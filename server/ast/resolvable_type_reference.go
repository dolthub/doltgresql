// Copyright 2023 Dolthub, Inc.
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

package ast

import (
	"fmt"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/postgres/parser/types"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// nodeResolvableTypeReference handles tree.ResolvableTypeReference nodes.
func nodeResolvableTypeReference(typ tree.ResolvableTypeReference) (*vitess.ConvertType, pgtypes.DoltgresType, error) {
	if typ == nil {
		return nil, nil, nil
	}

	var columnTypeName string
	var columnTypeLength *vitess.SQLVal
	var columnTypeScale *vitess.SQLVal
	var resolvedType pgtypes.DoltgresType
	switch columnType := typ.(type) {
	case *tree.ArrayTypeReference:
		return nil, nil, fmt.Errorf("the given array type is not yet supported")
	case *tree.OIDTypeReference:
		return nil, nil, fmt.Errorf("referencing types by their OID is not yet supported")
	case *tree.UnresolvedObjectName:
		return nil, nil, fmt.Errorf("type declaration format is not yet supported")
	case *types.GeoMetadata:
		return nil, nil, fmt.Errorf("geometry types are not yet supported")
	case *types.T:
		columnTypeName = columnType.SQLStandardName()
		if columnType.Family() == types.ArrayFamily {
			_, baseResolvedType, err := nodeResolvableTypeReference(columnType.ArrayContents())
			if err != nil {
				return nil, nil, err
			}
			resolvedType = baseResolvedType.ToArrayType()
		} else {
			switch columnType.Oid() {
			case oid.T_bool:
				resolvedType = pgtypes.Bool
			case oid.T_bytea:
				resolvedType = pgtypes.Bytea
			case oid.T_bpchar:
				width := uint32(columnType.Width())
				if width > pgtypes.StringMaxLength {
					return nil, nil, fmt.Errorf("length for type bpchar cannot exceed %d", pgtypes.StringMaxLength)
				}
				if width == 0 {
					resolvedType = pgtypes.BpChar
				} else {
					resolvedType = pgtypes.CharType{Length: width}
				}
			case oid.T_char:
				width := uint32(columnType.Width())
				if width > pgtypes.StringMaxLength {
					return nil, nil, fmt.Errorf("length for type char cannot exceed %d", pgtypes.StringMaxLength)
				}
				if width == 0 {
					width = 1
				}
				resolvedType = pgtypes.CharType{Length: width}
			case oid.T_date:
				resolvedType = pgtypes.Date
			case oid.T_float4:
				resolvedType = pgtypes.Float32
			case oid.T_float8:
				resolvedType = pgtypes.Float64
			case oid.T_int2:
				resolvedType = pgtypes.Int16
			case oid.T_int4:
				resolvedType = pgtypes.Int32
			case oid.T_int8:
				resolvedType = pgtypes.Int64
			case oid.T_json:
				columnTypeName = "JSON"
			case oid.T_jsonb:
				columnTypeName = "JSON"
			case oid.T_name:
				resolvedType = pgtypes.Name
			case oid.T_numeric:
				if columnType.Precision() == 0 && columnType.Scale() == 0 {
					resolvedType = pgtypes.Numeric
				} else {
					resolvedType = pgtypes.NumericType{
						Precision: columnType.Precision(),
						Scale:     columnType.Scale(),
					}
				}
			case oid.T_oid:
				resolvedType = pgtypes.Oid
			case oid.T_text:
				resolvedType = pgtypes.Text
			case oid.T_time:
				resolvedType = pgtypes.Time
			case oid.T_timestamp:
				resolvedType = pgtypes.Timestamp
			case oid.T_timestamptz:
				resolvedType = pgtypes.TimestampTZ
			case oid.T_timetz:
				resolvedType = pgtypes.TimeTZ
			case oid.T_uuid:
				resolvedType = pgtypes.Uuid
			case oid.T_varchar:
				width := uint32(columnType.Width())
				if width > pgtypes.StringMaxLength {
					return nil, nil, fmt.Errorf("length for type varchar cannot exceed %d", pgtypes.StringMaxLength)
				}
				resolvedType = pgtypes.VarCharType{Length: width}
			case oid.T_xid:
				resolvedType = pgtypes.Xid
			default:
				return nil, nil, fmt.Errorf("unknown type with oid: %d", uint32(columnType.Oid()))
			}
		}
	}

	return &vitess.ConvertType{
		Type:    columnTypeName,
		Length:  columnTypeLength,
		Scale:   columnTypeScale,
		Charset: "", // TODO
	}, resolvedType, nil
}
