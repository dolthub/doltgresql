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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = '!~' ORDER BY o.oprcode::varchar;

// initBinaryRegexNotMatch registers the functions to the catalog.
func initBinaryRegexNotMatch() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryRegexNotMatch, bpcharregexne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryRegexNotMatch, nameregexne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryRegexNotMatch, textregexne)
}

// bpcharregexne represents the PostgreSQL function of the same name, taking the same parameters.
var bpcharregexne = framework.Function2{
	Name:       "bpcharregexne",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.BpChar, pgtypes.Text},
	Strict:     true,
	Callable:   textregexne_callable,
}

// nameregexne represents the PostgreSQL function of the same name, taking the same parameters.
var nameregexne = framework.Function2{
	Name:       "nameregexne",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Name, pgtypes.Text},
	Strict:     true,
	Callable:   textregexne_callable,
}

// textregexne_callable is the callable logic for the textregexne function.
func textregexne_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	matches, err := regexMatches(ctx, val1, val2)
	return !matches, err
}

// textregexne represents the PostgreSQL function of the same name, taking the same parameters.
var textregexne = framework.Function2{
	Name:       "textregexne",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable:   textregexne_callable,
}
