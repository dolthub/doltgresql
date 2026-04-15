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

package core

// IsValidPostgresIdentifier returns true according to Postgres quoted identifier rules.
// Quoted identifiers can contain any character except the null character (code zero),
// including supplementary Unicode (emoji, code points above U+FFFF) unlike MySQL.
// https://www.postgresql.org/docs/current/sql-syntax-lexical.html
func IsValidPostgresIdentifier(name string) bool {
	if len(name) == 0 {
		return false
	}
	for _, c := range name {
		if c == 0x0000 {
			return false
		}
	}
	return true
}
