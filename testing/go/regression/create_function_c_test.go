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

package regression

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestCreateFunctionC(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_create_function_c)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_create_function_c,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `\getenv libdir PG_LIBDIR
\getenv dlsuffix PG_DLSUFFIX
\set regresslib :libdir '/regress' :dlsuffix
LOAD :'regresslib';`,
			},
			{
				Statement: `CREATE FUNCTION test1 (int) RETURNS int LANGUAGE C
    AS 'nosuchfile';`,
				ErrorString: `could not access file "nosuchfile": No such file or directory`,
			},
			{
				Statement: `\set VERBOSITY sqlstate
CREATE FUNCTION test1 (int) RETURNS int LANGUAGE C
    AS :'regresslib', 'nosuchsymbol';`,
				ErrorString: `42883`,
			},
			{
				Statement: `\set VERBOSITY default
SELECT regexp_replace(:'LAST_ERROR_MESSAGE', 'file ".*"', 'file "..."');`,
				Results: []sql.Row{{`could not find function "nosuchsymbol" in file "..."`}},
			},
			{
				Statement: `CREATE FUNCTION test1 (int) RETURNS int LANGUAGE internal
    AS 'nosuch';`,
				ErrorString: `there is no built-in function named "nosuch"`,
			},
		},
	})
}
