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

package pgcatalog

import (
	"fmt"
	"io"
	"sort"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/config"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgSettingsName is a constant to the pg_settings name.
const PgSettingsName = "pg_settings"

// InitPgSettings handles registration of the pg_settings handler.
func InitPgSettings() {
	tables.AddHandler(PgCatalogName, PgSettingsName, PgSettingsHandler{})
}

// PgSettingsHandler is the handler for the pg_settings table.
type PgSettingsHandler struct{}

var _ tables.Handler = PgSettingsHandler{}

// Name implements the interface tables.Handler.
func (p PgSettingsHandler) Name() string {
	return PgSettingsName
}

// RowIter implements the interface tables.Handler.
func (p PgSettingsHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	configParams := config.PostgresConfigParameters()
	names := make([]string, 0, len(configParams))
	for name := range configParams {
		names = append(names, name)
	}
	sort.Strings(names)

	params := make([]*config.Parameter, 0, len(names))
	for _, name := range names {
		if param, ok := configParams[name].(*config.Parameter); ok {
			params = append(params, param)
		}
	}
	return &pgSettingsRowIter{
		params: params,
		idx:    0,
	}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgSettingsHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgSettingsSchema,
		PkOrdinals: nil,
	}
}

// pgSettingsSchema is the schema for pg_settings.
var pgSettingsSchema = sql.Schema{
	{Name: "name", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "setting", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "unit", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "category", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "short_desc", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "extra_desc", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "context", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "vartype", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "source", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "min_val", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "max_val", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "enumvals", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "boot_val", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "reset_val", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "sourcefile", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "sourceline", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "pending_restart", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgSettingsName},
}

// pgSettingsVarType returns the pg_settings vartype string for the given configuration parameter type.
// The system variable type structs in go-mysql-server are mostly unexported, so we rely on their stable
// String() identifiers (e.g. "system_bool") to distinguish them.
func pgSettingsVarType(t sql.Type) string {
	switch t.String() {
	case "system_bool":
		return "bool"
	case "system_int":
		return "integer"
	case "system_double":
		return "real"
	case "system_enum":
		return "enum"
	case "system_string":
		return "string"
	default:
		return "string"
	}
}

// formatSettingValue formats a configuration parameter value for display in pg_settings. Boolean
// parameters are stored as int8 (or bool) values, but Postgres displays them as "on"/"off".
func formatSettingValue(param *config.Parameter, val any) string {
	if val == nil {
		return ""
	}
	if pgSettingsVarType(param.Type) == "bool" {
		switch v := val.(type) {
		case bool:
			if v {
				return "on"
			}
			return "off"
		case int8:
			if v != 0 {
				return "on"
			}
			return "off"
		case int64:
			if v != 0 {
				return "on"
			}
			return "off"
		}
	}
	return fmt.Sprintf("%v", val)
}

// currentSettingValue returns the current session value of the given configuration parameter,
// falling back to the parameter's default value if the session value is unavailable.
func currentSettingValue(ctx *sql.Context, param *config.Parameter) string {
	val, err := ctx.GetSessionVariable(ctx, param.Name)
	if err != nil || val == nil {
		val = param.Default
	}
	return formatSettingValue(param, val)
}

// pgSettingsRowIter is the sql.RowIter for the pg_settings table.
type pgSettingsRowIter struct {
	params []*config.Parameter
	idx    int
}

var _ sql.RowIter = (*pgSettingsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgSettingsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.params) {
		return nil, io.EOF
	}
	iter.idx++
	param := iter.params[iter.idx-1]

	// TODO: fill in unit, extra_desc, min_val, max_val, and enumvals from the parameter definitions
	return sql.Row{
		param.Name,                               // name
		currentSettingValue(ctx, param),          // setting
		nil,                                      // unit
		param.Category,                           // category
		param.ShortDesc,                          // short_desc
		nil,                                      // extra_desc
		string(param.Context),                    // context
		pgSettingsVarType(param.Type),            // vartype
		string(param.Source),                     // source
		nil,                                      // min_val
		nil,                                      // max_val
		nil,                                      // enumvals
		formatSettingValue(param, param.Default), // boot_val
		formatSettingValue(param, param.ResetVal), // reset_val
		nil,   // sourcefile
		nil,   // sourceline
		false, // pending_restart
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgSettingsRowIter) Close(ctx *sql.Context) error {
	return nil
}
