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

package _go

import (
	"context"
	"testing"

	"github.com/dolthub/dolt/go/libraries/utils/svcs"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMultipleStatements is a test for: https://github.com/dolthub/doltgresql/issues/2175
func TestMultipleStatements(t *testing.T) {
	ctx := context.Background()
	var conn *Connection
	if runOnPostgres {
		pgxConn, err := pgx.Connect(ctx, "postgres://postgres:password@127.0.0.1:5432/postgres?sslmode=disable")
		require.NoError(t, err)
		conn = &Connection{
			Default: pgxConn,
			Current: pgxConn,
		}
		require.NoError(t, pgxConn.Ping(ctx))
		defer func() {
			conn.Close(ctx)
		}()
	} else {
		var controller *svcs.Controller
		ctx, conn, controller = CreateServer(t, "postgres")
		defer func() {
			conn.Close(ctx)
			controller.Stop()
			err := controller.WaitForStop()
			require.NoError(t, err)
		}()
	}
	queries := []string{
		`BEGIN;`,
		`DROP TABLE IF EXISTS migrations;`,
		`DROP TABLE IF EXISTS animals;`,
		`CREATE TABLE IF NOT EXISTS migrations (file_name TEXT NOT NULL, file_hash TEXT NOT NULL);`,
		`CREATE TABLE IF NOT EXISTS animals (id SERIAL PRIMARY KEY NOT NULL, name TEXT NOT NULL);`,
		`;`, // This should be ignored in the output
		`INSERT INTO migrations (file_name, file_hash) VALUES ('2021-09-07T154500-create-animals-table.sql', '42331f4277227d09e9bb32eeaf7e04d9c7fe320160e05372ed0ef010cfbf666b');`,
		`INSERT INTO animals(name) VALUES('Alpaca');`,
		`INSERT INTO animals(name) VALUES('Highland cow');`,
		`INSERT INTO animals(name) VALUES('Aardvark');`,
		`INSERT INTO migrations (file_name, file_hash) VALUES ('2021-09-07T154700-insert-animals.sql', '3223d0deb6fb7fb2accf6abffc0667ebe4503379987c472d10a585a553f9b3b6');`,
		`SELECT * FROM migrations ORDER BY file_name;`,
		`SELECT * FROM animals ORDER BY id;`,
		`COMMIT;`,
	}
	combinedQueries := ""
	for _, query := range queries {
		// We do this just to homogenize the queries, even though we're adding the delimiter right back
		query = sql.RemoveSpaceAndDelimiter(query, ';')
		combinedQueries += query + ";"
	}
	// First we'll check all invalid modes that fail immediately
	invalidModes := []pgx.QueryExecMode{
		pgx.QueryExecModeCacheStatement,
		pgx.QueryExecModeCacheDescribe,
		pgx.QueryExecModeDescribeExec,
		pgx.QueryExecModeExec,
	}
	for _, mode := range invalidModes {
		rows, err := conn.Current.Query(ctx, combinedQueries, mode)
		if mode == pgx.QueryExecModeExec {
			// This mode requires reading from the returned rows to find the error, rather than erroring immediately
			require.NoError(t, err)
			_ = rows.Next()
			err = rows.Err()
		} else {
			require.Error(t, err)
		}
		require.Contains(t, err.Error(), "cannot insert multiple commands into a prepared statement")
	}
	// Then we'll check the singular valid mode
	rows, err := conn.Current.Query(ctx, combinedQueries, pgx.QueryExecModeSimpleProtocol)
	require.NoError(t, err)
	require.False(t, rows.Next()) // Simple mode doesn't return results with multiple statements
	rows.Close()
	// Now we'll use the underlying connection to verify all returned results
	mrr := conn.Current.PgConn().Exec(ctx, combinedQueries)
	results, err := mrr.ReadAll()
	require.NoError(t, err)
	if assert.Len(t, results, len(testMultipleStatementsResults)) {
		for resultIdx, expected := range testMultipleStatementsResults {
			result := results[resultIdx]
			if assert.Equal(t, len(expected.FieldDescriptions), len(result.FieldDescriptions)) {
				for fieldIdx, expectedField := range expected.FieldDescriptions {
					resultField := result.FieldDescriptions[fieldIdx]
					assert.Equal(t, expectedField.Name, resultField.Name)
					assert.Equal(t, expectedField.DataTypeOID, resultField.DataTypeOID)
					assert.Equal(t, expectedField.DataTypeSize, resultField.DataTypeSize)
					assert.Equal(t, expectedField.TypeModifier, resultField.TypeModifier)
					assert.Equal(t, expectedField.Format, resultField.Format)
				}
			}
			if assert.Equal(t, len(expected.Rows), len(result.Rows)) {
				for rowIdx, expectedRow := range expected.Rows {
					resultRow := result.Rows[rowIdx]
					for columnIdx, expectedCol := range expectedRow {
						assert.Equal(t, expectedCol, resultRow[columnIdx])
					}
				}
			}
			assert.Equal(t, expected.CommandTag, result.CommandTag)
		}
	}
	require.NoError(t, mrr.Close())

	// Now we'll ensure that errors are properly handled within multiple statements
	queries = []string{
		`INSERT INTO animals(name) VALUES('Pigeon');`,
		`SELECT * FROM non_existent;`,
		`INSERT INTO animals(name) VALUES('Elephant');`,
	}
	combinedQueries = ""
	for _, query := range queries {
		query = sql.RemoveSpaceAndDelimiter(query, ';')
		combinedQueries += query + ";"
	}
	mrr = conn.Current.PgConn().Exec(ctx, combinedQueries)
	results, err = mrr.ReadAll()
	require.Error(t, err)
	require.Contains(t, err.Error(), "non_existent")
	if assert.Len(t, results, 1) {
		assert.Equal(t, results[0].CommandTag, pgconn.NewCommandTag("INSERT 0 1"))
	}
}

// testMultipleStatementsResults are used within TestMultipleStatements
var testMultipleStatementsResults = []pgconn.Result{
	{CommandTag: pgconn.NewCommandTag("BEGIN")},
	{CommandTag: pgconn.NewCommandTag("DROP TABLE")},
	{CommandTag: pgconn.NewCommandTag("DROP TABLE")},
	{CommandTag: pgconn.NewCommandTag("CREATE TABLE")},
	{CommandTag: pgconn.NewCommandTag("CREATE TABLE")},
	{CommandTag: pgconn.NewCommandTag("INSERT 0 1")},
	{CommandTag: pgconn.NewCommandTag("INSERT 0 1")},
	{CommandTag: pgconn.NewCommandTag("INSERT 0 1")},
	{CommandTag: pgconn.NewCommandTag("INSERT 0 1")},
	{CommandTag: pgconn.NewCommandTag("INSERT 0 1")},
	{
		FieldDescriptions: []pgconn.FieldDescription{
			{
				Name:         "file_name",
				DataTypeOID:  25,
				DataTypeSize: -1,
				TypeModifier: -1,
				Format:       0,
			},
			{
				Name:         "file_hash",
				DataTypeOID:  25,
				DataTypeSize: -1,
				TypeModifier: -1,
				Format:       0,
			},
		},
		Rows: [][][]byte{
			{
				[]byte("2021-09-07T154500-create-animals-table.sql"),
				[]byte("42331f4277227d09e9bb32eeaf7e04d9c7fe320160e05372ed0ef010cfbf666b"),
			},
			{
				[]byte("2021-09-07T154700-insert-animals.sql"),
				[]byte("3223d0deb6fb7fb2accf6abffc0667ebe4503379987c472d10a585a553f9b3b6"),
			},
		},
		CommandTag: pgconn.NewCommandTag("SELECT 2"),
	},
	{
		FieldDescriptions: []pgconn.FieldDescription{
			{
				Name:         "id",
				DataTypeOID:  23,
				DataTypeSize: 4,
				TypeModifier: -1,
				Format:       0,
			},
			{
				Name:         "name",
				DataTypeOID:  25,
				DataTypeSize: -1,
				TypeModifier: -1,
				Format:       0,
			},
		},
		Rows: [][][]byte{
			{
				[]byte("1"),
				[]byte("Alpaca"),
			},
			{
				[]byte("2"),
				[]byte("Highland cow"),
			},
			{
				[]byte("3"),
				[]byte("Aardvark"),
			},
		},
		CommandTag: pgconn.NewCommandTag("SELECT 3"),
	},
	{CommandTag: pgconn.NewCommandTag("COMMIT")},
}
