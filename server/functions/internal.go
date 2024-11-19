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

// initInternal registers the functions to the catalog.
func initInternal() {
	framework.RegisterFunction(internal_in)
	framework.RegisterFunction(internal_out)
}

// internal_in represents the PostgreSQL function of internal type IO input.
var internal_in = framework.Function1{
	Name:       "internal_in",
	Return:     pgtypes.Internal,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Cstring},
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		return []byte(val.(string)), nil
	},
}

// internal_out represents the PostgreSQL function of internal type IO output.
var internal_out = framework.Function1{
	Name:       "internal_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		return string(val.([]byte)), nil
	},
}
