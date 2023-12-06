// Copyright 2023 Dolthub, Inc.
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

package utils

import (
	"cmp"
	"sort"
)

// GetMapKeys returns the map's keys as an unsorted slice.
func GetMapKeys[K comparable, V any](m map[K]V) []K {
	allKeys := make([]K, len(m))
	i := 0
	for k := range m {
		allKeys[i] = k
		i++
	}
	return allKeys
}

// GetMapKeysSorted returns the map's keys as a sorted slice.
func GetMapKeysSorted[K cmp.Ordered, V any](m map[K]V) []K {
	allKeys := make([]K, len(m))
	i := 0
	for k := range m {
		allKeys[i] = k
		i++
	}
	sort.Slice(allKeys, func(i, j int) bool {
		return allKeys[i] < allKeys[j]
	})
	return allKeys
}

// GetMapValues returns the map's values as a slice. Due to Go's map iteration, the values will be in a
// non-deterministic order.
func GetMapValues[K comparable, V any](m map[K]V) []V {
	allValues := make([]V, len(m))
	i := 0
	for _, v := range m {
		allValues[i] = v
		i++
	}
	return allValues
}
