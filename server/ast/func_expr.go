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
)

// nodeFuncExpr handles *tree.FuncExpr nodes.
func nodeFuncExpr(ctx *Context, node *tree.FuncExpr) (*vitess.FuncExpr, error) {
	if node == nil {
		return nil, nil
	}
	if node.Filter != nil {
		return nil, fmt.Errorf("function filters are not yet supported")
	}
	if node.AggType == tree.OrderedSetAgg {
		return nil, fmt.Errorf("WITHIN GROUP is not yet supported")
	}
	if len(node.OrderBy) > 0 {
		return nil, fmt.Errorf("function ORDER BY is not yet supported")
	}
	var qualifier vitess.TableIdent
	var name vitess.ColIdent
	switch funcRef := node.Func.FunctionReference.(type) {
	case *tree.FunctionDefinition:
		name = vitess.NewColIdent(funcRef.Name)
	case *tree.UnresolvedName:
		if funcRef.NumParts > 2 {
			return nil, fmt.Errorf("referencing items outside the schema or database is not yet supported")
		}
		if funcRef.NumParts == 2 {
			qualifier = vitess.NewTableIdent(funcRef.Parts[1])
		}
		name = vitess.NewColIdent(funcRef.Parts[0])
	default:
		return nil, fmt.Errorf("unknown function reference")
	}
	var distinct bool
	switch node.Type {
	case 0, tree.AllFuncType:
		distinct = false
	case tree.DistinctFuncType:
		distinct = true
	default:
		return nil, fmt.Errorf("unknown function spec type %d", node.Type)
	}
	windowDef, err := nodeWindowDef(ctx, node.WindowDef)
	if err != nil {
		return nil, err
	}
	exprs, err := nodeExprsToSelectExprs(ctx, node.Exprs)
	if err != nil {
		return nil, err
	}
	return &vitess.FuncExpr{
		Qualifier: qualifier,
		Name:      name,
		Distinct:  distinct,
		Exprs:     exprs,
		Over:      (*vitess.Over)(windowDef),
	}, nil
}
