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
	"encoding/json"
	"math"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initJsonB handles all casts that are built-in. This comprises only the source types.
func initJsonB(builtInCasts map[id.Cast]casts.Cast) {
	jsonbExplicit(builtInCasts)
	jsonbAssignment(builtInCasts)
}

// jsonbGetInterface extracts the native Go value from a JSONB value (sql.JSONWrapper or string).
func jsonbGetInterface(ctx *sql.Context, val any) (any, error) {
	switch v := val.(type) {
	case sql.JSONWrapper:
		return v.ToInterface(ctx)
	case string:
		var result any
		if err := json.Unmarshal([]byte(v), &result); err != nil {
			return nil, errors.Errorf("invalid JSON: %v", err)
		}
		return result, nil
	default:
		return nil, errors.Errorf("unexpected JSONB value type: %T", val)
	}
}

// jsonbNumberToDecimal converts various numeric types from JSON deserialization to decimal.Decimal.
func jsonbNumberToDecimal(v any) (*apd.Decimal, bool) {
	switch n := v.(type) {
	case float64:
		d, _ := apd.New(0, 0).SetFloat64(n)
		return d, true
	case float32:
		d, _ := apd.New(0, 0).SetFloat64(float64(n))
		return d, true
	case json.Number:
		d, _, err := apd.NewFromString(n.String())
		if err != nil {
			return nil, false
		}
		return d, true
	case int64:
		return apd.New(n, 0), true
	case int32:
		return apd.New(int64(n), 0), true
	case *apd.Decimal:
		return n, true
	}
	return nil, false
}

// jsonbDecimalToInt rounds the decimal half-to-even (matching Postgres'
// rules for numeric → integer casts) and then verifies that the rounded
// value lies within [min, max]. Rounding happens before the bounds check so
// values like 32767.4 round down to 32767 and still fit in int16, while
// 32767.5 rounds up to 32768 and is rejected.
func jsonbDecimalToInt(d *apd.Decimal, min, max *apd.Decimal, rangeMsg string) (int64, error) {
	if d.Form != apd.Finite {
		return 0, errors.Wrap(pgtypes.ErrCastOutOfRange, rangeMsg)
	}
	// Round half-to-even ("banker's rounding"), matching Postgres' numeric →
	// integer semantics. The default rounding mode on apd.BaseContext is
	// RoundHalfUp, so we have to construct a context explicitly.
	rounded := new(apd.Decimal)
	p := d.NumDigits() + int64(math.Abs(float64(d.Exponent))) + 1
	ctx := sql.DecimalCtx.WithPrecision(uint32(p))
	ctx.Rounding = apd.RoundHalfEven
	if _, err := ctx.Quantize(rounded, d, 0); err != nil {
		return 0, err
	}
	if rounded.Cmp(min) < 0 || rounded.Cmp(max) > 0 {
		return 0, errors.Wrap(pgtypes.ErrCastOutOfRange, rangeMsg)
	}
	i, err := rounded.Int64()
	if err != nil {
		return 0, err
	}
	return i, nil
}

// jsonbDecimalToFloat64 converts the decimal to a float64, returning an
// out-of-range error when the magnitude exceeds what float64 can represent
// as a finite value.
func jsonbDecimalToFloat64(d *apd.Decimal, rangeMsg string) (float64, error) {
	if d.Form != apd.Finite {
		return 0, errors.Wrap(pgtypes.ErrCastOutOfRange, rangeMsg)
	}
	f, err := d.Float64()
	if err != nil || math.IsInf(f, 0) {
		return 0, errors.Wrap(pgtypes.ErrCastOutOfRange, rangeMsg)
	}
	return f, nil
}

// jsonbDecimalToFloat32 converts the decimal to a float32, returning an
// out-of-range error when the magnitude exceeds what float32 can represent
// as a finite value.
func jsonbDecimalToFloat32(d *apd.Decimal, rangeMsg string) (float32, error) {
	f, err := jsonbDecimalToFloat64(d, rangeMsg)
	if err != nil {
		return 0, err
	}
	f32 := float32(f)
	if math.IsInf(float64(f32), 0) {
		return 0, errors.Wrap(pgtypes.ErrCastOutOfRange, rangeMsg)
	}
	return f32, nil
}

// jsonbExplicit registers all explicit casts. This comprises only the source types.
func jsonbExplicit(builtInCasts map[id.Cast]casts.Cast) {
	framework.MustAddExplicitTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.JsonB,
		ToType:   pgtypes.Bool,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			v, err := jsonbGetInterface(ctx, val)
			if err != nil {
				return nil, err
			}
			switch value := v.(type) {
			case map[string]interface{}:
				return nil, errors.Errorf("cannot cast jsonb object to type %s", targetType.String())
			case []interface{}:
				return nil, errors.Errorf("cannot cast jsonb array to type %s", targetType.String())
			case string:
				return nil, errors.Errorf("cannot cast jsonb string to type %s", targetType.String())
			case bool:
				return value, nil
			case nil:
				return nil, errors.Errorf("cannot cast jsonb null to type %s", targetType.String())
			default:
				if _, ok := jsonbNumberToDecimal(v); ok {
					return nil, errors.Errorf("cannot cast jsonb numeric to type %s", targetType.String())
				}
				return nil, errors.Errorf("unexpected jsonb value type: %T", v)
			}
		},
	})
	framework.MustAddExplicitTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.JsonB,
		ToType:   pgtypes.Float32,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			v, err := jsonbGetInterface(ctx, val)
			if err != nil {
				return nil, err
			}
			switch v.(type) {
			case map[string]interface{}:
				return nil, errors.Errorf("cannot cast jsonb object to type %s", targetType.String())
			case []interface{}:
				return nil, errors.Errorf("cannot cast jsonb array to type %s", targetType.String())
			case string:
				return nil, errors.Errorf("cannot cast jsonb string to type %s", targetType.String())
			case bool:
				return nil, errors.Errorf("cannot cast jsonb boolean to type %s", targetType.String())
			case nil:
				return nil, errors.Errorf("cannot cast jsonb null to type %s", targetType.String())
			default:
				d, ok := jsonbNumberToDecimal(v)
				if !ok {
					return nil, errors.Errorf("unexpected jsonb value type: %T", v)
				}
				return jsonbDecimalToFloat32(d, "real out of range")
			}
		},
	})
	framework.MustAddExplicitTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.JsonB,
		ToType:   pgtypes.Float64,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			v, err := jsonbGetInterface(ctx, val)
			if err != nil {
				return nil, err
			}
			switch v.(type) {
			case map[string]interface{}:
				return nil, errors.Errorf("cannot cast jsonb object to type %s", targetType.String())
			case []interface{}:
				return nil, errors.Errorf("cannot cast jsonb array to type %s", targetType.String())
			case string:
				return nil, errors.Errorf("cannot cast jsonb string to type %s", targetType.String())
			case bool:
				return nil, errors.Errorf("cannot cast jsonb boolean to type %s", targetType.String())
			case nil:
				return nil, errors.Errorf("cannot cast jsonb null to type %s", targetType.String())
			default:
				d, ok := jsonbNumberToDecimal(v)
				if !ok {
					return nil, errors.Errorf("unexpected jsonb value type: %T", v)
				}
				return jsonbDecimalToFloat64(d, "double precision out of range")
			}
		},
	})
	framework.MustAddExplicitTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.JsonB,
		ToType:   pgtypes.Int16,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			v, err := jsonbGetInterface(ctx, val)
			if err != nil {
				return nil, err
			}
			switch v.(type) {
			case map[string]interface{}:
				return nil, errors.Errorf("cannot cast jsonb object to type %s", targetType.String())
			case []interface{}:
				return nil, errors.Errorf("cannot cast jsonb array to type %s", targetType.String())
			case string:
				return nil, errors.Errorf("cannot cast jsonb string to type %s", targetType.String())
			case bool:
				return nil, errors.Errorf("cannot cast jsonb boolean to type %s", targetType.String())
			case nil:
				return nil, errors.Errorf("cannot cast jsonb null to type %s", targetType.String())
			default:
				d, ok := jsonbNumberToDecimal(v)
				if !ok {
					return nil, errors.Errorf("unexpected jsonb value type: %T", v)
				}
				i, err := jsonbDecimalToInt(d, pgtypes.NumericValueMinInt16, pgtypes.NumericValueMaxInt16, "smallint out of range")
				if err != nil {
					return nil, err
				}
				return int16(i), nil
			}
		},
	})
	framework.MustAddExplicitTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.JsonB,
		ToType:   pgtypes.Int32,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			v, err := jsonbGetInterface(ctx, val)
			if err != nil {
				return nil, err
			}
			switch v.(type) {
			case map[string]interface{}:
				return nil, errors.Errorf("cannot cast jsonb object to type %s", targetType.String())
			case []interface{}:
				return nil, errors.Errorf("cannot cast jsonb array to type %s", targetType.String())
			case string:
				return nil, errors.Errorf("cannot cast jsonb string to type %s", targetType.String())
			case bool:
				return nil, errors.Errorf("cannot cast jsonb boolean to type %s", targetType.String())
			case nil:
				return nil, errors.Errorf("cannot cast jsonb null to type %s", targetType.String())
			default:
				d, ok := jsonbNumberToDecimal(v)
				if !ok {
					return nil, errors.Errorf("unexpected jsonb value type: %T", v)
				}
				i, err := jsonbDecimalToInt(d, pgtypes.NumericValueMinInt32, pgtypes.NumericValueMaxInt32, "integer out of range")
				if err != nil {
					return nil, err
				}
				return int32(i), nil
			}
		},
	})
	framework.MustAddExplicitTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.JsonB,
		ToType:   pgtypes.Int64,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			v, err := jsonbGetInterface(ctx, val)
			if err != nil {
				return nil, err
			}
			switch v.(type) {
			case map[string]interface{}:
				return nil, errors.Errorf("cannot cast jsonb object to type %s", targetType.String())
			case []interface{}:
				return nil, errors.Errorf("cannot cast jsonb array to type %s", targetType.String())
			case string:
				return nil, errors.Errorf("cannot cast jsonb string to type %s", targetType.String())
			case bool:
				return nil, errors.Errorf("cannot cast jsonb boolean to type %s", targetType.String())
			case nil:
				return nil, errors.Errorf("cannot cast jsonb null to type %s", targetType.String())
			default:
				d, ok := jsonbNumberToDecimal(v)
				if !ok {
					return nil, errors.Errorf("unexpected jsonb value type: %T", v)
				}
				return jsonbDecimalToInt(d, pgtypes.NumericValueMinInt64, pgtypes.NumericValueMaxInt64, "bigint out of range")
			}
		},
	})
	framework.MustAddExplicitTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.JsonB,
		ToType:   pgtypes.Numeric,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			v, err := jsonbGetInterface(ctx, val)
			if err != nil {
				return nil, err
			}
			switch v.(type) {
			case map[string]interface{}:
				return nil, errors.Errorf("cannot cast jsonb object to type %s", targetType.String())
			case []interface{}:
				return nil, errors.Errorf("cannot cast jsonb array to type %s", targetType.String())
			case string:
				return nil, errors.Errorf("cannot cast jsonb string to type %s", targetType.String())
			case bool:
				return nil, errors.Errorf("cannot cast jsonb boolean to type %s", targetType.String())
			case nil:
				return nil, errors.Errorf("cannot cast jsonb null to type %s", targetType.String())
			default:
				d, ok := jsonbNumberToDecimal(v)
				if !ok {
					return nil, errors.Errorf("unexpected jsonb value type: %T", v)
				}
				// Apply the target type's precision/scale typmod, matching
				// what the numeric → numeric cast does.
				return pgtypes.GetNumericValueWithTypmod(d, targetType.GetAttTypMod())
			}
		},
	})
}

// jsonbAssignment registers all assignment casts. This comprises only the source types.
func jsonbAssignment(builtInCasts map[id.Cast]casts.Cast) {
	framework.MustAddAssignmentTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.JsonB,
		ToType:   pgtypes.Json,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			return pgtypes.JsonB.IoOutput(ctx, val)
		},
	})
}
