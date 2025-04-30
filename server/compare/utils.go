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
func CompareRecords(ctx *sql.Context, op framework.Operator, v1 interface{}, v2 interface{}) (result any, err error) {
	leftRecord, rightRecord, err := checkRecordArgs(v1, v2)
	if err != nil {
		return nil, err
	}

	hasNull := false
	hasEqualFields := true

	for i := 0; i < len(leftRecord); i++ {
		typ1 := leftRecord[i].Type
		typ2 := rightRecord[i].Type

		// NULL values are by definition not comparable, so they need special handling depending
		// on what type of comparison we're performing.
		if leftRecord[i].Value == nil || rightRecord[i].Value == nil {
			switch op {
			case framework.Operator_BinaryEqual:
				// If we're comparing for equality, then any presence of a NULL value means
				// we don't have enough information to determine equality, so return nil.
				return nil, nil

			case framework.Operator_BinaryLessThan, framework.Operator_BinaryGreaterThan,
				framework.Operator_BinaryLessOrEqual, framework.Operator_BinaryGreaterOrEqual:
				// If we haven't seen a prior field with non-equivalent values, then we
				// don't have enough certainty to make a comparison, so return nil.
				if hasEqualFields {
					return nil, nil
				}
			}

			// Otherwise, mark that we've seen a NULL and skip over it
			hasNull = true
			continue
		}

		leftLiteral := expression.NewLiteral(leftRecord[i].Value, typ1)
		rightLiteral := expression.NewLiteral(rightRecord[i].Value, typ2)

		// For >= and <=, we need to distinguish between < and = (and > and =). Records
		// are compared by evaluating each field, in order of significance, so for >= if
		// the field is greater than, then we can stop comparing and return true immediately.
		// If the field is equal, then we need to look at the next field. For this reason,
		// we have to break >= and <= into separate comparisons for > or < and =.
		switch op {
		case framework.Operator_BinaryLessThan, framework.Operator_BinaryLessOrEqual:
			if res, err := callComparisonFunction(ctx, framework.Operator_BinaryLessThan, leftLiteral, rightLiteral); err != nil {
				return false, err
			} else if res == true {
				return true, nil
			}
		case framework.Operator_BinaryGreaterThan, framework.Operator_BinaryGreaterOrEqual:
			if res, err := callComparisonFunction(ctx, framework.Operator_BinaryGreaterThan, leftLiteral, rightLiteral); err != nil {
				return false, err
			} else if res == true {
				return true, nil
			}
		}

		// After we've determined > and <, we can look at the equality comparison. For < and >, we've already returned
		// true if that initial comparison was true. Now we need to determine if the two fields are equal, in which case
		// we continue on to check the next field. If the two fields are NOT equal, then we can return false immediately.
		switch op {
		case framework.Operator_BinaryGreaterOrEqual, framework.Operator_BinaryLessOrEqual, framework.Operator_BinaryEqual:
			if res, err := callComparisonFunction(ctx, framework.Operator_BinaryEqual, leftLiteral, rightLiteral); err != nil {
				return false, err
			} else if res == false {
				return false, nil
			}
		case framework.Operator_BinaryNotEqual:
			if res, err := callComparisonFunction(ctx, framework.Operator_BinaryNotEqual, leftLiteral, rightLiteral); err != nil {
				return false, err
			} else if res == true {
				return true, nil
			}
		case framework.Operator_BinaryLessThan, framework.Operator_BinaryGreaterThan:
			if res, err := callComparisonFunction(ctx, framework.Operator_BinaryEqual, leftLiteral, rightLiteral); err != nil {
				return false, err
			} else if res == false {
				// For < and >, we still need to check equality to know if we need to continue checking additional fields.
				// If
				hasEqualFields = false
			}
		default:
			return false, fmt.Errorf("unsupportd binary operator: %s", op)
		}
	}

	// If the records contain any NULL fields, but all non-NULL fields are equal, then we
	// don't have enough certainty to return a result.
	if hasNull && hasEqualFields {
		return nil, nil
	}

	return true, nil
}

// checkRecordArgs asserts that |v1| and |v2| are both []pgtypes.RecordValue, and that they have the same number of
// elements, then returns them. If any problems were detected, an error is returnd instead.
func checkRecordArgs(v1, v2 interface{}) (leftRecord, rightRecord []pgtypes.RecordValue, err error) {
	leftRecord, ok1 := v1.([]pgtypes.RecordValue)
	rightRecord, ok2 := v2.([]pgtypes.RecordValue)
	if !ok1 {
		return nil, nil, fmt.Errorf("expected a RecordValue, but got %T", v1)
	} else if !ok2 {
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
		"_internal_record_comparison_function", leftLiteral, rightLiteral)
	return compiledFunction.Eval(ctx, nil)
}
