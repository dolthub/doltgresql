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

package oid

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/postgres/parser/types"
)

// regtype_IoInput is the implementation for IoInput that avoids circular dependencies by being declared in a separate
// package.
func regtype_IoInput(ctx *sql.Context, input string) (uint32, error) {
	// If the string just represents a number, then we return it.
	if parsedOid, err := strconv.ParseUint(input, 10, 32); err == nil {
		return uint32(parsedOid), nil
	}
	sections, err := ioInputSections(input)
	if err != nil {
		return 0, err
	}
	if err = regtype_IoInputValidation(ctx, input, sections); err != nil {
		return 0, err
	}
	var searchSchemas []string
	var typeName string
	switch len(sections) {
	case 1:
		// TODO: we should make use of the search path, but it needs to include an implicit "pg_catalog" before we can
		typeName = sections[0]
	case 3:
		searchSchemas = []string{sections[0]}
		typeName = sections[2]
	default:
		return 0, fmt.Errorf("regtype failed validation")
	}
	// Remove everything after the first parenthesis
	typeName = strings.Split(typeName, "(")[0]

	resultOid := uint32(0)
	err = IterateCurrentDatabase(ctx, Callbacks{
		Type: func(ctx *sql.Context, typ ItemType) (cont bool, err error) {
			if typeName == typ.Item.String() || typeName == typ.Item.BaseName() || (typeName == "char" && typ.Item.BaseName() == "bpchar") {
				resultOid = typ.OID
				return false, nil
			} else if t, ok := types.OidToType[oid.Oid(typ.OID)]; ok && typeName == t.SQLStandardName() {
				resultOid = typ.OID
				return false, nil
			}
			return true, nil
		},
		SearchSchemas: searchSchemas,
	})
	if err != nil || resultOid != 0 {
		return resultOid, err
	}
	return 0, fmt.Errorf(`type "%s" does not exist`, input)
}

// regtype_IoOutput is the implementation for IoOutput that avoids circular dependencies by being declared in a separate
// package.
func regtype_IoOutput(ctx *sql.Context, toid uint32) (string, error) {
	output := strconv.FormatUint(uint64(toid), 10)
	err := RunCallback(ctx, toid, Callbacks{
		Type: func(ctx *sql.Context, typ ItemType) (cont bool, err error) {
			if t, ok := types.OidToType[oid.Oid(toid)]; ok {
				output = t.SQLStandardName()
			}
			return false, nil
		},
	})
	return output, err
}

// regtype_IoInputValidation handles the validation for the parsed sections in regtype_IoInput.
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
