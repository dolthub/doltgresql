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

package config

import (
	"errors"
	"regexp"
	"time"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
)

// offsetRegex is a regex for matching MySQL offsets (e.g. +01:00).
var offsetRegex = regexp.MustCompile(`(?m)^([+\-])(\d{2}):(\d{2})$`)

// MySQLOffsetToDuration takes in a MySQL timezone offset (e.g. "+01:00") and returns it as a time.Duration.
// If any problems are encountered, an error is returned.
func MySQLOffsetToDuration(d string) (time.Duration, error) {
	matches := offsetRegex.FindStringSubmatch(d)
	if len(matches) == 4 {
		symbol := matches[1]
		hours := matches[2]
		mins := matches[3]
		return time.ParseDuration(symbol + hours + "h" + mins + "m")
	} else {
		return -1, errors.New("error: unable to process time")
	}
}

// AddNecessaryMySQLSystemVariables adds some of MySQL system variables as they are frequently used. E.g. 'autocommit'
// TODO: support MySQL system parameters to the extent that we have to, but we'll eventually move completely away from them.
func AddNecessaryMySQLSystemVariables() {
	sql.SystemVariables.AddSystemVariables([]sql.SystemVariable{
		// accessed before server starts
		&sql.MysqlSystemVariable{
			Name:              sql.AutoCommitSessionVar,
			Scope:             sql.SystemVariableScope_Both,
			Dynamic:           true,
			SetVarHintApplies: false,
			Type:              types.NewSystemBoolType(sql.AutoCommitSessionVar),
			Default:           int8(1),
		},
		&sql.MysqlSystemVariable{
			Name:              "max_connections",
			Scope:             sql.SystemVariableScope_Global,
			Dynamic:           true,
			SetVarHintApplies: false,
			Type:              types.NewSystemIntType("max_connections", 1, 100000, false),
			Default:           int64(151),
		},
		&sql.MysqlSystemVariable{
			Name:              "net_write_timeout",
			Scope:             sql.SystemVariableScope_Both,
			Dynamic:           true,
			SetVarHintApplies: false,
			Type:              types.NewSystemIntType("net_write_timeout", 1, 9223372036854775807, false),
			Default:           int64(60),
		},
		&sql.MysqlSystemVariable{
			Name:              "net_read_timeout",
			Scope:             sql.SystemVariableScope_Both,
			Dynamic:           true,
			SetVarHintApplies: false,
			Type:              types.NewSystemIntType("net_read_timeout", 1, 9223372036854775807, false),
			Default:           int64(60),
		},
		&sql.MysqlSystemVariable{
			Name:              "secure_file_priv",
			Scope:             sql.SystemVariableScope_Global,
			Dynamic:           false,
			SetVarHintApplies: false,
			Type:              types.NewSystemStringType("secure_file_priv"),
			Default:           "",
		},
		// accessed after? server starts
		&sql.MysqlSystemVariable{
			Name:              "foreign_key_checks",
			Scope:             sql.SystemVariableScope_Both,
			Dynamic:           true,
			SetVarHintApplies: true,
			Type:              types.NewSystemBoolType("foreign_key_checks"),
			Default:           int8(1),
		},
		&sql.MysqlSystemVariable{
			Name:              "sql_mode",
			Scope:             sql.SystemVariableScope_Both,
			Dynamic:           true,
			SetVarHintApplies: true,
			Type:              types.NewSystemSetType("sql_mode", "ALLOW_INVALID_DATES", "ANSI_QUOTES", "ERROR_FOR_DIVISION_BY_ZERO", "HIGH_NOT_PRECEDENCE", "IGNORE_SPACE", "NO_AUTO_VALUE_ON_ZERO", "NO_BACKSLASH_ESCAPES", "NO_DIR_IN_CREATE", "NO_ENGINE_SUBSTITUTION", "NO_UNSIGNED_SUBTRACTION", "NO_ZERO_DATE", "NO_ZERO_IN_DATE", "ONLY_FULL_GROUP_BY", "PAD_CHAR_TO_FULL_LENGTH", "PIPES_AS_CONCAT", "REAL_AS_FLOAT", "STRICT_ALL_TABLES", "STRICT_TRANS_TABLES", "TIME_TRUNCATE_FRACTIONAL", "TRADITIONAL", "ANSI"),
			Default:           "NO_ENGINE_SUBSTITUTION,ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES",
		},
	})
}
