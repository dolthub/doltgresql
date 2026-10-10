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

//! The objects that use what an extension declares, which DROP EXTENSION refuses to leave behind unless CASCADE drops
//! them, as Postgres reports them.

use pg_query::NodeRef;

use crate::catalog::id::{self, SECTION_TYPE};
use crate::engine::quote_identifier as quote;
use crate::error::{PgError, Result};
use crate::extensions::Extension;
use crate::query::Ctx;

/// Dependent is an object that uses an extension's object: its description, the description of what it uses, where
/// Postgres lists it, which is by the position of that object among the extension's objects with uses of a type's
/// array type first, and the statement that drops it for CASCADE.
pub struct Dependent {
    pub object: String,
    pub on: String,
    order: (usize, bool),
    pub cascade: String,
}

/// Member is an object that an extension declares, which expressions name: a routine or a type.
enum Member {
    Routine(String, usize),
    Type(String),
}

/// references returns the positions of the extension's members that a statement names, with whether each type that
/// it casts to is that type's array type.
fn references(members: &[Member], statement: &str) -> Vec<(usize, Option<bool>)> {
    let Ok(parsed) = pg_query::parse(statement) else { return Vec::new() };
    let mut found = Vec::new();
    for (node, ..) in parsed.protobuf.nodes() {
        match node {
            NodeRef::FuncCall(call) => {
                let name = call.funcname.last().and_then(crate::expr::node_name).unwrap_or_default();
                let matching = |arity: usize| {
                    members.iter().position(|m| matches!(m, Member::Routine(n, a) if n == name && *a == arity))
                };
                if let Some(i) = matching(call.args.len())
                    .or_else(|| members.iter().position(|m| matches!(m, Member::Routine(n, _) if n == name)))
                {
                    found.push((i, None));
                }
            }
            NodeRef::TypeCast(cast) => {
                let Some(type_name) = cast.type_name.as_ref() else { continue };
                let name = type_name.names.last().and_then(crate::expr::node_name).unwrap_or_default();
                if let Some(i) = members.iter().position(|m| matches!(m, Member::Type(n) if n == name)) {
                    found.push((i, Some(!type_name.array_bounds.is_empty())));
                }
            }
            _ => {}
        }
    }
    found
}

impl Ctx<'_> {
    /// extension_dependents returns the objects that use the objects of an extension installed in a schema, each once,
    /// in the order of the extension's objects that they use, as Postgres lists them.
    pub fn extension_dependents(&mut self, extension: &Extension, schema: &str) -> Result<Vec<Dependent>> {
        let mut members: Vec<Member> =
            extension.routines.iter().map(|r| Member::Routine(r.name.clone(), r.params.len())).collect();
        let first_type = members.len();
        members.extend(extension.types.iter().map(|t| Member::Type(t.name.to_string())));
        let type_oid = |name: &str| crate::catalog::oids::oid(&id::new(SECTION_TYPE, &[schema, name]));
        let routines = self.routines()?;
        let describe = |ctx: &mut Ctx<'_>, member: usize, array: Option<bool>| -> Result<String> {
            Ok(match &members[member] {
                Member::Routine(name, arity) => {
                    let routine =
                        routines.iter().find(|r| r.schema == schema && r.name == *name && r.inputs().count() == *arity);
                    let oid = routine.map(|r| crate::pgcatalog::routines::routine_oid(r)).unwrap_or_default();
                    let shown = ctx.proc_name(oid, crate::oid::REGPROCEDURE)?.unwrap_or_else(|| format!("{name}()"));
                    format!("function {shown}")
                }
                Member::Type(name) => {
                    let oid = type_oid(name);
                    let oid = match array {
                        Some(true) => crate::expr::array_of(oid),
                        _ => oid,
                    };
                    format!("type {}", crate::cast::type_display(oid))
                }
            })
        };
        let mut dependents: Vec<Dependent> = Vec::new();
        let mut add = |ctx: &mut Ctx<'_>, object: String, uses: Vec<(usize, Option<bool>)>, cascade: String| {
            let Some(&(member, array)) = uses.iter().min_by_key(|(m, _)| *m) else { return Ok(()) };
            let on = describe(ctx, member, array)?;
            dependents.push(Dependent { object, on, order: (member, array != Some(true)), cascade });
            Ok::<(), PgError>(())
        };
        let snapshot = self.snapshot()?;
        for table in &snapshot.tables {
            let table_name = format!("{}.{}", quote(&table.schema), quote(&table.name));
            for column in &table.columns {
                let element = crate::expr::element_type(column.ty.oid);
                let typed = extension.types.iter().position(|t| [column.ty.oid, element].contains(&type_oid(t.name)));
                let object = format!("column {} of table {}", column.name, table.name);
                let drop_column = format!("ALTER TABLE {table_name} DROP COLUMN {}", quote(&column.name));
                if let Some(i) = typed {
                    let array = type_oid(extension.types[i].name) != column.ty.oid;
                    add(self, object, vec![(first_type + i, Some(array))], drop_column)?;
                    continue;
                }
                if column.default.is_empty() {
                    continue;
                }
                let uses = references(&members, &format!("SELECT {}", column.default));
                match column.generated {
                    true => add(self, object, uses, drop_column)?,
                    false => add(
                        self,
                        format!("default value for column {} of table {}", column.name, table.name),
                        uses,
                        format!("ALTER TABLE {table_name} ALTER COLUMN {} DROP DEFAULT", quote(&column.name)),
                    )?,
                }
            }
            for check in &table.checks {
                add(
                    self,
                    format!("constraint {} on table {}", check.name, table.name),
                    references(&members, &format!("SELECT {}", check.expression)),
                    format!("ALTER TABLE {table_name} DROP CONSTRAINT {}", quote(&check.name)),
                )?;
            }
        }
        for view in &snapshot.views {
            add(
                self,
                format!("view {}", view.name),
                references(&members, &view.statement),
                format!("DROP VIEW {}.{}", quote(&view.schema), quote(&view.name)),
            )?;
        }
        for user_type in self.user_types()?.values() {
            let crate::usertypes::Kind::Domain(domain) = &user_type.kind else { continue };
            let domain_name = format!("{}.{}", quote(&user_type.schema), quote(&user_type.name));
            let drop_domain = format!("DROP DOMAIN {domain_name}");
            if let Some(i) = extension.types.iter().position(|t| type_oid(t.name) == domain.base.oid) {
                add(self, format!("type {}", user_type.name), vec![(first_type + i, None)], drop_domain)?;
                continue;
            }
            if let Some(default) = &domain.default {
                let uses = references(&members, &format!("SELECT {default}"));
                if !uses.is_empty() {
                    add(self, format!("type {}", user_type.name), uses, drop_domain)?;
                    continue;
                }
            }
            for (name, expression) in &domain.checks {
                add(
                    self,
                    format!("constraint {name}"),
                    references(&members, &format!("SELECT {expression}")),
                    format!("ALTER DOMAIN {domain_name} DROP CONSTRAINT {}", quote(name)),
                )?;
            }
        }
        dependents.sort_by_key(|d| d.order);
        Ok(dependents)
    }

    /// drop_extension_dependents runs the statements that CASCADE runs to drop objects, telling the client about each.
    pub fn drop_extension_dependents(&mut self, dependents: &[Dependent]) -> Result<()> {
        self.notice_cascades(dependents.iter().map(|d| format!("drop cascades to {}", d.object)).collect());
        for dependent in dependents {
            for statement in crate::parse::parse(&dependent.cascade)? {
                let crate::parse::Statement::Postgres { node, .. } = statement else {
                    return Err(PgError::internal("a cascading drop statement did not parse"));
                };
                self.run(&node)?;
            }
        }
        Ok(())
    }
}
