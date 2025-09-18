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
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initCurrentSetting registers the functions to the catalog.
func initCurrentSetting() {
	framework.RegisterFunction(current_setting_text)
	framework.RegisterFunction(current_setting_text_bool)
}

// current_setting_text represents the PostgreSQL function of the same name, taking the same parameters.
var current_setting_text = framework.Function1{
	Name:       "current_setting",
	Return:     pgtypes.Text, // TODO: it would be nice to support non-text values as well, but this is all postgres supports
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		return getCurSetting(ctx, val1.(string), false)
	},
}

// current_setting_text_bool represents the PostgreSQL function of the same name, taking the same parameters.
var current_setting_text_bool = framework.Function2{
	Name:       "current_setting",
	Return:     pgtypes.Text, // TODO: it would be nice to support non-text values as well, but this is all postgres supports
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Bool},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		return getCurSetting(ctx, val1.(string), val2.(bool))
	},
}

// getCurSetting returns value set for given user variable. It returns nil instead of an error
// if it doesn't exist and missingOk is set to true.
func getCurSetting(ctx *sql.Context, s string, missingOk bool) (any, error) {
	_, variable, err := ctx.GetUserVariable(ctx, s)
	if err != nil {
		if missingOk {
			return nil, nil
		}
		return nil, err
	}

	if variable != nil {
		return fmt.Sprintf("%v", variable), nil
	}

	variable, err = ctx.GetSessionVariable(ctx, s)
	if err != nil {
		if missingOk {
			return nil, nil
		}
		return nil, errors.Errorf(`unrecognized configuration parameter "%s"`, s)
	}

	if variable != nil {
		return fmt.Sprintf("%v", variable), nil
	}

	if missingOk {
		return nil, nil
	}
	return nil, errors.Errorf(`unrecognized configuration parameter "%s"`, s)
}
