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
	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// initRegrole registers the functions to the catalog.
func initRegrole() {
	framework.RegisterFunction(regrolein)
	framework.RegisterFunction(regroleout)
	framework.RegisterFunction(regrolerecv)
	framework.RegisterFunction(regrolesend)
}

// regrolein represents the PostgreSQL function of regrole type IO input.
var regrolein = framework.Function1{
	Name:       "regrolein",
	Return:     pgtypes.Regrole,
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

		var exists bool
		auth.LockRead(func() {
			exists = auth.RoleExists(sections[0])
		})
		if !exists {
			return id.Null, errors.Errorf(`role "%s" does not exist`, sections[0])
		}
		// This matches the OID used by pg_authid and related tables
		return id.NewId(id.Section_User, sections[0]), nil
	},
}

// regroleout represents the PostgreSQL function of regrole type IO output.
var regroleout = framework.Function1{
	Name:       "regroleout",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Regrole},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		input := val.(id.Id)
		if input.Section() == id.Section_User {
			roleName := input.Segment(0)
			var exists bool
			auth.LockRead(func() {
				exists = auth.RoleExists(roleName)
			})
			if exists {
				return roleName, nil
			}
		} else if input.Section() == id.Section_OID {
			return input.Segment(0), nil
		}
		// Roles that do not exist are displayed using their numeric OID
		return strconv.FormatUint(uint64(id.Cache().ToOID(input)), 10), nil
	},
}

// regrolerecv represents the PostgreSQL function of regrole type IO receive.
var regrolerecv = framework.Function1{
	Name:       "regrolerecv",
	Return:     pgtypes.Regrole,
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

// regrolesend represents the PostgreSQL function of regrole type IO send.
var regrolesend = framework.Function1{
	Name:       "regrolesend",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Regrole},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		writer := utils.NewWireWriter()
		writer.WriteUint32(id.Cache().ToOID(val.(id.Id)))
		return writer.BufferData(), nil
	},
}
