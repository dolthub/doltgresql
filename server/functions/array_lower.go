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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

func initArrayLower() { framework.RegisterFunction(array_lower) }

var array_lower = framework.Function2{
	Name:       "array_lower",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [3]*pgtypes.DoltgresType, val, dimension any) (any, error) {
		dims := pgtypes.ArrayDims(val.([]any), t[0].ArrayBaseType())
		dim := dimension.(int32)
		if dim < 1 || int(dim) > len(dims) {
			return nil, nil
		}
		if t[0].IsVectorType() {
			return int32(0), nil
		}
		return int32(1), nil
	},
}
