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
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeUnresolvedObjectName handles *tree.UnresolvedObjectName nodes.
func nodeUnresolvedObjectName(node *tree.UnresolvedObjectName) (vitess.TableName, error) {
	if node == nil {
		return vitess.TableName{}, nil
	}

	var tableName vitess.TableIdent
	var dbQual vitess.TableIdent
	var schemaQual vitess.TableIdent

	tableName = vitess.NewTableIdent(node.Parts[0])

	if node.NumParts > 1 {
		schemaQual = vitess.NewTableIdent(node.Parts[1])
	}

	if node.NumParts > 2 {
		dbQual = vitess.NewTableIdent(node.Parts[2])
	}

	return vitess.TableName{
		Name:            tableName,
		DbQualifier:     dbQual,
		SchemaQualifier: schemaQual,
	}, nil
}
