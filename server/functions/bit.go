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

package functions

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// initBit registers the functions to the catalog.
func initBit() {
	framework.RegisterFunction(bitin)
	framework.RegisterFunction(bitout)
	framework.RegisterFunction(bitrecv)
	framework.RegisterFunction(bitsend)
	framework.RegisterFunction(bittypmodin)
	framework.RegisterFunction(bittypmodout)
}

// bitin represents the PostgreSQL function of bit type IO input.
var bitin = framework.Function3{
	Name:       "bit_in",
	Return:     pgtypes.Bit,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Cstring, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		input := val1.(string)
		typmod := val3.(int32)
		
		// Parse bit string - remove leading 'B' or 'b' prefix if present
		bitStr := strings.TrimSpace(input)
		if len(bitStr) > 0 && (bitStr[0] == 'B' || bitStr[0] == 'b') {
			bitStr = bitStr[1:]
			// Remove quotes if present
			if len(bitStr) > 0 && (bitStr[0] == '\'' || bitStr[0] == '"') {
				if len(bitStr) > 1 && bitStr[len(bitStr)-1] == bitStr[0] {
					bitStr = bitStr[1 : len(bitStr)-1]
				}
			}
		}
		
		// Validate that all characters are '0' or '1'
		for _, r := range bitStr {
			if r != '0' && r != '1' {
				return nil, pgtypes.ErrInvalidSyntaxForType.New("bit", input)
			}
		}
		
		// Check length against typmod
		if typmod != -1 {
			expectedLength := pgtypes.GetCharLengthFromTypmod(typmod)
			if int32(len(bitStr)) != expectedLength {
				return nil, pgtypes.ErrInvalidSyntaxForType.New("bit", input)
			}
		}
		
		return bitStr, nil
	},
}

// bitout represents the PostgreSQL function of bit type IO output.
var bitout = framework.Function1{
	Name:       "bit_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Bit},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		bitStr := val.(string)
		typ := t[0]
		tm := typ.GetAttTypMod()
		if tm != -1 {
			expectedLength := pgtypes.GetCharLengthFromTypmod(tm)
			// Pad with zeros if needed (shouldn't happen for fixed-length bit)
			if int32(len(bitStr)) < expectedLength {
				bitStr = bitStr + strings.Repeat("0", int(expectedLength-int32(len(bitStr))))
			} else if int32(len(bitStr)) > expectedLength {
				bitStr = bitStr[:expectedLength]
			}
		}
		return bitStr, nil
	},
}

// bitrecv represents the PostgreSQL function of bit type IO receive.
var bitrecv = framework.Function3{
	Name:       "bit_recv",
	Return:     pgtypes.Bit,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		data := val1.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		reader := utils.NewReader(data)
		return reader.String(), nil
	},
}

// bitsend represents the PostgreSQL function of bit type IO send.
var bitsend = framework.Function1{
	Name:       "bit_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Bit},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		bitStr := val.(string)
		writer := utils.NewWriter(uint64(len(bitStr) + 4))
		writer.String(bitStr)
		return writer.Data(), nil
	},
}

// bittypmodin represents the PostgreSQL function of bit type IO typmod input.
var bittypmodin = framework.Function1{
	Name:       "bittypmodin",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.CstringArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return getTypModFromStringArr("bit", val.([]any))
	},
}

// bittypmodout represents the PostgreSQL function of bit type IO typmod output.
var bittypmodout = framework.Function1{
	Name:       "bittypmodout",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		typmod := val.(int32)
		if typmod < 5 {
			return "", nil
		}
		bitLength := pgtypes.GetCharLengthFromTypmod(typmod)
		return fmt.Sprintf("(%v)", bitLength), nil
	},
}
