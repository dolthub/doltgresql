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
	"fmt"

	"github.com/dolthub/dolt/go/libraries/doltcore/schema/typeinfo"
	doltTypes "github.com/dolthub/dolt/go/store/types"
	"github.com/dolthub/dolt/go/store/val"
	"github.com/dolthub/go-mysql-server/sql"
)

// typeInfo is the implementation of typeinfo.TypeInfo for DoltgresType.
type typeInfo struct {
	Type     *DoltgresType
	encoding val.Encoding
}

var _ typeinfo.TypeInfo = (*typeInfo)(nil)

// Equals implements typeinfo.TypeInfo.
func (t typeInfo) Equals(other typeinfo.TypeInfo) bool {
	ot, ok := other.(typeInfo)
	return ok && t.Type.Equals(ot.Type) && t.encoding == ot.encoding
}

// NomsKind implements typeinfo.TypeInfo.
func (t typeInfo) NomsKind() doltTypes.NomsKind {
	// This kind is only ever used when determining column tags, so we won't worry about encoding here.
	return doltTypes.ExtendedKind
}

// ToSqlType implements typeinfo.TypeInfo.
func (t typeInfo) ToSqlType() sql.Type {
	return t.Type
}

// Encoding implements typeinfo.TypeInfo.
func (t typeInfo) Encoding() val.Encoding {
	if t.encoding > 0 {
		return t.encoding
	}

	switch t.Type.ID.TypeName() {
	case "int2":
		return val.Int16Enc
	case "int4":
		return val.Int32Enc
	case "int8":
		return val.Int64Enc
	case "float4":
		return val.Float32Enc
	case "float8":
		return val.Float64Enc
	case "numeric", "decimal":
		return val.DecimalEnc
	case "bytea":
		return val.BytesAdaptiveEnc
	// TODO: use dolt JSON document encoding here
	// case "json", "jsonb":
	// 	return val.JSONAddrEnc
	case "xid":
		return val.Uint32Enc
		// TODO: uuid is represented as a uuid.Uuid in doltgres, but dolt wants []byte for BytesAdaptiveEnc
	// case "uuid":
	// 	return val.BytesAdaptiveEnc
	case "varchar":
		if t.Type.attTypMod == -1 {
			return val.StringAdaptiveEnc
		}
		return val.StringEnc
	case "name", "char":
		return val.StringEnc
	case "bpchar", "text":
		return val.StringAdaptiveEnc
	default:
		switch t.Type.MaxSerializedWidth() {
		case sql.ExtendedTypeSerializedWidth_64K:
			return val.ExtendedEnc
		case sql.ExtendedTypeSerializedWidth_Unbounded:
			return val.ExtendedAdaptiveEnc
		default:
			panic(fmt.Errorf("unknown extended type serialization width"))
		}
	}
}

// WithEncoding implements typeinfo.TypeInfo.
func (t typeInfo) WithEncoding(enc val.Encoding) typeinfo.TypeInfo {
	return typeInfo{
		Type:     t.Type,
		encoding: enc,
	}
}

// String implements typeinfo.TypeInfo.
func (t typeInfo) String() string {
	return fmt.Sprintf("TypeInfo(%s, encoding=%d)", t.Type.String(), t.encoding)
}
