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
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPgGetUserbyid registers the functions to the catalog.
func initPgGetUserbyid() {
	framework.RegisterFunction(pg_get_userbyid_oid)
}

// pg_get_userbyid_oid represents the PostgreSQL system catalog information function.
var pg_get_userbyid_oid = framework.Function1{
	Name:               "pg_get_userbyid",
	Return:             pgtypes.Text,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Oid},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		internalID := val.(id.Id)
		// TODO: replace the id.Null owner placeholders in older catalog handlers.
		if internalID == id.Null {
			return "postgres", nil
		}
		var role auth.Role
		auth.LockRead(func() {
			if internalID.Section() == id.Section_User {
				role = auth.GetRole(internalID.Segment(0))
			}
		})
		if role.IsValid() {
			return role.Name, nil
		}
		return fmt.Sprintf("unknown (OID=%d)", id.Cache().ToOID(internalID)), nil
	},
}
