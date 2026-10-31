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

package functions

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/require"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// TestLpadRpadEmptyFill checks that an empty fill string never panics: PostgreSQL pads nothing and only truncates.
func TestLpadRpadEmptyFill(t *testing.T) {
	var types [4]*pgtypes.DoltgresType
	tests := []struct {
		name     string
		callable func(ctx *sql.Context, str string, length int32, fill string) (any, error)
		str      string
		length   int32
		fill     string
		expected string
	}{
		{"lpad shorter than length", lpadCall(&types), "hi", 5, "", "hi"},
		{"lpad longer than length", lpadCall(&types), "hello", 3, "", "hel"},
		{"lpad equal length", lpadCall(&types), "hi", 2, "", "hi"},
		{"lpad non-empty fill", lpadCall(&types), "hi", 5, "xy", "xyxhi"},
		{"lpad non-empty fill truncates", lpadCall(&types), "hello", 3, "xy", "hel"},
		{"rpad shorter than length", rpadCall(&types), "hi", 5, "", "hi"},
		{"rpad longer than length", rpadCall(&types), "hello", 3, "", "hel"},
		{"rpad non-empty fill", rpadCall(&types), "hi", 5, "xy", "hixyx"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.callable(nil, test.str, test.length, test.fill)
			require.NoError(t, err)
			require.Equal(t, test.expected, result)
		})
	}
}

func lpadCall(types *[4]*pgtypes.DoltgresType) func(*sql.Context, string, int32, string) (any, error) {
	return func(ctx *sql.Context, str string, length int32, fill string) (any, error) {
		return lpad_text_int32_text.Callable(ctx, *types, str, length, fill)
	}
}

func rpadCall(types *[4]*pgtypes.DoltgresType) func(*sql.Context, string, int32, string) (any, error) {
	return func(ctx *sql.Context, str string, length int32, fill string) (any, error) {
		return rpad_text_int32_text.Callable(ctx, *types, str, length, fill)
	}
}
