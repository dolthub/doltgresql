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
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// information_schema columns are one of these 5 types https://www.postgresql.org/docs/current/infoschema-datatypes.html
var cardinal_number = pgtypes.Int32
var character_data = pgtypes.Text
var sql_identifier = pgtypes.NewVarCharType(64)
var yes_or_no = pgtypes.NewVarCharType(3)
