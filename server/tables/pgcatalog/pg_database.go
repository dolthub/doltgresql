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
	"io"
	"sort"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/server/types/oid"
)

// PgDatabaseName is a constant to the pg_database name.
const PgDatabaseName = "pg_database"

// InitPgDatabase handles registration of the pg_database handler.
func InitPgDatabase() {
	tables.AddHandler(PgCatalogName, PgDatabaseName, PgDatabaseHandler{})
}

// PgDatabaseHandler is the handler for the pg_database table.
type PgDatabaseHandler struct{}

var _ tables.Handler = PgDatabaseHandler{}

// Name implements the interface tables.Handler.
func (p PgDatabaseHandler) Name() string {
	return PgDatabaseName
}

// RowIter implements the interface tables.Handler.
func (p PgDatabaseHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Should the catalog be passed to RowIter like it is for the information_schema tables RowIter?
	doltSession := dsess.DSessFromSess(ctx.Session)
	c := sqle.NewDefault(doltSession.Provider()).Analyzer.Catalog

	databases := c.AllDatabases(ctx)
	dbs := make([]sql.Database, 0, len(databases))
	for _, db := range databases {
		name := db.Name()
		if name == "information_schema" || name == "pg_catalog" || name == "performance_schema" {
			continue
		}
		dbs = append(dbs, db)
	}
	sort.Slice(dbs, func(i, j int) bool {
		return dbs[i].Name() < dbs[j].Name()
	})

	return &pgDatabaseRowIter{
		dbs: dbs,
		idx: 0,
	}, nil
}

// Schema implements the interface tables.Handler.
func (p PgDatabaseHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgDatabaseSchema,
		PkOrdinals: nil,
	}
}

// pgDatabaseSchema is the schema for pg_database.
var pgDatabaseSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "datname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "datdba", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "encoding", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "datlocprovider", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "datistemplate", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "datallowconn", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "datconnlimit", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "datfrozenxid", Type: pgtypes.Xid, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "datminmxid", Type: pgtypes.Xid, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "dattablespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "datcollate", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgDatabaseName},  // TODO: collation C
	{Name: "datctype", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgDatabaseName},    // TODO: collation C
	{Name: "daticulocale", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgDatabaseName}, // TODO: collation C
	{Name: "daticurules", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "datcollversion", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgDatabaseName}, // TODO: collation C
	{Name: "datacl", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgDatabaseName},    // TODO: type aclitem[]
}

// pgDatabaseRowIter is the sql.RowIter for the pg_database table.
type pgDatabaseRowIter struct {
	dbs []sql.Database
	idx int
}

var _ sql.RowIter = (*pgDatabaseRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgDatabaseRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.dbs) {
		return nil, io.EOF
	}
	iter.idx++
	db := iter.dbs[iter.idx-1]
	dbOid := oid.CreateOID(oid.Section_Database, 0, iter.idx-1)

	// TODO: Add the rest of the pg_database columns
	return sql.Row{
		dbOid,     // oid
		db.Name(), // datname
		uint32(0), // datdba
		int32(6),  // encoding
		"i",       // datlocprovider
		false,     // datistemplate
		true,      // datallowconn
		int32(-1), // datconnlimit
		uint32(0), // datfrozenxid
		uint32(0), // datminmxid
		uint32(0), // dattablespace
		"",        // datcollate
		"",        // datctype
		nil,       // daticulocale
		"",        // daticurules
		nil,       // datcollversion
		nil,       // datacl
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgDatabaseRowIter) Close(ctx *sql.Context) error {
	return nil
}
