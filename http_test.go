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
	"testing"
)

func TestCheckAPIError(t *testing.T) {
	d, err := ioutil.ReadFile("testdata/error.xml")
	if err != nil {
		t.Fatalf("Error reading test data file: %+v", err)
	}
	err = checkAPIError(d)
	if err == nil {
		t.Fatalf("Did not get error")
	}
	exp := "API error: Banned"
	if err.Error() != exp {
		t.Errorf("err.Error() = %#v, expected %#v", err.Error(), exp)
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
