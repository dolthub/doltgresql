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

package expression

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/stretchr/testify/require"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

func TestCoalesceIsNullable(t *testing.T) {
	ctx := sql.NewEmptyContext()

	notNullField := expression.NewGetField(0, pgtypes.Int32, "a", false)
	nullableField := expression.NewGetField(1, pgtypes.Int32, "b", true)

	tests := []struct {
		name string
		args []sql.Expression
		want bool
	}{
		{
			name: "all arguments non-nullable",
			args: []sql.Expression{notNullField, notNullField},
			want: false,
		},
		{
			// A non-nullable argument guarantees a non-null result once reached,
			// regardless of position, so the whole expression is non-nullable.
			name: "mix of nullable and non-nullable arguments",
			args: []sql.Expression{nullableField, notNullField},
			want: false,
		},
		{
			name: "all arguments nullable",
			args: []sql.Expression{nullableField, nullableField},
			want: true,
		},
		{
			name: "single non-nullable argument",
			args: []sql.Expression{notNullField},
			want: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c, err := NewPgCoalesce(ctx, test.args...)
			require.NoError(t, err)
			require.Equal(t, test.want, c.IsNullable(ctx))
		})
	}
}
