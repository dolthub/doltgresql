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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/types"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initRegtype registers the functions to the catalog.
func initRegtype() {
	framework.RegisterFunction(regtypein)
	framework.RegisterFunction(regtypeout)
	framework.RegisterFunction(regtyperecv)
	framework.RegisterFunction(regtypesend)
}

// regtypein represents the PostgreSQL function of regtype type IO input.
var regtypein = framework.Function1{
	Name:       "regtypein",
	Return:     pgtypes.Regtype,
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
		if err = regtype_IoInputValidation(ctx, input, sections); err != nil {
			return id.Null, err
		}
		var schema string
		var typeName string
		switch len(sections) {
		case 1:
			// TODO: we should make use of the search path, but it needs to include an implicit "pg_catalog" before we can
			typeName = sections[0]
		case 3:
			// TODO: sections[0] is the schema that we need to search in
			schema = sections[0]
			typeName = sections[2]
			if schema == "pg_catalog" && typeName == "char" { // Sad but true
				typeName = `"char"`
			}
		default:
			return id.Null, fmt.Errorf("regtype failed validation")
		}
		// Remove everything after the first parenthesis
		typeName = strings.Split(typeName, "(")[0]

		if typeName == "char" && schema == "" {
			return id.NewInternalType("pg_catalog", "bpchar").Internal(), nil
		}
		if internalID, ok := pgtypes.NameToInternalID[typeName]; ok && (internalID.SchemaName() == schema || schema == "") {
			return internalID.Internal(), nil
		}
		return id.Null, pgtypes.ErrTypeDoesNotExist.New(input)
	},
}

// regtypeout represents the PostgreSQL function of regtype type IO output.
var regtypeout = framework.Function1{
	Name:       "regtypeout",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Regtype},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		internalID := val.(id.Internal)
		if internalID.Section() == id.Section_OID {
			return internalID.Segment(0), nil
		}
		toid := id.Cache().ToOID(internalID)
		if t, ok := types.OidToType[oid.Oid(toid)]; ok {
			return t.SQLStandardName(), nil
		} else {
			return internalID.Segment(1), nil
		}
	},
}

// regtyperecv represents the PostgreSQL function of regtype type IO receive.
var regtyperecv = framework.Function1{
	Name:       "regtyperecv",
	Return:     pgtypes.Regtype,
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

// regtypesend represents the PostgreSQL function of regtype type IO send.
var regtypesend = framework.Function1{
	Name:       "regtypesend",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Regtype},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return []byte(val.(id.Internal)), nil
	},
}

// regtype_IoInputValidation handles the validation for the parsed sections in regtypein.
func regtype_IoInputValidation(ctx *sql.Context, input string, sections []string) error {
	switch len(sections) {
	case 1:
		return nil
	case 3:
		// We check for name validity before checking the schema name
		if sections[1] != "." {
			return fmt.Errorf("invalid name syntax")
		}
		return nil
	case 5:
		if sections[1] != "." || sections[3] != "." {
			return fmt.Errorf("invalid name syntax")
		}
		return fmt.Errorf("cross-database references are not implemented: %s", input)
	case 7:
		if sections[1] != "." || sections[3] != "." || sections[5] != "." {
			return fmt.Errorf("invalid name syntax")
		}
		return fmt.Errorf("improper qualified name (too many dotted names): %s", input)
	default:
		return fmt.Errorf("invalid name syntax")
	}
}
