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

package settings

import (
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/config"
)

// Set writes a PostgreSQL built-in or custom setting. Both SET statements and
// set_config use this path so validation and transaction scoping agree.
func Set(ctx *sql.Context, name string, value any, reset, local bool) error {
	name = strings.ToLower(name)
	if config.IsValidPostgresConfigParameter(name) {
		_ = core.SetDateStyleOutputFormat(ctx, "")
		if reset {
			var err error
			value, err = ctx.GetSessionVariableDefault(ctx, name)
			if err != nil {
				return err
			}
		}
		return config.SetPostgresParameter(ctx, name, value, local)
	}
	if config.IsValidCustomParameterName(name) {
		if reset {
			value = ""
		}
		return core.SetSetting(ctx, name, settingText(value), local)
	}
	if strings.Contains(name, ".") {
		return pgerror.Newf(pgcode.InvalidName, `invalid configuration parameter name "%s"`, name)
	}
	return pgerror.Newf(pgcode.UndefinedObject, `unrecognized configuration parameter "%s"`, name)
}

func settingText(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprint(value)
	}
}
