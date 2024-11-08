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
	"fmt"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeAlterProcedure handles *tree.AlterProcedure nodes.
func nodeAlterProcedure(ctx *Context, node *tree.AlterProcedure) (vitess.Statement, error) {
	err := verifyRedundantRoutineOption(ctx, node.Options)
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("ALTER PROCEDURE statement is not yet supported")
}
