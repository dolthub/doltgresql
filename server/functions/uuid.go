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

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/uuid"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initUuid registers the functions to the catalog.
func initUuid() {
	framework.RegisterFunction(uuid_in)
	framework.RegisterFunction(uuid_out)
	framework.RegisterFunction(uuid_recv)
	framework.RegisterFunction(uuid_send)
	framework.RegisterFunction(uuid_cmp)
}

// uuid_in represents the PostgreSQL function of uuid type IO input.
var uuid_in = framework.Function1{
	Name:       "uuid_in",
	Return:     pgtypes.Uuid,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return uuid.FromString(val.(string))
	},
}

// uuid_out represents the PostgreSQL function of uuid type IO output.
var uuid_out = framework.Function1{
	Name:       "uuid_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Uuid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return val.(uuid.UUID).String(), nil
	},
}

// uuid_recv represents the PostgreSQL function of uuid type IO receive.
var uuid_recv = framework.Function1{
	Name:       "uuid_recv",
	Return:     pgtypes.Uuid,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		return uuid.FromBytes(data)
	},
}

// uuid_send represents the PostgreSQL function of uuid type IO send.
var uuid_send = framework.Function1{
	Name:       "uuid_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Uuid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return val.(uuid.UUID).GetBytes(), nil
	},
}

// uuid_cmp represents the PostgreSQL function of uuid type compare.
var uuid_cmp = framework.Function2{
	Name:       "uuid_cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Uuid, pgtypes.Uuid},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(uuid.UUID)
		bb := val2.(uuid.UUID)
		return int32(bytes.Compare(ab.GetBytesMut(), bb.GetBytesMut())), nil
	},
}
