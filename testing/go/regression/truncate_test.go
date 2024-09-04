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

func TestTruncate(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_truncate)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_truncate,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE truncate_a (col1 integer primary key);`,
			},
			{
				Statement: `INSERT INTO truncate_a VALUES (1);`,
			},
			{
				Statement: `INSERT INTO truncate_a VALUES (2);`,
			},
			{
				Statement: `SELECT * FROM truncate_a;`,
				Results:   []sql.Row{{1}, {2}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `TRUNCATE truncate_a;`,
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `SELECT * FROM truncate_a;`,
				Results:   []sql.Row{{1}, {2}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `TRUNCATE truncate_a;`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement: `SELECT * FROM truncate_a;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `CREATE TABLE trunc_b (a int REFERENCES truncate_a);`,
			},
			{
				Statement: `CREATE TABLE trunc_c (a serial PRIMARY KEY);`,
			},
			{
				Statement: `CREATE TABLE trunc_d (a int REFERENCES trunc_c);`,
			},
			{
				Statement: `CREATE TABLE trunc_e (a int REFERENCES truncate_a, b int REFERENCES trunc_c);`,
			},
			{
				Statement:   `TRUNCATE TABLE truncate_a;		-- fail`,
				ErrorString: `cannot truncate a table referenced in a foreign key constraint`,
			},
			{
				Statement:   `TRUNCATE TABLE truncate_a,trunc_b;		-- fail`,
				ErrorString: `cannot truncate a table referenced in a foreign key constraint`,
			},
			{
				Statement: `TRUNCATE TABLE truncate_a,trunc_b,trunc_e;	-- ok`,
			},
			{
				Statement:   `TRUNCATE TABLE truncate_a,trunc_e;		-- fail`,
				ErrorString: `cannot truncate a table referenced in a foreign key constraint`,
			},
			{
				Statement:   `TRUNCATE TABLE trunc_c;		-- fail`,
				ErrorString: `cannot truncate a table referenced in a foreign key constraint`,
			},
			{
				Statement:   `TRUNCATE TABLE trunc_c,trunc_d;		-- fail`,
				ErrorString: `cannot truncate a table referenced in a foreign key constraint`,
			},
			{
				Statement: `TRUNCATE TABLE trunc_c,trunc_d,trunc_e;	-- ok`,
			},
			{
				Statement:   `TRUNCATE TABLE trunc_c,trunc_d,trunc_e,truncate_a;	-- fail`,
				ErrorString: `cannot truncate a table referenced in a foreign key constraint`,
			},
			{
				Statement: `TRUNCATE TABLE trunc_c,trunc_d,trunc_e,truncate_a,trunc_b;	-- ok`,
			},
			{
				Statement:   `TRUNCATE TABLE truncate_a RESTRICT; -- fail`,
				ErrorString: `cannot truncate a table referenced in a foreign key constraint`,
			},
			{
				Statement: `TRUNCATE TABLE truncate_a CASCADE;  -- ok`,
			},
			{
				Statement: `ALTER TABLE truncate_a ADD FOREIGN KEY (col1) REFERENCES trunc_c;`,
			},
			{
				Statement: `INSERT INTO trunc_c VALUES (1);`,
			},
			{
				Statement: `INSERT INTO truncate_a VALUES (1);`,
			},
			{
				Statement: `INSERT INTO trunc_b VALUES (1);`,
			},
			{
				Statement: `INSERT INTO trunc_d VALUES (1);`,
			},
			{
				Statement: `INSERT INTO trunc_e VALUES (1,1);`,
			},
			{
				Statement:   `TRUNCATE TABLE trunc_c;`,
				ErrorString: `cannot truncate a table referenced in a foreign key constraint`,
			},
			{
				Statement:   `TRUNCATE TABLE trunc_c,truncate_a;`,
				ErrorString: `cannot truncate a table referenced in a foreign key constraint`,
			},
			{
				Statement:   `TRUNCATE TABLE trunc_c,truncate_a,trunc_d;`,
				ErrorString: `cannot truncate a table referenced in a foreign key constraint`,
			},
			{
				Statement:   `TRUNCATE TABLE trunc_c,truncate_a,trunc_d,trunc_e;`,
				ErrorString: `cannot truncate a table referenced in a foreign key constraint`,
			},
			{
				Statement: `TRUNCATE TABLE trunc_c,truncate_a,trunc_d,trunc_e,trunc_b;`,
			},
			{
				Statement: `SELECT * FROM truncate_a
   UNION ALL
 SELECT * FROM trunc_c
   UNION ALL
 SELECT * FROM trunc_b
   UNION ALL
 SELECT * FROM trunc_d;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT * FROM trunc_e;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `INSERT INTO trunc_c VALUES (1);`,
			},
			{
				Statement: `INSERT INTO truncate_a VALUES (1);`,
			},
			{
				Statement: `INSERT INTO trunc_b VALUES (1);`,
			},
			{
				Statement: `INSERT INTO trunc_d VALUES (1);`,
			},
			{
				Statement: `INSERT INTO trunc_e VALUES (1,1);`,
			},
			{
				Statement: `TRUNCATE TABLE trunc_c CASCADE;  -- ok`,
			},
			{
				Statement: `SELECT * FROM truncate_a
   UNION ALL
 SELECT * FROM trunc_c
   UNION ALL
 SELECT * FROM trunc_b
   UNION ALL
 SELECT * FROM trunc_d;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT * FROM trunc_e;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP TABLE truncate_a,trunc_c,trunc_b,trunc_d,trunc_e CASCADE;`,
			},
			{
				Statement: `CREATE TABLE trunc_f (col1 integer primary key);`,
			},
			{
				Statement: `INSERT INTO trunc_f VALUES (1);`,
			},
			{
				Statement: `INSERT INTO trunc_f VALUES (2);`,
			},
			{
				Statement: `CREATE TABLE trunc_fa (col2a text) INHERITS (trunc_f);`,
			},
			{
				Statement: `INSERT INTO trunc_fa VALUES (3, 'three');`,
			},
			{
				Statement: `CREATE TABLE trunc_fb (col2b int) INHERITS (trunc_f);`,
			},
			{
				Statement: `INSERT INTO trunc_fb VALUES (4, 444);`,
			},
			{
				Statement: `CREATE TABLE trunc_faa (col3 text) INHERITS (trunc_fa);`,
			},
			{
				Statement: `INSERT INTO trunc_faa VALUES (5, 'five', 'FIVE');`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT * FROM trunc_f;`,
				Results:   []sql.Row{{1}, {2}, {3}, {4}, {5}},
			},
			{
				Statement: `TRUNCATE trunc_f;`,
			},
			{
				Statement: `SELECT * FROM trunc_f;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT * FROM trunc_f;`,
				Results:   []sql.Row{{1}, {2}, {3}, {4}, {5}},
			},
			{
				Statement: `TRUNCATE ONLY trunc_f;`,
			},
			{
				Statement: `SELECT * FROM trunc_f;`,
				Results:   []sql.Row{{3}, {4}, {5}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT * FROM trunc_f;`,
				Results:   []sql.Row{{1}, {2}, {3}, {4}, {5}},
			},
			{
				Statement: `SELECT * FROM trunc_fa;`,
				Results:   []sql.Row{{3, `three`}, {5, `five`}},
			},
			{
				Statement: `SELECT * FROM trunc_faa;`,
				Results:   []sql.Row{{5, `five`, `FIVE`}},
			},
			{
				Statement: `TRUNCATE ONLY trunc_fb, ONLY trunc_fa;`,
			},
			{
				Statement: `SELECT * FROM trunc_f;`,
				Results:   []sql.Row{{1}, {2}, {5}},
			},
			{
				Statement: `SELECT * FROM trunc_fa;`,
				Results:   []sql.Row{{5, `five`}},
			},
			{
				Statement: `SELECT * FROM trunc_faa;`,
				Results:   []sql.Row{{5, `five`, `FIVE`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SELECT * FROM trunc_f;`,
				Results:   []sql.Row{{1}, {2}, {3}, {4}, {5}},
			},
			{
				Statement: `SELECT * FROM trunc_fa;`,
				Results:   []sql.Row{{3, `three`}, {5, `five`}},
			},
			{
				Statement: `SELECT * FROM trunc_faa;`,
				Results:   []sql.Row{{5, `five`, `FIVE`}},
			},
			{
				Statement: `TRUNCATE ONLY trunc_fb, trunc_fa;`,
			},
			{
				Statement: `SELECT * FROM trunc_f;`,
				Results:   []sql.Row{{1}, {2}},
			},
			{
				Statement: `SELECT * FROM trunc_fa;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT * FROM trunc_faa;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `DROP TABLE trunc_f CASCADE;`,
			},
			{
				Statement: `CREATE TABLE trunc_trigger_test (f1 int, f2 text, f3 text);`,
			},
			{
				Statement: `CREATE TABLE trunc_trigger_log (tgop text, tglevel text, tgwhen text,
        tgargv text, tgtable name, rowcount bigint);`,
			},
			{
				Statement: `CREATE FUNCTION trunctrigger() RETURNS trigger as $$
declare c bigint;`,
			},
			{
				Statement: `begin
    execute 'select count(*) from ' || quote_ident(tg_table_name) into c;`,
			},
			{
				Statement: `    insert into trunc_trigger_log values
      (TG_OP, TG_LEVEL, TG_WHEN, TG_ARGV[0], tg_table_name, c);`,
			},
			{
				Statement: `    return null;`,
			},
			{
				Statement: `end;`,
			},
			{
				Statement: `$$ LANGUAGE plpgsql;`,
			},
			{
				Statement: `INSERT INTO trunc_trigger_test VALUES(1, 'foo', 'bar'), (2, 'baz', 'quux');`,
			},
			{
				Statement: `CREATE TRIGGER t
BEFORE TRUNCATE ON trunc_trigger_test
FOR EACH STATEMENT
EXECUTE PROCEDURE trunctrigger('before trigger truncate');`,
			},
			{
				Statement: `SELECT count(*) as "Row count in test table" FROM trunc_trigger_test;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT * FROM trunc_trigger_log;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `TRUNCATE trunc_trigger_test;`,
			},
			{
				Statement: `SELECT count(*) as "Row count in test table" FROM trunc_trigger_test;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT * FROM trunc_trigger_log;`,
				Results:   []sql.Row{{`TRUNCATE`, `STATEMENT`, `BEFORE`, `before trigger truncate`, `trunc_trigger_test`, 2}},
			},
			{
				Statement: `DROP TRIGGER t ON trunc_trigger_test;`,
			},
			{
				Statement: `truncate trunc_trigger_log;`,
			},
			{
				Statement: `INSERT INTO trunc_trigger_test VALUES(1, 'foo', 'bar'), (2, 'baz', 'quux');`,
			},
			{
				Statement: `CREATE TRIGGER tt
AFTER TRUNCATE ON trunc_trigger_test
FOR EACH STATEMENT
EXECUTE PROCEDURE trunctrigger('after trigger truncate');`,
			},
			{
				Statement: `SELECT count(*) as "Row count in test table" FROM trunc_trigger_test;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `SELECT * FROM trunc_trigger_log;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `TRUNCATE trunc_trigger_test;`,
			},
			{
				Statement: `SELECT count(*) as "Row count in test table" FROM trunc_trigger_test;`,
				Results:   []sql.Row{{0}},
			},
			{
				Statement: `SELECT * FROM trunc_trigger_log;`,
				Results:   []sql.Row{{`TRUNCATE`, `STATEMENT`, `AFTER`, `after trigger truncate`, `trunc_trigger_test`, 0}},
			},
			{
				Statement: `DROP TABLE trunc_trigger_test;`,
			},
			{
				Statement: `DROP TABLE trunc_trigger_log;`,
			},
			{
				Statement: `DROP FUNCTION trunctrigger();`,
			},
			{
				Statement: `CREATE SEQUENCE truncate_a_id1 START WITH 33;`,
			},
			{
				Statement: `CREATE TABLE truncate_a (id serial,
                         id1 integer default nextval('truncate_a_id1'));`,
			},
			{
				Statement: `ALTER SEQUENCE truncate_a_id1 OWNED BY truncate_a.id1;`,
			},
			{
				Statement: `INSERT INTO truncate_a DEFAULT VALUES;`,
			},
			{
				Statement: `INSERT INTO truncate_a DEFAULT VALUES;`,
			},
			{
				Statement: `SELECT * FROM truncate_a;`,
				Results:   []sql.Row{{1, 33}, {2, 34}},
			},
			{
				Statement: `TRUNCATE truncate_a;`,
			},
			{
				Statement: `INSERT INTO truncate_a DEFAULT VALUES;`,
			},
			{
				Statement: `INSERT INTO truncate_a DEFAULT VALUES;`,
			},
			{
				Statement: `SELECT * FROM truncate_a;`,
				Results:   []sql.Row{{3, 35}, {4, 36}},
			},
			{
				Statement: `TRUNCATE truncate_a RESTART IDENTITY;`,
			},
			{
				Statement: `INSERT INTO truncate_a DEFAULT VALUES;`,
			},
			{
				Statement: `INSERT INTO truncate_a DEFAULT VALUES;`,
			},
			{
				Statement: `SELECT * FROM truncate_a;`,
				Results:   []sql.Row{{1, 33}, {2, 34}},
			},
			{
				Statement: `CREATE TABLE truncate_b (id int GENERATED ALWAYS AS IDENTITY (START WITH 44));`,
			},
			{
				Statement: `INSERT INTO truncate_b DEFAULT VALUES;`,
			},
			{
				Statement: `INSERT INTO truncate_b DEFAULT VALUES;`,
			},
			{
				Statement: `SELECT * FROM truncate_b;`,
				Results:   []sql.Row{{44}, {45}},
			},
			{
				Statement: `TRUNCATE truncate_b;`,
			},
			{
				Statement: `INSERT INTO truncate_b DEFAULT VALUES;`,
			},
			{
				Statement: `INSERT INTO truncate_b DEFAULT VALUES;`,
			},
			{
				Statement: `SELECT * FROM truncate_b;`,
				Results:   []sql.Row{{46}, {47}},
			},
			{
				Statement: `TRUNCATE truncate_b RESTART IDENTITY;`,
			},
			{
				Statement: `INSERT INTO truncate_b DEFAULT VALUES;`,
			},
			{
				Statement: `INSERT INTO truncate_b DEFAULT VALUES;`,
			},
			{
				Statement: `SELECT * FROM truncate_b;`,
				Results:   []sql.Row{{44}, {45}},
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `TRUNCATE truncate_a RESTART IDENTITY;`,
			},
			{
				Statement: `INSERT INTO truncate_a DEFAULT VALUES;`,
			},
			{
				Statement: `SELECT * FROM truncate_a;`,
				Results:   []sql.Row{{1, 33}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement: `INSERT INTO truncate_a DEFAULT VALUES;`,
			},
			{
				Statement: `INSERT INTO truncate_a DEFAULT VALUES;`,
			},
			{
				Statement: `SELECT * FROM truncate_a;`,
				Results:   []sql.Row{{1, 33}, {2, 34}, {3, 35}, {4, 36}},
			},
			{
				Statement: `DROP TABLE truncate_a;`,
			},
			{
				Statement:   `SELECT nextval('truncate_a_id1'); -- fail, seq should have been dropped`,
				ErrorString: `relation "truncate_a_id1" does not exist`,
			},
			{
				Statement: `CREATE TABLE truncparted (a int, b char) PARTITION BY LIST (a);`,
			},
			{
				Statement:   `TRUNCATE ONLY truncparted;`,
				ErrorString: `cannot truncate only a partitioned table`,
			},
			{
				Statement: `CREATE TABLE truncparted1 PARTITION OF truncparted FOR VALUES IN (1);`,
			},
			{
				Statement: `INSERT INTO truncparted VALUES (1, 'a');`,
			},
			{
				Statement:   `TRUNCATE ONLY truncparted;`,
				ErrorString: `cannot truncate only a partitioned table`,
			},
			{
				Statement: `TRUNCATE truncparted;`,
			},
			{
				Statement: `DROP TABLE truncparted;`,
			},
			{
				Statement: `CREATE FUNCTION tp_ins_data() RETURNS void LANGUAGE plpgsql AS $$
  BEGIN
	INSERT INTO truncprim VALUES (1), (100), (150);`,
			},
			{
				Statement: `	INSERT INTO truncpart VALUES (1), (100), (150);`,
			},
			{
				Statement: `  END
$$;`,
			},
			{
				Statement: `CREATE FUNCTION tp_chk_data(OUT pktb regclass, OUT pkval int, OUT fktb regclass, OUT fkval int)
  RETURNS SETOF record LANGUAGE plpgsql AS $$
  BEGIN
    RETURN QUERY SELECT
      pk.tableoid::regclass, pk.a, fk.tableoid::regclass, fk.a
    FROM truncprim pk FULL JOIN truncpart fk USING (a)
    ORDER BY 2, 4;`,
			},
			{
				Statement: `  END
$$;`,
			},
			{
				Statement: `CREATE TABLE truncprim (a int PRIMARY KEY);`,
			},
			{
				Statement: `CREATE TABLE truncpart (a int REFERENCES truncprim)
  PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE truncpart_1 PARTITION OF truncpart FOR VALUES FROM (0) TO (100);`,
			},
			{
				Statement: `CREATE TABLE truncpart_2 PARTITION OF truncpart FOR VALUES FROM (100) TO (200)
  PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE truncpart_2_1 PARTITION OF truncpart_2 FOR VALUES FROM (100) TO (150);`,
			},
			{
				Statement: `CREATE TABLE truncpart_2_d PARTITION OF truncpart_2 DEFAULT;`,
			},
			{
				Statement:   `TRUNCATE TABLE truncprim;	-- should fail`,
				ErrorString: `cannot truncate a table referenced in a foreign key constraint`,
			},
			{
				Statement: `select tp_ins_data();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `TRUNCATE TABLE truncprim, truncpart;`,
			},
			{
				Statement: `select * from tp_chk_data();`,
				Results:   []sql.Row{},
			},
			{
				Statement: `select tp_ins_data();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `TRUNCATE TABLE truncprim CASCADE;`,
			},
			{
				Statement: `SELECT * FROM tp_chk_data();`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT tp_ins_data();`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `TRUNCATE TABLE truncpart;`,
			},
			{
				Statement: `SELECT * FROM tp_chk_data();`,
				Results:   []sql.Row{{`truncprim`, 1, ``, ``}, {`truncprim`, 100, ``, ``}, {`truncprim`, 150, ``, ``}},
			},
			{
				Statement: `DROP TABLE truncprim, truncpart;`,
			},
			{
				Statement: `DROP FUNCTION tp_ins_data(), tp_chk_data();`,
			},
			{
				Statement: `CREATE TABLE trunc_a (a INT PRIMARY KEY) PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE trunc_a1 PARTITION OF trunc_a FOR VALUES FROM (0) TO (10);`,
			},
			{
				Statement: `CREATE TABLE trunc_a2 PARTITION OF trunc_a FOR VALUES FROM (10) TO (20)
  PARTITION BY RANGE (a);`,
			},
			{
				Statement: `CREATE TABLE trunc_a21 PARTITION OF trunc_a2 FOR VALUES FROM (10) TO (12);`,
			},
			{
				Statement: `CREATE TABLE trunc_a22 PARTITION OF trunc_a2 FOR VALUES FROM (12) TO (16);`,
			},
			{
				Statement: `CREATE TABLE trunc_a2d PARTITION OF trunc_a2 DEFAULT;`,
			},
			{
				Statement: `CREATE TABLE trunc_a3 PARTITION OF trunc_a FOR VALUES FROM (20) TO (30);`,
			},
			{
				Statement: `INSERT INTO trunc_a VALUES (0), (5), (10), (15), (20), (25);`,
			},
			{
				Statement: `CREATE TABLE ref_b (
    b INT PRIMARY KEY,
    a INT REFERENCES trunc_a(a) ON DELETE CASCADE
);`,
			},
			{
				Statement: `INSERT INTO ref_b VALUES (10, 0), (50, 5), (100, 10), (150, 15);`,
			},
			{
				Statement: `TRUNCATE TABLE trunc_a1 CASCADE;`,
			},
			{
				Statement: `SELECT a FROM ref_b;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `DROP TABLE ref_b;`,
			},
			{
				Statement: `CREATE TABLE ref_c (
    c INT PRIMARY KEY,
    a INT REFERENCES trunc_a(a) ON DELETE CASCADE
) PARTITION BY RANGE (c);`,
			},
			{
				Statement: `CREATE TABLE ref_c1 PARTITION OF ref_c FOR VALUES FROM (100) TO (200);`,
			},
			{
				Statement: `CREATE TABLE ref_c2 PARTITION OF ref_c FOR VALUES FROM (200) TO (300);`,
			},
			{
				Statement: `INSERT INTO ref_c VALUES (100, 10), (150, 15), (200, 20), (250, 25);`,
			},
			{
				Statement: `TRUNCATE TABLE trunc_a21 CASCADE;`,
			},
			{
				Statement: `SELECT a as "from table ref_c" FROM ref_c;`,
				Results:   []sql.Row{},
			},
			{
				Statement: `SELECT a as "from table trunc_a" FROM trunc_a ORDER BY a;`,
				Results:   []sql.Row{{15}, {20}, {25}},
			},
			{
				Statement: `DROP TABLE trunc_a, ref_c;`,
			},
		},
	})
}
