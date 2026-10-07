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

//! User-defined types: enums, composites, and domains, stored with their array types as Go's type root objects, and
//! the registry that makes the types of the working root known to the code that converts, prints, and stores values.

use std::cell::RefCell;
use std::collections::HashMap;
use std::sync::Arc;

use doltdb::database::Database;
use doltdb::root::Root;
use objects::{CompositeAttribute, EnumLabel, SerializedType, TypeCheck};
use store::Hash;

use crate::catalog::id::{self, SECTION_FUNCTION, SECTION_TYPE};
use crate::catalog::{ColumnType, builtin_type, builtin_type_by_id, oids};
use crate::error::{PgError, Result, code};
use crate::query::Ctx;

/// COLLECTION is the position of the type collection among a root value's root object collections.
pub const COLLECTION: usize = 1;

/// SECTION_ENUM_LABEL is the ID section of enum labels.
const SECTION_ENUM_LABEL: u8 = 7;

/// Kind is what a user-defined type is.
#[derive(Clone, Debug, PartialEq)]
pub enum Kind {
    /// An enum, with its labels in sort order.
    Enum(Vec<String>),
    /// A composite type, with its attributes' names and types.
    Composite(Vec<(String, ColumnType)>),
    Domain(Domain),
    /// An array of the element type.
    Array(u32),
    /// A base type that an extension provides.
    Base(&'static crate::extensions::BaseType),
    /// A shell type, which CREATE TYPE names before defining it.
    Shell,
}

/// Domain is a type that restricts the values of its base type.
#[derive(Clone, Debug, PartialEq)]
pub struct Domain {
    pub base: ColumnType,
    pub not_null: bool,
    /// Each check's name and expression, which names the value `VALUE`.
    pub checks: Vec<(String, String)>,
    pub default: Option<String>,
}

/// UserType is a user-defined type with its definition as Go stores it.
#[derive(Clone, Debug, PartialEq)]
pub struct UserType {
    pub oid: u32,
    pub schema: String,
    pub name: String,
    pub kind: Kind,
    /// The OID of the type's array type, or 0 without one.
    pub array: u32,
    pub definition: SerializedType,
}

/// Types are user-defined types by OID.
pub type Types = HashMap<u32, Arc<UserType>>;

/// Registry is the user-defined types known to the statement running on this thread, with the search path that
/// unqualified type names resolve through.
#[derive(Default)]
struct Registry {
    types: Types,
    search_path: Vec<String>,
}

thread_local! {
    /// REGISTRY is the registry of the statement running on this thread.
    static REGISTRY: RefCell<Registry> = RefCell::new(Registry::default());
}

/// type_oid returns the OID of the type a stored type ID names.
pub fn type_oid(type_id: &[u8]) -> u32 {
    builtin_type_by_id(type_id).map_or_else(|| oids::oid(type_id), |t| t.oid)
}

/// type_id returns the stored ID of a type by OID, which is empty for an unknown type.
pub fn type_id(type_oid: u32) -> Vec<u8> {
    match builtin_type(type_oid) {
        Some(t) => t.definition.id.clone(),
        None => get(type_oid).map(|t| t.definition.id.clone()).unwrap_or_default(),
    }
}

impl UserType {
    /// from_definition reads a type from its stored definition.
    pub fn from_definition(definition: SerializedType) -> UserType {
        let mut segments = id::segments(&definition.id).into_iter();
        let schema = segments.next().unwrap_or_default();
        let name = segments.next().unwrap_or_default();
        let kind = match definition.typ_type.as_slice() {
            b"e" => {
                let mut labels: Vec<(f32, String)> = definition
                    .enum_labels
                    .iter()
                    .map(|l| (l.sort_order, id::segments(&l.id).pop().unwrap_or_default()))
                    .collect();
                labels.sort_by(|a, b| a.0.total_cmp(&b.0));
                Kind::Enum(labels.into_iter().map(|(_, label)| label).collect())
            }
            b"c" => Kind::Composite(
                definition
                    .composite_attrs
                    .iter()
                    .map(|a| {
                        let ty = ColumnType { oid: type_oid(&a.type_id), modifier: -1 };
                        (String::from_utf8_lossy(&a.name).into_owned(), ty)
                    })
                    .collect(),
            ),
            b"d" => Kind::Domain(Domain {
                base: ColumnType { oid: type_oid(&definition.base_type), modifier: definition.typ_mod },
                not_null: definition.not_null,
                checks: definition
                    .checks
                    .iter()
                    .map(|c| {
                        (
                            String::from_utf8_lossy(&c.name).into_owned(),
                            String::from_utf8_lossy(&c.expression).into_owned(),
                        )
                    })
                    .collect(),
                default: (!definition.default.is_empty())
                    .then(|| String::from_utf8_lossy(&definition.default).into_owned()),
            }),
            b"p" => Kind::Shell,
            _ if definition.elem.is_empty() => {
                let send = id::segments(&definition.send_func).into_iter().nth(1).unwrap_or_default();
                match crate::extensions::base_type(&send) {
                    Some(base) => Kind::Base(base),
                    None => Kind::Array(0),
                }
            }
            _ => Kind::Array(type_oid(&definition.elem)),
        };
        UserType {
            oid: oids::oid(&definition.id),
            schema,
            name,
            kind,
            array: if definition.array.is_empty() { 0 } else { type_oid(&definition.array) },
            definition,
        }
    }

    /// is_array reports whether the type is an array type.
    pub fn is_array(&self) -> bool {
        matches!(self.kind, Kind::Array(_))
    }
}

/// get returns a user-defined type known to this thread's statement by OID.
pub fn get(type_oid: u32) -> Option<Arc<UserType>> {
    REGISTRY.with(|r| r.borrow().types.get(&type_oid).cloned())
}

/// register makes a type known to this thread's statement, as a table's column definitions do, returning its OID.
pub fn register(definition: SerializedType) -> u32 {
    let user_type = UserType::from_definition(definition);
    let oid = user_type.oid;
    REGISTRY.with(|r| {
        r.borrow_mut().types.entry(oid).or_insert_with(|| Arc::new(user_type));
    });
    oid
}

/// register_row_type makes a table's row type and its array type known to this thread's statement, returning the row
/// type's OID.
pub fn register_row_type(table: &crate::catalog::table::TableDef) -> u32 {
    let definition = row_type(table);
    register(array_type(&definition));
    register(definition)
}

/// table_row_type returns the OID of the registered row type of the table with the OID.
pub fn table_row_type(table_oid: u32) -> Option<u32> {
    REGISTRY.with(|r| {
        r.borrow()
            .types
            .values()
            .find(|t| !t.definition.rel_id.is_empty() && oids::oid(&t.definition.rel_id) == table_oid)
            .map(|t| t.oid)
    })
}

/// install makes the types and search path those of the statement starting on this thread.
pub fn install(types: &Types, search_path: Vec<String>) {
    REGISTRY.with(|r| *r.borrow_mut() = Registry { types: types.clone(), search_path });
}

/// lookup finds a type by name, in the schema or else in the search path's schemas in order.
pub fn lookup(schema: Option<&str>, name: &str) -> Option<Arc<UserType>> {
    REGISTRY.with(|r| {
        let registry = r.borrow();
        let find = |schema: &str| registry.types.values().find(|t| t.schema == schema && t.name == name).cloned();
        match schema {
            Some(schema) => find(schema),
            None => registry.search_path.iter().find_map(|s| find(s)),
        }
    })
}

/// in_search_path reports whether a schema is on the search path of the statement running on this thread.
pub fn in_search_path(schema: &str) -> bool {
    REGISTRY.with(|r| r.borrow().search_path.iter().any(|s| s == schema))
}

/// base_type returns the type whose values a type holds: a domain's base type, and any other type itself.
pub fn base_type(ty: ColumnType) -> ColumnType {
    match get(ty.oid).map(|t| t.kind.clone()) {
        Some(Kind::Domain(domain)) => base_type(domain.base),
        _ => ty,
    }
}

/// all returns the user-defined types of a root value.
fn all(db: &mut Database, root: &Root) -> Result<Types> {
    let mut types = Types::new();
    for (_, address) in root.objects(db, COLLECTION)? {
        let user_type = UserType::from_definition(SerializedType::deserialize(&prolly::read_blob(db, &address)?)?);
        types.insert(user_type.oid, Arc::new(user_type));
    }
    Ok(types)
}

/// store writes a type into a root value.
pub fn store(db: &mut Database, root: &mut Root, definition: &SerializedType) -> Result<()> {
    let data = definition.serialize();
    let mut sink = |_: Hash, bytes: &[u8]| {
        db.write_value(bytes.to_vec()).map(|_| ()).map_err(|e| store::Error::Corrupt(e.to_string()))
    };
    let (address, _) = prolly::write_blob(&data, &mut sink)?.ok_or_else(|| PgError::internal("an empty type"))?;
    root.put_object(db, COLLECTION, &definition.id, Some(address))?;
    Ok(())
}

/// function_ref returns the stored reference to a pg_catalog function with parameters of the named types.
fn function_ref(name: &str, params: &[&str]) -> Vec<u8> {
    let types: Vec<String> = params
        .iter()
        .map(|p| String::from_utf8_lossy(&id::new(SECTION_TYPE, &["pg_catalog", p])).into_owned())
        .collect();
    let mut segments = vec!["pg_catalog", name];
    segments.extend(types.iter().map(String::as_str));
    id::new(SECTION_FUNCTION, &segments)
}

/// new_type returns a type definition with Go's settings that every kind of type shares.
fn new_type(schema: &str, name: &str) -> SerializedType {
    SerializedType {
        version: 0,
        id: id::new(SECTION_TYPE, &[schema, name]),
        typ_length: -1,
        passed_by_val: false,
        typ_type: Vec::new(),
        typ_category: Vec::new(),
        is_preferred: false,
        is_defined: true,
        delimiter: b",".to_vec(),
        rel_id: Vec::new(),
        subscript_func: Vec::new(),
        elem: Vec::new(),
        array: id::new(SECTION_TYPE, &[schema, &format!("_{name}")]),
        input_func: Vec::new(),
        output_func: Vec::new(),
        receive_func: Vec::new(),
        send_func: Vec::new(),
        mod_in_func: Vec::new(),
        mod_out_func: Vec::new(),
        analyze_func: Vec::new(),
        align: b"i".to_vec(),
        storage: b"p".to_vec(),
        not_null: false,
        base_type: Vec::new(),
        typ_mod: -1,
        n_dims: 0,
        typ_collation: Vec::new(),
        default_bin: Vec::new(),
        default: Vec::new(),
        acl: Vec::new(),
        checks: Vec::new(),
        att_typ_mod: -1,
        compare_func: Vec::new(),
        enum_labels: Vec::new(),
        composite_attrs: Vec::new(),
        internal_name: Vec::new(),
    }
}

/// enum_type returns the definition Go stores for an enum with the labels in order.
pub fn enum_type(schema: &str, name: &str, labels: &[String]) -> SerializedType {
    let mut t = new_type(schema, name);
    t.typ_length = 4;
    t.passed_by_val = true;
    t.typ_type = b"e".to_vec();
    t.typ_category = b"E".to_vec();
    t.input_func = function_ref("enum_in", &["cstring", "oid"]);
    t.output_func = function_ref("enum_out", &["anyenum"]);
    t.receive_func = function_ref("enum_recv", &["internal", "oid"]);
    t.send_func = function_ref("enum_send", &["anyenum"]);
    t.compare_func = function_ref("enum_cmp", &["anyenum", "anyenum"]);
    let type_id = String::from_utf8_lossy(&t.id).into_owned();
    t.enum_labels = labels
        .iter()
        .enumerate()
        .map(|(i, label)| EnumLabel { id: id::new(SECTION_ENUM_LABEL, &[&type_id, label]), sort_order: (i + 1) as f32 })
        .collect();
    t
}

/// extension_type returns the definition Go stores for a base type that an extension provides, whose support routines
/// share its schema and are named after it, as pgvector's are.
pub fn extension_type(schema: &str, name: &str) -> SerializedType {
    let own = String::from_utf8_lossy(&id::new(SECTION_TYPE, &[schema, name])).into_owned();
    let builtin = |n: &str| String::from_utf8_lossy(&id::new(SECTION_TYPE, &["pg_catalog", n])).into_owned();
    let routine = |suffix: &str, params: &[&str]| {
        let routine_name = format!("{name}{suffix}");
        let mut segments = vec![schema, &routine_name];
        segments.extend_from_slice(params);
        id::new(SECTION_FUNCTION, &segments)
    };
    let (cstring, oid, int4, internal) = (builtin("cstring"), builtin("oid"), builtin("int4"), builtin("internal"));
    let mut t = new_type(schema, name);
    t.typ_type = b"b".to_vec();
    t.typ_category = b"U".to_vec();
    t.storage = b"e".to_vec();
    t.input_func = routine("_in", &[&cstring, &oid, &int4]);
    t.output_func = routine("_out", &[&own]);
    t.receive_func = routine("_recv", &[&internal, &oid, &int4]);
    t.send_func = routine("_send", &[&own]);
    t.mod_in_func = routine("_typmod_in", &[&builtin("_cstring")]);
    t.compare_func = routine("_cmp", &[&own, &own]);
    t
}

/// composite_type returns the definition Go stores for a composite type of the attributes.
pub fn composite_type(schema: &str, name: &str, attributes: &[(String, ColumnType)]) -> SerializedType {
    let mut t = new_type(schema, name);
    t.typ_type = b"c".to_vec();
    t.typ_category = b"C".to_vec();
    t.input_func = function_ref("record_in", &["cstring", "oid", "int4"]);
    t.output_func = function_ref("record_out", &["record"]);
    t.receive_func = function_ref("record_recv", &["internal", "oid", "int4"]);
    t.send_func = function_ref("record_send", &["record"]);
    t.compare_func = function_ref("btrecordcmp", &["record", "record"]);
    t.align = b"d".to_vec();
    t.storage = b"x".to_vec();
    t.composite_attrs = attributes
        .iter()
        .enumerate()
        .map(|(i, (attribute, ty))| CompositeAttribute {
            rel_id: Vec::new(),
            name: attribute.clone().into_bytes(),
            type_id: type_id(ty.oid),
            num: (i + 1) as i16,
            collation: Vec::new(),
        })
        .collect();
    t
}

/// domain_type returns the definition Go stores for a domain over the base type.
pub fn domain_type(schema: &str, name: &str, base: ColumnType, domain: &Domain) -> Result<SerializedType> {
    let base_definition = match builtin_type(base.oid) {
        Some(t) => t.definition.clone(),
        None => get(base.oid).map(|t| t.definition.clone()).ok_or_else(|| PgError::internal("an unknown base type"))?,
    };
    let mut t = new_type(schema, name);
    t.typ_length = base_definition.typ_length;
    t.passed_by_val = base_definition.passed_by_val;
    t.typ_type = b"d".to_vec();
    t.typ_category = base_definition.typ_category.clone();
    t.is_preferred = base_definition.is_preferred;
    t.input_func = function_ref("domain_in", &["cstring", "oid", "int4"]);
    t.output_func = base_definition.output_func.clone();
    t.receive_func = function_ref("domain_recv", &["internal", "oid", "int4"]);
    t.send_func = base_definition.send_func.clone();
    t.mod_in_func = base_definition.mod_in_func.clone();
    t.mod_out_func = base_definition.mod_out_func.clone();
    t.align = base_definition.align.clone();
    t.storage = base_definition.storage.clone();
    t.not_null = domain.not_null;
    t.base_type = base_definition.id.clone();
    t.default = domain.default.clone().unwrap_or_default().into_bytes();
    t.checks = domain
        .checks
        .iter()
        .map(|(check, expression)| TypeCheck {
            name: check.clone().into_bytes(),
            expression: expression.clone().into_bytes(),
        })
        .collect();
    t.compare_func = base_definition.compare_func.clone();
    Ok(t)
}

/// transient returns the OID of a composite type of the columns that only this thread's statement knows, which names
/// the fields of a row that no declared type describes, such as a whole-row reference or a PL/pgSQL record.
pub fn transient(name: &str, columns: &[(String, ColumnType)]) -> u32 {
    let shape: Vec<String> = columns.iter().map(|(n, t)| format!("{n}:{}", t.oid)).collect();
    let mut definition = composite_type("", name, columns);
    definition.id = id::new(SECTION_TYPE, &["", name, &shape.join(",")]);
    register(definition)
}

/// row_type returns the definition Go stores for the row type of a table, a composite type tied to the table.
pub fn row_type(table: &crate::catalog::table::TableDef) -> SerializedType {
    let columns: Vec<(String, ColumnType)> = table.columns.iter().map(|c| (c.name.clone(), c.ty)).collect();
    let mut definition = composite_type(&table.schema, &table.name, &columns);
    definition.rel_id = id::new(crate::catalog::id::SECTION_TABLE, &[&table.schema, &table.name]);
    for attribute in &mut definition.composite_attrs {
        attribute.rel_id = definition.rel_id.clone();
    }
    definition
}

/// shell_type returns the definition Go stores for a shell type.
fn shell_type(schema: &str, name: &str) -> SerializedType {
    let mut t = new_type(schema, name);
    t.typ_length = 4;
    t.passed_by_val = true;
    t.typ_type = b"p".to_vec();
    t.typ_category = b"P".to_vec();
    t.is_defined = false;
    t.array = Vec::new();
    t.input_func = function_ref("shell_in", &["cstring"]);
    t.output_func = function_ref("shell_out", &["void"]);
    t
}

/// array_type returns the definition Go stores for the array type of a type.
pub fn array_type(base: &SerializedType) -> SerializedType {
    let mut segments = id::segments(&base.id).into_iter();
    let schema = segments.next().unwrap_or_default();
    let name = segments.next().unwrap_or_default();
    let mut t = new_type(&schema, &format!("_{name}"));
    t.version = 1;
    t.typ_type = b"b".to_vec();
    t.typ_category = b"A".to_vec();
    t.subscript_func = function_ref("array_subscript_handler", &["internal"]);
    t.elem = base.id.clone();
    t.array = Vec::new();
    t.input_func = function_ref("array_in", &["cstring", "oid", "int4"]);
    t.output_func = function_ref("array_out", &["anyarray"]);
    t.receive_func = function_ref("array_recv", &["internal", "oid", "int4"]);
    t.send_func = function_ref("array_send", &["anyarray"]);
    t.mod_in_func = base.mod_in_func.clone();
    t.mod_out_func = base.mod_out_func.clone();
    t.analyze_func = function_ref("array_typanalyze", &["internal"]);
    t.align = if base.align == b"d" { b"d".to_vec() } else { b"i".to_vec() };
    t.storage = b"x".to_vec();
    t.typ_collation = base.typ_collation.clone();
    t.att_typ_mod = base.att_typ_mod;
    t.compare_func = function_ref("btarraycmp", &["anyarray", "anyarray"]);
    t.internal_name = format!("{name}[]").into_bytes();
    t
}

impl Ctx<'_> {
    /// user_types returns the user-defined types of the working root, reusing them while the type collection is
    /// unchanged.
    pub fn user_types(&mut self) -> Result<Arc<Types>> {
        let address = self.txn.root.root_objects[COLLECTION];
        if let Some((cached, types)) = &self.session.user_types
            && *cached == address
        {
            return Ok(types.clone());
        }
        let types = Arc::new(all(self.db, &self.txn.root)?);
        self.session.user_types = Some((address, types.clone()));
        Ok(types)
    }

    /// prepare_type makes the row type of the table that a type name names known to this thread, so that resolving the
    /// name finds it.
    pub fn prepare_type(&mut self, type_name: &pg_query::protobuf::TypeName) -> Result<()> {
        let names: Vec<&str> = type_name.names.iter().filter_map(crate::expr::node_name).collect();
        let (schema, name) = match names.as_slice() {
            [name] => (None, *name),
            [schema, name] | [_, schema, name] => (Some(*schema), *name),
            _ => return Ok(()),
        };
        if schema.is_none_or(|s| s == "pg_catalog") && crate::catalog::builtin_type_named(name).is_some()
            || lookup(schema, name).is_some()
        {
            return Ok(());
        }
        let schemas = match schema {
            Some(schema) => vec![schema.to_string()],
            None => self.session.search_path(),
        };
        for schema in schemas {
            if let Some(table) = self.txn.table(self.db, &schema, name)? {
                register_row_type(&table);
                return Ok(());
            }
        }
        Ok(())
    }

    /// install_types makes the working root's user-defined types and the search path known to this thread.
    pub fn install_types(&mut self) -> Result<()> {
        let types = self.user_types()?;
        install(&types, self.session.search_path());
        Ok(())
    }
}

/// type_names returns the schema, which is empty when unqualified, and name of a qualified type name.
fn type_names(nodes: &[pg_query::Node]) -> (String, String) {
    let names: Vec<&str> = nodes.iter().filter_map(crate::expr::node_name).collect();
    match names.as_slice() {
        [name] => (String::new(), name.to_string()),
        [.., schema, name] => (schema.to_string(), name.to_string()),
        [] => (String::new(), String::new()),
    }
}

impl Ctx<'_> {
    /// new_type_schema returns the schema a new type of the name goes in, failing as Postgres does when a type or a
    /// table's row type already has the name.
    fn new_type_schema(&mut self, schema: &str, name: &str) -> Result<String> {
        let schema = self.target_schema(schema, -1)?;
        let existing = self.user_types()?.values().find(|t| t.schema == schema && t.name == name).cloned();
        if let Some(array) = existing.as_ref().filter(|t| t.is_array()) {
            self.move_array_type(array)?;
        }
        let taken =
            self.user_types()?.values().any(|t| t.schema == schema && t.name == name && !matches!(t.kind, Kind::Shell))
                || crate::catalog::builtin_type_named(name).is_some() && schema == "pg_catalog"
                || self.txn.root.table(self.db, &schema, name)?.is_some();
        if taken {
            return Err(PgError::new(code::DUPLICATE_OBJECT, format!("type \"{name}\" already exists")));
        }
        Ok(schema)
    }

    /// move_array_type gives an array type the next name that adds underscores to its own, freeing its name for a new
    /// type, as Postgres' moveArrayTypeName does.
    fn move_array_type(&mut self, array: &UserType) -> Result<()> {
        let types = self.user_types()?;
        let mut name = format!("_{}", array.name);
        while types.values().any(|t| t.schema == array.schema && t.name == name) {
            name.insert(0, '_');
        }
        let mut moved = array.definition.clone();
        moved.id = id::new(SECTION_TYPE, &[&array.schema, &name]);
        moved.internal_name = Vec::new();
        self.txn.root.put_object(self.db, COLLECTION, &array.definition.id, None)?;
        store(self.db, &mut self.txn.root, &moved)?;
        if let Some(element) = types.values().find(|t| t.definition.array == array.definition.id) {
            let mut definition = element.definition.clone();
            definition.array = moved.id.clone();
            store(self.db, &mut self.txn.root, &definition)?;
        }
        Ok(())
    }

    /// create_shell runs CREATE TYPE with only a name, which records a shell type to define later.
    pub fn create_shell(&mut self, define: &pg_query::protobuf::DefineStmt) -> Result<crate::Outcome> {
        let (schema, name) = type_names(&define.defnames);
        let schema = self.new_type_schema(&schema, &name)?;
        store(self.db, &mut self.txn.root, &shell_type(&schema, &name))?;
        Ok(crate::Outcome::command("CREATE TYPE"))
    }

    /// store_type writes a type and its array type to the working root.
    pub(crate) fn store_type(&mut self, definition: SerializedType) -> Result<()> {
        let array = array_type(&definition);
        store(self.db, &mut self.txn.root, &definition)?;
        store(self.db, &mut self.txn.root, &array)?;
        Ok(())
    }

    /// create_enum runs CREATE TYPE ... AS ENUM.
    pub fn create_enum(&mut self, stmt: &pg_query::protobuf::CreateEnumStmt) -> Result<crate::Outcome> {
        let (schema, name) = type_names(&stmt.type_name);
        let schema = self.new_type_schema(&schema, &name)?;
        let labels: Vec<String> = stmt.vals.iter().filter_map(crate::expr::node_name).map(str::to_string).collect();
        for (i, label) in labels.iter().enumerate() {
            check_label(label)?;
            if labels[..i].contains(label) {
                let oid = oids::oid(&id::new(SECTION_TYPE, &[&schema, &name]));
                return Err(PgError {
                    detail: Some(format!("Key (enumtypid, enumlabel)=({oid}, {label}) already exists.")),
                    objects: Some(Box::new(crate::error::ErrorObjects {
                        schema: Some("pg_catalog".into()),
                        table: Some("pg_enum".into()),
                        constraint: Some("pg_enum_typid_label_index".into()),
                        ..Default::default()
                    })),
                    ..PgError::new(
                        code::UNIQUE_VIOLATION,
                        "duplicate key value violates unique constraint \"pg_enum_typid_label_index\"",
                    )
                });
            }
        }
        self.store_type(enum_type(&schema, &name, &labels))?;
        Ok(crate::Outcome::command("CREATE TYPE"))
    }

    /// create_composite runs CREATE TYPE ... AS (...).
    pub fn create_composite(&mut self, stmt: &pg_query::protobuf::CompositeTypeStmt) -> Result<crate::Outcome> {
        let relation = stmt.typevar.as_ref().ok_or_else(|| PgError::internal("CREATE TYPE without a name"))?;
        let schema = self.new_type_schema(&relation.schemaname, &relation.relname)?;
        let mut attributes = Vec::with_capacity(stmt.coldeflist.len());
        for node in &stmt.coldeflist {
            let Some(pg_query::NodeEnum::ColumnDef(column)) = node.node.as_ref() else { continue };
            let type_name =
                column.type_name.as_ref().ok_or_else(|| PgError::internal("an attribute without a type"))?;
            self.prepare_type(type_name)?;
            let ty = crate::expr::resolve_type_name(type_name).map_err(|err| PgError { position: None, ..err })?;
            if ty.oid == crate::oid::RECORD {
                return Err(PgError::new(
                    code::INVALID_TABLE_DEFINITION,
                    format!("column \"{}\" has pseudo-type record", column.colname),
                ));
            }
            if attributes.iter().any(|(n, _)| *n == column.colname) {
                return Err(PgError::new(
                    code::DUPLICATE_COLUMN,
                    format!("column \"{}\" specified more than once", column.colname),
                ));
            }
            attributes.push((column.colname.clone(), ty));
        }
        self.store_type(composite_type(&schema, &relation.relname, &attributes))?;
        Ok(crate::Outcome::command("CREATE TYPE"))
    }

    /// create_domain runs CREATE DOMAIN, checking its constraints as Postgres does.
    pub fn create_domain(&mut self, stmt: &pg_query::protobuf::CreateDomainStmt) -> Result<crate::Outcome> {
        use pg_query::protobuf::ConstrType;
        let (schema, name) = type_names(&stmt.domainname);
        let schema = self.new_type_schema(&schema, &name)?;
        let type_name = stmt.type_name.as_ref().ok_or_else(|| PgError::internal("CREATE DOMAIN without a type"))?;
        self.prepare_type(type_name)?;
        let base = crate::expr::resolve_type_name(type_name).map_err(|err| PgError { position: None, ..err })?;
        if base.oid == crate::oid::RECORD {
            return Err(PgError::new(code::DATATYPE_MISMATCH, "\"record\" is not a valid base type for a domain"));
        }
        let mut domain = Domain { base, not_null: false, checks: Vec::new(), default: None };
        let mut nullability = None;
        for node in &stmt.constraints {
            let Some(pg_query::NodeEnum::Constraint(constraint)) = node.node.as_ref() else { continue };
            match ConstrType::try_from(constraint.contype).unwrap_or(ConstrType::Undefined) {
                kind @ (ConstrType::ConstrNotnull | ConstrType::ConstrNull) => {
                    let not_null = kind == ConstrType::ConstrNotnull;
                    if nullability.is_some_and(|n| n != not_null) {
                        return Err(PgError::new(code::SYNTAX_ERROR, "conflicting NULL/NOT NULL constraints"));
                    }
                    nullability = Some(not_null);
                    domain.not_null = not_null;
                }
                ConstrType::ConstrDefault => {
                    if domain.default.is_some() {
                        return Err(PgError::new(code::SYNTAX_ERROR, "multiple default expressions"));
                    }
                    let expr = constraint.raw_expr.as_deref().ok_or_else(|| PgError::internal("a default"))?;
                    let mut binder = crate::expr::Binder::new(self, crate::expr::Scope::default());
                    binder.clause = "DEFAULT expressions";
                    binder.definition = true;
                    let bound = binder.bind(expr)?;
                    crate::expr::coerce(bound, base, false, crate::expr::arg_location(expr))?;
                    domain.default = Some(crate::ddl::expression_text(expr)?);
                }
                ConstrType::ConstrCheck => {
                    let expr = constraint.raw_expr.as_deref().ok_or_else(|| PgError::internal("a check"))?;
                    self.bind_domain_check(expr, base).map_err(|err| PgError { position: None, ..err })?;
                    let check = if constraint.conname.is_empty() {
                        let count =
                            domain.checks.iter().filter(|(n, _)| n.starts_with(&format!("{name}_check"))).count();
                        if count == 0 { format!("{name}_check") } else { format!("{name}_check{count}") }
                    } else {
                        constraint.conname.clone()
                    };
                    domain.checks.push((check, crate::ddl::expression_text(expr)?));
                }
                _ => return Err(PgError::unsupported("this domain constraint")),
            }
        }
        self.store_type(domain_type(&schema, &name, base, &domain)?)?;
        Ok(crate::Outcome::command("CREATE DOMAIN"))
    }

    /// bind_domain_check binds a domain check over its value, failing unless it is boolean.
    fn bind_domain_check(&mut self, expr: &pg_query::Node, base: ColumnType) -> Result<crate::expr::Expr> {
        let scope = crate::expr::Scope {
            columns: vec![crate::expr::ScopeColumn {
                table: String::new(),
                name: "value".into(),
                ty: base,
                hidden: false,
                origin: (0, 0),
            }],
        };
        let mut binder = crate::expr::Binder::new(self, scope);
        binder.clause = "check constraints";
        binder.definition = true;
        let (bound, ty) = binder.bind(expr)?;
        if ty.oid != crate::oid::BOOL {
            return Err(PgError::new(
                code::DATATYPE_MISMATCH,
                format!("argument of CHECK must be type boolean, not type {}", crate::cast::type_display(ty.oid)),
            ));
        }
        Ok(bound)
    }

    /// check_domain fails as Postgres does when a value is not a value of a domain, checking each domain a domain is
    /// based on.
    pub fn check_domain(&mut self, value: &crate::types::Value, ty: ColumnType) -> Result<()> {
        let Some(user_type) = get(ty.oid) else { return Ok(()) };
        let Kind::Domain(domain) = &user_type.kind else { return Ok(()) };
        self.check_domain(value, domain.base)?;
        let objects = || {
            Some(Box::new(crate::error::ErrorObjects {
                schema: Some(user_type.schema.clone()),
                data_type: Some(user_type.name.clone()),
                ..Default::default()
            }))
        };
        if value.is_null() && domain.not_null {
            return Err(PgError {
                objects: objects(),
                ..PgError::new(
                    code::NOT_NULL_VIOLATION,
                    format!("domain {} does not allow null values", user_type.name),
                )
            });
        }
        for (name, text) in &domain.checks {
            let node = crate::parse::expression_node(text)?;
            let check = self.bind_domain_check(&node, domain.base)?;
            if matches!(check.eval(self, std::slice::from_ref(value))?, crate::types::Value::Bool(false)) {
                let mut objects = objects();
                if let Some(objects) = objects.as_mut() {
                    objects.constraint = Some(name.clone());
                }
                return Err(PgError {
                    objects,
                    ..PgError::new(
                        code::CHECK_VIOLATION,
                        format!("value for domain {} violates check constraint \"{name}\"", user_type.name),
                    )
                });
            }
        }
        Ok(())
    }

    /// drop_types runs DROP TYPE or DROP DOMAIN, resolving every type first so that a missing one drops none.
    pub fn drop_types(&mut self, drop: &pg_query::protobuf::DropStmt, domains: bool) -> Result<crate::Outcome> {
        let mut doomed = Vec::new();
        for object in &drop.objects {
            let names: Vec<pg_query::Node> = match object.node.as_ref() {
                Some(pg_query::NodeEnum::TypeName(t)) => t.names.clone(),
                Some(pg_query::NodeEnum::List(list)) => list.items.clone(),
                _ => continue,
            };
            let (schema, name) = type_names(&names);
            let found = lookup((!schema.is_empty()).then_some(schema.as_str()), &name).filter(|t| !t.is_array());
            let Some(user_type) = found else {
                let shown = if schema.is_empty() { name.clone() } else { format!("{schema}.{name}") };
                if drop.missing_ok {
                    self.session.notice(PgError::notice("00000", format!("type \"{shown}\" does not exist, skipping")));
                    continue;
                }
                return Err(undefined(&shown));
            };
            if domains && !matches!(user_type.kind, Kind::Domain(_)) {
                return Err(PgError::new(code::WRONG_OBJECT_TYPE, format!("\"{name}\" is not a domain")));
            }
            let dependents = self.type_dependents(&user_type)?;
            if !dependents.is_empty() {
                return Err(PgError {
                    detail: Some(dependents.join("\n")),
                    hint: Some("Use DROP ... CASCADE to drop the dependent objects too.".into()),
                    ..PgError::new(
                        code::DEPENDENT_OBJECTS_STILL_EXIST,
                        format!("cannot drop type {} because other objects depend on it", user_type.name),
                    )
                });
            }
            doomed.push(user_type);
        }
        for user_type in doomed {
            self.txn.root.put_object(self.db, COLLECTION, &user_type.definition.id, None)?;
            if !user_type.definition.array.is_empty() {
                self.txn.root.put_object(self.db, COLLECTION, &user_type.definition.array, None)?;
            }
        }
        Ok(crate::Outcome::command(if domains { "DROP DOMAIN" } else { "DROP TYPE" }))
    }

    /// type_dependents describes the table columns and types that use a type, or its array type.
    fn type_dependents(&mut self, user_type: &UserType) -> Result<Vec<String>> {
        let uses = |oid: u32| oid == user_type.oid || (user_type.array != 0 && oid == user_type.array);
        let mut dependents = Vec::new();
        for (key, address) in self.txn.root.tables(self.db)? {
            let text = String::from_utf8_lossy(&key).into_owned();
            let mut parts = text.splitn(3, '\0').skip(1);
            let (Some(schema), Some(table)) = (parts.next(), parts.next()) else { continue };
            let table = crate::catalog::table::TableDef::load(self.db, schema, table, address)?;
            for column in table.columns.iter().filter(|c| uses(c.ty.oid)) {
                let array = column.ty.oid == user_type.array;
                let shown = if array { format!("{}[]", user_type.name) } else { user_type.name.clone() };
                let dependent = format!("column {} of table {} depends on type {shown}", column.name, table.name);
                if array { dependents.insert(0, dependent) } else { dependents.push(dependent) }
            }
        }
        for other in self.user_types()?.values() {
            let depends = match &other.kind {
                Kind::Composite(attributes) => attributes.iter().any(|(_, t)| uses(t.oid)),
                Kind::Domain(domain) => uses(domain.base.oid),
                _ => false,
            };
            if depends && other.oid != user_type.oid {
                let kind = if matches!(other.kind, Kind::Domain(_)) { "type" } else { "composite type" };
                dependents.push(format!("{kind} {} depends on type {}", other.name, user_type.name));
            }
        }
        Ok(dependents)
    }

    /// alter_enum runs ALTER TYPE ... ADD VALUE and ALTER TYPE ... RENAME VALUE.
    pub fn alter_enum(&mut self, stmt: &pg_query::protobuf::AlterEnumStmt) -> Result<crate::Outcome> {
        let (schema, name) = type_names(&stmt.type_name);
        let user_type =
            lookup((!schema.is_empty()).then_some(schema.as_str()), &name).ok_or_else(|| undefined(&name))?;
        let Kind::Enum(labels) = &user_type.kind else {
            return Err(PgError::new(code::WRONG_OBJECT_TYPE, format!("\"{name}\" is not an enum")));
        };
        let missing = |label: &str| {
            PgError::new(code::INVALID_PARAMETER_VALUE, format!("\"{label}\" is not an existing enum label"))
        };
        let mut definition = user_type.definition.clone();
        let type_id = String::from_utf8_lossy(&definition.id).into_owned();
        if !stmt.old_val.is_empty() {
            if !labels.contains(&stmt.old_val) {
                return Err(missing(&stmt.old_val));
            }
            if labels.contains(&stmt.new_val) {
                return Err(PgError::new(
                    code::DUPLICATE_OBJECT,
                    format!("enum label \"{}\" already exists", stmt.new_val),
                ));
            }
            check_label(&stmt.new_val)?;
            for label in &mut definition.enum_labels {
                if id::segments(&label.id).pop().as_deref() == Some(stmt.old_val.as_str()) {
                    label.id = id::new(SECTION_ENUM_LABEL, &[&type_id, &stmt.new_val]);
                }
            }
        } else {
            if labels.contains(&stmt.new_val) {
                let message = format!("enum label \"{}\" already exists", stmt.new_val);
                if stmt.skip_if_new_val_exists {
                    self.session.notice(PgError::notice(code::DUPLICATE_OBJECT, format!("{message}, skipping")));
                    return Ok(crate::Outcome::command("ALTER TYPE"));
                }
                return Err(PgError::new(code::DUPLICATE_OBJECT, message));
            }
            check_label(&stmt.new_val)?;
            let mut orders: Vec<f32> = definition.enum_labels.iter().map(|l| l.sort_order).collect();
            orders.sort_by(f32::total_cmp);
            let order_of = |label: &str| {
                definition
                    .enum_labels
                    .iter()
                    .find(|l| id::segments(&l.id).pop().as_deref() == Some(label))
                    .map(|l| l.sort_order)
            };
            let sort_order = if stmt.new_val_neighbor.is_empty() {
                orders.last().map_or(1.0, |o| o + 1.0)
            } else {
                let neighbor = order_of(&stmt.new_val_neighbor).ok_or_else(|| missing(&stmt.new_val_neighbor))?;
                let position = orders.iter().position(|o| *o == neighbor).unwrap_or(0);
                if stmt.new_val_is_after {
                    orders.get(position + 1).map_or(neighbor + 1.0, |next| (neighbor + next) / 2.0)
                } else if position == 0 {
                    neighbor - 1.0
                } else {
                    (orders[position - 1] + neighbor) / 2.0
                }
            };
            definition
                .enum_labels
                .push(EnumLabel { id: id::new(SECTION_ENUM_LABEL, &[&type_id, &stmt.new_val]), sort_order });
        }
        store(self.db, &mut self.txn.root, &definition)?;
        Ok(crate::Outcome::command("ALTER TYPE"))
    }
}

/// check_label fails as Postgres does for an enum label that is too long.
fn check_label(label: &str) -> Result<()> {
    if label.len() > 63 {
        return Err(PgError {
            detail: Some("Labels must be 63 bytes or less.".into()),
            ..PgError::new(code::INVALID_NAME, format!("invalid enum label \"{label}\""))
        });
    }
    Ok(())
}

/// undefined returns Postgres' error for a type name that names no type.
pub fn undefined(name: &str) -> PgError {
    PgError::new(code::UNDEFINED_OBJECT, format!("type \"{name}\" does not exist"))
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::expr::typ;

    /// hex returns the bytes that hexadecimal text spells.
    fn hex(text: &str) -> Vec<u8> {
        (0..text.len()).step_by(2).map(|i| u8::from_str_radix(&text[i..i + 2], 16).unwrap()).collect()
    }

    #[test]
    fn definitions_match_go() {
        let mood = enum_type("public", "mood", &["sad".into(), "ok".into(), "happy".into()]);
        let go_mood = "000e230206047075626c69636d6f6f64800401016501450001012c0000000f230206057075626c69635f6d6f6f643d0f040a07151170675f636174616c6f67656e756d5f696e23020a0770675f636174616c6f6763737472696e6723020a0370675f636174616c6f676f69642c0f030a081570675f636174616c6f67656e756d5f6f757423020a0770675f636174616c6f67616e79656e756d400f040a09161170675f636174616c6f67656e756d5f7265637623020a0870675f636174616c6f67696e7465726e616c23020a0370675f636174616c6f676f69642d0f030a091570675f636174616c6f67656e756d5f73656e6423020a0770675f636174616c6f67616e79656e756d0000000169017000007fffffff8000000000000000007fffffff420f040a08151570675f636174616c6f67656e756d5f636d7023020a0770675f636174616c6f67616e79656e756d23020a0770675f636174616c6f67616e79656e756d031407020e02230206047075626c69636d6f6f646f6bc00000001507020e03230206047075626c69636d6f6f64736164bf8000001707020e05230206047075626c69636d6f6f646861707079c04000000000";
        assert_eq!(mood.serialize(), hex(go_mood));
        let go_mood_array = "010f230206057075626c69635f6d6f6f647fff00016201410001012c003c0f030a171670675f636174616c6f6761727261795f7375627363726970745f68616e646c657223020a0870675f636174616c6f67696e7465726e616c0e230206047075626c69636d6f6f6400510f050a0815111270675f636174616c6f6761727261795f696e23020a0770675f636174616c6f6763737472696e6723020a0370675f636174616c6f676f696423020a0470675f636174616c6f67696e74342e0f030a091670675f636174616c6f6761727261795f6f757423020a0870675f636174616c6f67616e796172726179540f050a0a16111270675f636174616c6f6761727261795f7265637623020a0870675f636174616c6f67696e7465726e616c23020a0370675f636174616c6f676f696423020a0470675f636174616c6f67696e74342f0f030a0a1670675f636174616c6f6761727261795f73656e6423020a0870675f636174616c6f67616e7961727261790000350f030a101670675f636174616c6f6761727261795f747970616e616c797a6523020a0870675f636174616c6f67696e7465726e616c0169017800007fffffff8000000000000000007fffffff460f040a0a161670675f636174616c6f6762746172726179636d7023020a0870675f636174616c6f67616e79617272617923020a0870675f636174616c6f67616e7961727261790000066d6f6f645b5d";
        assert_eq!(array_type(&mood).serialize(), hex(go_mood_array));
        let pair = composite_type(
            "public",
            "pair",
            &[("x".into(), typ(crate::oid::INT4)), ("y".into(), typ(crate::oid::TEXT))],
        );
        let go_pair = "000e230206047075626c6963706169727fff00016301430001012c0000000f230206057075626c69635f70616972520f050a0915111270675f636174616c6f677265636f72645f696e23020a0770675f636174616c6f6763737472696e6723020a0370675f636174616c6f676f696423020a0470675f636174616c6f67696e74342d0f030a0a1470675f636174616c6f677265636f72645f6f757423020a0670675f636174616c6f677265636f7264550f050a0b16111270675f636174616c6f677265636f72645f7265637623020a0870675f636174616c6f67696e7465726e616c23020a0370675f636174616c6f676f696423020a0470675f636174616c6f67696e74342e0f030a0b1470675f636174616c6f677265636f72645f73656e6423020a0670675f636174616c6f677265636f72640000000164017800007fffffff8000000000000000007fffffff430f040a0b141470675f636174616c6f6762747265636f7264636d7023020a0670675f636174616c6f677265636f726423020a0670675f636174616c6f677265636f726400020001781223020a0470675f636174616c6f67696e74348001000001791223020a0470675f636174616c6f677465787480020000";
        assert_eq!(pair.serialize(), hex(go_pair));
        let domain = Domain {
            base: typ(crate::oid::INT4),
            not_null: true,
            checks: vec![("posint_check".into(), "VALUE > 0".into())],
            default: Some("1".into()),
        };
        let posint = domain_type("public", "posint", typ(crate::oid::INT4), &domain).unwrap();
        let go_posint = "0010230206067075626c6963706f73696e748004010164014e0001012c00000011230206077075626c69635f706f73696e74520f050a0915111270675f636174616c6f67646f6d61696e5f696e23020a0770675f636174616c6f6763737472696e6723020a0370675f636174616c6f676f696423020a0470675f636174616c6f67696e7434280f030a071270675f636174616c6f67696e74346f757423020a0470675f636174616c6f67696e7434550f050a0b16111270675f636174616c6f67646f6d61696e5f7265637623020a0870675f636174616c6f67696e7465726e616c23020a0370675f636174616c6f676f696423020a0470675f636174616c6f67696e7434290f030a081270675f636174616c6f67696e743473656e6423020a0470675f636174616c6f67696e743400000001690170011223020a0470675f636174616c6f67696e74347fffffff800000000000013100010c706f73696e745f636865636b0956414c5545203e20307fffffff3d0f040a09121270675f636174616c6f676274696e7434636d7023020a0470675f636174616c6f67696e743423020a0470675f636174616c6f67696e7434000000";
        assert_eq!(posint.serialize(), hex(go_posint));
        assert_eq!(UserType::from_definition(posint).kind, Kind::Domain(domain));
    }
}
