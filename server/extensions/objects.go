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

package extensions

import (
	"slices"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/aggregates"
	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/core/functions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/operators"
	"github.com/dolthub/doltgresql/core/procedures"
	"github.com/dolthub/doltgresql/core/typecollection"
	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
	"github.com/dolthub/doltgresql/server/hook"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// CreateObjects writes every object that the given extension declares into the given schema.
func CreateObjects(ctx *sql.Context, ext *extdef.Extension, schemaName string) error {
	typColl, err := core.GetTypesCollectionFromContext(ctx, "")
	if err != nil {
		return err
	}
	return extensionObjects{ext: ext, schemaName: schemaName, typColl: typColl}.materialize(ctx)
}

// CheckDependents returns an error if a table, a domain, a CHECK constraint, or a view depends on an object that the
// given extension declares in the given schema.
func CheckDependents(ctx *sql.Context, ext *extdef.Extension, schemaName string) error {
	typColl, err := core.GetTypesCollectionFromContext(ctx, "")
	if err != nil {
		return err
	}
	return extensionObjects{ext: ext, schemaName: schemaName, typColl: typColl}.checkDependents(ctx)
}

// DropObjects removes every object that the given extension declares from the given schema.
func DropObjects(ctx *sql.Context, ext *extdef.Extension, schemaName string) error {
	typColl, err := core.GetTypesCollectionFromContext(ctx, "")
	if err != nil {
		return err
	}
	return extensionObjects{ext: ext, schemaName: schemaName, typColl: typColl}.drop(ctx)
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
		newType := pgtypes.NewBaseType(ctx, id.NewType(e.schemaName, declared.Name), def)
		if err = e.typColl.CreateType(ctx, newType); err != nil {
			return err
		}
		if err = e.typColl.CreateType(ctx, pgtypes.CreateArrayTypeFromBaseType(newType)); err != nil {
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

// drop removes every object that the extension declares, in the reverse order of materialize.
func (e extensionObjects) drop(ctx *sql.Context) error {
	if err := e.dropAggregates(ctx); err != nil {
		return err
	}
	if err := e.dropCasts(ctx); err != nil {
		return err
	}
	if err := e.dropOperators(ctx); err != nil {
		return err
	}
	if err := e.dropRoutines(ctx); err != nil {
		return err
	}
	return e.dropTypes(ctx)
}

// checkDependents returns an error if a table, a domain, a CHECK constraint, or a view depends on one of the declared
// objects.
func (e extensionObjects) checkDependents(ctx *sql.Context) error {
	finder := dependencyFinder{ext: e.ext, typeIDs: e.declaredTypeIDs()}
	db, err := core.GetSqlDatabaseFromContext(ctx, "")
	if err != nil {
		return err
	}
	schemaDb, ok := db.(sql.SchemaDatabase)
	if !ok {
		return nil
	}
	schemas, err := schemaDb.AllSchemas(ctx)
	if err != nil {
		return err
	}
	for _, schema := range schemas {
		tableNames, err := schema.GetTableNames(ctx)
		if err != nil {
			return err
		}
		for _, tableName := range tableNames {
			table, ok, err := schema.GetTableInsensitive(ctx, tableName)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			if err = finder.walkTable(ctx, table); err != nil {
				return err
			}
		}
	}
	if !finder.found {
		err = e.typColl.IterateTypes(ctx, func(typ *pgtypes.DoltgresType) (stop bool, err error) {
			if typ.TypType == pgtypes.TypeType_Domain && !finder.found {
				err = finder.walkDomain(typ)
			}
			return false, err
		})
		if err != nil {
			return err
		}
	}
	if !finder.found {
		if err = hook.WalkViewExpressions(ctx, &finder); err != nil {
			return err
		}
	}
	if finder.found {
		return pgerror.Newf(pgcode.DependentObjectsStillExist,
			"cannot drop extension %s because other objects depend on it", e.ext.Name)
	}
	return nil
}

// dropAggregates removes the declared aggregates from the aggregates collection.
func (e extensionObjects) dropAggregates(ctx *sql.Context) error {
	if len(e.ext.Aggregates) == 0 {
		return nil
	}
	aggCollection, err := core.GetAggregatesCollectionFromContext(ctx, "")
	if err != nil {
		return err
	}
	aggregateIDs := make([]id.Function, len(e.ext.Aggregates))
	for i, declared := range e.ext.Aggregates {
		paramTypes, err := e.parameterTypes(ctx, declared.Parameters)
		if err != nil {
			return err
		}
		aggregateIDs[i] = id.NewFunction(e.schemaName, declared.Name, paramTypes...)
	}
	return aggCollection.DropAggregate(ctx, aggregateIDs...)
}

// dropCasts removes the declared casts from the casts collection.
func (e extensionObjects) dropCasts(ctx *sql.Context) error {
	if len(e.ext.Casts) == 0 {
		return nil
	}
	castCollection, err := core.GetCastsCollectionFromContext(ctx, "")
	if err != nil {
		return err
	}
	castIDs := make([]id.Cast, len(e.ext.Casts))
	for i, declared := range e.ext.Casts {
		sourceType, err := e.typeID(ctx, declared.Source)
		if err != nil {
			return err
		}
		targetType, err := e.typeID(ctx, declared.Target)
		if err != nil {
			return err
		}
		castIDs[i] = id.NewCast(sourceType, targetType)
	}
	return castCollection.DropCast(ctx, castIDs...)
}

// dropOperators removes the declared operators from the operators collection.
func (e extensionObjects) dropOperators(ctx *sql.Context) error {
	if len(e.ext.Operators) == 0 {
		return nil
	}
	opCollection, err := core.GetOperatorsCollectionFromContext(ctx, "")
	if err != nil {
		return err
	}
	operatorIDs := make([]id.Operator, len(e.ext.Operators))
	for i, declared := range e.ext.Operators {
		leftType, err := e.typeID(ctx, declared.Left)
		if err != nil {
			return err
		}
		rightType, err := e.typeID(ctx, declared.Right)
		if err != nil {
			return err
		}
		operatorIDs[i] = id.NewOperator(e.schemaName, declared.Symbol, leftType, rightType)
	}
	return opCollection.DropOperator(ctx, operatorIDs...)
}

// dropRoutines removes the declared routines from the functions collection.
func (e extensionObjects) dropRoutines(ctx *sql.Context) error {
	if len(e.ext.Routines) == 0 {
		return nil
	}
	funcCollection, err := core.GetFunctionsCollectionFromContext(ctx, "")
	if err != nil {
		return err
	}
	funcIDs := make([]id.Function, len(e.ext.Routines))
	for i, routine := range e.ext.Routines {
		if funcIDs[i], err = e.routineID(ctx, routine); err != nil {
			return err
		}
	}
	return funcCollection.DropFunction(ctx, funcIDs...)
}

// dropTypes removes the declared types and their array types from the types collection.
func (e extensionObjects) dropTypes(ctx *sql.Context) error {
	if len(e.ext.Types) == 0 {
		return nil
	}
	return e.typColl.DropType(ctx, e.declaredTypeIDs()...)
}

// declaredTypeIDs returns the IDs of the declared types and their array types.
func (e extensionObjects) declaredTypeIDs() []id.Type {
	typeIDs := make([]id.Type, 0, len(e.ext.Types)*2)
	for _, declared := range e.ext.Types {
		typeIDs = append(typeIDs, id.NewType(e.schemaName, declared.Name), id.NewType(e.schemaName, "_"+declared.Name))
	}
	return typeIDs
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
		return id.NullType, pgtypes.ErrTypeDoesNotExist.New(name)
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
	return pgtypes.ToFuncID(funcID), nil
}

// dependencyFinder finds whether a table, a domain, a CHECK constraint, or a view uses an object that an extension
// declares. Routines are matched by name, since stored expressions do not keep the schema of the routines they call.
type dependencyFinder struct {
	ext     *extdef.Extension
	typeIDs []id.Type
	found   bool
}

var _ tree.Visitor = (*dependencyFinder)(nil)

// VisitPre implements the interface tree.Visitor.
func (f *dependencyFinder) VisitPre(expr tree.Expr) (recurse bool, newExpr tree.Expr) {
	switch expr := expr.(type) {
	case *tree.FuncExpr:
		if name, ok := expr.Func.FunctionReference.(*tree.UnresolvedName); ok {
			f.found = f.found || slices.ContainsFunc(f.ext.Routines, func(routine extdef.Routine) bool {
				return routine.Name == name.Parts[0]
			})
		}
	case *tree.CastExpr:
		f.found = f.found || f.namesDeclaredType(expr.Type)
	}
	return !f.found, expr
}

// VisitPost implements the interface tree.Visitor.
func (f *dependencyFinder) VisitPost(expr tree.Expr) tree.Expr {
	return expr
}

// walkTable checks the column types, the stored default and generated expressions, and the CHECK constraints of the
// given table.
func (f *dependencyFinder) walkTable(ctx *sql.Context, table sql.Table) error {
	for _, col := range table.Schema(ctx) {
		if colType, ok := col.Type.(*pgtypes.DoltgresType); ok && slices.Contains(f.typeIDs, colType.ID) {
			f.found = true
		}
		for _, colDefault := range []*sql.ColumnDefaultValue{col.Default, col.Generated} {
			if colDefault == nil {
				continue
			}
			if unresolved, ok := colDefault.Expr.(*sql.UnresolvedColumnDefault); ok {
				if err := f.walkText(unresolved.String()); err != nil {
					return err
				}
			}
		}
	}
	checkTable, ok := table.(sql.CheckTable)
	if !ok {
		return nil
	}
	checks, err := checkTable.GetChecks(ctx)
	if err != nil {
		return err
	}
	for _, check := range checks {
		if err = f.walkText(check.CheckExpression); err != nil {
			return err
		}
	}
	return nil
}

// walkDomain checks the base type, the default, and the CHECK constraints of the given domain.
func (f *dependencyFinder) walkDomain(domain *pgtypes.DoltgresType) error {
	if slices.Contains(f.typeIDs, domain.BaseTypeType.ID) {
		f.found = true
	}
	if len(domain.Default) > 0 {
		if err := f.walkText(domain.Default); err != nil {
			return err
		}
	}
	for _, check := range domain.Checks {
		if err := f.walkText(check.CheckExpression); err != nil {
			return err
		}
	}
	return nil
}

// walkText parses the given expression and walks it.
func (f *dependencyFinder) walkText(text string) error {
	expr, err := parser.ParseExpr(text)
	if err != nil {
		return err
	}
	tree.WalkExpr(f, expr)
	return nil
}

// namesDeclaredType returns whether the given type reference names one of the declared types or their array types.
func (f *dependencyFinder) namesDeclaredType(ref tree.ResolvableTypeReference) bool {
	switch ref := ref.(type) {
	case *tree.ArrayTypeReference:
		return f.namesDeclaredType(ref.ElementType)
	case *tree.ModifiedTypeReference:
		return f.namesDeclaredType(ref.Name)
	case *tree.UnresolvedObjectName:
		return slices.ContainsFunc(f.ext.Types, func(declared extdef.Type) bool {
			return declared.Name == ref.Parts[0]
		})
	}
	return false
}
