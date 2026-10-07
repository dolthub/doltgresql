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

//! The rows of the pg_catalog relations.

use crate::array::Array;
use crate::catalog::table::TableDef;
use crate::catalog::{ColumnType, builtin_type, id, oids};

/// SECTION_COLUMN_DEFAULT is the ID section of column defaults.
const SECTION_COLUMN_DEFAULT: u8 = 5;
use crate::error::Result;
use crate::foreign::Rule;
use crate::oid as types;
use crate::pgcatalog::snapshot::{
    Snapshot, constraint_oid, database_oid, index_oid, namespace_oid, sequence_oid, table_oid, view_oid,
};
use crate::pgcatalog::{Rows, boolean, int2, int4, oid, text};
use crate::query::Ctx;
use crate::types::{Reg, Value};

/// SUPERUSER is the OID of the bootstrap superuser.
pub const SUPERUSER: u32 = 10;

/// PREDEFINED_ROLES are Postgres 15's predefined roles.
pub(crate) const PREDEFINED_ROLES: [(u32, &str); 12] = [
    (3373, "pg_monitor"),
    (3374, "pg_read_all_settings"),
    (3375, "pg_read_all_stats"),
    (3377, "pg_stat_scan_tables"),
    (4200, "pg_signal_backend"),
    (4544, "pg_checkpoint"),
    (4569, "pg_read_server_files"),
    (4570, "pg_write_server_files"),
    (4571, "pg_execute_server_program"),
    (6171, "pg_database_owner"),
    (6181, "pg_read_all_data"),
    (6182, "pg_write_all_data"),
];

/// DEFAULT_BTREE_OPCLASSES maps a type to the OID and name of its default btree operator class.
const DEFAULT_BTREE_OPCLASSES: [(u32, u32, &str); 22] = [
    (16, 10003, "bool_ops"),
    (17, 10006, "bytea_ops"),
    (18, 10007, "char_ops"),
    (19, 10028, "name_ops"),
    (20, 3124, "int8_ops"),
    (21, 1979, "int2_ops"),
    (23, 1978, "int4_ops"),
    (25, 3126, "text_ops"),
    (26, 1981, "oid_ops"),
    (700, 10012, "float4_ops"),
    (701, 3123, "float8_ops"),
    (1042, 10004, "bpchar_ops"),
    (1043, 3126, "text_ops"),
    (1082, 3122, "date_ops"),
    (1083, 10038, "time_ops"),
    (1114, 3128, "timestamp_ops"),
    (1184, 3127, "timestamptz_ops"),
    (1186, 10022, "interval_ops"),
    (1266, 10041, "timetz_ops"),
    (1700, 3125, "numeric_ops"),
    (2950, 10065, "uuid_ops"),
    (3802, 10088, "jsonb_ops"),
];

/// default_opclass returns the default btree operator class of a type.
pub fn default_opclass(type_oid: u32) -> Option<(u32, &'static str)> {
    if crate::array::is_array_type(type_oid) {
        return Some((10000, "array_ops"));
    }
    DEFAULT_BTREE_OPCLASSES.iter().find(|(t, ..)| *t == type_oid).map(|&(_, o, n)| (o, n))
}

/// is_builtin_schema reports whether a schema is one that every Postgres database has.
pub fn is_builtin_schema(schema: &str) -> bool {
    matches!(schema, "pg_catalog" | "pg_toast" | "information_schema" | "public")
}

/// DEFAULT_COLLATION and C_COLLATION are the OIDs of the database's default collation and of the C collation.
const DEFAULT_COLLATION: u32 = 100;
const C_COLLATION: u32 = 950;

/// TypeInfo is what pg_attribute and pg_type repeat about a type.
struct TypeInfo {
    len: i16,
    by_value: bool,
    align: String,
    storage: String,
    collation: u32,
}

/// type_info returns what the catalogs show about a type.
fn type_info(type_oid: u32) -> TypeInfo {
    let Some(t) = builtin_type(type_oid) else {
        return TypeInfo { len: -1, by_value: false, align: "i".into(), storage: "x".into(), collation: 0 };
    };
    let d = &t.definition;
    let collation = match id::segments(&d.typ_collation).last().map(String::as_str) {
        Some("C") => C_COLLATION,
        Some(_) => DEFAULT_COLLATION,
        None => 0,
    };
    TypeInfo {
        len: d.typ_length,
        by_value: d.passed_by_val,
        align: String::from_utf8_lossy(&d.align).into_owned(),
        storage: String::from_utf8_lossy(&d.storage).into_owned(),
        collation,
    }
}

/// regproc returns a regproc value for a function ID, which is 0 and prints as `-` for an empty ID.
pub(super) fn regproc(function: &[u8]) -> Value {
    let name = id::segments(function).get(1).cloned();
    Value::Reg(Box::new(Reg {
        type_oid: types::REGPROC,
        oid: if name.is_some() { oids::oid(function) } else { 0 },
        name: name.unwrap_or_else(|| "-".into()),
    }))
}

/// builtin_proc returns a regproc value for a built-in function a type definition refers to, which is 0 and prints as
/// `-` for none.
fn builtin_proc(function: &[u8]) -> Value {
    match id::segments(function).get(1) {
        Some(name) => proc_named(name),
        None => regproc(&[]),
    }
}

/// proc_named returns a regproc value for a built-in function by name.
fn proc_named(name: &str) -> Value {
    let oid = crate::pgcatalog::reg::builtin_column("pg_proc", "proname")
        .into_iter()
        .find(|(_, n)| n.output().as_deref() == Some(name))
        .map_or(0, |(o, _)| o);
    Value::Reg(Box::new(Reg { type_oid: types::REGPROC, oid, name: name.to_string() }))
}

/// int2_array returns a smallint array.
fn int2_array(values: impl IntoIterator<Item = i16>) -> Value {
    Value::Array(Box::new(Array::one_dimensional(types::INT2, values.into_iter().map(Value::Int2).collect())))
}

/// vector returns the text of an int2vector or oidvector.
fn vector<T: ToString>(values: impl IntoIterator<Item = T>) -> Value {
    text(values.into_iter().map(|v| v.to_string()).collect::<Vec<_>>().join(" "))
}

/// role_oid returns the OID of a role, which is 10 for the superuser, as for Postgres' bootstrap superuser.
pub fn role_oid(name: &str, superuser: &str) -> u32 {
    if name == superuser { SUPERUSER } else { oids::oid(&id::new(id::SECTION_USER, &[name])) }
}

/// row_type_oid returns the OID of a table's or view's row type.
pub fn row_type_oid(schema: &str, name: &str) -> u32 {
    oids::oid(&id::new(id::SECTION_TYPE, &[schema, name]))
}

/// Attribute is a column of a relation, as pg_attribute shows it.
struct Attribute {
    relation: u32,
    name: String,
    ty: ColumnType,
    number: i16,
    not_null: bool,
    has_default: bool,
    generated: bool,
}

/// SYSTEM_COLUMNS are the system columns of every table, with their numbers and types.
const SYSTEM_COLUMNS: [(&str, i16, u32); 6] = [
    ("tableoid", -6, types::OID),
    ("cmax", -5, types::CID),
    ("xmax", -4, types::XID),
    ("cmin", -3, types::CID),
    ("xmin", -2, types::XID),
    ("ctid", -1, 27),
];

/// index_name returns the name of a table's index, where an empty name is its primary key.
fn index_name(table: &TableDef, index: &str) -> String {
    if index.is_empty() { table.primary_name() } else { index.to_string() }
}

/// TableIndex is an index of a table as the catalogs show it: its name, columns, and kind.
pub struct TableIndex {
    pub name: String,
    pub columns: Vec<usize>,
    pub unique: bool,
    pub primary: bool,
    pub descending: Vec<bool>,
    pub nulls_first: Vec<bool>,
    /// The distance of a vector index.
    pub vector: Option<prolly::Distance>,
    /// Whether a unique index's constraint is DEFERRABLE, and whether it is INITIALLY DEFERRED.
    pub deferrable: bool,
    pub initially_deferred: bool,
}

/// table_indexes returns a table's primary key index and its visible secondary indexes.
pub fn table_indexes(table: &TableDef) -> Vec<TableIndex> {
    let mut out = Vec::new();
    if !table.key_columns.is_empty() {
        out.push(TableIndex {
            name: index_name(table, ""),
            columns: table.key_columns.clone(),
            unique: true,
            primary: true,
            descending: vec![false; table.key_columns.len()],
            nulls_first: vec![false; table.key_columns.len()],
            vector: None,
            deferrable: table.primary.deferrable,
            initially_deferred: table.primary.initially_deferred,
        });
    }
    for index in table.indexes.iter().filter(|i| !i.system) {
        out.push(TableIndex {
            name: index.name.clone(),
            columns: index.columns.clone(),
            unique: index.unique,
            primary: false,
            descending: index.descending.clone(),
            nulls_first: index.nulls_last.iter().map(|&l| !l).collect(),
            vector: index.vector,
            deferrable: index.deferrable,
            initially_deferred: index.initially_deferred,
        });
    }
    out
}

impl Ctx<'_> {
    /// pg_catalog_rows fills the rows of a pg_catalog relation.
    pub(super) fn pg_catalog_rows(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        match rows.table.name {
            "pg_database" => self.pg_database(rows),
            "pg_namespace" => {
                self.pg_namespace(rows);
                Ok(())
            }
            "pg_authid" | "pg_roles" | "pg_user" | "pg_shadow" => self.pg_roles(rows),
            "pg_auth_members" | "pg_group" => self.pg_auth_members(rows),
            "pg_type" => self.pg_type(rows),
            "pg_class" => self.pg_class(rows),
            "pg_attribute" => self.pg_attribute(rows),
            "pg_index" => self.pg_index(rows),
            "pg_attrdef" => self.pg_attrdef(rows),
            "pg_indexes" => self.pg_indexes(rows),
            "pg_constraint" => self.pg_constraint(rows),
            "pg_tables" => self.pg_tables(rows),
            "pg_views" => self.pg_views(rows),
            "pg_sequence" | "pg_sequences" => self.pg_sequences(rows),
            "pg_proc" => self.pg_proc(rows),
            "pg_cast" => self.pg_cast(rows),
            "pg_aggregate" => self.pg_aggregate(rows),
            "pg_operator" => self.pg_operator(rows),
            "pg_enum" => self.pg_enum(rows),
            "pg_trigger" => self.pg_trigger(rows),
            "pg_extension" => self.pg_extension(rows),
            "pg_available_extensions" | "pg_available_extension_versions" => self.pg_available_extensions(rows),
            "pg_am" => self.pg_am(rows),
            "pg_opclass" => self.pg_opclass(rows),
            "pg_settings" => {
                self.pg_settings(rows);
                Ok(())
            }
            _ => Ok(()),
        }
    }

    /// pg_settings lists the configuration parameters with the session's values.
    fn pg_settings(&mut self, rows: &mut Rows<'_>) {
        let optional = |v: &str| if v.is_empty() { Value::Null } else { text(v) };
        for s in crate::settings::all_settings() {
            let value = self.session.settings.get(&s.name).unwrap_or_default();
            let enum_values = if s.enum_values.is_empty() {
                Value::Null
            } else {
                Value::Array(Box::new(Array::one_dimensional(
                    types::TEXT,
                    s.enum_values.iter().map(|v| text(v.clone())).collect(),
                )))
            };
            rows.push(vec![
                ("source", text(if value == s.default { "default" } else { "session" })),
                ("name", text(s.name.clone())),
                ("setting", text(value)),
                ("unit", optional(&s.unit)),
                ("category", text(s.category.clone())),
                ("short_desc", text(s.description.clone())),
                ("extra_desc", optional(&s.extra_description)),
                ("context", text(s.context.clone())),
                ("vartype", text(s.kind.clone())),
                ("min_val", optional(&s.min)),
                ("max_val", optional(&s.max)),
                ("enumvals", enum_values),
                ("boot_val", text(s.default.clone())),
                ("reset_val", text(s.default.clone())),
                ("pending_restart", boolean(false)),
            ]);
        }
    }

    /// catalog_database_names returns the databases, with the template databases that Postgres always has.
    pub(crate) fn catalog_database_names(&self) -> Vec<String> {
        let mut names = self.session.database_names();
        names.extend(["template0".to_string(), "template1".to_string()]);
        names
    }

    /// pg_database lists the databases, with the template databases that Postgres always has.
    fn pg_database(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        for name in self.catalog_database_names() {
            let template = name.starts_with("template") && !self.session.database_names().contains(&name);
            rows.push(vec![
                ("oid", oid(database_oid(&name))),
                ("datname", text(name.clone())),
                ("datdba", oid(SUPERUSER)),
                ("encoding", int4(6)),
                ("datlocprovider", text("c")),
                ("datistemplate", boolean(template)),
                ("datallowconn", boolean(name != "template0")),
                ("datconnlimit", int4(-1)),
                ("datfrozenxid", oid(716)),
                ("datminmxid", oid(1)),
                ("dattablespace", oid(1663)),
                ("datcollate", text("C")),
                ("datctype", text("C")),
            ]);
        }
        Ok(())
    }

    /// pg_namespace lists the schemas.
    fn pg_namespace(&mut self, rows: &mut Rows<'_>) {
        for schema in self.schema_names().into_iter().filter(|s| !is_builtin_schema(s)) {
            rows.push(vec![
                ("oid", oid(namespace_oid(&schema))),
                ("nspname", text(schema)),
                ("nspowner", oid(SUPERUSER)),
            ]);
        }
    }

    /// schema_names returns the names of the schemas of the root value, without Doltgres' own `dolt` schema.
    pub fn schema_names(&self) -> Vec<String> {
        let mut schemas: Vec<String> = self
            .txn
            .root
            .schemas
            .iter()
            .map(|s| String::from_utf8_lossy(s).into_owned())
            .filter(|s| s != "dolt")
            .collect();
        if schemas.is_empty() {
            schemas.push("public".into());
        }
        schemas.sort();
        schemas
    }

    /// pg_roles lists the roles, with Postgres' predefined roles, as pg_authid, pg_roles, pg_user, or pg_shadow.
    fn pg_roles(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let auth = self.auth()?.clone();
        let mut roles: Vec<(u32, crate::auth::Role)> = auth
            .roles
            .values()
            .filter(|r| r.name != crate::auth::PUBLIC)
            .map(|r| (role_oid(&r.name, &self.session.superuser), r.clone()))
            .collect();
        roles.extend(PREDEFINED_ROLES.iter().map(|&(o, n)| (o, crate::auth::Role::new(0, n))));
        let users = matches!(rows.table.name, "pg_user" | "pg_shadow");
        for (role_oid, role) in roles {
            if users && !role.login {
                continue;
            }
            let secret = role.password.as_ref().map_or(Value::Null, |p| text(p.text()));
            let masked =
                if rows.table.name == "pg_user" || rows.table.name == "pg_roles" { text("********") } else { secret };
            let valid_until = role.valid_until.map_or(Value::Null, |t| Value::TimestampTz(t - 946_684_800_000_000));
            rows.push(vec![
                ("oid", oid(role_oid)),
                ("rolname", text(role.name.clone())),
                ("rolsuper", boolean(role.superuser)),
                ("rolinherit", boolean(role.inherit)),
                ("rolcreaterole", boolean(role.create_role)),
                ("rolcreatedb", boolean(role.create_db)),
                ("rolcanlogin", boolean(role.login)),
                ("rolreplication", boolean(role.replication)),
                ("rolbypassrls", boolean(role.bypass_rls)),
                ("rolconnlimit", int4(role.connection_limit)),
                ("rolpassword", masked.clone()),
                ("rolvaliduntil", valid_until.clone()),
                ("usename", text(role.name.clone())),
                ("usesysid", oid(role_oid)),
                ("usecreatedb", boolean(role.create_db)),
                ("usesuper", boolean(role.superuser)),
                ("userepl", boolean(role.replication)),
                ("usebypassrls", boolean(role.bypass_rls)),
                ("passwd", masked),
                ("valuntil", valid_until),
            ]);
        }
        Ok(())
    }

    /// pg_auth_members lists role memberships, with pg_monitor's, or pg_group the members of each role that cannot log
    /// in.
    fn pg_auth_members(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let auth = self.auth()?.clone();
        let superuser = self.session.superuser.clone();
        let name_oid = |id: &u64| auth.roles.get(id).map_or(0, |r| role_oid(&r.name, &superuser));
        if rows.table.name == "pg_group" {
            for &(role, name) in &PREDEFINED_ROLES {
                let members: Vec<Value> = if matches!(role, 3374 | 3375 | 3377) { vec![oid(3373)] } else { Vec::new() };
                rows.push(vec![
                    ("groname", text(name)),
                    ("grosysid", oid(role)),
                    ("grolist", Value::Array(Box::new(Array::one_dimensional(types::OID, members)))),
                ]);
            }
            for role in auth.roles.values().filter(|r| !r.login && r.name != crate::auth::PUBLIC) {
                let members: Vec<Value> = auth
                    .memberships
                    .iter()
                    .filter(|(_, groups)| groups.contains_key(&role.id))
                    .map(|(member, _)| oid(name_oid(member)))
                    .collect();
                rows.push(vec![
                    ("groname", text(role.name.clone())),
                    ("grosysid", oid(role_oid(&role.name, &superuser))),
                    ("grolist", Value::Array(Box::new(Array::one_dimensional(types::OID, members)))),
                ]);
            }
            return Ok(());
        }
        for group in [3374, 3375, 3377] {
            rows.push(vec![
                ("roleid", oid(group)),
                ("member", oid(3373)),
                ("grantor", oid(SUPERUSER)),
                ("admin_option", boolean(false)),
            ]);
        }
        for (member, groups) in &auth.memberships {
            for (group, membership) in groups {
                rows.push(vec![
                    ("roleid", oid(name_oid(group))),
                    ("member", oid(name_oid(member))),
                    ("grantor", oid(name_oid(&membership.granted_by))),
                    ("admin_option", boolean(membership.admin)),
                ]);
            }
        }
        Ok(())
    }

    /// pg_type lists the row types of the user tables and views.
    fn pg_type(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        for user_type in self.user_types()?.values() {
            let definition = &user_type.definition;
            let (base, element) = match &user_type.kind {
                crate::usertypes::Kind::Domain(domain) => (domain.base.oid, 0),
                crate::usertypes::Kind::Array(element) => (0, *element),
                _ => (0, 0),
            };
            let relation = match user_type.kind {
                crate::usertypes::Kind::Composite(_) => table_oid(&user_type.schema, &user_type.name),
                _ => 0,
            };
            let typmod = if base == 0 { -1 } else { definition.typ_mod };
            rows.push(vec![
                ("oid", oid(user_type.oid)),
                ("typname", text(user_type.name.clone())),
                ("typnamespace", oid(namespace_oid(&user_type.schema))),
                ("typowner", oid(SUPERUSER)),
                ("typlen", int2(definition.typ_length)),
                ("typbyval", boolean(definition.passed_by_val)),
                ("typtype", text(String::from_utf8_lossy(&definition.typ_type).into_owned())),
                ("typcategory", text(String::from_utf8_lossy(&definition.typ_category).into_owned())),
                ("typispreferred", boolean(definition.is_preferred)),
                ("typisdefined", boolean(true)),
                ("typdelim", text(",")),
                ("typrelid", oid(relation)),
                ("typsubscript", builtin_proc(&definition.subscript_func)),
                ("typelem", oid(element)),
                ("typarray", oid(user_type.array)),
                ("typinput", builtin_proc(&definition.input_func)),
                ("typoutput", builtin_proc(&definition.output_func)),
                ("typreceive", builtin_proc(&definition.receive_func)),
                ("typsend", builtin_proc(&definition.send_func)),
                ("typmodin", builtin_proc(&definition.mod_in_func)),
                ("typmodout", builtin_proc(&definition.mod_out_func)),
                ("typanalyze", builtin_proc(&definition.analyze_func)),
                ("typalign", text(String::from_utf8_lossy(&definition.align).into_owned())),
                ("typstorage", text(String::from_utf8_lossy(&definition.storage).into_owned())),
                ("typnotnull", boolean(definition.not_null)),
                ("typbasetype", oid(base)),
                ("typtypmod", int4(typmod)),
                ("typndims", int4(0)),
                ("typcollation", oid(0)),
                (
                    "typdefault",
                    if definition.default.is_empty() {
                        Value::Null
                    } else {
                        text(String::from_utf8_lossy(&definition.default).into_owned())
                    },
                ),
            ]);
        }
        let snapshot = self.snapshot()?;
        let relations = snapshot
            .tables
            .iter()
            .map(|t| (t.schema.clone(), t.name.clone(), table_oid(&t.schema, &t.name)))
            .chain(snapshot.views.iter().map(|v| (v.schema.clone(), v.name.clone(), view_oid(&v.schema, &v.name))));
        for (schema, name, relation) in relations {
            let array = oids::oid(&id::new(id::SECTION_TYPE, &[&schema, &format!("_{name}")]));
            rows.push(vec![
                ("oid", oid(array)),
                ("typname", text(format!("_{name}"))),
                ("typnamespace", oid(namespace_oid(&schema))),
                ("typowner", oid(SUPERUSER)),
                ("typlen", int2(-1)),
                ("typbyval", boolean(false)),
                ("typtype", text("b")),
                ("typcategory", text("A")),
                ("typispreferred", boolean(false)),
                ("typisdefined", boolean(true)),
                ("typdelim", text(",")),
                ("typrelid", oid(0)),
                ("typsubscript", proc_named("array_subscript_handler")),
                ("typelem", oid(row_type_oid(&schema, &name))),
                ("typarray", oid(0)),
                ("typinput", proc_named("array_in")),
                ("typoutput", proc_named("array_out")),
                ("typreceive", proc_named("array_recv")),
                ("typsend", proc_named("array_send")),
                ("typmodin", regproc(&[])),
                ("typmodout", regproc(&[])),
                ("typanalyze", proc_named("array_typanalyze")),
                ("typalign", text("d")),
                ("typstorage", text("x")),
                ("typnotnull", boolean(false)),
                ("typbasetype", oid(0)),
                ("typtypmod", int4(-1)),
                ("typndims", int4(0)),
                ("typcollation", oid(0)),
            ]);
            rows.push(vec![
                ("oid", oid(row_type_oid(&schema, &name))),
                ("typname", text(name)),
                ("typnamespace", oid(namespace_oid(&schema))),
                ("typowner", oid(SUPERUSER)),
                ("typlen", int2(-1)),
                ("typbyval", boolean(false)),
                ("typtype", text("c")),
                ("typcategory", text("C")),
                ("typispreferred", boolean(false)),
                ("typisdefined", boolean(true)),
                ("typdelim", text(",")),
                ("typrelid", oid(relation)),
                ("typsubscript", regproc(&[])),
                ("typelem", oid(0)),
                ("typarray", oid(array)),
                ("typinput", proc_named("record_in")),
                ("typoutput", proc_named("record_out")),
                ("typreceive", proc_named("record_recv")),
                ("typsend", proc_named("record_send")),
                ("typmodin", regproc(&[])),
                ("typmodout", regproc(&[])),
                ("typanalyze", regproc(&[])),
                ("typalign", text("d")),
                ("typstorage", text("x")),
                ("typnotnull", boolean(false)),
                ("typbasetype", oid(0)),
                ("typtypmod", int4(-1)),
                ("typndims", int4(0)),
                ("typcollation", oid(0)),
            ]);
        }
        Ok(())
    }

    /// pg_class lists the user tables, indexes, views, and sequences.
    fn pg_class(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let snapshot = self.snapshot()?;
        let triggered = self.triggered_tables()?;
        for table in &snapshot.tables {
            let relation = table_oid(&table.schema, &table.name);
            let namespace = namespace_oid(&table.schema);
            let indexes = table_indexes(table);
            let mut row = class_row(relation, &table.name, namespace, "r", table.columns.len() as i16, 2);
            row.extend([
                ("reltype", oid(row_type_oid(&table.schema, &table.name))),
                ("relfilenode", oid(relation)),
                ("relhasindex", boolean(!indexes.is_empty())),
                ("relchecks", int2(table.checks.len() as i16)),
                (
                    "relhastriggers",
                    boolean(
                        has_foreign_keys(&snapshot, table)
                            || triggered.contains(&(table.schema.clone(), table.name.clone())),
                    ),
                ),
                ("relreplident", text("d")),
                ("relminmxid", oid(1)),
            ]);
            rows.push(row);
            for index in indexes {
                let index_relation = index_oid(&table.schema, &table.name, &index.name);
                let method = match index.vector {
                    Some(_) => super::extensions::access_method_oid("hnsw"),
                    None => 403,
                };
                let mut row =
                    class_row(index_relation, &index.name, namespace, "i", index.columns.len() as i16, method);
                row.extend([
                    ("relfilenode", oid(index_relation)),
                    ("relpages", int4(1)),
                    ("reltuples", Value::Float4(0.0)),
                ]);
                rows.push(row);
            }
        }
        for view in &snapshot.views {
            let columns = self.view_columns(&view.schema, &view.name).map_or(0, |c| c.len());
            let mut row = class_row(
                view_oid(&view.schema, &view.name),
                &view.name,
                namespace_oid(&view.schema),
                "v",
                columns as i16,
                0,
            );
            row.extend([("reltype", oid(row_type_oid(&view.schema, &view.name))), ("relhasrules", boolean(true))]);
            rows.push(row);
        }
        for user_type in self.user_types()?.values() {
            let crate::usertypes::Kind::Composite(fields) = &user_type.kind else { continue };
            let relation = table_oid(&user_type.schema, &user_type.name);
            let namespace = namespace_oid(&user_type.schema);
            let mut row = class_row(relation, &user_type.name, namespace, "c", fields.len() as i16, 0);
            row.push(("reltype", oid(user_type.oid)));
            rows.push(row);
        }
        for sequence in &snapshot.sequences {
            let (schema, name) = crate::sequences::schema_and_name(sequence);
            let relation = sequence_oid(&schema, &name);
            let mut row = class_row(relation, &name, namespace_oid(&schema), "S", 3, 0);
            row.extend([("relfilenode", oid(relation)), ("relpages", int4(1)), ("reltuples", Value::Float4(1.0))]);
            rows.push(row);
        }
        Ok(())
    }

    /// view_columns returns the names and types of a view's columns, by planning its query.
    pub fn view_columns(&mut self, schema: &str, name: &str) -> Option<Vec<(String, ColumnType)>> {
        let (_, fragment) = self.find_view(schema, name).ok()??;
        let (select, aliases) = crate::views::view_query(&fragment).ok()?;
        let query = crate::plan::Planner { ctx: self, outer: Vec::new() }.plan_query(&select).ok()?;
        Some(
            query
                .columns
                .iter()
                .zip(&query.types)
                .enumerate()
                .map(|(i, (c, &ty))| (aliases.get(i).cloned().unwrap_or_else(|| c.name.clone()), ty))
                .collect(),
        )
    }

    /// pg_attribute lists the columns of the user relations, with the system columns of tables.
    fn pg_attribute(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let mut attributes: Vec<Attribute> = Vec::new();
        let system = |attributes: &mut Vec<Attribute>, relation: u32| {
            for (name, number, type_oid) in SYSTEM_COLUMNS {
                attributes.push(Attribute {
                    relation,
                    name: name.into(),
                    ty: ColumnType { oid: type_oid, modifier: -1 },
                    number,
                    not_null: true,
                    has_default: false,
                    generated: false,
                });
            }
        };
        let snapshot = self.snapshot()?;
        for table in &snapshot.tables {
            let relation = table_oid(&table.schema, &table.name);
            system(&mut attributes, relation);
            for (i, column) in table.columns.iter().enumerate() {
                attributes.push(Attribute {
                    relation,
                    name: column.name.clone(),
                    ty: column.ty,
                    number: i as i16 + 1,
                    not_null: !column.nullable,
                    has_default: !column.default.is_empty(),
                    generated: column.generated,
                });
            }
            for index in table_indexes(table) {
                let index_relation = index_oid(&table.schema, &table.name, &index.name);
                for (i, &c) in index.columns.iter().enumerate() {
                    attributes.push(Attribute {
                        relation: index_relation,
                        name: table.columns[c].name.clone(),
                        ty: table.columns[c].ty,
                        number: i as i16 + 1,
                        not_null: false,
                        has_default: false,
                        generated: false,
                    });
                }
            }
        }
        for user_type in self.user_types()?.values() {
            let crate::usertypes::Kind::Composite(fields) = &user_type.kind else { continue };
            let relation = table_oid(&user_type.schema, &user_type.name);
            for (i, (name, ty)) in fields.iter().enumerate() {
                attributes.push(Attribute {
                    relation,
                    name: name.clone(),
                    ty: *ty,
                    number: i as i16 + 1,
                    not_null: false,
                    has_default: false,
                    generated: false,
                });
            }
        }
        for view in &snapshot.views {
            let relation = view_oid(&view.schema, &view.name);
            for (i, (name, ty)) in
                self.view_columns(&view.schema, &view.name).unwrap_or_default().into_iter().enumerate()
            {
                attributes.push(Attribute {
                    relation,
                    name,
                    ty,
                    number: i as i16 + 1,
                    not_null: false,
                    has_default: false,
                    generated: false,
                });
            }
        }
        for sequence in &snapshot.sequences {
            let (schema, name) = crate::sequences::schema_and_name(sequence);
            let relation = sequence_oid(&schema, &name);
            for (i, (column, type_oid)) in
                [("last_value", types::INT8), ("log_cnt", types::INT8), ("is_called", types::BOOL)]
                    .into_iter()
                    .enumerate()
            {
                attributes.push(Attribute {
                    relation,
                    name: column.into(),
                    ty: ColumnType { oid: type_oid, modifier: -1 },
                    number: i as i16 + 1,
                    not_null: true,
                    has_default: false,
                    generated: false,
                });
            }
        }
        for a in attributes {
            let info = type_info(a.ty.oid);
            rows.push(vec![
                ("attrelid", oid(a.relation)),
                ("attname", text(a.name)),
                ("atttypid", oid(a.ty.oid)),
                ("attstattarget", int4(-1)),
                ("attlen", int2(info.len)),
                ("attnum", int2(a.number)),
                ("attndims", int4(crate::array::is_array_type(a.ty.oid) as i32)),
                ("attcacheoff", int4(-1)),
                ("atttypmod", int4(a.ty.modifier)),
                ("attbyval", boolean(info.by_value)),
                ("attalign", text(info.align)),
                ("attstorage", text(info.storage)),
                ("attcompression", text("")),
                ("attnotnull", boolean(a.not_null)),
                ("atthasdef", boolean(a.has_default)),
                ("atthasmissing", boolean(false)),
                ("attidentity", text("")),
                ("attgenerated", text(if a.generated { "s" } else { "" })),
                ("attisdropped", boolean(false)),
                ("attislocal", boolean(true)),
                ("attinhcount", int4(0)),
                ("attcollation", oid(info.collation)),
            ]);
        }
        Ok(())
    }

    /// pg_index lists the indexes of the user tables.
    fn pg_index(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let snapshot = self.snapshot()?;
        for table in &snapshot.tables {
            for index in table_indexes(table) {
                let types: Vec<u32> = index.columns.iter().map(|&c| table.columns[c].ty.oid).collect();
                rows.push(vec![
                    ("indexrelid", oid(index_oid(&table.schema, &table.name, &index.name))),
                    ("indrelid", oid(table_oid(&table.schema, &table.name))),
                    ("indnatts", int2(index.columns.len() as i16)),
                    ("indnkeyatts", int2(index.columns.len() as i16)),
                    ("indisunique", boolean(index.unique)),
                    ("indnullsnotdistinct", boolean(false)),
                    ("indisprimary", boolean(index.primary)),
                    ("indisexclusion", boolean(false)),
                    ("indimmediate", boolean(!index.deferrable)),
                    ("indisclustered", boolean(false)),
                    ("indisvalid", boolean(true)),
                    ("indcheckxmin", boolean(false)),
                    ("indisready", boolean(true)),
                    ("indislive", boolean(true)),
                    ("indisreplident", boolean(false)),
                    ("indkey", vector(index.columns.iter().map(|c| c + 1))),
                    ("indcollation", vector(types.iter().map(|&t| type_info(t).collation))),
                    ("indclass", vector(types.iter().map(|&t| default_opclass(t).map_or(0, |o| o.0)))),
                    (
                        "indoption",
                        vector(
                            index.descending.iter().zip(&index.nulls_first).map(|(&d, &f)| d as i32 | (f as i32) << 1),
                        ),
                    ),
                ]);
            }
        }
        Ok(())
    }

    /// pg_indexes lists each index of the user tables with its definition.
    fn pg_indexes(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let snapshot = self.snapshot()?;
        for table in &snapshot.tables {
            for index in table_indexes(table) {
                rows.push(vec![
                    ("schemaname", text(table.schema.clone())),
                    ("tablename", text(table.name.clone())),
                    ("indexname", text(index.name.clone())),
                    ("indexdef", text(index_definition(table, &index, true))),
                ]);
            }
        }
        Ok(())
    }

    /// index_definition_of returns the definition of the index with the OID as pg_get_indexdef prints it, prettily
    /// when asked, or the name of one of its columns for a nonzero column number, or None when no index has the OID.
    pub(crate) fn index_definition_of(&mut self, index: u32, column: i32, pretty: bool) -> Result<Option<String>> {
        let snapshot = self.snapshot()?;
        for table in &snapshot.tables {
            for candidate in table_indexes(table) {
                if index_oid(&table.schema, &table.name, &candidate.name) != index {
                    continue;
                }
                return Ok(Some(match usize::try_from(column) {
                    Ok(0) => index_definition(
                        table,
                        &candidate,
                        !pretty || !self.session.search_path().contains(&table.schema),
                    ),
                    Ok(position) => candidate
                        .columns
                        .get(position - 1)
                        .map(|&c| crate::engine::quote_identifier(&table.columns[c].name))
                        .unwrap_or_default(),
                    Err(_) => String::new(),
                }));
            }
        }
        Ok(None)
    }

    /// pg_attrdef lists the defaults and generation expressions of the user tables' columns.
    fn pg_attrdef(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let snapshot = self.snapshot()?;
        for table in &snapshot.tables {
            let columns: Vec<(String, ColumnType)> = table.columns.iter().map(|c| (c.name.clone(), c.ty)).collect();
            for (i, column) in table.columns.iter().enumerate().filter(|(_, c)| !c.default.is_empty()) {
                let mut analyzer = crate::ruleutils::Analyzer::new(self, columns.clone());
                let expression = analyzer.deparse(&column.default, Some(column.ty), false)?;
                rows.push(vec![
                    (
                        "oid",
                        oid(oids::oid(&id::new(SECTION_COLUMN_DEFAULT, &[&table.schema, &table.name, &column.name]))),
                    ),
                    ("adrelid", oid(table_oid(&table.schema, &table.name))),
                    ("adnum", int2(i as i16 + 1)),
                    ("adbin", text(expression)),
                ]);
            }
        }
        Ok(())
    }

    /// expression_definition prints an expression over a relation's columns as pg_get_expr does.
    pub(crate) fn expression_definition(&mut self, expression: &str, relation: u32, pretty: bool) -> Result<String> {
        let snapshot = self.snapshot()?;
        let columns = snapshot
            .tables
            .iter()
            .find(|t| table_oid(&t.schema, &t.name) == relation)
            .map(|t| t.columns.iter().map(|c| (c.name.clone(), c.ty)).collect())
            .unwrap_or_default();
        crate::ruleutils::Analyzer::new(self, columns).deparse(expression, None, pretty)
    }

    /// constraint_definition_of prints the constraint with the OID as pg_get_constraintdef does, or returns None when
    /// no constraint has the OID.
    pub(crate) fn constraint_definition_of(&mut self, constraint: u32, pretty: bool) -> Result<Option<String>> {
        let snapshot = self.snapshot()?;
        let deferral = |deferrable: bool, deferred: bool| {
            let mut suffix = String::new();
            if deferrable {
                suffix.push_str(" DEFERRABLE");
            }
            if deferred {
                suffix.push_str(" INITIALLY DEFERRED");
            }
            suffix
        };
        let names = |table: &TableDef, columns: &[usize]| -> String {
            let names: Vec<String> =
                columns.iter().map(|&c| crate::engine::quote_identifier(&table.columns[c].name)).collect();
            names.join(", ")
        };
        for table in &snapshot.tables {
            for index in table_indexes(table).into_iter().filter(|i| i.unique) {
                let section = if index.primary { 23 } else { 36 };
                if constraint_oid(section, &table.schema, &table.name, &index.name) == constraint {
                    let kind = if index.primary { "PRIMARY KEY" } else { "UNIQUE" };
                    let suffix = deferral(index.deferrable, index.initially_deferred);
                    return Ok(Some(format!("{kind} ({}){suffix}", names(table, &index.columns))));
                }
            }
            for check in &table.checks {
                if constraint_oid(3, &table.schema, &table.name, &check.name) == constraint {
                    let columns = table.columns.iter().map(|c| (c.name.clone(), c.ty)).collect();
                    let text =
                        crate::ruleutils::Analyzer::new(self, columns).deparse(&check.expression, None, pretty)?;
                    return Ok(Some(format!("CHECK ({text})")));
                }
            }
            for fk in
                snapshot.foreign_keys.iter().filter(|f| f.child_schema == table.schema && f.child_table == table.name)
            {
                if constraint_oid(11, &table.schema, &table.name, &fk.name) != constraint {
                    continue;
                }
                let quote = |names: &[String]| {
                    names.iter().map(|n| crate::engine::quote_identifier(n)).collect::<Vec<_>>().join(", ")
                };
                let parent = if self.session.search_path().contains(&fk.parent_schema) {
                    crate::engine::quote_identifier(&fk.parent_table)
                } else {
                    format!(
                        "{}.{}",
                        crate::engine::quote_identifier(&fk.parent_schema),
                        crate::engine::quote_identifier(&fk.parent_table)
                    )
                };
                let mut text = format!(
                    "FOREIGN KEY ({}) REFERENCES {parent}({})",
                    quote(&fk.child_columns),
                    quote(&fk.parent_columns)
                );
                let action = |rule: Rule| match rule {
                    Rule::NoAction => None,
                    Rule::Restrict => Some("RESTRICT"),
                    Rule::Cascade => Some("CASCADE"),
                    Rule::SetNull => Some("SET NULL"),
                    Rule::SetDefault => Some("SET DEFAULT"),
                };
                if let Some(action) = action(fk.on_update) {
                    text.push_str(&format!(" ON UPDATE {action}"));
                }
                if let Some(action) = action(fk.on_delete) {
                    text.push_str(&format!(" ON DELETE {action}"));
                }
                if fk.match_full {
                    text.push_str(" MATCH FULL");
                }
                text.push_str(&deferral(fk.deferrable, fk.initially_deferred));
                if fk.not_valid {
                    text.push_str(" NOT VALID");
                }
                return Ok(Some(text));
            }
        }
        Ok(None)
    }

    /// pg_constraint lists the primary key, unique, check, and foreign key constraints of the user tables.
    fn pg_constraint(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let snapshot = self.snapshot()?;
        for table in &snapshot.tables {
            let relation = table_oid(&table.schema, &table.name);
            let namespace = namespace_oid(&table.schema);
            let base = |name: &str, kind: &str, section: u8, (deferrable, deferred): (bool, bool)| {
                vec![
                    ("oid", oid(constraint_oid(section, &table.schema, &table.name, name))),
                    ("conname", text(name)),
                    ("connamespace", oid(namespace)),
                    ("contype", text(kind)),
                    ("condeferrable", boolean(deferrable)),
                    ("condeferred", boolean(deferred)),
                    ("convalidated", boolean(true)),
                    ("conrelid", oid(relation)),
                    ("contypid", oid(0)),
                    ("conparentid", oid(0)),
                    ("confrelid", oid(0)),
                    ("confupdtype", text(" ")),
                    ("confdeltype", text(" ")),
                    ("confmatchtype", text(" ")),
                    ("conislocal", boolean(true)),
                    ("coninhcount", int4(0)),
                    ("connoinherit", boolean(kind != "c")),
                ]
            };
            for index in table_indexes(table).into_iter().filter(|i| i.unique) {
                let (kind, section) = if index.primary { ("p", 23) } else { ("u", 36) };
                let mut row = base(&index.name, kind, section, (index.deferrable, index.initially_deferred));
                row.extend([
                    ("conindid", oid(index_oid(&table.schema, &table.name, &index.name))),
                    ("conkey", int2_array(index.columns.iter().map(|&c| c as i16 + 1))),
                ]);
                rows.push(row);
            }
            for check in &table.checks {
                let mut row = base(&check.name, "c", 3, (false, false));
                let columns: Vec<i16> = table
                    .columns
                    .iter()
                    .enumerate()
                    .filter(|(_, c)| crate::alter::references(&check.expression, &c.name))
                    .map(|(i, _)| i as i16 + 1)
                    .collect();
                row.extend([("conindid", oid(0)), ("conkey", int2_array(columns))]);
                rows.push(row);
            }
            for fk in
                snapshot.foreign_keys.iter().filter(|f| f.child_schema == table.schema && f.child_table == table.name)
            {
                let Some(parent) = snapshot.table(&fk.parent_schema, &fk.parent_table) else { continue };
                let position =
                    |t: &TableDef, c: &String| t.columns.iter().position(|col| col.name == *c).unwrap_or(0) as i16 + 1;
                let parent_index =
                    if fk.parent_index.is_empty() { index_name(parent, "") } else { fk.parent_index.clone() };
                let mut row = base(&fk.name, "f", 11, (fk.deferrable, fk.initially_deferred));
                row.extend([
                    ("conindid", oid(index_oid(&parent.schema, &parent.name, &parent_index))),
                    ("confrelid", oid(table_oid(&parent.schema, &parent.name))),
                    ("confupdtype", text(rule_letter(fk.on_update))),
                    ("confdeltype", text(rule_letter(fk.on_delete))),
                    ("confmatchtype", text(if fk.match_full { "f" } else { "s" })),
                    ("convalidated", boolean(!fk.not_valid)),
                    ("conkey", int2_array(fk.child_columns.iter().map(|c| position(table, c)))),
                    ("confkey", int2_array(fk.parent_columns.iter().map(|c| position(parent, c)))),
                ]);
                rows.push(row);
            }
        }
        Ok(())
    }

    /// triggered_tables returns the schema and name of each table with a user trigger.
    fn triggered_tables(&mut self) -> Result<std::collections::HashSet<(String, String)>> {
        Ok(self
            .triggers()?
            .iter()
            .map(|t| {
                let (schema, table, _) = crate::triggers::names(t);
                (schema, table)
            })
            .collect())
    }

    /// pg_enum lists the labels of the user-defined enums.
    fn pg_enum(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        for user_type in self.user_types()?.values() {
            for label in &user_type.definition.enum_labels {
                rows.push(vec![
                    ("oid", oid(oids::oid(&label.id))),
                    ("enumtypid", oid(user_type.oid)),
                    ("enumsortorder", Value::Float4(label.sort_order)),
                    ("enumlabel", text(id::segments(&label.id).pop().unwrap_or_default())),
                ]);
            }
        }
        Ok(())
    }

    /// pg_tables lists the user tables.
    fn pg_tables(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let owner = self.session.superuser.clone();
        let snapshot = self.snapshot()?;
        let triggered = self.triggered_tables()?;
        for table in &snapshot.tables {
            let has_triggers =
                has_foreign_keys(&snapshot, table) || triggered.contains(&(table.schema.clone(), table.name.clone()));
            rows.push(vec![
                ("schemaname", text(table.schema.clone())),
                ("tablename", text(table.name.clone())),
                ("tableowner", text(owner.clone())),
                ("hasindexes", boolean(!table_indexes(table).is_empty())),
                ("hasrules", boolean(false)),
                ("hastriggers", boolean(has_triggers)),
                ("rowsecurity", boolean(false)),
            ]);
        }
        Ok(())
    }

    /// pg_views lists the user views.
    fn pg_views(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let owner = self.session.superuser.clone();
        for view in self.snapshot()?.views {
            let definition = crate::views::view_definition(&view.statement).unwrap_or_default();
            rows.push(vec![
                ("schemaname", text(view.schema)),
                ("viewname", text(view.name)),
                ("viewowner", text(owner.clone())),
                ("definition", text(definition)),
            ]);
        }
        Ok(())
    }

    /// pg_sequences fills pg_sequence or the pg_sequences view.
    fn pg_sequences(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let owner = self.session.superuser.clone();
        for sequence in self.snapshot()?.sequences {
            let (schema, name) = crate::sequences::schema_and_name(&sequence);
            let data_type = crate::catalog::builtin_type_by_id(&sequence.data_type_id).map_or(types::INT8, |t| t.oid);
            let display = crate::cast::type_display(data_type).into_owned();
            rows.push(vec![
                ("seqrelid", oid(sequence_oid(&schema, &name))),
                ("seqtypid", oid(data_type)),
                ("seqstart", Value::Int8(sequence.start)),
                ("seqincrement", Value::Int8(sequence.increment)),
                ("seqmax", Value::Int8(sequence.maximum)),
                ("seqmin", Value::Int8(sequence.minimum)),
                ("seqcache", Value::Int8(sequence.cache)),
                ("seqcycle", boolean(sequence.cycle)),
                ("schemaname", text(schema)),
                ("sequencename", text(name)),
                ("sequenceowner", text(owner.clone())),
                ("data_type", Value::Reg(Box::new(Reg { type_oid: types::REGTYPE, oid: data_type, name: display }))),
                ("start_value", Value::Int8(sequence.start)),
                ("min_value", Value::Int8(sequence.minimum)),
                ("max_value", Value::Int8(sequence.maximum)),
                ("increment_by", Value::Int8(sequence.increment)),
                ("cycle", boolean(sequence.cycle)),
                ("cache_size", Value::Int8(sequence.cache)),
                (
                    "last_value",
                    match (sequence.has_been_called, sequence.is_at_end) {
                        (false, _) => Value::Null,
                        (true, true) => Value::Int8(sequence.current),
                        (true, false) => Value::Int8(sequence.current - sequence.increment),
                    },
                ),
            ]);
        }
        Ok(())
    }
}

/// class_row returns the pg_class columns shared by every kind of relation.
fn class_row(
    relation: u32,
    name: &str,
    namespace: u32,
    kind: &str,
    columns: i16,
    access_method: u32,
) -> Vec<(&'static str, Value)> {
    let replica_identity = if kind == "r" { "d" } else { "n" };
    vec![
        ("oid", oid(relation)),
        ("relname", text(name)),
        ("relnamespace", oid(namespace)),
        ("reltype", oid(0)),
        ("reloftype", oid(0)),
        ("relowner", oid(SUPERUSER)),
        ("relam", oid(access_method)),
        ("relfilenode", oid(0)),
        ("reltablespace", oid(0)),
        ("relpages", int4(0)),
        ("reltuples", Value::Float4(-1.0)),
        ("relallvisible", int4(0)),
        ("reltoastrelid", oid(0)),
        ("relhasindex", boolean(false)),
        ("relisshared", boolean(false)),
        ("relpersistence", text("p")),
        ("relkind", text(kind)),
        ("relnatts", int2(columns)),
        ("relchecks", int2(0)),
        ("relhasrules", boolean(false)),
        ("relhastriggers", boolean(false)),
        ("relhassubclass", boolean(false)),
        ("relrowsecurity", boolean(false)),
        ("relforcerowsecurity", boolean(false)),
        ("relispopulated", boolean(true)),
        ("relreplident", text(replica_identity)),
        ("relispartition", boolean(false)),
        ("relrewrite", oid(0)),
        ("relfrozenxid", oid(0)),
        ("relminmxid", oid(0)),
    ]
}

/// has_foreign_keys reports whether a table refers to another or another refers to it, which Postgres enforces with
/// triggers.
fn has_foreign_keys(snapshot: &Snapshot, table: &TableDef) -> bool {
    snapshot.foreign_keys.iter().any(|fk| {
        (fk.child_schema == table.schema && fk.child_table == table.name)
            || (fk.parent_schema == table.schema && fk.parent_table == table.name)
    })
}

/// rule_letter returns the letter pg_constraint shows for a foreign key action.
fn rule_letter(rule: Rule) -> &'static str {
    match rule {
        Rule::NoAction => "a",
        Rule::Restrict => "r",
        Rule::Cascade => "c",
        Rule::SetNull => "n",
        Rule::SetDefault => "d",
    }
}

/// index_definition returns an index's CREATE INDEX statement as pg_get_indexdef prints it, naming its table with
/// the schema when asked.
pub fn index_definition(table: &TableDef, index: &TableIndex, qualified: bool) -> String {
    let rendering = index.vector.and_then(|distance| {
        crate::extensions::vector_rendering(distance, table.columns[*index.columns.first()?].ty.oid)
    });
    let columns: Vec<String> = index
        .columns
        .iter()
        .enumerate()
        .map(|(i, &c)| {
            let mut column = crate::engine::quote_identifier(&table.columns[c].name);
            match (index.descending[i], index.nulls_first[i]) {
                (true, true) => column.push_str(" DESC"),
                (true, false) => column.push_str(" DESC NULLS LAST"),
                (false, true) => column.push_str(" NULLS FIRST"),
                (false, false) => {}
            }
            if let Some((_, class)) = &rendering {
                column.push(' ');
                column.push_str(class);
            }
            column
        })
        .collect();
    let schema = if qualified { format!("{}.", crate::engine::quote_identifier(&table.schema)) } else { String::new() };
    format!(
        "CREATE {}INDEX {} ON {schema}{} USING {} ({})",
        if index.unique { "UNIQUE " } else { "" },
        crate::engine::quote_identifier(&index.name),
        crate::engine::quote_identifier(&table.name),
        rendering.as_ref().map_or("btree", |(method, _)| method),
        columns.join(", ")
    )
}
