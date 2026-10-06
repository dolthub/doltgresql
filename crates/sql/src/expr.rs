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

use crate::array::{Array, is_array_type};
use crate::cast::{cast_value, type_display};
use crate::catalog::{ColumnType, builtin_type, resolve_type};
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

/// DateOp is an operator on dates, times, timestamps, and intervals.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum DateOp {
    DatePlusDays,
    DateMinusDays,
    DateMinusDate,
    /// A timestamp, or timestamptz when true, plus or minus an interval.
    TimestampPlusInterval(bool),
    TimestampMinusInterval(bool),
    TimestampMinusTimestamp,
    TimePlusInterval,
    TimeMinusInterval,
    TimeMinusTime,
    DatePlusTime,
    IntervalPlusInterval,
    IntervalMinusInterval,
    IntervalTimesFloat,
    IntervalDivFloat,
}

/// ArrayOp is an operator over arrays.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum ArrayOp {
    Concat,
    Append,
    Prepend,
    Contains,
    ContainedBy,
    Overlaps,
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
    /// The default of a column of the table being written, by position.
    Default(usize),
    /// A date and time operator.
    DateTime(DateOp, Box<Expr>, Box<Expr>),
    /// An ARRAY constructor of the element type, whose items are themselves arrays when it is multidimensional.
    Array(u32, Vec<Expr>, bool),
    /// Subscripts of an array, as lower and upper bounds, which select a slice when the flag is set.
    Subscript(Box<Expr>, Vec<(Option<Expr>, Option<Expr>)>, bool),
    /// A comparison against each element of an array, which holds for every element when the flag is set.
    AnyArray(Box<Expr>, Box<Expr>, bool),
    ArrayOp(ArrayOp, Box<Expr>, Box<Expr>),
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
            _ if implicitly_converts(result.oid, ty.oid) => typ(ty.oid),
            _ if implicitly_converts(ty.oid, result.oid) => typ(result.oid),
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
                let bound = match arg.node.as_ref() {
                    Some(NodeEnum::AArrayExpr(array)) if is_array_type(target.oid) => {
                        self.array_expr(array, Some(target.oid))?
                    }
                    _ => self.bind(arg)?,
                };
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
                    Ok(SqlValueFunctionOp::SvfopCurrentDate) => "current_date",
                    Ok(SqlValueFunctionOp::SvfopCurrentTime | SqlValueFunctionOp::SvfopCurrentTimeN) => "current_time",
                    Ok(SqlValueFunctionOp::SvfopCurrentTimestamp | SqlValueFunctionOp::SvfopCurrentTimestampN) => "now",
                    Ok(SqlValueFunctionOp::SvfopLocaltime | SqlValueFunctionOp::SvfopLocaltimeN) => "localtime",
                    Ok(SqlValueFunctionOp::SvfopLocaltimestamp | SqlValueFunctionOp::SvfopLocaltimestampN) => {
                        "localtimestamp"
                    }
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
            NodeEnum::AArrayExpr(array) => self.array_expr(array, None),
            NodeEnum::AIndirection(indirection) => self.indirection(indirection),
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
            AExprKind::AexprOpAny | AExprKind::AexprOpAll => {
                let left = self.bind(operand(&e.lexpr)?)?;
                let mut right = self.bind(operand(&e.rexpr)?)?;
                if right.1.oid == oid::UNKNOWN {
                    let array_type = if left.1.oid == oid::UNKNOWN { oid::TEXT_ARRAY } else { array_of(left.1.oid) };
                    right = coerce(right, typ(array_type), false, e.location)?;
                }
                if !is_array_type(right.1.oid) {
                    return Err(PgError {
                        position: position(e.location),
                        ..PgError::new(code::WRONG_OBJECT_TYPE, "op ANY/ALL (array) requires array on right side")
                    });
                }
                let element = ColumnType { oid: element_type(right.1.oid), ..right.1 };
                let (comparison, _) = self.binary(&op, left, (Expr::SubqueryValue, element), e.location)?;
                let all = kind == AExprKind::AexprOpAll;
                Ok((Expr::AnyArray(Box::new(comparison), Box::new(right.0), all), typ(oid::BOOL)))
            }
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
        if is_array_type(lt) || is_array_type(rt) {
            return self.array_binary(op, left, right, location).and_then(|b| b.ok_or_else(missing));
        }
        if is_datetime(lt) || is_datetime(rt) {
            if let Some(bound) = self.datetime_binary(op, &left, &right, location)? {
                return Ok(bound);
            }
            if !(is_datetime(lt) && (rt == oid::UNKNOWN || rt == lt) || is_datetime(rt) && lt == oid::UNKNOWN) {
                return Err(missing());
            }
        }
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

impl<'b, 'a> Binder<'b, 'a> {
    /// array_expr binds an ARRAY constructor, whose type comes from a cast around it when it is empty.
    fn array_expr(&mut self, array: &pg_query::protobuf::AArrayExpr, target: Option<u32>) -> Result<Bound> {
        let nested = array.elements.iter().any(|e| matches!(e.node.as_ref(), Some(NodeEnum::AArrayExpr(_))));
        let mut bound = Vec::with_capacity(array.elements.len());
        for element in &array.elements {
            let item = match element.node.as_ref() {
                Some(NodeEnum::AArrayExpr(inner)) => self.array_expr(inner, target)?,
                _ => self.bind(element)?,
            };
            bound.push((item, arg_location(element)));
        }
        if bound.is_empty() {
            let Some(target) = target else {
                return Err(PgError {
                    position: position(array.location),
                    hint: Some("Explicitly cast to the desired type, for example ARRAY[]::integer[].".into()),
                    ..PgError::new(code::INDETERMINATE_DATATYPE, "cannot determine type of empty array")
                });
            };
            return Ok((Expr::Array(element_type(target), Vec::new(), false), typ(target)));
        }
        let types: Vec<(ColumnType, i32)> = bound.iter().map(|((_, t), l)| (*t, *l)).collect();
        let mut ty = match target {
            Some(target) if types.iter().all(|(t, _)| t.oid == oid::UNKNOWN) => {
                typ(if nested { target } else { element_type(target) })
            }
            _ => common_type(&types, "ARRAY")?,
        };
        if nested && !is_array_type(ty.oid) {
            ty = typ(array_of(ty.oid));
        }
        let mut items = Vec::with_capacity(bound.len());
        for (item, location) in bound {
            items.push(coerce(item, ty, false, location)?.0);
        }
        if nested {
            return Ok((Expr::Array(element_type(ty.oid), items, true), typ(ty.oid)));
        }
        Ok((Expr::Array(ty.oid, items, false), typ(array_of(ty.oid))))
    }

    /// indirection binds subscripts of an array.
    fn indirection(&mut self, indirection: &pg_query::protobuf::AIndirection) -> Result<Bound> {
        let arg = indirection.arg.as_deref().ok_or_else(|| PgError::internal("no subscripted value"))?;
        let (base, ty) = self.bind(arg)?;
        let mut subscripts = Vec::new();
        let mut slice = false;
        for item in &indirection.indirection {
            let Some(NodeEnum::AIndices(indices)) = item.node.as_ref() else {
                return Err(PgError::unsupported("field selection"));
            };
            slice |= indices.is_slice;
            let mut bound_index = |node: &Option<Box<Node>>| -> Result<Option<Expr>> {
                match node.as_deref() {
                    Some(node) => {
                        let bound = self.bind(node)?;
                        let location = arg_location(node);
                        Ok(Some(subscript_int(bound, location)?))
                    }
                    None => Ok(None),
                }
            };
            let lower = bound_index(&indices.lidx)?;
            let upper = bound_index(&indices.uidx)?;
            subscripts.push(if indices.is_slice { (lower, upper) } else { (Some(Expr::Const(Value::Int4(1))), upper) });
        }
        if !is_array_type(ty.oid) {
            return Err(PgError {
                position: position(arg_location(arg)),
                ..PgError::new(
                    code::DATATYPE_MISMATCH,
                    format!("cannot subscript type {} because it does not support subscripting", type_display(ty.oid)),
                )
            });
        }
        let result = if slice { ty } else { ColumnType { oid: element_type(ty.oid), ..ty } };
        Ok((Expr::Subscript(Box::new(base), subscripts, slice), result))
    }

    /// array_binary binds an operator with an array operand, or returns None when no array operator matches.
    fn array_binary(&mut self, op: &str, left: Bound, right: Bound, location: i32) -> Result<Option<Bound>> {
        let (lt, rt) = (left.1.oid, right.1.oid);
        let (l_array, r_array) = (is_array_type(lt), is_array_type(rt));
        let array_op = match op {
            "@>" => Some(ArrayOp::Contains),
            "<@" => Some(ArrayOp::ContainedBy),
            "&&" => Some(ArrayOp::Overlaps),
            "||" => Some(ArrayOp::Concat),
            _ => None,
        };
        if op == "||" && (!l_array || !r_array) && lt != oid::UNKNOWN && rt != oid::UNKNOWN {
            let (array, element, kind) =
                if l_array { (&left.1, rt, ArrayOp::Append) } else { (&right.1, lt, ArrayOp::Prepend) };
            let common = common_type(&[(typ(element_type(array.oid)), location), (typ(element), location)], "ARRAY")
                .map_err(|_| ())
                .ok();
            let Some(common) = common else { return Ok(None) };
            let array_type = typ(array_of(common.oid));
            let (l, r) = if l_array {
                (coerce(left, array_type, false, location)?.0, coerce(right, common, false, location)?.0)
            } else {
                (coerce(left, common, false, location)?.0, coerce(right, array_type, false, location)?.0)
            };
            return Ok(Some((Expr::ArrayOp(kind, Box::new(l), Box::new(r)), array_type)));
        }
        let domain = match (l_array, r_array) {
            (true, true) => match common_type(&[(typ(lt), location), (typ(rt), location)], "ARRAY") {
                Ok(t) => t,
                Err(_) if op == "||" => {
                    let (le, re) = (element_type(lt), element_type(rt));
                    if le != re {
                        return Ok(None);
                    }
                    left.1
                }
                Err(_) => return Ok(None),
            },
            (true, false) if rt == oid::UNKNOWN => left.1,
            (false, true) if lt == oid::UNKNOWN => right.1,
            _ => return Ok(None),
        };
        let domain = ColumnType { modifier: -1, ..domain };
        let l = coerce(left, domain, false, location)?.0;
        let r = coerce(right, domain, false, location)?.0;
        if let Some(kind) = array_op {
            let ret = if kind == ArrayOp::Concat { domain } else { typ(oid::BOOL) };
            return Ok(Some((Expr::ArrayOp(kind, Box::new(l), Box::new(r)), ret)));
        }
        let cmp = match op {
            "=" => CmpOp::Eq,
            "<>" | "!=" => CmpOp::Ne,
            "<" => CmpOp::Lt,
            "<=" => CmpOp::Le,
            ">" => CmpOp::Gt,
            ">=" => CmpOp::Ge,
            _ => return Ok(None),
        };
        Ok(Some((Expr::Compare(cmp, Box::new(l), Box::new(r)), typ(oid::BOOL))))
    }
}

/// subscript_int converts an array subscript to an integer.
fn subscript_int(bound: Bound, location: i32) -> Result<Expr> {
    if !matches!(
        bound.1.oid,
        oid::INT2 | oid::INT4 | oid::INT8 | oid::UNKNOWN | oid::NUMERIC | oid::FLOAT4 | oid::FLOAT8
    ) {
        return Err(PgError {
            position: position(location),
            ..PgError::new(code::DATATYPE_MISMATCH, "array subscript must have type integer")
        });
    }
    Ok(coerce(bound, typ(oid::INT4), true, location)?.0)
}

/// implicit_datetime reports whether Postgres converts one date or time type to another without being asked.
fn implicit_datetime(from: u32, to: u32) -> bool {
    matches!(
        (from, to),
        (oid::DATE, oid::TIMESTAMP | oid::TIMESTAMPTZ)
            | (oid::TIMESTAMP, oid::TIMESTAMPTZ)
            | (oid::TIME, oid::TIMETZ | oid::INTERVAL)
    )
}

/// assignable_datetime reports whether a date or time type converts to another on assignment.
fn assignable_datetime(from: u32, to: u32) -> bool {
    implicit_datetime(from, to)
        || matches!(
            (from, to),
            (oid::TIMESTAMP, oid::DATE | oid::TIME)
                | (oid::TIMESTAMPTZ, oid::DATE | oid::TIME | oid::TIMETZ | oid::TIMESTAMP)
                | (oid::INTERVAL, oid::TIME)
                | (oid::TIMETZ, oid::TIME)
        )
}

/// is_datetime reports whether a type is a date, time, timestamp, or interval type.
fn is_datetime(type_oid: u32) -> bool {
    matches!(type_oid, oid::DATE | oid::TIME | oid::TIMETZ | oid::TIMESTAMP | oid::TIMESTAMPTZ | oid::INTERVAL)
}

impl<'b, 'a> Binder<'b, 'a> {
    /// datetime_binary binds an arithmetic or comparison operator with a date or time operand, returning None when
    /// the operands compare as one type and need the general comparison.
    fn datetime_binary(&mut self, op: &str, left: &Bound, right: &Bound, location: i32) -> Result<Option<Bound>> {
        let (mut lt, mut rt) = (left.1.oid, right.1.oid);
        // An untyped operand of + or - with a date or time is an interval, except that minus takes the other type
        // first, as Postgres' operator resolution does.
        if rt == oid::UNKNOWN && lt != oid::UNKNOWN {
            rt = match (op, lt) {
                ("-", _) => lt,
                ("+", oid::DATE) => {
                    return Err(PgError {
                        position: position(location),
                        hint: Some(
                            "Could not choose a best candidate operator. You might need to add explicit type casts."
                                .into(),
                        ),
                        ..PgError::new(code::AMBIGUOUS_FUNCTION, "operator is not unique: date + unknown")
                    });
                }
                ("+", oid::TIMESTAMP | oid::TIMESTAMPTZ | oid::TIME) => oid::INTERVAL,
                _ => lt,
            };
        } else if lt == oid::UNKNOWN && rt != oid::UNKNOWN {
            lt = if op == "+" && matches!(rt, oid::TIMESTAMP | oid::TIMESTAMPTZ | oid::TIME | oid::DATE) {
                oid::INTERVAL
            } else {
                rt
            };
        }
        let int = |t: u32| matches!(t, oid::INT2 | oid::INT4 | oid::INT8);
        let num = |t: u32| numeric_rank(t).is_some();
        use DateOp as D;
        let (op_kind, lcast, rcast, ret) = match (op, lt, rt) {
            ("+", oid::DATE, t) if int(t) => (D::DatePlusDays, oid::DATE, oid::INT4, oid::DATE),
            ("+", t, oid::DATE) if int(t) => (D::DatePlusDays, oid::INT4, oid::DATE, oid::DATE),
            ("-", oid::DATE, t) if int(t) => (D::DateMinusDays, oid::DATE, oid::INT4, oid::DATE),
            ("-", oid::DATE, oid::DATE) => (D::DateMinusDate, oid::DATE, oid::DATE, oid::INT4),
            ("+", oid::DATE, oid::TIME) | ("+", oid::TIME, oid::DATE) => (D::DatePlusTime, lt, rt, oid::TIMESTAMP),
            ("+", oid::DATE | oid::TIMESTAMP, oid::INTERVAL) => {
                (D::TimestampPlusInterval(false), oid::TIMESTAMP, oid::INTERVAL, oid::TIMESTAMP)
            }
            ("+", oid::INTERVAL, oid::DATE | oid::TIMESTAMP) => {
                (D::TimestampPlusInterval(false), oid::INTERVAL, oid::TIMESTAMP, oid::TIMESTAMP)
            }
            ("-", oid::DATE | oid::TIMESTAMP, oid::INTERVAL) => {
                (D::TimestampMinusInterval(false), oid::TIMESTAMP, oid::INTERVAL, oid::TIMESTAMP)
            }
            ("+", oid::TIMESTAMPTZ, oid::INTERVAL) => {
                (D::TimestampPlusInterval(true), oid::TIMESTAMPTZ, oid::INTERVAL, oid::TIMESTAMPTZ)
            }
            ("+", oid::INTERVAL, oid::TIMESTAMPTZ) => {
                (D::TimestampPlusInterval(true), oid::INTERVAL, oid::TIMESTAMPTZ, oid::TIMESTAMPTZ)
            }
            ("-", oid::TIMESTAMPTZ, oid::INTERVAL) => {
                (D::TimestampMinusInterval(true), oid::TIMESTAMPTZ, oid::INTERVAL, oid::TIMESTAMPTZ)
            }
            ("-", oid::TIMESTAMP, oid::TIMESTAMP) => {
                (D::TimestampMinusTimestamp, oid::TIMESTAMP, oid::TIMESTAMP, oid::INTERVAL)
            }
            ("-", oid::TIMESTAMPTZ | oid::TIMESTAMP | oid::DATE, oid::TIMESTAMPTZ)
            | ("-", oid::TIMESTAMPTZ, oid::TIMESTAMP | oid::DATE) => {
                (D::TimestampMinusTimestamp, oid::TIMESTAMPTZ, oid::TIMESTAMPTZ, oid::INTERVAL)
            }
            ("+", oid::TIME, oid::INTERVAL) => (D::TimePlusInterval, oid::TIME, oid::INTERVAL, oid::TIME),
            ("+", oid::INTERVAL, oid::TIME) => (D::TimePlusInterval, oid::INTERVAL, oid::TIME, oid::TIME),
            ("-", oid::TIME, oid::INTERVAL) => (D::TimeMinusInterval, oid::TIME, oid::INTERVAL, oid::TIME),
            ("-", oid::TIME, oid::TIME) => (D::TimeMinusTime, oid::TIME, oid::TIME, oid::INTERVAL),
            ("+", oid::INTERVAL, oid::INTERVAL) => {
                (D::IntervalPlusInterval, oid::INTERVAL, oid::INTERVAL, oid::INTERVAL)
            }
            ("-", oid::INTERVAL, oid::INTERVAL) => {
                (D::IntervalMinusInterval, oid::INTERVAL, oid::INTERVAL, oid::INTERVAL)
            }
            ("*", oid::INTERVAL, t) if num(t) || t == oid::UNKNOWN => {
                (D::IntervalTimesFloat, oid::INTERVAL, oid::FLOAT8, oid::INTERVAL)
            }
            ("*", t, oid::INTERVAL) if num(t) || t == oid::UNKNOWN => {
                (D::IntervalTimesFloat, oid::FLOAT8, oid::INTERVAL, oid::INTERVAL)
            }
            ("/", oid::INTERVAL, t) if num(t) || t == oid::UNKNOWN => {
                (D::IntervalDivFloat, oid::INTERVAL, oid::FLOAT8, oid::INTERVAL)
            }
            ("=" | "<>" | "!=" | "<" | "<=" | ">" | ">=", l, r) if l != r && is_datetime(l) && is_datetime(r) => {
                // Dates promote to timestamps, and timestamps to timestamptz, so that mixed comparisons agree.
                let common = if l == oid::TIMESTAMPTZ || r == oid::TIMESTAMPTZ {
                    oid::TIMESTAMPTZ
                } else if matches!((l, r), (oid::DATE, oid::TIMESTAMP) | (oid::TIMESTAMP, oid::DATE)) {
                    oid::TIMESTAMP
                } else if matches!((l, r), (oid::TIME, oid::TIMETZ) | (oid::TIMETZ, oid::TIME)) {
                    oid::TIMETZ
                } else {
                    return Ok(None);
                };
                let l = coerce(left.clone(), typ(common), false, location)?.0;
                let r = coerce(right.clone(), typ(common), false, location)?.0;
                return Ok(Some(self.binary(op, (l, typ(common)), (r, typ(common)), location)?));
            }
            _ => return Ok(None),
        };
        for (bound, ty) in [(left, lcast), (right, rcast)] {
            if let Expr::Param(i) = bound.0
                && self.ctx.parameters[i] == 0
            {
                self.ctx.parameters[i] = ty;
            }
        }
        let l = coerce(left.clone(), typ(lcast), false, location)?.0;
        let r = coerce(right.clone(), typ(rcast), false, location)?.0;
        Ok(Some((Expr::DateTime(op_kind, Box::new(l), Box::new(r)), typ(ret))))
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
        "-" if ty.oid == oid::INTERVAL => Ok((Expr::Neg(Box::new(expr), ty), ty)),
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
    let allowed = explicit && (is_array_type(from.oid) == is_array_type(to.oid) || is_string(from.oid))
        || implicitly_converts(from.oid, to.oid);
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

/// implicitly_converts reports whether Postgres converts a value of one type to the other without being asked.
fn implicitly_converts(from: u32, to: u32) -> bool {
    if is_array_type(from) && is_array_type(to) {
        return implicitly_converts(element_type(from), element_type(to));
    }
    let numeric = numeric_rank(from).zip(numeric_rank(to));
    from == to
        || numeric.is_some_and(|(f, t)| f <= t)
        || (is_string(from) && is_string(to))
        || implicit_datetime(from, to)
}

/// assignable reports whether a value of one type converts to the other on assignment.
fn assignable(from: u32, to: u32) -> bool {
    if is_array_type(from) && is_array_type(to) {
        return assignable(element_type(from), element_type(to));
    }
    from == oid::UNKNOWN
        || from == to
        || (numeric_rank(from).is_some() && numeric_rank(to).is_some())
        || assignable_datetime(from, to)
        || (is_string(to) && !is_array_type(from))
}

/// element_type returns the element type of an array type.
pub fn element_type(array_type: u32) -> u32 {
    builtin_type(array_type).map_or(oid::TEXT, |t| t.elem)
}

/// array_of returns the array type of an element type.
pub fn array_of(element: u32) -> u32 {
    builtin_type(element).map_or(oid::TEXT_ARRAY, |t| t.array)
}

/// assign converts a bound expression to a column's type as an assignment does, which also allows numeric narrowing
/// and conversions to text, and names the column in its error.
pub fn assign(bound: Bound, to: ColumnType, column: &str, location: i32) -> Result<Bound> {
    let from = bound.1.oid;
    if !assignable(from, to.oid) {
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
    figure_name_strength(node).0
}

/// figure_name_strength returns the name Postgres gives an expression's column, with how strongly the expression
/// names it: 2 for columns and functions, 1 for casts and constructs, and 0 for none, as FigureColnameInternal does.
fn figure_name_strength(node: &Node) -> (String, u8) {
    let strong = |name: &str| (name.to_string(), 2);
    match node.node.as_ref() {
        Some(NodeEnum::ColumnRef(c)) => match c.fields.iter().filter_map(node_name).next_back() {
            Some(name) => strong(name),
            None => ("?column?".into(), 0),
        },
        Some(NodeEnum::FuncCall(f)) => {
            strong(f.funcname.iter().filter_map(node_name).next_back().unwrap_or("?column?"))
        }
        Some(NodeEnum::TypeCast(cast)) => match cast.arg.as_deref().map(figure_name_strength) {
            Some((name, strength)) if strength > 1 => (name, strength),
            _ => match cast.type_name.as_ref().and_then(|t| t.names.iter().filter_map(node_name).next_back()) {
                Some(name) => (name.to_string(), 1),
                None => ("?column?".into(), 0),
            },
        },
        Some(NodeEnum::CaseExpr(c)) => match c.defresult.as_deref().map(figure_name_strength) {
            Some((name, strength)) if strength > 1 => (name, strength),
            _ => ("case".into(), 1),
        },
        Some(NodeEnum::AArrayExpr(_)) => strong("array"),
        Some(NodeEnum::RowExpr(_)) => strong("row"),
        Some(NodeEnum::AIndirection(i)) => match i.indirection.iter().filter_map(node_name).next_back() {
            Some(name) => strong(name),
            None => i.arg.as_deref().map_or_else(|| ("?column?".into(), 0), figure_name_strength),
        },
        Some(NodeEnum::CoalesceExpr(_)) => strong("coalesce"),
        Some(NodeEnum::MinMaxExpr(m)) => {
            strong(if m.op == pg_query::protobuf::MinMaxOp::IsGreatest as i32 { "greatest" } else { "least" })
        }
        Some(NodeEnum::AExpr(e)) if e.kind == AExprKind::AexprNullif as i32 => strong("nullif"),
        Some(NodeEnum::SubLink(link)) => match pg_query::protobuf::SubLinkType::try_from(link.sub_link_type) {
            Ok(pg_query::protobuf::SubLinkType::ExistsSublink) => strong("exists"),
            Ok(pg_query::protobuf::SubLinkType::ArraySublink) => strong("array"),
            Ok(pg_query::protobuf::SubLinkType::ExprSublink) => {
                match link.subselect.as_deref().and_then(|n| n.node.as_ref()) {
                    Some(NodeEnum::SelectStmt(select)) => select
                        .target_list
                        .first()
                        .and_then(|t| match t.node.as_ref() {
                            Some(NodeEnum::ResTarget(t)) if !t.name.is_empty() => Some(strong(&t.name)),
                            Some(NodeEnum::ResTarget(t)) => t.val.as_deref().map(figure_name_strength),
                            _ => None,
                        })
                        .unwrap_or_else(|| ("?column?".into(), 0)),
                    _ => ("?column?".into(), 0),
                }
            }
            _ => ("?column?".into(), 0),
        },
        Some(NodeEnum::SqlvalueFunction(f)) => strong(match SqlValueFunctionOp::try_from(f.op) {
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
        }),
        _ => ("?column?".into(), 0),
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
        (Value::Date(l), Value::Date(r)) => l.cmp(r),
        (Value::Time(l), Value::Time(r)) => l.cmp(r),
        (Value::Timestamp(l), Value::Timestamp(r)) | (Value::TimestampTz(l), Value::TimestampTz(r)) => l.cmp(r),
        (Value::TimeTz(lt, lz), Value::TimeTz(rt, rz)) => ((*lt as i128) + (*lz as i128) * 1_000_000)
            .cmp(&((*rt as i128) + (*rz as i128) * 1_000_000))
            .then(lz.cmp(rz)),
        (Value::Interval(l), Value::Interval(r)) => l.cmp_key().cmp(&r.cmp_key()),
        (Value::Array(l), Value::Array(r)) => crate::array::compare(l, r),
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
                Value::Interval(iv) => Value::Interval(crate::functions::datetime::negate_interval(iv)?),
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
            Expr::Default(_) => return Err(PgError::internal("a default outside a written row")),
            Expr::DateTime(op, left, right) => {
                let (l, r) = (left.eval(ctx, row)?, right.eval(ctx, row)?);
                if l.is_null() || r.is_null() {
                    return Ok(Value::Null);
                }
                date_op(*op, l, r)?
            }
            Expr::SubqueryValue => ctx.subquery_value.clone(),
            Expr::Array(element, items, nested) => {
                let values = items.iter().map(|i| i.eval(ctx, row)).collect::<Result<Vec<_>>>()?;
                if *nested {
                    Value::Array(Box::new(crate::array::nest(*element, values)?))
                } else {
                    Value::Array(Box::new(Array::one_dimensional(*element, values)))
                }
            }
            Expr::Subscript(base, subscripts, slice) => {
                let Value::Array(array) = base.eval(ctx, row)? else { return Ok(Value::Null) };
                let mut bounds = Vec::with_capacity(subscripts.len());
                for (lower, upper) in subscripts {
                    let mut bound = |e: &Option<Expr>| -> Result<Option<Option<i32>>> {
                        match e {
                            Some(e) => match e.eval(ctx, row)? {
                                Value::Int4(i) => Ok(Some(Some(i))),
                                _ => Ok(None),
                            },
                            None => Ok(Some(None)),
                        }
                    };
                    let (Some(lower), Some(upper)) = (bound(lower)?, bound(upper)?) else { return Ok(Value::Null) };
                    bounds.push((lower, upper));
                }
                if *slice {
                    Value::Array(Box::new(crate::array::slice(&array, &bounds)))
                } else {
                    let indexes: Vec<i32> = bounds.iter().map(|(_, u)| u.unwrap_or(0)).collect();
                    crate::array::element(&array, &indexes).cloned().unwrap_or(Value::Null)
                }
            }
            Expr::AnyArray(comparison, array, all) => {
                let Value::Array(array) = array.eval(ctx, row)? else { return Ok(Value::Null) };
                let mut saw_null = false;
                let previous = std::mem::replace(&mut ctx.subquery_value, Value::Null);
                let mut result = None;
                for value in array.values {
                    ctx.subquery_value = value;
                    match comparison.eval(ctx, row) {
                        Ok(Value::Bool(b)) if b != *all => {
                            result = Some(b);
                            break;
                        }
                        Ok(Value::Null) => saw_null = true,
                        Ok(_) => {}
                        Err(err) => {
                            ctx.subquery_value = previous;
                            return Err(err);
                        }
                    }
                }
                ctx.subquery_value = previous;
                match result {
                    Some(b) => Value::Bool(b),
                    None if saw_null => Value::Null,
                    None => Value::Bool(*all),
                }
            }
            Expr::ArrayOp(op, left, right) => {
                let (l, r) = (left.eval(ctx, row)?, right.eval(ctx, row)?);
                crate::array::operate(*op, l, r)?
            }
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
            Expr::DateTime(op, l, r) => {
                let l = b(l);
                Expr::DateTime(op, l, b(r))
            }
            Expr::Array(t, items, n) => Expr::Array(t, items.into_iter().map(&mut *f).collect(), n),
            Expr::Subscript(base, subscripts, slice) => {
                let base = b(base);
                let subscripts = subscripts.into_iter().map(|(l, u)| (l.map(&mut *f), u.map(&mut *f))).collect();
                Expr::Subscript(base, subscripts, slice)
            }
            Expr::AnyArray(c, a, all) => {
                let c = b(c);
                Expr::AnyArray(c, b(a), all)
            }
            Expr::ArrayOp(op, l, r) => {
                let l = b(l);
                Expr::ArrayOp(op, l, b(r))
            }
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
            | Expr::DateTime(_, l, r)
            | Expr::AnyArray(l, r, _)
            | Expr::ArrayOp(_, l, r)
            | Expr::DistinctFrom(l, r, _) => {
                l.visit(f);
                r.visit(f);
            }
            Expr::Func(_, args) | Expr::Coalesce(args) | Expr::MinMax(_, args) | Expr::Array(_, args, _) => {
                args.iter().for_each(|a| a.visit(f))
            }
            Expr::Subscript(base, subscripts, _) => {
                base.visit(f);
                for (l, u) in subscripts {
                    l.iter().chain(u).for_each(|e| e.visit(f));
                }
            }
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

/// date_op applies a date and time operator to two non-NULL values.
fn date_op(op: DateOp, l: Value, r: Value) -> Result<Value> {
    use crate::datetime::{self as dt, USECS_PER_DAY};
    use crate::functions::datetime::{interval_multiply, justify_hours_of, negate_interval, timestamp_plus_interval};
    let date_range = || PgError::new(code::DATETIME_FIELD_OVERFLOW, "date out of range");
    let interval_range = || PgError::new(code::DATETIME_FIELD_OVERFLOW, "interval out of range");
    let float = |v: &Value| match v {
        Value::Float8(f) => *f,
        _ => 0.0,
    };
    Ok(match (op, l, r) {
        (DateOp::DatePlusDays, Value::Date(d), Value::Int4(n))
        | (DateOp::DatePlusDays, Value::Int4(n), Value::Date(d)) => {
            if d == dt::DATE_NOBEGIN || d == dt::DATE_NOEND {
                Value::Date(d)
            } else {
                Value::Date(d.checked_add(n).ok_or_else(date_range)?)
            }
        }
        (DateOp::DateMinusDays, Value::Date(d), Value::Int4(n)) => {
            if d == dt::DATE_NOBEGIN || d == dt::DATE_NOEND {
                Value::Date(d)
            } else {
                Value::Date(d.checked_sub(n).ok_or_else(date_range)?)
            }
        }
        (DateOp::DateMinusDate, Value::Date(a), Value::Date(b)) => {
            if [a, b].iter().any(|d| *d == dt::DATE_NOBEGIN || *d == dt::DATE_NOEND) {
                return Err(PgError::new(code::DATETIME_FIELD_OVERFLOW, "cannot subtract infinite dates"));
            }
            Value::Int4(a - b)
        }
        (DateOp::DatePlusTime, Value::Date(d), Value::Time(t))
        | (DateOp::DatePlusTime, Value::Time(t), Value::Date(d)) => Value::Timestamp(d as i64 * USECS_PER_DAY + t),
        (DateOp::TimestampPlusInterval(tz), Value::Timestamp(ts) | Value::TimestampTz(ts), Value::Interval(iv))
        | (DateOp::TimestampPlusInterval(tz), Value::Interval(iv), Value::Timestamp(ts) | Value::TimestampTz(ts)) => {
            let result = timestamp_plus_interval(ts, iv, tz)?;
            if tz { Value::TimestampTz(result) } else { Value::Timestamp(result) }
        }
        (DateOp::TimestampMinusInterval(tz), Value::Timestamp(ts) | Value::TimestampTz(ts), Value::Interval(iv)) => {
            let result = timestamp_plus_interval(ts, negate_interval(iv)?, tz)?;
            if tz { Value::TimestampTz(result) } else { Value::Timestamp(result) }
        }
        (
            DateOp::TimestampMinusTimestamp,
            Value::Timestamp(a) | Value::TimestampTz(a),
            Value::Timestamp(b) | Value::TimestampTz(b),
        ) => {
            if [a, b].iter().any(|t| *t == dt::TIMESTAMP_NOBEGIN || *t == dt::TIMESTAMP_NOEND) {
                return Err(PgError::new(code::DATETIME_FIELD_OVERFLOW, "cannot subtract infinite timestamps"));
            }
            let micros = a.checked_sub(b).ok_or_else(interval_range)?;
            Value::Interval(justify_hours_of(dt::Interval { months: 0, days: 0, micros }))
        }
        (DateOp::TimePlusInterval, Value::Time(t), Value::Interval(iv))
        | (DateOp::TimePlusInterval, Value::Interval(iv), Value::Time(t)) => {
            Value::Time((t + iv.micros).rem_euclid(USECS_PER_DAY))
        }
        (DateOp::TimeMinusInterval, Value::Time(t), Value::Interval(iv)) => {
            Value::Time((t - iv.micros).rem_euclid(USECS_PER_DAY))
        }
        (DateOp::TimeMinusTime, Value::Time(a), Value::Time(b)) => {
            Value::Interval(dt::Interval { months: 0, days: 0, micros: a - b })
        }
        (DateOp::IntervalPlusInterval, Value::Interval(a), Value::Interval(b)) => Value::Interval(dt::Interval {
            months: a.months.checked_add(b.months).ok_or_else(interval_range)?,
            days: a.days.checked_add(b.days).ok_or_else(interval_range)?,
            micros: a.micros.checked_add(b.micros).ok_or_else(interval_range)?,
        }),
        (DateOp::IntervalMinusInterval, Value::Interval(a), Value::Interval(b)) => Value::Interval(dt::Interval {
            months: a.months.checked_sub(b.months).ok_or_else(interval_range)?,
            days: a.days.checked_sub(b.days).ok_or_else(interval_range)?,
            micros: a.micros.checked_sub(b.micros).ok_or_else(interval_range)?,
        }),
        (DateOp::IntervalTimesFloat, Value::Interval(iv), f) | (DateOp::IntervalTimesFloat, f, Value::Interval(iv)) => {
            Value::Interval(interval_multiply(iv, float(&f), false)?)
        }
        (DateOp::IntervalDivFloat, Value::Interval(iv), f) => Value::Interval(interval_multiply(iv, float(&f), true)?),
        _ => return Err(PgError::internal("a date operator on the wrong values")),
    })
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
