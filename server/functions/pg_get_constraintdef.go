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

// initPgGetConstraintDef registers the functions to the catalog.
func initPgGetConstraintDef() {
	framework.RegisterFunction(pg_get_constraintdef1)
	framework.RegisterFunction(pg_get_constraintdef2)
}

// pg_get_constraintdef represents the PostgreSQL system catalog information function taking 1 parameter.
var pg_get_constraintdef1 = framework.Function1{
	Name:               "pg_get_constraintdef",
	Return:             pgtypes.Text,
	Parameters:         []pgtypes.DoltgresType{pgtypes.Oid},
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, val any) (any, error) {
		return nil, nil
	},
	Strict: true,
}

// pg_get_constraintdef represents the PostgreSQL system catalog information function taking 2 parameters.
var pg_get_constraintdef2 = framework.Function2{
	Name:               "pg_get_constraintdef",
	Return:             pgtypes.Text,
	Parameters:         []pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Bool},
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, val1, val2 any) (any, error) {
		return nil, nil
	},
	Strict: true,
}
