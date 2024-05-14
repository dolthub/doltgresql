// Copyright 2023 Dolthub, Inc.
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

func TestIssues(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Issue #25",
			SetUpScript: []string{
				"create table tbl (pk int);",
				"insert into tbl values (1);",
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:       `call dolt_add(".");`,
					ExpectedErr: "abc",
				},
				{
					Query:            `call dolt_add('.');`,
					SkipResultsCheck: true,
				},
				{
					Query:       `call dolt_commit("-m", "look ma");`,
					ExpectedErr: "abc",
				},
				{
					Query:            `call dolt_commit('-m', 'look ma');`,
					SkipResultsCheck: true,
				},
				{
					Query:       `call dolt_branch("br1");`,
					ExpectedErr: "abc",
				},
				{
					Query:            `call dolt_branch('br1');`,
					SkipResultsCheck: true,
				},
			},
		},
	})
}
