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

package sql_test

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/stretchr/testify/require"

	psql "github.com/dolthub/doltgresql/postgres/parser/parser/sql"
)

// Statements that do not touch variables must parse under a context with no engine variables.
func TestParseWithUnboundContext(t *testing.T) {
	parser := psql.NewPostgresParser()
	ctx := sql.NewEmptyContext()

	statement, _, _, err := parser.Parse(ctx, "SELECT 1", false)
	require.NoError(t, err)
	require.NotNil(t, statement)

	statement, _, err = parser.ParseOneWithOptions(ctx, "SELECT 1", vitess.ParserOptions{})
	require.NoError(t, err)
	require.NotNil(t, statement)
}
