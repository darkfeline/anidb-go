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

	"golang.org/x/time/rate"
)

// UDP proto ver
const protoVer = "3"

const defaultServer = "api.anidb.net:9000"

// An UDPConfig is used for starting UDP sessions.
type UDPConfig struct {
	// If nil, use default server.
	Server *net.UDPAddr
	// Local source port
	Local         *net.UDPAddr
	UserName      string
	UserPassword  string
	ClientName    string
	ClientVersion int32
	// For encryption, optional.
	APIKey string
	// Logger should add a prefix if needed.  Optional.
	Logger Logger
}

// StartUDP starts a UDP session.
// context is used for initializing the session only.
// You must close the session after use.
func StartUDP(ctx context.Context, c *UDPConfig) (*Session, error) {
	srv := c.Server
	if srv == nil {
		var err error
		srv, err = net.ResolveUDPAddr("udp", defaultServer)
		if err != nil {
			return nil, fmt.Errorf("start anidb UDP: %s", err)
		}
	}
	conn, err := net.DialUDP("udp", c.Local, srv)
	if err != nil {
		return nil, fmt.Errorf("start anidb UDP: %s", err)
	}
	s := &Session{
		conn:    conn,
		limiter: newUDPLimiter(),
		logger:  c.Logger,
	}
	s.responses.logger = c.Logger
	if c.APIKey != "" {
		if err := s.encrypt(ctx, c.UserName, c.APIKey); err != nil {
			return nil, fmt.Errorf("start anidb UDP: %s", err)
		}
	}
	if err := s.auth(ctx, c); err != nil {
		return nil, fmt.Errorf("start anidb UDP: %s", err)
	}
	if s.isNAT {
		// XXXXXXXXXXXX
		// ping
	}
	// XXXXXXXXXXXX
	// keepalive
	// logout

	go s.handleResponses()
	return s, nil
}

// A Logger can be used for logging.
type Logger interface {
	// Printf must be concurrency safe.
	Printf(string, ...interface{})
}

// A Session represents an authenticated UDP session.
// A Session's methods are concurrency safe.
type Session struct {
	// Concurrency safe
	wg         sync.WaitGroup
	responses  responseMap
	tagCounter tagCounter

	// Set on init
	conn    *net.UDPConn
	limiter Limiter
	logger  Logger

	// Mutex protected
	muSessionKey sync.Mutex
	sessionKey   string
	muIsNAT      sync.Mutex
	isNAT        bool

	// Unsafe concurrent set
	block cipher.Block
}

// Close immediately closes the session.
// Waits for any goroutines to exit.
func (s *Session) Close() {
	// XXXX logout
	_ = s.conn.Close()
	// Won't have new requests since connection is closed.
	s.responses.close()
	s.wg.Wait()
}

// request performs a UDP request.  Handles retries.
// args is modified with a new tag.
func (s *Session) request(ctx context.Context, cmd string, args url.Values) (response, error) {
	for ctx.Err() == nil {
		resp, err := s.requestOnce(ctx, cmd, args)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				// XXXXXXXX retry
			}
			return response{}, fmt.Errorf("anidb UDP request: %s", err)
		}
		// XXXXXXXX check for retriable codes
		return resp, nil
	}
	return response{}, fmt.Errorf("anidb UDP request: %s", ctx.Err())
}

// requestOnce sends a single UDP request packet.  No retries.
// args is modified with a new tag.
// Returned error may be errors.Is:
//  context.DeadlineExceeded
//  returnCode
func (s *Session) requestOnce(ctx context.Context, cmd string, args url.Values) (response, error) {
	ctx, cf := context.WithTimeout(ctx, 5*time.Second)
	defer cf()
	t := s.tagCounter.next()
	args.Set("tag", string(t))
	req := []byte(cmd + " " + args.Encode())
	if s.block != nil {
		req = encrypt(s.block, req)
	}
	if err := s.limiter.Wait(ctx); err != nil {
		return response{}, err
	}
	if _, err := s.conn.Write(req); err != nil {
		return response{}, err
	}
	b, err := s.responses.waitFor(ctx, t)
	if err != nil {
		return response{}, err
	}
	resp, err := parseResponse(b)
	if err != nil {
		return response{}, err
	}
	return resp, nil
}

// handleResponses handles incoming responses.
// Should be a called as a goroutine.
func (s *Session) handleResponses() {
	s.wg.Add(1)
	defer s.wg.Done()
	buf := make([]byte, 1400) // Max UDP size
	for {
		n, readErr := s.conn.Read(buf)
		if n > 0 {
			s.handleResponseData(buf[:n])
		}
		if readErr != nil {
			if readErr == io.EOF {
				return
			}
			var err net.Error
			if errors.As(readErr, &err) && !err.Temporary() {
				return
			}
			s.log("error reading from UDP conn: %s", readErr)
		}
	}
}

// handleResponseData handles one incoming response packet.
// Does decryption and decompression, as it is needed to match the response tag.
func (s *Session) handleResponseData(data []byte) {
	if s.block != nil {
		var err error
		data, err = decrypt(s.block, data)
		if err != nil {
			s.log("error: %s", err)
			return
		}
	}
	if len(data) > 2 && data[0] == 0 && data[1] == 0 {
		var err error
		data, err = decompress(data[2:])
		if err != nil {
			s.log("error: %s", err)
			return
		}
	}
	parts := bytes.SplitN(data, []byte(" "), 2)
	tag := responseTag(parts[0])
	switch len(parts) {
	case 1:
		s.responses.deliver(tag, nil)
	case 2:
		s.responses.deliver(tag, parts[1])
	}
}

// concurrent safe
func (s *Session) log(format string, v ...interface{}) {
	if s.logger == nil {
		return
	}
	s.logger.Printf(format, v...)
}

// A responseMap tracks pending UDP responses by tag, so they can be
// delivered out of order.
// This is concurrent safe.
type responseMap struct {
	m      sync.Map
	logger Logger
}

func (m *responseMap) waitFor(ctx context.Context, t responseTag) ([]byte, error) {
	c := make(chan []byte, 1)
	_, loaded := m.m.LoadOrStore(t, c)
	if loaded {
		panic(fmt.Sprintf("dupe tag %q", t))
	}
	select {
	case b := <-c:
		return b, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (m *responseMap) deliver(t responseTag, b []byte) {
	v, loaded := m.m.LoadAndDelete(t)
	if !loaded {
		m.log("Unknown tag %q for response", t)
		return
	}
	c := v.(chan<- []byte)
	c <- b
	close(c)
}

func (m *responseMap) log(format string, v ...interface{}) {
	if m.logger != nil {
		m.logger.Printf(format, v)
	}
}

// close delivers empty bytes to all pending responses.
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

type response struct {
	code   returnCode
	header string
	rows   [][]string
}

// parseResponse parses UDP responses, without the tag.
func parseResponse(b []byte) (response, error) {
	s := string(b)
	lines := strings.Split(s, "\n")
	parts := strings.SplitN(lines[0], " ", 2)
	r := response{}
	code, err := strconv.Atoi(parts[0])
	if err != nil {
		return r, fmt.Errorf("parse response: %s", err)
	}
	// TODO Check if it's a known code
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
	return fmt.Sprintf("return code %d %s", c, c)
}

// A udpLimiter complies with AniDB UDP API recommendations.
type udpLimiter struct {
	short *rate.Limiter
	long  *rate.Limiter
}

func newUDPLimiter() udpLimiter {
	return udpLimiter{
		// Every 2 sec short term
		short: rate.NewLimiter(0.5, 1),
		// Every 4 sec long term after 60 seconds
		long: rate.NewLimiter(0.25, 60/2),
	}
}

func (l udpLimiter) Wait(ctx context.Context) error {
	if err := l.long.Wait(ctx); err != nil {
		return err
	}
	if err := l.short.Wait(ctx); err != nil {
		return err
	}
	return nil
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
func encrypt(c cipher.Block, b []byte) []byte {
	bs := c.BlockSize()
	if bs > 256 {
		panic(fmt.Sprintf("Unsupported block size %d", bs))
	}
	for i := 0; i < len(b); i += bs {
		// PKCS#5 padding
		if i+bs >= len(b) {
			gap := bs - (len(b) % bs)
			pad := make([]byte, gap)
			for i := range pad {
				pad[i] = byte(gap)
			}
			b = append(b, pad...)
		}
		c.Encrypt(b[i:], b[i:])
	}
	// PKCS#5 padding for full block
	if len(b)%bs == 0 {
		pad := make([]byte, bs)
		for i := range pad {
			pad[i] = byte(bs)
		}
		b = append(b, pad...)
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
