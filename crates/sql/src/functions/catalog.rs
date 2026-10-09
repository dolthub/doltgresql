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

//! System catalog functions: name lookups, visibility, and descriptions of catalog objects.

use super::{ANY, Function};
use crate::error::{PgError, Result, code};
use crate::oid::{
    BOOL, INT4, INT8, NAME, NUMERIC, OID, REGCLASS, REGNAMESPACE, REGPROC, REGPROCEDURE, REGROLE, REGTYPE, TEXT,
};
use crate::query::Ctx;
use crate::types::Value;

/// INFORMATION_SCHEMA are the built-in functions that information_schema holds, which calls name with that schema.
pub const INFORMATION_SCHEMA: &[&str] = &["_pg_char_max_length", "_pg_truetypid"];

/// PG_TYPE is the row type of the pg_type catalog.
const PG_TYPE: u32 = 71;

/// PG_ATTRIBUTE is the row type of the pg_attribute catalog.
const PG_ATTRIBUTE: u32 = 75;

/// f declares a strict catalog function.
const fn f(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}

/// FUNCTIONS are the catalog functions.
pub const FUNCTIONS: &[Function] = &[
    f("to_regclass", &[TEXT], REGCLASS, to_regclass),
    f("_pg_char_max_length", &[OID, INT4], INT4, pg_char_max_length),
    f("_pg_truetypid", &[PG_ATTRIBUTE, PG_TYPE], OID, pg_truetypid),
    f("to_regtype", &[TEXT], REGTYPE, to_regtype),
    f("to_regproc", &[TEXT], REGPROC, to_regproc),
    f("to_regprocedure", &[TEXT], REGPROCEDURE, to_regprocedure),
    f("to_regnamespace", &[TEXT], REGNAMESPACE, to_regnamespace),
    f("to_regrole", &[TEXT], REGROLE, to_regrole),
    f("pg_table_is_visible", &[OID], BOOL, pg_table_is_visible),
    f("pg_type_is_visible", &[OID], BOOL, pg_type_is_visible),
    f("pg_function_is_visible", &[OID], BOOL, pg_function_is_visible),
    f("pg_collation_is_visible", &[OID], BOOL, pg_collation_is_visible),
    f("pg_conversion_is_visible", &[OID], BOOL, pg_conversion_is_visible),
    f("pg_operator_is_visible", &[OID], BOOL, pg_operator_is_visible),
    f("pg_opclass_is_visible", &[OID], BOOL, pg_opclass_is_visible),
    f("pg_opfamily_is_visible", &[OID], BOOL, pg_opfamily_is_visible),
    f("pg_statistics_obj_is_visible", &[OID], BOOL, pg_statistics_obj_is_visible),
    f("pg_ts_config_is_visible", &[OID], BOOL, pg_ts_config_is_visible),
    f("pg_ts_dict_is_visible", &[OID], BOOL, pg_ts_dict_is_visible),
    f("pg_ts_parser_is_visible", &[OID], BOOL, pg_ts_parser_is_visible),
    f("pg_ts_template_is_visible", &[OID], BOOL, pg_ts_template_is_visible),
    f("pg_get_userbyid", &[OID], NAME, pg_get_userbyid),
    f("getdatabaseencoding", &[], NAME, getdatabaseencoding),
    f("obj_description", &[OID], TEXT, obj_description),
    f("obj_description", &[OID, NAME], TEXT, obj_description),
    f("col_description", &[OID, INT4], TEXT, col_description),
    f("shobj_description", &[OID, NAME], TEXT, obj_description),
    f("pg_get_serial_sequence", &[TEXT, TEXT], TEXT, pg_get_serial_sequence),
    f("pg_partition_ancestors", &[REGCLASS], REGCLASS, pg_partition_ancestors),
    f("min_scale", &[NUMERIC], INT4, min_scale),
    f("pg_relation_size", &[REGCLASS], INT8, relation_size),
    f("pg_relation_size", &[REGCLASS, TEXT], INT8, relation_size),
    f("pg_table_size", &[REGCLASS], INT8, relation_size),
    f("pg_indexes_size", &[REGCLASS], INT8, relation_size),
    f("pg_total_relation_size", &[REGCLASS], INT8, relation_size),
    f("pg_relation_is_publishable", &[REGCLASS], BOOL, pg_relation_is_publishable),
    f("pg_get_partkeydef", &[OID], TEXT, pg_get_partkeydef),
    f("pg_tablespace_location", &[OID], TEXT, pg_tablespace_location),
    f("pg_stat_get_numscans", &[OID], INT8, pg_stat_get_numscans),
    Function { name: "num_nulls", args: &[ANY], ret: INT4, strict: false, variadic: true, implementation: num_nulls },
    Function {
        name: "num_nonnulls",
        args: &[ANY],
        ret: INT4,
        strict: false,
        variadic: true,
        implementation: num_nonnulls,
    },
];

/// OUT_COLUMNS are the result columns of the catalog functions that return rows.
pub const OUT_COLUMNS: &[(&str, &[(&str, u32)])] = &[("pg_partition_ancestors", &[("relid", REGCLASS)])];

/// pg_partition_ancestors returns the partitioned tables a partition belongs to, which are none, since Doltgres has
/// no partitioned tables.
fn pg_partition_ancestors(_: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Set(Vec::new()))
}

/// text returns the text of a string argument.
fn text(value: &Value) -> &str {
    match value {
        Value::Text(s) => s,
        _ => "",
    }
}

/// oid returns the OID of an OID argument.
fn oid(value: &Value) -> u32 {
    match value {
        Value::Oid(o) => *o,
        Value::Reg(reg) => reg.oid,
        Value::Int4(i) => *i as u32,
        Value::Int8(i) => *i as u32,
        _ => 0,
    }
}

/// to_regclass returns the relation that text names, or NULL.
fn to_regclass(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    ctx.to_reg(text(&args[0]), REGCLASS)
}

/// to_regtype returns the type that text names, or NULL.
fn to_regtype(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    ctx.to_reg(text(&args[0]), REGTYPE)
}

/// to_regproc returns the function that text names, or NULL.
fn to_regproc(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    ctx.to_reg(text(&args[0]), REGPROC)
}

/// to_regprocedure returns the function that text names with its argument types, or NULL.
fn to_regprocedure(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    ctx.to_reg(text(&args[0]), REGPROCEDURE)
}

/// to_regnamespace returns the schema that text names, or NULL.
fn to_regnamespace(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    ctx.to_reg(text(&args[0]), REGNAMESPACE)
}

/// to_regrole returns the role that text names, or NULL.
fn to_regrole(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    ctx.to_reg(text(&args[0]), REGROLE)
}

/// visibility declares a function that reports whether an object of a catalog is visible in the search path.
macro_rules! visibility {
    ($($name:ident => $catalog:literal),* $(,)?) => {$(
        #[doc = concat!(stringify!($name), " reports whether an object of ", $catalog, " is visible in the search path.")]
        fn $name(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
            ctx.is_visible(oid(&args[0]), $catalog)
        }
    )*};
}

visibility! {
    pg_table_is_visible => "pg_class",
    pg_type_is_visible => "pg_type",
    pg_function_is_visible => "pg_proc",
    pg_collation_is_visible => "pg_collation",
    pg_conversion_is_visible => "pg_conversion",
    pg_operator_is_visible => "pg_operator",
    pg_opclass_is_visible => "pg_opclass",
    pg_opfamily_is_visible => "pg_opfamily",
    pg_statistics_obj_is_visible => "pg_statistic_ext",
    pg_ts_config_is_visible => "pg_ts_config",
    pg_ts_dict_is_visible => "pg_ts_dict",
    pg_ts_parser_is_visible => "pg_ts_parser",
    pg_ts_template_is_visible => "pg_ts_template",
}

/// pg_get_userbyid returns the name of a role, or a placeholder naming the OID when there is no such role.
fn pg_get_userbyid(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let role = oid(&args[0]);
    Ok(Value::Text(ctx.role_of_oid(role).unwrap_or_else(|| format!("unknown (OID={role})"))))
}

/// getdatabaseencoding returns the database's encoding, which is always UTF8.
fn getdatabaseencoding(_: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Text("UTF8".into()))
}

/// obj_description returns the comment on an object of a catalog, which pg_class is without one.
fn obj_description(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let catalog = args.get(1).map_or("pg_class", text);
    Ok(ctx.description(oid(&args[0]), catalog, 0)?.map_or(Value::Null, Value::Text))
}

/// col_description returns the comment on a column of a relation.
fn col_description(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let column = match args[1] {
        Value::Int4(i) => i,
        _ => 0,
    };
    Ok(ctx.description(oid(&args[0]), "pg_class", column)?.map_or(Value::Null, Value::Text))
}

/// pg_get_serial_sequence returns the qualified name of the sequence that a column owns, or NULL without one.
fn pg_get_serial_sequence(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let names = crate::sequences::parse_qualified_name(text(&args[0]))?;
    let (schema, name) = match names.as_slice() {
        [name] => (String::new(), name.clone()),
        [schema, name] | [_, schema, name] => (schema.clone(), name.clone()),
        _ => {
            return Err(PgError::new(
                code::SYNTAX_ERROR,
                format!("improper relation name (too many dotted names): {}", text(&args[0])),
            ));
        }
    };
    if !schema.is_empty() && !ctx.schema_names().contains(&schema) {
        return Err(PgError::new(code::INVALID_SCHEMA_NAME, format!("schema \"{schema}\" does not exist")));
    }
    let relation = pg_query::protobuf::RangeVar {
        schemaname: schema.clone(),
        relname: name.clone(),
        inh: true,
        ..Default::default()
    };
    let table = match ctx.resolve_table(&relation) {
        Ok(table) => table,
        Err(err) => {
            let schemas = if schema.is_empty() { ctx.session.search_path() } else { vec![schema] };
            let snapshot = ctx.snapshot()?;
            let system =
                schemas.iter().find_map(|s| snapshot.system.iter().find(|(t, _)| t.schema == *s && t.name == name));
            match system {
                Some((table, _)) => table.clone(),
                None => return Err(PgError { position: None, ..err }),
            }
        }
    };
    let column = text(&args[1]);
    if !table.columns.iter().any(|c| c.name == column) {
        return Err(PgError::new(
            code::UNDEFINED_COLUMN,
            format!("column \"{column}\" of relation \"{}\" does not exist", table.name),
        ));
    }
    let owner = crate::catalog::id::new(crate::catalog::id::SECTION_TABLE, &[&table.schema, &table.name]);
    for sequence in ctx.snapshot()?.sequences.iter() {
        if sequence.owner_table == owner && sequence.owner_column == column.as_bytes() {
            let (schema, name) = crate::sequences::schema_and_name(sequence);
            return Ok(Value::Text(format!(
                "{}.{}",
                crate::engine::quote_identifier(&schema),
                crate::engine::quote_identifier(&name)
            )));
        }
    }
    Ok(Value::Null)
}

/// min_scale returns the fewest decimal digits that represent a numeric exactly.
fn min_scale(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let Value::Numeric(n) = &args[0] else { return Ok(Value::Null) };
    let text = n.to_string();
    if text == "NaN" || text.contains("Infinity") {
        return Ok(Value::Null);
    }
    let scale = text.split_once('.').map_or(0, |(_, fraction)| fraction.trim_end_matches('0').len());
    Ok(Value::Int4(scale as i32))
}

/// num_nulls counts its NULL arguments.
fn num_nulls(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Int4(args.iter().filter(|v| v.is_null()).count() as i32))
}

/// num_nonnulls counts its arguments that are not NULL.
fn num_nonnulls(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(Value::Int4(args.iter().filter(|v| !v.is_null()).count() as i32))
}

/// relation_size returns the disk space a relation uses, which Doltgres reports as 0 since its storage is shared
/// between tables, or NULL for an OID that no relation has.
fn relation_size(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    if let Some(fork) = args.get(1)
        && !matches!(text(fork), "main" | "fsm" | "vm" | "init")
    {
        return Err(PgError {
            hint: Some("Valid fork names are \"main\", \"fsm\", \"vm\", and \"init\".".into()),
            ..PgError::new(code::INVALID_PARAMETER_VALUE, "invalid fork name")
        });
    }
    Ok(if ctx.relation_exists(oid(&args[0]))? { Value::Int8(0) } else { Value::Null })
}

/// pg_relation_is_publishable reports whether a relation is a user table, or returns NULL for an OID that no relation
/// has.
fn pg_relation_is_publishable(ctx: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    Ok(ctx.is_publishable(oid(&args[0]))?.map_or(Value::Null, Value::Bool))
}

/// pg_get_partkeydef returns a partitioned table's partition key, which is always NULL, since Doltgres has no
/// partitioned tables.
fn pg_get_partkeydef(_: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Null)
}

/// pg_tablespace_location returns a tablespace's directory, which is empty for the built-in tablespaces that are
/// the only ones Doltgres has.
fn pg_tablespace_location(_: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Text(String::new()))
}

/// pg_stat_get_numscans returns how many scans used a relation, which Doltgres does not count.
fn pg_stat_get_numscans(_: &mut Ctx<'_>, _: &[Value]) -> Result<Value> {
    Ok(Value::Int8(0))
}

/// pg_char_max_length returns the length limit of a character or bit string type's modifier, as
/// information_schema._pg_char_max_length does.
fn pg_char_max_length(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let (Value::Oid(type_oid), Value::Int4(modifier)) = (&args[0], &args[1]) else { return Ok(Value::Null) };
    Ok(match (*type_oid, *modifier) {
        (_, -1) => Value::Null,
        (crate::oid::BPCHAR | crate::oid::VARCHAR, modifier) => Value::Int4(modifier - 4),
        (crate::oid::BIT | crate::oid::VARBIT, modifier) => Value::Int4(modifier),
        _ => Value::Null,
    })
}

/// pg_truetypid returns a column's type, or a domain's base type when the column's type is a domain, given the column's
/// pg_attribute row and its type's pg_type row, as information_schema._pg_truetypid does, where a row of only NULLs
/// comes from the missing side of an outer join and so is NULL.
fn pg_truetypid(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let fields = |row: &Value| match row {
        Value::Composite(c) => c.fields.clone(),
        Value::Record(fields) => fields.clone(),
        _ => Vec::new(),
    };
    if args.iter().any(|row| fields(row).iter().all(Value::is_null)) {
        return Ok(Value::Null);
    }
    let field = |row: &Value, catalog: &str, column: &str| {
        let fields = fields(row);
        crate::pgcatalog::lookup("pg_catalog", catalog)
            .and_then(|t| t.column(column))
            .and_then(|i| fields.get(i).cloned())
            .unwrap_or(Value::Null)
    };
    Ok(match field(&args[1], "pg_type", "typtype") {
        Value::Text(kind) if kind == "d" => field(&args[1], "pg_type", "typbasetype"),
        _ => field(&args[0], "pg_attribute", "atttypid"),
    })
}
