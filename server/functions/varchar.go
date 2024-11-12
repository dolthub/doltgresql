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

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// initVarChar registers the functions to the catalog.
func initVarChar() {
	framework.RegisterFunction(varcharin)
	framework.RegisterFunction(varcharout)
	framework.RegisterFunction(varcharrecv)
	framework.RegisterFunction(varcharsend)
	framework.RegisterFunction(varchartypmodin)
	framework.RegisterFunction(varchartypmodout)
}

// varcharin represents the PostgreSQL function of varchar type IO input.
var varcharin = framework.Function3{
	Name:       "varcharin",
	Return:     pgtypes.VarChar,
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Cstring, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		input := val1.(string)
		typmod := val3.(int32)
		maxChars := pgtypes.GetCharLengthFromTypmod(typmod)
		if maxChars < pgtypes.StringUnbounded {
			return input, nil
		}
		input, runeLength := truncateString(input, maxChars)
		if runeLength > maxChars {
			return input, fmt.Errorf("value too long for type varying(%v)", maxChars)
		} else {
			return input, nil
		}
	},
}

// varcharout represents the PostgreSQL function of varchar type IO output.
var varcharout = framework.Function1{
	Name:       "varcharout",
	Return:     pgtypes.Cstring,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.VarChar},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]pgtypes.DoltgresType, val any) (any, error) {
		v := val.(string)
		typ := t[0]
		if typ.AttTypMod != -1 {
			str, _ := truncateString(v, pgtypes.GetCharLengthFromTypmod(typ.AttTypMod))
			return str, nil
		} else {
			return v, nil
		}
	},
}

// varcharrecv represents the PostgreSQL function of varchar type IO receive.
var varcharrecv = framework.Function3{
	Name:       "varcharrecv",
	Return:     pgtypes.VarChar,
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		data := val1.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		reader := utils.NewReader(data)
		return reader.String(), nil
	},
}

// varcharsend represents the PostgreSQL function of varchar type IO send.
var varcharsend = framework.Function1{
	Name:       "varcharsend",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.VarChar},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		str := val.(string)
		writer := utils.NewWriter(uint64(len(str) + 4))
		writer.String(str)
		return writer.Data(), nil
	},
}

// varchartypmodin represents the PostgreSQL function of varchar type IO typmod input.
var varchartypmodin = framework.Function1{
	Name:       "varchartypmodin",
	Return:     pgtypes.Int32,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.CstringArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return getTypModFromStringArr("varchar", val.([]any))
	},
}

// varchartypmodout represents the PostgreSQL function of varchar type IO typmod output.
var varchartypmodout = framework.Function1{
	Name:       "varchartypmodout",
	Return:     pgtypes.Cstring,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		typmod := val.(int32)
		if typmod < 5 {
			return "", nil
		}
		maxChars := pgtypes.GetCharLengthFromTypmod(typmod)
		return fmt.Sprintf("(%v)", maxChars), nil
	},
}
