// Copyright 2025 Dolthub, Inc.
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

package analyzer

import (
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"

	"github.com/dolthub/doltgresql/server/functions/framework"
	"github.com/dolthub/doltgresql/server/types"
)

// validateForeignKeyDefinition validates that the given foreign key definition is valid for creation
func validateForeignKeyDefinition(ctx *sql.Context, fkDef sql.ForeignKeyConstraint, cols map[string]*sql.Column, parentCols map[string]*sql.Column) error {
	for i := range fkDef.Columns {
		col := cols[strings.ToLower(fkDef.Columns[i])]
		parentCol := parentCols[strings.ToLower(fkDef.ParentColumns[i])]
		if !foreignKeyComparableTypes(col.Type, parentCol.Type) {
			return errors.Errorf("Key columns %q and %q are of incompatible types: %s and %s", col.Name, parentCol.Name, col.Type.String(), parentCol.Type.String())
		}
	}
	return nil
}

// foreignKeyComparableTypes returns whether the two given types are able to be used as parent/child columns in a
// foreign key.
func foreignKeyComparableTypes(from sql.Type, to sql.Type) bool {
	dtFrom, ok := from.(*types.DoltgresType)
	if !ok {
		return false // should never be possible
	}

	dtTo, ok := to.(*types.DoltgresType)
	if !ok {
		return false // should never be possible
	}

	if dtFrom.Equals(dtTo) {
		return true
	}

	fromLiteral := expression.NewLiteral(dtFrom.Zero(), from)
	toLiteral := expression.NewLiteral(dtTo.Zero(), to)

	// a foreign key between two different types is valid if there is an equality operator on the two types
	// TODO: there are some subtleties in postgres not captured by this logic, e.g. a foreign key from double -> int
	//  is valid, but the reverse is not. This works fine, but is more permissive than postgres is.
	eq := framework.GetBinaryFunction(framework.Operator_BinaryEqual).Compile("=", fromLiteral, toLiteral)
	if eq == nil || eq.StashedError() != nil {
		return false
	}

	// Additionally, we need to be able to convert freely between the two types in both directions, since we do this
	// during the process of enforcing the constraints
	forwardConversion := types.GetAssignmentCast(dtFrom, dtTo)
	reverseConversion := types.GetAssignmentCast(dtTo, dtFrom)

	return forwardConversion != nil && reverseConversion != nil
}
