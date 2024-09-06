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

package information_schema

import (
	"github.com/dolthub/go-mysql-server/sql/information_schema"
)

// Init handles initialization of all Postgres-specific and Doltgres-specific information_schema tables.
func Init() {
	information_schema.NewColumnsTable = newColumnsTable
	information_schema.NewSchemataTable = newSchemataTable
	information_schema.AllDatabasesWithNames = allDatabasesWithNames
}
