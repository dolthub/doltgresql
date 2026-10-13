// Copyright 2026 Dolthub, Inc.
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

package binary

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/expression/function"
	"github.com/dolthub/go-mysql-server/sql/types"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initRegexInsensitive registers case-insensitive regex matching and its negation.
func initRegexInsensitive() {
	for _, op := range []struct {
		operator framework.Operator
		name     string
		negate   bool
	}{
		{framework.Operator_BinaryRegexInsensitive, "texticregexeq", false},
		{framework.Operator_BinaryNotRegexInsensitive, "texticregexne", true},
	} {
		negate := op.negate
		framework.RegisterBinaryFunction(op.operator, framework.Function2{
			Name:       op.name,
			Return:     pgtypes.Bool,
			Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
			Strict:     true,
			Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, left, right any) (any, error) {
				text, err := framework.UnwrapString(ctx, left)
				if err != nil {
					return nil, err
				}
				pattern, err := framework.UnwrapString(ctx, right)
				if err != nil {
					return nil, err
				}
				// PostgreSQL's empty regex matches every non-NULL string. The shared
				// MySQL matcher rejects empty patterns, so handle this before compiling.
				if pattern == "" {
					return !negate, nil
				}
				// Reuse the engine used by ~ and !~. PostgreSQL's default is to let
				// dot match newlines; MySQL's "n" flag enables that behavior.
				// TODO: Support PostgreSQL-specific ARE syntax and embedded option flags
				// that differ from the shared regex engine.
				matcher, err := function.NewRegexpLike(ctx,
					expression.NewLiteral(text, types.LongText),
					expression.NewLiteral(pattern, types.LongText),
					expression.NewLiteral("in", types.LongText),
				)
				if err != nil {
					return nil, err
				}
				defer matcher.(sql.Disposable).Dispose(ctx)
				value, err := matcher.Eval(ctx, nil)
				if err != nil {
					return nil, pgerror.WithCandidateCode(err, pgcode.InvalidRegularExpression)
				}
				return (value.(int8) != 0) != negate, nil
			},
		})
	}
}
