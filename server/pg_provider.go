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

package server

import (
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/tables"
)

// DoltgresDatabaseProvider wraps DoltDatabaseProvider to enforce PostgreSQL specific
// requirements, such as requiring all relation names (e.g. tables, indexes, sequences, views)
// are unique within a schema.
type DoltgresDatabaseProvider struct {
	*sqle.DoltDatabaseProvider
}

var _ sql.DatabaseProvider = (*DoltgresDatabaseProvider)(nil)

// Database overrides DoltDatabaseProvider.Database to wrap the returned sql.Database
// with PgDatabase, enabling relation-name uniqueness enforcement.
func (p *DoltgresDatabaseProvider) Database(ctx *sql.Context, name string) (sql.Database, error) {
	db, err := p.DoltDatabaseProvider.Database(ctx, name)
	if err != nil {
		return nil, err
	}
	return tables.WrapSqlDatabase(db), nil
}

// AllDatabases overrides DoltDatabaseProvider.AllDatabases to wrap each returned database
// with PgDatabase so that AllSchemas returns properly wrapped schemas (including pg_catalog
// virtual-table schemas) regardless of which code path iterates the catalog.
func (p *DoltgresDatabaseProvider) AllDatabases(ctx *sql.Context) []sql.Database {
	all := p.DoltDatabaseProvider.AllDatabases(ctx)
	for i, db := range all {
		all[i] = tables.WrapSqlDatabase(db)
	}
	return all
}

// UnderlyingDoltProvider implements sqle.DoltProviderUnwrapper so that NewSqlEngine can
// access the wrapped *DoltDatabaseProvider for Dolt-specific configuration.
func (p *DoltgresDatabaseProvider) UnderlyingDoltProvider() *sqle.DoltDatabaseProvider {
	return p.DoltDatabaseProvider
}

// DoltgresProviderFactory is the ProviderFactory for Doltgres. It embeds DoltProviderFactory
// and overrides NewProvider to wrap the result in DoltgresDatabaseProvider, so that
// Database() returns PgDatabase-wrapped databases enforcing PostgreSQL relation-name uniqueness.
type DoltgresProviderFactory struct {
	sqle.DoltProviderFactory
}

var _ sqle.ProviderFactory = DoltgresProviderFactory{}

// NewProvider overrides DoltProviderFactory.NewProvider to wrap the created provider in
// DoltgresDatabaseProvider before returning it.
func (f DoltgresProviderFactory) NewProvider(defaultBranch string, fs filesys.Filesys, databases []dsess.SqlDatabase, locations []filesys.Filesys, overrides sql.EngineOverrides) (sql.DatabaseProvider, error) {
	inner, err := f.DoltProviderFactory.NewProvider(defaultBranch, fs, databases, locations, overrides)
	if err != nil {
		return nil, err
	}
	return &DoltgresDatabaseProvider{inner.(*sqle.DoltDatabaseProvider)}, nil
}
