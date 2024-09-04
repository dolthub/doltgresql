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

func TestUuid(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_uuid)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_uuid,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE guid1
(
	guid_field UUID,
	text_field TEXT DEFAULT(now())
);`,
			},
			{
				Statement: `CREATE TABLE guid2
(
	guid_field UUID,
	text_field TEXT DEFAULT(now())
);`,
			},
			{
				Statement:   `INSERT INTO guid1(guid_field) VALUES('11111111-1111-1111-1111-111111111111F');`,
				ErrorString: `invalid input syntax for type uuid: "11111111-1111-1111-1111-111111111111F"`,
			},
			{
				Statement:   `INSERT INTO guid1(guid_field) VALUES('{11111111-1111-1111-1111-11111111111}');`,
				ErrorString: `invalid input syntax for type uuid: "{11111111-1111-1111-1111-11111111111}"`,
			},
			{
				Statement:   `INSERT INTO guid1(guid_field) VALUES('111-11111-1111-1111-1111-111111111111');`,
				ErrorString: `invalid input syntax for type uuid: "111-11111-1111-1111-1111-111111111111"`,
			},
			{
				Statement:   `INSERT INTO guid1(guid_field) VALUES('{22222222-2222-2222-2222-222222222222 ');`,
				ErrorString: `invalid input syntax for type uuid: "{22222222-2222-2222-2222-222222222222 "`,
			},
			{
				Statement:   `INSERT INTO guid1(guid_field) VALUES('11111111-1111-1111-G111-111111111111');`,
				ErrorString: `invalid input syntax for type uuid: "11111111-1111-1111-G111-111111111111"`,
			},
			{
				Statement:   `INSERT INTO guid1(guid_field) VALUES('11+11111-1111-1111-1111-111111111111');`,
				ErrorString: `invalid input syntax for type uuid: "11+11111-1111-1111-1111-111111111111"`,
			},
			{
				Statement: `INSERT INTO guid1(guid_field) VALUES('11111111-1111-1111-1111-111111111111');`,
			},
			{
				Statement: `INSERT INTO guid1(guid_field) VALUES('{22222222-2222-2222-2222-222222222222}');`,
			},
			{
				Statement: `INSERT INTO guid1(guid_field) VALUES('3f3e3c3b3a3039383736353433a2313e');`,
			},
			{
				Statement: `SELECT guid_field FROM guid1;`,
				Results:   []sql.Row{{`11111111-1111-1111-1111-111111111111`}, {`22222222-2222-2222-2222-222222222222`}, {`3f3e3c3b-3a30-3938-3736-353433a2313e`}},
			},
			{
				Statement: `SELECT guid_field FROM guid1 ORDER BY guid_field ASC;`,
				Results:   []sql.Row{{`11111111-1111-1111-1111-111111111111`}, {`22222222-2222-2222-2222-222222222222`}, {`3f3e3c3b-3a30-3938-3736-353433a2313e`}},
			},
			{
				Statement: `SELECT guid_field FROM guid1 ORDER BY guid_field DESC;`,
				Results:   []sql.Row{{`3f3e3c3b-3a30-3938-3736-353433a2313e`}, {`22222222-2222-2222-2222-222222222222`}, {`11111111-1111-1111-1111-111111111111`}},
			},
			{
				Statement: `SELECT COUNT(*) FROM guid1 WHERE guid_field = '3f3e3c3b-3a30-3938-3736-353433a2313e';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT COUNT(*) FROM guid1 WHERE guid_field <> '11111111111111111111111111111111';`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT COUNT(*) FROM guid1 WHERE guid_field < '22222222-2222-2222-2222-222222222222';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT COUNT(*) FROM guid1 WHERE guid_field <= '22222222-2222-2222-2222-222222222222';`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT COUNT(*) FROM guid1 WHERE guid_field > '22222222-2222-2222-2222-222222222222';`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `SELECT COUNT(*) FROM guid1 WHERE guid_field >= '22222222-2222-2222-2222-222222222222';`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `CREATE INDEX guid1_btree ON guid1 USING BTREE (guid_field);`,
			},
			{
				Statement: `CREATE INDEX guid1_hash  ON guid1 USING HASH  (guid_field);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX guid1_unique_BTREE ON guid1 USING BTREE (guid_field);`,
			},
			{
				Statement:   `INSERT INTO guid1(guid_field) VALUES('11111111-1111-1111-1111-111111111111');`,
				ErrorString: `duplicate key value violates unique constraint "guid1_unique_btree"`,
			},
			{
				Statement: `SELECT count(*) FROM pg_class WHERE relkind='i' AND relname LIKE 'guid%';`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `INSERT INTO guid1(guid_field) VALUES('44444444-4444-4444-4444-444444444444');`,
			},
			{
				Statement: `INSERT INTO guid2(guid_field) VALUES('11111111-1111-1111-1111-111111111111');`,
			},
			{
				Statement: `INSERT INTO guid2(guid_field) VALUES('{22222222-2222-2222-2222-222222222222}');`,
			},
			{
				Statement: `INSERT INTO guid2(guid_field) VALUES('3f3e3c3b3a3039383736353433a2313e');`,
			},
			{
				Statement: `SELECT COUNT(*) FROM guid1 g1 INNER JOIN guid2 g2 ON g1.guid_field = g2.guid_field;`,
				Results:   []sql.Row{{3}},
			},
			{
				Statement: `SELECT COUNT(*) FROM guid1 g1 LEFT JOIN guid2 g2 ON g1.guid_field = g2.guid_field WHERE g2.guid_field IS NULL;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `TRUNCATE guid1;`,
			},
			{
				Statement: `INSERT INTO guid1 (guid_field) VALUES (gen_random_uuid());`,
			},
			{
				Statement: `INSERT INTO guid1 (guid_field) VALUES (gen_random_uuid());`,
			},
			{
				Statement: `SELECT count(DISTINCT guid_field) FROM guid1;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `DROP TABLE guid1, guid2 CASCADE;`,
			},
		},
	})
}
