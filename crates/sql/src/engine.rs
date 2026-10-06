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
use pg_query::NodeEnum;
use pg_query::protobuf::TransactionStmtKind;

use crate::error::{PgError, Result, code};
use crate::parse::{self, Extras, Statement};
use crate::query::Ctx;
use crate::txn::{DbHandle, Txn};
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
    databases: Mutex<HashMap<String, DbHandle>>,
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
        let mut databases = lock(&self.shared.databases)?;
        if let Some(handle) = databases.get(name) {
            return Ok(handle.clone());
        }
        let handle = Arc::new(Mutex::new(Database::open(&self.shared.data_dir.join(name).join(".dolt/noms"))?));
        databases.insert(name.to_string(), handle.clone());
        Ok(handle)
    }

    /// session starts a session for the user, connected from the host, on a database or a branch of one written as
    /// `database/branch`.
    pub fn session(&self, user: &str, host: &str, database: &str) -> Result<Session> {
        let mut session = Session {
            engine: self.clone(),
            user: user.to_string(),
            host: host.to_string(),
            database: String::new(),
            branch: DEFAULT_BRANCH.to_string(),
            display: String::new(),
            txn: None,
            explicit: false,
            failed: false,
            notices: Vec::new(),
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
    pub user: String,
    /// The address the client connected from.
    pub host: String,
    database: String,
    branch: String,
    /// The current database as the session named it, which includes the branch when one was named.
    display: String,
    txn: Option<Txn>,
    /// Whether the open transaction began with BEGIN.
    explicit: bool,
    /// Whether a statement failed in the explicit transaction, which then only ends.
    failed: bool,
    notices: Vec<PgError>,
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
        &self.display
    }

    /// tx_status returns the transaction status that ReadyForQuery reports: idle, in a transaction, or failed.
    pub fn tx_status(&self) -> u8 {
        match (self.explicit, self.failed) {
            (true, true) => b'E',
            (true, false) => b'T',
            _ => b'I',
        }
    }

    /// take_notices returns the notices raised since the last call.
    pub fn take_notices(&mut self) -> Vec<PgError> {
        std::mem::take(&mut self.notices)
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
        self.database = database.to_string();
        self.branch = branch.to_string();
        self.display = target.to_string();
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
        if let Some(Statement::Postgres { node, .. }) = &statement
            && describable(node)
        {
            columns = self.with_ctx(&mut parameters, |ctx| ctx.describe(node))?;
            if !self.explicit {
                self.txn = None;
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
        if self.explicit {
            self.failed = true;
        } else {
            self.txn = None;
        }
        err
    }

    /// end_implicit commits the transaction when it is implicit.
    fn end_implicit(&mut self) -> Result<()> {
        if self.explicit {
            return Ok(());
        }
        self.commit()
    }

    /// commit commits and ends the open transaction.
    fn commit(&mut self) -> Result<()> {
        let Some(txn) = self.txn.take() else { return Ok(()) };
        let handle = txn.handle.clone();
        let mut db = lock(&handle)?;
        txn.commit(&mut db, &self.user, &self.host)
    }

    /// with_ctx runs a function with the planning context of the open transaction, beginning one when needed.
    fn with_ctx<T>(&mut self, parameters: &mut Vec<u32>, f: impl FnOnce(&mut Ctx<'_>) -> Result<T>) -> Result<T> {
        if self.txn.is_none() {
            let handle = self.engine.database(&self.database)?;
            self.txn = Some(Txn::begin(handle, &self.database, &self.branch)?);
        }
        let txn = self.txn.as_mut().expect("an open transaction");
        let handle = txn.handle.clone();
        let mut db = lock(&handle)?;
        let mut ctx = Ctx { db: &mut db, txn, parameters, notices: &mut self.notices };
        f(&mut ctx)
    }

    /// run runs one statement with the parameter values.
    fn run(&mut self, statement: &Statement, params: &[Value]) -> Result<Outcome> {
        let kind = transaction_kind(statement);
        if self.failed {
            return match kind {
                Some(TransactionStmtKind::TransStmtCommit | TransactionStmtKind::TransStmtRollback) => {
                    self.txn = None;
                    self.explicit = false;
                    self.failed = false;
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
            Statement::SetExpression { .. } => Err(PgError::unsupported("SET to an expression")),
            Statement::Postgres { node, extras } => self.postgres(node, extras, params),
        }
    }

    /// transaction runs BEGIN, COMMIT, or ROLLBACK.
    fn transaction(&mut self, kind: TransactionStmtKind) -> Result<Outcome> {
        match kind {
            TransactionStmtKind::TransStmtBegin | TransactionStmtKind::TransStmtStart => {
                if self.explicit {
                    self.notices.push(PgError {
                        severity: "WARNING",
                        ..PgError::new(code::ACTIVE_SQL_TRANSACTION, "there is already a transaction in progress")
                    });
                }
                self.explicit = true;
                Ok(Outcome::command("BEGIN"))
            }
            TransactionStmtKind::TransStmtCommit => {
                if !self.explicit {
                    self.notices.push(PgError {
                        severity: "WARNING",
                        ..PgError::new(code::NO_ACTIVE_SQL_TRANSACTION, "there is no transaction in progress")
                    });
                }
                self.explicit = false;
                self.commit()?;
                Ok(Outcome::command("COMMIT"))
            }
            TransactionStmtKind::TransStmtRollback => {
                if !self.explicit {
                    self.notices.push(PgError {
                        severity: "WARNING",
                        ..PgError::new(code::NO_ACTIVE_SQL_TRANSACTION, "there is no transaction in progress")
                    });
                }
                self.explicit = false;
                self.txn = None;
                Ok(Outcome::command("ROLLBACK"))
            }
            other => Err(PgError::unsupported(format!("the transaction statement {other:?}"))),
        }
    }

    /// create_database runs CREATE DATABASE, which only Doltgres allows with IF NOT EXISTS.
    fn create_database(&mut self, name: &str, if_not_exists: bool) -> Result<Outcome> {
        if self.explicit {
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
        doltdb::create::create_database(&dir, DEFAULT_BRANCH, &self.user, &self.host, &create_times())?;
        Ok(Outcome::command("CREATE DATABASE"))
    }

    /// postgres runs a statement of Postgres' grammar.
    fn postgres(&mut self, node: &NodeEnum, extras: &Extras, params: &[Value]) -> Result<Outcome> {
        if let NodeEnum::CreatedbStmt(create) = node {
            return self.create_database(&create.dbname, extras.if_not_exists);
        }
        let mut parameters = Vec::new();
        self.with_ctx(&mut parameters, |ctx| ctx.run(node, params))
    }
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
            NodeEnum::SelectStmt(select) => Some(self.plan_select(select)?.columns),
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
    fn run(&mut self, node: &NodeEnum, params: &[Value]) -> Result<Outcome> {
        match node {
            NodeEnum::SelectStmt(select) => {
                let plan = self.plan_select(select)?;
                let rows = plan.run(self.db, params)?;
                let tag = format!("SELECT {}", rows.len());
                Ok(Outcome::Rows { columns: plan.columns, rows, tag })
            }
            NodeEnum::InsertStmt(insert) => self.plan_insert(insert)?.run(self.db, self.txn, params),
            NodeEnum::UpdateStmt(update) => self.plan_update(update)?.run(self.db, self.txn, params),
            NodeEnum::DeleteStmt(delete) => self.plan_delete(delete)?.run(self.db, self.txn, params),
            NodeEnum::CreateStmt(create) => self.create_table(create),
            _ => Err(PgError::unsupported("this statement")),
        }
    }
}
