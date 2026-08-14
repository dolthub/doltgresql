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
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/operators"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// CreateOperator implements CREATE OPERATOR.
type CreateOperator struct {
	Symbol     string
	Left       *pgtypes.DoltgresType // This is nil for prefix operators
	Right      *pgtypes.DoltgresType
	FuncSchema string
	FuncName   string
	Commutator string
	Negator    string
	Hashes     bool
	Merges     bool
}

var _ sql.ExecSourceRel = (*CreateOperator)(nil)
var _ vitess.Injectable = (*CreateOperator)(nil)

// NewCreateOperator returns a new *CreateOperator.
func NewCreateOperator(
	symbol string,
	left, right *pgtypes.DoltgresType,
	funcSchema, funcName string,
	commutator, negator string,
	hashes, merges bool) *CreateOperator {
	return &CreateOperator{
		Symbol:     symbol,
		Left:       left,
		Right:      right,
		FuncSchema: funcSchema,
		FuncName:   funcName,
		Commutator: commutator,
		Negator:    negator,
		Hashes:     hashes,
		Merges:     merges,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *CreateOperator) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *CreateOperator) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *CreateOperator) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *CreateOperator) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	if c.Left == nil {
		if c.Commutator != "" {
			return nil, errors.New("only binary operators can have commutators")
		}
		if c.Hashes {
			return nil, errors.New("only binary operators can hash")
		}
		if c.Merges {
			return nil, errors.New("only binary operators can merge join")
		}
	}
	funcSchema, err := core.GetSchemaName(ctx, nil, c.FuncSchema)
	if err != nil {
		return nil, err
	}
	var paramTypes []id.Type
	var paramNames []string
	if c.Left != nil {
		paramTypes = append(paramTypes, c.Left.ID)
		paramNames = append(paramNames, c.Left.String())
	}
	paramTypes = append(paramTypes, c.Right.ID)
	paramNames = append(paramNames, c.Right.String())
	funcID := id.NewFunction(funcSchema, c.FuncName, paramTypes...)
	funcCollection, err := core.GetFunctionsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	if !funcCollection.HasFunction(ctx, funcID) {
		return nil, errors.Errorf("function %s(%s) does not exist", c.FuncName, strings.Join(paramNames, ", "))
	}
	f, err := funcCollection.GetFunction(ctx, funcID)
	if err != nil {
		return nil, err
	}
	if f.ReturnType != pgtypes.Bool.ID {
		if c.Negator != "" {
			return nil, errors.New("only boolean operators can have negators")
		}
		if c.Hashes {
			return nil, errors.New("only boolean operators can hash")
		}
		if c.Merges {
			return nil, errors.New("only boolean operators can merge join")
		}
	}
	schemaName, err := core.GetSchemaName(ctx, nil, "")
	if err != nil {
		return nil, err
	}
	leftTypeID := id.NullType
	if c.Left != nil {
		leftTypeID = c.Left.ID
	}
	opCollection, err := core.GetOperatorsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	operatorID := id.NewOperator(schemaName, c.Symbol, leftTypeID, c.Right.ID)
	if opCollection.HasOperator(ctx, operatorID) {
		return nil, errors.Errorf("operator %s already exists", c.Symbol)
	}
	err = opCollection.AddOperator(ctx, operators.Operator{
		ID:         operatorID,
		Function:   funcID,
		ReturnType: f.ReturnType,
		Commutator: c.Commutator,
		Negator:    c.Negator,
		Hashes:     c.Hashes,
		Merges:     c.Merges,
	})
	if err != nil {
		return nil, err
	}
	if c.Commutator != "" {
		commutatorID := id.NewOperator(schemaName, c.Commutator, c.Right.ID, leftTypeID)
		if commutatorID != operatorID {
			if err = setCommutatorBackLink(ctx, opCollection, commutatorID, c.Symbol); err != nil {
				return nil, err
			}
		}
	}
	if c.Negator != "" {
		negatorID := id.NewOperator(schemaName, c.Negator, leftTypeID, c.Right.ID)
		if negatorID != operatorID {
			if err = setNegatorBackLink(ctx, opCollection, negatorID, c.Symbol); err != nil {
				return nil, err
			}
		}
	}
	return sql.RowsToRowIter(), nil
}

// setCommutatorBackLink sets the commutator of the given operator to the given symbol if the operator exists and has
// no commutator.
func setCommutatorBackLink(ctx *sql.Context, opCollection *operators.Collection, operatorID id.Operator, symbol string) error {
	if !opCollection.HasOperator(ctx, operatorID) {
		return nil
	}
	op, err := opCollection.GetOperator(ctx, operatorID)
	if err != nil || op.Commutator != "" {
		return err
	}
	op.Commutator = symbol
	if err = opCollection.DropOperator(ctx, operatorID); err != nil {
		return err
	}
	return opCollection.AddOperator(ctx, op)
}

// setNegatorBackLink sets the negator of the given operator to the given symbol if the operator exists and has no
// negator.
func setNegatorBackLink(ctx *sql.Context, opCollection *operators.Collection, operatorID id.Operator, symbol string) error {
	if !opCollection.HasOperator(ctx, operatorID) {
		return nil
	}
	op, err := opCollection.GetOperator(ctx, operatorID)
	if err != nil || op.Negator != "" {
		return err
	}
	op.Negator = symbol
	if err = opCollection.DropOperator(ctx, operatorID); err != nil {
		return err
	}
	return opCollection.AddOperator(ctx, op)
}

// Schema implements the interface sql.ExecSourceRel.
func (c *CreateOperator) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *CreateOperator) String() string {
	return fmt.Sprintf("CREATE OPERATOR %s", c.Symbol)
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *CreateOperator) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *CreateOperator) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
