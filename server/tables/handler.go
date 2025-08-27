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

import "github.com/dolthub/go-mysql-server/sql"

// Handler is an interface that controls how data is represented for some table.
type Handler interface {
	// Name returns the name of the table.
	Name() string
	// RowIter returns a sql.RowIter that returns the rows of the table.
	RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error)
	// Schema returns the table's schema.
	Schema() sql.PrimaryKeySchema
}

type IndexedTableHandler interface {
	Handler
	// Indexes returns the table's indexes.
	Indexes() ([]sql.Index, error)
	// LookupPartitions returns a sql.PartitionIter that can be used to look up rows in the table using the given lookup
	LookupPartitions(context *sql.Context, lookup sql.IndexLookup) (sql.PartitionIter, error)
}

// handlers is a map from the schema name, to the table name, to the handler.
var handlers = map[string]map[string]Handler{}

// AddHandler adds the given handler to the handler set.
func AddHandler(schemaName string, tableName string, handler Handler) {
	tableMap, ok := handlers[schemaName]
	if !ok {
		tableMap = make(map[string]Handler)
		handlers[schemaName] = tableMap
	}
	tableMap[tableName] = handler
}
