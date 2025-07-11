// Copyright 2023 Dolthub, Inc.
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

func TestSmokeTests(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Simple statements",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE TABLE test2 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test VALUES (1, 1), (2, 2);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test2 VALUES (3, 3), (4, 4);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 1},
						{2, 2},
					},
				},
				{
					Query: "SELECT * FROM test2;",
					Expected: []sql.Row{
						{3, 3},
						{4, 4},
					},
				},
				{
					Query: "SELECT test2.pk FROM test2;",
					Expected: []sql.Row{
						{3},
						{4},
					},
				},
				{
					Query: "SELECT * FROM test ORDER BY 1 LIMIT 1 OFFSET 1;",
					Expected: []sql.Row{
						{2, 2},
					},
				},
				{
					Query:    "SELECT NULL = NULL",
					Expected: []sql.Row{{nil}},
				},
				{
					Query:    ";",
					Expected: []sql.Row{},
				},
				{
					Query:    " ; ",
					Expected: []sql.Row{},
				},
				{
					Query:    "-- this is only a comment",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "Insert statements",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT8 PRIMARY KEY, v1 INT4, v2 INT2);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "INSERT INTO test VALUES (1, 2, 3);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test (v1, pk) VALUES (5, 4);",
					Expected: []sql.Row{},
				},
				{
					Query:    "INSERT INTO test (pk, v2) SELECT pk + 5, v2 + 10 FROM test WHERE v2 IS NOT NULL;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 2, 3},
						{4, 5, nil},
						{6, nil, 13},
					},
				},
			},
		},
		{
			Name: "Update statements",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT8 PRIMARY KEY, v1 INT4, v2 INT2);",
				"INSERT INTO test VALUES (1, 2, 3), (4, 5, 6), (7, 8, 9);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "UPDATE test SET v2 = 10;",
					Expected: []sql.Row{},
				},
				{
					Query:    "UPDATE test SET v1 = pk + v2;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 11, 10},
						{4, 14, 10},
						{7, 17, 10},
					},
				},
				{
					Query:    "UPDATE test SET pk = subquery.val FROM (SELECT 22 as val) AS subquery WHERE pk >= 7;",
					Skip:     true, // FROM not yet supported
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Skip:  true, // Above query doesn't run yet
					Expected: []sql.Row{
						{1, 11, 10},
						{4, 14, 10},
						{22, 17, 10},
					},
				},
			},
		},
		{
			Name: "Delete statements",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT8 PRIMARY KEY, v1 INT4, v2 INT2);",
				"INSERT INTO test VALUES (1, 1, 1), (2, 3, 4), (5, 7, 9);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "DELETE FROM test WHERE v2 = 9;",
					Expected: []sql.Row{},
				},
				{
					Query:    "DELETE FROM test WHERE v1 = pk;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{2, 3, 4},
					},
				},
			},
		},
		{
			Name: "USE statements",
			SetUpScript: []string{
				"CREATE DATABASE test",
				"USE test",
				"CREATE TABLE t1 (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				"INSERT INTO t1 VALUES (1, 1), (2, 2);",
				"call dolt_commit('-Am', 'initial commit');",
				"call dolt_branch('b1');",
				"call dolt_checkout('b1');",
				"INSERT INTO t1 VALUES (3, 3), (4, 4);",
				"call dolt_commit('-Am', 'commit b1');",
				"call dolt_tag('tag1')",
				"INSERT INTO t1 VALUES (5, 5), (6, 6);",
				"call dolt_checkout('main');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "select * from t1 order by 1;",
					Expected: []sql.Row{
						{1, 1},
						{2, 2},
					},
				},
				{
					Query:            "USE test/b1",
					SkipResultsCheck: true,
				},
				{
					Query: "select * from t1 order by 1;",
					Expected: []sql.Row{
						{1, 1},
						{2, 2},
						{3, 3},
						{4, 4},
						{5, 5},
						{6, 6},
					},
				},
				{
					Query:            "USE \"test/main\"",
					SkipResultsCheck: true,
				},
				{
					Query: "select * from t1 order by 1;",
					Expected: []sql.Row{
						{1, 1},
						{2, 2},
					},
				},
				{
					Query:            "USE 'test/tag1'",
					SkipResultsCheck: true,
				},
				{
					Query: "select * from t1 order by 1;",
					Expected: []sql.Row{
						{1, 1},
						{2, 2},
						{3, 3},
						{4, 4},
					},
				},
			},
		},
		{
			Name: "Boolean results",
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT 1 IN (2);",
					Expected: []sql.Row{
						{"f"},
					},
				},
				{
					Query: "SELECT 2 IN (2);",
					Expected: []sql.Row{
						{"t"},
					},
				},
			},
		},
		{
			Name: "Commit and diff across branches",
			SetUpScript: []string{
				"CREATE TABLE test (pk BIGINT PRIMARY KEY, v1 BIGINT);",
				"INSERT INTO test VALUES (1, 1), (2, 2);",
				"CALL DOLT_ADD('-A');",
				"CALL DOLT_COMMIT('-m', 'initial commit');",
				"CALL DOLT_BRANCH('other');",
				"UPDATE test SET v1 = 3;",
				"CALL DOLT_ADD('-A');",
				"CALL DOLT_COMMIT('-m', 'commit main');",
				"CALL DOLT_CHECKOUT('other');",
				"UPDATE test SET v1 = 4 WHERE pk = 2;",
				"CALL DOLT_ADD('-A');",
				"CALL DOLT_COMMIT('-m', 'commit other');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:            "CALL DOLT_CHECKOUT('main');",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 3},
						{2, 3},
					},
				},
				{
					Query:            "CALL DOLT_CHECKOUT('other');",
					SkipResultsCheck: true,
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1, 1},
						{2, 4},
					},
				},
				{
					Query: "SELECT from_pk, to_pk, from_v1, to_v1 FROM dolt_diff_test;",
					Expected: []sql.Row{
						{2, 2, 2, 4},
						{nil, 1, nil, 1},
						{nil, 2, nil, 2},
					},
				},
			},
		},
		{
			Name: "ARRAY expression",
			SetUpScript: []string{
				"CREATE TABLE test1 (id INTEGER primary key, v1 BOOLEAN);",
				"INSERT INTO test1 VALUES (1, 'true'), (2, 'false');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT ARRAY[v1]::boolean[] FROM test1 ORDER BY id;",
					Expected: []sql.Row{
						{"{t}"},
						{"{f}"},
					},
				},
				{
					Query: "SELECT ARRAY[v1] FROM test1 ORDER BY id;",
					Expected: []sql.Row{
						{"{t}"},
						{"{f}"},
					},
				},
				{
					Query: "SELECT ARRAY[v1, true, v1] FROM test1 ORDER BY id;",
					Expected: []sql.Row{
						{"{t,t,t}"},
						{"{f,t,f}"},
					},
				},
				{
					Query: "SELECT ARRAY[1::float8, 2::numeric];",
					Expected: []sql.Row{
						{"{1,2}"},
					},
				},
				{
					Query: "SELECT ARRAY[1::float8, NULL];",
					Expected: []sql.Row{
						{"{1,NULL}"},
					},
				},
				{
					Query: "SELECT ARRAY[1::int2, 2::int4, 3::int8]::varchar[];",
					Expected: []sql.Row{
						{"{1,2,3}"},
					},
				},
				{
					Query:       "SELECT ARRAY[1::int8]::int;",
					ExpectedErr: "cast from `bigint[]` to `integer` does not exist",
				},
				{
					Query:       "SELECT ARRAY[1::int8, 2::varchar];",
					ExpectedErr: "ARRAY types bigint and varchar cannot be matched",
				},
			},
		},
		{
			Name: "Array casting",
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT '{true,false,true}'::boolean[];`,
					Expected: []sql.Row{
						{`{t,f,t}`},
					},
				},
				{
					Skip:  true, // TODO: result differs from Postgres
					Query: `SELECT '{"\x68656c6c6f", "\x776f726c64", "\x6578616d706c65"}'::bytea[]::text[];`,
					Expected: []sql.Row{
						{`{"\\x7836383635366336633666","\\x7837373666373236633634","\\x783635373836313664373036633635"}`},
					},
				},
				{
					Skip:  true, // TODO: result differs from Postgres
					Query: `SELECT '{"\\x68656c6c6f", "\\x776f726c64", "\\x6578616d706c65"}'::bytea[]::text[];`,
					Expected: []sql.Row{
						{`{"\\x68656c6c6f", "\\x776f726c64", "\\x6578616d706c65"}`},
					},
				},
				{
					Query: `SELECT '{"abcd", "efgh", "ijkl"}'::char(3)[];`,
					Expected: []sql.Row{
						{`{abc,efg,ijk}`},
					},
				},
				{
					Query: `SELECT '{"2020-02-03", "2020-04-05", "2020-06-06"}'::date[];`,
					Expected: []sql.Row{
						{`{2020-02-03,2020-04-05,2020-06-06}`},
					},
				},
				{
					Query: `SELECT '{1.25,2.5,3.75}'::float4[];`,
					Expected: []sql.Row{
						{`{1.25,2.5,3.75}`},
					},
				},
				{
					Query: `SELECT '{4.25,5.5,6.75}'::float8[];`,
					Expected: []sql.Row{
						{`{4.25,5.5,6.75}`},
					},
				},
				{
					Query: `SELECT '{1,2,3}'::int2[];`,
					Expected: []sql.Row{
						{`{1,2,3}`},
					},
				},
				{
					Query: `SELECT '{4,5,6}'::int4[];`,
					Expected: []sql.Row{
						{`{4,5,6}`},
					},
				},
				{
					Query: `SELECT '{7,8,9}'::int8[];`,
					Expected: []sql.Row{
						{`{7,8,9}`},
					},
				},
				{
					Query: `SELECT '{"{\"a\":\"val1\"}", "{\"b\":\"value2\"}", "{\"c\": \"object_value3\"}"}'::json[];`,
					Expected: []sql.Row{
						{`{"{\"a\":\"val1\"}","{\"b\":\"value2\"}","{\"c\": \"object_value3\"}"}`},
					},
				},
				{
					Query: `SELECT '{"{\"d\":\"val1\"}", "{\"e\":\"value2\"}", "{\"f\": \"object_value3\"}"}'::jsonb[];`,
					Expected: []sql.Row{
						{`{"{\"d\": \"val1\"}","{\"e\": \"value2\"}","{\"f\": \"object_value3\"}"}`},
					},
				},
				{
					Query: `SELECT '{"the", "legendary", "formula"}'::name[];`,
					Expected: []sql.Row{
						{`{the,legendary,formula}`},
					},
				},
				{
					Query: `SELECT '{10.01,20.02,30.03}'::numeric[];`,
					Expected: []sql.Row{
						{`{10.01,20.02,30.03}`},
					},
				},
				{
					Query: `SELECT '{1,10,100}'::oid[];`,
					Expected: []sql.Row{
						{`{1,10,100}`},
					},
				},
				{
					Query: `SELECT '{"this", "is", "some", "text"}'::text[], '{text,without,quotes}'::text[], '{null,NULL,"NULL","quoted"}'::text[];`,
					Expected: []sql.Row{
						{`{this,is,some,text}`, `{text,without,quotes}`, `{NULL,NULL,"NULL",quoted}`},
					},
				},
				{
					Query: `SELECT '{"12:12:13", "14:14:15", "16:16:17"}'::time[];`,
					Expected: []sql.Row{
						{`{12:12:13,14:14:15,16:16:17}`},
					},
				},
				{
					Query: `SELECT '{"2020-02-03 12:13:14", "2020-04-05 15:16:17", "2020-06-06 18:19:20"}'::timestamp[];`,
					Expected: []sql.Row{
						{`{"2020-02-03 12:13:14","2020-04-05 15:16:17","2020-06-06 18:19:20"}`},
					},
				},
				{
					Query: `SELECT '{"3920fd79-7b53-437c-b647-d450b58b4532", "a594c217-4c63-4669-96ec-40eed180b7cf", "4367b70d-8d8b-4969-a1aa-bf59536455fb"}'::uuid[];`,
					Expected: []sql.Row{
						{`{3920fd79-7b53-437c-b647-d450b58b4532,a594c217-4c63-4669-96ec-40eed180b7cf,4367b70d-8d8b-4969-a1aa-bf59536455fb}`},
					},
				},
				{
					Query: `SELECT '{"somewhere", "over", "the", "rainbow"}'::varchar(5)[];`,
					Expected: []sql.Row{
						{`{somew,over,the,rainb}`},
					},
				},
				{
					Query: `SELECT '{1,2,3}'::xid[];`,
					Expected: []sql.Row{
						{`{1,2,3}`},
					},
				},
				{
					Query:       `SELECT '{"abc""","def"}'::text[];`,
					ExpectedErr: "malformed",
				},
				{
					Query:       `SELECT '{a,b,c'::text[];`,
					ExpectedErr: "malformed",
				},
				{
					Query:       `SELECT 'a,b,c}'::text[];`,
					ExpectedErr: "malformed",
				},
				{
					Query:       `SELECT '{"a,b,c}'::text[];`,
					ExpectedErr: "malformed",
				},
				{
					Query:       `SELECT '{a",b,c}'::text[];`,
					ExpectedErr: "malformed",
				},
				{
					Query:       `SELECT '{a,b,"c}'::text[];`,
					ExpectedErr: "malformed",
				},
				{
					Query:       `SELECT '{a,b,c"}'::text[];`,
					ExpectedErr: "malformed",
				},
			},
		},
		{
			Name: "BETWEEN",
			SetUpScript: []string{
				"CREATE TABLE test (v1 FLOAT8);",
				"INSERT INTO test VALUES (1), (3), (7);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN 1 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(1)},
						{float64(3)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN 2 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(3)},
					},
				},
				{
					Query:    "SELECT * FROM test WHERE v1 BETWEEN 4 AND 2 ORDER BY v1;",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN SYMMETRIC 1 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(1)},
						{float64(3)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN SYMMETRIC 2 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(3)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 BETWEEN SYMMETRIC 4 AND 2 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(3)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN 1 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(7)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN 2 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(1)},
						{float64(7)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN 4 AND 2 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(1)},
						{float64(3)},
						{float64(7)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN SYMMETRIC 1 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(7)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN SYMMETRIC 2 AND 4 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(1)},
						{float64(7)},
					},
				},
				{
					Query: "SELECT * FROM test WHERE v1 NOT BETWEEN SYMMETRIC 4 AND 2 ORDER BY v1;",
					Expected: []sql.Row{
						{float64(1)},
						{float64(7)},
					},
				},
			},
		},
		{
			Name: "IN",
			SetUpScript: []string{
				"CREATE TABLE test(v1 INT4, v2 INT4);",
				"INSERT INTO test VALUES (1, 1), (2, 2), (3, 3), (4, 4), (5, 5);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test WHERE v1 IN (2, '3', 4) ORDER BY v1;",
					Expected: []sql.Row{
						{2, 2},
						{3, 3},
						{4, 4},
					},
				},
				{
					Query:    "CREATE INDEX v2_idx ON test(v2);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test WHERE v2 IN (2, '3', 4) ORDER BY v1;",
					Expected: []sql.Row{
						{2, 2},
						{3, 3},
						{4, 4},
					},
				},
			},
		},
		{
			Name: "SUM",
			SetUpScript: []string{
				"CREATE TABLE test(pk SERIAL PRIMARY KEY, v1 INT4);",
				"INSERT INTO test (v1) VALUES (1), (2), (3), (4), (5);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT SUM(v1) FROM test WHERE v1 BETWEEN 3 AND 5;",
					Expected: []sql.Row{
						{12.0},
					},
				},
				{
					Query:    "CREATE INDEX v1_idx ON test(v1);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT SUM(v1) FROM test WHERE v1 BETWEEN 3 AND 5;",
					Expected: []sql.Row{
						{12.0},
					},
				},
			},
		},
		{
			Name: "Empty statement",
			Assertions: []ScriptTestAssertion{
				{
					Query:    ";",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "Unsupported MySQL statements",
			Assertions: []ScriptTestAssertion{
				{
					Query:       "SHOW CREATE TABLE;",
					ExpectedErr: "syntax error",
				},
			},
		},
		{
			Name: "querying tables with same name as pg_catalog tables",
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT attname FROM pg_catalog.pg_attribute ORDER BY attname LIMIT 3;",
					Expected: []sql.Row{
						{"ACTION_CONDITION"},
						{"ACTION_ORDER"},
						{"ACTION_ORIENTATION"},
					},
				},
				{
					Query: "SELECT attname FROM pg_attribute ORDER BY attname LIMIT 3;",
					Expected: []sql.Row{
						{"ACTION_CONDITION"},
						{"ACTION_ORDER"},
						{"ACTION_ORIENTATION"},
					},
				},
				{
					Query:    "CREATE TABLE pg_attribute (id INT);",
					Expected: []sql.Row{},
				},
				{
					Query:       "insert into pg_attribute values (1);",
					ExpectedErr: "Column count doesn't match value count at row 1",
				},
				{
					Query:    "insert into public.pg_attribute values (1);",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT attname FROM pg_attribute ORDER BY attname LIMIT 3;",
					Expected: []sql.Row{
						{"ACTION_CONDITION"},
						{"ACTION_ORDER"},
						{"ACTION_ORIENTATION"},
					},
				},
				{
					Query:    "SELECT * FROM public.pg_attribute;",
					Expected: []sql.Row{{1}},
				},
				{
					Query:       "drop table pg_attribute;",
					ExpectedErr: "tables cannot be dropped on database pg_catalog",
				},
				{
					Query:    "drop table public.pg_attribute;",
					Expected: []sql.Row{},
				},
				{
					Query:       "SELECT * FROM public.pg_attribute;",
					ExpectedErr: "table not found: pg_attribute",
				},
			},
		},
		{
			Name: "200 Row Test",
			SetUpScript: []string{
				"CREATE TABLE test (pk INT8 PRIMARY KEY);",
				"INSERT INTO test VALUES " +
					"(1),   (2),   (3),   (4),   (5),   (6),   (7),   (8),   (9),   (10)," +
					"(11),  (12),  (13),  (14),  (15),  (16),  (17),  (18),  (19),  (20)," +
					"(21),  (22),  (23),  (24),  (25),  (26),  (27),  (28),  (29),  (30)," +
					"(31),  (32),  (33),  (34),  (35),  (36),  (37),  (38),  (39),  (40)," +
					"(41),  (42),  (43),  (44),  (45),  (46),  (47),  (48),  (49),  (50)," +
					"(51),  (52),  (53),  (54),  (55),  (56),  (57),  (58),  (59),  (60)," +
					"(61),  (62),  (63),  (64),  (65),  (66),  (67),  (68),  (69),  (70)," +
					"(71),  (72),  (73),  (74),  (75),  (76),  (77),  (78),  (79),  (80)," +
					"(81),  (82),  (83),  (84),  (85),  (86),  (87),  (88),  (89),  (90)," +
					"(91),  (92),  (93),  (94),  (95),  (96),  (97),  (98),  (99),  (100)," +
					"(101), (102), (103), (104), (105), (106), (107), (108), (109), (110)," +
					"(111), (112), (113), (114), (115), (116), (117), (118), (119), (120)," +
					"(121), (122), (123), (124), (125), (126), (127), (128), (129), (130)," +
					"(131), (132), (133), (134), (135), (136), (137), (138), (139), (140)," +
					"(141), (142), (143), (144), (145), (146), (147), (148), (149), (150)," +
					"(151), (152), (153), (154), (155), (156), (157), (158), (159), (160)," +
					"(161), (162), (163), (164), (165), (166), (167), (168), (169), (170)," +
					"(171), (172), (173), (174), (175), (176), (177), (178), (179), (180)," +
					"(181), (182), (183), (184), (185), (186), (187), (188), (189), (190)," +
					"(191), (192), (193), (194), (195), (196), (197), (198), (199), (200);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test ORDER BY pk;",
					Expected: []sql.Row{
						{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10},
						{11}, {12}, {13}, {14}, {15}, {16}, {17}, {18}, {19}, {20},
						{21}, {22}, {23}, {24}, {25}, {26}, {27}, {28}, {29}, {30},
						{31}, {32}, {33}, {34}, {35}, {36}, {37}, {38}, {39}, {40},
						{41}, {42}, {43}, {44}, {45}, {46}, {47}, {48}, {49}, {50},
						{51}, {52}, {53}, {54}, {55}, {56}, {57}, {58}, {59}, {60},
						{61}, {62}, {63}, {64}, {65}, {66}, {67}, {68}, {69}, {70},
						{71}, {72}, {73}, {74}, {75}, {76}, {77}, {78}, {79}, {80},
						{81}, {82}, {83}, {84}, {85}, {86}, {87}, {88}, {89}, {90},
						{91}, {92}, {93}, {94}, {95}, {96}, {97}, {98}, {99}, {100},
						{101}, {102}, {103}, {104}, {105}, {106}, {107}, {108}, {109}, {110},
						{111}, {112}, {113}, {114}, {115}, {116}, {117}, {118}, {119}, {120},
						{121}, {122}, {123}, {124}, {125}, {126}, {127}, {128}, {129}, {130},
						{131}, {132}, {133}, {134}, {135}, {136}, {137}, {138}, {139}, {140},
						{141}, {142}, {143}, {144}, {145}, {146}, {147}, {148}, {149}, {150},
						{151}, {152}, {153}, {154}, {155}, {156}, {157}, {158}, {159}, {160},
						{161}, {162}, {163}, {164}, {165}, {166}, {167}, {168}, {169}, {170},
						{171}, {172}, {173}, {174}, {175}, {176}, {177}, {178}, {179}, {180},
						{181}, {182}, {183}, {184}, {185}, {186}, {187}, {188}, {189}, {190},
						{191}, {192}, {193}, {194}, {195}, {196}, {197}, {198}, {199}, {200},
					},
				},
			},
		},
	})
}
