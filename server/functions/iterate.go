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

package functions

import (
	"sort"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/doltgresql/core/typecollection"
	"github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/sequences"
)

// Callbacks are a set of callbacks that are used to simplify and coalesce all iteration involving database elements and
// their OIDs. All callbacks should be left nil except for the ones that are desired. Search paths are also supported
// through SearchSchemas.
type Callbacks struct {
	// Check is the callback for check constraints.
	Check func(ctx *sql.Context, schema ItemSchema, table ItemTable, check ItemCheck) (cont bool, err error)
	// ColumnDefault is the callback for column defaults.
	ColumnDefault func(ctx *sql.Context, schema ItemSchema, table ItemTable, check ItemColumnDefault) (cont bool, err error)
	// ForeignKey is the callback for foreign keys.
	ForeignKey func(ctx *sql.Context, schema ItemSchema, table ItemTable, foreignKey ItemForeignKey) (cont bool, err error)
	// Index is the callback for indexes.
	Index func(ctx *sql.Context, schema ItemSchema, table ItemTable, index ItemIndex) (cont bool, err error)
	// Schema is the callback for schemas/namespaces.
	Schema func(ctx *sql.Context, schema ItemSchema) (cont bool, err error)
	// Sequence is the callback for sequences.
	Sequence func(ctx *sql.Context, schema ItemSchema, sequence ItemSequence) (cont bool, err error)
	// Table is the callback for tables.
	Table func(ctx *sql.Context, schema ItemSchema, table ItemTable) (cont bool, err error)
	// Type is the callback for types.
	Type func(ctx *sql.Context, schema ItemSchema, typ ItemType) (cont bool, err error)
	// View is the callback for views.
	View func(ctx *sql.Context, schema ItemSchema, view ItemView) (cont bool, err error)
	// SearchSchemas represents the search path. If left empty, then all schemas are iterated over. If supplied, then
	// schemas are iterated by their given order.
	SearchSchemas []string
}

// ItemCheck contains the relevant information to pass to the Check callback.
type ItemCheck struct {
	OID  id.Check
	Item sql.CheckDefinition
}

// ColumnWithIndex is a helper struct to pass the column and its index to the ColumnDefault callback.
type ColumnWithIndex struct {
	Column      *sql.Column
	ColumnIndex int
}

// ItemColumnDefault contains the relevant information to pass to the ColumnDefault callback.
type ItemColumnDefault struct {
	OID  id.ColumnDefault
	Item ColumnWithIndex
}

// ItemForeignKey contains the relevant information to pass to the ForeignKey callback.
type ItemForeignKey struct {
	OID  id.ForeignKey
	Item sql.ForeignKeyConstraint
}

// ItemIndex contains the relevant information to pass to the Index callback.
type ItemIndex struct {
	OID  id.Index
	Item sql.Index
}

// ItemSchema contains the relevant information to pass to the Schema callback.
type ItemSchema struct {
	OID  id.Namespace
	Item sql.DatabaseSchema
}

func (is ItemSchema) IsSystemSchema() bool {
	return is.Item.SchemaName() == "information_schema" || is.Item.SchemaName() == "pg_catalog"
}

// ItemSequence contains the relevant information to pass to the Sequence callback.
type ItemSequence struct {
	OID  id.Sequence
	Item *sequences.Sequence
}

// ItemTable contains the relevant information to pass to the Table callback.
type ItemTable struct {
	OID  id.Table
	Item sql.Table
}

// ItemType contains the relevant information to pass to the Type callback.
type ItemType struct {
	Oid  id.Type
	Item *types.DoltgresType
}

// ItemView contains the relevant information to pass to the View callback.
type ItemView struct {
	OID  id.View
	Item sql.ViewDefinition
}

// IterateDatabase iterates over the provided database, calling each callback as the relevant items are iterated
// over. This is a central function that homogenizes all iteration, since OIDs depend on a deterministic iteration over
// items. This function should be expanded as we add more items to iterate over.
func IterateDatabase(ctx *sql.Context, database string, callbacks Callbacks) error {
	sess := ctx.Session.(*dsess.DoltSession)
	currentDatabase, err := sess.Provider().Database(ctx, database)
	if err != nil {
		return err
	}

	// Then we'll iterate over everything that is contained within a schema
	if currentSchemaDatabase, ok := currentDatabase.(sql.SchemaDatabase); ok && callbacks.iteratesOverSchemas() {
		// Load and sort all schemas by name ascending
		schemas, err := currentSchemaDatabase.AllSchemas(ctx)
		if err != nil {
			return err
		}

		sort.Slice(schemas, func(i, j int) bool {
			return schemas[i].SchemaName() < schemas[j].SchemaName()
		})
		// We'll load the sequences early (if the callback exists) since they're stored on the root.
		// Within the schema iteration, we can simply
		var sequenceMap map[string][]*sequences.Sequence
		if callbacks.Sequence != nil {
			collection, err := core.GetSequencesCollectionFromContext(ctx, database)
			if err != nil {
				return err
			}
			sequenceMap, _, _, err = collection.GetAllSequences(ctx)
			if err != nil {
				return err
			}
		}

		var typeMap map[string][]*types.DoltgresType
		if callbacks.Type != nil {
			coll, err := core.GetTypesCollectionFromContext(ctx)
			if err != nil {
				return err
			}

			typeMap = typesBySchema(ctx, coll)
		}

		if err = iterateSchemas(ctx, callbacks, schemas, sequenceMap, typeMap); err != nil {
			return err
		}
	}
	return nil
}

// typesBySchema returns a map of schema name to types within that schema.
func typesBySchema(ctx *sql.Context, coll *typecollection.TypeCollection) map[string][]*types.DoltgresType {
	m := make(map[string][]*types.DoltgresType)
	_ = coll.IterateTypes(ctx, func(typ *types.DoltgresType) (stop bool, err error) {
		m[typ.Schema()] = append(m[typ.Schema()], typ)
		return false, nil
	})
	return m
}

// IterateCurrentDatabase iterates over the current database, calling each callback as the relevant items are iterated
// over. This is a central function that homogenizes all iteration, since OIDs depend on a deterministic iteration over
// items. This function should be expanded as we add more items to iterate over.
func IterateCurrentDatabase(ctx *sql.Context, callbacks Callbacks) error {
	return IterateDatabase(ctx, ctx.GetCurrentDatabase(), callbacks)
}

// iterateSchemas is called by IterateCurrentDatabase to handle schemas and elements contained within schemas.
func iterateSchemas(
	ctx *sql.Context,
	callbacks Callbacks,
	sortedSchemas []sql.DatabaseSchema,
	sequenceMap map[string][]*sequences.Sequence,
	typeMap map[string][]*types.DoltgresType,
) error {
	// Iterate over the sorted schemas by the iteration order
	for _, schemaIndex := range callbacks.schemaIterationOrder(sortedSchemas) {
		schema := sortedSchemas[schemaIndex]
		itemSchema := ItemSchema{
			OID:  id.NewNamespace(schema.SchemaName()),
			Item: schema,
		}
		// Check for a schema callback
		if callbacks.Schema != nil {
			if cont, err := callbacks.Schema(ctx, itemSchema); err != nil {
				return err
			} else if !cont {
				return nil
			}
		}
		// Check for a view callback
		if callbacks.View != nil {
			if err := iterateViews(ctx, callbacks, itemSchema); err != nil {
				return err
			}
		}

		err := iterateSequences(ctx, callbacks, sequenceMap, schema, itemSchema)
		if err != nil {
			return err
		}

		// Check if we need to iterate over tables
		if callbacks.iteratesOverTables() {
			tableNames, err := schema.GetTableNames(ctx)
			if err != nil {
				return err
			}
			sort.Slice(tableNames, func(i, j int) bool {
				return tableNames[i] < tableNames[j]
			})
			if err = iterateTables(ctx, callbacks, itemSchema, tableNames); err != nil {
				return err
			}
		}

		if callbacks.iteratesOverTypes() {
			if err := iterateTypes(ctx, callbacks, itemSchema, typeMap); err != nil {
				return err
			}
		}
	}
	return nil
}

// iterateSequences is called by iterateSchemas to handle sequence callbacks
func iterateSequences(ctx *sql.Context, callbacks Callbacks, sequenceMap map[string][]*sequences.Sequence, schema sql.DatabaseSchema, itemSchema ItemSchema) error {
	for _, sequence := range sequenceMap[schema.SchemaName()] {
		itemSequence := ItemSequence{
			OID:  sequence.Id,
			Item: sequence,
		}
		if cont, err := callbacks.Sequence(ctx, itemSchema, itemSequence); err != nil {
			return err
		} else if !cont {
			return nil
		}
	}
	return nil
}

// iterateTypes is called by iterateSchemas to handle type callbacks
func iterateTypes(ctx *sql.Context, callbacks Callbacks, itemSchema ItemSchema, typeMap map[string][]*types.DoltgresType) error {
	for _, typ := range typeMap[itemSchema.Item.SchemaName()] {
		itemSchemaType := ItemType{
			Oid:  typ.ID,
			Item: typ,
		}
		cont, err := callbacks.Type(ctx, itemSchema, itemSchemaType)
		if err != nil {
			return err
		} else if !cont {
			return nil
		}
	}
	return nil
}

// iterateViews is called by iterateSchemas to handle views.
func iterateViews(ctx *sql.Context, callbacks Callbacks, itemSchema ItemSchema) error {
	if viewDatabase, ok := itemSchema.Item.(sql.ViewDatabase); ok {
		views, err := viewDatabase.AllViews(ctx)
		if err != nil {
			return err
		}
		sort.Slice(views, func(i, j int) bool {
			return views[i].Name < views[j].Name
		})
		for _, view := range views {
			itemView := ItemView{
				OID:  id.NewView(itemSchema.Item.SchemaName(), view.Name),
				Item: view,
			}
			if cont, err := callbacks.View(ctx, itemSchema, itemView); err != nil {
				return err
			} else if !cont {
				return nil
			}
		}
	}
	return nil
}

// iterateTables is called by iterateSchemas to handle tables and elements contained within tables.
func iterateTables(ctx *sql.Context, callbacks Callbacks, itemSchema ItemSchema, sortedTableNames []string) error {
	// Start all of the counts at -1, since we always increment before using the count.
	checkCount := -1
	columnDefaultCount := -1
	foreignKeyCount := -1
	indexCount := -1

	// Iterate over the sorted table names
	for _, tableName := range sortedTableNames {
		table, ok, err := itemSchema.Item.GetTableInsensitive(ctx, tableName)
		if err != nil && !errors.Is(err, doltdb.ErrTableNotFound) {
			return err
		} else if !ok {
			// We receive these names from the database, so these must be the names of root objects
			continue
		}
		itemTable := ItemTable{
			OID:  id.NewTable(itemSchema.Item.SchemaName(), table.Name()),
			Item: table,
		}

		// Check for a check constraint callback
		if callbacks.Check != nil {
			if err = iterateChecks(ctx, callbacks, itemSchema, itemTable, &checkCount); err != nil {
				return err
			}
		}
		// Check for a column default callback
		if callbacks.ColumnDefault != nil {
			// System tables don't have column defaults
			if itemSchema.IsSystemSchema() {
				continue
			}
			if err = iterateColumnDefaults(ctx, callbacks, itemSchema, itemTable, &columnDefaultCount); err != nil {
				return err
			}
		}
		// Check for a foreign key callback
		if callbacks.ForeignKey != nil {
			// System tables don't have foreign keys
			if itemSchema.IsSystemSchema() {
				continue
			}
			if err = iterateForeignKeys(ctx, callbacks, itemSchema, itemTable, &foreignKeyCount); err != nil {
				return err
			}
		}
		// Check for an index callback
		if callbacks.Index != nil {
			if err = iterateIndexes(ctx, callbacks, itemSchema, itemTable, &indexCount); err != nil {
				return err
			}
		}
		// Check for a table callback
		if callbacks.Table != nil {
			if cont, err := callbacks.Table(ctx, itemSchema, itemTable); err != nil {
				return err
			} else if !cont {
				return nil
			}
		}
	}
	return nil
}

// iterateChecks is called by iterateTables to handle check constraints.
func iterateChecks(ctx *sql.Context, callbacks Callbacks, itemSchema ItemSchema, itemTable ItemTable, checkCount *int) error {
	if checkTable, ok := itemTable.Item.(sql.CheckTable); ok {
		checks, err := checkTable.GetChecks(ctx)
		if err != nil {
			return err
		}
		sort.Slice(checks, func(i, j int) bool {
			return checks[i].Name < checks[j].Name
		})
		for _, check := range checks {
			*checkCount++
			itemCheck := ItemCheck{
				OID:  id.NewCheck(itemSchema.Item.SchemaName(), itemTable.Item.Name(), check.Name),
				Item: check,
			}
			if cont, err := callbacks.Check(ctx, itemSchema, itemTable, itemCheck); err != nil {
				return err
			} else if !cont {
				return nil
			}
		}
	}
	return nil
}

// iterateColumnDefaults is called by iterateTables to handle column defaults.
func iterateColumnDefaults(ctx *sql.Context, callbacks Callbacks, itemSchema ItemSchema, itemTable ItemTable, columnDefaultCount *int) error {
	columns := itemTable.Item.Schema()
	for i, col := range columns {
		if col.Default != nil {
			*columnDefaultCount++
			itemColDefault := ItemColumnDefault{
				OID:  id.NewColumnDefault(itemSchema.Item.SchemaName(), itemTable.Item.Name(), col.Name),
				Item: ColumnWithIndex{col, i},
			}
			if cont, err := callbacks.ColumnDefault(ctx, itemSchema, itemTable, itemColDefault); err != nil {
				return err
			} else if !cont {
				return nil
			}
		}
	}
	return nil
}

// iterateForeignKeys is called by iterateTables to handle foreign keys.
func iterateForeignKeys(ctx *sql.Context, callbacks Callbacks, itemSchema ItemSchema, itemTable ItemTable, foreignKeyCount *int) error {
	if fkTable, ok := itemTable.Item.(sql.ForeignKeyTable); ok {
		foreignKeys, err := fkTable.GetDeclaredForeignKeys(ctx)
		if err != nil {
			return err
		}
		sort.Slice(foreignKeys, func(i, j int) bool {
			return foreignKeys[i].Name < foreignKeys[j].Name
		})
		for _, foreignKey := range foreignKeys {
			*foreignKeyCount++
			itemForeignKey := ItemForeignKey{
				OID:  id.NewForeignKey(itemSchema.Item.SchemaName(), itemTable.Item.Name(), foreignKey.Name),
				Item: foreignKey,
			}
			if cont, err := callbacks.ForeignKey(ctx, itemSchema, itemTable, itemForeignKey); err != nil {
				return err
			} else if !cont {
				return nil
			}
		}
	}
	return nil
}

// iterateIndexes is called by iterateTables to handle indexes.
func iterateIndexes(ctx *sql.Context, callbacks Callbacks, itemSchema ItemSchema, itemTable ItemTable, indexCount *int) error {
	if indexedTable, ok := itemTable.Item.(sql.IndexAddressable); ok {
		indexes, err := indexedTable.GetIndexes(ctx)
		if err != nil {
			return err
		}
		sort.Slice(indexes, func(i, j int) bool {
			return indexes[i].ID() < indexes[j].ID()
		})
		for _, index := range indexes {
			*indexCount++
			itemIndex := ItemIndex{
				OID:  id.NewIndex(itemSchema.Item.SchemaName(), itemTable.Item.Name(), index.ID()),
				Item: index,
			}
			if cont, err := callbacks.Index(ctx, itemSchema, itemTable, itemIndex); err != nil {
				return err
			} else if !cont {
				return nil
			}
		}
	}
	return nil
}

// RunCallback iterates over schemas, etc. to find the item that the given oid points to. Once the item has been found,
// the relevant callback is called with the item. This means that, at most, only one callback will be called. If the
// item cannot be found, then no callbacks are called.
func RunCallback(ctx *sql.Context, internalID id.Id, callbacks Callbacks) error {
	if ok := runCallbackValidation(ctx, internalID, callbacks); !ok {
		return nil
	}
	// We know we have the relevant callback for the given section, so we'll grab the schema
	doltSession := dsess.DSessFromSess(ctx.Session)
	currentDatabase, err := doltSession.Provider().Database(ctx, ctx.GetCurrentDatabase())
	if err != nil {
		return err
	}

	if currentSchemaDatabase, ok := currentDatabase.(sql.SchemaDatabase); ok {
		schemas, err := currentSchemaDatabase.AllSchemas(ctx)
		if err != nil {
			return err
		}
		sort.Slice(schemas, func(i, j int) bool {
			return schemas[i].SchemaName() < schemas[j].SchemaName()
		})
		// First check if we're only looking for the schema
		if internalID.Section() == id.Section_Namespace {
			return runNamespace(ctx, internalID, callbacks, schemas)
		}
		// We're not looking for the schema, so we'll just select the target schema containing what we're looking for
		var itemSchema ItemSchema
		for _, schema := range schemas {
			if schema.SchemaName() == internalID.Segment(0) {
				itemSchema = ItemSchema{
					OID:  id.NewNamespace(schema.SchemaName()),
					Item: schema,
				}
			}
		}
		if !itemSchema.OID.IsValid() {
			return nil
		}
		// Check if we're looking for a sequence
		if internalID.Section() == id.Section_Sequence {
			return runSequence(ctx, internalID, callbacks, itemSchema)
		}
		// Check if we're looking for a view
		if internalID.Section() == id.Section_View {
			return runView(ctx, internalID, callbacks, itemSchema)
		}
		// Before diving inside of tables, we'll check if we're looking for a table
		tableNames, err := itemSchema.Item.GetTableNames(ctx)
		if err != nil {
			return err
		}
		sort.Slice(tableNames, func(i, j int) bool {
			return tableNames[i] < tableNames[j]
		})
		if internalID.Section() == id.Section_Table {
			return runTable(ctx, internalID, callbacks, itemSchema, tableNames)
		}
		// All other sections are items on tables, so we'll iterate over those now.
		// We hold the count here since it increments across tables.
		countedIndex := 0
		for _, tableName := range tableNames {
			table, ok, err := itemSchema.Item.GetTableInsensitive(ctx, tableName)
			if err != nil && !errors.Is(err, doltdb.ErrTableNotFound) {
				return err
			} else if !ok {
				// We receive these names from the schema, so these must be the names of root objects
				continue
			}
			itemTable := ItemTable{
				OID:  id.NewTable(itemSchema.Item.SchemaName(), table.Name()),
				Item: table,
			}
			switch internalID.Section() {
			case id.Section_Check:
				if ok, err := runCheck(ctx, internalID, callbacks, itemSchema, itemTable, &countedIndex); err != nil {
					return err
				} else if ok {
					continue
				}
				return nil
			case id.Section_ColumnDefault:
				if ok, err := runColumnDefault(ctx, internalID, callbacks, itemSchema, itemTable, &countedIndex); err != nil {
					return err
				} else if ok {
					continue
				}
				return nil
			case id.Section_ForeignKey:
				if ok, err := runForeignKey(ctx, internalID, callbacks, itemSchema, itemTable, &countedIndex); err != nil {
					return err
				} else if ok {
					continue
				}
				return nil
			case id.Section_Index:
				if ok, err := runIndex(ctx, internalID, callbacks, itemSchema, itemTable); err != nil {
					return err
				} else if ok {
					continue
				}
				return nil
			default: // This is unnecessary, but the linter complains without it
				return nil
			}
		}
	}
	return nil
}

// runCheck is called by RunCallback to handle Section_Check.
func runCheck(ctx *sql.Context, internalID id.Id, callbacks Callbacks, itemSchema ItemSchema, itemTable ItemTable, countedIndex *int) (cont bool, err error) {
	if itemSchema.Item.SchemaName() != internalID.Segment(0) && itemTable.Item.Name() != internalID.Segment(1) {
		return true, nil
	}
	if checkTable, ok := itemTable.Item.(sql.CheckTable); ok {
		checks, err := checkTable.GetChecks(ctx)
		if err != nil {
			return false, err
		}
		for _, check := range checks {
			if check.Name == internalID.Segment(2) {
				itemCheck := ItemCheck{
					OID:  id.Check(internalID),
					Item: check,
				}
				_, err = callbacks.Check(ctx, itemSchema, itemTable, itemCheck)
				return false, err
			}
		}
	}
	return true, nil
}

// runColumnDefault is called by RunCallback to handle Section_Column.
func runColumnDefault(ctx *sql.Context, internalID id.Id, callbacks Callbacks, itemSchema ItemSchema, itemTable ItemTable, countedIndex *int) (cont bool, err error) {
	if itemSchema.Item.SchemaName() != internalID.Segment(0) && itemTable.Item.Name() != internalID.Segment(1) {
		return true, nil
	}
	columns := itemTable.Item.Schema()
	var colDefaults []ColumnWithIndex
	for i, col := range columns {
		if col.Default != nil {
			colDefaults = append(colDefaults, ColumnWithIndex{col, i})
		}
	}
	for _, col := range colDefaults {
		if col.Column.Name == internalID.Segment(2) {
			itemColDefault := ItemColumnDefault{
				OID:  id.ColumnDefault(internalID),
				Item: col,
			}
			_, err = callbacks.ColumnDefault(ctx, itemSchema, itemTable, itemColDefault)
			if err != nil {
				return false, err
			}
		}
	}
	return true, nil
}

// runForeignKey is called by RunCallback to handle Section_ForeignKey.
func runForeignKey(ctx *sql.Context, internalID id.Id, callbacks Callbacks, itemSchema ItemSchema, itemTable ItemTable, countedIndex *int) (cont bool, err error) {
	if fkTable, ok := itemTable.Item.(sql.ForeignKeyTable); ok && itemSchema.Item.SchemaName() == internalID.Segment(0) && itemTable.Item.Name() == internalID.Segment(1) {
		foreignKeys, err := fkTable.GetDeclaredForeignKeys(ctx)
		if err != nil {
			return false, err
		}
		for _, foreignKey := range foreignKeys {
			if foreignKey.Name == internalID.Segment(2) {
				itemForeignKey := ItemForeignKey{
					OID:  id.ForeignKey(internalID),
					Item: foreignKey,
				}
				_, err = callbacks.ForeignKey(ctx, itemSchema, itemTable, itemForeignKey)
				return false, err
			}
		}
	}
	return true, nil
}

// runIndex is called by RunCallback to handle Section_Index.
func runIndex(ctx *sql.Context, internalID id.Id, callbacks Callbacks, itemSchema ItemSchema, itemTable ItemTable) (cont bool, err error) {
	if indexedTable, ok := itemTable.Item.(sql.IndexAddressable); ok && itemSchema.Item.SchemaName() == internalID.Segment(0) && itemTable.Item.Name() == internalID.Segment(1) {
		indexes, err := indexedTable.GetIndexes(ctx)
		if err != nil {
			return false, err
		}
		for _, index := range indexes {
			if index.ID() == internalID.Segment(2) && itemTable.Item.Name() == index.Table() {
				_, err = callbacks.Index(ctx, itemSchema, itemTable, ItemIndex{
					OID:  id.Index(internalID),
					Item: index,
				})
				return false, err
			}
		}
		return true, nil
	}
	return true, nil
}

// runNamespace is called by RunCallback to handle Section_Namespace.
func runNamespace(ctx *sql.Context, internalID id.Id, callbacks Callbacks, sortedSchemas []sql.DatabaseSchema) error {
	for _, schema := range sortedSchemas {
		if schema.SchemaName() == internalID.Segment(0) {
			itemSchema := ItemSchema{
				OID:  id.Namespace(internalID),
				Item: schema,
			}
			_, err := callbacks.Schema(ctx, itemSchema)
			return err
		}
	}
	return nil
}

// runSequence is called by RunCallback to handle Section_Sequence.
func runSequence(ctx *sql.Context, internalID id.Id, callbacks Callbacks, itemSchema ItemSchema) error {
	collection, err := core.GetSequencesCollectionFromContext(ctx, itemSchema.Item.Name())
	if err != nil {
		return err
	}
	sequenceMap, _, _, err := collection.GetAllSequences(ctx)
	if err != nil {
		return err
	}
	sequencesInSchema, ok := sequenceMap[itemSchema.Item.SchemaName()]
	if !ok {
		return nil
	}
	for _, seq := range sequencesInSchema {
		if id.Id(seq.Id) == internalID {
			_, err = callbacks.Sequence(ctx, itemSchema, ItemSequence{
				OID:  id.Sequence(internalID),
				Item: seq,
			})
			return err
		}
	}
	return nil
}

// runTable is called by RunCallback to handle Section_Table.
func runTable(ctx *sql.Context, internalID id.Id, callbacks Callbacks, itemSchema ItemSchema, sortedTableNames []string) error {
	table, ok, err := itemSchema.Item.GetTableInsensitive(ctx, internalID.Segment(1))
	if err != nil {
		return err
	} else if !ok {
		return sql.ErrTableNotFound.New(internalID.Segment(1))
	}
	itemTable := ItemTable{
		OID:  id.Table(internalID),
		Item: table,
	}
	_, err = callbacks.Table(ctx, itemSchema, itemTable)
	return err
}

// runView is called by RunCallback to handle Section_View.
func runView(ctx *sql.Context, internalID id.Id, callbacks Callbacks, itemSchema ItemSchema) error {
	if viewDatabase, ok := itemSchema.Item.(sql.ViewDatabase); ok && itemSchema.Item.SchemaName() == internalID.Segment(0) {
		views, err := viewDatabase.AllViews(ctx)
		if err != nil {
			return err
		}
		for _, view := range views {
			if view.Name == internalID.Segment(1) {
				_, err = callbacks.View(ctx, itemSchema, ItemView{
					OID:  id.View(internalID),
					Item: view,
				})
				return err
			}
		}
	}
	return nil
}

// runCallbackValidation ensures that the callbacks match the given oid.
func runCallbackValidation(ctx *sql.Context, internalID id.Id, callbacks Callbacks) bool {
	// Check that we have the relevant callback, and return early if we do not
	switch internalID.Section() {
	case id.Section_Check:
		if callbacks.Check == nil {
			return false
		}
	case id.Section_ColumnDefault:
		if callbacks.ColumnDefault == nil {
			return false
		}
	case id.Section_Database:
		// TODO: we inject information_schema, so we need to figure out how to model that here
		return false
	case id.Section_ForeignKey:
		if callbacks.ForeignKey == nil {
			return false
		}
	case id.Section_Index:
		if callbacks.Index == nil {
			return false
		}
	case id.Section_Namespace:
		if callbacks.Schema == nil {
			return false
		}
	case id.Section_Sequence:
		if callbacks.Sequence == nil {
			return false
		}
	case id.Section_Table:
		if callbacks.Table == nil {
			return false
		}
	case id.Section_View:
		if callbacks.View == nil {
			return false
		}
	default:
		return false
	}
	return true
}

// iteratesOverSchemas returns whether we need to iterate over schemas based on the given callbacks.
func (iter Callbacks) iteratesOverSchemas() bool {
	return iter.Check != nil ||
		iter.ColumnDefault != nil ||
		iter.ForeignKey != nil ||
		iter.Index != nil ||
		iter.Schema != nil ||
		iter.Sequence != nil ||
		iter.Table != nil ||
		iter.Type != nil ||
		iter.View != nil
}

// iteratesOverTables returns whether we need to iterate over tables based on the given callbacks.
func (iter Callbacks) iteratesOverTables() bool {
	return iter.Check != nil ||
		iter.ColumnDefault != nil ||
		iter.ForeignKey != nil ||
		iter.Index != nil ||
		iter.Table != nil
}

// iteratesOverTypes returns whether we need to iterate over types based on the given callbacks.
func (iter Callbacks) iteratesOverTypes() bool {
	return iter.Type != nil
}

// schemaIterationOrder returns the order that the given schemas should be iterated over.
func (iter Callbacks) schemaIterationOrder(sortedSchemas []sql.DatabaseSchema) []int {
	// If no search schemas are set, then we'll iterate over all of the schemas in sorted order
	if len(iter.SearchSchemas) == 0 {
		order := make([]int, len(sortedSchemas))
		for i := range sortedSchemas {
			order[i] = i
		}
		return order
	}
	sortedSchemasMap := make(map[string]int)
	for i, schema := range sortedSchemas {
		sortedSchemasMap[schema.SchemaName()] = i
	}
	// We only add the schemas that we can find
	order := make([]int, 0, len(iter.SearchSchemas))
	for _, searchSchema := range iter.SearchSchemas {
		if schemaPosition, ok := sortedSchemasMap[searchSchema]; ok {
			order = append(order, schemaPosition)
		}
	}
	return order
}
