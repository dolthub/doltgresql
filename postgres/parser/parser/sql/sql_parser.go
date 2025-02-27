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

package sql

import (
	"context"
	"fmt"
	"strings"

	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/server/ast"

	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
)

var _ sql.Parser = &PostgresParser{}

// PostgresParser is a postgres syntax parser.
// This parser is used as parser in the engine for Doltgres.
type PostgresParser struct{}

// NewPostgresParser creates new PostgresParser.
func NewPostgresParser() *PostgresParser { return &PostgresParser{} }

// ParseSimple implements sql.Parser interface.
func (p *PostgresParser) ParseSimple(query string) (vitess.Statement, error) {
	stmt, _, _, err := p.ParseWithOptions(context.Background(), query, ';', false, vitess.ParserOptions{})
	return stmt, err
}

// Parse implements sql.Parser interface.
func (p *PostgresParser) Parse(ctx *sql.Context, query string, multi bool) (vitess.Statement, string, string, error) {
	return p.ParseWithOptions(ctx, query, ';', multi, vitess.ParserOptions{})
}

// ParseWithOptions implements sql.Parser interface.
func (p *PostgresParser) ParseWithOptions(ctx context.Context, query string, delimiter rune, _ bool, _ vitess.ParserOptions) (vitess.Statement, string, string, error) {
	q := sql.RemoveSpaceAndDelimiter(query, delimiter)
	stmts, err := parser.Parse(q)
	if err != nil {
		return nil, "", "", err
	}
	if len(stmts) > 1 {
		return nil, "", "", fmt.Errorf("only a single statement at a time is currently supported")
	}
	if len(stmts) == 0 {
		return nil, q, "", vitess.ErrEmpty
	}

	vitessAST, err := ast.Convert(stmts[0])
	if err != nil {
		return nil, "", "", err
	}
	if vitessAST == nil {
		q = stmts[0].AST.String()
	}

	return vitessAST, q, "", nil
}

// ParseOneWithOptions implements sql.Parser interface.
func (p *PostgresParser) ParseOneWithOptions(_ context.Context, query string, _ vitess.ParserOptions) (vitess.Statement, int, error) {
	stmt, err := parser.ParseOne(query)
	if err != nil {
		return nil, 0, err
	}
	vitessAST, err := ast.Convert(stmt)
	if err != nil {
		return nil, 0, err
	}
	return vitessAST, 0, nil
}

func (p *PostgresParser) QuoteIdentifier(identifier string) string {
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(identifier, `"`, `""`))
}
