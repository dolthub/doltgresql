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
	"fmt"

	"github.com/dolthub/doltgresql/server/auth/rfc5802"
)

// ScramSha256Password is the struct form of an encrypted password.
type ScramSha256Password struct {
	Iterations uint32
	Salt       rfc5802.OctetString
	StoredKey  rfc5802.OctetString
	ServerKey  rfc5802.OctetString
}

// NewScramSha256Password creates a ScramSha256Password with a randomly-generated salt.
func NewScramSha256Password(rawPassword string) (*ScramSha256Password, error) {
	// This is unlikely to change, but it's defined here since we don't want to reference it elsewhere accidentally
	const iterations = 4096
	salt := GenerateRandomOctetString(16)
	saltedPassword, err := rfc5802.SaltedPassword(rawPassword, salt, iterations)
	if err != nil {
		return nil, err
	}
	storedKey := rfc5802.StoredKey(rfc5802.ClientKey(saltedPassword))
	serverKey := rfc5802.ServerKey(saltedPassword)
	return &ScramSha256Password{
		Iterations: iterations,
		Salt:       salt,
		StoredKey:  storedKey,
		ServerKey:  serverKey,
	}, nil
}

// AsPasswordString returns the password as defined in https://www.postgresql.org/docs/15/catalog-pg-authid.html
func (password ScramSha256Password) AsPasswordString() string {
	return fmt.Sprintf(`SCRAM-SHA-256$%d:%s$%s:%s`,
		password.Iterations, password.Salt.ToBase64(), password.StoredKey.ToBase64(), password.ServerKey.ToBase64())
}
