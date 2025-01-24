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

package cast

import (
	"github.com/cockroachdb/errors"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initJsonB handles all casts that are built-in. This comprises only the "From" types.
func initJsonB() {
	jsonbExplicit()
	jsonbAssignment()
}

// jsonbExplicit registers all explicit casts. This comprises only the "From" types.
func jsonbExplicit() {
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.JsonB,
		ToType:   pgtypes.Bool,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			switch value := val.(pgtypes.JsonDocument).Value.(type) {
			case pgtypes.JsonValueObject:
				return nil, errors.Errorf("cannot cast jsonb object to type %s", targetType.String())
			case pgtypes.JsonValueArray:
				return nil, errors.Errorf("cannot cast jsonb array to type %s", targetType.String())
			case pgtypes.JsonValueString:
				return nil, errors.Errorf("cannot cast jsonb string to type %s", targetType.String())
			case pgtypes.JsonValueNumber:
				return nil, errors.Errorf("cannot cast jsonb numeric to type %s", targetType.String())
			case pgtypes.JsonValueBoolean:
				return bool(value), nil
			case pgtypes.JsonValueNull:
				return nil, errors.Errorf("cannot cast jsonb null to type %s", targetType.String())
			default:
				return nil, errors.Errorf("")
			}
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.JsonB,
		ToType:   pgtypes.Float32,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			switch value := val.(pgtypes.JsonDocument).Value.(type) {
			case pgtypes.JsonValueObject:
				return nil, errors.Errorf("cannot cast jsonb object to type %s", targetType.String())
			case pgtypes.JsonValueArray:
				return nil, errors.Errorf("cannot cast jsonb array to type %s", targetType.String())
			case pgtypes.JsonValueString:
				return nil, errors.Errorf("cannot cast jsonb string to type %s", targetType.String())
			case pgtypes.JsonValueNumber:
				f, _ := decimal.Decimal(value).Float64()
				return float32(f), nil
			case pgtypes.JsonValueBoolean:
				return nil, errors.Errorf("cannot cast jsonb boolean to type %s", targetType.String())
			case pgtypes.JsonValueNull:
				return nil, errors.Errorf("cannot cast jsonb null to type %s", targetType.String())
			default:
				return nil, errors.Errorf("")
			}
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.JsonB,
		ToType:   pgtypes.Float64,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			switch value := val.(pgtypes.JsonDocument).Value.(type) {
			case pgtypes.JsonValueObject:
				return nil, errors.Errorf("cannot cast jsonb object to type %s", targetType.String())
			case pgtypes.JsonValueArray:
				return nil, errors.Errorf("cannot cast jsonb array to type %s", targetType.String())
			case pgtypes.JsonValueString:
				return nil, errors.Errorf("cannot cast jsonb string to type %s", targetType.String())
			case pgtypes.JsonValueNumber:
				f, _ := decimal.Decimal(value).Float64()
				return f, nil
			case pgtypes.JsonValueBoolean:
				return nil, errors.Errorf("cannot cast jsonb boolean to type %s", targetType.String())
			case pgtypes.JsonValueNull:
				return nil, errors.Errorf("cannot cast jsonb null to type %s", targetType.String())
			default:
				return nil, errors.Errorf("")
			}
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.JsonB,
		ToType:   pgtypes.Int16,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			switch value := val.(pgtypes.JsonDocument).Value.(type) {
			case pgtypes.JsonValueObject:
				return nil, errors.Errorf("cannot cast jsonb object to type %s", targetType.String())
			case pgtypes.JsonValueArray:
				return nil, errors.Errorf("cannot cast jsonb array to type %s", targetType.String())
			case pgtypes.JsonValueString:
				return nil, errors.Errorf("cannot cast jsonb string to type %s", targetType.String())
			case pgtypes.JsonValueNumber:
				d := decimal.Decimal(value)
				if d.LessThan(pgtypes.NumericValueMinInt16) || d.GreaterThan(pgtypes.NumericValueMaxInt16) {
					return nil, errors.Errorf("smallint out of range")
				}
				return int16(d.IntPart()), nil
			case pgtypes.JsonValueBoolean:
				return nil, errors.Errorf("cannot cast jsonb boolean to type %s", targetType.String())
			case pgtypes.JsonValueNull:
				return nil, errors.Errorf("cannot cast jsonb null to type %s", targetType.String())
			default:
				return nil, errors.Errorf("")
			}
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.JsonB,
		ToType:   pgtypes.Int32,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			switch value := val.(pgtypes.JsonDocument).Value.(type) {
			case pgtypes.JsonValueObject:
				return nil, errors.Errorf("cannot cast jsonb object to type %s", targetType.String())
			case pgtypes.JsonValueArray:
				return nil, errors.Errorf("cannot cast jsonb array to type %s", targetType.String())
			case pgtypes.JsonValueString:
				return nil, errors.Errorf("cannot cast jsonb string to type %s", targetType.String())
			case pgtypes.JsonValueNumber:
				d := decimal.Decimal(value)
				if d.LessThan(pgtypes.NumericValueMinInt32) || d.GreaterThan(pgtypes.NumericValueMaxInt32) {
					return nil, errors.Errorf("integer out of range")
				}
				return int32(d.IntPart()), nil
			case pgtypes.JsonValueBoolean:
				return nil, errors.Errorf("cannot cast jsonb boolean to type %s", targetType.String())
			case pgtypes.JsonValueNull:
				return nil, errors.Errorf("cannot cast jsonb null to type %s", targetType.String())
			default:
				return nil, errors.Errorf("")
			}
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.JsonB,
		ToType:   pgtypes.Int64,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			switch value := val.(pgtypes.JsonDocument).Value.(type) {
			case pgtypes.JsonValueObject:
				return nil, errors.Errorf("cannot cast jsonb object to type %s", targetType.String())
			case pgtypes.JsonValueArray:
				return nil, errors.Errorf("cannot cast jsonb array to type %s", targetType.String())
			case pgtypes.JsonValueString:
				return nil, errors.Errorf("cannot cast jsonb string to type %s", targetType.String())
			case pgtypes.JsonValueNumber:
				d := decimal.Decimal(value)
				if d.LessThan(pgtypes.NumericValueMinInt64) || d.GreaterThan(pgtypes.NumericValueMaxInt64) {
					return nil, errors.Errorf("bigint out of range")
				}
				return int64(d.IntPart()), nil
			case pgtypes.JsonValueBoolean:
				return nil, errors.Errorf("cannot cast jsonb boolean to type %s", targetType.String())
			case pgtypes.JsonValueNull:
				return nil, errors.Errorf("cannot cast jsonb null to type %s", targetType.String())
			default:
				return nil, errors.Errorf("")
			}
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.JsonB,
		ToType:   pgtypes.Numeric,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			switch value := val.(pgtypes.JsonDocument).Value.(type) {
			case pgtypes.JsonValueObject:
				return nil, errors.Errorf("cannot cast jsonb object to type %s", targetType.String())
			case pgtypes.JsonValueArray:
				return nil, errors.Errorf("cannot cast jsonb array to type %s", targetType.String())
			case pgtypes.JsonValueString:
				return nil, errors.Errorf("cannot cast jsonb string to type %s", targetType.String())
			case pgtypes.JsonValueNumber:
				return decimal.Decimal(value), nil
			case pgtypes.JsonValueBoolean:
				return nil, errors.Errorf("cannot cast jsonb boolean to type %s", targetType.String())
			case pgtypes.JsonValueNull:
				return nil, errors.Errorf("cannot cast jsonb null to type %s", targetType.String())
			default:
				return nil, errors.Errorf("")
			}
		},
	})
}

// jsonbAssignment registers all assignment casts. This comprises only the "From" types.
func jsonbAssignment() {
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.JsonB,
		ToType:   pgtypes.Json,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			return pgtypes.JsonB.IoOutput(ctx, val)
		},
	})
}
