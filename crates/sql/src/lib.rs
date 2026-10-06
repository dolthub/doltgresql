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

//! The SQL engine. An engine serves the databases of a data directory, and each connection runs statements in a
//! session, which parses them with Postgres' own grammar and the Doltgres-only syntax it rejects.

pub mod error;
pub mod parse;
pub mod types;

use std::path::{Path, PathBuf};
use std::sync::Arc;
use std::time::{SystemTime, UNIX_EPOCH};

use pg_query::NodeEnum;
use pg_query::protobuf::a_const::Val;

pub use error::{PgError, Result, code};
use parse::{Extras, Statement};
pub use types::Value;

/// DEFAULT_BRANCH is the branch a new database starts on.
pub const DEFAULT_BRANCH: &str = "main";

/// Type OIDs the engine returns.
pub mod oid {
    pub const BOOL: u32 = 16;
    pub const INT8: u32 = 20;
    pub const INT2: u32 = 21;
    pub const INT4: u32 = 23;
    pub const TEXT: u32 = 25;
    pub const UNKNOWN: u32 = 705;
    pub const NUMERIC: u32 = 1700;
}

/// Column describes a result column.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Column {
    pub name: String,
    pub type_oid: u32,
    /// The type's size in bytes, or -1 for a variable-length type.
    pub type_size: i16,
    pub type_modifier: i32,
}

/// Outcome is what one statement produced.
#[derive(Clone, Debug, PartialEq)]
pub enum Outcome {
    /// Rows, with the command tag that ends them.
    Rows { columns: Vec<Column>, rows: Vec<Vec<Value>>, tag: String },
    /// A command without rows.
    Command { tag: String },
    /// An empty query.
    Empty,
}

/// Prepared is a parsed statement, ready to bind parameters to and execute.
#[derive(Clone, Debug)]
pub struct Prepared {
    /// The statement, or None for an empty query.
    pub statement: Option<Statement>,
    pub parameter_types: Vec<u32>,
    /// The result columns, or None when the statement returns no rows.
    pub columns: Option<Vec<Column>>,
}

/// Engine serves the databases in a data directory. Clones share the same databases.
#[derive(Clone)]
pub struct Engine {
    shared: Arc<Shared>,
}

/// Shared is what every clone of an engine shares.
struct Shared {
    data_dir: PathBuf,
}

/// unix_millis returns the current Unix time in milliseconds.
fn unix_millis() -> u64 {
    SystemTime::now().duration_since(UNIX_EPOCH).map_or(0, |d| d.as_millis() as u64)
}

/// create_times returns the clock readings of creating a database now.
fn create_times() -> doltdb::create::CreateTimes {
    let millis = unix_millis();
    doltdb::create::CreateTimes {
        init_author_millis: millis as i64,
        init_committer_millis: millis,
        environment_seconds: millis / 1000,
        session_seconds: millis / 1000,
        commit_millis: millis,
    }
}

impl Engine {
    /// open opens the data directory, creating it and the default database, named after the superuser, when they
    /// do not exist, as the Go server does on its first start.
    pub fn open(data_dir: &Path, superuser: &str) -> Result<Engine> {
        std::fs::create_dir_all(data_dir.join(".dolt")).map_err(PgError::internal)?;
        let engine = Engine { shared: Arc::new(Shared { data_dir: data_dir.to_path_buf() }) };
        if !engine.database_exists(superuser) {
            doltdb::create::create_database(
                &data_dir.join(superuser),
                DEFAULT_BRANCH,
                superuser,
                "localhost",
                &create_times(),
            )?;
        }
        Ok(engine)
    }

    /// database_exists reports whether the data directory holds the database.
    pub fn database_exists(&self, name: &str) -> bool {
        !name.is_empty() && !name.contains(['/', '\\']) && self.shared.data_dir.join(name).join(".dolt").is_dir()
    }

    /// session starts a session on the database for the user, connected from the host.
    pub fn session(&self, user: &str, host: &str, database: &str) -> Result<Session> {
        if !self.database_exists(database) {
            return Err(PgError::fatal(code::INVALID_CATALOG_NAME, format!("database \"{database}\" does not exist")));
        }
        Ok(Session {
            engine: self.clone(),
            user: user.to_string(),
            host: host.to_string(),
            database: database.to_string(),
        })
    }

    /// branch_exists reports whether the database has the branch.
    fn branch_exists(&self, database: &str, branch: &str) -> Result<bool> {
        let noms = self.shared.data_dir.join(database).join(".dolt/noms");
        let mut db = doltdb::database::Database::open(&noms)?;
        let found = db.head(&doltdb::create::branch_ref(branch))?.is_some();
        db.close()?;
        Ok(found)
    }
}

/// Session runs statements for one connection.
pub struct Session {
    engine: Engine,
    pub user: String,
    /// The address the client connected from.
    pub host: String,
    /// The current database, written as `database/branch` when the session switched to a branch.
    pub database: String,
}

impl Session {
    /// execute runs the statements of a simple query, stopping at the first error, and returns what each produced.
    pub fn execute(&mut self, query: &str) -> (Vec<Outcome>, Option<PgError>) {
        let statements = match parse::parse(query) {
            Ok(statements) => statements,
            Err(err) => return (Vec::new(), Some(err)),
        };
        if statements.is_empty() {
            return (vec![Outcome::Empty], None);
        }
        let mut outcomes = Vec::new();
        for statement in &statements {
            match self.statement(statement) {
                Ok(outcome) => outcomes.push(outcome),
                Err(err) => return (outcomes, Some(err)),
            }
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
        let columns = match &statement {
            Some(Statement::Postgres { node: NodeEnum::SelectStmt(select), .. }) if select.from_clause.is_empty() => {
                Some(select_constants(select)?.0)
            }
            _ => None,
        };
        Ok(Prepared { statement, parameter_types: parameter_types.to_vec(), columns })
    }

    /// execute_prepared runs a prepared statement with the parameter values.
    pub fn execute_prepared(&mut self, prepared: &Prepared, _parameters: &[Value]) -> Result<Outcome> {
        match &prepared.statement {
            Some(statement) => self.statement(statement),
            None => Ok(Outcome::Empty),
        }
    }

    /// statement runs one parsed statement.
    fn statement(&mut self, statement: &Statement) -> Result<Outcome> {
        match statement {
            Statement::Use(target) => self.use_database(target),
            Statement::Postgres { node, extras } => self.postgres_statement(node, extras),
            Statement::SetExpression { .. } => Err(PgError::unsupported("SET to an expression")),
        }
    }

    /// use_database switches the session to a database, or to a branch of one, as Go's USE does.
    fn use_database(&mut self, target: &str) -> Result<Outcome> {
        let not_found = || PgError::new(code::INVALID_CATALOG_NAME, format!("database not found: {target}"));
        let (database, branch) = match target.split_once('/') {
            Some((database, branch)) => (database, Some(branch)),
            None => (target, None),
        };
        if !self.engine.database_exists(database) {
            return Err(not_found());
        }
        if let Some(branch) = branch
            && !self.engine.branch_exists(database, branch)?
        {
            return Err(not_found());
        }
        self.database = target.to_string();
        Ok(Outcome::Command { tag: "SET".into() })
    }

    /// create_database runs CREATE DATABASE, which only Doltgres allows with IF NOT EXISTS.
    fn create_database(&mut self, name: &str, if_not_exists: bool) -> Result<Outcome> {
        let tag = Outcome::Command { tag: "CREATE DATABASE".into() };
        if name.contains(['/', '\\']) || name.is_empty() {
            return Err(PgError::internal(format!("Incorrect database name '{name}'")));
        }
        if self.engine.database_exists(name) {
            if if_not_exists {
                return Ok(tag);
            }
            return Err(PgError::new(code::DUPLICATE_DATABASE, format!("database \"{name}\" already exists")));
        }
        let dir = self.engine.shared.data_dir.join(name);
        doltdb::create::create_database(&dir, DEFAULT_BRANCH, &self.user, &self.host, &create_times())?;
        Ok(tag)
    }

    /// postgres_statement runs a statement of Postgres' grammar.
    fn postgres_statement(&mut self, node: &NodeEnum, extras: &Extras) -> Result<Outcome> {
        match node {
            NodeEnum::CreatedbStmt(create) => self.create_database(&create.dbname, extras.if_not_exists),
            NodeEnum::SelectStmt(select) if select.from_clause.is_empty() => {
                let (columns, row) = select_constants(select)?;
                Ok(Outcome::Rows { columns, rows: vec![row], tag: "SELECT 1".into() })
            }
            _ => Err(PgError::unsupported("this statement")),
        }
    }
}

/// select_constants returns the columns and the row of a SELECT of constants.
fn select_constants(select: &pg_query::protobuf::SelectStmt) -> Result<(Vec<Column>, Vec<Value>)> {
    let mut columns = Vec::new();
    let mut row = Vec::new();
    for target in &select.target_list {
        let Some(NodeEnum::ResTarget(target)) = target.node.as_ref() else {
            return Err(PgError::unsupported("this target"));
        };
        let value = target.val.as_ref().and_then(|v| v.node.as_ref());
        let (type_oid, type_size, value) = match value {
            Some(NodeEnum::AConst(c)) => match &c.val {
                Some(Val::Ival(i)) => (oid::INT4, 4, Value::Int4(i.ival)),
                Some(Val::Sval(s)) => (oid::UNKNOWN, -2, Value::Text(s.sval.clone())),
                Some(Val::Boolval(b)) => (oid::BOOL, 1, Value::Bool(b.boolval)),
                None if c.isnull => (oid::UNKNOWN, -2, Value::Null),
                _ => return Err(PgError::unsupported("this constant")),
            },
            _ => return Err(PgError::unsupported("this expression")),
        };
        let name = if target.name.is_empty() { "?column?".to_string() } else { target.name.clone() };
        columns.push(Column { name, type_oid, type_size, type_modifier: -1 });
        row.push(value);
    }
    Ok((columns, row))
}
