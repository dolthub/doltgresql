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

//! EXPLAIN: a query's plan printed in Postgres' text format, with Doltgres-specific `Index Columns` and `Index Ranges`
//! lines that give an index scan's key columns and its ranges in go-mysql-server's notation.

use pg_query::NodeEnum;
use pg_query::protobuf::ExplainStmt;

use crate::Outcome;
use crate::error::{PgError, Result};
use crate::expr::{ArithOp, CmpOp, Expr};
use crate::plan::{JoinKind, Plan, Planner, SetOp};
use crate::query::Ctx;
use crate::types::Value;

/// cmp_text returns a comparison operator's symbol.
fn cmp_text(op: CmpOp) -> &'static str {
    match op {
        CmpOp::Eq => "=",
        CmpOp::Ne => "<>",
        CmpOp::Lt => "<",
        CmpOp::Le => "<=",
        CmpOp::Gt => ">",
        CmpOp::Ge => ">=",
    }
}

/// expr_text prints an expression over rows with the named columns, as EXPLAIN shows conditions and keys.
fn expr_text(e: &Expr, columns: &[String]) -> String {
    let text = |e: &Expr| expr_text(e, columns);
    match e {
        Expr::Column(i) => columns.get(*i).cloned().unwrap_or_else(|| "?".into()),
        Expr::Const(Value::Null) => "NULL".into(),
        Expr::Const(Value::Text(t)) => format!("'{}'::text", t.replace('\'', "''")),
        Expr::Const(v) => v.output().unwrap_or_default(),
        Expr::Param(i) => format!("${}", i + 1),
        Expr::Compare(op, l, r) => format!("({} {} {})", text(l), cmp_text(*op), text(r)),
        Expr::And(l, r) => format!("({} AND {})", text(l), text(r)),
        Expr::Or(l, r) => format!("({} OR {})", text(l), text(r)),
        Expr::Not(inner) => format!("(NOT {})", text(inner)),
        Expr::IsNull(inner, negated) => format!("({} IS {}NULL)", text(inner), if *negated { "NOT " } else { "" }),
        Expr::Cast(inner, ty, _) => format!("({})::{}", text(inner), crate::cast::type_display(ty.oid)),
        Expr::Concat(l, r) => format!("({} || {})", text(l), text(r)),
        Expr::Arith(op, l, r, _) => {
            let op = match op {
                ArithOp::Add => "+",
                ArithOp::Sub => "-",
                ArithOp::Mul => "*",
                ArithOp::Div => "/",
                ArithOp::Mod => "%",
            };
            format!("({} {op} {})", text(l), text(r))
        }
        Expr::Operator(name, _, l, r) => format!("({} {name} {})", text(l), text(r)),
        Expr::Func(index, args) => {
            let args: Vec<String> = args.iter().map(text).collect();
            format!("{}({})", crate::functions::function(*index).name, args.join(", "))
        }
        _ => "?".into(),
    }
}

/// columns returns the names of a plan's columns, which conditions over its rows print.
fn columns(plan: &Plan) -> Vec<String> {
    match plan {
        Plan::Scan(table) => table.columns.iter().map(|c| c.name.clone()).collect(),
        Plan::IndexScan(scan) => scan.table.columns.iter().map(|c| c.name.clone()).collect(),
        Plan::Filter { input, .. }
        | Plan::Sort { input, .. }
        | Plan::Limit { input, .. }
        | Plan::Distinct { input, .. } => columns(input),
        Plan::Project { input, exprs } => {
            let names = columns(input);
            exprs.iter().map(|e| expr_text(e, &names)).collect()
        }
        Plan::Join { left, right, .. } => {
            let mut names = columns(left);
            names.extend(columns(right));
            names
        }
        Plan::Aggregate { input, groups, .. } => {
            let names = columns(input);
            groups.iter().map(|e| expr_text(e, &names)).collect()
        }
        _ => Vec::new(),
    }
}

/// Printer collects the lines of a plan in Postgres' text format.
struct Printer {
    lines: Vec<String>,
}

impl Printer {
    /// node prints a plan node at a depth below the root, with the filters that the plans above it applied to its rows,
    /// as Postgres attaches filters to the nodes they test.
    fn node(&mut self, plan: &Plan, depth: usize, mut filters: Vec<String>) {
        let (name, mut properties, children): (String, Vec<String>, Vec<&Plan>) = match plan {
            Plan::Filter { input, predicate } => {
                filters.insert(0, format!("Filter: {}", expr_text(predicate, &columns(input))));
                return self.node(input, depth, filters);
            }
            Plan::Project { input, .. } => return self.node(input, depth, filters),
            Plan::Scan(table) => {
                (format!("Seq Scan on {}", crate::engine::quote_identifier(&table.name)), vec![], vec![])
            }
            Plan::IndexScan(scan) => {
                let descending: Vec<bool> = match scan.index {
                    Some(i) => scan.table.indexes[i].descending.clone(),
                    None => Vec::new(),
                };
                let names: Vec<String> = scan
                    .index_columns()
                    .iter()
                    .enumerate()
                    .map(|(i, &c)| {
                        let desc = if descending.get(i).copied().unwrap_or(false) { " DESC" } else { "" };
                        format!("{}{desc}", scan.table.columns[c].name)
                    })
                    .collect();
                let mut properties = vec![format!("Index Columns: {}", names.join(", "))];
                match &scan.nearest {
                    Some(nearest) => {
                        properties.push(format!("Order By: {}", expr_text(&nearest.order, &columns(plan))))
                    }
                    None => properties.push(format!("Index Ranges: {}", crate::ranges::ranges_text(&scan.ranges))),
                }
                (
                    format!(
                        "Index Scan{} using {} on {}",
                        if scan.reverse { " Backward" } else { "" },
                        crate::engine::quote_identifier(&scan.index_name()),
                        crate::engine::quote_identifier(&scan.table.name)
                    ),
                    properties,
                    vec![],
                )
            }
            Plan::Sort { input, keys } => {
                let names = columns(input);
                let keys: Vec<String> = keys
                    .iter()
                    .map(|k| {
                        let order = match (k.descending, k.nulls_first) {
                            (false, false) | (true, true) => "",
                            (false, true) => " NULLS FIRST",
                            (true, false) => " NULLS LAST",
                        };
                        format!("{}{}{order}", expr_text(&k.expr, &names), if k.descending { " DESC" } else { "" })
                    })
                    .collect();
                ("Sort".into(), vec![format!("Sort Key: {}", keys.join(", "))], vec![input.as_ref()])
            }
            Plan::Limit { input, .. } => ("Limit".into(), vec![], vec![input.as_ref()]),
            Plan::Distinct { input, .. } => ("Unique".into(), vec![], vec![input.as_ref()]),
            Plan::Aggregate { input, groups, .. } => {
                let names = columns(input);
                if groups.is_empty() {
                    ("Aggregate".into(), vec![], vec![input.as_ref()])
                } else {
                    let keys: Vec<String> = groups.iter().map(|g| expr_text(g, &names)).collect();
                    ("HashAggregate".into(), vec![format!("Group Key: {}", keys.join(", "))], vec![input.as_ref()])
                }
            }
            Plan::Join { left, right, kind, condition, .. } => {
                let name = match kind {
                    JoinKind::Inner => "Nested Loop",
                    JoinKind::Left => "Nested Loop Left Join",
                    JoinKind::Right => "Nested Loop Right Join",
                    JoinKind::Full => "Nested Loop Full Join",
                };
                let mut names = columns(left);
                names.extend(columns(right));
                let properties = condition.iter().map(|c| format!("Join Filter: {}", expr_text(c, &names))).collect();
                (name.into(), properties, vec![left.as_ref(), right.as_ref()])
            }
            Plan::SetOp { op, all, left, right } => {
                let name = match (op, all) {
                    (SetOp::Union, true) => "Append",
                    (SetOp::Union, false) => "Unique",
                    (SetOp::Intersect, true) => "HashSetOp Intersect All",
                    (SetOp::Intersect, false) => "HashSetOp Intersect",
                    (SetOp::Except, true) => "HashSetOp Except All",
                    (SetOp::Except, false) => "HashSetOp Except",
                };
                (name.into(), vec![], vec![left.as_ref(), right.as_ref()])
            }
            Plan::Values(_) => ("Values Scan on \"*VALUES*\"".into(), vec![], vec![]),
            Plan::Function { .. } | Plan::RowsFrom { .. } => ("Function Scan".into(), vec![], vec![]),
            Plan::QueryDiff(diff, _) => ("Query Diff".into(), vec![], vec![&diff.from, &diff.to]),
            Plan::XmlTable(_) | Plan::JsonTable(_) => ("Table Function Scan".into(), vec![], vec![]),
            Plan::Catalog(table) => (format!("Seq Scan on {}", table.name), vec![], vec![]),
            Plan::System(_) => ("Seq Scan on a Dolt system table".into(), vec![], vec![]),
            Plan::OneRow => ("Result".into(), vec![], vec![]),
            Plan::Recursive { anchor, step, .. } => {
                ("Recursive Union".into(), vec![], vec![anchor.as_ref(), step.as_ref()])
            }
            Plan::WorkTable(_) => ("WorkTable Scan".into(), vec![], vec![]),
            Plan::ProjectSet { input, .. } => ("ProjectSet".into(), vec![], vec![input.as_ref()]),
            Plan::Window { input, .. } => ("WindowAgg".into(), vec![], vec![input.as_ref()]),
        };
        properties.extend(filters);
        let prefix = if depth == 0 { String::new() } else { format!("{}->  ", " ".repeat(6 * (depth - 1) + 2)) };
        self.lines.push(format!("{prefix}{name}"));
        let pad = " ".repeat(6 * depth + 2);
        for property in properties {
            self.lines.push(format!("{pad}{property}"));
        }
        for child in children {
            self.node(child, depth + 1, Vec::new());
        }
    }
}

/// lines returns the lines that EXPLAIN prints for a plan.
pub fn lines(plan: &Plan) -> Vec<String> {
    let mut printer = Printer { lines: Vec::new() };
    printer.node(plan, 0, Vec::new());
    printer.lines
}

impl Ctx<'_> {
    /// explain_plan plans the query of an EXPLAIN.
    fn explain_plan(&mut self, stmt: &ExplainStmt) -> Result<Plan> {
        let Some(NodeEnum::SelectStmt(select)) = stmt.query.as_deref().and_then(|n| n.node.as_ref()) else {
            return Err(PgError::unsupported("EXPLAIN of this statement"));
        };
        Ok(Planner { ctx: self, outer: Vec::new() }.plan_query(select)?.plan)
    }

    /// explain_columns plans the query of an EXPLAIN for its parameter types, returning EXPLAIN's one column.
    pub fn explain_columns(&mut self, stmt: &ExplainStmt) -> Result<Vec<crate::Column>> {
        self.explain_plan(stmt)?;
        Ok(vec![crate::query::column("QUERY PLAN".into(), crate::expr::typ(crate::oid::TEXT))])
    }

    /// explain runs EXPLAIN of a query, returning its plan's lines in the QUERY PLAN column.
    pub fn explain(&mut self, stmt: &ExplainStmt) -> Result<Outcome> {
        let plan = self.explain_plan(stmt)?;
        let rows = lines(&plan).into_iter().map(|line| vec![Value::Text(line)]).collect();
        let columns = vec![crate::query::column("QUERY PLAN".into(), crate::expr::typ(crate::oid::TEXT))];
        Ok(Outcome::Rows { columns, rows, tag: "EXPLAIN".into() })
    }
}
