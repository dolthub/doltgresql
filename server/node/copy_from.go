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
	"io"
	"os"
	"strings"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// CopyFrom handles the COPY ... FROM ... statement.
type CopyFrom struct {
	DatabaseName string
	TableName    doltdb.TableName
	File         string
	Delimiter    string
	Null         string
}

var _ vitess.Injectable = (*CopyFrom)(nil)
var _ sql.ExecSourceRel = (*CopyFrom)(nil)

// NewCopyFrom returns a new *CopyFrom.
func NewCopyFrom(databaseName string, tableName doltdb.TableName, fileName string) *CopyFrom {
	return &CopyFrom{
		DatabaseName: databaseName,
		TableName:    tableName,
		File:         fileName,
		Delimiter:    "\t",
		Null:         `\N`,
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

// RowIter implements the interface sql.ExecSourceRel.
func (cf *CopyFrom) RowIter(ctx *sql.Context, r sql.Row) (_ sql.RowIter, err error) {
	// Fetch the table
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

	// Get the row inserter and set the defers
	rowInserter := insertable.Inserter(ctx)
	defer func() {
		nErr := rowInserter.Close(ctx)
		if err == nil {
			err = nErr
		}
	}()
	rowInserter.StatementBegin(ctx)
	defer func() {
		if err == nil {
			err = rowInserter.StatementComplete(ctx)
		} else {
			err = rowInserter.DiscardChanges(ctx, err)
		}
	}()

	// Get the column's types, which we'll use for casting
	sch := insertable.Schema()
	colTypes := make([]pgtypes.DoltgresType, len(sch))
	for i, col := range sch {
		var ok bool
		colTypes[i], ok = col.Type.(pgtypes.DoltgresType)
		if !ok {
			return nil, fmt.Errorf("COPY FROM only works for tables with all Postgres types")
		}
	}

	// Write the data to the table
	row := make(sql.Row, len(colTypes))
	foundEOF := false
	for !foundEOF {
		// Read the next line from the file
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			foundEOF = true
		} else {
			// If we've not reached EOF, then there will be a newline appended to the end that we must remove.
			// We'll check though just for thoroughness.
			line = strings.TrimSuffix(line, "\n")
		}
		if len(line) == 0 {
			continue
		}
		// Split the values by the delimiter, ensuring the correct number of values have been read
		values := strings.Split(line, cf.Delimiter)
		if len(values) > len(colTypes) {
			return nil, fmt.Errorf("extra data after last expected column")
		} else if len(values) < len(colTypes) {
			return nil, fmt.Errorf(`missing data for column "%s"`, sch[len(values)].Name)
		}
		// Cast the values using I/O input
		for i := range colTypes {
			if values[i] == cf.Null {
				row[i] = nil
			} else {
				row[i], err = colTypes[i].IoInput(ctx, values[i])
				if err != nil {
					return nil, err
				}
			}
		}
		// Insert the read row
		err = rowInserter.Insert(ctx, row)
		if err != nil {
			return nil, err
		}
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
		return nil, sql.ErrInvalidChildrenNumber.New(cf, len(children), 1)
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
