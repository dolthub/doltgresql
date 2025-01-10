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

package types

import (
	"sort"

	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/types"
)

// TypeAlignment represents the alignment required when storing a value of this type.
type TypeAlignment string

const (
	TypeAlignment_Char   TypeAlignment = "c"
	TypeAlignment_Short  TypeAlignment = "s"
	TypeAlignment_Int    TypeAlignment = "i"
	TypeAlignment_Double TypeAlignment = "d"
)

// TypeCategory represents the type category that a type belongs to. These are used by Postgres to group similar types
// for parameter resolution, operator resolution, etc.
type TypeCategory string

const (
	TypeCategory_ArrayTypes          TypeCategory = "A"
	TypeCategory_BooleanTypes        TypeCategory = "B"
	TypeCategory_CompositeTypes      TypeCategory = "C"
	TypeCategory_DateTimeTypes       TypeCategory = "D"
	TypeCategory_EnumTypes           TypeCategory = "E"
	TypeCategory_GeometricTypes      TypeCategory = "G"
	TypeCategory_NetworkAddressTypes TypeCategory = "I"
	TypeCategory_NumericTypes        TypeCategory = "N"
	TypeCategory_PseudoTypes         TypeCategory = "P"
	TypeCategory_RangeTypes          TypeCategory = "R"
	TypeCategory_StringTypes         TypeCategory = "S"
	TypeCategory_TimespanTypes       TypeCategory = "T"
	TypeCategory_UserDefinedTypes    TypeCategory = "U"
	TypeCategory_BitStringTypes      TypeCategory = "V"
	TypeCategory_UnknownTypes        TypeCategory = "X"
	TypeCategory_InternalUseTypes    TypeCategory = "Z"
)

// TypeStorage represents the storage strategy for storing `varlena` (typlen = -1) types.
type TypeStorage string

const (
	TypeStorage_Plain    TypeStorage = "p"
	TypeStorage_External TypeStorage = "e"
	TypeStorage_Main     TypeStorage = "m"
	TypeStorage_Extended TypeStorage = "x"
)

// TypeType represents the type of types that can be created/used.
// This includes 'base', 'composite', 'domain', 'enum', 'shell', 'range' and  'multirange'.
type TypeType string

const (
	TypeType_Base       TypeType = "b"
	TypeType_Composite  TypeType = "c"
	TypeType_Domain     TypeType = "d"
	TypeType_Enum       TypeType = "e"
	TypeType_Pseudo     TypeType = "p"
	TypeType_Range      TypeType = "r"
	TypeType_MultiRange TypeType = "m"
)

// GetTypeByID returns the DoltgresType matching the given Internal ID.
// If the Internal ID does not match a type, then nil is returned.
func GetTypeByID(internalID id.InternalType) *DoltgresType {
	t, ok := InternalToBuiltInDoltgresType[internalID]
	if !ok {
		// TODO: return UNKNOWN?
		return nil
	}
	return t
}

// GetAllBuitInTypes returns a slice containing all registered types.
// The slice is sorted by each type's ID.
func GetAllBuitInTypes() []*DoltgresType {
	pgTypes := make([]*DoltgresType, 0, len(InternalToBuiltInDoltgresType))
	for internalID, typ := range InternalToBuiltInDoltgresType {
		if typ.ID == Unknown.ID && internalID.TypeName() != "unknown" {
			continue
		}
		pgTypes = append(pgTypes, typ)
	}
	sort.Slice(pgTypes, func(i, j int) bool {
		return pgTypes[i].ID < pgTypes[j].ID
	})
	return pgTypes
}

// InternalToBuiltInDoltgresType is a map of id.Internal to Doltgres' built-in type.
var InternalToBuiltInDoltgresType = map[id.InternalType]*DoltgresType{
	toInternal("_abstime"):         Unknown,
	toInternal("_aclitem"):         Unknown,
	toInternal("_bit"):             Unknown,
	toInternal("_bool"):            BoolArray,
	toInternal("_box"):             Unknown,
	toInternal("_bpchar"):          BpCharArray,
	toInternal("_bytea"):           ByteaArray,
	toInternal("_char"):            InternalCharArray,
	toInternal("_cid"):             Unknown,
	toInternal("_cidr"):            Unknown,
	toInternal("_circle"):          Unknown,
	toInternal("_cstring"):         CstringArray,
	toInternal("_date"):            DateArray,
	toInternal("_daterange"):       Unknown,
	toInternal("_float4"):          Float32Array,
	toInternal("_float8"):          Float64Array,
	toInternal("_gtsvector"):       Unknown,
	toInternal("_inet"):            Unknown,
	toInternal("_int2"):            Int16Array,
	toInternal("_int2vector"):      Unknown,
	toInternal("_int4"):            Int32Array,
	toInternal("_int4range"):       Unknown,
	toInternal("_int8"):            Int64Array,
	toInternal("_int8range"):       Unknown,
	toInternal("_interval"):        IntervalArray,
	toInternal("_json"):            JsonArray,
	toInternal("_jsonb"):           JsonBArray,
	toInternal("_line"):            Unknown,
	toInternal("_lseg"):            Unknown,
	toInternal("_macaddr"):         Unknown,
	toInternal("_money"):           Unknown,
	toInternal("_name"):            NameArray,
	toInternal("_numeric"):         NumericArray,
	toInternal("_numrange"):        Unknown,
	toInternal("_oid"):             OidArray,
	toInternal("_oidvector"):       Unknown,
	toInternal("_path"):            Unknown,
	toInternal("_pg_lsn"):          Unknown,
	toInternal("_point"):           Unknown,
	toInternal("_polygon"):         Unknown,
	toInternal("_record"):          RecordArray,
	toInternal("_refcursor"):       Unknown,
	toInternal("_regclass"):        RegclassArray,
	toInternal("_regconfig"):       Unknown,
	toInternal("_regdictionary"):   Unknown,
	toInternal("_regnamespace"):    Unknown,
	toInternal("_regoper"):         Unknown,
	toInternal("_regoperator"):     Unknown,
	toInternal("_regproc"):         RegprocArray,
	toInternal("_regprocedure"):    Unknown,
	toInternal("_regrole"):         Unknown,
	toInternal("_regtype"):         RegtypeArray,
	toInternal("_reltime"):         Unknown,
	toInternal("_text"):            TextArray,
	toInternal("_tid"):             Unknown,
	toInternal("_time"):            TimeArray,
	toInternal("_timestamp"):       TimestampArray,
	toInternal("_timestamptz"):     TimestampTZArray,
	toInternal("_timetz"):          TimeTZArray,
	toInternal("_tinterval"):       Unknown,
	toInternal("_tsquery"):         Unknown,
	toInternal("_tsrange"):         Unknown,
	toInternal("_tstzrange"):       Unknown,
	toInternal("_tsvector"):        Unknown,
	toInternal("_txid_snapshot"):   Unknown,
	toInternal("_uuid"):            UuidArray,
	toInternal("_varbit"):          Unknown,
	toInternal("_varchar"):         VarCharArray,
	toInternal("_xid"):             XidArray,
	toInternal("_xml"):             Unknown,
	toInternal("abstime"):          Unknown,
	toInternal("aclitem"):          Unknown,
	toInternal("any"):              Any,
	toInternal("anyarray"):         AnyArray,
	toInternal("anyelement"):       AnyElement,
	toInternal("anyenum"):          AnyEnum,
	toInternal("anynonarray"):      AnyNonArray,
	toInternal("anyrange"):         Unknown,
	toInternal("bit"):              Unknown,
	toInternal("bool"):             Bool,
	toInternal("box"):              Unknown,
	toInternal("bpchar"):           BpChar,
	toInternal("bytea"):            Bytea,
	toInternal("char"):             InternalChar,
	toInternal("cid"):              Unknown,
	toInternal("cidr"):             Unknown,
	toInternal("circle"):           Unknown,
	toInternal("cstring"):          Cstring,
	toInternal("date"):             Date,
	toInternal("daterange"):        Unknown,
	toInternal("event_trigger"):    Unknown,
	toInternal("fdw_handler"):      Unknown,
	toInternal("float4"):           Float32,
	toInternal("float8"):           Float64,
	toInternal("gtsvector"):        Unknown,
	toInternal("index_am_handler"): Unknown,
	toInternal("inet"):             Unknown,
	toInternal("int2"):             Int16,
	toInternal("int2vector"):       Unknown,
	toInternal("int4"):             Int32,
	toInternal("int4range"):        Unknown,
	toInternal("int8"):             Int64,
	toInternal("int8range"):        Unknown,
	toInternal("internal"):         Internal,
	toInternal("interval"):         Interval,
	toInternal("json"):             Json,
	toInternal("jsonb"):            JsonB,
	toInternal("language_handler"): Unknown,
	toInternal("line"):             Unknown,
	toInternal("lseg"):             Unknown,
	toInternal("macaddr"):          Unknown,
	toInternal("money"):            Unknown,
	toInternal("name"):             Name,
	toInternal("numeric"):          Numeric,
	toInternal("numrange"):         Unknown,
	toInternal("oid"):              Oid,
	toInternal("oidvector"):        Unknown,
	toInternal("opaque"):           Unknown,
	toInternal("path"):             Unknown,
	toInternal("pg_attribute"):     Unknown,
	toInternal("pg_auth_members"):  Unknown,
	toInternal("pg_authid"):        Unknown,
	toInternal("pg_class"):         Unknown,
	toInternal("pg_database"):      Unknown,
	toInternal("pg_ddl_command"):   Unknown,
	toInternal("pg_lsn"):           Unknown,
	toInternal("pg_node_tree"):     Unknown,
	toInternal("pg_proc"):          Unknown,
	toInternal("pg_shseclabel"):    Unknown,
	toInternal("pg_type"):          Unknown,
	toInternal("point"):            Unknown,
	toInternal("polygon"):          Unknown,
	toInternal("record"):           Record,
	toInternal("refcursor"):        Unknown,
	toInternal("regclass"):         Regclass,
	toInternal("regconfig"):        Unknown,
	toInternal("regdictionary"):    Unknown,
	toInternal("regnamespace"):     Unknown,
	toInternal("regoper"):          Unknown,
	toInternal("regoperator"):      Unknown,
	toInternal("regproc"):          Regproc,
	toInternal("regprocedure"):     Unknown,
	toInternal("regrole"):          Unknown,
	toInternal("regtype"):          Regtype,
	toInternal("reltime"):          Unknown,
	toInternal("smgr"):             Unknown,
	toInternal("text"):             Text,
	toInternal("tid"):              Unknown,
	toInternal("time"):             Time,
	toInternal("timestamp"):        Timestamp,
	toInternal("timestamptz"):      TimestampTZ,
	toInternal("timetz"):           TimeTZ,
	toInternal("tinterval"):        Unknown,
	toInternal("trigger"):          Unknown,
	toInternal("tsm_handler"):      Unknown,
	toInternal("tsquery"):          Unknown,
	toInternal("tsrange"):          Unknown,
	toInternal("tstzrange"):        Unknown,
	toInternal("tsvector"):         Unknown,
	toInternal("txid_snapshot"):    Unknown,
	toInternal("unknown"):          Unknown,
	toInternal("uuid"):             Uuid,
	toInternal("varbit"):           Unknown,
	toInternal("varchar"):          VarChar,
	toInternal("void"):             Void,
	toInternal("xid"):              Xid,
	toInternal("xml"):              Unknown,
}

// NameToInternalID is a mapping from a given name to its respective Internal ID.
var NameToInternalID = map[string]id.InternalType{}

// init, for now, fills the contents of NameToInternalID, so that we may search for types using regtype. This should be
// replaced with a better abstraction at some point.
func init() {
	for _, t := range GetAllBuitInTypes() {
		NameToInternalID[t.Name()] = t.ID
		pt, ok := types.OidToType[oid.Oid(id.Cache().ToOID(t.ID.Internal()))]
		if ok {
			NameToInternalID[pt.SQLStandardName()] = t.ID
		}
	}
}
