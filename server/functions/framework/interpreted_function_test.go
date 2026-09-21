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

package framework_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/functions/framework"
	"github.com/dolthub/doltgresql/server/plpgsql"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// TestMain initializes the function registry, which can only be done once for the whole package.
func TestMain(m *testing.M) {
	functions.Init()
	framework.Initialize(nil)
	os.Exit(m.Run())
}

// TestApplyBindings_RendersRegclassVariableThroughSessionContext asserts
// that ApplyBindings renders a regclass-typed variable using the
// session context it is given.
//
// See https://github.com/dolthub/doltgresql/issues/1142.
func TestApplyBindings_RendersRegclassVariableThroughSessionContext(t *testing.T) {
	t.Parallel()
	ctx := sql.NewEmptyContext()
	stack := plpgsql.NewInterpreterStack(nil)
	stack.NewVariableWithValue("rel", pgtypes.Regclass, id.NewOID(1259).AsId())

	require.NotPanics(t, func() {
		_, _, _ = framework.InterpretedFunction{}.ApplyBindings(ctx, stack, "SELECT $1", []string{"rel"}, false)
	})

	// Every occurrence must be replaced, and shorter placeholder indexes must not alter longer ones.
	bindings := make([]string, 10)
	for i := range bindings {
		bindings[i] = fmt.Sprintf("v%d", i+1)
		stack.NewVariableWithValue(bindings[i], pgtypes.Int32, int32(i+1))
	}

	stmt, found, err := framework.InterpretedFunction{}.ApplyBindings(
		ctx, stack, "SELECT $1, $1, $10, $10, $2", bindings, false)
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "SELECT 1, 1, 10, 10, 2", stmt)
}

// TestApplyBindings_RendersWholeRecord asserts that a reference to a record as a whole, such as a trigger's
// OLD or NEW, renders as an expression carrying the record's field names.
func TestApplyBindings_RendersWholeRecord(t *testing.T) {
	t.Parallel()
	ctx := sql.NewEmptyContext()
	stack := plpgsql.NewInterpreterStack(nil)
	sch := sql.Schema{
		{Name: "pk", Type: pgtypes.Int32},
		{Name: "v1", Type: pgtypes.Text},
		{Name: "v2", Type: pgtypes.Text},
	}
	stack.NewRecord("new", sch, sql.Row{int32(1), "hi", nil})

	t.Run("enforced type carries the field names", func(t *testing.T) {
		stmt, varFound, err := framework.InterpretedFunction{}.ApplyBindings(ctx, stack, "SELECT to_jsonb($1)", []string{"new"}, true)
		require.NoError(t, err)
		require.True(t, varFound)
		require.Equal(t,
			`SELECT to_jsonb((SELECT "record" FROM (SELECT ((1)::integer) AS "pk", (('hi')::text) AS "v1", ((NULL)::text) AS "v2") "record"))`,
			stmt)
	})

	t.Run("unenforced type renders the record's text form", func(t *testing.T) {
		stmt, varFound, err := framework.InterpretedFunction{}.ApplyBindings(ctx, stack, "$1", []string{"new"}, false)
		require.NoError(t, err)
		require.True(t, varFound)
		require.Equal(t, `(1,hi,)`, stmt)
	})

	t.Run("a record that has no shape yet has no fields to render", func(t *testing.T) {
		stack.NewRecord("r", nil, nil)
		_, varFound, err := framework.InterpretedFunction{}.ApplyBindings(ctx, stack, "SELECT $1", []string{"r"}, true)
		require.True(t, varFound)
		require.True(t, plpgsql.ErrRecordNotAssigned.Is(err), "got %v", err)
	})
}
