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

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPgGetExpr registers the functions to the catalog.
func initPgGetExpr() {
	framework.RegisterFunction(pg_get_expr_pgnodetree_oid)
	framework.RegisterFunction(pg_get_expr_pgnodetree_oid_bool)
}

// pg_get_expr_pgnodetree_oid represents the PostgreSQL function of the same name, taking the same parameters.
var pg_get_expr_pgnodetree_oid = framework.Function2{
	Name:       "pg_get_expr",
	Return:     pgtypes.Text,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Text, pgtypes.Oid}, // TODO: First parameter should be pg_node_tree
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		// TODO: Implement this when the pg_node_tree type exists
		return nil, fmt.Errorf("pg_get_expr is not yet supported")
	},
}

// pg_get_expr_pgnodetree_oid_bool represents the PostgreSQL function of the same name, taking the same parameters.
var pg_get_expr_pgnodetree_oid_bool = framework.Function3{
	Name:       "pg_get_expr",
	Return:     pgtypes.Text,
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Text, pgtypes.Oid, pgtypes.Bool}, // TODO: First parameter should be pg_node_tree
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		// TODO: Implement this when the pg_node_tree type exists
		return nil, fmt.Errorf("pg_get_expr is not yet supported")
	},
}
