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

package users

import (
	"fmt"

	"github.com/dolthub/doltgresql/server/users/rfc5802"
)

// TODO: this will actually hold whatever needs to be held for proper authentication and role management.
//  For now though, this just exists to test that passwords for users are being stored and retrieved correctly.

var (
	globalMockDatabase MockDatabase
)

// MockDatabase is a temporary database to hold user passwords.
type MockDatabase struct {
	Users map[string]MockUser
}

// MockUser is a temporary user to hold a user's password.
type MockUser struct {
	Name     string
	Exists   bool
	Password ScramSha256Password
}

// ScramSha256Password is the struct form of an encrypted password.
type ScramSha256Password struct {
	Iterations uint32
	Salt       rfc5802.OctetString
	StoredKey  rfc5802.OctetString
	ServerKey  rfc5802.OctetString
}

// AsPasswordString returns the password as defined in https://www.postgresql.org/docs/15/catalog-pg-authid.html
func (password ScramSha256Password) AsPasswordString() string {
	return fmt.Sprintf(`SCRAM-SHA-256$%d:%s$%s:%s`,
		password.Iterations, password.Salt.ToBase64(), password.StoredKey.ToBase64(), password.ServerKey.ToBase64())
}

// GetUser returns the user with the given name.
func GetUser(name string) MockUser {
	return globalMockDatabase.Users[name]
}

// init initializes the mock database and fills it with default users for testing.
func init() {
	globalMockDatabase = MockDatabase{make(map[string]MockUser)}
	createUser("doltgres", "", "h2745oyhgwek4j")
	createUser("postgres", "password", "87835hg29u4has")
	createUser("user1", "abc123z", "g842hkaASF5320")
	createUser("user2", "bad_pass", "u924gf190yg4rb")
}

// createUser is called from within init to create a default set of users.
func createUser(name string, password string, salt string) {
	scramPassword := ScramSha256Password{
		Iterations: 4096,
		Salt:       rfc5802.OctetString(salt),
	}
	postgresSaltedPassword := rfc5802.SaltedPassword(password, scramPassword.Salt, scramPassword.Iterations)
	scramPassword.StoredKey = rfc5802.StoredKey(rfc5802.ClientKey(postgresSaltedPassword))
	scramPassword.ServerKey = rfc5802.ServerKey(postgresSaltedPassword)
	globalMockDatabase.Users[name] = MockUser{
		Name:     name,
		Exists:   true,
		Password: scramPassword,
	}
}
