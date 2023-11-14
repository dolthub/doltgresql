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

// nodeSelectClause handles tree.SelectClause nodes.
func nodeSelectClause(node *tree.SelectClause) (*vitess.Select, error) {
	if node == nil {
		return nil, nil
	}
	selectExprs, err := nodeSelectExprs(node.Exprs)
	if err != nil {
		return nil, err
	}
	from, err := nodeFrom(node.From)
	if err != nil {
		return nil, err
	}
	var distinct bool
	if node.Distinct {
		distinct = true
	}
	if len(node.DistinctOn) > 0 {
		return nil, fmt.Errorf("DISTINCT ON is not yet supported")
	}
	where, err := nodeWhere(node.Where)
	if err != nil {
		return nil, err
	}
	having, err := nodeWhere(node.Having)
	if err != nil {
		return nil, err
	}
	groupBy, err := nodeGroupBy(node.GroupBy)
	if err != nil {
		return nil, err
	}
	window, err := nodeWindow(node.Window)
	if err != nil {
		return nil, err
	}
	return &vitess.Select{
		QueryOpts:   vitess.QueryOpts{Distinct: distinct},
		SelectExprs: selectExprs,
		From:        from,
		Where:       where,
		GroupBy:     groupBy,
		Having:      having,
		Window:      window,
	}, nil
}
