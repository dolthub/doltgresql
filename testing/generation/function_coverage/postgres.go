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

package main

import (
	"context"
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5"
)

var postgresConnection *pgx.Conn

// QueryPostgres runs the query on a Postgres server and returns the results, along with the OIDs of each result column.
// This assumes that a valid Postgres (not Doltgres) instance is on port 5432.
func QueryPostgres(query string) ([]sql.Row, []uint32, error) {
	var err error
	ctx := context.Background()
	if postgresConnection == nil {
		connectionString := fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/", 5432)
		postgresConnection, err = pgx.Connect(ctx, connectionString)
		if err != nil {
			return nil, nil, err
		}
	}
	r, err := postgresConnection.Query(ctx, query)
	if err != nil {
		return nil, nil, err
	}
	var oids []uint32
	for _, desc := range r.FieldDescriptions() {
		oids = append(oids, desc.DataTypeOID)
	}
	var allRows []sql.Row
	for r.Next() {
		if err = r.Err(); err != nil {
			return nil, nil, err
		}
		row, err := r.Values()
		if err != nil {
			return nil, nil, err
		}
		allRows = append(allRows, row)
	}
	return allRows, oids, r.Err()
}

// ExecPostgres runs the query on a Postgres server without checking the results. This assumes that a valid Postgres
// (not Doltgres) instance is on port 5432.
func ExecPostgres(query string) error {
	var err error
	ctx := context.Background()
	if postgresConnection == nil {
		connectionString := fmt.Sprintf("postgres://postgres:password@127.0.0.1:%d/", 5432)
		postgresConnection, err = pgx.Connect(ctx, connectionString)
		if err != nil {
			return err
		}
	}
	_, err = postgresConnection.Exec(ctx, query)
	return err
}
