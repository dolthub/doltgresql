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

//! Window functions: functions over a row's partition and frame, which a Window plan node computes for every row.

use std::cmp::Ordering;

use pg_query::protobuf::{FuncCall, WindowDef};
use pg_query::{Node, NodeEnum};

use crate::catalog::ColumnType;
use crate::error::{PgError, Result, code};
use crate::expr::{Binder, Expr, arg_location, coerce, compare_values, position, typ};
use crate::functions::aggregate::{Accumulator, AggCall};
use crate::oid;
use crate::plan::SortKey;
use crate::query::Ctx;
use crate::types::Value;

/// Frame option bits of a window definition, as Postgres' parser sets them.
pub mod frame {
    pub const NONDEFAULT: i32 = 0x1;
    pub const RANGE: i32 = 0x2;
    pub const ROWS: i32 = 0x4;
    pub const GROUPS: i32 = 0x8;
    pub const START_UNBOUNDED_PRECEDING: i32 = 0x20;
    pub const END_UNBOUNDED_FOLLOWING: i32 = 0x100;
    pub const START_CURRENT_ROW: i32 = 0x200;
    pub const END_CURRENT_ROW: i32 = 0x400;
    pub const START_OFFSET_PRECEDING: i32 = 0x800;
    pub const END_OFFSET_PRECEDING: i32 = 0x1000;
    pub const START_OFFSET_FOLLOWING: i32 = 0x2000;
    pub const END_OFFSET_FOLLOWING: i32 = 0x4000;
    pub const EXCLUDE_CURRENT_ROW: i32 = 0x8000;
    pub const EXCLUDE_GROUP: i32 = 0x10000;
    pub const EXCLUDE_TIES: i32 = 0x20000;
}

/// WindowKind is what a window function computes.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub enum WindowKind {
    RowNumber,
    Rank,
    DenseRank,
    PercentRank,
    CumeDist,
    Ntile,
    Lag,
    Lead,
    FirstValue,
    LastValue,
    NthValue,
    /// An aggregate over the frame, by its index among the aggregates.
    Aggregate(usize),
}

/// Bound is a bound of a window frame.
#[derive(Clone, Debug, PartialEq)]
pub enum Bound {
    UnboundedPreceding,
    Preceding(Expr),
    CurrentRow,
    Following(Expr),
    UnboundedFollowing,
}

/// WindowCall is a window function call: its function, arguments, partition, order, and frame, all over the rows the
/// window node reads.
#[derive(Clone, Debug, PartialEq)]
pub struct WindowCall {
    pub kind: WindowKind,
    pub args: Vec<Expr>,
    pub distinct: bool,
    pub filter: Option<Expr>,
    pub partition: Vec<Expr>,
    pub order: Vec<SortKey>,
    /// The frame's mode bits, start, and end.
    pub options: i32,
    pub start: Bound,
    pub end: Bound,
    pub ret: ColumnType,
    /// The comparisons of a RANGE frame with offsets.
    pub range: Option<RangeKeys>,
    /// The conditions that hold of the call's values only up to some row of each partition, as Postgres'
    /// WindowFunc.runCondition holds them.
    pub run_condition: Vec<RunCondition>,
}

/// RunCondition is a comparison of a window function's value with a value that does not change, which holds until
/// some row of a partition and never again after it, as Postgres' WindowFuncRunCondition is: the comparison, whether
/// the call is on its left, and the other value.
#[derive(Clone, Debug, PartialEq)]
pub struct RunCondition {
    pub op: crate::expr::CmpOp,
    pub wfunc_left: bool,
    pub arg: Expr,
}

/// Monotonic is how a window function's value moves over the rows of a partition, as Postgres' MonotonicFunction
/// reports it: whether it never decreases, and whether it never increases.
#[derive(Clone, Copy, Debug, Default)]
pub struct Monotonic {
    pub increasing: bool,
    pub decreasing: bool,
}

/// RangeKeys are what a RANGE frame with offsets compares rows by: the ordering value as the type of the ordering value
/// plus an offset, and for each offset bound that sum or difference with whether rows in the frame are at most it, as
/// Postgres' in_range functions decide.
#[derive(Clone, Debug, PartialEq)]
pub struct RangeKeys {
    pub key: Expr,
    pub start: Option<(Expr, bool)>,
    pub end: Option<(Expr, bool)>,
}

/// window_kind returns the window function of a name, or None for a name that is not one.
fn window_kind(name: &str) -> Option<(WindowKind, &'static [u32], u32)> {
    Some(match name {
        "row_number" => (WindowKind::RowNumber, &[], oid::INT8),
        "rank" => (WindowKind::Rank, &[], oid::INT8),
        "dense_rank" => (WindowKind::DenseRank, &[], oid::INT8),
        "percent_rank" => (WindowKind::PercentRank, &[], oid::FLOAT8),
        "cume_dist" => (WindowKind::CumeDist, &[], oid::FLOAT8),
        "ntile" => (WindowKind::Ntile, &[oid::INT4], oid::INT4),
        "lag" => (WindowKind::Lag, &[], 0),
        "lead" => (WindowKind::Lead, &[], 0),
        "first_value" => (WindowKind::FirstValue, &[], 0),
        "last_value" => (WindowKind::LastValue, &[], 0),
        "nth_value" => (WindowKind::NthValue, &[], 0),
        _ => return None,
    })
}

/// is_window_function reports whether a name is a window function.
pub fn is_window_function(name: &str) -> bool {
    window_kind(name).is_some()
}

/// has_window reports whether an expression calls a window function.
pub fn has_window(node: &Node) -> bool {
    node.node.as_ref().is_some_and(|n| {
        n.nodes().into_iter().any(|(n, ..)| matches!(n, pg_query::NodeRef::FuncCall(f) if f.over.is_some()))
    })
}

impl<'b, 'a> Binder<'b, 'a> {
    /// window_definition returns a call's window, resolving a named window and merging a window it refers to.
    fn window_definition(&self, over: &WindowDef) -> Result<WindowDef> {
        let named = |name: &str| {
            self.named_windows.iter().find(|w| w.name == name).cloned().ok_or_else(|| PgError {
                position: position(over.location),
                ..PgError::new(code::UNDEFINED_OBJECT, format!("window \"{name}\" does not exist"))
            })
        };
        if !over.name.is_empty() && over.partition_clause.is_empty() && over.order_clause.is_empty() {
            let resolved = self.window_definition(&WindowDef { name: String::new(), ..named(&over.name)? })?;
            return Ok(WindowDef { name: over.name.clone(), ..resolved });
        }
        if over.refname.is_empty() {
            return Ok(over.clone());
        }
        let base = self.window_definition(&WindowDef { name: String::new(), ..named(&over.refname)? })?;
        if !over.partition_clause.is_empty() {
            return Err(PgError {
                position: position(over.location),
                ..PgError::new(
                    code::WINDOWING_ERROR,
                    format!("cannot override PARTITION BY clause of window \"{}\"", over.refname),
                )
            });
        }
        if !over.order_clause.is_empty() && !base.order_clause.is_empty() {
            return Err(PgError {
                position: position(over.location),
                ..PgError::new(
                    code::WINDOWING_ERROR,
                    format!("cannot override ORDER BY clause of window \"{}\"", over.refname),
                )
            });
        }
        if base.frame_options & 1 != 0 {
            return Err(PgError {
                position: position(over.location),
                ..PgError::new(
                    code::WINDOWING_ERROR,
                    format!("cannot copy window \"{}\" because it has a frame clause", over.refname),
                )
            });
        }
        Ok(WindowDef {
            partition_clause: base.partition_clause,
            order_clause: if over.order_clause.is_empty() { base.order_clause } else { over.order_clause.clone() },
            ..over.clone()
        })
    }

    /// window_call binds a call with an OVER clause.
    pub fn window_call(&mut self, name: &str, call: &FuncCall) -> Result<(Expr, ColumnType)> {
        let Some(mut windows) = self.windows.take() else {
            return Err(PgError {
                position: position(call.location),
                ..PgError::new(code::WINDOWING_ERROR, format!("window functions are not allowed in {}", self.clause))
            });
        };
        let result = self.window_call_inner(name, call);
        let (window, ty) = match result {
            Ok(r) => r,
            Err(err) => {
                self.windows = Some(windows);
                return Err(err);
            }
        };
        windows.push(window);
        let k = windows.len() - 1;
        self.windows = Some(windows);
        Ok((Expr::WindowRef(k), ty))
    }

    /// window_call_inner binds a window call's function, arguments, and window.
    fn window_call_inner(&mut self, name: &str, call: &FuncCall) -> Result<(WindowCall, ColumnType)> {
        let over = call.over.as_deref().ok_or_else(|| PgError::internal("a window call without OVER"))?;
        let over = self.window_definition(over)?;
        let mut bound = Vec::with_capacity(call.args.len());
        for arg in &call.args {
            if has_window(arg) {
                return Err(PgError {
                    position: position(arg_location(arg)),
                    ..PgError::new(code::WINDOWING_ERROR, "window function calls cannot be nested")
                });
            }
            bound.push(self.bind(arg)?);
        }
        let types: Vec<u32> = bound.iter().map(|(_, t)| t.oid).collect();
        let (kind, args, ret) = match window_kind(name) {
            Some((kind, params, ret)) => {
                let mut args = Vec::new();
                let ret = match kind {
                    WindowKind::Lag
                    | WindowKind::Lead
                    | WindowKind::FirstValue
                    | WindowKind::LastValue
                    | WindowKind::NthValue => {
                        let expected = match kind {
                            WindowKind::Lag | WindowKind::Lead => 1..=3,
                            WindowKind::NthValue => 2..=2,
                            _ => 1..=1,
                        };
                        if !expected.contains(&bound.len()) {
                            return Err(missing_function(name, &types, call.location));
                        }
                        let value_type = if types[0] == oid::UNKNOWN { typ(oid::TEXT) } else { bound[0].1 };
                        for (i, (b, node)) in bound.into_iter().zip(&call.args).enumerate() {
                            let target = if i == 0 || (i == 2 && kind != WindowKind::NthValue) {
                                value_type
                            } else {
                                typ(oid::INT4)
                            };
                            args.push(coerce(b, target, false, arg_location(node))?.0);
                        }
                        value_type
                    }
                    _ => {
                        if bound.len() != params.len() {
                            return Err(missing_function(name, &types, call.location));
                        }
                        for ((b, node), &p) in bound.into_iter().zip(&call.args).zip(params) {
                            args.push(coerce(b, typ(p), false, arg_location(node))?.0);
                        }
                        typ(ret)
                    }
                };
                (kind, args, ret)
            }
            None if call.agg_star || crate::functions::aggregate::exists(None, name) => {
                let (index, arg_types, ret) = crate::functions::aggregate::resolve(name, &types, call.location)?;
                let mut args = Vec::new();
                for ((b, &t), node) in bound.into_iter().zip(&arg_types).zip(&call.args) {
                    args.push(coerce(b, typ(t), false, arg_location(node))?.0);
                }
                (WindowKind::Aggregate(index), args, typ(ret))
            }
            None => {
                let message = match call.agg_distinct {
                    true => format!("DISTINCT specified, but {name} is not an aggregate function"),
                    false => format!("OVER specified, but {name} is not a window function nor an aggregate function"),
                };
                return Err(PgError {
                    position: position(call.location),
                    ..PgError::new(code::WRONG_OBJECT_TYPE, message)
                });
            }
        };
        if call.agg_distinct {
            return Err(PgError {
                position: position(call.location),
                ..PgError::new(code::FEATURE_NOT_SUPPORTED, "DISTINCT is not implemented for window functions")
            });
        }
        let filter = match call.agg_filter.as_deref() {
            Some(node) => Some(coerce(self.bind(node)?, typ(oid::BOOL), false, -1)?.0),
            None => None,
        };
        let mut partition = Vec::new();
        for node in &over.partition_clause {
            partition.push(self.bind(node)?.0);
        }
        let (mut order, mut order_types) = (Vec::new(), Vec::new());
        for sort in &over.order_clause {
            let Some(NodeEnum::SortBy(sort)) = sort.node.as_ref() else { continue };
            let node = sort.node.as_deref().ok_or_else(|| PgError::internal("ORDER BY without a key"))?;
            let descending = pg_query::protobuf::SortByDir::try_from(sort.sortby_dir)
                == Ok(pg_query::protobuf::SortByDir::SortbyDesc);
            let nulls_first = match pg_query::protobuf::SortByNulls::try_from(sort.sortby_nulls) {
                Ok(pg_query::protobuf::SortByNulls::SortbyNullsFirst) => true,
                Ok(pg_query::protobuf::SortByNulls::SortbyNullsLast) => false,
                _ => descending,
            };
            let (expr, ty) = self.bind(node)?;
            order.push(SortKey { expr, descending, nulls_first });
            order_types.push(ty);
        }
        let options = over.frame_options;
        let offsets = frame::START_OFFSET_PRECEDING
            | frame::START_OFFSET_FOLLOWING
            | frame::END_OFFSET_PRECEDING
            | frame::END_OFFSET_FOLLOWING;
        let ranged = options & frame::RANGE != 0 && options & offsets != 0;
        if ranged && order.len() != 1 {
            return Err(PgError {
                position: position(over.location),
                ..PgError::new(
                    code::WINDOWING_ERROR,
                    "RANGE with offset PRECEDING/FOLLOWING requires exactly one ORDER BY column",
                )
            });
        }
        let column = order_types.first().copied().unwrap_or(typ(oid::INT8));
        let target = match column.oid {
            oid::INT2 | oid::INT4 | oid::INT8 => oid::INT8,
            oid::FLOAT4 | oid::FLOAT8 => oid::FLOAT8,
            oid::NUMERIC => oid::NUMERIC,
            oid::DATE | oid::TIME | oid::TIMETZ | oid::TIMESTAMP | oid::TIMESTAMPTZ | oid::INTERVAL => oid::INTERVAL,
            _ => 0,
        };
        let mut offset = |node: &Option<Box<Node>>| -> Result<Expr> {
            let node = node.as_deref().ok_or_else(|| PgError::internal("a frame offset without a value"))?;
            let bound = self.bind(node)?;
            if !ranged {
                return Ok(coerce(bound, typ(oid::INT8), false, arg_location(node))?.0);
            }
            let column_name = crate::cast::type_display(column.oid);
            let unsupported = |message: String| PgError {
                position: position(arg_location(node)),
                ..PgError::new(code::FEATURE_NOT_SUPPORTED, message)
            };
            if target == 0 {
                return Err(unsupported(format!(
                    "RANGE with offset PRECEDING/FOLLOWING is not supported for column type {column_name}"
                )));
            }
            let from = crate::cast::type_display(bound.1.oid).into_owned();
            coerce(bound, typ(target), false, arg_location(node)).map(|b| b.0).map_err(|_| PgError {
                hint: Some("Cast the offset value to an appropriate type.".into()),
                ..unsupported(format!(
                    "RANGE with offset PRECEDING/FOLLOWING is not supported for column type {column_name} and offset type {from}"
                ))
            })
        };
        let start = if options & frame::START_CURRENT_ROW != 0 {
            Bound::CurrentRow
        } else if options & frame::START_OFFSET_PRECEDING != 0 {
            Bound::Preceding(offset(&over.start_offset)?)
        } else if options & frame::START_OFFSET_FOLLOWING != 0 {
            Bound::Following(offset(&over.start_offset)?)
        } else {
            Bound::UnboundedPreceding
        };
        let end = if options & frame::END_UNBOUNDED_FOLLOWING != 0 {
            Bound::UnboundedFollowing
        } else if options & frame::END_OFFSET_PRECEDING != 0 {
            Bound::Preceding(offset(&over.end_offset)?)
        } else if options & frame::END_OFFSET_FOLLOWING != 0 {
            Bound::Following(offset(&over.end_offset)?)
        } else {
            Bound::CurrentRow
        };
        let range = match ranged {
            false => None,
            true => {
                let descending = order[0].descending;
                let mut sum_type = column;
                let mut sum = |bound: &Bound, at_most: bool| -> Result<Option<(Expr, bool)>> {
                    let (offset, preceding) = match bound {
                        Bound::Preceding(e) => (e, true),
                        Bound::Following(e) => (e, false),
                        _ => return Ok(None),
                    };
                    let op = if preceding != descending { "-" } else { "+" };
                    let (sum, ty) =
                        self.binary(op, (order[0].expr.clone(), column), (offset.clone(), typ(target)), -1)?;
                    sum_type = ty;
                    Ok(Some((sum, at_most != descending)))
                };
                let (start, end) = (sum(&start, false)?, sum(&end, true)?);
                let key = coerce((order[0].expr.clone(), column), sum_type, true, -1)?.0;
                Some(RangeKeys { key, start, end })
            }
        };
        let run_condition = Vec::new();
        let call = WindowCall {
            kind,
            args,
            distinct: false,
            filter,
            partition,
            order,
            options,
            start,
            end,
            ret,
            range,
            run_condition,
        };
        Ok((call, ret))
    }
}

/// missing_function returns Postgres' error for a window function called with the wrong arguments.
fn missing_function(name: &str, types: &[u32], location: i32) -> PgError {
    PgError {
        position: position(location),
        hint: Some(
            "No function matches the given name and argument types. You might need to add explicit type casts.".into(),
        ),
        ..PgError::new(
            code::UNDEFINED_FUNCTION,
            format!(
                "function {name}({}) does not exist",
                types.iter().map(|&t| crate::cast::type_display(t)).collect::<Vec<_>>().join(", ")
            ),
        )
    }
}

/// compare_keys orders rows by sort key values, as ORDER BY does.
fn compare_keys(keys: &[SortKey], a: &[Value], b: &[Value]) -> Ordering {
    for (i, key) in keys.iter().enumerate() {
        let ordering = match (&a[i], &b[i]) {
            (Value::Null, Value::Null) => Ordering::Equal,
            (Value::Null, _) => {
                if key.nulls_first {
                    Ordering::Less
                } else {
                    Ordering::Greater
                }
            }
            (_, Value::Null) => {
                if key.nulls_first {
                    Ordering::Greater
                } else {
                    Ordering::Less
                }
            }
            (l, r) => {
                let o = compare_values(l, r);
                if key.descending { o.reverse() } else { o }
            }
        };
        if ordering != Ordering::Equal {
            return ordering;
        }
    }
    Ordering::Equal
}

/// offset evaluates a frame offset, which must not be NULL or negative.
fn offset(ctx: &mut Ctx<'_>, expr: &Expr, row: &[Value]) -> Result<usize> {
    match expr.eval(ctx, row)? {
        Value::Int8(n) if n >= 0 => Ok(n as usize),
        Value::Null => Err(PgError::new(code::NULL_VALUE_NOT_ALLOWED, "frame starting offset must not be null")),
        _ => Err(PgError::new(code::INVALID_PRECEDING_OR_FOLLOWING_SIZE, "frame starting offset must not be negative")),
    }
}

/// Partition is one partition of a window's rows in window order, with its rows' peer groups.
struct Partition<'a> {
    /// Each row's partition key values, ordering values, and position in the input.
    part: &'a [(Vec<Value>, Vec<Value>, usize)],
    members: Vec<&'a Vec<Value>>,
    /// The first position of each peer group, in order.
    group_starts: Vec<usize>,
    /// The peer group of each position.
    group_of: Vec<usize>,
}

impl Partition<'_> {
    /// group_end returns the last position of a peer group.
    fn group_end(&self, group: usize) -> usize {
        self.group_starts.get(group + 1).map_or(self.part.len(), |&s| s) - 1
    }

    /// peers returns the first and last positions of the rows that sort equal to the row at a position.
    fn peers(&self, position: usize) -> (usize, usize) {
        let group = self.group_of[position];
        (self.group_starts[group], self.group_end(group))
    }
}

impl WindowCall {
    /// monotonic returns how the call's value moves over the rows of a partition, as the support functions of Postgres'
    /// ranking functions and of count answer SupportRequestWFuncMonotonic: ranks never decrease, and a count never
    /// decreases when its frame starts at the partition's start and never increases when it ends at the partition's
    /// end, and stays the same without an ORDER BY.
    pub fn monotonic(&self) -> Monotonic {
        match self.kind {
            WindowKind::RowNumber
            | WindowKind::Rank
            | WindowKind::DenseRank
            | WindowKind::PercentRank
            | WindowKind::CumeDist
            | WindowKind::Ntile => Monotonic { increasing: true, decreasing: false },
            WindowKind::Aggregate(index) if crate::functions::aggregate::AGGREGATES[index].name == "count" => {
                match self.order.is_empty() {
                    true => Monotonic { increasing: true, decreasing: true },
                    false => Monotonic {
                        increasing: self.start == Bound::UnboundedPreceding,
                        decreasing: self.end == Bound::UnboundedFollowing,
                    },
                }
            }
            _ => Monotonic::default(),
        }
    }

    /// ignores_frame reports whether the call's value is the same in any frame, so that its window's frame can become
    /// ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW, as the support functions of Postgres' ranking functions answer
    /// SupportRequestOptimizeWindowClause.
    pub fn ignores_frame(&self) -> bool {
        matches!(
            self.kind,
            WindowKind::RowNumber
                | WindowKind::Rank
                | WindowKind::DenseRank
                | WindowKind::PercentRank
                | WindowKind::CumeDist
                | WindowKind::Ntile
        )
    }

    /// same_window reports whether two calls share their window: its partition, order, and frame.
    pub fn same_window(&self, other: &WindowCall) -> bool {
        self.partition == other.partition
            && self.order == other.order
            && self.options == other.options
            && self.start == other.start
            && self.end == other.end
    }

    /// name returns the name of the call's function.
    pub fn name(&self) -> &'static str {
        match self.kind {
            WindowKind::RowNumber => "row_number",
            WindowKind::Rank => "rank",
            WindowKind::DenseRank => "dense_rank",
            WindowKind::PercentRank => "percent_rank",
            WindowKind::CumeDist => "cume_dist",
            WindowKind::Ntile => "ntile",
            WindowKind::Lag => "lag",
            WindowKind::Lead => "lead",
            WindowKind::FirstValue => "first_value",
            WindowKind::LastValue => "last_value",
            WindowKind::NthValue => "nth_value",
            WindowKind::Aggregate(index) => crate::functions::aggregate::AGGREGATES[index].name,
        }
    }

    /// compute returns the call's value for each row of the input, in input order.
    pub fn compute(&self, ctx: &mut Ctx<'_>, rows: &[Vec<Value>]) -> Result<Vec<Value>> {
        let mut keyed: Vec<(Vec<Value>, Vec<Value>, usize)> = Vec::with_capacity(rows.len());
        for (i, row) in rows.iter().enumerate() {
            let partition = self.partition.iter().map(|e| e.eval(ctx, row)).collect::<Result<Vec<_>>>()?;
            let mut order = self.order.iter().map(|k| k.expr.eval(ctx, row)).collect::<Result<Vec<_>>>()?;
            if let Some(range) = &self.range {
                order.push(range.key.eval(ctx, row)?);
            }
            keyed.push((partition, order, i));
        }
        let partition_keys: Vec<SortKey> =
            self.partition.iter().map(|e| SortKey { expr: e.clone(), descending: false, nulls_first: false }).collect();
        keyed.sort_by(|a, b| {
            compare_keys(&partition_keys, &a.0, &b.0).then_with(|| compare_keys(&self.order, &a.1, &b.1))
        });
        let mut out = vec![Value::Null; rows.len()];
        let mut start = 0;
        while start < keyed.len() {
            let mut end = start + 1;
            while end < keyed.len() && compare_keys(&partition_keys, &keyed[start].0, &keyed[end].0) == Ordering::Equal
            {
                end += 1;
            }
            let part = &keyed[start..end];
            let (mut group_starts, mut group_of) = (Vec::new(), Vec::with_capacity(part.len()));
            for i in 0..part.len() {
                if i == 0
                    || (!self.order.is_empty()
                        && compare_keys(&self.order, &part[i - 1].1, &part[i].1) != Ordering::Equal)
                {
                    group_starts.push(i);
                }
                group_of.push(group_starts.len() - 1);
            }
            let partition =
                Partition { part, members: part.iter().map(|k| &rows[k.2]).collect(), group_starts, group_of };
            self.partition_values(ctx, &partition, &mut out)?;
            start = end;
        }
        Ok(out)
    }

    /// partition_values computes the call for every row of a partition.
    fn partition_values(&self, ctx: &mut Ctx<'_>, p: &Partition<'_>, out: &mut [Value]) -> Result<()> {
        let excludes = frame::EXCLUDE_CURRENT_ROW | frame::EXCLUDE_GROUP | frame::EXCLUDE_TIES;
        let call = match self.kind {
            WindowKind::Aggregate(index) => Some(AggCall {
                index,
                args: self.args.clone(),
                distinct: self.distinct,
                filter: self.filter.clone(),
                order: Vec::new(),
                ret: self.ret.oid,
                user: None,
            }),
            _ => None,
        };
        if let Some(call) = &call
            && self.start == Bound::UnboundedPreceding
            && self.options & excludes == 0
        {
            return self.running(ctx, p, call, out);
        }
        if let Some(call) = &call
            && self.options & excludes == 0
            && Accumulator::new(call).invertible()
        {
            return self.moving(ctx, p, call, out);
        }
        for position in 0..p.part.len() {
            out[p.part[position].2] = self.value(ctx, p, position, call.as_ref())?;
        }
        Ok(())
    }

    /// running computes an aggregate whose frames start at the partition's first row, adding each row to one running
    /// aggregate as the frames grow, as Postgres' window aggregation does when frame heads never move.
    fn running(&self, ctx: &mut Ctx<'_>, p: &Partition<'_>, call: &AggCall, out: &mut [Value]) -> Result<()> {
        let mut accumulator = Accumulator::new(call);
        let mut added = 0;
        let mut previous: Option<(usize, Value)> = None;
        for position in 0..p.part.len() {
            let (_, end) = self.bounds(ctx, p, position)?;
            let end = (end + 1).clamp(0, p.part.len() as isize) as usize;
            if end < added {
                accumulator = Accumulator::new(call);
                added = 0;
            }
            while added < end {
                accumulator.add(ctx, call, p.members[added])?;
                added += 1;
            }
            let value = match &previous {
                Some((last, value)) if *last == end => value.clone(),
                _ => accumulator.peek(ctx, call)?,
            };
            out[p.part[position].2] = value.clone();
            previous = Some((end, value));
        }
        Ok(())
    }

    /// moving computes an aggregate whose frames move forward through the partition, adding the rows that enter each
    /// frame and removing the ones that leave it, as Postgres does for aggregates with inverse transitions, and
    /// starting over when the aggregate can't remove a row.
    fn moving(&self, ctx: &mut Ctx<'_>, p: &Partition<'_>, call: &AggCall, out: &mut [Value]) -> Result<()> {
        let n = p.part.len();
        let mut accumulator = Accumulator::new(call);
        let (mut first, mut end) = (0, 0);
        for position in 0..n {
            let (start, last) = self.bounds(ctx, p, position)?;
            let start = start.min(n);
            let stop = ((last + 1).clamp(0, n as isize) as usize).max(start);
            let mut restart = start < first || stop < end;
            while !restart && first < start {
                if first < end && !accumulator.remove(ctx, call, p.members[first])? {
                    restart = true;
                }
                first += 1;
            }
            if restart {
                accumulator = Accumulator::new(call);
                (first, end) = (start, start);
            }
            end = end.max(first);
            while end < stop {
                accumulator.add(ctx, call, p.members[end])?;
                end += 1;
            }
            out[p.part[position].2] = accumulator.peek(ctx, call)?;
        }
        Ok(())
    }

    /// range_edge returns the first position of a RANGE frame with an offset start, or the last position of one with an
    /// offset end, where the rows of a NULL ordering value frame only their peers, as Postgres' window aggregation does.
    fn range_edge(
        &self,
        ctx: &mut Ctx<'_>,
        p: &Partition<'_>,
        row: &[Value],
        position: usize,
        offset: &Expr,
        starting: bool,
    ) -> Result<isize> {
        let part = p.part;
        let sums = self.range.as_ref().and_then(|r| if starting { r.start.as_ref() } else { r.end.as_ref() });
        let (sum, at_most) = sums.ok_or_else(|| PgError::internal("a RANGE frame bound without an offset"))?;
        let negative = match offset.eval(ctx, row)? {
            Value::Null => {
                let which = if starting { "starting" } else { "ending" };
                return Err(PgError::new(
                    code::NULL_VALUE_NOT_ALLOWED,
                    format!("frame {which} offset must not be null"),
                ));
            }
            Value::Int8(n) => n < 0,
            Value::Float8(f) => f.is_nan() || f < 0.0,
            Value::Numeric(n) => n.is_negative() || matches!(n, crate::numeric::Numeric::NaN),
            Value::Interval(iv) => iv.cmp_key() < 0,
            _ => false,
        };
        if negative {
            return Err(PgError::new(
                code::INVALID_PRECEDING_OR_FOLLOWING_SIZE,
                "invalid preceding or following size in window function",
            ));
        }
        if part[position].1[0].is_null() {
            let (first, last) = p.peers(position);
            return Ok(if starting { first } else { last } as isize);
        }
        let bound = sum.eval(ctx, row)?;
        let k = self.order.len();
        let lo = part.iter().position(|p| !p.1[k].is_null()).unwrap_or(part.len());
        let hi = part.iter().rposition(|p| !p.1[k].is_null()).map_or(lo, |i| i + 1);
        let inside = |p: &(Vec<Value>, Vec<Value>, usize)| match compare_values(&p.1[k], &bound) {
            Ordering::Greater => !at_most,
            Ordering::Less => *at_most,
            Ordering::Equal => true,
        };
        let rows = &part[lo..hi];
        Ok(match starting {
            true => (lo + rows.partition_point(|p| !inside(p))) as isize,
            false => (lo + rows.partition_point(inside)) as isize - 1,
        })
    }

    /// bounds returns the first and last positions of a row's frame before any exclusion, where a last position
    /// before the first makes the frame empty.
    fn bounds(&self, ctx: &mut Ctx<'_>, p: &Partition<'_>, position: usize) -> Result<(usize, isize)> {
        let n = p.part.len();
        let rows_mode = self.options & frame::ROWS != 0;
        let groups_mode = self.options & frame::GROUPS != 0;
        let (peer_first, peer_last) = p.peers(position);
        let row = p.members[position];
        let group = p.group_of[position];
        let start = match &self.start {
            Bound::Preceding(e) | Bound::Following(e) if self.range.is_some() => {
                self.range_edge(ctx, p, row, position, e, true)? as usize
            }
            Bound::UnboundedPreceding => 0,
            Bound::CurrentRow if rows_mode => position,
            Bound::CurrentRow => peer_first,
            Bound::Preceding(e) => {
                let k = offset(ctx, e, row)?;
                if groups_mode { p.group_starts[group.saturating_sub(k)] } else { position.saturating_sub(k) }
            }
            Bound::Following(e) => {
                let k = offset(ctx, e, row)?;
                if groups_mode { p.group_starts.get(group + k).copied().unwrap_or(n) } else { position + k }
            }
            Bound::UnboundedFollowing => n,
        };
        let end = match &self.end {
            Bound::Preceding(e) | Bound::Following(e) if self.range.is_some() => {
                self.range_edge(ctx, p, row, position, e, false)?
            }
            Bound::UnboundedFollowing => n as isize - 1,
            Bound::CurrentRow if rows_mode => position as isize,
            Bound::CurrentRow => peer_last as isize,
            Bound::Preceding(e) => {
                let k = offset(ctx, e, row)? as isize;
                if groups_mode {
                    let g = group as isize - k;
                    if g < 0 { -1 } else { p.group_end(g as usize) as isize }
                } else {
                    position as isize - k
                }
            }
            Bound::Following(e) => {
                let k = offset(ctx, e, row)?;
                if groups_mode {
                    let g = group + k;
                    if g >= p.group_starts.len() { n as isize - 1 } else { p.group_end(g) as isize }
                } else {
                    (position + k).min(n - 1) as isize
                }
            }
            Bound::UnboundedPreceding => -1,
        };
        Ok((start, end))
    }

    /// frame returns the positions of the rows in a row's frame.
    fn frame(&self, ctx: &mut Ctx<'_>, p: &Partition<'_>, position: usize) -> Result<Vec<usize>> {
        let n = p.part.len();
        let (start, end) = self.bounds(ctx, p, position)?;
        let (peer_first, peer_last) = p.peers(position);
        let mut positions: Vec<usize> =
            if end < start as isize { Vec::new() } else { (start..=(end as usize).min(n - 1)).collect() };
        if self.options & frame::EXCLUDE_CURRENT_ROW != 0 {
            positions.retain(|&i| i != position);
        } else if self.options & frame::EXCLUDE_GROUP != 0 {
            positions.retain(|&i| i < peer_first || i > peer_last);
        } else if self.options & frame::EXCLUDE_TIES != 0 {
            positions.retain(|&i| i == position || i < peer_first || i > peer_last);
        }
        Ok(positions)
    }

    /// value computes the call for the row at a position of a sorted partition, given the aggregate call of an
    /// aggregate window function.
    fn value(&self, ctx: &mut Ctx<'_>, p: &Partition<'_>, position: usize, call: Option<&AggCall>) -> Result<Value> {
        let n = p.part.len();
        let row = p.members[position];
        let rank = |position: usize| p.peers(position).0 + 1;
        Ok(match self.kind {
            WindowKind::RowNumber => Value::Int8(position as i64 + 1),
            WindowKind::Rank => Value::Int8(rank(position) as i64),
            WindowKind::DenseRank => Value::Int8(p.group_of[position] as i64 + 1),
            WindowKind::PercentRank => {
                Value::Float8(if n <= 1 { 0.0 } else { (rank(position) - 1) as f64 / (n - 1) as f64 })
            }
            WindowKind::CumeDist => Value::Float8((p.peers(position).1 + 1) as f64 / n as f64),
            WindowKind::Ntile => {
                let buckets = match self.args[0].eval(ctx, row)? {
                    Value::Int4(b) if b > 0 => b as usize,
                    Value::Null => return Ok(Value::Null),
                    _ => {
                        return Err(PgError::new(
                            code::INVALID_ARGUMENT_FOR_NTILE,
                            "argument of ntile must be greater than zero",
                        ));
                    }
                };
                let (size, extra) = (n / buckets, n % buckets);
                let big = (size + 1) * extra;
                let bucket =
                    if position < big { position / (size + 1) } else { extra + (position - big) / size.max(1) };
                Value::Int4(bucket as i32 + 1)
            }
            WindowKind::Lag | WindowKind::Lead => {
                let k = match self.args.get(1).map(|e| e.eval(ctx, row)).transpose()? {
                    Some(Value::Int4(k)) => k as i64,
                    Some(_) => return Ok(Value::Null),
                    None => 1,
                };
                let target = if self.kind == WindowKind::Lag { position as i64 - k } else { position as i64 + k };
                if target >= 0 && (target as usize) < n {
                    self.args[0].eval(ctx, p.members[target as usize])?
                } else {
                    match self.args.get(2) {
                        Some(default) => default.eval(ctx, row)?,
                        None => Value::Null,
                    }
                }
            }
            WindowKind::FirstValue | WindowKind::LastValue | WindowKind::NthValue => {
                let positions = self.frame(ctx, p, position)?;
                let chosen = match self.kind {
                    WindowKind::FirstValue => positions.first().copied(),
                    WindowKind::LastValue => positions.last().copied(),
                    _ => match self.args[1].eval(ctx, row)? {
                        Value::Int4(k) if k > 0 => positions.get(k as usize - 1).copied(),
                        Value::Null => None,
                        _ => {
                            return Err(PgError::new(
                                code::INVALID_ARGUMENT_FOR_NTH_VALUE,
                                "argument of nth_value must be greater than zero",
                            ));
                        }
                    },
                };
                match chosen {
                    Some(i) => self.args[0].eval(ctx, p.members[i])?,
                    None => Value::Null,
                }
            }
            WindowKind::Aggregate(_) => {
                let call = call.expect("an aggregate call");
                let mut accumulator = Accumulator::new(call);
                for i in self.frame(ctx, p, position)? {
                    accumulator.add(ctx, call, p.members[i])?;
                }
                accumulator.finish(ctx, call)?
            }
        })
    }
}

/// window_names returns the named windows of a SELECT's WINDOW clause.
pub fn window_names(clause: &[Node]) -> Vec<WindowDef> {
    clause
        .iter()
        .filter_map(|n| match n.node.as_ref() {
            Some(NodeEnum::WindowDef(w)) => Some((**w).clone()),
            _ => None,
        })
        .collect()
}
