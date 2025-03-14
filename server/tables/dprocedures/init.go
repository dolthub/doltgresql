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

package dprocedures

import (
	"fmt"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dprocedures"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/auth"
)

// Init handles initialization of all Postgres-specific and Doltgres-specific Dolt procedures.
func Init() {
	dprocedures.UserHasSuperAccess = userHasSuperAccess
}

// userHasSuperAccess is a function that checks if the current user has super access.
func userHasSuperAccess(ctx *sql.Context) (bool, error) {
	var userRole auth.Role
	auth.LockRead(func() {
		userRole = auth.GetRole(ctx.Client().User)
	})
	if !userRole.IsValid() {
		return false, fmt.Errorf(`role "%s" does not exist`, ctx.Client().User)
	}
	return userRole.IsSuperUser, nil
}
