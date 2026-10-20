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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initHasFunctionPrivilege registers the functions to the catalog.
func initHasFunctionPrivilege() {
	framework.RegisterFunction(has_function_privilege_name_text_text)
	framework.RegisterFunction(has_function_privilege_name_oid_text)
	framework.RegisterFunction(has_function_privilege_oid_text_text)
	framework.RegisterFunction(has_function_privilege_oid_oid_text)
	framework.RegisterFunction(has_function_privilege_text_text)
	framework.RegisterFunction(has_function_privilege_oid_text)
}

// has_function_privilege_name_text_text represents the PostgreSQL function of the same name, taking the same parameters.
var has_function_privilege_name_text_text = framework.Function3{
	Name:               "has_function_privilege",
	Return:             pgtypes.Bool,
	Parameters:         [3]*pgtypes.DoltgresType{pgtypes.Name, pgtypes.Text, pgtypes.Text},
	Strict:             true,
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		return hasFunctionPrivilege(ctx, val1, val2, val3)
	},
}

// has_function_privilege_name_oid_text represents the PostgreSQL function of the same name, taking the same parameters.
var has_function_privilege_name_oid_text = framework.Function3{
	Name:               "has_function_privilege",
	Return:             pgtypes.Bool,
	Parameters:         [3]*pgtypes.DoltgresType{pgtypes.Name, pgtypes.Oid, pgtypes.Text},
	Strict:             true,
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		return hasFunctionPrivilege(ctx, val1, val2, val3)
	},
}

// has_function_privilege_oid_text_text represents the PostgreSQL function of the same name, taking the same parameters.
var has_function_privilege_oid_text_text = framework.Function3{
	Name:               "has_function_privilege",
	Return:             pgtypes.Bool,
	Parameters:         [3]*pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Text, pgtypes.Text},
	Strict:             true,
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		return hasFunctionPrivilege(ctx, val1, val2, val3)
	},
}

// has_function_privilege_oid_oid_text represents the PostgreSQL function of the same name, taking the same parameters.
var has_function_privilege_oid_oid_text = framework.Function3{
	Name:               "has_function_privilege",
	Return:             pgtypes.Bool,
	Parameters:         [3]*pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Oid, pgtypes.Text},
	Strict:             true,
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		return hasFunctionPrivilege(ctx, val1, val2, val3)
	},
}

// has_function_privilege_text_text represents the PostgreSQL function of the same name, taking the same parameters.
var has_function_privilege_text_text = framework.Function2{
	Name:               "has_function_privilege",
	Return:             pgtypes.Bool,
	Parameters:         [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:             true,
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		return hasFunctionPrivilege(ctx, nil, val1, val2)
	},
}

// has_function_privilege_oid_text represents the PostgreSQL function of the same name, taking the same parameters.
var has_function_privilege_oid_text = framework.Function2{
	Name:               "has_function_privilege",
	Return:             pgtypes.Bool,
	Parameters:         [2]*pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Text},
	Strict:             true,
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		return hasFunctionPrivilege(ctx, nil, val1, val2)
	},
}

// hasFunctionPrivilege uses the same routine ACLs as execution. A nil user selects
// the effective current role; SQL NULL arguments are handled by Strict.
func hasFunctionPrivilege(ctx *sql.Context, user, target, privilege any) (any, error) {
	var actor auth.Role
	var err error
	auth.LockRead(func() {
		if user == nil {
			actor, err = auth.CurrentRoleLocked(ctx)
		} else if _, oid := user.(id.Id); oid {
			actor, _, err = pgHasRoleLookup(ctx, user)
		} else {
			var name string
			name, err = framework.UnwrapString(ctx, user)
			if err != nil {
				return
			}
			if name == "public" {
				actor = auth.GetRole("public")
			} else {
				actor, _, err = pgHasRoleLookup(ctx, user)
			}
		}
	})
	if err != nil {
		return nil, err
	}
	routine, byOID := target.(id.Id)
	if !byOID {
		name, err := framework.UnwrapString(ctx, target)
		if err != nil {
			return nil, err
		}
		resolved, err := pgtypes.Regprocedure.IoInput(ctx, name)
		if err != nil {
			return nil, err
		}
		routine = resolved.(id.Id)
		if id.Cache().ToOID(routine) == 0 {
			return nil, pgerror.Newf(pgcode.UndefinedFunction, "function %q does not exist", name)
		}
	}
	privileges, err := framework.UnwrapString(ctx, privilege)
	if err != nil {
		return nil, err
	}
	var execute, grant bool
	for _, item := range strings.Split(privileges, ",") {
		switch strings.ToUpper(strings.TrimSpace(item)) {
		case "EXECUTE":
			execute = true
		case "EXECUTE WITH GRANT OPTION":
			grant = true
		default:
			return nil, pgerror.Newf(pgcode.InvalidParameterValue, "unrecognized privilege type: %q", strings.TrimSpace(item))
		}
	}
	exists, builtin, err := functionPrivilegeRoutineExists(ctx, routine)
	if err != nil {
		return nil, err
	}
	if !exists {
		if byOID {
			return nil, nil
		}
		// PostgreSQL's text overload accepts numeric regprocedure input. Its
		// ACL check lets superusers bypass a missing nonzero numeric OID.
		if actor.IsSuperUser {
			return true, nil
		}
		return nil, pgerror.Newf(pgcode.UndefinedFunction, "function with OID %d does not exist", id.Cache().ToOID(routine))
	}
	var result bool
	auth.LockRead(func() {
		key := auth.RoutinePrivilegeKey{Role: actor.ID(), Schema: routine.Segment(0), Name: routine.Segment(1)}
		publicKey := key
		public := auth.GetRole("public")
		publicKey.Role = public.ID()
		// Like checkPrivilegeOnRoutine, ACLs currently apply to all overloads of a
		// name. Built-ins have implicit PUBLIC EXECUTE until revoking it is supported.
		result = (execute && (builtin || auth.HasRoutinePrivilege(key, auth.Privilege_EXECUTE) || auth.HasRoutinePrivilege(publicKey, auth.Privilege_EXECUTE))) ||
			(grant && (auth.HasRoutinePrivilegeGrantOption(key, auth.Privilege_EXECUTE).IsValid() || auth.HasRoutinePrivilegeGrantOption(publicKey, auth.Privilege_EXECUTE).IsValid()))
	})
	return result, nil
}

// functionPrivilegeRoutineExists checks the live catalog, not just the OID cache,
// which can still contain IDs for dropped routines.
func functionPrivilegeRoutineExists(ctx *sql.Context, routine id.Id) (exists, builtin bool, err error) {
	switch routine.Section() {
	case id.Section_Function:
		if isBuiltinRoutine(routine) {
			return true, true, nil
		}
		collection, err := core.GetFunctionsCollectionFromContext(ctx, "")
		if err != nil {
			return false, false, err
		}
		return collection.HasFunction(ctx, id.Function(routine)), false, nil
	case id.Section_Procedure:
		collection, err := core.GetProceduresCollectionFromContext(ctx, "")
		if err != nil {
			return false, false, err
		}
		return collection.HasProcedure(ctx, id.Procedure(routine)), false, nil
	default:
		return false, false, nil
	}
}
