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

package anidb

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

// A closeLimiter is a Limiter that has a Close method to unblock all waiters.
type closeLimiter interface {
	Limiter
	// close unblocks all waiters.
	// This method must be safe to call concurrently.
	// All Wait calls afterward must also be unblocked.
	close()
}

// A reqPipe serializes and demuxes AniDB UDP requests.
type reqPipe struct {
	// Concurrency safe
	wg         sync.WaitGroup
	responses  responseMap
	tagCounter tagCounter

	// Set on init
	conn    net.Conn
	limiter closeLimiter
	logger  Logger

	// Mutex protected
	block   cipher.Block
	blockMu sync.Mutex
}

func newReqPipe(conn net.Conn, l closeLimiter, logger Logger) *reqPipe {
	if logger == nil {
		logger = nullLogger{}
	}
	p := &reqPipe{
		conn:    conn,
		limiter: l,
		logger:  logger,
	}
	p.responses.logger = logger
	go p.handleResponses()
	return p
}

// request performs a UDP request.  Handles retries.
// args is modified with a new tag.
// Concurrency safe.
func (p *reqPipe) request(ctx context.Context, cmd string, args url.Values) (response, error) {
	p.logger.Printf("Starting request cmd %s", cmd)
	for ctx.Err() == nil {
		resp, err := p.requestOnce(ctx, cmd, args)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				// XXXXXXXX retry
			}
			return response{}, fmt.Errorf("reqpipe request: %s", err)
		}
		// XXXXXXXX check for retriable returnCode
		return resp, nil
	}
	return response{}, fmt.Errorf("reqpipe request: %w", ctx.Err())
}

// setBlock sets the cipher block to use for future requests.
// Set to nil to unset.
// Concurrency safe.
func (p *reqPipe) setBlock(b cipher.Block) {
	p.blockMu.Lock()
	p.block = b
	p.blockMu.Unlock()
}

// close immediately closes the pipe.
// Waits for any goroutines to exit.
// Concurrency safe.
func (p *reqPipe) close() {
	_ = p.conn.Close()
	p.limiter.close()
	p.responses.close()
	p.wg.Wait()
}

// requestOnce sends a single UDP request packet.  No retries.
// args is modified with a new tag.
// Returned error may be errors.Is:
//  context.DeadlineExceeded
//  returnCode
func (p *reqPipe) requestOnce(ctx context.Context, cmd string, args url.Values) (response, error) {
	ctx, cf := context.WithTimeout(ctx, 5*time.Second)
	defer cf()
	t := p.tagCounter.next()
	args.Set("tag", string(t))
	req := []byte(cmd + " " + args.Encode())
	if b := p.getBlock(); b != nil {
		req = encrypt(b, req)
	}
	p.logger.Printf("Waiting to send cmd %s", cmd)
	if err := p.limiter.Wait(ctx); err != nil {
		return response{}, err
	}
	c := p.responses.waitFor(t)
	defer p.responses.cancel(t)
	p.logger.Printf("Sending cmd %s", cmd)
	if _, err := p.conn.Write(req); err != nil {
		return response{}, err
	}
	select {
	case <-ctx.Done():
		return response{}, ctx.Err()
	case d := <-c:
		resp, err := parseResponse(d)
		if err != nil {
			return response{}, err
		}
		return resp, nil
	}
}

// handleResponses handles incoming responses.
// Should be a called as a goroutine.
// Will exit when connection is closed.
func (p *reqPipe) handleResponses() {
	p.wg.Add(1)
	defer p.wg.Done()
	buf := make([]byte, 1400) // Max UDP size
	for {
		n, readErr := p.conn.Read(buf)
		if n > 0 {
			p.handleResponseData(buf[:n])
		}
		if readErr != nil {
			if readErr == io.EOF {
				return
			}
			var err net.Error
			if errors.As(readErr, &err) && !err.Temporary() {
				return
			}
			p.logger.Printf("error reading from UDP conn: %s", readErr)
		}
	}
}

// handleResponseData handles one incoming response packet.
// Does decryption and decompression, as it is needed to match the response tag.
func (p *reqPipe) handleResponseData(data []byte) {
	if b := p.getBlock(); b != nil {
		var err error
		data, err = decrypt(b, data)
		if err != nil {
			p.logger.Printf("error: %s", err)
			return
		}
	}
	if len(data) > 2 && data[0] == 0 && data[1] == 0 {
		var err error
		data, err = decompress(data[2:])
		if err != nil {
			p.logger.Printf("error: %s", err)
			return
		}
	}
	p.responses.deliver(splitTag(data))
}

func (p *reqPipe) getBlock() cipher.Block {
	p.blockMu.Lock()
	defer p.blockMu.Unlock()
	return p.block
}

// A responseMap tracks pending UDP responses by tag, so they can be
// delivered out of order.
// This is concurrent safe.
type responseMap struct {
	m      sync.Map
	logger Logger
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

type response struct {
	code   returnCode
	header string
	rows   [][]string
}

// parseResponse parses UDP responses, without the tag.
func parseResponse(b []byte) (response, error) {
	p := string(b)
	lines := strings.Split(p, "\n")
	parts := strings.SplitN(lines[0], " ", 2)
	r := response{}
	code, err := strconv.Atoi(parts[0])
	if err != nil {
		return r, fmt.Errorf("parse response: %s", err)
	}
	r.code = returnCode(code)
	if len(parts) > 1 {
		r.header = parts[1]
	}
	for _, line := range lines[1:] {
		if line == "" {
			continue
		}
		row := strings.Split(line, "|")
		for i, f := range row {
			row[i] = unescapeField(f)
		}
		r.rows = append(r.rows, row)
	}
	return r, nil
}

// UDP API return code.
// Note that returnCode implements error, but not all codes should be
// considered errors.
type returnCode int

const (
	// 505 ILLEGAL INPUT OR ACCESS DENIED
	illegalInput returnCode = 505
	// 555 BANNED
	// {str reason}
	banned returnCode = 555
	// 598 UNKNOWN COMMAND
	unknownCmd returnCode = 598
	// 600 INTERNAL SERVER ERROR
	internalErr returnCode = 600
	// 601 ANIDB OUT OF SERVICE - TRY AGAIN LATER
	outOfService returnCode = 601
	// 602 SERVER BUSY - TRY AGAIN LATER
	serverBusy returnCode = 602
	// 604 TIMEOUT - DELAY AND RESUBMIT
	timeout returnCode = 604

	// Additional return codes for all commands that require login:
	// 501 LOGIN FIRST
	loginFirst returnCode = 501
	// 502 ACCESS DENIED
	accessDenied returnCode = 502
	// 506 INVALID SESSION
	invalidSession returnCode = 506
)

//go:generate stringer -type=returnCode

func (c returnCode) Error() string {
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
