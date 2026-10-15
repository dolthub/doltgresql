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

package ast_test

import (
	"context"
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/server/ast"
)

func TestConvertOptionsWithContext(t *testing.T) {
	bound := sqle.NewEngineVariables()
	ctx := sql.NewContextWithVariables(context.Background(), bound)

	options := ast.ConvertOptions{}.WithContext(ctx)
	require.True(t, options.SystemVariables == bound.System)

	explicit := sqle.NewEngineVariables()
	options = ast.ConvertOptions{SystemVariables: explicit.System}.WithContext(ctx)
	require.True(t, options.SystemVariables == explicit.System, "an explicit registry wins over the context")

	require.Nil(t, ast.ConvertOptions{}.WithContext(sql.NewEmptyContext()).SystemVariables)
	require.Nil(t, ast.ConvertOptions{}.WithContext(context.Background()).SystemVariables)
	require.Nil(t, ast.ConvertOptions{}.WithContext(nil).SystemVariables)
}
