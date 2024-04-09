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
	"strings"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeCreateView handles *tree.CreateView nodes.
func nodeCreateView(node *tree.CreateView) (*vitess.DDL, error) {
	if node == nil {
		return nil, nil
	}
	if node.Persistence.IsTemporary() {
		return nil, fmt.Errorf("CREATE TEMPORARY VIEW is not yet supported")
	}
	if node.IsRecursive {
		return nil, fmt.Errorf("CREATE RECURSIVE VIEW is not yet supported")
	}
	var checkOption = tree.ViewCheckOptionUnspecified
	var sqlSecurity string
	if node.Options != nil {
		for _, opt := range node.Options {
			switch strings.ToLower(opt.Name) {
			case "check_option":
				switch strings.ToLower(opt.CheckOpt) {
				case "local":
					checkOption = tree.ViewCheckOptionLocal
				case "cascaded":
					checkOption = tree.ViewCheckOptionCascaded
				default:
					return nil, fmt.Errorf(`"ERROR:  syntax error at or near "%s"`, opt.Name)
				}
			case "security_barrier":
				return nil, fmt.Errorf("CREATE VIEW '%s' option is not yet supported", opt.Name)
			case "security_invoker":
				if opt.Security {
					sqlSecurity = "invoker"
				} else {
					sqlSecurity = "definer"
				}
			default:
				return nil, fmt.Errorf(`"ERROR:  syntax error at or near "%s"`, opt.Name)
			}
		}
	}
	if checkOption == tree.ViewCheckOptionLocal {
		return nil, fmt.Errorf("CREATE VIEW ... WITH LOCAL CHECK OPTION is not yet supported")
	}

	tableName, err := nodeTableName(&node.Name)
	if err != nil {
		return nil, err
	}
	selectStmt, err := nodeSelectStatement(node.AsSource.Select)
	if err != nil {
		return nil, err
	}
	var cols = make(vitess.Columns, len(node.ColumnNames))
	for i, col := range node.ColumnNames {
		cols[i] = vitess.NewColIdent(col.String())
	}

	// TODO: need a way to get this information in better way?
	stmtStr := "CREATE "
	if node.Replace {
		stmtStr = fmt.Sprintf("%sOR REPLACE ", stmtStr)
	}
	if sqlSecurity != "" {
		stmtStr = fmt.Sprintf("%sSQL SECURITY %s ", stmtStr, sqlSecurity)
	}
	stmtStr = fmt.Sprintf("%sVIEW %s", stmtStr, tableName.String())
	if node.ColumnNames != nil {
		stmtStr = fmt.Sprintf("%s(%s)", stmtStr, strings.Join(node.ColumnNames.ToStrings(), ", "))
	}
	stmtStr = fmt.Sprintf("%s AS ", stmtStr)
	posStart := len(stmtStr)
	posEnd := posStart + len(node.AsSource.Select.String())

	stmt := &vitess.DDL{
		Action:    vitess.CreateStr,
		OrReplace: node.Replace,
		ViewSpec: &vitess.ViewSpec{
			ViewName: tableName,
			ViewExpr: selectStmt,
			Columns:  cols,
			Security: sqlSecurity,
		},
		SubStatementPositionStart: posStart,
		SubStatementPositionEnd:   posEnd,
	}
	return stmt, nil
}
