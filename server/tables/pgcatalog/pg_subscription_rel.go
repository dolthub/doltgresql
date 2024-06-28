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

	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgSubscriptionRelName is a constant to the pg_subscription_rel name.
const PgSubscriptionRelName = "pg_subscription_rel"

// InitPgSubscriptionRel handles registration of the pg_subscription_rel handler.
func InitPgSubscriptionRel() {
	tables.AddHandler(PgCatalogName, PgSubscriptionRelName, PgSubscriptionRelHandler{})
}

// PgSubscriptionRelHandler is the handler for the pg_subscription_rel table.
type PgSubscriptionRelHandler struct{}

var _ tables.Handler = PgSubscriptionRelHandler{}

// Name implements the interface tables.Handler.
func (p PgSubscriptionRelHandler) Name() string {
	return PgSubscriptionRelName
}

// RowIter implements the interface tables.Handler.
func (p PgSubscriptionRelHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_subscription_rel row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgSubscriptionRelHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgSubscriptionRelSchema,
		PkOrdinals: nil,
	}
}

// pgSubscriptionRelSchema is the schema for pg_subscription_rel.
var pgSubscriptionRelSchema = sql.Schema{
	{Name: "srsubid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgSubscriptionRelName},
	{Name: "srrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgSubscriptionRelName},
	{Name: "srsubstate", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgSubscriptionRelName},
	{Name: "srsublsn", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSubscriptionRelName}, // TODO: pg_lsn type
}

// pgSubscriptionRelRowIter is the sql.RowIter for the pg_subscription_rel table.
type pgSubscriptionRelRowIter struct {
}

var _ sql.RowIter = (*pgSubscriptionRelRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgSubscriptionRelRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgSubscriptionRelRowIter) Close(ctx *sql.Context) error {
	return nil
}
