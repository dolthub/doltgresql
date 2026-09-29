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
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/schema"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/publications"
)

// TestPublicationPersistence exercises absent fields on old roots and preservation by unrelated root rewrites.
func TestPublicationPersistence(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	root := newTestRoot(t, ctx)
	initialHash, err := root.HashOf()
	require.NoError(t, err)
	coll, err := publications.LoadPublications(ctx, root)
	require.NoError(t, err)
	p := publications.Publication{ID: id.NewPublication("supabase_realtime"), OwnerRoleID: 42,
		AllTables: true, PublishInsert: true, PublishDelete: true, PublishViaRoot: true}
	require.NoError(t, coll.CreatePublication(ctx, p))
	updated, err := coll.UpdateRoot(ctx, root)
	require.NoError(t, err)
	root = updated.(*RootValue)
	// Adding a schema rebuilds the flatbuffer and must preserve the new collection.
	withSchema, err := root.CreateDatabaseSchema(ctx, schema.DatabaseSchema{Name: "public"})
	require.NoError(t, err)
	root = withSchema.(*RootValue)
	reloaded, err := publications.LoadPublications(ctx, root)
	require.NoError(t, err)
	actual, err := reloaded.GetPublication(ctx, p.ID)
	require.NoError(t, err)
	require.Equal(t, p, actual)
	_, plainID, err := reloaded.ResolveName(ctx, doltdb.TableName{Name: p.ID.Name()})
	require.NoError(t, err)
	require.False(t, plainID.IsValid(), "publications must not resolve as SQL relations")
	_, objectID, err := reloaded.ResolveName(ctx, p.Name())
	require.NoError(t, err)
	require.Equal(t, p.ID.AsId(), objectID)
	// Dropping validates the whole list before making any change.
	require.Error(t, reloaded.DropPublication(ctx, p.ID, id.NewPublication("missing")))
	require.True(t, reloaded.HasPublication(ctx, p.ID))
	require.NoError(t, reloaded.DropPublication(ctx, p.ID))
	updated, err = reloaded.UpdateRoot(ctx, root)
	require.NoError(t, err)
	empty, err := publications.LoadPublications(ctx, updated)
	require.NoError(t, err)
	require.False(t, empty.HasPublication(ctx, p.ID))
	// An add/drop round trip on the original root must retain its canonical empty hash.
	original := newTestRoot(t, ctx)
	col, err := publications.LoadPublications(ctx, original)
	require.NoError(t, err)
	require.NoError(t, col.CreatePublication(ctx, p))
	require.NoError(t, col.DropPublication(ctx, p.ID))
	updated, err = col.UpdateRoot(ctx, original)
	require.NoError(t, err)
	h, err := updated.HashOf()
	require.NoError(t, err)
	require.Equal(t, initialHash, h)
}

// TestPublicationMerge ensures one-sided edits merge and competing definitions remain conflicts.
func TestPublicationMerge(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	base := publications.Publication{ID: id.NewPublication("p"), OwnerRoleID: 1, PublishInsert: true}
	ours := base
	ours.PublishUpdate = true
	theirs := base
	theirs.PublishDelete = true
	coll := &publications.Collection{}
	diffs, merged, err := coll.DiffRootObjects(ctx, "", ours, base, base)
	require.NoError(t, err)
	require.Empty(t, diffs)
	require.Equal(t, ours, merged)
	diffs, merged, err = coll.DiffRootObjects(ctx, "", base, theirs, base)
	require.NoError(t, err)
	require.Empty(t, diffs)
	require.Equal(t, theirs, merged)
	diffs, _, err = coll.DiffRootObjects(ctx, "", ours, theirs, base)
	require.NoError(t, err)
	require.Len(t, diffs, 1)
	resolved, err := coll.UpdateField(ctx, ours, diffs[0].FieldName, diffs[0].TheirValue)
	require.NoError(t, err)
	require.Equal(t, theirs, resolved)
}
