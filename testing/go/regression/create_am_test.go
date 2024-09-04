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

func TestCreateAm(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_create_am)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_create_am,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE ACCESS METHOD gist2 TYPE INDEX HANDLER gisthandler;`,
			},
			{
				Statement:   `CREATE ACCESS METHOD bogus TYPE INDEX HANDLER int4in;`,
				ErrorString: `function int4in(internal) does not exist`,
			},
			{
				Statement:   `CREATE ACCESS METHOD bogus TYPE INDEX HANDLER heap_tableam_handler;`,
				ErrorString: `function heap_tableam_handler must return type index_am_handler`,
			},
			{
				Statement:   `CREATE INDEX grect2ind2 ON fast_emp4000 USING gist2 (home_base);`,
				ErrorString: `data type box has no default operator class for access method "gist2"`,
			},
			{
				Statement: `CREATE OPERATOR CLASS box_ops DEFAULT
	FOR TYPE box USING gist2 AS
	OPERATOR 1	<<,
	OPERATOR 2	&<,
	OPERATOR 3	&&,
	OPERATOR 4	&>,
	OPERATOR 5	>>,
	OPERATOR 6	~=,
	OPERATOR 7	@>,
	OPERATOR 8	<@,
	OPERATOR 9	&<|,
	OPERATOR 10	<<|,
	OPERATOR 11	|>>,
	OPERATOR 12	|&>,
	FUNCTION 1	gist_box_consistent(internal, box, smallint, oid, internal),
	FUNCTION 2	gist_box_union(internal, internal),
	-- don't need compress, decompress, or fetch functions
	FUNCTION 5	gist_box_penalty(internal, internal, internal),
	FUNCTION 6	gist_box_picksplit(internal, internal),
	FUNCTION 7	gist_box_same(box, box, internal);`,
			},
			{
				Statement: `CREATE INDEX grect2ind2 ON fast_emp4000 USING gist2 (home_base);`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `DROP INDEX grect2ind;`,
			},
			{
				Statement: `SET enable_seqscan = OFF;`,
			},
			{
				Statement: `SET enable_indexscan = ON;`,
			},
			{
				Statement: `SET enable_bitmapscan = OFF;`,
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT * FROM fast_emp4000
    WHERE home_base <@ '(200,200),(2000,1000)'::box
    ORDER BY (home_base[0])[0];`,
				Results: []sql.Row{{`Sort`}, {`Sort Key: ((home_base[0])[0])`}, {`->  Index Only Scan using grect2ind2 on fast_emp4000`}, {`Index Cond: (home_base <@ '(2000,1000),(200,200)'::box)`}},
			},
			{
				Statement: `SELECT * FROM fast_emp4000
    WHERE home_base <@ '(200,200),(2000,1000)'::box
    ORDER BY (home_base[0])[0];`,
				Results: []sql.Row{{`(337,455),(240,359)`}, {`(1444,403),(1346,344)`}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM fast_emp4000 WHERE home_base && '(1000,1000,0,0)'::box;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using grect2ind2 on fast_emp4000`}, {`Index Cond: (home_base && '(1000,1000),(0,0)'::box)`}},
			},
			{
				Statement: `SELECT count(*) FROM fast_emp4000 WHERE home_base && '(1000,1000,0,0)'::box;`,
				Results:   []sql.Row{{2}},
			},
			{
				Statement: `EXPLAIN (COSTS OFF)
SELECT count(*) FROM fast_emp4000 WHERE home_base IS NULL;`,
				Results: []sql.Row{{`Aggregate`}, {`->  Index Only Scan using grect2ind2 on fast_emp4000`}, {`Index Cond: (home_base IS NULL)`}},
			},
			{
				Statement: `SELECT count(*) FROM fast_emp4000 WHERE home_base IS NULL;`,
				Results:   []sql.Row{{278}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement:   `DROP ACCESS METHOD gist2;`,
				ErrorString: `cannot drop access method gist2 because other objects depend on it`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `LOCK TABLE fast_emp4000;`,
			},
			{
				Statement: `DROP ACCESS METHOD gist2 CASCADE;`,
			},
			{
				Statement: `COMMIT;`,
			},
			{
				Statement:   `SET default_table_access_method = '';`,
				ErrorString: `invalid value for parameter "default_table_access_method": ""`,
			},
			{
				Statement:   `SET default_table_access_method = 'I do not exist AM';`,
				ErrorString: `invalid value for parameter "default_table_access_method": "I do not exist AM"`,
			},
			{
				Statement:   `SET default_table_access_method = 'btree';`,
				ErrorString: `access method "btree" is not of type TABLE`,
			},
			{
				Statement: `CREATE ACCESS METHOD heap2 TYPE TABLE HANDLER heap_tableam_handler;`,
			},
			{
				Statement:   `CREATE ACCESS METHOD bogus TYPE TABLE HANDLER int4in;`,
				ErrorString: `function int4in(internal) does not exist`,
			},
			{
				Statement:   `CREATE ACCESS METHOD bogus TYPE TABLE HANDLER bthandler;`,
				ErrorString: `function bthandler must return type table_am_handler`,
			},
			{
				Statement: `SELECT amname, amhandler, amtype FROM pg_am where amtype = 't' ORDER BY 1, 2;`,
				Results:   []sql.Row{{`heap`, `heap_tableam_handler`, true}, {`heap2`, `heap_tableam_handler`, true}},
			},
			{
				Statement: `CREATE TABLE tableam_tbl_heap2(f1 int) USING heap2;`,
			},
			{
				Statement: `INSERT INTO tableam_tbl_heap2 VALUES(1);`,
			},
			{
				Statement: `SELECT f1 FROM tableam_tbl_heap2 ORDER BY f1;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement: `CREATE TABLE tableam_tblas_heap2 USING heap2 AS SELECT * FROM tableam_tbl_heap2;`,
			},
			{
				Statement: `SELECT f1 FROM tableam_tbl_heap2 ORDER BY f1;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement:   `SELECT INTO tableam_tblselectinto_heap2 USING heap2 FROM tableam_tbl_heap2;`,
				ErrorString: `syntax error at or near "USING"`,
			},
			{
				Statement:   `CREATE VIEW tableam_view_heap2 USING heap2 AS SELECT * FROM tableam_tbl_heap2;`,
				ErrorString: `syntax error at or near "USING"`,
			},
			{
				Statement:   `CREATE SEQUENCE tableam_seq_heap2 USING heap2;`,
				ErrorString: `syntax error at or near "USING"`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW tableam_tblmv_heap2 USING heap2 AS SELECT * FROM tableam_tbl_heap2;`,
			},
			{
				Statement: `SELECT f1 FROM tableam_tblmv_heap2 ORDER BY f1;`,
				Results:   []sql.Row{{1}},
			},
			{
				Statement:   `CREATE TABLE tableam_parted_heap2 (a text, b int) PARTITION BY list (a) USING heap2;`,
				ErrorString: `specifying a table access method is not supported on a partitioned table`,
			},
			{
				Statement: `CREATE TABLE tableam_parted_heap2 (a text, b int) PARTITION BY list (a);`,
			},
			{
				Statement: `SET default_table_access_method = 'heap';`,
			},
			{
				Statement: `CREATE TABLE tableam_parted_a_heap2 PARTITION OF tableam_parted_heap2 FOR VALUES IN ('a');`,
			},
			{
				Statement: `SET default_table_access_method = 'heap2';`,
			},
			{
				Statement: `CREATE TABLE tableam_parted_b_heap2 PARTITION OF tableam_parted_heap2 FOR VALUES IN ('b');`,
			},
			{
				Statement: `RESET default_table_access_method;`,
			},
			{
				Statement: `CREATE TABLE tableam_parted_c_heap2 PARTITION OF tableam_parted_heap2 FOR VALUES IN ('c') USING heap;`,
			},
			{
				Statement: `CREATE TABLE tableam_parted_d_heap2 PARTITION OF tableam_parted_heap2 FOR VALUES IN ('d') USING heap2;`,
			},
			{
				Statement: `SELECT
    pc.relkind,
    pa.amname,
    CASE WHEN relkind = 't' THEN
        (SELECT 'toast for ' || relname::regclass FROM pg_class pcm WHERE pcm.reltoastrelid = pc.oid)
    ELSE
        relname::regclass::text
    END COLLATE "C" AS relname
FROM pg_class AS pc,
    pg_am AS pa
WHERE pa.oid = pc.relam
   AND pa.amname = 'heap2'
ORDER BY 3, 1, 2;`,
				Results: []sql.Row{{`r`, `heap2`, `tableam_parted_b_heap2`}, {`r`, `heap2`, `tableam_parted_d_heap2`}, {`r`, `heap2`, `tableam_tbl_heap2`}, {`r`, `heap2`, `tableam_tblas_heap2`}, {`m`, `heap2`, `tableam_tblmv_heap2`}, {true, `heap2`, `toast for tableam_parted_b_heap2`}, {true, `heap2`, `toast for tableam_parted_d_heap2`}},
			},
			{
				Statement: `SELECT pg_describe_object(classid,objid,objsubid) AS obj
FROM pg_depend, pg_am
WHERE pg_depend.refclassid = 'pg_am'::regclass
    AND pg_am.oid = pg_depend.refobjid
    AND pg_am.amname = 'heap2'
ORDER BY classid, objid, objsubid;`,
				Results: []sql.Row{{`table tableam_tbl_heap2`}, {`table tableam_tblas_heap2`}, {`materialized view tableam_tblmv_heap2`}, {`table tableam_parted_b_heap2`}, {`table tableam_parted_d_heap2`}},
			},
			{
				Statement: `CREATE TABLE heaptable USING heap AS
  SELECT a, repeat(a::text, 100) FROM generate_series(1,9) AS a;`,
			},
			{
				Statement: `SELECT amname FROM pg_class c, pg_am am
  WHERE c.relam = am.oid AND c.oid = 'heaptable'::regclass;`,
				Results: []sql.Row{{`heap`}},
			},
			{
				Statement: `ALTER TABLE heaptable SET ACCESS METHOD heap2;`,
			},
			{
				Statement: `SELECT pg_describe_object(classid, objid, objsubid) as obj,
       pg_describe_object(refclassid, refobjid, refobjsubid) as objref,
       deptype
  FROM pg_depend
  WHERE classid = 'pg_class'::regclass AND
        objid = 'heaptable'::regclass
  ORDER BY 1, 2;`,
				Results: []sql.Row{{`table heaptable`, `access method heap2`, `n`}, {`table heaptable`, `schema public`, `n`}},
			},
			{
				Statement: `ALTER TABLE heaptable SET ACCESS METHOD heap;`,
			},
			{
				Statement: `SELECT pg_describe_object(classid, objid, objsubid) as obj,
       pg_describe_object(refclassid, refobjid, refobjsubid) as objref,
       deptype
  FROM pg_depend
  WHERE classid = 'pg_class'::regclass AND
        objid = 'heaptable'::regclass
  ORDER BY 1, 2;`,
				Results: []sql.Row{{`table heaptable`, `schema public`, `n`}},
			},
			{
				Statement: `ALTER TABLE heaptable SET ACCESS METHOD heap2;`,
			},
			{
				Statement: `SELECT amname FROM pg_class c, pg_am am
  WHERE c.relam = am.oid AND c.oid = 'heaptable'::regclass;`,
				Results: []sql.Row{{`heap2`}},
			},
			{
				Statement: `SELECT COUNT(a), COUNT(1) FILTER(WHERE a=1) FROM heaptable;`,
				Results:   []sql.Row{{9, 1}},
			},
			{
				Statement: `CREATE MATERIALIZED VIEW heapmv USING heap AS SELECT * FROM heaptable;`,
			},
			{
				Statement: `SELECT amname FROM pg_class c, pg_am am
  WHERE c.relam = am.oid AND c.oid = 'heapmv'::regclass;`,
				Results: []sql.Row{{`heap`}},
			},
			{
				Statement: `ALTER MATERIALIZED VIEW heapmv SET ACCESS METHOD heap2;`,
			},
			{
				Statement: `SELECT amname FROM pg_class c, pg_am am
  WHERE c.relam = am.oid AND c.oid = 'heapmv'::regclass;`,
				Results: []sql.Row{{`heap2`}},
			},
			{
				Statement: `SELECT COUNT(a), COUNT(1) FILTER(WHERE a=1) FROM heapmv;`,
				Results:   []sql.Row{{9, 1}},
			},
			{
				Statement:   `ALTER TABLE heaptable SET ACCESS METHOD heap, SET ACCESS METHOD heap2;`,
				ErrorString: `cannot have multiple SET ACCESS METHOD subcommands`,
			},
			{
				Statement:   `ALTER MATERIALIZED VIEW heapmv SET ACCESS METHOD heap, SET ACCESS METHOD heap2;`,
				ErrorString: `cannot have multiple SET ACCESS METHOD subcommands`,
			},
			{
				Statement: `DROP MATERIALIZED VIEW heapmv;`,
			},
			{
				Statement: `DROP TABLE heaptable;`,
			},
			{
				Statement: `CREATE TABLE am_partitioned(x INT, y INT)
  PARTITION BY hash (x);`,
			},
			{
				Statement:   `ALTER TABLE am_partitioned SET ACCESS METHOD heap2;`,
				ErrorString: `cannot change access method of a partitioned table`,
			},
			{
				Statement: `DROP TABLE am_partitioned;`,
			},
			{
				Statement: `BEGIN;`,
			},
			{
				Statement: `SET LOCAL default_table_access_method = 'heap2';`,
			},
			{
				Statement: `CREATE TABLE tableam_tbl_heapx(f1 int);`,
			},
			{
				Statement: `CREATE TABLE tableam_tblas_heapx AS SELECT * FROM tableam_tbl_heapx;`,
			},
			{
				Statement: `SELECT INTO tableam_tblselectinto_heapx FROM tableam_tbl_heapx;`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW tableam_tblmv_heapx USING heap2 AS SELECT * FROM tableam_tbl_heapx;`,
			},
			{
				Statement: `CREATE TABLE tableam_parted_heapx (a text, b int) PARTITION BY list (a);`,
			},
			{
				Statement: `CREATE TABLE tableam_parted_1_heapx PARTITION OF tableam_parted_heapx FOR VALUES IN ('a', 'b');`,
			},
			{
				Statement: `CREATE TABLE tableam_parted_2_heapx PARTITION OF tableam_parted_heapx FOR VALUES IN ('c', 'd') USING heap;`,
			},
			{
				Statement: `CREATE VIEW tableam_view_heapx AS SELECT * FROM tableam_tbl_heapx;`,
			},
			{
				Statement: `CREATE SEQUENCE tableam_seq_heapx;`,
			},
			{
				Statement: `CREATE FOREIGN DATA WRAPPER fdw_heap2 VALIDATOR postgresql_fdw_validator;`,
			},
			{
				Statement: `CREATE SERVER fs_heap2 FOREIGN DATA WRAPPER fdw_heap2 ;`,
			},
			{
				Statement: `CREATE FOREIGN table tableam_fdw_heapx () SERVER fs_heap2;`,
			},
			{
				Statement: `SELECT
    pc.relkind,
    pa.amname,
    CASE WHEN relkind = 't' THEN
        (SELECT 'toast for ' || relname::regclass FROM pg_class pcm WHERE pcm.reltoastrelid = pc.oid)
    ELSE
        relname::regclass::text
    END COLLATE "C" AS relname
FROM pg_class AS pc
    LEFT JOIN pg_am AS pa ON (pa.oid = pc.relam)
WHERE pc.relname LIKE 'tableam_%_heapx'
ORDER BY 3, 1, 2;`,
				Results: []sql.Row{{false, ``, `tableam_fdw_heapx`}, {`r`, `heap2`, `tableam_parted_1_heapx`}, {`r`, `heap`, `tableam_parted_2_heapx`}, {`p`, ``, `tableam_parted_heapx`}, {`S`, ``, `tableam_seq_heapx`}, {`r`, `heap2`, `tableam_tbl_heapx`}, {`r`, `heap2`, `tableam_tblas_heapx`}, {`m`, `heap2`, `tableam_tblmv_heapx`}, {`r`, `heap2`, `tableam_tblselectinto_heapx`}, {`v`, ``, `tableam_view_heapx`}},
			},
			{
				Statement: `ROLLBACK;`,
			},
			{
				Statement:   `CREATE TABLE i_am_a_failure() USING "";`,
				ErrorString: `zero-length delimited identifier at or near """"`,
			},
			{
				Statement:   `CREATE TABLE i_am_a_failure() USING i_do_not_exist_am;`,
				ErrorString: `access method "i_do_not_exist_am" does not exist`,
			},
			{
				Statement:   `CREATE TABLE i_am_a_failure() USING "I do not exist AM";`,
				ErrorString: `access method "I do not exist AM" does not exist`,
			},
			{
				Statement:   `CREATE TABLE i_am_a_failure() USING "btree";`,
				ErrorString: `access method "btree" is not of type TABLE`,
			},
			{
				Statement:   `DROP ACCESS METHOD heap2;`,
				ErrorString: `cannot drop access method heap2 because other objects depend on it`,
			},
			{
				Statement: `table tableam_tblas_heap2 depends on access method heap2
materialized view tableam_tblmv_heap2 depends on access method heap2
table tableam_parted_b_heap2 depends on access method heap2
table tableam_parted_d_heap2 depends on access method heap2
HINT:  Use DROP ... CASCADE to drop the dependent objects too.`,
			},
		},
	})
}
