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

package types

import "strings"

// QuoteString will quote the string according to the type given. This means that some types will quote, and others will
// not, or they may quote in a special way that is unique to that type.
func QuoteString(baseID DoltgresTypeBaseID, str string) string {
	switch baseID {
	case DoltgresTypeBaseID_Char, DoltgresTypeBaseID_Name, DoltgresTypeBaseID_Text, DoltgresTypeBaseID_VarChar:
		return `'` + strings.ReplaceAll(str, `'`, `''`) + `'`
	default:
		return str
	}
}
