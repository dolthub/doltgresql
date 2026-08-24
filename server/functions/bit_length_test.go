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
	"math"
	"testing"

	"github.com/stretchr/testify/require"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

func TestBitStringLength(t *testing.T) {
	result, err := bitStringLength(math.MaxInt32)
	require.NoError(t, err)
	require.Equal(t, int32(math.MaxInt32), result)

	result, err = bitStringLength(math.MaxInt32 + 1)
	require.Zero(t, result)
	require.Error(t, err)
	require.True(t, pgtypes.ErrOutOfRange.Is(err))
	require.EqualError(t, err, "integer out of range")
}
