// Copyright 2023 Dolthub, Inc.
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
	"fmt"
	"strings"
)

// Scanner allows for scanning through the synopsis to generate tokens.
type Scanner struct {
	source     []rune
	tokens     []Token
	sourceIdx  int
	isFinished bool
	savedErr   error
}

// NewScanner returns a new *Scanner by reading through the synopsis.
func NewScanner(synopsis string) *Scanner {
	scanner := &Scanner{
		source:     []rune(strings.ReplaceAll(synopsis, "\r", "")),
		tokens:     nil,
		sourceIdx:  -1,
		isFinished: false,
		savedErr:   nil,
	}
	_, _ = scanner.Process()
	return scanner
}

// Next returns the next rune while advancing the scanner. Returns false if there are no more runes.
func (scanner *Scanner) Next() (rune, bool) {
	if scanner.sourceIdx+1 >= len(scanner.source) {
		return 0, false
	}
	scanner.sourceIdx++
	return scanner.source[scanner.sourceIdx], true
}

// Peek returns the next rune. Does not advance the scanner. Returns false if there are no more runes.
func (scanner *Scanner) Peek() (rune, bool) {
	if scanner.sourceIdx+1 >= len(scanner.source) {
		return 0, false
	}
	return scanner.source[scanner.sourceIdx+1], true
}

// PeekBy returns the next rune that is n positions from the current rune. Does not advance the scanner. Returns false
// if we are peeking beyond the source.
func (scanner *Scanner) PeekBy(n int) (rune, bool) {
	if scanner.sourceIdx+n >= len(scanner.source) || scanner.sourceIdx+n < 0 {
		return 0, false
	}
	return scanner.source[scanner.sourceIdx+n], true
}

// PeekMatch returns whether the given string exactly matches runes starting after the current position. This is
// equivalent to calling PeekMatchOffset with an offset of 1.
func (scanner *Scanner) PeekMatch(str string) bool {
	return scanner.PeekMatchOffset(str, 1)
}

// PeekMatchOffset returns whether the given string exactly matches runes starting from the offset applied to the
// current position. An offset of 0 means that we are including the rune at the current position.
func (scanner *Scanner) PeekMatchOffset(str string, offset int) bool {
	for i, r := range []rune(str) {
		sr, ok := scanner.PeekBy(i + offset)
		if !ok || sr != r {
			return false
		}
	}
	return true
}

// Advance is equivalent to calling AdvanceBy(1).
func (scanner *Scanner) Advance() {
	scanner.AdvanceBy(1)
}

// AdvanceBy advances the scanner by the given amount. If the amount is greater than the number of remaining runes, then
// it advances to the end. Cannot advance backwards.
func (scanner *Scanner) AdvanceBy(n int) {
	if n < 0 {
		n = 0
	}
	if scanner.sourceIdx+n >= len(scanner.source) {
		n = len(scanner.source) - scanner.sourceIdx - 1
	}
	scanner.sourceIdx += n
}

// Process processes the synopsis and returns the generated tokens.
func (scanner *Scanner) Process() ([]Token, error) {
	if scanner.isFinished || scanner.savedErr != nil {
		return scanner.tokens, scanner.savedErr
	}
ScannerLoop:
	for {
		r, ok := scanner.Next()
		if !ok {
			break ScannerLoop
		}
		switch r {
		case '$':
			var name []rune
			for r, ok = scanner.Peek(); ok && r != '$'; r, ok = scanner.Peek() {
				scanner.Advance()
				name = append(name, r)
			}
			if !ok {
				scanner.savedErr = fmt.Errorf("unexpected EOF when reading variable name")
				break ScannerLoop
			}
			scanner.Advance()
			scanner.tokens = append(scanner.tokens, Token{
				Type:    TokenType_Variable,
				Literal: string(name),
			})
		case '\t':
			scanner.savedErr = fmt.Errorf("tab found, remove all tabs from the synopsis")
			break ScannerLoop
		case ' ', '\n':
			spaceCount := 0
			lineCount := 0
			if r == ' ' {
				spaceCount += 1
			}
			if r == '\n' {
				lineCount++
			}
			commentMode := false
		WhitespaceLoop:
			for r, ok = scanner.Peek(); ok; r, ok = scanner.Peek() {
				if commentMode {
					scanner.Advance()
					if r == '\n' {
						commentMode = false
					}
				} else {
					switch r {
					case ' ':
						scanner.Advance()
						spaceCount += 1
					case '\n':
						scanner.Advance()
						lineCount++
					case '/':
						if next, _ := scanner.PeekBy(2); next == '/' {
							if last, _ := scanner.PeekBy(0); last == '\n' {
								scanner.Advance()
								commentMode = true
								continue WhitespaceLoop
							}
						}
						break WhitespaceLoop
					default:
						break WhitespaceLoop
					}
				}
			}
			// EOF, no need to add the last bit of whitespace
			if !ok {
				break ScannerLoop
			}
			if spaceCount == 1 && lineCount == 0 {
				scanner.tokens = append(scanner.tokens, Token{Type: TokenType_ShortSpace})
			} else if spaceCount > 1 && lineCount <= 1 {
				scanner.tokens = append(scanner.tokens, Token{Type: TokenType_MediumSpace})
			} else {
				// This will match the case where spaceCount == 0 and lineCount == 1.
				// This may seem counter-intuitive, however it's consistent with how new statements are defined.
				scanner.tokens = append(scanner.tokens, Token{Type: TokenType_LongSpace})
			}
		case '.':
			dotCount := 1
			// All dot repetition blocks look like `...`
		DotRepetitionLoop:
			for n := 1; true; n++ {
				peek, _ := scanner.PeekBy(n)
				switch peek {
				case '.':
					dotCount++
				default:
					scanner.AdvanceBy(n - 1)
					// If the dot count is different from 3, then we'll treat it as a string
					if dotCount == 3 {
						scanner.tokens = append(scanner.tokens, Token{Type: TokenType_Repeat})
					} else {
						scanner.tokens = append(scanner.tokens, Token{
							Type:    TokenType_Text,
							Literal: strings.Repeat(".", dotCount),
						})
					}
					break DotRepetitionLoop
				}
			}
		case ',':
			scanner.tokens = append(scanner.tokens, Token{
				Type:    TokenType_Text,
				Literal: ",",
			})
		case '[':
			if scanner.PeekMatchOffset("[ ... ]", 0) {
				scanner.AdvanceBy(6)
				scanner.tokens = append(scanner.tokens, Token{
					Type:    TokenType_OptionalRepeat,
					Literal: "",
				})
			} else if scanner.PeekMatchOffset("[ , ... ]", 0) {
				scanner.AdvanceBy(8)
				scanner.tokens = append(scanner.tokens, Token{
					Type:    TokenType_OptionalRepeat,
					Literal: ",",
				})
			} else if scanner.PeekMatchOffset("[ AND ... ]", 0) {
				scanner.AdvanceBy(10)
				scanner.tokens = append(scanner.tokens, Token{
					Type:    TokenType_OptionalRepeat,
					Literal: "AND",
				})
			} else if scanner.PeekMatchOffset("[ OR ... ]", 0) {
				scanner.AdvanceBy(9)
				scanner.tokens = append(scanner.tokens, Token{
					Type:    TokenType_OptionalRepeat,
					Literal: "OR",
				})
			} else {
				scanner.tokens = append(scanner.tokens, Token{Type: TokenType_OptionalOpen})
			}
		case ']':
			scanner.tokens = append(scanner.tokens, Token{Type: TokenType_OptionalClose})
		case '{':
			scanner.tokens = append(scanner.tokens, Token{Type: TokenType_OneOfOpen})
		case '}':
			scanner.tokens = append(scanner.tokens, Token{Type: TokenType_OneOfClose})
		case '(':
			scanner.tokens = append(scanner.tokens, Token{Type: TokenType_ParenOpen})
		case ')':
			scanner.tokens = append(scanner.tokens, Token{Type: TokenType_ParenClose})
		case '|':
			scanner.tokens = append(scanner.tokens, Token{Type: TokenType_Or})
		default:
			if scanner.PeekMatchOffset("where $", 0) {
				// Skip past what we peeked at
				scanner.AdvanceBy(6)
				var varName []rune
			VariableDescriptionLoop:
				for r, ok = scanner.Peek(); ok; r, ok = scanner.Peek() {
					switch r {
					case '$':
						scanner.tokens = append(scanner.tokens, Token{
							Type:    TokenType_VariableDefinition,
							Literal: string(varName),
						})
						if !scanner.PeekMatch("$ is:") {
							scanner.savedErr = fmt.Errorf("invalid variable definition format")
							break ScannerLoop
						}
						// Skip past what we peeked at
						scanner.AdvanceBy(5)
						break VariableDescriptionLoop
					default:
						scanner.Advance()
						varName = append(varName, r)
					}
				}
			} else {
				text := []rune{r}
			TextLoop:
				for r, ok = scanner.Peek(); ok; r, ok = scanner.Peek() {
					switch r {
					case '$', ' ', '\t', '\n', '.', ',', '[', ']', '{', '}', '(', ')', '|':
						scanner.tokens = append(scanner.tokens, Token{
							Type:    TokenType_Text,
							Literal: string(text),
						})
						break TextLoop
					default:
						scanner.Advance()
						text = append(text, r)
					}
				}
				// If we hit EOF, then we haven't added the word yet, so we'll do it here
				if !ok {
					scanner.tokens = append(scanner.tokens, Token{
						Type:    TokenType_Text,
						Literal: string(text),
					})
				}
			}
		}
	}
	// Remove any ending spaces
	if len(scanner.tokens) > 0 && scanner.tokens[len(scanner.tokens)-1].IsSpace() {
		scanner.tokens = scanner.tokens[:len(scanner.tokens)-1]
	}
	// Add an EOF
	scanner.tokens = append(scanner.tokens, Token{Type: TokenType_EOF})
	// Set that we're finished now (we'll also finish on errors)
	scanner.isFinished = true
	return scanner.tokens, scanner.savedErr
}

// String returns the processed contents of the scanner as a string. This means that it will not be the original string,
// as the processed string may differ in format.
func (scanner *Scanner) String() string {
	sb := strings.Builder{}
	for _, token := range scanner.tokens {
		str := token.String()
		if len(str) > 0 {
			sb.WriteString(str)
		}
	}
	return sb.String()
}
