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
	"fmt"
	"strconv"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// initTid registers the functions to the catalog.
func initTid() {
	framework.RegisterFunction(tidin)
	framework.RegisterFunction(tidout)
	framework.RegisterFunction(tidrecv)
	framework.RegisterFunction(tidsend)
	framework.RegisterFunction(bttidcmp)
	framework.RegisterFunction(tid_block)
	framework.RegisterFunction(tid_offset)
}

// tidin represents the PostgreSQL function of tid type IO input.
var tidin = framework.Function1{
	Name:       "tidin",
	Return:     pgtypes.Tid,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		valStr, err := framework.UnwrapString(ctx, val)
		if err != nil {
			return nil, err
		}
		input := strings.TrimSpace(valStr)
		commaIdx := strings.Index(input, ",")
		if len(input) < 5 || input[0] != '(' || input[len(input)-1] != ')' || commaIdx == -1 {
			return nil, pgtypes.ErrInvalidSyntaxForType.New(pgtypes.Tid.Name(), input)
		}
		block, err := strconv.ParseUint(input[1:commaIdx], 10, 32)
		if err != nil {
			return nil, pgtypes.ErrInvalidSyntaxForType.New(pgtypes.Tid.Name(), input)
		}
		offset, err := strconv.ParseUint(input[commaIdx+1:len(input)-1], 10, 16)
		if err != nil {
			return nil, pgtypes.ErrInvalidSyntaxForType.New(pgtypes.Tid.Name(), input)
		}
		return pgtypes.TidValue{
			Block:  uint32(block),
			Offset: uint16(offset),
		}, nil
	},
}

// tidout represents the PostgreSQL function of tid type IO output.
var tidout = framework.Function1{
	Name:       "tidout",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Tid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		tidValue := val.(pgtypes.TidValue)
		return fmt.Sprintf("(%d,%d)", tidValue.Block, tidValue.Offset), nil
	},
}

// tidrecv represents the PostgreSQL function of tid type IO receive.
var tidrecv = framework.Function1{
	Name:       "tidrecv",
	Return:     pgtypes.Tid,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		if data == nil {
			return nil, nil
		}
		reader := utils.NewWireReader(data)
		block := reader.ReadUint32()
		offset := reader.ReadUint16()
		return pgtypes.TidValue{
			Block:  block,
			Offset: offset,
		}, nil
	},
}

// tidsend represents the PostgreSQL function of tid type IO send.
var tidsend = framework.Function1{
	Name:       "tidsend",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Tid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		tidValue := val.(pgtypes.TidValue)
		writer := utils.NewWireWriter()
		writer.WriteUint32(tidValue.Block)
		writer.WriteUint16(tidValue.Offset)
		return writer.BufferData(), nil
	},
}

// bttidcmp represents the PostgreSQL function of tid type compare.
var bttidcmp = framework.Function2{
	Name:       "bttidcmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Tid, pgtypes.Tid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		res, err := pgtypes.Tid.Compare(ctx, val1, val2)
		return int32(res), err
	},
}

// tid_block represents the PostgreSQL function of tid type IO output.
var tid_block = framework.Function1{
	Name:       "tid_block",
	Return:     pgtypes.Int64,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Tid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return int64(val.(pgtypes.TidValue).Block), nil
	},
}

// tid_offset represents the PostgreSQL function of tid type IO output.
var tid_offset = framework.Function1{
	Name:       "tid_offset",
	Return:     pgtypes.Int64,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Tid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return int64(val.(pgtypes.TidValue).Offset), nil
	},
}
