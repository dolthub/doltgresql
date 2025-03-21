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
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/resolve"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPgGetSerialSequence registers the functions to the catalog.
func initPgGetSerialSequence() {
	framework.RegisterFunction(pg_get_serial_sequence_text_text)
}

// pg_get_serial_sequence_text_text represents the PostgreSQL function of the same name, taking the same parameters.
var pg_get_serial_sequence_text_text = framework.Function2{
	Name:               "pg_get_serial_sequence",
	Return:             pgtypes.Text,
	Parameters:         [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Variadic:           false,
	IsNonDeterministic: false,
	Strict:             true,
	Callable: func(ctx *sql.Context, paramsAndReturn [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		tableName := val1.(string)
		columnName := val2.(string)

		// Parse out the schema if one was supplied
		var err error
		schemaName := ""
		if strings.Contains(tableName, ".") {
			// TODO: parseRelationName() will return the first schema from the search_path if one is not included
			//       in the relation name, but that doesn't mean it's the correct schema. It should be updated to
			//       not return any schema name if one wasn't explicitly specified, then we should search for the
			//       table on the search_path and find the first schema that contains a table with that name.
			schemaName, tableName, err = parseRelationName(ctx, tableName)
			if err != nil {
				return nil, err
			}
		}

		// Resolve the table's schema if it wasn't specified
		if schemaName == "" {
			doltSession := dsess.DSessFromSess(ctx.Session)
			roots, ok := doltSession.GetRoots(ctx, ctx.GetCurrentDatabase())
			if !ok {
				return nil, errors.Errorf("unable to get roots")
			}
			foundTableName, _, ok, err := resolve.TableWithSearchPath(ctx, roots.Working, tableName)
			if err != nil {
				return nil, err
			}
			if !ok {
				return nil, errors.Errorf(`relation "%s" does not exist`, tableName)
			}
			schemaName = foundTableName.Schema
		}

		// Validate the full schema + table name and grab the columns
		table, err := core.GetSqlTableFromContext(ctx, "", doltdb.TableName{
			Schema: schemaName,
			Name:   tableName,
		})
		if err != nil {
			return nil, err
		}
		if table == nil {
			return nil, errors.Errorf(`relation "%s" does not exist`, tableName)
		}
		tableSchema := table.Schema()

		// Find the column in the table's schema
		columnIndex := tableSchema.IndexOfColName(columnName)
		if columnIndex < 0 {
			return nil, errors.Errorf(`column "%s" of relation "%s" does not exist`, columnName, tableName)
		}
		column := tableSchema[columnIndex]

		// Find any sequence associated with the column
		sequenceCollection, err := core.GetSequencesCollectionFromContext(ctx)
		if err != nil {
			return nil, err
		}
		sequences, err := sequenceCollection.GetSequencesWithTable(ctx, doltdb.TableName{
			Name:   tableName,
			Schema: schemaName,
		})
		if err != nil {
			return nil, err
		}
		for _, sequence := range sequences {
			if sequence.OwnerColumn == column.Name {
				// pg_get_serial_sequence() always includes the schema name in its output
				return schemaName + "." + sequence.Id.SequenceName(), nil
			}
		}

		return nil, nil
	},
}
