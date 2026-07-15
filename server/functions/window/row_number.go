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

// initRanking registers the ranking window functions to the catalog.
func initRanking() {
	framework.RegisterWindowFunction(rowNumber)
	framework.RegisterWindowFunction(rank)
	framework.RegisterWindowFunction(denseRank)
	framework.RegisterWindowFunction(percentRank)
}

// rowNumber represents the PostgreSQL row_number() window function, returning a bigint directly rather than
// GMS's native uint64/int representation.
var rowNumber = framework.Func0Window{
	Function0: framework.Function0{
		Name:   "row_number",
		Return: pgtypes.Int64,
	},
	NewWinFunc: newRowNumberWindowFunction,
}

// rowNumberWindowFunction is the sql.WindowFunction used for row_number() within an OVER(...) clause. Per the
// SQL standard, row_number() ignores any explicit frame clause and always numbers rows across the whole
// partition, so DefaultFramer always returns a partition-wide framer regardless of window.Frame.
type rowNumberWindowFunction struct {
	pos int64
}

var _ sql.WindowFunction = (*rowNumberWindowFunction)(nil)

// newRowNumberWindowFunction creates the sql.WindowFunction for row_number().
func newRowNumberWindowFunction(_ *sql.WindowDefinition) (sql.WindowFunction, error) {
	return &rowNumberWindowFunction{}, nil
}

// StartPartition implements the sql.WindowFunction interface.
func (w *rowNumberWindowFunction) StartPartition(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) error {
	w.pos = 1
	return nil
}

// DefaultFramer implements the sql.WindowFunction interface.
func (w *rowNumberWindowFunction) DefaultFramer() sql.WindowFramer {
	return aggregation.NewPartitionFramer()
}

// Compute implements the sql.WindowFunction interface.
func (w *rowNumberWindowFunction) Compute(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) (interface{}, error) {
	if interval.End-interval.Start < 1 {
		return nil, nil
	}
	defer func() { w.pos++ }()
	return w.pos, nil
}

// Dispose implements the sql.WindowFunction interface.
func (w *rowNumberWindowFunction) Dispose(ctx *sql.Context) {}
