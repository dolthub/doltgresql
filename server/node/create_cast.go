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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// CreateCast implements CREATE CAST.
type CreateCast struct {
	Source     *pgtypes.DoltgresType
	Target     *pgtypes.DoltgresType
	Type       casts.CastType
	CreateType tree.CreateCastType
	FuncSchema string
	FuncName   string
	FuncParams []RoutineParam
}

var _ sql.ExecSourceRel = (*CreateCast)(nil)
var _ vitess.Injectable = (*CreateCast)(nil)

// NewCreateCast returns a new *CreateCast.
func NewCreateCast(
	source, target *pgtypes.DoltgresType,
	typ casts.CastType,
	createType tree.CreateCastType,
	funcSch, funcName string,
	params []RoutineParam) *CreateCast {
	return &CreateCast{
		Source:     source,
		Target:     target,
		Type:       typ,
		CreateType: createType,
		FuncSchema: funcSch,
		FuncName:   funcName,
		FuncParams: params,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *CreateCast) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *CreateCast) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *CreateCast) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *CreateCast) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	castCollection, err := core.GetCastsCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	newCast := casts.Cast{
		ID:       id.NewCast(c.Source.ID, c.Target.ID),
		CastType: c.Type,
	}
	if c.Source.ID == c.Target.ID {
		return nil, errors.New("source data type and target data type are the same")
	}
	switch c.CreateType {
	case tree.CreateCastType_WithFunction:
		if len(c.FuncParams) < 1 || len(c.FuncParams) > 3 {
			return nil, errors.New("cast function must take one to three arguments")
		}
		if len(c.FuncParams) >= 2 && c.FuncParams[1].Type.ID != pgtypes.Int32.ID {
			return nil, errors.New("second argument of cast function must be type integer")
		}
		if len(c.FuncParams) >= 3 && c.FuncParams[2].Type.ID != pgtypes.Bool.ID {
			return nil, errors.New("third argument of cast function must be type boolean")
		}
		schemaName, err := core.GetSchemaName(ctx, nil, c.FuncSchema)
		if err != nil {
			return nil, err
		}
		paramTypes := make([]id.Type, len(c.FuncParams))
		for i, param := range c.FuncParams {
			paramTypes[i] = param.Type.ID
		}
		funcID := id.NewFunction(schemaName, c.FuncName, paramTypes...)
		funcCollection, err := core.GetFunctionsCollectionFromContext(ctx)
		if err != nil {
			return nil, err
		}
		if !funcCollection.HasFunction(ctx, funcID) {
			return nil, errors.Errorf("function %s does not exist", funcID.DisplayString())
		}
		f, err := funcCollection.GetFunction(ctx, funcID)
		if err != nil {
			return nil, err
		}
		if f.ParameterTypes[0] != c.Source.ID {
			// Although the error mentions binary-coercible, we can't actually support that due to how types work in Go.
			// Our bit representations of values will differ from Postgres.
			return nil, errors.New("argument of cast function must match or be binary-coercible from source data type")
		}
		if f.ReturnType != c.Target.ID {
			// Although the error mentions binary-coercible, we can't actually support that due to how types work in Go.
			// Our bit representations of values will differ from Postgres.
			return nil, errors.New("return data type of cast function must match or be binary-coercible to target data type")
		}
		newCast.Function = funcID
	case tree.CreateCastType_WithoutFunction:
		if c.Source.IsCompositeType() || c.Target.IsCompositeType() {
			return nil, errors.New("composite data types are not binary-compatible")
		}
		// Due to the differences in how we handle values at the bit level compared to Postgres, we must mark all of
		// these types of casts as physically incompatible
		return nil, errors.New("source and target data types are not physically compatible")
	case tree.CreateCastType_Inout:
		newCast.UseInOut = true
	}
	if castCollection.HasCast(ctx, newCast.ID) {
		return nil, errors.Errorf("cast from type %s to type %s already exists", c.Source.ID.TypeName(), c.Target.ID.TypeName())
	}
	if err = castCollection.AddCast(ctx, newCast); err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *CreateCast) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *CreateCast) String() string {
	return "CREATE CAST"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *CreateCast) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *CreateCast) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
