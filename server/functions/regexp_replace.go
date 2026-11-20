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

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initRegexpReplace registers the functions to the catalog.
func initRegexpReplace() {
	framework.RegisterFunction(regexp_replace_text_text_text)
	framework.RegisterFunction(regexp_replace_text_text_text_text)
}

// regexp_replace_text_text_text represents the PostgreSQL function of the same name, taking the same parameters.
var regexp_replace_text_text_text = framework.Function3{
	Name:       "regexp_replace",
	Return:     pgtypes.Text,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, source, pattern, replacement any) (any, error) {
		return regexpReplace(ctx, source, pattern, replacement, "")
	},
}

// regexp_replace_text_text_text_text represents the PostgreSQL function of the same name, taking the same parameters.
var regexp_replace_text_text_text_text = framework.Function4{
	Name:       "regexp_replace",
	Return:     pgtypes.Text,
	Parameters: [4]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text, pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [5]*pgtypes.DoltgresType, source, pattern, replacement, flags any) (any, error) {
		return regexpReplace(ctx, source, pattern, replacement, flags)
	},
}

// regexpReplace replaces the first match of |pattern| in |source| (or every match when the "g" flag is given) with
// |replacement|, expanding PostgreSQL's replacement escapes.
func regexpReplace(ctx *sql.Context, source, pattern, replacement, flags any) (any, error) {
	sourceText, err := framework.UnwrapString(ctx, source)
	if err != nil {
		return nil, err
	}
	patternText, err := framework.UnwrapString(ctx, pattern)
	if err != nil {
		return nil, err
	}
	replacementText, err := framework.UnwrapString(ctx, replacement)
	if err != nil {
		return nil, err
	}
	flagText, err := framework.UnwrapString(ctx, flags)
	if err != nil {
		return nil, err
	}
	global := false
	re, err := compilePostgresRegexp(patternText, flagText, &global)
	if err != nil {
		return nil, err
	}
	limit := 1
	if global {
		limit = -1
	}
	matches := re.FindAllStringSubmatchIndex(sourceText, limit)
	if len(matches) == 0 {
		return sourceText, nil
	}
	var result strings.Builder
	previousEnd := 0
	for _, match := range matches {
		result.WriteString(sourceText[previousEnd:match[0]])
		appendRegexpReplacement(&result, sourceText, replacementText, match)
		previousEnd = match[1]
	}
	result.WriteString(sourceText[previousEnd:])
	return result.String(), nil
}

// appendRegexpReplacement writes |replacement| to |result|, substituting PostgreSQL's escapes: \1 through \9 insert
// the corresponding capture group (empty if it did not participate or does not exist), \& inserts the whole match,
// and \\ inserts a single backslash. A backslash followed by anything else is written literally.
func appendRegexpReplacement(result *strings.Builder, source, replacement string, match []int) {
	for i := 0; i < len(replacement); i++ {
		c := replacement[i]
		if c != '\\' || i+1 >= len(replacement) {
			result.WriteByte(c)
			continue
		}
		next := replacement[i+1]
		switch {
		case next >= '1' && next <= '9':
			group := int(next - '0')
			if 2*group+1 < len(match) && match[2*group] >= 0 {
				result.WriteString(source[match[2*group]:match[2*group+1]])
			}
			i++
		case next == '&':
			result.WriteString(source[match[0]:match[1]])
			i++
		case next == '\\':
			result.WriteByte('\\')
			i++
		default:
			result.WriteByte('\\')
		}
	}
}
