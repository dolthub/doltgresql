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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PostgreSQL command-type bits used by pg_relation_is_updatable.
const (
	RelationUpdatable  int32 = 4
	RelationInsertable int32 = 8
	RelationDeletable  int32 = 16
)

// initPgRelationIsUpdatable registers the function to the catalog.
func initPgRelationIsUpdatable() {
	framework.RegisterFunction(pg_relation_is_updatable)
}

// pg_relation_is_updatable reports supported writes, independently of the caller's privileges.
var pg_relation_is_updatable = framework.Function2{
	Name:               "pg_relation_is_updatable",
	Return:             pgtypes.Int32,
	Parameters:         [2]*pgtypes.DoltgresType{pgtypes.Regclass, pgtypes.Bool},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, relation, includeTriggers any) (any, error) {
		// includeTriggers would include writes enabled by INSTEAD OF triggers.
		// Doltgres does not support those triggers yet, so it does not affect the mask.
		var operations int32
		// No View callback: views currently support no writes in Doltgres.
		err := RunCallback(ctx, relation.(id.Id), Callbacks{
			Table: func(ctx *sql.Context, schema ItemSchema, table ItemTable) (bool, error) {
				if _, ok := table.Item.(sql.UpdatableTable); ok {
					operations |= RelationUpdatable
				}
				if _, ok := table.Item.(sql.InsertableTable); ok {
					operations |= RelationInsertable
				}
				if _, ok := table.Item.(sql.DeletableTable); ok {
					operations |= RelationDeletable
				}
				return false, nil
			},
		})
		// Cached OIDs can outlive their relations.
		if sql.ErrTableNotFound.Is(err) {
			return int32(0), nil
		}
		return operations, err
	},
}
