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

package id

import (
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
)

// GetFromTable returns the id from the table.
func GetFromTable(ctx *sql.Context, tbl sql.Table) (Table, bool, error) {
	schTbl, ok := tbl.(sql.DatabaseSchemaTable)
	if !ok {
		return NullTable, false, errors.Newf(`table "%s" does not specify a schema`, tbl.Name())
	}
	return NewTable(schTbl.DatabaseSchema().SchemaName(), schTbl.Name()), true, nil
}
