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
	"crypto/aes"
	"crypto/md5"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
)

const protoVer = "3"
const defaultServer = "api.anidb.net:9000"

// A Client is an AniDB UDP API client.
//
// The client handles basic rate limiting.
// The client does not handle retries or most errors.
//
// Due to the complexity of correctly using the UDP API, there may be
// bugs and/or feature gaps.
type Client struct {
	conn          net.Conn
	m             *Mux
	limiter       *limiter
	sessionKeyMu  sync.Mutex
	sessionKey    string
	ClientName    string
	ClientVersion int32
}

// NewClient creates a new Client.
// The caller should set ClientName and ClientVersion on the returned Client.
func NewClient(o ...ClientOption) (*Client, error) {
	conn, err := net.Dial("udp", defaultServer)
	if err != nil {
		return nil, fmt.Errorf("udpapi: %w", err)
	}
	c := &Client{
		conn:    conn,
		limiter: newLimiter(),
	}
	c.m = NewMux(conn)
	return c, nil
}

// A ClientOption is passed to NewClient for configuration.
type ClientOption interface {
	clientOption()
}

// Close closes the Client.
func (c *Client) Close() {
	// BUG(darkfeline): Client.Close should try to logout first.
	c.m.Close()
}

// A UserInfo contains user information for authentication and encryption.
type UserInfo struct {
	UserName     string
	UserPassword string
	APIKey       string // required for encryption, optional otherwise
}

// Encrypt calls the ENCRYPT command.
func (c *Client) Encrypt(ctx context.Context, u UserInfo) error {
	if u.APIKey == "" {
		return errors.New("udpapi: APIKey require for encryption")
	}
	v := url.Values{}
	v.Set("user", u.UserName)
	v.Set("type", "1")
	resp, err := c.m.Request(ctx, "ENCRYPT", v)
	if err != nil {
		return err
	}
	switch resp.Code {
	case 209:
		parts := strings.SplitN(resp.Header, " ", 2)
		salt := parts[0]
		sum := md5.Sum([]byte(u.APIKey + salt))
		b, err := aes.NewCipher(sum[:])
		if err != nil {
			return fmt.Errorf("udpapi: %s", err)
		}
		c.m.SetBlock(b)
		return nil
	default:
		return fmt.Errorf("udpapi: bad code %d %q", resp.Code, resp.Header)
	}
}

// Auth calls the AUTH command.
func (c *Client) Auth(ctx context.Context, u UserInfo) error {
	v := url.Values{}
	v.Set("user", u.UserName)
	v.Set("pass", u.UserPassword)
	v.Set("protover", protoVer)
	v.Set("client", c.ClientName)
	v.Set("clientver", strconv.Itoa(int(c.ClientVersion)))
	v.Set("nat", "1") // TODO use this
	v.Set("comp", "1")
	resp, err := c.m.Request(ctx, "AUTH", v)
	if err != nil {
		return err
	}
	switch resp.Code {
	case 201:
		// TODO New anidb UDP API version available
		fallthrough
	case 200:
		parts := strings.SplitN(resp.Header, " ", 3)
		if len(parts) < 3 {
			return fmt.Errorf("udpapi: invalid response header %q", resp.Header)
		}
		c.sessionKeyMu.Lock()
		c.sessionKey = parts[0]
		c.sessionKeyMu.Unlock()
		// TODO Make address comparison more reliable
		if our := c.conn.LocalAddr().String(); our != parts[1] {
			// TODO Detected NAT
		}
		return nil
	default:
		return fmt.Errorf("udpapi: bad code %d %q", resp.Code, resp.Header)
	}
}

// Logout calls the LOGOUT command.
func (c *Client) Logout(ctx context.Context) error {
	v, err := c.sessionValues()
	if err != nil {
		return err
	}
	resp, err := c.m.Request(ctx, "LOGOUT", v)
	if err != nil {
		return err
	}
	c.m.SetBlock(nil)
	c.sessionKeyMu.Lock()
	c.sessionKey = ""
	c.sessionKeyMu.Unlock()
	switch resp.Code {
	case 203:
		return nil
	default:
		return fmt.Errorf("udpapi: bad code %d %q", resp.Code, resp.Header)
	}
}

// FileByHash calls the FILE command by size+ed2k hash.
func (c *Client) FileByHash(ctx context.Context, size int64, hash string, fmask FMask, amask FAMask) ([]string, error) {
	v, err := c.sessionValues()
	if err != nil {
		return nil, err
	}
	v.Set("size", fmt.Sprintf("%d", size))
	v.Set("ed2k", hash)
	v.Set("fmask", fmt.Sprintf("%x", fmask))
	v.Set("amask", fmt.Sprintf("%x", amask))
	resp, err := c.m.Request(ctx, "FILE", v)
	if err != nil {
		return nil, err
	}
	if resp.Code != 220 {
		return nil, fmt.Errorf("udpapi: FileByHash got bad return code %s", resp.Code)
	}
	if n := len(resp.Rows); n != 1 {
		return nil, fmt.Errorf("udpapi: FileByHash got unexpected number of rows %d", n)
	}
	return resp.Rows[0], nil
}

func (c *Client) sessionValues() (url.Values, error) {
	v := url.Values{}
	c.sessionKeyMu.Lock()
	defer c.sessionKeyMu.Unlock()
	if c.sessionKey == "" {
		return nil, errors.New("udpapi: no session key (auth first)")
	}
	v.Set("s", c.sessionKey)
	return v, nil
}
