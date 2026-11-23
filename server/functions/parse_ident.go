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

package functions

import (
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/lex"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initParseIdent registers the functions to the catalog.
func initParseIdent() {
	framework.RegisterFunction(parse_ident_text)
	framework.RegisterFunction(parse_ident_text_bool)
}

// parse_ident_text represents the PostgreSQL function with strict mode defaulting to true.
var parse_ident_text = framework.Function1{
	Name:       "parse_ident",
	Return:     pgtypes.TextArray,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		input, err := framework.UnwrapString(ctx, val)
		if err != nil {
			return nil, err
		}
		return parseIdent(input, true)
	},
}

// parse_ident_text_bool represents the PostgreSQL function with an explicit strict mode.
var parse_ident_text_bool = framework.Function2{
	Name:       "parse_ident",
	Return:     pgtypes.TextArray,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Bool},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		input, err := framework.UnwrapString(ctx, val1)
		if err != nil {
			return nil, err
		}
		return parseIdent(input, val2.(bool))
	},
}

// parseIdent splits a qualified identifier without truncating its components. In non-strict mode, trailing input
// after a complete identifier is ignored, but a dot always requires another valid identifier.
func parseIdent(input string, strict bool) ([]any, error) {
	invalidIdentifier := func() ([]any, error) {
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, `string is not a valid identifier: "%s"`, input)
	}
	// Match PostgreSQL's scanner whitespace, which excludes vertical tabs and non-ASCII spaces.
	const whitespace = " \t\n\r\f"
	remaining := strings.TrimLeft(input, whitespace)
	var identifiers []any
	for {
		if len(remaining) == 0 {
			return invalidIdentifier()
		}

		var identifier strings.Builder
		pos := 0
		if remaining[0] == '"' {
			pos++
			for {
				if pos == len(remaining) {
					return invalidIdentifier()
				}
				ch := remaining[pos]
				pos++
				if ch == '"' {
					if pos < len(remaining) && remaining[pos] == '"' {
						// Doubled double quotes represent a literal quote within an identifier.
						pos++
						identifier.WriteByte('"')
						continue
					}
					break
				}
				identifier.WriteByte(ch)
			}
			if identifier.Len() == 0 {
				return invalidIdentifier()
			}
		} else {
			if !lex.IsIdentStart(int(remaining[0])) {
				return invalidIdentifier()
			}
			for pos < len(remaining) && lex.IsIdentMiddle(int(remaining[pos])) {
				ch := remaining[pos]
				// PostgreSQL folds only ASCII uppercase letters in UTF-8 identifiers.
				if ch >= 'A' && ch <= 'Z' {
					ch += 'a' - 'A'
				}
				identifier.WriteByte(ch)
				pos++
			}
		}
		identifiers = append(identifiers, identifier.String())
		remaining = strings.TrimLeft(remaining[pos:], whitespace)
		if len(remaining) == 0 {
			return identifiers, nil
		}
		if remaining[0] != '.' {
			if strict {
				return invalidIdentifier()
			}
			return identifiers, nil
		}
		remaining = strings.TrimLeft(remaining[1:], whitespace)
	}
}
