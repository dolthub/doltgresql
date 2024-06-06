// Copyright 2024 Dolthub, Inc.
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

package tables

import (
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/go-mysql-server/sql"
)

// InitializeTable is a function that will handle the creation of a table, along with setting its initial data.
type InitializeTable func(ctx *sql.Context, db sqle.Database) error

// AddInitializeTable adds the given function to the array of table initializations for the given schema. Tables are
// initialized in the order that they were added.
func AddInitializeTable(schema string, f InitializeTable) {
	if m, ok := initializeTableCallbacks[schema]; ok {
		initializeTableCallbacks[schema] = append(m, f)
	} else {
		initializeTableCallbacks[schema] = []InitializeTable{f}
	}
}

// initializeTableCallbacks contains all InitializeTable functions that will be called by InitializeTables.
var initializeTableCallbacks = map[string][]InitializeTable{}

// InitializeTables initializes all tables for the given schema.
func InitializeTables(ctx *sql.Context, db sqle.Database, schema string) error {
	for _, f := range initializeTableCallbacks[schema] {
		if err := f(ctx, db); err != nil {
			return err
		}
	}
	return nil
}
