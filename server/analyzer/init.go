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
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/planbuilder"

	"github.com/dolthub/doltgresql/server/index"

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
	ruleId_ReplaceSerial                                                 // replaceSerial
	ruleId_InsertContextRootFinalizer                                    // insertContextRootFinalizer
	ruleId_ResolveType                                                   // resolveType
	ruleId_ReplaceArithmeticExpressions                                  // replaceArithmeticExpressions
	ruleId_OptimizeFunctions                                             // optimizeFunctions
	ruleId_ValidateColumnDefaults                                        // validateColumnDefaults
	ruleId_ValidateCreateTable                                           // validateCreateTable
	ruleId_ResolveAlterColumn                                            // resolveAlterColumn
)

// Init adds additional rules to the analyzer to handle Doltgres-specific functionality.
func Init() {
	analyzer.AlwaysBeforeDefault = append(analyzer.AlwaysBeforeDefault,
		analyzer.Rule{Id: ruleId_ResolveType, Apply: ResolveType},
		analyzer.Rule{Id: ruleId_TypeSanitizer, Apply: TypeSanitizer},
		analyzer.Rule{Id: ruleId_GenerateForeignKeyName, Apply: generateForeignKeyName},
		analyzer.Rule{Id: ruleId_AddDomainConstraints, Apply: AddDomainConstraints},
		analyzer.Rule{Id: ruleId_ValidateColumnDefaults, Apply: ValidateColumnDefaults},
		analyzer.Rule{Id: ruleId_AssignInsertCasts, Apply: AssignInsertCasts},
		analyzer.Rule{Id: ruleId_AssignUpdateCasts, Apply: AssignUpdateCasts},
		analyzer.Rule{Id: ruleId_AssignTriggers, Apply: AssignTriggers},
		analyzer.Rule{Id: ruleId_ReplaceIndexedTables, Apply: ReplaceIndexedTables},
	)

	analyzer.OnceBeforeDefault = append([]analyzer.Rule{
		{Id: ruleId_ApplyTablesForAnalyzeAllTables, Apply: applyTablesForAnalyzeAllTables},
		{Id: ruleId_ConvertDropPrimaryKeyConstraint, Apply: convertDropPrimaryKeyConstraint}},
		analyzer.OnceBeforeDefault...)

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

	analyzer.OnceAfterDefault = append(analyzer.OnceAfterDefault,
		analyzer.Rule{Id: ruleId_ReplaceSerial, Apply: ReplaceSerial},
		analyzer.Rule{Id: ruleId_ReplaceArithmeticExpressions, Apply: ReplaceArithmeticExpressions},
	)

	// The auto-commit rule writes the contents of the context, so we need to insert our finalizer before that.
	// We also should optimize functions last, since other rules may change the underlying expressions, potentially changing their return types.
	analyzer.OnceAfterAll = insertAnalyzerRules(analyzer.OnceAfterAll, analyzer.QuoteDefaultColumnValueNamesId, false,
		analyzer.Rule{Id: ruleId_OptimizeFunctions, Apply: OptimizeFunctions},
		// AddDomainConstraintsToCasts needs to run after 'assignExecIndexes' rule in GMS.
		analyzer.Rule{Id: ruleId_AddDomainConstraintsToCasts, Apply: AddDomainConstraintsToCasts},
		analyzer.Rule{Id: ruleId_ReplaceNode, Apply: ReplaceNode},
		analyzer.Rule{Id: ruleId_InsertContextRootFinalizer, Apply: InsertContextRootFinalizer})

	initEngine()
}

func initEngine() {
	// This technically takes place at execution time rather than as part of analysis, but we don't have a better
	// place to put it. Our foreign key validation logic is different from MySQL's, and since it's not an analyzer rule
	// we can't swap out a rule like the rest of the logic in this package, we have to do a function swap.
	plan.ValidateForeignKeyDefinition = validateForeignKeyDefinition

	planbuilder.IsAggregateFunc = IsAggregateFunc

	expression.DefaultExpressionFactory = pgexpression.PostgresExpressionFactory{}

	analyzer.SplitConjunction = index.SplitConjunction
}

// IsAggregateFunc checks if the given function name is an aggregate function. This is the entire set supported by
// MySQL plus some postgres specific ones.
func IsAggregateFunc(name string) bool {
	if planbuilder.IsMySQLAggregateFuncName(name) {
		return true
	}

	switch name {
	case "array_agg", "bool_and", "bool_or":
		return true
	}

	return false
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
