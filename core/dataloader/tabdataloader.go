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

	"github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/go-mysql-server/sql"
)

const defaultTextDelimiter = "\t"
const defaultNullChar = "\\N"

// TabularDataLoader tracks the state of a load data operation from a tabular data source.
type TabularDataLoader struct {
	results       LoadDataResults
	partialLine   string
	nextDataChunk *bufio.Reader
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
func NewTabularDataLoader(colNames []string, sch sql.Schema, delimiterChar, nullChar string, header bool) (*TabularDataLoader, error) {
	colTypes, err := getColumnTypes(colNames, sch)
	if err != nil {
		return nil, err
	}
	
	if delimiterChar == "" {
		delimiterChar = defaultTextDelimiter
	}

	if nullChar == "" {
		nullChar = defaultNullChar
	}

	return &TabularDataLoader{
		colTypes:      colTypes,
		sch:           sch,
		delimiterChar: delimiterChar,
		nullChar:      nullChar,
		removeHeader:  header,
	}, nil
}

// nextRow returns the next SQL row from the reader provided, using any previously saved partial line. Returns true if
// there was another row.
func (tdl *TabularDataLoader) nextRow(ctx *sql.Context, data *bufio.Reader) (sql.Row, bool, error) {
	if tdl.removeHeader {
		_, err := data.ReadString('\n')
		tdl.removeHeader = false
		if err != nil {
			return nil, false, err
		}
	}

	for {
		// Read the next line from the file
		line, err := data.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return nil, false, err
			}

			// bufio.Reader.ReadString will return an error AND a line
			// if the final contents of the data does NOT end in the
			// delimiter. In this case, that means that we need to save
			// the partial line and use it in the next chunk.
			tdl.partialLine = line
			return nil, false, nil
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
			return nil, false, nil
		}

		// Skip over empty lines
		if len(line) == 0 {
			continue
		}
		
		// Split the values by the delimiter, ensuring the correct number of values have been read
		values := strings.Split(line, tdl.delimiterChar)
		if len(values) > len(tdl.colTypes) {
			return nil, false, fmt.Errorf("extra data after last expected column")
		} else if len(values) < len(tdl.colTypes) {
			return nil, false, fmt.Errorf(`missing data for column "%s"`, tdl.sch[len(values)].Name)
		}
		
		// Cast the values using I/O input
		row := make(sql.Row, len(tdl.colTypes))
		for i := range tdl.colTypes {
			if values[i] == tdl.nullChar {
				row[i] = nil
			} else {
				row[i], err = tdl.colTypes[i].IoInput(ctx, values[i])
				if err != nil {
					return nil, false, err
				}
			}
		}
		
		return row, true, nil
	}
}

func (tdl *TabularDataLoader) SetNextDataChunk(ctx *sql.Context, data *bufio.Reader) error {
	tdl.nextDataChunk = data
	return nil
}

// Finish completes the current load data operation and finalizes the data that has been inserted.
func (tdl *TabularDataLoader) Finish(ctx *sql.Context) (*LoadDataResults, error) {
	// If there is partial data from the last chunk that hasn't been inserted, return an error.
	if tdl.partialLine != "" {
		return nil, fmt.Errorf("partial line found at end of data load")
	}

	return &tdl.results, nil
}

func (tdl *TabularDataLoader) Resolved() bool {
	return true
}

func (tdl *TabularDataLoader) String() string {
	return "TabularDataLoader"
}

func (tdl *TabularDataLoader) Schema() sql.Schema {
	return nil
}

func (tdl *TabularDataLoader) Children() []sql.Node {
	return nil
}

func (tdl *TabularDataLoader) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(tdl, len(children), 0)
	}
	return tdl, nil
}

func (tdl *TabularDataLoader) IsReadOnly() bool {
	return true
}

type tabularRowIter struct {
	tdl *TabularDataLoader
	reader *bufio.Reader
}

func (t tabularRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	row, hasNext, err := t.tdl.nextRow(ctx, t.reader)
	if err != nil {
		return nil, err
	}

	if !hasNext {
		return nil, io.EOF
	}

	return row, nil
}

func (t tabularRowIter) Close(context *sql.Context) error {
	return nil
}

func (tdl *TabularDataLoader) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	return &tabularRowIter{tdl: tdl, reader: tdl.nextDataChunk}, nil
}