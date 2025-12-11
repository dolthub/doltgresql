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
	"os"

	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/go-mysql-server/sql"
)

// doltgresPasswordEnvVar is the name of the environment variable that can be used to set the password for the
// default user.
const doltgresPasswordEnvVar = "DOLTGRES_PASSWORD"

// doltgresUserEnvVar is the name of the environment variable that can be used to set the password for the
// default user.
const doltgresUserEnvVar = "DOLTGRES_USER"

// Init handles all initialization needs in this package.
func Init(dEnv *env.DoltEnv) {
	dbInit(dEnv)
	sql.SetAuthorizationHandlerFactory(AuthorizationHandlerFactory{})
}

// GetSuperUserAndPassword returns the superuser and password for the server to use, as defined in the environment
func GetSuperUserAndPassword() (string, string) {
	user := "postgres"
	if envUser := os.Getenv(doltgresUserEnvVar); envUser != "" {
		user = envUser
	}

	password := "password"
	if envPassword := os.Getenv(doltgresPasswordEnvVar); envPassword != "" {
		password = envPassword
	}

	return user, password
}
