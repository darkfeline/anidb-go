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
	"io/ioutil"
	"reflect"
	"testing"
)

func TestDecodeAnime(t *testing.T) {
	d, err := ioutil.ReadFile("testdata/anime.xml")
	if err != nil {
		t.Fatalf("Error reading test data file: %+v", err)
	}
	a, err := decodeAnime(d)
	if err != nil {
		t.Errorf("Error decoding titles: %+v", err)
	}
	e := []Episode{
		{
			EID:    113,
			EpNo:   "1",
			Length: 25,
			Titles: []EpTitle{
				{Title: "使徒, 襲来", Lang: "ja"},
				{Title: "Angel Attack!", Lang: "en"},
				{Title: "Shito, Shuurai", Lang: "x-jat"},
			},
		},
		{
			EID:    28864,
			EpNo:   "S1",
			Length: 75,
			Titles: []EpTitle{
				{Title: "Revival of Evangelion Extras Disc", Lang: "en"},
			},
		},
	}
	exp := &Anime{
		AID:          22,
		Type:         "TV Series",
		EpisodeCount: 26,
		StartDate:    "1995-10-04",
		EndDate:      "1996-03-27",
		Titles: []Title{
			{Name: "Shinseiki Evangelion", Type: "main", Lang: "x-jat"},
			{Name: "Neon Genesis Evangelion", Type: "official", Lang: "en"},
		},
		Episodes: e,
	}
	if !reflect.DeepEqual(a, exp) {
		t.Errorf("Expected %#v, got %#v", exp, a)
	}
}

func TestCheckAPIError(t *testing.T) {
	d, err := ioutil.ReadFile("testdata/error.xml")
	if err != nil {
		t.Fatalf("Error reading test data file: %+v", err)
	}
	err = checkAPIError(d)
	if err == nil {
		t.Errorf("Did not get error")
	}
}

func TestCheckAPIErrorGood(t *testing.T) {
	d, err := ioutil.ReadFile("testdata/anime.xml")
	if err != nil {
		t.Fatalf("Error reading test data file: %+v", err)
	}
	err = checkAPIError(d)
	if err != nil {
		t.Errorf("Got unexpected error %+v", err)
	}
}
