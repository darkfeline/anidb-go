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
	"context"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

// A Client describe the AniDB API client in use.
// Read the AniDB API documentation about registering a client.
type Client struct {
	Name    string
	Version int
	// Limiter specifies a rate limiter to use.
	// If unset, no rate limiting is done.
	Limiter Limiter
}

// A Limiter implements rate limiting.
// The golang.org/x/time/rate package provides an implementation.
type Limiter interface {
	Wait(context.Context) error
}

var httpClient = http.Client{}

func (c *Client) httpAPI(params map[string]string) ([]byte, error) {
	if c.Limiter != nil {
		if err := c.Limiter.Wait(context.Background()); err != nil {
			return nil, err
		}
	}
	u := c.apiRequestURL(params)
	resp, err := httpClient.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, err
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

func (c *Client) apiRequestURL(params map[string]string) string {
	vals := url.Values{}
	vals.Set("client", c.Name)
	vals.Set("clientver", strconv.Itoa(c.Version))
	vals.Set("protover", "1")
	for k, v := range params {
		vals.Set(k, v)
	}
	return "http://api.anidb.net:9001/httpapi?" + vals.Encode()
}

// RequestAnime requests anime information from AniDB.
func (c *Client) RequestAnime(aid int) (*Anime, error) {
	d, err := c.httpAPI(map[string]string{
		"request": "anime",
		"aid":     strconv.Itoa(aid),
	})
	if err != nil {
		return nil, fmt.Errorf("anidb request anime %d: %s", aid, err)
	}
	a, err := decodeAnime(d)
	if err != nil {
		return nil, fmt.Errorf("anidb request anime %d: %s", aid, err)
	}
	return a, nil
}

// RequestAnime requests anime information from AniDB.
// This is deprecated; use the Client.RequestAnime method instead.
func RequestAnime(c Client, aid int) (*Anime, error) {
	return c.RequestAnime(aid)
}

// An Anime holds information for an anime.
type Anime struct {
	AID          int       `xml:"id,attr"`
	Titles       []Title   `xml:"titles>title"`
	Type         string    `xml:"type"`
	EpisodeCount int       `xml:"episodecount"`
	StartDate    string    `xml:"startdate"`
	EndDate      string    `xml:"enddate"`
	Episodes     []Episode `xml:"episodes>episode"`
}

// A Title holds information for a single anime title.
type Title struct {
	Name string `xml:",chardata"`
	Type string `xml:"type,attr"`
	Lang string `xml:"http://www.w3.org/XML/1998/namespace lang,attr"`
}

// An Episode holds information for an episode.
type Episode struct {
	// EpNo is a concatenation of a type string and episode number.  It
	// should be unique among the episodes for an anime, so it can serve
	// as a unique identifier.
	EpNo string `xml:"epno"`
	// Length is the length of the episode in minutes.
	Length int       `xml:"length"`
	Titles []EpTitle `xml:"title"`
}

// An EpTitle holds information for a single episode title.
type EpTitle struct {
	Title string `xml:",chardata"`
	Lang  string `xml:"http://www.w3.org/XML/1998/namespace lang,attr"`
}

func decodeAnime(d []byte) (*Anime, error) {
	var r Anime
	if err := xml.Unmarshal(d, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

// checkAPIError checks for in-band AniDB API errors.
func checkAPIError(d []byte) error {
	var n xml.Name
	_ = xml.Unmarshal(d, &n)
	if n.Local != "error" {
		return nil
	}
	var a struct {
		Text string `xml:",innerxml"`
	}
	if err := xml.Unmarshal(d, &a); err != nil {
		// Unmarshaling should never fail.
		panic(err)
	}
	return fmt.Errorf("API error %s", a.Text)
}
