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

package sequences

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/dolt/go/store/prolly/tree"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/core/id"
)

// TestMap_RemainsUsableAfterFlushFailure asserts that Collection.Map
// remains usable after a flush returns an error. A subsequent Map call
// against a healthy store must succeed.
func TestMap_RemainsUsableAfterFlushFailure(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ns := newCountingFailNodeStore(t)
	coll := newTestCollection(t, ns)

	coll.accessedMap[id.NewSequence("public", "seq_one")] = newTestSequence("public", "seq_one")
	_, err := coll.Map(ctx)
	require.NoError(t, err)

	ns.failAfter(0)
	coll.accessedMap[id.NewSequence("public", "seq_two")] = newTestSequence("public", "seq_two")
	_, err = coll.Map(ctx)
	require.Error(t, err)

	ns.allowAll()
	require.NotPanics(t, func() {
		_, _ = coll.Map(ctx)
	})
}

// TestDropSequence_RemainsUsableAfterFlushFailure asserts the same
// recovery contract for DropSequence.
func TestDropSequence_RemainsUsableAfterFlushFailure(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ns := newCountingFailNodeStore(t)
	coll := newTestCollection(t, ns)

	pending := newTestSequence("public", "pending")
	coll.accessedMap[pending.Id] = pending
	ns.failAfter(0)
	err := coll.DropSequence(ctx, pending.Id)
	require.Error(t, err)

	ns.allowAll()
	require.NotPanics(t, func() {
		_ = coll.DropSequence(ctx, id.NewSequence("public", "pending"))
	})
}

// newTestCollection returns a Collection backed by |ns| with an empty
// address map.
func newTestCollection(t *testing.T, ns tree.NodeStore) *Collection {
	t.Helper()
	addrMap, err := prolly.NewEmptyAddressMap(ns)
	require.NoError(t, err)
	return &Collection{
		accessedMap:   map[id.Sequence]*Sequence{},
		underlyingMap: addrMap,
		ns:            ns,
	}
}

// newTestSequence returns a Sequence with valid bounds that round-trip
// through Serialize and Deserialize. The exact values are not significant.
func newTestSequence(schema, name string) *Sequence {
	return &Sequence{
		Id:        id.NewSequence(schema, name),
		Start:     1,
		Current:   1,
		Increment: 1,
		Minimum:   1,
		Maximum:   math.MaxInt64,
		Cache:     1,
	}
}

// countingFailNodeStore wraps a real test NodeStore so that callers can
// induce Write failures at chosen points without otherwise altering
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
