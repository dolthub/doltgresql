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
        Expr::Const(Value::Bool(b)) => b.to_string(),
        Expr::Const(v) => v.output().unwrap_or_default(),
        Expr::Param(i) => format!("${}", i + 1),
        Expr::Compare(op, l, r) => format!("({} {} {})", text(l), cmp_text(*op), text(r)),
        Expr::RowCompare(op, l, r) => {
            let row = |fields: &[Expr]| fields.iter().map(text).collect::<Vec<String>>().join(", ");
            format!("(ROW({}) {} ROW({}))", row(l), cmp_text(*op), row(r))
        }
        Expr::And(..) => {
            let args: Vec<String> = crate::indexscan::conjuncts(e).into_iter().map(text).collect();
            format!("({})", args.join(" AND "))
        }
        Expr::Or(..) => {
            let args: Vec<String> = crate::optimizer::or_args(e).into_iter().map(text).collect();
            format!("({})", args.join(" OR "))
        }
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
        Expr::Outer(depth, i) => ENCLOSING.with(|e| {
            let enclosing = e.borrow();
            let row = enclosing.len().checked_sub(*depth).and_then(|level| enclosing.get(level));
            row.and_then(|names| names.get(*i)).cloned().unwrap_or_else(|| "?".into())
        }),
        _ => "?".into(),
    }
}

thread_local! {
    /// ENCLOSING are the names of the columns of the rows that the lateral joins around the node being printed push
    /// as its enclosing rows, innermost last.
    static ENCLOSING: RefCell<Vec<Vec<String>>> = const { RefCell::new(Vec::new()) };

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
        Plan::BitmapHeapScan(scan) => name(&scan.table),
        Plan::Join { left, right, .. } | Plan::SetOp { left, right, .. } => {
            name_relations(left, names, taken);
            name_relations(right, names, taken);
        }
        Plan::MergeAppend { inputs, .. } => inputs.iter().for_each(|input| name_relations(input, names, taken)),
        Plan::Memoize { input, .. } | Plan::OneTimeFilter { input, .. } => name_relations(input, names, taken),
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
        Plan::BitmapHeapScan(scan) => scan.table.columns.iter().map(|c| c.name.clone()).collect(),
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
        Plan::BitmapHeapScan(scan) => {
            let relation = relation_name(&scan.table);
            scan.table.columns.iter().map(|c| qualify(&relation, &c.name)).collect()
        }
        Plan::Catalog(table) => table.columns.iter().map(|c| qualify(table.name, c.name)).collect(),
        Plan::CatalogIndexScan(scan) => scan.table.columns.iter().map(|c| qualify(scan.table.name, c.name)).collect(),
        Plan::Filter { input, .. }
        | Plan::Sort { input, .. }
        | Plan::Limit { input, .. }
        | Plan::Distinct { input, .. }
        | Plan::Memoize { input, .. }
        | Plan::OneTimeFilter { input, .. }
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
    /// A plan under a node of the given name that holds its rows, such as a Hash or a Materialize.
    Held(&'static str, &'p Plan),
    Lookup(Lookup<'p>),
    /// A node of the tree of index scans that a bitmap heap scan reads the keys of.
    Bitmap(&'p crate::indexscan::Bitmap),
    /// The right input of a lateral join, which reads a left row of the given columns as its enclosing row.
    Lateral(&'p Plan, Vec<String>),
    /// A child under a Memoize of its rows by the keys, which read a left row of the given columns.
    Memoized(&'p [Expr], bool, Vec<String>, Box<Child<'p>>),
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
                let mut properties = vec![index_columns(scan)];
                match &scan.nearest {
                    Some(nearest) => {
                        properties.push(format!("Order By: {}", expr_text(&nearest.order, &own_columns(plan))))
                    }
                    None => properties.push(index_ranges(scan)),
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
            Plan::BitmapHeapScan(scan) => (
                format!("Bitmap Heap Scan on {}", scan_target(&scan.table)),
                scan.recheck.iter().map(|e| format!("Recheck Cond: {}", expr_text(e, &own_columns(plan)))).collect(),
                vec![Child::Bitmap(&scan.bitmap)],
            ),
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
            Plan::Join { left, right, kind, condition, method, lateral, .. } => {
                evaluated.extend(condition);
                let kind = match kind {
                    JoinKind::Inner => "",
                    JoinKind::Left => " Left Join",
                    JoinKind::Right => " Right Join",
                    JoinKind::Full => " Full Join",
                    JoinKind::Anti => " Anti Join",
                    JoinKind::Semi => " Semi Join",
                    JoinKind::RightSemi => " Right Semi Join",
                    JoinKind::RightAnti => " Right Anti Join",
                };
                let mut names = columns(left);
                names.extend(columns(right));
                let printed =
                    |key: &str| condition.iter().map(|c| format!("{key}: {}", expr_text(c, &names))).collect();
                let conditions = |(key, clauses): (&str, Vec<&Expr>)| {
                    let texts: Vec<String> = clauses.into_iter().map(|c| expr_text(c, &names)).collect();
                    match texts.as_slice() {
                        [] => None,
                        [one] => Some(format!("{key}: {one}")),
                        many => Some(format!("{key}: ({})", many.join(" AND "))),
                    }
                };
                match method {
                    JoinMethod::Hash => {
                        let width = left.width();
                        let (hashed, rest): (Vec<&Expr>, Vec<&Expr>) = condition
                            .iter()
                            .flat_map(crate::indexscan::conjuncts)
                            .partition(|c| !crate::plan::join_keys(c, width).0.is_empty());
                        (
                            format!("Hash {}", if kind.is_empty() { "Join" } else { kind.trim_start() }),
                            [("Hash Cond", hashed), ("Join Filter", rest)].into_iter().flat_map(conditions).collect(),
                            vec![Child::Plan(left), Child::Held("Hash", right)],
                        )
                    }
                    JoinMethod::Merge { clauses, materialized } => {
                        let mut merged: Vec<&Expr> = condition.iter().flat_map(crate::indexscan::conjuncts).collect();
                        let rest = merged.split_off(*clauses);
                        let right = match materialized {
                            true => Child::Held("Materialize", right),
                            false => Child::Plan(right),
                        };
                        (
                            format!("Merge {}", if kind.is_empty() { "Join" } else { kind.trim_start() }),
                            [("Merge Cond", merged), ("Join Filter", rest)].into_iter().flat_map(conditions).collect(),
                            vec![Child::Plan(left), right],
                        )
                    }
                    JoinMethod::Lookup { .. } | JoinMethod::CatalogLookup { .. } => {
                        let (memoize, right) = match &**right {
                            Plan::Memoize { input, keys, binary } => (Some((keys, *binary)), &**input),
                            other => (None, other),
                        };
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
                        let lookup = match memoize {
                            Some((keys, binary)) => {
                                Child::Memoized(keys, binary, columns(left), Box::new(Child::Lookup(lookup)))
                            }
                            None => Child::Lookup(lookup),
                        };
                        (format!("Nested Loop{kind}"), properties, vec![Child::Plan(left), lookup])
                    }
                    JoinMethod::MaterializedLoop => (
                        format!("Nested Loop{kind}"),
                        printed("Join Filter"),
                        vec![Child::Plan(left), Child::Held("Materialize", right)],
                    ),
                    _ => {
                        let right = match lateral {
                            true => Child::Lateral(right, columns(left)),
                            false => Child::Plan(right),
                        };
                        (format!("Nested Loop{kind}"), printed("Join Filter"), vec![Child::Plan(left), right])
                    }
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
            Plan::Memoize { input, keys, binary } => {
                ("Memoize".into(), memoize_properties(keys, *binary), vec![Child::Plan(input)])
            }
            Plan::OneTimeFilter { input, condition } => {
                let children = match &**input {
                    Plan::OneRow => Vec::new(),
                    Plan::Project { input, .. } if **input == Plan::OneRow => Vec::new(),
                    input => vec![Child::Plan(input)],
                };
                ("Result".into(), vec![format!("One-Time Filter: {}", expr_text(condition, &[]))], children)
            }
            Plan::MergeAppend { inputs, keys } => {
                let names = columns(plan);
                let keys: Vec<String> = keys.iter().map(|k| expr_text(&k.expr, &names)).collect();
                let children = inputs.iter().map(Child::Plan).collect();
                ("Merge Append".into(), vec![format!("Sort Key: {}", keys.join(", "))], children)
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
            Plan::CteScan(def) => (format!("CTE Scan on {}", def.name), vec![], vec![]),
            Plan::ProjectSet { input, .. } => ("ProjectSet".into(), vec![], vec![Child::Plan(input)]),
            Plan::Window { input, calls } => {
                let columns = &columns(input);
                let conditions: Vec<String> = calls
                    .iter()
                    .flat_map(|call| {
                        let args: Vec<String> = call.args.iter().map(|a| expr_text(a, columns)).collect();
                        let args = if args.is_empty() && call.name() == "count" { "*".into() } else { args.join(", ") };
                        let wfunc = format!("{}({args}) OVER (?)", call.name());
                        call.run_condition.iter().map(move |rc| {
                            let (op, arg) = (cmp_text(rc.op), expr_text(&rc.arg, columns));
                            match rc.wfunc_left {
                                true => format!("({wfunc} {op} {arg})"),
                                false => format!("({arg} {op} {wfunc})"),
                            }
                        })
                    })
                    .collect();
                let properties = match conditions.len() {
                    0 => vec![],
                    1 => vec![format!("Run Condition: {}", conditions[0])],
                    _ => vec![format!("Run Condition: ({})", conditions.join(" AND "))],
                };
                ("WindowAgg".into(), properties, vec![Child::Plan(input)])
            }
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
                let Some(subquery) = e.subquery() else { return };
                self.subplans += 1;
                self.lines.push(format!("{pad}SubPlan {}", self.subplans));
                ENCLOSING.with(|e| e.borrow_mut().push(Vec::new()));
                self.node(subquery, depth + 1, Vec::new(), Vec::new());
                ENCLOSING.with(|e| e.borrow_mut().pop());
            });
        }
        for child in children {
            self.child(child, depth);
        }
    }

    /// child prints a child of a node printed at a depth.
    fn child(&mut self, child: Child<'_>, depth: usize) {
        match child {
            Child::Plan(plan) => self.node(plan, depth + 1, Vec::new(), Vec::new()),
            Child::Held(name, plan) => {
                self.lines.push(format!("{}->  {name}", " ".repeat(6 * depth + 2)));
                self.node(plan, depth + 2, Vec::new(), Vec::new());
            }
            Child::Lookup(lookup) => self.lookup(&lookup, depth + 1),
            Child::Bitmap(bitmap) => self.bitmap(bitmap, depth + 1),
            Child::Lateral(plan, names) => {
                ENCLOSING.with(|e| e.borrow_mut().push(names));
                self.node(plan, depth + 1, Vec::new(), Vec::new());
                ENCLOSING.with(|e| e.borrow_mut().pop());
            }
            Child::Memoized(keys, binary, names, child) => {
                self.lines.push(format!("{}->  Memoize", " ".repeat(6 * depth + 2)));
                ENCLOSING.with(|e| e.borrow_mut().push(names));
                let properties = memoize_properties(keys, binary);
                ENCLOSING.with(|e| e.borrow_mut().pop());
                let pad = " ".repeat(6 * (depth + 1) + 2);
                self.lines.extend(properties.into_iter().map(|property| format!("{pad}{property}")));
                self.child(*child, depth + 1);
            }
        }
    }

    /// bitmap prints a node of the tree of index scans that a bitmap heap scan reads the keys of, as Postgres prints
    /// its bitmap index scans, BitmapAnds, and BitmapOrs.
    fn bitmap(&mut self, bitmap: &crate::indexscan::Bitmap, depth: usize) {
        let (name, properties, children) = match bitmap {
            crate::indexscan::Bitmap::Index(scan) => {
                let name = format!("Bitmap Index Scan on {}", crate::engine::quote_identifier(&scan.index_name()));
                (name, vec![index_columns(scan), index_ranges(scan)], &[][..])
            }
            crate::indexscan::Bitmap::And(children) => ("BitmapAnd".to_string(), Vec::new(), &children[..]),
            crate::indexscan::Bitmap::Or(children) => ("BitmapOr".to_string(), Vec::new(), &children[..]),
        };
        self.lines.push(format!("{}->  {name}", " ".repeat(6 * (depth - 1) + 2)));
        let pad = " ".repeat(6 * depth + 2);
        for property in properties {
            self.lines.push(format!("{pad}{property}"));
        }
        for child in children {
            self.bitmap(child, depth + 1);
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
        let right = match lookup.right {
            Plan::Project { input, .. } => &**input,
            other => other,
        };
        if let Plan::Filter { input, predicate } = right {
            self.lines.push(format!("{pad}Filter: {}", expr_text(predicate, &own_columns(input))));
        }
    }
}

/// memoize_properties returns the `Cache Key` and `Cache Mode` lines of a Memoize by keys that read its enclosing row.
fn memoize_properties(keys: &[Expr], binary: bool) -> Vec<String> {
    let keys: Vec<String> = keys.iter().map(|k| expr_text(k, &[])).collect();
    vec![
        format!("Cache Key: {}", keys.join(", ")),
        format!("Cache Mode: {}", if binary { "binary" } else { "logical" }),
    ]
}

/// index_ranges returns the `Index Ranges` line of a scan of an index, or the `Index Cond` line of the conditions
/// over its enclosing row that build its ranges each time it runs.
fn index_ranges(scan: &crate::indexscan::IndexScan) -> String {
    match &scan.parameterized {
        Some(cond) => {
            let names: Vec<String> = scan.table.columns.iter().map(|c| c.name.clone()).collect();
            format!("Index Cond: {}", expr_text(cond, &names))
        }
        None => format!("Index Ranges: {}", crate::ranges::ranges_text(&scan.ranges)),
    }
}

/// index_columns returns the `Index Columns` line of a scan of an index: its columns, each marked DESC when the index
/// orders it descending, then the predicate of a partial index.
fn index_columns(scan: &crate::indexscan::IndexScan) -> String {
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
    format!("Index Columns: {}", names.join(", "))
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
