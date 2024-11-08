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

package ast

import (
	"fmt"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeCreateFunction handles *tree.CreateFunction nodes.
func nodeCreateFunction(ctx *Context, node *tree.CreateFunction) (vitess.Statement, error) {
	err := verifyRedundantRoutineOption(ctx, node.Options)
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("CREATE FUNCTION statement is not yet supported")
}

// verifyRedundantRoutineOption checks for each option defined only once.
// If there is multiple definition of the same option, it returns an error.
func verifyRedundantRoutineOption(ctx *Context, options []tree.RoutineOption) error {
	var optDefined = make(map[tree.FunctionOption]struct{})
	for _, opt := range options {
		if _, ok := optDefined[opt.OptionType]; ok {
			return fmt.Errorf("ERROR:  conflicting or redundant options")
		} else {
			optDefined[opt.OptionType] = struct{}{}
		}
	}
	return nil
}
