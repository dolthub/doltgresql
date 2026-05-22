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
		return apd.NewWithBigInt(apd.NewBigInt(n), 1), true
	case int32:
		return apd.New(int64(n), 1), true
	case *apd.Decimal:
		return n, true
	}
	return nil, false
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
				f, _ := d.Float64()
				return float32(f), nil
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
				f, _ := d.Float64()
				return f, nil
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
				// TODO: range check the value fits in int16, return an error if not
				i, err := d.Int64()
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
				// TODO: range check the value fits in int32, return an error if not
				i, err := d.Int64()
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
				// TODO: range check the value fits in int64, return an error if not
				i, err := d.Int64()
				if err != nil {
					return nil, err
				}
				return i, nil
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
				return d, nil
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
