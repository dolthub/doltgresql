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

// nodeShowVar handles *tree.ShowVar nodes.
func nodeShowVar(node *tree.ShowVar) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}

	if strings.ToLower(node.Name) == "is_superuser" {
		return nil, fmt.Errorf("SHOW IS_SUPERUSER is not yet supported")
	} else if strings.ToLower(node.Name) == "all" {
		// TODO: need this soon
		return nil, fmt.Errorf("SHOW ALL is not yet supported")
	}

	// TODO: this is a temporary way to get the param value for the current implementation
	//   need better way to get these info
	// We treat namespaced variables (e.g. myvar.myvalue) as user variables.
	// See set_var.go
	isUserVar := strings.Index(node.Name, ".") > 0
	if isUserVar {
		varName := vitess.NewColIdent(node.Name)
		return &vitess.Select{
			SelectExprs: vitess.SelectExprs{
				&vitess.AliasedExpr{
					Expr: &vitess.FuncExpr{
						Name: vitess.NewColIdent("current_setting"),
						Exprs: []vitess.SelectExpr{
							&vitess.AliasedExpr{
								Expr: &vitess.SQLVal{Type: vitess.StrVal, Val: []byte(node.Name)},
							},
						},
					},
					StartParsePos: 7,
					EndParsePos:   7 + len(varName.String()),
					As:            varName,
				},
			},
		}, nil
	} else {
		varName := vitess.NewColIdent("@@session." + node.Name)
		return &vitess.Select{
			SelectExprs: vitess.SelectExprs{
				&vitess.AliasedExpr{
					Expr: &vitess.ColName{
						Name:      varName,
						Qualifier: vitess.TableName{},
					},
					StartParsePos: 7,
					EndParsePos:   7 + len(varName.String()),
					As:            varName,
				},
			},
		}, nil
	}
}
