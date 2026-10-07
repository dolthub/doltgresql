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

//! Role statements, GRANT and REVOKE, and the privilege checks of other statements.

use std::sync::MutexGuard;

use pg_query::protobuf::a_const::Val;
use pg_query::protobuf::{
    AlterRoleStmt, CreateRoleStmt, DropRoleStmt, GrantRoleStmt, GrantStmt, GrantTargetType, ObjectType, RoleSpec,
    RoleSpecType, RoleStmtType,
};
use pg_query::{Node, NodeEnum};

use crate::Outcome;
use crate::auth::{AuthDb, Membership, Object, PUBLIC, Password, Role, insufficient};
use crate::catalog::table::TableDef;
use crate::error::{PgError, Result, code};
use crate::expr::{node_name, position};
use crate::query::Ctx;

/// TABLE_PRIVILEGES, SCHEMA_PRIVILEGES, SEQUENCE_PRIVILEGES, DATABASE_PRIVILEGES, and ROUTINE_PRIVILEGES are the
/// privilege letters that ALL grants on each kind of object.
const TABLE_PRIVILEGES: &[&str] = &["a", "r", "w", "d", "D", "x", "t"];
const SCHEMA_PRIVILEGES: &[&str] = &["U", "C"];
const SEQUENCE_PRIVILEGES: &[&str] = &["r", "w", "U"];
const DATABASE_PRIVILEGES: &[&str] = &["C", "T", "c"];
const ROUTINE_PRIVILEGES: &[&str] = &["X"];

/// privilege_letter returns the letter of a privilege named in GRANT or REVOKE.
fn privilege_letter(name: &str) -> Option<&'static str> {
    Some(match name {
        "select" => "r",
        "insert" => "a",
        "update" => "w",
        "delete" => "d",
        "truncate" => "D",
        "references" => "x",
        "trigger" => "t",
        "create" => "C",
        "connect" => "c",
        "temporary" | "temp" => "T",
        "execute" => "X",
        "usage" => "U",
        _ => return None,
    })
}

impl Ctx<'_> {
    /// alter_owner runs ALTER ... OWNER TO for the objects whose ownership Doltgres does not record, checking the new
    /// owner and then the object as Postgres' ExecAlterOwnerStmt does.
    pub fn alter_owner(&mut self, stmt: &pg_query::protobuf::AlterOwnerStmt) -> Result<crate::Outcome> {
        use pg_query::protobuf::ObjectType as O;
        self.check_new_owner(stmt.newowner.as_ref())?;
        let kind = O::try_from(stmt.object_type).unwrap_or(O::Undefined);
        let object = stmt.object.as_deref().and_then(|o| o.node.as_ref());
        let names: Vec<String> = match object {
            Some(pg_query::NodeEnum::String(s)) => vec![s.sval.clone()],
            Some(pg_query::NodeEnum::List(list)) => {
                list.items.iter().filter_map(crate::expr::node_name).map(str::to_string).collect()
            }
            _ => Vec::new(),
        };
        let name = names.last().cloned().unwrap_or_default();
        let tag = match kind {
            O::ObjectDatabase => {
                if !self.session.engine.database_exists(&name) {
                    return Err(PgError::new(
                        code::INVALID_CATALOG_NAME,
                        format!("database \"{name}\" does not exist"),
                    ));
                }
                "ALTER DATABASE"
            }
            O::ObjectSchema => {
                if !self.txn.root.schemas.iter().any(|s| *s == name.as_bytes()) {
                    return Err(PgError::new(code::INVALID_SCHEMA_NAME, format!("schema \"{name}\" does not exist")));
                }
                "ALTER SCHEMA"
            }
            O::ObjectType | O::ObjectDomain => {
                let schema = (names.len() > 1).then(|| names[names.len() - 2].as_str());
                if crate::usertypes::lookup(schema, &name).is_none()
                    && crate::catalog::builtin_type_named(&name).is_none()
                {
                    return Err(PgError::new(
                        code::UNDEFINED_OBJECT,
                        format!("type \"{}\" does not exist", names.join(".")),
                    ));
                }
                if kind == O::ObjectType { "ALTER TYPE" } else { "ALTER DOMAIN" }
            }
            O::ObjectFunction | O::ObjectProcedure | O::ObjectRoutine => {
                if let Some(pg_query::NodeEnum::ObjectWithArgs(target)) = object {
                    let procedures = match kind {
                        O::ObjectFunction => Some(false),
                        O::ObjectProcedure => Some(true),
                        _ => None,
                    };
                    self.find_routine(target, false, procedures)?;
                }
                match kind {
                    O::ObjectFunction => "ALTER FUNCTION",
                    O::ObjectProcedure => "ALTER PROCEDURE",
                    _ => "ALTER ROUTINE",
                }
            }
            _ => return Err(PgError::unsupported("this ALTER OWNER")),
        };
        Ok(crate::Outcome::command(tag))
    }

    /// check_new_owner fails as Postgres does when OWNER TO names a role that does not exist.
    pub(crate) fn check_new_owner(&mut self, owner: Option<&pg_query::protobuf::RoleSpec>) -> Result<()> {
        let Some(owner) = owner else { return Ok(()) };
        let named = pg_query::protobuf::RoleSpecType::try_from(owner.roletype)
            == Ok(pg_query::protobuf::RoleSpecType::RolespecCstring);
        if named && self.auth()?.role(&owner.rolename).is_none() {
            return Err(role_does_not_exist(&owner.rolename));
        }
        Ok(())
    }
}

/// role_does_not_exist returns Postgres' error for a missing role.
fn role_does_not_exist(name: &str) -> PgError {
    PgError::new(code::UNDEFINED_OBJECT, format!("role \"{name}\" does not exist"))
}

/// option_bool reads a role option's boolean argument.
fn option_bool(arg: Option<&Node>) -> bool {
    match arg.and_then(|a| a.node.as_ref()) {
        Some(NodeEnum::Boolean(b)) => b.boolval,
        Some(NodeEnum::Integer(i)) => i.ival != 0,
        _ => true,
    }
}

/// option_text reads a role option's string argument, which is None for NULL.
fn option_text(arg: Option<&Node>) -> Option<String> {
    match arg.and_then(|a| a.node.as_ref()) {
        Some(NodeEnum::String(s)) => Some(s.sval.clone()),
        Some(NodeEnum::AConst(c)) => match &c.val {
            Some(Val::Sval(s)) => Some(s.sval.clone()),
            _ => None,
        },
        _ => None,
    }
}

/// option_int reads a role option's integer argument.
fn option_int(arg: Option<&Node>) -> i32 {
    match arg.and_then(|a| a.node.as_ref()) {
        Some(NodeEnum::Integer(i)) => i.ival,
        _ => -1,
    }
}

impl Ctx<'_> {
    /// auth locks the roles and privileges.
    pub fn auth(&self) -> Result<MutexGuard<'_, AuthDb>> {
        self.session.auth.lock().map_err(|_| PgError::internal("the auth lock was poisoned"))
    }

    /// current_role returns the current role, which a superuser stands in for when it no longer exists.
    pub fn current_role(&self) -> Result<Role> {
        let auth = self.auth()?;
        auth.role(&self.session.role).cloned().ok_or_else(|| role_does_not_exist(&self.session.role))
    }

    /// role_name resolves a role specification to a role name.
    fn role_name(&self, spec: &RoleSpec) -> String {
        match RoleSpecType::try_from(spec.roletype) {
            Ok(RoleSpecType::RolespecCurrentRole | RoleSpecType::RolespecCurrentUser) => self.session.role.clone(),
            Ok(RoleSpecType::RolespecSessionUser) => self.session.user.clone(),
            Ok(RoleSpecType::RolespecPublic) => PUBLIC.into(),
            _ => spec.rolename.clone(),
        }
    }

    /// apply_role_options sets a role's attributes from CREATE ROLE's or ALTER ROLE's options.
    fn apply_role_options(&self, role: &mut Role, options: &[Node]) -> Result<()> {
        for option in options {
            let Some(NodeEnum::DefElem(def)) = option.node.as_ref() else { continue };
            let arg = def.arg.as_deref();
            match def.defname.as_str() {
                "password" => role.password = option_text(arg).map(|p| Password::new(&p)),
                "superuser" => role.superuser = option_bool(arg),
                "createdb" => role.create_db = option_bool(arg),
                "createrole" => role.create_role = option_bool(arg),
                "inherit" => role.inherit = option_bool(arg),
                "canlogin" => role.login = option_bool(arg),
                "isreplication" => role.replication = option_bool(arg),
                "bypassrls" => role.bypass_rls = option_bool(arg),
                "connectionlimit" => role.connection_limit = option_int(arg),
                "validUntil" => {
                    role.valid_until = match option_text(arg) {
                        Some(text) => match crate::cast::input(&text, crate::oid::TIMESTAMPTZ)? {
                            crate::types::Value::TimestampTz(t) => Some(t + 946_684_800_000_000),
                            _ => None,
                        },
                        None => None,
                    }
                }
                "addroleto" | "rolemembers" | "adminmembers" => {}
                other => return Err(PgError::unsupported(format!("the role option {other}"))),
            }
        }
        Ok(())
    }

    /// create_role runs CREATE ROLE, CREATE USER, and CREATE GROUP, which Doltgres also allows with IF NOT EXISTS.
    pub fn create_role(&mut self, stmt: &CreateRoleStmt, if_not_exists: bool) -> Result<Outcome> {
        let actor = self.current_role()?;
        if stmt.role == PUBLIC || stmt.role.starts_with("pg_") {
            return Err(PgError {
                detail: stmt.role.starts_with("pg_").then(|| "Role names starting with \"pg_\" are reserved.".into()),
                ..PgError::new(code::RESERVED_NAME, format!("role name \"{}\" is reserved", stmt.role))
            });
        }
        let mut auth = self.auth()?;
        if auth.role(&stmt.role).is_some() {
            let message = format!("role \"{}\" already exists", stmt.role);
            if if_not_exists {
                drop(auth);
                self.session.notice(PgError::notice(code::DUPLICATE_OBJECT, format!("{message}, skipping")));
                return Ok(Outcome::command("CREATE ROLE"));
            }
            return Err(PgError::new(code::DUPLICATE_OBJECT, message));
        }
        if !actor.superuser && !actor.create_role {
            return Err(insufficient("permission denied to create role"));
        }
        let mut role = Role::new(auth.next_id(), &stmt.role);
        role.login = RoleStmtType::try_from(stmt.stmt_type) == Ok(RoleStmtType::RolestmtUser);
        self.apply_role_options(&mut role, &stmt.options)?;
        if role.superuser && !actor.superuser {
            return Err(insufficient("must be superuser to create superusers"));
        }
        let id = role.id;
        auth.roles.insert(id, role);
        for option in &stmt.options {
            let Some(NodeEnum::DefElem(def)) = option.node.as_ref() else { continue };
            let names: Vec<String> = match def.arg.as_deref().and_then(|a| a.node.as_ref()) {
                Some(NodeEnum::List(list)) => list
                    .items
                    .iter()
                    .filter_map(|n| match n.node.as_ref() {
                        Some(NodeEnum::RoleSpec(spec)) => Some(spec.rolename.clone()),
                        _ => None,
                    })
                    .collect(),
                _ => continue,
            };
            for name in names {
                let other = auth.role(&name).ok_or_else(|| role_does_not_exist(&name))?.id;
                let (member, group, admin) = match def.defname.as_str() {
                    "addroleto" => (id, other, false),
                    "rolemembers" => (other, id, false),
                    _ => (other, id, true),
                };
                auth.memberships.entry(member).or_default().insert(group, Membership { admin, granted_by: actor.id });
            }
        }
        auth.persist()?;
        Ok(Outcome::command("CREATE ROLE"))
    }

    /// alter_role runs ALTER ROLE and ALTER USER.
    pub fn alter_role(&mut self, stmt: &AlterRoleStmt) -> Result<Outcome> {
        let actor = self.current_role()?;
        let name = stmt.role.as_ref().map(|r| self.role_name(r)).unwrap_or_default();
        let mut auth = self.auth()?;
        let mut role = auth.role(&name).cloned().ok_or_else(|| role_does_not_exist(&name))?;
        let before = role.clone();
        self.apply_role_options(&mut role, &stmt.options)?;
        let only_password = role.clone() == Role { password: role.password.clone(), ..before.clone() };
        if !actor.superuser {
            if role.superuser != before.superuser || before.superuser {
                return Err(insufficient("must be superuser to alter superuser roles or change superuser attribute"));
            }
            if !actor.create_role && !(only_password && actor.id == role.id) {
                return Err(insufficient("permission denied"));
            }
        }
        auth.roles.insert(role.id, role);
        auth.persist()?;
        Ok(Outcome::command("ALTER ROLE"))
    }

    /// rename_role runs ALTER ROLE ... RENAME TO.
    pub fn rename_role(&mut self, old: &str, new: &str) -> Result<Outcome> {
        let actor = self.current_role()?;
        if !actor.superuser && !actor.create_role {
            return Err(insufficient("permission denied to rename role"));
        }
        let mut auth = self.auth()?;
        if auth.role(new).is_some() {
            return Err(PgError::new(code::DUPLICATE_OBJECT, format!("role \"{new}\" already exists")));
        }
        let mut role = auth.role(old).cloned().ok_or_else(|| role_does_not_exist(old))?;
        if role.name == self.session.user {
            return Err(PgError::new(code::FEATURE_NOT_SUPPORTED, "session user cannot be renamed"));
        }
        role.name = new.to_string();
        auth.roles.insert(role.id, role);
        auth.persist()?;
        Ok(Outcome::command("ALTER ROLE"))
    }

    /// drop_role runs DROP ROLE, DROP USER, and DROP GROUP.
    pub fn drop_role(&mut self, stmt: &DropRoleStmt) -> Result<Outcome> {
        let actor = self.current_role()?;
        let names: Vec<String> = stmt
            .roles
            .iter()
            .filter_map(|n| match n.node.as_ref() {
                Some(NodeEnum::RoleSpec(spec)) => Some(self.role_name(spec)),
                _ => None,
            })
            .collect();
        let path = self.session.search_path();
        let auth_db = self.session.auth.clone();
        let mut auth = auth_db.lock().map_err(|_| PgError::internal("the auth lock was poisoned"))?;
        let mut dropped = Vec::new();
        for name in names {
            let Some(role) = auth.role(&name).cloned() else {
                if stmt.missing_ok {
                    self.session.notice(PgError::notice("00000", format!("role \"{name}\" does not exist, skipping")));
                    continue;
                }
                return Err(role_does_not_exist(&name));
            };
            if name == PUBLIC {
                return Err(PgError::new(code::RESERVED_NAME, "cannot use special role specifier in DROP ROLE"));
            }
            if name == self.session.role {
                return Err(PgError::new(code::OBJECT_IN_USE, "current user cannot be dropped"));
            }
            if name == self.session.user {
                return Err(PgError::new(code::OBJECT_IN_USE, "session user cannot be dropped"));
            }
            let dependents = Self::role_dependents(&auth, role.id, &path);
            if !dependents.is_empty() {
                return Err(PgError {
                    detail: Some(dependents.join("\n")),
                    ..PgError::new(
                        code::DEPENDENT_OBJECTS_STILL_EXIST,
                        format!("role \"{name}\" cannot be dropped because some objects depend on it"),
                    )
                });
            }
            if !actor.superuser && (role.superuser || !actor.create_role) {
                return Err(insufficient(if role.superuser {
                    "must be superuser to drop superusers"
                } else {
                    "permission denied to drop role"
                }));
            }
            dropped.push(role.id);
        }
        for id in dropped {
            auth.drop_role(id);
        }
        auth.persist()?;
        Ok(Outcome::command("DROP ROLE"))
    }

    /// grant_role runs GRANT and REVOKE of role memberships.
    pub fn grant_role(&mut self, stmt: &GrantRoleStmt) -> Result<Outcome> {
        let actor = self.current_role()?;
        let admin =
            stmt.opt.iter().any(|o| matches!(o.node.as_ref(), Some(NodeEnum::DefElem(d)) if d.defname == "admin"));
        let groups: Vec<String> = stmt
            .granted_roles
            .iter()
            .filter_map(|n| match n.node.as_ref() {
                Some(NodeEnum::AccessPriv(p)) => Some(p.priv_name.clone()),
                _ => None,
            })
            .collect();
        let members: Vec<String> = stmt
            .grantee_roles
            .iter()
            .filter_map(|n| match n.node.as_ref() {
                Some(NodeEnum::RoleSpec(spec)) => Some(self.role_name(spec)),
                _ => None,
            })
            .collect();
        let mut auth = self.auth()?;
        for group in &groups {
            let group_role = auth.role(group).cloned().ok_or_else(|| role_does_not_exist(group))?;
            let can_admin = actor.superuser
                || (actor.create_role && !group_role.superuser)
                || auth.memberships.get(&actor.id).and_then(|g| g.get(&group_role.id)).is_some_and(|m| m.admin);
            if !can_admin {
                return Err(PgError {
                    detail: Some(format!(
                        "Only roles with the ADMIN option on role \"{group}\" may {} this role.",
                        if stmt.is_grant { "grant" } else { "revoke" }
                    )),
                    ..insufficient(format!(
                        "permission denied to {} role \"{group}\"",
                        if stmt.is_grant { "grant" } else { "revoke" }
                    ))
                });
            }
            for member in &members {
                let member_role = auth.role(member).cloned().ok_or_else(|| role_does_not_exist(member))?;
                if stmt.is_grant {
                    if member_role.id == group_role.id || auth.groups(group_role.id, false).contains(&member_role.id) {
                        return Err(PgError::new(
                            code::INVALID_GRANT_OPERATION,
                            format!("role \"{group}\" is a member of role \"{member}\""),
                        ));
                    }
                    let entry = auth.memberships.entry(member_role.id).or_default();
                    let existing = entry.entry(group_role.id).or_insert(Membership { admin, granted_by: actor.id });
                    existing.admin |= admin;
                } else if let Some(groups) = auth.memberships.get_mut(&member_role.id) {
                    if admin {
                        if let Some(m) = groups.get_mut(&group_role.id) {
                            m.admin = false;
                        }
                    } else {
                        groups.remove(&group_role.id);
                    }
                    if groups.is_empty() {
                        auth.memberships.remove(&member_role.id);
                    }
                }
            }
        }
        auth.persist()?;
        Ok(Outcome::command(if stmt.is_grant { "GRANT ROLE" } else { "REVOKE ROLE" }))
    }

    /// grant runs GRANT and REVOKE of privileges on objects.
    pub fn grant(&mut self, stmt: &GrantStmt) -> Result<Outcome> {
        let actor = self.current_role()?;
        let object_type = ObjectType::try_from(stmt.objtype).unwrap_or(ObjectType::Undefined);
        let all_in_schema = GrantTargetType::try_from(stmt.targtype) == Ok(GrantTargetType::AclTargetAllInSchema);
        let (available, kind): (&[&str], &str) = match object_type {
            ObjectType::ObjectTable => (TABLE_PRIVILEGES, "relation"),
            ObjectType::ObjectSchema => (SCHEMA_PRIVILEGES, "schema"),
            ObjectType::ObjectSequence => (SEQUENCE_PRIVILEGES, "sequence"),
            ObjectType::ObjectDatabase => (DATABASE_PRIVILEGES, "database"),
            ObjectType::ObjectFunction | ObjectType::ObjectProcedure | ObjectType::ObjectRoutine => {
                (ROUTINE_PRIVILEGES, "routine")
            }
            _ => return Err(PgError::unsupported("GRANT on this kind of object")),
        };
        let privileges: Vec<&str> = if stmt.privileges.is_empty() {
            available.to_vec()
        } else {
            let mut out = Vec::new();
            for privilege in &stmt.privileges {
                let Some(NodeEnum::AccessPriv(p)) = privilege.node.as_ref() else { continue };
                if privilege_letter(&p.priv_name).is_none() {
                    return Err(PgError::new(
                        code::SYNTAX_ERROR,
                        format!("unrecognized privilege type \"{}\"", p.priv_name),
                    ));
                }
                let letter = privilege_letter(&p.priv_name).filter(|l| available.contains(l)).ok_or_else(|| {
                    PgError::new(
                        code::INVALID_GRANT_OPERATION,
                        format!("invalid privilege type {} for {kind}", p.priv_name.to_uppercase()),
                    )
                })?;
                out.push(letter);
            }
            out
        };
        let objects = self.grant_objects(stmt, object_type, all_in_schema)?;
        let grantees: Vec<String> = stmt
            .grantees
            .iter()
            .filter_map(|n| match n.node.as_ref() {
                Some(NodeEnum::RoleSpec(spec)) => Some(self.role_name(spec)),
                _ => None,
            })
            .collect();
        let mut auth = self.auth()?;
        let mut ids = Vec::new();
        for grantee in &grantees {
            ids.push(auth.role(grantee).ok_or_else(|| role_does_not_exist(grantee))?.id);
        }
        for object in &objects {
            for &privilege in &privileges {
                if !actor.superuser && auth.owner(object) != Some(actor.id) {
                    let can_grant = auth
                        .privileges
                        .get(&(actor.id, object.clone()))
                        .and_then(|p| p.get(privilege))
                        .is_some_and(|g| g.values().any(|&o| o));
                    if !can_grant {
                        return Err(insufficient(format!("permission denied for {}", object_name(object))));
                    }
                }
                for &id in &ids {
                    if stmt.is_grant {
                        auth.grant(id, object.clone(), privilege, actor.id, stmt.grant_option);
                    } else {
                        auth.revoke(id, object, privilege, stmt.grant_option);
                    }
                }
            }
        }
        auth.persist()?;
        Ok(Outcome::command(if stmt.is_grant { "GRANT" } else { "REVOKE" }))
    }

    /// grant_objects resolves the objects that GRANT or REVOKE names.
    fn grant_objects(&mut self, stmt: &GrantStmt, object_type: ObjectType, all_in_schema: bool) -> Result<Vec<Object>> {
        let mut objects = Vec::new();
        for object in &stmt.objects {
            match object.node.as_ref() {
                Some(NodeEnum::RangeVar(relation)) => {
                    if object_type == ObjectType::ObjectSequence {
                        let text = if relation.schemaname.is_empty() {
                            relation.relname.clone()
                        } else {
                            format!("{}.{}", relation.schemaname, relation.relname)
                        };
                        let sequence = self.resolve_sequence(&text, relation.location)?;
                        let (schema, name) = crate::sequences::schema_and_name(&sequence);
                        objects.push(Object::Sequence(schema, name));
                    } else if let Some((schema, _)) = self.find_view(&relation.schemaname, &relation.relname)? {
                        objects.push(Object::Table(schema, relation.relname.clone()));
                    } else {
                        let table = self.resolve_table(relation).map_err(|err| PgError { position: None, ..err })?;
                        objects.push(Object::Table(table.schema, table.name));
                    }
                }
                Some(NodeEnum::String(name)) if all_in_schema => {
                    if object_type == ObjectType::ObjectSequence {
                        for sequence in crate::sequences::all(self.db, &self.txn.root)? {
                            let (schema, sequence_name) = crate::sequences::schema_and_name(&sequence);
                            if schema == name.sval {
                                objects.push(Object::Sequence(schema, sequence_name));
                            }
                        }
                    } else {
                        let snapshot = self.snapshot()?;
                        for table in snapshot.tables.iter().filter(|t| t.schema == name.sval) {
                            objects.push(Object::Table(table.schema.clone(), table.name.clone()));
                        }
                        for view in snapshot.views.iter().filter(|v| v.schema == name.sval) {
                            objects.push(Object::Table(view.schema.clone(), view.name.clone()));
                        }
                    }
                }
                Some(NodeEnum::String(name)) if object_type == ObjectType::ObjectSchema => {
                    if !self.schema_names().contains(&name.sval) {
                        return Err(PgError::new(
                            code::INVALID_SCHEMA_NAME,
                            format!("schema \"{}\" does not exist", name.sval),
                        ));
                    }
                    objects.push(Object::Schema(name.sval.clone()));
                }
                Some(NodeEnum::String(name)) => {
                    if !self.session.database_names().contains(&name.sval) {
                        return Err(PgError::new(
                            code::INVALID_CATALOG_NAME,
                            format!("database \"{}\" does not exist", name.sval),
                        ));
                    }
                    objects.push(Object::Database(name.sval.clone()));
                }
                Some(NodeEnum::ObjectWithArgs(function)) => {
                    let names: Vec<&str> = function.objname.iter().filter_map(node_name).collect();
                    let (schema, name) = match names.as_slice() {
                        [schema, name] => (schema.to_string(), name.to_string()),
                        [name] => (self.creation_schema()?, name.to_string()),
                        _ => return Err(PgError::unsupported("this routine name")),
                    };
                    objects.push(Object::Routine(schema, name, String::new()));
                }
                _ => return Err(PgError::unsupported("GRANT on this object")),
            }
        }
        Ok(objects)
    }

    /// role_dependents describes the objects a role owns and the privileges granted to it, as DROP ROLE lists them.
    fn role_dependents(auth: &AuthDb, role: u64, path: &[String]) -> Vec<String> {
        let qualified = |schema: &str, name: &str| {
            if path.iter().any(|s| s == schema) { name.to_string() } else { format!("{schema}.{name}") }
        };
        let mut owned = Vec::new();
        let mut granted = Vec::new();
        for ((holder, object), privileges) in &auth.privileges {
            if *holder != role || matches!(object, Object::Table(_, n) if n.is_empty()) {
                continue;
            }
            let description = match object {
                Object::Database(name) => format!("database {name}"),
                Object::Schema(name) => format!("schema {name}"),
                Object::Table(schema, name) => format!("table {}", qualified(schema, name)),
                Object::Sequence(schema, name) => format!("sequence {}", qualified(schema, name)),
                Object::Routine(schema, name, args) => format!("function {}({args})", qualified(schema, name)),
            };
            if privileges.values().any(|g| g.contains_key(&role)) {
                owned.push(format!("owner of {description}"));
            } else {
                granted.push(format!("privileges for {description}"));
            }
        }
        owned.extend(granted);
        owned
    }

    /// set_role checks SET ROLE and RESET ROLE, where None returns to the session user.
    pub fn set_role(&mut self, name: Option<&str>) -> Result<()> {
        let name = match name {
            None | Some("none") => self.session.user.clone(),
            Some(name) => name.to_string(),
        };
        let auth = self.auth()?;
        let target = auth
            .role(&name)
            .cloned()
            .ok_or_else(|| PgError::new(code::INVALID_PARAMETER_VALUE, format!("role \"{name}\" does not exist")))?;
        let user = auth.role(&self.session.user).cloned().ok_or_else(|| role_does_not_exist(&self.session.user))?;
        if !user.superuser && target.id != user.id && !auth.groups(user.id, false).contains(&target.id) {
            return Err(insufficient(format!("permission denied to set role \"{name}\"")));
        }
        Ok(())
    }

    /// set_session_authorization checks SET SESSION AUTHORIZATION, which only a superuser that logged in may run.
    pub fn set_session_authorization(&mut self, name: &str) -> Result<()> {
        let auth = self.auth()?;
        let authenticated = auth.role(&self.session.authenticated).is_some_and(|r| r.superuser);
        if auth.role(name).is_none() {
            return Err(PgError::new(code::INVALID_PARAMETER_VALUE, format!("role \"{name}\" does not exist")));
        }
        if !authenticated && name != self.session.authenticated {
            return Err(insufficient(format!("permission denied to set session authorization \"{name}\"")));
        }
        Ok(())
    }

    /// is_superuser reports whether the current role is a superuser.
    fn is_superuser(&self) -> bool {
        self.auth().ok().and_then(|a| a.role(&self.session.role).map(|r| r.superuser)).unwrap_or(false)
    }

    /// require fails as Postgres does unless the current role holds a privilege on an object, checking the schema of
    /// tables and sequences for USAGE first.
    pub fn require(&mut self, object: &Object, privilege: &str, location: i32) -> Result<()> {
        if self.is_superuser() {
            return Ok(());
        }
        let auth = self.auth()?;
        let role = auth.role(&self.session.role).map_or(0, |r| r.id);
        let schema = match object {
            Object::Table(schema, _) | Object::Sequence(schema, _) => Some(schema.clone()),
            _ => None,
        };
        if let Some(schema) = schema
            && !self.holds_schema(&auth, role, &schema, "U")
        {
            return Err(PgError {
                position: position(location),
                ..insufficient(format!("permission denied for schema {schema}"))
            });
        }
        let held = match object {
            Object::Schema(schema) => self.holds_schema(&auth, role, schema, privilege),
            Object::Database(_) => auth.holds(role, object, privilege) || matches!(privilege, "c" | "T"),
            Object::Routine(..) => auth.holds(role, object, privilege) || privilege == "X",
            _ => auth.holds(role, object, privilege) || auth.owner(object) == Some(role),
        };
        if held {
            return Ok(());
        }
        let located = matches!(object, Object::Schema(_));
        Err(PgError {
            position: if located { position(location) } else { None },
            ..insufficient(format!("permission denied for {}", object_name(object)))
        })
    }

    /// holds_schema reports whether a role holds a privilege on a schema, where everyone may use pg_catalog,
    /// information_schema, and public.
    fn holds_schema(&self, auth: &AuthDb, role: u64, schema: &str, privilege: &str) -> bool {
        let object = Object::Schema(schema.to_string());
        (privilege == "U" && matches!(schema, "pg_catalog" | "information_schema" | "public" | "dolt"))
            || auth.holds(role, &object, privilege)
            || auth.owner(&object) == Some(role)
    }

    /// has_privilege reports whether a role holds any of the privileges on a schema or database, each with its grant
    /// option when asked, as has_schema_privilege and has_database_privilege do.
    pub fn has_privilege(&self, role: u64, object: &Object, privileges: &[(&str, bool)]) -> Result<bool> {
        let auth = self.auth()?;
        if auth.owner(object) == Some(role) || auth.roles.get(&role).is_some_and(|r| r.superuser) {
            return Ok(true);
        }
        Ok(privileges.iter().any(|&(privilege, option)| match object {
            _ if option => auth.holds_option(role, object, privilege),
            Object::Schema(schema) => self.holds_schema(&auth, role, schema, privilege),
            _ => auth.holds(role, object, privilege) || matches!(privilege, "c" | "T"),
        }))
    }

    /// require_create_db fails unless the current role may create databases.
    pub fn require_create_db(&self) -> Result<()> {
        let role = self.current_role()?;
        if role.superuser || role.create_db {
            return Ok(());
        }
        Err(insufficient("permission denied to create database"))
    }

    /// own records the current role as the owner of an object it created, by granting it every privilege on the
    /// object, unless it is a superuser, which holds them all anyway.
    pub fn own(&mut self, object: Object) -> Result<()> {
        let role = self.current_role()?;
        if role.superuser {
            return Ok(());
        }
        let privileges: &[&str] = match object {
            Object::Table(..) => TABLE_PRIVILEGES,
            Object::Sequence(..) => SEQUENCE_PRIVILEGES,
            Object::Schema(_) => SCHEMA_PRIVILEGES,
            Object::Database(_) => DATABASE_PRIVILEGES,
            Object::Routine(..) => ROUTINE_PRIVILEGES,
        };
        let mut auth = self.auth()?;
        for privilege in privileges {
            auth.grant(role.id, object.clone(), privilege, role.id, true);
        }
        auth.persist()
    }

    /// forget_object removes the privileges granted on a dropped object.
    pub fn forget_object(&mut self, object: &Object) -> Result<()> {
        let mut auth = self.auth()?;
        auth.drop_object(object);
        auth.persist()
    }

    /// require_owner fails as Postgres does unless the current role owns an object or is a superuser.
    pub fn require_owner(&mut self, object: &Object) -> Result<()> {
        if self.is_superuser() {
            return Ok(());
        }
        let auth = self.auth()?;
        let role = auth.role(&self.session.role).map_or(0, |r| r.id);
        if auth.owner(object) == Some(role) {
            return Ok(());
        }
        let kind = match object {
            Object::Table(..) => "table",
            Object::Sequence(..) => "sequence",
            Object::Schema(_) => "schema",
            Object::Database(_) => "database",
            Object::Routine(..) => "function",
        };
        let name = match object {
            Object::Table(_, n) | Object::Sequence(_, n) | Object::Routine(_, n, _) => n,
            Object::Schema(n) | Object::Database(n) => n,
        };
        Err(insufficient(format!("must be owner of {kind} {name}")))
    }
}

/// RESTRICTED_CATALOGS are the pg_catalog relations that Postgres does not let every role read.
const RESTRICTED_CATALOGS: [&str; 14] = [
    "pg_authid",
    "pg_backend_memory_contexts",
    "pg_config",
    "pg_file_settings",
    "pg_hba_file_rules",
    "pg_ident_file_mappings",
    "pg_largeobject",
    "pg_replication_origin_status",
    "pg_shadow",
    "pg_shmem_allocations",
    "pg_statistic",
    "pg_statistic_ext_data",
    "pg_subscription",
    "pg_user_mapping",
];

impl Ctx<'_> {
    /// require_sequence fails as Postgres does unless the current role holds one of the privileges on a sequence.
    pub fn require_sequence(&mut self, schema: &str, name: &str, privileges: &[&str]) -> Result<()> {
        let object = Object::Sequence(schema.to_string(), name.to_string());
        let mut last = None;
        for privilege in privileges {
            match self.require(&object, privilege, -1) {
                Ok(()) => return Ok(()),
                Err(err) => last = Some(err),
            }
        }
        last.map_or(Ok(()), Err)
    }

    /// require_view fails as Postgres does unless the current role holds a privilege on a view.
    pub fn require_view(&mut self, schema: &str, name: &str, privilege: &str, location: i32) -> Result<()> {
        self.require(&Object::Table(schema.to_string(), name.to_string()), privilege, location)
            .map_err(|err| PgError { message: err.message.replacen("for table ", "for view ", 1), ..err })
    }

    /// resolve_target resolves the table that INSERT, UPDATE, or DELETE changes, checking the privilege on a view of
    /// that name first, as Postgres does before it rewrites a change of a view, and creating the docs table on the
    /// first change of it and a schema's dolt_ignore table on the first INSERT into it, as Dolt does.
    pub fn resolve_target(&mut self, relation: &pg_query::protobuf::RangeVar, privilege: &str) -> Result<TableDef> {
        match self.resolve_table(relation) {
            Ok(table) => Ok(table),
            Err(err) => match self.find_view(&relation.schemaname, &relation.relname)? {
                Some((schema, _)) => {
                    self.require_view(&schema, &relation.relname, privilege, -1)?;
                    Err(PgError::unsupported("changing the rows of a view"))
                }
                None if crate::dolt::docs::is_docs(&relation.schemaname, &relation.relname) => {
                    crate::dolt::docs::table(self)
                }
                None if relation.relname == crate::dolt::ignore::TABLE && privilege == "a" => {
                    let schema = match relation.schemaname.as_str() {
                        "" => self.creation_schema()?,
                        schema if self.txn.root.schemas.iter().any(|s| s == schema.as_bytes()) => schema.to_string(),
                        _ => return Err(err),
                    };
                    crate::dolt::ignore::create(self, &schema)
                }
                None => Err(err),
            },
        }
    }

    /// require_catalog fails as Postgres does when a role that is not a superuser reads a restricted catalog.
    pub fn require_catalog(&self, name: &str) -> Result<()> {
        if RESTRICTED_CATALOGS.contains(&name) && !self.is_superuser() {
            return Err(insufficient(format!("permission denied for table {name}")));
        }
        Ok(())
    }

    /// owner_name returns the name of the role that owns an object, which is the superuser when no other role does.
    pub fn owner_name(&self, object: &Object) -> Result<String> {
        let auth = self.auth()?;
        Ok(auth
            .owner(object)
            .and_then(|id| auth.roles.get(&id))
            .map_or(self.session.superuser.clone(), |r| r.name.clone()))
    }
}

/// object_name names an object as Postgres' permission errors do.
fn object_name(object: &Object) -> String {
    match object {
        Object::Table(_, name) => format!("table {name}"),
        Object::Sequence(_, name) => format!("sequence {name}"),
        Object::Schema(name) => format!("schema {name}"),
        Object::Database(name) => format!("database {name}"),
        Object::Routine(_, name, _) => format!("function {name}"),
    }
}
