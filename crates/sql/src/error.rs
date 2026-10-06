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
    pub const UNDEFINED_OBJECT: &str = "42704";
    pub const INVALID_TEXT_REPRESENTATION: &str = "22P02";
    pub const NUMERIC_VALUE_OUT_OF_RANGE: &str = "22003";
    pub const DIVISION_BY_ZERO: &str = "22012";
    pub const STRING_DATA_RIGHT_TRUNCATION: &str = "22001";
    pub const AMBIGUOUS_COLUMN: &str = "42702";
    pub const UNDEFINED_TABLE: &str = "42P01";
    pub const UNDEFINED_COLUMN: &str = "42703";
    pub const UNDEFINED_FUNCTION: &str = "42883";
    pub const DATATYPE_MISMATCH: &str = "42804";
    pub const CANNOT_COERCE: &str = "42846";
    pub const DUPLICATE_TABLE: &str = "42P07";
    pub const DUPLICATE_COLUMN: &str = "42701";
    pub const INVALID_TABLE_DEFINITION: &str = "42P16";
    pub const NOT_NULL_VIOLATION: &str = "23502";
    pub const UNIQUE_VIOLATION: &str = "23505";
    pub const CHECK_VIOLATION: &str = "23514";
    pub const DEPENDENT_OBJECTS_STILL_EXIST: &str = "2BP01";
    pub const DUPLICATE_SCHEMA: &str = "42P06";
    pub const ACTIVE_SQL_TRANSACTION: &str = "25001";
    pub const IN_FAILED_SQL_TRANSACTION: &str = "25P02";
    pub const NO_ACTIVE_SQL_TRANSACTION: &str = "25P01";
    pub const INVALID_SCHEMA_NAME: &str = "3F000";
    pub const SYNTAX_ERROR_OR_ACCESS_RULE: &str = "42000";
    pub const INVALID_COLUMN_REFERENCE: &str = "42P10";
    pub const GROUPING_ERROR: &str = "42803";
    pub const INVALID_ROW_COUNT_IN_LIMIT_CLAUSE: &str = "2201W";
    pub const INVALID_ROW_COUNT_IN_RESULT_OFFSET_CLAUSE: &str = "2201X";
    pub const SERIALIZATION_FAILURE: &str = "40001";
    pub const CANT_CHANGE_RUNTIME_PARAM: &str = "55P02";
    pub const PROGRAM_LIMIT_EXCEEDED: &str = "54000";
    pub const SUBSTRING_ERROR: &str = "22011";
    pub const AMBIGUOUS_FUNCTION: &str = "42725";
    pub const WRONG_OBJECT_TYPE: &str = "42809";
    pub const CARDINALITY_VIOLATION: &str = "21000";
    pub const DUPLICATE_ALIAS: &str = "42712";
    pub const INVALID_DATETIME_FORMAT: &str = "22007";
    pub const DATETIME_FIELD_OVERFLOW: &str = "22008";
    pub const INVALID_TIME_ZONE_DISPLACEMENT: &str = "22009";
    pub const INVALID_ARGUMENT_FOR_LOG: &str = "2201E";
    pub const INVALID_ARGUMENT_FOR_POWER: &str = "2201F";
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
    /// ERROR, FATAL, or PANIC, or NOTICE or WARNING for a notice.
    pub severity: &'static str,
    pub code: &'static str,
    pub message: String,
    pub detail: Option<String>,
    pub hint: Option<String>,
    /// The 1-based character position in the query that the error refers to.
    pub position: Option<u32>,
    /// The context and objects of the error, which few errors have.
    pub objects: Option<Box<ErrorObjects>>,
}

/// ErrorObjects are an error's context and the objects it is about.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct ErrorObjects {
    /// The context the error happened in, such as the PL/pgSQL statement running.
    pub where_: Option<String>,
    pub schema: Option<String>,
    pub table: Option<String>,
    pub column: Option<String>,
    pub data_type: Option<String>,
    pub constraint: Option<String>,
}

impl PgError {
    /// new returns an ERROR with the code and message.
    pub fn new(code: &'static str, message: impl Into<String>) -> PgError {
        PgError {
            severity: "ERROR",
            code,
            message: message.into(),
            detail: None,
            hint: None,
            position: None,
            objects: None,
        }
    }

    /// notice returns a NOTICE, which reports something without failing the statement.
    pub fn notice(code: &'static str, message: impl Into<String>) -> PgError {
        PgError { severity: "NOTICE", ..PgError::new(code, message) }
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
