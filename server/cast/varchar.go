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
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/postgres/parser/uuid"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// init handles all explicit and implicit casts that are built-in. This comprises only the "From" types.
func init() {
	varcharExplicit()
	varcharImplicit()
}

// varcharExplicit registers all explicit casts. This comprises only the "From" types.
func varcharExplicit() {
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Bool,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			switch strings.TrimSpace(strings.ToLower(val.(string))) {
			case "true", "y", "ye", "yes", "on", "1", "t":
				return true, nil
			case "false", "n", "no", "off", "0", "f":
				return false, nil
			default:
				return nil, fmt.Errorf("invalid input syntax for type %s: %q", targetType.String(), val)
			}
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.BpChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return handleCharExplicitCast(val.(string), targetType)
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Bytea,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			if strings.HasPrefix(val.(string), `\x`) {
				return hex.DecodeString(val.(string)[2:])
			} else {
				return []byte(val.(string)), nil
			}
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Float32,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			out, err := strconv.ParseFloat(strings.TrimSpace(val.(string)), 32)
			if err != nil {
				return nil, fmt.Errorf("invalid input syntax for type %s: %q", targetType.String(), val.(string))
			}
			return float32(out), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Float64,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			out, err := strconv.ParseFloat(strings.TrimSpace(val.(string)), 64)
			if err != nil {
				return nil, fmt.Errorf("invalid input syntax for type %s: %q", targetType.String(), val.(string))
			}
			return out, nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Int16,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			out, err := strconv.ParseInt(strings.TrimSpace(val.(string)), 10, 16)
			if err != nil {
				return nil, fmt.Errorf("invalid input syntax for type %s: %q", targetType.String(), val.(string))
			}
			if out > 32767 || out < -32768 {
				return nil, fmt.Errorf("value %q is out of range for type %s", val.(string), targetType.String())
			}
			return int16(out), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Int32,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			out, err := strconv.ParseInt(strings.TrimSpace(val.(string)), 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid input syntax for type %s: %q", targetType.String(), val.(string))
			}
			if out > 2147483647 || out < -2147483648 {
				return nil, fmt.Errorf("value %q is out of range for type %s", val.(string), targetType.String())
			}
			return int32(out), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Int64,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			out, err := strconv.ParseInt(strings.TrimSpace(val.(string)), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid input syntax for type %s: %q", targetType.String(), val.(string))
			}
			return out, nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Numeric,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			d, err := decimal.NewFromString(strings.TrimSpace(val.(string)))
			if err != nil {
				return nil, fmt.Errorf("invalid input syntax for type %s: %q", targetType.String(), val.(string))
			}
			return d, nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Text,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return val, nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Uuid,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			u, err := uuid.FromString(strings.TrimSpace(val.(string)))
			if err != nil {
				return nil, fmt.Errorf("invalid input syntax for type %s: %q", targetType.String(), val.(string))
			}
			return u, nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.VarChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return handleCharExplicitCast(val.(string), targetType)
		},
	})
}

// varcharImplicit registers all implicit casts. This comprises only the "From" types.
func varcharImplicit() {
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Bool,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			switch strings.TrimSpace(strings.ToLower(val.(string))) {
			case "true", "y", "ye", "yes", "on", "1", "t":
				return true, nil
			case "false", "n", "no", "off", "0", "f":
				return false, nil
			default:
				return nil, fmt.Errorf("invalid input syntax for type %s: %q", targetType.String(), val)
			}
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.BpChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return handleCharImplicitCast(val.(string), targetType)
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Bytea,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			if strings.HasPrefix(val.(string), `\x`) {
				return hex.DecodeString(val.(string)[2:])
			} else {
				return []byte(val.(string)), nil
			}
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Float32,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			out, err := strconv.ParseFloat(strings.TrimSpace(val.(string)), 32)
			if err != nil {
				return nil, fmt.Errorf("invalid input syntax for type %s: %q", targetType.String(), val.(string))
			}
			return float32(out), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Float64,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			out, err := strconv.ParseFloat(strings.TrimSpace(val.(string)), 64)
			if err != nil {
				return nil, fmt.Errorf("invalid input syntax for type %s: %q", targetType.String(), val.(string))
			}
			return out, nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Int16,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			out, err := strconv.ParseInt(strings.TrimSpace(val.(string)), 10, 16)
			if err != nil {
				return nil, fmt.Errorf("invalid input syntax for type %s: %q", targetType.String(), val.(string))
			}
			if out > 32767 || out < -32768 {
				return nil, fmt.Errorf("value %q is out of range for type %s", val.(string), targetType.String())
			}
			return int16(out), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Int32,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			out, err := strconv.ParseInt(strings.TrimSpace(val.(string)), 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid input syntax for type %s: %q", targetType.String(), val.(string))
			}
			if out > 2147483647 || out < -2147483648 {
				return nil, fmt.Errorf("value %q is out of range for type %s", val.(string), targetType.String())
			}
			return int32(out), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Int64,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			out, err := strconv.ParseInt(strings.TrimSpace(val.(string)), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid input syntax for type %s: %q", targetType.String(), val.(string))
			}
			return out, nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Numeric,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			d, err := decimal.NewFromString(strings.TrimSpace(val.(string)))
			if err != nil {
				return nil, fmt.Errorf("invalid input syntax for type %s: %q", targetType.String(), val.(string))
			}
			return d, nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Text,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return val, nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.Uuid,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			u, err := uuid.FromString(strings.TrimSpace(val.(string)))
			if err != nil {
				return nil, fmt.Errorf("invalid input syntax for type %s: %q", targetType.String(), val.(string))
			}
			return u, nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarChar,
		ToType:   pgtypes.VarChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return handleCharImplicitCast(val.(string), targetType)
		},
	})
}
