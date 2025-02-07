// Copyright 2025 Dolthub, Inc.
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

// FunctionProvider is the special sql.FunctionProvider for Doltgres that allows us to handle functions that were
// are created by users.
type FunctionProvider struct{}

var _ sql.FunctionProvider = (*FunctionProvider)(nil)

// Function implements the interface sql.FunctionProvider.
func (fp *FunctionProvider) Function(ctx *sql.Context, name string) (sql.Function, bool) {
	// TODO: this should be configurable from within Dolt, rather than set on an external variable
	// TODO: user functions should be accessible from the context, just like how sequences and types are handled
	//  For now, this just reads our global map (which also needs to be changed, since functions should not be global)
	if f, ok := compiledCatalog[name]; ok {
		return sql.FunctionN{
			Name: name,
			Fn:   f,
		}, true
	}
	return nil, false
}
