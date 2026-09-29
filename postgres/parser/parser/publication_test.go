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

package parser

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

func TestPublicationRoundTrip(t *testing.T) {
	for _, query := range []string{
		`CREATE PUBLICATION p`,
		`CREATE PUBLICATION "quoted publication" FOR ALL TABLES WITH (publish = 'insert, update', publish_via_partition_root)`,
		`CREATE PUBLICATION p WITH (publish = insert, publish_via_partition_root = true)`,
		`CREATE PUBLICATION p WITH (publish_via_partition_root = 0)`,
		`DROP PUBLICATION IF EXISTS "quoted publication", p CASCADE`,
		`DROP PUBLICATION p RESTRICT`,
	} {
		t.Run(query, func(t *testing.T) {
			stmt, err := ParseOne(query)
			require.NoError(t, err)
			formatted := tree.AsStringWithFlags(stmt.AST, tree.FmtParsable)
			reparsed, err := ParseOne(formatted)
			require.NoError(t, err)
			require.Equal(t, formatted, tree.AsStringWithFlags(reparsed.AST, tree.FmtParsable))
			require.Equal(t, tree.DDL, stmt.AST.StatementType())
			if _, ok := stmt.AST.(*tree.CreatePublication); ok {
				require.Equal(t, "CREATE PUBLICATION", stmt.AST.StatementTag())
			} else {
				require.Equal(t, "DROP PUBLICATION", stmt.AST.StatementTag())
			}
		})
	}
}
