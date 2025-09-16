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

package main

import (
	"bytes"
	"fmt"
	"os"
)

// limitedSequences contains indices for sequences that should use the limited character set.
var limitedSequences = make(map[int]struct{})

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Expected two arguments, the first containing the SQL file, and the second containing the output file.")
		os.Exit(1)
	}
	sqlFileBytes, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	var sequences [][]rune            // This holds all sequences, where a sequence is a run of special characters or non-special characters
	lastReadSpecial := false          // When this is true, the last rune that was read was a special rune
	runeCache := make([]rune, 0, 512) // The rune cache holds all of the runes read since the last special (or normal) rune

	// This loop gathers all sequences
	for _, r := range string(sqlFileBytes) {
		isSpecial := isSpecialRune(r)
		if isSpecial != lastReadSpecial && len(runeCache) > 0 {
			// We must preserve control characters, which will always start with a special rune (preserving them)
			if isControlCharacter(runeCache, r) {
				sequence := make([]rune, len(runeCache)+1)
				copy(sequence, runeCache)
				sequence[len(sequence)-1] = r
				sequences = append(sequences, sequence)
				runeCache = runeCache[:0]
				lastReadSpecial = true
				// The sequences following a hex control character will use a limited character set for replacements
				if r == 'x' {
					limitedSequences[len(sequences)] = struct{}{}
				}
				continue
			}
			sequence := make([]rune, len(runeCache))
			copy(sequence, runeCache)
			sequences = append(sequences, sequence)
			runeCache = runeCache[:0]
		}
		runeCache = append(runeCache, r)
		lastReadSpecial = isSpecial
	}

	// Look for UUIDs to ensure that they're handled correctly
	handleUUIDs(sequences)

	// We write all sequences to the buffer, inserting the replacements when applicable
	buffer := bytes.Buffer{}
	buffer.Grow(int(float64(len(sqlFileBytes)) * 1.5))
	for seqIdx, sequence := range sequences {
		_, ok := limitedSequences[seqIdx]
		buffer.WriteString(getReplacement(sequence, ok))
	}

	// Now we write the buffer to the output file
	err = os.WriteFile(os.Args[2], buffer.Bytes(), 0644)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
