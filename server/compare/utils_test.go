// Copyright 2026 Dolthub, Inc.
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

package compare_test

import (
	"sync"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/server/compare"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/functions/binary"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

var compareRecordsInitOnce sync.Once

func TestCompareRecordsUsesLexicographicRowComparison(t *testing.T) {
	initCompareRecordsTests()

	ctx := sql.NewEmptyContext()
	tests := []struct {
		name     string
		op       framework.Operator
		left     []pgtypes.RecordValue
		right    []pgtypes.RecordValue
		expected any
	}{
		{
			name:     "greater than false when first field is less",
			op:       framework.Operator_BinaryGreaterThan,
			left:     recordValues(1, 1),
			right:    recordValues(999, 999),
			expected: false,
		},
		{
			name:     "less than true when first field is less",
			op:       framework.Operator_BinaryLessThan,
			left:     recordValues(1, 1),
			right:    recordValues(999, 999),
			expected: true,
		},
		{
			name:     "greater than true when later field decides",
			op:       framework.Operator_BinaryGreaterThan,
			left:     recordValues(1, 2),
			right:    recordValues(1, 1),
			expected: true,
		},
		{
			name:     "less than or equal true for equal records",
			op:       framework.Operator_BinaryLessOrEqual,
			left:     recordValues(1, 1),
			right:    recordValues(1, 1),
			expected: true,
		},
		{
			name:     "greater than or equal true for equal records",
			op:       framework.Operator_BinaryGreaterOrEqual,
			left:     recordValues(1, 1),
			right:    recordValues(1, 1),
			expected: true,
		},
		{
			name:     "not equal false for equal records",
			op:       framework.Operator_BinaryNotEqual,
			left:     recordValues(1, 1),
			right:    recordValues(1, 1),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := compare.CompareRecords(ctx, test.op, test.left, test.right)
			require.NoError(t, err)
			assert.Equal(t, test.expected, res)
		})
	}
}

func TestCompareRecordsHandlesNullRowComparison(t *testing.T) {
	initCompareRecordsTests()

	ctx := sql.NewEmptyContext()
	tests := []struct {
		name     string
		op       framework.Operator
		left     []pgtypes.RecordValue
		right    []pgtypes.RecordValue
		expected any
	}{
		{
			name:     "equality false when later non null field differs",
			op:       framework.Operator_BinaryEqual,
			left:     recordValues(nil, 1),
			right:    recordValues(nil, 2),
			expected: false,
		},
		{
			name:     "not equal true when later non null field differs",
			op:       framework.Operator_BinaryNotEqual,
			left:     recordValues(nil, 1),
			right:    recordValues(nil, 2),
			expected: true,
		},
		{
			name:     "equality unknown when only null fields are inconclusive",
			op:       framework.Operator_BinaryEqual,
			left:     recordValues(1, nil),
			right:    recordValues(1, nil),
			expected: nil,
		},
		{
			name:     "ordering unknown when decisive field contains null",
			op:       framework.Operator_BinaryGreaterThan,
			left:     recordValues(1, 2),
			right:    recordValues(1, nil),
			expected: nil,
		},
		{
			name:     "ordering true when earlier non null field decides before null",
			op:       framework.Operator_BinaryGreaterThan,
			left:     recordValues(2, 1),
			right:    recordValues(1, nil),
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := compare.CompareRecords(ctx, test.op, test.left, test.right)
			require.NoError(t, err)
			assert.Equal(t, test.expected, res)
		})
	}
}

func initCompareRecordsTests() {
	compareRecordsInitOnce.Do(func() {
		core.Init()
		pgtypes.Init()
		functions.Init()
		binary.Init()
		framework.Initialize()
	})
}

func recordValues(values ...any) []pgtypes.RecordValue {
	record := make([]pgtypes.RecordValue, len(values))
	for i, value := range values {
		if intValue, ok := value.(int); ok {
			value = int32(intValue)
		}
		record[i] = pgtypes.RecordValue{
			Type:  pgtypes.Int32,
			Value: value,
		}
	}
	return record
}
