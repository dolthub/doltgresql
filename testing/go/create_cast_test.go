/// Copyright 2026 Dolthub, Inc.
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
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestCasts(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "CREATE CAST with function creates explicit-only cast",
			SetUpScript: []string{
				`CREATE TABLE cast_explicit_src (v text);`,
				`CREATE TABLE cast_explicit_dst (v text, tag text);`,
				`CREATE FUNCTION cast_explicit_fn(src cast_explicit_src) RETURNS cast_explicit_dst
					AS $$ SELECT ROW((src).v, 'explicit')::cast_explicit_dst $$ LANGUAGE SQL;`,
				`CREATE FUNCTION cast_explicit_accept(dst cast_explicit_dst) RETURNS text
					AS $$ SELECT (dst).v || ':' || (dst).tag $$ LANGUAGE SQL;`,
				`CREATE TABLE cast_explicit_holder (v cast_explicit_dst);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE CAST (cast_explicit_src AS cast_explicit_dst) WITH FUNCTION cast_explicit_fn(cast_explicit_src);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT cast_explicit_accept((ROW('one')::cast_explicit_src)::cast_explicit_dst);`,
					Expected: []sql.Row{{"one:explicit"}},
				},
				{
					Query:       `SELECT cast_explicit_accept(ROW('one')::cast_explicit_src);`,
					ExpectedErr: "does not exist",
				},
				{
					Query:       `INSERT INTO cast_explicit_holder VALUES (ROW('one')::cast_explicit_src);`,
					ExpectedErr: "cast_explicit_src",
				},
				{
					Query: `SELECT c.castcontext::text, c.castmethod::text
						FROM pg_cast c
						JOIN pg_type src ON src.oid = c.castsource
						JOIN pg_type dst ON dst.oid = c.casttarget
						WHERE src.typname = 'cast_explicit_src'
						  AND dst.typname = 'cast_explicit_dst';`,
					Expected: []sql.Row{{"e", "f"}},
				},
				{
					Query:       `CREATE CAST (cast_explicit_src AS cast_explicit_dst) WITH FUNCTION cast_explicit_fn(cast_explicit_src);`,
					ExpectedErr: "already exists",
				},
			},
		},
		{
			Name: "CREATE CAST with PL/pgSQL function creates explicit-only cast",
			SetUpScript: []string{
				`CREATE TABLE cast_explicit_src (v text);`,
				`CREATE TABLE cast_explicit_dst (v text, tag text);`,
				`CREATE FUNCTION cast_explicit_fn(src cast_explicit_src) RETURNS cast_explicit_dst AS $$ BEGIN
						RETURN ROW((src).v, 'explicit')::cast_explicit_dst;
					END; $$ LANGUAGE plpgsql;`,
				`CREATE FUNCTION cast_explicit_accept(dst cast_explicit_dst) RETURNS text AS $$ BEGIN
						RETURN (dst).v || ':' || (dst).tag;
					END; $$ LANGUAGE plpgsql;`,
				`CREATE TABLE cast_explicit_holder (v cast_explicit_dst);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE CAST (cast_explicit_src AS cast_explicit_dst) WITH FUNCTION cast_explicit_fn(cast_explicit_src);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT cast_explicit_accept((ROW('one')::cast_explicit_src)::cast_explicit_dst);`,
					Expected: []sql.Row{{"one:explicit"}},
				},
				{
					Query:       `SELECT cast_explicit_accept(ROW('one')::cast_explicit_src);`,
					ExpectedErr: "does not exist",
				},
				{
					Query:       `INSERT INTO cast_explicit_holder VALUES (ROW('one')::cast_explicit_src);`,
					ExpectedErr: "cast_explicit_src",
				},
				{
					Query: `SELECT c.castcontext::text, c.castmethod::text
						FROM pg_cast c
						JOIN pg_type src ON src.oid = c.castsource
						JOIN pg_type dst ON dst.oid = c.casttarget
						WHERE src.typname = 'cast_explicit_src'
						  AND dst.typname = 'cast_explicit_dst';`,
					Expected: []sql.Row{{"e", "f"}},
				},
				{
					Query:       `CREATE CAST (cast_explicit_src AS cast_explicit_dst) WITH FUNCTION cast_explicit_fn(cast_explicit_src);`,
					ExpectedErr: "already exists",
				},
			},
		},
		{
			Name: "CREATE CAST AS ASSIGNMENT works for assignment contexts only",
			SetUpScript: []string{
				`CREATE TABLE cast_assignment_src (v text);`,
				`CREATE TABLE cast_assignment_dst (v text, tag text);`,
				`CREATE FUNCTION cast_assignment_fn(cast_assignment_src) RETURNS cast_assignment_dst
					AS $$ SELECT ROW(($1).v, 'assignment')::cast_assignment_dst $$ LANGUAGE SQL;`,
				`CREATE FUNCTION cast_assignment_accept(cast_assignment_dst) RETURNS text
					AS $$ SELECT ($1).v || ':' || ($1).tag $$ LANGUAGE SQL;`,
				`CREATE TABLE cast_assignment_holder (v cast_assignment_dst);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE CAST (cast_assignment_src AS cast_assignment_dst) WITH FUNCTION cast_assignment_fn(cast_assignment_src) AS ASSIGNMENT;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT cast_assignment_accept((ROW('on')::cast_assignment_src)::cast_assignment_dst);`,
					Expected: []sql.Row{{"on:assignment"}},
				},
				{
					Query:    `INSERT INTO cast_assignment_holder VALUES (ROW('on')::cast_assignment_src);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT cast_assignment_accept(v) FROM cast_assignment_holder;`,
					Expected: []sql.Row{{"on:assignment"}},
				},
				{
					Query:       `SELECT cast_assignment_accept(ROW('off')::cast_assignment_src);`,
					ExpectedErr: "does not exist",
				},
				{
					Query: `SELECT c.castcontext::text, c.castmethod::text
						FROM pg_cast c
						JOIN pg_type src ON src.oid = c.castsource
						JOIN pg_type dst ON dst.oid = c.casttarget
						WHERE src.typname = 'cast_assignment_src'
						  AND dst.typname = 'cast_assignment_dst';`,
					Expected: []sql.Row{{"a", "f"}},
				},
			},
		},
		{
			Name: "CREATE CAST AS IMPLICIT works for function resolution",
			SetUpScript: []string{
				`CREATE TABLE cast_implicit_src (v text);`,
				`CREATE TABLE cast_implicit_dst (v text, tag text);`,
				`CREATE FUNCTION cast_implicit_fn(cast_implicit_src) RETURNS cast_implicit_dst
					AS $$ SELECT ROW(($1).v, 'implicit')::cast_implicit_dst $$ LANGUAGE SQL;`,
				`CREATE FUNCTION cast_implicit_accept(cast_implicit_dst) RETURNS text
					AS $$ SELECT ($1).v || ':' || ($1).tag $$ LANGUAGE SQL;`,
				`CREATE TABLE cast_implicit_holder (v cast_implicit_dst);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE CAST (cast_implicit_src AS cast_implicit_dst) WITH FUNCTION cast_implicit_fn(cast_implicit_src) AS IMPLICIT;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT cast_implicit_accept((ROW('x')::cast_implicit_src)::cast_implicit_dst);`,
					Expected: []sql.Row{{"x:implicit"}},
				},
				{
					Query:    `SELECT cast_implicit_accept(ROW('y')::cast_implicit_src);`,
					Expected: []sql.Row{{"y:implicit"}},
				},
				{
					Query:    `INSERT INTO cast_implicit_holder VALUES (ROW('z')::cast_implicit_src);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT cast_implicit_accept(v) FROM cast_implicit_holder;`,
					Expected: []sql.Row{{"z:implicit"}},
				},
				{
					Query: `SELECT c.castcontext::text, c.castmethod::text
						FROM pg_cast c
						JOIN pg_type src ON src.oid = c.castsource
						JOIN pg_type dst ON dst.oid = c.casttarget
						WHERE src.typname = 'cast_implicit_src'
						  AND dst.typname = 'cast_implicit_dst';`,
					Expected: []sql.Row{{"i", "f"}},
				},
			},
		},
		{
			Name: "CREATE CAST WITH INOUT",
			SetUpScript: []string{
				`CREATE TABLE cast_inout_src (v text);`,
				`CREATE TABLE cast_inout_dst (v int);`,
				`CREATE FUNCTION cast_inout_accept(cast_inout_dst) RETURNS text
					AS $$ SELECT (($1).v)::text $$ LANGUAGE SQL;`,
				`CREATE TABLE cast_inout_holder (v cast_inout_dst);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE CAST (cast_inout_src AS cast_inout_dst) WITH INOUT AS ASSIGNMENT;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT cast_inout_accept((ROW('42')::cast_inout_src)::cast_inout_dst);`,
					Expected: []sql.Row{{"42"}},
				},
				{
					Query:    `INSERT INTO cast_inout_holder VALUES (ROW('99')::cast_inout_src);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT cast_inout_accept(v) FROM cast_inout_holder;`,
					Expected: []sql.Row{{"99"}},
				},
				{
					Query:       `SELECT cast_inout_accept((ROW('not_an_int')::cast_inout_src)::cast_inout_dst);`,
					ExpectedErr: "invalid input syntax",
				},
				{
					Query: `SELECT c.castcontext::text, c.castmethod::text
						FROM pg_cast c
						JOIN pg_type src ON src.oid = c.castsource
						JOIN pg_type dst ON dst.oid = c.casttarget
						WHERE src.typname = 'cast_inout_src'
						  AND dst.typname = 'cast_inout_dst';`,
					Expected: []sql.Row{{"a", "i"}},
				},
			},
		},
		{
			Name: "CREATE CAST with three-argument function receives explicit flag",
			SetUpScript: []string{
				`CREATE TABLE cast_three_arg_src (v text);`,
				`CREATE TABLE cast_three_arg_dst (v text, tag text);`,
				`CREATE FUNCTION cast_three_arg_fn(cast_three_arg_src, integer, boolean) RETURNS cast_three_arg_dst
					AS $$
						SELECT ROW(
							($1).v,
							CASE WHEN $3 THEN 'explicit' ELSE 'implicit_or_assignment' END
						)::cast_three_arg_dst
					$$ LANGUAGE SQL;`,
				`CREATE FUNCTION cast_three_arg_accept(cast_three_arg_dst) RETURNS text
					AS $$ SELECT ($1).v || ':' || ($1).tag $$ LANGUAGE SQL;`,
				`CREATE TABLE cast_three_arg_holder (v cast_three_arg_dst);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE CAST (cast_three_arg_src AS cast_three_arg_dst) WITH FUNCTION cast_three_arg_fn(cast_three_arg_src, integer, boolean) AS IMPLICIT;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT cast_three_arg_accept((ROW('a')::cast_three_arg_src)::cast_three_arg_dst);`,
					Expected: []sql.Row{{"a:explicit"}},
				},
				{
					Query:    `SELECT cast_three_arg_accept(ROW('b')::cast_three_arg_src);`,
					Expected: []sql.Row{{"b:implicit_or_assignment"}},
				},
				{
					Query:    `INSERT INTO cast_three_arg_holder VALUES (ROW('c')::cast_three_arg_src);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT cast_three_arg_accept(v) FROM cast_three_arg_holder;`,
					Expected: []sql.Row{{"c:implicit_or_assignment"}},
				},
				{
					Query: `SELECT c.castcontext::text, c.castmethod::text
						FROM pg_cast c
						JOIN pg_type src ON src.oid = c.castsource
						JOIN pg_type dst ON dst.oid = c.casttarget
						WHERE src.typname = 'cast_three_arg_src'
						  AND dst.typname = 'cast_three_arg_dst';`,
					Expected: []sql.Row{{"i", "f"}},
				},
			},
		},
		{
			Name: "DROP CAST removes catalog entry and allows recreation",
			SetUpScript: []string{
				`CREATE TABLE cast_drop_src (v text);`,
				`CREATE TABLE cast_drop_dst (v text, tag text);`,
				`CREATE FUNCTION cast_drop_fn(cast_drop_src) RETURNS cast_drop_dst
					AS $$ SELECT ROW(($1).v, 'drop')::cast_drop_dst $$ LANGUAGE SQL;`,
				`CREATE FUNCTION cast_drop_accept(cast_drop_dst) RETURNS text
					AS $$ SELECT ($1).v || ':' || ($1).tag $$ LANGUAGE SQL;`,
				`CREATE CAST (cast_drop_src AS cast_drop_dst) WITH FUNCTION cast_drop_fn(cast_drop_src);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT cast_drop_accept((ROW('before')::cast_drop_src)::cast_drop_dst);`,
					Expected: []sql.Row{{"before:drop"}},
				},
				{
					Query: `SELECT EXISTS (
						SELECT 1
						FROM pg_cast c
						JOIN pg_type src ON src.oid = c.castsource
						JOIN pg_type dst ON dst.oid = c.casttarget
						WHERE src.typname = 'cast_drop_src'
						  AND dst.typname = 'cast_drop_dst'
					);`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `DROP CAST (cast_drop_src AS cast_drop_dst);`,
					Expected: []sql.Row{},
				},
				{
					Query: `SELECT EXISTS (
						SELECT 1
						FROM pg_cast c
						JOIN pg_type src ON src.oid = c.castsource
						JOIN pg_type dst ON dst.oid = c.casttarget
						WHERE src.typname = 'cast_drop_src'
						  AND dst.typname = 'cast_drop_dst'
					);`,
					Expected: []sql.Row{{"f"}},
				},
				{
					Query:       `SELECT cast_drop_accept((ROW('after')::cast_drop_src)::cast_drop_dst);`,
					ExpectedErr: "cast_drop_src",
				},
				{
					Query:    `CREATE CAST (cast_drop_src AS cast_drop_dst) WITH FUNCTION cast_drop_fn(cast_drop_src);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT cast_drop_accept((ROW('after')::cast_drop_src)::cast_drop_dst);`,
					Expected: []sql.Row{{"after:drop"}},
				},
			},
		},
		{
			Name: "CREATE CAST function is invoked for NULL when function is not STRICT",
			SetUpScript: []string{
				`CREATE TABLE cast_null_src (v text);`,
				`CREATE TABLE cast_null_dst (v text, tag text);`,
				`CREATE FUNCTION cast_null_fn(cast_null_src) RETURNS cast_null_dst
					AS $$ SELECT CASE
						WHEN $1 IS NULL THEN ROW('saw_null', 'called')::cast_null_dst
						ELSE ROW('hi', 'nonnull')::cast_null_dst
					END $$ LANGUAGE SQL;`,
				`CREATE CAST (cast_null_src AS cast_null_dst) WITH FUNCTION cast_null_fn(cast_null_src);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT ((NULL::cast_null_src)::cast_null_dst);`,
					Expected: []sql.Row{{"(saw_null,called)"}},
				},
				{
					Query:    `SELECT ((ROW('hi')::cast_null_src)::cast_null_dst);`,
					Expected: []sql.Row{{"(hi,nonnull)"}},
				},
			},
		},
		{
			Name: "CREATE CAST STRICT function is not invoked for NULL",
			SetUpScript: []string{
				`CREATE TABLE cast_strict_null_src (v text);`,
				`CREATE TABLE cast_strict_null_dst (v text, tag text);`,
				`CREATE FUNCTION cast_strict_null_fn(cast_strict_null_src) RETURNS cast_strict_null_dst
					AS $$ SELECT ROW('bad', 'called')::cast_strict_null_dst $$ LANGUAGE SQL STRICT;`,
				`CREATE CAST (cast_strict_null_src AS cast_strict_null_dst) WITH FUNCTION cast_strict_null_fn(cast_strict_null_src);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT ((NULL::cast_strict_null_src)::cast_strict_null_dst) IS NULL;`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:    `SELECT ((ROW('there')::cast_strict_null_src)::cast_strict_null_dst);`,
					Expected: []sql.Row{{"(bad,called)"}},
				},
			},
		},
		{
			Name: "CREATE CAST validation",
			SetUpScript: []string{
				`CREATE TABLE cast_bad_src (v text);`,
				`CREATE TABLE cast_bad_dst (v text, tag text);`,
				`CREATE FUNCTION cast_bad_wrong_return(cast_bad_src) RETURNS text
					AS $$ SELECT ($1).v $$ LANGUAGE SQL;`,
				`CREATE FUNCTION cast_bad_wrong_source(text) RETURNS cast_bad_dst
					AS $$ SELECT ROW($1, 'wrong_source')::cast_bad_dst $$ LANGUAGE SQL;`,
				`CREATE FUNCTION cast_bad_wrong_second(cast_bad_src, text) RETURNS cast_bad_dst
					AS $$ SELECT ROW(($1).v, 'wrong_second')::cast_bad_dst $$ LANGUAGE SQL;`,
				`CREATE FUNCTION cast_bad_wrong_third(cast_bad_src, integer, integer) RETURNS cast_bad_dst
					AS $$ SELECT ROW(($1).v, 'wrong_third')::cast_bad_dst $$ LANGUAGE SQL;`,
				`CREATE FUNCTION cast_bad_too_many(cast_bad_src, integer, boolean, integer) RETURNS cast_bad_dst
					AS $$ SELECT ROW(($1).v, 'too_many')::cast_bad_dst $$ LANGUAGE SQL;`,
				`CREATE FUNCTION cast_good(cast_bad_src) RETURNS cast_bad_dst
					AS $$ SELECT ROW(($1).v, 'good')::cast_bad_dst $$ LANGUAGE SQL;`,
				`CREATE TABLE cast_without_src (v text);`,
				`CREATE TABLE cast_without_dst (v text, tag text);`,
				`CREATE CAST (cast_bad_src AS cast_bad_dst) WITH FUNCTION cast_good(cast_bad_src);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `CREATE CAST (cast_bad_src AS cast_bad_dst) WITH FUNCTION cast_bad_missing(cast_bad_src);`,
					ExpectedErr: "does not exist",
				},
				{
					Query:       `CREATE CAST (cast_bad_src AS cast_bad_dst) WITH FUNCTION cast_bad_wrong_return(cast_bad_src);`,
					ExpectedErr: "return data type",
				},
				{
					Query:       `CREATE CAST (cast_bad_src AS cast_bad_dst) WITH FUNCTION cast_bad_wrong_source(text);`,
					ExpectedErr: "source data type",
				},
				{
					Query:       `CREATE CAST (cast_bad_src AS cast_bad_dst) WITH FUNCTION cast_bad_wrong_second(cast_bad_src, text);`,
					ExpectedErr: "second argument",
				},
				{
					Query:       `CREATE CAST (cast_bad_src AS cast_bad_dst) WITH FUNCTION cast_bad_wrong_third(cast_bad_src, integer, integer);`,
					ExpectedErr: "third argument",
				},
				{
					Query:       `CREATE CAST (cast_bad_src AS cast_bad_dst) WITH FUNCTION cast_bad_too_many(cast_bad_src, integer, boolean, integer);`,
					ExpectedErr: "one to three arguments",
				},
				{
					Query:       `CREATE CAST (cast_without_src AS cast_without_dst) WITHOUT FUNCTION;`,
					ExpectedErr: "composite data types are not binary-compatible",
				},
				{
					Query:       `CREATE CAST (int4 AS float8) WITHOUT FUNCTION;`,
					ExpectedErr: "source and target data types are not physically compatible",
				},
				{
					Query:       `CREATE CAST (cast_bad_src AS cast_bad_src) WITH INOUT;`,
					ExpectedErr: "source data type and target data type are the same",
				},
				{
					Query:       `CREATE CAST (cast_bad_src AS cast_bad_dst) WITH FUNCTION cast_good(cast_bad_src);`,
					ExpectedErr: "cast from type cast_bad_src to type cast_bad_dst already exists",
				},
			},
		},
	})
}
