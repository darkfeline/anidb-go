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
	"net/url"
	"sync"
	"time"
)

const (
	minKeepAliveInterval = 30 * time.Second
	maxKeepAliveInterval = 5 * time.Minute
)

type udpRequester interface {
	request(context.Context, string, url.Values) (response, error)
}

var _ udpRequester = &reqPipe{}

type keepAlive struct {
	r      udpRequester
	logger Logger // Must be non-nil

	wg         sync.WaitGroup
	sleepTimer *time.Timer
	ctx        context.Context
	cf         context.CancelFunc

	lastRequest   time.Time
	lastRequestMu sync.Mutex
	lastPort      string
	interval      time.Duration
	timeoutHit    bool
}

// newKeepAlive starts a keepalive goroutine to keep the AniDB UDP
// connection alive behind NAT.
// You must call start to actually start the keepalive.
func newKeepAlive(r udpRequester, l Logger) *keepAlive {
	k := &keepAlive{
		r:      r,
		logger: l,
	}
	return k
}

// start starts the keepalive.
// You must call stop after use.
func (k *keepAlive) start() error {
	if err := k.initialize(); err != nil {
		fmt.Errorf("start keepalive: %s", err)
	}
	k.wg.Add(1)
	go func() {
		defer k.wg.Done()
		k.background()
	}()
	return nil
}

// notify notifies keepAlive that a packet was sent at the given time in
// order to accurately calibrate the keepalive interval.
// Concurrent safe.
func (k *keepAlive) notify(t time.Time) {
	k.lastRequestMu.Lock()
	k.lastRequest = t
	k.lastRequestMu.Unlock()
}

func (k *keepAlive) stop() {
	k.cf()
	k.wg.Wait()
}

// initialize keepalive, but without starting background goroutine.
// For testing.
func (k *keepAlive) initialize() error {
	port, err := keepAlivePing(context.Background(), k.r)
	if err != nil {
		return err
	}
	k.notify(time.Now())
	k.lastPort = port
	k.sleepTimer = time.NewTimer(time.Hour)
	k.interval = time.Minute
	k.ctx, k.cf = context.WithCancel(context.Background())
	return nil
}

// background goroutine
func (k *keepAlive) background() {
	for {
		if err := k.sleepUntilInterval(k.ctx); err != nil {
			return
		}
		port, err := keepAlivePing(k.ctx, k.r)
		if err != nil {
			// TODO Faster retry on error
			k.logger.Printf("Error: %s", err)
			continue
		}
		k.updateInterval(time.Now(), port)
	}
}

func (k *keepAlive) updateInterval(t time.Time, port string) {
	k.lastRequestMu.Lock()
	interval := t.Sub(k.lastRequest)
	k.lastRequest = t
	k.lastRequestMu.Unlock()
	if k.lastPort != port {
		k.timeoutHit = true
		k.interval = interval - (10 * time.Second)
		k.logger.Printf("Port reset, lowering interval to %s", k.interval)
		k.lastPort = port
	} else if !k.timeoutHit {
		k.interval = k.interval + (10 * time.Second)
		k.logger.Printf("Timeout not hit, raising interval to %s", k.interval)
	}
}

// sleepUntilInterval sleeps until the interval is reached since last
// request or context expires.
// Returns an error for context expiration.
func (k *keepAlive) sleepUntilInterval(ctx context.Context) error {
	elapsed := time.Now().Sub(k.lastRequest)
	for elapsed < k.interval {
		if !k.sleepTimer.Stop() {
			<-k.sleepTimer.C
		}
		k.sleepTimer.Reset(k.interval - elapsed)
		select {
		case t := <-k.sleepTimer.C:
			elapsed = t.Sub(k.lastRequest)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

// An inactiveSleeper tracks sleeping for a period of inactivity.
type inactiveSleeper struct {
	interval time.Duration
}

func (s *inactiveSleeper) activate(t time.Time) {

}

func (s *inactiveSleeper) sleep(t time.Time) {

}

func keepAlivePing(ctx context.Context, r udpRequester) (port string, _ error) {
	ctx, cf := context.WithTimeout(ctx, 2*time.Second)
	defer cf()
	resp, err := r.request(ctx, "PING", url.Values{"nat": []string{"1"}})
	if err != nil {
		return "", err
	}
	// TODO check for bad returnCode, retries
	if len(resp.rows) < 1 || len(resp.rows[0]) < 1 {
		return "", fmt.Errorf("ping: unexpected response rows")
	}
	return resp.rows[0][0], nil
}
