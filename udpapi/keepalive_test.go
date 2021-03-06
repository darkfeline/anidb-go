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
	"net/url"
	"testing"
	"time"
)

func TestKeepAlive(t *testing.T) {
	t.Parallel()
	r := &fakeRequester{
		resp: Response{
			Code:   300,
			Header: "PONG",
			Rows:   [][]string{{"123"}},
		},
	}
	k := newKeepAlive(r, testLogger{t, "keepalive: "})
	if err := k.initialize(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(k.stop)
	t.Run("raise", func(t *testing.T) {
		prevInterval := k.interval
		newTime := k.sleeper.afterActive(prevInterval)
		k.updateInterval(newTime, "123")
		if k.interval <= prevInterval {
			t.Errorf("Expected new interval greater than %s; got %s",
				prevInterval, k.interval)
		}
	})
	t.Run("raise 2", func(t *testing.T) {
		prevInterval := k.interval
		newTime := k.sleeper.afterActive(prevInterval)
		k.updateInterval(newTime, "123")
		if k.interval <= prevInterval {
			t.Errorf("Expected new interval greater than %s; got %s",
				prevInterval, k.interval)
		}
	})
	t.Run("timeout", func(t *testing.T) {
		prevInterval := k.interval
		newTime := k.sleeper.afterActive(prevInterval)
		k.updateInterval(newTime, "555")
		if k.interval >= prevInterval {
			t.Errorf("Expected new interval less than %s; got %s",
				prevInterval, k.interval)
		}
	})
	t.Run("sustain", func(t *testing.T) {
		prevInterval := k.interval
		newTime := k.sleeper.afterActive(prevInterval)
		k.updateInterval(newTime, "555")
		if k.interval != prevInterval {
			t.Errorf("Expected new interval equal to %s; got %s",
				prevInterval, k.interval)
		}
	})
}

func TestKeepAlive_large_interval_okay(t *testing.T) {
	t.Parallel()
	r := &fakeRequester{
		resp: Response{
			Code:   300,
			Header: "PONG",
			Rows:   [][]string{{"123"}},
		},
	}
	k := newKeepAlive(r, testLogger{t, "keepalive: "})
	if err := k.initialize(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(k.stop)
	prevInterval := k.interval
	newTime := k.sleeper.afterActive(prevInterval).Add(time.Minute)
	k.updateInterval(newTime, "123")
	if k.interval <= prevInterval+(time.Minute/2) {
		t.Errorf("Expected new interval greater than %s; got %s",
			prevInterval, k.interval)
	}
}

func TestKeepAlive_large_interval_timeout(t *testing.T) {
	t.Parallel()
	r := &fakeRequester{
		resp: Response{
			Code:   300,
			Header: "PONG",
			Rows:   [][]string{{"123"}},
		},
	}
	k := newKeepAlive(r, testLogger{t, "keepalive: "})
	if err := k.initialize(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(k.stop)
	prevInterval := k.interval
	newTime := k.sleeper.afterActive(prevInterval).Add(time.Minute)
	k.updateInterval(newTime, "555")
	if k.interval < prevInterval {
		t.Errorf("Interval %s shouldn't be lower than previous %s",
			k.interval, prevInterval)
	}
}

type fakeRequester struct {
	resp Response
	err  error
}

func (r *fakeRequester) Request(ctx context.Context, cmd string, v url.Values) (Response, error) {
	return r.resp, r.err
}
