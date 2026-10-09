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

//! Text search functions over tsvector and tsquery values: matching a query against a vector as Postgres'
//! TS_execute does, the tsquery operators, and the tsvector functions that need no parser.

use super::Function;
use crate::basetypes::textsearch::{
    Lexeme, MAX_POSITION, Mode, Operator, Query, node_count, normalized, query_bytes, query_out, query_tree,
    vector_bytes, vector_lexemes,
};
use crate::dolt::procedures::RECORD;
use crate::error::{PgError, Result, code};
use crate::json::{Json, Raw, RawKind};
use crate::oid::{BOOL, CHAR, FLOAT4, INT2, INT4, JSON, JSONB, OID, REGCONFIG, REGDICTIONARY, TEXT, TEXT_ARRAY};
use crate::query::Ctx;
use crate::tsearch;
use crate::types::{BaseValue, Value};

/// TSVECTOR and TSQUERY are the text search types' OIDs.
const TSVECTOR: u32 = 3614;
const TSQUERY: u32 = 3615;

/// CHAR_ARRAY is the OID of the "char" array type.
const CHAR_ARRAY: u32 = 1002;

/// FLOAT4_ARRAY is the OID of the real array type.
const FLOAT4_ARRAY: u32 = 1021;

/// f declares a strict text search function.
const fn f(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}

/// FUNCTIONS are the text search functions.
pub const FUNCTIONS: &[Function] = &[
    f("@@", &[TSVECTOR, TSQUERY], BOOL, |_, args| matches(&args[0], &args[1])),
    f("@@", &[TSQUERY, TSVECTOR], BOOL, |_, args| matches(&args[1], &args[0])),
    f("@@@", &[TSVECTOR, TSQUERY], BOOL, |_, args| matches(&args[0], &args[1])),
    f("@@@", &[TSQUERY, TSVECTOR], BOOL, |_, args| matches(&args[1], &args[0])),
    f("ts_match_vq", &[TSVECTOR, TSQUERY], BOOL, |_, args| matches(&args[0], &args[1])),
    f("ts_match_qv", &[TSQUERY, TSVECTOR], BOOL, |_, args| matches(&args[1], &args[0])),
    f("&&", &[TSQUERY, TSQUERY], TSQUERY, |_, args| combine(&args[0], &args[1], Operator::And, 0)),
    f("||", &[TSQUERY, TSQUERY], TSQUERY, |_, args| combine(&args[0], &args[1], Operator::Or, 0)),
    f("<->", &[TSQUERY, TSQUERY], TSQUERY, |_, args| combine(&args[0], &args[1], Operator::Phrase, 1)),
    f("tsquery_and", &[TSQUERY, TSQUERY], TSQUERY, |_, args| combine(&args[0], &args[1], Operator::And, 0)),
    f("tsquery_or", &[TSQUERY, TSQUERY], TSQUERY, |_, args| combine(&args[0], &args[1], Operator::Or, 0)),
    f("tsquery_phrase", &[TSQUERY, TSQUERY], TSQUERY, |_, args| combine(&args[0], &args[1], Operator::Phrase, 1)),
    f("tsquery_phrase", &[TSQUERY, TSQUERY, INT4], TSQUERY, phrase_distance),
    f("!!", &[TSQUERY], TSQUERY, negate),
    f("tsquery_not", &[TSQUERY], TSQUERY, negate),
    f("numnode", &[TSQUERY], INT4, |_, args| {
        Ok(Value::Int4(query(&args[0])?.as_ref().map_or(0, |q| node_count(q) as i32)))
    }),
    f("querytree", &[TSQUERY], TEXT, querytree),
    f("@>", &[TSQUERY, TSQUERY], BOOL, |_, args| contains(&args[0], &args[1])),
    f("<@", &[TSQUERY, TSQUERY], BOOL, |_, args| contains(&args[1], &args[0])),
    f("||", &[TSVECTOR, TSVECTOR], TSVECTOR, concat),
    f("tsvector_concat", &[TSVECTOR, TSVECTOR], TSVECTOR, concat),
    f("strip", &[TSVECTOR], TSVECTOR, strip),
    f("length", &[TSVECTOR], INT4, |_, args| Ok(Value::Int4(lexemes(&args[0])?.len() as i32))),
    f("setweight", &[TSVECTOR, CHAR], TSVECTOR, setweight),
    f("setweight", &[TSVECTOR, CHAR, TEXT_ARRAY], TSVECTOR, setweight),
    f("ts_delete", &[TSVECTOR, TEXT], TSVECTOR, delete),
    f("ts_delete", &[TSVECTOR, TEXT_ARRAY], TSVECTOR, delete),
    f("ts_filter", &[TSVECTOR, CHAR_ARRAY], TSVECTOR, filter),
    f("tsvector_to_array", &[TSVECTOR], TEXT_ARRAY, to_array),
    f("array_to_tsvector", &[TEXT_ARRAY], TSVECTOR, from_array),
    f("unnest", &[TSVECTOR], RECORD, unnest),
    f("ts_rank", &[TSVECTOR, TSQUERY], FLOAT4, |_, args| rank(args, false)),
    f("ts_rank", &[TSVECTOR, TSQUERY, INT4], FLOAT4, |_, args| rank(args, false)),
    f("ts_rank", &[FLOAT4_ARRAY, TSVECTOR, TSQUERY], FLOAT4, |_, args| rank(args, false)),
    f("ts_rank", &[FLOAT4_ARRAY, TSVECTOR, TSQUERY, INT4], FLOAT4, |_, args| rank(args, false)),
    f("ts_rank_cd", &[TSVECTOR, TSQUERY], FLOAT4, |_, args| rank(args, true)),
    f("ts_rank_cd", &[TSVECTOR, TSQUERY, INT4], FLOAT4, |_, args| rank(args, true)),
    f("ts_rank_cd", &[FLOAT4_ARRAY, TSVECTOR, TSQUERY], FLOAT4, |_, args| rank(args, true)),
    f("ts_rank_cd", &[FLOAT4_ARRAY, TSVECTOR, TSQUERY, INT4], FLOAT4, |_, args| rank(args, true)),
    f("to_tsvector", &[REGCONFIG, TEXT], TSVECTOR, |ctx, args| to_tsvector(ctx, Some(&args[0]), &args[1])),
    f("to_tsvector", &[TEXT], TSVECTOR, |ctx, args| to_tsvector(ctx, None, &args[0])),
    f("to_tsvector", &[REGCONFIG, JSON], TSVECTOR, |ctx, args| json_vector(ctx, Some(&args[0]), &args[1], None)),
    f("to_tsvector", &[JSON], TSVECTOR, |ctx, args| json_vector(ctx, None, &args[0], None)),
    f("to_tsvector", &[REGCONFIG, JSONB], TSVECTOR, |ctx, args| json_vector(ctx, Some(&args[0]), &args[1], None)),
    f("to_tsvector", &[JSONB], TSVECTOR, |ctx, args| json_vector(ctx, None, &args[0], None)),
    f("json_to_tsvector", &[REGCONFIG, JSON, JSONB], TSVECTOR, |ctx, args| {
        json_vector(ctx, Some(&args[0]), &args[1], Some(&args[2]))
    }),
    f("json_to_tsvector", &[JSON, JSONB], TSVECTOR, |ctx, args| json_vector(ctx, None, &args[0], Some(&args[1]))),
    f("jsonb_to_tsvector", &[REGCONFIG, JSONB, JSONB], TSVECTOR, |ctx, args| {
        json_vector(ctx, Some(&args[0]), &args[1], Some(&args[2]))
    }),
    f("jsonb_to_tsvector", &[JSONB, JSONB], TSVECTOR, |ctx, args| json_vector(ctx, None, &args[0], Some(&args[1]))),
    f("to_tsquery", &[REGCONFIG, TEXT], TSQUERY, |ctx, args| {
        to_query(ctx, Some(&args[0]), &args[1], Mode::Standard, Operator::Phrase)
    }),
    f("to_tsquery", &[TEXT], TSQUERY, |ctx, args| to_query(ctx, None, &args[0], Mode::Standard, Operator::Phrase)),
    f("plainto_tsquery", &[REGCONFIG, TEXT], TSQUERY, |ctx, args| {
        to_query(ctx, Some(&args[0]), &args[1], Mode::Plain, Operator::And)
    }),
    f("plainto_tsquery", &[TEXT], TSQUERY, |ctx, args| to_query(ctx, None, &args[0], Mode::Plain, Operator::And)),
    f("phraseto_tsquery", &[REGCONFIG, TEXT], TSQUERY, |ctx, args| {
        to_query(ctx, Some(&args[0]), &args[1], Mode::Plain, Operator::Phrase)
    }),
    f("phraseto_tsquery", &[TEXT], TSQUERY, |ctx, args| to_query(ctx, None, &args[0], Mode::Plain, Operator::Phrase)),
    f("websearch_to_tsquery", &[REGCONFIG, TEXT], TSQUERY, |ctx, args| {
        to_query(ctx, Some(&args[0]), &args[1], Mode::Web, Operator::Phrase)
    }),
    f("websearch_to_tsquery", &[TEXT], TSQUERY, |ctx, args| to_query(ctx, None, &args[0], Mode::Web, Operator::Phrase)),
    f("ts_lexize", &[REGDICTIONARY, TEXT], TEXT_ARRAY, lexize),
    f("ts_token_type", &[TEXT], RECORD, |_, args| token_types(&args[0])),
    f("ts_token_type", &[OID], RECORD, |_, args| token_types(&args[0])),
    f("ts_parse", &[TEXT, TEXT], RECORD, |_, args| parse(&args[0], &args[1])),
    f("ts_parse", &[OID, TEXT], RECORD, |_, args| parse(&args[0], &args[1])),
    f("ts_debug", &[REGCONFIG, TEXT], RECORD, |ctx, args| debug(ctx, Some(&args[0]), &args[1])),
    f("ts_debug", &[TEXT], RECORD, |ctx, args| debug(ctx, None, &args[0])),
    f("get_current_ts_config", &[], REGCONFIG, |ctx, _| {
        let name = ctx.session.settings.get("default_text_search_config").unwrap_or_default();
        ctx.reg_value(Value::Text(name), REGCONFIG)
    }),
];

/// OUT_COLUMNS are the columns of the text search functions that return rows of several columns.
pub const OUT_COLUMNS: &[(&str, &[(&str, u32)])] = &[
    ("ts_token_type", &[("tokid", INT4), ("alias", TEXT), ("description", TEXT)]),
    ("ts_parse", &[("tokid", INT4), ("token", TEXT)]),
    (
        "ts_debug",
        &[
            ("alias", TEXT),
            ("description", TEXT),
            ("token", TEXT),
            ("dictionaries", REGDICTIONARY_ARRAY),
            ("dictionary", REGDICTIONARY),
            ("lexemes", TEXT_ARRAY),
        ],
    ),
];

/// DEFAULT_PARSER is the OID of the default text search parser.
const DEFAULT_PARSER: u32 = 3722;

/// REGDICTIONARY_ARRAY is the OID of the regdictionary array type.
const REGDICTIONARY_ARRAY: u32 = 3770;

/// UNNEST_COLUMNS are the columns of unnest over a tsvector.
pub const UNNEST_COLUMNS: &[(&str, u32)] = &[("lexeme", TEXT), ("positions", 1005), ("weights", TEXT_ARRAY)];

/// lexemes returns a tsvector argument's lexemes.
fn lexemes(value: &Value) -> Result<Vec<Lexeme>> {
    match value {
        Value::Base(base) => vector_lexemes(&base.data),
        _ => Err(PgError::internal("a tsvector that is not one")),
    }
}

/// query returns a tsquery argument's tree, which is None for an empty query.
fn query(value: &Value) -> Result<Option<Query>> {
    match value {
        Value::Base(base) => query_tree(&base.data),
        _ => Err(PgError::internal("a tsquery that is not one")),
    }
}

/// vector returns a tsvector value of lexemes.
fn vector(lexemes: &[Lexeme]) -> Value {
    Value::Base(Box::new(BaseValue { type_oid: TSVECTOR, data: vector_bytes(lexemes) }))
}

/// query_value returns a tsquery value of a tree.
fn query_value(query: Option<&Query>) -> Value {
    Value::Base(Box::new(BaseValue { type_oid: TSQUERY, data: query_bytes(query) }))
}

/// Check reports whether a query lexeme matches and at which positions, which are None when the matched lexemes have
/// none.
type Check<'a> = &'a dyn Fn(&Query) -> (bool, Option<Vec<u16>>);

/// matches reports whether a tsvector matches a tsquery, as Postgres' ts_match_vq does.
fn matches(vector: &Value, query_value: &Value) -> Result<Value> {
    let lexemes = lexemes(vector)?;
    let Some(query) = query(query_value)? else { return Ok(Value::Bool(false)) };
    Ok(Value::Bool(execute(&query, &|q| in_vector(&lexemes, q))))
}

/// in_vector checks a query lexeme against a vector's lexemes.
fn in_vector(lexemes: &[Lexeme], query: &Query) -> (bool, Option<Vec<u16>>) {
    match query {
        Query::Lexeme { word, weights, prefix } => found(lexemes, word, *weights, *prefix),
        _ => (false, None),
    }
}

/// found returns the positions of a query lexeme in a vector, of the weights it asks for, and whether the vector holds
/// the lexeme at all, matching a prefix when the lexeme asks for one.
fn found(lexemes: &[Lexeme], word: &[u8], weights: u8, prefix: bool) -> (bool, Option<Vec<u16>>) {
    let mut any = false;
    let mut positions = Vec::new();
    let mut positional = true;
    for (lexeme, entries) in lexemes {
        let hit = if prefix { lexeme.starts_with(word) } else { lexeme.as_slice() == word };
        if !hit {
            continue;
        }
        if entries.is_empty() {
            positional = false;
            any = true;
            continue;
        }
        for &entry in entries {
            let weight = 1u8 << (entry >> 14);
            if weights == 0 || weights & weight != 0 {
                any = true;
                positions.push(entry & MAX_POSITION);
            }
        }
    }
    positions.sort_unstable();
    positions.dedup();
    (any, positional.then_some(positions))
}

/// execute matches a query, as Postgres' TS_execute_recurse does, checking each lexeme with `check`.
fn execute(query: &Query, check: Check<'_>) -> bool {
    match query {
        Query::Lexeme { .. } => check(query).0,
        Query::Not(inner) => !execute(inner, check),
        Query::Binary { operator: Operator::And, left, right, .. } => execute(left, check) && execute(right, check),
        Query::Binary { operator: Operator::Or, left, right, .. } => execute(left, check) || execute(right, check),
        Query::Binary { .. } => phrase(query, check).is_some_and(|p| p.negate || !p.positions.is_empty()),
    }
}

/// Phrase is where part of a phrase matches in a vector: the positions where it ends, or where it does not when
/// `negate` is set, and how many positions it spans before its end.
struct Phrase {
    positions: Vec<u16>,
    width: u16,
    negate: bool,
}

/// Emit is which positions joining two parts of a phrase keeps, as Postgres' TSPO flags name them: positions both
/// parts share, and those of only the left or only the right part.
struct Emit {
    both: bool,
    left_only: bool,
    right_only: bool,
}

/// output joins two parts' positions, the left ones moved by an offset, keeping those that `emit` asks for, as
/// Postgres' TS_phrase_output does.
fn output(left: &[u16], right: &[u16], left_offset: u16, emit: Emit) -> Vec<u16> {
    let shifted: Vec<u16> = left.iter().map(|&l| l.saturating_add(left_offset)).collect();
    let mut out = Vec::new();
    for &l in &shifted {
        let shared = right.contains(&l);
        if (shared && emit.both) || (!shared && emit.left_only) {
            out.push(l);
        }
    }
    for &r in right {
        if !shifted.contains(&r) && emit.right_only {
            out.push(r);
        }
    }
    out.sort_unstable();
    out.dedup();
    out
}

/// phrase returns where a part of a phrase matches, as Postgres' TS_phrase_execute does, or None when the matched
/// lexemes lack the positions to tell.
fn phrase(query: &Query, check: Check<'_>) -> Option<Phrase> {
    let emit = |both, left_only, right_only| Emit { both, left_only, right_only };
    let no = || Some(Phrase { positions: Vec::new(), width: 0, negate: false });
    let is_no = |p: &Option<Phrase>| p.as_ref().is_some_and(|p| !p.negate && p.positions.is_empty());
    match query {
        Query::Lexeme { .. } => match check(query) {
            (_, Some(positions)) => Some(Phrase { positions, width: 0, negate: false }),
            (true, None) => None,
            (false, None) => no(),
        },
        Query::Not(inner) => phrase(inner, check).map(|p| Phrase { negate: !p.negate, ..p }),
        Query::Binary { operator, distance, left, right } => {
            let left = phrase(left, check);
            if *operator != Operator::Or && is_no(&left) {
                return no();
            }
            let right = phrase(right, check);
            if is_no(&right) && (*operator != Operator::Or || is_no(&left)) {
                return no();
            }
            let (mut left, mut right) = (left?, right?);
            if *operator == Operator::Or {
                for side in [&mut left, &mut right] {
                    if !side.negate && side.positions.is_empty() {
                        side.width = 0;
                    }
                }
            }
            let (l, r) = (&left.positions, &right.positions);
            Some(match operator {
                Operator::Phrase => {
                    let offset = right.width + distance;
                    let (positions, negate) = match (left.negate, right.negate) {
                        (false, false) => (output(l, r, offset, emit(true, false, false)), false),
                        (true, false) => (output(l, r, offset, emit(false, false, true)), false),
                        (false, true) => (output(l, r, offset, emit(false, true, false)), false),
                        (true, true) => (output(l, r, offset, emit(true, true, true)), true),
                    };
                    Phrase { positions, width: left.width + distance + right.width, negate }
                }
                Operator::And => {
                    let (positions, negate) = match (left.negate, right.negate) {
                        (true, true) => (output(l, r, 0, emit(true, true, true)), true),
                        (true, false) => (output(l, r, 0, emit(false, false, true)), false),
                        (false, true) => (output(l, r, 0, emit(false, true, false)), false),
                        (false, false) => (output(l, r, 0, emit(true, false, false)), false),
                    };
                    Phrase { positions, width: left.width.max(right.width), negate }
                }
                _ => {
                    let (positions, negate) = match (left.negate, right.negate) {
                        (true, true) => (output(l, r, 0, emit(true, false, false)), true),
                        (true, false) => (output(l, r, 0, emit(false, true, false)), true),
                        (false, true) => (output(l, r, 0, emit(false, false, true)), true),
                        (false, false) => (output(l, r, 0, emit(true, true, true)), false),
                    };
                    Phrase { positions, width: left.width.max(right.width), negate }
                }
            })
        }
    }
}

/// combine joins two queries with an operator, returning the other query when one is empty, as Postgres'
/// tsquery_and, tsquery_or, and tsquery_phrase do.
fn combine(left: &Value, right: &Value, operator: Operator, distance: u16) -> Result<Value> {
    let (left, right) = (query(left)?, query(right)?);
    let joined = match (left, right) {
        (None, other) | (other, None) => other,
        (Some(left), Some(right)) => {
            Some(Query::Binary { operator, distance, left: Box::new(left), right: Box::new(right) })
        }
    };
    Ok(query_value(joined.as_ref()))
}

/// phrase_distance joins two queries with the phrase operator at a distance, as tsquery_phrase does.
fn phrase_distance(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let distance = match &args[2] {
        Value::Int4(d) if (0..=i32::from(MAX_POSITION)).contains(d) => *d as u16,
        _ => {
            return Err(PgError::new(
                code::INVALID_PARAMETER_VALUE,
                format!(
                    "distance in phrase operator must be an integer value between zero and {MAX_POSITION} inclusive"
                ),
            ));
        }
    };
    combine(&args[0], &args[1], Operator::Phrase, distance)
}

/// negate negates a query, as tsquery_not does.
fn negate(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let negated = query(&args[0])?.map(|q| Query::Not(Box::new(q)));
    Ok(query_value(negated.as_ref()))
}

/// querytree returns the part of a query that an index can use, which leaves out its negations, as Postgres'
/// querytree does, or T when nothing is left.
fn querytree(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    /// clean drops negations from a query.
    fn clean(query: Query) -> Option<Query> {
        match query {
            Query::Not(_) => None,
            Query::Binary { operator, distance, left, right } => match (clean(*left), clean(*right)) {
                (Some(left), Some(right)) => {
                    Some(Query::Binary { operator, distance, left: Box::new(left), right: Box::new(right) })
                }
                (Some(only), None) | (None, Some(only)) if operator != Operator::Or => Some(only),
                _ => None,
            },
            lexeme => Some(lexeme),
        }
    }
    Ok(match query(&args[0])? {
        None => Value::Text(String::new()),
        Some(tree) => Value::Text(clean(tree).map_or_else(|| "T".to_string(), |q| query_out(Some(&q)))),
    })
}

/// contains reports whether a query holds every lexeme of another, as tsq_mcontains does.
fn contains(outer: &Value, inner: &Value) -> Result<Value> {
    /// words collects a query's lexemes.
    fn words(query: &Query, out: &mut Vec<Vec<u8>>) {
        match query {
            Query::Lexeme { word, .. } => out.push(word.clone()),
            Query::Not(inner) => words(inner, out),
            Query::Binary { left, right, .. } => {
                words(left, out);
                words(right, out);
            }
        }
    }
    let (mut outer_words, mut inner_words) = (Vec::new(), Vec::new());
    if let Some(q) = query(outer)? {
        words(&q, &mut outer_words);
    }
    if let Some(q) = query(inner)? {
        words(&q, &mut inner_words);
    }
    Ok(Value::Bool(inner_words.iter().all(|w| outer_words.contains(w))))
}

/// concat joins two vectors, moving the second one's positions past the first one's last position, as
/// tsvector_concat does.
fn concat(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let left = lexemes(&args[0])?;
    let right = lexemes(&args[1])?;
    let shift = left.iter().flat_map(|(_, p)| p.iter().map(|e| e & MAX_POSITION)).max().unwrap_or(0);
    let shifted: Vec<Lexeme> = right
        .into_iter()
        .map(|(word, positions)| {
            let positions = positions
                .into_iter()
                .map(|e| {
                    let position = (e & MAX_POSITION).saturating_add(shift).min(MAX_POSITION);
                    (e & !MAX_POSITION) | position
                })
                .collect();
            (word, positions)
        })
        .collect();
    Ok(vector(&normalized(left.into_iter().chain(shifted).collect())))
}

/// strip removes a vector's positions.
fn strip(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let stripped: Vec<Lexeme> = lexemes(&args[0])?.into_iter().map(|(word, _)| (word, Vec::new())).collect();
    Ok(vector(&stripped))
}

/// weight_of returns the two weight bits that a weight letter stands for, failing as Postgres does for another.
fn weight_of(value: &Value) -> Result<u16> {
    let letter = value.output().unwrap_or_default();
    match letter.to_ascii_uppercase().as_str() {
        "A" => Ok(3),
        "B" => Ok(2),
        "C" => Ok(1),
        "D" => Ok(0),
        _ => Err(PgError::new(
            code::INTERNAL_ERROR,
            format!("unrecognized weight: {}", letter.chars().next().map_or(0, |c| c as u32)),
        )),
    }
}

/// text_items returns the texts of a text array, skipping its NULL elements as Postgres does.
fn text_items(value: &Value) -> Vec<String> {
    let Value::Array(array) = value else { return vec![value.output().unwrap_or_default()] };
    array.values.iter().filter_map(Value::output).collect()
}

/// setweight gives a weight to every position of a vector, or to those of the listed lexemes.
fn setweight(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let weight = weight_of(&args[1])?;
    let only = args.get(2).map(text_items);
    let weighted: Vec<Lexeme> = lexemes(&args[0])?
        .into_iter()
        .map(|(word, positions)| {
            let chosen = only.as_ref().is_none_or(|o| o.iter().any(|w| w.as_bytes() == word.as_slice()));
            let positions = match chosen {
                true => positions.into_iter().map(|e| (e & MAX_POSITION) | (weight << 14)).collect(),
                false => positions,
            };
            (word, positions)
        })
        .collect();
    Ok(vector(&weighted))
}

/// delete removes lexemes from a vector.
fn delete(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let doomed = text_items(&args[1]);
    let kept: Vec<Lexeme> = lexemes(&args[0])?
        .into_iter()
        .filter(|(word, _)| !doomed.iter().any(|d| d.as_bytes() == word.as_slice()))
        .collect();
    Ok(vector(&kept))
}

/// filter keeps the positions of a vector with the given weights, and the lexemes left with any.
fn filter(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let Value::Array(weights) = &args[1] else { return Ok(args[0].clone()) };
    let mut mask = 0u8;
    for weight in &weights.values {
        if weight.is_null() {
            return Err(PgError::new(code::NULL_VALUE_NOT_ALLOWED, "weight array may not contain nulls"));
        }
        mask |= 1 << weight_of(weight)?;
    }
    let kept: Vec<Lexeme> = lexemes(&args[0])?
        .into_iter()
        .filter_map(|(word, positions)| {
            let positions: Vec<u16> = positions.into_iter().filter(|e| mask & (1 << (e >> 14)) != 0).collect();
            (!positions.is_empty()).then_some((word, positions))
        })
        .collect();
    Ok(vector(&kept))
}

/// to_array returns a vector's lexemes.
fn to_array(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let words: Vec<Value> = lexemes(&args[0])?
        .into_iter()
        .map(|(word, _)| Value::Text(String::from_utf8_lossy(&word).into_owned()))
        .collect();
    Ok(Value::Array(Box::new(crate::array::Array::one_dimensional(TEXT, words))))
}

/// from_array builds a vector without positions from lexemes.
fn from_array(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    if let Value::Array(array) = &args[0]
        && array.values.iter().any(Value::is_null)
    {
        return Err(PgError::new(code::NULL_VALUE_NOT_ALLOWED, "lexeme array may not contain nulls"));
    }
    let words = text_items(&args[0]);
    if words.iter().any(String::is_empty) {
        return Err(PgError::new("2200F", "lexeme array may not contain empty strings"));
    }
    let lexemes: Vec<Lexeme> = words.into_iter().map(|w| (w.into_bytes(), Vec::new())).collect();
    Ok(vector(&normalized(lexemes)))
}

/// unnest returns a vector's lexemes as rows of the lexeme, its positions, and their weights.
fn unnest(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let rows = lexemes(&args[0])?
        .into_iter()
        .map(|(word, positions)| {
            let (values, weights) = match positions.is_empty() {
                true => (Value::Null, Value::Null),
                false => (
                    Value::Array(Box::new(crate::array::Array::one_dimensional(
                        INT2,
                        positions.iter().map(|e| Value::Int2((e & MAX_POSITION) as i16)).collect(),
                    ))),
                    Value::Array(Box::new(crate::array::Array::one_dimensional(
                        TEXT,
                        positions
                            .iter()
                            .map(|e| Value::Text(["D", "C", "B", "A"][usize::from(e >> 14)].to_string()))
                            .collect(),
                    ))),
                ),
            };
            Value::Record(vec![Value::Text(String::from_utf8_lossy(&word).into_owned()), values, weights])
        })
        .collect();
    Ok(Value::Set(rows))
}

/// WEIGHTS are the default weights of D, C, B, and A positions when ranking.
const WEIGHTS: [f32; 4] = [0.1, 0.2, 0.4, 1.0];

/// rank ranks a vector against a query as Postgres' ts_rank does, or as ts_rank_cd does when `cover` is set, taking
/// optional weights before them and an optional normalization method after them.
fn rank(args: &[Value], cover: bool) -> Result<Value> {
    let (weights, args) = match &args[0] {
        Value::Array(array) => (rank_weights(array)?, &args[1..]),
        _ => (WEIGHTS, args),
    };
    let method = match args.get(2) {
        Some(Value::Int4(method)) => *method,
        _ => 0,
    };
    let lexemes = lexemes(&args[0])?;
    let Some(query) = query(&args[1])? else { return Ok(Value::Float4(0.0)) };
    if lexemes.is_empty() {
        return Ok(Value::Float4(0.0));
    }
    Ok(Value::Float4(match cover {
        true => rank_cover(&weights, &lexemes, &query, method),
        false => rank_words(&weights, &lexemes, &query, method),
    }))
}

/// rank_weights reads the weights given to ranking, keeping the default for each negative one, as Postgres'
/// getWeights does.
fn rank_weights(array: &crate::array::Array) -> Result<[f32; 4]> {
    if array.dims.len() != 1 {
        return Err(PgError::new(code::ARRAY_SUBSCRIPT_ERROR, "array of weight must be one-dimensional"));
    }
    if array.values.len() < WEIGHTS.len() {
        return Err(PgError::new(code::ARRAY_SUBSCRIPT_ERROR, "array of weight is too short"));
    }
    let mut weights = WEIGHTS;
    for (weight, value) in weights.iter_mut().zip(&array.values) {
        match value {
            Value::Float4(given) if *given >= 0.0 => *weight = *given,
            Value::Float4(_) => {}
            _ => return Err(PgError::new(code::NULL_VALUE_NOT_ALLOWED, "array of weight must not contain nulls")),
        }
        if *weight > 1.0 {
            return Err(PgError::new(code::INVALID_PARAMETER_VALUE, "weight out of range"));
        }
    }
    Ok(weights)
}

/// operands returns a query's lexemes in the order Postgres stores them.
fn operands<'q>(query: &'q Query, out: &mut Vec<&'q Query>) {
    match query {
        Query::Lexeme { .. } => out.push(query),
        Query::Not(inner) => operands(inner, out),
        Query::Binary { left, right, .. } => {
            operands(right, out);
            operands(left, out);
        }
    }
}

/// entries returns the indexes of a vector's lexemes that a query lexeme matches, as Postgres' find_wordentry finds
/// them.
fn entries(lexemes: &[Lexeme], operand: &Query) -> Vec<usize> {
    let Query::Lexeme { word, prefix, .. } = operand else { return Vec::new() };
    let hit = |l: &[u8]| if *prefix { l.starts_with(word) } else { l == word.as_slice() };
    (0..lexemes.len()).filter(|&i| hit(&lexemes[i].0)).collect()
}

/// word_of returns a query lexeme's text.
fn word_of(operand: &Query) -> &[u8] {
    match operand {
        Query::Lexeme { word, .. } => word,
        _ => &[],
    }
}

/// weight_of_position returns the weight that ranking gives a position.
fn weight_of_position(weights: &[f32; 4], entry: u16) -> f32 {
    weights[usize::from(entry >> 14)]
}

/// word_count returns how many words a vector stands for, as Postgres' cnt_length counts them.
fn word_count(lexemes: &[Lexeme]) -> usize {
    lexemes.iter().map(|(_, positions)| positions.len().max(1)).sum()
}

/// rank_words is Postgres' calc_rank, which ranks by how often and how close together the query's words appear.
fn rank_words(weights: &[f32; 4], lexemes: &[Lexeme], query: &Query, method: i32) -> f32 {
    let mut items = Vec::new();
    operands(query, &mut items);
    items.sort_by(|a, b| word_of(a).cmp(word_of(b)));
    items.dedup_by(|a, b| word_of(a) == word_of(b));
    let mut res = match query {
        Query::Binary { operator: Operator::And | Operator::Phrase, .. } if items.len() >= 2 => {
            rank_and(weights, lexemes, &items)
        }
        _ => rank_or(weights, lexemes, &items),
    };
    if res < 0.0 {
        res = 1e-20;
    }
    if method & 0x01 != 0 {
        res = (f64::from(res) / ((word_count(lexemes) as f64 + 1.0).ln() / 2f64.ln())) as f32;
    }
    if method & 0x02 != 0 {
        let len = word_count(lexemes);
        if len > 0 {
            res /= len as f32;
        }
    }
    if method & 0x08 != 0 {
        res /= lexemes.len() as f32;
    }
    if method & 0x10 != 0 {
        res = (f64::from(res) / ((lexemes.len() as f64 + 1.0).ln() / 2f64.ln())) as f32;
    }
    if method & 0x20 != 0 {
        res /= res + 1.0;
    }
    res
}

/// word_distance returns the weight of two words a distance apart, as Postgres' word_distance does.
fn word_distance(distance: i32) -> f32 {
    if distance > 100 {
        return 1e-30;
    }
    (1.0 / (1.005 + 0.05 * (f64::from(distance as f32) / 1.5 - 2.0).exp())) as f32
}

/// rank_and is Postgres' calc_rank_and, which ranks by how close together the query's words appear.
fn rank_and(weights: &[f32; 4], lexemes: &[Lexeme], items: &[&Query]) -> f32 {
    let mut res = -1.0f32;
    let mut seen: Vec<Option<(Vec<u16>, bool)>> = vec![None; items.len()];
    for i in 0..items.len() {
        for entry in entries(lexemes, items[i]) {
            let positions = &lexemes[entry].1;
            let current = match positions.is_empty() {
                true => (vec![MAX_POSITION], true),
                false => (positions.clone(), false),
            };
            for earlier in seen[..i].iter().flatten() {
                for &l in &current.0 {
                    for &p in &earlier.0 {
                        let mut distance = (i32::from(l & MAX_POSITION) - i32::from(p & MAX_POSITION)).abs();
                        if distance == 0 && !current.1 && !earlier.1 {
                            continue;
                        }
                        if distance == 0 {
                            distance = i32::from(MAX_POSITION) + 1;
                        }
                        let product =
                            weight_of_position(weights, l) * weight_of_position(weights, p) * word_distance(distance);
                        let curw = f64::from(product).sqrt() as f32;
                        res = match res < 0.0 {
                            true => curw,
                            false => (1.0 - (1.0 - f64::from(res)) * (1.0 - f64::from(curw))) as f32,
                        };
                    }
                }
            }
            seen[i] = Some(current);
        }
    }
    res
}

/// rank_or is Postgres' calc_rank_or, which ranks by how often the query's words appear and with what weights.
fn rank_or(weights: &[f32; 4], lexemes: &[Lexeme], items: &[&Query]) -> f32 {
    let mut res = 0.0f32;
    for item in items {
        for entry in entries(lexemes, item) {
            let positions = match lexemes[entry].1.is_empty() {
                true => vec![0],
                false => lexemes[entry].1.clone(),
            };
            let (mut resj, mut wjm, mut jm) = (0.0f32, -1.0f32, 0usize);
            for (j, &position) in positions.iter().enumerate() {
                let weight = weight_of_position(weights, position);
                resj += weight / ((j + 1) * (j + 1)) as f32;
                if weight > wjm {
                    wjm = weight;
                    jm = j;
                }
            }
            let sum = wjm + resj - wjm / ((jm + 1) * (jm + 1)) as f32;
            res = (f64::from(res) + f64::from(sum) / 1.64493406685) as f32;
        }
    }
    if !items.is_empty() {
        res /= items.len() as f32;
    }
    res
}

/// DocEntry is a position where query lexemes match a vector, as Postgres' DocRepresentation is: the position with
/// its weight, the vector lexeme there, and the query lexemes that match it.
struct DocEntry<'q> {
    position: u16,
    entry: usize,
    items: Vec<&'q Query>,
}

/// Cover is the search state of Postgres' Cover: where the next search starts, and the found cover's first and last
/// document entries and positions.
#[derive(Default)]
struct Cover {
    next: usize,
    begin: usize,
    end: usize,
    first: i32,
    last: i32,
}

/// document returns where a query's lexemes match a vector, in order of position, as Postgres' get_docrep does.
fn document<'q>(lexemes: &[Lexeme], query: &'q Query) -> Vec<DocEntry<'q>> {
    let mut items = Vec::new();
    operands(query, &mut items);
    let mut found: Vec<(u16, usize, &Query)> = Vec::new();
    for item in items {
        let Query::Lexeme { weights, .. } = item else { continue };
        for entry in entries(lexemes, item) {
            for &position in &lexemes[entry].1 {
                if *weights == 0 || weights & (1 << (position >> 14)) != 0 {
                    found.push((position, entry, item));
                }
            }
        }
    }
    found.sort_by_key(|&(position, entry, _)| (position & MAX_POSITION, position >> 14, entry));
    let mut doc: Vec<DocEntry<'q>> = Vec::new();
    for (position, entry, item) in found {
        match doc.last_mut() {
            Some(last) if last.position == position && last.entry == entry => last.items.push(item),
            _ => doc.push(DocEntry { position, entry, items: vec![item] }),
        }
    }
    doc
}

/// cover_matches reports whether the query matches the document entries in a range, scanned in the given order, as
/// Postgres' Cover checks with its QueryRepresentation.
fn cover_matches(query: &Query, seen: &[(&Query, Vec<u16>)]) -> bool {
    execute(query, &|operand| match seen.iter().find(|(q, _)| std::ptr::eq(*q, operand)) {
        Some((_, positions)) => {
            let mut positions: Vec<u16> = positions.iter().map(|p| p & MAX_POSITION).collect();
            positions.sort_unstable();
            (true, Some(positions))
        }
        None => (false, Some(Vec::new())),
    })
}

/// see records a document entry's positions for each query lexeme that matches it, skipping a position the lexeme
/// already has last, as Postgres' fillQueryRepresentationData does.
fn see<'q>(seen: &mut Vec<(&'q Query, Vec<u16>)>, doc: &DocEntry<'q>) {
    for &item in &doc.items {
        match seen.iter_mut().find(|(q, _)| std::ptr::eq(*q, item)) {
            Some((_, positions)) => {
                if positions.last().is_none_or(|&p| p & MAX_POSITION != doc.position & MAX_POSITION) {
                    positions.push(doc.position);
                }
            }
            None => seen.push((item, vec![doc.position])),
        }
    }
}

/// next_cover finds the next shortest stretch of the document that matches the query, as Postgres' Cover does.
fn next_cover(doc: &[DocEntry<'_>], query: &Query, cover: &mut Cover) -> bool {
    loop {
        let mut seen = Vec::new();
        let (mut first, mut last, mut last_index) = (i32::MAX, 0, cover.next);
        let mut found = false;
        for (i, entry) in doc.iter().enumerate().skip(cover.next) {
            see(&mut seen, entry);
            if cover_matches(query, &seen) {
                last = i32::from(entry.position & MAX_POSITION);
                cover.end = i;
                last_index = i;
                found = true;
                break;
            }
        }
        if !found {
            return false;
        }
        let mut seen = Vec::new();
        let mut stop = cover.next;
        for i in (cover.next..=last_index).rev() {
            stop = i;
            see(&mut seen, &doc[i]);
            if cover_matches(query, &seen) {
                if i32::from(doc[i].position & MAX_POSITION) < first {
                    cover.begin = i;
                    first = i32::from(doc[i].position & MAX_POSITION);
                }
                break;
            }
        }
        cover.first = first;
        cover.last = last;
        if first <= last {
            cover.next = stop + 1;
            return true;
        }
        cover.next += 1;
    }
}

/// rank_cover is Postgres' calc_rank_cd, which ranks by the stretches of the vector that match the query, favoring
/// short ones.
fn rank_cover(weights: &[f32; 4], lexemes: &[Lexeme], query: &Query, method: i32) -> f32 {
    let inverse: Vec<f64> = weights.iter().map(|&w| 1.0 / f64::from(w)).collect();
    let doc = document(lexemes, query);
    if doc.is_empty() {
        return 0.0;
    }
    let (mut wdoc, mut sum_distance, mut previous, mut extents) = (0.0f64, 0.0f64, 0.0f64, 0);
    let mut cover = Cover::default();
    while next_cover(&doc, query, &mut cover) {
        let inverse_sum: f64 =
            doc[cover.begin..=cover.end].iter().map(|e| inverse[usize::from(e.position >> 14)]).sum();
        let length = (cover.end - cover.begin) as i32;
        let cpos = f64::from(length + 1) / inverse_sum;
        let mut noise = (cover.last - cover.first) - length;
        if noise < 0 {
            noise = length / 2;
        }
        wdoc += cpos / f64::from(1 + noise);
        let position = f64::from(cover.last + cover.first) / 2.0;
        if extents > 0 && position > previous {
            sum_distance += 1.0 / (position - previous);
        }
        previous = position;
        extents += 1;
    }
    if method & 0x01 != 0 {
        wdoc /= (word_count(lexemes) as f64 + 1.0).ln();
    }
    if method & 0x02 != 0 {
        let len = word_count(lexemes);
        if len > 0 {
            wdoc /= len as f64;
        }
    }
    if method & 0x04 != 0 && extents > 0 && sum_distance > 0.0 {
        wdoc /= f64::from(extents) / sum_distance;
    }
    if method & 0x08 != 0 {
        wdoc /= lexemes.len() as f64;
    }
    if method & 0x10 != 0 {
        wdoc /= (lexemes.len() as f64 + 1.0).ln() / 2f64.ln();
    }
    if method & 0x20 != 0 {
        wdoc /= wdoc + 1.0;
    }
    wdoc as f32
}

/// text_argument returns the text of a text argument.
fn text_argument(value: &Value) -> String {
    value.output().unwrap_or_default()
}

/// config returns the configuration that a regconfig argument names, or the default_text_search_config setting's
/// without one.
fn config(ctx: &Ctx<'_>, value: Option<&Value>) -> Result<tsearch::Config> {
    match value {
        Some(Value::Reg(reg)) => tsearch::Config::named(&reg.name),
        _ => tsearch::Config::named(&ctx.session.settings.get("default_text_search_config").unwrap_or_default()),
    }
}

/// to_tsvector parses text into a vector of its lexemes, as Postgres' to_tsvector_byid does.
fn to_tsvector(ctx: &mut Ctx<'_>, config_value: Option<&Value>, text: &Value) -> Result<Value> {
    let config = config(ctx, config_value)?;
    let mut notices = Vec::new();
    let words = tsearch::parse_text(config, &text_argument(text), &mut 0, &mut notices);
    for notice in notices {
        ctx.session.notice(notice);
    }
    Ok(vector(&tsearch::vector_lexemes(words)))
}

/// to_query parses text into a query in a mode, joining the lexemes of each operand with an operator, as Postgres'
/// to_tsquery, plainto_tsquery, phraseto_tsquery, and websearch_to_tsquery do.
fn to_query(
    ctx: &mut Ctx<'_>,
    config_value: Option<&Value>,
    text: &Value,
    mode: Mode,
    join: Operator,
) -> Result<Value> {
    let config = config(ctx, config_value)?;
    let (query, notices) = tsearch::to_tsquery(config, &text_argument(text), mode, join)?;
    for notice in notices {
        ctx.session.notice(notice);
    }
    Ok(query_value(query.as_ref()))
}

/// lexize returns the lexemes that a dictionary makes of a token, as Postgres' ts_lexize does.
fn lexize(_: &mut Ctx<'_>, args: &[Value]) -> Result<Value> {
    let Value::Reg(reg) = &args[0] else { return Err(PgError::internal("a regdictionary that is not one")) };
    let lexemes = tsearch::Dictionary::named(&reg.name)?.lexize(&text_argument(&args[1]));
    let values = lexemes.into_iter().map(Value::Text).collect();
    Ok(Value::Array(Box::new(crate::array::Array::one_dimensional(TEXT, values))))
}

/// The kinds of JSON parts that json_to_tsvector can take text from, as Postgres' JsonToIndex flags name them.
const JSON_KEYS: u8 = 0x01;
const JSON_STRINGS: u8 = 0x02;
const JSON_NUMBERS: u8 = 0x04;
const JSON_BOOLEANS: u8 = 0x08;

/// json_flags reads the kinds of JSON parts that a jsonb array of their names asks for, as Postgres'
/// parse_jsonb_index_flags does.
fn json_flags(value: &Value) -> Result<u8> {
    let hint = "Possible values are: \"string\", \"numeric\", \"boolean\", \"key\", and \"all\".";
    let json = match value {
        Value::Jsonb(json) => json.as_ref(),
        _ => return Err(PgError::internal("flags that are not jsonb")),
    };
    let names = match json {
        Json::Array(names) => names.as_slice(),
        Json::Object(_) => {
            return Err(PgError::new(
                code::INVALID_PARAMETER_VALUE,
                "wrong flag type, only arrays and scalars are allowed",
            ));
        }
        scalar => std::slice::from_ref(scalar),
    };
    let mut flags = 0;
    for name in names {
        let Json::String(name) = name else {
            return Err(PgError {
                hint: Some(hint.into()),
                ..PgError::new(code::INVALID_PARAMETER_VALUE, "flag array element is not a string")
            });
        };
        flags |= match name.to_ascii_lowercase().as_str() {
            "all" => JSON_KEYS | JSON_STRINGS | JSON_NUMBERS | JSON_BOOLEANS,
            "key" => JSON_KEYS,
            "string" => JSON_STRINGS,
            "numeric" => JSON_NUMBERS,
            "boolean" => JSON_BOOLEANS,
            _ => {
                return Err(PgError {
                    hint: Some(hint.into()),
                    ..PgError::new(code::INVALID_PARAMETER_VALUE, format!("wrong flag in flag array: \"{name}\""))
                });
            }
        };
    }
    Ok(flags)
}

/// jsonb_texts collects the texts of a jsonb value's parts of the kinds that flags ask for, in the order Postgres'
/// iterate_jsonb_values visits them.
fn jsonb_texts(json: &Json, flags: u8, out: &mut Vec<String>) {
    match json {
        Json::Object(pairs) => {
            for (key, value) in pairs {
                if flags & JSON_KEYS != 0 {
                    out.push(key.clone());
                }
                jsonb_texts(value, flags, out);
            }
        }
        Json::Array(items) => items.iter().for_each(|item| jsonb_texts(item, flags, out)),
        Json::String(text) if flags & JSON_STRINGS != 0 => out.push(text.clone()),
        Json::Number(number) if flags & JSON_NUMBERS != 0 => out.push(number.to_string()),
        Json::Bool(b) if flags & JSON_BOOLEANS != 0 => out.push(b.to_string()),
        _ => {}
    }
}

/// json_texts collects the texts of a json value's parts of the kinds that flags ask for, in the order Postgres'
/// iterate_json_values visits them, keeping numbers as written.
fn json_texts(raw: &Raw<'_>, flags: u8, out: &mut Vec<String>) {
    match &raw.kind {
        RawKind::Object(pairs) => {
            for (key, value) in pairs {
                if flags & JSON_KEYS != 0 {
                    out.push(key.clone());
                }
                json_texts(value, flags, out);
            }
        }
        RawKind::Array(items) => items.iter().for_each(|item| json_texts(item, flags, out)),
        RawKind::Scalar(Json::String(text)) if flags & JSON_STRINGS != 0 => out.push(text.clone()),
        RawKind::Scalar(Json::Number(_)) if flags & JSON_NUMBERS != 0 => out.push(raw.text.to_string()),
        RawKind::Scalar(Json::Bool(b)) if flags & JSON_BOOLEANS != 0 => out.push(b.to_string()),
        _ => {}
    }
}

/// json_vector parses the parts of a json or jsonb value that flags ask for, or its strings without flags, into one
/// vector, leaving a position free between the parts as Postgres' add_to_tsvector does.
fn json_vector(ctx: &mut Ctx<'_>, config_value: Option<&Value>, value: &Value, flags: Option<&Value>) -> Result<Value> {
    let config = config(ctx, config_value)?;
    let flags = match flags {
        Some(flags) => json_flags(flags)?,
        None => JSON_STRINGS,
    };
    let mut texts = Vec::new();
    match value {
        Value::Jsonb(json) => jsonb_texts(json, flags, &mut texts),
        Value::Json(text) => json_texts(&crate::json::parse_raw(text)?, flags, &mut texts),
        _ => return Err(PgError::internal("a json value that is not one")),
    }
    let (mut words, mut position, mut notices) = (Vec::new(), 0, Vec::new());
    for text in texts {
        let parsed = tsearch::parse_text(config, &text, &mut position, &mut notices);
        if !parsed.is_empty() {
            position += 1;
        }
        words.extend(parsed);
    }
    for notice in notices {
        ctx.session.notice(notice);
    }
    Ok(vector(&tsearch::vector_lexemes(words)))
}

/// check_parser fails as Postgres does unless a parser name or OID names the default parser, the only one there is.
fn check_parser(value: &Value) -> Result<()> {
    let found = match value {
        Value::Oid(oid) => *oid == DEFAULT_PARSER,
        other => {
            let name = text_argument(other);
            name.strip_prefix("pg_catalog.").unwrap_or(&name) == "default"
        }
    };
    match found {
        true => Ok(()),
        false => Err(PgError::new(
            code::UNDEFINED_OBJECT,
            format!("text search parser \"{}\" does not exist", text_argument(value)),
        )),
    }
}

/// token_types returns the default parser's token types, as Postgres' ts_token_type does.
fn token_types(parser_value: &Value) -> Result<Value> {
    check_parser(parser_value)?;
    let rows = tsearch::parser::TOKEN_TYPES
        .iter()
        .zip(1..)
        .map(|((alias, description), id)| {
            Value::Record(vec![Value::Int4(id), Value::Text(alias.to_string()), Value::Text(description.to_string())])
        })
        .collect();
    Ok(Value::Set(rows))
}

/// parse returns the default parser's tokens of text, as Postgres' ts_parse does.
fn parse(parser_value: &Value, text: &Value) -> Result<Value> {
    check_parser(parser_value)?;
    let text = text_argument(text);
    let mut parser = tsearch::parser::Parser::new(&text);
    let mut rows = Vec::new();
    while let Some((token_type, token)) = parser.next_token() {
        rows.push(Value::Record(vec![Value::Int4(i32::from(token_type)), Value::Text(token.to_string())]));
    }
    Ok(Value::Set(rows))
}

/// debug returns each token of text with its type, the dictionaries its type goes to, and the lexemes the first of
/// them makes of it, as Postgres' ts_debug does.
fn debug(ctx: &mut Ctx<'_>, config_value: Option<&Value>, text: &Value) -> Result<Value> {
    let config = config(ctx, config_value)?;
    let text = text_argument(text);
    let mut parser = tsearch::parser::Parser::new(&text);
    let mut rows = Vec::new();
    while let Some((token_type, token)) = parser.next_token() {
        let (alias, description) = tsearch::parser::TOKEN_TYPES[usize::from(token_type) - 1];
        let dictionaries = config.dictionaries(token_type);
        let mut names = Vec::new();
        for dictionary in dictionaries {
            names.push(ctx.reg_value(Value::Text(dictionary.name().to_string()), REGDICTIONARY)?);
        }
        let (dictionary, lexemes) = match (names.first(), dictionaries.first()) {
            (Some(name), Some(dictionary)) => {
                let lexemes = dictionary.lexize(token).into_iter().map(Value::Text).collect();
                (name.clone(), Value::Array(Box::new(crate::array::Array::one_dimensional(TEXT, lexemes))))
            }
            _ => (Value::Null, Value::Null),
        };
        rows.push(Value::Record(vec![
            Value::Text(alias.to_string()),
            Value::Text(description.to_string()),
            Value::Text(token.to_string()),
            Value::Array(Box::new(crate::array::Array::one_dimensional(REGDICTIONARY, names))),
            dictionary,
            lexemes,
        ]));
    }
    Ok(Value::Set(rows))
}
