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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/lib/pq/oid"
	"gopkg.in/src-d/go-errors.v1"
)

var ErrTypeAlreadyExists = errors.NewKind(`type "%s" already exists`)
var ErrTypeDoesNotExist = errors.NewKind(`type "%s" does not exist`)

// Type represents a single type.
type Type struct {
	Name          string
	Owner         string // TODO: should be `uint32`.
	Length        int16
	PassedByVal   bool
	Typ           TypeType
	Category      TypeCategory
	IsPreferred   bool
	IsDefined     bool
	Delimiter     string
	RelID         uint32 // for Composite types
	SubscriptFunc string
	Elem          uint32
	Array         uint32
	InputFunc     string
	OutputFunc    string
	ReceiveFunc   string
	SendFunc      string
	ModInFunc     string
	ModOutFunc    string
	AnalyzeFunc   string
	Align         TypeAlignment
	Storage       TypeStorage
	NotNull       bool   // for Domain types
	BaseTypeOID   uint32 // for Domain types
	TypMod        int32  // for Domain types
	NDims         int32  // for Domain types
	Collation     uint32
	DefaulBin     string // for Domain types
	Default       string
	Acl           string                 // TODO: list of privileges
	Checks        []*sql.CheckDefinition // TODO: this is not part of `pg_type` instead `pg_constraint` for Domain types.
}

// DoltgresType is a type that is distinct from the MySQL types in GMS.
type DoltgresType interface {
	types.ExtendedType
	// Alignment returns a char representing the alignment required when storing a value of this type.
	Alignment() TypeAlignment
	// BaseID returns the DoltgresTypeBaseID for this type.
	BaseID() DoltgresTypeBaseID
	// BaseName returns the name of the type displayed in pg_catalog tables.
	BaseName() string
	// Category returns a char representing an arbitrary classification of data types that is used by the parser to determine which implicit casts should be “preferred”.
	Category() TypeCategory
	// GetSerializationID returns the SerializationID for this type.
	GetSerializationID() SerializationID
	// IoInput returns a value from the given input string. This function mirrors Postgres' I/O input function. Such
	// strings are intended for serialization and automatic cross-type conversion. An input string will never represent
	// NULL.
	IoInput(ctx *sql.Context, input string) (any, error)
	// IoOutput returns a string from the given output value. This function mirrors Postgres' I/O output function. These
	// strings are not intended for output, but are instead intended for serialization and cross-type conversion. Output
	// values will always be non-NULL.
	IoOutput(ctx *sql.Context, output any) (string, error)
	// IsPreferredType returns true if the type is preferred type.
	IsPreferredType() bool
	// IsUnbounded returns whether the type is unbounded. Unbounded types do not enforce a length, precision, etc. on
	// values. All values are still bound by the field size limit, but that differs from any type-enforced limits.
	IsUnbounded() bool
	// OID returns an OID that we are associating with this type. OIDs are not unique, and are not guaranteed to be the
	// same between versions of Postgres. However, they've so far appeared relatively stable, and many libraries rely on
	// them for type identification, so we return them here. These should not be used for any sort of identification on
	// our side. For that, we should use DoltgresTypeBaseID, which we can guarantee will be unique and non-changing once
	// we've stabilized development.
	OID() uint32
	// SerializeType returns a byte slice representing the serialized form of the type. All serialized types MUST start
	// with their SerializationID. Deserialization is done through the DeserializeType function.
	SerializeType() ([]byte, error)
	// deserializeType returns a new type based on the given version and metadata. The metadata is all data after the
	// serialization header. This is called from within the types package. To deserialize types normally, use
	// DeserializeType, which will call this as needed.
	deserializeType(version uint16, metadata []byte) (DoltgresType, error)
	// ToArrayType converts the calling DoltgresType into its corresponding array type. When called on a
	// DoltgresArrayType, then it simply returns itself, as a multidimensional or nested array is equivalent to a
	// standard array.
	ToArrayType() DoltgresArrayType
}

// DoltgresArrayType is a DoltgresType that represents an array variant of a non-array type.
type DoltgresArrayType interface {
	DoltgresType
	// BaseType is the inner type of the array. This will always be a non-array type.
	BaseType() DoltgresType
}

// DoltgresPolymorphicType is a DoltgresType that represents one of the polymorphic types. These types are special
// built-in pseudo-types that are used during function resolution to allow a function to handle multiple types from a
// single definition. All polymorphic types have "any" as a prefix. The exception is the "any" type, which is not a
// polymorphic type.
type DoltgresPolymorphicType interface {
	DoltgresType
	// IsValid returns whether the given type is valid for the calling polymorphic type.
	IsValid(target DoltgresType) bool
}

// typesFromBaseID contains a map from a DoltgresTypeBaseID to its originating type.
var typesFromBaseID = map[DoltgresTypeBaseID]DoltgresType{
	AnyArray.BaseID():          AnyArray,
	AnyElement.BaseID():        AnyElement,
	AnyNonArray.BaseID():       AnyNonArray,
	BpChar.BaseID():            BpChar,
	BpCharArray.BaseID():       BpCharArray,
	Bool.BaseID():              Bool,
	BoolArray.BaseID():         BoolArray,
	Bytea.BaseID():             Bytea,
	ByteaArray.BaseID():        ByteaArray,
	Date.BaseID():              Date,
	DateArray.BaseID():         DateArray,
	Float32.BaseID():           Float32,
	Float32Array.BaseID():      Float32Array,
	Float64.BaseID():           Float64,
	Float64Array.BaseID():      Float64Array,
	Int16.BaseID():             Int16,
	Int16Array.BaseID():        Int16Array,
	Int16Serial.BaseID():       Int16Serial,
	Int32.BaseID():             Int32,
	Int32Array.BaseID():        Int32Array,
	Int32Serial.BaseID():       Int32Serial,
	Int64.BaseID():             Int64,
	Int64Array.BaseID():        Int64Array,
	Int64Serial.BaseID():       Int64Serial,
	InternalChar.BaseID():      InternalChar,
	InternalCharArray.BaseID(): InternalCharArray,
	Interval.BaseID():          Interval,
	IntervalArray.BaseID():     IntervalArray,
	Json.BaseID():              Json,
	JsonArray.BaseID():         JsonArray,
	JsonB.BaseID():             JsonB,
	JsonBArray.BaseID():        JsonBArray,
	Name.BaseID():              Name,
	NameArray.BaseID():         NameArray,
	Numeric.BaseID():           Numeric,
	NumericArray.BaseID():      NumericArray,
	Oid.BaseID():               Oid,
	OidArray.BaseID():          OidArray,
	Regclass.BaseID():          Regclass,
	RegclassArray.BaseID():     RegclassArray,
	Regproc.BaseID():           Regproc,
	RegprocArray.BaseID():      RegprocArray,
	Regtype.BaseID():           Regtype,
	RegtypeArray.BaseID():      RegtypeArray,
	Text.BaseID():              Text,
	TextArray.BaseID():         TextArray,
	Time.BaseID():              Time,
	TimeArray.BaseID():         TimeArray,
	Timestamp.BaseID():         Timestamp,
	TimestampArray.BaseID():    TimestampArray,
	TimestampTZ.BaseID():       TimestampTZ,
	TimestampTZArray.BaseID():  TimestampTZArray,
	TimeTZ.BaseID():            TimeTZ,
	TimeTZArray.BaseID():       TimeTZArray,
	Uuid.BaseID():              Uuid,
	UuidArray.BaseID():         UuidArray,
	Unknown.BaseID():           Unknown,
	VarChar.BaseID():           VarChar,
	VarCharArray.BaseID():      VarCharArray,
	Xid.BaseID():               Xid,
	XidArray.BaseID():          XidArray,
}

// GetAllTypes returns a slice containing all registered types. The slice is sorted by each type's base ID.
func GetAllTypes() []DoltgresType {
	pgTypes := make([]DoltgresType, 0, len(typesFromBaseID))
	for _, typ := range typesFromBaseID {
		pgTypes = append(pgTypes, typ)
	}
	sort.Slice(pgTypes, func(i, j int) bool {
		return pgTypes[i].BaseID() < pgTypes[j].BaseID()
	})
	return pgTypes
}

// OidToBuildInDoltgresType is map of oid to built-in Doltgres type.
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
