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

package types

import (
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/schema/typeinfo"
	"github.com/dolthub/dolt/go/store/val"
	"github.com/stretchr/testify/require"
)

// TestPreserveEncoding checks which encodings of an existing column are kept when the column is altered. Columns
// created before v0.56.3 use an extended encoding, which must be kept, since altering a column doesn't always rewrite
// its values.
func TestPreserveEncoding(t *testing.T) {
	varChar10, err := NewVarCharType(10)
	require.NoError(t, err)

	tests := []struct {
		name     string
		old      typeinfo.TypeInfo
		new      typeinfo.TypeInfo
		expected val.Encoding
	}{
		{
			name:     "legacy int4 stays extended",
			old:      Int32.TypeInfo().WithEncoding(val.ExtendedEnc),
			new:      Int32.TypeInfo(),
			expected: val.ExtendedEnc,
		},
		{
			name:     "legacy int4 to int8 stays extended",
			old:      Int32.TypeInfo().WithEncoding(val.ExtendedEnc),
			new:      Int64.TypeInfo(),
			expected: val.ExtendedEnc,
		},
		{
			name:     "legacy jsonb stays extended",
			old:      JsonB.TypeInfo().WithEncoding(val.ExtendedAdaptiveEnc),
			new:      JsonB.TypeInfo(),
			expected: val.ExtendedAdaptiveEnc,
		},
		{
			name:     "legacy text stays extended",
			old:      Text.TypeInfo().WithEncoding(val.ExtendedAdaptiveEnc),
			new:      Text.TypeInfo(),
			expected: val.ExtendedAdaptiveEnc,
		},
		{
			name:     "legacy varchar(10) to text uses the new encoding",
			old:      varChar10.TypeInfo().WithEncoding(val.ExtendedEnc),
			new:      Text.TypeInfo(),
			expected: val.StringAdaptiveEnc,
		},
		{
			name:     "int4 stays native",
			old:      Int32.TypeInfo(),
			new:      Int32.TypeInfo(),
			expected: val.Int32Enc,
		},
		{
			name:     "int4 to int8 uses the new encoding",
			old:      Int32.TypeInfo(),
			new:      Int64.TypeInfo(),
			expected: val.Int64Enc,
		},
		{
			name:     "timestamp stays extended",
			old:      Timestamp.TypeInfo(),
			new:      Timestamp.TypeInfo(),
			expected: val.ExtendedEnc,
		},
		{
			name:     "non-doltgres type uses the new encoding",
			old:      typeinfo.Int32Type,
			new:      Int32.TypeInfo(),
			expected: val.Int32Enc,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			preserved := typeinfo.PreserveAdaptiveEncoding(test.old, test.new)
			require.Equal(t, test.expected, preserved.Encoding())
			require.True(t, preserved.ToSqlType().Equals(test.new.ToSqlType()))
		})
	}
}
