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
	"math"
	"testing"
)

// TestExtraFloatDigitsWithoutContextUsesDefault verifies context-free formatting uses PostgreSQL's default precision.
func TestExtraFloatDigitsWithoutContextUsesDefault(t *testing.T) {
	got, err := ExtraFloatDigits(nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != 1 {
		t.Fatalf("got %d, want 1", got)
	}
}

// TestAppendFloatTextMatchesPostgres verifies default and legacy output modes against PostgreSQL 15.
func TestAppendFloatTextMatchesPostgres(t *testing.T) {
	tests := []struct {
		name             string
		got              []byte
		extraFloatDigits int
		want             string
	}{
		{"float4 smallest subnormal", AppendFloat32Text(nil, math.SmallestNonzeroFloat32, 1), 1, "1e-45"},
		{"float8 smallest subnormal", AppendFloat64Text(nil, math.SmallestNonzeroFloat64, 1), 1, "5e-324"},
		{"float4 negative zero", AppendFloat32Text(nil, float32(math.Copysign(0, -1)), 1), 1, "-0"},
		{"float8 negative zero", AppendFloat64Text(nil, math.Copysign(0, -1), 1), 1, "-0"},
		{"float4 lower fixed", AppendFloat32Text(nil, 1e-4, 1), 1, "0.0001"},
		{"float4 lower exponent", AppendFloat32Text(nil, 1e-5, 1), 1, "1e-05"},
		{"float4 upper fixed", AppendFloat32Text(nil, 1e5, 1), 1, "100000"},
		{"float4 upper exponent", AppendFloat32Text(nil, 1e6, 1), 1, "1e+06"},
		{"float8 lower fixed", AppendFloat64Text(nil, 1e-4, 1), 1, "0.0001"},
		{"float8 lower exponent", AppendFloat64Text(nil, 1e-5, 1), 1, "1e-05"},
		{"float8 upper fixed", AppendFloat64Text(nil, 1e14, 1), 1, "100000000000000"},
		{"float8 upper exponent", AppendFloat64Text(nil, 1e15, 1), 1, "1e+15"},
		{"float4 legacy precision", AppendFloat32Text(nil, float32(1.17549435e-38), 0), 0, "1.17549e-38"},
		{"float8 legacy precision", AppendFloat64Text(nil, 1.234567890123456, 0), 0, "1.23456789012346"},
		{"float4 minimum precision", AppendFloat32Text(nil, float32(1.17549435e-38), -15), -15, "1e-38"},
		{"float8 minimum precision", AppendFloat64Text(nil, 1.234567890123456, -15), -15, "1"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if string(test.got) != test.want {
				t.Fatalf("extra_float_digits=%d: got %q want %q", test.extraFloatDigits, test.got, test.want)
			}
		})
	}
}
