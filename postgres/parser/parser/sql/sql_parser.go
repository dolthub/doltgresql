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

type PostgresFormatter struct {}

var _ sql.SchemaFormatter = PostgresFormatter{}

// NewPostgresSchemaFormatter creates a new PostgresFormatter.
func NewPostgresSchemaFormatter() PostgresFormatter {
	return PostgresFormatter{}
}

// GenerateCreateTableStatement implements sql.SchemaFormatter interface.
func (p PostgresFormatter) GenerateCreateTableStatement(tblName string, colStmts []string, temp, autoInc, tblCharsetName, tblCollName, comment string) string {
	return fmt.Sprintf(
		"CREATE%s TABLE %s (\n%s\n)",
		temp,
		p.QuoteIdentifier(tblName),
		strings.Join(colStmts, ",\n"),
	)
}

func (p PostgresFormatter) GenerateCreateTableColumnDefinition(col *sql.Column, colDefault, onUpdate string, tableCollation sql.CollationID) string {
	var colTypeString = col.Type.String()
	if collationType, ok := col.Type.(sql.TypeWithCollation); ok {
		colTypeString = collationType.StringWithTableCollation(tableCollation)
	} 
	
	stmt := fmt.Sprintf("  %s %s", p.QuoteIdentifier(col.Name), colTypeString)
	if !col.Nullable {
		stmt = fmt.Sprintf("%s NOT NULL", stmt)
	}
	
	if col.Generated != nil {
		storedStr := " STORED"
		stmt = fmt.Sprintf("%s GENERATED ALWAYS AS %s%s", stmt, col.Generated.String(), storedStr)
	}

	if col.Default != nil && col.Generated == nil {
		stmt = fmt.Sprintf("%s DEFAULT %s", stmt, colDefault)
	}

	// TODO: comments
	return stmt
}

func (p PostgresFormatter) GenerateCreateTablePrimaryKeyDefinition(pkCols []string) string {
	return fmt.Sprintf("  PRIMARY KEY (%s)", strings.Join(p.QuoteIdentifiers(pkCols), ","))
}

func (p PostgresFormatter) GenerateCreateTableIndexDefinition(isUnique, isSpatial, isFullText, isVector bool, indexID string, indexCols []string, comment string) (string, bool) {
	if isUnique {
		return fmt.Sprintf("  UNIQUE %s (%s)", p.QuoteIdentifier(indexID), strings.Join(indexCols, ",")), true
	}

	// TODO: this interface is not sufficient for SHOW CREATE TABLE output, where we will need to return multiple
	//  statements to capture index creation for non-unique indexes
	return "", false
}

func (p PostgresFormatter) GenerateCreateTableForiegnKeyDefinition(fkName string, fkCols []string, parentTbl string, parentCols []string, onDelete, onUpdate string) string {
	keyCols := strings.Join(p.QuoteIdentifiers(fkCols), ",")
	refCols := strings.Join(p.QuoteIdentifiers(parentCols), ",")
	fkey := fmt.Sprintf("  CONSTRAINT %s FOREIGN KEY (%s) REFERENCES %s (%s)", p.QuoteIdentifier(fkName), keyCols, p.QuoteIdentifier(parentTbl), refCols)
	if onDelete != "" {
		fkey = fmt.Sprintf("%s ON DELETE %s", fkey, onDelete)
	}
	if onUpdate != "" {
		fkey = fmt.Sprintf("%s ON UPDATE %s", fkey, onUpdate)
	}
	return fkey
}

func (p PostgresFormatter) GenerateCreateTableCheckConstraintClause(checkName, checkExpr string, enforced bool) string {
	return fmt.Sprintf("  CONSTRAINT %s CHECK (%s)", p.QuoteIdentifier(checkName), checkExpr)
}

func (p PostgresFormatter) QuoteIdentifier(identifier string) string {
	return fmt.Sprintf(`"%s"`, strings.ReplaceAll(identifier, `"`, `""`))
}

func (p PostgresFormatter) QuoteIdentifiers(ids []string) []string {
	quoted := make([]string, len(ids))
	for i, id := range ids {
		quoted[i] = p.QuoteIdentifier(id)
	}
	return quoted
}
