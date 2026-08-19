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

// TabularDataWriter encodes rows for a COPY TO operation in the text format, mirroring the format accepted by
// TabularDataLoader.
type TabularDataWriter struct {
	colNames      []string
	delimiterChar string
	nullChar      string
	header        bool
}

var _ DataWriter = (*TabularDataWriter)(nil)

// NewTabularDataWriter creates a new TabularDataWriter for the columns named, using the specified |delimiterChar| and
// |nullChar|. If |header| is true, a header line containing the column names will be returned by WriteHeader.
func NewTabularDataWriter(colNames []string, delimiterChar, nullChar string, header bool) *TabularDataWriter {
	if delimiterChar == "" {
		delimiterChar = defaultTextDelimiter
	}

	if nullChar == "" {
		nullChar = defaultNullChar
	}

	return &TabularDataWriter{
		colNames:      colNames,
		delimiterChar: delimiterChar,
		nullChar:      nullChar,
		header:        header,
	}
}

// WriteHeader implements the DataWriter interface.
func (tdw *TabularDataWriter) WriteHeader() ([]byte, error) {
	if !tdw.header {
		return nil, nil
	}
	var buf bytes.Buffer
	for i, colName := range tdw.colNames {
		if i > 0 {
			buf.WriteString(tdw.delimiterChar)
		}
		tdw.writeEscaped(&buf, []byte(colName))
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

// WriteRow implements the DataWriter interface.
func (tdw *TabularDataWriter) WriteRow(vals [][]byte) ([]byte, error) {
	var buf bytes.Buffer
	for i, val := range vals {
		if i > 0 {
			buf.WriteString(tdw.delimiterChar)
		}
		if val == nil {
			buf.WriteString(tdw.nullChar)
		} else {
			tdw.writeEscaped(&buf, val)
		}
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}

// WriteFooter implements the DataWriter interface.
func (tdw *TabularDataWriter) WriteFooter() ([]byte, error) {
	return nil, nil
}

// writeEscaped writes the value given to the buffer, escaping any control characters, backslashes, and delimiter
// characters with a backslash, matching the escaping performed by Postgres for the text format.
func (tdw *TabularDataWriter) writeEscaped(buf *bytes.Buffer, val []byte) {
	delimiterByte := byte(0)
	if len(tdw.delimiterChar) == 1 {
		delimiterByte = tdw.delimiterChar[0]
	}
	for _, b := range val {
		switch b {
		case '\\':
			buf.WriteString(`\\`)
		case '\b':
			buf.WriteString(`\b`)
		case '\f':
			buf.WriteString(`\f`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		case '\v':
			buf.WriteString(`\v`)
		default:
			if b == delimiterByte {
				buf.WriteByte('\\')
			}
			buf.WriteByte(b)
		}
	}
}
