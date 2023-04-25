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
	"fmt"
	"net/url"
	"sync"
	"time"
)

// A requester is used to issue requests to AniDB UDP API.
type requester interface {
	Request(context.Context, string, url.Values) (Response, error)
}

var _ requester = &Mux{}

type keepAlive struct {
	r        requester
	logger   Logger
	interval time.Duration

	wg  sync.WaitGroup
	t   *time.Timer
	ctx context.Context
	cf  context.CancelFunc

	lastPort   string
	timeoutHit bool
}

// newKeepAlive starts a keepalive goroutine to keep the AniDB UDP
// connection alive behind NAT.
// You must call start to actually start the keepalive.
// Logger must be non-nil.
func newKeepAlive(c *keepAliveConfig) *keepAlive {
	k := &keepAlive{
		r:        c.r,
		logger:   c.logger,
		interval: c.interval,
		t:        time.NewTimer(0),
	}
	return k
}

type keepAliveConfig struct {
	r        requester
	logger   Logger
	interval time.Duration
}

// start starts the keepalive.
// You must call stop after use.
func (k *keepAlive) start() error {
	if err := k.initialize(); err != nil {
		return fmt.Errorf("start keepalive: %s", err)
	}
	k.wg.Add(1)
	go func() {
		defer k.wg.Done()
		k.background()
	}()
	return nil
}

func (k *keepAlive) stop() {
	k.cf()
	k.wg.Wait()
}

// initialize keepalive, but without starting background goroutine.
// This is a separate method for testing purposes.
func (k *keepAlive) initialize() error {
	ctx := context.Background()
	port, err := keepAlivePing(ctx, k.r)
	if err != nil {
		return err
	}
	k.lastPort = port
	k.ctx, k.cf = context.WithCancel(ctx)
	return nil
}

// background goroutine
func (k *keepAlive) background() {
	for {
		if !k.t.Stop() {
			<-k.t.C
		}
		k.t.Reset(k.interval)
		select {
		case <-k.ctx.Done():
			return
		case <-k.t.C:
		}
		port, err := keepAlivePing(k.ctx, k.r)
		if err != nil {
			k.logger.Printf("keepalive ping error: %s", err)
			continue
		}
		if port != k.lastPort {
			new := k.interval / 2
			k.logger.Printf("keepalive: port %d != lastPort %d; NAT expired, lowering interval from %s to %s",
				port, k.lastPort, k.interval)
			k.interval = new
		}
	}
}

// keepAlivePing sends one PING request.
// This function sets a timeout on the request.
func keepAlivePing(ctx context.Context, r requester) (port string, _ error) {
	ctx, cf := context.WithTimeout(ctx, 2*time.Second)
	defer cf()
	resp, err := r.Request(ctx, "PING", url.Values{"nat": []string{"1"}})
	if err != nil {
		return "", err
	}
	if resp.Code != 300 {
		return "", fmt.Errorf("ping: unexpected return code %d", resp.Code)
	}
	if len(resp.Rows) < 1 || len(resp.Rows[0]) < 1 {
		return "", fmt.Errorf("ping: unexpected response rows: %v", resp.Rows)
	}
	return resp.Rows[0][0], nil
}
