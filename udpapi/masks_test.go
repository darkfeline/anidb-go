// Copyright (C) 2023 Allen Li
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

package udpapi

import "testing"

func TestFileFmask_Test(t *testing.T) {
	t.Parallel()
	var m FileFmask
	m.Set("aid")
	want := FileFmask{0b100_0000, 0, 0, 0, 0}
	if m != want {
		t.Errorf("Got %v; want %v", m, want)
	}
}

func TestFileAmask_Test(t *testing.T) {
	t.Parallel()
	var m FileAmask
	m.Set("epno")
	want := FileAmask{0, 0, 0b1000_0000, 0}
	if m != want {
		t.Errorf("Got %v; want %v", m, want)
	}
}
