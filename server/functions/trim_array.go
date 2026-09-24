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

package functions

import (
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/go-mysql-server/sql"
)

func initTrimArray() { framework.RegisterFunction(trim_array) }

var trim_array = framework.Function2{
	Name: "trim_array", Return: pgtypes.AnyArray, Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.Int32}, Strict: true,
	Callable: func(ctx *sql.Context, t [3]*pgtypes.DoltgresType, val, count any) (any, error) {
		vals := val.([]any)
		n := count.(int32)
		if n < 0 || int64(n) > int64(len(vals)) {
			return nil, pgerror.Newf(pgcode.ArraySubscript, "number of elements to trim must be between 0 and %d", len(vals))
		}
		result := make([]any, len(vals)-int(n))
		copy(result, vals)
		return result, nil
	},
}
