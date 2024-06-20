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
	"math"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
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
	// TODO: Implement pg_type row iter
	doltgresTypes := pgtypes.GetAllPgTypes()
	types := make([]pgtypes.DoltgresType, 0, len(doltgresTypes))
	typNames := make([]string, 0, len(doltgresTypes))
	for name, typ := range doltgresTypes {
		types = append(types, typ)
		typNames = append(typNames, name)
	}
	return &pgTypeRowIter{types: types, typNames: typNames, idx: 0}, nil
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
	{Name: "typtype", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typcategory", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typispreferred", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typisdefined", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typdelim", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgTypeName},
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
	{Name: "typalign", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgTypeName},
	{Name: "typstorage", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgTypeName},
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
	types    []pgtypes.DoltgresType
	typNames []string
	idx      int
}

var _ sql.RowIter = (*pgTypeRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgTypeRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.types) {
		return nil, io.EOF
	}
	iter.idx++
	typ := iter.types[iter.idx-1]
	typName := iter.typNames[iter.idx-1]

	var (
		typLen     int16
		typByVal   = false
		typAlign   = pgtypes.TypeAlignment_Double
		typStorage = "p"
	)
	if l := typ.MaxTextResponseByteLength(ctx); l == math.MaxUint32 {
		typLen = -1
	} else {
		typLen = int16(l)
		// TODO: below can be of different value for some exceptions
		typByVal = true
		typStorage = "x"
	}

	// TODO: use the type information to fill these rather than manually doing it
	switch typ.(type) {
	case pgtypes.UnknownType:
		typLen = -2
		typAlign = pgtypes.TypeAlignment_Char
	case pgtypes.NumericType:
		typStorage = "m"
		typAlign = pgtypes.TypeAlignment_Int
	case pgtypes.BoolType, pgtypes.CharType, pgtypes.NameType, pgtypes.UuidType:
		typAlign = pgtypes.TypeAlignment_Char
	case pgtypes.Int16Type:
		typAlign = pgtypes.TypeAlignment_Short
	case pgtypes.ByteaType, pgtypes.Int32Type, pgtypes.TextType, pgtypes.OidType, pgtypes.XidType,
		pgtypes.JsonType, pgtypes.Float32Type, pgtypes.VarCharType, pgtypes.DateType, pgtypes.JsonBType:
		typAlign = pgtypes.TypeAlignment_Int
	}

	typCategory := typ.BaseID().GetTypeCategory()
	typIsPreferred := typ.BaseID() == typCategory.GetPreferredType()

	// TODO: fix some types get underscore as spacing (e.g. uuid_in, json_in, etc.)
	typIn := fmt.Sprintf("%sin", typName)
	typOut := fmt.Sprintf("%sout", typName)
	typRec := fmt.Sprintf("%srec", typName)
	typSend := fmt.Sprintf("%ssend", typName)

	// TODO: not all columns are populated
	return sql.Row{
		typ.OID(),           // oid
		typName,             //typname
		uint32(0),           //typnamespace
		uint32(0),           //typowner
		typLen,              //typlen
		typByVal,            //typbyval
		"b",                 //typtype
		string(typCategory), //typcategory
		typIsPreferred,      //typispreferred
		true,                //typisdefined
		",",                 //typdelim
		uint32(0),           //typrelid
		"-",                 //typsubscript
		uint32(0),           //typelem
		uint32(0),           //typarray
		typIn,               //typinput
		typOut,              //typoutput
		typRec,              //typreceive
		typSend,             //typsend
		"-",                 //typmodin
		"-",                 //typmodout
		"-",                 //typanalyze
		string(typAlign),    //typalign
		typStorage,          //typstorage
		false,               //typnotnull
		uint32(0),           //typbasetype
		int32(0),            //typtypmod
		int32(0),            //typndims
		uint32(0),           //typcollation
		nil,                 //typdefaultbin
		nil,                 //typdefault
		nil,                 //typacl
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgTypeRowIter) Close(ctx *sql.Context) error {
	return nil
}
