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
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/aggregates"
	"github.com/dolthub/doltgresql/core/casts"
	coreextensions "github.com/dolthub/doltgresql/core/extensions"
	"github.com/dolthub/doltgresql/core/functions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/operators"
	"github.com/dolthub/doltgresql/core/procedures"
	"github.com/dolthub/doltgresql/core/typecollection"
	"github.com/dolthub/doltgresql/server/extensions"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	"github.com/dolthub/doltgresql/server/types"
)

// CreateExtension implements CREATE EXTENSION.
type CreateExtension struct {
	Name        string
	IfNotExists bool
	SchemaName  string
	Version     string
	Cascade     bool
}

var _ sql.ExecSourceRel = (*CreateExtension)(nil)
var _ vitess.Injectable = (*CreateExtension)(nil)

// NewCreateExtension returns a new *CreateExtension.
func NewCreateExtension(name string, ifNotExists bool, schemaName string, version string, cascade bool) *CreateExtension {
	return &CreateExtension{
		Name:        name,
		IfNotExists: ifNotExists,
		SchemaName:  schemaName,
		Version:     version,
		Cascade:     cascade,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *CreateExtension) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *CreateExtension) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *CreateExtension) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *CreateExtension) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	typColl, err := core.GetTypesCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	extCollection, err := core.GetExtensionsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	if extCollection.HasLoadedExtension(ctx, id.NewExtension(c.Name)) {
		if c.IfNotExists {
			return sql.RowsToRowIter(), nil
		}
		return nil, errors.Errorf(`extension "%s" already exists`, c.Name)
	}
	// TODO: install the extensions named by Control.Requires, once an emulated extension declares any
	ext, err := extensions.Get(c.Name)
	if err != nil {
		return nil, err
	}

	schemaName, err := core.GetSchemaName(ctx, nil, c.SchemaName)
	if err != nil {
		return nil, err
	}
	if err = (extensionObjects{ext: ext, schemaName: schemaName, typColl: typColl}).materialize(ctx); err != nil {
		return nil, err
	}
	err = extCollection.AddLoadedExtension(ctx, coreextensions.Extension{
		ExtName:     id.NewExtension(c.Name),
		Namespace:   id.NewNamespace(schemaName),
		Relocatable: ext.Control.Relocatable,
		Version:     ext.Control.DefaultVersion,
	})
	if err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *CreateExtension) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *CreateExtension) String() string {
	return fmt.Sprintf("CREATE EXTENSION %s", c.Name)
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *CreateExtension) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *CreateExtension) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}

// extensionObjects materializes the objects that an extension declares.
type extensionObjects struct {
	ext        *extdef.Extension
	schemaName string
	typColl    *typecollection.TypeCollection
}

// materialize writes every object that the extension declares.
func (e extensionObjects) materialize(ctx *sql.Context) error {
	if err := e.materializeTypes(ctx); err != nil {
		return err
	}
	if err := e.materializeRoutines(ctx); err != nil {
		return err
	}
	if err := e.materializeOperators(ctx); err != nil {
		return err
	}
	if err := e.materializeCasts(ctx); err != nil {
		return err
	}
	return e.materializeAggregates(ctx)
}

// materializeTypes writes the declared types into the types collection.
func (e extensionObjects) materializeTypes(ctx *sql.Context) error {
	for _, declared := range e.ext.Types {
		var err error
		def := declared.Definition
		if def.InputFunc, err = e.supportFuncID(ctx, declared.Input); err != nil {
			return err
		}
		if def.OutputFunc, err = e.supportFuncID(ctx, declared.Output); err != nil {
			return err
		}
		if def.ReceiveFunc, err = e.supportFuncID(ctx, declared.Receive); err != nil {
			return err
		}
		if def.SendFunc, err = e.supportFuncID(ctx, declared.Send); err != nil {
			return err
		}
		if def.ModInFunc, err = e.supportFuncID(ctx, declared.ModIn); err != nil {
			return err
		}
		if def.ModOutFunc, err = e.supportFuncID(ctx, declared.ModOut); err != nil {
			return err
		}
		if def.CompareFunc, err = e.supportFuncID(ctx, declared.Compare); err != nil {
			return err
		}
		newType := types.NewBaseType(ctx, id.NewType(e.schemaName, declared.Name), def)
		if err = e.typColl.CreateType(ctx, newType); err != nil {
			return err
		}
		if err = e.typColl.CreateType(ctx, types.CreateArrayTypeFromBaseType(newType)); err != nil {
			return err
		}
	}
	return nil
}

// materializeRoutines writes the declared routines into the functions collection.
func (e extensionObjects) materializeRoutines(ctx *sql.Context) error {
	if len(e.ext.Routines) == 0 {
		return nil
	}
	funcCollection, err := core.GetFunctionsCollectionFromContext(ctx, "")
	if err != nil {
		return err
	}
	for _, routine := range e.ext.Routines {
		returnType, err := e.typeID(ctx, routine.Returns)
		if err != nil {
			return err
		}
		paramTypes, err := e.parameterTypes(ctx, routine.Parameters)
		if err != nil {
			return err
		}
		allParams := make([]procedures.Parameter, len(routine.Parameters))
		for i, param := range routine.Parameters {
			allParams[i] = procedures.Parameter{Name: param.Name, Type: paramTypes[i]}
		}
		err = funcCollection.AddFunction(ctx, functions.Function{
			ID:                 id.NewFunction(e.schemaName, routine.Name, paramTypes...),
			ReturnType:         returnType,
			AllParams:          allParams,
			IsNonDeterministic: true,
			Strict:             routine.Strict,
			ExtensionName:      e.ext.Name,
			ExtensionSymbol:    routine.Symbol,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// materializeOperators writes the declared operators into the operators collection.
func (e extensionObjects) materializeOperators(ctx *sql.Context) error {
	if len(e.ext.Operators) == 0 {
		return nil
	}
	opCollection, err := core.GetOperatorsCollectionFromContext(ctx, "")
	if err != nil {
		return err
	}
	for _, declared := range e.ext.Operators {
		leftType, err := e.typeID(ctx, declared.Left)
		if err != nil {
			return err
		}
		rightType, err := e.typeID(ctx, declared.Right)
		if err != nil {
			return err
		}
		routine, err := e.routine(declared.Routine)
		if err != nil {
			return err
		}
		returnType, err := e.typeID(ctx, routine.Returns)
		if err != nil {
			return err
		}
		funcID, err := e.routineID(ctx, routine)
		if err != nil {
			return err
		}
		err = opCollection.AddOperator(ctx, operators.Operator{
			ID:         id.NewOperator(e.schemaName, declared.Symbol, leftType, rightType),
			Function:   funcID,
			ReturnType: returnType,
			Commutator: declared.Commutator,
			Negator:    declared.Negator,
			Hashes:     declared.Hashes,
			Merges:     declared.Merges,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// materializeCasts writes the declared casts into the casts collection.
func (e extensionObjects) materializeCasts(ctx *sql.Context) error {
	if len(e.ext.Casts) == 0 {
		return nil
	}
	castCollection, err := core.GetCastsCollectionFromContext(ctx, "")
	if err != nil {
		return err
	}
	for _, declared := range e.ext.Casts {
		sourceType, err := e.typeID(ctx, declared.Source)
		if err != nil {
			return err
		}
		targetType, err := e.typeID(ctx, declared.Target)
		if err != nil {
			return err
		}
		funcID, err := e.optionalRoutineID(ctx, declared.Routine)
		if err != nil {
			return err
		}
		err = castCollection.AddCast(ctx, casts.Cast{
			ID:       id.NewCast(sourceType, targetType),
			CastType: declared.CastType,
			Function: funcID,
			UseInOut: !funcID.IsValid(),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// materializeAggregates writes the declared aggregates into the aggregates collection.
func (e extensionObjects) materializeAggregates(ctx *sql.Context) error {
	if len(e.ext.Aggregates) == 0 {
		return nil
	}
	aggCollection, err := core.GetAggregatesCollectionFromContext(ctx, "")
	if err != nil {
		return err
	}
	for _, declared := range e.ext.Aggregates {
		returnType, err := e.typeID(ctx, declared.Returns)
		if err != nil {
			return err
		}
		stateType, err := e.typeID(ctx, declared.StateType)
		if err != nil {
			return err
		}
		paramTypes, err := e.parameterTypes(ctx, declared.Parameters)
		if err != nil {
			return err
		}
		transitionFunc, err := e.optionalRoutineID(ctx, declared.Transition)
		if err != nil {
			return err
		}
		finalFunc, err := e.optionalRoutineID(ctx, declared.Final)
		if err != nil {
			return err
		}
		combineFunc, err := e.optionalRoutineID(ctx, declared.Combine)
		if err != nil {
			return err
		}
		err = aggCollection.AddAggregate(ctx, aggregates.Aggregate{
			ID:          id.NewFunction(e.schemaName, declared.Name, paramTypes...),
			ReturnType:  returnType,
			SFunc:       transitionFunc,
			SType:       stateType,
			FinalFunc:   finalFunc,
			CombineFunc: combineFunc,
			InitCond:    declared.InitCond,
			HasInitCond: declared.HasInitCond,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// typeID returns the type ID matching the given name.
func (e extensionObjects) typeID(ctx *sql.Context, name string) (id.Type, error) {
	for _, declared := range e.ext.Types {
		if declared.Name == name {
			return id.NewType(e.schemaName, name), nil
		}
	}
	_, typeID, err := e.typColl.ResolveName(ctx, doltdb.TableName{Name: name})
	if err != nil {
		return id.NullType, err
	}
	if !typeID.IsValid() {
		return id.NullType, types.ErrTypeDoesNotExist.New(name)
	}
	return id.Type(typeID), nil
}

// parameterTypes returns the type IDs of the given parameters.
func (e extensionObjects) parameterTypes(ctx *sql.Context, params []extdef.Parameter) ([]id.Type, error) {
	paramTypes := make([]id.Type, len(params))
	for i, param := range params {
		paramType, err := e.typeID(ctx, param.Type)
		if err != nil {
			return nil, err
		}
		paramTypes[i] = paramType
	}
	return paramTypes, nil
}

// routine returns the routine that the extension declares under the given symbol.
func (e extensionObjects) routine(symbol string) (extdef.Routine, error) {
	for _, routine := range e.ext.Routines {
		if routine.Symbol == symbol {
			return routine, nil
		}
	}
	return extdef.Routine{}, errors.Errorf(`extension "%s" does not declare the function "%s"`, e.ext.Name, symbol)
}

// routineID returns the ID that the given routine is materialized under.
func (e extensionObjects) routineID(ctx *sql.Context, routine extdef.Routine) (id.Function, error) {
	paramTypes, err := e.parameterTypes(ctx, routine.Parameters)
	if err != nil {
		return id.NullFunction, err
	}
	return id.NewFunction(e.schemaName, routine.Name, paramTypes...), nil
}

// optionalRoutineID returns the ID of the routine with the given symbol, or a null ID when the declaration omitted it.
func (e extensionObjects) optionalRoutineID(ctx *sql.Context, symbol string) (id.Function, error) {
	if len(symbol) == 0 {
		return id.NullFunction, nil
	}
	routine, err := e.routine(symbol)
	if err != nil {
		return id.NullFunction, err
	}
	return e.routineID(ctx, routine)
}

// supportFuncID returns the function registry ID of the routine with the given symbol, or zero when the type omitted
// it.
func (e extensionObjects) supportFuncID(ctx *sql.Context, symbol string) (uint32, error) {
	funcID, err := e.optionalRoutineID(ctx, symbol)
	if err != nil || !funcID.IsValid() {
		return 0, err
	}
	return types.ToFuncID(funcID), nil
}
