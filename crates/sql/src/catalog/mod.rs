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

//! The catalog: built-in types, resolving the type names that statements use, and table definitions.

pub mod id;
pub mod oids;
pub mod table;

use std::collections::HashMap;
use std::sync::OnceLock;

use objects::SerializedType;
use prolly::val::encoding;

use crate::error::{PgError, Result, code};

/// BUILTIN_TYPES lists every built-in type as `oid name hex`, where the hex is the type as Doltgres serializes it.
const BUILTIN_TYPES: &str = include_str!("builtin_types.txt");

/// BuiltinType is a built-in type.
#[derive(Debug)]
pub struct BuiltinType {
    pub oid: u32,
    pub name: &'static str,
    pub definition: SerializedType,
    /// The OID of the element type of an array type, or 0.
    pub elem: u32,
    /// The OID of the array type whose elements are this type, or 0.
    pub array: u32,
}

/// Builtins indexes the built-in types.
struct Builtins {
    types: Vec<BuiltinType>,
    /// The position of each type by OID, or `usize::MAX` for an OID that no built-in type has.
    by_oid: Vec<usize>,
    by_name: HashMap<&'static str, usize>,
    by_id: HashMap<Vec<u8>, usize>,
}

/// builtins returns the built-in types, decoding them on first use.
fn builtins() -> &'static Builtins {
    static BUILTINS: OnceLock<Builtins> = OnceLock::new();
    BUILTINS.get_or_init(|| {
        let mut types = Vec::new();
        let mut by_id = HashMap::new();
        for line in BUILTIN_TYPES.lines() {
            let mut fields = line.split(' ');
            let (Some(oid), Some(name), Some(hex)) = (fields.next(), fields.next(), fields.next()) else { continue };
            let bytes: Vec<u8> =
                (0..hex.len()).step_by(2).map(|i| u8::from_str_radix(&hex[i..i + 2], 16).unwrap()).collect();
            let definition = SerializedType::deserialize(&bytes).expect("a valid built-in type");
            let oid: u32 = oid.parse().expect("a built-in type OID");
            by_id.insert(definition.id.clone(), oid);
            types.push(BuiltinType { oid, name, definition, elem: 0, array: 0 });
        }
        for t in &mut types {
            t.elem = by_id.get(&t.definition.elem).copied().unwrap_or(0);
            t.array = by_id.get(&t.definition.array).copied().unwrap_or(0);
        }
        let mut by_oid = vec![usize::MAX; types.iter().map(|t| t.oid as usize + 1).max().unwrap_or(0)];
        for (i, t) in types.iter().enumerate() {
            by_oid[t.oid as usize] = i;
        }
        let by_name = types.iter().enumerate().map(|(i, t)| (t.name, i)).collect();
        let by_id = types.iter().enumerate().map(|(i, t)| (t.definition.id.clone(), i)).collect();
        Builtins { types, by_oid, by_name, by_id }
    })
}

/// builtin_types returns every built-in type.
pub fn builtin_types() -> &'static [BuiltinType] {
    &builtins().types
}

/// builtin_type returns a built-in type by OID.
pub fn builtin_type(oid: u32) -> Option<&'static BuiltinType> {
    let b = builtins();
    b.by_oid.get(oid as usize).and_then(|&i| b.types.get(i))
}

/// builtin_type_named returns a built-in type by its name in pg_catalog.
pub fn builtin_type_named(name: &str) -> Option<&'static BuiltinType> {
    let b = builtins();
    b.by_name.get(name).map(|&i| &b.types[i])
}

/// builtin_type_by_id returns a built-in type by its internal ID.
pub fn builtin_type_by_id(id: &[u8]) -> Option<&'static BuiltinType> {
    let b = builtins();
    b.by_id.get(id).map(|&i| &b.types[i])
}

/// ColumnType is the type of a column: a type OID and its type modifier, which is -1 without one.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct ColumnType {
    pub oid: u32,
    pub modifier: i32,
}

impl ColumnType {
    /// serialized returns the column type as a Dolt schema stores it: `extended_` and the hex of the serialized type.
    pub fn serialized(&self) -> Result<String> {
        let mut definition = match builtin_type(self.oid) {
            Some(t) => t.definition.clone(),
            None => crate::usertypes::get(self.oid)
                .map(|t| t.definition.clone())
                .ok_or_else(|| PgError::internal(format!("unknown type OID {}", self.oid)))?,
        };
        definition.att_typ_mod = self.modifier;
        let hex: String = definition.serialize().iter().map(|b| format!("{b:02x}")).collect();
        Ok(format!("extended_{hex}"))
    }

    /// encoding returns the field encoding of the column type in tuples, as Doltgres chooses it.
    pub fn encoding(&self) -> u8 {
        let Some(t) = builtin_type(self.oid) else {
            return match crate::usertypes::get(self.oid) {
                Some(t) if t.definition.typ_length > 0 => encoding::EXTENDED,
                _ => encoding::EXTENDED_ADAPTIVE,
            };
        };
        match t.name {
            "int2" => encoding::INT16,
            "int4" => encoding::INT32,
            "int8" => encoding::INT64,
            "float4" => encoding::FLOAT32,
            "float8" => encoding::FLOAT64,
            "numeric" => encoding::DECIMAL,
            "bytea" => encoding::BYTES_ADAPTIVE,
            "json" | "jsonb" => encoding::JSON_ADAPTIVE,
            "xid" => encoding::UINT32,
            "varchar" if self.modifier == -1 => encoding::STRING_ADAPTIVE,
            "varchar" | "name" | "char" => encoding::STRING,
            "bpchar" | "text" | "xml" | "jsonpath" => encoding::STRING_ADAPTIVE,
            _ if t.definition.typ_length > 0 => encoding::EXTENDED,
            _ => encoding::EXTENDED_ADAPTIVE,
        }
    }

    /// name returns the type's name in pg_catalog, or the name of a user-defined type.
    pub fn name(&self) -> std::borrow::Cow<'static, str> {
        match builtin_type(self.oid) {
            Some(t) => t.name.into(),
            None => crate::usertypes::get(self.oid).map_or("unknown".into(), |t| t.name.clone().into()),
        }
    }
}

/// resolve_type returns the column type that a type name and its modifiers denote, where `position` is the 1-based
/// position of the name for errors.
pub fn resolve_type(names: &[String], modifiers: &[String], array: bool, position: Option<u32>) -> Result<ColumnType> {
    let name = match names {
        [name] => name.as_str(),
        [schema, name] if schema == "pg_catalog" => name.as_str(),
        _ => "",
    };
    let not_found = || PgError {
        position,
        ..PgError::new(code::UNDEFINED_OBJECT, format!("type \"{}\" does not exist", names.join(".")))
    };
    let user_type = match names {
        [name] => crate::usertypes::lookup(None, name),
        [schema, name] | [_, schema, name] => crate::usertypes::lookup(Some(schema), name),
        _ => None,
    };
    if let Some(user_type) = user_type {
        if matches!(user_type.kind, crate::usertypes::Kind::Shell) {
            return Err(PgError {
                position,
                ..PgError::new(code::UNDEFINED_OBJECT, format!("type \"{}\" is only a shell", names.join(".")))
            });
        }
        let oid = match (array, user_type.is_array()) {
            (false, _) => user_type.oid,
            (true, false) => user_type.array,
            (true, true) => 0,
        };
        if oid == 0 {
            return Err(not_found());
        }
        let modifier = match &user_type.kind {
            crate::usertypes::Kind::Domain(domain) if !array => domain.base.modifier,
            crate::usertypes::Kind::Base(definition) if !modifiers.is_empty() => {
                (definition.typmod_in)(modifiers).map_err(|err| PgError { position, ..err })?
            }
            _ => -1,
        };
        return Ok(ColumnType { oid, modifier });
    }
    let t = builtin_type_named(name).ok_or_else(not_found)?;
    let numbers = modifiers
        .iter()
        .map(|m| m.parse().map_err(|_| PgError { position, ..crate::cast::invalid_syntax(crate::oid::INT4, m) }))
        .collect::<Result<Vec<i32>>>()?;
    let modifier = type_modifier(t.name, &numbers, position)?;
    let oid = if array { t.array } else { t.oid };
    if oid == 0 {
        return Err(not_found());
    }
    Ok(ColumnType { oid, modifier })
}

/// type_modifier computes a type's modifier from the numbers written after its name, as the type's typmodin does.
fn type_modifier(name: &str, modifiers: &[i32], position: Option<u32>) -> Result<i32> {
    let error = |message: String| PgError { position, ..PgError::new(code::INVALID_PARAMETER_VALUE, message) };
    match (name, modifiers) {
        (_, []) => Ok(-1),
        ("varchar" | "bpchar", [length]) => {
            let kind = if name == "varchar" { "varchar" } else { "char" };
            if *length < 1 {
                return Err(error(format!("length for type {kind} must be at least 1")));
            }
            if *length > 10_485_760 {
                return Err(error(format!("length for type {kind} cannot exceed 10485760")));
            }
            Ok(length + 4)
        }
        ("numeric", [precision, rest @ ..]) if rest.len() <= 1 => {
            let scale = rest.first().copied().unwrap_or(0);
            if !(1..=1000).contains(precision) {
                return Err(error(format!("NUMERIC precision {precision} must be between 1 and 1000")));
            }
            if !(-1000..=1000).contains(&scale) {
                return Err(error(format!("NUMERIC scale {scale} must be between -1000 and 1000")));
            }
            Ok(((precision << 16) | (scale & 0x7ff)) + 4)
        }
        ("timestamp" | "timestamptz" | "time" | "timetz", [precision]) => {
            if *precision < 0 {
                let kind = name.to_uppercase();
                return Err(error(format!("{kind}({precision}) precision must not be negative")));
            }
            Ok((*precision).min(6))
        }
        ("interval", [range, rest @ ..]) if rest.len() <= 1 => {
            if !crate::datetime::INTERVAL_RANGES.iter().any(|(r, _)| r == range) {
                return Err(error("invalid INTERVAL type modifier".into()));
            }
            let precision = match rest.first() {
                None if *range == crate::datetime::INTERVAL_FULL_RANGE => return Ok(-1),
                None => crate::datetime::INTERVAL_FULL_PRECISION,
                Some(p) if *p < 0 => return Err(error(format!("INTERVAL({p}) precision must not be negative"))),
                Some(p) => (*p).min(6),
            };
            Ok((range << 16) | precision)
        }
        ("bit" | "varbit", [length]) => {
            if *length < 1 {
                let kind = if name == "bit" { "bit" } else { "bit varying" };
                return Err(error(format!("length for type {kind} must be at least 1")));
            }
            Ok(*length)
        }
        _ => Err(PgError {
            position,
            ..PgError::new(code::SYNTAX_ERROR, format!("type modifier is not allowed for type \"{name}\""))
        }),
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn builtin_types_have_their_arrays() {
        let int8 = builtin_type(20).unwrap();
        assert_eq!((int8.name, int8.array, int8.definition.typ_length), ("int8", 1016, 8));
        assert_eq!(builtin_type(1016).unwrap().elem, 20);
        assert_eq!(builtin_types_count(), 113);
    }

    #[test]
    fn column_types_serialize_as_go_does() {
        let int8 = ColumnType { oid: 20, modifier: -1 };
        assert!(int8.serialized().unwrap().ends_with("3823020a0470675f636174616c6f67696e7438000006626967696e74"));
        assert_eq!(int8.encoding(), encoding::INT64);
        let varchar = resolve_type(&["pg_catalog".into(), "varchar".into()], &["10".into()], false, None).unwrap();
        assert_eq!((varchar.modifier, varchar.encoding()), (14, encoding::STRING));
        let err = resolve_type(&["nope".into()], &[], false, Some(5)).unwrap_err();
        assert_eq!((err.code, err.message.as_str(), err.position), ("42704", "type \"nope\" does not exist", Some(5)));
    }

    /// builtin_types_count returns the number of built-in types.
    fn builtin_types_count() -> usize {
        builtins().types.len()
    }
}
