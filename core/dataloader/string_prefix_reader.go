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

import "io"

// stringPrefixReader is an io.ReadCloser that reads from a string prefix before reading from
// another io.Reader. This is used for reassembling partial records across multi-message
// exchanges, since the end of a wire message does not typically line up with the end of a record.
type stringPrefixReader struct {
	prefix         string
	prefixPosition uint
	reader         io.Reader
}

var _ io.ReadCloser = (*stringPrefixReader)(nil)

// NewStringPrefixReader creates a new stringPrefixReader that first returns the data in |prefix| and
// then returns data from |reader|.
func NewStringPrefixReader(prefix string, reader io.Reader) *stringPrefixReader {
	return &stringPrefixReader{
		prefix: prefix,
		reader: reader,
	}
}

// Read implements the io.Reader interface
func (spr *stringPrefixReader) Read(p []byte) (n int, err error) {
	if spr.prefixPosition < uint(len(spr.prefix)) {
		n = copy(p, spr.prefix[spr.prefixPosition:])
		spr.prefixPosition += uint(n)
		if n == len(p) {
			return n, nil
		}
	}

	read, err := spr.reader.Read(p[n:])
	return n + read, err
}

// Close implements the io.Closer interface
func (spr *stringPrefixReader) Close() error {
	if closer, ok := spr.reader.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}
