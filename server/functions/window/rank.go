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

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// rank represents the PostgreSQL rank() window function, returning a bigint directly rather than GMS's
// native uint64 representation.
var rank = framework.Func0Window{
	Function0: framework.Function0{
		Name:   "rank",
		Return: pgtypes.Int64,
	},
	NewWinFunc: newRankWindowFunction,
}

// denseRank represents the PostgreSQL dense_rank() window function, returning a bigint directly rather than
// GMS's native uint64 representation.
var denseRank = framework.Func0Window{
	Function0: framework.Function0{
		Name:   "dense_rank",
		Return: pgtypes.Int64,
	},
	NewWinFunc: newDenseRankWindowFunction,
}

// percentRank represents the PostgreSQL percent_rank() window function.
var percentRank = framework.Func0Window{
	Function0: framework.Function0{
		Name:   "percent_rank",
		Return: pgtypes.Float64,
	},
	NewWinFunc: newPercentRankWindowFunction,
}

// rankWindowFunction is the sql.WindowFunction used for rank() within an OVER(...) clause. Its algorithm
// mirrors GMS's own aggregation.rankBase.Compute exactly (see sql/expression/function/aggregation/
// window_functions.go), just returning int64 instead of uint64: the rank of a row is the count of rows
// strictly before its peer group (rows sharing the same ORDER BY key), plus 1. Per the SQL standard, RANK
// ignores any explicit frame clause and groups rows into "peers" by the window's ORDER BY expressions, so
// DefaultFramer always returns a peer-group framer regardless of window.Frame.
type rankWindowFunction struct {
	orderBy                      []sql.Expression
	partitionStart, partitionEnd int
	pos                          int
}

var _ sql.WindowFunction = (*rankWindowFunction)(nil)

// newRankWindowFunction creates the sql.WindowFunction for rank().
func newRankWindowFunction(window *sql.WindowDefinition) (sql.WindowFunction, error) {
	var orderBy []sql.Expression
	if window != nil {
		orderBy = window.OrderBy.ToExpressions()
	}
	return &rankWindowFunction{partitionStart: -1, partitionEnd: -1, pos: -1, orderBy: orderBy}, nil
}

// StartPartition implements the sql.WindowFunction interface.
func (w *rankWindowFunction) StartPartition(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) error {
	w.partitionStart, w.partitionEnd = interval.Start, interval.End
	w.pos = w.partitionStart
	return nil
}

// DefaultFramer implements the sql.WindowFunction interface.
func (w *rankWindowFunction) DefaultFramer() sql.WindowFramer {
	return aggregation.NewPeerGroupFramer(w.orderBy)
}

// Compute implements the sql.WindowFunction interface.
func (w *rankWindowFunction) Compute(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) (interface{}, error) {
	if interval.End-interval.Start < 1 {
		return nil, nil
	}
	defer func() { w.pos++ }()
	switch {
	case w.pos == 0:
		return int64(1), nil
	case w.partitionEnd-w.partitionStart == 1:
		return int64(1), nil
	default:
		return int64(interval.Start-w.partitionStart) + 1, nil
	}
}

// Dispose implements the sql.WindowFunction interface.
func (w *rankWindowFunction) Dispose(ctx *sql.Context) {}

// denseRankWindowFunction is the sql.WindowFunction used for dense_rank() within an OVER(...) clause. It
// wraps rankWindowFunction's peer-group counting the same way GMS's own DenseRank wraps rankBase: instead of
// counting all preceding rows, it counts preceding distinct peer groups.
type denseRankWindowFunction struct {
	*rankWindowFunction
	prevRank  int64
	denseRank int64
}

var _ sql.WindowFunction = (*denseRankWindowFunction)(nil)

// newDenseRankWindowFunction creates the sql.WindowFunction for dense_rank().
func newDenseRankWindowFunction(window *sql.WindowDefinition) (sql.WindowFunction, error) {
	r, err := newRankWindowFunction(window)
	if err != nil {
		return nil, err
	}
	return &denseRankWindowFunction{rankWindowFunction: r.(*rankWindowFunction)}, nil
}

// Compute implements the sql.WindowFunction interface.
func (w *denseRankWindowFunction) Compute(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) (interface{}, error) {
	r, err := w.rankWindowFunction.Compute(ctx, interval, buf)
	if err != nil || r == nil {
		return r, err
	}
	rankVal := r.(int64)
	if rankVal == 1 {
		w.prevRank = 1
		w.denseRank = 1
	} else if rankVal != w.prevRank {
		w.prevRank = rankVal
		w.denseRank++
	}
	return w.denseRank, nil
}

// percentRankWindowFunction is the sql.WindowFunction used for percent_rank() within an OVER(...) clause. It
// reuses rankWindowFunction's peer-group counting the same way GMS's own PercentRank wraps rankBase: the
// result is (rank-1)/(partition size-1), or 0 for a single-row partition (rather than a divide-by-zero).
type percentRankWindowFunction struct {
	*rankWindowFunction
}

var _ sql.WindowFunction = (*percentRankWindowFunction)(nil)

// newPercentRankWindowFunction creates the sql.WindowFunction for percent_rank().
func newPercentRankWindowFunction(window *sql.WindowDefinition) (sql.WindowFunction, error) {
	r, err := newRankWindowFunction(window)
	if err != nil {
		return nil, err
	}
	return &percentRankWindowFunction{rankWindowFunction: r.(*rankWindowFunction)}, nil
}

// Compute implements the sql.WindowFunction interface.
func (w *percentRankWindowFunction) Compute(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) (interface{}, error) {
	partitionSize := w.partitionEnd - w.partitionStart
	r, err := w.rankWindowFunction.Compute(ctx, interval, buf)
	if err != nil || r == nil {
		return r, err
	}
	if partitionSize <= 1 {
		return float64(0), nil
	}
	return float64(r.(int64)-1) / float64(partitionSize-1), nil
}
