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

package binary

import (
	"regexp"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = '~' ORDER BY o.oprcode::varchar;

// initBinaryRegMatch registers the functions to the catalog.
func initBinaryRegMatch() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryRegMatch, textregexeq)
}

// textregexeq_callable is the callable logic for the textregexeq function.
func textregexeq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	input, err := unwrapTextOperand(ctx, val1, "regex input")
	if err != nil {
		return nil, err
	}
	pattern, err := unwrapTextOperand(ctx, val2, "regex pattern")
	if err != nil {
		return nil, err
	}
	return regexp.MatchString(pattern, input)
}

// unwrapTextOperand converts a text operator argument to a string.
func unwrapTextOperand(ctx *sql.Context, val any, operandName string) (string, error) {
	str, ok, err := sql.Unwrap[string](ctx, val)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.Errorf("unexpected type for %s, expected string, got %T", operandName, val)
	}
	return str, nil
}

// textregexeq represents the PostgreSQL function of the same name, taking the same parameters.
var textregexeq = framework.Function2{
	Name:       "textregexeq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable:   textregexeq_callable,
}
