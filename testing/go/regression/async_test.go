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

package regression

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestAsync(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_async)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_async,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT pg_notify('notify_async1','sample message1');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_notify('notify_async1','');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT pg_notify('notify_async1',NULL);`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement:   `SELECT pg_notify('','sample message1');`,
				ErrorString: `channel name cannot be empty`,
			},
			{
				Statement:   `SELECT pg_notify(NULL,'sample message1');`,
				ErrorString: `channel name cannot be empty`,
			},
			{
				Statement:   `SELECT pg_notify('notify_async_channel_name_too_long______________________________','sample_message1');`,
				ErrorString: `channel name too long`,
			},
			{
				Statement: `NOTIFY notify_async2;`,
			},
			{
				Statement: `LISTEN notify_async2;`,
			},
			{
				Statement: `UNLISTEN notify_async2;`,
			},
			{
				Statement: `UNLISTEN *;`,
			},
			{
				Statement: `SELECT pg_notification_queue_usage();`,
				Results:   []sql.Row{{0}},
			},
		},
	})
}
