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
	"bytes"
	"compress/gzip"
	"io/ioutil"
	"net/http"
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

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// fakeTitlesHTTPClient sets up a fake httpClient that serves gzipped AniDB titles data
// loaded from testdata/titles.xml. It returns the raw gzipped bytes for validation.
func fakeTitlesHTTPClient(t *testing.T) []byte {
	// Read original xml data
	xmlData, err := ioutil.ReadFile("testdata/titles.xml")
	if err != nil {
		t.Fatalf("Failed to read testdata/titles.xml: %s", err)
	}

	// Gzip the xml data
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(xmlData); err != nil {
		t.Fatalf("Failed to gzip XML: %s", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("Failed to close gzip writer: %s", err)
	}
	gzippedBytes := buf.Bytes()

	origClient := httpClient
	t.Cleanup(func() { httpClient = origClient })

	httpClient = http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.String() != "http://anidb.net/api/anime-titles.xml.gz" {
				t.Errorf("Unexpected URL: %s", req.URL.String())
			}
			if req.Header.Get("User-Agent") == "" {
				t.Error("Missing User-Agent header")
			}
			return &http.Response{
				StatusCode: 200,
				Body:       ioutil.NopCloser(bytes.NewReader(gzippedBytes)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	return gzippedBytes
}
