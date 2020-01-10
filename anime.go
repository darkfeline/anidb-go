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
func RequestAnime(c Client, aid int) (*Anime, error) {
	d, err := httpAPI(c, map[string]string{
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

func decodeAnime(d []byte) (*Anime, error) {
	var r Anime
	if err := xml.Unmarshal(d, &r); err != nil {
		return nil, err
	}
	return &r, nil
}
