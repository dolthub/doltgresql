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

// Command compresscatalog writes a gzip-compressed copy of a file, without a timestamp so that its output only depends
// on its input. It compresses crs_catalog.tsv for binaries built with the compressed_catalog tag.
package main

import (
	"compress/gzip"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: compresscatalog <input> <output>")
		os.Exit(2)
	}
	if err := compress(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// compress writes the gzip-compressed contents of the input file to the output file.
func compress(input string, output string) error {
	data, err := os.ReadFile(input)
	if err != nil {
		return err
	}
	file, err := os.Create(output)
	if err != nil {
		return err
	}
	writer, err := gzip.NewWriterLevel(file, gzip.BestCompression)
	if err != nil {
		file.Close()
		return err
	}
	if _, err = writer.Write(data); err != nil {
		file.Close()
		return err
	}
	if err = writer.Close(); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}
