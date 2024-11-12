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

// typesFromOID contains a map from a OID to its originating type.
var typesFromOID = map[uint32]DoltgresType{
	AnyArray.OID:          AnyArray,
	AnyElement.OID:        AnyElement,
	AnyNonArray.OID:       AnyNonArray,
	BpChar.OID:            BpChar,
	BpCharArray.OID:       BpCharArray,
	Bool.OID:              Bool,
	BoolArray.OID:         BoolArray,
	Bytea.OID:             Bytea,
	ByteaArray.OID:        ByteaArray,
	Date.OID:              Date,
	DateArray.OID:         DateArray,
	Float32.OID:           Float32,
	Float32Array.OID:      Float32Array,
	Float64.OID:           Float64,
	Float64Array.OID:      Float64Array,
	Int16.OID:             Int16,
	Int16Array.OID:        Int16Array,
	Int32.OID:             Int32,
	Int32Array.OID:        Int32Array,
	Int64.OID:             Int64,
	Int64Array.OID:        Int64Array,
	InternalChar.OID:      InternalChar,
	InternalCharArray.OID: InternalCharArray,
	Interval.OID:          Interval,
	IntervalArray.OID:     IntervalArray,
	Json.OID:              Json,
	JsonArray.OID:         JsonArray,
	JsonB.OID:             JsonB,
	JsonBArray.OID:        JsonBArray,
	Name.OID:              Name,
	NameArray.OID:         NameArray,
	Numeric.OID:           Numeric,
	NumericArray.OID:      NumericArray,
	Oid.OID:               Oid,
	OidArray.OID:          OidArray,
	Regclass.OID:          Regclass,
	RegclassArray.OID:     RegclassArray,
	Regproc.OID:           Regproc,
	RegprocArray.OID:      RegprocArray,
	Regtype.OID:           Regtype,
	RegtypeArray.OID:      RegtypeArray,
	Text.OID:              Text,
	TextArray.OID:         TextArray,
	Time.OID:              Time,
	TimeArray.OID:         TimeArray,
	Timestamp.OID:         Timestamp,
	TimestampArray.OID:    TimestampArray,
	TimestampTZ.OID:       TimestampTZ,
	TimestampTZArray.OID:  TimestampTZArray,
	TimeTZ.OID:            TimeTZ,
	TimeTZArray.OID:       TimeTZArray,
	Uuid.OID:              Uuid,
	UuidArray.OID:         UuidArray,
	Unknown.OID:           Unknown,
	VarChar.OID:           VarChar,
	VarCharArray.OID:      VarCharArray,
	Xid.OID:               Xid,
	XidArray.OID:          XidArray,
}

// GetTypeByOID returns the DoltgresType matching the given OID. If the OID does not match a type, then nil is returned.
func GetTypeByOID(oid uint32) DoltgresType {
	t, ok := typesFromOID[oid]
	if !ok {
		return DoltgresType{}
	}
	return t
}

// GetAllTypes returns a slice containing all registered types. The slice is sorted by each type's base ID.
func GetAllTypes() []DoltgresType {
	pgTypes := make([]DoltgresType, 0, len(typesFromOID))
	for _, typ := range typesFromOID {
		pgTypes = append(pgTypes, typ)
	}
	sort.Slice(pgTypes, func(i, j int) bool {
		return pgTypes[i].OID < pgTypes[j].OID
	})
	return pgTypes
}

// OidToBuildInDoltgresType is a map of oid to built-in Doltgres type.
var OidToBuildInDoltgresType = map[uint32]DoltgresType{
	uint32(oid.T_bool):             Bool,
	uint32(oid.T_bytea):            Bytea,
	uint32(oid.T_char):             InternalChar,
	uint32(oid.T_name):             Name,
	uint32(oid.T_int8):             Int64,
	uint32(oid.T_int2):             Int16,
	uint32(oid.T_int2vector):       Unknown,
	uint32(oid.T_int4):             Int32,
	uint32(oid.T_regproc):          Regproc,
	uint32(oid.T_text):             Text,
	uint32(oid.T_oid):              Oid,
	uint32(oid.T_tid):              Unknown,
	uint32(oid.T_xid):              Xid,
	uint32(oid.T_cid):              Unknown,
	uint32(oid.T_oidvector):        Unknown,
	uint32(oid.T_pg_ddl_command):   Unknown,
	uint32(oid.T_pg_type):          Unknown,
	uint32(oid.T_pg_attribute):     Unknown,
	uint32(oid.T_pg_proc):          Unknown,
	uint32(oid.T_pg_class):         Unknown,
	uint32(oid.T_json):             Json,
	uint32(oid.T_xml):              Unknown,
	uint32(oid.T__xml):             Unknown,
	uint32(oid.T_pg_node_tree):     Unknown,
	uint32(oid.T__json):            JsonArray,
	uint32(oid.T_smgr):             Unknown,
	uint32(oid.T_index_am_handler): Unknown,
	uint32(oid.T_point):            Unknown,
	uint32(oid.T_lseg):             Unknown,
	uint32(oid.T_path):             Unknown,
	uint32(oid.T_box):              Unknown,
	uint32(oid.T_polygon):          Unknown,
	uint32(oid.T_line):             Unknown,
	uint32(oid.T__line):            Unknown,
	uint32(oid.T_cidr):             Unknown,
	uint32(oid.T__cidr):            Unknown,
	uint32(oid.T_float4):           Float32,
	uint32(oid.T_float8):           Float64,
	uint32(oid.T_abstime):          Unknown,
	uint32(oid.T_reltime):          Unknown,
	uint32(oid.T_tinterval):        Unknown,
	uint32(oid.T_unknown):          Unknown,
	uint32(oid.T_circle):           Unknown,
	uint32(oid.T__circle):          Unknown,
	uint32(oid.T_money):            Unknown,
	uint32(oid.T__money):           Unknown,
	uint32(oid.T_macaddr):          Unknown,
	uint32(oid.T_inet):             Unknown,
	uint32(oid.T__bool):            BoolArray,
	uint32(oid.T__bytea):           ByteaArray,
	uint32(oid.T__char):            InternalCharArray,
	uint32(oid.T__name):            NameArray,
	uint32(oid.T__int2):            Int16Array,
	uint32(oid.T__int2vector):      Unknown,
	uint32(oid.T__int4):            Int32Array,
	uint32(oid.T__regproc):         RegprocArray,
	uint32(oid.T__text):            TextArray,
	uint32(oid.T__tid):             Unknown,
	uint32(oid.T__xid):             XidArray,
	uint32(oid.T__cid):             Unknown,
	uint32(oid.T__oidvector):       Unknown,
	uint32(oid.T__bpchar):          BpCharArray,
	uint32(oid.T__varchar):         VarCharArray,
	uint32(oid.T__int8):            Int64Array,
	uint32(oid.T__point):           Unknown,
	uint32(oid.T__lseg):            Unknown,
	uint32(oid.T__path):            Unknown,
	uint32(oid.T__box):             Unknown,
	uint32(oid.T__float4):          Float32Array,
	uint32(oid.T__float8):          Float64Array,
	uint32(oid.T__abstime):         Unknown,
	uint32(oid.T__reltime):         Unknown,
	uint32(oid.T__tinterval):       Unknown,
	uint32(oid.T__polygon):         Unknown,
	uint32(oid.T__oid):             OidArray,
	uint32(oid.T_aclitem):          Unknown,
	uint32(oid.T__aclitem):         Unknown,
	uint32(oid.T__macaddr):         Unknown,
	uint32(oid.T__inet):            Unknown,
	uint32(oid.T_bpchar):           BpChar,
	uint32(oid.T_varchar):          VarChar,
	uint32(oid.T_date):             Date,
	uint32(oid.T_time):             Time,
	uint32(oid.T_timestamp):        Timestamp,
	uint32(oid.T__timestamp):       TimestampArray,
	uint32(oid.T__date):            DateArray,
	uint32(oid.T__time):            TimeArray,
	uint32(oid.T_timestamptz):      TimestampTZ,
	uint32(oid.T__timestamptz):     TimestampTZArray,
	uint32(oid.T_interval):         Interval,
	uint32(oid.T__interval):        IntervalArray,
	uint32(oid.T__numeric):         NumericArray,
	uint32(oid.T_pg_database):      Unknown,
	uint32(oid.T__cstring):         Unknown,
	uint32(oid.T_timetz):           TimeTZ,
	uint32(oid.T__timetz):          TimeTZArray,
	uint32(oid.T_bit):              Unknown,
	uint32(oid.T__bit):             Unknown,
	uint32(oid.T_varbit):           Unknown,
	uint32(oid.T__varbit):          Unknown,
	uint32(oid.T_numeric):          Numeric,
	uint32(oid.T_refcursor):        Unknown,
	uint32(oid.T__refcursor):       Unknown,
	uint32(oid.T_regprocedure):     Unknown,
	uint32(oid.T_regoper):          Unknown,
	uint32(oid.T_regoperator):      Unknown,
	uint32(oid.T_regclass):         Regclass,
	uint32(oid.T_regtype):          Regtype,
	uint32(oid.T__regprocedure):    Unknown,
	uint32(oid.T__regoper):         Unknown,
	uint32(oid.T__regoperator):     Unknown,
	uint32(oid.T__regclass):        RegclassArray,
	uint32(oid.T__regtype):         RegtypeArray,
	uint32(oid.T_record):           Unknown,
	uint32(oid.T_cstring):          Unknown,
	uint32(oid.T_any):              Unknown,
	uint32(oid.T_anyarray):         AnyArray,
	uint32(oid.T_void):             Unknown,
	uint32(oid.T_trigger):          Unknown,
	uint32(oid.T_language_handler): Unknown,
	uint32(oid.T_internal):         Unknown,
	uint32(oid.T_opaque):           Unknown,
	uint32(oid.T_anyelement):       AnyElement,
	uint32(oid.T__record):          Unknown,
	uint32(oid.T_anynonarray):      AnyNonArray,
	uint32(oid.T_pg_authid):        Unknown,
	uint32(oid.T_pg_auth_members):  Unknown,
	uint32(oid.T__txid_snapshot):   Unknown,
	uint32(oid.T_uuid):             Uuid,
	uint32(oid.T__uuid):            UuidArray,
	uint32(oid.T_txid_snapshot):    Unknown,
	uint32(oid.T_fdw_handler):      Unknown,
	uint32(oid.T_pg_lsn):           Unknown,
	uint32(oid.T__pg_lsn):          Unknown,
	uint32(oid.T_tsm_handler):      Unknown,
	uint32(oid.T_anyenum):          Unknown,
	uint32(oid.T_tsvector):         Unknown,
	uint32(oid.T_tsquery):          Unknown,
	uint32(oid.T_gtsvector):        Unknown,
	uint32(oid.T__tsvector):        Unknown,
	uint32(oid.T__gtsvector):       Unknown,
	uint32(oid.T__tsquery):         Unknown,
	uint32(oid.T_regconfig):        Unknown,
	uint32(oid.T__regconfig):       Unknown,
	uint32(oid.T_regdictionary):    Unknown,
	uint32(oid.T__regdictionary):   Unknown,
	uint32(oid.T_jsonb):            JsonB,
	uint32(oid.T__jsonb):           JsonBArray,
	uint32(oid.T_anyrange):         Unknown,
	uint32(oid.T_event_trigger):    Unknown,
	uint32(oid.T_int4range):        Unknown,
	uint32(oid.T__int4range):       Unknown,
	uint32(oid.T_numrange):         Unknown,
	uint32(oid.T__numrange):        Unknown,
	uint32(oid.T_tsrange):          Unknown,
	uint32(oid.T__tsrange):         Unknown,
	uint32(oid.T_tstzrange):        Unknown,
	uint32(oid.T__tstzrange):       Unknown,
	uint32(oid.T_daterange):        Unknown,
	uint32(oid.T__daterange):       Unknown,
	uint32(oid.T_int8range):        Unknown,
	uint32(oid.T__int8range):       Unknown,
	uint32(oid.T_pg_shseclabel):    Unknown,
	uint32(oid.T_regnamespace):     Unknown,
	uint32(oid.T__regnamespace):    Unknown,
	uint32(oid.T_regrole):          Unknown,
	uint32(oid.T__regrole):         Unknown,
}
