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

package types

import (
	"github.com/dolthub/go-mysql-server/sql/types"
)

// TypeCodec is a storage codec for extension types, replacing the type's send and receive functions for storage.
type TypeCodec struct {
	Serialize         func(val any) ([]byte, error)
	Deserialize       func(data []byte) (any, error)
	SerializedCompare func(left []byte, right []byte) (int, error)
	VectorIndexable   bool // VectorIndexable is true when Deserialize yields a value that sql.ConvertToVector accepts.
}

// LoadTypeCodec returns the storage codec registered under the given name, or nil when the type serializes through its
// send and receive functions. This is used to avoid a dependency cycle.
var LoadTypeCodec func(sendFuncName string) *TypeCodec

var _ types.VectorIndexableType = &DoltgresType{}

// codec returns the storage codec registered for this type, or nil when the type has none.
func (t *DoltgresType) codec() *TypeCodec {
	if LoadTypeCodec == nil {
		return nil
	}
	return LoadTypeCodec(globalFunctionRegistry.GetString(t.SendFunc))
}

// VectorIndexable implements the types.VectorIndexableType interface.
func (t *DoltgresType) VectorIndexable() bool {
	codec := t.codec()
	return codec != nil && codec.VectorIndexable
}
