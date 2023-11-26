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
	"context"
	"log/slog"
)

type nullHandler struct{}

func (nullHandler) Enabled(context.Context, slog.Level) bool {
	return false
}

func (nullHandler) Handle(context.Context, slog.Record) error {
	return nil
}

func (h nullHandler) WithAttrs([]slog.Attr) slog.Handler {
	return h
}

func (h nullHandler) WithGroup(string) slog.Handler {
	return h
}
