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

func TestPublication(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_publication)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_publication,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE ROLE regress_publication_user LOGIN SUPERUSER;`,
			},
			{
				Statement: `CREATE ROLE regress_publication_user2;`,
			},
			{
				Statement: `CREATE ROLE regress_publication_user_dummy LOGIN NOSUPERUSER;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION 'regress_publication_user';`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_default;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `COMMENT ON PUBLICATION testpub_default IS 'test publication';`,
			},
			{
				Statement: `SELECT obj_description(p.oid, 'pg_publication') FROM pg_publication p;`,
				Results:   []sql.Row{{`test publication`}},
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpib_ins_trunct WITH (publish = insert);`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_default SET (publish = update);`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_xxx WITH (foo);`,
				ErrorString: `unrecognized publication parameter: "foo"`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_xxx WITH (publish = 'cluster, vacuum');`,
				ErrorString: `unrecognized value for publication option "publish": "cluster"`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_xxx WITH (publish_via_partition_root = 'true', publish_via_partition_root = '0');`,
				ErrorString: `conflicting or redundant options`,
			},
			{
				Statement: `\dRp
                                              List of publications
        Name        |          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------+--------------------------+------------+---------+---------+---------+-----------+----------
 testpib_ins_trunct | regress_publication_user | f          | t       | f       | f       | f         | f
 testpub_default    | regress_publication_user | f          | f       | t       | f       | f         | f
(2 rows)
ALTER PUBLICATION testpub_default SET (publish = 'insert, update, delete');`,
			},
			{
				Statement: `\dRp
                                              List of publications
        Name        |          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------+--------------------------+------------+---------+---------+---------+-----------+----------
 testpib_ins_trunct | regress_publication_user | f          | t       | f       | f       | f         | f
 testpub_default    | regress_publication_user | f          | t       | t       | t       | f         | f
(2 rows)
CREATE SCHEMA pub_test;`,
			},
			{
				Statement: `CREATE TABLE testpub_tbl1 (id serial primary key, data text);`,
			},
			{
				Statement: `CREATE TABLE pub_test.testpub_nopk (foo int, bar int);`,
			},
			{
				Statement: `CREATE VIEW testpub_view AS SELECT 1;`,
			},
			{
				Statement: `CREATE TABLE testpub_parted (a int) PARTITION BY LIST (a);`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_foralltables FOR ALL TABLES WITH (publish = 'insert');`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_foralltables SET (publish = 'insert, update');`,
			},
			{
				Statement: `CREATE TABLE testpub_tbl2 (id serial primary key, data text);`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub_foralltables ADD TABLE testpub_tbl2;`,
				ErrorString: `publication "testpub_foralltables" is defined as FOR ALL TABLES`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub_foralltables DROP TABLE testpub_tbl2;`,
				ErrorString: `publication "testpub_foralltables" is defined as FOR ALL TABLES`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub_foralltables SET TABLE pub_test.testpub_nopk;`,
				ErrorString: `publication "testpub_foralltables" is defined as FOR ALL TABLES`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub_foralltables ADD TABLES IN SCHEMA pub_test;`,
				ErrorString: `publication "testpub_foralltables" is defined as FOR ALL TABLES`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub_foralltables DROP TABLES IN SCHEMA pub_test;`,
				ErrorString: `publication "testpub_foralltables" is defined as FOR ALL TABLES`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub_foralltables SET TABLES IN SCHEMA pub_test;`,
				ErrorString: `publication "testpub_foralltables" is defined as FOR ALL TABLES`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_fortable FOR TABLE testpub_tbl1;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_fortable ADD TABLES IN SCHEMA pub_test;`,
			},
			{
				Statement: `\dRp+ testpub_fortable
                                Publication testpub_fortable
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables:
    "public.testpub_tbl1"
Tables from schemas:
    "pub_test"
ALTER PUBLICATION testpub_fortable DROP TABLES IN SCHEMA pub_test;`,
			},
			{
				Statement: `\dRp+ testpub_fortable
                                Publication testpub_fortable
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables:
    "public.testpub_tbl1"
ALTER PUBLICATION testpub_fortable SET TABLES IN SCHEMA pub_test;`,
			},
			{
				Statement: `\dRp+ testpub_fortable
                                Publication testpub_fortable
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "pub_test"
SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_forschema FOR TABLES IN SCHEMA pub_test;`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_for_tbl_schema FOR TABLES IN SCHEMA pub_test, TABLE pub_test.testpub_nopk;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `\dRp+ testpub_for_tbl_schema
                             Publication testpub_for_tbl_schema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables:
    "pub_test.testpub_nopk"
Tables from schemas:
    "pub_test"
CREATE PUBLICATION testpub_parsertst FOR TABLE pub_test.testpub_nopk, CURRENT_SCHEMA;`,
				ErrorString: `invalid table name`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_parsertst FOR TABLES IN SCHEMA foo, test.foo;`,
				ErrorString: `invalid schema name`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_forschema ADD TABLE pub_test.testpub_nopk;`,
			},
			{
				Statement: `\dRp+ testpub_forschema
                               Publication testpub_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables:
    "pub_test.testpub_nopk"
Tables from schemas:
    "pub_test"
ALTER PUBLICATION testpub_forschema DROP TABLE pub_test.testpub_nopk;`,
			},
			{
				Statement: `\dRp+ testpub_forschema
                               Publication testpub_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "pub_test"
ALTER PUBLICATION testpub_forschema DROP TABLE pub_test.testpub_nopk;`,
				ErrorString: `relation "testpub_nopk" is not part of the publication`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_forschema SET TABLE pub_test.testpub_nopk;`,
			},
			{
				Statement: `\dRp+ testpub_forschema
                               Publication testpub_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables:
    "pub_test.testpub_nopk"
SELECT pubname, puballtables FROM pg_publication WHERE pubname = 'testpub_foralltables';`,
				Results: []sql.Row{{`testpub_foralltables`, true}},
			},
			{
				Statement: `\d+ testpub_tbl2
                                                Table "public.testpub_tbl2"
 Column |  Type   | Collation | Nullable |                 Default                  | Storage  | Stats target | Description 
--------+---------+-----------+----------+------------------------------------------+----------+--------------+-------------
 id     | integer |           | not null | nextval('testpub_tbl2_id_seq'::regclass) | plain    |              | 
 data   | text    |           |          |                                          | extended |              | 
Indexes:
    "testpub_tbl2_pkey" PRIMARY KEY, btree (id)
Publications:
    "testpub_foralltables"
\dRp+ testpub_foralltables
                              Publication testpub_foralltables
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | t          | t       | t       | f       | f         | f
(1 row)
DROP TABLE testpub_tbl2;`,
			},
			{
				Statement: `DROP PUBLICATION testpub_foralltables, testpub_fortable, testpub_forschema, testpub_for_tbl_schema;`,
			},
			{
				Statement: `CREATE TABLE testpub_tbl3 (a int);`,
			},
			{
				Statement: `CREATE TABLE testpub_tbl3a (b text) INHERITS (testpub_tbl3);`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub3 FOR TABLE testpub_tbl3;`,
			},
			{
				Statement: `CREATE PUBLICATION testpub4 FOR TABLE ONLY testpub_tbl3;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `\dRp+ testpub3
                                    Publication testpub3
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables:
    "public.testpub_tbl3"
    "public.testpub_tbl3a"
\dRp+ testpub4
                                    Publication testpub4
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables:
    "public.testpub_tbl3"
DROP TABLE testpub_tbl3, testpub_tbl3a;`,
			},
			{
				Statement: `DROP PUBLICATION testpub3, testpub4;`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_forparted;`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_forparted1;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `CREATE TABLE testpub_parted1 (LIKE testpub_parted);`,
			},
			{
				Statement: `CREATE TABLE testpub_parted2 (LIKE testpub_parted);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_forparted1 SET (publish='insert');`,
			},
			{
				Statement: `ALTER TABLE testpub_parted ATTACH PARTITION testpub_parted1 FOR VALUES IN (1);`,
			},
			{
				Statement: `ALTER TABLE testpub_parted ATTACH PARTITION testpub_parted2 FOR VALUES IN (2);`,
			},
			{
				Statement: `UPDATE testpub_parted1 SET a = 1;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_forparted ADD TABLE testpub_parted;`,
			},
			{
				Statement: `\dRp+ testpub_forparted
                               Publication testpub_forparted
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables:
    "public.testpub_parted"
UPDATE testpub_parted SET a = 1 WHERE false;`,
			},
			{
				Statement:   `UPDATE testpub_parted1 SET a = 1;`,
				ErrorString: `cannot update table "testpub_parted1" because it does not have a replica identity and publishes updates`,
			},
			{
				Statement: `ALTER TABLE testpub_parted DETACH PARTITION testpub_parted1;`,
			},
			{
				Statement: `UPDATE testpub_parted1 SET a = 1;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_forparted SET (publish_via_partition_root = true);`,
			},
			{
				Statement: `\dRp+ testpub_forparted
                               Publication testpub_forparted
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | t
Tables:
    "public.testpub_parted"
UPDATE testpub_parted2 SET a = 2;`,
				ErrorString: `cannot update table "testpub_parted2" because it does not have a replica identity and publishes updates`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_forparted DROP TABLE testpub_parted;`,
			},
			{
				Statement: `UPDATE testpub_parted2 SET a = 2;`,
			},
			{
				Statement: `DROP TABLE testpub_parted1, testpub_parted2;`,
			},
			{
				Statement: `DROP PUBLICATION testpub_forparted, testpub_forparted1;`,
			},
			{
				Statement: `CREATE TABLE testpub_rf_tbl1 (a integer, b text);`,
			},
			{
				Statement: `CREATE TABLE testpub_rf_tbl2 (c text, d integer);`,
			},
			{
				Statement: `CREATE TABLE testpub_rf_tbl3 (e integer);`,
			},
			{
				Statement: `CREATE TABLE testpub_rf_tbl4 (g text);`,
			},
			{
				Statement: `CREATE TABLE testpub_rf_tbl5 (a xml);`,
			},
			{
				Statement: `CREATE SCHEMA testpub_rf_schema1;`,
			},
			{
				Statement: `CREATE TABLE testpub_rf_schema1.testpub_rf_tbl5 (h integer);`,
			},
			{
				Statement: `CREATE SCHEMA testpub_rf_schema2;`,
			},
			{
				Statement: `CREATE TABLE testpub_rf_schema2.testpub_rf_tbl6 (i integer);`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub5 FOR TABLE testpub_rf_tbl1, testpub_rf_tbl2 WHERE (c <> 'test' AND d < 5) WITH (publish = 'insert');`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `\dRp+ testpub5
                                    Publication testpub5
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | f       | f       | f         | f
Tables:
    "public.testpub_rf_tbl1"
    "public.testpub_rf_tbl2" WHERE ((c <> 'test'::text) AND (d < 5))
\d testpub_rf_tbl3
          Table "public.testpub_rf_tbl3"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 e      | integer |           |          | 
ALTER PUBLICATION testpub5 ADD TABLE testpub_rf_tbl3 WHERE (e > 1000 AND e < 2000);`,
			},
			{
				Statement: `\dRp+ testpub5
                                    Publication testpub5
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | f       | f       | f         | f
Tables:
    "public.testpub_rf_tbl1"
    "public.testpub_rf_tbl2" WHERE ((c <> 'test'::text) AND (d < 5))
    "public.testpub_rf_tbl3" WHERE ((e > 1000) AND (e < 2000))
\d testpub_rf_tbl3
          Table "public.testpub_rf_tbl3"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 e      | integer |           |          | 
Publications:
    "testpub5" WHERE ((e > 1000) AND (e < 2000))
ALTER PUBLICATION testpub5 DROP TABLE testpub_rf_tbl2;`,
			},
			{
				Statement: `\dRp+ testpub5
                                    Publication testpub5
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | f       | f       | f         | f
Tables:
    "public.testpub_rf_tbl1"
    "public.testpub_rf_tbl3" WHERE ((e > 1000) AND (e < 2000))
ALTER PUBLICATION testpub5 SET TABLE testpub_rf_tbl3 WHERE (e > 300 AND e < 500);`,
			},
			{
				Statement: `\dRp+ testpub5
                                    Publication testpub5
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | f       | f       | f         | f
Tables:
    "public.testpub_rf_tbl3" WHERE ((e > 300) AND (e < 500))
\d testpub_rf_tbl3
          Table "public.testpub_rf_tbl3"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 e      | integer |           |          | 
Publications:
    "testpub5" WHERE ((e > 300) AND (e < 500))
SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_rf_yes FOR TABLE testpub_rf_tbl1 WHERE (a > 1) WITH (publish = 'insert');`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_rf_no FOR TABLE testpub_rf_tbl1;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `\d testpub_rf_tbl1
          Table "public.testpub_rf_tbl1"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 a      | integer |           |          | 
 b      | text    |           |          | 
Publications:
    "testpub_rf_no"
    "testpub_rf_yes" WHERE (a > 1)
DROP PUBLICATION testpub_rf_yes, testpub_rf_no;`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_syntax1 FOR TABLE testpub_rf_tbl1, ONLY testpub_rf_tbl3 WHERE (e < 999) WITH (publish = 'insert');`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `\dRp+ testpub_syntax1
                                Publication testpub_syntax1
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | f       | f       | f         | f
Tables:
    "public.testpub_rf_tbl1"
    "public.testpub_rf_tbl3" WHERE (e < 999)
DROP PUBLICATION testpub_syntax1;`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_syntax2 FOR TABLE testpub_rf_tbl1, testpub_rf_schema1.testpub_rf_tbl5 WHERE (h < 999) WITH (publish = 'insert');`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `\dRp+ testpub_syntax2
                                Publication testpub_syntax2
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | f       | f       | f         | f
Tables:
    "public.testpub_rf_tbl1"
    "testpub_rf_schema1.testpub_rf_tbl5" WHERE (h < 999)
DROP PUBLICATION testpub_syntax2;`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_syntax3 FOR TABLES IN SCHEMA testpub_rf_schema1 WHERE (a = 123);`,
				ErrorString: `syntax error at or near "WHERE"`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_syntax3 FOR TABLES IN SCHEMA testpub_rf_schema1, testpub_rf_schema1 WHERE (a = 123);`,
				ErrorString: `WHERE clause not allowed for schema`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_dups FOR TABLE testpub_rf_tbl1 WHERE (a = 1), testpub_rf_tbl1 WITH (publish = 'insert');`,
				ErrorString: `conflicting or redundant WHERE clauses for table "testpub_rf_tbl1"`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_dups FOR TABLE testpub_rf_tbl1, testpub_rf_tbl1 WHERE (a = 2) WITH (publish = 'insert');`,
				ErrorString: `conflicting or redundant WHERE clauses for table "testpub_rf_tbl1"`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub5 SET TABLE testpub_rf_tbl3 WHERE (1234);`,
				ErrorString: `argument of PUBLICATION WHERE must be type boolean, not type integer`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub5 SET TABLE testpub_rf_tbl3 WHERE (e < AVG(e));`,
				ErrorString: `aggregate functions are not allowed in WHERE`,
			},
			{
				Statement: `CREATE FUNCTION testpub_rf_func1(integer, integer) RETURNS boolean AS $$ SELECT hashint4($1) > $2 $$ LANGUAGE SQL;`,
			},
			{
				Statement: `CREATE OPERATOR =#> (PROCEDURE = testpub_rf_func1, LEFTARG = integer, RIGHTARG = integer);`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub6 FOR TABLE testpub_rf_tbl3 WHERE (e =#> 27);`,
				ErrorString: `invalid publication WHERE expression`,
			},
			{
				Statement: `CREATE FUNCTION testpub_rf_func2() RETURNS integer AS $$ BEGIN RETURN 123; END; $$ LANGUAGE plpgsql;`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub5 ADD TABLE testpub_rf_tbl1 WHERE (a >= testpub_rf_func2());`,
				ErrorString: `invalid publication WHERE expression`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub5 ADD TABLE testpub_rf_tbl1 WHERE (a < random());`,
				ErrorString: `invalid publication WHERE expression`,
			},
			{
				Statement: `CREATE COLLATION user_collation FROM "C";`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub5 ADD TABLE testpub_rf_tbl1 WHERE (b < '2' COLLATE user_collation);`,
				ErrorString: `invalid publication WHERE expression`,
			},
			{
				Statement: `ALTER PUBLICATION testpub5 SET TABLE testpub_rf_tbl1 WHERE (NULLIF(1,2) = a);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub5 SET TABLE testpub_rf_tbl1 WHERE (a IS NULL);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub5 SET TABLE testpub_rf_tbl1 WHERE ((a > 5) IS FALSE);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub5 SET TABLE testpub_rf_tbl1 WHERE (a IS DISTINCT FROM 5);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub5 SET TABLE testpub_rf_tbl1 WHERE ((a, a + 1) < (2, 3));`,
			},
			{
				Statement: `ALTER PUBLICATION testpub5 SET TABLE testpub_rf_tbl1 WHERE (b::varchar < '2');`,
			},
			{
				Statement: `ALTER PUBLICATION testpub5 SET TABLE testpub_rf_tbl4 WHERE (length(g) < 6);`,
			},
			{
				Statement: `CREATE TYPE rf_bug_status AS ENUM ('new', 'open', 'closed');`,
			},
			{
				Statement: `CREATE TABLE rf_bug (id serial, description text, status rf_bug_status);`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub6 FOR TABLE rf_bug WHERE (status = 'open') WITH (publish = 'insert');`,
				ErrorString: `invalid publication WHERE expression`,
			},
			{
				Statement: `DROP TABLE rf_bug;`,
			},
			{
				Statement: `DROP TYPE rf_bug_status;`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub6 FOR TABLE testpub_rf_tbl1 WHERE (a IN (SELECT generate_series(1,5)));`,
				ErrorString: `invalid publication WHERE expression`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub6 FOR TABLE testpub_rf_tbl1 WHERE ('(0,1)'::tid = ctid);`,
				ErrorString: `invalid publication WHERE expression`,
			},
			{
				Statement: `ALTER PUBLICATION testpub5 SET TABLE testpub_rf_tbl5 WHERE (a IS DOCUMENT);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub5 SET TABLE testpub_rf_tbl5 WHERE (xmlexists('//foo[text() = ''bar'']' PASSING BY VALUE a));`,
			},
			{
				Statement: `ALTER PUBLICATION testpub5 SET TABLE testpub_rf_tbl1 WHERE (NULLIF(1, 2) = a);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub5 SET TABLE testpub_rf_tbl1 WHERE (CASE a WHEN 5 THEN true ELSE false END);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub5 SET TABLE testpub_rf_tbl1 WHERE (COALESCE(b, 'foo') = 'foo');`,
			},
			{
				Statement: `ALTER PUBLICATION testpub5 SET TABLE testpub_rf_tbl1 WHERE (GREATEST(a, 10) > 10);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub5 SET TABLE testpub_rf_tbl1 WHERE (a IN (2, 4, 6));`,
			},
			{
				Statement: `ALTER PUBLICATION testpub5 SET TABLE testpub_rf_tbl1 WHERE (ARRAY[a] <@ ARRAY[2, 4, 6]);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub5 SET TABLE testpub_rf_tbl1 WHERE (ROW(a, 2) IS NULL);`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub5 DROP TABLE testpub_rf_tbl1 WHERE (e < 27);`,
				ErrorString: `cannot use a WHERE clause when removing a table from a publication`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub6 FOR TABLES IN SCHEMA testpub_rf_schema2;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLES IN SCHEMA testpub_rf_schema2, TABLE testpub_rf_schema2.testpub_rf_tbl6 WHERE (i < 99);`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `\dRp+ testpub6
                                    Publication testpub6
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables:
    "testpub_rf_schema2.testpub_rf_tbl6" WHERE (i < 99)
Tables from schemas:
    "testpub_rf_schema2"
DROP TABLE testpub_rf_tbl1;`,
			},
			{
				Statement: `DROP TABLE testpub_rf_tbl2;`,
			},
			{
				Statement: `DROP TABLE testpub_rf_tbl3;`,
			},
			{
				Statement: `DROP TABLE testpub_rf_tbl4;`,
			},
			{
				Statement: `DROP TABLE testpub_rf_tbl5;`,
			},
			{
				Statement: `DROP TABLE testpub_rf_schema1.testpub_rf_tbl5;`,
			},
			{
				Statement: `DROP TABLE testpub_rf_schema2.testpub_rf_tbl6;`,
			},
			{
				Statement: `DROP SCHEMA testpub_rf_schema1;`,
			},
			{
				Statement: `DROP SCHEMA testpub_rf_schema2;`,
			},
			{
				Statement: `DROP PUBLICATION testpub5;`,
			},
			{
				Statement: `DROP PUBLICATION testpub6;`,
			},
			{
				Statement: `DROP OPERATOR =#>(integer, integer);`,
			},
			{
				Statement: `DROP FUNCTION testpub_rf_func1(integer, integer);`,
			},
			{
				Statement: `DROP FUNCTION testpub_rf_func2();`,
			},
			{
				Statement: `DROP COLLATION user_collation;`,
			},
			{
				Statement: `CREATE TABLE rf_tbl_abcd_nopk(a int, b int, c int, d int);`,
			},
			{
				Statement: `CREATE TABLE rf_tbl_abcd_pk(a int, b int, c int, d int, PRIMARY KEY(a,b));`,
			},
			{
				Statement: `CREATE TABLE rf_tbl_abcd_part_pk (a int PRIMARY KEY, b int) PARTITION by RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE rf_tbl_abcd_part_pk_1 (b int, a int PRIMARY KEY);`,
			},
			{
				Statement: `ALTER TABLE rf_tbl_abcd_part_pk ATTACH PARTITION rf_tbl_abcd_part_pk_1 FOR VALUES FROM (1) TO (10);`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub6 FOR TABLE rf_tbl_abcd_pk WHERE (a > 99);`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `UPDATE rf_tbl_abcd_pk SET a = 1;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_pk WHERE (b > 99);`,
			},
			{
				Statement: `UPDATE rf_tbl_abcd_pk SET a = 1;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_pk WHERE (c > 99);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_pk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_pk"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_pk WHERE (d > 99);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_pk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_pk"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_nopk WHERE (a > 99);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_nopk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_nopk"`,
			},
			{
				Statement: `ALTER TABLE rf_tbl_abcd_pk REPLICA IDENTITY FULL;`,
			},
			{
				Statement: `ALTER TABLE rf_tbl_abcd_nopk REPLICA IDENTITY FULL;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_pk WHERE (c > 99);`,
			},
			{
				Statement: `UPDATE rf_tbl_abcd_pk SET a = 1;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_nopk WHERE (a > 99);`,
			},
			{
				Statement: `UPDATE rf_tbl_abcd_nopk SET a = 1;`,
			},
			{
				Statement: `ALTER TABLE rf_tbl_abcd_pk REPLICA IDENTITY NOTHING;`,
			},
			{
				Statement: `ALTER TABLE rf_tbl_abcd_nopk REPLICA IDENTITY NOTHING;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_pk WHERE (a > 99);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_pk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_pk"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_pk WHERE (c > 99);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_pk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_pk"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_nopk WHERE (a > 99);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_nopk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_nopk"`,
			},
			{
				Statement: `ALTER TABLE rf_tbl_abcd_pk ALTER COLUMN c SET NOT NULL;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX idx_abcd_pk_c ON rf_tbl_abcd_pk(c);`,
			},
			{
				Statement: `ALTER TABLE rf_tbl_abcd_pk REPLICA IDENTITY USING INDEX idx_abcd_pk_c;`,
			},
			{
				Statement: `ALTER TABLE rf_tbl_abcd_nopk ALTER COLUMN c SET NOT NULL;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX idx_abcd_nopk_c ON rf_tbl_abcd_nopk(c);`,
			},
			{
				Statement: `ALTER TABLE rf_tbl_abcd_nopk REPLICA IDENTITY USING INDEX idx_abcd_nopk_c;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_pk WHERE (a > 99);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_pk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_pk"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_pk WHERE (c > 99);`,
			},
			{
				Statement: `UPDATE rf_tbl_abcd_pk SET a = 1;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_nopk WHERE (a > 99);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_nopk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_nopk"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_nopk WHERE (c > 99);`,
			},
			{
				Statement: `UPDATE rf_tbl_abcd_nopk SET a = 1;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET (PUBLISH_VIA_PARTITION_ROOT=0);`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_part_pk WHERE (a > 99);`,
				ErrorString: `cannot use publication WHERE clause for relation "rf_tbl_abcd_part_pk"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_part_pk_1 WHERE (a > 99);`,
			},
			{
				Statement: `UPDATE rf_tbl_abcd_part_pk SET a = 1;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET (PUBLISH_VIA_PARTITION_ROOT=1);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_part_pk WHERE (a > 99);`,
			},
			{
				Statement: `UPDATE rf_tbl_abcd_part_pk SET a = 1;`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub6 SET (PUBLISH_VIA_PARTITION_ROOT=0);`,
				ErrorString: `cannot set parameter "publish_via_partition_root" to false for publication "testpub6"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_part_pk;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET (PUBLISH_VIA_PARTITION_ROOT=0);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_part_pk_1 WHERE (b > 99);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET (PUBLISH_VIA_PARTITION_ROOT=0);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_part_pk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_part_pk_1"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET (PUBLISH_VIA_PARTITION_ROOT=1);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_part_pk WHERE (b > 99);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_part_pk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_part_pk_1"`,
			},
			{
				Statement: `DROP PUBLICATION testpub6;`,
			},
			{
				Statement: `DROP TABLE rf_tbl_abcd_pk;`,
			},
			{
				Statement: `DROP TABLE rf_tbl_abcd_nopk;`,
			},
			{
				Statement: `DROP TABLE rf_tbl_abcd_part_pk;`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_dups FOR TABLE testpub_tbl1 (a), testpub_tbl1 WITH (publish = 'insert');`,
				ErrorString: `conflicting or redundant column lists for table "testpub_tbl1"`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_dups FOR TABLE testpub_tbl1, testpub_tbl1 (a) WITH (publish = 'insert');`,
				ErrorString: `conflicting or redundant column lists for table "testpub_tbl1"`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_fortable FOR TABLE testpub_tbl1;`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_fortable_insert WITH (publish = 'insert');`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `CREATE TABLE testpub_tbl5 (a int PRIMARY KEY, b text, c text,
	d int generated always as (a + length(b)) stored);`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub_fortable ADD TABLE testpub_tbl5 (a, x);`,
				ErrorString: `column "x" of relation "testpub_tbl5" does not exist`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_fortable ADD TABLE testpub_tbl5 (b, c);`,
			},
			{
				Statement:   `UPDATE testpub_tbl5 SET a = 1;`,
				ErrorString: `cannot update table "testpub_tbl5"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_fortable DROP TABLE testpub_tbl5;`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub_fortable ADD TABLE testpub_tbl5 (a, d);`,
				ErrorString: `cannot use generated column "d" in publication column list`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub_fortable ADD TABLE testpub_tbl5 (a, ctid);`,
				ErrorString: `cannot use system column "ctid" in publication column list`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_fortable ADD TABLE testpub_tbl5 (a, c);`,
			},
			{
				Statement:   `ALTER TABLE testpub_tbl5 DROP COLUMN c;		-- no dice`,
				ErrorString: `cannot drop column c of table testpub_tbl5 because other objects depend on it`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_fortable_insert ADD TABLE testpub_tbl5 (b, c);`,
			},
			{
				Statement: `/* not all replica identities are good enough */
CREATE UNIQUE INDEX testpub_tbl5_b_key ON testpub_tbl5 (b, c);`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl5 ALTER b SET NOT NULL, ALTER c SET NOT NULL;`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl5 REPLICA IDENTITY USING INDEX testpub_tbl5_b_key;`,
			},
			{
				Statement:   `UPDATE testpub_tbl5 SET a = 1;`,
				ErrorString: `cannot update table "testpub_tbl5"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_fortable DROP TABLE testpub_tbl5;`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl5 REPLICA IDENTITY USING INDEX testpub_tbl5_b_key;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_fortable ADD TABLE testpub_tbl5 (a, c);`,
			},
			{
				Statement:   `UPDATE testpub_tbl5 SET a = 1;`,
				ErrorString: `cannot update table "testpub_tbl5"`,
			},
			{
				Statement: `/* But if upd/del are not published, it works OK */
SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_table_ins WITH (publish = 'insert, truncate');`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_table_ins ADD TABLE testpub_tbl5 (a);		-- ok`,
			},
			{
				Statement: `\dRp+ testpub_table_ins
                               Publication testpub_table_ins
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | f       | f       | t         | f
Tables:
    "public.testpub_tbl5" (a)
CREATE TABLE testpub_tbl6 (a int, b text, c text);`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl6 REPLICA IDENTITY FULL;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_fortable ADD TABLE testpub_tbl6 (a, b, c);`,
			},
			{
				Statement:   `UPDATE testpub_tbl6 SET a = 1;`,
				ErrorString: `cannot update table "testpub_tbl6"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_fortable DROP TABLE testpub_tbl6;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_fortable ADD TABLE testpub_tbl6; -- ok`,
			},
			{
				Statement: `UPDATE testpub_tbl6 SET a = 1;`,
			},
			{
				Statement: `CREATE TABLE testpub_tbl7 (a int primary key, b text, c text);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_fortable ADD TABLE testpub_tbl7 (a, b);`,
			},
			{
				Statement: `\d+ testpub_tbl7
                                Table "public.testpub_tbl7"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 a      | integer |           | not null |         | plain    |              | 
 b      | text    |           |          |         | extended |              | 
 c      | text    |           |          |         | extended |              | 
Indexes:
    "testpub_tbl7_pkey" PRIMARY KEY, btree (a)
Publications:
    "testpub_fortable" (a, b)
ALTER PUBLICATION testpub_fortable SET TABLE testpub_tbl7 (a, b);`,
			},
			{
				Statement: `\d+ testpub_tbl7
                                Table "public.testpub_tbl7"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 a      | integer |           | not null |         | plain    |              | 
 b      | text    |           |          |         | extended |              | 
 c      | text    |           |          |         | extended |              | 
Indexes:
    "testpub_tbl7_pkey" PRIMARY KEY, btree (a)
Publications:
    "testpub_fortable" (a, b)
ALTER PUBLICATION testpub_fortable SET TABLE testpub_tbl7 (a, c);`,
			},
			{
				Statement: `\d+ testpub_tbl7
                                Table "public.testpub_tbl7"
 Column |  Type   | Collation | Nullable | Default | Storage  | Stats target | Description 
--------+---------+-----------+----------+---------+----------+--------------+-------------
 a      | integer |           | not null |         | plain    |              | 
 b      | text    |           |          |         | extended |              | 
 c      | text    |           |          |         | extended |              | 
Indexes:
    "testpub_tbl7_pkey" PRIMARY KEY, btree (a)
Publications:
    "testpub_fortable" (a, c)
CREATE TABLE testpub_tbl8 (a int, b text, c text) PARTITION BY HASH (a);`,
			},
			{
				Statement: `CREATE TABLE testpub_tbl8_0 PARTITION OF testpub_tbl8 FOR VALUES WITH (modulus 2, remainder 0);`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl8_0 ADD PRIMARY KEY (a);`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl8_0 REPLICA IDENTITY USING INDEX testpub_tbl8_0_pkey;`,
			},
			{
				Statement: `CREATE TABLE testpub_tbl8_1 PARTITION OF testpub_tbl8 FOR VALUES WITH (modulus 2, remainder 1);`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl8_1 ADD PRIMARY KEY (b);`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl8_1 REPLICA IDENTITY USING INDEX testpub_tbl8_1_pkey;`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_col_list FOR TABLE testpub_tbl8 (a, b) WITH (publish_via_partition_root = 'true');`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_col_list DROP TABLE testpub_tbl8;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_col_list ADD TABLE testpub_tbl8 (a, b);`,
			},
			{
				Statement: `UPDATE testpub_tbl8 SET a = 1;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_col_list DROP TABLE testpub_tbl8;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_col_list ADD TABLE testpub_tbl8 (a, c);`,
			},
			{
				Statement:   `UPDATE testpub_tbl8 SET a = 1;`,
				ErrorString: `cannot update table "testpub_tbl8_1"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_col_list DROP TABLE testpub_tbl8;`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl8_1 REPLICA IDENTITY FULL;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_col_list ADD TABLE testpub_tbl8 (a, c);`,
			},
			{
				Statement:   `UPDATE testpub_tbl8 SET a = 1;`,
				ErrorString: `cannot update table "testpub_tbl8_1"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_col_list DROP TABLE testpub_tbl8;`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl8_1 REPLICA IDENTITY USING INDEX testpub_tbl8_1_pkey;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_col_list ADD TABLE testpub_tbl8 (a, b);`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl8_1 REPLICA IDENTITY FULL;`,
			},
			{
				Statement:   `UPDATE testpub_tbl8 SET a = 1;`,
				ErrorString: `cannot update table "testpub_tbl8_1"`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl8_1 DROP CONSTRAINT testpub_tbl8_1_pkey;`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl8_1 ADD PRIMARY KEY (c);`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl8_1 REPLICA IDENTITY USING INDEX testpub_tbl8_1_pkey;`,
			},
			{
				Statement:   `UPDATE testpub_tbl8 SET a = 1;`,
				ErrorString: `cannot update table "testpub_tbl8_1"`,
			},
			{
				Statement: `DROP TABLE testpub_tbl8;`,
			},
			{
				Statement: `CREATE TABLE testpub_tbl8 (a int, b text, c text) PARTITION BY HASH (a);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_col_list ADD TABLE testpub_tbl8 (a, b);`,
			},
			{
				Statement: `CREATE TABLE testpub_tbl8_0 (a int, b text, c text);`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl8_0 ADD PRIMARY KEY (a);`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl8_0 REPLICA IDENTITY USING INDEX testpub_tbl8_0_pkey;`,
			},
			{
				Statement: `CREATE TABLE testpub_tbl8_1 (a int, b text, c text);`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl8_1 ADD PRIMARY KEY (c);`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl8_1 REPLICA IDENTITY USING INDEX testpub_tbl8_1_pkey;`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl8 ATTACH PARTITION testpub_tbl8_0 FOR VALUES WITH (modulus 2, remainder 0);`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl8 ATTACH PARTITION testpub_tbl8_1 FOR VALUES WITH (modulus 2, remainder 1);`,
			},
			{
				Statement:   `UPDATE testpub_tbl8 SET a = 1;`,
				ErrorString: `cannot update table "testpub_tbl8_1"`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl8_0 REPLICA IDENTITY FULL;`,
			},
			{
				Statement:   `UPDATE testpub_tbl8 SET a = 1;`,
				ErrorString: `cannot update table "testpub_tbl8_0"`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_tbl9 FOR TABLES IN SCHEMA public, TABLE public.testpub_tbl7(a);`,
				ErrorString: `cannot use column list for relation "public.testpub_tbl7" in publication "testpub_tbl9"`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_tbl9 FOR TABLES IN SCHEMA public;`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub_tbl9 ADD TABLE public.testpub_tbl7(a);`,
				ErrorString: `cannot use column list for relation "public.testpub_tbl7" in publication "testpub_tbl9"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_tbl9 SET TABLE public.testpub_tbl7(a);`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub_tbl9 ADD TABLES IN SCHEMA public;`,
				ErrorString: `cannot add schema to publication "testpub_tbl9"`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub_tbl9 SET TABLES IN SCHEMA public, TABLE public.testpub_tbl7(a);`,
				ErrorString: `cannot use column list for relation "public.testpub_tbl7" in publication "testpub_tbl9"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_tbl9 DROP TABLE public.testpub_tbl7;`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub_tbl9 ADD TABLES IN SCHEMA public, TABLE public.testpub_tbl7(a);`,
				ErrorString: `cannot use column list for relation "public.testpub_tbl7" in publication "testpub_tbl9"`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `DROP TABLE testpub_tbl5, testpub_tbl6, testpub_tbl7, testpub_tbl8, testpub_tbl8_1;`,
			},
			{
				Statement: `DROP PUBLICATION testpub_table_ins, testpub_fortable, testpub_fortable_insert, testpub_col_list, testpub_tbl9;`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_both_filters;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `CREATE TABLE testpub_tbl_both_filters (a int, b int, c int, PRIMARY KEY (a,c));`,
			},
			{
				Statement: `ALTER TABLE testpub_tbl_both_filters REPLICA IDENTITY USING INDEX testpub_tbl_both_filters_pkey;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_both_filters ADD TABLE testpub_tbl_both_filters (a,c) WHERE (c != 1);`,
			},
			{
				Statement: `\dRp+ testpub_both_filters
                              Publication testpub_both_filters
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables:
    "public.testpub_tbl_both_filters" (a, c) WHERE (c <> 1)
\d+ testpub_tbl_both_filters
                         Table "public.testpub_tbl_both_filters"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 a      | integer |           | not null |         | plain   |              | 
 b      | integer |           |          |         | plain   |              | 
 c      | integer |           | not null |         | plain   |              | 
Indexes:
    "testpub_tbl_both_filters_pkey" PRIMARY KEY, btree (a, c) REPLICA IDENTITY
Publications:
    "testpub_both_filters" (a, c) WHERE (c <> 1)
DROP TABLE testpub_tbl_both_filters;`,
			},
			{
				Statement: `DROP PUBLICATION testpub_both_filters;`,
			},
			{
				Statement: `CREATE TABLE rf_tbl_abcd_nopk(a int, b int, c int, d int);`,
			},
			{
				Statement: `CREATE TABLE rf_tbl_abcd_pk(a int, b int, c int, d int, PRIMARY KEY(a,b));`,
			},
			{
				Statement: `CREATE TABLE rf_tbl_abcd_part_pk (a int PRIMARY KEY, b int) PARTITION by RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE rf_tbl_abcd_part_pk_1 (b int, a int PRIMARY KEY);`,
			},
			{
				Statement: `ALTER TABLE rf_tbl_abcd_part_pk ATTACH PARTITION rf_tbl_abcd_part_pk_1 FOR VALUES FROM (1) TO (10);`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub6 FOR TABLE rf_tbl_abcd_pk (a, b);`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `UPDATE rf_tbl_abcd_pk SET a = 1;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_pk (a, b, c);`,
			},
			{
				Statement: `UPDATE rf_tbl_abcd_pk SET a = 1;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_pk (a);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_pk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_pk"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_pk (b);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_pk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_pk"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_nopk (a);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_nopk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_nopk" because it does not have a replica identity and publishes updates`,
			},
			{
				Statement: `ALTER TABLE rf_tbl_abcd_pk REPLICA IDENTITY FULL;`,
			},
			{
				Statement: `ALTER TABLE rf_tbl_abcd_nopk REPLICA IDENTITY FULL;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_pk (c);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_pk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_pk"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_nopk (a, b, c, d);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_nopk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_nopk"`,
			},
			{
				Statement: `ALTER TABLE rf_tbl_abcd_pk REPLICA IDENTITY NOTHING;`,
			},
			{
				Statement: `ALTER TABLE rf_tbl_abcd_nopk REPLICA IDENTITY NOTHING;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_pk (a);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_pk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_pk" because it does not have a replica identity and publishes updates`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_pk (a, b, c, d);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_pk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_pk" because it does not have a replica identity and publishes updates`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_nopk (d);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_nopk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_nopk" because it does not have a replica identity and publishes updates`,
			},
			{
				Statement: `ALTER TABLE rf_tbl_abcd_pk ALTER COLUMN c SET NOT NULL;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX idx_abcd_pk_c ON rf_tbl_abcd_pk(c);`,
			},
			{
				Statement: `ALTER TABLE rf_tbl_abcd_pk REPLICA IDENTITY USING INDEX idx_abcd_pk_c;`,
			},
			{
				Statement: `ALTER TABLE rf_tbl_abcd_nopk ALTER COLUMN c SET NOT NULL;`,
			},
			{
				Statement: `CREATE UNIQUE INDEX idx_abcd_nopk_c ON rf_tbl_abcd_nopk(c);`,
			},
			{
				Statement: `ALTER TABLE rf_tbl_abcd_nopk REPLICA IDENTITY USING INDEX idx_abcd_nopk_c;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_pk (a);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_pk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_pk"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_pk (c);`,
			},
			{
				Statement: `UPDATE rf_tbl_abcd_pk SET a = 1;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_nopk (a);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_nopk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_nopk"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_nopk (c);`,
			},
			{
				Statement: `UPDATE rf_tbl_abcd_nopk SET a = 1;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET (PUBLISH_VIA_PARTITION_ROOT=0);`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_part_pk (a);`,
				ErrorString: `cannot use column list for relation "public.rf_tbl_abcd_part_pk" in publication "testpub6"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_part_pk_1 (a);`,
			},
			{
				Statement: `UPDATE rf_tbl_abcd_part_pk SET a = 1;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET (PUBLISH_VIA_PARTITION_ROOT=1);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_part_pk (a);`,
			},
			{
				Statement: `UPDATE rf_tbl_abcd_part_pk SET a = 1;`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub6 SET (PUBLISH_VIA_PARTITION_ROOT=0);`,
				ErrorString: `cannot set parameter "publish_via_partition_root" to false for publication "testpub6"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_part_pk;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET (PUBLISH_VIA_PARTITION_ROOT=0);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_part_pk_1 (b);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET (PUBLISH_VIA_PARTITION_ROOT=0);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_part_pk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_part_pk_1"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET (PUBLISH_VIA_PARTITION_ROOT=1);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub6 SET TABLE rf_tbl_abcd_part_pk (b);`,
			},
			{
				Statement:   `UPDATE rf_tbl_abcd_part_pk SET a = 1;`,
				ErrorString: `cannot update table "rf_tbl_abcd_part_pk_1"`,
			},
			{
				Statement: `DROP PUBLICATION testpub6;`,
			},
			{
				Statement: `DROP TABLE rf_tbl_abcd_pk;`,
			},
			{
				Statement: `DROP TABLE rf_tbl_abcd_nopk;`,
			},
			{
				Statement: `DROP TABLE rf_tbl_abcd_part_pk;`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE TABLE testpub_tbl4(a int);`,
			},
			{
				Statement: `INSERT INTO testpub_tbl4 values(1);`,
			},
			{
				Statement: `UPDATE testpub_tbl4 set a = 2;`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_foralltables FOR ALL TABLES;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement:   `UPDATE testpub_tbl4 set a = 3;`,
				ErrorString: `cannot update table "testpub_tbl4" because it does not have a replica identity and publishes updates`,
			},
			{
				Statement: `DROP PUBLICATION testpub_foralltables;`,
			},
			{
				Statement: `UPDATE testpub_tbl4 set a = 3;`,
			},
			{
				Statement: `DROP TABLE testpub_tbl4;`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_fortbl FOR TABLE testpub_view;`,
				ErrorString: `cannot add relation "testpub_view" to publication`,
			},
			{
				Statement: `CREATE TEMPORARY TABLE testpub_temptbl(a int);`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_fortemptbl FOR TABLE testpub_temptbl;`,
				ErrorString: `cannot add relation "testpub_temptbl" to publication`,
			},
			{
				Statement: `DROP TABLE testpub_temptbl;`,
			},
			{
				Statement: `CREATE UNLOGGED TABLE testpub_unloggedtbl(a int);`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_forunloggedtbl FOR TABLE testpub_unloggedtbl;`,
				ErrorString: `cannot add relation "testpub_unloggedtbl" to publication`,
			},
			{
				Statement: `DROP TABLE testpub_unloggedtbl;`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_forsystemtbl FOR TABLE pg_publication;`,
				ErrorString: `cannot add relation "pg_publication" to publication`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_fortbl FOR TABLE testpub_tbl1, pub_test.testpub_nopk;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub_fortbl ADD TABLE testpub_tbl1;`,
				ErrorString: `relation "testpub_tbl1" is already member of publication "testpub_fortbl"`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_fortbl FOR TABLE testpub_tbl1;`,
				ErrorString: `publication "testpub_fortbl" already exists`,
			},
			{
				Statement: `\dRp+ testpub_fortbl
                                 Publication testpub_fortbl
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables:
    "pub_test.testpub_nopk"
    "public.testpub_tbl1"
ALTER PUBLICATION testpub_default ADD TABLE testpub_view;`,
				ErrorString: `cannot add relation "testpub_view" to publication`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_default ADD TABLE testpub_tbl1;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_default SET TABLE testpub_tbl1;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_default ADD TABLE pub_test.testpub_nopk;`,
			},
			{
				Statement: `ALTER PUBLICATION testpib_ins_trunct ADD TABLE pub_test.testpub_nopk, testpub_tbl1;`,
			},
			{
				Statement: `\d+ pub_test.testpub_nopk
                              Table "pub_test.testpub_nopk"
 Column |  Type   | Collation | Nullable | Default | Storage | Stats target | Description 
--------+---------+-----------+----------+---------+---------+--------------+-------------
 foo    | integer |           |          |         | plain   |              | 
 bar    | integer |           |          |         | plain   |              | 
Publications:
    "testpib_ins_trunct"
    "testpub_default"
    "testpub_fortbl"
\d+ testpub_tbl1
                                                Table "public.testpub_tbl1"
 Column |  Type   | Collation | Nullable |                 Default                  | Storage  | Stats target | Description 
--------+---------+-----------+----------+------------------------------------------+----------+--------------+-------------
 id     | integer |           | not null | nextval('testpub_tbl1_id_seq'::regclass) | plain    |              | 
 data   | text    |           |          |                                          | extended |              | 
Indexes:
    "testpub_tbl1_pkey" PRIMARY KEY, btree (id)
Publications:
    "testpib_ins_trunct"
    "testpub_default"
    "testpub_fortbl"
\dRp+ testpub_default
                                Publication testpub_default
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | f         | f
Tables:
    "pub_test.testpub_nopk"
    "public.testpub_tbl1"
ALTER PUBLICATION testpub_default DROP TABLE testpub_tbl1, pub_test.testpub_nopk;`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub_default DROP TABLE pub_test.testpub_nopk;`,
				ErrorString: `relation "testpub_nopk" is not part of the publication`,
			},
			{
				Statement: `\d+ testpub_tbl1
                                                Table "public.testpub_tbl1"
 Column |  Type   | Collation | Nullable |                 Default                  | Storage  | Stats target | Description 
--------+---------+-----------+----------+------------------------------------------+----------+--------------+-------------
 id     | integer |           | not null | nextval('testpub_tbl1_id_seq'::regclass) | plain    |              | 
 data   | text    |           |          |                                          | extended |              | 
Indexes:
    "testpub_tbl1_pkey" PRIMARY KEY, btree (id)
Publications:
    "testpib_ins_trunct"
    "testpub_fortbl"
CREATE TABLE pub_test.testpub_addpk (id int not null, data int);`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_default ADD TABLE pub_test.testpub_addpk;`,
			},
			{
				Statement: `INSERT INTO pub_test.testpub_addpk VALUES(1, 11);`,
			},
			{
				Statement: `CREATE UNIQUE INDEX testpub_addpk_id_idx ON pub_test.testpub_addpk(id);`,
			},
			{
				Statement:   `UPDATE pub_test.testpub_addpk SET id = 2;`,
				ErrorString: `cannot update table "testpub_addpk" because it does not have a replica identity and publishes updates`,
			},
			{
				Statement: `ALTER TABLE pub_test.testpub_addpk ADD PRIMARY KEY USING INDEX testpub_addpk_id_idx;`,
			},
			{
				Statement: `UPDATE pub_test.testpub_addpk SET id = 2;`,
			},
			{
				Statement: `DROP TABLE pub_test.testpub_addpk;`,
			},
			{
				Statement: `SET ROLE regress_publication_user2;`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub2;  -- fail`,
				ErrorString: `permission denied for database regression`,
			},
			{
				Statement: `SET ROLE regress_publication_user;`,
			},
			{
				Statement: `GRANT CREATE ON DATABASE regression TO regress_publication_user2;`,
			},
			{
				Statement: `SET ROLE regress_publication_user2;`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub2;  -- ok`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub3 FOR TABLES IN SCHEMA pub_test;  -- fail`,
				ErrorString: `must be superuser to create FOR TABLES IN SCHEMA publication`,
			},
			{
				Statement: `CREATE PUBLICATION testpub3;  -- ok`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub2 ADD TABLE testpub_tbl1;  -- fail`,
				ErrorString: `must be owner of table testpub_tbl1`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub3 ADD TABLES IN SCHEMA pub_test;  -- fail`,
				ErrorString: `must be superuser to add or set schemas`,
			},
			{
				Statement: `SET ROLE regress_publication_user;`,
			},
			{
				Statement: `GRANT regress_publication_user TO regress_publication_user2;`,
			},
			{
				Statement: `SET ROLE regress_publication_user2;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub2 ADD TABLE testpub_tbl1;  -- ok`,
			},
			{
				Statement: `DROP PUBLICATION testpub2;`,
			},
			{
				Statement: `DROP PUBLICATION testpub3;`,
			},
			{
				Statement: `SET ROLE regress_publication_user;`,
			},
			{
				Statement: `CREATE ROLE regress_publication_user3;`,
			},
			{
				Statement: `GRANT regress_publication_user2 TO regress_publication_user3;`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub4 FOR TABLES IN SCHEMA pub_test;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub4 OWNER TO regress_publication_user3;`,
			},
			{
				Statement: `SET ROLE regress_publication_user3;`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub4 owner to regress_publication_user2; -- fail`,
				ErrorString: `permission denied to change owner of publication "testpub4"`,
			},
			{
				Statement: `ALTER PUBLICATION testpub4 owner to regress_publication_user; -- ok`,
			},
			{
				Statement: `SET ROLE regress_publication_user;`,
			},
			{
				Statement: `DROP PUBLICATION testpub4;`,
			},
			{
				Statement: `DROP ROLE regress_publication_user3;`,
			},
			{
				Statement: `REVOKE CREATE ON DATABASE regression FROM regress_publication_user2;`,
			},
			{
				Statement: `DROP TABLE testpub_parted;`,
			},
			{
				Statement: `DROP TABLE testpub_tbl1;`,
			},
			{
				Statement: `\dRp+ testpub_default
                                Publication testpub_default
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | f         | f
(1 row)
SET ROLE regress_publication_user_dummy;`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub_default RENAME TO testpub_dummy;`,
				ErrorString: `must be owner of publication testpub_default`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_default RENAME TO testpub_foo;`,
			},
			{
				Statement: `\dRp testpub_foo
                                           List of publications
    Name     |          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
-------------+--------------------------+------------+---------+---------+---------+-----------+----------
 testpub_foo | regress_publication_user | f          | t       | t       | t       | f         | f
(1 row)
ALTER PUBLICATION testpub_foo RENAME TO testpub_default;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub_default OWNER TO regress_publication_user2;`,
			},
			{
				Statement: `\dRp testpub_default
                                             List of publications
      Name       |           Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
-----------------+---------------------------+------------+---------+---------+---------+-----------+----------
 testpub_default | regress_publication_user2 | f          | t       | t       | t       | f         | f
(1 row)
CREATE SCHEMA pub_test1;`,
			},
			{
				Statement: `CREATE SCHEMA pub_test2;`,
			},
			{
				Statement: `CREATE SCHEMA pub_test3;`,
			},
			{
				Statement: `CREATE SCHEMA "CURRENT_SCHEMA";`,
			},
			{
				Statement: `CREATE TABLE pub_test1.tbl (id int, data text);`,
			},
			{
				Statement: `CREATE TABLE pub_test1.tbl1 (id serial primary key, data text);`,
			},
			{
				Statement: `CREATE TABLE pub_test2.tbl1 (id serial primary key, data text);`,
			},
			{
				Statement: `CREATE TABLE "CURRENT_SCHEMA"."CURRENT_SCHEMA"(id int);`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub1_forschema FOR TABLES IN SCHEMA pub_test1;`,
			},
			{
				Statement: `\dRp+ testpub1_forschema
                               Publication testpub1_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "pub_test1"
CREATE PUBLICATION testpub2_forschema FOR TABLES IN SCHEMA pub_test1, pub_test2, pub_test3;`,
			},
			{
				Statement: `\dRp+ testpub2_forschema
                               Publication testpub2_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "pub_test1"
    "pub_test2"
    "pub_test3"
CREATE PUBLICATION testpub3_forschema FOR TABLES IN SCHEMA CURRENT_SCHEMA;`,
			},
			{
				Statement: `CREATE PUBLICATION testpub4_forschema FOR TABLES IN SCHEMA "CURRENT_SCHEMA";`,
			},
			{
				Statement: `CREATE PUBLICATION testpub5_forschema FOR TABLES IN SCHEMA CURRENT_SCHEMA, "CURRENT_SCHEMA";`,
			},
			{
				Statement: `CREATE PUBLICATION testpub6_forschema FOR TABLES IN SCHEMA "CURRENT_SCHEMA", CURRENT_SCHEMA;`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_fortable FOR TABLE "CURRENT_SCHEMA"."CURRENT_SCHEMA";`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `\dRp+ testpub3_forschema
                               Publication testpub3_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "public"
\dRp+ testpub4_forschema
                               Publication testpub4_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "CURRENT_SCHEMA"
\dRp+ testpub5_forschema
                               Publication testpub5_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "CURRENT_SCHEMA"
    "public"
\dRp+ testpub6_forschema
                               Publication testpub6_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "CURRENT_SCHEMA"
    "public"
\dRp+ testpub_fortable
                                Publication testpub_fortable
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables:
    "CURRENT_SCHEMA.CURRENT_SCHEMA"
SET SEARCH_PATH='';`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_forschema FOR TABLES IN SCHEMA CURRENT_SCHEMA;`,
				ErrorString: `no schema has been selected for CURRENT_SCHEMA`,
			},
			{
				Statement: `RESET SEARCH_PATH;`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_forschema1 FOR CURRENT_SCHEMA;`,
				ErrorString: `invalid publication object list`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_forschema1 FOR TABLE CURRENT_SCHEMA;`,
				ErrorString: `syntax error at or near "CURRENT_SCHEMA"`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_forschema FOR TABLES IN SCHEMA non_existent_schema;`,
				ErrorString: `schema "non_existent_schema" does not exist`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub_forschema FOR TABLES IN SCHEMA pg_catalog;`,
				ErrorString: `cannot add schema "pg_catalog" to publication`,
			},
			{
				Statement:   `CREATE PUBLICATION testpub1_forschema1 FOR TABLES IN SCHEMA testpub_view;`,
				ErrorString: `schema "testpub_view" does not exist`,
			},
			{
				Statement: `DROP SCHEMA pub_test3;`,
			},
			{
				Statement: `\dRp+ testpub2_forschema
                               Publication testpub2_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "pub_test1"
    "pub_test2"
ALTER SCHEMA pub_test1 RENAME to pub_test1_renamed;`,
			},
			{
				Statement: `\dRp+ testpub2_forschema
                               Publication testpub2_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "pub_test1_renamed"
    "pub_test2"
ALTER SCHEMA pub_test1_renamed RENAME to pub_test1;`,
			},
			{
				Statement: `\dRp+ testpub2_forschema
                               Publication testpub2_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "pub_test1"
    "pub_test2"
ALTER PUBLICATION testpub1_forschema ADD TABLES IN SCHEMA pub_test2;`,
			},
			{
				Statement: `\dRp+ testpub1_forschema
                               Publication testpub1_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "pub_test1"
    "pub_test2"
ALTER PUBLICATION testpub1_forschema ADD TABLES IN SCHEMA non_existent_schema;`,
				ErrorString: `schema "non_existent_schema" does not exist`,
			},
			{
				Statement: `\dRp+ testpub1_forschema
                               Publication testpub1_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "pub_test1"
    "pub_test2"
ALTER PUBLICATION testpub1_forschema ADD TABLES IN SCHEMA pub_test1;`,
				ErrorString: `schema "pub_test1" is already member of publication "testpub1_forschema"`,
			},
			{
				Statement: `\dRp+ testpub1_forschema
                               Publication testpub1_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "pub_test1"
    "pub_test2"
ALTER PUBLICATION testpub1_forschema DROP TABLES IN SCHEMA pub_test2;`,
			},
			{
				Statement: `\dRp+ testpub1_forschema
                               Publication testpub1_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "pub_test1"
ALTER PUBLICATION testpub1_forschema DROP TABLES IN SCHEMA pub_test2;`,
				ErrorString: `tables from schema "pub_test2" are not part of the publication`,
			},
			{
				Statement: `\dRp+ testpub1_forschema
                               Publication testpub1_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "pub_test1"
ALTER PUBLICATION testpub1_forschema DROP TABLES IN SCHEMA non_existent_schema;`,
				ErrorString: `schema "non_existent_schema" does not exist`,
			},
			{
				Statement: `\dRp+ testpub1_forschema
                               Publication testpub1_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "pub_test1"
ALTER PUBLICATION testpub1_forschema DROP TABLES IN SCHEMA pub_test1;`,
			},
			{
				Statement: `\dRp+ testpub1_forschema
                               Publication testpub1_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
(1 row)
ALTER PUBLICATION testpub1_forschema SET TABLES IN SCHEMA pub_test1, pub_test2;`,
			},
			{
				Statement: `\dRp+ testpub1_forschema
                               Publication testpub1_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "pub_test1"
    "pub_test2"
ALTER PUBLICATION testpub1_forschema SET TABLES IN SCHEMA non_existent_schema;`,
				ErrorString: `schema "non_existent_schema" does not exist`,
			},
			{
				Statement: `\dRp+ testpub1_forschema
                               Publication testpub1_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "pub_test1"
    "pub_test2"
ALTER PUBLICATION testpub1_forschema SET TABLES IN SCHEMA pub_test1, pub_test1;`,
			},
			{
				Statement: `\dRp+ testpub1_forschema
                               Publication testpub1_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "pub_test1"
ALTER PUBLICATION testpub1_forschema ADD TABLES IN SCHEMA foo (a, b);`,
				ErrorString: `syntax error at or near "("`,
			},
			{
				Statement:   `ALTER PUBLICATION testpub1_forschema ADD TABLES IN SCHEMA foo, bar (a, b);`,
				ErrorString: `column specification not allowed for schema`,
			},
			{
				Statement: `ALTER PUBLICATION testpub2_forschema DROP TABLES IN SCHEMA pub_test1;`,
			},
			{
				Statement: `DROP PUBLICATION testpub3_forschema, testpub4_forschema, testpub5_forschema, testpub6_forschema, testpub_fortable;`,
			},
			{
				Statement: `DROP SCHEMA "CURRENT_SCHEMA" CASCADE;`,
			},
			{
				Statement: `INSERT INTO pub_test1.tbl VALUES(1, 'test');`,
			},
			{
				Statement:   `UPDATE pub_test1.tbl SET id = 2;`,
				ErrorString: `cannot update table "tbl" because it does not have a replica identity and publishes updates`,
			},
			{
				Statement: `ALTER PUBLICATION testpub1_forschema DROP TABLES IN SCHEMA pub_test1;`,
			},
			{
				Statement: `UPDATE pub_test1.tbl SET id = 2;`,
			},
			{
				Statement: `ALTER PUBLICATION testpub1_forschema SET TABLES IN SCHEMA pub_test1;`,
			},
			{
				Statement:   `UPDATE pub_test1.tbl SET id = 2;`,
				ErrorString: `cannot update table "tbl" because it does not have a replica identity and publishes updates`,
			},
			{
				Statement: `CREATE SCHEMA pub_testpart1;`,
			},
			{
				Statement: `CREATE SCHEMA pub_testpart2;`,
			},
			{
				Statement: `CREATE TABLE pub_testpart1.parent1 (a int) partition by list (a);`,
			},
			{
				Statement: `CREATE TABLE pub_testpart2.child_parent1 partition of pub_testpart1.parent1 for values in (1);`,
			},
			{
				Statement: `INSERT INTO pub_testpart2.child_parent1 values(1);`,
			},
			{
				Statement: `UPDATE pub_testpart2.child_parent1 set a = 1;`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpubpart_forschema FOR TABLES IN SCHEMA pub_testpart1;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement:   `UPDATE pub_testpart1.parent1 set a = 1;`,
				ErrorString: `cannot update table "child_parent1" because it does not have a replica identity and publishes updates`,
			},
			{
				Statement:   `UPDATE pub_testpart2.child_parent1 set a = 1;`,
				ErrorString: `cannot update table "child_parent1" because it does not have a replica identity and publishes updates`,
			},
			{
				Statement: `DROP PUBLICATION testpubpart_forschema;`,
			},
			{
				Statement: `CREATE TABLE pub_testpart2.parent2 (a int) partition by list (a);`,
			},
			{
				Statement: `CREATE TABLE pub_testpart1.child_parent2 partition of pub_testpart2.parent2 for values in (1);`,
			},
			{
				Statement: `INSERT INTO pub_testpart1.child_parent2 values(1);`,
			},
			{
				Statement: `UPDATE pub_testpart1.child_parent2 set a = 1;`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpubpart_forschema FOR TABLES IN SCHEMA pub_testpart2;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement:   `UPDATE pub_testpart2.child_parent1 set a = 1;`,
				ErrorString: `cannot update table "child_parent1" because it does not have a replica identity and publishes updates`,
			},
			{
				Statement:   `UPDATE pub_testpart2.parent2 set a = 1;`,
				ErrorString: `cannot update table "child_parent2" because it does not have a replica identity and publishes updates`,
			},
			{
				Statement:   `UPDATE pub_testpart1.child_parent2 set a = 1;`,
				ErrorString: `cannot update table "child_parent2" because it does not have a replica identity and publishes updates`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub3_forschema;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `\dRp+ testpub3_forschema
                               Publication testpub3_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
(1 row)
ALTER PUBLICATION testpub3_forschema SET TABLES IN SCHEMA pub_test1;`,
			},
			{
				Statement: `\dRp+ testpub3_forschema
                               Publication testpub3_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables from schemas:
    "pub_test1"
SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_forschema_fortable FOR TABLES IN SCHEMA pub_test1, TABLE pub_test2.tbl1;`,
			},
			{
				Statement: `CREATE PUBLICATION testpub_fortable_forschema FOR TABLE pub_test2.tbl1, TABLES IN SCHEMA pub_test1;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `\dRp+ testpub_forschema_fortable
                           Publication testpub_forschema_fortable
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables:
    "pub_test2.tbl1"
Tables from schemas:
    "pub_test1"
\dRp+ testpub_fortable_forschema
                           Publication testpub_fortable_forschema
          Owner           | All tables | Inserts | Updates | Deletes | Truncates | Via root 
--------------------------+------------+---------+---------+---------+-----------+----------
 regress_publication_user | f          | t       | t       | t       | t         | f
Tables:
    "pub_test2.tbl1"
Tables from schemas:
    "pub_test1"
CREATE PUBLICATION testpub_error FOR pub_test2.tbl1;`,
				ErrorString: `invalid publication object list`,
			},
			{
				Statement: `DROP VIEW testpub_view;`,
			},
			{
				Statement: `DROP PUBLICATION testpub_default;`,
			},
			{
				Statement: `DROP PUBLICATION testpib_ins_trunct;`,
			},
			{
				Statement: `DROP PUBLICATION testpub_fortbl;`,
			},
			{
				Statement: `DROP PUBLICATION testpub1_forschema;`,
			},
			{
				Statement: `DROP PUBLICATION testpub2_forschema;`,
			},
			{
				Statement: `DROP PUBLICATION testpub3_forschema;`,
			},
			{
				Statement: `DROP PUBLICATION testpub_forschema_fortable;`,
			},
			{
				Statement: `DROP PUBLICATION testpub_fortable_forschema;`,
			},
			{
				Statement: `DROP PUBLICATION testpubpart_forschema;`,
			},
			{
				Statement: `DROP SCHEMA pub_test CASCADE;`,
			},
			{
				Statement: `DROP SCHEMA pub_test1 CASCADE;`,
			},
			{
				Statement: `DROP SCHEMA pub_test2 CASCADE;`,
			},
			{
				Statement: `DROP SCHEMA pub_testpart1 CASCADE;`,
			},
			{
				Statement: `DROP SCHEMA pub_testpart2 CASCADE;`,
			},
			{
				Statement: `SET client_min_messages = 'ERROR';`,
			},
			{
				Statement: `CREATE SCHEMA sch1;`,
			},
			{
				Statement: `CREATE SCHEMA sch2;`,
			},
			{
				Statement: `CREATE TABLE sch1.tbl1 (a int) PARTITION BY RANGE(a);`,
			},
			{
				Statement: `CREATE TABLE sch2.tbl1_part1 PARTITION OF sch1.tbl1 FOR VALUES FROM (1) to (10);`,
			},
			{
				Statement: `CREATE PUBLICATION pub FOR TABLES IN SCHEMA sch2 WITH (PUBLISH_VIA_PARTITION_ROOT=1);`,
			},
			{
				Statement: `SELECT * FROM pg_publication_tables;`,
				Results:   []sql.Row{{`pub`, `sch2`, `tbl1_part1`, `{a}`, ``}},
			},
			{
				Statement: `DROP PUBLICATION pub;`,
			},
			{
				Statement: `CREATE PUBLICATION pub FOR TABLE sch2.tbl1_part1 WITH (PUBLISH_VIA_PARTITION_ROOT=1);`,
			},
			{
				Statement: `SELECT * FROM pg_publication_tables;`,
				Results:   []sql.Row{{`pub`, `sch2`, `tbl1_part1`, `{a}`, ``}},
			},
			{
				Statement: `ALTER PUBLICATION pub ADD TABLE sch1.tbl1;`,
			},
			{
				Statement: `SELECT * FROM pg_publication_tables;`,
				Results:   []sql.Row{{`pub`, `sch1`, `tbl1`, `{a}`, ``}},
			},
			{
				Statement: `DROP PUBLICATION pub;`,
			},
			{
				Statement: `CREATE PUBLICATION pub FOR TABLES IN SCHEMA sch2 WITH (PUBLISH_VIA_PARTITION_ROOT=0);`,
			},
			{
				Statement: `SELECT * FROM pg_publication_tables;`,
				Results:   []sql.Row{{`pub`, `sch2`, `tbl1_part1`, `{a}`, ``}},
			},
			{
				Statement: `DROP PUBLICATION pub;`,
			},
			{
				Statement: `CREATE PUBLICATION pub FOR TABLE sch2.tbl1_part1 WITH (PUBLISH_VIA_PARTITION_ROOT=0);`,
			},
			{
				Statement: `SELECT * FROM pg_publication_tables;`,
				Results:   []sql.Row{{`pub`, `sch2`, `tbl1_part1`, `{a}`, ``}},
			},
			{
				Statement: `ALTER PUBLICATION pub ADD TABLE sch1.tbl1;`,
			},
			{
				Statement: `SELECT * FROM pg_publication_tables;`,
				Results:   []sql.Row{{`pub`, `sch2`, `tbl1_part1`, `{a}`, ``}},
			},
			{
				Statement: `DROP PUBLICATION pub;`,
			},
			{
				Statement: `DROP TABLE sch2.tbl1_part1;`,
			},
			{
				Statement: `DROP TABLE sch1.tbl1;`,
			},
			{
				Statement: `CREATE TABLE sch1.tbl1 (a int) PARTITION BY RANGE(a);`,
			},
			{
				Statement: `CREATE TABLE sch1.tbl1_part1 PARTITION OF sch1.tbl1 FOR VALUES FROM (1) to (10);`,
			},
			{
				Statement: `CREATE TABLE sch1.tbl1_part2 PARTITION OF sch1.tbl1 FOR VALUES FROM (10) to (20);`,
			},
			{
				Statement: `CREATE TABLE sch1.tbl1_part3 (a int) PARTITION BY RANGE(a);`,
			},
			{
				Statement: `ALTER TABLE sch1.tbl1 ATTACH PARTITION sch1.tbl1_part3 FOR VALUES FROM (20) to (30);`,
			},
			{
				Statement: `CREATE PUBLICATION pub FOR TABLES IN SCHEMA sch1 WITH (PUBLISH_VIA_PARTITION_ROOT=1);`,
			},
			{
				Statement: `SELECT * FROM pg_publication_tables;`,
				Results:   []sql.Row{{`pub`, `sch1`, `tbl1`, `{a}`, ``}},
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `DROP PUBLICATION pub;`,
			},
			{
				Statement: `DROP TABLE sch1.tbl1;`,
			},
			{
				Statement: `DROP SCHEMA sch1 cascade;`,
			},
			{
				Statement: `DROP SCHEMA sch2 cascade;`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `DROP ROLE regress_publication_user, regress_publication_user2;`,
			},
			{
				Statement: `DROP ROLE regress_publication_user_dummy;`,
			},
		},
	})
}
