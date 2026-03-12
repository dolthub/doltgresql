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

package auth

import "github.com/dolthub/go-mysql-server/sql"

// PrivilegeSetLayer is used to allow some functions that inspect the GMS privilege set (such as branch control) to
// interface with Doltgres' auth system.
type PrivilegeSetLayer struct {
	Role RoleID
}

var _ sql.PrivilegeSet = (*PrivilegeSetLayer)(nil)

// NewPrivilegeSetLayer creates a new PrivilegeSetLayer for the user in the given context's session.
func NewPrivilegeSetLayer(ctx *sql.Context) *PrivilegeSetLayer {
	return &PrivilegeSetLayer{
		Role: GetRole(ctx.Client().User).id,
	}
}

// Has implements the interface sql.PrivilegeSet.
func (privSet *PrivilegeSetLayer) Has(privileges ...sql.PrivilegeType) bool {
	return IsSuperUser(privSet.Role)
}

// HasPrivileges implements the interface sql.PrivilegeSet.
func (privSet *PrivilegeSetLayer) HasPrivileges() bool {
	return IsSuperUser(privSet.Role)
}

// Count implements the interface sql.PrivilegeSet.
func (privSet *PrivilegeSetLayer) Count() int {
	if IsSuperUser(privSet.Role) {
		return 31 // The current number in GMS
	}
	return 0
}

// Database implements the interface sql.PrivilegeSet.
func (privSet *PrivilegeSetLayer) Database(dbName string) sql.PrivilegeSetDatabase {
	return &PrivilegeSetLayerDatabase{
		Db:   dbName,
		Role: privSet.Role,
	}
}

// GetDatabases implements the interface sql.PrivilegeSet.
func (privSet *PrivilegeSetLayer) GetDatabases() []sql.PrivilegeSetDatabase {
	return nil
}

// Equals implements the interface sql.PrivilegeSet.
func (privSet *PrivilegeSetLayer) Equals(otherPs sql.PrivilegeSet) bool {
	if other, ok := otherPs.(*PrivilegeSetLayer); ok {
		return privSet.Role == other.Role
	}
	return false
}

// ToSlice implements the interface sql.PrivilegeSet.
func (privSet *PrivilegeSetLayer) ToSlice() []sql.PrivilegeType {
	if IsSuperUser(privSet.Role) {
		return []sql.PrivilegeType{sql.PrivilegeType_Select,
			sql.PrivilegeType_Insert,
			sql.PrivilegeType_Update,
			sql.PrivilegeType_Delete,
			sql.PrivilegeType_Create,
			sql.PrivilegeType_Drop,
			sql.PrivilegeType_Reload,
			sql.PrivilegeType_Shutdown,
			sql.PrivilegeType_Process,
			sql.PrivilegeType_File,
			sql.PrivilegeType_GrantOption,
			sql.PrivilegeType_References,
			sql.PrivilegeType_Index,
			sql.PrivilegeType_Alter,
			sql.PrivilegeType_ShowDB,
			sql.PrivilegeType_Super,
			sql.PrivilegeType_CreateTempTable,
			sql.PrivilegeType_LockTables,
			sql.PrivilegeType_Execute,
			sql.PrivilegeType_ReplicationSlave,
			sql.PrivilegeType_ReplicationClient,
			sql.PrivilegeType_CreateView,
			sql.PrivilegeType_ShowView,
			sql.PrivilegeType_CreateRoutine,
			sql.PrivilegeType_AlterRoutine,
			sql.PrivilegeType_CreateUser,
			sql.PrivilegeType_Event,
			sql.PrivilegeType_Trigger,
			sql.PrivilegeType_CreateTablespace,
			sql.PrivilegeType_CreateRole,
			sql.PrivilegeType_DropRole}
	}
	return nil
}

// PrivilegeSetLayerDatabase is the database portion of PrivilegeSetLayer.
type PrivilegeSetLayerDatabase struct {
	Db   string
	Role RoleID
}

var _ sql.PrivilegeSetDatabase = (*PrivilegeSetLayerDatabase)(nil)

// Name implements the interface sql.PrivilegeSetDatabase.
func (privSet *PrivilegeSetLayerDatabase) Name() string {
	return privSet.Db
}

// Has implements the interface sql.PrivilegeSetDatabase.
func (privSet *PrivilegeSetLayerDatabase) Has(privileges ...sql.PrivilegeType) bool {
	return IsSuperUser(privSet.Role)
}

// HasPrivileges implements the interface sql.PrivilegeSetDatabase.
func (privSet *PrivilegeSetLayerDatabase) HasPrivileges() bool {
	return IsSuperUser(privSet.Role)
}

// Count implements the interface sql.PrivilegeSetDatabase.
func (privSet *PrivilegeSetLayerDatabase) Count() int {
	if IsSuperUser(privSet.Role) {
		return 31 // The current number in GMS
	}
	return 0
}

// Table implements the interface sql.PrivilegeSetDatabase.
func (privSet *PrivilegeSetLayerDatabase) Table(tblName string) sql.PrivilegeSetTable {
	panic("Table is not yet implemented for the Doltgres privilege layer")
}

// GetTables implements the interface sql.PrivilegeSetDatabase.
func (privSet *PrivilegeSetLayerDatabase) GetTables() []sql.PrivilegeSetTable {
	return nil
}

// Routine implements the interface sql.PrivilegeSetDatabase.
func (privSet *PrivilegeSetLayerDatabase) Routine(routineName string, isProcedure bool) sql.PrivilegeSetRoutine {
	panic("Routine is not yet implemented for the Doltgres privilege layer")
}

// GetRoutines implements the interface sql.PrivilegeSetDatabase.
func (privSet *PrivilegeSetLayerDatabase) GetRoutines() []sql.PrivilegeSetRoutine {
	return nil
}

// Equals implements the interface sql.PrivilegeSetDatabase.
func (privSet *PrivilegeSetLayerDatabase) Equals(otherPs sql.PrivilegeSetDatabase) bool {
	if other, ok := otherPs.(*PrivilegeSetLayerDatabase); ok {
		return privSet.Role == other.Role && privSet.Db == other.Db
	}
	return false
}

// ToSlice implements the interface sql.PrivilegeSetDatabase.
func (privSet *PrivilegeSetLayerDatabase) ToSlice() []sql.PrivilegeType {
	if IsSuperUser(privSet.Role) {
		return []sql.PrivilegeType{sql.PrivilegeType_Select,
			sql.PrivilegeType_Insert,
			sql.PrivilegeType_Update,
			sql.PrivilegeType_Delete,
			sql.PrivilegeType_Create,
			sql.PrivilegeType_Drop,
			sql.PrivilegeType_Reload,
			sql.PrivilegeType_Shutdown,
			sql.PrivilegeType_Process,
			sql.PrivilegeType_File,
			sql.PrivilegeType_GrantOption,
			sql.PrivilegeType_References,
			sql.PrivilegeType_Index,
			sql.PrivilegeType_Alter,
			sql.PrivilegeType_ShowDB,
			sql.PrivilegeType_Super,
			sql.PrivilegeType_CreateTempTable,
			sql.PrivilegeType_LockTables,
			sql.PrivilegeType_Execute,
			sql.PrivilegeType_ReplicationSlave,
			sql.PrivilegeType_ReplicationClient,
			sql.PrivilegeType_CreateView,
			sql.PrivilegeType_ShowView,
			sql.PrivilegeType_CreateRoutine,
			sql.PrivilegeType_AlterRoutine,
			sql.PrivilegeType_CreateUser,
			sql.PrivilegeType_Event,
			sql.PrivilegeType_Trigger,
			sql.PrivilegeType_CreateTablespace,
			sql.PrivilegeType_CreateRole,
			sql.PrivilegeType_DropRole}
	}
	return nil
}
