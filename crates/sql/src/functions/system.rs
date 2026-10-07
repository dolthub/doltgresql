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
use crate::auth::Object;
use crate::error::{PgError, Result, code};
use crate::oid::{BOOL, FLOAT8, INT4, INT8, INTERVAL, NAME, OID, TEXT, TIMESTAMPTZ};
use crate::query::Ctx;
use crate::types::Value;

/// f declares a strict session or server function.
const fn f(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}

/// SCHEMA_PRIVILEGES and DATABASE_PRIVILEGES map the privilege names that has_schema_privilege and
/// has_database_privilege accept to their letters.
const SCHEMA_PRIVILEGES: &[(&str, &str)] = &[("CREATE", "C"), ("USAGE", "U")];
const DATABASE_PRIVILEGES: &[(&str, &str)] = &[("CREATE", "C"), ("TEMPORARY", "T"), ("TEMP", "T"), ("CONNECT", "c")];

/// FUNCTIONS are the session and server functions.
pub const FUNCTIONS: &[Function] = &[
    Function {
        name: "pg_get_indexdef",
        args: &[crate::oid::OID],
        ret: TEXT,
        strict: true,
        variadic: false,
        implementation: pg_get_indexdef,
    },
    Function {
        name: "pg_get_indexdef",
        args: &[crate::oid::OID, crate::oid::INT4, BOOL],
        ret: TEXT,
        strict: true,
        variadic: false,
        implementation: pg_get_indexdef,
    },
    Function {
        name: "pg_get_expr",
        args: &[super::ANY, crate::oid::OID],
        ret: TEXT,
        strict: true,
        variadic: false,
        implementation: pg_get_expr,
    },
    Function {
        name: "pg_get_expr",
        args: &[super::ANY, crate::oid::OID, BOOL],
        ret: TEXT,
        strict: true,
        variadic: false,
        implementation: pg_get_expr,
    },
    Function {
        name: "pg_get_constraintdef",
        args: &[crate::oid::OID],
        ret: TEXT,
        strict: true,
        variadic: false,
        implementation: pg_get_constraintdef,
    },
    Function {
        name: "pg_get_constraintdef",
        args: &[crate::oid::OID, BOOL],
        ret: TEXT,
        strict: true,
        variadic: false,
        implementation: pg_get_constraintdef,
    },
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
    f("pg_sleep", &[FLOAT8], crate::routines::VOID, pg_sleep),
    f("pg_sleep_for", &[INTERVAL], crate::routines::VOID, pg_sleep),
    f("pg_sleep_until", &[TIMESTAMPTZ], crate::routines::VOID, pg_sleep),
    f("pg_backend_pid", &[], INT4, pg_backend_pid),
    f("txid_current", &[], INT8, txid_current),
    f("pg_postmaster_start_time", &[], TIMESTAMPTZ, pg_postmaster_start_time),
    f("pg_is_in_recovery", &[], BOOL, pg_is_in_recovery),
    f("pg_is_wal_replay_paused", &[], BOOL, pg_is_wal_replay_paused),
    f("has_schema_privilege", &[NAME, TEXT, TEXT], BOOL, has_schema_privilege),
    f("has_schema_privilege", &[NAME, OID, TEXT], BOOL, has_schema_privilege),
    f("has_schema_privilege", &[OID, TEXT, TEXT], BOOL, has_schema_privilege),
    f("has_schema_privilege", &[OID, OID, TEXT], BOOL, has_schema_privilege),
    f("has_schema_privilege", &[TEXT, TEXT], BOOL, has_schema_privilege),
    f("has_schema_privilege", &[OID, TEXT], BOOL, has_schema_privilege),
    f("has_database_privilege", &[NAME, TEXT, TEXT], BOOL, has_database_privilege),
    f("has_database_privilege", &[NAME, OID, TEXT], BOOL, has_database_privilege),
    f("has_database_privilege", &[OID, TEXT, TEXT], BOOL, has_database_privilege),
    f("has_database_privilege", &[OID, OID, TEXT], BOOL, has_database_privilege),
    f("has_database_privilege", &[TEXT, TEXT], BOOL, has_database_privilege),
    f("has_database_privilege", &[OID, TEXT], BOOL, has_database_privilege),
    f("pg_get_functiondef", &[OID], TEXT, pg_get_functiondef),
    f("pg_get_function_arguments", &[OID], TEXT, pg_get_function_arguments),
    f("pg_get_function_identity_arguments", &[OID], TEXT, pg_get_function_identity_arguments),
    f("pg_get_function_result", &[OID], TEXT, pg_get_function_result),
    f("pg_get_function_sqlbody", &[OID], TEXT, pg_get_function_sqlbody),
    f("pg_get_triggerdef", &[OID], TEXT, pg_get_triggerdef),
    f("pg_get_triggerdef", &[OID, BOOL], TEXT, pg_get_triggerdef),
    f("pg_get_viewdef", &[OID], TEXT, pg_get_viewdef),
    f("pg_get_viewdef", &[OID, BOOL], TEXT, pg_get_viewdef),
    f("pg_get_viewdef", &[OID, INT4], TEXT, pg_get_viewdef),
    f("pg_get_viewdef", &[TEXT], TEXT, pg_get_viewdef),
    f("pg_get_viewdef", &[TEXT, BOOL], TEXT, pg_get_viewdef),
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

/// oid_arg returns the OID an argument holds.
fn oid_arg(value: &Value) -> u32 {
    match value {
        Value::Oid(oid) => *oid,
        Value::Reg(reg) => reg.oid,
        Value::Int4(i) => *i as u32,
        Value::Int8(i) => *i as u32,
        _ => 0,
    }
}

/// pg_get_expr prints an expression over a relation's columns, prettily when asked.
fn pg_get_expr(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let expression = args[0].output().unwrap_or_default();
    let pretty = matches!(args.get(2), Some(Value::Bool(true)));
    Ok(Value::Text(ctx.expression_definition(&expression, oid_arg(&args[1]), pretty)?))
}

/// pg_get_constraintdef prints a constraint's definition, prettily when asked, or returns NULL for an OID that no
/// constraint has.
fn pg_get_constraintdef(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let pretty = matches!(args.get(1), Some(Value::Bool(true)));
    Ok(ctx.constraint_definition_of(oid_arg(&args[0]), pretty)?.map_or(Value::Null, Value::Text))
}

/// pg_get_indexdef returns an index's CREATE INDEX statement, prettily when asked, or one of its columns for a positive
/// column number, or NULL for an OID that no index has.
fn pg_get_indexdef(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let index = match &args[0] {
        Value::Oid(oid) => *oid,
        Value::Reg(reg) => reg.oid,
        _ => 0,
    };
    let column = match args.get(1) {
        Some(Value::Int4(column)) => *column,
        _ => 0,
    };
    let pretty = matches!(args.get(2), Some(Value::Bool(true)));
    Ok(ctx.index_definition_of(index, column, pretty)?.map_or(Value::Null, Value::Text))
}

/// pg_sleep waits for a number of seconds, for an interval, or until a time.
fn pg_sleep(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let seconds = match &args[0] {
        Value::Float8(f) => *f,
        Value::Interval(iv) => {
            let days = iv.months as i64 * 30 + iv.days as i64;
            (days * crate::datetime::USECS_PER_DAY + iv.micros) as f64 / 1e6
        }
        Value::TimestampTz(t) => (t - crate::datetime::clock()) as f64 / 1e6,
        _ => 0.0,
    };
    if seconds.is_finite() && seconds > 0.0 {
        std::thread::sleep(std::time::Duration::from_secs_f64(seconds));
    }
    Ok(Value::Text(String::new()))
}

/// pg_backend_pid returns the session's number, which the server also sends the client as its process ID.
fn pg_backend_pid(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Int4(ctx.session.id as i32))
}

/// txid_current returns the transaction's ID, which is always 0, since Doltgres does not number transactions.
fn txid_current(_: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Int8(0))
}

/// pg_postmaster_start_time returns when the server started.
fn pg_postmaster_start_time(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::TimestampTz(ctx.session.engine.started()))
}

/// pg_is_in_recovery reports whether the server is replaying a write-ahead log, which it never does.
fn pg_is_in_recovery(_: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Bool(false))
}

/// pg_is_wal_replay_paused fails as Postgres does outside recovery, which is where the server always is.
fn pg_is_wal_replay_paused(_: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Err(PgError {
        hint: Some("Recovery control functions can only be executed during recovery.".into()),
        ..PgError::new(code::OBJECT_NOT_IN_PREREQUISITE_STATE, "recovery is not in progress")
    })
}

/// has_schema_privilege reports whether a role, the current one by default, holds any of the comma-separated
/// privileges on a schema, or returns NULL for an OID that no schema has.
fn has_schema_privilege(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    has_privilege(ctx, args, SCHEMA_PRIVILEGES)
}

/// has_database_privilege reports whether a role, the current one by default, holds any of the comma-separated
/// privileges on a database, or returns NULL for an OID that no database has.
fn has_database_privilege(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    has_privilege(ctx, args, DATABASE_PRIVILEGES)
}

/// has_privilege runs has_schema_privilege or has_database_privilege, checking the role, then the object's name,
/// then the privileges, as Postgres does.
fn has_privilege(ctx: &mut Ctx<'_>, args: &[Value], names: &[(&str, &str)]) -> Result<Value> {
    let (role, object, privileges) = match args {
        [role, object, privileges] => (Some(role), object, privileges),
        [object, privileges] => (None, object, privileges),
        _ => return Ok(Value::Null),
    };
    let role = {
        let auth = ctx.auth()?;
        match role {
            None => auth.role(&ctx.session.role).map_or(0, |r| r.id),
            Some(Value::Text(name)) if name == "public" => auth.public_id(),
            Some(Value::Text(name)) => match auth.role(name) {
                Some(role) => role.id,
                None => {
                    return Err(PgError::new(code::UNDEFINED_OBJECT, format!("role \"{name}\" does not exist")));
                }
            },
            Some(other) => ctx.role_of_oid(oid_arg(other)).and_then(|name| auth.role(&name).map(|r| r.id)).unwrap_or(0),
        }
    };
    let schema = names == SCHEMA_PRIVILEGES;
    let object = match object {
        Value::Text(name) if schema => match ctx.namespaces().iter().any(|(n, _)| n == name) {
            true => Some(Object::Schema(name.clone())),
            false => {
                return Err(PgError::new(code::INVALID_SCHEMA_NAME, format!("schema \"{name}\" does not exist")));
            }
        },
        Value::Text(name) => match ctx.catalog_database_names().contains(name) {
            true => Some(Object::Database(name.clone())),
            false => {
                return Err(PgError::new(code::INVALID_CATALOG_NAME, format!("database \"{name}\" does not exist")));
            }
        },
        other if schema => {
            let oid = oid_arg(other);
            ctx.namespaces().into_iter().find(|(_, o)| *o == oid).map(|(n, _)| Object::Schema(n))
        }
        other => {
            let oid = oid_arg(other);
            ctx.catalog_database_names()
                .into_iter()
                .find(|n| crate::pgcatalog::snapshot::database_oid(n) == oid)
                .map(Object::Database)
        }
    };
    let mut wanted = Vec::new();
    for chunk in text(privileges).split(',') {
        let chunk = chunk.trim_matches(|c: char| c.is_ascii_whitespace());
        let upper = chunk.to_ascii_uppercase();
        let (name, option) = match upper.strip_suffix(" WITH GRANT OPTION") {
            Some(name) => (name, true),
            None => (upper.as_str(), false),
        };
        let Some((_, letter)) = names.iter().find(|(n, _)| *n == name) else {
            return Err(PgError::new(
                code::INVALID_PARAMETER_VALUE,
                format!("unrecognized privilege type: \"{chunk}\""),
            ));
        };
        wanted.push((*letter, option));
    }
    match object {
        Some(object) => Ok(Value::Bool(ctx.has_privilege(role, &object, &wanted)?)),
        None => Ok(Value::Null),
    }
}

/// pg_get_functiondef prints the CREATE OR REPLACE statement of a function or procedure, or returns NULL for an OID
/// that none has.
fn pg_get_functiondef(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    match ctx.routine_def(oid_arg(&args[0]))? {
        Some(def) => Ok(Value::Text(def.definition()?)),
        None => Ok(Value::Null),
    }
}

/// pg_get_function_arguments prints the parameters of a function or procedure with their defaults, or returns NULL
/// for an OID that none has.
fn pg_get_function_arguments(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(ctx.routine_def(oid_arg(&args[0]))?.map_or(Value::Null, |def| Value::Text(def.arguments(true))))
}

/// pg_get_function_identity_arguments prints the parameters that identify a function or procedure, or returns NULL
/// for an OID that none has.
fn pg_get_function_identity_arguments(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(ctx.routine_def(oid_arg(&args[0]))?.map_or(Value::Null, |def| Value::Text(def.arguments(false))))
}

/// pg_get_function_result prints the result type of a function, or returns NULL for a procedure or an OID that no
/// function has.
fn pg_get_function_result(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(ctx.routine_def(oid_arg(&args[0]))?.and_then(|def| def.result()).map_or(Value::Null, Value::Text))
}

/// pg_get_function_sqlbody prints the SQL-standard body of a function or procedure, or returns NULL for one without
/// such a body or an OID that none has.
fn pg_get_function_sqlbody(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(ctx.routine_def(oid_arg(&args[0]))?.and_then(|def| def.sql_body()).map_or(Value::Null, Value::Text))
}

/// pg_get_triggerdef prints a trigger's CREATE TRIGGER statement, prettily when asked, or returns NULL for an OID
/// that no trigger has.
fn pg_get_triggerdef(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let pretty = matches!(args.get(1), Some(Value::Bool(true)));
    Ok(ctx.trigger_definition_of(oid_arg(&args[0]), pretty)?.map_or(Value::Null, Value::Text))
}

/// pg_get_viewdef prints a view's query, prettily when asked or when given a column to wrap its lists after, where a
/// view of pg_catalog or information_schema prints as pg_views shows it, or returns NULL for a relation that is not
/// a view.
fn pg_get_viewdef(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let relation = match &args[0] {
        Value::Text(name) => match ctx.reg_value(Value::Text(name.clone()), crate::oid::REGCLASS)? {
            Value::Reg(reg) => reg.oid,
            _ => 0,
        },
        other => oid_arg(other),
    };
    let (pretty, wrap) = match args.get(1) {
        Some(Value::Bool(pretty)) => (*pretty, 0),
        Some(Value::Int4(wrap)) => (true, *wrap),
        _ => (false, 0),
    };
    let snapshot = ctx.snapshot()?;
    let view = snapshot.views.iter().find(|v| crate::pgcatalog::snapshot::view_oid(&v.schema, &v.name) == relation);
    match view {
        Some(view) => Ok(Value::Text(ctx.view_definition(&view.statement, pretty, wrap)?)),
        None => Ok(crate::pgcatalog::builtin_view_definition(relation).map_or(Value::Null, |d| Value::Text(d.into()))),
    }
}
