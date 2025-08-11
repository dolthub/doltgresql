// Copyright 2025 Dolthub, Inc.
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

package conflicts

import (
	"context"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	"github.com/dolthub/doltgresql/utils"
)

// Serialize returns the Conflict as a byte slice. If the Conflict is invalid, then this returns a nil slice.
func (conflict Conflict) Serialize(ctx context.Context) (_ []byte, err error) {
	if !conflict.ID.IsValid() {
		return nil, nil
	}

	// Write all of the conflicts to the writer
	writer := utils.NewWriter(512)
	writer.VariableUint(0) // Version
	// Serialize "ours", "theirs", and "ancestor"
	var ours, theirs, ancestor []byte
	if conflict.Ours != nil {
		ours, err = conflict.Ours.Serialize(ctx)
		if err != nil {
			return nil, err
		}
	}
	if conflict.Theirs != nil {
		theirs, err = conflict.Theirs.Serialize(ctx)
		if err != nil {
			return nil, err
		}
	}
	if conflict.Ancestor != nil {
		ancestor, err = conflict.Ancestor.Serialize(ctx)
		if err != nil {
			return nil, err
		}
	}
	// Write the conflict data
	writer.Id(conflict.ID)
	writer.String(conflict.FromHash)
	writer.Int64(int64(conflict.RootObjectID))
	writer.Bool(conflict.Ours != nil)
	writer.Bool(conflict.Theirs != nil)
	writer.Bool(conflict.Ancestor != nil)
	writer.ByteSlice(ours)
	writer.ByteSlice(theirs)
	writer.ByteSlice(ancestor)
	// Returns the data
	return writer.Data(), nil
}

// DeserializeConflict returns the Conflict that was serialized in the byte slice. Returns an empty Conflict (invalid
// ID) if data is nil or empty.
func DeserializeConflict(ctx context.Context, data []byte) (_ Conflict, err error) {
	if len(data) == 0 {
		return Conflict{}, nil
	}
	reader := utils.NewReader(data)
	version := reader.VariableUint()
	if version > 0 {
		return Conflict{}, errors.Errorf("version %d of conflicts is not supported, please upgrade the server", version)
	}

	// Read from the reader
	conflict := Conflict{}
	conflict.ID = reader.Id()
	conflict.FromHash = reader.String()
	conflict.RootObjectID = objinterface.RootObjectID(reader.Int64())
	hasOurs := reader.Bool()
	hasTheirs := reader.Bool()
	hasAncestor := reader.Bool()
	ours := reader.ByteSlice()
	theirs := reader.ByteSlice()
	ancestor := reader.ByteSlice()
	// Deserialize "ours", "theirs", and "ancestor"
	if hasOurs {
		conflict.Ours, err = DeserializeRootObject(ctx, conflict.RootObjectID, ours)
		if err != nil {
			return Conflict{}, err
		}
	}
	if hasTheirs {
		conflict.Theirs, err = DeserializeRootObject(ctx, conflict.RootObjectID, theirs)
		if err != nil {
			return Conflict{}, err
		}
	}
	if hasAncestor {
		conflict.Ancestor, err = DeserializeRootObject(ctx, conflict.RootObjectID, ancestor)
		if err != nil {
			return Conflict{}, err
		}
	}
	if !reader.IsEmpty() {
		return Conflict{}, errors.Errorf("extra data found while deserializing a conflict")
	}
	// Return the deserialized object
	return conflict, nil
}

// DeserializeRootObject handles root object deserialization, and is declared in a different package. It is assigned
// here by an Init function to get around import cycles.
var DeserializeRootObject = func(ctx context.Context, rootObjID objinterface.RootObjectID, data []byte) (objinterface.RootObject, error) {
	return nil, errors.New("DeserializeRootObject was never initialized")
}
