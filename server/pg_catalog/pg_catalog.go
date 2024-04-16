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

package pg_catalog

import (
	"bytes"
	"fmt"
	"github.com/dolthub/go-mysql-server/sql"
	"io"
)

const (
	// PgCatalogName is the name of the pg_catalog system schema.
	PgCatalogName = "pg_catalog"
)

var (
	_ sql.Database      = (*pgCatalogDatabase)(nil)
	_ sql.Table         = (*pgCatalogTable)(nil)
	_ sql.Databaseable  = (*pgCatalogTable)(nil)
	_ sql.CatalogTable  = (*pgCatalogTable)(nil)
	_ sql.Partition     = (*pgCatalogPartition)(nil)
	_ sql.PartitionIter = (*pgCatalogPartitionIter)(nil)
)

type pgCatalogDatabase struct {
	name   string
	tables map[string]sql.Table
}

func (pdb *pgCatalogDatabase) Name() string {
	return pdb.name
}

func (pdb *pgCatalogDatabase) GetTableInsensitive(ctx *sql.Context, tblName string) (sql.Table, bool, error) {
	tbl, ok := sql.GetTableInsensitive(tblName, pdb.tables)
	return tbl, ok, nil
}

func (pdb *pgCatalogDatabase) GetTableNames(ctx *sql.Context) ([]string, error) {
	tblNames := make([]string, 0, len(pdb.tables))
	for k := range pdb.tables {
		tblNames = append(tblNames, k)
	}
	return tblNames, nil
}

type pgCatalogTable struct {
	name    string
	schema  sql.Schema
	catalog sql.Catalog
	reader  func(*sql.Context, sql.Catalog) (sql.RowIter, error)
}

func (pt *pgCatalogTable) Name() string {
	return pt.name
}

func (pt *pgCatalogTable) String() string {
	return printTable(pt.Name(), pt.Schema())
}

func (pt *pgCatalogTable) Schema() sql.Schema {
	return pt.schema
}

func (pt *pgCatalogTable) Collation() sql.CollationID {
	// TODO: the default collation of pg_catalog is 'C'.
	return sql.Collation_Information_Schema_Default
}

func (pt *pgCatalogTable) Partitions(ctx *sql.Context) (sql.PartitionIter, error) {
	return &pgCatalogPartitionIter{pgCatalogPartition: pgCatalogPartition{partitionKey(pt.Name())}}, nil
}

func (pt *pgCatalogTable) PartitionRows(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	if !bytes.Equal(partition.Key(), partitionKey(pt.Name())) {
		return nil, sql.ErrPartitionNotFound.New(partition.Key())
	}
	if pt.reader == nil {
		return sql.RowsToRowIter(), nil
	}
	if pt.catalog == nil {
		return nil, fmt.Errorf("nil catalog for info schema table %s", pt.name)
	}
	return pt.reader(ctx, pt.catalog)
}

// Database implements the sql.Databaseable interface.
func (pt *pgCatalogTable) Database() string {
	return PgCatalogName
}

func (pt *pgCatalogTable) AssignCatalog(cat sql.Catalog) sql.Table {
	pt.catalog = cat
	return pt
}

type pgCatalogPartition struct {
	key []byte
}

func (pp *pgCatalogPartition) Key() []byte {
	return pp.key
}

type pgCatalogPartitionIter struct {
	pgCatalogPartition
	pos int
}

func (ppi *pgCatalogPartitionIter) Close(cxt *sql.Context) error {
	ppi.pos = 0
	return nil
}

func (ppi *pgCatalogPartitionIter) Next(cxt *sql.Context) (sql.Partition, error) {
	if ppi.pos == 0 {
		ppi.pos++
		return ppi, nil
	}
	return nil, io.EOF
}

// emptyRowIter implements the sql.RowIter for empty table.
func emptyRowIter(ctx *sql.Context, c sql.Catalog) (sql.RowIter, error) {
	return sql.RowsToRowIter(), nil
}

// NewPgCatalogDatabase creates a new pg_catalog Database.
func NewPgCatalogDatabase() sql.Database {
	return pgCatalogDb
}

func printTable(name string, tableSchema sql.Schema) string {
	p := sql.NewTreePrinter()
	_ = p.WriteNode("Table(%s)", name)
	var schema = make([]string, len(tableSchema))
	for i, col := range tableSchema {
		schema[i] = fmt.Sprintf(
			"Column(%s, %s, nullable=%v)",
			col.Name,
			col.Type.String(),
			col.Nullable,
		)
	}
	_ = p.WriteChildren(schema...)
	return p.String()
}

func partitionKey(tableName string) []byte {
	return []byte(PgCatalogName + "." + tableName)
}
