// Copyright 2026 Dolthub, Inc.
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

// DataWriter encodes rows of data for a COPY TO operation, one row at a time. Rows are provided as the text-encoded
// values of each column, with a nil value indicating NULL. Each call to WriteRow returns the encoded bytes for a
// single row (including the row terminator), which callers either stream to the client as CopyData messages or write
// to a file.
type DataWriter interface {
	// WriteHeader returns the encoded header line for the data, or nil if no header was requested.
	WriteHeader() ([]byte, error)

	// WriteRow returns the encoded bytes for the row given. |vals| contains the encoded value of each column
	// (text-encoded for the text and CSV formats, binary-encoded for the binary format), with nil indicating a
	// NULL value.
	WriteRow(vals [][]byte) ([]byte, error)

	// WriteFooter returns the encoded footer for the data, or nil if this format has no footer.
	WriteFooter() ([]byte, error)
}
