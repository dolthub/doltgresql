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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/vitess/go/sqltypes"
)

// IDs are basically arbitrary, we just need to ensure that they do not conflict with existing IDs
// Comments are to match the Stringer formatting rules in the original rule definition file, but we can't generate
// human-readable strings for these extended types because they are in another package.
const (
	ruleId_TypeSanitizer                analyzer.RuleId = iota + 1000 // typeSanitizer
	ruleId_AddDomainConstraints                                       // addDomainConstraints
	ruleId_AddDomainConstraintsToCasts                                // addDomainConstraintsToCasts
	ruleId_AssignInsertCasts                                          // assignInsertCasts
	ruleId_AssignUpdateCasts                                          // assignUpdateCasts
	ruleId_ReplaceIndexedTables                                       // replaceIndexedTables
	ruleId_ReplaceNode                                                // replaceNode
	ruleId_ReplaceSerial                                              // replaceSerial
	ruleId_AddImplicitPrefixLengths                                   // addImplicitPrefixLengths
	ruleId_InsertContextRootFinalizer                                 // insertContextRootFinalizer
	ruleId_ResolveType                                                // resolveType
	ruleId_ReplaceArithmeticExpressions                               // replaceArithmeticExpressions
	ruleId_OptimizeFunctions                                          // optimizeFunctions
	ruleId_ValidateColumnDefaults                                     // validateColumnDefaults
	ruleId_ValidateCreateTable                                        // validateCreateTable
	rulesId_ResolveAlterColumn                                        // resolveAlterColumn
)

// Init adds additional rules to the analyzer to handle Doltgres-specific functionality.
func Init() {
	analyzer.AlwaysBeforeDefault = append(analyzer.AlwaysBeforeDefault,
		analyzer.Rule{Id: ruleId_ResolveType, Apply: ResolveType},
		analyzer.Rule{Id: ruleId_TypeSanitizer, Apply: TypeSanitizer},
		analyzer.Rule{Id: ruleId_AddDomainConstraints, Apply: AddDomainConstraints},
		analyzer.Rule{Id: ruleId_ValidateColumnDefaults, Apply: ValidateColumnDefaults},
		analyzer.Rule{Id: ruleId_AssignInsertCasts, Apply: AssignInsertCasts},
		analyzer.Rule{Id: ruleId_AssignUpdateCasts, Apply: AssignUpdateCasts},
		analyzer.Rule{Id: ruleId_ReplaceIndexedTables, Apply: ReplaceIndexedTables},
	)

	// PostgreSQL doesn't have the concept of prefix lengths, so we add a rule to implicitly add them
	// TODO: this should be replaced by implementing automatic toast semantics for blob types
	analyzer.OnceBeforeDefault = append([]analyzer.Rule{{Id: ruleId_AddImplicitPrefixLengths, Apply: AddImplicitPrefixLengths}},
		analyzer.OnceBeforeDefault...)

	// We remove several validation rules and substitute our own
	analyzer.OnceBeforeDefault = insertAnalyzerRules(analyzer.OnceBeforeDefault, analyzer.ValidateCreateTableId, true,
		analyzer.Rule{Id: ruleId_ValidateCreateTable, Apply: validateCreateTable})
	analyzer.OnceBeforeDefault = insertAnalyzerRules(analyzer.OnceBeforeDefault, analyzer.ResolveAlterColumnId, true,
		analyzer.Rule{Id: rulesId_ResolveAlterColumn, Apply: resolveAlterColumn})

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
	plan.ValidateForeignKeyDefinition = validateForeignKeyDefinition
}

// validateForeignKeyDefinition validates that the given foreign key definition is valid for creation
func validateForeignKeyDefinition(ctx *sql.Context, fkDef sql.ForeignKeyConstraint, cols map[string]*sql.Column, parentCols map[string]*sql.Column) error {
	// TODO: this check is too permissive, we should be doing some type checks here
	for i := range fkDef.Columns {
		col := cols[strings.ToLower(fkDef.Columns[i])]
		parentCol := parentCols[strings.ToLower(fkDef.ParentColumns[i])]
		if !foreignKeyComparableTypes(ctx, col.Type, parentCol.Type) {
			return sql.ErrForeignKeyColumnTypeMismatch.New(fkDef.Columns[i], fkDef.ParentColumns[i])
		}
	}
	return nil
}

// foreignKeyComparableTypes returns whether the two given types are able to be used as parent/child columns in a
// foreign key.
func foreignKeyComparableTypes(ctx *sql.Context, type1 sql.Type, type2 sql.Type) bool {
	if !type1.Equals(type2) {
		// There seems to be a special case where CHAR/VARCHAR/BINARY/VARBINARY can have unequal lengths.
		// Have not tested every type nor combination, but this seems specific to those 4 types.
		if type1.Type() == type2.Type() {
			switch type1.Type() {
			case sqltypes.Char, sqltypes.VarChar, sqltypes.Binary, sqltypes.VarBinary:
				type1String := type1.(sql.StringType)
				type2String := type2.(sql.StringType)
				if type1String.Collation().CharacterSet() != type2String.Collation().CharacterSet() {
					return false
				}
			default:
				return false
			}
		} else {
			return false
		}
	}
	return true
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
