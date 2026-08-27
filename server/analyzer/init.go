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

package analyzer

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/planbuilder"

	"github.com/dolthub/doltgresql/core"
	pgexpression "github.com/dolthub/doltgresql/server/expression"
)

// IDs are basically arbitrary, we just need to ensure that they do not conflict with existing IDs
// Comments are to match the Stringer formatting rules in the original rule definition file, but we can't generate
// human-readable strings for these extended types because they are in another package.
const (
	ruleId_TypeSanitizer                   analyzer.RuleId = iota + 1000 // typeSanitizer
	ruleId_AddDomainConstraints                                          // addDomainConstraints
	ruleId_AddDomainConstraintsToCasts                                   // addDomainConstraintsToCasts
	ruleId_ApplyTablesForAnalyzeAllTables                                // applyTablesForAnalyzeAllTables
	ruleId_AssignInsertCasts                                             // assignInsertCasts
	ruleId_AssignTriggers                                                // assignTriggers
	ruleId_AssignUpdateCasts                                             // assignUpdateCasts
	ruleId_ConvertDropPrimaryKeyConstraint                               // convertDropPrimaryKeyConstraint
	ruleId_GenerateForeignKeyName                                        // generateForeignKeyName
	ruleId_ReplaceIndexedTables                                          // replaceIndexedTables
	ruleId_ReplaceNode                                                   // replaceNode
	ruleId_TransformRecordFilter                                         // transformRecordFilter
	ruleId_ReplaceSerial                                                 // replaceSerial
	ruleId_InsertContextRootFinalizer                                    // insertContextRootFinalizer
	ruleId_ResolveType                                                   // resolveType
	ruleId_ReplaceArithmeticExpressions                                  // replaceArithmeticExpressions
	ruleId_OptimizeFunctions                                             // optimizeFunctions
	ruleId_ValidateColumnDefaults                                        // validateColumnDefaults
	ruleId_ValidateCreateTable                                           // validateCreateTable
	ruleId_ValidateCreateSchema                                          // validateCreateSchema
	ruleId_ResolveAlterColumn                                            // resolveAlterColumn
	ruleId_ValidateCreateFunction                                        // validateCreateFunction
	ruleId_ResolveValuesTypes                                            // resolveValuesTypes
	ruleId_ResolveProcedureDefaults                                      // resolveProcedureDefaults
	ruleId_SetRunner                                                     // setRunner
	ruleId_TypeSanitizeExistsSubquery                                    // typeSanitizeExistsSubquery
	ruleId_ResolveRenameIndex                                            // resolveRenameIndex
)

// Init adds additional rules to the analyzer to handle Doltgres-specific functionality.
func Init() {
	// OnceBeforeDefault runs before AlwaysBeforeDefault in GMS
	analyzer.OnceBeforeDefault = append([]analyzer.Rule{
		{Id: ruleId_ResolveType, Apply: ResolveType}, // ResolveType rule must run before simplifyFilters rule in GMS
		{Id: ruleId_ApplyTablesForAnalyzeAllTables, Apply: applyTablesForAnalyzeAllTables},
		{Id: ruleId_ConvertDropPrimaryKeyConstraint, Apply: convertDropPrimaryKeyConstraint},
		{Id: ruleId_ResolveRenameIndex, Apply: resolveRenameIndex}},
		analyzer.OnceBeforeDefault...)

	analyzer.AlwaysBeforeDefault = append(analyzer.AlwaysBeforeDefault,
		// ResolveType rule must run in this batch in addition to OnceBeforeDefault batch
		// because of custom batch set optimization in GMS skipping OnceBeforeDefault batch for some nodes.
		analyzer.Rule{Id: ruleId_ResolveType, Apply: ResolveType},
		analyzer.Rule{Id: ruleId_SetRunner, Apply: SetRunner},
		analyzer.Rule{Id: ruleId_TypeSanitizer, Apply: TypeSanitizer},
		analyzer.Rule{Id: ruleId_ResolveValuesTypes, Apply: ResolveValuesTypes},
		analyzer.Rule{Id: ruleId_GenerateForeignKeyName, Apply: generateForeignKeyName},
		analyzer.Rule{Id: ruleId_AddDomainConstraints, Apply: AddDomainConstraints},
		analyzer.Rule{Id: ruleId_ValidateColumnDefaults, Apply: ValidateColumnDefaults},
		analyzer.Rule{Id: ruleId_AssignInsertCasts, Apply: AssignInsertCasts},
		analyzer.Rule{Id: ruleId_AssignUpdateCasts, Apply: AssignUpdateCasts},
		analyzer.Rule{Id: ruleId_AssignTriggers, Apply: AssignTriggers},
		analyzer.Rule{Id: ruleId_ValidateCreateFunction, Apply: ValidateCreateFunction},
		analyzer.Rule{Id: ruleId_ValidateCreateSchema, Apply: ValidateCreateSchema},
		analyzer.Rule{Id: ruleId_ResolveProcedureDefaults, Apply: ResolveProcedureDefaults},
	)

	// We remove several validation rules and substitute our own
	analyzer.OnceBeforeDefault = insertAnalyzerRules(analyzer.OnceBeforeDefault, analyzer.ValidateCreateTableId, true,
		analyzer.Rule{Id: ruleId_ValidateCreateTable, Apply: validateCreateTable})
	analyzer.OnceBeforeDefault = insertAnalyzerRules(analyzer.OnceBeforeDefault, analyzer.ResolveAlterColumnId, true,
		analyzer.Rule{Id: ruleId_ResolveAlterColumn, Apply: resolveAlterColumn})

	analyzer.OnceBeforeDefault = removeAnalyzerRules(
		analyzer.OnceBeforeDefault,
		analyzer.ValidateColumnDefaultsId,
		analyzer.ValidateCreateTableId,
		analyzer.ResolveAlterColumnId,
	)

	// Remove all other validation rules that do not apply to Postgres
	analyzer.DefaultValidationRules = removeAnalyzerRules(analyzer.DefaultValidationRules, analyzer.ValidateOperandsId)

	analyzer.DefaultRules = append(analyzer.DefaultRules,
		analyzer.Rule{Id: ruleId_TransformRecordFilter, Apply: TransformRecordFilter},
	)

	analyzer.OnceAfterDefault = append(analyzer.OnceAfterDefault,
		analyzer.Rule{Id: ruleId_ReplaceSerial, Apply: ReplaceSerial},
		analyzer.Rule{Id: ruleId_ReplaceArithmeticExpressions, Apply: ReplaceArithmeticExpressions},
		// Must run after GMS's unnestExistsSubqueries rule, so decorrelation gets a chance to see
		// a bare *plan.ExistsSubquery before it's cast-wrapped.
		analyzer.Rule{Id: ruleId_TypeSanitizeExistsSubquery, Apply: TypeSanitizeExistsSubquery},
	)

	// The auto-commit rule writes the contents of the context, so we need to insert our finalizer before that.
	// We also should optimize functions last, since other rules may change the underlying expressions, potentially changing their return types.
	analyzer.OnceAfterAll = insertAnalyzerRules(analyzer.OnceAfterAll, analyzer.QuoteDefaultColumnValueNamesId, false,
		analyzer.Rule{Id: ruleId_OptimizeFunctions, Apply: OptimizeFunctions},
		// AddDomainConstraintsToCasts needs to run after 'assignExecIndexes' rule in GMS.
		analyzer.Rule{Id: ruleId_AddDomainConstraintsToCasts, Apply: AddDomainConstraintsToCasts},
		analyzer.Rule{Id: ruleId_ReplaceNode, Apply: ReplaceNode},
		analyzer.Rule{Id: ruleId_InsertContextRootFinalizer, Apply: InsertContextRootFinalizer},
	)

	initEngine()
}

// TODO: introduce a real pluggable architecture for this instead of swapping function pointers
func initEngine() {
	// This technically takes place at execution time rather than as part of analysis, but we don't have a better
	// place to put it. Our foreign key validation logic is different from MySQL's, and since it's not an analyzer rule
	// we can't swap out a rule like the rest of the logic in this package, we have to do a function swap.
	plan.ValidateForeignKeyDefinition = validateForeignKeyDefinition

	planbuilder.IsAggregateFunc = IsAggregateFunc
	planbuilder.IsWindowFunc = IsWindowFunc

	expression.DefaultExpressionFactory = pgexpression.PostgresExpressionFactory{}

	expression.SplitConjunction = splitConjunction
}

// postgresOnlyAggregateFuncNames holds Postgres aggregate functions with no MySQL equivalent. Every name
// here must be recognized by both IsAggregateFunc and IsWindowFunc: buildScalar in GMS's planbuilder only
// routes a call with an OVER(...) clause into the window-building path if IsWindowFunc recognizes its name,
// and Postgres allows any aggregate to be used as a window function.
var postgresOnlyAggregateFuncNames = map[string]bool{
	"array_agg": true,
	"bool_and":  true,
	"bool_or":   true,
}

// postgresOnlyWindowFuncNames holds Postgres functions that may only be used as window functions (i.e.
// within an OVER(...) clause) with no MySQL equivalent and no GROUP BY aggregate form.
var postgresOnlyWindowFuncNames = map[string]bool{
	"cume_dist": true,
	"nth_value": true,
}

// IsAggregateFunc checks if the given function name is an aggregate function. This is the entire set supported by
// MySQL plus some postgres specific ones, along with every user-defined aggregate.
func IsAggregateFunc(ctx *sql.Context, name string) (bool, error) {
	isAggregate, err := planbuilder.IsMySQLAggregateFuncName(ctx, name)
	if err != nil {
		return false, err
	}
	if isAggregate || postgresOnlyAggregateFuncNames[name] {
		return true, nil
	}
	collection, err := core.GetAggregatesCollectionFromContext(ctx, "")
	if err != nil {
		return false, err
	}
	return collection.HasAggregateName(ctx, name)
}

// IsWindowFunc checks if the given function name is a window function. This is the entire set supported by
// MySQL plus some postgres specific ones, along with every user-defined aggregate.
func IsWindowFunc(ctx *sql.Context, name string) (bool, error) {
	isWindow, err := planbuilder.IsMySQLWindowFuncName(ctx, name)
	if err != nil {
		return false, err
	}
	if isWindow || postgresOnlyAggregateFuncNames[name] || postgresOnlyWindowFuncNames[name] {
		return true, nil
	}
	collection, err := core.GetAggregatesCollectionFromContext(ctx, "")
	if err != nil {
		return false, err
	}
	return collection.HasAggregateName(ctx, name)
}

// insertAnalyzerRules inserts the given rule(s) before or after the given analyzer.RuleId, returning an updated slice.
func insertAnalyzerRules(rules []analyzer.Rule, id analyzer.RuleId, before bool, additionalRules ...analyzer.Rule) []analyzer.Rule {
	inserted := false
	newRules := make([]analyzer.Rule, len(rules)+len(additionalRules))
	for i, rule := range rules {
		if rule.Id == id {
			inserted = true
			if before {
				copy(newRules, rules[:i])
				copy(newRules[i:], additionalRules)
				copy(newRules[i+len(additionalRules):], rules[i:])
			} else {
				copy(newRules, rules[:i+1])
				copy(newRules[i+1:], additionalRules)
				copy(newRules[i+1+len(additionalRules):], rules[i+1:])
			}
			break
		}
	}

	if !inserted {
		panic("no rules were inserted")
	}

	return newRules
}

// removeAnalyzerRules removes the given analyzer.RuleId(s), returning an updated slice.
func removeAnalyzerRules(rules []analyzer.Rule, remove ...analyzer.RuleId) []analyzer.Rule {
	ids := make(map[analyzer.RuleId]struct{})
	for _, removal := range remove {
		ids[removal] = struct{}{}
	}

	removedIds := 0
	var newRules []analyzer.Rule
	for _, rule := range rules {
		if _, ok := ids[rule.Id]; !ok {
			newRules = append(newRules, rule)
		} else {
			removedIds++
		}
	}

	if removedIds < len(remove) {
		panic("one or more rules were not removed, this is a bug")
	}

	return newRules
}
