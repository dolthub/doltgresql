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

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// cumeDist represents the PostgreSQL cume_dist() window function.
var cumeDist = framework.Func0Window{
	Function0: framework.Function0{
		Name:   "cume_dist",
		Return: pgtypes.Float64,
	},
	NewWinFunc: newCumeDistWindowFunction,
}

// cumeDistWindowFunction is the sql.WindowFunction used for cume_dist() within an OVER(...) clause. It
// reuses rankWindowFunction's partition tracking and peer-group framing the same way percentRank does:
// per the SQL standard, cume_dist ignores any explicit frame clause and groups rows into peers by the
// window's ORDER BY expressions. The result is (number of rows preceding or peer with the current row) /
// (partition size).
type cumeDistWindowFunction struct {
	*rankWindowFunction
}

var _ sql.WindowFunction = (*cumeDistWindowFunction)(nil)

// newCumeDistWindowFunction creates the sql.WindowFunction for cume_dist().
func newCumeDistWindowFunction(window *sql.WindowDefinition) (sql.WindowFunction, error) {
	r, err := newRankWindowFunction(window)
	if err != nil {
		return nil, err
	}
	return &cumeDistWindowFunction{rankWindowFunction: r.(*rankWindowFunction)}, nil
}

// Compute implements the sql.WindowFunction interface.
func (w *cumeDistWindowFunction) Compute(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) (interface{}, error) {
	partitionSize := w.partitionEnd - w.partitionStart
	if partitionSize <= 0 {
		return nil, nil
	}
	return float64(interval.End-w.partitionStart) / float64(partitionSize), nil
}
