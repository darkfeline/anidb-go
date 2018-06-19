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

type Anime struct {
	AID             int       `xml:"aid,attr"`
	Titles          []Title   `xml:"titles>title"`
	Type            string    `xml:"type"`
	EpisodeCode     int       `xml:"episodecount"`
	StartDateString string    `xml:"startdate"`
	EndDateString   string    `xml:"enddate"`
	Episodes        []Episode `xml:"episodes>episode"`
}

type AnimeT struct {
	AID    int     `xml:"aid,attr"`
	Titles []Title `xml:"title"`
}

type Title struct {
	Name string `xml:",innerxml"`
	Type string `xml:"type,attr"`
	Lang string `xml:"http://www.w3.org/XML/1998/namespace lang,attr"`
}

// Episode holds information for an episode.
//
// EpNo is a concatenation of a type string and episode number.  It
// should be unique among the episodes for an anime, so it can serve
// as a unique identifier.
//
// Type is the episode type code.
//
// Length is the length of the episode in minutes.
//
// Title is the episode title.
type Episode struct {
	EpNo   string
	Type   int
	Length int
	Title  EpisodeTitle
}

type EpisodeTitle struct {
	Title, Lang string
}
