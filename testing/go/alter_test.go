// Copyright 2025 Dolthub, Inc.
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

import "testing"

// TestAlterStatements tests ALTER statements other than ALTER TABLE, mostly for stub functionality.
// These tests should move into their respective test files as real functionality for these ALTER statements is added.
func TestAlterStatements(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "alter database",
			Assertions: []ScriptTestAssertion{
				{
					Query: "ALTER DATABASE postgres OWNER TO foo",
					ExpectedNotices: []ExpectedNotice{
						{
							Severity: "WARNING",
							Message:  "owners are not yet supported",
						},
					},
				},
			},
		},
		{
			Name: "alter sequence",
			SetUpScript: []string{
				"CREATE SEQUENCE testseq",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "ALTER SEQUENCE testseq OWNER TO foo",
					ExpectedNotices: []ExpectedNotice{
						{
							Severity: "WARNING",
							Message:  "owners are not yet supported",
						},
					},
				},
			},
		},
		{
			Name: "alter type",
			SetUpScript: []string{
				"CREATE TYPE testtype AS ENUM ('a', 'b', 'c')",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "ALTER TYPE testtype OWNER TO foo",
					ExpectedNotices: []ExpectedNotice{
						{
							Severity: "WARNING",
							Message:  "owners are not yet supported",
						},
					},
				},
			},
		},
		{
			Name: "alter function",
			SetUpScript: []string{
				"CREATE FUNCTION testfunc() RETURNS int AS $$ BEGIN RETURN 1; END; $$ LANGUAGE plpgsql",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "ALTER FUNCTION testfunc() OWNER TO foo",
					ExpectedNotices: []ExpectedNotice{
						{
							Severity: "WARNING",
							Message:  "owners are not yet supported",
						},
					},
				},
			},
		},
		{
			Name: "alter schema",
			SetUpScript: []string{
				"CREATE SCHEMA testschema",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "ALTER schema testschema OWNER TO foo",
					ExpectedNotices: []ExpectedNotice{
						{
							Severity: "WARNING",
							Message:  "owners are not yet supported",
						},
					},
				},
			},
		},
		{
			Name: "alter view",
			SetUpScript: []string{
				"CREATE VIEW testview AS SELECT 1",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: "ALTER VIEW testview OWNER TO foo",
					ExpectedNotices: []ExpectedNotice{
						{
							Severity: "WARNING",
							Message:  "owners are not yet supported",
						},
					},
				},
			},
		},
	})
}
