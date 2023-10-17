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

type bitSpec struct{ byte, bit int }

// A FileFmask is a mask for the FILE command fmask field.
type FileFmask [5]byte

var fileFmaskFields = map[string]bitSpec{
	"aid":   {0, 6},
	"eid":   {0, 5},
	"gid":   {0, 4},
	"state": {0, 0},

	"anidb file name": {3, 0},
}

func (m *FileFmask) Set(f string) {
	s, ok := fileFmaskFields[f]
	if !ok {
		panic(f)
	}
	m[s.byte] |= 1 << s.bit
}

// A FileAmask is a mask for the FILE command amask field.
type FileAmask [4]byte

func formatMask(m []byte) string {
	var sb strings.Builder
	for _, b := range m {
		fmt.Fprintf(&sb, "%x", b)
	}
	return sb.String()
}
