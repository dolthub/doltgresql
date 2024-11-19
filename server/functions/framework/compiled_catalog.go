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

import "github.com/dolthub/go-mysql-server/sql"

// compiledCatalog contains all of PostgreSQL functions in their compiled forms.
var compiledCatalog = map[string]sql.CreateFuncNArgs{}

// GetFunction returns the compiled function with the given name and parameters. Returns false if the function could not
// be found.
func GetFunction(functionName string, params ...sql.Expression) (*CompiledFunction, bool, error) {
	if createFunc, ok := compiledCatalog[functionName]; ok {
		expr, err := createFunc(params...)
		if err != nil {
			return nil, false, err
		}
		return expr.(*CompiledFunction), true, nil
	}
	return nil, false, nil
}
