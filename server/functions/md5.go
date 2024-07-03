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
	md5_package "crypto/md5"
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initMd5 registers the functions to the catalog.
func initMd5() {
	framework.RegisterFunction(md5_varchar)
}

// md5_varchar represents the PostgreSQL function of the same name, taking the same parameters.
var md5_varchar = framework.Function1{
	Name:       "md5",
	Return:     pgtypes.VarChar,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.VarChar},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val1 any) (any, error) {
		return fmt.Sprintf("%x", md5_package.Sum([]byte(val1.(string)))), nil
	},
}
