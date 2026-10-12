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

//! Catalog rows for installed extensions and the access methods and operator classes they provide.

use crate::catalog::id::{self, SECTION_FUNCTION, SECTION_TYPE};
use crate::catalog::oids;
use crate::error::Result;
use crate::pgcatalog::rows::SUPERUSER;
use crate::pgcatalog::snapshot::namespace_oid;
use crate::pgcatalog::{Rows, boolean, oid, text};
use crate::query::Ctx;
use crate::types::{Reg, Value};

/// The ID sections of access methods, operator classes, and operator families.
const SECTION_ACCESS_METHOD: u8 = 1;
const SECTION_OPERATOR_CLASS: u8 = 21;
const SECTION_OPERATOR_FAMILY: u8 = 22;

/// access_method_oid returns the OID of an access method that an extension provides.
pub fn access_method_oid(name: &str) -> u32 {
    oids::oid(&id::new(SECTION_ACCESS_METHOD, &[name]))
}

impl Ctx<'_> {
    /// installed returns each installed extension's definition with the schema it was installed in.
    fn installed(&mut self) -> Result<Vec<(&'static crate::extensions::Extension, objects::Extension, String)>> {
        let mut installed = Vec::new();
        for extension in self.installed_extensions()? {
            let name = id::segments(&extension.ext_name).into_iter().next().unwrap_or_default();
            let schema = id::segments(&extension.namespace).into_iter().next().unwrap_or_default();
            if let Some(definition) = crate::extensions::get(&name) {
                installed.push((definition, extension, schema));
            }
        }
        Ok(installed)
    }

    /// pg_extension lists the installed extensions after the ones Postgres always has.
    pub(super) fn pg_extension(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        for (definition, extension, schema) in self.installed()? {
            rows.push(vec![
                ("oid", oid(oids::oid(&extension.ext_name))),
                ("extname", text(definition.name)),
                ("extowner", oid(SUPERUSER)),
                ("extnamespace", oid(namespace_oid(&schema))),
                ("extrelocatable", boolean(extension.relocatable)),
                ("extversion", text(String::from_utf8_lossy(&extension.version).into_owned())),
            ]);
        }
        Ok(())
    }

    /// pg_available_extensions lists the extensions that Doltgres can install, which are PL/pgSQL and the emulated
    /// ones, in pg_available_extensions and pg_available_extension_versions.
    pub(super) fn pg_available_extensions(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        let installed = self.installed()?;
        rows.push(vec![
            ("name", text("plpgsql")),
            ("default_version", text("1.0")),
            ("version", text("1.0")),
            ("installed_version", text("1.0")),
            ("installed", boolean(true)),
            ("superuser", boolean(true)),
            ("trusted", boolean(true)),
            ("relocatable", boolean(false)),
            ("schema", text("pg_catalog")),
            ("comment", text("PL/pgSQL procedural language")),
        ]);
        for extension in crate::extensions::all() {
            let version = installed
                .iter()
                .find(|(d, ..)| d.name == extension.name)
                .map(|(_, e, _)| String::from_utf8_lossy(&e.version).into_owned());
            let control = &extension.control;
            rows.push(vec![
                ("name", text(extension.name)),
                ("default_version", text(control.default_version)),
                ("version", text(control.default_version)),
                ("installed_version", version.clone().map_or(Value::Null, text)),
                ("installed", boolean(version.as_deref() == Some(control.default_version))),
                ("superuser", boolean(control.superuser)),
                ("trusted", boolean(control.trusted)),
                ("relocatable", boolean(control.relocatable)),
                ("comment", text(control.comment)),
            ]);
        }
        Ok(())
    }

    /// pg_am lists the access methods of the installed extensions after Postgres' own.
    pub(super) fn pg_am(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        for (definition, _, schema) in self.installed()? {
            for method in &definition.access_methods {
                let internal =
                    String::from_utf8_lossy(&id::new(SECTION_TYPE, &["pg_catalog", "internal"])).into_owned();
                let handler = id::new(SECTION_FUNCTION, &[&schema, method.handler, &internal]);
                rows.push(vec![
                    ("oid", oid(access_method_oid(method.name))),
                    ("amname", text(method.name)),
                    (
                        "amhandler",
                        Value::Reg(Box::new(Reg {
                            type_oid: crate::oid::REGPROC,
                            oid: oids::oid(&handler),
                            name: method.handler.to_string(),
                        })),
                    ),
                    ("amtype", text("i")),
                ]);
            }
        }
        Ok(())
    }

    /// pg_opclass lists the operator classes of the installed extensions after Postgres' own.
    pub(super) fn pg_opclass(&mut self, rows: &mut Rows<'_>) -> Result<()> {
        for (definition, _, schema) in self.installed()? {
            for class in &definition.operator_classes {
                let input = match crate::catalog::builtin_type_named(class.type_name) {
                    Some(builtin) => builtin.oid,
                    None => oids::oid(&id::new(SECTION_TYPE, &[&schema, class.type_name])),
                };
                for method in &class.access_methods {
                    rows.push(vec![
                        ("oid", oid(oids::oid(&id::new(SECTION_OPERATOR_CLASS, &[method, &class.name])))),
                        ("opcmethod", oid(access_method_oid(method))),
                        ("opcname", text(class.name.clone())),
                        ("opcnamespace", oid(namespace_oid(&schema))),
                        ("opcowner", oid(SUPERUSER)),
                        ("opcfamily", oid(oids::oid(&id::new(SECTION_OPERATOR_FAMILY, &[method, &class.name])))),
                        ("opcintype", oid(input)),
                        ("opcdefault", boolean(class.default_for.contains(method))),
                        ("opckeytype", oid(0)),
                    ]);
                }
            }
        }
        Ok(())
    }
}
