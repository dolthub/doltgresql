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
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

// TestPartialIndexMergeJoinDoesNotLoseRows verifies that a partial index is not
// treated as a complete input when its leading column is used as a join key.
// Regression test for https://github.com/dolthub/doltgresql/issues/3100.
func TestPartialIndexMergeJoinDoesNotLoseRows(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "partial index on join key",
			SetUpScript: []string{
				"CREATE TABLE profiles (id int PRIMARY KEY);",
				"CREATE TABLE favorites (id int PRIMARY KEY, profile_id int NOT NULL, flag boolean NOT NULL DEFAULT false);",
				"CREATE TABLE profile_ranges (id int PRIMARY KEY, min_id int NOT NULL, max_id int NOT NULL);",
				"CREATE INDEX favorites_partial ON favorites (profile_id) WHERE flag;",
				"CREATE INDEX profile_ranges_min ON profile_ranges (min_id);",
				"CREATE INDEX profile_ranges_max ON profile_ranges (max_id);",
				"INSERT INTO profiles VALUES (1), (2), (3), (4), (5), (6), (7);",
				"INSERT INTO favorites (id, profile_id) SELECT id, id FROM profiles;",
				"UPDATE favorites SET flag = true WHERE profile_id IN (1, 7);",
				"INSERT INTO profile_ranges VALUES (1, 1, 7);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT count(*) FROM favorites f JOIN profiles p ON p.id = f.profile_id;",
					Expected: []sql.Row{{7}},
				},
				{
					Query:    "SELECT count(*) FROM favorites f LEFT JOIN profiles p ON p.id = f.profile_id;",
					Expected: []sql.Row{{7}},
				},
				{
					Query:    "SELECT count(*), count(f.id) FROM profiles p LEFT JOIN favorites f ON f.profile_id = p.id;",
					Expected: []sql.Row{{7, 7}},
				},
				{
					Query:    "SELECT count(*) FROM profiles p WHERE EXISTS (SELECT 1 FROM favorites f WHERE f.profile_id = p.id);",
					Expected: []sql.Row{{7}},
				},
				{
					Query:    "SELECT count(*) FROM profiles p WHERE NOT EXISTS (SELECT 1 FROM favorites f WHERE f.profile_id = p.id);",
					Expected: []sql.Row{{0}},
				},
				{
					Query:    "SELECT count(*) FROM favorites f JOIN profile_ranges r ON f.profile_id BETWEEN r.min_id AND r.max_id;",
					Expected: []sql.Row{{7}},
				},
				{
					Query:    "SELECT count(*) FROM favorites f JOIN profiles p ON p.id = f.profile_id WHERE f.flag;",
					Expected: []sql.Row{{2}},
				},
			},
		},
		{
			Name: "partial index created after rows exist",
			SetUpScript: []string{
				"CREATE TABLE existing_profiles (id int PRIMARY KEY);",
				"CREATE TABLE existing_favorites (id int PRIMARY KEY, profile_id int NOT NULL, flag boolean NOT NULL DEFAULT false);",
				"INSERT INTO existing_profiles VALUES (1), (2), (3), (4), (5), (6), (7);",
				"INSERT INTO existing_favorites (id, profile_id) SELECT id, id FROM existing_profiles;",
				"UPDATE existing_favorites SET flag = true WHERE profile_id IN (1, 7);",
				"CREATE INDEX existing_favorites_partial ON existing_favorites (profile_id) WHERE flag;",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT count(*) FROM existing_favorites f JOIN existing_profiles p ON p.id = f.profile_id;",
					Expected: []sql.Row{{7}},
				},
				{
					Query:    "SELECT count(*) FROM existing_favorites f JOIN existing_profiles p ON p.id = f.profile_id WHERE f.flag;",
					Expected: []sql.Row{{2}},
				},
			},
		},
	})
}
