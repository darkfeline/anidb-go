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

package anidb

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestTitlesCache(t *testing.T) {
	f, err := ioutil.TempFile("", "test")
	if err != nil {
		t.Fatalf("Error creating temporary dir: %s", err)
	}
	t.Cleanup(func() { os.Remove(f.Name()) })
	f.Close()
	ts := []AnimeT{{AID: 22, Titles: []Title{
		{
			Name: "Neon Genesis Evangelion",
			Type: "official",
			Lang: "en",
		},
		{
			Name: "Shinseiki Evangelion",
			Type: "main",
			Lang: "x-jat",
		},
	}}}
	c := &TitlesCache{
		Path:   f.Name(),
		Titles: ts,
	}
	if err := c.Save(); err != nil {
		t.Fatalf("Error saving: %s", err)
	}
	c, err = OpenTitlesCache(f.Name())
	if err != nil {
		t.Errorf("Error loading: %s", err)
	}
	if !reflect.DeepEqual(c.Titles, ts) {
		t.Errorf("got %#v; want %#v", c.Titles, ts)
	}
}
