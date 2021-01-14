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
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
)

// A TitlesCache represents a cache for AniDB titles data.
type TitlesCache struct {
	// Path is the path to the cache file.
	Path string
	// Titles is the titles loaded from the cache.
	Titles []AnimeT
	// Updated indicates if the cached titles were updated.
	// This is set to true when any method updates the cache.
	Updated bool
}

// DefaultTitlesCache opens a TitlesCache at a default location,
// using XDG_CACHE_DIR.
func DefaultTitlesCache() (*TitlesCache, error) {
	return OpenTitlesCache(defaultTitlesCacheFile())
}

// OpenTitlesCache opens a TitlesCache.
func OpenTitlesCache(path string) (*TitlesCache, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &TitlesCache{Path: path}, nil
		}
		return nil, fmt.Errorf("open titles cache: %s", err)
	}
	defer f.Close()
	c := &TitlesCache{
		Path: path,
	}
	if err := gob.NewDecoder(f).Decode(&c.Titles); err != nil {
		return nil, fmt.Errorf("open titles cache %s: %s", path, err)
	}
	return c, nil
}

// GetTitles gets titles from the cache.
// If the cache has not been populated yet, downloads titles from AniDB.
func (c *TitlesCache) GetTitles() ([]AnimeT, error) {
	if len(c.Titles) > 0 {
		return c.Titles, nil
	}
	return c.GetFreshTitles()
}

// GetFreshTitles downloads titles from AniDB and stores it in the cache.
// See AniDB API documentation about rate limits.
func (c *TitlesCache) GetFreshTitles() ([]AnimeT, error) {
	t, err := RequestTitles()
	if err != nil {
		return nil, err
	}
	c.Titles = t
	c.Updated = true
	return t, nil
}

// Save saves the cached titles to the cache file.
// This method sets Updated to false if successful.
func (c *TitlesCache) Save() error {
	if err := os.MkdirAll(filepath.Dir(c.Path), 0777); err != nil {
		return fmt.Errorf("save titles cache: %s", err)
	}
	f, err := os.Create(c.Path)
	if err != nil {
		return fmt.Errorf("save titles cache: %s", err)
	}
	defer f.Close()
	if err := gob.NewEncoder(f).Encode(c.Titles); err != nil {
		return fmt.Errorf("save titles cache %s: %s", c.Path, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("save titles cache: %s", err)
	}
	c.Updated = false
	return nil
}

// SaveIfUpdated saves the cached titles to the cache file if they
// have been updated.
// This method sets Updated to false if successful.
func (c *TitlesCache) SaveIfUpdated() error {
	if !c.Updated {
		return nil
	}
	return c.Save()
}

func defaultTitlesCacheFile() string {
	return filepath.Join(cacheDir(), "go.felesatra.moe_anidb", "titles.gob")
}

func cacheDir() string {
	if p := os.Getenv("XDG_CACHE_HOME"); p != "" {
		return p
	}
	return filepath.Join(os.Getenv("HOME"), ".cache")
}
