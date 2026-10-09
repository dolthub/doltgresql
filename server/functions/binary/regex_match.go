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

package binary

import (
	regex "github.com/dolthub/go-icu-regex"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = '~' ORDER BY o.oprcode::varchar;

// initBinaryRegexMatch registers the functions to the catalog.
func initBinaryRegexMatch() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryRegexMatch, bpcharregexeq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryRegexMatch, nameregexeq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryRegexMatch, textregexeq)
}

// bpcharregexeq represents the PostgreSQL function of the same name, taking the same parameters.
var bpcharregexeq = framework.Function2{
	Name:       "bpcharregexeq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.BpChar, pgtypes.Text},
	Strict:     true,
	Callable:   textregexeq_callable,
}

// nameregexeq represents the PostgreSQL function of the same name, taking the same parameters.
var nameregexeq = framework.Function2{
	Name:       "nameregexeq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Name, pgtypes.Text},
	Strict:     true,
	Callable:   textregexeq_callable,
}

// textregexeq_callable is the callable logic for the textregexeq function.
func textregexeq_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	return regexMatches(ctx, val1, val2)
}

// textregexeq represents the PostgreSQL function of the same name, taking the same parameters.
var textregexeq = framework.Function2{
	Name:       "textregexeq",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable:   textregexeq_callable,
}

// regexMatches returns whether the regular expression `pattern` matches any part of `str`.
func regexMatches(ctx *sql.Context, str any, pattern any) (bool, error) {
	strVal, err := framework.UnwrapString(ctx, str)
	if err != nil {
		return false, err
	}
	patternVal, err := framework.UnwrapString(ctx, pattern)
	if err != nil {
		return false, err
	}
	if len(patternVal) == 0 {
		return true, nil
	}
	re := regex.CreateRegex(0)
	defer re.Close()
	if err = re.SetRegexString(ctx, patternVal, regex.RegexFlags_None); err != nil {
		return false, pgerror.Newf(pgcode.InvalidRegularExpression, "invalid regular expression: %s", err.Error())
	}
	if err = re.SetMatchString(ctx, strVal); err != nil {
		return false, err
	}
	return re.Matches(ctx, 0, 0)
}
