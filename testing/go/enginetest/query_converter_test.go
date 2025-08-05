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

	"github.com/dolthub/doltgresql/postgres/parser/parser"
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
		switch stmt.Action {
		case "create":
			return transformCreateTable(stmt)
		case "drop":
			return transformDrop(query, stmt)
		case "rename":
			return transformRename(stmt)
		}
	case *sqlparser.Set:
		return transformSet(stmt)
	case *sqlparser.Select:
		return transformSelect(stmt)
	case *sqlparser.Insert:
		return transformInsert(stmt)
	case *sqlparser.AlterTable:
		return transformAlterTable(stmt)
	}

	return nil, false
}

func transformRename(stmt *sqlparser.DDL) ([]string, bool) {
	rename := &tree.RenameTable{
		Name:    TableNameToUnresolvedObjectName(stmt.FromTables[0]),
		NewName: TableNameToUnresolvedObjectName(stmt.ToTables[0]),
	}

	ctx := formatNodeWithUnqualifiedTableNames(rename)
	return []string{ctx.String()}, true
}

func transformInsert(stmt *sqlparser.Insert) ([]string, bool) {
	// only bother translating inserts if there's an ON DUPLICATE KEY UPDATE clause, maybe revisit this later
	table := stmt.Table
	if len(stmt.OnDup) > 0 {
		tableName := translateTableName(table)

		var colList tree.NameList
		if len(stmt.Columns) > 0 {
			colList = make(tree.NameList, len(stmt.Columns))
			for i, col := range stmt.Columns {
				colList[i] = tree.Name(col.String())
			}
		}

		rows := rowsForInsert(stmt.Rows)

		onConflict := tree.OnConflict{
			Exprs:   convertUpdateExprs(sqlparser.AssignmentExprs(stmt.OnDup)),
			Columns: tree.NameList{tree.Name("fake")}, // column list ignored but must be present for valid syntax
		}

		insert := tree.Insert{
			Table:      tableName,
			Columns:    colList,
			Rows:       rows,
			OnConflict: &onConflict,
			Returning:  &tree.NoReturningClause{},
		}

		ctx := formatNodeWithUnqualifiedTableNames(&insert)
		return []string{ctx.String()}, true
	} else if stmt.Ignore == "ignore " {
		tableName := tree.NewTableName(tree.Name(table.DbQualifier.String()), tree.Name(table.Name.String()))

		var colList tree.NameList
		if len(stmt.Columns) > 0 {
			colList = make(tree.NameList, len(stmt.Columns))
			for i, col := range stmt.Columns {
				colList[i] = tree.Name(col.String())
			}
		}

		rows := rowsForInsert(stmt.Rows)

		onConflict := tree.OnConflict{
			Columns:   tree.NameList{tree.Name("fake")}, // column list ignored but must be present for valid syntax
			DoNothing: true,
		}

		insert := tree.Insert{
			Table:      tableName,
			Columns:    colList,
			Rows:       rows,
			OnConflict: &onConflict,
			Returning:  &tree.NoReturningClause{},
		}

		ctx := formatNodeWithUnqualifiedTableNames(&insert)
		return []string{ctx.String()}, true
	}

	return nil, false
}

func translateTableName(table sqlparser.TableName) *tree.TableName {
	return tree.NewTableName(tree.Name(table.DbQualifier.String()), tree.Name(table.Name.String()))
}

func TableNameToUnresolvedObjectName(table sqlparser.TableName) *tree.UnresolvedObjectName {
	if !table.DbQualifier.IsEmpty() {
		panic(fmt.Sprintf("unhandled case: db qualifier present %v", table))
	}

	name, err := tree.NewUnresolvedObjectName(1, [3]string{table.Name.String(), "", ""}, 0)
	if err != nil {
		panic(err)
	}
	return name
}

func convertUpdateExprs(exprs sqlparser.AssignmentExprs) tree.UpdateExprs {
	updateExprs := make(tree.UpdateExprs, len(exprs))
	for i, expr := range exprs {
		updateExprs[i] = &tree.UpdateExpr{
			Names: tree.NameList{tree.Name(expr.Name.String())},
			Expr:  convertExpr(expr.Expr),
		}
	}
	return updateExprs
}

func rowsForInsert(rows sqlparser.InsertRows) *tree.Select {
	switch rows := rows.(type) {
	case sqlparser.Values:
		return &tree.Select{
			Select: &tree.ValuesClause{
				Rows: insertValuesToExprs(rows),
			},
		}
	case *sqlparser.Select:
		return &tree.Select{
			Select: convertSelect(rows),
		}
	case *sqlparser.ParenSelect:
		return &tree.Select{
			Select: &tree.ParenSelect{
				Select: convertParentSelect(rows.Select),
			},
		}
	case *sqlparser.AliasedValues:
		return &tree.Select{
			Select: &tree.ValuesClause{
				Rows: insertValuesToExprs(rows.Values),
			},
		}
	case *sqlparser.SetOp:
		return &tree.Select{
			Select: convertSelectStatement(rows),
		}
	default:
		panic(fmt.Sprintf("unhandled type: %T", rows))
	}
}

func convertParentSelect(statement sqlparser.SelectStatement) *tree.Select {
	switch statement := statement.(type) {
	case *sqlparser.Select:
		sel := convertSelect(statement)
		return &tree.Select{
			Select: sel,
		}
	default:
		panic(fmt.Sprintf("unhandled type: %T", statement))
	}
}

func convertSelect(sel *sqlparser.Select) *tree.SelectClause {
	return &tree.SelectClause{
		Distinct: sel.QueryOpts.Distinct,
		Exprs:    convertSelectExprs(sel.SelectExprs),
		From:     convertFrom(sel.From),
		Where:    convertWhere(sel.Where),
		GroupBy:  convertGroupBy(sel.GroupBy),
		Having:   convertHaving(sel.Having),
	}
}

func convertSelectStatement(sel sqlparser.SelectStatement) tree.SelectStatement {
	switch sel := sel.(type) {
	case *sqlparser.Select:
		return convertSelect(sel)
	case *sqlparser.SetOp:
		return convertSetOp(sel)
	default:
		panic(fmt.Sprintf("unhandled type: %T", sel))
	}
}

func convertSetOp(sel *sqlparser.SetOp) tree.SelectStatement {
	switch sel.Type {
	case sqlparser.UnionStr:
		left := convertSelectStatement(sel.Left)
		right := convertSelectStatement(sel.Right)
		return &tree.UnionClause{
			Type:  tree.UnionOp,
			Left:  selectFromSelectClause(left.(*tree.SelectClause)),
			Right: selectFromSelectClause(right.(*tree.SelectClause)),
		}
	default:
		panic(fmt.Sprintf("unhandled type: %s", sel.Type))
	}
}

func selectFromSelectClause(clause *tree.SelectClause) *tree.Select {
	return &tree.Select{
		Select: clause,
	}
}

func convertHaving(having *sqlparser.Where) *tree.Where {
	return convertWhere(having)
}

func convertGroupBy(groupBy sqlparser.GroupBy) tree.GroupBy {
	return convertExprs(sqlparser.Exprs(groupBy))
}

func convertWhere(where *sqlparser.Where) *tree.Where {
	if where == nil {
		return nil
	}
	return &tree.Where{
		Type: tree.AstWhere,
		Expr: convertExpr(where.Expr),
	}
}

func convertFrom(from sqlparser.TableExprs) tree.From {
	tables := make(tree.TableExprs, len(from))

	for i, table := range from {
		tables[i] = convertTableExpr(table)
	}
	return tree.From{
		Tables: tables,
	}
}

func convertTableExpr(table sqlparser.TableExpr) tree.TableExpr {
	switch table := table.(type) {
	case *sqlparser.AliasedTableExpr:
		switch tableExpr := table.Expr.(type) {
		case sqlparser.TableName:
			return &tree.AliasedTableExpr{
				Expr: tree.NewTableName(tree.Name(tableExpr.DbQualifier.String()), tree.Name(tableExpr.Name.String())),
				As: tree.AliasClause{
					Alias: tree.Name(table.As.String()),
				},
			}
		default:
			panic(fmt.Sprintf("unhandled type: %T", table))
		}
	default:
		panic(fmt.Sprintf("unhandled type: %T", table))
	}
}

func convertSelectExprs(exprs sqlparser.SelectExprs) tree.SelectExprs {
	es := make(tree.SelectExprs, len(exprs))
	for i, expr := range exprs {
		es[i] = convertSelectExpr(expr)
	}
	return es
}

func insertValuesToExprs(values sqlparser.Values) []tree.Exprs {
	exprs := make([]tree.Exprs, len(values))
	for i, row := range values {
		exprs[i] = make(tree.Exprs, len(row))
		for j, val := range row {
			exprs[i][j] = convertValue(val)
		}
	}
	return exprs
}

func convertValue(val sqlparser.Expr) tree.Expr {
	switch val := val.(type) {
	case *sqlparser.SQLVal:
		return convertSQLVal(val)
	case *sqlparser.NullVal:
		return tree.DNull
	case *sqlparser.FuncExpr:
		return convertFuncExpr(val)
	default:
		panic(fmt.Sprintf("unhandled type: %T", val))
	}
}

func convertFuncExpr(val *sqlparser.FuncExpr) tree.Expr {
	fnName := tree.NewUnresolvedName(val.Name.String())
	exprs := make(tree.Exprs, len(val.Exprs))

	for i, expr := range val.Exprs {
		e := convertSelectExpr(expr)
		exprs[i] = e.Expr
	}
	return &tree.FuncExpr{
		Func: tree.ResolvableFunctionReference{
			FunctionReference: fnName,
		},
		Exprs: exprs,
	}
}

func convertSelectExpr(expr sqlparser.SelectExpr) tree.SelectExpr {
	switch val := expr.(type) {
	case *sqlparser.AliasedExpr:
		e := convertExpr(val.Expr)
		return tree.SelectExpr{
			Expr: e,
			As:   tree.UnrestrictedName(val.As.String()),
		}
	case *sqlparser.StarExpr:
		return tree.SelectExpr{
			Expr: tree.StarExpr(),
		}
	default:
		panic(fmt.Sprintf("unhandled type: %T", val))
	}
}

func convertExprs(exprs sqlparser.Exprs) []tree.Expr {
	es := make([]tree.Expr, len(exprs))
	for i, expr := range exprs {
		es[i] = convertExpr(expr)
	}
	return es
}

func convertExpr(expr sqlparser.Expr) tree.Expr {
	switch val := expr.(type) {
	case nil:
		return nil
	case *sqlparser.SQLVal:
		return convertSQLVal(val)
	case *sqlparser.ColName:
		return tree.NewUnresolvedName(val.Name.String())
	case *sqlparser.FuncExpr:
		return convertFuncExpr(val)
	case *sqlparser.ValuesFuncExpr:
		return tree.NewStrVal(val.Name.String())
	case *sqlparser.BinaryExpr:
		return convertBinaryExpr(val)
	case *sqlparser.ComparisonExpr:
		return convertComparisonExpr(val)
	case *sqlparser.Subquery:
		return convertSubquery(val)
	case *sqlparser.ParenExpr:
		return convertExpr(val.Expr)
	case sqlparser.ValTuple:
		return convertValTuple(val)
	case *sqlparser.NullVal:
		return tree.DNull
	case sqlparser.BoolVal:
		boolVal := tree.DBool(bool(val))
		return &boolVal
	default:
		panic(fmt.Sprintf("unhandled type: %T", val))
	}
}

func convertValTuple(val sqlparser.ValTuple) tree.Expr {
	exprs := make([]tree.Expr, len(val))
	for i, expr := range val {
		exprs[i] = convertExpr(expr)
	}
	return &tree.Tuple{Exprs: exprs}
}

func convertSubquery(val *sqlparser.Subquery) tree.Expr {
	return &tree.Subquery{
		Select: &tree.ParenSelect{
			// TODO: order by, limit
			Select: &tree.Select{
				Select: convertSelectStatement(val.Select),
			},
		},
	}
}

func convertComparisonExpr(val *sqlparser.ComparisonExpr) tree.Expr {
	var op tree.ComparisonOperator
	switch val.Operator {
	case sqlparser.EqualStr:
		op = tree.EQ
	case sqlparser.LessThanStr:
		op = tree.LT
	case sqlparser.LessEqualStr:
		op = tree.LE
	case sqlparser.GreaterThanStr:
		op = tree.GT
	case sqlparser.GreaterEqualStr:
		op = tree.GE
	case sqlparser.NotEqualStr:
		op = tree.NE
	case sqlparser.InStr:
		op = tree.In
	case sqlparser.NotInStr:
		op = tree.NotIn
	case sqlparser.LikeStr:
		op = tree.Like
	case sqlparser.NotLikeStr:
		op = tree.NotLike
	case sqlparser.RegexpStr:
		op = tree.RegMatch
	case sqlparser.NotRegexpStr:
		op = tree.NotRegMatch
	default:
		panic(fmt.Sprintf("unhandled operator: %s", val.Operator))
	}

	return &tree.ComparisonExpr{
		Operator: op,
		Left:     convertExpr(val.Left),
		Right:    convertExpr(val.Right),
		// Fn:       nil,
	}
}

func convertBinaryExpr(val *sqlparser.BinaryExpr) tree.Expr {
	var op tree.BinaryOperator
	switch val.Operator {
	case sqlparser.BitAndStr:
		op = tree.Bitand
	case sqlparser.BitOrStr:
		op = tree.Bitor
	case sqlparser.BitXorStr:
		op = tree.Bitxor
	case sqlparser.PlusStr:
		op = tree.Plus
	case sqlparser.MinusStr:
		op = tree.Minus
	case sqlparser.MultStr:
		op = tree.Mult
	case sqlparser.DivStr:
		op = tree.Div
	case sqlparser.ModStr:
		op = tree.Mod
	case sqlparser.ShiftLeftStr:
		op = tree.LShift
	case sqlparser.ShiftRightStr:
		op = tree.RShift
	default:
		panic(fmt.Sprintf("unhandled operator: %s", val.Operator))
	}

	return &tree.BinaryExpr{
		Operator: op,
		Left:     convertExpr(val.Left),
		Right:    convertExpr(val.Right),
		// Fn:       nil,
	}
}

func convertSQLVal(val *sqlparser.SQLVal) tree.Expr {
	switch val.Type {
	case sqlparser.StrVal:
		return tree.NewStrVal(string(val.Val))
	case sqlparser.IntVal:
		i, err := strconv.Atoi(string(val.Val))
		if err != nil {
			panic(err)
		}
		return tree.NewDInt(tree.DInt(i))
	case sqlparser.FloatVal:
		f, err := strconv.ParseFloat(string(val.Val), 64)
		if err != nil {
			panic(err)
		}
		return tree.NewDFloat(tree.DFloat(f))
	case sqlparser.HexVal:
		return tree.NewStrVal(fmt.Sprintf("x'%s'", val.Val))
	case sqlparser.HexNum:
		return tree.NewStrVal(fmt.Sprintf("x'%s'", val.Val))
	default:
		panic(fmt.Sprintf("unhandled type: %v", val.Type))
	}
}

func transformDrop(query string, stmt *sqlparser.DDL) ([]string, bool) {
	// TODO
	return nil, false
}

func transformAlterTable(stmt *sqlparser.AlterTable) ([]string, bool) {
	var outputStmts []string
	for _, statement := range stmt.Statements {
		converted, ok := convertDdlStatement(statement)
		if !ok {
			return nil, false
		}
		outputStmts = append(outputStmts, converted...)
	}
	return outputStmts, true
}

func convertDdlStatement(statement *sqlparser.DDL) ([]string, bool) {
	switch statement.Action {
	case "alter":
		if statement.ColumnAction != "" {
			switch statement.ColumnAction {
			case "modify":
				if len(statement.TableSpec.Columns) != 1 {
					return nil, false
				}

				stmts := make([]string, 0)

				col := statement.TableSpec.Columns[0]
				tableName, err := tree.NewUnresolvedObjectName(1, [3]string{statement.Table.Name.String(), "", ""}, 0)
				if err != nil {
					panic(err)
				}

				newType := convertTypeDef(col.Type)
				alter := tree.AlterTable{
					Table: tableName,
					Cmds: []tree.AlterTableCmd{
						&tree.AlterTableAlterColumnType{
							Column: tree.Name(col.Name.String()),
							ToType: newType,
						},
					},
				}

				ctx := formatNodeWithUnqualifiedTableNames(&alter)
				stmts = append(stmts, ctx.String())

				// constraints
				if col.Type.NotNull {
					alter.Cmds = []tree.AlterTableCmd{
						&tree.AlterTableSetNotNull{
							Column: tree.Name(col.Name.String()),
						},
					}
					ctx = formatNodeWithUnqualifiedTableNames(&alter)
					stmts = append(stmts, ctx.String())
				} else {
					alter.Cmds = []tree.AlterTableCmd{
						&tree.AlterTableDropNotNull{
							Column: tree.Name(col.Name.String()),
						},
					}
					ctx = formatNodeWithUnqualifiedTableNames(&alter)
					stmts = append(stmts, ctx.String())
				}

				// rename
				if statement.Column.String() != col.Name.String() {
					alter.Cmds = []tree.AlterTableCmd{
						&tree.AlterTableRenameColumn{
							Column:  tree.Name(statement.Column.String()),
							NewName: tree.Name(col.Name.String()),
						},
					}
					ctx = formatNodeWithUnqualifiedTableNames(&alter)
					stmts = append(stmts, ctx.String())
				}

				return stmts, true
			default:
				return nil, false
			}
		}
		if statement.IndexSpec != nil {
			switch statement.IndexSpec.Action {
			case "drop":
				tableName := tree.NewTableName(tree.Name(""), tree.Name(statement.Table.Name.String()))
				indexName := statement.IndexSpec.ToName.String()
				if statement.IndexSpec.Type == "primary" {
					indexName = "PRIMARY"
				}
				dropIndex := tree.DropIndex{
					IndexList: tree.TableIndexNames{
						{
							Table: *tableName,
							Index: tree.UnrestrictedName(indexName),
						},
					},
				}

				ctx := formatNodeWithUnqualifiedTableNames(&dropIndex)
				return []string{ctx.String()}, true
			default:
				return nil, false
			}
		}

		return nil, false
	default:
		return nil, false
	}
}

// transformSelect converts a MySQL SELECT statement to a postgres-compatible SELECT statement.
// This is a very broad surface area, so we do this very selectively
func transformSelect(stmt *sqlparser.Select) ([]string, bool) {
	if !shouldRewriteSelect(stmt) {
		return nil, false
	}
	return []string{formatNode(stmt)}, true
}

func shouldRewriteSelect(stmt *sqlparser.Select) bool {
	return containsUserVars(stmt) ||
		containsBinaryConversion(stmt)
}

func containsBinaryConversion(stmt *sqlparser.Select) bool {
	foundBinaryConversionExpr := false
	fn := func(node sqlparser.SQLNode) (bool, error) {
		switch node := node.(type) {
		case *sqlparser.BinaryExpr:
			if node.Operator == "binary " {
				foundBinaryConversionExpr = true
				return false, nil
			}
		case *sqlparser.UnaryExpr:
			if node.Operator == "binary " {
				foundBinaryConversionExpr = true
				return false, nil
			}
		}
		return true, nil
	}

	for _, sel := range stmt.SelectExprs {
		sqlparser.Walk(fn, sel)
	}

	if stmt.Where != nil {
		sqlparser.Walk(fn, stmt.Where)
	}

	return foundBinaryConversionExpr
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
		if strings.HasPrefix(node.String(), "@@") {
			buf.Myprintf("current_setting('.%s')", strings.TrimLeft(node.String(), "@"))
		} else if strings.HasPrefix(node.String(), "@") {
			buf.Myprintf("current_setting('doltgres_enginetest.%s')", strings.TrimLeft(node.String(), "@"))
		} else {
			buf.Myprintf("%s", node.Lowered())
		}
	case *sqlparser.UnaryExpr:
		if node.Operator == "binary " {
			buf.Myprintf("%v::text::bytea", node.Expr)
		} else {
			buf.Myprintf("%v", node)
		}
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

var sequenceNum int

func transformCreateTable(stmt *sqlparser.DDL) ([]string, bool) {
	if stmt.TableSpec == nil {
		return nil, false
	}

	createTable := tree.CreateTable{
		IfNotExists: stmt.IfNotExists,
		Table:       tree.MakeTableNameWithSchema("", "", tree.Name(stmt.Table.Name.String())), // TODO: qualified names
	}

	var queries []string
	var autoIncColumn string
	for _, col := range stmt.TableSpec.Columns {
		defVal := convertExpr(col.Type.Default)

		if col.Type.Autoincrement {
			autoIncColumn = col.Name.String()
			defVal = &tree.FuncExpr{
				Func: tree.WrapFunction("nextval"),
				Exprs: []tree.Expr{
					tree.NewStrVal(fmt.Sprintf("seq_%d", sequenceNum)),
				},
			}
		}

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
				Expr:           defVal,
				ConstraintName: "", // TODO
			},
			CheckExprs: nil, // TODO
		})
	}

	// convert any primary key indexes
	if len(stmt.TableSpec.Indexes) > 0 {
		for _, index := range stmt.TableSpec.Indexes {
			if !index.Info.Primary {
				continue
			}

			indexCols := make(tree.IndexElemList, len(index.Columns))
			for i, col := range index.Columns {
				colName := col.Column.String()
				indexCols[i] = tree.IndexElem{
					Column: tree.Name(colName),
				}
			}

			indexDef := &tree.UniqueConstraintTableDef{
				PrimaryKey: true,
				IndexTableDef: tree.IndexTableDef{
					Columns: indexCols,
				},
			}

			createTable.Defs = append(createTable.Defs, indexDef)
		}
	}

	if autoIncColumn != "" {
		queries = append(queries, fmt.Sprintf("CREATE SEQUENCE seq_%d", sequenceNum))
		sequenceNum++
	}

	ctx := formatNodeWithUnqualifiedTableNames(&createTable)
	query := ctx.String()

	// this is a very odd quirk for only the char type, not sure why the postgres parser does this but it doesn't
	// parse in a CREATE TABLE statement
	query = strings.ReplaceAll(query, `"char"`, `char`)
	queries = append(queries, query)

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

	// convert constraints into separate statements as well
	for _, c := range stmt.TableSpec.Constraints {
		switch c := c.Details.(type) {
		case *sqlparser.ForeignKeyDefinition:
			queries = append(queries, createForeignKeyStatement(createTable.Table, c))
		case *sqlparser.CheckConstraintDefinition:
			queries = append(queries, createCheckConstraintStatement(createTable.Table, c))
		default:
			// do nothing, unsupported
		}
	}

	return queries, true
}

func createCheckConstraintStatement(table tree.TableName, c *sqlparser.CheckConstraintDefinition) string {
	name, err := tree.NewUnresolvedObjectName(1, [3]string{table.Table(), "", ""}, 0)
	if err != nil {
		panic(err)
	}

	alter := tree.AlterTable{
		Table: name,
	}

	alter.Cmds = append(alter.Cmds, &tree.AlterTableAddConstraint{
		ConstraintDef: &tree.CheckConstraintTableDef{
			Expr: convertExpr(c.Expr),
		},
	})

	ctx := formatNodeWithUnqualifiedTableNames(&alter)
	return ctx.String()
}

func createForeignKeyStatement(table tree.TableName, c *sqlparser.ForeignKeyDefinition) string {
	name, err := tree.NewUnresolvedObjectName(1, [3]string{table.Table(), "", ""}, 0)
	if err != nil {
		panic(err)
	}

	alter := tree.AlterTable{
		Table: name,
	}

	var fromCols, toCols tree.NameList
	for _, col := range c.Source {
		fromCols = append(fromCols, tree.Name(col.String()))
	}
	for _, col := range c.ReferencedColumns {
		toCols = append(toCols, tree.Name(col.String()))
	}

	onDelete := translateRefAction(c.OnDelete)
	onUpdate := translateRefAction(c.OnUpdate)

	alter.Cmds = append(alter.Cmds, &tree.AlterTableAddConstraint{
		ConstraintDef: &tree.ForeignKeyConstraintTableDef{
			FromCols: fromCols,
			Table:    tree.MakeTableName(tree.Name(""), tree.Name(c.ReferencedTable.Name.String())),
			ToCols:   toCols,
			Actions: tree.ReferenceActions{
				Delete: onDelete,
				Update: onUpdate,
			},
		},
	})

	ctx := formatNodeWithUnqualifiedTableNames(&alter)
	return ctx.String()
}

func translateRefAction(action sqlparser.ReferenceAction) tree.RefAction {
	switch action {
	case sqlparser.Cascade:
		return tree.RefAction{
			Action: tree.Cascade,
		}
	case sqlparser.SetNull:
		return tree.RefAction{
			Action: tree.SetNull,
		}
	case sqlparser.NoAction:
		return tree.RefAction{
			Action: tree.NoAction,
		}
	case sqlparser.Restrict:
		return tree.RefAction{
			Action: tree.Restrict,
		}
	case sqlparser.SetDefault:
		return tree.RefAction{
			Action: tree.SetDefault,
		}
	case sqlparser.DefaultAction:
		return tree.RefAction{
			Action: tree.Restrict, // is this correct?
		}
	default:
		panic(fmt.Sprintf("unhandled on delete action: %v", action))
	}
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
	case "boolean":
		return &types.T{
			InternalType: types.InternalType{
				Family: types.BoolFamily,
				Oid:    oid.T_bool,
			},
		}
	case "year":
		return &types.T{
			InternalType: types.InternalType{
				Family: types.IntFamily,
				Width:  16,
				Oid:    oid.T_int2,
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

// TestPostgresQueryFormat is a utility function to test how postgres parses and formats queries
func TestPostgresQueryFormat(t *testing.T) {
	type test struct {
		input    string
		expected string
	}
	tests := []test{
		{
			input:    "CREATE TABLE foo (a INT primary key default nextval('myseq'))",
			expected: "CREATE TABLE foo (a INTEGER DEFAULT nextval('myseq') PRIMARY KEY)",
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			s, err := parser.Parse(test.input)
			require.NoError(t, err)

			ctx := formatNodeWithUnqualifiedTableNames(s[0].AST)
			query := ctx.String()
			require.Equal(t, test.expected, query)
		})
	}
}

func TestConvertQuery(t *testing.T) {
	type test struct {
		input    string
		expected []string
		pattern  bool
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
			input: "CREATE TABLE test (pk BIGINT PRIMARY KEY AUTO_INCREMENT, v1 BIGINT);",
			// this is a pattern match because when run with other tests in the same process, the name of the sequence created is changed
			pattern: true,
			expected: []string{
				"CREATE SEQUENCE .+",
				"CREATE TABLE test \\(pk BIGINT NOT NULL DEFAULT nextval\\('.+'\\) PRIMARY KEY, v1 BIGINT NULL\\)"},
		},
		{
			input:    "CREATE TABLE foo (a INT, b int, primary key (b,a))",
			expected: []string{"CREATE TABLE foo (a INTEGER NULL, b INTEGER NULL, PRIMARY KEY (b, a))"},
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
		{
			input: "INSERT INTO foo (a, b) VALUES (1, 2), (3, 4) on duplicate key update a = 5",
			expected: []string{
				"INSERT INTO foo(a, b) VALUES (1, 2), (3, 4) ON CONFLICT (fake) DO UPDATE SET a = 5",
			},
		},
		{
			input: "INSERT INTO foo VALUES (1, 2), (3, 4) on duplicate key update a = 5",
			expected: []string{
				"INSERT INTO foo VALUES (1, 2), (3, 4) ON CONFLICT (fake) DO UPDATE SET a = 5",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			actual := convertQuery(test.input)
			if test.pattern {
				require.Equal(t, len(test.expected), len(actual))
				for i := range test.expected {
					require.Regexp(t, test.expected[i], actual[i])
				}
			} else {
				require.Equal(t, test.expected, actual)
			}
		})
	}
}

// TestBoolValSupport tests that query converter can handle boolean literals
// Related to issue #1708: Query converter crashes on boolean literals
func TestBoolValSupport(t *testing.T) {
	// This query contains boolean literals that should be converted properly
	// Before the fix: panics with "unhandled type: sqlparser.BoolVal"
	// After the fix: should convert successfully without panic
	result := convertQuery("CREATE TABLE test_table (id INT, archived BOOLEAN DEFAULT FALSE)")
	
	// Should not panic and should return converted query
	require.NotEmpty(t, result, "Query conversion should succeed and return converted SQL")
	require.Len(t, result, 1, "Should return exactly one converted statement")
	require.Contains(t, result[0], "CREATE TABLE", "Result should contain CREATE TABLE")
	require.Contains(t, result[0], "false", "Result should contain converted boolean literal")
}
