// Copyright 2024 Dolthub, Inc.
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

package functions

import (
	"io"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initForeachSlice registers the internal iterator used by PL/pgSQL FOREACH SLICE.
func initForeachSlice() { framework.RegisterFunction(foreach_slice) }

var foreach_slice = framework.Function2{
	Name:       "__doltgres_foreach_slice",
	Return:     pgtypes.RowTypeWithReturnType(pgtypes.AnyArray),
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.Int32},
	Strict:     true,
	SRF:        true,
	Callable: func(ctx *sql.Context, t [3]*pgtypes.DoltgresType, val, slice any) (any, error) {
		vals := val.([]any)
		dims := pgtypes.ArrayDims(vals, t[0].ArrayBaseType())
		n := int(slice.(int32))
		if n < 0 || n > len(dims) {
			return nil, pgerror.Newf(pgcode.ArraySubscript, "slice dimension (%d) is out of the valid range 0..%d", n, len(dims))
		}
		// Collect references to complete subarrays; the iteration never modifies them.
		var rows []any
		var collect func([]any, int)
		collect = func(a []any, depth int) {
			if depth == 0 {
				rows = append(rows, a)
				return
			}
			for _, v := range a {
				collect(v.([]any), depth-1)
			}
		}
		if n == 0 {
			return nil, pgerror.New(pgcode.InvalidParameterValue, "slice dimension must be greater than zero")
		}
		collect(vals, len(dims)-n)
		i := 0
		return pgtypes.NewSetReturningFunctionRowIter(func(ctx *sql.Context) (sql.Row, error) {
			if i >= len(rows) {
				return nil, io.EOF
			}
			row := sql.Row{rows[i]}
			i++
			return row, nil
		}), nil
	},
}
