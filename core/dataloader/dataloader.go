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

package dataloader

import (
	"bufio"
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/types"
)

// DataLoader allows callers to insert rows from multiple chunks into a table. Rows encoded in each chunk will not
// necessarily end cleanly on a chunk boundary, so DataLoader implementations must handle recognizing partial, or
// incomplete records, and saving that partial record until the next call to LoadChunk, so that it may be prefixed
// with the incomplete record.
type DataLoader interface {
	sql.ExecSourceRel
	
	// SetNextDataChunk sets the next data chunk to be processed by the DataLoader.  Data records
	// are not guaranteed to start and end cleanly on chunk boundaries, so implementations must recognize incomplete
	// records and save them to prepend on the next processed chunk.
	SetNextDataChunk(ctx *sql.Context, data *bufio.Reader) error

	// Finish finalizes the current load operation and cleans up any resources used. Implementations should check that 
	// the last call to LoadChunk did not end with an incomplete record and return an error to the caller if so. The
	// returned LoadDataResults describe the load operation, including how many rows were inserted.
	Finish(ctx *sql.Context) (*LoadDataResults, error)
}

// LoadDataResults contains the results of a load data operation, including the number of rows loaded.
type LoadDataResults struct {
	// RowsLoaded contains the total number of rows inserted during a load data operation.
	RowsLoaded int32
}

// getColumnTypes examines |sch| and returns a slice of DoltgresTypes in the order of the schema's columns. If any
// columns in the schema are not DoltgresType instances, an error is returned.
func getColumnTypes(colNames []string, sch sql.Schema) ([]*types.DoltgresType, error) {
	colTypes := make([]*types.DoltgresType, len(colNames))
	for i, colName := range colNames {
		colIdx := sch.IndexOfColName(colName)
		if colIdx < 0 {
			// should be impossible
			return nil, fmt.Errorf("column %s not found in schema", colName)
		}
		col := sch[colIdx]
		var ok bool
		colTypes[i], ok = col.Type.(*types.DoltgresType)
		if !ok {
			return nil, fmt.Errorf("unsupported column type: name: %s, type: %T", col.Name, col.Type)
		}
	}

	return colTypes, nil
}
