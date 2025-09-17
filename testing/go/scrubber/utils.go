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

import "math/rand/v2"

// replacements contains all word replacements, to ensure that similar sequences throughout the file are replaced with
// the same values. This allows for foreign keys to continue functioning.
var replacements = make(map[string]string)

// createdReplacements contains all replacement strings that have been created, in an effort to prevent unintended
// duplicates (which can cause primary/unique keys to unexpectedly fail).
var createdReplacements = make(map[string]struct{})

// replacementCharacters contains all of the characters that are valid substitutions for replacement.
var replacementCharacters = []rune{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z'}

// replacementCharactersLimited contains all of the characters that are valid substitutions in a limited context (such as hex-encoded strings).
var replacementCharactersLimited = []rune{'a', 'b', 'c', 'd', 'e', 'f', '0', '1', '2', '3', '4', '5', '6', '7', '8', '9'}

// getReplacement returns the replacement of the given string if applicable.
func getReplacement(sequence []rune, useLimited bool) string {
	original := string(sequence)
	if len(sequence) == 0 || isSpecialRune(sequence[0]) || isReserved(original) {
		return original
	}
	replacement, ok := replacements[original]
	if ok {
		return replacement
	}
	allNumbers := true
	for i := 0; i < len(sequence); i++ {
		if isNumber(sequence[i]) { // We don't replace numbers
			continue
		} else {
			allNumbers = false
			if useLimited {
				sequence[i] = replacementCharactersLimited[rand.Uint32()%uint32(len(replacementCharactersLimited))]
			} else {
				sequence[i] = replacementCharacters[rand.Uint32()%uint32(len(replacementCharacters))]
			}
		}
	}
	if allNumbers {
		return original
	}
	newStr := string(sequence)
	if _, ok = createdReplacements[newStr]; (ok || isReserved(newStr)) && !useLimited {
		// If we've somehow randomly created a reserved or duplicate word, then we recursively try again
		return getReplacement([]rune(original), useLimited)
	}
	replacements[original] = newStr
	createdReplacements[newStr] = struct{}{}
	return newStr
}

// handleUUIDs exists as UUIDs must use a limited set of replacement runes since not all runes are valid.
func handleUUIDs(sequences [][]rune) {
	for i := range sequences {
		// Ensure that there are enough sequences
		if i+8 >= len(sequences) {
			continue
		}
		// Check that the sequence lengths are valid
		if len(sequences[i]) != 8 ||
			len(sequences[i+1]) != 1 ||
			len(sequences[i+2]) != 4 ||
			len(sequences[i+3]) != 1 ||
			len(sequences[i+4]) != 4 ||
			len(sequences[i+5]) != 1 ||
			len(sequences[i+6]) != 4 ||
			len(sequences[i+7]) != 1 ||
			len(sequences[i+8]) != 12 {
			continue
		}
		// Check for the dashes that separate UUIDs
		if sequences[i+1][0] != '-' ||
			sequences[i+3][0] != '-' ||
			sequences[i+5][0] != '-' ||
			sequences[i+7][0] != '-' {
			continue
		}
		// Check that the characters are all alphanumeric
		if !matches(sequences[i], isAlphanumeric) ||
			!matches(sequences[i+2], isAlphanumeric) ||
			!matches(sequences[i+4], isAlphanumeric) ||
			!matches(sequences[i+6], isAlphanumeric) ||
			!matches(sequences[i+8], isAlphanumeric) {
			continue
		}
		var newStr string
		for seqIdx, seq := range [][]rune{sequences[i], sequences[i+2], sequences[i+4], sequences[i+6], sequences[i+8]} {
			if seqIdx > 0 {
				newStr += "-"
			}
			newStr += getReplacement(seq, true)
		}
		// Set the UUID as a replacement
		sequences[i] = []rune(newStr)
		sequences[i+1] = nil
		sequences[i+2] = nil
		sequences[i+3] = nil
		sequences[i+4] = nil
		sequences[i+5] = nil
		sequences[i+6] = nil
		sequences[i+7] = nil
		sequences[i+8] = nil
		replacements[newStr] = newStr
	}
}

// isSpecialRune returns whether the given rune is a control character or symbol (excluding underscores).
func isSpecialRune(r rune) bool {
	return (r >= 0 && r <= '/') || (r >= ':' && r <= '@') || (r >= '[' && r <= '^') || r == '`' || (r >= '{' && r <= 191)
}

// isNumber returns whether the given rune is a number.
func isNumber(r rune) bool {
	return r >= '0' && r <= '9'
}

// isCharacter returns whether the given rune is a character.
func isCharacter(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}

// isAlphanumeric returns whether the given rune is a numeber or character.
func isAlphanumeric(r rune) bool {
	return isCharacter(r) || isNumber(r)
}

// isControlCharacter returns whether the current runeCache and the given nextRune creates a control character (such as
// the null character \N).
func isControlCharacter(runeCache []rune, nextRune rune) bool {
	if len(runeCache) == 0 {
		return false
	}
	if runeCache[len(runeCache)-1] == '\\' {
		switch nextRune {
		case 'b', 'f', 'n', 'r', 't', 'v', 'x', 'N':
			return true
		}
	}
	return false
}

// matches returns whether each rune in the sequence returns true when passed to the given function.
func matches(sequence []rune, f func(rune) bool) bool {
	for _, r := range sequence {
		if !f(r) {
			return false
		}
	}
	return true
}
