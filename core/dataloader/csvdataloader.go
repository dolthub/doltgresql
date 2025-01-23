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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/table"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/types"
)

// CsvDataLoader is an implementation of DataLoader that reads data from chunks of CSV files and inserts them into a table.
type CsvDataLoader struct {
	results       LoadDataResults
	partialRecord string
	nextDataChunk *bufio.Reader
	colTypes      []*types.DoltgresType
	sch           sql.Schema
	removeHeader  bool
	delimiter     string
}

func (cdl *CsvDataLoader) SetNextDataChunk(ctx *sql.Context, data *bufio.Reader) error {
	cdl.nextDataChunk = data
	return nil
}

var _ DataLoader = (*CsvDataLoader)(nil)

const defaultCsvDelimiter = ","

// NewCsvDataLoader creates a new DataLoader instance that will produce rows for the schema provided.
// |header| is true, the first line of the data will be treated as a header and ignored. If |delimiter| is not the empty
// string, it will be used as the delimiter separating value.
func NewCsvDataLoader(colNames []string, sch sql.Schema, delimiter string, header bool) (*CsvDataLoader, error) {
	colTypes, reducedSch, err := getColumnTypes(colNames, sch)
	if err != nil {
		return nil, err
	}

	if delimiter == "" {
		delimiter = defaultCsvDelimiter
	}

	return &CsvDataLoader{
		colTypes:     colTypes,
		sch:          reducedSch,
		removeHeader: header,
		delimiter:    delimiter,
	}, nil
}

// nextRow attempts to read the next row from the data and return it, and returns true if a row was read
func (cdl *CsvDataLoader) nextRow(ctx *sql.Context, reader *csvReader) (sql.Row, bool, error) {
	if cdl.removeHeader {
		_, err := reader.readLine()
		cdl.removeHeader = false
		if err != nil {
			return nil, false, err
		}
	}

	record, err := reader.ReadSqlRow()
	if err != nil {
		if ple, ok := err.(*partialLineError); ok {
			cdl.partialRecord = ple.partialLine
			return nil, false, nil
		}

		// csvReader will return a BadRow error if it encounters an input line without the
		// correct number of columns. If we see the end of data marker, then break out of the
		// loop and return from this function without returning an error.
		if _, ok := err.(*table.BadRow); ok {
			if len(record) == 1 && record[0] == "\\." {
				return nil, false, nil
			}
		}

		if err != io.EOF {
			return nil, false, err
		}

		recordValues := make([]string, 0, len(record))
		for _, v := range record {
			recordValues = append(recordValues, fmt.Sprintf("%v", v))
		}
		cdl.partialRecord = strings.Join(recordValues, ",")
		return nil, false, nil
	}

	// If we see the end of data marker, then break out of the loop. Normally this will happen in the code
	// above when we receive a BadRow error, since there won't be enough values, but if a table only has
	// one column, we won't get a BadRow error, and we'll handle the end of data marker here.
	if len(record) == 1 && record[0] == "\\." {
		return nil, false, nil
	}

	if len(record) > len(cdl.colTypes) {
		return nil, false, errors.Errorf("extra data after last expected column")
	} else if len(record) < len(cdl.colTypes) {
		return nil, false, errors.Errorf(`missing data for column "%s"`, cdl.sch[len(record)].Name)
	}

	// Cast the values using I/O input
	row := make(sql.Row, len(cdl.colTypes))
	for i := range cdl.colTypes {
		if record[i] == nil {
			row[i] = nil
		} else {
			row[i], err = cdl.colTypes[i].IoInput(ctx, fmt.Sprintf("%v", record[i]))
			if err != nil {
				return nil, false, err
			}
		}
	}

	return row, true, nil
}

// Finish implements the DataLoader interface
func (cdl *CsvDataLoader) Finish(ctx *sql.Context) (*LoadDataResults, error) {
	// If there is partial data from the last chunk that hasn't been inserted, return an error.
	if cdl.partialRecord != "" {
		return nil, errors.Errorf("partial record (%s) found at end of data load", cdl.partialRecord)
	}

	return &cdl.results, nil
}

func (cdl *CsvDataLoader) Resolved() bool {
	return true
}

func (cdl *CsvDataLoader) String() string {
	return "CsvDataLoader"
}

func (cdl *CsvDataLoader) Schema() sql.Schema {
	return cdl.sch
}

func (cdl *CsvDataLoader) Children() []sql.Node {
	return nil
}

func (cdl *CsvDataLoader) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(cdl, len(children), 0)
	}
	return cdl, nil
}

func (cdl *CsvDataLoader) IsReadOnly() bool {
	return true
}

type csvRowIter struct {
	cdl    *CsvDataLoader
	reader *csvReader
}

func (c csvRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	row, hasNext, err := c.cdl.nextRow(ctx, c.reader)
	if err != nil {
		return nil, err
	}

	// TODO: this isn't the best way to handle the count of rows, something like a RowUpdateAccumulator would be better
	if hasNext {
		c.cdl.results.RowsLoaded++
	} else {
		return nil, io.EOF
	}

	return row, nil
}

func (c csvRowIter) Close(context *sql.Context) error {
	return nil
}

var _ sql.RowIter = (*csvRowIter)(nil)

func (cdl *CsvDataLoader) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	combinedReader := NewStringPrefixReader(cdl.partialRecord, cdl.nextDataChunk)
	cdl.partialRecord = ""

	csvReader, err := newCsvReaderWithDelimiter(combinedReader, cdl.delimiter)
	if err != nil {
		return nil, err
	}

	return &csvRowIter{cdl: cdl, reader: csvReader}, nil
}
