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
	"io"
	"net/http"

	"github.com/pkg/errors"
)

type Client struct {
	Name    string
	Version int
}

type Anime struct {
	AID    int     `xml:"aid,attr"`
	Titles []Title `xml:"title"`
}

type Title struct {
	Name string `xml:",innerxml"`
	Type string `xml:"type,attr"`
	Lang string `xml:"http://www.w3.org/XML/1998/namespace lang,attr"`
}

func RequestTitles() ([]Anime, error) {
	resp, err := http.Get("http://anidb.net/api/anime-titles.xml.gz")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, errors.Errorf("Bad status %d", resp.StatusCode)
	}
	return decodeTitles(resp.Body)
}

func decodeTitles(r io.Reader) ([]Anime, error) {
	d := xml.NewDecoder(r)
	var a struct {
		Anime []Anime `xml:"anime"`
	}
	err := d.Decode(&a)
	if err != nil {
		return nil, err
	}
	return a.Anime, nil
}
