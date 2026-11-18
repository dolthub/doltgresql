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

//go:build !postgis_native

package spatial

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStubReturnsUnavailable(t *testing.T) {
	g := NewGEOS()
	defer g.Close()
	point := Geometry{Type: Point, Coords: []float64{1, 2}}
	_, err := g.Intersects(point, point)
	require.ErrorIs(t, err, ErrNativeUnavailable)
	_, err = g.Buffer(point, 1, BufferParams{QuadSegs: 8, EndCap: BufferRound, Join: BufferRound})
	require.ErrorIs(t, err, ErrNativeUnavailable)
	_, err = NewTransformer("EPSG:4326", "EPSG:3857")
	require.ErrorIs(t, err, ErrNativeUnavailable)
	_, err = DatabaseCRSInfo("EPSG", nil)
	require.ErrorIs(t, err, ErrNativeUnavailable)
	_, err = WagyuClip([][]int64{{0, 0, 1, 0, 1, 1, 0, 0}}, 0, 0, 1, 1)
	require.ErrorIs(t, err, ErrNativeUnavailable)
}

func TestStubPanicsWithUnavailable(t *testing.T) {
	require.PanicsWithError(t, ErrNativeUnavailable.Error(), func() { GEOSVersion() })
}

func TestStubReadsNothing(t *testing.T) {
	require.False(t, Reads("EPSG:4326"))
	require.False(t, IsGeographic("EPSG:4326"))
}
