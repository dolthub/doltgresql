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
func nodeCreateView(ctx *Context, node *tree.CreateView) (*vitess.DDL, error) {
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

	if checkOption != tree.ViewCheckOptionUnspecified && node.CheckOption != tree.ViewCheckOptionUnspecified {
		return nil, fmt.Errorf(`ERROR:  parameter "check_option" specified more than once`)
	} else {
		checkOption = node.CheckOption
	}

	vCheckOpt := vitess.ViewCheckOptionUnspecified
	switch checkOption {
	case tree.ViewCheckOptionCascaded:
		vCheckOpt = vitess.ViewCheckOptionCascaded
	case tree.ViewCheckOptionLocal:
		vCheckOpt = vitess.ViewCheckOptionLocal
	default:
	}

	tableName, err := nodeTableName(ctx, &node.Name)
	if err != nil {
		return nil, err
	}
	selectStmt, err := nodeSelectStatement(ctx, node.AsSource.Select)
	if err != nil {
		return nil, err
	}
	var cols = make(vitess.Columns, len(node.ColumnNames))
	for i, col := range node.ColumnNames {
		cols[i] = vitess.NewColIdent(col.String())
	}

	stmt := &vitess.DDL{
		Action:    vitess.CreateStr,
		OrReplace: node.Replace,
		ViewSpec: &vitess.ViewSpec{
			ViewName:    tableName,
			ViewExpr:    selectStmt,
			Columns:     cols,
			Security:    sqlSecurity,
			CheckOption: vCheckOpt,
		},
		SubStatementStr: node.AsSource.Select.String(),
	}
	return stmt, nil
}
