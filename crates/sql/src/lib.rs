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
//! session, which parses them with Postgres' own grammar.

pub mod error;

use std::path::{Path, PathBuf};
use std::time::{SystemTime, UNIX_EPOCH};

use pg_query::NodeEnum;
use pg_query::protobuf::a_const::Val;

pub use error::{PgError, Result, code};

/// DEFAULT_BRANCH is the branch a new database starts on.
pub const DEFAULT_BRANCH: &str = "main";

/// Type OIDs the engine returns.
pub mod oid {
    pub const BOOL: u32 = 16;
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
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Outcome {
    /// Rows, in text format, with the command tag that ends them.
    Rows { columns: Vec<Column>, rows: Vec<Vec<Option<Vec<u8>>>>, tag: String },
    /// A command without rows.
    Command { tag: String },
    /// An empty query.
    Empty,
}

/// Engine serves the databases in a data directory.
pub struct Engine {
    data_dir: PathBuf,
}

/// unix_millis returns the current Unix time in milliseconds.
fn unix_millis() -> u64 {
    SystemTime::now().duration_since(UNIX_EPOCH).map_or(0, |d| d.as_millis() as u64)
}

impl Engine {
    /// open opens the data directory, creating it and the default database, named after the superuser, when they
    /// do not exist, as the Go server does on its first start.
    pub fn open(data_dir: &Path, superuser: &str) -> Result<Engine> {
        std::fs::create_dir_all(data_dir.join(".dolt")).map_err(PgError::internal)?;
        let database = data_dir.join(superuser);
        if !database.join(".dolt").is_dir() {
            let millis = unix_millis();
            let times = doltdb::create::CreateTimes {
                init_author_millis: millis as i64,
                init_committer_millis: millis,
                environment_seconds: millis / 1000,
                session_seconds: millis / 1000,
                commit_millis: millis,
            };
            doltdb::create::create_database(&database, DEFAULT_BRANCH, superuser, "localhost", &times)?;
        }
        Ok(Engine { data_dir: data_dir.to_path_buf() })
    }

    /// database_exists reports whether the data directory holds the database.
    pub fn database_exists(&self, name: &str) -> bool {
        !name.is_empty() && !name.contains(['/', '\\']) && self.data_dir.join(name).join(".dolt").is_dir()
    }

    /// session starts a session for the user on the database.
    pub fn session(&self, user: &str, database: &str) -> Result<Session> {
        if !self.database_exists(database) {
            return Err(PgError::fatal(code::INVALID_CATALOG_NAME, format!("database \"{database}\" does not exist")));
        }
        Ok(Session { user: user.to_string(), database: database.to_string() })
    }
}

/// Session runs statements for one connection.
pub struct Session {
    pub user: String,
    pub database: String,
}

impl Session {
    /// execute runs the statements of a simple query, stopping at the first error, and returns what each produced.
    pub fn execute(&mut self, query: &str) -> (Vec<Outcome>, Option<PgError>) {
        let parsed = match pg_query::parse(query) {
            Ok(parsed) => parsed,
            Err(err) => return (Vec::new(), Some(parse_error(err))),
        };
        if parsed.protobuf.stmts.is_empty() {
            return (vec![Outcome::Empty], None);
        }
        let mut outcomes = Vec::new();
        for statement in &parsed.protobuf.stmts {
            let node = statement.stmt.as_ref().and_then(|node| node.node.as_ref());
            match node.map(|node| self.statement(node)) {
                Some(Ok(outcome)) => outcomes.push(outcome),
                Some(Err(err)) => return (outcomes, Some(err)),
                None => outcomes.push(Outcome::Empty),
            }
        }
        (outcomes, None)
    }

    /// statement runs one parsed statement.
    fn statement(&mut self, node: &NodeEnum) -> Result<Outcome> {
        match node {
            NodeEnum::SelectStmt(select) if select.from_clause.is_empty() => {
                let mut columns = Vec::new();
                let mut row = Vec::new();
                for target in &select.target_list {
                    let Some(NodeEnum::ResTarget(target)) = target.node.as_ref() else {
                        return Err(PgError::unsupported("this target"));
                    };
                    let value = target.val.as_ref().and_then(|v| v.node.as_ref());
                    let (type_oid, type_size, text) = match value {
                        Some(NodeEnum::AConst(c)) => match &c.val {
                            Some(Val::Ival(i)) => (oid::INT4, 4, Some(i.ival.to_string())),
                            Some(Val::Sval(s)) => (oid::UNKNOWN, -2, Some(s.sval.clone())),
                            Some(Val::Boolval(b)) => (oid::BOOL, 1, Some(if b.boolval { "t" } else { "f" }.into())),
                            None if c.isnull => (oid::UNKNOWN, -2, None),
                            _ => return Err(PgError::unsupported("this constant")),
                        },
                        _ => return Err(PgError::unsupported("this expression")),
                    };
                    let name = if target.name.is_empty() { "?column?".to_string() } else { target.name.clone() };
                    columns.push(Column { name, type_oid, type_size, type_modifier: -1 });
                    row.push(text.map(String::into_bytes));
                }
                Ok(Outcome::Rows { columns, rows: vec![row], tag: "SELECT 1".into() })
            }
            _ => Err(PgError::unsupported("this statement")),
        }
    }
}

/// parse_error converts a parser error to Postgres' syntax error.
fn parse_error(err: pg_query::Error) -> PgError {
    let message = match err {
        pg_query::Error::Parse(message) => message,
        other => other.to_string(),
    };
    PgError::new(code::SYNTAX_ERROR, message)
}
