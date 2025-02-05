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

package functions

import (
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initInterpretedExamples registers example functions to the catalog. These are temporary, and exist solely to test the
// interpreter functionality.
func initInterpretedExamples() {
	framework.RegisterFunction(interpretedAssignment)
	framework.RegisterFunction(interpretedAlias)
}

// interpretedAssignment is roughly equivalent to the (expected) parsed output of the following function definition:
/* CREATE FUNCTION interpreted_assignment(input TEXT) RETURNS TEXT AS $$
*  DECLARE
*      var1 TEXT;
*  BEGIN
*      var1 := 'Initial: ' || input;
*      IF input = 'Hello' THEN
*          var1 := var1 || ' - Greeting';
*      ELSIF input = 'Bye' THEN
*          var1 := var1 || ' - Farewell';
*      ELSIF length(input) > 5 THEN
*          var1 := var1 || ' - Over 5';
*      ELSE
*          var1 := var1 || ' - Else';
*      END IF;
*      RETURN var1;
*  END;
*  $$ LANGUAGE plpgsql;
 */
var interpretedAssignment = framework.InterpretedFunction{
	ID:                 id.NewFunction("pg_catalog", "interpreted_assignment", pgtypes.Text.ID),
	ReturnType:         pgtypes.Text,
	ParameterNames:     []string{"input"},
	ParameterTypes:     []*pgtypes.DoltgresType{pgtypes.Text},
	Variadic:           false,
	IsNonDeterministic: false,
	Strict:             true,
	Labels:             nil,
	Statements: []framework.InterpreterOperation{
		{ // 0
			OpCode: framework.OpCode_ScopeBegin,
		},
		{ // 1
			OpCode:      framework.OpCode_Declare,
			PrimaryData: `text`,
			Target:      `var1`,
		},
		{ // 2
			OpCode:        framework.OpCode_Assign,
			PrimaryData:   `SELECT 'Initial: ' || $1;`,
			SecondaryData: []string{`input`},
			Target:        `var1`,
		},
		{ // 3
			OpCode: framework.OpCode_ScopeBegin,
		},
		{ // 4
			OpCode:        framework.OpCode_If,
			PrimaryData:   `SELECT $1 = 'Hello';`,
			SecondaryData: []string{`input`},
			Index:         6,
		},
		{ // 5
			OpCode: framework.OpCode_Goto,
			Index:  8,
		},
		{ // 6
			OpCode:        framework.OpCode_Assign,
			PrimaryData:   `SELECT $1 || ' - Greeting';`,
			SecondaryData: []string{`var1`},
			Target:        `var1`,
		},
		{ // 7
			OpCode: framework.OpCode_Goto,
			Index:  17,
		},
		{ // 8
			OpCode:        framework.OpCode_If,
			PrimaryData:   `SELECT $1 = 'Bye';`,
			SecondaryData: []string{`input`},
			Index:         10,
		},
		{ // 9
			OpCode: framework.OpCode_Goto,
			Index:  12,
		},
		{ // 10
			OpCode:        framework.OpCode_Assign,
			PrimaryData:   `SELECT $1 || ' - Farewell';`,
			SecondaryData: []string{`var1`},
			Target:        `var1`,
		},
		{ // 11
			OpCode: framework.OpCode_Goto,
			Index:  17,
		},
		{ // 12
			OpCode:        framework.OpCode_If,
			PrimaryData:   `SELECT length($1) > 5;`,
			SecondaryData: []string{`input`},
			Index:         14,
		},
		{ // 13
			OpCode: framework.OpCode_Goto,
			Index:  16,
		},
		{ // 14
			OpCode:        framework.OpCode_Assign,
			PrimaryData:   `SELECT $1 || ' - Over 5';`,
			SecondaryData: []string{`var1`},
			Target:        `var1`,
		},
		{ // 15
			OpCode: framework.OpCode_Goto,
			Index:  17,
		},
		{ // 16
			OpCode:        framework.OpCode_Assign,
			PrimaryData:   `SELECT $1 || ' - Else';`,
			SecondaryData: []string{`var1`},
			Target:        `var1`,
		},
		{ // 17
			OpCode: framework.OpCode_ScopeEnd,
		},
		{ // 18
			OpCode:        framework.OpCode_Return,
			PrimaryData:   `SELECT $1;`,
			SecondaryData: []string{`var1`},
		},
		{ // 19
			OpCode: framework.OpCode_ScopeEnd,
		},
	},
}

// interpretedAlias is roughly equivalent to the (expected) parsed output of the following function definition:
/*
CREATE FUNCTION interpreted_alias(input TEXT)
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
*/
var interpretedAlias = framework.InterpretedFunction{
	ID:                 id.NewFunction("pg_catalog", "interpreted_alias", pgtypes.Text.ID),
	ReturnType:         pgtypes.Text,
	ParameterNames:     []string{"input"},
	ParameterTypes:     []*pgtypes.DoltgresType{pgtypes.Text},
	Variadic:           false,
	IsNonDeterministic: false,
	Strict:             true,
	Labels:             nil,
	Statements: []framework.InterpreterOperation{
		{ // 0
			OpCode: framework.OpCode_ScopeBegin,
		},
		{ // 1
			OpCode:      framework.OpCode_Declare,
			Target:      `var1`,
			PrimaryData: `text`,
		},
		{ // 2
			OpCode:      framework.OpCode_Declare,
			Target:      `var2`,
			PrimaryData: `text`,
		},
		{ // 3
			OpCode: framework.OpCode_ScopeBegin,
		},
		{ // 4
			OpCode:      framework.OpCode_Alias,
			Target:      `alias1`,
			PrimaryData: `var1`,
		},
		{ // 5
			OpCode:      framework.OpCode_Alias,
			Target:      `alias2`,
			PrimaryData: `alias1`,
		},
		{ // 6
			OpCode:      framework.OpCode_Alias,
			Target:      `alias3`,
			PrimaryData: `input`,
		},
		{ // 7
			OpCode: framework.OpCode_ScopeBegin,
		},
		{ // 8
			OpCode:        framework.OpCode_Assign,
			PrimaryData:   `SELECT $1;`,
			SecondaryData: []string{`alias3`},
			Target:        `alias2`,
		},
		{ // 9
			OpCode: framework.OpCode_ScopeEnd,
		},
		{ // 10
			OpCode:        framework.OpCode_Return,
			PrimaryData:   `SELECT $1;`,
			SecondaryData: []string{`var1`},
		},
		{ // 11
			OpCode: framework.OpCode_ScopeEnd,
		},
	},
}
