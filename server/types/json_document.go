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
	"strings"

	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/utils"
)

// JsonValueType represents the type of a JSON value. These values are serialized, and therefore should never be modified.
type JsonValueType byte

const (
	JsonValueType_Object  JsonValueType = 0
	JsonValueType_Array   JsonValueType = 1
	JsonValueType_String  JsonValueType = 2
	JsonValueType_Number  JsonValueType = 3
	JsonValueType_Boolean JsonValueType = 4
	JsonValueType_Null    JsonValueType = 5
)

// JsonDocument represents an entire JSON document.
type JsonDocument struct {
	Value JsonValue
}

// JsonValue is a value that represents some kind of data in JSON.
type JsonValue interface {
	// enforceJsonInterfaceInheritance is a special function that ensures only the expected types inherit this interface.
	enforceJsonInterfaceInheritance(error)
}

// JsonValueObject represents a JSON object.
type JsonValueObject struct {
	Items []JsonValueObjectItem
	Index map[string]int
}

// JsonValueObjectItem represents a specific item inside a JsonObject.
type JsonValueObjectItem struct {
	Key   string
	Value JsonValue
}

// JsonValueArray represents a JSON array.
type JsonValueArray []JsonValue

// JsonValueString represents a string value.
type JsonValueString string

// JsonValueNumber represents a number.
type JsonValueNumber decimal.Decimal

// JsonValueBoolean represents a boolean value.
type JsonValueBoolean bool

// JsonValueNull represents a null value.
type JsonValueNull byte

var _ JsonValue = JsonValueObject{}
var _ JsonValue = JsonValueArray{}
var _ JsonValue = JsonValueString("")
var _ JsonValue = JsonValueNumber{}
var _ JsonValue = JsonValueBoolean(false)
var _ JsonValue = JsonValueNull(0)

// enforceJsonInterfaceInheritance implements the JsonValue interface.
func (JsonValueObject) enforceJsonInterfaceInheritance(error) {}

// enforceJsonInterfaceInheritance implements the JsonValue interface.
func (JsonValueArray) enforceJsonInterfaceInheritance(error) {}

// enforceJsonInterfaceInheritance implements the JsonValue interface.
func (JsonValueString) enforceJsonInterfaceInheritance(error) {}

// enforceJsonInterfaceInheritance implements the JsonValue interface.
func (JsonValueNumber) enforceJsonInterfaceInheritance(error) {}

// enforceJsonInterfaceInheritance implements the JsonValue interface.
func (JsonValueBoolean) enforceJsonInterfaceInheritance(error) {}

// enforceJsonInterfaceInheritance implements the JsonValue interface.
func (JsonValueNull) enforceJsonInterfaceInheritance(error) {}

// JsonValueCopy returns a new copy of the given JsonValue that may be freely modified.
func JsonValueCopy(value JsonValue) JsonValue {
	switch value := value.(type) {
	case JsonValueObject:
		newItems := make([]JsonValueObjectItem, len(value.Items))
		newIndex := make(map[string]int)
		for i := range value.Items {
			newItems[i].Key = value.Items[i].Key
			newItems[i].Value = JsonValueCopy(value.Items[i].Value)
			newIndex[newItems[i].Key] = i
		}
		return JsonValueObject{
			Items: newItems,
			Index: newIndex,
		}
	case JsonValueArray:
		newArray := make(JsonValueArray, len(value))
		for i := range value {
			newArray[i] = JsonValueCopy(value[i])
		}
		return newArray
	default:
		return value
	}
}

// jsonValueSerialize is the recursive serializer for JSON values.
func jsonValueSerialize(writer *utils.Writer, value JsonValue) {
	switch value := value.(type) {
	case JsonValueObject:
		writer.Byte(byte(JsonValueType_Object))
		writer.VariableUint(uint64(len(value.Items)))
		for _, item := range value.Items {
			writer.String(item.Key)
			jsonValueSerialize(writer, item.Value)
		}
	case JsonValueArray:
		writer.Byte(byte(JsonValueType_Array))
		writer.VariableUint(uint64(len(value)))
		for _, item := range value {
			jsonValueSerialize(writer, item)
		}
	case JsonValueString:
		writer.Byte(byte(JsonValueType_String))
		writer.String(string(value))
	case JsonValueNumber:
		writer.Byte(byte(JsonValueType_Number))
		// MarshalBinary cannot error, so we can safely ignore it
		bytes, _ := decimal.Decimal(value).MarshalBinary()
		writer.ByteSlice(bytes)
	case JsonValueBoolean:
		writer.Byte(byte(JsonValueType_Boolean))
		writer.Bool(bool(value))
	case JsonValueNull:
		writer.Byte(byte(JsonValueType_Null))
	}
}

// jsonValueDeserialize is the recursive deserializer for JSON values.
func jsonValueDeserialize(reader *utils.Reader) (_ JsonValue, err error) {
	switch JsonValueType(reader.Byte()) {
	case JsonValueType_Object:
		items := make([]JsonValueObjectItem, reader.VariableUint())
		index := make(map[string]int)
		for i := range items {
			items[i].Key = reader.String()
			items[i].Value, err = jsonValueDeserialize(reader)
			if err != nil {
				return nil, err
			}
			index[items[i].Key] = i
		}
		return JsonValueObject{
			Items: items,
			Index: index,
		}, nil
	case JsonValueType_Array:
		values := make(JsonValueArray, reader.VariableUint())
		for i := range values {
			values[i], err = jsonValueDeserialize(reader)
			if err != nil {
				return nil, err
			}
		}
		return values, nil
	case JsonValueType_String:
		return JsonValueString(reader.String()), nil
	case JsonValueType_Number:
		d := decimal.Decimal{}
		err = d.UnmarshalBinary(reader.ByteSlice())
		return JsonValueNumber(d), err
	case JsonValueType_Boolean:
		return JsonValueBoolean(reader.Bool()), nil
	case JsonValueType_Null:
		return JsonValueNull(0), nil
	default:
		return nil, fmt.Errorf("unknown json value type")
	}
}

// jsonValueFormatter is the recursive formatter for JSON values.
func jsonValueFormatter(sb *strings.Builder, value JsonValue) {
	switch value := value.(type) {
	case JsonValueObject:
		sb.WriteRune('{')
		for i, item := range value.Items {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteRune('"')
			sb.WriteString(strings.ReplaceAll(item.Key, `"`, `\"`))
			sb.WriteString(`": `)
			jsonValueFormatter(sb, item.Value)
		}
		sb.WriteRune('}')
	case JsonValueArray:
		sb.WriteRune('[')
		for i, item := range value {
			if i > 0 {
				sb.WriteString(", ")
			}
			jsonValueFormatter(sb, item)
		}
		sb.WriteRune(']')
	case JsonValueString:
		sb.WriteRune('"')
		sb.WriteString(strings.ReplaceAll(string(value), `"`, `\"`))
		sb.WriteRune('"')
	case JsonValueNumber:
		sb.WriteString(decimal.Decimal(value).String())
	case JsonValueBoolean:
		if value {
			sb.WriteString(`true`)
		} else {
			sb.WriteString(`false`)
		}
	case JsonValueNull:
		sb.WriteString(`null`)
	}
}
