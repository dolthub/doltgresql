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

package _go

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

func TestParseIdent(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "signatures and array result",
			Assertions: []ScriptTestAssertion{
				{
					Query:            `SELECT parse_ident('public'), parse_ident('"SomeSchema".some_table');`,
					Expected:         []sql.Row{{"{public}", "{SomeSchema,some_table}"}},
					ExpectedColTypes: []id.Type{pgtypes.TextArray.ID, pgtypes.TextArray.ID},
				},
				{
					Query:    `SELECT parse_ident('Schema.Table'::text, true), parse_ident('Schema.Table'::text, false);`,
					Expected: []sql.Row{{"{schema,table}", "{schema,table}"}},
				},
				{
					Query:    `SELECT pg_catalog.parse_ident('public'), pg_catalog.parse_ident('"SomeSchema".some_table', false);`,
					Expected: []sql.Row{{"{public}", "{SomeSchema,some_table}"}},
				},
				{
					Query:    `SELECT to_regprocedure('pg_catalog.parse_ident(text, boolean)')::oid;`,
					Expected: []sql.Row{{1268}},
				},
				{
					Query:    `SELECT parse_ident('Schema.Table'::varchar), parse_ident('public'::name);`,
					Expected: []sql.Row{{"{schema,table}", "{public}"}},
				},
				{
					Query:    `SELECT (parse_ident('"SomeSchema".some_table'))[1], (parse_ident('"SomeSchema".some_table'))[2], (parse_ident('public'))[2];`,
					Expected: []sql.Row{{"SomeSchema", "some_table", nil}},
				},
				{
					Query:    `SELECT pg_typeof(parse_ident('public')), array_length(parse_ident('a.b.c.d.e'), 1);`,
					Expected: []sql.Row{{"text[]", 5}},
				},
				{
					Query:    `SELECT parse_ident($1::text, $2::boolean);`,
					BindVars: []any{`"SomeSchema".SomeTable(integer)`, false},
					Expected: []sql.Row{{"{SomeSchema,sometable}"}},
				},
			},
		},
		{
			Name: "quoted identifiers",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT parse_ident('"SomeSchema".someTable'), parse_ident('"SomeSchema"."SomeTable"');`,
					Expected: []sql.Row{{"{SomeSchema,sometable}", "{SomeSchema,SomeTable}"}},
				},
				{
					Query:    `SELECT (parse_ident('"schema.with.dots"."table name"'))[1], (parse_ident('"schema.with.dots"."table name"'))[2];`,
					Expected: []sql.Row{{"schema.with.dots", "table name"}},
				},
				{
					Query:    `SELECT (parse_ident('"a""b"."c""d"'))[1], (parse_ident('"a""b"."c""d"'))[2];`,
					Expected: []sql.Row{{`a"b`, `c"d`}},
				},
				{
					Query:    `SELECT (parse_ident('"""foo"""'))[1], (parse_ident('""""'))[1];`,
					Expected: []sql.Row{{`"foo"`, `"`}},
				},
				{
					Query:    `SELECT (parse_ident('"  spaced  "'))[1], (parse_ident('"123"'))[1], (parse_ident('"char"'))[1];`,
					Expected: []sql.Row{{"  spaced  ", "123", "char"}},
				},
				{
					Query:    `SELECT (parse_ident('"comma,brace{and}slash\"'))[1], (parse_ident('"O''Brien"'))[1];`,
					Expected: []sql.Row{{`comma,brace{and}slash\`, "O'Brien"}},
				},
				{
					Query:    `SELECT (parse_ident(E'"line\nnext"'))[1];`,
					Expected: []sql.Row{{"line\nnext"}},
				},
				{
					Query:    `SELECT parse_ident('"NULL"."a,b"');`,
					Expected: []sql.Row{{`{"NULL","a,b"}`}},
				},
			},
		},
		{
			Name: "unquoted identifiers and whitespace",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT parse_ident('_Schema9.Table$1'), parse_ident('select.from');`,
					Expected: []sql.Row{{"{_schema9,table$1}", "{select,from}"}},
				},
				{
					Query:    `SELECT parse_ident('  First . "  Second  " . Third  ');`,
					Expected: []sql.Row{{`{first,"  Second  ",third}`}},
				},
				{
					Query:    `SELECT parse_ident(E' \t\n\r\fFirst\t.\nSecond\r\f ');`,
					Expected: []sql.Row{{"{first,second}"}},
				},
				{
					// PostgreSQL folds ASCII letters, preserving non-ASCII case in UTF-8.
					Query:    `SELECT parse_ident('ÄBC.ÖDEF'), parse_ident('日本語.テーブル'), parse_ident('🐘.TABLE');`,
					Expected: []sql.Row{{"{Äbc,Ödef}", "{日本語,テーブル}", "{🐘,table}"}},
				},
				{
					// Identifier parsing does not perform Unicode normalization.
					Query:    "SELECT (parse_ident('E\u0301COLE'))[1];",
					Expected: []sql.Row{{"e\u0301cole"}},
				},
				{
					// Non-breaking spaces are identifier characters, not scanner whitespace.
					Query:    "SELECT (parse_ident('\u00a0FOO\u00a0'))[1];",
					Expected: []sql.Row{{"\u00a0foo\u00a0"}},
				},
			},
		},
		{
			Name: "non-strict trailing input",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT parse_ident('Schema.Function(integer, text)', false);`,
					Expected: []sql.Row{{"{schema,function}"}},
				},
				{
					Query:    `SELECT parse_ident('foo.boo[]', false), parse_ident('aaa.a%b', false);`,
					Expected: []sql.Row{{"{foo,boo}", "{aaa,a}"}},
				},
				{
					Query:    `SELECT parse_ident('foo bar.baz', false), parse_ident('foo;bar', false);`,
					Expected: []sql.Row{{"{foo}", "{foo}"}},
				},
				{
					Query:    `SELECT parse_ident('"Foo"bar', false), parse_ident('foo"bar', false);`,
					Expected: []sql.Row{{"{Foo}", "{foo}"}},
				},
				{
					Query:    `SELECT parse_ident('"Foo"."Bar" (integer)', false);`,
					Expected: []sql.Row{{"{Foo,Bar}"}},
				},
				{
					Query:    `SELECT parse_ident(E'foo\013bar', false), parse_ident('foo/*comment*/.bar', false);`,
					Expected: []sql.Row{{"{foo}", "{foo}"}},
				},
			},
		},
		{
			Name: "long identifiers",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT (parse_ident(repeat('A', 100)))[1];`,
					Expected: []sql.Row{{strings.Repeat("a", 100)}},
				},
				{
					Query:    `SELECT length((parse_ident('"' || repeat('X', 100) || '".' || repeat('Y', 100)))[1]), length((parse_ident('"' || repeat('X', 100) || '".' || repeat('Y', 100)))[2]);`,
					Expected: []sql.Row{{100, 100}},
				},
				{
					Query:    `SELECT length((parse_ident(repeat('Ä', 100)))[1]), octet_length((parse_ident(repeat('Ä', 100)))[1]);`,
					Expected: []sql.Row{{100, 200}},
				},
			},
		},
		{
			Name: "null arguments",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT parse_ident(NULL), parse_ident(NULL::text);`,
					Expected: []sql.Row{{nil, nil}},
				},
				{
					Query:    `SELECT parse_ident(NULL, true), parse_ident(NULL, false), parse_ident(NULL, NULL);`,
					Expected: []sql.Row{{nil, nil, nil}},
				},
				{
					Query:    `SELECT parse_ident('public', NULL::boolean), parse_ident('invalid..name', NULL::boolean);`,
					Expected: []sql.Row{{nil, nil}},
				},
			},
		},
		{
			Name: "column arguments",
			SetUpScript: []string{
				`CREATE TABLE parse_ident_inputs (id integer PRIMARY KEY, input text, strict_mode boolean);`,
				`INSERT INTO parse_ident_inputs VALUES (1, 'PUBLIC', true), (2, '"SomeSchema".SomeTable', false), (3, NULL, true), (4, 'invalid..name', NULL);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT id, parse_ident(input, strict_mode) FROM parse_ident_inputs ORDER BY id;`,
					Expected: []sql.Row{{1, "{public}"}, {2, "{SomeSchema,sometable}"}, {3, nil}, {4, nil}},
				},
				{
					Query:    `SELECT id, parse_ident(input) FROM parse_ident_inputs WHERE id < 4 ORDER BY id;`,
					Expected: []sql.Row{{1, "{public}"}, {2, "{SomeSchema,sometable}"}, {3, nil}},
				},
			},
		},
		{
			Name: "PostgREST computed relationship schema",
			SetUpScript: []string{
				`CREATE FUNCTION public.parse_ident_computed_rel(integer) RETURNS integer LANGUAGE SQL AS 'SELECT $1';`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT (parse_ident(p.pronamespace::regnamespace::text))[1] AS schema FROM pg_catalog.pg_proc p WHERE p.proname = 'parse_ident_computed_rel';`,
					Expected: []sql.Row{{"public"}},
				},
			},
		},
	})
}

func TestParseIdentInvalidInput(t *testing.T) {
	// These inputs must fail even in non-strict mode: no valid identifier was
	// found, or a dot requires another identifier, or the quoted part is invalid.
	invalidInputs := []string{
		`''`,
		`' '`,
		`E'\t\n\r\f'`,
		`'.'`,
		`'.foo'`,
		`'foo.'`,
		`'foo . '`,
		`'foo..bar'`,
		`'foo.123'`,
		`'123'`,
		`'$foo'`,
		`'""'`,
		`'"unterminated'`,
		`'foo."unterminated'`,
		`'foo.""'`,
		`E'\013foo'`,
		`'/*comment*/foo'`,
	}
	var assertions []ScriptTestAssertion
	for _, input := range invalidInputs {
		for _, mode := range []string{"", ", true", ", false"} {
			assertions = append(assertions, ScriptTestAssertion{
				Query:           fmt.Sprintf("SELECT parse_ident(%s%s);", input, mode),
				ExpectedErr:     "string is not a valid identifier",
				ExpectedErrCode: "22023",
			})
		}
	}
	RunScripts(t, []ScriptTest{{Name: "invalid identifiers in either mode", Assertions: assertions}})
}

func TestParseIdentStrictInput(t *testing.T) {
	// Both the default and explicit strict mode reject characters left after the
	// final identifier. Non-strict counterparts are covered in TestParseIdent.
	trailingInputs := []string{
		`'Schema.Function(integer, text)'`,
		`'foo.boo[]'`,
		`'aaa.a%b'`,
		`'foo bar.baz'`,
		`'foo;bar'`,
		`'"Foo"bar'`,
		`'foo"bar'`,
		`'"Foo"."Bar" (integer)'`,
		`E'foo\013bar'`,
		`'foo/*comment*/.bar'`,
	}
	var assertions []ScriptTestAssertion
	for _, input := range trailingInputs {
		for _, mode := range []string{"", ", true"} {
			assertions = append(assertions, ScriptTestAssertion{
				Query:           fmt.Sprintf("SELECT parse_ident(%s%s);", input, mode),
				ExpectedErr:     "string is not a valid identifier",
				ExpectedErrCode: "22023",
			})
		}
	}
	RunScripts(t, []ScriptTest{{Name: "strict trailing input", Assertions: assertions}})
}
