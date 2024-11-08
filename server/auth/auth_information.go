// Copyright 2024 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

// These AuthType_ enums are used as the AuthType in vitess.AuthInformation.
const (
	AuthType_IGNORE       = "IGNORE"
	AuthType_ALTER_SYSTEM = "ALTER_SYSTEM"
	AuthType_CONNECT      = "CONNECT"
	AuthType_CREATE       = "CREATE"
	AuthType_DELETE       = "DELETE"
	AuthType_EXECUTE      = "EXECUTE"
	AuthType_INSERT       = "INSERT"
	AuthType_REFERENCES   = "REFERENCES"
	AuthType_SELECT       = "SELECT"
	AuthType_SET          = "SET"
	AuthType_TEMPORARY    = "TEMPORARY"
	AuthType_TRIGGER      = "TRIGGER"
	AuthType_TRUNCATE     = "TRUNCATE"
	AuthType_UPDATE       = "UPDATE"
	AuthType_USAGE        = "USAGE"
)

// These AuthTargetType_ enums are used as the TargetType in vitess.AuthInformation.
const (
	AuthTargetType_Ignore                   = "IGNORE"
	AuthTargetType_DatabaseIdentifiers      = "DB_IDENTS"
	AuthTargetType_Global                   = "GLOBAL"
	AuthTargetType_MultipleTableIdentifiers = "DB_TABLE_IDENTS"
	AuthTargetType_SingleTableIdentifier    = "DB_TABLE_IDENT"
	AuthTargetType_TableColumn              = "DB_TABLE_COLUMN_IDENT"
	AuthTargetType_TODO                     = "TODO"
)
