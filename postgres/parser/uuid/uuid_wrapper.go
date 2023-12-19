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

// Copyright 2016 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package uuid

import (
	"crypto/rand"
	"encoding/json"
	"fmt"

	"github.com/dolthub/doltgresql/postgres/parser/utils"
)

// Short returns the first eight characters of the output of String().
func (u UUID) Short() string {
	return u.String()[:8]
}

// ShortStringer implements fmt.Stringer to output Short() on String().
type ShortStringer UUID

// String is part of fmt.Stringer.
func (s ShortStringer) String() string {
	return UUID(s).Short()
}

var _ fmt.Stringer = ShortStringer{}

// Equal returns true iff the receiver equals the argument.
//
// This method exists only to conform to the API expected by gogoproto's
// generated Equal implementations.
func (u UUID) Equal(t UUID) bool {
	return u == t
}

// GetBytes returns the UUID as a byte slice. It incurs an allocation if
// the return value escapes.
func (u UUID) GetBytes() []byte {
	return u.bytes()
}

// GetBytesMut returns the UUID as a mutable byte slice. Unlike GetBytes,
// it does not necessarily incur an allocation if the return value escapes.
// Instead, the return value escaping will cause the method's receiver (and
// any struct that it is a part of) to escape. Use only if GetBytes is causing
// an allocation and the UUID is already on the heap.
func (u *UUID) GetBytesMut() []byte {
	return u.bytesMut()
}

// ToUint128 returns the UUID as a Uint128.
func (u UUID) ToUint128() utils.Uint128 {
	return utils.FromBytes(u.bytes())
}

// Size returns the marshaled size of u, in bytes.
func (u UUID) Size() int {
	return len(u)
}

// MarshalTo marshals u to data.
func (u UUID) MarshalTo(data []byte) (int, error) {
	return copy(data, u.GetBytes()), nil
}

// Unmarshal unmarshals data to u.
func (u *UUID) Unmarshal(data []byte) error {
	return u.UnmarshalBinary(data)
}

// MarshalJSON returns the JSON encoding of u.
func (u UUID) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.String())
}

// UnmarshalJSON unmarshals the JSON encoded data into u.
func (u *UUID) UnmarshalJSON(data []byte) error {
	var uuidString string
	if err := json.Unmarshal(data, &uuidString); err != nil {
		return err
	}
	uuid, err := FromString(uuidString)
	*u = uuid
	return err
}

// defaultRandReader is an io.Reader that calls through to "math/rand".Read
// which is safe for concurrent use.
type defaultRandReader struct{}

func (r defaultRandReader) Read(p []byte) (n int, err error) {
	return rand.Read(p)
}
