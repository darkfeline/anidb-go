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
	"compress/gzip"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

// RequestTitles requests title information from AniDB.
//
// TitlesCache is more convenient to use, as AniDB has severe rate
// limits on this.
func RequestTitles() ([]AnimeT, error) {
	d, err := downloadTitles()
	if err != nil {
		return nil, fmt.Errorf("anidb request titles: %s", err)
	}
	ts, err := DecodeTitles(d)
	if err != nil {
		return nil, fmt.Errorf("anidb request titles: %s", err)
	}
	return ts, nil
}

const (
	packageVersion = "1.1.0"
	userAgent      = "go.felesatra.moe/anidb " + packageVersion
)

const titlesURL = "http://anidb.net/api/anime-titles.xml.gz"

func downloadTitles() ([]byte, error) {
	req, err := http.NewRequest("GET", titlesURL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("User-Agent", userAgent)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, err
	}
	r, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	d, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// DecodeTitles decodes XML title information from an AniDB title dump.
// The input should be uncompressed XML.
func DecodeTitles(d []byte) ([]AnimeT, error) {
	var r struct {
		Anime []AnimeT `xml:"anime"`
	}
	if err := xml.Unmarshal(d, &r); err != nil {
		return nil, fmt.Errorf("anidb decode titles: %s", err)
	}
	return r.Anime, nil
}

// An AnimeT is like Anime but holds title information only.
type AnimeT struct {
	AID    int     `xml:"aid,attr"`
	Titles []Title `xml:"title"`
}
