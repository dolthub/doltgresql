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

package auth

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestAuthorizationHandlerDbNameStripsRevisions(t *testing.T) {
	ctx := sql.NewEmptyContext()
	ctx.SetCurrentDatabase("current/branch")
	h := &AuthorizationHandler{}
	for input, expected := range map[string]string{
		"":           "current",
		"db":         "db",
		"db/b1":      "db",
		"db@abc":     "db",
		"Db/Feature": "Db",
	} {
		if actual := h.dbName(ctx, input); actual != expected {
			t.Errorf("dbName(%q) = %q, expected %q", input, actual, expected)
		}
	}
}
