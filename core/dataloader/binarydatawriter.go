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
	"encoding/binary"

	"github.com/cockroachdb/errors"
)

// binarySignature is the 11-byte sequence that begins the binary COPY format, identifying the format and providing
// early detection of file corruption (e.g. newline translation by a file transfer).
// See https://www.postgresql.org/docs/15/sql-copy.html#id-1.9.3.55.9.4 for the format specification.
var binarySignature = []byte("PGCOPY\n\377\r\n\000")

// BinaryDataWriter encodes rows for a COPY TO operation in the binary format. Values are expected to already be
// encoded in each type's binary send format (as produced for binary result values on the wire); this writer adds the
// file header, per-tuple field counts and lengths, and the file trailer that make up the binary COPY format.
type BinaryDataWriter struct{}

var _ DataWriter = (*BinaryDataWriter)(nil)

// NewBinaryDataWriter creates a new BinaryDataWriter.
func NewBinaryDataWriter() *BinaryDataWriter {
	return &BinaryDataWriter{}
}

// WriteHeader implements the DataWriter interface. The binary format always begins with a header, consisting of the
// format signature, a flags field, and a header extension area (empty, since we don't include OIDs or define any
// extensions).
func (bdw *BinaryDataWriter) WriteHeader() ([]byte, error) {
	var buf bytes.Buffer
	buf.Write(binarySignature)
	_ = binary.Write(&buf, binary.BigEndian, int32(0)) // flags field
	_ = binary.Write(&buf, binary.BigEndian, int32(0)) // header extension area length
	return buf.Bytes(), nil
}

// WriteRow implements the DataWriter interface. Each tuple consists of a 16-bit field count, followed by each field
// as a 32-bit byte length (-1 for NULL, with no value bytes following) and the field data in its binary send format.
func (bdw *BinaryDataWriter) WriteRow(vals [][]byte) ([]byte, error) {
	if len(vals) > 32767 {
		// should be impossible, Postgres's column limit is far lower
		return nil, errors.Errorf("too many columns for binary COPY: %d", len(vals))
	}
	var buf bytes.Buffer
	_ = binary.Write(&buf, binary.BigEndian, int16(len(vals)))
	for _, val := range vals {
		if val == nil {
			_ = binary.Write(&buf, binary.BigEndian, int32(-1))
			continue
		}
		_ = binary.Write(&buf, binary.BigEndian, int32(len(val)))
		buf.Write(val)
	}
	return buf.Bytes(), nil
}

// WriteFooter implements the DataWriter interface. The binary format ends with a trailer consisting of a 16-bit
// word containing -1, distinguishing a complete file from a truncated one (a tuple would begin with a non-negative
// field count).
func (bdw *BinaryDataWriter) WriteFooter() ([]byte, error) {
	return []byte{0xff, 0xff}, nil
}
