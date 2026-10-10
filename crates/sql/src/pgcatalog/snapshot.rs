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
#[derive(Clone)]
pub struct ViewDef {
    pub schema: String,
    pub name: String,
    pub statement: String,
}

/// Snapshot is every user object of the working root value.
pub struct Snapshot {
    /// The schemas, in name order.
    pub schemas: Vec<String>,
    /// The tables, without Dolt's own tables and with the tables that dolt_nonlocal_tables names, in schema and name
    /// order.
    pub tables: Vec<TableDef>,
    pub views: Vec<ViewDef>,
    pub sequences: Vec<Sequence>,
    pub foreign_keys: Vec<ForeignKeyDef>,
    /// Dolt's own tables with their indexes, which the catalogs list when dolt_show_system_tables is on.
    pub system: Vec<(TableDef, Vec<crate::pgcatalog::rows::TableIndex>)>,
}

/// CatalogCache holds what a statement has read of the catalogs, for the root value it read them from.
pub struct CatalogCache {
    pub root: doltdb::root::Root,
    pub snapshot: std::sync::Arc<Snapshot>,
    /// Whether the snapshot lists Dolt's own tables, as dolt_show_system_tables asked when it was read.
    pub system: bool,
    /// The relations that regclass can name, once a lookup asks for them.
    pub relations: Option<std::sync::Arc<Vec<crate::pgcatalog::reg::Relation>>>,
}

impl Snapshot {
    /// table returns a table by schema and name.
    pub fn table(&self, schema: &str, name: &str) -> Option<&TableDef> {
        self.tables.iter().find(|t| t.schema == schema && t.name == name)
    }

    /// listed returns the tables that the catalogs list, with their indexes: the user tables, then Dolt's own tables
    /// when dolt_show_system_tables is on.
    pub fn listed(&self) -> Vec<(&TableDef, Vec<crate::pgcatalog::rows::TableIndex>)> {
        let user = self.tables.iter().map(|t| (t, crate::pgcatalog::rows::table_indexes(t)));
        user.chain(self.system.iter().map(|(t, indexes)| (t, indexes.clone()))).collect()
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

/// index_oid returns the OID of an index, where a primary key's index is named `PRIMARY`.
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

/// not_null_constraints returns the name and column number of the NOT NULL constraint of each column of a table that
/// cannot be NULL.
pub fn not_null_constraints(table: &TableDef) -> Vec<(String, i16)> {
    let columns = table.columns.iter().enumerate().filter(|(_, c)| !c.nullable);
    columns.map(|(i, c)| (table.not_null_name(c), i as i16 + 1)).collect()
}

/// constraint_oid returns the OID of a constraint in its ID section.
pub fn constraint_oid(section: u8, schema: &str, table: &str, name: &str) -> u32 {
    oids::oid(&id::new(section, &[schema, table, name]))
}

impl Ctx<'_> {
    /// snapshot returns every user object of the working root value, reading them once for each root value that a
    /// statement sees.
    pub fn snapshot(&mut self) -> Result<std::sync::Arc<Snapshot>> {
        let system = self.session.setting_on("dolt_show_system_tables");
        if let Some(cache) = &self.catalog
            && cache.root == self.txn.root
            && cache.system == system
        {
            return Ok(cache.snapshot.clone());
        }
        let snapshot = std::sync::Arc::new(self.read_snapshot(system)?);
        self.catalog =
            Some(CatalogCache { root: self.txn.root.clone(), snapshot: snapshot.clone(), system, relations: None });
        Ok(snapshot)
    }

    /// read_snapshot reads every user object of the working root value, with Dolt's own tables and each table's blame
    /// view when asked.
    fn read_snapshot(&mut self, system: bool) -> Result<Snapshot> {
        let mut schemas: Vec<String> =
            self.txn.root.schemas.iter().map(|s| String::from_utf8_lossy(s).into_owned()).collect();
        schemas.sort();
        let (mut tables, mut stored_system) = (Vec::new(), Vec::new());
        for (key, address) in self.txn.root.tables(self.db)? {
            let text = String::from_utf8_lossy(&key).into_owned();
            let mut parts = text.splitn(3, '\0').skip(1);
            let (Some(schema), Some(name)) = (parts.next(), parts.next()) else { continue };
            if schema == "dolt" {
                continue;
            }
            match name.starts_with("dolt_") {
                true if system => stored_system.push(TableDef::load(self.db, schema, name, address)?),
                true => {}
                false => tables.push(TableDef::load(self.db, schema, name, address)?),
            }
        }
        let show_system = system;
        let system = match system {
            true => crate::pgcatalog::systables::generated(&schemas, &tables, stored_system),
            false => Vec::new(),
        };
        for schema in &schemas {
            for table in self.nonlocal_tables(schema)? {
                if !tables.iter().any(|t| t.schema == table.schema && t.name == table.name) {
                    tables.push(table);
                }
            }
        }
        tables.sort_by(|a, b| (&a.schema, &a.name).cmp(&(&b.schema, &b.name)));
        let mut views = Vec::new();
        for schema in &schemas {
            for (name, statement) in self.views(schema)? {
                views.push(ViewDef { schema: schema.clone(), name, statement });
            }
        }
        for table in tables.iter().filter(|t| show_system && !t.keyless()) {
            let name = format!("dolt_blame_{}", table.name);
            let Some(statement) = crate::dolt::diff::blame_view(self, &table.schema, &name)? else { continue };
            views.retain(|v| v.schema != table.schema || v.name != name);
            views.push(ViewDef { schema: table.schema.clone(), name, statement });
        }
        let sequences = crate::sequences::all(self.db, &self.txn.root)?;
        let foreign_keys = self.foreign_keys()?;
        Ok(Snapshot { schemas, tables, views, sequences, foreign_keys, system })
    }
}
