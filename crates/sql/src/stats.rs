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

//! Table statistics: the histograms that ANALYZE builds over each index of a table, with a bucket for each node of
//! one level of the index's prolly tree as Dolt's statspro builds them, which the engine keeps in memory and
//! dolt_statistics lists.

use doltdb::database::Database;
use pg_query::NodeEnum;
use pg_query::protobuf::VacuumStmt;
use prolly::Node;
use prolly::Tuple;

use crate::Outcome;
use crate::catalog::ColumnType;
use crate::catalog::table::TableDef;
use crate::error::{PgError, Result, code};
use crate::query::Ctx;
use crate::types::Value;

/// LOW_BUCKETS is the number of tree nodes that a histogram descends levels until it has, as Dolt's bucketLowCnt.
const LOW_BUCKETS: usize = 20;

/// MCV_COUNT is how many most common keys each bucket keeps, as Dolt's mcvCnt.
const MCV_COUNT: usize = 3;

/// Bucket is a histogram bucket, which counts the keys of one tree node's subtree.
#[derive(Clone, Debug)]
pub struct Bucket {
    pub rows: u64,
    pub distinct: u64,
    pub nulls: u64,
    /// The bucket's last key, and how many rows have it.
    pub upper_bound: Vec<Value>,
    pub upper_bound_count: u64,
    /// The most common keys, least common first, with how many rows have each.
    pub mcvs: Vec<(Vec<Value>, u64)>,
}

/// Statistic is the histogram of one index of a table.
#[derive(Clone, Debug)]
pub struct Statistic {
    pub schema: String,
    pub table: String,
    pub index: String,
    pub columns: Vec<String>,
    pub types: Vec<String>,
    /// When ANALYZE built the histogram, as a UTC timestamp.
    pub created: i64,
    pub buckets: Vec<Bucket>,
}

/// KeyShape is how an index's keys begin: the encoding and type of each indexed column.
struct KeyShape {
    fields: Vec<(u8, ColumnType)>,
}

impl KeyShape {
    /// same reports whether two keys have equal indexed columns.
    fn same(&self, left: &[u8], right: &[u8]) -> Result<bool> {
        for (i, &(encoding, ty)) in self.fields.iter().enumerate() {
            let (l, r) = (Tuple(left).field(i)?, Tuple(right).field(i)?);
            if crate::storage::compare_key_field(encoding, ty, l, r) != std::cmp::Ordering::Equal {
                return Ok(false);
            }
        }
        Ok(true)
    }

    /// has_null reports whether a key has a NULL indexed column.
    fn has_null(&self, key: &[u8]) -> Result<bool> {
        for i in 0..self.fields.len() {
            if Tuple(key).field(i)?.is_none() {
                return Ok(true);
            }
        }
        Ok(false)
    }

    /// values decodes a key's indexed columns.
    fn values(&self, db: &Database, key: &[u8]) -> Result<Vec<Value>> {
        let mut out = Vec::with_capacity(self.fields.len());
        for (i, &(encoding, ty)) in self.fields.iter().enumerate() {
            out.push(crate::storage::decode_field(db, Tuple(key).field(i)?, encoding, ty)?);
        }
        Ok(out)
    }
}

/// histogram_level returns the nodes of the highest level of a tree with at least LOW_BUCKETS nodes, or its leaves,
/// as Dolt's GetHistogramLevel does.
fn histogram_level(db: &Database, root: Node) -> Result<Vec<Node>> {
    if root.tree_count() == 0 {
        return Ok(Vec::new());
    }
    let mut level = vec![root];
    while level.len() < LOW_BUCKETS && level.first().is_some_and(|n| n.level() > 0) {
        let mut next = Vec::new();
        for node in &level {
            for i in 0..node.count() {
                next.push(Node::load(db, &node.child(i)?)?);
            }
        }
        level = next;
    }
    Ok(level)
}

/// bucket counts the keys under a node, as Dolt's bucketBuilder does.
fn bucket(db: &Database, node: &Node, shape: &KeyShape) -> Result<Bucket> {
    let mut keys = Vec::new();
    prolly::walk_leaves(db, node, &mut |key, _| {
        keys.push(key.to_vec());
        Ok(())
    })?;
    let (mut rows, mut distinct, mut nulls) = (0u64, 0u64, 0u64);
    let mut current: Option<Vec<u8>> = None;
    let mut current_count = 0u64;
    let mut mcvs: Vec<(Vec<u8>, u64)> = Vec::new();
    let keep = |mcvs: &mut Vec<(Vec<u8>, u64)>, key: Vec<u8>, count: u64| {
        mcvs.push((key, count));
        if mcvs.len() > MCV_COUNT
            && let Some(least) = mcvs.iter().enumerate().min_by_key(|(_, (_, c))| *c).map(|(i, _)| i)
        {
            mcvs.remove(least);
        }
    };
    for key in keys {
        let same = match &current {
            Some(previous) => shape.same(previous, &key)?,
            None => false,
        };
        if same {
            current_count += 1;
        } else {
            if let Some(previous) = current.take() {
                keep(&mut mcvs, previous, current_count);
            }
            distinct += 1;
            current_count = 1;
            current = Some(key.clone());
        }
        rows += 1;
        if shape.has_null(&key)? {
            nulls += 1;
        }
    }
    if let Some(last) = &current {
        keep(&mut mcvs, last.clone(), current_count);
    }
    mcvs.sort_by_key(|(_, c)| *c);
    let cutoff = 2.0 * rows as f64 / distinct.max(1) as f64;
    let start = mcvs.iter().position(|(_, c)| *c as f64 >= cutoff).unwrap_or(mcvs.len());
    let mcvs = mcvs[start..].iter().map(|(k, c)| Ok((shape.values(db, k)?, *c))).collect::<Result<Vec<_>>>()?;
    let upper_bound = match &current {
        Some(key) => shape.values(db, key)?,
        None => Vec::new(),
    };
    Ok(Bucket { rows, distinct, nulls, upper_bound, upper_bound_count: current_count, mcvs })
}

/// build returns the histograms of a table's primary key and secondary indexes, leaving out vector indexes and the
/// primary key of a keyless table.
pub fn build(db: &Database, table: &TableDef, created: i64) -> Result<Vec<Statistic>> {
    let mut out = Vec::new();
    let mut add = |name: &str, root: Node, columns: &[usize]| -> Result<()> {
        let column = |c: usize| table.index_column(c).expect("an index column");
        let shape = KeyShape { fields: columns.iter().map(|&c| (column(c).encoding, column(c).ty)).collect() };
        let buckets = histogram_level(db, root)?.iter().map(|n| bucket(db, n, &shape)).collect::<Result<_>>()?;
        out.push(Statistic {
            schema: table.schema.clone(),
            table: table.name.clone(),
            index: name.to_string(),
            columns: columns.iter().map(|&c| column(c).name.to_lowercase()).collect(),
            types: columns.iter().map(|&c| crate::cast::type_display(column(c).ty.oid).into_owned()).collect(),
            created,
            buckets,
        });
        Ok(())
    };
    if !table.keyless() {
        add("primary", Node::decode(table.table.primary_index.clone())?, &table.key_columns)?;
    }
    for index in table.indexes.iter().filter(|i| i.vector.is_none()) {
        add(&index.name, Node::load(db, &index.root)?, &index.columns)?;
    }
    Ok(out)
}

/// COLUMNS are the columns of dolt_statistics, with Doltgres' types for Dolt's.
pub const COLUMNS: &[(&str, u32)] = &[
    ("database_name", crate::oid::TEXT),
    ("table_name", crate::oid::TEXT),
    ("index_name", crate::oid::TEXT),
    ("row_count", crate::oid::NUMERIC),
    ("distinct_count", crate::oid::NUMERIC),
    ("null_count", crate::oid::NUMERIC),
    ("columns", crate::oid::TEXT),
    ("types", crate::oid::TEXT),
    ("upper_bound", crate::oid::TEXT),
    ("upper_bound_cnt", crate::oid::NUMERIC),
    ("created_at", crate::oid::TIMESTAMP),
    ("mcv1", crate::oid::TEXT),
    ("mcv2", crate::oid::TEXT),
    ("mcv3", crate::oid::TEXT),
    ("mcv4", crate::oid::TEXT),
    ("mcv_counts", crate::oid::TEXT),
];

/// key_text joins a key's values with commas, as GMS's StringifyKey does.
fn key_text(values: &[Value]) -> String {
    values.iter().map(|v| v.output().unwrap_or_default()).collect::<Vec<_>>().join(",")
}

impl Ctx<'_> {
    /// analyze runs ANALYZE, which builds the histograms of the named tables, or of every table without names.
    pub fn analyze(&mut self, stmt: &VacuumStmt) -> Result<Outcome> {
        if stmt.is_vacuumcmd {
            return Err(PgError::unsupported("VACUUM"));
        }
        let mut tables = Vec::new();
        for node in &stmt.rels {
            let Some(NodeEnum::VacuumRelation(relation)) = node.node.as_ref() else { continue };
            let Some(relation) = &relation.relation else { continue };
            if !relation.catalogname.is_empty() && relation.catalogname != self.session.database {
                return Err(PgError::new(
                    code::FEATURE_NOT_SUPPORTED,
                    format!(
                        "cross-database references are not implemented: \"{}.{}.{}\"",
                        relation.catalogname, relation.schemaname, relation.relname
                    ),
                ));
            }
            tables.push(self.resolve_table(relation)?);
        }
        if stmt.rels.is_empty() {
            tables = self.snapshot()?.tables;
        }
        let created = crate::datetime::clock();
        for table in tables {
            let statistics = build(self.db, &table, created)?;
            let (database, branch) = (self.session.database.clone(), self.session.branch.clone());
            self.session.engine.put_statistics(&database, &branch, &table.schema, &table.name, statistics);
        }
        Ok(Outcome::command("ANALYZE"))
    }

    /// statistics_rows returns the rows of dolt_statistics: a row for each bucket of the histograms of the tables
    /// that still exist.
    pub fn statistics_rows(&mut self) -> Result<Vec<Vec<Value>>> {
        let snapshot = self.snapshot()?;
        let (database, branch) = (self.session.database.clone(), self.session.branch.clone());
        let mut rows = Vec::new();
        for statistic in self.session.engine.statistics(&database, &branch) {
            if !snapshot.tables.iter().any(|t| t.schema == statistic.schema && t.name == statistic.table) {
                continue;
            }
            for bucket in &statistic.buckets {
                let count = |n: u64| Value::Numeric(crate::numeric::Numeric::from_i64(n as i64));
                let mut mcvs: Vec<Value> = bucket.mcvs.iter().map(|(k, _)| Value::Text(key_text(k))).collect();
                mcvs.resize(4, Value::Text(String::new()));
                let counts: Vec<String> = bucket.mcvs.iter().map(|(_, c)| c.to_string()).collect();
                let mut row = vec![
                    Value::Text(database.clone()),
                    Value::Text(statistic.table.clone()),
                    Value::Text(statistic.index.clone()),
                    count(bucket.rows),
                    count(bucket.distinct),
                    count(bucket.nulls),
                    Value::Text(statistic.columns.join(",")),
                    Value::Text(statistic.types.join(",")),
                    Value::Text(key_text(&bucket.upper_bound)),
                    count(bucket.upper_bound_count),
                    Value::Timestamp(statistic.created),
                ];
                row.extend(mcvs);
                row.push(Value::Text(counts.join(",")));
                rows.push(row);
            }
        }
        Ok(rows)
    }
}
