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
	"strings"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgnodes "github.com/dolthub/doltgresql/server/node"
)

// nodeAlterRoleSetting handles the SET and RESET forms of ALTER ROLE and ALTER DATABASE. The returned node only needs
// its target to be filled in.
func nodeAlterRoleSetting(setVar *tree.SetVar, resetAll bool) (*pgnodes.AlterRoleSetting, error) {
	if resetAll {
		return &pgnodes.AlterRoleSetting{ResetAll: true}, nil
	}
	name := setVar.Name
	if setVar.Namespace != "" {
		name = setVar.Namespace + "." + setVar.Name
	}
	if setVar.Namespace == "" && (strings.EqualFold(name, "role") || strings.EqualFold(name, "session_authorization")) {
		// These are handled specially by SET, and cannot be stored as settings
		return nil, pgerror.Newf(pgcode.CantChangeRuntimeParam, `parameter "%s" cannot be changed now`, name)
	}
	setting := &pgnodes.AlterRoleSetting{Name: name}
	switch {
	case setVar.Reset:
		setting.Reset = true
	case setVar.FromCurrent:
		setting.FromCurrent = true
	case len(setVar.Values) == 1 && isDefaultVal(setVar.Values[0]):
		// SET ... TO DEFAULT is equivalent to RESET
		setting.Reset = true
	default:
		value, err := roleSettingValue(name, setVar.Values)
		if err != nil {
			return nil, err
		}
		setting.Value = value
	}
	return setting, nil
}

// roleSettingValue renders the values of a SET clause as the text that is stored for the parameter, which is applied
// when a session starts. This matches Postgres's flatten_set_variable_args.
func roleSettingValue(name string, values tree.Exprs) (string, error) {
	if flattened, ok := flattenIdentifierList(name, values); ok {
		return flattened, nil
	}
	vals := make([]string, len(values))
	for i, value := range values {
		val, ok := constantSettingValue(value)
		if !ok {
			return "", pgerror.Newf(pgcode.Syntax, `syntax error at or near "%s"`, tree.AsString(value))
		}
		vals[i] = val
	}
	return strings.Join(vals, ", "), nil
}

// constantSettingValue returns the text of a constant value given in a SET clause. Returns false if the value is not
// a constant.
func constantSettingValue(value tree.Expr) (string, bool) {
	switch value := value.(type) {
	case *tree.StrVal:
		return value.RawString(), true
	case *tree.NumVal:
		return value.FormattedString(), true
	case *tree.UnresolvedName:
		if value.NumParts != 1 || value.Star {
			return "", false
		}
		return value.Parts[0], true
	case *tree.DBool:
		if *value {
			return "true", true
		}
		return "false", true
	case *tree.UnaryExpr:
		if num, ok := value.Expr.(*tree.NumVal); ok && value.Operator == tree.UnaryMinus {
			return "-" + num.FormattedString(), true
		}
	}
	return "", false
}

// isDefaultVal returns whether the expression is DEFAULT.
func isDefaultVal(expr tree.Expr) bool {
	_, ok := expr.(tree.DefaultVal)
	return ok
}

// alterRoleSettingStatement wraps the node as a vitess statement.
func alterRoleSettingStatement(setting *pgnodes.AlterRoleSetting) vitess.Statement {
	return vitess.InjectedStatement{
		Statement: setting,
		Children:  nil,
	}
}
