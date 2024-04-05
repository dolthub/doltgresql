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
	"strings"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeTableName handles *tree.TableName nodes.
func nodeTableName(node *tree.TableName) (vitess.TableName, error) {
	if node == nil {
		return vitess.TableName{}, nil
	}

	if node.ExplicitCatalog || node.ExplicitSchema {
		if strings.ToLower(string(node.SchemaName)) == "information_schema" {
			return vitess.TableName{
				Name:        vitess.NewTableIdent(string(node.ObjectName)),
				DbQualifier: vitess.NewTableIdent(string(node.SchemaName)),
			}, nil
		} else if !node.ExplicitCatalog && node.ExplicitSchema && strings.ToLower(string(node.SchemaName)) == "public" {
			// the "public" schema is the default schema in PostgreSQL, so treat it as if it were not explicitly specified
			return vitess.TableName{
				Name:        vitess.NewTableIdent(string(node.ObjectName)),
				DbQualifier: vitess.NewTableIdent(""),
			}, nil
		}
		return vitess.TableName{}, fmt.Errorf("referencing items outside the schema or database is not yet supported")
	}

	return vitess.TableName{
		Name:        vitess.NewTableIdent(string(node.ObjectName)),
		DbQualifier: vitess.NewTableIdent(string(node.SchemaName)),
	}, nil
}
