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

package tables

import (
	"io"

	"github.com/dolthub/go-mysql-server/sql"
)

// partitionIter is a partition iterator that returns a single partition.
type partitionIter struct {
	used bool
}

var _ sql.PartitionIter = (*partitionIter)(nil)

// Close implements the interface sql.PartitionIter.
func (iter *partitionIter) Close(ctx *sql.Context) error {
	return nil
}

// Next implements the interface sql.PartitionIter.
func (iter *partitionIter) Next(ctx *sql.Context) (sql.Partition, error) {
	if iter.used {
		return nil, io.EOF
	}
	iter.used = true
	return partition{}, nil
}

// partition is a dummy value that is returned from partitionIter.
type partition struct{}

var _ sql.Partition = partition{}

// Key implements the interface sql.Partition.
func (p partition) Key() []byte {
	return nil
}
