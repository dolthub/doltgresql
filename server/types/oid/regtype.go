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

	"github.com/dolthub/doltgresql/postgres/parser/types"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/lib/pq/oid"
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
	var relationName string
	switch len(sections) {
	case 1:
		relationName = sections[0]
	case 3:
		relationName = sections[2]
	default:
		return 0, fmt.Errorf("regtype failed validation")
	}
	relationName = normalizeTypeRelationName(relationName)

	// Iterate over all of the items to find which relation matches.
	// Postgres does not need to worry about name conflicts since everything is created in the same naming space, but
	// GMS and Dolt use different naming spaces, so for now we just ignore potential name conflicts and return the first
	// match found.
	resultOid := uint32(0)
	err = IterateTypes(ctx, func(typ pgtypes.DoltgresType) (cont bool, err error) {
		stringNoSpace := removeSpaces(typ.String())
		standardName := removeSpaces(getStandardNameFromOid(typ.OID()))
		if relationName == stringNoSpace || relationName == typ.BaseName() || relationName == standardName {
			resultOid = typ.OID()
			return false, nil
		}
		return true, nil
	})
	if err != nil || resultOid != 0 {
		return resultOid, err
	}
	return 0, fmt.Errorf(`type "%s" does not exist`, input)
}

// regtype_IoOutput is the implementation for IoOutput that avoids circular dependencies by being declared in a separate
// package.
func regtype_IoOutput(ctx *sql.Context, toid uint32) (string, error) {
	name := getStandardNameFromOid(toid)
	if name == "" {
		return strconv.FormatUint(uint64(toid), 10), nil
	}
	return name, nil
}

// regtype_IoInputValidation handles the validation for the parsed sections in regtype_IoInput.
func regtype_IoInputValidation(ctx *sql.Context, input string, sections []string) error {
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

// normalizeTypeRelationName removes everything after the first parenthesis from
// the relation name.
func normalizeTypeRelationName(name string) string {
	split := strings.Split(name, "(")
	return split[0]
}

// getStandardNameFromOid returns the SQL standard name of an OID if it exists.
func getStandardNameFromOid(toid uint32) string {
	if t, ok := types.OidToType[oid.Oid(toid)]; ok {
		return t.SQLStandardName()
	}
	return ""
}

// removeSpaces removes all spaces from a string.
func removeSpaces(s string) string {
	return strings.ReplaceAll(s, " ", "")
}
