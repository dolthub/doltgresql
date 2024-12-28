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

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgnodes "github.com/dolthub/doltgresql/server/node"
)

// nodeCopyFrom handles *tree.CopyFrom nodes.
func nodeCopyFrom(ctx *Context, node *tree.CopyFrom) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	if node.Options.CopyFormat == tree.CopyFormatBinary {
		return nil, fmt.Errorf("COPY FROM does not support format BINARY")
	}

	// We start by creating a stub insert statement for the COPY FROM statement, which we will use to build a basic 
	// INSERT plan for. At runtime we will swap out the bogus row values for our actual data read from STDIN. 
	var columns []vitess.ColIdent
	if len(node.Columns) > 0 {
		columns = make([]vitess.ColIdent, len(node.Columns))
		for i := range node.Columns {
			columns[i] = vitess.NewColIdent(string(node.Columns[i]))
		}
	}

	tableName, err := nodeTableName(ctx, &node.Table)
	if err != nil {
		return nil, err
	}

	stubValues := make(vitess.Values, 1)
	stubValues[0] = make(vitess.ValTuple, len(columns))
	for i := range columns {
		// TODO: does this actually work? A select might be better
		stubValues[0][i] = &vitess.NullVal{}
	}
	
	return vitess.InjectedStatement{
		Statement: pgnodes.NewCopyFrom(
			node.Table.Catalog(),
			doltdb.TableName{
				Name:   node.Table.Object(),
				Schema: node.Table.Schema(),
			},
			node.Options,
			node.File,
			node.Stdin,
			node.Columns,
			&vitess.Insert{
				Action:  vitess.InsertStr,
				Table:   tableName,
				Columns: columns,
				Rows:    &vitess.AliasedValues{
					Values:  stubValues,
				},
			},
		),
		Children: nil,
	}, nil
}
