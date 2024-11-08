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
		return nil, fmt.Errorf("RETURNING is not yet supported")
	}
	var ignore string
	if node.OnConflict != nil {
		// Currently, only ON CONFLICT DO NOTHING is supported, which is equivalent to INSERT IGNORE in GMS
		if node.OnConflict.Columns != nil ||
			node.OnConflict.ArbiterPredicate != nil ||
			node.OnConflict.Exprs != nil ||
			node.OnConflict.Where != nil ||
			!node.OnConflict.DoNothing {
			return nil, fmt.Errorf("the ON CONFLICT clause provided is not yet supported")
		}
		ignore = vitess.IgnoreStr
	}
	var tableName vitess.TableName
	switch node := node.Table.(type) {
	case *tree.AliasedTableExpr:
		return nil, fmt.Errorf("aliased inserts are not yet supported")
	case *tree.TableName:
		var err error
		tableName, err = nodeTableName(ctx, node)
		if err != nil {
			return nil, err
		}
	case *tree.TableRef:
		return nil, fmt.Errorf("table refs are not yet supported")
	default:
		return nil, fmt.Errorf("unknown table name type in INSERT: `%T`", node)
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

	// GMS For a ValuesStatement with simple rows, GMS expects AliasedValues
	if vSelect, ok := rows.(*vitess.Select); ok && len(vSelect.From) == 1 {
		if valsStmt, ok := vSelect.From[0].(*vitess.ValuesStatement); ok {
			rows = &vitess.AliasedValues{
				Values: valsStmt.Rows,
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
		Auth: vitess.AuthInformation{
			AuthType:    auth.AuthType_INSERT,
			TargetType:  auth.AuthTargetType_SingleTableIdentifier,
			TargetNames: []string{tableName.SchemaQualifier.String(), tableName.Name.String()},
		},
	}, nil
}
