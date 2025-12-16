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

package analyzer

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/expression"

	pgexprs "github.com/dolthub/doltgresql/server/expression"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// SplitConjunction breaks AND expressions into their left and right parts, recursively.
func SplitConjunction(expr sql.Expression) []sql.Expression {
	if expr == nil {
		return nil
	}
	switch expr := expr.(type) {
	case *expression.And:
		return append(
			SplitConjunction(expr.LeftChild),
			SplitConjunction(expr.RightChild)...,
		)
	case *pgexprs.GMSCast:
		// We should check to see if we need to preserve the cast on each child individually
		split := SplitConjunction(expr.Child())
		for i := range split {
			if _, ok := split[i].Type().(*pgtypes.DoltgresType); !ok {
				split[i] = pgexprs.NewGMSCast(split[i])
			}
		}
		return split
	default:
		return []sql.Expression{expr}
	}
}

// LogicTreeWalker is a walker that removes GMSCast and other Doltgres specific expression nodes from
// logic expression trees. This allows the analyzer logic to correctly reason about expressions in filters
// to apply indexes.
type LogicTreeWalker struct{}

var _ analyzer.LogicTreeWalker = &LogicTreeWalker{}

func (l *LogicTreeWalker) Next(e sql.Expression) sql.Expression {
	switch expr := e.(type) {
	case *pgexprs.GMSCast:
		return l.Next(expr.Child())
	default:
		return e
	}
}
