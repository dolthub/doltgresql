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

#![forbid(unsafe_code)]

//! Temporary tooling that ports the Go test suite to Rust. It is removed along with the Go code.

mod dump;

use std::collections::VecDeque;
use std::io::Write;
use std::sync::{Arc, Mutex, mpsc};

use base64::Engine;
use base64::engine::general_purpose::STANDARD;
use harness::script::{Capture, Observation, capture_script};
use harness::server::Target;
use pgproto_fields::fields_json;
use serde_json::{Value, json};

/// main dispatches the subcommand.
fn main() {
    let args: Vec<String> = std::env::args().collect();
    match args.get(1).map(String::as_str) {
        Some("capture") if args.len() >= 5 => {
            let jobs = args.get(5).and_then(|j| j.parse().ok()).unwrap_or(4);
            let filter = args.get(6).cloned();
            capture(&args[2], &args[3], &args[4], jobs, filter);
        }
        _ => {
            eprintln!("usage: goport capture <dump.jsonl> <target> <out.jsonl> [jobs] [test name filter]");
            std::process::exit(2);
        }
    }
}

/// Job is one script to capture.
struct Job {
    record: usize,
    test: usize,
    go_test: String,
    runner: String,
    repetitions: usize,
    script: Result<harness::script::ScriptTest, String>,
}

/// capture runs every script of the dump against the target and writes what each returned.
fn capture(dump_path: &str, target: &str, out_path: &str, jobs: usize, filter: Option<String>) {
    let target = Target::parse(target).unwrap_or_else(|err| panic!("{err}"));
    let records = dump::read_dump(dump_path);
    let mut queue = VecDeque::new();
    for record in &records {
        let transaction = record.runner == "RunTransactionTests";
        let repetitions = record.extra.get("n").and_then(Value::as_u64).unwrap_or(1) as usize;
        match record.runner.as_str() {
            "RunScripts" | "RunTransactionTests" | "RunScriptN" => {}
            _ => continue,
        }
        if let Some(filter) = &filter
            && !record.test.contains(filter.as_str())
        {
            continue;
        }
        for (test, value) in record.tests.iter().enumerate() {
            queue.push_back(Job {
                record: record.index,
                test,
                go_test: record.test.clone(),
                runner: record.runner.clone(),
                repetitions,
                script: dump::to_script(value, transaction),
            });
        }
    }
    let total = queue.len();
    eprintln!("capturing {total} scripts with {jobs} workers");
    let queue = Arc::new(Mutex::new(queue));
    let (sender, receiver) = mpsc::channel::<Value>();
    let mut workers = Vec::new();
    for _ in 0..jobs {
        let queue = queue.clone();
        let sender = sender.clone();
        let target = target.clone();
        workers.push(std::thread::spawn(move || {
            loop {
                let Some(job) = queue.lock().unwrap().pop_front() else { break };
                let capture = match &job.script {
                    Ok(script) => capture_script(&target, script, job.repetitions),
                    Err(err) => Capture { setup_error: Some(format!("conversion: {err}")), ..Capture::default() },
                };
                let name = job.script.as_ref().map(|s| s.name).unwrap_or_default();
                let _ = sender.send(json!({
                    "record": job.record,
                    "test": job.test,
                    "go_test": job.go_test,
                    "runner": job.runner,
                    "name": name,
                    "setup_error": capture.setup_error,
                    "setup_failed_at": capture.setup_failed_at,
                    "finish_errors": capture.finish_errors,
                    "observations": capture.observations.iter().map(observation_json).collect::<Vec<_>>(),
                }));
            }
        }));
    }
    drop(sender);
    let mut out = std::io::BufWriter::new(std::fs::File::create(out_path).unwrap());
    for (done, value) in receiver.iter().enumerate() {
        writeln!(out, "{value}").unwrap();
        if (done + 1) % 50 == 0 {
            eprintln!("{} / {total}", done + 1);
            out.flush().unwrap();
        }
    }
    for worker in workers {
        worker.join().unwrap();
    }
    eprintln!("done");
}

/// observation_json converts an observation into JSON.
fn observation_json(observation: &Observation) -> Value {
    json!({
        "columns": observation.columns,
        "rows": observation.rows,
        "queried": observation.queried,
        "tag": observation.tag,
        "error": observation.error.as_ref().map(fields_json),
        "client_error": observation.client_error,
        "notices": observation.notices.iter().map(fields_json).collect::<Vec<_>>(),
        "copy_out": observation.copy_out.as_ref().map(|data| STANDARD.encode(data)),
        "skipped": observation.skipped,
    })
}

/// pgproto_fields converts error fields into JSON.
mod pgproto_fields {
    use serde_json::{Value, json};

    /// fields_json converts the fields of an error or notice into JSON, omitting absent fields.
    pub fn fields_json(fields: &pgproto::ErrorFields) -> Value {
        let mut value = json!({
            "severity": fields.severity,
            "code": fields.code,
            "message": fields.message,
        });
        let object = value.as_object_mut().unwrap();
        let mut put = |name: &str, text: &str| {
            if !text.is_empty() {
                object.insert(name.to_string(), json!(text));
            }
        };
        put("severity_unlocalized", &fields.severity_unlocalized);
        put("detail", &fields.detail);
        put("hint", &fields.hint);
        put("internal_query", &fields.internal_query);
        put("where", &fields.where_);
        put("schema", &fields.schema_name);
        put("table", &fields.table_name);
        put("column", &fields.column_name);
        put("data_type", &fields.data_type_name);
        put("constraint", &fields.constraint_name);
        put("file", &fields.file);
        put("routine", &fields.routine);
        if fields.position != 0 {
            object.insert("position".into(), json!(fields.position));
        }
        if fields.internal_position != 0 {
            object.insert("internal_position".into(), json!(fields.internal_position));
        }
        if fields.line != 0 {
            object.insert("line".into(), json!(fields.line));
        }
        value
    }
}
