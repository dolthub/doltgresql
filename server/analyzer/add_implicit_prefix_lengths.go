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

package analyzer

import (
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
	"github.com/lib/pq/oid"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// defaultIndexPrefixLength is the index prefix length that this analyzer rule applies automatically to TEXT columns
// in secondary indexes. 768 is the limit for the prefix length in MySQL and is also enforced in Dolt/GMS, so this
// is currently the largest size we can support.
const defaultIndexPrefixLength = 768

// AddImplicitPrefixLengths searches the |node| tree for any nodes creating an index, and plugs in a default index
// prefix length for any TEXT columns in those new indexes. This rule is intended to be used for Postgres compatibility,
// since Postgres does not require specifying prefix lengths for TEXT columns.
func AddImplicitPrefixLengths(_ *sql.Context, _ *analyzer.Analyzer, node sql.Node, _ *plan.Scope, _ analyzer.RuleSelector, _ *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	var targetSchema sql.Schema
	transform.Inspect(node, func(node sql.Node) bool {
		if st, ok := node.(sql.SchemaTarget); ok {
			targetSchema = st.TargetSchema().Copy()
			return false
		}
		return true
	})

	// Recurse through the node tree to fill in prefix lengths. Note that some statements come in as Block nodes
	// that contain multiple nodes, so we need to recurse through and handle all of them.
	return transform.Node(node, func(node sql.Node) (sql.Node, transform.TreeIdentity, error) {
		switch node := node.(type) {
		case *plan.AddColumn:
			// For any AddColumn nodes, we need to update the target schema with the column being added, otherwise
			// we won't be able to find those columns if they are also being added to a secondary index.
			var err error
			targetSchema, err = analyzer.ValidateAddColumn(targetSchema, node)
			if err != nil {
				return nil, transform.SameTree, err
			}

		case *plan.CreateTable:
			newIndexes := make([]*sql.IndexDef, len(node.Indexes()))
			for i := range node.Indexes() {
				copy := *node.Indexes()[i]
				newIndexes[i] = &copy
			}
			indexModified := false
			for _, index := range newIndexes {
				targetSchema := node.TargetSchema()
				colMap := schToColMap(targetSchema)
				for i := range index.Columns {
					col, ok := colMap[strings.ToLower(index.Columns[i].Name)]
					if !ok {
						return nil, false, fmt.Errorf("indexed column %s not found in schema", index.Columns[i].Name)
					}
					if dt, ok := col.Type.(pgtypes.DoltgresType); ok && dt.OID == uint32(oid.T_text) && index.Columns[i].Length == 0 {
						index.Columns[i].Length = defaultIndexPrefixLength
						indexModified = true
					}
				}
			}
			if indexModified {
				newNode, err := node.WithIndexDefs(newIndexes)
				return newNode, transform.NewTree, err
			}

		case *plan.AlterIndex:
			if node.Action == plan.IndexAction_Create {
				colMap := schToColMap(targetSchema)
				newColumns := make([]sql.IndexColumn, len(node.Columns))
				for i := range node.Columns {
					copy := node.Columns[i]
					newColumns[i] = copy
				}
				indexModified := false
				for i := range newColumns {
					col, ok := colMap[strings.ToLower(newColumns[i].Name)]
					if !ok {
						return nil, false, fmt.Errorf("indexed column %s not found in schema", newColumns[i].Name)
					}
					if dt, ok := col.Type.(pgtypes.DoltgresType); ok && dt.OID == uint32(oid.T_text) && newColumns[i].Length == 0 {
						newColumns[i].Length = defaultIndexPrefixLength
						indexModified = true
					}
				}
				if indexModified {
					newNode, err := node.WithColumns(newColumns)
					return newNode, transform.NewTree, err
				}
			}
		}
		return node, transform.SameTree, nil
	})
}

func schToColMap(sch sql.Schema) map[string]*sql.Column {
	colMap := make(map[string]*sql.Column, len(sch))
	for _, col := range sch {
		colMap[strings.ToLower(col.Name)] = col
	}
	return colMap
}
