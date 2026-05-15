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

package casts

import (
	"context"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/utils"
)

// Serialize returns the Cast as a byte slice. If the Cast is invalid, then this returns a nil slice.
func (cast Cast) Serialize(ctx context.Context) ([]byte, error) {
	if !cast.ID.IsValid() {
		return nil, nil
	}

	// Initialize the writer and version
	writer := utils.NewWriter(256)
	writer.VariableUint(0) // Version
	// Write the cast data
	writer.Id(cast.ID.AsId())
	writer.Uint8(uint8(cast.CastType))
	writer.Id(cast.Function.AsId())
	writer.Bool(cast.UseInOut)
	// Returns the data
	return writer.Data(), nil
}

// DeserializeCast returns the Cast that was serialized in the byte slice. Returns an empty Cast (invalid ID) if data is
// nil or empty.
func DeserializeCast(ctx context.Context, data []byte) (Cast, error) {
	if len(data) == 0 {
		return Cast{}, nil
	}
	reader := utils.NewReader(data)
	version := reader.VariableUint()
	if version != 0 {
		return Cast{}, errors.Errorf("version %d of casts is not supported, please upgrade the server", version)
	}

	// Read from the reader
	t := Cast{}
	t.ID = id.Cast(reader.Id())
	t.CastType = CastType(reader.Uint8())
	t.Function = id.Function(reader.Id())
	t.UseInOut = reader.Bool()
	if !reader.IsEmpty() {
		return Cast{}, errors.Errorf("extra data found while deserializing a cast")
	}
	// Return the deserialized object
	return t, nil
}
