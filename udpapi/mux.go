// Copyright (C) 2021 Allen Li
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package udpapi provides Go bindings for the AniDB UDP API.
//
// Documentation for the API can be found at
// https://wiki.anidb.net/UDP_API_Definition.
package udpapi

import (
	"bytes"
	"compress/flate"
	"context"
	"crypto/cipher"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// A Mux multiplexes AniDB UDP API requests on a single connection.
//
// Multiple goroutines may invoke methods on a Mux simultaneously.
type Mux struct {
	// Concurrency safe
	wg         sync.WaitGroup
	responses  responseMap
	tagCounter tagCounter

	// Set on init
	conn   net.Conn
	logger Logger

	// Mutex protected
	block   cipher.Block
	blockMu sync.Mutex
}

// NewMux makes a new Mux.
// You must call Close after use.
// The underlying conn will be closed internally and should not
// be closed directly by the caller.
func NewMux(conn net.Conn, o ...MuxOption) *Mux {
	m := &Mux{
		conn:   conn,
		logger: nullLogger{},
	}
	for _, o := range o {
		o.apply(m)
	}
	m.wg.Add(1)
	go func() {
		defer m.wg.Done()
		m.handleResponses()
	}()
	return m
}

// A MuxOption is passed to NewMux for configuration.
type MuxOption interface {
	apply(*Mux)
}

// A Logger can be used for logging.
// A Logger must be safe to use concurrently.
type Logger interface {
	Printf(string, ...interface{})
}

// UseLogger returns a MuxOption for setting a Logger.
func UseLogger(l Logger) MuxOption {
	return loggerOpt{l}
}

type loggerOpt struct {
	logger Logger
}

func (loggerOpt) muxOption() {}
func (o loggerOpt) apply(m *Mux) {
	m.logger = o.logger
}

// Request performs an AniDB UDP API request.
// args is modified by setting a new tag.
// This method does not handle retries or rate limiting.
//
// This method handles decompression and decryption, as they are
// necessary to parse response tags.
//
// See the AniDB UDP API documentation for more information.
//
// The returned error may be errors.Is with these errors:
//  context.DeadlineExceeded
//  net.Error
func (m *Mux) Request(ctx context.Context, cmd string, args url.Values) (Response, error) {
	ctx, cf := context.WithTimeout(ctx, 5*time.Second)
	defer cf()
	m.logger.Printf("Starting request cmd %s", cmd)
	t := m.tagCounter.next()
	args.Set("tag", string(t))
	req := []byte(cmd + " " + args.Encode())
	if b := m.getBlock(); b != nil {
		req = encrypt(b, req)
	}
	c := m.responses.waitFor(t)
	defer m.responses.cancel(t)
	m.logger.Printf("Sending cmd %s", cmd)
	// BUG(darkfeline): Network writes aren't governed by context deadlines.
	if _, err := m.conn.Write(req); err != nil {
		return Response{}, fmt.Errorf("udpapi: %w", err)
	}
	select {
	case <-ctx.Done():
		return Response{}, ctx.Err()
	case d := <-c:
		resp, err := parseResponse(d)
		if err != nil {
			return Response{}, fmt.Errorf("udpapi: %s", err)
		}
		return resp, nil
	}
}

// SetBlock sets the cipher block to use for future requests and responses.
// Set to nil to disable encryption and decryption.
//
// See the AniDB UDP API documentation for more information.
func (m *Mux) SetBlock(b cipher.Block) {
	m.blockMu.Lock()
	m.block = b
	m.blockMu.Unlock()
}

// Close immediately closes the Mux.
// The underlying connection is closed.
// No new requests will be accepted (as the connection is closed).
// Any Request calls waiting for responses will be unblocked.
func (m *Mux) Close() {
	_ = m.conn.Close()
	m.responses.close()
	m.wg.Wait()
}

// handleResponses handles incoming responses.
// Should be a called as a goroutine.
// Will exit when connection is closed.
func (m *Mux) handleResponses() {
	buf := make([]byte, 1400) // Max UDP size
	for {
		n, readErr := m.conn.Read(buf)
		if n > 0 {
			m.handleResponseData(buf[:n])
		}
		if readErr != nil {
			if errors.Is(readErr, net.ErrClosed) {
				return
			}
			var err net.Error
			if errors.As(readErr, &err) && !err.Temporary() {
				return
			}
			m.logger.Printf("Error reading from UDP conn: %s", readErr)
		}
	}
}

// handleResponseData handles one incoming response packet.
// Does decryption and decompression, as it is needed to match the response tag.
func (m *Mux) handleResponseData(data []byte) {
	if b := m.getBlock(); b != nil {
		var err error
		data, err = decrypt(b, data)
		if err != nil {
			m.logger.Printf("Error handling response: %s", err)
			return
		}
	}
	if len(data) > 2 && data[0] == 0 && data[1] == 0 {
		var err error
		data, err = decompress(data[2:])
		if err != nil {
			m.logger.Printf("Error handling response: %s", err)
			return
		}
	}
	m.responses.deliver(splitTag(data))
}

func (m *Mux) getBlock() cipher.Block {
	m.blockMu.Lock()
	defer m.blockMu.Unlock()
	return m.block
}

// A responseMap tracks pending UDP responses by tag, so they can be
// delivered out of order.
// This is concurrent safe.
type responseMap struct {
	m      sync.Map
	logger Logger // must be non-nil
}

func (m *responseMap) waitFor(t responseTag) <-chan []byte {
	c := make(chan []byte, 1)
	_, loaded := m.m.LoadOrStore(t, c)
	if loaded {
		panic(fmt.Sprintf("dupe tag %q", t))
	}
	return c
}

func (m *responseMap) deliver(t responseTag, b []byte) {
	v, loaded := m.m.LoadAndDelete(t)
	if !loaded {
		m.logger.Printf("Unknown tag %q for response", t)
		return
	}
	c := v.(chan []byte)
	c <- b
	close(c)
}

func (m *responseMap) cancel(t responseTag) {
	m.m.Delete(t)
}

// close delivers empty bytes to all pending responses.
// Doesn't handle any new pending responses created while close is running.
func (m *responseMap) close() {
	m.m.Range(func(key, value interface{}) bool {
		m.deliver(key.(responseTag), nil)
		return true
	})
}

type responseTag string

// A tagCounter generates sequential responseTags.
// This is concurrency safe.
type tagCounter struct {
	mu sync.Mutex
	c  int
}

func (c *tagCounter) next() responseTag {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.c++
	return responseTag(fmt.Sprintf("%x", c.c))
}

// splitTag splits the tag off a UDP response body.
func splitTag(b []byte) (responseTag, []byte) {
	parts := bytes.SplitN(b, []byte(" "), 2)
	tag := responseTag(parts[0])
	switch len(parts) {
	case 1:
		return tag, nil
	case 2:
		return tag, parts[1]
	default:
		panic(fmt.Sprintf("unexpected length %d", len(parts)))
	}
}

// A Response is an AniDB UDP API response.
type Response struct {
	Code   ReturnCode
	Header string
	Rows   [][]string
}

// parseResponse parses UDP responses, without the tag.
func parseResponse(b []byte) (Response, error) {
	m := string(b)
	lines := strings.Split(m, "\n")
	parts := strings.SplitN(lines[0], " ", 2)
	r := Response{}
	code, err := strconv.Atoi(parts[0])
	if err != nil {
		return r, fmt.Errorf("parse response: %s", err)
	}
	r.Code = ReturnCode(code)
	if len(parts) > 1 {
		r.Header = parts[1]
	}
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		row := strings.Split(line, "|")
		for i, f := range row {
			row[i] = unescapeField(f)
		}
		r.Rows = append(r.Rows, row)
	}
	return r, nil
}

// A ReturnCode is an AniDB UDP API return code.
// Note that even though ReturnCode implements error, not all ReturnCode values should be
// considered errors.
type ReturnCode int

const (
	// 505 ILLEGAL INPUT OR ACCESS DENIED
	IllegalInput ReturnCode = 505
	// 555 BANNED
	// {str reason}
	Banned ReturnCode = 555
	// 598 UNKNOWN COMMAND
	UnknownCmd ReturnCode = 598
	// 600 INTERNAL SERVER ERROR
	InternalErr ReturnCode = 600
	// 601 ANIDB OUT OF SERVICE - TRY AGAIN LATER
	OutOfService ReturnCode = 601
	// 602 SERVER BUSY - TRY AGAIN LATER
	ServerBusy ReturnCode = 602
	// 604 TIMEOUT - DELAY AND RESUBMIT
	Timeout ReturnCode = 604

	// Additional return codes for all commands that require login:
	// 501 LOGIN FIRST
	LoginFirst ReturnCode = 501
	// 502 ACCESS DENIED
	AccessDenied ReturnCode = 502
	// 506 INVALID SESSION
	InvalidSession ReturnCode = 506
)

//go:generate stringer -type=ReturnCode

func (c ReturnCode) Error() string {
	return fmt.Sprintf("return code %d %s", c, c.String())
}

// DEFLATE
func decompress(b []byte) ([]byte, error) {
	r := flate.NewReader(bytes.NewReader(b))
	defer r.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, fmt.Errorf("decompress: %s", err)
	}
	return buf.Bytes(), nil
}

// in place
// ECB, blockwise encryption
// PKCS#5 padding
func encrypt(c cipher.Block, b []byte) []byte {
	bs := c.BlockSize()
	if bs > 256 {
		panic(fmt.Sprintf("Unsupported block size %d", bs))
	}
	gap := bs - (len(b) % bs)
	pad := make([]byte, gap)
	for i := range pad {
		pad[i] = byte(gap)
	}
	b = append(b, pad...)
	for i := 0; i < len(b); i += bs {
		c.Encrypt(b[i:], b[i:])
	}
	return b
}

// in place
func decrypt(c cipher.Block, b []byte) ([]byte, error) {
	bs := c.BlockSize()
	if len(b)%bs != 0 {
		return nil, fmt.Errorf("decrypt blocks: incomplete blocks")
	}
	for i := 0; i < len(b); i += bs {
		c.Decrypt(b[i:], b[i:])
	}
	// PKCS#5 padding
	pad := b[len(b)-1]
	return b[:len(b)-int(pad)], nil
}

// unescape UDP field
func unescapeField(s string) string {
	s = strings.ReplaceAll(s, "<br />", "\n")
	s = strings.ReplaceAll(s, "`", "'")
	s = strings.ReplaceAll(s, "/", "|")
	return s
}

type nullLogger struct{}

func (nullLogger) Printf(string, ...interface{}) {}
