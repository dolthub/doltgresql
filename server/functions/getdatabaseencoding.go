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

package functions

import (
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initGetdatabaseencoding registers the functions to the catalog.
func initGetdatabaseencoding() {
	framework.RegisterFunction(getdatabaseencoding)
}

// getdatabaseencoding represents the PostgreSQL system information function of the same name, taking no parameters.
var getdatabaseencoding = framework.Function0{
	Name:   "getdatabaseencoding",
	Return: pgtypes.Name,
	Strict: true,
	Callable: func(ctx *sql.Context) (any, error) {
		// TODO: UTF8 is the only server encoding supported for now
		return "UTF8", nil
	},
}
