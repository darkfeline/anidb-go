// Copyright (C) 2020 Allen Li
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

package anidb_test

import (
	"fmt"
	"strings"

	"go.felesatra.moe/anidb"
)

func ExampleClient() {
	c := anidb.Client{
		Name:    "go.felesatra.moe/anidb example",
		Version: 1,
	}
	a, err := c.RequestAnime(8076)
	if err != nil {
		panic(err)
	}
	fmt.Print(a.EpisodeCount)
}

func ExampleTitlesCache() {
	c, err := anidb.DefaultTitlesCache()
	if err != nil {
		panic(err)
	}
	defer c.SaveIfUpdated()
	titles, err := c.GetTitles()
	if err != nil {
		panic(err)
	}
	var matched []anidb.AnimeT
	for _, anime := range titles {
		for _, t := range anime.Titles {
			if strings.Index(t.Name, "bofuri") >= 0 {
				matched = append(matched)
			}
		}
	}
}
