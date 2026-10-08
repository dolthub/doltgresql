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

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPgGetTriggerDef registers the functions to the catalog.
func initPgGetTriggerDef() {
	framework.RegisterFunction(pg_get_triggerdef_oid)
	framework.RegisterFunction(pg_get_triggerdef_oid_bool)
}

// pg_get_triggerdef_oid represents the PostgreSQL system catalog information function taking 1 parameter.
var pg_get_triggerdef_oid = framework.Function1{
	Name:               "pg_get_triggerdef",
	Return:             pgtypes.Text,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Oid},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return getTriggerDef(ctx, val.(id.Id))
	},
}

// pg_get_triggerdef_oid_bool represents the PostgreSQL system catalog information function taking 2 parameters.
var pg_get_triggerdef_oid_bool = framework.Function2{
	Name:               "pg_get_triggerdef",
	Return:             pgtypes.Text,
	Parameters:         [2]*pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Bool},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		// The pretty flag only affects the formatting of expressions, which we don't reproduce, so
		// we return the same text either way.
		return getTriggerDef(ctx, val1.(id.Id))
	},
}

// getTriggerDef returns the definition of the trigger for the given OID, or an empty string if the OID doesn't refer
// to an existing trigger. Postgres reconstructs the statement from the catalog, whereas we return the statement that
// created the trigger.
func getTriggerDef(ctx *sql.Context, oidVal id.Id) (string, error) {
	if oidVal.Section() != id.Section_Trigger {
		return "", nil
	}
	collection, err := core.GetTriggersCollectionFromContext(ctx, ctx.GetCurrentDatabase())
	if err != nil {
		return "", err
	}
	trigger, err := collection.GetTrigger(ctx, id.Trigger(oidVal))
	if err != nil {
		return "", err
	}
	if !trigger.ID.IsValid() {
		return "", nil
	}
	return strings.TrimSuffix(strings.TrimSpace(trigger.Definition), ";"), nil
}
