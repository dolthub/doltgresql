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

package pgcatalog

import (
	"hash/fnv"
	"math"
	"strings"
)

// hashStringToUint32 hashes a string to a uint32 using the FNV-1a hash function.
func hashStringToUint32(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	hash := h.Sum32()
	// Ensure the hash is in the top half of the key space (>= 2^31)
	if hash < math.MaxUint32/2 {
		hash += math.MaxUint32 / 2
	}
	return hash
}

// genOid generates an OID from a list of names.
func genOid(names ...string) uint32 {
	id := strings.Join(names, ".")
	return hashStringToUint32(id)
}
