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

package core

import (
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/resolve"
	"github.com/dolthub/go-mysql-server/sql"
)

// GetCurrentSchema returns the current schema used by the context. Defaults to "public" if the context does not specify
// a schema.
func GetCurrentSchema(ctx *sql.Context) (string, error) {
	_, root, err := getRootFromContext(ctx)
	if err != nil {
		return "", nil
	}

	return resolve.FirstExistingSchemaOnSearchPath(ctx, root)
}
