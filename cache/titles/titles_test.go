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

package titles

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"go.felesatra.moe/anidb"
)

func TestSaveAndLoad(t *testing.T) {
	d, err := ioutil.TempDir("", "test")
	if err != nil {
		t.Fatalf("Error creating temporary dir: %s", err)
	}
	defer os.RemoveAll(d)
	a := []anidb.AnimeT{{AID: 22, Titles: []anidb.Title{
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
	f := filepath.Join(d, "foo.gob")
	err = Save(f, a)
	if err != nil {
		t.Fatalf("Error saving: %s", err)
	}
	got, err := Load(f)
	if err != nil {
		t.Errorf("Error loading: %s", err)
	}
	if !reflect.DeepEqual(got, a) {
		t.Errorf("Expected %#v, got %#v", a, got)
	}
}
