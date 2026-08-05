// Copyright 2026 Dolthub, Inc.
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

package functions

import (
	"io"
	"sort"

	"github.com/dolthub/doltgresql/core/extensions"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPgAvailableExtensionVersions registers the functions to the catalog.
func initPgAvailableExtensionVersions() {
	framework.RegisterFunction(pg_available_extension_versions)
}

// pgAvailableExtensionVersionsName is the name for pg_available_extension_versions function.
const pgAvailableExtensionVersionsName = "pg_available_extension_versions"

// pg_available_extension_versions represents the PostgreSQL function of the same name, taking the same parameters.
var pg_available_extension_versions = framework.Function0{
	Name:               pgAvailableExtensionVersionsName,
	Return:             pgtypes.Record, // SETOF record
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context) (any, error) {
		allExtensions, err := extensions.GetAllExtensions()
		if err != nil {
			// Extensions cannot be loaded when there is no local Postgres installation, so we report that no extensions
			// are available rather than returning an error.
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
			extVersions = append(extVersions, extVersion)
		}
		sort.Slice(extVersions, func(i, j int) bool {
			return extVersions[i].name < extVersions[j].name
		})

		idx := 0
		numExtVersions := len(extVersions)
		return pgtypes.NewSetReturningFunctionRowIter(func(ctx *sql.Context) (sql.Row, error) {
			defer func() { idx++ }()
			if idx >= numExtVersions {
				return nil, io.EOF
			}
			ext := extVersions[idx]
			return sql.Row{
				ext.name,        // name
				ext.version,     // version
				ext.superuser,   // superuser
				ext.trusted,     // trusted
				ext.relocatable, // relocatable
				ext.schema,      // schema
				ext.requires,    // requires
				ext.comment,     // comment
			}, nil
		}), nil
	},
	OutParams: pgAvailableExtensionVersionsSchema,
}

// pgAvailableExtensionVersion represents a row in the pg_available_extension_versions table.
type pgAvailableExtensionVersion struct {
	name        string
	version     string
	superuser   bool
	trusted     bool
	relocatable bool
	schema      any
	requires    any
	comment     any
}

// pgAvailableExtensionVersionsSchema is the schema for pg_available_extension_versions.
var pgAvailableExtensionVersionsSchema = sql.Schema{
	{Name: "name", Type: pgtypes.Name, Default: nil, Nullable: true, Source: pgAvailableExtensionVersionsName},
	{Name: "version", Type: pgtypes.Text, Default: nil, Nullable: true, Source: pgAvailableExtensionVersionsName},
	{Name: "superuser", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: pgAvailableExtensionVersionsName},
	{Name: "trusted", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: pgAvailableExtensionVersionsName},
	{Name: "relocatable", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: pgAvailableExtensionVersionsName},
	{Name: "schema", Type: pgtypes.Name, Default: nil, Nullable: true, Source: pgAvailableExtensionVersionsName},
	{Name: "requires", Type: pgtypes.NameArray, Default: nil, Nullable: true, Source: pgAvailableExtensionVersionsName},
	{Name: "comment", Type: pgtypes.Text, Default: nil, Nullable: true, Source: pgAvailableExtensionVersionsName},
}
