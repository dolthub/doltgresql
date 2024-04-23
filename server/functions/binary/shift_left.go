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
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = '<<' ORDER BY o.oprcode::varchar;

// initBinaryShiftLeft registers the functions to the catalog.
func initBinaryShiftLeft() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryShiftLeft, int2shl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryShiftLeft, int4shl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryShiftLeft, int8shl)
}

// int2shl represents the PostgreSQL function of the same name, taking the same parameters.
var int2shl = framework.Function2{
	Name:       "int2shl",
	Return:     pgtypes.Int16,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int32},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		return int16(int32(val1.(int16)) << val2.(int32)), nil
	},
}

// int4shl represents the PostgreSQL function of the same name, taking the same parameters.
var int4shl = framework.Function2{
	Name:       "int4shl",
	Return:     pgtypes.Int32,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		return int32(val1.(int32) << val2.(int32)), nil
	},
}

// int8shl represents the PostgreSQL function of the same name, taking the same parameters.
var int8shl = framework.Function2{
	Name:       "int8shl",
	Return:     pgtypes.Int64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int32},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		return int64(val1.(int64) << int64(val2.(int32))), nil
	},
}
