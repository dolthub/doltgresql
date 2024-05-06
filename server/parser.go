package server

import (
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/server/ast"
)

var _ sql.Parser = &PostgresParser{}

type PostgresParser struct{}

func NewPostgresParser() *PostgresParser { return &PostgresParser{} }

func (p *PostgresParser) ParseSimple(query string) (vitess.Statement, error) {
	stmt, _, _, err := p.ParseWithOptions(query, ';', false, vitess.ParserOptions{})
	return stmt, err
}

func (p *PostgresParser) Parse(_ *sql.Context, query string, multi bool) (vitess.Statement, string, string, error) {
	return p.ParseWithOptions(query, ';', multi, vitess.ParserOptions{})
}

func (p *PostgresParser) ParseWithOptions(query string, delimiter rune, _ bool, _ vitess.ParserOptions) (vitess.Statement, string, string, error) {
	q := sql.RemoveSpaceAndDelimiter(query, delimiter)
	if vitessStmt, err := vitess.Parse(q); err == nil {
		return vitessStmt, q, "", err
	}
	stmts, err := parser.Parse(q)
	if err != nil {
		return nil, "", "", err
	}
	if len(stmts) > 1 {
		return nil, "", "", fmt.Errorf("only a single statement at a time is currently supported")
	}
	if len(stmts) == 0 {
		return nil, q, "", nil
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

func (p *PostgresParser) ParseOneWithOptions(query string, _ vitess.ParserOptions) (vitess.Statement, int, error) {
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
