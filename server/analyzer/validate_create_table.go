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

package analyzer

import (
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
)

// validateCreateTable validates that a table can be created as specified
func validateCreateTable(ctx *sql.Context, a *analyzer.Analyzer, n sql.Node, scope *plan.Scope, sel analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	ct, ok := n.(*plan.CreateTable)
	if !ok {
		return n, transform.SameTree, nil
	}

	err := validateIdentifiers(ct)
	if err != nil {
		return nil, transform.SameTree, err
	}

	sch := ct.PkSchema().Schema
	idxs := ct.Indexes()
	err = validateIndexes(ctx, sch, idxs)
	if err != nil {
		return nil, transform.SameTree, err
	}

	return n, transform.SameTree, nil
}

// validateIdentifiers validates the names of all schema elements for validity
// TODO: we use 64 character as the max length for an identifier, postgres uses 63
func validateIdentifiers(ct *plan.CreateTable) error {
	err := analyzer.ValidateIdentifier(ct.Name())
	if err != nil {
		return err
	}

	colNames := make(map[string]bool)
	for _, col := range ct.PkSchema().Schema {
		err = analyzer.ValidateIdentifier(col.Name)
		if err != nil {
			return err
		}
		lower := strings.ToLower(col.Name)
		if colNames[lower] {
			return sql.ErrDuplicateColumn.New(col.Name)
		}
		colNames[lower] = true
	}

	for _, chDef := range ct.Checks() {
		err = analyzer.ValidateIdentifier(chDef.Name)
		if err != nil {
			return err
		}
	}

	for _, idxDef := range ct.Indexes() {
		err = analyzer.ValidateIdentifier(idxDef.Name)
		if err != nil {
			return err
		}
	}

	for _, fkDef := range ct.ForeignKeys() {
		err = analyzer.ValidateIdentifier(fkDef.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

// validateIndexes validates that the index definitions being created are valid
func validateIndexes(ctx *sql.Context, sch sql.Schema, idxDefs sql.IndexDefs) error {
	colMap := schToColMap(sch)
	for _, idxDef := range idxDefs {
		if err := validateIndex(ctx, colMap, idxDef); err != nil {
			return err
		}
	}

	return nil
}

// validateIndex ensures that the Index Definition is valid for the table schema.
// This function will throw errors and warnings as needed.
// All columns in the index must be:
//   - in the schema
//   - not duplicated
//   - a compatible type for an index
//
// TODO: there are other constraints on indexes that we could enforce and are not yet (e.g. JSON as an index)
func validateIndex(ctx *sql.Context, colMap map[string]*sql.Column, idxDef *sql.IndexDef) error {
	seenCols := make(map[string]struct{})
	for _, idxCol := range idxDef.Columns {
		schCol, exists := colMap[strings.ToLower(idxCol.Name)]
		if !exists {
			return sql.ErrKeyColumnDoesNotExist.New(idxCol.Name)
		}
		if _, ok := seenCols[schCol.Name]; ok {
			return sql.ErrDuplicateColumn.New(schCol.Name)
		}
		seenCols[schCol.Name] = struct{}{}
		if idxDef.IsFullText() {
			continue
		}
	}

	if idxDef.IsSpatial() {
		return errors.Errorf("spatial indexes are not supported")
	}

	return nil
}

// resolveAlterColumn is a validation rule that validates the schema changes in an ALTER TABLE statement.
func resolveAlterColumn(ctx *sql.Context, a *analyzer.Analyzer, n sql.Node, scope *plan.Scope, sel analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	if !analyzer.FlagIsSet(qFlags, sql.QFlagAlterTable) {
		return n, transform.SameTree, nil
	}

	var sch sql.Schema
	var indexes []string
	var validator sql.SchemaValidator
	keyedColumns := make(map[string]bool)
	var err error
	transform.Inspect(n, func(n sql.Node) bool {
		if st, ok := n.(sql.SchemaTarget); ok {
			sch = st.TargetSchema()
		}
		switch n := n.(type) {
		case *plan.ModifyColumn:
			if rt, ok := n.Table.(*plan.ResolvedTable); ok {
				if sv, ok := rt.UnwrappedDatabase().(sql.SchemaValidator); ok {
					validator = sv
				}
			}
			keyedColumns, err = analyzer.GetTableIndexColumns(ctx, n.Table)
			return false
		case *plan.RenameColumn:
			if rt, ok := n.Table.(*plan.ResolvedTable); ok {
				if sv, ok := rt.UnwrappedDatabase().(sql.SchemaValidator); ok {
					validator = sv
				}
			}
			return false
		case *plan.AddColumn:
			if rt, ok := n.Table.(*plan.ResolvedTable); ok {
				if sv, ok := rt.UnwrappedDatabase().(sql.SchemaValidator); ok {
					validator = sv
				}
			}
			keyedColumns, err = analyzer.GetTableIndexColumns(ctx, n.Table)
			return false
		case *plan.DropColumn:
			if rt, ok := n.Table.(*plan.ResolvedTable); ok {
				if sv, ok := rt.UnwrappedDatabase().(sql.SchemaValidator); ok {
					validator = sv
				}
			}
			return false
		case *plan.AlterIndex:
			if rt, ok := n.Table.(*plan.ResolvedTable); ok {
				if sv, ok := rt.UnwrappedDatabase().(sql.SchemaValidator); ok {
					validator = sv
				}
			}
			indexes, err = analyzer.GetTableIndexNames(ctx, a, n.Table)
		default:
		}
		return true
	})

	if err != nil {
		return nil, transform.SameTree, err
	}

	// Skip this validation if we didn't find one or more of the above node types
	if len(sch) == 0 {
		return n, transform.SameTree, nil
	}

	sch = sch.Copy() // Make a copy of the original schema to deal with any references to the original table.
	initialSch := sch

	// Need a TransformUp here because multiple of these statement types can be nested under a Block node.
	// It doesn't look it, but this is actually an iterative loop over all the independent clauses in an ALTER statement
	n, same, err := transform.Node(n, func(n sql.Node) (sql.Node, transform.TreeIdentity, error) {
		switch nn := n.(type) {
		case *plan.ModifyColumn:
			n, err := nn.WithTargetSchema(sch.Copy())
			if err != nil {
				return nil, transform.SameTree, err
			}

			sch, err = analyzer.ValidateModifyColumn(ctx, initialSch, sch, n.(*plan.ModifyColumn), keyedColumns)
			if err != nil {
				return nil, transform.SameTree, err
			}
			return n, transform.NewTree, nil
		case *plan.RenameColumn:
			n, err := nn.WithTargetSchema(sch.Copy())
			if err != nil {
				return nil, transform.SameTree, err
			}
			sch, err = analyzer.ValidateRenameColumn(initialSch, sch, n.(*plan.RenameColumn))
			if err != nil {
				return nil, transform.SameTree, err
			}
			return n, transform.NewTree, nil
		case *plan.AddColumn:
			n, err := nn.WithTargetSchema(sch.Copy())
			if err != nil {
				return nil, transform.SameTree, err
			}

			sch, err = analyzer.ValidateAddColumn(sch, n.(*plan.AddColumn))
			if err != nil {
				return nil, transform.SameTree, err
			}

			return n, transform.NewTree, nil
		case *plan.DropColumn:
			n, err := nn.WithTargetSchema(sch.Copy())
			if err != nil {
				return nil, transform.SameTree, err
			}
			sch, err = analyzer.ValidateDropColumn(initialSch, sch, n.(*plan.DropColumn))
			if err != nil {
				return nil, transform.SameTree, err
			}
			delete(keyedColumns, nn.Column)

			return n, transform.NewTree, nil
		case *plan.AlterIndex:
			n, err := nn.WithTargetSchema(sch.Copy())
			if err != nil {
				return nil, transform.SameTree, err
			}
			indexes, err = validateAlterIndex(ctx, initialSch, sch, n.(*plan.AlterIndex), indexes)
			if err != nil {
				return nil, transform.SameTree, err
			}

			keyedColumns = analyzer.UpdateKeyedColumns(keyedColumns, nn)
			return n, transform.NewTree, nil
		case *plan.AlterPK:
			n, err := nn.WithTargetSchema(sch.Copy())
			if err != nil {
				return nil, transform.SameTree, err
			}
			sch, err = validatePrimaryKey(ctx, initialSch, sch, n.(*plan.AlterPK))
			if err != nil {
				return nil, transform.SameTree, err
			}
			return n, transform.NewTree, nil
		case *plan.AlterDefaultSet:
			n, err := nn.WithTargetSchema(sch.Copy())
			if err != nil {
				return nil, transform.SameTree, err
			}
			sch, err = analyzer.ValidateAlterDefault(initialSch, sch, n.(*plan.AlterDefaultSet))
			if err != nil {
				return nil, transform.SameTree, err
			}
			return n, transform.NewTree, nil
		case *plan.AlterDefaultDrop:
			n, err := nn.WithTargetSchema(sch.Copy())
			if err != nil {
				return nil, transform.SameTree, err
			}
			sch, err = analyzer.ValidateDropDefault(initialSch, sch, n.(*plan.AlterDefaultDrop))
			if err != nil {
				return nil, transform.SameTree, err
			}
			return n, transform.NewTree, nil
		}
		return n, transform.SameTree, nil
	})

	if err != nil {
		return nil, transform.SameTree, err
	}

	if validator != nil {
		if err := validator.ValidateSchema(sch); err != nil {
			return nil, transform.SameTree, err
		}
	}

	return n, same, nil
}

// Returns the underlying table name for the node given
func getTableName(node sql.Node) string {
	var tableName string
	transform.Inspect(node, func(node sql.Node) bool {
		switch node := node.(type) {
		case *plan.TableAlias:
			tableName = node.Name()
			return false
		case *plan.ResolvedTable:
			tableName = node.Name()
			return false
		case *plan.UnresolvedTable:
			tableName = node.Name()
			return false
		case *plan.IndexedTableAccess:
			tableName = node.Name()
			return false
		}
		return true
	})

	return tableName
}

// validatePrimaryKey validates a primary key add or drop operation.
func validatePrimaryKey(ctx *sql.Context, initialSch, sch sql.Schema, ai *plan.AlterPK) (sql.Schema, error) {
	tableName := getTableName(ai.Table)
	switch ai.Action {
	case plan.PrimaryKeyAction_Create:
		if analyzer.HasPrimaryKeys(sch) {
			return nil, sql.ErrMultiplePrimaryKeysDefined.New()
		}

		colMap := schToColMap(sch)
		idxDef := &sql.IndexDef{
			Name:       "PRIMARY",
			Columns:    ai.Columns,
			Constraint: sql.IndexConstraint_Primary,
		}

		err := validateIndex(ctx, colMap, idxDef)
		if err != nil {
			return nil, err
		}

		for _, idxCol := range ai.Columns {
			schCol := colMap[strings.ToLower(idxCol.Name)]
			if schCol.Virtual {
				return nil, sql.ErrVirtualColumnPrimaryKey.New()
			}
		}

		// Set the primary keys
		for _, col := range ai.Columns {
			sch[sch.IndexOf(col.Name, tableName)].PrimaryKey = true
		}

		return sch, nil
	case plan.PrimaryKeyAction_Drop:
		if !analyzer.HasPrimaryKeys(sch) {
			return nil, sql.ErrCantDropFieldOrKey.New("PRIMARY")
		}

		for _, col := range sch {
			if col.PrimaryKey {
				col.PrimaryKey = false
			}
		}

		return sch, nil
	default:
		return sch, nil
	}
}

// validateAlterIndex validates the specified column can have an index added, dropped, or renamed. Returns an updated
// list of index name given the add, drop, or rename operations.
func validateAlterIndex(ctx *sql.Context, initialSch, sch sql.Schema, ai *plan.AlterIndex, indexes []string) ([]string, error) {
	switch ai.Action {
	case plan.IndexAction_Create:
		err := analyzer.ValidateIdentifier(ai.IndexName)
		if err != nil {
			return nil, err
		}
		colMap := schToColMap(sch)

		// TODO: plan.AlterIndex should just have a sql.IndexDef
		indexDef := &sql.IndexDef{
			Name:       ai.IndexName,
			Columns:    ai.Columns,
			Constraint: ai.Constraint,
			Storage:    ai.Using,
			Comment:    ai.Comment,
		}

		err = validateIndex(ctx, colMap, indexDef)
		if err != nil {
			return nil, err
		}
		return append(indexes, ai.IndexName), nil
	case plan.IndexAction_Drop:
		savedIdx := -1
		for i, idx := range indexes {
			if strings.EqualFold(idx, ai.IndexName) {
				savedIdx = i
				break
			}
		}
		if savedIdx == -1 {
			return nil, sql.ErrCantDropFieldOrKey.New(ai.IndexName)
		}
		// Remove the index from the list
		return append(indexes[:savedIdx], indexes[savedIdx+1:]...), nil
	case plan.IndexAction_Rename:
		err := analyzer.ValidateIdentifier(ai.IndexName)
		if err != nil {
			return nil, err
		}
		savedIdx := -1
		for i, idx := range indexes {
			if strings.EqualFold(idx, ai.PreviousIndexName) {
				savedIdx = i
			}
		}
		if savedIdx == -1 {
			return nil, sql.ErrCantDropFieldOrKey.New(ai.IndexName)
		}
		// Simulate the rename by deleting the old name and adding the new one.
		return append(append(indexes[:savedIdx], indexes[savedIdx+1:]...), ai.IndexName), nil
	}

	return indexes, nil
}
