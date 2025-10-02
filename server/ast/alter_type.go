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

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeAlterType handles *tree.AlterType nodes.
func nodeAlterType(ctx *Context, node *tree.AlterType) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}

	// We can handle the common ALTER TYPE .. TO OWNER case since it's a no-op
	if _, ok := node.Cmd.(*tree.AlterTypeOwner); ok {
		return NewNoOp("owners are unsupported"), nil
	}

	return NotYetSupportedError("ALTER TYPE is not yet supported")
}
