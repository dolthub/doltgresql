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
	"unicode"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// initRegprocedure registers the functions to the catalog.
func initRegprocedure() {
	framework.RegisterFunction(regprocedurein)
	framework.RegisterFunction(regprocedureout)
	framework.RegisterFunction(regprocedurerecv)
	framework.RegisterFunction(regproceduresend)
}

// regprocedurein represents the PostgreSQL function of regprocedure type IO input.
var regprocedurein = framework.Function1{
	Name:       "regprocedurein",
	Return:     pgtypes.Regprocedure,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		input, err := framework.UnwrapString(ctx, val)
		if err != nil {
			return nil, err
		}
		// If the string just represents a number, then we return it.
		if parsedOid, err := strconv.ParseUint(input, 10, 32); err == nil {
			if internalID := id.Cache().ToInternal(uint32(parsedOid)); internalID.IsValid() {
				return internalID, nil
			}
			return id.NewOID(uint32(parsedOid)).AsId(), nil
		}
		return regprocedure_Resolve(ctx, input)
	},
}

// regprocedureout represents the PostgreSQL function of regprocedure type IO output.
var regprocedureout = framework.Function1{
	Name:       "regprocedureout",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Regprocedure},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		input := val.(id.Id)
		if input.Section() == id.Section_OID {
			return input.Segment(0), nil
		}
		name := input.Segment(1)
		if name == "" {
			return "-", nil
		}
		paramTypes := make([]id.Type, input.SegmentCount()-2)
		typeNames := make([]string, len(paramTypes))
		for i, paramType := range input.Data()[2:] {
			paramTypes[i] = id.Type(paramType)
			typeName, err := pgtypes.Regtype.IoOutput(ctx, paramTypes[i].AsId())
			if err != nil {
				return "", err
			}
			typeNames[i] = typeName
		}
		output := fmt.Sprintf("%s(%s)", name, strings.Join(typeNames, ","))
		searchPath, err := core.SearchPath(ctx)
		if err != nil {
			return "", err
		}
		resolvedID, err := regprocedure_FindRoutine(ctx, searchPath, name, paramTypes)
		if err != nil {
			return "", err
		}
		if regproc_IsVisible(input, searchPath, []id.Id{resolvedID}) {
			return output, nil
		}
		return fmt.Sprintf("%s.%s", input.Segment(0), output), nil
	},
}

// regprocedurerecv represents the PostgreSQL function of regprocedure type IO receive.
var regprocedurerecv = framework.Function1{
	Name:       "regprocedurerecv",
	Return:     pgtypes.Regprocedure,
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

// regproceduresend represents the PostgreSQL function of regprocedure type IO send.
var regproceduresend = framework.Function1{
	Name:       "regproceduresend",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Regprocedure},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		writer := utils.NewWireWriter()
		writer.WriteUint32(id.Cache().ToOID(val.(id.Id)))
		return writer.BufferData(), nil
	},
}

// regprocedure_Resolve returns the ID of the routine named by `input`, such as "public.f(int, text)". Returns an
// UndefinedFunction error when no routine has that name and those argument types.
func regprocedure_Resolve(ctx *sql.Context, input string) (id.Id, error) {
	name, args, err := regprocedure_Split(input)
	if err != nil {
		return id.Null, err
	}
	schemas, funcName, err := regproc_SchemasAndName(ctx, name)
	if err != nil {
		return id.Null, err
	}
	paramTypes := make([]id.Type, len(args))
	for i, arg := range args {
		typeID, err := pgtypes.Regtype.IoInput(ctx, arg)
		if err != nil {
			return id.Null, err
		}
		paramTypes[i] = id.Type(typeID.(id.Id))
	}
	// TODO: handle aggregate functions and window functions
	routineID, err := regprocedure_FindRoutine(ctx, schemas, funcName, paramTypes)
	if err != nil || routineID.IsValid() {
		return routineID, err
	}
	return id.Null, pgerror.Newf(pgcode.UndefinedFunction, `function "%s" does not exist`, input)
}

// regprocedure_Split separates `input` into the routine name and its argument types, so "f(int, text)" becomes "f"
// with the arguments "int" and "text".
func regprocedure_Split(input string) (name string, args []string, err error) {
	leftParen := -1
	inQuotes := false
	for i, char := range input {
		if char == '"' {
			inQuotes = !inQuotes
		} else if char == '(' && !inQuotes {
			leftParen = i
			break
		}
	}
	if leftParen == -1 {
		return "", nil, pgerror.New(pgcode.InvalidTextRepresentation, "expected a left parenthesis")
	}
	argList, ok := strings.CutSuffix(strings.TrimRightFunc(input[leftParen+1:], unicode.IsSpace), ")")
	if !ok {
		return "", nil, pgerror.New(pgcode.InvalidTextRepresentation, "expected a right parenthesis")
	}
	if len(strings.TrimSpace(argList)) == 0 {
		return input[:leftParen], nil, nil
	}
	depth := 0
	argStart := 0
	inQuotes = false
	for i, char := range argList {
		if char == '"' {
			inQuotes = !inQuotes
		} else if !inQuotes {
			switch char {
			case '(', '[':
				depth++
			case ')', ']':
				depth--
			case ',':
				if depth == 0 {
					args = append(args, strings.TrimSpace(argList[argStart:i]))
					argStart = i + 1
				}
			}
		}
	}
	if inQuotes || depth != 0 {
		return "", nil, pgerror.New(pgcode.InvalidTextRepresentation, "improper type name")
	}
	lastArg := strings.TrimSpace(argList[argStart:])
	if len(lastArg) == 0 {
		return "", nil, pgerror.New(pgcode.InvalidTextRepresentation, "expected a type name")
	}
	return input[:leftParen], append(args, lastArg), nil
}

// regprocedure_FindRoutine returns the ID of the first built-in function, user-defined function, or procedure with the
// given name and parameter types in the given schemas. Returns `id.Null` if none are found.
func regprocedure_FindRoutine(ctx *sql.Context, schemas []string, name string, paramTypes []id.Type) (id.Id, error) {
	funcCollection, err := core.GetFunctionsCollectionFromContext(ctx, "")
	if err != nil {
		return id.Null, err
	}
	procCollection, err := core.GetProceduresCollectionFromContext(ctx, "")
	if err != nil {
		return id.Null, err
	}
	for _, schema := range schemas {
		funcID := id.NewFunction(schema, name, paramTypes...)
		if schema == "pg_catalog" {
			for _, f := range framework.Catalog[name] {
				if f.InternalID() == funcID.AsId() {
					return funcID.AsId(), nil
				}
			}
		}
		if funcCollection.HasFunction(ctx, funcID) {
			return funcID.AsId(), nil
		}
		if procID := id.NewProcedure(schema, name, paramTypes...); procCollection.HasProcedure(ctx, procID) {
			return procID.AsId(), nil
		}
	}
	return id.Null, nil
}
