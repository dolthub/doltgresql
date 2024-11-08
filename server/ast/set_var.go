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
	"github.com/dolthub/doltgresql/server/config"
)

// nodeSetVar handles *tree.SetVar nodes.
func nodeSetVar(ctx *Context, node *tree.SetVar) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	// USE statement alias
	if node.Name == "database" {
		// strip off all quotes from the database name
		dbName := strings.TrimPrefix(strings.TrimSuffix(node.Values[0].String(), "'"), "'")
		dbName = strings.TrimPrefix(strings.TrimSuffix(dbName, "\""), "\"")
		dbName = strings.TrimPrefix(strings.TrimSuffix(dbName, "`"), "`")
		return &vitess.Use{DBName: vitess.NewTableIdent(dbName)}, nil
	}
	if node.Namespace == "" && !config.IsValidPostgresConfigParameter(node.Name) && !config.IsValidDoltConfigParameter(node.Name) {
		return nil, fmt.Errorf(`ERROR: unrecognized configuration parameter "%s"`, node.Name)
	}
	if node.IsLocal {
		// TODO: takes effect for only the current transaction rather than the current session.
		return nil, fmt.Errorf("SET LOCAL is not yet supported")
	}
	var expr vitess.Expr
	var err error
	if len(node.Values) == 0 {
		// sanity check
		return nil, fmt.Errorf(`ERROR: syntax error at or near ";"'`)
	} else if len(node.Values) > 1 {
		vals := make([]string, len(node.Values))
		for i, val := range node.Values {
			vals[i] = val.String()
		}
		expr = &vitess.ColName{
			Name: vitess.NewColIdent(strings.Join(vals, ", ")),
		}
	} else {
		expr, err = nodeExpr(ctx, node.Values[0])
		if err != nil {
			return nil, err
		}
	}

	if node.Namespace == "" {
		return &vitess.Set{
			Exprs: vitess.SetVarExprs{&vitess.SetVarExpr{
				Scope: vitess.SetScope_Session,
				Name: &vitess.ColName{
					Name: vitess.NewColIdent(node.Name),
				},
				Expr: expr,
			}},
		}, nil
	} else {
		return &vitess.Set{
			Exprs: vitess.SetVarExprs{&vitess.SetVarExpr{
				Scope: vitess.SetScope_User,
				Name: &vitess.ColName{
					Name: vitess.NewColIdent(fmt.Sprintf("%s.%s", node.Namespace, node.Name)),
				},
				Expr: expr,
			}},
		}, nil
	}
}
