// Copyright 2024 Dolthub, Inc.
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

package regression

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestCopy2(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_copy2)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_copy2,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TEMP TABLE x (
	a serial,
	b int,
	c text not null default 'stuff',
	d text,
	e text
) ;`,
			},
			{
				Statement: `CREATE FUNCTION fn_x_before () RETURNS TRIGGER AS '
  BEGIN
		NEW.e := ''before trigger fired''::text;`,
			},
			{
				Statement: `		return NEW;`,
			},
			{
				Statement: `	END;`,
			},
			{
				Statement: `' LANGUAGE plpgsql;`,
			},
			{
				Statement: `CREATE FUNCTION fn_x_after () RETURNS TRIGGER AS '
  BEGIN
		UPDATE x set e=''after trigger fired'' where c=''stuff'';`,
			},
			{
				Statement: `		return NULL;`,
			},
			{
				Statement: `	END;`,
			},
			{
				Statement: `' LANGUAGE plpgsql;`,
			},
			{
				Statement: `CREATE TRIGGER trg_x_after AFTER INSERT ON x
FOR EACH ROW EXECUTE PROCEDURE fn_x_after();`,
			},
			{
				Statement: `CREATE TRIGGER trg_x_before BEFORE INSERT ON x
FOR EACH ROW EXECUTE PROCEDURE fn_x_before();`,
			},
			{
				Statement: `COPY x (a, b, c, d, e) from stdin;`,
			},
			{
				Statement: `COPY x (b, d) from stdin;`,
			},
			{
				Statement: `COPY x (b, d) from stdin;`,
			},
			{
				Statement: `COPY x (a, b, c, d, e) from stdin;`,
			},
			{
				Statement:   `COPY x (xyz) from stdin;`,
				ErrorString: `column "xyz" of relation "x" does not exist`,
			},
			{
				Statement:   `COPY x from stdin (format CSV, FORMAT CSV);`,
				ErrorString: `conflicting or redundant options`,
			},
			{
				Statement:   `COPY x from stdin (freeze off, freeze on);`,
				ErrorString: `conflicting or redundant options`,
			},
			{
				Statement:   `COPY x from stdin (delimiter ',', delimiter ',');`,
				ErrorString: `conflicting or redundant options`,
			},
			{
				Statement:   `COPY x from stdin (null ' ', null ' ');`,
				ErrorString: `conflicting or redundant options`,
			},
			{
				Statement:   `COPY x from stdin (header off, header on);`,
				ErrorString: `conflicting or redundant options`,
			},
			{
				Statement:   `COPY x from stdin (quote ':', quote ':');`,
				ErrorString: `conflicting or redundant options`,
			},
			{
				Statement:   `COPY x from stdin (escape ':', escape ':');`,
				ErrorString: `conflicting or redundant options`,
			},
			{
				Statement:   `COPY x from stdin (force_quote (a), force_quote *);`,
				ErrorString: `conflicting or redundant options`,
			},
			{
				Statement:   `COPY x from stdin (force_not_null (a), force_not_null (b));`,
				ErrorString: `conflicting or redundant options`,
			},
			{
				Statement:   `COPY x from stdin (force_null (a), force_null (b));`,
				ErrorString: `conflicting or redundant options`,
			},
			{
				Statement:   `COPY x from stdin (convert_selectively (a), convert_selectively (b));`,
				ErrorString: `conflicting or redundant options`,
			},
			{
				Statement:   `COPY x from stdin (encoding 'sql_ascii', encoding 'sql_ascii');`,
				ErrorString: `conflicting or redundant options`,
			},
			{
				Statement:   `COPY x (a, b, c, d, e, d, c) from stdin;`,
				ErrorString: `column "d" specified more than once`,
			},
			{
				Statement:   `COPY x from stdin;`,
				ErrorString: `invalid input syntax for type integer: ""`,
			},
			{
				Statement: `CONTEXT:  COPY x, line 1, column a: ""
COPY x from stdin;`,
				ErrorString: `missing data for column "e"`,
			},
			{
				Statement: `CONTEXT:  COPY x, line 1: "2000	230	23	23"
COPY x from stdin;`,
				ErrorString: `missing data for column "e"`,
			},
			{
				Statement: `CONTEXT:  COPY x, line 1: "2001	231	\N	\N"
COPY x from stdin;`,
				ErrorString: `extra data after last expected column`,
			},
			{
				Statement: `CONTEXT:  COPY x, line 1: "2002	232	40	50	60	70	80"
COPY x (b, c, d, e) from stdin delimiter ',' null 'x';`,
			},
			{
				Statement: `COPY x from stdin WITH DELIMITER AS ';' NULL AS '';`,
			},
			{
				Statement: `COPY x from stdin WITH DELIMITER AS ':' NULL AS E'\\X' ENCODING 'sql_ascii';`,
			},
			{
				Statement:   `COPY x TO stdout WHERE a = 1;`,
				ErrorString: `WHERE clause not allowed with COPY TO`,
			},
			{
				Statement: `COPY x from stdin WHERE a = 50004;`,
			},
			{
				Statement: `COPY x from stdin WHERE a > 60003;`,
			},
			{
				Statement:   `COPY x from stdin WHERE f > 60003;`,
				ErrorString: `column "f" does not exist`,
			},
			{
				Statement:   `COPY x from stdin WHERE a = max(x.b);`,
				ErrorString: `aggregate functions are not allowed in COPY FROM WHERE conditions`,
			},
			{
				Statement:   `COPY x from stdin WHERE a IN (SELECT 1 FROM x);`,
				ErrorString: `cannot use subquery in COPY FROM WHERE condition`,
			},
			{
				Statement:   `COPY x from stdin WHERE a IN (generate_series(1,5));`,
				ErrorString: `set-returning functions are not allowed in COPY FROM WHERE conditions`,
			},
			{
				Statement:   `COPY x from stdin WHERE a = row_number() over(b);`,
				ErrorString: `window functions are not allowed in COPY FROM WHERE conditions`,
			},
			{
				Statement: `SELECT * FROM x;`,
				Results:   []sql.Row{{9999, ``, `\N`, `NN`, `before trigger fired`}, {10000, 21, 31, 41, `before trigger fired`}, {10001, 22, 32, 42, `before trigger fired`}, {10002, 23, 33, 43, `before trigger fired`}, {10003, 24, 34, 44, `before trigger fired`}, {10004, 25, 35, 45, `before trigger fired`}, {10005, 26, 36, 46, `before trigger fired`}, {6, ``, 45, 80, `before trigger fired`}, {7, ``, `x`, `\x`, `before trigger fired`}, {8, ``, `,`, `\,`, `before trigger fired`}, {3000, ``, `c`, ``, `before trigger fired`}, {4000, ``, `C`, ``, `before trigger fired`}, {4001, 1, `empty`, ``, `before trigger fired`}, {4002, 2, `null`, ``, `before trigger fired`}, {4003, 3, `Backslash`, `\`, `before trigger fired`}, {4004, 4, `BackslashX`, `\X`, `before trigger fired`}, {4005, 5, `N`, `N`, `before trigger fired`}, {4006, 6, `BackslashN`, `\N`, `before trigger fired`}, {4007, 7, `XX`, `XX`, `before trigger fired`}, {4008, 8, `Delimiter`, `:`, `before trigger fired`}, {50004, 25, 35, 45, `before trigger fired`}, {60004, 25, 35, 45, `before trigger fired`}, {60005, 26, 36, 46, `before trigger fired`}, {1, 1, `stuff`, `test_1`, `after trigger fired`}, {2, 2, `stuff`, `test_2`, `after trigger fired`}, {3, 3, `stuff`, `test_3`, `after trigger fired`}, {4, 4, `stuff`, `test_4`, `after trigger fired`}, {5, 5, `stuff`, `test_5`, `after trigger fired`}},
			},
			{
				Statement: `COPY x TO stdout;`,
			},
			{
				Statement: `9999	\N	\\N	NN	before trigger fired
10000	21	31	41	before trigger fired
10001	22	32	42	before trigger fired
10002	23	33	43	before trigger fired
10003	24	34	44	before trigger fired
10004	25	35	45	before trigger fired
10005	26	36	46	before trigger fired
6	\N	45	80	before trigger fired
7	\N	x	\\x	before trigger fired
8	\N	,	\\,	before trigger fired
3000	\N	c	\N	before trigger fired
4000	\N	C	\N	before trigger fired
4001	1	empty		before trigger fired
4002	2	null	\N	before trigger fired
4003	3	Backslash	\\	before trigger fired
4004	4	BackslashX	\\X	before trigger fired
4005	5	N	N	before trigger fired
4006	6	BackslashN	\\N	before trigger fired
4007	7	XX	XX	before trigger fired
4008	8	Delimiter	:	before trigger fired
50004	25	35	45	before trigger fired
60004	25	35	45	before trigger fired
60005	26	36	46	before trigger fired
1	1	stuff	test_1	after trigger fired
2	2	stuff	test_2	after trigger fired
3	3	stuff	test_3	after trigger fired
4	4	stuff	test_4	after trigger fired
5	5	stuff	test_5	after trigger fired
COPY x (c, e) TO stdout;`,
			},
			{
				Statement: `\\N	before trigger fired
31	before trigger fired
32	before trigger fired
33	before trigger fired
34	before trigger fired
35	before trigger fired
36	before trigger fired
45	before trigger fired
x	before trigger fired
,	before trigger fired
c	before trigger fired
C	before trigger fired
empty	before trigger fired
null	before trigger fired
Backslash	before trigger fired
BackslashX	before trigger fired
N	before trigger fired
BackslashN	before trigger fired
XX	before trigger fired
Delimiter	before trigger fired
35	before trigger fired
35	before trigger fired
36	before trigger fired
stuff	after trigger fired
stuff	after trigger fired
stuff	after trigger fired
stuff	after trigger fired
stuff	after trigger fired
COPY x (b, e) TO stdout WITH NULL 'I''m null';`,
			},
			{
				Statement: `I'm null	before trigger fired
21	before trigger fired
22	before trigger fired
23	before trigger fired
24	before trigger fired
25	before trigger fired
26	before trigger fired
I'm null	before trigger fired
I'm null	before trigger fired
I'm null	before trigger fired
I'm null	before trigger fired
I'm null	before trigger fired
1	before trigger fired
2	before trigger fired
3	before trigger fired
4	before trigger fired
5	before trigger fired
6	before trigger fired
7	before trigger fired
8	before trigger fired
25	before trigger fired
25	before trigger fired
26	before trigger fired
1	after trigger fired
2	after trigger fired
3	after trigger fired
4	after trigger fired
5	after trigger fired
CREATE TEMP TABLE y (
	col1 text,
	col2 text
);`,
			},
			{
				Statement: `INSERT INTO y VALUES ('Jackson, Sam', E'\\h');`,
			},
			{
				Statement: `INSERT INTO y VALUES ('It is "perfect".',E'\t');`,
			},
			{
				Statement: `INSERT INTO y VALUES ('', NULL);`,
			},
			{
				Statement: `COPY y TO stdout WITH CSV;`,
			},
			{
				Statement: `"Jackson, Sam",\h
"It is ""perfect"".",	
"",
COPY y TO stdout WITH CSV QUOTE '''' DELIMITER '|';`,
			},
			{
				Statement: `Jackson, Sam|\h
It is "perfect".|	
''|
COPY y TO stdout WITH CSV FORCE QUOTE col2 ESCAPE E'\\' ENCODING 'sql_ascii';`,
			},
			{
				Statement: `"Jackson, Sam","\\h"
"It is \"perfect\".","	"
"",
COPY y TO stdout WITH CSV FORCE QUOTE *;`,
			},
			{
				Statement: `"Jackson, Sam","\h"
"It is ""perfect"".","	"
"",
COPY y TO stdout (FORMAT CSV);`,
			},
			{
				Statement: `"Jackson, Sam",\h
"It is ""perfect"".",	
"",
COPY y TO stdout (FORMAT CSV, QUOTE '''', DELIMITER '|');`,
			},
			{
				Statement: `Jackson, Sam|\h
It is "perfect".|	
''|
COPY y TO stdout (FORMAT CSV, FORCE_QUOTE (col2), ESCAPE E'\\');`,
			},
			{
				Statement: `"Jackson, Sam","\\h"
"It is \"perfect\".","	"
"",
COPY y TO stdout (FORMAT CSV, FORCE_QUOTE *);`,
			},
			{
				Statement: `"Jackson, Sam","\h"
"It is ""perfect"".","	"
"",
\copy y TO stdout (FORMAT CSV)
"Jackson, Sam",\h
"It is ""perfect"".",	
"",
\copy y TO stdout (FORMAT CSV, QUOTE '''', DELIMITER '|')
Jackson, Sam|\h
It is "perfect".|	
''|
\copy y TO stdout (FORMAT CSV, FORCE_QUOTE (col2), ESCAPE E'\\')
"Jackson, Sam","\\h"
"It is \"perfect\".","	"
"",
\copy y TO stdout (FORMAT CSV, FORCE_QUOTE *)
"Jackson, Sam","\h"
"It is ""perfect"".","	"
"",
CREATE TEMP TABLE testnl (a int, b text, c int);`,
			},
			{
				Statement: `COPY testnl FROM stdin CSV;`,
			},
			{
				Statement: `CREATE TEMP TABLE testeoc (a text);`,
			},
			{
				Statement: `COPY testeoc FROM stdin CSV;`,
			},
			{
				Statement: `COPY testeoc TO stdout CSV;`,
			},
			{
				Statement: `a\.
\.b
c\.d
"\."
CREATE TEMP TABLE testnull(a int, b text);`,
			},
			{
				Statement: `INSERT INTO testnull VALUES (1, E'\\0'), (NULL, NULL);`,
			},
			{
				Statement: `COPY testnull TO stdout WITH NULL AS E'\\0';`,
			},
			{
				Statement: `1	\\0
\0	\0
COPY testnull FROM stdin WITH NULL AS E'\\0';`,
			},
			{
				Statement: `SELECT * FROM testnull;`,
				Results:   []sql.Row{{1, `\0`}, {``, ``}, {42, `\0`}, {``, ``}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE TABLE vistest (LIKE testeoc);`,
			},
			{
				Statement: `COPY vistest FROM stdin CSV;`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT * FROM vistest;`,
				Results:   []sql.Row{{`a0`}, {`b`}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `TRUNCATE vistest;`,
			},
			{
				Statement: `COPY vistest FROM stdin CSV;`,
			},
			{
				Statement: `SELECT * FROM vistest;`,
				Results:   []sql.Row{{`a1`}, {`b`}},
			},
			{
				Statement: `SAVEPOINT s1;`,
			},
			{
				Statement: `TRUNCATE vistest;`,
			},
			{
				Statement: `COPY vistest FROM stdin CSV;`,
			},
			{
				Statement: `SELECT * FROM vistest;`,
				Results:   []sql.Row{{`d1`}, {`e`}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT * FROM vistest;`,
				Results:   []sql.Row{{`d1`}, {`e`}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `TRUNCATE vistest;`,
			},
			{
				Statement: `COPY vistest FROM stdin CSV FREEZE;`,
			},
			{
				Statement: `SELECT * FROM vistest;`,
				Results:   []sql.Row{{`a2`}, {`b`}},
			},
			{
				Statement: `SAVEPOINT s1;`,
			},
			{
				Statement: `TRUNCATE vistest;`,
			},
			{
				Statement: `COPY vistest FROM stdin CSV FREEZE;`,
			},
			{
				Statement: `SELECT * FROM vistest;`,
				Results:   []sql.Row{{`d2`}, {`e`}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT * FROM vistest;`,
				Results:   []sql.Row{{`d2`}, {`e`}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `TRUNCATE vistest;`,
			},
			{
				Statement: `COPY vistest FROM stdin CSV FREEZE;`,
			},
			{
				Statement: `SELECT * FROM vistest;`,
				Results:   []sql.Row{{`x`}, {`y`}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `TRUNCATE vistest;`,
			},
			{
				Statement:   `COPY vistest FROM stdin CSV FREEZE;`,
				ErrorString: `cannot perform COPY FREEZE because the table was not created or truncated in the current subtransaction`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `TRUNCATE vistest;`,
			},
			{
				Statement: `SAVEPOINT s1;`,
			},
			{
				Statement:   `COPY vistest FROM stdin CSV FREEZE;`,
				ErrorString: `cannot perform COPY FREEZE because the table was not created or truncated in the current subtransaction`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO vistest VALUES ('z');`,
			},
			{
				Statement: `SAVEPOINT s1;`,
			},
			{
				Statement: `TRUNCATE vistest;`,
			},
			{
				Statement: `ROLLBACK TO SAVEPOINT s1;`,
			},
			{
				Statement:   `COPY vistest FROM stdin CSV FREEZE;`,
				ErrorString: `cannot perform COPY FREEZE because the table was not created or truncated in the current subtransaction`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `CREATE FUNCTION truncate_in_subxact() RETURNS VOID AS
$$
BEGIN
	TRUNCATE vistest;`,
			},
			{
				Statement: `EXCEPTION
  WHEN OTHERS THEN
	INSERT INTO vistest VALUES ('subxact failure');`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$ language plpgsql;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `INSERT INTO vistest VALUES ('z');`,
			},
			{
				Statement: `SELECT truncate_in_subxact();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `COPY vistest FROM stdin CSV FREEZE;`,
			},
			{
				Statement: `SELECT * FROM vistest;`,
				Results:   []sql.Row{{`d4`}, {`e`}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT * FROM vistest;`,
				Results:   []sql.Row{{`d4`}, {`e`}},
			},
			{
				Statement: `CREATE TEMP TABLE forcetest (
    a INT NOT NULL,
    b TEXT NOT NULL,
    c TEXT,
    d TEXT,
    e TEXT
);`,
			},
			{
				Statement: `\pset null NULL
BEGIN;`,
			},
			{
				Statement: `COPY forcetest (a, b, c) FROM STDIN WITH (FORMAT csv, FORCE_NOT_NULL(b), FORCE_NULL(c));`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT b, c FROM forcetest WHERE a = 1;`,
				Results:   []sql.Row{{``, `NULL`}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `COPY forcetest (a, b, c, d) FROM STDIN WITH (FORMAT csv, FORCE_NOT_NULL(c,d), FORCE_NULL(c,d));`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT c, d FROM forcetest WHERE a = 2;`,
				Results:   []sql.Row{{``, `NULL`}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `COPY forcetest (a, b, c) FROM STDIN WITH (FORMAT csv, FORCE_NULL(b), FORCE_NOT_NULL(c));`,
				ErrorString: `null value in column "b" of relation "forcetest" violates not-null constraint`,
			},
			{
				Statement: `CONTEXT:  COPY forcetest, line 1: "3,,"""
ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `COPY forcetest (d, e) FROM STDIN WITH (FORMAT csv, FORCE_NOT_NULL(b));`,
				ErrorString: `FORCE_NOT_NULL column "b" not referenced by COPY`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement:   `COPY forcetest (d, e) FROM STDIN WITH (FORMAT csv, FORCE_NULL(b));`,
				ErrorString: `FORCE_NULL column "b" not referenced by COPY`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `\pset null ''
create table check_con_tbl (f1 int);`,
			},
			{
				Statement: `create function check_con_function(check_con_tbl) returns bool as $$
begin
  raise notice 'input = %', row_to_json($1);`,
			},
			{
				Statement: `  return $1.f1 > 0;`,
			},
			{
				Statement: `end $$ language plpgsql immutable;`,
			},
			{
				Statement: `alter table check_con_tbl add check (check_con_function(check_con_tbl.*));`,
			},
			{
				Statement: `\d+ check_con_tbl
                               Table "public.check_con_tbl"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 f1     | integer |           |          |         | plain   |              | 
Check constraints:
    "check_con_tbl_check" CHECK (check_con_function(check_con_tbl.*))
copy check_con_tbl from stdin;`,
			},
			{
				Statement:   `copy check_con_tbl from stdin;`,
				ErrorString: `new row for relation "check_con_tbl" violates check constraint "check_con_tbl_check"`,
			},
			{
				Statement: `CONTEXT:  COPY check_con_tbl, line 1: "0"
select * from check_con_tbl;`,
				Results: []sql.Row{{1}, {``}},
			},
			{
				Statement: `CREATE ROLE regress_rls_copy_user;`,
			},
			{
				Statement: `CREATE ROLE regress_rls_copy_user_colperms;`,
			},
			{
				Statement: `CREATE TABLE rls_t1 (a int, b int, c int);`,
			},
			{
				Statement: `COPY rls_t1 (a, b, c) from stdin;`,
			},
			{
				Statement: `CREATE POLICY p1 ON rls_t1 FOR SELECT USING (a % 2 = 0);`,
			},
			{
				Statement: `ALTER TABLE rls_t1 ENABLE ROW LEVEL SECURITY;`,
			},
			{
				Statement: `ALTER TABLE rls_t1 FORCE ROW LEVEL SECURITY;`,
			},
			{
				Statement: `GRANT SELECT ON TABLE rls_t1 TO regress_rls_copy_user;`,
			},
			{
				Statement: `GRANT SELECT (a, b) ON TABLE rls_t1 TO regress_rls_copy_user_colperms;`,
			},
			{
				Statement: `COPY rls_t1 TO stdout;`,
			},
			{
				Statement: `1	4	1
2	3	2
3	2	3
4	1	4
COPY rls_t1 (a, b, c) TO stdout;`,
			},
			{
				Statement: `1	4	1
2	3	2
3	2	3
4	1	4
COPY rls_t1 (a) TO stdout;`,
			},
			{
				Statement: `1
2
3
4
COPY rls_t1 (a, b) TO stdout;`,
			},
			{
				Statement: `1	4
2	3
3	2
4	1
COPY rls_t1 (b, a) TO stdout;`,
			},
			{
				Statement: `4	1
3	2
2	3
1	4
SET SESSION AUTHORIZATION regress_rls_copy_user;`,
			},
			{
				Statement: `COPY rls_t1 TO stdout;`,
			},
			{
				Statement: `2	3	2
4	1	4
COPY rls_t1 (a, b, c) TO stdout;`,
			},
			{
				Statement: `2	3	2
4	1	4
COPY rls_t1 (a) TO stdout;`,
			},
			{
				Statement: `2
4
COPY rls_t1 (a, b) TO stdout;`,
			},
			{
				Statement: `2	3
4	1
COPY rls_t1 (b, a) TO stdout;`,
			},
			{
				Statement: `3	2
1	4
RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_rls_copy_user_colperms;`,
			},
			{
				Statement:   `COPY rls_t1 TO stdout;`,
				ErrorString: `permission denied for table rls_t1`,
			},
			{
				Statement:   `COPY rls_t1 (a, b, c) TO stdout;`,
				ErrorString: `permission denied for table rls_t1`,
			},
			{
				Statement:   `COPY rls_t1 (c) TO stdout;`,
				ErrorString: `permission denied for table rls_t1`,
			},
			{
				Statement: `COPY rls_t1 (a) TO stdout;`,
			},
			{
				Statement: `2
4
COPY rls_t1 (a, b) TO stdout;`,
			},
			{
				Statement: `2	3
4	1
RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `CREATE TABLE instead_of_insert_tbl(id serial, name text);`,
			},
			{
				Statement: `CREATE VIEW instead_of_insert_tbl_view AS SELECT ''::text AS str;`,
			},
			{
				Statement:   `COPY instead_of_insert_tbl_view FROM stdin; -- fail`,
				ErrorString: `cannot copy to view "instead_of_insert_tbl_view"`,
			},
			{
				Statement: `CREATE FUNCTION fun_instead_of_insert_tbl() RETURNS trigger AS $$
BEGIN
  INSERT INTO instead_of_insert_tbl (name) VALUES (NEW.str);`,
			},
			{
				Statement: `  RETURN NULL;`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `CREATE TRIGGER trig_instead_of_insert_tbl_view
  INSTEAD OF INSERT ON instead_of_insert_tbl_view
  FOR EACH ROW EXECUTE PROCEDURE fun_instead_of_insert_tbl();`,
			},
			{
				Statement: `COPY instead_of_insert_tbl_view FROM stdin;`,
			},
			{
				Statement: `SELECT * FROM instead_of_insert_tbl;`,
				Results:   []sql.Row{{1, `test1`}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `CREATE VIEW instead_of_insert_tbl_view_2 as select ''::text as str;`,
			},
			{
				Statement: `CREATE TRIGGER trig_instead_of_insert_tbl_view_2
  INSTEAD OF INSERT ON instead_of_insert_tbl_view_2
  FOR EACH ROW EXECUTE PROCEDURE fun_instead_of_insert_tbl();`,
			},
			{
				Statement: `COPY instead_of_insert_tbl_view_2 FROM stdin;`,
			},
			{
				Statement: `SELECT * FROM instead_of_insert_tbl;`,
				Results:   []sql.Row{{1, `test1`}, {2, `test1`}},
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `DROP TABLE forcetest;`,
			},
			{
				Statement: `DROP TABLE vistest;`,
			},
			{
				Statement: `DROP FUNCTION truncate_in_subxact();`,
			},
			{
				Statement: `DROP TABLE x, y;`,
			},
			{
				Statement: `DROP TABLE rls_t1 CASCADE;`,
			},
			{
				Statement: `DROP ROLE regress_rls_copy_user;`,
			},
			{
				Statement: `DROP ROLE regress_rls_copy_user_colperms;`,
			},
			{
				Statement: `DROP FUNCTION fn_x_before();`,
			},
			{
				Statement: `DROP FUNCTION fn_x_after();`,
			},
			{
				Statement: `DROP TABLE instead_of_insert_tbl;`,
			},
			{
				Statement: `DROP VIEW instead_of_insert_tbl_view;`,
			},
			{
				Statement: `DROP VIEW instead_of_insert_tbl_view_2;`,
			},
			{
				Statement: `DROP FUNCTION fun_instead_of_insert_tbl();`,
			},
		},
	})
}
