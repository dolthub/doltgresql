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

package tables

import (
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/auth"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// sequenceTable exposes a sequence's state as PostgreSQL's read-only relation.
type sequenceTable struct {
	database string
	name     doltdb.TableName
}

// resolveSequenceTable resolves a sequence after regular table lookup fails.
func resolveSequenceTable(ctx *sql.Context, schema sql.DatabaseSchema, tblName string) (sql.Table, bool, error) {
	collection, err := core.GetSequencesCollectionFromContext(ctx, schema.Name())
	if err != nil {
		return nil, false, err
	}
	name := doltdb.TableName{Schema: schema.SchemaName(), Name: tblName}
	if name.Schema == "" {
		path, err := core.SearchPath(ctx)
		if err != nil {
			return nil, false, err
		}
		for _, schemaName := range path {
			sequence, err := collection.GetSequence(ctx, id.NewSequence(schemaName, tblName))
			if err != nil {
				return nil, false, err
			}
			if sequence != nil {
				name.Schema = schemaName
				break
			}
		}
	}
	sequence, err := collection.GetSequence(ctx, id.NewSequence(name.Schema, name.Name))
	if err != nil || sequence == nil {
		return nil, false, err
	}
	return NewVirtualTable(sequenceTable{database: schema.Name(), name: name}, schema), true, nil
}

// Name implements Handler.
func (s sequenceTable) Name() string { return s.name.Name }

// PkSchema implements Handler with PostgreSQL's sequence relation columns.
func (s sequenceTable) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{Schema: sql.Schema{
		{Name: "last_value", Type: pgtypes.Int64, Source: s.name.Name},
		// Doltgres does not use PostgreSQL's WAL sequence preallocation counter.
		{Name: "log_cnt", Type: pgtypes.Int64, Source: s.name.Name},
		{Name: "is_called", Type: pgtypes.Bool, Source: s.name.Name},
	}}
}

// RowIter implements Handler. Check authorization on every execution, including
// previously analyzed plans, before returning any sequence state.
func (s sequenceTable) RowIter(ctx *sql.Context, _ sql.Partition) (sql.RowIter, error) {
	if err := auth.CheckSequencePrivileges(ctx, s.name.Schema, s.name.Name, auth.Privilege_SELECT); err != nil {
		return nil, err
	}
	collection, err := core.GetSequencesCollectionFromContext(ctx, s.database)
	if err != nil {
		return nil, err
	}
	sequence, err := collection.GetSequence(ctx, id.NewSequence(s.name.Schema, s.name.Name))
	if err != nil {
		return nil, err
	}
	if sequence == nil {
		return nil, sql.ErrTableNotFound.New(s.name.Name)
	}
	last := sequence.Current
	if sequence.HasBeenCalled && !sequence.IsAtEnd {
		last -= sequence.Increment
	}
	return sql.RowsToRowIter(sql.Row{last, int64(0), sequence.HasBeenCalled}), nil
}
