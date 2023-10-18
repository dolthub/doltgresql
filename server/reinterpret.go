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

package server

import (
	"fmt"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// reinterpretQuery takes the given Postgres query, and reinterprets it as a ParsedQuery that will work with the handler.
func (l *Listener) reinterpretQuery(query string) (ParsedQuery, error) {
	s, err := parser.Parse(query)
	if err != nil {
		return ParsedQuery{}, err
	}
	if len(s) > 1 {
		return ParsedQuery{}, fmt.Errorf("only a single statement at a time is currently supported")
	}
	parsedAST := s[0].AST
	// Proof-of-concept on how this can be expanded and used. We'll eventually have a full translation layer to convert
	// from one AST to the other. For now, this lets us parse CREATE DATABASE while ignoring extra options like templates.
	switch ast := parsedAST.(type) {
	case *tree.CreateDatabase:
		vitessParsed := &vitess.DBDDL{
			Action:      vitess.CreateStr,
			DBName:      ast.Name.String(),
			IfNotExists: ast.IfNotExists,
		}
		// Normally we'd pass the original query in rather than use the empty string (for tracking purposes).
		// However, for the sake of demonstration, we're using an empty string so that it's clear that it's working.
		return ParsedQuery{"", vitessParsed}, nil
	default:
		return ParsedQuery{parsedAST.String(), nil}, nil
	}
}
