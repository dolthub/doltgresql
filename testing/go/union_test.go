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
	"github.com/dolthub/go-mysql-server/sql"
"testing"

	)

func TestUnion(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "union tests",
			SetUpScript: []string{
				`CREATE TABLE t1 (i INT PRIMARY KEY);`,
				`CREATE TABLE t2 (j INT PRIMARY KEY);`,
				`INSERT INTO t1 VALUES (1), (2), (3);`,
				`INSERT INTO t2 VALUES (2), (3), (4);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT * FROM t1 UNION SELECT * FROM t2;`,
					Expected: []sql.Row{
						{1},
						{2},
						{3},
						{4},
					},
				},
				{
					Query: `SELECT 123 UNION SELECT 456;`,
					Expected: []sql.Row{
						{123},
						{456},
					},
				},
				{
					Query: `SELECT * FROM (VALUES (123), (456)) a UNION SELECT * FROM (VALUES (456), (789)) b;`,
					Expected: []sql.Row{
						{123},
						{456},
						{789},
					},
				},
			},
		},
	})
}
