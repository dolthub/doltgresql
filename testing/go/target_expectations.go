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

package _go

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

type serverTestTarget string

const (
	targetDoltgres     serverTestTarget = "doltgres"
	targetPostgres     serverTestTarget = "postgres"
	targetPostgresFork serverTestTarget = "postgres-fork"
)

// An explicit target selects expectations independently of how the server starts.
// Never infer compatibility from a query result or silently accept either answer.
func parseTestTarget(value string) (serverTestTarget, error) {
	switch serverTestTarget(value) {
	case "", targetDoltgres:
		return targetDoltgres, nil
	case targetPostgres:
		return targetPostgres, nil
	case targetPostgresFork:
		return targetPostgresFork, nil
	default:
		return "", fmt.Errorf("invalid DOLTGRES_TEST_TARGET %q (want doltgres, postgres, or postgres-fork)", value)
	}
}

func testTarget(t testing.TB) serverTestTarget {
	t.Helper()
	target, err := parseTestTarget(os.Getenv("DOLTGRES_TEST_TARGET"))
	if err != nil {
		t.Fatal(err)
	}
	return target
}

func assertionName(a ScriptTestAssertion) string {
	if a.variant == "" {
		return a.Query
	}
	return a.Query + " [" + a.variant + "]"
}

func expandAssertionsForTarget(assertions []ScriptTestAssertion, target serverTestTarget) []ScriptTestAssertion {
	var result []ScriptTestAssertion
	for _, a := range assertions {
		if a.Postgres == nil && a.PostgresFork == nil {
			result = append(result, a)
			continue
		}
		original := a
		original.Postgres, original.PostgresFork = nil, nil
		original.variant, original.variantEnabled = "doltgres", target == targetDoltgres
		result = append(result, original)
		for _, v := range []struct {
			name      string
			assertion *ScriptTestAssertion
			enabled   bool
		}{
			{"postgres", a.Postgres, target == targetPostgres || target == targetPostgresFork && a.PostgresFork == nil},
			{"postgres-fork", a.PostgresFork, target == targetPostgresFork},
		} {
			if v.assertion == nil {
				continue
			}
			alternate := *v.assertion
			if alternate.Query == "" {
				alternate.Query = a.Query
			}
			// A pre-existing unconditional skip must not accidentally enable an unimplemented case.
			alternate.Skip = alternate.Skip || a.Skip
			alternate.Focus = a.Focus
			alternate.Postgres, alternate.PostgresFork = nil, nil
			alternate.variant, alternate.variantEnabled = v.name, v.enabled
			result = append(result, alternate)
		}
	}
	return result
}

func expandTargetAssertions(t testing.TB, assertions []ScriptTestAssertion) []ScriptTestAssertion {
	t.Helper()
	return expandAssertionsForTarget(assertions, testTarget(t))
}

// runTargetPair gives hand-written wire tests the same visible skip contract.
func runTargetPair(t *testing.T, test func(*testing.T, bool)) {
	target := testTarget(t)
	for _, name := range []string{"doltgres", "postgres"} {
		t.Run(name, func(t *testing.T) {
			postgres := name == "postgres"
			if postgres != (target != targetDoltgres) {
				t.Skip("counterpart selected by DOLTGRES_TEST_TARGET")
			}
			test(t, postgres)
		})
	}
}

func checkPlanRelations(plan string, relations []string) error {
	var roots []struct{ Plan map[string]any }
	if err := json.Unmarshal([]byte(plan), &roots); err != nil {
		return err
	}
	if len(roots) != 1 || roots[0].Plan["Node Type"] == nil {
		return fmt.Errorf("missing native plan root")
	}
	found := map[string]bool{}
	var walk func(map[string]any)
	walk = func(node map[string]any) {
		if name, ok := node["Relation Name"].(string); ok {
			found[name] = true
		}
		if plans, ok := node["Plans"].([]any); ok {
			for _, p := range plans {
				if child, ok := p.(map[string]any); ok {
					walk(child)
				}
			}
		}
	}
	walk(roots[0].Plan)
	for _, relation := range relations {
		if !found[relation] {
			return fmt.Errorf("plan does not reference expected relation %s", relation)
		}
	}
	return nil
}

func checkCatalogScanAliases(plan string, aliases []string) error {
	if len(aliases) == 0 {
		return nil
	}
	var roots []struct{ Plan map[string]any }
	if err := json.Unmarshal([]byte(plan), &roots); err != nil {
		return err
	}
	found := map[string]bool{}
	var walk func(map[string]any)
	walk = func(node map[string]any) {
		if node["Node Type"] == "Custom Scan" && node["Custom Plan Provider"] == "DoltCatalogScan" {
			if alias, ok := node["Alias"].(string); ok {
				found[alias] = true
			}
		}
		if plans, ok := node["Plans"].([]any); ok {
			for _, p := range plans {
				if child, ok := p.(map[string]any); ok {
					walk(child)
				}
			}
		}
	}
	for _, root := range roots {
		walk(root.Plan)
	}
	for _, alias := range aliases {
		if !found[alias] {
			return fmt.Errorf("missing DoltCatalogScan for alias %s", alias)
		}
	}
	return nil
}
