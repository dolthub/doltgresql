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

// nodeCall handles *tree.Call nodes.
func nodeCall(ctx *Context, node *tree.Call) (*vitess.Call, error) {
	if node == nil {
		return nil, nil
	}
	if node.Procedure.Type != 0 {
		return nil, fmt.Errorf("procedure distinction is not yet supported")
	}
	if node.Procedure.Filter != nil {
		return nil, fmt.Errorf("procedure filters are not yet supported")
	}
	if node.Procedure.WindowDef != nil {
		return nil, fmt.Errorf("procedure window definitions are not yet supported")
	}
	if node.Procedure.AggType != tree.GeneralAgg {
		return nil, fmt.Errorf("procedure aggregation is not yet supported")
	}
	if len(node.Procedure.OrderBy) > 0 {
		return nil, fmt.Errorf("procedure ORDER BY is not yet supported")
	}
	var qualifier vitess.TableIdent
	var name vitess.ColIdent
	switch funcRef := node.Procedure.Func.FunctionReference.(type) {
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
	exprs, err := nodeExprs(ctx, node.Procedure.Exprs)
	if err != nil {
		return nil, err
	}
	return &vitess.Call{
		ProcName: vitess.ProcedureName{
			Name:      name,
			Qualifier: qualifier,
		},
		Params: exprs,
	}, nil
}
