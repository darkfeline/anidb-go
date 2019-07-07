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

func TestDecodeTitles(t *testing.T) {
	d, err := ioutil.ReadFile("testdata/titles.xml")
	if err != nil {
		t.Fatalf("Error reading test data file: %+v", err)
	}
	a, err := DecodeTitles(d)
	if err != nil {
		t.Errorf("Error decoding titles: %+v", err)
	}
	exp := []AnimeT{{AID: 22, Titles: []Title{
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
	if !reflect.DeepEqual(a, exp) {
		t.Errorf("DecodeTitles(%#v) = %#v, expected %#v", d, a, exp)
	}
}
