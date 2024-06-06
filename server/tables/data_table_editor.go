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
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
)

// DataTableEditor is an editor that encapsulates a table's row storage.
type DataTableEditor struct {
	editor dsess.TableWriter
}

// dataTableEditorInterface is the interface that handles all aspects of the DataTableEditor outside of those that are
// controlled by the handler.
type dataTableEditorInterface struct {
	*DataTableEditor
	handler DataTableHandler
}

var _ dsess.TableWriter = (*DataTableEditor)(nil)

// newDataTableEditorInterface creates a new dataTableEditorInterface from the given dsess.TableWriter.
func newDataTableEditorInterface(doltEditor dsess.TableWriter, handler DataTableHandler) dataTableEditorInterface {
	return dataTableEditorInterface{
		DataTableEditor: &DataTableEditor{
			editor: doltEditor,
		},
		handler: handler,
	}
}

// AcquireAutoIncrementLock implements the interface dsess.TableWriter.
func (dte *DataTableEditor) AcquireAutoIncrementLock(ctx *sql.Context) (func(), error) {
	return dte.editor.AcquireAutoIncrementLock(ctx)
}

// Close implements the interface dsess.TableWriter.
func (dte *DataTableEditor) Close(ctx *sql.Context) error {
	return dte.editor.Close(ctx)
}

// Delete implements the interface dsess.TableWriter.
func (dte *DataTableEditor) Delete(ctx *sql.Context, row sql.Row) error {
	return dte.editor.Delete(ctx, row)
}

// DiscardChanges implements the interface dsess.TableWriter.
func (dte *DataTableEditor) DiscardChanges(ctx *sql.Context, errorEncountered error) error {
	return dte.editor.DiscardChanges(ctx, errorEncountered)
}

// GetIndexes implements the interface dsess.TableWriter.
func (dte *DataTableEditor) GetIndexes(ctx *sql.Context) ([]sql.Index, error) {
	return dte.editor.GetIndexes(ctx)
}

// IndexedAccess implements the interface dsess.TableWriter.
func (dte *DataTableEditor) IndexedAccess(lookup sql.IndexLookup) sql.IndexedTable {
	return dte.editor.IndexedAccess(lookup)
}

// Insert implements the interface dsess.TableWriter.
func (dte *DataTableEditor) Insert(ctx *sql.Context, row sql.Row) error {
	return dte.editor.Insert(ctx, row)
}

// PreciseMatch implements the interface dsess.TableWriter.
func (dte *DataTableEditor) PreciseMatch() bool {
	return dte.editor.PreciseMatch()
}

// SetAutoIncrementValue implements the interface dsess.TableWriter.
func (dte *DataTableEditor) SetAutoIncrementValue(ctx *sql.Context, u uint64) error {
	return dte.editor.SetAutoIncrementValue(ctx, u)
}

// StatementBegin implements the interface dsess.TableWriter.
func (dte *DataTableEditor) StatementBegin(ctx *sql.Context) {
	dte.editor.StatementBegin(ctx)
}

// StatementComplete implements the interface dsess.TableWriter.
func (dte *DataTableEditor) StatementComplete(ctx *sql.Context) error {
	return dte.editor.StatementComplete(ctx)
}

// Update implements the interface dsess.TableWriter.
func (dte *DataTableEditor) Update(ctx *sql.Context, old sql.Row, new sql.Row) error {
	return dte.editor.Update(ctx, old, new)
}

// Delete implements the interface dsess.TableWriter.
func (iface dataTableEditorInterface) Delete(ctx *sql.Context, row sql.Row) error {
	return iface.handler.Delete(ctx, iface.DataTableEditor, row)
}

// Insert implements the interface dsess.TableWriter.
func (iface dataTableEditorInterface) Insert(ctx *sql.Context, row sql.Row) error {
	return iface.handler.Insert(ctx, iface.DataTableEditor, row)
}

// Update implements the interface dsess.TableWriter.
func (iface dataTableEditorInterface) Update(ctx *sql.Context, old sql.Row, new sql.Row) error {
	return iface.handler.Update(ctx, iface.DataTableEditor, old, new)
}
