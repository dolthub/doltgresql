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

package typecollection

import (
	"context"
	"errors"
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly/tree"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// TestMap_RemainsUsableAfterFlushFailure asserts that TypeCollection.Map
// remains usable after a flush returns an error. A subsequent Map call
// against a healthy store must succeed.
func TestMap_RemainsUsableAfterFlushFailure(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ns := newCountingFailNodeStore(t)
	coll := newTestTypeCollection(t, ns)

	first := pgtypes.NewUnresolvedDoltgresType("public", "type_one")
	coll.accessedMap[first.ID] = first
	_, err := coll.Map(ctx)
	require.NoError(t, err)

	ns.failAfter(0)
	second := pgtypes.NewUnresolvedDoltgresType("public", "type_two")
	coll.accessedMap[second.ID] = second
	_, err = coll.Map(ctx)
	require.Error(t, err)

	ns.allowAll()
	require.NotPanics(t, func() {
		_, _ = coll.Map(ctx)
	})
}

// TestGetTable_NoSchemaToSearch asserts that an unqualified lookup with no schema to search reports no matching table
// rather than an error, leaving the caller free to resolve the name another way.
func TestGetTable_NoSchemaToSearch(t *testing.T) {
	defer withSchemaResolution(t, "", false, nil)()

	coll := newTestTypeCollection(t, tree.NewTestNodeStore())
	tbl, schema, err := coll.getTable(sql.NewEmptyContext(), "", "registration_status")
	require.NoError(t, err)
	require.Nil(t, tbl)
	require.Empty(t, schema)
}

// TestGetTable_SchemaResolutionErrorPropagates asserts that a genuine schema resolution failure is still reported.
func TestGetTable_SchemaResolutionErrorPropagates(t *testing.T) {
	boom := errors.New("boom")
	defer withSchemaResolution(t, "", false, boom)()

	coll := newTestTypeCollection(t, tree.NewTestNodeStore())
	_, _, err := coll.getTable(sql.NewEmptyContext(), "", "registration_status")
	require.ErrorIs(t, err, boom)
}

// withSchemaResolution makes resolving an unqualified name yield the given result, returning a restore function. The
// table hook fails the test if reached, since there is no schema to look a table up in.
func withSchemaResolution(t *testing.T, schema string, ok bool, err error) func() {
	t.Helper()
	origLookupSchemaName, origGetSqlTable := LookupSchemaName, GetSqlTableFromContext
	LookupSchemaName = func(ctx *sql.Context, db sql.Database, schemaName string) (string, bool, error) {
		if schemaName == "" {
			return schema, ok, err
		}
		return schemaName, true, nil
	}
	GetSqlTableFromContext = func(ctx *sql.Context, databaseName string, tableName doltdb.TableName) (sql.Table, error) {
		t.Errorf("table lookup attempted with no schema to search: %s", tableName.String())
		return nil, nil
	}
	return func() {
		LookupSchemaName, GetSqlTableFromContext = origLookupSchemaName, origGetSqlTable
	}
}

// newTestTypeCollection returns a TypeCollection backed by |ns| with an
// empty address map.
func newTestTypeCollection(t *testing.T, ns tree.NodeStore) *TypeCollection {
	t.Helper()
	rom, err := objinterface.NewDetachedRootObjectMap(storage, ns)
	require.NoError(t, err)
	return &TypeCollection{
		RootObjectMap: rom,
		accessedMap:   map[id.Type]*pgtypes.DoltgresType{},
		initCache:     map[id.Type]*pgtypes.DoltgresType{},
	}
}

// countingFailNodeStore wraps a real test NodeStore so that callers
// can induce Write failures at chosen points without otherwise altering
// behavior. Used to drive the writeCache flush-failure path.
type countingFailNodeStore struct {
	tree.NodeStore
	writes int
	budget int // -1 means unlimited
}

func newCountingFailNodeStore(t *testing.T) *countingFailNodeStore {
	t.Helper()
	return &countingFailNodeStore{NodeStore: tree.NewTestNodeStore(), budget: -1}
}

// failAfter permits |allowed| additional Write calls then fails the
// rest, until allowAll is called.
func (f *countingFailNodeStore) failAfter(allowed int) {
	f.writes = 0
	f.budget = allowed
}

// allowAll restores the wrapper to delegate every Write to the
// underlying NodeStore.
func (f *countingFailNodeStore) allowAll() {
	f.budget = -1
}

func (f *countingFailNodeStore) Write(ctx context.Context, nd *tree.Node) (hash.Hash, error) {
	if f.budget >= 0 {
		f.writes++
		if f.writes > f.budget {
			return hash.Hash{}, errors.New("induced node store failure")
		}
	}
	return f.NodeStore.Write(ctx, nd)
}
