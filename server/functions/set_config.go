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

package functions

import (
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

func initSetConfig() {
	framework.RegisterFunction(set_config_text_text_boolean)
}

// set_config_text_text_boolean implements the set_config() function
// https://www.postgresql.org/docs/current/functions-admin.html#FUNCTIONS-ADMIN-SET
var set_config_text_text_boolean = framework.Function3{
	Name:       "set_config",
	Return:     pgtypes.Text,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text, pgtypes.Bool},
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, settingName any, newValue any, isLocal any) (any, error) {
		if settingName == nil {
			return nil, errors.Errorf("NULL value not allowed for configuration setting name")
		}

		// NULL is not supported for configuration values, and gets turned into the empty string
		if newValue == nil {
			newValue = ""
		}

		if isLocal == true {
			// TODO: If isLocal is true, then the config setting should only persist for the current transaction
			return nil, errors.Errorf("setting configuration values for the current transaction is not supported yet")
		}

		// set_config can set system configuration or user configuration. System configuration settings are in top
		// level settings, while user configuration settings are namespaced.
		isUserConfig := strings.Contains(settingName.(string), ".")
		if isUserConfig {
			if err := ctx.SetUserVariable(ctx, settingName.(string), newValue.(string), pgtypes.Text); err != nil {
				return nil, err
			}
		} else {
			if err := ctx.SetSessionVariable(ctx, settingName.(string), newValue.(string)); err != nil {
				return nil, err
			}
		}

		return newValue.(string), nil
	},
}
