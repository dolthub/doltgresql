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

func TestMiscSanity(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_misc_sanity)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_misc_sanity,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SELECT *
FROM pg_depend as d1
WHERE refclassid = 0 OR refobjid = 0 OR
      classid = 0 OR objid = 0 OR
      deptype NOT IN ('a', 'e', 'i', 'n', 'x', 'P', 'S');`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT *
FROM pg_shdepend as d1
WHERE refclassid = 0 OR refobjid = 0 OR
      classid = 0 OR objid = 0 OR
      deptype NOT IN ('a', 'o', 'r', 't');`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT relname, attname, atttypid::regtype
FROM pg_class c JOIN pg_attribute a ON c.oid = attrelid
WHERE c.oid < 16384 AND
      reltoastrelid = 0 AND
      relkind = 'r' AND
      attstorage != 'p'
ORDER BY 1, 2;`,
				Results: []sql.Row{{`pg_attribute`, `attacl`, `aclitem[]`}, {`pg_attribute`, `attfdwoptions`, `text[]`}, {`pg_attribute`, `attmissingval`, `anyarray`}, {`pg_attribute`, `attoptions`, `text[]`}, {`pg_class`, `relacl`, `aclitem[]`}, {`pg_class`, `reloptions`, `text[]`}, {`pg_class`, `relpartbound`, `pg_node_tree`}, {`pg_index`, `indexprs`, `pg_node_tree`}, {`pg_index`, `indpred`, `pg_node_tree`}, {`pg_largeobject`, `data`, `bytea`}, {`pg_largeobject_metadata`, `lomacl`, `aclitem[]`}},
			},
			{
				Statement: `SELECT relname
FROM pg_class
WHERE relnamespace = 'pg_catalog'::regnamespace AND relkind = 'r'
      AND pg_class.oid NOT IN (SELECT indrelid FROM pg_index WHERE indisprimary)
ORDER BY 1;`,
				Results: []sql.Row{{`pg_depend`}, {`pg_shdepend`}},
			},
			{
				Statement: `SELECT relname
FROM pg_class c JOIN pg_index i ON c.oid = i.indexrelid
WHERE relnamespace = 'pg_catalog'::regnamespace AND relkind = 'i'
      AND i.indisunique
      AND c.oid NOT IN (SELECT conindid FROM pg_constraint)
ORDER BY 1;`,
				Results: []sql.Row{},
			},
		},
	})
}
