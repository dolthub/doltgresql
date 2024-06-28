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

import "github.com/dolthub/go-mysql-server/sql/types"

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
	IoInput(input string) (any, error)
	// IoOutput returns a string from the given output value. This function mirrors Postgres' I/O output function. These
	// strings are not intended for output, but are instead intended for serialization and cross-type conversion. Output
	// values will always be non-NULL.
	IoOutput(output any) (string, error)
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

// typesFromBaseID contains a map from a DoltgresTypeBaseID to its originating type.
var typesFromBaseID = map[DoltgresTypeBaseID]DoltgresType{
	AnyArray.BaseID():         AnyArray,
	BpChar.BaseID():           BpChar,
	BpCharArray.BaseID():      BpCharArray,
	Bool.BaseID():             Bool,
	BoolArray.BaseID():        BoolArray,
	Bytea.BaseID():            Bytea,
	ByteaArray.BaseID():       ByteaArray,
	Date.BaseID():             Date,
	DateArray.BaseID():        DateArray,
	Float32.BaseID():          Float32,
	Float32Array.BaseID():     Float32Array,
	Float64.BaseID():          Float64,
	Float64Array.BaseID():     Float64Array,
	Int16.BaseID():            Int16,
	Int16Array.BaseID():       Int16Array,
	Int16Serial.BaseID():      Int16Serial,
	Int32.BaseID():            Int32,
	Int32Array.BaseID():       Int32Array,
	Int32Serial.BaseID():      Int32Serial,
	Int64.BaseID():            Int64,
	Int64Array.BaseID():       Int64Array,
	Int64Serial.BaseID():      Int64Serial,
	Json.BaseID():             Json,
	JsonArray.BaseID():        JsonArray,
	JsonB.BaseID():            JsonB,
	JsonBArray.BaseID():       JsonBArray,
	Name.BaseID():             Name,
	NameArray.BaseID():        NameArray,
	Null.BaseID():             Null,
	Numeric.BaseID():          Numeric,
	NumericArray.BaseID():     NumericArray,
	Oid.BaseID():              Oid,
	OidArray.BaseID():         OidArray,
	Text.BaseID():             Text,
	TextArray.BaseID():        TextArray,
	Time.BaseID():             Time,
	TimeArray.BaseID():        TimeArray,
	Timestamp.BaseID():        Timestamp,
	TimestampArray.BaseID():   TimestampArray,
	TimestampTZ.BaseID():      TimestampTZ,
	TimestampTZArray.BaseID(): TimestampTZArray,
	TimeTZ.BaseID():           TimeTZ,
	TimeTZArray.BaseID():      TimeTZArray,
	Uuid.BaseID():             Uuid,
	UuidArray.BaseID():        UuidArray,
	Unknown.BaseID():          Unknown,
	VarChar.BaseID():          VarChar,
	VarCharArray.BaseID():     VarCharArray,
	Xid.BaseID():              Xid,
	XidArray.BaseID():         XidArray,
}

// GetPgTypes returns an array of DoltgresTypes with types that should be displayed in `pg_catalog.pg_type` table.
// It filters out the array types, null type and serial types as they are not present in the pg_type table.
// It also adds the `Internal "char"` type as it is present in the pg_table table.
func GetPgTypes() []DoltgresType {
	pgTypes := make([]DoltgresType, 0, len(typesFromBaseID))
	for _, typ := range typesFromBaseID {
		switch typ.(type) {
		case DoltgresArrayType, NullType, Int16TypeSerial,
			Int32TypeSerial, Int64TypeSerial:
			// these types are not present in pg_type table
			continue
		default:
			pgTypes = append(pgTypes, typ)
		}
	}
	// internal "char" type is present in pg_type table.
	pgTypes = append(pgTypes, InternalChar)
	return pgTypes
}
