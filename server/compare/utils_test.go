package compare_test

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/assert"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/server/compare"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/functions/binary"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

func TestCompareRecords(t *testing.T) {
	core.Init()
	pgtypes.Init()
	binary.Init()
	functions.Init()
	framework.Initialize()
	ctx := sql.NewEmptyContext()

	// (1, 1) > (999, 999) should be false
	v1 := []pgtypes.RecordValue{
		{Value: int32(1), Type: pgtypes.Int32},
		{Value: int32(1), Type: pgtypes.Int32},
	}
	v2 := []pgtypes.RecordValue{
		{Value: int32(999), Type: pgtypes.Int32},
		{Value: int32(999), Type: pgtypes.Int32},
	}

	res, err := compare.CompareRecords(ctx, framework.Operator_BinaryGreaterThan, v1, v2)
	assert.NoError(t, err)
	assert.Equal(t, false, res, "(1, 1) > (999, 999) should be false")

	// (1, 1) < (999, 999) should be true
	res, err = compare.CompareRecords(ctx, framework.Operator_BinaryLessThan, v1, v2)
	assert.NoError(t, err)
	assert.Equal(t, true, res, "(1, 1) < (999, 999) should be true")

	// (999, 999) > (1, 1) should be true
	res, err = compare.CompareRecords(ctx, framework.Operator_BinaryGreaterThan, v2, v1)
	assert.NoError(t, err)
	assert.Equal(t, true, res, "(999, 999) > (1, 1) should be true")

	// (1, 1) = (1, 1) should be true
	res, err = compare.CompareRecords(ctx, framework.Operator_BinaryEqual, v1, v1)
	assert.NoError(t, err)
	assert.Equal(t, true, res, "(1, 1) = (1, 1) should be true")

	// (1, 1) > (1, 1) should be false
	res, err = compare.CompareRecords(ctx, framework.Operator_BinaryGreaterThan, v1, v1)
	assert.NoError(t, err)
	assert.Equal(t, false, res, "(1, 1) > (1, 1) should be false")

	// (1, 1) >= (1, 1) should be true
	res, err = compare.CompareRecords(ctx, framework.Operator_BinaryGreaterOrEqual, v1, v1)
	assert.NoError(t, err)
	assert.Equal(t, true, res, "(1, 1) >= (1, 1) should be true")

	// (1, 1) <= (1, 1) should be true
	res, err = compare.CompareRecords(ctx, framework.Operator_BinaryLessOrEqual, v1, v1)
	assert.NoError(t, err)
	assert.Equal(t, true, res, "(1, 1) <= (1, 1) should be true")

	// (1, 1) < (1, 1) should be false
	res, err = compare.CompareRecords(ctx, framework.Operator_BinaryLessThan, v1, v1)
	assert.NoError(t, err)
	assert.Equal(t, false, res, "(1, 1) < (1, 1) should be false")

	// (1, 1) != (1, 1) should be false
	res, err = compare.CompareRecords(ctx, framework.Operator_BinaryNotEqual, v1, v1)
	assert.NoError(t, err)
	assert.Equal(t, false, res, "(1, 1) != (1, 1) should be false")

	// (1, 2) > (1, 1) should be true
	v1_2 := []pgtypes.RecordValue{
		{Value: int32(1), Type: pgtypes.Int32},
		{Value: int32(2), Type: pgtypes.Int32},
	}
	res, err = compare.CompareRecords(ctx, framework.Operator_BinaryGreaterThan, v1_2, v1)
	assert.NoError(t, err)
	assert.Equal(t, true, res, "(1, 2) > (1, 1) should be true")

	// NULL cases
	// (1, NULL) = (1, NULL) should be NULL (nil)
	v1_null := []pgtypes.RecordValue{
		{Value: int32(1), Type: pgtypes.Int32},
		{Value: nil, Type: pgtypes.Int32},
	}
	res, err = compare.CompareRecords(ctx, framework.Operator_BinaryEqual, v1_null, v1_null)
	assert.NoError(t, err)
	assert.Nil(t, res, "(1, NULL) = (1, NULL) should be NULL")

	// (1, 2) > (1, NULL) should be NULL
	res, err = compare.CompareRecords(ctx, framework.Operator_BinaryGreaterThan, v1_2, v1_null)
	assert.NoError(t, err)
	assert.Nil(t, res, "(1, 2) > (1, NULL) should be NULL")

	// (2, 1) > (1, NULL) should be true (first element determines result)
	v2_1 := []pgtypes.RecordValue{
		{Value: int32(2), Type: pgtypes.Int32},
		{Value: int32(1), Type: pgtypes.Int32},
	}
	res, err = compare.CompareRecords(ctx, framework.Operator_BinaryGreaterThan, v2_1, v1_null)
	assert.NoError(t, err)
	assert.Equal(t, true, res, "(2, 1) > (1, NULL) should be true")
}
