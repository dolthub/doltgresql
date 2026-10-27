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

package _go

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/require"
	"testing"
)

// A width check alone would accept unrelated identifiers. Capture the mutation's
// result once, then verify its encoding, interface agreement and retained data.
func TestCommitHashCompatibility(t *testing.T) {
	if testTarget(t) == targetPostgres {
		t.Skip("stock PostgreSQL has no version-control interfaces")
	}
	runTargetPair(t, func(t *testing.T, postgres bool) {
		ctx, conn, controller := CreateServer(t, "postgres")
		defer func() { conn.Close(ctx); controller.Stop(); require.NoError(t, controller.WaitForStop()) }()
		_, err := conn.Exec(ctx, "CREATE TABLE hash_identity (id int PRIMARY KEY)")
		require.NoError(t, err)
		_, err = conn.Exec(ctx, "INSERT INTO hash_identity VALUES (1)")
		require.NoError(t, err)
		var hash string
		require.NoError(t, conn.Default.QueryRow(ctx, "SELECT hash FROM dolt_commit('-Am', 'hash identity')").Scan(&hash))
		if postgres {
			require.Regexp(t, "^[0-9a-f]{64}$", hash)
		} else {
			require.Regexp(t, "^[0-9a-v]{32}$", hash)
		}
		runScript(t, ctx, ScriptTest{Assertions: []ScriptTestAssertion{
			{Query: "SELECT count(*) FROM dolt_log l JOIN dolt_branches b ON b.hash=l.commit_hash WHERE b.name='main' AND l.message='hash identity' AND l.commit_hash=$1", BindVars: []any{hash}, Expected: []sql.Row{{1}}},
			{Query: "INSERT INTO hash_identity VALUES (2)"},
			{Query: "SELECT id FROM hash_identity ORDER BY id", Expected: []sql.Row{{1}, {2}}},
			{Query: "SELECT id FROM hash_identity AS OF 'HEAD' ORDER BY id", Expected: []sql.Row{{1}}},
			{Query: "SELECT count(*) FROM dolt_log l JOIN dolt_branches b ON b.hash=l.commit_hash WHERE b.name='main' AND l.message='hash identity' AND l.commit_hash=$1", BindVars: []any{hash}, Expected: []sql.Row{{1}}},
		}}, conn, true)
	})
}
