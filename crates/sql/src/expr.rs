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

//! Expressions: binding the parser's expressions to typed expressions over a row, and evaluating them.

use std::cmp::Ordering;

use pg_query::protobuf::a_const::Val;
use pg_query::protobuf::{AExprKind, BoolExprType, NullTestType, SqlValueFunctionOp};
use pg_query::{Node, NodeEnum};

use crate::cast::{cast_value, type_display};
use crate::catalog::{ColumnType, resolve_type};
use crate::error::{PgError, Result, code};
use crate::functions;
use crate::functions::aggregate::AggCall;
use crate::numeric::Numeric;
use crate::oid;
use crate::plan::{Plan, Planner};
use crate::query::Ctx;
use crate::types::Value;

/// ScopeColumn is a column that expressions can refer to, from the table it belongs to.
#[derive(Clone, Debug)]
pub struct ScopeColumn {
    /// The name or alias of the table, empty for a column outside any table.
    pub table: String,
    pub name: String,
    pub ty: ColumnType,
    /// Whether the column is reachable only by its table's name, as a column that USING merged is.
    pub hidden: bool,
}

/// Scope is the columns of the rows that expressions evaluate over, in row order.
#[derive(Clone, Debug, Default)]
pub struct Scope {
    pub columns: Vec<ScopeColumn>,
}

/// ArithOp is an arithmetic operator.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum ArithOp {
    Add,
    Sub,
    Mul,
    Div,
    Mod,
}

/// CmpOp is a comparison operator.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum CmpOp {
    Eq,
    Ne,
    Lt,
    Le,
    Gt,
    Ge,
}

/// Expr is a bound expression.
#[derive(Clone, Debug, PartialEq)]
pub enum Expr {
    Const(Value),
    /// A column of the input row by position.
    Column(usize),
    /// A statement parameter by zero-based position.
    Param(usize),
    /// A cast to the type, which is explicit when written by the user.
    Cast(Box<Expr>, ColumnType, bool),
    /// Arithmetic with the result type.
    Arith(ArithOp, Box<Expr>, Box<Expr>, ColumnType),
    Neg(Box<Expr>, ColumnType),
    Compare(CmpOp, Box<Expr>, Box<Expr>),
    Concat(Box<Expr>, Box<Expr>),
    And(Box<Expr>, Box<Expr>),
    Or(Box<Expr>, Box<Expr>),
    Not(Box<Expr>),
    /// IS NULL, or IS NOT NULL when negated.
    IsNull(Box<Expr>, bool),
    /// A call of the built-in function at the index.
    Func(usize, Vec<Expr>),
    /// A column of an enclosing query's row, by how many queries out it is and its position.
    Outer(usize, usize),
    /// A column of a grouped query's input, which grouping must replace.
    InputColumn(usize),
    /// The result of a grouped query's aggregate call, which grouping replaces.
    AggRef(usize),
    /// The first argument that is not NULL.
    Coalesce(Vec<Expr>),
    /// The result of the first condition that holds, or the default.
    Case(Vec<(Expr, Expr)>, Box<Expr>),
    /// NULL when the comparison holds, and otherwise the first value.
    NullIf(Box<Expr>, Box<Expr>),
    /// The greatest or least argument that is not NULL.
    MinMax(bool, Vec<Expr>),
    /// IS DISTINCT FROM, or IS NOT DISTINCT FROM when negated.
    DistinctFrom(Box<Expr>, Box<Expr>, bool),
    /// IS TRUE, IS FALSE, or IS UNKNOWN, negated with NOT.
    BoolTest(Box<Expr>, Option<bool>, bool),
    /// Whether the subquery returns a row.
    Exists(Box<Plan>),
    /// The one value the subquery returns, or NULL without rows.
    Scalar(Box<Plan>),
    /// Whether the comparison holds for any, or for all, of the subquery's values.
    AnySubquery(Box<Expr>, Box<Plan>, bool),
    /// The subquery value a comparison of AnySubquery tests.
    SubqueryValue,
}

/// Bound is a bound expression with its type.
pub type Bound = (Expr, ColumnType);

/// Binder binds expressions over a scope inside the scopes of enclosing queries.
pub struct Binder<'b, 'a> {
    pub ctx: &'b mut Ctx<'a>,
    /// The scopes of the enclosing queries, innermost last, and then the current scope.
    pub scopes: Vec<Scope>,
    /// The aggregate calls of a grouped query, or None where aggregates aren't allowed.
    pub aggregates: Option<Vec<AggCall>>,
    /// The current scope's columns that were referred to, with the locations of the references.
    pub columns: Vec<(usize, i32)>,
}

impl<'b, 'a> Binder<'b, 'a> {
    /// new returns a binder over a scope.
    pub fn new(ctx: &'b mut Ctx<'a>, scope: Scope) -> Binder<'b, 'a> {
        Binder::with_scopes(ctx, vec![scope])
    }

    /// with_scopes returns a binder over the last scope inside the others.
    pub fn with_scopes(ctx: &'b mut Ctx<'a>, scopes: Vec<Scope>) -> Binder<'b, 'a> {
        Binder { ctx, scopes, aggregates: None, columns: Vec::new() }
    }

    /// compare binds a comparison of two bound expressions.
    pub fn compare(&mut self, op: &str, left: Bound, right: Bound, location: i32) -> Result<Bound> {
        self.binary(op, left, right, location)
    }
}

/// common_type returns the type that values of the types convert to, as Postgres chooses it for UNION, CASE, VALUES,
/// and the like, failing for types that do not match in the context.
pub fn common_type(types: &[(ColumnType, i32)], context: &str) -> Result<ColumnType> {
    let known: Vec<&(ColumnType, i32)> = types.iter().filter(|(t, _)| t.oid != oid::UNKNOWN).collect();
    let Some(&&(first, _)) = known.first() else { return Ok(typ(oid::TEXT)) };
    let mut result = first;
    for &&(ty, location) in &known[1..] {
        if ty.oid == result.oid {
            if ty.modifier != result.modifier {
                result.modifier = -1;
            }
            continue;
        }
        result = match (numeric_rank(result.oid), numeric_rank(ty.oid)) {
            (Some(a), Some(b)) => {
                if b > a {
                    typ(ty.oid)
                } else {
                    typ(result.oid)
                }
            }
            _ if is_string(result.oid) && is_string(ty.oid) => typ(oid::TEXT),
            _ => {
                return Err(PgError {
                    position: position(location),
                    ..PgError::new(
                        code::DATATYPE_MISMATCH,
                        format!(
                            "{context} types {} and {} cannot be matched",
                            type_display(result.oid),
                            type_display(ty.oid)
                        ),
                    )
                });
            }
        };
    }
    Ok(result)
}

/// typ returns a column type without a modifier.
pub fn typ(type_oid: u32) -> ColumnType {
    ColumnType { oid: type_oid, modifier: -1 }
}

/// numeric_rank orders the numeric types by how Postgres promotes them, or None for a type outside the category.
fn numeric_rank(type_oid: u32) -> Option<u8> {
    match type_oid {
        oid::INT2 => Some(0),
        oid::INT4 => Some(1),
        oid::INT8 => Some(2),
        oid::NUMERIC => Some(3),
        oid::FLOAT4 => Some(4),
        oid::FLOAT8 => Some(5),
        _ => None,
    }
}

/// is_string reports whether the type is in the string category, counting untyped literals.
fn is_string(type_oid: u32) -> bool {
    matches!(type_oid, oid::TEXT | oid::VARCHAR | oid::BPCHAR | oid::NAME | oid::UNKNOWN)
}

/// position returns a parser location as a 1-based error position.
pub fn position(location: i32) -> Option<u32> {
    u32::try_from(location).ok().map(|l| l + 1)
}

/// node_name returns the string of a String node.
pub fn node_name(node: &Node) -> Option<&str> {
    match node.node.as_ref() {
        Some(NodeEnum::String(s)) => Some(&s.sval),
        _ => None,
    }
}

/// type_name_parts returns a type name's parts, modifiers, and whether it names an array, for resolving it.
pub fn resolve_type_name(type_name: &pg_query::protobuf::TypeName) -> Result<ColumnType> {
    let names: Vec<String> = type_name.names.iter().filter_map(node_name).map(str::to_string).collect();
    let mut modifiers = Vec::new();
    for modifier in &type_name.typmods {
        match modifier.node.as_ref() {
            Some(NodeEnum::AConst(c)) => match &c.val {
                Some(Val::Ival(i)) => modifiers.push(i.ival),
                _ => {
                    return Err(PgError {
                        position: position(c.location),
                        ..PgError::new(code::SYNTAX_ERROR, "type modifiers must be simple constants or identifiers")
                    });
                }
            },
            _ => return Err(PgError::unsupported("this type modifier")),
        }
    }
    resolve_type(&names, &modifiers, !type_name.array_bounds.is_empty(), position(type_name.location))
}

impl<'b, 'a> Binder<'b, 'a> {
    /// bind binds an expression.
    pub fn bind(&mut self, node: &Node) -> Result<Bound> {
        let Some(node) = node.node.as_ref() else { return Err(PgError::internal("an empty expression")) };
        match node {
            NodeEnum::AConst(c) => constant(c),
            NodeEnum::ColumnRef(column) => self.column(column),
            NodeEnum::ParamRef(param) => {
                let index = param.number as usize - 1;
                if self.ctx.parameters.len() <= index {
                    self.ctx.parameters.resize(index + 1, 0);
                }
                Ok((
                    Expr::Param(index),
                    typ(match self.ctx.parameters[index] {
                        0 => oid::UNKNOWN,
                        t => t,
                    }),
                ))
            }
            NodeEnum::TypeCast(cast) => {
                let target = resolve_type_name(cast.type_name.as_ref().ok_or_else(|| PgError::internal("no type"))?)?;
                let arg = cast.arg.as_deref().ok_or_else(|| PgError::internal("no cast argument"))?;
                let bound = self.bind(arg)?;
                coerce(bound, target, true, arg_location(arg))
            }
            NodeEnum::AExpr(e) => self.a_expr(e),
            NodeEnum::BoolExpr(e) => {
                let mut args = Vec::with_capacity(e.args.len());
                for arg in &e.args {
                    let bound = self.bind(arg)?;
                    args.push(coerce(bound, typ(oid::BOOL), false, arg_location(arg))?.0);
                }
                let kind = BoolExprType::try_from(e.boolop).unwrap_or(BoolExprType::Undefined);
                let mut args = args.into_iter();
                let first = args.next().ok_or_else(|| PgError::internal("an empty boolean expression"))?;
                let expr = match kind {
                    BoolExprType::NotExpr => Expr::Not(Box::new(first)),
                    BoolExprType::AndExpr => args.fold(first, |l, r| Expr::And(Box::new(l), Box::new(r))),
                    _ => args.fold(first, |l, r| Expr::Or(Box::new(l), Box::new(r))),
                };
                Ok((expr, typ(oid::BOOL)))
            }
            NodeEnum::FuncCall(call) => self.func_call(call),
            NodeEnum::SqlvalueFunction(f) => {
                let name = match SqlValueFunctionOp::try_from(f.op) {
                    Ok(SqlValueFunctionOp::SvfopCurrentUser | SqlValueFunctionOp::SvfopCurrentRole) => "current_user",
                    Ok(SqlValueFunctionOp::SvfopUser) => "current_user",
                    Ok(SqlValueFunctionOp::SvfopSessionUser) => "session_user",
                    Ok(SqlValueFunctionOp::SvfopCurrentCatalog) => "current_database",
                    Ok(SqlValueFunctionOp::SvfopCurrentSchema) => "current_schema",
                    _ => return Err(PgError::unsupported("this SQL value function")),
                };
                let resolved = functions::resolve(name, &[], f.location)?;
                Ok((Expr::Func(resolved.index, Vec::new()), typ(resolved.ret)))
            }
            NodeEnum::CoalesceExpr(c) => {
                let (args, ty) = self.common_args(&c.args, "COALESCE")?;
                Ok((Expr::Coalesce(args), ty))
            }
            NodeEnum::MinMaxExpr(m) => {
                let greatest = m.op == pg_query::protobuf::MinMaxOp::IsGreatest as i32;
                let (args, ty) = self.common_args(&m.args, if greatest { "GREATEST" } else { "LEAST" })?;
                Ok((Expr::MinMax(greatest, args), ty))
            }
            NodeEnum::CaseExpr(c) => self.case(c),
            NodeEnum::BooleanTest(test) => {
                let arg = test.arg.as_deref().ok_or_else(|| PgError::internal("no boolean test argument"))?;
                let bound = self.bind(arg)?;
                let expr = coerce(bound, typ(oid::BOOL), false, arg_location(arg))?.0;
                use pg_query::protobuf::BoolTestType as T;
                let (value, negated) = match T::try_from(test.booltesttype) {
                    Ok(T::IsTrue) => (Some(true), false),
                    Ok(T::IsNotTrue) => (Some(true), true),
                    Ok(T::IsFalse) => (Some(false), false),
                    Ok(T::IsNotFalse) => (Some(false), true),
                    Ok(T::IsUnknown) => (None, false),
                    _ => (None, true),
                };
                Ok((Expr::BoolTest(Box::new(expr), value, negated), typ(oid::BOOL)))
            }
            NodeEnum::SubLink(link) => self.sublink(link),
            NodeEnum::NullTest(test) => {
                let arg = test.arg.as_deref().ok_or_else(|| PgError::internal("no null test argument"))?;
                let (expr, _) = self.bind(arg)?;
                let negated = NullTestType::try_from(test.nulltesttype) == Ok(NullTestType::IsNotNull);
                Ok((Expr::IsNull(Box::new(expr), negated), typ(oid::BOOL)))
            }
            _ => Err(PgError::unsupported(format!("the expression {}", node_kind(node)))),
        }
    }

    /// column binds a column reference.
    fn column(&mut self, column: &pg_query::protobuf::ColumnRef) -> Result<Bound> {
        let names: Vec<&str> = column.fields.iter().filter_map(node_name).collect();
        if names.len() != column.fields.len() {
            return Err(PgError::unsupported("this column reference"));
        }
        let (table, name) = match names.as_slice() {
            [name] => (None, *name),
            [table, name] => (Some(*table), *name),
            [_, table, name] => (Some(*table), *name),
            _ => return Err(PgError::unsupported("this column reference")),
        };
        let full = names.join(".");
        let depth_count = self.scopes.len();
        for depth in 0..depth_count {
            let scope = &self.scopes[depth_count - 1 - depth];
            let mut found = scope.columns.iter().enumerate().filter(|(_, c)| {
                c.name == name
                    && match table {
                        Some(t) => c.table == t,
                        None => !c.hidden,
                    }
            });
            match (found.next(), found.next()) {
                (Some((i, c)), None) => {
                    let ty = c.ty;
                    if depth == 0 {
                        self.columns.push((i, column.location));
                        return Ok((Expr::Column(i), ty));
                    }
                    return Ok((Expr::Outer(depth, i), ty));
                }
                (Some(_), Some(_)) => {
                    return Err(PgError {
                        position: position(column.location),
                        ..PgError::new(code::AMBIGUOUS_COLUMN, format!("column reference \"{full}\" is ambiguous"))
                    });
                }
                _ => {}
            }
        }
        if let Some(table) = table
            && !self.scopes.iter().any(|s| s.columns.iter().any(|c| c.table == table))
        {
            return Err(PgError {
                position: position(column.location),
                ..PgError::new(code::UNDEFINED_TABLE, format!("missing FROM-clause entry for table \"{table}\""))
            });
        }
        Err(PgError {
            position: position(column.location),
            ..PgError::new(code::UNDEFINED_COLUMN, format!("column \"{full}\" does not exist"))
        })
    }
}

impl<'b, 'a> Binder<'b, 'a> {
    /// func_call binds a call of a built-in function.
    fn func_call(&mut self, call: &pg_query::protobuf::FuncCall) -> Result<Bound> {
        let names: Vec<&str> = call.funcname.iter().filter_map(node_name).collect();
        let name = match names.as_slice() {
            [name] => *name,
            ["pg_catalog", name] => *name,
            _ => {
                return Err(PgError {
                    position: position(call.location),
                    ..PgError::new(code::UNDEFINED_FUNCTION, format!("function {}() does not exist", names.join(".")))
                });
            }
        };
        if call.over.is_some() {
            return Err(PgError::unsupported("window functions"));
        }
        if call.agg_star || functions::aggregate::exists(name) {
            return self.aggregate_call(name, call);
        }
        if call.agg_distinct || call.agg_filter.is_some() || !call.agg_order.is_empty() {
            return Err(PgError {
                position: position(call.location),
                ..PgError::new(
                    code::WRONG_OBJECT_TYPE,
                    format!("DISTINCT specified, but {name} is not an aggregate function"),
                )
            });
        }
        let mut bound = Vec::with_capacity(call.args.len());
        for arg in &call.args {
            bound.push(self.bind(arg)?);
        }
        let types: Vec<u32> = bound.iter().map(|(_, t)| t.oid).collect();
        let resolved = functions::resolve(name, &types, call.location)?;
        let mut args = Vec::with_capacity(bound.len());
        for (((expr, ty), &target), node) in bound.into_iter().zip(&resolved.arg_types).zip(&call.args) {
            if let Expr::Param(i) = expr
                && self.ctx.parameters[i] == 0
            {
                self.ctx.parameters[i] = target;
            }
            if target == functions::ANY {
                args.push(expr);
            } else {
                args.push(coerce((expr, ty), typ(target), false, arg_location(node))?.0);
            }
        }
        Ok((Expr::Func(resolved.index, args), typ(resolved.ret)))
    }
}

impl<'b, 'a> Binder<'b, 'a> {
    /// aggregate_call binds a call of an aggregate in a grouped query, whose arguments are over the input rows.
    fn aggregate_call(&mut self, name: &str, call: &pg_query::protobuf::FuncCall) -> Result<Bound> {
        let Some(mut aggregates) = self.aggregates.take() else {
            return Err(PgError {
                position: position(call.location),
                ..PgError::new(code::GROUPING_ERROR, "aggregate function calls cannot be nested")
            });
        };
        let result = (|| -> Result<(AggCall, u32)> {
            let mut bound = Vec::with_capacity(call.args.len());
            for arg in &call.args {
                bound.push(self.bind(arg)?);
            }
            let types: Vec<u32> = bound.iter().map(|(_, t)| t.oid).collect();
            let (index, arg_types, ret) = functions::aggregate::resolve(name, &types, call.location)?;
            let mut args = Vec::with_capacity(bound.len());
            for ((bound, &target), node) in bound.into_iter().zip(&arg_types).zip(&call.args) {
                args.push(coerce(bound, typ(target), false, arg_location(node))?.0);
            }
            let filter = match call.agg_filter.as_deref() {
                Some(node) => Some(coerce(self.bind(node)?, typ(oid::BOOL), false, -1)?.0),
                None => None,
            };
            let mut order = Vec::new();
            for sort in &call.agg_order {
                let Some(NodeEnum::SortBy(sort)) = sort.node.as_ref() else { continue };
                let node = sort.node.as_deref().ok_or_else(|| PgError::internal("ORDER BY without a key"))?;
                let descending = pg_query::protobuf::SortByDir::try_from(sort.sortby_dir)
                    == Ok(pg_query::protobuf::SortByDir::SortbyDesc);
                let nulls_first = match pg_query::protobuf::SortByNulls::try_from(sort.sortby_nulls) {
                    Ok(pg_query::protobuf::SortByNulls::SortbyNullsFirst) => true,
                    Ok(pg_query::protobuf::SortByNulls::SortbyNullsLast) => false,
                    _ => descending,
                };
                order.push((self.bind(node)?.0, descending, nulls_first));
            }
            Ok((AggCall { index, args, distinct: call.agg_distinct, filter, order, ret }, ret))
        })();
        let (agg, ret) = match result {
            Ok(r) => r,
            Err(err) => {
                self.aggregates = Some(aggregates);
                return Err(err);
            }
        };
        let k = aggregates.len();
        aggregates.push(agg);
        self.aggregates = Some(aggregates);
        Ok((Expr::AggRef(k), typ(ret)))
    }
}

/// operand returns an operand of an operator expression.
fn operand(side: &Option<Box<Node>>) -> Result<&Node> {
    side.as_deref().ok_or_else(|| PgError::internal("no operand"))
}

impl<'b, 'a> Binder<'b, 'a> {
    /// a_expr binds an operator expression.
    fn a_expr(&mut self, e: &pg_query::protobuf::AExpr) -> Result<Bound> {
        let kind = AExprKind::try_from(e.kind).unwrap_or(AExprKind::Undefined);
        let op = e.name.iter().filter_map(node_name).next_back().unwrap_or_default().to_string();
        match kind {
            AExprKind::AexprOp => {
                if e.lexpr.is_none() {
                    let right = self.bind(operand(&e.rexpr)?)?;
                    return unary(&op, right, e.location);
                }
                let left = self.bind(operand(&e.lexpr)?)?;
                let right = self.bind(operand(&e.rexpr)?)?;
                self.binary(&op, left, right, e.location)
            }
            AExprKind::AexprIn => {
                let left_node = operand(&e.lexpr)?;
                let Some(NodeEnum::List(list)) = operand(&e.rexpr)?.node.as_ref() else {
                    return Err(PgError::unsupported("IN with a subquery"));
                };
                let (cmp, join): (&str, fn(Expr, Expr) -> Expr) = if op == "<>" {
                    ("<>", |l, r| Expr::And(Box::new(l), Box::new(r)))
                } else {
                    ("=", |l, r| Expr::Or(Box::new(l), Box::new(r)))
                };
                let mut result: Option<Expr> = None;
                for item in &list.items {
                    let left = self.bind(left_node)?;
                    let right = self.bind(item)?;
                    let (test, _) = self.binary(cmp, left, right, e.location)?;
                    result = Some(match result {
                        Some(previous) => join(previous, test),
                        None => test,
                    });
                }
                Ok((result.ok_or_else(|| PgError::internal("an empty IN list"))?, typ(oid::BOOL)))
            }
            AExprKind::AexprBetween
            | AExprKind::AexprNotBetween
            | AExprKind::AexprBetweenSym
            | AExprKind::AexprNotBetweenSym => {
                let value = operand(&e.lexpr)?;
                let Some(NodeEnum::List(bounds)) = operand(&e.rexpr)?.node.as_ref() else {
                    return Err(PgError::internal("BETWEEN without bounds"));
                };
                let [low, high] = bounds.items.as_slice() else { return Err(PgError::internal("BETWEEN bounds")) };
                let symmetric = matches!(kind, AExprKind::AexprBetweenSym | AExprKind::AexprNotBetweenSym);
                let mut expr = self.between(value, low, high, e.location)?;
                if symmetric {
                    expr = Expr::Or(Box::new(expr), Box::new(self.between(value, high, low, e.location)?));
                }
                if matches!(kind, AExprKind::AexprNotBetween | AExprKind::AexprNotBetweenSym) {
                    expr = Expr::Not(Box::new(expr));
                }
                Ok((expr, typ(oid::BOOL)))
            }
            AExprKind::AexprNullif => {
                let left = self.bind(operand(&e.lexpr)?)?;
                let ty = left.1;
                let right = self.bind(operand(&e.rexpr)?)?;
                let left_expr = left.0.clone();
                let (test, _) = self.binary("=", left, right, e.location)?;
                Ok((Expr::NullIf(Box::new(left_expr), Box::new(test)), ty))
            }
            AExprKind::AexprDistinct | AExprKind::AexprNotDistinct => {
                let left = self.bind(operand(&e.lexpr)?)?;
                let right = self.bind(operand(&e.rexpr)?)?;
                let (test, _) = self.binary("=", left, right, e.location)?;
                let Expr::Compare(_, l, r) = test else { return Err(PgError::internal("a distinct test")) };
                Ok((Expr::DistinctFrom(l, r, kind == AExprKind::AexprNotDistinct), typ(oid::BOOL)))
            }
            AExprKind::AexprOpAny | AExprKind::AexprOpAll => Err(PgError::unsupported("ANY and ALL with arrays")),
            _ => Err(PgError::unsupported(format!("the operator expression {kind:?}"))),
        }
    }

    /// common_args binds arguments that must share a type, converting them to their common type.
    fn common_args(&mut self, nodes: &[Node], context: &str) -> Result<(Vec<Expr>, ColumnType)> {
        let mut bound = Vec::with_capacity(nodes.len());
        for node in nodes {
            bound.push((self.bind(node)?, arg_location(node)));
        }
        let types: Vec<(ColumnType, i32)> = bound.iter().map(|((_, t), l)| (*t, *l)).collect();
        let ty = common_type(&types, context)?;
        let args = bound.into_iter().map(|(b, l)| coerce(b, ty, false, l).map(|b| b.0)).collect::<Result<_>>()?;
        Ok((args, ty))
    }

    /// case binds a CASE expression, whose results take their common type.
    fn case(&mut self, c: &pg_query::protobuf::CaseExpr) -> Result<Bound> {
        let mut conditions = Vec::new();
        let mut results = Vec::new();
        for when in &c.args {
            let Some(NodeEnum::CaseWhen(when)) = when.node.as_ref() else { continue };
            let condition = when.expr.as_deref().ok_or_else(|| PgError::internal("WHEN without a condition"))?;
            let test = match c.arg.as_deref() {
                Some(operand) => {
                    let left = self.bind(operand)?;
                    let right = self.bind(condition)?;
                    self.binary("=", left, right, when.location)?.0
                }
                None => coerce(self.bind(condition)?, typ(oid::BOOL), false, arg_location(condition))?.0,
            };
            conditions.push(test);
            let result = when.result.as_deref().ok_or_else(|| PgError::internal("WHEN without a result"))?;
            results.push((self.bind(result)?, arg_location(result)));
        }
        let default = match c.defresult.as_deref() {
            Some(node) => Some((self.bind(node)?, arg_location(node))),
            None => None,
        };
        let mut types: Vec<(ColumnType, i32)> = results.iter().map(|((_, t), l)| (*t, *l)).collect();
        if let Some(((_, t), l)) = &default {
            types.push((*t, *l));
        }
        let ty = common_type(&types, "CASE")?;
        let mut whens = Vec::new();
        for (condition, (result, location)) in conditions.into_iter().zip(results) {
            whens.push((condition, coerce(result, ty, false, location)?.0));
        }
        let otherwise = match default {
            Some((bound, location)) => coerce(bound, ty, false, location)?.0,
            None => Expr::Const(Value::Null),
        };
        Ok((Expr::Case(whens, Box::new(otherwise)), ty))
    }

    /// sublink binds a subquery expression.
    fn sublink(&mut self, link: &pg_query::protobuf::SubLink) -> Result<Bound> {
        use pg_query::protobuf::SubLinkType as T;
        let Some(NodeEnum::SelectStmt(select)) = link.subselect.as_deref().and_then(|n| n.node.as_ref()) else {
            return Err(PgError::internal("a subquery without a SELECT"));
        };
        let query = Planner { ctx: &mut *self.ctx, outer: self.scopes.clone() }.plan_query(select)?;
        let kind = T::try_from(link.sub_link_type).unwrap_or(T::ExprSublink);
        match kind {
            T::ExistsSublink => Ok((Expr::Exists(Box::new(query.plan)), typ(oid::BOOL))),
            T::ExprSublink => {
                if query.columns.len() != 1 {
                    return Err(PgError {
                        position: position(link.location),
                        ..PgError::new(code::SYNTAX_ERROR, "subquery must return only one column")
                    });
                }
                Ok((Expr::Scalar(Box::new(query.plan)), query.types[0]))
            }
            T::AnySublink | T::AllSublink => {
                if query.columns.len() != 1 {
                    let message = if query.columns.len() > 1 {
                        "subquery has too many columns"
                    } else {
                        "subquery has too few columns"
                    };
                    return Err(PgError {
                        position: position(link.location),
                        ..PgError::new(code::SYNTAX_ERROR, message)
                    });
                }
                let test =
                    link.testexpr.as_deref().ok_or_else(|| PgError::internal("a subquery test without a value"))?;
                let op = link.oper_name.iter().filter_map(node_name).next_back().unwrap_or("=").to_string();
                let left = self.bind(test)?;
                let (comparison, _) = self.binary(&op, left, (Expr::SubqueryValue, query.types[0]), link.location)?;
                Ok((
                    Expr::AnySubquery(Box::new(comparison), Box::new(query.plan), kind == T::AllSublink),
                    typ(oid::BOOL),
                ))
            }
            _ => Err(PgError::unsupported("this kind of subquery")),
        }
    }

    /// between binds `value >= low AND value <= high`.
    fn between(&mut self, value: &Node, low: &Node, high: &Node, location: i32) -> Result<Expr> {
        let (value_bound, low) = (self.bind(value)?, self.bind(low)?);
        let (lower, _) = self.binary(">=", value_bound, low, location)?;
        let (value_bound, high) = (self.bind(value)?, self.bind(high)?);
        let (upper, _) = self.binary("<=", value_bound, high, location)?;
        Ok(Expr::And(Box::new(lower), Box::new(upper)))
    }

    /// binary binds a binary operator, resolving its operand types as Postgres does for the built-in operators.
    fn binary(&mut self, op: &str, left: Bound, right: Bound, location: i32) -> Result<Bound> {
        let (lt, rt) = (left.1.oid, right.1.oid);
        let missing = || PgError {
            position: position(location),
            hint: Some(
                "No operator matches the given name and argument types. You might need to add explicit type casts."
                    .into(),
            ),
            ..PgError::new(
                code::UNDEFINED_FUNCTION,
                format!("operator does not exist: {} {op} {}", type_display(lt), type_display(rt)),
            )
        };
        // An untyped operand takes the other operand's type, and two untyped operands are text.
        let domain = match (lt == oid::UNKNOWN, rt == oid::UNKNOWN) {
            (true, true) => typ(oid::TEXT),
            (true, false) => right.1,
            (false, true) => left.1,
            (false, false) => match (numeric_rank(lt), numeric_rank(rt)) {
                (Some(l), Some(r)) if l <= 2 && r <= 2 => {
                    if l >= r {
                        left.1
                    } else {
                        right.1
                    }
                }
                (Some(4), Some(4)) => typ(oid::FLOAT4),
                (Some(l), Some(r)) if l >= 4 || r >= 4 => typ(oid::FLOAT8),
                (Some(_), Some(_)) => typ(oid::NUMERIC),
                _ if is_string(lt) && is_string(rt) => typ(oid::TEXT),
                _ if lt == rt => left.1,
                _ => return Err(missing()),
            },
        };
        let domain = ColumnType { modifier: -1, ..domain };
        let infer = |binder: &mut Self, bound: &Bound, ty: ColumnType| {
            if let Expr::Param(i) = bound.0
                && binder.ctx.parameters[i] == 0
            {
                binder.ctx.parameters[i] = ty.oid;
            }
        };
        infer(self, &left, domain);
        infer(self, &right, domain);
        let cmp = match op {
            "=" => Some(CmpOp::Eq),
            "<>" | "!=" => Some(CmpOp::Ne),
            "<" => Some(CmpOp::Lt),
            "<=" => Some(CmpOp::Le),
            ">" => Some(CmpOp::Gt),
            ">=" => Some(CmpOp::Ge),
            _ => None,
        };
        let is_int = |t: u32| matches!(t, oid::INT2 | oid::INT4 | oid::INT8);
        // Integer operands keep their own widths, since the integer operators take mixed widths.
        let operand_type = |t: ColumnType| if is_int(domain.oid) && is_int(t.oid) { t } else { domain };
        let left_type = operand_type(left.1);
        let right_type = operand_type(right.1);
        let left = coerce(left, left_type, false, location)?.0;
        let right = coerce(right, right_type, false, location)?.0;
        if let Some(cmp) = cmp {
            return Ok((Expr::Compare(cmp, Box::new(left), Box::new(right)), typ(oid::BOOL)));
        }
        if op == "||" && is_string(domain.oid) {
            return Ok((Expr::Concat(Box::new(left), Box::new(right)), typ(oid::TEXT)));
        }
        let arith = match op {
            "+" => ArithOp::Add,
            "-" => ArithOp::Sub,
            "*" => ArithOp::Mul,
            "/" => ArithOp::Div,
            "%" if numeric_rank(domain.oid).is_some_and(|r| r <= 3) => ArithOp::Mod,
            _ => return Err(missing()),
        };
        if numeric_rank(domain.oid).is_none() {
            return Err(missing());
        }
        Ok((Expr::Arith(arith, Box::new(left), Box::new(right), domain), domain))
    }
}

/// unary binds a prefix operator.
fn unary(op: &str, (expr, ty): Bound, location: i32) -> Result<Bound> {
    let ty = if ty.oid == oid::UNKNOWN { typ(oid::FLOAT8) } else { ty };
    match op {
        "-" if numeric_rank(ty.oid).is_some() => {
            Ok((Expr::Neg(Box::new(coerce((expr, ty), ty, false, location)?.0), ty), ty))
        }
        "+" if numeric_rank(ty.oid).is_some() => Ok((expr, ty)),
        _ => Err(PgError {
            position: position(location),
            hint: Some(
                "No operator matches the given name and argument type. You might need to add an explicit type cast."
                    .into(),
            ),
            ..PgError::new(code::UNDEFINED_FUNCTION, format!("operator does not exist: {op} {}", type_display(ty.oid)))
        }),
    }
}

/// arg_location returns the location of an expression node, or -1.
pub fn arg_location(node: &Node) -> i32 {
    match node.node.as_ref() {
        Some(NodeEnum::AConst(c)) => c.location,
        Some(NodeEnum::ColumnRef(c)) => c.location,
        Some(NodeEnum::AExpr(e)) => e.location,
        Some(NodeEnum::TypeCast(c)) => c.location,
        Some(NodeEnum::ParamRef(p)) => p.location,
        Some(NodeEnum::FuncCall(f)) => f.location,
        _ => -1,
    }
}

/// node_kind names an expression node for errors.
fn node_kind(node: &NodeEnum) -> String {
    let debug = format!("{node:?}");
    debug.split('(').next().unwrap_or_default().to_string()
}

/// constant binds a constant.
fn constant(c: &pg_query::protobuf::AConst) -> Result<Bound> {
    Ok(match &c.val {
        _ if c.isnull => (Expr::Const(Value::Null), typ(oid::UNKNOWN)),
        Some(Val::Ival(i)) => (Expr::Const(Value::Int4(i.ival)), typ(oid::INT4)),
        Some(Val::Fval(f)) => {
            let text = &f.fval;
            match text.parse::<i64>() {
                Ok(i) if !text.contains(['.', 'e', 'E']) => (Expr::Const(Value::Int8(i)), typ(oid::INT8)),
                _ => (Expr::Const(Value::Numeric(Numeric::parse(text)?)), typ(oid::NUMERIC)),
            }
        }
        Some(Val::Sval(s)) => (Expr::Const(Value::Text(s.sval.clone())), typ(oid::UNKNOWN)),
        Some(Val::Boolval(b)) => (Expr::Const(Value::Bool(b.boolval)), typ(oid::BOOL)),
        _ => return Err(PgError::unsupported("this constant")),
    })
}

/// coerce converts a bound expression to the type, folding untyped constants by reading them as the type, and
/// rejecting implicit conversions Postgres does not allow.
pub fn coerce((expr, from): Bound, to: ColumnType, explicit: bool, location: i32) -> Result<Bound> {
    if from == to {
        return Ok((expr, to));
    }
    if from.oid == oid::UNKNOWN {
        if let Expr::Const(value) = &expr {
            let value = match value {
                Value::Text(text) => cast_value(
                    crate::cast::input(text, to.oid).map_err(|err| PgError { position: position(location), ..err })?,
                    to,
                    explicit,
                )?,
                other => other.clone(),
            };
            return Ok((Expr::Const(value), to));
        }
        return Ok((Expr::Cast(Box::new(expr), to, explicit), to));
    }
    if from.oid == to.oid && !explicit && to.modifier == -1 {
        return Ok((expr, to));
    }
    let numeric = numeric_rank(from.oid).zip(numeric_rank(to.oid));
    let allowed = explicit
        || from.oid == to.oid
        || numeric.is_some_and(|(f, t)| f <= t)
        || (is_string(from.oid) && is_string(to.oid));
    if !allowed {
        return Err(PgError {
            position: position(location),
            ..PgError::new(
                code::DATATYPE_MISMATCH,
                format!("cannot cast type {} to {}", type_display(from.oid), type_display(to.oid)),
            )
        });
    }
    if let Expr::Const(value) = &expr {
        return Ok((Expr::Const(cast_value(value.clone(), to, explicit)?), to));
    }
    Ok((Expr::Cast(Box::new(expr), to, explicit), to))
}

/// assign converts a bound expression to a column's type as an assignment does, which also allows numeric narrowing
/// and conversions to text, and names the column in its error.
pub fn assign(bound: Bound, to: ColumnType, column: &str, location: i32) -> Result<Bound> {
    let from = bound.1.oid;
    let allowed = from == oid::UNKNOWN
        || from == to.oid
        || (numeric_rank(from).is_some() && numeric_rank(to.oid).is_some())
        || is_string(to.oid);
    if !allowed {
        return Err(PgError {
            position: position(location),
            hint: Some("You will need to rewrite or cast the expression.".into()),
            ..PgError::new(
                code::DATATYPE_MISMATCH,
                format!(
                    "column \"{column}\" is of type {} but expression is of type {}",
                    type_display(to.oid),
                    type_display(from)
                ),
            )
        });
    }
    if from == to.oid && to.modifier == -1 {
        return Ok((bound.0, to));
    }
    if let Expr::Const(value) = &bound.0 {
        let value = match value {
            Value::Text(text) if from == oid::UNKNOWN => {
                crate::cast::input(text, to.oid).map_err(|err| PgError { position: position(location), ..err })?
            }
            other => other.clone(),
        };
        return Ok((Expr::Const(cast_value(value, to, false)?), to));
    }
    Ok((Expr::Cast(Box::new(bound.0), to, false), to))
}

/// figure_name returns the name Postgres gives a result column for an expression.
pub fn figure_name(node: &Node) -> String {
    match node.node.as_ref() {
        Some(NodeEnum::ColumnRef(c)) => {
            c.fields.iter().filter_map(node_name).next_back().unwrap_or("?column?").to_string()
        }
        Some(NodeEnum::FuncCall(f)) => {
            f.funcname.iter().filter_map(node_name).next_back().unwrap_or("?column?").to_string()
        }
        Some(NodeEnum::TypeCast(cast)) => match cast.arg.as_deref().map(figure_name) {
            Some(name) if name != "?column?" => name,
            _ => cast
                .type_name
                .as_ref()
                .and_then(|t| t.names.iter().filter_map(node_name).next_back().map(str::to_string))
                .unwrap_or_else(|| "?column?".into()),
        },
        Some(NodeEnum::CaseExpr(_)) => "case".into(),
        Some(NodeEnum::CoalesceExpr(_)) => "coalesce".into(),
        Some(NodeEnum::MinMaxExpr(m)) => {
            if m.op == pg_query::protobuf::MinMaxOp::IsGreatest as i32 { "greatest" } else { "least" }.into()
        }
        Some(NodeEnum::AExpr(e)) if e.kind == AExprKind::AexprNullif as i32 => "nullif".into(),
        Some(NodeEnum::SubLink(link)) => match pg_query::protobuf::SubLinkType::try_from(link.sub_link_type) {
            Ok(pg_query::protobuf::SubLinkType::ExistsSublink) => "exists".into(),
            Ok(pg_query::protobuf::SubLinkType::ArraySublink) => "array".into(),
            Ok(pg_query::protobuf::SubLinkType::ExprSublink) => {
                match link.subselect.as_deref().and_then(|n| n.node.as_ref()) {
                    Some(NodeEnum::SelectStmt(select)) => select
                        .target_list
                        .first()
                        .and_then(|t| match t.node.as_ref() {
                            Some(NodeEnum::ResTarget(t)) if !t.name.is_empty() => Some(t.name.clone()),
                            Some(NodeEnum::ResTarget(t)) => t.val.as_deref().map(figure_name),
                            _ => None,
                        })
                        .unwrap_or_else(|| "?column?".into()),
                    _ => "?column?".into(),
                }
            }
            _ => "?column?".into(),
        },
        Some(NodeEnum::SqlvalueFunction(f)) => match SqlValueFunctionOp::try_from(f.op) {
            Ok(SqlValueFunctionOp::SvfopCurrentUser) => "current_user",
            Ok(SqlValueFunctionOp::SvfopCurrentRole) => "current_role",
            Ok(SqlValueFunctionOp::SvfopUser) => "user",
            Ok(SqlValueFunctionOp::SvfopSessionUser) => "session_user",
            Ok(SqlValueFunctionOp::SvfopCurrentCatalog) => "current_catalog",
            Ok(SqlValueFunctionOp::SvfopCurrentSchema) => "current_schema",
            Ok(SqlValueFunctionOp::SvfopCurrentDate) => "current_date",
            Ok(SqlValueFunctionOp::SvfopCurrentTime | SqlValueFunctionOp::SvfopCurrentTimeN) => "current_time",
            Ok(SqlValueFunctionOp::SvfopCurrentTimestamp | SqlValueFunctionOp::SvfopCurrentTimestampN) => {
                "current_timestamp"
            }
            Ok(SqlValueFunctionOp::SvfopLocaltime | SqlValueFunctionOp::SvfopLocaltimeN) => "localtime",
            Ok(SqlValueFunctionOp::SvfopLocaltimestamp | SqlValueFunctionOp::SvfopLocaltimestampN) => "localtimestamp",
            _ => "?column?",
        }
        .into(),
        Some(NodeEnum::AArrayExpr(_)) => "array".into(),
        _ => "?column?".into(),
    }
}

/// as_i64 returns an integer value widened to 64 bits.
fn as_i64(value: &Value) -> Option<i64> {
    match value {
        Value::Int2(i) => Some(*i as i64),
        Value::Int4(i) => Some(*i as i64),
        Value::Int8(i) => Some(*i),
        _ => None,
    }
}

/// int_result narrows a 64-bit result to the integer type, failing as Postgres does when it is out of range.
fn int_result(value: Option<i64>, ty: ColumnType) -> Result<Value> {
    let out_of_range = |name: &str| PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, format!("{name} out of range"));
    match ty.oid {
        oid::INT2 => value.and_then(|v| i16::try_from(v).ok()).map(Value::Int2).ok_or_else(|| out_of_range("smallint")),
        oid::INT4 => value.and_then(|v| i32::try_from(v).ok()).map(Value::Int4).ok_or_else(|| out_of_range("integer")),
        _ => value.map(Value::Int8).ok_or_else(|| out_of_range("bigint")),
    }
}

/// compare_values orders two non-NULL values of the same domain.
pub fn compare_values(left: &Value, right: &Value) -> Ordering {
    match (left, right) {
        (Value::Bool(l), Value::Bool(r)) => l.cmp(r),
        (Value::Float4(l), Value::Float4(r)) => compare_floats(*l as f64, *r as f64),
        (Value::Float8(l), Value::Float8(r)) => compare_floats(*l, *r),
        (Value::Text(l), Value::Text(r)) => l.as_bytes().cmp(r.as_bytes()),
        (Value::Numeric(l), Value::Numeric(r)) => l.cmp_numeric(r),
        (l, r) => match (as_i64(l), as_i64(r)) {
            (Some(l), Some(r)) => l.cmp(&r),
            _ => Ordering::Equal,
        },
    }
}

/// compare_floats orders floats as Postgres does, with NaN above every other value.
fn compare_floats(left: f64, right: f64) -> Ordering {
    left.partial_cmp(&right).unwrap_or_else(|| left.is_nan().cmp(&right.is_nan()))
}

/// division_by_zero returns Postgres' division by zero error.
fn division_by_zero() -> PgError {
    PgError::new(code::DIVISION_BY_ZERO, "division by zero")
}

/// float_result checks a float result for overflow and underflow as Postgres does.
fn float_result(value: f64, inputs_finite: bool) -> Result<f64> {
    if value.is_infinite() && inputs_finite {
        return Err(PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, "value out of range: overflow"));
    }
    Ok(value)
}

impl Expr {
    /// eval evaluates the expression over a row.
    pub fn eval(&self, ctx: &mut Ctx<'_>, row: &[Value]) -> Result<Value> {
        Ok(match self {
            Expr::Const(value) => value.clone(),
            Expr::Column(i) => row[*i].clone(),
            Expr::Param(i) => ctx.params.get(*i).cloned().unwrap_or(Value::Null),
            Expr::Cast(expr, ty, explicit) => {
                let value = expr.eval(ctx, row)?;
                match value {
                    Value::Text(text) if !matches!(ty.oid, oid::TEXT | oid::VARCHAR | oid::BPCHAR | oid::NAME) => {
                        cast_value(crate::cast::input(&text, ty.oid)?, *ty, *explicit)?
                    }
                    value => cast_value(value, *ty, *explicit)?,
                }
            }
            Expr::Arith(op, left, right, ty) => {
                let (left, right) = (left.eval(ctx, row)?, right.eval(ctx, row)?);
                if left.is_null() || right.is_null() {
                    return Ok(Value::Null);
                }
                arith(*op, &left, &right, *ty)?
            }
            Expr::Neg(expr, ty) => match expr.eval(ctx, row)? {
                Value::Null => Value::Null,
                Value::Numeric(n) => Value::Numeric(n.negate()),
                Value::Float4(f) => Value::Float4(-f),
                Value::Float8(f) => Value::Float8(-f),
                value => int_result(as_i64(&value).and_then(i64::checked_neg), *ty)?,
            },
            Expr::Compare(op, left, right) => {
                let (left, right) = (left.eval(ctx, row)?, right.eval(ctx, row)?);
                if left.is_null() || right.is_null() {
                    return Ok(Value::Null);
                }
                let ordering = compare_values(&left, &right);
                Value::Bool(match op {
                    CmpOp::Eq => ordering == Ordering::Equal,
                    CmpOp::Ne => ordering != Ordering::Equal,
                    CmpOp::Lt => ordering == Ordering::Less,
                    CmpOp::Le => ordering != Ordering::Greater,
                    CmpOp::Gt => ordering == Ordering::Greater,
                    CmpOp::Ge => ordering != Ordering::Less,
                })
            }
            Expr::Concat(left, right) => match (left.eval(ctx, row)?, right.eval(ctx, row)?) {
                (Value::Null, _) | (_, Value::Null) => Value::Null,
                (l, r) => Value::Text(format!("{}{}", l.output().unwrap_or_default(), r.output().unwrap_or_default())),
            },
            Expr::And(left, right) => match left.eval(ctx, row)? {
                Value::Bool(false) => Value::Bool(false),
                l => match (l, right.eval(ctx, row)?) {
                    (_, Value::Bool(false)) => Value::Bool(false),
                    (Value::Bool(true), Value::Bool(true)) => Value::Bool(true),
                    _ => Value::Null,
                },
            },
            Expr::Or(left, right) => match left.eval(ctx, row)? {
                Value::Bool(true) => Value::Bool(true),
                l => match (l, right.eval(ctx, row)?) {
                    (_, Value::Bool(true)) => Value::Bool(true),
                    (Value::Bool(false), Value::Bool(false)) => Value::Bool(false),
                    _ => Value::Null,
                },
            },
            Expr::Not(expr) => match expr.eval(ctx, row)? {
                Value::Bool(b) => Value::Bool(!b),
                _ => Value::Null,
            },
            Expr::IsNull(expr, negated) => Value::Bool(expr.eval(ctx, row)?.is_null() != *negated),
            Expr::Func(index, args) => {
                let values = args.iter().map(|a| a.eval(ctx, row)).collect::<Result<Vec<_>>>()?;
                functions::call(ctx, *index, &values)?
            }
            Expr::Outer(depth, i) => {
                let level = ctx
                    .outer
                    .len()
                    .checked_sub(*depth)
                    .ok_or_else(|| PgError::internal("an outer column without its query"))?;
                ctx.outer[level].get(*i).cloned().unwrap_or(Value::Null)
            }
            Expr::InputColumn(_) | Expr::AggRef(_) => return Err(PgError::internal("an ungrouped expression")),
            Expr::SubqueryValue => ctx.subquery_value.clone(),
            Expr::Coalesce(args) => {
                for arg in args {
                    let value = arg.eval(ctx, row)?;
                    if !value.is_null() {
                        return Ok(value);
                    }
                }
                Value::Null
            }
            Expr::Case(whens, otherwise) => {
                for (condition, result) in whens {
                    if condition.is_true(ctx, row)? {
                        return result.eval(ctx, row);
                    }
                }
                otherwise.eval(ctx, row)?
            }
            Expr::NullIf(value, test) => {
                if test.is_true(ctx, row)? {
                    Value::Null
                } else {
                    value.eval(ctx, row)?
                }
            }
            Expr::MinMax(greatest, args) => {
                let mut best: Option<Value> = None;
                for arg in args {
                    let value = arg.eval(ctx, row)?;
                    if value.is_null() {
                        continue;
                    }
                    let better = best.as_ref().is_none_or(|b| {
                        let ordering = compare_values(&value, b);
                        if *greatest { ordering == Ordering::Greater } else { ordering == Ordering::Less }
                    });
                    if better {
                        best = Some(value);
                    }
                }
                best.unwrap_or(Value::Null)
            }
            Expr::DistinctFrom(left, right, negated) => {
                let (l, r) = (left.eval(ctx, row)?, right.eval(ctx, row)?);
                let distinct = match (&l, &r) {
                    (Value::Null, Value::Null) => false,
                    (Value::Null, _) | (_, Value::Null) => true,
                    (l, r) => compare_values(l, r) != Ordering::Equal,
                };
                Value::Bool(distinct != *negated)
            }
            Expr::BoolTest(expr, value, negated) => {
                let v = expr.eval(ctx, row)?;
                let holds = match (value, &v) {
                    (None, Value::Null) => true,
                    (Some(b), Value::Bool(x)) => b == x,
                    _ => false,
                };
                Value::Bool(holds != *negated)
            }
            Expr::Exists(plan) => {
                ctx.outer.push(row.to_vec());
                let rows = plan.run(ctx);
                ctx.outer.pop();
                Value::Bool(!rows?.is_empty())
            }
            Expr::Scalar(plan) => {
                ctx.outer.push(row.to_vec());
                let rows = plan.run(ctx);
                ctx.outer.pop();
                let rows = rows?;
                if rows.len() > 1 {
                    return Err(PgError::new(
                        code::CARDINALITY_VIOLATION,
                        "more than one row returned by a subquery used as an expression",
                    ));
                }
                rows.into_iter().next().and_then(|r| r.into_iter().next()).unwrap_or(Value::Null)
            }
            Expr::AnySubquery(comparison, plan, all) => {
                ctx.outer.push(row.to_vec());
                let rows = plan.run(ctx);
                ctx.outer.pop();
                let mut saw_null = false;
                let previous = std::mem::replace(&mut ctx.subquery_value, Value::Null);
                let mut result = None;
                for r in rows? {
                    ctx.subquery_value = r.into_iter().next().unwrap_or(Value::Null);
                    match comparison.eval(ctx, row)? {
                        Value::Bool(b) if b != *all => {
                            result = Some(b);
                            break;
                        }
                        Value::Null => saw_null = true,
                        _ => {}
                    }
                }
                ctx.subquery_value = previous;
                match result {
                    Some(b) => Value::Bool(b),
                    None if saw_null => Value::Null,
                    None => Value::Bool(*all),
                }
            }
        })
    }

    /// is_true evaluates a condition, treating NULL as false.
    pub fn is_true(&self, ctx: &mut Ctx<'_>, row: &[Value]) -> Result<bool> {
        Ok(matches!(self.eval(ctx, row)?, Value::Bool(true)))
    }
}

impl Expr {
    /// map_children rebuilds the expression with each child replaced, leaving subquery plans alone.
    pub fn map_children(self, f: &mut dyn FnMut(Expr) -> Expr) -> Expr {
        let mut b = |e: Box<Expr>| Box::new(f(*e));
        match self {
            Expr::Cast(e, t, x) => Expr::Cast(b(e), t, x),
            Expr::Arith(op, l, r, t) => {
                let l = b(l);
                Expr::Arith(op, l, b(r), t)
            }
            Expr::Neg(e, t) => Expr::Neg(b(e), t),
            Expr::Compare(op, l, r) => {
                let l = b(l);
                Expr::Compare(op, l, b(r))
            }
            Expr::Concat(l, r) => {
                let l = b(l);
                Expr::Concat(l, b(r))
            }
            Expr::And(l, r) => {
                let l = b(l);
                Expr::And(l, b(r))
            }
            Expr::Or(l, r) => {
                let l = b(l);
                Expr::Or(l, b(r))
            }
            Expr::Not(e) => Expr::Not(b(e)),
            Expr::IsNull(e, n) => Expr::IsNull(b(e), n),
            Expr::Func(i, args) => Expr::Func(i, args.into_iter().map(&mut *f).collect()),
            Expr::Coalesce(args) => Expr::Coalesce(args.into_iter().map(&mut *f).collect()),
            Expr::MinMax(g, args) => Expr::MinMax(g, args.into_iter().map(&mut *f).collect()),
            Expr::Case(whens, otherwise) => {
                let whens = whens.into_iter().map(|(c, r)| (f(c), f(r))).collect();
                Expr::Case(whens, Box::new(f(*otherwise)))
            }
            Expr::NullIf(v, t) => {
                let v = b(v);
                Expr::NullIf(v, b(t))
            }
            Expr::DistinctFrom(l, r, n) => {
                let l = b(l);
                Expr::DistinctFrom(l, b(r), n)
            }
            Expr::BoolTest(e, v, n) => Expr::BoolTest(b(e), v, n),
            Expr::AnySubquery(c, p, all) => Expr::AnySubquery(b(c), p, all),
            other => other,
        }
    }

    /// visit calls the function on the expression and each of its descendants, outside subquery plans.
    pub fn visit(&self, f: &mut dyn FnMut(&Expr)) {
        f(self);
        match self {
            Expr::Cast(e, ..) | Expr::Neg(e, _) | Expr::Not(e) | Expr::IsNull(e, _) | Expr::BoolTest(e, ..) => {
                e.visit(f)
            }
            Expr::Arith(_, l, r, _)
            | Expr::Compare(_, l, r)
            | Expr::Concat(l, r)
            | Expr::And(l, r)
            | Expr::Or(l, r)
            | Expr::NullIf(l, r)
            | Expr::DistinctFrom(l, r, _) => {
                l.visit(f);
                r.visit(f);
            }
            Expr::Func(_, args) | Expr::Coalesce(args) | Expr::MinMax(_, args) => args.iter().for_each(|a| a.visit(f)),
            Expr::Case(whens, otherwise) => {
                for (c, r) in whens {
                    c.visit(f);
                    r.visit(f);
                }
                otherwise.visit(f);
            }
            Expr::AnySubquery(c, ..) => c.visit(f),
            _ => {}
        }
    }
}

/// arith applies an arithmetic operator to two non-NULL values of the result type's domain.
fn arith(op: ArithOp, left: &Value, right: &Value, ty: ColumnType) -> Result<Value> {
    if let (Some(l), Some(r)) = (as_i64(left), as_i64(right)) {
        let result = match op {
            ArithOp::Add => l.checked_add(r),
            ArithOp::Sub => l.checked_sub(r),
            ArithOp::Mul => l.checked_mul(r),
            ArithOp::Div if r == 0 => return Err(division_by_zero()),
            ArithOp::Div => l.checked_div(r),
            ArithOp::Mod if r == 0 => return Err(division_by_zero()),
            ArithOp::Mod => Some(l.checked_rem(r).unwrap_or(0)),
        };
        return int_result(result, ty);
    }
    if let (Value::Numeric(l), Value::Numeric(r)) = (left, right) {
        return Ok(Value::Numeric(match op {
            ArithOp::Add => l.add(r),
            ArithOp::Sub => l.sub(r),
            ArithOp::Mul => l.mul(r),
            ArithOp::Div => l.div(r)?,
            ArithOp::Mod => l.rem(r)?,
        }));
    }
    let float = |value: &Value| match value {
        Value::Float4(f) => *f as f64,
        Value::Float8(f) => *f,
        other => as_i64(other).unwrap_or_default() as f64,
    };
    let (l, r) = (float(left), float(right));
    let result = match op {
        ArithOp::Add => l + r,
        ArithOp::Sub => l - r,
        ArithOp::Mul => l * r,
        ArithOp::Div if r == 0.0 => return Err(division_by_zero()),
        ArithOp::Div => l / r,
        ArithOp::Mod => return Err(PgError::internal("modulo of floats")),
    };
    let finite = l.is_finite() && r.is_finite();
    Ok(if ty.oid == oid::FLOAT4 {
        Value::Float4(float_result(result as f32 as f64, finite)? as f32)
    } else {
        Value::Float8(float_result(result, finite)?)
    })
}
