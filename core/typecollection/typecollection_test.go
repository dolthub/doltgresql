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

	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/dolt/go/store/prolly/tree"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/core/id"
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

// newTestTypeCollection returns a TypeCollection backed by |ns| with an
// empty address map.
func newTestTypeCollection(t *testing.T, ns tree.NodeStore) *TypeCollection {
	t.Helper()
	addrMap, err := prolly.NewEmptyAddressMap(ns)
	require.NoError(t, err)
	return &TypeCollection{
		accessedMap:   map[id.Type]*pgtypes.DoltgresType{},
		underlyingMap: addrMap,
		ns:            ns,
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
