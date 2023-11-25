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

import (
	"fmt"
	"testing"
)

func TestPrefixLogger(t *testing.T) {
	t.Parallel()
	var s spyLogger
	p := prefixLogger{
		prefix: "mika:",
		logger: &s,
	}
	p.Printf("%s %s", "azusa", "hifumi")
	got := s.msg
	const want = "mika:azusa hifumi"
	if got != want {
		t.Errorf("got log message %q; want %q", got, want)
	}
}

type spyLogger struct {
	msg string
}

func (l *spyLogger) Printf(format string, a ...any) {
	l.msg = fmt.Sprintf(format, a...)
}
