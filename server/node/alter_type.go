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
	"fmt"
	"maps"
	"sort"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/schema"
	"github.com/dolthub/dolt/go/libraries/doltcore/schema/typeinfo"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/types"
)

// AlterTypeAddValue adds a label without changing existing enum identities.
type AlterTypeAddValue struct {
	SchemaName, Name, Value string
	IfNotExists             bool
	Placement               *tree.AlterTypeAddValuePlacement
}

var _ sql.ExecSourceRel = (*AlterTypeAddValue)(nil)
var _ vitess.Injectable = (*AlterTypeAddValue)(nil)

func (a *AlterTypeAddValue) Children() []sql.Node { return nil }

func (a *AlterTypeAddValue) IsReadOnly() bool { return false }

func (a *AlterTypeAddValue) Resolved() bool { return true }

func (a *AlterTypeAddValue) Schema(ctx *sql.Context) sql.Schema { return nil }

func (a *AlterTypeAddValue) String() string {
	return fmt.Sprintf("ALTER TYPE %s ADD VALUE %q", a.Name, a.Value)
}

func (a *AlterTypeAddValue) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(a, children...)
}

func (a *AlterTypeAddValue) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return a, nil
}

func (a *AlterTypeAddValue) RowIter(ctx *sql.Context, _ sql.Row) (sql.RowIter, error) {
	var err error
	auth.LockRead(func() { _, err = auth.CurrentRoleLocked(ctx) })
	if err != nil {
		return nil, err
	}
	collection, err := core.GetTypesCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	schemas := []string{a.SchemaName}
	if a.SchemaName == "" {
		schemas, err = core.SearchPath(ctx)
		if err != nil {
			return nil, err
		}
	}
	var typ *types.DoltgresType
	for _, schemaName := range schemas {
		typ, err = collection.GetType(ctx, id.NewType(schemaName, a.Name))
		if err != nil {
			return nil, err
		}
		if typ != nil {
			break
		}
	}
	if typ == nil {
		return nil, pgerror.Newf(pgcode.UndefinedObject, "unable to resolve type `%s`", a.Name)
	}
	if typ.TypType != types.TypeType_Enum {
		return nil, pgerror.Newf(pgcode.WrongObjectType, "%s is not an enum", typ.Name())
	}
	if len(a.Value) > 63 {
		return nil, pgerror.New(pgcode.InvalidName, "invalid enum label: labels must be 63 bytes or less")
	}
	if _, ok := typ.EnumLabels[a.Value]; ok {
		if !a.IfNotExists {
			return nil, pgerror.Newf(pgcode.DuplicateObject, "enum label %q already exists", a.Value)
		}
		dsess.DSessFromSess(ctx.Session).Notice(&pgproto3.NoticeResponse{Severity: "NOTICE", Code: "42710", Message: fmt.Sprintf("enum label %q already exists, skipping", a.Value)})
		return sql.RowsToRowIter(), nil
	}
	labels := maps.Clone(typ.EnumLabels)
	if labels == nil {
		labels = make(map[string]types.EnumLabel)
	}
	ordered := make([]types.EnumLabel, 0, len(labels))
	for _, label := range labels {
		ordered = append(ordered, label)
	}
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].SortOrder < ordered[j].SortOrder })
	at := len(ordered)
	if a.Placement != nil {
		at = -1
		for i, label := range ordered {
			if label.ID.Label() == a.Placement.ExistingVal {
				at = i
				break
			}
		}
		if at == -1 {
			return nil, pgerror.Newf(pgcode.InvalidParameterValue, "%q is not an existing enum label", a.Placement.ExistingVal)
		}
		if !a.Placement.Before {
			at++
		}
	}
	order := func() float32 {
		if len(ordered) == 0 {
			return 1
		}
		if at == 0 {
			return ordered[0].SortOrder - 1
		}
		if at == len(ordered) {
			return ordered[at-1].SortOrder + 1
		}
		return (ordered[at-1].SortOrder + ordered[at].SortOrder) / 2
	}
	position := order()
	// Repeated insertions can exhaust float32 spacing. Renumber metadata only;
	// label IDs and serialized row values remain unchanged.
	if (at > 0 && position <= ordered[at-1].SortOrder) || (at < len(ordered) && position >= ordered[at].SortOrder) {
		for i := range ordered {
			ordered[i].SortOrder = float32(i + 1)
			labels[ordered[i].ID.Label()] = ordered[i]
		}
		position = order()
	}
	labels[a.Value] = types.NewEnumLabel(ctx, id.NewEnumLabel(typ.ID, a.Value), position)

	updated := typ.Copy()
	updated.EnumLabels = labels
	updated.Array = typ.Array.Copy()
	updated.Array.Elem = updated
	session, root, err := core.GetRootFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var newRoot doltdb.RootValue = root
	// Dolt's index comparisons can run without a SQL context. Refresh the stored
	// column definitions during DDL, keeping the row and index data unchanged.
	err = root.IterTables(ctx, func(name doltdb.TableName, table *doltdb.Table, sch schema.Schema) (bool, error) {
		cols := append([]schema.Column(nil), sch.GetAllCols().GetColumns()...)
		changed := false
		for i := range cols {
			colType, ok := cols[i].TypeInfo.ToSqlType().(*types.DoltgresType)
			if !ok {
				continue
			}
			replacement, replaced := replaceEnumType(colType, updated, make(map[*types.DoltgresType]bool))
			if replaced {
				cols[i].TypeInfo = typeinfo.CreateExtendedTypeFromSqlType(replacement).WithEncoding(cols[i].TypeInfo.Encoding())
				changed = true
			}
		}
		if !changed {
			return false, nil
		}
		updatedSchema, err := schema.NewSchema(schema.NewColCollection(cols...), sch.GetPkOrdinals(), sch.GetCollation(), nil, sch.Checks().Copy())
		if err != nil {
			return false, err
		}
		updatedSchema.Indexes().AddIndex(sch.Indexes().AllIndexes()...)
		updatedSchema.SetComment(sch.GetComment())
		updatedSchema.SetTargetRowSize(sch.GetTargetRowSize())
		table, err = table.UpdateSchema(ctx, updatedSchema)
		if err != nil {
			return false, err
		}
		newRoot, err = newRoot.PutTable(ctx, name, table)
		return false, err
	})
	if err != nil {
		return nil, err
	}
	// Persist type and dependent schema metadata in the same working-root update.
	typ.EnumLabels = labels
	rootWithType, err := collection.UpdateRoot(ctx, newRoot.(*core.RootValue))
	if err != nil {
		return nil, err
	}
	if err = core.RecordEnumValueAdded(ctx, typ.ID, a.Value); err != nil {
		return nil, err
	}
	if err = session.SetWorkingRoot(ctx, ctx.GetCurrentDatabase(), rootWithType.(*core.RootValue)); err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// replaceEnumType refreshes enum dependencies without modifying cached types.
func replaceEnumType(typ, updated *types.DoltgresType, seen map[*types.DoltgresType]bool) (*types.DoltgresType, bool) {
	if typ == nil || seen[typ] {
		return typ, false
	}
	if typ.ID == updated.ID {
		return updated, true
	}
	seen[typ] = true
	defer delete(seen, typ)
	replacement := typ.Copy()
	changed := false
	var replaced bool
	replacement.Elem, replaced = replaceEnumType(typ.Elem, updated, seen)
	changed = changed || replaced
	replacement.BaseTypeType, replaced = replaceEnumType(typ.BaseTypeType, updated, seen)
	changed = changed || replaced
	replacement.CompositeAttrs = append([]types.CompositeAttribute(nil), typ.CompositeAttrs...)
	for i := range replacement.CompositeAttrs {
		replacement.CompositeAttrs[i].Type, replaced = replaceEnumType(typ.CompositeAttrs[i].Type, updated, seen)
		changed = changed || replaced
	}
	if !changed {
		return typ, false
	}
	return replacement, true
}
