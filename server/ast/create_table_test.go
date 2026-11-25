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
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/server/ast"
)

func TestCreateTableTablespacePgDefault(t *testing.T) {
	// pg_default is PostgreSQL's default tablespace, so naming it is a no-op.
	// pgAdmin and pg_dump emit the clause on ordinary tables.
	stmts, err := parser.Parse("CREATE TABLE t (c int) TABLESPACE pg_default;")
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	_, err = ast.Convert(stmts[0])
	require.NoError(t, err)

	// A tablespace other than pg_default is still rejected.
	stmts, err = parser.Parse("CREATE TABLE t (c int) TABLESPACE fast_ssd;")
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	_, err = ast.Convert(stmts[0])
	require.Error(t, err)
}

func TestAlterTableConstraintTablespacePgDefault(t *testing.T) {
	stmts, err := parser.Parse("ALTER TABLE t ADD CONSTRAINT u UNIQUE (c) USING INDEX TABLESPACE pg_default;")
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	_, err = ast.Convert(stmts[0])
	require.NoError(t, err)

	stmts, err = parser.Parse("ALTER TABLE t ADD CONSTRAINT u UNIQUE (c) USING INDEX TABLESPACE fast_ssd;")
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	_, err = ast.Convert(stmts[0])
	require.Error(t, err)
}
