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

package utils

import (
	"golang.org/x/exp/constraints"
)

// Min returns the smallest value of the given parameters.
func Min[T constraints.Ordered](vals ...T) (minimum T) {
	if len(vals) == 0 {
		return minimum
	}
	minimum = vals[0]
	for i := 1; i < len(vals); i++ {
		if vals[i] < minimum {
			minimum = vals[i]
		}
	}
	return minimum
}
