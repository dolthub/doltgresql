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
	"unicode/utf8"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initRegexpReplace registers the functions to the catalog.
func initRegexpReplace() {
	framework.RegisterFunction(regexp_replace_text_text_text)
	framework.RegisterFunction(regexp_replace_text_text_text_text)
	framework.RegisterFunction(regexp_replace_text_text_text_int4)
	framework.RegisterFunction(regexp_replace_text_text_text_int4_int4)
	framework.RegisterFunction(regexp_replace_text_text_text_int4_int4_text)
}

// regexp_replace_text_text_text represents the PostgreSQL function of the same name, taking the same parameters.
var regexp_replace_text_text_text = framework.Function3{
	Name:       "regexp_replace",
	Return:     pgtypes.Text,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, source, pattern, replacement any) (any, error) {
		return regexpReplace(ctx, source, pattern, replacement, "", 1, -1)
	},
}

// regexp_replace_text_text_text_text represents the PostgreSQL function of the same name, taking the same parameters.
var regexp_replace_text_text_text_text = framework.Function4{
	Name:       "regexp_replace",
	Return:     pgtypes.Text,
	Parameters: [4]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text, pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [5]*pgtypes.DoltgresType, source, pattern, replacement, flags any) (any, error) {
		return regexpReplace(ctx, source, pattern, replacement, flags, 1, -1)
	},
}

// regexp_replace_text_text_text_int4 represents the PostgreSQL function of the same name, taking the same parameters.
var regexp_replace_text_text_text_int4 = framework.Function4{
	Name:       "regexp_replace",
	Return:     pgtypes.Text,
	Parameters: [4]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text, pgtypes.Text, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [5]*pgtypes.DoltgresType, source, pattern, replacement, start any) (any, error) {
		return regexpReplace(ctx, source, pattern, replacement, "", start.(int32), -1)
	},
}

// regexp_replace_text_text_text_int4_int4 represents the PostgreSQL function of the same name, taking the same parameters.
var regexp_replace_text_text_text_int4_int4 = framework.Function5{
	Name:       "regexp_replace",
	Return:     pgtypes.Text,
	Parameters: [5]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text, pgtypes.Text, pgtypes.Int32, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [6]*pgtypes.DoltgresType, source, pattern, replacement, start, n any) (any, error) {
		if n.(int32) < 0 {
			return nil, invalidRegexpReplaceN(n.(int32))
		}
		return regexpReplace(ctx, source, pattern, replacement, "", start.(int32), n.(int32))
	},
}

// regexp_replace_text_text_text_int4_int4_text represents the PostgreSQL function of the same name, taking the same parameters.
var regexp_replace_text_text_text_int4_int4_text = framework.Function6{
	Name:       "regexp_replace",
	Return:     pgtypes.Text,
	Parameters: [6]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text, pgtypes.Text, pgtypes.Int32, pgtypes.Int32, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [7]*pgtypes.DoltgresType, source, pattern, replacement, start, n, flags any) (any, error) {
		if n.(int32) < 0 {
			return nil, invalidRegexpReplaceN(n.(int32))
		}
		return regexpReplace(ctx, source, pattern, replacement, flags, start.(int32), n.(int32))
	},
}

// regexpReplace replaces matches of |pattern| in |source| with |replacement|, expanding PostgreSQL's replacement
// escapes. The search begins at the 1-based character position |start|. When |n| is positive, only the n'th match is
// replaced, and when it is zero, every match is. A negative |n| means it was not given, in which case the first match is
// replaced, or every match when the "g" flag is given.
func regexpReplace(ctx *sql.Context, source, pattern, replacement, flags any, start, n int32) (any, error) {
	if start <= 0 {
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, "invalid value for parameter \"start\": %d", start)
	}
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
	if n < 0 {
		if global {
			n = 0
		} else {
			n = 1
		}
	}
	// Convert the character position to a byte offset; a start past the end of the string matches nothing.
	offset := 0
	for i := int32(1); i < start; i++ {
		if offset >= len(sourceText) {
			return sourceText, nil
		}
		_, size := utf8.DecodeRuneInString(sourceText[offset:])
		offset += size
	}
	// TODO: The search runs on the text after |offset|, so anchors and word boundaries treat that position as the
	// start of the string. PostgreSQL matches against the whole string, so ^ cannot match there.
	searchText := sourceText[offset:]
	limit := -1
	if n > 0 {
		limit = int(n)
	}
	matches := re.FindAllStringSubmatchIndex(searchText, limit)
	if n > 0 {
		if len(matches) < int(n) {
			return sourceText, nil
		}
		matches = matches[n-1:]
	}
	if len(matches) == 0 {
		return sourceText, nil
	}
	var result strings.Builder
	result.WriteString(sourceText[:offset])
	previousEnd := 0
	for _, match := range matches {
		result.WriteString(searchText[previousEnd:match[0]])
		appendRegexpReplacement(&result, searchText, replacementText, match)
		previousEnd = match[1]
	}
	result.WriteString(searchText[previousEnd:])
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

// invalidRegexpReplaceN returns the error for a negative match number, which is distinct from the internal "not given"
// value that regexpReplace accepts.
func invalidRegexpReplaceN(n int32) error {
	return pgerror.Newf(pgcode.InvalidParameterValue, "invalid value for parameter \"n\": %d", n)
}
