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

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
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

		// validation and normalization
		bitStr, err := tree.ParseDBitArray(input)
		if err != nil {
			return nil, err
		}

		// Check length against typmod (varbit allows up to typmod length)
		if typmod != -1 {
			if int32(bitStr.BitLen()) > typmod {
				return nil, pgtypes.ErrVarBitLengthExceeded.New(typmod)
			}
		}

		return tree.AsStringWithFlags(bitStr, tree.FmtPgwireText), nil
	},
}

// varbitout represents the PostgreSQL function of varbit type IO output.
var varbitout = framework.Function1{
	Name:       "varbit_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.VarBit},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		bitStr, ok, err := sql.Unwrap[string](ctx, val)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("varbit_out function returned false")
		}
		return bitStr, nil
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
		return reader.String(), nil
	},
}

// varbitsend represents the PostgreSQL function of varbit type IO send.
var varbitsend = framework.Function1{
	Name:       "varbit_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.VarBit},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		bitStr := val.(string)
		writer := utils.NewWriter(uint64(len(bitStr) + 4))
		writer.String(bitStr)
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
		typmod, err := getTypModFromStringArr("bit varying", val.([]any))
		if err != nil {
			return nil, err
		}
		// getTypModFromStringArr always adds 4, so we remove 4 here since it doesn't apply to varbit types
		return typmod - 4, nil
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
		if typmod < 1 {
			return "", nil
		}
		return fmt.Sprintf("(%v)", typmod), nil
	},
}
