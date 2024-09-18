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
	"testing"

	regex "github.com/dolthub/go-icu-regex"
	"github.com/stretchr/testify/require"
)

func TestRegressionTests(t *testing.T) {
	// We'll only run this on GitHub Actions, so set this environment variable to run locally
	if _, ok := os.LookupEnv("REGRESSION_TESTING"); !ok {
		t.Skip()
	}
	regex.ShouldPanic = false // Something that is occurring in a test is causing this to panic, so we disable for now
	controller, port, err := CreateDoltgresServer()
	require.NoError(t, err)
	defer func() {
		controller.Stop()
		err = controller.WaitForStop()
		require.NoError(t, err)
	}()

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
			FailQueries: []string{
				`CREATE VIEW lock_view7 AS SELECT * from lock_view2;`,
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
			},
		})
		require.NoError(t, err)
		trackers = append(trackers, tracker)
	}
	fmt.Printf("Finished, writing output to `%s`\n", regressionFolder.GetAbsolutePath("out/results.trackers"))
	err = regressionFolder.WriteReplayTrackers("out/results.trackers", trackers, 0644)
	require.NoError(t, err)
}
