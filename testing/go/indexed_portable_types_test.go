// Copyright 2026 Dolthub, Inc.
// Licensed under the Apache License, Version 2.0.
package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestIndexedPortableTypes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "explicit OID payloads retain unsigned ordering",
			SetUpScript: []string{
				`CREATE TABLE objects(v oid PRIMARY KEY, label text)`,
				`INSERT INTO objects VALUES (4294967295,'max'),(23,'builtin'),(0,'zero'),(2147483648,'high')`,
			},
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT v::text AS value,label FROM objects ORDER BY v`, Expected: []sql.Row{{"0", "zero"}, {"23", "builtin"}, {"2147483648", "high"}, {"4294967295", "max"}}},
				{Query: `SELECT label FROM objects WHERE v>=2147483648::oid ORDER BY v DESC`, Expected: []sql.Row{{"max"}, {"high"}}},
				{Query: `INSERT INTO objects VALUES(23,'duplicate')`, ExpectedErrCode: "23505"},
				{Query: `UPDATE objects SET v=24 WHERE v=23`, SkipResultsCheck: true},
				{Query: `SELECT label FROM objects WHERE v=24::oid`, Expected: []sql.Row{{"builtin"}}},
				{Query: `DELETE FROM objects WHERE v=4294967295::oid`, SkipResultsCheck: true},
				{Query: `SELECT v::text AS value FROM objects ORDER BY v DESC`, Expected: []sql.Row{{"2147483648"}, {"24"}, {"0"}}},
			},
		},
		{
			Name: "bit string indexes preserve prefixes",
			SetUpScript: []string{
				`CREATE TABLE bits(v bit(4) PRIMARY KEY)`,
				`INSERT INTO bits VALUES(B'1111'),(B'0001'),(B'1000'),(B'0000')`,
				`CREATE TABLE varying(v varbit PRIMARY KEY)`,
				`INSERT INTO varying VALUES(B'1'),(B'10'),(B'0'),(B''),(B'00'),(B'100000001')`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// Doltgres does not yet resolve comparison operators on bit(n).
					Query:           `SELECT v::text AS value FROM bits WHERE v>=B'0001'::bit(4) ORDER BY v`,
					ExpectedErrCode: "42883",
					Postgres:        &ScriptTestAssertion{Expected: []sql.Row{{"0001"}, {"1000"}, {"1111"}}},
				},
				{Query: `SELECT v::text AS value FROM varying ORDER BY v`, Expected: []sql.Row{{""}, {"0"}, {"00"}, {"1"}, {"10"}, {"100000001"}}},
				{Query: `INSERT INTO varying VALUES(B'00')`, ExpectedErrCode: "23505"},
			},
		},
		{
			Name: "network indexes compare prefixes before host bits",
			SetUpScript: []string{
				`CREATE TABLE addresses(v inet PRIMARY KEY)`,
				`INSERT INTO addresses VALUES('10.9.8.7/8'),('10.0.0.0/16'),('::1'),('10.9.8.7/32')`,
			},
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT v::text AS value FROM addresses ORDER BY v`, Expected: []sql.Row{{"10.9.8.7/8"}, {"10.0.0.0/16"}, {"10.9.8.7/32"}, {"::1/128"}}},
				{Query: `SELECT v::text AS value FROM addresses WHERE v<'10.9.8.7'::inet ORDER BY v`, Expected: []sql.Row{{"10.9.8.7/8"}, {"10.0.0.0/16"}}},
			},
		},
	})
}

func TestIndexedStructuralTypes(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "JSONB index structural ordering and numeric equality",
			SetUpScript: []string{
				`CREATE TABLE docs(v jsonb PRIMARY KEY)`,
				`INSERT INTO docs VALUES ('{}'),('[]'),('null'),('1.0'),('false'),('[1]'),('{"a":2}')`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// Doltgres uses different JSON type ranking and normalizes numeric scale.
					// The PostgreSQL counterpart is verified against stock PostgreSQL 18.
					Query:    `SELECT v::text AS value FROM docs ORDER BY v`,
					Expected: []sql.Row{{"null"}, {"1"}, {"{}"}, {`{"a": 2}`}, {"[]"}, {"[1]"}, {"false"}},
					Postgres: &ScriptTestAssertion{Expected: []sql.Row{{"[]"}, {"null"}, {"1.0"}, {"false"}, {"[1]"}, {"{}"}, {`{"a": 2}`}}},
				},
				{Query: `INSERT INTO docs VALUES ('1.00')`, ExpectedErrCode: "23505"},
				{
					Query:    `SELECT v::text AS value FROM docs WHERE v>'false'::jsonb ORDER BY v`,
					Expected: []sql.Row{},
					Postgres: &ScriptTestAssertion{Expected: []sql.Row{{"[1]"}, {"{}"}, {`{"a": 2}`}}},
				},
			},
		},
		{
			Name: "composite index uses field ordering",
			SetUpScript: []string{
				`CREATE TYPE pair AS (n numeric,s text)`,
				`CREATE TABLE records(v pair PRIMARY KEY)`,
				`INSERT INTO records VALUES ('(10,a)'),('(2,b)'),('(2,a)'),('(,)')`,
			},
			Assertions: []ScriptTestAssertion{
				{
					// Native record comparison places NULL fields after non-NULL fields.
					Query:    `SELECT v::text AS value FROM records ORDER BY v`,
					Expected: []sql.Row{{"(,)"}, {"(2,a)"}, {"(2,b)"}, {"(10,a)"}},
					Postgres: &ScriptTestAssertion{Expected: []sql.Row{{"(2,a)"}, {"(2,b)"}, {"(10,a)"}, {"(,)"}}},
				},
				{Query: `INSERT INTO records VALUES ('(2.00,a)')`, ExpectedErrCode: "23505"},
				{Query: `SELECT v::text AS value FROM records WHERE v<'(10,a)'::pair ORDER BY v`, Expected: []sql.Row{{"(2,a)"}, {"(2,b)"}}},
			},
		},
		{
			Name: "range index compares bound values and inclusivity",
			// Doltgres does not implement the numrange type.
			Skip: testTarget(t) == targetDoltgres,
			SetUpScript: []string{
				`CREATE TABLE spans(v numrange PRIMARY KEY)`,
				`INSERT INTO spans VALUES ('empty'),('(,)'),('[1,2)'),('(1,2)'),('[1,2]'),('[2,3)')`,
			},
			Assertions: []ScriptTestAssertion{
				{Query: `SELECT v::text AS value FROM spans ORDER BY v`, Expected: []sql.Row{{"empty"}, {"(,)"}, {"[1,2)"}, {"[1,2]"}, {"(1,2)"}, {"[2,3)"}}},
				{Query: `INSERT INTO spans VALUES ('[1.00,2.0)')`, ExpectedErrCode: "23505"},
				{Query: `SELECT v::text AS value FROM spans WHERE v>'[1,2)'::numrange ORDER BY v`, Expected: []sql.Row{{"[1,2]"}, {"(1,2)"}, {"[2,3)"}}},
			},
		},
	})
}
