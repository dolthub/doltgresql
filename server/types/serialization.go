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

package types

import (
	"fmt"

	"github.com/dolthub/go-mysql-server/sql/types"
)

type serializationID byte

const (
	serializationID_Bool      serializationID = 1
	serializationID_BoolArray serializationID = 2
)

// init sets the serialization and deserialization functions.
func init() {
	types.SetExtendedTypeSerializers(SerializeType, DeserializeType)
}

// SerializeType is able to serialize the given extended type into a byte slice. All extended types will be defined
// by DoltgreSQL.
func SerializeType(extendedType types.ExtendedType) ([]byte, error) {
	switch extendedType.(type) {
	case BoolType:
		return []byte{byte(serializationID_Bool)}, nil
	case BoolArrayType:
		return []byte{byte(serializationID_BoolArray)}, nil
	default:
		return nil, fmt.Errorf("unknown type to serialize")
	}
}

// MustSerializeType internally calls SerializeType and panics on error. In general, panics should only occur when a
// type has not yet had its serialization implemented yet.
func MustSerializeType(extendedType types.ExtendedType) []byte {
	serializedType, err := SerializeType(extendedType)
	if err != nil {
		panic(err)
	}
	return serializedType
}

// DeserializeType is able to deserialize the given serialized type into an appropriate extended type. All extended
// types will be defined by DoltgreSQL.
func DeserializeType(serializedType []byte) (types.ExtendedType, error) {
	if len(serializedType) == 0 {
		return nil, fmt.Errorf("cannot deserialize an empty type")
	}
	switch serializationID(serializedType[0]) {
	case serializationID_Bool:
		return Bool, nil
	case serializationID_BoolArray:
		return BoolArray, nil
	default:
		return nil, fmt.Errorf("unknown type to deserialize")
	}
}
