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

// Package titles provides a cache for AniDB titles data.
//
// This package is deprecated as [go.felesatra.moe/anidb] now provides a
// titles cache.
package titles

import (
	"encoding/gob"
	"os"
	"path/filepath"

	"go.felesatra.moe/anidb"
)

// Load loads cached anime title data.
func Load(path string) ([]anidb.AnimeT, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	d := gob.NewDecoder(f)
	var a []anidb.AnimeT
	err = d.Decode(&a)
	if err != nil {
		return nil, err
	}
	return a, nil
}

var titlesPath string

func init() {
	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		cacheDir = filepath.Join(os.Getenv("HOME"), ".cache")
	}
	titlesPath = filepath.Join(cacheDir, "go.felesatra.moe_anidb", "titles.gob")
}

// LoadDefault loads cached anime title data from a default cache path.
func LoadDefault() ([]anidb.AnimeT, error) {
	return Load(titlesPath)
}

// Save saves anime title data to a cache.
func Save(path string, a []anidb.AnimeT) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	e := gob.NewEncoder(f)
	return e.Encode(a)
}

// SaveDefault saves anime title data to a default cache path.
func SaveDefault(a []anidb.AnimeT) error {
	err := os.MkdirAll(filepath.Dir(titlesPath), 0777)
	if err != nil {
		return err
	}
	return Save(titlesPath, a)
}
