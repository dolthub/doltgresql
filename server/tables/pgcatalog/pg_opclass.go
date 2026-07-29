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
	"io"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgOpclassName is a constant to the pg_opclass name.
const PgOpclassName = "pg_opclass"

// InitPgOpclass handles registration of the pg_opclass handler.
func InitPgOpclass() {
	tables.AddHandler(PgCatalogName, PgOpclassName, PgOpclassHandler{})
}

// PgOpclassHandler is the handler for the pg_opclass table.
type PgOpclassHandler struct{}

var _ tables.Handler = PgOpclassHandler{}

// Name implements the interface tables.Handler.
func (p PgOpclassHandler) Name() string {
	return PgOpclassName
}

// RowIter implements the interface tables.Handler.
func (p PgOpclassHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	return &pgOpclassRowIter{
		classes: defaultOperatorClasses,
		idx:     0,
	}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgOpclassHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgOpclassSchema,
		PkOrdinals: nil,
	}
}

// pgOpclassSchema is the schema for pg_opclass.
var pgOpclassSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOpclassName},
	{Name: "opcmethod", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOpclassName},
	{Name: "opcname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgOpclassName},
	{Name: "opcnamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOpclassName},
	{Name: "opcowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOpclassName},
	{Name: "opcfamily", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOpclassName},
	{Name: "opcintype", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOpclassName},
	{Name: "opcdefault", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgOpclassName},
	{Name: "opckeytype", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOpclassName},
}

// operatorClass describes a built-in operator class.
type operatorClass struct {
	am         string
	name       string
	familyName string // the operator family this class belongs to, within the same access method
	inputType  *pgtypes.DoltgresType
	isDefault  bool
}

// oid returns the ID of this operator class, whose OIDs are registered in core/id/cache_operator_class_defaults.go.
func (c operatorClass) oid() id.Id {
	return id.NewId(id.Section_OperatorClass, c.am, c.name)
}

// defaultOperatorClasses is the list of built-in operator classes available in Postgres for the access methods and
// types that Doltgres supports. Unlike operator families, Postgres assigns most operator class OIDs dynamically during
// initdb, so Doltgres assigns its own fixed OIDs (registered in core/id/cache_operator_class_defaults.go).
// TODO: Postgres defines more operator classes (gin, gist, brin, spgist, and additional types); add them as the
// related types and access methods gain support.
var defaultOperatorClasses = []operatorClass{
	{am: "btree", name: "array_ops", familyName: "array_ops", inputType: pgtypes.AnyArray, isDefault: true},
	{am: "btree", name: "bool_ops", familyName: "bool_ops", inputType: pgtypes.Bool, isDefault: true},
	{am: "btree", name: "bpchar_ops", familyName: "bpchar_ops", inputType: pgtypes.BpChar, isDefault: true},
	{am: "btree", name: "bytea_ops", familyName: "bytea_ops", inputType: pgtypes.Bytea, isDefault: true},
	{am: "btree", name: "char_ops", familyName: "char_ops", inputType: pgtypes.InternalChar, isDefault: true},
	{am: "btree", name: "date_ops", familyName: "datetime_ops", inputType: pgtypes.Date, isDefault: true},
	{am: "btree", name: "float4_ops", familyName: "float_ops", inputType: pgtypes.Float32, isDefault: true},
	{am: "btree", name: "float8_ops", familyName: "float_ops", inputType: pgtypes.Float64, isDefault: true},
	{am: "btree", name: "int2_ops", familyName: "integer_ops", inputType: pgtypes.Int16, isDefault: true},
	{am: "btree", name: "int4_ops", familyName: "integer_ops", inputType: pgtypes.Int32, isDefault: true},
	{am: "btree", name: "int8_ops", familyName: "integer_ops", inputType: pgtypes.Int64, isDefault: true},
	{am: "btree", name: "interval_ops", familyName: "interval_ops", inputType: pgtypes.Interval, isDefault: true},
	{am: "btree", name: "jsonb_ops", familyName: "jsonb_ops", inputType: pgtypes.JsonB, isDefault: true},
	{am: "btree", name: "name_ops", familyName: "text_ops", inputType: pgtypes.Name, isDefault: true},
	{am: "btree", name: "numeric_ops", familyName: "numeric_ops", inputType: pgtypes.Numeric, isDefault: true},
	{am: "btree", name: "oid_ops", familyName: "oid_ops", inputType: pgtypes.Oid, isDefault: true},
	{am: "btree", name: "record_ops", familyName: "record_ops", inputType: pgtypes.Record, isDefault: true},
	{am: "btree", name: "text_ops", familyName: "text_ops", inputType: pgtypes.Text, isDefault: true},
	{am: "btree", name: "time_ops", familyName: "time_ops", inputType: pgtypes.Time, isDefault: true},
	{am: "btree", name: "timestamp_ops", familyName: "datetime_ops", inputType: pgtypes.Timestamp, isDefault: true},
	{am: "btree", name: "timestamptz_ops", familyName: "datetime_ops", inputType: pgtypes.TimestampTZ, isDefault: true},
	{am: "btree", name: "timetz_ops", familyName: "timetz_ops", inputType: pgtypes.TimeTZ, isDefault: true},
	{am: "btree", name: "uuid_ops", familyName: "uuid_ops", inputType: pgtypes.Uuid, isDefault: true},
	// varchar_ops operates on text, matching Postgres (varchar has no operators of its own)
	{am: "btree", name: "varchar_ops", familyName: "text_ops", inputType: pgtypes.Text, isDefault: false},
	{am: "hash", name: "bool_ops", familyName: "bool_ops", inputType: pgtypes.Bool, isDefault: true},
	{am: "hash", name: "bpchar_ops", familyName: "bpchar_ops", inputType: pgtypes.BpChar, isDefault: true},
	{am: "hash", name: "bytea_ops", familyName: "bytea_ops", inputType: pgtypes.Bytea, isDefault: true},
	{am: "hash", name: "char_ops", familyName: "char_ops", inputType: pgtypes.InternalChar, isDefault: true},
	{am: "hash", name: "date_ops", familyName: "datetime_ops", inputType: pgtypes.Date, isDefault: true},
	{am: "hash", name: "float4_ops", familyName: "float_ops", inputType: pgtypes.Float32, isDefault: true},
	{am: "hash", name: "float8_ops", familyName: "float_ops", inputType: pgtypes.Float64, isDefault: true},
	{am: "hash", name: "int2_ops", familyName: "integer_ops", inputType: pgtypes.Int16, isDefault: true},
	{am: "hash", name: "int4_ops", familyName: "integer_ops", inputType: pgtypes.Int32, isDefault: true},
	{am: "hash", name: "int8_ops", familyName: "integer_ops", inputType: pgtypes.Int64, isDefault: true},
	{am: "hash", name: "jsonb_ops", familyName: "jsonb_ops", inputType: pgtypes.JsonB, isDefault: true},
	{am: "hash", name: "numeric_ops", familyName: "numeric_ops", inputType: pgtypes.Numeric, isDefault: true},
	{am: "hash", name: "oid_ops", familyName: "oid_ops", inputType: pgtypes.Oid, isDefault: true},
	{am: "hash", name: "text_ops", familyName: "text_ops", inputType: pgtypes.Text, isDefault: true},
	{am: "hash", name: "timestamp_ops", familyName: "datetime_ops", inputType: pgtypes.Timestamp, isDefault: true},
	{am: "hash", name: "timestamptz_ops", familyName: "datetime_ops", inputType: pgtypes.TimestampTZ, isDefault: true},
	{am: "hash", name: "uuid_ops", familyName: "uuid_ops", inputType: pgtypes.Uuid, isDefault: true},
	// varchar_ops operates on text, matching Postgres (varchar has no operators of its own)
	{am: "hash", name: "varchar_ops", familyName: "text_ops", inputType: pgtypes.Text, isDefault: false},
}

// pgOpclassRowIter is the sql.RowIter for the pg_opclass table.
type pgOpclassRowIter struct {
	classes []operatorClass
	idx     int
}

var _ sql.RowIter = (*pgOpclassRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgOpclassRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.classes) {
		return nil, io.EOF
	}
	iter.idx++
	class := iter.classes[iter.idx-1]

	return sql.Row{
		class.oid(),                          // oid
		id.NewAccessMethod(class.am).AsId(),  // opcmethod
		class.name,                           // opcname
		id.NewNamespace("pg_catalog").AsId(), // opcnamespace
		id.Null,                              // opcowner (TODO: object ownership is not tracked)
		operatorFamily{am: class.am, name: class.familyName}.oid(), // opcfamily
		class.inputType.ID.AsId(),                                  // opcintype
		class.isDefault,                                            // opcdefault
		id.Null,                                                    // opckeytype
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgOpclassRowIter) Close(ctx *sql.Context) error {
	return nil
}
