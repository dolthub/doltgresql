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

package expression

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	gmsexpression "github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/stretchr/testify/require"
)

func TestApplyRowUpdate(t *testing.T) {
	ctx := sql.NewEmptyContext()
	schema := sql.Schema{
		{Name: "a", Type: types.Int64},
		{Name: "b", Type: types.Int64},
		{Name: "derived", Type: types.Int64},
	}
	// The leading value is an outer-scope field, not part of the table schema.
	oldRow := sql.Row{int64(99), int64(1), int64(0), int64(0)}
	a := gmsexpression.NewGetField(1, types.Int64, "a", false)
	b := gmsexpression.NewGetField(2, types.Int64, "b", false)
	derived := gmsexpression.NewGetField(3, types.Int64, "derived", false)
	for _, tt := range []struct {
		name      string
		exprs     *sql.UpdateExprs
		expected  sql.Row
		wantError bool
	}{
		{
			name: "old explicit values and new derived values with outer scope",
			exprs: sql.NewUpdateExprs([]sql.Expression{
				gmsexpression.NewSetField(a, gmsexpression.NewLiteral(int64(2), types.Int64)),
				gmsexpression.NewSetField(b, a),
				gmsexpression.NewSetField(derived, b),
			}, 2),
			expected: sql.Row{int64(99), int64(2), int64(1), int64(1)},
		},
		{
			name: "unchanged row does not evaluate derived updates",
			exprs: sql.NewUpdateExprs([]sql.Expression{
				gmsexpression.NewSetField(a, a),
				gmsexpression.NewSetField(derived, gmsexpression.NewLiteral("invalid integer", types.Text)),
			}, 1),
			expected: oldRow,
		},
		{
			name: "conversion error after an earlier assignment preserves input",
			exprs: sql.NewUpdateExprs([]sql.Expression{
				gmsexpression.NewSetField(a, gmsexpression.NewLiteral(int64(2), types.Int64)),
				gmsexpression.NewSetField(b, gmsexpression.NewLiteral("invalid integer", types.Text)),
			}, 2),
			wantError: true,
		},
		{
			name: "derived error preserves input",
			exprs: sql.NewUpdateExprs([]sql.Expression{
				gmsexpression.NewSetField(a, gmsexpression.NewLiteral(int64(2), types.Int64)),
				gmsexpression.NewSetField(derived, gmsexpression.NewLiteral("invalid integer", types.Text)),
			}, 1),
			wantError: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			input := oldRow.Copy()
			result, err := (UpdateExpressionApplier{}).ApplyRowUpdate(ctx, tt.exprs, schema, input, false)
			if tt.wantError {
				require.Error(t, err)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.expected, result)
			}
			require.Equal(t, oldRow, input)
		})
	}
}
