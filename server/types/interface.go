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

// DoltgresTypeBaseID is an ID that is common between all variations of a DoltgresType. For example, VARCHAR(3) and
// VARCHAR(6) are different types, however they will return the same DoltgresTypeBaseID. This ID is not suitable for
// serialization, as it may change over time.
type DoltgresTypeBaseID uint32

// DoltgresType is a type that is distinct from the MySQL types in GMS.
type DoltgresType interface {
	types.ExtendedType
	// BaseID returns the DoltgresTypeBaseID for this type.
	BaseID() DoltgresTypeBaseID
	// OID returns an OID that we are associating with this type. OIDs are not unique, and are not guaranteed to be the
	// same between versions of Postgres. However, they've so far appeared relatively stable, and many libraries rely on
	// them for type identification, so we return them here. These should not be used for any sort of identification on
	// our side. For that, we should use DoltgresTypeBaseID, which we can guarantee will be unique and non-changing once
	// we've stabilized development.
	OID() uint32
}

// FromBaseID returns a DoltgresType that matches the Base ID. This type will usually be the most permissive version of
// the type, along with any default values.
func FromBaseID(baseID DoltgresTypeBaseID) DoltgresType {
	t, ok := typesFromBaseID[baseID]
	if !ok {
		panic("unknown Doltgres base id")
	}
	return t
}

// typesFromBaseID contains a map from a DoltgresTypeBaseID to its originating type.
var typesFromBaseID = map[DoltgresTypeBaseID]DoltgresType{
	Bool.BaseID():       Bool,
	BoolArray.BaseID():  BoolArray,
	Float32.BaseID():    Float32,
	Float64.BaseID():    Float64,
	Int16.BaseID():      Int16,
	Int32.BaseID():      Int32,
	Int64.BaseID():      Int64,
	Null.BaseID():       Null,
	Numeric.BaseID():    Numeric,
	Uuid.BaseID():       Uuid,
	VarCharMax.BaseID(): VarCharMax,
}
