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
	"bytes"
	"fmt"
	"github.com/dolthub/doltgresql/utils"
	"strings"
	"unicode/utf8"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initBpChar registers the functions to the catalog.
func initBpChar() {
	framework.RegisterFunction(bpcharin)
	framework.RegisterFunction(bpcharout)
	framework.RegisterFunction(bpcharrecv)
	framework.RegisterFunction(bpcharsend)
	framework.RegisterFunction(bpchartypmodin)
	framework.RegisterFunction(bpchartypmodout)
	framework.RegisterFunction(bpcharcmp)
}

// bpcharin represents the PostgreSQL function of bpchar type IO input.
var bpcharin = framework.Function3{
	Name:       "bpcharin",
	Return:     pgtypes.BpChar,
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Text, pgtypes.Oid, pgtypes.Int32}, // cstring
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		input := val1.(string)
		oid := val2.(uint32) // TODO: what is this for?
		typmod := val3.(int32)
		baseType := pgtypes.OidToBuildInDoltgresType[oid]
		if typmod == -1 {
			return input, nil
		} else {
			input, runeLength := truncateString(input, typmod)
			if runeLength > typmod {
				return input, fmt.Errorf("value too long for type %s", baseType.String())
			} else if runeLength < typmod {
				return input + strings.Repeat(" ", int(typmod-runeLength)), nil
			} else {
				return input, nil
			}
		}
	},
}

// bpcharout represents the PostgreSQL function of bpchar type IO output.
var bpcharout = framework.Function1{
	Name:       "bpcharout",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.BpChar},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO: need length information OR is it expected to be within length limit?
		typ := t[0]
		typLen := int32(typ.Length)
		if typLen == -1 {
			return val.(string), nil
		} else {
			str, runeCount := truncateString(val.(string), typLen)
			if runeCount < typLen {
				return str + strings.Repeat(" ", int(typLen-runeCount)), nil
			}
			return str, nil
		}
	},
}

// bpcharrecv represents the PostgreSQL function of bpchar type IO receive.
var bpcharrecv = framework.Function3{
	Name:       "bpcharrecv",
	Return:     pgtypes.BpChar,
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		// TODO: use typmod
		data := val1.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		reader := utils.NewReader(data)
		return reader.String(), nil
	},
}

// bpcharsend represents the PostgreSQL function of bpchar type IO send.
var bpcharsend = framework.Function1{
	Name:       "bpcharsend",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.BpChar},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		str := val.(string)
		writer := utils.NewWriter(uint64(len(str) + 4))
		writer.String(str)
		return writer.Data(), nil
	},
}

// bpchartypmodin represents the PostgreSQL function of bpchar type IO typmod input.
var bpchartypmodin = framework.Function1{
	Name:       "bpchartypmodin",
	Return:     pgtypes.Int32,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.TextArray}, // cstring[]
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		return nil, nil
	},
}

// bpchartypmodout represents the PostgreSQL function of bpchar type IO typmod output.
var bpchartypmodout = framework.Function1{
	Name:       "bpchartypmodout",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		return nil, nil
	},
}

// bpcharcmp represents the PostgreSQL function of bpchar type compare.
var bpcharcmp = framework.Function2{
	Name:       "bpcharcmp",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.BpChar, pgtypes.BpChar},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		return int32(bytes.Compare([]byte(val1.(string)), []byte(val2.(string)))), nil
	},
}

// truncateString returns a string that has been truncated to the given length. Uses the rune count rather than the
// byte count. Returns the input string if it's smaller than the length. Also returns the rune count of the string.
func truncateString(val string, runeLimit int32) (string, int32) {
	runeLength := int32(utf8.RuneCountInString(val))
	if runeLength > runeLimit {
		// TODO: figure out if there's a faster way to truncate based on rune count
		startString := val
		for i := int32(0); i < runeLimit; i++ {
			_, size := utf8.DecodeRuneInString(val)
			val = val[size:]
		}
		return startString[:len(startString)-len(val)], runeLength
	}
	return val, runeLength
}
