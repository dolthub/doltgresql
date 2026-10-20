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
	"fmt"
	"slices"
	"strings"

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

// CreateObjects writes every object that the given extension declares into the given schema, running the statements
// that create its tables and views on the given runner.
func CreateObjects(ctx *sql.Context, runner sql.StatementRunner, ext *extdef.Extension, schemaName string) error {
	typColl, err := core.GetTypesCollectionFromContext(ctx, "")
	if err != nil {
		return err
	}
	return extensionObjects{ext: ext, schemaName: schemaName, typColl: typColl, runner: runner}.materialize(ctx)
}

// UpdateObjects rewrites every object that the given extension declares in the given schema to match the declaration,
// creating the objects that are missing and removing the extension's functions that it no longer declares. A declared
// table that already exists keeps its rows, and the statements that write tables and views run on the given runner.
func UpdateObjects(ctx *sql.Context, runner sql.StatementRunner, ext *extdef.Extension, schemaName string) error {
	typColl, err := core.GetTypesCollectionFromContext(ctx, "")
	if err != nil {
		return err
	}
	objects := extensionObjects{ext: ext, schemaName: schemaName, typColl: typColl, runner: runner, replace: true}
	if err = objects.materialize(ctx); err != nil {
		return err
	}
	return objects.dropUndeclaredRoutines(ctx)
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

// DropObjects removes every object that the given extension declares from the given schema, running the statements
// that drop its tables and views on the given runner.
func DropObjects(ctx *sql.Context, runner sql.StatementRunner, ext *extdef.Extension, schemaName string) error {
	typColl, err := core.GetTypesCollectionFromContext(ctx, "")
	if err != nil {
		return err
	}
	return extensionObjects{ext: ext, schemaName: schemaName, typColl: typColl, runner: runner}.drop(ctx)
}

// extensionObjects materializes the objects that an extension declares.
type extensionObjects struct {
	ext        *extdef.Extension
	schemaName string
	typColl    *typecollection.TypeCollection
	runner     sql.StatementRunner
	// replace is whether materializing removes an existing object before writing the declared one.
	replace bool
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
	if err := e.materializeAggregates(ctx); err != nil {
		return err
	}
	if err := e.createTables(ctx); err != nil {
		return err
	}
	return e.createViews(ctx)
}

// createTables creates the declared tables and inserts their rows, leaving a table that already exists alone when
// replacing.
func (e extensionObjects) createTables(ctx *sql.Context) error {
	for _, table := range e.ext.Tables {
		if e.replace {
			existing, err := core.GetSqlTableFromContext(ctx, "", doltdb.TableName{Name: table.Name, Schema: e.schemaName})
			if err != nil {
				return err
			}
			if existing != nil {
				continue
			}
		}
		name := e.qualifiedName(table.Name)
		if err := hook.RunStatement(ctx, e.runner, "CREATE TABLE "+name+" "+table.Definition); err != nil {
			return err
		}
		rows, err := table.Rows(ctx)
		if err != nil {
			return err
		}
		for start := 0; start < len(rows); start += insertBatchSize {
			var sb strings.Builder
			sb.WriteString("INSERT INTO " + name + " VALUES ")
			for i, row := range rows[start:min(start+insertBatchSize, len(rows))] {
				if i > 0 {
					sb.WriteString(", ")
				}
				sb.WriteString("(")
				for j, val := range row {
					if j > 0 {
						sb.WriteString(", ")
					}
					sb.WriteString(sqlLiteral(val))
				}
				sb.WriteString(")")
			}
			if err = hook.RunStatement(ctx, e.runner, sb.String()); err != nil {
				return err
			}
		}
	}
	return nil
}

// insertBatchSize is the number of rows that each INSERT statement of a declared table's rows holds.
const insertBatchSize = 500

// sqlLiteral returns the SQL literal for a value of a declared table's row, which is nil, an integer, or a string.
func sqlLiteral(val any) string {
	switch val := val.(type) {
	case nil:
		return "NULL"
	case string:
		return "'" + strings.ReplaceAll(val, "'", "''") + "'"
	default:
		return fmt.Sprint(val)
	}
}

// createViews creates the declared views, replacing a view that already exists when replacing.
func (e extensionObjects) createViews(ctx *sql.Context) error {
	for _, view := range e.ext.Views {
		if e.replace {
			if err := hook.RunStatement(ctx, e.runner, "DROP VIEW IF EXISTS "+e.qualifiedName(view.Name)); err != nil {
				return err
			}
		}
		if err := hook.RunStatement(ctx, e.runner, "CREATE VIEW "+e.qualifiedName(view.Name)+" AS "+view.Query); err != nil {
			return err
		}
	}
	return nil
}

// qualifiedName returns the quoted name of a relation in the extension's schema.
func (e extensionObjects) qualifiedName(name string) string {
	return hook.QuoteIdentifier(e.schemaName) + "." + hook.QuoteIdentifier(name)
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
		if e.replace && e.typColl.HasType(ctx, newType.ID) {
			if err = e.typColl.DropType(ctx, newType.ID, id.NewType(e.schemaName, "_"+declared.Name)); err != nil {
				return err
			}
		}
		if err = e.typColl.CreateType(ctx, newType); err != nil {
			return err
		}
		if err = e.typColl.CreateType(ctx, pgtypes.CreateArrayTypeFromBaseType(newType)); err != nil {
			return err
		}
	}
	for _, declared := range e.ext.Composites {
		attrs := make([]pgtypes.CompositeAttribute, len(declared.Attributes))
		for i, attr := range declared.Attributes {
			attrTypeID, err := e.typeID(ctx, attr.Type)
			if err != nil {
				return err
			}
			attrType, err := e.typColl.GetType(ctx, attrTypeID)
			if err != nil {
				return err
			}
			attrs[i] = pgtypes.NewCompositeAttribute(ctx, id.Null, attr.Name, attrType, int16(i+1), "")
		}
		typeID := id.NewType(e.schemaName, declared.Name)
		arrayType := &pgtypes.DoltgresType{ID: id.NewType(e.schemaName, "_"+declared.Name), IsUnresolved: true}
		newType := pgtypes.NewCompositeType(ctx, id.Null, arrayType, typeID, attrs)
		if e.replace && e.typColl.HasType(ctx, typeID) {
			if err := e.typColl.DropType(ctx, typeID, arrayType.ID); err != nil {
				return err
			}
		}
		if err := e.typColl.CreateType(ctx, newType); err != nil {
			return err
		}
		if err := e.typColl.CreateType(ctx, pgtypes.CreateArrayTypeFromBaseType(newType)); err != nil {
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
		paramTypes, err := e.parameterTypes(ctx, routine.Parameters)
		if err != nil {
			return err
		}
		allParams := make([]procedures.Parameter, len(routine.Parameters))
		for i, param := range routine.Parameters {
			allParams[i] = procedures.Parameter{Name: param.Name, Type: paramTypes[i], Default: param.Default}
		}
		outputTypes, err := e.parameterTypes(ctx, routine.Outputs)
		if err != nil {
			return err
		}
		returnType, err := e.returnType(ctx, routine, outputTypes)
		if err != nil {
			return err
		}
		for i, output := range routine.Outputs {
			allParams = append(allParams, procedures.Parameter{Mode: procedures.ParameterMode_OUT, Name: output.Name, Type: outputTypes[i]})
		}
		volatility := routine.Volatility
		if volatility == 0 {
			volatility = tree.VolatilityVolatile
		}
		funcID := id.NewFunction(e.schemaName, routine.Name, paramTypes...)
		if e.replace && funcCollection.HasFunction(ctx, funcID) {
			if err = funcCollection.DropFunction(ctx, funcID); err != nil {
				return err
			}
		}
		err = funcCollection.AddFunction(ctx, functions.Function{
			ID:                 funcID,
			ReturnType:         returnType,
			AllParams:          allParams,
			IsNonDeterministic: true,
			Strict:             routine.Strict,
			SetOf:              routine.SetOf,
			Window:             routine.WindowImpl != nil,
			Volatility:         volatility,
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
		operatorID := id.NewOperator(e.schemaName, declared.Symbol, leftType, rightType)
		if e.replace && opCollection.HasOperator(ctx, operatorID) {
			if err = opCollection.DropOperator(ctx, operatorID); err != nil {
				return err
			}
		}
		err = opCollection.AddOperator(ctx, operators.Operator{
			ID:         operatorID,
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
		castID := id.NewCast(sourceType, targetType)
		if e.replace && castCollection.HasCast(ctx, castID) {
			if err = castCollection.DropCast(ctx, castID); err != nil {
				return err
			}
		}
		err = castCollection.AddCast(ctx, casts.Cast{
			ID:       castID,
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
		aggregateID := id.NewFunction(e.schemaName, declared.Name, paramTypes...)
		if e.replace && aggCollection.HasAggregate(ctx, aggregateID) {
			if err = aggCollection.DropAggregate(ctx, aggregateID); err != nil {
				return err
			}
		}
		err = aggCollection.AddAggregate(ctx, aggregates.Aggregate{
			ID:          aggregateID,
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
	for _, view := range e.ext.Views {
		if err := hook.RunStatement(ctx, e.runner, "DROP VIEW IF EXISTS "+e.qualifiedName(view.Name)); err != nil {
			return err
		}
	}
	for _, table := range e.ext.Tables {
		if err := hook.RunStatement(ctx, e.runner, "DROP TABLE IF EXISTS "+e.qualifiedName(table.Name)); err != nil {
			return err
		}
	}
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
			if schema.SchemaName() == e.schemaName && slices.ContainsFunc(e.ext.Tables, func(table extdef.Table) bool {
				return strings.EqualFold(table.Name, tableName)
			}) {
				continue
			}
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
		ownView := func(schemaName string, viewName string) bool {
			return schemaName == e.schemaName && slices.ContainsFunc(e.ext.Views, func(view extdef.View) bool {
				return strings.EqualFold(view.Name, viewName)
			})
		}
		if err = hook.WalkViewExpressions(ctx, &finder, ownView); err != nil {
			return err
		}
	}
	if !finder.found && len(e.ext.Tables)+len(e.ext.Views) > 0 {
		var relations []doltdb.TableName
		for _, table := range e.ext.Tables {
			relations = append(relations, doltdb.TableName{Name: table.Name, Schema: e.schemaName})
		}
		for _, view := range e.ext.Views {
			relations = append(relations, doltdb.TableName{Name: view.Name, Schema: e.schemaName})
		}
		if finder.found, err = hook.DependentViewExists(ctx, relations); err != nil {
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
	var aggregateIDs []id.Function
	for _, declared := range e.ext.Aggregates {
		paramTypes, err := e.parameterTypes(ctx, declared.Parameters)
		if err != nil {
			return err
		}
		if aggregateID := id.NewFunction(e.schemaName, declared.Name, paramTypes...); aggCollection.HasAggregate(ctx, aggregateID) {
			aggregateIDs = append(aggregateIDs, aggregateID)
		}
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
	var castIDs []id.Cast
	for _, declared := range e.ext.Casts {
		sourceType, err := e.typeID(ctx, declared.Source)
		if err != nil {
			return err
		}
		targetType, err := e.typeID(ctx, declared.Target)
		if err != nil {
			return err
		}
		if castID := id.NewCast(sourceType, targetType); castCollection.HasCast(ctx, castID) {
			castIDs = append(castIDs, castID)
		}
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
	var operatorIDs []id.Operator
	for _, declared := range e.ext.Operators {
		leftType, err := e.typeID(ctx, declared.Left)
		if err != nil {
			return err
		}
		rightType, err := e.typeID(ctx, declared.Right)
		if err != nil {
			return err
		}
		if operatorID := id.NewOperator(e.schemaName, declared.Symbol, leftType, rightType); opCollection.HasOperator(ctx, operatorID) {
			operatorIDs = append(operatorIDs, operatorID)
		}
	}
	return opCollection.DropOperator(ctx, operatorIDs...)
}

// dropRoutines removes every function that belongs to the extension from the functions collection, including the
// functions that an earlier declaration created.
func (e extensionObjects) dropRoutines(ctx *sql.Context) error {
	return e.dropRoutinesExcept(ctx, nil)
}

// dropUndeclaredRoutines removes the functions that belong to the extension but that its declaration no longer has.
func (e extensionObjects) dropUndeclaredRoutines(ctx *sql.Context) error {
	declared := make(map[id.Function]struct{}, len(e.ext.Routines))
	for _, routine := range e.ext.Routines {
		funcID, err := e.routineID(ctx, routine)
		if err != nil {
			return err
		}
		declared[funcID] = struct{}{}
	}
	return e.dropRoutinesExcept(ctx, declared)
}

// dropRoutinesExcept removes the functions that belong to the extension in its schema, other than the given ones.
func (e extensionObjects) dropRoutinesExcept(ctx *sql.Context, kept map[id.Function]struct{}) error {
	funcCollection, err := core.GetFunctionsCollectionFromContext(ctx, "")
	if err != nil {
		return err
	}
	var funcIDs []id.Function
	err = funcCollection.IterateFunctions(ctx, func(f functions.Function) (stop bool, err error) {
		if _, ok := kept[f.ID]; !ok && f.ExtensionName == e.ext.Name && f.ID.SchemaName() == e.schemaName {
			funcIDs = append(funcIDs, f.ID)
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	return funcCollection.DropFunction(ctx, funcIDs...)
}

// dropTypes removes the declared types and their array types from the types collection.
func (e extensionObjects) dropTypes(ctx *sql.Context) error {
	if len(e.ext.Types) == 0 && len(e.ext.Composites) == 0 {
		return nil
	}
	var typeIDs []id.Type
	for _, typeID := range e.declaredTypeIDs() {
		if e.typColl.HasType(ctx, typeID) {
			typeIDs = append(typeIDs, typeID)
		}
	}
	return e.typColl.DropType(ctx, typeIDs...)
}

// declaredTypeIDs returns the IDs of the declared types and their array types.
func (e extensionObjects) declaredTypeIDs() []id.Type {
	typeIDs := make([]id.Type, 0, (len(e.ext.Types)+len(e.ext.Composites))*2)
	for _, declared := range e.ext.Types {
		typeIDs = append(typeIDs, id.NewType(e.schemaName, declared.Name), id.NewType(e.schemaName, "_"+declared.Name))
	}
	for _, declared := range e.ext.Composites {
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
	for _, declared := range e.ext.Composites {
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

// returnType returns the type ID that a routine returns, given the types of its outputs. A routine with outputs
// returns the anonymous composite type of its outputs, as a function declared with RETURNS TABLE does.
func (e extensionObjects) returnType(ctx *sql.Context, routine extdef.Routine, outputTypes []id.Type) (id.Type, error) {
	if len(routine.Outputs) == 0 {
		return e.typeID(ctx, routine.Returns)
	}
	attributes := make([]string, len(routine.Outputs))
	for i, output := range routine.Outputs {
		attributes[i] = output.Name + ":" + outputTypes[i].SchemaName() + "." + outputTypes[i].TypeName()
	}
	return id.NewType("", "table("+strings.Join(attributes, ",")+")"), nil
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
