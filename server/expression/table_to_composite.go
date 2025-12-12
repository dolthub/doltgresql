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

package expression

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// TableToComposite is a set of sql.Expressions wrapped together in a single value.
type TableToComposite struct {
	fields []sql.Expression
	typ    *pgtypes.DoltgresType
}

var _ sql.Expression = (*TableToComposite)(nil)

// NewTableToComposite creates a new composite table type.
func NewTableToComposite(ctx *sql.Context, tableName string, fields []sql.Expression) (sql.Expression, error) {
	coll, err := core.GetTypesCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	// TODO: we need to get the schema, but the GMS builder doesn't have that information
	typ, err := coll.GetType(ctx, id.NewType("", tableName))
	if err != nil {
		return nil, err
	}
	if typ == nil {
		return nil, errors.New(fmt.Sprintf(`could not create a composite type for table "%s"`, tableName))
	}
	return &TableToComposite{
		fields: fields,
		typ:    typ,
	}, nil
}

// Resolved implements the sql.Expression interface.
func (t *TableToComposite) Resolved() bool {
	for _, expr := range t.fields {
		if !expr.Resolved() {
			return false
		}
	}
	return true
}

// String implements the sql.Expression interface.
func (t *TableToComposite) String() string {
	return "TABLE TO COMPOSITE"
}

// Type implements the sql.Expression interface.
func (t *TableToComposite) Type() sql.Type {
	return t.typ
}

// IsNullable implements the sql.Expression interface.
func (t *TableToComposite) IsNullable() bool {
	return false
}

// Eval implements the sql.Expression interface.
func (t *TableToComposite) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	vals := make([]pgtypes.RecordValue, len(t.fields))
	for i, expr := range t.fields {
		val, err := expr.Eval(ctx, row)
		if err != nil {
			return nil, err
		}

		typ, ok := expr.Type().(*pgtypes.DoltgresType)
		if !ok {
			return nil, fmt.Errorf("expected a DoltgresType, but got %T", expr.Type())
		}
		vals[i] = pgtypes.RecordValue{
			Value: val,
			Type:  typ,
		}
	}

	return vals, nil
}

// Children implements the sql.Expression interface.
func (t *TableToComposite) Children() []sql.Expression {
	return t.fields
}

// WithChildren implements the sql.Expression interface.
func (t *TableToComposite) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	tCopy := *t
	tCopy.fields = children
	return &tCopy, nil
}
