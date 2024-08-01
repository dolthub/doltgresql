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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/server/types/oid"
)

// initPgGetConstraintdef registers the functions to the catalog.
func initPgGetConstraintdef() {
	framework.RegisterFunction(pg_get_constraintdef1)
	framework.RegisterFunction(pg_get_constraintdef2)
}

// pg_get_constraintdef1 represents the PostgreSQL function pg_get_constraintdef, taking one OID parameter.
var pg_get_constraintdef1 = framework.Function1{
	Name:       "pg_get_constraintdef",
	Return:     pgtypes.Text,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Oid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val1 any) (any, error) {
		oidVal := val1.(uint32)
		def, err := getConstraintDef(ctx, oidVal)
		return def, err
	},
}

// pg_get_constraintdef2 represents the PostgreSQL function pg_get_constraintdef, taking two parameters.
var pg_get_constraintdef2 = framework.Function2{
	Name:       "pg_get_constraintdef",
	Return:     pgtypes.Text,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Bool},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		oidVal := val1.(uint32)
		pretty := val2.(bool)
		if pretty {
			return "", fmt.Errorf("pretty printing is not yet supported")
		}
		def, err := getConstraintDef(ctx, oidVal)
		if err != nil {
			return nil, err
		}
		return def, nil
	},
}

// getConstraintDef returns the definition of the constraint for the given OID.
func getConstraintDef(ctx *sql.Context, oidVal uint32) (string, error) {
	var result string
	err := oid.RunCallback(ctx, oidVal, oid.Callbacks{
		Check: func(ctx *sql.Context, schema oid.ItemSchema, table oid.ItemTable, check oid.ItemCheck) (cont bool, err error) {
			name := check.Item.Name
			if len(name) > 0 {
				name += " "
			}
			not := ""
			if !check.Item.Enforced {
				not = "not "
			}
			result = fmt.Sprintf("%sCHECK %s %sENFORCED", name, check.Item.CheckExpression, not)
			return false, nil
		},
		ForeignKey: func(ctx *sql.Context, schema oid.ItemSchema, table oid.ItemTable, fk oid.ItemForeignKey) (cont bool, err error) {
			result = fmt.Sprintf(
				"FOREIGN KEY %s (%s) REFERENCES %s (%s)",
				fk.Item.Name,
				getColumnNamesString(fk.Item.Columns),
				fk.Item.ParentTable,
				getColumnNamesString(fk.Item.ParentColumns),
			)
			return false, nil
		},
		Index: func(ctx *sql.Context, schema oid.ItemSchema, table oid.ItemTable, index oid.ItemIndex) (cont bool, err error) {
			colsStr := getColumnNamesString(index.Item.Expressions())
			if strings.ToLower(index.Item.ID()) == "primary" {
				result = fmt.Sprintf("PRIMARY KEY (%s)", colsStr)
			} else {
				result = fmt.Sprintf("UNIQUE (%s)", colsStr)
			}
			return false, nil
		},
	})
	if err != nil {
		return "", err
	}
	return result, nil
}

// getColumnNamesString returns a comma-separated string of column names with
// the table names removed from a list of expressions.
func getColumnNamesString(exprs []string) string {
	colNames := make([]string, len(exprs))
	for i, expr := range exprs {
		split := strings.Split(expr, ".")
		if len(split) == 0 {
			return ""
		}
		if len(split) == 1 {
			colNames[i] = split[0]
		} else {
			colNames[i] = split[1]
		}
	}
	return strings.Join(colNames, ", ")
}
