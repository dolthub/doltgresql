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
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeIndexTableDef handles *tree.IndexTableDef nodes. The parser does not store type information in the index
// definition (PRIMARY KEY, UNIQUE, etc.) so it must be added to this definition by the caller.
func nodeIndexTableDef(node *tree.IndexTableDef) (*vitess.IndexDefinition, error) {
	if node == nil {
		return nil, nil
	}
	if node.IndexParams.IncludeColumns != nil {
		return nil, fmt.Errorf("include columns is not yet supported")
	}
	if len(node.IndexParams.StorageParams) > 0 {
		return nil, fmt.Errorf("storage parameters is not yet supported")
	}
	if node.IndexParams.Tablespace != "" {
		return nil, fmt.Errorf("tablespace is not yet supported")
	}
	//if node.Predicate != nil {
	//	return nil, fmt.Errorf("WHERE is not yet supported")
	//}
	columns := make([]*vitess.IndexColumn, len(node.Columns))
	for i, indexElem := range node.Columns {
		if indexElem.Expr != nil {
			return nil, fmt.Errorf("expression index attribute is not yet supported")
		}
		if indexElem.Collation != "" {
			return nil, fmt.Errorf("index attribute collation is not yet supported")
		}
		if indexElem.OpClass != nil {
			return nil, fmt.Errorf("index attribute operator class is not yet supported")
		}
		switch indexElem.Direction {
		case tree.DefaultDirection:
			// Defaults to ASC
		case tree.Ascending:
			// The only default supported in GMS for now
		case tree.Descending:
			logrus.Warn("descending indexes are not yet supported, ignoring sort order")
		default:
			return nil, fmt.Errorf("unknown index sorting direction encountered")
		}
		if indexElem.Direction != tree.Ascending {
			logrus.Warn("descending indexes are not yet supported, ignoring sort order")
		}
		switch indexElem.NullsOrder {
		case tree.DefaultNullsOrder:
			//TODO: the default NULL order is reversed compared to MySQL, so the default is technically always wrong.
			// To prevent choking on every index, we allow this to proceed (even with incorrect results) for now.
		case tree.NullsFirst:
			// The only form supported in GMS for now
		case tree.NullsLast:
			return nil, fmt.Errorf("NULLS LAST for indexes is not yet supported")
		default:
			return nil, fmt.Errorf("unknown NULL ordering for index")
		}
		if indexElem.ExcludeOp != nil {
			return nil, fmt.Errorf("index attribute exclude operator is not yet supported")
		}
		columns[i] = &vitess.IndexColumn{
			Column: vitess.NewColIdent(string(indexElem.Column)),
			Order:  vitess.AscScr,
		}
	}
	return &vitess.IndexDefinition{
		Info: &vitess.IndexInfo{
			Type: "",
			Name: vitess.NewColIdent(string(node.Name)),
		},
		Columns: columns,
	}, nil
}
