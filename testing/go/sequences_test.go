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

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestSequences(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Basic CREATE SEQUENCE and DROP SEQUENCE",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE SEQUENCE test;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT nextval('test');",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT nextval('test');",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "DROP SEQUENCE test;",
					Expected: []sql.Row{},
				},
				{
					Query:       "SELECT nextval('test');",
					ExpectedErr: true,
				},
			},
		},
		{
			Name: "CREATE SEQUENCE IF NOT EXISTS",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE SEQUENCE test1;",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE SEQUENCE test1;",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT nextval('test1');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "CREATE SEQUENCE IF NOT EXISTS test1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test1');",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "CREATE SEQUENCE IF NOT EXISTS test2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "CREATE SEQUENCE IF NOT EXISTS test2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{2}},
				},
			},
		},
		{
			Name: "DROP SEQUENCE IF NOT EXISTS",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE SEQUENCE test1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE SEQUENCE test2;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test1');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "DROP SEQUENCE test1;",
					Expected: []sql.Row{},
				},
				{
					Query:       "DROP SEQUENCE test1;",
					ExpectedErr: true,
				},
				{
					Query:    "DROP SEQUENCE IF EXISTS test1;",
					Expected: []sql.Row{},
				},
				{
					Query:       "SELECT nextval('test1');",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "DROP SEQUENCE IF EXISTS test2;",
					Expected: []sql.Row{},
				},
				{
					Query:       "SELECT nextval('test2');",
					ExpectedErr: true,
				},
				{
					Query:    "DROP SEQUENCE IF EXISTS test2;",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name: "MINVALUE and MAXVALUE with DATA TYPE",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE SEQUENCE test1 AS SMALLINT MINVALUE -32768;",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE SEQUENCE test2 AS SMALLINT MINVALUE -32769;",
					ExpectedErr: true,
				},
				{
					Query:    "CREATE SEQUENCE test3 AS SMALLINT MAXVALUE 32767;",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE SEQUENCE test4 AS SMALLINT MINVALUE 32768;",
					ExpectedErr: true,
				},
				{
					Query:    "CREATE SEQUENCE test5 AS INTEGER MINVALUE -2147483648;",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE SEQUENCE test6 AS INTEGER MINVALUE -2147483649;",
					ExpectedErr: true,
				},
				{
					Query:    "CREATE SEQUENCE test7 AS INTEGER MAXVALUE 2147483647;",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE SEQUENCE test8 AS INTEGER MINVALUE 2147483648;",
					ExpectedErr: true,
				},
				{
					Query:    "CREATE SEQUENCE test9 AS BIGINT MINVALUE -9223372036854775808;",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE SEQUENCE test10 AS BIGINT MINVALUE -9223372036854775809;",
					ExpectedErr: true,
				},
				{
					Query:    "CREATE SEQUENCE test11 AS BIGINT MAXVALUE 9223372036854775807;",
					Expected: []sql.Row{},
				},
				{
					Query:       "CREATE SEQUENCE test12 AS BIGINT MINVALUE 9223372036854775808;",
					ExpectedErr: true,
				},
			},
		},
		{
			Name: "CREATE SEQUENCE START",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE SEQUENCE test1 START 39;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test1');",
					Expected: []sql.Row{{39}},
				},
				{
					Query:       "CREATE SEQUENCE test2 START 0;",
					ExpectedErr: true,
				},
				{
					Query:    "CREATE SEQUENCE test2 MINVALUE 0 START 0;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "CREATE SEQUENCE test3 MINVALUE -100 START -7;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test3');",
					Expected: []sql.Row{{-7}},
				},
				{
					Query:       "CREATE SEQUENCE test4 START -5 INCREMENT 1;",
					ExpectedErr: true,
				},
				{
					Query:    "CREATE SEQUENCE test4 START -5 INCREMENT -1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test4');",
					Expected: []sql.Row{{-5}},
				},
				{
					Query:       "CREATE SEQUENCE test5 START 25 INCREMENT -1;",
					ExpectedErr: true,
				},
				{
					Query:    "CREATE SEQUENCE test5 START 25 MAXVALUE 25 INCREMENT -1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test5');",
					Expected: []sql.Row{{25}},
				},
				{
					Query:    "SELECT nextval('test5');",
					Expected: []sql.Row{{24}},
				},
			},
		},
		{
			Name: "CYCLE and NO CYCLE",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE SEQUENCE test1 MINVALUE 0 MAXVALUE 3 START 2 INCREMENT 1 NO CYCLE;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test1');",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT nextval('test1');",
					Expected: []sql.Row{{3}},
				},
				{
					Query:       "SELECT nextval('test1');",
					ExpectedErr: true,
				},
				{
					Query:    "CREATE SEQUENCE test2 MINVALUE 0 MAXVALUE 3 START 2 INCREMENT 1 CYCLE;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "CREATE SEQUENCE test3 MINVALUE 0 MAXVALUE 3 START 1 INCREMENT -1 NO CYCLE;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test3');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT nextval('test3');",
					Expected: []sql.Row{{0}},
				},
				{
					Query:       "SELECT nextval('test3');",
					ExpectedErr: true,
				},
				{
					Query:    "CREATE SEQUENCE test4 MINVALUE 0 MAXVALUE 3 START 1 INCREMENT -1 CYCLE;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test4');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT nextval('test4');",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "SELECT nextval('test4');",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "CREATE SEQUENCE test5 MINVALUE 1 MAXVALUE 7 START 1 INCREMENT 5 CYCLE;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test5');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT nextval('test5');",
					Expected: []sql.Row{{6}},
				},
				{
					Query:    "SELECT nextval('test5');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "CREATE SEQUENCE test6 MINVALUE 1 MAXVALUE 7 START 6 INCREMENT -5 CYCLE;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT nextval('test6');",
					Expected: []sql.Row{{6}},
				},
				{
					Query:    "SELECT nextval('test6');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT nextval('test6');",
					Expected: []sql.Row{{7}},
				},
				{
					Query:    "SELECT nextval('test6');",
					Expected: []sql.Row{{2}},
				},
			},
		},
		{
			Name: "setval()",
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CREATE SEQUENCE test1 MINVALUE 1 MAXVALUE 10 START 5 INCREMENT 1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE SEQUENCE test2 MINVALUE 1 MAXVALUE 10 START 5 INCREMENT -1;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT setval('test1', 2);",
					Expected: []sql.Row{{2}},
				},
				{
					Query:    "SELECT nextval('test1');",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT setval('test1', 10);",
					Expected: []sql.Row{{10}},
				},
				{
					Query:       "SELECT nextval('test1');",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT setval('test1', 10, false);",
					Expected: []sql.Row{{10}},
				},
				{
					Query:    "SELECT nextval('test1');",
					Expected: []sql.Row{{10}},
				},
				{
					Query:    "SELECT setval('test1', 10, true);",
					Expected: []sql.Row{{10}},
				},
				{
					Query:       "SELECT nextval('test1');",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT setval('test2', 9);",
					Expected: []sql.Row{{9}},
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{8}},
				},
				{
					Query:    "SELECT setval('test2', 1);",
					Expected: []sql.Row{{1}},
				},
				{
					Query:       "SELECT nextval('test2');",
					ExpectedErr: true,
				},
				{
					Query:    "SELECT setval('test2', 1, false);",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT nextval('test2');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT setval('test2', 1, true);",
					Expected: []sql.Row{{1}},
				},
				{
					Query:       "SELECT nextval('test2');",
					ExpectedErr: true,
				},
				{
					Query:    "CREATE SEQUENCE test3 MINVALUE 3 MAXVALUE 7 START 5 INCREMENT 1 CYCLE;",
					Expected: []sql.Row{},
				},
				{
					Query:    "CREATE SEQUENCE test4 MINVALUE 3 MAXVALUE 7 START 5 INCREMENT -1 CYCLE;",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT setval('test3', 7, true);",
					Expected: []sql.Row{{7}},
				},
				{
					Query:    "SELECT nextval('test3');",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT setval('test4', 3, true);",
					Expected: []sql.Row{{3}},
				},
				{
					Query:    "SELECT nextval('test4');",
					Expected: []sql.Row{{7}},
				},
			},
		},
	})
}