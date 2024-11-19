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
	"encoding/binary"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initRegproc registers the functions to the catalog.
func initRegproc() {
	framework.RegisterFunction(regprocin)
	framework.RegisterFunction(regprocout)
	framework.RegisterFunction(regprocrecv)
	framework.RegisterFunction(regprocsend)
}

// regprocin represents the PostgreSQL function of regproc type IO input.
var regprocin = framework.Function1{
	Name:       "regprocin",
	Return:     pgtypes.Regproc,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return pgtypes.Regproc_IoInput(ctx, val.(string))
	},
}

// regprocout represents the PostgreSQL function of regproc type IO output.
var regprocout = framework.Function1{
	Name:       "regprocout",
	Return:     pgtypes.Cstring,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Regproc},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return pgtypes.Regproc_IoOutput(ctx, val.(uint32))
	},
}

// regprocrecv represents the PostgreSQL function of regproc type IO receive.
var regprocrecv = framework.Function1{
	Name:       "regprocrecv",
	Return:     pgtypes.Regproc,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		return binary.BigEndian.Uint32(data), nil
	},
}

// regprocsend represents the PostgreSQL function of regproc type IO send.
var regprocsend = framework.Function1{
	Name:       "regprocsend",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Regproc},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		retVal := make([]byte, 4)
		binary.BigEndian.PutUint32(retVal, val.(uint32))
		return retVal, nil
	},
}
