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

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// initVarBit registers the functions to the catalog.
func initVarBit() {
	framework.RegisterFunction(varbitin)
	framework.RegisterFunction(varbitout)
	framework.RegisterFunction(varbitrecv)
	framework.RegisterFunction(varbitsend)
	framework.RegisterFunction(varbittypmodin)
	framework.RegisterFunction(varbittypmodout)
}

// varbitin represents the PostgreSQL function of varbit type IO input.
var varbitin = framework.Function3{
	Name:       "varbit_in",
	Return:     pgtypes.VarBit,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Cstring, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, _, val3 any) (any, error) {
		input := val1.(string)
		typmod := val3.(int32)

		bitStr, err := tree.ParseDBitArray(input)
		if err != nil {
			return nil, err
		}

		// Check length against typmod (varbit allows up to typmod length)
		if typmod != -1 {
			maxLength := pgtypes.GetCharLengthFromTypmod(typmod)
			if int32(bitStr.BitLen()) > maxLength {
				return nil, pgtypes.ErrVarBitLengthExceeded.New(maxLength)
			}
		}

		return bitStr, nil
	},
}

// varbitout represents the PostgreSQL function of varbit type IO output.
var varbitout = framework.Function1{
	Name:       "varbit_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.VarBit},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		var bitArray *tree.DBitArray
		bitStr, ok, err := sql.Unwrap[string](ctx, val)
		if err != nil {
			return nil, err
		}
		if ok {
			bitArray, err = tree.ParseDBitArray(bitStr)
			if err != nil {
				return nil, err
			}
		} else {
			bitArray = val.(*tree.DBitArray)
		}
		return tree.AsStringWithFlags(bitArray, tree.FmtPgwireText), nil
	},
}

// varbitrecv represents the PostgreSQL function of varbit type IO receive.
var varbitrecv = framework.Function3{
	Name:       "varbit_recv",
	Return:     pgtypes.VarBit,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		data := val1.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		reader := utils.NewReader(data)
		return tree.ParseDBitArray(reader.String())
	},
}

// varbitsend represents the PostgreSQL function of varbit type IO send.
var varbitsend = framework.Function1{
	Name:       "varbit_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.VarBit},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		bitStr := val.(*tree.DBitArray)
		writer := utils.NewWriter(uint64(bitStr.BitLen() + 4))
		writer.String(tree.AsStringWithFlags(bitStr, tree.FmtPgwireText))
		return writer.Data(), nil
	},
}

// varbittypmodin represents the PostgreSQL function of varbit type IO typmod input.
var varbittypmodin = framework.Function1{
	Name:       "varbittypmodin",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.CstringArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return getTypModFromStringArr("bit varying", val.([]any))
	},
}

// varbittypmodout represents the PostgreSQL function of varbit type IO typmod output.
var varbittypmodout = framework.Function1{
	Name:       "varbittypmodout",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		typmod := val.(int32)
		if typmod < 5 {
			return "", nil
		}
		maxLength := pgtypes.GetCharLengthFromTypmod(typmod)
		return fmt.Sprintf("(%v)", maxLength), nil
	},
}
