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

package procedures

import (
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
)

// DropSequenceName is the name of the stored procedure for dropping sequences.
const DropSequenceName = "doltgres_drop_sequence"

// dropSequence is the details for the stored procedure that drops sequences.
var dropSequence = sql.ExternalStoredProcedureDetails{
	Name:      DropSequenceName,
	Schema:    nil,
	Function:  dropSequenceFunction,
	ReadOnly:  false,
	AdminOnly: false,
}

// dropSequenceFunction is the stored procedure function for dropping sequences.
func dropSequenceFunction(ctx *sql.Context, ifExists bool, schema, name string, cascade bool) (sql.RowIter, error) {
	if len(schema) == 0 {
		schema = core.GetCurrentSchema(ctx)
	}
	err := core.UseRootInSession(ctx, func(ctx *sql.Context, root *core.RootValue) (*core.RootValue, error) {
		collection, err := root.GetSequences(ctx)
		if err != nil {
			return nil, err
		}
		if ifExists && !collection.HasSequence(schema, name) {
			// TODO: issue a notice
			return nil, nil
		}
		// TODO: handle cascade
		if err = collection.DropSequence(schema, name); err != nil {
			return nil, err
		}
		return root.PutSequences(ctx.Context, collection)
	})
	return nil, err
}
