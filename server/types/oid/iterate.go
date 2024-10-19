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

package oid

import (
	"sort"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression/function"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/sequences"
	pgtypes "github.com/dolthub/doltgresql/server/types"
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
	// Function is the callback for functions.
	Function func(ctx *sql.Context, function ItemFunction) (cont bool, err error)
	// Index is the callback for indexes.
	Index func(ctx *sql.Context, schema ItemSchema, table ItemTable, index ItemIndex) (cont bool, err error)
	// Schema is the callback for schemas/namespaces.
	Schema func(ctx *sql.Context, schema ItemSchema) (cont bool, err error)
	// Sequence is the callback for sequences.
	Sequence func(ctx *sql.Context, schema ItemSchema, sequence ItemSequence) (cont bool, err error)
	// Table is the callback for tables.
	Table func(ctx *sql.Context, schema ItemSchema, table ItemTable) (cont bool, err error)
	// Types is the callback for types.
	Type func(ctx *sql.Context, typ ItemType) (cont bool, err error)
	// View is the callback for views.
	View func(ctx *sql.Context, schema ItemSchema, view ItemView) (cont bool, err error)
	// SearchSchemas represents the search path. If left empty, then all schemas are iterated over. If supplied, then
	// schemas are iterated by their given order.
	SearchSchemas []string
}

// ItemCheck contains the relevant information to pass to the Check callback.
type ItemCheck struct {
	Index int
	OID   uint32
	Item  sql.CheckDefinition
}

// ColumnWithIndex is a helper struct to pass the column and its index to the ColumnDefault callback.
type ColumnWithIndex struct {
	Column      *sql.Column
	ColumnIndex int
}

// ItemColumnDefault contains the relevant information to pass to the ColumnDefault callback.
type ItemColumnDefault struct {
	Index int
	OID   uint32
	Item  ColumnWithIndex
}

// ItemForeignKey contains the relevant information to pass to the ForeignKey callback.
type ItemForeignKey struct {
	Index int
	OID   uint32
	Item  sql.ForeignKeyConstraint
}

// ItemFunction contains the relevant information to pass to the Function callback.
type ItemFunction struct {
	Index int
	OID   uint32
	Item  sql.Function
}

// ItemIndex contains the relevant information to pass to the Index callback.
type ItemIndex struct {
	Index int
	OID   uint32
	Item  sql.Index
}

// ItemSchema contains the relevant information to pass to the Schema callback.
type ItemSchema struct {
	Index int
	OID   uint32
	Item  sql.DatabaseSchema
}

func (is ItemSchema) IsSystemSchema() bool {
	return is.Item.SchemaName() == "information_schema" || is.Item.SchemaName() == "pg_catalog"
}

// ItemSequence contains the relevant information to pass to the Sequence callback.
type ItemSequence struct {
	Index int
	OID   uint32
	Item  *sequences.Sequence
}

// ItemTable contains the relevant information to pass to the Table callback.
type ItemTable struct {
	Index int
	OID   uint32
	Item  sql.Table
}

// ItemType contains the relevant information to pass to the Type callback.
type ItemType struct {
	// TODO: add Index when we add custom types
	OID  uint32
	Item pgtypes.DoltgresType
}

// ItemView contains the relevant information to pass to the View callback.
type ItemView struct {
	Index int
	OID   uint32
	Item  sql.ViewDefinition
}

// IterateDatabase iterates over the provided database, calling each callback as the relevant items are iterated
// over. This is a central function that homogenizes all iteration, since OIDs depend on a deterministic iteration over
// items. This function should be expanded as we add more items to iterate over.
func IterateDatabase(ctx *sql.Context, database string, callbacks Callbacks) error {
	// Functions and types aren't contained within a schema for now, so we'll iterate over those separately
	if callbacks.Function != nil {
		if err := iterateFunctions(ctx, callbacks); err != nil {
			return err
		}
	}
	if callbacks.Type != nil {
		if err := iterateTypes(ctx, callbacks); err != nil {
			return err
		}
	}

	doltSession := dsess.DSessFromSess(ctx.Session)

	cat := sqle.NewDefault(doltSession.Provider()).Analyzer.Catalog
	currentDatabase, err := cat.Database(ctx, database)
	if err != nil {
		return err
	}

	// Then we'll iterate over everything that is contained within a schema
	if currentSchemaDatabase, ok := currentDatabase.(sql.SchemaDatabase); ok && callbacks.iteratesOverSchemas() {
		// Load and sort all of the schemas by name ascending
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
			collection, err := core.GetSequencesCollectionFromContext(ctx)
			if err != nil {
				return err
			}
			sequenceMap, _, _ = collection.GetAllSequences()
		}
		if err = iterateSchemas(ctx, callbacks, schemas, sequenceMap); err != nil {
			return err
		}
	}
	return nil
}

// IterateCurrentDatabase iterates over the current database, calling each callback as the relevant items are iterated
// over. This is a central function that homogenizes all iteration, since OIDs depend on a deterministic iteration over
// items. This function should be expanded as we add more items to iterate over.
func IterateCurrentDatabase(ctx *sql.Context, callbacks Callbacks) error {
	return IterateDatabase(ctx, ctx.GetCurrentDatabase(), callbacks)
}

// iterateFunctions is called by IterateCurrentDatabase to handle functions.
func iterateFunctions(ctx *sql.Context, callbacks Callbacks) error {
	for functionIndex, f := range function.BuiltIns {
		itemFunction := ItemFunction{
			Index: functionIndex,
			OID:   CreateOID(Section_Function, 0, functionIndex),
			Item:  f,
		}
		if cont, err := callbacks.Function(ctx, itemFunction); err != nil {
			return err
		} else if !cont {
			return nil
		}
	}
	return nil
}

// iterateTypes is called by IterateCurrentDatabase to handle types
func iterateTypes(ctx *sql.Context, callbacks Callbacks) error {
	// We only iterate over the types that are present in the pg_type table.
	// This means that we ignore the schema if one is given and it's not equal to "pg_catalog".
	// If no schemas were given, then we'll automatically look for the types in "pg_catalog".
	if len(callbacks.SearchSchemas) > 0 {
		containsPgCatalog := false
		for _, schema := range callbacks.SearchSchemas {
			if schema == "pg_catalog" {
				containsPgCatalog = true
				break
			}
		}
		if !containsPgCatalog {
			return nil
		}
	}
	// this gets all built-in types
	for _, t := range pgtypes.GetAllTypes() {
		if t.BaseID().HasUniqueOID() {
			cont, err := callbacks.Type(ctx, ItemType{
				OID:  t.OID(),
				Item: t,
			})
			if err != nil {
				return err
			}
			if !cont {
				return nil
			}
		}
	}
	// TODO: add domain and custom types when supported
	return nil
}

// iterateSchemas is called by IterateCurrentDatabase to handle schemas and elements contained within schemas.
func iterateSchemas(ctx *sql.Context, callbacks Callbacks, sortedSchemas []sql.DatabaseSchema, sequenceMap map[string][]*sequences.Sequence) error {
	// Iterate over the sorted schemas by the iteration order
	for _, schemaIndex := range callbacks.schemaIterationOrder(sortedSchemas) {
		schema := sortedSchemas[schemaIndex]
		itemSchema := ItemSchema{
			Index: schemaIndex,
			OID:   CreateOID(Section_Namespace, 0, schemaIndex),
			Item:  schema,
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
		// Iterate over sequences. The map will only be populated if the sequence callback exists.
		for sequenceIndex, sequence := range sequenceMap[schema.SchemaName()] {
			itemSequence := ItemSequence{
				Index: sequenceIndex,
				OID:   CreateOID(Section_Sequence, schemaIndex, sequenceIndex),
				Item:  sequence,
			}
			if cont, err := callbacks.Sequence(ctx, itemSchema, itemSequence); err != nil {
				return err
			} else if !cont {
				return nil
			}
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
		for viewIndex, view := range views {
			itemView := ItemView{
				Index: viewIndex,
				OID:   CreateOID(Section_View, itemSchema.Index, viewIndex),
				Item:  view,
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
	for tableIndex, tableName := range sortedTableNames {
		table, ok, err := itemSchema.Item.GetTableInsensitive(ctx, tableName)
		if err != nil {
			return err
		} else if !ok {
			return sql.ErrTableNotFound.New(tableName)
		}
		itemTable := ItemTable{
			Index: tableIndex,
			OID:   CreateOID(Section_Table, itemSchema.Index, tableIndex),
			Item:  table,
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
				Index: *checkCount,
				OID:   CreateOID(Section_Check, itemSchema.Index, *checkCount),
				Item:  check,
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
				Index: *columnDefaultCount,
				OID:   CreateOID(Section_ColumnDefault, itemSchema.Index, *columnDefaultCount),
				Item:  ColumnWithIndex{col, i},
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
				Index: *foreignKeyCount,
				OID:   CreateOID(Section_ForeignKey, itemSchema.Index, *foreignKeyCount),
				Item:  foreignKey,
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
				Index: *indexCount,
				OID:   CreateOID(Section_Index, itemSchema.Index, *indexCount),
				Item:  index,
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
func RunCallback(ctx *sql.Context, oid uint32, callbacks Callbacks) error {
	section, schemaIndex, dataIndex := ParseOID(oid)
	if ok := runCallbackValidation(ctx, oid, callbacks); !ok {
		return nil
	}
	// Functions and types aren't contained within a schema for now, so we'll iterate over those here
	if section == Section_Function {
		return runFunction(ctx, oid, callbacks)
	}
	if section == Section_BuiltIn {
		return runType(ctx, oid, callbacks)
	}
	// We know we have the relevant callback for the given section, so we'll grab the schema
	doltSession := dsess.DSessFromSess(ctx.Session)
	currentDatabase, err := sqle.NewDefault(doltSession.Provider()).Analyzer.Catalog.Database(ctx, ctx.GetCurrentDatabase())
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
		if section == Section_Namespace {
			return runNamespace(ctx, oid, callbacks, schemas)
		}
		// We're not looking for the schema, so we'll just select the target schema containing what we're looking for
		if schemaIndex >= len(schemas) {
			return nil
		}
		itemSchema := ItemSchema{
			Index: schemaIndex,
			OID:   oid,
			Item:  schemas[schemaIndex],
		}
		// Check if we're looking for a sequence
		if section == Section_Sequence {
			return runSequence(ctx, oid, callbacks, itemSchema)
		}
		// Check if we're looking for a view
		if section == Section_View {
			return runView(ctx, oid, callbacks, itemSchema)
		}
		// Before diving inside of tables, we'll check if we're looking for a table
		tableNames, err := itemSchema.Item.GetTableNames(ctx)
		if err != nil {
			return err
		}
		sort.Slice(tableNames, func(i, j int) bool {
			return tableNames[i] < tableNames[j]
		})
		if section == Section_Table {
			return runTable(ctx, oid, callbacks, itemSchema, tableNames)
		}
		// All other sections are items on tables, so we'll iterate over those now.
		// We hold the count here since it increments across tables.
		countedIndex := 0
		for _, tableName := range tableNames {
			table, ok, err := itemSchema.Item.GetTableInsensitive(ctx, tableName)
			if err != nil {
				return err
			} else if !ok {
				return sql.ErrTableNotFound.New(tableName)
			}
			itemTable := ItemTable{
				Index: dataIndex,
				OID:   oid,
				Item:  table,
			}
			switch section {
			case Section_Check:
				if ok, err := runCheck(ctx, oid, callbacks, itemSchema, itemTable, &countedIndex); err != nil {
					return err
				} else if ok {
					continue
				}
				return nil
			case Section_ColumnDefault:
				if ok, err := runColumnDefault(ctx, oid, callbacks, itemSchema, itemTable, &countedIndex); err != nil {
					return err
				} else if ok {
					continue
				}
				return nil
			case Section_ForeignKey:
				if ok, err := runForeignKey(ctx, oid, callbacks, itemSchema, itemTable, &countedIndex); err != nil {
					return err
				} else if ok {
					continue
				}
				return nil
			case Section_Index:
				if ok, err := runIndex(ctx, oid, callbacks, itemSchema, itemTable, &countedIndex); err != nil {
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
func runCheck(ctx *sql.Context, oid uint32, callbacks Callbacks, itemSchema ItemSchema, itemTable ItemTable, countedIndex *int) (cont bool, err error) {
	_, _, dataIndex := ParseOID(oid)
	if checkTable, ok := itemTable.Item.(sql.CheckTable); ok {
		checks, err := checkTable.GetChecks(ctx)
		if err != nil {
			return false, err
		}
		if dataIndex >= *countedIndex+len(checks) {
			*countedIndex += len(checks)
			return true, nil
		}
		sort.Slice(checks, func(i, j int) bool {
			return checks[i].Name < checks[j].Name
		})
		itemCheck := ItemCheck{
			Index: dataIndex,
			OID:   oid,
			Item:  checks[dataIndex-(*countedIndex)],
		}
		_, err = callbacks.Check(ctx, itemSchema, itemTable, itemCheck)
		return false, err
	}
	return true, nil
}

// runColumnDefault is called by RunCallback to handle Section_ColumnDefault.
func runColumnDefault(ctx *sql.Context, oid uint32, callbacks Callbacks, itemSchema ItemSchema, itemTable ItemTable, countedIndex *int) (cont bool, err error) {
	_, _, dataIndex := ParseOID(oid)
	columns := itemTable.Item.Schema()

	var colDefaults []ColumnWithIndex
	for i, col := range columns {
		if col.Default != nil {
			colDefaults = append(colDefaults, ColumnWithIndex{col, i})
		}
	}
	if dataIndex >= *countedIndex+len(colDefaults) {
		*countedIndex += len(colDefaults)
		return true, nil
	}

	itemColDefault := ItemColumnDefault{
		Index: dataIndex,
		OID:   oid,
		Item:  colDefaults[dataIndex-(*countedIndex)],
	}
	_, err = callbacks.ColumnDefault(ctx, itemSchema, itemTable, itemColDefault)
	if err != nil {
		return false, err
	}
	return true, nil
}

// runForeignKey is called by RunCallback to handle Section_ForeignKey.
func runForeignKey(ctx *sql.Context, oid uint32, callbacks Callbacks, itemSchema ItemSchema, itemTable ItemTable, countedIndex *int) (cont bool, err error) {
	_, _, dataIndex := ParseOID(oid)
	if fkTable, ok := itemTable.Item.(sql.ForeignKeyTable); ok {
		foreignKeys, err := fkTable.GetDeclaredForeignKeys(ctx)
		if err != nil {
			return false, err
		}
		if dataIndex >= *countedIndex+len(foreignKeys) {
			*countedIndex += len(foreignKeys)
			return true, nil
		}
		sort.Slice(foreignKeys, func(i, j int) bool {
			return foreignKeys[i].Name < foreignKeys[j].Name
		})
		itemForeignKey := ItemForeignKey{
			Index: dataIndex,
			OID:   oid,
			Item:  foreignKeys[dataIndex-(*countedIndex)],
		}
		_, err = callbacks.ForeignKey(ctx, itemSchema, itemTable, itemForeignKey)
		return false, err
	}
	return true, nil
}

// runFunction is called by RunCallback to handle Section_Function.
func runFunction(ctx *sql.Context, oid uint32, callbacks Callbacks) error {
	_, _, dataIndex := ParseOID(oid)
	if dataIndex >= len(function.BuiltIns) {
		return nil
	}
	itemFunction := ItemFunction{
		Index: dataIndex,
		OID:   oid,
		Item:  function.BuiltIns[dataIndex],
	}
	_, err := callbacks.Function(ctx, itemFunction)
	return err
}

// runIndex is called by RunCallback to handle Section_Index.
func runIndex(ctx *sql.Context, oid uint32, callbacks Callbacks, itemSchema ItemSchema, itemTable ItemTable, countedIndex *int) (cont bool, err error) {
	_, _, dataIndex := ParseOID(oid)
	if indexedTable, ok := itemTable.Item.(sql.IndexAddressable); ok {
		indexes, err := indexedTable.GetIndexes(ctx)
		if err != nil {
			return false, err
		}
		if dataIndex >= *countedIndex+len(indexes) {
			*countedIndex += len(indexes)
			return true, nil
		}
		sort.Slice(indexes, func(i, j int) bool {
			return indexes[i].ID() < indexes[j].ID()
		})
		itemIndex := ItemIndex{
			Index: dataIndex,
			OID:   oid,
			Item:  indexes[dataIndex-*countedIndex],
		}
		_, err = callbacks.Index(ctx, itemSchema, itemTable, itemIndex)
		return false, err
	}
	return true, nil
}

// runNamespace is called by RunCallback to handle Section_Namespace.
func runNamespace(ctx *sql.Context, oid uint32, callbacks Callbacks, sortedSchemas []sql.DatabaseSchema) error {
	_, _, dataIndex := ParseOID(oid)
	if dataIndex >= len(sortedSchemas) {
		return nil
	}
	itemSchema := ItemSchema{
		Index: dataIndex,
		OID:   oid,
		Item:  sortedSchemas[dataIndex],
	}
	_, err := callbacks.Schema(ctx, itemSchema)
	return err
}

// runSequence is called by RunCallback to handle Section_Sequence.
func runSequence(ctx *sql.Context, oid uint32, callbacks Callbacks, itemSchema ItemSchema) error {
	_, _, dataIndex := ParseOID(oid)
	collection, err := core.GetSequencesCollectionFromContext(ctx)
	if err != nil {
		return err
	}
	sequenceMap, _, _ := collection.GetAllSequences()
	sequencesInSchema, ok := sequenceMap[itemSchema.Item.SchemaName()]
	if !ok || dataIndex >= len(sequencesInSchema) {
		return nil
	}
	itemSequence := ItemSequence{
		Index: dataIndex,
		OID:   oid,
		Item:  sequencesInSchema[dataIndex],
	}
	_, err = callbacks.Sequence(ctx, itemSchema, itemSequence)
	return err
}

// runTable is called by RunCallback to handle Section_Table.
func runTable(ctx *sql.Context, oid uint32, callbacks Callbacks, itemSchema ItemSchema, sortedTableNames []string) error {
	_, _, dataIndex := ParseOID(oid)
	if dataIndex >= len(sortedTableNames) {
		return nil
	}
	table, ok, err := itemSchema.Item.GetTableInsensitive(ctx, sortedTableNames[dataIndex])
	if err != nil {
		return err
	} else if !ok {
		return sql.ErrTableNotFound.New(sortedTableNames[dataIndex])
	}
	itemTable := ItemTable{
		Index: dataIndex,
		OID:   oid,
		Item:  table,
	}
	_, err = callbacks.Table(ctx, itemSchema, itemTable)
	return err
}

// runType is called by RunCallback to handle types within Section_BuiltIn.
func runType(ctx *sql.Context, toid uint32, callbacks Callbacks) error {
	if t := pgtypes.GetTypeByOID(toid); t != nil {
		itemType := ItemType{
			OID:  toid,
			Item: t,
		}
		_, err := callbacks.Type(ctx, itemType)
		return err
	}
	return nil
}

// runView is called by RunCallback to handle Section_View.
func runView(ctx *sql.Context, oid uint32, callbacks Callbacks, itemSchema ItemSchema) error {
	_, _, dataIndex := ParseOID(oid)
	if viewDatabase, ok := itemSchema.Item.(sql.ViewDatabase); ok {
		views, err := viewDatabase.AllViews(ctx)
		if err != nil {
			return err
		}
		sort.Slice(views, func(i, j int) bool {
			return views[i].Name < views[j].Name
		})
		if dataIndex >= len(views) {
			return nil
		}
		itemView := ItemView{
			Index: dataIndex,
			OID:   oid,
			Item:  views[dataIndex],
		}
		_, err = callbacks.View(ctx, itemSchema, itemView)
		return err
	}
	return nil
}

// runCallbackValidation ensures that the callbacks match the given oid.
func runCallbackValidation(ctx *sql.Context, oid uint32, callbacks Callbacks) bool {
	section, _, _ := ParseOID(oid)
	// Check that we have the relevant callback, and return early if we do not
	switch section {
	case Section_BuiltIn:
		// For now, only the built-in types are checked in the built-in section
		if callbacks.Type == nil {
			return false
		}
	case Section_Check:
		if callbacks.Check == nil {
			return false
		}
	case Section_ColumnDefault:
		if callbacks.ColumnDefault == nil {
			return false
		}
	case Section_Database:
		// TODO: we inject information_schema, so we need to figure out how to model that here
		return false
	case Section_ForeignKey:
		if callbacks.ForeignKey == nil {
			return false
		}
	case Section_Function:
		if callbacks.Function == nil {
			return false
		}
	case Section_Index:
		if callbacks.Index == nil {
			return false
		}
	case Section_Namespace:
		if callbacks.Schema == nil {
			return false
		}
	case Section_Sequence:
		if callbacks.Sequence == nil {
			return false
		}
	case Section_Table:
		if callbacks.Table == nil {
			return false
		}
	case Section_View:
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
