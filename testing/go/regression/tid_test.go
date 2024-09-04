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

func TestTid(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_tid)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_tid,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT
  '(0,0)'::tid as tid00,
  '(0,1)'::tid as tid01,
  '(-1,0)'::tid as tidm10,
  '(4294967295,65535)'::tid as tidmax;`,
				Results: []sql.Row{{`(0,0)`, `(0,1)`, `(4294967295,0)`, `(4294967295,65535)`}},
			},
			{
				Statement:   `SELECT '(4294967296,1)'::tid;  -- error`,
				ErrorString: `invalid input syntax for type tid: "(4294967296,1)"`,
			},
			{
				Statement:   `SELECT '(1,65536)'::tid;  -- error`,
				ErrorString: `invalid input syntax for type tid: "(1,65536)"`,
			},
			{
				Statement: `CREATE TABLE tid_tab (a int);`,
			},
			{
				Statement: `INSERT INTO tid_tab VALUES (1), (2);`,
			},
			{
				Statement: `SELECT min(ctid) FROM tid_tab;`,
				Results:   []sql.Row{{`(0,1)`}},
			},
			{
				Statement: `SELECT max(ctid) FROM tid_tab;`,
				Results:   []sql.Row{{`(0,2)`}},
			},
			{
				Statement: `TRUNCATE tid_tab;`,
			},
			{
				Statement: `CREATE MATERIALIZED VIEW tid_matview AS SELECT a FROM tid_tab;`,
			},
			{
				Statement:   `SELECT currtid2('tid_matview'::text, '(0,1)'::tid); -- fails`,
				ErrorString: `tid (0, 1) is not valid for relation "tid_matview"`,
			},
			{
				Statement: `INSERT INTO tid_tab VALUES (1);`,
			},
			{
				Statement: `REFRESH MATERIALIZED VIEW tid_matview;`,
			},
			{
				Statement: `SELECT currtid2('tid_matview'::text, '(0,1)'::tid); -- ok`,
				Results:   []sql.Row{{`(0,1)`}},
			},
			{
				Statement: `DROP MATERIALIZED VIEW tid_matview;`,
			},
			{
				Statement: `TRUNCATE tid_tab;`,
			},
			{
				Statement: `CREATE SEQUENCE tid_seq;`,
			},
			{
				Statement: `SELECT currtid2('tid_seq'::text, '(0,1)'::tid); -- ok`,
				Results:   []sql.Row{{`(0,1)`}},
			},
			{
				Statement: `DROP SEQUENCE tid_seq;`,
			},
			{
				Statement: `CREATE INDEX tid_ind ON tid_tab(a);`,
			},
			{
				Statement:   `SELECT currtid2('tid_ind'::text, '(0,1)'::tid); -- fails`,
				ErrorString: `"tid_ind" is an index`,
			},
			{
				Statement: `DROP INDEX tid_ind;`,
			},
			{
				Statement: `CREATE TABLE tid_part (a int) PARTITION BY RANGE (a);`,
			},
			{
				Statement:   `SELECT currtid2('tid_part'::text, '(0,1)'::tid); -- fails`,
				ErrorString: `cannot look at latest visible tid for relation "public.tid_part"`,
			},
			{
				Statement: `DROP TABLE tid_part;`,
			},
			{
				Statement: `CREATE VIEW tid_view_no_ctid AS SELECT a FROM tid_tab;`,
			},
			{
				Statement:   `SELECT currtid2('tid_view_no_ctid'::text, '(0,1)'::tid); -- fails`,
				ErrorString: `currtid cannot handle views with no CTID`,
			},
			{
				Statement: `DROP VIEW tid_view_no_ctid;`,
			},
			{
				Statement: `CREATE VIEW tid_view_with_ctid AS SELECT ctid, a FROM tid_tab;`,
			},
			{
				Statement:   `SELECT currtid2('tid_view_with_ctid'::text, '(0,1)'::tid); -- fails`,
				ErrorString: `tid (0, 1) is not valid for relation "tid_tab"`,
			},
			{
				Statement: `INSERT INTO tid_tab VALUES (1);`,
			},
			{
				Statement: `SELECT currtid2('tid_view_with_ctid'::text, '(0,1)'::tid); -- ok`,
				Results:   []sql.Row{{`(0,1)`}},
			},
			{
				Statement: `DROP VIEW tid_view_with_ctid;`,
			},
			{
				Statement: `TRUNCATE tid_tab;`,
			},
			{
				Statement: `CREATE VIEW tid_view_fake_ctid AS SELECT 1 AS ctid, 2 AS a;`,
			},
			{
				Statement:   `SELECT currtid2('tid_view_fake_ctid'::text, '(0,1)'::tid); -- fails`,
				ErrorString: `ctid isn't of type TID`,
			},
			{
				Statement: `DROP VIEW tid_view_fake_ctid;`,
			},
			{
				Statement: `DROP TABLE tid_tab CASCADE;`,
			},
		},
	})
}
