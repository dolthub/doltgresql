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

package ast

import (
	"fmt"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeUniqueConstraintTableDef converts a tree.UniqueConstraintTableDef instance
// into a vitess.DDL instance that can be executed by GMS. |tableName| identifies
// the table being altered, and |ifExists| indicates whether the IF EXISTS clause
// was specified.
func nodeUniqueConstraintTableDef(
	node *tree.UniqueConstraintTableDef,
	tableName vitess.TableName,
	ifExists bool) (*vitess.DDL, error) {

	if len(node.IndexParams.StorageParams) > 0 {
		return nil, fmt.Errorf("STORAGE parameters not yet supported for indexes")
	}

	if node.IndexParams.Tablespace != "" {
		return nil, fmt.Errorf("TABLESPACE is not yet supported")
	}

	if node.NullsNotDistinct {
		return nil, fmt.Errorf("NULLS NOT DISTINCT is not yet supported")
	}

	columns, err := nodeIndexElemList(node.Columns)
	if err != nil {
		return nil, err
	}

	if node.PrimaryKey {
		return &vitess.DDL{
			Action:   "alter",
			Table:    tableName,
			IfExists: ifExists,
			IndexSpec: &vitess.IndexSpec{
				Action:  "create",
				Type:    "primary",
				Columns: columns,
			},
		}, nil
	} else {
		return nil, fmt.Errorf("Only PRIMARY KEY constraints are supported currently")
	}
}
