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
