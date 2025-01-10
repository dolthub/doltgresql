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
	"strconv"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/resolve"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions/framework"
	"github.com/dolthub/doltgresql/server/settings"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initRegclass registers the functions to the catalog.
func initRegclass() {
	framework.RegisterFunction(regclassin)
	framework.RegisterFunction(regclassout)
	framework.RegisterFunction(regclassrecv)
	framework.RegisterFunction(regclasssend)
}

// regclassin represents the PostgreSQL function of regclass type IO input.
var regclassin = framework.Function1{
	Name:       "regclassin",
	Return:     pgtypes.Regclass,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		// If the string just represents a number, then we return it.
		input := val.(string)
		if parsedOid, err := strconv.ParseUint(input, 10, 32); err == nil {
			if internalID := id.Cache().ToInternal(uint32(parsedOid)); internalID.IsValid() {
				return internalID, nil
			}
			return id.NewInternalOID(uint32(parsedOid)).Internal(), nil
		}
		sections, err := ioInputSections(input)
		if err != nil {
			return id.Null, err
		}
		if err = regclass_IoInputValidation(ctx, input, sections); err != nil {
			return id.Null, err
		}

		var database string
		var searchSchemas []string
		var relationName string
		switch len(sections) {
		case 1:
			database = ctx.GetCurrentDatabase()
			searchSchemas, err = resolve.SearchPath(ctx)
			if err != nil {
				return id.Null, err
			}
			relationName = sections[0]
		case 3:
			database = ctx.GetCurrentDatabase()
			searchSchemas = []string{sections[0]}
			relationName = sections[2]
		case 5:
			database = sections[0]
			searchSchemas = []string{sections[2]}
			relationName = sections[4]
		default:
			return id.Null, fmt.Errorf("regclass failed validation")
		}

		// Iterate over all of the items to find which relation matches.
		// Postgres does not need to worry about name conflicts since everything is created in the same naming space, but
		// GMS and Dolt use different naming spaces, so for now we just ignore potential name conflicts and return the first
		// match found.
		var resultOid id.Internal
		err = IterateDatabase(ctx, database, Callbacks{
			Index: func(ctx *sql.Context, schema ItemSchema, table ItemTable, index ItemIndex) (cont bool, err error) {
				idxName := index.Item.ID()
				if idxName == "PRIMARY" {
					idxName = fmt.Sprintf("%s_pkey", index.Item.Table())
				}
				if relationName == idxName {
					resultOid = index.OID.Internal()
					return false, nil
				}
				return true, nil
			},
			Sequence: func(ctx *sql.Context, schema ItemSchema, sequence ItemSequence) (cont bool, err error) {
				if sequence.Item.Name.SequenceName() == relationName {
					resultOid = sequence.OID.Internal()
					return false, nil
				}
				return true, nil
			},
			Table: func(ctx *sql.Context, schema ItemSchema, table ItemTable) (cont bool, err error) {
				if table.Item.Name() == relationName {
					resultOid = table.OID.Internal()
					return false, nil
				}
				return true, nil
			},
			View: func(ctx *sql.Context, schema ItemSchema, view ItemView) (cont bool, err error) {
				if view.Item.Name == relationName {
					resultOid = view.OID.Internal()
					return false, nil
				}
				return true, nil
			},
			SearchSchemas: searchSchemas,
		})
		if err != nil || resultOid.IsValid() {
			return resultOid, err
		}
		return id.Null, fmt.Errorf(`relation "%s" does not exist`, input)
	},
}

// regclassout represents the PostgreSQL function of regclass type IO output.
var regclassout = framework.Function1{
	Name:       "regclassout",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Regclass},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		// Find all the schemas on the search path. If a schema is on the search path, then it is not included in the
		// output of relation name. If the relation's schema is not on the search path, then it is explicitly included.
		schemasMap, err := settings.GetCurrentSchemasAsMap(ctx)
		if err != nil {
			return "", err
		}

		// The pg_catalog schema is always implicitly part of the search path
		// https://www.postgresql.org/docs/current/ddl-schemas.html#DDL-SCHEMAS-CATALOG
		schemasMap["pg_catalog"] = struct{}{}

		input := val.(id.Internal)
		if input.Section() == id.Section_OID {
			return input.Segment(0), nil
		}
		var output string
		err = RunCallback(ctx, input, Callbacks{
			Index: func(ctx *sql.Context, schema ItemSchema, table ItemTable, index ItemIndex) (cont bool, err error) {
				output = index.Item.ID()
				if output == "PRIMARY" {
					schemaName := schema.Item.SchemaName()
					if _, ok := schemasMap[schemaName]; ok {
						output = fmt.Sprintf("%s_pkey", index.Item.Table())
					} else {
						output = fmt.Sprintf("%s.%s_pkey", schemaName, index.Item.Table())
					}
				}
				return false, nil
			},
			Sequence: func(ctx *sql.Context, schema ItemSchema, sequence ItemSequence) (cont bool, err error) {
				schemaName := schema.Item.SchemaName()
				if _, ok := schemasMap[schemaName]; ok {
					output = sequence.Item.Name.SequenceName()
				} else {
					output = fmt.Sprintf("%s.%s", schemaName, sequence.Item.Name.SequenceName())
				}
				return false, nil
			},
			Table: func(ctx *sql.Context, schema ItemSchema, table ItemTable) (cont bool, err error) {
				schemaName := schema.Item.SchemaName()
				if _, ok := schemasMap[schemaName]; ok {
					output = table.Item.Name()
				} else {
					output = fmt.Sprintf("%s.%s", schemaName, table.Item.Name())
				}
				return false, nil
			},
			View: func(ctx *sql.Context, schema ItemSchema, view ItemView) (cont bool, err error) {
				schemaName := schema.Item.SchemaName()
				if _, ok := schemasMap[schemaName]; ok {
					output = view.Item.Name
				} else {
					output = fmt.Sprintf("%s.%s", schemaName, view.Item.Name)
				}
				return false, nil
			},
		})
		return output, err
	},
}

// regclassrecv represents the PostgreSQL function of regclass type IO receive.
var regclassrecv = framework.Function1{
	Name:       "regclassrecv",
	Return:     pgtypes.Regclass,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		return id.Internal(data), nil
	},
}

// regclasssend represents the PostgreSQL function of regclass type IO send.
var regclasssend = framework.Function1{
	Name:       "regclasssend",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Regclass},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return []byte(val.(id.Internal)), nil
	},
}

// regclass_IoInputValidation handles the validation for the parsed sections in regclass_IoInput.
func regclass_IoInputValidation(ctx *sql.Context, input string, sections []string) error {
	switch len(sections) {
	case 1:
		return nil
	case 3:
		if sections[1] != "." {
			return fmt.Errorf("invalid name syntax")
		}
		return nil
	case 5:
		if sections[1] != "." || sections[3] != "." {
			return fmt.Errorf("invalid name syntax")
		}
		return nil
	case 7:
		if sections[1] != "." || sections[3] != "." || sections[5] != "." {
			return fmt.Errorf("invalid name syntax")
		}
		return fmt.Errorf("improper qualified name (too many dotted names): %s", input)
	default:
		return fmt.Errorf("invalid name syntax")
	}
}
