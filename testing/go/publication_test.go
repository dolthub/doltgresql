// Copyright 2026 Dolthub, Inc.
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

package _go

import "testing"

func TestPublicationDDL(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{Name: "creation options and names", Assertions: []ScriptTestAssertion{
			{Query: `CREATE PUBLICATION p`, ExpectedTag: "CREATE PUBLICATION", ExpectedNotices: []ExpectedNotice{{Severity: "WARNING", Message: "publications are stored as metadata; logical replication publishing is not supported"}}},
			{Query: `CREATE PUBLICATION p WITH (unknown = true)`, ExpectedErr: `publication "p" already exists`, ExpectedErrCode: "42710"},
			{Query: `CREATE TABLE p (i int)`},
			{Query: `CREATE PUBLICATION "Quoted Name" FOR ALL TABLES WITH (publish='INSERT, insert', publish_via_partition_root)`, ExpectedTag: "CREATE PUBLICATION"},
			{Query: `CREATE PUBLICATION flags WITH (publish='', publish_via_partition_root=off)`},
			{Query: `DROP PUBLICATION p,p,"Quoted Name",flags CASCADE`, ExpectedTag: "DROP PUBLICATION"},
			{Query: `DROP PUBLICATION p`, ExpectedErr: `publication "p" does not exist`, ExpectedErrCode: "42704"},
			{Query: `DROP PUBLICATION IF EXISTS p RESTRICT`, ExpectedTag: "DROP PUBLICATION", ExpectedNotices: []ExpectedNotice{{Severity: "NOTICE", Message: `publication "p" does not exist, skipping`}}},
			{Query: `DROP PUBLICATION IF EXISTS p,p`, ExpectedTag: "DROP PUBLICATION", ExpectedNotices: []ExpectedNotice{
				{Severity: "NOTICE", Message: `publication "p" does not exist, skipping`},
				{Severity: "NOTICE", Message: `publication "p" does not exist, skipping`},
			}},
		}},
		{Name: "invalid options do not create publications", Assertions: []ScriptTestAssertion{
			{Query: `CREATE PUBLICATION p WITH (publish)`, ExpectedErr: "publish requires a parameter", ExpectedErrCode: "42601"},
			{Query: `CREATE PUBLICATION p WITH (unknown=true)`, ExpectedErr: `unrecognized publication parameter: "unknown"`, ExpectedErrCode: "42601"},
			{Query: `CREATE PUBLICATION p WITH (publish='insert',publish='delete')`, ExpectedErr: "conflicting or redundant options", ExpectedErrCode: "42601"},
			{Query: `CREATE PUBLICATION p WITH (publish='insert,')`, ExpectedErr: `invalid list syntax in parameter "publish"`, ExpectedErrCode: "42601"},
			{Query: `CREATE PUBLICATION p WITH (publish='merge')`, ExpectedErr: `unrecognized value for publication option "publish": "merge"`, ExpectedErrCode: "42601"},
			{Query: `CREATE PUBLICATION p WITH (publish='"INSERT"')`, ExpectedErr: `unrecognized value for publication option "publish": "INSERT"`, ExpectedErrCode: "42601"},
			{Query: `CREATE PUBLICATION p WITH (publish_via_partition_root='yes')`, ExpectedErr: "requires a Boolean value", ExpectedErrCode: "42601"},
			{Query: `CREATE PUBLICATION p WITH (publish_via_partition_root='1')`, ExpectedErr: "requires a Boolean value", ExpectedErrCode: "42601"},
			{Query: `CREATE PUBLICATION p WITH (publish_via_partition_root=1.0)`, ExpectedErr: "requires a Boolean value", ExpectedErrCode: "42601"},
			{Query: `CREATE PUBLICATION p WITH (publish_via_partition_root=-1)`, ExpectedErr: "requires a Boolean value", ExpectedErrCode: "42601"},
			{Query: `DROP PUBLICATION p`, ExpectedErr: `publication "p" does not exist`},
			{Query: `CREATE PUBLICATION p WITH (publish=insert,publish_via_partition_root=1)`},
			{Query: `DROP PUBLICATION p, missing`, ExpectedErr: `publication "missing" does not exist`},
			{Query: `DROP PUBLICATION p`},
		}},
		{Name: "unsupported selectors", Assertions: []ScriptTestAssertion{
			{Query: `CREATE PUBLICATION p FOR TABLE t`, ExpectedErr: "CREATE PUBLICATION FOR TABLE", ExpectedErrCode: "0A000"},
			{Query: `CREATE PUBLICATION p FOR TABLE t(i) WHERE (i>0)`, ExpectedErr: "CREATE PUBLICATION FOR TABLE", ExpectedErrCode: "0A000"},
			{Query: `CREATE PUBLICATION p FOR TABLES IN SCHEMA public`, ExpectedErr: "CREATE PUBLICATION FOR TABLES IN SCHEMA", ExpectedErrCode: "0A000"},
			{Query: `DROP PUBLICATION p`, ExpectedErr: `publication "p" does not exist`},
		}},
		{Name: "transaction rollback", Assertions: []ScriptTestAssertion{
			{Query: `BEGIN`}, {Query: `CREATE PUBLICATION p`}, {Query: `ROLLBACK`},
			{Query: `DROP PUBLICATION p`, ExpectedErr: `publication "p" does not exist`},
			{Query: `CREATE PUBLICATION p`}, {Query: `BEGIN`}, {Query: `DROP PUBLICATION p`}, {Query: `ROLLBACK`}, {Query: `DROP PUBLICATION p`},
		}},
		{Name: "read only transactions", Assertions: []ScriptTestAssertion{
			{Query: `CREATE PUBLICATION p`},
			{Query: `BEGIN READ ONLY`},
			{Query: `CREATE PUBLICATION q`, ExpectedErr: "read-only", ExpectedErrCode: "25006"},
			{Query: `ROLLBACK`},
			{Query: `BEGIN READ ONLY`},
			{Query: `DROP PUBLICATION p`, ExpectedErr: "read-only", ExpectedErrCode: "25006"},
			{Query: `ROLLBACK`},
			{Query: `DROP PUBLICATION p`},
		}},
		{Name: "publication privileges", SetUpScript: []string{
			`CREATE ROLE publisher LOGIN PASSWORD 'test'`, `CREATE ROLE other_user LOGIN PASSWORD 'test'`, `CREATE ROLE member_user LOGIN PASSWORD 'test'`,
			`GRANT publisher TO member_user`,
		}, Assertions: []ScriptTestAssertion{
			{Query: `CREATE PUBLICATION denied`, Username: "publisher", Password: "test", ExpectedErr: "permission denied for database postgres", ExpectedErrCode: "42501"},
			{Query: `GRANT CREATE ON DATABASE postgres TO publisher`},
			{Query: `CREATE PUBLICATION denied FOR ALL TABLES`, Username: "publisher", Password: "test", ExpectedErr: "must be superuser", ExpectedErrCode: "42501"},
			{Query: `CREATE PUBLICATION owned`, Username: "publisher", Password: "test"},
			{Query: `DROP PUBLICATION owned`, Username: "other_user", Password: "test", ExpectedErr: "must be owner of publication owned", ExpectedErrCode: "42501"},
			{Query: `DROP PUBLICATION owned`, Username: "member_user", Password: "test"},
			{Query: `CREATE PUBLICATION owned`, Username: "publisher", Password: "test"},
			{Query: `DROP PUBLICATION owned`},
		}},
	})
}
