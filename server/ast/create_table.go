// Copyright 2023 Dolthub, Inc.
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

package ast

import (
	"fmt"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeCreateTable handles *tree.CreateTable nodes.
func nodeCreateTable(node *tree.CreateTable) (*vitess.DDL, error) {
	if node == nil {
		return nil, nil
	}
	if node.Interleave != nil {
		return nil, fmt.Errorf("INTERLEAVE is not yet supported")
	}
	if node.PartitionBy != nil {
		return nil, fmt.Errorf("PARTITION BY is not yet supported")
	}
	if len(node.StorageParams) > 0 {
		return nil, fmt.Errorf("storage parameters are not yet supported")
	}
	if node.OnCommit != tree.CreateTableOnCommitUnset {
		return nil, fmt.Errorf("ON COMMIT is not yet supported")
	}
	tableName, err := nodeTableName(&node.Table)
	if err != nil {
		return nil, err
	}
	var isTemporary bool
	switch node.Persistence {
	case tree.PersistencePermanent:
		isTemporary = false
	case tree.PersistenceTemporary:
		isTemporary = true
	case tree.PersistenceUnlogged:
		return nil, fmt.Errorf("UNLOGGED is not yet supported")
	default:
		return nil, fmt.Errorf("unknown persistence strategy encountered")
	}
	var optSelect *vitess.OptSelect
	if node.Using != "" {
		return nil, fmt.Errorf("USING is not yet supported")
	}
	if node.Tablespace != "" {
		return nil, fmt.Errorf("TABLESPACE is not yet supported")
	}
	if node.AsSource != nil {
		selectStmt, err := nodeSelect(node.AsSource)
		if err != nil {
			return nil, err
		}
		optSelect = &vitess.OptSelect{
			Select: selectStmt,
		}
	}
	if node.WithNoData {
		return nil, fmt.Errorf("WITH NO DATA is not yet supported")
	}
	ddl := &vitess.DDL{
		Action:      vitess.CreateStr,
		Table:       tableName,
		IfNotExists: node.IfNotExists,
		Temporary:   isTemporary,
		OptSelect:   optSelect,
	}
	if err = assignTableDefs(node.Defs, ddl); err != nil {
		return nil, err
	}
	return ddl, nil
}
