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

use std::cell::RefCell;
use std::collections::HashMap;

use pg_query::NodeEnum;
use pg_query::protobuf::ExplainStmt;

use crate::Outcome;
use crate::catalog::table::TableDef;
use crate::error::{PgError, Result};
use crate::expr::{ArithOp, CmpOp, Expr};
use crate::plan::{JoinKind, JoinMethod, Plan, Planner, SetOp};
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
        Expr::Spread(inner) => format!("VARIADIC {}", text(inner)),
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

thread_local! {
    /// RELATIONS are the names that the plan being printed gives the tables it scans, by the addresses of their
    /// definitions, and whether the conditions of the nodes above its scans qualify column names with them, as
    /// Postgres does once a query reads more than one relation.
    static RELATIONS: RefCell<(HashMap<usize, String>, bool)> = RefCell::new((HashMap::new(), false));
}

/// relation_name returns the name that EXPLAIN gives a scanned table: its alias in the query or its own name, made
/// unique among the plan's relations as Postgres' select_rtable_names_for_explain makes it.
fn relation_name(table: &TableDef) -> String {
    let address = table as *const TableDef as usize;
    RELATIONS
        .with(|r| r.borrow().0.get(&address).cloned())
        .unwrap_or_else(|| table.alias.clone().unwrap_or_else(|| table.name.clone()))
}

/// scan_target returns how a scan line names a table: its name, followed by the name the plan gives it when that
/// differs.
fn scan_target(table: &TableDef) -> String {
    let name = relation_name(table);
    match name == table.name {
        true => crate::engine::quote_identifier(&table.name),
        false => {
            format!("{} {}", crate::engine::quote_identifier(&table.name), crate::engine::quote_identifier(&name))
        }
    }
}

/// name_relations names the tables that a plan scans, in the order it reads them, numbering a name that repeats.
fn name_relations(plan: &Plan, names: &mut HashMap<usize, String>, taken: &mut HashMap<String, usize>) {
    let mut name = |table: &TableDef| {
        let base = table.alias.clone().unwrap_or_else(|| table.name.clone());
        let count = taken.entry(base.clone()).or_insert(0);
        let unique = if *count == 0 { base.clone() } else { format!("{base}_{count}") };
        *count += 1;
        names.insert(table as *const TableDef as usize, unique);
    };
    match plan {
        Plan::Scan(table, _) => name(table),
        Plan::IndexScan(scan) => name(&scan.table),
        Plan::Join { left, right, .. } | Plan::SetOp { left, right, .. } => {
            name_relations(left, names, taken);
            name_relations(right, names, taken);
        }
        Plan::Recursive { anchor, step, .. } => {
            name_relations(anchor, names, taken);
            name_relations(step, names, taken);
        }
        Plan::Filter { input, .. }
        | Plan::Project { input, .. }
        | Plan::Aggregate { input, .. }
        | Plan::Sort { input, .. }
        | Plan::Distinct { input, .. }
        | Plan::Limit { input, .. }
        | Plan::ProjectSet { input, .. }
        | Plan::Window { input, .. }
        | Plan::Once(input) => name_relations(input, names, taken),
        _ => {}
    }
}

/// qualify returns a column name prefixed with its relation's name when the plan's conditions qualify names.
fn qualify(relation: &str, column: &str) -> String {
    match RELATIONS.with(|r| r.borrow().1) {
        true => format!("{relation}.{column}"),
        false => column.to_string(),
    }
}

/// own_columns returns the names of a scan's columns as its own conditions print them, unqualified, or those of any
/// other plan as `columns` gives them.
fn own_columns(plan: &Plan) -> Vec<String> {
    match plan {
        Plan::Scan(table, _) => table.columns.iter().map(|c| c.name.clone()).collect(),
        Plan::IndexScan(scan) => scan.table.columns.iter().map(|c| c.name.clone()).collect(),
        Plan::Catalog(table) => table.columns.iter().map(|c| c.name.to_string()).collect(),
        Plan::CatalogIndexScan(scan) => scan.table.columns.iter().map(|c| c.name.to_string()).collect(),
        other => columns(other),
    }
}

/// columns returns the names of a plan's columns, which conditions over its rows print.
fn columns(plan: &Plan) -> Vec<String> {
    match plan {
        Plan::Scan(table, _) => {
            let relation = relation_name(table);
            table.columns.iter().map(|c| qualify(&relation, &c.name)).collect()
        }
        Plan::IndexScan(scan) => {
            let relation = relation_name(&scan.table);
            scan.table.columns.iter().map(|c| qualify(&relation, &c.name)).collect()
        }
        Plan::Catalog(table) => table.columns.iter().map(|c| qualify(table.name, c.name)).collect(),
        Plan::CatalogIndexScan(scan) => scan.table.columns.iter().map(|c| qualify(scan.table.name, c.name)).collect(),
        Plan::Filter { input, .. }
        | Plan::Sort { input, .. }
        | Plan::Limit { input, .. }
        | Plan::Distinct { input, .. }
        | Plan::Once(input) => columns(input),
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
    /// The number of subquery plans printed so far, which numbers them.
    subplans: usize,
}

/// Child is a node that EXPLAIN prints below another: a plan, the hash table that a hash join builds from a plan, or
/// the index lookups that a join makes for each of its left rows.
enum Child<'p> {
    Plan(&'p Plan),
    Hash(&'p Plan),
    Lookup(Lookup<'p>),
}

/// Lookup is the index lookups that a join makes for each of its left rows: the index and the relation it belongs
/// to, the index's columns, the key expressions over the left row's columns, and the right input.
struct Lookup<'p> {
    index: String,
    relation: String,
    columns: Vec<String>,
    keys: &'p [Expr],
    left_columns: Vec<String>,
    right: &'p Plan,
}

impl Printer {
    /// node prints a plan node at a depth below the root, with the filters that the plans above it applied to its rows,
    /// as Postgres attaches filters to the nodes they test, and the plans of the subqueries in those filters and in the
    /// expressions of the projections above it.
    fn node<'p>(&mut self, plan: &'p Plan, depth: usize, mut filters: Vec<String>, mut evaluated: Vec<&'p Expr>) {
        let (name, mut properties, children): (String, Vec<String>, Vec<Child<'_>>) = match plan {
            Plan::Filter { input, predicate } => {
                filters.insert(0, format!("Filter: {}", expr_text(predicate, &own_columns(input))));
                evaluated.push(predicate);
                return self.node(input, depth, filters, evaluated);
            }
            Plan::Project { input, exprs } => {
                evaluated.extend(exprs);
                return self.node(input, depth, filters, evaluated);
            }
            Plan::Once(input) => return self.node(input, depth, filters, evaluated),
            Plan::Scan(table, _) => (format!("Seq Scan on {}", scan_target(table)), vec![], vec![]),
            Plan::IndexScan(scan) => {
                let descending: Vec<bool> = match scan.index {
                    Some(i) => scan.table.indexes[i].descending.clone(),
                    None => Vec::new(),
                };
                let mut names: Vec<String> = scan
                    .index_columns()
                    .iter()
                    .enumerate()
                    .map(|(i, &c)| {
                        let desc = if descending.get(i).copied().unwrap_or(false) { " DESC" } else { "" };
                        format!("{}{desc}", scan.table.index_column(c).map_or("", |c| c.name.as_str()))
                    })
                    .collect();
                if let Some(index) = scan.index.map(|i| &scan.table.indexes[i]).filter(|i| !i.predicate.is_empty()) {
                    names.push(index.predicate.clone());
                }
                let mut properties = vec![format!("Index Columns: {}", names.join(", "))];
                match &scan.nearest {
                    Some(nearest) => {
                        properties.push(format!("Order By: {}", expr_text(&nearest.order, &own_columns(plan))))
                    }
                    None => properties.push(format!("Index Ranges: {}", crate::ranges::ranges_text(&scan.ranges))),
                }
                (
                    format!(
                        "Index Scan{} using {} on {}",
                        if scan.reverse { " Backward" } else { "" },
                        crate::engine::quote_identifier(&scan.index_name()),
                        scan_target(&scan.table)
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
                ("Sort".into(), vec![format!("Sort Key: {}", keys.join(", "))], vec![Child::Plan(input)])
            }
            Plan::Limit { input, .. } => ("Limit".into(), vec![], vec![Child::Plan(input)]),
            Plan::Distinct { input, .. } => ("Unique".into(), vec![], vec![Child::Plan(input)]),
            Plan::Aggregate { input, groups, .. } => {
                let names = columns(input);
                if groups.is_empty() {
                    ("Aggregate".into(), vec![], vec![Child::Plan(input)])
                } else {
                    let keys: Vec<String> = groups.iter().map(|g| expr_text(g, &names)).collect();
                    ("HashAggregate".into(), vec![format!("Group Key: {}", keys.join(", "))], vec![Child::Plan(input)])
                }
            }
            Plan::Join { left, right, kind, condition, method, .. } => {
                evaluated.extend(condition);
                let kind = match kind {
                    JoinKind::Inner => "",
                    JoinKind::Left => " Left Join",
                    JoinKind::Right => " Right Join",
                    JoinKind::Full => " Full Join",
                    JoinKind::Anti => " Anti Join",
                    JoinKind::Semi => " Semi Join",
                };
                let mut names = columns(left);
                names.extend(columns(right));
                let printed =
                    |key: &str| condition.iter().map(|c| format!("{key}: {}", expr_text(c, &names))).collect();
                match method {
                    JoinMethod::Hash => {
                        let width = left.width();
                        let (hashed, rest): (Vec<&Expr>, Vec<&Expr>) = condition
                            .iter()
                            .flat_map(crate::indexscan::conjuncts)
                            .partition(|c| !crate::plan::join_keys(c, width).0.is_empty());
                        let mut properties = Vec::new();
                        for (key, clauses) in [("Hash Cond", hashed), ("Join Filter", rest)] {
                            let texts: Vec<String> = clauses.into_iter().map(|c| expr_text(c, &names)).collect();
                            match texts.as_slice() {
                                [] => {}
                                [one] => properties.push(format!("{key}: {one}")),
                                many => properties.push(format!("{key}: ({})", many.join(" AND "))),
                            }
                        }
                        (
                            format!("Hash {}", if kind.is_empty() { "Join" } else { kind.trim_start() }),
                            properties,
                            vec![Child::Plan(left), Child::Hash(right)],
                        )
                    }
                    JoinMethod::Lookup { .. } | JoinMethod::CatalogLookup { .. } => {
                        let lookup = Lookup::of(method, columns(left), right);
                        let keys = lookup.keys;
                        let looked_up = |c: &&Expr| matches!(c, Expr::Compare(CmpOp::Eq, a, b) if keys.contains(a) || keys.contains(b));
                        let rest: Vec<String> = condition
                            .iter()
                            .flat_map(crate::indexscan::conjuncts)
                            .filter(|c| !looked_up(c))
                            .map(|c| expr_text(c, &names))
                            .collect();
                        let properties = match rest.is_empty() {
                            true => Vec::new(),
                            false => vec![format!("Join Filter: {}", rest.join(" AND "))],
                        };
                        (format!("Nested Loop{kind}"), properties, vec![Child::Plan(left), Child::Lookup(lookup)])
                    }
                    _ => (
                        format!("Nested Loop{kind}"),
                        printed("Join Filter"),
                        vec![Child::Plan(left), Child::Plan(right)],
                    ),
                }
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
                (name.into(), vec![], vec![Child::Plan(left), Child::Plan(right)])
            }
            Plan::Values(_) => ("Values Scan on \"*VALUES*\"".into(), vec![], vec![]),
            Plan::Function { .. } | Plan::RowsFrom { .. } => ("Function Scan".into(), vec![], vec![]),
            Plan::QueryDiff(diff, _) => {
                ("Query Diff".into(), vec![], vec![Child::Plan(&diff.from), Child::Plan(&diff.to)])
            }
            Plan::XmlTable(_) | Plan::JsonTable(_) => ("Table Function Scan".into(), vec![], vec![]),
            Plan::Catalog(table) => (format!("Seq Scan on {}", table.name), vec![], vec![]),
            Plan::CatalogIndexScan(scan) => (
                format!("Index Scan using {} on {}", scan.index.name, scan.table.name),
                vec![
                    format!("Index Columns: {}", scan.index.columns.join(", ")),
                    format!("Index Ranges: {}", crate::ranges::ranges_text(&scan.ranges)),
                ],
                vec![],
            ),
            Plan::System(crate::dolt::tables::SystemTable::User(table)) => match &table.lookup {
                Some((column, commit)) => {
                    let commit = match commit {
                        Expr::Const(Value::Text(text)) => text.clone(),
                        other => expr_text(other, &[]),
                    };
                    (
                        format!("Index Scan using {column} on {}", table.table_name()),
                        vec![format!("Index Columns: {column}"), format!("Index Ranges: [{{[{commit}, {commit}]}}]")],
                        vec![],
                    )
                }
                None => (format!("Seq Scan on {}", table.table_name()), vec![], vec![]),
            },
            Plan::System(_) => ("Seq Scan on a Dolt system table".into(), vec![], vec![]),
            Plan::OneRow => ("Result".into(), vec![], vec![]),
            Plan::Recursive { anchor, step, .. } => {
                ("Recursive Union".into(), vec![], vec![Child::Plan(anchor), Child::Plan(step)])
            }
            Plan::WorkTable(..) => ("WorkTable Scan".into(), vec![], vec![]),
            Plan::ProjectSet { input, .. } => ("ProjectSet".into(), vec![], vec![Child::Plan(input)]),
            Plan::Window { input, .. } => ("WindowAgg".into(), vec![], vec![Child::Plan(input)]),
        };
        properties.extend(filters);
        let prefix = if depth == 0 { String::new() } else { format!("{}->  ", " ".repeat(6 * (depth - 1) + 2)) };
        self.lines.push(format!("{prefix}{name}"));
        let pad = " ".repeat(6 * depth + 2);
        for property in properties {
            self.lines.push(format!("{pad}{property}"));
        }
        for expr in evaluated {
            expr.visit(&mut |e| {
                let (Expr::Exists(subquery)
                | Expr::Scalar(subquery)
                | Expr::ArraySubquery(subquery, _)
                | Expr::AnySubquery(_, subquery, _)) = e
                else {
                    return;
                };
                self.subplans += 1;
                self.lines.push(format!("{pad}SubPlan {}", self.subplans));
                self.node(subquery, depth + 1, Vec::new(), Vec::new());
            });
        }
        for child in children {
            match child {
                Child::Plan(plan) => self.node(plan, depth + 1, Vec::new(), Vec::new()),
                Child::Hash(plan) => {
                    self.lines.push(format!("{}->  Hash", " ".repeat(6 * depth + 2)));
                    self.node(plan, depth + 2, Vec::new(), Vec::new());
                }
                Child::Lookup(lookup) => self.lookup(&lookup, depth + 1),
            }
        }
    }

    /// lookup prints the index lookups that a join makes for each left row, as Postgres prints the inner index scan of
    /// a nested loop, with the right input's filter.
    fn lookup(&mut self, lookup: &Lookup<'_>, depth: usize) {
        let pad = " ".repeat(6 * depth + 2);
        self.lines.push(format!(
            "{}->  Index Scan using {} on {}",
            " ".repeat(6 * (depth - 1) + 2),
            crate::engine::quote_identifier(&lookup.index),
            lookup.relation
        ));
        self.lines.push(format!("{pad}Index Columns: {}", lookup.columns.join(", ")));
        let conditions: Vec<String> = lookup
            .keys
            .iter()
            .zip(&lookup.columns)
            .map(|(key, column)| format!("({column} = {})", expr_text(key, &lookup.left_columns)))
            .collect();
        self.lines.push(format!("{pad}Index Cond: {}", conditions.join(" AND ")));
        if let Plan::Filter { input, predicate } = lookup.right {
            self.lines.push(format!("{pad}Filter: {}", expr_text(predicate, &own_columns(input))));
        }
    }
}

impl<'p> Lookup<'p> {
    /// of returns the lookups of a join that looks rows up, given the names of its left input's columns.
    fn of(method: &'p JoinMethod, left_columns: Vec<String>, right: &'p Plan) -> Lookup<'p> {
        let (index, relation, columns, keys) = match method {
            JoinMethod::Lookup { scan, keys } => (
                scan.index_name(),
                scan_target(&scan.table),
                scan.index_columns()
                    .iter()
                    .map(|&c| scan.table.index_column(c).map_or_else(String::new, |c| c.name.clone()))
                    .collect(),
                keys.as_slice(),
            ),
            JoinMethod::CatalogLookup { index, keys } => (
                index.name.to_string(),
                catalog_name(right),
                index.columns.iter().map(|c| c.to_string()).collect(),
                keys.as_slice(),
            ),
            _ => (String::new(), String::new(), Vec::new(), &[][..]),
        };
        Lookup { index, relation, columns, keys, left_columns, right }
    }
}

/// catalog_name returns the name of the system catalog relation that a plan reads.
fn catalog_name(plan: &Plan) -> String {
    match plan {
        Plan::Filter { input, .. } => catalog_name(input),
        Plan::Catalog(table) => table.name.to_string(),
        Plan::CatalogIndexScan(scan) => scan.table.name.to_string(),
        _ => String::new(),
    }
}

/// lines returns the lines that EXPLAIN prints for a plan.
pub fn lines(plan: &Plan) -> Vec<String> {
    let mut names = HashMap::new();
    name_relations(plan, &mut names, &mut HashMap::new());
    let qualified = names.len() > 1;
    RELATIONS.with(|r| *r.borrow_mut() = (names, qualified));
    let mut printer = Printer { lines: Vec::new(), subplans: 0 };
    printer.node(plan, 0, Vec::new(), Vec::new());
    RELATIONS.with(|r| *r.borrow_mut() = (HashMap::new(), false));
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
