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

// PgSubscriptionName is a constant to the pg_subscription name.
const PgSubscriptionName = "pg_subscription"

// InitPgSubscription handles registration of the pg_subscription handler.
func InitPgSubscription() {
	tables.AddHandler(PgCatalogName, PgSubscriptionName, PgSubscriptionHandler{})
}

// PgSubscriptionHandler is the handler for the pg_subscription table.
type PgSubscriptionHandler struct{}

var _ tables.Handler = PgSubscriptionHandler{}

// Name implements the interface tables.Handler.
func (p PgSubscriptionHandler) Name() string {
	return PgSubscriptionName
}

// RowIter implements the interface tables.Handler.
func (p PgSubscriptionHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_subscription row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgSubscriptionHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgSubscriptionSchema,
		PkOrdinals: nil,
	}
}

// pgSubscriptionSchema is the schema for pg_subscription.
var pgSubscriptionSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgSubscriptionName},
	{Name: "subdbid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgSubscriptionName},
	{Name: "subskiplsn", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgSubscriptionName}, // TODO: pg_lsn type
	{Name: "subname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgSubscriptionName},
	{Name: "subowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgSubscriptionName},
	{Name: "subenabled", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgSubscriptionName},
	{Name: "subbinary", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgSubscriptionName},
	{Name: "substream", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgSubscriptionName},
	{Name: "subtwophasestate", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgSubscriptionName},
	{Name: "subdisableonerr", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgSubscriptionName},
	{Name: "subconninfo", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgSubscriptionName}, // TODO: collation C
	{Name: "subslotname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgSubscriptionName},
	{Name: "subsynccommit", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgSubscriptionName},        // TODO: collation C
	{Name: "subpublications", Type: pgtypes.TextArray, Default: nil, Nullable: false, Source: PgSubscriptionName}, // TODO: collation C
}

// pgSubscriptionRowIter is the sql.RowIter for the pg_subscription table.
type pgSubscriptionRowIter struct {
}

var _ sql.RowIter = (*pgSubscriptionRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgSubscriptionRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgSubscriptionRowIter) Close(ctx *sql.Context) error {
	return nil
}
