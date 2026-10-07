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

//! The engine, which serves a data directory's databases, and the sessions that run statements in transactions.

use std::collections::HashMap;
use std::path::{Path, PathBuf};
use std::sync::{Arc, Mutex};
use std::time::{SystemTime, UNIX_EPOCH};

use doltdb::database::Database;
use pg_query::protobuf::a_const::Val;
use pg_query::protobuf::{RangeVar, TransactionStmtKind, VariableSetKind, VariableSetStmt};
use pg_query::{Node, NodeEnum};

use crate::error::{PgError, Result, code};
use crate::parse::{self, Extras, Statement};
use crate::plan::Planner;
use crate::query::{Ctx, column};
use crate::settings::{Settings, setting};
use crate::txn::{DbHandle, SequenceTracker, Txn};
use crate::types::Value;
use crate::{Column, DEFAULT_BRANCH, Outcome, Prepared, Results};

/// Engine serves the databases in a data directory. Clones share the same databases.
#[derive(Clone)]
pub struct Engine {
    shared: Arc<Shared>,
}

/// Shared is what every clone of an engine shares.
struct Shared {
    data_dir: PathBuf,
    /// The superuser, whose name the default database takes.
    superuser: String,
    /// The roles and privileges.
    auth: Arc<Mutex<crate::auth::AuthDb>>,
    /// The branch control tables.
    branch_control: Mutex<crate::dolt::branch_control::Controller>,
    /// The port the server listens on, which the `port` parameter shows.
    port: std::sync::atomic::AtomicU16,
    databases: Mutex<HashMap<String, (DbHandle, SequenceTracker)>>,
    /// The advisory locks that sessions hold.
    advisory: Arc<crate::advisory::AdvisoryLocks>,
    /// When the engine opened, as a UTC timestamp.
    started: i64,
    /// The histograms that ANALYZE built, by database and branch.
    statistics: Mutex<HashMap<(String, String), Vec<crate::stats::Statistic>>>,
    /// What each open session is doing, by session ID.
    activity: Mutex<std::collections::BTreeMap<u64, Activity>>,
    /// The sessions that a garbage collection ended because they had a transaction open, by session ID.
    ended: Mutex<std::collections::HashSet<u64>>,
}

/// Activity is what a session is doing, as pg_stat_activity shows it.
#[derive(Clone, Debug, Default)]
pub struct Activity {
    pub database: String,
    pub user: String,
    /// The address the client connected from.
    pub host: String,
    /// The running statement, or the last one when the session is idle.
    pub query: String,
    /// When the running statement began, as a UTC timestamp, or None when the session is idle.
    pub started: Option<i64>,
    /// Whether the session has an explicit transaction open.
    pub in_transaction: bool,
}

/// undrop_hint lists the dropped databases that dolt_undrop can restore, as Dolt's CreateUndropErrorMessage does.
pub fn undrop_hint(available: &[String]) -> String {
    match available.is_empty() {
        true => "there are no databases currently available to be undropped".to_string(),
        false => format!("available databases that can be undropped: {}", available.join(", ")),
    }
}

/// DROPPED_DATABASES is the directory in the data directory that holds dropped databases.
const DROPPED_DATABASES: &str = ".dolt_dropped_databases";

/// create_times returns the clock readings of creating a database now, with the CREATE DATABASE commit a millisecond
/// after the initial one so that ordering commits by date never ties them.
fn create_times() -> doltdb::create::CreateTimes {
    let millis = SystemTime::now().duration_since(UNIX_EPOCH).map_or(0, |d| d.as_millis() as u64);
    doltdb::create::CreateTimes {
        init_author_millis: millis as i64,
        init_committer_millis: millis,
        environment_seconds: millis / 1000,
        session_seconds: millis / 1000,
        commit_millis: millis + 1,
    }
}

/// lock locks a shared value, failing when a panic poisoned it.
fn lock<T>(mutex: &Mutex<T>) -> Result<std::sync::MutexGuard<'_, T>> {
    mutex.lock().map_err(|_| PgError::internal("a lock was poisoned"))
}

impl Engine {
    /// open opens the data directory, the auth file, and the branch control file when there is one, creating them,
    /// the default database named after the superuser, and the superuser's role when they do not exist, as the Go
    /// server does on its first start.
    pub fn open(
        data_dir: &Path,
        superuser: &str,
        password: &str,
        auth_file: &Path,
        branch_control_file: Option<&Path>,
    ) -> Result<Engine> {
        std::fs::create_dir_all(data_dir.join(".dolt")).map_err(PgError::internal)?;
        let auth = crate::auth::AuthDb::open(auth_file, superuser, password)?;
        let branch_control = crate::dolt::branch_control::Controller::load(branch_control_file.map(Path::to_path_buf))?;
        let engine = Engine {
            shared: Arc::new(Shared {
                data_dir: data_dir.to_path_buf(),
                superuser: superuser.to_string(),
                auth: Arc::new(Mutex::new(auth)),
                branch_control: Mutex::new(branch_control),
                port: std::sync::atomic::AtomicU16::new(5432),
                databases: Mutex::new(HashMap::new()),
                advisory: Arc::default(),
                started: crate::datetime::clock(),
                statistics: Mutex::default(),
                activity: Mutex::default(),
                ended: Mutex::default(),
            }),
        };
        if !engine.database_exists(superuser) {
            let dir = data_dir.join(superuser);
            doltdb::create::create_database(&dir, DEFAULT_BRANCH, superuser, "localhost", &create_times())?;
        }
        Ok(engine)
    }

    /// set_port records the port the server listens on.
    pub fn set_port(&self, port: u16) {
        self.shared.port.store(port, std::sync::atomic::Ordering::Relaxed);
    }

    /// branch_control returns the branch control tables.
    pub fn branch_control(&self) -> &Mutex<crate::dolt::branch_control::Controller> {
        &self.shared.branch_control
    }

    /// put_statistics replaces the histograms of a table on a branch of a database.
    pub fn put_statistics(
        &self,
        database: &str,
        branch: &str,
        schema: &str,
        table: &str,
        statistics: Vec<crate::stats::Statistic>,
    ) {
        let Ok(mut all) = self.shared.statistics.lock() else { return };
        let entry = all.entry((database.to_string(), branch.to_string())).or_default();
        entry.retain(|s| s.schema != schema || s.table != table);
        entry.extend(statistics);
    }

    /// statistics returns the histograms of the tables on a branch of a database.
    pub fn statistics(&self, database: &str, branch: &str) -> Vec<crate::stats::Statistic> {
        let Ok(all) = self.shared.statistics.lock() else { return Vec::new() };
        all.get(&(database.to_string(), branch.to_string())).cloned().unwrap_or_default()
    }

    /// activity returns what each open session is doing, by session ID.
    pub fn activity(&self) -> Vec<(u64, Activity)> {
        let Ok(all) = self.shared.activity.lock() else { return Vec::new() };
        all.iter().map(|(id, activity)| (*id, activity.clone())).collect()
    }

    /// update_activity changes what a session is doing.
    fn update_activity(&self, id: u64, update: impl FnOnce(&mut Activity)) {
        let Ok(mut all) = self.shared.activity.lock() else { return };
        update(all.entry(id).or_default());
    }

    /// forget_activity forgets a session that closed.
    fn forget_activity(&self, id: u64) {
        let Ok(mut all) = self.shared.activity.lock() else { return };
        all.remove(&id);
    }

    /// collected records a garbage collection of a database, which ends the other sessions that use it with a
    /// transaction open, since the collection may have removed chunks that only their transactions refer to.
    pub fn collected(&self, database: &str, caller: u64) {
        let in_transaction: Vec<u64> = self
            .activity()
            .into_iter()
            .filter(|(id, a)| *id != caller && a.database == database && a.in_transaction)
            .map(|(id, _)| id)
            .collect();
        if let Ok(mut ended) = self.shared.ended.lock() {
            ended.extend(in_transaction);
        }
    }

    /// ended reports whether a garbage collection ended a session.
    fn ended(&self, id: u64) -> bool {
        self.shared.ended.lock().is_ok_and(|ended| ended.contains(&id))
    }

    /// started returns when the engine opened, as a UTC timestamp.
    pub fn started(&self) -> i64 {
        self.shared.started
    }

    /// login returns the stored password of a role and whether it may log in, or None when the role does not exist.
    pub fn login(&self, user: &str) -> Option<(Option<crate::auth::Password>, bool)> {
        let auth = self.shared.auth.lock().ok()?;
        let role = auth.role(user)?;
        Some((role.password.clone(), role.login))
    }

    /// database_exists reports whether the data directory holds the database.
    pub fn database_exists(&self, name: &str) -> bool {
        !name.is_empty() && !name.contains(['/', '\\']) && self.shared.data_dir.join(name).join(".dolt").is_dir()
    }

    /// create_database creates a database in the data directory, as the user connected from the host.
    pub fn create_database(&self, name: &str, user: &str, host: &str) -> Result<()> {
        let dir = self.shared.data_dir.join(name);
        Ok(doltdb::create::create_database(&dir, DEFAULT_BRANCH, user, host, &create_times())?)
    }

    /// drop_database closes a database and moves its directory into the dropped databases directory, moving aside
    /// an earlier dropped database of the same name, so that dolt_undrop can restore it, as Dolt's
    /// droppedDatabaseManager does.
    pub fn drop_database(&self, name: &str) -> Result<()> {
        lock(&self.shared.databases)?.remove(name);
        let dropped = self.shared.data_dir.join(DROPPED_DATABASES);
        std::fs::create_dir_all(&dropped).map_err(PgError::internal)?;
        let target = dropped.join(name);
        if target.exists() {
            let millis = SystemTime::now().duration_since(UNIX_EPOCH).map_or(0, |d| d.as_millis());
            std::fs::rename(&target, dropped.join(format!("{name}.backup.{millis}"))).map_err(|e| {
                PgError::internal(format!("unable to move existing dropped database out of the way: {e}"))
            })?;
        }
        std::fs::rename(self.shared.data_dir.join(name), target).map_err(PgError::internal)
    }

    /// dropped_databases returns the names of the dropped databases that dolt_undrop can restore.
    pub fn dropped_databases(&self) -> Vec<String> {
        let mut names: Vec<String> = std::fs::read_dir(self.shared.data_dir.join(DROPPED_DATABASES))
            .map(|entries| entries.filter_map(|e| e.ok()?.file_name().into_string().ok()).collect())
            .unwrap_or_default();
        names.sort();
        names
    }

    /// undrop_database moves a dropped database back into the data directory, matching its name without regard to
    /// case, and returns its name as it was dropped.
    pub fn undrop_database(&self, name: &str) -> Result<String> {
        let dropped = self.shared.data_dir.join(DROPPED_DATABASES);
        let available = self.dropped_databases();
        let Some(exact) = available.iter().find(|n| n.eq_ignore_ascii_case(name)).cloned() else {
            return Err(PgError::internal(format!(
                "no database named '{name}' found to undrop. {}",
                undrop_hint(&available)
            )));
        };
        let existing = std::fs::read_dir(&self.shared.data_dir).map_err(PgError::internal)?;
        if existing.flatten().any(|e| e.file_name().to_string_lossy().eq_ignore_ascii_case(&exact)) {
            return Err(PgError::internal(format!(
                "unable to undrop database '{exact}'; another database already exists with the same case-insensitive \
                 name"
            )));
        }
        std::fs::rename(dropped.join(&exact), self.shared.data_dir.join(&exact)).map_err(PgError::internal)?;
        Ok(exact)
    }

    /// purge_dropped_databases deletes every dropped database.
    pub fn purge_dropped_databases(&self) -> Result<()> {
        let dropped = self.shared.data_dir.join(DROPPED_DATABASES);
        if !dropped.exists() {
            return Ok(());
        }
        for entry in std::fs::read_dir(&dropped).map_err(PgError::internal)?.flatten() {
            let path = entry.path();
            let removed = if path.is_dir() { std::fs::remove_dir_all(&path) } else { std::fs::remove_file(&path) };
            removed.map_err(PgError::internal)?;
        }
        Ok(())
    }

    /// database returns the shared handle of a database, opening it on first use.
    fn database(&self, name: &str) -> Result<DbHandle> {
        Ok(self.open_database(name)?.0)
    }

    /// open_database returns the shared handle and sequence tracker of a database, opening it on first use.
    fn open_database(&self, name: &str) -> Result<(DbHandle, SequenceTracker)> {
        let mut databases = lock(&self.shared.databases)?;
        if let Some(entry) = databases.get(name) {
            return Ok(entry.clone());
        }
        let handle = Arc::new(Mutex::new(Database::open(&self.shared.data_dir.join(name).join(".dolt/noms"))?));
        let entry = (handle, SequenceTracker::default());
        databases.insert(name.to_string(), entry.clone());
        Ok(entry)
    }

    /// session starts a session for the user, connected from the host, on a database or a branch of one written as
    /// `database/branch`, with the parameters the client sent at startup.
    pub fn session(&self, user: &str, host: &str, database: &str, startup: &[(String, String)]) -> Result<Session> {
        let mut session = Session {
            state: SessionState {
                engine: self.clone(),
                user: user.to_string(),
                host: host.to_string(),
                database: String::new(),
                branch: DEFAULT_BRANCH.to_string(),
                display: String::new(),
                source: String::new(),
                notices: Vec::new(),
                settings: Settings::new(startup, self.shared.port.load(std::sync::atomic::Ordering::Relaxed))
                    .map_err(|err| PgError { severity: "FATAL", ..err })?,
                explicit: false,
                sequence_values: HashMap::new(),
                last_sequence: None,
                data_dir: self.shared.data_dir.clone(),
                superuser: self.shared.superuser.clone(),
                auth: self.shared.auth.clone(),
                role: user.to_string(),
                authenticated: user.to_string(),
                view_schema: None,
                routines: None,
                triggers: None,
                operators: None,
                casts: None,
                aggregates: None,
                user_types: None,
                call_depth: 0,
                id: NEXT_SESSION.fetch_add(1, std::sync::atomic::Ordering::Relaxed),
                advisory: self.shared.advisory.clone(),
                pending_copy: None,
                as_of: Vec::new(),
                deferred: crate::deferred::Deferred::default(),
                checked_out: HashMap::new(),
            },
            txns: Vec::new(),
            pending: None,
            failed: false,
            savepoints: Vec::new(),
            reported: HashMap::new(),
        };
        session.state.settings.set_raw("session_authorization", Some(user.to_string()), false, false);
        session.switch(database).map_err(|_| {
            PgError::fatal(code::INVALID_CATALOG_NAME, format!("database \"{database}\" does not exist"))
        })?;
        session.report_activity(None);
        Ok(session)
    }
}

/// Session runs statements for one connection.
pub struct Session {
    pub state: SessionState,
    /// The open transaction's view of each branch it touched.
    txns: Vec<Txn>,
    /// Whether a statement failed in the explicit transaction, which then only ends or rolls back to a savepoint.
    failed: bool,
    /// The savepoints of the explicit transaction, each with the transaction's state and the settings when it was
    /// made.
    savepoints: Vec<(String, Vec<Txn>, crate::settings::Settings)>,
    /// The reported parameters as the client last heard them.
    reported: HashMap<String, String>,
    /// The statements of a simple query that wait for its COPY FROM STDIN to finish, or None when the extended
    /// protocol began the copy.
    pending: Option<Vec<Statement>>,
}

/// RoutineCache is the functions and procedures of a root value, with the addresses of their collections.
pub type RoutineCache =
    ((Option<store::Hash>, Option<store::Hash>, Option<store::Hash>), Arc<Vec<Arc<crate::routines::Routine>>>);

/// TriggerCache is the triggers of a root value, with the address of their collection.
pub type TriggerCache = (Option<store::Hash>, Arc<Vec<Arc<objects::Trigger>>>);

/// SessionState is the part of a session that statements and functions can read and change.
pub struct SessionState {
    /// The engine the session runs on.
    pub engine: Engine,
    pub user: String,
    /// The address the client connected from.
    pub host: String,
    pub database: String,
    pub branch: String,
    /// The current database as the session named it, which includes the branch when one was named.
    pub display: String,
    /// The text of the query running, which statement locations point into.
    pub source: String,
    pub notices: Vec<PgError>,
    pub settings: Settings,
    /// Whether the open transaction began with BEGIN.
    pub explicit: bool,
    /// The last value nextval returned for each sequence in this session, by sequence ID.
    pub sequence_values: HashMap<Vec<u8>, i64>,
    /// The sequence ID and value of the session's most recent nextval.
    pub last_sequence: Option<(Vec<u8>, i64)>,
    /// The directory that holds the databases.
    pub data_dir: PathBuf,
    pub superuser: String,
    pub auth: Arc<Mutex<crate::auth::AuthDb>>,
    /// The current role, which SET ROLE changes from the session user.
    pub role: String,
    /// The user that logged in, which SET SESSION AUTHORIZATION checks.
    pub authenticated: String,
    /// The schema of the view whose query is being planned, which its unqualified names search first, as if they
    /// were bound when the view was created.
    pub view_schema: Option<String>,

    /// The functions and procedures last loaded, with the addresses of the collections they were loaded from.
    pub routines: Option<RoutineCache>,
    /// The triggers last loaded, with the address of the trigger collection they were loaded from.
    pub triggers: Option<TriggerCache>,
    /// The user-defined types last loaded, with the address of the type collection they were loaded from.
    pub user_types: Option<(Option<store::Hash>, Arc<crate::usertypes::Types>)>,
    /// The operators last loaded, with the addresses of the collections they were loaded from.
    pub operators: Option<crate::operators::OperatorCache>,
    /// The casts last loaded, with the addresses of the collections they were loaded from.
    pub casts: Option<crate::casts::CastCache>,
    /// The aggregates last loaded, with the addresses of the collections they were loaded from.
    pub aggregates: Option<crate::aggregates::AggregateCache>,
    /// How many function calls are running inside one another.
    pub call_depth: usize,
    /// The session's number among the engine's sessions, which advisory locks record their holders by.
    pub id: u64,
    /// The engine's advisory locks.
    pub advisory: Arc<crate::advisory::AdvisoryLocks>,
    /// The COPY FROM STDIN waiting for its data.
    pub pending_copy: Option<Box<crate::copy::CopyFrom>>,
    /// The `AS OF` revisions of the running statement's tables, each with the location of the table it follows.
    pub as_of: Vec<(i32, pg_query::Node)>,
    /// The transaction's constraint modes and the checks its deferred constraints owe.
    pub deferred: crate::deferred::Deferred,
    /// The branch the session last had checked out in each database it left, which a USE of the database without a
    /// branch returns to.
    pub checked_out: HashMap<String, String>,
}

/// NEXT_SESSION numbers the sessions of the process.
static NEXT_SESSION: std::sync::atomic::AtomicU64 = std::sync::atomic::AtomicU64::new(1);

impl Drop for Session {
    fn drop(&mut self) {
        self.state.advisory.release_all(self.state.id, false, true);
        self.state.engine.forget_activity(self.state.id);
        if let Ok(mut ended) = self.state.engine.shared.ended.lock() {
            ended.remove(&self.state.id);
        }
    }
}

impl SessionState {
    /// end_transaction ends the transaction's settings, undoing every change when it rolled back, and releases the
    /// advisory locks it took.
    pub fn end_transaction(&mut self, committed: bool) {
        self.settings.end_transaction(committed);
        self.deferred = crate::deferred::Deferred::default();
        self.advisory.release_all(self.id, true, false);
    }

    /// checked_out_branch returns the branch the session last had checked out in a database it left, or the
    /// database's default branch.
    pub fn checked_out_branch(&self, database: &str) -> String {
        if let Some(branch) = self.checked_out.get(database) {
            return branch.clone();
        }
        crate::dolt::remotes::RepoState::load(&self.data_dir.join(database))
            .ok()
            .and_then(|state| state.head.strip_prefix("refs/heads/").map(str::to_string))
            .unwrap_or_else(|| DEFAULT_BRANCH.to_string())
    }

    /// sync_identity sets the session user and the current role from the parameters that SET SESSION AUTHORIZATION
    /// and SET ROLE change, which transactions can undo.
    pub fn sync_identity(&mut self) {
        self.user = self.settings.raw("session_authorization").unwrap_or_else(|| self.authenticated.clone());
        self.role = match self.settings.raw("role") {
            Some(role) if role != "none" => role,
            _ => self.user.clone(),
        };
    }

    /// database_names returns the names of the databases in the data directory, in name order.
    pub fn database_names(&self) -> Vec<String> {
        let mut names: Vec<String> = std::fs::read_dir(&self.data_dir)
            .map(|entries| {
                entries
                    .filter_map(|e| e.ok())
                    .filter(|e| e.path().join(".dolt").is_dir())
                    .filter_map(|e| e.file_name().into_string().ok())
                    .collect()
            })
            .unwrap_or_default();
        names.sort();
        names
    }

    /// setting_on reports whether a boolean setting is on.
    pub fn setting_on(&self, name: &str) -> bool {
        self.settings.get(name).is_some_and(|v| matches!(v.to_ascii_lowercase().as_str(), "on" | "true" | "1" | "yes"))
    }

    /// search_path returns the schemas that unqualified names resolve in.
    pub fn search_path(&self) -> Vec<String> {
        let path = self.settings.get("search_path").unwrap_or_default();
        let path = match &self.view_schema {
            Some(schema) => format!("{},{path}", crate::engine::quote_identifier(schema)),
            None => path,
        };
        let mut seen = std::collections::HashSet::new();
        path.split(',')
            .map(|s| s.trim())
            .filter(|s| !s.is_empty())
            .map(|s| {
                let s = if s.starts_with('"') && s.ends_with('"') && s.len() > 1 {
                    s[1..s.len() - 1].replace("\"\"", "\"")
                } else {
                    s.to_ascii_lowercase()
                };
                if s == "$user" { self.role.clone() } else { s }
            })
            .filter(|s| self.can_use_schema(s) && seen.insert(s.clone()))
            .collect()
    }

    /// can_use_schema reports whether the current role may use a schema, which unqualified names skip otherwise.
    fn can_use_schema(&self, schema: &str) -> bool {
        if matches!(schema, "pg_catalog" | "information_schema" | "public") {
            return true;
        }
        let Ok(auth) = self.auth.lock() else { return true };
        let Some(role) = auth.role(&self.role) else { return true };
        let object = crate::auth::Object::Schema(schema.to_string());
        role.superuser || auth.holds(role.id, &object, "U") || auth.owner(&object) == Some(role.id)
    }

    /// install_format installs the session's DateStyle, IntervalStyle, time zone, and bytea_output for printing values.
    pub fn install_format(&self) {
        let get = |name: &str| self.settings.get(name).unwrap_or_default();
        crate::datetime::install_format(crate::datetime::Format::from_settings(
            &get("DateStyle"),
            &get("IntervalStyle"),
            &get("TimeZone"),
        ));
        crate::binary::install_output(&get("bytea_output"));
        crate::xml::install_options(&get("xmloption"), &get("xmlbinary"), &get("client_encoding"));
    }

    /// notice records a notice for the client.
    pub fn notice(&mut self, notice: PgError) {
        self.notices.push(notice);
    }
}

/// REPORTED_PARAMETERS are the parameters whose values the server reports to the client with ParameterStatus.
const REPORTED_PARAMETERS: [&str; 13] = [
    "application_name",
    "client_encoding",
    "DateStyle",
    "default_transaction_read_only",
    "in_hot_standby",
    "integer_datetimes",
    "IntervalStyle",
    "is_superuser",
    "server_encoding",
    "server_version",
    "session_authorization",
    "standard_conforming_strings",
    "TimeZone",
];

/// quote_identifier quotes an identifier when Postgres would.
pub fn quote_identifier(name: &str) -> String {
    let simple = name.chars().next().is_some_and(|c| c.is_ascii_lowercase() || c == '_')
        && name.chars().all(|c| c.is_ascii_lowercase() || c.is_ascii_digit() || c == '_' || c == '$');
    let keyword =
        pg_query::scan(name).ok().and_then(|scan| scan.tokens.first().map(|t| t.keyword_kind > 1)).unwrap_or(false);
    if simple && !keyword { name.to_string() } else { format!("\"{}\"", name.replace('"', "\"\"")) }
}

/// transaction_statement returns a transaction statement.
fn transaction_statement(statement: &Statement) -> Option<&pg_query::protobuf::TransactionStmt> {
    match statement {
        Statement::Postgres { node: NodeEnum::TransactionStmt(t), .. } => Some(t),
        _ => None,
    }
}

/// transaction_kind returns the kind of a transaction statement.
fn transaction_kind(statement: &Statement) -> Option<TransactionStmtKind> {
    match statement {
        Statement::Postgres { node: NodeEnum::TransactionStmt(t), .. } => TransactionStmtKind::try_from(t.kind).ok(),
        _ => None,
    }
}

impl Session {
    /// current_database returns the session's database as it was named, with any branch.
    pub fn current_database(&self) -> &str {
        &self.state.display
    }

    /// tx_status returns the transaction status that ReadyForQuery reports: idle, in a transaction, or failed.
    pub fn tx_status(&self) -> u8 {
        match (self.state.explicit, self.failed) {
            (true, true) => b'E',
            (true, false) => b'T',
            _ => b'I',
        }
    }

    /// parameter_changes returns the reported parameters whose values changed since the client last heard them, all
    /// of them the first time.
    pub fn parameter_changes(&mut self) -> Vec<(String, String)> {
        let mut changes = Vec::new();
        for name in REPORTED_PARAMETERS {
            let value = match name {
                "session_authorization" => self.state.user.clone(),
                "server_version" => crate::SERVER_VERSION.to_string(),
                "is_superuser" => "on".to_string(),
                _ => self.state.settings.show(name).unwrap_or_default(),
            };
            if self.reported.get(name) != Some(&value) {
                self.reported.insert(name.to_string(), value.clone());
                changes.push((name.to_string(), value));
            }
        }
        changes
    }

    /// take_notices returns the notices raised since the last call.
    pub fn take_notices(&mut self) -> Vec<PgError> {
        std::mem::take(&mut self.state.notices)
    }

    /// switch makes a database, or a branch of one written as `database/branch`, the session's database.
    fn switch(&mut self, target: &str) -> Result<()> {
        let not_found = || PgError::new(code::INVALID_CATALOG_NAME, format!("database \"{target}\" does not exist"));
        let (database, branch) = match target.split_once('/') {
            Some((database, branch)) => (database, branch.to_string()),
            None => (target, String::new()),
        };
        if !self.state.engine.database_exists(database) {
            return Err(not_found());
        }
        let branch = match branch.is_empty() {
            true => self.state.checked_out_branch(database),
            false => branch,
        };
        let branch = branch.as_str();
        let handle = self.state.engine.database(database)?;
        let mut db = lock(&handle)?;
        if db.head(&doltdb::create::branch_ref(branch))?.is_none()
            && crate::dolt::history::resolve(&mut db, store::Hash::default(), branch).is_err()
        {
            return Err(not_found());
        }
        drop(db);
        if !self.state.display.is_empty() && !self.state.display.contains('/') {
            self.state.checked_out.insert(self.state.database.clone(), self.state.branch.clone());
        }
        self.state.database = database.to_string();
        self.state.branch = branch.to_string();
        self.state.display = target.to_string();
        Ok(())
    }

    /// report_activity records the statement that the session runs, or that it went idle after its last statement
    /// for None, keeping the statement's text as the session's source.
    fn report_activity(&mut self, query: Option<&str>) {
        if let Some(query) = query {
            self.state.source = query.to_string();
        }
        let state = &self.state;
        state.engine.update_activity(state.id, |activity| {
            activity.database.clone_from(&state.database);
            activity.user.clone_from(&state.user);
            activity.host.clone_from(&state.host);
            if let Some(query) = query {
                activity.query = query.to_string();
            }
            activity.started = query.map(|_| crate::datetime::clock());
            activity.in_transaction = state.explicit;
        });
    }

    /// execute runs the statements of a simple query, stopping at the first error, and returns what each produced.
    /// The statements run in one implicit transaction unless they manage their own.
    pub fn execute(&mut self, query: &str) -> (Results, Option<PgError>) {
        self.report_activity(Some(query));
        let result = match parse::parse(query) {
            Ok(statements) if statements.is_empty() => (vec![(Vec::new(), Outcome::Empty)], None),
            Ok(statements) => self.run_batch(statements, Vec::new()),
            Err(err) => (Vec::new(), Some(self.fail(err))),
        };
        self.report_activity(None);
        result
    }

    /// run_batch runs the statements of a simple query after the outcomes of the ones before them, pausing at a COPY
    /// FROM STDIN until the client sends its data.
    fn run_batch(&mut self, statements: Vec<Statement>, mut outcomes: Results) -> (Results, Option<PgError>) {
        let mut statements = statements.into_iter();
        while let Some(statement) = statements.next() {
            match self.run(&statement, &[]) {
                Ok(outcome @ Outcome::CopyIn { .. }) => {
                    self.pending = Some(statements.collect());
                    outcomes.push((self.take_notices(), outcome));
                    return (outcomes, None);
                }
                Ok(outcome) => outcomes.push((self.take_notices(), outcome)),
                Err(err) => return (outcomes, Some(self.fail(err))),
            }
        }
        if let Err(err) = self.end_implicit() {
            return (outcomes, Some(self.fail(err)));
        }
        (outcomes, None)
    }

    /// copy_data finishes a COPY FROM STDIN with the data the client sent, then runs the rest of its query.
    pub fn copy_data(&mut self, data: &[u8]) -> (Results, Option<PgError>) {
        let Some(copy) = self.state.pending_copy.take() else { return (Vec::new(), None) };
        let mut parameters = Vec::new();
        match self.with_ctx(&mut parameters, &[], |ctx| ctx.copy_rows(&copy, data)) {
            Ok(outcome) => {
                let outcome = (self.take_notices(), outcome);
                match self.pending.take() {
                    Some(pending) => self.run_batch(pending, vec![outcome]),
                    None => (vec![outcome], None),
                }
            }
            Err(err) => {
                self.pending = None;
                (Vec::new(), Some(self.fail(err)))
            }
        }
    }

    /// abort_copy ends a COPY FROM STDIN with an error at a line of its data, along with the rest of its query.
    pub fn abort_copy(&mut self, err: PgError, line: usize) -> PgError {
        let table = self.state.pending_copy.take().map(|copy| copy.table.name).unwrap_or_default();
        self.pending = None;
        self.fail(crate::copy::with_context(err, crate::copy::context(&table, line)))
    }

    /// prepare parses a query of at most one statement and describes its parameters and results.
    pub fn prepare(&mut self, query: &str, parameter_types: &[u32]) -> Result<Prepared> {
        self.state.source = query.to_string();
        let mut statements = parse::parse(query)?;
        if statements.len() > 1 {
            return Err(PgError::new(code::SYNTAX_ERROR, "cannot insert multiple commands into a prepared statement"));
        }
        let statement = statements.pop();
        let mut parameters = parameter_types.to_vec();
        let mut columns = None;
        if let Some(Statement::Postgres { node: NodeEnum::VariableShowStmt(show), .. }) = &statement {
            if show.name != "all" {
                self.state.settings.show(&show.name)?;
            }
            columns = Some(show_columns(&show.name));
        } else if let Some(Statement::Describe { .. }) = &statement {
            columns = Some(crate::listing::describe_columns());
        } else if let Some(Statement::ShowCreateTable { .. }) = &statement {
            columns = Some(crate::listing::show_create_columns());
        } else if let Some(Statement::Listing { kind, from }) = &statement {
            columns = Some(self.with_ctx(&mut parameters, &[], |ctx| ctx.plan_listing(kind, from))?.columns);
        } else if let Some(Statement::Postgres { node, .. }) = &statement
            && describable(node)
        {
            columns = self.with_ctx(&mut parameters, &[], |ctx| ctx.describe(node))?;
        }
        for parameter in &mut parameters {
            if *parameter == 0 {
                *parameter = crate::oid::TEXT;
            }
        }
        Ok(Prepared { query: query.to_string(), statement, parameter_types: parameters, columns })
    }

    /// execute_prepared runs a prepared statement with the parameter values, in the implicit transaction that lasts
    /// until the next Sync.
    pub fn execute_prepared(&mut self, prepared: &Prepared, parameters: &[Value]) -> Result<Outcome> {
        let Some(statement) = &prepared.statement else { return Ok(Outcome::Empty) };
        self.report_activity(Some(&prepared.query));
        let result = match self.reg_parameters(&prepared.parameter_types, parameters) {
            Ok(parameters) => self.run(statement, &parameters).map_err(|err| self.fail(err)),
            Err(err) => Err(self.fail(err)),
        };
        self.report_activity(None);
        result
    }

    /// reg_parameters looks up the objects that the text of reg-typed parameters names.
    fn reg_parameters(&mut self, types: &[u32], parameters: &[Value]) -> Result<Vec<Value>> {
        if !types.iter().any(|&t| crate::cast::is_reg_type(t)) {
            return Ok(parameters.to_vec());
        }
        let mut scratch = Vec::new();
        self.with_ctx(&mut scratch, &[], |ctx| {
            parameters
                .iter()
                .zip(types)
                .map(|(value, &ty)| match value {
                    Value::Text(_) if crate::cast::is_reg_type(ty) => ctx.reg_value(value.clone(), ty),
                    other => Ok(other.clone()),
                })
                .collect()
        })
    }

    /// sync commits the implicit transaction of the extended protocol's messages since the last Sync.
    pub fn sync(&mut self) -> Result<()> {
        self.end_implicit().map_err(|err| self.fail(err))
    }

    /// abort ends an implicit transaction, or marks an explicit one failed, after an error in a protocol message.
    pub fn abort(&mut self, err: PgError) -> PgError {
        self.fail(err)
    }

    /// fail ends an implicit transaction, or marks an explicit one failed, after an error.
    fn fail(&mut self, err: PgError) -> PgError {
        let err = self.fail_transaction(err);
        self.state.sync_identity();
        err
    }

    /// fail_transaction ends an implicit transaction, or marks an explicit one failed, after an error.
    fn fail_transaction(&mut self, err: PgError) -> PgError {
        if self.state.explicit {
            self.failed = true;
        } else {
            self.txns.clear();
            self.state.end_transaction(false);
        }
        err
    }

    /// end_implicit commits the transaction when it is implicit.
    fn end_implicit(&mut self) -> Result<()> {
        if self.state.explicit {
            return Ok(());
        }
        self.check_deferred()?;
        self.commit()
    }

    /// check_deferred runs the checks that deferred constraints owe before the transaction commits, rolling it back
    /// when one fails.
    fn check_deferred(&mut self) -> Result<()> {
        if self.state.deferred.pending.is_empty() {
            return Ok(());
        }
        let mut parameters = Vec::new();
        if let Err(err) = self.with_ctx(&mut parameters, &[], |ctx| ctx.run_deferred(true)) {
            self.txns.clear();
            self.state.end_transaction(false);
            return Err(err);
        }
        Ok(())
    }

    /// commit commits and ends the open transaction, refusing a working set with conflicts or constraint violations
    /// as Dolt does unless the session allows them.
    fn commit(&mut self) -> Result<()> {
        self.state.end_transaction(true);
        let allow_conflicts = self.state.setting_on("dolt_allow_commit_conflicts");
        let force = self.state.setting_on("dolt_force_transaction_commit");
        let autocommit = !self.state.explicit;
        for txn in std::mem::take(&mut self.txns) {
            let handle = txn.handle.clone();
            let mut db = lock(&handle)?;
            if txn.changed() {
                let schema_conflicts = txn.merge.as_ref().is_some_and(|m| !m.unmergable_tables.is_empty());
                crate::dolt::conflicts::commit_check(
                    &mut db,
                    &txn.root,
                    schema_conflicts,
                    allow_conflicts,
                    force,
                    autocommit,
                )?;
            }
            txn.commit(&mut db, &self.state.user, &self.state.host)?;
        }
        Ok(())
    }

    /// with_ctx runs a function with the planning context of the open transaction, beginning one when needed.
    fn with_ctx<T>(
        &mut self,
        parameters: &mut Vec<u32>,
        params: &[Value],
        f: impl FnOnce(&mut Ctx<'_>) -> Result<T>,
    ) -> Result<T> {
        self.state.sync_identity();
        let (database, branch) = (&self.state.database, &self.state.branch);
        let index = match self.txns.iter().position(|t| t.database == *database && t.branch == *branch) {
            Some(index) => index,
            None => {
                let (handle, tracker) = self.state.engine.open_database(database)?;
                let mut txn = Txn::begin(handle, tracker, database, branch)?;
                if let Some(first) = self.txns.first() {
                    txn.started = first.started;
                }
                self.txns.push(txn);
                self.txns.len() - 1
            }
        };
        let mut txn = self.txns.remove(index);
        crate::datetime::install_now(txn.started);
        self.state.install_format();
        let handle = txn.handle.clone();
        let mut db = lock(&handle)?;
        let mut ctx = Ctx {
            db: &mut db,
            txn: &mut txn,
            branches: &mut self.txns,
            session: &mut self.state,
            parameters,
            params,
            outer: Vec::new(),
            subquery_value: Value::Null,
            ctes: Vec::new(),
            work_tables: std::collections::HashMap::new(),
            named_params: None,
        };
        let result = (|| {
            ctx.install_types()?;
            ctx.install_casts()?;
            ctx.install_aggregates()?;
            f(&mut ctx)
        })();
        self.txns.insert(index, txn);
        result
    }

    /// run runs one statement with the parameter values.
    fn run(&mut self, statement: &Statement, params: &[Value]) -> Result<Outcome> {
        if self.state.engine.ended(self.state.id) {
            return Err(PgError::fatal(
                code::INTERNAL_ERROR,
                "this connection was established when this server performed an online garbage collection. this \
                 connection can no longer be used. please reconnect.",
            ));
        }
        let result = self.run_statement(statement, params);
        for warning in crate::xml::take_warnings() {
            self.state.notices.push(PgError { severity: "WARNING", ..PgError::new("01000", warning) });
        }
        self.state.sync_identity();
        result
    }

    /// run_statement runs one statement with the parameter values.
    fn run_statement(&mut self, statement: &Statement, params: &[Value]) -> Result<Outcome> {
        let kind = transaction_kind(statement);
        if self.failed {
            return match kind {
                Some(TransactionStmtKind::TransStmtCommit | TransactionStmtKind::TransStmtRollback) => {
                    self.txns.clear();
                    self.savepoints.clear();
                    self.state.explicit = false;
                    self.failed = false;
                    self.state.end_transaction(false);
                    Ok(Outcome::command("ROLLBACK"))
                }
                Some(TransactionStmtKind::TransStmtRollbackTo) => {
                    let name = transaction_statement(statement).map(|t| t.savepoint_name.clone()).unwrap_or_default();
                    self.rollback_to(&name)
                }
                _ => Err(PgError::new(
                    code::IN_FAILED_SQL_TRANSACTION,
                    "current transaction is aborted, commands ignored until end of transaction block",
                )),
            };
        }
        if let Some(kind) = kind {
            let name = transaction_statement(statement).map(|t| t.savepoint_name.clone()).unwrap_or_default();
            let begun = !self.state.explicit;
            let outcome = self.transaction(kind, &name)?;
            if matches!(kind, TransactionStmtKind::TransStmtBegin | TransactionStmtKind::TransStmtStart) {
                self.begin_modes(transaction_statement(statement), begun)?;
            }
            return Ok(outcome);
        }
        match statement {
            Statement::Use(target) => {
                self.commit()?;
                self.switch(target)?;
                Ok(Outcome::command("SET"))
            }
            Statement::SetExpression { name, local, value } => {
                let mut parameters = Vec::new();
                let value = self.with_ctx(&mut parameters, params, |ctx| ctx.constant_text(value))?;
                self.state.settings.set(name, Some(&value), *local, self.state.explicit)?;
                Ok(Outcome::command("SET"))
            }
            Statement::Describe { relation, extras } => {
                let mut parameters = Vec::new();
                self.with_ctx(&mut parameters, params, |ctx| ctx.describe_table(relation, extras))
            }
            Statement::ShowCreateTable { relation, extras } => {
                let mut parameters = Vec::new();
                self.with_ctx(&mut parameters, params, |ctx| ctx.show_create_table(relation, extras))
            }
            Statement::Listing { kind, from } => {
                let mut parameters = Vec::new();
                self.with_ctx(&mut parameters, params, |ctx| ctx.list_objects(kind, from))
            }
            Statement::Postgres { node, extras } => self.postgres(node, extras, params),
        }
    }

    /// savepoint_index returns the position of the newest savepoint of a name, failing when there is none.
    fn savepoint_index(&self, name: &str) -> Result<usize> {
        self.savepoints.iter().rposition(|(n, _, _)| n == name).ok_or_else(|| {
            PgError::new(code::INVALID_SAVEPOINT_SPECIFICATION, format!("savepoint \"{name}\" does not exist"))
        })
    }

    /// rollback_to runs ROLLBACK TO SAVEPOINT, which restores the transaction and the settings as the savepoint saw
    /// them and keeps the savepoint.
    fn rollback_to(&mut self, name: &str) -> Result<Outcome> {
        let index = self.savepoint_index(name)?;
        let (_, txns, settings) = self.savepoints[index].clone();
        self.savepoints.truncate(index + 1);
        self.txns = txns;
        self.state.settings = settings;
        self.failed = false;
        Ok(Outcome::command("ROLLBACK"))
    }

    /// begin_modes applies the transaction modes of a BEGIN, which a new transaction otherwise takes from
    /// default_transaction_read_only, as Postgres does even for a BEGIN inside a transaction block.
    fn begin_modes(&mut self, stmt: Option<&pg_query::protobuf::TransactionStmt>, begun: bool) -> Result<()> {
        if begun {
            let read_only = self.state.settings.get("default_transaction_read_only").unwrap_or_default();
            self.state.settings.set("transaction_read_only", Some(&read_only), true, true)?;
        }
        for option in stmt.map(|s| s.options.as_slice()).unwrap_or_default() {
            let Some(NodeEnum::DefElem(def)) = option.node.as_ref() else { continue };
            let value = match def.arg.as_deref().and_then(|a| a.node.as_ref()) {
                Some(NodeEnum::AConst(c)) => match &c.val {
                    Some(Val::Ival(i)) => if i.ival == 0 { "off" } else { "on" }.to_string(),
                    Some(Val::Sval(s)) => s.sval.clone(),
                    _ => continue,
                },
                _ => continue,
            };
            self.state.settings.set(&def.defname, Some(&value), true, true)?;
        }
        Ok(())
    }

    /// transaction runs BEGIN, COMMIT, ROLLBACK, SAVEPOINT, RELEASE, or ROLLBACK TO.
    fn transaction(&mut self, kind: TransactionStmtKind, name: &str) -> Result<Outcome> {
        let outside = |statement: &str| {
            PgError::new(code::NO_ACTIVE_SQL_TRANSACTION, format!("{statement} can only be used in transaction blocks"))
        };
        match kind {
            TransactionStmtKind::TransStmtSavepoint => {
                if !self.state.explicit {
                    return Err(outside("SAVEPOINT"));
                }
                self.savepoints.push((name.to_string(), self.txns.clone(), self.state.settings.clone()));
                return Ok(Outcome::command("SAVEPOINT"));
            }
            TransactionStmtKind::TransStmtRelease => {
                if !self.state.explicit {
                    return Err(outside("RELEASE SAVEPOINT"));
                }
                let index = self.savepoint_index(name)?;
                self.savepoints.truncate(index);
                return Ok(Outcome::command("RELEASE"));
            }
            TransactionStmtKind::TransStmtRollbackTo => {
                if !self.state.explicit {
                    return Err(outside("ROLLBACK TO SAVEPOINT"));
                }
                return self.rollback_to(name);
            }
            TransactionStmtKind::TransStmtCommit | TransactionStmtKind::TransStmtRollback => self.savepoints.clear(),
            _ => {}
        }
        match kind {
            TransactionStmtKind::TransStmtBegin | TransactionStmtKind::TransStmtStart => {
                if self.state.explicit {
                    self.state.notices.push(PgError {
                        severity: "WARNING",
                        ..PgError::new(code::ACTIVE_SQL_TRANSACTION, "there is already a transaction in progress")
                    });
                }
                self.state.explicit = true;
                let start = kind == TransactionStmtKind::TransStmtStart;
                Ok(Outcome::command(if start { "START TRANSACTION" } else { "BEGIN" }))
            }
            TransactionStmtKind::TransStmtCommit => {
                if !self.state.explicit {
                    self.state.notices.push(PgError {
                        severity: "WARNING",
                        ..PgError::new(code::NO_ACTIVE_SQL_TRANSACTION, "there is no transaction in progress")
                    });
                }
                self.state.explicit = false;
                self.check_deferred()?;
                self.commit()?;
                Ok(Outcome::command("COMMIT"))
            }
            TransactionStmtKind::TransStmtRollback => {
                if !self.state.explicit {
                    self.state.notices.push(PgError {
                        severity: "WARNING",
                        ..PgError::new(code::NO_ACTIVE_SQL_TRANSACTION, "there is no transaction in progress")
                    });
                }
                self.state.explicit = false;
                self.txns.clear();
                self.state.end_transaction(false);
                Ok(Outcome::command("ROLLBACK"))
            }
            other => Err(PgError::unsupported(format!("the transaction statement {other:?}"))),
        }
    }

    /// create_database runs CREATE DATABASE, which only Doltgres allows with IF NOT EXISTS.
    fn create_database(&mut self, name: &str, if_not_exists: bool) -> Result<Outcome> {
        if self.state.explicit {
            return Err(PgError::new(
                code::ACTIVE_SQL_TRANSACTION,
                "CREATE DATABASE cannot run inside a transaction block",
            ));
        }
        if name.contains(['/', '\\']) || name.is_empty() {
            return Err(PgError {
                detail: Some("Database names cannot be empty or contain \"/\" or \"\\\", which name branches.".into()),
                ..PgError::new(code::INVALID_NAME, format!("invalid database name \"{name}\""))
            });
        }
        if self.state.engine.database_exists(name) {
            if if_not_exists {
                return Ok(Outcome::command("CREATE DATABASE"));
            }
            return Err(PgError::new(code::DUPLICATE_DATABASE, format!("database \"{name}\" already exists")));
        }
        self.state.engine.create_database(name, &self.state.user, &self.state.host)?;
        Ok(Outcome::command("CREATE DATABASE"))
    }

    /// drop_database runs DROP DATABASE, which only superusers may run since databases record no owner, refusing
    /// the session's own database as Postgres does.
    fn drop_database(&mut self, name: &str, missing_ok: bool, superuser: bool) -> Result<Outcome> {
        if self.state.explicit {
            return Err(PgError::new(
                code::ACTIVE_SQL_TRANSACTION,
                "DROP DATABASE cannot run inside a transaction block",
            ));
        }
        if !self.state.engine.database_exists(name) {
            if missing_ok {
                self.state
                    .notices
                    .push(PgError::notice("00000", format!("database \"{name}\" does not exist, skipping")));
                return Ok(Outcome::command("DROP DATABASE"));
            }
            return Err(PgError::new(code::INVALID_CATALOG_NAME, format!("database \"{name}\" does not exist")));
        }
        if !superuser {
            return Err(PgError::new(code::INSUFFICIENT_PRIVILEGE, format!("must be owner of database {name}")));
        }
        if self.state.database == name {
            return Err(PgError::new(code::OBJECT_IN_USE, "cannot drop the currently open database"));
        }
        self.state.engine.drop_database(name)?;
        Ok(Outcome::command("DROP DATABASE"))
    }

    /// postgres runs a statement of Postgres' grammar.
    fn postgres(&mut self, node: &NodeEnum, extras: &Extras, params: &[Value]) -> Result<Outcome> {
        self.state.as_of = extras.as_of.clone();
        match node {
            NodeEnum::DropdbStmt(drop) => {
                let mut parameters = Vec::new();
                let superuser = self.with_ctx(&mut parameters, params, |ctx| Ok(ctx.current_role()?.superuser))?;
                return self.drop_database(&drop.dbname, drop.missing_ok, superuser);
            }
            NodeEnum::CreatedbStmt(create) => {
                let mut parameters = Vec::new();
                self.with_ctx(&mut parameters, params, |ctx| ctx.require_create_db())?;
                return self.create_database(&create.dbname, extras.if_not_exists);
            }
            NodeEnum::CreateRoleStmt(create) => {
                let mut parameters = Vec::new();
                return self.with_ctx(&mut parameters, params, |ctx| ctx.create_role(create, extras.if_not_exists));
            }
            NodeEnum::CreateFunctionStmt(create) => {
                let mut parameters = Vec::new();
                return self.with_ctx(&mut parameters, params, |ctx| ctx.create_function(create, &extras.text));
            }
            NodeEnum::DefineStmt(define)
                if define.kind == pg_query::protobuf::ObjectType::ObjectType as i32 && define.definition.is_empty() =>
            {
                let mut parameters = Vec::new();
                return self.with_ctx(&mut parameters, params, |ctx| ctx.create_shell(define));
            }
            NodeEnum::DefineStmt(define) if define.kind == pg_query::protobuf::ObjectType::ObjectAggregate as i32 => {
                let mut parameters = Vec::new();
                return self.with_ctx(&mut parameters, params, |ctx| ctx.create_aggregate(define));
            }
            NodeEnum::DefineStmt(define) if define.kind == pg_query::protobuf::ObjectType::ObjectOperator as i32 => {
                let mut parameters = Vec::new();
                return self.with_ctx(&mut parameters, params, |ctx| ctx.create_operator(define));
            }
            NodeEnum::CreateCastStmt(create) => {
                let mut parameters = Vec::new();
                return self.with_ctx(&mut parameters, params, |ctx| ctx.create_cast(create));
            }
            NodeEnum::DoStmt(stmt) => {
                let mut parameters = Vec::new();
                return self.with_ctx(&mut parameters, params, |ctx| ctx.do_block(stmt, &extras.text));
            }
            NodeEnum::CreateTrigStmt(create) => {
                let mut parameters = Vec::new();
                return self.with_ctx(&mut parameters, params, |ctx| ctx.create_trigger(create, &extras.text));
            }
            NodeEnum::ViewStmt(stmt) => {
                let mut parameters = Vec::new();
                return self.with_ctx(&mut parameters, params, |ctx| ctx.create_view(stmt, &extras.text));
            }
            NodeEnum::VariableSetStmt(set) if matches!(set.name.as_str(), "role" | "session_authorization") => {
                return self.set_role(set);
            }
            NodeEnum::VariableSetStmt(set) => return self.set(set),
            NodeEnum::VariableShowStmt(show) => return self.show(&show.name),
            NodeEnum::DeallocateStmt(stmt) if stmt.isall => return Ok(Outcome::command("DEALLOCATE ALL")),
            NodeEnum::DiscardStmt(discard) => {
                let mode = pg_query::protobuf::DiscardMode::try_from(discard.target);
                if mode != Ok(pg_query::protobuf::DiscardMode::DiscardAll) {
                    let tag = match mode {
                        Ok(pg_query::protobuf::DiscardMode::DiscardPlans) => "DISCARD PLANS",
                        Ok(pg_query::protobuf::DiscardMode::DiscardSequences) => "DISCARD SEQUENCES",
                        _ => "DISCARD TEMP",
                    };
                    if tag == "DISCARD SEQUENCES" {
                        self.state.sequence_values.clear();
                        self.state.last_sequence = None;
                    }
                    return Ok(Outcome::command(tag));
                }
                if self.state.explicit {
                    return Err(PgError::new(
                        code::ACTIVE_SQL_TRANSACTION,
                        "DISCARD ALL cannot run inside a transaction block",
                    ));
                }
                self.state.settings.reset_all(false);
                let user = self.state.authenticated.clone();
                self.state.settings.set_raw("session_authorization", Some(user), false, false);
                self.state.settings.set_raw("role", None, false, false);
                self.state.sequence_values.clear();
                self.state.last_sequence = None;
                return Ok(Outcome::command("DISCARD ALL"));
            }
            _ => {}
        }
        self.wait_for_advisory_locks(node);
        let mut parameters = Vec::new();
        self.with_ctx(&mut parameters, params, |ctx| ctx.run(node))
    }

    /// wait_for_advisory_locks waits until the advisory locks that a statement takes with constant keys are free,
    /// before the statement takes its database, since a statement cannot wait while it holds its database.
    fn wait_for_advisory_locks(&self, node: &NodeEnum) {
        for (node, ..) in node.nodes() {
            let pg_query::NodeRef::FuncCall(call) = node else { continue };
            let exclusive = match call.funcname.iter().filter_map(crate::expr::node_name).next_back() {
                Some("pg_advisory_lock" | "pg_advisory_xact_lock") => true,
                Some("pg_advisory_lock_shared" | "pg_advisory_xact_lock_shared") => false,
                _ => continue,
            };
            let ints: Option<Vec<i64>> = call
                .args
                .iter()
                .map(|arg| match arg.node.as_ref() {
                    Some(NodeEnum::AConst(c)) => match &c.val {
                        Some(Val::Ival(i)) => Some(i.ival as i64),
                        Some(Val::Fval(f)) => f.fval.parse().ok(),
                        _ => None,
                    },
                    _ => None,
                })
                .collect();
            if let Some(ints) = ints {
                let key = crate::advisory::Key::new(&self.state.database, &ints);
                self.state.advisory.wait(self.state.id, &key, exclusive);
            }
        }
    }
}

impl Session {
    /// set_role runs SET ROLE, RESET ROLE, SET SESSION AUTHORIZATION, and RESET SESSION AUTHORIZATION.
    fn set_role(&mut self, set: &VariableSetStmt) -> Result<Outcome> {
        let kind = VariableSetKind::try_from(set.kind).unwrap_or(VariableSetKind::Undefined);
        let name = match set.args.first().and_then(|a| a.node.as_ref()) {
            Some(NodeEnum::AConst(c)) => match &c.val {
                Some(Val::Sval(s)) => Some(s.sval.clone()),
                _ => None,
            },
            _ => None,
        };
        let tag = if kind == VariableSetKind::VarReset { "RESET" } else { "SET" };
        let mut parameters = Vec::new();
        let explicit = self.state.explicit;
        self.warn_set_local(set);
        if set.name == "session_authorization" {
            let user = name.filter(|n| kind == VariableSetKind::VarSetValue && n != "default");
            let user = user.unwrap_or_else(|| self.state.authenticated.clone());
            self.with_ctx(&mut parameters, &[], |ctx| ctx.set_session_authorization(&user))?;
            self.state.settings.set_raw("session_authorization", Some(user), set.is_local, explicit);
            self.state.settings.set_raw("role", None, set.is_local, explicit);
        } else {
            let role = name.filter(|n| kind == VariableSetKind::VarSetValue && n != "none");
            self.with_ctx(&mut parameters, &[], |ctx| ctx.set_role(role.as_deref()))?;
            self.state.settings.set_raw("role", role, set.is_local, explicit);
        }
        self.state.sync_identity();
        Ok(Outcome::command(tag))
    }

    /// warn_set_local warns as Postgres does about SET LOCAL outside a transaction block, where it has no effect.
    fn warn_set_local(&mut self, set: &VariableSetStmt) {
        if set.is_local && !self.state.explicit {
            self.state.notices.push(PgError {
                severity: "WARNING",
                ..PgError::new(code::NO_ACTIVE_SQL_TRANSACTION, "SET LOCAL can only be used in transaction blocks")
            });
        }
    }

    /// set runs SET and RESET.
    fn set(&mut self, set: &VariableSetStmt) -> Result<Outcome> {
        let kind = VariableSetKind::try_from(set.kind).unwrap_or(VariableSetKind::Undefined);
        let in_transaction = self.state.explicit;
        let reset =
            matches!(kind, VariableSetKind::VarReset | VariableSetKind::VarResetAll | VariableSetKind::VarSetDefault);
        let tag =
            if matches!(kind, VariableSetKind::VarReset | VariableSetKind::VarResetAll) { "RESET" } else { "SET" };
        let transactional =
            matches!(set.name.as_str(), "transaction_isolation" | "transaction_read_only" | "transaction_deferrable");
        if reset && set.name == "transaction_isolation" && !in_transaction {
            self.state.notices.push(PgError {
                severity: "WARNING",
                ..PgError::new(
                    code::NO_ACTIVE_SQL_TRANSACTION,
                    "RESET TRANSACTION can only be used in transaction blocks",
                )
            });
        }
        self.warn_set_local(set);
        let local = set.is_local || transactional;
        match kind {
            VariableSetKind::VarSetValue => {
                let value = set_value(&set.name, &set.args)?;
                self.state.settings.set(&set.name, Some(&value), local, in_transaction)?;
            }
            VariableSetKind::VarSetDefault | VariableSetKind::VarReset => {
                self.state.settings.set(&set.name, None, local, in_transaction)?;
            }
            VariableSetKind::VarResetAll => self.state.settings.reset_all(in_transaction),
            VariableSetKind::VarSetMulti => {}
            _ => return Err(PgError::unsupported("this SET")),
        }
        Ok(Outcome::command(tag))
    }

    /// show runs SHOW.
    fn show(&mut self, name: &str) -> Result<Outcome> {
        if name == "all" {
            let mut rows = Vec::new();
            for definition in crate::settings::all_settings() {
                let value = self.state.settings.show(&definition.name)?;
                rows.push(vec![
                    Value::Text(definition.name.clone()),
                    Value::Text(value),
                    Value::Text(definition.description.clone()),
                ]);
            }
            let tag = format!("SHOW {}", rows.len());
            return Ok(Outcome::Rows { columns: show_columns(name), rows, tag });
        }
        let value = match name {
            "session_authorization" => self.state.user.clone(),
            "server_version" => crate::SERVER_VERSION.to_string(),
            _ => self.state.settings.show(name)?,
        };
        Ok(Outcome::Rows { columns: show_columns(name), rows: vec![vec![Value::Text(value)]], tag: "SHOW".into() })
    }
}

/// show_columns returns the result columns of SHOW for a parameter, or for every parameter.
fn show_columns(name: &str) -> Vec<Column> {
    let text = crate::expr::typ(crate::oid::TEXT);
    if name == "all" {
        return vec![column("name".into(), text), column("setting".into(), text), column("description".into(), text)];
    }
    vec![column(setting(name).map_or(name.to_string(), |s| s.name.clone()), text)]
}

/// LIST_QUOTE_SETTINGS are the list settings whose items SET quotes as identifiers.
const LIST_QUOTE_SETTINGS: [&str; 5] = [
    "search_path",
    "temp_tablespaces",
    "session_preload_libraries",
    "shared_preload_libraries",
    "local_preload_libraries",
];

/// set_value flattens SET's arguments into the text of the value, as Postgres' flatten_set_variable_args does.
fn set_value(name: &str, args: &[Node]) -> Result<String> {
    let quote = LIST_QUOTE_SETTINGS.contains(&name);
    let mut parts = Vec::new();
    for arg in args {
        let text = match arg.node.as_ref() {
            Some(NodeEnum::AConst(c)) => match &c.val {
                Some(Val::Ival(i)) => i.ival.to_string(),
                Some(Val::Fval(f)) => f.fval.clone(),
                Some(Val::Sval(s)) if quote => quote_identifier(&s.sval),
                Some(Val::Sval(s)) => s.sval.clone(),
                Some(Val::Boolval(b)) => if b.boolval { "on" } else { "off" }.to_string(),
                _ => return Err(PgError::unsupported("this SET value")),
            },
            Some(NodeEnum::TypeCast(cast)) if name == "timezone" => {
                let Some(NodeEnum::AConst(c)) = cast.arg.as_deref().and_then(|a| a.node.as_ref()) else {
                    return Err(PgError::unsupported("this time zone"));
                };
                let Some(Val::Sval(s)) = &c.val else { return Err(PgError::unsupported("this time zone")) };
                let text = s.sval.trim();
                let (sign, rest) = text.strip_prefix('-').map_or((1.0, text.trim_start_matches('+')), |r| (-1.0, r));
                let mut fields = rest.split(':').map(|f| f.parse::<f64>().unwrap_or(0.0));
                let hours = fields.next().unwrap_or(0.0) + fields.next().unwrap_or(0.0) / 60.0;
                (sign * hours).to_string()
            }
            _ => return Err(PgError::unsupported("this SET value")),
        };
        parts.push(text);
    }
    Ok(parts.join(", "))
}

/// command_name names a statement that changes data as Postgres' CreateCommandName does.
fn command_name(node: &NodeEnum) -> &'static str {
    match node {
        NodeEnum::InsertStmt(_) => "INSERT",
        NodeEnum::UpdateStmt(_) => "UPDATE",
        NodeEnum::DeleteStmt(_) => "DELETE",
        NodeEnum::TruncateStmt(_) => "TRUNCATE TABLE",
        NodeEnum::CreateStmt(_) => "CREATE TABLE",
        NodeEnum::CreateTableAsStmt(_) => "CREATE TABLE AS",
        NodeEnum::IndexStmt(_) => "CREATE INDEX",
        NodeEnum::ViewStmt(_) => "CREATE VIEW",
        NodeEnum::DropStmt(drop) => match pg_query::protobuf::ObjectType::try_from(drop.remove_type) {
            Ok(pg_query::protobuf::ObjectType::ObjectIndex) => "DROP INDEX",
            Ok(pg_query::protobuf::ObjectType::ObjectView) => "DROP VIEW",
            Ok(pg_query::protobuf::ObjectType::ObjectSequence) => "DROP SEQUENCE",
            _ => "DROP TABLE",
        },
        _ => "ALTER TABLE",
    }
}

/// describable reports whether describing a statement needs the catalog.
pub(crate) fn describable(node: &NodeEnum) -> bool {
    matches!(
        node,
        NodeEnum::SelectStmt(_)
            | NodeEnum::InsertStmt(_)
            | NodeEnum::UpdateStmt(_)
            | NodeEnum::DeleteStmt(_)
            | NodeEnum::CallStmt(_)
            | NodeEnum::ExplainStmt(_)
    )
}

impl Ctx<'_> {
    /// describe plans a statement for its result columns, collecting its parameter types.
    pub(crate) fn describe(&mut self, node: &NodeEnum) -> Result<Option<Vec<Column>>> {
        Ok(match node {
            NodeEnum::SelectStmt(select) => Some(Planner { ctx: self, outer: Vec::new() }.plan_query(select)?.columns),
            NodeEnum::InsertStmt(_) | NodeEnum::UpdateStmt(_) | NodeEnum::DeleteStmt(_)
                if let Some(columns) = self.on_target_branch(node, |ctx, node| ctx.describe(node))? =>
            {
                columns
            }
            NodeEnum::InsertStmt(_) | NodeEnum::UpdateStmt(_) | NodeEnum::DeleteStmt(_)
                if self.is_branch_control_dml(node)? =>
            {
                None
            }
            NodeEnum::InsertStmt(insert) => self.plan_insert(insert)?.returning.map(|r| r.columns),
            NodeEnum::UpdateStmt(update) if self.is_conflicts_table(update.relation.as_ref())? => None,
            NodeEnum::DeleteStmt(delete) if self.is_conflicts_table(delete.relation.as_ref())? => None,
            NodeEnum::UpdateStmt(update) => self.plan_update(update)?.returning.map(|r| r.columns),
            NodeEnum::DeleteStmt(delete) => self.plan_delete(delete)?.returning.map(|r| r.columns),
            NodeEnum::CallStmt(call) => self.call_columns(call)?,
            NodeEnum::ExplainStmt(stmt) => Some(self.explain_columns(stmt)?),
            _ => None,
        })
    }

    /// on_target_branch runs a function with an INSERT, UPDATE, or DELETE on the branch whose table it changes, with its
    /// table named as that branch names it, when a `database/branch` qualifier or a dolt_nonlocal_tables rule puts
    /// that table on another branch, or returns None otherwise.
    fn on_target_branch<T>(
        &mut self,
        node: &NodeEnum,
        f: impl FnOnce(&mut Ctx<'_>, &NodeEnum) -> Result<T>,
    ) -> Result<Option<T>> {
        let relation = match node {
            NodeEnum::InsertStmt(insert) => insert.relation.as_ref(),
            NodeEnum::UpdateStmt(update) => update.relation.as_ref(),
            NodeEnum::DeleteStmt(delete) => delete.relation.as_ref(),
            _ => None,
        };
        let Some(relation) = relation else { return Ok(None) };
        let qualified = relation.catalogname.split_once('/').filter(|(d, _)| *d == self.session.database);
        let (branch, relation) = match qualified {
            Some((_, branch)) => (branch.to_string(), RangeVar { catalogname: String::new(), ..relation.clone() }),
            None => match self.nonlocal_target(relation)? {
                Some((branch, _)) if branch != self.txn.branch && self.branch_root(&branch)?.is_none() => {
                    let command = match node {
                        NodeEnum::InsertStmt(_) => "INSERT INTO",
                        NodeEnum::UpdateStmt(_) => "UPDATE",
                        _ => "DELETE FROM",
                    };
                    return Err(crate::dolt::args::error(format!("table doesn't support {command}")));
                }
                Some((branch, renamed)) => {
                    let catalogname = format!("{}/{branch}", self.session.database);
                    (branch, RangeVar { catalogname, ..renamed })
                }
                None => return Ok(None),
            },
        };
        if branch == self.txn.branch || self.branch_root(&branch)?.is_none() {
            return Ok(None);
        }
        let mut node = node.clone();
        match &mut node {
            NodeEnum::InsertStmt(insert) => insert.relation = Some(relation),
            NodeEnum::UpdateStmt(update) => update.relation = Some(relation),
            NodeEnum::DeleteStmt(delete) => delete.relation = Some(relation),
            _ => {}
        }
        self.on_branch(&branch, |ctx| f(ctx, &node)).map(Some)
    }

    /// run plans and runs a statement, after checking that the session may write to its branch when the statement
    /// changes tables.
    pub(crate) fn run(&mut self, node: &NodeEnum) -> Result<Outcome> {
        let writes = matches!(
            node,
            NodeEnum::InsertStmt(_)
                | NodeEnum::UpdateStmt(_)
                | NodeEnum::DeleteStmt(_)
                | NodeEnum::TruncateStmt(_)
                | NodeEnum::CreateStmt(_)
                | NodeEnum::CreateTableAsStmt(_)
                | NodeEnum::DropStmt(_)
                | NodeEnum::AlterTableStmt(_)
                | NodeEnum::RenameStmt(_)
                | NodeEnum::IndexStmt(_)
                | NodeEnum::ViewStmt(_)
        );
        if writes
            && (self.session.setting_on("transaction_read_only")
                || !self.session.explicit && self.session.setting_on("default_transaction_read_only"))
        {
            return Err(PgError::new(
                code::READ_ONLY_SQL_TRANSACTION,
                format!("cannot execute {} in a read-only transaction", command_name(node)),
            ));
        }
        if writes && self.txn.detached {
            return Err(PgError::internal(format!("Database {} is read-only.", self.session.display)));
        }
        if writes && !self.is_branch_control_dml(node)? {
            self.check_branch_write()?;
        }
        match node {
            NodeEnum::SelectStmt(select) => {
                let query = Planner { ctx: self, outer: Vec::new() }.plan_query(select)?;
                let rows = query.plan.run(self)?;
                let tag = format!("SELECT {}", rows.len());
                Ok(Outcome::Rows { columns: query.columns, rows, tag })
            }
            NodeEnum::InsertStmt(_) | NodeEnum::UpdateStmt(_) | NodeEnum::DeleteStmt(_)
                if let Some(outcome) = self.on_target_branch(node, |ctx, node| ctx.run(node))? =>
            {
                Ok(outcome)
            }
            NodeEnum::InsertStmt(_) | NodeEnum::UpdateStmt(_) | NodeEnum::DeleteStmt(_)
                if let Some(outcome) = self.branch_control_dml(node)? =>
            {
                Ok(outcome)
            }
            NodeEnum::InsertStmt(insert) => self.plan_insert(insert)?.run(self),
            NodeEnum::UpdateStmt(update) => match self.update_object_conflicts(update)? {
                Some(outcome) => Ok(outcome),
                None => match self.update_table_conflicts(update)? {
                    Some(outcome) => Ok(outcome),
                    None => self.plan_update(update)?.run(self),
                },
            },
            NodeEnum::DeleteStmt(delete) => match self.delete_artifacts(delete)? {
                Some(outcome) => Ok(outcome),
                None => self.plan_delete(delete)?.run(self),
            },
            NodeEnum::CreateStmt(create) => self.create_table(create),
            NodeEnum::CreateTableAsStmt(create) => self.create_table_as(create),
            NodeEnum::CreateSchemaStmt(create) => self.create_schema(create),
            NodeEnum::DropStmt(drop) => self.drop(drop),
            NodeEnum::TruncateStmt(truncate) => self.truncate(truncate),
            NodeEnum::IndexStmt(stmt) => self.create_index(stmt),
            NodeEnum::CreateSeqStmt(stmt) => self.create_sequence(stmt),
            NodeEnum::AlterTableStmt(stmt) => self.alter_table(stmt),
            NodeEnum::RenameStmt(stmt) => self.rename(stmt),
            NodeEnum::ViewStmt(stmt) => {
                let text = pg_query::NodeRef::ViewStmt(stmt).deparse().map_err(PgError::internal)?;
                self.create_view(stmt, &text)
            }
            NodeEnum::DoStmt(stmt) => {
                let text = pg_query::NodeRef::DoStmt(stmt).deparse().map_err(PgError::internal)?;
                self.do_block(stmt, &text)
            }
            NodeEnum::AlterRoleStmt(stmt) => self.alter_role(stmt),
            NodeEnum::DropRoleStmt(stmt) => self.drop_role(stmt),
            NodeEnum::GrantStmt(stmt) => self.grant(stmt),
            NodeEnum::GrantRoleStmt(stmt) => self.grant_role(stmt),
            NodeEnum::CallStmt(stmt) => self.call_procedure(stmt),
            NodeEnum::CreateEnumStmt(stmt) => self.create_enum(stmt),
            NodeEnum::CompositeTypeStmt(stmt) => self.create_composite(stmt),
            NodeEnum::CreateDomainStmt(stmt) => self.create_domain(stmt),
            NodeEnum::AlterEnumStmt(stmt) => self.alter_enum(stmt),
            NodeEnum::CreateExtensionStmt(stmt) => self.create_extension(stmt),
            NodeEnum::CopyStmt(stmt) => self.copy(stmt),
            NodeEnum::ConstraintsSetStmt(stmt) => self.set_constraints(stmt),
            NodeEnum::AlterSeqStmt(stmt) => self.alter_sequence(stmt),
            NodeEnum::AlterOwnerStmt(stmt) => self.alter_owner(stmt),
            NodeEnum::VacuumStmt(stmt) => self.analyze(stmt),
            NodeEnum::ExplainStmt(stmt) => self.explain(stmt),
            _ => Err(PgError::unsupported("this statement")),
        }
    }
}
