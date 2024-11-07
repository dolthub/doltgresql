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

	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/server/types/oid"
)

// PgTypeName is a constant to the pg_type name.
const PgTypeName = "pg_type"

// InitPgType handles registration of the pg_type handler.
func InitPgType() {
	tables.AddHandler(PgCatalogName, PgTypeName, PgTypeHandler{})
}

// PgTypeHandler is the handler for the pg_type table.
type PgTypeHandler struct{}

var _ tables.Handler = PgTypeHandler{}

// Name implements the interface tables.Handler.
func (p PgTypeHandler) Name() string {
	return PgTypeName
}

// RowIter implements the interface tables.Handler.
func (p PgTypeHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	var pgCatalogOid uint32
	err := oid.IterateCurrentDatabase(ctx, oid.Callbacks{
		Schema: func(ctx *sql.Context, schema oid.ItemSchema) (cont bool, err error) {
			if schema.Item.SchemaName() == PgCatalogName {
				pgCatalogOid = schema.OID
				return false, nil
			}
			return true, nil
		},
	})
	if err != nil {
		return nil, err
	}

	var displayTypes []pgtypes.DoltgresType
	err = oid.IterateCurrentDatabase(ctx, oid.Callbacks{
		Type: func(ctx *sql.Context, typ oid.ItemType) (cont bool, err error) {
			displayTypes = append(displayTypes, typ.Item)
			return true, nil
		},
	})
	if err != nil {
		return nil, err
	}
	return &pgTypeRowIter{pgCatalogOid: pgCatalogOid, types: displayTypes, idx: 0}, nil
}

// Schema implements the interface tables.Handler.
func (p PgTypeHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgTypeSchema,
		PkOrdinals: nil,
	}
}

// pgTypeSchema is the schema for pg_type.
var pgTypeSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typnamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typlen", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typbyval", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typtype", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typcategory", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typispreferred", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typisdefined", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typdelim", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typsubscript", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgTypeName}, // TODO: type regproc
	{Name: "typelem", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typarray", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typinput", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgTypeName},   // TODO: type regproc
	{Name: "typoutput", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgTypeName},  // TODO: type regproc
	{Name: "typreceive", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgTypeName}, // TODO: type regproc
	{Name: "typsend", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgTypeName},    // TODO: type regproc
	{Name: "typmodin", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgTypeName},   // TODO: type regproc
	{Name: "typmodout", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgTypeName},  // TODO: type regproc
	{Name: "typanalyze", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgTypeName}, // TODO: type regproc
	{Name: "typalign", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typstorage", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typnotnull", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typbasetype", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typtypmod", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typndims", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typcollation", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typdefaultbin", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgTypeName}, // TODO: type pg_node_tree, collation C
	{Name: "typdefault", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgTypeName},    // TODO: collation C
	{Name: "typacl", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgTypeName},   // TODO: type aclitem[]
}

// pgTypeRowIter is the sql.RowIter for the pg_type table.
type pgTypeRowIter struct {
	pgCatalogOid uint32
	types        []pgtypes.DoltgresType
	idx          int
}

var _ sql.RowIter = (*pgTypeRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgTypeRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.types) {
		return nil, io.EOF
	}
	iter.idx++
	typ := iter.types[iter.idx-1]
	// TODO: typ.Acl is stored as []string
	typAcl := []any(nil)

	return sql.Row{
		typ.OID,                 //oid
		typ.Name,                //typname
		iter.pgCatalogOid,       //typnamespace
		uint32(0),               //typowner
		typ.TypLength,           //typlen
		typ.PassedByVal,         //typbyval
		string(typ.TypType),     //typtype
		string(typ.TypCategory), //typcategory
		typ.IsPreferred,         //typispreferred
		typ.IsDefined,           //typisdefined
		typ.Delimiter,           //typdelim
		typ.RelID,               //typrelid
		typ.SubscriptFunc,       //typsubscript
		typ.Elem,                //typelem
		typ.Array,               //typarray
		typ.InputFunc,           //typinput
		typ.OutputFunc,          //typoutput
		typ.ReceiveFunc,         //typreceive
		typ.SendFunc,            //typsend
		typ.ModInFunc,           //typmodin
		typ.ModOutFunc,          //typmodout
		typ.AnalyzeFunc,         //typanalyze
		string(typ.Align),       //typalign
		string(typ.Storage),     //typstorage
		typ.NotNull,             //typnotnull
		typ.BaseTypeOID,         //typbasetype
		typ.TypMod,              //typtypmod
		typ.NDims,               //typndims
		typ.TypCollation,        //typcollation
		typ.DefaulBin,           //typdefaultbin
		typ.Default,             //typdefault
		typAcl,                  //typacl
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgTypeRowIter) Close(ctx *sql.Context) error {
	return nil
}
