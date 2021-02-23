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

package udpapi

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// UDP proto ver
const protoVer = "3"

const defaultServer = "api.anidb.net:9000"

// An sessionConfig is used for starting an AniDB UDP session.
type sessionConfig struct {
	UserName      string
	UserPassword  string
	ClientName    string
	ClientVersion int32
	// For encryption, optional.
	APIKey string
	// Logger should add a prefix if needed.  Optional.
	Logger Logger
}

// A udpSession represents an authenticated UDP session.
// A udpSession's methods are concurrency safe.
type udpSession struct {
	// Set on init
	p      *Mux
	logger Logger

	// Mutex protected
	sessionKeyMu sync.Mutex
	sessionKey   string
	isNATMu      sync.Mutex
	isNAT        bool
}

// startUDPSession starts a UDP session.
// context is used for initializing the session only.
// Muxs must only be used with a single session at a time.
// You must close the session after use. XXXXXXXXXXXXXXXXXX
func startUDPSession(ctx context.Context, p *Mux, c *sessionConfig) (_ *udpSession, err error) {
	logger := c.Logger
	if logger == nil {
		logger = nullLogger{}
	}
	s := &udpSession{
		p:      p,
		logger: logger,
	}
	////////////////// handle existing session
	if c.APIKey != "" {
		if err := s.encrypt(ctx, c.UserName, c.APIKey); err != nil {
			return nil, fmt.Errorf("start UDP session: %s", err)
		}
	}
	if err := s.auth(ctx, c); err != nil {
		return nil, fmt.Errorf("start UDP session: %s", err)
	}
	if s.isNAT {
		// XXXXXXXXXXXX
		// ping
	}
	// XXXXXXXXXXXX
	// keepalive
	// logout

	return s, nil
}

// close immediately closes the session.
func (s *udpSession) close() {
	ctx, cf := context.WithTimeout(context.Background(), 2*time.Second)
	defer cf()
	_ = s.logout(ctx) // XXXXXXXXXX shouldn't always logout?
	s.p.Close()
}

func (s *udpSession) sessionValues() url.Values {
	v := url.Values{}
	s.sessionKeyMu.Lock()
	v.Set("user", s.sessionKey)
	s.sessionKeyMu.Unlock()
	return v
}

// XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXx
// request performs a UDP request.  Handles retries.
// args is modified with a new tag.
// Concurrency safe.
func (m *Mux) tmpRequest(ctx context.Context, cmd string, args url.Values) (Response, error) {
	m.logger.Printf("Starting request cmd %s", cmd)
	for ctx.Err() == nil {
		resp, err := m.Request(ctx, cmd, args)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				// XXXXXXXX retry
			}
			return Response{}, fmt.Errorf("reqpipe request: %s", err)
		}
		// XXXXXXXX check for retriable returnCode
		return resp, nil
	}
	return Response{}, fmt.Errorf("reqpipe request: %w", ctx.Err())
}

func retryCommand(ctx context.Context, m *Mux, cmd string, args url.Values) (Response, error) {
	panic("Not implemented")
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

func (l udpLimiter) close() {
	l.short.SetLimit(rate.Inf)
	l.long.SetLimit(rate.Inf)
}
