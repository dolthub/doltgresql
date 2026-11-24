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

// TestLookupJoinExtendedKeyStoragePosition tests lookup joins whose key column is an
// extended-encoded type (e.g. uuid) that sits at a non-zero storage position among the
// source table's non-primary-key columns.
// Regression test for https://github.com/dolthub/doltgresql/issues/2979
func TestLookupJoinExtendedKeyStoragePosition(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "join key is the second non-PK column",
			SetUpScript: []string{
				"CREATE TABLE c (id uuid PRIMARY KEY, name text);",
				"CREATE TABLE m (seq int PRIMARY KEY, filler text, company_id uuid, label text);",
				"INSERT INTO c VALUES ('11111111-1111-1111-1111-111111111111', 'acme');",
				"INSERT INTO m VALUES (1, 'f', '11111111-1111-1111-1111-111111111111', 'L');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select /*+ lookup_join(m, c) */ HINT count(*) from m join c on c.id = m.company_id;",
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "join key is the first non-PK column",
			SetUpScript: []string{
				"CREATE TABLE c (id uuid PRIMARY KEY, name text);",
				"CREATE TABLE m (seq int PRIMARY KEY, company_id uuid, label text);",
				"INSERT INTO c VALUES ('11111111-1111-1111-1111-111111111111', 'acme');",
				"INSERT INTO m VALUES (1, '11111111-1111-1111-1111-111111111111', 'L');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select /*+ lookup_join(m, c) */ HINT count(*) from m join c on c.id = m.company_id;",
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "join key is the third non-PK column",
			SetUpScript: []string{
				"CREATE TABLE c (id uuid PRIMARY KEY, name text);",
				"CREATE TABLE m (seq int PRIMARY KEY, filler1 text, filler2 text, company_id uuid, label text);",
				"INSERT INTO c VALUES ('11111111-1111-1111-1111-111111111111', 'acme');",
				"INSERT INTO m VALUES (1, 'f1', 'f2', '11111111-1111-1111-1111-111111111111', 'L');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "select /*+ lookup_join(m, c) */ HINT count(*) from m join c on c.id = m.company_id;",
					Expected: []sql.Row{{1}},
				},
			},
		},
		{
			Name: "composite key with both columns at non-zero storage positions",
			SetUpScript: []string{
				"CREATE TABLE c (id uuid PRIMARY KEY, tag text, code uuid, label text);",
				"CREATE TABLE m (seq int PRIMARY KEY, filler text, ref_code uuid, ref_label text);",
				"INSERT INTO c VALUES ('11111111-1111-1111-1111-111111111111', 'c1', '22222222-2222-2222-2222-222222222222', 'match');",
				"INSERT INTO c VALUES ('33333333-3333-3333-3333-333333333333', 'c2', '22222222-2222-2222-2222-222222222222', 'other');",
				"INSERT INTO m VALUES (1, 'f', '22222222-2222-2222-2222-222222222222', 'match');",
				"INSERT INTO m VALUES (2, 'f', '22222222-2222-2222-2222-222222222222', 'other');",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "select /*+ lookup_join(m, c) */ HINT m.seq, c.tag from m join c on c.code = m.ref_code and c.label = m.ref_label order by m.seq;",
					Expected: []sql.Row{
						{1, "c1"},
						{2, "c2"},
					},
				},
			},
		},
	})
}

// Expected results were checked against PostgreSQL 18.
func TestRowLookupJoinConversions(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "large integer to float",
			SetUpScript: []string{
				`CREATE TABLE lsrc (id INT PRIMARY KEY, v BIGINT)`,
				`CREATE TABLE rdst (id INT PRIMARY KEY, v DOUBLE PRECISION)`,
				`CREATE INDEX v_idx ON rdst(v)`,
				`INSERT INTO lsrc VALUES (1,9007199254740992),(2,9007199254740993),(3,9007199254740994),(4,NULL)`,
				`INSERT INTO rdst VALUES (1,9007199254740992),(2,9007199254740994),(3,NULL)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
JOIN rdst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, 1}, {2, 1}, {3, 2}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, 1}, {2, 1}, {3, 2}, {4, nil}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v AND r.id > 1
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, 2}, {4, nil}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v AND r.id < 0
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, nil}, {4, nil}},
				},
			},
		},
		{
			Name: "text to varchar",
			SetUpScript: []string{
				`CREATE TABLE lsrc (id INT PRIMARY KEY, v TEXT)`,
				`CREATE TABLE rdst (id INT PRIMARY KEY, v VARCHAR(3))`,
				`CREATE INDEX v_idx ON rdst(v)`,
				`INSERT INTO lsrc VALUES (1,'a'),(2,'abc'),(3,'abcd'),(4,''),(5,'é'),(6,NULL),(7,'b')`,
				`INSERT INTO rdst VALUES (1,'a'),(2,'abc'),(3,''),(4,'é'),(5,NULL),(6,'b')`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
JOIN rdst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, 1}, {2, 2}, {4, 3}, {5, 4}, {7, 6}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, 1}, {2, 2}, {3, nil}, {4, 3}, {5, 4}, {6, nil}, {7, 6}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v AND r.id > 1
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, 2}, {3, nil}, {4, 3}, {5, 4}, {6, nil}, {7, 6}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v AND r.id < 0
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, nil}, {4, nil}, {5, nil}, {6, nil}, {7, nil}},
				},
			},
		},
		{
			Name: "varchar to text",
			SetUpScript: []string{
				`CREATE TABLE lsrc (id INT PRIMARY KEY, v VARCHAR(20))`,
				`CREATE TABLE rdst (id INT PRIMARY KEY, v TEXT)`,
				`CREATE INDEX v_idx ON rdst(v)`,
				`INSERT INTO lsrc VALUES (1,'alpha'),(2,'beta'),(3,'missing'),(4,NULL)`,
				`INSERT INTO rdst VALUES (1,'alpha'),(2,'beta'),(3,NULL)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
JOIN rdst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, 1}, {2, 2}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, 1}, {2, 2}, {3, nil}, {4, nil}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v AND r.id > 1
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, 2}, {3, nil}, {4, nil}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v AND r.id < 0
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, nil}, {4, nil}},
				},
			},
		},
		{
			Name: "bigint to smallint",
			SetUpScript: []string{
				`CREATE TABLE lsrc (id INT PRIMARY KEY, v BIGINT)`,
				`CREATE TABLE rdst (id INT PRIMARY KEY, v SMALLINT)`,
				`CREATE INDEX v_idx ON rdst(v)`,
				`INSERT INTO lsrc VALUES (1,-32769),(2,-32768),(3,32767),(4,32768),(5,NULL),(6,1)`,
				`INSERT INTO rdst VALUES (1,-32768),(2,32767),(3,NULL),(4,1)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
JOIN rdst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{2, 1}, {3, 2}, {6, 4}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, 1}, {3, 2}, {4, nil}, {5, nil}, {6, 4}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v AND r.id > 1
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, 2}, {4, nil}, {5, nil}, {6, 4}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v AND r.id < 0
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, nil}, {4, nil}, {5, nil}, {6, nil}},
				},
			},
		},
		{
			Name: "integer to bigint",
			SetUpScript: []string{
				`CREATE TABLE lsrc (id INT PRIMARY KEY, v INTEGER)`,
				`CREATE TABLE rdst (id INT PRIMARY KEY, v BIGINT)`,
				`CREATE INDEX v_idx ON rdst(v)`,
				`INSERT INTO lsrc VALUES (1,-2147483648),(2,2147483647),(3,NULL),(4,1)`,
				`INSERT INTO rdst VALUES (1,-2147483648),(2,2147483647),(3,2147483648),(4,NULL),(5,1)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
JOIN rdst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, 1}, {2, 2}, {4, 5}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, 1}, {2, 2}, {3, nil}, {4, 5}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v AND r.id > 1
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, 2}, {3, nil}, {4, 5}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v AND r.id < 0
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, nil}, {4, nil}},
				},
			},
		},
		{
			Name: "numeric to integer",
			SetUpScript: []string{
				`CREATE TABLE lsrc (id INT PRIMARY KEY, v NUMERIC(12,3))`,
				`CREATE TABLE rdst (id INT PRIMARY KEY, v INTEGER)`,
				`CREATE INDEX v_idx ON rdst(v)`,
				`INSERT INTO lsrc VALUES (1,-1.5),(2,-1),(3,0.5),(4,1),(5,1.5),(6,2),(7,NULL)`,
				`INSERT INTO rdst VALUES (1,-2),(2,-1),(3,0),(4,1),(5,2),(6,NULL)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
JOIN rdst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{2, 2}, {4, 4}, {6, 5}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, 2}, {3, nil}, {4, 4}, {5, nil}, {6, 5}, {7, nil}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v AND r.id > 1
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, 2}, {3, nil}, {4, 4}, {5, nil}, {6, 5}, {7, nil}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v AND r.id < 0
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, nil}, {4, nil}, {5, nil}, {6, nil}, {7, nil}},
				},
			},
		},
		{
			Name: "float to integer",
			SetUpScript: []string{
				`CREATE TABLE lsrc (id INT PRIMARY KEY, v DOUBLE PRECISION)`,
				`CREATE TABLE rdst (id INT PRIMARY KEY, v INTEGER)`,
				`CREATE INDEX v_idx ON rdst(v)`,
				`INSERT INTO lsrc VALUES (1,-1.5),(2,0.5),(3,1),(4,1.49),(5,1.5),(6,2),(7,NULL)`,
				`INSERT INTO rdst VALUES (1,-2),(2,0),(3,1),(4,2),(5,NULL)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
JOIN rdst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{3, 3}, {6, 4}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, 3}, {4, nil}, {5, nil}, {6, 4}, {7, nil}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v AND r.id > 1
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, 3}, {4, nil}, {5, nil}, {6, 4}, {7, nil}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v AND r.id < 0
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, nil}, {4, nil}, {5, nil}, {6, nil}, {7, nil}},
				},
			},
		},
		{
			Name: "numeric scale",
			SetUpScript: []string{
				`CREATE TABLE lsrc (id INT PRIMARY KEY, v NUMERIC(12,3))`,
				`CREATE TABLE rdst (id INT PRIMARY KEY, v NUMERIC(5,2))`,
				`CREATE INDEX v_idx ON rdst(v)`,
				`INSERT INTO lsrc VALUES (1,1.234),(2,1.235),(3,1.230),(4,-1.235),(5,999.999),(6,NULL),(7,1.2)`,
				`INSERT INTO rdst VALUES (1,1.23),(2,1.24),(3,-1.24),(4,999.99),(5,NULL),(6,1.2)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
JOIN rdst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{3, 1}, {7, 6}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, 1}, {4, nil}, {5, nil}, {6, nil}, {7, 6}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v AND r.id > 1
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, nil}, {4, nil}, {5, nil}, {6, nil}, {7, 6}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v AND r.id < 0
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, nil}, {4, nil}, {5, nil}, {6, nil}, {7, nil}},
				},
				{
					Query:       "INSERT INTO rdst VALUES (100, 1000.00)",
					ExpectedErr: "numeric field overflow",
				},
				{
					Query:       "SELECT CAST(1000 AS NUMERIC(5,2))",
					ExpectedErr: "numeric field overflow",
				},
				{
					Query:    "SELECT COUNT(*) FROM rdst WHERE id=100",
					Expected: []sql.Row{{0}},
				},
			},
		},
		{
			Name: "date to timestamp",
			SetUpScript: []string{
				`CREATE TABLE lsrc (id INT PRIMARY KEY, v DATE)`,
				`CREATE TABLE rdst (id INT PRIMARY KEY, v TIMESTAMP)`,
				`CREATE INDEX v_idx ON rdst(v)`,
				`INSERT INTO lsrc VALUES (1,'2020-01-01'),(2,'2020-01-02'),(3,NULL)`,
				`INSERT INTO rdst VALUES (1,'2020-01-01'),(2,'2020-01-01 12:00:00'),(3,'2020-01-02'),(4,NULL)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
JOIN rdst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, 1}, {2, 3}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, 1}, {2, 3}, {3, nil}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v AND r.id > 1
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, 3}, {3, nil}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v AND r.id < 0
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, nil}},
				},
			},
		},
		{
			Name: "uuid row source",
			SetUpScript: []string{
				`CREATE TABLE lsrc (id INT PRIMARY KEY, v UUID)`,
				`CREATE TABLE rdst (id INT PRIMARY KEY, v UUID)`,
				`CREATE INDEX v_idx ON rdst(v)`,
				`INSERT INTO lsrc VALUES (1,'11111111-1111-1111-1111-111111111111'),(2,'22222222-2222-2222-2222-222222222222'),(3,NULL)`,
				`INSERT INTO rdst VALUES (1,'11111111-1111-1111-1111-111111111111'),(2,NULL)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
JOIN rdst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, 1}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, 1}, {2, nil}, {3, nil}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v AND r.id > 1
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, nil}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
FROM (SELECT id, v FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v = l.v AND r.id < 0
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, nil}},
				},
			},
		},
		{
			Name: "out of band text to varchar",
			SetUpScript: []string{`CREATE TABLE lsrc (id INT PRIMARY KEY, v TEXT)`,
				`CREATE TABLE rdst (id INT PRIMARY KEY, v VARCHAR(25000))`,
				`CREATE INDEX v_idx ON rdst(v)`,
				`INSERT INTO lsrc VALUES (1, REPEAT('x',20000)), (2,REPEAT('y',20000)), (3,NULL), (4, REPEAT('x',20000))`,
				`INSERT INTO rdst VALUES (1,REPEAT('x',20000)), (2,NULL)`,
			},
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT /*+ LOOKUP_JOIN(l,r) */ HINT l.id, r.id
FROM (SELECT * FROM lsrc LIMIT 1000) l
LEFT JOIN rdst r ON r.v=l.v
ORDER BY l.id, r.id`, Expected: []sql.Row{{1, 1}, {2, nil}, {3, nil}, {4, 1}}},
			},
		},
	})
}

// Expected results were checked against PostgreSQL 18.6.
func TestRowLookupJoinShapes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "multi-table lookup join shapes",
			SetUpScript: []string{
				`CREATE TABLE shape_src (id INT PRIMARY KEY, v BIGINT)`,
				`CREATE TABLE shape_mid (id INT PRIMARY KEY, v SMALLINT)`,
				`CREATE INDEX mid_v ON shape_mid(v)`,
				`CREATE TABLE shape_tail (id INT PRIMARY KEY, v INT)`,
				`CREATE INDEX tail_v ON shape_tail(v)`,
				`CREATE TABLE shape_bridge (v SMALLINT, tail_id INT, payload INT, PRIMARY KEY(v,tail_id))`,
				`INSERT INTO shape_src VALUES (1,1),(2,2),(3,32768),(4,NULL),(5,3),(6,1),(7,9)`,
				`INSERT INTO shape_mid VALUES (11,1),(12,1),(13,2),(14,NULL),(15,3)`,
				`INSERT INTO shape_tail VALUES (21,11),(22,11),(23,13),(24,NULL),(25,99)`,
				`INSERT INTO shape_bridge VALUES (1,21,101),(1,22,102),(2,23,103)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// subquery table subquery.
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r,q) */ HINT l.id, r.id, q.id
FROM (SELECT * FROM shape_src LIMIT 100) l
JOIN shape_mid r ON r.v = l.v
JOIN (SELECT * FROM shape_tail LIMIT 100) q ON q.v = r.id
ORDER BY l.id NULLS FIRST, r.id NULLS FIRST, q.id NULLS FIRST`,
					Expected: []sql.Row{{1, 11, 21}, {1, 11, 22}, {2, 13, 23}, {6, 11, 21}, {6, 11, 22}},
				},
				{
					// outer subquery table subquery.
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r,q) */ HINT l.id, r.id, q.id
FROM (SELECT * FROM shape_src LIMIT 100) l
LEFT JOIN shape_mid r ON r.v = l.v
LEFT JOIN (SELECT * FROM shape_tail LIMIT 100) q ON q.v = r.id
ORDER BY l.id NULLS FIRST, r.id NULLS FIRST, q.id NULLS FIRST`,
					Expected: []sql.Row{{1, 11, 21}, {1, 11, 22}, {1, 12, nil}, {2, 13, 23}, {3, nil, nil}, {4, nil, nil}, {5, 15, nil}, {6, 11, 21}, {6, 11, 22}, {6, 12, nil}, {7, nil, nil}},
				},
				{
					// two consecutive index lookups.
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) LOOKUP_JOIN(r,q) JOIN_ORDER(l,r,q) */ HINT l.id, r.id, q.id
FROM (SELECT * FROM shape_src LIMIT 100) l
LEFT JOIN shape_mid r ON r.v = l.v
LEFT JOIN shape_tail q ON q.v = r.id
ORDER BY l.id NULLS FIRST, r.id NULLS FIRST, q.id NULLS FIRST`,
					Expected: []sql.Row{{1, 11, 21}, {1, 11, 22}, {1, 12, nil}, {2, 13, 23}, {3, nil, nil}, {4, nil, nil}, {5, 15, nil}, {6, 11, 21}, {6, 11, 22}, {6, 12, nil}, {7, nil, nil}},
				},
				{
					// two subqueries feed an index lookup.
					Query: `SELECT /*+ LOOKUP_JOIN(q,r) JOIN_ORDER(l,q,r) */ HINT l.id, q.id, r.id
FROM (SELECT * FROM shape_src LIMIT 100) l
JOIN (SELECT * FROM shape_src LIMIT 100) q ON q.v = l.v AND q.id > l.id
LEFT JOIN shape_mid r ON r.v = q.v
ORDER BY l.id NULLS FIRST, q.id NULLS FIRST, r.id NULLS FIRST`,
					Expected: []sql.Row{{1, 6, 11}, {1, 6, 12}},
				},
				{
					// materialized join feeds an index lookup.
					Query: `SELECT /*+ LOOKUP_JOIN(j,q) JOIN_ORDER(j,q) */ HINT j.lid, j.rid, q.id
FROM (
 SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid
 FROM (SELECT * FROM shape_src LIMIT 100) l
 LEFT JOIN shape_mid r ON r.v = l.v
 LIMIT 100
) j
LEFT JOIN shape_tail q ON q.v = j.rid
ORDER BY j.lid NULLS FIRST, j.rid NULLS FIRST, q.id NULLS FIRST`,
					Expected: []sql.Row{{1, 11, 21}, {1, 11, 22}, {1, 12, nil}, {2, 13, 23}, {3, nil, nil}, {4, nil, nil}, {5, 15, nil}, {6, 11, 21}, {6, 11, 22}, {6, 12, nil}, {7, nil, nil}},
				},
				{
					// aggregates on both sides of indexed table.
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r,q) */ HINT l.v, l.n, r.id, q.n
FROM (SELECT v, COUNT(*) AS n FROM shape_src GROUP BY v) l
LEFT JOIN shape_mid r ON r.v = l.v
LEFT JOIN (SELECT v, COUNT(*) AS n FROM shape_tail GROUP BY v) q ON q.v = r.id
ORDER BY l.v NULLS FIRST, r.id NULLS FIRST`,
					Expected: []sql.Row{{nil, 1, nil, nil}, {1, 2, 11, 2}, {1, 2, 12, nil}, {2, 1, 13, 1}, {3, 1, 15, nil}, {9, 1, nil, nil}, {32768, 1, nil, nil}},
				},
				{
					// union source and filtered subquery.
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r,q) */ HINT l.id, r.id, q.id
FROM (
 SELECT id, v FROM shape_src WHERE id <= 2
 UNION ALL
 SELECT id, v FROM shape_src WHERE id = 1
) l
LEFT JOIN shape_mid r ON r.v = l.v AND r.id <> 12
LEFT JOIN (SELECT * FROM shape_tail WHERE id > 21 LIMIT 100) q ON q.v = r.id
ORDER BY l.id NULLS FIRST, r.id NULLS FIRST, q.id NULLS FIRST`,
					Expected: []sql.Row{{1, 11, 22}, {1, 11, 22}, {2, 13, 23}},
				},
				{
					// composite lookup uses two earlier relations.
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) LOOKUP_JOIN(r,b) JOIN_ORDER(l,r,q,b) */ HINT l.id, r.id, q.id, b.payload
FROM (SELECT * FROM shape_src LIMIT 100) l
LEFT JOIN shape_mid r ON r.v = l.v
LEFT JOIN (SELECT * FROM shape_tail LIMIT 100) q ON q.v = r.id
LEFT JOIN shape_bridge b ON b.v = r.v AND b.tail_id = q.id
ORDER BY l.id NULLS FIRST, r.id NULLS FIRST, q.id NULLS FIRST`,
					Expected: []sql.Row{{1, 11, 21, 101}, {1, 11, 22, 102}, {1, 12, nil, nil}, {2, 13, 23, 103}, {3, nil, nil, nil}, {4, nil, nil, nil}, {5, 15, nil, nil}, {6, 11, 21, 101}, {6, 11, 22, 102}, {6, 12, nil, nil}, {7, nil, nil, nil}},
				},
				{
					// reused cte around indexed table.
					Query: `WITH s AS (SELECT * FROM shape_src LIMIT 100)
SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r,q) */ HINT l.id, r.id, q.id
FROM s l
JOIN shape_mid r ON r.v = l.v
JOIN s q ON q.v = r.v AND q.id > l.id
ORDER BY l.id NULLS FIRST, r.id NULLS FIRST, q.id NULLS FIRST`,
					Expected: []sql.Row{{1, 11, 6}, {1, 12, 6}},
				},
				{
					// computed join output feeds lookup.
					Query: `SELECT /*+ LOOKUP_JOIN(j,q) JOIN_ORDER(j,q) */ HINT j.lid, j.rid, q.id
FROM (
 SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id AS lid, r.id AS rid, r.id + 10 AS next_id
 FROM (SELECT * FROM shape_src LIMIT 100) l
 LEFT JOIN shape_mid r ON r.v = l.v
 LIMIT 100
) j
LEFT JOIN shape_tail q ON q.id = j.next_id
ORDER BY j.lid NULLS FIRST, j.rid NULLS FIRST, q.id NULLS FIRST`,
					Expected: []sql.Row{{1, 11, 21}, {1, 12, 22}, {2, 13, 23}, {3, nil, nil}, {4, nil, nil}, {5, 15, 25}, {6, 11, 21}, {6, 12, 22}, {7, nil, nil}},
				},
				{
					// rejected middle rows propagate nulls.
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) LOOKUP_JOIN(r,q) JOIN_ORDER(l,r,q) */ HINT l.id, r.id, q.id
FROM (SELECT * FROM shape_src LIMIT 100) l
LEFT JOIN shape_mid r ON r.v = l.v AND r.id < 0
LEFT JOIN shape_tail q ON q.v = r.id
ORDER BY l.id NULLS FIRST, r.id NULLS FIRST, q.id NULLS FIRST`,
					Expected: []sql.Row{{1, nil, nil}, {2, nil, nil}, {3, nil, nil}, {4, nil, nil}, {5, nil, nil}, {6, nil, nil}, {7, nil, nil}},
				},
				{
					// outer then inner join rejects missing middle rows.
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) LOOKUP_JOIN(r,q) JOIN_ORDER(l,r,q) */ HINT l.id, r.id, q.id
FROM (SELECT * FROM shape_src LIMIT 100) l
LEFT JOIN shape_mid r ON r.v = l.v
JOIN shape_tail q ON q.v = r.id
ORDER BY l.id NULLS FIRST, r.id NULLS FIRST, q.id NULLS FIRST`,
					Expected: []sql.Row{{1, 11, 21}, {1, 11, 22}, {2, 13, 23}, {6, 11, 21}, {6, 11, 22}},
				},
				{
					// limited result across nested joins.
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r,q) */ HINT l.id, r.id, q.id
FROM (SELECT * FROM shape_src LIMIT 100) l
LEFT JOIN shape_mid r ON r.v = l.v
LEFT JOIN (SELECT * FROM shape_tail LIMIT 100) q ON q.v = r.id
ORDER BY l.id NULLS FIRST, r.id NULLS FIRST, q.id NULLS FIRST LIMIT 3 OFFSET 1`,
					Expected: []sql.Row{{1, 11, 22}, {1, 12, nil}, {2, 13, 23}},
				},
			},
		},
	})
}

// Expected results were checked against PostgreSQL 18.6. These indexes cover
// complete adaptive values, including values stored out of band.
func TestRowLookupJoinAdaptiveKeys(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "bounded varchar to adaptive text",
			SetUpScript: []string{
				`CREATE TABLE adaptive_src (id INT PRIMARY KEY, v VARCHAR(3000))`,
				`CREATE TABLE adaptive_dst (id INT PRIMARY KEY, v TEXT, bucket INT)`,
				`CREATE INDEX adaptive_v ON adaptive_dst(v,bucket)`,
				`INSERT INTO adaptive_src VALUES (1,''),(2,'abc'),(3,'abcd'),(4,'é日本語'),(5,repeat('x',2200)),(6,NULL),(7,'missing')`,
				`INSERT INTO adaptive_dst VALUES (1,'',0),(2,'abc',1),(3,'é日本語',0),(4,repeat('x',2200),1),(5,NULL,0),(6,'abc',1)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id, r.id
FROM (SELECT * FROM adaptive_src LIMIT 100) l
JOIN adaptive_dst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, 1}, {2, 2}, {2, 6}, {4, 3}, {5, 4}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id, r.id
FROM (SELECT * FROM adaptive_src LIMIT 100) l
LEFT JOIN adaptive_dst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, 1}, {2, 2}, {2, 6}, {3, nil}, {4, 3}, {5, 4}, {6, nil}, {7, nil}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id, r.id
FROM (SELECT * FROM adaptive_src LIMIT 100) l
LEFT JOIN adaptive_dst r ON r.v = l.v AND r.bucket = 1
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, 2}, {2, 6}, {3, nil}, {4, nil}, {5, 4}, {6, nil}, {7, nil}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id, r.id
FROM (SELECT * FROM adaptive_src LIMIT 100) l
LEFT JOIN adaptive_dst r ON r.v = l.v AND r.id < 0
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, nil}, {4, nil}, {5, nil}, {6, nil}, {7, nil}},
				},
			},
		},
		{
			Name: "adaptive text to unbounded varchar",
			SetUpScript: []string{
				`CREATE TABLE adaptive_src (id INT PRIMARY KEY, v TEXT)`,
				`CREATE TABLE adaptive_dst (id INT PRIMARY KEY, v VARCHAR, bucket INT)`,
				`CREATE INDEX adaptive_v ON adaptive_dst(v,bucket)`,
				`INSERT INTO adaptive_src VALUES (1,''),(2,'abc'),(3,'abcd'),(4,'é日本語'),(5,repeat('x',20000)),(6,NULL),(7,'missing')`,
				`INSERT INTO adaptive_dst VALUES (1,'',0),(2,'abc',1),(3,'é日本語',0),(4,repeat('x',20000),1),(5,NULL,0),(6,'abc',1)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id, r.id
FROM (SELECT * FROM adaptive_src LIMIT 100) l
JOIN adaptive_dst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, 1}, {2, 2}, {2, 6}, {4, 3}, {5, 4}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id, r.id
FROM (SELECT * FROM adaptive_src LIMIT 100) l
LEFT JOIN adaptive_dst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, 1}, {2, 2}, {2, 6}, {3, nil}, {4, 3}, {5, 4}, {6, nil}, {7, nil}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id, r.id
FROM (SELECT * FROM adaptive_src LIMIT 100) l
LEFT JOIN adaptive_dst r ON r.v = l.v AND r.bucket = 1
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, 2}, {2, 6}, {3, nil}, {4, nil}, {5, 4}, {6, nil}, {7, nil}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id, r.id
FROM (SELECT * FROM adaptive_src LIMIT 100) l
LEFT JOIN adaptive_dst r ON r.v = l.v AND r.id < 0
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, nil}, {4, nil}, {5, nil}, {6, nil}, {7, nil}},
				},
			},
		},
		{
			Name: "adaptive bytes",
			SetUpScript: []string{
				`CREATE TABLE adaptive_src (id INT PRIMARY KEY, v BYTEA)`,
				`CREATE TABLE adaptive_dst (id INT PRIMARY KEY, v BYTEA, bucket INT)`,
				`CREATE INDEX adaptive_v ON adaptive_dst(v,bucket)`,
				`INSERT INTO adaptive_src VALUES (1,decode('', 'hex')),(2,decode('610062','hex')),(3,decode('61006200','hex')),(4,decode('ff00','hex')),(5,decode(repeat('ab',20000),'hex')),(6,NULL),(7,decode('ff','hex'))`,
				`INSERT INTO adaptive_dst VALUES (1,decode('', 'hex'),0),(2,decode('610062','hex'),1),(3,decode('ff00','hex'),0),(4,decode(repeat('ab',20000),'hex'),1),(5,NULL,0),(6,decode('610062','hex'),1)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id, r.id
FROM (SELECT * FROM adaptive_src LIMIT 100) l
JOIN adaptive_dst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, 1}, {2, 2}, {2, 6}, {4, 3}, {5, 4}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id, r.id
FROM (SELECT * FROM adaptive_src LIMIT 100) l
LEFT JOIN adaptive_dst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, 1}, {2, 2}, {2, 6}, {3, nil}, {4, 3}, {5, 4}, {6, nil}, {7, nil}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id, r.id
FROM (SELECT * FROM adaptive_src LIMIT 100) l
LEFT JOIN adaptive_dst r ON r.v = l.v AND r.bucket = 1
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, 2}, {2, 6}, {3, nil}, {4, nil}, {5, 4}, {6, nil}, {7, nil}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id, r.id
FROM (SELECT * FROM adaptive_src LIMIT 100) l
LEFT JOIN adaptive_dst r ON r.v = l.v AND r.id < 0
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, nil}, {4, nil}, {5, nil}, {6, nil}, {7, nil}},
				},
			},
		},
		{
			Name: "adaptive jsonb",
			SetUpScript: []string{
				`CREATE TABLE adaptive_src (id INT PRIMARY KEY, v JSONB)`,
				`CREATE TABLE adaptive_dst (id INT PRIMARY KEY, v JSONB, bucket INT)`,
				`CREATE INDEX adaptive_v ON adaptive_dst(v,bucket)`,
				`INSERT INTO adaptive_src VALUES (1,'{}'),(2,'{"a":1,"b":2}'),(3,'{"a":1}'),(4,'[1,2]'),(5,to_jsonb(repeat('x',20000))),(6,NULL),(7,'null')`,
				`INSERT INTO adaptive_dst VALUES (1,'{}',0),(2,'{"b":2,"a":1}',1),(3,'[1,2]',0),(4,to_jsonb(repeat('x',20000)),1),(5,NULL,0),(6,'{"a":1,"b":2}',1)`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id, r.id
FROM (SELECT * FROM adaptive_src LIMIT 100) l
JOIN adaptive_dst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, 1}, {2, 2}, {2, 6}, {4, 3}, {5, 4}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id, r.id
FROM (SELECT * FROM adaptive_src LIMIT 100) l
LEFT JOIN adaptive_dst r ON r.v = l.v
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, 1}, {2, 2}, {2, 6}, {3, nil}, {4, 3}, {5, 4}, {6, nil}, {7, nil}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id, r.id
FROM (SELECT * FROM adaptive_src LIMIT 100) l
LEFT JOIN adaptive_dst r ON r.v = l.v AND r.bucket = 1
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, 2}, {2, 6}, {3, nil}, {4, nil}, {5, 4}, {6, nil}, {7, nil}},
				},
				{
					Query: `SELECT /*+ LOOKUP_JOIN(l,r) JOIN_ORDER(l,r) */ HINT l.id, r.id
FROM (SELECT * FROM adaptive_src LIMIT 100) l
LEFT JOIN adaptive_dst r ON r.v = l.v AND r.id < 0
ORDER BY l.id, r.id`,
					Expected: []sql.Row{{1, nil}, {2, nil}, {3, nil}, {4, nil}, {5, nil}, {6, nil}, {7, nil}},
				},
			},
		},
	})
}
