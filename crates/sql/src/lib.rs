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

mod alter;
pub mod array;
pub mod cast;
pub mod catalog;
pub mod datetime;
mod ddl;
pub mod dml;
pub mod dolt;
mod engine;
pub mod error;
pub mod expr;
mod foreign;
pub mod functions;
pub mod json;
pub mod numeric;
pub mod parse;
pub mod plan;
pub mod query;
pub mod sequences;
pub mod settings;
pub mod storage;
pub mod txn;
pub mod types;
pub mod views;
pub mod window;

pub use engine::{Engine, Session};
pub use error::{PgError, Result, code};
use parse::Statement;
pub use types::Value;

/// DOLTGRES_VERSION is the Doltgres version the server reports.
pub const DOLTGRES_VERSION: &str = "1.4.0";

/// SERVER_VERSION is the Postgres version the server reports.
pub const SERVER_VERSION: &str = "15.17";

/// DEFAULT_BRANCH is the branch a new database starts on.
pub const DEFAULT_BRANCH: &str = "main";

/// Type OIDs the engine returns.
pub mod oid {
    pub const BOOL: u32 = 16;
    pub const NAME: u32 = 19;
    pub const INT8: u32 = 20;
    pub const INT2: u32 = 21;
    pub const INT4: u32 = 23;
    pub const TEXT: u32 = 25;
    pub const TEXT_ARRAY: u32 = 1009;
    pub const JSON: u32 = 114;
    pub const JSONB: u32 = 3802;
    pub const RECORD: u32 = 2249;
    pub const FLOAT4: u32 = 700;
    pub const FLOAT8: u32 = 701;
    pub const UNKNOWN: u32 = 705;
    pub const BPCHAR: u32 = 1042;
    pub const VARCHAR: u32 = 1043;
    pub const DATE: u32 = 1082;
    pub const TIME: u32 = 1083;
    pub const TIMESTAMP: u32 = 1114;
    pub const TIMESTAMPTZ: u32 = 1184;
    pub const INTERVAL: u32 = 1186;
    pub const TIMETZ: u32 = 1266;
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

impl Outcome {
    /// command returns the outcome of a command with the tag.
    pub fn command(tag: impl Into<String>) -> Outcome {
        Outcome::Command { tag: tag.into() }
    }
}
