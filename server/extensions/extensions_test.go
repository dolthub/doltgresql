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

package extensions

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/server/extensions/extdef"
)

// testFunction is a stand-in implementation for the extension that these tests register.
func testFunction(ctx *sql.Context, args ...any) (any, error) {
	return "called", nil
}

func TestRegistry(t *testing.T) {
	register(&extdef.Extension{
		Name:     "doltgres_test",
		Control:  extdef.Control{DefaultVersion: "2.5", Comment: "a test extension", Relocatable: true},
		Routines: []extdef.Routine{{Name: "alpha", Symbol: "alpha", Returns: "uuid", Impl: testFunction}},
	})

	ext, err := Get("doltgres_test")
	require.NoError(t, err)
	require.Equal(t, "alpha", ext.Routines[0].Name)
	require.Equal(t, "2.5", ext.Control.DefaultVersion)
	require.Equal(t, "a test extension", ext.Control.Comment)
	require.True(t, ext.Control.Relocatable)
	require.Contains(t, GetAll(), "doltgres_test")

	f, err := GetFunction("doltgres_test", "alpha")
	require.NoError(t, err)
	result, err := f(nil)
	require.NoError(t, err)
	require.Equal(t, "called", result)

	_, err = Get("doltgres_test_missing")
	require.ErrorContains(t, err, `extension "doltgres_test_missing" is not available`)
	// Extension names are case-sensitive in Postgres, so a differently-cased name does not match
	_, err = Get("DOLTGRES_TEST")
	require.ErrorContains(t, err, `extension "DOLTGRES_TEST" is not available`)
	_, err = GetFunction("doltgres_test", "beta")
	require.ErrorContains(t, err, `extension "doltgres_test" does not declare the function "beta"`)
	_, err = GetFunction("doltgres_test_missing", "alpha")
	require.ErrorContains(t, err, `extension "doltgres_test_missing" is not available`)

	// Registering the same extension twice would silently replace the first, so it panics instead
	require.Panics(t, func() {
		register(&extdef.Extension{Name: "doltgres_test", Control: extdef.Control{DefaultVersion: "2.5"}})
	})
	// A symbol may only back one routine, since it is what dispatches a call to its implementation
	require.Panics(t, func() {
		register(&extdef.Extension{
			Name: "doltgres_test_symbols",
			Routines: []extdef.Routine{
				{Name: "alpha", Symbol: "alpha", Impl: testFunction},
				{Name: "beta", Symbol: "alpha", Impl: testFunction},
			},
		})
	})
}
