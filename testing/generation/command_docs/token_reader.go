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

// TokenReader may be used to easily read through a collection of tokens generated from a synopsis. Of note, a
// TokenReader will skip all short and medium spaces, as clients of a reader will only care about statement boundaries.
type TokenReader struct {
	tokens []Token
	index  int
}

// NewTokenReader returns a new *TokenReader created from the given tokens.
func NewTokenReader(tokens []Token) *TokenReader {
	sanitizedTokens := make([]Token, 0, len(tokens))
	for _, token := range tokens {
		if token.IsStandardSpace() {
			continue
		}
		sanitizedTokens = append(sanitizedTokens, token)
	}
	return &TokenReader{
		tokens: sanitizedTokens,
		index:  -1,
	}
}

// Next returns the next token while advancing the reader. Returns false if there are no more tokens.
func (reader *TokenReader) Next() (Token, bool) {
	if reader.index+1 >= len(reader.tokens) {
		return Token{}, false
	}
	reader.index++
	return reader.tokens[reader.index], true
}

// Peek returns the next token. Does not advance the reader. Returns false if there are no more tokens.
func (reader *TokenReader) Peek() (Token, bool) {
	if reader.index+1 >= len(reader.tokens) {
		return Token{}, false
	}
	return reader.tokens[reader.index+1], true
}

// PeekBy returns the next token that is n positions from the current token. Does not advance the reader. Returns false
// if we are peeking beyond the slice.
func (reader *TokenReader) PeekBy(n int) (Token, bool) {
	if reader.index+n >= len(reader.tokens) || reader.index+n < 0 {
		return Token{}, false
	}
	return reader.tokens[reader.index+n], true
}

// Advance is equivalent to calling AdvanceBy(1).
func (reader *TokenReader) Advance() {
	reader.AdvanceBy(1)
}

// AdvanceBy advances the reader by the given amount. If the amount is greater than the number of remaining tokens, then
// it advances to the end. Cannot advance backwards.
func (reader *TokenReader) AdvanceBy(n int) {
	if n < 0 {
		n = 0
	}
	if reader.index+n >= len(reader.tokens) {
		n = len(reader.tokens) - reader.index - 1
	}
	reader.index += n
}
