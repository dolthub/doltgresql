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

package oid

import (
	"fmt"
	"strings"
	"unicode"
)

// ioInputSections converts the input string for IoInput into a sectioned form according to the rules defined by Postgres:
// https://www.postgresql.org/docs/15/datatype-oid.html
func ioInputSections(input string) ([]string, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("invalid name syntax")
	}
	runeInput := []rune(strings.TrimSpace(input))
	var sections []string
	var sectionBuilder strings.Builder
	var inQuotes bool
	for i := 0; i < len(runeInput); i++ {
		char := runeInput[i]
		switch char {
		case '"':
			if inQuotes {
				if i < len(runeInput)-1 && runeInput[i+1] == '"' {
					sectionBuilder.WriteRune(char)
					i++
				} else {
					inQuotes = false
					section := sectionBuilder.String()
					sectionBuilder.Reset()
					if len(section) == 0 || section == `"` {
						return nil, fmt.Errorf("invalid name syntax")
					}
					sections = append(sections, section)
				}
			} else {
				inQuotes = true
			}
		default:
			if inQuotes {
				sectionBuilder.WriteRune(char)
			} else if char != ' ' {
				if char == '.' {
					section := sectionBuilder.String()
					sectionBuilder.Reset()
					if len(section) > 0 {
						sections = append(sections, section)
					}
					sections = append(sections, string(char))
				} else {
					sectionBuilder.WriteRune(unicode.ToLower(char))
				}
			}
		}
	}
	if sectionBuilder.Len() > 0 {
		section := sectionBuilder.String()
		if inQuotes {
			// For some reason, you can have an unmatched double quote at the end, so we're duplicating that behavior
			if input[len(input)-1] != '"' {
				return nil, fmt.Errorf("invalid name syntax")
			}
			section += `"`
		}
		sections = append(sections, section)
	}
	return sections, nil
}
