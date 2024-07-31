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
	"fmt"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/go-mysql-server/sql"
)

// initPgGetTriggerDef registers the functions to the catalog.
func initPgGetTriggerDef() {
	framework.RegisterFunction(pg_get_triggerdef1)
	framework.RegisterFunction(pg_get_triggerdef2)
}

// pg_get_triggerdef1 represents the PostgreSQL system catalog information function taking 1 parameter.
var pg_get_triggerdef1 = framework.Function1{
	Name:               "pg_get_triggerdef",
	Return:             pgtypes.Text,
	Parameters:         [1]pgtypes.DoltgresType{pgtypes.Oid},
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO: triggers are not supported yet
		return "", nil
	},
	Strict: true,
}

// pg_get_triggerdef2 represents the PostgreSQL system catalog information function taking 2 parameters.
var pg_get_triggerdef2 = framework.Function2{
	Name:               "pg_get_triggerdef",
	Return:             pgtypes.Text,
	Parameters:         [2]pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Bool},
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		pretty := val2.(bool)
		if pretty {
			return "", fmt.Errorf("pretty printing is not yet supported")
		}
		// TODO: triggers are not supported yet
		return "", nil
	},
	Strict: true,
}
