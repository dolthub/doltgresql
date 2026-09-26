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

	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/config"
	"github.com/dolthub/doltgresql/server/functions/framework"
	"github.com/dolthub/doltgresql/server/settings"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

func initSetConfig() {
	framework.RegisterFunction(set_config_text_text_boolean)
}

// set_config_text_text_boolean implements the set_config() function
// https://www.postgresql.org/docs/current/functions-admin.html#FUNCTIONS-ADMIN-SET
var set_config_text_text_boolean = framework.Function3{
	Name:               "set_config",
	IsNonDeterministic: true,
	Return:             pgtypes.Text,
	Parameters:         [3]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text, pgtypes.Bool},
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, settingName any, newValue any, isLocal any) (any, error) {
		if settingName == nil {
			return nil, errors.Errorf("NULL value not allowed for configuration setting name")
		}

		reset := newValue == nil
		if reset {
			newValue = ""
		}

		settingNameStr, err := framework.UnwrapString(ctx, settingName)
		if err != nil {
			return nil, err
		}
		newValueStr, err := framework.UnwrapString(ctx, newValue)
		if err != nil {
			return nil, err
		}
		if strings.EqualFold(settingNameStr, "role") {
			if err := auth.ApplySetRole(ctx, newValueStr, strings.EqualFold(newValueStr, "none"), reset || strings.EqualFold(newValueStr, "default"), isLocal == true); err != nil {
				return nil, err
			}
			return auth.SelectedRoleSetting(ctx)
		}
		if strings.EqualFold(settingNameStr, "session_authorization") {
			if err := auth.ApplySessionAuthorization(ctx, newValueStr, reset || strings.EqualFold(newValueStr, "default"), isLocal == true); err != nil {
				return nil, err
			}
			return auth.SessionAuthorizationSetting(ctx)
		}

		if config.IsValidDoltConfigParameter(settingNameStr) && !config.IsValidPostgresConfigParameter(settingNameStr) {
			if isLocal == true {
				if err := ctx.Session.SetTransactionLocalVariable(ctx, settingNameStr, newValueStr); err != nil {
					return nil, err
				}
			} else if err := ctx.SetSessionVariable(ctx, settingNameStr, newValueStr); err != nil {
				return nil, err
			}
		} else {
			if err := settings.Set(ctx, settingNameStr, newValueStr, reset, isLocal == true); err != nil {
				return nil, err
			}
		}
		return getCurSetting(ctx, settingNameStr, false)
	},
}
