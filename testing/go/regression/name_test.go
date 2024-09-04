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

func TestName(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_name)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_name,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT name 'name string' = name 'name string' AS "True";`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT name 'name string' = name 'name string ' AS "False";`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `CREATE TABLE NAME_TBL(f1 name);`,
			},
			{
				Statement: `INSERT INTO NAME_TBL(f1) VALUES ('1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQR');`,
			},
			{
				Statement: `INSERT INTO NAME_TBL(f1) VALUES ('1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopqr');`,
			},
			{
				Statement: `INSERT INTO NAME_TBL(f1) VALUES ('asdfghjkl;');`,
			},
			{
				Statement: `INSERT INTO NAME_TBL(f1) VALUES ('343f%2a');`,
			},
			{
				Statement: `INSERT INTO NAME_TBL(f1) VALUES ('d34aaasdf');`,
			},
			{
				Statement: `INSERT INTO NAME_TBL(f1) VALUES ('');`,
			},
			{
				Statement: `INSERT INTO NAME_TBL(f1) VALUES ('1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ');`,
			},
			{
				Statement: `SELECT * FROM NAME_TBL;`,
				Results:   []sql.Row{{`1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQ`}, {`1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopq`}, {`asdfghjkl;`}, {`343f%2a`}, {`d34aaasdf`}, {``}, {`1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQ`}},
			},
			{
				Statement: `SELECT c.f1 FROM NAME_TBL c WHERE c.f1 <> '1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQR';`,
				Results:   []sql.Row{{`1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopq`}, {`asdfghjkl;`}, {`343f%2a`}, {`d34aaasdf`}, {``}},
			},
			{
				Statement: `SELECT c.f1 FROM NAME_TBL c WHERE c.f1 = '1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQR';`,
				Results:   []sql.Row{{`1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQ`}, {`1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQ`}},
			},
			{
				Statement: `SELECT c.f1 FROM NAME_TBL c WHERE c.f1 < '1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQR';`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `SELECT c.f1 FROM NAME_TBL c WHERE c.f1 <= '1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQR';`,
				Results:   []sql.Row{{`1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQ`}, {``}, {`1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQ`}},
			},
			{
				Statement: `SELECT c.f1 FROM NAME_TBL c WHERE c.f1 > '1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQR';`,
				Results:   []sql.Row{{`1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopq`}, {`asdfghjkl;`}, {`343f%2a`}, {`d34aaasdf`}},
			},
			{
				Statement: `SELECT c.f1 FROM NAME_TBL c WHERE c.f1 >= '1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQR';`,
				Results:   []sql.Row{{`1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQ`}, {`1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopq`}, {`asdfghjkl;`}, {`343f%2a`}, {`d34aaasdf`}, {`1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQ`}},
			},
			{
				Statement: `SELECT c.f1 FROM NAME_TBL c WHERE c.f1 ~ '.*';`,
				Results:   []sql.Row{{`1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQ`}, {`1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopq`}, {`asdfghjkl;`}, {`343f%2a`}, {`d34aaasdf`}, {``}, {`1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQ`}},
			},
			{
				Statement: `SELECT c.f1 FROM NAME_TBL c WHERE c.f1 !~ '.*';`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT c.f1 FROM NAME_TBL c WHERE c.f1 ~ '[0-9]';`,
				Results:   []sql.Row{{`1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQ`}, {`1234567890abcdefghijklmnopqrstuvwxyz1234567890abcdefghijklmnopq`}, {`343f%2a`}, {`d34aaasdf`}, {`1234567890ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890ABCDEFGHIJKLMNOPQ`}},
			},
			{
				Statement: `SELECT c.f1 FROM NAME_TBL c WHERE c.f1 ~ '.*asdf.*';`,
				Results:   []sql.Row{{`asdfghjkl;`}, {`d34aaasdf`}},
			},
			{
				Statement: `DROP TABLE NAME_TBL;`,
			},
			{
				Statement: `DO $$
DECLARE r text[];`,
			},
			{
				Statement: `BEGIN
  r := parse_ident('Schemax.Tabley');`,
			},
			{
				Statement: `  RAISE NOTICE '%', format('%I.%I', r[1], r[2]);`,
			},
			{
				Statement: `  r := parse_ident('"SchemaX"."TableY"');`,
			},
			{
				Statement: `  RAISE NOTICE '%', format('%I.%I', r[1], r[2]);`,
			},
			{
				Statement: `END;`,
			},
			{
				Statement: `$$;`,
			},
			{
				Statement: `SELECT parse_ident('foo.boo');`,
				Results:   []sql.Row{{`{foo,boo}`}},
			},
			{
				Statement:   `SELECT parse_ident('foo.boo[]'); -- should fail`,
				ErrorString: `string is not a valid identifier: "foo.boo[]"`,
			},
			{
				Statement: `SELECT parse_ident('foo.boo[]', strict => false); -- ok`,
				Results:   []sql.Row{{`{foo,boo}`}},
			},
			{
				Statement:   `SELECT parse_ident(' ');`,
				ErrorString: `string is not a valid identifier: " "`,
			},
			{
				Statement:   `SELECT parse_ident(' .aaa');`,
				ErrorString: `string is not a valid identifier: " .aaa"`,
			},
			{
				Statement:   `SELECT parse_ident(' aaa . ');`,
				ErrorString: `string is not a valid identifier: " aaa . "`,
			},
			{
				Statement:   `SELECT parse_ident('aaa.a%b');`,
				ErrorString: `string is not a valid identifier: "aaa.a%b"`,
			},
			{
				Statement:   `SELECT parse_ident(E'X\rXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX');`,
				ErrorString: `string is not a valid identifier: "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"`,
			},
			{
				Statement: `SELECT length(a[1]), length(a[2]) from parse_ident('"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx".yyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyyy') as a ;`,
				Results:   []sql.Row{{414, 289}},
			},
			{
				Statement: `SELECT parse_ident(' first . "  second  " ."   third   ". "  ' || repeat('x',66) || '"');`,
				Results:   []sql.Row{{`{first,"  second  ","   third   ","  xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}`}},
			},
			{
				Statement: `SELECT parse_ident(' first . "  second  " ."   third   ". "  ' || repeat('x',66) || '"')::name[];`,
				Results:   []sql.Row{{`{first,"  second  ","   third   ","  xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}`}},
			},
			{
				Statement:   `SELECT parse_ident(E'"c".X XXXX\002XXXXXX');`,
				ErrorString: `string is not a valid identifier: ""c".X XXXXXXXXXX"`, //lint:ignore ST1018 Should not modify test
			},
			{
				Statement:   `SELECT parse_ident('1020');`,
				ErrorString: `string is not a valid identifier: "1020"`,
			},
			{
				Statement:   `SELECT parse_ident('10.20');`,
				ErrorString: `string is not a valid identifier: "10.20"`,
			},
			{
				Statement:   `SELECT parse_ident('.');`,
				ErrorString: `string is not a valid identifier: "."`,
			},
			{
				Statement:   `SELECT parse_ident('.1020');`,
				ErrorString: `string is not a valid identifier: ".1020"`,
			},
			{
				Statement:   `SELECT parse_ident('xxx.1020');`,
				ErrorString: `string is not a valid identifier: "xxx.1020"`,
			},
		},
	})
}
