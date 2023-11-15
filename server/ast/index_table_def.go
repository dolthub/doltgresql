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
	if node.Sharded != nil {
		return nil, fmt.Errorf("sharding is not yet supported")
	}
	if len(node.Storing) > 0 {
		return nil, fmt.Errorf("INCLUDE is not yet supported")
	}
	if node.Interleave != nil {
		return nil, fmt.Errorf("INTERLEAVE is not yet supported")
	}
	if node.Inverted {
		return nil, fmt.Errorf("inverted indexes are not yet supported")
	}
	if node.PartitionBy != nil {
		return nil, fmt.Errorf("PARTITION BY is not yet supported")
	}
	if len(node.StorageParams) > 0 {
		return nil, fmt.Errorf("storage parameters are not yet supported")
	}
	if node.Predicate != nil {
		return nil, fmt.Errorf("WHERE is not yet supported")
	}
	columns := make([]*vitess.IndexColumn, len(node.Columns))
	for i, indexElem := range node.Columns {
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
