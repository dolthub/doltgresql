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

package enginetest

import (
	"fmt"
	denginetest "github.com/dolthub/dolt/go/libraries/doltcore/sqle/enginetest"
	"github.com/dolthub/go-mysql-server/enginetest"
	"github.com/dolthub/go-mysql-server/enginetest/queries"
	"github.com/dolthub/go-mysql-server/sql"
	"strings"
	"testing"
	"time"
)

func nativeViewAsOf() queries.ScriptTest {
	s := denginetest.ViewsWithAsOfScriptTest
	s.SetUpScript = append([]string(nil), s.SetUpScript...)
	s.Assertions = append([]queries.ScriptTestAssertion(nil), s.Assertions...)
	var setup []string
	for _, q := range s.SetUpScript {
		if strings.HasPrefix(q, "CREATE VIEW") {
			continue
		}
		if strings.Contains(q, "'cm5'") {
			q = "CALL dolt_commit('--allow-empty','-am','cm5')"
		}
		setup = append(setup, q)
		if strings.HasPrefix(q, "CREATE TABLE t2") {
			// PostgreSQL requires explicit UNION coercion. Version the view when its
			// dependencies first exist, so historical reads resolve that definition.
			setup = append(setup, "CREATE VIEW v1 AS SELECT pk, c0::text AS c0 FROM t1 UNION SELECT pk, vc::text FROM t2")
		}
	}
	s.SetUpScript = setup
	for i := range s.Assertions {
		a := &s.Assertions[i]
		if a.ExpectedErrStr != "" {
			a.ExpectedErrStr = "does not exist in revision"
		}
		if a.Query == "select * from v1 as of HEAD" {
			a.Query = "select * from v1 as of 'HEAD'"
		}
		if strings.Contains(a.Query, "HEAD.ASDF") {
			a.Query = "select * from v1 as of 'HEAD.ASDF'"
			a.ExpectedErrStr = "unknown revision"
		}
	}
	return s
}

func nativeDescribeAsOf() queries.ScriptTest {
	s := denginetest.DescribeTableAsOfScriptTest
	s.SetUpScript = nil
	for _, q := range denginetest.DescribeTableAsOfScriptTest.SetUpScript {
		if strings.HasPrefix(q, "set @") {
			continue
		}
		if strings.HasPrefix(q, "call dolt_commit_hash_out(") {
			q = "CALL dolt_commit(" + q[strings.Index(q, ",")+1:]
		}
		s.SetUpScript = append(s.SetUpScript, q)
	}
	s.Assertions = []queries.ScriptTestAssertion{{Query: "SELECT * FROM a AS OF 'HEAD~3'", ExpectedErr: sql.ErrTableNotFound}}
	for _, r := range []struct {
		revision string
		rows     []sql.Row
	}{
		{revision: "HEAD~2", rows: []sql.Row{{"pk", "integer", true}, {"c1", "integer", false}}},
		{revision: "HEAD~1", rows: []sql.Row{{"pk", "integer", true}, {"c1", "integer", false}, {"c2", "character varying(20)", false}}},
		{revision: "HEAD", rows: []sql.Row{{"pk", "integer", true}, {"c2", "character varying(20)", false}}},
	} {
		s.Assertions = append(s.Assertions, queries.ScriptTestAssertion{
			Query:    fmt.Sprintf("SELECT a.attname,format_type(a.atttypid,a.atttypmod),a.attnotnull FROM pg_attribute AS OF '%s' a JOIN pg_class AS OF '%s' c ON c.oid=a.attrelid WHERE c.relname='a' AND a.attnum>0 AND NOT a.attisdropped ORDER BY a.attnum", r.revision, r.revision),
			Expected: r.rows,
		})
	}
	return s
}

// Native functions accept ordinary PostgreSQL expressions, have fixed typed
// signatures, and use the authenticated PostgreSQL identity by default.
func nativeLogScript(s queries.ScriptTest) queries.ScriptTest {
	s.SetUpScript = append([]string(nil), s.SetUpScript...)
	s.Assertions = append([]queries.ScriptTestAssertion(nil), s.Assertions...)
	for i := range s.Assertions {
		a := &s.Assertions[i]
		a.Query = strings.ReplaceAll(a.Query, "having cnt > 1", "having count(*) > 1")
		a.Query = strings.ReplaceAll(a.Query, " join dolt_diff(", " CROSS JOIN dolt_diff(")
		a.Query = strings.ReplaceAll(a.Query, " join lateral ", " CROSS JOIN LATERAL ")
		a.Query = strings.ReplaceAll(a.Query, "hashof(", "dolt_hashof(")
		a.Query = strings.ReplaceAll(a.Query, "SUBSTRING_INDEX(parents, ', ', 1)", "split_part(parents, ', ', 1)")
		a.Query = strings.ReplaceAll(a.Query, "SUBSTRING_INDEX(parents, ', ', -1)", "split_part(parents, ', ', -1)")
		if a.ExpectedErr != nil || a.ExpectedErrStr != "" {
			a.ExpectedErr = nil
			switch {
			case strings.Contains(a.Query, "null"):
				a.ExpectedErrStr = "SQLSTATE 22004"
			case strings.Contains(a.Query, "123"):
				a.ExpectedErrStr = "SQLSTATE 42883"
			case strings.Contains(a.Query, "fake"):
				a.ExpectedErrStr = "unknown revision: fake-branch"
			case strings.Contains(a.Query, "dolt_hashof('main')"):
				a.Query = "SELECT count(*) FROM dolt_log(dolt_hashof('main'))"
				a.ExpectedErrStr = ""
				a.Expected = []sql.Row{{3}}
			case strings.Contains(a.Query, "LOWER(@Commit2)"):
				a.Query = "SELECT count(*) FROM dolt_log(@Commit1, LOWER(@Commit2))"
				a.ExpectedErrStr = ""
				a.Expected = []sql.Row{{3}}
			default:
				a.ExpectedErrStr = "unsupported version-control option or operation"
			}
		}
		if strings.HasPrefix(a.Query, "USE mydb/") {
			// Explicit revision-qualified catalogs exercise the same detached metadata
			// without MySQL's session USE statement, which PostgreSQL does not have.
			revision := strings.TrimSuffix(strings.TrimPrefix(a.Query, "USE mydb/"), ";")
			a.Query = fmt.Sprintf("SELECT count(*) FROM pg_class AS OF '%s' WHERE relname='test_table'", revision)
			a.Expected = []sql.Row{{1}}
		}
		if a.Query == "describe dolt_log;" {
			a.Query = "SELECT attname,format_type(atttypid,atttypmod) FROM pg_attribute WHERE attrelid='pg_catalog.dolt_log'::regclass AND attnum>0 ORDER BY attnum"
			a.Expected = []sql.Row{{"commit_hash", "text"}, {"committer", "text"}, {"email", "text"}, {"date", "timestamp without time zone"}, {"message", "text"}, {"commit_order", "bigint"}, {"parents", "text"}, {"refs", "text"}, {"signature", "text"}, {"author", "text"}, {"author_email", "text"}, {"author_date", "timestamp without time zone"}}
		}
		// MySQL's warning number is not a PostgreSQL protocol field. The value
		// assertion still verifies the documented auto-to-short fallback.
		a.ExpectedWarning = 0
		a.ExpectedWarningsCount = 0
		a.ExpectedWarningMessageSubstring = ""
		rows := make([]sql.Row, len(a.Expected))
		for r, row := range a.Expected {
			rows[r] = append(sql.Row(nil), row...)
			for c, v := range rows[r] {
				if v == "billy bob" {
					rows[r][c] = "postgres"
				}
				if v == "bigbillieb@fake.horse" {
					rows[r][c] = "postgres@localhost"
				}
				if v == "root" {
					rows[r][c] = "postgres"
				}
				if v == "root@localhost" {
					rows[r][c] = "postgres@127.0.0.1"
				}
			}
		}
		a.Expected = rows
	}
	return s
}

func TestNativeVersionFixtureExpressions(t *testing.T) {
	if !nativePostgresEngine() {
		t.Skip("PostgreSQL fixture translation")
	}
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestScript(t, h, queries.ScriptTest{
		Name: "computed user variables and folded table renames",
		SetUpScript: []string{
			"CREATE TABLE renamedTable (id int PRIMARY KEY)",
			"INSERT INTO renamedTable VALUES (1)",
			"RENAME TABLE renamedTable TO newRenamedTable",
			"SET @sum=1+2", "SET @copy=@sum",
			"CREATE TABLE fk_parent (id int PRIMARY KEY)",
			"CREATE TABLE fk_child (id int, CONSTRAINT named_parent FOREIGN KEY (id) REFERENCES fk_parent(id))",
			"ALTER TABLE fk_child DROP FOREIGN KEY named_parent",
			"ALTER TABLE fk_child ADD CONSTRAINT renamed_parent FOREIGN KEY (id) REFERENCES fk_parent(id)",
		},
		Assertions: []queries.ScriptTestAssertion{
			{Query: "SELECT id FROM newRenamedTable", Expected: []sql.Row{{1}}},
			{Query: "SELECT @sum,@copy", Expected: []sql.Row{{"3", "3"}}},
			{Query: "SELECT conname::text FROM pg_constraint WHERE conrelid='fk_child'::regclass AND contype='f'", Expected: []sql.Row{{"renamed_parent"}}},
			{Query: "SELECT SLEEP(0)", Expected: []sql.Row{{0}}},
			{Query: "SELECT a.id,b.id FROM newRenamedTable a JOIN newRenamedTable b", Expected: []sql.Row{{1, 1}}},
			{Query: "SELECT '2020-01-02 03:04:05+00'::timestamptz", Expected: []sql.Row{{time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)}}},
		},
	})
}

// Canonical object names include their namespace. Row/cell counts and change
// flags are deliberately retained from the shared fixtures.
func nativeDiffMetadataScript(s queries.ScriptTest, nameColumns int) queries.ScriptTest {
	if s.Name == "raising the autoincrement value should be seen as a schema change" {
		// PostgreSQL sequence allocation is deliberately shared, unversioned
		// state. Advancing a counter alone must not manufacture a schema diff.
		return queries.ScriptTest{
			Name: s.Name,
			SetUpScript: []string{
				"CREATE TABLE t(pk serial PRIMARY KEY) USING dolt_prolly_table",
				"INSERT INTO t DEFAULT VALUES",
				"CALL dolt_commit('-Am','first allocation')",
				"INSERT INTO t DEFAULT VALUES",
				"DELETE FROM t WHERE pk=2",
				"CALL dolt_commit('--allow-empty','-am','second allocation')",
			},
			Assertions: []queries.ScriptTestAssertion{
				{Query: "SELECT * FROM dolt_diff_summary('HEAD~','HEAD')", Expected: []sql.Row{}},
				{Query: "SELECT currval(pg_get_serial_sequence('t','pk'))", Expected: []sql.Row{{2}}},
			},
		}
	}
	s.SetUpScript = append([]string(nil), s.SetUpScript...)
	for i, q := range s.SetUpScript {
		// These fixtures create an unnamed primary key and a named foreign key.
		if strings.EqualFold(q, "ALTER TABLE t drop primary key;") {
			q = "ALTER TABLE t DROP CONSTRAINT t_pkey"
		}
		s.SetUpScript[i] = q
	}
	s.Assertions = append([]queries.ScriptTestAssertion(nil), s.Assertions...)
	for i := range s.Assertions {
		a := &s.Assertions[i]
		// In this fixture Commit5 is HEAD; Commit3 is exactly HEAD~2.
		// The native AS OF binder requires a constant revision.
		a.Query = strings.ReplaceAll(a.Query, "as of @Commit3", "AS OF 'HEAD~2'")
		if a.ExpectedErr != nil || a.ExpectedErrStr != "" {
			primaryKeyChange := strings.Contains(a.ExpectedErrStr, "primary key set changed")
			missingTable := a.ExpectedErr == sql.ErrTableNotFound
			a.ExpectedErr = nil
			switch {
			case primaryKeyChange:
				a.ExpectedErrStr = "primary key set changed"
			case missingTable:
				a.ExpectedErrStr = "SQLSTATE 42P01"
			case strings.Contains(strings.ToLower(a.Query), "null"):
				// Native overloads are STRICT, so null arguments produce no rows.
				a.ExpectedErrStr = ""
				a.Expected = []sql.Row{}
			case strings.Contains(a.Query, "fake"):
				a.ExpectedErrStr = "unknown revision: fake-branch"
			case strings.Contains(a.Query, "hashof('main')"):
				a.Query = strings.Replace(a.Query, "SELECT *", "SELECT count(*)", 1)
				a.ExpectedErrStr = ""
				a.Expected = []sql.Row{{0}}
			case strings.Contains(a.Query, "LOWER('T')"):
				a.Query = strings.Replace(a.Query, "SELECT *", "SELECT count(*)", 1)
				a.ExpectedErrStr = ""
				a.Expected = []sql.Row{{1}}
			case strings.Contains(a.Query, "('t');"):
				a.ExpectedErrStr = "requires two revisions or a from..to revision range"
			default:
				a.ExpectedErrStr = "SQLSTATE 42883"
			}
		}
		// PostgreSQL warnings carry SQLSTATE, not MySQL warning numbers.
		a.ExpectedWarning = 0
		a.ExpectedWarningsCount = 0
		a.ExpectedWarningMessageSubstring = ""
		if strings.Contains(strings.ToLower(a.Query), "select *") {
			rows := make([]sql.Row, len(a.Expected))
			for r, row := range a.Expected {
				rows[r] = append(sql.Row(nil), row...)
				for c := 0; c < nameColumns && c < len(row); c++ {
					if name, ok := row[c].(string); ok && name != "" && !strings.Contains(name, ".") {
						rows[r][c] = "public." + strings.ToLower(name)
					}
				}
			}
			a.Expected = rows
		}
	}
	return s
}
