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

func TestTypes(t *testing.T) {
	RunScripts(t, typesTests)
}

var typesTests = []ScriptTest{
	{
		Name: "Bigint type",
		SetUpScript: []string{
			"CREATE TABLE t_bigint (id INTEGER primary key, v1 BIGINT);",
			"INSERT INTO t_bigint VALUES (1, 123456789012345), (2, 987654321098765);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_bigint ORDER BY id;",
				Expected: []sql.Row{
					{1, 123456789012345},
					{2, 987654321098765},
				},
			},
			{
				Query:    `SELECT 1::pg_catalog.int8;`,
				Expected: []sql.Row{{1}},
			},
		},
	},
	{
		Name: "Bigint key",
		SetUpScript: []string{
			"CREATE TABLE t_bigint (id BIGINT primary key, v1 BIGINT);",
			"INSERT INTO t_bigint VALUES (1, 123456789012345), (2, 987654321098765);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_bigint WHERE id = 1 ORDER BY id;",
				Expected: []sql.Row{
					{1, 123456789012345},
				},
			},
		},
	},
	{
		Name: "Bigint array type",
		SetUpScript: []string{
			"CREATE TABLE t_bigint (id INTEGER primary key, v1 BIGINT[]);",
			"INSERT INTO t_bigint VALUES (1, ARRAY[123456789012345, NULL]), (2, ARRAY[987654321098765, 5]), (3, ARRAY[4, 5]);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_bigint ORDER BY id;",
				Expected: []sql.Row{
					{1, "{123456789012345,NULL}"},
					{2, "{987654321098765,5}"},
					{3, "{4,5}"},
				},
			},
		},
	},
	{
		Name: "Bit type",
		Skip: true, // no pgx support: unknown type with oid: 1560
		SetUpScript: []string{
			"CREATE TABLE t_bit (id INTEGER primary key, v1 BIT(8));",
			"INSERT INTO t_bit VALUES (1, B'11011010'), (2, B'00101011');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_bit ORDER BY id;",
				Expected: []sql.Row{
					{1, []byte{0xDA}},
					{2, []byte{0x2B}},
				},
			},
		},
	},
	{
		Name: "Bit key",
		Skip: true, // no pgx support: unknown type with oid: 1560
		SetUpScript: []string{
			"CREATE TABLE t_bit (id BIT(8) primary key, v1 BIT(8));",
			"INSERT INTO t_bit VALUES (B'11011010', B'11011010'), (B'00101011', B'00101011');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_bit WHERE id = B'11011010' ORDER BY id;",
				Expected: []sql.Row{
					{[]byte{0xDA}, []byte{0xDA}},
				},
			},
		},
	},
	{
		Name: "Boolean type",
		SetUpScript: []string{
			"CREATE TABLE t_boolean (id INTEGER primary key, v1 BOOLEAN);",
			"INSERT INTO t_boolean VALUES (1, true), (2, 'false'), (3, NULL);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_boolean ORDER BY id;",
				Skip:  true, // Proper NULL-ordering has not yet been implemented
				Expected: []sql.Row{
					{1, "t"},
					{2, "f"},
					{3, nil},
				},
			},
			{
				Query: "SELECT * FROM t_boolean ORDER BY v1;",
				Skip:  true, // Proper NULL-ordering has not yet been implemented
				Expected: []sql.Row{
					{2, "f"},
					{1, "t"},
					{3, nil},
				},
			},
			{
				Query: "SELECT * FROM t_boolean WHERE v1 IS NOT NULL ORDER BY id;",
				Expected: []sql.Row{
					{1, "t"},
					{2, "f"},
				},
			},
			{
				Query: "SELECT * FROM t_boolean WHERE v1 IS NOT NULL ORDER BY v1;",
				Expected: []sql.Row{
					{2, "f"},
					{1, "t"},
				},
			},
		},
	},
	{
		Name: "Boolean key",
		SetUpScript: []string{
			"CREATE TABLE t_boolean (id boolean primary key, v1 BOOLEAN);",
			"INSERT INTO t_boolean VALUES (true, true), (false, 'false')",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_boolean where id ORDER BY id;",
				Skip:  true, // Proper NULL-ordering has not yet been implemented
				Expected: []sql.Row{
					{"t", "t"},
				},
			},
		},
	},
	{
		Name: "boolean indexes",
		Skip: true, // panic
		SetUpScript: []string{
			"create table t (b bool);",
			"insert into t values (false);",
			"create table t_idx (b bool);",
			"create index idx on t_idx(b);",
			"insert into t_idx values (false);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "select * from t where (b in (false));",
				Expected: []sql.Row{
					{0},
				},
			},
			{
				Query: "select * from t_idx where (b in (false));",
				Expected: []sql.Row{
					{0},
				},
			},
		},
	},
	{
		Name: "Boolean array type",
		SetUpScript: []string{
			"CREATE TABLE t_boolean_array (id INTEGER primary key, v1 BOOLEAN[]);",
			"INSERT INTO t_boolean_array VALUES (1, ARRAY[true, false]), (2, ARRAY[false, true]), (3, ARRAY[true, true]), (4, ARRAY[false, false]), (5, ARRAY[true]), (6, ARRAY[false]), (7, NULL);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_boolean_array ORDER BY id;",
				Skip:  true, // Proper NULL-ordering has not yet been implemented
				Expected: []sql.Row{
					{1, "{t,f}"},
					{2, "{f,t}"},
					{3, "{t,t}"},
					{4, "{f,f}"},
					{5, "{t}"},
					{6, "{f}"},
					{7, nil},
				},
			},
			{
				Query: "SELECT * FROM t_boolean_array ORDER BY v1;",
				Skip:  true, // Proper NULL-ordering has not yet been implemented
				Expected: []sql.Row{
					{6, "{f}"},
					{4, "{f,f}"},
					{2, "{f,t}"},
					{5, "{t}"},
					{1, "{t,f}"},
					{3, "{t,t}"},
					{7, nil},
				},
			},
			{
				Query: "SELECT * FROM t_boolean_array WHERE v1 IS NOT NULL ORDER BY id;",
				Expected: []sql.Row{
					{1, "{t,f}"},
					{2, "{f,t}"},
					{3, "{t,t}"},
					{4, "{f,f}"},
					{5, "{t}"},
					{6, "{f}"},
				},
			},
			{
				Query: "SELECT * FROM t_boolean_array WHERE v1 IS NOT NULL ORDER BY v1;",
				Expected: []sql.Row{
					{6, "{f}"},
					{4, "{f,f}"},
					{2, "{f,t}"},
					{5, "{t}"},
					{1, "{t,f}"},
					{3, "{t,t}"},
				},
			},
		},
	},
	{
		Name: "Bigserial type",
		SetUpScript: []string{
			"CREATE TABLE t_bigserial (id INTEGER primary key, v1 BIGSERIAL);",
			"INSERT INTO t_bigserial VALUES (1, 123456789012345), (2, 987654321098765);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_bigserial ORDER BY id;",
				Expected: []sql.Row{
					{1, 123456789012345},
					{2, 987654321098765},
				},
			},
		},
	},
	{
		Name: "Bigserial key",
		SetUpScript: []string{
			"CREATE TABLE t_bigserial (id BIGSERIAL primary key, v1 BIGSERIAL);",
			"INSERT INTO t_bigserial VALUES (123456789012345, 123456789012345), (987654321098765, 987654321098765);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_bigserial where ID = 987654321098765 ORDER BY id;",
				Expected: []sql.Row{
					{987654321098765, 987654321098765},
				},
			},
		},
	},
	{
		Name: "Bit varying type",
		Skip: true,
		SetUpScript: []string{
			"CREATE TABLE t_bit_varying (id INTEGER primary key, v1 BIT VARYING(16));",
			"INSERT INTO t_bit_varying VALUES (1, B'1101101010101010'), (2, B'0010101101010101');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_bit_varying ORDER BY id;",
				Expected: []sql.Row{
					{1, []byte{0xDA, 0xAA}},
					{2, []byte{0x2B, 0xA5}},
				},
			},
		},
	},
	{
		Name: "Box type",
		Skip: true,
		SetUpScript: []string{
			"CREATE TABLE t_box (id INTEGER primary key, v1 BOX);",
			"INSERT INTO t_box VALUES (1, '(1,2),(3,4)'), (2, '(5,6),(7,8)');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_box ORDER BY id;",
				// TODO: the output and ordering of points here varies from postgres, probably need a GMS type, not a string
				Expected: []sql.Row{
					{1, "((1,2),(3,4))"},
					{2, "((5,6),(7,8))"},
				},
			},
		},
	},
	{
		Name: "Bytea type",
		SetUpScript: []string{
			"CREATE TABLE t_bytea (id INTEGER primary key, v1 BYTEA);",
			"INSERT INTO t_bytea VALUES (1, E'\\\\xDEADBEEF'), (2, '\\xC0FFEE'), (3, ''), (4, NULL);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_bytea ORDER BY id;",
				Expected: []sql.Row{
					{1, []byte{0xDE, 0xAD, 0xBE, 0xEF}},
					{2, []byte{0xC0, 0xFF, 0xEE}},
					{3, []byte{}},
					{4, nil},
				},
			},
		},
	},
	{
		Name: "Bytea key",
		Skip: true, // blob/text column 'id' used in key specification without a key length
		SetUpScript: []string{
			"CREATE TABLE t_bytea (id BYTEA primary key, v1 BYTEA);",
			"INSERT INTO t_bytea VALUES (E'\\\\xCAFEBABE', E'\\\\xDEADBEEF'), ('\\xBADD00D5', '\\xC0FFEE');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_bytea WHERE ID = E'\\\\xCAFEBABE' ORDER BY id;",
				Expected: []sql.Row{
					{[]byte{0xCA, 0xFE, 0xBA, 0xBE}, []byte{0xDE, 0xAD, 0xBE, 0xEF}},
				},
			},
		},
	},
	{
		Name: "Character type",
		SetUpScript: []string{
			"CREATE TABLE t_character (id INTEGER primary key, v1 CHARACTER(5));",
			"INSERT INTO t_character VALUES (1, 'abcde'), (2, 'vwxyz'), (3, 'ghi'), (4, ''), (5, NULL);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_character ORDER BY id;",
				Expected: []sql.Row{
					{1, "abcde"},
					{2, "vwxyz"},
					{3, "ghi  "},
					{4, "     "},
					{5, nil},
				},
			},
			{
				Query: "SELECT length(v1) FROM t_character ORDER BY id;",
				Expected: []sql.Row{
					{5},
					{5},
					{3},
					{0},
					{nil},
				},
			},
			{
				Query:    `SELECT char(20) 'characters' || ' and text' AS "Concat char to unknown type";`,
				Expected: []sql.Row{{"characters and text"}},
			},
			{
				Query: "SELECT true::char, false::char;",
				Expected: []sql.Row{
					{"t", "f"},
				},
			},
			{
				Query: "SELECT true::character(5), false::character(5);",
				Expected: []sql.Row{
					{"true ", "false"},
				},
			},
			{
				Query: "SELECT char 'c' = char 'c' AS true;",
				Expected: []sql.Row{
					{"t"},
				},
			},
		},
	},
	{
		Name: "Character key",
		SetUpScript: []string{
			"CREATE TABLE t_character (id CHAR(5) primary key, v1 CHARACTER(5));",
			"INSERT INTO t_character VALUES ('abcde', 'fghjk'), ('vwxyz', '12345'), ('vwxy', '1234')",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_character WHERE ID = 'vwxyz' ORDER BY id;",
				Expected: []sql.Row{
					{"vwxyz", "12345"},
				},
			},
			{
				Query: "SELECT length(id) FROM t_character;",
				Expected: []sql.Row{
					{5}, {5}, {4},
				},
			},
		},
	},
	{
		Name: "Internal char type",
		SetUpScript: []string{
			`CREATE TABLE t_char (id INTEGER primary key, v1 "char");`,
			`INSERT INTO t_char VALUES (1, 'abcde'), (2, 'vwxyz'), (3, '123'), (4, ''), (5, NULL), (100, 'こんにちは');`,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_char ORDER BY id;",
				Expected: []sql.Row{
					{1, "a"},
					{2, "v"},
					{3, "1"},
					{4, ""},
					{5, nil},
					{100, "\343"},
				},
			},
			{
				Query:       "INSERT INTO t_char VALUES (6, 7);",
				ExpectedErr: `target is of type "char" but expression is of type integer`,
			},
			{
				Query:       "INSERT INTO t_char VALUES (6, true);",
				ExpectedErr: `target is of type "char" but expression is of type boolean`,
			},
			{
				Query:       `SELECT true::"char";`,
				ExpectedErr: "cast from `boolean` to `\"char\"` does not exist",
			},
			{
				Query:       `SELECT 100000::bigint::"char";`,
				ExpectedErr: "cast from `bigint` to `\"char\"` does not exist",
			},
			{
				Query: `SELECT 'abc'::"char", '123'::varchar(3)::"char";`,
				Expected: []sql.Row{
					{"a", "1"},
				},
			},
			{
				Query: `SELECT 'def'::name::"char";`,
				Expected: []sql.Row{
					{"d"},
				},
			},
			{
				Query: `SELECT id, v1::int, v1::text FROM t_char WHERE id < 10;`,
				Expected: []sql.Row{
					{1, 97, "a"},
					{2, 118, "v"},
					{3, 1, "1"},
					{4, 0, ""},
					{5, nil, nil},
				},
			},
			{
				Skip:  true, // TODO: We currently return '227'
				Query: `SELECT v1::int FROM t_char WHERE id = 100;`,
				Expected: []sql.Row{
					{-29},
				},
			},
			{
				Query:    "INSERT INTO t_char VALUES (6, '0123456789012345678901234567890123456789012345678901234567890123456789');",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT * FROM t_char WHERE id=6;",
				Expected: []sql.Row{
					{6, "0"},
				},
			},
			{
				Query:       "INSERT INTO t_char VALUES (7, 'abc'::name);",
				ExpectedErr: "expression is of type",
			},
			{
				Query:    "INSERT INTO t_char VALUES (8, 'def'::text);",
				Expected: []sql.Row{},
			},
			{
				Query:    "INSERT INTO t_char VALUES (9, 'ghi'::varchar);",
				Expected: []sql.Row{},
			},
			{
				Query: `SELECT * FROM t_char WHERE id >= 7 AND id < 10 ORDER BY id;`,
				Expected: []sql.Row{
					{8, "d"},
					{9, "g"},
				},
			},
		},
	},
	{
		Name: "Character varying type",
		SetUpScript: []string{
			"CREATE TABLE t_varchar (id INTEGER primary key, v1 CHARACTER VARYING(10));",
			"INSERT INTO t_varchar VALUES (1, 'abcdefghij'), (2, 'klmnopqrst'), (3, ''), (4, NULL);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_varchar ORDER BY id;",
				Expected: []sql.Row{
					{1, "abcdefghij"},
					{2, "klmnopqrst"},
					{3, ""},
					{4, nil},
				},
			},
			{
				Query: "SELECT true::character varying(10), false::character varying(10);",
				Expected: []sql.Row{
					{"true", "false"},
				},
			},
		},
	},
	{
		Name: "Character varying type as primary key",
		SetUpScript: []string{
			"CREATE TABLE t_varchar (id INTEGER, v1 CHARACTER VARYING(10) primary key);",
			"INSERT INTO t_varchar VALUES (1, 'abcdefghij'), (2, 'klmnopqrst'), (3, '');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_varchar ORDER BY id;",
				Expected: []sql.Row{
					{1, "abcdefghij"},
					{2, "klmnopqrst"},
					{3, ""},
				},
			},
			{
				Query: "SELECT true::character varying(10), false::character varying(10);",
				Expected: []sql.Row{
					{"true", "false"},
				},
			},
		},
	},
	{
		Name: "Character varying array type, with length",
		SetUpScript: []string{
			"CREATE TABLE t_varchar1 (v1 CHARACTER VARYING[]);",
			"CREATE TABLE t_varchar2 (v1 CHARACTER VARYING(1)[]);",
			"INSERT INTO t_varchar1 VALUES (ARRAY['ab''cdef', 'what', 'is,hi', 'wh\"at']);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: `SELECT v1::varchar(1)[] FROM t_varchar1;`,
				Expected: []sql.Row{
					{"{a,w,i,w}"},
				},
			},
			{
				Query:       "INSERT INTO t_varchar2 VALUES (ARRAY['ab''cdef', 'what', 'is,hi', 'wh\"at']);",
				ExpectedErr: "too long",
			},
			{
				Query:    "INSERT INTO t_varchar2 VALUES (ARRAY['a', 'w', 'i', 'w']);",
				Expected: []sql.Row{},
			},
			{
				Query: `SELECT * FROM t_varchar2;`,
				Expected: []sql.Row{
					{"{a,w,i,w}"},
				},
			},
		},
	},
	{
		Name: "Character varying type, no length",
		SetUpScript: []string{
			"CREATE TABLE t_varchar (id INTEGER primary key, v1 CHARACTER VARYING);",
			"INSERT INTO t_varchar VALUES (1, 'abcdefghij'), (2, 'klmnopqrst');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_varchar ORDER BY id;",
				Expected: []sql.Row{
					{1, "abcdefghij"},
					{2, "klmnopqrst"},
				},
			},
		},
	},
	{
		Name: "Character varying type, no length, as primary key",
		Skip: true, // panic
		SetUpScript: []string{
			"CREATE TABLE t_varchar (id INTEGER, v1 CHARACTER VARYING primary key);",
			"INSERT INTO t_varchar VALUES (1, 'abcdefghij'), (2, 'klmnopqrst');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_varchar ORDER BY id;",
				Skip:  true, // missing the second row
				Expected: []sql.Row{
					{1, "abcdefghij"},
					{2, "klmnopqrst"},
				},
			},
		},
	},
	{
		Name: "Character varying array type, no length",
		SetUpScript: []string{
			"CREATE TABLE t_varchar (id INTEGER primary key, v1 CHARACTER VARYING[]);",
			"INSERT INTO t_varchar VALUES (1, '{abcdefghij, NULL}'), (2, ARRAY['ab''cdef', 'what', 'is,hi', 'wh\"at', '}', '{', '{}']);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_varchar ORDER BY id;",
				Expected: []sql.Row{
					{1, "{abcdefghij,NULL}"},
					{2, `{ab'cdef,what,"is,hi","wh\"at","}","{","{}"}`},
				},
			},
		},
	},
	{
		Name: "2D array",
		Skip: true, // multiple dimensions not supported yet
		SetUpScript: []string{
			"CREATE TABLE t_varchar (id INTEGER primary key, v1 CHARACTER VARYING[][]);",
			"INSERT INTO t_varchar VALUES (1, '{{abcdefghij, NULL}, {1234, abc}}'), (2, ARRAY['ab''cdef', 'what', 'is,hi', 'wh\"at', '}', '{', '{}']);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_varchar ORDER BY id;",
				Expected: []sql.Row{
					{1, "{abcdefghij,NULL}"},
					{2, `{ab'cdef,what,"is,hi","wh\"at","}","{","{}"}`},
				},
			},
		},
	}, {
		Name: "Cidr type",
		Skip: true,
		SetUpScript: []string{
			"CREATE TABLE t_cidr (id INTEGER primary key, v1 CIDR);",
			"INSERT INTO t_cidr VALUES (1, '192.168.1.0/24'), (2, '10.0.0.0/8');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_cidr ORDER BY id;",
				Expected: []sql.Row{
					{1, "192.168.1.0/24"},
					{2, "10.0.0.0/8"},
				},
			},
		},
	},
	{
		Name: "Circle type",
		Skip: true,
		SetUpScript: []string{
			"CREATE TABLE t_circle (id INTEGER primary key, v1 CIRCLE);",
			"INSERT INTO t_circle VALUES (1, '<(1,2),3>'), (2, '<(4,5),6>');",
		},
		Assertions: []ScriptTestAssertion{
			{
				// TODO: might need a GMS type here, not a string
				Query: "SELECT * FROM t_circle ORDER BY id;",
				Expected: []sql.Row{
					{1, "<(1,2),3>"},
					{2, "<(4,5),6>"},
				},
			},
		},
	},
	{
		Name: "Date type",
		SetUpScript: []string{
			"CREATE TABLE t_date (id INTEGER primary key, v1 DATE);",
			"INSERT INTO t_date VALUES (1, '2023-01-01'), (2, '2023-02-02');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_date ORDER BY id;",
				Expected: []sql.Row{
					{1, "2023-01-01"},
					{2, "2023-02-02"},
				},
			},
			{
				Query: "SELECT date '2022-2-2'",
				Expected: []sql.Row{
					{"2022-02-02"},
				},
			},
			{
				Query: "SELECT date '2022-02-02'",
				Expected: []sql.Row{
					{"2022-02-02"},
				},
			},
			{
				Query: "select '2024-10-31'::date;",
				Expected: []sql.Row{
					{"2024-10-31"},
				},
			},
			{
				Query: "select '2024-OCT-31'::date;",
				Expected: []sql.Row{
					{"2024-10-31"},
				},
			},
			{
				Query: "select '20241031'::date;",
				Expected: []sql.Row{
					{"2024-10-31"},
				},
			},
			{
				Query: "select '2024Oct31'::date;",
				Expected: []sql.Row{
					{"2024-10-31"},
				},
			},
			{
				Query: "select '10 31 2024'::date;",
				Expected: []sql.Row{
					{"2024-10-31"},
				},
			},
			{
				Query: "select 'Oct 31 2024'::date;",
				Expected: []sql.Row{
					{"2024-10-31"},
				},
			},
			{
				Query: "SELECT date 'J2451187';",
				Expected: []sql.Row{
					{"1999-01-08"},
				},
			},
			{
				Query:    `SELECT date '08-Jan-99';`,
				Expected: []sql.Row{{"1999-01-08"}},
			},
			{
				Query:    `SELECT date '2025-07-21' - 1;`,
				Expected: []sql.Row{{"2025-07-20"}},
			},
			{
				Query:    `SELECT date '2025-07-21' - date '2025-07-18';`,
				Expected: []sql.Row{{3}},
			},
			{
				Query:    `SELECT date '2025-07-21' - interval '2 days';`,
				Expected: []sql.Row{{"2025-07-19 00:00:00"}},
			},
			{
				Query:    `SELECT date '1991-02-03' - time '04:05:06';`,
				Expected: []sql.Row{{"1991-02-02 19:54:54"}},
			},
			{
				Query:    `SELECT date '2025-07-21' - 1;`,
				Expected: []sql.Row{{"2025-07-20"}},
			},
			{
				Query:    `SELECT date '1991-02-03' - time '04:05:06';`,
				Expected: []sql.Row{{"1991-02-02 19:54:54"}},
			},
			{
				Query:    `SELECT date '2025-07-21' + 1;`,
				Expected: []sql.Row{{"2025-07-22"}},
			},
			{
				Query:    `SELECT date '2025-07-21' + interval '2 days';`,
				Expected: []sql.Row{{"2025-07-23 00:00:00"}},
			},
			{
				Query:    `SELECT date '2025-07-21' + time '04:05:06';`,
				Expected: []sql.Row{{"2025-07-21 04:05:06"}},
			},
			{
				Query:    `SELECT date '2025-07-21' + time '04:05:06 UTC';`,
				Expected: []sql.Row{{"2025-07-21 04:05:06"}},
			},
		},
	},
	{
		Name: "Date key",
		SetUpScript: []string{
			"CREATE TABLE t_date (id DATE primary key, v1 DATE);",
			"INSERT INTO t_date VALUES ('2025-01-01', '2023-01-01'), ('2026-01-01', '2023-02-02');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_date where Id = '2025-01-01' ORDER BY id;",
				Expected: []sql.Row{
					{"2025-01-01", "2023-01-01"},
				},
			},
		},
	},
	{
		Name: "Double precision type",
		SetUpScript: []string{
			"CREATE TABLE t_double_precision (id INTEGER primary key, v1 DOUBLE PRECISION);",
			"INSERT INTO t_double_precision VALUES (1, 123.456), (2, 789.012);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_double_precision ORDER BY id;",
				Expected: []sql.Row{
					{1, 123.456},
					{2, 789.012},
				},
			},
		},
	},
	{
		Name: "Double precision key",
		SetUpScript: []string{
			"CREATE TABLE t_double_precision (id DOUBLE PRECISION primary key, v1 DOUBLE PRECISION);",
			"INSERT INTO t_double_precision VALUES (456.789, 123.456), (123.456, 789.012);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_double_precision WHERE id = 456.789 ORDER BY id;",
				Expected: []sql.Row{
					{456.789, 123.456},
				},
			},
		},
	},
	{
		Name: "Double precision array type",
		SetUpScript: []string{
			"CREATE TABLE t_double_precision (id INTEGER primary key, v1 DOUBLE PRECISION[]);",
			"INSERT INTO t_double_precision VALUES (1, ARRAY[123.456, NULL]), (2, ARRAY[789.012, 125.125]);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_double_precision ORDER BY id;",
				Expected: []sql.Row{
					{1, "{123.456,NULL}"},
					{2, "{789.012,125.125}"},
				},
			},
		},
	},
	{
		Name: "Inet type",
		Skip: true,
		SetUpScript: []string{
			"CREATE TABLE t_inet (id INTEGER primary key, v1 INET);",
			"INSERT INTO t_inet VALUES (1, '192.168.1.1'), (2, '10.0.0.1');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_inet ORDER BY id;",
				Expected: []sql.Row{
					{1, "192.168.1.1"},
					{2, "10.0.0.1"},
				},
			},
		},
	},
	{
		Name: "Integer type",
		SetUpScript: []string{
			"CREATE TABLE t_integer (id INTEGER primary key, v1 INTEGER);",
			"INSERT INTO t_integer VALUES (1, 123), (2, 456);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_integer ORDER BY id;",
				Expected: []sql.Row{
					{1, 123},
					{2, 456},
				},
			},
		},
	},
	{
		Name: "Integer array type",
		SetUpScript: []string{
			"CREATE TABLE t_integer (id INTEGER primary key, v1 INTEGER[]);",
			"INSERT INTO t_integer VALUES (1, ARRAY[123,NULL]), (2, ARRAY[456,823753913]);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_integer ORDER BY id;",
				Expected: []sql.Row{
					{1, "{123,NULL}"},
					{2, "{456,823753913}"},
				},
			},
		},
	},
	{
		Name: "Interval type",
		SetUpScript: []string{
			"CREATE TABLE t_interval (id INTEGER primary key, v1 INTERVAL);",
			"INSERT INTO t_interval VALUES (1, '1 day 3 hours'), (2, '23 hours 30 minutes'), (3, '@ 1 minute');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_interval ORDER BY id;",
				Expected: []sql.Row{
					{1, "1 day 03:00:00"},
					{2, "23:30:00"},
					{3, "00:01:00"},
				},
			},
			{
				Query: "SELECT * FROM t_interval ORDER BY v1;",
				Expected: []sql.Row{
					{3, "00:01:00"},
					{2, "23:30:00"},
					{1, "1 day 03:00:00"},
				},
			},
			{
				Query: `SELECT id, v1::char, v1::name FROM t_interval;`,
				Expected: []sql.Row{
					{1, "1", "1 day 03:00:00"},
					{2, "2", "23:30:00"},
					{3, "0", "00:01:00"},
				},
			},
			{
				Query:    `SELECT '2 years 15 months 100 weeks 99 hours 123456789 milliseconds'::interval;`,
				Expected: []sql.Row{{"3 years 3 mons 700 days 133:17:36.789"}},
			},
			{
				Query:    `SELECT '2 years 15 months 100 weeks 99 hours 123456789 milliseconds'::interval::char;`,
				Expected: []sql.Row{{"3"}},
			},
			{
				Query:    `SELECT '2 years 15 months 100 weeks 99 hours 123456789 milliseconds'::interval::text;`,
				Expected: []sql.Row{{"3 years 3 mons 700 days 133:17:36.789"}},
			},
			{
				Query:    `SELECT '2 years 15 months 100 weeks 99 hours 123456789 milliseconds'::char::interval;`,
				Expected: []sql.Row{{"00:00:02"}},
			},
			{
				Query:    `SELECT '13 months'::name::interval;`,
				Expected: []sql.Row{{"1 year 1 mon"}},
			},
			{
				Query:    `SELECT '13 months'::bpchar::interval;`,
				Expected: []sql.Row{{"1 year 1 mon"}},
			},
			{
				Query:    `SELECT '13 months'::varchar::interval;`,
				Expected: []sql.Row{{"1 year 1 mon"}},
			},
			{
				Query:    `SELECT '13 months'::text::interval;`,
				Expected: []sql.Row{{"1 year 1 mon"}},
			},
			{
				Query:    `SELECT '13 months'::char::interval;`,
				Expected: []sql.Row{{"00:00:01"}},
			},
			{
				Query:       "INSERT INTO t_interval VALUES (3, 7);",
				ExpectedErr: `ASSIGNMENT_CAST: target is of type interval but expression is of type integer: 7`,
			},
			{
				Query:       "INSERT INTO t_interval VALUES (3, true);",
				ExpectedErr: `ASSIGNMENT_CAST: target is of type interval but expression is of type boolean: true`,
			},
			{
				Query:    `SELECT CAST(interval '02:03' AS time) AS "02:03:00";`,
				Expected: []sql.Row{{"02:03:00"}},
			},
		},
	},
	{
		Name: "Interval key",
		SetUpScript: []string{
			"CREATE TABLE t_interval (id interval primary key, v1 INTERVAL);",
			"INSERT INTO t_interval VALUES ('1 hour', '1 day 3 hours'), ('2 days', '23 hours 30 minutes');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_interval WHERE id = '1 hour' ORDER BY id;",
				Expected: []sql.Row{
					{"01:00:00", "1 day 03:00:00"},
				},
			},
		},
	},
	{
		Name: "Interval array type",
		SetUpScript: []string{
			"CREATE TABLE t_interval_array (id INTEGER primary key, v1 INTERVAL[]);",
			"INSERT INTO t_interval_array VALUES (1, ARRAY['1 day 3 hours'::interval,'5 days 2 hours'::interval]), (2, ARRAY['3 years 3 mons 700 days 133:17:36.789'::interval,'200 hours'::interval]);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_interval_array ORDER BY id;",
				Expected: []sql.Row{
					{1, `{"1 day 03:00:00","5 days 02:00:00"}`},
					{2, `{"3 years 3 mons 700 days 133:17:36.789",200:00:00}`},
				},
			},
		},
	},
	{
		Name:        "JSON key",
		SetUpScript: []string{},
		Assertions: []ScriptTestAssertion{
			{
				Query:       "CREATE TABLE t_json (id JSON primary key, v1 JSON);",
				ExpectedErr: "data type json has no default operator class for access method \"btree\"",
				Skip:        true, // current error message is blob/text column 'id' used in key specification without a key length
			},
		},
	},
	{
		Name: "JSON type",
		SetUpScript: []string{
			"CREATE TABLE t_json (id INTEGER primary key, v1 JSON);",
			`INSERT INTO t_json VALUES (1, '{"key1": {"key": "value"}}'), (2, '{"num":42}'), (3, '{"key1": "value1", "key2": "value2"}'), (4, '{"key1": {"key": [2,3]}}');`,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_json ORDER BY 1;",
				Expected: []sql.Row{
					{1, `{"key1": {"key": "value"}}`},
					{2, `{"num":42}`},
					{3, `{"key1": "value1", "key2": "value2"}`},
					{4, `{"key1": {"key": [2,3]}}`},
				},
			},
			{
				Query: "SELECT * FROM t_json ORDER BY id;",
				Expected: []sql.Row{
					{1, `{"key1": {"key": "value"}}`},
					{2, `{"num":42}`},
					{3, `{"key1": "value1", "key2": "value2"}`},
					{4, `{"key1": {"key": [2,3]}}`},
				},
			},
			{
				Query: "SELECT '5'::json;",
				Expected: []sql.Row{
					{`5`},
				},
			},
			{
				Query: "SELECT 'false'::json;",
				Expected: []sql.Row{
					{`false`},
				},
			},
			{
				Query: `SELECT '"hi"'::json;`,
				Expected: []sql.Row{
					{`"hi"`},
				},
			},
			{
				Query: `SELECT 'null'::json;`,
				Expected: []sql.Row{
					{`null`},
				},
			},
			{
				Query: `SELECT '{"reading": 1.230e-5}'::json;`,
				Expected: []sql.Row{
					{`{"reading": 1.230e-5}`},
				},
			},
			{
				Query: `select json '{ "a":  "\ud83d\ude04\ud83d\udc36" }' -> 'a'`,
				Expected: []sql.Row{
					{`"\ud83d\ude04\ud83d\udc36"`},
				},
			},
		},
	},
	{
		Name: "JSON column default",
		SetUpScript: []string{
			`CREATE TABLE t_json (id INTEGER primary key, v1 JSON DEFAULT '{"num": 42}'::JSON);`,
			`INSERT INTO t_json VALUES (1, '{"key1": {"key": "value"}}');`,
			`INSERT INTO t_json (id) VALUES (2);`,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_json ORDER BY 1;",
				Expected: []sql.Row{
					{1, `{"key1": {"key": "value"}}`},
					{2, `{"num":42}`},
				},
			},
		},
	},
	{
		Name: "JSONB type",
		SetUpScript: []string{
			"CREATE TABLE t_jsonb (id INTEGER primary key, v1 JSONB);",
			"INSERT INTO t_jsonb VALUES (1, '{\"key\": \"value\"}'), (2, '{\"num\": 42}');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_jsonb ORDER BY id;",
				Expected: []sql.Row{
					{1, `{"key": "value"}`},
					{2, `{"num": 42}`},
				},
			},
			{
				Query: `SELECT '{"bar": "baz", "balance": 7.77, "active":false}'::jsonb;`,
				Expected: []sql.Row{
					{`{"bar": "baz", "active": false, "balance": 7.77}`},
				},
			},
			{
				Query: `SELECT '{"active": "baz", "active":false, "balance": 7.77}'::jsonb;`,
				Expected: []sql.Row{
					{`{"active": false, "balance": 7.77}`},
				},
			},
			{
				Query: `SELECT '{"active":false, "balance": 7.77, "bar": "baz"}'::jsonb;`,
				Expected: []sql.Row{
					{`{"bar": "baz", "active": false, "balance": 7.77}`},
				},
			},
			{
				Query: `SELECT jsonb '{"a":null, "b":"qq"}' ? 'a';`,
				Expected: []sql.Row{
					{"t"},
				},
			},
		},
	},
	{
		Name: "JSONB column default",
		SetUpScript: []string{
			`CREATE TABLE t_json (id INTEGER primary key, v1 JSONB DEFAULT '{"num": 42}'::JSONB);`,
			`INSERT INTO t_json VALUES (1, '{"key1": {"key": "value"}}');`,
			`INSERT INTO t_json (id) VALUES (2);`,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_json ORDER BY 1;",
				Expected: []sql.Row{
					{1, `{"key1": {"key": "value"}}`},
					{2, `{"num": 42}`},
				},
			},
		},
	},
	{
		Name: "JSONB ORDER BY",
		SetUpScript: []string{
			`CREATE TABLE t_jsonb (v1 JSONB);`,
			`INSERT INTO t_jsonb VALUES
				('["string_with_emoji_😊"]'),
				('[null, "null_as_string", false, 0]'),
				('{"key1": "value1", "key2": "value2", "key3": "value3"}'),
				('{"simple": "object"}'),
				('["special_chars_!@#$%^&*()_+", {"more": "!@#$"}]'),
				('[null, 1, "two", true, {"five": 5}]'),
				('[true, false, true]'),
				('{"key1": 123, "key2": "duplicate_key", "common_key": "same_value"}'),
				('["emoji_😀", "nested_😂", {"key": "value"}]'),
				('{"common_key": 456}'),
				('{"common_key": 123}'),
				('{"mixed_data": {"number": 100, "string": "text", "bool": false, "null": null}}'),
				('{"nested": {"level1": {"level2": {"key": "deep_value"}}}}'),
				('[1.1, 2.2, 3.3, 4.4, 5.5]'),
				('[{"nested_array": [1, 2, {"deep": {"inner": "value"}}]}, "text"]'),
				('{"common_key": "same_value"}'),
				('["end", "of", "array", 123, true]'),
				('"random string"'),
				('{"unicode": "こんにちは", "emoji": "😊"}'),
				('{"keyX": "string_value", "keyY": 123.456, "keyZ": null}'),
				('[{"key1": "value1"}, {"key2": "value2"}]'),
				('{"array_of_arrays": {"array1": [1, 2, 3], "array2": [4, 5, 6], "array3": [7, 8, 9]}}'),
				('{"key1": 123, "key2": "value", "key3": true}'),
				('{"key1": 1, "key2": 2, "key3": 3, "key4": 4, "key5": 5}'),
				('{"numbers": [1, 2, 3], "strings": ["a", "b", "c"], "booleans": [true, false]}'),
				('{"unicode_chars": {"char1": "あ", "char2": "い", "char3": "う"}}'),
				('[true, null, "string", 3.14]'),
				('{"array_of_bools": [true, false, true]}'),
				('[-1, -2, -3, -4]'),
				('[{"nested_array": [1, 2, 3]}, {"nested_object": {"inner_key": "inner_value"}}]'),
				('{"single": 1, "double": 2, "triple": 3, "quadruple": 4}'),
				('true'),
				('{"complex_array": {"array1": [1, 2, 3], "array2": ["a", "b", "c"]}}'),
				('["mixed", 123, false, null, {"complex": {"key": "value"}}]'),
				('{"array_of_strings": ["one", "two", "three"]}'),
				('["simple_text"]'),
				('{"mixed": {"number": 100, "string": "text", "bool": false, "null": null}}'),
				('{"boolean_true": true, "boolean_false": false, "null_value": null}'),
				('[{"deep": {"structure": {"key": "value"}}}, 123, false]'),
				('{"nested_numbers": {"one": 1, "two": 2, "three": 3}}'),
				('[{"emoji": "😊"}, {"another_emoji": "😢"}]'),
				('["just_text"]'),
				('{"common_key": "different_value"}'),
				('[[], [], []]'),
				('{"array_of_objects": [{"key1": "value1"}, {"key2": "value2"}, {"key3": "value3"}]}'),
				('{"combos": [{"number": 1}, {"string": "two"}, {"boolean": true}]}'),
				('{"keyA": 456, "keyB": "another_value", "keyC": false, "keyD": [1, 2, 3]}'),
				('[true, false, true, false, null]'),
				('[{"deep_nested": {"level1": {"level2": {"level3": "value"}}}}, 42, "text"]'),
				('{"empty": {}}'),
				('{"common_key": {"nested_key": "different_value"}}'),
				('["a", "b", "c", {"nested": {"key": "value"}}]'),
				('{"deep_nesting": {"level1": {"level2": {"level3": {"key": "value"}}}}}'),
				('{"random_text": "Lorem ipsum dolor sit amet"}'),
				('{"nested_string": {"outer": {"inner": "text"}}}'),
				('[1, 2, 3, 4, 5]'),
				('{"single_bool": true}'),
				('[1234567890, "large_number", false]'),
				('{"array_of_numbers": [1, 2, 3]}'),
				('[3.14159, 2.71828, 1.61803]'),
				('{"common_key": {"nested_key": "value"}}'),
				('["string1", "string2", "string3"]'),
				('{"single_string": "hello"}'),
				('{"nested_mixed": {"key1": 1, "key2": [true, false], "key3": {"inner_key": "inner_value"}}}'),
				('[0.1, 0.2, 0.3, 0.4]'),
				('[{"unicode": "こんにちは"}, {"another": "你好"}]'),
				('[1, "two", true, null, [1, 2, 3]]'),
				('["flat", "array", "of", "strings"]'),
				('123456'),
				('{"nested_object": {"subkey1": 789, "subkey2": [true, false], "subkey3": {"deep": "value"}}}'),
				('[{"key": {"subkey": [1, 2, 3]}}, 42, "text", false]'),
				('{"string_with_numbers": {"key": "123abc", "another_key": "456def"}}'),
				('{"unicode_string": {"greeting": "你好"}}'),
				('[{"key": "value"}, {"array": [1, 2, 3]}, {"nested": {"inner": "deep"}}]'),
				('["simple", "array", "of", "strings"]'),
				('{"text": "simple_string", "integer": 123, "float": 3.14}'),
				('[[], ["nested", "array"], 123]'),
				('{"object_in_array": [{"key": "value"}, {"another": "one"}]}'),
				('{"single_number": 42}'),
				('[null, null, null]'),
				('{"random_mixed": {"number": 1, "string": "two", "boolean": true, "null": null}}'),
				('null'),
				('["varied", "types", true, 123, {"key": "value"}]'),
				('[true, false, null, "end"]'),
				('789.123'),
				('["unicode_안녕하세요", "string"]'),
				('{"empty_object": {}, "empty_array": [], "boolean": true}'),
				('["text", 123, false, {"key": "value"}, [1, 2, 3]]'),
				('["multiple", "types", 123, true, {"key": "value"}]'),
				('{"boolean_mixed": {"true": true, "false": false, "null": null}}'),
				('{"object_in_array": {"array": [1, 2, 3], "nested": {"key": "value"}}}'),
				('[123, 456, 789]'),
				('[{"obj_in_array": {"key": "value"}}, [1, 2, 3], false]'),
				('false'),
				('[{"complex": {"nested": {"structure": "value"}}}, [1, 2, 3], false]'),
				('{"simple_object": {"key": "value"}}'),
				('{"number_key": {"integer": 1, "float": 2.3, "negative": -1}}'),
				('{"complex_object": {"key1": {"subkey": "value1"}, "key2": {"subkey": "value2"}}}'),
				('[1, "two", true, null, {"key": "value"}]');`,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_jsonb ORDER BY v1;",
				Expected: []sql.Row{
					{`null`},
					{`"random string"`},
					{`789.123`},
					{`123456`},
					{`false`},
					{`true`},
					{`["just_text"]`},
					{`["simple_text"]`},
					{`["string_with_emoji_😊"]`},
					{`["special_chars_!@#$%^&*()_+", {"more": "!@#$"}]`},
					{`["unicode_안녕하세요", "string"]`},
					{`[{"emoji": "😊"}, {"another_emoji": "😢"}]`},
					{`[{"key1": "value1"}, {"key2": "value2"}]`},
					{`[{"nested_array": [1, 2, 3]}, {"nested_object": {"inner_key": "inner_value"}}]`},
					{`[{"nested_array": [1, 2, {"deep": {"inner": "value"}}]}, "text"]`},
					{`[{"unicode": "こんにちは"}, {"another": "你好"}]`},
					{`[null, null, null]`},
					{`["emoji_😀", "nested_😂", {"key": "value"}]`},
					{`["string1", "string2", "string3"]`},
					{`[3.14159, 2.71828, 1.61803]`},
					{`[123, 456, 789]`},
					{`[1234567890, "large_number", false]`},
					{`[true, false, true]`},
					{`[[], [], []]`},
					{`[[], ["nested", "array"], 123]`},
					{`[{"complex": {"nested": {"structure": "value"}}}, [1, 2, 3], false]`},
					{`[{"deep": {"structure": {"key": "value"}}}, 123, false]`},
					{`[{"deep_nested": {"level1": {"level2": {"level3": "value"}}}}, 42, "text"]`},
					{`[{"key": "value"}, {"array": [1, 2, 3]}, {"nested": {"inner": "deep"}}]`},
					{`[{"obj_in_array": {"key": "value"}}, [1, 2, 3], false]`},
					{`[null, "null_as_string", false, 0]`},
					{`["a", "b", "c", {"nested": {"key": "value"}}]`},
					{`["flat", "array", "of", "strings"]`},
					{`["simple", "array", "of", "strings"]`},
					{`[-1, -2, -3, -4]`},
					{`[0.1, 0.2, 0.3, 0.4]`},
					{`[true, null, "string", 3.14]`},
					{`[true, false, null, "end"]`},
					{`[{"key": {"subkey": [1, 2, 3]}}, 42, "text", false]`},
					{`[null, 1, "two", true, {"five": 5}]`},
					{`["end", "of", "array", 123, true]`},
					{`["mixed", 123, false, null, {"complex": {"key": "value"}}]`},
					{`["multiple", "types", 123, true, {"key": "value"}]`},
					{`["text", 123, false, {"key": "value"}, [1, 2, 3]]`},
					{`["varied", "types", true, 123, {"key": "value"}]`},
					{`[1, "two", true, null, [1, 2, 3]]`},
					{`[1, "two", true, null, {"key": "value"}]`},
					{`[1, 2, 3, 4, 5]`},
					{`[1.1, 2.2, 3.3, 4.4, 5.5]`},
					{`[true, false, true, false, null]`},
					{`{"array_of_arrays": {"array1": [1, 2, 3], "array2": [4, 5, 6], "array3": [7, 8, 9]}}`},
					{`{"array_of_bools": [true, false, true]}`},
					{`{"array_of_numbers": [1, 2, 3]}`},
					{`{"array_of_objects": [{"key1": "value1"}, {"key2": "value2"}, {"key3": "value3"}]}`},
					{`{"array_of_strings": ["one", "two", "three"]}`},
					{`{"boolean_mixed": {"null": null, "true": true, "false": false}}`},
					{`{"combos": [{"number": 1}, {"string": "two"}, {"boolean": true}]}`},
					{`{"common_key": "different_value"}`},
					{`{"common_key": "same_value"}`},
					{`{"common_key": 123}`},
					{`{"common_key": 456}`},
					{`{"common_key": {"nested_key": "different_value"}}`},
					{`{"common_key": {"nested_key": "value"}}`},
					{`{"complex_array": {"array1": [1, 2, 3], "array2": ["a", "b", "c"]}}`},
					{`{"complex_object": {"key1": {"subkey": "value1"}, "key2": {"subkey": "value2"}}}`},
					{`{"deep_nesting": {"level1": {"level2": {"level3": {"key": "value"}}}}}`},
					{`{"empty": {}}`},
					{`{"mixed": {"bool": false, "null": null, "number": 100, "string": "text"}}`},
					{`{"mixed_data": {"bool": false, "null": null, "number": 100, "string": "text"}}`},
					{`{"nested": {"level1": {"level2": {"key": "deep_value"}}}}`},
					{`{"nested_mixed": {"key1": 1, "key2": [true, false], "key3": {"inner_key": "inner_value"}}}`},
					{`{"nested_numbers": {"one": 1, "two": 2, "three": 3}}`},
					{`{"nested_object": {"subkey1": 789, "subkey2": [true, false], "subkey3": {"deep": "value"}}}`},
					{`{"nested_string": {"outer": {"inner": "text"}}}`},
					{`{"number_key": {"float": 2.3, "integer": 1, "negative": -1}}`},
					{`{"object_in_array": [{"key": "value"}, {"another": "one"}]}`},
					{`{"object_in_array": {"array": [1, 2, 3], "nested": {"key": "value"}}}`},
					{`{"random_mixed": {"null": null, "number": 1, "string": "two", "boolean": true}}`},
					{`{"random_text": "Lorem ipsum dolor sit amet"}`},
					{`{"simple": "object"}`},
					{`{"simple_object": {"key": "value"}}`},
					{`{"single_bool": true}`},
					{`{"single_number": 42}`},
					{`{"single_string": "hello"}`},
					{`{"string_with_numbers": {"key": "123abc", "another_key": "456def"}}`},
					{`{"unicode_chars": {"char1": "あ", "char2": "い", "char3": "う"}}`},
					{`{"unicode_string": {"greeting": "你好"}}`},
					{`{"emoji": "😊", "unicode": "こんにちは"}`},
					{`{"boolean": true, "empty_array": [], "empty_object": {}}`},
					{`{"key1": "value1", "key2": "value2", "key3": "value3"}`},
					{`{"key1": 123, "key2": "duplicate_key", "common_key": "same_value"}`},
					{`{"key1": 123, "key2": "value", "key3": true}`},
					{`{"keyX": "string_value", "keyY": 123.456, "keyZ": null}`},
					{`{"null_value": null, "boolean_true": true, "boolean_false": false}`},
					{`{"numbers": [1, 2, 3], "strings": ["a", "b", "c"], "booleans": [true, false]}`},
					{`{"text": "simple_string", "float": 3.14, "integer": 123}`},
					{`{"double": 2, "single": 1, "triple": 3, "quadruple": 4}`},
					{`{"keyA": 456, "keyB": "another_value", "keyC": false, "keyD": [1, 2, 3]}`},
					{`{"key1": 1, "key2": 2, "key3": 3, "key4": 4, "key5": 5}`},
				},
			},
		},
	},
	{
		Name: "Line type",
		Skip: true,
		SetUpScript: []string{
			"CREATE TABLE t_line (id INTEGER primary key, v1 LINE);",
			"INSERT INTO t_line VALUES (1, '{1,2,3}'), (2, '{4,5,6}');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_line ORDER BY id;",
				Expected: []sql.Row{
					{1, "{1,2,3}"},
					{2, "{4,5,6}"},
				},
			},
		},
	},
	{
		Name: "Lseg type",
		Skip: true,
		SetUpScript: []string{
			"CREATE TABLE t_lseg (id INTEGER primary key, v1 LSEG);",
			"INSERT INTO t_lseg VALUES (1, '((1,2),(3,4))'), (2, '((5,6),(7,8))');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_lseg ORDER BY id;",
				Expected: []sql.Row{
					{1, "((1,2),(3,4))"},
					{2, "((5,6),(7,8))"},
				},
			},
		},
	},
	{
		Name: "Macaddr type",
		Skip: true,
		SetUpScript: []string{
			"CREATE TABLE t_macaddr (id INTEGER primary key, v1 MACADDR);",
			"INSERT INTO t_macaddr VALUES (1, '08:00:2b:01:02:03'), (2, '00:11:22:33:44:55');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_macaddr ORDER BY id;",
				Expected: []sql.Row{
					{1, "08:00:2b:01:02:03"},
					{2, "00:11:22:33:44:55"},
				},
			},
		},
	},
	{
		Name: "Money type",
		Skip: true,
		SetUpScript: []string{
			"CREATE TABLE t_money (id INTEGER primary key, v1 MONEY);",
			"INSERT INTO t_money VALUES (1, '$100.25'), (2, '$50.50');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_money ORDER BY id;",
				Expected: []sql.Row{
					{1, "$100.25"},
					{2, "$50.50"},
				},
			},
		},
	},
	{
		Name: "Name type",
		SetUpScript: []string{
			"CREATE TABLE t_name (id INTEGER primary key, v1 NAME);",
			"INSERT INTO t_name VALUES (1, 'abcdefghij'), (2, 'klmnopqrst');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_name ORDER BY id;",
				Expected: []sql.Row{
					{1, "abcdefghij"},
					{2, "klmnopqrst"},
				},
			},
			{
				Query: "SELECT * FROM t_name ORDER BY v1 DESC;",
				Expected: []sql.Row{
					{2, "klmnopqrst"},
					{1, "abcdefghij"},
				},
			},
			{
				Query: "SELECT v1::char(1) FROM t_name WHERE v1='klmnopqrst';",
				Expected: []sql.Row{
					{"k"},
				},
			},
			{
				Query:    "UPDATE t_name SET v1='tuvwxyz' WHERE id=2;",
				Expected: []sql.Row{},
			},
			{
				Query:    "DELETE FROM t_name WHERE v1='abcdefghij';",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT id::name, v1::text FROM t_name ORDER BY id;",
				Expected: []sql.Row{
					{"2", "tuvwxyz"},
				},
			},
			{
				Query:    "INSERT INTO t_name VALUES (3, '0123456789012345678901234567890123456789012345678901234567890123456789');",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT * FROM t_name ORDER BY id;",
				Expected: []sql.Row{
					{2, "tuvwxyz"},
					{3, "012345678901234567890123456789012345678901234567890123456789012"},
				},
			},
			{
				Query:    "INSERT INTO t_name VALUES (4, 12345);",
				Skip:     true, // TODO: according to casting rules this shouldn't work but it does, investigate why
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT * FROM t_name ORDER BY id;",
				Skip:  true, // This is skipped because the one above is skipped
				Expected: []sql.Row{
					{2, "tuvwxyz"},
					{3, "012345678901234567890123456789012345678901234567890123456789012"},
					{4, "12345"},
				},
			},
			{
				Query:    `SELECT name 'name string' = name 'name string' AS "True";`,
				Expected: []sql.Row{{"t"}},
			},
		},
	},
	{
		Name: "Name key",
		SetUpScript: []string{
			"CREATE TABLE t_name (id NAME primary key, v1 NAME);",
			"INSERT INTO t_name VALUES ('wxyz', 'abcdefghij'), ('abcd', 'klmnopqrst');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_name WHERE id = 'wxyz' ORDER BY id;",
				Expected: []sql.Row{
					{"wxyz", "abcdefghij"},
				},
			},
		},
	},
	{
		Name: "Name type, explicit casts",
		SetUpScript: []string{
			"CREATE TABLE t_name (id INTEGER primary key, v1 NAME);",
			"INSERT INTO t_name VALUES (1, 'abcdefghij'), (2, '12345');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_name ORDER BY id;",
				Expected: []sql.Row{
					{1, "abcdefghij"},
					{2, "12345"},
				},
			},
			// Cast from Name to types
			{
				Query: "SELECT v1::char(1), v1::varchar(2), v1::text FROM t_name WHERE id=1;",
				Expected: []sql.Row{
					{"a", "ab", "abcdefghij"},
				},
			},
			{
				Query: "SELECT v1::smallint, v1::integer, v1::bigint, v1::float4, v1::float8, v1::numeric FROM t_name WHERE id=2;",
				Expected: []sql.Row{
					{12345, 12345, 12345, float64(12345), float64(12345), Numeric("12345")},
				},
			},
			{
				Query: "SELECT v1::oid, v1::xid FROM t_name WHERE id=2;",
				Expected: []sql.Row{
					{12345, 12345},
				},
			},
			{
				Query: "SELECT v1::xid FROM t_name WHERE id=1;",
				Expected: []sql.Row{
					{0},
				},
			},
			{
				Query: "SELECT ('0'::name)::boolean, ('1'::name)::boolean;",
				Expected: []sql.Row{
					{"f", "t"},
				},
			},
			{
				Query:       "SELECT v1::smallint FROM t_name WHERE id=1;",
				ExpectedErr: "invalid input syntax for type",
			},
			{
				Query:       "SELECT v1::integer FROM t_name WHERE id=1;",
				ExpectedErr: "invalid input syntax for type",
			},
			{
				Query:       "SELECT v1::bigint FROM t_name WHERE id=1;",
				ExpectedErr: "invalid input syntax for type",
			},
			{
				Query:       "SELECT v1::float4 FROM t_name WHERE id=1;",
				ExpectedErr: "invalid input syntax for type",
			},
			{
				Query:       "SELECT v1::float8 FROM t_name WHERE id=1;",
				ExpectedErr: "invalid input syntax for type",
			},
			{
				Query:       "SELECT v1::numeric FROM t_name WHERE id=1;",
				ExpectedErr: "invalid input syntax for type",
			},
			{
				Query:       "SELECT v1::boolean FROM t_name WHERE id=1;",
				ExpectedErr: "invalid input syntax for type",
			},
			{
				Query:       "SELECT v1::oid FROM t_name WHERE id=1;",
				ExpectedErr: "invalid input syntax for type",
			},
			// Cast to Name from types
			{
				Query: "SELECT ('abc'::char(3))::name, ('abc'::varchar)::name, ('abc'::text)::name;",
				Expected: []sql.Row{
					{"abc", "abc", "abc"},
				},
			},
			{
				Query: "SELECT (10::int2)::name, (100::int4)::name, (1000::int8)::name;",
				Expected: []sql.Row{
					{"10", "100", "1000"},
				},
			},
			{
				Query: "SELECT (1.1::float4)::name, (10.1::float8)::name;",
				Expected: []sql.Row{
					{"1.1", "10.1"},
				},
			},
			{
				Query: "SELECT (100.0::numeric)::name;",
				Skip:  true, // TODO: Should return 100.0 instead of 100
				Expected: []sql.Row{
					{"100.0"},
				},
			},
			{
				Query: "SELECT false::name, true::name, ('0'::boolean)::name, ('1'::boolean)::name;",
				Expected: []sql.Row{
					{"f", "t", "f", "t"},
				},
			},
			{
				Query: "SELECT ('123'::xid)::name, (123::oid)::name;",
				Expected: []sql.Row{
					{"123", "123"},
				},
			},
		},
	},
	{
		Name: "Name array type",
		SetUpScript: []string{
			"CREATE TABLE t_namea (id INTEGER primary key, v1 NAME[], v2 CHARACTER(100), v3 BOOLEAN);",
			"INSERT INTO t_namea VALUES (1, ARRAY['ab''cdef', 'what', 'is,hi', 'wh\"at'], '1234567890123456789012345678901234567890123456789012345678901234567890', true);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: `SELECT v1::varchar(1)[] FROM t_namea;`,
				Expected: []sql.Row{
					{"{a,w,i,w}"},
				},
			},
			{
				Query: `SELECT v2::name, v3::name FROM t_namea;`,
				Expected: []sql.Row{
					{"123456789012345678901234567890123456789012345678901234567890123", "t"},
				},
			},
		},
	},
	{
		Name: "Numeric type",
		SetUpScript: []string{
			"CREATE TABLE t_numeric (id INTEGER primary key, v1 NUMERIC(5,2));",
			"INSERT INTO t_numeric VALUES (1, 123.45), (2, 67.89), (3, 100.3);",
			"CREATE TABLE fract_only (id int, val numeric(4,4));",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_numeric ORDER BY id;",
				Expected: []sql.Row{
					{1, Numeric("123.45")},
					{2, Numeric("67.89")},
					{3, Numeric("100.30")},
				},
			},
			{
				Query:    "INSERT INTO fract_only VALUES (1, '0.0');",
				Expected: []sql.Row{},
			},
			{
				Query:    "SELECT numeric '10.00';",
				Expected: []sql.Row{{Numeric("10.00")}},
			},
			{
				Query:    "SELECT numeric '-10.00';",
				Expected: []sql.Row{{Numeric("-10.00")}},
			},
			{
				Query:    "select 0.03::numeric(3,3);",
				Expected: []sql.Row{{Numeric("0.030")}},
			},
			{
				Query:       "select 1.03::numeric(2,2);",
				ExpectedErr: `numeric field overflow`,
			},
			{
				Query:       "select 1.03::float4::numeric(2,2);",
				ExpectedErr: `numeric field overflow`,
			},
		},
	},
	{
		Name: "Numeric key",
		SetUpScript: []string{
			"CREATE TABLE t_numeric (id numeric(5,2) primary key, v1 NUMERIC(5,2));",
			"INSERT INTO t_numeric VALUES (123.45, 67.89), (67.89, 100.3);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_numeric;",
				Skip:  true, // test setup problem, values are logically equivalent but don't match
				Expected: []sql.Row{
					{Numeric("123.45"), Numeric("67.89")},
					{Numeric("67.89"), Numeric("100.3")},
				},
			},
			{
				Query: "SELECT * FROM t_numeric WHERE ID = 123.45 ORDER BY id;",
				Skip:  true, // value not found
				Expected: []sql.Row{
					{Numeric("123.45"), Numeric("67.89")},
				},
			},
		},
	},
	{
		Name: "Numeric type, no scale or precision",
		SetUpScript: []string{
			"CREATE TABLE t_numeric (id INTEGER primary key, v1 NUMERIC);",
			"INSERT INTO t_numeric VALUES (1, 123.45), (2, 67.875), (3, 100.3);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_numeric ORDER BY id;",
				Expected: []sql.Row{
					{1, Numeric("123.45")},
					{2, Numeric("67.875")},
					{3, Numeric("100.3")},
				},
			},
		},
	},
	{
		Name: "Numeric array type, no scale or precision",
		SetUpScript: []string{
			"CREATE TABLE t_numeric (id INTEGER primary key, v1 NUMERIC[]);",
			"INSERT INTO t_numeric VALUES (1, ARRAY[NULL,123.45]), (2, ARRAY[67.89,572903.1468]);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_numeric ORDER BY id;",
				Expected: []sql.Row{
					{1, "{NULL,123.45}"},
					{2, "{67.89,572903.1468}"},
				},
			},
		},
	},
	{
		Name: "Oid type",
		SetUpScript: []string{
			"CREATE TABLE t_oid (id INTEGER primary key, v1 OID);",
			"INSERT INTO t_oid VALUES (1, 1234), (2, 5678);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_oid ORDER BY id;",
				Expected: []sql.Row{
					{1, 1234},
					{2, 5678},
				},
			},
			{
				Query: "SELECT * FROM t_oid ORDER BY v1 DESC;",
				Expected: []sql.Row{
					{2, 5678},
					{1, 1234},
				},
			},
			{
				Query:    "UPDATE t_oid SET v1=9012 WHERE id=2;",
				Expected: []sql.Row{},
			},
			{
				Query:    "DELETE FROM t_oid WHERE v1=1234;",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT * FROM t_oid ORDER BY id;",
				Expected: []sql.Row{
					{2, 9012},
				},
			},
			{
				Query:    "INSERT INTO t_oid VALUES (3, '2345');",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT * FROM t_oid ORDER BY id;",
				Expected: []sql.Row{
					{2, 9012},
					{3, 2345},
				},
			},
			{
				Query:    "INSERT INTO t_oid VALUES (4, 4294967295);",
				Expected: []sql.Row{},
			},
			{
				Query:       "INSERT INTO t_oid VALUES (5, 4294967296);",
				ExpectedErr: "out of range",
			},
			{
				Query:    "INSERT INTO t_oid VALUES (6, 0);",
				Expected: []sql.Row{},
			},
			{
				Query:    "INSERT INTO t_oid VALUES (7, -1);",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT * FROM t_oid ORDER BY id;",
				Expected: []sql.Row{
					{2, 9012},
					{3, 2345},
					{4, 4294967295},
					{6, 0},
					{7, 4294967295},
				},
			},
			{
				Query:    "select oid '20304';",
				Expected: []sql.Row{{20304}},
			},
		},
	},
	{
		Name: "Oid type, explicit casts",
		SetUpScript: []string{
			"CREATE TABLE t_oid (id INTEGER primary key, coid OID);",
			"INSERT INTO t_oid VALUES (1, 1234), (2, 4294967295);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_oid ORDER BY id;",
				Expected: []sql.Row{
					{1, 1234},
					{2, 4294967295},
				},
			},
			// Cast from OID to types
			{
				Query: "SELECT coid::char(1) FROM t_oid WHERE id=1;",
				Expected: []sql.Row{
					{"1"},
				},
			},
			{
				Query: "SELECT coid::varchar(2) FROM t_oid WHERE id=1;",
				Expected: []sql.Row{
					{"12"},
				},
			},
			{
				Query: "SELECT coid::text FROM t_oid WHERE id=1;",
				Expected: []sql.Row{
					{"1234"},
				},
			},
			{
				Query:       "SELECT coid::smallint FROM t_oid WHERE id=1;",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT coid::smallint FROM t_oid WHERE id=2;",
				ExpectedErr: "does not exist",
			},
			{
				Query: "SELECT coid::integer FROM t_oid WHERE id=1;",
				Expected: []sql.Row{
					{1234},
				},
			},
			{
				Query: "SELECT coid::integer FROM t_oid WHERE id=2;",
				Expected: []sql.Row{
					{-1},
				},
			},
			{
				Query: "SELECT coid::bigint FROM t_oid WHERE id=1;",
				Expected: []sql.Row{
					{1234},
				},
			},
			{
				Query: "SELECT coid::name FROM t_oid WHERE id=1;",
				Expected: []sql.Row{
					{"1234"},
				},
			},
			{
				Query: "SELECT coid::bigint FROM t_oid WHERE id=2;",
				Expected: []sql.Row{
					{4294967295},
				},
			},
			{
				Query:       "SELECT coid::float4 FROM t_oid WHERE id=1;",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT coid::float8 FROM t_oid WHERE id=1;",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT coid::numeric FROM t_oid WHERE id=1;",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT coid::xid FROM t_oid WHERE id=1;",
				ExpectedErr: "does not exist",
			},
			// Cast to OID from types
			{
				Query: "SELECT ('123'::char(3))::oid, ('123'::varchar)::oid, ('0'::text)::oid, ('400'::name)::oid;",
				Expected: []sql.Row{
					{123, 123, 0, 400},
				},
			},
			{
				Query: "SELECT ('-1'::char(3))::oid, ('-1'::varchar)::oid, ('-1'::text)::oid, ('-1'::name)::oid;",
				Expected: []sql.Row{
					{4294967295, 4294967295, 4294967295, 4294967295},
				},
			},
			{
				Query: "SELECT ('-2147483648'::char(11))::oid, ('-2147483648'::varchar)::oid, ('-2147483648'::text)::oid, ('-2147483648'::name)::oid;",
				Expected: []sql.Row{
					{2147483648, 2147483648, 2147483648, 2147483648},
				},
			},
			{
				Query: "SELECT (10::int2)::oid, (10::int4)::oid, (100::int8)::oid;",
				Expected: []sql.Row{
					{10, 10, 100},
				},
			},
			{
				Query: "SELECT (-1::int2)::oid, (-1::int4)::oid;",
				Expected: []sql.Row{
					{4294967295, 4294967295},
				},
			},
			{
				Query:       "SELECT (-1::int8)::oid;",
				ExpectedErr: "out of range",
			},
			{
				Query:       "SELECT (922337203685477580::int8)::oid;",
				ExpectedErr: "out of range",
			},
			{
				Query:       "SELECT (1.1::float4)::oid;",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT (1.1::float8)::oid;",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT (1.1::decimal)::oid;",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT ('922337203685477580'::text)::oid;",
				ExpectedErr: "out of range",
			},
			{
				Query:       "SELECT ('abc'::char(3))::oid;",
				ExpectedErr: "invalid input syntax",
			},
			{
				Query:       "SELECT ('-2147483649'::char(11))::oid;",
				ExpectedErr: "out of range",
			},
			{
				Query:       "SELECT ('-2147483649'::varchar)::oid;",
				ExpectedErr: "out of range",
			},
			{
				Query:       "SELECT ('-2147483649'::text)::oid;",
				ExpectedErr: "out of range",
			},
			{
				Query:       "SELECT ('-2147483649'::name)::oid;",
				ExpectedErr: "out of range",
			},
		},
	},
	{
		Name: "Oid array type",
		SetUpScript: []string{
			"CREATE TABLE t_oid (id INTEGER primary key, v1 OID[], v2 CHARACTER(100), v3 BOOLEAN);",
			"INSERT INTO t_oid VALUES (1, ARRAY[123, 456, 789, 101], '1234567890', true);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: `SELECT v1::varchar(1)[] FROM t_oid;`,
				Expected: []sql.Row{
					{"{1,4,7,1}"},
				},
			},
			{
				Query:       `SELECT v2::oid, v3::oid FROM t_oid;`,
				ExpectedErr: "cast",
			},
		},
	},
	{
		Name: "Path type",
		Skip: true,
		SetUpScript: []string{
			"CREATE TABLE t_path (id INTEGER primary key, v1 PATH);",
			"INSERT INTO t_path VALUES (1, '((1,2),(3,4),(5,6))'), (2, '((7,8),(9,10),(11,12))');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_path ORDER BY id;",
				Expected: []sql.Row{
					{1, "((1,2),(3,4),(5,6))"},
					{2, "((7,8),(9,10),(11,12))"},
				},
			},
		},
	},
	{
		Name: "Pg_lsn type",
		Skip: true,
		SetUpScript: []string{
			"CREATE TABLE t_pg_lsn (id INTEGER primary key, v1 PG_LSN);",
			"INSERT INTO t_pg_lsn VALUES (1, '16/B8E36C60'), (2, '16/B8E36C70');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_pg_lsn ORDER BY id;",
				Expected: []sql.Row{
					{1, "16/B8E36C60"},
					{2, "16/B8E36C70"},
				},
			},
		},
	},
	{
		Name: "Point type",
		Skip: true,
		SetUpScript: []string{
			"CREATE TABLE t_point (id INTEGER primary key, v1 POINT);",
			"INSERT INTO t_point VALUES (1, '(1,2)'), (2, '(3,4)');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_point ORDER BY id;",
				Expected: []sql.Row{
					{1, "(1,2)"},
					{2, "(3,4)"},
				},
			},
		},
	},
	{
		Name: "Polygon type",
		Skip: true,
		SetUpScript: []string{
			"CREATE TABLE t_polygon (id INTEGER primary key, v1 POLYGON);",
			"INSERT INTO t_polygon VALUES (1, '((1,2),(3,4),(5,6))'), (2, '((7,8),(9,10),(11,12))');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_polygon ORDER BY id;",
				Expected: []sql.Row{
					{1, "((1,2),(3,4),(5,6))"},
					{2, "((7,8),(9,10),(11,12))"},
				},
			},
		},
	},
	{
		Name: "Real type",
		SetUpScript: []string{
			"CREATE TABLE t_real (id INTEGER primary key, v1 REAL);",
			"INSERT INTO t_real VALUES (1, 123.875), (2, 67.125);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_real ORDER BY id;",
				Expected: []sql.Row{
					{1, 123.875},
					{2, 67.125},
				},
			},
		},
	},
	{
		Name: "Real key",
		SetUpScript: []string{
			"CREATE TABLE t_real (id REAL primary key, v1 REAL);",
			"INSERT INTO t_real VALUES (123.875, 67.125), (67.125, 123.875);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_real WHERE ID = 123.875 ORDER BY id;",
				Expected: []sql.Row{
					{123.875, 67.125},
				},
			},
		},
	},
	{
		Name: "Real array type",
		SetUpScript: []string{
			"CREATE TABLE t_real (id INTEGER primary key, v1 REAL[]);",
			"INSERT INTO t_real VALUES (1, ARRAY[NULL,123.875]), (2, ARRAY[67.125, 84256]);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_real ORDER BY id;",
				Expected: []sql.Row{
					{1, "{NULL,123.875}"},
					{2, "{67.125,84256}"},
				},
			},
		},
	},
	{
		Name: "Regclass type",
		SetUpScript: []string{
			`CREATE TABLE testing (pk INT primary key, v1 INT UNIQUE);`,
			`CREATE TABLE "Testing2" (pk INT primary key, v1 INT);`,
			`CREATE VIEW testview AS SELECT * FROM testing LIMIT 1;`,
			`CREATE SEQUENCE seq1;`,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: `SELECT 'testing'::regclass;`,
				Expected: []sql.Row{
					{"testing"},
				},
			},
			{
				Query: `SELECT 'public.testing'::regclass;`,
				Expected: []sql.Row{
					{"testing"},
				},
			},
			{
				Query: `SELECT 'postgres.public.testing'::regclass;`,
				Expected: []sql.Row{
					{"testing"},
				},
			},
			{
				Query:       `SELECT 'doesnotexist.public.testing'::regclass;`,
				ExpectedErr: "database not found",
			},
			{
				Query: `SELECT 'testview'::regclass;`,
				Expected: []sql.Row{
					{"testview"},
				},
			},
			{
				Query: `SELECT ' testing'::regclass;`,
				Expected: []sql.Row{
					{"testing"},
				},
			},
			{
				Query: `SELECT 'seq1'::regclass;`,
				Expected: []sql.Row{
					{"seq1"},
				},
			},
			{
				Query:       `SELECT 'Testing2'::regclass;`,
				ExpectedErr: "does not exist",
			},
			{
				Query: `SELECT '"Testing2"'::regclass;`,
				Expected: []sql.Row{
					{"Testing2"},
				},
			},
			{ // This tests that an invalid OID returns itself in string form
				Query: `SELECT 4294967295::regclass;`,
				Expected: []sql.Row{
					{"4294967295"},
				},
			},
			{
				Query: "SELECT relname FROM pg_catalog.pg_class WHERE oid = 'testing'::regclass;",
				Skip:  true, // panic converting string to regclass
				Expected: []sql.Row{
					{"testing"},
				},
			},
			{
				// schema-qualified relation names are not returned if the schema is on the search path
				Query: `SELECT 'public.testing'::regclass, 'public.seq1'::regclass, 'public.testview'::regclass, 'public.testing_pkey'::regclass;`,
				Expected: []sql.Row{
					{"testing", "seq1", "testview", "testing_pkey"},
				},
			},
			{
				// Clear out the current search_path setting to test schema-qualified relation names
				Query:    `SET search_path = '';`,
				Expected: []sql.Row{},
			},
			{
				// Without 'public' on search_path, we get a does not exist error
				Query:       `SELECT 'testing'::regclass;`,
				ExpectedErr: "does not exist",
			},
			{
				Query: `SELECT 'public.testing'::regclass, 'public.seq1'::regclass, 'public.testview'::regclass, 'public.testing_pkey'::regclass;`,
				Expected: []sql.Row{
					{"public.testing", "public.seq1", "public.testview", "public.testing_pkey"},
				},
			},
		},
	},
	{
		Name: "Regproc type",
		Assertions: []ScriptTestAssertion{
			{
				Query: `SELECT 'acos'::regproc;`,
				Expected: []sql.Row{
					{"acos"},
				},
			},
			{
				Query: `SELECT ' acos'::regproc;`,
				Expected: []sql.Row{
					{"acos"},
				},
			},
			{
				Query: `SELECT '"acos"'::regproc;`,
				Expected: []sql.Row{
					{"acos"},
				},
			},
			{ // This tests that a raw OID properly converts
				Query: `SELECT (('acos'::regproc)::oid)::regproc;`,
				Expected: []sql.Row{
					{"acos"},
				},
			},
			{ // This tests that a string representing a raw OID converts the same as a raw OID
				Query: `SELECT ((('acos'::regproc)::oid)::text)::regproc;`,
				Expected: []sql.Row{
					{"acos"},
				},
			},
			{ // This tests that an invalid OID returns itself in string form
				Query: `SELECT 4294967295::regproc;`,
				Expected: []sql.Row{
					{"4294967295"},
				},
			},
			{
				Query:       `SELECT '"Abs"'::regproc;`,
				ExpectedErr: "does not exist",
			},
			{
				Query:       `SELECT '"acos'::regproc;`,
				ExpectedErr: "invalid name syntax",
			},
			{
				Query:       `SELECT 'acos"'::regproc;`,
				ExpectedErr: "does not exist",
			},
			{
				Query:       `SELECT '""acos'::regproc;`,
				ExpectedErr: "invalid name syntax",
			},
		},
	},
	{
		Name: "Regtype type",
		Assertions: []ScriptTestAssertion{
			{
				Skip:             true, // TODO: Column should be regtype, not "integer"
				Query:            `SELECT 'integer'::regtype;`,
				ExpectedColNames: []string{"regtype"},
				Expected: []sql.Row{
					{"integer"},
				},
			},
			{
				Query: `SELECT 'integer'::regtype;`,
				Expected: []sql.Row{
					{"integer"},
				},
			},
			{
				Query: `SELECT 'integer[]'::regtype;`,
				Expected: []sql.Row{
					{"integer[]"},
				},
			},
			{
				Query: `SELECT 'int4'::regtype;`,
				Expected: []sql.Row{
					{"integer"},
				},
			},
			{
				Query: `SELECT 'float8'::regtype;`,
				Expected: []sql.Row{
					{"double precision"},
				},
			},
			{
				Query: `SELECT 'character varying'::regtype;`,
				Expected: []sql.Row{
					{"character varying"},
				},
			},
			{
				Query: `SELECT '"char"'::regtype;`,
				Expected: []sql.Row{
					{`"char"`},
				},
			},
			{
				Query: `SELECT 'char'::regtype;`,
				Expected: []sql.Row{
					{"character"},
				},
			},
			{
				Query: `SELECT 'char(10)'::regtype;`,
				Expected: []sql.Row{
					{"character"},
				},
			},
			{
				Query: `SELECT '"char"'::regtype::oid;`,
				Expected: []sql.Row{
					{18},
				},
			},
			{
				Query: `SELECT 'char'::regtype::oid;`,
				Expected: []sql.Row{
					{1042},
				},
			},
			{
				Query: `SELECT '"char"[]'::regtype;`,
				Expected: []sql.Row{
					{"\"char\"[]"},
				},
			},
			{
				Query: `SELECT ' integer'::regtype;`,
				Expected: []sql.Row{
					{"integer"},
				},
			},
			{
				Query: `SELECT '"integer"'::regtype;`,
				Expected: []sql.Row{
					{"integer"},
				},
			},
			{ // This tests that a raw OID properly converts
				Query: `SELECT (('integer'::regtype)::oid)::regtype;`,
				Expected: []sql.Row{
					{"integer"},
				},
			},
			{ // This tests that a string representing a raw OID converts the same as a raw OID
				Query: `SELECT ((('integer'::regtype)::oid)::text)::regtype;`,
				Expected: []sql.Row{
					{"integer"},
				},
			},
			{ // This tests that an invalid OID returns itself in string form
				Query: `SELECT 4294967295::regtype;`,
				Expected: []sql.Row{
					{"4294967295"},
				},
			},
			{
				Query:       `SELECT '"Integer"'::regtype;`,
				ExpectedErr: "does not exist",
			},
			{
				Query:       `SELECT '"integer'::regtype;`,
				ExpectedErr: "invalid name syntax",
			},
			{
				Query:       `SELECT 'integer"'::regtype;`,
				ExpectedErr: "does not exist",
			},
			{
				Query:       `SELECT '""integer'::regtype;`,
				ExpectedErr: "invalid name syntax",
			},
		},
	},
	{
		Name: "Smallint type",
		SetUpScript: []string{
			"CREATE TABLE t_smallint (id INTEGER primary key, v1 SMALLINT);",
			"INSERT INTO t_smallint VALUES (1, 42), (2, 99);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_smallint ORDER BY id;",
				Expected: []sql.Row{
					{1, 42},
					{2, 99},
				},
			},
		},
	},
	{
		Name: "Smallint key",
		SetUpScript: []string{
			"CREATE TABLE t_smallint (id smallint primary key, v1 SMALLINT);",
			"INSERT INTO t_smallint VALUES (1, 42), (2, 99);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_smallint WHERE ID = 1 ORDER BY id;",
				Expected: []sql.Row{
					{1, 42},
				},
			},
		},
	},
	{
		Name: "Smallint array type",
		SetUpScript: []string{
			"CREATE TABLE t_smallint (id INTEGER primary key, v1 SMALLINT[]);",
			"INSERT INTO t_smallint VALUES (1, ARRAY[42,NULL]), (2, ARRAY[99,126]);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_smallint ORDER BY id;",
				Expected: []sql.Row{
					{1, "{42,NULL}"},
					{2, "{99,126}"},
				},
			},
		},
	},
	{
		Name: "Smallserial type",
		SetUpScript: []string{
			"CREATE TABLE t_smallserial (id SERIAL primary key, v1 SMALLSERIAL);",
			"INSERT INTO t_smallserial (v1) VALUES (42), (99);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_smallserial ORDER BY id;",
				Expected: []sql.Row{
					{1, 42},
					{2, 99},
				},
			},
		},
	},
	{
		Name: "Smallserial key",
		SetUpScript: []string{
			"CREATE TABLE t_smallserial (id smallserial primary key, v1 SMALLSERIAL);",
			"INSERT INTO t_smallserial (v1) VALUES (42), (99);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_smallserial WHERE ID = 1 ORDER BY id;",
				Expected: []sql.Row{
					{1, 42},
				},
			},
		},
	},
	{
		Name: "Serial type",
		SetUpScript: []string{
			"CREATE TABLE t_serial (id SERIAL primary key, v1 SERIAL);",
			"INSERT INTO t_serial (v1) VALUES (123), (456);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_serial ORDER BY id;",
				Expected: []sql.Row{
					{1, 123},
					{2, 456},
				},
			},
			{
				Query: "SELECT * FROM t_serial WHERE ID = 2 ORDER BY id;",
				Expected: []sql.Row{
					{2, 456},
				},
			},
		},
	},
	{
		Name: "Text type",
		SetUpScript: []string{
			// Test a table with a TEXT column
			"CREATE TABLE t_text (id INTEGER primary key, v1 TEXT);",
			"INSERT INTO t_text VALUES (1, 'Hello'), (2, 'World'), (3, ''), (4, NULL);",

			// Test a table created with a TEXT column in a unique, secondary index
			"CREATE TABLE t_text_unique (id INTEGER primary key, v1 TEXT, v2 TEXT NOT NULL UNIQUE);",
			"INSERT INTO t_text_unique VALUES (1, 'Hello', 'Bonjour'), (2, 'World', 'tout le monde'), (3, '', ''), (4, NULL, '!');",
		},
		Assertions: []ScriptTestAssertion{
			{
				// Use the text keyword to cast
				Query:    `SELECT text 'text' || ' and unknown';`,
				Expected: []sql.Row{{"text and unknown"}},
			},
			{
				// Use the text keyword to cast
				Query:    `SELECT text 'this is a text string' = text 'this is a text string' AS true;`,
				Expected: []sql.Row{{"t"}},
			},
			{
				// Basic select from a table with a TEXT column
				Query: "SELECT * FROM t_text ORDER BY id;",
				Expected: []sql.Row{
					{1, "Hello"},
					{2, "World"},
					{3, ""},
					{4, nil},
				},
			},
			{
				// Create a unique, secondary index on a TEXT column
				Query:    "CREATE UNIQUE INDEX v1_unique ON t_text(v1);",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT * FROM t_text WHERE v1 = 'World';",
				Skip:  true, // text indexes are broken
				Expected: []sql.Row{
					{2, "World"},
				},
			},
			{
				// Test the new unique constraint on the TEXT column
				Query:       "INSERT INTO t_text VALUES (5, 'World');",
				ExpectedErr: "unique",
			},
			{
				Query: "SELECT * FROM t_text_unique WHERE v2 = '!';",
				Expected: []sql.Row{
					{4, nil, "!"},
				},
			},
			{
				Query: "SELECT * FROM t_text_unique WHERE v2 >= '!' ORDER BY v2;",
				Expected: []sql.Row{
					{4, nil, "!"},
					{1, "Hello", "Bonjour"},
					{2, "World", "tout le monde"},
				},
			},
			{
				// Test ordering by TEXT column in a secondary index
				Query: "SELECT * FROM t_text_unique ORDER BY v2;",
				Expected: []sql.Row{
					{3, "", ""},
					{4, nil, "!"},
					{1, "Hello", "Bonjour"},
					{2, "World", "tout le monde"},
				},
			},
			{
				Query: "SELECT * FROM t_text_unique ORDER BY id;",
				Expected: []sql.Row{
					{1, "Hello", "Bonjour"},
					{2, "World", "tout le monde"},
					{3, "", ""},
					{4, nil, "!"},
				},
			},
			{
				Query:       "INSERT INTO t_text_unique VALUES (5, 'Another', 'Bonjour');",
				ExpectedErr: "unique",
			},
			{
				// Create a secondary index over multiple text fields
				Query:    "CREATE INDEX on t_text_unique(v1, v2);",
				Expected: []sql.Row{},
			},
			{
				Query:    "SELECT id FROM t_text_unique WHERE v1='Hello' and v2='Bonjour';",
				Expected: []sql.Row{{1}},
			},
			{
				// Create a table with a TEXT column to test adding a non-unique, secondary index
				Query:    `CREATE TABLE t2 (pk int primary key, c1 TEXT);`,
				Expected: []sql.Row{},
			},
			{
				Query:    `CREATE INDEX idx1 ON t2(c1);`,
				Expected: []sql.Row{},
			},
			{
				Query:    `INSERT INTO t2 VALUES (1, 'one'), (2, 'two');`,
				Expected: []sql.Row{},
			},
			{
				Query:    `SELECT c1 from t2 order by c1;`,
				Skip:     true, // ordering is broken due to text indexes being broken
				Expected: []sql.Row{{"one"}, {"two"}},
			},
		},
	},
	{
		Name: "Text key",
		Skip: true, // text indexes are broken
		SetUpScript: []string{
			"CREATE TABLE t_text (id TEXT primary key, v1 TEXT);",
			"INSERT INTO t_text VALUES ('Hello', 'World'), ('goodbye', 'cruel world');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_text where id = 'goodbye' ORDER BY id;",
				Expected: []sql.Row{
					{"goodbye", "cruel world"},
				},
			},
		},
	},
	{
		Name: "Time without time zone type",
		SetUpScript: []string{
			"CREATE TABLE t_time_without_zone (id INTEGER primary key, v1 TIME);",
			"INSERT INTO t_time_without_zone VALUES (1, '12:34:56'), (2, '23:45:01'), (3, '02:03 EDT');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_time_without_zone ORDER BY id;",
				Expected: []sql.Row{
					{1, "12:34:56"},
					{2, "23:45:01"},
					{3, "02:03:00"},
				},
			},
			{
				Query: "SELECT v1::interval FROM t_time_without_zone ORDER BY id;",
				Expected: []sql.Row{
					{"12:34:56"},
					{"23:45:01"},
					{"02:03:00"},
				},
			},
			{
				Query: `SELECT '00:00:00'::time;`,
				Expected: []sql.Row{
					{"00:00:00"},
				},
			},
			{
				Query: `SELECT '23:59:59.999999'::time;`,
				Expected: []sql.Row{
					{"23:59:59.999999"},
				},
			},
			{
				Query:    "SELECT time without time zone '040506.789+08';",
				Expected: []sql.Row{{"04:05:06.789"}},
			},
			{
				Query:    `SELECT time '04:05:06' + date '2025-07-21';`,
				Expected: []sql.Row{{"2025-07-21 04:05:06"}},
			},
			{
				Query:    `SELECT time without time zone '04:05:06' + interval '2 minutes';`,
				Expected: []sql.Row{{"04:07:06"}},
			},
		},
	},
	{
		Name: "Time without time zone key",
		SetUpScript: []string{
			"CREATE TABLE t_time_without_zone (id TIME primary key, v1 TIME);",
			"INSERT INTO t_time_without_zone VALUES ('12:34:56', '23:45:01'), ('23:45:01', '12:34:56');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_time_without_zone WHERE ID = '12:34:56' ORDER BY id;",
				Expected: []sql.Row{
					{"12:34:56", "23:45:01"},
				},
			},
		},
	},
	{ // TODO: timezone representation is reported via local time, need to account for that in testing?
		Name: "Time with time zone type",
		SetUpScript: []string{
			"CREATE TABLE t_time_with_zone (id INTEGER primary key, v1 TIME WITH TIME ZONE);",
			"INSERT INTO t_time_with_zone VALUES (1, '12:34:56 UTC'), (2, '23:45:01-0200'), (3, '2025-06-03 02:03 EDT');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_time_with_zone ORDER BY id;",
				Expected: []sql.Row{
					{1, "12:34:56+00"},
					{2, "23:45:01-02"},
					{3, "02:03:00-04"},
				},
			},
			{
				Query:    `SET TIMEZONE TO 'UTC';`,
				Expected: []sql.Row{},
			},
			{
				Query: `SELECT '00:00:00'::timetz;`,
				Expected: []sql.Row{
					{"00:00:00+00"},
				},
			},
			{
				Query:    `SELECT time with time zone '04:05:06 UTC' + date '2025-07-21';`,
				Expected: []sql.Row{{"2025-07-21 04:05:06+00"}},
			},
			{
				Query:    `SELECT time with time zone '04:05:06 UTC' + interval '2 minutes';`,
				Expected: []sql.Row{{"04:07:06+00"}},
			},
			{
				Query:    `SET TIMEZONE TO DEFAULT;`,
				Expected: []sql.Row{},
			},
			{
				Query: `SELECT '00:00:00-07'::timetz;`,
				Expected: []sql.Row{
					{"00:00:00-07"},
				},
			},
		},
	},
	{
		Name: "Timestamp without time zone type",
		SetUpScript: []string{
			"CREATE TABLE t_timestamp_without_zone (id INTEGER primary key, v1 TIMESTAMP);",
			"INSERT INTO t_timestamp_without_zone VALUES (1, '2022-01-01 12:34:56'), (2, '2022-02-01 23:45:01'), (3, 'Feb 10 5:32PM 1997'), (4, 'Feb 10 16:32:05 99');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_timestamp_without_zone ORDER BY id;",
				Expected: []sql.Row{
					{1, "2022-01-01 12:34:56"},
					{2, "2022-02-01 23:45:01"},
					{3, "1997-02-10 17:32:00"},
					{4, "1999-02-10 16:32:05"},
				},
			},
			{
				Query: "SELECT '2000-01-01'::timestamp;",
				Expected: []sql.Row{
					{"2000-01-01 00:00:00"},
				},
			},
			{
				Query: `SELECT '2000-01-01 00:00:00'::timestamp;`,
				Expected: []sql.Row{
					{"2000-01-01 00:00:00"},
				},
			},
			{
				Query:    `SELECT timestamp without time zone '2025-07-21 04:05:06' + interval '2 minutes';`,
				Expected: []sql.Row{{"2025-07-21 04:07:06"}},
			},
		},
	},
	{
		Name: "Timestamp with time zone type",
		SetUpScript: []string{
			"CREATE TABLE t_timestamp_with_zone (id INTEGER primary key, v1 TIMESTAMP WITH TIME ZONE);",
			"INSERT INTO t_timestamp_with_zone VALUES (1, '2022-01-01 12:34:56 UTC'), (2, '2022-02-01 23:45:01 America/New_York');",
		},
		Assertions: []ScriptTestAssertion{
			{
				// timezone representation is reported via local time, need to account for that in testing
				Query:    "SET timezone TO '-04:25'",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT * FROM t_timestamp_with_zone ORDER BY id;",
				Expected: []sql.Row{
					{1, "2022-01-01 16:59:56+04:25"},
					{2, "2022-02-02 09:10:01+04:25"},
				},
			},
			{
				Query: "SELECT '2000-01-01'::timestamptz;",
				Expected: []sql.Row{
					{"2000-01-01 00:00:00+04:25"},
				},
			},
			{
				Query: `SELECT '2000-01-01 00:00:00'::timestamptz;`,
				Expected: []sql.Row{
					{"2000-01-01 00:00:00+04:25"},
				},
			},
			{
				// timezone representation is reported via local time, need to account for that in testing
				Query:    "SET timezone TO '-06:00'",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT * FROM t_timestamp_with_zone ORDER BY id;",
				Expected: []sql.Row{
					{1, "2022-01-01 18:34:56+06"},
					{2, "2022-02-02 10:45:01+06"},
				},
			},
			{
				Query:    `SELECT timestamp with time zone '2025-07-21 04:05:06 UTC' + interval '2 minutes';`,
				Expected: []sql.Row{{"2025-07-21 10:07:06+06"}},
			},
			{
				Query:    "SET timezone TO default",
				Expected: []sql.Row{},
			},
		},
	},
	{
		Name: "Tsquery type",
		Skip: true,
		SetUpScript: []string{
			"CREATE TABLE t_tsquery (id INTEGER primary key, v1 TSQUERY);",
			"INSERT INTO t_tsquery VALUES (1, 'word'), (2, 'phrase & (another | term)');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_tsquery ORDER BY id;",
				Expected: []sql.Row{
					{1, "word"},
					{2, "phrase & (another | term)"},
				},
			},
		},
	},
	{
		Name: "Tsvector type",
		Skip: true,
		SetUpScript: []string{
			"CREATE TABLE t_tsvector (id INTEGER primary key, v1 TSVECTOR);",
			"INSERT INTO t_tsvector VALUES (1, 'simple'), (2, 'complex & (query | terms)');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_tsvector ORDER BY id;",
				// TODO: output differs from postgres, may need a custom type, not a string
				Expected: []sql.Row{
					{1, "simple"},
					{2, "complex & (query | terms)"},
				},
			},
		},
	},
	{
		// This syntax had a very bad error message ("unsupported: this syntax"), this test just assert it's better
		// It can be retired when we support the type.
		Name: "tsvector unsupported error",
		Assertions: []ScriptTestAssertion{
			{
				Query:       `CREATE TABLE t_tsvector (id INTEGER primary key, v1 TSVECTOR);`,
				ExpectedErr: `type "tsvector" does not exist`,
			},
		},
	},
	{
		Name: "Uuid type",
		SetUpScript: []string{
			"CREATE TABLE t_uuid (id INTEGER primary key, v1 UUID);",
			"INSERT INTO t_uuid VALUES (1, 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'), (2, 'f47ac10b58cc4372a567-0e02b2c3d479');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_uuid ORDER BY id;",
				Expected: []sql.Row{
					{1, "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"},
					{2, "f47ac10b-58cc-4372-a567-0e02b2c3d479"},
				},
			},
			{
				Query:    "select uuid 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11';",
				Expected: []sql.Row{{"a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"}},
			},
		},
	},
	{
		Name: "Uuid default value",
		SetUpScript: []string{
			"CREATE TABLE t_uuid (id INTEGER primary key, v1 UUID default 'a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'::uuid);",
			"INSERT INTO t_uuid VALUES (1, 'f47ac10b58cc4372a567-0e02b2c3d479');",
			"INSERT INTO t_uuid (id) VALUES (2);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_uuid ORDER BY id;",
				Expected: []sql.Row{
					{1, "f47ac10b-58cc-4372-a567-0e02b2c3d479"},
					{2, "a1eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"},
				},
			},
		},
	},
	{
		Name: "Uuid key",
		SetUpScript: []string{
			"CREATE TABLE t_uuid (id UUID primary key, v1 UUID);",
			"INSERT INTO t_uuid VALUES ('f47ac10b58cc4372a567-0e02b2c3d479', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'), ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'f47ac10b58cc4372a567-0e02b2c3d479');",
			"create table t_uuid2 (id int primary key, v1 uuid, v2 uuid);",
			"create index on t_uuid2(v1, v2);",
			"insert into t_uuid2 values " +
				"(1, 'f47ac10b58cc4372a567-0e02b2c3d479', 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'), " +
				"(2, 'dcf783c8-49c2-44b4-8b90-34ad8c52ea1e', 'f99802e8-0018-4913-806c-bcad5d246d46');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_uuid WHERE ID = 'f47ac10b58cc4372a567-0e02b2c3d479' ORDER BY id;",
				Expected: []sql.Row{
					{"f47ac10b-58cc-4372-a567-0e02b2c3d479", "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"},
				},
			},
			{
				Query: "SELECT * FROM t_uuid2 WHERE v1 = 'f47ac10b58cc4372a567-0e02b2c3d479' and v2 = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11' ORDER BY id;",
				Expected: []sql.Row{
					{1, "f47ac10b-58cc-4372-a567-0e02b2c3d479", "a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11"},
				},
			},
			{
				Query: "SELECT * FROM t_uuid2 WHERE v1 < 'f47ac10b58cc4372a567-0e02b2c3d479' ORDER BY id;",
				Expected: []sql.Row{
					{2, "dcf783c8-49c2-44b4-8b90-34ad8c52ea1e", "f99802e8-0018-4913-806c-bcad5d246d46"},
				},
			},
		},
	},
	{
		Name: "Uuid array type",
		SetUpScript: []string{
			"CREATE TABLE t_uuid (id INTEGER primary key, v1 UUID[]);",
			"INSERT INTO t_uuid VALUES (1, ARRAY['a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'::uuid, NULL]), (2, ARRAY[NULL, 'f47ac10b58cc4372a567-0e02b2c3d479'::uuid]);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_uuid ORDER BY id;",
				Expected: []sql.Row{
					{1, "{a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11,NULL}"},
					{2, "{NULL,f47ac10b-58cc-4372-a567-0e02b2c3d479}"},
				},
			},
		},
	},
	{
		Name: "Xid type",
		SetUpScript: []string{
			"CREATE TABLE t_xid (id INTEGER primary key, v1 XID, v2 VARCHAR(20));",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:       "INSERT INTO t_xid VALUES (1, 1234, '100');",
				ExpectedErr: "expression is of type",
			},
			{
				Query:       "INSERT INTO t_xid VALUES (1, 1234::xid, '100');",
				ExpectedErr: "does not exist",
			},
			{
				Query:    "INSERT INTO t_xid VALUES (1, NULL, '100');",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT * FROM t_xid ORDER BY id;",
				Expected: []sql.Row{
					{1, nil, "100"},
				},
			},
			{
				Query:    "INSERT INTO t_xid VALUES (2, '100', '101');",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT * FROM t_xid WHERE v1 IS NOT NULL;",
				Expected: []sql.Row{
					{2, 100, "101"},
				},
			},
			{
				Query:    "UPDATE t_xid SET v1='9012' WHERE id=1;",
				Expected: []sql.Row{},
			},
			{
				Query:    "DELETE FROM t_xid WHERE v1=100;",
				Skip:     true, // TODO: need to implement comparisons, cast interface isn't adequate enough
				Expected: []sql.Row{},
			},
			{
				Query:       "SELECT * FROM t_xid ORDER BY v1 DESC;",
				Skip:        true, // TODO: should error with "could not identify an ordering operator for type xid"
				ExpectedErr: "does not exist",
			},
			{
				Query:    "INSERT INTO t_xid VALUES (4, '4294967295', 'a');",
				Expected: []sql.Row{},
			},
			{
				Query:    "INSERT INTO t_xid VALUES (5, '4294967296', 'b');",
				Expected: []sql.Row{},
			},
			{
				Query:    "INSERT INTO t_xid VALUES (6, '0', 'c');",
				Expected: []sql.Row{},
			},
			{
				Query:    "INSERT INTO t_xid VALUES (7, '-1', 'd');",
				Expected: []sql.Row{},
			},
			{
				Query:    "INSERT INTO t_xid VALUES (8, 'abc', 'd');",
				Expected: []sql.Row{},
			},
			{
				Query: "SELECT * FROM t_xid ORDER BY id;",
				Expected: []sql.Row{
					{1, 9012, "100"},
					{2, 100, "101"},
					{4, 4294967295, "a"},
					{5, 0, "b"},
					{6, 0, "c"},
					{7, 4294967295, "d"},
					{8, 0, "d"},
				},
			},
		},
	},
	{
		Name: "Xid type, explicit casts",
		SetUpScript: []string{
			"CREATE TABLE t_xid (id INTEGER primary key, v1 XID);",
			"INSERT INTO t_xid VALUES (1, '1234'), (2, '4294967295');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_xid ORDER BY id;",
				Expected: []sql.Row{
					{1, 1234},
					{2, 4294967295},
				},
			},
			// Cast from XID to types
			{
				Query: "SELECT v1::char(1), v1::varchar(2), v1::text, v1::name FROM t_xid WHERE id=1;",
				Expected: []sql.Row{
					{"1", "12", "1234", "1234"},
				},
			},
			{
				Query:       "SELECT v1::smallint FROM t_xid WHERE id=1;",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT v1::integer FROM t_xid WHERE id=1;",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT v1::bigint FROM t_xid WHERE id=1;",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT v1::oid FROM t_xid WHERE id=1;",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT v1::float4 FROM t_xid WHERE id=1;",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT v1::float8 FROM t_xid WHERE id=1;",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT v1::numeric FROM t_xid WHERE id=1;",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT v1::boolean FROM t_xid WHERE id=1;",
				ExpectedErr: "does not exist",
			},
			// Cast to XID from types
			{
				Query: "SELECT ('123'::char(3))::xid, ('123'::varchar)::xid, ('0'::text)::xid, ('400'::name)::xid;",
				Expected: []sql.Row{
					{123, 123, 0, 400},
				},
			},
			{
				Query: "SELECT ('-1'::char(3))::xid, ('-1'::varchar)::xid, ('-1'::text)::xid, ('-1'::name)::xid;",
				Expected: []sql.Row{
					{4294967295, 4294967295, 4294967295, 4294967295},
				},
			},
			{
				Query: "SELECT ('-2147483648'::char(11))::xid, ('-2147483648'::varchar)::xid, ('-2147483648'::text)::xid, ('-2147483648'::name)::xid;",
				Expected: []sql.Row{
					{2147483648, 2147483648, 2147483648, 2147483648},
				},
			},
			{
				Query:       "SELECT (10::int2)::xid;",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT (10::boolean)::xid;",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT (10::int4)::xid;",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT (10::int8)::xid;",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT (1.1::float4)::xid;",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT (1.1::float8)::xid;",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT (1.1::decimal)::xid;",
				ExpectedErr: "does not exist",
			},
			{
				Query: "SELECT ('4294967295'::text)::xid, ('4294967297'::text)::xid;",
				Expected: []sql.Row{
					{4294967295, 1},
				},
			},
			{
				Query: "SELECT ('-4294967295'::text)::xid, ('-4294967297'::text)::xid;",
				Expected: []sql.Row{
					{1, 4294967295},
				},
			},
			{
				Query: "SELECT ('4294967295'::varchar)::xid, ('4294967296232'::varchar)::xid;",
				Expected: []sql.Row{
					{4294967295, 232},
				},
			},
			{
				Query: "SELECT ('-4294967295'::varchar)::xid, ('-4294967296232'::varchar)::xid;",
				Expected: []sql.Row{
					{1, 4294967064},
				},
			},
			{
				Query: "SELECT ('4294967295'::char(11))::xid, ('4294967296'::char(11))::xid;",
				Expected: []sql.Row{
					{4294967295, 0},
				},
			},
			{
				Query: "SELECT ('4294967295'::name)::xid, ('4294967296'::name)::xid;",
				Expected: []sql.Row{
					{4294967295, 0},
				},
			},
			{
				Query: "SELECT ('abc'::text)::xid, ('abc'::char(3))::xid, ('abc'::varchar)::xid, ('abc'::name)::xid;",
				Expected: []sql.Row{
					{0, 0, 0, 0},
				},
			},
		},
	},
	{
		Name: "Xid array type",
		SetUpScript: []string{
			"CREATE TABLE t_xid (id INTEGER primary key, v1 XID[], v2 CHARACTER(100), v3 BOOLEAN);",
			"INSERT INTO t_xid VALUES (2, '{123, 456, 789, 101}', '1234567890', true);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: `SELECT v1::varchar(1)[] FROM t_xid;`,
				Expected: []sql.Row{
					{"{1,4,7,1}"},
				},
			},
			{
				Query:       `INSERT INTO t_xid VALUES (2, ARRAY[123, 456, 789, 101], '1234567890', true);`,
				ExpectedErr: "is of type",
			},
		},
	},
	{
		Name: "Xml type",
		Skip: true,
		SetUpScript: []string{
			"CREATE TABLE t_xml (id INTEGER primary key, v1 XML);",
			"INSERT INTO t_xml VALUES (1, '<note><to>Tove</to><from>Jani</from><body>Don''t forget me this weekend!</body></note>'), (2, '<book><title>Introduction to Golang</title><author>John Doe</author></book>');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_xml ORDER BY id;",
				Expected: []sql.Row{
					{1, "<note><to>Tove</to><from>Jani</from><body>Don't forget me this weekend!</body></note>"},
					{2, "<book><title>Introduction to Golang</title><author>John Doe</author></book>"},
				},
			},
		},
	},
	{
		Name: "Polymorphic types",
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT array_append(ARRAY[1], 2);",
				Expected: []sql.Row{
					{"{1,2}"},
				},
			},
			{
				Query: "SELECT array_append(ARRAY['abc','def'], 'ghi');",
				Expected: []sql.Row{
					{"{abc,def,ghi}"},
				},
			},
			{
				Query: "SELECT array_append(ARRAY['abc','def'], null);",
				Expected: []sql.Row{
					{"{abc,def,NULL}"},
				},
			},
			{
				Query: "SELECT array_append(null, null);",
				Expected: []sql.Row{
					{"{NULL}"},
				},
			},
			{
				Query: "SELECT array_append(null, 'ghi');",
				Expected: []sql.Row{
					{"{ghi}"},
				},
			},
			{
				Query: "SELECT array_append(null, 3);",
				Expected: []sql.Row{
					{"{3}"},
				},
			},
			{
				Query:       "SELECT array_append(1, 2);",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT array_append(1, ARRAY[2]);",
				ExpectedErr: "does not exist",
			},
			{
				Query:       "SELECT array_append(ARRAY[1], ARRAY[2]);",
				ExpectedErr: "does not exist",
			},
		},
	},
}

func TestSameTypes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Integer types",
			SetUpScript: []string{
				"CREATE TABLE test1 (v1 SMALLINT, v2 INTEGER, v3 BIGINT);",
				"CREATE TABLE test2 (v1 INT2, v2 INT4, v3 INT8);",
				"INSERT INTO test1 VALUES (1, 2, 3), (4, 5, 6);",
				"INSERT INTO test2 VALUES (1, 2, 3), (4, 5, 6);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test1 ORDER BY 1;",
					Expected: []sql.Row{
						{1, 2, 3},
						{4, 5, 6},
					},
				},
				{
					Query: "SELECT * FROM test1 ORDER BY v1;",
					Expected: []sql.Row{
						{1, 2, 3},
						{4, 5, 6},
					},
				},
				{
					Query: "SELECT * FROM test2 ORDER BY 1;",
					Expected: []sql.Row{
						{1, 2, 3},
						{4, 5, 6},
					},
				},
				{
					Query: "SELECT * FROM test2 ORDER BY v1;",
					Expected: []sql.Row{
						{1, 2, 3},
						{4, 5, 6},
					},
				},
				{
					Query:    "select int2 '2', int4 '3', int8 '4'",
					Expected: []sql.Row{{2, 3, 4}},
				},
			},
		},
		{
			Name: "Arbitrary precision types",
			SetUpScript: []string{
				"CREATE TABLE test (v1 DECIMAL(10, 1), v2 NUMERIC(11, 2));",
				"INSERT INTO test VALUES (14854.5, 2504.25), (566821525.5, 735134574.75), (21525, 134574.7);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test ORDER BY 1;",
					Expected: []sql.Row{
						{Numeric("14854.5"), Numeric("2504.25")},
						{Numeric("21525.0"), Numeric("134574.70")},
						{Numeric("566821525.5"), Numeric("735134574.75")},
					},
				},
			},
		},
		{
			Name: "Floating point types",
			SetUpScript: []string{
				"CREATE TABLE test1 (v1 REAL, v2 DOUBLE PRECISION);",
				"CREATE TABLE test2 (v1 FLOAT4, v2 FLOAT8);",
				"INSERT INTO test1 VALUES (10.125, 20.4), (40.875, 81.6);",
				"INSERT INTO test2 VALUES (10.125, 20.4), (40.875, 81.6);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test1 ORDER BY 1;",
					Expected: []sql.Row{
						{10.125, 20.4},
						{40.875, 81.6},
					},
				},
				{
					Query: "SELECT * FROM test2 ORDER BY 1;",
					Expected: []sql.Row{
						{10.125, 20.4},
						{40.875, 81.6},
					},
				},
			},
		},
		{
			// TIME has the same name, but operates a bit differently, so it's not included as a "same type"
			Name: "Date and time types",
			SetUpScript: []string{
				"CREATE TABLE test (v1 TIMESTAMP, v2 DATE);",
				"INSERT INTO test VALUES ('1986-08-02 17:04:22', '2023-09-03');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test ORDER BY 1;",
					Expected: []sql.Row{
						{"1986-08-02 17:04:22", "2023-09-03"},
					},
				},
			},
		},
		{
			// ENUM exists, but features too many differences to incorporate as a "same type"
			// BLOB exists, but functions as a BYTEA, which operates differently than a BINARY/VARBINARY in MySQL
			Name: "Text types",
			SetUpScript: []string{
				"CREATE TABLE test (v1 CHARACTER VARYING(255), v2 CHARACTER(3), v3 TEXT);",
				"INSERT INTO test VALUES ('abc', 'def', 'ghi'), ('jkl', 'mno', 'pqr');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test ORDER BY 1;",
					Expected: []sql.Row{
						{"abc", "def", "ghi"},
						{"jkl", "mno", "pqr"},
					},
				},
			},
		},
	})
}

func TestEnumTypes(t *testing.T) {
	RunScripts(t, enumTypeTests)
}

var enumTypeTests = []ScriptTest{
	{
		Name: "create enum type",
		SetUpScript: []string{
			`CREATE TYPE mood AS ENUM ('sad', 'ok', 'happy')`,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    `CREATE TABLE person (name text, current_mood mood);`,
				Expected: []sql.Row{},
			},
			{
				Query:    `INSERT INTO person VALUES ('Moe', 'happy'), ('Larry', 'sad'), ('Curly', 'ok');`,
				Expected: []sql.Row{},
			},
			{
				Query:    `SELECT * FROM person order by current_mood;`,
				Expected: []sql.Row{{"Larry", "sad"}, {"Curly", "ok"}, {"Moe", "happy"}},
			},
			{
				Query:    `SELECT * FROM person order by name;`,
				Expected: []sql.Row{{"Curly", "ok"}, {"Larry", "sad"}, {"Moe", "happy"}},
			},
			{
				Query:    `SELECT * FROM person;`,
				Expected: []sql.Row{{"Moe", "happy"}, {"Larry", "sad"}, {"Curly", "ok"}},
			},
			{
				Query:    `SELECT * FROM person WHERE current_mood = 'happy';`,
				Expected: []sql.Row{{"Moe", "happy"}},
			},
			{
				Query:    `SELECT * FROM person WHERE current_mood > 'sad';`,
				Expected: []sql.Row{{"Moe", "happy"}, {"Curly", "ok"}},
			},
			{
				Query:    `SELECT * FROM person WHERE current_mood > 'sad' ORDER BY current_mood;`,
				Expected: []sql.Row{{"Curly", "ok"}, {"Moe", "happy"}},
			},
			{
				Query:       `INSERT INTO person VALUES ('Joey', 'invalid');`,
				ExpectedErr: `invalid input value for enum mood: "invalid"`,
			},
			{
				Query:       `CREATE TYPE failure AS ENUM ('ok','ok');`,
				ExpectedErr: `duplicate key value violates unique constraint "pg_enum_typid_label_index"`,
			},
			{
				Query:    `CREATE TYPE empty_mood AS ENUM ();`,
				Expected: []sql.Row{},
			},
		},
	},
	{
		Name: "drop enum type",
		SetUpScript: []string{
			`CREATE TYPE mood AS ENUM ('sad', 'ok', 'happy')`,
			`CREATE TYPE empty_enum AS ENUM ()`,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:       `DROP TYPE mood, empty_enum;`,
				ExpectedErr: `dropping multiple types in DROP TYPE is not yet supported`,
			},
			{
				Query:    `DROP TYPE empty_enum;`,
				Expected: []sql.Row{},
			},
			{
				Query:       `DROP TYPE empty_enum;`,
				ExpectedErr: `type "empty_enum" does not exist`,
			},
			{
				Query:    `DROP TYPE IF EXISTS empty_enum;`,
				Expected: []sql.Row{},
			},
			{
				Query:       `DROP TYPE _mood;`,
				ExpectedErr: `cannot drop type mood[] because type mood requires it`,
			},
		},
	},
	{
		Name: "enum type cast",
		SetUpScript: []string{
			`CREATE TYPE mood AS ENUM ('sad', 'ok', 'happy')`,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    `select 'sad'::mood`,
				Expected: []sql.Row{{"sad"}},
			},
			{
				Query:       `select 'invalid'::mood`,
				ExpectedErr: `invalid input value for enum mood: "invalid"`,
			},
		},
	},
	{
		Skip: true,
		Name: "enum type function",
		SetUpScript: []string{
			`CREATE TYPE mood AS ENUM ('sad', 'ok', 'happy')`,
		},
		Assertions: []ScriptTestAssertion{
			{
				// oid of type 'mood' = 16675
				Query:    `select enum_in('sad'::cstring, 16675);`,
				Expected: []sql.Row{{"sad"}},
			},
		},
	},
	{
		Skip: true,
		Name: "create type with existing array type name updates the name of the array type",
		Assertions: []ScriptTestAssertion{
			{
				Query:    `CREATE TYPE my_type AS ENUM ();`,
				Expected: []sql.Row{},
			},
			{
				Query:    `CREATE TYPE _my_type;`,
				Expected: []sql.Row{},
			},
			{
				Query:    `SELECT typname from pg_type where typname like '%my_type'`,
				Expected: []sql.Row{{"my_type"}, {"_my_type"}, {"__my_type"}},
			},
			{
				Query:    `DROP TYPE my_type;`,
				Expected: []sql.Row{},
			},
			{
				Query:    `DROP TYPE _my_type;`,
				Expected: []sql.Row{},
			},
		},
	},
}

func TestShellTypes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "shell type use cases",
			Assertions: []ScriptTestAssertion{
				{
					Query:    `CREATE TYPE undefined_type;`,
					Expected: []sql.Row{},
				},
				{
					Query:       `select 1::undefined_type;`,
					ExpectedErr: `type "undefined_type" is only a shell`,
				},
				{
					Query:    `DROP TYPE undefined_type;`,
					Expected: []sql.Row{},
				},
				{
					Query:    `DROP TYPE IF EXISTS undefined_type;`,
					Expected: []sql.Row{},
				},
			},
		},
	})
}
