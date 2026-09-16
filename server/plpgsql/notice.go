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

package plpgsql

import (
	"strconv"
	"strings"
)

// NoticeLevel represents the severity, or level, of a notice created by a RAISE statement.
type NoticeLevel uint8

const (
	NoticeLevelDebug     NoticeLevel = 14
	NoticeLevelLog       NoticeLevel = 15
	NoticeLevelInfo      NoticeLevel = 17
	NoticeLevelNotice    NoticeLevel = 18
	NoticeLevelWarning   NoticeLevel = 19
	NoticeLevelException NoticeLevel = 21
)

// String returns a string representation of this NoticeLevel.
func (nl NoticeLevel) String() string {
	switch nl {
	case NoticeLevelDebug:
		return "DEBUG"
	case NoticeLevelLog:
		return "LOG"
	case NoticeLevelInfo:
		return "INFO"
	case NoticeLevelNotice:
		return "NOTICE"
	case NoticeLevelWarning:
		return "WARNING"
	case NoticeLevelException:
		return "EXCEPTION"
	default:
		return "UNKNOWN"
	}
}

// NoticeOptionType represents the type of option specified for a notice in the USING clause of a RAISE statement.
type NoticeOptionType uint8

const (
	NoticeOptionTypeErrCode    NoticeOptionType = 0
	NoticeOptionTypeMessage    NoticeOptionType = 1
	NoticeOptionTypeDetail     NoticeOptionType = 2
	NoticeOptionTypeHint       NoticeOptionType = 3
	NoticeOptionTypeConstraint NoticeOptionType = 5
	NoticeOptionTypeDataType   NoticeOptionType = 6
	NoticeOptionTypeTable      NoticeOptionType = 7
	NoticeOptionTypeSchema     NoticeOptionType = 8
)

// errCodeOptionKey is the Options key under which an operation carries the SQLSTATE that a RAISE reports.
// It is the same key the USING clause's ERRCODE option lands under, so a generated RAISE and one written in
// a function body name their code the same way.
var errCodeOptionKey = strconv.Itoa(int(NoticeOptionTypeErrCode))

// sqlStateFromErrCode resolves the value of a RAISE statement's ERRCODE option into the SQLSTATE it names,
// reporting whether it named one. The value arrives as the raw source text of the USING clause's expression
// when the RAISE was written in a function body, so a quoted literal is unwrapped here.
//
// TODO: PostgreSQL also accepts a condition name, such as division_by_zero, which needs the table mapping
// every name to its SQLSTATE. Until that exists, a name reports as unresolved and the RAISE keeps its
// default code, rather than reporting the name itself as though it were a code.
func sqlStateFromErrCode(value string) (string, bool) {
	sqlState := strings.TrimSpace(value)
	if len(sqlState) >= 2 && sqlState[0] == '\'' && sqlState[len(sqlState)-1] == '\'' {
		sqlState = sqlState[1 : len(sqlState)-1]
	}
	// A SQLSTATE is five characters drawn from the digits and the upper-case letters.
	if len(sqlState) != 5 {
		return "", false
	}
	for _, r := range sqlState {
		if (r < '0' || r > '9') && (r < 'A' || r > 'Z') {
			return "", false
		}
	}
	return sqlState, true
}
