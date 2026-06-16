// Copyright 2026 Dolthub, Inc.
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

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

type ResFunc func(rows pgx.Rows) error

type StmtTest struct {
	Name  string
	Query string
	Args  []any
	Res   []ResFunc
}

func main() {
	user := os.Args[1]
	port := os.Args[2]

	ctx := context.Background()
	connStr := fmt.Sprintf("postgres://%s:password@localhost:%s/postgres", user, port)
	conn, err := pgx.Connect(ctx, connStr)
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	var pk int
	err = conn.QueryRow(ctx, "SELECT pk FROM test_table LIMIT 1").Scan(&pk)
	if err != nil {
		panic(err)
	}
	if pk != 1 {
		panic(fmt.Sprintf("expected pk=1, got %d", pk))
	}

	_, err = conn.Exec(ctx, "INSERT INTO test_table VALUES (2)")
	if err != nil {
		panic(err)
	}

	var count int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM test_table").Scan(&count)
	if err != nil {
		panic(err)
	}
	if count != 2 {
		panic(fmt.Sprintf("expected count=2, got %d", count))
	}

	_, err = conn.Prepare(ctx, "select_pk", "SELECT pk FROM test_table WHERE pk = $1")
	if err != nil {
		panic(err)
	}
	err = conn.QueryRow(ctx, "select_pk", 1).Scan(&pk)
	if err != nil {
		panic(err)
	}
	if pk != 1 {
		panic(fmt.Sprintf("expected pk=1 from prepared stmt, got %d", pk))
	}

	// dolt workflow: create table, insert, commit, branch, insert, commit, merge
	for _, q := range []string{
		"DROP TABLE IF EXISTS test",
		"CREATE TABLE test (pk int, value int, PRIMARY KEY(pk))",
		"INSERT INTO test (pk, value) VALUES (0, 0)",
		"SELECT dolt_add('-A')",
		"SELECT dolt_commit('-m', 'added table test')",
		"SELECT dolt_checkout('-b', 'mybranch')",
		"INSERT INTO test VALUES (1, 1)",
		"SELECT dolt_commit('-a', '-m', 'updated test')",
		"SELECT dolt_checkout('main')",
		"SELECT dolt_merge('mybranch')",
	} {
		if _, err = conn.Exec(ctx, q); err != nil {
			panic(fmt.Sprintf("failed to execute %q: %v", q, err))
		}
	}

	stmtTests := []StmtTest{
		{
			Name:  "select_test_by_pk",
			Query: "SELECT pk, value FROM test WHERE pk = $1",
			Args:  []any{int64(0)},
			Res: []ResFunc{
				func(rows pgx.Rows) error {
					var pk, value int64
					if err := rows.Scan(&pk, &value); err != nil {
						return err
					}
					if pk != 0 || value != 0 {
						return fmt.Errorf("expected pk=0 value=0, got pk=%d value=%d", pk, value)
					}
					return nil
				},
			},
		},
		{
			Name:  "count_dolt_log",
			Query: "SELECT COUNT(*) FROM dolt_log",
			Res: []ResFunc{
				func(rows pgx.Rows) error {
					var size int64
					if err := rows.Scan(&size); err != nil {
						return err
					}
					if size != 4 {
						return fmt.Errorf("expected 4 dolt_log entries, got %d", size)
					}
					return nil
				},
			},
		},
		{
			Name:  "count_test",
			Query: "SELECT COUNT(*) FROM test",
			Res: []ResFunc{
				func(rows pgx.Rows) error {
					var size int64
					if err := rows.Scan(&size); err != nil {
						return err
					}
					if size != 2 {
						return fmt.Errorf("expected 2 rows in test, got %d", size)
					}
					return nil
				},
			},
		},
	}

	for _, test := range stmtTests {
		func() {
			if _, err := conn.Prepare(ctx, test.Name, test.Query); err != nil {
				panic(fmt.Sprintf("prepare %q: %v", test.Query, err))
			}
			rows, err := conn.Query(ctx, test.Name, test.Args...)
			if err != nil {
				panic(fmt.Sprintf("query %q: %v", test.Query, err))
			}
			defer rows.Close()

			i := 0
			for rows.Next() {
				if i >= len(test.Res) {
					panic(fmt.Sprintf("too many rows for %q", test.Query))
				}
				if err := test.Res[i](rows); err != nil {
					panic(fmt.Sprintf("result %d of %q: %v", i, test.Query, err))
				}
				i++
			}
			if err := rows.Err(); err != nil {
				panic(fmt.Sprintf("rows.Err() for %q: %v", test.Query, err))
			}
		}()
	}

	fmt.Println("pgx test passed")
}
