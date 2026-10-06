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

//! The vector extension, version 0.8.6: the vector, halfvec, and sparsevec types, their functions and operators, and
//! the hnsw and ivfflat index access methods.

use std::cmp::Ordering;

use super::{
    AccessMethod, Aggregate, BaseType, Cast, Control, Extension, Implementation, Operator, OperatorClass, Routine,
};
use crate::array::Array;
use crate::catalog::ColumnType;
use crate::error::{PgError, Result, code};
use crate::numeric::Numeric;
use crate::query::Ctx;
use crate::types::{BaseValue, Value};

/// MAX_DENSE_DIMS is the most dimensions a vector or halfvec has.
const MAX_DENSE_DIMS: usize = 16000;

/// MAX_SPARSE_DIMS is the most dimensions a sparsevec has.
const MAX_SPARSE_DIMS: i64 = 1_000_000_000;

/// MAX_SPARSE_NON_ZERO is the most non-zero elements a sparsevec has.
const MAX_SPARSE_NON_ZERO: usize = 16000;

/// extension returns the definition of the extension.
pub fn extension() -> Extension {
    Extension {
        name: "vector",
        control: Control {
            default_version: "0.8.6",
            comment: "vector data type and ivfflat and hnsw access methods",
            superuser: true,
            trusted: false,
            relocatable: true,
        },
        types: vec![
            BaseType {
                name: "vector",
                input: dense_input::<false>,
                output: dense_output::<false>,
                receive: dense_receive::<false>,
                send: dense_send::<false>,
                typmod_in: |modifiers| typmod_in("vector", MAX_DENSE_DIMS as i64, modifiers),
                typmod: dense_typmod::<false>,
                compare: dense_compare::<false>,
            },
            BaseType {
                name: "halfvec",
                input: dense_input::<true>,
                output: dense_output::<true>,
                receive: dense_receive::<true>,
                send: dense_send::<true>,
                typmod_in: |modifiers| typmod_in("halfvec", MAX_DENSE_DIMS as i64, modifiers),
                typmod: dense_typmod::<true>,
                compare: dense_compare::<true>,
            },
            BaseType {
                name: "sparsevec",
                input: sparse_input,
                output: sparse_output,
                receive: sparse_receive,
                send: sparse_send,
                typmod_in: |modifiers| typmod_in("sparsevec", MAX_SPARSE_DIMS, modifiers),
                typmod: sparse_typmod,
                compare: sparse_compare,
            },
        ],
        routines: routines(),
        operators: operators(),
        casts: casts(),
        aggregates: aggregates(),
        operator_classes: operator_classes(),
        access_methods: vec![
            AccessMethod { name: "hnsw", handler: "hnswhandler" },
            AccessMethod { name: "ivfflat", handler: "ivfflathandler" },
        ],
    }
}

/// routine declares a strict routine whose parameters are unnamed.
fn routine(name: &str, symbol: &str, params: &[&str], returns: &str, implementation: Implementation) -> Routine {
    Routine {
        name: name.into(),
        symbol: symbol.into(),
        params: params.iter().map(|p| ("", p.to_string())).collect(),
        returns: returns.into(),
        strict: true,
        implementation,
    }
}

/// io_routines declares the support routines of a type.
fn io_routines(name: &str, input: Implementation, output: Implementation, typmod: Implementation) -> Vec<Routine> {
    vec![
        routine(&format!("{name}_in"), &format!("{name}_in"), &["cstring", "oid", "int4"], name, input),
        routine(&format!("{name}_out"), &format!("{name}_out"), &[name], "cstring", output),
        routine(&format!("{name}_typmod_in"), &format!("{name}_typmod_in"), &["_cstring"], "int4", typmod),
        routine(&format!("{name}_recv"), &format!("{name}_recv"), &["internal", "oid", "int4"], name, receive),
        routine(&format!("{name}_send"), &format!("{name}_send"), &[name], "bytea", send),
    ]
}

/// comparison_routines declares the comparison functions of a type.
fn comparison_routines(name: &str) -> Vec<Routine> {
    let compare: [(&str, Implementation); 7] = [
        ("_lt", |_, a, _| compare_values(a).map(|o| Value::Bool(o.is_lt()))),
        ("_le", |_, a, _| compare_values(a).map(|o| Value::Bool(o.is_le()))),
        ("_eq", |_, a, _| compare_values(a).map(|o| Value::Bool(o.is_eq()))),
        ("_ne", |_, a, _| compare_values(a).map(|o| Value::Bool(o.is_ne()))),
        ("_ge", |_, a, _| compare_values(a).map(|o| Value::Bool(o.is_ge()))),
        ("_gt", |_, a, _| compare_values(a).map(|o| Value::Bool(o.is_gt()))),
        ("_cmp", |_, a, _| compare_values(a).map(|o| Value::Int4(o as i32))),
    ];
    compare
        .into_iter()
        .map(|(suffix, implementation)| {
            let name = format!("{name}{suffix}");
            let returns = if suffix == "_cmp" { "int4" } else { "bool" };
            routine(
                &name,
                &name,
                &[name.trim_end_matches(suffix), name.trim_end_matches(suffix)],
                returns,
                implementation,
            )
        })
        .collect()
}

/// array_routines declares the conversions from arrays to a type, which share a C symbol in pgvector and so take their
/// source type's name as a suffix.
fn array_routines(name: &str, implementation: Implementation) -> Vec<Routine> {
    ["_int4", "_float4", "_float8", "_numeric"]
        .into_iter()
        .map(|source| {
            routine(
                &format!("array_to_{name}"),
                &format!("array_to_{name}{source}"),
                &[source, "int4", "bool"],
                name,
                implementation,
            )
        })
        .collect()
}

/// routines returns every routine of the extension.
fn routines() -> Vec<Routine> {
    let mut all = io_routines("vector", type_input, type_output, vector_typmod_in);
    let vector = |name: &str, params: &[&str], returns: &str, implementation: Implementation| {
        routine(name, name, params, returns, implementation)
    };
    all.extend([
        vector("binary_quantize", &["vector"], "bit", binary_quantize::<false>),
        vector("cosine_distance", &["vector", "vector"], "float8", cosine_distance::<false>),
        vector("inner_product", &["vector", "vector"], "float8", inner_product::<false>),
        vector("l1_distance", &["vector", "vector"], "float8", l1_distance::<false>),
        vector("l2_distance", &["vector", "vector"], "float8", l2_distance::<false>),
        vector("l2_normalize", &["vector"], "vector", l2_normalize::<false>),
        vector("subvector", &["vector", "int4", "int4"], "vector", subvector::<false>),
        vector("vector_add", &["vector", "vector"], "vector", add::<false>),
        vector("vector_concat", &["vector", "vector"], "vector", concat::<false>),
        vector("vector_dims", &["vector"], "int4", dims::<false>),
        vector("vector_mul", &["vector", "vector"], "vector", mul::<false>),
        vector("vector_norm", &["vector"], "float8", norm::<false>),
        vector("vector_sub", &["vector", "vector"], "vector", sub::<false>),
    ]);
    all.extend(comparison_routines("vector"));
    all.extend([
        vector("vector", &["vector", "int4", "bool"], "vector", dense_with_typmod::<false>),
        vector("vector_accum", &["_float8", "vector"], "_float8", accum::<false>),
        vector("vector_avg", &["_float8"], "vector", avg::<false>),
        vector("vector_combine", &["_float8", "_float8"], "_float8", combine::<false>),
        vector("vector_l2_squared_distance", &["vector", "vector"], "float8", l2_squared_distance::<false>),
        vector("vector_negative_inner_product", &["vector", "vector"], "float8", negative_inner_product::<false>),
        vector("vector_spherical_distance", &["vector", "vector"], "float8", spherical_distance::<false>),
        vector("vector_to_float4", &["vector", "int4", "bool"], "_float4", dense_to_float4::<false>),
    ]);
    all.extend(array_routines("vector", array_to_dense::<false>));

    all.extend(io_routines("halfvec", type_input, type_output, halfvec_typmod_in));
    let half = |name: &str, symbol: &str, params: &[&str], returns: &str, implementation: Implementation| {
        routine(name, symbol, params, returns, implementation)
    };
    all.extend([
        half("binary_quantize", "halfvec_binary_quantize", &["halfvec"], "bit", binary_quantize::<true>),
        half("cosine_distance", "halfvec_cosine_distance", &["halfvec", "halfvec"], "float8", cosine_distance::<true>),
        half("halfvec_add", "halfvec_add", &["halfvec", "halfvec"], "halfvec", add::<true>),
        half("halfvec_concat", "halfvec_concat", &["halfvec", "halfvec"], "halfvec", concat::<true>),
        half("halfvec_mul", "halfvec_mul", &["halfvec", "halfvec"], "halfvec", mul::<true>),
        half("halfvec_sub", "halfvec_sub", &["halfvec", "halfvec"], "halfvec", sub::<true>),
        half("inner_product", "halfvec_inner_product", &["halfvec", "halfvec"], "float8", inner_product::<true>),
        half("l1_distance", "halfvec_l1_distance", &["halfvec", "halfvec"], "float8", l1_distance::<true>),
        half("l2_distance", "halfvec_l2_distance", &["halfvec", "halfvec"], "float8", l2_distance::<true>),
        half("l2_norm", "halfvec_l2_norm", &["halfvec"], "float8", norm::<true>),
        half("l2_normalize", "halfvec_l2_normalize", &["halfvec"], "halfvec", l2_normalize::<true>),
        half("subvector", "halfvec_subvector", &["halfvec", "int4", "int4"], "halfvec", subvector::<true>),
        half("vector_dims", "halfvec_vector_dims", &["halfvec"], "int4", dims::<true>),
    ]);
    all.extend(comparison_routines("halfvec"));
    let same = |name: &str, params: &[&str], returns: &str, implementation: Implementation| {
        routine(name, name, params, returns, implementation)
    };
    all.extend([
        same("halfvec", &["halfvec", "int4", "bool"], "halfvec", dense_with_typmod::<true>),
        same("halfvec_accum", &["_float8", "halfvec"], "_float8", accum::<true>),
        same("halfvec_avg", &["_float8"], "halfvec", avg::<true>),
        same("halfvec_combine", &["_float8", "_float8"], "_float8", combine::<true>),
        same("halfvec_l2_squared_distance", &["halfvec", "halfvec"], "float8", l2_squared_distance::<true>),
        same("halfvec_negative_inner_product", &["halfvec", "halfvec"], "float8", negative_inner_product::<true>),
        same("halfvec_spherical_distance", &["halfvec", "halfvec"], "float8", spherical_distance::<true>),
        same("halfvec_to_float4", &["halfvec", "int4", "bool"], "_float4", dense_to_float4::<true>),
        same("halfvec_to_vector", &["halfvec", "int4", "bool"], "vector", dense_to_dense::<false>),
        same("vector_to_halfvec", &["vector", "int4", "bool"], "halfvec", dense_to_dense::<true>),
    ]);
    all.extend(array_routines("halfvec", array_to_dense::<true>));

    all.extend([
        same("hamming_distance", &["bit", "bit"], "float8", hamming_distance),
        same("jaccard_distance", &["bit", "bit"], "float8", jaccard_distance),
    ]);

    all.extend(io_routines("sparsevec", type_input, type_output, sparsevec_typmod_in));
    all.extend([
        half(
            "cosine_distance",
            "sparsevec_cosine_distance",
            &["sparsevec", "sparsevec"],
            "float8",
            sparse_cosine_distance,
        ),
        half("inner_product", "sparsevec_inner_product", &["sparsevec", "sparsevec"], "float8", sparse_inner_product),
        half("l1_distance", "sparsevec_l1_distance", &["sparsevec", "sparsevec"], "float8", sparse_l1_distance),
        half("l2_distance", "sparsevec_l2_distance", &["sparsevec", "sparsevec"], "float8", sparse_l2_distance),
        half("l2_norm", "sparsevec_l2_norm", &["sparsevec"], "float8", sparse_norm),
        half("l2_normalize", "sparsevec_l2_normalize", &["sparsevec"], "sparsevec", sparse_l2_normalize),
    ]);
    all.extend(comparison_routines("sparsevec"));
    all.extend([
        same("halfvec_to_sparsevec", &["halfvec", "int4", "bool"], "sparsevec", dense_to_sparse::<true>),
        same("sparsevec", &["sparsevec", "int4", "bool"], "sparsevec", sparse_with_typmod),
        same("sparsevec_l2_squared_distance", &["sparsevec", "sparsevec"], "float8", sparse_l2_squared_distance),
        same("sparsevec_negative_inner_product", &["sparsevec", "sparsevec"], "float8", sparse_negative_inner_product),
        same("sparsevec_to_halfvec", &["sparsevec", "int4", "bool"], "halfvec", sparse_to_dense::<true>),
        same("sparsevec_to_vector", &["sparsevec", "int4", "bool"], "vector", sparse_to_dense::<false>),
        same("vector_to_sparsevec", &["vector", "int4", "bool"], "sparsevec", dense_to_sparse::<false>),
    ]);
    all.extend(array_routines("sparsevec", array_to_sparse));
    all
}

/// operator declares an operator over two values of the same type.
fn operator(name: &'static str, ty: &str, routine: &str, commutator: &'static str, negator: &'static str) -> Operator {
    Operator { name, left: ty.into(), right: ty.into(), routine: routine.into(), commutator, negator }
}

/// distance_operators declares the four distance operators of a type.
fn distance_operators(ty: &str, l2: &str, negative_inner_product: &str, cosine: &str, l1: &str) -> Vec<Operator> {
    vec![
        operator("<->", ty, l2, "<->", ""),
        operator("<#>", ty, negative_inner_product, "<#>", ""),
        operator("<=>", ty, cosine, "<=>", ""),
        operator("<+>", ty, l1, "<+>", ""),
    ]
}

/// arithmetic_operators declares the arithmetic and concatenation operators of a dense type.
fn arithmetic_operators(ty: &str) -> Vec<Operator> {
    vec![
        operator("+", ty, &format!("{ty}_add"), "+", ""),
        operator("-", ty, &format!("{ty}_sub"), "", ""),
        operator("*", ty, &format!("{ty}_mul"), "*", ""),
        operator("||", ty, &format!("{ty}_concat"), "", ""),
    ]
}

/// comparison_operators declares the six comparison operators of a type.
fn comparison_operators(ty: &str) -> Vec<Operator> {
    vec![
        operator("<", ty, &format!("{ty}_lt"), ">", ">="),
        operator("<=", ty, &format!("{ty}_le"), ">=", ">"),
        operator("=", ty, &format!("{ty}_eq"), "=", "<>"),
        operator("<>", ty, &format!("{ty}_ne"), "<>", "="),
        operator(">=", ty, &format!("{ty}_ge"), "<=", "<"),
        operator(">", ty, &format!("{ty}_gt"), "<", "<="),
    ]
}

/// operators returns every operator of the extension.
fn operators() -> Vec<Operator> {
    let mut all =
        distance_operators("vector", "l2_distance", "vector_negative_inner_product", "cosine_distance", "l1_distance");
    all.extend(arithmetic_operators("vector"));
    all.extend(comparison_operators("vector"));
    all.extend(distance_operators(
        "halfvec",
        "halfvec_l2_distance",
        "halfvec_negative_inner_product",
        "halfvec_cosine_distance",
        "halfvec_l1_distance",
    ));
    all.extend(arithmetic_operators("halfvec"));
    all.extend(comparison_operators("halfvec"));
    all.push(operator("<~>", "bit", "hamming_distance", "<~>", ""));
    all.push(operator("<%>", "bit", "jaccard_distance", "<%>", ""));
    all.extend(distance_operators(
        "sparsevec",
        "sparsevec_l2_distance",
        "sparsevec_negative_inner_product",
        "sparsevec_cosine_distance",
        "sparsevec_l1_distance",
    ));
    all.extend(comparison_operators("sparsevec"));
    all
}

/// casts returns every cast of the extension.
fn casts() -> Vec<Cast> {
    let cast = |source: &str, target: &str, routine: &str, context: u8| Cast {
        source: source.into(),
        target: target.into(),
        routine: routine.into(),
        context,
    };
    let mut all = vec![
        cast("halfvec", "_float4", "halfvec_to_float4", crate::casts::ASSIGNMENT),
        cast("halfvec", "halfvec", "halfvec", crate::casts::IMPLICIT),
        cast("halfvec", "sparsevec", "halfvec_to_sparsevec", crate::casts::IMPLICIT),
        cast("halfvec", "vector", "halfvec_to_vector", crate::casts::ASSIGNMENT),
        cast("sparsevec", "halfvec", "sparsevec_to_halfvec", crate::casts::ASSIGNMENT),
        cast("sparsevec", "sparsevec", "sparsevec", crate::casts::IMPLICIT),
        cast("sparsevec", "vector", "sparsevec_to_vector", crate::casts::ASSIGNMENT),
        cast("vector", "_float4", "vector_to_float4", crate::casts::IMPLICIT),
        cast("vector", "halfvec", "vector_to_halfvec", crate::casts::IMPLICIT),
        cast("vector", "sparsevec", "vector_to_sparsevec", crate::casts::IMPLICIT),
        cast("vector", "vector", "vector", crate::casts::IMPLICIT),
    ];
    for target in ["vector", "halfvec", "sparsevec"] {
        for source in ["_int4", "_float4", "_float8", "_numeric"] {
            all.push(cast(source, target, &format!("array_to_{target}{source}"), crate::casts::ASSIGNMENT));
        }
    }
    all
}

/// aggregates returns the avg and sum aggregates of the dense types.
fn aggregates() -> Vec<Aggregate> {
    let mut all = Vec::new();
    for ty in ["vector", "halfvec"] {
        all.push(Aggregate {
            name: "avg",
            params: vec![ty.into()],
            returns: ty.into(),
            state_type: "_float8".into(),
            transition: format!("{ty}_accum"),
            final_routine: format!("{ty}_avg"),
            combine: format!("{ty}_combine"),
            init_cond: Some("{0}"),
        });
        all.push(Aggregate {
            name: "sum",
            params: vec![ty.into()],
            returns: ty.into(),
            state_type: ty.into(),
            transition: format!("{ty}_add"),
            final_routine: String::new(),
            combine: format!("{ty}_add"),
            init_cond: None,
        });
    }
    all
}

/// operator_classes returns the extension's operator classes.
fn operator_classes() -> Vec<OperatorClass> {
    let mut all = Vec::new();
    for ty in ["vector", "halfvec"] {
        for metric in ["_l2_ops", "_ip_ops", "_cosine_ops", "_l1_ops"] {
            let name = format!("{ty}{metric}");
            let default_for = if name == "vector_l2_ops" { vec!["ivfflat"] } else { Vec::new() };
            let access_methods = if metric == "_l1_ops" { vec!["hnsw"] } else { vec!["hnsw", "ivfflat"] };
            all.push(OperatorClass { name, access_methods, default_for, type_name: ty });
        }
    }
    for metric in ["_l2_ops", "_ip_ops", "_cosine_ops", "_l1_ops"] {
        all.push(OperatorClass {
            name: format!("sparsevec{metric}"),
            access_methods: vec!["hnsw"],
            default_for: Vec::new(),
            type_name: "sparsevec",
        });
    }
    all.push(OperatorClass {
        name: "bit_hamming_ops".into(),
        access_methods: vec!["hnsw", "ivfflat"],
        default_for: Vec::new(),
        type_name: "bit",
    });
    all.push(OperatorClass {
        name: "bit_jaccard_ops".into(),
        access_methods: vec!["hnsw"],
        default_for: Vec::new(),
        type_name: "bit",
    });
    all
}

/// data_error returns pgvector's error for invalid data.
fn data_error(message: impl Into<String>) -> PgError {
    PgError::new(code::DATA_EXCEPTION, message)
}

/// invalid_syntax returns pgvector's error for text that is not a value of the type.
fn invalid_syntax(type_name: &str, text: &str) -> PgError {
    PgError::new(code::INVALID_TEXT_REPRESENTATION, format!("invalid input syntax for type {type_name}: \"{text}\""))
}

/// out_of_range returns Postgres' error for an overflowing or underflowing result.
fn out_of_range(kind: &str) -> PgError {
    PgError::new(code::NUMERIC_VALUE_OUT_OF_RANGE, format!("value out of range: {kind}"))
}

/// at_least_one_dimension returns pgvector's error for a vector without dimensions.
fn at_least_one_dimension(type_name: &str) -> PgError {
    data_error(format!("{type_name} must have at least 1 dimension"))
}

/// too_many_dimensions returns pgvector's error for a vector over its type's dimension limit.
fn too_many_dimensions(type_name: &str, limit: i64) -> PgError {
    PgError::new(code::PROGRAM_LIMIT_EXCEEDED, format!("{type_name} cannot have more than {limit} dimensions"))
}

/// check_expected_dimensions checks a dimension count against a type modifier.
fn check_expected_dimensions(typmod: i32, dimensions: usize) -> Result<()> {
    if typmod != -1 && typmod as i64 != dimensions as i64 {
        return Err(data_error(format!("expected {typmod} dimensions, not {dimensions}")));
    }
    Ok(())
}

/// check_dense_dimensions checks that a dense vector has between 1 and the most dimensions.
fn check_dense_dimensions(type_name: &str, dimensions: usize) -> Result<()> {
    if dimensions < 1 {
        return Err(at_least_one_dimension(type_name));
    }
    if dimensions > MAX_DENSE_DIMS {
        return Err(too_many_dimensions(type_name, MAX_DENSE_DIMS as i64));
    }
    Ok(())
}

/// check_element rejects the NaN and infinite elements that pgvector disallows.
fn check_element(type_name: &str, value: f32) -> Result<()> {
    if value.is_nan() {
        return Err(data_error(format!("NaN not allowed in {type_name}")));
    }
    if value.is_infinite() {
        return Err(data_error(format!("infinite value not allowed in {type_name}")));
    }
    Ok(())
}

/// typmod_in computes a vector type's modifier: its one dimension count.
fn typmod_in(type_name: &str, limit: i64, modifiers: &[String]) -> Result<i32> {
    let [modifier] = modifiers else {
        return Err(PgError::new(code::INVALID_PARAMETER_VALUE, "invalid type modifier"));
    };
    let dimensions: i64 = match modifier.trim().parse::<i64>() {
        Ok(d) if i32::try_from(d).is_ok() => d,
        Ok(_) => {
            return Err(PgError::new(
                code::NUMERIC_VALUE_OUT_OF_RANGE,
                format!("value \"{modifier}\" is out of range for type integer"),
            ));
        }
        Err(_) => return Err(crate::cast::invalid_syntax(crate::oid::INT4, modifier)),
    };
    if dimensions < 1 {
        return Err(PgError::new(
            code::INVALID_PARAMETER_VALUE,
            format!("dimensions for type {type_name} must be at least 1"),
        ));
    }
    if dimensions > limit {
        return Err(PgError::new(
            code::INVALID_PARAMETER_VALUE,
            format!("dimensions for type {type_name} cannot exceed {limit}"),
        ));
    }
    Ok(dimensions as i32)
}

/// strtof reads the longest prefix of text that is a float as C's strtof does, returning the value, the bytes read,
/// and whether the value overflowed or underflowed, or None when no prefix is a float.
fn strtof(text: &str) -> Option<(f32, usize, bool)> {
    let bytes = text.as_bytes();
    let mut i = usize::from(matches!(bytes.first(), Some(b'+' | b'-')));
    let rest = text[i..].to_ascii_lowercase();
    for word in ["infinity", "inf", "nan"] {
        if rest.starts_with(word) {
            let value = if word == "nan" { f32::NAN } else { f32::INFINITY };
            let value = if bytes[0] == b'-' { -value } else { value };
            return Some((value, i + word.len(), false));
        }
    }
    if rest.starts_with("0x") && rest[2..].starts_with(|c: char| c.is_ascii_hexdigit() || c == '.') {
        let digits = &rest[2..];
        let mantissa_end = digits.find(|c: char| !(c.is_ascii_hexdigit() || c == '.')).unwrap_or(digits.len());
        let (whole, fraction) = digits[..mantissa_end].split_once('.').unwrap_or((&digits[..mantissa_end], ""));
        let mut value = 0f64;
        for c in whole.chars().chain(fraction.chars()) {
            value = value * 16.0 + c.to_digit(16).unwrap_or(0) as f64;
        }
        value /= 16f64.powi(fraction.len() as i32);
        let mut end = 2 + mantissa_end;
        let exponent = &rest[end..];
        if let Some(digits) = exponent.strip_prefix('p') {
            let sign = usize::from(digits.starts_with(['+', '-']));
            let count = digits[sign..].bytes().take_while(u8::is_ascii_digit).count();
            if count > 0 {
                let power: i32 = digits[..sign + count].parse().unwrap_or(0);
                value *= 2f64.powi(power);
                end += 1 + sign + count;
            }
        }
        let value = if bytes[0] == b'-' { -value } else { value };
        let narrowed = value as f32;
        return Some((narrowed, i + end, narrowed.is_infinite() || (narrowed == 0.0 && value != 0.0)));
    }
    let start = i;
    let digits = |i: &mut usize| {
        let begin = *i;
        while bytes.get(*i).is_some_and(u8::is_ascii_digit) {
            *i += 1;
        }
        *i - begin
    };
    let mut count = digits(&mut i);
    if bytes.get(i) == Some(&b'.') {
        i += 1;
        count += digits(&mut i);
    }
    if count == 0 {
        return None;
    }
    if matches!(bytes.get(i), Some(b'e' | b'E')) {
        let mut j = i + 1;
        if matches!(bytes.get(j), Some(b'+' | b'-')) {
            j += 1;
        }
        if digits(&mut j) > 0 {
            i = j;
        }
    }
    let token = &text[..i];
    let value: f32 = token.parse().ok()?;
    let mantissa = &text[start..i];
    let nonzero = mantissa.split(['e', 'E']).next().is_some_and(|m| m.bytes().any(|b| (b'1'..=b'9').contains(&b)));
    Some((value, i, value.is_infinite() || (value == 0.0 && nonzero)))
}

/// is_space reports whether a byte is whitespace as pgvector's parsers see it.
fn is_space(b: u8) -> bool {
    matches!(b, b' ' | b'\t' | b'\n' | b'\r' | b'\x0b' | b'\x0c')
}

/// skip_spaces returns the position of the first byte at or after `i` that is not whitespace.
fn skip_spaces(bytes: &[u8], mut i: usize) -> usize {
    while bytes.get(i).copied().is_some_and(is_space) {
        i += 1;
    }
    i
}

/// half_bits converts a float to the bits of the nearest half-precision value, rounding to nearest even.
fn half_bits(f: f32) -> u16 {
    let bits = f.to_bits();
    let sign = ((bits >> 16) & 0x8000) as u16;
    let exponent = ((bits >> 23) & 0xff) as i32;
    let mut mantissa = bits & 0x7fffff;
    if exponent == 0xff {
        return sign | if mantissa == 0 { 0x7c00 } else { 0x7e00 };
    }
    let half_exponent = exponent - 127 + 15;
    if half_exponent >= 0x1f {
        return sign | 0x7c00;
    }
    if half_exponent <= 0 {
        if half_exponent < -10 {
            return sign;
        }
        mantissa |= 0x800000;
        let shift = (14 - half_exponent) as u32;
        let mut half = (mantissa >> shift) as u16;
        let round = 1u32 << (shift - 1);
        if mantissa & round != 0 && mantissa & (3 * round - 1) != 0 {
            half += 1;
        }
        return sign | half;
    }
    let mut half = ((half_exponent as u16) << 10) | (mantissa >> 13) as u16;
    if mantissa & 0x1000 != 0 && mantissa & (3 * 0x1000 - 1) != 0 {
        half += 1;
    }
    sign | half
}

/// half_value converts the bits of a half-precision value to the float it represents exactly.
fn half_value(half: u16) -> f32 {
    let sign = ((half & 0x8000) as u32) << 16;
    let exponent = ((half >> 10) & 0x1f) as u32;
    let mantissa = (half & 0x3ff) as u32;
    match exponent {
        0x1f => f32::from_bits(sign | 0x7f800000 | mantissa << 13),
        0 => f32::from_bits(sign | (mantissa as f32 * f32::from_bits(0x33800000)).to_bits()),
        _ => f32::from_bits(sign | (exponent + 112) << 23 | mantissa << 13),
    }
}

/// narrow rounds a float to the precision of a dense type, failing when a finite value overflows it.
fn narrow<const HALF: bool>(value: f32) -> Result<f32> {
    let rounded = if HALF { half_value(half_bits(value)) } else { value };
    if rounded.is_infinite() && !value.is_infinite() || (!HALF && value.is_infinite()) {
        return Err(out_of_range("overflow"));
    }
    Ok(rounded)
}

/// to_half rounds a float to half precision, failing for a finite value outside the half-precision range.
fn to_half(value: f32) -> Result<f32> {
    let rounded = half_value(half_bits(value));
    if rounded.is_infinite() && !value.is_infinite() {
        return Err(PgError::new(
            code::NUMERIC_VALUE_OUT_OF_RANGE,
            format!("\"{}\" is out of range for type halfvec", format_element(value)),
        ));
    }
    Ok(rounded)
}

/// dense_name returns the name of a dense type.
fn dense_name<const HALF: bool>() -> &'static str {
    if HALF { "halfvec" } else { "vector" }
}

/// decode_dense reads the stored elements of a dense value.
fn decode_dense<const HALF: bool>(data: &[u8]) -> Vec<f32> {
    if HALF {
        data.as_chunks::<2>().0.iter().map(|c| half_value(u16::from_le_bytes(*c))).collect()
    } else {
        data.as_chunks::<4>().0.iter().map(|c| f32::from_le_bytes(*c)).collect()
    }
}

/// encode_dense writes the stored form of a dense value's elements.
fn encode_dense<const HALF: bool>(values: &[f32]) -> Vec<u8> {
    if HALF {
        values.iter().flat_map(|v| half_bits(*v).to_le_bytes()).collect()
    } else {
        values.iter().flat_map(|v| v.to_le_bytes()).collect()
    }
}

/// base_data returns the stored bytes of an extension type's value.
fn base_data(value: &Value) -> &[u8] {
    match value {
        Value::Base(base) => &base.data,
        _ => &[],
    }
}

/// dense returns the elements of a dense argument.
fn dense<const HALF: bool>(value: &Value) -> Vec<f32> {
    decode_dense::<HALF>(base_data(value))
}

/// base returns a value of an extension type from its stored bytes.
fn base(ty: ColumnType, data: Vec<u8>) -> Value {
    Value::Base(Box::new(BaseValue { type_oid: ty.oid, data }))
}

/// float returns a float8 result.
fn float(value: f64) -> Result<Value> {
    Ok(Value::Float8(value))
}

/// int returns an integer argument.
fn int(value: &Value) -> i32 {
    match value {
        Value::Int4(i) => *i,
        _ => -1,
    }
}

/// dense_input reads a dense vector from its `[x,y,...]` text format.
fn dense_input<const HALF: bool>(text: &str, typmod: i32) -> Result<Vec<u8>> {
    let name = dense_name::<HALF>();
    let bytes = text.as_bytes();
    let mut i = skip_spaces(bytes, 0);
    if bytes.get(i) != Some(&b'[') {
        return Err(PgError {
            detail: Some("Vector contents must start with \"[\".".into()),
            ..invalid_syntax(name, text)
        });
    }
    i = skip_spaces(bytes, i + 1);
    if bytes.get(i) == Some(&b']') {
        return Err(at_least_one_dimension(name));
    }
    let mut values = Vec::new();
    loop {
        if values.len() == MAX_DENSE_DIMS {
            return Err(too_many_dimensions(name, MAX_DENSE_DIMS as i64));
        }
        i = skip_spaces(bytes, i);
        let (value, length, range_error) = strtof(&text[i..]).ok_or_else(|| invalid_syntax(name, text))?;
        if range_error && value.is_infinite() {
            return Err(PgError::new(
                code::NUMERIC_VALUE_OUT_OF_RANGE,
                format!("\"{}\" is out of range for type {name}", &text[i..i + length]),
            ));
        }
        check_element(name, value)?;
        let value = if HALF {
            let rounded = half_value(half_bits(value));
            if rounded.is_infinite() {
                return Err(PgError::new(
                    code::NUMERIC_VALUE_OUT_OF_RANGE,
                    format!("\"{}\" is out of range for type {name}", &text[i..i + length]),
                ));
            }
            rounded
        } else {
            value
        };
        values.push(value);
        i = skip_spaces(bytes, i + length);
        match bytes.get(i) {
            Some(b',') => i += 1,
            Some(b']') => {
                i += 1;
                break;
            }
            _ => return Err(invalid_syntax(name, text)),
        }
    }
    if skip_spaces(bytes, i) != bytes.len() {
        return Err(PgError { detail: Some("Junk after closing right brace.".into()), ..invalid_syntax(name, text) });
    }
    check_expected_dimensions(typmod, values.len())?;
    Ok(encode_dense::<HALF>(&values))
}

/// format_element prints an element as Postgres prints a real.
fn format_element(value: f32) -> String {
    Value::Float4(value).output().unwrap_or_default()
}

/// dense_output prints a dense vector in its `[x,y,...]` text format.
fn dense_output<const HALF: bool>(data: &[u8]) -> String {
    let elements: Vec<String> = decode_dense::<HALF>(data).into_iter().map(format_element).collect();
    format!("[{}]", elements.join(","))
}

/// dense_receive reads a dense vector from its binary format: the dimensions, an unused field, and the elements.
fn dense_receive<const HALF: bool>(data: &[u8], typmod: i32) -> Result<Vec<u8>> {
    let name = dense_name::<HALF>();
    let insufficient = || PgError::new(code::PROTOCOL_VIOLATION, "insufficient data left in message");
    let header = data.get(..4).ok_or_else(insufficient)?;
    let dimensions = i16::from_be_bytes([header[0], header[1]]);
    let unused = i16::from_be_bytes([header[2], header[3]]);
    check_dense_dimensions(name, dimensions.max(0) as usize)?;
    if unused != 0 {
        return Err(data_error(format!("expected unused to be 0, not {unused}")));
    }
    let width = if HALF { 2 } else { 4 };
    let body = &data[4..];
    if body.len() != dimensions as usize * width {
        return Err(insufficient());
    }
    let values: Vec<f32> = if HALF {
        body.as_chunks::<2>().0.iter().map(|c| half_value(u16::from_be_bytes(*c))).collect()
    } else {
        body.as_chunks::<4>().0.iter().map(|c| f32::from_be_bytes(*c)).collect()
    };
    for value in &values {
        check_element(name, *value)?;
    }
    check_expected_dimensions(typmod, values.len())?;
    Ok(encode_dense::<HALF>(&values))
}

/// dense_send writes a dense vector's binary format.
fn dense_send<const HALF: bool>(data: &[u8]) -> Vec<u8> {
    let values = decode_dense::<HALF>(data);
    let mut out = (values.len() as u16).to_be_bytes().to_vec();
    out.extend_from_slice(&[0, 0]);
    for value in values {
        if HALF {
            out.extend_from_slice(&half_bits(value).to_be_bytes());
        } else {
            out.extend_from_slice(&value.to_be_bytes());
        }
    }
    out
}

/// dense_typmod checks a dense vector's dimensions against a type modifier.
fn dense_typmod<const HALF: bool>(data: &[u8], typmod: i32) -> Result<()> {
    check_expected_dimensions(typmod, data.len() / if HALF { 2 } else { 4 })
}

/// dense_compare orders dense vectors by their elements, with a vector before a longer one that it prefixes.
fn dense_compare<const HALF: bool>(left: &[u8], right: &[u8]) -> Ordering {
    let (left, right) = (decode_dense::<HALF>(left), decode_dense::<HALF>(right));
    for (l, r) in left.iter().zip(&right) {
        if l < r {
            return Ordering::Less;
        }
        if l > r {
            return Ordering::Greater;
        }
    }
    left.len().cmp(&right.len())
}

/// SparseVector is a sparsevec's non-zero elements, by 0-based index in ascending order.
#[derive(Default)]
struct SparseVector {
    dimensions: i32,
    indices: Vec<i32>,
    values: Vec<f32>,
}

impl SparseVector {
    /// decode reads a sparsevec's stored form: the dimensions, the count, the indices, and the values, little-endian.
    fn decode(data: &[u8]) -> SparseVector {
        let int = |i: usize| data.get(i..i + 4).map_or(0, |b| i32::from_le_bytes([b[0], b[1], b[2], b[3]]));
        let count = int(4).max(0) as usize;
        SparseVector {
            dimensions: int(0),
            indices: (0..count).map(|i| int(8 + i * 4)).collect(),
            values: (0..count).map(|i| f32::from_bits(int(8 + (count + i) * 4) as u32)).collect(),
        }
    }

    /// encode writes a sparsevec's stored form.
    fn encode(&self) -> Vec<u8> {
        let mut out = Vec::with_capacity(8 + self.indices.len() * 8);
        out.extend_from_slice(&self.dimensions.to_le_bytes());
        out.extend_from_slice(&(self.indices.len() as i32).to_le_bytes());
        for index in &self.indices {
            out.extend_from_slice(&index.to_le_bytes());
        }
        for value in &self.values {
            out.extend_from_slice(&value.to_le_bytes());
        }
        out
    }

    /// from_dense keeps the non-zero elements of a dense vector.
    fn from_dense(values: &[f32]) -> SparseVector {
        let mut sparse = SparseVector { dimensions: values.len() as i32, ..SparseVector::default() };
        for (i, value) in values.iter().enumerate().filter(|(_, v)| **v != 0.0) {
            sparse.indices.push(i as i32);
            sparse.values.push(*value);
        }
        sparse
    }

    /// to_dense returns every element, filling zeros between the non-zero ones.
    fn to_dense(&self) -> Vec<f32> {
        let mut values = vec![0.0; self.dimensions.max(0) as usize];
        for (index, value) in self.indices.iter().zip(&self.values) {
            values[*index as usize] = *value;
        }
        values
    }
}

/// sparse returns a sparsevec argument.
fn sparse(value: &Value) -> SparseVector {
    SparseVector::decode(base_data(value))
}

/// strtol reads the longest prefix of text that is an integer as C's strtol does, clamping to the 64-bit range,
/// returning the value and the bytes read, or None when no prefix is an integer.
fn strtol(text: &str) -> Option<(i64, usize)> {
    let bytes = text.as_bytes();
    let sign = usize::from(matches!(bytes.first(), Some(b'+' | b'-')));
    let count = bytes[sign..].iter().take_while(|b| b.is_ascii_digit()).count();
    if count == 0 {
        return None;
    }
    let negative = bytes.first() == Some(&b'-');
    let value = text[sign..sign + count]
        .parse::<i64>()
        .map_or(if negative { i64::MIN } else { i64::MAX }, |v| if negative { -v } else { v });
    Some((value, sign + count))
}

/// sparse_input reads a sparsevec from its `{index:value,...}/dimensions` text format, with 1-based indices.
fn sparse_input(text: &str, typmod: i32) -> Result<Vec<u8>> {
    let syntax = || invalid_syntax("sparsevec", text);
    let bytes = text.as_bytes();
    let mut i = skip_spaces(bytes, 0);
    if bytes.get(i) != Some(&b'{') {
        return Err(PgError { detail: Some("Vector contents must start with \"{\".".into()), ..syntax() });
    }
    i = skip_spaces(bytes, i + 1);
    let mut elements: Vec<(i64, f32)> = Vec::new();
    if bytes.get(i) == Some(&b'}') {
        i += 1;
    } else {
        loop {
            if elements.len() == MAX_SPARSE_NON_ZERO {
                return Err(PgError::new(
                    code::PROGRAM_LIMIT_EXCEEDED,
                    format!("sparsevec cannot have more than {MAX_SPARSE_NON_ZERO} non-zero elements"),
                ));
            }
            i = skip_spaces(bytes, i);
            let (index, length) = strtol(&text[i..]).ok_or_else(syntax)?;
            i = skip_spaces(bytes, i + length);
            if bytes.get(i) != Some(&b':') {
                return Err(syntax());
            }
            i = skip_spaces(bytes, i + 1);
            let (value, length, range_error) = strtof(&text[i..]).ok_or_else(syntax)?;
            if range_error {
                return Err(PgError::new(
                    code::NUMERIC_VALUE_OUT_OF_RANGE,
                    format!("\"{}\" is out of range for type sparsevec", &text[i..i + length]),
                ));
            }
            check_element("sparsevec", value)?;
            if value != 0.0 {
                elements.push((index, value));
            }
            i = skip_spaces(bytes, i + length);
            match bytes.get(i) {
                Some(b',') => i += 1,
                Some(b'}') => {
                    i += 1;
                    break;
                }
                _ => return Err(syntax()),
            }
        }
    }
    i = skip_spaces(bytes, i);
    match bytes.get(i) {
        Some(b'/') => {}
        Some(_) => return Err(syntax()),
        None => return Err(PgError { detail: Some("Unexpected end of input.".into()), ..syntax() }),
    }
    i = skip_spaces(bytes, i + 1);
    let (dimensions, length) = strtol(&text[i..]).ok_or_else(syntax)?;
    if skip_spaces(bytes, i + length) != bytes.len() {
        return Err(PgError { detail: Some("Junk after closing.".into()), ..syntax() });
    }
    if dimensions < 1 {
        return Err(at_least_one_dimension("sparsevec"));
    }
    if dimensions > MAX_SPARSE_DIMS {
        return Err(too_many_dimensions("sparsevec", MAX_SPARSE_DIMS));
    }
    check_expected_dimensions(typmod, dimensions as usize)?;
    elements.sort_by_key(|(index, _)| *index);
    let mut sparse = SparseVector { dimensions: dimensions as i32, ..SparseVector::default() };
    for (i, (index, value)) in elements.iter().enumerate() {
        if *index < 1 || *index > dimensions {
            return Err(data_error("sparsevec index out of bounds"));
        }
        if i > 0 && elements[i - 1].0 == *index {
            return Err(data_error("sparsevec indices must not contain duplicates"));
        }
        sparse.indices.push(*index as i32 - 1);
        sparse.values.push(*value);
    }
    Ok(sparse.encode())
}

/// sparse_output prints a sparsevec in its `{index:value,...}/dimensions` text format.
fn sparse_output(data: &[u8]) -> String {
    let sparse = SparseVector::decode(data);
    let elements: Vec<String> = sparse
        .indices
        .iter()
        .zip(&sparse.values)
        .map(|(index, value)| format!("{}:{}", *index as i64 + 1, format_element(*value)))
        .collect();
    format!("{{{}}}/{}", elements.join(","), sparse.dimensions)
}

/// sparse_receive reads a sparsevec from its binary format: the dimensions, the count, an unused field, the indices,
/// and the values.
fn sparse_receive(data: &[u8], typmod: i32) -> Result<Vec<u8>> {
    let insufficient = || PgError::new(code::PROTOCOL_VIOLATION, "insufficient data left in message");
    let int =
        |i: usize| data.get(i..i + 4).map(|b| i32::from_be_bytes([b[0], b[1], b[2], b[3]])).ok_or_else(insufficient);
    let (dimensions, count, unused) = (int(0)?, int(4)?, int(8)?);
    if dimensions < 1 {
        return Err(at_least_one_dimension("sparsevec"));
    }
    if dimensions as i64 > MAX_SPARSE_DIMS {
        return Err(too_many_dimensions("sparsevec", MAX_SPARSE_DIMS));
    }
    if count < 0 || count as usize > MAX_SPARSE_NON_ZERO {
        return Err(PgError::new(
            code::PROGRAM_LIMIT_EXCEEDED,
            format!("sparsevec cannot have more than {MAX_SPARSE_NON_ZERO} non-zero elements"),
        ));
    }
    if unused != 0 {
        return Err(data_error(format!("expected unused to be 0, not {unused}")));
    }
    let count = count as usize;
    if data.len() != 12 + count * 8 {
        return Err(insufficient());
    }
    let mut sparse = SparseVector { dimensions, ..SparseVector::default() };
    for i in 0..count {
        let index = int(12 + i * 4)?;
        if index < 0 || index >= dimensions {
            return Err(data_error("sparsevec index out of bounds"));
        }
        if i > 0 && index <= sparse.indices[i - 1] {
            let message = if index == sparse.indices[i - 1] {
                "sparsevec indices must not contain duplicates"
            } else {
                "sparsevec indices must be in ascending order"
            };
            return Err(data_error(message));
        }
        sparse.indices.push(index);
    }
    for i in 0..count {
        let value = f32::from_bits(int(12 + (count + i) * 4)? as u32);
        check_element("sparsevec", value)?;
        if value == 0.0 {
            return Err(data_error("sparsevec elements must not be zero"));
        }
        sparse.values.push(value);
    }
    check_expected_dimensions(typmod, dimensions as usize)?;
    Ok(sparse.encode())
}

/// sparse_send writes a sparsevec's binary format.
fn sparse_send(data: &[u8]) -> Vec<u8> {
    let sparse = SparseVector::decode(data);
    let mut out = Vec::with_capacity(12 + sparse.indices.len() * 8);
    out.extend_from_slice(&sparse.dimensions.to_be_bytes());
    out.extend_from_slice(&(sparse.indices.len() as i32).to_be_bytes());
    out.extend_from_slice(&0i32.to_be_bytes());
    for index in &sparse.indices {
        out.extend_from_slice(&index.to_be_bytes());
    }
    for value in &sparse.values {
        out.extend_from_slice(&value.to_be_bytes());
    }
    out
}

/// sparse_typmod checks a sparsevec's dimensions against a type modifier.
fn sparse_typmod(data: &[u8], typmod: i32) -> Result<()> {
    check_expected_dimensions(typmod, SparseVector::decode(data).dimensions as usize)
}

/// sparse_compare orders sparsevecs as their dense forms, then by their dimensions.
fn sparse_compare(left: &[u8], right: &[u8]) -> Ordering {
    let (a, b) = (SparseVector::decode(left), SparseVector::decode(right));
    let (mut i, mut j) = (0, 0);
    while i < a.indices.len() || j < b.indices.len() {
        let (mut l, mut r) = (0.0, 0.0);
        if j >= b.indices.len() || (i < a.indices.len() && a.indices[i] < b.indices[j]) {
            l = a.values[i];
            i += 1;
        } else if i >= a.indices.len() || b.indices[j] < a.indices[i] {
            r = b.values[j];
            j += 1;
        } else {
            (l, r) = (a.values[i], b.values[j]);
            i += 1;
            j += 1;
        }
        match l.partial_cmp(&r) {
            Some(Ordering::Equal) | None => {}
            Some(ordering) => return ordering,
        }
    }
    a.dimensions.cmp(&b.dimensions)
}

/// type_of returns the extension type of a routine's value argument or result.
fn type_of(value: &Value) -> Option<&'static super::BaseType> {
    let Value::Base(base) = value else { return None };
    match &crate::usertypes::get(base.type_oid)?.kind {
        crate::usertypes::Kind::Base(definition) => Some(*definition),
        _ => None,
    }
}

/// type_input implements the input routines, which read a value in the type that the second argument names.
fn type_input(_: &mut Ctx<'_>, args: &[Value], ty: ColumnType) -> Result<Value> {
    let Value::Text(text) = &args[0] else { return Ok(Value::Null) };
    let definition = match crate::usertypes::get(ty.oid).map(|t| t.kind.clone()) {
        Some(crate::usertypes::Kind::Base(definition)) => definition,
        _ => return Err(PgError::internal("an input routine of an unknown type")),
    };
    Ok(base(ty, (definition.input)(text, int(&args[2]))?))
}

/// type_output implements the output routines.
fn type_output(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    Ok(Value::Text(type_of(&args[0]).map(|t| (t.output)(base_data(&args[0]))).unwrap_or_default()))
}

/// modifiers returns the modifiers that a type modifier routine receives.
fn modifiers(value: &Value) -> Vec<String> {
    match value {
        Value::Array(array) => array.values.iter().filter_map(Value::output).collect(),
        _ => Vec::new(),
    }
}

/// vector_typmod_in implements vector_typmod_in.
fn vector_typmod_in(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    typmod_in("vector", MAX_DENSE_DIMS as i64, &modifiers(&args[0])).map(Value::Int4)
}

/// halfvec_typmod_in implements halfvec_typmod_in.
fn halfvec_typmod_in(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    typmod_in("halfvec", MAX_DENSE_DIMS as i64, &modifiers(&args[0])).map(Value::Int4)
}

/// sparsevec_typmod_in implements sparsevec_typmod_in.
fn sparsevec_typmod_in(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    typmod_in("sparsevec", MAX_SPARSE_DIMS, &modifiers(&args[0])).map(Value::Int4)
}

/// receive implements the receive routines, which Doltgres only calls for binary parameters.
fn receive(_: &mut Ctx<'_>, _: &[Value], _: ColumnType) -> Result<Value> {
    Err(PgError::unsupported("calling a receive routine"))
}

/// send implements the send routines.
fn send(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    Ok(Value::Bytea(type_of(&args[0]).map(|t| (t.send)(base_data(&args[0]))).unwrap_or_default()))
}

/// compare_values compares the two arguments of a comparison routine.
fn compare_values(args: &[Value]) -> Result<Ordering> {
    let definition = type_of(&args[0]).ok_or_else(|| PgError::internal("a comparison of an unknown type"))?;
    Ok((definition.compare)(base_data(&args[0]), base_data(&args[1])))
}

/// pair returns the elements of two dense arguments, which must have the same dimensions.
fn pair<const HALF: bool>(args: &[Value]) -> Result<(Vec<f32>, Vec<f32>)> {
    let (a, b) = (dense::<HALF>(&args[0]), dense::<HALF>(&args[1]));
    if a.len() != b.len() {
        return Err(data_error(format!("different {} dimensions {} and {}", dense_name::<HALF>(), a.len(), b.len())));
    }
    Ok((a, b))
}

/// l2_squared returns the squared Euclidean distance, accumulating in single precision as pgvector does.
fn l2_squared(a: &[f32], b: &[f32]) -> f32 {
    a.iter().zip(b).fold(0.0, |sum, (x, y)| sum + (x - y) * (x - y))
}

/// dot returns the inner product, accumulating in single precision as pgvector does.
fn dot(a: &[f32], b: &[f32]) -> f32 {
    a.iter().zip(b).fold(0.0, |sum, (x, y)| sum + x * y)
}

/// cosine returns the cosine distance from an inner product and two squared norms, keeping the similarity within
/// [-1, 1] while letting NaN through.
fn cosine(product: f32, norm_a: f32, norm_b: f32) -> f64 {
    let similarity = (product as f64 / (norm_a as f64 * norm_b as f64).sqrt()).clamp(-1.0, 1.0);
    1.0 - similarity
}

/// l2_distance implements l2_distance.
fn l2_distance<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    let (a, b) = pair::<HALF>(args)?;
    float((l2_squared(&a, &b) as f64).sqrt())
}

/// l2_squared_distance implements the squared distance support routines.
fn l2_squared_distance<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    let (a, b) = pair::<HALF>(args)?;
    float(l2_squared(&a, &b) as f64)
}

/// inner_product implements inner_product.
fn inner_product<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    let (a, b) = pair::<HALF>(args)?;
    float(dot(&a, &b) as f64)
}

/// negative_inner_product implements the negative inner product that the `<#>` operator computes.
fn negative_inner_product<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    let (a, b) = pair::<HALF>(args)?;
    float(-(dot(&a, &b) as f64))
}

/// cosine_distance implements cosine_distance.
fn cosine_distance<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    let (a, b) = pair::<HALF>(args)?;
    float(cosine(dot(&a, &b), dot(&a, &a), dot(&b, &b)))
}

/// spherical_distance implements the spherical distance support routines.
fn spherical_distance<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    let (a, b) = pair::<HALF>(args)?;
    float((dot(&a, &b) as f64).clamp(-1.0, 1.0).acos() / std::f64::consts::PI)
}

/// l1_distance implements l1_distance.
fn l1_distance<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    let (a, b) = pair::<HALF>(args)?;
    float(a.iter().zip(&b).fold(0f32, |sum, (x, y)| sum + (x - y).abs()) as f64)
}

/// dims implements vector_dims.
fn dims<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    Ok(Value::Int4(dense::<HALF>(&args[0]).len() as i32))
}

/// dense_norm returns the Euclidean norm, accumulating in double precision.
fn dense_norm(values: &[f32]) -> f64 {
    values.iter().fold(0.0, |sum, v| sum + *v as f64 * *v as f64).sqrt()
}

/// norm implements vector_norm and l2_norm.
fn norm<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    float(dense_norm(&dense::<HALF>(&args[0])))
}

/// l2_normalize implements l2_normalize, leaving a zero vector unchanged.
fn l2_normalize<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], ty: ColumnType) -> Result<Value> {
    let values = dense::<HALF>(&args[0]);
    let norm = dense_norm(&values);
    if norm <= 0.0 {
        return Ok(base(ty, encode_dense::<HALF>(&values)));
    }
    let normalized = values.iter().map(|v| narrow::<HALF>((*v as f64 / norm) as f32)).collect::<Result<Vec<_>>>()?;
    Ok(base(ty, encode_dense::<HALF>(&normalized)))
}

/// binary_quantize implements binary_quantize, setting a bit for every positive element.
fn binary_quantize<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    Ok(Value::Bit(dense::<HALF>(&args[0]).iter().map(|v| if *v > 0.0 { '1' } else { '0' }).collect()))
}

/// subvector implements subvector, taking a count of elements from a 1-based start as pgvector clamps them.
fn subvector<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], ty: ColumnType) -> Result<Value> {
    let name = dense_name::<HALF>();
    let values = dense::<HALF>(&args[0]);
    let (mut start, count) = (int(&args[1]) as i64, int(&args[2]) as i64);
    if count < 1 {
        return Err(at_least_one_dimension(name));
    }
    let dimensions = values.len() as i64;
    let end = if start > dimensions - count { dimensions + 1 } else { start + count };
    if start < 1 {
        start = 1;
    } else if start > dimensions {
        return Err(at_least_one_dimension(name));
    }
    check_dense_dimensions(name, (end - start) as usize)?;
    Ok(base(ty, encode_dense::<HALF>(&values[(start - 1) as usize..(end - 1) as usize])))
}

/// arithmetic combines two dense vectors element by element, where `underflow` also rejects a zero product of
/// non-zero elements.
fn arithmetic<const HALF: bool>(
    args: &[Value],
    ty: ColumnType,
    op: fn(f32, f32) -> f32,
    underflow: bool,
) -> Result<Value> {
    let (a, b) = pair::<HALF>(args)?;
    let mut result = Vec::with_capacity(a.len());
    for (x, y) in a.iter().zip(&b) {
        let value = narrow::<HALF>(op(*x, *y))?;
        if underflow && value == 0.0 && *x != 0.0 && *y != 0.0 {
            return Err(out_of_range("underflow"));
        }
        result.push(value);
    }
    Ok(base(ty, encode_dense::<HALF>(&result)))
}

/// add implements vector_add and halfvec_add.
fn add<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], ty: ColumnType) -> Result<Value> {
    arithmetic::<HALF>(args, ty, |x, y| x + y, false)
}

/// sub implements vector_sub and halfvec_sub.
fn sub<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], ty: ColumnType) -> Result<Value> {
    arithmetic::<HALF>(args, ty, |x, y| x - y, false)
}

/// mul implements vector_mul and halfvec_mul.
fn mul<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], ty: ColumnType) -> Result<Value> {
    arithmetic::<HALF>(args, ty, |x, y| x * y, true)
}

/// concat implements vector_concat and halfvec_concat.
fn concat<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], ty: ColumnType) -> Result<Value> {
    let mut values = dense::<HALF>(&args[0]);
    values.extend(dense::<HALF>(&args[1]));
    check_dense_dimensions(dense_name::<HALF>(), values.len())?;
    Ok(base(ty, encode_dense::<HALF>(&values)))
}

/// state returns the values of an avg state array, `[count, sums...]`, for the named routine.
fn state(routine: &str, value: &Value) -> Result<Vec<f64>> {
    let expected = || PgError::internal(format!("{routine}: expected state array"));
    let Value::Array(array) = value else { return Err(expected()) };
    if array.dims.len() != 1 || array.values.is_empty() {
        return Err(expected());
    }
    array.values.iter().map(|v| if let Value::Float8(f) = v { Ok(*f) } else { Err(expected()) }).collect()
}

/// state_array returns an avg state array.
fn state_array(values: Vec<f64>) -> Value {
    Value::Array(Box::new(Array::one_dimensional(crate::oid::FLOAT8, values.into_iter().map(Value::Float8).collect())))
}

/// accum implements the avg transition routines, adding a vector into the state.
fn accum<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    let state = state(&format!("{}_accum", dense_name::<HALF>()), &args[0])?;
    let values = dense::<HALF>(&args[1]);
    if state.len() == 1 {
        let mut next = vec![1.0];
        next.extend(values.iter().map(|v| *v as f64));
        return Ok(state_array(next));
    }
    check_expected_dimensions(state.len() as i32 - 1, values.len())?;
    let mut next = vec![state[0] + 1.0];
    for (sum, value) in state[1..].iter().zip(&values) {
        let total = sum + *value as f64;
        if total.is_infinite() {
            return Err(out_of_range("overflow"));
        }
        next.push(total);
    }
    Ok(state_array(next))
}

/// combine implements the avg combine routines, merging two states.
fn combine<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    let routine = format!("{}_combine", dense_name::<HALF>());
    let (mut a, mut b) = (state(&routine, &args[0])?, state(&routine, &args[1])?);
    if a[0] == 0.0 {
        std::mem::swap(&mut a, &mut b);
    }
    check_dense_dimensions(dense_name::<HALF>(), a.len() - 1)?;
    if b[0] == 0.0 {
        return Ok(state_array(a));
    }
    check_expected_dimensions(a.len() as i32 - 1, b.len() - 1)?;
    let mut next = vec![a[0] + b[0]];
    for (x, y) in a[1..].iter().zip(&b[1..]) {
        let total = x + y;
        if total.is_infinite() {
            return Err(out_of_range("overflow"));
        }
        next.push(total);
    }
    Ok(state_array(next))
}

/// avg implements the avg final routines, which return NULL for no rows.
fn avg<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], ty: ColumnType) -> Result<Value> {
    let state = state(&format!("{}_avg", dense_name::<HALF>()), &args[0])?;
    let count = state[0];
    if count == 0.0 {
        return Ok(Value::Null);
    }
    check_dense_dimensions(dense_name::<HALF>(), state.len() - 1)?;
    let values = state[1..].iter().map(|sum| narrow::<HALF>((sum / count) as f32)).collect::<Result<Vec<_>>>()?;
    Ok(base(ty, encode_dense::<HALF>(&values)))
}

/// dense_with_typmod implements the length coercions vector(vector, integer, boolean) and halfvec(halfvec, integer,
/// boolean).
fn dense_with_typmod<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], ty: ColumnType) -> Result<Value> {
    let data = base_data(&args[0]);
    dense_typmod::<HALF>(data, int(&args[1]))?;
    Ok(base(ty, data.to_vec()))
}

/// dense_to_float4 implements the casts from the dense types to real[].
fn dense_to_float4<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    let values = dense::<HALF>(&args[0]).into_iter().map(Value::Float4).collect();
    Ok(Value::Array(Box::new(Array::one_dimensional(crate::oid::FLOAT4, values))))
}

/// dense_to_dense implements the casts between the dense types, converting to halfvec when `HALF` is set.
fn dense_to_dense<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], ty: ColumnType) -> Result<Value> {
    let values = if HALF { dense::<false>(&args[0]) } else { dense::<true>(&args[0]) };
    check_expected_dimensions(int(&args[1]), values.len())?;
    let values = if HALF { values.into_iter().map(to_half).collect::<Result<Vec<_>>>()? } else { values };
    Ok(base(ty, encode_dense::<HALF>(&values)))
}

/// array_elements returns the elements of an array argument as floats, which must be one-dimensional without NULLs.
fn array_elements(value: &Value) -> Result<Vec<f32>> {
    let Value::Array(array) = value else { return Ok(Vec::new()) };
    if array.dims.len() > 1 {
        return Err(data_error("array must be 1-D"));
    }
    if array.values.iter().any(Value::is_null) {
        return Err(PgError::new(code::NULL_VALUE_NOT_ALLOWED, "array must not contain nulls"));
    }
    array
        .values
        .iter()
        .map(|v| match v {
            Value::Int4(i) => Ok(*i as f32),
            Value::Float4(f) => Ok(*f),
            Value::Float8(f) => Ok(*f as f32),
            Value::Numeric(n) => {
                let value = Numeric::to_f64(n) as f32;
                if value.is_infinite() && !matches!(n, Numeric::Infinity | Numeric::NegativeInfinity) {
                    return Err(out_of_range("overflow"));
                }
                Ok(value)
            }
            _ => Ok(0.0),
        })
        .collect()
}

/// array_to_dense implements the casts from arrays to the dense types.
fn array_to_dense<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], ty: ColumnType) -> Result<Value> {
    let name = dense_name::<HALF>();
    let values = array_elements(&args[0])?;
    check_dense_dimensions(name, values.len())?;
    check_expected_dimensions(int(&args[1]), values.len())?;
    let mut result = Vec::with_capacity(values.len());
    for value in values {
        check_element(name, value)?;
        result.push(if HALF { to_half(value)? } else { value });
    }
    Ok(base(ty, encode_dense::<HALF>(&result)))
}

/// hamming_distance implements hamming_distance, counting the bits that differ.
fn hamming_distance(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    let (a, b) = bit_pair(args)?;
    float(a.bytes().zip(b.bytes()).filter(|(x, y)| x != y).count() as f64)
}

/// jaccard_distance implements jaccard_distance, which is 1 when the bit strings share no set bit.
fn jaccard_distance(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    let (a, b) = bit_pair(args)?;
    let ones = |s: &str| s.bytes().filter(|b| *b == b'1').count();
    let both = a.bytes().zip(b.bytes()).filter(|(x, y)| *x == b'1' && *y == b'1').count();
    if both == 0 {
        return float(1.0);
    }
    float(1.0 - both as f64 / (ones(a) + ones(b) - both) as f64)
}

/// bits returns a bit string argument.
fn bits(value: &Value) -> &str {
    if let Value::Bit(bits) = value { bits } else { "" }
}

/// bit_pair returns two bit string arguments, which must have the same length.
fn bit_pair(args: &[Value]) -> Result<(&str, &str)> {
    let (a, b) = (bits(&args[0]), bits(&args[1]));
    if a.len() != b.len() {
        return Err(data_error(format!("different bit lengths {} and {}", a.len(), b.len())));
    }
    Ok((a, b))
}

/// sparse_pair returns two sparsevec arguments, which must have the same dimensions.
fn sparse_pair(args: &[Value]) -> Result<(SparseVector, SparseVector)> {
    let (a, b) = (sparse(&args[0]), sparse(&args[1]));
    if a.dimensions != b.dimensions {
        return Err(data_error(format!("different sparsevec dimensions {} and {}", a.dimensions, b.dimensions)));
    }
    Ok((a, b))
}

/// merge visits the elements of two sparsevecs in index order, passing each one's value at every index either has.
fn merge(a: &SparseVector, b: &SparseVector, mut visit: impl FnMut(f32, f32)) {
    let (mut i, mut j) = (0, 0);
    while i < a.indices.len() || j < b.indices.len() {
        if j >= b.indices.len() || (i < a.indices.len() && a.indices[i] < b.indices[j]) {
            visit(a.values[i], 0.0);
            i += 1;
        } else if i >= a.indices.len() || b.indices[j] < a.indices[i] {
            visit(0.0, b.values[j]);
            j += 1;
        } else {
            visit(a.values[i], b.values[j]);
            i += 1;
            j += 1;
        }
    }
}

/// sparse_l2_squared returns the squared Euclidean distance between two sparsevecs.
fn sparse_l2_squared(a: &SparseVector, b: &SparseVector) -> f32 {
    let mut distance = 0f32;
    merge(a, b, |x, y| distance += (x - y) * (x - y));
    distance
}

/// sparse_dot returns the inner product of two sparsevecs, to which only their shared indices contribute.
fn sparse_dot(a: &SparseVector, b: &SparseVector) -> f32 {
    let mut product = 0f32;
    merge(a, b, |x, y| product += x * y);
    product
}

/// sparse_l2_distance implements l2_distance for sparsevecs.
fn sparse_l2_distance(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    let (a, b) = sparse_pair(args)?;
    float((sparse_l2_squared(&a, &b) as f64).sqrt())
}

/// sparse_l2_squared_distance implements sparsevec_l2_squared_distance.
fn sparse_l2_squared_distance(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    let (a, b) = sparse_pair(args)?;
    float(sparse_l2_squared(&a, &b) as f64)
}

/// sparse_inner_product implements inner_product for sparsevecs.
fn sparse_inner_product(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    let (a, b) = sparse_pair(args)?;
    float(sparse_dot(&a, &b) as f64)
}

/// sparse_negative_inner_product implements sparsevec_negative_inner_product.
fn sparse_negative_inner_product(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    let (a, b) = sparse_pair(args)?;
    float(-(sparse_dot(&a, &b) as f64))
}

/// sparse_cosine_distance implements cosine_distance for sparsevecs.
fn sparse_cosine_distance(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    let (a, b) = sparse_pair(args)?;
    let squares = |s: &SparseVector| s.values.iter().fold(0f32, |sum, v| sum + v * v);
    float(cosine(sparse_dot(&a, &b), squares(&a), squares(&b)))
}

/// sparse_l1_distance implements l1_distance for sparsevecs.
fn sparse_l1_distance(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    let (a, b) = sparse_pair(args)?;
    let mut distance = 0f32;
    merge(&a, &b, |x, y| distance += (x - y).abs());
    float(distance as f64)
}

/// sparse_norm implements l2_norm for sparsevecs.
fn sparse_norm(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    float(dense_norm(&sparse(&args[0]).values))
}

/// sparse_l2_normalize implements l2_normalize for sparsevecs, leaving a zero vector unchanged.
fn sparse_l2_normalize(_: &mut Ctx<'_>, args: &[Value], ty: ColumnType) -> Result<Value> {
    let vector = sparse(&args[0]);
    let norm = dense_norm(&vector.values);
    if norm <= 0.0 {
        return Ok(base(ty, vector.encode()));
    }
    let mut result = SparseVector { dimensions: vector.dimensions, ..SparseVector::default() };
    for (index, value) in vector.indices.iter().zip(&vector.values) {
        let normalized = narrow::<false>((*value as f64 / norm) as f32)?;
        if normalized != 0.0 {
            result.indices.push(*index);
            result.values.push(normalized);
        }
    }
    Ok(base(ty, result.encode()))
}

/// sparse_with_typmod implements the length coercion sparsevec(sparsevec, integer, boolean).
fn sparse_with_typmod(_: &mut Ctx<'_>, args: &[Value], ty: ColumnType) -> Result<Value> {
    let data = base_data(&args[0]);
    sparse_typmod(data, int(&args[1]))?;
    Ok(base(ty, data.to_vec()))
}

/// dense_to_sparse implements the casts from the dense types to sparsevec.
fn dense_to_sparse<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], ty: ColumnType) -> Result<Value> {
    let values = dense::<HALF>(&args[0]);
    check_expected_dimensions(int(&args[1]), values.len())?;
    Ok(base(ty, SparseVector::from_dense(&values).encode()))
}

/// sparse_to_dense implements the casts from sparsevec to the dense types.
fn sparse_to_dense<const HALF: bool>(_: &mut Ctx<'_>, args: &[Value], ty: ColumnType) -> Result<Value> {
    let vector = sparse(&args[0]);
    if vector.dimensions as usize > MAX_DENSE_DIMS {
        return Err(too_many_dimensions(dense_name::<HALF>(), MAX_DENSE_DIMS as i64));
    }
    check_expected_dimensions(int(&args[1]), vector.dimensions as usize)?;
    let values = vector.to_dense();
    let values = if HALF { values.into_iter().map(to_half).collect::<Result<Vec<_>>>()? } else { values };
    Ok(base(ty, encode_dense::<HALF>(&values)))
}

/// array_to_sparse implements the casts from arrays to sparsevec.
fn array_to_sparse(_: &mut Ctx<'_>, args: &[Value], ty: ColumnType) -> Result<Value> {
    let values = array_elements(&args[0])?;
    if values.is_empty() {
        return Err(at_least_one_dimension("sparsevec"));
    }
    check_expected_dimensions(int(&args[1]), values.len())?;
    for value in &values {
        check_element("sparsevec", *value)?;
    }
    let vector = SparseVector::from_dense(&values);
    if vector.indices.len() > MAX_SPARSE_NON_ZERO {
        return Err(PgError::new(
            code::PROGRAM_LIMIT_EXCEEDED,
            format!("sparsevec cannot have more than {MAX_SPARSE_NON_ZERO} non-zero elements"),
        ));
    }
    Ok(base(ty, vector.encode()))
}
