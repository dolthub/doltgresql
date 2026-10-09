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
pub use catalog::OUT_COLUMNS as CATALOG_OUT_COLUMNS;
pub use json::OUT_COLUMNS as JSON_OUT_COLUMNS;
mod advisory;
mod array;
mod binary;
pub(crate) mod catalog;
pub mod datetime;
pub mod json;
mod jsonpath;
mod math;
pub(crate) mod pattern;
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
            jsonpath::FUNCTIONS,
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
    "__doltgres_foreach_slice",
    "regexp_matches",
    "regexp_split_to_table",
    "dolt_log",
    "dolt_diff_summary",
    "dolt_diff_stat",
    "dolt_preview_merge_conflicts_summary",
    "jsonb_path_query",
    "jsonb_path_query_tz",
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
    "pg_partition_ancestors",
];

/// returns_set reports whether a function returns rows.
/// PATH_PARAMETERS are the parameters of the SQL/JSON path functions, with the defaults of the last two.
const PATH_PARAMETERS: (&[&str], &[&str]) = (&["target", "path", "vars", "silent"], &["'{}'::jsonb", "false"]);

/// PARAMETERS names the parameters of the built-in functions that calls may pass by name or leave out, with the
/// defaults of their trailing parameters, as pg_proc's proargnames and proargdefaults hold them.
pub const PARAMETERS: &[(&str, &[&str], &[&str])] = &[
    ("jsonb_path_exists", PATH_PARAMETERS.0, PATH_PARAMETERS.1),
    ("jsonb_path_exists_tz", PATH_PARAMETERS.0, PATH_PARAMETERS.1),
    ("jsonb_path_match", PATH_PARAMETERS.0, PATH_PARAMETERS.1),
    ("jsonb_path_match_tz", PATH_PARAMETERS.0, PATH_PARAMETERS.1),
    ("jsonb_path_query", PATH_PARAMETERS.0, PATH_PARAMETERS.1),
    ("jsonb_path_query_tz", PATH_PARAMETERS.0, PATH_PARAMETERS.1),
    ("jsonb_path_query_array", PATH_PARAMETERS.0, PATH_PARAMETERS.1),
    ("jsonb_path_query_array_tz", PATH_PARAMETERS.0, PATH_PARAMETERS.1),
    ("jsonb_path_query_first", PATH_PARAMETERS.0, PATH_PARAMETERS.1),
    ("jsonb_path_query_first_tz", PATH_PARAMETERS.0, PATH_PARAMETERS.1),
    (
        "make_interval",
        &["years", "months", "weeks", "days", "hours", "mins", "secs"],
        &["0", "0", "0", "0", "0", "0", "0.0"],
    ),
];

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
        || (crate::cast::is_reg_type(to) && matches!(from, oid::OID | oid::INT2 | oid::INT4 | oid::INT8))
        || (to == oid::REGCLASS && matches!(from, oid::TEXT | oid::VARCHAR))
        || (is_array(from)
            && is_array(to)
            && implicitly_castable(crate::expr::element_type(from), crate::expr::element_type(to)))
}

/// is_preferred reports whether a type is the preferred type of its category.
pub(crate) fn is_preferred(type_oid: u32) -> bool {
    builtin_type(type_oid).is_some_and(|t| t.definition.is_preferred)
}

/// Resolved is a chosen overload with the type each argument converts to and the result type.
pub struct Resolved {
    pub index: usize,
    pub arg_types: Vec<u32>,
    pub ret: u32,
}

/// parameter_types returns the parameter types an overload gives a number of arguments, repeating a variadic one,
/// which takes at least one argument unless it is the only parameter and its type is not `any`.
fn parameter_types(f: &Function, count: usize) -> Option<Vec<u32>> {
    if f.variadic {
        let (last, fixed) = f.args.split_last()?;
        if count < fixed.len() || (count == fixed.len() && (!fixed.is_empty() || *last == ANY)) {
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
    let candidates: Vec<(usize, Vec<u32>)> = indexes
        .iter()
        .filter_map(|&i| Some((i, parameter_types(r.functions[i], types.len())?)))
        .filter(|(_, params)| consistent(params, types))
        .collect();
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
    // A polymorphic parameter takes the common type of its arguments, and the result follows it.
    let mut element: Option<u32> = None;
    for (&p, &t) in params.iter().zip(types).filter(|(_, t)| **t != oid::UNKNOWN) {
        let implied = match p {
            ANYELEMENT | ANYNONARRAY => Some(t),
            ANYARRAY => match builtin_type(t) {
                Some(b) => Some(b.elem),
                None => is_array(t).then(|| crate::expr::element_type(t)),
            },
            _ => None,
        };
        element = match (element, implied) {
            (Some(current), Some(implied)) if current != implied && implicitly_castable(current, implied) => {
                Some(implied)
            }
            (current, implied) => current.or(implied),
        };
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

/// consistent reports whether the arguments of an overload's polymorphic parameters, read as their domains' base
/// types, can share one element type, as Postgres requires.
fn consistent(params: &[u32], types: &[u32]) -> bool {
    let mut element = None;
    for (&p, &t) in params.iter().zip(types).filter(|(_, t)| **t != oid::UNKNOWN) {
        let t = crate::usertypes::base_type(crate::expr::typ(t)).oid;
        let implied = match p {
            ANYELEMENT | ANYNONARRAY => t,
            ANYARRAY => crate::expr::element_type(t),
            _ => continue,
        };
        let first = *element.get_or_insert(implied);
        if !implicitly_castable(first, implied) && !implicitly_castable(implied, first) {
            return false;
        }
    }
    true
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
        params
            .iter()
            .zip(types)
            .filter(|(p, t)| **t != oid::UNKNOWN && *p != *t && is_preferred(**p) && category(**p) == category(**t))
            .count()
    });
    if candidates.len() > 1 {
        keep_unknown_categories(types, &mut candidates);
    }
    candidates
}

/// category returns the category letter of a type.
fn category(type_oid: u32) -> Option<u8> {
    builtin_type(type_oid).and_then(|t| t.definition.typ_category.first().copied())
}

/// keep_unknown_categories keeps the candidates whose parameters for untyped arguments are all in the category chosen
/// for each, which is the string category when any candidate takes a string there and otherwise the one category
/// every candidate takes, and are the preferred type of that category when some candidate takes it, as Postgres'
/// func_select_candidate does. It keeps every candidate when some untyped argument has no such category.
fn keep_unknown_categories<C>(types: &[u32], candidates: &mut Vec<(C, Vec<u32>)>) {
    let mut chosen = Vec::new();
    for (i, _) in types.iter().enumerate().filter(|(_, t)| **t == oid::UNKNOWN) {
        let categories: Vec<Option<u8>> = candidates.iter().map(|(_, p)| category(p[i])).collect();
        let selected = if categories.contains(&Some(b'S')) {
            Some(b'S')
        } else if categories.windows(2).all(|w| w[0] == w[1]) {
            categories[0]
        } else {
            return;
        };
        let preferred = candidates.iter().any(|(_, p)| category(p[i]) == selected && is_preferred(p[i]));
        chosen.push((i, selected, preferred));
    }
    candidates.retain(|(_, p)| {
        chosen.iter().all(|&(i, selected, preferred)| category(p[i]) == selected && (!preferred || is_preferred(p[i])))
    });
}

/// call runs a function on arguments already converted to its parameter types.
pub fn call(ctx: &mut Ctx<'_>, index: usize, args: &[Value]) -> Result<Value> {
    let f = function(index);
    if f.strict && args.iter().any(Value::is_null) {
        return Ok(Value::Null);
    }
    implement(ctx, f, args)
}

/// implement runs a function's implementation, without the session's temporary tables for Dolt's procedures, since
/// version control never sees them.
fn implement(ctx: &mut Ctx<'_>, f: &Function, args: &[Value]) -> Result<Value> {
    match f.name.starts_with("dolt_") {
        true => ctx.without_temp(|ctx| (f.implementation)(ctx, args)),
        false => (f.implementation)(ctx, args),
    }
}

/// call_set runs a set-returning function, returning its rows' values.
pub fn call_set(ctx: &mut Ctx<'_>, index: usize, args: &[Value]) -> Result<Vec<Value>> {
    let f = function(index);
    if f.strict && args.iter().any(Value::is_null) {
        return Ok(Vec::new());
    }
    match implement(ctx, f, args)? {
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
