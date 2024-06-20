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

package pgcatalog

// PgCatalogName is a constant to the pg_catalog name.
const PgCatalogName = "pg_catalog"

// Init initializes everything necessary for the pg_catalog tables.
func Init() {
	InitPgAggregate()
	InitPgAm()
	InitPgAmop()
	InitPgAmproc()
	InitPgAttrdef()
	InitPgAttribute()
	InitPgAuthMembers()
	InitPgAuthid()
	InitPgCast()
	InitPgClass()
	InitPgCollation()
	InitPgConstraint()
	InitPgConversion()
	InitPgDatabase()
	InitPgDbRoleSetting()
	InitPgDefaultAcl()
	InitPgDepend()
	InitPgDescription()
	InitPgEnum()
	InitPgEventTrigger()
	InitPgExtension()
	InitPgForeignDataWrapper()
	InitPgForeignServer()
	InitPgForeignTable()
	InitPgIndex()
	InitPgNamespace()
	InitPgProc()
	InitPgSequence()
	InitPgTrigger()
	InitPgType()
}
