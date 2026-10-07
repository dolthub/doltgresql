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

//! Statements that define schema objects.

use doltdb::table::{Table, empty_rows};
use doltdb::tags::{EXTENDED_KIND, STRING_KIND, auto_generate_tag};
use pg_query::protobuf::{
    ConstrType, CreateSchemaStmt, CreateStmt, CreateTableAsStmt, DropBehavior, DropStmt, IndexStmt, ObjectType,
    OnCommitAction, RangeVar, ResTarget, SelectStmt, SortByDir, SortByNulls, TruncateStmt,
};
use pg_query::{Node, NodeEnum};
use store::Hash;

use crate::Outcome;
use crate::auth::Object;
use crate::cast::type_display;
use crate::catalog::ColumnType;
use crate::catalog::table::{Check, ColumnDef, HIDDEN_BASE, IndexDef, Primary, TableDef, schema_message};
use crate::error::{PgError, Result, code};
use crate::expr::{
    Binder, Scope, ScopeColumn, arg_location, assign, assignable, node_name, position, resolve_type_name,
};
use crate::plan::Planner;
use crate::query::Ctx;

/// type_conflict fails when a column that a new table merges with an inherited one has another type, naming it as
/// Postgres' MergeAttributes does.
fn type_conflict(what: &str, name: &str, inherited: ColumnType, other: ColumnType) -> Result<()> {
    if inherited == other {
        return Ok(());
    }
    Err(PgError {
        detail: Some(format!("{} versus {}", type_display(inherited.oid), type_display(other.oid))),
        ..PgError::new(code::DATATYPE_MISMATCH, format!("{what} \"{name}\" has a type conflict"))
    })
}

/// constraint_type returns a constraint node's type.
pub(crate) fn constraint_type(constraint: &pg_query::protobuf::Constraint) -> ConstrType {
    ConstrType::try_from(constraint.contype).unwrap_or(ConstrType::Undefined)
}

/// expression_text returns an expression as SQL text, which defaults and check constraints store, by deparsing it
/// inside a SELECT.
pub fn expression_text(expr: &Node) -> Result<String> {
    let select = SelectStmt {
        target_list: vec![Node {
            node: Some(NodeEnum::ResTarget(Box::new(ResTarget {
                val: Some(Box::new(expr.clone())),
                ..Default::default()
            }))),
        }],
        limit_option: pg_query::protobuf::LimitOption::Default as i32,
        op: pg_query::protobuf::SetOperation::SetopNone as i32,
        ..Default::default()
    };
    let text = pg_query::NodeRef::SelectStmt(&select).deparse().map_err(PgError::internal)?;
    Ok(text.strip_prefix("SELECT ").unwrap_or(&text).to_string())
}

/// check_constraint_attributes fails as Postgres' grammar does for DEFERRABLE, NOT DEFERRABLE, INITIALLY DEFERRED, and
/// INITIALLY IMMEDIATE clauses that do not follow a constraint that takes them, or that contradict each other.
fn check_constraint_attributes(constraints: &[Node]) -> Result<()> {
    let mut takes_attributes = false;
    let (mut deferrable, mut initially): (Option<bool>, Option<bool>) = (None, None);
    for constraint in constraints {
        let Some(NodeEnum::Constraint(constraint)) = constraint.node.as_ref() else { continue };
        let kind = constraint_type(constraint);
        let error = |message: &str| PgError {
            position: position(constraint.location),
            ..PgError::new(code::SYNTAX_ERROR, message)
        };
        let clause = match kind {
            ConstrType::ConstrAttrDeferrable => "DEFERRABLE",
            ConstrType::ConstrAttrNotDeferrable => "NOT DEFERRABLE",
            ConstrType::ConstrAttrDeferred => "INITIALLY DEFERRED",
            ConstrType::ConstrAttrImmediate => "INITIALLY IMMEDIATE",
            other => {
                takes_attributes = matches!(
                    other,
                    ConstrType::ConstrUnique
                        | ConstrType::ConstrPrimary
                        | ConstrType::ConstrForeign
                        | ConstrType::ConstrExclusion
                );
                (deferrable, initially) = (None, None);
                continue;
            }
        };
        if !takes_attributes {
            return Err(error(&format!("misplaced {clause} clause")));
        }
        match kind {
            ConstrType::ConstrAttrDeferrable | ConstrType::ConstrAttrNotDeferrable => {
                if deferrable.is_some() {
                    return Err(error("multiple DEFERRABLE/NOT DEFERRABLE clauses not allowed"));
                }
                deferrable = Some(kind == ConstrType::ConstrAttrDeferrable);
            }
            _ => {
                if initially.is_some() {
                    return Err(error("multiple INITIALLY IMMEDIATE/DEFERRED clauses not allowed"));
                }
                initially = Some(kind == ConstrType::ConstrAttrDeferred);
            }
        }
        if deferrable == Some(false) && initially == Some(true) {
            return Err(error("constraint declared INITIALLY DEFERRED must be DEFERRABLE"));
        }
    }
    Ok(())
}

/// Deferral is whether a constraint is DEFERRABLE and whether it is INITIALLY DEFERRED.
pub(crate) type Deferral = (bool, bool);

/// deferral returns a table constraint's DEFERRABLE and INITIALLY DEFERRED settings, where INITIALLY DEFERRED implies
/// DEFERRABLE.
pub(crate) fn deferral(constraint: &pg_query::protobuf::Constraint) -> Deferral {
    (constraint.deferrable || constraint.initdeferred, constraint.initdeferred)
}

/// column_deferrals returns each column constraint's DEFERRABLE and INITIALLY DEFERRED settings, applying the
/// attribute clauses that follow a constraint to it, as Postgres' transformConstraintAttrs does.
fn column_deferrals(constraints: &[Node]) -> Vec<Deferral> {
    let mut out: Vec<Deferral> = Vec::with_capacity(constraints.len());
    let mut last = None;
    for node in constraints {
        let Some(NodeEnum::Constraint(constraint)) = node.node.as_ref() else {
            out.push((false, false));
            continue;
        };
        let target: Option<&mut Deferral> = last.and_then(|i: usize| out.get_mut(i));
        match (constraint_type(constraint), target) {
            (ConstrType::ConstrAttrDeferrable, Some(entry)) => entry.0 = true,
            (ConstrType::ConstrAttrNotDeferrable, Some(entry)) => entry.0 = false,
            (ConstrType::ConstrAttrDeferred, Some(entry)) => *entry = (true, true),
            (ConstrType::ConstrAttrImmediate, Some(entry)) => entry.1 = false,
            (
                ConstrType::ConstrAttrDeferrable
                | ConstrType::ConstrAttrNotDeferrable
                | ConstrType::ConstrAttrDeferred
                | ConstrType::ConstrAttrImmediate,
                None,
            ) => {}
            _ => last = Some(out.len()),
        }
        out.push(deferral(constraint));
    }
    out
}

/// both_default_and_generated returns Postgres' error for a column with both a default and a generation expression.
fn both_default_and_generated(column: &str, table: &str, location: i32) -> PgError {
    PgError {
        position: position(location),
        ..PgError::new(
            code::SYNTAX_ERROR,
            format!("both default and generation expression specified for column \"{column}\" of table \"{table}\""),
        )
    }
}

/// TableParts are what a table's definition elements add up to, before the checks and indexes are named.
#[derive(Default)]
pub(crate) struct TableParts {
    pub columns: Vec<ColumnDef>,
    pub primary_key: Vec<usize>,
    /// The primary key constraint's name and deferral.
    pub primary: Primary,
    /// Each check constraint's name, empty when unnamed, and expression.
    pub checks: Vec<(String, Node)>,
    /// Each unique constraint's name, empty when unnamed, columns, and deferral.
    pub uniques: Vec<(String, Vec<usize>, Deferral)>,
    /// Each serial or identity column with its sequence's data type and options.
    pub generated: Vec<(usize, &'static str, Vec<Node>)>,
    /// Each foreign key with its referencing columns.
    pub foreign: Vec<(Vec<usize>, pg_query::protobuf::Constraint)>,
    /// Each column default's column and expression.
    pub defaults: Vec<(usize, Node)>,
    /// Each generated column's column and expression.
    pub generation: Vec<(usize, Node)>,
}

impl TableParts {
    /// add_column adds a column definition with its column constraints.
    pub fn add_column(&mut self, table: &str, def: &pg_query::protobuf::ColumnDef) -> Result<()> {
        if self.columns.iter().any(|c| c.name == def.colname) {
            return Err(PgError::new(
                code::DUPLICATE_COLUMN,
                format!("column \"{}\" specified more than once", def.colname),
            ));
        }
        let index = self.columns.len();
        let type_name = def.type_name.as_ref().ok_or_else(|| PgError::internal("a column without a type"))?;
        let serial = serial_type(type_name);
        let ty = match serial {
            Some(data_type) => crate::catalog::resolve_type(&[data_type.to_string()], &[], false, None)?,
            None => resolve_type_name(type_name)?,
        };
        if ty.oid == crate::oid::RECORD {
            return Err(PgError::new(
                code::INVALID_TABLE_DEFINITION,
                format!("column \"{}\" has pseudo-type record", def.colname),
            ));
        }
        if let Some(data_type) = serial {
            self.generated.push((index, data_type, Vec::new()));
        }
        let mut column = ColumnDef {
            name: def.colname.clone(),
            ty,
            tag: 0,
            encoding: ty.encoding(),
            nullable: true,
            primary_key: false,
            default: String::new(),
            generated: false,
            mysql_type: String::new(),
            comment: String::new(),
        };
        check_constraint_attributes(&def.constraints)?;
        let deferrals = column_deferrals(&def.constraints);
        for (constraint, &deferral) in def.constraints.iter().zip(&deferrals) {
            let Some(NodeEnum::Constraint(constraint)) = constraint.node.as_ref() else { continue };
            match constraint_type(constraint) {
                ConstrType::ConstrAttrDeferrable
                | ConstrType::ConstrAttrNotDeferrable
                | ConstrType::ConstrAttrImmediate
                | ConstrType::ConstrAttrDeferred => {}
                ConstrType::ConstrNotnull => column.nullable = false,
                ConstrType::ConstrNull => column.nullable = true,
                ConstrType::ConstrPrimary => {
                    if !self.primary_key.is_empty() {
                        return Err(multiple_primary_keys(table, constraint.location));
                    }
                    self.primary_key.push(index);
                    self.primary = primary(table, &constraint.conname, deferral);
                }
                ConstrType::ConstrGenerated => {
                    if !column.default.is_empty() {
                        return Err(both_default_and_generated(&column.name, table, constraint.location));
                    }
                    let expr =
                        constraint.raw_expr.as_deref().ok_or_else(|| PgError::internal("an empty generation"))?;
                    column.default = expression_text(expr)?;
                    column.generated = true;
                    self.generation.push((index, expr.clone()));
                }
                ConstrType::ConstrDefault => {
                    if column.generated {
                        return Err(both_default_and_generated(&column.name, table, constraint.location));
                    }
                    let expr = constraint.raw_expr.as_deref().ok_or_else(|| PgError::internal("an empty DEFAULT"))?;
                    column.default = expression_text(expr)?;
                    self.defaults.push((index, expr.clone()));
                }
                ConstrType::ConstrCheck => {
                    let expr = constraint.raw_expr.as_deref().ok_or_else(|| PgError::internal("an empty CHECK"))?;
                    self.checks.push((constraint.conname.clone(), expr.clone()));
                }
                ConstrType::ConstrUnique => self.uniques.push((constraint.conname.clone(), vec![index], deferral)),
                ConstrType::ConstrForeign => {
                    let mut constraint = (**constraint).clone();
                    (constraint.deferrable, constraint.initdeferred) = deferral;
                    self.foreign.push((vec![index], constraint));
                }
                ConstrType::ConstrIdentity => {
                    let data_type = match ty.oid {
                        crate::oid::INT2 => "int2",
                        crate::oid::INT4 => "int4",
                        crate::oid::INT8 => "int8",
                        _ => {
                            return Err(PgError::new(
                                code::INVALID_PARAMETER_VALUE,
                                "identity column type must be smallint, integer, or bigint",
                            ));
                        }
                    };
                    if !column.default.is_empty() {
                        return Err(PgError {
                            position: position(constraint.location),
                            ..PgError::new(
                                code::SYNTAX_ERROR,
                                format!(
                                    "both default and identity specified for column \"{}\" of table \"{table}\"",
                                    def.colname
                                ),
                            )
                        });
                    }
                    self.generated.push((index, data_type, constraint.options.clone()));
                }
                other => return Err(PgError::unsupported(format!("the column constraint {other:?}"))),
            }
        }
        self.columns.push(column);
        Ok(())
    }

    /// key_columns returns the columns a constraint's key names.
    fn key_columns(&self, constraint: &pg_query::protobuf::Constraint) -> Result<Vec<usize>> {
        let mut keys = Vec::new();
        for key in &constraint.keys {
            let key = node_name(key).unwrap_or_default();
            keys.push(self.columns.iter().position(|c| c.name == key).ok_or_else(|| PgError {
                position: position(constraint.location),
                ..PgError::new(code::UNDEFINED_COLUMN, format!("column \"{key}\" named in key does not exist"))
            })?);
        }
        Ok(keys)
    }

    /// add_constraint adds a table constraint.
    pub fn add_constraint(&mut self, table: &str, constraint: &pg_query::protobuf::Constraint) -> Result<()> {
        match constraint_type(constraint) {
            ConstrType::ConstrPrimary => {
                if !self.primary_key.is_empty() {
                    return Err(multiple_primary_keys(table, constraint.location));
                }
                self.primary_key = self.key_columns(constraint)?;
                self.primary = primary(table, &constraint.conname, deferral(constraint));
            }
            ConstrType::ConstrCheck => {
                let expr = constraint.raw_expr.as_deref().ok_or_else(|| PgError::internal("an empty CHECK"))?;
                self.checks.push((constraint.conname.clone(), expr.clone()));
            }
            ConstrType::ConstrUnique => {
                let keys = self.key_columns(constraint)?;
                self.uniques.push((constraint.conname.clone(), keys, deferral(constraint)));
            }
            ConstrType::ConstrForeign => {
                let mut keys = Vec::new();
                for key in constraint.fk_attrs.iter().filter_map(node_name) {
                    keys.push(self.columns.iter().position(|c| c.name == key).ok_or_else(|| PgError {
                        position: position(constraint.location),
                        ..PgError::new(
                            code::UNDEFINED_COLUMN,
                            format!("column \"{key}\" referenced in foreign key constraint does not exist"),
                        )
                    })?);
                }
                self.foreign.push((keys, constraint.clone()));
            }
            other => return Err(PgError::unsupported(format!("the table constraint {other:?}"))),
        }
        Ok(())
    }
}

/// primary returns a primary key constraint of a table with its name, which is stored only when it is not the
/// default, and its deferral.
fn primary(table: &str, name: &str, (deferrable, initially_deferred): Deferral) -> Primary {
    let name = if name == format!("{table}_pkey") { String::new() } else { name.to_string() };
    Primary { name, deferrable, initially_deferred }
}

/// serial_type returns the integer type of a serial pseudo-type name, or None for any other type.
fn serial_type(type_name: &pg_query::protobuf::TypeName) -> Option<&'static str> {
    if !type_name.array_bounds.is_empty() {
        return None;
    }
    let names: Vec<&str> = type_name.names.iter().filter_map(node_name).collect();
    match names.as_slice() {
        ["smallserial" | "serial2"] | ["pg_catalog", "smallserial" | "serial2"] => Some("int2"),
        ["serial" | "serial4"] | ["pg_catalog", "serial" | "serial4"] => Some("int4"),
        ["bigserial" | "serial8"] | ["pg_catalog", "bigserial" | "serial8"] => Some("int8"),
        _ => None,
    }
}

/// check_column returns the column a check expression references when it references exactly one, which Postgres
/// names an unnamed check constraint after.
pub(crate) fn check_column(expr: &Node) -> Option<String> {
    let node = expr.node.as_ref()?;
    let mut names: Vec<String> = Vec::new();
    for (node, ..) in node.nodes() {
        if let pg_query::NodeRef::ColumnRef(column) = node
            && let Some(name) = column.fields.iter().filter_map(node_name).next_back()
            && !names.iter().any(|n| n == name)
        {
            names.push(name.to_string());
        }
    }
    if names.len() == 1 { names.pop() } else { None }
}

/// schema_element returns an element of CREATE SCHEMA with its relation in the schema being created, failing as
/// Postgres' transformCreateSchemaStmt does when the element names another schema.
fn schema_element(mut element: NodeEnum, schema: &str) -> Result<NodeEnum> {
    let relation = match &mut element {
        NodeEnum::CreateStmt(s) => s.relation.as_mut(),
        NodeEnum::ViewStmt(s) => s.view.as_mut(),
        NodeEnum::IndexStmt(s) => s.relation.as_mut(),
        NodeEnum::CreateSeqStmt(s) => s.sequence.as_mut(),
        NodeEnum::CreateTrigStmt(s) => s.relation.as_mut(),
        _ => None,
    };
    if let Some(relation) = relation {
        if relation.schemaname.is_empty() {
            relation.schemaname = schema.to_string();
        } else if relation.schemaname != schema {
            return Err(PgError::new(
                code::INVALID_SCHEMA_DEFINITION,
                format!(
                    "CREATE specifies a schema ({}) different from the one being created ({schema})",
                    relation.schemaname
                ),
            ));
        }
    }
    Ok(element)
}

/// object_names returns the schema and name a qualified object name list names, with the schema empty when it is
/// unqualified.
fn object_names(names: &[Node]) -> (String, String) {
    let parts: Vec<&str> = names.iter().filter_map(node_name).collect();
    match parts.as_slice() {
        [name] => (String::new(), name.to_string()),
        [schema, name] => (schema.to_string(), name.to_string()),
        [_, schema, name] => (schema.to_string(), name.to_string()),
        _ => (String::new(), String::new()),
    }
}

impl Ctx<'_> {
    /// comment runs COMMENT ON, which keeps the comments of tables and columns in their Dolt schemas, and accepts the
    /// comments of other objects without keeping them, warning as the Go server does.
    pub(crate) fn comment(&mut self, stmt: &pg_query::protobuf::CommentStmt) -> Result<Outcome> {
        let items = match stmt.object.as_deref().and_then(|n| n.node.as_ref()) {
            Some(NodeEnum::List(list)) => list.items.clone(),
            _ => Vec::new(),
        };
        let (relation, column) = match ObjectType::try_from(stmt.objtype) {
            Ok(ObjectType::ObjectTable) => (&items[..], None),
            Ok(ObjectType::ObjectColumn) => match items.split_last() {
                Some((column, relation)) if !relation.is_empty() => (relation, node_name(column)),
                _ => return Err(PgError::new(code::SYNTAX_ERROR, "column name must be qualified")),
            },
            _ => {
                let warning = "COMMENT ON is not yet supported for this kind of object";
                self.session.notice(PgError { severity: "WARNING", ..PgError::new("01000", warning) });
                return Ok(Outcome::command("COMMENT"));
            }
        };
        let (schemaname, relname) = object_names(relation);
        let relation = pg_query::protobuf::RangeVar { schemaname, relname, ..Default::default() };
        let mut table = self.resolve_table(&relation).map_err(|err| PgError { position: None, ..err })?;
        self.require_owner(&Object::Table(table.schema.clone(), table.name.clone()))?;
        match column {
            Some(name) => {
                let Some(column) = table.columns.iter_mut().find(|c| c.name == name) else {
                    return Err(PgError::new(
                        code::UNDEFINED_COLUMN,
                        format!("column \"{name}\" of relation \"{}\" does not exist", table.name),
                    ));
                };
                column.comment = stmt.comment.clone();
            }
            None => table.comment = stmt.comment.clone(),
        }
        let mut stored = table.table.clone();
        stored.schema = self.db.write_value(table.schema_message()?)?;
        let address = stored.write(self.db)?;
        self.txn.root.put_table(self.db, &table.schema, &table.name, Some(address))?;
        Ok(Outcome::command("COMMENT"))
    }

    /// check_generation fails as Postgres does for an expression that a generated column can't be generated from:
    /// one that uses another generated column, a subquery, an aggregate, or a function that is not immutable.
    pub(crate) fn check_generation(
        &mut self,
        expr: &Node,
        table: &str,
        columns: &[ColumnDef],
        column: usize,
    ) -> Result<()> {
        let scope = Scope {
            columns: columns
                .iter()
                .map(|c| ScopeColumn {
                    table: table.to_string(),
                    name: c.name.clone(),
                    ty: c.ty,
                    hidden: false,
                    origin: (0, 0),
                })
                .collect(),
        };
        let mut binder = Binder::new(self, scope);
        binder.clause = "column generation expressions";
        binder.definition = true;
        let bound = binder.bind(expr)?;
        for &(i, location) in &binder.columns {
            if columns[i].generated {
                return Err(PgError {
                    position: position(location),
                    detail: Some("A generated column cannot reference another generated column.".into()),
                    ..PgError::new(
                        code::INVALID_OBJECT_DEFINITION,
                        format!("cannot use generated column \"{}\" in column generation expression", columns[i].name),
                    )
                });
            }
        }
        if !crate::pgcatalog::is_immutable(expr) {
            return Err(PgError::new(code::INVALID_OBJECT_DEFINITION, "generation expression is not immutable"));
        }
        let target = &columns[column];
        assign(bound, ColumnType { modifier: -1, ..target.ty }, &target.name, arg_location(expr)).map(|_| ())
    }

    /// check_default fails as Postgres does for an expression that a column's default can't be.
    pub(crate) fn check_default(&mut self, expr: &Node, column: &ColumnDef) -> Result<()> {
        let mut binder = Binder::new(self, Scope::default());
        binder.clause = "DEFAULT expressions";
        binder.definition = true;
        let bound = binder.bind(expr)?;
        if !assignable(bound.1.oid, column.ty.oid) {
            return Err(PgError {
                hint: Some("You will need to rewrite or cast the expression.".into()),
                ..PgError::new(
                    code::DATATYPE_MISMATCH,
                    format!(
                        "column \"{}\" is of type {} but default expression is of type {}",
                        column.name,
                        type_display(column.ty.oid),
                        type_display(bound.1.oid)
                    ),
                )
            });
        }
        let ty = ColumnType { modifier: -1, ..column.ty };
        assign(bound, ty, &column.name, arg_location(expr)).map(|_| ())
    }

    /// create_table runs CREATE TABLE.
    pub fn create_table(&mut self, create: &CreateStmt) -> Result<Outcome> {
        let relation = create.relation.as_ref().ok_or_else(|| PgError::internal("CREATE TABLE without a name"))?;
        let schema = self
            .relation_schema(relation)
            .map_err(|err| PgError { position: err.position.or(position(relation.location)), ..err })?;
        self.require(&Object::Schema(schema.clone()), "U", -1)?;
        if self.nonlocal_table(relation)?.is_some() {
            let message = format!("relation \"{}\" already exists", relation.relname);
            if create.if_not_exists {
                self.session.notice(PgError::notice(code::DUPLICATE_TABLE, format!("{message}, skipping")));
                return Ok(Outcome::command("CREATE TABLE"));
            }
            return Err(PgError::new(code::DUPLICATE_TABLE, message));
        }
        self.check_nonlocal_name(&relation.relname)?;
        let mut liked = Vec::new();
        for element in &create.table_elts {
            match element.node.as_ref() {
                Some(NodeEnum::TableLikeClause(clause)) => liked.extend(self.like(clause)?),
                _ => liked.push(element.clone()),
            }
        }
        let create = &CreateStmt { table_elts: liked, ..create.clone() };
        let outcome = if create.inh_relations.is_empty() {
            self.create_table_in(create, schema)?
        } else {
            let table_elts = self.inherit(create)?;
            self.create_table_in(&CreateStmt { table_elts, inh_relations: Vec::new(), ..create.clone() }, schema)?
        };
        self.on_commit(relation, create.oncommit)?;
        Ok(outcome)
    }

    /// like returns the table elements that `LIKE table` copies, as Postgres' transformTableLikeClause does: the
    /// columns with their NOT NULL constraints, their defaults and generation expressions when it includes them, the
    /// check constraints with INCLUDING CONSTRAINTS, and the primary key and unique constraints with INCLUDING INDEXES.
    fn like(&mut self, clause: &pg_query::protobuf::TableLikeClause) -> Result<Vec<Node>> {
        //TODO: copy the table's other indexes with INCLUDING INDEXES, and its comments with INCLUDING COMMENTS
        let relation = clause.relation.as_ref().ok_or_else(|| PgError::internal("LIKE without a table"))?;
        let table = self.resolve_table(relation)?;
        let including = |bit: u32| clause.options & (1 << bit) != 0;
        let mut definitions: Vec<String> = table
            .columns
            .iter()
            .map(|c| {
                let kept = if c.generated { including(4) } else { including(3) };
                let column = ColumnDef {
                    default: if kept { c.default.clone() } else { String::new() },
                    generated: c.generated && kept,
                    ..c.clone()
                };
                crate::dolt::patch::column_definition(&column)
            })
            .collect();
        if including(2) {
            definitions.extend(
                table.checks.iter().map(|c| {
                    format!("CONSTRAINT {} CHECK ({})", crate::engine::quote_identifier(&c.name), c.expression)
                }),
            );
        }
        if including(6) {
            let names = |columns: &[usize]| {
                let names: Vec<String> =
                    columns.iter().map(|&c| crate::engine::quote_identifier(&table.columns[c].name)).collect();
                names.join(", ")
            };
            if !table.keyless() {
                definitions.push(format!("PRIMARY KEY ({})", names(&table.key_columns)));
            }
            let uniques =
                table.indexes.iter().filter(|i| i.unique && !i.system && i.columns.iter().all(|&c| c < HIDDEN_BASE));
            definitions.extend(uniques.map(|i| format!("UNIQUE ({})", names(&i.columns))));
        }
        let parsed =
            pg_query::parse(&format!("CREATE TABLE liked ({})", definitions.join(", "))).map_err(PgError::internal)?;
        let Some(NodeEnum::CreateStmt(liked)) =
            parsed.protobuf.stmts.into_iter().next().and_then(|s| s.stmt).and_then(|s| s.node)
        else {
            return Err(PgError::internal("the columns of LIKE"));
        };
        Ok(liked.table_elts)
    }

    /// inherit returns a CREATE TABLE's elements with the columns and check constraints of the tables it inherits from
    /// first, merging columns of one name as Postgres' MergeAttributes does, since tables keep no link to their parents
    /// and Go copies the columns as LIKE does.
    fn inherit(&mut self, create: &CreateStmt) -> Result<Vec<Node>> {
        let mut columns: Vec<ColumnDef> = Vec::new();
        let mut checks: Vec<Check> = Vec::new();
        for relation in create.inh_relations.iter().filter_map(|n| match n.node.as_ref() {
            Some(NodeEnum::RangeVar(relation)) => Some(relation),
            _ => None,
        }) {
            let parent = self.resolve_table(relation).map_err(|err| PgError { position: None, ..err })?;
            for column in parent.columns {
                match columns.iter_mut().find(|c| c.name == column.name) {
                    Some(existing) => {
                        self.session.notice(PgError::notice(
                            "00000",
                            format!("merging multiple inherited definitions of column \"{}\"", column.name),
                        ));
                        type_conflict("inherited column", &column.name, existing.ty, column.ty)?;
                        existing.nullable &= column.nullable;
                    }
                    None => columns.push(ColumnDef { primary_key: false, ..column }),
                }
            }
            checks.extend(
                parent.checks.into_iter().filter(|c| !checks.iter().any(|k| k.name == c.name)).collect::<Vec<_>>(),
            );
        }
        let definitions: Vec<String> =
            columns
                .iter()
                .map(crate::dolt::patch::column_definition)
                .chain(checks.iter().map(|c| {
                    format!("CONSTRAINT {} CHECK ({})", crate::engine::quote_identifier(&c.name), c.expression)
                }))
                .collect();
        let parsed = pg_query::parse(&format!("CREATE TABLE inherited ({})", definitions.join(", ")))
            .map_err(PgError::internal)?;
        let Some(NodeEnum::CreateStmt(inherited)) =
            parsed.protobuf.stmts.into_iter().next().and_then(|s| s.stmt).and_then(|s| s.node)
        else {
            return Err(PgError::internal("the inherited columns"));
        };
        let mut elements = inherited.table_elts;
        let mut local = Vec::new();
        for element in &create.table_elts {
            let Some(NodeEnum::ColumnDef(def)) = element.node.as_ref() else {
                local.push(element.clone());
                continue;
            };
            let Some(i) = columns.iter().position(|c| c.name == def.colname) else {
                local.push(element.clone());
                continue;
            };
            self.session.notice(PgError::notice(
                "00000",
                format!("merging column \"{}\" with inherited definition", def.colname),
            ));
            if let Some(type_name) = &def.type_name {
                type_conflict("column", &def.colname, columns[i].ty, resolve_type_name(type_name)?)?;
            }
            let Some(NodeEnum::ColumnDef(inherited)) = elements[i].node.as_ref() else { continue };
            let mut merged = def.clone();
            for constraint in &inherited.constraints {
                let Some(NodeEnum::Constraint(c)) = constraint.node.as_ref() else { continue };
                if !def
                    .constraints
                    .iter()
                    .any(|d| matches!(d.node.as_ref(), Some(NodeEnum::Constraint(d)) if d.contype == c.contype))
                {
                    merged.constraints.push(constraint.clone());
                }
            }
            elements[i] = Node { node: Some(NodeEnum::ColumnDef(merged)) };
        }
        let checks = elements.split_off(columns.len());
        elements.extend(local);
        elements.extend(checks);
        Ok(elements)
    }

    /// create_table_in runs CREATE TABLE in a schema, which may be the `dolt` schema that holds Dolt's own tables.
    pub(crate) fn create_table_in(&mut self, create: &CreateStmt, schema: String) -> Result<Outcome> {
        let relation = create.relation.as_ref().ok_or_else(|| PgError::internal("CREATE TABLE without a name"))?;
        let name = relation.relname.as_str();
        let user_type = crate::usertypes::lookup(Some(&schema), name).filter(|t| !t.is_array());
        let composite = user_type.as_ref().is_some_and(|t| matches!(t.kind, crate::usertypes::Kind::Composite(_)));
        if composite || self.relation_names(&schema)?.iter().any(|n| n == name) {
            let message = format!("relation \"{name}\" already exists");
            if create.if_not_exists {
                self.session.notice(PgError::notice(code::DUPLICATE_TABLE, format!("{message}, skipping")));
                return Ok(Outcome::command("CREATE TABLE"));
            }
            return Err(PgError::new(code::DUPLICATE_TABLE, message));
        }
        if user_type.is_some() {
            return Err(PgError {
                hint: Some(
                    "A relation has an associated type of the same name, so you must use a name that doesn't conflict \
                     with any existing type."
                        .into(),
                ),
                ..PgError::new(code::DUPLICATE_OBJECT, format!("type \"{name}\" already exists"))
            });
        }
        for element in &create.table_elts {
            if let Some(NodeEnum::ColumnDef(def)) = element.node.as_ref()
                && let Some(type_name) = &def.type_name
            {
                self.prepare_type(type_name)?;
            }
        }
        let mut parts = TableParts::default();
        for element in &create.table_elts {
            match element.node.as_ref() {
                Some(NodeEnum::ColumnDef(def)) => parts.add_column(name, def)?,
                Some(NodeEnum::Constraint(constraint)) => parts.add_constraint(name, constraint)?,
                _ => return Err(PgError::unsupported("this table element")),
            }
        }
        let TableParts {
            mut columns,
            primary_key,
            primary,
            checks: pending_checks,
            uniques,
            generated,
            foreign,
            defaults,
            generation,
        } = parts;
        for (column, expr) in &defaults {
            self.check_default(expr, &columns[*column])?;
        }
        for (column, expr) in &generation {
            self.check_generation(expr, name, &columns, *column)?;
        }
        for &i in &primary_key {
            columns[i].primary_key = true;
            columns[i].nullable = false;
        }
        let mut constraints = self.constraint_names(&schema)?;
        let mut checks: Vec<Check> = Vec::new();
        let mut check_nodes = Vec::new();
        for (constraint, expr) in pending_checks {
            let check_name = if constraint.is_empty() {
                let column = check_column(&expr);
                choose_relation_name(name, column.as_deref().unwrap_or(""), "check", &constraints)
            } else if checks.iter().any(|c| c.name == constraint) {
                return Err(PgError::new(
                    code::DUPLICATE_OBJECT,
                    format!("check constraint \"{constraint}\" already exists"),
                ));
            } else {
                constraint
            };
            constraints.push(check_name.clone());
            checks.push(Check { name: check_name, expression: expression_text(&expr)? });
            check_nodes.push(expr);
        }
        let mut taken = self.relation_names(&schema)?;
        taken.push(name.to_string());
        if !primary_key.is_empty() {
            taken.push(if primary.name.is_empty() { format!("{name}_pkey") } else { primary.name.clone() });
        }
        for (column, data_type, options) in generated {
            let column_name = columns[column].name.clone();
            columns[column].default =
                self.create_owned_sequence(&schema, name, &column_name, data_type, &options, &mut taken)?;
            columns[column].nullable = false;
        }
        let mut indexes = Vec::new();
        for (constraint, keys, (deferrable, initially_deferred)) in uniques {
            let index_name = if constraint.is_empty() {
                let names: Vec<&str> = keys.iter().map(|&k| columns[k].name.as_str()).collect();
                choose_relation_name(name, &names.join("_"), "key", &taken)
            } else {
                constraint
            };
            taken.push(index_name.clone());
            indexes.push(IndexDef { deferrable, initially_deferred, ..new_index(index_name, keys, true) });
        }
        self.write_new_table(&schema, name, columns, (primary_key, primary), checks, indexes)?;
        let table = self.txn.table(self.db, &schema, name)?.ok_or_else(|| PgError::internal("the new table"))?;
        for expr in &check_nodes {
            let mut binder = Binder::new(self, crate::dml::table_scope(&table, None));
            (binder.clause, binder.definition) = ("check constraints", true);
            crate::expr::condition(binder.bind(expr)?, "CHECK", arg_location(expr))?;
        }
        for (keys, constraint) in foreign {
            let table =
                self.txn.table(self.db, &schema, name)?.ok_or_else(|| PgError::internal("a new table vanished"))?;
            self.add_foreign_key(&table, &keys, &constraint)?;
        }
        Ok(Outcome::command("CREATE TABLE"))
    }

    /// target_schema returns the schema a new object goes in: the named one, which must exist, or the first existing
    /// schema of the search path.
    pub(crate) fn target_schema(&mut self, named: &str, location: i32) -> Result<String> {
        let schema = if named.is_empty() && self.session.temp_first() {
            return self.temp_schema();
        } else if named.is_empty() {
            self.creation_schema()?
        } else if named == "pg_temp" || named == self.session.temp_schema() {
            return self.temp_schema();
        } else if !self.txn.root.schemas.iter().any(|s| s == named.as_bytes()) {
            return Err(PgError {
                position: position(location),
                ..PgError::new(code::INVALID_SCHEMA_NAME, format!("schema \"{named}\" does not exist"))
            });
        } else {
            named.to_string()
        };
        self.require(&Object::Schema(schema.clone()), "C", -1)?;
        Ok(schema)
    }

    /// relation_schema returns the schema that a new relation goes in, which is the session's temporary schema for a
    /// temporary relation.
    pub(crate) fn relation_schema(&mut self, relation: &RangeVar) -> Result<String> {
        if relation.relpersistence != "t" {
            return self.target_schema(&relation.schemaname, relation.location);
        }
        if !matches!(relation.schemaname.as_str(), "" | "pg_temp") && relation.schemaname != self.session.temp_schema()
        {
            return Err(PgError {
                position: position(relation.location),
                ..PgError::new(
                    code::INVALID_TABLE_DEFINITION,
                    "cannot create temporary relation in non-temporary schema",
                )
            });
        }
        self.temp_schema()
    }

    /// temp_schema returns the session's temporary schema, making it when the session has none in the database.
    pub(crate) fn temp_schema(&mut self) -> Result<String> {
        self.require(&Object::Database(self.session.database.clone()), "T", -1)?;
        let schema = self.session.temp_schema();
        self.session.temp.entry(self.session.database.clone()).or_default();
        self.session.temp_used = true;
        if self.txn.temp_schema.is_none() {
            self.txn.inject_temp(self.db, &schema, &Default::default())?;
        }
        Ok(schema)
    }

    /// on_commit records what commits do to a new table, which only a temporary table may ask for.
    fn on_commit(&mut self, relation: &RangeVar, action: i32) -> Result<()> {
        let drop = match OnCommitAction::try_from(action) {
            Ok(OnCommitAction::OncommitDeleteRows) => false,
            Ok(OnCommitAction::OncommitDrop) => true,
            _ => return Ok(()),
        };
        if relation.relpersistence != "t" {
            return Err(PgError::new(code::INVALID_TABLE_DEFINITION, "ON COMMIT can only be used on temporary tables"));
        }
        if let Some(temp) = self.session.temp.get_mut(&self.session.database) {
            temp.on_commit.retain(|(name, _)| *name != relation.relname);
            temp.on_commit.push((relation.relname.clone(), drop));
        }
        Ok(())
    }

    /// write_new_table chooses the columns' tags and writes a new empty table to the working root.
    pub(crate) fn write_new_table(
        &mut self,
        schema: &str,
        name: &str,
        mut columns: Vec<ColumnDef>,
        (primary_key, primary): (Vec<usize>, Primary),
        checks: Vec<Check>,
        indexes: Vec<IndexDef>,
    ) -> Result<()> {
        let mut tags = self.txn.all_tags(self.db)?;
        let mut kinds = Vec::new();
        for column in &mut columns {
            let kind = if column.mysql_type.is_empty() { EXTENDED_KIND } else { STRING_KIND };
            column.tag = auto_generate_tag(&tags, name, &kinds, &column.name, kind);
            tags.insert(column.tag);
            kinds.push(kind);
        }
        let value_columns: Vec<usize> = (0..columns.len()).filter(|i| !primary_key.contains(i)).collect();
        let message = schema_message(&columns, &[], (&primary_key, &value_columns), &checks, &indexes, &primary, "")?;
        let (mut address, mut table) = Table::create(self.db, message)?;
        if !indexes.is_empty() {
            let empty = Hash::of(&empty_rows());
            for index in &indexes {
                table.put_index(self.db, &index.name, Some(empty))?;
            }
            address = table.write(self.db)?;
        }
        self.txn.root.put_table(self.db, schema, name, Some(address))?;
        self.own(Object::Table(schema.to_string(), name.to_string()))
    }

    /// create_table_as runs CREATE TABLE AS, which makes a keyless table of the query's columns and inserts its rows.
    pub fn create_table_as(&mut self, create: &CreateTableAsStmt) -> Result<Outcome> {
        let into = create.into.as_ref().ok_or_else(|| PgError::internal("CREATE TABLE AS without a target"))?;
        let relation = into.rel.as_ref().ok_or_else(|| PgError::internal("CREATE TABLE AS without a name"))?;
        let schema = self.relation_schema(relation)?;
        let name = relation.relname.clone();
        if self.txn.root.table(self.db, &schema, &name)?.is_some() {
            let message = format!("relation \"{name}\" already exists");
            if create.if_not_exists {
                self.session.notice(PgError::notice(code::DUPLICATE_TABLE, format!("{message}, skipping")));
                return Ok(Outcome::command("CREATE TABLE AS"));
            }
            return Err(PgError::new(code::DUPLICATE_TABLE, message));
        }
        let Some(NodeEnum::SelectStmt(select)) = create.query.as_deref().and_then(|n| n.node.as_ref()) else {
            return Err(PgError::unsupported("CREATE TABLE AS without a SELECT"));
        };
        let query = Planner { ctx: self, outer: Vec::new() }.plan_query(select)?;
        let renames: Vec<&str> = into.col_names.iter().filter_map(node_name).collect();
        let columns: Vec<ColumnDef> = query
            .columns
            .iter()
            .zip(&query.types)
            .enumerate()
            .map(|(i, (c, ty))| {
                let ty = if ty.oid == crate::oid::UNKNOWN {
                    ColumnType { oid: crate::oid::TEXT, modifier: -1 }
                } else {
                    *ty
                };
                ColumnDef {
                    name: renames.get(i).map_or(c.name.clone(), |r| r.to_string()),
                    ty,
                    tag: 0,
                    encoding: ty.encoding(),
                    nullable: true,
                    primary_key: false,
                    default: String::new(),
                    generated: false,
                    mysql_type: String::new(),
                    comment: String::new(),
                }
            })
            .collect();
        self.write_new_table(&schema, &name, columns, (Vec::new(), Primary::default()), Vec::new(), Vec::new())?;
        let rows = if into.skip_data { Vec::new() } else { query.plan.run(self)? };
        let count = rows.len();
        let table =
            self.txn.table(self.db, &schema, &name)?.ok_or_else(|| PgError::internal("a new table vanished"))?;
        crate::dml::insert_rows(self, &table, rows)?;
        self.on_commit(relation, into.on_commit)?;
        Ok(Outcome::command(format!("SELECT {count}")))
    }

    /// create_schema runs CREATE SCHEMA.
    pub fn create_schema(&mut self, create: &CreateSchemaStmt) -> Result<Outcome> {
        self.check_new_owner(create.authrole.as_ref())?;
        let name = match (create.schemaname.as_str(), create.authrole.as_ref()) {
            ("", Some(role)) if role.rolename.is_empty() => self.session.role.clone(),
            ("", Some(role)) => role.rolename.clone(),
            (name, _) => name.to_string(),
        };
        if self.txn.root.schemas.iter().any(|s| s == name.as_bytes()) {
            if create.if_not_exists {
                self.session.notice(PgError::notice(
                    code::DUPLICATE_SCHEMA,
                    format!("schema \"{name}\" already exists, skipping"),
                ));
                return Ok(Outcome::command("CREATE SCHEMA"));
            }
            return Err(PgError::new(code::DUPLICATE_SCHEMA, format!("schema \"{name}\" already exists")));
        }
        self.require(&Object::Database(self.session.database.clone()), "C", -1)?;
        self.txn.root.schemas.push(name.clone().into_bytes());
        self.txn.root.schemas.sort();
        self.own(Object::Schema(name.clone()))?;
        let mut elements = Vec::with_capacity(create.schema_elts.len());
        for element in create.schema_elts.iter().filter_map(|e| e.node.clone()) {
            elements.push(schema_element(element, &name)?);
        }
        elements.sort_by_key(|element| match element {
            NodeEnum::CreateSeqStmt(_) => 0,
            NodeEnum::CreateStmt(_) => 1,
            NodeEnum::ViewStmt(_) => 2,
            NodeEnum::IndexStmt(_) => 3,
            NodeEnum::CreateTrigStmt(_) => 4,
            _ => 5,
        });
        let outer = self.session.view_schema.replace(name);
        let result = elements.iter().try_for_each(|element| self.run(element).map(|_| ()));
        self.session.view_schema = outer;
        result?;
        Ok(Outcome::command("CREATE SCHEMA"))
    }

    /// drop runs DROP TABLE and DROP SCHEMA.
    pub fn drop(&mut self, drop: &DropStmt) -> Result<Outcome> {
        let kind = ObjectType::try_from(drop.remove_type).unwrap_or(ObjectType::Undefined);
        let cascade = DropBehavior::try_from(drop.behavior) == Ok(DropBehavior::DropCascade);
        match kind {
            ObjectType::ObjectTable => self.drop_tables(drop),
            ObjectType::ObjectSchema => self.drop_schemas(drop, cascade),
            ObjectType::ObjectIndex => self.drop_indexes(drop),
            ObjectType::ObjectSequence => self.drop_sequences(drop),
            ObjectType::ObjectView => self.drop_views(drop),
            ObjectType::ObjectFunction => self.drop_routines(drop, Some(false)),
            ObjectType::ObjectProcedure => self.drop_routines(drop, Some(true)),
            ObjectType::ObjectRoutine => self.drop_routines(drop, None),
            ObjectType::ObjectTrigger => self.drop_triggers(drop),
            ObjectType::ObjectCast => self.drop_casts(drop),
            ObjectType::ObjectAggregate => self.drop_aggregates(drop),
            ObjectType::ObjectOperator => self.drop_operators(drop),
            ObjectType::ObjectType => self.drop_types(drop, false),
            ObjectType::ObjectDomain => self.drop_types(drop, true),
            ObjectType::ObjectExtension => self.drop_extensions(drop),
            other => Err(PgError::unsupported(format!("DROP {other:?}"))),
        }
    }

    /// missing_relation reports a relation that DROP did not find, as an error or as a notice that IF EXISTS skips
    /// it, naming the schema instead when the relation's schema does not exist, as Postgres does.
    pub(crate) fn missing_relation(
        &mut self,
        kind: &str,
        schema: &str,
        shown: &str,
        missing_ok: bool,
        error_code: &'static str,
    ) -> Result<()> {
        let schema_missing = !schema.is_empty()
            && !matches!(schema, "pg_catalog" | "information_schema" | "dolt")
            && !self.txn.root.schemas.iter().any(|s| s == schema.as_bytes());
        let (error_code, message) = if schema_missing {
            (code::INVALID_SCHEMA_NAME, format!("schema \"{schema}\" does not exist"))
        } else {
            (error_code, format!("{kind} \"{shown}\" does not exist"))
        };
        if !missing_ok {
            return Err(PgError::new(error_code, message));
        }
        self.session.notice(PgError::notice("00000", format!("{message}, skipping")));
        Ok(())
    }

    /// drop_tables runs DROP TABLE, resolving every table first so that a missing one drops none.
    fn drop_tables(&mut self, drop: &DropStmt) -> Result<Outcome> {
        let mut doomed = Vec::new();
        for object in &drop.objects {
            let Some(NodeEnum::List(list)) = object.node.as_ref() else { continue };
            let (schema, name) = object_names(&list.items);
            if self.system_catalog(&schema, &name).is_some() {
                return Err(PgError::new(
                    code::INSUFFICIENT_PRIVILEGE,
                    format!("permission denied: \"{name}\" is a system catalog"),
                ));
            }
            let schemas =
                if schema.is_empty() { self.session.search_path() } else { vec![self.session.named_schema(&schema)] };
            let mut found = None;
            for s in schemas {
                if self.txn.root.table(self.db, &s, &name)?.is_some() {
                    found = Some(s);
                    break;
                }
            }
            match found {
                Some(s) => {
                    self.session.temp_used |= s == self.session.temp_schema();
                    self.require_owner(&Object::Table(s.clone(), name.clone()))?;
                    doomed.push((s, name))
                }
                None => {
                    let shown = if schema.is_empty() { name.clone() } else { format!("{schema}.{name}") };
                    if self.find_view(&schema, &name)?.is_some() {
                        return Err(PgError {
                            hint: Some("Use DROP VIEW to remove a view.".into()),
                            ..PgError::new(code::WRONG_OBJECT_TYPE, format!("\"{shown}\" is not a table"))
                        });
                    }
                    self.missing_relation("table", &schema, &shown, drop.missing_ok, code::UNDEFINED_TABLE)?;
                }
            }
        }
        let cascade = DropBehavior::try_from(drop.behavior) == Ok(DropBehavior::DropCascade);
        for (schema, name) in &doomed {
            self.drop_row_type_dependents(schema, name, cascade, &doomed)?;
            self.drop_table_foreign_keys(schema, name, drop.behavior, &doomed)?;
        }
        for (schema, name) in doomed {
            self.txn.root.put_table(self.db, &schema, &name, None)?;
            self.drop_table_triggers(&schema, &name)?;
            self.forget_object(&Object::Table(schema.clone(), name.clone()))?;
            self.drop_owned_sequences(&schema, &name)?;
        }
        Ok(Outcome::command("DROP TABLE"))
    }

    /// drop_row_type_dependents fails as Postgres does when another table's column or a routine uses the row type of
    /// a table being dropped, unless the drop cascades, which drops those columns and routines.
    fn drop_row_type_dependents(
        &mut self,
        schema: &str,
        name: &str,
        cascade: bool,
        dropping: &[(String, String)],
    ) -> Result<()> {
        let row_type = crate::pgcatalog::row_type_oid(schema, name);
        let mut columns = Vec::new();
        for ((s, t), address) in crate::dolt::procedures::table_map(self.db, &self.txn.root.clone())? {
            if dropping.iter().any(|(ds, dt)| *ds == s && *dt == t) {
                continue;
            }
            let table = TableDef::load(self.db, &s, &t, address)?;
            columns.extend(
                table.columns.iter().filter(|c| c.ty.oid == row_type).map(|c| (s.clone(), t.clone(), c.name.clone())),
            );
        }
        let routines: Vec<_> = self
            .routines()?
            .iter()
            .filter(|r| r.ret.oid == row_type || r.params.iter().any(|p| p.ty.oid == row_type))
            .cloned()
            .collect();
        if columns.is_empty() && routines.is_empty() {
            return Ok(());
        }
        let dependents: Vec<String> = columns
            .iter()
            .map(|(s, t, c)| format!("column {c} of table {}", self.shown_relation(s, t)))
            .chain(routines.iter().map(|r| format!("function {}", r.signature())))
            .collect();
        let shown = self.shown_relation(schema, name);
        if !cascade {
            return Err(PgError {
                detail: Some(
                    dependents.iter().map(|d| format!("{d} depends on type {shown}")).collect::<Vec<_>>().join("\n"),
                ),
                hint: Some("Use DROP ... CASCADE to drop the dependent objects too.".into()),
                ..PgError::new(
                    code::DEPENDENT_OBJECTS_STILL_EXIST,
                    format!("cannot drop table {shown} because other objects depend on it"),
                )
            });
        }
        self.notice_cascades(dependents.iter().map(|d| format!("drop cascades to {d}")).collect());
        let quote = crate::engine::quote_identifier;
        let mut statements: Vec<String> = columns
            .iter()
            .map(|(s, t, c)| format!("ALTER TABLE {}.{} DROP COLUMN {}", quote(s), quote(t), quote(c)))
            .collect();
        statements.extend(routines.iter().map(|r| format!("DROP ROUTINE {}.{}", quote(&r.schema), r.signature())));
        for statement in statements {
            match crate::parse::parse(&statement)?.into_iter().next() {
                Some(crate::parse::Statement::Postgres { node: NodeEnum::AlterTableStmt(alter), .. }) => {
                    self.alter_table(&alter)?;
                }
                Some(crate::parse::Statement::Postgres { node: NodeEnum::DropStmt(drop), .. }) => {
                    self.drop_routines(&drop, None)?;
                }
                _ => return Err(PgError::internal("a cascading drop statement did not parse")),
            }
        }
        Ok(())
    }

    /// drop_schemas runs DROP SCHEMA, which fails for a schema with tables unless it cascades to them.
    fn drop_schemas(&mut self, drop: &DropStmt, cascade: bool) -> Result<Outcome> {
        let mut doomed = Vec::new();
        for object in &drop.objects {
            let Some(name) = node_name(object) else { continue };
            if !self.txn.root.schemas.iter().any(|s| s == name.as_bytes()) {
                if !drop.missing_ok {
                    return Err(PgError::new(code::INVALID_SCHEMA_NAME, format!("schema \"{name}\" does not exist")));
                }
                self.session.notice(PgError::notice("00000", format!("schema \"{name}\" does not exist, skipping")));
                continue;
            }
            self.require_owner(&Object::Schema(name.to_string()))?;
            let prefix = doltdb::root::table_key(name, "");
            let tables: Vec<String> = self
                .txn
                .root
                .tables(self.db)?
                .into_iter()
                .filter_map(|(key, _)| {
                    key.strip_prefix(prefix.as_slice()).map(|n| String::from_utf8_lossy(n).into_owned())
                })
                .collect();
            let mut types: Vec<(String, bool)> = self
                .user_types()?
                .values()
                .filter(|t| t.schema == name && !t.is_array())
                .map(|t| (t.name.clone(), matches!(t.kind, crate::usertypes::Kind::Domain(_))))
                .collect();
            types.sort();
            let sequences: Vec<String> = crate::sequences::all(self.db, &self.txn.root)?
                .into_iter()
                .filter(|s| s.owner_table.is_empty())
                .map(|s| crate::catalog::id::segments(&s.id))
                .filter(|segments| segments.first().is_some_and(|s| s == name))
                .filter_map(|segments| segments.get(1).cloned())
                .collect();
            let views: Vec<String> = self.views(name)?.into_iter().map(|(view, _)| view).collect();
            let dependents: Vec<String> = tables
                .iter()
                .filter(|t| !t.starts_with("dolt_"))
                .map(|t| format!("table {name}.{t}"))
                .chain(views.iter().map(|v| format!("view {name}.{v}")))
                .chain(types.iter().map(|(t, _)| format!("type {name}.{t}")))
                .chain(sequences.iter().map(|s| format!("sequence {name}.{s}")))
                .collect();
            if !dependents.is_empty() && !cascade {
                let detail = dependents.iter().map(|d| format!("{d} depends on schema {name}")).collect::<Vec<_>>();
                return Err(PgError {
                    detail: Some(detail.join("\n")),
                    hint: Some("Use DROP ... CASCADE to drop the dependent objects too.".into()),
                    ..PgError::new(
                        code::DEPENDENT_OBJECTS_STILL_EXIST,
                        format!("cannot drop schema {name} because other objects depend on it"),
                    )
                });
            }
            doomed.push((name.to_string(), tables, types, sequences, dependents));
        }
        for (name, tables, types, sequences, dependents) in doomed {
            match dependents.len() {
                0 => {}
                1 => self.session.notice(PgError::notice("00000", format!("drop cascades to {}", dependents[0]))),
                n => {
                    let detail = dependents.iter().map(|d| format!("drop cascades to {d}")).collect::<Vec<_>>();
                    self.session.notice(PgError {
                        detail: Some(detail.join("\n")),
                        ..PgError::notice("00000", format!("drop cascades to {n} other objects"))
                    });
                }
            }
            for table in tables {
                self.txn.root.put_table(self.db, &name, &table, None)?;
                self.drop_table_triggers(&name, &table)?;
                self.forget_object(&Object::Table(name.clone(), table))?;
            }
            let quoted = |object: &str| {
                format!("{}.{}", crate::engine::quote_identifier(&name), crate::engine::quote_identifier(object))
            };
            let mut statements: Vec<String> = types
                .iter()
                .map(|(t, domain)| format!("DROP {} {} CASCADE", if *domain { "DOMAIN" } else { "TYPE" }, quoted(t)))
                .collect();
            statements.extend(sequences.iter().map(|s| format!("DROP SEQUENCE {}", quoted(s))));
            for statement in statements {
                let parsed = pg_query::parse(&statement).map_err(PgError::internal)?;
                let Some(NodeEnum::DropStmt(stmt)) =
                    parsed.protobuf.stmts.into_iter().next().and_then(|s| s.stmt).and_then(|s| s.node)
                else {
                    return Err(PgError::internal("a schema's dependent object"));
                };
                self.drop(&stmt)?;
            }
            self.txn.root.schemas.retain(|s| s != name.as_bytes());
            self.forget_object(&Object::Schema(name))?;
        }
        Ok(Outcome::command("DROP SCHEMA"))
    }

    /// truncate runs TRUNCATE, which empties the tables.
    pub fn truncate(&mut self, truncate: &TruncateStmt) -> Result<Outcome> {
        let mut tables = Vec::new();
        for relation in &truncate.relations {
            let Some(NodeEnum::RangeVar(relation)) = relation.node.as_ref() else { continue };
            let table = self.resolve_table(relation).map_err(|err| PgError { position: None, ..err })?;
            self.require(&Object::Table(table.schema.clone(), table.name.clone()), "D", -1)?;
            tables.push(table);
        }
        self.check_truncate(&tables)?;
        for table in tables {
            let mut stored = table.table.clone();
            stored.primary_index = empty_rows();
            for index in &table.indexes {
                let empty = index.empty_root(self.db)?;
                stored.put_index(self.db, &index.name, Some(empty))?;
            }
            let address = stored.write(self.db)?;
            self.txn.root.put_table(self.db, &table.schema, &table.name, Some(address))?;
        }
        Ok(Outcome::command("TRUNCATE TABLE"))
    }
}

impl Ctx<'_> {
    /// relation_names returns the names of the tables, indexes, sequences, and views in a schema, which new relations
    /// must avoid.
    pub(crate) fn relation_names(&mut self, schema: &str) -> Result<Vec<String>> {
        let mut names = Vec::new();
        let prefix = doltdb::root::table_key(schema, "");
        for (key, address) in self.txn.root.tables(self.db)? {
            let Some(name) = key.strip_prefix(prefix.as_slice()) else { continue };
            let name = String::from_utf8_lossy(name).into_owned();
            let table = TableDef::load(self.db, schema, &name, address)?;
            names.extend(table.indexes.iter().filter(|i| !i.system).map(|i| i.name.clone()));
            if !table.key_columns.is_empty() {
                names.push(table.primary_name());
            }
            names.push(name);
        }
        for sequence in crate::sequences::all(self.db, &self.txn.root)? {
            let (sequence_schema, name) = crate::sequences::schema_and_name(&sequence);
            if sequence_schema == schema {
                names.push(name);
            }
        }
        names.extend(self.views(schema)?.into_iter().map(|(name, _)| name));
        Ok(names)
    }

    /// constraint_names returns the names of the constraints of every table in a schema, which new constraints must
    /// avoid.
    pub(crate) fn constraint_names(&mut self, schema: &str) -> Result<Vec<String>> {
        let mut names = Vec::new();
        let prefix = doltdb::root::table_key(schema, "");
        for (key, address) in self.txn.root.tables(self.db)? {
            let Some(name) = key.strip_prefix(prefix.as_slice()) else { continue };
            let name = String::from_utf8_lossy(name).into_owned();
            let table = TableDef::load(self.db, schema, &name, address)?;
            names.extend(table.checks.iter().map(|c| c.name.clone()));
            names.extend(table.indexes.iter().filter(|i| i.unique).map(|i| i.name.clone()));
            if !table.key_columns.is_empty() {
                names.push(table.primary_name());
            }
        }
        names.extend(self.foreign_keys()?.into_iter().filter(|fk| fk.child_schema == schema).map(|fk| fk.name));
        Ok(names)
    }

    /// create_index runs CREATE INDEX, building the index from the table's rows.
    pub fn create_index(&mut self, stmt: &IndexStmt) -> Result<Outcome> {
        let relation = stmt.relation.as_ref().ok_or_else(|| PgError::internal("CREATE INDEX without a table"))?;
        let table = self.resolve_table(relation)?;
        self.require_owner(&Object::Table(table.schema.clone(), table.name.clone()))?;
        let method = stmt.access_method.as_str();
        if !matches!(method, "" | "btree" | "hash") {
            if let Some((extension, access_method)) = self.access_method(method)? {
                return self.create_vector_index(stmt, table, extension, access_method);
            }
            if !matches!(method, "gist" | "gin" | "spgist" | "brin") {
                return Err(PgError::new(code::UNDEFINED_OBJECT, format!("access method \"{method}\" does not exist")));
            }
            for param in &stmt.index_params {
                let Some(NodeEnum::IndexElem(elem)) = param.node.as_ref() else { continue };
                let column = table.columns.iter().find(|c| c.name == elem.name);
                if let Some(column) = column.filter(|c| crate::types::base_type(c.ty.oid).is_some())
                    && elem.opclass.is_empty()
                {
                    return Err(crate::extensions::no_default_class(column.ty.oid, method));
                }
            }
            return Err(PgError::unsupported(format!("indexes using {method}")));
        }
        if !matches!(stmt.table_space.as_str(), "" | "pg_default") {
            let message = format!("tablespace \"{}\" does not exist", stmt.table_space);
            return Err(PgError::new(code::UNDEFINED_OBJECT, message));
        }
        let mut columns = Vec::new();
        let mut names = Vec::new();
        let mut expressions = Vec::new();
        let mut descending = Vec::new();
        let mut nulls_last = Vec::new();
        let mut op_classes = Vec::new();
        for param in &stmt.index_params {
            let Some(NodeEnum::IndexElem(elem)) = param.node.as_ref() else { continue };
            let (column, ty) = match &elem.expr {
                Some(expr) => {
                    let ty = self.index_expression(&table, expr)?;
                    names.push(crate::expr::figure_index_name(expr));
                    expressions.push((columns.len(), expression_text(expr)?, ty));
                    (HIDDEN_BASE + table.hidden.len() + expressions.len() - 1, ty)
                }
                None => {
                    let column = table.columns.iter().position(|c| c.name == elem.name).ok_or_else(|| {
                        PgError::new(code::UNDEFINED_COLUMN, format!("column \"{}\" does not exist", elem.name))
                    })?;
                    names.push(elem.name.clone());
                    (column, table.columns[column].ty)
                }
            };
            let desc = SortByDir::try_from(elem.ordering) == Ok(SortByDir::SortbyDesc);
            let last = match SortByNulls::try_from(elem.nulls_ordering) {
                Ok(SortByNulls::SortbyNullsFirst) => false,
                Ok(SortByNulls::SortbyNullsLast) => true,
                _ => !desc,
            };
            columns.push(column);
            descending.push(desc);
            nulls_last.push(last);
            op_classes.push(operator_class(elem, ty, method)?);
        }
        let predicate = match stmt.where_clause.as_deref() {
            Some(node) => {
                let mut binder = Binder::new(self, crate::dml::table_scope(&table, None));
                (binder.clause, binder.definition) = ("index predicates", true);
                crate::expr::condition(binder.bind(node)?, "WHERE", arg_location(node))?;
                if !crate::pgcatalog::is_immutable(node) {
                    return Err(PgError::new(
                        code::INVALID_OBJECT_DEFINITION,
                        "functions in index predicate must be marked IMMUTABLE",
                    ));
                }
                expression_text(node)?
            }
            None => String::new(),
        };
        let names = unique_column_names(names);
        let Some(name) = self.index_name(stmt, &table, &names)? else { return Ok(Outcome::command("CREATE INDEX")) };
        if !stmt.index_including_params.is_empty() {
            return Err(PgError::unsupported("indexes with INCLUDE"));
        }
        let mut table = table;
        let mut tags = self.txn.all_tags(self.db)?;
        tags.extend(table.columns.iter().chain(&table.hidden).map(|c| c.tag));
        for (position, text, ty) in expressions {
            let kinds = vec![EXTENDED_KIND; table.columns.len() + table.hidden.len()];
            let column_name = format!("!hidden!{}!{position}!0", name.to_lowercase());
            let tag = auto_generate_tag(&tags, &table.name, &kinds, &column_name, EXTENDED_KIND);
            tags.insert(tag);
            table.hidden.push(ColumnDef {
                name: column_name,
                ty,
                tag,
                encoding: ty.encoding(),
                nullable: true,
                primary_key: false,
                default: format!("({text})"),
                generated: true,
                mysql_type: String::new(),
                comment: String::new(),
            });
        }
        let index = IndexDef { descending, nulls_last, op_classes, predicate, ..new_index(name, columns, stmt.unique) };
        self.build_index(table, index)?;
        Ok(Outcome::command("CREATE INDEX"))
    }

    /// index_expression checks an index expression over a table's columns as Postgres does, returning its type.
    fn index_expression(&mut self, table: &TableDef, expr: &Node) -> Result<ColumnType> {
        let mut binder = Binder::new(self, crate::dml::table_scope(table, None));
        (binder.clause, binder.definition) = ("index expressions", true);
        let (_, ty) = binder.bind(expr)?;
        if !crate::pgcatalog::is_immutable(expr) {
            return Err(PgError::new(
                code::INVALID_OBJECT_DEFINITION,
                "functions in index expression must be marked IMMUTABLE",
            ));
        }
        Ok(ColumnType { modifier: -1, ..ty })
    }

    /// index_name chooses the name of a new index of a table's columns, named by their column names or, for
    /// expressions, as Postgres names them, returning None after the notice of IF NOT EXISTS when the name is taken.
    pub(crate) fn index_name(
        &mut self,
        stmt: &IndexStmt,
        table: &TableDef,
        names: &[String],
    ) -> Result<Option<String>> {
        let taken = self.relation_names(&table.schema)?;
        let name = if stmt.idxname.is_empty() {
            choose_relation_name(&table.name, &names.join("_"), "idx", &taken)
        } else {
            stmt.idxname.clone()
        };
        if taken.contains(&name) {
            let message = format!("relation \"{name}\" already exists");
            if stmt.if_not_exists {
                self.session.notice(PgError::notice(code::DUPLICATE_TABLE, format!("{message}, skipping")));
                return Ok(None);
            }
            return Err(PgError::new(code::DUPLICATE_TABLE, message));
        }
        Ok(Some(name))
    }

    /// build_index adds an index to a table and fills it from the table's rows that its predicate holds for, failing
    /// for duplicates in a unique index as Postgres does.
    pub(crate) fn build_index(&mut self, mut table: TableDef, index: IndexDef) -> Result<()> {
        table.indexes.push(index.clone());
        let rules = self.index_rules(&table)?;
        let mut keys = Vec::new();
        for row in crate::query::scan(self.db, &table)? {
            let (row, held) = rules.indexed(self, &row)?;
            if !held[table.indexes.len() - 1] {
                continue;
            }
            let (primary, _) = table.encode_row(self.db, &row)?;
            keys.push((table.index_key(self.db, &index, &row, &primary)?, row.into_owned()));
        }
        keys.sort_by(|a, b| table.compare_index_keys(&index, &a.0, &b.0));
        if index.unique {
            let width = index.columns.len();
            for pair in keys.windows(2) {
                let value = |c: usize| &pair[0].1[table.row_position(c)];
                let null = index.columns.iter().any(|&c| value(c).is_null());
                if !null
                    && table.compare_index_prefix(&index, width, &pair[0].0, &pair[1].0) == std::cmp::Ordering::Equal
                {
                    let names: Vec<&str> = index.columns.iter().map(|&c| rules.column_name(&table, c)).collect();
                    let values: Vec<String> =
                        index.columns.iter().map(|&c| value(c).output().unwrap_or_default()).collect();
                    return Err(PgError {
                        detail: Some(format!("Key ({})=({}) is duplicated.", names.join(", "), values.join(", "))),
                        objects: Some(Box::new(crate::error::ErrorObjects {
                            schema: Some(table.schema.clone()),
                            table: Some(table.name.clone()),
                            constraint: Some(index.name.clone()),
                            ..Default::default()
                        })),
                        ..PgError::new(
                            code::UNIQUE_VIOLATION,
                            format!("could not create unique index \"{}\"", index.name),
                        )
                    });
                }
            }
        }
        let mut stored = table.table.clone();
        if let Some(distance) = index.vector {
            let root = table.write_vector_index(self.db, &index, distance)?;
            stored.put_index(self.db, &index.name, Some(root))?;
        } else {
            let empty = Hash::of(&empty_rows());
            stored.put_index(self.db, &index.name, Some(empty))?;
            keys.dedup_by(|a, b| table.compare_index_keys(&index, &a.0, &b.0) == std::cmp::Ordering::Equal);
            let edits = keys.into_iter().map(|(k, _)| (k, Some(prolly::val::build_tuple(&[])))).collect();
            let compare = |a: &[u8], b: &[u8]| table.compare_index_keys(&index, a, b);
            stored.edit_index(self.db, &index.name, empty, edits, &compare, &table.index_encodings(&index))?;
        }
        stored.schema = self.db.write_value(table.schema_message()?)?;
        let address = stored.write(self.db)?;
        self.txn.root.put_table(self.db, &table.schema, &table.name, Some(address))?;
        Ok(())
    }

    /// drop_indexes runs DROP INDEX.
    fn drop_indexes(&mut self, drop: &DropStmt) -> Result<Outcome> {
        let mut doomed = Vec::new();
        for object in &drop.objects {
            let Some(NodeEnum::List(list)) = object.node.as_ref() else { continue };
            let (schema, name) = object_names(&list.items);
            let schemas =
                if schema.is_empty() { self.session.search_path() } else { vec![self.session.named_schema(&schema)] };
            let mut found = None;
            'search: for s in schemas {
                let prefix = doltdb::root::table_key(&s, "");
                for (key, address) in self.txn.root.tables(self.db)? {
                    let Some(table) = key.strip_prefix(prefix.as_slice()) else { continue };
                    let table = TableDef::load(self.db, &s, &String::from_utf8_lossy(table), address)?;
                    if table.indexes.iter().any(|i| i.name == name && !i.system) {
                        found = Some(table);
                        break 'search;
                    }
                }
            }
            match found {
                Some(table) => {
                    self.require_owner(&Object::Table(table.schema.clone(), table.name.clone()))
                        .map_err(|err| PgError { message: format!("must be owner of index {name}"), ..err })?;
                    doomed.push((table, name))
                }
                None => {
                    let shown = if schema.is_empty() { name } else { format!("{schema}.{name}") };
                    self.missing_relation("index", &schema, &shown, drop.missing_ok, code::UNDEFINED_OBJECT)?;
                }
            }
        }
        let cascade = DropBehavior::try_from(drop.behavior) == Ok(DropBehavior::DropCascade);
        for (table, name) in doomed {
            let mut table = match self.txn.table(self.db, &table.schema, &table.name)? {
                Some(table) => table,
                None => continue,
            };
            self.drop_referencing_foreign_keys(&mut table, &name, &format!("index {name}"), cascade)?;
            table.drop_index(&name);
            let mut stored = table.table.clone();
            stored.put_index(self.db, &name, None)?;
            stored.schema = self.db.write_value(table.schema_message()?)?;
            let address = stored.write(self.db)?;
            self.txn.root.put_table(self.db, &table.schema, &table.name, Some(address))?;
            self.replace_child_index(&table.schema, &table.name, &name)?;
        }
        Ok(Outcome::command("DROP INDEX"))
    }
}

/// new_index returns an index of the columns, ascending with NULLs last, whose root the caller sets.
pub(crate) fn new_index(name: String, columns: Vec<usize>, unique: bool) -> IndexDef {
    let count = columns.len();
    IndexDef {
        name,
        columns,
        unique,
        descending: vec![false; count],
        nulls_last: vec![true; count],
        op_classes: vec![String::new(); count],
        comment: String::new(),
        predicate: String::new(),
        root: Hash::of(&empty_rows()),
        system: false,
        vector: None,
        deferrable: false,
        initially_deferred: false,
    }
}

/// operator_class returns the operator class that an index column names, checking that it exists for the access
/// method and accepts the column's type as Postgres does, or an empty name for the type's default.
fn operator_class(elem: &pg_query::protobuf::IndexElem, ty: ColumnType, method: &str) -> Result<String> {
    let names: Vec<&str> = elem.opclass.iter().filter_map(node_name).collect();
    let Some(&name) = names.last() else { return Ok(String::new()) };
    let method = if method.is_empty() { "btree" } else { method };
    let class = match names.as_slice() {
        [_] | ["pg_catalog", _] => {
            let method_oid = if method == "hash" { 405 } else { 403 };
            crate::pgcatalog::operator_classes(method_oid).into_iter().find(|c| c.name == name)
        }
        _ => None,
    };
    let Some(class) = class else {
        return Err(PgError::new(
            code::UNDEFINED_OBJECT,
            format!("operator class \"{}\" does not exist for access method \"{method}\"", names.join(".")),
        ));
    };
    if !crate::pgcatalog::binary_coercible(ty.oid, class.input) {
        return Err(PgError::new(
            code::DATATYPE_MISMATCH,
            format!("operator class \"{name}\" does not accept data type {}", type_display(ty.oid)),
        ));
    }
    if !elem.opclassopts.is_empty() {
        return Err(PgError::new(code::INVALID_PARAMETER_VALUE, format!("operator class {name} has no options")));
    }
    Ok(class.name)
}

/// unique_column_names returns the names of an index's columns with numbers added to repeated ones, as Postgres'
/// ChooseIndexColumnNames does.
pub(crate) fn unique_column_names(names: Vec<String>) -> Vec<String> {
    let mut out: Vec<String> = Vec::with_capacity(names.len());
    for name in names {
        let mut candidate = name.clone();
        let mut suffix = 1;
        while out.contains(&candidate) {
            let number = suffix.to_string();
            candidate = format!("{}{number}", clip(&name, NAMEDATALEN_MAX - number.len()));
            suffix += 1;
        }
        out.push(candidate);
    }
    out
}

/// NAMEDATALEN_MAX is the longest identifier Postgres keeps, in bytes.
pub(crate) const NAMEDATALEN_MAX: usize = 63;

/// clip returns the longest prefix of a name within a byte length that ends on a character boundary.
pub(crate) fn clip(name: &str, len: usize) -> &str {
    let mut end = len.min(name.len());
    while !name.is_char_boundary(end) {
        end -= 1;
    }
    &name[..end]
}

/// make_object_name joins two names and a label with underscores, shortening the longer name until the result fits,
/// as Postgres' makeObjectName does.
fn make_object_name(name1: &str, name2: &str, label: &str) -> String {
    let overhead = label.len() + 1 + if name2.is_empty() { 0 } else { 1 };
    let available = NAMEDATALEN_MAX - overhead;
    let (mut len1, mut len2) = (name1.len(), name2.len());
    while len1 + len2 > available {
        if len1 > len2 {
            len1 -= 1;
        } else {
            len2 -= 1;
        }
    }
    let (part1, part2) = (clip(name1, len1), clip(name2, len2));
    if part2.is_empty() { format!("{part1}_{label}") } else { format!("{part1}_{part2}_{label}") }
}

/// choose_relation_name returns a name from two names and a label that no existing relation takes, adding a number
/// to the label when needed, as Postgres' ChooseRelationName does.
pub(crate) fn choose_relation_name(name1: &str, name2: &str, label: &str, taken: &[String]) -> String {
    let mut name = make_object_name(name1, name2, label);
    let mut pass = 0;
    while taken.contains(&name) {
        pass += 1;
        name = make_object_name(name1, name2, &format!("{label}{pass}"));
    }
    name
}

/// multiple_primary_keys returns Postgres' error for a second primary key.
fn multiple_primary_keys(table: &str, location: i32) -> PgError {
    PgError {
        position: position(location),
        ..PgError::new(
            code::INVALID_TABLE_DEFINITION,
            format!("multiple primary keys for table \"{table}\" are not allowed"),
        )
    }
}
