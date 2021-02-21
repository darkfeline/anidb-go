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
	"crypto/aes"
	"crypto/rand"
	"reflect"
	"testing"
	"time"
)

func TestReqPipe(t *testing.T) {
	t.Parallel()
}

func TestResponseMap(t *testing.T) {
	t.Parallel()
	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		m := responseMap{logger: testLogger{t, "response map: "}}
		ctx, cf := context.WithTimeout(context.Background(), time.Second)
		t.Cleanup(cf)
		t.Run("first tag", func(t *testing.T) {
			c := m.waitFor("shefi")
			t.Parallel()
			select {
			case got := <-c:
				const want = "shifuna"
				if string(got) != want {
					t.Errorf("Got %q, want %q", got, want)
				}
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
		})
		t.Run("second tag", func(t *testing.T) {
			c := m.waitFor("kyaru")
			t.Parallel()
			select {
			case got := <-c:
				const want = "kiruya"
				if string(got) != want {
					t.Errorf("Got %q, want %q", got, want)
				}
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
		})
		m.deliver("kyaru", []byte("kiruya"))
		m.deliver("shefi", []byte("shifuna"))
	})
	t.Run("close", func(t *testing.T) {
		t.Parallel()
		m := responseMap{logger: testLogger{t, "response map: "}}
		ctx, cf := context.WithTimeout(context.Background(), time.Second)
		t.Cleanup(cf)
		t.Run("first tag", func(t *testing.T) {
			c := m.waitFor("shefi")
			t.Parallel()
			select {
			case got := <-c:
				const want = ""
				if string(got) != want {
					t.Errorf("Got %q, want %q", got, want)
				}
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
		})
		m.close()
	})
}

func TestParseResponse(t *testing.T) {
	t.Parallel()
	const data = `720 1234 NOTIFICATION - NEW FILE
1234|12|34`
	got, err := parseResponse([]byte(data))
	if err != nil {
		t.Fatal(err)
	}
	want := response{
		code:   720,
		header: "1234 NOTIFICATION - NEW FILE",
		rows: [][]string{
			{"1234", "12", "34"},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Got %#v, want %#v", got, want)
	}
}

func TestEncryptDecrypt(t *testing.T) {
	t.Parallel()
	// AES-128, 16 bytes
	const key = "\x80\xa2_\xcaa\xb6\f\xa9X\xa5\xff\x9am\xebי"
	cb, err := aes.NewCipher([]byte(key))
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		desc string
		size int
	}{
		{desc: "3 bytes", size: 3},
		{desc: "16 bytes", size: 16},
		{desc: "17 bytes", size: 17},
		{desc: "31 bytes", size: 31},
		{desc: "32 bytes", size: 32},
		{desc: "33 bytes", size: 33},
		{desc: "64 bytes", size: 64},
	}
	for _, c := range cases {
		c := c
		t.Run(c.desc, func(t *testing.T) {
			t.Parallel()
			orig := make([]byte, c.size)
			if _, err := rand.Read(orig); err != nil {
				t.Fatal(err)
			}
			data := make([]byte, len(orig))
			copy(data, orig)
			data = encrypt(cb, data)
			if reflect.DeepEqual(orig, data) {
				t.Fatalf("data not encrypted")
			}
			t.Logf("encrypted data is %d bytes", len(data))
			data, err = decrypt(cb, data)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(orig, data) {
				t.Errorf("decrypted not equal, got %x, want %x", data, orig)
			}
		})
	}
}

type testLogger struct {
	t      *testing.T
	prefix string
}

func (l testLogger) Printf(format string, v ...interface{}) {
	l.t.Helper()
	l.t.Logf(l.prefix+format, v...)
}
