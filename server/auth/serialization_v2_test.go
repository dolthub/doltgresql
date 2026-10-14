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

package auth

import (
	"testing"

	"github.com/dolthub/doltgresql/utils"
)

func TestDeserializeV2WithoutRoleSettings(t *testing.T) {
	// Construct the original v2 format independently of the current serializer:
	// version, role count, then seven empty privilege/membership collections.
	// Existing v2 files end after default privileges, without a role-settings count.
	writer := utils.NewWriter(64)
	writer.Uint32(2)
	writer.Uint32(0)
	for range 7 {
		writer.Uint64(0)
	}
	loaded := newEmptyDatabase()
	if err := loaded.deserialize(writer.Data()); err != nil {
		t.Fatal(err)
	}
	if len(loaded.roleSettings.Data) != 0 {
		t.Fatalf("expected no role settings in a legacy v2 auth database, got %v", loaded.roleSettings.Data)
	}
}
