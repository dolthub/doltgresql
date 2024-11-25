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

package dtables

import (
	"fmt"
	"strings"

	"github.com/dolthub/dolt/go/libraries/doltcore/rebase"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dprocedures"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// getRebaseSchema returns the schema for the rebase table.
func getRebaseSchema() sql.Schema {
	return []*sql.Column{
		{Name: "rebase_order", Type: pgtypes.Float32, Nullable: false, PrimaryKey: true}, // TODO: cannot have numeric key
		{Name: "action", Type: pgtypes.MustCreateNewVarCharType(6), Nullable: false},     // TODO: Should be enum(pick, squash, fixup, drop, reword)
		{Name: "commit_hash", Type: pgtypes.Text, Nullable: false},
		{Name: "commit_message", Type: pgtypes.Text, Nullable: false},
	}
}

// convertRebasePlanStepToRow converts a RebasePlanStep to a sql.Row.
func convertRebasePlanStepToRow(planMember rebase.RebasePlanStep) (sql.Row, error) {
	actionEnumValue := dprocedures.RebaseActionEnumType.IndexOf(strings.ToLower(planMember.Action))
	if actionEnumValue == -1 {
		return nil, fmt.Errorf("invalid rebase action: %s", planMember.Action)
	}

	return sql.Row{
		planMember.RebaseOrderAsFloat(),
		planMember.Action,
		planMember.CommitHash,
		planMember.CommitMsg,
	}, nil
}

// convertRowToRebasePlanStep converts a sql.Row to a RebasePlanStep.
func convertRowToRebasePlanStep(row sql.Row) (rebase.RebasePlanStep, error) {
	order, ok := row[0].(float32)
	if !ok {
		return rebase.RebasePlanStep{}, fmt.Errorf("invalid order value in rebase plan: %v (%T)", row[0], row[0])
	}

	rebaseAction, ok := row[1].(string)
	if !ok {
		return rebase.RebasePlanStep{}, fmt.Errorf("invalid enum value in rebase plan: %v (%T)", row[1], row[1])
	}

	rebaseIdx := dprocedures.RebaseActionEnumType.IndexOf(rebaseAction)
	if rebaseIdx < 0 {
		return rebase.RebasePlanStep{}, fmt.Errorf("invalid enum value in rebase plan: %v (%T)", row[1], row[1])
	}

	return rebase.RebasePlanStep{
		RebaseOrder: decimal.NewFromFloat32(order),
		Action:      rebaseAction,
		CommitHash:  row[2].(string),
		CommitMsg:   row[3].(string),
	}, nil
}

// getRebaseTableName returns the name of the rebase table.
func getRebaseTableName() string {
	return "rebase"
}
