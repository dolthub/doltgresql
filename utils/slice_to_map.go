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

// SliceToMapKeys converts the given slice into a map where all of the keys are represented by all of the slice's values.
func SliceToMapKeys[T comparable](slice []T) map[T]struct{} {
	m := make(map[T]struct{})
	for _, val := range slice {
		m[val] = struct{}{}
	}
	return m
}

// SliceToMapValues converts the given slice into a map where all of the slice's values are keyed by the output of the
// `getKey` function, which takes a value and returns a key. It is up to the caller to return a unique key.
func SliceToMapValues[K comparable, V any](slice []V, getKey func(V) K) map[K]V {
	m := make(map[K]V)
	for _, val := range slice {
		m[getKey(val)] = val
	}
	return m
}
