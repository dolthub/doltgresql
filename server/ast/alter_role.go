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
	"errors"
	"fmt"
	"strings"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgnodes "github.com/dolthub/doltgresql/server/node"
)

// nodeAlterRole handles *tree.AlterRole nodes.
func nodeAlterRole(ctx *Context, node *tree.AlterRole) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	if len(node.Name) == 0 {
		// The parser should make this impossible, but extra error checking is never bad
		return nil, errors.New(`role name cannot be empty`)
	}
	// Some of the keys are used as markers, and do not contain values.
	// Therefore, the values will be nil since they're ignored.
	options := make(map[string]any)
	for _, kvOption := range node.KVOptions {
		optionName := strings.ToUpper(string(kvOption.Key))
		switch optionName {
		case "BYPASSRLS":
			options["BYPASSRLS"] = nil
		case "CONNECTION_LIMIT":
			switch value := kvOption.Value.(type) {
			case *tree.DInt:
				if value == nil {
					options["CONNECTION_LIMIT"] = int32(-1)
				} else {
					// We enforce that only int32 values will fit here in the parser
					options["CONNECTION_LIMIT"] = int32(*value)
				}
			case tree.NullLiteral:
				options["CONNECTION_LIMIT"] = int32(-1)
			default:
				return nil, fmt.Errorf(`unknown role option value (%T) for option "%s"`, kvOption.Value, kvOption.Key)
			}
		case "CREATEDB":
			options["CREATEDB"] = nil
		case "CREATEROLE":
			options["CREATEROLE"] = nil
		case "INHERIT":
			options["INHERIT"] = nil
		case "LOGIN":
			options["LOGIN"] = nil
		case "NOBYPASSRLS":
			options["NOBYPASSRLS"] = nil
		case "NOCREATEDB":
			options["NOCREATEDB"] = nil
		case "NOCREATEROLE":
			options["NOCREATEROLE"] = nil
		case "NOINHERIT":
			options["NOINHERIT"] = nil
		case "NOLOGIN":
			options["NOLOGIN"] = nil
		case "NOREPLICATION":
			options["NOREPLICATION"] = nil
		case "NOSUPERUSER":
			options["NOSUPERUSER"] = nil
		case "PASSWORD":
			switch value := kvOption.Value.(type) {
			case *tree.DString:
				if value == nil {
					options["PASSWORD"] = nil
				} else {
					options["PASSWORD"] = (*string)(value)
				}
			case tree.NullLiteral:
				options["PASSWORD"] = nil
			default:
				return nil, fmt.Errorf(`unknown role option value (%T) for option "%s"`, kvOption.Value, kvOption.Key)
			}
		case "REPLICATION":
			options["REPLICATION"] = nil
		case "SUPERUSER":
			options["SUPERUSER"] = nil
		case "SYSID":
			// This is an option that is ignored by Postgres. Assuming it used to be relevant, but not any longer.
		case "VALID_UNTIL":
			strVal, ok := kvOption.Value.(*tree.DString)
			if !ok {
				return nil, fmt.Errorf(`unknown role option value (%T) for option "%s"`, kvOption.Value, kvOption.Key)
			}
			if strVal == nil {
				options["VALID_UNTIL"] = nil
			} else {
				options["VALID_UNTIL"] = (*string)(strVal)
			}
		default:
			return nil, fmt.Errorf(`unknown role option "%s"`, kvOption.Key)
		}
	}
	return vitess.InjectedStatement{
		Statement: &pgnodes.AlterRole{
			Name:    node.Name,
			Options: options,
		},
		Children: nil,
	}, nil
}
