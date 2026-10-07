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

package types

import (
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
)

// PgAttribute is the row type of pg_catalog.pg_attribute.
var PgAttribute = newCatalogRowType("pg_attribute", PgAttributeSchema)

// PgType is the row type of pg_catalog.pg_type.
var PgType = newCatalogRowType("pg_type", PgTypeSchema)

// newCatalogRowType uses the catalog table's schema so its composite fields have
// the same types and order as whole-row references to the table.
func newCatalogRowType(name string, schema sql.Schema) *DoltgresType {
	relID := id.NewTable("pg_catalog", name).AsId()
	attrs := make([]CompositeAttribute, len(schema))
	for i, col := range schema {
		attrs[i] = NewCompositeAttribute(nil, relID, col.Name, col.Type.(*DoltgresType), int16(i+1), "")
	}
	return NewCompositeType(nil, relID, nil, id.NewType("pg_catalog", name), attrs)
}

// PgAttributeSchema is the schema for pg_catalog.pg_attribute.
var PgAttributeSchema = sql.Schema{
	{Name: "attrelid", Type: Oid, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "attname", Type: Name, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "atttypid", Type: Oid, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "attlen", Type: Int16, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "attnum", Type: Int16, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "attcacheoff", Type: Int32, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "atttypmod", Type: Int32, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "attndims", Type: Int16, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "attbyval", Type: Bool, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "attalign", Type: InternalChar, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "attstorage", Type: InternalChar, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "attcompression", Type: InternalChar, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "attnotnull", Type: Bool, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "atthasdef", Type: Bool, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "atthasmissing", Type: Bool, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "attidentity", Type: InternalChar, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "attgenerated", Type: InternalChar, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "attisdropped", Type: Bool, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "attislocal", Type: Bool, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "attinhcount", Type: Int16, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "attstattarget", Type: Int16, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "attcollation", Type: Oid, Default: nil, Nullable: false, Source: "pg_attribute"},
	{Name: "attacl", Type: TextArray, Default: nil, Nullable: true, Source: "pg_attribute"},        // TODO: type aclitem[]
	{Name: "attoptions", Type: TextArray, Default: nil, Nullable: true, Source: "pg_attribute"},    // TODO: collation C
	{Name: "attfdwoptions", Type: TextArray, Default: nil, Nullable: true, Source: "pg_attribute"}, // TODO: collation C
	{Name: "attmissingval", Type: AnyArray, Default: nil, Nullable: true, Source: "pg_attribute"},
}

// PgTypeSchema is the schema for pg_catalog.pg_type.
var PgTypeSchema = sql.Schema{
	{Name: "oid", Type: Oid, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typname", Type: Name, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typnamespace", Type: Oid, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typowner", Type: Oid, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typlen", Type: Int16, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typbyval", Type: Bool, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typtype", Type: InternalChar, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typcategory", Type: InternalChar, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typispreferred", Type: Bool, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typisdefined", Type: Bool, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typdelim", Type: InternalChar, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typrelid", Type: Oid, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typsubscript", Type: Regproc, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typelem", Type: Oid, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typarray", Type: Oid, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typinput", Type: Regproc, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typoutput", Type: Regproc, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typreceive", Type: Regproc, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typsend", Type: Regproc, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typmodin", Type: Regproc, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typmodout", Type: Regproc, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typanalyze", Type: Regproc, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typalign", Type: InternalChar, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typstorage", Type: InternalChar, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typnotnull", Type: Bool, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typbasetype", Type: Oid, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typtypmod", Type: Int32, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typndims", Type: Int32, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typcollation", Type: Oid, Default: nil, Nullable: false, Source: "pg_type"},
	{Name: "typdefaultbin", Type: Text, Default: nil, Nullable: true, Source: "pg_type"}, // TODO: type pg_node_tree, collation C
	{Name: "typdefault", Type: Text, Default: nil, Nullable: true, Source: "pg_type"},    // TODO: collation C
	{Name: "typacl", Type: TextArray, Default: nil, Nullable: true, Source: "pg_type"},   // TODO: type aclitem[]
}
