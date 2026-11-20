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
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeLockingClause handles *tree.LockingClause nodes. It returns nil for a SELECT that has no locking clause.
func nodeLockingClause(ctx *Context, node tree.LockingClause) (*vitess.Lock, error) {
	if len(node) == 0 {
		return nil, nil
	}

	items := make([]vitess.RowLockItem, len(node))
	for i, item := range node {
		var strength vitess.RowLockStrength
		switch item.Strength {
		case tree.ForKeyShare:
			strength = vitess.RowLockForKeyShare
		case tree.ForShare:
			strength = vitess.RowLockForShare
		case tree.ForNoKeyUpdate:
			strength = vitess.RowLockForNoKeyUpdate
		case tree.ForUpdate:
			strength = vitess.RowLockForUpdate
		default:
			return nil, pgerror.Newf(pgcode.FeatureNotSupported, "unknown locking strength: %d", item.Strength)
		}

		var waitPolicy vitess.RowLockWaitPolicy
		switch item.WaitPolicy {
		case tree.LockWaitBlock:
			waitPolicy = vitess.RowLockWait
		case tree.LockWaitSkip:
			waitPolicy = vitess.RowLockSkipLocked
		case tree.LockWaitError:
			waitPolicy = vitess.RowLockNoWait
		default:
			return nil, pgerror.Newf(pgcode.FeatureNotSupported, "unknown locking wait policy: %d", item.WaitPolicy)
		}

		var tables vitess.TableNames
		for _, target := range item.Targets {
			if target.ExplicitSchema || target.ExplicitCatalog {
				return nil, pgerror.Newf(pgcode.Syntax, "%s must specify unqualified relation names", item.Strength)
			}

			tables = append(tables, vitess.TableName{Name: vitess.NewTableIdent(string(target.ObjectName))})
		}

		items[i] = vitess.RowLockItem{
			Strength:   strength,
			WaitPolicy: waitPolicy,
			Tables:     tables,
		}
	}

	return &vitess.Lock{
		Type:  vitess.ForUpdateStr,
		Items: items,
	}, nil
}

// applyLockingClause attaches a locking clause to the SELECT it was written on, after checking that the SELECT is of
// a kind whose rows can be locked: one that returns rows of its tables as they are, and not combinations of them.
func applyLockingClause(node tree.LockingClause, lock *vitess.Lock, selectStmt vitess.SelectStatement) error {
	if lock == nil {
		return nil
	}

	strength := node[0].Strength
	switch selectStmt := selectStmt.(type) {
	case *vitess.ParenSelect:
		return applyLockingClause(node, lock, selectStmt.Select)
	case *vitess.SetOp:
		return lockingNotAllowed(strength, "UNION/INTERSECT/EXCEPT")
	case *vitess.Select:
		switch {
		case selectStmt.QueryOpts.Distinct || len(selectStmt.QueryOpts.DistinctOn) > 0:
			return lockingNotAllowed(strength, "DISTINCT clause")
		case len(selectStmt.GroupBy) > 0:
			return lockingNotAllowed(strength, "GROUP BY clause")
		case selectStmt.Having != nil:
			return lockingNotAllowed(strength, "HAVING clause")
		case len(selectStmt.Window) > 0:
			return lockingNotAllowed(strength, "window functions")
		}

		var err error
		_ = vitess.Walk(func(n vitess.SQLNode) (bool, error) {
			switch n := n.(type) {
			case *vitess.Subquery:
				// A subquery's rows are its own business.
				return false, nil
			case *vitess.FuncExpr:
				if n.Over != nil {
					err = lockingNotAllowed(strength, "window functions")
				} else if n.IsAggregate() {
					err = lockingNotAllowed(strength, "aggregate functions")
				}
			}

			return err == nil, nil
		}, selectStmt.SelectExprs)
		if err != nil {
			return err
		}

		selectStmt.Lock = lock
		return nil
	default:
		return lockingNotAllowed(strength, "this kind of SELECT")
	}
}

func lockingNotAllowed(strength tree.LockingStrength, with string) error {
	return pgerror.Newf(pgcode.FeatureNotSupported, "%s is not allowed with %s", strength, with)
}
