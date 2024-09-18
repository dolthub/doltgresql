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
	"io"
	"strings"

	"github.com/dolthub/dolt/go/libraries/doltcore/table"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/server/types"
)

// CsvDataLoader is an implementation of DataLoader that reads data from chunks of CSV files and inserts them into a table.
type CsvDataLoader struct {
	results       LoadDataResults
	partialRecord string
	rowInserter   sql.RowInserter
	colTypes      []types.DoltgresType
	sch           sql.Schema
}

var _ DataLoader = (*CsvDataLoader)(nil)

// NewCsvDataLoader creates a new DataLoader instance that will insert records from chunks of CSV data into |table|.
func NewCsvDataLoader(ctx *sql.Context, table sql.InsertableTable) (*CsvDataLoader, error) {
	colTypes, err := getColumnTypes(table.Schema())
	if err != nil {
		return nil, err
	}

	rowInserter := table.Inserter(ctx)
	rowInserter.StatementBegin(ctx)

	return &CsvDataLoader{
		rowInserter: rowInserter,
		colTypes:    colTypes,
		sch:         table.Schema(),
	}, nil
}

// LoadChunk implements the DataLoader interface
func (tdl *CsvDataLoader) LoadChunk(ctx *sql.Context, data *bufio.Reader) error {
	combinedReader := newStringPrefixReader(tdl.partialRecord, data)
	tdl.partialRecord = ""

	reader, err := newCsvReader(combinedReader)
	if err != nil {
		return err
	}

	for {
		// Read the next record from the data
		record, err := reader.ReadSqlRow()
		if err != nil {
			if ple, ok := err.(*partialLineError); ok {
				tdl.partialRecord = ple.partialLine
				break
			}

			// csvReader will return a BadRow error if it encounters an input line without the
			// correct number of columns. If we see the end of data marker, then break out of the
			// loop and return from this function without returning an error.
			if _, ok := err.(*table.BadRow); ok {
				if len(record) == 1 && record[0] == "\\." {
					break
				}
			}

			if err != io.EOF {
				return err
			}

			recordValues := make([]string, 0, len(record))
			for _, v := range record {
				recordValues = append(recordValues, fmt.Sprintf("%v", v))
			}
			tdl.partialRecord = strings.Join(recordValues, ",")
			break
		}

		// If we see the end of data marker, then break out of the loop. Normally this will happen in the code
		// above when we receive a BadRow error, since there won't be enough values, but if a table only has
		// one column, we won't get a BadRow error, and we'll handle the end of data marker here.
		if len(record) == 1 && record[0] == "\\." {
			break
		}

		if len(record) > len(tdl.colTypes) {
			return fmt.Errorf("extra data after last expected column")
		} else if len(record) < len(tdl.colTypes) {
			return fmt.Errorf(`missing data for column "%s"`, tdl.sch[len(record)].Name)
		}

		// Cast the values using I/O input
		row := make(sql.Row, len(tdl.colTypes))
		for i := range tdl.colTypes {
			if record[i] == nil {
				row[i] = nil
			} else {
				row[i], err = tdl.colTypes[i].IoInput(ctx, fmt.Sprintf("%v", record[i]))
				if err != nil {
					return err
				}
			}
		}

		// Insert the row
		if err = tdl.rowInserter.Insert(ctx, row); err != nil {
			return err
		}
		tdl.results.RowsLoaded += 1
	}

	return nil
}

// Abort implements the DataLoader interface
func (tdl *CsvDataLoader) Abort(ctx *sql.Context) error {
	defer func() {
		if closeErr := tdl.rowInserter.Close(ctx); closeErr != nil {
			logrus.Warnf("error closing rowInserter: %v", closeErr)
		}
	}()

	return tdl.rowInserter.DiscardChanges(ctx, nil)
}

// Finish implements the DataLoader interface
func (tdl *CsvDataLoader) Finish(ctx *sql.Context) (*LoadDataResults, error) {
	defer func() {
		if closeErr := tdl.rowInserter.Close(ctx); closeErr != nil {
			logrus.Warnf("error closing rowInserter: %v", closeErr)
		}
	}()

	// If there is partial data from the last chunk that hasn't been inserted, return an error.
	if tdl.partialRecord != "" {
		return nil, fmt.Errorf("partial record (%s) found at end of data load", tdl.partialRecord)
	}

	err := tdl.rowInserter.StatementComplete(ctx)
	if err != nil {
		err = tdl.rowInserter.DiscardChanges(ctx, err)
		return nil, err
	}

	return &tdl.results, nil
}
