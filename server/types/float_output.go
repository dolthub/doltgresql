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
	"math"
	"strconv"

	"github.com/dolthub/go-mysql-server/sql"
)

// ExtraFloatDigits returns the PostgreSQL float-output precision setting.
func ExtraFloatDigits(ctx *sql.Context) (int, error) {
	if ctx == nil {
		return 1, nil
	}
	value, err := ctx.GetSessionVariable(ctx, "extra_float_digits")
	if err != nil {
		return 0, err
	}
	switch value := value.(type) {
	case int:
		return value, nil
	case int64:
		return int(value), nil
	default:
		return 0, fmt.Errorf("extra_float_digits has unexpected type %T", value)
	}
}

// AppendFloat32Text appends PostgreSQL-compatible float4 text to dst.
func AppendFloat32Text(dst []byte, value float32, extraFloatDigits int) []byte {
	return appendFloatText(dst, float64(value), 32, extraFloatDigits)
}

// AppendFloat64Text appends PostgreSQL-compatible float8 text to dst.
func AppendFloat64Text(dst []byte, value float64, extraFloatDigits int) []byte {
	return appendFloatText(dst, value, 64, extraFloatDigits)
}

// appendFloatText implements PostgreSQL's special values, shortest mode, and legacy precision mode.
func appendFloatText(dst []byte, value float64, bitSize int, extraFloatDigits int) []byte {
	if math.IsInf(value, 1) {
		return append(dst, "Infinity"...)
	}
	if math.IsInf(value, -1) {
		return append(dst, "-Infinity"...)
	}
	if math.IsNaN(value) {
		return append(dst, "NaN"...)
	}
	if extraFloatDigits <= 0 {
		precision := 15 + extraFloatDigits
		if bitSize == 32 {
			precision = 6 + extraFloatDigits
		}
		if precision < 1 {
			precision = 1
		}
		return strconv.AppendFloat(dst, value, 'g', precision, bitSize)
	}

	abs := math.Abs(value)
	format := byte('f')
	lowerFixed := 1e-4
	if bitSize == 32 {
		lowerFixed = float64(float32(1e-4))
	}
	if abs != 0 && (abs < lowerFixed || bitSize == 32 && abs >= 1e6 || bitSize == 64 && abs >= 1e15) {
		format = 'e'
	}
	return strconv.AppendFloat(dst, value, format, -1, bitSize)
}
