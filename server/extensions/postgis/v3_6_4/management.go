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

package v3_6_4

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	"github.com/dolthub/doltgresql/server/hook"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// quoteLiteral returns the SQL string literal for the text.
func quoteLiteral(text string) string {
	return "'" + strings.ReplaceAll(text, "'", "''") + "'"
}

// managedTypeNames holds the mixed-case name of each type that AddGeometryColumn accepts, keyed by its upper-case
// name without a dimension suffix.
var managedTypeNames = map[string]string{
	"POINT":              "Point",
	"MULTIPOINT":         "MultiPoint",
	"LINESTRING":         "LineString",
	"MULTILINESTRING":    "MultiLineString",
	"POLYGON":            "Polygon",
	"MULTIPOLYGON":       "MultiPolygon",
	"CIRCULARSTRING":     "CircularString",
	"COMPOUNDCURVE":      "CompoundCurve",
	"MULTICURVE":         "MultiCurve",
	"CURVEPOLYGON":       "CurvePolygon",
	"MULTISURFACE":       "MultiSurface",
	"GEOMETRY":           "Geometry",
	"GEOMETRYCOLLECTION": "GeometryCollection",
	"TRIANGLE":           "Triangle",
	"POLYHEDRALSURFACE":  "PolyhedralSurface",
	"TIN":                "Tin",
}

// managedTypeModifier returns the type name with its dimension suffix that a geometry column of the given type and
// number of dimensions takes in its type modifier. A type without a suffix takes Z for three dimensions and ZM for
// four, and a type with a suffix only allows the dimensions that the suffix names.
func managedTypeModifier(typeName string, dims int32) (string, error) {
	upper := strings.ToUpper(typeName)
	base, suffix := upper, ""
	for _, candidate := range []string{"ZM", "Z", "M"} {
		if trimmed, ok := strings.CutSuffix(upper, candidate); ok && managedTypeNames[trimmed] != "" {
			base, suffix = trimmed, candidate
			break
		}
	}
	name, ok := managedTypeNames[base]
	valid := ok && (suffix == "" && dims >= 2 && dims <= 4 || (suffix == "Z" || suffix == "M") && dims == 3 || suffix == "ZM" && dims == 4)
	if !valid {
		return "", pgerror.Newf(pgcode.RaiseException, "Invalid type name \"%s(%d)\" - valid ones are:\n\tPOINT, MULTIPOINT,\n\t"+
			"LINESTRING, MULTILINESTRING,\n\tPOLYGON, MULTIPOLYGON,\n\tCIRCULARSTRING, COMPOUNDCURVE, MULTICURVE,\n\t"+
			"CURVEPOLYGON, MULTISURFACE,\n\tGEOMETRY, GEOMETRYCOLLECTION,\n\tPOINTM, MULTIPOINTM,\n\t"+
			"LINESTRINGM, MULTILINESTRINGM,\n\tPOLYGONM, MULTIPOLYGONM,\n\tCIRCULARSTRINGM, COMPOUNDCURVEM, MULTICURVEM\n\t"+
			"CURVEPOLYGONM, MULTISURFACEM, TRIANGLE, TRIANGLEM,\n\tPOLYHEDRALSURFACE, POLYHEDRALSURFACEM, TIN, TINM\n\t"+
			"or GEOMETRYCOLLECTIONM", typeName, dims)
	}
	if strings.HasSuffix(suffix, "M") && dims != 3 {
		return "", pgerror.New(pgcode.RaiseException, "TypeM needs 3 dimensions")
	}
	if suffix == "" && dims == 3 {
		suffix = "Z"
	} else if suffix == "" && dims == 4 {
		suffix = "ZM"
	}
	return name + suffix, nil
}

// managedSchema returns the schema of a table for the column management functions: the given schema when there is
// one, and otherwise the first schema on the search path that holds the table, or false when none does.
func managedSchema(ctx *sql.Context, runner sql.StatementRunner, schemaName string, tableName string) (string, bool, error) {
	if schemaName != "" {
		return schemaName, true, nil
	}
	searchPath, err := core.SearchPath(ctx)
	if err != nil {
		return "", false, err
	}
	for _, schema := range searchPath {
		rows, err := queryRows(ctx, runner, fmt.Sprintf("SELECT 1 FROM pg_class c JOIN pg_namespace n ON c.relnamespace = n.oid WHERE n.nspname = %s AND c.relname = %s",
			quoteLiteral(schema), quoteLiteral(tableName)))
		if err != nil {
			return "", false, err
		}
		if len(rows) > 0 {
			return schema, true, nil
		}
	}
	return "", false, nil
}

// knownSRID returns whether spatial_ref_sys holds the SRID.
func knownSRID(ctx *sql.Context, runner sql.StatementRunner, srid int32) (bool, error) {
	rows, err := queryRows(ctx, runner, fmt.Sprintf("SELECT 1 FROM spatial_ref_sys WHERE srid = %d", srid))
	return len(rows) > 0, err
}

// addGeometryColumn implements AddGeometryColumn, which adds a geometry column whose type modifier holds the type,
// number of dimensions, and SRID, and describes the new column. The catalog name, when given, is ignored.
func addGeometryColumn(ctx *sql.Context, runner sql.StatementRunner, args ...any) (any, error) {
	for len(args) < 8 {
		args = append([]any{""}, args...)
	}
	schemaName, tableName, columnName := args[1].(string), args[2].(string), args[3].(string)
	srid, typeName, dims, useTypmod := args[4].(int32), args[5].(string), args[6].(int32), args[7].(bool)
	modifier, err := managedTypeModifier(typeName, dims)
	if err != nil {
		return nil, err
	}
	if srid < 0 {
		sendNotice(ctx, "NOTICE", fmt.Sprintf("SRID value %d converted to the officially unknown SRID value 0", srid))
		srid = 0
	} else if srid != 0 {
		known, err := knownSRID(ctx, runner, srid)
		if err != nil {
			return nil, err
		}
		if !known {
			return nil, pgerror.New(pgcode.RaiseException, "AddGeometryColumn() - invalid SRID")
		}
	}
	if schemaName != "" {
		rows, err := queryRows(ctx, runner, "SELECT 1 FROM pg_namespace WHERE nspname = "+quoteLiteral(schemaName))
		if err != nil {
			return nil, err
		}
		if len(rows) == 0 {
			return nil, pgerror.Newf(pgcode.RaiseException, "Schema '%s' is not a valid schemaname", schemaName)
		}
	}
	schemaName, ok, err := managedSchema(ctx, runner, schemaName, tableName)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, pgerror.Newf(pgcode.RaiseException, "Table '%s' does not occur in the search_path", tableName)
	}
	if !useTypmod {
		//TODO: add the CHECK constraints that describe the column without a type modifier, once constraints may call
		// extension functions
		return nil, pgerror.New(pgcode.FeatureNotSupported, "AddGeometryColumn without a type modifier is not yet supported")
	}
	statement := fmt.Sprintf("ALTER TABLE %s.%s ADD COLUMN %s geometry(%s, %d)", hook.QuoteIdentifier(schemaName),
		hook.QuoteIdentifier(tableName), hook.QuoteIdentifier(columnName), modifier, srid)
	if err = hook.RunStatement(ctx, runner, statement); err != nil {
		return nil, err
	}
	return fmt.Sprintf("%s.%s.%s SRID:%d TYPE:%s DIMS:%d ", schemaName, tableName, columnName, srid, typeName, dims), nil
}

// registeredColumn returns the type and number of dimensions of a column that geometry_columns describes, or false
// when it describes no such column.
func registeredColumn(ctx *sql.Context, runner sql.StatementRunner, schemaName string, tableName string, columnName string) (string, int32, bool, error) {
	rows, err := queryRows(ctx, runner, fmt.Sprintf("SELECT type, coord_dimension FROM geometry_columns WHERE f_table_schema = %s AND f_table_name = %s AND f_geometry_column = %s",
		quoteLiteral(schemaName), quoteLiteral(tableName), quoteLiteral(columnName)))
	if err != nil || len(rows) == 0 {
		return "", 0, false, err
	}
	return rows[0][0].(string), rows[0][1].(int32), true, nil
}

// dropGeometryColumn implements DropGeometryColumn, which drops a column that geometry_columns describes. The catalog
// name, when given, is ignored.
func dropGeometryColumn(ctx *sql.Context, runner sql.StatementRunner, args ...any) (any, error) {
	for len(args) < 4 {
		args = append([]any{""}, args...)
	}
	schemaName, ok, err := managedSchema(ctx, runner, args[1].(string), args[2].(string))
	if err != nil {
		return nil, err
	}
	tableName, columnName := args[2].(string), args[3].(string)
	if ok {
		_, _, ok, err = registeredColumn(ctx, runner, schemaName, tableName, columnName)
		if err != nil {
			return nil, err
		}
	}
	if !ok {
		return nil, pgerror.New(pgcode.RaiseException, "column not found in geometry_columns table")
	}
	statement := fmt.Sprintf("ALTER TABLE %s.%s DROP COLUMN %s", hook.QuoteIdentifier(schemaName), hook.QuoteIdentifier(tableName), hook.QuoteIdentifier(columnName))
	if err = hook.RunStatement(ctx, runner, statement); err != nil {
		return nil, err
	}
	return fmt.Sprintf("%s.%s.%s effectively removed.", schemaName, tableName, columnName), nil
}

// dropGeometryTable implements DropGeometryTable, which drops a table if it exists. The catalog name, when given, is
// ignored.
func dropGeometryTable(ctx *sql.Context, runner sql.StatementRunner, args ...any) (any, error) {
	for len(args) < 3 {
		args = append([]any{""}, args...)
	}
	schemaName, tableName := args[1].(string), args[2].(string)
	if schemaName == "" {
		rows, err := queryRows(ctx, runner, "SELECT current_schema()")
		if err != nil {
			return nil, err
		}
		schemaName = rows[0][0].(string)
	}
	if err := hook.RunStatement(ctx, runner, fmt.Sprintf("DROP TABLE IF EXISTS %s.%s", hook.QuoteIdentifier(schemaName), hook.QuoteIdentifier(tableName))); err != nil {
		return nil, err
	}
	return fmt.Sprintf("%s.%s dropped.", schemaName, tableName), nil
}

// updateGeometrySRID implements UpdateGeometrySRID, which changes the SRID of a column that geometry_columns describes
// and of every value in it. The catalog name, when given, is ignored.
func updateGeometrySRID(ctx *sql.Context, runner sql.StatementRunner, args ...any) (any, error) {
	for len(args) < 5 {
		args = append([]any{""}, args...)
	}
	tableName, columnName, srid := args[2].(string), args[3].(string), args[4].(int32)
	schemaName, ok, err := managedSchema(ctx, runner, args[1].(string), tableName)
	if err != nil {
		return nil, err
	}
	var typeName string
	var dims int32
	if ok {
		if typeName, dims, ok, err = registeredColumn(ctx, runner, schemaName, tableName, columnName); err != nil {
			return nil, err
		}
	}
	if !ok {
		return nil, pgerror.New(pgcode.RaiseException, "column not found in geometry_columns table")
	}
	if srid != 0 {
		known, err := knownSRID(ctx, runner, srid)
		if err != nil {
			return nil, err
		}
		if !known {
			return nil, pgerror.Newf(pgcode.RaiseException, "invalid SRID: %d not found in spatial_ref_sys", srid)
		}
	}
	if strings.HasSuffix(typeName, "M") {
		dims = 3
	}
	modifier, err := managedTypeModifier(typeName, dims)
	if err != nil {
		return nil, err
	}
	column := hook.QuoteIdentifier(columnName)
	statement := fmt.Sprintf("ALTER TABLE %s.%s ALTER COLUMN %s TYPE geometry(%s, %d) USING ST_SetSRID(%s, %d)", hook.QuoteIdentifier(schemaName),
		hook.QuoteIdentifier(tableName), column, modifier, srid, column, srid)
	if err = hook.RunStatement(ctx, runner, statement); err != nil {
		return nil, err
	}
	return fmt.Sprintf("%s.%s.%s SRID changed to %d", schemaName, tableName, columnName, srid), nil
}

// findSRID implements Find_SRID, which returns the SRID that geometry_columns gives a column. A table name may carry
// its schema when the schema is empty.
func findSRID(ctx *sql.Context, runner sql.StatementRunner, args ...any) (any, error) {
	schemaName, tableName, columnName := args[0].(string), args[1].(string), args[2].(string)
	if schemaName == "" {
		if schema, table, ok := strings.Cut(tableName, "."); ok {
			schemaName, tableName = schema, table
		}
	}
	rows, err := queryRows(ctx, runner, fmt.Sprintf("SELECT srid FROM geometry_columns WHERE (f_table_schema = %s OR %s = '') AND f_table_name = %s AND f_geometry_column = %s",
		quoteLiteral(schemaName), quoteLiteral(schemaName), quoteLiteral(tableName), quoteLiteral(columnName)))
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, pgerror.New(pgcode.RaiseException, "find_srid() - could not find the corresponding SRID - is the geometry registered in the GEOMETRY_COLUMNS table?  Is there an uppercase/lowercase mismatch?")
	}
	return rows[0][0], nil
}

// populateColumns gives each geometry column of the tables that the condition selects, which lacks a type modifier,
// the type, number of dimensions, and SRID of its first value that is not null. A MULTI type converts the other
// values with ST_Multi. A column whose values do not all convert is skipped with a warning. Every column is checked
// before any is altered, since a failed statement undoes the changes that came before it. It returns the number of
// geometry columns that it looked at and the number that it converted.
func populateColumns(ctx *sql.Context, runner sql.StatementRunner, condition string, useTypmod bool) (int32, int32, error) {
	if !useTypmod {
		//TODO: add the CHECK constraints that describe the columns without a type modifier, once constraints may call
		// extension functions
		return 0, 0, pgerror.New(pgcode.FeatureNotSupported, "populate_geometry_columns without a type modifier is not yet supported")
	}
	columns, err := queryRows(ctx, runner, "SELECT n.nspname, c.relname, a.attname, postgis_typmod_type(a.atttypmod) FROM pg_class c "+
		"JOIN pg_attribute a ON a.attrelid = c.oid AND NOT a.attisdropped JOIN pg_namespace n ON c.relnamespace = n.oid "+
		"JOIN pg_type t ON a.atttypid = t.oid WHERE c.relkind IN ('r', 'p') AND t.typname = 'geometry' AND "+condition+
		" ORDER BY n.nspname, c.relname, a.attnum")
	if err != nil {
		return 0, 0, err
	}
	var statements []string
	for _, column := range columns {
		schemaName, tableName, columnName := column[0].(string), column[1].(string), column[2].(string)
		if column[3] != "Geometry" {
			continue
		}
		table := hook.QuoteIdentifier(schemaName) + "." + hook.QuoteIdentifier(tableName)
		quotedColumn := hook.QuoteIdentifier(columnName)
		rows, err := queryRows(ctx, runner, fmt.Sprintf("SELECT ST_SRID(%s), GeometryType(%s), ST_NDims(%s) FROM %s WHERE %s IS NOT NULL LIMIT 1",
			quotedColumn, quotedColumn, quotedColumn, table, quotedColumn))
		if err != nil {
			return 0, 0, err
		}
		if len(rows) == 0 {
			sendNotice(ctx, "WARNING", fmt.Sprintf("No data in table %s.%s, so no information to determine geometry type and srid", schemaName, tableName))
			continue
		}
		srid, typeName := rows[0][0].(int32), rows[0][1].(string)
		modifier, err := managedTypeModifier(typeName, int32(rows[0][2].(int16)))
		if err != nil {
			return 0, 0, err
		}
		converted := quotedColumn
		if strings.HasPrefix(typeName, "MULTI") {
			converted = "ST_Multi(" + quotedColumn + ")"
		}
		columnType := fmt.Sprintf("geometry(%s, %d)", modifier, srid)
		if _, err = queryRows(ctx, runner, fmt.Sprintf("SELECT count(CAST(%s AS %s)) FROM %s", converted, columnType, table)); err != nil {
			sendNotice(ctx, "WARNING", fmt.Sprintf("Could not convert '%s' in '%s.%s' to use typmod with srid %d, type %s: %s",
				columnName, schemaName, tableName, srid, modifier, err.Error()))
			continue
		}
		statements = append(statements, fmt.Sprintf("ALTER TABLE %s ALTER COLUMN %s TYPE %s USING %s", table, quotedColumn, columnType, converted))
	}
	for _, statement := range statements {
		if err = hook.RunStatement(ctx, runner, statement); err != nil {
			return 0, 0, err
		}
	}
	return int32(len(columns)), int32(len(statements)), nil
}

// populateGeometryColumns implements populate_geometry_columns(use_typmod), which converts the geometry columns of
// every table and reports how many it looked at and how many it converted.
func populateGeometryColumns(ctx *sql.Context, runner sql.StatementRunner, args ...any) (any, error) {
	probed, converted, err := populateColumns(ctx, runner, "true", args[0].(bool))
	if err != nil {
		return nil, err
	}
	return fmt.Sprintf("probed:%d inserted:%d", probed, converted), nil
}

// populateTableGeometryColumns implements populate_geometry_columns(tbl_oid, use_typmod), which converts the geometry
// columns of one table and returns how many it converted.
func populateTableGeometryColumns(ctx *sql.Context, runner sql.StatementRunner, args ...any) (any, error) {
	_, converted, err := populateColumns(ctx, runner, fmt.Sprintf("c.oid = %d", id.Cache().ToOID(args[0].(id.Id))), args[1].(bool))
	return converted, err
}

// constraintValue returns the value that a CHECK constraint on a column compares the result of the named function
// with, such as 4326 for `st_srid(geom) = 4326`, or nil when no constraint on the column calls the function.
func constraintValue(ctx *sql.Context, runner sql.StatementRunner, schemaName string, tableName string, columnName string, function string) (any, error) {
	rows, err := queryRows(ctx, runner, fmt.Sprintf("SELECT pg_get_constraintdef(s.oid) FROM pg_constraint s JOIN pg_class c ON s.conrelid = c.oid "+
		"JOIN pg_namespace n ON s.connamespace = n.oid JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum = ANY (s.conkey) "+
		"WHERE n.nspname = %s AND c.relname = %s AND a.attname = %s", quoteLiteral(schemaName), quoteLiteral(tableName), quoteLiteral(columnName)))
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		definition := row[0].(string)
		call := strings.Index(definition, function+"(")
		if call < 0 {
			continue
		}
		_, value, ok := strings.Cut(definition[call:], " = ")
		if !ok {
			continue
		}
		if function == "geometrytype" {
			_, quoted, _ := strings.Cut(value, "'")
			name, _, _ := strings.Cut(quoted, "'")
			return name, nil
		}
		number, err := strconv.ParseInt(strings.Trim(value, "()"), 10, 32)
		if err != nil {
			return nil, err
		}
		return int32(number), nil
	}
	return nil, nil
}

// constraintRoutine returns the implementation of a postgis_constraint function, which reads the value that a CHECK
// constraint on a column compares the result of the named function with.
func constraintRoutine(function string) extdef.QueryFunction {
	return func(ctx *sql.Context, runner sql.StatementRunner, args ...any) (any, error) {
		return constraintValue(ctx, runner, args[0].(string), args[1].(string), args[2].(string), function)
	}
}

// stFindExtent implements ST_FindExtent, which returns the extent of a column, given its table and optionally the
// table's schema first.
func stFindExtent(ctx *sql.Context, runner sql.StatementRunner, args ...any) (any, error) {
	table := hook.QuoteIdentifier(args[len(args)-2].(string))
	if len(args) == 3 {
		table = hook.QuoteIdentifier(args[0].(string)) + "." + table
	}
	rows, err := queryRows(ctx, runner, fmt.Sprintf("SELECT ST_Extent(%s) FROM %s", hook.QuoteIdentifier(args[len(args)-1].(string)), table))
	if err != nil {
		return nil, err
	}
	return rows[0][0], nil
}

// stEstimatedExtent implements ST_EstimatedExtent, which estimates the extent of a column from the statistics that
// ANALYZE gathers, and warns and returns null when there are none.
func stEstimatedExtent(ctx *sql.Context, args ...any) (any, error) {
	tableName, columnName := args[0].(string), args[1].(string)
	if len(args) > 2 {
		tableName, columnName = args[1].(string), args[2].(string)
	}
	//TODO: gather the statistics of geometry columns, which ANALYZE does not do yet
	sendNotice(ctx, "WARNING", fmt.Sprintf(`stats for "%s.%s" do not exist`, tableName, columnName))
	return nil, nil
}

// postgisExtensionsUpgrade implements postgis_extensions_upgrade, which updates PostGIS to the given version, or to
// its newest version when none is given. PostGIS passes through the version "ANY" on the way, which its errors name.
// The emulated version is the only one, so an update to it changes nothing.
func postgisExtensionsUpgrade(ctx *sql.Context, args ...any) (any, error) {
	sendNotice(ctx, "NOTICE", "Updating extension postgis "+postgisVersion)
	target := postgisVersion
	if args[0] != nil {
		target = args[0].(string)
	}
	if target == "" {
		return nil, pgerror.New(pgcode.Syntax, `zero-length delimited identifier at or near """"`)
	}
	if target != postgisVersion {
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, `extension "postgis" has no update path from version "ANY" to version "%s"`, target)
	}
	return fmt.Sprintf("Upgrade to version %s completed, run SELECT postgis_full_version(); for details", target), nil
}

// cacheBBoxTrigger implements postgis_cache_bbox for triggers, which PostGIS fires to cache the bounding boxes of the
// geometry column that the trigger's one argument names. Doltgres caches no bounding boxes, so it only checks the
// trigger and keeps the row. A trigger fired by DELETE, after the change, or once per statement does nothing but warn.
func cacheBBoxTrigger(ctx *sql.Context, sch sql.Schema, oldRow sql.Row, newRow sql.Row, variables map[string]any) (any, error) {
	args, _ := variables["TG_ARGV"].([]any)
	if len(args) != 1 {
		return nil, pgerror.New(pgcode.Internal, "trigger 'cache_bbox' must be called with one argument")
	}
	if variables["TG_OP"] == "DELETE" {
		sendNotice(ctx, "NOTICE", "Useless cache_box trigger fired by DELETE")
		return oldRow, nil
	}
	if variables["TG_WHEN"] == "AFTER" {
		sendNotice(ctx, "NOTICE", "Useless cache_box trigger fired AFTER")
		return newRow, nil
	}
	if variables["TG_LEVEL"] == "STATEMENT" {
		sendNotice(ctx, "NOTICE", "Useless cache_box trigger fired for STATEMENT")
		return nil, nil
	}
	column := args[0].(string)
	for _, col := range sch {
		if col.Name == column {
			if t, ok := col.Type.(*pgtypes.DoltgresType); !ok || t.Name() != "geometry" {
				return nil, pgerror.Newf(pgcode.Internal, "trigger %s requested to apply to a non-geometry field (%s)", variables["TG_NAME"], column)
			}
			return newRow, nil
		}
	}
	return nil, pgerror.Newf(pgcode.Internal, "trigger %s can't find attribute %s", variables["TG_NAME"], column)
}

// managementRoutines returns the routines that manage geometry columns.
func managementRoutines() []extdef.Routine {
	addParams := func(names ...string) []extdef.Parameter {
		ps := make([]extdef.Parameter, len(names))
		for i, name := range names {
			ps[i] = extdef.Parameter{Name: name, Type: "varchar"}
		}
		return append(ps, extdef.Parameter{Name: "new_srid", Type: "int4"}, extdef.Parameter{Name: "new_type", Type: "varchar"},
			extdef.Parameter{Name: "new_dim", Type: "int4"}, extdef.Parameter{Name: "use_typmod", Type: "bool", Default: `true`})
	}
	names := func(names ...string) []extdef.Parameter {
		ps := make([]extdef.Parameter, len(names))
		for i, name := range names {
			ps[i] = extdef.Parameter{Name: name, Type: "varchar"}
		}
		return ps
	}
	catalogAdd := addParams("catalog_name", "schema_name", "table_name", "column_name")
	catalogAdd[4].Name = "new_srid_in"
	return []extdef.Routine{
		{Name: "addgeometrycolumn", Symbol: "addgeometrycolumn_varchar_varchar_varchar_varchar_int4_varchar_int4_bool", Parameters: catalogAdd, Returns: "text", Strict: true, QueryImpl: addGeometryColumn},
		{Name: "addgeometrycolumn", Symbol: "addgeometrycolumn_varchar_varchar_varchar_int4_varchar_int4_bool", Parameters: addParams("schema_name", "table_name", "column_name"), Returns: "text", Strict: true, QueryImpl: addGeometryColumn},
		{Name: "addgeometrycolumn", Symbol: "addgeometrycolumn_varchar_varchar_int4_varchar_int4_bool", Parameters: addParams("table_name", "column_name"), Returns: "text", Strict: true, QueryImpl: addGeometryColumn},
		{Name: "dropgeometrycolumn", Symbol: "dropgeometrycolumn_varchar_varchar_varchar_varchar", Parameters: names("catalog_name", "schema_name", "table_name", "column_name"), Returns: "text", Strict: true, QueryImpl: dropGeometryColumn},
		{Name: "dropgeometrycolumn", Symbol: "dropgeometrycolumn_varchar_varchar_varchar", Parameters: names("schema_name", "table_name", "column_name"), Returns: "text", Strict: true, QueryImpl: dropGeometryColumn},
		{Name: "dropgeometrycolumn", Symbol: "dropgeometrycolumn_varchar_varchar", Parameters: names("table_name", "column_name"), Returns: "text", Strict: true, QueryImpl: dropGeometryColumn},
		{Name: "dropgeometrytable", Symbol: "dropgeometrytable_varchar_varchar_varchar", Parameters: names("catalog_name", "schema_name", "table_name"), Returns: "text", Strict: true, QueryImpl: dropGeometryTable},
		{Name: "dropgeometrytable", Symbol: "dropgeometrytable_varchar_varchar", Parameters: names("schema_name", "table_name"), Returns: "text", Strict: true, QueryImpl: dropGeometryTable},
		{Name: "dropgeometrytable", Symbol: "dropgeometrytable_varchar", Parameters: names("table_name"), Returns: "text", Strict: true, QueryImpl: dropGeometryTable},
		{Name: "find_srid", Symbol: "find_srid_varchar_varchar_varchar", Parameters: params("varchar", "varchar", "varchar"), Returns: "int4", Strict: true, QueryImpl: findSRID},
		{Name: "updategeometrysrid", Symbol: "updategeometrysrid_varchar_varchar_varchar_varchar_int4", Parameters: append(names("catalogn_name", "schema_name", "table_name", "column_name"), extdef.Parameter{Name: "new_srid_in", Type: "int4"}), Returns: "text", Strict: true, QueryImpl: updateGeometrySRID},
		{Name: "updategeometrysrid", Symbol: "updategeometrysrid_varchar_varchar_varchar_int4", Parameters: params("varchar", "varchar", "varchar", "int4"), Returns: "text", Strict: true, QueryImpl: updateGeometrySRID},
		{Name: "updategeometrysrid", Symbol: "updategeometrysrid_varchar_varchar_int4", Parameters: params("varchar", "varchar", "int4"), Returns: "text", Strict: true, QueryImpl: updateGeometrySRID},
		{Name: "populate_geometry_columns", Symbol: "populate_geometry_columns_bool", Parameters: []extdef.Parameter{{Name: "use_typmod", Type: "bool", Default: `true`}}, Returns: "text", QueryImpl: populateGeometryColumns},
		{Name: "populate_geometry_columns", Symbol: "populate_geometry_columns_oid_bool", Parameters: []extdef.Parameter{{Name: "tbl_oid", Type: "oid"}, {Name: "use_typmod", Type: "bool", Default: `true`}}, Returns: "int4", QueryImpl: populateTableGeometryColumns},
		{Name: "st_estimatedextent", Symbol: "st_estimatedextent_text_text", Parameters: params("text", "text"), Returns: "box2d", Strict: true, Impl: stEstimatedExtent},
		{Name: "st_estimatedextent", Symbol: "st_estimatedextent_text_text_text", Parameters: params("text", "text", "text"), Returns: "box2d", Strict: true, Impl: stEstimatedExtent},
		{Name: "st_estimatedextent", Symbol: "gserialized_estimated_extent", Parameters: params("text", "text", "text", "bool"), Returns: "box2d", Strict: true, Impl: stEstimatedExtent},
		{Name: "st_findextent", Symbol: "st_findextent_text_text", Parameters: params("text", "text"), Returns: "box2d", Strict: true, QueryImpl: stFindExtent},
		{Name: "st_findextent", Symbol: "st_findextent_text_text_text", Parameters: params("text", "text", "text"), Returns: "box2d", Strict: true, QueryImpl: stFindExtent},
		{Name: "postgis_constraint_dims", Symbol: "postgis_constraint_dims", Parameters: []extdef.Parameter{{Name: "geomschema", Type: "text"}, {Name: "geomtable", Type: "text"}, {Name: "geomcolumn", Type: "text"}}, Returns: "int4", Strict: true, QueryImpl: constraintRoutine("st_ndims")},
		{Name: "postgis_constraint_srid", Symbol: "postgis_constraint_srid", Parameters: []extdef.Parameter{{Name: "geomschema", Type: "text"}, {Name: "geomtable", Type: "text"}, {Name: "geomcolumn", Type: "text"}}, Returns: "int4", Strict: true, QueryImpl: constraintRoutine("st_srid")},
		{Name: "postgis_constraint_type", Symbol: "postgis_constraint_type", Parameters: []extdef.Parameter{{Name: "geomschema", Type: "text"}, {Name: "geomtable", Type: "text"}, {Name: "geomcolumn", Type: "text"}}, Returns: "varchar", Strict: true, QueryImpl: constraintRoutine("geometrytype")},
		{Name: "postgis_extensions_upgrade", Symbol: "postgis_extensions_upgrade", Parameters: []extdef.Parameter{{Name: "target_version", Type: "text", Default: "NULL::text"}}, Returns: "text", Impl: postgisExtensionsUpgrade},
		{Name: "postgis_cache_bbox", Symbol: "cache_bbox", Returns: "trigger", TriggerImpl: cacheBBoxTrigger, Impl: func(ctx *sql.Context, args ...any) (any, error) {
			return nil, pgerror.New(pgcode.Internal, "cache_bbox: not called by trigger manager")
		}},
	}
}
