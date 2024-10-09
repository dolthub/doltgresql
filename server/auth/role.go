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
	"time"
)

// Role represents a role/user.
type Role struct {
	Name                      string               // rolname
	IsSuperUser               bool                 // rolsuper
	InheritPrivileges         bool                 // rolinherit
	CanCreateRoles            bool                 // rolcreaterole
	CanCreateDB               bool                 // rolcreatedb
	CanLogin                  bool                 // rolcanlogin
	IsReplicationRole         bool                 // rolreplication
	CanBypassRowLevelSecurity bool                 // rolbypassrls
	ConnectionLimit           int32                // rolconnlimit
	Password                  *ScramSha256Password // rolpassword
	ValidUntil                *time.Time           // rolvaliduntil
}

// CreateDefaultRole creates the given role object with all default values set.
func CreateDefaultRole(name string) Role {
	return Role{
		Name:                      name,
		IsSuperUser:               false,
		InheritPrivileges:         true,
		CanCreateRoles:            false,
		CanCreateDB:               false,
		CanLogin:                  false,
		IsReplicationRole:         false,
		CanBypassRowLevelSecurity: false,
		ConnectionLimit:           -1,
		Password:                  nil,
		ValidUntil:                nil,
	}
}
