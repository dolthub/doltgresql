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

package node

import (
	"context"
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/functions"
)

// SetConstraints handles the SET CONSTRAINTS statement, treating every constraint as NOT DEFERRABLE.
type SetConstraints struct {
	Names    []*tree.UnresolvedObjectName
	Deferred bool
}

var _ sql.ExecSourceRel = (*SetConstraints)(nil)
var _ vitess.Injectable = (*SetConstraints)(nil)

// Children implements the interface sql.ExecSourceRel.
func (c *SetConstraints) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *SetConstraints) IsReadOnly() bool {
	return true
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *SetConstraints) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *SetConstraints) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	//TODO: support DEFERRABLE constraints, which CREATE TABLE and ALTER TABLE currently reject
	for _, name := range c.Names {
		if name.HasExplicitCatalog() && name.Catalog() != ctx.GetCurrentDatabase() {
			return nil, pgerror.Newf(pgcode.FeatureNotSupported, `cross-database references are not implemented: "%s"`, name.UnquotedString())
		}
		searchSchemas := []string{name.Schema()}
		if !name.HasExplicitSchema() {
			var err error
			if searchSchemas, err = core.SearchPath(ctx); err != nil {
				return nil, err
			}
		}
		schemaExists := false
		constraintExists := false
		err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
			Schema: func(ctx *sql.Context, schema functions.ItemSchema) (cont bool, err error) {
				schemaExists = true
				return true, nil
			},
			Check: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable, check functions.ItemCheck) (cont bool, err error) {
				constraintExists = constraintExists || check.Item.Name == name.Object()
				return true, nil
			},
			ForeignKey: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable, foreignKey functions.ItemForeignKey) (cont bool, err error) {
				constraintExists = constraintExists || foreignKey.Item.Name == name.Object()
				return true, nil
			},
			Index: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable, index functions.ItemIndex) (cont bool, err error) {
				indexName := index.Item.ID()
				if indexName == "PRIMARY" {
					indexName = fmt.Sprintf("%s_pkey", index.Item.Table())
				} else if !index.Item.IsUnique() {
					return true, nil
				}
				constraintExists = constraintExists || indexName == name.Object()
				return true, nil
			},
			Type: func(ctx *sql.Context, schema functions.ItemSchema, typ functions.ItemType) (cont bool, err error) {
				for _, check := range typ.Item.Checks {
					constraintExists = constraintExists || check.Name == name.Object()
				}
				return true, nil
			},
			SearchSchemas: searchSchemas,
		})
		if err != nil {
			return nil, err
		}
		if name.HasExplicitSchema() && !schemaExists {
			return nil, pgerror.Newf(pgcode.UndefinedSchema, `schema "%s" does not exist`, name.Schema())
		}
		if !constraintExists {
			return nil, pgerror.Newf(pgcode.UndefinedObject, `constraint "%s" does not exist`, name.Object())
		}
		if c.Deferred {
			return nil, pgerror.Newf(pgcode.WrongObjectType, `constraint "%s" is not deferrable`, name.Object())
		}
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *SetConstraints) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *SetConstraints) String() string {
	return "SET CONSTRAINTS"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *SetConstraints) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *SetConstraints) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
