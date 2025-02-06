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

func TestCreateFunction(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "Interpreter Assignment Example",
			Skip: true, // TODO: need to use a Doltgres function provider, as the current one doesn't allow for adding functions
			SetUpScript: []string{`CREATE FUNCTION interpreted_assignment(input TEXT) RETURNS TEXT AS $$
DECLARE
    var1 TEXT;
BEGIN
    var1 := 'Initial: ' || input;
    IF input = 'Hello' THEN
        var1 := var1 || ' - Greeting';
    ELSIF input = 'Bye' THEN
        var1 := var1 || ' - Farewell';
    ELSIF length(input) > 5 THEN
        var1 := var1 || ' - Over 5';
    ELSE
        var1 := var1 || ' - Else';
    END IF;
    RETURN var1;
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT interpreted_assignment('Hello');",
					Expected: []sql.Row{
						{"Initial: Hello - Greeting"},
					},
				},
				{
					Query: "SELECT interpreted_assignment('Bye');",
					Expected: []sql.Row{
						{"Initial: Bye - Farewell"},
					},
				},
				{
					Query: "SELECT interpreted_assignment('abc');",
					Expected: []sql.Row{
						{"Initial: abc - Else"},
					},
				},
				{
					Query: "SELECT interpreted_assignment('something');",
					Expected: []sql.Row{
						{"Initial: something - Over 5"},
					},
				},
			},
		},
		{
			Name: "Interpreter Alias Example",
			// TODO: need to use a Doltgres function provider, and need to implement the
			//       OpCode conversion for parsed ALIAS statements.
			Skip: true,
			SetUpScript: []string{
				`CREATE FUNCTION interpreted_alias(input TEXT)
				RETURNS TEXT AS $$
				DECLARE
					var1 TEXT;
					var2 TEXT;
				BEGIN
					DECLARE
						alias1 ALIAS FOR var1;
						alias2 ALIAS FOR alias1;
						alias3 ALIAS FOR input;
					BEGIN
						alias2 := alias3;
					END;
					RETURN var1;
				END;
				$$ LANGUAGE plpgsql;
				`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT interpreted_alias('123');",
					Expected: []sql.Row{{"123"}},
				},
			},
		},
	})
}
