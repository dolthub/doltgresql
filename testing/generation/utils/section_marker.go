// Copyright 2023-2024 Dolthub, Inc.
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

package utils

import (
	"fmt"
	"strings"
)

// SectionMarker returns a marker that may be used to denote sections.
//
// For example, SectionMarker("abc", '-', 21) would return:
//
// -------- abc --------
func SectionMarker(centeredText string, fillerCharacter rune, totalLength int) string {
	fillerStr := string(fillerCharacter)
	remainingLength := totalLength - (len(centeredText) + 2)
	if remainingLength <= 0 {
		return fmt.Sprintf(" %s ", centeredText)
	}
	left := remainingLength / 2
	right := remainingLength - left // Integer division doesn't do fractions, so this will handle odd counts
	return fmt.Sprintf("%s %s %s",
		strings.Repeat(fillerStr, left), centeredText, strings.Repeat(fillerStr, right))
}

// PrintSectionMarker is a convenience function that prints the result of SectionMarker with a newline appended.
func PrintSectionMarker(centeredText string, fillerCharacter rune, totalLength int) {
	fmt.Println(SectionMarker(centeredText, fillerCharacter, totalLength))
}
