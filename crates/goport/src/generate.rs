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

//! Generates Rust script tests from the dumped Go tests and their captures. Expectations come from Postgres,
//! except where Postgres cannot run a statement because it depends on Dolt, where they come from the Go server.

use std::collections::{BTreeMap, HashMap};
use std::fmt::Write as _;

use harness::script::{Expected, Flow, ScriptTestAssertion, effective_flow};
use serde_json::Value;

use crate::dump::{Record, go_flow, to_assertion};
use crate::rust;

/// Source is where an assertion's expectation came from.
#[derive(Clone, Copy, Debug, PartialEq, Eq, PartialOrd, Ord)]
pub enum Source {
    /// Postgres ran the statement.
    Postgres,
    /// The statement depends on Dolt, so the Go server's output is used.
    DoltStatement,
    /// The script's setup depends on Dolt, so the Go server's output is used for the whole script.
    DoltScript,
    /// An earlier Dolt statement changed state that Postgres does not have, and the outputs differ.
    AfterDivergence,
    /// Neither capture can produce an expectation.
    Unavailable,
}

/// Captures holds the capture of every script, by record and test index.
pub type Captures = HashMap<(usize, usize), Value>;

/// read_captures reads a capture file.
pub fn read_captures(path: &str) -> Captures {
    let text = std::fs::read_to_string(path).unwrap_or_else(|err| panic!("cannot read {path}: {err}"));
    text.lines()
        .map(|line| {
            let value: Value = serde_json::from_str(line).unwrap();
            ((value["record"].as_u64().unwrap() as usize, value["test"].as_u64().unwrap() as usize), value)
        })
        .collect()
}

/// mentions_dolt reports whether a statement uses a Dolt feature that Postgres lacks.
pub fn mentions_dolt(query: &str) -> bool {
    let lower = query.to_lowercase();
    lower.contains("dolt") || lower.contains(" as of ") || lower.trim_start().starts_with("use ")
}

/// GeneratedAssertion is the Rust source of one assertion and where its expectation came from.
pub struct GeneratedAssertion {
    /// The source code.
    pub code: String,
    /// The expectation's source.
    pub source: Source,
    /// A note for the review report.
    pub note: Option<String>,
}

/// observation_equal reports whether two captured observations match, ignoring Postgres-internal error fields.
fn observation_equal(a: &Value, b: &Value) -> bool {
    let strip = |v: &Value| {
        let mut v = v.clone();
        for key in ["file", "line", "routine"] {
            if let Some(error) = v.get_mut("error").and_then(Value::as_object_mut) {
                error.remove(key);
            }
        }
        v
    };
    strip(a) == strip(b)
}

/// expectation renders an observation as an expected outcome, returning the code and the outcome's kind.
fn expectation(observation: &Value) -> (String, Expected) {
    if let Some(error) = observation.get("error").filter(|e| !e.is_null()) {
        return (format!("Expected::Error({})", rust::diagnostic(error)), Expected::Error(harness::script::E));
    }
    let tag = observation["tag"].as_str().unwrap_or_default();
    let columns = observation["columns"].as_array().cloned().unwrap_or_default();
    let rows = observation["rows"].as_array().cloned().unwrap_or_default();
    if observation["queried"].as_bool() == Some(true) && !(columns.is_empty() && rows.is_empty()) {
        let mut code = String::from("Expected::Rows {\n");
        code.push_str("                        columns: &[");
        code.push_str(
            &columns
                .iter()
                .map(|c| {
                    format!(
                        "Column({}, {})",
                        rust::string(c[0].as_str().unwrap()),
                        rust::type_oid(c[1].as_u64().unwrap() as u32)
                    )
                })
                .collect::<Vec<_>>()
                .join(", "),
        );
        code.push_str("],\n                        rows: &[");
        if rows.is_empty() {
            code.push_str("],\n");
        } else {
            code.push('\n');
            for row in &rows {
                let cells: Vec<String> = row.as_array().unwrap().iter().map(rust::cell).collect();
                let _ = writeln!(code, "                            &[{}],", cells.join(", "));
            }
            code.push_str("                        ],\n");
        }
        let _ = write!(code, "                        tag: {},\n                    }}", rust::string(tag));
        return (code, Expected::Rows { columns: &[], rows: &[], tag: "" });
    }
    (format!("Expected::Tag({})", rust::string(tag)), Expected::Tag(""))
}

/// generate_assertion renders one assertion from its Go definition and chosen observation.
fn generate_assertion(
    go: &Value,
    transaction: bool,
    observation: Option<&Value>,
    source: Source,
) -> GeneratedAssertion {
    let converted: ScriptTestAssertion = match to_assertion(go, transaction) {
        Ok(assertion) => assertion,
        Err(err) => {
            return GeneratedAssertion {
                code: format!("                // cannot convert: {err}\n"),
                source: Source::Unavailable,
                note: Some(err),
            };
        }
    };
    let mut fields = vec![format!("query: {}", rust::string(converted.query))];
    if !converted.bind_vars.is_empty() {
        fields.push(format!(
            "bind_vars: &[{}]",
            converted.bind_vars.iter().map(rust::bind_var).collect::<Vec<_>>().join(", ")
        ));
    }
    let mut note = None;
    let mut source = source;
    let blocking_or_close = converted.expected_blocking || converted.close_client;
    let mut kind = Expected::Ok;
    match observation {
        _ if blocking_or_close => {}
        Some(observation) if observation["client_error"].is_null() => {
            let (code, expected_kind) = expectation(observation);
            kind = expected_kind;
            fields.push(format!("expected: {code}"));
            let notices = observation["notices"].as_array().cloned().unwrap_or_default();
            if !notices.is_empty() {
                fields.push(format!(
                    "notices: &[{}]",
                    notices.iter().map(rust::diagnostic).collect::<Vec<_>>().join(", ")
                ));
            }
        }
        Some(observation)
            if observation["client_error"].as_str().is_some_and(|err| err.starts_with("failed to encode")) =>
        {
            let err = observation["client_error"].as_str().unwrap();
            fields.push(format!("expected: Expected::ClientError({})", rust::string(err)));
            kind = Expected::ClientError("");
        }
        Some(observation) => {
            note = Some(format!("client error: {}", observation["client_error"]));
            source = Source::Unavailable;
        }
        None => {
            note = Some("no observation".to_string());
            source = Source::Unavailable;
        }
    }
    let wanted = go_flow(go);
    let probe = ScriptTestAssertion { expected: kind, flow: Flow::Auto, ..converted };
    if effective_flow(&probe) != wanted && !blocking_or_close {
        fields.push(format!("flow: {}", rust::flow(wanted)));
    }
    if blocking_or_close {
        fields.push("flow: Flow::Exec".to_string());
    }
    for (name, value) in [
        ("username", converted.username),
        ("password", converted.password),
        ("client", converted.client),
        ("copy_from_stdin_file", converted.copy_from_stdin_file),
        ("copy_to_stdout_file", converted.copy_to_stdout_file),
        ("copy_round_trip_stdin_query", converted.copy_round_trip_stdin_query),
    ] {
        if !value.is_empty() {
            fields.push(format!("{name}: {}", rust::string(value)));
        }
    }
    if converted.expected_blocking {
        fields.push("expected_blocking: true".to_string());
    }
    if converted.close_client {
        fields.push("close_client: true".to_string());
    }
    if source == Source::Unavailable {
        fields.push(format!("skip: Some({})", rust::string(note.as_deref().unwrap_or("no expectation"))));
    }
    let mut code = String::from("                ScriptTestAssertion {\n");
    for field in fields {
        let _ = writeln!(code, "                    {field},");
    }
    code.push_str("                    ..A\n                },\n");
    GeneratedAssertion { code, source, note }
}

/// Report collects statistics and review notes.
#[derive(Default)]
pub struct Report {
    /// The number of assertions by expectation source.
    pub sources: BTreeMap<String, usize>,
    /// Lines for the review file.
    pub notes: Vec<String>,
}

/// generate_script renders one script.
fn generate_script(
    record: &Record,
    test_index: usize,
    test: &Value,
    pg: Option<&Value>,
    go: Option<&Value>,
    report: &mut Report,
) -> String {
    let transaction = record.runner == "RunTransactionTests";
    let name = test["Name"].as_str().unwrap_or_default();
    let assertions = test["Assertions"].as_array().cloned().unwrap_or_default();
    let pg_setup_failed = pg.is_none_or(|pg| !pg["setup_error"].is_null());
    let go_setup_failed = go.is_none_or(|go| !go["setup_error"].is_null());
    let observations = |capture: Option<&Value>| -> Vec<Value> {
        capture.and_then(|c| c["observations"].as_array().cloned()).unwrap_or_default()
    };
    let pg_observations = observations(pg);
    let go_observations = observations(go);
    let mut diverged = false;
    let mut body = String::new();
    for (index, assertion) in assertions.iter().enumerate() {
        let query = assertion["Query"].as_str().unwrap_or_default();
        let pg_observation = pg_observations.get(index);
        let go_observation = go_observations.get(index);
        let (observation, source) = if !pg_setup_failed {
            let pg_observation = pg_observation.unwrap();
            let pg_failed = !pg_observation["error"].is_null();
            if pg_failed && mentions_dolt(query) && !go_setup_failed {
                if go_observation.is_some_and(|o| o["error"].is_null()) {
                    diverged = true;
                }
                (go_observation, Source::DoltStatement)
            } else if diverged
                && !go_setup_failed
                && go_observation.is_some_and(|go| !observation_equal(go, pg_observation))
            {
                (go_observation, Source::AfterDivergence)
            } else {
                (Some(pg_observation), Source::Postgres)
            }
        } else if !go_setup_failed {
            (go_observation, Source::DoltScript)
        } else {
            (None, Source::Unavailable)
        };
        let generated = generate_assertion(assertion, transaction, observation, source);
        *report.sources.entry(format!("{:?}", generated.source)).or_default() += 1;
        if generated.source != Source::Postgres || generated.note.is_some() {
            report.notes.push(format!(
                "{} / {} / {}: {:?} {}",
                record.test,
                name,
                query.lines().next().unwrap_or_default(),
                generated.source,
                generated.note.unwrap_or_default()
            ));
        }
        body.push_str(&generated.code);
    }
    let mut code = String::from("        ScriptTest {\n");
    let _ = writeln!(code, "            name: {},", rust::string(name));
    if let Some(database) = test["Database"].as_str().filter(|d| !d.is_empty()) {
        let _ = writeln!(code, "            database: {},", rust::string(database));
    }
    let set_up: Vec<&str> =
        test["SetUpScript"].as_array().map(|s| s.iter().filter_map(Value::as_str).collect()).unwrap_or_default();
    if !set_up.is_empty() {
        code.push_str("            set_up_script: &[\n");
        for statement in set_up {
            let _ = writeln!(code, "                {},", rust::string(statement));
        }
        code.push_str("            ],\n");
    }
    if pg_setup_failed && go_setup_failed {
        let reason = format!(
            "setup fails on Postgres ({}) and on the Go server ({})",
            pg.map(|c| c["setup_error"].to_string()).unwrap_or_default(),
            go.map(|c| c["setup_error"].to_string()).unwrap_or_default()
        );
        let _ = writeln!(code, "            skip: Some({}),", rust::string(&reason));
        report.notes.push(format!("{} / {}: script skipped: {reason}", record.test, name));
    }
    code.push_str("            assertions: &[\n");
    code.push_str(&body);
    code.push_str("            ],\n            ..S\n        },\n");
    let _ = test_index;
    code
}

/// generate_wire_test renders one wire conversation, with received messages from the Postgres capture, or from the
/// Go capture when Postgres could not run it.
fn generate_wire_test(
    record: &Record,
    test: &Value,
    pg: Option<&Value>,
    go: Option<&Value>,
    report: &mut Report,
) -> String {
    let name = test["Name"].as_str().unwrap_or_default();
    let converted = match record.runner.as_str() {
        "RunWireScripts" => crate::wire::from_wire_script(test),
        _ => crate::wire::from_message_flow(test),
    };
    let converted = match converted {
        Ok(converted) => converted,
        Err(err) => {
            report.notes.push(format!("{} / {name}: cannot convert: {err}", record.test));
            *report.sources.entry("WireUnavailable".into()).or_default() += 1;
            return format!("        // {name}: cannot convert: {err}\n");
        }
    };
    let pg_ok = pg.is_some_and(|c| c["wire"]["error"].is_null());
    let go_ok = go.is_some_and(|c| c["wire"]["error"].is_null());
    let (capture, source) = match (pg_ok, go_ok) {
        (true, _) => (pg, "WirePostgres"),
        (false, true) => (go, "WireDolt"),
        _ => (None, "WireUnavailable"),
    };
    *report.sources.entry(source.into()).or_default() += 1;
    if source != "WirePostgres" {
        report.notes.push(format!(
            "{} / {name}: {source}: Postgres {}, Go {}",
            record.test,
            pg.map(|c| c["wire"]["error"].to_string()).unwrap_or_default(),
            go.map(|c| c["wire"]["error"].to_string()).unwrap_or_default()
        ));
    }
    let received = |step: usize| -> Vec<String> {
        capture
            .and_then(|c| c["wire"]["received"].as_array())
            .and_then(|r| r.iter().find(|x| x["step"].as_u64() == Some(step as u64)))
            .map(|x| x["messages"].as_array().unwrap().iter().map(|m| m.as_str().unwrap().to_string()).collect())
            .unwrap_or_default()
    };
    let other_rows = |step: usize| -> Vec<Value> {
        capture
            .and_then(|c| c["wire"]["other_rows"].as_array())
            .and_then(|r| r.iter().find(|x| x["step"].as_u64() == Some(step as u64)))
            .and_then(|x| x["rows"].as_array().cloned())
            .unwrap_or_default()
    };
    let mut code = String::from("        WireTest {\n");
    let _ = writeln!(code, "            name: {},", rust::string(name));
    if !converted.set_up_script.is_empty() {
        code.push_str("            set_up_script: &[\n");
        for statement in converted.set_up_script {
            let _ = writeln!(code, "                {},", rust::string(statement));
        }
        code.push_str("            ],\n");
    }
    if source == "WireUnavailable" {
        let _ = writeln!(
            code,
            "            skip: Some(\"neither Postgres nor the Go server completes this conversation\"),"
        );
    }
    code.push_str("            steps: &[\n");
    for (index, step) in converted.steps.iter().enumerate() {
        match step {
            harness::wire::Step::Send(messages) => {
                code.push_str("                Step::Send(&[\n");
                for message in *messages {
                    let _ = writeln!(code, "                    {},", crate::wire::send_code(message));
                }
                code.push_str("                ]),\n");
            }
            harness::wire::Step::Receive(_) => {
                code.push_str("                Step::Receive(&[\n");
                for message in received(index) {
                    let _ = writeln!(code, "                    {message},");
                }
                code.push_str("                ]),\n");
            }
            harness::wire::Step::OtherQuery { query, .. } => {
                let rows = other_rows(index);
                let _ = write!(code, "                Step::OtherQuery {{ query: {}, rows: &[", rust::string(query));
                for row in rows {
                    let cells: Vec<String> = row.as_array().unwrap().iter().map(rust::cell).collect();
                    let _ = write!(code, "&[{}], ", cells.join(", "));
                }
                code.push_str("] },\n");
            }
        }
    }
    code.push_str("            ],\n            ..W\n        },\n");
    code
}

/// go_file returns the Go test file that called the runner.
pub fn go_file(record: &Record) -> String {
    record
        .callers
        .iter()
        .map(|caller| caller.rsplit_once(':').map(|(file, _)| file).unwrap_or(caller))
        .find(|file| file.ends_with("_test.go"))
        .map(|file| file.rsplit('/').next().unwrap().trim_end_matches("_test.go").to_string())
        .unwrap_or_else(|| "unknown".to_string())
}

/// generate writes a Rust module for every Go test file, returning the report.
pub fn generate(records: &[Record], pg: &Captures, go: &Captures, out_dir: &str) -> Report {
    let mut report = Report::default();
    let mut files: BTreeMap<String, BTreeMap<String, Vec<String>>> = BTreeMap::new();
    for record in records {
        if matches!(record.runner.as_str(), "RunWireScripts" | "RunMessageFlowTests") {
            let mut call = String::from("run_wire_tests(&[\n");
            for (index, test) in record.tests.iter().enumerate() {
                call.push_str(&generate_wire_test(
                    record,
                    test,
                    pg.get(&(record.index, index)),
                    go.get(&(record.index, index)),
                    &mut report,
                ));
            }
            call.push_str("    ]);\n");
            let test_name = record.test.split('/').next().unwrap().to_string();
            files.entry(go_file(record)).or_default().entry(test_name).or_default().push(call);
            continue;
        }
        let runner = match record.runner.as_str() {
            "RunScripts" | "RunTransactionTests" => "run_scripts(&[\n".to_string(),
            "RunScriptN" => String::new(),
            _ => continue,
        };
        let repetitions = record.extra.get("n").and_then(Value::as_u64).unwrap_or(1);
        let mut call = if runner.is_empty() { "run_scripts_repeated(&[\n".to_string() } else { runner };
        for (index, test) in record.tests.iter().enumerate() {
            call.push_str(&generate_script(
                record,
                index,
                test,
                pg.get(&(record.index, index)),
                go.get(&(record.index, index)),
                &mut report,
            ));
        }
        if record.runner == "RunScriptN" {
            let _ = writeln!(call, "    ], {repetitions});");
        } else {
            call.push_str("    ]);\n");
        }
        let test_name = record.test.split('/').next().unwrap().to_string();
        files.entry(go_file(record)).or_default().entry(test_name).or_default().push(call);
    }
    std::fs::create_dir_all(out_dir).unwrap();
    let mut modules = Vec::new();
    for (file, tests) in &files {
        let mut code = String::from(rust::LICENSE_HEADER);
        code.push_str(
            "\nuse harness::oid::*;\nuse harness::pgx::Time;\nuse harness::script::Cell::{Null, Text as T};\n",
        );
        code.push_str(
            "use harness::script::{A, BindVar, Column, Diagnostic, E, Expected, Flow, N, S, ScriptTest, \
             ScriptTestAssertion, USER_DEFINED, run_scripts, run_scripts_repeated};\n",
        );
        code.push_str(
            "use harness::wire::{Datum, F, Field, Fields, Receive, Send, Step, W, WireTest, run_wire_tests};\n",
        );
        for (test, calls) in tests {
            let _ = write!(code, "\n#[test]\nfn {}() {{\n", rust::snake_case(test));
            for call in calls {
                code.push_str("    ");
                code.push_str(call);
            }
            code.push_str("}\n");
        }
        std::fs::write(format!("{out_dir}/{file}.rs"), code).unwrap();
        modules.push(file.clone());
    }
    let mut main = String::from(rust::LICENSE_HEADER);
    main.push_str(
        "\n//! Script tests ported from the Go test suite, whose expectations are what Postgres returns.\n\n",
    );
    main.push_str("#![allow(unused_imports)]\n\n");
    for module in &modules {
        let _ = writeln!(main, "mod {module};");
    }
    std::fs::write(format!("{out_dir}/main.rs"), main).unwrap();
    report
}
