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
	"regexp"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/utils"
)

// jsonDocumentStringUnicodeRegex is used on a JsonDocument's string to find all Unicode escape sequences that have an
// additional backslash.
var jsonDocumentStringUnicodeRegex = regexp.MustCompile(`\\\\u([0-9A-Fa-f]{4})`)

// JsonValueType represents a JSON value type. These values are serialized, and therefore should never be modified.
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

// JsonValueSerialize is the recursive serializer for JSON values.
func JsonValueSerialize(writer *utils.Writer, value JsonValue) {
	switch value := value.(type) {
	case JsonValueObject:
		writer.Byte(byte(JsonValueType_Object))
		writer.VariableUint(uint64(len(value.Items)))
		for _, item := range value.Items {
			writer.String(item.Key)
			JsonValueSerialize(writer, item.Value)
		}
	case JsonValueArray:
		writer.Byte(byte(JsonValueType_Array))
		writer.VariableUint(uint64(len(value)))
		for _, item := range value {
			JsonValueSerialize(writer, item)
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

// JsonValueDeserialize is the recursive deserializer for JSON values.
func JsonValueDeserialize(reader *utils.Reader) (_ JsonValue, err error) {
	switch JsonValueType(reader.Byte()) {
	case JsonValueType_Object:
		items := make([]JsonValueObjectItem, reader.VariableUint())
		index := make(map[string]int)
		for i := range items {
			items[i].Key = reader.String()
			items[i].Value, err = JsonValueDeserialize(reader)
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
			values[i], err = JsonValueDeserialize(reader)
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
		return nil, errors.Errorf("unknown json value type")
	}
}

// ConvertToJsonDocument recursively constructs a valid JsonDocument based on the structures returned by the decoder.
func ConvertToJsonDocument(val interface{}) (JsonValue, error) {
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
			items[i].Value, err = ConvertToJsonDocument(val[key])
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
			values[i], err = ConvertToJsonDocument(item)
			if err != nil {
				return nil, err
			}
		}
		return values, nil
	case string:
		// JSON parsing will convert some escaped whitespace characters to their actual characters, which is incorrect.
		// We must retain their escaped form to be considered valid JSON.
		val = strings.ReplaceAll(val, "\\", `\\`)
		val = strings.ReplaceAll(val, "\n", `\n`)
		val = strings.ReplaceAll(val, "\t", `\t`)
		val = strings.ReplaceAll(val, "\r", `\r`)
		// We specifically don't want Unicode escape sequences to be replaced, so we revert those.
		// This is safe as we double backslashes before this step, so this will return it to its original input.
		val = jsonDocumentStringUnicodeRegex.ReplaceAllString(val, `\u$1`)
		return JsonValueString(val), nil
	case float64:
		// TODO: handle this as a proper numeric as float64 is not precise enough
		return JsonValueNumber(decimal.NewFromFloat(val)), nil
	case bool:
		return JsonValueBoolean(val), nil
	case nil:
		return JsonValueNull(0), nil
	default:
		return nil, errors.Errorf("unexpected type while constructing JsonDocument: %T", val)
	}
}
