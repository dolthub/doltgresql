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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/sequences"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// ReplaceSerial replaces a CreateTable node containing a SERIAL type with a node that can create sequences alongside
// the table.
func ReplaceSerial(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	createTable, ok := node.(*plan.CreateTable)
	if !ok {
		return node, transform.SameTree, nil
	}

	var ctSequences []*pgnodes.CreateSequence
	for _, col := range createTable.PkSchema().Schema {
		if doltgresType, ok := col.Type.(*pgtypes.DoltgresType); ok {
			if doltgresType.IsSerial {
				var maxValue int64
				switch doltgresType.Name() {
				case "smallserial":
					col.Type = pgtypes.Int16
					maxValue = 32767
				case "serial":
					col.Type = pgtypes.Int32
					maxValue = 2147483647
				case "bigserial":
					col.Type = pgtypes.Int64
					maxValue = 9223372036854775807
				}

				baseSequenceName := fmt.Sprintf("%s_%s_seq", createTable.Name(), col.Name)
				sequenceName := baseSequenceName
				schemaName, err := core.GetSchemaName(ctx, createTable.Db, "")
				if err != nil {
					return nil, false, err
				}

				relationType, err := core.GetRelationType(ctx, schemaName, baseSequenceName)
				if err != nil {
					return nil, transform.NewTree, err
				}
				if relationType != core.RelationType_DoesNotExist {
					seqIndex := 1
					for ; seqIndex <= 100; seqIndex++ {
						sequenceName = fmt.Sprintf("%s%d", baseSequenceName, seqIndex)
						relationType, err = core.GetRelationType(ctx, schemaName, baseSequenceName)
						if err != nil {
							return nil, transform.NewTree, err
						}
						if relationType == core.RelationType_DoesNotExist {
							break
						}
					}
					if seqIndex > 100 {
						return nil, transform.NewTree, errors.Errorf("SERIAL sequence name reached max iterations")
					}
				}

				seqName := doltdb.TableName{Name: sequenceName, Schema: schemaName}.String()
				nextVal, ok, err := framework.GetFunction("nextval", pgexprs.NewTextLiteral(seqName))
				if err != nil {
					return nil, transform.NewTree, err
				}
				if !ok {
					return nil, transform.NewTree, errors.Errorf(`function "nextval" could not be found for SERIAL default`)
				}
				col.Default = &sql.ColumnDefaultValue{
					Expr:          nextVal,
					OutType:       pgtypes.Int64,
					Literal:       false,
					ReturnNil:     false,
					Parenthesized: false,
				}
				ctSequences = append(ctSequences, pgnodes.NewCreateSequence(false, "", &sequences.Sequence{
					Id:          id.NewSequence("", sequenceName),
					DataTypeID:  col.Type.(*pgtypes.DoltgresType).ID,
					Persistence: sequences.Persistence_Permanent,
					Start:       1,
					Current:     1,
					Increment:   1,
					Minimum:     1,
					Maximum:     maxValue,
					Cache:       1,
					Cycle:       false,
					IsAtEnd:     false,
					OwnerTable:  id.NewTable("", createTable.Name()),
					OwnerColumn: col.Name,
				}))
			}
		}
	}
	return pgnodes.NewCreateTable(createTable, ctSequences), transform.NewTree, nil
}
