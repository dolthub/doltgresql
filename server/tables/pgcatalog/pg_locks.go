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

package pgcatalog

import (
	"fmt"
	"io"
	"sort"

	"github.com/dolthub/dolt/go/libraries/doltcore/rowjournal"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgLocksName is a constant to the pg_locks name.
const PgLocksName = "pg_locks"

// InitPgLocks handles registration of the pg_locks handler.
func InitPgLocks() {
	tables.AddHandler(PgCatalogName, PgLocksName, PgLocksHandler{})
}

// PgLocksHandler is the handler for the pg_locks table.
type PgLocksHandler struct{}

var _ tables.Handler = PgLocksHandler{}

// Name implements the interface tables.Handler.
func (p PgLocksHandler) Name() string {
	return PgLocksName
}

// RowIter implements the interface tables.Handler.
func (p PgLocksHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	journal := functions.RowJournal(ctx)
	if journal == nil {
		return emptyRowIter()
	}

	// Row locks and the locks that DDL and version control operations take on whole tables are reported. Other kinds
	// of lock that Postgres reports here, such as advisory locks, are not.
	locks := journal.Locks()
	sort.Slice(locks, func(i, j int) bool {
		if locks[i].Txn != locks[j].Txn {
			return locks[i].Txn < locks[j].Txn
		}

		return locks[i].Granted && !locks[j].Granted
	})

	rows := make([]sql.Row, len(locks))
	for i, lock := range locks {
		lockType := "tuple"
		if lock.Target.WholeTable {
			lockType = "relation"
		}

		schemaName := lock.Relation.Table.Schema
		if schemaName == "" {
			schemaName = "public"
		}

		rows[i] = sql.Row{
			lockType, // locktype
			nil,      // database
			id.NewTable(schemaName, lock.Relation.Table.Name).AsId(), // relation
			nil,              // page
			nil,              // tuple
			nil,              // virtualxid
			uint32(lock.Txn), // transactionid
			nil,              // classid
			nil,              // objid
			nil,              // objsubid
			fmt.Sprintf("%d/%d", lock.SessionID, lock.Txn), // virtualtransaction
			int32(lock.SessionID),                          // pid
			pgLockModeName(lock.Mode),                      // mode
			lock.Granted,                                   // granted
			false,                                          // fastpath
			nil,                                            // waitstart
		}
	}

	return sql.RowsToRowIter(rows...), nil
}

// pgLockModeName returns the name Postgres reports for the lock on a row that a row lock of the mode given
// corresponds to.
func pgLockModeName(mode rowjournal.LockMode) string {
	switch mode {
	case rowjournal.LockKeyShare:
		return "AccessShareLock"
	case rowjournal.LockShare:
		return "RowShareLock"
	case rowjournal.LockNoKeyUpdate:
		return "ExclusiveLock"
	case rowjournal.LockUpdate:
		return "AccessExclusiveLock"
	default:
		return mode.String()
	}
}

// PkSchema implements the interface tables.Handler.
func (p PgLocksHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgLocksSchema,
		PkOrdinals: nil,
	}
}

// pgLocksSchema is the schema for pg_locks.
var pgLocksSchema = sql.Schema{
	{Name: "locktype", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "database", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "relation", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "page", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "tuple", Type: pgtypes.Int16, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "virtualxid", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "transactionid", Type: pgtypes.Xid, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "classid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "objid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "objsubid", Type: pgtypes.Int16, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "virtualtransaction", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "pid", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "mode", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "granted", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "fastpath", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "waitstart", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgLocksName},
}

// pgLocksRowIter is the sql.RowIter for the pg_locks table.
type pgLocksRowIter struct {
}

var _ sql.RowIter = (*pgLocksRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgLocksRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgLocksRowIter) Close(ctx *sql.Context) error {
	return nil
}
