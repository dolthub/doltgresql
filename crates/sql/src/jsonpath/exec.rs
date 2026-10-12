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

//! Evaluation of SQL/JSON paths over JSON items, following Postgres' jsonpath_exec.c: lax mode unwraps arrays and
//! ignores structural errors, strict mode reports them, and errors that are not thrown make predicates unknown.

use std::cmp::Ordering;
use std::collections::HashMap;

use regex::{Regex, RegexBuilder};

use super::{Accessor, ArithOp, CmpOp, JsonPath, Method, Node, TimeMethod};
use crate::datetime::{self as dt, Fields, POSTGRES_EPOCH_JDATE, USECS_PER_SEC};
use crate::error::{PgError, Result, code};
use crate::json::Json;
use crate::numeric::Numeric;
use crate::types::Value;

/// USECS_PER_DAY is the number of microseconds in a day.
const USECS_PER_DAY: i64 = 86_400 * USECS_PER_SEC;

/// Item is a SQL/JSON item: a JSON value, or a datetime that a path method or variable produced.
#[derive(Clone, Debug, PartialEq)]
pub enum Item {
    Json(Json),
    /// A date, time, timetz, timestamp, or timestamptz value, with the offset east of UTC that a timestamptz prints
    /// with when it has one.
    DateTime(Value, Option<i32>),
}

impl Item {
    /// type_name returns the name that the `.type()` method returns for the item.
    pub fn type_name(&self) -> &'static str {
        match self {
            Item::Json(Json::Null) => "null",
            Item::Json(Json::Bool(_)) => "boolean",
            Item::Json(Json::Number(_)) => "number",
            Item::Json(Json::String(_)) => "string",
            Item::Json(Json::Array(_)) => "array",
            Item::Json(Json::Object(_)) => "object",
            Item::DateTime(Value::Date(_), _) => "date",
            Item::DateTime(Value::Time(_), _) => "time without time zone",
            Item::DateTime(Value::TimeTz(..), _) => "time with time zone",
            Item::DateTime(Value::Timestamp(_), _) => "timestamp without time zone",
            Item::DateTime(..) => "timestamp with time zone",
        }
    }

    /// to_json returns the item as a JSON value, where a datetime becomes the ISO 8601 string that Postgres'
    /// JsonEncodeDateTime writes.
    pub fn to_json(&self) -> Json {
        match self {
            Item::Json(json) => json.clone(),
            Item::DateTime(value, offset) => Json::String(iso_text(value, *offset)),
        }
    }
}

/// iso_text writes a datetime value in ISO 8601, with a `T` between the date and time and a full zone offset.
pub fn iso_text(value: &Value, offset: Option<i32>) -> String {
    let date = |days: i64| {
        let (y, m, d) = dt::j2date(days + POSTGRES_EPOCH_JDATE);
        if y > 0 { format!("{y:04}-{m:02}-{d:02}") } else { format!("{:04}-{m:02}-{d:02} BC", 1 - y) }
    };
    let zone = |east: i32| {
        let sign = if east < 0 { '-' } else { '+' };
        let east = east.abs();
        let text = format!("{sign}{:02}:{:02}", east / 3600, east / 60 % 60);
        if east % 60 != 0 { format!("{text}:{:02}", east % 60) } else { text }
    };
    match value {
        Value::Date(days) => date(i64::from(*days)),
        Value::Time(micros) => dt::format_time(*micros),
        Value::TimeTz(micros, west) => format!("{}{}", dt::format_time(*micros), zone(-west)),
        Value::Timestamp(ts) => {
            let (days, time) = (ts.div_euclid(USECS_PER_DAY), ts.rem_euclid(USECS_PER_DAY));
            format!("{}T{}", date(days), dt::format_time(time))
        }
        Value::TimestampTz(ts) => {
            let east = offset.unwrap_or_else(|| dt::with_format(|f| f.zone.offset_at(*ts).0));
            let local = ts + i64::from(east) * USECS_PER_SEC;
            let (days, time) = (local.div_euclid(USECS_PER_DAY), local.rem_euclid(USECS_PER_DAY));
            format!("{}T{}{}", date(days), dt::format_time(time), zone(east))
        }
        other => other.output().unwrap_or_default(),
    }
}

/// Res is the outcome of evaluating part of a path: items found, none found, or an error that was not thrown.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum Res {
    Ok,
    NotFound,
    Error,
}

/// Tri is the three-valued result of a predicate.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
enum Tri {
    True,
    False,
    Unknown,
}

/// Found collects the items a path finds, or is None when only whether any exist matters.
type Found<'a> = Option<&'a mut Vec<Item>>;

/// PredicateFn evaluates a predicate over one item and, for a binary predicate, one item of its right operand.
type PredicateFn<'f, 'a> = dyn FnMut(&mut Exec<'a>, &Item, Option<&Item>) -> Result<Tri> + 'f;

/// Exec is the state of one path evaluation.
struct Exec<'a> {
    vars: &'a HashMap<String, Item>,
    root: Item,
    current: Item,
    lax: bool,
    ignore_structural: bool,
    throw: bool,
    use_tz: bool,
    innermost_size: Option<usize>,
}

/// regex compiles a `like_regex` pattern with its flags.
pub fn regex(pattern: &str, flags: &str) -> Result<Regex> {
    let quoted = flags.contains('q');
    let pattern = if quoted { regex::escape(pattern) } else { crate::functions::pattern::translate_regex(pattern) };
    let mut builder = RegexBuilder::new(&pattern);
    builder
        .case_insensitive(flags.contains('i'))
        .dot_matches_new_line(flags.contains('s'))
        .multi_line(flags.contains('m'));
    builder.build().map_err(|e| {
        PgError::new(
            code::INVALID_REGULAR_EXPRESSION,
            format!("invalid regular expression: {}", crate::functions::pattern::regex_error(&e)),
        )
    })
}

/// error returns a SQL/JSON error with the code.
fn error(code: &'static str, message: impl Into<String>) -> PgError {
    PgError::new(code, message)
}

/// method_name returns the name of an item method as errors spell it.
fn method_name(method: Method) -> &'static str {
    match method {
        Method::Abs => "abs",
        Method::Size => "size",
        Method::Type => "type",
        Method::Floor => "floor",
        Method::Double => "double",
        Method::Ceiling => "ceiling",
        Method::KeyValue => "keyvalue",
        Method::Bigint => "bigint",
        Method::Boolean => "boolean",
        Method::Date => "date",
        Method::Integer => "integer",
        Method::Number => "number",
        Method::String => "string",
    }
}

/// time_method_name returns the name of a time method as errors spell it.
fn time_method_name(method: TimeMethod) -> &'static str {
    match method {
        TimeMethod::Time => "time",
        TimeMethod::TimeTz => "time_tz",
        TimeMethod::Timestamp => "timestamp",
        TimeMethod::TimestampTz => "timestamp_tz",
    }
}

/// arith_name returns the symbol of an arithmetic operator.
fn arith_name(op: ArithOp) -> &'static str {
    match op {
        ArithOp::Add => "+",
        ArithOp::Sub => "-",
        ArithOp::Mul => "*",
        ArithOp::Div => "/",
        ArithOp::Mod => "%",
    }
}

/// float_numeric converts a float to a numeric as Postgres' float8_numeric does, through 15 significant digits.
fn float_numeric(value: f64) -> Numeric {
    let text = format!("{value:.14e}");
    let (mantissa, exponent) = text.split_once('e').unwrap_or((&text, "0"));
    let mantissa = if mantissa.contains('.') { mantissa.trim_end_matches('0').trim_end_matches('.') } else { mantissa };
    Numeric::parse(&format!("{mantissa}e{exponent}")).unwrap_or(Numeric::NaN)
}

/// integer_part returns a numeric rounded toward negative infinity, or toward positive infinity when asked.
fn integer_part(n: &Numeric, ceiling: bool) -> Numeric {
    let whole = match n {
        Numeric::Finite { negative, coefficient, scale } => {
            let divisor = num_bigint::BigUint::from(10u32).pow(*scale);
            let quotient = coefficient / divisor;
            Numeric::Finite {
                negative: *negative && quotient != num_bigint::BigUint::ZERO,
                coefficient: quotient,
                scale: 0,
            }
        }
        other => return other.clone(),
    };
    match (ceiling, whole.cmp_numeric(n)) {
        (true, Ordering::Less) => whole.add(&Numeric::from_i64(1)),
        (false, Ordering::Greater) => whole.sub(&Numeric::from_i64(1)),
        _ => whole,
    }
}

impl<'a> Exec<'a> {
    /// soft returns an error as Postgres' RETURN_ERROR does: thrown when errors are thrown, and otherwise an error
    /// outcome.
    fn soft(&self, err: PgError) -> Result<Res> {
        if self.throw { Err(err) } else { Ok(Res::Error) }
    }

    /// item evaluates a node over an item, passing each result through the accessors that follow it.
    fn item(&mut self, node: &Node, rest: &[Accessor], jb: &Item, found: Found<'_>) -> Result<Res> {
        match node {
            Node::Null => self.next(rest, Item::Json(Json::Null), found),
            Node::Bool(b) => self.next(rest, Item::Json(Json::Bool(*b)), found),
            Node::Number(n) => self.next(rest, Item::Json(Json::Number(n.clone())), found),
            Node::String(s) => self.next(rest, Item::Json(Json::String(s.clone())), found),
            Node::Variable(name) => {
                let value = self.vars.get(name).cloned().ok_or_else(|| {
                    error(code::UNDEFINED_OBJECT, format!("could not find jsonpath variable \"{name}\""))
                })?;
                self.next(rest, value, found)
            }
            Node::Root => {
                let root = self.root.clone();
                self.next(rest, root, found)
            }
            Node::Current => {
                let current = self.current.clone();
                self.next(rest, current, found)
            }
            Node::Last => {
                let size = self
                    .innermost_size
                    .ok_or_else(|| PgError::internal("evaluating jsonpath LAST outside of array subscript"))?;
                self.next(rest, Item::Json(Json::Number(Numeric::from_i64(size as i64 - 1))), found)
            }
            Node::Chain(primary, accessors) => {
                if rest.is_empty() {
                    self.item(primary, accessors, jb, found)
                } else {
                    let combined: Vec<Accessor> = accessors.iter().chain(rest).cloned().collect();
                    self.item(primary, &combined, jb, found)
                }
            }
            Node::Unary(negate, operand) => self.unary(*negate, operand, rest, jb, found),
            Node::Arith(op, left, right) => self.arith(*op, left, right, rest, jb, found),
            _ => {
                let value = match self.bool_item(node, jb)? {
                    Tri::True => Json::Bool(true),
                    Tri::False => Json::Bool(false),
                    Tri::Unknown => Json::Null,
                };
                self.next(rest, Item::Json(value), found)
            }
        }
    }

    /// next passes an item through the accessors that remain, collecting it when none do.
    fn next(&mut self, rest: &[Accessor], value: Item, found: Found<'_>) -> Result<Res> {
        match rest.split_first() {
            None => {
                if let Some(found) = found {
                    found.push(value);
                }
                Ok(Res::Ok)
            }
            Some((accessor, rest)) => {
                let unwrap = self.lax;
                self.step(accessor, rest, &value, found, unwrap)
            }
        }
    }

    /// elements returns the elements of an array, or None for any other item.
    fn elements(jb: &Item) -> Option<&Vec<Json>> {
        match jb {
            Item::Json(Json::Array(elements)) => Some(elements),
            _ => None,
        }
    }

    /// unwrap_array applies the accessor, or collects the elements when there is none, to each element of an array,
    /// as Postgres' executeItemUnwrapTargetArray does.
    fn unwrap_array(
        &mut self,
        accessor: Option<(&Accessor, &[Accessor])>,
        jb: &Item,
        mut found: Found<'_>,
        unwrap_elements: bool,
    ) -> Result<Res> {
        let elements = Self::elements(jb).cloned().unwrap_or_default();
        let mut res = Res::NotFound;
        for element in elements {
            let element = Item::Json(element);
            match accessor {
                Some((accessor, rest)) => {
                    let r = self.step(accessor, rest, &element, found.as_deref_mut(), unwrap_elements)?;
                    if r == Res::Error {
                        return Ok(r);
                    }
                    if r == Res::Ok {
                        if found.is_none() {
                            return Ok(Res::Ok);
                        }
                        res = Res::Ok;
                    }
                }
                None => match found.as_deref_mut() {
                    Some(found) => {
                        found.push(element);
                        res = Res::Ok;
                    }
                    None => return Ok(Res::Ok),
                },
            }
        }
        Ok(res)
    }

    /// any_item visits the values of an array or object down to the levels, applying the accessors that follow to
    /// those at the requested levels, as Postgres' executeAnyItem does.
    #[allow(clippy::too_many_arguments)]
    fn any_item(
        &mut self,
        rest: Option<&[Accessor]>,
        jb: &Json,
        mut found: Found<'_>,
        level: u32,
        first: u32,
        last: u32,
        ignore_structural: bool,
        unwrap_next: bool,
    ) -> Result<Res> {
        let mut res = Res::NotFound;
        if level > last {
            return Ok(res);
        }
        let values: Vec<&Json> = match jb {
            Json::Array(elements) => elements.iter().collect(),
            Json::Object(pairs) => pairs.iter().map(|(_, v)| v).collect(),
            _ => return Ok(res),
        };
        for value in values {
            let container = matches!(value, Json::Array(_) | Json::Object(_));
            if level >= first || (first == super::ANY_LAST && last == super::ANY_LAST && !container) {
                match rest {
                    Some(rest) => {
                        let saved = self.ignore_structural;
                        if ignore_structural {
                            self.ignore_structural = true;
                        }
                        let item = Item::Json(value.clone());
                        res = match rest.split_first() {
                            Some((accessor, rest)) => {
                                self.step(accessor, rest, &item, found.as_deref_mut(), unwrap_next)?
                            }
                            None => self.next(&[], item, found.as_deref_mut())?,
                        };
                        self.ignore_structural = saved;
                        if res == Res::Error || (res == Res::Ok && found.is_none()) {
                            break;
                        }
                    }
                    None => match found.as_deref_mut() {
                        Some(found) => {
                            found.push(Item::Json(value.clone()));
                            res = Res::Ok;
                        }
                        None => return Ok(Res::Ok),
                    },
                }
            }
            if level < last && container {
                res = self.any_item(
                    rest,
                    value,
                    found.as_deref_mut(),
                    level + 1,
                    first,
                    last,
                    ignore_structural,
                    unwrap_next,
                )?;
                if res == Res::Error || (res == Res::Ok && found.is_none()) {
                    break;
                }
            }
        }
        Ok(res)
    }

    /// step applies one accessor to an item, as Postgres' executeItemOptUnwrapTarget does for that item type.
    fn step(
        &mut self,
        accessor: &Accessor,
        rest: &[Accessor],
        jb: &Item,
        mut found: Found<'_>,
        unwrap: bool,
    ) -> Result<Res> {
        let is_array = Self::elements(jb).is_some();
        match accessor {
            Accessor::Key(key) => match jb {
                Item::Json(Json::Object(pairs)) => match pairs.iter().find(|(k, _)| k == key) {
                    Some((_, value)) => self.next(rest, Item::Json(value.clone()), found),
                    None if !self.ignore_structural => self.soft(error(
                        code::SQL_JSON_MEMBER_NOT_FOUND,
                        format!("JSON object does not contain key \"{key}\""),
                    )),
                    None => Ok(Res::NotFound),
                },
                _ if unwrap && is_array => self.unwrap_array(Some((accessor, rest)), jb, found, false),
                _ if !self.ignore_structural => self.soft(error(
                    code::SQL_JSON_MEMBER_NOT_FOUND,
                    "jsonpath member accessor can only be applied to an object",
                )),
                _ => Ok(Res::NotFound),
            },
            Accessor::AnyArray => {
                if is_array {
                    let lax = self.lax;
                    if rest.is_empty() {
                        self.unwrap_array(None, jb, found, lax)
                    } else {
                        self.unwrap_array(Some((&rest[0], &rest[1..])), jb, found, lax)
                    }
                } else if self.lax {
                    self.next(rest, jb.clone(), found)
                } else if !self.ignore_structural {
                    self.soft(error(
                        code::SQL_JSON_ARRAY_NOT_FOUND,
                        "jsonpath wildcard array accessor can only be applied to an array",
                    ))
                } else {
                    Ok(Res::NotFound)
                }
            }
            Accessor::AnyKey => match jb {
                Item::Json(object @ Json::Object(_)) => {
                    let lax = self.lax;
                    self.any_item(Some(rest), object, found, 1, 1, 1, false, lax)
                }
                _ if unwrap && is_array => self.unwrap_array(Some((accessor, rest)), jb, found, false),
                _ if !self.ignore_structural => self.soft(error(
                    code::SQL_JSON_OBJECT_NOT_FOUND,
                    "jsonpath wildcard member accessor can only be applied to an object",
                )),
                _ => Ok(Res::NotFound),
            },
            Accessor::Index(subscripts) => {
                if !is_array && !self.lax {
                    if !self.ignore_structural {
                        return self.soft(error(
                            code::SQL_JSON_ARRAY_NOT_FOUND,
                            "jsonpath array accessor can only be applied to an array",
                        ));
                    }
                    return Ok(Res::NotFound);
                }
                let elements = Self::elements(jb).cloned();
                let size = elements.as_ref().map_or(1, Vec::len);
                let saved = self.innermost_size.replace(size);
                let result = self.subscripts(subscripts, rest, jb, elements.as_deref(), size, found.as_deref_mut());
                self.innermost_size = saved;
                result
            }
            Accessor::Any(first, last) => {
                if *first == 0 {
                    let saved = self.ignore_structural;
                    self.ignore_structural = true;
                    let res = self.next(rest, jb.clone(), found.as_deref_mut());
                    self.ignore_structural = saved;
                    let res = res?;
                    if res == Res::Ok && found.is_none() {
                        return Ok(res);
                    }
                }
                match jb {
                    Item::Json(container @ (Json::Array(_) | Json::Object(_))) => {
                        let lax = self.lax;
                        self.any_item(Some(rest), container, found, 1, *first, *last, true, lax)
                    }
                    _ => Ok(Res::NotFound),
                }
            }
            Accessor::Filter(predicate) => {
                if unwrap && is_array {
                    return self.unwrap_array(Some((accessor, rest)), jb, found, false);
                }
                if self.nested_bool(predicate, jb)? != Tri::True {
                    return Ok(Res::NotFound);
                }
                self.next(rest, jb.clone(), found)
            }
            Accessor::Method(Method::Type) => self.next(rest, Item::Json(Json::String(jb.type_name().into())), found),
            Accessor::Method(Method::Size) => {
                let size = match Self::elements(jb) {
                    Some(elements) => elements.len(),
                    None if !self.lax => {
                        if !self.ignore_structural {
                            return self.soft(error(
                                code::SQL_JSON_ARRAY_NOT_FOUND,
                                "jsonpath item method .size() can only be applied to an array",
                            ));
                        }
                        return Ok(Res::NotFound);
                    }
                    None => 1,
                };
                self.next(rest, Item::Json(Json::Number(Numeric::from_i64(size as i64))), found)
            }
            Accessor::Method(method) => {
                if unwrap && is_array {
                    return self.unwrap_array(Some((accessor, rest)), jb, found, false);
                }
                self.method(*method, rest, jb, found)
            }
            Accessor::Datetime(_) | Accessor::Time(..) => {
                if unwrap && is_array {
                    return self.unwrap_array(Some((accessor, rest)), jb, found, false);
                }
                self.datetime_method(accessor, rest, jb, found)
            }
            Accessor::Decimal(precision, scale) => {
                if unwrap && is_array {
                    return self.unwrap_array(Some((accessor, rest)), jb, found, false);
                }
                self.decimal_method(precision.as_deref(), scale.as_deref(), rest, jb, found)
            }
        }
    }

    /// subscripts applies an array accessor's subscripts to an array, or to a non-array treated as an array of
    /// itself.
    fn subscripts(
        &mut self,
        subscripts: &[(Node, Option<Node>)],
        rest: &[Accessor],
        jb: &Item,
        elements: Option<&[Json]>,
        size: usize,
        mut found: Found<'_>,
    ) -> Result<Res> {
        let mut res = Res::NotFound;
        for (from, to) in subscripts {
            let from = match self.array_index(from, jb)? {
                Ok(index) => index,
                Err(res) => return Ok(res),
            };
            let to = match to {
                Some(to) => match self.array_index(to, jb)? {
                    Ok(index) => index,
                    Err(res) => return Ok(res),
                },
                None => from,
            };
            if !self.ignore_structural && (from < 0 || from > to || to >= size as i64) {
                return self.soft(error(code::INVALID_SQL_JSON_SUBSCRIPT, "jsonpath array subscript is out of bounds"));
            }
            let (from, to) = (from.max(0), to.min(size as i64 - 1));
            let mut index = from;
            while index <= to {
                let value = match elements {
                    Some(elements) => Item::Json(elements[index as usize].clone()),
                    None => jb.clone(),
                };
                let r = self.next(rest, value, found.as_deref_mut())?;
                if r == Res::Error {
                    return Ok(r);
                }
                if r == Res::Ok {
                    if found.is_none() {
                        return Ok(Res::Ok);
                    }
                    res = Res::Ok;
                }
                index += 1;
            }
        }
        Ok(res)
    }

    /// array_index evaluates an array subscript to an integer, or returns the outcome of an error that was not
    /// thrown.
    fn array_index(&mut self, node: &Node, jb: &Item) -> Result<std::result::Result<i64, Res>> {
        let mut values = Vec::new();
        let res = self.item(node, &[], jb, Some(&mut values))?;
        if res == Res::Error {
            return Ok(Err(res));
        }
        let number = match values.as_slice() {
            [Item::Json(Json::Number(n))] => n.clone(),
            _ => {
                return self
                    .soft(error(
                        code::INVALID_SQL_JSON_SUBSCRIPT,
                        "jsonpath array subscript is not a single numeric value",
                    ))
                    .map(Err);
            }
        };
        match integer_part(&number, number.is_negative()).to_i64().and_then(|i| i32::try_from(i).ok()) {
            Some(index) => Ok(Ok(i64::from(index))),
            None => self
                .soft(error(code::INVALID_SQL_JSON_SUBSCRIPT, "jsonpath array subscript is out of integer range"))
                .map(Err),
        }
    }

    /// unwrapped_results evaluates a node, unwrapping the arrays among its results one level in lax mode when asked,
    /// as Postgres' executeItemOptUnwrapResult does.
    fn unwrapped_results(&mut self, node: &Node, jb: &Item, unwrap: bool, out: &mut Vec<Item>) -> Result<Res> {
        let lax = self.lax;
        if !(unwrap && lax) {
            return self.item(node, &[], jb, Some(out));
        }
        let mut seq = Vec::new();
        let res = self.item(node, &[], jb, Some(&mut seq))?;
        if res == Res::Error {
            return Ok(res);
        }
        for value in seq {
            match value {
                Item::Json(Json::Array(elements)) => out.extend(elements.into_iter().map(Item::Json)),
                other => out.push(other),
            }
        }
        Ok(res)
    }

    /// unwrapped_results_quiet evaluates as unwrapped_results does without throwing errors.
    fn unwrapped_results_quiet(&mut self, node: &Node, jb: &Item, unwrap: bool, out: &mut Vec<Item>) -> Result<Res> {
        let saved = std::mem::replace(&mut self.throw, false);
        let res = self.unwrapped_results(node, jb, unwrap, out);
        self.throw = saved;
        res
    }

    /// unary applies a unary plus or minus to each numeric result of its operand.
    fn unary(
        &mut self,
        negate: bool,
        operand: &Node,
        rest: &[Accessor],
        jb: &Item,
        mut found: Found<'_>,
    ) -> Result<Res> {
        let mut seq = Vec::new();
        let res = self.unwrapped_results(operand, jb, true, &mut seq)?;
        if res == Res::Error {
            return Ok(res);
        }
        let mut res = Res::NotFound;
        let has_next = !rest.is_empty();
        for value in seq {
            let number = match value {
                Item::Json(Json::Number(n)) => n,
                _ => {
                    if found.is_none() && !has_next {
                        continue;
                    }
                    return self.soft(error(
                        code::SQL_JSON_NUMBER_NOT_FOUND,
                        format!(
                            "operand of unary jsonpath operator {} is not a numeric value",
                            if negate { "-" } else { "+" }
                        ),
                    ));
                }
            };
            if found.is_none() && !has_next {
                return Ok(Res::Ok);
            }
            let number = if negate { number.negate() } else { number };
            let r = self.next(rest, Item::Json(Json::Number(number)), found.as_deref_mut())?;
            if r == Res::Error {
                return Ok(r);
            }
            if r == Res::Ok {
                if found.is_none() {
                    return Ok(Res::Ok);
                }
                res = Res::Ok;
            }
        }
        Ok(res)
    }

    /// arith applies a binary arithmetic operator to the single numeric results of its operands.
    fn arith(
        &mut self,
        op: ArithOp,
        left: &Node,
        right: &Node,
        rest: &[Accessor],
        jb: &Item,
        found: Found<'_>,
    ) -> Result<Res> {
        let (mut lseq, mut rseq) = (Vec::new(), Vec::new());
        let res = self.unwrapped_results(left, jb, true, &mut lseq)?;
        if res == Res::Error {
            return Ok(res);
        }
        let res = self.unwrapped_results(right, jb, true, &mut rseq)?;
        if res == Res::Error {
            return Ok(res);
        }
        let single = |seq: &[Item]| match seq {
            [Item::Json(Json::Number(n))] => Some(n.clone()),
            _ => None,
        };
        let Some(l) = single(&lseq) else {
            return self.soft(error(
                code::SINGLETON_SQL_JSON_ITEM_REQUIRED,
                format!("left operand of jsonpath operator {} is not a single numeric value", arith_name(op)),
            ));
        };
        let Some(r) = single(&rseq) else {
            return self.soft(error(
                code::SINGLETON_SQL_JSON_ITEM_REQUIRED,
                format!("right operand of jsonpath operator {} is not a single numeric value", arith_name(op)),
            ));
        };
        let result = match op {
            ArithOp::Add => Ok(l.add(&r)),
            ArithOp::Sub => Ok(l.sub(&r)),
            ArithOp::Mul => Ok(l.mul(&r)),
            ArithOp::Div => l.div(&r),
            ArithOp::Mod => l.rem(&r),
        };
        let value = match result {
            Ok(value) => value,
            Err(err) => return self.soft(err),
        };
        if rest.is_empty() && found.is_none() {
            return Ok(Res::Ok);
        }
        self.next(rest, Item::Json(Json::Number(value)), found)
    }

    /// method applies an item method other than `.type()` and `.size()`.
    fn method(&mut self, method: Method, rest: &[Accessor], jb: &Item, mut found: Found<'_>) -> Result<Res> {
        let name = method_name(method);
        let invalid = |value: &str, type_name: &str| {
            error(
                code::NON_NUMERIC_SQL_JSON_ITEM,
                format!("argument \"{value}\" of jsonpath item method .{name}() is invalid for type {type_name}"),
            )
        };
        let value = match (method, jb) {
            (Method::Abs | Method::Floor | Method::Ceiling, Item::Json(Json::Number(n))) => {
                let n = match method {
                    Method::Abs if n.is_negative() => n.negate(),
                    Method::Abs => n.clone(),
                    Method::Floor => integer_part(n, false),
                    _ => integer_part(n, true),
                };
                Json::Number(n)
            }
            (Method::Abs | Method::Floor | Method::Ceiling, _) => {
                return self.soft(error(
                    code::NON_NUMERIC_SQL_JSON_ITEM,
                    format!("jsonpath item method .{name}() can only be applied to a numeric value"),
                ));
            }
            (Method::Double, Item::Json(Json::Number(n))) => {
                let f = n.to_f64();
                if !f.is_finite() {
                    return self.soft(error(
                        code::NON_NUMERIC_SQL_JSON_ITEM,
                        format!("NaN or Infinity is not allowed for jsonpath item method .{name}()"),
                    ));
                }
                Json::Number(float_numeric(f))
            }
            (Method::Double, Item::Json(Json::String(s))) => {
                let Ok(Value::Float8(f)) = crate::cast::input(s, crate::oid::FLOAT8) else {
                    return self.soft(invalid(s, "double precision"));
                };
                if !f.is_finite() {
                    return self.soft(error(
                        code::NON_NUMERIC_SQL_JSON_ITEM,
                        format!("NaN or Infinity is not allowed for jsonpath item method .{name}()"),
                    ));
                }
                Json::Number(float_numeric(f))
            }
            (Method::Number, Item::Json(Json::Number(n))) => Json::Number(n.clone()),
            (Method::Number, Item::Json(Json::String(s))) => match Numeric::parse(s) {
                Ok(n @ Numeric::Finite { .. }) => Json::Number(n),
                Ok(_) => {
                    return self.soft(error(
                        code::NON_NUMERIC_SQL_JSON_ITEM,
                        format!("NaN or Infinity is not allowed for jsonpath item method .{name}()"),
                    ));
                }
                Err(_) => return self.soft(invalid(s, "numeric")),
            },
            (Method::Bigint | Method::Integer, Item::Json(Json::Number(n))) => {
                let rounded = n.with_scale(0).to_i64();
                let fits = match method {
                    Method::Integer => rounded.filter(|i| i32::try_from(*i).is_ok()),
                    _ => rounded,
                };
                match fits {
                    Some(i) => Json::Number(Numeric::from_i64(i)),
                    None => {
                        return self.soft(invalid(
                            &n.to_string(),
                            if method == Method::Integer { "integer" } else { "bigint" },
                        ));
                    }
                }
            }
            (Method::Bigint | Method::Integer, Item::Json(Json::String(s))) => {
                let target = if method == Method::Integer { crate::oid::INT4 } else { crate::oid::INT8 };
                match crate::cast::input(s, target) {
                    Ok(Value::Int4(i)) => Json::Number(Numeric::from_i64(i64::from(i))),
                    Ok(Value::Int8(i)) => Json::Number(Numeric::from_i64(i)),
                    _ => return self.soft(invalid(s, if method == Method::Integer { "integer" } else { "bigint" })),
                }
            }
            (Method::Boolean, Item::Json(Json::Bool(b))) => Json::Bool(*b),
            (Method::Boolean, Item::Json(Json::Number(n))) => {
                match n.with_scale(0).to_i64().filter(|_| n.scale() == 0 || n.with_scale(0).cmp_numeric(n).is_eq()) {
                    Some(i) if i32::try_from(i).is_ok() => Json::Bool(i != 0),
                    _ => return self.soft(invalid(&n.to_string(), "boolean")),
                }
            }
            (Method::Boolean, Item::Json(Json::String(s))) => match crate::cast::input(s, crate::oid::BOOL) {
                Ok(Value::Bool(b)) => Json::Bool(b),
                _ => return self.soft(invalid(s, "boolean")),
            },
            (Method::String, Item::Json(Json::String(s))) => Json::String(s.clone()),
            (Method::String, Item::Json(Json::Number(n))) => Json::String(n.to_string()),
            (Method::String, Item::Json(Json::Bool(b))) => Json::String(b.to_string()),
            (Method::String, Item::DateTime(value, _)) => Json::String(value.output().unwrap_or_default()),
            (Method::String, _) => {
                return self.soft(error(
                    code::NON_NUMERIC_SQL_JSON_ITEM,
                    format!(
                        "jsonpath item method .{name}() can only be applied to a boolean, string, numeric, or datetime value"
                    ),
                ));
            }
            (Method::Boolean, _) => {
                return self.soft(error(
                    code::NON_NUMERIC_SQL_JSON_ITEM,
                    format!(
                        "jsonpath item method .{name}() can only be applied to a boolean, string, or numeric value"
                    ),
                ));
            }
            (Method::KeyValue, Item::Json(Json::Object(pairs))) => {
                let mut res = Res::NotFound;
                for (key, value) in pairs.clone() {
                    let object = Json::Object(vec![
                        ("id".into(), Json::Number(Numeric::from_i64(0))),
                        ("key".into(), Json::String(key)),
                        ("value".into(), value),
                    ]);
                    let r = self.next(rest, Item::Json(object), found.as_deref_mut())?;
                    if r == Res::Error {
                        return Ok(r);
                    }
                    if r == Res::Ok {
                        if found.is_none() {
                            return Ok(Res::Ok);
                        }
                        res = Res::Ok;
                    }
                }
                return Ok(res);
            }
            (Method::KeyValue, _) => {
                return self.soft(error(
                    code::SQL_JSON_OBJECT_NOT_FOUND,
                    format!("jsonpath item method .{name}() can only be applied to an object"),
                ));
            }
            (Method::Date, _) => return self.datetime_method(&Accessor::Method(Method::Date), rest, jb, found),
            _ => {
                return self.soft(error(
                    code::NON_NUMERIC_SQL_JSON_ITEM,
                    format!("jsonpath item method .{name}() can only be applied to a string or numeric value"),
                ));
            }
        };
        self.next(rest, Item::Json(value), found)
    }

    /// decimal_method applies the `.decimal()` method with its optional precision and scale.
    fn decimal_method(
        &mut self,
        precision: Option<&Node>,
        scale: Option<&Node>,
        rest: &[Accessor],
        jb: &Item,
        found: Found<'_>,
    ) -> Result<Res> {
        let number = match jb {
            Item::Json(Json::Number(n)) => n.clone(),
            Item::Json(Json::String(s)) => match Numeric::parse(s) {
                Ok(n @ Numeric::Finite { .. }) => n,
                _ => {
                    return self.soft(error(
                        code::NON_NUMERIC_SQL_JSON_ITEM,
                        format!("argument \"{s}\" of jsonpath item method .decimal() is invalid for type numeric"),
                    ));
                }
            },
            _ => {
                return self.soft(error(
                    code::NON_NUMERIC_SQL_JSON_ITEM,
                    "jsonpath item method .decimal() can only be applied to a string or numeric value",
                ));
            }
        };
        let constant = |node: Option<&Node>| match node {
            Some(Node::Number(n)) => n.to_i64(),
            Some(Node::Unary(true, inner)) => match inner.as_ref() {
                Node::Number(n) => n.to_i64().map(|i| -i),
                _ => None,
            },
            _ => None,
        };
        let value = match constant(precision) {
            Some(p) => {
                let s = constant(scale).unwrap_or(0);
                let typmod = ((p << 16) | (s & 0x7ff)) as i32 + 4;
                match number.apply_typmod(typmod) {
                    Ok(n) => n,
                    Err(_) => {
                        return self.soft(error(
                            code::NON_NUMERIC_SQL_JSON_ITEM,
                            format!(
                                "argument \"{number}\" of jsonpath item method .decimal() is invalid for type numeric"
                            ),
                        ));
                    }
                }
            }
            None => number,
        };
        self.next(rest, Item::Json(Json::Number(value)), found)
    }

    /// datetime_method applies `.datetime()`, `.date()`, or a time method to a string, as Postgres'
    /// executeDateTimeMethod does.
    fn datetime_method(&mut self, accessor: &Accessor, rest: &[Accessor], jb: &Item, found: Found<'_>) -> Result<Res> {
        let name = match accessor {
            Accessor::Datetime(_) => "datetime",
            Accessor::Time(kind, _) => time_method_name(*kind),
            _ => "date",
        };
        let Item::Json(Json::String(text)) = jb else {
            return self.soft(error(
                code::INVALID_ARGUMENT_FOR_SQL_JSON_DATETIME_FUNCTION,
                format!("jsonpath item method .{name}() can only be applied to a string"),
            ));
        };
        let parsed = match accessor {
            Accessor::Datetime(Some(template)) => match formatted_datetime(text, template) {
                Ok(parsed) => parsed,
                Err(err) => return self.soft(err),
            },
            _ => match iso_datetime(text) {
                Some(parsed) => parsed,
                None => {
                    return self.soft(PgError {
                        hint: Some("Use a datetime template argument to specify the input data format.".into()),
                        ..error(
                            code::INVALID_ARGUMENT_FOR_SQL_JSON_DATETIME_FUNCTION,
                            format!("{name} format is not recognized: \"{text}\""),
                        )
                    });
                }
            },
        };
        let not_recognized = || {
            error(
                code::INVALID_ARGUMENT_FOR_SQL_JSON_DATETIME_FUNCTION,
                format!("{name} format is not recognized: \"{text}\""),
            )
        };
        let (value, offset) = parsed;
        let converted = match accessor {
            Accessor::Datetime(_) => (value, offset),
            Accessor::Method(_) => match value {
                Value::Date(d) => (Value::Date(d), None),
                Value::Timestamp(ts) => (Value::Date(ts.div_euclid(USECS_PER_DAY) as i32), None),
                Value::TimestampTz(ts) => {
                    self.require_tz("timestamptz", "date")?;
                    let local = ts + i64::from(dt::with_format(|f| f.zone.offset_at(ts).0)) * USECS_PER_SEC;
                    (Value::Date(local.div_euclid(USECS_PER_DAY) as i32), None)
                }
                _ => return self.soft(not_recognized()),
            },
            Accessor::Time(kind, precision) => {
                let value = match (kind, value) {
                    (TimeMethod::Time, Value::Time(t)) => Value::Time(t),
                    (TimeMethod::Time, Value::TimeTz(t, _)) => {
                        self.require_tz("timetz", "time")?;
                        Value::Time(t)
                    }
                    (TimeMethod::Time, Value::Timestamp(ts)) => Value::Time(ts.rem_euclid(USECS_PER_DAY)),
                    (TimeMethod::Time, Value::TimestampTz(ts)) => {
                        self.require_tz("timestamptz", "time")?;
                        let local = ts + i64::from(dt::with_format(|f| f.zone.offset_at(ts).0)) * USECS_PER_SEC;
                        Value::Time(local.rem_euclid(USECS_PER_DAY))
                    }
                    (TimeMethod::TimeTz, Value::TimeTz(t, w)) => Value::TimeTz(t, w),
                    (TimeMethod::TimeTz, Value::Time(t)) => {
                        self.require_tz("time", "timetz")?;
                        Value::TimeTz(t, -dt::with_format(|f| f.zone.offset_at(dt::clock()).0))
                    }
                    (TimeMethod::TimeTz, Value::TimestampTz(ts)) => {
                        let east = offset.unwrap_or_else(|| dt::with_format(|f| f.zone.offset_at(ts).0));
                        let local = ts + i64::from(east) * USECS_PER_SEC;
                        Value::TimeTz(local.rem_euclid(USECS_PER_DAY), -east)
                    }
                    (TimeMethod::Timestamp, Value::Timestamp(ts)) => Value::Timestamp(ts),
                    (TimeMethod::Timestamp, Value::Date(d)) => Value::Timestamp(i64::from(d) * USECS_PER_DAY),
                    (TimeMethod::Timestamp, Value::TimestampTz(ts)) => {
                        self.require_tz("timestamptz", "timestamp")?;
                        Value::Timestamp(ts + i64::from(dt::with_format(|f| f.zone.offset_at(ts).0)) * USECS_PER_SEC)
                    }
                    (TimeMethod::TimestampTz, Value::TimestampTz(ts)) => Value::TimestampTz(ts),
                    (TimeMethod::TimestampTz, Value::Date(d)) => {
                        self.require_tz("date", "timestamptz")?;
                        let local = i64::from(d) * USECS_PER_DAY;
                        Value::TimestampTz(
                            local - i64::from(dt::with_format(|f| f.zone.offset_for_local(local))) * USECS_PER_SEC,
                        )
                    }
                    (TimeMethod::TimestampTz, Value::Timestamp(local)) => {
                        self.require_tz("timestamp", "timestamptz")?;
                        Value::TimestampTz(
                            local - i64::from(dt::with_format(|f| f.zone.offset_for_local(local))) * USECS_PER_SEC,
                        )
                    }
                    _ => return self.soft(not_recognized()),
                };
                let value = match precision {
                    Some(p) => round_precision(value, *p),
                    None => value,
                };
                let offset = if matches!(value, Value::TimestampTz(_)) { offset } else { None };
                (value, offset)
            }
            _ => (value, offset),
        };
        self.next(rest, Item::DateTime(converted.0, converted.1), found)
    }

    /// require_tz fails as Postgres does when a conversion between datetime types needs the time zone that only the
    /// `_tz` functions allow.
    fn require_tz(&self, from: &str, to: &str) -> Result<()> {
        if self.use_tz {
            return Ok(());
        }
        Err(PgError {
            hint: Some("Use *_tz() function for time zone support.".into()),
            ..error(
                code::FEATURE_NOT_SUPPORTED,
                format!("cannot convert value from {from} to {to} without time zone usage"),
            )
        })
    }

    /// nested_bool evaluates a filter's predicate with the item as `@`.
    fn nested_bool(&mut self, node: &Node, jb: &Item) -> Result<Tri> {
        let saved = std::mem::replace(&mut self.current, jb.clone());
        let result = self.bool_item(node, jb);
        self.current = saved;
        result
    }

    /// bool_item evaluates a predicate, as Postgres' executeBoolItem does.
    fn bool_item(&mut self, node: &Node, jb: &Item) -> Result<Tri> {
        match node {
            Node::And(left, right) => {
                let l = self.bool_item(left, jb)?;
                if l == Tri::False {
                    return Ok(Tri::False);
                }
                let r = self.bool_item(right, jb)?;
                Ok(if r == Tri::True { l } else { r })
            }
            Node::Or(left, right) => {
                let l = self.bool_item(left, jb)?;
                if l == Tri::True {
                    return Ok(Tri::True);
                }
                let r = self.bool_item(right, jb)?;
                Ok(if r == Tri::False { l } else { r })
            }
            Node::Not(inner) => Ok(match self.bool_item(inner, jb)? {
                Tri::True => Tri::False,
                Tri::False => Tri::True,
                Tri::Unknown => Tri::Unknown,
            }),
            Node::IsUnknown(inner) => {
                Ok(if self.bool_item(inner, jb)? == Tri::Unknown { Tri::True } else { Tri::False })
            }
            Node::Compare(op, left, right) => {
                let op = *op;
                self.predicate(left, Some(right), jb, true, &mut |exec, l, r| exec.compare(op, l, r.unwrap_or(l)))
            }
            Node::StartsWith(whole, initial) => {
                self.predicate(whole, Some(initial), jb, false, &mut |_, w, i| {
                    Ok(match (w, i) {
                        (Item::Json(Json::String(w)), Some(Item::Json(Json::String(i)))) => {
                            if w.starts_with(i.as_str()) { Tri::True } else { Tri::False }
                        }
                        _ => Tri::Unknown,
                    })
                })
            }
            Node::LikeRegex(operand, pattern, flags) => {
                let regex = regex(pattern, flags)?;
                self.predicate(operand, None, jb, false, &mut |_, s, _| {
                    Ok(match s {
                        Item::Json(Json::String(s)) => {
                            if regex.is_match(s) {
                                Tri::True
                            } else {
                                Tri::False
                            }
                        }
                        _ => Tri::Unknown,
                    })
                })
            }
            Node::Exists(inner) => {
                if !self.lax {
                    let mut values = Vec::new();
                    let res = self.unwrapped_results_quiet(inner, jb, false, &mut values)?;
                    if res == Res::Error {
                        return Ok(Tri::Unknown);
                    }
                    return Ok(if values.is_empty() { Tri::False } else { Tri::True });
                }
                let saved = std::mem::replace(&mut self.throw, false);
                let res = self.item(inner, &[], jb, None);
                self.throw = saved;
                Ok(match res? {
                    Res::Error => Tri::Unknown,
                    Res::Ok => Tri::True,
                    Res::NotFound => Tri::False,
                })
            }
            _ => Err(PgError::internal("invalid boolean jsonpath item type")),
        }
    }

    /// predicate evaluates a predicate over every pair of its operands' results, as Postgres' executePredicate does.
    fn predicate(
        &mut self,
        left: &Node,
        right: Option<&Node>,
        jb: &Item,
        unwrap_right: bool,
        exec: &mut PredicateFn<'_, 'a>,
    ) -> Result<Tri> {
        let mut lseq = Vec::new();
        if self.unwrapped_results_quiet(left, jb, true, &mut lseq)? == Res::Error {
            return Ok(Tri::Unknown);
        }
        let mut rseq = Vec::new();
        if let Some(right) = right
            && self.unwrapped_results_quiet(right, jb, unwrap_right, &mut rseq)? == Res::Error
        {
            return Ok(Tri::Unknown);
        }
        let (mut error, mut found) = (false, false);
        for l in &lseq {
            let rights: Vec<Option<&Item>> = if right.is_some() { rseq.iter().map(Some).collect() } else { vec![None] };
            for r in rights {
                match exec(self, l, r)? {
                    Tri::Unknown => {
                        if !self.lax {
                            return Ok(Tri::Unknown);
                        }
                        error = true;
                    }
                    Tri::True => {
                        if self.lax {
                            return Ok(Tri::True);
                        }
                        found = true;
                    }
                    Tri::False => {}
                }
            }
        }
        Ok(if found {
            Tri::True
        } else if error {
            Tri::Unknown
        } else {
            Tri::False
        })
    }

    /// compare compares two items, as Postgres' compareItems does.
    fn compare(&mut self, op: CmpOp, a: &Item, b: &Item) -> Result<Tri> {
        let ordering = match (a, b) {
            (Item::Json(Json::Null), Item::Json(Json::Null)) => Ordering::Equal,
            (Item::Json(Json::Null), _) | (_, Item::Json(Json::Null)) => {
                return Ok(if op == CmpOp::Ne { Tri::True } else { Tri::False });
            }
            (Item::Json(Json::Bool(x)), Item::Json(Json::Bool(y))) => x.cmp(y),
            (Item::Json(Json::Number(x)), Item::Json(Json::Number(y))) => x.cmp_numeric(y),
            (Item::Json(Json::String(x)), Item::Json(Json::String(y))) => x.as_bytes().cmp(y.as_bytes()),
            (Item::DateTime(x, _), Item::DateTime(y, _)) => match self.compare_datetimes(x, y)? {
                Some(ordering) => ordering,
                None => return Ok(Tri::Unknown),
            },
            _ => return Ok(Tri::Unknown),
        };
        let result = match op {
            CmpOp::Eq => ordering.is_eq(),
            CmpOp::Ne => ordering.is_ne(),
            CmpOp::Lt => ordering.is_lt(),
            CmpOp::Le => ordering.is_le(),
            CmpOp::Gt => ordering.is_gt(),
            CmpOp::Ge => ordering.is_ge(),
        };
        Ok(if result { Tri::True } else { Tri::False })
    }

    /// compare_datetimes orders two datetime values, converting between their types as Postgres' compareDatetime
    /// does, or returns None for types that do not compare.
    fn compare_datetimes(&self, a: &Value, b: &Value) -> Result<Option<Ordering>> {
        let local_zone = |local: i64| i64::from(dt::with_format(|f| f.zone.offset_for_local(local))) * USECS_PER_SEC;
        let as_utc = |value: &Value| -> Option<i64> {
            match value {
                Value::Date(d) => {
                    let local = i64::from(*d) * USECS_PER_DAY;
                    Some(local - local_zone(local))
                }
                Value::Timestamp(local) => Some(local - local_zone(*local)),
                Value::TimestampTz(ts) => Some(*ts),
                _ => None,
            }
        };
        let name = |value: &Value| match value {
            Value::Date(_) => "date",
            Value::Time(_) => "time",
            Value::TimeTz(..) => "timetz",
            Value::Timestamp(_) => "timestamp",
            _ => "timestamptz",
        };
        Ok(match (a, b) {
            (Value::Date(x), Value::Date(y)) => Some(x.cmp(y)),
            (Value::Time(x), Value::Time(y)) => Some(x.cmp(y)),
            (Value::TimeTz(x, xw), Value::TimeTz(y, yw)) => {
                Some((x + i64::from(*xw) * USECS_PER_SEC).cmp(&(y + i64::from(*yw) * USECS_PER_SEC)))
            }
            (Value::Timestamp(x), Value::Timestamp(y)) | (Value::TimestampTz(x), Value::TimestampTz(y)) => {
                Some(x.cmp(y))
            }
            (Value::Date(d), Value::Timestamp(ts)) => Some((i64::from(*d) * USECS_PER_DAY).cmp(ts)),
            (Value::Timestamp(ts), Value::Date(d)) => Some(ts.cmp(&(i64::from(*d) * USECS_PER_DAY))),
            (Value::Time(t), Value::TimeTz(..)) | (Value::TimeTz(..), Value::Time(t)) => {
                self.require_tz("time", "timetz")?;
                let east = dt::with_format(|f| f.zone.offset_at(dt::clock()).0);
                let tz = match (a, b) {
                    (Value::TimeTz(x, w), _) | (_, Value::TimeTz(x, w)) => x + i64::from(*w) * USECS_PER_SEC,
                    _ => 0,
                };
                let plain = t - i64::from(east) * USECS_PER_SEC;
                Some(if matches!(a, Value::Time(_)) { plain.cmp(&tz) } else { tz.cmp(&plain) })
            }
            (
                Value::Date(_) | Value::Timestamp(_) | Value::TimestampTz(_),
                Value::Date(_) | Value::Timestamp(_) | Value::TimestampTz(_),
            ) => {
                let zoneless = if matches!(a, Value::TimestampTz(_)) { b } else { a };
                self.require_tz(name(zoneless), "timestamptz")?;
                match (as_utc(a), as_utc(b)) {
                    (Some(x), Some(y)) => Some(x.cmp(&y)),
                    _ => None,
                }
            }
            _ => None,
        })
    }
}

/// round_precision rounds a time or timestamp value's fractional seconds to a precision.
fn round_precision(value: Value, precision: i32) -> Value {
    let precision = precision.clamp(0, 6);
    let scale = 10i64.pow(6 - precision as u32);
    let round = |micros: i64| {
        let half = scale / 2;
        if micros >= 0 { (micros + half) / scale * scale } else { -((-micros + half) / scale * scale) }
    };
    match value {
        Value::Time(t) => Value::Time(round(t)),
        Value::TimeTz(t, w) => Value::TimeTz(round(t), w),
        Value::Timestamp(ts) => Value::Timestamp(round(ts)),
        Value::TimestampTz(ts) => Value::TimestampTz(round(ts)),
        other => other,
    }
}

/// Parsed is a datetime value that `.datetime()` read, with the offset east of UTC that its text named.
type Parsed = (Value, Option<i32>);

/// datetime_value builds the datetime value that a date, a time, and an offset make, as the type the fields present
/// choose.
fn datetime_value(fields: &Fields, dated: bool, timed: bool, offset: Option<i64>) -> Option<Parsed> {
    let time = ((fields.hour * 60 + fields.minute) * 60 + fields.second) * USECS_PER_SEC + fields.micros;
    let days = dt::date2j(fields.year, fields.month, fields.day) - POSTGRES_EPOCH_JDATE;
    let offset = offset.and_then(|o| i32::try_from(o).ok());
    Some(match (dated, timed, offset) {
        (true, false, None) => (Value::Date(i32::try_from(days).ok()?), None),
        (false, _, None) => (Value::Time(time), None),
        (false, _, Some(east)) => (Value::TimeTz(time, -east), None),
        (true, _, None) => (Value::Timestamp(days * USECS_PER_DAY + time), None),
        (true, _, Some(east)) => {
            (Value::TimestampTz(days * USECS_PER_DAY + time - i64::from(east) * USECS_PER_SEC), Some(east))
        }
    })
}

/// formatted_datetime reads text with a `.datetime()` template.
fn formatted_datetime(text: &str, template: &str) -> Result<Parsed> {
    let parts = crate::formatting::parse_datetime(text, template)?;
    let offset = if parts.zoned { Some(parts.offset.unwrap_or(0)) } else { None };
    datetime_value(&parts.fields, parts.dated || !parts.timed, parts.timed, offset)
        .ok_or_else(|| error(code::DATETIME_FIELD_OVERFLOW, format!("date/time field value out of range: \"{text}\"")))
}

/// iso_datetime reads text in one of the ISO 8601 formats that `.datetime()` recognizes without a template.
fn iso_datetime(text: &str) -> Option<Parsed> {
    let bytes = text.as_bytes();
    let digits = |from: usize, count: usize| -> Option<i64> {
        let slice = bytes.get(from..from + count)?;
        if !slice.iter().all(u8::is_ascii_digit) {
            return None;
        }
        std::str::from_utf8(slice).ok()?.parse().ok()
    };
    let mut fields = Fields { month: 1, day: 1, ..Fields::default() };
    let mut position = 0;
    let dated = bytes.len() >= 10 && bytes[4] == b'-' && bytes[7] == b'-';
    if dated {
        fields.year = digits(0, 4)?;
        fields.month = digits(5, 2)?;
        fields.day = digits(8, 2)?;
        if !(1..=12).contains(&fields.month)
            || fields.day < 1
            || fields.day > dt::days_in_month(fields.year, fields.month)
        {
            return None;
        }
        position = 10;
        if position == bytes.len() {
            return datetime_value(&fields, true, false, None);
        }
        if !matches!(bytes[position], b' ' | b'T') {
            return None;
        }
        position += 1;
    }
    fields.hour = digits(position, 2)?;
    fields.minute = digits(position + 3, 2)?;
    fields.second = digits(position + 6, 2)?;
    if bytes.get(position + 2) != Some(&b':') || bytes.get(position + 5) != Some(&b':') {
        return None;
    }
    if fields.hour > 23 || fields.minute > 59 || fields.second > 59 {
        return None;
    }
    position += 8;
    if bytes.get(position) == Some(&b'.') {
        let start = position + 1;
        let mut end = start;
        while end < bytes.len() && bytes[end].is_ascii_digit() && end - start < 6 {
            end += 1;
        }
        if end == start {
            return None;
        }
        let fraction = std::str::from_utf8(&bytes[start..end]).ok()?;
        fields.micros = format!("{fraction:0<6}").parse().ok()?;
        position = end;
    }
    let mut offset = None;
    if position < bytes.len() {
        let sign = match bytes[position] {
            b'+' => 1,
            b'-' => -1,
            _ => return None,
        };
        let hours = digits(position + 1, 2).or_else(|| digits(position + 1, 1))?;
        let mut rest = position + 1 + if digits(position + 1, 2).is_some() { 2 } else { 1 };
        let mut minutes = 0;
        if bytes.get(rest) == Some(&b':') {
            minutes = digits(rest + 1, 2)?;
            rest += 3;
        } else if let Some(m) = digits(rest, 2) {
            minutes = m;
            rest += 2;
        }
        if rest != bytes.len() || hours > 15 || minutes > 59 {
            return None;
        }
        offset = Some(sign * (hours * 3600 + minutes * 60));
    }
    datetime_value(&fields, dated, true, offset)
}

/// Options are how a path evaluates: its variables, whether errors are thrown, and whether datetime conversions may
/// use the session's time zone.
pub struct Options<'a> {
    pub vars: &'a HashMap<String, Item>,
    pub throw: bool,
    pub use_tz: bool,
}

/// query returns the items that a path finds in a document, or None when an error occurred that was not thrown.
pub fn query(path: &JsonPath, document: &Item, options: &Options<'_>) -> Result<Option<Vec<Item>>> {
    let mut items = Vec::new();
    Ok(query_into(path, document, options, &mut items)?.then_some(items))
}

/// query_into adds the items that a path finds in a document to a list, keeping the ones it found before an error
/// that was not thrown, as Postgres' executeJsonPath does, and reports whether no such error occurred.
pub fn query_into(path: &JsonPath, document: &Item, options: &Options<'_>, items: &mut Vec<Item>) -> Result<bool> {
    let mut exec = Exec {
        vars: options.vars,
        root: document.clone(),
        current: document.clone(),
        lax: path.lax,
        ignore_structural: path.lax,
        throw: options.throw,
        use_tz: options.use_tz,
        innermost_size: None,
    };
    Ok(exec.item(&path.expr, &[], document, Some(items))? != Res::Error)
}

/// exists reports whether a path finds any item in a document, or returns None when an error occurred that was not
/// thrown.
pub fn exists(path: &JsonPath, document: &Item, options: &Options<'_>) -> Result<Option<bool>> {
    if !path.lax {
        return Ok(query(path, document, options)?.map(|items| !items.is_empty()));
    }
    let mut exec = Exec {
        vars: options.vars,
        root: document.clone(),
        current: document.clone(),
        lax: true,
        ignore_structural: true,
        throw: options.throw,
        use_tz: options.use_tz,
        innermost_size: None,
    };
    Ok(match exec.item(&path.expr, &[], document, None)? {
        Res::Error => None,
        Res::Ok => Some(true),
        Res::NotFound => Some(false),
    })
}
