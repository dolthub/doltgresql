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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// initRegnamespace registers the functions to the catalog.
func initRegnamespace() {
	framework.RegisterFunction(regnamespacein)
	framework.RegisterFunction(regnamespaceout)
	framework.RegisterFunction(regnamespacerecv)
	framework.RegisterFunction(regnamespacesend)
}

// regnamespacein represents the PostgreSQL function of regnamespace type IO input.
var regnamespacein = framework.Function1{
	Name:       "regnamespacein",
	Return:     pgtypes.Regnamespace,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		input, err := framework.UnwrapString(ctx, val)
		if err != nil {
			return nil, err
		}
		if parsedOid, err := strconv.ParseUint(input, 10, 32); err == nil {
			if internalID := id.Cache().ToInternal(uint32(parsedOid)); internalID.IsValid() {
				return internalID, nil
			}
			return id.NewOID(uint32(parsedOid)).AsId(), nil
		}
		sections, err := ioInputSections(input)
		if err != nil {
			return id.Null, err
		}
		if len(sections) != 1 {
			return id.Null, errors.Errorf("invalid name syntax")
		}

		var resultOid id.Id
		err = IterateCurrentDatabase(ctx, Callbacks{
			Schema: func(ctx *sql.Context, schema ItemSchema) (cont bool, err error) {
				if schema.Item.SchemaName() == sections[0] {
					resultOid = schema.OID.AsId()
					return false, nil
				}
				return true, nil
			},
		})
		if err != nil || resultOid.IsValid() {
			return resultOid, err
		}
		return id.Null, errors.Errorf(`schema "%s" does not exist`, sections[0])
	},
}

// regnamespaceout represents the PostgreSQL function of regnamespace type IO output.
var regnamespaceout = framework.Function1{
	Name:       "regnamespaceout",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Regnamespace},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		input := val.(id.Id)
		if input.Section() == id.Section_OID {
			return input.Segment(0), nil
		}
		return id.Namespace(input).SchemaName(), nil
	},
}

// regnamespacerecv represents the PostgreSQL function of regnamespace type IO receive.
var regnamespacerecv = framework.Function1{
	Name:       "regnamespacerecv",
	Return:     pgtypes.Regnamespace,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		data, err := framework.UnwrapBytes(ctx, val)
		if err != nil {
			return nil, err
		}
		if data == nil {
			return nil, nil
		}
		reader := utils.NewWireReader(data)
		cachedID := id.Cache().ToInternal(reader.ReadUint32())
		return cachedID, nil
	},
}

// regnamespacesend represents the PostgreSQL function of regnamespace type IO send.
var regnamespacesend = framework.Function1{
	Name:       "regnamespacesend",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Regnamespace},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		writer := utils.NewWireWriter()
		writer.WriteUint32(id.Cache().ToOID(val.(id.Id)))
		return writer.BufferData(), nil
	},
}
