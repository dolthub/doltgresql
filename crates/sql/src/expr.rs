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
use pg_query::protobuf::{AExprKind, BoolExprType, NullTestType};
use pg_query::{Node, NodeEnum};

use crate::cast::{cast_value, type_display};
use crate::catalog::{ColumnType, resolve_type};
use crate::error::{PgError, Result, code};
use crate::numeric::Numeric;
use crate::oid;
use crate::types::Value;

/// ScopeColumn is a column that expressions can refer to, from the table it belongs to.
#[derive(Clone, Debug)]
pub struct ScopeColumn {
    /// The name or alias of the table, empty for a column outside any table.
    pub table: String,
    pub name: String,
    pub ty: ColumnType,
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
}

/// Bound is a bound expression with its type.
pub type Bound = (Expr, ColumnType);

/// Binder binds expressions over a scope, collecting the types of the parameters it sees.
pub struct Binder<'a> {
    pub scope: &'a Scope,
    /// The types of the statement's parameters, where 0 is a type not yet known.
    pub parameters: &'a mut Vec<u32>,
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

impl Binder<'_> {
    /// bind binds an expression.
    pub fn bind(&mut self, node: &Node) -> Result<Bound> {
        let Some(node) = node.node.as_ref() else { return Err(PgError::internal("an empty expression")) };
        match node {
            NodeEnum::AConst(c) => constant(c),
            NodeEnum::ColumnRef(column) => self.column(column),
            NodeEnum::ParamRef(param) => {
                let index = param.number as usize - 1;
                if self.parameters.len() <= index {
                    self.parameters.resize(index + 1, 0);
                }
                Ok((
                    Expr::Param(index),
                    typ(match self.parameters[index] {
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
        let mut found =
            self.scope.columns.iter().enumerate().filter(|(_, c)| c.name == name && table.is_none_or(|t| c.table == t));
        let full = names.join(".");
        match (found.next(), found.next()) {
            (Some((i, c)), None) => Ok((Expr::Column(i), c.ty)),
            (Some(_), Some(_)) => Err(PgError {
                position: position(column.location),
                ..PgError::new(code::AMBIGUOUS_COLUMN, format!("column reference \"{full}\" is ambiguous"))
            }),
            _ => {
                if let Some(table) = table
                    && !self.scope.columns.iter().any(|c| c.table == table)
                {
                    return Err(PgError {
                        position: position(column.location),
                        ..PgError::new(
                            code::UNDEFINED_TABLE,
                            format!("missing FROM-clause entry for table \"{table}\""),
                        )
                    });
                }
                Err(PgError {
                    position: position(column.location),
                    ..PgError::new(code::UNDEFINED_COLUMN, format!("column \"{full}\" does not exist"))
                })
            }
        }
    }
}

/// operand returns an operand of an operator expression.
fn operand(side: &Option<Box<Node>>) -> Result<&Node> {
    side.as_deref().ok_or_else(|| PgError::internal("no operand"))
}

impl Binder<'_> {
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
            _ => Err(PgError::unsupported(format!("the operator expression {kind:?}"))),
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
                && binder.parameters[i] == 0
            {
                binder.parameters[i] = ty.oid;
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
    /// eval evaluates the expression over a row with the statement's parameters.
    pub fn eval(&self, row: &[Value], params: &[Value]) -> Result<Value> {
        Ok(match self {
            Expr::Const(value) => value.clone(),
            Expr::Column(i) => row[*i].clone(),
            Expr::Param(i) => params.get(*i).cloned().unwrap_or(Value::Null),
            Expr::Cast(expr, ty, explicit) => {
                let value = expr.eval(row, params)?;
                match value {
                    Value::Text(text) if !matches!(ty.oid, oid::TEXT | oid::VARCHAR | oid::BPCHAR | oid::NAME) => {
                        cast_value(crate::cast::input(&text, ty.oid)?, *ty, *explicit)?
                    }
                    value => cast_value(value, *ty, *explicit)?,
                }
            }
            Expr::Arith(op, left, right, ty) => {
                let (left, right) = (left.eval(row, params)?, right.eval(row, params)?);
                if left.is_null() || right.is_null() {
                    return Ok(Value::Null);
                }
                arith(*op, &left, &right, *ty)?
            }
            Expr::Neg(expr, ty) => match expr.eval(row, params)? {
                Value::Null => Value::Null,
                Value::Numeric(n) => Value::Numeric(n.negate()),
                Value::Float4(f) => Value::Float4(-f),
                Value::Float8(f) => Value::Float8(-f),
                value => int_result(as_i64(&value).and_then(i64::checked_neg), *ty)?,
            },
            Expr::Compare(op, left, right) => {
                let (left, right) = (left.eval(row, params)?, right.eval(row, params)?);
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
            Expr::Concat(left, right) => match (left.eval(row, params)?, right.eval(row, params)?) {
                (Value::Null, _) | (_, Value::Null) => Value::Null,
                (l, r) => Value::Text(format!("{}{}", l.output().unwrap_or_default(), r.output().unwrap_or_default())),
            },
            Expr::And(left, right) => match left.eval(row, params)? {
                Value::Bool(false) => Value::Bool(false),
                l => match (l, right.eval(row, params)?) {
                    (_, Value::Bool(false)) => Value::Bool(false),
                    (Value::Bool(true), Value::Bool(true)) => Value::Bool(true),
                    _ => Value::Null,
                },
            },
            Expr::Or(left, right) => match left.eval(row, params)? {
                Value::Bool(true) => Value::Bool(true),
                l => match (l, right.eval(row, params)?) {
                    (_, Value::Bool(true)) => Value::Bool(true),
                    (Value::Bool(false), Value::Bool(false)) => Value::Bool(false),
                    _ => Value::Null,
                },
            },
            Expr::Not(expr) => match expr.eval(row, params)? {
                Value::Bool(b) => Value::Bool(!b),
                _ => Value::Null,
            },
            Expr::IsNull(expr, negated) => Value::Bool(expr.eval(row, params)?.is_null() != *negated),
        })
    }

    /// is_true evaluates a condition, treating NULL as false.
    pub fn is_true(&self, row: &[Value], params: &[Value]) -> Result<bool> {
        Ok(matches!(self.eval(row, params)?, Value::Bool(true)))
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
