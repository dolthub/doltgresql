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

package index

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"

	pgexprs "github.com/dolthub/doltgresql/server/expression"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// SplitDisjunction breaks OR expressions into their left and right parts, recursively. Also handles expressions that
// can be approximated as OR expressions, such as IN for tuples.
func SplitDisjunction(expr sql.Expression) []sql.Expression {
	if expr == nil {
		return nil
	}
	switch expr := expr.(type) {
	case *expression.Or:
		return append(
			SplitDisjunction(expr.LeftChild),
			SplitDisjunction(expr.RightChild)...,
		)
	case *pgexprs.GMSCast:
		// We should check to see if we need to preserve the cast on each child individually
		split := SplitDisjunction(expr.Child())
		for i := range split {
			if _, ok := split[i].Type().(*pgtypes.DoltgresType); !ok {
				split[i] = pgexprs.NewGMSCast(split[i])
			}
		}
		return split
	case *pgexprs.InTuple:
		return SplitDisjunction(expr.Decay())
	default:
		return []sql.Expression{expr}
	}
}

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

// SplitDisjunctions performs the same operation as SplitDisjunction, except that it applies to a slice.
func SplitDisjunctions(exprs []sql.Expression) []sql.Expression {
	if len(exprs) == 0 {
		return nil
	}
	// New slice will be at least the size of the incoming slice
	newExprs := make([]sql.Expression, 0, len(exprs))
	for _, expr := range exprs {
		newExprs = append(newExprs, SplitDisjunction(expr)...)
	}
	return newExprs
}

// SplitConjunctions performs the same operation as SplitConjunction, except that it applies to a slice.
func SplitConjunctions(exprs []sql.Expression) []sql.Expression {
	if len(exprs) == 0 {
		return nil
	}
	// New slice will be at least the size of the incoming slice
	newExprs := make([]sql.Expression, 0, len(exprs))
	for _, expr := range exprs {
		newExprs = append(newExprs, SplitConjunction(expr)...)
	}
	return newExprs
}
