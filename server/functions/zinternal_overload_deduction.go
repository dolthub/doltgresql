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
	"strconv"
	"time"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
)

// OverloadDeduction handles resolving which function to call by iterating over the parameter expressions. This also
// handles casting between types if an exact function match is not found.
type OverloadDeduction struct {
	Function      reflect.Value
	ReturnSqlType sql.Type
	ReturnValType ParameterType
	Parameter     [ParameterType_Length]*OverloadDeduction
}

// Resolve returns an overload that either matches the given parameters exactly, or is a viable match after casting.
// This will modify the parameter slice in-place. Returns a nil OverloadDeduction if a viable match is not found.
func (overload *OverloadDeduction) Resolve(parameters []IntermediateParameter) (*OverloadDeduction, error) {
	parameterTypes := make([]ParameterType, len(parameters))
	for i := range parameters {
		parameterTypes[i] = parameters[i].OriginalType
	}
	resultOverload, resultTypes := overload.ResolveByType(parameterTypes)
	// If we receive a nil overload, then no valid overloads were found
	if resultOverload == nil {
		return nil, nil
	}
	// If any of the result types are different from their originals, then we need to cast them to their resulting types
	// if it's possible.
	for i, t := range resultTypes {
		parameters[i].CurrentType = t
		if parameters[i].OriginalType == t {
			continue
		}

		var err error
		switch parameters[i].OriginalType {
		case ParameterType_Null:
			// Since nulls are typeless, we pretend that the current type was also the original type
			parameters[i].OriginalType = t
			switch t {
			case ParameterType_Integer:
				parameters[i].Value = int64(0)
			case ParameterType_Float:
				parameters[i].Value = float64(0)
			case ParameterType_Numeric:
				//TODO: properly handle decimal types
				parameters[i].Value = float64(0)
			case ParameterType_String:
				parameters[i].Value = ""
			case ParameterType_Timestamp:
				parameters[i].Value = time.Time{}
			default:
				return nil, fmt.Errorf("invalid `%s` cast to `%s`", parameters[i].OriginalType.String(), t.String())
			}
		case ParameterType_Integer:
			switch t {
			case ParameterType_Float:
				parameters[i].Value = float64(parameters[i].Value.(int64))
			case ParameterType_Numeric:
				//TODO: properly handle decimal types
				parameters[i].Value = float64(parameters[i].Value.(int64))
			case ParameterType_String:
				parameters[i].Value = strconv.FormatInt(parameters[i].Value.(int64), 10)
			default:
				return nil, fmt.Errorf("invalid `%s` cast to `%s`", parameters[i].OriginalType.String(), t.String())
			}
		case ParameterType_Float:
			switch t {
			case ParameterType_Numeric:
				//TODO: properly handle decimal types, this is a redundant assignment but serves as a reminder
				parameters[i].Value = parameters[i].Value.(float64)
			case ParameterType_String:
				parameters[i].Value = strconv.FormatFloat(parameters[i].Value.(float64), 'f', -1, 64)
			default:
				return nil, fmt.Errorf("invalid `%s` cast to `%s`", parameters[i].OriginalType.String(), t.String())
			}
		case ParameterType_Numeric:
			switch t {
			case ParameterType_String:
				//TODO: properly handle decimal types
				parameters[i].Value = strconv.FormatFloat(parameters[i].Value.(float64), 'f', -1, 64)
			default:
				return nil, fmt.Errorf("invalid `%s` cast to `%s`", parameters[i].OriginalType.String(), t.String())
			}
		case ParameterType_String:
			switch t {
			case ParameterType_Integer:
				parameters[i].Value, err = strconv.ParseInt(parameters[i].Value.(string), 10, 64)
				if err != nil {
					return nil, fmt.Errorf("cannot cast `%s` to type `%s`", parameters[i].Value.(string), t.String())
				}
				// It looks like string constants are treated as native integer types, so we'll mimic this here
				if parameters[i].Source == Source_Constant {
					parameters[i].OriginalType = ParameterType_Integer
				}
			case ParameterType_Float:
				parameters[i].Value, err = strconv.ParseFloat(parameters[i].Value.(string), 64)
				if err != nil {
					return nil, fmt.Errorf("cannot cast `%s` to type `%s`", parameters[i].Value.(string), t.String())
				}
				// It looks like string constants are treated as native float types, so we'll mimic this here
				if parameters[i].Source == Source_Constant {
					parameters[i].OriginalType = ParameterType_Float
				}
			case ParameterType_Numeric:
				//TODO: properly handle decimal types
				parameters[i].Value, err = strconv.ParseFloat(parameters[i].Value.(string), 64)
				if err != nil {
					return nil, fmt.Errorf("cannot cast `%s` to type `%s`", parameters[i].Value.(string), t.String())
				}
				// It looks like string constants are treated as native numeric types, so we'll mimic this here
				if parameters[i].Source == Source_Constant {
					parameters[i].OriginalType = ParameterType_Numeric
				}
			case ParameterType_Timestamp:
				//TODO: properly handle timestamps
				parameters[i].Value, _, err = types.Datetime.Convert(parameters[i].Value)
				if err != nil {
					return nil, fmt.Errorf("cannot cast `%s` to type `%s`", parameters[i].Value.(string), t.String())
				}
			default:
				return nil, fmt.Errorf("invalid `%s` cast to `%s`", parameters[i].OriginalType.String(), t.String())
			}
		case ParameterType_Timestamp:
			return nil, fmt.Errorf("invalid `%s` cast to `%s`", parameters[i].OriginalType.String(), t.String())
		default:
			return nil, fmt.Errorf("unhandled parameter type in %T::Resolve", overload)
		}
	}
	return resultOverload, nil
}

// ResolveByType returns the best matching overload for the given types. The returned types represent the actual types
// used by the overload, which may differ from the calling types. It is up to the caller to cast the parameters to match
// the types expected by the returned overload. Returns a nil OverloadDeduction if a viable match is not found.
func (overload *OverloadDeduction) ResolveByType(originalTypes []ParameterType) (*OverloadDeduction, []ParameterType) {
	resultTypes := make([]ParameterType, len(originalTypes))
	copy(resultTypes, originalTypes)
	return overload.resolveByType(originalTypes, resultTypes), resultTypes
}

// resolveByType is the recursive implementation of ResolveByType.
func (overload *OverloadDeduction) resolveByType(originalTypes []ParameterType, resultTypes []ParameterType) *OverloadDeduction {
	if overload == nil {
		return nil
	}
	if len(originalTypes) == 0 {
		if overload.Function.IsValid() && !overload.Function.IsNil() {
			return overload
		}
		return nil
	}

	// Check if we're able to resolve the original type
	t := originalTypes[0]
	resultOverload := overload.Parameter[t].resolveByType(originalTypes[1:], resultTypes[1:])
	if resultOverload != nil {
		resultTypes[0] = t
		return resultOverload
	}

	// We did not find a resolution for the original type, so we'll look through each cast
	for _, cast := range t.PotentialCasts() {
		resultOverload = overload.Parameter[cast].resolveByType(originalTypes[1:], resultTypes[1:])
		if resultOverload != nil {
			resultTypes[0] = cast
			return resultOverload
		}
	}
	// We did not find any potential matches, so we'll return nil
	return nil
}
