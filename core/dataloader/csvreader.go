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
	"bytes"
	"context"
	"encoding/csv"
	"io"
	"unicode/utf8"

	"github.com/dolthub/dolt/go/libraries/doltcore/table"
	"github.com/dolthub/go-mysql-server/sql"
	textunicode "golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// csvReadBufSize is the size of the buffer used when reading the csv file.
var csvReadBufSize = 256 * 1024

// partialLineError is an error type that is returned when an incomplete record is read from a CSV
// file. This can occur when a CSV document is split across multiple messages and the message
// boundaries don't line up with CSV record boundaries. Callers should use this error to record the
// partial line, so that it can be prepended to the next message.
type partialLineError struct {
	partialLine string
}

var _ error = partialLineError{}

func (ple partialLineError) Error() string {
	return "incomplete record found at end of CSV data: " + ple.partialLine
}

// csvReader implements TableReader.  It reads csv files and returns rows.
//
// This implementation is adapted from the CSVReader in dolt, which is a fork
// of the standard Golang CSV reader.  The main differences with the Golang std
// library implementation are that this parser has been adapted to differentiate
// between quoted and unquoted empty strings (for distinguishing between the empty
// string and NULL), and to use multi-rune delimiters. This adaptation removes the
// comment feature and the lazyQuotes option.
//
// Additionally, this fork of the dolt implementation removes some dolt specific
// features and adds support for a few Postgres requirements, such as allowing for
// the full CSV document to be arbitrarily split into multiple messages and for
// incomplete/partial lines to be communicated to the caller.
type csvReader struct {
	closer          io.Closer
	bRd             *bufio.Reader
	isDone          bool
	delim           []byte
	numLine         int
	fieldsPerRecord int
}

// NewCsvReader creates a csvReader from a given ReadCloser.
//
// The interpretation of the bytes of the supplied reader is a little murky. If
// there is a UTF8, UTF16LE or UTF16BE BOM as the first bytes read, then the
// BOM is stripped and the remaining contents of the reader are treated as that
// encoding. If we are not in any of those marked encodings, then some of the
// bytes go uninterpreted until we get to the SQL layer. It is currently the
// case that newlines must be encoded as a '0xa' byte.
func NewCsvReader(r io.ReadCloser) (*csvReader, error) {
	return newCsvReaderWithDelimiter(r, ",")
}

// newCsvReaderWithDelimiter creates a csvReader from a given ReadCloser, |r|, using
// the |delimiter| as the field delimiter in the parsed data.
func newCsvReaderWithDelimiter(r io.ReadCloser, delimiter string) (*csvReader, error) {
	textReader := transform.NewReader(r, textunicode.BOMOverride(transform.Nop))
	br := bufio.NewReaderSize(textReader, csvReadBufSize)

	return &csvReader{
		closer: r,
		bRd:    br,
		isDone: false,
		delim:  []byte(delimiter),
	}, nil
}

func (csvr *csvReader) ReadSqlRow() (sql.Row, error) {
	if csvr.isDone {
		return nil, io.EOF
	}

	rowVals, err := csvr.csvReadRecords(nil)
	if err == io.EOF {
		csvr.isDone = true
		return nil, io.EOF
	}

	sqlRows := rowValsToSQLRows(rowVals)
	if err != nil {
		if _, ok := err.(*partialLineError); ok {
			return nil, err
		}
		return sqlRows, table.NewBadRow(nil, err.Error())
	}

	return sqlRows, nil
}

func rowValsToSQLRows(rowVals []*string) sql.Row {
	var sqlRow sql.Row
	for _, rowVal := range rowVals {
		if rowVal == nil {
			sqlRow = append(sqlRow, nil)
		} else {
			sqlRow = append(sqlRow, *rowVal)
		}
	}

	return sqlRow
}

// Close should release resources being held
func (csvr *csvReader) Close(ctx context.Context) error {
	if csvr.closer != nil {
		err := csvr.closer.Close()
		csvr.closer = nil

		return err
	}
	return nil
}

// Functions below this line are borrowed or adapted from encoding/csv/reader.go

// lengthNL returns 1 if the last byte in b is a newline, 0 otherwise.
func lengthNL(b []byte) int {
	if len(b) > 0 && b[len(b)-1] == '\n' {
		return 1
	}
	return 0
}

// readLine reads the next line (with the trailing endline).
// If EOF is hit without a trailing endline, it will be omitted.
// If some bytes were read, then the error is never io.EOF.
// The result is only valid until the next call to readLine.
func (csvr *csvReader) readLine() ([]byte, error) {
	var rawBuffer []byte

	line, err := csvr.bRd.ReadSlice('\n')
	if err == bufio.ErrBufferFull {
		rawBuffer = append(rawBuffer[:0], line...)
		for err == bufio.ErrBufferFull {
			line, err = csvr.bRd.ReadSlice('\n')
			rawBuffer = append(rawBuffer, line...)
		}
		line = rawBuffer
	}
	if len(line) > 0 && err == io.EOF {
		err = nil
		// For backwards compatibility, drop trailing \r before EOF.
		if line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
	}
	csvr.numLine++
	// Normalize \r\n to \n on all input lines.
	if n := len(line); n >= 2 && line[n-2] == '\r' && line[n-1] == '\n' {
		line[n-2] = '\n'
		line = line[:n-1]
	}

	// If the line does NOT end with a newline, then we must have read a partial record
	if len(line) > 0 && lengthNL(line) == 0 {
		return nil, &partialLineError{string(line)}
	}

	return line, err
}

type recordState struct {
	line []byte
	// recordBuffer holds the unescaped fields, one after another.
	// The fields can be accessed by using the indexes in fieldIndexes.
	// E.g., For the row `a,"b","c""d",e`, recordBuffer will contain `abc"de`
	// and fieldIndexes will contain the indexes [1, 2, 5, 6].
	recordBuffer []byte
	fieldIndexes []int
	rawData      []byte
}

func (csvr *csvReader) csvReadRecords(dst []*string) ([]*string, error) {
	recordStartline := csvr.numLine // Starting line for record

	var rs recordState
	var err error
	for err == nil {
		rs = recordState{}
		rs.line, err = csvr.readLine()
		rs.rawData = append(rs.rawData, rs.line...)

		if err == nil && len(rs.line) == lengthNL(rs.line) {
			continue // Skip empty lines
		}
		break
	}
	if err != nil {
		return nil, err
	}

	// nullString indicates whether to interpret an empty string as a NULL
	// only empty strings escaped with double quotes will be non-null
	nullString := make(map[int]bool)
	fieldIdx := 0

	kontinue := true
	for kontinue {
		// Parse each field in the record.
		keep := true
		if len(rs.line) == 0 || rs.line[0] != '"' {
			kontinue, keep, err = csvr.parseField(&rs)
			if !keep {
				nullString[fieldIdx] = true
			}
		} else {
			kontinue, err = csvr.parseQuotedField(&rs)
			if err != nil {
				return nil, err
			}
		}
		fieldIdx++
	}

	// Create a single string and create slices out of it.
	// This pins the memory of the fields together, but allocates once.
	str := string(rs.recordBuffer) // Convert to string once to batch allocations
	dst = dst[:0]
	if cap(dst) < len(rs.fieldIndexes) {
		dst = make([]*string, len(rs.fieldIndexes))
	}
	dst = dst[:len(rs.fieldIndexes)]
	var preIdx int
	for i, idx := range rs.fieldIndexes {
		_, ok := nullString[i]
		if ok {
			dst[i] = nil
		} else {
			s := str[preIdx:idx]
			dst[i] = &s
		}
		preIdx = idx
	}

	// Check or update the expected fields per record.
	if csvr.fieldsPerRecord > 0 {
		if len(dst) != csvr.fieldsPerRecord && err == nil {
			err = &csv.ParseError{StartLine: recordStartline, Line: csvr.numLine, Err: csv.ErrFieldCount}
		}
	} else if csvr.fieldsPerRecord == 0 {
		csvr.fieldsPerRecord = len(dst)
	}

	return dst, err
}

func (csvr *csvReader) parseField(rs *recordState) (kontinue bool, keep bool, err error) {
	i := bytes.Index(rs.line, csvr.delim)
	field := rs.line
	if i >= 0 {
		field = field[:i]
	} else {
		field = field[:len(field)-lengthNL(field)]
	}
	rs.recordBuffer = append(rs.recordBuffer, field...)
	rs.fieldIndexes = append(rs.fieldIndexes, len(rs.recordBuffer))
	keep = len(field) != 0 // discard unquoted empty strings
	if i >= 0 {
		dl := len(csvr.delim)
		rs.line = rs.line[i+dl:]
		return true, keep, err
	}
	return false, keep, err
}

func (csvr *csvReader) parseQuotedField(rs *recordState) (kontinue bool, err error) {
	const quoteLen = len(`"`)
	dl := len(csvr.delim)
	recordStartLine := csvr.numLine
	fullField := rs.line

	// Quoted string field
	rs.line = rs.line[quoteLen:]
	for {
		i := bytes.IndexByte(rs.line, '"')
		if i >= 0 {
			// Hit next quote.
			rs.recordBuffer = append(rs.recordBuffer, rs.line[:i]...)
			rs.line = rs.line[i+quoteLen:]

			atDelimiter := len(rs.line) >= dl && bytes.Equal(rs.line[:dl], csvr.delim)
			nextRune, _ := utf8.DecodeRune(rs.line)

			switch {
			case atDelimiter:
				// `"<delimiter>` sequence (end of field).
				rs.line = rs.line[dl:]
				rs.fieldIndexes = append(rs.fieldIndexes, len(rs.recordBuffer))
				return true, err
			case nextRune == '"':
				// `""` sequence (append quote).
				rs.recordBuffer = append(rs.recordBuffer, '"')
				rs.line = rs.line[quoteLen:]
			case lengthNL(rs.line) == len(rs.line):
				// `"\n` sequence (end of line).
				rs.fieldIndexes = append(rs.fieldIndexes, len(rs.recordBuffer))
				return false, err
			default:
				// `"*` sequence (invalid non-escaped quote).
				col := utf8.RuneCount(fullField[:len(fullField)-len(rs.line)-quoteLen])
				err = &csv.ParseError{StartLine: recordStartLine, Line: csvr.numLine, Column: col, Err: csv.ErrQuote}
				return false, err
			}
		} else if len(rs.line) > 0 {
			// Hit end of line (copy all data so far).
			rs.recordBuffer = append(rs.recordBuffer, rs.line...)
			if err != nil {
				return false, err
			}

			rs.line, err = csvr.readLine()
			rs.rawData = append(rs.rawData, rs.line...)
			if err == io.EOF {
				err = nil
			}
			// If we get a partialLineError, populate the partialLine field with the full record data
			// since quoted fields can span multiple lines, otherwise we wouldn't capture the initial
			// lines of this record.
			if ple, ok := err.(*partialLineError); ok {
				ple.partialLine = string(rs.rawData)
				return true, ple
			}
			fullField = append(fullField, rs.line...)
		} else {
			// Abrupt end of file
			if err == nil {
				return false, &partialLineError{string(rs.rawData)}
			}
			rs.fieldIndexes = append(rs.fieldIndexes, len(rs.recordBuffer))
			return false, err
		}
	}
}
