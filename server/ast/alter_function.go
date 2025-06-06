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

package ast

import (
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeAlterFunction handles *tree.AlterFunction nodes.
func nodeAlterFunction(ctx *Context, node *tree.AlterFunction) (vitess.Statement, error) {
	_, err := validateRoutineOptions(ctx, node.Options)
	if err != nil {
		return nil, err
	}

	// We can handle the common ALTER FUNCTION .. TO OWNER case since it's a no-op
	if node.Owner != "" {
		return NewNoOp([]string{"owners are unsupported"}), nil
	}

	return NotYetSupportedError("ALTER FUNCTION statement is not yet supported")
}
