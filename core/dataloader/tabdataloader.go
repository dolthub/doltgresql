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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/server/types"
)

const defaultTextDelimiter = "\t"
const defaultNullChar = "\\N"

// TabularDataLoader tracks the state of a load data operation from a tabular data source.
type TabularDataLoader struct {
	results       LoadDataResults
	partialLine   string
	rowInserter   sql.RowInserter
	colTypes      []*types.DoltgresType
	sch           sql.Schema
	delimiterChar string
	nullChar      string
	removeHeader  bool
}

var _ DataLoader = (*TabularDataLoader)(nil)

// NewTabularDataLoader creates a new TabularDataLoader to insert into the specified |table| using the specified
// |delimiterChar| and |nullChar|. If |header| is true, the first line of the data will be treated as a header and
// ignored.
func NewTabularDataLoader(ctx *sql.Context, table sql.InsertableTable, delimiterChar, nullChar string, header bool) (*TabularDataLoader, error) {
	colTypes, err := getColumnTypes(table.Schema())
	if err != nil {
		return nil, err
	}

	rowInserter := table.Inserter(ctx)
	rowInserter.StatementBegin(ctx)

	if delimiterChar == "" {
		delimiterChar = defaultTextDelimiter
	}

	if nullChar == "" {
		nullChar = defaultNullChar
	}

	return &TabularDataLoader{
		rowInserter:   rowInserter,
		colTypes:      colTypes,
		sch:           table.Schema(),
		delimiterChar: delimiterChar,
		nullChar:      nullChar,
		removeHeader:  header,
	}, nil
}

// LoadChunk loads a chunk of data from the specified |data| reader into the table for this data loader. Note that
// the chunk does not need to end on a line boundary – the loader will handle partial lines at the end of the chunk
// by saving them for the next chunk.
func (tdl *TabularDataLoader) LoadChunk(ctx *sql.Context, data *bufio.Reader) error {
	for {
		row := make(sql.Row, len(tdl.colTypes))

		// Read the next line from the file
		line, err := data.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return err
			}

			// bufio.Reader.ReadString will return an error AND a line
			// if the final contents of the data does NOT end in the
			// delimiter. In this case, that means that we need to save
			// the partial line and use it in the next chunk.
			tdl.partialLine = line
			break
		}

		if tdl.removeHeader {
			// Skip the first line if it is a header
			tdl.removeHeader = false
			continue
		}

		// If we've not reached EOF, then there will be a newline appended to the end that we must remove.
		line = strings.TrimSuffix(line, "\n")
		// Data with windows line endings will also have a carriage return character that we need to remove.
		line = strings.TrimSuffix(line, "\r")

		if tdl.partialLine != "" {
			line = tdl.partialLine + line
			tdl.partialLine = ""
		}

		// If we see the end of data marker, return early
		if line == "\\." {
			return nil
		}

		// Skip over empty lines
		if len(line) == 0 {
			continue
		}
		// Split the values by the delimiter, ensuring the correct number of values have been read
		values := strings.Split(line, tdl.delimiterChar)
		if len(values) > len(tdl.colTypes) {
			return fmt.Errorf("extra data after last expected column")
		} else if len(values) < len(tdl.colTypes) {
			return fmt.Errorf(`missing data for column "%s"`, tdl.sch[len(values)].Name)
		}
		// Cast the values using I/O input
		for i := range tdl.colTypes {
			if values[i] == tdl.nullChar {
				row[i] = nil
			} else {
				row[i], err = tdl.colTypes[i].IoInput(ctx, values[i])
				if err != nil {
					return err
				}
			}
		}
		// Insert the read row
		err = tdl.rowInserter.Insert(ctx, row)
		if err != nil {
			return err
		}
		tdl.results.RowsLoaded += 1
	}

	return nil
}

// Abort ends the current load data operation and discards any changes that have been made.
func (tdl *TabularDataLoader) Abort(ctx *sql.Context) error {
	defer func() {
		if closeErr := tdl.rowInserter.Close(ctx); closeErr != nil {
			logrus.Warnf("error closing rowInserter: %v", closeErr)
		}
	}()

	return tdl.rowInserter.DiscardChanges(ctx, nil)
}

// Finish completes the current load data operation and finalizes the data that has been inserted.
func (tdl *TabularDataLoader) Finish(ctx *sql.Context) (*LoadDataResults, error) {
	defer func() {
		if closeErr := tdl.rowInserter.Close(ctx); closeErr != nil {
			logrus.Warnf("error closing rowInserter: %v", closeErr)
		}
	}()

	// If there is partial data from the last chunk that hasn't been inserted, return an error.
	if tdl.partialLine != "" {
		return nil, fmt.Errorf("partial line found at end of data load")
	}

	err := tdl.rowInserter.StatementComplete(ctx)
	if err != nil {
		err = tdl.rowInserter.DiscardChanges(ctx, err)
		return nil, err
	}

	return &tdl.results, nil
}

func (tdl *TabularDataLoader) Resolved() bool {
	// TODO implement me
	panic("implement me")
}

func (tdl *TabularDataLoader) String() string {
	// TODO implement me
	panic("implement me")
}

func (tdl *TabularDataLoader) Schema() sql.Schema {
	// TODO implement me
	panic("implement me")
}

func (tdl *TabularDataLoader) Children() []sql.Node {
	// TODO implement me
	panic("implement me")
}

func (tdl *TabularDataLoader) WithChildren(children ...sql.Node) (sql.Node, error) {
	// TODO implement me
	panic("implement me")
}

func (tdl *TabularDataLoader) IsReadOnly() bool {
	// TODO implement me
	panic("implement me")
}

func (tdl *TabularDataLoader) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	// TODO implement me
	panic("implement me")
}
