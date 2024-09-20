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

// nodeIndexElemList converts a tree.IndexElemList to a slice of vitess.IndexColumn.
func nodeIndexElemList(node tree.IndexElemList) ([]*vitess.IndexColumn, error) {
	vitessIndexColumns := make([]*vitess.IndexColumn, 0, len(node))
	for i := range node {
		inputColumn := node[i]
		if inputColumn.Collation != "" {
			return nil, fmt.Errorf("collation is not yet supported " +
				"for index column constraint definitions")
		}

		if inputColumn.Expr != nil {
			return nil, fmt.Errorf("expressions are not yet supported " +
				"for index column constraint definitions")
		}

		if inputColumn.ExcludeOp != nil {
			return nil, fmt.Errorf("EXCLUDE is not yet supported")
		}

		if inputColumn.NullsOrder == tree.NullsLast {
			return nil, fmt.Errorf("NULLS LAST is not yet supported")
		}

		if inputColumn.OpClass != nil {
			return nil, fmt.Errorf("operator classes are not yet supported")
		}

		order := vitess.AscScr
		if inputColumn.Direction == tree.Descending {
			order = vitess.DescScr
		}

		vitessIndexColumns = append(vitessIndexColumns, &vitess.IndexColumn{
			Column: vitess.NewColIdent(inputColumn.Column.String()),
			Order:  order,
		})
	}

	return vitessIndexColumns, nil
}
