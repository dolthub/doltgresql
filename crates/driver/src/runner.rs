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

use std::collections::HashMap;
use std::path::Path;
use std::sync::Arc;

use regex::Regex;
use yaml_rust2::Yaml;

use crate::client::Db;
use crate::model::{Connection, Query, Server, Test, TestDef, TestRepo, WithFile, parse_test_def};
use crate::process::{DoltUser, RepoStore, ServerOptions, SqlServer, sanitize, wait_briefly};
use crate::resources::Resources;
use crate::yaml;

/// GEN_DIR is the directory of generated certificates and tokens, which `$TESTGENDIR` stands for.
static GEN_DIR: std::sync::OnceLock<std::path::PathBuf> = std::sync::OnceLock::new();

/// gen_dir returns the directory of generated certificates, generating them into a new directory the first time.
pub fn gen_dir() -> Result<&'static Path, String> {
    if let Some(dir) = GEN_DIR.get() {
        return Ok(dir);
    }
    static GENERATING: std::sync::Mutex<()> = std::sync::Mutex::new(());
    let _guard = GENERATING.lock().unwrap();
    if let Some(dir) = GEN_DIR.get() {
        return Ok(dir);
    }
    let dir = crate::process::make_temp_dir("go-sql-server-driver-gen-")?;
    crate::certs::generate_x509_certs(&dir)?;
    Ok(GEN_DIR.get_or_init(|| dir))
}

/// Outcome is a test's result.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Outcome {
    Pass,
    Skip(String),
    Fail(Vec<String>),
}

/// Failure ends a test attempt.
type Failure = String;

/// replace_gen_dir replaces `$TESTGENDIR` with the directory of generated files.
fn replace_gen_dir(text: &str) -> Result<String, Failure> {
    if !text.contains("$TESTGENDIR") {
        return Ok(text.to_string());
    }
    Ok(text.replace("$TESTGENDIR", &gen_dir()?.to_string_lossy()))
}

/// regex_matches reports whether the pattern matches the text, like testify's Regexp.
fn regex_matches(pattern: &str, text: &str) -> Result<bool, Failure> {
    Regex::new(pattern).map(|r| r.is_match(text)).map_err(|e| format!("invalid regex {pattern:?}: {e}"))
}

/// write_file writes a file into a directory, templating its contents.
fn write_file(file: &WithFile, dir: &Path, base: &Path, resources: &mut Resources) -> Result<(), Failure> {
    let path = dir.join(&file.name);
    if let Some(parent) = path.parent() {
        std::fs::create_dir_all(parent).map_err(|e| e.to_string())?;
    }
    let contents = if file.source_path.is_empty() {
        file.contents.clone()
    } else {
        let source = base.join(replace_gen_dir(&file.source_path)?);
        String::from_utf8_lossy(&std::fs::read(&source).map_err(|e| format!("{}: {e}", source.display()))?).into_owned()
    };
    std::fs::write(&path, resources.apply_template(&contents)?).map_err(|e| format!("{}: {e}", path.display()))
}

/// make_repo creates a repo in the store with its files and remotes.
fn make_repo(store: &RepoStore, repo: &TestRepo, base: &Path, resources: &mut Resources) -> Result<(), Failure> {
    let created = store.make_repo(&repo.name)?;
    for file in &repo.with_files {
        write_file(file, &store.dir, base, resources)?;
    }
    for remote in &repo.with_remotes {
        created.create_remote(&remote.name, &resources.apply_template(&remote.url)?)?;
    }
    Ok(())
}

/// prepare_server_args merges the listener into the config file named by the arguments, writes the merged config,
/// and returns the arguments that point the server at it.
pub fn prepare_server_args(cwd: &Path, name: &str, port: u16, args: &[String]) -> Result<Vec<String>, Failure> {
    let mut passthrough = Vec::new();
    let mut config_path = None;
    let mut i = 0;
    while i < args.len() {
        let arg = &args[i];
        if arg == "--config" || arg == "-config" {
            config_path = args.get(i + 1).cloned();
            i += 1;
        } else if let Some(path) = arg.strip_prefix("--config=").or_else(|| arg.strip_prefix("-config=")) {
            config_path = Some(path.to_string());
        } else {
            passthrough.push(arg.clone());
        }
        i += 1;
    }
    let mut base = yaml_rust2::yaml::Hash::new();
    if let Some(path) = config_path {
        let path = if Path::new(&path).is_absolute() { Path::new(&path).to_path_buf() } else { cwd.join(path) };
        let contents = std::fs::read_to_string(&path).map_err(|e| format!("{}: {e}", path.display()))?;
        if let Ok(Yaml::Hash(hash)) = yaml::parse(&contents).map(|n| n.to_yaml()) {
            base = hash;
        }
    }
    let listener_key = Yaml::String("listener".to_string());
    let mut listener = match base.get(&listener_key) {
        Some(Yaml::Hash(hash)) => hash.clone(),
        _ => yaml_rust2::yaml::Hash::new(),
    };
    let socket = std::env::temp_dir().join(format!("dg-{port}.sock"));
    listener.insert(Yaml::String("host".into()), Yaml::String("127.0.0.1".into()));
    listener.insert(Yaml::String("port".into()), Yaml::Integer(port as i64));
    listener.insert(Yaml::String("socket".into()), Yaml::String(socket.to_string_lossy().into_owned()));
    base.insert(listener_key, Yaml::Hash(listener));
    let generated = cwd.join(format!(".generated-{}-config.yaml", sanitize(name)));
    std::fs::write(&generated, yaml::emit(&Yaml::Hash(base))?).map_err(|e| e.to_string())?;
    let has_data_dir = passthrough
        .iter()
        .any(|a| a == "--data-dir" || a == "-data-dir" || a.starts_with("--data-dir=") || a.starts_with("-data-dir="));
    let mut result = passthrough;
    if !has_data_dir {
        result.push("--data-dir=.".to_string());
    }
    result.push(format!("--config={}", generated.display()));
    Ok(result)
}

/// Env holds a test's resources and servers, and the checks that run when it finishes.
#[derive(Default)]
pub struct Env {
    pub resources: Resources,
    pub base: std::path::PathBuf,
    users: Vec<Arc<DoltUser>>,
    servers: HashMap<String, SqlServer>,
    order: Vec<String>,
    log_checks: HashMap<String, (Vec<String>, Vec<String>)>,
}

impl Env {
    /// new returns an environment that reads relative source paths from the base directory.
    pub fn new(base: &Path) -> Env {
        Env { base: base.to_path_buf(), ..Env::default() }
    }

    /// user creates a user that lives until the environment is dropped.
    pub fn user(&mut self) -> Result<Arc<DoltUser>, Failure> {
        let user = Arc::new(DoltUser::new()?);
        self.users.push(user.clone());
        Ok(user)
    }

    /// server returns a started server by its key.
    pub fn server(&mut self, key: &str) -> &mut SqlServer {
        self.servers.get_mut(key).unwrap()
    }

    /// start_server starts a server under the key, returning false when it was expected to fail and did.
    pub fn start_server(
        &mut self,
        key: &str,
        store: &RepoStore,
        server: &Server,
        visitor: Option<crate::process::Visitor>,
    ) -> Result<bool, Failure> {
        let name = if server.name.is_empty() { key.to_string() } else { server.name.clone() };
        match self.make_server(store, server, &name, visitor)? {
            Some(started) => {
                self.add_server(key, started);
                Ok(true)
            }
            None => Ok(false),
        }
    }

    /// make_server starts a server, or checks the failure of one that is expected to fail.
    fn make_server(
        &mut self,
        store: &RepoStore,
        server: &Server,
        name: &str,
        visitor: Option<crate::process::Visitor>,
    ) -> Result<Option<SqlServer>, Failure> {
        if server.port != 0 {
            return Err("cannot specify s.Port on these tests; please use {{get_port ...}} and dynamic_port: to \
                        specify a dynamic port."
                .to_string());
        }
        if server.dynamic_port.is_empty() {
            return Err("you must specify s.DynamicPort on these tests; please use {{get_port ...}} and \
                        dynamic_port: to specify a dynamic port."
                .to_string());
        }
        let port = self.resources.port(&server.dynamic_port)?;
        let args = server.args.iter().map(|a| self.resources.apply_template(a)).collect::<Result<Vec<_>, _>>()?;
        let args = prepare_server_args(&store.dir, name, port, &args)?;
        let mut started = SqlServer::start(
            store,
            ServerOptions { name: name.to_string(), args, envs: server.envs.clone(), port, visitor },
        )?;
        if !server.error_matches.is_empty() {
            if started.error_stop().is_ok() {
                return Err(format!("expected server {name} to exit with an error"));
            }
            let output = started.output_text();
            for pattern in &server.error_matches {
                if !regex_matches(pattern, &output)? {
                    return Err(format!("expected {pattern:?} to match the server output:\n{output}"));
                }
            }
            return Ok(None);
        }
        self.log_checks.insert(name.to_string(), (server.log_matches.clone(), server.log_not_matches.clone()));
        Ok(Some(started))
    }

    fn add_server(&mut self, key: &str, server: SqlServer) {
        self.order.push(key.to_string());
        self.servers.insert(key.to_string(), server);
    }

    /// finish stops the servers in reverse order and checks their logs, returning the problems.
    pub fn finish(&mut self) -> Vec<String> {
        let mut problems = Vec::new();
        for key in self.order.iter().rev() {
            let Some(server) = self.servers.get_mut(key) else {
                continue;
            };
            if let Err(err) = server.graceful_stop() {
                problems.push(format!("server {}: {err}", server.name));
                continue;
            }
            let output = server.output_text();
            if let Some((matches, not_matches)) = self.log_checks.get(&server.name) {
                for pattern in matches {
                    match regex_matches(pattern, &output) {
                        Ok(true) => {}
                        Ok(false) => problems.push(format!("expected {pattern:?} to match the log of {}", server.name)),
                        Err(e) => problems.push(e),
                    }
                }
                for pattern in not_matches {
                    match regex_matches(pattern, &output) {
                        Ok(false) => {}
                        Ok(true) => {
                            problems.push(format!("expected {pattern:?} not to match the log of {}", server.name))
                        }
                        Err(e) => problems.push(e),
                    }
                }
            }
        }
        problems
    }
}

/// retry runs the attempt until it succeeds, up to the given number of times, waiting between attempts.
fn retry(attempts: i64, mut attempt: impl FnMut() -> Result<(), Failure>) -> Result<(), Failure> {
    let attempts = attempts.max(1);
    let mut last = Ok(());
    for i in 0..attempts {
        if i != 0 {
            wait_briefly();
        }
        last = attempt();
        if last.is_ok() {
            return last;
        }
    }
    last
}

/// password returns the connection's password, from its password file when it has one.
fn password(connection: &Connection) -> Result<String, Failure> {
    if !connection.password_file.is_empty() {
        let path = replace_gen_dir(&connection.password_file)?;
        return Ok(std::fs::read_to_string(&path).map_err(|e| format!("{path}: {e}"))?.trim().to_string());
    }
    Ok(if connection.password.is_empty() { "password".to_string() } else { connection.password.clone() })
}

/// open connects for a connection block.
fn open(server: &SqlServer, connection: &Connection) -> Result<Db, Failure> {
    server.db(&connection.user, &password(connection)?, &connection.database, &connection.driver_params)
}

/// run_query_attempt runs a query or statement once and checks its result.
pub fn run_query_attempt(db: &mut Db, query: &Query, resources: &mut Resources) -> Result<(), Failure> {
    if !query.query.is_empty() {
        let result = db.query(&query.query, &query.args);
        if !query.error_match.is_empty() {
            let err = result.err().ok_or_else(|| format!("expected error running query {}", query.query))?;
            if !regex_matches(&query.error_match, &err)? {
                return Err(format!("expected {:?} to match the error {err:?}", query.error_match));
            }
            return Ok(());
        }
        let result = result.map_err(|e| format!("query {}: {e}", query.query))?;
        if result.columns != query.columns {
            return Err(format!(
                "query {}: expected columns {:?}, got {:?}",
                query.query, query.columns, result.columns
            ));
        }
        if let Some(options) = &query.rows {
            let mut expanded = Vec::with_capacity(options.len());
            for option in options {
                let mut rows = Vec::with_capacity(option.len());
                for row in option {
                    rows.push(row.iter().map(|v| resources.apply_template(v)).collect::<Result<Vec<_>, _>>()?);
                }
                expanded.push(rows);
            }
            if !expanded.contains(&result.rows) {
                return Err(format!("query {}: expected one of {expanded:?}, got {:?}", query.query, result.rows));
            }
        }
    } else if !query.exec.is_empty() {
        let exec = resources.apply_template(&query.exec)?;
        let result = db.exec(&exec, &query.args);
        if query.error_match.is_empty() {
            result.map_err(|e| format!("error running query {}: {e}", query.exec))?;
        } else {
            let err = result.err().ok_or_else(|| format!("expected error running statement {}", query.exec))?;
            if !regex_matches(&query.error_match, &err)? {
                return Err(format!("expected {:?} to match the error {err:?}", query.error_match));
            }
        }
    }
    Ok(())
}

/// run_steps creates the test's repos and servers and runs its connections.
fn run_steps(test: &Test, run: &mut Env) -> Result<(), Failure> {
    for repo in &test.repos {
        let user = run.user()?;
        let store = user.make_repo_store()?;
        make_repo(&store, repo, &run.base, &mut run.resources)?;
        if let Some(server) = &repo.server {
            let name = if server.name.is_empty() { repo.name.clone() } else { server.name.clone() };
            if let Some(mut started) = run.make_server(&store, server, &name, None)? {
                started.db_name = repo.name.clone();
                run.add_server(&repo.name, started);
            }
        }
    }
    for multi in &test.multi_repos {
        let user = run.user()?;
        let store = user.make_repo_store()?;
        for repo in &multi.repos {
            make_repo(&store, repo, &run.base, &mut run.resources)?;
        }
        for file in &multi.with_files {
            write_file(file, &store.dir, &run.base, &mut run.resources)?;
        }
        if let Some(server) = &multi.server {
            let name = if server.name.is_empty() { multi.name.clone() } else { server.name.clone() };
            if let Some(started) = run.make_server(&store, server, &name, None)? {
                run.add_server(&multi.name, started);
            }
        }
    }
    for (index, connection) in test.connections.iter().enumerate() {
        let server = run.servers.get(&connection.on).ok_or_else(|| {
            format!("error in test spec: could not find server {} for connection {index}", connection.on)
        })?;
        let resources = &mut run.resources;
        if connection.retry_attempts > 1 {
            retry(connection.retry_attempts, || {
                let mut db = open(server, connection)?;
                for query in &connection.queries {
                    run_query_attempt(&mut db, query, resources)?;
                }
                Ok(())
            })?;
        } else {
            let mut db = open(server, connection)?;
            for query in &connection.queries {
                retry(query.retry_attempts, || run_query_attempt(&mut db, query, resources))?;
            }
            db.close();
        }
        if let Some(restart) = &connection.restart_server {
            let server = run.servers.get_mut(&connection.on).unwrap();
            let args = match &restart.args {
                Some(args) => {
                    let args = args.iter().map(|a| run.resources.apply_template(a)).collect::<Result<Vec<_>, _>>()?;
                    Some(prepare_server_args(&server.dir, &server.name, server.port, &args)?)
                }
                None => None,
            };
            server.restart(args, restart.envs.clone())?;
        }
    }
    Ok(())
}

/// Slot is a permit to run a test, of which there are as many as Go's default test parallelism.
struct Slot;

static RUNNING: (std::sync::Mutex<usize>, std::sync::Condvar) = (std::sync::Mutex::new(0), std::sync::Condvar::new());

impl Slot {
    fn acquire() -> Slot {
        let limit = std::thread::available_parallelism().map_or(1, |n| n.get());
        let mut running = RUNNING.0.lock().unwrap();
        while *running >= limit {
            running = RUNNING.1.wait(running).unwrap();
        }
        *running += 1;
        Slot
    }
}

impl Drop for Slot {
    fn drop(&mut self) {
        *RUNNING.0.lock().unwrap() -= 1;
        RUNNING.1.notify_one();
    }
}

/// run_test runs a test, stopping its servers and checking their logs when it finishes.
pub fn run_test(test: &Test, base: &Path) -> Outcome {
    if !test.skip.is_empty() {
        return Outcome::Skip(test.skip.clone());
    }
    let _slot = Slot::acquire();
    let mut run = Env::new(base);
    let mut problems = Vec::new();
    if let Err(err) = run_steps(test, &mut run) {
        problems.push(err);
    }
    problems.extend(run.finish());
    if problems.is_empty() { Outcome::Pass } else { Outcome::Fail(problems) }
}

/// parse_tests_file parses a test file.
pub fn parse_tests_file(path: &Path) -> Result<TestDef, String> {
    let text = std::fs::read_to_string(path).map_err(|e| format!("{}: {e}", path.display()))?;
    parse_test_def(&yaml::parse(&text)?)
}

/// run_tests_file runs every test of a file, in parallel when the file asks for it, returning each outcome.
pub fn run_tests_file(path: &Path) -> Result<Vec<(String, Outcome)>, String> {
    let def = parse_tests_file(path)?;
    let base = path.parent().and_then(Path::parent).unwrap_or(Path::new("."));
    if !def.parallel {
        return Ok(def.tests.iter().map(|test| (test.name.clone(), run_test(test, base))).collect());
    }
    std::thread::scope(|scope| {
        let handles: Vec<_> =
            def.tests.iter().map(|test| (test.name.clone(), scope.spawn(move || run_test(test, base)))).collect();
        Ok(handles
            .into_iter()
            .map(|(name, handle)| (name, handle.join().unwrap_or_else(|_| Outcome::Fail(vec!["panicked".into()]))))
            .collect())
    })
}
