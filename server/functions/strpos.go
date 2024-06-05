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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initStrpos registers the functions to the catalog.
func initStrpos() {
	framework.RegisterFunction(strpos_varchar)
}

// strpos_varchar represents the PostgreSQL function of the same name, taking the same parameters.
var strpos_varchar = framework.Function2{
	Name:       "strpos",
	Return:     pgtypes.Int32,
	Parameters: []pgtypes.DoltgresType{pgtypes.VarChar, pgtypes.VarChar},
	Callable: func(ctx *sql.Context, str any, substring any) (any, error) {
		if str == nil || substring == nil {
			return nil, nil
		}
		idx := strings.Index(str.(string), substring.(string))
		if idx == -1 {
			return int32(0), nil
		}
		return int32(idx + 1), nil
	},
}
