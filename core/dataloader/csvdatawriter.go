// Copyright 2025 Dolthub, Inc.
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
)

// CsvDataWriter encodes rows for a COPY TO operation in the CSV format, mirroring the format accepted by
// CsvDataLoader.
type CsvDataWriter struct {
	colNames      []string
	delimiterChar string
	header        bool
}

var _ DataWriter = (*CsvDataWriter)(nil)

// NewCsvDataWriter creates a new CsvDataWriter for the columns named, using the specified |delimiterChar|. If
// |header| is true, a header line containing the column names will be returned by WriteHeader.
func NewCsvDataWriter(colNames []string, delimiterChar string, header bool) *CsvDataWriter {
	if delimiterChar == "" {
		delimiterChar = defaultCsvDelimiter
	}

	return &CsvDataWriter{
		colNames:      colNames,
		delimiterChar: delimiterChar,
		header:        header,
	}
}

// WriteHeader implements the DataWriter interface.
func (cdw *CsvDataWriter) WriteHeader() ([]byte, error) {
	if !cdw.header {
		return nil, nil
	}
	var buf bytes.Buffer
	for i, colName := range cdw.colNames {
		if i > 0 {
			buf.WriteString(cdw.delimiterChar)
		}
		cdw.writeField(&buf, []byte(colName), len(cdw.colNames) == 1)
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

// WriteRow implements the DataWriter interface.
func (cdw *CsvDataWriter) WriteRow(vals [][]byte) ([]byte, error) {
	var buf bytes.Buffer
	for i, val := range vals {
		if i > 0 {
			buf.WriteString(cdw.delimiterChar)
		}
		// A NULL value is written as an unquoted empty field, whereas an empty string value is quoted, which is how
		// the two are distinguished from one another.
		if val == nil {
			continue
		}
		cdw.writeField(&buf, val, len(vals) == 1)
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

// WriteFooter implements the DataWriter interface.
func (cdw *CsvDataWriter) WriteFooter() ([]byte, error) {
	return nil, nil
}

// writeField writes a single field value to the buffer, quoting it if necessary. |singleColumn| indicates that this
// field makes up the entire row, in which case a value matching the end-of-data marker must be quoted so that it
// isn't misinterpreted when read back.
func (cdw *CsvDataWriter) writeField(buf *bytes.Buffer, val []byte, singleColumn bool) {
	if !cdw.needsQuoting(val, singleColumn) {
		buf.Write(val)
		return
	}
	buf.WriteByte('"')
	for _, b := range val {
		if b == '"' {
			buf.WriteByte('"')
		}
		buf.WriteByte(b)
	}
	buf.WriteByte('"')
}

// needsQuoting returns whether the value given must be quoted in the CSV output. This matches Postgres's behavior:
// values containing the delimiter, quote, or newline characters are quoted, as are empty strings (to distinguish
// them from NULLs) and values that would be mistaken for the end-of-data marker.
func (cdw *CsvDataWriter) needsQuoting(val []byte, singleColumn bool) bool {
	if len(val) == 0 {
		return true
	}
	if singleColumn && bytes.Equal(val, []byte(`\.`)) {
		return true
	}
	delimiterByte := byte(0)
	if len(cdw.delimiterChar) == 1 {
		delimiterByte = cdw.delimiterChar[0]
	}
	for _, b := range val {
		switch b {
		case '"', '\n', '\r', delimiterByte:
			return true
		}
	}
	return false
}
