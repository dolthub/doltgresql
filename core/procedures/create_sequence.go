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
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/sequences"
)

// CreateSequenceName is the name of the stored procedure for creating sequences.
const CreateSequenceName = "doltgres_create_sequence"

// createSequence is the details for the stored procedure that creates sequences.
var createSequence = sql.ExternalStoredProcedureDetails{
	Name:      CreateSequenceName,
	Schema:    nil,
	Function:  createSequenceFunction,
	ReadOnly:  false,
	AdminOnly: false,
}

// createSequenceFunction is the stored procedure function for creating sequences.
func createSequenceFunction(ctx *sql.Context, ifNotExists bool, schema, name string, typeOID, min, max, start, increment int64, cycle bool, table, column string) (sql.RowIter, error) {
	if strings.HasPrefix(strings.ToLower(name), "dolt") {
		return nil, fmt.Errorf("sequences cannot be prefixed with 'dolt'")
	}
	if len(schema) == 0 {
		schema = core.GetCurrentSchema(ctx)
	}
	// TODO: The sequence name must be distinct from the name of any other relation (table, sequence, index, view, materialized view, or foreign table) in the same schema.
	err := core.UseRootInSession(ctx, func(ctx *sql.Context, root *core.RootValue) (*core.RootValue, error) {
		collection, err := root.GetSequences(ctx)
		if err != nil {
			return nil, err
		}
		if ifNotExists && collection.HasSequence(schema, name) {
			// TODO: issue a notice
			return nil, nil
		}
		if err = collection.CreateSequence(schema, &sequences.Sequence{
			Name:        name,
			DataTypeOID: uint32(typeOID),
			Persistence: sequences.Persistence_Permanent,
			Start:       start,
			Current:     start,
			Increment:   increment,
			Minimum:     min,
			Maximum:     max,
			Cache:       1,
			Cycle:       cycle,
			IsAtEnd:     false,
			OwnerUser:   "",
			OwnerTable:  table,
			OwnerColumn: column,
		}); err != nil {
			return nil, err
		}
		return root.PutSequences(ctx.Context, collection)
	})
	return nil, err
}
