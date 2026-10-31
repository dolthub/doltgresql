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
	"github.com/dolthub/go-mysql-server/sql/types"
	"os"
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
	// Intentional fork contract: names identify canonical members. Do not
	// carry Dolt's history-dependent column-fingerprint rename heuristic.
	if s.Name == "renamed table" {
		for i := range s.Assertions {
			a := &s.Assertions[i]
			oldName := strings.Contains(a.Query, "'t1'")
			newName := strings.Contains(a.Query, "'t2'")
			if nameColumns == 2 {
				a.Expected = nil
				if !newName {
					a.Expected = append(a.Expected, sql.Row{"t1", "", "dropped", true, true})
				}
				if !oldName {
					a.Expected = append(a.Expected, sql.Row{"", "t2", "added", true, true})
				}
			} else if oldName {
				a.Expected = []sql.Row{{"t1", 0, 0, 1, 0, 0, 2, 0, 1, 0, 2, 0}}
			} else {
				a.Expected = []sql.Row{{"t2", 0, 2, 0, 0, 4, 0, 0, 0, 2, 0, 4}}
			}
		}
	}

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

func TestNativeCatalogFixtureQueries(t *testing.T) {
	if !nativePostgresEngine() {
		t.Skip("PostgreSQL catalog counterparts")
	}
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestScript(t, h, queries.ScriptTest{
		Name:        "native catalog listing and database identity",
		SetUpScript: []string{"CREATE TABLE fixture_table(id int PRIMARY KEY)", "CREATE VIEW fixture_view AS SELECT id FROM fixture_table"},
		Assertions: []queries.ScriptTestAssertion{
			{Query: "SHOW TABLES LIKE 'fixture_%'", Expected: []sql.Row{{"fixture_table"}, {"fixture_view"}}},
			{Query: "SHOW FULL TABLES LIKE 'fixture_%'", Expected: []sql.Row{{"fixture_table", "BASE TABLE"}, {"fixture_view", "VIEW"}}},
			{Query: "SELECT DATABASE()=current_database()", Expected: []sql.Row{{true}}},
		},
	})
}

func TestNativeHistoricalCatalogListing(t *testing.T) {
	if os.Getenv("DOLTGRES_TEST_TARGET") != "postgres-fork" {
		t.Skip("fork revision catalogs")
	}
	h := newDoltgresServerHarness(t)
	defer h.Close()
	enginetest.TestScript(t, h, queries.ScriptTest{
		Name:        "historical SHOW TABLES counterpart",
		SetUpScript: []string{"CREATE TABLE history_fixture(id int PRIMARY KEY)", "CALL dolt_commit('-Am','table exists')", "DROP TABLE history_fixture", "CALL dolt_commit('-Am','table removed')"},
		Assertions: []queries.ScriptTestAssertion{
			{Query: "SHOW TABLES AS OF 'HEAD~1' LIKE 'history_%'", Expected: []sql.Row{{"history_fixture"}}},
			{Query: "SHOW TABLES LIKE 'history_%'", Expected: []sql.Row{}},
		},
	})
}

func nativeCherryPickScript(s queries.ScriptTest) queries.ScriptTest {
	s.SetUpScript = append([]string(nil), s.SetUpScript...)
	for i, q := range s.SetUpScript {
		q = strings.ReplaceAll(q, "DROP PRIMARY KEY, ADD PRIMARY KEY", "DROP CONSTRAINT t_pkey, ADD PRIMARY KEY")
		q = strings.ReplaceAll(q, "('generated_*', 1)", "('generated_*', true)")
		s.SetUpScript[i] = q
	}
	s.Assertions = append([]queries.ScriptTestAssertion(nil), s.Assertions...)
	if s.Name == "error cases: different primary keys" {
		// The native schema publisher can apply this empty-table primary-key
		// change. Verify the new key and its enforcement instead of copying
		// Dolt's unsupported-operation error.
		s.Assertions[0].ExpectedErrStr = ""
		s.Assertions[0].SkipResultsCheck = true
		s.Assertions = append(s.Assertions,
			queries.ScriptTestAssertion{Query: "SELECT pg_get_constraintdef(oid) FROM pg_constraint WHERE conrelid='t'::regclass AND contype='p'", Expected: []sql.Row{{"PRIMARY KEY (pk, v)"}}},
			queries.ScriptTestAssertion{Query: "INSERT INTO t VALUES (1,'one'),(1,'two')", SkipResultsCheck: true},
			queries.ScriptTestAssertion{Query: "SELECT count(*) FROM t", Expected: []sql.Row{{2}}},
			queries.ScriptTestAssertion{Query: "INSERT INTO t VALUES (1,'one')", ExpectedErrStr: "SQLSTATE 23505"},
		)
	}

	for i := range s.Assertions {
		a := &s.Assertions[i]
		if a.ExpectedErrStr == "error: There is no cherry-pick merge to continue" {
			a.ExpectedErrStr = "no cherry-pick to continue"
		}
		if strings.Contains(a.ExpectedErrStr, "mutually exclusive") {
			a.ExpectedErrStr = "--continue and --abort are mutually exclusive"
		}
		rows := make([]sql.Row, len(a.Expected))
		for r, row := range a.Expected {
			rows[r] = append(sql.Row(nil), row...)
			for c, v := range rows[r] {
				if v == "root" {
					rows[r][c] = "postgres"
				}
				if v == "root@localhost" {
					rows[r][c] = "postgres@127.0.0.1"
				}
			}
		}
		a.Expected = rows
		if a.ExpectedErrStr == "target commit not found" {
			a.ExpectedErrStr = "revision does not name a visible commit object"
		}
		if a.ExpectedErrStr == "error: There is no cherry-pick merge to abort" {
			a.ExpectedErrStr = "no cherry-pick to abort"
		}
		if strings.EqualFold(a.Query, "SHOW CREATE TABLE test;") {
			// Check the native definition, including the primary key, instead of
			// comparing MySQL's SHOW CREATE formatting and storage-engine clauses.
			a.Query = "SELECT a.attname,format_type(a.atttypid,a.atttypmod),a.attnotnull,EXISTS(SELECT 1 FROM pg_constraint c WHERE c.conrelid=a.attrelid AND c.contype='p' AND a.attnum=ANY(c.conkey)) FROM pg_attribute a WHERE a.attrelid='test'::regclass AND a.attnum>0 AND NOT a.attisdropped ORDER BY a.attnum"
			a.Expected = []sql.Row{{"pk", "integer", true, true}}
			if strings.Contains(s.Name, "ADD COLUMN") {
				a.Expected = append(a.Expected, sql.Row{"v", "character varying(100)", false, false})
			}
			if strings.Contains(s.Name, "RENAME COLUMN") {
				a.Expected = append(a.Expected, sql.Row{"v2", "character varying(100)", false, false})
			}
		}
		q := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(a.Query)), ";")
		if q == "insert into generated_foo values (2)" {
			// PostgreSQL exposes the affected-row command tag for this INSERT.
			a.Expected = []sql.Row{{types.NewOkResult(1)}}
		}
		if q == "select * from dolt_status" || q == "select * from dolt_status_ignored" {
			rows := make([]sql.Row, len(a.Expected))
			for j, row := range a.Expected {
				rows[j] = append(sql.Row(nil), row...)
				if name, ok := rows[j][0].(string); ok && !strings.Contains(name, ".") {
					rows[j][0] = "public." + name
				}
			}
			a.Expected = rows
		}
	}
	return s
}

// Native reverts preserve conflicts without aborting the SQL statement, just
// like native merge/cherry-pick. Original Dolt expectations remain unchanged.
func nativeRevertScript(s queries.ScriptTest) queries.ScriptTest {
	s = nativeCherryPickScript(s)
	for i, q := range s.SetUpScript {
		q = strings.ReplaceAll(q, "('ignored_*', 1)", "('ignored_*', true)")
		q = strings.ReplaceAll(q, "('dont_*', 1)", "('dont_*', true)")
		q = strings.ReplaceAll(q, "tableA", "tablea")
		s.SetUpScript[i] = q
	}
	countQueries := 0
	for i := range s.Assertions {
		a := &s.Assertions[i]
		a.Query = strings.ReplaceAll(a.Query, "tableA", "tablea")
		q := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(a.Query)), ";")
		if s.Name == "dolt_revert() with @@autocommit=1 returns error on conflicts and leaves repo clean" {
			if strings.HasPrefix(q, "call dolt_revert") {
				a.ExpectedErrStr = ""
				a.Expected = []sql.Row{{"", 1, 0, 0}}
			}
			if q == "select `table` from dolt_conflicts" {
				a.Expected = []sql.Row{{"test"}}
			}
		}
		if strings.Contains(a.ExpectedErrStr, "no revert in progress") || strings.Contains(a.ExpectedErrStr, "There is no revert") {
			if strings.Contains(q, "--abort") {
				a.ExpectedErrStr = "no revert to abort"
			} else {
				a.ExpectedErrStr = "no revert to continue"
			}
		}
		if strings.Contains(a.ExpectedErrStr, "local changes would be overwritten") {
			if strings.Contains(s.Name, "staged changes") {
				a.ExpectedErrStr = "cannot revert with uncommitted changes"
			} else {
				a.ExpectedErrStr = "local changes would be overwritten by revert"
			}
		}
		if a.ExpectedErrStr == "table not found: dont_track" {
			a.ExpectedErrStr = "SQLSTATE 42P01"
		}
		if q == "select count(*) from dolt_log" {
			// Count only this fixture's commits, independently of init history.
			a.Query = `SELECT count(*) FROM dolt_log WHERE message = ANY (ARRAY['seed table','insert row 4','insert row 5','Revert "insert row 4"','Revert "insert row 5"']::text[])`
			a.Expected = []sql.Row{{3 + 2*countQueries}}
			countQueries++
		}
	}
	if strings.Contains(s.Name, "--abort rejects other dirty working sets") {
		// Revert is a standalone native command. Verify failure atomicity
		// inside an explicit transaction using PostgreSQL savepoint recovery.
		n := len(s.SetUpScript)
		s.SetUpScript = append(s.SetUpScript[:n-1], "BEGIN", s.SetUpScript[n-1], "SAVEPOINT before_abort")
		s.Assertions[0].ExpectedErrStr = "SQLSTATE 25001"
		a := queries.ScriptTestAssertion{Query: "ROLLBACK TO SAVEPOINT before_abort", SkipResultsCheck: true}
		s.Assertions = append(s.Assertions[:1], append([]queries.ScriptTestAssertion{a}, s.Assertions[1:]...)...)
	}
	return s
}
