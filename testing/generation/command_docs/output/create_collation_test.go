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

func TestCreateCollation(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("CREATE COLLATION name ( LOCALE = locale )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LOCALE = locale )"),
		Unimplemented("CREATE COLLATION name ( LC_COLLATE = lc_collate )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_COLLATE = lc_collate )"),
		Unimplemented("CREATE COLLATION name ( LC_COLLATE = lc_collate , LC_CTYPE = lc_ctype )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_COLLATE = lc_collate , LC_CTYPE = lc_ctype )"),
		Unimplemented("CREATE COLLATION name ( LC_CTYPE = lc_ctype )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_CTYPE = lc_ctype )"),
		Unimplemented("CREATE COLLATION name ( LOCALE = locale , PROVIDER = provider )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LOCALE = locale , PROVIDER = provider )"),
		Unimplemented("CREATE COLLATION name ( LC_COLLATE = lc_collate , PROVIDER = provider )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_COLLATE = lc_collate , PROVIDER = provider )"),
		Unimplemented("CREATE COLLATION name ( LC_COLLATE = lc_collate , LC_CTYPE = lc_ctype , PROVIDER = provider )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_COLLATE = lc_collate , LC_CTYPE = lc_ctype , PROVIDER = provider )"),
		Unimplemented("CREATE COLLATION name ( LC_CTYPE = lc_ctype , PROVIDER = provider )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_CTYPE = lc_ctype , PROVIDER = provider )"),
		Unimplemented("CREATE COLLATION name ( LOCALE = locale , DETERMINISTIC = true )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LOCALE = locale , DETERMINISTIC = true )"),
		Unimplemented("CREATE COLLATION name ( LC_COLLATE = lc_collate , DETERMINISTIC = true )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_COLLATE = lc_collate , DETERMINISTIC = true )"),
		Unimplemented("CREATE COLLATION name ( LC_COLLATE = lc_collate , LC_CTYPE = lc_ctype , DETERMINISTIC = true )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_COLLATE = lc_collate , LC_CTYPE = lc_ctype , DETERMINISTIC = true )"),
		Unimplemented("CREATE COLLATION name ( LC_CTYPE = lc_ctype , DETERMINISTIC = true )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_CTYPE = lc_ctype , DETERMINISTIC = true )"),
		Unimplemented("CREATE COLLATION name ( LOCALE = locale , PROVIDER = provider , DETERMINISTIC = true )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LOCALE = locale , PROVIDER = provider , DETERMINISTIC = true )"),
		Unimplemented("CREATE COLLATION name ( LC_COLLATE = lc_collate , PROVIDER = provider , DETERMINISTIC = true )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_COLLATE = lc_collate , PROVIDER = provider , DETERMINISTIC = true )"),
		Unimplemented("CREATE COLLATION name ( LC_COLLATE = lc_collate , LC_CTYPE = lc_ctype , PROVIDER = provider , DETERMINISTIC = true )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_COLLATE = lc_collate , LC_CTYPE = lc_ctype , PROVIDER = provider , DETERMINISTIC = true )"),
		Unimplemented("CREATE COLLATION name ( LC_CTYPE = lc_ctype , PROVIDER = provider , DETERMINISTIC = true )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_CTYPE = lc_ctype , PROVIDER = provider , DETERMINISTIC = true )"),
		Unimplemented("CREATE COLLATION name ( LOCALE = locale , VERSION = version )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LOCALE = locale , VERSION = version )"),
		Unimplemented("CREATE COLLATION name ( LC_COLLATE = lc_collate , VERSION = version )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_COLLATE = lc_collate , VERSION = version )"),
		Unimplemented("CREATE COLLATION name ( LC_COLLATE = lc_collate , LC_CTYPE = lc_ctype , VERSION = version )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_COLLATE = lc_collate , LC_CTYPE = lc_ctype , VERSION = version )"),
		Unimplemented("CREATE COLLATION name ( LC_CTYPE = lc_ctype , VERSION = version )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_CTYPE = lc_ctype , VERSION = version )"),
		Unimplemented("CREATE COLLATION name ( LOCALE = locale , PROVIDER = provider , VERSION = version )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LOCALE = locale , PROVIDER = provider , VERSION = version )"),
		Unimplemented("CREATE COLLATION name ( LC_COLLATE = lc_collate , PROVIDER = provider , VERSION = version )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_COLLATE = lc_collate , PROVIDER = provider , VERSION = version )"),
		Unimplemented("CREATE COLLATION name ( LC_COLLATE = lc_collate , LC_CTYPE = lc_ctype , PROVIDER = provider , VERSION = version )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_COLLATE = lc_collate , LC_CTYPE = lc_ctype , PROVIDER = provider , VERSION = version )"),
		Unimplemented("CREATE COLLATION name ( LC_CTYPE = lc_ctype , PROVIDER = provider , VERSION = version )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_CTYPE = lc_ctype , PROVIDER = provider , VERSION = version )"),
		Unimplemented("CREATE COLLATION name ( LOCALE = locale , DETERMINISTIC = true , VERSION = version )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LOCALE = locale , DETERMINISTIC = true , VERSION = version )"),
		Unimplemented("CREATE COLLATION name ( LC_COLLATE = lc_collate , DETERMINISTIC = true , VERSION = version )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_COLLATE = lc_collate , DETERMINISTIC = true , VERSION = version )"),
		Unimplemented("CREATE COLLATION name ( LC_COLLATE = lc_collate , LC_CTYPE = lc_ctype , DETERMINISTIC = true , VERSION = version )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_COLLATE = lc_collate , LC_CTYPE = lc_ctype , DETERMINISTIC = true , VERSION = version )"),
		Unimplemented("CREATE COLLATION name ( LC_CTYPE = lc_ctype , DETERMINISTIC = true , VERSION = version )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_CTYPE = lc_ctype , DETERMINISTIC = true , VERSION = version )"),
		Unimplemented("CREATE COLLATION name ( LOCALE = locale , PROVIDER = provider , DETERMINISTIC = true , VERSION = version )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LOCALE = locale , PROVIDER = provider , DETERMINISTIC = true , VERSION = version )"),
		Unimplemented("CREATE COLLATION name ( LC_COLLATE = lc_collate , PROVIDER = provider , DETERMINISTIC = true , VERSION = version )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_COLLATE = lc_collate , PROVIDER = provider , DETERMINISTIC = true , VERSION = version )"),
		Unimplemented("CREATE COLLATION name ( LC_COLLATE = lc_collate , LC_CTYPE = lc_ctype , PROVIDER = provider , DETERMINISTIC = true , VERSION = version )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_COLLATE = lc_collate , LC_CTYPE = lc_ctype , PROVIDER = provider , DETERMINISTIC = true , VERSION = version )"),
		Unimplemented("CREATE COLLATION name ( LC_CTYPE = lc_ctype , PROVIDER = provider , DETERMINISTIC = true , VERSION = version )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( LC_CTYPE = lc_ctype , PROVIDER = provider , DETERMINISTIC = true , VERSION = version )"),
		Unimplemented("CREATE COLLATION name ( PROVIDER = provider )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( PROVIDER = provider )"),
		Unimplemented("CREATE COLLATION name ( PROVIDER = provider , DETERMINISTIC = true )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( PROVIDER = provider , DETERMINISTIC = true )"),
		Unimplemented("CREATE COLLATION name ( PROVIDER = provider , VERSION = version )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( PROVIDER = provider , VERSION = version )"),
		Unimplemented("CREATE COLLATION name ( PROVIDER = provider , DETERMINISTIC = true , VERSION = version )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( PROVIDER = provider , DETERMINISTIC = true , VERSION = version )"),
		Unimplemented("CREATE COLLATION name ( DETERMINISTIC = true )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( DETERMINISTIC = true )"),
		Unimplemented("CREATE COLLATION name ( DETERMINISTIC = true , VERSION = version )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( DETERMINISTIC = true , VERSION = version )"),
		Unimplemented("CREATE COLLATION name ( VERSION = version )"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name ( VERSION = version )"),
		Unimplemented("CREATE COLLATION name FROM existing_collation"),
		Unimplemented("CREATE COLLATION IF NOT EXISTS name FROM existing_collation"),
	}
	RunTests(t, tests)
}
