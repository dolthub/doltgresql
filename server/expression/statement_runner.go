// Copyright 2025 Dolthub, Inc.
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

package expression

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// StatementRunner is an expression that can be added to a node to grab the statement runner.
type StatementRunner struct {
	Runner analyzer.StatementRunner
}

var _ sql.Expression = StatementRunner{}
var _ analyzer.Interpreter = StatementRunner{}

// Children implements the sql.Expression interface.
func (StatementRunner) Children() []sql.Expression {
	return nil
}

// Eval implements the sql.Expression interface.
func (StatementRunner) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	return nil, nil
}

// IsNullable implements the sql.Expression interface.
func (StatementRunner) IsNullable() bool {
	return false
}

// Resolved implements the sql.Expression interface.
func (StatementRunner) Resolved() bool {
	return true
}

// SetStatementRunner implements the sql.Expression interface.
func (sr StatementRunner) SetStatementRunner(ctx *sql.Context, runner analyzer.StatementRunner) sql.Expression {
	sr.Runner = runner
	return sr
}

// String implements the sql.Expression interface.
func (StatementRunner) String() string {
	return "StatementRunner"
}

// Type implements the sql.Expression interface.
func (StatementRunner) Type() sql.Type {
	return pgtypes.Unknown
}

// WithChildren implements the sql.Expression interface.
func (sr StatementRunner) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(sr, len(children), 0)
	}
	return sr, nil
}
