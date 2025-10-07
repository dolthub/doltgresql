package pgcatalog

import (
	"testing"

	assert "github.com/stretchr/testify/require"
)

func TestIndexes(t *testing.T) {
	// Only the fields that impact indexing are set here
	con1 := &pgConstraint{
		oidNative:       100,
		name:            "abc",
		tableOidNative:  300,
		typeOidNative:   0, // this means the constraint is on a table, not a type
		schemaOidNative: 500,
	}

	relidTypNameIdx := NewUniqueInMemIndexStorage[*pgConstraint](lessConstraintRelidTypeName)

	relidTypNameIdx.Add(con1)

	var foundElements []*pgConstraint
	cb := func(c *pgConstraint) bool {
		foundElements = append(foundElements, c)
		return true
	}

	// lookup by relid
	relidTypNameIdx.uniqTree.AscendRange(
		&pgConstraint{tableOidNative: 300},
		&pgConstraint{tableOidNative: 301},
		cb,
	)

	assert.Equal(t, []*pgConstraint{con1}, foundElements)

	foundElements = nil
	// lookup by relid, typeid
	relidTypNameIdx.uniqTree.AscendRange(
		&pgConstraint{tableOidNative: 300, typeOidNative: 0},
		&pgConstraint{tableOidNative: 300, typeOidNative: 1},
		cb,
	)
	
	assert.Equal(t, []*pgConstraint{con1}, foundElements)
}
