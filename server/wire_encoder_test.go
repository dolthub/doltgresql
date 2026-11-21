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
	"bytes"
	"math"
	"sync"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dolthub/doltgresql/server/config"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

func TestWireRowEncoderMatchesGenericPath(t *testing.T) {
	initWireEncoderTestFunctions()
	ctx := sql.NewEmptyContext()
	schema := sql.Schema{
		&sql.Column{Name: "b", Type: pgtypes.Bool},
		&sql.Column{Name: "i2", Type: pgtypes.Int16},
		&sql.Column{Name: "i4", Type: pgtypes.Int32},
		&sql.Column{Name: "i8", Type: pgtypes.Int64},
		&sql.Column{Name: "f4", Type: pgtypes.Float32},
		&sql.Column{Name: "f8", Type: pgtypes.Float64},
		&sql.Column{Name: "s", Type: pgtypes.Text},
		&sql.Column{Name: "n", Type: pgtypes.Text},
	}
	rows := []sql.Row{
		{true, int16(-12), int32(123456), int64(-9876543210), float32(1.25), float64(-3.5), "hello", nil},
		{false, int16(0), int32(-1), int64(0), float32(0), float64(1.0 / 3.0), "", "世界"},
	}
	plan, err := newWireRowEncoder(ctx, schema, make([]int16, len(schema)))
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		want, err := rowToBytes(ctx, schema, row, nil)
		if err != nil {
			t.Fatal(err)
		}
		got, err := plan.encode(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != len(want) {
			t.Fatalf("length: got %d want %d", len(got), len(want))
		}
		for i := range got {
			if !bytes.Equal(got[i], want[i]) {
				t.Errorf("column %d: got %q want %q", i, got[i], want[i])
			}
		}
	}
}

func TestWireRowEncoderFallsBackForNonSimpleTypes(t *testing.T) {
	initWireEncoderTestFunctions()
	ctx := sql.NewEmptyContext()
	schema := sql.Schema{
		&sql.Column{Name: "limited", Type: pgtypes.VarChar.WithAttTypMod(7)},
		&sql.Column{Name: "array", Type: pgtypes.Int32Array},
	}
	row := sql.Row{"abcdef", []any{int32(1), nil, int32(3)}}
	plan, err := newWireRowEncoder(ctx, schema, make([]int16, len(schema)))
	if err != nil {
		t.Fatal(err)
	}
	want, err := rowToBytes(ctx, schema, row, nil)
	if err != nil {
		t.Fatal(err)
	}
	got, err := plan.encode(ctx, row)
	if err != nil {
		t.Fatal(err)
	}
	for i := range want {
		if !bytes.Equal(got[i], want[i]) {
			t.Fatalf("column %d: got %q want %q", i, got[i], want[i])
		}
	}
}

func TestWireRowEncoderPreservesBinaryFormat(t *testing.T) {
	initWireEncoderTestFunctions()
	ctx := sql.NewEmptyContext()
	schema := sql.Schema{&sql.Column{Name: "i4", Type: pgtypes.Int32}}
	row := sql.Row{int32(-123456)}
	formats := []int16{pgtype.BinaryFormatCode}
	plan, err := newWireRowEncoder(ctx, schema, formats)
	if err != nil {
		t.Fatal(err)
	}
	want, err := rowToBytes(ctx, schema, row, formats)
	if err != nil {
		t.Fatal(err)
	}
	got, err := plan.encode(ctx, row)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got[0], want[0]) {
		t.Fatalf("got %v want %v", got[0], want[0])
	}
}

func TestWireRowEncoderPreservesSpecialFloats(t *testing.T) {
	initWireEncoderTestFunctions()
	ctx := sql.NewEmptyContext()
	schema := sql.Schema{
		&sql.Column{Name: "f4", Type: pgtypes.Float32},
		&sql.Column{Name: "f8", Type: pgtypes.Float64},
	}
	rows := []sql.Row{
		{float32(math.NaN()), math.Inf(1)},
		{float32(math.Inf(-1)), math.Copysign(0, -1)},
	}
	plan, err := newWireRowEncoder(ctx, schema, make([]int16, len(schema)))
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range rows {
		want, err := rowToBytes(ctx, schema, row, nil)
		if err != nil {
			t.Fatal(err)
		}
		got, err := plan.encode(ctx, row)
		if err != nil {
			t.Fatal(err)
		}
		for i := range want {
			if !bytes.Equal(got[i], want[i]) {
				t.Fatalf("column %d: got %q want %q", i, got[i], want[i])
			}
		}
	}
}

func TestWireRowEncoderFormatsFiniteFloatBoundariesLikePostgres(t *testing.T) {
	initWireEncoderTestFunctions()
	ctx := sql.NewEmptyContext()
	schema := sql.Schema{
		&sql.Column{Name: "f4_min", Type: pgtypes.Float32},
		&sql.Column{Name: "f4_max", Type: pgtypes.Float32},
		&sql.Column{Name: "f8_min", Type: pgtypes.Float64},
		&sql.Column{Name: "f8_max", Type: pgtypes.Float64},
	}
	row := sql.Row{
		float32(1.17549435e-38),
		float32(3.4028235e38),
		2.2250738585072014e-308,
		1.7976931348623157e308,
	}
	want := [][]byte{
		[]byte("1.1754944e-38"),
		[]byte("3.4028235e+38"),
		[]byte("2.2250738585072014e-308"),
		[]byte("1.7976931348623157e+308"),
	}
	plan, err := newWireRowEncoder(ctx, schema, make([]int16, len(schema)))
	if err != nil {
		t.Fatal(err)
	}
	got, err := plan.encode(ctx, row)
	if err != nil {
		t.Fatal(err)
	}
	for i := range want {
		if !bytes.Equal(got[i], want[i]) {
			t.Fatalf("column %d: got %q want PostgreSQL output %q", i, got[i], want[i])
		}
	}
}

func TestWireRowEncoderRejectsMismatchedShapes(t *testing.T) {
	if _, err := newWireRowEncoder(sql.NewEmptyContext(), sql.Schema{&sql.Column{Type: pgtypes.Text}}, nil); err == nil {
		t.Fatal("expected mismatched schema and format code counts to return an error")
	}
	plan := &wireRowEncoder{columns: []wireColumnEncoder{genericWireColumnEncoder(pgtypes.Text, pgtype.TextFormatCode)}}
	if err := plan.encodeInto(sql.NewEmptyContext(), sql.Row{"a"}, nil); err == nil {
		t.Fatal("expected mismatched output shape to return an error")
	}
	if err := plan.encodeInto(sql.NewEmptyContext(), sql.Row{"a", "b"}, make([][]byte, 2)); err == nil {
		t.Fatal("expected mismatched row shape to return an error")
	}
}

func BenchmarkWireRowEncoding(b *testing.B) {
	initWireEncoderTestFunctions()
	ctx := sql.NewEmptyContext()
	schema := sql.Schema{
		&sql.Column{Name: "id", Type: pgtypes.Int32},
		&sql.Column{Name: "flag", Type: pgtypes.Bool},
		&sql.Column{Name: "value", Type: pgtypes.Float64},
		&sql.Column{Name: "name", Type: pgtypes.Text},
	}
	row := sql.Row{int32(123456), true, 1234.5678, "representative table scan value"}
	plan, err := newWireRowEncoder(ctx, schema, make([]int16, len(schema)))
	if err != nil {
		b.Fatal(err)
	}
	b.Run("generic", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if _, err := rowToBytes(ctx, schema, row, nil); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("planned", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			if _, err := plan.encode(ctx, row); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func initWireEncoderTestFunctions() {
	wireEncoderInitOnce.Do(func() {
		pgtypes.Init()
		config.Init()
		functions.Init()
		framework.Initialize(nil)
	})
}

var wireEncoderInitOnce sync.Once
