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

package spatial

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStubReturnsUnavailable(t *testing.T) {
	g := NewGEOS()
	defer g.Close()
	point := Geometry{Type: Point, Coords: []float64{1, 2}}
	_, err := g.MakeValid(point, false, false)
	require.ErrorIs(t, err, ErrNotSupported)
	_, err = g.Intersection(point, point, 0.5)
	require.ErrorIs(t, err, ErrNotSupported)
	_, err = WagyuClip([][]int64{{0, 0, 1, 0, 1, 1, 0, 0}}, 0, 0, 1, 1)
	require.ErrorIs(t, err, ErrNotSupported)
}

func TestStubPanicsWithUnavailable(t *testing.T) {
	require.PanicsWithError(t, ErrNotSupported.Error(), func() { GEOSVersion() })
}
