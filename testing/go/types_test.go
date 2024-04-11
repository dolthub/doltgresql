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
		Skip: true,
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
		Skip: true,
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
			"INSERT INTO t_bytea VALUES (1, E'\\\\xDEADBEEF'), (2, '\\xC0FFEE');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_bytea ORDER BY id;",
				Expected: []sql.Row{
					{1, []byte{0xDE, 0xAD, 0xBE, 0xEF}},
					{2, []byte{0xC0, 0xFF, 0xEE}},
				},
			},
		},
	},
	{
		Name: "Character type",
		SetUpScript: []string{
			"CREATE TABLE t_character (id INTEGER primary key, v1 CHARACTER(5));",
			"INSERT INTO t_character VALUES (1, 'abcde'), (2, 'vwxyz'), (3, 'ghi');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_character ORDER BY id;",
				Expected: []sql.Row{
					{1, "abcde"},
					{2, "vwxyz"},
					{3, "ghi  "},
				},
			},
		},
	},
	{
		Name: "Character varying type",
		SetUpScript: []string{
			"CREATE TABLE t_varchar (id INTEGER primary key, v1 CHARACTER VARYING(10));",
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
		Name: "Character varying array type, with length",
		SetUpScript: []string{
			"CREATE TABLE t_varchar1 (v1 CHARACTER VARYING[]);",
			"CREATE TABLE t_varchar2 (v1 CHARACTER VARYING(1)[]);",
			"INSERT INTO t_varchar1 VALUES (ARRAY['ab''cdef', 'what', 'is,hi', 'wh\"at']);",
			"INSERT INTO t_varchar2 VALUES (ARRAY['ab''cdef', 'what', 'is,hi', 'wh\"at']);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: `SELECT v1::varchar(1)[] FROM t_varchar1;`,
				Expected: []sql.Row{
					{"{a,w,i,w}"},
				},
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
		Name: "Character varying array type, no length",
		SetUpScript: []string{
			"CREATE TABLE t_varchar (id INTEGER primary key, v1 CHARACTER VARYING[]);",
			"INSERT INTO t_varchar VALUES (1, ARRAY['abcdefghij', NULL]), (2, ARRAY['ab''cdef', 'what', 'is,hi', 'wh\"at', '}', '{', '{}']);",
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
		Skip: true,
		SetUpScript: []string{
			"CREATE TABLE t_interval (id INTEGER primary key, v1 INTERVAL);",
			"INSERT INTO t_interval VALUES (1, '1 day 3 hours'), (2, '2 hours 30 minutes');",
		},
		Assertions: []ScriptTestAssertion{
			{
				// TODO: might need a GMS type here, not a string (psql output is different than below)
				Query: "SELECT * FROM t_interval ORDER BY id;",
				Expected: []sql.Row{
					{1, "1 day 3 hours"},
					{2, "2 hours 30 minutes"},
				},
			},
		},
	},
	{
		Name: "JSON type",
		SetUpScript: []string{
			"CREATE TABLE t_json (id INTEGER primary key, v1 JSON);",
			"INSERT INTO t_json VALUES (1, '{\"key\": \"value\"}'), (2, '{\"num\": 42}');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_json ORDER BY id;",
				Expected: []sql.Row{
					{1, `{"key":"value"}`},
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
					{1, `{"key":"value"}`},
					{2, `{"num":42}`},
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
		Name: "Numeric type",
		SetUpScript: []string{
			"CREATE TABLE t_numeric (id INTEGER primary key, v1 NUMERIC(5,2));",
			"INSERT INTO t_numeric VALUES (1, 123.45), (2, 67.89);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_numeric ORDER BY id;",
				Expected: []sql.Row{
					{1, 123.45},
					{2, 67.89},
				},
			},
		},
	},
	{
		Name: "Numeric type, no scale or precision",
		SetUpScript: []string{
			"CREATE TABLE t_numeric (id INTEGER primary key, v1 NUMERIC);",
			"INSERT INTO t_numeric VALUES (1, 123.45), (2, 67.875);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_numeric ORDER BY id;",
				Expected: []sql.Row{
					{1, 123.45},
					{2, 67.875},
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
		Skip: true,
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
		Name: "Serial type",
		Skip: true,
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
		},
	},
	{
		Name: "Text type",
		SetUpScript: []string{
			"CREATE TABLE t_text (id INTEGER primary key, v1 TEXT);",
			"INSERT INTO t_text VALUES (1, 'Hello'), (2, 'World');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_text ORDER BY id;",
				Expected: []sql.Row{
					{1, "Hello"},
					{2, "World"},
				},
			},
		},
	},
	{
		Name: "Time without time zone type",
		SetUpScript: []string{
			"CREATE TABLE t_time_without_zone (id INTEGER primary key, v1 TIME);",
			"INSERT INTO t_time_without_zone VALUES (1, '12:34:56'), (2, '23:45:01');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_time_without_zone ORDER BY id;",
				Expected: []sql.Row{
					{1, "12:34:56"},
					{2, "23:45:01"},
				},
			},
		},
	},
	{ // TODO: timezone representation is reported via local time, need to account for that in testing?
		Name: "Time with time zone type",
		Skip: true,
		SetUpScript: []string{
			"CREATE TABLE t_time_with_zone (id INTEGER primary key, v1 TIME WITH TIME ZONE);",
			"INSERT INTO t_time_with_zone VALUES (1, '12:34:56 UTC'), (2, '23:45:01 America/New_York');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_time_with_zone ORDER BY id;",
				Expected: []sql.Row{
					{1, "12:34:56 UTC"},
					{2, "23:45:01 America/New_York"},
				},
			},
		},
	},
	{
		Name: "Timestamp without time zone type",
		SetUpScript: []string{
			"CREATE TABLE t_timestamp_without_zone (id INTEGER primary key, v1 TIMESTAMP);",
			"INSERT INTO t_timestamp_without_zone VALUES (1, '2022-01-01 12:34:56'), (2, '2022-02-01 23:45:01');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_timestamp_without_zone ORDER BY id;",
				Expected: []sql.Row{
					{1, "2022-01-01 12:34:56"},
					{2, "2022-02-01 23:45:01"},
				},
			},
		},
	},
	{ // TODO: timezone representation is reported via local time, need to account for that in testing?
		Name: "Timestamp with time zone type",
		Skip: true,
		SetUpScript: []string{
			"CREATE TABLE t_timestamp_with_zone (id INTEGER primary key, v1 TIMESTAMP WITH TIME ZONE);",
			"INSERT INTO t_timestamp_with_zone VALUES (1, '2022-01-01 12:34:56 UTC'), (2, '2022-02-01 23:45:01 America/New_York');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "SELECT * FROM t_timestamp_with_zone ORDER BY id;",
				Expected: []sql.Row{
					{1, "2022-01-01 12:34:56 UTC"},
					{2, "2022-02-01 23:45:01 America/New_York"},
				},
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
}

func TestTypes(t *testing.T) {
	RunScripts(t, typesTests)
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
			},
		},
		{
			Name: "Arbitrary precision types",
			SetUpScript: []string{
				"CREATE TABLE test (v1 DECIMAL(10, 1), v2 NUMERIC(11, 2));",
				"INSERT INTO test VALUES (14854.5, 2504.25), (566821525.5, 735134574.75);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test ORDER BY 1;",
					Expected: []sql.Row{
						{14854.5, 2504.25},
						{566821525.5, 735134574.75},
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
					Skip: true, // type length info is not being passed correctly to the engine, which causes the
					// select to fail with 'invalid length for "char": 3'
				},
			},
		},
		{
			Name: "JSON type",
			SetUpScript: []string{
				"CREATE TABLE test (v1 INT, v2 JSON);",
				`INSERT INTO test VALUES (1, '{"key1": {"key": "value"}}'), (2, '{"key1": "value1", "key2": "value2"}'), (3, '{"key1": {"key": [2,3]}}');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT * FROM test ORDER BY 1;",
					Expected: []sql.Row{
						{1, `{"key1":{"key":"value"}}`},
						{2, `{"key1":"value1","key2":"value2"}`},
						{3, `{"key1":{"key":[2,3]}}`},
					},
				},
				{
					Query: "SELECT * FROM test ORDER BY v1;",
					Expected: []sql.Row{
						{1, `{"key1":{"key":"value"}}`},
						{2, `{"key1":"value1","key2":"value2"}`},
						{3, `{"key1":{"key":[2,3]}}`},
					},
				},
			},
		},
	})
}
