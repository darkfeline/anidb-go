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

// Package codes contains return codes for the AniDB UDP API
package codes

// A ReturnCode is an AniDB UDP API return code.
// Note that even though ReturnCode implements error, not all
// ReturnCode values should be considered errors.
type ReturnCode int

const (
	LoginFirst     ReturnCode = 501 // 501 LOGIN_FIRST
	AccessDenied   ReturnCode = 502 // 502 ACCESS_DENIED
	ClientBanned   ReturnCode = 504 // 504 CLIENT_BANNED
	IllegalInput   ReturnCode = 505 // 505 ILLEGAL_INPUT_OR_ACCESS_DENIED
	InvalidSession ReturnCode = 506 // 506 INVALID_SESSION
	Banned         ReturnCode = 555 // 555 BANNED
	UnknownCmd     ReturnCode = 598 // 598 UNKNOWN_COMMAND
	InternalErr    ReturnCode = 600 // 600 INTERNAL_SERVER_ERROR
	OutOfService   ReturnCode = 601 // 601 ANIDB_OUT_OF_SERVICE
	ServerBusy     ReturnCode = 602 // 602 SERVER_BUSY
	Timeout        ReturnCode = 604 // 604 TIMEOUT - DELAY AND RESUBMIT
)

//go:generate stringer -type=ReturnCode -linecomment

func (c ReturnCode) Error() string {
	return c.String()
}
