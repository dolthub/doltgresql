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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/sequences"
	"github.com/dolthub/doltgresql/server/types"

	"github.com/dolthub/go-mysql-server/sql"
)

// ShowSequences is a node that implements the SHOW SCHEMAS	statement.
type ShowSequences struct {
	// TODO: we need planbuilder integration to support SHOW SCHEMAS, rather than getting everything at runtime
	database string
}

var _ sql.ExecSourceRel = (*DropTable)(nil)

// NewDropTable returns a new *DropTable.
func NewShowSequences(database string) *ShowSequences {
	return &ShowSequences{
		database: database,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (s *ShowSequences) Children() []sql.Node {
	return []sql.Node{}
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (s *ShowSequences) IsReadOnly() bool {
	return true
}

// Resolved implements the interface sql.ExecSourceRel.
func (s *ShowSequences) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (s *ShowSequences) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	database := s.database
	if database == "" {
		database = ctx.GetCurrentDatabase()
		if database == "" {
			return nil, errors.New("no database selected (this is a bug)")
		}
	}
	
	seqs, err := core.GetSequencesCollectionFromContextForDatabase(ctx, database)
	if err != nil {
		return nil, err
	}
	
	var rows []sql.Row
	err = seqs.IterateSequences(ctx, func(seq *sequences.Sequence) (stop bool, err error) {
		name := seq.Name()
		rows = append(rows, sql.Row{name.Schema, name.Name})
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	
	return sql.RowsToRowIter(rows...), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (s *ShowSequences) Schema() sql.Schema {
	return sql.Schema{
		{Name: "sequence_schema", Type: types.Text, Source: "show sequences"},
		{Name: "sequence_name", Type: types.Text, Source: "show sequences"},
	}
}

// String implements the interface sql.ExecSourceRel.
func (s *ShowSequences) String() string {
	if s.database == "" {
		return "SHOW SEQUENCES FROM " + s.database
	}
	return "SHOW SEQUENCES"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (s *ShowSequences) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 0 {
		return nil, errors.New("SHOW SCHEMAS does not support children")
	}
	return s, nil
}

func (s *ShowSequences) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, errors.New("SHOW SCHEMAS does not support children")
	}
	return s, nil
}
