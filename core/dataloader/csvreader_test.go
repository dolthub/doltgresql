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
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// basicCsvData is used for a smoke test with CSV parsing
const basicCsvData = `
1,foo,bar
2,  ,bash
`

// wrongNumberOfFieldsCsvData tests the case where records have different
// numbers of values in them.
const wrongNumberOfFieldsCsvData = `
1,foo,bar,baz,bash
2,blue
3,boop,beep,bop,blorp
`

// partialLineErrorCsvData tests the case where the last line of the CSV data
// does not end with a newline character.
const partialLineErrorCsvData = `
1,foo,bar,baz,bash
2,boop,beep,bop,blorp
3,blue,g`

// nullAndEmptyStringQuotingCsvData tests the difference between representing
// NULL and an empty string.
const nullAndEmptyStringQuotingCsvData = `
1,,NULL,"NULL",""
`

// escapedQuotesCsvData tests escaped quotes in CSV data.
const escapedQuotesCsvData = `
1,'',"'","""",','''
`

// newLineInQuotedFieldCsvData tests when a quoted field contains a newline.
const newLineInQuotedFieldCsvData = `
1,foo,"baz
bar
bash"
`

// endOfDataMarkerCsvData tests when a quoted field contains the end of data marker.
const endOfDataMarkerCsvData = `
1,foo,"baz
\.
bash"
`

// TestCsvReader tests various cases of CSV data parsing.
func TestCsvReader(t *testing.T) {
	t.Run("basic CSV data", func(t *testing.T) {
		csvReader, err := newCsvReader(newReader(basicCsvData))
		require.NoError(t, err)

		// Read the first row
		row, err := csvReader.ReadSqlRow()
		require.NoError(t, err)
		assert.Equal(t, "1", row[0])
		assert.Equal(t, "foo", row[1])
		assert.Equal(t, "bar", row[2])

		// Read the second row
		row, err = csvReader.ReadSqlRow()
		require.NoError(t, err)
		assert.Equal(t, "2", row[0])
		assert.Equal(t, "  ", row[1])
		assert.Equal(t, "bash", row[2])

		// Read the EOF error
		_, err = csvReader.ReadSqlRow()
		require.Equal(t, io.EOF, err)
	})

	t.Run("wrong number of fields", func(t *testing.T) {
		csvReader, err := newCsvReader(newReader(wrongNumberOfFieldsCsvData))
		require.NoError(t, err)

		// Read the first row
		row, err := csvReader.ReadSqlRow()
		require.NoError(t, err)
		require.Equal(t, "1", row[0])
		require.Equal(t, "foo", row[1])
		require.Equal(t, "bar", row[2])
		require.Equal(t, "baz", row[3])
		require.Equal(t, "bash", row[4])

		// Read the second row
		_, err = csvReader.ReadSqlRow()
		require.Error(t, err)
		require.Equal(t, "record on line 3: wrong number of fields", err.Error())
	})

	t.Run("incomplete line, no newline ending", func(t *testing.T) {
		csvReader, err := newCsvReader(newReader(partialLineErrorCsvData))
		require.NoError(t, err)

		// Read the first row
		row, err := csvReader.ReadSqlRow()
		require.NoError(t, err)
		assert.Equal(t, "1", row[0])
		assert.Equal(t, "foo", row[1])
		assert.Equal(t, "bar", row[2])
		assert.Equal(t, "baz", row[3])
		assert.Equal(t, "bash", row[4])

		// Read the second row
		row, err = csvReader.ReadSqlRow()
		require.NoError(t, err)
		assert.Equal(t, "2", row[0])
		assert.Equal(t, "boop", row[1])
		assert.Equal(t, "beep", row[2])
		assert.Equal(t, "bop", row[3])
		assert.Equal(t, "blorp", row[4])

		// Third row should trigger a partialLineError
		_, err = csvReader.ReadSqlRow()
		require.Error(t, err)
		require.Equal(t, "incomplete record found at end of CSV data: 3,blue,g", err.Error())
	})

	t.Run("null and empty string quoting", func(t *testing.T) {
		csvReader, err := newCsvReader(newReader(nullAndEmptyStringQuotingCsvData))
		require.NoError(t, err)

		// Read the first row
		row, err := csvReader.ReadSqlRow()
		assert.NoError(t, err)
		assert.Equal(t, "1", row[0])
		assert.Equal(t, nil, row[1])
		assert.Equal(t, "NULL", row[2])
		assert.Equal(t, "NULL", row[3])
		assert.Equal(t, "", row[4])

		// Read the EOF error
		_, err = csvReader.ReadSqlRow()
		require.Equal(t, io.EOF, err)
	})

	t.Run("quote escaping", func(t *testing.T) {
		csvReader, err := newCsvReader(newReader(escapedQuotesCsvData))
		require.NoError(t, err)

		// Read the first row
		row, err := csvReader.ReadSqlRow()
		assert.NoError(t, err)
		assert.Equal(t, "1", row[0])
		assert.Equal(t, "''", row[1])
		assert.Equal(t, "'", row[2])
		assert.Equal(t, "\"", row[3])
		assert.Equal(t, "'", row[4])
		assert.Equal(t, "'''", row[5])

		// Read the EOF error
		_, err = csvReader.ReadSqlRow()
		require.Equal(t, io.EOF, err)
	})

	t.Run("quoted newlines", func(t *testing.T) {
		csvReader, err := newCsvReader(newReader(newLineInQuotedFieldCsvData))
		require.NoError(t, err)

		// Read the first row
		row, err := csvReader.ReadSqlRow()
		assert.NoError(t, err)
		assert.Equal(t, "1", row[0])
		assert.Equal(t, "foo", row[1])
		assert.Equal(t, "baz\nbar\nbash", row[2])

		// Read the EOF error
		_, err = csvReader.ReadSqlRow()
		require.Equal(t, io.EOF, err)
	})

	t.Run("quoted end of data marker", func(t *testing.T) {
		csvReader, err := newCsvReader(newReader(endOfDataMarkerCsvData))
		require.NoError(t, err)

		// Read the first row
		row, err := csvReader.ReadSqlRow()
		assert.NoError(t, err)
		assert.Equal(t, "1", row[0])
		assert.Equal(t, "foo", row[1])
		assert.Equal(t, "baz\n\\.\nbash", row[2])

		// Read the EOF error
		_, err = csvReader.ReadSqlRow()
		require.Equal(t, io.EOF, err)
	})

}

// testReader is a simple io.ReadCloser implementation that delegates reads to an io.Reader
// and implements Close() as a no-op.
type testReader struct {
	io.Reader
}

var _ io.ReadCloser = (*testReader)(nil)

// Close implements the io.Closer interface
func (frc *testReader) Close() error {
	return nil
}

// newReader returns an io.ReadCloser instance that reads the data from the specified
// string |s| and is a no-op when Close() is called.
func newReader(s string) io.ReadCloser {
	return &testReader{
		bytes.NewReader([]byte(s)),
	}
}
