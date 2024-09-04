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

func TestSecurityLabel(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_security_label)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_security_label,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `SET client_min_messages TO 'warning';`,
			},
			{
				Statement: `DROP ROLE IF EXISTS regress_seclabel_user1;`,
			},
			{
				Statement: `DROP ROLE IF EXISTS regress_seclabel_user2;`,
			},
			{
				Statement: `RESET client_min_messages;`,
			},
			{
				Statement: `CREATE USER regress_seclabel_user1 WITH CREATEROLE;`,
			},
			{
				Statement: `CREATE USER regress_seclabel_user2;`,
			},
			{
				Statement: `CREATE TABLE seclabel_tbl1 (a int, b text);`,
			},
			{
				Statement: `CREATE TABLE seclabel_tbl2 (x int, y text);`,
			},
			{
				Statement: `CREATE VIEW seclabel_view1 AS SELECT * FROM seclabel_tbl2;`,
			},
			{
				Statement: `CREATE FUNCTION seclabel_four() RETURNS integer AS $$SELECT 4$$ language sql;`,
			},
			{
				Statement: `CREATE DOMAIN seclabel_domain AS text;`,
			},
			{
				Statement: `ALTER TABLE seclabel_tbl1 OWNER TO regress_seclabel_user1;`,
			},
			{
				Statement: `ALTER TABLE seclabel_tbl2 OWNER TO regress_seclabel_user2;`,
			},
			{
				Statement:   `SECURITY LABEL ON TABLE seclabel_tbl1 IS 'classified';			-- fail`,
				ErrorString: `no security label providers have been loaded`,
			},
			{
				Statement:   `SECURITY LABEL FOR 'dummy' ON TABLE seclabel_tbl1 IS 'classified';		-- fail`,
				ErrorString: `security label provider "dummy" is not loaded`,
			},
			{
				Statement:   `SECURITY LABEL ON TABLE seclabel_tbl1 IS '...invalid label...';		-- fail`,
				ErrorString: `no security label providers have been loaded`,
			},
			{
				Statement:   `SECURITY LABEL ON TABLE seclabel_tbl3 IS 'unclassified';			-- fail`,
				ErrorString: `no security label providers have been loaded`,
			},
			{
				Statement:   `SECURITY LABEL ON ROLE regress_seclabel_user1 IS 'classified';			-- fail`,
				ErrorString: `no security label providers have been loaded`,
			},
			{
				Statement:   `SECURITY LABEL FOR 'dummy' ON ROLE regress_seclabel_user1 IS 'classified';		-- fail`,
				ErrorString: `security label provider "dummy" is not loaded`,
			},
			{
				Statement:   `SECURITY LABEL ON ROLE regress_seclabel_user1 IS '...invalid label...';		-- fail`,
				ErrorString: `no security label providers have been loaded`,
			},
			{
				Statement:   `SECURITY LABEL ON ROLE regress_seclabel_user3 IS 'unclassified';			-- fail`,
				ErrorString: `no security label providers have been loaded`,
			},
			{
				Statement: `DROP FUNCTION seclabel_four();`,
			},
			{
				Statement: `DROP DOMAIN seclabel_domain;`,
			},
			{
				Statement: `DROP VIEW seclabel_view1;`,
			},
			{
				Statement: `DROP TABLE seclabel_tbl1;`,
			},
			{
				Statement: `DROP TABLE seclabel_tbl2;`,
			},
			{
				Statement: `DROP USER regress_seclabel_user1;`,
			},
			{
				Statement: `DROP USER regress_seclabel_user2;`,
			},
		},
	})
}
