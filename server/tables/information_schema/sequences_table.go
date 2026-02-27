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

package information_schema

import (
	"github.com/dolthub/doltgresql/server/functions"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/information_schema"
	"strconv"
)

const SequencesTableName = "sequences"

// newSequencesTable returns a InformationSchemaTable for MySQL.
func newSequencesTable() *information_schema.InformationSchemaTable {
	return &information_schema.InformationSchemaTable{
		TableName:   SequencesTableName,
		TableSchema: sequencesSchema,
		Reader:      sequencesRowIter,
	}
}

// tablesSchema is the schema for the information_schema.TABLES table.
var sequencesSchema = sql.Schema{
	{Name: "sequence_catalog", Type: sql_identifier, Default: nil, Nullable: true, Source: SequencesTableName},
	{Name: "sequence_schema", Type: sql_identifier, Default: nil, Nullable: true, Source: SequencesTableName},
	{Name: "sequence_name", Type: sql_identifier, Default: nil, Nullable: true, Source: SequencesTableName},
	{Name: "data_type", Type: character_data, Default: nil, Nullable: true, Source: SequencesTableName},
	{Name: "numeric_precision", Type: cardinal_number, Default: nil, Nullable: true, Source: SequencesTableName},
	{Name: "numeric_precision_radix", Type: cardinal_number, Default: nil, Nullable: true, Source: SequencesTableName},
	{Name: "numeric_scale", Type: cardinal_number, Default: nil, Nullable: true, Source: SequencesTableName},
	{Name: "start_value", Type: character_data, Default: nil, Nullable: true, Source: SequencesTableName},
	{Name: "minimum_value", Type: character_data, Default: nil, Nullable: true, Source: SequencesTableName},
	{Name: "maximum_value", Type: character_data, Default: nil, Nullable: true, Source: SequencesTableName},
	{Name: "increment", Type: character_data, Default: nil, Nullable: true, Source: SequencesTableName},
	{Name: "cycle_option", Type: yes_or_no, Default: nil, Nullable: true, Source: SequencesTableName},
}

// sequencesRowIter implements the sql.RowIter for the information_schema.Sequences table.
func sequencesRowIter(ctx *sql.Context, catalog sql.Catalog) (sql.RowIter, error) {
	var rows []sql.Row

	err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
		Sequence: func(_ *sql.Context, schema functions.ItemSchema, sequence functions.ItemSequence) (cont bool, err error) {
			sequenceType := pgtypes.GetTypeByID(sequence.Item.DataTypeID)

			var precision, radix, scale interface{}
			if sequenceType != nil {
				precision, radix, scale = getColumnPrecisionAndScale(sequenceType)
			}

			cycle_option := "NO"
			if sequence.Item.Cycle {
				cycle_option = "YES"
			}

			rows = append(rows, sql.Row{
				schema.Item.Name(),              //sequence_catalog
				schema.Item.SchemaName(),        //sequence_schema
				sequence.Item.Id.SequenceName(), //sequence_name
				sequenceType.String(),           //data_type
				precision,                       //numeric_precision
				radix,                           //numeric_precision_radix
				scale,                           //numeric_scale
				strconv.FormatInt(sequence.Item.Start, 10),     //start_value
				strconv.FormatInt(sequence.Item.Minimum, 10),   //minimum_value
				strconv.FormatInt(sequence.Item.Maximum, 10),   //maximum_value
				strconv.FormatInt(sequence.Item.Increment, 10), //increment
				cycle_option, //cycle_option
			})
			return true, nil
		},
	})
	if err != nil {
		return nil, err
	}

	return sql.RowsToRowIter(rows...), nil
}
