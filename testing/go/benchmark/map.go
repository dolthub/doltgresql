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

package main

import (
	"cmp"
	"sort"
)

// KeyValue represents an entry in a map.
type KeyValue[K comparable, V any] struct {
	Key   K
	Value V
}

// GetMapKVs returns the map's KeyValue entries as an unsorted slice.
func GetMapKVs[K comparable, V any](m map[K]V) []KeyValue[K, V] {
	allEntries := make([]KeyValue[K, V], len(m))
	i := 0
	for k, v := range m {
		allEntries[i] = KeyValue[K, V]{Key: k, Value: v}
		i++
	}
	return allEntries
}

// GetMapKVsSorted returns the map's KeyValue entries as a sorted slice. The keys are sorted in ascending order.
func GetMapKVsSorted[K cmp.Ordered, V any](m map[K]V) []KeyValue[K, V] {
	allEntries := GetMapKVs(m)
	sort.Slice(allEntries, func(i, j int) bool {
		return allEntries[i].Key < allEntries[j].Key
	})
	return allEntries
}
