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
	"strconv"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/postgres/parser/types"
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
	if node.PrimaryKey.Sharded {
		return nil, fmt.Errorf("sharded columns are not yet supported")
	}
	if node.Family.Create || len(node.Family.Name) > 0 {
		return nil, fmt.Errorf("FAMILY is not yet supported")
	}
	var columnTypeName string
	var columnTypeLength *vitess.SQLVal
	var columnTypeScale *vitess.SQLVal
	switch columnType := node.Type.(type) {
	case *tree.ArrayTypeReference:
		return nil, fmt.Errorf("array types are not yet supported")
	case *tree.OIDTypeReference:
		return nil, fmt.Errorf("referencing types by their OID is not yet supported")
	case *tree.UnresolvedObjectName:
		return nil, fmt.Errorf("type declaration format is not yet supported")
	case *types.GeoMetadata:
		return nil, fmt.Errorf("geometry types are not yet supported")
	case *types.T:
		columnTypeName = columnType.SQLStandardName()
		switch columnType.Family() {
		case types.DecimalFamily:
			columnTypeLength = vitess.NewIntVal([]byte(strconv.Itoa(int(columnType.Precision()))))
			columnTypeScale = vitess.NewIntVal([]byte(strconv.Itoa(int(columnType.Scale()))))
		case types.JsonFamily:
			columnTypeName = "JSON"
		case types.StringFamily:
			columnTypeLength = vitess.NewIntVal([]byte(strconv.Itoa(int(columnType.Width()))))
		case types.TimestampFamily:
			columnTypeName = columnType.Name()
		}
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
	return &vitess.ColumnDefinition{
		Name: vitess.NewColIdent(string(node.Name)),
		Type: vitess.ColumnType{
			Type:          columnTypeName,
			Null:          isNull,
			NotNull:       isNotNull,
			Autoincrement: vitess.BoolVal(node.IsSerial),
			Default:       defaultExpr,
			Length:        columnTypeLength,
			Scale:         columnTypeScale,
			KeyOpt:        keyOpt,
			ForeignKeyDef: fkDef,
			GeneratedExpr: generated,
			Stored:        generatedStored,
		},
	}, nil
}
