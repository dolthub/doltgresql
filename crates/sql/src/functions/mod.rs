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

//! Built-in functions, and choosing among a function's overloads as Postgres does.

pub mod aggregate;
pub use array::value_type;
pub use json::OUT_COLUMNS as JSON_OUT_COLUMNS;
mod advisory;
mod array;
mod binary;
mod catalog;
pub mod datetime;
pub mod json;
mod math;
mod pattern;
mod series;
mod string;
mod system;
pub mod xml;

use std::collections::HashMap;
use std::sync::OnceLock;

use crate::cast::type_display;
use crate::catalog::builtin_type;
use crate::error::{PgError, Result, code};
use crate::expr::position;
use crate::oid;
use crate::query::Ctx;
use crate::types::Value;

/// Implementation computes a function's result from its arguments.
pub type Implementation = fn(&mut Ctx<'_>, &[Value]) -> Result<Value>;

/// Function is one overload of a built-in function.
#[derive(Debug)]
pub struct Function {
    pub name: &'static str,
    /// The parameter types, which may be the polymorphic types ANYELEMENT, ANYARRAY, and ANYNONARRAY.
    pub args: &'static [u32],
    /// The result type, which is the type a polymorphic parameter resolved to when it is polymorphic.
    pub ret: u32,
    /// Whether the function returns NULL, without running, when an argument is NULL.
    pub strict: bool,
    /// Whether the last parameter takes any number of arguments of its type.
    pub variadic: bool,
    pub implementation: Implementation,
}

/// The polymorphic pseudo-types.
pub const ANY: u32 = 2276;
pub const ANYARRAY: u32 = 2277;
pub const ANYELEMENT: u32 = 2283;
pub const ANYNONARRAY: u32 = 2776;

/// Registry indexes the built-in functions by name.
struct Registry {
    functions: Vec<&'static Function>,
    by_name: HashMap<&'static str, Vec<usize>>,
}

/// registry returns the built-in functions, indexing them on first use.
fn registry() -> &'static Registry {
    static REGISTRY: OnceLock<Registry> = OnceLock::new();
    REGISTRY.get_or_init(|| {
        let functions: Vec<&'static Function> = [
            system::FUNCTIONS,
            string::FUNCTIONS,
            math::FUNCTIONS,
            series::FUNCTIONS,
            datetime::FUNCTIONS,
            array::FUNCTIONS,
            pattern::FUNCTIONS,
            json::FUNCTIONS,
            binary::FUNCTIONS,
            xml::FUNCTIONS,
            catalog::FUNCTIONS,
            advisory::FUNCTIONS,
            crate::dolt::procedures::FUNCTIONS,
            crate::sequences::FUNCTIONS,
        ]
        .into_iter()
        .flatten()
        .collect();
        let mut by_name: HashMap<&'static str, Vec<usize>> = HashMap::new();
        for (i, f) in functions.iter().enumerate() {
            by_name.entry(f.name).or_default().push(i);
        }
        Registry { functions, by_name }
    })
}

/// function returns a built-in function by its index.
pub fn function(index: usize) -> &'static Function {
    registry().functions[index]
}

/// SET_RETURNING are the functions that return rows.
const SET_RETURNING: &[&str] = &[
    "generate_series",
    "generate_subscripts",
    "unnest",
    "regexp_matches",
    "regexp_split_to_table",
    "dolt_log",
    "dolt_diff_summary",
    "dolt_diff_stat",
    "dolt_preview_merge_conflicts_summary",
    "jsonb_object_keys",
    "json_object_keys",
    "jsonb_array_elements",
    "json_array_elements",
    "jsonb_array_elements_text",
    "json_array_elements_text",
    "jsonb_each",
    "json_each",
    "jsonb_each_text",
    "json_each_text",
];

/// returns_set reports whether a function returns rows.
pub fn returns_set(name: &str) -> bool {
    SET_RETURNING.contains(&name)
}

/// exists reports whether a function of the name exists.
pub fn exists(name: &str) -> bool {
    registry().by_name.contains_key(name)
}

/// is_array reports whether a type is an array type.
fn is_array(type_oid: u32) -> bool {
    crate::array::is_array_type(type_oid)
}

/// numeric_rank orders the numeric types by implicit promotion.
fn numeric_rank(type_oid: u32) -> Option<u8> {
    match type_oid {
        oid::INT2 => Some(0),
        oid::INT4 => Some(1),
        oid::INT8 => Some(2),
        oid::NUMERIC => Some(3),
        oid::FLOAT4 => Some(4),
        oid::FLOAT8 => Some(5),
        _ => None,
    }
}

/// is_string reports whether a type is in the string category.
fn is_string(type_oid: u32) -> bool {
    matches!(type_oid, oid::TEXT | oid::VARCHAR | oid::BPCHAR | oid::NAME)
}

/// implicitly_castable reports whether Postgres converts a value of one type to the other without being asked.
pub fn implicitly_castable(from: u32, to: u32) -> bool {
    from == to
        || from == oid::UNKNOWN
        || to == ANY
        || numeric_rank(from).zip(numeric_rank(to)).is_some_and(|(f, t)| f < t)
        || (is_string(from) && matches!(to, oid::TEXT | oid::VARCHAR | oid::BPCHAR | oid::NAME))
        || (to == ANYELEMENT && from != oid::UNKNOWN)
        || (to == ANYARRAY && is_array(from))
        || (to == ANYNONARRAY && !is_array(from))
        || (matches!(from, oid::BIT | oid::VARBIT) && matches!(to, oid::BIT | oid::VARBIT))
        || crate::casts::context(from, to) == Some(crate::casts::IMPLICIT)
        || crate::expr::implicit_datetime(from, to)
        || (to == oid::OID && (crate::cast::is_reg_type(from) || matches!(from, oid::INT2 | oid::INT4 | oid::INT8)))
        || (is_array(from)
            && is_array(to)
            && implicitly_castable(crate::expr::element_type(from), crate::expr::element_type(to)))
}

/// is_preferred reports whether a type is the preferred type of its category.
fn is_preferred(type_oid: u32) -> bool {
    builtin_type(type_oid).is_some_and(|t| t.definition.is_preferred)
}

/// Resolved is a chosen overload with the type each argument converts to and the result type.
pub struct Resolved {
    pub index: usize,
    pub arg_types: Vec<u32>,
    pub ret: u32,
}

/// parameter_types returns the parameter types an overload gives a number of arguments, repeating a variadic one.
fn parameter_types(f: &Function, count: usize) -> Option<Vec<u32>> {
    if f.variadic {
        let (last, fixed) = f.args.split_last()?;
        if count < fixed.len() {
            return None;
        }
        let mut types = fixed.to_vec();
        types.resize(count, *last);
        return Some(types);
    }
    (f.args.len() == count).then(|| f.args.to_vec())
}

/// resolve chooses the overload of a function for arguments of the types, as Postgres' function resolution does:
/// keep the overloads every argument converts to, prefer the most exact matches, then the most preferred types, and
/// finally, for untyped arguments, the string category.
pub fn resolve(name: &str, types: &[u32], location: i32) -> Result<Resolved> {
    let r = registry();
    let not_found = || PgError {
        position: position(location),
        hint: Some(
            "No function matches the given name and argument types. You might need to add explicit type casts.".into(),
        ),
        ..PgError::new(
            code::UNDEFINED_FUNCTION,
            format!(
                "function {name}({}) does not exist",
                types.iter().map(|&t| type_display(t)).collect::<Vec<_>>().join(", ")
            ),
        )
    };
    let indexes = r.by_name.get(name).ok_or_else(not_found)?;
    let candidates: Vec<(usize, Vec<u32>)> =
        indexes.iter().filter_map(|&i| Some((i, parameter_types(r.functions[i], types.len())?))).collect();
    let mut candidates = best_candidates(types, candidates);
    if candidates.is_empty() {
        return Err(not_found());
    }
    if candidates.len() > 1 {
        return Err(PgError {
            position: position(location),
            hint: Some("Could not choose a best candidate function. You might need to add explicit type casts.".into()),
            ..PgError::new(
                code::AMBIGUOUS_FUNCTION,
                format!(
                    "function {name}({}) is not unique",
                    types.iter().map(|&t| type_display(t)).collect::<Vec<_>>().join(", ")
                ),
            )
        });
    }
    let (index, params) = candidates.pop().expect("a candidate");
    let f = r.functions[index];
    // A polymorphic parameter takes the type of its argument, and the result follows it.
    let mut element = None;
    for (&p, &t) in params.iter().zip(types) {
        match p {
            ANYELEMENT | ANYNONARRAY => element = element.or(Some(t)),
            ANYARRAY => {
                element = element.or_else(|| match builtin_type(t) {
                    Some(b) => Some(b.elem),
                    None => is_array(t).then(|| crate::expr::element_type(t)),
                })
            }
            _ => {}
        }
    }
    let element = element.filter(|&e| e != oid::UNKNOWN).unwrap_or(oid::TEXT);
    let arg_types = params
        .iter()
        .zip(types)
        .map(|(&p, &t)| match p {
            ANYELEMENT | ANYNONARRAY => element,
            ANYARRAY if builtin_type(element).is_none() && crate::usertypes::get(element).is_some() => {
                crate::expr::array_of(element)
            }
            ANYARRAY => builtin_type(element).map_or(t, |b| b.array),
            ANY => t,
            _ => p,
        })
        .collect();
    let ret = match f.ret {
        ANYELEMENT | ANYNONARRAY => element,
        ANYARRAY if builtin_type(element).is_none() && crate::usertypes::get(element).is_some() => {
            crate::expr::array_of(element)
        }
        ANYARRAY => builtin_type(element).map_or(oid::TEXT, |b| b.array),
        other => other,
    };
    Ok(Resolved { index, arg_types, ret })
}

/// overload_types returns the parameter types each built-in overload of the name gives a number of arguments.
pub fn overload_types(name: &str, count: usize) -> Vec<Vec<u32>> {
    let r = registry();
    r.by_name
        .get(name)
        .map(|indexes| indexes.iter().filter_map(|&i| parameter_types(r.functions[i], count)).collect())
        .unwrap_or_default()
}

/// best_candidates keeps the overloads, each given with its parameter types, that a call with arguments of the types
/// may choose, as Postgres' function resolution does: those every argument converts to, then those with the most
/// exact matches, then the most preferred types, and finally, for untyped arguments, the string category. More than
/// one remaining means the call is ambiguous.
pub fn best_candidates<C>(types: &[u32], candidates: Vec<(C, Vec<u32>)>) -> Vec<(C, Vec<u32>)> {
    let mut candidates: Vec<(C, Vec<u32>)> = candidates
        .into_iter()
        .filter(|(_, params)| params.iter().zip(types).all(|(&p, &t)| implicitly_castable(t, p)))
        .collect();
    let keep_best = |candidates: &mut Vec<(C, Vec<u32>)>, score: &dyn Fn(&[u32]) -> usize| {
        let best = candidates.iter().map(|(_, p)| score(p)).max().unwrap_or(0);
        candidates.retain(|(_, p)| score(p) == best);
    };
    keep_best(&mut candidates, &|params| params.iter().zip(types).filter(|(p, t)| *p == *t).count());
    keep_best(&mut candidates, &|params| {
        params.iter().zip(types).filter(|(p, t)| *p != *t && is_preferred(**p)).count()
    });
    if candidates.len() > 1 {
        keep_best(&mut candidates, &|params| {
            params.iter().zip(types).filter(|(p, t)| **t == oid::UNKNOWN && is_string(**p)).count()
        });
    }
    candidates
}

/// call runs a function on arguments already converted to its parameter types.
pub fn call(ctx: &mut Ctx<'_>, index: usize, args: &[Value]) -> Result<Value> {
    let f = function(index);
    if f.strict && args.iter().any(Value::is_null) {
        return Ok(Value::Null);
    }
    (f.implementation)(ctx, args)
}

/// call_set runs a set-returning function, returning its rows' values.
pub fn call_set(ctx: &mut Ctx<'_>, index: usize, args: &[Value]) -> Result<Vec<Value>> {
    let f = function(index);
    if f.strict && args.iter().any(Value::is_null) {
        return Ok(Vec::new());
    }
    match (f.implementation)(ctx, args)? {
        Value::Set(values) => Ok(values),
        value => Ok(vec![value]),
    }
}

/// text returns a text argument.
fn text(value: &Value) -> &str {
    match value {
        Value::Text(s) => s,
        _ => "",
    }
}
