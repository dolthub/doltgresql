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
				ExpectedErr: "ERROR: database not found: doesnotexists (SQLSTATE 3D000)",
			},
		},
	},
	{
		// https://github.com/dolthub/doltgresql/issues/3918
		Name: "ANALYZE options and column lists",
		SetUpScript: []string{
			"CREATE TABLE an_t (id int PRIMARY KEY, version int NOT NULL);",
			"CREATE TABLE an_u (id int PRIMARY KEY);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "ANALYZE (SKIP_LOCKED) an_t;",
				Expected: []sql.Row{},
			},
			{
				Query:    "ANALYZE (VERBOSE) an_t;",
				Expected: []sql.Row{},
			},
			{
				Query:    "ANALYZE (SKIP_LOCKED false, VERBOSE false) an_t, an_u (id);",
				Expected: []sql.Row{},
			},
			{
				Query:    "ANALYZE VERBOSE an_t;",
				Expected: []sql.Row{},
			},
			{
				Query:    "ANALYZE an_t (version);",
				Expected: []sql.Row{},
			},
			{
				Query:    "ANALYSE (BUFFER_USAGE_LIMIT '256kB') an_t;",
				Expected: []sql.Row{},
			},
			{
				Query:    "ANALYZE (BUFFER_USAGE_LIMIT 1024) an_t;",
				Expected: []sql.Row{},
			},
			{
				Query:    "ANALYZE (SKIP_LOCKED);",
				Expected: []sql.Row{},
			},
			{
				Query:       "ANALYZE (FULL) an_t;",
				ExpectedErr: `unrecognized ANALYZE option "full"`,
			},
			{
				Query:       "ANALYZE (FREEZE) an_t;",
				ExpectedErr: `unrecognized ANALYZE option "freeze"`,
			},
			{
				Query:       "ANALYZE an_t, nope;",
				ExpectedErr: "nope",
			},
			{
				Query:       "ANALYZE an_t (no_such_col);",
				ExpectedErr: `column "no_such_col" of relation "an_t" does not exist`,
				Skip:        true, // column lists are not validated
			},
			{
				Query:       "ANALYZE an_t (id, id);",
				ExpectedErr: `column "id" of relation "an_t" appears more than once`,
				Skip:        true, // column lists are not validated
			},
		},
	},
}
