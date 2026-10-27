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

package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dolthub/dolt/go/libraries/utils/svcs"
	"github.com/jackc/pgx/v5"

	"github.com/dolthub/doltgresql/server"

	"github.com/stretchr/testify/require"
)

func TestRegressionTests(t *testing.T) {
	// We'll only run this on GitHub Actions, so set this environment variable to run locally
	if _, ok := os.LookupEnv("REGRESSION_TESTING"); !ok {
		t.Skip()
	}
	server.EnableAuthentication = false // We have to disable authentication, since we can't replay the messages due to nonces
	port, err := externalServerPort()
	require.NoError(t, err)
	timeout, err := replayTimeout()
	require.NoError(t, err)
	if port == 0 {
		var controller *svcs.Controller
		controller, port, err = CreateDoltgresServer()
		require.NoError(t, err)
		defer func() {
			controller.Stop()
			err = controller.WaitForStop()
			require.NoError(t, err)
		}()
	}

	outputPath := "out/results.trackers"
	if path, ok := os.LookupEnv("REGRESSION_OUTPUT"); ok {
		outputPath = path
	}

	failQueries, err := extraSkipQueries()
	require.NoError(t, err)

	fileNames, err := selectedFileNames()
	require.NoError(t, err)

	trackers := make([]*ReplayTracker, 0, len(fileNames))
	for _, fileName := range fileNames {
		messages, err := regressionFolder.ReadMessages(fileName)
		require.NoError(t, err)
		tracker, err := Replay(ReplayOptions{
			File:         fileName,
			Port:         port,
			Messages:     messages,
			PrintQueries: false,
			FailPSQL:     true,
			FailQueries:  failQueries,
			Timeout:      timeout,
		})
		if err != nil {
			// An external server may crash or drop the connection mid-file. Record the abort and keep going, so
			// one crash does not discard the results of every other file.
			fmt.Printf("REPLAY ABORTED in `%s`: %s\n", fileName, err.Error())
			tracker = NewReplayTracker(fileName)
			tracker.Failed = 1
			tracker.AddFailure(ReplayTrackerItem{
				Query:           "<replay aborted>",
				UnexpectedError: err.Error(),
			})
			waitForServer(port)
		}

		trackers = append(trackers, tracker)
		if port != 0 {
			// Write partial results after each file, since external replays can take many hours
			err = regressionFolder.WriteReplayTrackers(outputPath, trackers, 0644)
			require.NoError(t, err)
		}
	}

	fmt.Printf("Finished, writing output to `%s`\n", regressionFolder.GetAbsolutePath(outputPath))
	err = regressionFolder.WriteReplayTrackers(outputPath, trackers, 0644)
	require.NoError(t, err)
}

// waitForServer waits up to five minutes for the server on the given port to accept connections again, such as after
// it restarts following a backend crash.
func waitForServer(port int) {
	url := fmt.Sprintf("postgres://postgres@127.0.0.1:%d/postgres?sslmode=disable", port)
	deadline := time.Now().Add(5 * time.Minute)
	for time.Now().Before(deadline) {
		conn, err := pgx.Connect(context.Background(), url)
		if err == nil {
			_ = conn.Close(context.Background())
			return
		}

		time.Sleep(time.Second)
	}
}

// replayTimeout returns how long to wait for each server response, as given by REGRESSION_TIMEOUT (a Go duration such
// as "10m"). Slow servers otherwise time out mid-statement and the replay moves on while the statement still runs.
// Returns zero, meaning the replay default, when the variable is not set.
func replayTimeout() (time.Duration, error) {
	value, ok := os.LookupEnv("REGRESSION_TIMEOUT")
	if !ok {
		return 0, nil
	}

	timeout, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("invalid REGRESSION_TIMEOUT `%s`: %w", value, err)
	}

	return timeout, nil
}

func TestExtraSkipQueries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "skip.txt")
	require.NoError(t, os.WriteFile(path, []byte("create type nestedcomposite as (x int8_tbl);\n\n  select 1;  \n"), 0644))
	t.Setenv("REGRESSION_SKIP_QUERIES_FILE", path)

	queries, err := extraSkipQueries()
	require.NoError(t, err)
	require.Equal(t, len(queriesToSkip)+2, len(queries))
	require.Equal(t, queriesToSkip, queries[:len(queriesToSkip)])
	require.Equal(t, []string{"create type nestedcomposite as (x int8_tbl);", "select 1;"}, queries[len(queriesToSkip):])
}

// selectedFileNames returns the result files to replay, in schedule order. REGRESSION_FILES may name a comma-separated
// subset of tests (such as "test_setup,collate.icu.utf8"), which lets a long external replay resume after a statement
// that left the server unusable. Returns all files when the variable is not set.
func selectedFileNames() ([]string, error) {
	value, ok := os.LookupEnv("REGRESSION_FILES")
	if !ok {
		return AllTestResultFilesNames, nil
	}

	wanted := make(map[string]bool)
	for _, name := range strings.Split(value, ",") {
		if name = strings.TrimSpace(name); name != "" {
			wanted["results/"+name+".results"] = true
		}
	}

	var fileNames []string
	for _, fileName := range AllTestResultFilesNames {
		if wanted[fileName] {
			fileNames = append(fileNames, fileName)
			delete(wanted, fileName)
		}
	}

	if len(wanted) > 0 {
		return nil, fmt.Errorf("unknown REGRESSION_FILES entries: %v", wanted)
	}

	return fileNames, nil
}

func TestSelectedFileNames(t *testing.T) {
	t.Setenv("REGRESSION_FILES", "boolean, test_setup")
	fileNames, err := selectedFileNames()
	require.NoError(t, err)
	require.Equal(t, []string{"results/test_setup.results", "results/boolean.results"}, fileNames)

	t.Setenv("REGRESSION_FILES", "test_setup,no_such_test")
	_, err = selectedFileNames()
	require.Error(t, err)
}

// extraSkipQueries returns queriesToSkip plus any queries listed in the file named by REGRESSION_SKIP_QUERIES_FILE, one
// query substring per non-empty line. This lets a replay against an external server skip statements known to leave it
// unusable for the remainder of the run, without changing the default list.
func extraSkipQueries() ([]string, error) {
	path, ok := os.LookupEnv("REGRESSION_SKIP_QUERIES_FILE")
	if !ok {
		return queriesToSkip, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("invalid REGRESSION_SKIP_QUERIES_FILE `%s`: %w", path, err)
	}

	queries := append([]string{}, queriesToSkip...)
	for _, line := range strings.Split(string(data), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			queries = append(queries, line)
		}
	}

	return queries, nil
}

// externalServerPort returns the port of an already-running server named by REGRESSION_SERVER_PORT, such as a
// PostgreSQL build, so that the recorded sessions may be replayed against it instead of an in-memory Doltgres
// server. Returns zero when the variable is not set. The external server must accept the recorded `postgres` user on
// 127.0.0.1 without password authentication.
func externalServerPort() (int, error) {
	value, ok := os.LookupEnv("REGRESSION_SERVER_PORT")
	if !ok {
		return 0, nil
	}

	port, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid REGRESSION_SERVER_PORT `%s`: %w", value, err)
	}

	return port, nil
}

var queriesToSkip = []string{
	`create index testtable_apple_index on testtable_apple(logdate);`,
	`create index testtable_orange_index on testtable_orange(logdate);`,
	`create table child_0_10 partition of parent_tab
  for values from (0) to (10);`,
	`create table child_10_20 partition of parent_tab
  for values from (10) to (20);`,
	`create table child_20_30 partition of parent_tab
  for values from (20) to (30);`,
	`create table child_30_35 partition of child_30_40
  for values from (30) to (35);`,
	`create table child_35_40 partition of child_30_40
   for values from (35) to (40);`,
	`select count(*)
from
  (select t3.tenthous as x1, coalesce(t1.stringu1, t2.stringu1) as x2
   from tenk1 t1
   left join tenk1 t2 on t1.unique1 = t2.unique1
   join tenk1 t3 on t1.unique2 = t3.unique2) ss,
  tenk1 t4,
  tenk1 t5
where t4.thousand = t5.unique1 and ss.x1 = t4.tenthous and ss.x2 = t5.stringu1;`,
	`select count(*) from
  (select * from tenk1 x order by x.thousand, x.twothousand, x.fivethous) x
  left join
  (select * from tenk1 y order by y.unique2) y
  on x.thousand = y.unique2 and x.twothousand = y.hundred and x.fivethous = y.unique2;`,
	`select count(*) from tenk1 a, tenk1 b
  where a.hundred = b.thousand and (b.fivethous % 10) < 10;`,
	`select a.unique2, a.ten, b.tenthous, b.unique2, b.hundred
from tenk1 a left join tenk1 b on a.unique2 = b.tenthous
where a.unique1 = 42 and
      ((b.unique2 is null and a.ten = 2) or b.hundred = 3);`,
	`select
  (select max((select i.unique2 from tenk1 i where i.unique1 = o.unique1)))
from tenk1 o;`,
	`SELECT pg_class.relname FROM pg_index, pg_class, pg_class AS pg_class_2
WHERE pg_class.oid=indexrelid
	AND indrelid=pg_class_2.oid
	AND pg_class_2.relname = 'clstr_tst'
	AND indisclustered;`,
	`SELECT 1 FROM pg_catalog.pg_constraint WHERE conrelid = i.indrelid AND conindid = i.indexrelid`,
	`SELECT generate_series(1, generate_series(1, 3))`,
	`SELECT generate_series(generate_series(1,3), generate_series(2, 4));`,
}
