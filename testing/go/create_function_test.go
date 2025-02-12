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
			Name: "Assignment",
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
			Name: "SELECT INTO",
			SetUpScript: []string{`CREATE FUNCTION interpreted_select_into(input INT4) RETURNS TEXT AS $$
DECLARE
    ret TEXT;
	count INT4;
BEGIN
	DROP TABLE IF EXISTS temp_table;
    CREATE TABLE temp_table (pk SERIAL PRIMARY KEY, v1 TEXT NOT NULL);
    INSERT INTO temp_table (v1) VALUES ('abc'), ('def'), ('ghi');
	SELECT COUNT(*) INTO count FROM temp_table;
    IF input > 0 AND input <= count THEN
        SELECT v1 INTO ret FROM temp_table WHERE pk = input;
    ELSE
        ret := 'out of bounds';
    END IF;
    RETURN ret;
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query: "SELECT interpreted_select_into(1);",
					Expected: []sql.Row{
						{"abc"},
					},
				},
				{
					Query: "SELECT interpreted_select_into(2);",
					Expected: []sql.Row{
						{"def"},
					},
				},
				{
					Query: "SELECT interpreted_select_into(3);",
					Expected: []sql.Row{
						{"ghi"},
					},
				},
				{
					Query: "SELECT interpreted_select_into(4);",
					Expected: []sql.Row{
						{"out of bounds"},
					},
				},
			},
		},
		{
			Name: "ALIAS",
			// TODO: Implement OpCode conversion for parsed ALIAS statements.
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
		{
			Name: "PERFORM",
			SetUpScript: []string{
				`CREATE SEQUENCE test_sequence;`,
				`CREATE FUNCTION interpreted_perform() RETURNS VOID AS $$
BEGIN
	PERFORM nextval('test_sequence');
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT nextval('test_sequence');",
					Expected: []sql.Row{{1}},
				},
				{
					Query:    "SELECT interpreted_perform();",
					Expected: []sql.Row{{nil}}, // TODO: Postgres returns a value that's not null, but also not a value?
				},
				{
					Query:    "SELECT nextval('test_sequence');",
					Expected: []sql.Row{{3}},
				},
			},
		},
		{
			// TODO: Returning an integer seems to not be supported yet?
			Name: "WHILE",
			SetUpScript: []string{
				`CREATE FUNCTION interpreted_while(input INT4) RETURNS TEXT AS $$
DECLARE
	counter INT4;
BEGIN
	WHILE counter + input < 100 LOOP
		counter = counter + 1;
		counter = counter - 1;
		counter = counter + 1;
	END LOOP;
	RETURN counter::TEXT;
END;
$$ LANGUAGE plpgsql;`},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT interpreted_while(42);",
					Expected: []sql.Row{{"58"}},
				},
			},
		},
		{
			// Tests that variable names are correctly substituted with references
			// to the variables when the function is parsed.
			Name: "Variable reference substitution",
			SetUpScript: []string{`
CREATE FUNCTION test1(input TEXT) RETURNS TEXT AS $$
DECLARE
    var1 TEXT;
BEGIN
	var1 := 'input' || input;
	IF var1 = 'input' || input THEN
		RETURN var1 || 'var1';
	ELSE
		RETURN '!!!';
	END IF;
END;
$$ LANGUAGE plpgsql;`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query:    "SELECT test1('Hello');",
					Expected: []sql.Row{{"inputHellovar1"}},
				},
			},
		},
	})
}
