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

package framework

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression/function/aggregation"
)

// WindowFramerState holds the sql.WindowFramer setup shared by every native window-function implementation
// that respects an explicit frame clause rather than always operating over the whole partition - e.g. the
// native sum/avg aggregates and nth_value(). Bind a framer from the window's explicit frame clause if one
// was given; with no explicit frame, DefaultFramer supplies Postgres's default frame: RANGE UNBOUNDED
// PRECEDING TO CURRENT ROW when the window has an ORDER BY (so rows tied on the ORDER BY value share a
// peer group), otherwise the whole partition.
type WindowFramerState struct {
	framer sql.WindowFramer
	window *sql.WindowDefinition
}

// BindFramer builds and stores this window's framer, if it declared an explicit frame clause; with no
// explicit frame, DefaultFramer's fallback applies instead.
func (s *WindowFramerState) BindFramer(window *sql.WindowDefinition) error {
	s.window = window
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
func (s *WindowFramerState) StartPartition(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) error {
	return nil
}

// DefaultFramer implements the sql.WindowFunction interface; with no explicit frame, this supplies
// Postgres's default frame: RANGE UNBOUNDED PRECEDING TO CURRENT ROW when the window has an ORDER BY (so
// rows tied on the ORDER BY value share a peer group), otherwise the whole partition.
func (s *WindowFramerState) DefaultFramer() sql.WindowFramer {
	if s.framer != nil {
		return s.framer
	}
	if s.window == nil || len(s.window.OrderBy) < 1 {
		return aggregation.NewPartitionFramer()
	}
	framer, _ := aggregation.NewRangeUnboundedPrecedingToCurrentRowFramer(nil, s.window)
	return framer
}

// Dispose implements the sql.WindowFunction interface.
func (s *WindowFramerState) Dispose(ctx *sql.Context) {}
