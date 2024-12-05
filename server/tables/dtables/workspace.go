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
	"github.com/dolthub/go-mysql-server/sql"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// getDoltWorkspaceBaseSqlSchema returns the base sql schema for the dolt_workspace_* table.
func getDoltWorkspaceBaseSqlSchema() sql.Schema {
	return []*sql.Column{
		{Name: "id", Type: pgtypes.Int64, PrimaryKey: true, Nullable: false},
		{Name: "staged", Type: pgtypes.Bool, Nullable: false},
		{Name: "diff_type", Type: pgtypes.Text, Nullable: false},
	}
}
