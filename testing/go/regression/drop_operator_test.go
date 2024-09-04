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

func TestDropOperator(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_drop_operator)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_drop_operator,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE OPERATOR === (
        PROCEDURE = int8eq,
        LEFTARG = bigint,
        RIGHTARG = bigint,
        COMMUTATOR = ===
);`,
			},
			{
				Statement: `CREATE OPERATOR !== (
        PROCEDURE = int8ne,
        LEFTARG = bigint,
        RIGHTARG = bigint,
        NEGATOR = ===,
        COMMUTATOR = !==
);`,
			},
			{
				Statement: `DROP OPERATOR !==(bigint, bigint);`,
			},
			{
				Statement: `SELECT  ctid, oprcom
FROM    pg_catalog.pg_operator fk
WHERE   oprcom != 0 AND
        NOT EXISTS(SELECT 1 FROM pg_catalog.pg_operator pk WHERE pk.oid = fk.oprcom);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT  ctid, oprnegate
FROM    pg_catalog.pg_operator fk
WHERE   oprnegate != 0 AND
        NOT EXISTS(SELECT 1 FROM pg_catalog.pg_operator pk WHERE pk.oid = fk.oprnegate);`,
				Results: []sql.Row{},
			},
			{
				Statement: `DROP OPERATOR ===(bigint, bigint);`,
			},
			{
				Statement: `CREATE OPERATOR <| (
        PROCEDURE = int8lt,
        LEFTARG = bigint,
        RIGHTARG = bigint
);`,
			},
			{
				Statement: `CREATE OPERATOR |> (
        PROCEDURE = int8gt,
        LEFTARG = bigint,
        RIGHTARG = bigint,
        NEGATOR = <|,
        COMMUTATOR = <|
);`,
			},
			{
				Statement: `DROP OPERATOR |>(bigint, bigint);`,
			},
			{
				Statement: `SELECT  ctid, oprcom
FROM    pg_catalog.pg_operator fk
WHERE   oprcom != 0 AND
        NOT EXISTS(SELECT 1 FROM pg_catalog.pg_operator pk WHERE pk.oid = fk.oprcom);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT  ctid, oprnegate
FROM    pg_catalog.pg_operator fk
WHERE   oprnegate != 0 AND
        NOT EXISTS(SELECT 1 FROM pg_catalog.pg_operator pk WHERE pk.oid = fk.oprnegate);`,
				Results: []sql.Row{},
			},
			{
				Statement: `DROP OPERATOR <|(bigint, bigint);`,
			},
		},
	})
}
