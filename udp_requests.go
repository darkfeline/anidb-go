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
	"crypto/md5"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// encrypt RPC call.
// Not concurrent safe.
func (s *Session) encrypt(ctx context.Context, user string, key string) error {
	v := url.Values{}
	v.Set("user", user)
	v.Set("type", "1")
	resp, err := s.request(ctx, "ENCRYPT", v)
	if err != nil {
		return fmt.Errorf("encrypt: %s", err)
	}
	switch resp.code {
	case 209:
		parts := strings.SplitN(resp.header, " ", 2)
		salt := parts[0]
		sum := md5.Sum([]byte(key + salt))
		s.block, err = aes.NewCipher(sum[:])
		if err != nil {
			return fmt.Errorf("encrypt: %s", err)
		}
		return nil
	default:
		return fmt.Errorf("encrypt: bad code %d %q", resp.code, resp.header)
	}
}

// auth RPC call.
// Concurrent safe.
func (s *Session) auth(ctx context.Context, cfg *UDPConfig) error {
	v := url.Values{}
	v.Set("user", cfg.UserName)
	v.Set("pass", cfg.UserPassword)
	v.Set("protover", protoVer)
	v.Set("client", cfg.ClientName)
	v.Set("clientver", strconv.Itoa(int(cfg.ClientVersion)))
	v.Set("nat", "1")
	v.Set("comp", "1")
	resp, err := s.request(ctx, "AUTH", v)
	if err != nil {
		return fmt.Errorf("auth request: %s", err)
	}
	switch resp.code {
	case 201:
		s.log("new anidb UDP API version available")
		// TODO Expose update available info to library clients
		fallthrough
	case 200:
		parts := strings.SplitN(resp.header, " ", 3)
		if len(parts) < 3 {
			return fmt.Errorf("auth request: invalid response header %q", resp.header)
		}
		s.muSessionKey.Lock()
		s.sessionKey = parts[0]
		s.muSessionKey.Unlock()
		// TODO Make address comparison reliable
		if s.conn.LocalAddr().String() != parts[1] {
			s.muIsNAT.Lock()
			s.isNAT = true
			s.muIsNAT.Unlock()
		}
		return nil
	default:
		return fmt.Errorf("auth request: bad code %d %s", resp.code, resp.header)
	}
}

// logout RPC call.
// Concurrent safe.
func (s *Session) logout(ctx context.Context) error {
	v := s.sessionValues()
	resp, err := s.request(ctx, "LOGOUT", v)
	if err != nil {
		return fmt.Errorf("logout request: %s", err)
	}
	s.muSessionKey.Lock()
	s.sessionKey = ""
	s.muSessionKey.Unlock()
	switch resp.code {
	case 203:
		return nil
	default:
		return fmt.Errorf("logout request: bad code %d %s", resp.code, resp.header)
	}
}
