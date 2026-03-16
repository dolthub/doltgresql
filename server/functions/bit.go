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
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
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
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, _, val3 any) (any, error) {
		input := val1.(string)
		typmod := val3.(int32)

		// validation and normalization
		array, err := tree.ParseDBitArray(input)
		if err != nil {
			return nil, err
		}

		if array.BitLen() != uint(typmod) {
			return nil, pgtypes.ErrWrongLengthBit.New(len(input), typmod)
		}

		return tree.AsStringWithFlags(array, tree.FmtPgwireText), nil
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
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		if wrapper, ok := val.(sql.AnyWrapper); ok {
			var err error
			val, err = wrapper.UnwrapAny(ctx)
			if err != nil {
				return nil, err
			}
			if val == nil {
				return nil, nil
			}
		}
		// We process bits in chunks of 8, so we append zeroes until our string is evenly divisible by 8
		bitString := val.(string)
		if len(bitString)%8 != 0 {
			bitString += strings.Repeat("0", 8-(len(bitString)%8))
		}
		writer := utils.NewWireWriter()
		writer.Reserve(uint64(4 + (len(bitString) / 8)))
		writer.WriteInt32(t[0].GetAttTypMod())
		for bufIdx := 0; bufIdx < len(bitString); bufIdx += 8 {
			parsedByte, err := strconv.ParseUint(bitString[bufIdx:bufIdx+8], 2, 8)
			if err != nil {
				return nil, errors.Errorf(`error encountered while converting "BIT" to binary wire format:\n%s`, err.Error())
			}
			writer.WriteUint8(byte(parsedByte))
		}
		return writer.BufferData(), nil
	},
}

// bittypmodin represents the PostgreSQL function of bit type IO typmod input.
var bittypmodin = framework.Function1{
	Name:       "bittypmodin",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.CstringArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		typmod, err := getTypModFromStringArr("bit", val.([]any))
		if err != nil {
			return nil, err
		}
		// getTypModFromStringArr always adds 4, so we remove 4 here since it doesn't apply to bit types
		return typmod - 4, nil
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
		if typmod < 1 {
			return "", nil
		}
		return fmt.Sprintf("(%v)", typmod), nil
	},
}
