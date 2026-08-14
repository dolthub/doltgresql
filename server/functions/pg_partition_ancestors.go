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

package functions

import (
	"io"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPgPartitionAncestors registers the functions to the catalog.
func initPgPartitionAncestors() {
	framework.RegisterFunction(pg_partition_ancestors_regclass)
}

// pgPartitionAncestorsName is the name of the pg_partition_ancestors function.
const pgPartitionAncestorsName = "pg_partition_ancestors"

// pg_partition_ancestors_regclass represents the PostgreSQL partitioning information function of the same name. It
// lists the ancestors of the given partition, including the relation itself. Since we don't support table
// partitioning, a relation's only ancestor is itself; relations that don't exist return an empty set, matching
// Postgres.
var pg_partition_ancestors_regclass = framework.Function1{
	Name:               pgPartitionAncestorsName,
	Return:             pgtypes.RowTypeWithReturnType(pgtypes.Regclass),
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Regclass},
	IsNonDeterministic: true,
	Strict:             true,
	SRF:                true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		oidVal := val.(id.Id)
		exists := false
		err := RunCallback(ctx, oidVal, Callbacks{
			Table: func(ctx *sql.Context, schema ItemSchema, table ItemTable) (cont bool, err error) {
				exists = true
				return false, nil
			},
			Index: func(ctx *sql.Context, schema ItemSchema, table ItemTable, index ItemIndex) (cont bool, err error) {
				exists = true
				return false, nil
			},
		})
		if err != nil {
			return nil, err
		}
		returned := false
		return pgtypes.NewSetReturningFunctionRowIter(func(ctx *sql.Context) (sql.Row, error) {
			if !exists || returned {
				return nil, io.EOF
			}
			returned = true
			return sql.Row{oidVal}, nil
		}), nil
	},
	OutParams: sql.Schema{
		{Name: "relid", Type: pgtypes.Regclass, Default: nil, Nullable: false, Source: pgPartitionAncestorsName},
	},
}
