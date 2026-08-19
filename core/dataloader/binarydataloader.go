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
	"bufio"
	"bytes"
	"encoding/binary"
	"io"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/types"
)

// BinaryDataLoader tracks the state of a load data operation from a data source in the binary COPY format: a file
// header, followed by tuples consisting of a 16-bit field count and each field as a 32-bit byte length (-1 for NULL)
// and the field data in its binary send format, terminated by a trailer word of -1. Field values are decoded with
// each column type's binary receive function.
type BinaryDataLoader struct {
	results       LoadDataResults
	partialRecord []byte
	nextDataChunk *bufio.Reader
	colTypes      []*types.DoltgresType
	sch           sql.Schema
	headerRead    bool
	sawTrailer    bool
}

var _ DataLoader = (*BinaryDataLoader)(nil)

// NewBinaryDataLoader creates a new BinaryDataLoader that will produce rows for the schema provided.
func NewBinaryDataLoader(colNames []string, tableSch sql.Schema) (*BinaryDataLoader, error) {
	colTypes, reducedSch, err := getColumnTypes(colNames, tableSch)
	if err != nil {
		return nil, err
	}

	return &BinaryDataLoader{
		colTypes: colTypes,
		sch:      reducedSch,
	}, nil
}

// nextRow returns the next SQL row from the reader provided. Returns true if there was another row. Records (the
// file header and each tuple) are not guaranteed to end cleanly on chunk boundaries, so any bytes consumed for a
// record the chunk cut short are saved to partialRecord and re-processed at the start of the next chunk.
func (bdl *BinaryDataLoader) nextRow(ctx *sql.Context, reader io.Reader) (sql.Row, bool, error) {
	if bdl.sawTrailer {
		var b [1]byte
		if _, err := io.ReadFull(reader, b[:]); err == nil {
			return nil, false, errors.Errorf("received copy data after EOF marker")
		}
		return nil, false, nil
	}

	// record accumulates every byte consumed for the current record, so that a record cut short by the end of the
	// chunk can be saved and re-processed once the next chunk arrives.
	var record []byte
	// readFull reads exactly n bytes into record, returning them. complete is false if the chunk ended first.
	readFull := func(n int) (data []byte, complete bool, err error) {
		start := len(record)
		record = append(record, make([]byte, n)...)
		read, err := io.ReadFull(reader, record[start:])
		record = record[:start+read]
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			bdl.partialRecord = record
			return nil, false, nil
		}
		if err != nil {
			return nil, false, err
		}
		return record[start:], true, nil
	}

	if !bdl.headerRead {
		header, complete, err := readFull(len(binarySignature) + 8)
		if err != nil || !complete {
			return nil, false, err
		}
		if !bytes.Equal(header[:len(binarySignature)], binarySignature) {
			return nil, false, errors.Errorf("COPY file signature not recognized")
		}
		flags := binary.BigEndian.Uint32(header[len(binarySignature):])
		// The high 16 bits of the flags field denote critical file format issues (such as the presence of OIDs,
		// which modern Postgres versions no longer support); the low 16 bits may be ignored.
		if flags>>16 != 0 {
			return nil, false, errors.Errorf("unrecognized critical flags in COPY file header")
		}
		extensionLen := int32(binary.BigEndian.Uint32(header[len(binarySignature)+4:]))
		if extensionLen < 0 {
			return nil, false, errors.Errorf("invalid extension length in COPY file header")
		}
		if _, complete, err = readFull(int(extensionLen)); err != nil || !complete {
			return nil, false, err
		}
		bdl.headerRead = true
		// The header was consumed cleanly, so the next record starts fresh
		record = nil
		bdl.partialRecord = nil
	}

	fieldCountBytes, complete, err := readFull(2)
	if err != nil || !complete {
		return nil, false, err
	}
	fieldCount := int16(binary.BigEndian.Uint16(fieldCountBytes))
	if fieldCount == -1 {
		// The trailer word, marking the end of the data
		bdl.sawTrailer = true
		bdl.partialRecord = nil
		var b [1]byte
		if _, err = io.ReadFull(reader, b[:]); err == nil {
			return nil, false, errors.Errorf("received copy data after EOF marker")
		}
		return nil, false, nil
	}
	if int(fieldCount) != len(bdl.colTypes) {
		return nil, false, errors.Errorf("row field count %d, expected %d", fieldCount, len(bdl.colTypes))
	}

	row := make(sql.Row, len(bdl.colTypes))
	for i := range bdl.colTypes {
		lengthBytes, complete, err := readFull(4)
		if err != nil || !complete {
			return nil, false, err
		}
		fieldLen := int32(binary.BigEndian.Uint32(lengthBytes))
		if fieldLen == -1 {
			row[i] = nil
			continue
		}
		if fieldLen < 0 {
			return nil, false, errors.Errorf("invalid field size %d", fieldLen)
		}
		fieldData, complete, err := readFull(int(fieldLen))
		if err != nil || !complete {
			return nil, false, err
		}
		row[i], err = bdl.colTypes[i].CallReceive(ctx, fieldData)
		if err != nil {
			return nil, false, err
		}
	}

	return row, true, nil
}

func (bdl *BinaryDataLoader) SetNextDataChunk(ctx *sql.Context, data *bufio.Reader) error {
	bdl.nextDataChunk = data
	return nil
}

// Finish completes the current load data operation and finalizes the data that has been inserted.
func (bdl *BinaryDataLoader) Finish(ctx *sql.Context) (*LoadDataResults, error) {
	// If there is partial data from the last chunk that hasn't been inserted, return an error.
	if len(bdl.partialRecord) > 0 {
		return nil, errors.Errorf("partial record found at end of data load")
	}
	if !bdl.sawTrailer {
		return nil, errors.Errorf("received unexpected EOF in COPY data, missing file trailer")
	}

	return &bdl.results, nil
}

func (bdl *BinaryDataLoader) Resolved() bool {
	return true
}

func (bdl *BinaryDataLoader) String() string {
	return "BinaryDataLoader"
}

func (bdl *BinaryDataLoader) Schema(ctx *sql.Context) sql.Schema {
	return bdl.sch
}

func (bdl *BinaryDataLoader) Children() []sql.Node {
	return nil
}

func (bdl *BinaryDataLoader) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(bdl, len(children), 0)
	}
	return bdl, nil
}

func (bdl *BinaryDataLoader) IsReadOnly() bool {
	return true
}

type binaryRowIter struct {
	bdl    *BinaryDataLoader
	reader io.Reader
}

var _ sql.RowIter = (*binaryRowIter)(nil)

func (b *binaryRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	row, hasNext, err := b.bdl.nextRow(ctx, b.reader)
	if err != nil {
		return nil, err
	}

	// TODO: this isn't the best way to handle the count of rows, something like a RowUpdateAccumulator would be better
	if hasNext {
		b.bdl.results.RowsLoaded++
	} else {
		return nil, io.EOF
	}

	return row, nil
}

func (b *binaryRowIter) Close(context *sql.Context) error {
	return nil
}

func (bdl *BinaryDataLoader) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	// Any partial record left over from the previous chunk is processed again, prefixed to this chunk's data.
	reader := io.MultiReader(bytes.NewReader(bdl.partialRecord), bdl.nextDataChunk)
	bdl.partialRecord = nil
	return &binaryRowIter{bdl: bdl, reader: reader}, nil
}
