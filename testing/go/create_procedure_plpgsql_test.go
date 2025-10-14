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

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestCreateProcedureLanguagePlpgsql(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Smoke test",
			SetUpScript: []string{
				`CREATE TABLE test (v1 INT8);`,
				`CREATE PROCEDURE example(input INT8) AS $$
				BEGIN
					INSERT INTO test VALUES (input);
				END;
				$$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "CALL example(1);",
					Expected: []sql.Row{},
				},
				{
					Query:    "CALL example('2');",
					Expected: []sql.Row{},
				},
				{
					Query: "SELECT * FROM test;",
					Expected: []sql.Row{
						{1},
						{2},
					},
				},
			},
		},
	})
}
