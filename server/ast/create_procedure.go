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
	"github.com/cockroachdb/errors"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeCreateProcedure handles *tree.CreateProcedure nodes.
func nodeCreateProcedure(ctx *Context, node *tree.CreateProcedure) (vitess.Statement, error) {
	_, err := validateRoutineOptions(ctx, node.Options)
	if err != nil {
		return nil, err
	}

	return NotYetSupportedError("CREATE PROCEDURE statement is not yet supported")
}

// validateRoutineOptions ensures that each option is defined only once. Returns a map containing all options, or an
// error if an option is invalid or is defined multiple times.
func validateRoutineOptions(ctx *Context, options []tree.RoutineOption) (map[tree.FunctionOption]tree.RoutineOption, error) {
	var optDefined = make(map[tree.FunctionOption]tree.RoutineOption)
	for _, opt := range options {
		if _, ok := optDefined[opt.OptionType]; ok {
			return nil, errors.Errorf("ERROR:  conflicting or redundant options")
		} else {
			optDefined[opt.OptionType] = opt
		}
	}
	return optDefined, nil
}
