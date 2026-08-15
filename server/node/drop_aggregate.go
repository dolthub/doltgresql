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

package node

import (
	"context"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// AggregateToDrop represents an aggregate in DROP AGGREGATE.
type AggregateToDrop struct {
	SchemaName string
	Name       string
	ArgTypes   []*pgtypes.DoltgresType
}

// DropAggregate implements DROP AGGREGATE.
type DropAggregate struct {
	Aggregates []*AggregateToDrop
	IfExists   bool
}

var _ sql.ExecSourceRel = (*DropAggregate)(nil)
var _ vitess.Injectable = (*DropAggregate)(nil)

// NewDropAggregate returns a new *DropAggregate.
func NewDropAggregate(ifExists bool, aggregates []*AggregateToDrop) *DropAggregate {
	return &DropAggregate{
		IfExists:   ifExists,
		Aggregates: aggregates,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (d *DropAggregate) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (d *DropAggregate) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (d *DropAggregate) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (d *DropAggregate) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	aggCollection, err := core.GetAggregatesCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	funcCollection, err := core.GetFunctionsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	for _, agg := range d.Aggregates {
		schemaName, err := core.GetSchemaName(ctx, nil, agg.SchemaName)
		if err != nil {
			return nil, err
		}
		argIDs := make([]id.Type, len(agg.ArgTypes))
		argNames := make([]string, len(agg.ArgTypes))
		for i, argType := range agg.ArgTypes {
			argIDs[i] = argType.ID
			argNames[i] = argType.String()
		}
		aggregateID := id.NewFunction(schemaName, agg.Name, argIDs...)
		if !aggCollection.HasAggregate(ctx, aggregateID) {
			if funcCollection.HasFunction(ctx, aggregateID) {
				return nil, errors.Errorf("function %s(%s) is not an aggregate", agg.Name, strings.Join(argNames, ", "))
			}
			if d.IfExists {
				continue
			}
			return nil, errors.Errorf("aggregate %s(%s) does not exist", agg.Name, strings.Join(argNames, ", "))
		}
		if err = aggCollection.DropAggregate(ctx, aggregateID); err != nil {
			return nil, err
		}
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (d *DropAggregate) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (d *DropAggregate) String() string {
	return "DROP AGGREGATE"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (d *DropAggregate) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(d, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (d *DropAggregate) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return d, nil
}
