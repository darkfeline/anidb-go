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

// An FMask is a mask for the FILE command fmask field.
type FMask uint64

const (
	// Bit	Dec	Data Field
	// 0	1	unused
	_ FMask = 1 << iota
	// 1	2	str mylist other
	_
	// 2	4	str mylist source
	_
	// 3	8	str mylist storage
	_
	// 4	16	int4 mylist viewdate
	_
	// 5	32	int4 mylist viewed
	_
	// 6	64	int4 mylist filestate
	_
	// 7	128	int4 mylist state
	_
	// Bit	Dec	Data Field
	// 0	1	str anidb file name
	FMaskFileName
	// 1	2	unused
	_
	// 2	4	unused
	_
	// 3	8	int4 aired date
	_
	// 4	16	str description
	_
	// 5	32	int4 length in seconds
	_
	// 6	64	str sub language
	_
	// 7	128	str dub language
	_
	// Bit	Dec	Data Field
	// 0	1	str file type (extension)
	FMaskFileType
	// 1	2	str video resolution
	_
	// 2	4	int4 video bitrate
	_
	// 3	8	str video codec
	_
	// 4	16	int4 audio bitrate list
	_
	// 5	32	str audio codec list
	_
	// 6	64	str source
	_
	// 7	128	str quality
	_
	// Bit	Dec	Data Field
	// 0	1	reserved
	_
	// 1	2	video colour depth
	_
	// 2	4	unused
	_
	// 3	8	str crc32
	FMaskCRC32
	// 4	16	str sha1
	FMaskSHA1
	// 5	32	str md5
	FMaskMD5
	// 6	64	str ed2k
	FMaskED2k
	// 7	128	int8 size
	FMaskSize
	// Bit	Dec	Data Field
	// 0	1	int2 state
	FMaskState
	// 1	2	int2 IsDeprecated
	_
	// 2	4	list other episodes
	_
	// 3	8	int4 mylist id
	_
	// 4	16	int4 gid
	FMaskGID
	// 5	32	int4 eid
	FMaskEID
	// 6	64	int4 aid
	FMaskAID
	// 7	128	unused
	_
)

// An FAMask is a mask for the FILE command amask field.
type FAMask uint32

const (
	// Bit	Dec	Data Field
	// 0	1	int4 date aid record updated
	_ FAMask = 1 << iota
	// 1	2	unused
	_
	// 2	4	unused
	_
	// 3	8	unused
	_
	// 4	16	unused
	_
	// 5	32	unused
	_
	// 6	64	str group short name
	_
	// 7	128	str group name
	FAMaskGroupName
	// Bit	Dec	Data Field
	// 0	1	unused
	_
	// 1	2	unused
	_
	// 2	4	int4 episode vote count
	_
	// 3	8	int4 episode rating
	_
	// 4	16	str ep kanji name
	_
	// 5	32	str ep romaji name
	_
	// 6	64	str ep name
	_
	// 7	128	str epno
	FAMaskEpno
	// Bit	Dec	Data Field
	// 0	1	retired
	_
	// 1	2	retired
	_
	// 2	4	str synonym list
	_
	// 3	8	str short name list
	_
	// 4	16	str other name
	_
	// 5	32	str english name
	_
	// 6	64	str kanji name
	_
	// 7	128	str romaji name
	_
	// Bit	Dec	Data Field
	// 0	1	reserved
	_
	// 1	2	str category list
	_
	// 2	4	str related aid type
	_
	// 3	8	str related aid list
	_
	// 4	16	str type
	FAMaskType
	// 5	32	str year
	_
	// 6	64	int4 highest episode number
	_
	// 7	128	int4 anime total episodes
	_
)
