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

package cast

import (
	"fmt"
	"strings"
	"unicode/utf8"

	pgtypes "github.com/dolthub/doltgresql/server/types"
	"gopkg.in/src-d/go-errors.v1"
)

var (
	// errCannotCast is returned when a cast from one type to another is not possible.
	errCannotCast = errors.NewKind("cannot cast type %s to %s")

	// errOutOfRange is returned when a value is out of range for a given type.
	errOutOfRange = errors.NewKind("%s out of range")
)

// handleCharExplicitCast handles explicit casts to Char and VarChar types. Returns an error if other types are passed in.
func handleCharExplicitCast(str string, targetType pgtypes.DoltgresType) (string, error) {
	switch targetType := targetType.(type) {
	case pgtypes.CharType:
		if targetType.IsUnbounded() {
			return str, nil
		}
		str, runeCount := truncateString(str, targetType.Length)
		if runeCount < targetType.Length {
			return str + strings.Repeat(" ", int(targetType.Length-runeCount)), nil
		}
		return str, nil
	case pgtypes.NameType:
		str, _ = truncateString(str, targetType.Length)
		return str, nil
	case pgtypes.VarCharType:
		if targetType.IsUnbounded() {
			return str, nil
		}
		str, _ = truncateString(str, targetType.Length)
		return str, nil
	default:
		return "", fmt.Errorf("explicit cast called to handle non-char type")
	}
}

// handleCharImplicitCast handles implicit casts to Char and VarChar types. Returns an error if other types are passed in.
func handleCharImplicitCast(str string, targetType pgtypes.DoltgresType) (string, error) {
	switch targetType := targetType.(type) {
	case pgtypes.CharType:
		if targetType.IsUnbounded() {
			return str, nil
		} else {
			runeLength := uint32(utf8.RuneCountInString(str))
			if runeLength > targetType.Length {
				return "", fmt.Errorf("value too long for type %s", targetType.String())
			} else if runeLength < targetType.Length {
				return str + strings.Repeat(" ", int(targetType.Length-runeLength)), nil
			} else {
				return str, nil
			}
		}
	case pgtypes.NameType:
		if uint32(utf8.RuneCountInString(str)) > targetType.Length {
			return "", fmt.Errorf("value too long for type %s", targetType.String())
		} else {
			return str, nil
		}
	case pgtypes.VarCharType:
		if !targetType.IsUnbounded() && uint32(utf8.RuneCountInString(str)) > targetType.Length {
			return "", fmt.Errorf("value too long for type %s", targetType.String())
		} else {
			return str, nil
		}
	default:
		return "", fmt.Errorf("implicit cast called to handle non-char type")
	}
}

// truncateString returns a string that has been truncated to the given length. Uses the rune count rather than the
// byte count. Returns the input string if it's smaller than the length. Also returns the rune count of the string.
func truncateString(val string, runeLimit uint32) (string, uint32) {
	runeLength := uint32(utf8.RuneCountInString(val))
	if runeLength > runeLimit {
		// TODO: figure out if there's a faster way to truncate based on rune count
		startString := val
		for i := uint32(0); i < runeLimit; i++ {
			_, size := utf8.DecodeRuneInString(val)
			val = val[size:]
		}
		return startString[:len(startString)-len(val)], runeLimit
	}
	return val, runeLength
}
