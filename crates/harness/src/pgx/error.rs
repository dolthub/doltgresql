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

use std::fmt;

use pgproto::ErrorFields;

/// PgError is an error that the server reported with an ErrorResponse.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct PgError {
    /// Every field of the ErrorResponse.
    pub fields: ErrorFields,
}

impl fmt::Display for PgError {
    /// fmt implements the interface Display, matching the text of pgx's PgError.
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(f, "{}: {} (SQLSTATE {})", self.fields.severity, self.fields.message, self.fields.code)
    }
}

/// Error is any error returned by the client.
#[derive(Clone, Debug, PartialEq, Eq)]
pub enum Error {
    /// The server reported an error.
    Pg(Box<PgError>),
    /// A connection could not be established. Each attempt contributes one line, as pgx joins them.
    Connect {
        /// The user that was connecting.
        user: String,
        /// The database that was requested.
        database: String,
        /// The error of each connection attempt, in order.
        attempts: Vec<ConnectAttemptError>,
    },
    /// Any other failure, such as an IO error or an unexpected message.
    Other(String),
}

/// ConnectAttemptError is the error of a single connection attempt.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct ConnectAttemptError {
    /// The address that was dialed.
    pub address: String,
    /// The host name that resolved to the address.
    pub host: String,
    /// What failed, such as "tls error" or "server error".
    pub stage: String,
    /// The server's error when the stage is a server error.
    pub pg_error: Option<Box<PgError>>,
    /// The error text when there is no server error.
    pub message: String,
}

impl Error {
    /// pg returns the error for an ErrorResponse with the given fields.
    pub fn pg(fields: ErrorFields) -> Error {
        Error::Pg(Box::new(PgError { fields }))
    }

    /// pg_error returns the server's error, including one that ended a connection attempt.
    pub fn pg_error(&self) -> Option<&PgError> {
        match self {
            Error::Pg(err) => Some(err.as_ref()),
            Error::Connect { attempts, .. } => attempts.iter().rev().find_map(|attempt| attempt.pg_error.as_deref()),
            Error::Other(_) => None,
        }
    }
}

impl fmt::Display for ConnectAttemptError {
    /// fmt implements the interface Display, matching the text of pgx's perDialConnectError.
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match &self.pg_error {
            Some(err) => write!(f, "{} ({}): {}: {}", self.address, self.host, self.stage, err),
            None => write!(f, "{} ({}): {}: {}", self.address, self.host, self.stage, self.message),
        }
    }
}

impl fmt::Display for Error {
    /// fmt implements the interface Display, matching the text of the corresponding pgx errors.
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Error::Pg(err) => write!(f, "{err}"),
            Error::Connect { user, database, attempts } => {
                let prefix = format!("failed to connect to `user={user} database={database}`:");
                let details = attempts.iter().map(ToString::to_string).collect::<Vec<_>>().join("\n");
                if details.contains('\n') {
                    write!(f, "{prefix}\n\t{}", details.replace('\n', "\n\t"))
                } else {
                    write!(f, "{prefix} {details}")
                }
            }
            Error::Other(message) => write!(f, "{message}"),
        }
    }
}

impl std::error::Error for Error {}

impl From<std::io::Error> for Error {
    /// from implements the interface From.
    fn from(err: std::io::Error) -> Error {
        Error::Other(err.to_string())
    }
}

impl From<pgproto::DecodeError> for Error {
    /// from implements the interface From.
    fn from(err: pgproto::DecodeError) -> Error {
        Error::Other(err.to_string())
    }
}
