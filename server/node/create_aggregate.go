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
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/aggregates"
	"github.com/dolthub/doltgresql/core/functions"
	"github.com/dolthub/doltgresql/core/id"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// CreateAggregate implements CREATE AGGREGATE.
type CreateAggregate struct {
	SchemaName        string
	Name              string
	Replace           bool
	ArgTypes          []*pgtypes.DoltgresType
	SType             *pgtypes.DoltgresType
	SFuncSchema       string
	SFuncName         string
	FinalFuncSchema   string
	FinalFuncName     string
	CombineFuncSchema string
	CombineFuncName   string
	InitCond          string
	HasInitCond       bool
}

var _ sql.ExecSourceRel = (*CreateAggregate)(nil)
var _ vitess.Injectable = (*CreateAggregate)(nil)

// NewCreateAggregate returns a new *CreateAggregate.
func NewCreateAggregate(
	schemaName, name string,
	replace bool,
	argTypes []*pgtypes.DoltgresType,
	sType *pgtypes.DoltgresType,
	sFuncSchema, sFuncName string,
	finalFuncSchema, finalFuncName string,
	combineFuncSchema, combineFuncName string,
	initCond string,
	hasInitCond bool) *CreateAggregate {
	return &CreateAggregate{
		SchemaName:        schemaName,
		Name:              name,
		Replace:           replace,
		ArgTypes:          argTypes,
		SType:             sType,
		SFuncSchema:       sFuncSchema,
		SFuncName:         sFuncName,
		FinalFuncSchema:   finalFuncSchema,
		FinalFuncName:     finalFuncName,
		CombineFuncSchema: combineFuncSchema,
		CombineFuncName:   combineFuncName,
		InitCond:          initCond,
		HasInitCond:       hasInitCond,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *CreateAggregate) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *CreateAggregate) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *CreateAggregate) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *CreateAggregate) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	funcCollection, err := core.GetFunctionsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	argIDs := make([]id.Type, len(c.ArgTypes))
	argNames := make([]string, len(c.ArgTypes))
	for i, argType := range c.ArgTypes {
		argIDs[i] = argType.ID
		argNames[i] = argType.String()
	}
	sFuncID, err := c.lookupFunction(ctx, funcCollection, c.SFuncSchema, c.SFuncName,
		append([]id.Type{c.SType.ID}, argIDs...), append([]string{c.SType.String()}, argNames...))
	if err != nil {
		return nil, err
	}
	sFunc, err := funcCollection.GetFunction(ctx, sFuncID)
	if err != nil {
		return nil, err
	}
	if sFunc.ReturnType != c.SType.ID {
		return nil, errors.Errorf("return type of transition function %s is not %s", c.SFuncName, c.SType.String())
	}
	returnType := c.SType.ID
	finalFuncID := id.NullFunction
	if c.FinalFuncName != "" {
		finalFuncID, err = c.lookupFunction(ctx, funcCollection, c.FinalFuncSchema, c.FinalFuncName,
			[]id.Type{c.SType.ID}, []string{c.SType.String()})
		if err != nil {
			return nil, err
		}
		finalFunc, err := funcCollection.GetFunction(ctx, finalFuncID)
		if err != nil {
			return nil, err
		}
		returnType = finalFunc.ReturnType
	}
	combineFuncID := id.NullFunction
	if c.CombineFuncName != "" {
		combineFuncID, err = c.lookupFunction(ctx, funcCollection, c.CombineFuncSchema, c.CombineFuncName,
			[]id.Type{c.SType.ID, c.SType.ID}, []string{c.SType.String(), c.SType.String()})
		if err != nil {
			return nil, err
		}
		combineFunc, err := funcCollection.GetFunction(ctx, combineFuncID)
		if err != nil {
			return nil, err
		}
		if combineFunc.ReturnType != c.SType.ID {
			return nil, errors.Errorf("return type of combine function %s is not %s", c.CombineFuncName, c.SType.String())
		}
	}
	schemaName, err := core.GetSchemaName(ctx, nil, c.SchemaName)
	if err != nil {
		return nil, err
	}
	aggCollection, err := core.GetAggregatesCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	aggregateID := id.NewFunction(schemaName, c.Name, argIDs...)
	if funcCollection.HasFunction(ctx, aggregateID) {
		return nil, errors.Errorf(`function "%s" already exists with same argument types`, c.Name)
	}
	if aggCollection.HasAggregate(ctx, aggregateID) {
		if !c.Replace {
			return nil, errors.Errorf(`function "%s" already exists with same argument types`, c.Name)
		}
		if err = aggCollection.DropAggregate(ctx, aggregateID); err != nil {
			return nil, err
		}
	}
	err = aggCollection.AddAggregate(ctx, aggregates.Aggregate{
		ID:          aggregateID,
		ReturnType:  returnType,
		SFunc:       sFuncID,
		SType:       c.SType.ID,
		FinalFunc:   finalFuncID,
		CombineFunc: combineFuncID,
		InitCond:    c.InitCond,
		HasInitCond: c.HasInitCond,
	})
	if err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// lookupFunction returns the ID of the function with the given name and parameter types. Returns an error if no
// function was found.
func (c *CreateAggregate) lookupFunction(
	ctx *sql.Context,
	funcCollection *functions.Collection,
	schema string, name string,
	paramTypes []id.Type,
	paramNames []string) (id.Function, error) {
	funcSchema, err := core.GetSchemaName(ctx, nil, schema)
	if err != nil {
		return id.NullFunction, err
	}
	funcID := id.NewFunction(funcSchema, name, paramTypes...)
	if !funcCollection.HasFunction(ctx, funcID) {
		return id.NullFunction, errors.Errorf("function %s(%s) does not exist", name, strings.Join(paramNames, ", "))
	}
	return funcID, nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *CreateAggregate) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *CreateAggregate) String() string {
	return fmt.Sprintf("CREATE AGGREGATE %s", c.Name)
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *CreateAggregate) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *CreateAggregate) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
