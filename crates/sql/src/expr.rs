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
    /// The OID and attribute number of the table column it comes from, or zeros for any other column.
    pub origin: (u32, u16),
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
    DatePlusTimeTz,
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

/// Location is where an expression starts in the statement's text.
#[derive(Clone, Copy, Debug)]
pub struct Location(pub i32);

impl PartialEq for Location {
    /// eq treats every location as equal, since where an expression was written does not change what it computes.
    fn eq(&self, _: &Location) -> bool {
        true
    }
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
    /// A call of a user-defined function, with an argument for each of its input parameters.
    Routine(std::sync::Arc<crate::routines::Routine>, Vec<Expr>),
    /// A user-defined binary operator, by its symbol, with the routine that computes it.
    Operator(String, std::sync::Arc<crate::routines::Routine>, Box<Expr>, Box<Expr>),
    /// A field of a composite value by position.
    Field(Box<Expr>, usize),
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
    /// An array of the element type holding the values the subquery returns.
    ArraySubquery(Box<Plan>, u32),
    /// Whether the comparison holds for any, or for all, of the subquery's values.
    AnySubquery(Box<Expr>, Box<Plan>, bool),
    /// The subquery value a comparison of AnySubquery tests.
    SubqueryValue,
    /// The default of a column of the table being written, by position.
    Default(usize),
    /// A date and time operator.
    DateTime(DateOp, Box<Expr>, Box<Expr>),
    /// A window function call's result, by its position among the select list's window calls.
    WindowRef(usize),
    /// A row constructor, with each field's type and location, where a literal keeps the unknown type until the row
    /// converts to a composite type.
    Row(Vec<Expr>, Vec<(ColumnType, Location)>),
    /// A set-returning function call's current row, by its position among the select list's set-returning calls.
    SetRef(usize),
    /// An ARRAY constructor of the element type, whose items are themselves arrays when it is multidimensional.
    Array(u32, Vec<Expr>, bool),
    /// Subscripts of an array, as lower and upper bounds, which select a slice when the flag is set.
    Subscript(Box<Expr>, Vec<(Option<Expr>, Option<Expr>)>, bool),
    /// An array of the element type with a value stored at subscripts, or over a slice when the flag is set, which an
    /// UPDATE's subscripted assignment computes.
    SubscriptAssign(Box<Expr>, u32, Vec<(Option<Expr>, Option<Expr>)>, bool, Box<Expr>),
    /// A comparison against each element of an array, which holds for every element when the flag is set.
    AnyArray(Box<Expr>, Box<Expr>, bool),
    ArrayOp(ArrayOp, Box<Expr>, Box<Expr>),
    /// An expression over a value computed once, which the expression refers to as the subquery value.
    Shared(Box<Expr>, Box<Expr>),
    /// An SQL/XML expression over its arguments.
    Xml(crate::xml::sql::XmlOp, Vec<Expr>),
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
    /// The window function calls of the select list, or None where window functions aren't allowed.
    pub windows: Option<Vec<crate::window::WindowCall>>,
    /// The named windows of the query's WINDOW clause.
    pub named_windows: Vec<pg_query::protobuf::WindowDef>,
    /// The clause being bound, which errors about window functions name.
    pub clause: &'static str,
    /// The set-returning function calls of the select list, or None where they aren't allowed.
    pub set_functions: Option<Vec<Expr>>,
    /// Whether the expression is part of a definition named by `clause`, such as a default, where subqueries and
    /// aggregates aren't allowed.
    pub definition: bool,
}

impl<'b, 'a> Binder<'b, 'a> {
    /// new returns a binder over a scope.
    pub fn new(ctx: &'b mut Ctx<'a>, scope: Scope) -> Binder<'b, 'a> {
        Binder::with_scopes(ctx, vec![scope])
    }

    /// with_scopes returns a binder over the last scope inside the others.
    pub fn with_scopes(ctx: &'b mut Ctx<'a>, scopes: Vec<Scope>) -> Binder<'b, 'a> {
        Binder {
            ctx,
            scopes,
            aggregates: None,
            columns: Vec::new(),
            windows: None,
            named_windows: Vec::new(),
            clause: "this context",
            set_functions: None,
            definition: false,
        }
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
pub(crate) fn is_string(type_oid: u32) -> bool {
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

/// resolve_type_name resolves a type name with its modifiers and array bounds.
pub fn resolve_type_name(type_name: &pg_query::protobuf::TypeName) -> Result<ColumnType> {
    let names: Vec<String> = type_name.names.iter().filter_map(node_name).map(str::to_string).collect();
    let mut modifiers = Vec::new();
    for modifier in &type_name.typmods {
        match modifier.node.as_ref() {
            Some(NodeEnum::AConst(c)) => match &c.val {
                Some(Val::Ival(i)) => modifiers.push(i.ival.to_string()),
                Some(Val::Sval(s)) => modifiers.push(s.sval.clone()),
                Some(Val::Fval(f)) => modifiers.push(f.fval.clone()),
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
                let type_name = cast.type_name.as_ref().ok_or_else(|| PgError::internal("no type"))?;
                self.ctx.prepare_type(type_name)?;
                let target = resolve_type_name(type_name)?;
                let arg = cast.arg.as_deref().ok_or_else(|| PgError::internal("no cast argument"))?;
                let bound = match arg.node.as_ref() {
                    Some(NodeEnum::AArrayExpr(array))
                        if crate::array::is_vector_type(crate::usertypes::base_type(target).oid) =>
                    {
                        let vector = crate::usertypes::base_type(target);
                        let (expr, ty) = self.array_expr(array, Some(vector.oid))?;
                        let (element, from) = (typ(element_type(vector.oid)), typ(element_type(ty.oid)));
                        let expr = match expr {
                            Expr::Array(_, items, false) => {
                                let items = items
                                    .into_iter()
                                    .map(|item| coerce((item, from), element, true, cast.location).map(|b| b.0))
                                    .collect::<Result<Vec<_>>>()?;
                                Expr::Array(element.oid, items, false)
                            }
                            other => other,
                        };
                        return coerce((Expr::Cast(Box::new(expr), vector, true), vector), target, true, cast.location);
                    }
                    Some(NodeEnum::AArrayExpr(array)) if is_array_type(target.oid) => {
                        self.array_expr(array, Some(target.oid))?
                    }
                    _ => self.bind(arg)?,
                };
                if let Expr::Param(i) = bound.0
                    && self.ctx.parameters[i] == 0
                {
                    self.ctx.parameters[i] = target.oid;
                }
                if crate::cast::is_reg_type(target.oid)
                    && let (Expr::Const(value), oid::UNKNOWN) = (&bound.0, bound.1.oid)
                {
                    let value = self
                        .ctx
                        .reg_value(value.clone(), target.oid)
                        .map_err(|err| PgError { position: position(arg_location(arg)), ..err })?;
                    return Ok((Expr::Const(value), target));
                }
                let location = if bound.1.oid == oid::UNKNOWN { arg_location(arg) } else { cast.location };
                coerce(bound, target, true, location)
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
            NodeEnum::XmlExpr(x) => self.xml_expr(x),
            NodeEnum::XmlSerialize(x) => self.xml_serialize(x),
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
            NodeEnum::RowExpr(row) => {
                let (mut fields, mut types) = (Vec::with_capacity(row.args.len()), Vec::with_capacity(row.args.len()));
                for arg in &row.args {
                    if let Some(NodeEnum::ColumnRef(column)) = arg.node.as_ref()
                        && let [table, star] = column.fields.as_slice()
                        && matches!(star.node, Some(NodeEnum::AStar(_)))
                        && let Some(table) = node_name(table)
                    {
                        for (_, expr, ty) in self.whole_row_columns(table) {
                            fields.push(expr);
                            types.push((ty, Location(arg_location(arg))));
                        }
                        continue;
                    }
                    let (expr, ty) = self.bind(arg)?;
                    types.push((ty, Location(arg_location(arg))));
                    let ty = if ty.oid == oid::UNKNOWN { typ(oid::TEXT) } else { ty };
                    fields.push(coerce((expr, ty), ty, false, arg_location(arg))?.0);
                }
                Ok((Expr::Row(fields, types), typ(oid::RECORD)))
            }
            NodeEnum::AIndirection(indirection) => self.indirection(indirection),
            NodeEnum::NullTest(test) => {
                let arg = test.arg.as_deref().ok_or_else(|| PgError::internal("no null test argument"))?;
                let (expr, _) = self.bind(arg)?;
                let negated = NullTestType::try_from(test.nulltesttype) == Ok(NullTestType::IsNotNull);
                Ok((Expr::IsNull(Box::new(expr), negated), typ(oid::BOOL)))
            }
            NodeEnum::CollateClause(collate) => self.collate(collate),
            _ => Err(PgError::unsupported(format!("the expression {}", node_kind(node)))),
        }
    }

    /// collate binds an expression with a COLLATE clause, which keeps its value and type, since every collation
    /// sorts as C does, after checking that the collation exists and that the type has collations.
    fn collate(&mut self, collate: &pg_query::protobuf::CollateClause) -> Result<Bound> {
        let arg = collate.arg.as_deref().ok_or_else(|| PgError::internal("no collated value"))?;
        let bound = self.bind(arg)?;
        let name = collate.collname.iter().filter_map(node_name).next_back().unwrap_or_default();
        let known = crate::pgcatalog::reg::builtin_column("pg_collation", "collname")
            .iter()
            .any(|(_, n)| n.output().is_some_and(|n| n == name));
        if !known {
            return Err(PgError {
                position: position(collate.location),
                ..PgError::new(
                    code::UNDEFINED_OBJECT,
                    format!("collation \"{name}\" for encoding \"UTF8\" does not exist"),
                )
            });
        }
        let base = crate::usertypes::base_type(bound.1).oid;
        let element = if is_array_type(base) { element_type(base) } else { base };
        if !is_string(element) && element != oid::UNKNOWN {
            return Err(PgError {
                position: position(collate.location),
                ..PgError::new(
                    code::DATATYPE_MISMATCH,
                    format!("collations are not supported by type {}", type_display(bound.1.oid)),
                )
            });
        }
        Ok(bound)
    }

    /// column binds a column reference.
    fn column(&mut self, column: &pg_query::protobuf::ColumnRef) -> Result<Bound> {
        let names: Vec<&str> = column.fields.iter().filter_map(node_name).collect();
        let star = matches!(column.fields.last().and_then(|f| f.node.as_ref()), Some(NodeEnum::AStar(_)));
        if names.len() + usize::from(star) != column.fields.len() || (star && names.len() != 1) {
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
        for depth in (0..depth_count).filter(|_| !star) {
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
        if let [name] = names.as_slice() {
            let columns = self.whole_row_columns(name);
            if !columns.is_empty() {
                let types: Vec<(String, ColumnType)> = columns.iter().map(|(n, _, t)| (n.clone(), *t)).collect();
                let type_oid = match self.whole_row_table(name).and_then(crate::usertypes::table_row_type) {
                    Some(type_oid) => type_oid,
                    None => crate::usertypes::transient(name, &types),
                };
                let row = Expr::Row(
                    columns.iter().map(|(_, e, _)| e.clone()).collect(),
                    columns.iter().map(|(_, _, t)| (*t, Location(-1))).collect(),
                );
                return Ok((Expr::Cast(Box::new(row), typ(type_oid), false), typ(type_oid)));
            }
        }
        if star {
            return Err(PgError::unsupported("this column reference"));
        }
        if let Some((routine, params)) = &self.ctx.named_params
            && let Some(index) = match names.as_slice() {
                [name] => params.iter().position(|p| p == name),
                [function, name] if function == routine => params.iter().position(|p| p == name),
                _ => None,
            }
        {
            return Ok((Expr::Param(index), typ(self.ctx.parameters.get(index).copied().unwrap_or(oid::UNKNOWN))));
        }
        if self.definition && self.scopes.iter().all(|s| s.columns.is_empty()) {
            return Err(PgError {
                position: position(column.location),
                ..PgError::new(
                    code::FEATURE_NOT_SUPPORTED,
                    format!("cannot use column reference in {}", singular(self.clause)),
                )
            });
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
    /// whole_row_columns returns the visible columns of the FROM item a name refers to, or its hidden ones when it has
    /// only those, as a trigger condition's `old` and `new` do, with their expressions, in the innermost scope that has
    /// it, which a whole-row reference such as `t` or `t.*` stands for.
    fn whole_row_columns(&mut self, table: &str) -> Vec<(String, Expr, ColumnType)> {
        let depth_count = self.scopes.len();
        for depth in 0..depth_count {
            let scope = &self.scopes[depth_count - 1 - depth];
            let only_hidden = scope.columns.iter().filter(|c| c.table == table).all(|c| c.hidden);
            let columns: Vec<(usize, String, ColumnType)> = scope
                .columns
                .iter()
                .enumerate()
                .filter(|(_, c)| c.table == table && (only_hidden || !c.hidden))
                .map(|(i, c)| (i, c.name.clone(), c.ty))
                .collect();
            if columns.is_empty() {
                continue;
            }
            return columns
                .into_iter()
                .map(|(i, name, ty)| {
                    let expr = if depth == 0 {
                        self.columns.push((i, -1));
                        Expr::Column(i)
                    } else {
                        Expr::Outer(depth, i)
                    };
                    (name, expr, ty)
                })
                .collect();
        }
        Vec::new()
    }

    /// whole_row_table returns the OID of the table whose every column, in order, a whole-row reference covers.
    fn whole_row_table(&self, table: &str) -> Option<u32> {
        let scope = self.scopes.iter().rev().find(|s| s.columns.iter().any(|c| c.table == table))?;
        let only_hidden = scope.columns.iter().filter(|c| c.table == table).all(|c| c.hidden);
        let origins: Vec<(u32, u16)> =
            scope.columns.iter().filter(|c| c.table == table && (only_hidden || !c.hidden)).map(|c| c.origin).collect();
        let table_oid = origins.first()?.0;
        let in_order = origins.iter().enumerate().all(|(i, o)| o.0 == table_oid && usize::from(o.1) == i + 1);
        (table_oid != 0 && in_order).then_some(table_oid)
    }

    /// func_call binds a call of a built-in function.
    fn func_call(&mut self, call: &pg_query::protobuf::FuncCall) -> Result<Bound> {
        let names: Vec<&str> = call.funcname.iter().filter_map(node_name).collect();
        let (schema, name) = match names.as_slice() {
            [name] => (None, *name),
            [schema, name] | [_, schema, name] => (Some(*schema), *name),
            _ => {
                return Err(PgError {
                    position: position(call.location),
                    ..PgError::new(code::UNDEFINED_FUNCTION, format!("function {}() does not exist", names.join(".")))
                });
            }
        };
        if name == "pg_typeof" && schema.is_none_or(|s| s == "pg_catalog") && call.args.len() == 1 {
            let (_, ty) = self.bind(&call.args[0])?;
            let reg = crate::types::Reg {
                type_oid: oid::REGTYPE,
                oid: ty.oid,
                name: crate::cast::type_display(ty.oid).into_owned(),
            };
            return Ok((Expr::Const(Value::Reg(Box::new(reg))), typ(oid::REGTYPE)));
        }
        if matches!(name, "enum_first" | "enum_last" | "enum_range")
            && schema.is_none_or(|s| s == "pg_catalog")
            && call.args.len() == 1
        {
            let (_, ty) = self.bind(&call.args[0])?;
            if let Some(user_type) = crate::usertypes::get(ty.oid)
                && let crate::usertypes::Kind::Enum(labels) = &user_type.kind
            {
                let label = |l: &String| {
                    Value::Enum(Box::new(crate::types::EnumValue { type_oid: user_type.oid, label: l.clone() }))
                };
                return Ok(match name {
                    "enum_first" => (Expr::Const(labels.first().map_or(Value::Null, label)), ty),
                    "enum_last" => (Expr::Const(labels.last().map_or(Value::Null, label)), ty),
                    _ => {
                        let array = crate::array::Array::one_dimensional(ty.oid, labels.iter().map(label).collect());
                        (Expr::Const(Value::Array(Box::new(array))), typ(user_type.array))
                    }
                });
            }
        }
        if call.over.is_none() && !call.agg_star {
            let routines = self.ctx.routines_named(schema, name)?;
            if (!routines.is_empty()
                || schema.is_some_and(|s| s != "pg_catalog") && !crate::aggregates::exists(schema, name))
                && let Some(bound) = self.routine_call(call, schema, name, routines, false)?
            {
                return Ok(bound);
            }
        }
        if call.over.is_some() {
            let wrong = |message: String| PgError {
                position: position(call.location),
                ..PgError::new(code::WRONG_OBJECT_TYPE, message)
            };
            if call.agg_distinct && crate::aggregates::exists(schema, name) {
                return Err(PgError {
                    position: position(call.location),
                    ..PgError::new(code::FEATURE_NOT_SUPPORTED, "DISTINCT is not implemented for window functions")
                });
            }
            if schema.is_some_and(|s| s != "pg_catalog")
                && !crate::aggregates::exists(schema, name)
                && !self.ctx.routines_named(schema, name)?.is_empty()
            {
                let display = names.join(".");
                if call.agg_distinct {
                    return Err(wrong(format!("DISTINCT specified, but {display} is not an aggregate function")));
                }
                return Err(wrong(format!(
                    "OVER specified, but {display} is not a window function nor an aggregate function"
                )));
            }
            return self.window_call(name, call);
        }
        if crate::window::is_window_function(name) {
            return Err(PgError {
                position: position(call.location),
                ..PgError::new(code::WRONG_OBJECT_TYPE, format!("window function {name} requires an OVER clause"))
            });
        }
        if call.agg_star || functions::aggregate::exists(schema, name) {
            return self.aggregate_call(schema, name, call);
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
            let (expr, ty) = self.bind(arg)?;
            bound.push((expr, crate::usertypes::base_type(ty)));
        }
        let types: Vec<u32> = bound.iter().map(|(_, t)| t.oid).collect();
        let resolved = functions::resolve(name, &types, call.location)?;
        if matches!(name, "nextval" | "currval" | "setval")
            && let Some((Expr::Const(Value::Text(text)), _)) = bound.first()
        {
            self.ctx.resolve_sequence(text, arg_location(&call.args[0]))?;
        }
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
        let mut call_expr = Expr::Func(resolved.index, args);
        if functions::function(resolved.index).ret == functions::ANYARRAY
            && resolved.ret != functions::ANYARRAY
            && !functions::returns_set(name)
        {
            call_expr = Expr::Cast(Box::new(call_expr), typ(resolved.ret), false);
        }
        if functions::returns_set(name) {
            let Some(set_functions) = self.set_functions.as_mut() else {
                return Err(PgError {
                    position: position(call.location),
                    ..PgError::new(
                        code::FEATURE_NOT_SUPPORTED,
                        format!("set-returning functions are not allowed in {}", self.clause),
                    )
                });
            };
            set_functions.push(call_expr);
            return Ok((Expr::SetRef(set_functions.len() - 1), typ(resolved.ret)));
        }
        Ok((call_expr, typ(resolved.ret)))
    }
}

impl<'b, 'a> Binder<'b, 'a> {
    /// aggregate_call binds a call of an aggregate in a grouped query, whose arguments are over the input rows.
    fn aggregate_call(
        &mut self,
        schema: Option<&str>,
        name: &str,
        call: &pg_query::protobuf::FuncCall,
    ) -> Result<Bound> {
        if self.definition {
            return Err(PgError {
                position: position(call.location),
                ..PgError::new(code::GROUPING_ERROR, format!("aggregate functions are not allowed in {}", self.clause))
            });
        }
        let Some(mut aggregates) = self.aggregates.take() else {
            return Err(PgError {
                position: position(call.location),
                ..PgError::new(code::GROUPING_ERROR, "aggregate function calls cannot be nested")
            });
        };
        let result = (|| -> Result<(AggCall, u32)> {
            let mut bound = Vec::with_capacity(call.args.len());
            for arg in &call.args {
                let (expr, ty) = self.bind(arg)?;
                bound.push((expr, crate::usertypes::base_type(ty)));
            }
            let types: Vec<u32> = bound.iter().map(|(_, t)| t.oid).collect();
            let user = crate::aggregates::find(schema, name, &types);
            let (index, arg_types, ret) = match &user {
                Some(user) => (0, user.params.clone(), user.ret.oid),
                None => functions::aggregate::resolve(name, &types, call.location)?,
            };
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
            Ok((AggCall { index, args, distinct: call.agg_distinct, filter, order, ret, user }, ret))
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

/// row_items returns the fields of a row constructor, `ROW(...)` or `(a, b)`, or None for any other expression.
fn row_items(node: &Node) -> Option<&[Node]> {
    match node.node.as_ref() {
        Some(NodeEnum::RowExpr(row)) => Some(&row.args),
        _ => None,
    }
}

/// unequal_rows returns Postgres' error for comparing row constructors of different lengths.
fn unequal_rows(location: i32) -> PgError {
    PgError {
        position: position(location),
        ..PgError::new(code::SYNTAX_ERROR, "unequal number of entries in row expressions")
    }
}

/// zero_length_rows returns Postgres' error for comparing empty row constructors.
fn zero_length_rows(location: i32) -> PgError {
    PgError {
        position: position(location),
        ..PgError::new(code::FEATURE_NOT_SUPPORTED, "cannot compare rows of zero length")
    }
}

impl Binder<'_, '_> {
    /// row_compare binds a comparison of two row constructors field by field, as Postgres does: equality holds when
    /// every pair is equal, inequality when any pair differs, and an ordering is decided by the first pair that differs,
    /// with NULL wherever a NULL field leaves the answer unknown.
    fn row_compare(&mut self, op: &str, left: &[Node], right: &[Node], location: i32) -> Result<Expr> {
        if left.len() != right.len() {
            return Err(unequal_rows(location));
        }
        let left = left.iter().map(|n| self.bind(n)).collect::<Result<Vec<_>>>()?;
        let right = right.iter().map(|n| self.bind(n)).collect::<Result<Vec<_>>>()?;
        self.compare_rows(op, left, right, location)
    }

    /// compare_rows binds a field-by-field comparison of two rows of bound fields, as row_compare describes.
    fn compare_rows(&mut self, op: &str, left: Vec<Bound>, right: Vec<Bound>, location: i32) -> Result<Expr> {
        if left.len() != right.len() {
            return Err(unequal_rows(location));
        }
        if left.is_empty() {
            return Err(zero_length_rows(location));
        }
        let pair = |binder: &mut Self, op: &str, i: usize| -> Result<Expr> {
            Ok(binder.binary(op, left[i].clone(), right[i].clone(), location)?.0)
        };
        let and = |l: Expr, r: Expr| Expr::And(Box::new(l), Box::new(r));
        let or = |l: Expr, r: Expr| Expr::Or(Box::new(l), Box::new(r));
        let last = left.len() - 1;
        match op {
            "=" => {
                let mut result = pair(self, "=", 0)?;
                for i in 1..=last {
                    result = and(result, pair(self, "=", i)?);
                }
                Ok(result)
            }
            "<>" | "!=" => {
                let mut result = pair(self, "<>", 0)?;
                for i in 1..=last {
                    result = or(result, pair(self, "<>", i)?);
                }
                Ok(result)
            }
            "<" | "<=" | ">" | ">=" => {
                let strict = if op.starts_with('<') { "<" } else { ">" };
                let mut result = pair(self, op, last)?;
                for i in (0..last).rev() {
                    let decided = pair(self, strict, i)?;
                    let equal = pair(self, "=", i)?;
                    result = or(decided, and(equal, result));
                }
                Ok(result)
            }
            _ => Err(PgError {
                position: position(location),
                ..PgError::new(
                    code::UNDEFINED_FUNCTION,
                    format!("could not determine interpretation of row comparison operator {op}"),
                )
            }),
        }
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
            AExprKind::AexprOp
                if let [schema, _] = e.name.iter().filter_map(node_name).collect::<Vec<_>>().as_slice()
                    && !matches!(*schema, "pg_catalog" | "information_schema" | "pg_toast")
                    && !self.ctx.schema_names().iter().any(|s| s == schema) =>
            {
                Err(PgError {
                    position: position(e.location),
                    ..PgError::new(code::INVALID_SCHEMA_NAME, format!("schema \"{schema}\" does not exist"))
                })
            }
            AExprKind::AexprOp => {
                if e.lexpr.is_none() {
                    let right = self.bind(operand(&e.rexpr)?)?;
                    if let Some(operator) = self.user_operator(&op, 0, right.1.oid)? {
                        let right = coerce(right, typ(operator.right), false, e.location)?.0;
                        return Ok((Expr::Routine(operator.routine.clone(), vec![right]), operator.routine.ret));
                    }
                    return unary(&op, right, e.location);
                }
                if let (Some(left), Some(right)) = (row_items(operand(&e.lexpr)?), row_items(operand(&e.rexpr)?)) {
                    return Ok((self.row_compare(&op, left, right, e.location)?, typ(oid::BOOL)));
                }
                if let (Some(items), Some(NodeEnum::SubLink(link))) =
                    (row_items(operand(&e.lexpr)?), operand(&e.rexpr)?.node.as_ref())
                    && link.sub_link_type == pg_query::protobuf::SubLinkType::ExprSublink as i32
                    && let Some(NodeEnum::SelectStmt(select)) = link.subselect.as_deref().and_then(|n| n.node.as_ref())
                {
                    let query = Planner { ctx: &mut *self.ctx, outer: self.scopes.clone() }.plan_query(select)?;
                    if query.columns.len() != items.len() {
                        let message = if query.columns.len() > items.len() {
                            "subquery has too many columns"
                        } else {
                            "subquery has too few columns"
                        };
                        return Err(PgError {
                            position: position(e.location),
                            ..PgError::new(code::SYNTAX_ERROR, message)
                        });
                    }
                    let left = items.iter().map(|n| self.bind(n)).collect::<Result<Vec<_>>>()?;
                    let subquery = Expr::Scalar(Box::new(query.plan));
                    let right = (0..items.len())
                        .map(|i| (Expr::Field(Box::new(subquery.clone()), i), query.types[i]))
                        .collect();
                    return Ok((self.compare_rows(&op, left, right, e.location)?, typ(oid::BOOL)));
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
                let mut shared = None;
                for item in &list.items {
                    if let (Some(left), Some(right)) = (row_items(left_node), row_items(item)) {
                        let test = self.row_compare(cmp, left, right, e.location)?;
                        result = Some(match result {
                            Some(previous) => join(previous, test),
                            None => test,
                        });
                        continue;
                    }
                    let mut left = self.bind(left_node)?;
                    if !matches!(left.0, Expr::Column(_) | Expr::Const(_) | Expr::Param(_)) {
                        shared = Some(left.0);
                        left.0 = Expr::SubqueryValue;
                    }
                    let right = self.bind(item)?;
                    let (test, _) = self.binary(cmp, left, right, e.location)?;
                    result = Some(match result {
                        Some(previous) => join(previous, test),
                        None => test,
                    });
                }
                let result = result.ok_or_else(|| PgError::internal("an empty IN list"))?;
                match shared {
                    Some(value) => Ok((Expr::Shared(Box::new(value), Box::new(result)), typ(oid::BOOL))),
                    None => Ok((result, typ(oid::BOOL))),
                }
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
            AExprKind::AexprLike | AExprKind::AexprIlike | AExprKind::AexprSimilar => {
                let left = self.bind(operand(&e.lexpr)?)?;
                let right = self.bind(operand(&e.rexpr)?)?;
                self.binary(&op, left, right, e.location)
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
                if let (Some(left), Some(right)) = (row_items(operand(&e.lexpr)?), row_items(operand(&e.rexpr)?)) {
                    if left.len() != right.len() {
                        return Err(unequal_rows(e.location));
                    }
                    let mut result: Option<Expr> = None;
                    for (l, r) in left.iter().zip(right) {
                        let (l, r) = (self.bind(l)?, self.bind(r)?);
                        let (test, _) = self.binary("=", l, r, e.location)?;
                        let Expr::Compare(_, l, r) = test else { return Err(PgError::internal("a distinct test")) };
                        let distinct = Expr::DistinctFrom(l, r, false);
                        result = Some(match result {
                            Some(previous) => Expr::Or(Box::new(previous), Box::new(distinct)),
                            None => distinct,
                        });
                    }
                    let result = result.ok_or_else(|| zero_length_rows(e.location))?;
                    let result = if kind == AExprKind::AexprNotDistinct { Expr::Not(Box::new(result)) } else { result };
                    return Ok((result, typ(oid::BOOL)));
                }
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
        for ((expr, _), _) in &bound {
            if let Expr::Param(i) = expr
                && self.ctx.parameters[*i] == 0
            {
                self.ctx.parameters[*i] = ty.oid;
            }
        }
        let args = bound.into_iter().map(|(b, l)| coerce(b, ty, false, l).map(|b| b.0)).collect::<Result<_>>()?;
        Ok((args, ty))
    }

    /// xml_arg binds an argument of an SQL/XML expression that must be of a type, converting an untyped literal and,
    /// for a text argument, any other value.
    pub fn xml_arg(&mut self, node: &Node, wanted: u32, construct: &str) -> Result<Expr> {
        let location = arg_location(node);
        let bound = self.bind(node)?;
        if bound.1.oid == wanted {
            return Ok(bound.0);
        }
        if bound.1.oid == oid::UNKNOWN || wanted == oid::TEXT {
            return Ok(coerce(bound, typ(wanted), true, location)?.0);
        }
        Err(crate::xml::sql::wrong_type(
            construct,
            &crate::cast::type_display(wanted),
            &crate::cast::type_display(bound.1.oid),
            position(location),
        ))
    }

    /// reg_literal reads an untyped literal as a value of a reg type when that is the type it is assigned to, which
    /// looks the name up in the catalog.
    pub fn reg_literal(&mut self, bound: Bound, target: ColumnType, location: i32) -> Result<Bound> {
        match (&bound.0, bound.1.oid) {
            (Expr::Const(value), oid::UNKNOWN) if crate::cast::is_reg_type(target.oid) => {
                let value = self
                    .ctx
                    .reg_value(value.clone(), target.oid)
                    .map_err(|err| PgError { position: position(location), ..err })?;
                Ok((Expr::Const(value), target))
            }
            _ => Ok(bound),
        }
    }

    /// typed_arg binds an argument of a construct that must be of a type, converting an untyped literal and any value
    /// that converts implicitly.
    pub fn typed_arg(&mut self, node: &Node, wanted: ColumnType, construct: &str) -> Result<Expr> {
        let location = arg_location(node);
        let bound = self.bind(node)?;
        if bound.1.oid == wanted.oid
            || bound.1.oid == oid::UNKNOWN
            || functions::implicitly_castable(bound.1.oid, wanted.oid)
        {
            return Ok(coerce(bound, wanted, false, location)?.0);
        }
        Err(crate::xml::sql::wrong_type(
            construct,
            &crate::cast::type_display(wanted.oid),
            &crate::cast::type_display(bound.1.oid),
            position(location),
        ))
    }

    /// xml_names binds the named arguments of XMLELEMENT's attributes or of XMLFOREST, whose names default to the
    /// names of the columns they refer to.
    fn xml_names(&mut self, nodes: &[Node], what: &str, unique: bool) -> Result<(Vec<String>, Vec<Expr>)> {
        let (mut names, mut args) = (Vec::new(), Vec::new());
        for node in nodes {
            let Some(NodeEnum::ResTarget(target)) = node.node.as_ref() else { continue };
            let value = target.val.as_deref().ok_or_else(|| PgError::internal("an XML argument without a value"))?;
            let name = if target.name.is_empty() {
                match value.node.as_ref() {
                    Some(NodeEnum::ColumnRef(c)) => c.fields.last().and_then(node_name).unwrap_or_default().to_string(),
                    _ => {
                        return Err(PgError {
                            position: position(target.location),
                            ..PgError::new(
                                code::SYNTAX_ERROR,
                                format!("unnamed XML {what} value must be a column reference"),
                            )
                        });
                    }
                }
            } else {
                target.name.clone()
            };
            let name = crate::xml::sql::escape_name(&name);
            if unique && names.contains(&name) {
                return Err(PgError {
                    position: position(target.location),
                    ..PgError::new(code::SYNTAX_ERROR, format!("XML attribute name \"{name}\" appears more than once"))
                });
            }
            names.push(name);
            args.push(self.bind(value)?.0);
        }
        Ok((names, args))
    }

    /// xml_expr binds an SQL/XML expression.
    fn xml_expr(&mut self, x: &pg_query::protobuf::XmlExpr) -> Result<Bound> {
        use crate::xml::sql::XmlOp;
        use pg_query::protobuf::{XmlExprOp, XmlOptionType};
        let xml = typ(oid::XML);
        Ok(match XmlExprOp::try_from(x.op) {
            Ok(XmlExprOp::IsXmlelement) => {
                let (attributes, mut args) = self.xml_names(&x.named_args, "attribute", true)?;
                for arg in &x.args {
                    args.push(self.bind(arg)?.0);
                }
                (Expr::Xml(XmlOp::Element { name: crate::xml::sql::escape_name(&x.name), attributes }, args), xml)
            }
            Ok(XmlExprOp::IsXmlforest) => {
                let (names, args) = self.xml_names(&x.named_args, "element", false)?;
                (Expr::Xml(XmlOp::Forest(names), args), xml)
            }
            Ok(XmlExprOp::IsXmlconcat) => {
                let args = x.args.iter().map(|a| self.xml_arg(a, oid::XML, "XMLCONCAT")).collect::<Result<_>>()?;
                (Expr::Xml(XmlOp::Concat, args), xml)
            }
            Ok(XmlExprOp::IsXmlparse) => {
                let document = x.xmloption == XmlOptionType::XmloptionDocument as i32;
                let arg = self.xml_arg(&x.args[0], oid::TEXT, "XMLPARSE")?;
                (Expr::Xml(XmlOp::Parse { document }, vec![arg]), xml)
            }
            Ok(XmlExprOp::IsXmlpi) => {
                let target = crate::xml::sql::escape_name(&x.name);
                if target.eq_ignore_ascii_case("xml") {
                    return Err(PgError {
                        detail: Some(format!("XML processing instruction target name cannot be \"{target}\".")),
                        ..PgError::new(code::SYNTAX_ERROR, "invalid XML processing instruction")
                    });
                }
                let args = x.args.iter().map(|a| self.xml_arg(a, oid::TEXT, "XMLPI")).collect::<Result<_>>()?;
                (Expr::Xml(XmlOp::Pi(target), args), xml)
            }
            Ok(XmlExprOp::IsXmlroot) => {
                let value = self.xml_arg(&x.args[0], oid::XML, "XMLROOT")?;
                let version = self.xml_arg(&x.args[1], oid::TEXT, "XMLROOT")?;
                let standalone = match x.args.get(2).and_then(|a| a.node.as_ref()) {
                    Some(NodeEnum::AConst(c)) => match &c.val {
                        Some(Val::Ival(i)) => match i.ival {
                            0 => Some(Some(true)),
                            1 => Some(Some(false)),
                            2 => Some(None),
                            _ => None,
                        },
                        _ => None,
                    },
                    _ => None,
                };
                (Expr::Xml(XmlOp::Root(standalone), vec![value, version]), xml)
            }
            Ok(XmlExprOp::IsDocument) => {
                let arg = self.xml_arg(&x.args[0], oid::XML, "IS DOCUMENT")?;
                (Expr::Xml(XmlOp::IsDocument, vec![arg]), typ(oid::BOOL))
            }
            _ => return Err(PgError::unsupported("this XML expression")),
        })
    }

    /// xml_serialize binds XMLSERIALIZE, whose result converts to a character type.
    fn xml_serialize(&mut self, x: &pg_query::protobuf::XmlSerialize) -> Result<Bound> {
        use crate::xml::sql::XmlOp;
        let arg = x.expr.as_deref().ok_or_else(|| PgError::internal("XMLSERIALIZE without a value"))?;
        let value = self.xml_arg(arg, oid::XML, "XMLSERIALIZE")?;
        let type_name = x.type_name.as_ref().ok_or_else(|| PgError::internal("XMLSERIALIZE without a type"))?;
        let target = resolve_type_name(type_name)?;
        if !matches!(target.oid, oid::TEXT | oid::VARCHAR | oid::BPCHAR | oid::NAME) {
            return Err(PgError {
                position: position(x.location),
                ..PgError::new(
                    code::CANNOT_COERCE,
                    format!("cannot cast XMLSERIALIZE result to {}", crate::cast::type_display(target.oid)),
                )
            });
        }
        let document = x.xmloption == pg_query::protobuf::XmlOptionType::XmloptionDocument as i32;
        let serialized = Expr::Xml(XmlOp::Serialize { document }, vec![value]);
        Ok((Expr::Cast(Box::new(serialized), target, false), target))
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
            types.insert(0, (*t, *l));
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
        if self.definition {
            return Err(PgError {
                position: position(link.location),
                ..PgError::new(code::FEATURE_NOT_SUPPORTED, format!("cannot use subquery in {}", singular(self.clause)))
            });
        }
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
                let test =
                    link.testexpr.as_deref().ok_or_else(|| PgError::internal("a subquery test without a value"))?;
                if let Some(items) = row_items(test) {
                    if items.len() != query.columns.len() {
                        let message = if query.columns.len() > items.len() {
                            "subquery has too many columns"
                        } else {
                            "subquery has too few columns"
                        };
                        return Err(PgError {
                            position: position(link.location),
                            ..PgError::new(code::SYNTAX_ERROR, message)
                        });
                    }
                    let op = link.oper_name.iter().filter_map(node_name).next_back().unwrap_or("=").to_string();
                    let left = items.iter().map(|n| self.bind(n)).collect::<Result<Vec<_>>>()?;
                    let right = (0..items.len())
                        .map(|i| (Expr::Field(Box::new(Expr::SubqueryValue), i), query.types[i]))
                        .collect();
                    let comparison = self.compare_rows(&op, left, right, link.location)?;
                    return Ok((
                        Expr::AnySubquery(Box::new(comparison), Box::new(query.plan), kind == T::AllSublink),
                        typ(oid::BOOL),
                    ));
                }
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
            T::ArraySublink => {
                if query.columns.len() != 1 {
                    return Err(PgError {
                        position: position(link.location),
                        ..PgError::new(code::SYNTAX_ERROR, "subquery must return only one column")
                    });
                }
                let element = query.types[0].oid;
                let result = if is_array_type(element) { element } else { array_of(element) };
                Ok((Expr::ArraySubquery(Box::new(query.plan), element), typ(result)))
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

    /// user_operator returns the visible stored operator of the name for operands of the types, where the left type
    /// is 0 for a prefix operator. An untyped operand first takes the other operand's type, and for operators that
    /// Postgres lacks it then matches any type, preferring text when several operators match, as Postgres'
    /// oper_select_candidate does.
    fn user_operator(
        &mut self,
        op: &str,
        lt: u32,
        rt: u32,
    ) -> Result<Option<std::sync::Arc<crate::operators::UserOperator>>> {
        let operators = self.ctx.user_operators()?;
        let named: Vec<_> = operators
            .iter()
            .filter(|o| o.name == op && (o.left == 0) == (lt == 0) && crate::usertypes::in_search_path(&o.schema))
            .collect();
        let (exact_left, exact_right) = match (lt == oid::UNKNOWN, rt == oid::UNKNOWN) {
            (true, false) => (rt, rt),
            (false, true) => (lt, lt),
            _ => (lt, rt),
        };
        if let Some(found) = named.iter().find(|o| o.left == exact_left && o.right == exact_right) {
            return Ok(Some((*found).clone()));
        }
        if crate::operators::is_builtin(op) {
            return Ok(None);
        }
        let fits = |given: u32, operand: u32| given == operand || given == oid::UNKNOWN;
        let candidates: Vec<_> = named.into_iter().filter(|o| fits(lt, o.left) && fits(rt, o.right)).collect();
        let preferred = |given: u32, operand: u32| given != oid::UNKNOWN || operand == oid::TEXT;
        Ok(match candidates.as_slice() {
            [only] => Some((*only).clone()),
            _ => candidates.iter().find(|o| preferred(lt, o.left) && preferred(rt, o.right)).map(|o| (*o).clone()),
        })
    }

    /// operator_call binds a binary operator that a built-in function implements.
    fn operator_call(
        &mut self,
        resolved: functions::Resolved,
        left: Bound,
        right: Bound,
        location: i32,
    ) -> Result<Bound> {
        let mut args = Vec::with_capacity(2);
        for (bound, &target) in [left, right].into_iter().zip(&resolved.arg_types) {
            if let Expr::Param(i) = bound.0
                && self.ctx.parameters[i] == 0
            {
                self.ctx.parameters[i] = target;
            }
            args.push(coerce(bound, typ(target), false, location)?.0);
        }
        Ok((Expr::Func(resolved.index, args), typ(resolved.ret)))
    }

    /// binary binds a binary operator, resolving its operand types as Postgres does for the built-in operators.
    fn binary(&mut self, op: &str, left: Bound, right: Bound, location: i32) -> Result<Bound> {
        let mut left = (left.0, crate::usertypes::base_type(left.1));
        let mut right = (right.0, crate::usertypes::base_type(right.1));
        if is_composite(left.1.oid) && right.1.oid == oid::RECORD {
            right = coerce(right, left.1, false, location)?;
        } else if is_composite(right.1.oid) && left.1.oid == oid::RECORD {
            left = coerce(left, right.1, false, location)?;
        }
        let (lt, rt) = (left.1.oid, right.1.oid);
        if let Some(operator) = self.user_operator(op, lt, rt)? {
            let left = coerce(left, typ(operator.left), false, location)?.0;
            let right = coerce(right, typ(operator.right), false, location)?.0;
            let call = Expr::Operator(operator.name.clone(), operator.routine.clone(), Box::new(left), Box::new(right));
            return Ok((call, operator.routine.ret));
        }
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
        if !crate::operators::is_builtin(op) && op != "!=" {
            return Err(missing());
        }
        if [lt, rt].iter().any(|t| matches!(*t, oid::JSON | oid::JSONB)) {
            if functions::exists(op)
                && let Ok(resolved) = functions::resolve(op, &[lt, rt], location)
            {
                return self.operator_call(resolved, left, right, location);
            }
            if lt == oid::JSON || rt == oid::JSON || !matches!(op, "=" | "<>" | "!=" | "<" | "<=" | ">" | ">=") {
                return Err(missing());
            }
        }
        if (lt == oid::XML || rt == oid::XML) && matches!(op, "=" | "<>" | "!=" | "<" | "<=" | ">" | ">=") {
            return Err(missing());
        }
        if let Some(function) = pattern_function(op)
            && lt != oid::BYTEA
            && rt != oid::BYTEA
        {
            let types = [lt, rt].map(|t| if t == oid::UNKNOWN { oid::TEXT } else { t });
            if !types.iter().all(|&t| is_string(t)) {
                return Err(missing());
            }
            let resolved = functions::resolve(function, &[oid::TEXT, oid::TEXT], location)?;
            for bound in [&left, &right] {
                if let Expr::Param(i) = bound.0
                    && self.ctx.parameters[i] == 0
                {
                    self.ctx.parameters[i] = oid::TEXT;
                }
            }
            let args = vec![
                coerce(left, typ(oid::TEXT), false, location)?.0,
                coerce(right, typ(oid::TEXT), false, location)?.0,
            ];
            return Ok((Expr::Func(resolved.index, args), typ(oid::BOOL)));
        }
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
        let textual = |t: u32| is_string(t) || t == oid::UNKNOWN;
        if !(textual(lt) && textual(rt))
            && functions::exists(op)
            && let Ok(resolved) = functions::resolve(op, &[lt, rt], location)
        {
            return self.operator_call(resolved, left, right, location);
        }
        let other = |t: u32| !textual(t) && !matches!(t, oid::JSON | oid::JSONB) && !is_array_type(t);
        if op == "||" && ((textual(lt) && other(rt)) || (other(lt) && textual(rt))) {
            let left = coerce(left, typ(oid::TEXT), true, location)?.0;
            let right = coerce(right, typ(oid::TEXT), true, location)?.0;
            return Ok((Expr::Concat(Box::new(left), Box::new(right)), typ(oid::TEXT)));
        }
        let integer = |t: u32| matches!(t, oid::INT2 | oid::INT4 | oid::INT8);
        if matches!(op, "=" | "<>" | "!=") && ((lt == oid::XID && integer(rt)) || (integer(lt) && rt == oid::XID)) {
            let xid = |(expr, ty): Bound| {
                if ty.oid == oid::XID { expr } else { Expr::Cast(Box::new(expr), typ(oid::XID), true) }
            };
            let cmp = if op == "=" { CmpOp::Eq } else { CmpOp::Ne };
            return Ok((Expr::Compare(cmp, Box::new(xid(left)), Box::new(xid(right))), typ(oid::BOOL)));
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
                _ if matches!(lt, oid::BIT | oid::VARBIT) && matches!(rt, oid::BIT | oid::VARBIT) => typ(oid::VARBIT),
                _ if (lt == oid::CHAR && is_string(rt)) || (rt == oid::CHAR && is_string(lt)) => typ(oid::TEXT),
                _ if (is_oid_type(lt) || numeric_rank(lt).is_some_and(|r| r <= 2))
                    && (is_oid_type(rt) || numeric_rank(rt).is_some_and(|r| r <= 2)) =>
                {
                    typ(oid::OID)
                }
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
        let domain = if domain.oid == oid::BPCHAR { typ(oid::TEXT) } else { domain };
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
        let nested = nested || (is_array_type(ty.oid) && !crate::array::is_vector_type(ty.oid));
        let mut items = Vec::with_capacity(bound.len());
        for (item, location) in bound {
            items.push(coerce(item, ty, false, location)?.0);
        }
        if nested {
            return Ok((Expr::Array(element_type(ty.oid), items, true), typ(ty.oid)));
        }
        Ok((Expr::Array(ty.oid, items, false), typ(array_of(ty.oid))))
    }

    /// subscripts binds array subscripts as lower and upper bounds, reporting whether any selects a slice, which makes
    /// a lone subscript `n` the slice `1:n`.
    pub fn subscripts(&mut self, items: &[Node]) -> Result<(Subscripts, bool)> {
        let mut subscripts = Vec::new();
        let mut slice = false;
        for item in items {
            let Some(NodeEnum::AIndices(indices)) = item.node.as_ref() else {
                return Err(PgError::unsupported("field selection"));
            };
            slice |= indices.is_slice;
            let mut bound_index = |node: &Option<Box<Node>>| -> Result<Option<Expr>> {
                match node.as_deref() {
                    Some(node) => {
                        let bound = self.bind(node)?;
                        if let Expr::Param(i) = bound.0
                            && self.ctx.parameters[i] == 0
                        {
                            self.ctx.parameters[i] = oid::INT4;
                        }
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
        Ok((subscripts, slice))
    }

    /// indirection binds subscripts of an array.
    fn indirection(&mut self, indirection: &pg_query::protobuf::AIndirection) -> Result<Bound> {
        let arg = indirection.arg.as_deref().ok_or_else(|| PgError::internal("no subscripted value"))?;
        let (mut base, mut ty) = self.bind(arg)?;
        let mut items = indirection.indirection.as_slice();
        while let Some((first, rest)) = items.split_first() {
            let Some(NodeEnum::String(field)) = first.node.as_ref() else { break };
            let attributes = match crate::usertypes::get(ty.oid).map(|t| t.kind.clone()) {
                Some(crate::usertypes::Kind::Composite(attributes)) => attributes,
                _ => {
                    return Err(PgError {
                        position: position(arg_location(arg)),
                        ..PgError::new(
                            code::WRONG_OBJECT_TYPE,
                            format!(
                                "column notation .{} applied to type {}, which is not a composite type",
                                field.sval,
                                type_display(ty.oid)
                            ),
                        )
                    });
                }
            };
            let Some(index) = attributes.iter().position(|(name, _)| *name == field.sval) else {
                return Err(PgError {
                    position: position(arg_location(arg)),
                    ..PgError::new(
                        code::UNDEFINED_COLUMN,
                        format!("column \"{}\" not found in data type {}", field.sval, type_display(ty.oid)),
                    )
                });
            };
            base = Expr::Field(Box::new(base), index);
            ty = attributes[index].1;
            items = rest;
        }
        if items.is_empty() {
            return Ok((base, ty));
        }
        if ty.oid == oid::JSONB {
            return self.jsonb_subscripts((base, ty), items, arg_location(arg));
        }
        let (subscripts, slice) = self.subscripts(items)?;
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

    /// jsonb_subscripts binds subscripts of a jsonb value as the path of a `#>` lookup, each read as an integer or as
    /// text, as Postgres' jsonb_subscript_transform does.
    fn jsonb_subscripts(&mut self, base: Bound, items: &[Node], location: i32) -> Result<Bound> {
        let mut path = Vec::with_capacity(items.len());
        for item in items {
            let Some(NodeEnum::AIndices(indices)) = item.node.as_ref() else {
                return Err(PgError::unsupported("field selection"));
            };
            let index = indices.uidx.as_deref().ok_or_else(|| PgError::internal("a subscript without an index"))?;
            if indices.is_slice {
                return Err(PgError {
                    position: position(arg_location(index)),
                    ..PgError::new(code::DATATYPE_MISMATCH, "jsonb subscript does not support slices")
                });
            }
            let (expr, ty) = self.bind(index)?;
            let target = match ty.oid {
                oid::UNKNOWN => oid::TEXT,
                from => [oid::INT4, oid::TEXT]
                    .into_iter()
                    .find(|&to| from == to || implicitly_converts(from, to))
                    .ok_or_else(|| PgError {
                        position: position(arg_location(index)),
                        hint: Some("jsonb subscript must be coercible to either integer or text.".into()),
                        ..PgError::new(
                            code::DATATYPE_MISMATCH,
                            format!("subscript type {} is not supported", type_display(from)),
                        )
                    })?,
            };
            let bound = coerce((expr, ty), typ(target), false, arg_location(index))?;
            path.push(coerce(bound, typ(oid::TEXT), true, arg_location(index))?.0);
        }
        let path = (Expr::Array(oid::TEXT, path, false), typ(oid::TEXT_ARRAY));
        self.binary("#>", base, path, location)
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

/// Subscripts are an array's subscripts as lower and upper bounds.
pub type Subscripts = Vec<(Option<Expr>, Option<Expr>)>;

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

/// pattern_function returns the function behind a LIKE or regular expression operator.
fn pattern_function(op: &str) -> Option<&'static str> {
    Some(match op {
        "~~" => "textlike",
        "!~~" => "textnlike",
        "~~*" => "texticlike",
        "!~~*" => "texticnlike",
        "~" => "textregexeq",
        "!~" => "textregexne",
        "~*" => "texticregexeq",
        "!~*" => "texticregexne",
        _ => return None,
    })
}

/// implicit_datetime reports whether Postgres converts one date or time type to another without being asked.
pub(crate) fn implicit_datetime(from: u32, to: u32) -> bool {
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
            ("-", oid::DATE, oid::TIME) => {
                (D::TimestampMinusInterval(false), oid::TIMESTAMP, oid::INTERVAL, oid::TIMESTAMP)
            }
            ("+", oid::DATE, oid::TIMETZ) | ("+", oid::TIMETZ, oid::DATE) => {
                (D::DatePlusTimeTz, lt, rt, oid::TIMESTAMPTZ)
            }
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
            ("+", oid::TIMETZ, oid::INTERVAL) => (D::TimePlusInterval, oid::TIMETZ, oid::INTERVAL, oid::TIMETZ),
            ("+", oid::INTERVAL, oid::TIMETZ) => (D::TimePlusInterval, oid::INTERVAL, oid::TIMETZ, oid::TIMETZ),
            ("-", oid::TIMETZ, oid::INTERVAL) => (D::TimeMinusInterval, oid::TIMETZ, oid::INTERVAL, oid::TIMETZ),
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
        "~" if matches!(ty.oid, oid::INT2 | oid::INT4 | oid::INT8 | oid::BIT | oid::VARBIT) => {
            let resolved = functions::resolve("~", &[ty.oid], location)?;
            Ok((
                Expr::Func(resolved.index, vec![coerce((expr, ty), typ(resolved.arg_types[0]), false, location)?.0]),
                ty,
            ))
        }
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
        Some(NodeEnum::TypeCast(c)) => match c.arg.as_deref().map(|arg| (arg_location(arg), &arg.node)) {
            Some((_, Some(NodeEnum::RowExpr(_)))) => c.location,
            Some((arg, _)) if arg >= 0 && (arg < c.location || c.location < 0) => arg,
            _ => c.location,
        },
        Some(NodeEnum::ParamRef(p)) => p.location,
        Some(NodeEnum::FuncCall(f)) => f.location,
        Some(NodeEnum::RowExpr(r)) => r.location,
        Some(NodeEnum::SubLink(s)) => s.location,
        Some(NodeEnum::AArrayExpr(a)) => a.location,
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
        Some(Val::Bsval(b)) => {
            let bits =
                crate::binary::parse_bits(&b.bsval).map_err(|e| PgError { position: position(c.location), ..e })?;
            (Expr::Const(Value::Bit(bits)), typ(oid::BIT))
        }
        _ => return Err(PgError::unsupported("this constant")),
    })
}

/// coerce converts a bound expression to the type, folding untyped constants by reading them as the type, and
/// rejecting implicit conversions Postgres does not allow.
pub fn coerce((expr, from): Bound, to: ColumnType, explicit: bool, location: i32) -> Result<Bound> {
    if from == to {
        return Ok((expr, to));
    }
    let base = crate::usertypes::base_type(to);
    if base != to {
        let (expr, _) = coerce((expr, from), base, explicit, location)?;
        return Ok((Expr::Cast(Box::new(expr), to, explicit), to));
    }
    if let (Expr::Row(fields, types), oid::RECORD, Some(user_type)) = (&expr, from.oid, crate::usertypes::get(to.oid))
        && let crate::usertypes::Kind::Composite(attributes) = &user_type.kind
    {
        let error = |detail: String, at: i32| PgError {
            position: position(at),
            detail: Some(detail),
            ..PgError::new(code::CANNOT_COERCE, format!("cannot cast type record to {}", user_type.name))
        };
        if fields.len() != attributes.len() {
            let detail = if fields.len() < attributes.len() { "too few" } else { "too many" };
            return Err(error(format!("Input has {detail} columns."), location));
        }
        let mut converted = Vec::with_capacity(fields.len());
        for (i, ((field, &(ty, Location(at))), (_, attribute))) in fields.iter().zip(types).zip(attributes).enumerate()
        {
            let (field, ty) = match ty.oid {
                oid::UNKNOWN => (field.clone(), ty),
                _ if explicit => coerce((field.clone(), ty), *attribute, true, location)?,
                _ => assign((field.clone(), ty), *attribute, "", at).map_err(|err| match err.code {
                    code::DATATYPE_MISMATCH => error(
                        format!(
                            "Cannot cast type {} to {} in column {}.",
                            type_display(ty.oid),
                            type_display(attribute.oid),
                            i + 1
                        ),
                        at,
                    ),
                    _ => err,
                })?,
            };
            converted.push(coerce((field, ty), *attribute, explicit, at)?.0);
        }
        let types = attributes.iter().map(|(_, t)| (*t, Location(-1))).collect();
        return Ok((Expr::Cast(Box::new(Expr::Row(converted, types)), to, explicit), to));
    }
    let from = crate::usertypes::base_type(from);
    if (from.oid == oid::CHAR) != (to.oid == oid::CHAR) && from.oid != oid::UNKNOWN {
        return char_cast((expr, from), to, explicit, location);
    }
    if from.oid == oid::UNKNOWN {
        if let Expr::Const(value) = &expr
            && !crate::cast::is_reg_type(to.oid)
        {
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
    let opaque =
        |t: u32| matches!(t, oid::BYTEA | oid::UUID | oid::BIT | oid::VARBIT) || crate::basetypes::get(t).is_some();
    let bits_or_ints = |t: u32| matches!(t, oid::BIT | oid::VARBIT | oid::INT4 | oid::INT8);
    let textual = is_string(from.oid) || is_string(to.oid);
    let context = if explicit { crate::casts::EXPLICIT } else { crate::casts::IMPLICIT };
    if let Some(bound) = user_cast(&expr, from, to, explicit, context) {
        return Ok(bound);
    }
    let xml_only_textual = (from.oid == oid::XML) != (to.oid == oid::XML) && !textual;
    let transaction_id = (matches!(to.oid, oid::XID | oid::CID) || matches!(from.oid, oid::XID | oid::CID))
        && from.oid != to.oid
        && !textual;
    let composite =
        |t: u32| crate::usertypes::get(t).is_some_and(|u| matches!(u.kind, crate::usertypes::Kind::Composite(_)));
    let other = if from.oid == oid::BOOL { to.oid } else { from.oid };
    let boolean = (from.oid == oid::BOOL) != (to.oid == oid::BOOL) && !textual && other != oid::INT4;
    let allowed = (explicit
        && (is_array_type(from.oid) == is_array_type(to.oid) || textual)
        && !(composite(from.oid) && composite(to.oid))
        && (!(opaque(from.oid) || opaque(to.oid)) || textual || (bits_or_ints(from.oid) && bits_or_ints(to.oid)))
        && !xml_only_textual
        && !transaction_id
        && !oid_without_cast(from.oid, to.oid)
        && !boolean
        || implicitly_converts(from.oid, to.oid))
        && !(crate::array::is_vector_type(to.oid) && is_array_type(from.oid));
    if !allowed {
        return Err(PgError {
            position: position(location),
            ..PgError::new(
                if explicit { code::CANNOT_COERCE } else { code::DATATYPE_MISMATCH },
                format!("cannot cast type {} to {}", type_display(from.oid), type_display(to.oid)),
            )
        });
    }
    if from.oid == oid::BPCHAR && matches!(to.oid, oid::TEXT | oid::VARCHAR | oid::NAME) {
        let trimmed = match expr {
            Expr::Const(Value::Text(text)) => Expr::Const(Value::Text(text.trim_end_matches(' ').to_string())),
            other => Expr::Func(functions::resolve("rtrim", &[oid::TEXT], location)?.index, vec![other]),
        };
        return coerce((trimmed, typ(oid::TEXT)), to, explicit, location);
    }
    if let Expr::Const(value) = &expr
        && !crate::cast::is_reg_type(to.oid)
    {
        return Ok((Expr::Const(cast_value(value.clone(), to, explicit)?), to));
    }
    Ok((Expr::Cast(Box::new(expr), to, explicit), to))
}

/// oid_without_cast reports whether Postgres has no cast between oid and a numeric type: from oid to smallint or a
/// non-integer type, or from a non-integer type to oid.
fn oid_without_cast(from: u32, to: u32) -> bool {
    let fractional = |t: u32| matches!(t, oid::FLOAT4 | oid::FLOAT8 | oid::NUMERIC);
    (from == oid::OID && (to == oid::INT2 || fractional(to))) || (to == oid::OID && fractional(from))
}

/// user_cast calls the routine of a stored cast from one type to the other that a context allows, passing the
/// target's modifier and whether the cast is explicit when the routine takes them.
fn user_cast(expr: &Expr, from: ColumnType, to: ColumnType, explicit: bool, context: u8) -> Option<Bound> {
    let (allowed, routine) = crate::casts::find(from.oid, to.oid)?;
    if allowed < context {
        return None;
    }
    let Some(routine) = routine else {
        let text = Expr::Cast(Box::new(expr.clone()), typ(oid::TEXT), true);
        return Some((Expr::Cast(Box::new(text), to, true), to));
    };
    let mut args = vec![expr.clone()];
    if routine.params.len() > 1 {
        args.push(Expr::Const(Value::Int4(to.modifier)));
    }
    if routine.params.len() > 2 {
        args.push(Expr::Const(Value::Bool(explicit)));
    }
    Some((Expr::Routine(routine, args), to))
}

/// char_cast converts to or from the "char" type, which converts to and from integers by its byte, to the string types,
/// and only explicitly from name.
fn char_cast((expr, from): Bound, to: ColumnType, explicit: bool, location: i32) -> Result<Bound> {
    let integer = if from.oid == oid::CHAR { to.oid } else { from.oid };
    let allowed = if integer == oid::INT4 {
        explicit
    } else if from.oid == oid::CHAR {
        is_string(to.oid) && (explicit || to.oid == oid::TEXT || to.oid != oid::NAME)
    } else {
        is_string(from.oid) && (explicit || from.oid != oid::NAME)
    };
    if !allowed {
        return Err(PgError {
            position: position(location),
            ..PgError::new(
                code::CANNOT_COERCE,
                format!("cannot cast type {} to {}", type_display(from.oid), type_display(to.oid)),
            )
        });
    }
    if integer == oid::INT4 {
        let name = if to.oid == oid::CHAR { "char" } else { "int4" };
        let resolved = functions::resolve(name, &[from.oid], location)?;
        return Ok((Expr::Func(resolved.index, vec![expr]), to));
    }
    Ok((Expr::Cast(Box::new(expr), to, explicit), to))
}

/// implicitly_converts reports whether Postgres converts a value of one type to the other without being asked.
pub(crate) fn implicitly_converts(from: u32, to: u32) -> bool {
    if is_array_type(from) && is_array_type(to) {
        return implicitly_converts(element_type(from), element_type(to));
    }
    let numeric = numeric_rank(from).zip(numeric_rank(to));
    from == to
        || numeric.is_some_and(|(f, t)| f <= t)
        || (is_string(from) && is_string(to))
        || implicit_datetime(from, to)
        || (from == oid::CHAR && to == oid::TEXT)
        || (matches!(from, oid::INT2 | oid::INT4 | oid::INT8) && is_oid_type(to))
        || (is_oid_type(from) && is_oid_type(to) && (from == oid::OID || to == oid::OID))
        || (from == oid::RECORD && is_composite(to))
        || (matches!(from, oid::BIT | oid::VARBIT) && matches!(to, oid::BIT | oid::VARBIT))
        || crate::casts::context(from, to) == Some(crate::casts::IMPLICIT)
}

/// is_composite reports whether a type is a user-defined composite type.
pub fn is_composite(type_oid: u32) -> bool {
    crate::usertypes::get(type_oid).is_some_and(|t| matches!(t.kind, crate::usertypes::Kind::Composite(_)))
}

/// is_oid_type reports whether a type holds an OID: oid or one of the reg types.
pub fn is_oid_type(type_oid: u32) -> bool {
    type_oid == oid::OID || crate::cast::is_reg_type(type_oid)
}

/// singular returns the singular of a plural clause name, as Postgres' errors name a definition.
fn singular(clause: &str) -> &str {
    clause.strip_suffix('s').unwrap_or(clause)
}

/// assignable reports whether a value of one type converts to the other on assignment.
pub(crate) fn assignable(from: u32, to: u32) -> bool {
    let (from, to) = (crate::usertypes::base_type(typ(from)).oid, crate::usertypes::base_type(typ(to)).oid);
    if is_array_type(from) && is_array_type(to) {
        return assignable(element_type(from), element_type(to));
    }
    from == oid::UNKNOWN
        || from == to
        || (numeric_rank(from).is_some() && numeric_rank(to).is_some())
        || assignable_datetime(from, to)
        || (is_string(to) && !is_array_type(from))
        || implicitly_converts(from, to)
        || (is_oid_type(from) && matches!(to, oid::INT4 | oid::INT8))
        || (to == oid::CHAR && matches!(from, oid::TEXT | oid::VARCHAR | oid::BPCHAR))
        || (from == oid::CHAR && is_string(to))
        || crate::casts::context(from, to).is_some_and(|c| c >= crate::casts::ASSIGNMENT)
}

/// element_type returns the element type of an array type.
pub fn element_type(array_type: u32) -> u32 {
    match builtin_type(array_type) {
        Some(t) => t.elem,
        None => match crate::usertypes::get(array_type).map(|t| t.kind.clone()) {
            Some(crate::usertypes::Kind::Array(element)) => element,
            _ => oid::TEXT,
        },
    }
}

/// array_of returns the array type of an element type.
pub fn array_of(element: u32) -> u32 {
    match builtin_type(element) {
        Some(t) => t.array,
        None => crate::usertypes::get(element).map_or(oid::TEXT_ARRAY, |t| t.array),
    }
}

/// assign converts a bound expression to a column's type as an assignment does, which also allows numeric narrowing
/// and conversions to text, and names the column in its error.
pub fn assign(bound: Bound, to: ColumnType, column: &str, location: i32) -> Result<Bound> {
    if let (Expr::Row(..), true) = (&bound.0, is_composite(to.oid)) {
        return coerce(bound, to, false, location);
    }
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
    if let Some(bound) = user_cast(&bound.0, bound.1, to, false, crate::casts::ASSIGNMENT) {
        return Ok(bound);
    }
    if let Expr::Const(value) = &bound.0
        && !crate::cast::is_reg_type(to.oid)
    {
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

/// condition converts a bound expression to boolean as Postgres' coerce_to_boolean does, failing for another type with
/// the name of the clause whose argument it is.
pub fn condition(bound: Bound, clause: &str, location: i32) -> Result<Expr> {
    if !matches!(crate::usertypes::base_type(bound.1).oid, oid::BOOL | oid::UNKNOWN) {
        let shown = crate::cast::type_display(bound.1.oid);
        return Err(PgError {
            position: position(location),
            ..PgError::new(
                code::DATATYPE_MISMATCH,
                format!("argument of {clause} must be type boolean, not type {shown}"),
            )
        });
    }
    Ok(coerce(bound, typ(oid::BOOL), false, location)?.0)
}

/// figure_index_name returns the name Postgres gives an index column for an expression, which is `expr` when nothing
/// names it.
pub fn figure_index_name(node: &Node) -> String {
    match figure_name_strength(node) {
        (_, 0) => "expr".into(),
        (name, _) => name,
    }
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
        Some(NodeEnum::XmlExpr(x)) => {
            use pg_query::protobuf::XmlExprOp as X;
            match X::try_from(x.op) {
                Ok(X::IsXmlconcat) => strong("xmlconcat"),
                Ok(X::IsXmlelement) => strong("xmlelement"),
                Ok(X::IsXmlforest) => strong("xmlforest"),
                Ok(X::IsXmlparse) => strong("xmlparse"),
                Ok(X::IsXmlpi) => strong("xmlpi"),
                Ok(X::IsXmlroot) => strong("xmlroot"),
                Ok(X::IsXmlserialize) => strong("xmlserialize"),
                _ => ("?column?".into(), 0),
            }
        }
        Some(NodeEnum::XmlSerialize(_)) => strong("xmlserialize"),
        Some(NodeEnum::CollateClause(collate)) => match collate.arg.as_deref() {
            Some(arg) => figure_name_strength(arg),
            None => ("?column?".into(), 0),
        },
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
        Value::Oid(o) => Some(*o as i64),
        Value::Reg(reg) => Some(reg.oid as i64),
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
        (Value::Jsonb(l), Value::Jsonb(r)) => crate::json::compare(l, r),
        (Value::Enum(l), Value::Enum(r)) => {
            let position = |label: &str| match crate::usertypes::get(l.type_oid).map(|t| t.kind.clone()) {
                Some(crate::usertypes::Kind::Enum(labels)) => labels.iter().position(|x| x == label),
                _ => None,
            };
            position(&l.label).cmp(&position(&r.label))
        }
        (Value::Composite(l), Value::Composite(r)) => {
            compare_values(&Value::Record(l.fields.clone()), &Value::Record(r.fields.clone()))
        }
        (Value::Bytea(l), Value::Bytea(r)) => l.cmp(r),
        (Value::Uuid(l), Value::Uuid(r)) => l.cmp(r),
        (Value::Bit(l), Value::Bit(r)) => l.cmp(r),
        (Value::Base(l), Value::Base(r)) => match crate::types::base_type(l.type_oid) {
            Some(definition) => (definition.compare)(&l.data, &r.data),
            None => Ordering::Equal,
        },
        (Value::Record(l), Value::Record(r)) => {
            for (a, b) in l.iter().zip(r) {
                let ordering = match (a.is_null(), b.is_null()) {
                    (true, true) => Ordering::Equal,
                    (true, false) => Ordering::Greater,
                    (false, true) => Ordering::Less,
                    (false, false) => compare_values(a, b),
                };
                if ordering != Ordering::Equal {
                    return ordering;
                }
            }
            l.len().cmp(&r.len())
        }
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
                if crate::cast::is_reg_type(ty.oid) {
                    return ctx.reg_value(value, ty.oid);
                }
                let value = match value {
                    Value::Text(text) if !matches!(ty.oid, oid::TEXT | oid::VARCHAR | oid::BPCHAR | oid::NAME) => {
                        cast_value(crate::cast::input(&text, ty.oid)?, *ty, *explicit)?
                    }
                    value => cast_value(value, *ty, *explicit)?,
                };
                ctx.check_domain(&value, *ty)?;
                value
            }
            Expr::Field(expr, index) => match expr.eval(ctx, row)? {
                Value::Composite(c) => c.fields.get(*index).cloned().unwrap_or(Value::Null),
                Value::Record(fields) => fields.get(*index).cloned().unwrap_or(Value::Null),
                value if *index == 0 => value,
                _ => Value::Null,
            },
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
            Expr::IsNull(expr, negated) => match expr.eval(ctx, row)? {
                Value::Record(fields) => Value::Bool(fields.iter().all(|f| f.is_null() != *negated)),
                Value::Composite(c) => Value::Bool(c.fields.iter().all(|f| f.is_null() != *negated)),
                value => Value::Bool(value.is_null() != *negated),
            },
            Expr::Func(index, args) => {
                let values = args.iter().map(|a| a.eval(ctx, row)).collect::<Result<Vec<_>>>()?;
                functions::call(ctx, *index, &values)?
            }
            Expr::Routine(routine, args) => {
                let values = args.iter().map(|a| a.eval(ctx, row)).collect::<Result<Vec<_>>>()?;
                crate::routines::call(ctx, routine, values)?
            }
            Expr::Operator(_, routine, l, r) => {
                let values = vec![l.eval(ctx, row)?, r.eval(ctx, row)?];
                crate::routines::call(ctx, routine, values)?
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
            Expr::WindowRef(_) => return Err(PgError::internal("a window call outside its window")),
            Expr::SetRef(_) => return Err(PgError::internal("a set-returning call outside its select list")),
            Expr::Default(_) => return Err(PgError::internal("a default outside a written row")),
            Expr::DateTime(op, left, right) => {
                let (l, r) = (left.eval(ctx, row)?, right.eval(ctx, row)?);
                if l.is_null() || r.is_null() {
                    return Ok(Value::Null);
                }
                date_op(*op, l, r)?
            }
            Expr::SubqueryValue => ctx.subquery_value.clone(),
            Expr::Row(fields, _) => Value::Record(fields.iter().map(|f| f.eval(ctx, row)).collect::<Result<Vec<_>>>()?),
            Expr::Array(element, items, nested) => {
                let values = items.iter().map(|i| i.eval(ctx, row)).collect::<Result<Vec<_>>>()?;
                if *nested {
                    Value::Array(Box::new(crate::array::nest(*element, values)?))
                } else {
                    Value::Array(Box::new(Array::one_dimensional(*element, values)))
                }
            }
            Expr::SubscriptAssign(base, element, subscripts, slice, value) => {
                let array = match base.eval(ctx, row)? {
                    Value::Array(array) => *array,
                    _ => crate::array::Array { element: *element, dims: Vec::new(), values: Vec::new() },
                };
                let mut bounds = Vec::with_capacity(subscripts.len());
                for (lower, upper) in subscripts {
                    let mut bound = |e: &Option<Expr>| -> Result<Option<i32>> {
                        let Some(e) = e else { return Ok(None) };
                        match e.eval(ctx, row)? {
                            Value::Int4(i) => Ok(Some(i)),
                            _ => Err(PgError::new(
                                code::NULL_VALUE_NOT_ALLOWED,
                                "array subscript in assignment must not be null",
                            )),
                        }
                    };
                    bounds.push((bound(lower)?, bound(upper)?));
                }
                let assigned = if *slice {
                    let source = match value.eval(ctx, row)? {
                        Value::Array(source) => Some(*source),
                        _ => None,
                    };
                    crate::array::assign_slice(array, &bounds, source)?
                } else {
                    let indexes: Vec<i32> = bounds.iter().map(|(_, upper)| upper.unwrap_or_default()).collect();
                    crate::array::assign_element(array, &indexes, value.eval(ctx, row)?)?
                };
                Value::Array(Box::new(assigned))
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
            Expr::Shared(value, body) => {
                let value = value.eval(ctx, row)?;
                let previous = std::mem::replace(&mut ctx.subquery_value, value);
                let result = body.eval(ctx, row);
                ctx.subquery_value = previous;
                result?
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
            Expr::Xml(op, args) => {
                let values = args.iter().map(|a| a.eval(ctx, row)).collect::<Result<Vec<_>>>()?;
                crate::xml::sql::eval(op, values)?
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
                match rows.into_iter().next() {
                    Some(r) if r.len() > 1 => Value::Record(r),
                    Some(r) => r.into_iter().next().unwrap_or(Value::Null),
                    None => Value::Null,
                }
            }
            Expr::ArraySubquery(plan, element) => {
                ctx.outer.push(row.to_vec());
                let rows = plan.run(ctx);
                ctx.outer.pop();
                let values: Vec<Value> =
                    rows?.into_iter().map(|r| r.into_iter().next().unwrap_or(Value::Null)).collect();
                if is_array_type(*element) && !values.is_empty() {
                    return crate::functions::aggregate::array_agg_arrays(element_type(*element), values);
                }
                let element = if is_array_type(*element) { element_type(*element) } else { *element };
                Value::Array(Box::new(crate::array::Array::one_dimensional(element, values)))
            }
            Expr::AnySubquery(comparison, plan, all) => {
                ctx.outer.push(row.to_vec());
                let rows = plan.run(ctx);
                ctx.outer.pop();
                let mut saw_null = false;
                let previous = std::mem::replace(&mut ctx.subquery_value, Value::Null);
                let mut result = None;
                for r in rows? {
                    ctx.subquery_value =
                        if r.len() == 1 { r.into_iter().next().unwrap_or(Value::Null) } else { Value::Record(r) };
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
            Expr::Field(e, i) => Expr::Field(b(e), i),
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
            Expr::Routine(r, args) => Expr::Routine(r, args.into_iter().map(&mut *f).collect()),
            Expr::Operator(name, routine, l, r) => {
                let l = b(l);
                Expr::Operator(name, routine, l, b(r))
            }
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
            Expr::Row(fields, types) => Expr::Row(fields.into_iter().map(&mut *f).collect(), types),
            Expr::Subscript(base, subscripts, slice) => {
                let base = b(base);
                let subscripts = subscripts.into_iter().map(|(l, u)| (l.map(&mut *f), u.map(&mut *f))).collect();
                Expr::Subscript(base, subscripts, slice)
            }
            Expr::SubscriptAssign(base, element, subscripts, slice, value) => {
                let (base, value) = (b(base), b(value));
                let subscripts = subscripts.into_iter().map(|(l, u)| (l.map(&mut *f), u.map(&mut *f))).collect();
                Expr::SubscriptAssign(base, element, subscripts, slice, value)
            }
            Expr::AnyArray(c, a, all) => {
                let c = b(c);
                Expr::AnyArray(c, b(a), all)
            }
            Expr::Shared(value, body) => {
                let value = b(value);
                Expr::Shared(value, b(body))
            }
            Expr::ArrayOp(op, l, r) => {
                let l = b(l);
                Expr::ArrayOp(op, l, b(r))
            }
            Expr::Xml(op, args) => Expr::Xml(op, args.into_iter().map(&mut *f).collect()),
            other => other,
        }
    }

    /// visit calls the function on the expression and each of its descendants, outside subquery plans.
    pub fn visit(&self, f: &mut dyn FnMut(&Expr)) {
        f(self);
        match self {
            Expr::Cast(e, ..)
            | Expr::Neg(e, _)
            | Expr::Not(e)
            | Expr::IsNull(e, _)
            | Expr::BoolTest(e, ..)
            | Expr::Field(e, _) => e.visit(f),
            Expr::Arith(_, l, r, _)
            | Expr::Compare(_, l, r)
            | Expr::Concat(l, r)
            | Expr::And(l, r)
            | Expr::Or(l, r)
            | Expr::NullIf(l, r)
            | Expr::DateTime(_, l, r)
            | Expr::AnyArray(l, r, _)
            | Expr::Shared(l, r)
            | Expr::Operator(_, _, l, r)
            | Expr::ArrayOp(_, l, r)
            | Expr::DistinctFrom(l, r, _) => {
                l.visit(f);
                r.visit(f);
            }
            Expr::Func(_, args)
            | Expr::Routine(_, args)
            | Expr::Coalesce(args)
            | Expr::MinMax(_, args)
            | Expr::Array(_, args, _)
            | Expr::Xml(_, args)
            | Expr::Row(args, _) => args.iter().for_each(|a| a.visit(f)),
            Expr::Subscript(base, subscripts, _) => {
                base.visit(f);
                for (l, u) in subscripts {
                    l.iter().chain(u).for_each(|e| e.visit(f));
                }
            }
            Expr::SubscriptAssign(base, _, subscripts, _, value) => {
                base.visit(f);
                for (l, u) in subscripts {
                    l.iter().chain(u).for_each(|e| e.visit(f));
                }
                value.visit(f);
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
        (DateOp::DatePlusTimeTz, Value::Date(d), Value::TimeTz(t, west))
        | (DateOp::DatePlusTimeTz, Value::TimeTz(t, west), Value::Date(d)) => {
            Value::TimestampTz(d as i64 * USECS_PER_DAY + t + west as i64 * dt::USECS_PER_SEC)
        }
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
        (DateOp::TimePlusInterval, Value::TimeTz(t, zone), Value::Interval(iv))
        | (DateOp::TimePlusInterval, Value::Interval(iv), Value::TimeTz(t, zone)) => {
            Value::TimeTz((t + iv.micros).rem_euclid(USECS_PER_DAY), zone)
        }
        (DateOp::TimeMinusInterval, Value::TimeTz(t, zone), Value::Interval(iv)) => {
            Value::TimeTz((t - iv.micros).rem_euclid(USECS_PER_DAY), zone)
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
