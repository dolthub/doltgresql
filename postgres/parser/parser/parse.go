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

// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in licenses/BSD-vitess.txt.

// Portions of this file are additionally subject to the following
// license and copyright.
//
// Copyright 2015 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// This code was derived from https://github.com/youtube/vitess.

package parser

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/postgres/parser/types"
)

// Statement is the result of parsing a single statement. It contains the AST
// node along with other information.
type Statement struct {
	// AST is the root of the AST tree for the parsed statement.
	AST tree.Statement

	// SQL is the original SQL from which the statement was parsed. Note that this
	// is not appropriate for use in logging, as it may contain passwords and
	// other sensitive data.
	SQL string

	// NumPlaceholders indicates the number of arguments to the statement (which
	// are referenced through placeholders). This corresponds to the highest
	// argument position (i.e. the x in "$x") that appears in the query.
	//
	// Note: where there are "gaps" in the placeholder positions, this number is
	// based on the highest position encountered. For example, for `SELECT $3`,
	// NumPlaceholders is 3. These cases are malformed and will result in a
	// type-check error.
	NumPlaceholders int

	// NumAnnotations indicates the number of annotations in the tree. It is equal
	// to the maximum annotation index.
	NumAnnotations tree.AnnotationIdx
}

// Statements is a list of parsed statements.
type Statements []Statement

// String returns the AST formatted as a string.
func (stmts Statements) String() string {
	return stmts.StringWithFlags(tree.FmtSimple)
}

// StringWithFlags returns the AST formatted as a string (with the given flags).
func (stmts Statements) StringWithFlags(flags tree.FmtFlags) string {
	ctx := tree.NewFmtCtx(flags)
	for i, s := range stmts {
		if i > 0 {
			ctx.WriteString("; ")
		}
		ctx.FormatNode(s.AST)
	}
	return ctx.CloseAndGetString()
}

// Parser wraps a scanner, parser and other utilities present in the parser
// package.
type Parser struct {
	scanner    scanner
	lexer      lexer
	parserImpl sqlParserImpl
	tokBuf     [8]sqlSymType
	stmtBuf    [1]Statement
}

// This is the default integer type when no integer type is explicitly specified.
var defaultNakedIntType = types.Int4

// Parse parses the sql and returns a list of statements.
func (p *Parser) Parse(sql string) (Statements, error) {
	return p.parseWithDepth(1, sql, defaultNakedIntType)
}

// ParseWithInt parses a sql statement string and returns a list of
// Statements. The INT token will result in the specified TInt type.
func (p *Parser) ParseWithInt(sql string, nakedIntType *types.T) (Statements, error) {
	return p.parseWithDepth(1, sql, nakedIntType)
}

func (p *Parser) parseOneWithDepth(depth int, sql string) (Statement, error) {
	stmts, err := p.parseWithDepth(1, sql, defaultNakedIntType)
	if err != nil {
		return Statement{}, err
	}
	if len(stmts) > 1 {
		return Statement{}, errors.AssertionFailedf("expected 1 statement, but found %d", len(stmts))
	} else if len(stmts) == 0 {
		return Statement{}, vitess.ErrEmpty
	}
	return stmts[0], nil
}

func (p *Parser) scanOneStmt() (sql string, tokens []sqlSymType, done bool) {
	var lval sqlSymType
	tokens = p.tokBuf[:0]

	// Scan the first token.
	for {
		p.scanner.scan(&lval)
		if lval.id == 0 {
			return "", nil, true
		}
		if lval.id != ';' {
			break
		}
	}

	startPos := lval.pos
	// We make the resulting token positions match the returned string.
	lval.pos = 0
	tokens = append(tokens, lval)
	for {
		if lval.id == ERROR {
			return p.scanner.in[startPos:], tokens, true
		}
		posBeforeScan := p.scanner.pos
		p.scanner.scan(&lval)
		if lval.id == 0 || lval.id == ';' {
			return p.scanner.in[startPos:posBeforeScan], tokens, (lval.id == 0)
		}
		lval.pos -= startPos
		tokens = append(tokens, lval)
	}
}

func (p *Parser) parseWithDepth(depth int, sql string, nakedIntType *types.T) (Statements, error) {
	stmts := Statements(p.stmtBuf[:0])
	p.scanner.init(sql)
	defer p.scanner.cleanup()
	for {
		sql, tokens, done := p.scanOneStmt()
		stmt, err := p.parse(depth+1, sql, tokens, nakedIntType)
		if err != nil {
			return nil, err
		}
		if stmt.AST != nil {
			stmts = append(stmts, stmt)
		}
		if done {
			break
		}
	}
	return stmts, nil
}

// parse parses a statement from the given scanned tokens.
func (p *Parser) parse(
	depth int, sql string, tokens []sqlSymType, nakedIntType *types.T,
) (Statement, error) {
	p.lexer.init(sql, tokens, nakedIntType)
	defer p.lexer.cleanup()
	if p.parserImpl.Parse(&p.lexer) != 0 {
		if p.lexer.lastError == nil {
			// This should never happen -- there should be an error object
			// every time Parse() returns nonzero. We're just playing safe
			// here.
			p.lexer.Error("syntax error")
		}
		err := p.lexer.lastError

		// Compatibility with 19.1 telemetry: prefix the telemetry keys
		// with the "syntax." prefix.
		// TODO(knz): move the auto-prefixing of feature names to a
		// higher level in the call stack.
		tkeys := errors.GetTelemetryKeys(err)
		if len(tkeys) > 0 {
			for i := range tkeys {
				tkeys[i] = "syntax." + tkeys[i]
			}
			err = errors.WithTelemetry(err, tkeys...)
		}

		return Statement{}, err
	}
	return Statement{
		AST:             p.lexer.stmt,
		SQL:             sql,
		NumPlaceholders: p.lexer.numPlaceholders,
		NumAnnotations:  p.lexer.numAnnotations,
	}, nil
}

// unaryNegation constructs an AST node for a negation. This attempts
// to preserve constant NumVals and embed the negative sign inside
// them instead of wrapping in an UnaryExpr. This in turn ensures
// that negative numbers get considered as a single constant
// for the purpose of formatting and scrubbing.
func unaryNegation(e tree.Expr) tree.Expr {
	if cst, ok := e.(*tree.NumVal); ok {
		cst.Negate()
		return cst
	}

	// Common case.
	return &tree.UnaryExpr{Operator: tree.UnaryMinus, Expr: e}
}

// Parse parses a sql statement string and returns a list of Statements.
func Parse(sql string) (Statements, error) {
	var p Parser
	return p.parseWithDepth(1, sql, defaultNakedIntType)
}

// ParseOne parses a sql statement string, ensuring that it contains only a
// single statement, and returns that Statement. ParseOne will always
// interpret the INT and SERIAL types as 64-bit types, since this is
// used in various internal-execution paths where we might receive
// bits of SQL from other nodes. In general,earwe expect that all
// user-generated SQL has been run through the ParseWithInt() function.
func ParseOne(sql string) (Statement, error) {
	var p Parser
	return p.parseOneWithDepth(1, sql)
}

// HasMultipleStatements returns true if the sql string contains more than one
// statements.
func HasMultipleStatements(sql string) bool {
	var p Parser
	p.scanner.init(sql)
	defer p.scanner.cleanup()
	count := 0
	for {
		_, _, done := p.scanOneStmt()
		if done {
			break
		}
		count++
		if count > 1 {
			return true
		}
	}
	return false
}

// ParseQualifiedTableName parses a SQL string of the form
// `[ database_name . ] [ schema_name . ] table_name`.
func ParseQualifiedTableName(sql string) (*tree.TableName, error) {
	name, err := ParseTableName(sql)
	if err != nil {
		return nil, err
	}
	tn := name.ToTableName()
	return &tn, nil
}

// ParseTableName parses a table name.
func ParseTableName(sql string) (*tree.UnresolvedObjectName, error) {
	// We wrap the name we want to parse into a dummy statement since our parser
	// can only parse full statements.
	stmt, err := ParseOne(fmt.Sprintf("ALTER TABLE %s RENAME TO x", sql))
	if err != nil {
		return nil, err
	}
	rename, ok := stmt.AST.(*tree.RenameTable)
	if !ok {
		return nil, errors.AssertionFailedf("expected an ALTER TABLE statement, but found %T", stmt)
	}
	return rename.Name, nil
}

// ParseTableNameWithQualifiedNames can be used to parse an input table name that
// might be prefixed with an unquoted qualified name. The standard ParseTableName
// cannot do this due to limitations with our parser. In particular, the parser
// can't parse different productions individually -- it must parse them as part
// of a top level statement. This causes qualified names that contain keywords
// to require quotes, which are not required in some cases due to Postgres
// compatibility (in particular, as arguments to pg_dump). This function gets
// around this limitation by parsing the input table name as a column name
// with a fake non-keyword prefix, and then shifting the result down into an
// UnresolvedObjectName.
func ParseTableNameWithQualifiedNames(sql string) (*tree.UnresolvedObjectName, error) {
	stmt, err := ParseOne(fmt.Sprintf("SELECT fakeprefix.%s", sql))
	if err != nil {
		return nil, err
	}
	un := stmt.AST.(*tree.Select).Select.(*tree.SelectClause).Exprs[0].Expr.(*tree.UnresolvedName)
	var nameParts [3]string
	numParts := un.NumParts - 1
	for i := 0; i < numParts; i++ {
		nameParts[i] = un.Parts[i]
	}
	return tree.NewUnresolvedObjectName(numParts, nameParts, 0 /* annotationIdx */)
}

// parseExprs parses one or more sql expressions.
func parseExprs(exprs []string) (tree.Exprs, error) {
	stmt, err := ParseOne(fmt.Sprintf("SET ROW (%s)", strings.Join(exprs, ",")))
	if err != nil {
		return nil, err
	}
	set, ok := stmt.AST.(*tree.SetVar)
	if !ok {
		return nil, errors.AssertionFailedf("expected a SET statement, but found %T", stmt)
	}
	return set.Values, nil
}

// ParseExprs is a short-hand for parseExprs(sql)
func ParseExprs(sql []string) (tree.Exprs, error) {
	if len(sql) == 0 {
		return tree.Exprs{}, nil
	}
	return parseExprs(sql)
}

// ParseExpr is a short-hand for parseExprs([]string{sql})
func ParseExpr(sql string) (tree.Expr, error) {
	exprs, err := parseExprs([]string{sql})
	if err != nil {
		return nil, err
	}
	if len(exprs) != 1 {
		return nil, errors.AssertionFailedf("expected 1 expression, found %d", len(exprs))
	}
	return exprs[0], nil
}

// ParseType parses a column type.
func ParseType(sql string) (tree.ResolvableTypeReference, error) {
	expr, err := ParseExpr(fmt.Sprintf("1::%s", sql))
	if err != nil {
		return nil, err
	}

	cast, ok := expr.(*tree.CastExpr)
	if !ok {
		return nil, errors.AssertionFailedf("expected a tree.CastExpr, but found %T", expr)
	}

	return cast.Type, nil
}

var errBitLengthNotPositive = pgerror.WithCandidateCode(
	errors.New("length for type bit must be at least 1"), pgcode.InvalidParameterValue)

// newBitType creates a new BIT type with the given bit width.
func newBitType(width int32, varying bool) (*types.T, error) {
	if width < 1 {
		return nil, errBitLengthNotPositive
	}
	if varying {
		return types.MakeVarBit(width), nil
	}
	return types.MakeBit(width), nil
}

var errFloatPrecAtLeast1 = pgerror.WithCandidateCode(
	errors.New("precision for type float must be at least 1 bit"), pgcode.InvalidParameterValue)
var errFloatPrecMax54 = pgerror.WithCandidateCode(
	errors.New("precision for type float must be less than 54 bits"), pgcode.InvalidParameterValue)

// newFloat creates a type for FLOAT with the given precision.
func newFloat(prec int64) (*types.T, error) {
	if prec < 1 {
		return nil, errFloatPrecAtLeast1
	}
	if prec <= 24 {
		return types.Float4, nil
	}
	if prec <= 54 {
		return types.Float, nil
	}
	return nil, errFloatPrecMax54
}

// newDecimal creates a type for DECIMAL with the given precision and scale.
func newDecimal(prec, scale int32) (*types.T, error) {
	if scale > prec {
		err := pgerror.WithCandidateCode(
			errors.Newf("scale (%d) must be between 0 and precision (%d)", scale, prec),
			pgcode.InvalidParameterValue)
		return nil, err
	}
	return types.MakeDecimal(prec, scale), nil
}

// arrayOf creates a type alias for an array of the given element type and fixed
// bounds. The bounds are currently ignored.
func arrayOf(
	ref tree.ResolvableTypeReference, bounds []int32,
) (tree.ResolvableTypeReference, error) {
	// If the reference is a statically known type, then return an array type,
	// rather than an array type reference.
	if typ, ok := tree.GetStaticallyKnownType(ref); ok {
		return types.MakeArray(typ), nil
	}
	return &tree.ArrayTypeReference{ElementType: ref}, nil
}
