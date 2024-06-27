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
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/go-mysql-server/sql"
)

// initPgGetViewDef registers the functions to the catalog.
func initPgGetViewDef() {
	framework.RegisterFunction(pg_get_viewdef1)
	framework.RegisterFunction(pg_get_viewdef2bool)
	framework.RegisterFunction(pg_get_viewdef2int)
}

// pg_get_viewdef represents the PostgreSQL system catalog information function taking 1 parameter, {oid}.
var pg_get_viewdef1 = framework.Function1{
	Name:               "pg_get_viewdef",
	Return:             pgtypes.Text,
	Parameters:         []pgtypes.DoltgresType{pgtypes.Oid},
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, val1 any) (any, error) {
		return nil, nil
	},
	Strict: true,
}

// pg_get_viewdef represents the PostgreSQL system catalog information function taking 2 parameters, {oid, bool}.
var pg_get_viewdef2bool = framework.Function2{
	Name:               "pg_get_viewdef",
	Return:             pgtypes.Text,
	Parameters:         []pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Bool},
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, val1, val2 any) (any, error) {
		// TODO: if val2 { prettyprint }
		return nil, nil
	},
	Strict: true,
}

// pg_get_viewdef represents the PostgreSQL system catalog information function taking 2 parameters, {oid, int}.
var pg_get_viewdef2int = framework.Function2{
	Name:               "pg_get_viewdef",
	Return:             pgtypes.Text,
	Parameters:         []pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Int64},
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, val1, val2 any) (any, error) {
		// TODO: prettyprint is implied
		return nil, nil
	},
	Strict: true,
}
