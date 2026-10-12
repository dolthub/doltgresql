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

//! DOLT_QUERY_DIFF: the rows that differ between the results of two queries, as Dolt's QueryDiffTableFunction finds
//! them.

use std::cmp::Ordering;

use pg_query::NodeEnum;

use crate::catalog::ColumnType;
use crate::dolt::args::error;
use crate::error::Result;
use crate::expr::compare_values;
use crate::plan::{Plan, Planner};
use crate::query::Ctx;
use crate::types::Value;

/// QueryDiff compares the rows of two query plans, matching them by the primary key columns they share, or
/// reporting every row as deleted or added when the queries' columns differ.
#[derive(Clone, Debug, PartialEq)]
pub struct QueryDiff {
    pub from: Plan,
    pub to: Plan,
    /// The positions of the primary key columns, or None for queries whose columns differ.
    pub keys: Option<Vec<usize>>,
    pub from_width: usize,
    pub to_width: usize,
}

/// QueryColumns are the names and types of a query diff's columns.
pub type QueryColumns = Vec<(String, ColumnType)>;

impl Ctx<'_> {
    /// plan_query_diff plans the two queries of a DOLT_QUERY_DIFF call, returning the diff and its columns: each
    /// column of the first query prefixed with `from_`, of the second with `to_`, and diff_type.
    pub fn plan_query_diff(&mut self, args: &[String]) -> Result<(QueryDiff, QueryColumns)> {
        if args.len() != 2 {
            return Err(crate::dolt::args::argument_count("dolt_query_diff", 2, args.len()));
        }
        let mut planned = Vec::new();
        for text in args {
            let text = text.trim();
            if !text.to_lowercase().starts_with("select") {
                return Err(error("query must be a SELECT statement"));
            }
            let Some(crate::parse::Statement::Postgres { node: NodeEnum::SelectStmt(select), extras }) =
                crate::parse::parse(text)?.into_iter().next()
            else {
                return Err(error("query must be a SELECT statement"));
            };
            let as_of = std::mem::replace(&mut self.session.as_of, extras.as_of);
            let query = Planner { ctx: self, outer: Vec::new() }.plan_query(&select);
            self.session.as_of = as_of;
            planned.push(query?);
        }
        let to = planned.pop().ok_or_else(|| error("query must be a SELECT statement"))?;
        let from = planned.pop().ok_or_else(|| error("query must be a SELECT statement"))?;
        let same = from.columns.len() == to.columns.len()
            && from.columns.iter().zip(&to.columns).all(|(a, b)| a.name == b.name && a.origin.0 == b.origin.0)
            && from.types == to.types;
        let keys = match same {
            true => Some(self.key_positions(&from.columns)?),
            false => None,
        };
        let mut columns: QueryColumns = Vec::new();
        for (prefix, query) in [("from_", &from), ("to_", &to)] {
            columns.extend(query.columns.iter().zip(&query.types).map(|(c, &t)| (format!("{prefix}{}", c.name), t)));
        }
        columns.push(("diff_type".into(), crate::expr::typ(crate::oid::TEXT)));
        let (from_width, to_width) = (from.columns.len(), to.columns.len());
        Ok((QueryDiff { from: from.plan, to: to.plan, keys, from_width, to_width }, columns))
    }

    /// key_positions returns the positions of the query columns that come from a primary key column of their table.
    fn key_positions(&mut self, columns: &[crate::Column]) -> Result<Vec<usize>> {
        let snapshot = self.snapshot()?;
        Ok(columns
            .iter()
            .enumerate()
            .filter(|(_, c)| {
                c.origin.1 > 0
                    && snapshot.tables.iter().any(|t| {
                        crate::pgcatalog::snapshot::table_oid(&t.schema, &t.name) == c.origin.0
                            && t.key_columns.contains(&(c.origin.1 as usize - 1))
                    })
            })
            .map(|(i, _)| i)
            .collect())
    }
}

impl QueryDiff {
    /// run returns the differing rows, in the order of the queries' rows, as Dolt's pkRowIter and keylessRowIter
    /// find them.
    pub fn run(&self, ctx: &mut Ctx<'_>) -> Result<Vec<Vec<Value>>> {
        let (from, to) = (self.from.run(ctx)?, self.to.run(ctx)?);
        let deleted = |row: &[Value]| {
            let mut out = row.to_vec();
            out.resize(self.from_width + self.to_width, Value::Null);
            out.push(Value::Text("deleted".into()));
            out
        };
        let added = |row: &[Value]| {
            let mut out = vec![Value::Null; self.from_width];
            out.extend_from_slice(row);
            out.push(Value::Text("added".into()));
            out
        };
        let Some(keys) = &self.keys else {
            return Ok(from.iter().map(|r| deleted(r)).chain(to.iter().map(|r| added(r))).collect());
        };
        let mut out = Vec::new();
        let (mut i, mut j) = (0, 0);
        while i < from.len() && j < to.len() {
            let ordering = keys
                .iter()
                .map(|&k| compare_values(&from[i][k], &to[j][k]))
                .find(|o| o.is_ne())
                .unwrap_or(Ordering::Equal);
            match ordering {
                Ordering::Less => {
                    out.push(deleted(&from[i]));
                    i += 1;
                }
                Ordering::Greater => {
                    out.push(added(&to[j]));
                    j += 1;
                }
                Ordering::Equal => {
                    if from[i].iter().zip(&to[j]).any(|(a, b)| compare_values(a, b).is_ne()) {
                        let mut row = from[i].clone();
                        row.extend_from_slice(&to[j]);
                        row.push(Value::Text("modified".into()));
                        out.push(row);
                    }
                    i += 1;
                    j += 1;
                }
            }
        }
        out.extend(from[i..].iter().map(|r| deleted(r)));
        out.extend(to[j..].iter().map(|r| added(r)));
        Ok(out)
    }
}
