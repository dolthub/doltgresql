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

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Operator is a unary or binary operator.
type Operator byte

const (
	Operator_BinaryPlus                Operator = iota // +
	Operator_BinaryMinus                               // -
	Operator_BinaryMultiply                            // *
	Operator_BinaryDivide                              // /
	Operator_BinaryMod                                 // %
	Operator_BinaryShiftLeft                           // <<
	Operator_BinaryShiftRight                          // >>
	Operator_BinaryLessThan                            // <
	Operator_BinaryGreaterThan                         // >
	Operator_BinaryLessOrEqual                         // <=
	Operator_BinaryGreaterOrEqual                      // >=
	Operator_BinaryEqual                               // =
	Operator_BinaryNotEqual                            // <> or != (they're equivalent in all cases)
	Operator_BinaryBitAnd                              // &
	Operator_BinaryBitOr                               // |
	Operator_BinaryBitXor                              // ^
	Operator_BinaryConcatenate                         // ||
	Operator_BinaryJSONExtractJson                     // ->
	Operator_BinaryJSONExtractText                     // ->>
	Operator_BinaryJSONExtractPathJson                 // #>
	Operator_BinaryJSONExtractPathText                 // #>>
	Operator_BinaryJSONContainsRight                   // @>
	Operator_BinaryJSONContainsLeft                    // <@
	Operator_BinaryJSONTopLevel                        // ?
	Operator_BinaryJSONTopLevelAny                     // ?|
	Operator_BinaryJSONTopLevelAll                     // ?&
	Operator_UnaryPlus                                 // +
	Operator_UnaryMinus                                // -
)

// unaryFunction represents the signature for a unary function.
type unaryFunction struct {
	Operator Operator
	Type     pgtypes.DoltgresTypeBaseID
}

// binaryFunction represents the signature for a binary function.
type binaryFunction struct {
	Operator Operator
	Left     pgtypes.DoltgresTypeBaseID
	Right    pgtypes.DoltgresTypeBaseID
}

var (
	// unaryFunctions is a map from a unaryFunction signature to the associated function.
	unaryFunctions = map[unaryFunction]Function1{}
	// binaryFunctions is a map from a binaryFunction signature to the associated function.
	binaryFunctions = map[binaryFunction]Function2{}
	// unaryAggregateOverloads is a map from an operator to an Overload deducer that is the aggregate of all functions
	// for that operator.
	unaryAggregateOverloads = map[Operator]*Overloads{}
	// binaryAggregateOverloads is a map from an operator to an Overload deducer that is the aggregate of all functions
	// for that operator.
	binaryAggregateOverloads = map[Operator]*Overloads{}
	// unaryAggregatePermutations contains all of the permutations for each unary operator.
	unaryAggregatePermutations = map[Operator][]Overload{}
	// unaryAggregatePermutations contains all of the permutations for each binary operator.
	binaryAggregatePermutations = map[Operator][]Overload{}
)

// RegisterUnaryFunction registers the given function, so that it will be usable from a running server. This should
// only be used for unary functions, which are the underlying functions for unary operators such as negation, etc. This
// should be called from within an init().
func RegisterUnaryFunction(operator Operator, f Function1) {
	if !operator.IsUnary() {
		panic("non-unary operator: " + operator.String())
	}
	RegisterFunction(f)
	sig := unaryFunction{
		Operator: operator,
		Type:     f.Parameters[0].BaseID(),
	}
	if existingFunction, ok := unaryFunctions[sig]; ok {
		panic(fmt.Errorf("duplicate unary function for `%s`: `%s` and `%s`",
			operator.String(), existingFunction.Name, f.Name))
	}
	unaryFunctions[sig] = f
}

// RegisterBinaryFunction registers the given function, so that it will be usable from a running server. This should
// only be used for binary functions, which are the underlying functions for binary operators such as addition,
// subtraction, etc. This should be called from within an init().
func RegisterBinaryFunction(operator Operator, f Function2) {
	if !operator.IsBinary() {
		panic("non-binary operator: " + operator.String())
	}
	RegisterFunction(f)
	sig := binaryFunction{
		Operator: operator,
		Left:     f.Parameters[0].BaseID(),
		Right:    f.Parameters[1].BaseID(),
	}
	if existingFunction, ok := binaryFunctions[sig]; ok {
		panic(fmt.Errorf("duplicate binary function for `%s`: `%s` and `%s`",
			operator.String(), existingFunction.Name, f.Name))
	}
	binaryFunctions[sig] = f
}

// GetUnaryFunction returns the unary function that matches the given operator.
func GetUnaryFunction(operator Operator) IntermediateFunction {
	// Returns nil if not found, which is fine as IntermediateFunction will handle the nil deducer
	return IntermediateFunction{
		Functions:    unaryAggregateOverloads[operator],
		AllOverloads: unaryAggregatePermutations[operator],
		IsOperator:   true,
	}
}

// GetBinaryFunction returns the binary function that matches the given operator.
func GetBinaryFunction(operator Operator) IntermediateFunction {
	// Returns nil if not found, which is fine as IntermediateFunction will handle the nil deducer
	return IntermediateFunction{
		Functions:    binaryAggregateOverloads[operator],
		AllOverloads: binaryAggregatePermutations[operator],
		IsOperator:   true,
	}
}

// String returns the string form of the operator.
func (o Operator) String() string {
	switch o {
	case Operator_BinaryPlus, Operator_UnaryPlus:
		return "+"
	case Operator_BinaryMinus, Operator_UnaryMinus:
		return "-"
	case Operator_BinaryMultiply:
		return "*"
	case Operator_BinaryDivide:
		return "/"
	case Operator_BinaryMod:
		return "%"
	case Operator_BinaryShiftLeft:
		return "<<"
	case Operator_BinaryShiftRight:
		return ">>"
	case Operator_BinaryBitAnd:
		return "&"
	case Operator_BinaryBitOr:
		return "|"
	case Operator_BinaryBitXor:
		return "#"
	case Operator_BinaryConcatenate:
		return "||"
	case Operator_BinaryEqual:
		return "="
	case Operator_BinaryNotEqual:
		return "<>"
	case Operator_BinaryGreaterThan:
		return ">"
	case Operator_BinaryGreaterOrEqual:
		return ">="
	case Operator_BinaryLessThan:
		return "<"
	case Operator_BinaryLessOrEqual:
		return "<="
	default:
		return "unknown operator"
	}
}

// IsUnary returns whether the operator is a unary operator.
func (o Operator) IsUnary() bool {
	switch o {
	case Operator_UnaryPlus, Operator_UnaryMinus:
		return true
	default:
		return false
	}
}

// IsBinary returns whether the operator is a binary operator.
func (o Operator) IsBinary() bool {
	return !o.IsUnary()
}

// GetOperatorFromString returns the binary operator for the given subOperator.
func GetOperatorFromString(op string) (Operator, error) {
	switch op {
	case "=":
		return Operator_BinaryEqual, nil
	case "<>", "!=":
		return Operator_BinaryNotEqual, nil
	case "<":
		return Operator_BinaryLessThan, nil
	case "<=":
		return Operator_BinaryLessOrEqual, nil
	case ">":
		return Operator_BinaryGreaterThan, nil
	case ">=":
		return Operator_BinaryGreaterOrEqual, nil
	default:
		return 0, fmt.Errorf("unhandled Operator `%s`", op)
	}
}
