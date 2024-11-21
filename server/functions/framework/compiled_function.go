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
	Arguments     []sql.Expression
	IsOperator    bool
	overloads     *Overloads
	fnOverloads   []Overload
	overload      overloadMatch
	originalTypes []pgtypes.DoltgresType
	callResolved  []pgtypes.DoltgresType
	stashedErr    error
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
	fnOverloads []Overload,
	isOperator bool,
) *CompiledFunction {
	c := &CompiledFunction{
		Name:        name,
		Arguments:   args,
		IsOperator:  isOperator,
		overloads:   overloads,
		fnOverloads: fnOverloads,
	}
	// First we'll analyze all the parameters.
	originalTypes, err := c.analyzeParameters()
	if err != nil {
		// Errors should be returned from the call to Eval, so we'll stash it for now
		c.stashedErr = err
		return c
	}
	// Next we'll resolve the overload based on the parameters given.
	overload, err := c.resolve(overloads, fnOverloads, originalTypes)
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
	// We don't error until evaluation time, so we need to tell the engine we're resolved if there was a stashed error
	return c.stashedErr != nil || c.overload.Valid()
}

// StashedError returns the stashed error if one exists. Otherwise, returns nil.
func (c *CompiledFunction) StashedError() error {
	if c == nil {
		return nil
	}
	return c.stashedErr
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
	if c.overload.Valid() {
		return c.overload.Function().NonDeterministic()
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

	if c.overload.Function().IsStrict() {
		for i := range args {
			if args[i] == nil {
				return nil, nil
			}
		}
	}

	targetParamTypes := c.overload.Function().GetParameters()

	if len(c.overload.casts) > 0 {
		for i, arg := range args {
			// For variadic params, we need to identify the corresponding target type
			var targetType pgtypes.DoltgresType
			isVariadicArg := c.overload.params.variadic >= 0 && i >= len(c.overload.params.paramTypes)-1
			if isVariadicArg {
				targetType = targetParamTypes[c.overload.params.variadic]
				targetArrayType, ok := targetType.(pgtypes.DoltgresArrayType)
				if !ok {
					// should be impossible, we check this at function compile time
					return nil, fmt.Errorf("variadic arguments must be array types, was %T", targetType)
				}
				targetType = targetArrayType.BaseType()
			} else {
				targetType = targetParamTypes[i]
			}

			if c.overload.casts[i] != nil {
				args[i], err = c.overload.casts[i](ctx, arg, targetType)
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
	switch f := c.overload.Function().(type) {
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
	if len(children) != len(c.Arguments) {
		return nil, sql.ErrInvalidChildrenNumber.New(len(children), len(c.Arguments))
	}

	// We have to re-resolve here, since the change in children may require it (e.g. we have more type info than we did)
	return newCompiledFunctionInternal(c.Name, children, c.overloads, c.fnOverloads, c.IsOperator), nil
}

// resolve returns an overloadMatch that either matches the given parameters exactly, or is a viable match after casting.
// Returns an invalid overloadMatch if a viable match is not found.
func (c *CompiledFunction) resolve(
	overloads *Overloads,
	fnOverloads []Overload,
	argTypes []pgtypes.DoltgresType,
) (overloadMatch, error) {

	// First check for an exact match
	exactMatch, found := overloads.ExactMatchForTypes(argTypes)
	if found {
		baseTypes := overloads.baseIdsForTypes(argTypes)
		return overloadMatch{
			params: Overload{
				function:   exactMatch,
				paramTypes: baseTypes,
				argTypes:   baseTypes,
				variadic:   -1,
			},
		}, nil
	}
	// There are no exact matches, so now we'll look through all of the overloads to determine the best match. This is
	// much more work, but there's a performance penalty for runtime overload resolution in Postgres as well.
	if c.IsOperator {
		return c.resolveOperator(argTypes, overloads, fnOverloads)
	} else {
		return c.resolveFunction(argTypes, fnOverloads)
	}
}

// resolveOperator resolves an operator according to the rules defined by Postgres.
// https://www.postgresql.org/docs/15/typeconv-oper.html
func (c *CompiledFunction) resolveOperator(argTypes []pgtypes.DoltgresType, overloads *Overloads, fnOverloads []Overload) (overloadMatch, error) {
	// Binary operators treat unknown literals as the other type, so we'll account for that here to see if we can find
	// an "exact" match.
	if len(argTypes) == 2 {
		leftUnknownType := argTypes[0].BaseID() == pgtypes.DoltgresTypeBaseID_Unknown
		rightUnknownType := argTypes[1].BaseID() == pgtypes.DoltgresTypeBaseID_Unknown
		if (leftUnknownType && !rightUnknownType) || (!leftUnknownType && rightUnknownType) {
			var baseID pgtypes.DoltgresTypeBaseID
			casts := []TypeCastFunction{identityCast, identityCast}
			if leftUnknownType {
				casts[0] = UnknownLiteralCast
				baseID = argTypes[1].BaseID()
			} else {
				casts[1] = UnknownLiteralCast
				baseID = argTypes[0].BaseID()
			}
			if exactMatch, ok := overloads.ExactMatchForBaseIds(baseID, baseID); ok {
				return overloadMatch{
					params: Overload{
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
	return c.resolveFunction(argTypes, fnOverloads)
}

// resolveFunction resolves a function according to the rules defined by Postgres.
// https://www.postgresql.org/docs/15/typeconv-func.html
func (c *CompiledFunction) resolveFunction(argTypes []pgtypes.DoltgresType, overloads []Overload) (overloadMatch, error) {
	// First we'll discard all overloads that do not have implicitly-convertible param types
	compatibleOverloads := c.typeCompatibleOverloads(overloads, argTypes)

	// No compatible overloads available, return early
	if len(compatibleOverloads) == 0 {
		return overloadMatch{}, nil
	}

	// If we've found exactly one match then we'll return that one
	// TODO: we need to also prefer non-variadic functions here over variadic ones (no such conflict can exist for now)
	//  https://www.postgresql.org/docs/15/typeconv-func.html
	if len(compatibleOverloads) == 1 {
		return compatibleOverloads[0], nil
	}

	// Next rank the candidates by the number of params whose types match exactly
	closestMatches := c.closestTypeMatches(argTypes, compatibleOverloads)

	// Now check again for exactly one match
	if len(closestMatches) == 1 {
		return closestMatches[0], nil
	}

	// If there was more than a single match, try to find the one with the most preferred type conversions
	preferredOverloads := c.preferredTypeMatches(argTypes, closestMatches)

	// Check once more for exactly one match
	if len(preferredOverloads) == 1 {
		return preferredOverloads[0], nil
	}

	// Next we'll check the type categories for `unknown` types
	unknownOverloads, ok := c.unknownTypeCategoryMatches(argTypes, preferredOverloads)
	if !ok {
		return overloadMatch{}, nil
	}

	// Check again for exactly one match
	if len(unknownOverloads) == 1 {
		return unknownOverloads[0], nil
	}

	// No matching function overload found
	return overloadMatch{}, nil
}

// typeCompatibleOverloads returns all overloads that have a matching number of params whose types can be
// implicitly converted to the ones provided. This is the set of all possible overloads that could be used with the
// param types provided.
func (c *CompiledFunction) typeCompatibleOverloads(fnOverloads []Overload, argTypes []pgtypes.DoltgresType) []overloadMatch {
	var compatible []overloadMatch
	for _, overload := range fnOverloads {
		isConvertible := true
		overloadCasts := make([]TypeCastFunction, len(argTypes))
		// Polymorphic parameters must be gathered so that we can later verify that they all have matching base types
		var polymorphicParameters []pgtypes.DoltgresType
		var polymorphicTargets []pgtypes.DoltgresType
		for i := range argTypes {
			paramType := overload.argTypes[i]

			if polymorphicType, ok := paramType.GetRepresentativeType().(pgtypes.DoltgresPolymorphicType); ok && polymorphicType.IsValid(argTypes[i]) {
				overloadCasts[i] = identityCast
				polymorphicParameters = append(polymorphicParameters, polymorphicType)
				polymorphicTargets = append(polymorphicTargets, argTypes[i])
			} else {
				if overloadCasts[i] = GetImplicitCast(argTypes[i].BaseID(), paramType); overloadCasts[i] == nil {
					if argTypes[i].BaseID() == pgtypes.DoltgresTypeBaseID_Unknown {
						overloadCasts[i] = UnknownLiteralCast
					} else {
						isConvertible = false
						break
					}
				}
			}
		}

		if isConvertible && c.polymorphicTypesCompatible(polymorphicParameters, polymorphicTargets) {
			compatible = append(compatible, overloadMatch{params: overload, casts: overloadCasts})
		}
	}
	return compatible
}

// closestTypeMatches returns the set of overload candidates that have the most exact type matches for the arg types
// provided.
func (*CompiledFunction) closestTypeMatches(argTypes []pgtypes.DoltgresType, candidates []overloadMatch) []overloadMatch {
	matchCount := 0
	var matches []overloadMatch
	for _, cand := range candidates {
		currentMatchCount := 0
		for argIdx := range argTypes {
			argType := cand.params.argTypes[argIdx]

			argBaseId := argTypes[argIdx].BaseID()
			if argBaseId == argType || argBaseId == pgtypes.DoltgresTypeBaseID_Unknown {
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

// preferredTypeMatches returns the overload candidates that have the most preferred types for args that require casts.
func (*CompiledFunction) preferredTypeMatches(argTypes []pgtypes.DoltgresType, candidates []overloadMatch) []overloadMatch {
	preferredCount := 0
	var preferredOverloads []overloadMatch
	for _, cand := range candidates {
		currentPreferredCount := 0
		for argIdx := range argTypes {
			argType := cand.params.argTypes[argIdx]

			if argTypes[argIdx].BaseID() != argType && argType.GetTypeCategory().IsPreferredType(argType) {
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

// unknownTypeCategoryMatches checks the type categories of `unknown` types. These types have an inherent bias toward
// the string category since an `unknown` literal resembles a string. Returns false if the resolution should fail.
func (c *CompiledFunction) unknownTypeCategoryMatches(argTypes []pgtypes.DoltgresType, candidates []overloadMatch) ([]overloadMatch, bool) {
	matches := make([]overloadMatch, len(candidates))
	copy(matches, candidates)
	// For our first loop, we'll filter matches based on whether they accept the string category
	for argIdx := range argTypes {
		// We're only concerned with `unknown` types
		if argTypes[argIdx].BaseID() != pgtypes.DoltgresTypeBaseID_Unknown {
			continue
		}
		var newMatches []overloadMatch
		for _, match := range matches {
			if match.params.argTypes[argIdx].GetTypeCategory() == pgtypes.TypeCategory_StringTypes {
				newMatches = append(newMatches, match)
			}
		}
		// If we've found matches in this step, then we'll update our match set
		if len(newMatches) > 0 {
			matches = newMatches
		}
	}
	// Return early if we've filtered down to a single match
	if len(matches) == 1 {
		return matches, true
	}
	// TODO: implement the remainder of step 4.e. from the documentation (following code assumes it has been implemented)
	// ...

	// If we've discarded every function, then we'll actually return all of the original candidates
	if len(matches) == 0 {
		return candidates, true
	}
	// In this case, we've trimmed at least one candidate, so we'll return our new matches
	return matches, true
}

// polymorphicTypesCompatible returns whether any polymorphic types given are compatible with the expression types given
func (*CompiledFunction) polymorphicTypesCompatible(paramTypes []pgtypes.DoltgresType, exprTypes []pgtypes.DoltgresType) bool {
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
		if polymorphicParamType, ok := paramType.(pgtypes.DoltgresPolymorphicType); ok && exprTypes[i].BaseID() != pgtypes.DoltgresTypeBaseID_Unknown {
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
	// We can use the first polymorphic non-unknown type that we find, since we can morph it into any type that we need.
	// We've verified that all polymorphic types are compatible in a previous step, so this is safe to do.
	var firstPolymorphicType pgtypes.DoltgresType
	for i, functionInterfaceType := range functionInterfaceTypes {
		if _, ok = functionInterfaceType.(pgtypes.DoltgresPolymorphicType); ok && originalTypes[i].BaseID() != pgtypes.DoltgresTypeBaseID_Unknown {
			firstPolymorphicType = originalTypes[i]
			break
		}
	}

	// if all types are `unknown`, use `text` type
	if firstPolymorphicType == nil {
		firstPolymorphicType = pgtypes.Text
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
			case query.Type_DATE:
				args[i], _, _ = pgtypes.Date.Convert(args[i])
			case query.Type_DATETIME, query.Type_TIMESTAMP:
				args[i], _, _ = pgtypes.Timestamp.Convert(args[i])
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
func (c *CompiledFunction) analyzeParameters() (originalTypes []pgtypes.DoltgresType, err error) {
	originalTypes = make([]pgtypes.DoltgresType, len(c.Arguments))
	for i, param := range c.Arguments {
		returnType := param.Type()
		if extendedType, ok := returnType.(pgtypes.DoltgresType); ok {
			if domainType, ok := extendedType.(pgtypes.DomainType); ok {
				extendedType = domainType.UnderlyingBaseType()
			}
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
			case query.Type_NULL_TYPE:
				originalTypes[i] = pgtypes.Unknown
			default:
				return nil, fmt.Errorf("encountered a type that does not conform to the DoltgresType interface: %T", param.Type())
			}
		}
	}
	return originalTypes, nil
}
