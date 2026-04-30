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
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/functions/framework"
	"github.com/dolthub/doltgresql/server/plpgsql"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// TestApplyBindings_RendersRegclassVariableThroughSessionContext asserts
// that ApplyBindings renders a regclass-typed variable using the
// session context it is given.
//
// See https://github.com/dolthub/doltgresql/issues/1142.
func TestApplyBindings_RendersRegclassVariableThroughSessionContext(t *testing.T) {
	t.Parallel()
	functions.Init()
	framework.Initialize()
	ctx := sql.NewEmptyContext()
	stack := plpgsql.NewInterpreterStack(nil)
	stack.NewVariableWithValue("rel", pgtypes.Regclass, id.NewOID(1259).AsId())

	require.NotPanics(t, func() {
		_, _, _ = framework.InterpretedFunction{}.ApplyBindings(ctx, stack, "SELECT $1", []string{"rel"}, false)
	})
}
