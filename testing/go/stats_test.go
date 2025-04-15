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

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestStats(t *testing.T) {
	RunScripts(t, StatsTests)
}

var StatsTests = []ScriptTest{
	{
		Name: "ANALYZE statement",
		SetUpScript: []string{
			"CREATE TABLE t (pk int primary key);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "ANALYZE;",
				Expected: []sql.Row{},
			},
			{
				Query:    "ANALYZE t;",
				Expected: []sql.Row{},
			},
			{
				Query:    "ANALYZE public.t;",
				Expected: []sql.Row{},
			},
			{
				Query:    "ANALYZE postgres.public.t;",
				Expected: []sql.Row{},
			},
			{
				Query:       "ANALYZE doesnotexists.public.t;",
				ExpectedErr: "ERROR: database not found: doesnotexists (errno 1049) (sqlstate HY000) (SQLSTATE XX000)",
			},
		},
	},
}
