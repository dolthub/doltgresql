package enginetest

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/lib/pq/oid"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/postgres/parser/types"
)

func convertQuery(query string) []string {
	if queries, converted := transformAST(query); converted {
		return queries
	}

	query = normalizeStrings(query)
	query = convertDoltProcedureCalls(query)
	return []string{query}
}

func transformAST(query string) ([]string, bool) {
	parser := sql.NewMysqlParser()
	stmt, err := parser.ParseSimple(query)
	if err != nil {
		return nil, false
	}

	switch stmt := stmt.(type) {
	case *sqlparser.DDL:
		if stmt.Action == "create" {
			return transformCreateTable(query, stmt)
		}
	case *sqlparser.Set:
		return transformSet(stmt)
	case *sqlparser.Select:
		return transformSelect(stmt)
	}

	return nil, false
}

// transformSelect converts a MySQL SELECT statement to a postgres-compatible SELECT statement.
// This is a very broad surface area, so we do this very selectively
func transformSelect(stmt *sqlparser.Select) ([]string, bool) {
	if !containsUserVars(stmt) {
		return nil, false
	}
	return []string{"gottem"}, true
}

func containsUserVars(stmt *sqlparser.Select) bool {
	foundUserVar := false
	detectUserVar := func(node sqlparser.SQLNode) (bool, error) {
		switch node := node.(type) {
		case *sqlparser.ColName:
			if strings.HasPrefix(node.Name.String(), "@") && !strings.HasPrefix(node.Name.String(), "@@") {
				foundUserVar = true
				return false, nil
			}
		}
		return true, nil
	}

	for _, sel := range stmt.SelectExprs {
		sqlparser.Walk(detectUserVar, sel)
	}

	if foundUserVar {
		return true
	}

	if stmt.Where != nil {
		sqlparser.Walk(detectUserVar, stmt.Where)
	}

	return foundUserVar
}

func transformSet(stmt *sqlparser.Set) ([]string, bool) {
	var queries []string

	// the semantics aren't quite the same, but setting autocommit to false is the same as beginning a transaction
	// (for most scripts). Setting autocommit to true is a no-op.
	if len(stmt.Exprs) == 1 && strings.ToLower(stmt.Exprs[0].Name.String()) == "autocommit" {
		exprStr := strings.ToLower(formatNode(stmt.Exprs[0].Expr))
		if exprStr == "0" || exprStr == "off" || exprStr == "'off'" || exprStr == "false" {
			queries = append(queries, "START TRANSACTION")
			return queries, true
		} else {
			return []string{""}, true
		}
	}

	for _, expr := range stmt.Exprs {
		if expr.Scope == sqlparser.GlobalStr {
			queries = append(queries, fmt.Sprintf("SET GLOBAL %s = %s", expr.Name, expr.Expr))
		} else if expr.Scope == "user" {
			queries = append(queries, fmt.Sprintf("SET doltgres_enginetest.%s = %s", expr.Name, formatNode(expr.Expr)))
		} else {
			queries = append(queries, fmt.Sprintf("SET %s = %s", expr.Name, expr.Expr))
		}
	}

	return queries, true
}

func formatNode(node sqlparser.SQLNode) string {
	buf := sqlparser.NewTrackedBuffer(PostgresNodeFormatter)
	node.Format(buf)
	return buf.String()
}

func PostgresNodeFormatter(buf *sqlparser.TrackedBuffer, node sqlparser.SQLNode) {
	switch node := node.(type) {
	case sqlparser.ColIdent:
		buf.Myprintf("%s", node.Lowered())
	case *sqlparser.Limit:
		if node == nil {
			return
		}
		buf.Myprintf(" limit %v", node.Rowcount)
		if node.Offset != nil {
			buf.Myprintf(" offset %v", node.Offset)
		}
	default:
		node.Format(buf)
	}
}

func transformCreateTable(query string, stmt *sqlparser.DDL) ([]string, bool) {
	if stmt.TableSpec == nil {
		return nil, false
	}

	createTable := tree.CreateTable{
		IfNotExists: stmt.IfNotExists,
		Table:       tree.MakeTableNameWithSchema("", "", tree.Name(stmt.Table.Name.String())), // TODO: qualified names
	}

	var queries []string
	for _, col := range stmt.TableSpec.Columns {
		createTable.Defs = append(createTable.Defs, &tree.ColumnTableDef{
			Name:      tree.Name(col.Name.String()),
			Type:      convertTypeDef(col.Type),
			Collation: "", // TODO
			Nullable: struct {
				Nullability    tree.Nullability
				ConstraintName tree.Name
			}{
				Nullability: convertNullability(col.Type),
			},
			PrimaryKey: struct {
				IsPrimaryKey bool
			}{
				IsPrimaryKey: col.Type.KeyOpt == 1, // TODO: unexported const
			},
			Unique:               col.Type.KeyOpt == 3, // TODO: unexported const
			UniqueConstraintName: "",                   // TODO
			DefaultExpr: struct {
				Expr           tree.Expr
				ConstraintName tree.Name
			}{
				Expr:           nil, // TODO
				ConstraintName: "",  // TODO
			},
			CheckExprs: nil, // TODO
		})
	}

	ctx := formatNodeWithUnqualifiedTableNames(&createTable)
	queries = append(queries, ctx.String())

	// If there are additional (non-primary key) indexes defined, each one gets its own additional statement
	if len(stmt.TableSpec.Indexes) > 0 {
		for _, index := range stmt.TableSpec.Indexes {
			if index.Info.Primary {
				continue
			}

			createIndex := tree.CreateIndex{
				Name:    tree.Name(index.Info.Name.String()),
				Table:   tree.MakeTableNameWithSchema("", "", tree.Name(stmt.Table.Name.String())), // TODO: qualified
				Unique:  index.Info.Unique,
				Columns: make(tree.IndexElemList, len(index.Columns)),
			}

			for i, col := range index.Columns {
				createIndex.Columns[i] = tree.IndexElem{
					Column:    tree.Name(col.Column.String()),
					Direction: tree.Ascending,
				}
			}

			ctx := formatNodeWithUnqualifiedTableNames(&createIndex)

			queries = append(queries, ctx.String())
		}
	}

	return queries, true
}

// The default formatter always qualifies table names with db name and schema name, which we don't want in most cases
func formatNodeWithUnqualifiedTableNames(n tree.NodeFormatter) *tree.FmtCtx {
	ctx := tree.NewFmtCtx(tree.FmtSimple)
	ctx.SetReformatTableNames(func(ctx *tree.FmtCtx, tn *tree.TableName) {
		ctx.FormatNode(&tn.ObjectName)
	})
	ctx.FormatNode(n)
	return ctx
}

func convertNullability(typ sqlparser.ColumnType) tree.Nullability {
	if typ.NotNull {
		return tree.NotNull
	}
	if typ.KeyOpt == 1 { // primary key, unexported constant
		return tree.NotNull
	}

	return tree.Null
}

func convertTypeDef(columnType sqlparser.ColumnType) tree.ResolvableTypeReference {
	switch strings.ToLower(columnType.Type) {
	case "int", "mediumint", "integer":
		return &types.T{
			InternalType: types.InternalType{
				Family: types.IntFamily,
				Width:  32,
				Oid:    oid.T_int4,
			},
		}
	case "tinyint", "smallint", "bool":
		return &types.T{
			InternalType: types.InternalType{
				Family: types.IntFamily,
				Width:  16,
				Oid:    oid.T_int2,
			},
		}
	case "bigint":
		return &types.T{
			InternalType: types.InternalType{
				Family: types.IntFamily,
				Width:  64,
				Oid:    oid.T_int8,
			},
		}
	case "float", "real":
		return &types.T{
			InternalType: types.InternalType{
				Family: types.FloatFamily,
				Width:  32,
				Oid:    oid.T_float4,
			},
		}
	case "double precision", "double":
		return &types.T{
			InternalType: types.InternalType{
				Family: types.FloatFamily,
				Oid:    oid.T_float8,
			},
		}
	case "decimal":
		return &types.T{
			InternalType: types.InternalType{
				Family: types.DecimalFamily,
				Oid:    oid.T_numeric,
			},
		}
	case "varchar":
		return &types.T{
			InternalType: types.InternalType{
				Family: types.StringFamily,
				Width:  int32FromSqlVal(columnType.Length),
				Oid:    oid.T_varchar,
			},
		}
	case "char":
		return &types.T{
			InternalType: types.InternalType{
				Family: types.StringFamily,
				Width:  int32FromSqlVal(columnType.Length),
				Oid:    oid.T_char,
			},
		}
	case "text", "tinytext", "mediumtext", "longtext":
		return &types.T{
			InternalType: types.InternalType{
				Family: types.StringFamily,
				Width:  int32FromSqlVal(columnType.Length),
				Oid:    oid.T_text,
			},
		}
	case "blob", "binary", "varbinary", "tinyblob", "mediumblob", "longblob":
		return &types.T{
			InternalType: types.InternalType{
				Family: types.BytesFamily,
				Width:  int32FromSqlVal(columnType.Length),
				Oid:    oid.T_bytea,
			},
		}
	case "datetime", "timestamp":
		return &types.T{
			InternalType: types.InternalType{
				Family: types.TimestampFamily,
				Width:  int32FromSqlVal(columnType.Length),
				Oid:    oid.T_timestamp,
			},
		}
	case "date":
		return &types.T{
			InternalType: types.InternalType{
				Family: types.DateFamily,
				Width:  int32FromSqlVal(columnType.Length),
				Oid:    oid.T_date,
			},
		}
	case "time":
		return &types.T{
			InternalType: types.InternalType{
				Family: types.TimeFamily,
				Width:  int32FromSqlVal(columnType.Length),
				Oid:    oid.T_time,
			},
		}
	case "enum":
		panic(fmt.Sprintf("unhandled type: %s", columnType.Type))
	case "set":
		panic(fmt.Sprintf("unhandled type: %s", columnType.Type))
	case "bit":
		panic(fmt.Sprintf("unhandled type: %s", columnType.Type))
	case "json":
		return &types.T{
			InternalType: types.InternalType{
				Family: types.JsonFamily,
				Width:  int32FromSqlVal(columnType.Length),
				Oid:    oid.T_json,
			},
		}
	case "geometry", "point", "linestring", "polygon", "multipoint", "multilinestring", "multipolygon", "geometrycollection":
		panic(fmt.Sprintf("unhandled type: %s", columnType.Type))
	default:
		panic(fmt.Sprintf("unhandled type: %s", columnType.Type))
	}
}

func int32FromSqlVal(v *sqlparser.SQLVal) int32 {
	if v == nil {
		return 0
	}

	i, err := strconv.Atoi(string(v.Val))
	if err != nil {
		return 0
	}
	return int32(i)
}

var doltProcedureCall = regexp.MustCompile(`(?i)CALL DOLT_(\w+)`)

func convertDoltProcedureCalls(query string) string {
	return doltProcedureCall.ReplaceAllString(query, "SELECT DOLT_$1")
}

// little state machine for turning MySQL quote characters into their postgres equivalents:
/*
               ┌───────────────────*─────────────────────────┐
               │                   ┌─*─┐                     *
               │               ┌───┴───▼──────┐         ┌────┴─────────┐
               │     ┌────"───►│ In double    │◄───"────┤End double    │
               │     │         │ quoted string│────"───►│quoted string?│
               │     │         └──────────────┘         └──────────────┘
               ├─────(──────────────────*───────────────────┐
      ┌─*──┐   ▼     │                                      *
      │    ├─────────┴┐            ┌─*─┐                    │
      └───►│ Not in   │        ┌───┴───▼─────┐          ┌───┴──────────┐
           │ string   ├───'───►│In single    │◄────'────┤End single    │
  ────────►└─────────┬┘        │quoted string│─────'───►│quoted string?│
  START        ▲     │         └─────────────┘          └──────────────┘
               └─────(──────────────────*───────────────────┐
                     │            ┌─*──┐                    *
                     │        ┌───┴────▼────┐           ┌───┴──────────┐
                     └───`───►│In backtick  │◄─────`────┤End backtick  │
                              │quoted string│──────`───►│quoted string?│
                              └─────────────┘           └──────────────┘
*/
type stringParserState byte

const (
	notInString stringParserState = iota
	inDoubleQuote
	maybeEndDoubleQuote
	inSingleQuote
	maybeEndSingleQuote
	inBackticks
	maybeEndBackticks
)

const singleQuote = '\''
const doubleQuote = '"'
const backtick = '`'
const backslash = '\\'

// normalizeStrings normalizes a query string to convert any MySQL syntax to Postgres syntax
func normalizeStrings(q string) string {
	state := notInString
	lastCharWasBackslash := false
	normalized := strings.Builder{}

	for _, c := range q {
		switch state {
		case notInString:
			switch c {
			case singleQuote:
				state = inSingleQuote
				normalized.WriteRune(singleQuote)
			case doubleQuote:
				state = inDoubleQuote
				normalized.WriteRune(singleQuote)
			case backtick:
				state = inBackticks
				normalized.WriteRune(doubleQuote)
			default:
				normalized.WriteRune(c)
			}
		case inDoubleQuote:
			switch c {
			case backslash:
				if lastCharWasBackslash {
					normalized.WriteRune(c)
				}
				lastCharWasBackslash = !lastCharWasBackslash
			case doubleQuote:
				if lastCharWasBackslash {
					normalized.WriteRune(c)
					lastCharWasBackslash = false
				} else {
					state = maybeEndDoubleQuote
				}
			case singleQuote:
				normalized.WriteRune(singleQuote)
				normalized.WriteRune(singleQuote)
				lastCharWasBackslash = false
			default:
				lastCharWasBackslash = false
				normalized.WriteRune(c)
			}
		case maybeEndDoubleQuote:
			switch c {
			case doubleQuote:
				state = inDoubleQuote
				normalized.WriteRune(doubleQuote)
			default:
				state = notInString
				normalized.WriteRune(singleQuote)
				normalized.WriteRune(c)
			}
		case inSingleQuote:
			switch c {
			case backslash:
				if lastCharWasBackslash {
					normalized.WriteRune(c)
				}
				lastCharWasBackslash = !lastCharWasBackslash
			case singleQuote:
				if lastCharWasBackslash {
					normalized.WriteRune(c)
					normalized.WriteRune(c)
					lastCharWasBackslash = false
				} else {
					state = maybeEndSingleQuote
				}
			default:
				lastCharWasBackslash = false
				normalized.WriteRune(c)
			}
		case maybeEndSingleQuote:
			switch c {
			case singleQuote:
				state = inSingleQuote
				normalized.WriteRune(singleQuote)
				normalized.WriteRune(singleQuote)
			default:
				state = notInString
				normalized.WriteRune(singleQuote)
				normalized.WriteRune(c)
			}
		case inBackticks:
			switch c {
			case backtick:
				state = maybeEndBackticks
			default:
				normalized.WriteRune(c)
			}
		case maybeEndBackticks:
			switch c {
			case backtick:
				state = inBackticks
				normalized.WriteRune(backtick)
			default:
				state = notInString
				normalized.WriteRune(doubleQuote)
				normalized.WriteRune(c)
			}
		default:
			panic("unknown state")
		}
	}

	// If reached the end of input unsure whether to unquote a string, do so now
	switch state {
	case maybeEndDoubleQuote:
		normalized.WriteRune(singleQuote)
	case maybeEndSingleQuote:
		normalized.WriteRune(singleQuote)
	case maybeEndBackticks:
		normalized.WriteRune(doubleQuote)
	default: // do nothing
	}

	return normalized.String()
}

// Test converting MySQL strings to Postgres strings
func TestNormalizeStrings(t *testing.T) {
	type test struct {
		input    string
		expected string
	}
	tests := []test{
		{
			input:    "SELECT \"foo\" FROM `bar`",
			expected: `SELECT 'foo' FROM "bar"`,
		},
		{
			input:    `SELECT "foo"`,
			expected: `SELECT 'foo'`,
		},
		{
			input:    `SELECT "fo\"o"`,
			expected: `SELECT 'fo"o'`,
		},
		{
			input:    `SELECT "fo\'o"`,
			expected: `SELECT 'fo''o'`,
		},
		{
			input:    `SELECT 'fo\'o'`,
			expected: `SELECT 'fo''o'`,
		},
		{
			input:    `SELECT 'fo\"o'`,
			expected: `SELECT 'fo"o'`,
		},
		{
			input:    `SELECT 'fo\\"o'`,
			expected: `SELECT 'fo\"o'`,
		},
		{
			input:    `SELECT 'fo\\\'o'`,
			expected: `SELECT 'fo\''o'`,
		},
		{
			input:    `SELECT "fo\\'o"`,
			expected: `SELECT 'fo\''o'`,
		},
		{
			input:    `SELECT "fo\\\"o"`,
			expected: `SELECT 'fo\"o'`,
		},
		{
			input:    "SELECT 'fo''o'",
			expected: `SELECT 'fo''o'`,
		},
		{
			input:    "SELECT 'fo''''o'",
			expected: `SELECT 'fo''''o'`,
		},
		{
			input:    `SELECT "fo'o"`,
			expected: `SELECT 'fo''o'`,
		},
		{
			input:    `SELECT "fo''o"`,
			expected: `SELECT 'fo''''o'`,
		},
		{
			input:    `SELECT "fo""o"`,
			expected: `SELECT 'fo"o'`,
		},
		{
			input:    `SELECT "fo""""o"`,
			expected: `SELECT 'fo""o'`,
		},
		{
			input:    `SELECT 'fo""o'`,
			expected: `SELECT 'fo""o'`,
		},
		{
			input:    "SELECT `foo` FROM `bar`",
			expected: `SELECT "foo" FROM "bar"`,
		},
		{
			input:    "SELECT 'foo' FROM `bar`",
			expected: `SELECT 'foo' FROM "bar"`,
		},
		{
			input:    "SELECT `f\"o'o` FROM `ba``r`",
			expected: "SELECT \"f\"o'o\" FROM \"ba`r\"",
		},
		{
			input:    "SELECT \"foo\" from `bar` where `bar`.`baz` = \"qux\"",
			expected: `SELECT 'foo' from "bar" where "bar"."baz" = 'qux'`,
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			actual := normalizeStrings(test.input)
			require.Equal(t, test.expected, actual)
		})
	}
}

func TestConvertQuery(t *testing.T) {
	type test struct {
		input    string
		expected []string
	}
	tests := []test{
		{
			input:    "CREATE TABLE foo (a INT primary key)",
			expected: []string{"CREATE TABLE foo (a INTEGER NOT NULL PRIMARY KEY)"},
		},
		{
			input: "CREATE TABLE foo (a INT primary key, b int not null)",
			expected: []string{
				"CREATE TABLE foo (a INTEGER NOT NULL PRIMARY KEY, b INTEGER NOT NULL)",
			},
		},
		{
			input: "CREATE TABLE foo (a INT primary key, b int, key (b))",
			expected: []string{
				"CREATE TABLE foo (a INTEGER NOT NULL PRIMARY KEY, b INTEGER NULL)",
				"CREATE INDEX ON foo ( b ASC ) NULLS NOT DISTINCT ",
			},
		},
		{
			input: "CREATE TABLE foo (a INT primary key, b int, c int, key (c,b))",
			expected: []string{
				"CREATE TABLE foo (a INTEGER NOT NULL PRIMARY KEY, b INTEGER NULL, c INTEGER NULL)",
				"CREATE INDEX ON foo ( c ASC, b ASC ) NULLS NOT DISTINCT ",
			},
		},
		{
			input: "CREATE TABLE foo (a INT primary key, b int, c int not null, d int, key (c), key (b), key (b,c))",
			expected: []string{
				"CREATE TABLE foo (a INTEGER NOT NULL PRIMARY KEY, b INTEGER NULL, c INTEGER NOT NULL, d INTEGER NULL)",
				"CREATE INDEX ON foo ( c ASC ) NULLS NOT DISTINCT ",
				"CREATE INDEX ON foo ( b ASC ) NULLS NOT DISTINCT ",
				"CREATE INDEX ON foo ( b ASC, c ASC ) NULLS NOT DISTINCT ",
			},
		},
		{
			input:    "SET @@autocommit = 1",
			expected: []string{""},
		},
		{
			input:    "SET @@autocommit = 0",
			expected: []string{"START TRANSACTION"},
		},
		{
			input:    "SET @@autocommit = off",
			expected: []string{"START TRANSACTION"},
		},
		{
			input: "SET @@autocommit = 1, @@dolt_transaction_commit = off",
			expected: []string{
				"SET autocommit = 1",
				"SET dolt_transaction_commit = 'off'",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			actual := convertQuery(test.input)
			require.Equal(t, test.expected, actual)
		})
	}
}
