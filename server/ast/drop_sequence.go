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

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgnodes "github.com/dolthub/doltgresql/server/node"
)

// nodeDropSequence handles *tree.DropSequence nodes.
func nodeDropSequence(ctx *Context, node *tree.DropSequence) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	if len(node.Names) != 1 {
		return nil, errors.Errorf("dropping multiple sequences in DROP SEQUENCE is not yet supported")
	}
	name, err := nodeTableName(ctx, &node.Names[0])
	if err != nil {
		return nil, err
	}
	if len(name.DbQualifier.String()) > 0 {
		return nil, errors.Errorf("DROP SEQUENCE for the non-current database is not yet supported")
	}
	return vitess.InjectedStatement{
		Statement: pgnodes.NewDropSequence(node.IfExists, name.SchemaQualifier.String(), name.Name.String(),
			node.DropBehavior == tree.DropCascade),
		Children: nil,
	}, nil
}
