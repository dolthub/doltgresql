// Copyright 2023 Dolthub, Inc.
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

package ast

import (
	"fmt"

	"github.com/cockroachdb/errors"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgnodes "github.com/dolthub/doltgresql/server/node"
)

// nodeAlterSequence handles *tree.AlterSequence nodes.
func nodeAlterSequence(ctx *Context, node *tree.AlterSequence) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}

	var warnings []string
	if len(node.Owner) > 0 {
		// We intentionally don't support OWNER TO since we don't support owning objects
		if len(node.Options) == 0 {
			return NewNoOp("OWNER TO is unsupported and ignored"), nil
		} else {
			warnings = append(warnings, "OWNER TO is unsupported and ignored")
		}
	}
	if node.SetLog {
		return NotYetSupportedError("LOGGED and UNLOGGED are not yet supported")
	}

	nodeName := node.Name.ToTableName()
	name, err := nodeTableName(ctx, &nodeName)
	if err != nil {
		return nil, err
	}
	if len(name.DbQualifier.String()) > 0 {
		return NotYetSupportedError("ALTER SEQUENCE does not yet support specifying the database")
	}

	ownedBy := pgnodes.AlterSequenceOwnedBy{}
	for _, option := range node.Options {
		switch option.Name {
		case tree.SeqOptOwnedBy:
			ownedBy.IsSet = true
			// OWNED BY NONE is valid, so we have to check if a column was provided
			if option.ColumnItemVal != nil {
				expr, err := nodeExpr(ctx, option.ColumnItemVal)
				if err != nil {
					return nil, err
				}
				colName, ok := expr.(*vitess.ColName)
				if !ok {
					return nil, errors.New("expected sequence owner to be a table and column name")
				}
				if colName.Qualifier.SchemaQualifier.String() != name.SchemaQualifier.String() {
					return nil, errors.New("ALTER SEQUENCE must use the same schema for the sequence and owned table")
				}
				if len(colName.Qualifier.DbQualifier.String()) > 0 {
					return nil, errors.New("database specification is not yet supported for sequences")
				}
				if len(colName.Name.String()) == 0 || len(colName.Qualifier.Name.String()) == 0 {
					return nil, errors.New("invalid OWNED BY option")
				}
				ownedBy.Table = colName.Qualifier.Name.String()
				ownedBy.Column = colName.Name.String()
			}
		default:
			return NotYetSupportedError(fmt.Sprintf("%s is not yet supported", option.Name))
		}
	}

	return vitess.InjectedStatement{
		Statement: pgnodes.NewAlterSequence(
			node.IfExists,
			name.SchemaQualifier.String(),
			name.Name.String(),
			ownedBy,
			warnings...),
		Children: nil,
	}, nil
}
