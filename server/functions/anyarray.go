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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initAnyArray registers the functions to the catalog.
func initAnyArray() {
	framework.RegisterFunction(anyarray_in)
	framework.RegisterFunction(anyarray_out)
	framework.RegisterFunction(anyarray_recv)
	framework.RegisterFunction(anyarray_send)
}

// anyarray_in represents the PostgreSQL function of anyarray type IO input.
var anyarray_in = framework.Function1{
	Name:       "anyarray_in",
	Return:     pgtypes.AnyArray,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		return []any{}, nil
	},
}

// anyarray_out represents the PostgreSQL function of anyarray type IO output.
var anyarray_out = framework.Function1{
	Name:       "anyarray_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.AnyArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		return "", nil
	},
}

// anyarray_recv represents the PostgreSQL function of anyarray type IO receive.
var anyarray_recv = framework.Function1{
	Name:       "anyarray_recv",
	Return:     pgtypes.AnyArray,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		return []any{}, nil
	},
}

// anyarray_send represents the PostgreSQL function of anyarray type IO send.
var anyarray_send = framework.Function1{
	Name:       "anyarray_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.AnyArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		return []byte{}, nil
	},
}
