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

//! The corpus of oracle scenarios: making it from Dolt's recorded sessions, and replaying it against Doltgres.
//!
//! A scenario is the work of one oracle test case: its repositories, which become databases, and the steps run in
//! them. A corpus file holds a suite's scenarios as JSON lines. In a step's SQL, `@dbN@` stands for the scenario's
//! Nth database and `@root@` for the directory that holds its file remotes.

use std::collections::{BTreeMap, HashMap};
use std::path::{Path, PathBuf};
use std::time::Duration;

use harness::server::{Server, Target};
use harness::wire::RawConnection;
use pgproto::{BackendMessage, FrontendMessage};
use serde_json::{Value as Json, json};

use crate::mysql::{self, Translated};

/// QUERY_TIMEOUT is how long a statement may run before the scenario fails.
const QUERY_TIMEOUT: Duration = Duration::from_secs(300);

/// BOOL is the OID of the boolean type.
const BOOL: u32 = 16;

/// Outcome is what a statement did on Doltgres: failed, returned rows, or finished without rows.
#[derive(Clone, Debug, PartialEq)]
enum Outcome {
    Error(String),
    Rows { columns: Vec<String>, rows: Vec<Vec<Option<String>>> },
    Done,
}

/// Rows are result rows of text and nulls.
type Rows = Vec<Vec<Option<String>>>;

/// Expected is what a statement did on Dolt: its error, or its result sets' columns and rows.
#[derive(Clone, Debug)]
struct Expected {
    error: Option<String>,
    results: Vec<(Vec<String>, Rows)>,
}

/// Step is one thing a scenario does.
#[derive(Clone, Debug)]
enum Step {
    /// A statement run on the server's own database, such as one creating a scenario's database.
    Admin { sql: String, error: bool },
    /// A Dolt command run as a procedure in a database, whose failure matters but whose output does not.
    Command { database: String, sql: String, error: bool },
    /// A session of statements in a database.
    Session { database: String, keep_going: bool, statements: Vec<(String, Expected)> },
}

/// Scenario is a suite's test case: its name, the directory its repositories were in, and its steps.
struct Scenario {
    suite: String,
    name: String,
    root: String,
    databases: Vec<String>,
    steps: Vec<Step>,
}

/// generate makes the corpus from Dolt's recorded sessions, running each scenario on Doltgres to settle its
/// expectations.
pub fn generate(record: &str, binary: &str, corpus: &str) -> Result<bool, String> {
    let rewrites = Rewrites::load(&Path::new(corpus).join(REWRITES))?;
    let text = std::fs::read_to_string(record).map_err(|err| format!("{record}: {err}"))?;
    let records: Vec<Json> = text.lines().filter_map(|l| serde_json::from_str(l).ok()).collect();
    let scenarios = scenarios(&records);
    std::fs::create_dir_all(corpus).map_err(|err| err.to_string())?;
    let mut suites: BTreeMap<String, Vec<&Scenario>> = BTreeMap::new();
    for scenario in &scenarios {
        suites.entry(scenario.suite.clone()).or_default().push(scenario);
    }
    for (suite, scenarios) in suites {
        let server = Server::start(&Target::Doltgres { binary: PathBuf::from(binary) }, "")?;
        let mut lines = Vec::new();
        let (mut matched, mut diverged) = (0, 0);
        for (index, scenario) in scenarios.iter().enumerate() {
            let entry = generate_scenario(&server, index, scenario, &rewrites)?;
            for step in entry["steps"].as_array().into_iter().flatten() {
                for statement in step["statements"].as_array().into_iter().flatten() {
                    if statement.get("divergence").is_some() {
                        diverged += 1;
                    } else {
                        matched += 1;
                    }
                }
            }
            lines.push(entry.to_string());
        }
        let path = Path::new(corpus).join(format!("{suite}.jsonl"));
        std::fs::write(&path, lines.join("\n") + "\n").map_err(|err| err.to_string())?;
        println!("{suite}: {} scenarios, {matched} statements match Dolt, {diverged} diverge", scenarios.len());
    }
    Ok(true)
}

/// generate_scenario runs a scenario on Doltgres and returns its corpus entry.
fn generate_scenario(server: &Server, index: usize, scenario: &Scenario, rewrites: &Rewrites) -> Result<Json, String> {
    let mut runner = Runner::new(server, index)?;
    let (mut ours, mut dolt) = (Normalizer::new(&runner.root), Normalizer::new(&scenario.root));
    let mut steps = Vec::new();
    for step in &scenario.steps {
        match step {
            Step::Admin { sql, error } => {
                let failed = matches!(runner.admin(sql)?, Outcome::Error(_));
                steps.push(json!({"admin": sql, "error": error, "divergence": (failed != *error).then_some("")}));
            }
            Step::Command { database, sql, error } => {
                let failed = matches!(runner.command(database, sql)?, Outcome::Error(_));
                steps.push(json!({"command": sql, "database": database, "error": error,
                    "divergence": (failed != *error).then_some("")}));
            }
            Step::Session { database, keep_going, statements } => {
                let database = dolt.placeholders(database);
                let mut session = runner.session(&ours.restore(&database));
                let mut entries = Vec::new();
                for (dolt_sql, expected) in statements {
                    let translated = map_translated(&session.translate(dolt_sql), |s| dolt.placeholders(s));
                    let outcome = session.run(&map_translated(&translated, |s| ours.restore(s)))?;
                    let failed = matches!(outcome, Outcome::Error(_));
                    let ours_normal = ours.outcome(&outcome);
                    let dolt_normal = dolt.expected(expected);
                    let sql = translated_json(&translated);
                    let mut entry = if sql.as_str() == Some(dolt_sql.as_str()) {
                        json!({"sql": sql})
                    } else {
                        json!({"dolt": dolt_sql, "sql": sql})
                    };
                    if equivalent(&dolt_normal, &ours_normal) {
                        entry["expect"] = outcome_json(&ours_normal);
                    } else if let Some(reason) = rewrites.reason(&scenario.suite, &scenario.name, dolt_sql) {
                        entry["expect"] = outcome_json(&ours_normal);
                        entry["postgres"] = json!(reason);
                    } else {
                        entry["expect"] = expected_json(&dolt_normal);
                        entry["divergence"] = json!(describe(&ours_normal));
                    }
                    entries.push(entry);
                    if failed && !keep_going {
                        break;
                    }
                }
                steps.push(json!({"database": database, "continue": keep_going, "statements": entries}));
            }
        }
    }
    Ok(json!({"name": scenario.name, "databases": scenario.databases.len(), "steps": steps}))
}

/// REWRITES is the corpus file of the statements whose expectations are Doltgres' Postgres behavior rather than Dolt's.
const REWRITES: &str = "postgres.json";

/// Rewrites are the statements where Postgres rightly differs from Dolt, each with the reason: a statement of a suite's
/// scenario, of any scenario of a suite, or of any suite, named by its whole text or by text it contains.
struct Rewrites {
    rules: Vec<Rule>,
}

/// Rule is a rewrite: the suite and scenario it is limited to, the statement's text or a part of it, and the reason.
struct Rule {
    suite: Option<String>,
    scenario: Option<String>,
    dolt: Option<String>,
    contains: Option<String>,
    reason: String,
}

impl Rewrites {
    /// load reads the rewrites, a JSON array of objects with `reason`, `dolt` (the statement Dolt ran) or `contains`
    /// (text it contains), and optionally `suite` and `scenario`; a missing file holds none.
    fn load(path: &Path) -> Result<Rewrites, String> {
        let Ok(text) = std::fs::read_to_string(path) else { return Ok(Rewrites { rules: Vec::new() }) };
        let rules: Json = serde_json::from_str(&text).map_err(|err| format!("{}: {err}", path.display()))?;
        let field = |rule: &Json, name: &str| rule[name].as_str().map(str::to_string);
        Ok(Rewrites {
            rules: rules
                .as_array()
                .into_iter()
                .flatten()
                .map(|r| Rule {
                    suite: field(r, "suite"),
                    scenario: field(r, "scenario"),
                    dolt: field(r, "dolt"),
                    contains: field(r, "contains"),
                    reason: field(r, "reason").unwrap_or_default(),
                })
                .collect(),
        })
    }

    /// reason returns why a statement's expectation is Postgres', when it is.
    fn reason(&self, suite: &str, scenario: &str, dolt: &str) -> Option<&str> {
        self.rules
            .iter()
            .find(|r| {
                r.suite.as_deref().is_none_or(|s| s == suite)
                    && r.scenario.as_deref().is_none_or(|c| c == scenario)
                    && r.dolt.as_deref().is_none_or(|d| d == dolt)
                    && r.contains.as_deref().is_none_or(|c| dolt.contains(c))
            })
            .map(|r| r.reason.as_str())
    }
}

/// run replays the corpus files of the suites named, or all of them, and reports how their statements compare with
/// their expectations.
pub fn run(corpus: &str, binary: &str, suites: &[String]) -> Result<bool, String> {
    let mut files: Vec<PathBuf> = std::fs::read_dir(corpus)
        .map_err(|err| format!("{corpus}: {err}"))?
        .filter_map(|e| e.ok().map(|e| e.path()))
        .filter(|p| p.extension().is_some_and(|e| e == "jsonl"))
        .filter(|p| suites.is_empty() || suites.iter().any(|s| p.file_stem().is_some_and(|f| f == s.as_str())))
        .collect();
    files.sort();
    let mut clean = true;
    let (mut total, mut passed, mut known, mut fixed) = (0, 0, 0, 0);
    for file in files {
        let suite = file.file_stem().map(|s| s.to_string_lossy().into_owned()).unwrap_or_default();
        let server = Server::start(&Target::Doltgres { binary: PathBuf::from(binary) }, "")?;
        let text = std::fs::read_to_string(&file).map_err(|err| err.to_string())?;
        for (index, line) in text.lines().enumerate() {
            let scenario: Json = serde_json::from_str(line).map_err(|err| format!("{}: {err}", file.display()))?;
            let report = run_scenario(&server, index, &scenario)?;
            for (statement, verdict) in report {
                total += 1;
                match verdict {
                    Verdict::Pass => passed += 1,
                    Verdict::Known => known += 1,
                    Verdict::Fixed => {
                        fixed += 1;
                        println!("{suite} {}: now matches Dolt: {statement}", scenario["name"].as_str().unwrap_or(""));
                    }
                    Verdict::Fail(detail) => {
                        clean = false;
                        println!("{suite} {}: {statement}\n    {detail}", scenario["name"].as_str().unwrap_or(""));
                    }
                }
            }
        }
    }
    println!(
        "{total} statements: {passed} pass, {known} known divergences, {fixed} now match Dolt, {} fail",
        total - passed - known - fixed
    );
    Ok(clean)
}

/// Verdict is how a replayed statement compares with its expectation.
enum Verdict {
    Pass,
    /// It still diverges from Dolt as the corpus records.
    Known,
    /// It diverged from Dolt when the corpus was made and now matches Dolt.
    Fixed,
    Fail(String),
}

/// run_scenario replays a scenario and returns the verdict of each statement and step.
fn run_scenario(server: &Server, index: usize, scenario: &Json) -> Result<Vec<(String, Verdict)>, String> {
    let mut runner = Runner::new(server, index)?;
    let mut ours = Normalizer::new(&runner.root);
    let mut verdicts = Vec::new();
    for step in scenario["steps"].as_array().into_iter().flatten() {
        let diverges = step.get("divergence").is_some_and(|d| !d.is_null());
        if let Some(sql) = step["admin"].as_str().or(step["command"].as_str()) {
            let outcome = match step["database"].as_str() {
                Some(database) => runner.command(database, sql)?,
                None => runner.admin(sql)?,
            };
            let failed = matches!(outcome, Outcome::Error(_));
            let verdict = match (failed == step["error"].as_bool().unwrap_or(false), diverges) {
                (true, false) => Verdict::Pass,
                (true, true) => Verdict::Fixed,
                (false, true) => Verdict::Known,
                (false, false) => Verdict::Fail(describe(&ours.outcome(&outcome))),
            };
            verdicts.push((sql.to_string(), verdict));
            continue;
        }
        let database = step["database"].as_str().unwrap_or_default();
        let keep_going = step["continue"].as_bool().unwrap_or(true);
        let mut session = runner.session(&ours.restore(database));
        for statement in step["statements"].as_array().into_iter().flatten() {
            let translated = map_translated(&translated_from_json(&statement["sql"]), |s| ours.restore(s));
            let outcome = session.run(&translated)?;
            let failed = matches!(outcome, Outcome::Error(_));
            let normal = ours.outcome(&outcome);
            let expect = &statement["expect"];
            let verdict = match statement.get("divergence").is_some_and(|d| !d.is_null()) {
                true if equivalent(&expected_from_json(expect), &normal) => Verdict::Fixed,
                true => Verdict::Known,
                false if outcome_json(&normal) == *expect || same_rows(&outcome_json(&normal), expect) => Verdict::Pass,
                false => Verdict::Fail(format!("expected {expect}, got {}", outcome_json(&normal))),
            };
            let label = statement.get("dolt").unwrap_or(&statement["sql"]);
            verdicts.push((label.as_str().unwrap_or_default().to_string(), verdict));
            if failed && !keep_going {
                break;
            }
        }
    }
    Ok(verdicts)
}

/// same_rows reports whether two row results hold the same rows in another order.
fn same_rows(ours: &Json, expect: &Json) -> bool {
    let rows = |j: &Json| {
        let mut rows: Vec<String> = j["rows"].as_array()?.iter().map(Json::to_string).collect();
        rows.sort();
        Some(rows)
    };
    ours["columns"] == expect["columns"] && rows(ours).is_some() && rows(ours) == rows(expect)
}

/// scenarios groups recorded sessions into scenarios by the test case directory their repositories were in. A
/// repository that a session uses before any command made it, as a copy of another, gets a database of its own.
fn scenarios(records: &[Json]) -> Vec<Scenario> {
    let mut scenarios: Vec<Scenario> = Vec::new();
    let mut repositories: Vec<HashMap<String, usize>> = Vec::new();
    for record in records {
        let suite = record["suite"].as_str().unwrap_or_default();
        let cwd = record["cwd"].as_str().unwrap_or_default();
        let Some(root) = scenario_root(cwd) else { continue };
        let at = match scenarios.iter().rposition(|s| s.suite == suite && s.root == root) {
            Some(at) => at,
            None => {
                let name = Path::new(&root).file_name().map(|n| n.to_string_lossy().into_owned()).unwrap_or_default();
                scenarios.push(Scenario {
                    suite: suite.to_string(),
                    name,
                    root: root.clone(),
                    databases: Vec::new(),
                    steps: Vec::new(),
                });
                repositories.push(HashMap::new());
                scenarios.len() - 1
            }
        };
        let (scenario, repos) = (&mut scenarios[at], &mut repositories[at]);
        let mut database = |path: &str, scenario: &mut Scenario| -> String {
            let next = repos.len();
            let n = *repos.entry(path.to_string()).or_insert(next);
            if n == scenario.databases.len() {
                scenario.databases.push(path.to_string());
            }
            format!("@db{n}@")
        };
        let args: Vec<String> =
            record["args"].as_array().into_iter().flatten().filter_map(|a| a.as_str().map(str::to_string)).collect();
        let failed = record["rc"].as_i64().unwrap_or(0) != 0;
        let (globals, command, rest) = split_command(&args);
        let use_db = globals.windows(2).find(|w| w[0] == "--use-db").map(|w| w[1].clone());
        let remap = |s: &str| {
            let remapped = s.replace(&root, "@root@");
            match root.strip_prefix("/private") {
                Some(alias) => remapped.replace(alias, "@root@"),
                None => remapped,
            }
        };
        match command.as_str() {
            "init" => {
                let db = database(cwd, scenario);
                scenario.steps.push(Step::Admin { sql: format!("CREATE DATABASE \"{db}\""), error: failed });
            }
            "sql" => {
                let (path, branch) = match &use_db {
                    Some(name) => match name.split_once('/') {
                        Some((repo, branch)) => (format!("{cwd}/{repo}"), Some(branch.to_string())),
                        None => (format!("{cwd}/{name}"), None),
                    },
                    None => (cwd.to_string(), None),
                };
                let known = scenario.databases.contains(&path);
                let mut db = database(&path, scenario);
                if !known {
                    scenario.steps.push(Step::Admin { sql: format!("CREATE DATABASE \"{db}\""), error: false });
                }
                if let Some(branch) = branch {
                    db = format!("{db}/{branch}");
                }
                let keep_going = rest.iter().any(|a| a == "-c" || a == "--continue");
                let statements = record["statements"]
                    .as_array()
                    .into_iter()
                    .flatten()
                    .map(|s| (remap(s["sql"].as_str().unwrap_or_default()), expected_from_record(s)))
                    .collect();
                scenario.steps.push(Step::Session { database: db, keep_going, statements });
            }
            "clone" => {
                let positional: Vec<&String> = rest.iter().filter(|a| !a.starts_with('-')).collect();
                let Some(url) = positional.first() else { continue };
                let dir = positional
                    .get(1)
                    .map(|d| d.to_string())
                    .unwrap_or_else(|| url.trim_end_matches('/').rsplit('/').next().unwrap_or_default().to_string());
                let path = if dir == "." { cwd.to_string() } else { format!("{cwd}/{dir}") };
                let db = database(&path, scenario);
                let sql = format!("SELECT dolt_clone('{}', '{db}')", remap(url).replace('\'', "''"));
                scenario.steps.push(Step::Admin { sql, error: failed });
            }
            "remote" | "push" | "pull" | "fetch" | "checkout" | "branch" | "merge" | "commit" | "add" | "tag"
            | "reset" => {
                let db = database(cwd, scenario);
                let arguments: Vec<String> =
                    rest.iter().map(|a| format!("'{}'", remap(a).replace('\'', "''"))).collect();
                let sql = format!("SELECT dolt_{command}({})", arguments.join(", "));
                scenario.steps.push(Step::Command { database: db, sql, error: failed });
            }
            _ => {}
        }
    }
    scenarios.retain(|s| !s.steps.is_empty());
    scenarios
}

/// scenario_root returns the directory of the test case that a path is in: the directory just inside the test's
/// temporary directory, which mktemp names `tmp.` and a suffix.
fn scenario_root(path: &str) -> Option<String> {
    let parts: Vec<&str> = path.split('/').collect();
    let at = parts.iter().position(|p| p.starts_with("tmp.") && p.len() > 4)?;
    let end = (at + 2).min(parts.len());
    Some(parts[..end].join("/"))
}

/// split_command splits dolt's arguments into the global options before the command, the command, and the rest.
fn split_command(args: &[String]) -> (Vec<String>, String, Vec<String>) {
    let mut i = 0;
    while i < args.len() && args[i].starts_with('-') {
        i += if matches!(args[i].as_str(), "--use-db" | "--data-dir" | "--branch" | "-u" | "-p") { 2 } else { 1 };
    }
    let command = args.get(i).cloned().unwrap_or_default();
    (args[..i.min(args.len())].to_vec(), command, args.get(i + 1..).map(<[String]>::to_vec).unwrap_or_default())
}

/// expected_from_record reads a recorded statement's result.
fn expected_from_record(statement: &Json) -> Expected {
    let results = statement["results"]
        .as_array()
        .into_iter()
        .flatten()
        .map(|set| {
            let columns = set["columns"]
                .as_array()
                .into_iter()
                .flatten()
                .map(|c| c.as_str().unwrap_or_default().to_string())
                .collect();
            (columns, rows_from_json(&set["rows"]))
        })
        .collect();
    Expected { error: statement["error"].as_str().map(str::to_string), results }
}

/// rows_from_json reads rows of text and nulls.
fn rows_from_json(rows: &Json) -> Vec<Vec<Option<String>>> {
    rows.as_array()
        .into_iter()
        .flatten()
        .map(|row| row.as_array().into_iter().flatten().map(|v| v.as_str().map(str::to_string)).collect())
        .collect()
}

/// Runner runs a scenario's steps in its own databases on a server.
struct Runner<'s> {
    server: &'s Server,
    prefix: String,
    root: String,
    admin: RawConnection,
}

impl<'s> Runner<'s> {
    /// new prepares a scenario's run, with a fresh directory for its file remotes.
    fn new(server: &'s Server, index: usize) -> Result<Runner<'s>, String> {
        let root = std::env::temp_dir().join(format!("dolt-oracle-{}-{index}", std::process::id()));
        let _ = std::fs::remove_dir_all(&root);
        std::fs::create_dir_all(&root).map_err(|err| err.to_string())?;
        let admin = RawConnection::connect(server.port, &[("user", "postgres"), ("database", "postgres")])?;
        Ok(Runner { server, prefix: format!("o{index}_"), root: root.to_string_lossy().into_owned(), admin })
    }

    /// resolve replaces a step's placeholders with this run's databases and directory.
    fn resolve(&self, sql: &str) -> String {
        let mut out = sql.replace("@root@", &self.root);
        while let Some(at) = out.find("@db") {
            let Some(end) = out[at + 3..].find('@') else { break };
            out.replace_range(at..at + 3 + end + 1, &format!("{}{}", self.prefix, &out[at + 3..at + 3 + end]));
        }
        out
    }

    /// admin runs a statement on the server's own database.
    fn admin(&mut self, sql: &str) -> Result<Outcome, String> {
        let sql = self.resolve(sql);
        query(&mut self.admin, &sql)
    }

    /// command runs a Dolt command's procedure in a database.
    fn command(&mut self, database: &str, sql: &str) -> Result<Outcome, String> {
        let sql = self.resolve(sql);
        self.session(database).query(&sql)
    }

    /// session opens a session in a database, whose statements fail with the connection's error when it cannot
    /// connect, as a session in a revision that does not exist does.
    fn session(&self, database: &str) -> Session<'_> {
        let name = self.resolve(database);
        let conn = RawConnection::connect(self.server.port, &[("user", "postgres"), ("database", &name)]);
        Session { runner: self, name, conn, catalog: None, autocommit: true }
    }
}

/// Session is a connection to one of a scenario's databases, as one `dolt sql` invocation is.
struct Session<'r> {
    runner: &'r Runner<'r>,
    name: String,
    conn: Result<RawConnection, String>,
    catalog: Option<RawConnection>,
    autocommit: bool,
}

impl Session<'_> {
    /// query runs a statement in the session's database.
    fn query(&mut self, sql: &str) -> Result<Outcome, String> {
        match &mut self.conn {
            Ok(conn) => query(conn, sql),
            Err(err) => Ok(Outcome::Error(err.clone())),
        }
    }

    /// translate returns the Postgres statements that stand for a statement Dolt ran.
    fn translate(&mut self, dolt_sql: &str) -> Translated {
        let (port, name) = (self.runner.server.port, self.name.clone());
        let catalog = &mut self.catalog;
        mysql::translate(dolt_sql, &mut |q| {
            if catalog.is_none() {
                *catalog = RawConnection::connect(port, &[("user", "postgres"), ("database", &name)]).ok();
                if let Some(c) = catalog.as_mut() {
                    rows_of(c, "SET dolt_show_system_tables = on");
                }
            }
            catalog.as_mut().map(|c| rows_of(c, q)).unwrap_or_default()
        })
    }

    /// run runs a translated statement, returning the outcome of its last statement or its first error.
    fn run(&mut self, translated: &Translated) -> Result<Outcome, String> {
        match translated {
            Translated::Autocommit(on) => {
                if *on && !self.autocommit {
                    self.query("COMMIT")?;
                } else if !on && self.autocommit {
                    self.query("BEGIN")?;
                }
                self.autocommit = *on;
                Ok(Outcome::Done)
            }
            Translated::Statements(statements) => {
                let atomic = !self.autocommit && !statements.iter().any(|s| transaction_control(s));
                if atomic {
                    self.query("SAVEPOINT oracle_statement")?;
                }
                let mut outcome = Outcome::Done;
                for statement in statements {
                    let statement = self.runner.resolve(statement);
                    outcome = self.query(&statement)?;
                    if matches!(outcome, Outcome::Error(_)) {
                        break;
                    }
                }
                if atomic {
                    if matches!(outcome, Outcome::Error(_)) {
                        self.query("ROLLBACK TO SAVEPOINT oracle_statement")?;
                    }
                    self.query("RELEASE SAVEPOINT oracle_statement")?;
                }
                let ends = statements.last().is_some_and(|s| {
                    let lower = s.trim().to_ascii_lowercase();
                    (lower.starts_with("commit") || lower.starts_with("rollback")) && !lower.contains(" to ")
                });
                if !self.autocommit && ends {
                    self.query("BEGIN")?;
                }
                Ok(outcome)
            }
        }
    }
}

/// query runs a statement and returns its outcome: its first error, or its last result.
fn query(conn: &mut RawConnection, sql: &str) -> Result<Outcome, String> {
    conn.send(&[FrontendMessage::Query { query: sql.to_string() }])?;
    let (mut outcome, mut error) = (Outcome::Done, None);
    let mut types: Vec<u32> = Vec::new();
    loop {
        match conn.receive(QUERY_TIMEOUT)?.ok_or_else(|| format!("timed out running {sql}"))? {
            BackendMessage::RowDescription { fields } => {
                types = fields.iter().map(|f| f.data_type_oid).collect();
                outcome = Outcome::Rows { columns: fields.iter().map(|f| f.name.clone()).collect(), rows: Vec::new() };
            }
            BackendMessage::DataRow { values } => {
                if let Outcome::Rows { rows, .. } = &mut outcome {
                    rows.push(
                        values
                            .iter()
                            .enumerate()
                            .map(|(i, v)| {
                                v.as_ref().map(|b| {
                                    let text = String::from_utf8_lossy(b).into_owned();
                                    match (types.get(i), text.as_str()) {
                                        (Some(&BOOL), "t") => "true".to_string(),
                                        (Some(&BOOL), "f") => "false".to_string(),
                                        _ => text,
                                    }
                                })
                            })
                            .collect(),
                    );
                }
            }
            BackendMessage::ErrorResponse(fields) => {
                error.get_or_insert(fields.message);
            }
            BackendMessage::ReadyForQuery { .. } => break,
            _ => {}
        }
    }
    Ok(error.map_or(outcome, Outcome::Error))
}

/// rows_of runs a query and returns its rows as text, or none when it fails.
fn rows_of(conn: &mut RawConnection, sql: &str) -> Vec<Vec<String>> {
    match query(conn, sql) {
        Ok(Outcome::Rows { rows, .. }) => {
            rows.into_iter().map(|r| r.into_iter().map(Option::unwrap_or_default).collect()).collect()
        }
        _ => Vec::new(),
    }
}

/// Normalizer writes the parts of results that differ between runs as placeholders: commit hashes, numbered in the
/// order they first appear, times, and the scenario's directory, under each name the system gives it.
struct Normalizer {
    roots: Vec<String>,
    hashes: HashMap<String, usize>,
}

impl Normalizer {
    /// placeholders replaces the hashes the normalizer has seen with their placeholders.
    fn placeholders(&self, text: &str) -> String {
        self.hashes.iter().fold(text.to_string(), |out, (hash, n)| out.replace(hash.as_str(), &format!("@hash{n}@")))
    }

    /// restore replaces placeholders with the hashes the normalizer has seen for them.
    fn restore(&self, text: &str) -> String {
        self.hashes.iter().fold(text.to_string(), |out, (hash, n)| out.replace(&format!("@hash{n}@"), hash))
    }

    /// new returns a normalizer for a scenario whose directory is the root.
    fn new(root: &str) -> Normalizer {
        let mut roots = vec![root.to_string()];
        if let Ok(canonical) = std::fs::canonicalize(root) {
            roots.push(canonical.to_string_lossy().into_owned());
        }
        let aliases: Vec<String> =
            roots.iter().filter_map(|r| r.strip_prefix("/private").map(str::to_string)).collect();
        roots.extend(aliases);
        roots.sort_by_key(|r| std::cmp::Reverse(r.len()));
        roots.dedup();
        Normalizer { roots, hashes: HashMap::new() }
    }

    /// text normalizes a value.
    fn text(&mut self, value: &str) -> String {
        let mut value = value.to_string();
        for root in &self.roots {
            value = value.replace(root.as_str(), "@root@");
        }
        let mut out = String::with_capacity(value.len());
        let bytes = value.as_bytes();
        let mut i = 0;
        while i < bytes.len() {
            let word = |at: usize| bytes.get(at).is_some_and(|b| b.is_ascii_alphanumeric());
            let hash = i + 32 <= bytes.len()
                && (i == 0 || !word(i - 1))
                && !word(i + 32)
                && bytes[i..i + 32].iter().all(|b| b.is_ascii_digit() || (b'a'..=b'v').contains(b));
            if hash && bytes[i..i + 32].iter().any(|b| b.is_ascii_alphabetic()) {
                let next = self.hashes.len();
                let n = *self.hashes.entry(value[i..i + 32].to_string()).or_insert(next);
                out.push_str(&format!("@hash{n}@"));
                i += 32;
            } else if let Some(len) = time_at(bytes, i) {
                out.push_str("@time@");
                i += len;
            } else {
                let c = value[i..].chars().next().unwrap_or_default();
                out.push(c);
                i += c.len_utf8();
            }
        }
        out
    }

    /// outcome normalizes an outcome of Doltgres.
    fn outcome(&mut self, outcome: &Outcome) -> Outcome {
        match outcome {
            Outcome::Error(message) => Outcome::Error(self.text(message)),
            Outcome::Rows { columns, rows } => Outcome::Rows {
                columns: columns.clone(),
                rows: rows.iter().map(|r| r.iter().map(|v| v.as_deref().map(|v| self.text(v))).collect()).collect(),
            },
            Outcome::Done => Outcome::Done,
        }
    }

    /// expected normalizes a result of Dolt.
    fn expected(&mut self, expected: &Expected) -> Expected {
        Expected {
            error: expected.error.as_deref().map(|e| self.text(e)),
            results: expected
                .results
                .iter()
                .map(|(c, rows)| {
                    (
                        c.clone(),
                        rows.iter().map(|r| r.iter().map(|v| v.as_deref().map(|v| self.text(v))).collect()).collect(),
                    )
                })
                .collect(),
        }
    }
}

/// time_at returns the length of a date and time that starts at a position, as `2026-01-02 03:04:05` with optional
/// fractional seconds and offset.
fn time_at(bytes: &[u8], at: usize) -> Option<usize> {
    let pattern = b"dddd-dd-dd dd:dd:dd";
    if at + pattern.len() > bytes.len() || (at > 0 && bytes[at - 1].is_ascii_digit()) {
        return None;
    }
    for (i, &p) in pattern.iter().enumerate() {
        let b = bytes[at + i];
        let ok = match p {
            b'd' => b.is_ascii_digit(),
            b' ' => b == b' ' || b == b'T',
            _ => b == p,
        };
        if !ok {
            return None;
        }
    }
    let mut len = pattern.len();
    if bytes.get(at + len) == Some(&b'.') {
        len += 1;
        while bytes.get(at + len).is_some_and(u8::is_ascii_digit) {
            len += 1;
        }
    }
    if matches!(bytes.get(at + len), Some(b'+' | b'-')) && bytes.get(at + len + 1).is_some_and(u8::is_ascii_digit) {
        len += 3;
        if bytes.get(at + len) == Some(&b':') {
            len += 3;
        }
    }
    Some(len.min(bytes.len() - at))
}

/// equivalent reports whether Doltgres did what Dolt did: both failed, or both returned the same rows apart from how
/// each writes booleans and qualifies table names, or Dolt returned no rows to compare.
fn equivalent(dolt: &Expected, ours: &Outcome) -> bool {
    match (&dolt.error, ours) {
        (Some(_), Outcome::Error(_)) => true,
        (Some(_), _) | (None, Outcome::Error(_)) => false,
        (None, Outcome::Done) => dolt.results.iter().all(|(_, rows)| rows.is_empty()),
        (None, Outcome::Rows { columns, rows }) => match dolt.results.last() {
            None => true,
            Some((names, expected)) => {
                let expected = &project(names, expected, columns);
                if expected.len() != rows.len() {
                    return false;
                }
                let same = |a: &[Vec<Option<String>>], b: &[Vec<Option<String>>]| {
                    a.iter()
                        .zip(b)
                        .all(|(x, y)| x.len() == y.len() && x.iter().zip(y).all(|(d, g)| cell_equivalent(d, g)))
                };
                if same(expected, rows) {
                    return true;
                }
                let key = |r: &Vec<Option<String>>| {
                    r.iter().map(|v| canonical(v.as_deref())).collect::<Vec<_>>().join("\u{1}")
                };
                let mut a: Vec<String> = expected.iter().map(key).collect();
                let mut b: Vec<String> = rows.iter().map(key).collect();
                a.sort();
                b.sort();
                a == b
            }
        },
    }
}

/// map_translated rewrites the text of each statement of a translation.
fn map_translated(translated: &Translated, f: impl Fn(&str) -> String) -> Translated {
    match translated {
        Translated::Statements(statements) => Translated::Statements(statements.iter().map(|s| f(s)).collect()),
        Translated::Autocommit(on) => Translated::Autocommit(*on),
    }
}

/// transaction_control reports whether a statement begins or ends a transaction or works with savepoints, which a
/// session without autocommit runs as written rather than as one statement that MySQL rolls back alone when it fails.
fn transaction_control(statement: &str) -> bool {
    let lower = statement.trim_start().to_ascii_lowercase();
    ["begin", "start", "commit", "rollback", "savepoint", "release", "end", "abort"]
        .iter()
        .any(|word| lower.split(|c: char| !c.is_ascii_alphabetic()).next() == Some(*word))
}

/// project returns Dolt's rows with only the columns that a translation kept, when ours are fewer than Dolt's and
/// each has the name of one of Dolt's, as SHOW TRIGGERS' translation keeps only the columns Postgres can show.
fn project(names: &[String], rows: &Rows, ours: &[String]) -> Rows {
    let positions: Option<Vec<usize>> = ours.iter().map(|c| names.iter().position(|n| n == c)).collect();
    match positions {
        Some(positions) if ours.len() < names.len() => {
            rows.iter().map(|row| positions.iter().map(|&p| row.get(p).cloned().flatten()).collect()).collect()
        }
        _ => rows.clone(),
    }
}

/// cell_equivalent reports whether a value Dolt returned matches the one Doltgres returned.
fn cell_equivalent(dolt: &Option<String>, ours: &Option<String>) -> bool {
    canonical(dolt.as_deref()) == canonical(ours.as_deref())
}

/// canonical writes a value in lower case with booleans as 1 and 0 and table names without the public schema, in each
/// of the fields that tabs, bars, commas, colons, and equals signs separate, since MySQL keeps identifiers' case where
/// Postgres folds them.
fn canonical(value: Option<&str>) -> String {
    let Some(value) = value else { return "\u{0}".to_string() };
    let mut out = String::new();
    let mut field = String::new();
    let flush = |field: &mut String, out: &mut String| {
        let f = field.strip_prefix("public.").unwrap_or(field).replace(" -> public.", " -> ");
        out.push_str(match f.as_str() {
            "t" | "true" => "1",
            "f" | "false" => "0",
            other => other,
        });
        field.clear();
    };
    for c in value.chars() {
        if matches!(c, '\t' | '|' | ',' | ':' | '=') {
            flush(&mut field, &mut out);
            out.push(c);
        } else {
            field.push(c);
        }
    }
    flush(&mut field, &mut out);
    out.to_lowercase()
}

/// describe summarizes an outcome for a divergence note.
fn describe(outcome: &Outcome) -> String {
    let text = outcome_json(outcome).to_string();
    match text.char_indices().nth(300) {
        Some((at, _)) => format!("{}...", &text[..at]),
        None => text,
    }
}

/// outcome_json writes an outcome as an expectation.
fn outcome_json(outcome: &Outcome) -> Json {
    match outcome {
        Outcome::Error(message) => json!({"error": message}),
        Outcome::Rows { columns, rows } => json!({"columns": columns, "rows": rows}),
        Outcome::Done => json!({"ok": true}),
    }
}

/// expected_json writes Dolt's result as an expectation: its error, or its last result set.
fn expected_json(expected: &Expected) -> Json {
    match (&expected.error, expected.results.last()) {
        (Some(error), _) => json!({"error": error, "dolt": true}),
        (None, Some((columns, rows))) => json!({"columns": columns, "rows": rows, "dolt": true}),
        (None, None) => json!({"ok": true, "dolt": true}),
    }
}

/// expected_from_json reads an expectation that records Dolt's result.
fn expected_from_json(expect: &Json) -> Expected {
    let columns = expect["columns"]
        .as_array()
        .into_iter()
        .flatten()
        .map(|c| c.as_str().unwrap_or_default().to_string())
        .collect();
    let results =
        if expect.get("rows").is_some() { vec![(columns, rows_from_json(&expect["rows"]))] } else { Vec::new() };
    Expected { error: expect["error"].as_str().map(str::to_string), results }
}

/// translated_json writes a translated statement for the corpus.
fn translated_json(translated: &Translated) -> Json {
    match translated {
        Translated::Statements(statements) if statements.len() == 1 => json!(statements[0]),
        Translated::Statements(statements) => json!(statements),
        Translated::Autocommit(on) => json!({"autocommit": on}),
    }
}

/// translated_from_json reads a translated statement from the corpus.
fn translated_from_json(sql: &Json) -> Translated {
    match sql.get("autocommit").and_then(Json::as_bool) {
        Some(on) => Translated::Autocommit(on),
        None if sql.is_string() => Translated::Statements(vec![sql.as_str().unwrap_or_default().to_string()]),
        None => Translated::Statements(
            sql.as_array().into_iter().flatten().filter_map(|s| s.as_str().map(str::to_string)).collect(),
        ),
    }
}
