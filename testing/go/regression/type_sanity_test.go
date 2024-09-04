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

func TestTypeSanity(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_type_sanity)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_type_sanity,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup, RegressionFileName_multirangetypes},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT t1.oid, t1.typname
FROM pg_type as t1
WHERE t1.typnamespace = 0 OR
    (t1.typlen <= 0 AND t1.typlen != -1 AND t1.typlen != -2) OR
    (t1.typtype not in ('b', 'c', 'd', 'e', 'm', 'p', 'r')) OR
    NOT t1.typisdefined OR
    (t1.typalign not in ('c', 's', 'i', 'd')) OR
    (t1.typstorage not in ('p', 'x', 'e', 'm'));`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname
FROM pg_type as t1
WHERE t1.typbyval AND
    (t1.typlen != 1 OR t1.typalign != 'c') AND
    (t1.typlen != 2 OR t1.typalign != 's') AND
    (t1.typlen != 4 OR t1.typalign != 'i') AND
    (t1.typlen != 8 OR t1.typalign != 'd');`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname
FROM pg_type as t1
WHERE t1.typstorage != 'p' AND
    (t1.typbyval OR t1.typlen != -1);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname
FROM pg_type as t1
WHERE (t1.typtype = 'c' AND t1.typrelid = 0) OR
    (t1.typtype != 'c' AND t1.typrelid != 0);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname
FROM pg_type as t1
WHERE t1.typtype not in ('p') AND t1.typname NOT LIKE E'\\_%'
    AND NOT EXISTS
    (SELECT 1 FROM pg_type as t2
     WHERE t2.typname = ('_' || t1.typname)::name AND
           t2.typelem = t1.oid and t1.typarray = t2.oid)
ORDER BY t1.oid;`,
				Results: []sql.Row{{194, `pg_node_tree`}, {3361, `pg_ndistinct`}, {3402, `pg_dependencies`}, {4600, `pg_brin_bloom_summary`}, {4601, `pg_brin_minmax_multi_summary`}, {5017, `pg_mcv_list`}},
			},
			{
				Statement: `SELECT t1.oid, t1.typname as basetype, t2.typname as arraytype,
       t2.typsubscript
FROM   pg_type t1 LEFT JOIN pg_type t2 ON (t1.typarray = t2.oid)
WHERE  t1.typarray <> 0 AND
       (t2.oid IS NULL OR
        t2.typsubscript <> 'array_subscript_handler'::regproc);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname
FROM pg_type as t1
WHERE t1.typtype = 'r' AND
   NOT EXISTS(SELECT 1 FROM pg_range r WHERE rngtypid = t1.oid);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, t1.typalign, t2.typname, t2.typalign
FROM pg_type as t1
     LEFT JOIN pg_range as r ON rngtypid = t1.oid
     LEFT JOIN pg_type as t2 ON rngsubtype = t2.oid
WHERE t1.typtype = 'r' AND
    (t1.typalign != (CASE WHEN t2.typalign = 'd' THEN 'd'::"char"
                          ELSE 'i'::"char" END)
     OR t2.oid IS NULL);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname
FROM pg_type as t1
WHERE (t1.typinput = 0 OR t1.typoutput = 0);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, p1.oid, p1.proname
FROM pg_type AS t1, pg_proc AS p1
WHERE t1.typinput = p1.oid AND NOT
    ((p1.pronargs = 1 AND p1.proargtypes[0] = 'cstring'::regtype) OR
     (p1.pronargs = 2 AND p1.proargtypes[0] = 'cstring'::regtype AND
      p1.proargtypes[1] = 'oid'::regtype) OR
     (p1.pronargs = 3 AND p1.proargtypes[0] = 'cstring'::regtype AND
      p1.proargtypes[1] = 'oid'::regtype AND
      p1.proargtypes[2] = 'int4'::regtype));`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT oid::regprocedure, provariadic::regtype, proargtypes::regtype[]
FROM pg_proc
WHERE provariadic != 0
AND case proargtypes[array_length(proargtypes, 1)-1]
	WHEN '"any"'::regtype THEN '"any"'::regtype
	WHEN 'anyarray'::regtype THEN 'anyelement'::regtype
	WHEN 'anycompatiblearray'::regtype THEN 'anycompatible'::regtype
	ELSE (SELECT t.oid
		  FROM pg_type t
		  WHERE t.typarray = proargtypes[array_length(proargtypes, 1)-1])
	END  != provariadic;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT oid::regprocedure, proargmodes, provariadic
FROM pg_proc
WHERE (proargmodes IS NOT NULL AND 'v' = any(proargmodes))
    IS DISTINCT FROM
    (provariadic != 0);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, p1.oid, p1.proname
FROM pg_type AS t1, pg_proc AS p1
WHERE t1.typinput = p1.oid AND t1.typtype in ('b', 'p') AND NOT
    (t1.typelem != 0 AND t1.typlen < 0) AND NOT
    (p1.prorettype = t1.oid AND NOT p1.proretset)
ORDER BY 1;`,
				Results: []sql.Row{{1790, `refcursor`, 46, `textin`}},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, p1.oid, p1.proname
FROM pg_type AS t1, pg_proc AS p1
WHERE t1.typinput = p1.oid AND
    (t1.typelem != 0 AND t1.typlen < 0) AND NOT
    (p1.oid = 'array_in'::regproc)
ORDER BY 1;`,
				Results: []sql.Row{{22, `int2vector`, 40, `int2vectorin`}, {30, `oidvector`, 54, `oidvectorin`}},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, p1.oid, p1.proname
FROM pg_type AS t1, pg_proc AS p1
WHERE t1.typinput = p1.oid AND p1.provolatile NOT IN ('i', 's');`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT DISTINCT typtype, typinput
FROM pg_type AS t1
WHERE t1.typtype not in ('b', 'p')
ORDER BY 1;`,
				Results: []sql.Row{{`c`, `record_in`}, {`d`, `domain_in`}, {`e`, `enum_in`}, {`m`, `multirange_in`}, {`r`, `range_in`}},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, p1.oid, p1.proname
FROM pg_type AS t1, pg_proc AS p1
WHERE t1.typoutput = p1.oid AND t1.typtype in ('b', 'p') AND NOT
    (p1.pronargs = 1 AND
     (p1.proargtypes[0] = t1.oid OR
      (p1.oid = 'array_out'::regproc AND
       t1.typelem != 0 AND t1.typlen = -1)))
ORDER BY 1;`,
				Results: []sql.Row{{1790, `refcursor`, 47, `textout`}},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, p1.oid, p1.proname
FROM pg_type AS t1, pg_proc AS p1
WHERE t1.typoutput = p1.oid AND NOT
    (p1.prorettype = 'cstring'::regtype AND NOT p1.proretset);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, p1.oid, p1.proname
FROM pg_type AS t1, pg_proc AS p1
WHERE t1.typoutput = p1.oid AND p1.provolatile NOT IN ('i', 's');`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT DISTINCT typtype, typoutput
FROM pg_type AS t1
WHERE t1.typtype not in ('b', 'd', 'p')
ORDER BY 1;`,
				Results: []sql.Row{{`c`, `record_out`}, {`e`, `enum_out`}, {`m`, `multirange_out`}, {`r`, `range_out`}},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, t2.oid, t2.typname
FROM pg_type AS t1 LEFT JOIN pg_type AS t2 ON t1.typbasetype = t2.oid
WHERE t1.typtype = 'd' AND t1.typoutput IS DISTINCT FROM t2.typoutput;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, p1.oid, p1.proname
FROM pg_type AS t1, pg_proc AS p1
WHERE t1.typreceive = p1.oid AND NOT
    ((p1.pronargs = 1 AND p1.proargtypes[0] = 'internal'::regtype) OR
     (p1.pronargs = 2 AND p1.proargtypes[0] = 'internal'::regtype AND
      p1.proargtypes[1] = 'oid'::regtype) OR
     (p1.pronargs = 3 AND p1.proargtypes[0] = 'internal'::regtype AND
      p1.proargtypes[1] = 'oid'::regtype AND
      p1.proargtypes[2] = 'int4'::regtype));`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, p1.oid, p1.proname
FROM pg_type AS t1, pg_proc AS p1
WHERE t1.typreceive = p1.oid AND t1.typtype in ('b', 'p') AND NOT
    (t1.typelem != 0 AND t1.typlen < 0) AND NOT
    (p1.prorettype = t1.oid AND NOT p1.proretset)
ORDER BY 1;`,
				Results: []sql.Row{{1790, `refcursor`, 2414, `textrecv`}},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, p1.oid, p1.proname
FROM pg_type AS t1, pg_proc AS p1
WHERE t1.typreceive = p1.oid AND
    (t1.typelem != 0 AND t1.typlen < 0) AND NOT
    (p1.oid = 'array_recv'::regproc)
ORDER BY 1;`,
				Results: []sql.Row{{22, `int2vector`, 2410, `int2vectorrecv`}, {30, `oidvector`, 2420, `oidvectorrecv`}},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, p1.oid, p1.proname, p2.oid, p2.proname
FROM pg_type AS t1, pg_proc AS p1, pg_proc AS p2
WHERE t1.typinput = p1.oid AND t1.typreceive = p2.oid AND
    p1.pronargs != p2.pronargs;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, p1.oid, p1.proname
FROM pg_type AS t1, pg_proc AS p1
WHERE t1.typreceive = p1.oid AND p1.provolatile NOT IN ('i', 's');`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT DISTINCT typtype, typreceive
FROM pg_type AS t1
WHERE t1.typtype not in ('b', 'p')
ORDER BY 1;`,
				Results: []sql.Row{{`c`, `record_recv`}, {`d`, `domain_recv`}, {`e`, `enum_recv`}, {`m`, `multirange_recv`}, {`r`, `range_recv`}},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, p1.oid, p1.proname
FROM pg_type AS t1, pg_proc AS p1
WHERE t1.typsend = p1.oid AND t1.typtype in ('b', 'p') AND NOT
    (p1.pronargs = 1 AND
     (p1.proargtypes[0] = t1.oid OR
      (p1.oid = 'array_send'::regproc AND
       t1.typelem != 0 AND t1.typlen = -1)))
ORDER BY 1;`,
				Results: []sql.Row{{1790, `refcursor`, 2415, `textsend`}},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, p1.oid, p1.proname
FROM pg_type AS t1, pg_proc AS p1
WHERE t1.typsend = p1.oid AND NOT
    (p1.prorettype = 'bytea'::regtype AND NOT p1.proretset);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, p1.oid, p1.proname
FROM pg_type AS t1, pg_proc AS p1
WHERE t1.typsend = p1.oid AND p1.provolatile NOT IN ('i', 's');`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT DISTINCT typtype, typsend
FROM pg_type AS t1
WHERE t1.typtype not in ('b', 'd', 'p')
ORDER BY 1;`,
				Results: []sql.Row{{`c`, `record_send`}, {`e`, `enum_send`}, {`m`, `multirange_send`}, {`r`, `range_send`}},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, t2.oid, t2.typname
FROM pg_type AS t1 LEFT JOIN pg_type AS t2 ON t1.typbasetype = t2.oid
WHERE t1.typtype = 'd' AND t1.typsend IS DISTINCT FROM t2.typsend;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, p1.oid, p1.proname
FROM pg_type AS t1, pg_proc AS p1
WHERE t1.typmodin = p1.oid AND NOT
    (p1.pronargs = 1 AND
     p1.proargtypes[0] = 'cstring[]'::regtype AND
     p1.prorettype = 'int4'::regtype AND NOT p1.proretset);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, p1.oid, p1.proname
FROM pg_type AS t1, pg_proc AS p1
WHERE t1.typmodin = p1.oid AND p1.provolatile NOT IN ('i', 's');`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, p1.oid, p1.proname
FROM pg_type AS t1, pg_proc AS p1
WHERE t1.typmodout = p1.oid AND NOT
    (p1.pronargs = 1 AND
     p1.proargtypes[0] = 'int4'::regtype AND
     p1.prorettype = 'cstring'::regtype AND NOT p1.proretset);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, p1.oid, p1.proname
FROM pg_type AS t1, pg_proc AS p1
WHERE t1.typmodout = p1.oid AND p1.provolatile NOT IN ('i', 's');`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, t2.oid, t2.typname
FROM pg_type AS t1, pg_type AS t2
WHERE t1.typelem = t2.oid AND NOT
    (t1.typmodin = t2.typmodin AND t1.typmodout = t2.typmodout);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, t2.oid, t2.typname
FROM pg_type AS t1, pg_type AS t2
WHERE t1.typarray = t2.oid AND NOT (t1.typdelim = t2.typdelim);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, t1.typalign, t2.typname, t2.typalign
FROM pg_type AS t1, pg_type AS t2
WHERE t1.typarray = t2.oid AND
    t2.typalign != (CASE WHEN t1.typalign = 'd' THEN 'd'::"char"
                         ELSE 'i'::"char" END);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, t1.typelem
FROM pg_type AS t1
WHERE t1.typelem != 0 AND t1.typsubscript = 0;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname,
       t1.typelem, t1.typlen, t1.typbyval
FROM pg_type AS t1
WHERE t1.typsubscript = 'array_subscript_handler'::regproc AND NOT
    (t1.typelem != 0 AND t1.typlen = -1 AND NOT t1.typbyval);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname,
       t1.typelem, t1.typlen, t1.typbyval
FROM pg_type AS t1
WHERE t1.typsubscript = 'raw_array_subscript_handler'::regproc AND NOT
    (t1.typelem != 0 AND t1.typlen > 0 AND NOT t1.typbyval);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t1.oid, t1.typname, p1.oid, p1.proname
FROM pg_type AS t1, pg_proc AS p1
WHERE t1.typanalyze = p1.oid AND NOT
    (p1.pronargs = 1 AND
     p1.proargtypes[0] = 'internal'::regtype AND
     p1.prorettype = 'bool'::regtype AND NOT p1.proretset);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT d.oid, d.typname, d.typanalyze, t.oid, t.typname, t.typanalyze
FROM pg_type d JOIN pg_type t ON d.typbasetype = t.oid
WHERE d.typanalyze != t.typanalyze;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t.oid, t.typname, t.typanalyze
FROM pg_type t LEFT JOIN pg_range r on t.oid = r.rngtypid
WHERE t.typbasetype = 0 AND
    (t.typanalyze = 'range_typanalyze'::regproc) != (r.rngtypid IS NOT NULL);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT t.oid, t.typname, t.typanalyze
FROM pg_type t
WHERE t.typbasetype = 0 AND
    (t.typanalyze = 'array_typanalyze'::regproc) !=
    (t.typsubscript = 'array_subscript_handler'::regproc)
ORDER BY 1;`,
				Results: []sql.Row{{22, `int2vector`, `-`}, {30, `oidvector`, `-`}},
			},
			{
				Statement: `SELECT c1.oid, c1.relname
FROM pg_class as c1
WHERE relkind NOT IN ('r', 'i', 'S', 't', 'v', 'm', 'c', 'f', 'p') OR
    relpersistence NOT IN ('p', 'u', 't') OR
    relreplident NOT IN ('d', 'n', 'f', 'i');`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT c1.oid, c1.relname
FROM pg_class as c1
WHERE c1.relkind NOT IN ('S', 'v', 'f', 'c') and
    c1.relam = 0;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT c1.oid, c1.relname
FROM pg_class as c1
WHERE c1.relkind IN ('S', 'v', 'f', 'c') and
    c1.relam != 0;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT pc.oid, pc.relname, pa.amname, pa.amtype
FROM pg_class as pc JOIN pg_am AS pa ON (pc.relam = pa.oid)
WHERE pc.relkind IN ('i') and
    pa.amtype != 'i';`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT pc.oid, pc.relname, pa.amname, pa.amtype
FROM pg_class as pc JOIN pg_am AS pa ON (pc.relam = pa.oid)
WHERE pc.relkind IN ('r', 't', 'm') and
    pa.amtype != 't';`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT a1.attrelid, a1.attname
FROM pg_attribute as a1
WHERE a1.attrelid = 0 OR a1.atttypid = 0 OR a1.attnum = 0 OR
    a1.attcacheoff != -1 OR a1.attinhcount < 0 OR
    (a1.attinhcount = 0 AND NOT a1.attislocal);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT a1.attrelid, a1.attname, c1.oid, c1.relname
FROM pg_attribute AS a1, pg_class AS c1
WHERE a1.attrelid = c1.oid AND a1.attnum > c1.relnatts;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT c1.oid, c1.relname
FROM pg_class AS c1
WHERE c1.relnatts != (SELECT count(*) FROM pg_attribute AS a1
                      WHERE a1.attrelid = c1.oid AND a1.attnum > 0);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT a1.attrelid, a1.attname, t1.oid, t1.typname
FROM pg_attribute AS a1, pg_type AS t1
WHERE a1.atttypid = t1.oid AND
    (a1.attlen != t1.typlen OR
     a1.attalign != t1.typalign OR
     a1.attbyval != t1.typbyval OR
     (a1.attstorage != t1.typstorage AND a1.attstorage != 'p'));`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT r.rngtypid, r.rngsubtype
FROM pg_range as r
WHERE r.rngtypid = 0 OR r.rngsubtype = 0 OR r.rngsubopc = 0;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT r.rngtypid, r.rngsubtype, r.rngcollation, t.typcollation
FROM pg_range r JOIN pg_type t ON t.oid = r.rngsubtype
WHERE (rngcollation = 0) != (typcollation = 0);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT r.rngtypid, r.rngsubtype, o.opcmethod, o.opcname
FROM pg_range r JOIN pg_opclass o ON o.oid = r.rngsubopc
WHERE o.opcmethod != 403 OR
    ((o.opcintype != r.rngsubtype) AND NOT
     (o.opcintype = 'pg_catalog.anyarray'::regtype AND
      EXISTS(select 1 from pg_catalog.pg_type where
             oid = r.rngsubtype and typelem != 0 and
             typsubscript = 'array_subscript_handler'::regproc)));`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT r.rngtypid, r.rngsubtype, p.proname
FROM pg_range r JOIN pg_proc p ON p.oid = r.rngcanonical
WHERE pronargs != 1 OR proargtypes[0] != rngtypid OR prorettype != rngtypid;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT r.rngtypid, r.rngsubtype, p.proname
FROM pg_range r JOIN pg_proc p ON p.oid = r.rngsubdiff
WHERE pronargs != 2
    OR proargtypes[0] != rngsubtype OR proargtypes[1] != rngsubtype
    OR prorettype != 'pg_catalog.float8'::regtype;`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT r.rngtypid, r.rngsubtype, r.rngmultitypid
FROM pg_range r
WHERE r.rngmultitypid IS NULL OR r.rngmultitypid = 0;`,
				Results: []sql.Row{},
			},
			{
				Statement: `CREATE TABLE tab_core_types AS SELECT
  '(11,12)'::point,
  '(1,1),(2,2)'::line,
  '((11,11),(12,12))'::lseg,
  '((11,11),(13,13))'::box,
  '((11,12),(13,13),(14,14))'::path AS openedpath,
  '[(11,12),(13,13),(14,14)]'::path AS closedpath,
  '((11,12),(13,13),(14,14))'::polygon,
  '1,1,1'::circle,
  'today'::date,
  'now'::time,
  'now'::timestamp,
  'now'::timetz,
  'now'::timestamptz,
  '12 seconds'::interval,
  '{"reason":"because"}'::json,
  '{"when":"now"}'::jsonb,
  '$.a[*] ? (@ > 2)'::jsonpath,
  '127.0.0.1'::inet,
  '127.0.0.0/8'::cidr,
  '00:01:03:86:1c:ba'::macaddr8,
  '00:01:03:86:1c:ba'::macaddr,
  2::int2, 4::int4, 8::int8,
  4::float4, '8'::float8, pi()::numeric,
  'foo'::"char",
  'c'::bpchar,
  'abc'::varchar,
  'name'::name,
  'txt'::text,
  true::bool,
  E'\\xDEADBEEF'::bytea,
  B'10001'::bit,
  B'10001'::varbit AS varbit,
  '12.34'::money,
  'abc'::refcursor,
  '1 2'::int2vector,
  '1 2'::oidvector,
  format('%I=UC/%I', USER, USER)::aclitem AS aclitem,
  'a fat cat sat on a mat and ate a fat rat'::tsvector,
  'fat & rat'::tsquery,
  'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'::uuid,
  '11'::xid8,
  'pg_class'::regclass,
  'regtype'::regtype type,
  'pg_monitor'::regrole,
  'pg_class'::regclass::oid,
  '(1,1)'::tid, '2'::xid, '3'::cid,
  '10:20:10,14,15'::txid_snapshot,
  '10:20:10,14,15'::pg_snapshot,
  '16/B374D848'::pg_lsn,
  1::information_schema.cardinal_number,
  'l'::information_schema.character_data,
  'n'::information_schema.sql_identifier,
  'now'::information_schema.time_stamp,
  'YES'::information_schema.yes_or_no,
  '(1,2)'::int4range, '{(1,2)}'::int4multirange,
  '(3,4)'::int8range, '{(3,4)}'::int8multirange,
  '(3,4)'::numrange, '{(3,4)}'::nummultirange,
  '(2020-01-02, 2021-02-03)'::daterange,
  '{(2020-01-02, 2021-02-03)}'::datemultirange,
  '(2020-01-02 03:04:05, 2021-02-03 06:07:08)'::tsrange,
  '{(2020-01-02 03:04:05, 2021-02-03 06:07:08)}'::tsmultirange,
  '(2020-01-02 03:04:05, 2021-02-03 06:07:08)'::tstzrange,
  '{(2020-01-02 03:04:05, 2021-02-03 06:07:08)}'::tstzmultirange;`,
			},
			{
				Statement: `SELECT oid, typname, typtype, typelem, typarray
  FROM pg_type t
  WHERE oid < 16384 AND
    -- Exclude pseudotypes and composite types.
    typtype NOT IN ('p', 'c') AND
    -- These reg* types cannot be pg_upgraded, so discard them.
    oid != ALL(ARRAY['regproc', 'regprocedure', 'regoper',
                     'regoperator', 'regconfig', 'regdictionary',
                     'regnamespace', 'regcollation']::regtype[]) AND
    -- Discard types that do not accept input values as these cannot be
    -- tested easily.
    -- Note: XML might be disabled at compile-time.
    oid != ALL(ARRAY['gtsvector', 'pg_node_tree',
                     'pg_ndistinct', 'pg_dependencies', 'pg_mcv_list',
                     'pg_brin_bloom_summary',
                     'pg_brin_minmax_multi_summary', 'xml']::regtype[]) AND
    -- Discard arrays.
    NOT EXISTS (SELECT 1 FROM pg_type u WHERE u.typarray = t.oid)
    -- Exclude everything from the table created above.  This checks
    -- that no in-core types are missing in tab_core_types.
    AND NOT EXISTS (SELECT 1
                    FROM pg_attribute a
                    WHERE a.atttypid=t.oid AND
                          a.attnum > 0 AND
                          a.attrelid='tab_core_types'::regclass);`,
				Results: []sql.Row{},
			},
		},
	})
}
