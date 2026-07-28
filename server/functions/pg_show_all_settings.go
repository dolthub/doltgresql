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

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPgShowAllSettings registers the functions to the catalog.
func initPgShowAllSettings() {
	framework.RegisterFunction(pg_show_all_settings)
}

// pgShowAllSettingsName is the name for pg_show_all_settings function.
const pgShowAllSettingsName = "pg_show_all_settings"

// pg_show_all_settings represents the PostgreSQL function of the same name, taking the same parameters.
var pg_show_all_settings = framework.Function0{
	Name:               pgShowAllSettingsName,
	Return:             pgtypes.Record, // SETOF record
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context) (any, error) {
		//sysVarMap := sql.SystemVariables.GetAllSystemVariables()
		// TODO add all config parameters
		//  only bytea_output is added to support pgAdmin4 workbench
		row := sql.Row{
			"bytea_output", // name
			"hex",          // setting
			"",             // unit
			"Client Connection Defaults / Statement Behavior", // category
			"Sets the output format for bytea.",               // short_desc
			"",                                                // extra_desc
			"user",                                            // context
			"enum",                                            // vartype
			"default",                                         // source
			"",                                                // min_val
			"",                                                // max_val
			[]any{"escape", "hex"},                            // enumvals
			"hex",                                             // bool_val
			"hex",                                             // reset_val
			"",                                                // sourcefile
			"",                                                // sourceline
			false,                                             // pending_restart
		}

		var i = 0
		return pgtypes.NewSetReturningFunctionRowIter(func(ctx *sql.Context) (sql.Row, error) {
			defer func() { i++ }()

			if i >= 1 {
				return nil, io.EOF
			}
			return row, nil
		}), nil
	},
	OutParams: pgShowAllSettingsOutArgs,
}

// pgShowAllSettingsOutArgs is the schema for pg_show_all_settings table function. Each column is OUT argument.
var pgShowAllSettingsOutArgs = sql.Schema{
	{Name: "name", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgShowAllSettingsName},
	{Name: "setting", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgShowAllSettingsName},
	{Name: "unit", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgShowAllSettingsName},
	{Name: "category", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgShowAllSettingsName},
	{Name: "short_desc", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgShowAllSettingsName},
	{Name: "extra_desc", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgShowAllSettingsName},
	{Name: "context", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgShowAllSettingsName},
	{Name: "vartype", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgShowAllSettingsName},
	{Name: "source", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgShowAllSettingsName},
	{Name: "min_val", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgShowAllSettingsName},
	{Name: "max_val", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgShowAllSettingsName},
	{Name: "enumvals", Type: pgtypes.TextArray, Default: nil, Nullable: false, Source: pgShowAllSettingsName},
	{Name: "bool_val", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgShowAllSettingsName},
	{Name: "reset_val", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgShowAllSettingsName},
	{Name: "sourcefile", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgShowAllSettingsName},
	{Name: "sourceline", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: pgShowAllSettingsName},
	{Name: "pending_restart", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: pgShowAllSettingsName},
}
