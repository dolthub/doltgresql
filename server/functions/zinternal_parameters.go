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

package functions

import (
	"fmt"
	"reflect"
	"time"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
)

// ParameterType represents the type of a parameter.
type ParameterType uint8

const (
	ParameterType_Null      ParameterType = iota // The parameter is a NULL value, and is therefore typeless
	ParameterType_Integer                        // The parameter is an IntegerType type
	ParameterType_Float                          // The parameter is a FloatType type
	ParameterType_Numeric                        // The parameter is a NumericType type
	ParameterType_String                         // The parameter is a StringType type
	ParameterType_Timestamp                      // The parameter is a TimestampType type

	ParameterType_Length // The number of parameters. This should always be last in the enum declaration.
)

// ptCasts contains an array of all potential casts for each parameter type
var ptCasts [ParameterType_Length][]ParameterType

func init() {
	ptCasts[ParameterType_Null] = []ParameterType{ParameterType_Integer, ParameterType_Float, ParameterType_Numeric, ParameterType_String, ParameterType_Timestamp}
	ptCasts[ParameterType_Integer] = []ParameterType{ParameterType_Float, ParameterType_Numeric, ParameterType_String}
	ptCasts[ParameterType_Float] = []ParameterType{ParameterType_Numeric, ParameterType_String}
	ptCasts[ParameterType_Numeric] = []ParameterType{ParameterType_String}
	ptCasts[ParameterType_String] = []ParameterType{ParameterType_Integer, ParameterType_Float, ParameterType_Numeric, ParameterType_Timestamp}
	ptCasts[ParameterType_Timestamp] = []ParameterType{}
}

// PotentialCasts returns all potential casts for the current type. For example, an IntegerType may be cast to a FloatType.
// Casts may be bidirectional, as a StringType may cast to an IntegerType, and an IntegerType may cast to a StringType.
func (pt ParameterType) PotentialCasts() []ParameterType {
	return ptCasts[pt]
}

// PotentialCasts returns all potential casts for the current type. For example, an IntegerType may be cast to a FloatType.
// Casts may be bidirectional, as a StringType may cast to an IntegerType, and an IntegerType may cast to a StringType.
func (pt ParameterType) String() string {
	switch pt {
	case ParameterType_Null:
		return "null"
	case ParameterType_Integer:
		return "integer"
	case ParameterType_Float:
		return "double precision"
	case ParameterType_Numeric:
		return "numeric"
	case ParameterType_String:
		return "character varying"
	case ParameterType_Timestamp:
		return "timestamp"
	default:
		panic(fmt.Errorf("unhandled type in ParameterType::String (%d)", int(pt)))
	}
}

// ParameterTypeFromReflection returns the ParameterType and equivalent sql.Type from the given reflection type. If the
// given type does not match a ParameterType, then this returns false.
func ParameterTypeFromReflection(t reflect.Type) (ParameterType, sql.Type, bool) {
	switch t {
	case reflect.TypeOf(IntegerType{}):
		return ParameterType_Integer, types.Int64, true
	case reflect.TypeOf(FloatType{}):
		return ParameterType_Float, types.Float64, true
	case reflect.TypeOf(NumericType{}):
		//TODO: properly handle decimal types
		return ParameterType_Numeric, types.Float64, true
	case reflect.TypeOf(StringType{}):
		return ParameterType_String, types.LongText, true
	case reflect.TypeOf(TimestampType{}):
		return ParameterType_Timestamp, types.Datetime, true
	default:
		return ParameterType_Null, types.Null, false
	}
}

// IntermediateParameter is a parameter before it has been finalized.
type IntermediateParameter struct {
	Value        interface{}
	IsNull       bool
	OriginalType ParameterType
	CurrentType  ParameterType
	Source       Source
}

// IntegerType is an integer type (all integer types are upcast to int64).
type IntegerType struct {
	Value        int64
	IsNull       bool
	OriginalType ParameterType
	Source       Source
}

// FloatType is a floating point type (float32 is upcast to float64).
type FloatType struct {
	Value        float64
	IsNull       bool
	OriginalType ParameterType
	Source       Source
}

// NumericType is a decimal type (all integer and float types are upcast to decimal).
type NumericType struct {
	Value        float64 //TODO: should be decimal, but our engine support isn't quite there yet
	IsNull       bool
	OriginalType ParameterType
	Source       Source
}

// StringType is a string type.
type StringType struct {
	Value        string
	IsNull       bool
	OriginalType ParameterType
	Source       Source
}

// TimestampType is a timestamp type.
type TimestampType struct {
	Value        time.Time
	IsNull       bool
	OriginalType ParameterType
	Source       Source
}

// ToValue converts the intermediate parameter into a concrete parameter type (IntegerType, FloatType, etc.) and returns
// it as a reflect.Value, which may be passed to the matched function.
func (ip IntermediateParameter) ToValue() reflect.Value {
	switch ip.CurrentType {
	case ParameterType_Null:
		panic(fmt.Errorf("a NULL parameter type was not erased before the call to %T::ToValue", ip))
	case ParameterType_Integer:
		return reflect.ValueOf(IntegerType{
			Value:        ip.Value.(int64),
			IsNull:       ip.IsNull,
			OriginalType: ip.OriginalType,
			Source:       ip.Source,
		})
	case ParameterType_Float:
		return reflect.ValueOf(FloatType{
			Value:        ip.Value.(float64),
			IsNull:       ip.IsNull,
			OriginalType: ip.OriginalType,
			Source:       ip.Source,
		})
	case ParameterType_Numeric:
		return reflect.ValueOf(NumericType{
			Value:        ip.Value.(float64),
			IsNull:       ip.IsNull,
			OriginalType: ip.OriginalType,
			Source:       ip.Source,
		})
	case ParameterType_String:
		return reflect.ValueOf(StringType{
			Value:        ip.Value.(string),
			IsNull:       ip.IsNull,
			OriginalType: ip.OriginalType,
			Source:       ip.Source,
		})
	case ParameterType_Timestamp:
		return reflect.ValueOf(TimestampType{
			Value:        ip.Value.(time.Time),
			IsNull:       ip.IsNull,
			OriginalType: ip.OriginalType,
			Source:       ip.Source,
		})
	default:
		panic(fmt.Errorf("unhandled type in %T::ToValue", ip))
	}
}
