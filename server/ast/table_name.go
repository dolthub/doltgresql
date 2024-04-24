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

// nodeTableName handles *tree.TableName nodes.
func nodeTableName(node *tree.TableName) (vitess.TableName, error) {
	if node == nil {
		return vitess.TableName{}, nil
	}

	var dbName, schemaName vitess.TableIdent

	if node.ExplicitCatalog {
		dbName = vitess.NewTableIdent(string(node.CatalogName))
	}

	if node.ExplicitSchema {
		schemaName = vitess.NewTableIdent(string(node.SchemaName))
	}

	return vitess.TableName{
		Name:            vitess.NewTableIdent(string(node.ObjectName)),
		DbQualifier:     dbName,
		SchemaQualifier: schemaName,
	}, nil
}
