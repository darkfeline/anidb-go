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

type TitlesCache struct {
	Path   string
	Titles []AnimeT
}

func DefaultTitlesCache() (*TitlesCache, error) {
	return OpenTitlesCache(defaultTitlesCacheFile())
}

func OpenTitlesCache(path string) (*TitlesCache, error) {
	f, err := os.Open(path)
	if err != nil {
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

func (c *TitlesCache) GetTitles() ([]AnimeT, error) {
	if len(c.Titles) > 0 {
		return c.Titles, nil
	}
	return c.GetFreshTitles()
}

func (c *TitlesCache) GetFreshTitles() ([]AnimeT, error) {
	t, err := RequestTitles()
	if err != nil {
		return nil, err
	}
	c.Titles = t
	return t, nil
}

func (c *TitlesCache) Save() error {
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
	return nil
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
