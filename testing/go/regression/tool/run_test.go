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
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/dolthub/dolt/go/libraries/utils/svcs"

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

	trackers := make([]*ReplayTracker, 0, len(AllTestResultFilesNames))
	for _, fileName := range AllTestResultFilesNames {
		messages, err := regressionFolder.ReadMessages(fileName)
		require.NoError(t, err)
		tracker, err := Replay(ReplayOptions{
			File:         fileName,
			Port:         port,
			Messages:     messages,
			PrintQueries: false,
			FailPSQL:     true,
			FailQueries:  queriesToSkip,
			Timeout:      timeout,
		})
		require.NoError(t, err)
		trackers = append(trackers, tracker)
	}
	outputPath := "out/results.trackers"
	if path, ok := os.LookupEnv("REGRESSION_OUTPUT"); ok {
		outputPath = path
	}
	fmt.Printf("Finished, writing output to `%s`\n", regressionFolder.GetAbsolutePath(outputPath))
	err = regressionFolder.WriteReplayTrackers(outputPath, trackers, 0644)
	require.NoError(t, err)
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
