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

// PgAvailableExtensionVersionsName is a constant to the pg_available_extension_versions name.
const PgAvailableExtensionVersionsName = "pg_available_extension_versions"

// InitPgAvailableExtensionVersions handles registration of the pg_available_extension_versions handler.
func InitPgAvailableExtensionVersions() {
	tables.AddHandler(PgCatalogName, PgAvailableExtensionVersionsName, PgAvailableExtensionVersionsHandler{})
}

// PgAvailableExtensionVersionsHandler is the handler for the pg_available_extension_versions table.
type PgAvailableExtensionVersionsHandler struct{}

var _ tables.Handler = PgAvailableExtensionVersionsHandler{}

// Name implements the interface tables.Handler.
func (p PgAvailableExtensionVersionsHandler) Name() string {
	return PgAvailableExtensionVersionsName
}

// pgAvailableExtensionVersion represents a row in the pg_available_extension_versions table.
type pgAvailableExtensionVersion struct {
	name        string
	version     string
	installed   bool
	superuser   bool
	trusted     bool
	relocatable bool
	schema      any
	requires    any
	comment     any
}

// RowIter implements the interface tables.Handler.
func (p PgAvailableExtensionVersionsHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
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
	// TODO: Postgres lists a row for every version that has an installation script, but we only track the default
	//  version for each extension.
	extVersions := make([]pgAvailableExtensionVersion, 0, len(allExtensions))
	for name, extFiles := range allExtensions {
		extVersion := pgAvailableExtensionVersion{
			name:        name,
			version:     extFiles.Control.DefaultVersion.String(),
			superuser:   extFiles.Control.Superuser,
			trusted:     extFiles.Control.Trusted,
			relocatable: extFiles.Control.Relocatable,
		}
		if len(extFiles.Control.Schema) > 0 {
			extVersion.schema = extFiles.Control.Schema
		}
		if len(extFiles.Control.Requires) > 0 {
			requires := make([]any, len(extFiles.Control.Requires))
			for i, req := range extFiles.Control.Requires {
				requires[i] = req
			}
			extVersion.requires = requires
		}
		if len(extFiles.Control.Comment) > 0 {
			extVersion.comment = extFiles.Control.Comment
		}
		if installed, err := extCollection.GetLoadedExtension(ctx, id.NewExtension(name)); err != nil {
			return nil, err
		} else if installed.ExtName.IsValid() {
			extVersion.installed = installed.LibIdentifier.Version() == extFiles.Control.DefaultVersion
		}
		extVersions = append(extVersions, extVersion)
	}
	sort.Slice(extVersions, func(i, j int) bool {
		return extVersions[i].name < extVersions[j].name
	})
	return &pgAvailableExtensionVersionsRowIter{
		extensions: extVersions,
		idx:        0,
	}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgAvailableExtensionVersionsHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgAvailableExtensionVersionsSchema,
		PkOrdinals: nil,
	}
}

// pgAvailableExtensionVersionsSchema is the schema for pg_available_extension_versions.
var pgAvailableExtensionVersionsSchema = sql.Schema{
	{Name: "name", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgAvailableExtensionVersionsName},
	{Name: "version", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgAvailableExtensionVersionsName},
	{Name: "installed", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgAvailableExtensionVersionsName},
	{Name: "superuser", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgAvailableExtensionVersionsName},
	{Name: "trusted", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgAvailableExtensionVersionsName},
	{Name: "relocatable", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgAvailableExtensionVersionsName},
	{Name: "schema", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgAvailableExtensionVersionsName},
	{Name: "requires", Type: pgtypes.NameArray, Default: nil, Nullable: true, Source: PgAvailableExtensionVersionsName},
	{Name: "comment", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgAvailableExtensionVersionsName},
}

// pgAvailableExtensionVersionsRowIter is the sql.RowIter for the pg_available_extension_versions table.
type pgAvailableExtensionVersionsRowIter struct {
	extensions []pgAvailableExtensionVersion
	idx        int
}

var _ sql.RowIter = (*pgAvailableExtensionVersionsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgAvailableExtensionVersionsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.extensions) {
		return nil, io.EOF
	}
	iter.idx++
	ext := iter.extensions[iter.idx-1]
	return sql.Row{
		ext.name,        // name
		ext.version,     // version
		ext.installed,   // installed
		ext.superuser,   // superuser
		ext.trusted,     // trusted
		ext.relocatable, // relocatable
		ext.schema,      // schema
		ext.requires,    // requires
		ext.comment,     // comment
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgAvailableExtensionVersionsRowIter) Close(ctx *sql.Context) error {
	return nil
}
