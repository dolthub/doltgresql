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

// TokenType is the type of the token.
type TokenType uint8

const (
	TokenType_Text TokenType = iota
	TokenType_Variable
	TokenType_VariableDefinition
	TokenType_Or
	TokenType_Repeat
	TokenType_OptionalRepeat
	TokenType_ShortSpace
	TokenType_MediumSpace
	TokenType_LongSpace
	TokenType_ParenOpen
	TokenType_ParenClose
	TokenType_OptionalOpen
	TokenType_OptionalClose
	TokenType_OneOfOpen
	TokenType_OneOfClose
	TokenType_EOF
)

// Token represents a token in the synopsis. Unlike with traditional lexers, whitespace is tokenized, with multiple
// token types for differing lengths, as the whitespace has some significance to the definition.
type Token struct {
	Type    TokenType
	Literal string
}

// IsSpace returns whether the token represents one of the space types.
func (t Token) IsSpace() bool {
	return t.Type == TokenType_ShortSpace || t.Type == TokenType_MediumSpace || t.Type == TokenType_LongSpace
}

// IsStandardSpace returns whether the token represents a short or medium space, which will not end a statement.
func (t Token) IsStandardSpace() bool {
	return t.Type == TokenType_ShortSpace || t.Type == TokenType_MediumSpace
}

// IsNewStatement returns whether the token represents one a long space, which will end a statement
func (t Token) IsNewStatement() bool {
	return t.Type == TokenType_LongSpace
}

// CreatesNewScope returns whether the token creates a new scope.
func (t Token) CreatesNewScope() bool {
	return t.Type == TokenType_ParenOpen || t.Type == TokenType_OptionalOpen || t.Type == TokenType_OneOfOpen
}

// ExitsScope returns whether the token exits the current scope.
func (t Token) ExitsScope() bool {
	return t.Type == TokenType_ParenClose || t.Type == TokenType_OptionalClose || t.Type == TokenType_OneOfClose
}

// String returns a string representation of the token.
func (t Token) String() string {
	switch t.Type {
	case TokenType_Text:
		return t.Literal
	case TokenType_Variable:
		return "$" + t.Literal + "$"
	case TokenType_VariableDefinition:
		return "where $" + t.Literal + "$ is:"
	case TokenType_Or:
		return "|"
	case TokenType_Repeat:
		return "..."
	case TokenType_OptionalRepeat:
		if len(t.Literal) > 0 {
			return "[ " + t.Literal + " ... ]"
		} else {
			return "[ ... ]"
		}
	case TokenType_ShortSpace:
		return " "
	case TokenType_MediumSpace:
		return "\n    "
	case TokenType_LongSpace:
		return "\n\n"
	case TokenType_ParenOpen:
		return "("
	case TokenType_ParenClose:
		return ")"
	case TokenType_OptionalOpen:
		return "["
	case TokenType_OptionalClose:
		return "]"
	case TokenType_OneOfOpen:
		return "{"
	case TokenType_OneOfClose:
		return "}"
	case TokenType_EOF:
		return ""
	default:
		panic("unexpected token type")
	}
}
