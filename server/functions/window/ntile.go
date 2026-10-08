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

// ntile represents the PostgreSQL ntile(num_buckets) window function.
var ntile = framework.Func1Window{
	Function1: framework.Function1{
		Name:   "ntile",
		Return: pgtypes.Int32,
		Parameters: [1]*pgtypes.DoltgresType{
			pgtypes.Int32,
		},
		Callable: func(ctx *sql.Context, paramsAndReturn [2]*pgtypes.DoltgresType, val1 any) (any, error) {
			return nil, nil
		},
	},
	NewWinFunc: newNtileWindowFunction,
}

// ntileWindowFunction is the sql.WindowFunction used for ntile() within an OVER(...) clause. Per the SQL
// standard, ntile() ignores any explicit frame clause and always divides the whole partition into
// num_buckets buckets as evenly as possible, assigning any remainder one-per-bucket starting from the
// first bucket.
type ntileWindowFunction struct {
	numBucketsExpr sql.Expression

	pos        int64
	bucketSize int64
	bigBuckets int64
	bucket     int64
}

var _ sql.WindowFunction = (*ntileWindowFunction)(nil)

// newNtileWindowFunction creates the sql.WindowFunction for ntile().
func newNtileWindowFunction(exprs []sql.Expression, _ *sql.WindowDefinition) (sql.WindowFunction, error) {
	return &ntileWindowFunction{numBucketsExpr: exprs[0]}, nil
}

// DefaultFramer implements the sql.WindowFunction interface.
func (w *ntileWindowFunction) DefaultFramer() sql.WindowFramer {
	return aggregation.NewPartitionFramer()
}

// StartPartition implements the sql.WindowFunction interface.
func (w *ntileWindowFunction) StartPartition(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) error {
	numBucketsVal, err := w.numBucketsExpr.Eval(ctx, nil)
	if err != nil {
		return err
	}
	numBuckets, ok := numBucketsVal.(int32)
	if !ok || numBuckets <= 0 {
		return sql.ErrInvalidArgument.New("NTILE")
	}

	count := int64(interval.End - interval.Start)
	n := int64(numBuckets)
	if n > count {
		w.bucketSize = 1
		w.bigBuckets = 0
	} else {
		w.bucketSize = count / n
		w.bigBuckets = count % n
	}
	w.pos = 0
	w.bucket = 1
	return nil
}

// Compute implements the sql.WindowFunction interface.
func (w *ntileWindowFunction) Compute(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) (interface{}, error) {
	defer func() { w.pos++ }()
	if w.pos == 0 {
		return int32(w.bucket), nil
	}

	// The first w.bigBuckets buckets are of size w.bucketSize+1; the remaining buckets are of size
	// w.bucketSize.
	if w.bigBuckets > 0 && w.pos%(w.bucketSize+1) == 0 {
		w.bucket++
		w.bigBuckets--
		if w.bigBuckets == 0 {
			w.pos = 0
		}
	} else if w.bigBuckets == 0 && w.pos%w.bucketSize == 0 {
		w.bucket++
	}

	return int32(w.bucket), nil
}

// Dispose implements the sql.WindowFunction interface.
func (w *ntileWindowFunction) Dispose(ctx *sql.Context) {}
