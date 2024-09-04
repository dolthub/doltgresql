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

func TestCreateSchema(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_create_schema)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_create_schema,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE ROLE regress_create_schema_role SUPERUSER;`,
			},
			{
				Statement: `CREATE SCHEMA AUTHORIZATION regress_create_schema_role
  CREATE SEQUENCE schema_not_existing.seq;`,
				ErrorString: `CREATE specifies a schema (schema_not_existing) different from the one being created (regress_create_schema_role)`,
			},
			{
				Statement: `CREATE SCHEMA AUTHORIZATION regress_create_schema_role
  CREATE TABLE schema_not_existing.tab (id int);`,
				ErrorString: `CREATE specifies a schema (schema_not_existing) different from the one being created (regress_create_schema_role)`,
			},
			{
				Statement: `CREATE SCHEMA AUTHORIZATION regress_create_schema_role
  CREATE VIEW schema_not_existing.view AS SELECT 1;`,
				ErrorString: `CREATE specifies a schema (schema_not_existing) different from the one being created (regress_create_schema_role)`,
			},
			{
				Statement: `CREATE SCHEMA AUTHORIZATION regress_create_schema_role
  CREATE INDEX ON schema_not_existing.tab (id);`,
				ErrorString: `CREATE specifies a schema (schema_not_existing) different from the one being created (regress_create_schema_role)`,
			},
			{
				Statement: `CREATE SCHEMA AUTHORIZATION regress_create_schema_role
  CREATE TRIGGER schema_trig BEFORE INSERT ON schema_not_existing.tab
  EXECUTE FUNCTION schema_trig.no_func();`,
				ErrorString: `CREATE specifies a schema (schema_not_existing) different from the one being created (regress_create_schema_role)`,
			},
			{
				Statement: `SET ROLE regress_create_schema_role;`,
			},
			{
				Statement: `CREATE SCHEMA AUTHORIZATION CURRENT_ROLE
  CREATE SEQUENCE schema_not_existing.seq;`,
				ErrorString: `CREATE specifies a schema (schema_not_existing) different from the one being created (regress_create_schema_role)`,
			},
			{
				Statement: `CREATE SCHEMA AUTHORIZATION CURRENT_ROLE
  CREATE TABLE schema_not_existing.tab (id int);`,
				ErrorString: `CREATE specifies a schema (schema_not_existing) different from the one being created (regress_create_schema_role)`,
			},
			{
				Statement: `CREATE SCHEMA AUTHORIZATION CURRENT_ROLE
  CREATE VIEW schema_not_existing.view AS SELECT 1;`,
				ErrorString: `CREATE specifies a schema (schema_not_existing) different from the one being created (regress_create_schema_role)`,
			},
			{
				Statement: `CREATE SCHEMA AUTHORIZATION CURRENT_ROLE
  CREATE INDEX ON schema_not_existing.tab (id);`,
				ErrorString: `CREATE specifies a schema (schema_not_existing) different from the one being created (regress_create_schema_role)`,
			},
			{
				Statement: `CREATE SCHEMA AUTHORIZATION CURRENT_ROLE
  CREATE TRIGGER schema_trig BEFORE INSERT ON schema_not_existing.tab
  EXECUTE FUNCTION schema_trig.no_func();`,
				ErrorString: `CREATE specifies a schema (schema_not_existing) different from the one being created (regress_create_schema_role)`,
			},
			{
				Statement: `CREATE SCHEMA regress_schema_1 AUTHORIZATION CURRENT_ROLE
  CREATE SEQUENCE schema_not_existing.seq;`,
				ErrorString: `CREATE specifies a schema (schema_not_existing) different from the one being created (regress_schema_1)`,
			},
			{
				Statement: `CREATE SCHEMA regress_schema_1 AUTHORIZATION CURRENT_ROLE
  CREATE TABLE schema_not_existing.tab (id int);`,
				ErrorString: `CREATE specifies a schema (schema_not_existing) different from the one being created (regress_schema_1)`,
			},
			{
				Statement: `CREATE SCHEMA regress_schema_1 AUTHORIZATION CURRENT_ROLE
  CREATE VIEW schema_not_existing.view AS SELECT 1;`,
				ErrorString: `CREATE specifies a schema (schema_not_existing) different from the one being created (regress_schema_1)`,
			},
			{
				Statement: `CREATE SCHEMA regress_schema_1 AUTHORIZATION CURRENT_ROLE
  CREATE INDEX ON schema_not_existing.tab (id);`,
				ErrorString: `CREATE specifies a schema (schema_not_existing) different from the one being created (regress_schema_1)`,
			},
			{
				Statement: `CREATE SCHEMA regress_schema_1 AUTHORIZATION CURRENT_ROLE
  CREATE TRIGGER schema_trig BEFORE INSERT ON schema_not_existing.tab
  EXECUTE FUNCTION schema_trig.no_func();`,
				ErrorString: `CREATE specifies a schema (schema_not_existing) different from the one being created (regress_schema_1)`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `CREATE SCHEMA AUTHORIZATION regress_create_schema_role
  CREATE TABLE regress_create_schema_role.tab (id int);`,
			},
			{
				Statement: `\d regress_create_schema_role.tab
      Table "regress_create_schema_role.tab"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 id     | integer |           |          | 
DROP SCHEMA regress_create_schema_role CASCADE;`,
			},
			{
				Statement: `SET ROLE regress_create_schema_role;`,
			},
			{
				Statement: `CREATE SCHEMA AUTHORIZATION CURRENT_ROLE
  CREATE TABLE regress_create_schema_role.tab (id int);`,
			},
			{
				Statement: `\d regress_create_schema_role.tab
      Table "regress_create_schema_role.tab"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 id     | integer |           |          | 
DROP SCHEMA regress_create_schema_role CASCADE;`,
			},
			{
				Statement: `CREATE SCHEMA regress_schema_1 AUTHORIZATION CURRENT_ROLE
  CREATE TABLE regress_schema_1.tab (id int);`,
			},
			{
				Statement: `\d regress_schema_1.tab
           Table "regress_schema_1.tab"
 Column |  Type   | Collation | Nullable | Default 
--------+---------+-----------+----------+---------
 id     | integer |           |          | 
DROP SCHEMA regress_schema_1 CASCADE;`,
			},
			{
				Statement: `RESET ROLE;`,
			},
			{
				Statement: `DROP ROLE regress_create_schema_role;`,
			},
		},
	})
}
