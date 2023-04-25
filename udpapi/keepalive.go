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
	r      requester
	logger Logger

	wg      sync.WaitGroup
	sleeper inactiveSleeper
	ctx     context.Context
	cf      context.CancelFunc

	lastPort   string
	interval   time.Duration
	timeoutHit bool
}

// newKeepAlive starts a keepalive goroutine to keep the AniDB UDP
// connection alive behind NAT.
// You must call start to actually start the keepalive.
// Logger must be non-nil.
func newKeepAlive(r requester, l Logger) *keepAlive {
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
		return fmt.Errorf("start keepalive: %s", err)
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
	k.sleeper.activate(t)
}

func (k *keepAlive) stop() {
	k.cf()
	k.wg.Wait()
}

// initialize keepalive, but without starting background goroutine.
// This is a separate method for testing purposes.
func (k *keepAlive) initialize() error {
	port, err := keepAlivePing(context.Background(), k.r)
	if err != nil {
		return err
	}
	k.sleeper.activate(time.Now())
	k.lastPort = port
	k.interval = time.Minute
	k.ctx, k.cf = context.WithCancel(context.Background())
	return nil
}

// background goroutine
func (k *keepAlive) background() {
	for {
		if err := k.sleeper.sleep(k.ctx, k.interval); err != nil {
			// error indicates context cancellation
			return
		}
		port, err := keepAlivePing(k.ctx, k.r)
		if err != nil {
			k.logger.Printf("Error: %s", err)
			k.interval += 10 * time.Second
			continue
		}
		k.updateInterval(time.Now(), port)
	}
}

const (
	minKeepAliveInterval = 30 * time.Second
	maxKeepAliveInterval = 5 * time.Minute
)

func (k *keepAlive) updateInterval(t time.Time, port string) {
	interval := k.sleeper.sinceActive(t)
	k.sleeper.activate(t)
	if k.lastPort != port {
		// If the actual interval is much greater than the
		// planned interval, then we can't infer anything from
		// the port change.  This should only happen when the
		// ping fails multiple times and is retried.
		if interval-k.interval > 10*time.Second {
			k.logger.Printf("Port reset, but interval %s much larger than expected %s",
				interval, k.interval)
			return
		}
		k.timeoutHit = true
		k.interval = k.interval - (10 * time.Second)
		k.logger.Printf("Port reset, lowering interval to %s", k.interval)
		if k.interval < minKeepAliveInterval {
			k.interval = minKeepAliveInterval
			k.logger.Printf("Minimum interval restricted to %s", k.interval)
		}
		k.lastPort = port
	} else if !k.timeoutHit {
		k.interval = interval + (10 * time.Second)
		k.logger.Printf("Timeout not hit, raising interval to %s", k.interval)
		if k.interval > maxKeepAliveInterval {
			k.interval = maxKeepAliveInterval
			k.logger.Printf("Maximum interval restricted to %s", k.interval)
		}
	}
}

// An inactiveSleeper tracks sleeping for a period of inactivity.
// Zero value is ready for use.
type inactiveSleeper struct {
	tmr          *time.Timer
	lastActive   time.Time
	lastActiveMu sync.Mutex
}

// activate indicates activity at the given time.
// Safe to call concurrently.
func (s *inactiveSleeper) activate(t time.Time) {
	s.lastActiveMu.Lock()
	if t.After(s.lastActive) {
		s.lastActive = t
	}
	s.lastActiveMu.Unlock()
}

// sleep sleeps until the duration is reached since last activity or
// context expires.
// Returns an error for context expiration.
// Must be called from at most one goroutine.
func (s *inactiveSleeper) sleep(ctx context.Context, d time.Duration) error {
	if s.tmr == nil {
		s.tmr = time.NewTimer(d)
	}
	for {
		elapsed := s.sinceActive(time.Now())
		if elapsed >= d {
			break
		}
		if !s.tmr.Stop() {
			<-s.tmr.C
		}
		s.tmr.Reset(d - elapsed)
		select {
		case <-s.tmr.C:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
}

func (s *inactiveSleeper) sinceActive(t time.Time) time.Duration {
	s.lastActiveMu.Lock()
	defer s.lastActiveMu.Unlock()
	return t.Sub(s.lastActive)
}

func (s *inactiveSleeper) afterActive(d time.Duration) time.Time {
	s.lastActiveMu.Lock()
	defer s.lastActiveMu.Unlock()
	return s.lastActive.Add(d)
}

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
