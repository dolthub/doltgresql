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
	"bufio"
	"fmt"
	"os"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/core"
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
}

var _ vitess.Injectable = (*CopyFrom)(nil)
var _ sql.ExecSourceRel = (*CopyFrom)(nil)

// NewCopyFrom returns a new *CopyFrom.
func NewCopyFrom(databaseName string, tableName doltdb.TableName, options tree.CopyOptions, fileName string, stdin bool, columns tree.NameList) *CopyFrom {
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
	}
}

// CheckPrivileges implements the interface sql.ExecSourceRel.
func (cf *CopyFrom) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
	// https://www.postgresql.org/docs/15/sql-copy.html
	// ... database superusers or users who are granted one of the roles `pg_read_server_files`, `pg_write_server_files`, or `pg_execute_server_program` ...
	return true
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
	if len(cf.Columns) > 0 {
		if len(table.Schema()) != len(cf.Columns) {
			return fmt.Errorf("invalid column name list for table %s: %v", table.Name(), cf.Columns)
		}

		for i, col := range table.Schema() {
			name := cf.Columns[i]
			if name.String() != col.Name {
				return fmt.Errorf("invalid column name list for table %s: %v", table.Name(), cf.Columns)
			}
		}
	}

	return nil
}

// RowIter implements the interface sql.ExecSourceRel.
func (cf *CopyFrom) RowIter(ctx *sql.Context, _ sql.Row) (_ sql.RowIter, err error) {
	if err := cf.Validate(ctx); err != nil {
		return nil, err
	}

	table, err := core.GetSqlTableFromContext(ctx, cf.DatabaseName, cf.TableName)
	if err != nil {
		return nil, err
	}
	if table == nil {
		return nil, fmt.Errorf(`relation "%s" does not exist`, cf.TableName.String())
	}
	insertable, ok := table.(sql.InsertableTable)
	if !ok {
		return nil, fmt.Errorf(`table "%s" is read-only`, cf.TableName.String())
	}

	// Open the file
	openFile, err := os.Open(cf.File)
	if openFile == nil || err != nil {
		return nil, fmt.Errorf(`could not open file "%s" for reading: No such file or directory`, cf.File)
	}
	defer func() {
		nErr := openFile.Close()
		if err == nil {
			err = nErr
		}
	}()
	reader := bufio.NewReader(openFile)

	dataLoader, err := dataloader.NewTabularDataLoader(ctx, insertable, cf.CopyOptions.Delimiter, "", cf.CopyOptions.Header)
	if err != nil {
		return nil, err
	}

	// NOTE: when loading data from a specified file, there is only one chunk for the entire file
	if err = dataLoader.LoadChunk(ctx, reader); err != nil {
		if abortError := dataLoader.Abort(ctx); abortError != nil {
			logrus.Warnf("unable to cleanly abort data loader: %s", abortError.Error())
		}
		return nil, err
	}

	if _, err = dataLoader.Finish(ctx); err != nil {
		if abortError := dataLoader.Abort(ctx); err != nil {
			logrus.Warnf("unable to cleanly abort data loader: %s", abortError.Error())
		}
		return nil, err
	}

	return sql.RowsToRowIter(), nil
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
		return nil, fmt.Errorf("invalid vitess child count, expected `0` but got `%d`", len(children))
	}
	return cf, nil
}
