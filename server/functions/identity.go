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

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/sessionstate"
	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

func initIdentity() {
	for _, name := range []string{"current_user", "current_role", "user", "session_user"} {
		name := name
		framework.RegisterFunction(framework.Function0{
			Name: name, Return: pgtypes.Name, IsNonDeterministic: true,
			Callable: func(ctx *sql.Context) (any, error) {
				return identityName(ctx, name)
			},
		})
	}
}

func identityName(ctx *sql.Context, name string) (string, error) {
	identity, err := core.Identity(ctx)
	if err != nil {
		return "", err
	}
	var id sessionstate.RoleID
	if name == "session_user" {
		id = identity.SessionRole()
	} else {
		id = identity.CurrentRole()
	}
	role, err := auth.ResolveRoleID(id)
	if err != nil {
		return "", err
	}
	return role.Name, nil
}
