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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions/framework"
	"github.com/dolthub/doltgresql/server/tables"
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
		var result any
		err := RunCallback(ctx, val.(id.Id), Callbacks{
			Table: func(ctx *sql.Context, schema ItemSchema, table ItemTable) (bool, error) {
				result = IsPublishableTable(schema, table)
				return false, nil
			},
			View: func(ctx *sql.Context, schema ItemSchema, view ItemView) (bool, error) {
				result = false
				return false, nil
			},
			Sequence: func(ctx *sql.Context, schema ItemSchema, sequence ItemSequence) (bool, error) {
				result = false
				return false, nil
			},
			Index: func(ctx *sql.Context, schema ItemSchema, table ItemTable, index ItemIndex) (bool, error) {
				result = false
				return false, nil
			},
		})
		// OIDs can outlive the object they identify. PostgreSQL returns NULL for a missing relation.
		if sql.ErrTableNotFound.Is(err) || errors.Is(err, doltdb.ErrTableNotFound) {
			return nil, nil
		}
		return result, err
	},
}

// IsPublishableTable reports whether a base table is eligible for publication.
// Views and sequences are separate iteration callbacks. UNLOGGED and foreign
// tables are not supported by Doltgres, so all remaining user base tables persist.
func IsPublishableTable(schema ItemSchema, table ItemTable) bool {
	if schema.IsSystemSchema() || doltdb.IsSystemTable(doltdb.TableName{Schema: schema.Item.SchemaName(), Name: table.Item.Name()}) {
		return false
	}
	underlying := sql.GetUnderlyingTable(table.Item)
	if _, virtual := underlying.(*tables.VirtualTable); virtual {
		return false
	}
	if temporary, ok := underlying.(sql.TemporaryTable); ok && temporary.IsTemporary() {
		return false
	}
	return true
}
