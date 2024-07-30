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

package binary

import (
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = '>>' ORDER BY o.oprcode::varchar;

// initBinaryShiftRight registers the functions to the catalog.
func initBinaryShiftRight() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryShiftRight, int2shr)
	framework.RegisterBinaryFunction(framework.Operator_BinaryShiftRight, int4shr)
	framework.RegisterBinaryFunction(framework.Operator_BinaryShiftRight, int8shr)
}

// int2shr represents the PostgreSQL function of the same name, taking the same parameters.
var int2shr = framework.Function2{
	Name:       "int2shr",
	Return:     pgtypes.Int16,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any, varargs ...any) (any, error) {
		return int16(int32(val1.(int16)) >> val2.(int32)), nil
	},
}

// int4shr represents the PostgreSQL function of the same name, taking the same parameters.
var int4shr = framework.Function2{
	Name:       "int4shr",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any, varargs ...any) (any, error) {
		return int32(val1.(int32) >> val2.(int32)), nil
	},
}

// int8shr represents the PostgreSQL function of the same name, taking the same parameters.
var int8shr = framework.Function2{
	Name:       "int8shr",
	Return:     pgtypes.Int64,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any, varargs ...any) (any, error) {
		return int64(val1.(int64) >> int64(val2.(int32))), nil
	},
}
