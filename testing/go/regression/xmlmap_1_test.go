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

package regression

import (
	"testing"
)

func TestXmlmap1(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_xmlmap_1)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_xmlmap_1,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE SCHEMA testxmlschema;`,
			},
			{
				Statement: `CREATE TABLE testxmlschema.test1 (a int, b text);`,
			},
			{
				Statement: `INSERT INTO testxmlschema.test1 VALUES (1, 'one'), (2, 'two'), (-1, null);`,
			},
			{
				Statement: `CREATE DOMAIN testxmldomain AS varchar;`,
			},
			{
				Statement: `CREATE TABLE testxmlschema.test2 (z int, y varchar(500), x char(6),
    w numeric(9,2), v smallint, u bigint, t real,
    s time, stz timetz, r timestamp, rtz timestamptz, q date,
    p xml, o testxmldomain, n bool, m bytea, aaa text);`,
			},
			{
				Statement: `ALTER TABLE testxmlschema.test2 DROP COLUMN aaa;`,
			},
			{
				Statement: `INSERT INTO testxmlschema.test2 VALUES (55, 'abc', 'def',
    98.6, 2, 999, 0,
    '21:07', '21:11 +05', '2009-06-08 21:07:30', '2009-06-08 21:07:30 -07', '2009-06-08',
    NULL, 'ABC', true, 'XYZ');`,
			},
			{
				Statement:   `SELECT table_to_xml('testxmlschema.test1', false, false, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT table_to_xml('testxmlschema.test1', true, false, 'foo');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT table_to_xml('testxmlschema.test1', false, true, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT table_to_xml('testxmlschema.test1', true, true, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT table_to_xml('testxmlschema.test2', false, false, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT table_to_xmlschema('testxmlschema.test1', false, false, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT table_to_xmlschema('testxmlschema.test1', true, false, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT table_to_xmlschema('testxmlschema.test1', false, true, 'foo');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT table_to_xmlschema('testxmlschema.test1', true, true, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT table_to_xmlschema('testxmlschema.test2', false, false, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT table_to_xml_and_xmlschema('testxmlschema.test1', false, false, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT table_to_xml_and_xmlschema('testxmlschema.test1', true, false, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT table_to_xml_and_xmlschema('testxmlschema.test1', false, true, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT table_to_xml_and_xmlschema('testxmlschema.test1', true, true, 'foo');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT query_to_xml('SELECT * FROM testxmlschema.test1', false, false, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT query_to_xmlschema('SELECT * FROM testxmlschema.test1', false, false, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT query_to_xml_and_xmlschema('SELECT * FROM testxmlschema.test1', true, true, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `DECLARE xc CURSOR WITH HOLD FOR SELECT * FROM testxmlschema.test1 ORDER BY 1, 2;`,
			},
			{
				Statement:   `SELECT cursor_to_xml('xc'::refcursor, 5, false, true, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT cursor_to_xmlschema('xc'::refcursor, false, true, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `MOVE BACKWARD ALL IN xc;`,
			},
			{
				Statement:   `SELECT cursor_to_xml('xc'::refcursor, 5, true, false, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT cursor_to_xmlschema('xc'::refcursor, true, false, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT schema_to_xml('testxmlschema', false, true, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT schema_to_xml('testxmlschema', true, false, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT schema_to_xmlschema('testxmlschema', false, true, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT schema_to_xmlschema('testxmlschema', true, false, '');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT schema_to_xml_and_xmlschema('testxmlschema', true, true, 'foo');`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement: `CREATE DOMAIN testboolxmldomain AS bool;`,
			},
			{
				Statement: `CREATE DOMAIN testdatexmldomain AS date;`,
			},
			{
				Statement: `CREATE TABLE testxmlschema.test3
    AS SELECT true c1,
              true::testboolxmldomain c2,
              '2013-02-21'::date c3,
              '2013-02-21'::testdatexmldomain c4;`,
			},
			{
				Statement:   `SELECT xmlforest(c1, c2, c3, c4) FROM testxmlschema.test3;`,
				ErrorString: `unsupported XML feature`,
			},
			{
				Statement:   `SELECT table_to_xml('testxmlschema.test3', true, true, '');`,
				ErrorString: `unsupported XML feature`,
			},
		},
	})
}
