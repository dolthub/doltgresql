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
use pg_query::protobuf::{TransactionStmtKind, VariableSetKind, VariableSetStmt};
use pg_query::{Node, NodeEnum};

use crate::error::{PgError, Result, code};
use crate::parse::{self, Extras, Statement};
use crate::plan::Planner;
use crate::query::{Ctx, column};
use crate::settings::{Settings, setting};
use crate::txn::{DbHandle, SequenceTracker, Txn};
use crate::types::Value;
use crate::{Column, DEFAULT_BRANCH, Outcome, Prepared};

/// Engine serves the databases in a data directory. Clones share the same databases.
#[derive(Clone)]
pub struct Engine {
    shared: Arc<Shared>,
}

/// Shared is what every clone of an engine shares.
struct Shared {
    data_dir: PathBuf,
    databases: Mutex<HashMap<String, (DbHandle, SequenceTracker)>>,
}

/// create_times returns the clock readings of creating a database now.
fn create_times() -> doltdb::create::CreateTimes {
    let millis = SystemTime::now().duration_since(UNIX_EPOCH).map_or(0, |d| d.as_millis() as u64);
    doltdb::create::CreateTimes {
        init_author_millis: millis as i64,
        init_committer_millis: millis,
        environment_seconds: millis / 1000,
        session_seconds: millis / 1000,
        commit_millis: millis,
    }
}

/// lock locks a shared value, failing when a panic poisoned it.
fn lock<T>(mutex: &Mutex<T>) -> Result<std::sync::MutexGuard<'_, T>> {
    mutex.lock().map_err(|_| PgError::internal("a lock was poisoned"))
}

impl Engine {
    /// open opens the data directory, creating it and the default database, named after the superuser, when they
    /// do not exist, as the Go server does on its first start.
    pub fn open(data_dir: &Path, superuser: &str) -> Result<Engine> {
        std::fs::create_dir_all(data_dir.join(".dolt")).map_err(PgError::internal)?;
        let engine = Engine {
            shared: Arc::new(Shared { data_dir: data_dir.to_path_buf(), databases: Mutex::new(HashMap::new()) }),
        };
        if !engine.database_exists(superuser) {
            let dir = data_dir.join(superuser);
            doltdb::create::create_database(&dir, DEFAULT_BRANCH, superuser, "localhost", &create_times())?;
        }
        Ok(engine)
    }

    /// database_exists reports whether the data directory holds the database.
    pub fn database_exists(&self, name: &str) -> bool {
        !name.is_empty() && !name.contains(['/', '\\']) && self.shared.data_dir.join(name).join(".dolt").is_dir()
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
            engine: self.clone(),
            state: SessionState {
                user: user.to_string(),
                host: host.to_string(),
                database: String::new(),
                branch: DEFAULT_BRANCH.to_string(),
                display: String::new(),
                notices: Vec::new(),
                settings: Settings::new(startup).map_err(|err| PgError { severity: "FATAL", ..err })?,
                explicit: false,
                sequence_values: HashMap::new(),
                last_sequence: None,
            },
            txns: Vec::new(),
            failed: false,
            reported: HashMap::new(),
        };
        session.switch(database).map_err(|_| {
            PgError::fatal(code::INVALID_CATALOG_NAME, format!("database \"{database}\" does not exist"))
        })?;
        Ok(session)
    }
}

/// Session runs statements for one connection.
pub struct Session {
    engine: Engine,
    pub state: SessionState,
    /// The open transaction's view of each branch it touched.
    txns: Vec<Txn>,
    /// Whether a statement failed in the explicit transaction, which then only ends.
    failed: bool,
    /// The reported parameters as the client last heard them.
    reported: HashMap<String, String>,
}

/// SessionState is the part of a session that statements and functions can read and change.
pub struct SessionState {
    pub user: String,
    /// The address the client connected from.
    pub host: String,
    pub database: String,
    pub branch: String,
    /// The current database as the session named it, which includes the branch when one was named.
    pub display: String,
    pub notices: Vec<PgError>,
    pub settings: Settings,
    /// Whether the open transaction began with BEGIN.
    pub explicit: bool,
    /// The last value nextval returned for each sequence in this session, by sequence ID.
    pub sequence_values: HashMap<Vec<u8>, i64>,
    /// The sequence ID and value of the session's most recent nextval.
    pub last_sequence: Option<(Vec<u8>, i64)>,
}

impl SessionState {
    /// search_path returns the schemas that unqualified names resolve in.
    pub fn search_path(&self) -> Vec<String> {
        let path = self.settings.get("search_path").unwrap_or_default();
        path.split(',')
            .map(|s| s.trim())
            .filter(|s| !s.is_empty())
            .map(|s| {
                let s = if s.starts_with('"') && s.ends_with('"') && s.len() > 1 {
                    s[1..s.len() - 1].replace("\"\"", "\"")
                } else {
                    s.to_ascii_lowercase()
                };
                if s == "$user" { self.user.clone() } else { s }
            })
            .collect()
    }

    /// install_format installs the session's DateStyle, IntervalStyle, and time zone for printing values.
    pub fn install_format(&self) {
        let get = |name: &str| self.settings.get(name).unwrap_or_default();
        crate::datetime::install_format(crate::datetime::Format::from_settings(
            &get("DateStyle"),
            &get("IntervalStyle"),
            &get("TimeZone"),
        ));
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
        let not_found = || PgError::new(code::INVALID_CATALOG_NAME, format!("database not found: {target}"));
        let (database, branch) = match target.split_once('/') {
            Some((database, branch)) => (database, branch),
            None => (target, DEFAULT_BRANCH),
        };
        if !self.engine.database_exists(database) {
            return Err(not_found());
        }
        let handle = self.engine.database(database)?;
        if lock(&handle)?.head(&doltdb::create::branch_ref(branch))?.is_none() {
            return Err(not_found());
        }
        self.state.database = database.to_string();
        self.state.branch = branch.to_string();
        self.state.display = target.to_string();
        Ok(())
    }

    /// execute runs the statements of a simple query, stopping at the first error, and returns what each produced.
    /// The statements run in one implicit transaction unless they manage their own.
    pub fn execute(&mut self, query: &str) -> (Vec<Outcome>, Option<PgError>) {
        let statements = match parse::parse(query) {
            Ok(statements) => statements,
            Err(err) => return (Vec::new(), Some(self.fail(err))),
        };
        if statements.is_empty() {
            return (vec![Outcome::Empty], None);
        }
        let mut outcomes = Vec::new();
        for statement in &statements {
            match self.run(statement, &[]) {
                Ok(outcome) => outcomes.push(outcome),
                Err(err) => return (outcomes, Some(self.fail(err))),
            }
        }
        if let Err(err) = self.end_implicit() {
            return (outcomes, Some(self.fail(err)));
        }
        (outcomes, None)
    }

    /// prepare parses a query of at most one statement and describes its parameters and results.
    pub fn prepare(&mut self, query: &str, parameter_types: &[u32]) -> Result<Prepared> {
        let mut statements = parse::parse(query)?;
        if statements.len() > 1 {
            return Err(PgError::new(code::SYNTAX_ERROR, "cannot insert multiple commands into a prepared statement"));
        }
        let statement = statements.pop();
        let mut parameters = parameter_types.to_vec();
        let mut columns = None;
        if let Some(Statement::Postgres { node: NodeEnum::VariableShowStmt(show), .. }) = &statement {
            columns = Some(show_columns(&show.name));
        } else if let Some(Statement::Postgres { node, .. }) = &statement
            && describable(node)
        {
            columns = self.with_ctx(&mut parameters, &[], |ctx| ctx.describe(node))?;
            if !self.state.explicit {
                self.txns.clear();
            }
        }
        for parameter in &mut parameters {
            if *parameter == 0 {
                *parameter = crate::oid::TEXT;
            }
        }
        Ok(Prepared { statement, parameter_types: parameters, columns })
    }

    /// execute_prepared runs a prepared statement with the parameter values.
    pub fn execute_prepared(&mut self, prepared: &Prepared, parameters: &[Value]) -> Result<Outcome> {
        let Some(statement) = &prepared.statement else { return Ok(Outcome::Empty) };
        let result = self.run(statement, parameters).and_then(|outcome| self.end_implicit().map(|_| outcome));
        result.map_err(|err| self.fail(err))
    }

    /// fail ends an implicit transaction, or marks an explicit one failed, after an error.
    fn fail(&mut self, err: PgError) -> PgError {
        if self.state.explicit {
            self.failed = true;
        } else {
            self.txns.clear();
            self.state.settings.end_transaction(false);
        }
        err
    }

    /// end_implicit commits the transaction when it is implicit.
    fn end_implicit(&mut self) -> Result<()> {
        if self.state.explicit {
            return Ok(());
        }
        self.commit()
    }

    /// commit commits and ends the open transaction.
    fn commit(&mut self) -> Result<()> {
        self.state.settings.end_transaction(true);
        for txn in std::mem::take(&mut self.txns) {
            let handle = txn.handle.clone();
            let mut db = lock(&handle)?;
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
        let (database, branch) = (&self.state.database, &self.state.branch);
        let index = match self.txns.iter().position(|t| t.database == *database && t.branch == *branch) {
            Some(index) => index,
            None => {
                let (handle, tracker) = self.engine.open_database(database)?;
                let mut txn = Txn::begin(handle, tracker, database, branch)?;
                if let Some(first) = self.txns.first() {
                    txn.started = first.started;
                }
                self.txns.push(txn);
                self.txns.len() - 1
            }
        };
        let txn = &mut self.txns[index];
        crate::datetime::install_now(txn.started);
        self.state.install_format();
        let handle = txn.handle.clone();
        let mut db = lock(&handle)?;
        let mut ctx = Ctx {
            db: &mut db,
            txn,
            session: &mut self.state,
            parameters,
            params,
            outer: Vec::new(),
            subquery_value: Value::Null,
        };
        f(&mut ctx)
    }

    /// run runs one statement with the parameter values.
    fn run(&mut self, statement: &Statement, params: &[Value]) -> Result<Outcome> {
        let kind = transaction_kind(statement);
        if self.failed {
            return match kind {
                Some(TransactionStmtKind::TransStmtCommit | TransactionStmtKind::TransStmtRollback) => {
                    self.txns.clear();
                    self.state.explicit = false;
                    self.failed = false;
                    self.state.settings.end_transaction(false);
                    Ok(Outcome::command("ROLLBACK"))
                }
                _ => Err(PgError::new(
                    code::IN_FAILED_SQL_TRANSACTION,
                    "current transaction is aborted, commands ignored until end of transaction block",
                )),
            };
        }
        if let Some(kind) = kind {
            return self.transaction(kind);
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
            Statement::Postgres { node, extras } => self.postgres(node, extras, params),
        }
    }

    /// transaction runs BEGIN, COMMIT, or ROLLBACK.
    fn transaction(&mut self, kind: TransactionStmtKind) -> Result<Outcome> {
        match kind {
            TransactionStmtKind::TransStmtBegin | TransactionStmtKind::TransStmtStart => {
                if self.state.explicit {
                    self.state.notices.push(PgError {
                        severity: "WARNING",
                        ..PgError::new(code::ACTIVE_SQL_TRANSACTION, "there is already a transaction in progress")
                    });
                }
                self.state.explicit = true;
                Ok(Outcome::command("BEGIN"))
            }
            TransactionStmtKind::TransStmtCommit => {
                if !self.state.explicit {
                    self.state.notices.push(PgError {
                        severity: "WARNING",
                        ..PgError::new(code::NO_ACTIVE_SQL_TRANSACTION, "there is no transaction in progress")
                    });
                }
                self.state.explicit = false;
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
                self.state.settings.end_transaction(false);
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
            return Err(PgError::internal(format!("Incorrect database name '{name}'")));
        }
        if self.engine.database_exists(name) {
            if if_not_exists {
                return Ok(Outcome::command("CREATE DATABASE"));
            }
            return Err(PgError::new(code::DUPLICATE_DATABASE, format!("database \"{name}\" already exists")));
        }
        let dir = self.engine.shared.data_dir.join(name);
        doltdb::create::create_database(&dir, DEFAULT_BRANCH, &self.state.user, &self.state.host, &create_times())?;
        Ok(Outcome::command("CREATE DATABASE"))
    }

    /// postgres runs a statement of Postgres' grammar.
    fn postgres(&mut self, node: &NodeEnum, extras: &Extras, params: &[Value]) -> Result<Outcome> {
        match node {
            NodeEnum::CreatedbStmt(create) => return self.create_database(&create.dbname, extras.if_not_exists),
            NodeEnum::VariableSetStmt(set) => return self.set(set),
            NodeEnum::VariableShowStmt(show) => return self.show(&show.name),
            _ => {}
        }
        let mut parameters = Vec::new();
        self.with_ctx(&mut parameters, params, |ctx| ctx.run(node))
    }
}

impl Session {
    /// set runs SET and RESET.
    fn set(&mut self, set: &VariableSetStmt) -> Result<Outcome> {
        let kind = VariableSetKind::try_from(set.kind).unwrap_or(VariableSetKind::Undefined);
        let in_transaction = self.state.explicit;
        let reset =
            matches!(kind, VariableSetKind::VarReset | VariableSetKind::VarResetAll | VariableSetKind::VarSetDefault);
        let tag =
            if matches!(kind, VariableSetKind::VarReset | VariableSetKind::VarResetAll) { "RESET" } else { "SET" };
        let word = if reset { "RESET" } else { "SET" };
        // The transaction characteristics only exist in a transaction block.
        let transactional =
            matches!(set.name.as_str(), "transaction_isolation" | "transaction_read_only" | "transaction_deferrable");
        if transactional && !in_transaction && kind != VariableSetKind::VarResetAll {
            self.state.notices.push(PgError {
                severity: "WARNING",
                ..PgError::new(
                    code::NO_ACTIVE_SQL_TRANSACTION,
                    format!("{word} TRANSACTION can only be used in transaction blocks"),
                )
            });
            return Ok(Outcome::command(tag));
        }
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

/// describable reports whether describing a statement needs the catalog.
fn describable(node: &NodeEnum) -> bool {
    matches!(
        node,
        NodeEnum::SelectStmt(_) | NodeEnum::InsertStmt(_) | NodeEnum::UpdateStmt(_) | NodeEnum::DeleteStmt(_)
    )
}

impl Ctx<'_> {
    /// describe plans a statement for its result columns, collecting its parameter types.
    fn describe(&mut self, node: &NodeEnum) -> Result<Option<Vec<Column>>> {
        Ok(match node {
            NodeEnum::SelectStmt(select) => Some(Planner { ctx: self, outer: Vec::new() }.plan_query(select)?.columns),
            NodeEnum::InsertStmt(insert) => {
                self.plan_insert(insert)?;
                None
            }
            NodeEnum::UpdateStmt(update) => {
                self.plan_update(update)?;
                None
            }
            NodeEnum::DeleteStmt(delete) => {
                self.plan_delete(delete)?;
                None
            }
            _ => None,
        })
    }

    /// run plans and runs a statement.
    fn run(&mut self, node: &NodeEnum) -> Result<Outcome> {
        match node {
            NodeEnum::SelectStmt(select) => {
                let query = Planner { ctx: self, outer: Vec::new() }.plan_query(select)?;
                let rows = query.plan.run(self)?;
                let tag = format!("SELECT {}", rows.len());
                Ok(Outcome::Rows { columns: query.columns, rows, tag })
            }
            NodeEnum::InsertStmt(insert) => self.plan_insert(insert)?.run(self),
            NodeEnum::UpdateStmt(update) => self.plan_update(update)?.run(self),
            NodeEnum::DeleteStmt(delete) => self.plan_delete(delete)?.run(self),
            NodeEnum::CreateStmt(create) => self.create_table(create),
            NodeEnum::CreateTableAsStmt(create) => self.create_table_as(create),
            NodeEnum::CreateSchemaStmt(create) => self.create_schema(create),
            NodeEnum::DropStmt(drop) => self.drop(drop),
            NodeEnum::TruncateStmt(truncate) => self.truncate(truncate),
            NodeEnum::IndexStmt(stmt) => self.create_index(stmt),
            NodeEnum::CreateSeqStmt(stmt) => self.create_sequence(stmt),
            NodeEnum::AlterTableStmt(stmt) => self.alter_table(stmt),
            NodeEnum::RenameStmt(stmt) => self.rename(stmt),
            NodeEnum::ViewStmt(stmt) => self.create_view(stmt),
            _ => Err(PgError::unsupported("this statement")),
        }
    }
}
