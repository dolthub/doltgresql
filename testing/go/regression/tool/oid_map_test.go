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

package main

import (
	"testing"

	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOIDMapRewriteQuery(t *testing.T) {
	om := NewOIDMap()
	om.LearnAll(map[uint32]uint32{159776: 21005, 159780: 21009})
	assert.Equal(t, `SELECT 1 WHERE c.oid = '21005'`, om.RewriteQuery(`SELECT 1 WHERE c.oid = '159776'`))
	assert.Equal(t, `SELECT 1 WHERE c.oid = 21005 AND i.indexrelid = 21009`,
		om.RewriteQuery(`SELECT 1 WHERE c.oid = 159776 AND i.indexrelid = 159780`))
	// Unknown OIDs, short numbers, and numbers embedded in identifiers are untouched
	assert.Equal(t, `SELECT 159777, 1597, col159776, tbl_159776x FROM t`,
		om.RewriteQuery(`SELECT 159777, 1597, col159776, tbl_159776x FROM t`))
	// A replacement value is never rescanned as another recorded OID
	om2 := NewOIDMap()
	om2.LearnAll(map[uint32]uint32{100001: 100002, 100002: 100003})
	assert.Equal(t, `100002 100003`, om2.RewriteQuery(`100001 100002`))
	// The query is returned as-is when the map is empty
	empty := NewOIDMap()
	assert.Equal(t, `SELECT '159776'`, empty.RewriteQuery(`SELECT '159776'`))
}

func TestOIDMapCompareLearning(t *testing.T) {
	oidDesc := &pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{
		{Name: []byte("oid"), DataTypeOID: pgtype.OIDOID, Format: 0},
		{Name: []byte("relname"), DataTypeOID: pgtype.TextOID, Format: 0},
	}}
	rows := func(vals ...[2]string) []*pgproto3.DataRow {
		out := make([]*pgproto3.DataRow, len(vals))
		for i, v := range vals {
			out[i] = &pgproto3.DataRow{Values: [][]byte{[]byte(v[0]), []byte(v[1])}}
		}
		return out
	}

	// Ordered comparison learns a new user-OID pairing when everything else matches
	om := NewOIDMap()
	require.NoError(t, CompareRowsOrdered(om,
		oidDesc, oidDesc,
		rows([2]string{"159776", "attmp"}),
		rows([2]string{"21005", "attmp"})))
	mapped, ok := om.Get(159776)
	require.True(t, ok)
	assert.Equal(t, uint32(21005), mapped)

	// A learned mapping must stay consistent within a result set
	require.Error(t, CompareRowsOrdered(om,
		oidDesc, oidDesc,
		rows([2]string{"159776", "a"}, [2]string{"159776", "b"}),
		rows([2]string{"21005", "a"}, [2]string{"31000", "b"})))

	// Non-OID differences still fail, and nothing is learned from a failed comparison
	om2 := NewOIDMap()
	require.Error(t, CompareRowsOrdered(om2,
		oidDesc, oidDesc,
		rows([2]string{"159776", "attmp"}),
		rows([2]string{"21005", "other"})))
	_, ok = om2.Get(159776)
	assert.False(t, ok)

	// System OIDs (below 16384) are never treated as equal when they differ
	require.Error(t, CompareRowsOrdered(NewOIDMap(),
		oidDesc, oidDesc,
		rows([2]string{"1259", "pg_class"}),
		rows([2]string{"1260", "pg_class"})))

	// Unordered comparison learns pairings too, keyed off the non-OID columns
	om3 := NewOIDMap()
	require.NoError(t, CompareRowsUnordered(om3,
		oidDesc, oidDesc,
		rows([2]string{"159776", "a"}, [2]string{"159780", "b"}),
		rows([2]string{"21009", "b"}, [2]string{"21005", "a"})))
	mapped, ok = om3.Get(159776)
	require.True(t, ok)
	assert.Equal(t, uint32(21005), mapped)
	mapped, ok = om3.Get(159780)
	require.True(t, ok)
	assert.Equal(t, uint32(21009), mapped)

	// Unordered comparison uses previously learned mappings for translation
	require.NoError(t, CompareRowsUnordered(om3,
		oidDesc, oidDesc,
		rows([2]string{"159776", "a"}),
		rows([2]string{"21005", "a"})))
}
