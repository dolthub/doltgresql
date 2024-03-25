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

package main

import (
	"strings"

	"github.com/dolthub/doltgresql/testing/generation/utils"
)

// GlobalCustomVariables are variable definitions that are used when a synopsis does not define the definition itself,
// and there isn't a more specific definition in PrefixCustomVariables.
var GlobalCustomVariables = map[string]utils.StatementGenerator{
	"access_method_type":  customDefinition(`TABLE | INDEX`),
	"argmode":             customDefinition(`IN | VARIADIC`),
	"argtype":             customDefinition(`FLOAT8`),
	"boolean":             customDefinition(`true`),
	"cache":               customDefinition(`1`),
	"code":                customDefinition(`'code'`),
	"collatable":          customDefinition(`true`),
	"collation":           customDefinition(`en_US`),
	"column_definition":   customDefinition(`v1 INTEGER`),
	"column_number":       customDefinition(`1`),
	"connlimit":           customDefinition(`-1`),
	"cycle_mark_default":  customDefinition(`'cycle_mark_default'`),
	"cycle_mark_value":    customDefinition(`'cycle_mark_value'`),
	"delete":              customDefinition(`DELETE FROM tablename`),
	"dest_encoding":       customDefinition(`'UTF8'`),
	"domain_constraint":   customDefinition(`CONSTRAINT name CHECK (condition)`),
	"element":             customDefinition(`element_type`),
	"execution_cost":      customDefinition(`10`),
	"existing_enum_value": customDefinition(`'1'`),
	"filter_value":        customDefinition(`'Active'`),
	"from_item_recursive": customDefinition(`function_name()`),
	"increment":           customDefinition(`1`),
	"insert":              customDefinition(`INSERT INTO tablename VALUES (1)`),
	"integer":             customDefinition(`1`),
	"internallength":      customDefinition(`16`),
	"join_type":           customDefinition(`[ INNER ] JOIN | LEFT [ OUTER ] JOIN | RIGHT [ OUTER ] JOIN | FULL [ OUTER ] JOIN`),
	"large_object_oid":    customDefinition(`99999`),
	"loid":                customDefinition(`99999`),
	"maxvalue":            customDefinition(`1`),
	"minvalue":            customDefinition(`1`),
	"neighbor_enum_value": customDefinition(`'1'`),
	"new_enum_value":      customDefinition(`'1'`),
	"numeric_literal":     customDefinition(`1`),
	"operator":            customDefinition(`+`),
	"output_expression":   customDefinition(`colname`),
	"payload":             customDefinition(`'payload'`),
	"preferred":           customDefinition(`true`),
	"query":               customDefinition(`SELECT 1`),
	"restart":             customDefinition(`0`),
	"select":              customDefinition(`SELECT 1`),
	"sequence_options":    customDefinition(`NO MINVALUE`),
	"snapshot_id":         customDefinition(`'snapshot_id'`),
	"source_encoding":     customDefinition(`'UTF8'`),
	"source_query":        customDefinition(`SELECT 1`),
	"sql_body":            customDefinition(`BEGIN ATOMIC END | RETURN 1`),
	"start":               customDefinition(`0`),
	"storage_parameter":   customDefinition(`fillfactor`),
	"strategy_number":     customDefinition(`3`),
	"string_literal":      customDefinition(`'str'`),
	"sub-SELECT":          customDefinition(`SELECT 1`),
	"support_number":      customDefinition(`3`),
	"transaction_id":      customDefinition(`'id'`),
	"oid":                 customDefinition(`99999`),
	"operator_name":       customDefinition(`@@`),
	"result_rows":         customDefinition(`10`),
	"uid":                 customDefinition(`1`),
	"update":              customDefinition(`UPDATE tablename SET x = 1`),
	"values":              customDefinition(`VALUES (1)`),
	"with_query":          customDefinition(`queryname AS (select)`),
}

var PrefixCustomVariables = map[string]map[string]utils.StatementGenerator{
	"ALTER FOREIGN TABLE": {
		"index_parameters": customDefinition(`USING INDEX TABLESPACE tablespace_name`),
	},
	"ALTER OPERATOR": {
		"name": customDefinition(`@@`),
	},
	"ALTER STATISTICS": {
		"new_target": customDefinition(`1`),
	},
	"CREATE OPERATOR": {
		"name": customDefinition(`@@`),
	},
	"CREATE RULE": {
		"command": customDefinition(`SELECT 'abc'`),
	},
	"CREATE SCHEMA": {
		"schema_element": customDefinition(`CREATE TABLE tablename()`),
	},
	"DROP OPERATOR": {
		"name": customDefinition(`@@`),
	},
	"EXPLAIN": {
		"statement": customDefinition(`SELECT 1 | INSERT INTO tablename VALUES (1)`),
	},
	"MERGE": {
		"query": customDefinition(`SELECT 1`),
	},
	"PREPARE": {
		"statement": customDefinition(`SELECT 1 | INSERT INTO tablename VALUES (1)`),
	},
	"SET": {
		"value": customDefinition(`1`),
	},
}

// customDefinition returns a StatementGenerator for a custom variable definition. The variable definition should follow
// the same layout format as synopses.
func customDefinition(str string) utils.StatementGenerator {
	str = strings.TrimSpace(str)
	scanner := NewScanner(str)
	tokens, err := scanner.Process()
	if err != nil {
		panic(err)
	}
	stmtGen, err := utils.ParseTokens(tokens, true)
	if err != nil {
		panic(err)
	}
	if stmtGen == nil {
		panic("definition did not create a statement generator")
	}
	return stmtGen
}
