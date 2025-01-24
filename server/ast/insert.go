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
	"github.com/cockroachdb/errors"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/auth"
)

// nodeInsert handles *tree.Insert nodes.
func nodeInsert(ctx *Context, node *tree.Insert) (*vitess.Insert, error) {
	if node == nil {
		return nil, nil
	}
	ctx.Auth().PushAuthType(auth.AuthType_INSERT)
	defer ctx.Auth().PopAuthType()

	if _, ok := node.Returning.(*tree.NoReturningClause); !ok {
		return nil, errors.Errorf("RETURNING is not yet supported")
	}
	var ignore string
	var onDuplicate vitess.OnDup

	if node.OnConflict != nil {
		if isIgnore(node.OnConflict) {
			ignore = vitess.IgnoreStr
		} else if supportedOnConflictClause(node.OnConflict) {
			// TODO: we are ignoring the column names, which are used to infer which index under conflict is to be checked
			updateExprs, err := nodeUpdateExprs(ctx, node.OnConflict.Exprs)
			if err != nil {
				return nil, err
			}
			for _, updateExpr := range updateExprs {
				onDuplicate = append(onDuplicate, updateExpr)
			}
		} else {
			return nil, errors.Errorf("the ON CONFLICT clause provided is not yet supported")
		}
	}
	var tableName vitess.TableName
	switch node := node.Table.(type) {
	case *tree.AliasedTableExpr:
		return nil, errors.Errorf("aliased inserts are not yet supported")
	case *tree.TableName:
		var err error
		tableName, err = nodeTableName(ctx, node)
		if err != nil {
			return nil, err
		}
	case *tree.TableRef:
		return nil, errors.Errorf("table refs are not yet supported")
	default:
		return nil, errors.Errorf("unknown table name type in INSERT: `%T`", node)
	}
	var columns []vitess.ColIdent
	if len(node.Columns) > 0 {
		columns = make([]vitess.ColIdent, len(node.Columns))
		for i := range node.Columns {
			columns[i] = vitess.NewColIdent(string(node.Columns[i]))
		}
	}
	with, err := nodeWith(ctx, node.With)
	if err != nil {
		return nil, err
	}
	var rows vitess.InsertRows
	rows, err = nodeSelect(ctx, node.Rows)
	if err != nil {
		return nil, err
	}

	// For a ValuesStatement with simple rows, GMS expects AliasedValues
	if vSelect, ok := rows.(*vitess.Select); ok && len(vSelect.From) == 1 {
		if aliasedStmt, ok := vSelect.From[0].(*vitess.AliasedTableExpr); ok {
			if valsStmt, ok := aliasedStmt.Expr.(*vitess.ValuesStatement); ok {
				var vals vitess.Values
				if len(valsStmt.Rows) == 0 {
					vals = []vitess.ValTuple{{}}
				} else {
					vals = valsStmt.Rows
				}
				rows = &vitess.AliasedValues{
					Values: vals,
				}
			}
		}
	}
	return &vitess.Insert{
		Action:  vitess.InsertStr,
		Ignore:  ignore,
		Table:   tableName,
		With:    with,
		Columns: columns,
		Rows:    rows,
		OnDup:   onDuplicate,
		Auth: vitess.AuthInformation{
			AuthType:    auth.AuthType_INSERT,
			TargetType:  auth.AuthTargetType_TableIdentifiers,
			TargetNames: []string{tableName.DbQualifier.String(), tableName.SchemaQualifier.String(), tableName.Name.String()},
		},
	}, nil
}

// isIgnore returns true if the ON CONFLICT clause provided is equivalent to INSERT IGNORE in GMS
func isIgnore(conflict *tree.OnConflict) bool {
	return conflict.ArbiterPredicate == nil &&
		conflict.Exprs == nil &&
		conflict.Where == nil &&
		conflict.DoNothing
}

// supportedOnConflictClause returns true if the ON CONFLICT clause given can be represented as
// an ON DUPLICATE KEY UPDATE clause in GMS
func supportedOnConflictClause(conflict *tree.OnConflict) bool {
	if conflict.ArbiterPredicate != nil {
		return false
	}
	if conflict.Where != nil {
		return false
	}
	return true
}
