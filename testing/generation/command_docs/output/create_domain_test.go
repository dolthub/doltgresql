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

func TestCreateDomain(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("CREATE DOMAIN name data_type"),
		Unimplemented("CREATE DOMAIN name AS data_type"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression"),
		Unimplemented("CREATE DOMAIN name data_type NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression NULL"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type NOT NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type NOT NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US NOT NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US NOT NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression NOT NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression NOT NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression NOT NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression NOT NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name NOT NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name NOT NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name NOT NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name NOT NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name NOT NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name NOT NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NOT NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NOT NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NULL NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type CHECK ( expression ) NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type CHECK ( expression ) NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CHECK ( expression ) NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CHECK ( expression ) NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CHECK ( expression ) NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CHECK ( expression ) NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CHECK ( expression ) NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CHECK ( expression ) NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name CHECK ( expression ) NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name CHECK ( expression ) NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name CHECK ( expression ) NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name CHECK ( expression ) NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type NOT NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type NOT NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US NOT NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US NOT NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression NOT NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression NOT NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression NOT NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression NOT NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NULL CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type CHECK ( expression ) CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type CHECK ( expression ) CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CHECK ( expression ) CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CHECK ( expression ) CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CHECK ( expression ) CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CHECK ( expression ) CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CHECK ( expression ) CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CHECK ( expression ) CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name NOT NULL"),
		Unimplemented("CREATE DOMAIN name data_type NOT NULL NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type NOT NULL NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US NOT NULL NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US NOT NULL NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression NOT NULL NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression NOT NULL NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression NOT NULL NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression NOT NULL NULL"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name NOT NULL NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name NOT NULL NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name NOT NULL NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name NOT NULL NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name NOT NULL NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name NOT NULL NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NOT NULL NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NOT NULL NULL"),
		Unimplemented("CREATE DOMAIN name data_type NULL NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type NULL NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US NULL NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US NULL NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression NULL NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression NULL NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression NULL NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression NULL NULL"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name NULL NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name NULL NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name NULL NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name NULL NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name NULL NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name NULL NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NULL NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NULL NULL"),
		Unimplemented("CREATE DOMAIN name data_type CHECK ( expression ) NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type CHECK ( expression ) NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CHECK ( expression ) NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CHECK ( expression ) NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CHECK ( expression ) NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CHECK ( expression ) NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CHECK ( expression ) NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CHECK ( expression ) NULL"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name CHECK ( expression ) NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name CHECK ( expression ) NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name CHECK ( expression ) NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name CHECK ( expression ) NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) NULL"),
		Unimplemented("CREATE DOMAIN name data_type NOT NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type NOT NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US NOT NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US NOT NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression NOT NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression NOT NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression NOT NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression NOT NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NULL CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type CHECK ( expression ) CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type CHECK ( expression ) CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CHECK ( expression ) CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CHECK ( expression ) CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CHECK ( expression ) CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CHECK ( expression ) CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CHECK ( expression ) CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CHECK ( expression ) CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name NULL"),
		Unimplemented("CREATE DOMAIN name data_type NOT NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type NOT NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US NOT NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US NOT NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression NOT NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression NOT NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression NOT NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression NOT NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name NOT NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name NOT NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name NOT NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name NOT NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name NOT NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name NOT NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NOT NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NOT NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NULL CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type CHECK ( expression ) CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type CHECK ( expression ) CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CHECK ( expression ) CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CHECK ( expression ) CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CHECK ( expression ) CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CHECK ( expression ) CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CHECK ( expression ) CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CHECK ( expression ) CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name CHECK ( expression ) CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name CHECK ( expression ) CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name CHECK ( expression ) CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name CHECK ( expression ) CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type NOT NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type NOT NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US NOT NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US NOT NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression NOT NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression NOT NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression NOT NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression NOT NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NOT NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name NULL CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type CHECK ( expression ) CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type CHECK ( expression ) CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CHECK ( expression ) CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CHECK ( expression ) CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CHECK ( expression ) CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CHECK ( expression ) CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CHECK ( expression ) CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CHECK ( expression ) CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name CHECK ( expression )"),
		Unimplemented("CREATE DOMAIN name AS data_type COLLATE en_US DEFAULT expression CONSTRAINT constraint_name CHECK ( expression ) CONSTRAINT constraint_name CHECK ( expression )"),
	}
	RunTests(t, tests)
}
