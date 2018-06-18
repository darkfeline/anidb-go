// Copyright (C) 2018 Allen Li
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
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

type Client struct {
	Name    string
	Version int
}

func httpAPI(c Client, params map[string]string) (io.ReadCloser, error) {
	v := url.Values{}
	v.Set("client", c.Name)
	v.Set("clientver", strconv.Itoa(c.Version))
	v.Set("protover", "1")
	u, err := url.Parse("http://api.anidb.net:9001/httpapi")
	if err != nil {
		panic(err)
	}
	u.RawQuery = v.Encode()
	r, err := httpGet(u.String())
	if err != nil {
		return nil, err
	}
	return r, nil
}

func httpGet(url string) (io.ReadCloser, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, errors.Errorf("Bad status %d", resp.StatusCode)
	}
	return resp.Body, nil
}
