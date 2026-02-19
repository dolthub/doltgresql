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
	"strings"

	"github.com/cockroachdb/errors"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/postgres/parser/types"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// nodeResolvableTypeReference handles tree.ResolvableTypeReference nodes.
func nodeResolvableTypeReference(ctx *Context, typ tree.ResolvableTypeReference, mayBeTrigger bool) (*vitess.ConvertType, *pgtypes.DoltgresType, error) {
	if typ == nil {
		// TODO: use UNKNOWN?
		return nil, nil, nil
	}

	var columnTypeName string
	var columnTypeLength *vitess.SQLVal
	var columnTypeScale *vitess.SQLVal
	var doltgresType *pgtypes.DoltgresType
	var err error
	switch columnType := typ.(type) {
	case *tree.ArrayTypeReference:
		if uon, ok := columnType.ElementType.(*tree.UnresolvedObjectName); ok {
			return nodeResolvableTypeReference(ctx, uon, mayBeTrigger)
		}
		return nil, nil, errors.Errorf("the given array type is not yet supported")
	case *tree.OIDTypeReference:
		return nil, nil, errors.Errorf("referencing types by their OID is not yet supported")
	case *tree.UnresolvedObjectName:
		tn := columnType.ToTableName()
		columnTypeName = tn.Object()
		doltgresType = pgtypes.NewUnresolvedDoltgresType(tn.Schema(), columnTypeName)
	case *types.GeoMetadata:
		return nil, nil, errors.Errorf("geometry types are not yet supported")
	case *types.T:
		columnTypeName = columnType.SQLStandardName()
		if columnType.Family() == types.ArrayFamily {
			_, baseResolvedType, err := nodeResolvableTypeReference(ctx, columnType.ArrayContents(), mayBeTrigger)
			if err != nil {
				return nil, nil, err
			}
			if baseResolvedType.IsResolvedType() {
				// currently the built-in types will be resolved, so it can retrieve its array type
				doltgresType = baseResolvedType.ToArrayType()
			} else {
				// TODO: handle array type of non-built-in types
				baseResolvedType.TypCategory = pgtypes.TypeCategory_ArrayTypes
				doltgresType = baseResolvedType
			}
		} else if columnType.Family() == types.GeometryFamily {
			return nil, nil, errors.Errorf("geometry types are not yet supported")
		} else if columnType.Family() == types.GeographyFamily {
			return nil, nil, errors.Errorf("geography types are not yet supported")
		} else {
			switch columnType.Oid() {
			case oid.T_record:
				doltgresType = pgtypes.Record
			case oid.T_bool:
				doltgresType = pgtypes.Bool
			case oid.T_bytea:
				doltgresType = pgtypes.Bytea
			case oid.T_bpchar:
				width := uint32(columnType.Width())
				if width > pgtypes.StringMaxLength {
					return nil, nil, errors.Errorf("length for type bpchar cannot exceed %d", pgtypes.StringMaxLength)
				} else if width == 0 {
					// TODO: need to differentiate between definitions 'bpchar' (valid) and 'char(0)' (invalid)
					doltgresType = pgtypes.BpChar
				} else {
					doltgresType, err = pgtypes.NewCharType(int32(width))
					if err != nil {
						return nil, nil, err
					}
				}
			case oid.T_char:
				width := uint32(columnType.Width())
				if width > pgtypes.InternalCharLength {
					return nil, nil, errors.Errorf("length for type \"char\" cannot exceed %d", pgtypes.InternalCharLength)
				}
				if width == 0 {
					width = 1
				}
				doltgresType = pgtypes.InternalChar
			case oid.T_date:
				doltgresType = pgtypes.Date
			case oid.T_float4:
				doltgresType = pgtypes.Float32
			case oid.T_float8:
				doltgresType = pgtypes.Float64
			case oid.T_int2:
				doltgresType = pgtypes.Int16
			case oid.T_int4:
				doltgresType = pgtypes.Int32
			case oid.T_int8:
				doltgresType = pgtypes.Int64
			case oid.T_interval:
				doltgresType = pgtypes.Interval
			case oid.T_json:
				doltgresType = pgtypes.Json
			case oid.T_jsonb:
				doltgresType = pgtypes.JsonB
			case oid.T_name:
				doltgresType = pgtypes.Name
			case oid.T_numeric:
				if columnType.Precision() == 0 && columnType.Scale() == 0 {
					doltgresType = pgtypes.Numeric
				} else {
					doltgresType, err = pgtypes.NewNumericTypeWithPrecisionAndScale(columnType.Precision(), columnType.Scale())
					if err != nil {
						return nil, nil, err
					}
				}
			case oid.T_oid:
				doltgresType = pgtypes.Oid
			case oid.T_regclass:
				doltgresType = pgtypes.Regclass
			case oid.T_regproc:
				doltgresType = pgtypes.Regproc
			case oid.T_regtype:
				doltgresType = pgtypes.Regtype
			case oid.T_text:
				doltgresType = pgtypes.Text
			case oid.T_time:
				doltgresType = pgtypes.Time
			case oid.T_timestamp:
				doltgresType = pgtypes.Timestamp
			case oid.T_timestamptz:
				doltgresType = pgtypes.TimestampTZ
			case oid.T_timetz:
				doltgresType = pgtypes.TimeTZ
			case oid.T_uuid:
				doltgresType = pgtypes.Uuid
			case oid.T_varchar:
				width := uint32(columnType.Width())
				if width > pgtypes.StringMaxLength {
					return nil, nil, errors.Errorf("length for type varchar cannot exceed %d", pgtypes.StringMaxLength)
				} else if width == 0 {
					// TODO: need to differentiate between definitions 'varchar' (valid) and 'varchar(0)' (invalid)
					doltgresType = pgtypes.VarChar
				} else {
					doltgresType, err = pgtypes.NewVarCharType(int32(width))
					if err != nil {
						return nil, nil, err
					}
				}
			case oid.T_xid:
				doltgresType = pgtypes.Xid
			case oid.T_bit:
				width := uint32(columnType.Width())
				if width > pgtypes.StringMaxLength {
					return nil, nil, errors.Errorf("length for type bit cannot exceed %d", pgtypes.StringMaxLength)
				} else if width == 0 {
					// TODO: need to differentiate between definitions 'bit' (valid) and 'bit(0)' (invalid)
					doltgresType = pgtypes.Bit
				} else {
					doltgresType, err = pgtypes.NewBitType(int32(width))
					if err != nil {
						return nil, nil, err
					}
				}
			case oid.T_varbit:
				width := uint32(columnType.Width())
				if width > pgtypes.StringMaxLength {
					return nil, nil, errors.Errorf("length for type varbit cannot exceed %d", pgtypes.StringMaxLength)
				} else if width == 0 {
					// TODO: need to differentiate between definitions 'varbit' (valid) and 'varbit(0)' (invalid)
					doltgresType = pgtypes.VarBit
				} else {
					doltgresType, err = pgtypes.NewVarBitType(int32(width))
					if err != nil {
						return nil, nil, err
					}
				}
			default:
				return nil, nil, errors.Errorf("unknown type with oid: %d", uint32(columnType.Oid()))
			}
		}
	default:
		doltgresType = pgtypes.NewUnresolvedDoltgresType("", strings.ToLower(typ.SQLString()))
	}

	return &vitess.ConvertType{
		Type:    columnTypeName,
		Length:  columnTypeLength,
		Scale:   columnTypeScale,
		Charset: "", // TODO
	}, doltgresType, nil
}
