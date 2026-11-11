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
	"regexp"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initILike registers case-insensitive LIKE (ILIKE, ~~*) and its negation (NOT ILIKE, !~~*).
func initILike() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryILike, texticlike)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotILike, texticnlike)
}

// texticlike represents the PostgreSQL function of the same name, taking the same parameters.
var texticlike = framework.Function2{
	Name:       "texticlike",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, left, right any) (any, error) {
		return iLike(ctx, left, right, false)
	},
}

// texticnlike represents the PostgreSQL function of the same name, taking the same parameters.
var texticnlike = framework.Function2{
	Name:       "texticnlike",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, left, right any) (any, error) {
		return iLike(ctx, left, right, true)
	},
}

// iLike matches text against a LIKE pattern ignoring case, optionally negating the result.
func iLike(ctx *sql.Context, left, right any, negate bool) (any, error) {
	text, err := framework.UnwrapString(ctx, left)
	if err != nil {
		return nil, err
	}
	pattern, err := framework.UnwrapString(ctx, right)
	if err != nil {
		return nil, err
	}
	matcher, err := likeRegexp(pattern)
	if err != nil {
		return nil, err
	}
	return matcher.MatchString(text) != negate, nil
}

// likeRegexp translates a LIKE pattern into an anchored, case-insensitive regular expression.
// As in PostgreSQL, % matches any sequence of characters, _ matches any single character
// (newlines included), and a backslash makes the next character match literally.
func likeRegexp(pattern string) (*regexp.Regexp, error) {
	var b strings.Builder
	b.WriteString(`(?is)\A`)
	runes := []rune(pattern)
	for i := 0; i < len(runes); i++ {
		switch r := runes[i]; r {
		case '%':
			b.WriteString(`.*`)
		case '_':
			b.WriteString(`.`)
		case '\\':
			if i+1 == len(runes) {
				return nil, pgerror.New(pgcode.InvalidEscapeSequence, "LIKE pattern must not end with escape character")
			}
			i++
			b.WriteString(regexp.QuoteMeta(string(runes[i])))
		default:
			b.WriteString(regexp.QuoteMeta(string(r)))
		}
	}
	b.WriteString(`\z`)
	return regexp.Compile(b.String())
}
