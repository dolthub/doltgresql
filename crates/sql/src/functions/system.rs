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
use crate::oid::{BOOL, FLOAT8, INT2, INT4, INT8, INTERVAL, NAME, NUMERIC, OID, TEXT, TIMESTAMPTZ};
use crate::query::Ctx;
use crate::types::Value;

/// f declares a strict session or server function.
const fn f(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}

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
        name: "enum_in",
        args: &[crate::oid::CSTRING, crate::oid::OID],
        ret: crate::oid::ANYENUM,
        strict: true,
        variadic: false,
        implementation: enum_in,
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
    f("pg_my_temp_schema", &[], OID, pg_my_temp_schema),
    f("pg_is_other_temp_schema", &[OID], BOOL, pg_is_other_temp_schema),
    f("txid_current", &[], INT8, txid_current),
    f("pg_postmaster_start_time", &[], TIMESTAMPTZ, pg_postmaster_start_time),
    f("pg_is_in_recovery", &[], BOOL, pg_is_in_recovery),
    f("pg_is_wal_replay_paused", &[], BOOL, pg_is_wal_replay_paused),
    f("pg_size_pretty", &[INT8], TEXT, pg_size_pretty),
    f("pg_size_pretty", &[NUMERIC], TEXT, pg_size_pretty),
    f("pg_trigger_depth", &[], INT4, pg_trigger_depth),
    f("pg_database_size", &[NAME], INT8, pg_database_size),
    f("pg_database_size", &[OID], INT8, pg_database_size),
    f("has_table_privilege", &[NAME, TEXT, TEXT], BOOL, has_table_privilege),
    f("has_table_privilege", &[NAME, OID, TEXT], BOOL, has_table_privilege),
    f("has_table_privilege", &[OID, TEXT, TEXT], BOOL, has_table_privilege),
    f("has_table_privilege", &[OID, OID, TEXT], BOOL, has_table_privilege),
    f("has_table_privilege", &[TEXT, TEXT], BOOL, has_table_privilege),
    f("has_table_privilege", &[OID, TEXT], BOOL, has_table_privilege),
    f("has_any_column_privilege", &[NAME, TEXT, TEXT], BOOL, has_any_column_privilege),
    f("has_any_column_privilege", &[NAME, OID, TEXT], BOOL, has_any_column_privilege),
    f("has_any_column_privilege", &[OID, TEXT, TEXT], BOOL, has_any_column_privilege),
    f("has_any_column_privilege", &[OID, OID, TEXT], BOOL, has_any_column_privilege),
    f("has_any_column_privilege", &[TEXT, TEXT], BOOL, has_any_column_privilege),
    f("has_any_column_privilege", &[OID, TEXT], BOOL, has_any_column_privilege),
    f("has_sequence_privilege", &[NAME, TEXT, TEXT], BOOL, has_sequence_privilege),
    f("has_sequence_privilege", &[NAME, OID, TEXT], BOOL, has_sequence_privilege),
    f("has_sequence_privilege", &[OID, TEXT, TEXT], BOOL, has_sequence_privilege),
    f("has_sequence_privilege", &[OID, OID, TEXT], BOOL, has_sequence_privilege),
    f("has_sequence_privilege", &[TEXT, TEXT], BOOL, has_sequence_privilege),
    f("has_sequence_privilege", &[OID, TEXT], BOOL, has_sequence_privilege),
    f("has_function_privilege", &[NAME, TEXT, TEXT], BOOL, has_function_privilege),
    f("has_function_privilege", &[NAME, OID, TEXT], BOOL, has_function_privilege),
    f("has_function_privilege", &[OID, TEXT, TEXT], BOOL, has_function_privilege),
    f("has_function_privilege", &[OID, OID, TEXT], BOOL, has_function_privilege),
    f("has_function_privilege", &[TEXT, TEXT], BOOL, has_function_privilege),
    f("has_function_privilege", &[OID, TEXT], BOOL, has_function_privilege),
    f("has_type_privilege", &[NAME, TEXT, TEXT], BOOL, has_type_privilege),
    f("has_type_privilege", &[NAME, OID, TEXT], BOOL, has_type_privilege),
    f("has_type_privilege", &[OID, TEXT, TEXT], BOOL, has_type_privilege),
    f("has_type_privilege", &[OID, OID, TEXT], BOOL, has_type_privilege),
    f("has_type_privilege", &[TEXT, TEXT], BOOL, has_type_privilege),
    f("has_type_privilege", &[OID, TEXT], BOOL, has_type_privilege),
    f("has_language_privilege", &[NAME, TEXT, TEXT], BOOL, has_language_privilege),
    f("has_language_privilege", &[NAME, OID, TEXT], BOOL, has_language_privilege),
    f("has_language_privilege", &[OID, TEXT, TEXT], BOOL, has_language_privilege),
    f("has_language_privilege", &[OID, OID, TEXT], BOOL, has_language_privilege),
    f("has_language_privilege", &[TEXT, TEXT], BOOL, has_language_privilege),
    f("has_language_privilege", &[OID, TEXT], BOOL, has_language_privilege),
    f("has_tablespace_privilege", &[NAME, TEXT, TEXT], BOOL, has_tablespace_privilege),
    f("has_tablespace_privilege", &[NAME, OID, TEXT], BOOL, has_tablespace_privilege),
    f("has_tablespace_privilege", &[OID, TEXT, TEXT], BOOL, has_tablespace_privilege),
    f("has_tablespace_privilege", &[OID, OID, TEXT], BOOL, has_tablespace_privilege),
    f("has_tablespace_privilege", &[TEXT, TEXT], BOOL, has_tablespace_privilege),
    f("has_tablespace_privilege", &[OID, TEXT], BOOL, has_tablespace_privilege),
    f("has_server_privilege", &[NAME, TEXT, TEXT], BOOL, has_server_privilege),
    f("has_server_privilege", &[NAME, OID, TEXT], BOOL, has_server_privilege),
    f("has_server_privilege", &[OID, TEXT, TEXT], BOOL, has_server_privilege),
    f("has_server_privilege", &[OID, OID, TEXT], BOOL, has_server_privilege),
    f("has_server_privilege", &[TEXT, TEXT], BOOL, has_server_privilege),
    f("has_server_privilege", &[OID, TEXT], BOOL, has_server_privilege),
    f("has_foreign_data_wrapper_privilege", &[NAME, TEXT, TEXT], BOOL, has_foreign_data_wrapper_privilege),
    f("has_foreign_data_wrapper_privilege", &[NAME, OID, TEXT], BOOL, has_foreign_data_wrapper_privilege),
    f("has_foreign_data_wrapper_privilege", &[OID, TEXT, TEXT], BOOL, has_foreign_data_wrapper_privilege),
    f("has_foreign_data_wrapper_privilege", &[OID, OID, TEXT], BOOL, has_foreign_data_wrapper_privilege),
    f("has_foreign_data_wrapper_privilege", &[TEXT, TEXT], BOOL, has_foreign_data_wrapper_privilege),
    f("has_foreign_data_wrapper_privilege", &[OID, TEXT], BOOL, has_foreign_data_wrapper_privilege),
    f("has_column_privilege", &[NAME, TEXT, TEXT, TEXT], BOOL, has_column_privilege),
    f("has_column_privilege", &[NAME, TEXT, INT2, TEXT], BOOL, has_column_privilege),
    f("has_column_privilege", &[NAME, OID, TEXT, TEXT], BOOL, has_column_privilege),
    f("has_column_privilege", &[NAME, OID, INT2, TEXT], BOOL, has_column_privilege),
    f("has_column_privilege", &[OID, TEXT, TEXT, TEXT], BOOL, has_column_privilege),
    f("has_column_privilege", &[OID, TEXT, INT2, TEXT], BOOL, has_column_privilege),
    f("has_column_privilege", &[OID, OID, TEXT, TEXT], BOOL, has_column_privilege),
    f("has_column_privilege", &[OID, OID, INT2, TEXT], BOOL, has_column_privilege),
    f("has_column_privilege", &[TEXT, TEXT, TEXT], BOOL, has_column_privilege),
    f("has_column_privilege", &[TEXT, INT2, TEXT], BOOL, has_column_privilege),
    f("has_column_privilege", &[OID, TEXT, TEXT], BOOL, has_column_privilege),
    f("has_column_privilege", &[OID, INT2, TEXT], BOOL, has_column_privilege),
    f("has_parameter_privilege", &[NAME, TEXT, TEXT], BOOL, has_parameter_privilege),
    f("has_parameter_privilege", &[OID, TEXT, TEXT], BOOL, has_parameter_privilege),
    f("has_parameter_privilege", &[TEXT, TEXT], BOOL, has_parameter_privilege),
    f("pg_has_role", &[NAME, NAME, TEXT], BOOL, pg_has_role),
    f("pg_has_role", &[NAME, OID, TEXT], BOOL, pg_has_role),
    f("pg_has_role", &[OID, NAME, TEXT], BOOL, pg_has_role),
    f("pg_has_role", &[OID, OID, TEXT], BOOL, pg_has_role),
    f("pg_has_role", &[NAME, TEXT], BOOL, pg_has_role),
    f("pg_has_role", &[OID, TEXT], BOOL, pg_has_role),
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
    f("pg_get_ruledef", &[OID], TEXT, pg_get_ruledef),
    f("pg_get_ruledef", &[OID, BOOL], TEXT, pg_get_ruledef),
    f("pg_get_viewdef", &[OID], TEXT, pg_get_viewdef),
    f("pg_get_viewdef", &[OID, BOOL], TEXT, pg_get_viewdef),
    f("pg_get_viewdef", &[OID, INT4], TEXT, pg_get_viewdef),
    f("pg_get_viewdef", &[TEXT], TEXT, pg_get_viewdef),
    f("pg_get_viewdef", &[TEXT, BOOL], TEXT, pg_get_viewdef),
    f("load_file", &[TEXT], TEXT, load_file),
    f("pg_get_statisticsobjdef_columns", &[OID], TEXT, pg_get_statisticsobjdef_columns),
];

/// current_schemas returns the schemas of the search path that exist, with the ones searched implicitly when asked.
fn current_schemas(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let implicit = args[0] == Value::Bool(true);
    let path = if implicit { ctx.effective_search_path() } else { ctx.session.explicit_search_path() };
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

/// current_schema returns the first schema of the search path that exists, making the temporary schema when the path
/// names it first.
fn current_schema(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    if ctx.session.temp_first() {
        return ctx.temp_schema().map(Value::Text);
    }
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

/// enum_in reads a label as a value of the enum type with the OID, failing as Postgres does for an OID that names no
/// enum type.
fn enum_in(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let type_oid = oid_arg(&args[1]);
    match crate::usertypes::get(type_oid) {
        Some(t) if matches!(t.kind, crate::usertypes::Kind::Enum(_)) => {
            crate::cast::input(&args[0].output().unwrap_or_default(), type_oid)
        }
        _ => Err(PgError::new(code::INTERNAL_ERROR, format!("cache lookup failed for type {type_oid}"))),
    }
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
/// load_file returns the contents of a file relative to the server's working directory, or NULL when it is missing, as
/// go-mysql-server's LOAD_FILE does for Doltgres.
fn load_file(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    match std::fs::read(text(&args[0])) {
        Ok(bytes) => Ok(Value::Text(String::from_utf8_lossy(&bytes).into_owned())),
        Err(err) if err.kind() == std::io::ErrorKind::NotFound => Ok(Value::Null),
        Err(err) => Err(PgError::internal(err.to_string())),
    }
}

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

/// pg_my_temp_schema returns the OID of the session's temporary schema, or 0 when it has none.
fn pg_my_temp_schema(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Oid(match ctx.session.temp.contains_key(&ctx.session.database) {
        true => crate::pgcatalog::snapshot::namespace_oid(&ctx.session.temp_schema()),
        false => 0,
    }))
}

/// pg_is_other_temp_schema reports whether a schema is another session's temporary schema, which is never true, since
/// sessions never see each other's temporary schemas.
fn pg_is_other_temp_schema(_: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Bool(false))
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

/// Kind is the kind of object that one of the has_*_privilege functions asks about.
#[derive(Clone, Copy, PartialEq, Eq)]
enum Kind {
    Database,
    Schema,
    Table,
    Column,
    AnyColumn,
    Sequence,
    Function,
    Type,
    Language,
    Tablespace,
    Server,
    Wrapper,
    Parameter,
}

impl Kind {
    /// privileges returns the privilege names that the kind's function accepts, each with its letter.
    fn privileges(self) -> &'static [(&'static str, &'static str)] {
        match self {
            Kind::Database => &[("CREATE", "C"), ("TEMPORARY", "T"), ("TEMP", "T"), ("CONNECT", "c")],
            Kind::Schema => &[("CREATE", "C"), ("USAGE", "U")],
            Kind::Table => &[
                ("SELECT", "r"),
                ("INSERT", "a"),
                ("UPDATE", "w"),
                ("DELETE", "d"),
                ("TRUNCATE", "D"),
                ("REFERENCES", "x"),
                ("TRIGGER", "t"),
            ],
            Kind::Column | Kind::AnyColumn => &[("SELECT", "r"), ("INSERT", "a"), ("UPDATE", "w"), ("REFERENCES", "x")],
            Kind::Sequence => &[("USAGE", "U"), ("SELECT", "r"), ("UPDATE", "w")],
            Kind::Function => &[("EXECUTE", "X")],
            Kind::Type | Kind::Language | Kind::Server | Kind::Wrapper => &[("USAGE", "U")],
            Kind::Tablespace => &[("CREATE", "C")],
            Kind::Parameter => &[("SET", "s"), ("ALTER SYSTEM", "A")],
        }
    }
}

/// Target is what a has_*_privilege function's object name or OID resolves to: an object with privileges, a
/// privilege that everyone or only superusers hold, or nothing, for an OID that no object has.
enum Target {
    Object(Object),
    Everyone,
    Superusers,
    Missing,
}

/// has_schema_privilege reports whether a role, the current one by default, holds any of the comma-separated
/// privileges on a schema, or returns NULL for an OID that no schema has.
fn has_schema_privilege(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    has_privilege(ctx, args, Kind::Schema)
}

/// has_database_privilege reports whether a role, the current one by default, holds any of the comma-separated
/// privileges on a database, or returns NULL for an OID that no database has.
fn has_database_privilege(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    has_privilege(ctx, args, Kind::Database)
}

/// has_table_privilege reports whether a role holds any of the privileges on a table or view.
fn has_table_privilege(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    has_privilege(ctx, args, Kind::Table)
}

/// has_column_privilege reports whether a role holds any of the privileges on a column, which a table's privileges
/// grant.
fn has_column_privilege(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    has_privilege(ctx, args, Kind::Column)
}

/// has_any_column_privilege reports whether a role holds any of the privileges on some column of a table.
fn has_any_column_privilege(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    has_privilege(ctx, args, Kind::AnyColumn)
}

/// has_sequence_privilege reports whether a role holds any of the privileges on a sequence.
fn has_sequence_privilege(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    has_privilege(ctx, args, Kind::Sequence)
}

/// has_function_privilege reports whether a role may execute a function, which everyone may.
fn has_function_privilege(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    has_privilege(ctx, args, Kind::Function)
}

/// has_type_privilege reports whether a role may use a type, which everyone may.
fn has_type_privilege(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    has_privilege(ctx, args, Kind::Type)
}

/// has_language_privilege reports whether a role may use a language, which everyone may.
fn has_language_privilege(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    has_privilege(ctx, args, Kind::Language)
}

/// has_tablespace_privilege reports whether a role may create objects in a tablespace, which only superusers may.
fn has_tablespace_privilege(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    has_privilege(ctx, args, Kind::Tablespace)
}

/// has_server_privilege fails for any foreign server, since Doltgres has none.
fn has_server_privilege(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    has_privilege(ctx, args, Kind::Server)
}

/// has_foreign_data_wrapper_privilege fails for any foreign-data wrapper, since Doltgres has none.
fn has_foreign_data_wrapper_privilege(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    has_privilege(ctx, args, Kind::Wrapper)
}

/// has_parameter_privilege reports whether a role holds a privilege on a configuration parameter, which only
/// superusers do, since Doltgres has no grants on parameters.
fn has_parameter_privilege(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    has_privilege(ctx, args, Kind::Parameter)
}

/// role_id returns the ID of the role that a has_*_privilege function names, the current one when it names none.
fn role_id(ctx: &mut Ctx<'_>, role: Option<&Value>) -> Result<u64> {
    let auth = ctx.auth()?;
    Ok(match role {
        None => auth.role(&ctx.session.role).map_or(0, |r| r.id),
        Some(Value::Text(name)) if name == "public" => auth.public_id(),
        Some(Value::Text(name)) => match auth.role(name) {
            Some(role) => role.id,
            None => return Err(PgError::new(code::UNDEFINED_OBJECT, format!("role \"{name}\" does not exist"))),
        },
        Some(other) => ctx.role_of_oid(oid_arg(other)).and_then(|name| auth.role(&name).map(|r| r.id)).unwrap_or(0),
    })
}

/// relation resolves a relation's name or OID to its schema and name, failing as Postgres does for a name that no
/// relation has.
fn relation(ctx: &mut Ctx<'_>, value: &Value) -> Result<Option<(String, String)>> {
    let oid = match value {
        Value::Text(name) => oid_arg(&ctx.reg_value(Value::Text(name.clone()), crate::oid::REGCLASS)?),
        other => oid_arg(other),
    };
    ctx.relation_of_oid(oid)
}

/// target resolves the object that a has_*_privilege function names, given its arguments after the role.
fn target(ctx: &mut Ctx<'_>, kind: Kind, args: &[Value]) -> Result<Target> {
    let missing =
        |what: &str, name: &str| PgError::new(code::UNDEFINED_OBJECT, format!("{what} \"{name}\" does not exist"));
    Ok(match (kind, &args[0]) {
        (Kind::Schema, Value::Text(name)) => match ctx.namespaces().iter().any(|(n, _)| n == name) {
            true => Target::Object(Object::Schema(name.clone())),
            false => {
                return Err(PgError::new(code::INVALID_SCHEMA_NAME, format!("schema \"{name}\" does not exist")));
            }
        },
        (Kind::Schema, other) => {
            let oid = oid_arg(other);
            ctx.namespaces()
                .into_iter()
                .find(|(_, o)| *o == oid)
                .map_or(Target::Missing, |(n, _)| Target::Object(Object::Schema(n)))
        }
        (Kind::Database, Value::Text(name)) => match ctx.catalog_database_names().contains(name) {
            true => Target::Object(Object::Database(name.clone())),
            false => {
                return Err(PgError::new(code::INVALID_CATALOG_NAME, format!("database \"{name}\" does not exist")));
            }
        },
        (Kind::Database, other) => {
            let oid = oid_arg(other);
            ctx.catalog_database_names()
                .into_iter()
                .find(|n| crate::pgcatalog::snapshot::database_oid(n) == oid)
                .map_or(Target::Missing, |n| Target::Object(Object::Database(n)))
        }
        (Kind::Table | Kind::Column | Kind::AnyColumn | Kind::Sequence, value) => {
            let Some((schema, name)) = relation(ctx, value)? else { return Ok(Target::Missing) };
            if kind == Kind::Sequence {
                let sequences = crate::sequences::all(ctx.db, &ctx.txn.root)?;
                if !sequences.iter().any(|s| crate::sequences::schema_and_name(s) == (schema.clone(), name.clone())) {
                    return Err(PgError::new(code::WRONG_OBJECT_TYPE, format!("\"{name}\" is not a sequence")));
                }
                return Ok(Target::Object(Object::Sequence(schema, name)));
            }
            if kind == Kind::Column {
                let columns: Vec<String> = match crate::pgcatalog::lookup(&schema, &name) {
                    Some(catalog) => catalog.columns.iter().map(|c| c.name.to_string()).collect(),
                    None => match ctx.txn.table(ctx.db, &schema, &name)? {
                        Some(table) => table.columns.iter().map(|c| c.name.clone()).collect(),
                        None => {
                            ctx.view_columns(&schema, &name).unwrap_or_default().into_iter().map(|(n, _)| n).collect()
                        }
                    },
                };
                match &args[1] {
                    Value::Text(column) if !columns.contains(column) => {
                        return Err(PgError::new(
                            code::UNDEFINED_COLUMN,
                            format!("column \"{column}\" of relation \"{name}\" does not exist"),
                        ));
                    }
                    Value::Int2(number) if *number < 1 || *number as usize > columns.len() => {
                        return Ok(Target::Missing);
                    }
                    _ => {}
                }
            }
            Target::Object(Object::Table(schema, name))
        }
        (Kind::Function, Value::Text(name)) => {
            ctx.reg_value(Value::Text(name.clone()), crate::oid::REGPROCEDURE)?;
            Target::Everyone
        }
        (Kind::Type, Value::Text(name)) => {
            ctx.reg_value(Value::Text(name.clone()), crate::oid::REGTYPE)?;
            Target::Everyone
        }
        (Kind::Function | Kind::Type, other) => {
            let type_oid = if kind == Kind::Function { crate::oid::REGPROCEDURE } else { crate::oid::REGTYPE };
            match ctx.reg_value(Value::Oid(oid_arg(other)), type_oid)? {
                Value::Reg(reg) if reg.name != reg.oid.to_string() => Target::Everyone,
                _ => Target::Missing,
            }
        }
        (Kind::Language, value) => {
            let languages = crate::pgcatalog::reg::builtin_column("pg_language", "lanname");
            let found = match value {
                Value::Text(name) => languages.iter().any(|(_, n)| n.output().as_deref() == Some(name.as_str())),
                other => languages.iter().any(|(o, _)| *o == oid_arg(other)),
            };
            match (found, value) {
                (true, _) => Target::Everyone,
                (false, Value::Text(name)) => return Err(missing("language", name)),
                (false, _) => Target::Missing,
            }
        }
        (Kind::Tablespace, value) => match value {
            Value::Text(name) if matches!(name.as_str(), "pg_default" | "pg_global") => Target::Superusers,
            Value::Text(name) => return Err(missing("tablespace", name)),
            other if matches!(oid_arg(other), 1663 | 1664) => Target::Superusers,
            _ => Target::Missing,
        },
        (Kind::Server, Value::Text(name)) => return Err(missing("server", name)),
        (Kind::Wrapper, Value::Text(name)) => return Err(missing("foreign-data wrapper", name)),
        (Kind::Server | Kind::Wrapper, _) => Target::Missing,
        (Kind::Parameter, _) => Target::Superusers,
    })
}

/// has_privilege runs one of the has_*_privilege functions, checking the role, then the object's name, then the
/// privileges, as Postgres does.
fn has_privilege(ctx: &mut Ctx<'_>, args: &[Value], kind: Kind) -> Result<Value> {
    let named = if kind == Kind::Column { 3 } else { 2 };
    let (role, rest) = match args.len() > named {
        true => (Some(&args[0]), &args[1..]),
        false => (None, args),
    };
    let role = role_id(ctx, role)?;
    let target = target(ctx, kind, rest)?;
    let privileges = &rest[rest.len() - 1];
    let mut wanted = Vec::new();
    for chunk in text(privileges).split(',') {
        let chunk = chunk.trim_matches(|c: char| c.is_ascii_whitespace());
        let upper = chunk.to_ascii_uppercase();
        let (name, option) = match upper.strip_suffix(" WITH GRANT OPTION") {
            Some(name) => (name, true),
            None => (upper.as_str(), false),
        };
        let Some((_, letter)) = kind.privileges().iter().find(|(n, _)| *n == name) else {
            return Err(PgError::new(
                code::INVALID_PARAMETER_VALUE,
                format!("unrecognized privilege type: \"{chunk}\""),
            ));
        };
        wanted.push((*letter, option));
    }
    let superuser = ctx.auth()?.roles.get(&role).is_some_and(|r| r.superuser);
    Ok(match target {
        Target::Missing => Value::Null,
        Target::Everyone => Value::Bool(true),
        Target::Superusers => Value::Bool(superuser),
        Target::Object(Object::Table(schema, _))
            if matches!(schema.as_str(), "pg_catalog" | "information_schema")
                && wanted.iter().any(|(letter, option)| *letter == "r" && !option) =>
        {
            Value::Bool(true)
        }
        Target::Object(object) => Value::Bool(ctx.has_privilege(role, &object, &wanted)?),
    })
}

/// pg_has_role reports whether a role is a member of another, directly or through others, for MEMBER, or holds its
/// privileges through inheritance, for USAGE, as Postgres' pg_has_role does.
fn pg_has_role(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (role, rest) = match args.len() > 2 {
        true => (Some(&args[0]), &args[1..]),
        false => (None, args),
    };
    let role = role_id(ctx, role)?;
    let auth = ctx.auth()?;
    let target = match &rest[0] {
        Value::Text(name) => match auth.role(name) {
            Some(target) => target.id,
            None => return Err(PgError::new(code::UNDEFINED_OBJECT, format!("role \"{name}\" does not exist"))),
        },
        other => match ctx.role_of_oid(oid_arg(other)).and_then(|name| auth.role(&name).map(|r| r.id)) {
            Some(target) => target,
            None => return Ok(Value::Null),
        },
    };
    let superuser = auth.roles.get(&role).is_some_and(|r| r.superuser);
    let mut held = false;
    for chunk in text(&rest[1]).split(',') {
        let chunk = chunk.trim_matches(|c: char| c.is_ascii_whitespace());
        held |= match chunk.to_ascii_uppercase().as_str() {
            "MEMBER" => superuser || role == target || auth.groups(role, false).contains(&target),
            "USAGE" => superuser || role == target || auth.groups(role, true).contains(&target),
            "MEMBER WITH ADMIN OPTION"
            | "USAGE WITH ADMIN OPTION"
            | "MEMBER WITH GRANT OPTION"
            | "USAGE WITH GRANT OPTION" => {
                superuser || auth.memberships.get(&role).and_then(|m| m.get(&target)).is_some_and(|m| m.admin)
            }
            _ => {
                return Err(PgError::new(
                    code::INVALID_PARAMETER_VALUE,
                    format!("unrecognized privilege type: \"{chunk}\""),
                ));
            }
        };
    }
    Ok(Value::Bool(held))
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

/// pg_get_statisticsobjdef_columns returns NULL, since no extended statistics object exists for an OID to name.
fn pg_get_statisticsobjdef_columns(_: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Null)
}

/// pg_get_ruledef prints the `_RETURN` rule of a view as a CREATE RULE statement, prettily when asked, or returns NULL
/// for an OID that no rule has.
fn pg_get_ruledef(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    use crate::catalog::{id, oids};
    let rule = oid_arg(&args[0]);
    let pretty = matches!(args.get(1), Some(Value::Bool(true)));
    let snapshot = ctx.snapshot()?;
    let Some(view) = snapshot
        .views
        .iter()
        .find(|v| oids::oid(&id::new(id::SECTION_TRIGGER, &[&v.schema, &v.name, "_RETURN"])) == rule)
    else {
        return Ok(Value::Null);
    };
    let quote = crate::engine::quote_identifier;
    let relation = if pretty && ctx.session.search_path().contains(&view.schema) {
        quote(&view.name)
    } else {
        format!("{}.{}", quote(&view.schema), quote(&view.name))
    };
    let definition = ctx.view_definition(&view.statement, pretty, 0)?;
    Ok(Value::Text(format!("CREATE RULE \"_RETURN\" AS\n    ON SELECT TO {relation} DO INSTEAD {definition}")))
}

/// SIZE_UNITS are the units that pg_size_pretty prints sizes in: each unit's name, the size below which it is used,
/// whether it rounds halves away from zero, and its power of two.
const SIZE_UNITS: [(&str, i128, bool, u32); 6] = [
    ("bytes", 10 * 1024, false, 0),
    ("kB", 20 * 1024 - 1, true, 10),
    ("MB", 20 * 1024 - 1, true, 20),
    ("GB", 20 * 1024 - 1, true, 30),
    ("TB", 20 * 1024 - 1, true, 40),
    ("PB", 20 * 1024 - 1, true, 50),
];

/// pg_size_pretty prints a number of bytes in the largest unit that keeps it readable, as Postgres does.
fn pg_size_pretty(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let mut size: i128 = match &args[0] {
        Value::Int8(n) => *n as i128,
        Value::Numeric(n) => {
            let shown = n.to_string();
            match shown.split('.').next().and_then(|whole| whole.parse::<i128>().ok()) {
                Some(whole) if shown.contains('.') && whole.abs() < SIZE_UNITS[0].1 => {
                    return Ok(Value::Text(format!("{shown} bytes")));
                }
                Some(whole) => whole,
                None => return Ok(Value::Text(format!("{shown} bytes"))),
            }
        }
        _ => return Ok(Value::Null),
    };
    for (i, &(name, limit, round, bits)) in SIZE_UNITS.iter().enumerate() {
        let Some(&(_, _, next_round, next_bits)) = SIZE_UNITS.get(i + 1).filter(|_| size.abs() >= limit) else {
            if round {
                size = (size + if size < 0 { -1 } else { 1 }) / 2;
            }
            return Ok(Value::Text(format!("{size} {name}")));
        };
        let shift = next_bits - bits - u32::from(next_round) + u32::from(round);
        size /= 1i128 << shift;
    }
    Ok(Value::Null)
}

/// pg_trigger_depth returns how many trigger functions are running inside one another.
fn pg_trigger_depth(ctx: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Int4(ctx.session.trigger_depth))
}

/// pg_database_size returns the bytes that a database's files take, failing for a name that no database has and
/// returning NULL for such an OID.
fn pg_database_size(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let names = ctx.session.database_names();
    let name = match &args[0] {
        Value::Text(name) if names.contains(name) => name.clone(),
        Value::Text(name) => {
            return Err(PgError::new(code::INVALID_CATALOG_NAME, format!("database \"{name}\" does not exist")));
        }
        other => {
            let oid = oid_arg(other);
            match names.into_iter().find(|n| crate::pgcatalog::snapshot::database_oid(n) == oid) {
                Some(name) => name,
                None => return Ok(Value::Null),
            }
        }
    };
    Ok(Value::Int8(directory_size(&ctx.session.data_dir.join(name)) as i64))
}

/// directory_size adds up the sizes of the files under a directory.
fn directory_size(path: &std::path::Path) -> u64 {
    std::fs::read_dir(path).map_or(0, |entries| {
        entries
            .flatten()
            .map(|e| match e.file_type() {
                Ok(t) if t.is_dir() => directory_size(&e.path()),
                _ => e.metadata().map_or(0, |m| m.len()),
            })
            .sum()
    })
}
