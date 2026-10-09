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

//! The indexes of the system catalogs that Doltgres declares, which scans and join lookups of those catalogs use.

use crate::catalog::table::{ColumnDef, IndexDef, TableDef};
use crate::expr::Expr;
use crate::pgcatalog::CatalogTable;
use crate::query::Ctx;
use crate::ranges::Range;
use crate::types::Value;

/// CatalogIndex is an index of a system catalog relation.
#[derive(Debug, PartialEq)]
pub struct CatalogIndex {
    pub name: &'static str,
    pub columns: &'static [&'static str],
    pub unique: bool,
}

/// INDEXES are the indexes of the system catalogs, by relation, as the Go server declares them.
const INDEXES: &[(&str, CatalogIndex)] = &[
    (
        "pg_attribute",
        CatalogIndex { name: "pg_attribute_relid_attnum_index", columns: &["attrelid", "attnum"], unique: true },
    ),
    (
        "pg_attribute",
        CatalogIndex { name: "pg_attribute_relid_attnam_index", columns: &["attrelid", "attname"], unique: true },
    ),
    ("pg_class", CatalogIndex { name: "pg_class_oid_index", columns: &["oid"], unique: true }),
    (
        "pg_class",
        CatalogIndex { name: "pg_class_relname_nsp_index", columns: &["relname", "relnamespace"], unique: true },
    ),
    ("pg_constraint", CatalogIndex { name: "pg_constraint_oid_index", columns: &["oid"], unique: true }),
    (
        "pg_constraint",
        CatalogIndex {
            name: "pg_constraint_conrelid_contypid_conname_index",
            columns: &["conrelid", "contypid", "conname"],
            unique: true,
        },
    ),
    (
        "pg_constraint",
        CatalogIndex { name: "pg_constraint_conname_nsp_index", columns: &["conname", "connamespace"], unique: false },
    ),
    ("pg_constraint", CatalogIndex { name: "pg_constraint_contypid_index", columns: &["contypid"], unique: false }),
    ("pg_index", CatalogIndex { name: "pg_index_indexrelid_index", columns: &["indexrelid"], unique: true }),
    ("pg_index", CatalogIndex { name: "pg_index_indrelid_index", columns: &["indrelid"], unique: false }),
    ("pg_namespace", CatalogIndex { name: "pg_namespace_oid_index", columns: &["oid"], unique: true }),
    ("pg_namespace", CatalogIndex { name: "pg_namespace_nspname_index", columns: &["nspname"], unique: true }),
    ("pg_type", CatalogIndex { name: "pg_type_oid_index", columns: &["oid"], unique: true }),
    (
        "pg_type",
        CatalogIndex { name: "pg_type_typname_nsp_index", columns: &["typname", "typnamespace"], unique: true },
    ),
];

/// indexes returns the indexes of a system catalog relation.
pub fn indexes(table: &CatalogTable) -> Vec<&'static CatalogIndex> {
    match table.schema {
        "pg_catalog" => INDEXES.iter().filter(|(t, _)| *t == table.name).map(|(_, index)| index).collect(),
        _ => Vec::new(),
    }
}

/// CatalogIndexScan reads the rows of a system catalog relation whose keys in one of its indexes lie in ranges.
#[derive(Clone, Debug, PartialEq)]
pub struct CatalogIndexScan {
    pub table: &'static CatalogTable,
    pub index: &'static CatalogIndex,
    pub ranges: Vec<Range>,
    /// The positions of the index's columns among the relation's columns.
    pub key: Vec<usize>,
}

/// key_columns returns the positions of an index's columns among its relation's columns.
pub fn key_columns(table: &CatalogTable, index: &CatalogIndex) -> Vec<usize> {
    index.columns.iter().filter_map(|c| table.column(c)).collect()
}

/// table_def returns a table with a system catalog relation's columns and indexes and no storage, which index
/// planning can choose among the indexes of.
fn table_def(table: &CatalogTable, indexes: &[&CatalogIndex]) -> TableDef {
    let columns = table
        .columns
        .iter()
        .map(|c| ColumnDef {
            name: c.name.to_string(),
            ty: crate::expr::typ(c.type_oid),
            tag: 0,
            encoding: 0,
            nullable: !c.not_null,
            primary_key: false,
            default: String::new(),
            generated: false,
            mysql_type: String::new(),
            comment: String::new(),
            identity: 0,
            legacy_array: false,
        })
        .collect();
    let indexes = indexes
        .iter()
        .map(|index| IndexDef {
            name: index.name.to_string(),
            columns: key_columns(table, index),
            unique: index.unique,
            descending: vec![false; index.columns.len()],
            nulls_last: vec![false; index.columns.len()],
            op_classes: Vec::new(),
            comment: String::new(),
            predicate: String::new(),
            root: store::Hash::default(),
            system: false,
            vector: None,
            deferrable: false,
            initially_deferred: false,
            plain: false,
        })
        .collect();
    TableDef {
        schema: table.schema.to_string(),
        name: table.name.to_string(),
        primary: Default::default(),
        columns,
        hidden: Vec::new(),
        checks: Vec::new(),
        indexes,
        key_columns: Vec::new(),
        value_columns: Vec::new(),
        comment: String::new(),
        table: doltdb::table::Table {
            schema: store::Hash::default(),
            primary_index: Vec::new(),
            secondary_indexes: Vec::new(),
            auto_increment: 0,
            conflicts: Default::default(),
            violations: Vec::new(),
            artifacts: Vec::new(),
        },
    }
}

/// choose returns the scan of a system catalog relation's index that answers a filter best, with whether its ranges
/// hold exactly the rows that the filter keeps, when one of its indexes answers the filter.
pub fn choose(ctx: &mut Ctx<'_>, table: &'static CatalogTable, predicate: &Expr) -> Option<(CatalogIndexScan, bool)> {
    let indexes = indexes(table);
    if indexes.is_empty() {
        return None;
    }
    let (scan, covered) = crate::indexscan::choose_with_cover(ctx, &table_def(table, &indexes), predicate)?;
    let index = indexes[scan.index?];
    Some((CatalogIndexScan { table, index, ranges: scan.ranges, key: key_columns(table, index) }, covered))
}

impl CatalogIndexScan {
    /// run returns the rows whose keys lie in the scan's ranges, in the relation's order.
    pub fn run(&self, ctx: &mut Ctx<'_>) -> crate::error::Result<Vec<Vec<Value>>> {
        ctx.catalog_rows_in(self.table, Some(self))
    }

    /// contains returns whether a row's key lies in the scan's ranges.
    pub fn contains(&self, row: &[Value]) -> bool {
        self.ranges.iter().any(|range| range.iter().zip(&self.key).all(|(r, &c)| r.contains(&row[c])))
    }
}
