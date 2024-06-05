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

package framework

import (
	"github.com/dolthub/go-mysql-server/sql"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// IntermediateFunction is an expression that represents an incomplete PostgreSQL function.
type IntermediateFunction struct {
	Functions    *OverloadDeduction
	AllOverloads [][]pgtypes.DoltgresTypeBaseID
	IsOperator   bool
}

// Compile returns a CompiledFunction created from the calling IntermediateFunction. Returns a nil function if it could
// not be compiled.
func (f IntermediateFunction) Compile(name string, parameters ...sql.Expression) *CompiledFunction {
	if f.Functions == nil {
		return nil
	}
	return &CompiledFunction{
		Name:         name,
		Parameters:   parameters,
		Functions:    f.Functions,
		AllOverloads: f.AllOverloads,
		IsOperator:   f.IsOperator,
	}
}
