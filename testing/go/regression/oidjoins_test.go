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
)

func TestOidjoins(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_oidjoins)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_oidjoins,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `DO $doblock$
declare
  fk record;`,
			},
			{
				Statement: `  nkeys integer;`,
			},
			{
				Statement: `  cmd text;`,
			},
			{
				Statement: `  err record;`,
			},
			{
				Statement: `begin
  for fk in select * from pg_get_catalog_foreign_keys()
  loop
    raise notice 'checking % % => % %',
      fk.fktable, fk.fkcols, fk.pktable, fk.pkcols;`,
			},
			{
				Statement: `    nkeys := array_length(fk.fkcols, 1);`,
			},
			{
				Statement: `    cmd := 'SELECT ctid';`,
			},
			{
				Statement: `    for i in 1 .. nkeys loop
      cmd := cmd || ', ' || quote_ident(fk.fkcols[i]);`,
			},
			{
				Statement: `    end loop;`,
			},
			{
				Statement: `    if fk.is_array then
      cmd := cmd || ' FROM (SELECT ctid';`,
			},
			{
				Statement: `      for i in 1 .. nkeys-1 loop
        cmd := cmd || ', ' || quote_ident(fk.fkcols[i]);`,
			},
			{
				Statement: `      end loop;`,
			},
			{
				Statement: `      cmd := cmd || ', unnest(' || quote_ident(fk.fkcols[nkeys]);`,
			},
			{
				Statement: `      cmd := cmd || ') as ' || quote_ident(fk.fkcols[nkeys]);`,
			},
			{
				Statement: `      cmd := cmd || ' FROM ' || fk.fktable::text || ') fk WHERE ';`,
			},
			{
				Statement: `    else
      cmd := cmd || ' FROM ' || fk.fktable::text || ' fk WHERE ';`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    if fk.is_opt then
      for i in 1 .. nkeys loop
        cmd := cmd || quote_ident(fk.fkcols[i]) || ' != 0 AND ';`,
			},
			{
				Statement: `      end loop;`,
			},
			{
				Statement: `    end if;`,
			},
			{
				Statement: `    cmd := cmd || 'NOT EXISTS(SELECT 1 FROM ' || fk.pktable::text || ' pk WHERE ';`,
			},
			{
				Statement: `    for i in 1 .. nkeys loop
      if i > 1 then cmd := cmd || ' AND '; end if;`,
			},
			{
				Statement: `      cmd := cmd || 'pk.' || quote_ident(fk.pkcols[i]);`,
			},
			{
				Statement: `      cmd := cmd || ' = fk.' || quote_ident(fk.fkcols[i]);`,
			},
			{
				Statement: `    end loop;`,
			},
			{
				Statement: `    cmd := cmd || ')';`,
			},
			{
				Statement: `    -- raise notice 'cmd = %', cmd;`,
			},
			{
				Statement: `    for err in execute cmd loop
      raise warning 'FK VIOLATION IN %(%): %', fk.fktable, fk.fkcols, err;`,
			},
			{
				Statement: `    end loop;`,
			},
			{
				Statement: `  end loop;`,
			},
			{
				Statement: `end
$doblock$;`,
			},
		},
	})
}
