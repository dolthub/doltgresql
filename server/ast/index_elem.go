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
	"github.com/cockroachdb/errors"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeIndexElemList converts a tree.IndexElemList to a slice of vitess.IndexColumn.
func nodeIndexElemList(ctx *Context, node tree.IndexElemList) ([]*vitess.IndexColumn, error) {
	vitessIndexColumns := make([]*vitess.IndexColumn, 0, len(node))
	for _, inputColumn := range node {
		if inputColumn.Expr != nil {
			return nil, errors.Errorf("expression index attribute is not yet supported")
		}
		if inputColumn.Collation != "" {
			logrus.Warn("index attribute collation is not yet supported, ignoring")
		}
		if inputColumn.OpClass != nil {
			logrus.Warn("index attribute operator class is not yet supported, ignoring")
		}
		if inputColumn.ExcludeOp != nil {
			return nil, errors.Errorf("index attribute exclude operator is not yet supported")
		}

		switch inputColumn.Direction {
		case tree.DefaultDirection:
			// Defaults to ASC
		case tree.Ascending:
			// The only default supported in GMS for now
		case tree.Descending:
			logrus.Warn("descending indexes are not yet supported, ignoring sort order")
		default:
			return nil, errors.Errorf("unknown index sorting direction encountered")
		}

		switch inputColumn.NullsOrder {
		case tree.DefaultNullsOrder:
			// TODO: the default NULL order is reversed compared to MySQL, so the default is technically always wrong.
			//       To prevent choking on every index, we allow this to proceed (even with incorrect results) for now.
		case tree.NullsFirst:
			// The only form supported in GMS for now
		case tree.NullsLast:
			return nil, errors.Errorf("NULLS LAST for indexes is not yet supported")
		default:
			return nil, errors.Errorf("unknown NULL ordering for index")
		}

		vitessIndexColumns = append(vitessIndexColumns, &vitess.IndexColumn{
			Column: vitess.NewColIdent(string(inputColumn.Column)),
			Order:  vitess.AscScr,
		})
	}

	return vitessIndexColumns, nil
}
