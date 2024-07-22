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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/server/types/oid"
)

// PgIndexesName is a constant to the pg_indexes name.
const PgIndexesName = "pg_indexes"

// InitPgIndexes handles registration of the pg_indexes handler.
func InitPgIndexes() {
	tables.AddHandler(PgCatalogName, PgIndexesName, PgIndexesHandler{})
}

// PgIndexesHandler is the handler for the pg_indexes table.
type PgIndexesHandler struct{}

var _ tables.Handler = PgIndexesHandler{}

// Name implements the interface tables.Handler.
func (p PgIndexesHandler) Name() string {
	return PgIndexesName
}

// RowIter implements the interface tables.Handler.
func (p PgIndexesHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	var indexes []sql.Index
	var schemas []string

	err := oid.IterateCurrentDatabase(ctx, oid.Callbacks{
		Index: func(ctx *sql.Context, schema oid.ItemSchema, table oid.ItemTable, index oid.ItemIndex) (cont bool, err error) {
			indexes = append(indexes, index.Item)
			schemas = append(schemas, schema.Item.SchemaName())
			return true, nil
		},
	})
	if err != nil {
		return nil, err
	}

	return &pgIndexesRowIter{
		indexes: indexes,
		schemas: schemas,
		idx:     0,
	}, nil
}

// Schema implements the interface tables.Handler.
func (p PgIndexesHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgIndexesSchema,
		PkOrdinals: nil,
	}
}

// pgIndexesSchema is the schema for pg_indexes.
var pgIndexesSchema = sql.Schema{
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgIndexesName},
	{Name: "tablename", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgIndexesName},
	{Name: "indexname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgIndexesName},
	{Name: "tablespace", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgIndexesName},
	{Name: "indexdef", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgIndexesName},
}

// pgIndexesRowIter is the sql.RowIter for the pg_indexes table.
type pgIndexesRowIter struct {
	indexes []sql.Index
	schemas []string
	idx     int
}

var _ sql.RowIter = (*pgIndexesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgIndexesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.indexes) {
		return nil, io.EOF
	}
	iter.idx++
	index := iter.indexes[iter.idx-1]
	schema := iter.schemas[iter.idx-1]

	// TODO: Fill in the rest of the pg_indexes columns
	return sql.Row{
		schema,                     // schemaname
		index.Table(),              // tablename
		getIndexName(index),        // indexname
		"",                         // tablespace
		getIndexDef(index, schema), // indexdef
	}, nil
}

// getIndexName returns the definition of the index.
func getIndexDef(index sql.Index, schema string) string {
	name := getIndexName(index)
	using := strings.ToLower(index.IndexType())
	unique := ""
	if index.IsUnique() {
		unique = " UNIQUE"
	}

	cols := make([]string, len(index.Expressions()))
	for i, expr := range index.Expressions() {
		split := strings.Split(expr, ".")
		if len(split) > 1 {
			cols[i] = split[1]
		} else {
			cols[i] = expr
		}
	}
	colsStr := strings.Join(cols, ", ")

	return fmt.Sprintf("CREATE%s INDEX %s ON %s.%s USING %s (%s)", unique, name, schema, index.Table(), using, colsStr)
}

// Close implements the interface sql.RowIter.
func (iter *pgIndexesRowIter) Close(ctx *sql.Context) error {
	return nil
}
