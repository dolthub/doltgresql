// Copyright 2026 Dolthub, Inc.
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

	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	"github.com/dolthub/doltgresql/server/plpgsql"
)

// nodeDo converts an anonymous procedural code block.
func nodeDo(ctx *Context, stmt *tree.Do) (vitess.Statement, error) {
	if stmt.Language != "" && !strings.EqualFold(stmt.Language, "plpgsql") {
		if strings.EqualFold(stmt.Language, "sql") {
			return nil, pgerror.Newf(pgcode.FeatureNotSupported, "language %q does not support inline code execution", stmt.Language)
		}
		return nil, pgerror.Newf(pgcode.UndefinedObject, "language %q does not exist", stmt.Language)
	}
	operations, err := plpgsql.Parse(ctx.originalQuery)
	if err != nil {
		return nil, err
	}
	for i, operation := range operations {
		if operation.OpCode != plpgsql.OpCode_Declare {
			continue
		}
		parsedType, err := parser.ParseType(operation.PrimaryData)
		if err != nil {
			return nil, err
		}
		_, resolvedType, err := nodeResolvableTypeReference(ctx, parsedType, false)
		if err != nil {
			return nil, err
		}
		if resolvedType == nil {
			return nil, fmt.Errorf("type %q could not be resolved", operation.PrimaryData)
		}
		typeName := resolvedType.Name()
		if resolvedType.Schema() != "" {
			typeName = fmt.Sprintf("%s.%s", resolvedType.Schema(), typeName)
		}
		operations[i].PrimaryData = typeName
	}
	return vitess.InjectedStatement{Statement: pgnodes.NewDo(operations)}, nil
}
