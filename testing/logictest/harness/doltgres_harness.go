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

package harness

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/dolthub/sqllogictest/go/logictest"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var _ logictest.Harness = &PostgresqlServerHarness{}

// sqllogictest harness for postgres databases.
type PostgresqlServerHarness struct {
	dsn string
	db  *sql.DB
}

// compile check for interface compliance
var _ logictest.Harness = &PostgresqlServerHarness{}

// NewPostgresqlHarness returns a new Postgres test harness for the data source name given. Panics if it cannot open a
// connection using the DSN.
func NewPostgresqlHarness(dsn string) *PostgresqlServerHarness {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		panic(err)
	}
	return &PostgresqlServerHarness{
		dsn: dsn,
		db:  db,
	}
}

func (h *PostgresqlServerHarness) EngineStr() string {
	return "postgresql"
}

func (h *PostgresqlServerHarness) Init() error {
	if err := h.dropAllTables(); err != nil {
		return err
	}

	return h.dropAllViews()
}

// See Harness.ExecuteStatement
func (h *PostgresqlServerHarness) ExecuteStatement(statement string) error {
	_, err := h.db.Exec(statement)
	return err
}

// See Harness.ExecuteQuery
func (h *PostgresqlServerHarness) ExecuteQuery(statement string) (schema string, results []string, err error) {
	rows, err := h.db.Query(statement)
	if rows != nil {
		defer rows.Close()
	}

	if err != nil {
		return "", nil, err
	}

	schema, columns, err := columns(rows)
	if err != nil {
		return "", nil, err
	}

	for rows.Next() {
		err := rows.Scan(columns...)
		if err != nil {
			return "", nil, err
		}

		for _, col := range columns {
			results = append(results, stringVal(col))
		}
	}

	if rows.Err() != nil {
		return "", nil, rows.Err()
	}

	return schema, results, nil
}

func (h *PostgresqlServerHarness) dropAllTables() error {
	var rows *sql.Rows
	var err error
	rows, err = h.db.QueryContext(context.Background(), "SELECT table_name FROM information_schema.tables WHERE table_schema = 'sqllogictest' AND table_type = 'BASE TABLE';")
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return err
	}

	_, columns, err := columns(rows)
	if err != nil {
		return err
	}

	var tableNames []string
	for rows.Next() {
		err := rows.Scan(columns...)
		if err != nil {
			return err
		}

		tableName := columns[0].(*sql.NullString)
		tableNames = append(tableNames, tableName.String)
	}

	if len(tableNames) > 0 {
		dropTables := "drop table if exists " + strings.Join(tableNames, ",")
		_, err = h.db.Exec(dropTables)
		if err != nil {
			return err
		}
	}

	return nil
}

func (h *PostgresqlServerHarness) dropAllViews() error {
	rows, err := h.db.QueryContext(context.Background(), "select table_name from INFORMATION_SCHEMA.views")
	if rows != nil {
		defer rows.Close()
	}
	if err != nil {
		return err
	}

	_, columns, err := columns(rows)
	if err != nil {
		return err
	}

	var viewNames []string
	for rows.Next() {
		err := rows.Scan(columns...)
		if err != nil {
			return err
		}

		viewName := columns[0].(*sql.NullString)
		viewNames = append(viewNames, viewName.String)
	}

	if len(viewNames) > 0 {
		dropView := "drop view if exists " + strings.Join(viewNames, ",")
		_, err = h.db.Exec(dropView)
		if err != nil {
			return err
		}
	}

	return nil
}

// Returns the string representation of the column value given
func stringVal(col interface{}) string {
	switch v := col.(type) {
	case *sql.NullBool:
		if !v.Valid {
			return "NULL"
		}
		if v.Bool {
			return "1"
		} else {
			return "0"
		}
	case *sql.NullInt64:
		if !v.Valid {
			return "NULL"
		}
		return fmt.Sprintf("%d", v.Int64)
	case *sql.NullFloat64:
		if !v.Valid {
			return "NULL"
		}
		return fmt.Sprintf("%.3f", v.Float64)
	case *sql.NullString:
		if !v.Valid {
			return "NULL"
		}
		return v.String
	default:
		panic(fmt.Sprintf("unhandled type %T for value %v", v, v))
	}
}

// Returns the schema for the rows given, as well as a slice of columns suitable for scanning values into.
func columns(rows *sql.Rows) (string, []interface{}, error) {
	types, err := rows.ColumnTypes()
	if err != nil {
		return "", nil, err
	}

	sb := strings.Builder{}
	var columns []interface{}
	for _, columnType := range types {
		switch columnType.DatabaseTypeName() {
		case "BIT":
			colVal := sql.NullBool{}
			columns = append(columns, &colVal)
			sb.WriteString("I")
		case "TEXT", "VARCHAR", "MEDIUMTEXT", "CHAR", "TINYTEXT", "NAME", "BYTEA":
			colVal := sql.NullString{}
			columns = append(columns, &colVal)
			sb.WriteString("T")
		case "DECIMAL", "DOUBLE", "FLOAT", "FLOAT8", "NUMERIC":
			colVal := sql.NullFloat64{}
			columns = append(columns, &colVal)
			sb.WriteString("R")
		case "MEDIUMINT", "INT", "BIGINT", "TINYINT", "SMALLINT", "INT4", "INT8":
			colVal := sql.NullInt64{}
			columns = append(columns, &colVal)
			sb.WriteString("I")
		case "UNKNOWN": // used for NULL values
			colVal := sql.NullString{}
			columns = append(columns, &colVal)
			sb.WriteString("I") // is this right?
		default:
			return "", nil, fmt.Errorf("Unhandled type %s", columnType.DatabaseTypeName())
		}
	}

	return sb.String(), columns, nil
}
