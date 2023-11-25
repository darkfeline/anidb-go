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

// A Logger can be used for logging.
// A Logger must be safe to use concurrently.
type Logger interface {
	Printf(string, ...any)
}

type nullLogger struct{}

func (nullLogger) Printf(string, ...any) {}

type prefixLogger struct {
	prefix string
	logger Logger
}

func (l prefixLogger) Printf(format string, a ...any) {
	a2 := make([]any, len(a)+1)
	a2[0] = l.prefix
	copy(a2[1:], a)
	l.logger.Printf("%s"+format, a2...)
}
