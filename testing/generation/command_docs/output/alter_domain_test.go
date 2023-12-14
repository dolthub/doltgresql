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
		Unimplemented("ALTER DOMAIN name SET DEFAULT expression"),
		Unimplemented("ALTER DOMAIN name DROP DEFAULT"),
		Unimplemented("ALTER DOMAIN name SET NOT NULL"),
		Unimplemented("ALTER DOMAIN name DROP NOT NULL"),
		Unimplemented("ALTER DOMAIN name ADD CONSTRAINT name CHECK ( condition )"),
		Unimplemented("ALTER DOMAIN name ADD CONSTRAINT name CHECK ( condition ) NOT VALID"),
		Unimplemented("ALTER DOMAIN name DROP CONSTRAINT constraint_name"),
		Unimplemented("ALTER DOMAIN name DROP CONSTRAINT IF EXISTS constraint_name"),
		Unimplemented("ALTER DOMAIN name DROP CONSTRAINT constraint_name RESTRICT"),
		Unimplemented("ALTER DOMAIN name DROP CONSTRAINT IF EXISTS constraint_name RESTRICT"),
		Unimplemented("ALTER DOMAIN name DROP CONSTRAINT constraint_name CASCADE"),
		Unimplemented("ALTER DOMAIN name DROP CONSTRAINT IF EXISTS constraint_name CASCADE"),
		Unimplemented("ALTER DOMAIN name RENAME CONSTRAINT constraint_name TO new_constraint_name"),
		Unimplemented("ALTER DOMAIN name VALIDATE CONSTRAINT constraint_name"),
		Unimplemented("ALTER DOMAIN name OWNER TO new_owner"),
		Unimplemented("ALTER DOMAIN name OWNER TO CURRENT_ROLE"),
		Unimplemented("ALTER DOMAIN name OWNER TO CURRENT_USER"),
		Unimplemented("ALTER DOMAIN name OWNER TO SESSION_USER"),
		Unimplemented("ALTER DOMAIN name RENAME TO new_name"),
		Unimplemented("ALTER DOMAIN name SET SCHEMA new_schema"),
	}
	RunTests(t, tests)
}
