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

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgShadowName is a constant to the pg_shadow name.
const PgShadowName = "pg_shadow"

// InitPgShadow handles registration of the pg_shadow handler.
func InitPgShadow() {
	tables.AddHandler(PgCatalogName, PgShadowName, PgShadowHandler{})
}

// PgShadowHandler is the handler for the pg_shadow table.
type PgShadowHandler struct{}

var _ tables.Handler = PgShadowHandler{}

// Name implements the interface tables.Handler.
func (p PgShadowHandler) Name() string {
	return PgShadowName
}

// RowIter implements the interface tables.Handler.
func (p PgShadowHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	return &pgShadowRowIter{
		users: allLoginRoles(),
		idx:   0,
	}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgShadowHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgShadowSchema,
		PkOrdinals: nil,
	}
}

// pgShadowSchema is the schema for pg_shadow.
var pgShadowSchema = sql.Schema{
	{Name: "usename", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgShadowName},
	{Name: "usesysid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgShadowName},
	{Name: "usecreatedb", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgShadowName},
	{Name: "usesuper", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgShadowName},
	{Name: "userepl", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgShadowName},
	{Name: "usebypassrls", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgShadowName},
	{Name: "passwd", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgShadowName}, // TODO: collation C
	{Name: "valuntil", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgShadowName},
	{Name: "useconfig", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgShadowName}, // TODO: collation C
}

// pgShadowRowIter is the sql.RowIter for the pg_shadow table.
type pgShadowRowIter struct {
	users []auth.Role
	idx   int
}

var _ sql.RowIter = (*pgShadowRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgShadowRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.users) {
		return nil, io.EOF
	}
	iter.idx++
	role := iter.users[iter.idx-1]
	var passwd any
	if role.Password != nil {
		passwd = role.Password.AsPasswordString()
	}
	var valUntil any
	if role.ValidUntil != nil {
		valUntil = *role.ValidUntil
	}
	return sql.Row{
		role.Name,                      // usename
		roleOid(role.Name),             // usesysid
		role.CanCreateDB,               // usecreatedb
		role.IsSuperUser,               // usesuper
		role.IsReplicationRole,         // userepl
		role.CanBypassRowLevelSecurity, // usebypassrls
		passwd,                         // passwd
		valUntil,                       // valuntil
		nil,                            // useconfig (per-role settings are not supported)
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgShadowRowIter) Close(ctx *sql.Context) error {
	return nil
}
