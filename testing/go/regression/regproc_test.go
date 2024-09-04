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

func TestRegproc(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_regproc)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_regproc,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `/* If objects exist, return oids */
CREATE ROLE regress_regrole_test;`,
			},
			{
				Statement: `SELECT regoper('||/');`,
				Results:   []sql.Row{{`||/`}},
			},
			{
				Statement: `SELECT regoperator('+(int4,int4)');`,
				Results:   []sql.Row{{`+(integer,integer)`}},
			},
			{
				Statement: `SELECT regproc('now');`,
				Results:   []sql.Row{{`now`}},
			},
			{
				Statement: `SELECT regprocedure('abs(numeric)');`,
				Results:   []sql.Row{{`abs(numeric)`}},
			},
			{
				Statement: `SELECT regclass('pg_class');`,
				Results:   []sql.Row{{`pg_class`}},
			},
			{
				Statement: `SELECT regtype('int4');`,
				Results:   []sql.Row{{`integer`}},
			},
			{
				Statement: `SELECT regcollation('"POSIX"');`,
				Results:   []sql.Row{{"POSIX"}},
			},
			{
				Statement: `SELECT to_regoper('||/');`,
				Results:   []sql.Row{{`||/`}},
			},
			{
				Statement: `SELECT to_regoperator('+(int4,int4)');`,
				Results:   []sql.Row{{`+(integer,integer)`}},
			},
			{
				Statement: `SELECT to_regproc('now');`,
				Results:   []sql.Row{{`now`}},
			},
			{
				Statement: `SELECT to_regprocedure('abs(numeric)');`,
				Results:   []sql.Row{{`abs(numeric)`}},
			},
			{
				Statement: `SELECT to_regclass('pg_class');`,
				Results:   []sql.Row{{`pg_class`}},
			},
			{
				Statement: `SELECT to_regtype('int4');`,
				Results:   []sql.Row{{`integer`}},
			},
			{
				Statement: `SELECT to_regcollation('"POSIX"');`,
				Results:   []sql.Row{{"POSIX"}},
			},
			{
				Statement: `SELECT regoper('pg_catalog.||/');`,
				Results:   []sql.Row{{`||/`}},
			},
			{
				Statement: `SELECT regoperator('pg_catalog.+(int4,int4)');`,
				Results:   []sql.Row{{`+(integer,integer)`}},
			},
			{
				Statement: `SELECT regproc('pg_catalog.now');`,
				Results:   []sql.Row{{`now`}},
			},
			{
				Statement: `SELECT regprocedure('pg_catalog.abs(numeric)');`,
				Results:   []sql.Row{{`abs(numeric)`}},
			},
			{
				Statement: `SELECT regclass('pg_catalog.pg_class');`,
				Results:   []sql.Row{{`pg_class`}},
			},
			{
				Statement: `SELECT regtype('pg_catalog.int4');`,
				Results:   []sql.Row{{`integer`}},
			},
			{
				Statement: `SELECT regcollation('pg_catalog."POSIX"');`,
				Results:   []sql.Row{{"POSIX"}},
			},
			{
				Statement: `SELECT to_regoper('pg_catalog.||/');`,
				Results:   []sql.Row{{`||/`}},
			},
			{
				Statement: `SELECT to_regproc('pg_catalog.now');`,
				Results:   []sql.Row{{`now`}},
			},
			{
				Statement: `SELECT to_regprocedure('pg_catalog.abs(numeric)');`,
				Results:   []sql.Row{{`abs(numeric)`}},
			},
			{
				Statement: `SELECT to_regclass('pg_catalog.pg_class');`,
				Results:   []sql.Row{{`pg_class`}},
			},
			{
				Statement: `SELECT to_regtype('pg_catalog.int4');`,
				Results:   []sql.Row{{`integer`}},
			},
			{
				Statement: `SELECT to_regcollation('pg_catalog."POSIX"');`,
				Results:   []sql.Row{{"POSIX"}},
			},
			{
				Statement: `SELECT regrole('regress_regrole_test');`,
				Results:   []sql.Row{{`regress_regrole_test`}},
			},
			{
				Statement: `SELECT regrole('"regress_regrole_test"');`,
				Results:   []sql.Row{{`regress_regrole_test`}},
			},
			{
				Statement: `SELECT regnamespace('pg_catalog');`,
				Results:   []sql.Row{{`pg_catalog`}},
			},
			{
				Statement: `SELECT regnamespace('"pg_catalog"');`,
				Results:   []sql.Row{{`pg_catalog`}},
			},
			{
				Statement: `SELECT to_regrole('regress_regrole_test');`,
				Results:   []sql.Row{{`regress_regrole_test`}},
			},
			{
				Statement: `SELECT to_regrole('"regress_regrole_test"');`,
				Results:   []sql.Row{{`regress_regrole_test`}},
			},
			{
				Statement: `SELECT to_regnamespace('pg_catalog');`,
				Results:   []sql.Row{{`pg_catalog`}},
			},
			{
				Statement: `SELECT to_regnamespace('"pg_catalog"');`,
				Results:   []sql.Row{{`pg_catalog`}},
			},
			{
				Statement: `/* If objects don't exist, raise errors. */
DROP ROLE regress_regrole_test;`,
			},
			{
				Statement:   `SELECT regoper('||//');`,
				ErrorString: `operator does not exist: ||//`,
			},
			{
				Statement:   `SELECT regoperator('++(int4,int4)');`,
				ErrorString: `operator does not exist: ++(int4,int4)`,
			},
			{
				Statement:   `SELECT regproc('know');`,
				ErrorString: `function "know" does not exist`,
			},
			{
				Statement:   `SELECT regprocedure('absinthe(numeric)');`,
				ErrorString: `function "absinthe(numeric)" does not exist`,
			},
			{
				Statement:   `SELECT regclass('pg_classes');`,
				ErrorString: `relation "pg_classes" does not exist`,
			},
			{
				Statement:   `SELECT regtype('int3');`,
				ErrorString: `type "int3" does not exist`,
			},
			{
				Statement:   `SELECT regoper('ng_catalog.||/');`,
				ErrorString: `schema "ng_catalog" does not exist`,
			},
			{
				Statement:   `SELECT regoperator('ng_catalog.+(int4,int4)');`,
				ErrorString: `operator does not exist: ng_catalog.+(int4,int4)`,
			},
			{
				Statement:   `SELECT regproc('ng_catalog.now');`,
				ErrorString: `schema "ng_catalog" does not exist`,
			},
			{
				Statement:   `SELECT regprocedure('ng_catalog.abs(numeric)');`,
				ErrorString: `schema "ng_catalog" does not exist`,
			},
			{
				Statement:   `SELECT regclass('ng_catalog.pg_class');`,
				ErrorString: `schema "ng_catalog" does not exist`,
			},
			{
				Statement:   `SELECT regtype('ng_catalog.int4');`,
				ErrorString: `schema "ng_catalog" does not exist`,
			},
			{
				Statement:   `SELECT regcollation('ng_catalog."POSIX"');`,
				ErrorString: `schema "ng_catalog" does not exist`,
			},
			{
				Statement:   `SELECT regrole('regress_regrole_test');`,
				ErrorString: `role "regress_regrole_test" does not exist`,
			},
			{
				Statement:   `SELECT regrole('"regress_regrole_test"');`,
				ErrorString: `role "regress_regrole_test" does not exist`,
			},
			{
				Statement:   `SELECT regrole('Nonexistent');`,
				ErrorString: `role "nonexistent" does not exist`,
			},
			{
				Statement:   `SELECT regrole('"Nonexistent"');`,
				ErrorString: `role "Nonexistent" does not exist`,
			},
			{
				Statement:   `SELECT regrole('foo.bar');`,
				ErrorString: `invalid name syntax`,
			},
			{
				Statement:   `SELECT regnamespace('Nonexistent');`,
				ErrorString: `schema "nonexistent" does not exist`,
			},
			{
				Statement:   `SELECT regnamespace('"Nonexistent"');`,
				ErrorString: `schema "Nonexistent" does not exist`,
			},
			{
				Statement:   `SELECT regnamespace('foo.bar');`,
				ErrorString: `invalid name syntax`,
			},
			{
				Statement: `/* If objects don't exist, return NULL with no error. */
SELECT to_regoper('||//');`,
				Results: []sql.Row{{``}},
			},
			{
				Statement: `SELECT to_regoperator('++(int4,int4)');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT to_regproc('know');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT to_regprocedure('absinthe(numeric)');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT to_regclass('pg_classes');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT to_regtype('int3');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT to_regcollation('notacollation');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT to_regoper('ng_catalog.||/');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT to_regoperator('ng_catalog.+(int4,int4)');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT to_regproc('ng_catalog.now');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT to_regprocedure('ng_catalog.abs(numeric)');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT to_regclass('ng_catalog.pg_class');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT to_regtype('ng_catalog.int4');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT to_regcollation('ng_catalog."POSIX"');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT to_regrole('regress_regrole_test');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT to_regrole('"regress_regrole_test"');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement:   `SELECT to_regrole('foo.bar');`,
				ErrorString: `invalid name syntax`,
			},
			{
				Statement: `SELECT to_regrole('Nonexistent');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT to_regrole('"Nonexistent"');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement:   `SELECT to_regrole('foo.bar');`,
				ErrorString: `invalid name syntax`,
			},
			{
				Statement: `SELECT to_regnamespace('Nonexistent');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT to_regnamespace('"Nonexistent"');`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement:   `SELECT to_regnamespace('foo.bar');`,
				ErrorString: `invalid name syntax`,
			},
		},
	})
}
