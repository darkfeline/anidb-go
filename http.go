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
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
)

// Client describe the AniDB API client in use.  Read the AniDB API
// documentation about registering a client.
type Client struct {
	Name    string
	Version int
}

func httpAPI(c Client, params map[string]string) ([]byte, error) {
	v := url.Values{}
	v.Set("client", c.Name)
	v.Set("clientver", strconv.Itoa(c.Version))
	v.Set("protover", "1")
	u, err := url.Parse("http://api.anidb.net:9001/httpapi")
	if err != nil {
		panic(err)
	}
	u.RawQuery = v.Encode()
	return httpGet(u.String())
}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.Errorf("Bad status %d", resp.StatusCode)
	}
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := checkAPIError(d); err != nil {
		return nil, err
	}
	return d, nil
}

func checkAPIError(d []byte) error {
	var n xml.Name
	_ = xml.Unmarshal(d, &n)
	if n.Local != "error" {
		return nil
	}
	var a struct {
		Text string `xml:",innerxml"`
	}
	_ = xml.Unmarshal(d, &a)
	return fmt.Errorf("API error: %s", a.Text)
}
