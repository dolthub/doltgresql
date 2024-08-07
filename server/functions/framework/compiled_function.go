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
	Name      string
	Arguments []sql.Expression
	// TODO: separate the resolution vars from the run-time vars, they aren't needed after resolution
	OverloadTree      *Overloads
	ParamPermutations []functionOverload
	IsOperator        bool
	overload          overloadMatch
	callableFunc      FunctionInterface
	originalTypes     []pgtypes.DoltgresType
	callResolved      []pgtypes.DoltgresType
	stashedErr        error
}

var _ sql.FunctionExpression = (*CompiledFunction)(nil)
var _ sql.NonDeterministicExpression = (*CompiledFunction)(nil)

// NewCompiledFunction returns a newly compiled function.
func NewCompiledFunction(name string, args []sql.Expression, functions *Overloads, isOperator bool) *CompiledFunction {
	return newCompiledFunctionInternal(name, args, functions, functions.overloadsForParams(len(args)), isOperator)
}

// newCompiledFunctionInternal is called internally, which skips steps that may have already been processed.
func newCompiledFunctionInternal(
	name string,
	args []sql.Expression,
	overloads *Overloads,
	fnOverloads []functionOverload,
	isOperator bool,
) *CompiledFunction {
	c := &CompiledFunction{
		Name:              name,
		Arguments:         args,
		OverloadTree:      overloads,
		ParamPermutations: fnOverloads,
		IsOperator:        isOperator,
	}
	// First we'll analyze all of the parameters.
	originalTypes, sources, err := c.analyzeParameters()
	if err != nil {
		// Errors should be returned from the call to Eval, so we'll stash it for now
		c.stashedErr = err
		return c
	}
	// Next we'll resolve the overload based on the parameters given.
	overload, err := c.resolve(originalTypes, sources)
	if err != nil {
		c.stashedErr = err
		return c
	}
	// If we do not receive an overload, then the parameters given did not result in a valid match
	if !overload.Valid() {
		c.stashedErr = fmt.Errorf("function %s does not exist", c.OverloadString(originalTypes))
		return c
	}

	fn := overload.Function()

	// Then we'll handle the polymorphic types
	// https://www.postgresql.org/docs/15/extend-type-system.html#EXTEND-TYPES-POLYMORPHIC
	functionParameterTypes := fn.GetParameters()
	c.callResolved = make([]pgtypes.DoltgresType, len(functionParameterTypes)+1)
	hasPolymorphicParam := false
	for i, param := range functionParameterTypes {
		if _, ok := param.(pgtypes.DoltgresPolymorphicType); ok {
			// resolve will ensure that the parameter types are valid, so we can just assign them here
			hasPolymorphicParam = true
			c.callResolved[i] = originalTypes[i]
		} else {
			c.callResolved[i] = param
		}
	}
	returnType := fn.GetReturn()
	c.callResolved[len(c.callResolved)-1] = returnType
	if _, ok := returnType.(pgtypes.DoltgresPolymorphicType); ok {
		if hasPolymorphicParam {
			c.callResolved[len(c.callResolved)-1] = c.resolvePolymorphicReturnType(functionParameterTypes, originalTypes, returnType)
		} else {
			c.stashedErr = fmt.Errorf("A result of type %s requires at least one input of type "+
				"anyelement, anyarray, anynonarray, anyenum, anyrange, or anymultirange.", returnType.String())
			return c
		}
	}

	// Lastly, we assign everything to the function struct
	c.callableFunc = fn
	c.overload = overload
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
	for _, param := range c.Arguments {
		if !param.Resolved() {
			return false
		}
	}
	return c.overload.Valid()
}

// String implements the interface sql.Expression.
func (c *CompiledFunction) String() string {
	sb := strings.Builder{}
	sb.WriteString(c.Name + "(")
	for i, param := range c.Arguments {
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
	if len(c.callResolved) > 0 {
		return c.callResolved[len(c.callResolved)-1]
	}
	// Compilation must have errored, so we'll return the unknown type
	return pgtypes.Unknown
}

// IsNullable implements the interface sql.Expression.
func (c *CompiledFunction) IsNullable() bool {
	// All functions seem to return NULL when given a NULL value
	return true
}

// IsNonDeterministic implements the interface sql.NonDeterministicExpression.
func (c *CompiledFunction) IsNonDeterministic() bool {
	if c.callableFunc != nil {
		return c.callableFunc.NonDeterministic()
	}
	// Compilation must have errored, so we'll just return true
	return true
}

// Eval implements the interface sql.Expression.
func (c *CompiledFunction) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	// If we have a stashed error, then we should return that now. Errors are stashed when they're supposed to be
	// returned during the call to Eval. This helps to ensure consistency with how errors are returned in Postgres.
	if c.stashedErr != nil {
		return nil, c.stashedErr
	}

	// Evaluate all of the arguments.
	args, err := c.evalArgs(ctx, row)
	if err != nil {
		return nil, err
	}

	targetParamTypes := c.callableFunc.GetParameters()

	if c.callableFunc.IsStrict() {
		for i := range args {
			if args[i] == nil {
				return nil, nil
			}
		}
	}

	if len(c.overload.casts) > 0 {
		for i, arg := range args {
			paramIdx := i
			isVariadicArg := c.overload.params.variadic >= 0 && i >= len(c.overload.params.paramTypes)-1
			if isVariadicArg {
				paramIdx = len(c.overload.params.paramTypes) - 1
			}

			if c.overload.casts[paramIdx] != nil {
				targetType := targetParamTypes[paramIdx]
				if isVariadicArg {
					targetArrayType, ok := targetType.(pgtypes.DoltgresArrayType)
					if !ok {
						return nil, fmt.Errorf("variadic arguments must be array types, was %T", targetType)
					}
					targetType = targetArrayType.BaseType()
				}

				args[i], err = c.overload.casts[paramIdx](ctx, arg, targetType)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, fmt.Errorf("function %s is missing the appropriate implicit cast", c.OverloadString(c.originalTypes))
			}
		}
	}

	args = c.overload.params.coalesceVariadicValues(args)

	// Call the function
	switch f := c.callableFunc.(type) {
	case Function0:
		return f.Callable(ctx)
	case Function1:
		return f.Callable(ctx, ([2]pgtypes.DoltgresType)(c.callResolved), args[0])
	case Function2:
		return f.Callable(ctx, ([3]pgtypes.DoltgresType)(c.callResolved), args[0], args[1])
	case Function3:
		return f.Callable(ctx, ([4]pgtypes.DoltgresType)(c.callResolved), args[0], args[1], args[2])
	case Function4:
		return f.Callable(ctx, ([5]pgtypes.DoltgresType)(c.callResolved), args[0], args[1], args[2], args[3])
	default:
		return nil, fmt.Errorf("unknown function type in CompiledFunction::Eval")
	}
}

// Children implements the interface sql.Expression.
func (c *CompiledFunction) Children() []sql.Expression {
	return c.Arguments
}

// WithChildren implements the interface sql.Expression.
func (c *CompiledFunction) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	return newCompiledFunctionInternal(c.Name, children, c.OverloadTree, c.ParamPermutations, c.IsOperator), nil
}

// resolve returns an overloadMatch that either matches the given parameters exactly, or is a viable match after casting.
// Returns an empty overloadMatch if a viable match is not found.
func (c *CompiledFunction) resolve(argTypes []pgtypes.DoltgresType, sources []Source) (overloadMatch, error) {
	// First check for an exact match
	exactMatch, found := c.OverloadTree.ExactMatchForTypes(argTypes)
	if found {
		// empty overload match for an exact match
		baseTypes := baseIdsFortypes(argTypes)
		return overloadMatch{
			params: functionOverload{
				function:   exactMatch,
				paramTypes: baseTypes,
				argTypes:   baseTypes,
				variadic:   -1,
			},
		}, nil
	}
	// There are no exact matches, so now we'll look through all of the functions to determine the best match. This is
	// much more work, but there's a performance penalty for runtime overload resolution in Postgres as well.
	if c.IsOperator {
		return c.resolveOperator(argTypes, sources)
	} else {
		return c.resolveFunction(argTypes, sources)
	}
}

// overloadMatch is the result of a successful overload resolution, containing the types of the parameters as well
// as the type cast functions required to convert every argument to its appropriate parameter type
type overloadMatch struct {
	params functionOverload
	casts  []TypeCastFunction
}

// Valid returns whether this overload is valid (has a callable function)
func (o overloadMatch) Valid() bool {
	return o.params.function != nil
}

// Function returns the function for this overload
func (o overloadMatch) Function() FunctionInterface {
	return o.params.function
}

// resolveFunction resolves a function according to the rules defined by Postgres.
// https://www.postgresql.org/docs/15/typeconv-func.html
func (c *CompiledFunction) resolveFunction(argTypes []pgtypes.DoltgresType, sources []Source) (overloadMatch, error) {

	compatibleOverloads := c.typeCompatibleOverloads(argTypes, sources)
	if len(compatibleOverloads) == 0 {
		return overloadMatch{}, nil
	}

	// If we've found exactly one match then we'll return that one
	if len(compatibleOverloads) == 1 {
		return compatibleOverloads[0], nil
	}

	matches := exactTypeMatches(argTypes, compatibleOverloads)
	if len(matches) == 0 {
		return overloadMatch{}, nil
	}

	// Now check again for exactly one match
	if len(matches) == 1 {
		return matches[0], nil
	}

	preferredOverloads := preferredTypeMatches(argTypes, compatibleOverloads)
	if len(preferredOverloads) == 0 {
		return overloadMatch{}, nil
	}

	// Check once more for exactly one match
	if len(preferredOverloads) == 1 {
		return preferredOverloads[0], nil
	}

	// No matching function overload found
	return overloadMatch{}, nil
}

// preferredTypeMatches returns the overload candidates that have the most preferred types for args that require casts.
func preferredTypeMatches(argTypes []pgtypes.DoltgresType, candidates []overloadMatch) []overloadMatch {
	preferredCount := 0
	var preferredOverloads []overloadMatch
	for _, cand := range candidates {
		currentPreferredCount := 0
		for argIdx := range argTypes {
			paramType := cand.params.argTypes[argIdx]

			if argTypes[argIdx].BaseID() != paramType && paramType.GetTypeCategory().IsPreferredType(paramType) {
				currentPreferredCount++
			}
		}

		if currentPreferredCount > preferredCount {
			preferredCount = currentPreferredCount
			preferredOverloads = append([]overloadMatch{}, cand)
		} else if currentPreferredCount == preferredCount {
			preferredOverloads = append(preferredOverloads, cand)
		}
	}
	return preferredOverloads
}

// exactTypeMatches returns the set of overload candidates that have the most exact type matches for the arg types
// provided.
func exactTypeMatches(argTypes []pgtypes.DoltgresType, candidates []overloadMatch) []overloadMatch {
	matchCount := 0
	var matches []overloadMatch
	for _, cand := range candidates {
		currentMatchCount := 0
		for argIdx := range argTypes {
			paramType := cand.params.argTypes[argIdx]

			// NULL values count as exact matches, since all types accept NULL as a valid value
			paramBaseID := argTypes[argIdx].BaseID()
			if paramBaseID == paramType || paramBaseID == pgtypes.DoltgresTypeBaseID_Null {
				currentMatchCount++
			}
		}
		if currentMatchCount > matchCount {
			matchCount = currentMatchCount
			matches = append([]overloadMatch{}, cand)
		} else if currentMatchCount == matchCount {
			matches = append(matches, cand)
		}
	}
	return matches
}

// typeCompatibleOverloads returns all overloads that have a matching number of params whose types can be
// implicitly converted to the ones provided. This is the set of all possible overloads that could be used with the
// param types provided.
func (c *CompiledFunction) typeCompatibleOverloads(argTypes []pgtypes.DoltgresType, sources []Source) []overloadMatch {
	var compatible []overloadMatch
	for _, overload := range c.ParamPermutations {
		if !paramNumbersMatch(argTypes, overload) {
			continue
		}

		isConvertible := true
		overloadCasts := make([]TypeCastFunction, len(argTypes))
		// Polymorphic parameters must be gathered so that we can later verify that they all have matching base types
		var polymorphicParameters []pgtypes.DoltgresType
		var polymorphicTargets []pgtypes.DoltgresType
		for i := range argTypes {
			paramType := overload.argTypes[i]

			if argTypes[i].BaseID() == pgtypes.DoltgresTypeBaseID_Null {
				overloadCasts[i] = identityCast
			} else if polymorphicType, ok := paramType.GetRepresentativeType().(pgtypes.DoltgresPolymorphicType); ok && polymorphicType.IsValid(argTypes[i]) {
				overloadCasts[i] = identityCast
				polymorphicParameters = append(polymorphicParameters, polymorphicType)
				polymorphicTargets = append(polymorphicTargets, argTypes[i])
			} else {
				if overloadCasts[i] = GetImplicitCast(argTypes[i].BaseID(), paramType); overloadCasts[i] == nil {
					if sources[i] == Source_Constant && argTypes[i].BaseID().GetTypeCategory() == pgtypes.TypeCategory_StringTypes {
						overloadCasts[i] = stringLiteralCast
					} else {
						isConvertible = false
						break
					}
				}
			}
		}

		if isConvertible && polymorphicTypesCompatible(polymorphicParameters, polymorphicTargets) {
			compatible = append(compatible, overloadMatch{params: overload, casts: overloadCasts})
		}
	}
	return compatible
}

func paramNumbersMatch(paramTypes []pgtypes.DoltgresType, overload functionOverload) bool {
	return len(overload.paramTypes) == len(paramTypes) ||
		(overload.variadic >= 0 && len(paramTypes) >= len(overload.paramTypes)-1)
}

// resolveOperator resolves an operator according to the rules defined by Postgres.
// https://www.postgresql.org/docs/15/typeconv-oper.html
func (c *CompiledFunction) resolveOperator(argTypes []pgtypes.DoltgresType, sources []Source) (overloadMatch, error) {
	// Binary operators treat string literals as the other type, so we'll account for that here to see if we can find an
	// "exact" match.
	if len(argTypes) == 2 {
		leftStringLiteral := sources[0] == Source_Constant && argTypes[0].BaseID().GetTypeCategory() == pgtypes.TypeCategory_StringTypes
		rightStringLiteral := sources[1] == Source_Constant && argTypes[1].BaseID().GetTypeCategory() == pgtypes.TypeCategory_StringTypes
		if (leftStringLiteral && !rightStringLiteral) || (!leftStringLiteral && rightStringLiteral) {
			var baseID pgtypes.DoltgresTypeBaseID
			casts := []TypeCastFunction{identityCast, identityCast}
			if leftStringLiteral {
				casts[0] = stringLiteralCast
				baseID = argTypes[1].BaseID()
			} else {
				casts[1] = stringLiteralCast
				baseID = argTypes[0].BaseID()
			}
			if exactMatch, ok := c.OverloadTree.ExactMatchForBaseIds(baseID, baseID); ok {
				return overloadMatch{
					params: functionOverload{
						function:   exactMatch,
						paramTypes: []pgtypes.DoltgresTypeBaseID{baseID, baseID},
						argTypes:   []pgtypes.DoltgresTypeBaseID{baseID, baseID},
						variadic:   -1,
					},
					casts: casts,
				}, nil
			}
		}
	}
	// From this point, the steps appear to be the same for functions and operators
	return c.resolveFunction(argTypes, sources)
}

// polymorphicTypesCompatible returns whether any polymorphic types given are compatible with the expression types given
func polymorphicTypesCompatible(paramTypes []pgtypes.DoltgresType, exprTypes []pgtypes.DoltgresType) bool {
	if len(paramTypes) != len(exprTypes) {
		return false
	}
	// If there are less than two parameters then we don't even need to check
	if len(paramTypes) < 2 {
		return true
	}

	// If one of the types is anyarray, then anyelement behaves as anynonarray, so we can convert them to anynonarray
	for _, paramType := range paramTypes {
		if polymorphicParamType, ok := paramType.(pgtypes.DoltgresPolymorphicType); ok && polymorphicParamType.BaseID() == pgtypes.DoltgresTypeBaseID_AnyArray {
			// At least one parameter is anyarray, so copy all parameters to a new slice and replace anyelement with anynonarray
			newParamTypes := make([]pgtypes.DoltgresType, len(paramTypes))
			copy(newParamTypes, paramTypes)
			for i := range newParamTypes {
				if paramTypes[i].BaseID() == pgtypes.DoltgresTypeBaseID_AnyElement {
					newParamTypes[i] = pgtypes.AnyNonArray
				}
			}
			paramTypes = newParamTypes
			break
		}
	}

	// The base type is the type that must match between all polymorphic types.
	var baseType pgtypes.DoltgresType
	for i, paramType := range paramTypes {
		if polymorphicParamType, ok := paramType.(pgtypes.DoltgresPolymorphicType); ok {
			// Although we do this check before we ever reach this function, we do it again as we may convert anyelement
			// to anynonarray, which changes type validity
			if !polymorphicParamType.IsValid(exprTypes[i]) {
				return false
			}
			// Get the base expression type that we'll compare against
			baseExprType := exprTypes[i]
			if arrayBaseExprType, ok := baseExprType.(pgtypes.DoltgresArrayType); ok {
				baseExprType = arrayBaseExprType.BaseType()
			}
			// TODO: handle range types
			// Check that the base expression type matches the previously-found base type
			if baseType == nil {
				baseType = baseExprType
			} else if baseType.BaseID() != baseExprType.BaseID() {
				return false
			}
		}
	}
	return true
}

// resolvePolymorphicReturnType returns the type that should be used for the return type. If the return type is not a
// polymorphic type, then the return type is directly returned. However, if the return type is a polymorphic type, then
// the type is determined using the expression types and parameter types. This makes the assumption that everything has
// already been validated.
func (c *CompiledFunction) resolvePolymorphicReturnType(functionInterfaceTypes []pgtypes.DoltgresType, originalTypes []pgtypes.DoltgresType, returnType pgtypes.DoltgresType) pgtypes.DoltgresType {
	polymorphicReturnType, ok := returnType.(pgtypes.DoltgresPolymorphicType)
	if !ok {
		return returnType
	}
	// We can use the first polymorphic type that we find, since we can morph it into any type that we need.
	// We've verified that all polymorphic types are compatible in a previous step, so this is safe to do.
	var firstPolymorphicType pgtypes.DoltgresType
	for i, functionInterfaceType := range functionInterfaceTypes {
		if _, ok = functionInterfaceType.(pgtypes.DoltgresPolymorphicType); ok {
			firstPolymorphicType = originalTypes[i]
			break
		}
	}
	switch polymorphicReturnType.BaseID() {
	case pgtypes.DoltgresTypeBaseID_AnyElement, pgtypes.DoltgresTypeBaseID_AnyNonArray:
		// For return types, anyelement behaves the same as anynonarray.
		// This isn't explicitly in the documentation, however it does note that:
		// "...anynonarray and anyenum do not represent separate type variables; they are the same type as anyelement..."
		// The implication of this being that anyelement will always return the base type even for array types,
		// just like anynonarray would.
		if minimalArrayType, ok := firstPolymorphicType.(pgtypes.DoltgresArrayType); ok {
			return minimalArrayType.BaseType()
		} else {
			return firstPolymorphicType
		}
	case pgtypes.DoltgresTypeBaseID_AnyArray:
		// Array types will return themselves, so this is safe
		return firstPolymorphicType.ToArrayType()
	default:
		panic(fmt.Errorf("`%s` is not yet handled during function compilation", polymorphicReturnType.String()))
	}
}

// evalArgs evaluates the function args within an Eval call.
func (c *CompiledFunction) evalArgs(ctx *sql.Context, row sql.Row) ([]any, error) {
	args := make([]any, len(c.Arguments))
	for i, arg := range c.Arguments {
		var err error
		args[i], err = arg.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		// TODO: once we remove GMS types from all of our expressions, we can remove this step which ensures the correct type
		if _, ok := arg.Type().(pgtypes.DoltgresType); !ok {
			switch arg.Type().Type() {
			case query.Type_INT8, query.Type_INT16:
				args[i], _, _ = pgtypes.Int16.Convert(args[i])
			case query.Type_INT24, query.Type_INT32:
				args[i], _, _ = pgtypes.Int32.Convert(args[i])
			case query.Type_INT64:
				args[i], _, _ = pgtypes.Int64.Convert(args[i])
			case query.Type_UINT8, query.Type_UINT16, query.Type_UINT24, query.Type_UINT32, query.Type_UINT64:
				args[i], _, _ = pgtypes.Int64.Convert(args[i])
			case query.Type_YEAR:
				args[i], _, _ = pgtypes.Int16.Convert(args[i])
			case query.Type_FLOAT32:
				args[i], _, _ = pgtypes.Float32.Convert(args[i])
			case query.Type_FLOAT64:
				args[i], _, _ = pgtypes.Float64.Convert(args[i])
			case query.Type_DECIMAL:
				args[i], _, _ = pgtypes.Numeric.Convert(args[i])
			case query.Type_DATE, query.Type_DATETIME, query.Type_TIMESTAMP:
				return nil, fmt.Errorf("need to add DoltgresType equivalents to DATETIME")
			case query.Type_CHAR, query.Type_VARCHAR, query.Type_TEXT:
				args[i], _, _ = pgtypes.Text.Convert(args[i])
			case query.Type_ENUM:
				args[i], _, _ = pgtypes.Int16.Convert(args[i])
			case query.Type_SET:
				args[i], _, _ = pgtypes.Int64.Convert(args[i])
			default:
				return nil, fmt.Errorf("encountered a GMS type that cannot be handled")
			}
		}
	}
	return args, nil
}

// analyzeParameters analyzes the parameters within an Eval call.
func (c *CompiledFunction) analyzeParameters() (originalTypes []pgtypes.DoltgresType, sources []Source, err error) {
	originalTypes = make([]pgtypes.DoltgresType, len(c.Arguments))
	sources = make([]Source, len(c.Arguments))
	for i, param := range c.Arguments {
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
				originalTypes[i] = pgtypes.Timestamp
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
