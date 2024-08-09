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

	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeDiscard handles *tree.Discard nodes.
func nodeDiscard(node *tree.Discard) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	if node.Mode != tree.DiscardModeAll {
		return nil, fmt.Errorf("unhandled DISCARD mode: %v", node.Mode)
	}

	return vitess.InjectedStatement{
		Statement: DiscardStatement{},
	}, nil
}

// DiscardStatement is just a marker type, since all functionality is handled by the connection handler,
// rather than the engine. It has to conform to the sql.ExecSourceRel interface to be used in the handler, but this
// functionality is all unused.
type DiscardStatement struct{}

var _ vitess.Injectable = DiscardStatement{}
var _ sql.ExecSourceRel = DiscardStatement{}

func (d DiscardStatement) Resolved() bool {
	return true
}

func (d DiscardStatement) String() string {
	return "DISCARD ALL"
}

func (d DiscardStatement) Schema() sql.Schema {
	return nil
}

func (d DiscardStatement) Children() []sql.Node {
	return nil
}

func (d DiscardStatement) WithChildren(children ...sql.Node) (sql.Node, error) {
	return d, nil
}

func (d DiscardStatement) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
	return true
}

func (d DiscardStatement) IsReadOnly() bool {
	return true
}

func (d DiscardStatement) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	panic("DISCARD ALL should be handled by the connection handler")
}

func (d DiscardStatement) WithResolvedChildren(children []any) (any, error) {
	return d, nil
}
