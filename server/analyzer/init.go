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
	"fmt"

	"github.com/dolthub/go-mysql-server/sql/analyzer"
)

// IDs are basically arbitrary, we just need to ensure that they do not conflict with existing IDs
const (
	ruleId_TypeSanitizer analyzer.RuleId = iota + 1000
	ruleId_AddDomainConstraints
	ruleId_AssignInsertCasts
	ruleId_AssignUpdateCasts
	ruleId_ReplaceIndexedTables
	ruleId_ReplaceSerial
	ruleId_ReplaceDropTable
	ruleId_AddImplicitPrefixLengths
	ruleId_InsertContextRootFinalizer
	ruleId_ResolveType
	ruleId_ReplaceArithmeticExpressions
)

// Init adds additional rules to the analyzer to handle Doltgres-specific functionality.
func Init() {
	analyzer.AlwaysBeforeDefault = append(analyzer.AlwaysBeforeDefault,
		analyzer.Rule{Id: ruleId_ResolveType, Apply: ResolveType},
		analyzer.Rule{Id: ruleId_TypeSanitizer, Apply: TypeSanitizer},
		analyzer.Rule{Id: ruleId_AddDomainConstraints, Apply: AddDomainConstraints},
		getAnalyzerRule(analyzer.OnceBeforeDefault, analyzer.ValidateColumnDefaultsId),
		analyzer.Rule{Id: ruleId_AssignInsertCasts, Apply: AssignInsertCasts},
		analyzer.Rule{Id: ruleId_AssignUpdateCasts, Apply: AssignUpdateCasts},
		analyzer.Rule{Id: ruleId_ReplaceIndexedTables, Apply: ReplaceIndexedTables},
	)

	// Column default validation was moved to occur after type sanitization, so we'll remove it from its original place
	analyzer.OnceBeforeDefault = removeAnalyzerRules(analyzer.OnceBeforeDefault, analyzer.ValidateColumnDefaultsId)

	// PostgreSQL doesn't have the concept of prefix lengths, so we add a rule to implicitly add them
	analyzer.OnceBeforeDefault = append([]analyzer.Rule{{Id: ruleId_AddImplicitPrefixLengths, Apply: AddImplicitPrefixLengths}},
		analyzer.OnceBeforeDefault...)

	// Remove all other validation rules that do not apply to Postgres
	analyzer.DefaultValidationRules = removeAnalyzerRules(analyzer.DefaultValidationRules, analyzer.ValidateOperandsId)

	analyzer.OnceAfterDefault = append(analyzer.OnceAfterDefault,
		analyzer.Rule{Id: ruleId_ReplaceSerial, Apply: ReplaceSerial},
		analyzer.Rule{Id: ruleId_ReplaceDropTable, Apply: ReplaceDropTable},
		analyzer.Rule{Id: ruleId_ReplaceArithmeticExpressions, Apply: ReplaceArithmeticExpressions},
	)

	// The auto-commit rule writes the contents of the context, so we need to insert our finalizer before that
	analyzer.OnceAfterAll = insertAnalyzerRules(analyzer.OnceAfterAll, analyzer.BacktickDefaulColumnValueNamesId, false,
		analyzer.Rule{Id: ruleId_InsertContextRootFinalizer, Apply: InsertContextRootFinalizer})
}

// getAnalyzerRule returns the rule matching the given ID.
func getAnalyzerRule(rules []analyzer.Rule, id analyzer.RuleId) analyzer.Rule {
	for _, rule := range rules {
		if rule.Id == id {
			return rule
		}
	}
	// This will only occur if GMS has been changed
	panic(fmt.Errorf("rule not found: %d", id))
}

// insertAnalyzerRules inserts the given rule(s) before or after the given analyzer.RuleId, returning an updated slice.
func insertAnalyzerRules(rules []analyzer.Rule, id analyzer.RuleId, before bool, additionalRules ...analyzer.Rule) []analyzer.Rule {
	newRules := make([]analyzer.Rule, len(rules)+len(additionalRules))
	for i, rule := range rules {
		if rule.Id == id {
			if before {
				copy(newRules, analyzer.OnceAfterAll[:i])
				copy(newRules[i:], additionalRules)
				copy(newRules[i+len(additionalRules):], analyzer.OnceAfterAll[i:])
			} else {
				copy(newRules, analyzer.OnceAfterAll[:i+1])
				copy(newRules[i+1:], additionalRules)
				copy(newRules[i+1+len(additionalRules):], analyzer.OnceAfterAll[i+1:])
			}
			break
		}
	}
	return newRules
}

// removeAnalyzerRules removes the given analyzer.RuleId(s), returning an updated slice.
func removeAnalyzerRules(rules []analyzer.Rule, remove ...analyzer.RuleId) []analyzer.Rule {
	ids := make(map[analyzer.RuleId]struct{})
	for _, removal := range remove {
		ids[removal] = struct{}{}
	}
	newRules := make([]analyzer.Rule, 0, len(rules))
	for _, rule := range rules {
		if _, ok := ids[rule.Id]; !ok {
			newRules = append(newRules, rule)
		}
	}
	return newRules
}
