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

func TestExpressions(t *testing.T) {
	RunScriptsWithoutNormalization(t, []ScriptTest{
		{
			Name: "Any",
			SetUpScript: []string{
				`CREATE TABLE test (id INT);`,
				`INSERT INTO test VALUES (1), (3), (2);`,

				`CREATE TABLE test2 (id INT, test_id INT, txt text);`,
				`INSERT INTO test2 VALUES (1, 1, 'foo'), (2, 10, 'bar'), (3, 2, 'baz');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    `SELECT 3 = ANY (ARRAY[1, 2, 3, 4, 5]);`,
					Expected: []sql.Row{{"t"}},
				},
				{
					Query:       `SELECT 3 > ANY (ARRAY[1, 2, 3, 4, 5]);`,
					ExpectedErr: "ANY operator is not yet supported with suboperator >",
				},
				{
					Query:       `SELECT 6 < ANY (ARRAY[1, 2, 3, 4, 5]);`,
					ExpectedErr: "ANY operator is not yet supported with suboperator <",
				},
				{
					Query:    `SELECT * FROM test WHERE id = ANY(ARRAY[2, 3, 4, 5]);`,
					Expected: []sql.Row{{int32(3)}, {int32(2)}},
				},
				{
					Query:    `SELECT * FROM test WHERE id = ANY(ARRAY[4, 3, 2, 1, 0]);`,
					Expected: []sql.Row{{int32(1)}, {int32(3)}, {int32(2)}},
				},
				{
					Query:    `SELECT * FROM test WHERE id = ANY(ARRAY[4, 5, 6]);`,
					Expected: []sql.Row{},
				},
				{
					Query:    `SELECT * FROM test2 WHERE test_id = ANY(SELECT * FROM test WHERE id = 1);`,
					Expected: []sql.Row{{int32(1), int32(1), "foo"}},
				},
				{
					Skip:  true, // TODO: ERROR: the subquery returned more than 1 row
					Query: `SELECT * FROM test2 WHERE test_id > ANY(SELECT * FROM test);`,
					Expected: []sql.Row{
						{int32(2), int32(10), "foo"},
						{int32(3), int32(2), "baz"},
					},
				},
				{
					Skip:     true, // TODO: ERROR: the subquery returned more than 1 row
					Query:    `SELECT * FROM test2 WHERE test_id = ANY(SELECT * FROM test WHERE id > 1) AND txt = 'baz';`,
					Expected: []sql.Row{{int32(3), int32(2)}},
				},
				{
					Skip:  true, // TODO: ERROR: the subquery returned more than 1 row
					Query: `SELECT * FROM test2 WHERE test_id = ANY(SELECT * FROM test WHERE id > 0);`,
					Expected: []sql.Row{
						{int32(1), int32(1)},
						{int32(3), int32(2)},
					},
				},
			},
		},
	})
}
