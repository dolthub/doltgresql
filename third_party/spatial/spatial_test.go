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
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersions(t *testing.T) {
	require.Equal(t, "3.11.1-CAPI-1.17.1", GEOSVersion())
	require.Equal(t, "9.8.1", PROJVersion())
}

func TestVersionsConcurrently(t *testing.T) {
	var wg sync.WaitGroup
	versions := make([]string, 64)
	for i := range versions {
		wg.Go(func() {
			versions[i] = PROJVersion()
		})
	}
	wg.Wait()
	for _, version := range versions {
		require.Equal(t, "9.8.1", version)
	}
}

func TestWagyuClip(t *testing.T) {
	polygons, err := WagyuClip([][]int64{{-5, 15, 5, 15, 5, 5, -5, 5, -5, 15}}, 0, 0, 10, 10)
	require.NoError(t, err)
	require.Equal(t, [][][]int64{{{0, 10, 0, 5, 5, 5, 5, 10, 0, 10}}}, polygons)
	polygons, err = WagyuClip([][]int64{{0, 10, 10, 10, 10, 0, 0, 0, 0, 10}, {2, 8, 2, 2, 8, 2, 8, 8, 2, 8}}, 0, 0, 10, 10)
	require.NoError(t, err)
	require.Equal(t, [][][]int64{{{0, 10, 0, 0, 10, 0, 10, 10, 0, 10}, {8, 2, 2, 2, 2, 8, 8, 8, 8, 2}}}, polygons)
}
