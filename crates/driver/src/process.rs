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

use std::io::{BufRead, BufReader, Read};
use std::path::{Path, PathBuf};
use std::process::{Child, Command, ExitStatus, Stdio};
use std::sync::atomic::{AtomicU64, Ordering};
use std::sync::{Arc, Mutex};
use std::time::Duration;

use crate::client::Db;

/// BIN_PATH_ENV names the doltgres binary to test, which is otherwise found on the path.
pub const BIN_PATH_ENV: &str = "DOLTGRES_BIN_PATH";

/// TEST_USER_NAME and TEST_EMAIL are the commit author for every server.
pub const TEST_USER_NAME: &str = "Bats Tests";
pub const TEST_EMAIL: &str = "bats@email.fake";

/// doltgres_path returns the doltgres binary.
pub fn doltgres_path() -> PathBuf {
    PathBuf::from(std::env::var(BIN_PATH_ENV).unwrap_or_else(|_| "doltgres".to_string()))
}

/// make_temp_dir creates a new directory under the system's temporary directory.
pub fn make_temp_dir(prefix: &str) -> Result<PathBuf, String> {
    static COUNT: AtomicU64 = AtomicU64::new(0);
    let nanos = std::time::SystemTime::now().duration_since(std::time::UNIX_EPOCH).unwrap_or_default().as_nanos();
    let name = format!("{prefix}{}-{nanos}-{}", std::process::id(), COUNT.fetch_add(1, Ordering::Relaxed));
    let dir = std::env::temp_dir().join(name);
    std::fs::create_dir_all(&dir).map_err(|e| format!("{}: {e}", dir.display()))?;
    Ok(dir)
}

/// free_port returns a port that was free when it was checked.
pub fn free_port() -> Result<u16, String> {
    let listener = std::net::TcpListener::bind("127.0.0.1:0").map_err(|e| e.to_string())?;
    Ok(listener.local_addr().map_err(|e| e.to_string())?.port())
}

/// sanitize replaces path separators and spaces in a name.
pub fn sanitize(name: &str) -> String {
    name.chars().map(|c| if matches!(c, '/' | '\\' | ' ') { '_' } else { c }).collect()
}

/// DoltUser is a home for servers with its own global config, deleted when dropped.
pub struct DoltUser {
    pub dir: PathBuf,
}

impl DoltUser {
    /// new creates a user whose global config sets the commit author and disables metrics.
    pub fn new() -> Result<DoltUser, String> {
        let dir = make_temp_dir("go-sql-server-driver-")?;
        let config_dir = dir.join(".dolt");
        std::fs::create_dir_all(&config_dir).map_err(|e| e.to_string())?;
        let contents = format!(
            "{{\"metrics.disabled\":\"true\",\"user.name\":{TEST_USER_NAME:?},\"user.email\":{TEST_EMAIL:?}}}\n"
        );
        std::fs::write(config_dir.join("config_global.json"), contents).map_err(|e| e.to_string())?;
        Ok(DoltUser { dir })
    }

    /// command returns a doltgres command run from the directory with this user's global config.
    pub fn command(&self, dir: &Path, args: &[String]) -> Command {
        let mut command = Command::new(doltgres_path());
        command.args(args).current_dir(dir).env("DOLT_ROOT_PATH", &self.dir);
        command
    }

    /// make_repo_store creates a new data directory for this user.
    pub fn make_repo_store(self: &Arc<DoltUser>) -> Result<RepoStore, String> {
        let dir = self.dir.join(format!("repo-store-{}", self.count()));
        std::fs::create_dir_all(&dir).map_err(|e| e.to_string())?;
        Ok(RepoStore { user: self.clone(), dir })
    }

    fn count(&self) -> u64 {
        static COUNT: AtomicU64 = AtomicU64::new(0);
        COUNT.fetch_add(1, Ordering::Relaxed)
    }
}

impl Drop for DoltUser {
    fn drop(&mut self) {
        let _ = std::fs::remove_dir_all(&self.dir);
    }
}

/// DbFunction runs against a newly created database.
pub type DbFunction = dyn Fn(&mut Db) -> Result<(), String>;

/// RepoStore is a data directory holding databases.
#[derive(Clone)]
pub struct RepoStore {
    pub user: Arc<DoltUser>,
    pub dir: PathBuf,
}

impl RepoStore {
    /// command returns a doltgres command run from the store.
    pub fn command(&self, args: &[String]) -> Command {
        self.user.command(&self.dir, args)
    }

    /// make_repo creates a database in the store.
    pub fn make_repo(&self, name: &str) -> Result<Repo, String> {
        self.init_database(name, None)?;
        Ok(Repo { store: self.clone(), dir: self.dir.join(name), name: name.to_string() })
    }

    /// init_database briefly runs a server on the store to create the database, then calls the function with a
    /// connection to it.
    pub fn init_database(
        &self,
        name: &str,
        function: Option<&DbFunction>,
    ) -> Result<(), String> {
        let port = free_port()?;
        let config_path = self.dir.join(format!(".init-{}-config.yaml", sanitize(name)));
        let socket = std::env::temp_dir().join(format!("dg-init-{port}.sock"));
        let config =
            format!("log_level: warn\nlistener:\n  host: 127.0.0.1\n  port: {port}\n  socket: {}\n", socket.display());
        std::fs::write(&config_path, config).map_err(|e| e.to_string())?;
        let args = vec!["--data-dir=.".to_string(), format!("--config={}", config_path.display())];
        let output = Arc::new(Mutex::new(String::new()));
        let mut child = self
            .command(&args)
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .map_err(|e| format!("cannot start {}: {e}", doltgres_path().display()))?;
        let readers = capture(&mut child, output.clone(), String::new(), None, false);
        let result = (|| -> Result<(), String> {
            let mut db = Db::connect("postgres", "password", "", "127.0.0.1", port, &[]).map_err(|e| {
                format!("could not connect to init server for {name}: {e} (output: {})", output.lock().unwrap())
            })?;
            db.exec(&format!("CREATE DATABASE IF NOT EXISTS {name}"), &[])
                .map_err(|e| format!("could not create database {name}: {e} (output: {})", output.lock().unwrap()))?;
            if let Some(function) = function {
                let mut db = Db::connect("postgres", "password", name, "127.0.0.1", port, &[])
                    .map_err(|e| format!("could not connect to {name} after creating it: {e}"))?;
                function(&mut db)?;
            }
            Ok(())
        })();
        let stopped = interrupt(&child).and_then(|_| child.wait().map_err(|e| e.to_string()));
        for reader in readers {
            let _ = reader.join();
        }
        let _ = std::fs::remove_file(&config_path);
        result?;
        match stopped {
            Ok(status) if status.success() => Ok(()),
            Ok(status) => Err(format!(
                "init server for {name} did not exit cleanly: {} (output: {})",
                exit_error(status),
                output.lock().unwrap()
            )),
            Err(e) => Err(format!("could not interrupt init server for {name}: {e}")),
        }
    }
}

/// Repo is a database in a store.
#[derive(Clone)]
pub struct Repo {
    pub store: RepoStore,
    pub dir: PathBuf,
    pub name: String,
}

impl Repo {
    /// create_remote adds a remote to the database.
    pub fn create_remote(&self, name: &str, url: &str) -> Result<(), String> {
        let (name, url) = (name.to_string(), url.to_string());
        self.store.init_database(
            &self.name,
            Some(&move |db: &mut Db| {
                db.exec("SELECT dolt_remote('add', $1, $2)", &[name.clone(), url.clone()]).map_err(|e| e.to_string())
            }),
        )
    }
}

/// Visitor is called with each line the server prints.
pub type Visitor = Arc<dyn Fn(&str) + Send + Sync>;

/// capture copies a child's output into the buffer, printing each line with the name as a prefix when asked.
fn capture(
    child: &mut Child,
    output: Arc<Mutex<String>>,
    name: String,
    visitor: Option<Visitor>,
    print: bool,
) -> Vec<std::thread::JoinHandle<()>> {
    let pipes: Vec<Box<dyn Read + Send>> =
        vec![Box::new(child.stdout.take().unwrap()) as Box<dyn Read + Send>, Box::new(child.stderr.take().unwrap())];
    pipes
        .into_iter()
        .map(|pipe| {
            let (output, name, visitor) = (output.clone(), name.clone(), visitor.clone());
            std::thread::spawn(move || {
                let mut reader = BufReader::new(pipe);
                let mut line = Vec::new();
                while reader.read_until(b'\n', &mut line).unwrap_or(0) > 0 {
                    let text = String::from_utf8_lossy(&line).trim_end_matches('\n').to_string();
                    output.lock().unwrap().push_str(&format!("{text}\n"));
                    if print {
                        if name.is_empty() { println!("{text}") } else { println!("[{name}] {text}") }
                    }
                    if let Some(visitor) = &visitor {
                        visitor(&text);
                    }
                    line.clear();
                }
            })
        })
        .collect()
}

/// exit_error describes an unsuccessful exit the way Go's exec package does.
pub fn exit_error(status: ExitStatus) -> String {
    #[cfg(unix)]
    {
        use std::os::unix::process::ExitStatusExt;
        if let Some(signal) = status.signal() {
            let name = match signal {
                2 => "interrupt",
                9 => "killed",
                15 => "terminated",
                _ => "signal",
            };
            return format!("signal: {name}");
        }
    }
    format!("exit status {}", status.code().unwrap_or(-1))
}

/// signal sends a signal to the child.
#[cfg(unix)]
fn signal(child: &Child, signal: nix::sys::signal::Signal) -> Result<(), String> {
    let pid = nix::unistd::Pid::from_raw(child.id() as i32);
    nix::sys::signal::kill(pid, signal).map_err(|e| e.to_string())
}

/// interrupt asks the child to stop, like pressing Ctrl-C.
#[cfg(unix)]
fn interrupt(child: &Child) -> Result<(), String> {
    signal(child, nix::sys::signal::Signal::SIGINT)
}

/// terminate asks the child to stop gracefully.
#[cfg(unix)]
fn terminate(child: &Child) -> Result<(), String> {
    signal(child, nix::sys::signal::Signal::SIGTERM)
}

/// ServerOptions configures a server process.
#[derive(Clone, Default)]
pub struct ServerOptions {
    pub name: String,
    pub args: Vec<String>,
    pub envs: Vec<String>,
    pub port: u16,
    pub visitor: Option<Visitor>,
}

/// SqlServer is a running doltgres server.
pub struct SqlServer {
    pub name: String,
    pub port: u16,
    pub db_name: String,
    pub dir: PathBuf,
    pub args: Vec<String>,
    pub output: Arc<Mutex<String>>,
    store: RepoStore,
    envs: Vec<String>,
    visitor: Option<Visitor>,
    child: Option<Child>,
    status: Option<ExitStatus>,
    readers: Vec<std::thread::JoinHandle<()>>,
}

/// apply_envs sets `NAME=value` environment variables on a command.
fn apply_envs(command: &mut Command, envs: &[String]) {
    for env in envs {
        let (name, value) = env.split_once('=').unwrap_or((env, ""));
        command.env(name, value);
    }
}

impl SqlServer {
    /// start starts a server from the store with the options.
    pub fn start(store: &RepoStore, options: ServerOptions) -> Result<SqlServer, String> {
        let mut server = SqlServer {
            name: options.name,
            port: options.port,
            db_name: String::new(),
            dir: store.dir.clone(),
            args: options.args,
            output: Arc::new(Mutex::new(String::new())),
            store: store.clone(),
            envs: options.envs,
            visitor: options.visitor,
            child: None,
            status: None,
            readers: Vec::new(),
        };
        server.spawn()?;
        Ok(server)
    }

    fn spawn(&mut self) -> Result<(), String> {
        let mut command = self.store.command(&self.args);
        apply_envs(&mut command, &self.envs);
        let mut child = command
            .stdout(Stdio::piped())
            .stderr(Stdio::piped())
            .spawn()
            .map_err(|e| format!("cannot start {}: {e}", doltgres_path().display()))?;
        self.readers = capture(&mut child, self.output.clone(), self.name.clone(), self.visitor.clone(), true);
        self.child = Some(child);
        self.status = None;
        Ok(())
    }

    /// wait waits for the process to exit, returning its status.
    fn wait(&mut self) -> Result<ExitStatus, String> {
        if let Some(status) = self.status {
            return Ok(status);
        }
        let status = self.child.as_mut().unwrap().wait().map_err(|e| e.to_string())?;
        for reader in self.readers.drain(..) {
            let _ = reader.join();
        }
        self.status = Some(status);
        Ok(status)
    }

    /// exited returns the exit status when the process has already exited.
    pub fn exited(&mut self) -> Option<ExitStatus> {
        if self.status.is_none()
            && let Ok(Some(_)) = self.child.as_mut().unwrap().try_wait()
        {
            let _ = self.wait();
        }
        self.status
    }

    /// error_stop waits for a server that is expected to fail, returning an error unless it exited successfully.
    pub fn error_stop(&mut self) -> Result<(), String> {
        let status = self.wait()?;
        if status.success() { Ok(()) } else { Err(exit_error(status)) }
    }

    /// graceful_stop asks the server to stop and waits for it, returning an error unless it exited successfully.
    pub fn graceful_stop(&mut self) -> Result<(), String> {
        if self.exited().is_none() {
            terminate(self.child.as_ref().unwrap())?;
        }
        self.error_stop()
    }

    /// restart stops the server and starts it again, with new arguments when given and additional environment
    /// variables.
    pub fn restart(&mut self, args: Option<Vec<String>>, envs: Option<Vec<String>>) -> Result<(), String> {
        self.graceful_stop()?;
        if let Some(args) = args {
            self.args = args;
        }
        if let Some(envs) = envs {
            self.envs.extend(envs);
        }
        self.spawn()
    }

    /// output_text returns everything the server has printed.
    pub fn output_text(&self) -> String {
        self.output.lock().unwrap().clone()
    }

    /// db connects to the server, retrying while it starts.
    pub fn db(&self, user: &str, password: &str, database: &str, params: &[(String, String)]) -> Result<Db, String> {
        let database = if database.is_empty() { self.db_name.as_str() } else { database };
        let database = if database.is_empty() { user } else { database };
        Db::connect(user, password, database, "127.0.0.1", self.port, params)
    }
}

impl Drop for SqlServer {
    fn drop(&mut self) {
        if self.exited().is_none()
            && let Some(child) = self.child.as_mut()
        {
            let _ = child.kill();
            let _ = self.wait();
        }
    }
}

/// wait_briefly sleeps between retries.
pub fn wait_briefly() {
    std::thread::sleep(Duration::from_millis(50));
}
