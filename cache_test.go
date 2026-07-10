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
	"bytes"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"
)

func TestTitlesCache(t *testing.T) {
	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "titles.gob")
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
		Path:   cachePath,
		Titles: ts,
	}
	if err := c.Save(); err != nil {
		t.Fatalf("Error saving: %s", err)
	}
	loadedCache, err := OpenTitlesCache(cachePath)
	if err != nil {
		t.Errorf("Error loading: %s", err)
	}
	if !reflect.DeepEqual(loadedCache.Titles, ts) {
		t.Errorf("got %#v; want %#v", loadedCache.Titles, ts)
	}
}

func TestTitlesCache_GetFreshTitles(t *testing.T) {
	gzippedBytes := fakeTitlesHTTPClient(t)

	tempDir := t.TempDir()
	cachePath := filepath.Join(tempDir, "cache", "titles.gob")
	c := &TitlesCache{
		Path: cachePath,
	}

	titles, err := c.GetFreshTitles()
	if err != nil {
		t.Fatalf("GetFreshTitles failed: %s", err)
	}

	// Verify that the cached titles match the expected test data
	expectedTitles := []AnimeT{{AID: 22, Titles: []Title{
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
	if !reflect.DeepEqual(titles, expectedTitles) {
		t.Errorf("GetFreshTitles returned %#v; want %#v", titles, expectedTitles)
	}

	// Verify the copy of anime-titles.xml.gz is saved alongside the cache file
	gzPath := filepath.Join(tempDir, "cache", "anime-titles.xml.gz")
	savedGz, err := ioutil.ReadFile(gzPath)
	if err != nil {
		t.Fatalf("Failed to read saved gz file: %s", err)
	}

	if !bytes.Equal(savedGz, gzippedBytes) {
		t.Errorf("Saved gz content does not match original gzipped bytes")
	}
}
