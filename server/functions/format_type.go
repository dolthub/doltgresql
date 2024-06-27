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

package functions

import (
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/go-mysql-server/sql"
)

// initFormatType registers the functions to the catalog.
func initFormatType() {
	framework.RegisterFunction(format_type)
}

// format_type represents the PostgreSQL system information function.
var format_type = framework.Function2{
	Name:               "format_type",
	Return:             pgtypes.Text,
	Parameters:         []pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Int32},
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, val1, val2 any) (any, error) {
		if val1 == nil {
			return nil, nil
		}
		// TODO: retrieve type by its OID
		//  if the OID does not match any type, return "???"

		if val2 != nil {
			// val2 is "typtypmod" of "pg_type", which is optional
			// TODO: if it's not -1 in pg_type, then it gets concatenated to the output.
		}
		return "", nil
	},
	Strict: false,
}
