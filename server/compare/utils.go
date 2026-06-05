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

	for i := 0; i < len(leftRecord); i++ {
		typ1 := leftRecord[i].Type
		typ2 := rightRecord[i].Type

		// NULL values are by definition not comparable, so they need special handling depending
		// on what type of comparison we're performing.
		if leftRecord[i].Value == nil || rightRecord[i].Value == nil {
			switch op {
			case framework.Operator_BinaryLessThan, framework.Operator_BinaryGreaterThan,
				framework.Operator_BinaryLessOrEqual, framework.Operator_BinaryGreaterOrEqual:
				// The first non-equal field determines row ordering. A NULL at that
				// position makes the comparison unknown.
				return nil, nil
			}

			// Equality and inequality can still be determined if a later non-NULL
			// field pair is unequal.
			continue
		}

		leftLiteral := expression.NewLiteral(leftRecord[i].Value, typ1)
		rightLiteral := expression.NewLiteral(rightRecord[i].Value, typ2)

		equal, err := callComparisonFunction(ctx, framework.Operator_BinaryEqual, leftLiteral, rightLiteral)
		if err != nil {
			return false, err
		}
		if equal == true {
			continue
		}

		switch op {
		case framework.Operator_BinaryEqual:
			return false, nil
		case framework.Operator_BinaryNotEqual:
			return true, nil
		case framework.Operator_BinaryLessThan, framework.Operator_BinaryLessOrEqual:
			if res, err := callComparisonFunction(ctx, framework.Operator_BinaryLessThan, leftLiteral, rightLiteral); err != nil {
				return false, err
			} else {
				return res, nil
			}
		case framework.Operator_BinaryGreaterThan, framework.Operator_BinaryGreaterOrEqual:
			if res, err := callComparisonFunction(ctx, framework.Operator_BinaryGreaterThan, leftLiteral, rightLiteral); err != nil {
				return false, err
			} else {
				return res, nil
			}
		default:
			return false, fmt.Errorf("unsupported binary operator: %s", op)
		}
	}

	// If no non-NULL field pair determined the result and at least one field
	// pair is NULL, the comparison is unknown.
	if recordsContainNull(leftRecord, rightRecord) {
		return nil, nil
	}

	switch op {
	case framework.Operator_BinaryEqual, framework.Operator_BinaryLessOrEqual, framework.Operator_BinaryGreaterOrEqual:
		return true, nil
	case framework.Operator_BinaryNotEqual, framework.Operator_BinaryLessThan, framework.Operator_BinaryGreaterThan:
		return false, nil
	default:
		return false, fmt.Errorf("unsupported binary operator: %s", op)
	}
}

// recordsContainNull returns whether either record has a NULL value in any field.
func recordsContainNull(leftRecord, rightRecord []pgtypes.RecordValue) bool {
	for i := 0; i < len(leftRecord); i++ {
		if leftRecord[i].Value == nil || rightRecord[i].Value == nil {
			return true
		}
	}
	return false
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
	if compiledFunction == nil {
		return nil, fmt.Errorf("could not find comparison function for operator %s and types %s, %s",
			op, leftLiteral.Type(ctx).String(), rightLiteral.Type(ctx).String())
	}
	return compiledFunction.Eval(ctx, nil)
}
