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

// nodeForeignKeyConstraintTableDef handles *tree.ForeignKeyConstraintTableDef nodes.
func nodeForeignKeyConstraintTableDef(node *tree.ForeignKeyConstraintTableDef) (*vitess.ForeignKeyDefinition, error) {
	if node == nil {
		return nil, nil
	}
	switch node.Match {
	case tree.MatchSimple:
		// This is the default behavior
	case tree.MatchFull:
		return nil, fmt.Errorf("MATCH FULL is not yet supported")
	case tree.MatchPartial:
		return nil, fmt.Errorf("MATCH PARTIAL is not yet supported")
	default:
		return nil, fmt.Errorf("unknown foreign key MATCH strategy")
	}
	tableName, err := nodeTableName(&node.Table)
	if err != nil {
		return nil, err
	}
	fromCols := make([]vitess.ColIdent, len(node.FromCols))
	for i := range node.FromCols {
		fromCols[i] = vitess.NewColIdent(string(node.FromCols[i]))
	}
	toCols := make([]vitess.ColIdent, len(node.ToCols))
	for i := range node.ToCols {
		toCols[i] = vitess.NewColIdent(string(node.ToCols[i]))
	}
	var refActions [2]vitess.ReferenceAction
	for i, refAction := range []tree.RefAction{node.Actions.Delete, node.Actions.Update} {
		switch refAction.Action {
		case tree.NoAction:
			refActions[i] = vitess.NoAction
		case tree.Restrict:
			refActions[i] = vitess.Restrict
		case tree.SetNull:
			refActions[i] = vitess.SetNull
			if refAction.Columns != nil {
				return nil, fmt.Errorf("SET NULL <columns> is not yet supported")
			}
		case tree.SetDefault:
			// GMS doesn't support this as MySQL doesn't support this
			refActions[i] = vitess.SetDefault
			if refAction.Columns != nil {
				return nil, fmt.Errorf("SET NULL <columns> is not yet supported")
			}
		case tree.Cascade:
			refActions[i] = vitess.Cascade
		default:
			return nil, fmt.Errorf("unknown foreign key reference action encountered")
		}
	}
	return &vitess.ForeignKeyDefinition{
		Source:            fromCols,
		ReferencedTable:   tableName,
		ReferencedColumns: toCols,
		OnDelete:          refActions[0],
		OnUpdate:          refActions[1],
	}, nil
}
