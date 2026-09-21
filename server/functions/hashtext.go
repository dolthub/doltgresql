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

package functions

import (
	"encoding/binary"
	"math/bits"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initHashText registers the functions to the catalog.
func initHashText() {
	framework.RegisterFunction(hashtext_text)
}

// hashtext_text represents the PostgreSQL function of the same name, taking the same parameters.
var hashtext_text = framework.Function1{
	Name:       "hashtext",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		input, err := framework.UnwrapString(ctx, val)
		if err != nil {
			return nil, err
		}
		return int32(postgresHashBytes([]byte(input))), nil
	},
}

// postgresHashBytes implements PostgreSQL's 32-bit hash_bytes routine on a byte slice. PostgreSQL uses this hash for
// deterministic-collation text values; DoltgreSQL currently supports deterministic collations only.
func postgresHashBytes(data []byte) uint32 {
	a := uint32(0x9e3779b9 + len(data) + 3923095)
	b := a
	c := a

	remaining := data
	for len(remaining) >= 12 {
		a += binary.LittleEndian.Uint32(remaining[0:4])
		b += binary.LittleEndian.Uint32(remaining[4:8])
		c += binary.LittleEndian.Uint32(remaining[8:12])
		a, b, c = postgresHashMix(a, b, c)
		remaining = remaining[12:]
	}

	switch len(remaining) {
	case 11:
		c += uint32(remaining[10]) << 24
		fallthrough
	case 10:
		c += uint32(remaining[9]) << 16
		fallthrough
	case 9:
		c += uint32(remaining[8]) << 8
		fallthrough
	case 8:
		b += uint32(remaining[7]) << 24
		fallthrough
	case 7:
		b += uint32(remaining[6]) << 16
		fallthrough
	case 6:
		b += uint32(remaining[5]) << 8
		fallthrough
	case 5:
		b += uint32(remaining[4])
		fallthrough
	case 4:
		a += uint32(remaining[3]) << 24
		fallthrough
	case 3:
		a += uint32(remaining[2]) << 16
		fallthrough
	case 2:
		a += uint32(remaining[1]) << 8
		fallthrough
	case 1:
		a += uint32(remaining[0])
	}

	_, _, c = postgresHashFinal(a, b, c)
	return c
}

func postgresHashMix(a, b, c uint32) (uint32, uint32, uint32) {
	a -= c
	a ^= bits.RotateLeft32(c, 4)
	c += b
	b -= a
	b ^= bits.RotateLeft32(a, 6)
	a += c
	c -= b
	c ^= bits.RotateLeft32(b, 8)
	b += a
	a -= c
	a ^= bits.RotateLeft32(c, 16)
	c += b
	b -= a
	b ^= bits.RotateLeft32(a, 19)
	a += c
	c -= b
	c ^= bits.RotateLeft32(b, 4)
	b += a
	return a, b, c
}

func postgresHashFinal(a, b, c uint32) (uint32, uint32, uint32) {
	c ^= b
	c -= bits.RotateLeft32(b, 14)
	a ^= c
	a -= bits.RotateLeft32(c, 11)
	b ^= a
	b -= bits.RotateLeft32(a, 25)
	c ^= b
	c -= bits.RotateLeft32(b, 16)
	a ^= c
	a -= bits.RotateLeft32(c, 4)
	b ^= a
	b -= bits.RotateLeft32(a, 14)
	c ^= b
	c -= bits.RotateLeft32(b, 24)
	return a, b, c
}
