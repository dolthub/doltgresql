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

//! Dolt's tests: the dolt_tests table, which holds a schema's test queries and their assertions, and DOLT_TEST_RUN,
//! which runs them, as Dolt's TestsTable and TestsRunTableFunction do.

use chrono::{DateTime, NaiveDate, NaiveDateTime};
use pg_query::NodeEnum;

use crate::catalog::table::TableDef;
use crate::dolt::args::error;
use crate::error::Result;
use crate::numeric::Numeric;
use crate::parse::Statement;
use crate::query::Ctx;
use crate::types::Value;
use crate::{Outcome, oid};

/// TABLE is the name of the table that holds a schema's tests.
pub const TABLE: &str = "dolt_tests";

/// DEFINITION creates the tests table with Dolt's columns and the checks that Dolt enforces on writes to it.
const DEFINITION: &str = "(test_name text PRIMARY KEY, test_group text, test_query text NOT NULL, \
     assertion_type text NOT NULL, assertion_comparator text NOT NULL, assertion_value text, \
     CONSTRAINT assertion_type_check CHECK (assertion_type IN ('expected_rows', 'expected_columns', \
     'expected_single_value')), \
     CONSTRAINT assertion_comparator_check CHECK (assertion_comparator IN ('==', '!=', '<', '>', '<=', '>=')))";

/// COLUMNS are the tests table's columns, which it shows before anything creates it.
pub const COLUMNS: &[(&str, u32)] = &[
    ("test_name", oid::TEXT),
    ("test_group", oid::TEXT),
    ("test_query", oid::TEXT),
    ("assertion_type", oid::TEXT),
    ("assertion_comparator", oid::TEXT),
    ("assertion_value", oid::TEXT),
];

/// create creates a schema's tests table, which Dolt creates on the first write to it.
pub fn create(ctx: &mut Ctx<'_>, schema: &str) -> Result<TableDef> {
    crate::dolt::tables::create_backing(ctx, schema, TABLE, DEFINITION)
}

/// Test is a row of the tests table.
struct Test {
    name: String,
    group: Option<String>,
    query: String,
    assertion: String,
    comparator: String,
    value: Option<String>,
}

/// tests returns the tests that an argument selects: every test for `*`, otherwise the test it names or else the
/// tests of the group it names, failing as Dolt does when it selects none.
fn tests(ctx: &mut Ctx<'_>, argument: &str) -> Result<Vec<Test>> {
    let relation = pg_query::protobuf::RangeVar { relname: TABLE.to_string(), inh: true, ..Default::default() };
    let rows = match ctx.resolve_table(&relation) {
        Ok(table) => crate::query::scan(ctx.db, &table)?,
        Err(_) => Vec::new(),
    };
    let text = |value: &Value| match value {
        Value::Null => None,
        value => value.output(),
    };
    let all: Vec<Test> = rows
        .iter()
        .map(|row| Test {
            name: text(&row[0]).unwrap_or_default(),
            group: text(&row[1]),
            query: text(&row[2]).unwrap_or_default(),
            assertion: text(&row[3]).unwrap_or_default(),
            comparator: text(&row[4]).unwrap_or_default(),
            value: text(&row[5]),
        })
        .collect();
    let selected: Vec<Test> = match argument {
        "*" => all,
        _ => {
            let (named, rest): (Vec<Test>, Vec<Test>) = all.into_iter().partition(|t| t.name == argument);
            match named.is_empty() {
                true => rest.into_iter().filter(|t| t.group.as_deref() == Some(argument)).collect(),
                false => named,
            }
        }
    };
    if selected.is_empty() {
        return Err(error(format!("could not find tests for argument: {argument}")));
    }
    Ok(selected)
}

/// dolt_test_run runs the tests that its arguments select, every test without arguments, returning each test's name,
/// group, query, status, and the message that explains a failure.
pub fn dolt_test_run(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let mut arguments: Vec<String> = args.iter().map(|a| a.output().unwrap_or_default()).collect();
    if arguments.is_empty() {
        arguments.push("*".to_string());
    }
    let mut rows = Vec::new();
    for argument in &arguments {
        for test in tests(ctx, argument.trim_matches('\''))? {
            let message = run_test(ctx, &test);
            rows.push(Value::Record(vec![
                Value::Text(test.name),
                Value::Text(test.group.unwrap_or_default()),
                Value::Text(test.query),
                Value::Text(if message.is_empty() { "PASS" } else { "FAIL" }.to_string()),
                Value::Text(message),
            ]));
        }
    }
    Ok(Value::Set(rows))
}

/// run_test runs a test's query and checks its assertion, returning the message that explains a failure, which is
/// empty when the test passes.
fn run_test(ctx: &mut Ctx<'_>, test: &Test) -> String {
    let statement = match validate(&test.query) {
        Ok(statement) => statement,
        Err(message) => return message,
    };
    let outcome = match statement {
        Statement::Postgres { node: NodeEnum::SelectStmt(select), .. } => {
            let planned = ctx.nested(&mut Vec::new(), &[], None, |ctx| {
                crate::plan::Planner { ctx, outer: Vec::new() }.plan_query(&select)
            });
            match planned {
                Ok(query) => match ctx.nested(&mut Vec::new(), &[], None, |ctx| query.plan.run(ctx)) {
                    Ok(rows) => Ok(rows),
                    Err(err) => return format!("Query error: {}", err.message),
                },
                Err(err) => Err(err),
            }
        }
        Statement::Postgres { node, .. } => ctx.nested(&mut Vec::new(), &[], None, |ctx| ctx.run(&node)).map(rows),
        Statement::Describe { relation, extras } => ctx.describe_table(&relation, &extras).map(rows),
        Statement::ShowCreateTable { relation, extras } => ctx.show_create_table(&relation, &extras).map(rows),
        Statement::Listing { kind, from } => ctx.list_objects(&kind, &from).map(rows),
        _ => return "Cannot execute write queries".to_string(),
    };
    match outcome {
        Ok(rows) => assert_data(&test.assertion, &test.comparator, test.value.as_deref(), &rows),
        Err(err) => format!("query error: {}", err.message),
    }
}

/// rows returns the rows of a statement's outcome.
fn rows(outcome: Outcome) -> Vec<Vec<Value>> {
    match outcome {
        Outcome::Rows { rows, .. } => rows,
        _ => Vec::new(),
    }
}

/// validate parses a test's query, returning the reason it cannot run as a failure message: more than one statement,
/// a statement that writes, or a call of DOLT_TEST_RUN.
fn validate(query: &str) -> std::result::Result<Statement, String> {
    let mut statements = crate::parse::parse(query).map_err(|err| format!("query error: {}", err.message))?;
    if statements.len() != 1 {
        return Err("Can only run exactly one query".to_string());
    }
    let statement = statements.remove(0);
    if writes(&statement) {
        return Err("Cannot execute write queries".to_string());
    }
    if query.to_lowercase().contains("dolt_test_run(") {
        return Err("Cannot call dolt_test_run in dolt_tests".to_string());
    }
    Ok(statement)
}

/// writes reports whether a statement may change anything: anything but a query without INTO or data-modifying WITH
/// queries, EXPLAIN without ANALYZE, SHOW, and Doltgres' describing and listing statements.
fn writes(statement: &Statement) -> bool {
    let node = match statement {
        Statement::Postgres { node, .. } => node,
        Statement::Describe { .. } | Statement::ShowCreateTable { .. } | Statement::Listing { .. } => return false,
        _ => return true,
    };
    let reads = |node: &Option<Box<pg_query::Node>>| {
        node.as_deref().and_then(|n| n.node.as_ref()).is_some_and(|n| !writes_node(n))
    };
    match node {
        NodeEnum::ExplainStmt(explain) => {
            let analyze = explain.options.iter().any(
                |o| matches!(o.node.as_ref(), Some(NodeEnum::DefElem(d)) if d.defname.eq_ignore_ascii_case("analyze")),
            );
            analyze && !reads(&explain.query)
        }
        node => writes_node(node),
    }
}

/// writes_node reports whether a Postgres statement may change anything, which only a query without INTO or
/// data-modifying WITH queries and SHOW cannot.
fn writes_node(node: &NodeEnum) -> bool {
    match node {
        NodeEnum::SelectStmt(select) => {
            let modifying = select.with_clause.iter().flat_map(|w| &w.ctes).any(|cte| match cte.node.as_ref() {
                Some(NodeEnum::CommonTableExpr(cte)) => {
                    cte.ctequery.as_deref().and_then(|q| q.node.as_ref()).is_some_and(writes_node)
                }
                _ => false,
            });
            select.into_clause.is_some() || modifying
        }
        NodeEnum::VariableShowStmt(_) => false,
        _ => true,
    }
}

/// assert_data checks a query's rows against a test's assertion, returning the message that explains a failure, as
/// Dolt's AssertData does.
fn assert_data(assertion: &str, comparator: &str, value: Option<&str>, rows: &[Vec<Value>]) -> String {
    match assertion {
        "expected_rows" | "expected_columns" => {
            let Some(value) = value else { return "null is not a valid assertion for expected_rows".to_string() };
            let Ok(expected) = value.parse::<i64>() else {
                return format!("cannot run assertion on non integer value: {value}");
            };
            let actual = match assertion {
                "expected_rows" => rows.len(),
                _ => rows.first().map_or(0, Vec::len),
            };
            compare(comparator, expected, actual as i64, assertion)
        }
        "expected_single_value" => expect_single_value(comparator, value, rows),
        _ => format!("{assertion} is not a valid assertion type"),
    }
}

/// SINGLE is the name of the assertion of a single value.
const SINGLE: &str = "expected_single_value";

/// expect_single_value checks that a query returned exactly one value and compares it with the expected value, read
/// as the type of the value returned, as Dolt's expectSingleValue does.
fn expect_single_value(comparator: &str, value: Option<&str>, rows: &[Vec<Value>]) -> String {
    let Some(row) = rows.first() else {
        return "expected_single_value expects exactly one cell. Received 0 rows".to_string();
    };
    if row.len() != 1 {
        return "expected_single_value expects exactly one cell. Received multiple columns".to_string();
    }
    if rows.len() > 1 {
        return "expected_single_value expects exactly one cell. Received multiple rows".to_string();
    }
    let actual = &row[0];
    let Some(value) = value else {
        return match (comparator, actual) {
            ("==", Value::Null) => String::new(),
            ("==", actual) => format!("Assertion failed: {SINGLE} equal to NULL, got {}", go_value(actual)),
            ("!=", Value::Null) => format!("Assertion failed: {SINGLE} not equal to NULL, got NULL"),
            ("!=", _) => String::new(),
            _ => format!("{comparator} is not a valid comparison for NULL values"),
        };
    };
    if value != "0"
        && value != "1"
        && let Some(expected) = parse_bool(value)
    {
        let actual = match actual {
            Value::Bool(b) => *b,
            Value::Int2(i) => *i == 1,
            Value::Int4(i) => *i == 1,
            Value::Int8(i) => *i == 1,
            Value::Text(s) => s == "1",
            other => {
                return format!(
                    "Could not convert value to boolean: unexpected type {}, was expecting bool, int, or string",
                    type_name(other)
                );
            }
        };
        return match comparator {
            "==" if expected != actual => format!("Assertion failed: {SINGLE} equal to {expected}, got {actual}"),
            "!=" if expected == actual => format!("Assertion failed: {SINGLE} not equal to {expected}, got {actual}"),
            "==" | "!=" => String::new(),
            _ => format!("{comparator} is not a valid comparison for boolean values. Only '==' and '!=' are supported"),
        };
    }
    let integer = value.parse::<i64>().ok();
    match actual {
        Value::Int2(a) => match integer {
            Some(e) => compare(comparator, e as i16, *a, SINGLE),
            None => format!("Could not compare non integer value '{value}', with {a}"),
        },
        Value::Int4(a) => match integer {
            Some(e) => compare(comparator, e as i32, *a, SINGLE),
            None => format!("Could not compare non integer value '{value}', with {a}"),
        },
        Value::Int8(a) => match integer {
            Some(e) => compare(comparator, e, *a, SINGLE),
            None => format!("Could not compare non integer value '{value}', with {a}"),
        },
        Value::Float8(a) => match value.parse::<f64>() {
            Ok(e) => compare(comparator, e, *a, SINGLE),
            Err(_) => format!("Could not compare non float value '{value}', with {a:.6}"),
        },
        Value::Float4(a) => match value.parse::<f32>() {
            Ok(e) => compare(comparator, e, *a, SINGLE),
            Err(_) => format!("Could not compare non float value '{value}', with {a:.6}"),
        },
        Value::Numeric(a) => match Numeric::parse(value) {
            Ok(e) => compare_numeric(comparator, &e, a),
            Err(_) => format!("Could not compare non decimal value '{value}', with {a}"),
        },
        Value::Date(_) | Value::Timestamp(_) | Value::TimestampTz(_) => match (parse_date(value), timestamp(actual)) {
            (Some((expected, format)), Some(actual)) => compare_dates(comparator, expected, actual, format),
            _ => format!("{value} does not appear to be a valid date"),
        },
        Value::Text(a) => compare(comparator, value.to_string(), a.clone(), SINGLE),
        other => format!(
            "Type {} is not supported. Open an issue at https://github.com/dolthub/dolt/issues to see it added",
            type_name(other)
        ),
    }
}

/// parse_bool reads a boolean as Go's strconv.ParseBool does.
fn parse_bool(text: &str) -> Option<bool> {
    match text {
        "1" | "t" | "T" | "TRUE" | "true" | "True" => Some(true),
        "0" | "f" | "F" | "FALSE" | "false" | "False" => Some(false),
        _ => None,
    }
}

/// go_value writes a value as Go's `%v` writes the value Dolt reads, with booleans as `true` and `false`.
fn go_value(value: &Value) -> String {
    match value {
        Value::Bool(b) => b.to_string(),
        value => value.output().unwrap_or_else(|| "<nil>".to_string()),
    }
}

/// type_name names a value's kind in the messages about values a test cannot compare.
fn type_name(value: &Value) -> String {
    match value {
        Value::Null => "<nil>".to_string(),
        value => format!("{value:?}").split('(').next().unwrap_or_default().to_lowercase(),
    }
}

/// compare checks an actual value against the expected one with a comparator, returning the message that explains a
/// failure, as Dolt's compareTestAssertion does.
fn compare<T: PartialOrd + std::fmt::Display>(comparator: &str, expected: T, actual: T, assertion: &str) -> String {
    let (failed, relation) = match comparator {
        "==" => (actual != expected, "equal to"),
        "!=" => (actual == expected, "not equal to"),
        "<" => (actual >= expected, "less than"),
        "<=" => (actual > expected, "less than or equal to"),
        ">" => (actual <= expected, "greater than"),
        ">=" => (actual < expected, "greater than or equal to"),
        _ => return format!("{comparator} is not a valid comparison type"),
    };
    match failed {
        true => format!("Assertion failed: {assertion} {relation} {expected}, got {actual}"),
        false => String::new(),
    }
}

/// compare_numeric checks an actual numeric value against the expected one, as Dolt's compareDecimals does.
fn compare_numeric(comparator: &str, expected: &Numeric, actual: &Numeric) -> String {
    let order = actual.cmp_numeric(expected);
    let (failed, relation) = match comparator {
        "==" => (order.is_ne(), "equal to"),
        "!=" => (order.is_eq(), "not equal to"),
        "<" => (order.is_ge(), "less than"),
        "<=" => (order.is_gt(), "less than or equal to"),
        ">" => (order.is_le(), "greater than"),
        ">=" => (order.is_lt(), "greater than or equal to"),
        _ => return format!("{comparator} is not a valid comparison type"),
    };
    match failed {
        true => format!("Assertion failed: {SINGLE} {relation} {expected}, got {actual}"),
        false => String::new(),
    }
}

/// DATE_FORMATS are the formats a test's expected date may take, in the order Dolt tries them: Go's DateOnly,
/// DateTime, TimeOnly, RFC3339, and RFC1123Z layouts.
const DATE_FORMATS: &[&str] = &["%Y-%m-%d", "%Y-%m-%d %H:%M:%S", "%H:%M:%S", "%+", "%a, %d %b %Y %H:%M:%S %z"];

/// parse_date reads a test's expected date in the first format that fits it, as Dolt's parseTestsDate does, with the
/// format it fit.
fn parse_date(text: &str) -> Option<(NaiveDateTime, &'static str)> {
    for &format in DATE_FORMATS {
        let parsed = match format {
            "%Y-%m-%d" => NaiveDate::parse_from_str(text, format).ok().and_then(|d| d.and_hms_opt(0, 0, 0)),
            "%H:%M:%S" => chrono::NaiveTime::parse_from_str(text, format)
                .ok()
                .and_then(|t| NaiveDate::from_ymd_opt(0, 1, 1).map(|d| d.and_time(t))),
            "%Y-%m-%d %H:%M:%S" => NaiveDateTime::parse_from_str(text, format).ok(),
            _ => DateTime::parse_from_str(text, format).ok().map(|d| d.naive_utc()),
        };
        if let Some(parsed) = parsed {
            return Some((parsed, format));
        }
    }
    None
}

/// timestamp returns a date or timestamp as a time in UTC.
fn timestamp(value: &Value) -> Option<NaiveDateTime> {
    let epoch = NaiveDate::from_ymd_opt(2000, 1, 1)?.and_hms_opt(0, 0, 0)?;
    match value {
        Value::Date(days) => epoch.checked_add_signed(chrono::Duration::days(*days as i64)),
        Value::Timestamp(micros) | Value::TimestampTz(micros) => {
            epoch.checked_add_signed(chrono::Duration::microseconds(*micros))
        }
        _ => None,
    }
}

/// compare_dates checks an actual time against the expected one, writing both in the format the expected time was
/// written in, as Dolt's compareDates does.
fn compare_dates(comparator: &str, expected: NaiveDateTime, actual: NaiveDateTime, format: &str) -> String {
    let write = |t: NaiveDateTime| match format {
        "%+" => t.and_utc().to_rfc3339_opts(chrono::SecondsFormat::AutoSi, true),
        "%a, %d %b %Y %H:%M:%S %z" => t.and_utc().format("%a, %d %b %Y %H:%M:%S +0000").to_string(),
        _ => t.format(format).to_string(),
    };
    let (failed, relation) = match comparator {
        "==" => (actual != expected, "equal to"),
        "!=" => (actual == expected, "not equal to"),
        "<" => (actual >= expected, "less than"),
        "<=" => (actual > expected, "less than or equal to"),
        ">" => (actual <= expected, "greater than"),
        ">=" => (actual < expected, "greater than or equal to"),
        _ => return format!("{comparator} is not a valid comparison type"),
    };
    match failed {
        true => format!("Assertion failed: {SINGLE} {relation} {}, got {}", write(expected), write(actual)),
        false => String::new(),
    }
}
