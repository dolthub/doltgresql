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

package extensions

import (
	"context"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/utils"
)

// Serialize returns the Extension as a byte slice. If the Extension is invalid (invalid ExtName), then this returns a
// nil slice.
func (ext Extension) Serialize(ctx context.Context) ([]byte, error) {
	if !ext.ExtName.IsValid() {
		return nil, nil
	}

	// Initialize the writer
	writer := utils.NewWriter(256)
	writer.VariableUint(0) // Version
	// Write the extension data
	writer.Id(ext.ExtName.AsId())
	writer.Id(ext.Namespace.AsId())
	writer.Bool(ext.Relocatable)
	writer.String(string(ext.LibIdentifier))
	// Returns the data
	return writer.Data(), nil
}

// DeserializeExtension returns the Extension that was serialized in the byte slice. Returns an empty Extension (has an
// invalid ID) if data is nil or empty.
func DeserializeExtension(ctx context.Context, data []byte) (Extension, error) {
	if len(data) == 0 {
		return Extension{}, nil
	}
	reader := utils.NewReader(data)
	version := reader.VariableUint()
	if version != 0 {
		return Extension{}, errors.Errorf("version %d of extensions are not supported, please upgrade the server", version)
	}

	// Read from the reader
	ext := Extension{}
	ext.ExtName = id.Extension(reader.Id())
	ext.Namespace = id.Namespace(reader.Id())
	ext.Relocatable = reader.Bool()
	ext.LibIdentifier = LibraryIdentifier(reader.String())
	if !reader.IsEmpty() {
		return Extension{}, errors.Errorf("extra data found while deserializing an extension")
	}
	// Return the deserialized object
	return ext, nil
}
