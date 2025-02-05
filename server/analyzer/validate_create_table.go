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
	"github.com/dolthub/go-mysql-server/sql/types"
)

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
	if len(ct.Name()) > sql.MaxIdentifierLength {
		return sql.ErrInvalidIdentifier.New(ct.Name())
	}

	colNames := make(map[string]bool)
	for _, col := range ct.PkSchema().Schema {
		if len(col.Name) > sql.MaxIdentifierLength {
			return sql.ErrInvalidIdentifier.New(col.Name)
		}
		lower := strings.ToLower(col.Name)
		if colNames[lower] {
			return sql.ErrDuplicateColumn.New(col.Name)
		}
		colNames[lower] = true
	}

	for _, chDef := range ct.Checks() {
		if len(chDef.Name) > sql.MaxIdentifierLength {
			return sql.ErrInvalidIdentifier.New(chDef.Name)
		}
	}

	for _, idxDef := range ct.Indexes() {
		if len(idxDef.Name) > sql.MaxIdentifierLength {
			return sql.ErrInvalidIdentifier.New(idxDef.Name)
		}
	}

	for _, fkDef := range ct.ForeignKeys() {
		if len(fkDef.Name) > sql.MaxIdentifierLength {
			return sql.ErrInvalidIdentifier.New(fkDef.Name)
		}
	}

	return nil
}

// validateIndexes validates that the index definitions being create are valid
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
		if types.IsJSON(schCol.Type) && !idxDef.IsVector() {
			return sql.ErrJSONIndex.New(schCol.Name)
		}

		if idxDef.IsFullText() {
			continue
		}
	}

	if idxDef.IsSpatial() {
		return errors.Errorf("spatial indexes are not supported")
	}

	return nil
}
