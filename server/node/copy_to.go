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

package node

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// CopyTo handles the COPY ... TO STDOUT statement. Copying to a server-side file is not supported, as a security
// measure.
type CopyTo struct {
	DatabaseName string
	TableName    doltdb.TableName
	Columns      tree.NameList
	CopyOptions  tree.CopyOptions
	SelectStub   vitess.SelectStatement
}

var _ vitess.Injectable = (*CopyTo)(nil)
var _ sql.ExecSourceRel = (*CopyTo)(nil)

// NewCopyTo returns a new *CopyTo.
func NewCopyTo(
	databaseName string,
	tableName doltdb.TableName,
	options tree.CopyOptions,
	columns tree.NameList,
	selectStub vitess.SelectStatement,
) *CopyTo {
	switch options.CopyFormat {
	case tree.CopyFormatCsv, tree.CopyFormatText, tree.CopyFormatBinary:
		// no-op
	default:
		panic(fmt.Sprintf("unknown COPY TO format: %d", options.CopyFormat))
	}

	return &CopyTo{
		DatabaseName: databaseName,
		TableName:    tableName,
		Columns:      columns,
		CopyOptions:  options,
		SelectStub:   selectStub,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (ct *CopyTo) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (ct *CopyTo) IsReadOnly() bool {
	return true
}

// Resolved implements the interface sql.ExecSourceRel.
func (ct *CopyTo) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (ct *CopyTo) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	// COPY TO is handled directly by the connection handler, since it requires access to the wire connection to
	// stream results back to the client, so it should never be executed by the engine.
	return nil, errors.Errorf("COPY TO must be handled by the connection handler")
}

// Schema implements the interface sql.ExecSourceRel.
func (ct *CopyTo) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (ct *CopyTo) String() string {
	return "COPY TO STDOUT"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (ct *CopyTo) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(ct, len(children), 0)
	}
	return ct, nil
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (ct *CopyTo) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return ct, nil
}
