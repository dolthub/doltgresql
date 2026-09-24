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

package framework

import (
	"testing"

	"github.com/stretchr/testify/require"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

func TestSetReturningArrayType(t *testing.T) {
	require.Same(t, pgtypes.Int32Array, getTypeIfRowType(true, pgtypes.Int32Array))
	require.Same(t, pgtypes.Int32Array, getTypeIfRowType(true, pgtypes.RowTypeWithReturnType(pgtypes.Int32Array)))
	require.Same(t, pgtypes.Int32, getTypeIfRowType(true, pgtypes.RowTypeWithReturnType(pgtypes.Int32)))
}
