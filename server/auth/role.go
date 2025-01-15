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

	"github.com/dolthub/doltgresql/utils"
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
	id                        RoleID
}

// RoleID represents a Role's ID. IDs are assigned during load and will be stable throughout the server's current
// process. IDs are useful for referencing a specific role without using their name, since names can change. This is
// basically a special OID specific to roles. Eventually, we'll have a proper OID system, but this is a placeholder for
// now.
// TODO: need to replace with id.InternalUser
type RoleID uint64

// CreateDefaultRole creates the given role object with all default values set.
func CreateDefaultRole(name string) Role {
	r := createDefaultRoleWithoutID(name)
	r.id = RoleID(userIDCounter.Add(1))
	return r
}

// createDefaultRoleWithoutID creates a default role, but does not assign an ID.
func createDefaultRoleWithoutID(name string) Role {
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
		id:                        RoleID(0),
	}
}

// ID returns this Role's ID value.
func (r *Role) ID() RoleID {
	return r.id
}

// IsValid returns true when the role has a valid ID.
func (r *Role) IsValid() bool {
	return r.id.IsValid()
}

// IsValid returns true when the RoleID has a valid value. It does not indicate that the RoleID is attached to a role
// that actually exists.
func (id RoleID) IsValid() bool {
	return id != RoleID(0)
}

// serialize writes the Role to the given writer.
func (r *Role) serialize(writer *utils.Writer) {
	// Version 0
	writer.String(r.Name)
	writer.Bool(r.IsSuperUser)
	writer.Bool(r.InheritPrivileges)
	writer.Bool(r.CanCreateRoles)
	writer.Bool(r.CanCreateDB)
	writer.Bool(r.CanLogin)
	writer.Bool(r.IsReplicationRole)
	writer.Bool(r.CanBypassRowLevelSecurity)
	writer.Int32(r.ConnectionLimit)
	if r.Password != nil {
		writer.Bool(true)
		writer.Uint32(r.Password.Iterations)
		writer.ByteSlice(r.Password.Salt)
		writer.ByteSlice(r.Password.StoredKey)
		writer.ByteSlice(r.Password.ServerKey)
	} else {
		writer.Bool(false)
	}
	if r.ValidUntil != nil {
		writer.Bool(true)
		writer.Int64(r.ValidUntil.UnixMicro())
	} else {
		writer.Bool(false)
	}
	writer.Uint64(uint64(r.id))
}

// deserialize reads the Role from the given reader.
func (r *Role) deserialize(version uint32, reader *utils.Reader) {
	switch version {
	case 0:
		r.Name = reader.String()
		r.IsSuperUser = reader.Bool()
		r.InheritPrivileges = reader.Bool()
		r.CanCreateRoles = reader.Bool()
		r.CanCreateDB = reader.Bool()
		r.CanLogin = reader.Bool()
		r.IsReplicationRole = reader.Bool()
		r.CanBypassRowLevelSecurity = reader.Bool()
		r.ConnectionLimit = reader.Int32()
		if reader.Bool() {
			r.Password = &ScramSha256Password{}
			r.Password.Iterations = reader.Uint32()
			r.Password.Salt = reader.ByteSlice()
			r.Password.StoredKey = reader.ByteSlice()
			r.Password.ServerKey = reader.ByteSlice()
		}
		if reader.Bool() {
			t := time.UnixMicro(reader.Int64())
			r.ValidUntil = &t
		}
		r.id = RoleID(reader.Uint64())
	default:
		panic("unexpected version in Role")
	}
}
