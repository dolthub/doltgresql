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

package unary

import (
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = '+' AND o.oprleft = 0 ORDER BY o.oprcode::varchar;

// initUnaryPlus registers the functions to the catalog.
func initUnaryPlus() {
	framework.RegisterUnaryFunction(framework.Operator_UnaryPlus, float4up)
	framework.RegisterUnaryFunction(framework.Operator_UnaryPlus, float8up)
	framework.RegisterUnaryFunction(framework.Operator_UnaryPlus, int2up)
	framework.RegisterUnaryFunction(framework.Operator_UnaryPlus, int4up)
	framework.RegisterUnaryFunction(framework.Operator_UnaryPlus, int8up)
	framework.RegisterUnaryFunction(framework.Operator_UnaryPlus, numeric_uplus)
}

// float4up represents the PostgreSQL function of the same name, taking the same parameters.
var float4up = framework.Function1{
	Name:       "float4up",
	Return:     pgtypes.Float32,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float32},
	Callable: func(ctx *sql.Context, val1 any) (any, error) {
		if val1 == nil {
			return nil, nil
		}
		return val1, nil
	},
}

// float8up represents the PostgreSQL function of the same name, taking the same parameters.
var float8up = framework.Function1{
	Name:       "float8up",
	Return:     pgtypes.Float64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float64},
	Callable: func(ctx *sql.Context, val1 any) (any, error) {
		if val1 == nil {
			return nil, nil
		}
		return val1, nil
	},
}

// int2up represents the PostgreSQL function of the same name, taking the same parameters.
var int2up = framework.Function1{
	Name:       "int2up",
	Return:     pgtypes.Int16,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int16},
	Callable: func(ctx *sql.Context, val1 any) (any, error) {
		if val1 == nil {
			return nil, nil
		}
		return val1, nil
	},
}

// int4up represents the PostgreSQL function of the same name, taking the same parameters.
var int4up = framework.Function1{
	Name:       "int4up",
	Return:     pgtypes.Int32,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int32},
	Callable: func(ctx *sql.Context, val1 any) (any, error) {
		if val1 == nil {
			return nil, nil
		}
		return val1, nil
	},
}

// int8up represents the PostgreSQL function of the same name, taking the same parameters.
var int8up = framework.Function1{
	Name:       "int8up",
	Return:     pgtypes.Int64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int64},
	Callable: func(ctx *sql.Context, val1 any) (any, error) {
		if val1 == nil {
			return nil, nil
		}
		return val1, nil
	},
}

// numeric_uplus represents the PostgreSQL function of the same name, taking the same parameters.
var numeric_uplus = framework.Function1{
	Name:       "numeric_uplus",
	Return:     pgtypes.Numeric,
	Parameters: []pgtypes.DoltgresType{pgtypes.Numeric},
	Callable: func(ctx *sql.Context, val1 any) (any, error) {
		if val1 == nil {
			return nil, nil
		}
		return val1, nil
	},
}
