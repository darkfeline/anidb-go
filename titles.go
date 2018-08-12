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
)

// RequestTitles requests title information from AniDB.
func RequestTitles() ([]AnimeT, error) {
	d, err := httpGet("http://anidb.net/api/anime-titles.xml.gz")
	if err != nil {
		return nil, fmt.Errorf("anidb: titles request error: %s", err)
	}
	ts, err := decodeTitles(d)
	if err != nil {
		return nil, fmt.Errorf("anidb: decode titles error: %s", err)
	}
	return ts, nil
}

func decodeTitles(d []byte) ([]AnimeT, error) {
	var r struct {
		Anime []AnimeT `xml:"anime"`
	}
	if err := xml.Unmarshal(d, &r); err != nil {
		return nil, err
	}
	return r.Anime, nil
}
