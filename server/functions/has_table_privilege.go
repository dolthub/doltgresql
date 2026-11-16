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

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initHasTablePrivilege registers the functions to the catalog.
func initHasTablePrivilege() {
	framework.RegisterFunction(has_table_privilege_name_text_text)
	framework.RegisterFunction(has_table_privilege_name_oid_text)
	framework.RegisterFunction(has_table_privilege_oid_text_text)
	framework.RegisterFunction(has_table_privilege_oid_oid_text)
	framework.RegisterFunction(has_table_privilege_text_text)
	framework.RegisterFunction(has_table_privilege_oid_text)
}

// has_table_privilege_name_text_text represents the PostgreSQL function of the same name, taking the same parameters.
var has_table_privilege_name_text_text = framework.Function3{
	Name:               "has_table_privilege",
	Return:             pgtypes.Bool,
	Parameters:         [3]*pgtypes.DoltgresType{pgtypes.Name, pgtypes.Text, pgtypes.Text},
	Strict:             true,
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, user, relation, privilege any) (any, error) {
		return hasRelationPrivilege(ctx, user, relation, privilege, tablePrivilegeTypes)
	},
}

// has_table_privilege_name_oid_text represents the PostgreSQL function of the same name, taking the same parameters.
var has_table_privilege_name_oid_text = framework.Function3{
	Name:               "has_table_privilege",
	Return:             pgtypes.Bool,
	Parameters:         [3]*pgtypes.DoltgresType{pgtypes.Name, pgtypes.Oid, pgtypes.Text},
	Strict:             true,
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, user, relation, privilege any) (any, error) {
		return hasRelationPrivilege(ctx, user, relation, privilege, tablePrivilegeTypes)
	},
}

// has_table_privilege_oid_text_text represents the PostgreSQL function of the same name, taking the same parameters.
var has_table_privilege_oid_text_text = framework.Function3{
	Name:               "has_table_privilege",
	Return:             pgtypes.Bool,
	Parameters:         [3]*pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Text, pgtypes.Text},
	Strict:             true,
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, user, relation, privilege any) (any, error) {
		return hasRelationPrivilege(ctx, user, relation, privilege, tablePrivilegeTypes)
	},
}

// has_table_privilege_oid_oid_text represents the PostgreSQL function of the same name, taking the same parameters.
var has_table_privilege_oid_oid_text = framework.Function3{
	Name:               "has_table_privilege",
	Return:             pgtypes.Bool,
	Parameters:         [3]*pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Oid, pgtypes.Text},
	Strict:             true,
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, user, relation, privilege any) (any, error) {
		return hasRelationPrivilege(ctx, user, relation, privilege, tablePrivilegeTypes)
	},
}

// has_table_privilege_text_text represents the PostgreSQL function of the same name, taking the same parameters.
var has_table_privilege_text_text = framework.Function2{
	Name:               "has_table_privilege",
	Return:             pgtypes.Bool,
	Parameters:         [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:             true,
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, relation, privilege any) (any, error) {
		return hasRelationPrivilege(ctx, nil, relation, privilege, tablePrivilegeTypes)
	},
}

// has_table_privilege_oid_text represents the PostgreSQL function of the same name, taking the same parameters.
var has_table_privilege_oid_text = framework.Function2{
	Name:               "has_table_privilege",
	Return:             pgtypes.Bool,
	Parameters:         [2]*pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Text},
	Strict:             true,
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, relation, privilege any) (any, error) {
		return hasRelationPrivilege(ctx, nil, relation, privilege, tablePrivilegeTypes)
	},
}

// tablePrivilegeTypes are the privilege names accepted by has_table_privilege.
var tablePrivilegeTypes = map[string]auth.Privilege{
	"SELECT":     auth.Privilege_SELECT,
	"INSERT":     auth.Privilege_INSERT,
	"UPDATE":     auth.Privilege_UPDATE,
	"DELETE":     auth.Privilege_DELETE,
	"TRUNCATE":   auth.Privilege_TRUNCATE,
	"REFERENCES": auth.Privilege_REFERENCES,
	"TRIGGER":    auth.Privilege_TRIGGER,
}

// relationKind is the kind of relation that a privilege inquiry targets.
type relationKind uint8

const (
	relationKind_Table relationKind = iota
	relationKind_Sequence
	relationKind_Index
)

// requestedPrivilege is a single entry of the privilege list given to a privilege inquiry function.
type requestedPrivilege struct {
	privilege       auth.Privilege
	withGrantOption bool
}

// hasRelationPrivilege implements has_table_privilege and has_any_column_privilege, using the same table ACLs as query
// execution. A nil user selects the effective current role, and SQL NULL arguments are handled by Strict. The
// privilege string is a comma-separated list, and the result is true if any of the listed privileges are held.
// Doltgres does not support column-level privileges, so any column privilege is held exactly when the table privilege
// is held.
func hasRelationPrivilege(ctx *sql.Context, user, target, privilege any, privilegeTypes map[string]auth.Privilege) (any, error) {
	var actor auth.Role
	var err error
	auth.LockRead(func() {
		actor, err = privilegeInquiryRole(ctx, user)
	})
	if err != nil {
		return nil, err
	}

	relation, byOID := target.(id.Id)
	if !byOID {
		name, err := framework.UnwrapString(ctx, target)
		if err != nil {
			return nil, err
		}
		resolved, err := pgtypes.Regclass.IoInput(ctx, name)
		if err != nil {
			if strings.Contains(err.Error(), "does not exist") {
				return nil, pgerror.Newf(pgcode.UndefinedTable, "relation %q does not exist", name)
			}
			return nil, err
		}
		relation = resolved.(id.Id)
	}

	requested, err := parseRelationPrivileges(ctx, privilege, privilegeTypes)
	if err != nil {
		return nil, err
	}

	kind, schemaName, relationName, exists, err := resolvePrivilegeRelation(ctx, relation)
	if err != nil {
		return nil, err
	}
	if !exists {
		// Like PostgreSQL, the OID overloads return NULL for missing relations, while the name overloads error.
		if byOID {
			return nil, nil
		}
		return nil, pgerror.Newf(pgcode.UndefinedTable, "relation with OID %d does not exist", id.Cache().ToOID(relation))
	}

	var result bool
	auth.LockRead(func() {
		for _, req := range requested {
			if relationPrivilegeHeld(actor, kind, schemaName, relationName, req) {
				result = true
				return
			}
		}
	})
	return result, nil
}

// parseRelationPrivileges parses the comma-separated privilege list of a privilege inquiry function. Every entry is
// validated, so that an invalid entry is an error even when an earlier entry is held.
func parseRelationPrivileges(ctx *sql.Context, privilege any, privilegeTypes map[string]auth.Privilege) ([]requestedPrivilege, error) {
	privileges, err := framework.UnwrapString(ctx, privilege)
	if err != nil {
		return nil, err
	}

	var requested []requestedPrivilege
	for _, item := range strings.Split(privileges, ",") {
		trimmed := strings.TrimSpace(item)
		name := strings.ToUpper(trimmed)
		withGrantOption := false
		if base, ok := strings.CutSuffix(name, " WITH GRANT OPTION"); ok {
			name = base
			withGrantOption = true
		}
		p, ok := privilegeTypes[name]
		if !ok {
			return nil, pgerror.Newf(pgcode.InvalidParameterValue, "unrecognized privilege type: %q", trimmed)
		}
		requested = append(requested, requestedPrivilege{
			privilege:       p,
			withGrantOption: withGrantOption,
		})
	}
	return requested, nil
}

// resolvePrivilegeRelation finds the relation with the given ID in the live catalog, since the OID cache can still
// contain IDs for dropped relations.
func resolvePrivilegeRelation(ctx *sql.Context, relation id.Id) (kind relationKind, schemaName, relationName string, exists bool, err error) {
	err = RunCallback(ctx, relation, Callbacks{
		Index: func(ctx *sql.Context, schema ItemSchema, table ItemTable, index ItemIndex) (bool, error) {
			kind, schemaName, relationName, exists = relationKind_Index, schema.Item.SchemaName(), index.Item.ID(), true
			return false, nil
		},
		Sequence: func(ctx *sql.Context, schema ItemSchema, sequence ItemSequence) (bool, error) {
			kind, schemaName, relationName, exists = relationKind_Sequence, schema.Item.SchemaName(), sequence.Item.Id.SequenceName(), true
			return false, nil
		},
		Table: func(ctx *sql.Context, schema ItemSchema, table ItemTable) (bool, error) {
			kind, schemaName, relationName, exists = relationKind_Table, schema.Item.SchemaName(), table.Item.Name(), true
			return false, nil
		},
		View: func(ctx *sql.Context, schema ItemSchema, view ItemView) (bool, error) {
			kind, schemaName, relationName, exists = relationKind_Table, schema.Item.SchemaName(), view.Item.Name, true
			return false, nil
		},
	})
	if sql.ErrTableNotFound.Is(err) {
		return kind, "", "", false, nil
	}
	return kind, schemaName, relationName, exists, err
}

// relationPrivilegeHeld returns whether the role holds the requested privilege on the relation. The caller must hold
// the auth lock.
func relationPrivilegeHeld(actor auth.Role, kind relationKind, schemaName, relationName string, req requestedPrivilege) bool {
	if actor.IsSuperUser {
		return true
	}

	switch kind {
	case relationKind_Table:
		if req.withGrantOption {
			return auth.RoleHasTablePrivilegeGrantOption(actor.ID(), schemaName, relationName, req.privilege)
		}
		return auth.RoleHasTablePrivilege(actor.ID(), schemaName, relationName, req.privilege)
	case relationKind_Sequence:
		// Of the table privileges, only SELECT and UPDATE apply to sequences.
		if req.privilege != auth.Privilege_SELECT && req.privilege != auth.Privilege_UPDATE {
			return false
		}
		if req.withGrantOption {
			return auth.RoleHasSequencePrivilegeGrantOption(actor.ID(), schemaName, relationName, req.privilege)
		}
		return auth.RoleHasSequencePrivilege(actor.ID(), schemaName, relationName, req.privilege)
	default:
		// Indexes do not have privileges of their own, and Doltgres does not track relation ownership, so only
		// superusers hold privileges on them.
		return false
	}
}
