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
	"slices"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
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
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		// If the string just represents a number, then we return it.
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
		schemas, funcName, err := regproc_SchemasAndName(ctx, input)
		if err != nil {
			return id.Null, err
		}
		// TODO: handle aggregate functions and window functions
		routineIDs, err := regproc_FindRoutines(ctx, schemas, funcName)
		if err != nil {
			return id.Null, err
		}
		switch len(routineIDs) {
		case 0:
			return id.Null, pgerror.Newf(pgcode.UndefinedFunction, `function "%s" does not exist`, input)
		case 1:
			return routineIDs[0], nil
		default:
			return id.Null, pgerror.Newf(pgcode.AmbiguousFunction, `more than one function named "%s"`, input)
		}
	},
}

// regprocout represents the PostgreSQL function of regproc type IO output.
var regprocout = framework.Function1{
	Name:       "regprocout",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Regproc},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		input := val.(id.Id)
		if input.Section() == id.Section_OID {
			return input.Segment(0), nil
		}
		res := val.(id.Id).Segment(1)
		if res == "" {
			return "-", nil
		}
		searchPath, err := core.SearchPath(ctx)
		if err != nil {
			return "", err
		}
		routineIDs, err := regproc_FindRoutines(ctx, searchPath, res)
		if err != nil {
			return "", err
		}
		if regproc_IsVisible(input, searchPath, routineIDs) {
			return res, nil
		}
		return fmt.Sprintf("%s.%s", input.Segment(0), res), nil
	},
}

// regprocrecv represents the PostgreSQL function of regproc type IO receive.
var regprocrecv = framework.Function1{
	Name:       "regprocrecv",
	Return:     pgtypes.Regproc,
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

// regprocsend represents the PostgreSQL function of regproc type IO send.
var regprocsend = framework.Function1{
	Name:       "regprocsend",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Regproc},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		writer := utils.NewWireWriter()
		writer.WriteUint32(id.Cache().ToOID(val.(id.Id)))
		return writer.BufferData(), nil
	},
}

// regproc_IoInputValidation handles the validation for the parsed sections in regproc_IoInput.
func regproc_IoInputValidation(ctx *sql.Context, input string, sections []string) error {
	switch len(sections) {
	case 1:
		return nil
	case 3:
		if sections[1] != "." {
			return errors.Errorf("invalid name syntax")
		}
		return nil
	case 5:
		if sections[1] != "." || sections[3] != "." {
			return errors.Errorf("invalid name syntax")
		}
		return errors.Errorf("cross-database references are not implemented: %s", input)
	case 7:
		if sections[1] != "." || sections[3] != "." || sections[5] != "." {
			return errors.Errorf("invalid name syntax")
		}
		return errors.Errorf("improper qualified name (too many dotted names): %s", input)
	default:
		return errors.Errorf("invalid name syntax")
	}
}

// regproc_SchemasAndName parses the possibly-qualified routine name in `input`, returning the schemas to search in
// order along with the unqualified name.
func regproc_SchemasAndName(ctx *sql.Context, input string) (schemas []string, name string, err error) {
	sections, err := ioInputSections(input)
	if err != nil {
		return nil, "", err
	}
	if err = regproc_IoInputValidation(ctx, input, sections); err != nil {
		return nil, "", err
	}
	switch len(sections) {
	case 1:
		schemas, err = core.SearchPath(ctx)
		return schemas, sections[0], err
	case 3:
		return []string{sections[0]}, sections[2], nil
	default:
		return nil, "", errors.Errorf("regproc failed validation")
	}
}

// regproc_FindRoutines returns the IDs of every built-in function, user-defined function, and procedure with the given
// name in the given schemas.
func regproc_FindRoutines(ctx *sql.Context, schemas []string, name string) ([]id.Id, error) {
	funcCollection, err := core.GetFunctionsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	procCollection, err := core.GetProceduresCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	var routineIDs []id.Id
	for _, schema := range schemas {
		if schema == "pg_catalog" {
			for _, f := range framework.Catalog[name] {
				routineIDs = append(routineIDs, f.InternalID())
			}
		}
		funcs, err := funcCollection.GetFunctionOverloads(ctx, id.NewFunction(schema, name))
		if err != nil {
			return nil, err
		}
		for _, f := range funcs {
			routineIDs = append(routineIDs, f.ID.AsId())
		}
		procs, err := procCollection.GetProcedureOverloads(ctx, id.NewProcedure(schema, name))
		if err != nil {
			return nil, err
		}
		for _, p := range procs {
			routineIDs = append(routineIDs, p.ID.AsId())
		}
	}
	return routineIDs, nil
}

// regproc_IsVisible returns whether `routineID` may be displayed without its schema. Its schema must be on the search
// path, and `resolvedIDs`, the routines found by its unqualified name, must contain no other routine.
func regproc_IsVisible(routineID id.Id, searchPath []string, resolvedIDs []id.Id) bool {
	for _, resolvedID := range resolvedIDs {
		if resolvedID.IsValid() && resolvedID != routineID {
			return false
		}
	}
	return slices.Contains(searchPath, routineID.Segment(0))
}
