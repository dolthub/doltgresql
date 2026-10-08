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

import "testing"

func TestDropDatabase(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "simple create database",
			Assertions: []ScriptTestAssertion{
				{
					Query: "CREATE DATABASE testdb",
				},
				{
					Query: "DROP DATABASE testdb",
				},
			},
		},
		{
			Name: "with quotes",
			Assertions: []ScriptTestAssertion{
				{
					Query: "CREATE DATABASE testdb",
				},
				{
					Query: "DROP DATABASE \"testdb\"",
				},
			},
		},
		{
			Name: "with hyphen",
			Assertions: []ScriptTestAssertion{
				{
					Query: "CREATE DATABASE \"test-db\"",
				},
				{
					Query: "USE \"test-db\"",
				},
			},
		},
		{
			Name: "drop a database that was previously used in the session",
			Assertions: []ScriptTestAssertion{
				{
					Query: "CREATE DATABASE dropdb1",
				},
				{
					Query: "CREATE DATABASE dropdb2",
				},
				{
					Query: "USE dropdb2",
				},
				{
					Query: "CREATE TABLE t1 (a INT PRIMARY KEY)",
				},
				{
					Query: "USE dropdb1",
				},
				{
					Query: "DROP DATABASE dropdb2",
				},
			},
		},
		{
			Name: "if exists",
			Assertions: []ScriptTestAssertion{
				{
					Query: "DROP DATABASE IF EXISTS invalid",
				},
			},
		},
	})
}
