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

//! Values as tuple fields, in the encodings Doltgres stores each type with.

use prolly::val::encoding;

use crate::error::{PgError, Result};
use crate::types::Value;

/// encode_field returns a value as a tuple field of the encoding, or None for NULL.
pub fn encode_field(value: &Value, field_encoding: u8) -> Result<Option<Vec<u8>>> {
    Ok(Some(match (value, field_encoding) {
        (Value::Null, _) => return Ok(None),
        (Value::Int2(i), encoding::INT16) => i.to_le_bytes().to_vec(),
        (Value::Int4(i), encoding::INT32) => i.to_le_bytes().to_vec(),
        (Value::Int8(i), encoding::INT64) => i.to_le_bytes().to_vec(),
        (Value::Float4(f), encoding::FLOAT32) => f.to_le_bytes().to_vec(),
        (Value::Float8(f), encoding::FLOAT64) => f.to_le_bytes().to_vec(),
        (Value::Text(s), encoding::STRING) => {
            let mut field = s.clone().into_bytes();
            field.push(0);
            field
        }
        (value, field_encoding) => {
            return Err(PgError::unsupported(format!("storing {value:?} with encoding {field_encoding}")));
        }
    }))
}

/// decode_field reads a value from a tuple field of the encoding, where None is NULL.
pub fn decode_field(field: Option<&[u8]>, field_encoding: u8) -> Result<Value> {
    let Some(field) = field else { return Ok(Value::Null) };
    let corrupt = || PgError::internal(format!("a field of encoding {field_encoding} has {} bytes", field.len()));
    Ok(match field_encoding {
        encoding::INT16 => Value::Int2(i16::from_le_bytes(field.try_into().map_err(|_| corrupt())?)),
        encoding::INT32 => Value::Int4(i32::from_le_bytes(field.try_into().map_err(|_| corrupt())?)),
        encoding::INT64 => Value::Int8(i64::from_le_bytes(field.try_into().map_err(|_| corrupt())?)),
        encoding::FLOAT32 => Value::Float4(f32::from_le_bytes(field.try_into().map_err(|_| corrupt())?)),
        encoding::FLOAT64 => Value::Float8(f64::from_le_bytes(field.try_into().map_err(|_| corrupt())?)),
        encoding::STRING => {
            let bytes = field.strip_suffix(&[0]).ok_or_else(corrupt)?;
            Value::Text(String::from_utf8(bytes.to_vec()).map_err(|_| corrupt())?)
        }
        _ => return Err(PgError::unsupported(format!("reading fields of encoding {field_encoding}"))),
    })
}
