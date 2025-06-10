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
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	pgexprs "github.com/dolthub/doltgresql/server/expression"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// AssignUpdateCasts adds the appropriate assign casts for updates.
func AssignUpdateCasts(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	update, ok := node.(*plan.Update)
	if !ok {
		return node, transform.SameTree, nil
	}
	var newUpdate sql.Node
	switch child := update.Child.(type) {
	case *plan.UpdateSource:
		newUpdateSource, err := assignUpdateCastsHandleSource(child)
		if err != nil {
			return nil, transform.NewTree, err
		}
		newUpdate, err = update.WithChildren(newUpdateSource)
		if err != nil {
			return nil, transform.NewTree, err
		}
	case *plan.ForeignKeyHandler:
		updateSource, ok := child.OriginalNode.(*plan.UpdateSource)
		if !ok {
			return nil, transform.NewTree, errors.Errorf("UPDATE: assumption that Foreign Key child is always UpdateSource is incorrect: %T", child.OriginalNode)
		}
		newUpdateSource, err := assignUpdateCastsHandleSource(updateSource)
		if err != nil {
			return nil, transform.NewTree, err
		}
		newHandler, err := child.WithChildren(newUpdateSource)
		if err != nil {
			return nil, transform.NewTree, err
		}
		newUpdate, err = update.WithChildren(newHandler)
		if err != nil {
			return nil, transform.NewTree, err
		}
	case *plan.UpdateJoin:
		updateSource, ok := child.Child.(*plan.UpdateSource)
		if !ok {
			return nil, transform.NewTree, fmt.Errorf("UPDATE: unknown source type: %T", child.Child)
		}

		newUpdateSource, err := assignUpdateCastsHandleSource(updateSource)
		if err != nil {
			return nil, transform.NewTree, err
		}
		newHandler, err := child.WithChildren(newUpdateSource)
		if err != nil {
			return nil, transform.NewTree, err
		}
		newUpdate, err = update.WithChildren(newHandler)
		if err != nil {
			return nil, transform.NewTree, err
		}
	default:
		return nil, transform.NewTree, errors.Errorf("UPDATE: unknown source type: %T", child)
	}
	return newUpdate, transform.NewTree, nil
}

// assignUpdateCastsHandleSource handles the *plan.UpdateSource portion of AssignUpdateCasts.
func assignUpdateCastsHandleSource(updateSource *plan.UpdateSource) (*plan.UpdateSource, error) {
	updateExprs := updateSource.UpdateExprs
	newUpdateExprs, err := assignUpdateFieldCasts(updateExprs)
	if err != nil {
		return nil, err
	}
	newUpdateSource, err := updateSource.WithExpressions(newUpdateExprs...)
	if err != nil {
		return nil, err
	}
	return newUpdateSource.(*plan.UpdateSource), nil
}

func assignUpdateFieldCasts(updateExprs []sql.Expression) ([]sql.Expression, error) {
	newUpdateExprs := make([]sql.Expression, len(updateExprs))
	for i, updateExpr := range updateExprs {
		setField, ok := updateExpr.(*expression.SetField)
		if !ok {
			return nil, errors.Errorf("UPDATE: assumption that expression is always SetField is incorrect: %T", updateExpr)
		}
		fromType, ok := setField.RightChild.Type().(*pgtypes.DoltgresType)
		if !ok {
			return nil, errors.Errorf("UPDATE: non-Doltgres type found in source: %s", setField.RightChild.String())
		}
		toType, ok := setField.LeftChild.Type().(*pgtypes.DoltgresType)
		if !ok {
			return nil, errors.Errorf("UPDATE: non-Doltgres type found in destination: %s", setField.LeftChild.String())
		}
		// We only assign the existing expression if the types perfectly match (same parameters), otherwise we'll cast
		if fromType.Equals(toType) {
			newUpdateExprs[i] = setField
		} else {
			newSetField, err := setField.WithChildren(setField.LeftChild, pgexprs.NewAssignmentCast(setField.RightChild, fromType, toType))
			if err != nil {
				return nil, err
			}
			newUpdateExprs[i] = newSetField
		}
	}
	return newUpdateExprs, nil
}
