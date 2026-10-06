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

/// Error is a failure to read or write storage.
#[derive(Debug)]
pub enum Error {
    /// The file system failed.
    Io(std::io::Error),
    /// Stored data does not have the expected format.
    Corrupt(String),
}

/// Result is a storage result.
pub type Result<T> = std::result::Result<T, Error>;

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Error::Io(err) => write!(f, "{err}"),
            Error::Corrupt(message) => write!(f, "{message}"),
        }
    }
}

impl std::error::Error for Error {}

impl From<std::io::Error> for Error {
    fn from(err: std::io::Error) -> Error {
        Error::Io(err)
    }
}

/// corrupt returns a Corrupt error.
pub(crate) fn corrupt(message: impl Into<String>) -> Error {
    Error::Corrupt(message.into())
}
