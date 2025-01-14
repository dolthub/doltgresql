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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/utils"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initRecord registers the functions to the catalog.
func initRecord() {
	framework.RegisterFunction(record_in)
	framework.RegisterFunction(record_out)
	framework.RegisterFunction(record_recv)
	framework.RegisterFunction(record_send)
	framework.RegisterFunction(btrecordcmp)
	framework.RegisterFunction(btrecordimagecmp)
}

// record_in represents the PostgreSQL function of record type IO input.
var record_in = framework.Function3{
	Name:       "record_in",
	Return:     pgtypes.Record,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Cstring, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		//typOid := val2.(id.Id)
		return val1.(string), nil
	},
}

// record_out represents the PostgreSQL function of record type IO output.
var record_out = framework.Function1{
	Name:       "record_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Record},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		return val.(string), nil
	},
}

// record_recv represents the PostgreSQL function of record type IO receive.
var record_recv = framework.Function3{
	Name:       "record_recv",
	Return:     pgtypes.Record,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		// TODO
		// typOid := val2.(id.Id)
		data := val1.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		reader := utils.NewReader(data)
		return reader.String(), nil
	},
}

// record_send represents the PostgreSQL function of record type IO send.
var record_send = framework.Function1{
	Name:       "record_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Record},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		str := val.(string)
		writer := utils.NewWriter(uint64(len(str) + 4))
		writer.String(str)
		return writer.Data(), nil
	},
}

// btrecordcmp represents the PostgreSQL function of record type compare.
var btrecordcmp = framework.Function2{
	Name:       "btrecordcmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Record, pgtypes.Record},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		// TODO
		ab := val1.(string)
		bb := val2.(string)
		if ab == bb {
			return int32(0), nil
		} else if ab < bb {
			return int32(-1), nil
		} else {
			return int32(1), nil
		}
	},
}

// btrecordimagecmp represents the PostgreSQL function of record type compare.
var btrecordimagecmp = framework.Function2{
	Name:       "btrecordimagecmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Record, pgtypes.Record},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		// TODO
		return int32(1), nil
	},
}
