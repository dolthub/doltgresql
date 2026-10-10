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

//! Roles and privileges, which every database of a server shares, kept in the auth file in Go's format.

use std::collections::{BTreeMap, BTreeSet};
use std::path::{Path, PathBuf};

use hmac::{Hmac, Mac};
use objects::codec::{Reader, Writer};
use rand::RngCore;
use sha2::{Digest, Sha256};

use crate::error::{PgError, Result, code};

/// ITERATIONS is the PBKDF2 iteration count of stored passwords, Postgres' default.
pub const ITERATIONS: u32 = 4096;

/// Password is a SCRAM-SHA-256 password as the server keeps it: its salt and iteration count with the derived keys.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Password {
    pub iterations: u32,
    pub salt: Vec<u8>,
    pub stored_key: Vec<u8>,
    pub server_key: Vec<u8>,
}

/// hmac returns the HMAC-SHA-256 of the message with the key.
fn hmac(key: &[u8], message: &[u8]) -> Vec<u8> {
    let mut mac = Hmac::<Sha256>::new_from_slice(key).expect("HMAC takes keys of any length");
    mac.update(message);
    mac.finalize().into_bytes().to_vec()
}

impl Password {
    /// new derives the stored form of a password with a random salt.
    pub fn new(password: &str) -> Password {
        let mut salt = vec![0; 16];
        rand::thread_rng().fill_bytes(&mut salt);
        Password::with_salt(password, salt, ITERATIONS)
    }

    /// with_salt derives the stored form of a password with the salt and iteration count.
    pub fn with_salt(password: &str, salt: Vec<u8>, iterations: u32) -> Password {
        let mut salted = [0; 32];
        pbkdf2::pbkdf2_hmac::<Sha256>(password.as_bytes(), &salt, iterations, &mut salted);
        let client_key = hmac(&salted, b"Client Key");
        let stored_key = Sha256::digest(client_key).to_vec();
        let server_key = hmac(&salted, b"Server Key");
        Password { iterations, salt, stored_key, server_key }
    }

    /// text returns the password as pg_authid shows it.
    pub fn text(&self) -> String {
        use base64::Engine;
        let b64 = |b: &[u8]| base64::engine::general_purpose::STANDARD.encode(b);
        format!(
            "SCRAM-SHA-256${}:{}${}:{}",
            self.iterations,
            b64(&self.salt),
            b64(&self.stored_key),
            b64(&self.server_key)
        )
    }
}

/// Role is a role or user.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Role {
    pub id: u64,
    pub name: String,
    pub superuser: bool,
    pub inherit: bool,
    pub create_role: bool,
    pub create_db: bool,
    pub login: bool,
    pub replication: bool,
    pub bypass_rls: bool,
    pub connection_limit: i32,
    pub password: Option<Password>,
    /// The time the password expires, in microseconds since 1970.
    pub valid_until: Option<i64>,
}

impl Role {
    /// new returns a role with Postgres' defaults.
    pub fn new(id: u64, name: &str) -> Role {
        Role {
            id,
            name: name.to_string(),
            superuser: false,
            inherit: true,
            create_role: false,
            create_db: false,
            login: false,
            replication: false,
            bypass_rls: false,
            connection_limit: -1,
            password: None,
            valid_until: None,
        }
    }
}

/// Privileges are the privileges a role holds on an object: for each privilege letter, the roles that granted it and
/// whether each grant carries the grant option.
pub type Privileges = BTreeMap<String, BTreeMap<u64, bool>>;

/// Object is something privileges apply to, named as Go's auth file keys it.
#[derive(Clone, Debug, PartialEq, Eq, PartialOrd, Ord)]
pub enum Object {
    Database(String),
    Schema(String),
    /// A table or view by schema and name, where an empty name is every table of the schema.
    Table(String, String),
    Sequence(String, String),
    /// A function or procedure by schema, name, and argument types.
    Routine(String, String, String),
}

/// Membership is a role's membership in another: whether it can administer it and who granted it.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Membership {
    pub admin: bool,
    pub granted_by: u64,
}

/// AuthDb holds every role, the privileges granted to them, and their memberships.
#[derive(Clone, Debug, Default)]
pub struct AuthDb {
    pub roles: BTreeMap<u64, Role>,
    pub privileges: BTreeMap<(u64, Object), Privileges>,
    /// Each member's groups.
    pub memberships: BTreeMap<u64, BTreeMap<u64, Membership>>,
    /// Where the auth file is, which is None for a database kept in memory.
    path: Option<PathBuf>,
}

/// PUBLIC is the name of the role every role is a member of.
pub const PUBLIC: &str = "public";

/// corrupt returns the error for an auth file that cannot be read.
fn corrupt(err: impl std::fmt::Display) -> PgError {
    PgError::internal(format!("invalid auth database format: {err}"))
}

/// read_privileges reads one privilege section of the auth file: the keys the read function returns, and each key's
/// privileges.
fn read_privileges(
    reader: &mut Reader<'_>,
    db: &mut AuthDb,
    read_key: &dyn Fn(&mut Reader<'_>) -> store::Result<Object>,
) -> store::Result<()> {
    for _ in 0..reader.uint64()? {
        let role = reader.uint64()?;
        let object = read_key(reader)?;
        let mut privileges = Privileges::new();
        for _ in 0..reader.uint64()? {
            let privilege = String::from_utf8_lossy(&reader.string()?).into_owned();
            let mut grants = BTreeMap::new();
            for _ in 0..reader.uint32()? {
                let granted_by = reader.uint64()?;
                grants.insert(granted_by, reader.bool()?);
            }
            privileges.insert(privilege, grants);
        }
        db.privileges.insert((role, object), privileges);
    }
    Ok(())
}

/// lossy reads a string field.
fn lossy(reader: &mut Reader<'_>) -> store::Result<String> {
    Ok(String::from_utf8_lossy(&reader.string()?).into_owned())
}

impl AuthDb {
    /// open reads the auth file, creating it with the public role and the superuser when it does not exist, as Go
    /// does.
    pub fn open(path: &Path, superuser: &str, password: &str) -> Result<AuthDb> {
        match std::fs::read(path) {
            Ok(data) => {
                let mut db = AuthDb::deserialize(&data)?;
                db.path = Some(path.to_path_buf());
                Ok(db)
            }
            Err(err) if err.kind() == std::io::ErrorKind::NotFound => {
                let mut db = AuthDb { path: Some(path.to_path_buf()), ..AuthDb::default() };
                db.add_defaults(superuser, password);
                if let Some(dir) = path.parent().filter(|d| !d.as_os_str().is_empty()) {
                    std::fs::create_dir_all(dir).map_err(PgError::internal)?;
                }
                db.persist()?;
                Ok(db)
            }
            Err(err) => Err(PgError::internal(err)),
        }
    }

    /// in_memory returns a database with the defaults that is never written.
    pub fn in_memory(superuser: &str, password: &str) -> AuthDb {
        let mut db = AuthDb::default();
        db.add_defaults(superuser, password);
        db
    }

    /// add_defaults adds the public role and the superuser.
    fn add_defaults(&mut self, superuser: &str, password: &str) {
        let public = Role::new(self.next_id(), PUBLIC);
        self.roles.insert(public.id, public);
        let mut user = Role::new(self.next_id(), superuser);
        user.superuser = true;
        user.create_role = true;
        user.create_db = true;
        user.login = true;
        user.replication = true;
        user.bypass_rls = true;
        user.password = Some(Password::new(password));
        self.roles.insert(user.id, user);
    }

    /// next_id returns an ID that no role has.
    pub fn next_id(&self) -> u64 {
        self.roles.keys().max().map_or(1, |m| m + 1)
    }

    /// persist writes the auth file.
    pub fn persist(&self) -> Result<()> {
        match &self.path {
            Some(path) => std::fs::write(path, self.serialize()).map_err(PgError::internal),
            None => Ok(()),
        }
    }

    /// replace replaces the roles and privileges with serialized ones, as a standby takes its primary's, and persists
    /// them.
    pub fn replace(&mut self, data: &[u8]) -> Result<()> {
        let fresh = AuthDb::deserialize(data)?;
        *self = AuthDb { path: self.path.take(), ..fresh };
        self.persist()
    }

    /// deserialize reads an auth file, dropping records about roles that no longer exist.
    pub fn deserialize(data: &[u8]) -> Result<AuthDb> {
        let mut reader = Reader::new(data);
        let version = reader.uint32().map_err(corrupt)?;
        if version > 1 {
            return Err(PgError::internal(format!(
                "Authorization database format {version} is not supported, please upgrade Doltgres"
            )));
        }
        let mut db = AuthDb::default();
        (|| -> store::Result<()> {
            for _ in 0..reader.uint32()? {
                let name = lossy(&mut reader)?;
                let mut role = Role::new(0, &name);
                role.superuser = reader.bool()?;
                role.inherit = reader.bool()?;
                role.create_role = reader.bool()?;
                role.create_db = reader.bool()?;
                role.login = reader.bool()?;
                role.replication = reader.bool()?;
                role.bypass_rls = reader.bool()?;
                role.connection_limit = reader.int32()?;
                if reader.bool()? {
                    role.password = Some(Password {
                        iterations: reader.uint32()?,
                        salt: reader.bytes()?,
                        stored_key: reader.bytes()?,
                        server_key: reader.bytes()?,
                    });
                }
                if reader.bool()? {
                    role.valid_until = Some(reader.int64()?);
                }
                role.id = reader.uint64()?;
                db.roles.insert(role.id, role);
            }
            read_privileges(&mut reader, &mut db, &|r| Ok(Object::Database(lossy(r)?)))?;
            read_privileges(&mut reader, &mut db, &|r| Ok(Object::Schema(lossy(r)?)))?;
            read_privileges(&mut reader, &mut db, &|r| {
                let name = lossy(r)?;
                Ok(Object::Table(lossy(r)?, name))
            })?;
            read_privileges(&mut reader, &mut db, &|r| Ok(Object::Sequence(lossy(r)?, lossy(r)?)))?;
            read_privileges(&mut reader, &mut db, &|r| Ok(Object::Routine(lossy(r)?, lossy(r)?, lossy(r)?)))?;
            for _ in 0..reader.uint64()? {
                for _ in 0..reader.uint64()? {
                    let member = reader.uint64()?;
                    let group = reader.uint64()?;
                    let admin = reader.bool()?;
                    let granted_by = reader.uint64()?;
                    db.memberships.entry(member).or_default().insert(group, Membership { admin, granted_by });
                }
            }
            Ok(())
        })()
        .map_err(corrupt)?;
        db.remove_invalid_references();
        Ok(db)
    }

    /// remove_invalid_references drops privileges and memberships that refer to roles that do not exist.
    fn remove_invalid_references(&mut self) {
        let roles: BTreeSet<u64> = self.roles.keys().copied().collect();
        self.privileges.retain(|(role, _), privileges| {
            privileges.retain(|_, grants| {
                grants.retain(|by, _| roles.contains(by));
                !grants.is_empty()
            });
            roles.contains(role) && !privileges.is_empty()
        });
        self.memberships.retain(|member, groups| {
            groups.retain(|group, m| roles.contains(group) && roles.contains(&m.granted_by));
            roles.contains(member) && !groups.is_empty()
        });
    }

    /// serialize writes the auth file in Go's version 1 format.
    pub fn serialize(&self) -> Vec<u8> {
        let mut w = Writer::new();
        w.uint32(1);
        w.uint32(self.roles.len() as u32);
        for role in self.roles.values() {
            w.string(role.name.as_bytes());
            for flag in [
                role.superuser,
                role.inherit,
                role.create_role,
                role.create_db,
                role.login,
                role.replication,
                role.bypass_rls,
            ] {
                w.bool(flag);
            }
            w.int32(role.connection_limit);
            w.bool(role.password.is_some());
            if let Some(p) = &role.password {
                w.uint32(p.iterations);
                w.string(&p.salt);
                w.string(&p.stored_key);
                w.string(&p.server_key);
            }
            w.bool(role.valid_until.is_some());
            if let Some(t) = role.valid_until {
                w.int64(t);
            }
            w.uint64(role.id);
        }
        type KeyWriter = fn(&mut Writer, &Object) -> bool;
        let sections: [KeyWriter; 5] = [
            |w, o| match o {
                Object::Database(n) => {
                    w.string(n.as_bytes());
                    true
                }
                _ => false,
            },
            |w, o| match o {
                Object::Schema(n) => {
                    w.string(n.as_bytes());
                    true
                }
                _ => false,
            },
            |w, o| match o {
                Object::Table(s, n) => {
                    w.string(n.as_bytes());
                    w.string(s.as_bytes());
                    true
                }
                _ => false,
            },
            |w, o| match o {
                Object::Sequence(s, n) => {
                    w.string(s.as_bytes());
                    w.string(n.as_bytes());
                    true
                }
                _ => false,
            },
            |w, o| match o {
                Object::Routine(s, n, a) => {
                    w.string(s.as_bytes());
                    w.string(n.as_bytes());
                    w.string(a.as_bytes());
                    true
                }
                _ => false,
            },
        ];
        for section in sections {
            let entries: Vec<_> = self.privileges.iter().filter(|((_, o), _)| section(&mut Writer::new(), o)).collect();
            w.uint64(entries.len() as u64);
            for ((role, object), privileges) in entries {
                w.uint64(*role);
                section(&mut w, object);
                w.uint64(privileges.len() as u64);
                for (privilege, grants) in privileges {
                    w.string(privilege.as_bytes());
                    w.uint32(grants.len() as u32);
                    for (by, option) in grants {
                        w.uint64(*by);
                        w.bool(*option);
                    }
                }
            }
        }
        w.uint64(self.memberships.len() as u64);
        for (member, groups) in &self.memberships {
            w.uint64(groups.len() as u64);
            for (group, m) in groups {
                w.uint64(*member);
                w.uint64(*group);
                w.bool(m.admin);
                w.uint64(m.granted_by);
            }
        }
        w.data()
    }

    /// role returns a role by name.
    pub fn role(&self, name: &str) -> Option<&Role> {
        self.roles.values().find(|r| r.name == name)
    }

    /// public_id returns the ID of the public role.
    pub fn public_id(&self) -> u64 {
        self.role(PUBLIC).map_or(0, |r| r.id)
    }

    /// groups returns every role that a role is a member of, directly or through other memberships, following only
    /// members that inherit privileges when asked.
    pub fn groups(&self, role: u64, inheriting: bool) -> Vec<u64> {
        let mut found = Vec::new();
        let mut pending = vec![role];
        while let Some(member) = pending.pop() {
            if inheriting && !self.roles.get(&member).is_some_and(|r| r.inherit) {
                continue;
            }
            for &group in self.memberships.get(&member).map(|g| g.keys()).into_iter().flatten() {
                if group != role && !found.contains(&group) {
                    found.push(group);
                    pending.push(group);
                }
            }
        }
        found
    }

    /// is_admin_of reports whether a role may administer another: a superuser always may, no role may administer
    /// itself, and otherwise the role or one it is a member of must hold the other WITH ADMIN OPTION, as Postgres'
    /// is_admin_of_role decides.
    pub fn is_admin_of(&self, member: u64, role: u64) -> bool {
        if self.roles.get(&member).is_some_and(|r| r.superuser) {
            return true;
        }
        if member == role {
            return false;
        }
        std::iter::once(member)
            .chain(self.groups(member, false))
            .any(|m| self.memberships.get(&m).and_then(|g| g.get(&role)).is_some_and(|m| m.admin))
    }

    /// bootstrap_id returns the ID of the superuser that the database was created with, which Postgres names as the
    /// grantor of the memberships it grants on its own.
    pub fn bootstrap_id(&self) -> u64 {
        self.roles.values().filter(|r| r.superuser).map(|r| r.id).min().unwrap_or(0)
    }

    /// holds reports whether a role holds a privilege on an object itself, through the schema-wide grant of a table,
    /// or through a role it inherits from, where superusers hold every privilege.
    pub fn holds(&self, role: u64, object: &Object, privilege: &str) -> bool {
        if self.roles.get(&role).is_some_and(|r| r.superuser) {
            return true;
        }
        let public = self.public_id();
        let mut candidates = vec![role, public];
        candidates.extend(self.groups(role, true));
        candidates.iter().any(|&r| {
            let direct = self
                .privileges
                .get(&(r, object.clone()))
                .is_some_and(|p| p.get(privilege).is_some_and(|g| !g.is_empty()));
            let schema_wide = match object {
                Object::Table(schema, name) if !name.is_empty() => self
                    .privileges
                    .get(&(r, Object::Table(schema.clone(), String::new())))
                    .is_some_and(|p| p.get(privilege).is_some_and(|g| !g.is_empty())),
                _ => false,
            };
            direct || schema_wide
        })
    }

    /// holds_option reports whether a role holds a privilege on an object with the option to grant it, itself or
    /// through a role it inherits from, where superusers hold every privilege with it.
    pub fn holds_option(&self, role: u64, object: &Object, privilege: &str) -> bool {
        if self.roles.get(&role).is_some_and(|r| r.superuser) {
            return true;
        }
        let mut candidates = vec![role, self.public_id()];
        candidates.extend(self.groups(role, true));
        candidates.iter().any(|&r| {
            self.privileges
                .get(&(r, object.clone()))
                .and_then(|p| p.get(privilege))
                .is_some_and(|grants| grants.values().any(|&option| option))
        })
    }

    /// grant gives a role a privilege on an object, as granted by another role.
    pub fn grant(&mut self, role: u64, object: Object, privilege: &str, granted_by: u64, option: bool) {
        let grants = self.privileges.entry((role, object)).or_default().entry(privilege.to_string()).or_default();
        let existing = grants.entry(granted_by).or_insert(option);
        *existing |= option;
    }

    /// revoke takes a privilege on an object away from a role, or only its grant option.
    pub fn revoke(&mut self, role: u64, object: &Object, privilege: &str, grant_option_only: bool) {
        let key = (role, object.clone());
        if let Some(privileges) = self.privileges.get_mut(&key) {
            if grant_option_only {
                if let Some(grants) = privileges.get_mut(privilege) {
                    grants.values_mut().for_each(|o| *o = false);
                }
            } else {
                privileges.remove(privilege);
            }
            if privileges.is_empty() {
                self.privileges.remove(&key);
            }
        }
    }

    /// drop_role removes a role with every privilege granted to or by it and its memberships.
    pub fn drop_role(&mut self, id: u64) {
        self.roles.remove(&id);
        self.remove_invalid_references();
    }

    /// owner returns the role that owns an object: the one that granted itself privileges on it when it created it,
    /// or the superuser it was created by otherwise.
    pub fn owner(&self, object: &Object) -> Option<u64> {
        self.privileges
            .iter()
            .find(|((role, o), privileges)| o == object && privileges.values().any(|g| g.contains_key(role)))
            .map(|((role, _), _)| *role)
    }

    /// rename_object follows a rename of an object in every privilege granted on it.
    pub fn rename_object(&mut self, old: &Object, new: &Object) {
        let renamed: Vec<_> = self.privileges.keys().filter(|(_, o)| o == old).cloned().collect();
        for key in renamed {
            if let Some(privileges) = self.privileges.remove(&key) {
                self.privileges.insert((key.0, new.clone()), privileges);
            }
        }
    }

    /// drop_object removes every privilege granted on an object.
    pub fn drop_object(&mut self, object: &Object) {
        self.privileges.retain(|(_, o), _| o != object);
    }
}

/// insufficient returns Postgres' error for a missing privilege.
pub fn insufficient(message: impl Into<String>) -> PgError {
    PgError::new(code::INSUFFICIENT_PRIVILEGE, message)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn go_auth_files_read_and_round_trip() {
        let db = AuthDb::deserialize(include_bytes!("../tests/fixtures/go_auth.db")).unwrap();
        let names: BTreeSet<&str> = db.roles.values().map(|r| r.name.as_str()).collect();
        assert_eq!(names, BTreeSet::from(["postgres", "public", "u1"]));
        let user = db.role("u1").unwrap();
        assert!(user.login && !user.superuser && user.inherit);
        let stored = user.password.clone().unwrap();
        assert_eq!(Password::with_salt("p1", stored.salt.clone(), stored.iterations), stored);
        let again = AuthDb::deserialize(&db.serialize()).unwrap();
        assert_eq!(again.roles, db.roles);
    }
}
