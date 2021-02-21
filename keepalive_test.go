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
	"net/url"
	"testing"
)

func TestKeepAlive(t *testing.T) {
	t.Parallel()
	r := &fakeRequester{
		resp: response{
			code:   300,
			header: "PONG",
			rows:   [][]string{{"123"}},
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

type fakeRequester struct {
	resp response
	err  error
}

func (r *fakeRequester) request(ctx context.Context, cmd string, v url.Values) (response, error) {
	return r.resp, r.err
}
