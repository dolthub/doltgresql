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

package spatial

// Release builds embed the catalog compressed, which shrinks binaries by about 10MB. Its compressed form is not checked
// in, so builds with the compressed_catalog tag must first run `go generate ./server/extensions/postgis/spatial`, as
// scripts/build_binaries.sh does. Other builds embed the text itself.
//go:generate go run ./internal/compresscatalog crs_catalog.tsv crs_catalog.tsv.gz

import (
	"bytes"
	"compress/gzip"
	"io"

	"github.com/cockroachdb/errors"
)

// decompressCatalog returns the text of a gzip-compressed catalog.
func decompressCatalog(compressed []byte) string {
	reader, err := gzip.NewReader(bytes.NewReader(compressed))
	if err != nil {
		panic(errors.Wrap(err, "the embedded coordinate reference system catalog is corrupt"))
	}
	text, err := io.ReadAll(reader)
	if err != nil {
		panic(errors.Wrap(err, "the embedded coordinate reference system catalog is corrupt"))
	}
	return string(text)
}
