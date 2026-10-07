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

pub mod advisory;
pub mod aggregates;
mod alter;
pub mod array;
pub mod auth;
pub mod basetypes;
pub mod binary;
pub mod cast;
pub mod casts;
pub mod catalog;
pub mod cluster;
pub mod copy;
pub mod datetime;
mod ddl;
pub mod deferred;
pub mod dml;
pub mod dolt;
pub mod encodings;
mod engine;
pub mod error;
pub mod explain;
pub mod expr;
pub mod extensions;
mod foreign;
pub mod formatting;
pub mod functions;
pub mod indexscan;
pub mod integrity;
pub mod json;
pub mod jsonpath;
pub mod jsontable;
mod listing;
pub mod numeric;
pub mod numeric_math;
pub mod operators;
pub mod parse;
mod pgcatalog;
pub mod plan;
mod plpgsql;
pub mod query;
pub mod ranges;
mod roles;
pub mod routines;
pub mod ruleutils;
pub mod sequences;
pub mod settings;
pub mod stats;
pub mod storage;
mod triggers;
pub mod txn;
pub mod types;
pub mod usertypes;
pub mod views;
pub mod window;
pub mod xml;

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
    /// ACLITEM_ARRAY is the type of the catalogs' access privilege columns, an array of text-like elements to
    /// Doltgres, which has no aclitem type.
    pub const ACLITEM_ARRAY: u32 = 1034;
    pub const JSON: u32 = 114;
    pub const JSONB: u32 = 3802;
    pub const RECORD: u32 = 2249;
    pub const CSTRING: u32 = 2275;
    pub const ANYENUM: u32 = 3500;
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
    pub const CHAR: u32 = 18;
    pub const REGPROC: u32 = 24;
    pub const OID: u32 = 26;
    pub const XID: u32 = 28;
    pub const CID: u32 = 29;
    pub const REGPROCEDURE: u32 = 2202;
    pub const REGOPER: u32 = 2203;
    pub const REGOPERATOR: u32 = 2204;
    pub const REGCLASS: u32 = 2205;
    pub const REGTYPE: u32 = 2206;
    pub const REGNAMESPACE: u32 = 4089;
    pub const REGROLE: u32 = 4096;
    pub const BYTEA: u32 = 17;
    pub const UUID: u32 = 2950;
    pub const BIT: u32 = 1560;
    pub const VARBIT: u32 = 1562;
    pub const XML: u32 = 142;
    pub const INT2VECTOR: u32 = 22;
    pub const OIDVECTOR: u32 = 30;
    pub const XML_ARRAY: u32 = 143;
}

/// Column describes a result column.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Column {
    pub name: String,
    pub type_oid: u32,
    /// The type's size in bytes, or -1 for a variable-length type.
    pub type_size: i16,
    pub type_modifier: i32,
    /// The OID and attribute number of the table column it comes from, or zeros for any other column.
    pub origin: (u32, u16),
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
    /// A COPY FROM STDIN waiting for the client to send the data of its columns, in the binary format or as text.
    CopyIn { binary: bool, columns: usize },
    /// The data of a COPY TO STDOUT, in chunks to send in turn, with the command tag that ends it.
    CopyOut { binary: bool, columns: usize, chunks: Vec<Vec<u8>>, tag: String },
}

/// Results are the outcomes of a simple query's statements, each with the notices its statement raised first.
pub type Results = Vec<(Vec<PgError>, Outcome)>;

/// Prepared is a parsed statement, ready to bind parameters to and execute.
#[derive(Clone, Debug)]
pub struct Prepared {
    /// The query text, which pg_stat_activity shows while the statement runs.
    pub query: String,
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
