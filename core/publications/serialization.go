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

package publications

import (
	"context"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/utils"
)

// Serialize returns the Publication as a byte slice. If the Publication is invalid (invalid ID), then this returns a
// nil slice.
func (publication Publication) Serialize(ctx context.Context) ([]byte, error) {
	if !publication.ID.IsValid() {
		return nil, nil
	}

	// Initialize the writer
	writer := utils.NewWriter(256)
	writer.VariableUint(1) // Version
	// Write the publication data
	writer.Id(publication.ID.AsId())
	writer.Uint64(publication.OwnerRoleID)
	writer.Bool(publication.AllTables)
	writer.Bool(publication.PublishInsert)
	writer.Bool(publication.PublishUpdate)
	writer.Bool(publication.PublishDelete)
	writer.Bool(publication.PublishTruncate)
	writer.Bool(publication.PublishViaRoot)
	// Returns the data
	return writer.Data(), nil
}

// DeserializePublication returns the Publication that was serialized in the byte slice. Returns an empty Publication (has an
// invalid ID) if data is nil or empty.
func DeserializePublication(ctx context.Context, data []byte) (Publication, error) {
	if len(data) == 0 {
		return Publication{}, nil
	}
	reader := utils.NewReader(data)
	version := reader.VariableUint()
	switch version {
	case 1:
		// current version
	default:
		return Publication{}, errors.Errorf("version %d of publications are not supported, please upgrade the server", version)
	}

	// Read from the reader
	publication := Publication{}
	publication.ID = id.Publication(reader.Id())
	publication.OwnerRoleID = reader.Uint64()
	publication.AllTables = reader.Bool()
	publication.PublishInsert = reader.Bool()
	publication.PublishUpdate = reader.Bool()
	publication.PublishDelete = reader.Bool()
	publication.PublishTruncate = reader.Bool()
	publication.PublishViaRoot = reader.Bool()
	if !reader.IsEmpty() {
		return Publication{}, errors.Errorf("publicationra data found while deserializing a publication")
	}
	// Return the deserialized object
	return publication, nil
}
