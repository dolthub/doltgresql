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
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Text, pgtypes.Oid, pgtypes.Int32}, // cstring
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		input := val1.(string)
		typmod := val3.(int32)
		maxChars := typmod //TODO: decode
		if maxChars == pgtypes.StringUnbounded {
			return input, nil
		}
		input, runeLength := truncateString(input, maxChars)
		if runeLength > maxChars {
			return input, fmt.Errorf("value too long for type %s", "varchar")
		} else {
			return input, nil
		}
	},
}

// varcharout represents the PostgreSQL function of varchar type IO output.
var varcharout = framework.Function1{
	Name:       "varcharout",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.VarChar},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		//if b.IsUnbounded() {
		//	return val.(string), nil
		//}
		//str, _ := truncateString(converted.(string), b.MaxChars)
		return val.(string), nil
	},
}

// varcharrecv represents the PostgreSQL function of varchar type IO receive.
var varcharrecv = framework.Function3{
	Name:       "varcharrecv",
	Return:     pgtypes.VarChar,
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		// TODO: should the value be converted here according to typmod?
		switch v := val1.(type) {
		case string:
			return v, nil
		default:
			return nil, pgtypes.ErrUnhandledType.New("varchar", v)
		}
	},
}

// varcharsend represents the PostgreSQL function of varchar type IO send.
var varcharsend = framework.Function1{
	Name:       "varcharsend",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.VarChar},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		//if b.IsUnbounded() {
		//	return val.(string), nil
		//}
		//str, _ := truncateString(converted.(string), b.MaxChars)
		return []byte(val.(string)), nil
	},
}

// varchartypmodin represents the PostgreSQL function of varchar type IO typmod input.
var varchartypmodin = framework.Function1{
	Name:       "varchartypmodin",
	Return:     pgtypes.Int32,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.TextArray}, // cstring[]
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO: typmod=(precision<<16)∣scale
		return nil, nil
	},
}

// varchartypmodout represents the PostgreSQL function of varchar type IO typmod output.
var varchartypmodout = framework.Function1{
	Name:       "varchartypmodout",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		// Precision = typmod & 0xFFFF
		// Scale = (typmod >> 16) & 0xFFFF
		return nil, nil
	},
}
