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

package dtables

import (
	"fmt"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/adapters"
	doltdtables "github.com/dolthub/dolt/go/libraries/doltcore/sqle/dtables"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// DoltgresDoltStatusTableAdapter adapts the [doltdtables.StatusTable] into a Doltgres-compatible version.
//
// DoltgresDoltStatusTableAdapter implements the [adapters.TableAdapter] interface.
type DoltgresDoltStatusTableAdapter struct{}

var _ adapters.TableAdapter = DoltgresDoltStatusTableAdapter{}

// NewTable returns a new [sql.Table] for Doltgres' version of [doltdtables.StatusTable].
func (a DoltgresDoltStatusTableAdapter) NewTable(ctx *sql.Context, tableName string, ddb *doltdb.DoltDB, ws *doltdb.WorkingSet, rp env.RootsProvider[*sql.Context]) sql.Table {
	doltTable := doltdtables.NewStatusTableWithNoAdapter(ctx, tableName, ddb, ws, rp)
	return &doltgresDoltStatusTable{
		srcDoltStatus: doltTable.(*doltdtables.StatusTable),
		rootsProvider: rp,
	}
}

// TableName returns the table name for Doltgres' version of [doltdtables.StatusTable].
func (a DoltgresDoltStatusTableAdapter) TableName() string {
	return DoltgresDoltStatusTableName
}

// DoltgresDoltStatusTableName is the name of Dolt's status table following Doltgres' naming conventions.
const DoltgresDoltStatusTableName = "status"

// doltgresDoltStatusTable translates the [doltdtables.StatusTable] into a Doltgres-compatible version.
//
// doltgresDoltStatusTable implements the [sql.Table] and [sql.StatisticsTable] interfaces.
type doltgresDoltStatusTable struct {
	srcDoltStatus *doltdtables.StatusTable
	rootsProvider env.RootsProvider[*sql.Context]
}

var _ sql.Table = (*doltgresDoltStatusTable)(nil)
var _ sql.StatisticsTable = (*doltgresDoltStatusTable)(nil)

// Name returns the name of Doltgres' version of the Dolt status table.
func (w *doltgresDoltStatusTable) Name() string {
	return w.srcDoltStatus.Name()
}

// Schema returns the schema for Doltgres' version of the Dolt status table.
func (w *doltgresDoltStatusTable) Schema(ctx *sql.Context) sql.Schema {
	return []*sql.Column{
		{Name: "table_name", Type: pgtypes.Text, Source: DoltgresDoltStatusTableName, PrimaryKey: true, Nullable: false},
		{Name: "staged", Type: pgtypes.Bool, Source: DoltgresDoltStatusTableName, PrimaryKey: true, Nullable: false},
		{Name: "status", Type: pgtypes.Text, Source: DoltgresDoltStatusTableName, PrimaryKey: true, Nullable: false},
	}
}

// String returns the string representation of [doltdtables.StatusTable].
func (w *doltgresDoltStatusTable) String() string {
	return w.srcDoltStatus.String()
}

// Collation returns the [sql.CollationID] from [doltdtables.StatusTable].
func (w *doltgresDoltStatusTable) Collation() sql.CollationID {
	return w.srcDoltStatus.Collation()
}

// Partitions returns a [sql.PartitionIter] on the partitions of [doltdtables.StatusTable].
func (w *doltgresDoltStatusTable) Partitions(ctx *sql.Context) (sql.PartitionIter, error) {
	return w.srcDoltStatus.Partitions(ctx)
}

// PartitionRows returns a wrapped [sql.RowIter] for the rows in |partition| from
// [doltdtables.StatusTable.PartitionRows] to later apply column transformations that match Doltgres' version of the
// Dolt status table schema.
func (w *doltgresDoltStatusTable) PartitionRows(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	iter, err := w.srcDoltStatus.PartitionRows(ctx, partition)
	if err != nil {
		return nil, err
	}
	rootObjects, err := w.loadStatusRootObjects(ctx)
	if err != nil {
		return nil, err
	}
	return &doltgresDoltStatusRowIter{w, iter, rootObjects}, nil
}

// loadStatusRootObjects loads the root object kinds from each of the roots that the Dolt status table compares.
func (w *doltgresDoltStatusTable) loadStatusRootObjects(ctx *sql.Context) (statusRootObjects, error) {
	if w.rootsProvider == nil {
		return statusRootObjects{}, nil
	}
	roots, err := w.rootsProvider.GetRoots(ctx)
	if err != nil {
		return statusRootObjects{}, err
	}
	head, err := loadRootObjectIDs(ctx, roots.Head)
	if err != nil {
		return statusRootObjects{}, err
	}
	staged, err := loadRootObjectIDs(ctx, roots.Staged)
	if err != nil {
		return statusRootObjects{}, err
	}
	working, err := loadRootObjectIDs(ctx, roots.Working)
	if err != nil {
		return statusRootObjects{}, err
	}
	return statusRootObjects{head: head, staged: staged, working: working}, nil
}

// loadRootObjectIDs returns the root object ID of each root object in |root|, keyed by the root object's name.
func loadRootObjectIDs(ctx *sql.Context, root doltdb.RootValue) (map[string]objinterface.RootObjectID, error) {
	rootObjectIDs := make(map[string]objinterface.RootObjectID)
	if root == nil {
		return rootObjectIDs, nil
	}
	err := root.IterRootObjects(ctx, func(name doltdb.TableName, rootObj doltdb.RootObject) (stop bool, err error) {
		if doltgresRootObj, ok := rootObj.(objinterface.RootObject); ok {
			rootObjectIDs[name.String()] = doltgresRootObj.GetRootObjectID()
		}
		return false, nil
	})
	return rootObjectIDs, err
}

// DataLength returns the length of the data in bytes from [doltdtables.StatusTable].
func (w *doltgresDoltStatusTable) DataLength(ctx *sql.Context) (uint64, error) {
	return w.srcDoltStatus.DataLength(ctx)
}

// RowCount returns exact (true) or estimate (false) number of rows from [doltdtables.StatusTable].
func (w *doltgresDoltStatusTable) RowCount(ctx *sql.Context) (uint64, bool, error) {
	return w.srcDoltStatus.RowCount(ctx)
}

// doltgresDoltStatusRowIter wraps [doltdtables.StatusTable] [sql.RowIter] and applies transformations before returning
// its rows to make sure they're compatible with Doltgres' version of Dolt's status table.
type doltgresDoltStatusRowIter struct {
	doltStatusTable sql.Table
	rowIter         sql.RowIter
	rootObjects     statusRootObjects
}

// statusRootObjects holds the root object IDs, keyed by name, for each root that the Dolt status table compares.
type statusRootObjects struct {
	head    map[string]objinterface.RootObjectID
	staged  map[string]objinterface.RootObjectID
	working map[string]objinterface.RootObjectID
}

var _ sql.RowIter = (*doltgresDoltStatusRowIter)(nil)

// Next converts the 'staged' column from [doltdtables.StatusTable.Schema] from a byte into a bool since, unlike the
// MySQL wire protocol, Doltgres has a real bool type. It also replaces the table statuses that Dolt reports for root
// objects with statuses naming the kind of root object, such as "new function".
func (i *doltgresDoltStatusRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	row, err := i.rowIter.Next(ctx)
	if err != nil {
		return nil, err
	}

	// Dolt uses byte to avoid MySQL wire protocol ambiguity on tinyint(1) and bool.
	// See: https://github.com/dolthub/dolt/pull/10117
	stagedIndex := i.doltStatusTable.Schema(ctx).IndexOfColName("staged")
	stagedVal, ok := row[stagedIndex].(byte)
	if !ok {
		return nil, fmt.Errorf("expected staged column at index %d to be byte, got %T", stagedIndex, row[stagedIndex])
	}
	row[stagedIndex] = stagedVal != 0

	nameIndex := i.doltStatusTable.Schema(ctx).IndexOfColName("table_name")
	statusIndex := i.doltStatusTable.Schema(ctx).IndexOfColName("status")
	name, nameOk := row[nameIndex].(string)
	status, statusOk := row[statusIndex].(string)
	if nameOk && statusOk {
		// Deleted root objects only exist in the "from" root, while all others exist in the "to" root.
		var rootObjectIDs map[string]objinterface.RootObjectID
		switch {
		case stagedVal != 0 && status == statusDeleted:
			rootObjectIDs = i.rootObjects.head
		case stagedVal != 0 || status == statusDeleted:
			rootObjectIDs = i.rootObjects.staged
		default:
			rootObjectIDs = i.rootObjects.working
		}
		if rootObjectID, ok := rootObjectIDs[name]; ok {
			row[statusIndex] = rootObjectStatus(status, rootObjectID)
		}
	}

	return row, nil
}

// These are the statuses that Dolt reports for tables, which are also reported for root objects.
const (
	statusNewTable = "new table"
	statusDeleted  = "deleted"
	statusModified = "modified"
)

// rootObjectStatus returns the status for a root object of the given kind, replacing the given table status. Statuses
// that do not apply to a single root object, such as conflicts, are returned unchanged.
func rootObjectStatus(status string, rootObjectID objinterface.RootObjectID) string {
	var kind string
	switch rootObjectID {
	case objinterface.RootObjectID_Sequences:
		kind = "sequence"
	case objinterface.RootObjectID_Types:
		kind = "type"
	case objinterface.RootObjectID_Functions:
		kind = "function"
	case objinterface.RootObjectID_Triggers:
		kind = "trigger"
	case objinterface.RootObjectID_Extensions:
		kind = "extension"
	case objinterface.RootObjectID_Procedures:
		kind = "procedure"
	case objinterface.RootObjectID_Casts:
		kind = "cast"
	case objinterface.RootObjectID_Operators:
		kind = "operator"
	case objinterface.RootObjectID_Aggregates:
		kind = "aggregate"
	default:
		return status
	}
	switch status {
	case statusNewTable:
		return "new " + kind
	case statusDeleted:
		return "deleted " + kind
	case statusModified:
		return "modified " + kind
	default:
		return status
	}
}

// Close closes the wrapped [doltdtables.StatusTable] [sql.RowIter].
func (i *doltgresDoltStatusRowIter) Close(ctx *sql.Context) error {
	return i.rowIter.Close(ctx)
}
