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

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/extensions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgAvailableExtensionsName is a constant to the pg_available_extensions name.
const PgAvailableExtensionsName = "pg_available_extensions"

// InitPgAvailableExtensions handles registration of the pg_available_extensions handler.
func InitPgAvailableExtensions() {
	tables.AddHandler(PgCatalogName, PgAvailableExtensionsName, PgAvailableExtensionsHandler{})
}

// PgAvailableExtensionsHandler is the handler for the pg_available_extensions table.
type PgAvailableExtensionsHandler struct{}

var _ tables.Handler = PgAvailableExtensionsHandler{}

// Name implements the interface tables.Handler.
func (p PgAvailableExtensionsHandler) Name() string {
	return PgAvailableExtensionsName
}

// pgAvailableExtension represents a row in the pg_available_extensions table.
type pgAvailableExtension struct {
	name             string
	defaultVersion   string
	installedVersion any
	comment          any
}

// RowIter implements the interface tables.Handler.
func (p PgAvailableExtensionsHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	allExtensions, err := extensions.GetAllExtensions()
	if err != nil {
		// Extensions cannot be loaded when there is no local Postgres installation, so we report that no extensions
		// are available rather than returning an error.
		return emptyRowIter()
	}
	extCollection, err := core.GetExtensionsCollectionFromContext(ctx, ctx.GetCurrentDatabase())
	if err != nil {
		return nil, err
	}
	availableExtensions := make([]pgAvailableExtension, 0, len(allExtensions))
	for name, extFiles := range allExtensions {
		availableExtension := pgAvailableExtension{
			name:           name,
			defaultVersion: extFiles.Control.DefaultVersion.String(),
		}
		if len(extFiles.Control.Comment) > 0 {
			availableExtension.comment = extFiles.Control.Comment
		}
		if installed, err := extCollection.GetLoadedExtension(ctx, id.NewExtension(name)); err != nil {
			return nil, err
		} else if installed.ExtName.IsValid() {
			availableExtension.installedVersion = installed.LibIdentifier.Version().String()
		}
		availableExtensions = append(availableExtensions, availableExtension)
	}
	sort.Slice(availableExtensions, func(i, j int) bool {
		return availableExtensions[i].name < availableExtensions[j].name
	})
	return &pgAvailableExtensionsRowIter{
		extensions: availableExtensions,
		idx:        0,
	}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgAvailableExtensionsHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgAvailableExtensionsSchema,
		PkOrdinals: nil,
	}
}

// pgAvailableExtensionsSchema is the schema for pg_available_extensions.
var pgAvailableExtensionsSchema = sql.Schema{
	{Name: "name", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgAvailableExtensionsName},
	{Name: "default_version", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgAvailableExtensionsName},
	{Name: "installed_version", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgAvailableExtensionsName}, // TODO: collation C
	{Name: "comment", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgAvailableExtensionsName},
}

// pgAvailableExtensionsRowIter is the sql.RowIter for the pg_available_extensions table.
type pgAvailableExtensionsRowIter struct {
	extensions []pgAvailableExtension
	idx        int
}

var _ sql.RowIter = (*pgAvailableExtensionsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgAvailableExtensionsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.extensions) {
		return nil, io.EOF
	}
	iter.idx++
	ext := iter.extensions[iter.idx-1]
	return sql.Row{
		ext.name,             // name
		ext.defaultVersion,   // default_version
		ext.installedVersion, // installed_version
		ext.comment,          // comment
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgAvailableExtensionsRowIter) Close(ctx *sql.Context) error {
	return nil
}
