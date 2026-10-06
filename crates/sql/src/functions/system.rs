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

//! Functions about the session and the server.

use super::{Function, text};
use crate::error::Result;
use crate::oid::{BOOL, NAME, TEXT};
use crate::query::Ctx;
use crate::types::Value;

/// FUNCTIONS are the session and server functions.
pub const FUNCTIONS: &[Function] = &[
    Function {
        name: "format_type",
        args: &[crate::oid::OID, crate::oid::INT4],
        ret: TEXT,
        strict: false,
        variadic: false,
        implementation: format_type,
    },
    Function {
        name: "current_setting",
        args: &[TEXT],
        ret: TEXT,
        strict: true,
        variadic: false,
        implementation: current_setting,
    },
    Function {
        name: "current_setting",
        args: &[TEXT, BOOL],
        ret: TEXT,
        strict: true,
        variadic: false,
        implementation: current_setting_missing_ok,
    },
    Function {
        name: "set_config",
        args: &[TEXT, TEXT, BOOL],
        ret: TEXT,
        strict: false,
        variadic: false,
        implementation: set_config,
    },
    Function { name: "version", args: &[], ret: TEXT, strict: true, variadic: false, implementation: version },
    Function {
        name: "current_database",
        args: &[],
        ret: NAME,
        strict: true,
        variadic: false,
        implementation: current_database,
    },
    Function {
        name: "current_schema",
        args: &[],
        ret: NAME,
        strict: true,
        variadic: false,
        implementation: current_schema,
    },
    Function {
        name: "current_schemas",
        args: &[BOOL],
        ret: 1003,
        strict: true,
        variadic: false,
        implementation: current_schemas,
    },
    Function {
        name: "current_user",
        args: &[],
        ret: NAME,
        strict: true,
        variadic: false,
        implementation: current_user,
    },
    Function {
        name: "session_user",
        args: &[],
        ret: NAME,
        strict: true,
        variadic: false,
        implementation: session_user,
    },
];

/// current_schemas returns the schemas of the search path that exist, with the ones searched implicitly when asked.
fn current_schemas(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let implicit = args[0] == Value::Bool(true);
    let path = if implicit { ctx.effective_search_path() } else { ctx.session.search_path() };
    let existing = ctx.schema_names();
    let schemas: Vec<Value> =
        path.into_iter().filter(|s| existing.contains(s) || (implicit && s == "pg_catalog")).map(Value::Text).collect();
    Ok(Value::Array(Box::new(crate::array::Array::one_dimensional(NAME, schemas))))
}

/// current_user returns the current role.
fn current_user(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Text(ctx.session.role.clone()))
}

/// session_user returns the session's user.
fn session_user(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Text(ctx.session.user.clone()))
}

/// current_setting returns a parameter's value as SHOW prints it.
fn current_setting(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Text(ctx.session.settings.show(text(&args[0]))?))
}

/// current_setting_missing_ok returns a parameter's value, or NULL for an unknown parameter when asked.
fn current_setting_missing_ok(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    match ctx.session.settings.show(text(&args[0])) {
        Ok(value) => Ok(Value::Text(value)),
        Err(_) if args[1] == Value::Bool(true) => Ok(Value::Null),
        Err(err) => Err(err),
    }
}

/// set_config sets a parameter, for the session or the transaction, and returns its new value.
fn set_config(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let name = text(&args[0]);
    let value = match &args[1] {
        Value::Null => None,
        value => Some(text(value).to_string()),
    };
    let local = args[2] == Value::Bool(true);
    let in_transaction = ctx.session.explicit;
    match name.to_ascii_lowercase().as_str() {
        "role" => {
            let role = value.filter(|v| v != "none");
            ctx.set_role(role.as_deref())?;
            ctx.session.settings.set_raw("role", role, local, in_transaction);
        }
        "session_authorization" => {
            let user = value.unwrap_or_default();
            ctx.set_session_authorization(&user)?;
            ctx.session.settings.set_raw("session_authorization", Some(user), local, in_transaction);
            ctx.session.settings.set_raw("role", None, local, in_transaction);
        }
        _ => ctx.session.settings.set(name, value.as_deref(), local, in_transaction)?,
    }
    ctx.session.sync_identity();
    Ok(Value::Text(ctx.session.settings.show(name)?))
}

/// version returns the server's version string.
fn version(_: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Text(format!("PostgreSQL {}", crate::SERVER_VERSION)))
}

/// current_database returns the session's database, with its branch when the session named one.
fn current_database(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Text(ctx.session.display.clone()))
}

/// current_schema returns the first schema of the search path that exists.
fn current_schema(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(ctx.creation_schema().map(Value::Text).unwrap_or(Value::Null))
}

/// format_type returns a type's name with a type modifier, or "???" for an unknown type.
fn format_type(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let type_oid = match &args[0] {
        Value::Null => return Ok(Value::Null),
        Value::Oid(oid) => *oid,
        Value::Reg(reg) => reg.oid,
        _ => 0,
    };
    let modifier = match &args[1] {
        Value::Int4(m) => Some(*m),
        _ => None,
    };
    if type_oid == 0 {
        return Ok(Value::Text("-".into()));
    }
    Ok(Value::Text(crate::cast::format_type(type_oid, modifier).unwrap_or_else(|| "???".into())))
}
