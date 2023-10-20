// Copyright (C) 2021 Allen Li
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

import (
	"fmt"
	"strings"
)

// A bitSpec designates a bit in an API mask.
type bitSpec struct {
	byte int
	bit  int
	typ  string
	name string
}

// A FileFmask is a mask for the FILE command fmask field.
type FileFmask [5]byte

// FileFmaskFields describes the bit fields in a FILE fmask.
var FileFmaskFields = map[string]bitSpec{
	"aid":   {0, 6, "int4", "aid"},
	"eid":   {0, 5, "int4", "eid"},
	"gid":   {0, 4, "int4", "gid"},
	"state": {0, 0, "int2", "state"},

	"anidb file name": {3, 0, "str", "anidb file name"},
}

// Set sets a bit in the mask.
func (m *FileFmask) Set(f ...string) {
	for _, f := range f {
		setMaskBit(m[:], FileFmaskFields, f)
	}
}

// A FileAmask is a mask for the FILE command amask field.
type FileAmask [4]byte

// FileAmaskFields describes the bit fields in a FILE amask.
var FileAmaskFields = map[string]bitSpec{
	"epno":    {2, 7, "str", "epno"},
	"ep name": {2, 6, "str", "ep name"},
}

// Set sets a bit in the mask.
func (m *FileAmask) Set(f ...string) {
	for _, f := range f {
		setMaskBit(m[:], FileFmaskFields, f)
	}
}

func setMaskBit(b []byte, m map[string]bitSpec, name string) {
	s, ok := m[name]
	if !ok {
		panic(name)
	}
	b[s.byte] |= 1 << s.bit
}

func formatMask(m []byte) string {
	var sb strings.Builder
	for _, b := range m {
		fmt.Fprintf(&sb, "%x", b)
	}
	return sb.String()
}
