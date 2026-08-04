// Copyright 2026 Dolthub, Inc.
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

package window

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression/function/aggregation"
)

// windowFramerState holds the sql.WindowFramer setup for a window function that respects an explicit
// frame clause rather than always operating over the whole partition or peer group - e.g. nth_value().
type windowFramerState struct {
	framer sql.WindowFramer
}

// bindFramer builds and stores this window's framer, if it declared an explicit frame clause; with no
// explicit frame, DefaultFramer's fallback applies instead.
func (s *windowFramerState) bindFramer(window *sql.WindowDefinition) error {
	if window == nil || window.Frame == nil {
		return nil
	}
	framer, err := window.Frame.NewFramer(window)
	if err != nil {
		return err
	}
	s.framer = framer
	return nil
}

// StartPartition implements the sql.WindowFunction interface.
func (s *windowFramerState) StartPartition(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) error {
	return nil
}

// DefaultFramer implements the sql.WindowFunction interface; with no explicit frame, this supplies
// Postgres's default (unbounded preceding to current row).
func (s *windowFramerState) DefaultFramer() sql.WindowFramer {
	if s.framer != nil {
		return s.framer
	}
	return aggregation.NewUnboundedPrecedingToCurrentRowFramer()
}

// Dispose implements the sql.WindowFunction interface.
func (s *windowFramerState) Dispose(ctx *sql.Context) {}
