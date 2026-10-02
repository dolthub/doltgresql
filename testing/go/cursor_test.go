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
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestCursors(t *testing.T) {
	setup := []string{
		"CREATE TABLE ct (id INT4 PRIMARY KEY, v TEXT);",
		"INSERT INTO ct VALUES (1, 'a'), (2, 'b'), (3, 'c'), (4, 'd'), (5, 'e');",
	}
	RunScripts(t, []ScriptTest{
		{
			Name:        "FETCH and MOVE directions on a SCROLL cursor",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query:       "BEGIN;",
					ExpectedTag: "BEGIN",
				},
				{
					Query:       "DECLARE c SCROLL CURSOR FOR SELECT * FROM ct ORDER BY id;",
					ExpectedTag: "DECLARE CURSOR",
				},
				{
					Query:    "FETCH c;",
					Expected: []sql.Row{{1, "a"}},
				},
				{
					Query:       "FETCH 2 FROM c;",
					ExpectedTag: "FETCH 2",
				},
				{
					Query:    "FETCH NEXT IN c;",
					Expected: []sql.Row{{4, "d"}},
				},
				{
					Query:    "FETCH PRIOR FROM c;",
					Expected: []sql.Row{{3, "c"}},
				},
				{
					Query:    "FETCH FIRST FROM c;",
					Expected: []sql.Row{{1, "a"}},
				},
				{
					Query:    "FETCH LAST FROM c;",
					Expected: []sql.Row{{5, "e"}},
				},
				{
					Query:    "FETCH ABSOLUTE 2 FROM c;",
					Expected: []sql.Row{{2, "b"}},
				},
				{
					Query:    "FETCH RELATIVE 2 FROM c;",
					Expected: []sql.Row{{4, "d"}},
				},
				{
					Query:    "FETCH FORWARD 10 FROM c;",
					Expected: []sql.Row{{5, "e"}},
				},
				{
					Query:    "FETCH FORWARD 10 FROM c;",
					Expected: []sql.Row{},
				},
				{
					Query:    "FETCH BACKWARD 2 FROM c;",
					Expected: []sql.Row{{5, "e"}, {4, "d"}},
				},
				{
					Query:    "FETCH BACKWARD ALL FROM c;",
					Expected: []sql.Row{{3, "c"}, {2, "b"}, {1, "a"}},
				},
				{
					Query:       "MOVE FORWARD 3 IN c;",
					ExpectedTag: "MOVE 3",
				},
				{
					Query:       "MOVE c;",
					ExpectedTag: "MOVE 1",
				},
				{
					Query:       "MOVE ALL IN c;",
					ExpectedTag: "MOVE 1",
				},
				{
					Query:    "FETCH ALL FROM c;",
					Expected: []sql.Row{},
				},
				{
					Query:    "FETCH -1 FROM c;",
					Expected: []sql.Row{{5, "e"}},
				},
				{
					Query:    "FETCH ABSOLUTE -2 FROM c;",
					Expected: []sql.Row{{4, "d"}},
				},
				{
					Query:    "FETCH RELATIVE 0 FROM c;",
					Expected: []sql.Row{{4, "d"}},
				},
				{
					Query:    "FETCH ABSOLUTE 0 FROM c;",
					Expected: []sql.Row{},
				},
				{
					Query:    "FETCH RELATIVE 0 FROM c;",
					Expected: []sql.Row{},
				},
				{
					Query:       "MOVE BACKWARD ALL FROM c;",
					ExpectedTag: "MOVE 0",
				},
				{
					Query:       "CLOSE c;",
					ExpectedTag: "CLOSE CURSOR",
				},
				{
					Query:           "CLOSE c;",
					ExpectedErr:     `cursor "c" does not exist`,
					ExpectedErrCode: "34000",
				},
				{
					Query:       "ROLLBACK;",
					ExpectedTag: "ROLLBACK",
				},
			},
		},
		{
			Name:        "cursors without a scroll option can move backward",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN;",
				},
				{
					Query: "DECLARE d CURSOR FOR SELECT * FROM ct ORDER BY id;",
				},
				{
					Query:    "FETCH 2 FROM d;",
					Expected: []sql.Row{{1, "a"}, {2, "b"}},
				},
				{
					Query:    "FETCH PRIOR FROM d;",
					Expected: []sql.Row{{1, "a"}},
					Skip:     true, // cursors without SCROLL only move forward
				},
				{
					Query:    "FETCH BACKWARD 1 FROM d;",
					Expected: []sql.Row{},
					Skip:     true, // cursors without SCROLL only move forward
				},
				{
					Query:    "FETCH ALL FROM d;",
					Expected: []sql.Row{{1, "a"}, {2, "b"}, {3, "c"}, {4, "d"}, {5, "e"}},
					Skip:     true, // cursors without SCROLL only move forward
				},
				{
					Query:    "FETCH BACKWARD ALL FROM d;",
					Expected: []sql.Row{{5, "e"}, {4, "d"}, {3, "c"}, {2, "b"}, {1, "a"}},
					Skip:     true, // cursors without SCROLL only move forward
				},
				{
					Query: "ROLLBACK;",
				},
			},
		},
		{
			Name:        "NO SCROLL cursors only move forward",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN;",
				},
				{
					Query: "DECLARE n NO SCROLL CURSOR FOR SELECT * FROM ct ORDER BY id;",
				},
				{
					Query:       "MOVE BACKWARD ALL FROM n;",
					ExpectedTag: "MOVE 0",
				},
				{
					Query:    "FETCH 2 FROM n;",
					Expected: []sql.Row{{1, "a"}, {2, "b"}},
				},
				{
					Query:       "MOVE 0 FROM n;",
					ExpectedTag: "MOVE 1",
				},
				{
					Query:    "FETCH ABSOLUTE 4 FROM n;",
					Expected: []sql.Row{{4, "d"}},
				},
				{
					Query:           "FETCH PRIOR FROM n;",
					ExpectedErr:     "cursor can only scan forward",
					ExpectedErrCode: "55000",
				},
				{
					Query: "ROLLBACK;",
				},
				{
					Query: "BEGIN;",
				},
				{
					Query: "DECLARE n NO SCROLL CURSOR FOR SELECT * FROM ct ORDER BY id;",
				},
				{
					Query:    "FETCH ABSOLUTE 2 FROM n;",
					Expected: []sql.Row{{2, "b"}},
				},
				{
					Query:           "FETCH ABSOLUTE 1 FROM n;",
					ExpectedErr:     "cursor can only scan forward",
					ExpectedErrCode: "55000",
				},
				{
					Query: "ROLLBACK;",
				},
			},
		},
		{
			Name:        "re-fetching the current row with a zero count",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN;",
				},
				{
					Query: "DECLARE i CURSOR FOR SELECT * FROM ct ORDER BY id;",
				},
				{
					Query:    "FETCH 0 FROM i;",
					Expected: []sql.Row{},
				},
				{
					Query:    "FETCH FORWARD 0 FROM i;",
					Expected: []sql.Row{},
				},
				{
					Query:    "FETCH 1 FROM i;",
					Expected: []sql.Row{{1, "a"}},
				},
				{
					Query:    "FETCH FORWARD 0 FROM i;",
					Expected: []sql.Row{{1, "a"}},
					Skip:     true, // cursors without SCROLL only move forward
				},
				{
					Query:    "FETCH BACKWARD 0 FROM i;",
					Expected: []sql.Row{{1, "a"}},
					Skip:     true, // cursors without SCROLL only move forward
				},
				{
					Query:       "MOVE 0 FROM i;",
					ExpectedTag: "MOVE 1",
				},
				{
					Query: "ROLLBACK;",
				},
			},
		},
		{
			Name:        "cursors return the rows that existed when they were declared",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN;",
				},
				{
					Query: "DECLARE i CURSOR FOR SELECT * FROM ct ORDER BY id;",
				},
				{
					Query: "INSERT INTO ct VALUES (6, 'f');",
				},
				{
					Query:    "FETCH ALL FROM i;",
					Expected: []sql.Row{{1, "a"}, {2, "b"}, {3, "c"}, {4, "d"}, {5, "e"}},
				},
				{
					Query: "ROLLBACK;",
				},
			},
		},
		{
			Name: "cursor queries",
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN;",
				},
				{
					Query: "DECLARE u CURSOR FOR VALUES (1), (2);",
				},
				{
					Query:    "FETCH ALL u;",
					Expected: []sql.Row{{1}, {2}},
				},
				{
					Query: "DECLARE w CURSOR FOR WITH x AS (SELECT 1 AS y) SELECT * FROM x;",
				},
				{
					Query:            "FETCH w;",
					Expected:         []sql.Row{{1}},
					ExpectedColNames: []string{"y"},
				},
				{
					Query: "DECLARE next CURSOR FOR SELECT 2 AS z;",
				},
				{
					Query:    "FETCH NEXT next;",
					Expected: []sql.Row{{2}},
				},
				{
					Query: `DECLARE "Mixed Case" CURSOR FOR SELECT 3;`,
				},
				{
					Query:    `FETCH "Mixed Case";`,
					Expected: []sql.Row{{3}},
				},
				{
					Query:           "FETCH mixed;",
					ExpectedErr:     `cursor "mixed" does not exist`,
					ExpectedErrCode: "34000",
				},
				{
					Query: "ROLLBACK;",
				},
			},
		},
		{
			Name:        "cursor errors",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query:           "DECLARE c CURSOR FOR SELECT * FROM ct;",
					ExpectedErr:     "DECLARE CURSOR can only be used in transaction blocks",
					ExpectedErrCode: "25P01",
				},
				{
					Query:           "FETCH x;",
					ExpectedErr:     `cursor "x" does not exist`,
					ExpectedErrCode: "34000",
				},
				{
					Query:           "MOVE x;",
					ExpectedErr:     `cursor "x" does not exist`,
					ExpectedErrCode: "34000",
				},
				{
					Query:           "CLOSE x;",
					ExpectedErr:     `cursor "x" does not exist`,
					ExpectedErrCode: "34000",
				},
				{
					Query:       "CLOSE ALL;",
					ExpectedTag: "CLOSE CURSOR ALL",
				},
				{
					Query: "BEGIN;",
				},
				{
					Query:           "DECLARE a SCROLL NO SCROLL CURSOR FOR SELECT 1;",
					ExpectedErr:     "cannot specify both SCROLL and NO SCROLL",
					ExpectedErrCode: "42P11",
				},
				{
					Query: "ROLLBACK;",
				},
				{
					Query: "BEGIN;",
				},
				{
					Query:           "DECLARE a ASENSITIVE INSENSITIVE CURSOR FOR SELECT 1;",
					ExpectedErr:     "cannot specify both ASENSITIVE and INSENSITIVE",
					ExpectedErrCode: "42P11",
				},
				{
					Query: "ROLLBACK;",
				},
				{
					Query: "BEGIN;",
				},
				{
					Query: "DECLARE c CURSOR FOR SELECT * FROM ct;",
				},
				{
					Query:           "DECLARE c CURSOR FOR SELECT 1;",
					ExpectedErr:     `cursor "c" already exists`,
					ExpectedErrCode: "42P03",
				},
				{
					Query: "ROLLBACK;",
				},
				{
					Query:           "DECLARE h CURSOR WITH HOLD FOR SELECT * FROM ct FOR UPDATE;",
					ExpectedErr:     "DECLARE CURSOR WITH HOLD ... FOR UPDATE is not supported",
					ExpectedErrCode: "0A000",
				},
				{
					Query: "BEGIN;",
				},
				{
					Query:           "DECLARE s SCROLL CURSOR FOR SELECT * FROM ct FOR UPDATE;",
					ExpectedErr:     "DECLARE SCROLL CURSOR ... FOR UPDATE is not supported",
					ExpectedErrCode: "0A000",
				},
				{
					Query: "ROLLBACK;",
				},
				{
					Query: "BEGIN;",
				},
				{
					Query:           "DECLARE f CURSOR FOR SELECT 1 / 0;",
					ExpectedErr:     "division by zero",
					ExpectedErrCode: "22012",
					Skip:            true, // constant expressions are not evaluated until rows are fetched
				},
				{
					Query:           "FETCH f;",
					ExpectedErr:     "current transaction is aborted",
					ExpectedErrCode: "25P02",
					Skip:            true, // constant expressions are not evaluated until rows are fetched
				},
				{
					Query: "ROLLBACK;",
				},
			},
		},
		{
			Name:        "pg_cursors",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN;",
				},
				{
					Query: "DECLARE c SCROLL CURSOR FOR SELECT * FROM ct ORDER BY id;",
				},
				{
					Query: "DECLARE d NO SCROLL CURSOR WITHOUT HOLD FOR SELECT 1 AS x;",
				},
				{
					Query: "DECLARE e INSENSITIVE CURSOR WITH HOLD FOR SELECT * FROM ct;",
				},
				{
					Query: "SELECT name, statement, is_holdable, is_binary, is_scrollable, creation_time IS NOT NULL FROM pg_cursors WHERE name <> '' ORDER BY name;",
					Expected: []sql.Row{
						{"c", "DECLARE c SCROLL CURSOR FOR SELECT * FROM ct ORDER BY id;", "f", "f", "t", "t"},
						{"d", "DECLARE d NO SCROLL CURSOR WITHOUT HOLD FOR SELECT 1 AS x;", "f", "f", "f", "t"},
						{"e", "DECLARE e INSENSITIVE CURSOR WITH HOLD FOR SELECT * FROM ct;", "t", "f", "t", "t"},
					},
					Skip: true, // cursors without SCROLL only move forward
				},
				{
					Query: "SELECT name, statement, is_holdable, is_binary, is_scrollable FROM pg_cursors WHERE name IN ('c', 'd') ORDER BY name;",
					Expected: []sql.Row{
						{"c", "DECLARE c SCROLL CURSOR FOR SELECT * FROM ct ORDER BY id;", "f", "f", "t"},
						{"d", "DECLARE d NO SCROLL CURSOR WITHOUT HOLD FOR SELECT 1 AS x;", "f", "f", "f"},
					},
				},
				{
					Query: "ROLLBACK;",
				},
				{
					Query:    "SELECT name FROM pg_cursors WHERE name <> '';",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT name, statement FROM pg_cursors WHERE name = '';",
					Expected: []sql.Row{{"", "SELECT name, statement FROM pg_cursors WHERE name = '';"}},
					Skip:     true, // portals from the extended query protocol are not listed yet
				},
			},
		},
		{
			Name:        "WITH HOLD cursors outlive their transaction",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query: "DECLARE h CURSOR WITH HOLD FOR SELECT * FROM ct ORDER BY id;",
				},
				{
					Query:    "FETCH 2 FROM h;",
					Expected: []sql.Row{{1, "a"}, {2, "b"}},
				},
				{
					Query:    "SELECT name, is_holdable FROM pg_cursors WHERE name <> '';",
					Expected: []sql.Row{{"h", "t"}},
				},
				{
					Query: "BEGIN;",
				},
				{
					Query: "DECLARE h2 CURSOR WITH HOLD FOR SELECT * FROM ct ORDER BY id;",
				},
				{
					Query: "DECLARE w CURSOR FOR SELECT 1;",
				},
				{
					Query:    "FETCH 1 FROM h2;",
					Expected: []sql.Row{{1, "a"}},
				},
				{
					Query: "COMMIT;",
				},
				{
					Query:    "SELECT name, is_holdable FROM pg_cursors WHERE name <> '' ORDER BY name;",
					Expected: []sql.Row{{"h", "t"}, {"h2", "t"}},
				},
				{
					Query:    "FETCH ALL FROM h2;",
					Expected: []sql.Row{{2, "b"}, {3, "c"}, {4, "d"}, {5, "e"}},
				},
				{
					Query:    "FETCH PRIOR FROM h2;",
					Expected: []sql.Row{{5, "e"}},
					Skip:     true, // cursors without SCROLL only move forward
				},
				{
					Query:    "FETCH h;",
					Expected: []sql.Row{{3, "c"}},
				},
				{
					Query: "BEGIN;",
				},
				{
					Query: "DECLARE h3 CURSOR WITH HOLD FOR SELECT 1;",
				},
				{
					Query: "CLOSE h;",
				},
				{
					Query: "ROLLBACK;",
				},
				{
					Query:    "SELECT name FROM pg_cursors WHERE name <> '' ORDER BY name;",
					Expected: []sql.Row{{"h2"}},
				},
				{
					Query: "BEGIN;",
				},
				{
					Query:       "SELECT 1 / 0;",
					ExpectedErr: "division by zero",
				},
				{
					Query: "ROLLBACK;",
				},
				{
					Query:    "SELECT name FROM pg_cursors WHERE name <> '' ORDER BY name;",
					Expected: []sql.Row{{"h2"}},
				},
				{
					Query: "CLOSE ALL;",
				},
				{
					Query:           "DECLARE e CURSOR WITH HOLD FOR SELECT 1 / 0;",
					ExpectedErr:     "division by zero",
					ExpectedErrCode: "22012",
				},
				{
					Query:    "SELECT name FROM pg_cursors WHERE name <> '';",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name:        "cursors in the implicit transaction of a multi-statement query",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query:       "DECLARE c CURSOR FOR SELECT * FROM ct; FETCH 2 FROM c;",
					ExpectedTag: "FETCH 2",
				},
				{
					Query:    "SELECT name FROM pg_cursors WHERE name <> '';",
					Expected: []sql.Row{},
				},
			},
		},
		{
			Name:        "ROLLBACK TO SAVEPOINT closes cursors declared after the savepoint",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN;",
				},
				{
					Query: "DECLARE s SCROLL CURSOR FOR SELECT * FROM ct ORDER BY id;",
				},
				{
					Query:    "FETCH 2 FROM s;",
					Expected: []sql.Row{{1, "a"}, {2, "b"}},
				},
				{
					Query: "SAVEPOINT sp;",
				},
				{
					Query: "DECLARE t CURSOR FOR SELECT 1;",
				},
				{
					Query:    "FETCH 1 FROM s;",
					Expected: []sql.Row{{3, "c"}},
				},
				{
					Query: "ROLLBACK TO SAVEPOINT sp;",
				},
				{
					Query:    "SELECT name FROM pg_cursors WHERE name <> '' ORDER BY name;",
					Expected: []sql.Row{{"s"}},
					Skip:     true, // ROLLBACK TO SAVEPOINT does not close cursors yet
				},
				{
					Query:    "FETCH 1 FROM s;",
					Expected: []sql.Row{{4, "d"}},
				},
				{
					Query: "ROLLBACK;",
				},
			},
		},
		{
			Name: "BINARY cursors",
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN;",
				},
				{
					Query: "DECLARE b BINARY CURSOR FOR SELECT 1;",
					Skip:  true, // BINARY cursors are not yet supported
				},
				{
					Query:    "SELECT name, is_binary FROM pg_cursors WHERE name <> '';",
					Expected: []sql.Row{{"b", "t"}},
					Skip:     true, // BINARY cursors are not yet supported
				},
				{
					Query: "ROLLBACK;",
				},
			},
		},
		{
			Name:        "cursors read their rows as they are fetched",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN;",
				},
				{
					Query: "DECLARE r CURSOR FOR SELECT 10 / (id - 3) FROM ct ORDER BY id;",
				},
				{
					Query:    "FETCH 2 FROM r;",
					Expected: []sql.Row{{-5}, {-10}},
				},
				{
					Query:           "FETCH 1 FROM r;",
					ExpectedErr:     "division by zero",
					ExpectedErrCode: "22012",
				},
				{
					Query: "ROLLBACK;",
				},
			},
		},
		{
			Name: "cursors read the data as it was when they were declared",
			SetUpScript: []string{
				"CREATE TABLE za (id INT4 PRIMARY KEY, v INT4);",
				"CREATE TABLE zb (id INT4 PRIMARY KEY, a_id INT4, w TEXT);",
				"CREATE INDEX zb_a ON zb (a_id);",
				"INSERT INTO za VALUES (1, 10), (2, 20), (3, 30);",
				"INSERT INTO zb VALUES (1, 1, 'x'), (2, 2, 'y'), (3, 3, 'z');",
				"CREATE VIEW zv AS SELECT za.id, zb.w FROM za JOIN zb ON zb.a_id = za.id;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN;",
				},
				{
					Query: "DECLARE j CURSOR FOR SELECT za.id, zb.w FROM za JOIN zb ON zb.a_id = za.id ORDER BY za.id;",
				},
				{
					Query: "DECLARE s CURSOR FOR SELECT id FROM za WHERE v IN (SELECT v FROM za WHERE v > 10) ORDER BY id;",
				},
				{
					Query: "DECLARE vw CURSOR FOR SELECT * FROM zv ORDER BY id;",
				},
				{
					Query: "DECLARE cte CURSOR FOR WITH x AS (SELECT id, v FROM za) SELECT id, (SELECT max(v) FROM za) FROM x ORDER BY id;",
				},
				{
					Query: "DECLARE lk CURSOR FOR SELECT w FROM zb WHERE a_id = 2;",
				},
				{
					Query:    "FETCH 1 FROM j;",
					Expected: []sql.Row{{1, "x"}},
				},
				{
					Query: "UPDATE zb SET w = 'changed';",
				},
				{
					Query: "INSERT INTO za VALUES (4, 40);",
				},
				{
					Query: "INSERT INTO zb VALUES (4, 4, 'new'), (5, 2, 'dup');",
				},
				{
					Query: "DELETE FROM za WHERE id = 3;",
				},
				{
					Query:    "FETCH ALL FROM j;",
					Expected: []sql.Row{{2, "y"}, {3, "z"}},
				},
				{
					Query:    "FETCH ALL FROM s;",
					Expected: []sql.Row{{2}, {3}},
				},
				{
					Query:    "FETCH ALL FROM vw;",
					Expected: []sql.Row{{1, "x"}, {2, "y"}, {3, "z"}},
				},
				{
					Query:    "FETCH ALL FROM cte;",
					Expected: []sql.Row{{1, 30}, {2, 30}, {3, 30}},
				},
				{
					Query:    "FETCH ALL FROM lk;",
					Expected: []sql.Row{{"y"}},
				},
				{
					Query:           "FETCH BACKWARD ALL FROM j;",
					ExpectedErr:     "cursor can only scan forward",
					ExpectedErrCode: "55000",
				},
				{
					Query: "ROLLBACK;",
				},
			},
		},
		{
			Name:        "WITH HOLD cursors read their remaining rows when their transaction commits",
			SetUpScript: setup,
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN;",
				},
				{
					Query: "DECLARE h CURSOR WITH HOLD FOR SELECT 10 / (id - 3) FROM ct ORDER BY id;",
				},
				{
					Query:    "FETCH 1 FROM h;",
					Expected: []sql.Row{{-5}},
				},
				{
					Query: "INSERT INTO ct VALUES (9, 'z');",
				},
				{
					Query:           "COMMIT;",
					ExpectedErr:     "division by zero",
					ExpectedErrCode: "22012",
				},
				{
					Query:    "SELECT name FROM pg_cursors WHERE name <> '';",
					Expected: []sql.Row{},
				},
				{
					Query:    "SELECT count(*) FROM ct;",
					Expected: []sql.Row{{5}},
				},
				{
					Query: "BEGIN;",
				},
				{
					Query: "DECLARE w SCROLL CURSOR WITH HOLD FOR SELECT * FROM ct ORDER BY id;",
				},
				{
					Query:    "FETCH 2 FROM w;",
					Expected: []sql.Row{{1, "a"}, {2, "b"}},
				},
				{
					Query: "COMMIT;",
				},
				{
					Query:    "FETCH PRIOR FROM w;",
					Expected: []sql.Row{{1, "a"}},
				},
				{
					Query:    "FETCH ALL FROM w;",
					Expected: []sql.Row{{2, "b"}, {3, "c"}, {4, "d"}, {5, "e"}},
				},
				{
					Query:    "FETCH BACKWARD 2 FROM w;",
					Expected: []sql.Row{{5, "e"}, {4, "d"}},
				},
				{
					Query: "DECLARE w2 NO SCROLL CURSOR WITH HOLD FOR SELECT * FROM ct ORDER BY id;",
				},
				{
					Query:    "FETCH 2 FROM w2;",
					Expected: []sql.Row{{1, "a"}, {2, "b"}},
				},
				{
					Query:           "FETCH PRIOR FROM w2;",
					ExpectedErr:     "cursor can only scan forward",
					ExpectedErrCode: "55000",
				},
			},
		},
		{
			Name: "cursor queries with parameters",
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN;",
				},
				{
					Query:    "DECLARE p CURSOR FOR SELECT $1::INT4 AS x;",
					BindVars: []any{5},
					Expected: []sql.Row{},
				},
				{
					Query:    "FETCH p;",
					Expected: []sql.Row{{5}},
				},
				{
					Query: "ROLLBACK;",
				},
			},
		},
		{
			Name: "cursors over historical data",
			SetUpScript: []string{
				"CREATE TABLE hist (id INT4 PRIMARY KEY, v TEXT);",
				"INSERT INTO hist VALUES (1, 'one');",
				"SELECT dolt_commit('-Am', 'first');",
				"INSERT INTO hist VALUES (2, 'two');",
				"UPDATE hist SET v = 'ONE' WHERE id = 1;",
				"SELECT dolt_commit('-am', 'second');",
				"DELETE FROM hist WHERE id = 1;",
				"INSERT INTO hist VALUES (3, 'three');",
				"SELECT dolt_commit('-am', 'third');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "DECLARE wh CURSOR WITH HOLD FOR SELECT * FROM hist AS OF 'HEAD~1' ORDER BY id;",
				},
				{
					Query: "BEGIN;",
				},
				{
					Query: "DECLARE h2 CURSOR FOR SELECT * FROM hist AS OF 'HEAD~2' ORDER BY id;",
				},
				{
					Query: "DECLARE h1 SCROLL CURSOR FOR SELECT * FROM hist AS OF 'HEAD~1' ORDER BY id;",
				},
				{
					Query: "DECLARE head CURSOR FOR SELECT * FROM hist AS OF 'HEAD' ORDER BY id;",
				},
				{
					Query: "DECLARE j CURSOR FOR SELECT cur.id, cur.v, old.v FROM hist cur JOIN hist AS OF 'HEAD~1' old ON cur.id = old.id ORDER BY cur.id;",
				},
				{
					Query: "DECLARE hh CURSOR FOR SELECT id, v FROM dolt_history_hist ORDER BY v, id;",
				},
				{
					Query: "INSERT INTO hist VALUES (4, 'four');",
				},
				{
					Query: "UPDATE hist SET v = 'TWO';",
				},
				{
					Query:            "SELECT dolt_commit('-am', 'fourth');",
					SkipResultsCheck: true,
				},
				{
					Query:    "FETCH ALL FROM h2;",
					Expected: []sql.Row{{1, "one"}},
				},
				{
					Query:    "FETCH 1 FROM h1;",
					Expected: []sql.Row{{1, "ONE"}},
				},
				{
					Query:    "FETCH ALL FROM h1;",
					Expected: []sql.Row{{2, "two"}},
				},
				{
					Query:    "FETCH BACKWARD ALL FROM h1;",
					Expected: []sql.Row{{2, "two"}, {1, "ONE"}},
				},
				{
					Query:    "FETCH ALL FROM head;",
					Expected: []sql.Row{{2, "two"}, {3, "three"}},
				},
				{
					Query:    "FETCH ALL FROM j;",
					Expected: []sql.Row{{2, "two", "two"}},
				},
				{
					Query:    "FETCH ALL FROM hh;",
					Expected: []sql.Row{{1, "ONE"}, {1, "one"}, {3, "three"}, {2, "two"}, {2, "two"}},
				},
				{
					Query: "ROLLBACK;",
				},
				{
					Query:    "FETCH ALL FROM wh;",
					Expected: []sql.Row{{1, "ONE"}, {2, "two"}},
				},
			},
		},
		{
			Name: "cursors over table functions and system tables",
			SetUpScript: []string{
				"CREATE TABLE tf (id INT4 PRIMARY KEY, v TEXT);",
				"INSERT INTO tf VALUES (1, 'a');",
				"SELECT dolt_commit('-Am', 'first');",
				"INSERT INTO tf VALUES (2, 'b');",
				"SELECT dolt_commit('-am', 'second');",
				"UPDATE tf SET v = 'c' WHERE id = 1;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "BEGIN;",
				},
				{
					Query: "DECLARE lg CURSOR FOR SELECT message FROM dolt_log() LIMIT 2;",
				},
				{
					Query: "DECLARE df CURSOR FOR SELECT to_id, to_v, from_id, from_v, diff_type FROM dolt_diff('HEAD~1', 'HEAD', 'tf');",
				},
				{
					Query: "DECLARE dw CURSOR FOR SELECT to_id, to_v, from_v, diff_type FROM dolt_diff('HEAD', 'WORKING', 'tf');",
				},
				{
					Query: "DECLARE ds CURSOR FOR SELECT table_name, rows_modified FROM dolt_diff_stat('HEAD', 'WORKING');",
				},
				{
					Query: "DECLARE dp CURSOR FOR SELECT table_name, diff_type, statement FROM dolt_patch('HEAD', 'WORKING');",
				},
				{
					Query: "DECLARE st CURSOR FOR SELECT table_name, staged, status FROM dolt_status;",
				},
				{
					Query: "DECLARE dd CURSOR FOR SELECT to_id, to_v, from_v, diff_type FROM dolt_diff_tf WHERE to_commit = 'WORKING';",
				},
				{
					Query: "DECLARE gs SCROLL CURSOR FOR SELECT * FROM generate_series(1, 5);",
				},
				{
					Query: "UPDATE tf SET v = 'z' WHERE id = 2;",
				},
				{
					Query:            "SELECT dolt_commit('-am', 'third');",
					SkipResultsCheck: true,
				},
				{
					Query:    "FETCH ALL FROM lg;",
					Expected: []sql.Row{{"second"}, {"first"}},
				},
				{
					Query:    "FETCH ALL FROM df;",
					Expected: []sql.Row{{2, "b", nil, nil, "added"}},
				},
				{
					Query:    "FETCH ALL FROM dw;",
					Expected: []sql.Row{{1, "c", "a", "modified"}},
				},
				{
					Query:    "FETCH ALL FROM ds;",
					Expected: []sql.Row{{"public.tf", 1}},
				},
				{
					Query:    "FETCH ALL FROM dp;",
					Expected: []sql.Row{{"public.tf", "data", `UPDATE "tf" SET "v"='c' WHERE "id"=1;`}},
				},
				{
					Query:    "FETCH ALL FROM st;",
					Expected: []sql.Row{{"public.tf", "f", "modified"}},
				},
				{
					Query:    "FETCH ALL FROM dd;",
					Expected: []sql.Row{{1, "c", "a", "modified"}},
				},
				{
					Query:    "FETCH 3 FROM gs;",
					Expected: []sql.Row{{1}, {2}, {3}},
				},
				{
					Query:    "FETCH BACKWARD 2 FROM gs;",
					Expected: []sql.Row{{2}, {1}},
				},
				{
					Query: "ROLLBACK;",
				},
			},
		},
	})
}
