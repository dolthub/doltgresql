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
	"github.com/cockroachdb/errors"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeCreateProcedure handles *tree.CreateProcedure nodes.
func nodeCreateProcedure(ctx *Context, node *tree.CreateProcedure) (vitess.Statement, error) {
	err := verifyRedundantRoutineOption(ctx, node.Options)
	if err != nil {
		return nil, err
	}
	return nil, errors.Errorf("CREATE PROCEDURE statement is not yet supported")
}
