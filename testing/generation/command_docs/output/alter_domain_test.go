// Copyright 2023 Dolthub, Inc.
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

package output

import "testing"

func TestAlterDomain(t *testing.T) {
	tests := []QueryParses{
		Parses("ALTER DOMAIN name SET DEFAULT expression"),
		Parses("ALTER DOMAIN name DROP DEFAULT"),
		Parses("ALTER DOMAIN name SET NOT NULL"),
		Parses("ALTER DOMAIN name DROP NOT NULL"),
		Parses("ALTER DOMAIN name ADD CONSTRAINT name CHECK ( condition )"),
		Parses("ALTER DOMAIN name ADD CONSTRAINT name CHECK ( condition ) NOT VALID"),
		Parses("ALTER DOMAIN name DROP CONSTRAINT constraint_name"),
		Parses("ALTER DOMAIN name DROP CONSTRAINT IF EXISTS constraint_name"),
		Parses("ALTER DOMAIN name DROP CONSTRAINT constraint_name RESTRICT"),
		Parses("ALTER DOMAIN name DROP CONSTRAINT IF EXISTS constraint_name RESTRICT"),
		Parses("ALTER DOMAIN name DROP CONSTRAINT constraint_name CASCADE"),
		Parses("ALTER DOMAIN name DROP CONSTRAINT IF EXISTS constraint_name CASCADE"),
		Parses("ALTER DOMAIN name RENAME CONSTRAINT constraint_name TO new_constraint_name"),
		Parses("ALTER DOMAIN name VALIDATE CONSTRAINT constraint_name"),
		Parses("ALTER DOMAIN name OWNER TO new_owner"),
		Parses("ALTER DOMAIN name OWNER TO CURRENT_ROLE"),
		Parses("ALTER DOMAIN name OWNER TO CURRENT_USER"),
		Parses("ALTER DOMAIN name OWNER TO SESSION_USER"),
		Parses("ALTER DOMAIN name RENAME TO new_name"),
		Parses("ALTER DOMAIN name SET SCHEMA new_schema"),
	}
	RunTests(t, tests)
}
