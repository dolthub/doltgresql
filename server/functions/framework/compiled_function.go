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

package framework

import (
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/vitess/go/vt/proto/query"

	pgtypes "github.com/dolthub/doltgresql/server/types"
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
	return c.Functions != nil
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
func (c *CompiledFunction) OverloadString(types []pgtypes.DoltgresType) string {
	sb := strings.Builder{}
	sb.WriteString(c.Name + "(")
	for i, t := range types {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(t.String())
	}
	sb.WriteString(")")
	return sb.String()
}

// Type implements the interface sql.Expression.
func (c *CompiledFunction) Type() sql.Type {
	parameters := c.possibleParameterTypes()
	// resolveByType takes a slice of result types, so we need to create them here even though they're not used
	resultTypes := make([]pgtypes.DoltgresType, len(parameters))
	copy(resultTypes, parameters)
	sources := make([]Source, len(parameters))
	for i := range sources {
		sources[i] = Source_Constant
	}
	if resolvedFunction := c.Functions.resolveByType(parameters, resultTypes, sources); resolvedFunction != nil {
		return resolvedFunction.Function.GetReturn()
	}
	// We can't resolve to a function before evaluation in this case, so we'll return something arbitrary
	return pgtypes.VarCharMax
}

// IsNullable implements the interface sql.Expression.
func (c *CompiledFunction) IsNullable() bool {
	// All functions seem to return NULL when given a NULL value
	return true
}

// Eval implements the interface sql.Expression.
func (c *CompiledFunction) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	// First we'll analyze all of the parameters.
	originalTypes, sources, err := c.analyzeParameters()
	if err != nil {
		return nil, err
	}
	pgctx := Context{
		Context:       ctx,
		OriginalTypes: originalTypes,
		Sources:       sources,
	}
	// Next we'll resolve the overload based on the parameters given.
	overload, casts, err := c.Functions.Resolve(originalTypes, sources)
	if err != nil {
		return nil, err
	}
	// If we do not receive an overload, then the parameters given did not result in a valid match
	if overload == nil {
		return nil, fmt.Errorf("function %s does not exist", c.OverloadString(originalTypes))
	}
	// With the overload figured out, we evaluate all of the parameters.
	parameters, err := c.evalParameters(ctx, row)
	if err != nil {
		return nil, err
	}
	// Convert the parameter values into their correct types
	for i := range parameters {
		if casts[i] != nil {
			parameters[i], err = casts[i](pgctx, parameters[i])
			if err != nil {
				return nil, err
			}
		} else {
			return nil, fmt.Errorf("function %s is missing the appropriate implicit cast", c.OverloadString(originalTypes))
		}
	}
	// Pass the parameters to the function
	switch f := overload.Function.(type) {
	case Function0:
		return f.Callable(pgctx)
	case Function1:
		return f.Callable(pgctx, parameters[0])
	case Function2:
		return f.Callable(pgctx, parameters[0], parameters[1])
	case Function3:
		return f.Callable(pgctx, parameters[0], parameters[1], parameters[2])
	case Function4:
		return f.Callable(pgctx, parameters[0], parameters[1], parameters[2], parameters[3])
	default:
		return nil, fmt.Errorf("unknown function type in CompiledFunction::Eval")
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
func (c *CompiledFunction) evalParameters(ctx *sql.Context, row sql.Row) ([]any, error) {
	parameters := make([]any, len(c.Parameters))
	for i, param := range c.Parameters {
		var err error
		parameters[i], err = param.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		// TODO: once we remove GMS types from all of our expressions, we can remove this step which ensures the correct type
		if _, ok := param.Type().(pgtypes.DoltgresType); !ok {
			switch param.Type().Type() {
			case query.Type_INT8, query.Type_INT16:
				parameters[i], _, _ = pgtypes.Int16.Convert(parameters[i])
			case query.Type_INT24, query.Type_INT32:
				parameters[i], _, _ = pgtypes.Int32.Convert(parameters[i])
			case query.Type_INT64:
				parameters[i], _, _ = pgtypes.Int64.Convert(parameters[i])
			case query.Type_UINT8, query.Type_UINT16, query.Type_UINT24, query.Type_UINT32, query.Type_UINT64:
				parameters[i], _, _ = pgtypes.Int64.Convert(parameters[i])
			case query.Type_YEAR:
				parameters[i], _, _ = pgtypes.Int16.Convert(parameters[i])
			case query.Type_FLOAT32:
				parameters[i], _, _ = pgtypes.Float32.Convert(parameters[i])
			case query.Type_FLOAT64:
				parameters[i], _, _ = pgtypes.Float64.Convert(parameters[i])
			case query.Type_DECIMAL:
				parameters[i], _, _ = pgtypes.Numeric.Convert(parameters[i])
			case query.Type_DATE, query.Type_DATETIME, query.Type_TIMESTAMP:
				return nil, fmt.Errorf("need to add DoltgresType equivalents to DATETIME")
			case query.Type_CHAR, query.Type_VARCHAR:
				parameters[i], _, _ = pgtypes.VarCharMax.Convert(parameters[i])
			case query.Type_TEXT:
				parameters[i], _, _ = pgtypes.VarCharMax.Convert(parameters[i])
			case query.Type_ENUM:
				parameters[i], _, _ = pgtypes.Int16.Convert(parameters[i])
			case query.Type_SET:
				parameters[i], _, _ = pgtypes.Int64.Convert(parameters[i])
			default:
				return nil, fmt.Errorf("encountered a GMS type that cannot be handled")
			}
		}
	}
	return parameters, nil
}

// analyzeParameters analyzes the parameters within an Eval call.
func (c *CompiledFunction) analyzeParameters() (originalTypes []pgtypes.DoltgresType, sources []Source, err error) {
	// TODO: should this be within Eval or sometime before that?
	originalTypes = make([]pgtypes.DoltgresType, len(c.Parameters))
	sources = make([]Source, len(c.Parameters))
	for i, param := range c.Parameters {
		returnType := param.Type()
		if extendedType, ok := returnType.(pgtypes.DoltgresType); ok {
			originalTypes[i] = extendedType
		} else {
			// TODO: we need to remove GMS types from all of our expressions so that we can remove this
			switch param.Type().Type() {
			case query.Type_INT8, query.Type_INT16:
				originalTypes[i] = pgtypes.Int16
			case query.Type_INT24, query.Type_INT32:
				originalTypes[i] = pgtypes.Int32
			case query.Type_INT64:
				originalTypes[i] = pgtypes.Int64
			case query.Type_UINT8, query.Type_UINT16, query.Type_UINT24, query.Type_UINT32, query.Type_UINT64:
				originalTypes[i] = pgtypes.Int64
			case query.Type_YEAR:
				originalTypes[i] = pgtypes.Int16
			case query.Type_FLOAT32:
				originalTypes[i] = pgtypes.Float32
			case query.Type_FLOAT64:
				originalTypes[i] = pgtypes.Float64
			case query.Type_DECIMAL:
				originalTypes[i] = pgtypes.Numeric
			case query.Type_DATE, query.Type_DATETIME, query.Type_TIMESTAMP:
				return nil, nil, fmt.Errorf("need to add DoltgresType equivalents to DATETIME")
			case query.Type_CHAR, query.Type_VARCHAR:
				originalTypes[i] = pgtypes.VarCharMax
			case query.Type_TEXT:
				originalTypes[i] = pgtypes.VarCharMax
			case query.Type_ENUM:
				originalTypes[i] = pgtypes.Int16
			case query.Type_SET:
				originalTypes[i] = pgtypes.Int64
			default:
				return nil, nil, fmt.Errorf("encountered a type that does not conform to the DoltgresType interface")
			}
		}
		sources[i] = c.determineSource(param)
	}
	return originalTypes, sources, nil
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
		if _, ok := expr.(LiteralInterface); ok {
			return Source_Constant
		}
		return Source_Expression
	}
}

// possibleParameterTypes returns the parameter types of all of the expressions. This is accomplished by gathering the
// return types of each parameter.
func (c *CompiledFunction) possibleParameterTypes() []pgtypes.DoltgresType {
	possibleParamTypes := make([]pgtypes.DoltgresType, len(c.Parameters))
	for i, param := range c.Parameters {
		expressionType := param.Type()
		if extendedType, ok := expressionType.(pgtypes.DoltgresType); ok {
			possibleParamTypes[i] = extendedType
		} else {
			// TODO: we need to remove GMS types from all of our expressions so that we can remove this
			switch param.Type().Type() {
			case query.Type_INT8, query.Type_INT16:
				possibleParamTypes[i] = pgtypes.Int16
			case query.Type_INT24, query.Type_INT32:
				possibleParamTypes[i] = pgtypes.Int32
			case query.Type_INT64:
				possibleParamTypes[i] = pgtypes.Int64
			case query.Type_UINT8, query.Type_UINT16, query.Type_UINT24, query.Type_UINT32, query.Type_UINT64:
				possibleParamTypes[i] = pgtypes.Int64
			case query.Type_YEAR:
				possibleParamTypes[i] = pgtypes.Int16
			case query.Type_FLOAT32:
				possibleParamTypes[i] = pgtypes.Float32
			case query.Type_FLOAT64:
				possibleParamTypes[i] = pgtypes.Float64
			case query.Type_DECIMAL:
				possibleParamTypes[i] = pgtypes.Numeric
			case query.Type_DATE, query.Type_DATETIME, query.Type_TIMESTAMP:
				// TODO: need to add DoltgresType equivalents to DATETIME
				possibleParamTypes[i] = pgtypes.Null
			case query.Type_CHAR, query.Type_VARCHAR:
				possibleParamTypes[i] = pgtypes.VarCharMax
			case query.Type_TEXT:
				possibleParamTypes[i] = pgtypes.VarCharMax
			case query.Type_ENUM:
				possibleParamTypes[i] = pgtypes.Int16
			case query.Type_SET:
				possibleParamTypes[i] = pgtypes.Int64
			default:
				possibleParamTypes[i] = pgtypes.Null
			}
		}
	}
	return possibleParamTypes
}
