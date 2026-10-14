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
	"regexp"
	"regexp/syntax"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initRegexpSplitToArray registers both PostgreSQL overloads.
func initRegexpSplitToArray() {
	framework.RegisterFunction(regexp_split_to_array_text_text)
	framework.RegisterFunction(regexp_split_to_array_text_text_text)
}

// regexp_split_to_array_text_text uses PostgreSQL's default regex options.
var regexp_split_to_array_text_text = framework.Function2{
	Name:       "regexp_split_to_array",
	Return:     pgtypes.TextArray,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, input, pattern any) (any, error) {
		return regexpSplitToArray(ctx, input, pattern, "")
	},
}

// regexp_split_to_array_text_text_text accepts explicit regex options.
var regexp_split_to_array_text_text_text = framework.Function3{
	Name:       "regexp_split_to_array",
	Return:     pgtypes.TextArray,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, input, pattern, flags any) (any, error) {
		return regexpSplitToArray(ctx, input, pattern, flags)
	},
}

// regexpSplitToArray omits boundary empty matches and empty matches immediately
// following a delimiter, while preserving empty fields from nonempty delimiters.
func regexpSplitToArray(ctx *sql.Context, input, pattern, flags any) (any, error) {
	text, err := framework.UnwrapString(ctx, input)
	if err != nil {
		return nil, err
	}
	patternText, err := framework.UnwrapString(ctx, pattern)
	if err != nil {
		return nil, err
	}
	flagText, err := framework.UnwrapString(ctx, flags)
	if err != nil {
		return nil, err
	}
	re, err := compileSplitRegexp(patternText, flagText)
	if err != nil {
		return nil, err
	}
	result := make([]any, 0)
	previousEnd := 0
	for _, match := range re.FindAllStringIndex(text, -1) {
		if match[0] == match[1] && (match[0] == previousEnd || match[0] == len(text)) {
			continue
		}
		result = append(result, text[previousEnd:match[0]])
		previousEnd = match[1]
	}
	return append(result, text[previousEnd:]), nil
}

// compileSplitRegexp translates PostgreSQL options into Go regex parser flags.
// TODO: Support PostgreSQL ARE features absent from Go's regex engine, including
// backreferences, lookarounds, and locale-dependent character classes.
func compileSplitRegexp(pattern, options string) (*regexp.Regexp, error) {
	flags := syntax.Perl | syntax.DotNL
	expanded := false
	mode := byte('a') // advanced, extended, basic, or quoted
	apply := func(options string) error {
		for _, option := range options {
			switch option {
			case 'b', 'e', 'q':
				mode = byte(option)
			case 'c':
				flags &^= syntax.FoldCase
			case 'i':
				flags |= syntax.FoldCase
			case 'm', 'n':
				flags &^= syntax.DotNL | syntax.ClassNL | syntax.OneLine
			case 'p':
				flags &^= syntax.DotNL | syntax.ClassNL
				flags |= syntax.OneLine
			case 's':
				flags |= syntax.DotNL | syntax.ClassNL | syntax.OneLine
			case 'w':
				flags |= syntax.DotNL | syntax.ClassNL
				flags &^= syntax.OneLine
			case 't':
				expanded = false
			case 'x':
				expanded = true
			case 'g':
				// PostgreSQL rejects this flag for split functions, which already
				// process all matches; it is not accepted as a no-op.
				return pgerror.New(pgcode.InvalidParameterValue, "regexp_split_to_array() does not support the \"global\" option")
			default:
				return pgerror.Newf(pgcode.InvalidParameterValue, "invalid regular expression option: %q", string(option))
			}
		}
		return nil
	}
	if err := apply(options); err != nil {
		return nil, err
	}
	// Quoted mode treats the entire pattern literally, including directors.
	// Otherwise, directors and initial embedded options can override the flags.
	if mode != 'q' {
		if strings.HasPrefix(pattern, "***=") {
			mode, pattern = 'q', pattern[4:]
		} else if strings.HasPrefix(pattern, "***:") {
			mode, pattern = 'a', pattern[4:]
		}
		if mode == 'a' && strings.HasPrefix(pattern, "(?") {
			if end := strings.IndexByte(pattern, ')'); end > 2 {
				embedded := pattern[2:end]
				if strings.IndexFunc(embedded, func(r rune) bool { return r < 'a' || r > 'z' }) == -1 {
					if err := apply(embedded); err != nil {
						return nil, err
					}
					pattern = pattern[end+1:]
				}
			}
		}
	}
	if mode == 'q' {
		flags |= syntax.Literal
	} else {
		if expanded {
			pattern = compactExpandedRegexp(pattern)
		}
		if mode == 'b' {
			pattern = expandBasicRegexp(pattern)
		}
		if mode == 'b' || mode == 'e' {
			flags &^= syntax.PerlX | syntax.UnicodeGroups
		}
	}
	parsed, err := syntax.Parse(pattern, flags)
	if err != nil {
		return nil, pgerror.Newf(pgcode.InvalidRegularExpression, "invalid regular expression: %s", err)
	}
	re, err := regexp.Compile(parsed.String())
	if err != nil {
		return nil, pgerror.Newf(pgcode.InvalidRegularExpression, "invalid regular expression: %s", err)
	}
	// POSIX matches choose the longest alternative unless a quantifier requests
	// a shortest match. Submatch choices do not affect the returned fields.
	if !hasNonGreedyRegexp(parsed) {
		re.Longest()
	}
	return re, nil
}

// hasNonGreedyRegexp recognizes shortest-match quantifiers in the parsed regex.
func hasNonGreedyRegexp(re *syntax.Regexp) bool {
	if re.Flags&syntax.NonGreedy != 0 {
		return true
	}
	for _, sub := range re.Sub {
		if hasNonGreedyRegexp(sub) {
			return true
		}
	}
	return false
}

// compactExpandedRegexp removes unescaped whitespace and comments outside
// bracket expressions, preserving escaped characters and POSIX character classes.
func compactExpandedRegexp(pattern string) string {
	var result strings.Builder
	inClass := false
	classStart := 0
	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		if c == '\\' && i+1 < len(pattern) {
			result.WriteByte(c)
			i++
			result.WriteByte(pattern[i])
			continue
		}
		if inClass {
			if c == '[' && i+1 < len(pattern) && strings.ContainsRune(":.=", rune(pattern[i+1])) {
				end := strings.Index(pattern[i+2:], string(pattern[i+1])+"]")
				if end >= 0 {
					end += i + 4
					result.WriteString(pattern[i:end])
					i = end - 1
					continue
				}
			}
			if c == ']' && i > classStart {
				inClass = false
			}
		} else {
			if strings.ContainsRune(" \t\n\r\f\v", rune(c)) {
				continue
			}
			if c == '#' {
				for i+1 < len(pattern) && pattern[i+1] != '\n' {
					i++
				}
				continue
			}
			if c == '[' {
				inClass = true
				classStart = i + 1
				if i+1 < len(pattern) && pattern[i+1] == '^' {
					classStart++
				}
			}
		}
		result.WriteByte(c)
	}
	return result.String()
}

// expandBasicRegexp converts BRE grouping and bounds into ERE spelling while
// quoting metacharacters that are ordinary in PostgreSQL basic regex syntax.
// Escaped +, ?, and | remain literals too; PostgreSQL does not use the GNU BRE
// extensions that interpret them as quantifiers or alternation.
func expandBasicRegexp(pattern string) string {
	var result strings.Builder
	inClass := false
	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		if c == '\\' && i+1 < len(pattern) {
			i++
			if !inClass && strings.ContainsRune("(){}", rune(pattern[i])) {
				result.WriteByte(pattern[i])
			} else {
				result.WriteByte('\\')
				result.WriteByte(pattern[i])
			}
			continue
		}
		if c == '[' {
			inClass = true
		}
		if !inClass && strings.ContainsRune("+?|(){}", rune(c)) {
			result.WriteByte('\\')
		}
		result.WriteByte(c)
		if c == ']' {
			inClass = false
		}
	}
	return result.String()
}
