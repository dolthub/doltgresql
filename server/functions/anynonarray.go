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

func initAnyNonArray() {
	framework.RegisterFunction(anynonarray_in)
	framework.RegisterFunction(anynonarray_out)
}

// anynonarray_in represents the PostgreSQL function of anynonarray type IO input.
var anynonarray_in = framework.Function1{
	Name:       "anynonarray_in",
	Return:     pgtypes.AnyNonArray,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Text}, // cstring
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		return nil, nil
	},
}

// anynonarray_out represents the PostgreSQL function of anynonarray type IO output.
var anynonarray_out = framework.Function1{
	Name:       "anynonarray_out",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.AnyNonArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		return "", nil
	},
}
