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
	// BaseID returns the DoltgresTypeBaseID for this type.
	BaseID() DoltgresTypeBaseID
	// GetSerializationID returns the SerializationID for this type.
	GetSerializationID() SerializationID
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
	Int32.BaseID():            Int32,
	Int32Array.BaseID():       Int32Array,
	Int64.BaseID():            Int64,
	Int64Array.BaseID():       Int64Array,
	Null.BaseID():             Null,
	Numeric.BaseID():          Numeric,
	NumericArray.BaseID():     NumericArray,
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
}
