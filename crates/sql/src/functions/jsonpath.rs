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

//! The SQL/JSON path functions and operators over jsonb: jsonb_path_exists, jsonb_path_match, jsonb_path_query and
//! its array and first-item forms, each with a `_tz` form, and the `@?` and `@@` operators.

use std::collections::HashMap;

use super::Function;
use crate::error::{PgError, Result, code};
use crate::json::Json;
use crate::jsonpath::exec::{self, Item, Options};
use crate::oid::{BOOL, JSONB, JSONPATH};
use crate::query::Ctx;
use crate::types::Value;

/// f declares a strict function with fixed parameters.
const fn f(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}

/// path_functions declares each path function, whose calls the binder completes with the defaults of its vars and
/// silent parameters, followed by the operators.
macro_rules! path_functions {
    ($(($name:literal, $ret:expr, $implementation:ident, $tz:literal)),* $(,)?) => {
        &[
            $(
                f($name, &[JSONB, JSONPATH, JSONB, BOOL], $ret, |ctx, args| $implementation(ctx, args, $tz)),
            )*
            f("@?", &[JSONB, JSONPATH], BOOL, |ctx, args| path_exists(ctx, &silent(args), false)),
            f("@@", &[JSONB, JSONPATH], BOOL, |ctx, args| path_match(ctx, &silent(args), false)),
        ]
    };
}

pub const FUNCTIONS: &[Function] = path_functions![
    ("jsonb_path_exists", BOOL, path_exists, false),
    ("jsonb_path_exists_tz", BOOL, path_exists, true),
    ("jsonb_path_match", BOOL, path_match, false),
    ("jsonb_path_match_tz", BOOL, path_match, true),
    ("jsonb_path_query", JSONB, path_query, false),
    ("jsonb_path_query_tz", JSONB, path_query, true),
    ("jsonb_path_query_array", JSONB, path_query_array, false),
    ("jsonb_path_query_array_tz", JSONB, path_query_array, true),
    ("jsonb_path_query_first", JSONB, path_query_first, false),
    ("jsonb_path_query_first_tz", JSONB, path_query_first, true),
];

/// silent returns an operator's arguments as the arguments of the function it runs, with empty vars and errors
/// suppressed, as Postgres' `_opr` functions do.
fn silent(args: &[Value]) -> Vec<Value> {
    vec![args[0].clone(), args[1].clone(), Value::Jsonb(Box::new(Json::Object(Vec::new()))), Value::Bool(true)]
}

/// Call is a path function's parsed arguments.
struct Call {
    document: Item,
    path: crate::jsonpath::JsonPath,
    vars: HashMap<String, Item>,
    silent: bool,
}

impl Call {
    /// new reads a path function's document, path, vars, and silent arguments.
    fn new(args: &[Value]) -> Result<Call> {
        let document = match &args[0] {
            Value::Jsonb(json) => Item::Json((**json).clone()),
            other => return Err(PgError::internal(format!("a jsonb document, not {other:?}"))),
        };
        let path = crate::jsonpath::parse(&args[1].output().unwrap_or_default())?;
        let vars = match &args[2] {
            Value::Jsonb(json) => match &**json {
                Json::Object(members) => members.iter().map(|(k, v)| (k.clone(), Item::Json(v.clone()))).collect(),
                _ => {
                    return Err(PgError {
                        detail: Some(
                            "Jsonpath parameters should be encoded as key-value pairs of \"vars\" object.".into(),
                        ),
                        ..PgError::new(code::INVALID_PARAMETER_VALUE, "\"vars\" argument is not an object")
                    });
                }
            },
            _ => HashMap::new(),
        };
        Ok(Call { document, path, vars, silent: matches!(args[3], Value::Bool(true)) })
    }

    /// items returns the items that the path finds, with the ones found before an error that silent suppressed.
    fn items(&self, use_tz: bool) -> Result<Vec<Item>> {
        let mut items = Vec::new();
        let options = Options { vars: &self.vars, throw: !self.silent, use_tz };
        exec::query_into(&self.path, &self.document, &options, &mut items)?;
        Ok(items)
    }
}

/// path_exists reports whether a path finds an item, as jsonb_path_exists does, or returns NULL after a suppressed
/// error.
fn path_exists(_: &mut Ctx<'_>, args: &[Value], use_tz: bool) -> Result<Value> {
    let call = Call::new(args)?;
    let options = Options { vars: &call.vars, throw: !call.silent, use_tz };
    Ok(exec::exists(&call.path, &call.document, &options)?.map_or(Value::Null, Value::Bool))
}

/// path_match returns the single boolean that a path predicate yields, as jsonb_path_match does.
fn path_match(_: &mut Ctx<'_>, args: &[Value], use_tz: bool) -> Result<Value> {
    let call = Call::new(args)?;
    match call.items(use_tz)?.as_slice() {
        [Item::Json(Json::Bool(b))] => return Ok(Value::Bool(*b)),
        [Item::Json(Json::Null)] => return Ok(Value::Null),
        _ => {}
    }
    if call.silent {
        return Ok(Value::Null);
    }
    Err(PgError::new(code::SINGLETON_SQL_JSON_ITEM_REQUIRED, "single boolean result is expected"))
}

/// path_query returns a row for each item that a path finds.
fn path_query(_: &mut Ctx<'_>, args: &[Value], use_tz: bool) -> Result<Value> {
    let items = Call::new(args)?.items(use_tz)?;
    Ok(Value::Set(items.iter().map(|item| Value::Jsonb(Box::new(item.to_json()))).collect()))
}

/// path_query_array returns the items that a path finds as one array.
fn path_query_array(_: &mut Ctx<'_>, args: &[Value], use_tz: bool) -> Result<Value> {
    let items = Call::new(args)?.items(use_tz)?;
    Ok(Value::Jsonb(Box::new(Json::Array(items.iter().map(Item::to_json).collect()))))
}

/// path_query_first returns the first item that a path finds, or NULL when it finds none.
fn path_query_first(_: &mut Ctx<'_>, args: &[Value], use_tz: bool) -> Result<Value> {
    let items = Call::new(args)?.items(use_tz)?;
    Ok(items.first().map_or(Value::Null, |item| Value::Jsonb(Box::new(item.to_json()))))
}
