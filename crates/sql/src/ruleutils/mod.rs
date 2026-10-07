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

//! Postgres' ruleutils: shaping a parsed expression or query with the casts that Postgres' parse analysis adds for the
//! common built-in types, and printing it as pg_get_expr, pg_get_constraintdef, pg_get_viewdef, and EXPLAIN print
//! theirs, with or without pretty-printing.

mod query;

use pg_query::protobuf::a_const::Val;
use pg_query::protobuf::{AExprKind, BoolExprType, BoolTestType, MinMaxOp, NullTestType, SqlValueFunctionOp};
use pg_query::{Node, NodeEnum};

use crate::catalog::ColumnType;
use crate::error::Result;
use crate::expr::node_name;
use crate::oid;
use crate::query::Ctx;

/// TExpr is an analyzed expression, shaped as Postgres' parse analysis shapes it.
#[derive(Clone, Debug)]
enum TExpr {
    /// A column, with the relation name that qualifies it in a trigger's condition, or a whole row of a trigger's
    /// `old` or `new` as the column `*` of type record.
    Var(Option<String>, String, ColumnType),
    /// A constant's text in its type's output format, or None for NULL.
    Const(Option<String>, ColumnType),
    Param(usize),
    /// An operator call with one argument for a prefix operator, and the result type.
    Op(String, Vec<TExpr>, u32),
    /// AND, or OR when the flag is unset, over the arguments.
    Bool(bool, Vec<TExpr>),
    Not(Box<TExpr>),
    /// IS NULL, or IS NOT NULL when the flag is set.
    NullTest(Box<TExpr>, bool),
    /// A boolean test with its keywords, such as IS TRUE.
    BoolTest(Box<TExpr>, &'static str),
    /// IS DISTINCT FROM, or IS NOT DISTINCT FROM when the flag is set.
    Distinct(Box<TExpr>, Box<TExpr>, bool),
    Func(String, Vec<TExpr>, u32),
    /// A function that the SQL standard spells as a keyword, such as CURRENT_DATE.
    Keyword(&'static str, u32),
    /// A cast, which is implicit when parse analysis added it.
    Cast(Box<TExpr>, ColumnType, bool),
    /// A comparison with each element of an array, holding for any element, or for all when the flag is unset.
    ArrayOp(String, bool, Box<TExpr>, Box<TExpr>),
    Array(Vec<TExpr>, u32),
    /// CASE with its test value, the conditions or test values and results of its WHEN clauses, and default.
    Case(Option<Box<TExpr>>, Vec<(TExpr, TExpr)>, Box<TExpr>, u32),
    Coalesce(Vec<TExpr>, u32),
    /// GREATEST, or LEAST when the flag is unset.
    MinMax(bool, Vec<TExpr>, u32),
    NullIf(Box<TExpr>, Box<TExpr>, u32),
    Row(Vec<TExpr>),
    /// A subquery of a kind, with the value and operator that ANY and ALL compare its rows with, and its type.
    Sub(SubKind, Option<Box<TExpr>>, String, Box<query::TQuery>, u32),
    /// An aggregate or window function call with its modifiers.
    Agg(Box<query::Agg>),
    /// A function that the SQL standard spells with keywords, such as EXTRACT, by its function name, and its type.
    Syntax(String, Vec<TExpr>, u32),
    /// An expression this module does not shape, as pg_query deparses it.
    Raw(String),
}

/// SubKind is the kind of a subquery in an expression.
#[derive(Clone, Copy, Debug, PartialEq)]
enum SubKind {
    Exists,
    Any,
    All,
    /// The one value of a subquery's one row.
    Value,
    Array,
}

/// typ returns a column type without a modifier.
fn typ(type_oid: u32) -> ColumnType {
    ColumnType { oid: type_oid, modifier: -1 }
}

impl TExpr {
    /// ty returns the expression's type.
    fn ty(&self) -> ColumnType {
        match self {
            TExpr::Var(_, _, t) | TExpr::Const(_, t) | TExpr::Cast(_, t, _) => *t,
            TExpr::Op(_, _, t)
            | TExpr::Func(_, _, t)
            | TExpr::Keyword(_, t)
            | TExpr::Array(_, t)
            | TExpr::Case(.., t)
            | TExpr::Coalesce(_, t)
            | TExpr::MinMax(_, _, t)
            | TExpr::NullIf(_, _, t)
            | TExpr::Sub(.., t)
            | TExpr::Syntax(_, _, t) => typ(*t),
            TExpr::Agg(agg) => typ(agg.ret),
            TExpr::Bool(..)
            | TExpr::Not(_)
            | TExpr::NullTest(..)
            | TExpr::BoolTest(..)
            | TExpr::Distinct(..)
            | TExpr::ArrayOp(..) => typ(oid::BOOL),
            TExpr::Row(_) => typ(oid::RECORD),
            TExpr::Param(_) | TExpr::Raw(_) => typ(oid::UNKNOWN),
        }
    }
}

/// Integer, numeric, floating-point, and string type classes that operator resolution distinguishes.
fn is_integer(t: u32) -> bool {
    matches!(t, oid::INT2 | oid::INT4 | oid::INT8)
}

/// is_float reports whether a type is a floating-point type.
fn is_float(t: u32) -> bool {
    matches!(t, oid::FLOAT4 | oid::FLOAT8)
}

/// is_stringish reports whether a type is a string type whose values compare as text.
fn is_stringish(t: u32) -> bool {
    matches!(t, oid::TEXT | oid::VARCHAR | oid::BPCHAR | oid::NAME)
}

/// is_datetime reports whether a type is a date or timestamp type, which compare with each other directly.
fn is_datetime(t: u32) -> bool {
    matches!(t, oid::DATE | oid::TIMESTAMP | oid::TIMESTAMPTZ)
}

/// is_comparison reports whether an operator compares its operands and returns a boolean.
fn is_comparison(op: &str) -> bool {
    matches!(op, "=" | "<>" | "<" | "<=" | ">" | ">=" | "~~" | "!~~" | "~~*" | "!~~*" | "~" | "!~" | "~*" | "!~*")
}

/// Analyzer shapes parsed expressions over a table's columns.
pub struct Analyzer<'c, 'a> {
    ctx: &'c mut Ctx<'a>,
    columns: Vec<(String, ColumnType)>,
    /// The relation names that may qualify columns, which a trigger's condition gives as `old` and `new`.
    qualifiers: &'static [&'static str],
    /// The FROM items that a query's column references resolve against, for each enclosing query, innermost last.
    scopes: Vec<Vec<query::Rte>>,
    /// The common table expressions in scope with their columns, innermost last.
    ctes: Vec<(String, Vec<(String, ColumnType)>)>,
}

impl<'c, 'a> Analyzer<'c, 'a> {
    /// new returns an analyzer over columns.
    pub fn new(ctx: &'c mut Ctx<'a>, columns: Vec<(String, ColumnType)>) -> Analyzer<'c, 'a> {
        Analyzer { ctx, columns, qualifiers: &[], scopes: Vec::new(), ctes: Vec::new() }
    }

    /// trigger returns an analyzer over the columns of a trigger's table, which its condition names through `old`
    /// and `new`.
    pub fn trigger(ctx: &'c mut Ctx<'a>, columns: Vec<(String, ColumnType)>) -> Analyzer<'c, 'a> {
        Analyzer { ctx, columns, qualifiers: &["old", "new"], scopes: Vec::new(), ctes: Vec::new() }
    }

    /// deparse prints an expression's text as Postgres' deparse_expression does, at the top level where implicit casts
    /// do not show, after converting it to a column's type when given one, as a default is.
    pub fn deparse(&mut self, text: &str, target: Option<ColumnType>, pretty: bool) -> Result<String> {
        let node = crate::dml::parse_expression(text)?;
        let mut analyzed = self.analyze(&node)?;
        if let Some(target) = target {
            analyzed = self.coerce(analyzed, target);
        }
        Ok(Printer { pretty, indents: true, level: 0, wrap: 0 }.print(&analyzed, None, false))
    }

    /// constant returns a constant of a type, normalizing its text through the type's input and output when it can.
    fn constant(&mut self, text: &str, ty: ColumnType) -> TExpr {
        let normalized = if crate::cast::is_reg_type(ty.oid) {
            self.ctx.reg_value(crate::types::Value::Text(text.to_string()), ty.oid).ok().and_then(|v| v.output())
        } else {
            crate::cast::input(text, ty.oid).ok().and_then(|v| v.output())
        };
        TExpr::Const(Some(normalized.unwrap_or_else(|| text.to_string())), ty)
    }

    /// coerce converts an expression to a type, typing an untyped constant as the type and casting anything else
    /// implicitly.
    fn coerce(&mut self, e: TExpr, to: ColumnType) -> TExpr {
        let from = e.ty();
        if from.oid == to.oid {
            return e;
        }
        match e {
            TExpr::Const(Some(text), t) if t.oid == oid::UNKNOWN => {
                let constant = self.constant(&text, typ(to.oid));
                if to.modifier >= 0 { TExpr::Cast(Box::new(constant), to, true) } else { constant }
            }
            TExpr::Const(None, t) if t.oid == oid::UNKNOWN => TExpr::Const(None, to),
            TExpr::Array(items, _) if crate::array::is_array_type(to.oid) => {
                let element = typ(crate::expr::element_type(to.oid));
                if items.iter().all(|i| i.ty().oid == oid::UNKNOWN) {
                    let items = items.into_iter().map(|i| self.coerce(i, element)).collect();
                    return TExpr::Array(items, to.oid);
                }
                TExpr::Cast(Box::new(TExpr::Array(items, from.oid)), to, true)
            }
            other => TExpr::Cast(Box::new(other), to, true),
        }
    }

    /// operand_types returns the types Postgres' operator resolution gives the operands of a binary operator for the
    /// common built-in types: the type an untyped operand takes from the other, the cross-type integer and
    /// floating-point operators that need no casts, and promotion to numeric, floating-point, and text otherwise.
    fn operand_types(op: &str, l: u32, r: u32) -> (u32, u32) {
        let text_like = |t: u32| if t == oid::VARCHAR { oid::TEXT } else { t };
        match (l, r) {
            (oid::UNKNOWN, oid::UNKNOWN) => (oid::TEXT, oid::TEXT),
            (oid::UNKNOWN, t) => (text_like(t), text_like(t)),
            (t, oid::UNKNOWN) => (text_like(t), text_like(t)),
            _ if op.starts_with('~') || op.starts_with("!~") => (oid::TEXT, oid::TEXT),
            (a, b) if a == b => (text_like(a), text_like(b)),
            (a, b) if is_integer(a) && is_integer(b) || is_float(a) && is_float(b) => (a, b),
            (a, b) if is_datetime(a) && is_datetime(b) => (a, b),
            (a, oid::NUMERIC) if is_integer(a) => (oid::NUMERIC, oid::NUMERIC),
            (oid::NUMERIC, b) if is_integer(b) => (oid::NUMERIC, oid::NUMERIC),
            (a, b) if is_float(a) && (is_integer(b) || b == oid::NUMERIC) => (a, a),
            (a, b) if is_float(b) && (is_integer(a) || a == oid::NUMERIC) => (b, b),
            (a, b) if is_stringish(a) && is_stringish(b) => (oid::TEXT, oid::TEXT),
            (a, b) => (a, b),
        }
    }

    /// result_type returns the type of a binary operator's result over operands of the types.
    fn result_type(op: &str, l: u32, r: u32) -> u32 {
        if is_comparison(op) {
            return oid::BOOL;
        }
        if op == "||" {
            return if l == r { l } else { oid::TEXT };
        }
        let rank = |t: u32| match t {
            oid::INT2 => 1,
            oid::INT4 => 2,
            oid::INT8 => 3,
            _ => 0,
        };
        if is_integer(l) && is_integer(r) {
            return if rank(l) >= rank(r) { l } else { r };
        }
        if l == oid::FLOAT8 || r == oid::FLOAT8 {
            return oid::FLOAT8;
        }
        l
    }

    /// binary shapes a binary operator call.
    fn binary(&mut self, op: &str, l: TExpr, r: TExpr) -> TExpr {
        let (lt, rt) = Self::operand_types(op, l.ty().oid, r.ty().oid);
        let (l, r) = (self.coerce(l, typ(lt)), self.coerce(r, typ(rt)));
        let result = Self::result_type(op, lt, rt);
        TExpr::Op(op.to_string(), vec![l, r], result)
    }

    /// array_op shapes a comparison of a value with each element of an array.
    fn array_op(&mut self, op: &str, any: bool, l: TExpr, r: TExpr) -> TExpr {
        let element = crate::expr::element_type(r.ty().oid);
        let element = if r.ty().oid == oid::UNKNOWN { oid::UNKNOWN } else { element };
        let (lt, et) = Self::operand_types(op, l.ty().oid, element);
        let l = self.coerce(l, typ(lt));
        let array_type = crate::expr::array_of(et);
        let r = self.coerce(r, typ(array_type));
        TExpr::ArrayOp(op.to_string(), any, Box::new(l), Box::new(r))
    }

    /// common_type returns the type that a list of expressions resolves to, as Postgres' select_common_type does for
    /// the common built-in types, where untyped constants take the type of the others.
    fn common_type(items: &[TExpr]) -> u32 {
        let mut common = oid::UNKNOWN;
        for item in items {
            let t = item.ty().oid;
            if t == oid::UNKNOWN || t == common {
                continue;
            }
            common = match common {
                oid::UNKNOWN => t,
                c if is_integer(c) && is_integer(t) => {
                    if [oid::INT2, oid::INT4, oid::INT8].iter().position(|x| *x == t)
                        > [oid::INT2, oid::INT4, oid::INT8].iter().position(|x| *x == c)
                    {
                        t
                    } else {
                        c
                    }
                }
                c if (is_integer(c) || c == oid::NUMERIC) && (is_integer(t) || t == oid::NUMERIC) => oid::NUMERIC,
                c if is_float(c) || is_float(t) => oid::FLOAT8,
                c if is_stringish(c) && is_stringish(t) => oid::TEXT,
                c => c,
            };
        }
        if common == oid::UNKNOWN { oid::TEXT } else { common }
    }

    /// analyze shapes a parsed expression.
    fn analyze(&mut self, node: &Node) -> Result<TExpr> {
        let raw = || TExpr::Raw(crate::ddl::expression_text(node).unwrap_or_default());
        let Some(inner) = node.node.as_ref() else { return Ok(raw()) };
        Ok(match inner {
            NodeEnum::ColumnRef(column) => {
                let names: Vec<&str> = column.fields.iter().filter_map(node_name).collect();
                let (qualifier, name) = match names.as_slice() {
                    [relation, name] if self.qualifiers.contains(relation) => (Some(relation.to_string()), *name),
                    _ => (None, names.last().copied().unwrap_or_default()),
                };
                let star = column.fields.len() == names.len() + 1;
                if let Some(var) = self.resolve_column(&names, star) {
                    return Ok(var);
                }
                match self.columns.iter().find(|(n, _)| n == name) {
                    _ if star && names.len() == 1 && self.qualifiers.contains(&name) => {
                        TExpr::Var(Some(name.to_string()), "*".into(), typ(oid::RECORD))
                    }
                    Some((n, t)) if !star => TExpr::Var(qualifier, n.clone(), *t),
                    _ => raw(),
                }
            }
            NodeEnum::AConst(c) => match &c.val {
                None => TExpr::Const(None, typ(oid::UNKNOWN)),
                Some(Val::Ival(i)) => TExpr::Const(Some(i.ival.to_string()), typ(oid::INT4)),
                Some(Val::Fval(f)) => {
                    let text = f.fval.clone();
                    match text.parse::<i64>() {
                        Ok(_) if !text.contains(['.', 'e', 'E']) => TExpr::Const(Some(text), typ(oid::INT8)),
                        _ => self.constant(&text, typ(oid::NUMERIC)),
                    }
                }
                Some(Val::Sval(s)) => TExpr::Const(Some(s.sval.clone()), typ(oid::UNKNOWN)),
                Some(Val::Boolval(b)) => {
                    TExpr::Const(Some(if b.boolval { "true" } else { "false" }.into()), typ(oid::BOOL))
                }
                _ => raw(),
            },
            NodeEnum::ParamRef(p) => TExpr::Param(p.number as usize),
            NodeEnum::TypeCast(cast) => {
                let Some(type_name) = cast.type_name.as_ref() else { return Ok(raw()) };
                self.ctx.prepare_type(type_name)?;
                let Ok(target) = crate::expr::resolve_type_name(type_name) else { return Ok(raw()) };
                let Some(arg) = cast.arg.as_deref() else { return Ok(raw()) };
                let arg = self.analyze(arg)?;
                match arg {
                    TExpr::Const(Some(text), t) if t.oid == oid::UNKNOWN => {
                        let constant = self.constant(&text, typ(target.oid));
                        if target.modifier >= 0 { TExpr::Cast(Box::new(constant), target, false) } else { constant }
                    }
                    TExpr::Const(None, t) if t.oid == oid::UNKNOWN => TExpr::Const(None, target),
                    TExpr::Array(items, t) if t == oid::UNKNOWN || items.iter().all(|i| i.ty().oid == oid::UNKNOWN) => {
                        self.coerce(TExpr::Array(items, oid::UNKNOWN), target)
                    }
                    other if other.ty() == target => other,
                    other => TExpr::Cast(Box::new(other), target, false),
                }
            }
            NodeEnum::AExpr(e) => self.a_expr(e, &raw)?,
            NodeEnum::BoolExpr(b) => {
                let mut args = Vec::with_capacity(b.args.len());
                for arg in &b.args {
                    let analyzed = self.analyze(arg)?;
                    args.push(self.coerce(analyzed, typ(oid::BOOL)));
                }
                match BoolExprType::try_from(b.boolop) {
                    Ok(BoolExprType::AndExpr) => TExpr::Bool(true, args),
                    Ok(BoolExprType::OrExpr) => TExpr::Bool(false, args),
                    _ => TExpr::Not(Box::new(args.into_iter().next().unwrap_or_else(raw))),
                }
            }
            NodeEnum::NullTest(t) => {
                let Some(arg) = t.arg.as_deref() else { return Ok(raw()) };
                let arg = self.analyze(arg)?;
                TExpr::NullTest(Box::new(arg), NullTestType::try_from(t.nulltesttype) == Ok(NullTestType::IsNotNull))
            }
            NodeEnum::BooleanTest(t) => {
                let Some(arg) = t.arg.as_deref() else { return Ok(raw()) };
                let arg = self.analyze(arg)?;
                let arg = self.coerce(arg, typ(oid::BOOL));
                let keywords = match BoolTestType::try_from(t.booltesttype) {
                    Ok(BoolTestType::IsTrue) => " IS TRUE",
                    Ok(BoolTestType::IsNotTrue) => " IS NOT TRUE",
                    Ok(BoolTestType::IsFalse) => " IS FALSE",
                    Ok(BoolTestType::IsNotFalse) => " IS NOT FALSE",
                    Ok(BoolTestType::IsUnknown) => " IS UNKNOWN",
                    _ => " IS NOT UNKNOWN",
                };
                TExpr::BoolTest(Box::new(arg), keywords)
            }
            NodeEnum::SqlvalueFunction(f) => match SqlValueFunctionOp::try_from(f.op) {
                Ok(SqlValueFunctionOp::SvfopCurrentDate) => TExpr::Keyword("CURRENT_DATE", oid::DATE),
                Ok(SqlValueFunctionOp::SvfopCurrentTime) => TExpr::Keyword("CURRENT_TIME", oid::TIMETZ),
                Ok(SqlValueFunctionOp::SvfopCurrentTimestamp) => TExpr::Keyword("CURRENT_TIMESTAMP", oid::TIMESTAMPTZ),
                Ok(SqlValueFunctionOp::SvfopLocaltime) => TExpr::Keyword("LOCALTIME", oid::TIME),
                Ok(SqlValueFunctionOp::SvfopLocaltimestamp) => TExpr::Keyword("LOCALTIMESTAMP", oid::TIMESTAMP),
                Ok(SqlValueFunctionOp::SvfopCurrentRole) => TExpr::Keyword("CURRENT_ROLE", oid::NAME),
                Ok(SqlValueFunctionOp::SvfopCurrentUser) => TExpr::Keyword("CURRENT_USER", oid::NAME),
                Ok(SqlValueFunctionOp::SvfopUser) => TExpr::Keyword("USER", oid::NAME),
                Ok(SqlValueFunctionOp::SvfopSessionUser) => TExpr::Keyword("SESSION_USER", oid::NAME),
                Ok(SqlValueFunctionOp::SvfopCurrentCatalog) => TExpr::Keyword("CURRENT_CATALOG", oid::NAME),
                Ok(SqlValueFunctionOp::SvfopCurrentSchema) => TExpr::Keyword("CURRENT_SCHEMA", oid::NAME),
                _ => raw(),
            },
            NodeEnum::FuncCall(call) => return self.call(call),
            NodeEnum::SubLink(link) => return self.sublink(link),
            NodeEnum::CoalesceExpr(c) => {
                let items = c.args.iter().map(|a| self.analyze(a)).collect::<Result<Vec<_>>>()?;
                let common = Self::common_type(&items);
                let items = items.into_iter().map(|i| self.coerce(i, typ(common))).collect();
                TExpr::Coalesce(items, common)
            }
            NodeEnum::MinMaxExpr(m) => {
                let items = m.args.iter().map(|a| self.analyze(a)).collect::<Result<Vec<_>>>()?;
                let common = Self::common_type(&items);
                let items = items.into_iter().map(|i| self.coerce(i, typ(common))).collect();
                TExpr::MinMax(MinMaxOp::try_from(m.op) == Ok(MinMaxOp::IsGreatest), items, common)
            }
            NodeEnum::AArrayExpr(array) => {
                let items = array.elements.iter().map(|a| self.analyze(a)).collect::<Result<Vec<_>>>()?;
                if items.iter().all(|i| i.ty().oid == oid::UNKNOWN) {
                    TExpr::Array(items, oid::UNKNOWN)
                } else {
                    let common = Self::common_type(&items);
                    let items = items.into_iter().map(|i| self.coerce(i, typ(common))).collect();
                    TExpr::Array(items, crate::expr::array_of(common))
                }
            }
            NodeEnum::RowExpr(row) => TExpr::Row(row.args.iter().map(|a| self.analyze(a)).collect::<Result<Vec<_>>>()?),
            NodeEnum::CaseExpr(case) => {
                let mut arg = match case.arg.as_deref() {
                    Some(arg) => Some(self.analyze(arg)?),
                    None => None,
                };
                let mut whens = Vec::new();
                for when in &case.args {
                    let Some(NodeEnum::CaseWhen(w)) = when.node.as_ref() else { continue };
                    let (Some(condition), Some(result)) = (w.expr.as_deref(), w.result.as_deref()) else { continue };
                    whens.push((self.analyze(condition)?, self.analyze(result)?));
                }
                if let Some(test) = arg.take() {
                    let mut values = vec![test.clone()];
                    values.extend(whens.iter().map(|(v, _)| v.clone()));
                    let common = Self::common_type(&values);
                    arg = Some(self.coerce(test, typ(common)));
                    whens = whens.into_iter().map(|(v, r)| (self.coerce(v, typ(common)), r)).collect();
                } else {
                    whens = whens.into_iter().map(|(c, r)| (self.coerce(c, typ(oid::BOOL)), r)).collect();
                }
                let default = match case.defresult.as_deref() {
                    Some(d) => Some(self.analyze(d)?),
                    None => None,
                };
                let mut results: Vec<TExpr> = whens.iter().map(|(_, r)| r.clone()).collect();
                results.extend(default.clone());
                let common = Self::common_type(&results);
                let whens = whens.into_iter().map(|(c, r)| (c, self.coerce(r, typ(common)))).collect();
                let default = match default {
                    Some(d) => self.coerce(d, typ(common)),
                    None => TExpr::Const(None, typ(common)),
                };
                TExpr::Case(arg.map(Box::new), whens, Box::new(default), common)
            }
            _ => raw(),
        })
    }

    /// a_expr shapes an operator expression.
    fn a_expr(&mut self, e: &pg_query::protobuf::AExpr, raw: &dyn Fn() -> TExpr) -> Result<TExpr> {
        let op = e.name.iter().filter_map(node_name).next_back().unwrap_or_default().to_string();
        let kind = AExprKind::try_from(e.kind).unwrap_or(AExprKind::Undefined);
        let side = |n: &Option<Box<Node>>| n.as_deref().cloned();
        let (Some(right), left) = (side(&e.rexpr), side(&e.lexpr)) else { return Ok(raw()) };
        Ok(match kind {
            AExprKind::AexprOp => match left {
                None => {
                    let operand = self.analyze(&right)?;
                    let ty = operand.ty().oid;
                    TExpr::Op(op, vec![operand], ty)
                }
                Some(left) => {
                    let (l, r) = (self.analyze(&left)?, self.analyze(&right)?);
                    self.binary(&op, l, r)
                }
            },
            AExprKind::AexprLike | AExprKind::AexprIlike => {
                let Some(left) = left else { return Ok(raw()) };
                let (l, r) = (self.analyze(&left)?, self.analyze(&right)?);
                self.binary(&op, l, r)
            }
            AExprKind::AexprOpAny | AExprKind::AexprOpAll => {
                let Some(left) = left else { return Ok(raw()) };
                let (l, r) = (self.analyze(&left)?, self.analyze(&right)?);
                self.array_op(&op, kind == AExprKind::AexprOpAny, l, r)
            }
            AExprKind::AexprIn => {
                let Some(left) = left else { return Ok(raw()) };
                let Some(NodeEnum::List(list)) = right.node.as_ref() else { return Ok(raw()) };
                let l = self.analyze(&left)?;
                let items = list.items.iter().map(|i| self.analyze(i)).collect::<Result<Vec<_>>>()?;
                let mut all = vec![l.clone()];
                all.extend(items.iter().cloned());
                let common = Self::common_type(&all);
                let items: Vec<TExpr> = items.into_iter().map(|i| self.coerce(i, typ(common))).collect();
                let array = TExpr::Array(items, crate::expr::array_of(common));
                let any = op == "=";
                self.array_op(if any { "=" } else { "<>" }, any, l, array)
            }
            AExprKind::AexprDistinct | AExprKind::AexprNotDistinct => {
                let Some(left) = left else { return Ok(raw()) };
                let (l, r) = (self.analyze(&left)?, self.analyze(&right)?);
                let TExpr::Op(_, mut args, _) = self.binary("=", l, r) else { return Ok(raw()) };
                let r = args.pop().unwrap_or_else(raw);
                let l = args.pop().unwrap_or_else(raw);
                TExpr::Distinct(Box::new(l), Box::new(r), kind == AExprKind::AexprNotDistinct)
            }
            AExprKind::AexprNullif => {
                let Some(left) = left else { return Ok(raw()) };
                let (l, r) = (self.analyze(&left)?, self.analyze(&right)?);
                let ty = l.ty().oid;
                TExpr::NullIf(Box::new(l), Box::new(r), ty)
            }
            AExprKind::AexprBetween | AExprKind::AexprNotBetween => {
                let Some(left) = left else { return Ok(raw()) };
                let Some(NodeEnum::List(list)) = right.node.as_ref() else { return Ok(raw()) };
                let [low, high] = list.items.as_slice() else { return Ok(raw()) };
                let l = self.analyze(&left)?;
                let (low, high) = (self.analyze(low)?, self.analyze(high)?);
                if kind == AExprKind::AexprBetween {
                    let ge = self.binary(">=", l.clone(), low);
                    let le = self.binary("<=", l, high);
                    TExpr::Bool(true, vec![ge, le])
                } else {
                    let lt = self.binary("<", l.clone(), low);
                    let gt = self.binary(">", l, high);
                    TExpr::Bool(false, vec![lt, gt])
                }
            }
            _ => raw(),
        })
    }
}

/// Printer prints analyzed expressions as Postgres' get_rule_expr does, and queries as its get_query_def does, putting
/// CASE keywords and query clauses on their own lines when it indents.
struct Printer {
    pretty: bool,
    indents: bool,
    /// The indentation of the lines that keywords start, as Postgres' indentLevel counts it.
    level: i32,
    /// The column that target lists and FROM lists wrap after, or a negative number for no wrapping.
    wrap: i32,
}

/// The indentation steps of Postgres' ruleutils.c.
const INDENT_STD: i32 = 8;
const INDENT_JOIN: i32 = 4;
const INDENT_VAR: i32 = 4;
const INDENT_LIMIT: i32 = 40;

/// trim_spaces removes the spaces that end the text.
fn trim_spaces(out: &mut String) {
    let trimmed = out.trim_end_matches(' ').len();
    out.truncate(trimmed);
}

/// simple_op returns the operator of a binary operator call whose priority pretty-printing knows, `+ -` or `* / %`.
fn simple_op(e: &TExpr) -> Option<&str> {
    match e {
        TExpr::Op(op, args, _) if args.len() == 2 && matches!(op.as_str(), "+" | "-" | "*" | "/" | "%") => Some(op),
        _ => None,
    }
}

impl Printer {
    /// is_simple reports whether an expression prints without parentheses under its parent when pretty-printing,
    /// as Postgres' isSimpleNode decides.
    fn is_simple(&self, e: &TExpr, parent: &TExpr, first: bool) -> bool {
        match e {
            TExpr::Var(..)
            | TExpr::Const(..)
            | TExpr::Param(_)
            | TExpr::Func(..)
            | TExpr::Keyword(..)
            | TExpr::Array(..)
            | TExpr::Row(_)
            | TExpr::Coalesce(..)
            | TExpr::MinMax(..)
            | TExpr::NullIf(..)
            | TExpr::Case(..)
            | TExpr::Agg(_)
            | TExpr::Syntax(..)
            | TExpr::Raw(_) => true,
            TExpr::Cast(arg, _, _) => self.is_simple(arg, e, true),
            TExpr::Op(..) if matches!(parent, TExpr::Op(..)) => {
                let (Some(op), Some(parent_op)) = (simple_op(e), simple_op(parent)) else { return false };
                let high = |o: &str| matches!(o, "*" | "/" | "%");
                if high(op) && !high(parent_op) {
                    return true;
                }
                if !high(op) && high(parent_op) {
                    return false;
                }
                first
            }
            TExpr::Op(..) | TExpr::NullTest(..) | TExpr::BoolTest(..) | TExpr::Distinct(..) | TExpr::Sub(..) => {
                matches!(
                    parent,
                    TExpr::Bool(..)
                        | TExpr::Not(_)
                        | TExpr::Func(..)
                        | TExpr::Agg(_)
                        | TExpr::Array(..)
                        | TExpr::Row(_)
                        | TExpr::Coalesce(..)
                        | TExpr::MinMax(..)
                        | TExpr::NullIf(..)
                        | TExpr::Case(..)
                )
            }
            TExpr::Bool(and, _) => match parent {
                TExpr::Bool(parent_and, _) => *and || !parent_and,
                TExpr::Func(..)
                | TExpr::Agg(_)
                | TExpr::Array(..)
                | TExpr::Row(_)
                | TExpr::Coalesce(..)
                | TExpr::MinMax(..)
                | TExpr::NullIf(..)
                | TExpr::Case(..) => true,
                _ => false,
            },
            TExpr::Not(_) => matches!(parent, TExpr::Bool(..)) || matches!(parent, TExpr::Func(..)),
            TExpr::ArrayOp(..) => false,
        }
    }

    /// print_paren prints an expression under its parent, in parentheses when pretty-printing needs them.
    fn print_paren(&mut self, e: &TExpr, parent: &TExpr, showimplicit: bool, first: bool) -> String {
        let text = self.print(e, Some(parent), showimplicit);
        if self.pretty && !self.is_simple(e, parent, first) { format!("({text})") } else { text }
    }

    /// wrap puts text in parentheses unless pretty-printing.
    fn wrap(&self, text: String) -> String {
        if self.pretty { text } else { format!("({text})") }
    }

    /// constant prints a constant as Postgres' get_const_expr does, labeling it with its type when the text alone would
    /// not read back as that type, or never when `label` is unset.
    fn constant(text: Option<&str>, ty: ColumnType, label: bool) -> String {
        let type_name = || crate::cast::format_type(ty.oid, Some(ty.modifier)).unwrap_or_else(|| "unknown".into());
        let Some(text) = text else {
            return if label { format!("NULL::{}", type_name()) } else { "NULL".into() };
        };
        let quoted = format!("'{}'", text.replace('\'', "''"));
        let (body, needs_label) = match ty.oid {
            oid::INT4 if !text.starts_with('-') => (text.to_string(), false),
            oid::INT4 => (quoted, true),
            oid::NUMERIC if text.starts_with(|c: char| c.is_ascii_digit()) && text.contains(['e', 'E', '.']) => {
                (text.to_string(), ty.modifier >= 0)
            }
            oid::NUMERIC => (quoted, true),
            oid::BOOL => ((if text == "t" || text == "true" { "true" } else { "false" }).to_string(), false),
            oid::UNKNOWN => (quoted, false),
            _ => (quoted, true),
        };
        if label && needs_label { format!("{body}::{}", type_name()) } else { body }
    }

    /// context_keyword appends a keyword as Postgres' appendContextKeyword does: when the printer indents, on a new
    /// line indented to the level after adding `before` to it, plus `plus`, then adding `after` to the level.
    fn context_keyword(&mut self, out: &mut String, word: &str, before: i32, after: i32, plus: i32) {
        if self.indents {
            self.level += before;
            trim_spaces(out);
            out.push('\n');
            let amount = if self.level < INDENT_LIMIT {
                self.level.max(0) + plus
            } else {
                (INDENT_LIMIT + (self.level - INDENT_LIMIT) / 2) % INDENT_LIMIT + plus
            };
            out.push_str(&" ".repeat(amount as usize));
            out.push_str(word);
            self.level = (self.level + after).max(0);
        } else {
            out.push_str(word);
        }
    }

    /// print prints an expression as Postgres' get_rule_expr does.
    fn print(&mut self, e: &TExpr, parent: Option<&TExpr>, showimplicit: bool) -> String {
        match e {
            TExpr::Var(qualifier, name, ty) => {
                let name = if ty.oid == oid::RECORD { name.clone() } else { crate::engine::quote_identifier(name) };
                qualifier.as_ref().map_or(name.clone(), |q| format!("{}.{name}", crate::engine::quote_identifier(q)))
            }
            TExpr::Const(text, ty) => Self::constant(text.as_deref(), *ty, true),
            TExpr::Param(n) => format!("${n}"),
            TExpr::Raw(text) => text.clone(),
            TExpr::Sub(kind, test, op, query, _) => self.sublink(*kind, test.as_deref(), op, query, e),
            TExpr::Agg(agg) => self.aggregate(agg),
            TExpr::Syntax(name, args, _) => self.syntax(name, args, e),
            TExpr::Op(op, args, _) => {
                let text = match args.as_slice() {
                    [arg] => format!("{op} {}", self.print_paren(arg, e, true, true)),
                    [l, r] => {
                        format!("{} {op} {}", self.print_paren(l, e, true, true), self.print_paren(r, e, true, false))
                    }
                    _ => String::new(),
                };
                self.wrap(text)
            }
            TExpr::Bool(and, args) => {
                let joined: Vec<String> =
                    args.iter().enumerate().map(|(i, a)| self.print_paren(a, e, false, i == 0)).collect();
                self.wrap(joined.join(if *and { " AND " } else { " OR " }))
            }
            TExpr::Not(arg) => {
                let text = format!("NOT {}", self.print_paren(arg, e, false, true));
                self.wrap(text)
            }
            TExpr::NullTest(arg, not) => {
                let text =
                    format!("{} IS {}NULL", self.print_paren(arg, e, true, true), if *not { "NOT " } else { "" });
                self.wrap(text)
            }
            TExpr::BoolTest(arg, keywords) => {
                let text = format!("{}{keywords}", self.print_paren(arg, e, false, true));
                self.wrap(text)
            }
            TExpr::Distinct(l, r, not) => {
                let text = format!(
                    "{} IS {}DISTINCT FROM {}",
                    self.print_paren(l, e, true, true),
                    if *not { "NOT " } else { "" },
                    self.print_paren(r, e, true, false)
                );
                self.wrap(text)
            }
            TExpr::Func(name, args, _) => {
                if name.ends_with("(*)") {
                    return name.clone();
                }
                let args: Vec<String> = args.iter().map(|a| self.print(a, Some(e), true)).collect();
                format!("{name}({})", args.join(", "))
            }
            TExpr::Keyword(word, _) => word.to_string(),
            TExpr::Cast(arg, ty, implicit) => {
                if *implicit && !showimplicit {
                    return self.print_paren(arg, parent.unwrap_or(e), false, true);
                }
                let type_name = crate::cast::format_type(ty.oid, Some(ty.modifier)).unwrap_or_else(|| "unknown".into());
                let inner = match arg.as_ref() {
                    TExpr::Const(text, t) if t.oid == ty.oid && t.modifier == -1 => {
                        Self::constant(text.as_deref(), *t, false)
                    }
                    _ => {
                        let text = self.print_paren(arg, e, false, true);
                        self.wrap(text)
                    }
                };
                format!("{inner}::{type_name}")
            }
            TExpr::ArrayOp(op, any, l, r) => {
                let text = format!(
                    "{} {op} {} ({})",
                    self.print_paren(l, e, true, true),
                    if *any { "ANY" } else { "ALL" },
                    self.print_paren(r, e, true, false)
                );
                self.wrap(text)
            }
            TExpr::Array(items, _) => {
                let items: Vec<String> = items.iter().map(|i| self.print(i, Some(e), true)).collect();
                format!("ARRAY[{}]", items.join(", "))
            }
            TExpr::Row(items) => {
                let items: Vec<String> = items.iter().map(|i| self.print(i, Some(e), true)).collect();
                format!("ROW({})", items.join(", "))
            }
            TExpr::Coalesce(items, _) => {
                let items: Vec<String> = items.iter().map(|i| self.print(i, Some(e), true)).collect();
                format!("COALESCE({})", items.join(", "))
            }
            TExpr::MinMax(greatest, items, _) => {
                let items: Vec<String> = items.iter().map(|i| self.print(i, Some(e), true)).collect();
                format!("{}({})", if *greatest { "GREATEST" } else { "LEAST" }, items.join(", "))
            }
            TExpr::NullIf(l, r, _) => {
                format!("NULLIF({}, {})", self.print(l, Some(e), true), self.print(r, Some(e), true))
            }
            TExpr::Case(arg, whens, default, _) => {
                let mut out = String::new();
                self.context_keyword(&mut out, "CASE", 0, INDENT_VAR, 0);
                if let Some(arg) = arg {
                    out.push(' ');
                    out.push_str(&self.print(arg, Some(e), true));
                }
                for (condition, result) in whens {
                    if !self.indents {
                        out.push(' ');
                    }
                    self.context_keyword(&mut out, "WHEN ", 0, 0, 0);
                    out.push_str(&self.print(condition, Some(e), false));
                    out.push_str(" THEN ");
                    out.push_str(&self.print(result, Some(e), true));
                }
                if !self.indents {
                    out.push(' ');
                }
                self.context_keyword(&mut out, "ELSE ", 0, 0, 0);
                out.push_str(&self.print(default, Some(e), true));
                if !self.indents {
                    out.push(' ');
                }
                self.context_keyword(&mut out, "END", -INDENT_VAR, 0, 0);
                out
            }
        }
    }
}
