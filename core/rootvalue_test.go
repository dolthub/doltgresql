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

package core

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/stretchr/testify/require"
)

// TestGetForeignKeysForTable verifies that GetForeignKeysForTable matches KeysForTable on a copy of the root's
// collection, is unaffected by changes to such copies, is safe for concurrent use, and doesn't cost more as the root
// gains foreign keys on unrelated tables.
func TestGetForeignKeysForTable(t *testing.T) {
	ctx := context.Background()
	empty := newTestRootWithSchema(t, ctx)
	parent := doltdb.TableName{Name: "parent", Schema: "public"}
	declared, referencedBy, err := empty.GetForeignKeysForTable(ctx, parent)
	require.NoError(t, err)
	require.Empty(t, declared)
	require.Empty(t, referencedBy)

	// 400 tables each declare a key referencing |parent|, which declares a key of its own.
	newKey := func(name string, table doltdb.TableName, tag uint64, referenced doltdb.TableName) doltdb.ForeignKey {
		return doltdb.ForeignKey{
			Name:                   name,
			TableName:              table,
			TableColumns:           []uint64{tag},
			ReferencedTableName:    referenced,
			ReferencedTableColumns: []uint64{1},
		}
	}
	fkc, err := doltdb.NewForeignKeyCollection()
	require.NoError(t, err)
	for i := 0; i < 400; i++ {
		child := doltdb.TableName{Name: fmt.Sprintf("child_%04d", i), Schema: "public"}
		require.NoError(t, fkc.AddKeys(newKey(fmt.Sprintf("fk_%04d", i), child, uint64(i+10), parent)))
	}
	grandparent := doltdb.TableName{Name: "grandparent", Schema: "public"}
	require.NoError(t, fkc.AddKeys(newKey("parent_fk", parent, 2, grandparent)))
	root, err := empty.PutForeignKeyCollection(ctx, fkc)
	require.NoError(t, err)

	child := doltdb.TableName{Name: "child_0007", Schema: "public"}
	lookups := []doltdb.TableName{parent, child, grandparent, {Name: "missing", Schema: "public"}}
	requireMatchesCollection := func() {
		copied, err := root.GetForeignKeyCollection(ctx)
		require.NoError(t, err)
		for _, tableName := range lookups {
			expectedDeclared, expectedReferencedBy := copied.KeysForTable(tableName)
			declared, referencedBy, err := root.GetForeignKeysForTable(ctx, tableName)
			require.NoError(t, err)
			require.Equal(t, expectedDeclared, declared, "table %s", tableName)
			require.Equal(t, expectedReferencedBy, referencedBy, "table %s", tableName)
		}
	}

	// Sessions sharing a root may load its collection, and look keys up in it, concurrently.
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, referencedBy, err := root.GetForeignKeysForTable(ctx, parent)
			require.NoError(t, err)
			require.Len(t, referencedBy, 400)
		}()
	}
	wg.Wait()
	requireMatchesCollection()

	// Changing a collection returned by GetForeignKeyCollection must not change the keys this root returns.
	copied, err := root.GetForeignKeyCollection(ctx)
	require.NoError(t, err)
	require.NoError(t, copied.RemoveTables(ctx, child))
	declared, _, err = root.GetForeignKeysForTable(ctx, child)
	require.NoError(t, err)
	require.Len(t, declared, 1)
	requireMatchesCollection()

	// Planning a write to a table looks up the keys of every table that references it, so a lookup must not cost more
	// as the root gains keys on other tables. Copying this root's collection alone takes hundreds of allocations.
	allocs := testing.AllocsPerRun(10, func() {
		_, _, err := root.GetForeignKeysForTable(ctx, child)
		require.NoError(t, err)
	})
	require.LessOrEqual(t, allocs, 10.0)
}
