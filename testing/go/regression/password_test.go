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

func TestPassword(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_password)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_password,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement:   `SET password_encryption = 'novalue'; -- error`,
				ErrorString: `invalid value for parameter "password_encryption": "novalue"`,
			},
			{
				Statement:   `SET password_encryption = true; -- error`,
				ErrorString: `invalid value for parameter "password_encryption": "true"`,
			},
			{
				Statement: `SET password_encryption = 'md5'; -- ok`,
			},
			{
				Statement: `SET password_encryption = 'scram-sha-256'; -- ok`,
			},
			{
				Statement: `SET password_encryption = 'md5';`,
			},
			{
				Statement: `CREATE ROLE regress_passwd1 PASSWORD 'role_pwd1';`,
			},
			{
				Statement: `CREATE ROLE regress_passwd2 PASSWORD 'role_pwd2';`,
			},
			{
				Statement: `SET password_encryption = 'scram-sha-256';`,
			},
			{
				Statement: `CREATE ROLE regress_passwd3 PASSWORD 'role_pwd3';`,
			},
			{
				Statement: `CREATE ROLE regress_passwd4 PASSWORD NULL;`,
			},
			{
				Statement: `SELECT rolname, regexp_replace(rolpassword, '(SCRAM-SHA-256)\$(\d+):([a-zA-Z0-9+/=]+)\$([a-zA-Z0-9+=/]+):([a-zA-Z0-9+/=]+)', '\1$\2:<salt>$<storedkey>:<serverkey>') as rolpassword_masked
    FROM pg_authid
    WHERE rolname LIKE 'regress_passwd%'
    ORDER BY rolname, rolpassword;`,
				Results: []sql.Row{{`regress_passwd1`, `md5783277baca28003b33453252be4dbb34`}, {`regress_passwd2`, `md54044304ba511dd062133eb5b4b84a2a3`}, {`regress_passwd3`, `SCRAM-SHA-256$4096:<salt>$<storedkey>:<serverkey>`}, {`regress_passwd4`, ``}},
			},
			{
				Statement: `ALTER ROLE regress_passwd2 RENAME TO regress_passwd2_new;`,
			},
			{
				Statement: `SELECT rolname, rolpassword
    FROM pg_authid
    WHERE rolname LIKE 'regress_passwd2_new'
    ORDER BY rolname, rolpassword;`,
				Results: []sql.Row{{`regress_passwd2_new`, ``}},
			},
			{
				Statement: `ALTER ROLE regress_passwd2_new RENAME TO regress_passwd2;`,
			},
			{
				Statement: `SET password_encryption = 'md5';`,
			},
			{
				Statement: `ALTER ROLE regress_passwd2 PASSWORD 'foo';`,
			},
			{
				Statement: `ALTER ROLE regress_passwd1 PASSWORD 'md5cd3578025fe2c3d7ed1b9a9b26238b70';`,
			},
			{
				Statement: `ALTER ROLE regress_passwd3 PASSWORD 'SCRAM-SHA-256$4096:VLK4RMaQLCvNtQ==$6YtlR4t69SguDiwFvbVgVZtuz6gpJQQqUMZ7IQJK5yI=:ps75jrHeYU4lXCcXI4O8oIdJ3eO8o2jirjruw9phBTo=';`,
			},
			{
				Statement: `SET password_encryption = 'scram-sha-256';`,
			},
			{
				Statement: `ALTER ROLE  regress_passwd4 PASSWORD 'foo';`,
			},
			{
				Statement: `CREATE ROLE regress_passwd5 PASSWORD 'md5e73a4b11df52a6068f8b39f90be36023';`,
			},
			{
				Statement: `CREATE ROLE regress_passwd6 PASSWORD 'SCRAM-SHA-256$1234';`,
			},
			{
				Statement: `CREATE ROLE regress_passwd7 PASSWORD 'md5012345678901234567890123456789zz';`,
			},
			{
				Statement: `CREATE ROLE regress_passwd8 PASSWORD 'md501234567890123456789012345678901zz';`,
			},
			{
				Statement: `SELECT rolname, regexp_replace(rolpassword, '(SCRAM-SHA-256)\$(\d+):([a-zA-Z0-9+/=]+)\$([a-zA-Z0-9+=/]+):([a-zA-Z0-9+/=]+)', '\1$\2:<salt>$<storedkey>:<serverkey>') as rolpassword_masked
    FROM pg_authid
    WHERE rolname LIKE 'regress_passwd%'
    ORDER BY rolname, rolpassword;`,
				Results: []sql.Row{{`regress_passwd1`, `md5cd3578025fe2c3d7ed1b9a9b26238b70`}, {`regress_passwd2`, `md5dfa155cadd5f4ad57860162f3fab9cdb`}, {`regress_passwd3`, `SCRAM-SHA-256$4096:<salt>$<storedkey>:<serverkey>`}, {`regress_passwd4`, `SCRAM-SHA-256$4096:<salt>$<storedkey>:<serverkey>`}, {`regress_passwd5`, `md5e73a4b11df52a6068f8b39f90be36023`}, {`regress_passwd6`, `SCRAM-SHA-256$4096:<salt>$<storedkey>:<serverkey>`}, {`regress_passwd7`, `SCRAM-SHA-256$4096:<salt>$<storedkey>:<serverkey>`}, {`regress_passwd8`, `SCRAM-SHA-256$4096:<salt>$<storedkey>:<serverkey>`}},
			},
			{
				Statement: `CREATE ROLE regress_passwd_empty PASSWORD '';`,
			},
			{
				Statement: `ALTER ROLE regress_passwd_empty PASSWORD 'md585939a5ce845f1a1b620742e3c659e0a';`,
			},
			{
				Statement: `ALTER ROLE regress_passwd_empty PASSWORD 'SCRAM-SHA-256$4096:hpFyHTUsSWcR7O9P$LgZFIt6Oqdo27ZFKbZ2nV+vtnYM995pDh9ca6WSi120=:qVV5NeluNfUPkwm7Vqat25RjSPLkGeoZBQs6wVv+um4=';`,
			},
			{
				Statement: `SELECT rolpassword FROM pg_authid WHERE rolname='regress_passwd_empty';`,
				Results:   []sql.Row{{``}},
			},
			{
				Statement: `CREATE ROLE regress_passwd_sha_len0 PASSWORD 'SCRAM-SHA-256$4096:A6xHKoH/494E941doaPOYg==$Ky+A30sewHIH3VHQLRN9vYsuzlgNyGNKCh37dy96Rqw=:COPdlNiIkrsacU5QoxydEuOH6e/KfiipeETb/bPw8ZI=';`,
			},
			{
				Statement: `CREATE ROLE regress_passwd_sha_len1 PASSWORD 'SCRAM-SHA-256$4096:A6xHKoH/494E941doaPOYg==$Ky+A30sewHIH3VHQLRN9vYsuzlgNyGNKCh37dy96RqwAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=:COPdlNiIkrsacU5QoxydEuOH6e/KfiipeETb/bPw8ZI=';`,
			},
			{
				Statement: `CREATE ROLE regress_passwd_sha_len2 PASSWORD 'SCRAM-SHA-256$4096:A6xHKoH/494E941doaPOYg==$Ky+A30sewHIH3VHQLRN9vYsuzlgNyGNKCh37dy96Rqw=:COPdlNiIkrsacU5QoxydEuOH6e/KfiipeETb/bPw8ZIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=';`,
			},
			{
				Statement: `SELECT rolname, rolpassword not like '%A6xHKoH/494E941doaPOYg==%' as is_rolpassword_rehashed
    FROM pg_authid
    WHERE rolname LIKE 'regress_passwd_sha_len%'
    ORDER BY rolname;`,
				Results: []sql.Row{{`regress_passwd_sha_len0`, false}, {`regress_passwd_sha_len1`, true}, {`regress_passwd_sha_len2`, true}},
			},
			{
				Statement: `DROP ROLE regress_passwd1;`,
			},
			{
				Statement: `DROP ROLE regress_passwd2;`,
			},
			{
				Statement: `DROP ROLE regress_passwd3;`,
			},
			{
				Statement: `DROP ROLE regress_passwd4;`,
			},
			{
				Statement: `DROP ROLE regress_passwd5;`,
			},
			{
				Statement: `DROP ROLE regress_passwd6;`,
			},
			{
				Statement: `DROP ROLE regress_passwd7;`,
			},
			{
				Statement: `DROP ROLE regress_passwd8;`,
			},
			{
				Statement: `DROP ROLE regress_passwd_empty;`,
			},
			{
				Statement: `DROP ROLE regress_passwd_sha_len0;`,
			},
			{
				Statement: `DROP ROLE regress_passwd_sha_len1;`,
			},
			{
				Statement: `DROP ROLE regress_passwd_sha_len2;`,
			},
			{
				Statement: `SELECT rolname, rolpassword
    FROM pg_authid
    WHERE rolname LIKE 'regress_passwd%'
    ORDER BY rolname, rolpassword;`,
				Results: []sql.Row{},
			},
		},
	})
}
