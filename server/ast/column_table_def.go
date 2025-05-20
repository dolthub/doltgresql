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
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// DoltCreateTablePlaceholderSequenceName is a Placeholder name used in translating computed columns to generated
// columns that involve a sequence, used later in analysis
const DoltCreateTablePlaceholderSequenceName = "dolt_create_table_placeholder_sequence"

// nodeColumnTableDef handles *tree.ColumnTableDef nodes.
func nodeColumnTableDef(ctx *Context, node *tree.ColumnTableDef) (*vitess.ColumnDefinition, error) {
	if node == nil {
		return nil, nil
	}
	if len(node.Nullable.ConstraintName) > 0 ||
		len(node.DefaultExpr.ConstraintName) > 0 ||
		len(node.UniqueConstraintName) > 0 {
		return nil, errors.Errorf("non-foreign key column constraint names are not yet supported")
	}
	convertType, resolvedType, err := nodeResolvableTypeReference(ctx, node.Type)
	if err != nil {
		return nil, err
	}

	if resolvedType == pgtypes.Record {
		return nil, errors.Errorf(`column "%s" has pseudo-type record`, node.Name.String())
	}

	var isNull vitess.BoolVal
	var isNotNull vitess.BoolVal
	switch node.Nullable.Nullability {
	case tree.NotNull:
		isNull = false
		isNotNull = true
	case tree.Null:
		isNull = true
		isNotNull = false
	case tree.SilentNull:
		isNull = true
		isNotNull = false
	default:
		return nil, errors.Errorf("unknown NULL type encountered")
	}
	keyOpt := vitess.ColumnKeyOption(0) // colKeyNone, unexported for some reason
	if node.PrimaryKey.IsPrimaryKey {
		keyOpt = 1 // colKeyPrimary
		isNull = false
		isNotNull = true
	} else if node.Unique {
		keyOpt = 3 // colKeyUnique
	}
	defaultExpr, err := nodeExpr(ctx, node.DefaultExpr.Expr)
	if err != nil {
		return nil, err
	}
	// Wrap any default expression using a function call in parens to match MySQL's column default requirements
	if _, ok := defaultExpr.(*vitess.FuncExpr); ok {
		defaultExpr = &vitess.ParenExpr{Expr: defaultExpr}
	}

	var fkDef *vitess.ForeignKeyDefinition
	if node.References.Table != nil {
		if len(node.References.Col) == 0 {
			return nil, errors.Errorf("implicit primary key matching on column foreign key is not yet supported")
		}
		fkDef, err = nodeForeignKeyConstraintTableDef(ctx, &tree.ForeignKeyConstraintTableDef{
			Name:     node.References.ConstraintName,
			Table:    *node.References.Table,
			FromCols: tree.NameList{node.Name},
			ToCols:   tree.NameList{node.References.Col},
			Actions:  node.References.Actions,
			Match:    node.References.Match,
		})
		if err != nil {
			return nil, err
		}
	}

	if len(node.Computed.Options) > 0 {
		return nil, errors.Errorf("sequence options are not yet supported, create a sequence separately")
	}

	var generated vitess.Expr
	hasGeneratedExpr := node.IsComputed() && node.Computed.Expr != nil
	computedByDefaultAsIdentity := node.IsComputed() && !hasGeneratedExpr && node.Computed.ByDefault
	computedAsIdentity := node.IsComputed() && !hasGeneratedExpr && !node.Computed.ByDefault

	if hasGeneratedExpr {
		generated, err = nodeExpr(ctx, node.Computed.Expr)
		if err != nil {
			return nil, err
		}
	} else if computedAsIdentity {
		generated, err = nodeExpr(ctx, &tree.FuncExpr{
			Func: tree.WrapFunction("nextval"),
			Exprs: tree.Exprs{
				tree.NewStrVal(DoltCreateTablePlaceholderSequenceName),
			},
		})
		if err != nil {
			return nil, err
		}
	}

	if generated != nil {
		// GMS requires the AST to wrap function expressions in parens
		if _, ok := generated.(*vitess.FuncExpr); ok {
			generated = &vitess.ParenExpr{Expr: generated}
		}

		// clean up the expressions generated here. our default expression handling generates aliases that aren't
		// appropriate in this context.
		generated = clearAliases(generated)
	}

	if node.IsSerial || computedByDefaultAsIdentity || computedAsIdentity {
		if resolvedType.IsEmptyType() {
			return nil, errors.Errorf("serial type was not resolvable")
		}
		switch resolvedType.ID {
		case pgtypes.Int16.ID:
			resolvedType = pgtypes.Int16Serial
		case pgtypes.Int32.ID:
			resolvedType = pgtypes.Int32Serial
		case pgtypes.Int64.ID:
			resolvedType = pgtypes.Int64Serial
		default:
			return nil, errors.Errorf(`type "%s" cannot be serial`, resolvedType.String())
		}
		if defaultExpr != nil {
			return nil, errors.Errorf(`multiple default values specified for column "%s"`, node.Name)
		}
	}

	colDef := &vitess.ColumnDefinition{
		Name: vitess.NewColIdent(string(node.Name)),
		Type: vitess.ColumnType{
			Type:          convertType.Type,
			ResolvedType:  resolvedType,
			Null:          isNull,
			NotNull:       isNotNull,
			Autoincrement: false,
			Default:       defaultExpr,
			Length:        convertType.Length,
			Scale:         convertType.Scale,
			KeyOpt:        keyOpt,
			ForeignKeyDef: fkDef,
			GeneratedExpr: generated,
			Stored:        generated != nil, // postgres generated columns are always stored, never virtual
		},
	}

	if len(node.CheckExprs) > 0 {
		// TODO: vitess does not support multiple check constraint on a single column
		if len(node.CheckExprs) > 1 {
			return nil, errors.Errorf("column-declared multiple CHECK expressions are not yet supported")
		}
		var checkConstraints = make([]*vitess.ConstraintDefinition, len(node.CheckExprs))
		for i, checkExpr := range node.CheckExprs {
			expr, err := nodeExpr(ctx, checkExpr.Expr)
			if err != nil {
				return nil, err
			}
			checkConstraints[i] = &vitess.ConstraintDefinition{
				Name: string(checkExpr.ConstraintName),
				Details: &vitess.CheckConstraintDefinition{
					Expr:     expr,
					Enforced: true,
				},
			}
		}
		// until we support multiple constraints in a column
		colDef.Type.Constraint = checkConstraints[0]
	}
	return colDef, nil
}

// clearAliases removes As and InputExpression from any AliasedExpr in the expression tree given. This is required
// in some contexts where we expect the expression to serialize to a string without any alias names.
func clearAliases(e vitess.Expr) vitess.Expr {
	_ = vitess.Walk(func(node vitess.SQLNode) (kontinue bool, err error) {
		if expr, ok := node.(*vitess.AliasedExpr); ok {
			expr.As = vitess.ColIdent{}
			expr.InputExpression = ""
		}
		return true, nil
	}, e)
	return e
}
