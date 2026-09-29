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

package framework

import (
	"testing"

	"github.com/stretchr/testify/require"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// TestSetReturningArrayType checks that set-returning functions retain array and scalar element types.
func TestSetReturningArrayType(t *testing.T) {
	require.Same(t, pgtypes.Int32Array, getTypeIfRowType(true, pgtypes.Int32Array))
	require.Same(t, pgtypes.Int32Array, getTypeIfRowType(true, pgtypes.RowTypeWithReturnType(pgtypes.Int32Array)))
	require.Same(t, pgtypes.Int32, getTypeIfRowType(true, pgtypes.RowTypeWithReturnType(pgtypes.Int32)))
	require.Same(t, pgtypes.TextArray, getTypeIfRowType(true, pgtypes.RowTypeWithReturnType(pgtypes.TextArray)))
	require.Same(t, pgtypes.Int32Array, getTypeIfRowType(false, pgtypes.Int32Array))
}

// TestVectorPolymorphicReturnType checks that vector arguments resolve to the expected polymorphic return types.
func TestVectorPolymorphicReturnType(t *testing.T) {
	f := &CompiledFunction{}
	require.Equal(t, pgtypes.Int16.ID, f.resolvePolymorphicReturnType([]*pgtypes.DoltgresType{pgtypes.AnyArray}, []*pgtypes.DoltgresType{pgtypes.Int16vector}, pgtypes.AnyElement).ID)
	require.Same(t, pgtypes.Int16vector, f.resolvePolymorphicReturnType([]*pgtypes.DoltgresType{pgtypes.AnyElement}, []*pgtypes.DoltgresType{pgtypes.Int16vector}, pgtypes.AnyElement))
	require.Equal(t, pgtypes.Oid.ID, f.resolvePolymorphicReturnType([]*pgtypes.DoltgresType{pgtypes.AnyArray}, []*pgtypes.DoltgresType{pgtypes.Oidvector}, pgtypes.AnyElement).ID)
	require.Equal(t, pgtypes.Int32.ID, f.resolvePolymorphicReturnType([]*pgtypes.DoltgresType{pgtypes.AnyArray}, []*pgtypes.DoltgresType{pgtypes.Int32Array}, pgtypes.AnyElement).ID)
}
