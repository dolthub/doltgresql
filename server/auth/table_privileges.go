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

import "github.com/dolthub/dolt/go/libraries/doltcore/doltdb"

// TablePrivileges contains the privileges given to a role on a table.
type TablePrivileges struct {
	Data map[TablePrivilegeKey]TablePrivilegeValue
}

// TODO: doc
type TablePrivilegeKey struct {
	Role  RoleID
	Table doltdb.TableName
}

// TODO: doc
type TablePrivilegeValue struct {
	Key        TablePrivilegeKey
	Privileges map[Privilege]GrantedPrivilege
}

// TODO: doc
func NewTablePrivileges() *TablePrivileges {
	return &TablePrivileges{make(map[TablePrivilegeKey]TablePrivilegeValue)}
}

// TODO: doc
func AddTablePrivilege(key TablePrivilegeKey, privilege GrantedPrivilege) {
	if existingValue, ok := globalMockDatabase.tablePrivileges.Data[key]; ok {
		existingValue.Privileges[privilege.Privilege] = privilege
	} else {
		globalMockDatabase.tablePrivileges.Data[key] = TablePrivilegeValue{
			Key:        key,
			Privileges: map[Privilege]GrantedPrivilege{privilege.Privilege: privilege},
		}
	}
}

// TODO: doc
func GetTablePrivilege(key TablePrivilegeKey) TablePrivilegeValue {
	privilege, _ := globalMockDatabase.tablePrivileges.Data[key]
	return privilege
}

// TODO: doc
func RemoveTablePrivilege(key TablePrivilegeKey) {
	delete(globalMockDatabase.tablePrivileges.Data, key)
}
