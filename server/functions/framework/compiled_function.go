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
	Name          string
	Parameters    []sql.Expression
	Functions     *OverloadDeduction
	AllOverloads  [][]pgtypes.DoltgresTypeBaseID
	IsOperator    bool
	callableFunc  FunctionInterface
	casts         []TypeCastFunction
	originalTypes []pgtypes.DoltgresType
	stashedErr    error
}

var _ sql.FunctionExpression = (*CompiledFunction)(nil)
var _ sql.NonDeterministicExpression = (*CompiledFunction)(nil)

// NewCompiledFunction returns a newly compiled function.
func NewCompiledFunction(name string, parameters []sql.Expression, functions *OverloadDeduction, isOperator bool) *CompiledFunction {
	return newCompiledFunctionInternal(name, parameters, functions, functions.collectOverloadPermutations(), isOperator)
}

// newCompiledFunctionInternal is called internally, which skips steps that may have already been processed.
func newCompiledFunctionInternal(name string, params []sql.Expression, funcs *OverloadDeduction, allFuncs [][]pgtypes.DoltgresTypeBaseID, isOperator bool) *CompiledFunction {
	c := &CompiledFunction{
		Name:         name,
		Parameters:   params,
		Functions:    funcs,
		AllOverloads: allFuncs,
		IsOperator:   isOperator,
	}
	// First we'll analyze all of the parameters.
	originalTypes, sources, err := c.analyzeParameters()
	if err != nil {
		// Errors should be returned from the call to Eval, so we'll stash it for now
		c.stashedErr = err
		return c
	}
	// Next we'll resolve the overload based on the parameters given.
	overload, casts, err := c.resolve(originalTypes, sources)
	if err != nil {
		c.stashedErr = err
		return c
	}
	// If we do not receive an overload, then the parameters given did not result in a valid match
	if overload == nil || overload.Function == nil {
		c.stashedErr = fmt.Errorf("function %s does not exist", c.OverloadString(originalTypes))
		return c
	}
	c.callableFunc = overload.Function
	c.casts = casts
	c.originalTypes = originalTypes
	return c
}

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
		// Aliases will output the string "x as x", which is an artifact of how we build the AST, so we'll bypass it
		if alias, ok := param.(*expression.Alias); ok {
			param = alias.Child
		}
		if i > 0 {
			sb.WriteString(", ")
		}
		if doltgresType, ok := param.Type().(pgtypes.DoltgresType); ok {
			sb.WriteString(pgtypes.QuoteString(doltgresType.BaseID(), param.String()))
		} else {
			sb.WriteString(param.String())
		}
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
	parameters, sources := c.possibleParameterTypes()
	if resolvedFunction, _, _ := c.resolve(parameters, sources); resolvedFunction != nil {
		return resolvedFunction.Function.GetReturn()
	}
	// We can't resolve to a function before evaluation in this case, so we'll return something arbitrary
	return pgtypes.Unknown
}

// IsNullable implements the interface sql.Expression.
func (c *CompiledFunction) IsNullable() bool {
	// All functions seem to return NULL when given a NULL value
	return true
}

// IsNonDeterministic implements the interface sql.NonDeterministicExpression.
func (c *CompiledFunction) IsNonDeterministic() bool {
	// TODO: determine if this needs to inspect the parameters as well
	parameters, sources := c.possibleParameterTypes()
	if resolvedFunction, _, _ := c.resolve(parameters, sources); resolvedFunction != nil {
		return resolvedFunction.Function.GetIsNonDeterministic()
	}
	// We can't resolve to a function before evaluation in this case, so we'll return true to be on the safe side
	return true
}

// Eval implements the interface sql.Expression.
func (c *CompiledFunction) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	// If we have a stashed error, then we should return that now. Errors are stashed when they're supposed to be
	// returned during the call to Eval. This helps to ensure consistency with how errors are returned in Postgres.
	if c.stashedErr != nil {
		return nil, c.stashedErr
	}
	// Evaluate all of the parameters.
	parameters, err := c.evalParameters(ctx, row)
	if err != nil {
		return nil, err
	}
	// Convert the parameter values into their correct types
	resultTypes := c.callableFunc.GetParameters()
	if len(c.casts) > 0 {
		for i := range parameters {
			if parameters[i] == nil && c.callableFunc.GetIsStrict() {
				return nil, nil
			} else if c.casts[i] != nil {
				parameters[i], err = c.casts[i](ctx, parameters[i], resultTypes[i])
				if err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("function %s is missing the appropriate implicit cast", c.OverloadString(c.originalTypes))
			}
		}
	}
	// Pass the parameters to the function
	switch f := c.callableFunc.(type) {
	case Function0:
		return f.Callable(ctx)
	case Function1:
		return f.Callable(ctx, parameters[0])
	case Function2:
		return f.Callable(ctx, parameters[0], parameters[1])
	case Function3:
		return f.Callable(ctx, parameters[0], parameters[1], parameters[2])
	case Function4:
		return f.Callable(ctx, parameters[0], parameters[1], parameters[2], parameters[3])
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
	return newCompiledFunctionInternal(c.Name, children, c.Functions, c.AllOverloads, c.IsOperator), nil
}

// resolve returns an overload that either matches the given parameters exactly, or is a viable match after casting.
// Returns a nil OverloadDeduction if a viable match is not found.
func (c *CompiledFunction) resolve(parameters []pgtypes.DoltgresType, sources []Source) (*OverloadDeduction, []TypeCastFunction, error) {
	// First check for an exact match
	exactMatch := c.Functions
	for _, parameter := range parameters {
		var ok bool
		if exactMatch, ok = exactMatch.Parameter[parameter.BaseID()]; !ok {
			break
		}
	}
	if exactMatch != nil && exactMatch.Function != nil {
		return exactMatch, nil, nil
	}
	// There are no exact matches, so now we'll look through all of the functions to determine the best match. This is
	// much more work, but there's a performance penalty for runtime overload resolution in Postgres as well.
	if c.IsOperator {
		return c.resolveOperator(parameters, sources)
	} else {
		return c.resolveFunction(parameters, sources)
	}
}

// resolveFunction resolves a function according to the rules defined by Postgres.
// https://www.postgresql.org/docs/15/typeconv-func.html
func (c *CompiledFunction) resolveFunction(parameters []pgtypes.DoltgresType, sources []Source) (*OverloadDeduction, []TypeCastFunction, error) {
	// First we'll discard all overloads that have a different length, or that do not have implicitly-convertible params
	var convertibles [][]pgtypes.DoltgresTypeBaseID
	var casts [][]TypeCastFunction
	for _, overload := range c.AllOverloads {
		if len(overload) == len(parameters) {
			isConvertible := true
			overloadCasts := make([]TypeCastFunction, len(overload))
			for i, overloadParam := range overload {
				if overloadCasts[i] = GetImplicitCast(parameters[i].BaseID(), overloadParam); overloadCasts[i] == nil {
					if sources[i] == Source_Constant && parameters[i].BaseID().GetTypeCategory() == pgtypes.TypeCategory_StringTypes {
						overloadCasts[i] = stringLiteralCast
					} else {
						isConvertible = false
						break
					}
				}
			}
			if isConvertible {
				convertibles = append(convertibles, overload)
				casts = append(casts, overloadCasts)
			}
		}
	}
	// If we've found exactly one match then we'll return that one
	if len(convertibles) == 1 {
		matchedOverload := c.Functions
		for _, parameter := range convertibles[0] {
			matchedOverload = matchedOverload.Parameter[parameter]
		}
		return matchedOverload, casts[0], nil
	} else if len(convertibles) == 0 {
		return nil, nil, nil
	}
	// Next we'll keep the functions that have the most exact matches, or all of them if none have exact matches
	matchCount := 0
	var matches [][]pgtypes.DoltgresTypeBaseID
	var matchCasts [][]TypeCastFunction
	for convertibleIdx, convertible := range convertibles {
		currentMatchCount := 0
		for paramIdx, param := range convertible {
			if parameters[paramIdx].BaseID() == param {
				currentMatchCount++
			}
		}
		if currentMatchCount > matchCount {
			matchCount = currentMatchCount
			matches = append([][]pgtypes.DoltgresTypeBaseID{}, convertible)
			matchCasts = append([][]TypeCastFunction{}, casts[convertibleIdx])
		} else if currentMatchCount == matchCount {
			matches = append(matches, convertible)
			matchCasts = append(matchCasts, casts[convertibleIdx])
		}
	}
	// Now check again for exactly one match
	if len(matches) == 1 {
		matchedOverload := c.Functions
		for _, parameter := range matches[0] {
			matchedOverload = matchedOverload.Parameter[parameter]
		}
		return matchedOverload, matchCasts[0], nil
	} else if len(matches) == 0 {
		return nil, nil, nil
	}
	// Check for preferred types, retaining those that have the most preferred types for parameters that require casts
	preferredCount := 0
	var preferredOverloads [][]pgtypes.DoltgresTypeBaseID
	var preferredCasts [][]TypeCastFunction
	for matchIdx, match := range matches {
		currentPreferredCount := 0
		for paramIdx, param := range match {
			if parameters[paramIdx].BaseID() != param && param.GetTypeCategory().GetPreferredType() == param {
				currentPreferredCount++
			}
		}
		if currentPreferredCount > preferredCount {
			preferredCount = currentPreferredCount
			preferredOverloads = append([][]pgtypes.DoltgresTypeBaseID{}, match)
			preferredCasts = append([][]TypeCastFunction{}, matchCasts[matchIdx])
		} else if currentPreferredCount == preferredCount {
			preferredOverloads = append(preferredOverloads, match)
			preferredCasts = append(preferredCasts, matchCasts[matchIdx])
		}
	}
	// Check once more for exactly one match
	if len(preferredOverloads) == 1 {
		matchedOverload := c.Functions
		for _, parameter := range preferredOverloads[0] {
			matchedOverload = matchedOverload.Parameter[parameter]
		}
		return matchedOverload, preferredCasts[0], nil
	} else if len(preferredOverloads) == 0 {
		return nil, nil, nil
	}
	return nil, nil, nil
}

// resolveOperator resolves an operator according to the rules defined by Postgres.
// https://www.postgresql.org/docs/15/typeconv-oper.html
func (c *CompiledFunction) resolveOperator(parameters []pgtypes.DoltgresType, sources []Source) (*OverloadDeduction, []TypeCastFunction, error) {
	// Binary operators treat string literals as the other type, so we'll account for that here to see if we can find an
	// "exact" match.
	if len(parameters) == 2 {
		leftStringLiteral := sources[0] == Source_Constant && parameters[0].BaseID().GetTypeCategory() == pgtypes.TypeCategory_StringTypes
		rightStringLiteral := sources[1] == Source_Constant && parameters[1].BaseID().GetTypeCategory() == pgtypes.TypeCategory_StringTypes
		if (leftStringLiteral && !rightStringLiteral) || (!leftStringLiteral && rightStringLiteral) {
			var baseID pgtypes.DoltgresTypeBaseID
			casts := []TypeCastFunction{identityCast, identityCast}
			if leftStringLiteral {
				casts[0] = stringLiteralCast
				baseID = parameters[1].BaseID()
			} else {
				casts[1] = stringLiteralCast
				baseID = parameters[0].BaseID()
			}
			if exactMatch, ok := c.Functions.Parameter[baseID]; ok {
				if exactMatch, ok = exactMatch.Parameter[baseID]; ok {
					return exactMatch, casts, nil
				}
			}
		}
	}
	// From this point, the steps appear to be the same for functions and operators
	return c.resolveFunction(parameters, sources)
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
			case query.Type_CHAR, query.Type_VARCHAR, query.Type_TEXT:
				parameters[i], _, _ = pgtypes.Text.Convert(parameters[i])
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
			case query.Type_CHAR, query.Type_VARCHAR, query.Type_TEXT:
				originalTypes[i] = pgtypes.Text
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
func (c *CompiledFunction) possibleParameterTypes() ([]pgtypes.DoltgresType, []Source) {
	possibleParamTypes := make([]pgtypes.DoltgresType, len(c.Parameters))
	sources := make([]Source, len(c.Parameters))
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
				possibleParamTypes[i] = pgtypes.Text
			case query.Type_ENUM:
				possibleParamTypes[i] = pgtypes.Int16
			case query.Type_SET:
				possibleParamTypes[i] = pgtypes.Int64
			default:
				possibleParamTypes[i] = pgtypes.Null
			}
		}
		sources[i] = c.determineSource(param)
	}
	return possibleParamTypes, sources
}
