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

//! Automatically updatable views: INSERT, UPDATE, and DELETE on a simple view, rewritten into the same statement on
//! the view's base relation, as Postgres' rewriter does.

use pg_query::protobuf::{
    AExpr, BoolExpr, BoolExprType, ColumnRef, DeleteStmt, InsertStmt, RangeVar, ResTarget, ReturningClause, SelectStmt,
    UpdateStmt, ViewCheckOption,
};
use pg_query::{Node, NodeEnum};

use crate::error::{PgError, Result, code};
use crate::expr::node_name;
use crate::query::Ctx;
use crate::{Column, Outcome};

/// Command is the kind of statement that changes a view.
#[derive(Clone, Copy, PartialEq)]
enum Command {
    Insert,
    Update,
    Delete,
}

/// View is an automatically updatable view: its name, the relation it selects from as its FROM item names it, each of
/// its columns with the expression that defines it and the base column it is when it is one, its WHERE condition, and
/// its check option.
struct View {
    name: String,
    base: RangeVar,
    columns: Vec<(String, Node, Option<String>)>,
    condition: Option<Node>,
    check: ViewCheckOption,
}

/// Rewritten is a change of a view rewritten onto the relation underneath it, with how many RETURNING columns the
/// original statement asked for, the views whose check option the rewritten statement returns a check column for, in
/// order after those columns, and whether the original statement had a RETURNING list.
pub struct Rewritten {
    pub statement: NodeEnum,
    pub returning: usize,
    pub checks: Vec<String>,
    pub returns: bool,
}

/// Updatability is which changes a view takes automatically, and which of its columns take new values.
pub struct Updatability {
    pub insertable: bool,
    pub updatable: bool,
    pub deletable: bool,
    pub columns: Vec<bool>,
}

/// column_ref returns a column reference node of names.
fn column_ref(names: &[&str]) -> Node {
    let fields =
        names.iter().map(|n| Node { node: Some(NodeEnum::String(pg_query::protobuf::String { sval: n.to_string() })) });
    Node { node: Some(NodeEnum::ColumnRef(ColumnRef { fields: fields.collect(), location: -1 })) }
}

/// star_ref returns a reference to every column of a relation.
fn star_ref(relation: &str) -> Node {
    let fields = vec![
        Node { node: Some(NodeEnum::String(pg_query::protobuf::String { sval: relation.to_string() })) },
        Node { node: Some(NodeEnum::AStar(pg_query::protobuf::AStar {})) },
    ];
    Node { node: Some(NodeEnum::ColumnRef(ColumnRef { fields, location: -1 })) }
}

/// res_target returns a target of an expression under a name.
fn res_target(name: &str, value: Node) -> Node {
    Node {
        node: Some(NodeEnum::ResTarget(Box::new(ResTarget {
            name: name.to_string(),
            val: Some(Box::new(value)),
            location: -1,
            ..Default::default()
        }))),
    }
}

/// and returns the conjunction of two conditions.
fn and(left: Node, right: Node) -> Node {
    Node {
        node: Some(NodeEnum::BoolExpr(Box::new(BoolExpr {
            boolop: BoolExprType::AndExpr as i32,
            args: vec![left, right],
            location: -1,
            ..Default::default()
        }))),
    }
}

/// is_star reports whether a column reference ends with `*`.
fn is_star(column: &ColumnRef) -> bool {
    matches!(column.fields.last().and_then(|f| f.node.as_ref()), Some(NodeEnum::AStar(_)))
}

/// not_updatable returns Postgres' error for a change of a view that is not automatically updatable.
fn not_updatable(command: Command, view: &str, detail: &str) -> PgError {
    let (message, hint) = match command {
        Command::Insert => (
            format!("cannot insert into view \"{view}\""),
            "To enable inserting into the view, provide an INSTEAD OF INSERT trigger or an unconditional ON INSERT DO \
             INSTEAD rule.",
        ),
        Command::Update => (
            format!("cannot update view \"{view}\""),
            "To enable updating the view, provide an INSTEAD OF UPDATE trigger or an unconditional ON UPDATE DO \
             INSTEAD rule.",
        ),
        Command::Delete => (
            format!("cannot delete from view \"{view}\""),
            "To enable deleting from the view, provide an INSTEAD OF DELETE trigger or an unconditional ON DELETE DO \
             INSTEAD rule.",
        ),
    };
    PgError {
        detail: Some(detail.to_string()),
        hint: Some(hint.to_string()),
        ..PgError::new(code::OBJECT_NOT_IN_PREREQUISITE_STATE, message)
    }
}

/// column_not_updatable returns Postgres' error for a change of a view column that is not a column of its base
/// relation.
fn column_not_updatable(command: Command, column: &str, view: &str) -> PgError {
    let verb = if command == Command::Insert { "insert into" } else { "update" };
    PgError {
        detail: Some("View columns that are not columns of their base relation are not updatable.".into()),
        ..PgError::new(code::FEATURE_NOT_SUPPORTED, format!("cannot {verb} column \"{column}\" of view \"{view}\""))
    }
}

/// target_name returns the name a select list item gives its column, as Postgres' FigureColname does for the forms a
/// view's columns take.
fn target_name(target: &ResTarget) -> String {
    if !target.name.is_empty() {
        return target.name.clone();
    }
    match target.val.as_deref().and_then(|v| v.node.as_ref()) {
        Some(NodeEnum::ColumnRef(column)) => column.fields.last().and_then(node_name).unwrap_or("?column?").to_string(),
        Some(NodeEnum::FuncCall(call)) => call.funcname.last().and_then(node_name).unwrap_or("?column?").to_string(),
        Some(NodeEnum::TypeCast(cast)) => match cast.arg.as_deref().and_then(|a| a.node.as_ref()) {
            Some(NodeEnum::ColumnRef(column)) => {
                column.fields.last().and_then(node_name).unwrap_or("?column?").to_string()
            }
            _ => cast
                .type_name
                .as_ref()
                .and_then(|t| t.names.last())
                .and_then(node_name)
                .unwrap_or("?column?")
                .to_string(),
        },
        _ => "?column?".to_string(),
    }
}

/// uses reports whether an expression calls a function for which `test` holds, or has a window.
fn uses(node: &Node, test: &dyn Fn(&pg_query::protobuf::FuncCall) -> bool) -> bool {
    let mut found = false;
    let statement =
        NodeEnum::ResTarget(Box::new(ResTarget { val: Some(Box::new(node.clone())), ..Default::default() }));
    for (child, ..) in statement.nodes() {
        if let pg_query::NodeRef::FuncCall(call) = child
            && test(call)
        {
            found = true;
        }
    }
    found
}

impl Ctx<'_> {
    /// relation_columns returns the names of a table's or view's columns.
    fn relation_columns(&mut self, relation: &RangeVar) -> Result<Vec<String>> {
        let select = SelectStmt {
            target_list: vec![res_target("", star_ref_unqualified())],
            from_clause: vec![Node { node: Some(NodeEnum::RangeVar(RangeVar { alias: None, ..relation.clone() })) }],
            op: pg_query::protobuf::SetOperation::SetopNone as i32,
            limit_option: pg_query::protobuf::LimitOption::Default as i32,
            ..Default::default()
        };
        let query = crate::plan::Planner { ctx: self, outer: Vec::new() }.plan_query(&select)?;
        Ok(query.columns.into_iter().map(|c| c.name).collect())
    }

    /// updatable_view reads a view and checks that it is automatically updatable for a command, failing as Postgres
    /// does when it is not.
    fn updatable_view(&mut self, schema: &str, name: &str, fragment: &str, command: Command) -> Result<View> {
        let parsed = pg_query::parse(fragment, 0).map_err(PgError::internal)?;
        let statement = parsed.protobuf.stmts.into_iter().next().and_then(|s| s.stmt).and_then(|s| s.node);
        let Some(NodeEnum::ViewStmt(stmt)) = statement else {
            return Err(PgError::internal(format!("a stored view that is not one: {fragment}")));
        };
        let Some(NodeEnum::SelectStmt(select)) = stmt.query.as_ref().and_then(|q| q.node.clone()) else {
            return Err(PgError::internal(format!("a stored view without a query: {fragment}")));
        };
        let fail = |detail: &str| Err(not_updatable(command, name, detail));
        if !select.distinct_clause.is_empty() {
            return fail("Views containing DISTINCT are not automatically updatable.");
        }
        if !select.group_clause.is_empty() || select.group_distinct {
            return fail("Views containing GROUP BY are not automatically updatable.");
        }
        if select.having_clause.is_some() {
            return fail("Views containing HAVING are not automatically updatable.");
        }
        if select.op != pg_query::protobuf::SetOperation::SetopNone as i32 {
            return fail("Views containing UNION, INTERSECT, or EXCEPT are not automatically updatable.");
        }
        if select.with_clause.is_some() {
            return fail("Views containing WITH are not automatically updatable.");
        }
        if select.limit_count.is_some() || select.limit_offset.is_some() {
            return fail("Views containing LIMIT or OFFSET are not automatically updatable.");
        }
        let values: Vec<Node> = select
            .target_list
            .iter()
            .filter_map(|t| match t.node.as_ref() {
                Some(NodeEnum::ResTarget(target)) => target.val.as_deref().cloned(),
                _ => None,
            })
            .collect();
        let aggregate = |call: &pg_query::protobuf::FuncCall| {
            let name = call.funcname.last().and_then(node_name).unwrap_or_default();
            call.over.is_none() && (call.agg_star || crate::functions::aggregate::exists(None, name))
        };
        if values.iter().any(|v| uses(v, &aggregate)) {
            return fail("Views that return aggregate functions are not automatically updatable.");
        }
        if values.iter().any(|v| uses(v, &|call| call.over.is_some())) {
            return fail("Views that return window functions are not automatically updatable.");
        }
        let set_returning = |call: &pg_query::protobuf::FuncCall| {
            crate::functions::returns_set(call.funcname.last().and_then(node_name).unwrap_or_default())
        };
        if values.iter().any(|v| uses(v, &set_returning)) {
            return fail("Views that return set-returning functions are not automatically updatable.");
        }
        let single = "Views that do not select from a single table or view are not automatically updatable.";
        let base = match select.from_clause.as_slice() {
            [item] => match item.node.as_ref() {
                Some(NodeEnum::RangeVar(relation)) => relation.clone(),
                Some(NodeEnum::RangeTableSample(_)) => {
                    return fail("Views containing TABLESAMPLE are not automatically updatable.");
                }
                _ => return fail(single),
            },
            _ => return fail(single),
        };
        let qualified = RangeVar { schemaname: schema.to_string(), ..base.clone() };
        let base = match base.schemaname.is_empty() && self.relation_columns(&qualified).is_ok() {
            true => qualified,
            false => base,
        };
        let base_columns = self.relation_columns(&base)?;
        let qualifier = base.alias.as_ref().map_or(base.relname.clone(), |a| a.aliasname.clone());
        let mut columns = Vec::new();
        for target in &select.target_list {
            let Some(NodeEnum::ResTarget(target)) = target.node.as_ref() else { continue };
            let value = target.val.as_deref().cloned().unwrap_or_default();
            if let Some(NodeEnum::ColumnRef(column)) = value.node.as_ref()
                && is_star(column)
            {
                for base_column in &base_columns {
                    columns.push((
                        base_column.clone(),
                        column_ref(&[&qualifier, base_column]),
                        Some(base_column.clone()),
                    ));
                }
                continue;
            }
            let base_column = match value.node.as_ref() {
                Some(NodeEnum::ColumnRef(column)) => {
                    let names: Vec<&str> = column.fields.iter().filter_map(node_name).collect();
                    match names.as_slice() {
                        [c] | [_, c] if base_columns.iter().any(|b| b == c) => Some(c.to_string()),
                        _ => None,
                    }
                }
                _ => None,
            };
            columns.push((target_name(target), value, base_column));
        }
        for (column, alias) in columns.iter_mut().zip(stmt.aliases.iter().filter_map(node_name)) {
            column.0 = alias.to_string();
        }
        if command != Command::Delete && columns.iter().all(|c| c.2.is_none()) {
            return fail("Views that have no updatable columns are not automatically updatable.");
        }
        let option = stmt.options.iter().find_map(|o| match o.node.as_ref() {
            Some(NodeEnum::DefElem(def)) if def.defname == "check_option" => {
                def.arg.as_deref().and_then(node_name).map(str::to_string)
            }
            _ => None,
        });
        let check = match (option.as_deref(), ViewCheckOption::try_from(stmt.with_check_option)) {
            (Some("local"), _) | (_, Ok(ViewCheckOption::LocalCheckOption)) => ViewCheckOption::LocalCheckOption,
            (Some("cascaded"), _) | (_, Ok(ViewCheckOption::CascadedCheckOption)) => {
                ViewCheckOption::CascadedCheckOption
            }
            _ => ViewCheckOption::NoCheckOption,
        };
        Ok(View { name: name.to_string(), base, columns, condition: select.where_clause.map(|w| *w), check })
    }

    /// view_updatability reports which changes a view takes automatically, through any views underneath it, and which
    /// of its columns take new values, as Postgres' pg_relation_is_updatable and pg_column_is_updatable do, with its
    /// check option as information_schema names it.
    pub fn view_updatability(&mut self, schema: &str, name: &str, fragment: &str) -> Updatability {
        let none = || Updatability { insertable: false, updatable: false, deletable: false, columns: Vec::new() };
        let Ok(view) = self.updatable_view(schema, name, fragment, Command::Delete) else { return none() };
        let mut columns: Vec<bool> = view.columns.iter().map(|c| c.2.is_some()).collect();
        if self.resolve_table(&view.base).is_err() {
            let Ok(Some((base_schema, base_fragment))) = self.find_view(&view.base.schemaname, &view.base.relname)
            else {
                return none();
            };
            let base = self.view_updatability(&base_schema, &view.base.relname, &base_fragment);
            if !base.deletable {
                return none();
            }
            let base_columns = self.relation_columns(&view.base).unwrap_or_default();
            for (updatable, (_, _, base_column)) in columns.iter_mut().zip(&view.columns) {
                let position = base_column.as_ref().and_then(|b| base_columns.iter().position(|c| c == b));
                *updatable &= position.is_some_and(|p| base.columns.get(p).copied().unwrap_or(false));
            }
        }
        let any = columns.iter().any(|&c| c);
        Updatability { insertable: any, updatable: any, deletable: true, columns }
    }

    /// view_check_option returns a view's check option as information_schema names it.
    pub fn view_check_option(&mut self, schema: &str, name: &str, fragment: &str) -> &'static str {
        match self.updatable_view(schema, name, fragment, Command::Delete).map(|v| v.check) {
            Ok(ViewCheckOption::LocalCheckOption) => "LOCAL",
            Ok(ViewCheckOption::CascadedCheckOption) => "CASCADED",
            _ => "NONE",
        }
    }

    /// rewrite_view_change rewrites an INSERT, UPDATE, or DELETE of a view, and of any views underneath it, onto the
    /// table the views select from, or returns None when the statement changes a table.
    pub fn rewrite_view_change(&mut self, node: &NodeEnum) -> Result<Option<Rewritten>> {
        let mut statement = node.clone();
        let mut rewritten: Option<Rewritten> = None;
        let mut cascaded = false;
        loop {
            let (relation, command) = match &statement {
                NodeEnum::InsertStmt(s) => (s.relation.clone(), Command::Insert),
                NodeEnum::UpdateStmt(s) => (s.relation.clone(), Command::Update),
                NodeEnum::DeleteStmt(s) => (s.relation.clone(), Command::Delete),
                _ => return Ok(None),
            };
            let Some(relation) = relation else { return Ok(rewritten) };
            if self.resolve_table(&relation).is_ok() {
                break;
            }
            let Some((schema, fragment)) = self.find_view(&relation.schemaname, &relation.relname)? else {
                break;
            };
            let privilege = match command {
                Command::Insert => "a",
                Command::Update => "w",
                Command::Delete => "d",
            };
            self.require_view(&schema, &relation.relname, privilege, -1)?;
            let view = self.updatable_view(&schema, &relation.relname, &fragment, command)?;
            let returns = !returning_list(&statement).is_empty();
            let mut qualifiers = vec![relation.relname.clone()];
            if let Some(alias) = &relation.alias {
                qualifiers = vec![alias.aliasname.clone()];
            }
            statement = rewrite_once(&statement, &view, &qualifiers, command)?;
            let state = rewritten.get_or_insert_with(|| Rewritten {
                statement: statement.clone(),
                returning: returning_list(&statement).len(),
                checks: Vec::new(),
                returns,
            });
            cascaded |= view.check == ViewCheckOption::CascadedCheckOption;
            let checked = view.check != ViewCheckOption::NoCheckOption || cascaded;
            if checked
                && command != Command::Delete
                && let Some(condition) = &view.condition
            {
                let check = Node {
                    node: Some(NodeEnum::CoalesceExpr(Box::new(pg_query::protobuf::CoalesceExpr {
                        args: vec![condition.clone(), bool_const(false)],
                        location: -1,
                        ..Default::default()
                    }))),
                };
                let index = state.returning + state.checks.len();
                returning_list_mut(&mut statement).insert(index, res_target("?check?", check));
                state.checks.push(view.name.clone());
            }
            state.statement = statement.clone();
        }
        if let Some(state) = &mut rewritten
            && !state.checks.is_empty()
            && let Some(relation) = target_relation(&state.statement)
        {
            let qualifier = relation.alias.as_ref().map_or(relation.relname.clone(), |a| a.aliasname.clone());
            returning_list_mut(&mut state.statement).push(res_target("", star_ref(&qualifier)));
        }
        Ok(rewritten)
    }

    /// run_view_change runs a rewritten change of a view, failing when a new row breaks a view's check option, and
    /// returns what the original statement returns.
    pub fn run_view_change(&mut self, rewritten: Rewritten) -> Result<Outcome> {
        let outcome = self.run(&rewritten.statement)?;
        let Outcome::Rows { columns, rows, tag } = outcome else { return Ok(outcome) };
        let checks = rewritten.checks.len();
        for row in &rows {
            for (i, view) in rewritten.checks.iter().enumerate() {
                if row.get(rewritten.returning + i) != Some(&crate::types::Value::Bool(true)) {
                    let values: Vec<String> = row[rewritten.returning + checks..]
                        .iter()
                        .map(|v| v.output().unwrap_or_else(|| "null".into()))
                        .collect();
                    return Err(PgError {
                        detail: Some(format!("Failing row contains ({}).", values.join(", "))),
                        ..PgError::new("44000", format!("new row violates check option for view \"{view}\""))
                    });
                }
            }
        }
        if !rewritten.returns {
            return Ok(Outcome::command(tag));
        }
        let columns: Vec<Column> = columns.into_iter().take(rewritten.returning).collect();
        let rows = rows.into_iter().map(|row| row.into_iter().take(rewritten.returning).collect()).collect();
        Ok(Outcome::Rows { columns, rows, tag })
    }
}

/// star_ref_unqualified returns a reference to every column.
fn star_ref_unqualified() -> Node {
    Node {
        node: Some(NodeEnum::ColumnRef(ColumnRef {
            fields: vec![Node { node: Some(NodeEnum::AStar(pg_query::protobuf::AStar {})) }],
            location: -1,
        })),
    }
}

/// bool_const returns a boolean constant.
fn bool_const(value: bool) -> Node {
    Node {
        node: Some(NodeEnum::AConst(pg_query::protobuf::AConst {
            val: Some(pg_query::protobuf::a_const::Val::Boolval(pg_query::protobuf::Boolean { boolval: value })),
            ..Default::default()
        })),
    }
}

/// target_relation returns the relation a change names.
fn target_relation(statement: &NodeEnum) -> Option<&RangeVar> {
    match statement {
        NodeEnum::InsertStmt(s) => s.relation.as_ref(),
        NodeEnum::UpdateStmt(s) => s.relation.as_ref(),
        NodeEnum::DeleteStmt(s) => s.relation.as_ref(),
        _ => None,
    }
}

/// returning_list returns a change's RETURNING list.
fn returning_list(statement: &NodeEnum) -> &[Node] {
    let clause = match statement {
        NodeEnum::InsertStmt(s) => &s.returning_clause,
        NodeEnum::UpdateStmt(s) => &s.returning_clause,
        NodeEnum::DeleteStmt(s) => &s.returning_clause,
        _ => &None,
    };
    clause.as_ref().map_or(&[], |c| &c.exprs)
}

/// returning_list_mut returns a change's RETURNING list for changing.
fn returning_list_mut(statement: &mut NodeEnum) -> &mut Vec<Node> {
    let clause = match statement {
        NodeEnum::InsertStmt(s) => &mut s.returning_clause,
        NodeEnum::UpdateStmt(s) => &mut s.returning_clause,
        NodeEnum::DeleteStmt(s) => &mut s.returning_clause,
        _ => unreachable!("only changes have RETURNING lists"),
    };
    &mut clause.get_or_insert_default().exprs
}

/// rewrite_once rewrites a change of a view onto the relation the view selects from, replacing references to the
/// view's columns, by the names in `qualifiers` or unqualified, with the expressions that define them.
fn rewrite_once(statement: &NodeEnum, view: &View, qualifiers: &[String], command: Command) -> Result<NodeEnum> {
    let substitute = |node: &mut Node| substitute(node, view, qualifiers);
    let returning = |list: &[Node]| -> Vec<Node> {
        let mut out = Vec::new();
        for item in list {
            if let Some(NodeEnum::ResTarget(target)) = item.node.as_ref()
                && let Some(NodeEnum::ColumnRef(column)) = target.val.as_deref().and_then(|v| v.node.as_ref())
                && is_star(column)
            {
                let names: Vec<&str> = column.fields.iter().filter_map(node_name).collect();
                if names.is_empty() || names.last().is_some_and(|n| qualifiers.iter().any(|q| q == n)) {
                    out.extend(view.columns.iter().map(|(name, value, _)| res_target(name, value.clone())));
                    continue;
                }
            }
            let mut item = item.clone();
            if let Some(NodeEnum::ResTarget(target)) = item.node.as_mut() {
                if target.name.is_empty()
                    && let Some(NodeEnum::ColumnRef(column)) = target.val.as_deref().and_then(|v| v.node.as_ref())
                {
                    target.name = column.fields.last().and_then(node_name).unwrap_or_default().to_string();
                }
                if let Some(value) = target.val.as_deref_mut() {
                    substitute(value);
                }
            }
            out.push(item);
        }
        out
    };
    let base_column = |name: &str| -> Result<String> {
        match view.columns.iter().find(|(n, _, _)| n == name) {
            Some((_, _, Some(base))) => Ok(base.clone()),
            Some(_) => Err(column_not_updatable(command, name, &view.name)),
            None => Err(PgError::new(
                code::UNDEFINED_COLUMN,
                format!("column \"{name}\" of relation \"{}\" does not exist", view.name),
            )),
        }
    };
    let condition = |own: Option<Box<Node>>| -> Option<Box<Node>> {
        let own = own.map(|mut w| {
            substitute(&mut w);
            *w
        });
        match (own, view.condition.clone()) {
            (Some(own), Some(view)) => Some(Box::new(and(view, own))),
            (own, view) => own.or(view).map(Box::new),
        }
    };
    Ok(match statement {
        NodeEnum::UpdateStmt(update) => {
            let mut targets = Vec::with_capacity(update.target_list.len());
            for item in &update.target_list {
                let mut item = item.clone();
                if let Some(NodeEnum::ResTarget(target)) = item.node.as_mut() {
                    target.name = base_column(&target.name)?;
                    if let Some(value) = target.val.as_deref_mut() {
                        substitute(value);
                    }
                }
                targets.push(item);
            }
            NodeEnum::UpdateStmt(Box::new(UpdateStmt {
                relation: Some(view.base.clone()),
                target_list: targets,
                where_clause: condition(update.where_clause.clone()),
                returning_clause: update
                    .returning_clause
                    .as_ref()
                    .map(|c| ReturningClause { exprs: returning(&c.exprs), ..c.clone() }),
                ..*update.clone()
            }))
        }
        NodeEnum::DeleteStmt(delete) => NodeEnum::DeleteStmt(Box::new(DeleteStmt {
            relation: Some(view.base.clone()),
            where_clause: condition(delete.where_clause.clone()),
            returning_clause: delete
                .returning_clause
                .as_ref()
                .map(|c| ReturningClause { exprs: returning(&c.exprs), ..c.clone() }),
            ..*delete.clone()
        })),
        NodeEnum::InsertStmt(insert) => {
            let given = match insert.select_stmt.as_deref().and_then(|s| s.node.as_ref()) {
                Some(NodeEnum::SelectStmt(select)) => match select.values_lists.first().and_then(|l| l.node.as_ref()) {
                    Some(NodeEnum::List(list)) => list.items.len(),
                    _ => select.target_list.len(),
                },
                _ => 0,
            };
            let names: Vec<String> = if insert.cols.is_empty() {
                view.columns.iter().take(given).map(|(n, _, _)| n.clone()).collect()
            } else {
                insert
                    .cols
                    .iter()
                    .filter_map(|c| match c.node.as_ref() {
                        Some(NodeEnum::ResTarget(target)) => Some(target.name.clone()),
                        _ => None,
                    })
                    .collect()
            };
            let mut cols = Vec::with_capacity(names.len());
            for (i, name) in names.iter().enumerate() {
                let indirection = match insert.cols.get(i).and_then(|c| c.node.as_ref()) {
                    Some(NodeEnum::ResTarget(target)) => target.indirection.clone(),
                    _ => Vec::new(),
                };
                cols.push(Node {
                    node: Some(NodeEnum::ResTarget(Box::new(ResTarget {
                        name: base_column(name)?,
                        indirection,
                        location: -1,
                        ..Default::default()
                    }))),
                });
            }
            let mut on_conflict = insert.on_conflict_clause.clone();
            if let Some(clause) = on_conflict.as_mut() {
                if let Some(infer) = clause.infer.as_mut() {
                    for element in &mut infer.index_elems {
                        if let Some(NodeEnum::IndexElem(elem)) = element.node.as_mut()
                            && !elem.name.is_empty()
                        {
                            elem.name = base_column(&elem.name)?;
                        }
                    }
                }
                for item in &mut clause.target_list {
                    if let Some(NodeEnum::ResTarget(target)) = item.node.as_mut() {
                        target.name = base_column(&target.name)?;
                        if let Some(value) = target.val.as_deref_mut() {
                            substitute(value);
                        }
                    }
                }
                if let Some(where_clause) = clause.where_clause.as_deref_mut() {
                    substitute(where_clause);
                }
            }
            NodeEnum::InsertStmt(Box::new(InsertStmt {
                relation: Some(view.base.clone()),
                cols,
                on_conflict_clause: on_conflict,
                returning_clause: insert
                    .returning_clause
                    .as_ref()
                    .map(|c| ReturningClause { exprs: returning(&c.exprs), ..c.clone() }),
                ..*insert.clone()
            }))
        }
        other => other.clone(),
    })
}

/// substitute replaces each reference to a view column in an expression with the expression that defines the column.
fn substitute(node: &mut Node, view: &View, qualifiers: &[String]) {
    let Some(inner) = node.node.as_mut() else { return };
    if let NodeEnum::ColumnRef(column) = inner {
        let names: Vec<&str> = column.fields.iter().filter_map(node_name).collect();
        let name = match names.as_slice() {
            [name] if column.fields.len() == 1 => Some(*name),
            [qualifier, name] if qualifiers.iter().any(|q| q == qualifier) => Some(*name),
            _ => None,
        };
        if let Some((_, value, _)) = name.and_then(|n| view.columns.iter().find(|(c, _, _)| c == n)) {
            *node = value.clone();
        }
        return;
    }
    let mut each = |child: &mut Node| substitute(child, view, qualifiers);
    let each_box = |child: &mut Option<Box<Node>>| {
        if let Some(child) = child.as_deref_mut() {
            substitute(child, view, qualifiers);
        }
    };
    match inner {
        NodeEnum::AExpr(e) => {
            let AExpr { lexpr, rexpr, .. } = &mut **e;
            each_box(lexpr);
            each_box(rexpr);
        }
        NodeEnum::BoolExpr(e) => e.args.iter_mut().for_each(&mut each),
        NodeEnum::FuncCall(call) => call.args.iter_mut().for_each(&mut each),
        NodeEnum::TypeCast(cast) => each_box(&mut cast.arg),
        NodeEnum::NullTest(test) => each_box(&mut test.arg),
        NodeEnum::BooleanTest(test) => each_box(&mut test.arg),
        NodeEnum::CollateClause(collate) => each_box(&mut collate.arg),
        NodeEnum::NamedArgExpr(arg) => each_box(&mut arg.arg),
        NodeEnum::AIndirection(indirection) => each_box(&mut indirection.arg),
        NodeEnum::CoalesceExpr(e) => e.args.iter_mut().for_each(&mut each),
        NodeEnum::MinMaxExpr(e) => e.args.iter_mut().for_each(&mut each),
        NodeEnum::RowExpr(e) => e.args.iter_mut().for_each(&mut each),
        NodeEnum::AArrayExpr(e) => e.elements.iter_mut().for_each(&mut each),
        NodeEnum::List(list) => list.items.iter_mut().for_each(&mut each),
        NodeEnum::SubLink(link) => each_box(&mut link.testexpr),
        NodeEnum::CaseExpr(case) => {
            each_box(&mut case.arg);
            each_box(&mut case.defresult);
            for when in &mut case.args {
                if let Some(NodeEnum::CaseWhen(when)) = when.node.as_mut() {
                    each_box(&mut when.expr);
                    each_box(&mut when.result);
                }
            }
        }
        _ => {}
    }
}
