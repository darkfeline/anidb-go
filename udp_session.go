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
	"context"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// UDP proto ver
const protoVer = "3"

const defaultServer = "api.anidb.net:9000"

// An UDPConfig is used for configuring an AniDB UDP client.
type UDPConfig struct {
	// If empty, use default server.
	Server        string
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
	p      *reqPipe
	logger Logger

	// Mutex protected
	sessionKeyMu sync.Mutex
	sessionKey   string
	isNATMu      sync.Mutex
	isNAT        bool
}

// startUDPSession starts a UDP session.
// context is used for initializing the session only.
// You must close the session after use.
func startUDPSession(ctx context.Context, c *UDPConfig) (*udpSession, error) {
	srv := c.Server
	if srv == "" {
		srv = defaultServer
	}
	logger := c.Logger
	if logger == nil {
		logger = nullLogger{}
	}
	conn, err := net.Dial("udp", srv)
	if err != nil {
		return nil, fmt.Errorf("start UDP session: %s", err)
	}
	s := &udpSession{
		p:      newReqPipe(conn, newUDPLimiter(), logger),
		logger: logger,
	}
	// XXXXXXXXXXXX Close reqpipe
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

// Close immediately closes the session.
// Waits for any goroutines to exit.
func (s *udpSession) Close() {
	ctx, cf := context.WithTimeout(context.Background(), 2*time.Second)
	defer cf()
	_ = s.logout(ctx)
	s.p.close()
}

func (s *udpSession) sessionValues() url.Values {
	v := url.Values{}
	s.sessionKeyMu.Lock()
	v.Set("user", s.sessionKey)
	s.sessionKeyMu.Unlock()
	return v
}

// A Logger can be used for logging.
type Logger interface {
	// Printf must be concurrency safe.
	Printf(string, ...interface{})
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
