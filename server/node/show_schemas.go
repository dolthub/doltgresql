// Copyright 2025 Dolthub, Inc.
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

package node

import (
	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/server/types"

	"github.com/dolthub/go-mysql-server/sql"
)

// ShowSchemas is a node that implements the SHOW SCHEMAS	statement.
type ShowSchemas struct {
	// TODO: we need planbuilder integration to support SHOW SCHEMAS, rather than getting everything at runtime
	database string
}

var _ sql.ExecSourceRel = (*ShowSchemas)(nil)

// NewShowSchemas returns a new *ShowSchemas.
func NewShowSchemas(database string) *ShowSchemas {
	return &ShowSchemas{
		database: database,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (s *ShowSchemas) Children() []sql.Node {
	return []sql.Node{}
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (s *ShowSchemas) IsReadOnly() bool {
	return true
}

// Resolved implements the interface sql.ExecSourceRel.
func (s *ShowSchemas) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (s *ShowSchemas) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	database := s.database
	if database == "" {
		database = ctx.GetCurrentDatabase()
		if database == "" {
			return nil, errors.New("no database selected (this is a bug)")
		}
	}

	db, err := core.GetSqlDatabaseFromContext(ctx, database)
	if err != nil {
		return nil, err
	}

	if db == nil {
		return nil, errors.New("database not found: " + database)
	}

	sdb, ok := db.(sql.SchemaDatabase)
	if !ok {
		// This handles any database that doesn't support schemas (such as some system databases)
		// TODO: mirror the postgres behavior of returning, every database should have schemas
		return sql.RowsToRowIter(), nil
	}

	schemas, err := sdb.AllSchemas(ctx)
	if err != nil {
		return nil, err
	}

	rows := make([]sql.Row, len(schemas))
	for i, schema := range schemas {
		rows[i] = sql.Row{schema.SchemaName()}
	}

	return sql.RowsToRowIter(rows...), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (s *ShowSchemas) Schema() sql.Schema {
	return sql.Schema{
		{Name: "schema_name", Type: types.Text, Source: "show schemas"},
	}
}

// String implements the interface sql.ExecSourceRel.
func (s *ShowSchemas) String() string {
	if s.database == "" {
		return "SHOW SCHEMAS FROM " + s.database
	}
	return "SHOW SCHEMAS"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (s *ShowSchemas) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 0 {
		return nil, errors.New("SHOW SCHEMAS does not support children")
	}
	return s, nil
}

// WithResolvedChildren implements the interface vitess.InjectedStatement.
func (s *ShowSchemas) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, errors.New("SHOW SCHEMAS does not support children")
	}
	return s, nil
}
