// Copyright 2024 Dolthub, Inc.
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
	pgnodes "github.com/dolthub/doltgresql/server/node"
)

// nodeCreateDomain handles *tree.CreateDomain nodes.
func nodeCreateDomain(node *tree.CreateDomain) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	name, err := nodeUnresolvedObjectName(node.TypeName)
	if err != nil {
		return nil, err
	}
	_, dataType, err := nodeResolvableTypeReference(node.DataType)
	if err != nil {
		return nil, err
	}
	if node.Collate != "" {
		return nil, fmt.Errorf("domain collation is not yet supported")
	}
	var children []vitess.Expr
	if node.Default != nil {
		defExpr, err := nodeExpr(node.Default)
		if err != nil {
			return nil, err
		}
		// Wrap any default expression using a function call in parens to match MySQL's column default requirements
		if _, ok := defExpr.(*vitess.FuncExpr); ok {
			defExpr = &vitess.ParenExpr{Expr: defExpr}
		}
		children = append(children, defExpr)
	}

	var definedNotNull, definedNull bool
	var checkConstraintNames []string
	var checkConstraintExprs []vitess.Expr
	for _, constraint := range node.Constraints {
		if constraint.Check != nil {
			check, err := verifyAndReplaceValue(node.DataType, constraint.Check)
			if err != nil {
				return nil, err
			}

			expr, err := nodeExpr(check)
			if err != nil {
				return nil, err
			}

			checkConstraintNames = append(checkConstraintNames, string(constraint.Constraint))
			checkConstraintExprs = append(checkConstraintExprs, expr)
		} else if constraint.NotNull {
			definedNotNull = true
			if definedNull {
				return nil, fmt.Errorf("conflicting NULL/NOT NULL constraints")
			}
		} else {
			definedNull = true
			if definedNotNull {
				return nil, fmt.Errorf("conflicting NULL/NOT NULL constraints")
			}
		}
	}
	children = append(children, checkConstraintExprs...)
	return vitess.InjectedStatement{
		Statement: &pgnodes.CreateDomain{
			SchemaName:           name.SchemaQualifier.String(),
			Name:                 name.Name.String(),
			AsType:               dataType,
			Collation:            node.Collate,
			HasDefault:           node.Default != nil,
			IsNotNull:            definedNotNull,
			CheckConstraintNames: checkConstraintNames,
		},
		Children: children,
	}, nil
}

// verifyAndReplaceValue verifies that only VALUE is referenced and replaces it with DomainColumn.
// This function should be used for DOMAIN statements only.
func verifyAndReplaceValue(typ tree.ResolvableTypeReference, expr tree.Expr) (tree.Expr, error) {
	return tree.SimpleVisit(expr, func(visitingExpr tree.Expr) (recurse bool, newExpr tree.Expr, err error) {
		switch v := visitingExpr.(type) {
		case *tree.UnresolvedName:
			if strings.ToLower(v.String()) != "value" {
				return false, nil, fmt.Errorf(`column "%s" does not exist`, v.String())
			}
			return false, tree.DomainColumn{Typ: typ}, nil
		}
		return true, visitingExpr, nil
	})
}
