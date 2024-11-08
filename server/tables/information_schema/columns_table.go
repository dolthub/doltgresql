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

package information_schema

import (
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/information_schema"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"

	partypes "github.com/dolthub/doltgresql/postgres/parser/types"
	"github.com/dolthub/doltgresql/server/functions"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// maxCharacterLength is the maximum character length for a column.
const maxCharacterOctetLength = 1073741824

// typeToNumericPrecision is a map of sqltypes to their respective numeric precision.
var typeToNumericPrecision = map[query.Type]int32{
	sqltypes.Int8:    3,
	sqltypes.Int16:   16,
	sqltypes.Int32:   32,
	sqltypes.Int64:   64,
	sqltypes.Float32: 24,
	sqltypes.Float64: 53,
}

// newColumnsTable creates a new information_schema.COLUMNS table.
func newColumnsTable() *information_schema.ColumnsTable {
	return &information_schema.ColumnsTable{
		TableName:   information_schema.ColumnsTableName,
		TableSchema: columnsSchema,
		RowIter:     columnsRowIter,
	}
}

// columnsSchema is the schema for the information_schema.columns table.
var columnsSchema = sql.Schema{
	{Name: "table_catalog", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "table_schema", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "table_name", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "column_name", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "ordinal_position", Type: cardinal_number, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "column_default", Type: character_data, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "is_nullable", Type: yes_or_no, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "data_type", Type: character_data, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "character_maximum_length", Type: cardinal_number, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "character_octet_length", Type: cardinal_number, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "numeric_precision", Type: cardinal_number, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "numeric_precision_radix", Type: cardinal_number, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "numeric_scale", Type: cardinal_number, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "datetime_precision", Type: cardinal_number, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "interval_type", Type: character_data, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "interval_precision", Type: cardinal_number, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "character_set_catalog", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "character_set_schema", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "character_set_name", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "collation_catalog", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "collation_schema", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "collation_name", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "domain_catalog", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "domain_schema", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "domain_name", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "udt_catalog", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "udt_schema", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "udt_name", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "scope_catalog", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "scope_schema", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "scope_name", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "maximum_cardinality", Type: cardinal_number, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "dtd_identifier", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "is_self_referencing", Type: yes_or_no, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "is_identity", Type: yes_or_no, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "identity_generation", Type: character_data, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "identity_start", Type: character_data, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "identity_increment", Type: character_data, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "identity_maximum", Type: character_data, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "identity_minimum", Type: character_data, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "identity_cycle", Type: yes_or_no, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "is_generated", Type: character_data, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "generation_expression", Type: character_data, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
	{Name: "is_updatable", Type: yes_or_no, Default: nil, Nullable: true, Source: information_schema.ColumnsTableName},
}

// columnsRowIter implements the custom sql.RowIter for the information_schema.columns table.
func columnsRowIter(ctx *sql.Context, catalog sql.Catalog, allColsWithDefaultValue sql.Schema) (sql.RowIter, error) {
	var rows []sql.Row

	databases, err := information_schema.AllDatabasesWithNames(ctx, catalog, false)
	if err != nil {
		return nil, err
	}

	for _, db := range databases {
		rs, err := getRowsFromDatabase(ctx, db, allColsWithDefaultValue)
		if err != nil {
			return nil, err
		}
		rows = append(rows, rs...)

		rs, err = getRowsFromViews(ctx, db)
		if err != nil {
			return nil, err
		}
		rows = append(rows, rs...)
	}
	return sql.RowsToRowIter(rows...), nil
}

// getRowFromColumn returns a single row for given column. The arguments passed
// are used to define all row values. These include the current ordinal
// position, so this column will get the next position number, sql.Column
// object, database name, and table name.
func getRowFromColumn(ctx *sql.Context, curOrdPos int, col *sql.Column, catName, schName, tblName string) sql.Row {
	var (
		ordinalPos  = int32(curOrdPos + 1)
		nullable    = "NO"
		isGenerated = "NEVER"
	)

	dataType, udtName := getDataAndUdtType(col.Type, col.Name)

	if col.Nullable {
		nullable = "YES"
	}
	if col.Generated != nil {
		isGenerated = "ALWAYS"
	}

	charName, collName, charMaxLen, charOctetLen := getCharAndCollNamesAndCharMaxAndOctetLens(ctx, col.Type)
	numericPrecision, numericPrecisionRadix, numericScale := getColumnPrecisionAndScale(col.Type)
	datetimePrecision := getDatetimePrecision(col.Type)

	columnDefault := information_schema.GetColumnDefault(ctx, col.Default)

	return sql.Row{
		catName,               // table_catalog
		schName,               // table_schema
		tblName,               // table_name
		col.Name,              // column_name
		ordinalPos,            // ordinal_position
		columnDefault,         // column_default
		nullable,              // is_nullable
		dataType,              // data_type
		charMaxLen,            // character_maximum_length
		charOctetLen,          // character_octet_length
		numericPrecision,      // numeric_precision
		numericPrecisionRadix, // numeric_precision_radix
		numericScale,          // numeric_scale
		datetimePrecision,     // datetime_precision
		nil,                   // interval_type TODO
		nil,                   // interval_precision TODO
		nil,                   // character_set_catalog TODO
		nil,                   // character_set_schema TODO
		charName,              // character_set_name
		nil,                   // collation_catalog TODO
		nil,                   // collation_schema TODO
		collName,              // collation_name
		nil,                   // domain_catalog TODO
		nil,                   // domain_schema TODO
		nil,                   // domain_name TODO
		catName,               // udt_catalog
		"pg_catalog",          // udt_schema
		udtName,               // udt_name
		nil,                   // scope_catalog TODO
		nil,                   // scope_schema TODO
		nil,                   // scope_name TODO
		nil,                   // maximum_cardinality TODO
		nil,                   // dtd_identifier TODO
		"NO",                  // is_self_referencing TODO
		"NO",                  // is_identity TODO
		nil,                   // identity_generation TODO
		nil,                   // identity_start TODO
		nil,                   // identity_increment TODO
		nil,                   // identity_maximum TODO
		nil,                   // identity_minimum TODO
		"NO",                  // identity_cycle TODO
		isGenerated,           // is_generated
		nil,                   // generation_expression TODO
		"YES",                 // is_updatable
	}
}

// getRowsFromTable returns array of rows for all accessible columns of the given table.
func getRowsFromTable(ctx *sql.Context, db information_schema.DbWithNames, t sql.Table, allColsWithDefaultValue sql.Schema) ([]sql.Row, error) {
	var rows []sql.Row

	tblName := t.Name()
	for i, col := range information_schema.SchemaForTable(t, db.Database, allColsWithDefaultValue) {
		r := getRowFromColumn(ctx, i, col, db.CatalogName, db.SchemaName, tblName)
		if r != nil {
			rows = append(rows, r)
		}
	}

	return rows, nil
}

// getRowsFromViews returns array or rows for columns for all views for given database.
func getRowsFromViews(ctx *sql.Context, db information_schema.DbWithNames) ([]sql.Row, error) {
	var rows []sql.Row
	// TODO: View Definition is lacking information to properly fill out these table
	// TODO: Should somehow get reference to table(s) view is referencing
	// TODO: Each column that view references should also show up as unique entries as well
	views, err := information_schema.ViewsInDatabase(ctx, db.Database)
	if err != nil {
		return nil, err
	}

	for _, view := range views {
		rows = append(rows, sql.Row{
			db.CatalogName, // table_catalog
			db.SchemaName,  // table_schema
			view.Name,      // table_name
			"",             // column_name
			int32(0),       // ordinal_position
			nil,            // column_default
			"YES",          // is_nullable
			nil,            // data_type
			nil,            // character_maximum_length
			nil,            // character_octet_length
			nil,            // numeric_precision
			nil,            // numeric_precision_radix
			nil,            // numeric_scale
			nil,            // datetime_precision
			nil,            // interval_type
			nil,            // interval_precision
			nil,            // character_set_catalog
			nil,            // character_set_schema
			nil,            // character_set_name
			nil,            // collation_catalog
			nil,            // collation_schema
			nil,            // collation_name
			nil,            // domain_catalog
			nil,            // domain_schema
			nil,            // domain_name
			nil,            // udt_catalog
			nil,            // udt_schema
			nil,            // udt_name
			nil,            // scope_catalog
			nil,            // scope_schema
			nil,            // scope_name
			nil,            // maximum_cardinality
			nil,            // dtd_identifier
			"NO",           // is_self_referencing
			"NO",           // is_identity
			nil,            // identity_generation
			nil,            // identity_start
			nil,            // identity_increment
			nil,            // identity_maximum
			nil,            // identity_minimum
			"NO",           // identity_cycle
			"NO",           // is_generated
			nil,            // generation_expression
			"YES",          // is_updatable
		})
	}

	return rows, nil
}

// getRowsFromDatabase returns array of rows for all accessible columns of accessible table of the given database.
func getRowsFromDatabase(ctx *sql.Context, db information_schema.DbWithNames, allColsWithDefaultValue sql.Schema) ([]sql.Row, error) {
	var rows []sql.Row

	err := sql.DBTableIter(ctx, db.Database, func(t sql.Table) (cont bool, err error) {
		rs, err := getRowsFromTable(ctx, db, t, allColsWithDefaultValue)
		if err != nil {
			return false, err
		}
		rows = append(rows, rs...)
		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// getDataAndUdtType returns data types for given DoltgresType. udt_name is the
// base name of the type (i.e. "varchar"). data_type is the SQL standard name of
// the type (i.e. "character varying").
func getDataAndUdtType(colType sql.Type, colName string) (string, string) {
	udtName := ""
	dataType := ""
	dgType, ok := colType.(pgtypes.DoltgresType)
	if ok {
		udtName = dgType.Name
		if t, ok := partypes.OidToType[oid.Oid(dgType.OID)]; ok {
			dataType = t.SQLStandardName()
		}
	} else {
		dtdId := strings.Split(strings.Split(colType.String(), " COLLATE")[0], " CHARACTER SET")[0]

		// The DATA_TYPE value is the type name only with no other information
		dataType = strings.Split(dtdId, "(")[0]
		dataType = strings.Split(dataType, " ")[0]
		udtName = dataType
	}
	return dataType, udtName
}

// getColumnPrecisionAndScale returns the precision or a number of postgres type. For non-numeric or decimal types this
// function should return nil,nil.
func getColumnPrecisionAndScale(colType sql.Type) (interface{}, interface{}, interface{}) {
	dgt, ok := colType.(pgtypes.DoltgresType)
	if ok {
		switch oid.Oid(dgt.OID) {
		// TODO: BitType
		case oid.T_float4, oid.T_float8:
			return typeToNumericPrecision[colType.Type()], int32(2), nil
		case oid.T_int2, oid.T_int4, oid.T_int8:
			return typeToNumericPrecision[colType.Type()], int32(2), int32(0)
		case oid.T_numeric:
			var precision interface{}
			var scale interface{}
			if dgt.AttTypMod != -1 {
				precision, scale = functions.GetPrecisionAndScaleFromTypmod(dgt.AttTypMod)
			}
			return precision, int32(10), scale
		default:
			return nil, nil, nil
		}
	}
	return nil, nil, nil
}

// getCharAndCollNamesAndCharMaxAndOctetLens returns the character set name,
// collation name, character maximum length and character octet length
func getCharAndCollNamesAndCharMaxAndOctetLens(ctx *sql.Context, colType sql.Type) (interface{}, interface{}, interface{}, interface{}) {
	var (
		charName     interface{}
		collName     interface{}
		charMaxLen   interface{}
		charOctetLen interface{}
	)
	// TODO: This doesn't work for doltgres types
	if twc, ok := colType.(sql.TypeWithCollation); ok && !types.IsBinaryType(colType) {
		colColl := twc.Collation()
		collName = colColl.Name()
		charName = colColl.CharacterSet().String()
		if types.IsEnum(colType) || types.IsSet(colType) {
			charOctetLen = int32(colType.MaxTextResponseByteLength(ctx))
			charMaxLen = int32(colType.MaxTextResponseByteLength(ctx)) / int32(colColl.CharacterSet().MaxLength())
		}
	}

	switch t := colType.(type) {
	case pgtypes.DoltgresType:
		if t.TypCategory == pgtypes.TypeCategory_StringTypes {
			if t.AttTypMod == -1 {
				charOctetLen = int32(maxCharacterOctetLength)
			} else {
				l := pgtypes.GetMaxCharsFromTypmod(t.AttTypMod)
				charOctetLen = l * 4
				charMaxLen = l
			}
		}
	}

	return charName, collName, charMaxLen, charOctetLen
}

func getDatetimePrecision(colType sql.Type) interface{} {
	if dgType, ok := colType.(pgtypes.DoltgresType); ok {
		switch oid.Oid(dgType.OID) {
		case oid.T_date:
			return int32(0)
		case oid.T_time, oid.T_timetz, oid.T_timestamp, oid.T_timestamptz:
			// TODO: TIME length not yet supported
			return int32(6)
		default:
			return nil
		}
	}
	return nil
}
