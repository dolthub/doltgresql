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
	"bytes"
	"fmt"
	"math"
	"reflect"
	"sort"
	"strings"
	"unsafe"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/goccy/go-json"
	"github.com/lib/pq/oid"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/utils"
)

// JsonB is the deserialized and structured version of JSON that deals with JsonDocument.
var JsonB = JsonBType{}

// JsonBType is the extended type implementation of the PostgreSQL jsonb.
type JsonBType struct{}

var _ DoltgresType = JsonBType{}

// Alignment implements the DoltgresType interface.
func (b JsonBType) Alignment() TypeAlignment {
	return TypeAlignment_Int
}

// BaseID implements the DoltgresType interface.
func (b JsonBType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_JsonB
}

// BaseName implements the DoltgresType interface.
func (b JsonBType) BaseName() string {
	return "jsonb"
}

// Category implements the DoltgresType interface.
func (b JsonBType) Category() TypeCategory {
	return TypeCategory_UserDefinedTypes
}

// CollationCoercibility implements the DoltgresType interface.
func (b JsonBType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b JsonBType) Compare(v1 any, v2 any) (int, error) {
	if v1 == nil && v2 == nil {
		return 0, nil
	} else if v1 != nil && v2 == nil {
		return 1, nil
	} else if v1 == nil && v2 != nil {
		return -1, nil
	}

	ac, _, err := b.Convert(v1)
	if err != nil {
		return 0, err
	}
	bc, _, err := b.Convert(v2)
	if err != nil {
		return 0, err
	}
	ab := ac.(JsonDocument)
	bb := bc.(JsonDocument)

	return jsonValueCompare(ab.Value, bb.Value), nil
}

// Convert implements the DoltgresType interface.
func (b JsonBType) Convert(val any) (any, sql.ConvertInRange, error) {
	switch val := val.(type) {
	case JsonDocument:
		return val, sql.InRange, nil
	case nil:
		return nil, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", b.String(), val)
	}
}

// Equals implements the DoltgresType interface.
func (b JsonBType) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatValue implements the DoltgresType interface.
func (b JsonBType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return b.IoOutput(sql.NewEmptyContext(), val)
}

// GetSerializationID implements the DoltgresType interface.
func (b JsonBType) GetSerializationID() SerializationID {
	return SerializationID_JsonB
}

// IoInput implements the DoltgresType interface.
func (b JsonBType) IoInput(ctx *sql.Context, input string) (any, error) {
	inputBytes := unsafe.Slice(unsafe.StringData(input), len(input))
	if json.Valid(inputBytes) {
		doc, err := b.unmarshalToJsonDocument(inputBytes)
		return doc, err
	}
	return nil, fmt.Errorf("invalid input syntax for type json")
}

// IoOutput implements the DoltgresType interface.
func (b JsonBType) IoOutput(ctx *sql.Context, output any) (string, error) {
	converted, _, err := b.Convert(output)
	if err != nil {
		return "", err
	}
	sb := strings.Builder{}
	sb.Grow(256)
	jsonValueFormatter(&sb, converted.(JsonDocument).Value)
	return sb.String(), nil
}

// IsPreferredType implements the DoltgresType interface.
func (b JsonBType) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresType interface.
func (b JsonBType) IsUnbounded() bool {
	return true
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b JsonBType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_Unbounded
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b JsonBType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return math.MaxUint32
}

// OID implements the DoltgresType interface.
func (b JsonBType) OID() uint32 {
	return uint32(oid.T_jsonb)
}

// Promote implements the DoltgresType interface.
func (b JsonBType) Promote() sql.Type {
	return b
}

// SerializedCompare implements the DoltgresType interface.
func (b JsonBType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	if len(v1) == 0 && len(v2) == 0 {
		return 0, nil
	} else if len(v1) > 0 && len(v2) == 0 {
		return 1, nil
	} else if len(v1) == 0 && len(v2) > 0 {
		return -1, nil
	}

	v1Doc, err := b.DeserializeValue(v1)
	if err != nil {
		return 0, err
	}
	v2Doc, err := b.DeserializeValue(v2)
	if err != nil {
		return 0, err
	}
	return b.Compare(v1Doc, v2Doc)
}

// SQL implements the DoltgresType interface.
func (b JsonBType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	if v == nil {
		return sqltypes.NULL, nil
	}
	value, err := b.IoOutput(ctx, v)
	if err != nil {
		return sqltypes.Value{}, err
	}
	return sqltypes.MakeTrusted(sqltypes.Text, types.AppendAndSliceBytes(dest, []byte(value))), nil
}

// String implements the DoltgresType interface.
func (b JsonBType) String() string {
	return "jsonb"
}

// ToArrayType implements the DoltgresType interface.
func (b JsonBType) ToArrayType() DoltgresArrayType {
	return JsonBArray
}

// Type implements the DoltgresType interface.
func (b JsonBType) Type() query.Type {
	return sqltypes.Text
}

// ValueType implements the DoltgresType interface.
func (b JsonBType) ValueType() reflect.Type {
	return reflect.TypeOf(JsonDocument{})
}

// Zero implements the DoltgresType interface.
func (b JsonBType) Zero() any {
	return JsonDocument{Value: JsonValueNull(0)}
}

// SerializeType implements the DoltgresType interface.
func (b JsonBType) SerializeType() ([]byte, error) {
	return SerializationID_JsonB.ToByteSlice(0), nil
}

// deserializeType implements the DoltgresType interface.
func (b JsonBType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	switch version {
	case 0:
		return JsonB, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, b.String())
	}
}

// SerializeValue implements the DoltgresType interface.
func (b JsonBType) SerializeValue(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return nil, err
	}
	writer := utils.NewWriter(256)
	jsonValueSerialize(writer, converted.(JsonDocument).Value)
	return writer.Data(), nil
}

// DeserializeValue implements the DoltgresType interface.
func (b JsonBType) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	reader := utils.NewReader(val)
	jsonValue, err := jsonValueDeserialize(reader)
	return JsonDocument{Value: jsonValue}, err
}

// unmarshalToJsonDocument converts a JSON document byte slice into the actual JSON document.
func (b JsonBType) unmarshalToJsonDocument(val []byte) (JsonDocument, error) {
	var decoded interface{}
	if err := json.Unmarshal(val, &decoded); err != nil {
		return JsonDocument{}, err
	}
	jsonValue, err := b.ConvertToJsonDocument(decoded)
	if err != nil {
		return JsonDocument{}, err
	}
	return JsonDocument{Value: jsonValue}, nil
}

// ConvertToJsonDocument recursively constructs a valid JsonDocument based on the structures returned by the decoder.
func (b JsonBType) ConvertToJsonDocument(val interface{}) (JsonValue, error) {
	var err error
	switch val := val.(type) {
	case map[string]interface{}:
		keys := utils.GetMapKeys(val)
		sort.Slice(keys, func(i, j int) bool {
			// Key length is sorted before key contents
			if len(keys[i]) < len(keys[j]) {
				return true
			} else if len(keys[i]) > len(keys[j]) {
				return false
			} else {
				return keys[i] < keys[j]
			}
		})
		items := make([]JsonValueObjectItem, len(val))
		index := make(map[string]int)
		for i, key := range keys {
			items[i].Key = key
			items[i].Value, err = b.ConvertToJsonDocument(val[key])
			if err != nil {
				return nil, err
			}
			index[key] = i
		}
		return JsonValueObject{
			Items: items,
			Index: index,
		}, nil
	case []interface{}:
		values := make(JsonValueArray, len(val))
		for i, item := range val {
			values[i], err = b.ConvertToJsonDocument(item)
			if err != nil {
				return nil, err
			}
		}
		return values, nil
	case string:
		return JsonValueString(val), nil
	case float64:
		// TODO: handle this as a proper numeric as float64 is not precise enough
		return JsonValueNumber(decimal.NewFromFloat(val)), nil
	case bool:
		return JsonValueBoolean(val), nil
	case nil:
		return JsonValueNull(0), nil
	default:
		return nil, fmt.Errorf("unexpected type while constructing JsonDocument: %T", val)
	}
}
