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

package enginetest

import (
	"os"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	gmserrors "gopkg.in/src-d/go-errors.v1"
)

// Keep the original imported Dolt scripts and their native translations visibly
// separate. Translation never chooses expectations from the server's response.
func runEngineTargetPair(t *testing.T, run func(*testing.T)) {
	target := os.Getenv("DOLTGRES_TEST_TARGET")
	for _, name := range []string{"doltgres", "postgres"} {
		t.Run(name, func(t *testing.T) {
			native := target == "postgres" || target == "postgres-fork"
			if (name == "postgres") != native {
				t.Skip("counterpart selected by DOLTGRES_TEST_TARGET")
			}
			run(t)
		})
	}
}

// These counterparts assert the same error category, without matching MySQL prose.
var nativeErrorCodes = map[*gmserrors.Kind]string{
	sql.ErrTableNotFound:                 "42P01",
	sql.ErrColumnNotFound:                "42703",
	sql.ErrAmbiguousColumnName:           "42702",
	sql.ErrAmbiguousColumnOrAliasName:    "42702",
	sql.ErrAmbiguousColumnInOrderBy:      "42702",
	sql.ErrPrimaryKeyViolation:           "23505",
	sql.ErrUniqueKeyViolation:            "23505",
	sql.ErrForeignKeyChildViolation:      "23503",
	sql.ErrForeignKeyParentViolation:     "23503",
	sql.ErrForeignKeyColumnCountMismatch: "42830",
	sql.ErrForeignKeyColumnTypeMismatch:  "42804",
}
