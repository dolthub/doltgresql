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

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/operators"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgOperatorName is a constant to the pg_operator name.
const PgOperatorName = "pg_operator"

// InitPgOperator handles registration of the pg_operator handler.
func InitPgOperator() {
	tables.AddHandler(PgCatalogName, PgOperatorName, PgOperatorHandler{})
}

// PgOperatorHandler is the handler for the pg_operator table.
type PgOperatorHandler struct{}

var _ tables.Handler = PgOperatorHandler{}

// Name implements the interface tables.Handler.
func (p PgOperatorHandler) Name() string {
	return PgOperatorName
}

// RowIter implements the interface tables.Handler.
func (p PgOperatorHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: fill this in from the operator framework's built-in operators
	operatorCollection, err := core.GetOperatorsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	var ops []operators.Operator
	err = operatorCollection.IterateOperators(ctx, func(o operators.Operator) (stop bool, err error) {
		ops = append(ops, o)
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	return &pgOperatorRowIter{ops: ops}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgOperatorHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgOperatorSchema,
		PkOrdinals: nil,
	}
}

// pgOperatorSchema is the schema for pg_operator.
var pgOperatorSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOperatorName},
	{Name: "oprname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgOperatorName},
	{Name: "oprnamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOperatorName},
	{Name: "oprowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOperatorName},
	{Name: "oprkind", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgOperatorName},
	{Name: "oprcanmerge", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgOperatorName},
	{Name: "oprcanhash", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgOperatorName},
	{Name: "oprleft", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOperatorName},
	{Name: "oprright", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOperatorName},
	{Name: "oprresult", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOperatorName},
	{Name: "oprcom", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOperatorName},
	{Name: "oprnegate", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgOperatorName},
	{Name: "oprcode", Type: pgtypes.Regproc, Default: nil, Nullable: false, Source: PgOperatorName},
	{Name: "oprrest", Type: pgtypes.Regproc, Default: nil, Nullable: false, Source: PgOperatorName},
	{Name: "oprjoin", Type: pgtypes.Regproc, Default: nil, Nullable: false, Source: PgOperatorName},
}

// pgOperatorRowIter is the sql.RowIter for the pg_operator table.
type pgOperatorRowIter struct {
	ops []operators.Operator
	idx int
}

var _ sql.RowIter = (*pgOperatorRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgOperatorRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.ops) {
		return nil, io.EOF
	}
	op := iter.ops[iter.idx]
	iter.idx++
	commutator := id.Null
	if len(op.Commutator) > 0 {
		commutator = id.NewOperator(op.ID.SchemaName(), op.Commutator, op.ID.RightType(), op.ID.LeftType()).AsId()
	}
	negator := id.Null
	if len(op.Negator) > 0 {
		negator = id.NewOperator(op.ID.SchemaName(), op.Negator, op.ID.LeftType(), op.ID.RightType()).AsId()
	}
	return sql.Row{
		op.ID.AsId(),   // oid
		op.ID.Symbol(), // oprname
		id.NewNamespace(op.ID.SchemaName()).AsId(), // oprnamespace
		id.Null,                  // oprowner
		"b",                      // oprkind
		op.Merges,                // oprcanmerge
		op.Hashes,                // oprcanhash
		op.ID.LeftType().AsId(),  // oprleft
		op.ID.RightType().AsId(), // oprright
		op.ReturnType.AsId(),     // oprresult
		commutator,               // oprcom
		negator,                  // oprnegate
		op.Function.AsId(),       // oprcode
		id.Null,                  // oprrest
		id.Null,                  // oprjoin
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgOperatorRowIter) Close(ctx *sql.Context) error {
	return nil
}
