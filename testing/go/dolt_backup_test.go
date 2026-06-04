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
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/require"
)

// localBackupUrl returns a file:// URL for a subdirectory named |name| inside the test's
// auto-cleaned TempDir. Each call produces a unique parent directory, so two calls with the same
// |name| never collide. The directory is automatically removed when the test ends.
func localBackupUrl(t *testing.T, name string) string {
	t.Helper()
	return "file://" + filepath.ToSlash(filepath.Join(t.TempDir(), name))
}

// runBackupTest runs |script| inside a dedicated, freshly-started on-disk Doltgres server and
// tears the server down when the subtest completes.
//
// A new server is started for every call so that each backup/restore test begins with a clean
// database state. Without this isolation, objects created (or restored) in one test could leak
// into subsequent tests, producing false passes or false failures. The server runs against a
// temporary on-disk directory (via CreateServerLocalWithPort) because dolt_backup requires a real
// NBS chunk store — in-memory storage cannot be synced to a file:// backup URL.
//
// Callers must pre-compute any backup URLs with localBackupUrl(t, …) so the directories live
// inside the test's auto-cleaned TempDir rather than a persistent system path.
func runBackupTest(t *testing.T, script ScriptTest) {
	t.Helper()
	t.Run(script.Name, func(t *testing.T) {
		t.Helper()
		port, err := sql.GetEmptyPort()
		require.NoError(t, err)
		ctx, conn, controller := CreateServerLocalWithPort(t, "postgres", port)
		defer func() {
			conn.Close(ctx)
			controller.Stop()
			require.NoError(t, controller.WaitForStop())
		}()
		runScript(t, ctx, script, conn, true)
	})
}

func TestDoltBackup(t *testing.T) {
	backupUrl := localBackupUrl(t, "backup")
	runBackupTest(t, ScriptTest{
		Name: "sync-url restore preserves data and commit history",
		SetUpScript: []string{
			"create table items (id int primary key, label text not null);",
			"insert into items values (1, 'apple'), (2, 'banana');",
			"select dolt_commit('-Am', 'first commit: add apple and banana');",
			"insert into items values (3, 'cherry');",
			"select dolt_commit('-Am', 'second commit: add cherry');",
			fmt.Sprintf("select dolt_backup('sync-url', '%s');", backupUrl),
			fmt.Sprintf("select dolt_backup('restore', '%s', 'restored_syncurl');", backupUrl),
			"USE restored_syncurl",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select id, label from items order by id;",
				Expected: []sql.Row{{1, "apple"}, {2, "banana"}, {3, "cherry"}},
			},
			{
				// 4 commits: Initialize + CREATE DATABASE + 2 user commits.
				Query:    "select count(*) from dolt.commits;",
				Expected: []sql.Row{{int64(4)}},
			},
			{
				Query:    "select message from dolt.commits order by date;",
				Expected: []sql.Row{{"Initialize data repository"}, {"CREATE DATABASE"}, {"first commit: add apple and banana"}, {"second commit: add cherry"}},
			},
			{
				// NOT NULL constraint on label must survive restore.
				Query:    "select column_name, is_nullable from information_schema.columns where table_name = 'items' and table_schema = 'public' order by column_name;",
				Expected: []sql.Row{{"id", "NO"}, {"label", "NO"}},
			},
		},
	})

	backupUrl = localBackupUrl(t, "backup")
	runBackupTest(t, ScriptTest{
		Name: "named backup sync restore preserves data and commit history",
		SetUpScript: []string{
			"create table orders (id int primary key, item text, qty int);",
			"insert into orders values (1, 'widget', 10);",
			"select dolt_add('.');",
			"select dolt_commit('-m', 'initial order');",
			"insert into orders values (2, 'gadget', 5);",
			"select dolt_add('.');",
			"select dolt_commit('-m', 'second order');",
			fmt.Sprintf("select dolt_backup('add', 'named_bak', '%s');", backupUrl),
			"select dolt_backup('sync', 'named_bak');",
			fmt.Sprintf("select dolt_backup('restore', '%s', 'restored_named');", backupUrl),
			"USE restored_named",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select id, item, qty from orders order by id;",
				Expected: []sql.Row{{1, "widget", 10}, {2, "gadget", 5}},
			},
			{
				// 4 commits: Initialize + CREATE DATABASE + 2 user commits.
				Query:    "select count(*) from dolt.commits;",
				Expected: []sql.Row{{int64(4)}},
			},
			{
				Query:    "select message from dolt.commits order by date;",
				Expected: []sql.Row{{"Initialize data repository"}, {"CREATE DATABASE"}, {"initial order"}, {"second order"}},
			},
		},
	})

	urlA := localBackupUrl(t, "backup_a")
	urlB := localBackupUrl(t, "backup_b")
	runBackupTest(t, ScriptTest{
		Name: "dolt.dolt_backups table reflects add and remove",
		SetUpScript: []string{
			fmt.Sprintf("select dolt_backup('add', 'backup_a', '%s');", urlA),
			fmt.Sprintf("select dolt_backup('add', 'backup_b', '%s');", urlB),
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select name from dolt.dolt_backups order by name;",
				Expected: []sql.Row{{"backup_a"}, {"backup_b"}},
			},
			{
				Query:    "select dolt_backup('remove', 'backup_a');",
				Expected: []sql.Row{{"{0}"}},
			},
			{
				Query:    "select name from dolt.dolt_backups order by name;",
				Expected: []sql.Row{{"backup_b"}},
			},
			{
				Query:    "select dolt_backup('remove', 'backup_b');",
				Expected: []sql.Row{{"{0}"}},
			},
			{
				Query:    "select count(*) from dolt.dolt_backups;",
				Expected: []sql.Row{{int64(0)}},
			},
		},
	})

	backupUrl = localBackupUrl(t, "backup")
	runBackupTest(t, ScriptTest{
		Name: "restore --force overwrites an existing database and data is correct",
		SetUpScript: []string{
			"create table t (id int primary key, v text);",
			"insert into t values (1, 'original');",
			"select dolt_add('.');",
			"select dolt_commit('-m', 'seed');",
			fmt.Sprintf("select dolt_backup('sync-url', '%s');", backupUrl),
			// First restore creates db_to_overwrite.
			fmt.Sprintf("select dolt_backup('restore', '%s', 'db_to_overwrite');", backupUrl),
		},
		Assertions: []ScriptTestAssertion{
			{
				// Without --force a second restore to the same name must error.
				Query:       fmt.Sprintf("select dolt_backup('restore', '%s', 'db_to_overwrite');", backupUrl),
				ExpectedErr: "database 'db_to_overwrite' already exists, use '--force' to overwrite",
			},
			{
				Query:    fmt.Sprintf("select dolt_backup('restore', '--force', '%s', 'db_to_overwrite');", backupUrl),
				Expected: []sql.Row{{"{0}"}},
			},
			{
				Query: "USE db_to_overwrite;",
			},
			{
				Query:    "select id, v from t order by id;",
				Expected: []sql.Row{{1, "original"}},
			},
			{
				// Initialize + CREATE DATABASE + seed.
				Query:    "select count(*) from dolt.commits;",
				Expected: []sql.Row{{int64(3)}},
			},
			{
				Query: "USE postgres;",
			},
			{
				Query:    "drop database db_to_overwrite;",
				Expected: []sql.Row{},
			},
		},
	})

	backupUrl = localBackupUrl(t, "backup")
	runBackupTest(t, ScriptTest{
		Name: "multi-branch backup and restore preserves all branches",
		SetUpScript: []string{
			"create table products (id int primary key, name text);",
			"insert into products values (1, 'alpha');",
			"select dolt_add('.');",
			"select dolt_commit('-m', 'add alpha on main');",
			"select dolt_branch('feature');",
			"select dolt_checkout('feature');",
			"insert into products values (2, 'beta');",
			"select dolt_commit('-Am', 'add beta on feature');",
			"select dolt_checkout('main');",
			fmt.Sprintf("select dolt_backup('sync-url', '%s');", backupUrl),
			fmt.Sprintf("select dolt_backup('restore', '%s', 'restored_branches');", backupUrl),
			"USE restored_branches",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select id, name from products order by id;",
				Expected: []sql.Row{{1, "alpha"}},
			},
			{
				Query:    "select name from dolt.branches order by name;",
				Expected: []sql.Row{{"feature"}, {"main"}},
			},
			{
				Query:            "select dolt_checkout('feature');",
				SkipResultsCheck: true,
			},
			{
				Query:    "select id, name from products order by id;",
				Expected: []sql.Row{{1, "alpha"}, {2, "beta"}},
			},
			{
				// Feature branch must have its own commit on top of main.
				Query:    "select message from dolt.commits order by date desc limit 1;",
				Expected: []sql.Row{{"add beta on feature"}},
			},
		},
	})

	backupUrl = localBackupUrl(t, "backup")
	runBackupTest(t, ScriptTest{
		Name: "working set state is preserved in backup",
		SetUpScript: []string{
			"create table logs (id int primary key, msg text);",
			"insert into logs values (1, 'committed');",
			"select dolt_add('.');",
			"select dolt_commit('-m', 'first commit');",
			// Row 2: staged but not committed.
			"insert into logs values (2, 'staged only');",
			"select dolt_add('.');",
			// Row 3: unstaged.
			"insert into logs values (3, 'unstaged');",
			// sync-url flushes the working set before syncing, so all rows are captured.
			fmt.Sprintf("select dolt_backup('sync-url', '%s');", backupUrl),
			fmt.Sprintf("select dolt_backup('restore', '%s', 'restored_workingset');", backupUrl),
			"USE restored_workingset",
		},
		Assertions: []ScriptTestAssertion{
			{
				// All three rows are visible: staged and unstaged changes are flushed to the
				// working set before backup and are restored intact.
				Query:    "select id, msg from logs order by id;",
				Expected: []sql.Row{{1, "committed"}, {2, "staged only"}, {3, "unstaged"}},
			},
			{
				// Only the first user commit; rows 2 and 3 are in the working set only.
				Query:    "select count(*) from dolt.commits;",
				Expected: []sql.Row{{int64(3)}},
			},
		},
	})

	backupUrl = localBackupUrl(t, "backup")
	runBackupTest(t, ScriptTest{
		Name: "incremental sync captures new commits added after the first sync",
		SetUpScript: []string{
			fmt.Sprintf("select dolt_backup('add', 'incr_bak', '%s');", backupUrl),
			"create table events (id int primary key, name text);",
			"insert into events values (1, 'first');",
			"select dolt_add('.');",
			"select dolt_commit('-m', 'first event');",
			"select dolt_backup('sync', 'incr_bak');",
			"insert into events values (2, 'second');",
			"select dolt_add('.');",
			"select dolt_commit('-m', 'second event');",
			// Second (incremental) sync must include the new commit.
			"select dolt_backup('sync', 'incr_bak');",
			fmt.Sprintf("select dolt_backup('restore', '%s', 'restored_incr');", backupUrl),
			"select dolt_backup('remove', 'incr_bak');",
			"USE restored_incr",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select id, name from events order by id;",
				Expected: []sql.Row{{1, "first"}, {2, "second"}},
			},
			{
				// Initialize + CREATE DATABASE + 2 user commits.
				Query:    "select count(*) from dolt.commits;",
				Expected: []sql.Row{{int64(4)}},
			},
		},
	})

	runBackupTest(t, ScriptTest{
		Name: "sync nonexistent named backup returns error",
		Assertions: []ScriptTestAssertion{
			{
				Query:       "select dolt_backup('sync', 'does_not_exist');",
				ExpectedErr: "backup 'does_not_exist' not found",
			},
		},
	})

	runBackupTest(t, ScriptTest{
		Name: "remove nonexistent backup returns error",
		Assertions: []ScriptTestAssertion{
			{
				Query:       "select dolt_backup('remove', 'no_such_backup');",
				ExpectedErr: "backup 'no_such_backup' not found",
			},
		},
	})

	runBackupTest(t, ScriptTest{
		Name: "restore from nonexistent URL returns error",
		Assertions: []ScriptTestAssertion{
			{
				Query:       "select dolt_backup('restore', 'file:///nonexistent/doltgres/backup/path', 'new_db');",
				ExpectedErr: "ERROR",
			},
		},
	})

	urlA = localBackupUrl(t, "backup_a")
	urlB = localBackupUrl(t, "backup_b")
	runBackupTest(t, ScriptTest{
		Name: "adding a backup with a duplicate name returns error",
		SetUpScript: []string{
			fmt.Sprintf("select dolt_backup('add', 'mybackup', '%s');", urlA),
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:       fmt.Sprintf("select dolt_backup('add', 'mybackup', '%s');", urlB),
				ExpectedErr: "backup 'mybackup' already exists",
			},
			{
				Query:    "select dolt_backup('remove', 'mybackup');",
				Expected: []sql.Row{{"{0}"}},
			},
		},
	})

	runBackupTest(t, ScriptTest{
		Name: "unknown dolt_backup operation returns error",
		Assertions: []ScriptTestAssertion{
			{
				Query:       "select dolt_backup('typo', 'name', 'url');",
				ExpectedErr: "unrecognized dolt_backup parameter 'typo'",
			},
		},
	})

	runBackupTest(t, ScriptTest{
		Name: "dolt_backup with no args returns usage error",
		Assertions: []ScriptTestAssertion{
			{
				Query:       "select dolt_backup();",
				ExpectedErr: "use 'dolt_backups' table to list backups",
			},
		},
	})

	backupUrl = localBackupUrl(t, "backup")
	runBackupTest(t, ScriptTest{
		Name: "empty database backup and restore",
		SetUpScript: []string{
			fmt.Sprintf("select dolt_backup('sync-url', '%s');", backupUrl),
			fmt.Sprintf("select dolt_backup('restore', '%s', 'restored_empty');", backupUrl),
			"USE restored_empty",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select count(*) from information_schema.tables where table_schema = 'public';",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				// At minimum, Initialize + CREATE DATABASE commits must be present.
				Query:    "select count(*) >= 2 from dolt.commits;",
				Expected: []sql.Row{{"t"}},
			},
		},
	})

	backupUrl = localBackupUrl(t, "backup")
	runBackupTest(t, ScriptTest{
		Name: "standalone sequence state is preserved across backup and restore",
		SetUpScript: []string{
			"create sequence counter start 1 increment 5;",
			"select nextval('counter');", // advances to 1
			"select nextval('counter');", // advances to 6
			"select nextval('counter');", // advances to 11
			fmt.Sprintf("select dolt_backup('sync-url', '%s');", backupUrl),
			fmt.Sprintf("select dolt_backup('restore', '%s', 'restored_seq');", backupUrl),
			"USE restored_seq",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select sequence_name from information_schema.sequences where sequence_schema = 'public';",
				Expected: []sql.Row{{"counter"}},
			},
			{
				// Must continue from where it left off (after 11, next is 16), not reset to 1.
				Query:    "select nextval('counter');",
				Expected: []sql.Row{{int64(16)}},
			},
		},
	})

	backupUrl = localBackupUrl(t, "backup")
	runBackupTest(t, ScriptTest{
		Name: "serial column sequence state is preserved across backup and restore",
		SetUpScript: []string{
			"create table widgets (id serial primary key, name text);",
			"insert into widgets (name) values ('a'), ('b'), ('c');",
			"select dolt_add('.');",
			"select dolt_commit('-m', 'seed widgets');",
			fmt.Sprintf("select dolt_backup('sync-url', '%s');", backupUrl),
			fmt.Sprintf("select dolt_backup('restore', '%s', 'restored_serial');", backupUrl),
			"USE restored_serial",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select id, name from widgets order by id;",
				Expected: []sql.Row{{1, "a"}, {2, "b"}, {3, "c"}},
			},
			{
				// Auto-generated id must continue from 4, proving the sequence state was restored.
				Query:    "insert into widgets (name) values ('d') returning id;",
				Expected: []sql.Row{{4}},
			},
		},
	})

	backupUrl = localBackupUrl(t, "backup")
	runBackupTest(t, ScriptTest{
		Name: "views are preserved across backup and restore",
		SetUpScript: []string{
			"create table employees (id int primary key, dept text, salary int);",
			"insert into employees values (1, 'eng', 100000), (2, 'eng', 120000), (3, 'ops', 90000);",
			"create view eng_employees as select id, salary from employees where dept = 'eng';",
			"select dolt_add('.');",
			"select dolt_commit('-m', 'add employees and view');",
			fmt.Sprintf("select dolt_backup('sync-url', '%s');", backupUrl),
			fmt.Sprintf("select dolt_backup('restore', '%s', 'restored_views');", backupUrl),
			"USE restored_views",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select table_name from information_schema.views where table_schema = 'public';",
				Expected: []sql.Row{{"eng_employees"}},
			},
			{
				Query:    "select id, salary from eng_employees order by id;",
				Expected: []sql.Row{{1, 100000}, {2, 120000}},
			},
		},
	})

	backupUrl = localBackupUrl(t, "backup")
	runBackupTest(t, ScriptTest{
		Name: "foreign key constraints are preserved and enforced after restore",
		SetUpScript: []string{
			"create table categories (id int primary key, name text);",
			"create table items (id int primary key, cat_id int references categories(id), label text);",
			"insert into categories values (1, 'fruit'), (2, 'veg');",
			"insert into items values (1, 1, 'apple'), (2, 2, 'carrot');",
			"select dolt_add('.');",
			"select dolt_commit('-m', 'seed fk data');",
			fmt.Sprintf("select dolt_backup('sync-url', '%s');", backupUrl),
			fmt.Sprintf("select dolt_backup('restore', '%s', 'restored_fk');", backupUrl),
			"USE restored_fk",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select id, label from items order by id;",
				Expected: []sql.Row{{1, "apple"}, {2, "carrot"}},
			},
			{
				Query:       "insert into items values (3, 99, 'mystery');",
				ExpectedErr: "Foreign key",
			},
			{
				Query:    "select count(*) from information_schema.referential_constraints where constraint_schema = 'public';",
				Expected: []sql.Row{{int64(1)}},
			},
		},
	})

	backupUrl = localBackupUrl(t, "backup")
	runBackupTest(t, ScriptTest{
		Name: "secondary index is preserved and used after restore",
		SetUpScript: []string{
			"create table products (id int primary key, sku text not null, price int);",
			"create index idx_sku on products (sku);",
			"insert into products values (1, 'AAA', 10), (2, 'BBB', 20), (3, 'CCC', 30);",
			"select dolt_add('.');",
			"select dolt_commit('-m', 'seed products with index');",
			fmt.Sprintf("select dolt_backup('sync-url', '%s');", backupUrl),
			fmt.Sprintf("select dolt_backup('restore', '%s', 'restored_idx');", backupUrl),
			"USE restored_idx",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select indexname from pg_indexes where tablename = 'products' and indexname = 'idx_sku';",
				Expected: []sql.Row{{"idx_sku"}},
			},
			{
				Query:    "select id, sku, price from products order by id;",
				Expected: []sql.Row{{1, "AAA", 10}, {2, "BBB", 20}, {3, "CCC", 30}},
			},
			{
				// A lookup that the optimizer can satisfy via the index must return the correct row.
				Query:    "select id, price from products where sku = 'BBB';",
				Expected: []sql.Row{{2, 20}},
			},
			{
				// Uniqueness is NOT enforced by this non-unique index, but the index must still
				// accept new inserts with a duplicate sku value.
				Query:    "insert into products values (4, 'AAA', 99);",
				Expected: []sql.Row{},
			},
			{
				Query:    "select id from products where sku = 'AAA' order by id;",
				Expected: []sql.Row{{1}, {4}},
			},
		},
	})

	backupUrl = localBackupUrl(t, "backup")
	runBackupTest(t, ScriptTest{
		Name: "custom enum type is preserved across backup and restore",
		SetUpScript: []string{
			"create type my_mood as enum ('happy', 'ok', 'sad');",
			"create table user_states (id int primary key, name text, mood my_mood);",
			"insert into user_states values (1, 'alice', 'happy'), (2, 'bob', 'sad');",
			fmt.Sprintf("select dolt_backup('sync-url', '%s');", backupUrl),
			fmt.Sprintf("select dolt_backup('restore', '%s', 'restored_types');", backupUrl),
			"USE restored_types",
		},
		Assertions: []ScriptTestAssertion{
			{
				// pg_type is implemented and returns user-defined types.
				Query:    "select typname from pg_type where typname = 'my_mood';",
				Expected: []sql.Row{{"my_mood"}},
			},
			{
				Query:    "select 'ok'::my_mood;",
				Expected: []sql.Row{{"ok"}},
			},
			{
				Query:    "select id, name, mood::text from user_states order by id;",
				Expected: []sql.Row{{1, "alice", "happy"}, {2, "bob", "sad"}},
			},
		},
	})

	backupUrl = localBackupUrl(t, "backup")
	runBackupTest(t, ScriptTest{
		Name: "domain type is preserved across backup and restore",
		SetUpScript: []string{
			`create domain pos_int as integer check (value > 0);`,
			`create table measurements (id int primary key, val pos_int);`,
			`insert into measurements values (1, 42);`,
			fmt.Sprintf("select dolt_backup('sync-url', '%s');", backupUrl),
			fmt.Sprintf("select dolt_backup('restore', '%s', 'restored_domain');", backupUrl),
			"USE restored_domain",
		},
		Assertions: []ScriptTestAssertion{
			{
				// typtype = 'd' for domain types.
				Query:    "select typname, typtype from pg_type where typname = 'pos_int';",
				Expected: []sql.Row{{"pos_int", "d"}},
			},
			{
				Query:    "select val from measurements where id = 1;",
				Expected: []sql.Row{{42}},
			},
			{
				// Domain constraint must still be enforced after restore.
				Query:       "insert into measurements values (2, -1);",
				ExpectedErr: "pos_int_check",
			},
		},
	})

	backupUrl = localBackupUrl(t, "backup")
	runBackupTest(t, ScriptTest{
		Name: "user-defined composite type is preserved across backup and restore",
		SetUpScript: []string{
			`create type point_t as (x float8, y float8);`,
			`create function distance(p point_t) returns float8 as $$
			 begin return sqrt((p).x * (p).x + (p).y * (p).y); end;
			 $$ language plpgsql;`,
			fmt.Sprintf("select dolt_backup('sync-url', '%s');", backupUrl),
			fmt.Sprintf("select dolt_backup('restore', '%s', 'restored_composite');", backupUrl),
			"USE restored_composite",
		},
		Assertions: []ScriptTestAssertion{
			{
				// typtype = 'c' for composite types.
				Query:    "select typname, typtype from pg_type where typname = 'point_t';",
				Expected: []sql.Row{{"point_t", "c"}},
			},
			{
				Query:    "select (row(3.0, 4.0)::point_t).x;",
				Expected: []sql.Row{{float64(3.0)}},
			},
			{
				// A function accepting the composite type must still be callable after restore.
				Query:    "select distance(row(3.0, 4.0)::point_t);",
				Expected: []sql.Row{{float64(5.0)}},
			},
		},
	})

	backupUrl = localBackupUrl(t, "backup")
	runBackupTest(t, ScriptTest{
		Name: "user-defined function is preserved across backup and restore",
		SetUpScript: []string{
			`create function double_it(n int) returns int as $$
			 begin return n * 2; end;
			 $$ language plpgsql;`,
			fmt.Sprintf("select dolt_backup('sync-url', '%s');", backupUrl),
			fmt.Sprintf("select dolt_backup('restore', '%s', 'restored_funcs');", backupUrl),
			"USE restored_funcs",
		},
		Assertions: []ScriptTestAssertion{
			{
				// pg_proc is an unimplemented stub; calling the function is the best available proof.
				Query:    "select double_it(21);",
				Expected: []sql.Row{{42}},
			},
			{
				Query:    "select double_it(0);",
				Expected: []sql.Row{{0}},
			},
		},
	})

	backupUrl = localBackupUrl(t, "backup")
	runBackupTest(t, ScriptTest{
		Name: "stored procedure is preserved across backup and restore",
		SetUpScript: []string{
			"create table job_log (id int primary key, status text);",
			"insert into job_log values (1, 'pending');",
			`create procedure mark_done(job_id int) as $$
			 begin update job_log set status = 'done' where id = job_id; end;
			 $$ language plpgsql;`,
			fmt.Sprintf("select dolt_backup('sync-url', '%s');", backupUrl),
			fmt.Sprintf("select dolt_backup('restore', '%s', 'restored_procs');", backupUrl),
			"USE restored_procs",
		},
		Assertions: []ScriptTestAssertion{
			{
				// pg_proc is an unimplemented stub; calling the procedure is the best available proof.
				Query:            "call mark_done(1);",
				SkipResultsCheck: true,
			},
			{
				Query:    "select status from job_log where id = 1;",
				Expected: []sql.Row{{"done"}},
			},
		},
	})

	backupUrl = localBackupUrl(t, "backup")
	runBackupTest(t, ScriptTest{
		Name: "trigger is preserved and fires after restore",
		SetUpScript: []string{
			"create table readings (id int primary key, val int);",
			`create function clamp_val() returns trigger as $$
			 begin
			   if NEW.val > 100 then NEW.val := 100; end if;
			   return NEW;
			 end;
			 $$ language plpgsql;`,
			"create trigger clamp_trigger before insert on readings for each row execute function clamp_val();",
			fmt.Sprintf("select dolt_backup('sync-url', '%s');", backupUrl),
			fmt.Sprintf("select dolt_backup('restore', '%s', 'restored_triggers');", backupUrl),
			"USE restored_triggers",
		},
		Assertions: []ScriptTestAssertion{
			{
				// pg_trigger is an unimplemented stub; the clamp firing is the best available proof.
				// Inserting val=200 must be clamped to 100 by the trigger.
				Query:            "insert into readings values (1, 200);",
				SkipResultsCheck: true,
			},
			{
				Query:    "select val from readings where id = 1;",
				Expected: []sql.Row{{100}},
			},
		},
	})

	// TODO: Extension loading in Windows CI environments don't work currently
	if !(runtime.GOOS == "windows" && os.Getenv("CI") != "") {
		backupUrl = localBackupUrl(t, "backup")
		runBackupTest(t, ScriptTest{
			Name: "extension is preserved across backup and restore",
			SetUpScript: []string{
				`create extension "uuid-ossp";`,
				fmt.Sprintf("select dolt_backup('sync-url', '%s');", backupUrl),
				fmt.Sprintf("select dolt_backup('restore', '%s', 'restored_ext');", backupUrl),
				"USE restored_ext",
			},
			Assertions: []ScriptTestAssertion{
				{
					// pg_extension is an unimplemented stub; calling the extension function is the best available proof.
					// uuid_generate_v4() returns a 36-character UUID (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx).
					Query:    "select length(uuid_generate_v4()::text) = 36;",
					Expected: []sql.Row{{"t"}},
				},
			},
		})
	}

	backupUrl = localBackupUrl(t, "backup")
	runBackupTest(t, ScriptTest{
		Name: "named schema contents are preserved across backup and restore",
		SetUpScript: []string{
			"create schema inventory;",
			"create table inventory.products (id int primary key, name text);",
			"insert into inventory.products values (1, 'widget'), (2, 'gadget');",
			"select dolt_commit('-Am', 'add inventory schema');",
			fmt.Sprintf("select dolt_backup('sync-url', '%s');", backupUrl),
			fmt.Sprintf("select dolt_backup('restore', '%s', 'restored_schema');", backupUrl),
			"USE restored_schema",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select schema_name from information_schema.schemata where schema_name = 'inventory';",
				Expected: []sql.Row{{"inventory"}},
			},
			{
				Query:    "select id, name from inventory.products order by id;",
				Expected: []sql.Row{{1, "widget"}, {2, "gadget"}},
			},
		},
	})

	backupUrl = localBackupUrl(t, "backup")
	runBackupTest(t, ScriptTest{
		Name: "custom cast is preserved across backup and restore",
		SetUpScript: []string{
			// CREATE TABLE implicitly creates a composite row type of the same name.
			`CREATE TABLE cast_src (v text);`,
			`CREATE TABLE cast_dst (v text, tag text);`,
			`CREATE FUNCTION cast_src_to_dst(src cast_src) RETURNS cast_dst AS $$
			 SELECT ROW((src).v, 'casted')::cast_dst
			 $$ LANGUAGE SQL;`,
			`CREATE FUNCTION cast_verify(dst cast_dst) RETURNS text AS $$
			 SELECT (dst).v || ':' || (dst).tag
			 $$ LANGUAGE SQL;`,
			`CREATE CAST (cast_src AS cast_dst) WITH FUNCTION cast_src_to_dst(cast_src);`,
			fmt.Sprintf("select dolt_backup('sync-url', '%s');", backupUrl),
			fmt.Sprintf("select dolt_backup('restore', '%s', 'restored_casts');", backupUrl),
			"USE restored_casts",
		},
		Assertions: []ScriptTestAssertion{
			{
				// pg_cast is implemented and returns user-defined casts.
				// context 'e' = explicit, method 'f' = function.
				Query: `SELECT c.castcontext::text, c.castmethod::text
					FROM pg_cast c
					JOIN pg_type src ON src.oid = c.castsource
					JOIN pg_type dst ON dst.oid = c.casttarget
					WHERE src.typname = 'cast_src' AND dst.typname = 'cast_dst';`,
				Expected: []sql.Row{{"e", "f"}},
			},
			{
				Query:    `SELECT cast_verify((ROW('hello')::cast_src)::cast_dst);`,
				Expected: []sql.Row{{"hello:casted"}},
			},
		},
	})
}
