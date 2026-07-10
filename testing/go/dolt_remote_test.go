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
	"strings"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/require"
)

// localRemoteUrl returns a file:// URL for a fresh temp directory suitable for use as a Dolt
// remote. Each call produces a unique directory, so two calls with the same |name| never collide.
// We avoid using [testing.T.TempDir] here, because on Windows, NBS chunk files can still be
// memory-mapped and locked for a brief window after a server's controller reports itself stopped,
// which interfers with the cleanup.
func localRemoteUrl(t *testing.T, name string) string {
	t.Helper()
	safeName := strings.ReplaceAll(t.Name(), "/", "_")
	dir, err := os.MkdirTemp(os.TempDir(), safeName+"_"+name)
	require.NoError(t, err)
	return "file://" + filepath.ToSlash(dir)
}

// runRemoteTest runs |script| inside a dedicated, freshly-started on-disk Doltgres server and
// tears the server down when the subtest completes.
//
// A new server is started for every call so that each push/pull/clone test begins with a clean
// database state. The server runs against a temporary on-disk directory (via
// CreateServerLocalWithPort) because pushing to, pulling from, or cloning a file:// remote
// requires a real NBS chunk store -- in-memory storage cannot be synced to a file:// URL.
//
// dolt_push/dolt_pull/dolt_fetch/dolt_clone/dolt_remote are all invoked with SELECT rather than
// CALL: Doltgres only allows built-in Dolt procedures to be invoked as functions (see
// ErrDoltProcedureSelectOnly in server/functions/dolt_procedures.go).
func runRemoteTest(t *testing.T, script ScriptTest) {
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

// TestDoltRemote exercises the file-system remote push/pull/fetch/clone workflow end to end.
func TestDoltRemote(t *testing.T) {
	remoteUrl := localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "push and clone preserve table data and commit history",
		SetUpScript: []string{
			"create table items (id int primary key, label text not null);",
			"insert into items values (1, 'apple'), (2, 'banana');",
			"select dolt_commit('-Am', 'first commit: add apple and banana');",
			"insert into items values (3, 'cherry');",
			"select dolt_commit('-Am', 'second commit: add cherry');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			fmt.Sprintf("select dolt_clone('%s', 'cloned_items');", remoteUrl),
			"USE cloned_items",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select id, label from items order by id;",
				Expected: []sql.Row{{1, "apple"}, {2, "banana"}, {3, "cherry"}},
			},
			{
				// Initialize + CREATE DATABASE + 2 user commits.
				Query:    "select count(*) from dolt.commits;",
				Expected: []sql.Row{{int64(4)}},
			},
			{
				Query:    "select message from dolt.commits order by date;",
				Expected: []sql.Row{{"Initialize data repository"}, {"CREATE DATABASE"}, {"first commit: add apple and banana"}, {"second commit: add cherry"}},
			},
			{
				// dolt_clone automatically configures "origin" to point back at the source URL.
				Query:    "select name from dolt.remotes;",
				Expected: []sql.Row{{"origin"}},
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "standalone sequence state is preserved across push and clone",
		SetUpScript: []string{
			"create sequence counter start 1 increment 5;",
			"select nextval('counter');", // advances to 1
			"select nextval('counter');", // advances to 6
			"select nextval('counter');", // advances to 11
			"select dolt_commit('-Am', 'seed counter');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			fmt.Sprintf("select dolt_clone('%s', 'cloned_seq');", remoteUrl),
			"drop sequence counter", // delete the sequence in the first database
			"USE cloned_seq",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select sequence_name, sequence_catalog from information_schema.sequences where sequence_schema = 'public';",
				Expected: []sql.Row{{"counter", "cloned_seq"}},
			},
			{
				// Must continue from where it left off (after 11, next is 16), not reset to 1.
				Query:    "select nextval('counter');",
				Expected: []sql.Row{{int64(16)}},
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "serial column's owned sequence state is preserved across push and clone",
		SetUpScript: []string{
			"create table widgets (id serial primary key, name text);",
			"insert into widgets (name) values ('a'), ('b'), ('c');",
			"select dolt_commit('-Am', 'seed widgets');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			fmt.Sprintf("select dolt_clone('%s', 'cloned_widgets');", remoteUrl),
			"drop table widgets",
			"USE cloned_widgets",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select id, name from widgets order by id;",
				Expected: []sql.Row{{1, "a"}, {2, "b"}, {3, "c"}},
			},
			{
				// Auto-generated id must continue from 4, proving the owned sequence's state was pushed.
				Query:    "insert into widgets (name) values ('d') returning id;",
				Expected: []sql.Row{{4}},
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "identity column's owned sequence state is preserved across push and clone",
		SetUpScript: []string{
			"create table gadgets (id bigint generated by default as identity primary key, name text);",
			"insert into gadgets (name) values ('a'), ('b'), ('c');",
			"select dolt_add('.');",
			"select dolt_commit('-m', 'seed gadgets');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			fmt.Sprintf("select dolt_clone('%s', 'cloned_gadgets');", remoteUrl),
			"drop table gadgets",
			"USE cloned_gadgets",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select id, name from gadgets order by id;",
				Expected: []sql.Row{{1, "a"}, {2, "b"}, {3, "c"}},
			},
			{
				Query:    "insert into gadgets (name) values ('d') returning id;",
				Expected: []sql.Row{{4}},
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "sequence in a non-public schema is preserved across push and clone",
		SetUpScript: []string{
			"create schema myschema;",
			"create sequence myschema.seq2 start 100 increment 10;",
			"select nextval('myschema.seq2');", // advances to 100
			"select dolt_add('.');",
			"select dolt_commit('-m', 'seed myschema.seq2');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			fmt.Sprintf("select dolt_clone('%s', 'cloned_schema');", remoteUrl),
			"USE cloned_schema",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select sequence_schema, sequence_name from information_schema.sequences where sequence_name = 'seq2';",
				Expected: []sql.Row{{"myschema", "seq2"}},
			},
			{
				// Must continue from 100 (after one call), not reset to the start value.
				Query:    "select nextval('myschema.seq2');",
				Expected: []sql.Row{{int64(110)}},
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "incremental push and pull keep sequence and table state in sync",
		SetUpScript: []string{
			"create sequence counter start 1 increment 5;",
			"create table orders (id int primary key default nextval('counter'), item text);",
			"select dolt_commit('-Am', 'initial schema');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			fmt.Sprintf("select dolt_clone('%s', 'replica');", remoteUrl),
			// Advance the sequence and add data on the source side only.
			"USE postgres",
			"insert into orders (item) values ('widget');", // consumes nextval -> 1
			"select dolt_commit('-Am', 'first order');",
			"select dolt_push('origin', 'main');",
			// Pull the change into the replica.
			"USE replica",
			"select dolt_pull('origin');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select id, item from orders order by id;",
				Expected: []sql.Row{{1, "widget"}},
			},
			{
				Query:    "select nextval('counter');",
				Expected: []sql.Row{{int64(6)}},
			},
		},
	})

	runRemoteTest(t, ScriptTest{
		Name: "dolt_remotes reflects add and remove",
		SetUpScript: []string{
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", localRemoteUrl(t, "remote_a")),
			fmt.Sprintf("select dolt_remote('add', 'other', '%s');", localRemoteUrl(t, "remote_b")),
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select name from dolt_remotes order by name;",
				Expected: []sql.Row{{"origin"}, {"other"}},
			},
			{
				Query:    "select dolt_remote('remove', 'other');",
				Expected: []sql.Row{{"{0}"}},
			},
			{
				Query:    "select name from dolt_remotes;",
				Expected: []sql.Row{{"origin"}},
			},
			{
				Query:       "select dolt_remote('remove', 'other');",
				ExpectedErr: "unknown remote",
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "fetch updates remote-tracking branches without touching the working branch; pull merges",
		SetUpScript: []string{
			"create table events (id int primary key, name text);",
			"insert into events values (1, 'first');",
			"select dolt_commit('-Am', 'first event');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			fmt.Sprintf("select dolt_clone('%s', 'subscriber');", remoteUrl),
			"USE postgres",
			"insert into events values (2, 'second');",
			"select dolt_commit('-Am', 'second event');",
			"select dolt_push('origin', 'main');",
			"USE subscriber",
			"select dolt_fetch('origin', 'main');",
		},
		Assertions: []ScriptTestAssertion{
			{
				// fetch alone must not update the working branch.
				Query:    "select id, name from events order by id;",
				Expected: []sql.Row{{1, "first"}},
			},
			{
				Query:    "select name from dolt_remote_branches;",
				Expected: []sql.Row{{"remotes/origin/main"}},
			},
			{
				Query:    "select dolt_pull('origin');",
				Expected: []sql.Row{{`{1,0,"merge successful"}`}},
			},
			{
				// pull must fast-forward the working branch to match the remote.
				Query:    "select id, name from events order by id;",
				Expected: []sql.Row{{1, "first"}, {2, "second"}},
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "pull fails when the working set has uncommitted changes",
		SetUpScript: []string{
			"create table logs (id int primary key, msg text);",
			"insert into logs values (1, 'a');",
			"select dolt_commit('-Am', 'seed');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			fmt.Sprintf("select dolt_clone('%s', 'dirty_clone');", remoteUrl),
			"USE postgres",
			"insert into logs values (2, 'b');",
			"select dolt_commit('-Am', 'second');",
			"select dolt_push('origin', 'main');",
			"USE dirty_clone",
			// Leave an uncommitted change on the clone before pulling.
			"insert into logs values (3, 'uncommitted');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:       "select dolt_pull('origin');",
				ExpectedErr: "cannot merge with uncommitted changes",
			},
		},
	})

	runRemoteTest(t, ScriptTest{
		Name: "push without a configured remote returns an error",
		SetUpScript: []string{
			"create table t (id int primary key);",
			"select dolt_add('.');",
			"select dolt_commit('-m', 'seed');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:       "select dolt_push('origin', 'main');",
				ExpectedErr: "remote 'origin' not found",
			},
		},
	})

	runRemoteTest(t, ScriptTest{
		Name: "pull without a configured remote returns an error",
		SetUpScript: []string{
			"create table t (id int primary key);",
			"select dolt_add('.');",
			"select dolt_commit('-m', 'seed');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:       "select dolt_pull('origin');",
				ExpectedErr: "remote 'origin' not found",
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "custom enum type is preserved across push and clone",
		SetUpScript: []string{
			"create type mood as enum ('sad', 'ok', 'happy');",
			"create table moods (id int primary key, m mood);",
			"insert into moods values (1, 'happy'), (2, 'sad');",
			"select dolt_commit('-Am', 'seed moods');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			fmt.Sprintf("select dolt_clone('%s', 'cloned_moods');", remoteUrl),
			"drop table moods;",
			"drop type mood;",
			"USE cloned_moods",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select id, m::text from moods order by id;",
				Expected: []sql.Row{{1, "happy"}, {2, "sad"}},
			},
			{
				// The enum type definition itself (not just data using it) must have transferred.
				Query:    "insert into moods values (3, 'ok') returning m::text;",
				Expected: []sql.Row{{"ok"}},
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "user-defined function is preserved across push and clone",
		SetUpScript: []string{
			"create function double_it(x int) returns int as $$ begin return x * 2; end; $$ language plpgsql;",
			"select dolt_commit('-Am', 'seed function');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			fmt.Sprintf("select dolt_clone('%s', 'cloned_func');", remoteUrl),
			"drop function double_it();",
			"USE cloned_func",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select double_it(21);",
				Expected: []sql.Row{{42}},
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "pushing and cloning a non-default branch carries its own sequence and table state",
		SetUpScript: []string{
			"create sequence counter start 1 increment 1;",
			"create table gadgets (id int primary key default nextval('counter'), name text);",
			"insert into gadgets (name) values ('a'), ('b'), ('c');", // consumes 1, 2, 3
			"select dolt_commit('-Am', 'seed on main');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			"select dolt_checkout('-b', 'feature');",
			"insert into gadgets (name) values ('feature-only');", // consumes 4
			"select dolt_commit('-Am', 'feature work');",
			"select dolt_push('origin', 'feature');",
			fmt.Sprintf("select dolt_clone('--branch', 'feature', '%s', 'cloned_feature');", remoteUrl),
			"drop table gadgets;",
			"drop sequence counter;",
			"USE cloned_feature",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select name from gadgets order by id;",
				Expected: []sql.Row{{"a"}, {"b"}, {"c"}, {"feature-only"}},
			},
			{
				// The owned sequence must reflect the feature branch's state (after 4), not main's.
				Query:    "insert into gadgets (name) values ('next') returning id;",
				Expected: []sql.Row{{5}},
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "non-fast-forward push is rejected, then succeeds with --force",
		SetUpScript: []string{
			"create table t (id int primary key, note text);",
			"insert into t values (1, 'source');",
			"select dolt_commit('-Am', 'seed');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			fmt.Sprintf("select dolt_clone('%s', 'clone1');", remoteUrl),
			// Diverge history on both sides so neither is a fast-forward of the other.
			"insert into t values (2, 'source-diverged');",
			"select dolt_commit('-Am', 'source diverges');",
			"select dolt_push('origin', 'main');",
			"USE clone1",
			"insert into t values (3, 'clone-diverged');",
			"select dolt_commit('-Am', 'clone diverges');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:       "select dolt_push('origin', 'main');",
				ExpectedErr: "failed to push some refs",
			},
			{
				Query:            "select dolt_push('-f', 'origin', 'main');",
				SkipResultsCheck: true,
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "pushing a branch delete removes it from the remote",
		SetUpScript: []string{
			"create table t (id int primary key);",
			"insert into t values (1);",
			"select dolt_commit('-Am', 'seed');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			"select dolt_checkout('-b', 'doomed');",
			"insert into t values (2);",
			"select dolt_commit('-Am', 'doomed work');",
			"select dolt_push('origin', 'doomed');",
			fmt.Sprintf("select dolt_clone('%s', 'observer');", remoteUrl),
			"USE observer",
			"select dolt_fetch('origin', 'doomed');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select name from dolt_remote_branches order by name;",
				Expected: []sql.Row{{"remotes/origin/doomed"}, {"remotes/origin/main"}},
			},
			{
				Query:            "USE postgres",
				SkipResultsCheck: true,
			},
			{
				// A ":branch" refspec deletes the branch on the remote, just like the CLI's `dolt push origin :doomed`.
				Query:            "select dolt_push('origin', ':doomed');",
				SkipResultsCheck: true,
			},
			{
				Query:            "USE observer",
				SkipResultsCheck: true,
			},
			{
				Query:            "select dolt_fetch('--prune', 'origin');",
				SkipResultsCheck: true,
			},
			{
				Query:    "select name from dolt_remote_branches order by name;",
				Expected: []sql.Row{{"remotes/origin/main"}},
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "tags are pushed and fetched along with commits",
		SetUpScript: []string{
			"create table t (id int primary key);",
			"insert into t values (1);",
			"select dolt_commit('-Am', 'seed');",
			"select dolt_tag('v1.0', '-m', 'first release');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			"select dolt_push('origin', 'v1.0');",
			fmt.Sprintf("select dolt_clone('%s', 'cloned_tags');", remoteUrl),
			"USE cloned_tags",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select tag_name from dolt_tags;",
				Expected: []sql.Row{{"v1.0"}},
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "fetch supports an explicit refspec into a custom tracking ref",
		SetUpScript: []string{
			"create table t (id int primary key);",
			"insert into t values (1);",
			"select dolt_commit('-Am', 'seed');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			fmt.Sprintf("select dolt_clone('%s', 'observer');", remoteUrl),
			"USE observer",
			"select dolt_fetch('origin', 'refs/heads/main:refs/remotes/custom/main');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select name from dolt_remote_branches where name = 'remotes/custom/main';",
				Expected: []sql.Row{{"remotes/custom/main"}},
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "push --all pushes every local branch",
		SetUpScript: []string{
			"create table t (id int primary key);",
			"insert into t values (1);",
			"select dolt_commit('-Am', 'seed');",
			"select dolt_checkout('-b', 'branch_a');",
			"select dolt_checkout('main');",
			"select dolt_checkout('-b', 'branch_b');",
			"select dolt_checkout('main');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('--all', 'origin');",
			fmt.Sprintf("select dolt_clone('%s', 'cloned_all');", remoteUrl),
			"USE cloned_all",
		},
		Assertions: []ScriptTestAssertion{
			{
				// Clone only checks out a local branch for the default branch; the others land as
				// remote-tracking refs, proving --all pushed them all to the remote.
				Query:    "select name from dolt_remote_branches order by name;",
				Expected: []sql.Row{{"remotes/origin/branch_a"}, {"remotes/origin/branch_b"}, {"remotes/origin/main"}},
			},
		},
	})

	runRemoteTest(t, ScriptTest{
		Name: "adding a remote with a name that already exists returns an error",
		SetUpScript: []string{
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", localRemoteUrl(t, "remote_a")),
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:       fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", localRemoteUrl(t, "remote_b")),
				ExpectedErr: "remote already exists",
			},
		},
	})

	runRemoteTest(t, ScriptTest{
		Name: "adding a remote with an invalid name returns an error",
		Assertions: []ScriptTestAssertion{
			{
				Query:       fmt.Sprintf("select dolt_remote('add', 'bad name', '%s');", localRemoteUrl(t, "remote")),
				ExpectedErr: "remote name invalid",
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "fetching an invalid ref spec returns an error",
		SetUpScript: []string{
			"create table t (id int primary key);",
			"select dolt_commit('-Am', 'seed');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:       "select dolt_fetch('origin', 'garbage');",
				ExpectedErr: "invalid ref spec",
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "pull auto-merges non-conflicting divergent history",
		SetUpScript: []string{
			"create table t (id int primary key, v text);",
			"insert into t values (1, 'a'), (2, 'b');",
			"select dolt_commit('-Am', 'seed');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			fmt.Sprintf("select dolt_clone('%s', 'clone1');", remoteUrl),
			"USE postgres",
			"insert into t values (3, 'source-row');",
			"select dolt_commit('-Am', 'source adds row 3');",
			"select dolt_push('origin', 'main');",
			"USE clone1",
			"insert into t values (4, 'clone-row');",
			"select dolt_commit('-Am', 'clone adds row 4');",
		},
		Assertions: []ScriptTestAssertion{
			{
				// Divergent but non-conflicting: no fast-forward, no conflicts, and it merges automatically.
				Query:    "select dolt_pull('origin');",
				Expected: []sql.Row{{`{0,0,"merge successful"}`}},
			},
			{
				Query:    "select id, v from t order by id;",
				Expected: []sql.Row{{1, "a"}, {2, "b"}, {3, "source-row"}, {4, "clone-row"}},
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "pull surfaces real merge conflicts, resolvable via dolt_conflicts_resolve",
		SetUpScript: []string{
			"create table t (id int primary key, v text);",
			"insert into t values (1, 'a'), (2, 'b');",
			"select dolt_commit('-Am', 'seed');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			fmt.Sprintf("select dolt_clone('%s', 'clone1');", remoteUrl),
			"USE postgres",
			"update t set v = 'source-edit' where id = 1;",
			"select dolt_commit('-Am', 'source edits row 1');",
			"select dolt_push('origin', 'main');",
			"USE clone1",
			"update t set v = 'clone-edit' where id = 1;",
			"select dolt_commit('-Am', 'clone edits row 1');",
			// Conflicts abort the implicit transaction under autocommit unless explicitly allowed.
			"set dolt_allow_commit_conflicts to 1;",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select dolt_pull('origin');",
				Expected: []sql.Row{{`{0,1,"merge has unresolved conflicts or constraint violations"}`}},
			},
			{
				Query:    "select base_v, our_v, their_v from dolt_conflicts_t;",
				Expected: []sql.Row{{"a", "clone-edit", "source-edit"}},
			},
			{
				Query:            "select dolt_conflicts_resolve('--ours', 't');",
				SkipResultsCheck: true,
			},
			{
				Query:    "select count(*) from dolt_conflicts;",
				Expected: []sql.Row{{int64(0)}},
			},
			{
				// --ours keeps the clone's own edit.
				Query:    "select id, v from t order by id;",
				Expected: []sql.Row{{1, "clone-edit"}, {2, "b"}},
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "domain type and its check constraint are preserved across push and clone",
		SetUpScript: []string{
			"create domain pos_int as integer check (value > 0);",
			"create table measurements (id int primary key, val pos_int);",
			"insert into measurements values (1, 42);",
			"select dolt_commit('-Am', 'seed measurements');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			fmt.Sprintf("select dolt_clone('%s', 'cloned_domain');", remoteUrl),
			"drop table measurements;",
			"drop domain pos_int;",
			"USE cloned_domain",
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
				// The domain's check constraint must still be enforced after cloning.
				Query:       "insert into measurements values (2, -1);",
				ExpectedErr: "pos_int_check",
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "user-defined composite type is preserved across push and clone",
		SetUpScript: []string{
			`create type point_t as (x float8, y float8);`,
			`create function distance(p point_t) returns float8 as $$ begin return sqrt((p).x * (p).x + (p).y * (p).y); end; $$ language plpgsql;`,
			"select dolt_commit('-Am', 'seed composite type');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			fmt.Sprintf("select dolt_clone('%s', 'cloned_composite');", remoteUrl),
			"drop function distance(point_t);",
			"drop type point_t;",
			"USE cloned_composite",
		},
		Assertions: []ScriptTestAssertion{
			{
				// typtype = 'c' for composite types.
				Query:    "select typname, typtype from pg_type where typname = 'point_t';",
				Expected: []sql.Row{{"point_t", "c"}},
			},
			{
				Query:    "select distance(row(3.0, 4.0)::point_t);",
				Expected: []sql.Row{{float64(5.0)}},
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "stored procedure is preserved across push and clone",
		SetUpScript: []string{
			"create table job_log (id int primary key, status text);",
			"insert into job_log values (1, 'pending');",
			`create procedure mark_done(job_id int) as $$ begin update job_log set status = 'done' where id = job_id; end; $$ language plpgsql;`,
			"select dolt_commit('-Am', 'seed procedure');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			fmt.Sprintf("select dolt_clone('%s', 'cloned_proc');", remoteUrl),
			"drop procedure mark_done(int);",
			"drop table job_log;",
			"USE cloned_proc",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:            "call mark_done(1);",
				SkipResultsCheck: true,
			},
			{
				Query:    "select status from job_log where id = 1;",
				Expected: []sql.Row{{"done"}},
			},
		},
	})

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "trigger is preserved and fires after clone",
		SetUpScript: []string{
			"create table readings (id int primary key, val int);",
			`create function clamp_val() returns trigger as $$ begin if NEW.val > 100 then NEW.val := 100; end if; return NEW; end; $$ language plpgsql;`,
			"create trigger clamp_trigger before insert on readings for each row execute function clamp_val();",
			"select dolt_commit('-Am', 'seed trigger');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			fmt.Sprintf("select dolt_clone('%s', 'cloned_trigger');", remoteUrl),
			"drop trigger clamp_trigger on readings;",
			"drop function clamp_val();",
			"drop table readings;",
			"USE cloned_trigger",
		},
		Assertions: []ScriptTestAssertion{
			{
				// Inserting val=200 must be clamped to 100 by the trigger, proving it still fires post-clone.
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
		remoteUrl = localRemoteUrl(t, "remote")
		runRemoteTest(t, ScriptTest{
			Name: "extension is preserved across push and clone",
			SetUpScript: []string{
				`create extension "uuid-ossp";`,
				"select dolt_commit('-Am', 'seed extension');",
				fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
				"select dolt_push('origin', 'main');",
				fmt.Sprintf("select dolt_clone('%s', 'cloned_ext');", remoteUrl),
				"USE cloned_ext",
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

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "named schema contents are preserved across push and clone",
		SetUpScript: []string{
			"create schema inventory;",
			"create table inventory.products (id int primary key, name text);",
			"insert into inventory.products values (1, 'widget'), (2, 'gadget');",
			"select dolt_commit('-Am', 'add inventory schema');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			fmt.Sprintf("select dolt_clone('%s', 'cloned_schema_contents');", remoteUrl),
			"drop table inventory.products;",
			"drop schema inventory;",
			"USE cloned_schema_contents",
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

	remoteUrl = localRemoteUrl(t, "remote")
	runRemoteTest(t, ScriptTest{
		Name: "custom cast is preserved across push and clone",
		SetUpScript: []string{
			// CREATE TABLE implicitly creates a composite row type of the same name.
			`CREATE TABLE cast_src (v text);`,
			`CREATE TABLE cast_dst (v text, tag text);`,
			`CREATE FUNCTION cast_src_to_dst(src cast_src) RETURNS cast_dst AS $$ SELECT ROW((src).v, 'casted')::cast_dst $$ LANGUAGE SQL;`,
			`CREATE FUNCTION cast_verify(dst cast_dst) RETURNS text AS $$ SELECT (dst).v || ':' || (dst).tag $$ LANGUAGE SQL;`,
			`CREATE CAST (cast_src AS cast_dst) WITH FUNCTION cast_src_to_dst(cast_src);`,
			"select dolt_commit('-Am', 'seed cast');",
			fmt.Sprintf("select dolt_remote('add', 'origin', '%s');", remoteUrl),
			"select dolt_push('origin', 'main');",
			fmt.Sprintf("select dolt_clone('%s', 'cloned_cast');", remoteUrl),
			"drop cast (cast_src as cast_dst);",
			"drop function cast_src_to_dst(cast_src);",
			"drop function cast_verify(cast_dst);",
			"drop table cast_src;",
			"drop table cast_dst;",
			"USE cloned_cast",
		},
		Assertions: []ScriptTestAssertion{
			{
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
