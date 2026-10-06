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

//! Values and their text and binary wire formats.

use crate::error::{PgError, Result, code};
use crate::oid;

/// TEXT_FORMAT and BINARY_FORMAT are the wire format codes.
pub const TEXT_FORMAT: i16 = 0;
pub const BINARY_FORMAT: i16 = 1;

/// Value is a SQL value.
#[derive(Clone, Debug, PartialEq)]
pub enum Value {
    Null,
    Bool(bool),
    Int2(i16),
    Int4(i32),
    Int8(i64),
    /// A string of the text types, and the value of an untyped literal.
    Text(String),
}

impl Value {
    /// encode returns the value in the format for a column of the type, or None for NULL.
    pub fn encode(&self, type_oid: u32, format: i16) -> Option<Vec<u8>> {
        if format == BINARY_FORMAT {
            return self.send(type_oid);
        }
        self.output()
    }

    /// output returns the value's text format, or None for NULL.
    pub fn output(&self) -> Option<Vec<u8>> {
        Some(match self {
            Value::Null => return None,
            Value::Bool(b) => {
                if *b {
                    b"t".to_vec()
                } else {
                    b"f".to_vec()
                }
            }
            Value::Int2(i) => i.to_string().into_bytes(),
            Value::Int4(i) => i.to_string().into_bytes(),
            Value::Int8(i) => i.to_string().into_bytes(),
            Value::Text(s) => s.clone().into_bytes(),
        })
    }

    /// send returns the value's binary format, or None for NULL.
    fn send(&self, type_oid: u32) -> Option<Vec<u8>> {
        Some(match self {
            Value::Null => return None,
            Value::Bool(b) => vec![*b as u8],
            Value::Int2(i) => i.to_be_bytes().to_vec(),
            Value::Int4(i) => i.to_be_bytes().to_vec(),
            Value::Int8(i) => i.to_be_bytes().to_vec(),
            Value::Text(_) if type_oid == oid::UNKNOWN => return self.output(),
            Value::Text(s) => s.clone().into_bytes(),
        })
    }

    /// decode returns a parameter value sent in the format for the type, where a zero type OID means unspecified.
    pub fn decode(type_oid: u32, format: i16, bytes: Option<&[u8]>) -> Result<Value> {
        let Some(bytes) = bytes else { return Ok(Value::Null) };
        let invalid =
            || PgError::new(code::INVALID_BINARY_REPRESENTATION, "incorrect binary data format in bind parameter");
        if format == BINARY_FORMAT {
            return match type_oid {
                oid::BOOL => Ok(Value::Bool(*bytes.first().ok_or_else(invalid)? != 0)),
                oid::INT2 => Ok(Value::Int2(i16::from_be_bytes(bytes.try_into().map_err(|_| invalid())?))),
                oid::INT4 => Ok(Value::Int4(i32::from_be_bytes(bytes.try_into().map_err(|_| invalid())?))),
                oid::INT8 => Ok(Value::Int8(i64::from_be_bytes(bytes.try_into().map_err(|_| invalid())?))),
                oid::TEXT | oid::UNKNOWN | 0 => {
                    Ok(Value::Text(String::from_utf8(bytes.to_vec()).map_err(|_| invalid())?))
                }
                _ => Err(PgError::unsupported(format!("binary parameters of type {type_oid}"))),
            };
        }
        let text = std::str::from_utf8(bytes).map_err(|_| {
            PgError::new(code::CHARACTER_NOT_IN_REPERTOIRE, "invalid byte sequence for encoding \"UTF8\"")
        })?;
        match type_oid {
            oid::TEXT | oid::UNKNOWN | 0 => Ok(Value::Text(text.to_string())),
            _ => Err(PgError::unsupported(format!("text parameters of type {type_oid}"))),
        }
    }
}
