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
	"strconv"
)

// RequestAnime requests anime information from AniDB.
func (c *Client) RequestAnime(aid int) (*Anime, error) {
	d, err := httpAPI(*c, map[string]string{
		"request": "anime",
		"aid":     strconv.Itoa(aid),
	})
	if err != nil {
		return nil, fmt.Errorf("anidb: request anime %d: %s", aid, err)
	}
	a, err := decodeAnime(d)
	if err != nil {
		return nil, fmt.Errorf("anidb: request anime %d: %s", aid, err)
	}
	return a, nil
}

// RequestAnime requests anime information from AniDB.
// This is deprecated; use the Client.RequestAnime method instead.
func RequestAnime(c Client, aid int) (*Anime, error) {
	return c.RequestAnime(aid)
}

func decodeAnime(d []byte) (*Anime, error) {
	var r Anime
	if err := xml.Unmarshal(d, &r); err != nil {
		return nil, err
	}
	return &r, nil
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
