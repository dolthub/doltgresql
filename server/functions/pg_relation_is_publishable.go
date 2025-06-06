// Copyright 2025 Dolthub, Inc.
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

// initPgRelationIsPublishable registers the functions to the catalog.
func initPgRelationIsPublishable() {
	framework.RegisterFunction(pg_relation_is_publishable)
}

// pg_relation_is_publishable represents the PostgreSQL function of the same name, taking the same parameters.
var pg_relation_is_publishable = framework.Function1{
	Name:               "pg_relation_is_publishable",
	Return:             pgtypes.Bool,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Regclass},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		// TODO: true is returned if the relation is a non-system catalog, non-temporary table or partitioned table, otherwise false.
		return nil, nil
	},
}
