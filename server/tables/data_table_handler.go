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
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
)

// DataTableHandler controls the modification and exposure of data to a DataTable. As the data in a DataTable may not be
// stored in the root's table map, the handler provides an interface to access that data from its actual storage
// location, or to employ a hybrid model where only some data resides in the table map.
type DataTableHandler interface {
	// Insert handles the insertion of rows into the DataTable. The given editor handles the underlying table storage,
	// and the handler fully controls when (and if) data is written to the underlying table storage.
	Insert(ctx *sql.Context, editor *DataTableEditor, row sql.Row) error
	// Update handles the updating of rows in the DataTable. The given editor handles the underlying table storage, and
	// the handler fully controls when (and if) data is modified in the underlying table storage.
	Update(ctx *sql.Context, editor *DataTableEditor, old sql.Row, new sql.Row) error
	// Delete handles the deletion of rows from the DataTable. The given editor handles the underlying table storage,
	// and the handler fully controls when (and if) data is deleted from the underlying table storage.
	Delete(ctx *sql.Context, editor *DataTableEditor, row sql.Row) error
	// UsesIndexes returns whether indexes are exposed on this table. Tables that have their own form of storage (such
	// as being located on the root directly) will generally not use indexes, although hybrid storage models may use them.
	UsesIndexes() bool
	// RowIter receives a sql.RowIter on the underlying storage, and returns a sql.RowIter that contains the full
	// contents of a row. The incoming sql.RowIter may return a subset of rows due to an index, for handlers that make
	// use of indexes. For handlers that are stored exclusively outside the table map, they may safely ignore the
	// incoming sql.RowIter and return their own that iterates over the entire data set.
	RowIter(ctx *sql.Context, rowIter sql.RowIter) (sql.RowIter, error)
}

// handlers is a map from the schema name, to the table name, to the handler.
var handlers = map[string]map[string]DataTableHandler{}

// AddHandler adds the given handler to the handler set.
func AddHandler(schemaName string, tableName string, handler DataTableHandler) {
	tableMap, ok := handlers[schemaName]
	if !ok {
		tableMap = make(map[string]DataTableHandler)
		handlers[schemaName] = tableMap
	}
	tableMap[tableName] = handler
}

// getHandler gets the handler using the given schema and table name. Returns an invalidHandler if one cannot be found.
func getHandler(schemaName string, tableName string) DataTableHandler {
	if tableMap, ok := handlers[schemaName]; ok {
		if handler, ok := tableMap[tableName]; ok {
			return handler
		}
	}
	return invalidHandler{}
}

// invalidHandler is used when an appropriate handler cannot be found.
type invalidHandler struct{}

// Insert implements the interface tables.DataTableHandler.
func (invalidHandler) Insert(ctx *sql.Context, editor *DataTableEditor, row sql.Row) error {
	return fmt.Errorf("cannot insert using an invalid handler")
}

// Update implements the interface tables.DataTableHandler.
func (invalidHandler) Update(ctx *sql.Context, editor *DataTableEditor, old sql.Row, new sql.Row) error {
	return fmt.Errorf("cannot update using an invalid handler")
}

// Delete implements the interface tables.DataTableHandler.
func (invalidHandler) Delete(ctx *sql.Context, editor *DataTableEditor, row sql.Row) error {
	return fmt.Errorf("cannot delete using an invalid handler")
}

// UsesIndexes implements the interface tables.DataTableHandler.
func (invalidHandler) UsesIndexes() bool {
	return false
}

// RowIter implements the interface tables.DataTableHandler.
func (invalidHandler) RowIter(ctx *sql.Context, rowIter sql.RowIter) (sql.RowIter, error) {
	return nil, fmt.Errorf("cannot iterate rows using an invalid handler")
}
