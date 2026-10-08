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

package plpgsql

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/stretchr/testify/require"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// TestUnfoldedNameLookup covers the resolution of names that a trigger compiled by an older Doltgres left in
// its stored operations. Such a body was compiled before references were folded, so its operations name the
// trigger's records and TG_ variables as its source text spelled them, `NEW.v` where the variable is now
// registered as `new`. The operations are persisted, so those spellings outlive the version that produced
// them and have to keep resolving; only the names a caller supplies are matched this way, since a name the
// function itself declares is compiled to its folded form and must still match exactly.
func TestUnfoldedNameLookup(t *testing.T) {
	sch := sql.Schema{
		{Name: "pk", Type: pgtypes.Int32},
		{Name: "v", Type: pgtypes.Int32},
	}

	t.Run("a trigger record resolves from the spelling an older version stored", func(t *testing.T) {
		stack := NewInterpreterStack(nil)
		stack.NewRecord(TriggerNewRecordName, sch, sql.Row{int32(1), int32(10)})
		stack.markUnfoldedName(TriggerNewRecordName)

		ref, err := stack.GetVariableWithError("NEW.v")
		require.NoError(t, err)
		require.NotNil(t, ref.Type)
		require.Equal(t, int32(10), *ref.Value)

		// The two spellings name one variable rather than two, so a write through the stored spelling is
		// visible to a read through the folded one.
		*ref.Value = int32(11)
		folded, err := stack.GetVariableWithError("new.v")
		require.NoError(t, err)
		require.Equal(t, int32(11), *folded.Value)
	})

	t.Run("a trigger's TG_ variables resolve the same way", func(t *testing.T) {
		stack := NewInterpreterStack(nil)
		stack.NewVariableWithValue("tg_op", pgtypes.Text, "INSERT")
		stack.markUnfoldedName("tg_op")

		ref, err := stack.GetVariableWithError("TG_OP")
		require.NoError(t, err)
		require.Equal(t, "INSERT", *ref.Value)
	})

	t.Run("a declared name is still matched exactly", func(t *testing.T) {
		stack := NewInterpreterStack(nil)
		// A quoted declaration keeps its capitals, and an unquoted reference to it folds to a name that
		// PostgreSQL leaves unresolved. Nothing marked it as caller-supplied, so it stays unresolved.
		stack.NewVariableWithValue("MyVar", pgtypes.Int32, int32(1))

		_, err := stack.GetVariableWithError("myvar")
		require.Error(t, err)
		require.True(t, ErrVariableNotFound.Is(err))
	})

	t.Run("an unmarked name that only differs in case stays unresolved", func(t *testing.T) {
		stack := NewInterpreterStack(nil)
		stack.NewRecord("new", sch, sql.Row{int32(1), int32(10)})

		_, err := stack.GetVariableWithError("NEW.v")
		require.Error(t, err)
		require.True(t, ErrVariableNotFound.Is(err))
	})

	t.Run("a record assignment reaches the record under either spelling", func(t *testing.T) {
		stack := NewInterpreterStack(nil)
		stack.NewRecord(TriggerNewRecordName, sch, sql.Row{int32(1), int32(10)})
		stack.markUnfoldedName(TriggerNewRecordName)

		require.NoError(t, stack.UpdateRecord("NEW", sch, sql.Row{int32(2), int32(20)}))
		ref, err := stack.GetVariableWithError("new.pk")
		require.NoError(t, err)
		require.Equal(t, int32(2), *ref.Value)
	})
}
