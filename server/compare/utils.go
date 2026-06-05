// Copyright 2025 Dolthub, Inc.
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

package compare

import (
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// CompareRecords compares two record values using the specified comparison operator, |op| and returns a result
// indicating if the comparison was true, false, or indeterminate (nil).
//
// More info on rules for comparing records:
// https://www.postgresql.org/docs/current/functions-comparisons.html#ROW-WISE-COMPARISON
func CompareRecords(ctx *sql.Context, op framework.Operator, v1 interface{}, v2 interface{}) (res any, err error) {
	leftRecord, rightRecord, err := checkRecordArgs(v1, v2)
	if err != nil {
		return nil, err
	}

	// TODO: This can be a hot path when filtering a large table against a tuple, for example.
	//  We can reduce branching by splitting each case into their individual functions, which will also make the code
	//  more readable. A downside would be a ton of repeated code.
	var hasNull bool
	var leftLiteral, rightLiteral expression.Literal
	for i := 0; i < len(leftRecord); i++ {
		// NULL values are by definition not comparable
		if leftRecord[i].Value == nil || rightRecord[i].Value == nil {
			switch op {
			case framework.Operator_BinaryEqual, framework.Operator_BinaryNotEqual:
				hasNull = true
				continue
			default:
				return nil, nil
			}
		}
		leftLiteral.Val = leftRecord[i].Value
		leftLiteral.Typ = leftRecord[i].Type
		rightLiteral.Val = rightRecord[i].Value
		rightLiteral.Typ = rightRecord[i].Type

		switch op {
		case framework.Operator_BinaryEqual:
			res, err = callComparisonFunction(ctx, framework.Operator_BinaryEqual, &leftLiteral, &rightLiteral)
			if err != nil {
				return false, err
			}
			if res == false {
				return false, nil
			}
		case framework.Operator_BinaryNotEqual:
			res, err = callComparisonFunction(ctx, framework.Operator_BinaryNotEqual, &leftLiteral, &rightLiteral)
			if err != nil {
				return false, err
			}
			if res == true {
				return true, nil
			}
		// Records are compared by evaluating each field in order of significance, so if the fields are equal, we need
		// to compare the next field.
		case framework.Operator_BinaryLessThan, framework.Operator_BinaryLessOrEqual,
			framework.Operator_BinaryGreaterThan, framework.Operator_BinaryGreaterOrEqual:
			switch op {
			case framework.Operator_BinaryLessThan, framework.Operator_BinaryLessOrEqual:
				res, err = callComparisonFunction(ctx, framework.Operator_BinaryLessThan, &leftLiteral, &rightLiteral)
			default:
				res, err = callComparisonFunction(ctx, framework.Operator_BinaryGreaterThan, &leftLiteral, &rightLiteral)
			}
			if err != nil {
				return false, err
			}
			if res == true {
				return true, nil
			}
			// If equals fields are equal, move onto the next field, else return false
			res, err = callComparisonFunction(ctx, framework.Operator_BinaryEqual, &leftLiteral, &rightLiteral)
			if err != nil {
				return false, err
			}
			if res == false {
				return false, nil
			}
		default:
			return false, fmt.Errorf("unsupported binary operator: %s", op)
		}
	}

	if hasNull {
		return nil, nil
	}

	// Every field is equal
	switch op {
	case framework.Operator_BinaryEqual, framework.Operator_BinaryLessOrEqual, framework.Operator_BinaryGreaterOrEqual:
		return true, nil
	default:
		return false, nil
	}
}

// checkRecordArgs asserts that |v1| and |v2| are both []pgtypes.RecordValue, and that they have the same number of
// elements, then returns them. If any problems were detected, an error is returned instead.
func checkRecordArgs(v1, v2 interface{}) (leftRecord, rightRecord []pgtypes.RecordValue, err error) {
	var ok bool
	leftRecord, ok = v1.([]pgtypes.RecordValue)
	if !ok {
		return nil, nil, fmt.Errorf("expected a RecordValue, but got %T", v1)
	}
	rightRecord, ok = v2.([]pgtypes.RecordValue)
	if !ok {
		return nil, nil, fmt.Errorf("expected a RecordValue, but got %T", v2)
	}
	if len(leftRecord) != len(rightRecord) {
		return nil, nil, fmt.Errorf("unequal number of entries in row expressions")
	}

	return leftRecord, rightRecord, nil
}

// callComparisonFunction invokes the binary comparison function for the specified operator |op| with the two arguments
// |leftLiteral| and |rightLiteral|. The result and any error are returned.
func callComparisonFunction(ctx *sql.Context, op framework.Operator, leftLiteral, rightLiteral sql.Expression) (result any, err error) {
	intermediateFunction := framework.GetBinaryFunction(op)
	compiledFunction := intermediateFunction.Compile(
		ctx, "_internal_record_comparison_function", leftLiteral, rightLiteral)
	return compiledFunction.Eval(ctx, nil)
}
