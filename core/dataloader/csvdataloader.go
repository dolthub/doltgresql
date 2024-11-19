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

	"github.com/dolthub/doltgresql/server/functions/framework"
	"github.com/dolthub/doltgresql/server/types"
)

// CsvDataLoader is an implementation of DataLoader that reads data from chunks of CSV files and inserts them into a table.
type CsvDataLoader struct {
	results       LoadDataResults
	partialRecord string
	rowInserter   sql.RowInserter
	colTypes      []types.DoltgresType
	sch           sql.Schema
	removeHeader  bool
	delimiter     string
}

var _ DataLoader = (*CsvDataLoader)(nil)

const defaultCsvDelimiter = ","

// NewCsvDataLoader creates a new DataLoader instance that will insert records from chunks of CSV data into |table|. If
// |header| is true, the first line of the data will be treated as a header and ignored. If |delimiter| is not the empty
// string, it will be used as the delimiter separating value.
func NewCsvDataLoader(ctx *sql.Context, table sql.InsertableTable, delimiter string, header bool) (*CsvDataLoader, error) {
	colTypes, err := getColumnTypes(table.Schema())
	if err != nil {
		return nil, err
	}

	rowInserter := table.Inserter(ctx)
	rowInserter.StatementBegin(ctx)

	if delimiter == "" {
		delimiter = defaultCsvDelimiter
	}

	return &CsvDataLoader{
		rowInserter:  rowInserter,
		colTypes:     colTypes,
		sch:          table.Schema(),
		removeHeader: header,
		delimiter:    delimiter,
	}, nil
}

// LoadChunk implements the DataLoader interface
func (cdl *CsvDataLoader) LoadChunk(ctx *sql.Context, data *bufio.Reader) error {
	combinedReader := NewStringPrefixReader(cdl.partialRecord, data)
	cdl.partialRecord = ""

	reader, err := newCsvReaderWithDelimiter(combinedReader, cdl.delimiter)
	if err != nil {
		return err
	}

	for {
		// Read the next record from the data
		if cdl.removeHeader {
			_, err := reader.readLine()
			cdl.removeHeader = false
			if err != nil {
				return err
			}
		}

		record, err := reader.ReadSqlRow()
		if err != nil {
			if ple, ok := err.(*partialLineError); ok {
				cdl.partialRecord = ple.partialLine
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
			cdl.partialRecord = strings.Join(recordValues, ",")
			break
		}

		// If we see the end of data marker, then break out of the loop. Normally this will happen in the code
		// above when we receive a BadRow error, since there won't be enough values, but if a table only has
		// one column, we won't get a BadRow error, and we'll handle the end of data marker here.
		if len(record) == 1 && record[0] == "\\." {
			break
		}

		if len(record) > len(cdl.colTypes) {
			return fmt.Errorf("extra data after last expected column")
		} else if len(record) < len(cdl.colTypes) {
			return fmt.Errorf(`missing data for column "%s"`, cdl.sch[len(record)].Name)
		}

		// Cast the values using I/O input
		row := make(sql.Row, len(cdl.colTypes))
		for i := range cdl.colTypes {
			if record[i] == nil {
				row[i] = nil
			} else {
				row[i], err = framework.IoInput(ctx, cdl.colTypes[i], fmt.Sprintf("%v", record[i]))
				if err != nil {
					return err
				}
			}
		}

		// Insert the row
		if err = cdl.rowInserter.Insert(ctx, row); err != nil {
			return err
		}
		cdl.results.RowsLoaded += 1
	}

	return nil
}

// Abort implements the DataLoader interface
func (cdl *CsvDataLoader) Abort(ctx *sql.Context) error {
	defer func() {
		if closeErr := cdl.rowInserter.Close(ctx); closeErr != nil {
			logrus.Warnf("error closing rowInserter: %v", closeErr)
		}
	}()

	return cdl.rowInserter.DiscardChanges(ctx, nil)
}

// Finish implements the DataLoader interface
func (cdl *CsvDataLoader) Finish(ctx *sql.Context) (*LoadDataResults, error) {
	defer func() {
		if closeErr := cdl.rowInserter.Close(ctx); closeErr != nil {
			logrus.Warnf("error closing rowInserter: %v", closeErr)
		}
	}()

	// If there is partial data from the last chunk that hasn't been inserted, return an error.
	if cdl.partialRecord != "" {
		return nil, fmt.Errorf("partial record (%s) found at end of data load", cdl.partialRecord)
	}

	err := cdl.rowInserter.StatementComplete(ctx)
	if err != nil {
		err = cdl.rowInserter.DiscardChanges(ctx, err)
		return nil, err
	}

	return &cdl.results, nil
}
