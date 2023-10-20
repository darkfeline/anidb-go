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

// A BitSpec designates a bit in an API mask.
type BitSpec struct{ byte, bit int, typ string }

// A FileFmask is a mask for the FILE command fmask field.
type FileFmask [5]byte

// FileFmaskFields describes the bit fields in a FILE fmask.
var FileFmaskFields = map[string]BitSpec{
	"aid":   {0, 6, "int4"},
	"eid":   {0, 5, "int4"},
	"gid":   {0, 4, "int4"},
	"state": {0, 0, "int2"},

	"anidb file name": {3, 0, "str"},
}

// Set sets a bit in the mask.
func (m *FileFmask) Set(f string) {
	s, ok := FileFmaskFields[f]
	if !ok {
		panic(f)
	}
	m[s.byte] |= 1 << s.bit
}

// A FileAmask is a mask for the FILE command amask field.
type FileAmask [4]byte

// FileAmaskFields describes the bit fields in a FILE amask.
var FileAmaskFields = map[string]BitSpec{
	"epno":    {2, 7, "str"},
	"ep name": {2, 6, "str"},
}

// Set sets a bit in the mask.
func (m *FileAmask) Set(f string) {
	s, ok := FileAmaskFields[f]
	if !ok {
		panic(f)
	}
	m[s.byte] |= 1 << s.bit
}

func formatMask(m []byte) string {
	var sb strings.Builder
	for _, b := range m {
		fmt.Fprintf(&sb, "%x", b)
	}
	return sb.String()
}
