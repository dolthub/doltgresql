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

//! Errors as Postgres reports them: a severity, an SQLSTATE code, and a message with optional detail.

use std::fmt;

/// SQLSTATE codes the engine reports.
pub mod code {
    pub const SYNTAX_ERROR: &str = "42601";
    pub const FEATURE_NOT_SUPPORTED: &str = "0A000";
    pub const INTERNAL_ERROR: &str = "XX000";
    pub const INVALID_PARAMETER_VALUE: &str = "22023";
    pub const INVALID_CATALOG_NAME: &str = "3D000";
    pub const DUPLICATE_DATABASE: &str = "42P04";
    pub const INVALID_BINARY_REPRESENTATION: &str = "22P03";
    pub const CHARACTER_NOT_IN_REPERTOIRE: &str = "22021";
    pub const PROTOCOL_VIOLATION: &str = "08P01";
    pub const INVALID_SQL_STATEMENT_NAME: &str = "26000";
    pub const INVALID_CURSOR_NAME: &str = "34000";
    pub const DUPLICATE_PREPARED_STATEMENT: &str = "42P05";
}

/// PgError is an error to report to the client.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct PgError {
    /// ERROR, FATAL, or PANIC.
    pub severity: &'static str,
    pub code: &'static str,
    pub message: String,
    pub detail: Option<String>,
    pub hint: Option<String>,
    /// The 1-based character position in the query that the error refers to.
    pub position: Option<u32>,
}

impl PgError {
    /// new returns an ERROR with the code and message.
    pub fn new(code: &'static str, message: impl Into<String>) -> PgError {
        PgError { severity: "ERROR", code, message: message.into(), detail: None, hint: None, position: None }
    }

    /// fatal returns a FATAL error, which ends the connection.
    pub fn fatal(code: &'static str, message: impl Into<String>) -> PgError {
        PgError { severity: "FATAL", ..PgError::new(code, message) }
    }

    /// internal returns an error for a failure that is not the client's.
    pub fn internal(message: impl fmt::Display) -> PgError {
        PgError::new(code::INTERNAL_ERROR, message.to_string())
    }

    /// unsupported returns an error for something the engine cannot do yet.
    pub fn unsupported(what: impl fmt::Display) -> PgError {
        PgError::new(code::FEATURE_NOT_SUPPORTED, format!("{what} is not yet supported"))
    }
}

impl fmt::Display for PgError {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}: {} (SQLSTATE {})", self.severity, self.message, self.code)
    }
}

impl std::error::Error for PgError {}

impl From<doltdb::database::Error> for PgError {
    fn from(err: doltdb::database::Error) -> PgError {
        PgError::internal(err)
    }
}

impl From<store::Error> for PgError {
    fn from(err: store::Error) -> PgError {
        PgError::internal(err)
    }
}

/// Result is an engine result.
pub type Result<T> = std::result::Result<T, PgError>;
