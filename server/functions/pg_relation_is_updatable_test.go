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

package functions

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

// These test doubles expose only the selected write interfaces. Their methods
// are not called: capability detection must not create editors or modify rows.
type relationTestUpdate struct{}

func (relationTestUpdate) Updater(*sql.Context) sql.RowUpdater { panic("unexpected editor creation") }

type relationTestInsert struct{}

func (relationTestInsert) Inserter(*sql.Context) sql.RowInserter { panic("unexpected editor creation") }

type relationTestDelete struct{}

func (relationTestDelete) Deleter(*sql.Context) sql.RowDeleter { panic("unexpected editor creation") }

func TestRelationWriteCapabilities(t *testing.T) {
	for _, tc := range []struct {
		name     string
		table    sql.Table
		expected int32
	}{
		{"read only", struct{ sql.Table }{}, 0},
		{"update", struct {
			sql.Table
			relationTestUpdate
		}{}, 4},
		{"insert", struct {
			sql.Table
			relationTestInsert
		}{}, 8},
		{"delete", struct {
			sql.Table
			relationTestDelete
		}{}, 16},
		{"update and insert", struct {
			sql.Table
			relationTestUpdate
			relationTestInsert
		}{}, 12},
		{"update and delete", struct {
			sql.Table
			relationTestUpdate
			relationTestDelete
		}{}, 20},
		{"insert and delete", struct {
			sql.Table
			relationTestInsert
			relationTestDelete
		}{}, 24},
		{"all writes", struct {
			sql.Table
			relationTestUpdate
			relationTestInsert
			relationTestDelete
		}{}, 28},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := relationWriteCapabilities(tc.table); got != tc.expected {
				t.Fatalf("capabilities = %d, want %d", got, tc.expected)
			}
		})
	}
}
