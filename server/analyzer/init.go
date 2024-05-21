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
)

// Init adds additional rules to the analyzer to handle Doltgres-specific functionality.
func Init() {
	// IDs are basically arbitrary, we just need to ensure that they do not conflict with existing IDs
	analyzer.OnceAfterDefault = append(analyzer.OnceAfterDefault,
		analyzer.Rule{Id: 1000, Apply: ReplaceSerial},
	)
	newOnceAfterAll := make([]analyzer.Rule, len(analyzer.OnceAfterAll)+1)
	for i, onceAfterAllRule := range analyzer.OnceAfterAll {
		// The auto-commit rule writes the contents of the context, so we need to insert our finalizer before that
		if onceAfterAllRule.Id == analyzer.AutocommitId {
			copy(newOnceAfterAll, analyzer.OnceAfterAll[:i])
			newOnceAfterAll[i] = analyzer.Rule{Id: 2000, Apply: InsertContextRootFinalizer}
			copy(newOnceAfterAll[i+1:], analyzer.OnceAfterAll[i:])
		}
	}
	analyzer.OnceAfterAll = newOnceAfterAll
}
