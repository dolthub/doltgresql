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
	"strings"

	"github.com/cockroachdb/errors"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/postgres/parser/types"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	"github.com/dolthub/doltgresql/server/plpgsql"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// nodeCreateFunction handles *tree.CreateFunction nodes.
func nodeCreateFunction(ctx *Context, node *tree.CreateFunction) (vitess.Statement, error) {
	options, err := validateRoutineOptions(ctx, node.Options)
	if err != nil {
		return nil, err
	}
	// We only support PL/pgSQL for now, so we'll verify that first
	if languageOption, ok := options[tree.OptionLanguage]; ok {
		if strings.ToLower(languageOption.Language) != "plpgsql" {
			return nil, errors.Errorf("CREATE FUNCTION only supports PL/pgSQL for now")
		}
	} else {
		return nil, errors.Errorf("CREATE FUNCTION does not define an input language")
	}
	// PL/pgSQL is different from standard Postgres SQL, so we have to use a special parser to handle it.
	// This parser also requires the full `CREATE FUNCTION` string, so we'll pass that.
	parsedBody, err := plpgsql.Parse(ctx.originalQuery)
	if err != nil {
		return nil, err
	}
	// Grab the rest of the information that we'll need to create the function
	tableName := node.Name.ToTableName()
	schemaName := tableName.Schema()
	if len(schemaName) == 0 {
		// TODO: fix function finder such that it doesn't always assume pg_catalog
		schemaName = "pg_catalog"
	}
	retType := pgtypes.Void
	if len(node.RetType) == 1 {
		switch typ := node.RetType[0].Type.(type) {
		case *types.T:
			retType = pgtypes.NewUnresolvedDoltgresType("", strings.ToLower(typ.Name()))
		default:
			retType = pgtypes.NewUnresolvedDoltgresType("", strings.ToLower(typ.SQLString()))
		}
	}
	paramNames := make([]string, len(node.Args))
	paramTypes := make([]*pgtypes.DoltgresType, len(node.Args))
	for i, arg := range node.Args {
		paramNames[i] = arg.Name.String()
		switch argType := arg.Type.(type) {
		case *types.T:
			paramTypes[i] = pgtypes.NewUnresolvedDoltgresType("", strings.ToLower(argType.Name()))
		default:
			paramTypes[i] = pgtypes.NewUnresolvedDoltgresType("", strings.ToLower(argType.SQLString()))
		}
	}
	// Returns the stored procedure call with all options
	return vitess.InjectedStatement{
		Statement: pgnodes.NewCreateFunction(
			tableName.Table(),
			schemaName,
			node.Replace,
			retType,
			paramNames,
			paramTypes,
			true, // TODO: implement strict check
			parsedBody,
		),
		Children: nil,
	}, nil
}
