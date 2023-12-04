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

func TestAlterSubscription(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("ALTER SUBSCRIPTION name CONNECTION ' conninfo '"),
		Unimplemented("ALTER SUBSCRIPTION name SET PUBLICATION publication_name"),
		Unimplemented("ALTER SUBSCRIPTION name SET PUBLICATION publication_name , publication_name"),
		Unimplemented("ALTER SUBSCRIPTION name SET PUBLICATION publication_name WITH ( publication_option )"),
		Unimplemented("ALTER SUBSCRIPTION name SET PUBLICATION publication_name , publication_name WITH ( publication_option )"),
		Unimplemented("ALTER SUBSCRIPTION name SET PUBLICATION publication_name WITH ( publication_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name SET PUBLICATION publication_name , publication_name WITH ( publication_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name SET PUBLICATION publication_name WITH ( publication_option , publication_option )"),
		Unimplemented("ALTER SUBSCRIPTION name SET PUBLICATION publication_name , publication_name WITH ( publication_option , publication_option )"),
		Unimplemented("ALTER SUBSCRIPTION name SET PUBLICATION publication_name WITH ( publication_option = value , publication_option )"),
		Unimplemented("ALTER SUBSCRIPTION name SET PUBLICATION publication_name , publication_name WITH ( publication_option = value , publication_option )"),
		Unimplemented("ALTER SUBSCRIPTION name SET PUBLICATION publication_name WITH ( publication_option , publication_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name SET PUBLICATION publication_name , publication_name WITH ( publication_option , publication_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name SET PUBLICATION publication_name WITH ( publication_option = value , publication_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name SET PUBLICATION publication_name , publication_name WITH ( publication_option = value , publication_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name ADD PUBLICATION publication_name"),
		Unimplemented("ALTER SUBSCRIPTION name ADD PUBLICATION publication_name , publication_name"),
		Unimplemented("ALTER SUBSCRIPTION name ADD PUBLICATION publication_name WITH ( publication_option )"),
		Unimplemented("ALTER SUBSCRIPTION name ADD PUBLICATION publication_name , publication_name WITH ( publication_option )"),
		Unimplemented("ALTER SUBSCRIPTION name ADD PUBLICATION publication_name WITH ( publication_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name ADD PUBLICATION publication_name , publication_name WITH ( publication_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name ADD PUBLICATION publication_name WITH ( publication_option , publication_option )"),
		Unimplemented("ALTER SUBSCRIPTION name ADD PUBLICATION publication_name , publication_name WITH ( publication_option , publication_option )"),
		Unimplemented("ALTER SUBSCRIPTION name ADD PUBLICATION publication_name WITH ( publication_option = value , publication_option )"),
		Unimplemented("ALTER SUBSCRIPTION name ADD PUBLICATION publication_name , publication_name WITH ( publication_option = value , publication_option )"),
		Unimplemented("ALTER SUBSCRIPTION name ADD PUBLICATION publication_name WITH ( publication_option , publication_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name ADD PUBLICATION publication_name , publication_name WITH ( publication_option , publication_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name ADD PUBLICATION publication_name WITH ( publication_option = value , publication_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name ADD PUBLICATION publication_name , publication_name WITH ( publication_option = value , publication_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name DROP PUBLICATION publication_name"),
		Unimplemented("ALTER SUBSCRIPTION name DROP PUBLICATION publication_name , publication_name"),
		Unimplemented("ALTER SUBSCRIPTION name DROP PUBLICATION publication_name WITH ( publication_option )"),
		Unimplemented("ALTER SUBSCRIPTION name DROP PUBLICATION publication_name , publication_name WITH ( publication_option )"),
		Unimplemented("ALTER SUBSCRIPTION name DROP PUBLICATION publication_name WITH ( publication_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name DROP PUBLICATION publication_name , publication_name WITH ( publication_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name DROP PUBLICATION publication_name WITH ( publication_option , publication_option )"),
		Unimplemented("ALTER SUBSCRIPTION name DROP PUBLICATION publication_name , publication_name WITH ( publication_option , publication_option )"),
		Unimplemented("ALTER SUBSCRIPTION name DROP PUBLICATION publication_name WITH ( publication_option = value , publication_option )"),
		Unimplemented("ALTER SUBSCRIPTION name DROP PUBLICATION publication_name , publication_name WITH ( publication_option = value , publication_option )"),
		Unimplemented("ALTER SUBSCRIPTION name DROP PUBLICATION publication_name WITH ( publication_option , publication_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name DROP PUBLICATION publication_name , publication_name WITH ( publication_option , publication_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name DROP PUBLICATION publication_name WITH ( publication_option = value , publication_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name DROP PUBLICATION publication_name , publication_name WITH ( publication_option = value , publication_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name REFRESH PUBLICATION"),
		Unimplemented("ALTER SUBSCRIPTION name REFRESH PUBLICATION WITH ( refresh_option )"),
		Unimplemented("ALTER SUBSCRIPTION name REFRESH PUBLICATION WITH ( refresh_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name REFRESH PUBLICATION WITH ( refresh_option , refresh_option )"),
		Unimplemented("ALTER SUBSCRIPTION name REFRESH PUBLICATION WITH ( refresh_option = value , refresh_option )"),
		Unimplemented("ALTER SUBSCRIPTION name REFRESH PUBLICATION WITH ( refresh_option , refresh_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name REFRESH PUBLICATION WITH ( refresh_option = value , refresh_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name ENABLE"),
		Unimplemented("ALTER SUBSCRIPTION name DISABLE"),
		Unimplemented("ALTER SUBSCRIPTION name SET ( subscription_parameter )"),
		Unimplemented("ALTER SUBSCRIPTION name SET ( subscription_parameter = value )"),
		Unimplemented("ALTER SUBSCRIPTION name SET ( subscription_parameter , subscription_parameter )"),
		Unimplemented("ALTER SUBSCRIPTION name SET ( subscription_parameter = value , subscription_parameter )"),
		Unimplemented("ALTER SUBSCRIPTION name SET ( subscription_parameter , subscription_parameter = value )"),
		Unimplemented("ALTER SUBSCRIPTION name SET ( subscription_parameter = value , subscription_parameter = value )"),
		Unimplemented("ALTER SUBSCRIPTION name SKIP ( skip_option = value )"),
		Unimplemented("ALTER SUBSCRIPTION name OWNER TO new_owner"),
		Unimplemented("ALTER SUBSCRIPTION name OWNER TO CURRENT_ROLE"),
		Unimplemented("ALTER SUBSCRIPTION name OWNER TO CURRENT_USER"),
		Unimplemented("ALTER SUBSCRIPTION name OWNER TO SESSION_USER"),
		Unimplemented("ALTER SUBSCRIPTION name RENAME TO new_name"),
	}
	RunTests(t, tests)
}
