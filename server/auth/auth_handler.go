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

package auth

import (
	"errors"
	"fmt"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
)

// AuthorizationQueryState contains any cached state for a query.
type AuthorizationQueryState struct {
	role   Role
	public Role
	err    error
}

var _ sql.AuthorizationQueryState = AuthorizationQueryState{}

// Error implements the sql.AuthorizationQueryState interface.
func (state AuthorizationQueryState) Error() error {
	return state.err
}

// AuthorizationQueryStateImpl implements the sql.AuthorizationQueryState interface.
func (state AuthorizationQueryState) AuthorizationQueryStateImpl() {}

// AuthorizationHandlerFactory is the factory for Doltgres.
type AuthorizationHandlerFactory struct{}

var _ sql.AuthorizationHandlerFactory = AuthorizationHandlerFactory{}

// CreateHandler implements the sql.AuthorizationHandlerFactory interface.
func (h AuthorizationHandlerFactory) CreateHandler(cat sql.Catalog) sql.AuthorizationHandler {
	return &AuthorizationHandler{
		cat: cat,
	}
}

// AuthorizationHandler handles vitess.AuthInformation for Doltgres.
type AuthorizationHandler struct {
	cat sql.Catalog
}

var _ sql.AuthorizationHandler = (*AuthorizationHandler)(nil)

// NewQueryState implements the sql.AuthorizationHandler interface.
func (h *AuthorizationHandler) NewQueryState(ctx *sql.Context) sql.AuthorizationQueryState {
	state := AuthorizationQueryState{}
	LockRead(func() {
		state.role = GetRole(ctx.Client().User)
		if !state.role.IsValid() {
			state.err = fmt.Errorf(`role "%s" does not exist`, state.role.Name)
			return
		}
		state.public = GetRole("public")
		if !state.public.IsValid() {
			state.err = fmt.Errorf(`role "%s" does not exist`, state.public.Name)
			return
		}
	})
	return state
}

// HandleAuth implements the sql.AuthorizationHandler interface.
func (h *AuthorizationHandler) HandleAuth(ctx *sql.Context, aqs sql.AuthorizationQueryState, auth vitess.AuthInformation) error {
	// TODO: eventually we'll want all conversion paths to provide both the AuthType and TargetType, but this lets us iterate faster for now
	if len(auth.AuthType) == 0 && len(auth.TargetType) == 0 {
		return nil
	}
	if aqs == nil {
		aqs = h.NewQueryState(ctx)
	}
	state := aqs.(AuthorizationQueryState)
	if state.err != nil {
		return state.err
	}
	globalLock.RLock()
	defer globalLock.RUnlock()

	var privileges []Privilege
	switch auth.AuthType {
	case AuthType_DELETE:
		privileges = []Privilege{Privilege_DELETE}
	case AuthType_INSERT:
		privileges = []Privilege{Privilege_INSERT}
	case AuthType_SELECT:
		privileges = []Privilege{Privilege_SELECT}
	case AuthType_TRUNCATE:
		privileges = []Privilege{Privilege_TRUNCATE}
	case AuthType_UPDATE:
		privileges = []Privilege{Privilege_UPDATE}
	default:
		if len(auth.AuthType) == 0 {
			return errors.New("AuthType is empty")
		} else {
			return fmt.Errorf("AuthType not handled: `%s`", auth.AuthType)
		}
	}

	// TODO: implement the rest of these
	switch auth.TargetType {
	case AuthTargetType_Ignore:
		// This means that the AuthType did not need a TargetType, so we can safely ignore it
	case AuthTargetType_SingleTableIdentifier:
		schemaName, err := core.GetSchemaName(ctx, nil, auth.TargetNames[0])
		if err != nil {
			return sql.ErrTableNotFound.New(auth.TargetNames[1])
		}
		ownerKey := OwnershipKey{
			PrivilegeObject: PrivilegeObject_TABLE,
			Schema:          schemaName,
			Name:            auth.TargetNames[1],
		}
		roleTableKey := TablePrivilegeKey{
			Role:  state.role.ID(),
			Table: doltdb.TableName{Name: auth.TargetNames[1], Schema: schemaName},
		}
		publicTableKey := TablePrivilegeKey{
			Role:  state.public.ID(),
			Table: doltdb.TableName{Name: auth.TargetNames[1], Schema: schemaName},
		}
		for _, privilege := range privileges {
			if !state.role.IsSuperUser && !IsOwner(ownerKey, state.role.ID()) &&
				!HasTablePrivilege(roleTableKey, privilege) && !HasTablePrivilege(publicTableKey, privilege) {
				return fmt.Errorf("permission denied for table %s", auth.TargetNames[1])
			}
		}
	case AuthTargetType_TODO:
		// This is similar to IGNORE, except we're meant to replace this at some point
	default:
		if len(auth.TargetType) == 0 {
			return errors.New("TargetType is unexpectedly empty")
		} else {
			return fmt.Errorf("TargetType not handled: `%s`", auth.TargetType)
		}
	}
	return nil
}

// HandleAuthNode implements the sql.AuthorizationHandler interface.
func (h *AuthorizationHandler) HandleAuthNode(ctx *sql.Context, aqs sql.AuthorizationQueryState, node sql.AuthorizationCheckerNode) error {
	if aqs == nil {
		aqs = h.NewQueryState(ctx)
	}
	state := aqs.(AuthorizationQueryState)
	if state.err != nil {
		return state.err
	}
	// TODO: implement this
	return nil
}

// CheckDatabase implements the sql.AuthorizationHandler interface.
func (h *AuthorizationHandler) CheckDatabase(ctx *sql.Context, aqs sql.AuthorizationQueryState, dbName string) error {
	if aqs == nil {
		aqs = h.NewQueryState(ctx)
	}
	state := aqs.(AuthorizationQueryState)
	if state.err != nil {
		return state.err
	}
	// TODO: implement this
	return nil
}

// CheckSchema implements the sql.AuthorizationHandler interface.
func (h *AuthorizationHandler) CheckSchema(ctx *sql.Context, aqs sql.AuthorizationQueryState, dbName string, schemaName string) error {
	if aqs == nil {
		aqs = h.NewQueryState(ctx)
	}
	state := aqs.(AuthorizationQueryState)
	if state.err != nil {
		return state.err
	}
	// TODO: implement this
	return nil
}

// CheckTable implements the sql.AuthorizationHandler interface.
func (h *AuthorizationHandler) CheckTable(ctx *sql.Context, aqs sql.AuthorizationQueryState, dbName string, schemaName string, tableName string) error {
	if aqs == nil {
		aqs = h.NewQueryState(ctx)
	}
	state := aqs.(AuthorizationQueryState)
	if state.err != nil {
		return state.err
	}
	// TODO: implement this
	return nil
}
