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

// nodeColumnTableDef handles *tree.ColumnTableDef nodes.
func nodeColumnTableDef(node *tree.ColumnTableDef) (_ *vitess.ColumnDefinition, err error) {
	if node == nil {
		return nil, nil
	}
	if len(node.Nullable.ConstraintName) > 0 ||
		len(node.DefaultExpr.ConstraintName) > 0 ||
		len(node.UniqueConstraintName) > 0 {
		return nil, fmt.Errorf("non-foreign key column constraint names are not yet supported")
	}
	convertType, resolvedType, err := nodeResolvableTypeReference(node.Type)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("unknown NULL type encountered")
	}
	keyOpt := vitess.ColumnKeyOption(0) // colKeyNone, unexported for some reason
	if node.PrimaryKey.IsPrimaryKey {
		keyOpt = 1 // colKeyPrimary
		isNull = false
		isNotNull = true
	} else if node.Unique {
		keyOpt = 3 // colKeyUnique
	}
	defaultExpr, err := nodeExpr(node.DefaultExpr.Expr)
	if err != nil {
		return nil, err
	}
	if len(node.CheckExprs) > 0 {
		return nil, fmt.Errorf("column-declared CHECK expressions are not yet supported")
	}
	var fkDef *vitess.ForeignKeyDefinition
	if node.References.Table != nil {
		if len(node.References.Col) == 0 {
			return nil, fmt.Errorf("implicit primary key matching on column foreign key is not yet supported")
		}
		fkDef, err = nodeForeignKeyConstraintTableDef(&tree.ForeignKeyConstraintTableDef{
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
	var generated vitess.Expr
	var generatedStored vitess.BoolVal
	if node.Computed.Computed {
		generated, err = nodeExpr(node.Computed.Expr)
		if err != nil {
			return nil, err
		}
		//TODO: need to add support for VIRTUAL in the parser
		generatedStored = true
	}
	// TODO: need to add support for SEQUENCE.
	return &vitess.ColumnDefinition{
		Name: vitess.NewColIdent(string(node.Name)),
		Type: vitess.ColumnType{
			Type:          convertType.Type,
			ResolvedType:  resolvedType,
			Null:          isNull,
			NotNull:       isNotNull,
			Autoincrement: vitess.BoolVal(node.IsSerial),
			Default:       defaultExpr,
			Length:        convertType.Length,
			Scale:         convertType.Scale,
			KeyOpt:        keyOpt,
			ForeignKeyDef: fkDef,
			GeneratedExpr: generated,
			Stored:        generatedStored,
		},
	}, nil
}
