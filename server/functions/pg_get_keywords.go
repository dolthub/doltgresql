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
	"io"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/lex"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPgGetKeywords registers the functions to the catalog.
func initPgGetKeywords() {
	framework.RegisterFunction(pg_get_keywords)
}

// pgGetKeywordsName is the name for pg_get_keywords function.
const pgGetKeywordsName = "pg_get_keywords"

// pg_get_keywords represents the PostgreSQL system catalog information function.
var pg_get_keywords = framework.Function0{
	Name:               pgGetKeywordsName,
	Return:             pgtypes.Text,
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context) (any, error) {
		i := 0
		count := len(lex.KeywordNames)
		return pgtypes.NewSetReturningFunctionRowIter(func(ctx *sql.Context) (sql.Row, error) {
			defer func() { i++ }()

			if i >= count {
				return nil, io.EOF
			}

			keyword := lex.KeywordNames[i]
			catCode := lex.KeywordsCategories[keyword]
			catDesc := ""
			switch catCode {
			case "U":
				catDesc = "unreserved"
			case "R":
				catDesc = "reserved"
			case "T":
				catDesc = "reserved (can be function or type name)"
			case "C":
				catDesc = "unreserved (cannot be function or type name)"
			}
			_, ok := keywordsRequiresAs[keyword]
			bareLabel := !ok
			bareDesc := "can be bare label"
			if !bareLabel {
				bareDesc = "requires AS"
			}

			return sql.Row{
				keyword,   // word
				catCode,   // catcode
				bareLabel, // barelabel
				catDesc,   // catdesc
				bareDesc,  // baredesc
			}, nil
		}), nil
	},
	OutParams: pgGetKeywordsOutArgs,
}

// pgGetKeywordsOutArgs is the schema for pg_get_keywords table function. Each column is OUT argument.
var pgGetKeywordsOutArgs = sql.Schema{
	{Name: "word", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgGetKeywordsName},
	{Name: "catcode", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: pgGetKeywordsName},
	{Name: "barelabel", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: pgGetKeywordsName},
	{Name: "catdesc", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgGetKeywordsName},
	{Name: "baredesc", Type: pgtypes.Text, Default: nil, Nullable: false, Source: pgGetKeywordsName},
}

// keywordsRequiresAs is a list of keywords that requires AS (cannot be bare label).
var keywordsRequiresAs = map[string]struct{}{
	"array":     {},
	"as":        {},
	"char":      {},
	"character": {},
	"create":    {},
	"day":       {},
	"except":    {},
	"fetch":     {},
	"filter":    {},
	"for":       {},
	"from":      {},
	"grant":     {},
	"group":     {},
	"having":    {},
	"hour":      {},
	"intersect": {},
	"into":      {},
	"isnull":    {},
	"limit":     {},
	"minute":    {},
	"month":     {},
	"notnull":   {},
	"offset":    {},
	"on":        {},
	"order":     {},
	"over":      {},
	"overlaps":  {},
	"precision": {},
	"returning": {},
	"second":    {},
	"to":        {},
	"union":     {},
	"varying":   {},
	"where":     {},
	"window":    {},
	"with":      {},
	"within":    {},
	"without":   {},
	"year":      {},
}
