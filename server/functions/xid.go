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
	"strconv"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initXid registers the functions to the catalog.
func initXid() {
	framework.RegisterFunction(xidin)
	framework.RegisterFunction(xidout)
	framework.RegisterFunction(xidrecv)
	framework.RegisterFunction(xidsend)
}

// xidin represents the PostgreSQL function of xid type IO input.
var xidin = framework.Function1{
	Name:       "xidin",
	Return:     pgtypes.Xid,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Text}, // cstring
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		input := val.(string)
		uVal, err := strconv.ParseInt(strings.TrimSpace(input), 10, 64)
		if err != nil {
			return uint32(0), nil
		}
		return uint32(uVal), nil
	},
}

// xidout represents the PostgreSQL function of xid type IO output.
var xidout = framework.Function1{
	Name:       "xidout",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Xid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return strconv.FormatUint(uint64(val.(uint32)), 10), nil
	},
}

// xidrecv represents the PostgreSQL function of xid type IO receive.
var xidrecv = framework.Function1{
	Name:       "xidrecv",
	Return:     pgtypes.Xid,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		switch val := val.(type) {
		case uint32:
			return val, nil
		default:
			return nil, pgtypes.ErrUnhandledType.New("xid", val)
		}
	},
}

// xidsend represents the PostgreSQL function of xid type IO send.
var xidsend = framework.Function1{
	Name:       "xidsend",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Xid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return []byte(strconv.FormatUint(uint64(val.(uint32)), 10)), nil
	},
}
