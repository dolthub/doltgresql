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
	"strings"

	"github.com/cockroachdb/errors"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeIndexElemList converts a tree.IndexElemList to a slice of vitess.IndexField.
func nodeIndexElemList(ctx *Context, node tree.IndexElemList) ([]*vitess.IndexField, error) {
	vitessIndexColumns := make([]*vitess.IndexField, 0, len(node))
	for _, inputColumn := range node {
		if inputColumn.Collation != "" {
			logrus.Warn("index attribute collation is not yet supported, ignoring")
		}
		if inputColumn.ExcludeOp != nil {
			return nil, errors.Errorf("index attribute exclude operator is not yet supported")
		}
		var opClass string
		if inputColumn.OpClass != nil {
			if len(inputColumn.OpClass.Options) > 0 {
				return nil, errors.Errorf("operator class %s has no options", inputColumn.OpClass.Name)
			}
			opClass = strings.TrimPrefix(inputColumn.OpClass.Name, "pg_catalog.")
		}

		order := vitess.AscScr
		nullsOrder := vitess.NullsLastStr
		switch inputColumn.Direction {
		case tree.DefaultDirection, tree.Ascending:
		case tree.Descending:
			order = vitess.DescScr
			nullsOrder = vitess.NullsFirstStr
		default:
			return nil, errors.Errorf("unknown index sorting direction encountered")
		}

		switch inputColumn.NullsOrder {
		case tree.DefaultNullsOrder:
		case tree.NullsFirst:
			nullsOrder = vitess.NullsFirstStr
		case tree.NullsLast:
			nullsOrder = vitess.NullsLastStr
		default:
			return nil, errors.Errorf("unknown NULL ordering for index")
		}

		var expr vitess.Expr
		if inputColumn.Expr != nil {
			var err error
			expr, err = nodeExpr(ctx, inputColumn.Expr)
			if err != nil {
				return nil, err
			}
		}

		vitessIndexColumns = append(vitessIndexColumns, &vitess.IndexField{
			Column:     vitess.NewColIdent(string(inputColumn.Column)),
			Order:      order,
			NullsOrder: nullsOrder,
			OpClass:    opClass,
			Expression: expr,
		})
	}

	return vitessIndexColumns, nil
}
