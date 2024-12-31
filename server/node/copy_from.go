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
	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/dataloader"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
)

// TODO: Privilege Checking: https://www.postgresql.org/docs/15/sql-copy.html

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
		InsertStub: insertStub,
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

// Validate returns an error if the CopyFrom node is invalid, for example if it contains columns that
// are not in the table schema.
//
// TODO: This validation logic should be hooked into the analyzer so that it can be run in a consistent way.
func (cf *CopyFrom) Validate(ctx *sql.Context) error {
	table, err := core.GetSqlTableFromContext(ctx, cf.DatabaseName, cf.TableName)
	if err != nil {
		return err
	}
	if table == nil {
		return fmt.Errorf(`relation "%s" does not exist`, cf.TableName.String())
	}
	if _, ok := table.(sql.InsertableTable); !ok {
		return fmt.Errorf(`table "%s" is read-only`, cf.TableName.String())
	}

	// If a set of columns was explicitly specified, validate them
	// TODO: remove, this check should happen during analysis
	// if len(cf.Columns) > 0 {
	// 	sch := table.Schema()
	// 	for _, col := range cf.Columns {
	// 		if sch.IndexOfColName(col.String()) < 0 {
	// 			return fmt.Errorf("invalid column %s for table %s", col.String(), table.Name())
	// 		}
	// 	}
	// }

	return nil
}

// RowIter implements the interface sql.ExecSourceRel.
func (cf *CopyFrom) RowIter(ctx *sql.Context, r sql.Row) (_ sql.RowIter, err error) {
	// TODO: implement file support
	// // Open the file
	// openFile, err := os.Open(cf.File)
	// if openFile == nil || err != nil {
	// 	return nil, fmt.Errorf(`could not open file "%s" for reading: No such file or directory`, cf.File)
	// }
	// defer func() {
	// 	nErr := openFile.Close()
	// 	if err == nil {
	// 		err = nErr
	// 	}
	// }()
	// reader := bufio.NewReader(openFile)

	return cf.DataLoader.RowIter(ctx, r)
}

// Schema implements the interface sql.ExecSourceRel.
func (cf *CopyFrom) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (cf *CopyFrom) String() string {
	return "COPY FROM"
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