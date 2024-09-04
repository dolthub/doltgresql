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

func TestSanityCheck(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_sanity_check)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_sanity_check,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `VACUUM;`,
			},
			{
				Statement: `SELECT relname, nspname
 FROM pg_class c LEFT JOIN pg_namespace n ON n.oid = relnamespace JOIN pg_attribute a ON (attrelid = c.oid AND attname = 'oid')
 WHERE relkind = 'r' and c.oid < 16384
     AND ((nspname ~ '^pg_') IS NOT FALSE)
     AND NOT EXISTS (SELECT 1 FROM pg_index i WHERE indrelid = c.oid
                     AND indkey[0] = a.attnum AND indnatts = 1
                     AND indisunique AND indimmediate);`,
				Results: []sql.Row{},
			},
			{
				Statement: `SELECT relname, relkind
  FROM pg_class
 WHERE relkind IN ('v', 'c', 'f', 'p', 'I')
       AND relfilenode <> 0;`,
				Results: []sql.Row{},
			},
			{
				Statement: `WITH check_columns AS (
 SELECT relname, attname,
  array(
   SELECT t.oid
    FROM pg_type t JOIN pg_attribute pa ON t.oid = pa.atttypid
    WHERE pa.attrelid = a.attrelid AND
          pa.attnum > 0 AND pa.attnum < a.attnum
    ORDER BY pa.attnum) AS coltypes
 FROM pg_attribute a JOIN pg_class c ON c.oid = attrelid
  JOIN pg_namespace n ON c.relnamespace = n.oid
 WHERE attalign = 'd' AND relkind = 'r' AND
  attnotnull AND attlen <> -1 AND n.nspname = 'pg_catalog'
)
SELECT relname, attname, coltypes, get_columns_length(coltypes)
 FROM check_columns
 WHERE get_columns_length(coltypes) % 8 != 0 OR
       'name'::regtype::oid = ANY(coltypes);`,
				Results: []sql.Row{},
			},
		},
	})
}
