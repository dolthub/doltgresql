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
	"strconv"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// initCid registers the functions to the catalog.
func initCid() {
	framework.RegisterFunction(cidin)
	framework.RegisterFunction(cidout)
	framework.RegisterFunction(cidrecv)
	framework.RegisterFunction(cidsend)
}

// cidin represents the PostgreSQL function of cid type IO input.
var cidin = framework.Function1{
	Name:       "cidin",
	Return:     pgtypes.Cid,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		input := val.(string)
		uVal, err := strconv.ParseUint(strings.TrimSpace(input), 10, 32)
		if err != nil {
			return uint32(0), nil
		}
		return uint32(uVal), nil
	},
}

// cidout represents the PostgreSQL function of cid type IO output.
var cidout = framework.Function1{
	Name:       "cidout",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return strconv.FormatUint(uint64(val.(uint32)), 10), nil
	},
}

// cidrecv represents the PostgreSQL function of cid type IO receive.
var cidrecv = framework.Function1{
	Name:       "cidrecv",
	Return:     pgtypes.Cid,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		if data == nil {
			return nil, nil
		}
		reader := utils.NewWireReader(data)
		return reader.ReadUint32(), nil
	},
}

// cidsend represents the PostgreSQL function of cid type IO send.
var cidsend = framework.Function1{
	Name:       "cidsend",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		writer := utils.NewWireWriter()
		writer.WriteUint32(val.(uint32))
		return writer.BufferData(), nil
	},
}
