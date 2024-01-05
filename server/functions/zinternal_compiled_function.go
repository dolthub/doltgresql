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
	"strings"
	"time"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/shopspring/decimal"
)

// CompiledFunction is an expression that represents a fully-analyzed PostgreSQL function.
type CompiledFunction struct {
	Name       string
	Parameters []sql.Expression
	Functions  *OverloadDeduction
}

var _ sql.FunctionExpression = (*CompiledFunction)(nil)

// FunctionName implements the interface sql.Expression.
func (c *CompiledFunction) FunctionName() string {
	return c.Name
}

// Description implements the interface sql.Expression.
func (c *CompiledFunction) Description() string {
	return fmt.Sprintf("The PostgreSQL function `%s`", c.Name)
}

// Resolved implements the interface sql.Expression.
func (c *CompiledFunction) Resolved() bool {
	for _, param := range c.Parameters {
		if !param.Resolved() {
			return false
		}
	}
	return true
}

// String implements the interface sql.Expression.
func (c *CompiledFunction) String() string {
	sb := strings.Builder{}
	sb.WriteString(c.Name + "(")
	for i, param := range c.Parameters {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(param.String())
	}
	sb.WriteString(")")
	return sb.String()
}

// OverloadString returns the name of the function represented by the given overload.
func (c *CompiledFunction) OverloadString(types []IntermediateParameter) string {
	sb := strings.Builder{}
	sb.WriteString(c.Name + "(")
	for i, t := range types {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(t.CurrentType.String())
	}
	sb.WriteString(")")
	return sb.String()
}

// Type implements the interface sql.Expression.
func (c *CompiledFunction) Type() sql.Type {
	if resolvedFunction, _ := c.Functions.ResolveByType(c.possibleParameterTypes()); resolvedFunction != nil {
		return resolvedFunction.ReturnSqlType
	}
	// We can't resolve to a function before evaluation in this case, so we'll return something arbitrary
	return types.LongText
}

// IsNullable implements the interface sql.Expression.
func (c *CompiledFunction) IsNullable() bool {
	// We'll always return true, since it does not seem to have a truly negative impact if we return true for a function
	// that will never return NULL, however there is a negative impact for returning false when a function does return
	// NULL.
	return true
}

// Eval implements the interface sql.Expression.
func (c *CompiledFunction) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	// First we'll evaluate all of the parameters.
	parameters, err := c.evalParameters(ctx, row)
	if err != nil {
		return nil, err
	}
	// Next we'll resolve the overload based on the parameters given.
	overload, err := c.Functions.Resolve(parameters)
	if err != nil {
		return nil, err
	}
	// If we do not receive an overload, then the parameters given did not result in a valid match
	if overload == nil {
		return nil, fmt.Errorf("function %s does not exist", c.OverloadString(parameters))
	}
	// Convert the intermediate parameters into their concrete types, then pass them to the function
	concreteParameters := make([]reflect.Value, len(parameters))
	for i := range parameters {
		concreteParameters[i] = parameters[i].ToValue()
	}
	result := overload.Function.Call(concreteParameters)
	if !result[1].IsNil() {
		return nil, result[1].Interface().(error)
	}
	// Unpack the resulting value, returning it to the caller
	switch overload.ReturnValType {
	case ParameterType_Integer:
		resultVal := result[0].Interface().(IntegerType)
		if resultVal.IsNull {
			return nil, nil
		}
		return resultVal.Value, nil
	case ParameterType_Float:
		resultVal := result[0].Interface().(FloatType)
		if resultVal.IsNull {
			return nil, nil
		}
		return resultVal.Value, nil
	case ParameterType_Numeric:
		resultVal := result[0].Interface().(NumericType)
		if resultVal.IsNull {
			return nil, nil
		}
		return resultVal.Value, nil
	case ParameterType_String:
		resultVal := result[0].Interface().(StringType)
		if resultVal.IsNull {
			return nil, nil
		}
		return resultVal.Value, nil
	case ParameterType_Timestamp:
		resultVal := result[0].Interface().(TimestampType)
		if resultVal.IsNull {
			return nil, nil
		}
		return resultVal.Value, nil
	default:
		return nil, fmt.Errorf("unhandled parameter type in %T::Eval (%d)", c, overload.ReturnValType)
	}
}

// Children implements the interface sql.Expression.
func (c *CompiledFunction) Children() []sql.Expression {
	return c.Parameters
}

// WithChildren implements the interface sql.Expression.
func (c *CompiledFunction) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	return &CompiledFunction{
		Name:       c.Name,
		Parameters: children,
		Functions:  c.Functions,
	}, nil
}

// evalParameters evaluates the parameters within an Eval call.
func (c *CompiledFunction) evalParameters(ctx *sql.Context, row sql.Row) ([]IntermediateParameter, error) {
	parameters := make([]IntermediateParameter, len(c.Parameters))
	for i, param := range c.Parameters {
		evaluatedParam, err := param.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		parameters[i].Source = c.determineSource(param)
		switch evaluatedParam := evaluatedParam.(type) {
		case int8:
			parameters[i].Value = int64(evaluatedParam)
			parameters[i].OriginalType = ParameterType_Integer
		case int16:
			parameters[i].Value = int64(evaluatedParam)
			parameters[i].OriginalType = ParameterType_Integer
		case int32:
			parameters[i].Value = int64(evaluatedParam)
			parameters[i].OriginalType = ParameterType_Integer
		case int64:
			parameters[i].Value = evaluatedParam
			parameters[i].OriginalType = ParameterType_Integer
		case float32:
			parameters[i].Value = float64(evaluatedParam)
			parameters[i].OriginalType = ParameterType_Float
		case float64:
			parameters[i].Value = evaluatedParam
			parameters[i].OriginalType = ParameterType_Float
		case decimal.Decimal:
			//TODO: properly handle decimal types
			asFloat, _ := evaluatedParam.Float64()
			parameters[i].Value = asFloat
			parameters[i].OriginalType = ParameterType_Numeric
		case string:
			parameters[i].Value = evaluatedParam
			parameters[i].OriginalType = ParameterType_String
		case time.Time:
			parameters[i].Value = evaluatedParam
			parameters[i].OriginalType = ParameterType_Timestamp
		case nil:
			parameters[i].IsNull = true
			parameters[i].OriginalType = ParameterType_Null
		default:
			return nil, fmt.Errorf("PostgreSQL functions do not yet support parameters of type `%T`", evaluatedParam)
		}
		parameters[i].CurrentType = parameters[i].OriginalType
	}
	return parameters, nil
}

// determineSource determines what the source is, based on the expression given.
func (c *CompiledFunction) determineSource(expr sql.Expression) Source {
	switch expr := expr.(type) {
	case *expression.Alias:
		return c.determineSource(expr.Child)
	case *expression.GetField:
		return Source_Column
	case *expression.Literal:
		return Source_Constant
	default:
		return Source_Expression
	}
}

// possibleParameterTypes returns the parameter types of all of the expressions by guessing the return value from the
// type that each expression declares it will return. This is not guaranteed to be correct.
func (c *CompiledFunction) possibleParameterTypes() []ParameterType {
	possibleParamTypes := make([]ParameterType, len(c.Parameters))
	for i, param := range c.Parameters {
		switch param.Type().Type() {
		case query.Type_INT8, query.Type_INT16, query.Type_INT24, query.Type_INT32, query.Type_INT64:
			possibleParamTypes[i] = ParameterType_Integer
		case query.Type_UINT8, query.Type_UINT16, query.Type_UINT24, query.Type_UINT32, query.Type_UINT64:
			possibleParamTypes[i] = ParameterType_Integer
		case query.Type_YEAR:
			possibleParamTypes[i] = ParameterType_Integer
		case query.Type_FLOAT32, query.Type_FLOAT64:
			possibleParamTypes[i] = ParameterType_Float
		case query.Type_DECIMAL:
			//TODO: properly handle decimal types
			possibleParamTypes[i] = ParameterType_Float
		case query.Type_DATE, query.Type_DATETIME, query.Type_TIMESTAMP:
			possibleParamTypes[i] = ParameterType_Timestamp
		case query.Type_CHAR, query.Type_VARCHAR, query.Type_TEXT:
			possibleParamTypes[i] = ParameterType_String
		case query.Type_ENUM, query.Type_SET:
			possibleParamTypes[i] = ParameterType_Integer
		default:
			// We'll just use NULL for now, since we've got incomplete coverage of PostgreSQL types anyway
			possibleParamTypes[i] = ParameterType_Null
		}
	}
	return possibleParamTypes
}
