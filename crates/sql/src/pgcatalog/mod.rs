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

//! The system catalogs: the relations of pg_catalog and information_schema, with Postgres 15's columns, and rows
//! that describe the working root value.

mod builtin;
pub(crate) mod definitions;
mod extensions;
mod infoschema;
pub mod reg;
pub(crate) mod routines;
mod rows;

pub(crate) use rows::row_type_oid;
pub mod snapshot;

use std::collections::HashMap;
use std::sync::OnceLock;

use crate::error::Result;
use crate::query::Ctx;
use crate::types::Value;

/// COLUMNS lists the columns of every system catalog relation as Postgres 15 defines them, one per line: schema,
/// relation, relation kind, relation OID, column, type OID (the base type of a domain), and whether it is NOT NULL.
const COLUMNS: &str = include_str!("columns.tsv");

/// CatalogColumn is a column of a system catalog relation.
#[derive(Debug, PartialEq)]
pub struct CatalogColumn {
    pub name: &'static str,
    pub type_oid: u32,
    pub not_null: bool,
}

/// CatalogTable is a system catalog relation.
#[derive(Debug, PartialEq)]
pub struct CatalogTable {
    pub schema: &'static str,
    pub name: &'static str,
    /// The relation kind, `r` for a table and `v` for a view.
    pub kind: &'static str,
    pub oid: u32,
    pub columns: Vec<CatalogColumn>,
}

impl CatalogTable {
    /// column returns the position of a column.
    fn column(&self, name: &str) -> Option<usize> {
        self.columns.iter().position(|c| c.name == name)
    }
}

/// Catalogs indexes the system catalog relations.
struct Catalogs {
    tables: Vec<CatalogTable>,
    by_name: HashMap<(&'static str, &'static str), usize>,
}

/// catalogs returns the system catalog relations, reading them on first use.
fn catalogs() -> &'static Catalogs {
    static CATALOGS: OnceLock<Catalogs> = OnceLock::new();
    CATALOGS.get_or_init(|| {
        let mut tables: Vec<CatalogTable> = Vec::new();
        for line in COLUMNS.lines() {
            let fields: Vec<&'static str> = line.split('\t').collect();
            let [schema, name, kind, oid, column, type_oid, not_null] = fields[..] else { continue };
            if tables.last().is_none_or(|t| t.schema != schema || t.name != name) {
                let oid = oid.parse().expect("a catalog relation OID");
                tables.push(CatalogTable { schema, name, kind, oid, columns: Vec::new() });
            }
            let type_oid = type_oid.parse().expect("a catalog column type OID");
            tables.last_mut().expect("a table was just added").columns.push(CatalogColumn {
                name: column,
                type_oid,
                not_null: not_null == "t",
            });
        }
        let by_name = tables.iter().enumerate().map(|(i, t)| ((t.schema, t.name), i)).collect();
        Catalogs { tables, by_name }
    })
}

/// lookup returns the system catalog relation with the schema and name.
pub fn lookup(schema: &str, name: &str) -> Option<&'static CatalogTable> {
    let c = catalogs();
    c.by_name.get(&(schema, name)).map(|&i| &c.tables[i])
}

/// builtin_view_definition returns the definition of a view of pg_catalog or information_schema, as pg_views shows
/// it, by its OID.
pub fn builtin_view_definition(oid: u32) -> Option<&'static str> {
    static VIEWS: OnceLock<HashMap<u32, String>> = OnceLock::new();
    let views = VIEWS.get_or_init(|| {
        let (Some(class), Some(views)) = (lookup("pg_catalog", "pg_class"), lookup("pg_catalog", "pg_views")) else {
            return HashMap::new();
        };
        let text = |v: &Value| v.output().unwrap_or_default();
        let (Some(schema), Some(view), Some(definition)) =
            (views.column("schemaname"), views.column("viewname"), views.column("definition"))
        else {
            return HashMap::new();
        };
        let mut definitions: HashMap<(String, String), String> =
            builtin::rows(views).iter().map(|r| ((text(&r[schema]), text(&r[view])), text(&r[definition]))).collect();
        let (Some(oid), Some(name), Some(namespace)) =
            (class.column("oid"), class.column("relname"), class.column("relnamespace"))
        else {
            return HashMap::new();
        };
        builtin::rows(class)
            .iter()
            .filter_map(|r| {
                let Value::Oid(oid) = r[oid] else { return None };
                let schema = if r[namespace] == Value::Oid(11) { "pg_catalog" } else { "information_schema" };
                Some((oid, definitions.remove(&(schema.to_string(), text(&r[name])))?))
            })
            .collect()
    });
    views.get(&oid).map(String::as_str)
}

/// is_immutable reports whether an expression calls only functions that have an immutable form, as pg_proc's
/// provolatile shows them.
pub fn is_immutable(expr: &pg_query::Node) -> bool {
    let Some(proc) = lookup("pg_catalog", "pg_proc") else { return true };
    let (Some(name), Some(volatility)) = (proc.column("proname"), proc.column("provolatile")) else { return true };
    let rows = builtin::rows(proc);
    let wrapped = pg_query::NodeEnum::ResTarget(Box::new(pg_query::protobuf::ResTarget {
        val: Some(Box::new(expr.clone())),
        ..Default::default()
    }));
    wrapped.nodes().into_iter().all(|(node, ..)| {
        let pg_query::NodeRef::FuncCall(call) = node else { return true };
        let Some(function) = call.funcname.iter().filter_map(crate::expr::node_name).next_back() else { return true };
        let forms: Vec<&Vec<Value>> = rows.iter().filter(|r| r[name] == Value::Text(function.to_string())).collect();
        forms.is_empty() || forms.iter().any(|r| r[volatility] == Value::Text("i".into()))
    })
}

/// is_volatile reports whether a built-in function has a volatile form, as pg_proc's provolatile shows them.
pub fn is_volatile(function: &str) -> bool {
    static VOLATILE: std::sync::OnceLock<std::collections::HashSet<String>> = OnceLock::new();
    VOLATILE
        .get_or_init(|| {
            let Some(proc) = lookup("pg_catalog", "pg_proc") else { return Default::default() };
            let (Some(name), Some(volatility)) = (proc.column("proname"), proc.column("provolatile")) else {
                return Default::default();
            };
            builtin::rows(proc)
                .iter()
                .filter(|r| r[volatility] == Value::Text("v".into()))
                .map(|r| r[name].output().unwrap_or_default())
                .collect()
        })
        .contains(function)
}

/// OperatorClass is a built-in operator class: its OID, name, input type, and whether it is its type's default.
#[derive(Clone, Debug)]
pub struct OperatorClass {
    pub oid: u32,
    pub name: String,
    pub input: u32,
    pub default: bool,
}

/// operator_classes returns the built-in operator classes of an index access method, by the method's OID.
pub fn operator_classes(method: u32) -> Vec<OperatorClass> {
    static CLASSES: OnceLock<Vec<(u32, OperatorClass)>> = OnceLock::new();
    let classes = CLASSES.get_or_init(|| {
        let Some(table) = lookup("pg_catalog", "pg_opclass") else { return Vec::new() };
        let columns = ["oid", "opcmethod", "opcname", "opcintype", "opcdefault"].map(|c| table.column(c));
        let [Some(oid), Some(opcmethod), Some(name), Some(input), Some(default)] = columns else { return Vec::new() };
        let number = |v: &Value| v.output().and_then(|t| t.parse().ok()).unwrap_or(0);
        builtin::rows(table)
            .iter()
            .map(|r| {
                let class = OperatorClass {
                    oid: number(&r[oid]),
                    name: r[name].output().unwrap_or_default(),
                    input: number(&r[input]),
                    default: r[default] == Value::Bool(true),
                };
                (number(&r[opcmethod]), class)
            })
            .collect()
    });
    classes.iter().filter(|(m, _)| *m == method).map(|(_, c)| c.clone()).collect()
}

/// default_operator_class returns a type's default btree operator class, preferring one for the type itself, then
/// one for a preferred type it coerces to without conversion, as Postgres' GetDefaultOpClass does.
pub fn default_operator_class(type_oid: u32) -> Option<OperatorClass> {
    let defaults: Vec<OperatorClass> = operator_classes(403).into_iter().filter(|c| c.default).collect();
    let exact = defaults.iter().find(|c| c.input == type_oid);
    let coercible: Vec<&OperatorClass> = defaults.iter().filter(|c| binary_coercible(type_oid, c.input)).collect();
    let preferred = coercible.iter().find(|c| crate::functions::is_preferred(c.input)).or(coercible.first());
    exact.or(preferred.copied()).cloned()
}

/// binary_coercible reports whether values of a type serve as values of another without conversion, as Postgres'
/// IsBinaryCoercible decides from pg_cast's binary casts and the polymorphic types.
pub fn binary_coercible(from: u32, to: u32) -> bool {
    let from = crate::usertypes::base_type(crate::catalog::ColumnType { oid: from, modifier: -1 }).oid;
    if from == to {
        return true;
    }
    match to {
        crate::functions::ANYARRAY => return crate::array::is_array_type(from),
        crate::oid::ANYENUM => {
            return crate::usertypes::get(from).is_some_and(|t| matches!(t.kind, crate::usertypes::Kind::Enum(_)));
        }
        crate::oid::RECORD => {
            return crate::usertypes::get(from).is_some_and(|t| matches!(t.kind, crate::usertypes::Kind::Composite(_)));
        }
        _ => {}
    }
    static CASTS: OnceLock<Vec<(u32, u32)>> = OnceLock::new();
    CASTS
        .get_or_init(|| {
            let Some(table) = lookup("pg_catalog", "pg_cast") else { return Vec::new() };
            let columns = ["castsource", "casttarget", "castmethod"].map(|c| table.column(c));
            let [Some(source), Some(target), Some(method)] = columns else { return Vec::new() };
            let number = |v: &Value| v.output().and_then(|t| t.parse().ok()).unwrap_or(0);
            builtin::rows(table)
                .iter()
                .filter(|r| r[method].output().as_deref() == Some("b"))
                .map(|r| (number(&r[source]), number(&r[target])))
                .collect()
        })
        .contains(&(from, to))
}

/// Rows collects the rows of a system catalog relation, with each column NULL unless set.
pub struct Rows<'t> {
    table: &'t CatalogTable,
    rows: Vec<Vec<Value>>,
}

impl<'t> Rows<'t> {
    /// new starts the rows of a relation.
    fn new(table: &'t CatalogTable) -> Rows<'t> {
        Rows { table, rows: Vec::new() }
    }

    /// zeros returns a zero of each named column's type, for the counters of the statistics views.
    fn zeros<'n>(&self, names: &[&'n str]) -> Vec<(&'n str, Value)> {
        let zero = |name: &str| match self.table.column(name).map(|i| self.table.columns[i].type_oid) {
            Some(crate::oid::INT4) => Value::Int4(0),
            Some(crate::oid::FLOAT8) => Value::Float8(0.0),
            Some(crate::oid::NUMERIC) => Value::Numeric(crate::numeric::Numeric::from_i64(0)),
            _ => Value::Int8(0),
        };
        names.iter().map(|&name| (name, zero(name))).collect()
    }

    /// push adds a row with the given columns set, ignoring columns the relation lacks.
    fn push(&mut self, fields: Vec<(&str, Value)>) {
        let mut row = vec![Value::Null; self.table.columns.len()];
        for (name, value) in fields {
            if let Some(i) = self.table.column(name) {
                let type_oid = self.table.columns[i].type_oid;
                row[i] = match value {
                    Value::Text(text) if crate::array::is_vector_type(type_oid) => {
                        crate::cast::input(&text, type_oid).unwrap_or(Value::Text(text))
                    }
                    value => value,
                };
            }
        }
        self.rows.push(row);
    }
}

/// oid returns an OID value.
fn oid(value: u32) -> Value {
    Value::Oid(value)
}

/// text returns a value of a text type.
fn text(value: impl Into<String>) -> Value {
    Value::Text(value.into())
}

/// int2 returns a smallint value.
fn int2(value: i16) -> Value {
    Value::Int2(value)
}

/// int4 returns an integer value.
fn int4(value: i32) -> Value {
    Value::Int4(value)
}

/// boolean returns a boolean value.
fn boolean(value: bool) -> Value {
    Value::Bool(value)
}

impl Ctx<'_> {
    /// catalog_rows returns the rows of a system catalog relation.
    pub fn catalog_rows(&mut self, table: &'static CatalogTable) -> Result<Vec<Vec<Value>>> {
        let mut rows = Rows::new(table);
        rows.rows = builtin::rows(table);
        if table.schema == "information_schema" {
            let database = Value::Text(self.session.database.clone());
            let catalogs: Vec<usize> = table
                .columns
                .iter()
                .enumerate()
                .filter(|(_, c)| c.name.ends_with("_catalog"))
                .map(|(i, _)| i)
                .collect();
            for row in &mut rows.rows {
                for &i in &catalogs {
                    if row[i] == Value::Text("postgres".into()) {
                        row[i] = database.clone();
                    }
                }
            }
        }
        match table.schema {
            "pg_catalog" => self.pg_catalog_rows(&mut rows)?,
            _ => self.information_schema_rows(&mut rows)?,
        }
        Ok(rows.rows)
    }

    /// catalog_relation returns the system catalog relation that a possibly qualified name refers to: a relation of
    /// pg_catalog or information_schema when named in it, or of pg_catalog for an unqualified name that no schema
    /// before pg_catalog in the search path has.
    pub fn catalog_relation(&mut self, schema: &str, name: &str) -> Result<Option<&'static CatalogTable>> {
        if !schema.is_empty() {
            return Ok(lookup(schema, name));
        }
        let Some(table) = lookup("pg_catalog", name) else { return Ok(None) };
        let path = self.session.search_path();
        if !path.iter().any(|s| s == "pg_catalog") {
            return Ok(Some(table));
        }
        for schema in path {
            if schema == "pg_catalog" {
                break;
            }
            if self.txn.table(self.db, &schema, name)?.is_some() || self.views(&schema)?.iter().any(|(v, _)| v == name)
            {
                return Ok(None);
            }
        }
        Ok(Some(table))
    }
}
