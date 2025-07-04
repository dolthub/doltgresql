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
	"github.com/dolthub/go-mysql-server/sql/expression"
)

func init() {
	expression.DefaultExpressionFactory = PostgresExpressionFactory{}
}

// PostgresExpressionFactory implements the expression.ExpressionFactory interface and
// allows callers to produce expressions that have custom behavior for Postgres.
type PostgresExpressionFactory struct{}

var _ expression.ExpressionFactory = (*PostgresExpressionFactory)(nil)

// NewIsNull implements the expression.ExpressionFactory interface.
func (m PostgresExpressionFactory) NewIsNull(e sql.Expression) sql.Expression {
	return NewIsNull(e)
}

// NewIsNotNull implements the expression.ExpressionFactory interface.
func (m PostgresExpressionFactory) NewIsNotNull(e sql.Expression) sql.Expression {
	return NewIsNotNull(e)
}
