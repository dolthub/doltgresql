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

package plpgsql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDeclareDefaultLayout pins the layout of a declaration's default within the operation, which other
// Doltgres versions read. Versions that do not compile defaults read only DeclareDefaultSourceIndex, so
// that entry must keep holding the default exactly as it was written.
func TestDeclareDefaultLayout(t *testing.T) {
	ops, err := Parse(`CREATE FUNCTION f(p INT) RETURNS TEXT AS $$
DECLARE
	literal TEXT := 'it''s';
	param INT := p;
	expression TEXT[] := ARRAY['retired_at', 'deleted_at'];
	referencing TEXT := literal || p;
	none TEXT;
BEGIN
	RETURN literal;
END;
$$ LANGUAGE plpgsql;`)
	require.NoError(t, err)

	defaults := make(map[string][]string)
	for _, op := range ops {
		if op.OpCode == OpCode_Declare {
			defaults[op.Target] = op.SecondaryData
		}
	}
	// The source text keeps the escaped quote it was written with.
	assert.Equal(t, []string{`'it''s'`, `SELECT 'it''s' ;`}, defaults["literal"])
	assert.Equal(t, []string{`p`, `SELECT $1 ;`, `p`}, defaults["param"])
	assert.Equal(t, []string{
		`ARRAY['retired_at', 'deleted_at']`,
		`SELECT ARRAY [ 'retired_at' , 'deleted_at' ] ;`,
	}, defaults["expression"])
	assert.Equal(t, []string{`literal || p`, `SELECT $1 || $2 ;`, `literal`, `p`}, defaults["referencing"])
	// A declaration with no default carries neither form of one.
	assert.Empty(t, defaults["none"])
}

// TestDeclareDefaultLegacy covers declarations stored by a version that did not compile defaults. Those
// carry the source text alone, so the query is compiled when the declaration runs and must match what
// compiling the same default at CREATE time produces.
func TestDeclareDefaultLegacy(t *testing.T) {
	for _, test := range []struct {
		name     string
		source   string
		query    string
		bindings []string
	}{
		{
			name:   "literal",
			source: `'{A,B,C}'`,
			query:  `SELECT '{A,B,C}' ;`,
		},
		{
			// A bare parameter name was the only reference such a version recognized, and the
			// releases we test against declared the variable NULL instead of copying the value.
			name:     "parameter reference",
			source:   `p`,
			query:    `SELECT $1 ;`,
			bindings: []string{"p"},
		},
		{
			// Such a version ran this through the declared type's input function and failed, so there
			// is no earlier behavior to preserve.
			name:   "expression",
			source: `ARRAY['retired_at']`,
			query:  `SELECT ARRAY [ 'retired_at' ] ;`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			// One form answers a query that binds nothing with an empty slice and the other with
			// nil; both say the same thing.
			assertBindings := func(expected []string, actual []string) {
				if len(expected) == 0 {
					assert.Empty(t, actual)
				} else {
					assert.Equal(t, expected, actual)
				}
			}
			stack := NewInterpreterStack(nil)
			stack.NewVariableWithValue("p", nil, nil)

			legacy := InterpreterOperation{
				OpCode:        OpCode_Declare,
				PrimaryData:   "text",
				SecondaryData: []string{test.source},
				Target:        "v",
			}
			query, bindings, err := declareDefault(legacy, &stack)
			require.NoError(t, err)
			assert.Equal(t, test.query, query)
			assertBindings(test.bindings, bindings)

			// Both forms are stored side by side and either may be the one read, so compiling at
			// CREATE time has to agree.
			compiledQuery, compiledBindings, err := compileDeclareDefault(test.source, &stack)
			require.NoError(t, err)
			assert.Equal(t, query, compiledQuery)
			assertBindings(bindings, compiledBindings)

			// A declaration that stored its query is answered from that instead.
			current := legacy
			current.SecondaryData = append([]string{test.source, compiledQuery}, compiledBindings...)
			query, bindings, err = declareDefault(current, &stack)
			require.NoError(t, err)
			assert.Equal(t, test.query, query)
			assertBindings(test.bindings, bindings)
		})
	}
}
