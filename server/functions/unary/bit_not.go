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

package unary

import (
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/ipaddr"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = '~' AND o.oprleft = 0 ORDER BY o.oprcode::varchar;

// initUnaryBitNot registers the functions to the catalog.
func initUnaryBitNot() {
	framework.RegisterUnaryFunction(framework.Operator_UnaryBitNot, inetnot)
	framework.RegisterUnaryFunction(framework.Operator_UnaryBitNot, int2not)
	framework.RegisterUnaryFunction(framework.Operator_UnaryBitNot, int4not)
	framework.RegisterUnaryFunction(framework.Operator_UnaryBitNot, int8not)
}

// inetnot represents the PostgreSQL function of the same name, taking the same parameters.
var inetnot = framework.Function1{
	Name:       "inetnot",
	Return:     pgtypes.Inet,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Inet},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		ip := val1.(ipaddr.IPAddr)
		return ip.Complement(), nil
	},
}

// int2not represents the PostgreSQL function of the same name, taking the same parameters.
var int2not = framework.Function1{
	Name:       "int2not",
	Return:     pgtypes.Int16,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		return ^val1.(int16), nil
	},
}

// int4not represents the PostgreSQL function of the same name, taking the same parameters.
var int4not = framework.Function1{
	Name:       "int4not",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		return ^val1.(int32), nil
	},
}

// int8not represents the PostgreSQL function of the same name, taking the same parameters.
var int8not = framework.Function1{
	Name:       "int8not",
	Return:     pgtypes.Int64,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		return ^val1.(int64), nil
	},
}
