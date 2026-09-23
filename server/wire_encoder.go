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

package server

import (
	"fmt"
	"strconv"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/encodings"
	"github.com/jackc/pgx/v5/pgtype"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// wireRowEncoder resolves the text/binary conversion for each result column
// once per result iterator. The fallback deliberately remains rowToBytes'
// general conversion path so domains, arrays, session-sensitive types and all
// binary formats retain their existing semantics.
type wireRowEncoder struct {
	columns []wireColumnEncoder
}

// wireColumnEncoder converts one non-NULL result value to its wire payload.
type wireColumnEncoder func(*sql.Context, any) ([]byte, error)

// newWireRowEncoder plans one encoder for each schema column and canonical format code.
func newWireRowEncoder(ctx *sql.Context, schema sql.Schema, formatCodes []int16) (*wireRowEncoder, error) {
	if len(formatCodes) != len(schema) {
		return nil, fmt.Errorf("wire schema has %d columns and %d format codes", len(schema), len(formatCodes))
	}
	extraFloatDigits := 1
	floatSettingLoaded := false
	var err error
	p := &wireRowEncoder{columns: make([]wireColumnEncoder, len(schema))}
	for i, col := range schema {
		if !floatSettingLoaded && formatCodes[i] == pgtype.TextFormatCode {
			if pgType, ok := col.Type.(*pgtypes.DoltgresType); ok && pgType.TypType == pgtypes.TypeType_Base &&
				(pgType.ID == pgtypes.Float32.ID || pgType.ID == pgtypes.Float64.ID) {
				extraFloatDigits, err = pgtypes.ExtraFloatDigits(ctx)
				if err != nil {
					return nil, err
				}
				floatSettingLoaded = true
			}
		}
		p.columns[i] = planWireColumnEncoder(col.Type, formatCodes[i], extraFloatDigits)
	}
	return p, nil
}

// planWireColumnEncoder selects a proven specialized text encoder or the generic path.
func planWireColumnEncoder(typ sql.Type, format int16, extraFloatDigits int) wireColumnEncoder {
	pgType, ok := typ.(*pgtypes.DoltgresType)
	if format != pgtype.TextFormatCode || !ok || pgType.TypType != pgtypes.TypeType_Base {
		return genericWireColumnEncoder(typ, format)
	}
	fallback := genericWireColumnEncoder(typ, format)
	switch pgType.ID {
	case pgtypes.Bool.ID:
		return func(ctx *sql.Context, value any) ([]byte, error) {
			v, ok := value.(bool)
			if !ok {
				return fallback(ctx, value)
			}
			if v {
				return []byte{'t'}, nil
			}
			return []byte{'f'}, nil
		}
	case pgtypes.Int16.ID:
		return func(ctx *sql.Context, value any) ([]byte, error) {
			if v, ok := value.(int16); ok {
				return strconv.AppendInt(nil, int64(v), 10), nil
			}
			return fallback(ctx, value)
		}
	case pgtypes.Int32.ID:
		return func(ctx *sql.Context, value any) ([]byte, error) {
			if v, ok := value.(int32); ok {
				return strconv.AppendInt(nil, int64(v), 10), nil
			}
			return fallback(ctx, value)
		}
	case pgtypes.Int64.ID:
		return func(ctx *sql.Context, value any) ([]byte, error) {
			if v, ok := value.(int64); ok {
				return strconv.AppendInt(nil, v, 10), nil
			}
			return fallback(ctx, value)
		}
	case pgtypes.Float32.ID:
		return func(ctx *sql.Context, value any) ([]byte, error) {
			if v, ok := value.(float32); ok {
				return pgtypes.AppendFloat32Text(nil, v, extraFloatDigits), nil
			}
			return fallback(ctx, value)
		}
	case pgtypes.Float64.ID:
		return func(ctx *sql.Context, value any) ([]byte, error) {
			if v, ok := value.(float64); ok {
				return pgtypes.AppendFloat64Text(nil, v, extraFloatDigits), nil
			}
			return fallback(ctx, value)
		}
	case pgtypes.Text.ID:
		return stringWireColumnEncoder(fallback)
	case pgtypes.VarChar.ID:
		if pgType.GetAttTypMod() == -1 {
			return stringWireColumnEncoder(fallback)
		}
		return fallback
	default:
		return fallback
	}
}

// stringWireColumnEncoder returns string bytes without copying when the runtime type matches.
func stringWireColumnEncoder(fallback wireColumnEncoder) wireColumnEncoder {
	return func(ctx *sql.Context, value any) ([]byte, error) {
		if v, ok := value.(string); ok {
			return encodings.StringToBytes(v), nil
		}
		return fallback(ctx, value)
	}
}

// genericWireColumnEncoder preserves the existing conversion for unsupported columns.
func genericWireColumnEncoder(typ sql.Type, format int16) wireColumnEncoder {
	return func(ctx *sql.Context, v any) ([]byte, error) {
		return valueToBytes(ctx, typ, format, v)
	}
}

// encode allocates column metadata and encodes one row into it.
func (p *wireRowEncoder) encode(ctx *sql.Context, row sql.Row) ([][]byte, error) {
	out := make([][]byte, len(row))
	if err := p.encodeInto(ctx, row, out); err != nil {
		return nil, err
	}
	return out, nil
}

// encodeInto encodes one row into caller-owned column metadata.
func (p *wireRowEncoder) encodeInto(ctx *sql.Context, row sql.Row, out [][]byte) error {
	if len(row) != len(p.columns) || len(out) != len(row) {
		return fmt.Errorf("wire row has %d values, %d encoders, and %d output slots", len(row), len(p.columns), len(out))
	}
	for i, value := range row {
		if value == nil {
			out[i] = nil
			continue
		}
		encoded, err := p.columns[i](ctx, value)
		if err != nil {
			return err
		}
		out[i] = encoded
	}
	return nil
}
