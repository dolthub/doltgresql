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
	"os"
	"testing"

	doltserial "github.com/dolthub/dolt/go/gen/fb/serial"
	"github.com/dolthub/dolt/go/libraries/doltcore/schema"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly/tree"
	"github.com/dolthub/dolt/go/store/types"
	flatbuffers "github.com/dolthub/flatbuffers/v23/go"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	"github.com/dolthub/doltgresql/core/sequences"
	"github.com/dolthub/doltgresql/core/storage"
	"github.com/dolthub/doltgresql/flatbuffers/gen/serial"
)

func TestMain(m *testing.M) {
	rootobject.Init()
	os.Exit(m.Run())
}

func TestLoadAndSaveEmptyCollectionsPreserveRoot(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name string
		root func(t testing.TB, ctx context.Context) *RootValue
	}{
		{
			name: "root with no root object fields",
			root: newTestRoot,
		},
		{
			name: "legacy root with zero-hash root object fields",
			root: newLegacyTestRoot,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()
			root := test.root(t, ctx)
			startHash, err := root.HashOf()
			require.NoError(t, err)

			colls, err := rootobject.LoadAllCollections(ctx, root)
			require.NoError(t, err)
			require.NotEmpty(t, colls)
			for _, coll := range colls {
				differs, err := coll.DiffersFrom(ctx, root)
				require.NoError(t, err)
				require.False(t, differs, "freshly loaded collection %d reports unwritten changes", coll.GetID())
				require.False(t, coll.IsStale(ctx, root))

				newRoot, err := coll.UpdateRoot(ctx, root)
				require.NoError(t, err)
				newHash, err := newRoot.HashOf()
				require.NoError(t, err)
				require.Equal(t, startHash, newHash, "collection %d changed the root", coll.GetID())
			}
		})
	}
}

func TestEmptiedCollectionMatchesUnwrittenCollection(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := newTestRootWithSchema(t, ctx)
	startHash, err := root.HashOf()
	require.NoError(t, err)

	seqID := id.NewSequence("public", "seq")
	coll, err := sequences.LoadSequences(ctx, root)
	require.NoError(t, err)
	require.NoError(t, coll.CreateSequence(ctx, &sequences.Sequence{SequenceState: sequences.SequenceState{Id: seqID, Start: 1, Current: 1, Increment: 1}}))

	differs, err := coll.DiffersFrom(ctx, root)
	require.NoError(t, err)
	require.True(t, differs)
	populated, err := coll.UpdateRoot(ctx, root)
	require.NoError(t, err)
	populatedHash, err := populated.HashOf()
	require.NoError(t, err)
	require.NotEqual(t, startHash, populatedHash)

	require.NoError(t, coll.DropSequence(ctx, seqID))
	differs, err = coll.DiffersFrom(ctx, populated)
	require.NoError(t, err)
	require.True(t, differs)
	emptied, err := coll.UpdateRoot(ctx, populated)
	require.NoError(t, err)
	emptiedHash, err := emptied.HashOf()
	require.NoError(t, err)
	require.Equal(t, startHash, emptiedHash)
}

func TestCollectionStaleness(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := newTestRootWithSchema(t, ctx)

	ours, err := sequences.LoadSequences(ctx, root)
	require.NoError(t, err)
	require.False(t, ours.IsStale(ctx, root))

	// A second collection loaded from the same root writes a sequence to it, standing in for any other writer.
	theirs, err := sequences.LoadSequences(ctx, root)
	require.NoError(t, err)
	require.NoError(t, theirs.CreateSequence(ctx,
		&sequences.Sequence{SequenceState: sequences.SequenceState{Id: id.NewSequence("public", "theirs"), Start: 1, Current: 1, Increment: 1}}))
	written, err := theirs.UpdateRoot(ctx, root)
	require.NoError(t, err)

	require.True(t, ours.IsStale(ctx, written), "collection must be stale once another writer changes the root")
	require.False(t, theirs.IsStale(ctx, written), "the writing collection is in sync with the root it wrote")
	require.False(t, ours.IsStale(ctx, root), "collection remains usable against the root it was loaded from")

	// Changes to an unrelated collection leave the sequences collection usable.
	typesColl, err := rootobject.LoadCollection(ctx, written, objinterface.RootObjectID_Types)
	require.NoError(t, err)
	unrelated, err := typesColl.UpdateRoot(ctx, written)
	require.NoError(t, err)
	require.False(t, theirs.IsStale(ctx, unrelated))
}

func TestResolutionCollectionsAreReused(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := newTestRootWithSchema(t, ctx)

	first, err := root.ReadOnlyCollections(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, first)
	second, err := root.ReadOnlyCollections(ctx)
	require.NoError(t, err)
	require.Len(t, second, len(first))
	for i := range first {
		require.Same(t, first[i], second[i], "collection %d was reloaded", first[i].GetID())
	}

	derived, err := root.CreateDatabaseSchema(ctx, schema.DatabaseSchema{Name: "other"})
	require.NoError(t, err)
	derivedColls, err := derived.(*RootValue).ReadOnlyCollections(ctx)
	require.NoError(t, err)
	require.Len(t, derivedColls, len(first))
	for i := range first {
		require.NotSame(t, first[i], derivedColls[i], "collection %d was carried onto a new root", first[i].GetID())
	}
}

func TestUnwrittenRootObjectFieldIsCreated(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := newTestRoot(t, ctx)
	require.Empty(t, root.st.SRV.CastsBytes(), "test requires a root with no casts field")

	coll, err := rootobject.LoadCollection(ctx, root, objinterface.RootObjectID_Casts)
	require.NoError(t, err)
	require.NoError(t, coll.PutRootObject(ctx, newTestCast()))
	newRoot, err := coll.UpdateRoot(ctx, root)
	require.NoError(t, err)

	reloaded, err := rootobject.LoadCollection(ctx, newRoot, objinterface.RootObjectID_Casts)
	require.NoError(t, err)
	has, err := reloaded.HasRootObject(ctx, newTestCast().GetID())
	require.NoError(t, err)
	require.True(t, has, "cast written to a root lacking a casts field was dropped")
}

// newTestRoot returns an empty in-memory root.
func newTestRoot(t testing.TB, ctx context.Context) *RootValue {
	t.Helper()
	root, err := emptyRootValue(ctx, types.NewMemoryValueStore(), tree.NewTestNodeStore())
	require.NoError(t, err)
	return root.(*RootValue)
}

// newTestRootWithSchema returns an empty in-memory root that has a database schema and no root object fields.
func newTestRootWithSchema(t testing.TB, ctx context.Context) *RootValue {
	t.Helper()
	root, err := newTestRoot(t, ctx).CreateDatabaseSchema(ctx, schema.DatabaseSchema{Name: "public"})
	require.NoError(t, err)
	require.Empty(t, root.(*RootValue).st.SRV.CastsBytes())
	return root.(*RootValue)
}

// newLegacyTestRoot returns an empty in-memory root with every root object field written as a zero hash, matching
// the roots that the previous version serialized.
func newLegacyTestRoot(t testing.TB, ctx context.Context) *RootValue {
	t.Helper()
	base := newTestRoot(t, ctx)
	builder := flatbuffers.NewBuilder(80)
	tablesOffset := builder.CreateByteVector(base.st.SRV.TablesBytes())
	fkOffset := builder.CreateByteVector(base.st.SRV.ForeignKeyAddrBytes())
	var empty hash.Hash
	rootObjOffsets := make([]flatbuffers.UOffsetT, len(storage.RootObjectSerializations))
	for i := range storage.RootObjectSerializations {
		rootObjOffsets[i] = builder.CreateByteVector(empty[:])
	}
	serial.RootValueStart(builder)
	serial.RootValueAddFeatureVersion(builder, base.st.SRV.FeatureVersion())
	serial.RootValueAddCollation(builder, base.st.SRV.Collation())
	serial.RootValueAddTables(builder, tablesOffset)
	serial.RootValueAddForeignKeyAddr(builder, fkOffset)
	for i := range storage.RootObjectSerializations {
		storage.RootObjectSerializations[i].RootValueAdd(builder, rootObjOffsets[i])
	}
	bs := doltserial.FinishMessage(builder, serial.RootValueEnd(builder), []byte(doltserial.DoltgresRootValueFileID))
	root, err := newRootValue(ctx, base.vrw, base.ns, types.SerialMessage(bs))
	require.NoError(t, err)
	require.Len(t, root.(*RootValue).st.SRV.CastsBytes(), hash.ByteLen)
	return root.(*RootValue)
}

// newTestCast returns a cast. The types that it converts are not significant.
func newTestCast() objinterface.RootObject {
	return casts.Cast{
		ID:       id.NewCast(id.NewType("pg_catalog", "int4"), id.NewType("pg_catalog", "text")),
		CastType: casts.CastType_Explicit,
		Function: id.NullFunction,
		UseInOut: true,
	}
}
