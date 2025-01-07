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

package node

import (
	"fmt"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core/dataloader"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// CopyFrom handles the COPY ... FROM ... statement.
type CopyFrom struct {
	DatabaseName string
	TableName    doltdb.TableName
	File         string
	Stdin        bool
	Columns      tree.NameList
	CopyOptions  tree.CopyOptions
	InsertStub   *vitess.Insert
	DataLoader   dataloader.DataLoader
}

var _ vitess.Injectable = (*CopyFrom)(nil)
var _ sql.ExecSourceRel = (*CopyFrom)(nil)

// NewCopyFrom returns a new *CopyFrom.
func NewCopyFrom(
	databaseName string,
	tableName doltdb.TableName,
	options tree.CopyOptions,
	fileName string,
	stdin bool,
	columns tree.NameList,
	insertStub *vitess.Insert,
) *CopyFrom {
	switch options.CopyFormat {
	case tree.CopyFormatCsv, tree.CopyFormatText:
		// no-op
	case tree.CopyFormatBinary:
		panic("BINARY format is not supported for COPY FROM")
	default:
		panic(fmt.Sprintf("unknown COPY FROM format: %d", options.CopyFormat))
	}

	return &CopyFrom{
		DatabaseName: databaseName,
		TableName:    tableName,
		File:         fileName,
		Stdin:        stdin,
		Columns:      columns,
		CopyOptions:  options,
		InsertStub:   insertStub,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (cf *CopyFrom) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (cf *CopyFrom) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (cf *CopyFrom) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (cf *CopyFrom) RowIter(ctx *sql.Context, r sql.Row) (_ sql.RowIter, err error) {
	return cf.DataLoader.RowIter(ctx, r)
}

// Schema implements the interface sql.ExecSourceRel.
func (cf *CopyFrom) Schema() sql.Schema {
	// For Parse calls, we need access to the schema before we have a DataLoader created, so return a stub schema.
	if cf.DataLoader == nil {
		return nil
	}
	return cf.DataLoader.Schema()
}

// String implements the interface sql.ExecSourceRel.
func (cf *CopyFrom) String() string {
	source := "STDIN"
	if cf.File != "" {
		source = fmt.Sprintf("'%s'", cf.File)
	}
	return fmt.Sprintf("COPY FROM %s", source)
}

// WithChildren implements the interface sql.ExecSourceRel.
func (cf *CopyFrom) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(cf, len(children), 0)
	}
	return cf, nil
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (cf *CopyFrom) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return cf, nil
}
