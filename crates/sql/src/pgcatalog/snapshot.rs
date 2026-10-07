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

//! The user objects of the working root value, which the system catalogs describe.

use objects::Sequence;

use crate::catalog::table::TableDef;
use crate::catalog::{id, oids};
use crate::error::Result;
use crate::foreign::ForeignKeyDef;
use crate::query::Ctx;

/// ViewDef is a view: its schema, name, and CREATE VIEW statement.
pub struct ViewDef {
    pub schema: String,
    pub name: String,
    pub statement: String,
}

/// Snapshot is every user object of the working root value.
pub struct Snapshot {
    /// The schemas, in name order.
    pub schemas: Vec<String>,
    /// The tables, without Dolt's own tables, in schema and name order.
    pub tables: Vec<TableDef>,
    pub views: Vec<ViewDef>,
    pub sequences: Vec<Sequence>,
    pub foreign_keys: Vec<ForeignKeyDef>,
}

impl Snapshot {
    /// table returns a table by schema and name.
    pub fn table(&self, schema: &str, name: &str) -> Option<&TableDef> {
        self.tables.iter().find(|t| t.schema == schema && t.name == name)
    }
}

/// table_oid returns the OID of a table.
pub fn table_oid(schema: &str, name: &str) -> u32 {
    oids::oid(&id::new(id::SECTION_TABLE, &[schema, name]))
}

/// view_oid returns the OID of a view.
pub fn view_oid(schema: &str, name: &str) -> u32 {
    oids::oid(&id::new(id::SECTION_VIEW, &[schema, name]))
}

/// sequence_oid returns the OID of a sequence.
pub fn sequence_oid(schema: &str, name: &str) -> u32 {
    oids::oid(&id::new(id::SECTION_SEQUENCE, &[schema, name]))
}

/// index_oid returns the OID of an index, where the primary key's index is named `<table>_pkey`.
pub fn index_oid(schema: &str, table: &str, index: &str) -> u32 {
    oids::oid(&id::new(id::SECTION_INDEX, &[schema, table, index]))
}

/// database_oid returns the OID of a database.
pub fn database_oid(name: &str) -> u32 {
    oids::oid(&id::new(id::SECTION_DATABASE, &[name]))
}

/// namespace_oid returns the OID of a schema.
pub fn namespace_oid(schema: &str) -> u32 {
    oids::oid(&id::new(id::SECTION_NAMESPACE, &[schema]))
}

/// constraint_oid returns the OID of a constraint in its ID section.
pub fn constraint_oid(section: u8, schema: &str, table: &str, name: &str) -> u32 {
    oids::oid(&id::new(section, &[schema, table, name]))
}

impl Ctx<'_> {
    /// snapshot reads every user object of the working root value.
    pub fn snapshot(&mut self) -> Result<Snapshot> {
        let mut schemas: Vec<String> =
            self.txn.root.schemas.iter().map(|s| String::from_utf8_lossy(s).into_owned()).collect();
        schemas.sort();
        let mut tables = Vec::new();
        for (key, address) in self.txn.root.tables(self.db)? {
            let text = String::from_utf8_lossy(&key).into_owned();
            let mut parts = text.splitn(3, '\0').skip(1);
            let (Some(schema), Some(name)) = (parts.next(), parts.next()) else { continue };
            if name.starts_with("dolt_") || schema == "dolt" {
                continue;
            }
            tables.push(TableDef::load(self.db, schema, name, address)?);
        }
        tables.sort_by(|a, b| (&a.schema, &a.name).cmp(&(&b.schema, &b.name)));
        let mut views = Vec::new();
        for schema in &schemas {
            for (name, statement) in self.views(schema)? {
                views.push(ViewDef { schema: schema.clone(), name, statement });
            }
        }
        let sequences = crate::sequences::all(self.db, &self.txn.root)?;
        let foreign_keys = self.foreign_keys()?;
        Ok(Snapshot { schemas, tables, views, sequences, foreign_keys })
    }
}
