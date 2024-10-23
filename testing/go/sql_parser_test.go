// Copyright 2024 Dolthub, Inc.
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
	"context"
	"testing"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/postgres/parser/parser/sql"
)

// TestParserBehaviors tests behaviors, such as empty statement handling, for the Doltgres parser.
func TestParserBehaviors(t *testing.T) {
	parser := sql.NewPostgresParser()

	// Parser implementations should return Vitess' ErrEmpty error for empty statements.
	t.Run("empty statement parsing", func(t *testing.T) {
		emptyStatements := []string{"", " ", "\t", ";", "-- comment", "/* comment */"}
		for _, statement := range emptyStatements {
			parsed, err := parser.ParseSimple(statement)
			require.Nil(t, parsed)
			require.ErrorIs(t, err, vitess.ErrEmpty)

			parsed, _, _, err = parser.Parse(nil, statement, true)
			require.Nil(t, parsed)
			require.ErrorIs(t, err, vitess.ErrEmpty)

			parsed, _, err = parser.ParseOneWithOptions(context.Background(), statement, vitess.ParserOptions{})
			require.Nil(t, parsed)
			require.ErrorIs(t, err, vitess.ErrEmpty)

			parsed, _, _, err = parser.ParseWithOptions(context.Background(), statement, ';', false, vitess.ParserOptions{})
			require.Nil(t, parsed)
			require.ErrorIs(t, err, vitess.ErrEmpty)
		}
	})
}
