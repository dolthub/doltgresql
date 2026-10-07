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

//! ALTER TABLE and RENAME, which change a table's definition and rebuild its rows when their layout changes.

use doltdb::table::Table;
use doltdb::tags::{EXTENDED_KIND, auto_generate_tag};
use pg_query::protobuf::{
    AlterTableCmd, AlterTableStmt, AlterTableType, ConstrType, DropBehavior, ObjectType, RenameStmt,
};
use pg_query::{Node, NodeEnum};

use crate::Outcome;
use crate::catalog::table::{Check, HIDDEN_BASE, IndexDef, Primary, TableDef};
use crate::ddl::{TableParts, check_column, choose_relation_name, constraint_type, expression_text, new_index};
use crate::dml::{parse_expression, table_scope, write_rows};
use crate::error::{ErrorObjects, PgError, Result, code};
use crate::expr::{Binder, Expr, Scope, assign, coerce, resolve_type_name};
use crate::query::{Ctx, scan};
use crate::types::Value;

/// Alteration is a table being altered: its definition, its rows once a change needs them, whether its rows must be
/// rebuilt, and the foreign keys to add once it is written.
pub(crate) struct Alteration {
    table: TableDef,
    rows: Option<Vec<Vec<Value>>>,
    rebuild: bool,
    foreign: Vec<pg_query::protobuf::Constraint>,
}

impl Alteration {
    /// new starts altering a table.
    pub(crate) fn new(table: TableDef) -> Alteration {
        Alteration { table, rows: None, rebuild: false, foreign: Vec::new() }
    }
}

/// table_objects returns the schema and table that an error about a table names.
fn table_objects(table: &TableDef, column: Option<&str>, constraint: Option<&str>) -> Option<Box<ErrorObjects>> {
    Some(Box::new(ErrorObjects {
        schema: Some(table.schema.clone()),
        table: Some(table.name.clone()),
        column: column.map(str::to_string),
        constraint: constraint.map(str::to_string),
        ..ErrorObjects::default()
    }))
}

/// column_missing returns Postgres' error for a column that a table lacks.
fn column_missing(table: &TableDef, column: &str) -> PgError {
    PgError::new(code::UNDEFINED_COLUMN, format!("column \"{column}\" of relation \"{}\" does not exist", table.name))
}

/// rename_in_expression renames a column in the SQL text of a stored expression.
fn rename_in_expression(text: &str, old: &str, new: &str) -> Result<String> {
    let scan = pg_query::scan(text).map_err(PgError::internal)?;
    let mut out = String::new();
    let mut last = 0;
    for token in &scan.tokens {
        let (start, end) = (token.start as usize, token.end as usize);
        let word = &text[start..end];
        let unquoted =
            if word.starts_with('"') { word.trim_matches('"').replace("\"\"", "\"") } else { word.to_string() };
        if unquoted == old && (word.starts_with('"') || token.keyword_kind == 0) {
            out.push_str(&text[last..start]);
            out.push_str(&crate::engine::quote_identifier(new));
            last = end;
        }
    }
    out.push_str(&text[last..]);
    Ok(out)
}

/// references reports whether a stored expression refers to a column.
pub(crate) fn references(text: &str, column: &str) -> bool {
    parse_expression(text).is_ok_and(|node| {
        node.node.as_ref().is_some_and(|n| {
            n.nodes().into_iter().any(|(n, ..)| {
                matches!(n, pg_query::NodeRef::ColumnRef(c) if c.fields.iter().filter_map(crate::expr::node_name).next_back() == Some(column))
            })
        })
    })
}

impl Ctx<'_> {
    /// alter_table runs ALTER TABLE.
    pub fn alter_table(&mut self, stmt: &AlterTableStmt) -> Result<Outcome> {
        let relation = stmt.relation.as_ref().ok_or_else(|| PgError::internal("ALTER TABLE without a table"))?;
        if ObjectType::try_from(stmt.objtype) == Ok(ObjectType::ObjectSequence) {
            return self.alter_sequence_commands(stmt, relation);
        }
        if ObjectType::try_from(stmt.objtype) == Ok(ObjectType::ObjectView)
            && self.find_view(&relation.schemaname, &relation.relname)?.is_some()
        {
            for cmd in &stmt.cmds {
                let Some(NodeEnum::AlterTableCmd(cmd)) = cmd.node.as_ref() else { continue };
                if AlterTableType::try_from(cmd.subtype) != Ok(AlterTableType::AtChangeOwner) {
                    return Err(PgError::unsupported("this ALTER VIEW"));
                }
                self.check_new_owner(cmd.newowner.as_ref())?;
            }
            return Ok(Outcome::command("ALTER VIEW"));
        }
        let table = match self.resolve_table(relation) {
            Ok(table) => table,
            Err(_) if stmt.missing_ok => {
                self.session.notice(PgError::notice(
                    "00000",
                    format!("relation \"{}\" does not exist, skipping", relation.relname),
                ));
                return Ok(Outcome::command("ALTER TABLE"));
            }
            Err(err) => return Err(PgError { position: None, ..err }),
        };
        self.require_owner(&crate::auth::Object::Table(table.schema.clone(), table.name.clone()))?;
        let before: Vec<(u64, String)> = table.columns.iter().map(|c| (c.tag, c.name.clone())).collect();
        let mut alteration = Alteration::new(table);
        for cmd in &stmt.cmds {
            let Some(NodeEnum::AlterTableCmd(cmd)) = cmd.node.as_ref() else { continue };
            self.alter_command(&mut alteration, cmd)?;
        }
        let foreign = std::mem::take(&mut alteration.foreign);
        let (schema, name) = (alteration.table.schema.clone(), alteration.table.name.clone());
        self.finish_alteration(alteration)?;
        self.update_row_type_users(&schema, &name, &before)?;
        for constraint in foreign {
            let table = self
                .txn
                .table(self.db, &schema, &name)?
                .ok_or_else(|| PgError::internal("an altered table vanished"))?;
            let mut keys = Vec::new();
            for key in constraint.fk_attrs.iter().filter_map(crate::expr::node_name) {
                keys.push(table.columns.iter().position(|c| c.name == key).ok_or_else(|| {
                    PgError::new(
                        code::UNDEFINED_COLUMN,
                        format!("column \"{key}\" referenced in foreign key constraint does not exist"),
                    )
                })?);
            }
            self.add_foreign_key(&table, &keys, &constraint)?;
        }
        Ok(Outcome::command("ALTER TABLE"))
    }

    /// alter_sequence_commands runs the ALTER SEQUENCE forms that the ALTER TABLE grammar parses, such as OWNER TO.
    fn alter_sequence_commands(
        &mut self,
        stmt: &AlterTableStmt,
        relation: &pg_query::protobuf::RangeVar,
    ) -> Result<Outcome> {
        if self.find_sequence(relation)?.is_none() {
            let message = format!("relation \"{}\" does not exist", relation.relname);
            if stmt.missing_ok {
                self.session.notice(PgError::notice("00000", format!("{message}, skipping")));
                return Ok(Outcome::command("ALTER SEQUENCE"));
            }
            return Err(PgError::new(code::UNDEFINED_TABLE, message));
        }
        for cmd in &stmt.cmds {
            let Some(NodeEnum::AlterTableCmd(cmd)) = cmd.node.as_ref() else { continue };
            match AlterTableType::try_from(cmd.subtype) {
                Ok(AlterTableType::AtChangeOwner) => self.check_new_owner(cmd.newowner.as_ref())?,
                Ok(other) => return Err(PgError::unsupported(format!("ALTER SEQUENCE {other:?}"))),
                Err(_) => return Err(PgError::unsupported("this ALTER SEQUENCE")),
            }
        }
        Ok(Outcome::command("ALTER SEQUENCE"))
    }

    /// rows returns the altered table's rows, scanning them on first use.
    fn rows<'r>(&mut self, alteration: &'r mut Alteration) -> Result<&'r mut Vec<Vec<Value>>> {
        if alteration.rows.is_none() {
            alteration.rows = Some(scan(self.db, &alteration.table)?);
        }
        Ok(alteration.rows.as_mut().expect("rows were just scanned"))
    }

    /// alter_command applies one ALTER TABLE command.
    fn alter_command(&mut self, alteration: &mut Alteration, cmd: &AlterTableCmd) -> Result<()> {
        let kind = AlterTableType::try_from(cmd.subtype).unwrap_or(AlterTableType::Undefined);
        match kind {
            AlterTableType::AtAddColumn => self.add_column(alteration, cmd),
            AlterTableType::AtDropColumn => self.drop_column(alteration, cmd),
            AlterTableType::AtAddIdentity => self.add_identity(alteration, cmd),
            AlterTableType::AtColumnDefault => {
                let i = self.column_index(&alteration.table, &cmd.name)?;
                alteration.table.columns[i].default = match cmd.def.as_deref() {
                    Some(expr) => {
                        let column = alteration.table.columns[i].clone();
                        self.check_default(expr, &column).map_err(|err| PgError { position: None, ..err })?;
                        expression_text(expr)?
                    }
                    None => String::new(),
                };
                Ok(())
            }
            AlterTableType::AtSetNotNull => {
                let i = self.column_index(&alteration.table, &cmd.name)?;
                if self.rows(alteration)?.iter().any(|r| r[i].is_null()) {
                    let table = &alteration.table;
                    return Err(PgError {
                        objects: table_objects(table, Some(&cmd.name), None),
                        ..PgError::new(
                            code::NOT_NULL_VIOLATION,
                            format!("column \"{}\" of relation \"{}\" contains null values", cmd.name, table.name),
                        )
                    });
                }
                alteration.table.columns[i].nullable = false;
                Ok(())
            }
            AlterTableType::AtDropNotNull => {
                let i = self.column_index(&alteration.table, &cmd.name)?;
                if alteration.table.key_columns.contains(&i) {
                    return Err(PgError::new(
                        code::INVALID_TABLE_DEFINITION,
                        format!("column \"{}\" is in a primary key", cmd.name),
                    ));
                }
                alteration.table.columns[i].nullable = true;
                Ok(())
            }
            AlterTableType::AtAlterColumnType => self.alter_column_type(alteration, cmd),
            AlterTableType::AtAddConstraint => {
                let Some(NodeEnum::Constraint(constraint)) = cmd.def.as_deref().and_then(|d| d.node.as_ref()) else {
                    return Err(PgError::internal("ADD CONSTRAINT without a constraint"));
                };
                self.add_constraint(alteration, constraint)
            }
            AlterTableType::AtAlterConstraint => {
                let Some(NodeEnum::Constraint(constraint)) = cmd.def.as_deref().and_then(|d| d.node.as_ref()) else {
                    return Err(PgError::internal("ALTER CONSTRAINT without a constraint"));
                };
                self.alter_constraint(alteration, constraint)
            }
            AlterTableType::AtValidateConstraint => self.validate_constraint(alteration, &cmd.name),
            AlterTableType::AtDropConstraint => {
                let cascade = DropBehavior::try_from(cmd.behavior) == Ok(DropBehavior::DropCascade);
                self.drop_constraint(alteration, &cmd.name, cmd.missing_ok, cascade)
            }
            AlterTableType::AtChangeOwner => self.check_new_owner(cmd.newowner.as_ref()),
            AlterTableType::AtEnableTrig
            | AlterTableType::AtDisableTrig
            | AlterTableType::AtEnableTrigAll
            | AlterTableType::AtDisableTrigAll
            | AlterTableType::AtEnableTrigUser
            | AlterTableType::AtDisableTrigUser
            | AlterTableType::AtReplicaIdentity
            | AlterTableType::AtEnableRowSecurity
            | AlterTableType::AtDisableRowSecurity
            | AlterTableType::AtForceRowSecurity
            | AlterTableType::AtNoForceRowSecurity
            | AlterTableType::AtSetStatistics => Ok(()),
            other => Err(PgError::unsupported(format!("ALTER TABLE {other:?}"))),
        }
    }

    /// column_index returns a column's position, failing as Postgres does when the table lacks it.
    fn column_index(&self, table: &TableDef, name: &str) -> Result<usize> {
        table.columns.iter().position(|c| c.name == name).ok_or_else(|| column_missing(table, name))
    }

    /// add_identity runs ALTER COLUMN ADD GENERATED AS IDENTITY, which gives a column an owned sequence as its default.
    fn add_identity(&mut self, alteration: &mut Alteration, cmd: &AlterTableCmd) -> Result<()> {
        let i = self.column_index(&alteration.table, &cmd.name)?;
        let (table, column) = (alteration.table.name.clone(), alteration.table.columns[i].clone());
        let state = |problem: &str| {
            PgError::new(
                code::OBJECT_NOT_IN_PREREQUISITE_STATE,
                format!("column \"{}\" of relation \"{table}\" {problem}", column.name),
            )
        };
        if column.nullable {
            return Err(state("must be declared NOT NULL before identity can be added"));
        }
        if !column.default.is_empty() {
            return Err(state("already has a default value"));
        }
        let data_type = match column.ty.oid {
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
        let options = match cmd.def.as_deref().and_then(|d| d.node.as_ref()) {
            Some(NodeEnum::Constraint(constraint)) => constraint.options.clone(),
            _ => Vec::new(),
        };
        let schema = alteration.table.schema.clone();
        let mut taken = self.relation_names(&schema)?;
        let default = self.create_owned_sequence(&schema, &table, &column.name, data_type, &options, &mut taken)?;
        alteration.table.columns[i].default = default;
        Ok(())
    }

    /// add_column runs ADD COLUMN, filling existing rows with the column's default.
    fn add_column(&mut self, alteration: &mut Alteration, cmd: &AlterTableCmd) -> Result<()> {
        let Some(NodeEnum::ColumnDef(def)) = cmd.def.as_deref().and_then(|d| d.node.as_ref()) else {
            return Err(PgError::internal("ADD COLUMN without a column"));
        };
        let name = alteration.table.name.clone();
        if alteration.table.columns.iter().any(|c| c.name == def.colname) {
            let message = format!("column \"{}\" of relation \"{name}\" already exists", def.colname);
            if cmd.missing_ok {
                self.session.notice(PgError::notice(code::DUPLICATE_COLUMN, format!("{message}, skipping")));
                return Ok(());
            }
            return Err(PgError::new(code::DUPLICATE_COLUMN, message));
        }
        let mut parts = TableParts {
            columns: alteration.table.columns.clone(),
            primary_key: alteration.table.key_columns.clone(),
            ..TableParts::default()
        };
        if let Some(type_name) = &def.type_name {
            self.prepare_type(type_name)?;
        }
        parts.add_column(&name, def)?;
        let index = parts.columns.len() - 1;
        let mut column = parts.columns.pop().expect("a column was just added");
        for (_, expr) in &parts.defaults {
            self.check_default(expr, &column).map_err(|err| PgError { position: None, ..err })?;
        }
        if let Some((_, expr)) = parts.generation.first() {
            let mut columns = alteration.table.columns.clone();
            columns.push(column.clone());
            self.check_generation(expr, &name, &columns, columns.len() - 1)?;
        }
        let mut tags = self.txn.all_tags(self.db)?;
        tags.extend(alteration.table.columns.iter().map(|c| c.tag));
        let kinds = vec![EXTENDED_KIND; alteration.table.columns.len()];
        column.tag = auto_generate_tag(&tags, &name, &kinds, &column.name, EXTENDED_KIND);
        let schema = alteration.table.schema.clone();
        let mut taken = self.relation_names(&schema)?;
        for (_, data_type, options) in &parts.generated {
            column.default =
                self.create_owned_sequence(&schema, &name, &column.name, data_type, options, &mut taken)?;
            column.nullable = false;
        }
        if !column.default.is_empty() {
            self.check_row_type_unused(&alteration.table)?;
        }
        let primary = parts.primary_key.contains(&index);
        if !column.default.is_empty() || !column.nullable || primary {
            self.rows(alteration)?;
        }
        alteration.table.columns.push(column);
        let column = alteration.table.columns[index].clone();
        if !column.default.is_empty() {
            let node = parse_expression(&column.default)?;
            let mut existing = alteration.table.clone();
            existing.columns.truncate(index);
            let scope = if column.generated { table_scope(&existing, None) } else { Scope::default() };
            let bound = Binder::new(self, scope).bind(&node)?;
            let expr = assign(bound, column.ty, &column.name, -1)?.0;
            let rows = std::mem::take(self.rows(alteration)?);
            let mut filled = Vec::with_capacity(rows.len());
            for mut row in rows {
                let value = expr.eval(self, &row)?;
                row.push(value);
                filled.push(row);
            }
            alteration.rows = Some(filled);
            alteration.rebuild = true;
        } else if let Some(rows) = alteration.rows.as_mut() {
            rows.iter_mut().for_each(|r| r.push(Value::Null));
        }
        if !column.nullable && self.rows(alteration)?.iter().any(|r| r.get(index).is_none_or(Value::is_null)) {
            return Err(PgError {
                objects: table_objects(&alteration.table, Some(&column.name), None),
                ..PgError::new(
                    code::NOT_NULL_VIOLATION,
                    format!("column \"{}\" of relation \"{name}\" contains null values", column.name),
                )
            });
        }
        if alteration.rows.as_ref().is_some_and(|rows| rows.iter().any(|r| r.len() < alteration.table.columns.len())) {
            for row in alteration.rows.as_mut().expect("rows exist") {
                row.resize(alteration.table.columns.len(), Value::Null);
            }
        }
        if primary {
            self.set_primary_key(alteration, vec![index], parts.primary.clone())?;
        } else {
            alteration.table.value_columns.push(index);
        }
        for (constraint, expr) in parts.checks {
            self.add_check(alteration, &constraint, &expr, true)?;
        }
        for (constraint, keys, deferral) in parts.uniques {
            self.add_unique(alteration, &constraint, keys, deferral)?;
        }
        for (_, mut constraint) in parts.foreign {
            let name = pg_query::protobuf::String { sval: alteration.table.columns[index].name.clone() };
            constraint.fk_attrs = vec![Node { node: Some(NodeEnum::String(name)) }];
            alteration.foreign.push(constraint);
        }
        Ok(())
    }

    /// drop_column runs DROP COLUMN, dropping the indexes and checks that use the column.
    fn drop_column(&mut self, alteration: &mut Alteration, cmd: &AlterTableCmd) -> Result<()> {
        let Some(i) = alteration.table.columns.iter().position(|c| c.name == cmd.name) else {
            if cmd.missing_ok {
                self.session.notice(PgError::notice(
                    "00000",
                    format!(
                        "column \"{}\" of relation \"{}\" does not exist, skipping",
                        cmd.name, alteration.table.name
                    ),
                ));
                return Ok(());
            }
            return Err(column_missing(&alteration.table, &cmd.name));
        };
        let cascade = DropBehavior::try_from(cmd.behavior) == Ok(DropBehavior::DropCascade);
        let (schema, table) = (alteration.table.schema.clone(), alteration.table.name.clone());
        self.drop_trigger_column(&schema, &table, &cmd.name, cascade)?;
        self.drop_column_foreign_keys(&mut alteration.table, &cmd.name, cascade)?;
        self.rows(alteration)?;
        let table = &mut alteration.table;
        if table.key_columns.contains(&i) {
            table.key_columns.clear();
            for column in &mut table.columns {
                column.primary_key = false;
            }
        }
        let name = cmd.name.clone();
        let doomed: Vec<String> = table
            .indexes
            .iter()
            .filter(|index| {
                references(&index.predicate, &name)
                    || index.columns.iter().any(|&c| {
                        c == i || table.index_column(c).is_some_and(|h| h.generated && references(&h.default, &name))
                    })
            })
            .map(|index| index.name.clone())
            .collect();
        for index in doomed {
            table.drop_index(&index);
        }
        table.checks.retain(|check| !references(&check.expression, &name));
        table.columns.remove(i);
        let shift = |c: usize| if c > i && c < HIDDEN_BASE { c - 1 } else { c };
        for index in &mut table.indexes {
            index.columns = index.columns.iter().map(|&c| shift(c)).collect();
        }
        table.key_columns = table.key_columns.iter().map(|&c| shift(c)).collect();
        table.value_columns = (0..table.columns.len()).filter(|c| !table.key_columns.contains(c)).collect();
        for row in alteration.rows.as_mut().expect("rows were scanned") {
            row.remove(i);
        }
        alteration.rebuild = true;
        Ok(())
    }

    /// check_row_type_unused fails when a column of another table holds a table's row type, which a change that
    /// rewrites the table's rows cannot keep up to date.
    fn check_row_type_unused(&mut self, table: &TableDef) -> Result<()> {
        let row_type = crate::pgcatalog::row_type_oid(&table.schema, &table.name);
        for ((schema, name), address) in crate::dolt::procedures::table_map(self.db, &self.txn.root.clone())? {
            let other = TableDef::load(self.db, &schema, &name, address)?;
            if let Some(column) = other.columns.iter().find(|c| c.ty.oid == row_type) {
                return Err(PgError::new(
                    code::FEATURE_NOT_SUPPORTED,
                    format!(
                        "cannot alter table \"{}\" because column \"{name}.{}\" uses its row type",
                        table.name, column.name
                    ),
                ));
            }
        }
        Ok(())
    }

    /// update_row_type_users rewrites the columns of other tables that hold a table's row type after the table's
    /// columns, given as their tags and names, changed, keeping the fields of the columns that the table still has and
    /// leaving new ones NULL, as Go's AfterTableAddColumn and AfterTableDropColumn do.
    fn update_row_type_users(&mut self, schema: &str, name: &str, columns: &[(u64, String)]) -> Result<()> {
        let table =
            self.txn.table(self.db, schema, name)?.ok_or_else(|| PgError::internal("an altered table vanished"))?;
        if table.columns.iter().map(|c| (c.tag, c.name.clone())).eq(columns.iter().cloned()) {
            return Ok(());
        }
        let before: Vec<u64> = columns.iter().map(|(tag, _)| *tag).collect();
        let after: Vec<u64> = table.columns.iter().map(|c| c.tag).collect();
        let row_type = crate::pgcatalog::row_type_oid(schema, name);
        for ((other_schema, other_name), address) in
            crate::dolt::procedures::table_map(self.db, &self.txn.root.clone())?
        {
            let other = TableDef::load(self.db, &other_schema, &other_name, address)?;
            let users: Vec<usize> = (0..other.columns.len()).filter(|&i| other.columns[i].ty.oid == row_type).collect();
            if users.is_empty() {
                continue;
            }
            let mut rows = scan(self.db, &other)?;
            for row in &mut rows {
                for &i in &users {
                    if let Value::Composite(value) = &mut row[i] {
                        value.fields = after
                            .iter()
                            .map(|tag| before.iter().position(|b| b == tag).and_then(|k| value.fields.get(k).cloned()))
                            .map(|field| field.unwrap_or(Value::Null))
                            .collect();
                    }
                }
            }
            crate::usertypes::replace_row_type(&table);
            self.finish_alteration(Alteration { rows: Some(rows), rebuild: true, ..Alteration::new(other) })?;
        }
        Ok(())
    }

    /// alter_column_type runs ALTER COLUMN TYPE, converting each row's value with the USING expression or a cast.
    fn alter_column_type(&mut self, alteration: &mut Alteration, cmd: &AlterTableCmd) -> Result<()> {
        self.check_row_type_unused(&alteration.table)?;
        let Some(NodeEnum::ColumnDef(def)) = cmd.def.as_deref().and_then(|d| d.node.as_ref()) else {
            return Err(PgError::internal("ALTER COLUMN TYPE without a type"));
        };
        let i = self.column_index(&alteration.table, &cmd.name)?;
        let type_name = def.type_name.as_ref().ok_or_else(|| PgError::internal("ALTER COLUMN TYPE without a type"))?;
        let ty = resolve_type_name(type_name)?;
        if ty.oid == crate::oid::RECORD {
            return Err(PgError::new(
                code::INVALID_TABLE_DEFINITION,
                format!("column \"{}\" has pseudo-type record", cmd.name),
            ));
        }
        let old = alteration.table.columns[i].clone();
        let scope = table_scope(&alteration.table, None);
        let expr = match def.raw_default.as_deref() {
            Some(using) => {
                let bound = Binder::new(self, scope).bind(using)?;
                assign(bound, ty, &old.name, -1)?.0
            }
            None => {
                let shown = crate::cast::type_display(ty.oid);
                assign((Expr::Column(i), old.ty), ty, &old.name, -1)
                    .map_err(|_| PgError {
                        hint: Some(format!("You might need to specify \"USING {}::{shown}\".", old.name)),
                        ..PgError::new(
                            code::DATATYPE_MISMATCH,
                            format!("column \"{}\" cannot be cast automatically to type {shown}", old.name),
                        )
                    })?
                    .0
            }
        };
        let rows = std::mem::take(self.rows(alteration)?);
        let mut converted = Vec::with_capacity(rows.len());
        for mut row in rows {
            row[i] = expr.eval(self, &row)?;
            if row[i].is_null() && !old.nullable {
                return Err(PgError {
                    objects: table_objects(&alteration.table, Some(&old.name), None),
                    ..PgError::new(
                        code::NOT_NULL_VIOLATION,
                        format!(
                            "column \"{}\" of relation \"{}\" contains null values",
                            old.name, alteration.table.name
                        ),
                    )
                });
            }
            converted.push(row);
        }
        self.check_retyped_column(&alteration.table, &old.name, ty.oid)?;
        alteration.rows = Some(converted);
        let column = &mut alteration.table.columns[i];
        column.ty = ty;
        column.encoding = ty.encoding();
        alteration.rebuild = true;
        Ok(())
    }

    /// add_constraint runs ADD CONSTRAINT.
    fn add_constraint(
        &mut self,
        alteration: &mut Alteration,
        constraint: &pg_query::protobuf::Constraint,
    ) -> Result<()> {
        let mut parts = TableParts { columns: alteration.table.columns.clone(), ..TableParts::default() };
        match constraint_type(constraint) {
            ConstrType::ConstrPrimary => {
                if !alteration.table.key_columns.is_empty() {
                    return Err(PgError::new(
                        code::INVALID_TABLE_DEFINITION,
                        format!("multiple primary keys for table \"{}\" are not allowed", alteration.table.name),
                    ));
                }
                parts.add_constraint(&alteration.table.name, constraint)?;
                self.set_primary_key(alteration, parts.primary_key, parts.primary)
            }
            ConstrType::ConstrCheck | ConstrType::ConstrUnique => {
                parts.add_constraint(&alteration.table.name, constraint)?;
                for (name, expr) in parts.checks {
                    self.add_check(alteration, &name, &expr, !constraint.skip_validation)?;
                }
                for (name, keys, deferral) in parts.uniques {
                    self.add_unique(alteration, &name, keys, deferral)?;
                }
                Ok(())
            }
            ConstrType::ConstrForeign => {
                alteration.foreign.push(constraint.clone());
                Ok(())
            }
            other => Err(PgError::unsupported(format!("ADD CONSTRAINT {other:?}"))),
        }
    }

    /// alter_constraint runs ALTER CONSTRAINT, which changes a foreign key's deferral.
    fn alter_constraint(&mut self, alteration: &Alteration, constraint: &pg_query::protobuf::Constraint) -> Result<()> {
        let table = &alteration.table;
        let name = &constraint.conname;
        if self.set_foreign_key_deferral(table, name, crate::ddl::deferral(constraint))? {
            return Ok(());
        }
        let exists = table.checks.iter().any(|c| c.name == *name)
            || table.indexes.iter().any(|i| i.unique && i.name == *name)
            || !table.key_columns.is_empty() && table.primary_name() == *name;
        if exists {
            return Err(PgError::new(
                code::WRONG_OBJECT_TYPE,
                format!("constraint \"{name}\" of relation \"{}\" is not a foreign key constraint", table.name),
            ));
        }
        Err(PgError::new(
            code::UNDEFINED_OBJECT,
            format!("constraint \"{name}\" of relation \"{}\" does not exist", table.name),
        ))
    }

    /// add_check adds a check constraint, after checking every row against it unless it is NOT VALID.
    fn add_check(&mut self, alteration: &mut Alteration, name: &str, expr: &Node, validate: bool) -> Result<()> {
        let schema = alteration.table.schema.clone();
        let mut constraints = self.constraint_names(&schema)?;
        constraints.extend(alteration.table.checks.iter().map(|c| c.name.clone()));
        let name = if name.is_empty() {
            let column = check_column(expr);
            choose_relation_name(&alteration.table.name, column.as_deref().unwrap_or(""), "check", &constraints)
        } else if alteration.table.checks.iter().any(|c| c.name == name) {
            return Err(PgError::new(
                code::DUPLICATE_OBJECT,
                format!("constraint \"{name}\" for relation \"{}\" already exists", alteration.table.name),
            ));
        } else {
            name.to_string()
        };
        if validate {
            self.check_rows(alteration, &name, expr)?;
        } else {
            Binder::new(self, table_scope(&alteration.table, None)).bind(expr)?;
        }
        alteration.table.checks.push(Check { name, expression: expression_text(expr)? });
        Ok(())
    }

    /// check_rows fails as Postgres does when a row of the table breaks a check constraint.
    fn check_rows(&mut self, alteration: &mut Alteration, name: &str, expr: &Node) -> Result<()> {
        let bound = Binder::new(self, table_scope(&alteration.table, None)).bind(expr)?;
        let condition = coerce(bound, crate::expr::typ(crate::oid::BOOL), false, -1)?.0;
        let rows = self.rows(alteration)?.clone();
        for row in &rows {
            if condition.eval(self, row)? == Value::Bool(false) {
                return Err(PgError {
                    objects: table_objects(&alteration.table, None, Some(name)),
                    ..PgError::new(
                        code::CHECK_VIOLATION,
                        format!(
                            "check constraint \"{name}\" of relation \"{}\" is violated by some row",
                            alteration.table.name
                        ),
                    )
                });
            }
        }
        Ok(())
    }

    /// validate_constraint runs VALIDATE CONSTRAINT, checking every row against a check constraint or foreign key.
    fn validate_constraint(&mut self, alteration: &mut Alteration, name: &str) -> Result<()> {
        if let Some(check) = alteration.table.checks.iter().find(|c| c.name == name).cloned() {
            let expr = crate::parse::expression_node(&check.expression)?;
            return self.check_rows(alteration, name, &expr);
        }
        if self.validate_foreign_key(&alteration.table, name)? {
            return Ok(());
        }
        Err(PgError::new(
            code::UNDEFINED_OBJECT,
            format!("constraint \"{name}\" of relation \"{}\" does not exist", alteration.table.name),
        ))
    }

    /// add_unique adds a unique index after checking the rows for duplicates.
    fn add_unique(
        &mut self,
        alteration: &mut Alteration,
        name: &str,
        keys: Vec<usize>,
        (deferrable, initially_deferred): crate::ddl::Deferral,
    ) -> Result<()> {
        let schema = alteration.table.schema.clone();
        let mut taken = self.relation_names(&schema)?;
        taken.extend(alteration.table.indexes.iter().map(|i| i.name.clone()));
        let name = if name.is_empty() {
            let names: Vec<&str> = keys.iter().map(|&k| alteration.table.columns[k].name.as_str()).collect();
            choose_relation_name(&alteration.table.name, &names.join("_"), "key", &taken)
        } else if taken.iter().any(|t| t == name) {
            return Err(PgError::new(code::DUPLICATE_TABLE, format!("relation \"{name}\" already exists")));
        } else {
            name.to_string()
        };
        self.check_duplicates(alteration, &name, &keys)?;
        alteration.table.indexes.push(IndexDef { deferrable, initially_deferred, ..new_index(name, keys, true) });
        alteration.rebuild = true;
        Ok(())
    }

    /// check_duplicates fails as creating a unique index does when two rows share non-NULL values in the columns.
    fn check_duplicates(&mut self, alteration: &mut Alteration, name: &str, keys: &[usize]) -> Result<()> {
        let rows = self.rows(alteration)?.clone();
        let mut seen: Vec<Vec<Value>> = Vec::new();
        for row in rows {
            let values: Vec<Value> = keys.iter().map(|&k| row[k].clone()).collect();
            if values.iter().any(Value::is_null) {
                continue;
            }
            if seen.iter().any(|s| crate::plan::rows_equal(s, &values)) {
                let names: Vec<&str> = keys.iter().map(|&k| alteration.table.columns[k].name.as_str()).collect();
                let shown: Vec<String> = values.iter().map(|v| v.output().unwrap_or_default()).collect();
                return Err(PgError {
                    detail: Some(format!("Key ({})=({}) is duplicated.", names.join(", "), shown.join(", "))),
                    objects: table_objects(&alteration.table, None, Some(name)),
                    ..PgError::new(code::UNIQUE_VIOLATION, format!("could not create unique index \"{name}\""))
                });
            }
            seen.push(values);
        }
        Ok(())
    }

    /// set_primary_key makes columns the table's primary key, checking the rows for NULLs and duplicates.
    fn set_primary_key(&mut self, alteration: &mut Alteration, keys: Vec<usize>, primary: Primary) -> Result<()> {
        let table_name = alteration.table.name.clone();
        for &k in &keys {
            if self.rows(alteration)?.iter().any(|r| r.get(k).is_none_or(Value::is_null)) {
                let column = alteration.table.columns[k].name.clone();
                return Err(PgError {
                    objects: table_objects(&alteration.table, Some(&column), None),
                    ..PgError::new(
                        code::NOT_NULL_VIOLATION,
                        format!("column \"{column}\" of relation \"{table_name}\" contains null values"),
                    )
                });
            }
        }
        let name = if primary.name.is_empty() { format!("{table_name}_pkey") } else { primary.name.clone() };
        self.check_duplicates(alteration, &name, &keys)?;
        let table = &mut alteration.table;
        table.primary = primary;
        for &k in &keys {
            table.columns[k].primary_key = true;
            table.columns[k].nullable = false;
        }
        table.key_columns = keys;
        table.value_columns = (0..table.columns.len()).filter(|c| !table.key_columns.contains(c)).collect();
        alteration.rebuild = true;
        Ok(())
    }

    /// drop_constraint runs DROP CONSTRAINT on a check, unique, primary key, or foreign key constraint.
    fn drop_constraint(
        &mut self,
        alteration: &mut Alteration,
        name: &str,
        missing_ok: bool,
        cascade: bool,
    ) -> Result<()> {
        if self.drop_foreign_key(&mut alteration.table, name)? {
            return Ok(());
        }
        let table = &mut alteration.table;
        if let Some(i) = table.checks.iter().position(|c| c.name == name) {
            table.checks.remove(i);
            return Ok(());
        }
        let referenced = if table.key_columns.is_empty() || name != table.primary_name() {
            table.indexes.iter().find(|ix| ix.unique && ix.name == name).map(|ix| ix.name.clone())
        } else {
            Some(String::new())
        };
        if let Some(index) = referenced {
            let object = format!("constraint {name} on table {}", alteration.table.name);
            self.drop_referencing_foreign_keys(&mut alteration.table, &index, &object, cascade)?;
        }
        let table = &mut alteration.table;
        if let Some(i) = table.indexes.iter().position(|ix| ix.unique && ix.name == name) {
            table.indexes.remove(i);
            alteration.rebuild = true;
            return Ok(());
        }
        if !table.key_columns.is_empty() && name == table.primary_name() {
            self.rows(alteration)?;
            let table = &mut alteration.table;
            for column in &mut table.columns {
                column.primary_key = false;
            }
            table.key_columns.clear();
            table.primary = crate::catalog::table::Primary::default();
            table.value_columns = (0..table.columns.len()).collect();
            alteration.rebuild = true;
            return Ok(());
        }
        let message = format!("constraint \"{name}\" of relation \"{}\" does not exist", alteration.table.name);
        if missing_ok {
            self.session.notice(PgError::notice("00000", format!("{message}, skipping")));
            return Ok(());
        }
        Err(PgError::new(code::UNDEFINED_OBJECT, message))
    }

    /// finish_alteration writes the altered table: its schema alone, or a rebuilt table with its rows.
    pub(crate) fn finish_alteration(&mut self, mut alteration: Alteration) -> Result<()> {
        let (schema, name) = (alteration.table.schema.clone(), alteration.table.name.clone());
        let message = alteration.table.schema_message()?;
        if !alteration.rebuild {
            let mut stored = alteration.table.table.clone();
            stored.schema = self.db.write_value(message)?;
            let address = stored.write(self.db)?;
            return Ok(self.txn.root.put_table(self.db, &schema, &name, Some(address))?);
        }
        let rows = match alteration.rows.take() {
            Some(rows) => rows,
            None => scan(self.db, &alteration.table)?,
        };
        let (mut address, mut stored) = Table::create(self.db, message)?;
        for index in &mut alteration.table.indexes {
            let empty = index.empty_root(self.db)?;
            stored.put_index(self.db, &index.name, Some(empty))?;
            index.root = empty;
        }
        if !alteration.table.indexes.is_empty() {
            address = stored.write(self.db)?;
        }
        self.txn.root.put_table(self.db, &schema, &name, Some(address))?;
        let table =
            self.txn.table(self.db, &schema, &name)?.ok_or_else(|| PgError::internal("an altered table vanished"))?;
        write_rows(self, &table, &rows)
    }

    /// rename runs ALTER ... RENAME for tables, columns, constraints, indexes, and sequences.
    pub fn rename(&mut self, stmt: &RenameStmt) -> Result<Outcome> {
        let kind = ObjectType::try_from(stmt.rename_type).unwrap_or(ObjectType::Undefined);
        let tag = match kind {
            ObjectType::ObjectIndex => "ALTER INDEX",
            ObjectType::ObjectSequence => "ALTER SEQUENCE",
            _ => "ALTER TABLE",
        };
        if kind == ObjectType::ObjectRole {
            return self.rename_role(&stmt.subname, &stmt.newname);
        }
        let relation = stmt.relation.as_ref().ok_or_else(|| PgError::unsupported("this RENAME"))?;
        if kind == ObjectType::ObjectIndex {
            return self.rename_index(relation, &stmt.newname, stmt.missing_ok).map(|_| Outcome::command(tag));
        }
        if kind == ObjectType::ObjectSequence {
            return Err(PgError::unsupported("renaming sequences"));
        }
        let table = match self.resolve_table(relation) {
            Ok(table) => table,
            Err(err)
                if matches!(kind, ObjectType::ObjectTable | ObjectType::ObjectView)
                    && self.find_view(&relation.schemaname, &relation.relname)?.is_some() =>
            {
                let (schema, fragment) = self.find_view(&relation.schemaname, &relation.relname)?.ok_or(err)?;
                self.rename_view(&schema, &relation.relname, &stmt.newname, &fragment)?;
                return Ok(Outcome::command(if kind == ObjectType::ObjectView { "ALTER VIEW" } else { tag }));
            }
            Err(_) if stmt.missing_ok => {
                self.session.notice(PgError::notice(
                    "00000",
                    format!("relation \"{}\" does not exist, skipping", relation.relname),
                ));
                return Ok(Outcome::command(tag));
            }
            Err(err) => return Err(PgError { position: None, ..err }),
        };
        let mut alteration = Alteration::new(table);
        match kind {
            ObjectType::ObjectTable => {
                let (schema, old) = (alteration.table.schema.clone(), alteration.table.name.clone());
                if self.relation_names(&schema)?.contains(&stmt.newname) {
                    return Err(PgError::new(
                        code::DUPLICATE_TABLE,
                        format!("relation \"{}\" already exists", stmt.newname),
                    ));
                }
                let fks = self.foreign_keys()?;
                let address = self.txn.root.table(self.db, &schema, &old)?;
                self.txn.root.put_table(self.db, &schema, &old, None)?;
                self.txn.root.put_table(self.db, &schema, &stmt.newname, address)?;
                self.move_owned_sequences(&schema, &old, &stmt.newname, None)?;
                self.rename_in_foreign_keys(fks, &schema, &old, &stmt.newname)?;
                let mut auth = self.auth()?;
                auth.rename_object(
                    &crate::auth::Object::Table(schema.clone(), old),
                    &crate::auth::Object::Table(schema, stmt.newname.clone()),
                );
                auth.persist()?;
                return Ok(Outcome::command(tag));
            }
            ObjectType::ObjectColumn => {
                let before: Vec<(u64, String)> =
                    alteration.table.columns.iter().map(|c| (c.tag, c.name.clone())).collect();
                let i = self.column_index(&alteration.table, &stmt.subname)?;
                if alteration.table.columns.iter().any(|c| c.name == stmt.newname) {
                    return Err(PgError::new(
                        code::DUPLICATE_COLUMN,
                        format!("column \"{}\" of relation \"{}\" already exists", stmt.newname, alteration.table.name),
                    ));
                }
                alteration.table.columns[i].name = stmt.newname.clone();
                for check in &mut alteration.table.checks {
                    check.expression = rename_in_expression(&check.expression, &stmt.subname, &stmt.newname)?;
                }
                for column in &mut alteration.table.hidden {
                    column.default = rename_in_expression(&column.default, &stmt.subname, &stmt.newname)?;
                }
                for index in alteration.table.indexes.iter_mut().filter(|i| !i.predicate.is_empty()) {
                    index.predicate = rename_in_expression(&index.predicate, &stmt.subname, &stmt.newname)?;
                }
                let (schema, name) = (alteration.table.schema.clone(), alteration.table.name.clone());
                self.move_owned_sequences(&schema, &name, &name, Some((&stmt.subname, &stmt.newname)))?;
                let fks = self.foreign_keys()?;
                self.finish_alteration(alteration)?;
                self.rename_in_foreign_keys(fks, &schema, &name, &name)?;
                self.rename_trigger_column(&schema, &name, &stmt.subname, &stmt.newname)?;
                self.update_row_type_users(&schema, &name, &before)?;
                return Ok(Outcome::command(tag));
            }
            ObjectType::ObjectTabconstraint => {
                if self.rename_foreign_key(&alteration.table, &stmt.subname, &stmt.newname)? {
                    return Ok(Outcome::command(tag));
                }
                let table = &mut alteration.table;
                if let Some(check) = table.checks.iter_mut().find(|c| c.name == stmt.subname) {
                    check.name = stmt.newname.clone();
                } else if let Some(index) = table.indexes.iter().position(|ix| ix.name == stmt.subname) {
                    let root = table.indexes[index].root;
                    table.indexes[index].name = stmt.newname.clone();
                    table.table.put_index(self.db, &stmt.subname, None)?;
                    table.table.put_index(self.db, &stmt.newname, Some(root))?;
                } else if !table.key_columns.is_empty() && stmt.subname == table.primary_name() {
                    let default = stmt.newname == format!("{}_pkey", table.name);
                    table.primary.name = if default { String::new() } else { stmt.newname.clone() };
                } else {
                    return Err(PgError::new(
                        code::UNDEFINED_OBJECT,
                        format!("constraint \"{}\" for table \"{}\" does not exist", stmt.subname, table.name),
                    ));
                }
            }
            other => return Err(PgError::unsupported(format!("renaming {other:?}"))),
        }
        self.finish_alteration(alteration)?;
        Ok(Outcome::command(tag))
    }

    /// rename_index runs ALTER INDEX RENAME TO.
    fn rename_index(&mut self, relation: &pg_query::protobuf::RangeVar, new: &str, missing_ok: bool) -> Result<()> {
        let schemas =
            if relation.schemaname.is_empty() { self.session.search_path() } else { vec![relation.schemaname.clone()] };
        for schema in schemas {
            let prefix = doltdb::root::table_key(&schema, "");
            for (key, address) in self.txn.root.tables(self.db)? {
                let Some(name) = key.strip_prefix(prefix.as_slice()) else { continue };
                let mut table = TableDef::load(self.db, &schema, &String::from_utf8_lossy(name), address)?;
                let Some(index) = table.indexes.iter().position(|ix| ix.name == relation.relname) else { continue };
                if self.relation_names(&schema)?.iter().any(|n| n == new) {
                    return Err(PgError::new(code::DUPLICATE_TABLE, format!("relation \"{new}\" already exists")));
                }
                let root = table.indexes[index].root;
                table.indexes[index].name = new.to_string();
                table.table.put_index(self.db, &relation.relname, None)?;
                table.table.put_index(self.db, new, Some(root))?;
                return self.finish_alteration(Alteration::new(table));
            }
        }
        let shown = match relation.schemaname.as_str() {
            "" => relation.relname.clone(),
            schema => format!("{schema}.{}", relation.relname),
        };
        if missing_ok {
            self.session.notice(PgError::notice("00000", format!("relation \"{shown}\" does not exist, skipping")));
            return Ok(());
        }
        Err(PgError::new(code::UNDEFINED_TABLE, format!("relation \"{shown}\" does not exist")))
    }

    /// move_owned_sequences points the sequences a table owns at its new name, or one column at its new name.
    fn move_owned_sequences(&mut self, schema: &str, old: &str, new: &str, column: Option<(&str, &str)>) -> Result<()> {
        use crate::catalog::id::{self, SECTION_TABLE};
        let owner = id::new(SECTION_TABLE, &[schema, old]);
        for mut sequence in crate::sequences::all(self.db, &self.txn.root)? {
            if sequence.owner_table != owner {
                continue;
            }
            sequence.owner_table = id::new(SECTION_TABLE, &[schema, new]);
            if let Some((from, to)) = column
                && sequence.owner_column == from.as_bytes()
            {
                sequence.owner_column = to.as_bytes().to_vec();
            }
            crate::sequences::store(self.db, &mut self.txn.root, &sequence)?;
        }
        Ok(())
    }
}
