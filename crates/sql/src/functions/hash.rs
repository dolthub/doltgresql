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

//! The hash support functions of the built-in types, built on Bob Jenkins' lookup3 hash as Postgres' hashfn.c has it,
//! each with an extended form that takes a seed and returns 64 bits.

use super::{ANYARRAY, Function};
use crate::error::{PgError, Result, code};
use crate::json::Json;
use crate::numeric::Numeric;
use crate::oid::{
    ANYENUM, BPCHAR, CHAR, FLOAT4, FLOAT8, INT2, INT4, INT8, INTERVAL, JSONB, NAME, NUMERIC, OID, OIDVECTOR, RECORD,
    TEXT, TIME, TIMESTAMP, TIMETZ, UUID,
};
use crate::rangetypes::{ANYMULTIRANGE, ANYRANGE, Range};
use crate::types::Value;

/// MACADDR, INET, and PG_LSN are the OIDs of the built-in base types with hash functions.
const MACADDR: u32 = 829;
const INET: u32 = 869;
const PG_LSN: u32 = 3220;

/// f declares a strict hash function.
const fn f(name: &'static str, args: &'static [u32], ret: u32, implementation: super::Implementation) -> Function {
    Function { name, args, ret, strict: true, variadic: false, implementation }
}

/// FUNCTIONS are the hash functions, where those of polymorphic types hash by the value's own type.
pub const FUNCTIONS: &[Function] = &[
    f("hashchar", &[CHAR], INT4, |_, a| standard(&a[0], CHAR)),
    f("hashcharextended", &[CHAR, INT8], INT8, |_, a| extended(&a[0], CHAR, &a[1])),
    f("hashint2", &[INT2], INT4, |_, a| standard(&a[0], INT2)),
    f("hashint2extended", &[INT2, INT8], INT8, |_, a| extended(&a[0], INT2, &a[1])),
    f("hashint4", &[INT4], INT4, |_, a| standard(&a[0], INT4)),
    f("hashint4extended", &[INT4, INT8], INT8, |_, a| extended(&a[0], INT4, &a[1])),
    f("hashint8", &[INT8], INT4, |_, a| standard(&a[0], INT8)),
    f("hashint8extended", &[INT8, INT8], INT8, |_, a| extended(&a[0], INT8, &a[1])),
    f("hashoid", &[OID], INT4, |_, a| standard(&a[0], OID)),
    f("hashoidextended", &[OID, INT8], INT8, |_, a| extended(&a[0], OID, &a[1])),
    f("hashenum", &[ANYENUM], INT4, |_, a| standard(&a[0], 0)),
    f("hashenumextended", &[ANYENUM, INT8], INT8, |_, a| extended(&a[0], 0, &a[1])),
    f("hashfloat4", &[FLOAT4], INT4, |_, a| standard(&a[0], FLOAT4)),
    f("hashfloat4extended", &[FLOAT4, INT8], INT8, |_, a| extended(&a[0], FLOAT4, &a[1])),
    f("hashfloat8", &[FLOAT8], INT4, |_, a| standard(&a[0], FLOAT8)),
    f("hashfloat8extended", &[FLOAT8, INT8], INT8, |_, a| extended(&a[0], FLOAT8, &a[1])),
    f("hashoidvector", &[OIDVECTOR], INT4, |_, a| standard(&a[0], OIDVECTOR)),
    f("hashoidvectorextended", &[OIDVECTOR, INT8], INT8, |_, a| extended(&a[0], OIDVECTOR, &a[1])),
    f("hashname", &[NAME], INT4, |_, a| standard(&a[0], NAME)),
    f("hashnameextended", &[NAME, INT8], INT8, |_, a| extended(&a[0], NAME, &a[1])),
    f("hashtext", &[TEXT], INT4, |_, a| standard(&a[0], TEXT)),
    f("hashtextextended", &[TEXT, INT8], INT8, |_, a| extended(&a[0], TEXT, &a[1])),
    f("hashbpchar", &[BPCHAR], INT4, |_, a| standard(&a[0], BPCHAR)),
    f("hashbpcharextended", &[BPCHAR, INT8], INT8, |_, a| extended(&a[0], BPCHAR, &a[1])),
    f("hashmacaddr", &[MACADDR], INT4, |_, a| standard(&a[0], MACADDR)),
    f("hashmacaddrextended", &[MACADDR, INT8], INT8, |_, a| extended(&a[0], MACADDR, &a[1])),
    f("hashinet", &[INET], INT4, |_, a| standard(&a[0], INET)),
    f("hashinetextended", &[INET, INT8], INT8, |_, a| extended(&a[0], INET, &a[1])),
    f("hash_numeric", &[NUMERIC], INT4, |_, a| standard(&a[0], NUMERIC)),
    f("hash_numeric_extended", &[NUMERIC, INT8], INT8, |_, a| extended(&a[0], NUMERIC, &a[1])),
    f("hash_array", &[ANYARRAY], INT4, |_, a| standard(&a[0], 0)),
    f("hash_array_extended", &[ANYARRAY, INT8], INT8, |_, a| extended(&a[0], 0, &a[1])),
    f("time_hash", &[TIME], INT4, |_, a| standard(&a[0], TIME)),
    f("time_hash_extended", &[TIME, INT8], INT8, |_, a| extended(&a[0], TIME, &a[1])),
    f("timetz_hash", &[TIMETZ], INT4, |_, a| standard(&a[0], TIMETZ)),
    f("timetz_hash_extended", &[TIMETZ, INT8], INT8, |_, a| extended(&a[0], TIMETZ, &a[1])),
    f("interval_hash", &[INTERVAL], INT4, |_, a| standard(&a[0], INTERVAL)),
    f("interval_hash_extended", &[INTERVAL, INT8], INT8, |_, a| extended(&a[0], INTERVAL, &a[1])),
    f("timestamp_hash", &[TIMESTAMP], INT4, |_, a| standard(&a[0], TIMESTAMP)),
    f("timestamp_hash_extended", &[TIMESTAMP, INT8], INT8, |_, a| extended(&a[0], TIMESTAMP, &a[1])),
    f("uuid_hash", &[UUID], INT4, |_, a| standard(&a[0], UUID)),
    f("uuid_hash_extended", &[UUID, INT8], INT8, |_, a| extended(&a[0], UUID, &a[1])),
    f("pg_lsn_hash", &[PG_LSN], INT4, |_, a| standard(&a[0], PG_LSN)),
    f("pg_lsn_hash_extended", &[PG_LSN, INT8], INT8, |_, a| extended(&a[0], PG_LSN, &a[1])),
    f("jsonb_hash", &[JSONB], INT4, |_, a| standard(&a[0], JSONB)),
    f("jsonb_hash_extended", &[JSONB, INT8], INT8, |_, a| extended(&a[0], JSONB, &a[1])),
    f("hash_range", &[ANYRANGE], INT4, |_, a| standard(&a[0], 0)),
    f("hash_range_extended", &[ANYRANGE, INT8], INT8, |_, a| extended(&a[0], 0, &a[1])),
    f("hash_multirange", &[ANYMULTIRANGE], INT4, |_, a| standard(&a[0], 0)),
    f("hash_multirange_extended", &[ANYMULTIRANGE, INT8], INT8, |_, a| extended(&a[0], 0, &a[1])),
    f("hash_record", &[RECORD], INT4, |_, a| standard(&a[0], 0)),
    f("hash_record_extended", &[RECORD, INT8], INT8, |_, a| extended(&a[0], 0, &a[1])),
];

/// standard returns a value's 32-bit hash as an integer.
fn standard(value: &Value, type_oid: u32) -> Result<Value> {
    Ok(Value::Int4(hash_value(value, type_oid, None)? as u32 as i32))
}

/// extended returns a value's 64-bit hash with a seed as a bigint.
fn extended(value: &Value, type_oid: u32, seed: &Value) -> Result<Value> {
    let Value::Int8(seed) = seed else { return Err(PgError::internal("a hash seed that is not a bigint")) };
    Ok(Value::Int8(hash_value(value, type_oid, Some(*seed as u64))? as i64))
}

/// mix is lookup3's mixing of three words.
fn mix(a: &mut u32, b: &mut u32, c: &mut u32) {
    *a = a.wrapping_sub(*c);
    *a ^= c.rotate_left(4);
    *c = c.wrapping_add(*b);
    *b = b.wrapping_sub(*a);
    *b ^= a.rotate_left(6);
    *a = a.wrapping_add(*c);
    *c = c.wrapping_sub(*b);
    *c ^= b.rotate_left(8);
    *b = b.wrapping_add(*a);
    *a = a.wrapping_sub(*c);
    *a ^= c.rotate_left(16);
    *c = c.wrapping_add(*b);
    *b = b.wrapping_sub(*a);
    *b ^= a.rotate_left(19);
    *a = a.wrapping_add(*c);
    *c = c.wrapping_sub(*b);
    *c ^= b.rotate_left(4);
    *b = b.wrapping_add(*a);
}

/// finish is lookup3's final mixing of three words into the last two.
fn finish(a: &mut u32, b: &mut u32, c: &mut u32) {
    *c ^= *b;
    *c = c.wrapping_sub(b.rotate_left(14));
    *a ^= *c;
    *a = a.wrapping_sub(c.rotate_left(11));
    *b ^= *a;
    *b = b.wrapping_sub(a.rotate_left(25));
    *c ^= *b;
    *c = c.wrapping_sub(b.rotate_left(16));
    *a ^= *c;
    *a = a.wrapping_sub(c.rotate_left(4));
    *b ^= *a;
    *b = b.wrapping_sub(a.rotate_left(14));
    *c ^= *b;
    *c = c.wrapping_sub(b.rotate_left(24));
}

/// start returns lookup3's starting words for a key length, mixing in a nonzero seed.
fn start(length: usize, seed: u64) -> (u32, u32, u32) {
    let initial = 0x9e3779b9u32.wrapping_add(length as u32).wrapping_add(3923095);
    let (mut a, mut b, mut c) = (initial, initial, initial);
    if seed != 0 {
        a = a.wrapping_add((seed >> 32) as u32);
        b = b.wrapping_add(seed as u32);
        mix(&mut a, &mut b, &mut c);
    }
    (a, b, c)
}

/// result returns the hash in the last two words: the last for a standard hash, and both for an extended one.
fn result(b: u32, c: u32, seed: Option<u64>) -> u64 {
    match seed {
        None => c as u64,
        Some(_) => (b as u64) << 32 | c as u64,
    }
}

/// bytes hashes bytes as Postgres' hash_bytes and hash_bytes_extended do on a little-endian machine.
fn bytes(key: &[u8], seed: Option<u64>) -> u64 {
    let word = |bytes: &[u8]| u32::from_le_bytes([bytes[0], bytes[1], bytes[2], bytes[3]]);
    let (mut a, mut b, mut c) = start(key.len(), seed.unwrap_or(0));
    let mut rest = key;
    while rest.len() >= 12 {
        a = a.wrapping_add(word(&rest[0..4]));
        b = b.wrapping_add(word(&rest[4..8]));
        c = c.wrapping_add(word(&rest[8..12]));
        mix(&mut a, &mut b, &mut c);
        rest = &rest[12..];
    }
    for (i, &byte) in rest.iter().enumerate() {
        let byte = byte as u32;
        match i {
            0..=3 => a = a.wrapping_add(byte << (8 * i)),
            4..=7 => b = b.wrapping_add(byte << (8 * (i - 4))),
            _ => c = c.wrapping_add(byte << (8 * (i - 7))),
        }
    }
    finish(&mut a, &mut b, &mut c);
    result(b, c, seed)
}

/// uint32 hashes a 32-bit value as Postgres' hash_bytes_uint32 and hash_bytes_uint32_extended do.
fn uint32(key: u32, seed: Option<u64>) -> u64 {
    let (mut a, mut b, mut c) = start(4, seed.unwrap_or(0));
    a = a.wrapping_add(key);
    finish(&mut a, &mut b, &mut c);
    result(b, c, seed)
}

/// int8 hashes a bigint as Postgres' hashint8 does, folding its high half into its low half so that values that fit
/// in an integer hash as hashint4 hashes them.
fn int8(value: i64, seed: Option<u64>) -> u64 {
    let high = (value >> 32) as u32;
    uint32(value as u32 ^ if value >= 0 { high } else { !high }, seed)
}

/// float8 hashes a float as Postgres' hashfloat8 does, where both zeros hash alike and every NaN hashes as one.
fn float8(value: f64, seed: Option<u64>) -> u64 {
    match value {
        0.0 => seed.unwrap_or(0),
        v if v.is_nan() => bytes(&f64::NAN.to_le_bytes(), seed),
        v => bytes(&v.to_le_bytes(), seed),
    }
}

/// numeric hashes a numeric as Postgres' hash_numeric does: its base-10000 digits without leading or trailing zeros,
/// mixed with the weight of the first, so that equal values of different scales hash alike.
fn numeric(value: &Numeric, seed: Option<u64>) -> u64 {
    if !matches!(value, Numeric::Finite { .. }) {
        return seed.unwrap_or(0);
    }
    let sent = value.send();
    let count = u16::from_be_bytes([sent[0], sent[1]]) as usize;
    let weight = i16::from_be_bytes([sent[2], sent[3]]) as i64;
    if count == 0 {
        return seed.map_or(u32::MAX as u64, |s| s.wrapping_sub(1));
    }
    let digits: Vec<u8> = sent[8..8 + count * 2].chunks(2).flat_map(|d| [d[1], d[0]]).collect();
    match seed {
        None => (bytes(&digits, None) as u32 ^ weight as u32) as u64,
        Some(_) => bytes(&digits, seed) ^ weight as u64,
    }
}

/// JB_FARRAY and JB_FOBJECT are the container flags that Postgres' jsonb_hash mixes in at each array and object.
const JB_FARRAY: u64 = 0x40000000;
const JB_FOBJECT: u64 = 0x20000000;

/// jsonb hashes a jsonb value as Postgres' jsonb_hash does, where an empty array or object hashes as nothing.
fn jsonb(value: &Json, seed: Option<u64>) -> u64 {
    if matches!(value, Json::Array(items) if items.is_empty())
        || matches!(value, Json::Object(pairs) if pairs.is_empty())
    {
        return seed.unwrap_or(0);
    }
    let mut hash = JsonbHash { hash: 0, seed };
    if !matches!(value, Json::Array(_) | Json::Object(_)) {
        hash.container(JB_FARRAY);
    }
    hash.walk(value);
    hash.hash
}

/// JsonbHash is the running hash of a jsonb value's tokens.
struct JsonbHash {
    hash: u64,
    seed: Option<u64>,
}

impl JsonbHash {
    /// container mixes in the flag of an array or object that begins, in both halves of an extended hash.
    fn container(&mut self, flag: u64) {
        self.hash ^= if self.seed.is_some() { flag << 32 | flag } else { flag };
    }

    /// scalar rotates the hash and mixes in a key's or scalar's hash, as Postgres' JsonbHashScalarValue does.
    fn scalar(&mut self, value: &Json) {
        let seed = self.seed;
        let scalar = match value {
            Json::Null => seed.unwrap_or(0).wrapping_add(1),
            Json::String(s) => bytes(s.as_bytes(), seed),
            Json::Number(n) => numeric(n, seed),
            Json::Bool(b) if seed.is_some_and(|s| s != 0) => uint32(*b as u32, seed),
            Json::Bool(true) => 0x02,
            Json::Bool(false) => 0x04,
            Json::Array(_) | Json::Object(_) => 0,
        };
        self.hash = rotate(self.hash, seed) ^ scalar;
    }

    /// walk mixes in a value's tokens in jsonb's order: elements in order, and keys in jsonb's sorted order, each
    /// followed by its value.
    fn walk(&mut self, value: &Json) {
        match value {
            Json::Array(items) => {
                self.container(JB_FARRAY);
                for item in items {
                    self.walk(item);
                }
            }
            Json::Object(pairs) => {
                self.container(JB_FOBJECT);
                let mut sorted: Vec<&(String, Json)> = pairs.iter().collect();
                sorted.sort_by(|a, b| crate::json::compare_keys(&a.0, &b.0));
                for (key, item) in sorted {
                    self.scalar(&Json::String(key.clone()));
                    self.walk(item);
                }
            }
            scalar => self.scalar(scalar),
        }
    }
}

/// rotate rotates a running hash left one bit, rotating each half of an extended hash on its own, as Postgres'
/// pg_rotate_left32 and ROTATE_HIGH_AND_LOW_32BITS do.
fn rotate(hash: u64, seed: Option<u64>) -> u64 {
    match seed {
        None => (hash as u32).rotate_left(1) as u64,
        Some(_) => (hash << 1) & 0xfffffffefffffffe | (hash >> 31) & 0x100000001,
    }
}

/// combine adds an element's hash to a running hash of a sequence, as Postgres' hash_array and hash_record do.
fn combine(running: u64, element: u64) -> u64 {
    (running << 5).wrapping_sub(running).wrapping_add(element)
}

/// truncate keeps the low 32 bits of a standard hash, and all of an extended one.
fn truncate(hash: u64, seed: Option<u64>) -> u64 {
    if seed.is_none() { hash as u32 as u64 } else { hash }
}

/// range hashes a range from its flags and bounds as Postgres' hash_range does.
fn range(range: &Range, subtype: u32, seed: Option<u64>) -> Result<u64> {
    let flags = if range.empty {
        0x01
    } else {
        u32::from(range.lower.inclusive) << 1
            | u32::from(range.upper.inclusive) << 2
            | u32::from(range.lower.value.is_none()) << 3
            | u32::from(range.upper.value.is_none()) << 4
    };
    let bound = |bound: &Option<Value>| match bound {
        Some(value) if !range.empty => hash_value(value, subtype, seed),
        _ => Ok(0),
    };
    let (lower, upper) = (bound(&range.lower.value)?, bound(&range.upper.value)?);
    Ok(rotate(uint32(flags, seed) ^ lower, seed) ^ upper)
}

/// check_hashable fails as Postgres does for a type without a hash function.
fn check_hashable(type_oid: u32, seed: Option<u64>) -> Result<()> {
    if hashable(type_oid) { Ok(()) } else { Err(no_hash(type_oid, seed)) }
}

/// hashable reports whether a type has a hash function.
pub(crate) fn hashable(type_oid: u32) -> bool {
    match crate::usertypes::get(type_oid) {
        Some(user) => !matches!(user.kind, crate::usertypes::Kind::Base(_)),
        None => match type_oid {
            MACADDR | INET | 650 | PG_LSN | 27 => true,
            crate::oid::JSON | crate::oid::XML | crate::oid::BIT | crate::oid::VARBIT => false,
            _ => crate::basetypes::get(type_oid).is_none(),
        },
    }
}

/// no_hash is the error of hashing a value of a type without a hash function.
fn no_hash(type_oid: u32, seed: Option<u64>) -> PgError {
    PgError::new(
        code::UNDEFINED_FUNCTION,
        format!(
            "could not identify {} hash function for type {}",
            if seed.is_some() { "an extended" } else { "a" },
            crate::cast::type_display(type_oid)
        ),
    )
}

/// enum_oid returns the OID of an enum value's label.
fn enum_oid(value: &crate::types::EnumValue) -> u32 {
    let labels = crate::usertypes::get(value.type_oid).map(|t| t.definition.enum_labels.clone()).unwrap_or_default();
    labels
        .iter()
        .find(|l| crate::catalog::id::segments(&l.id).pop().as_deref() == Some(value.label.as_str()))
        .map_or(0, |l| crate::catalog::oids::oid(&l.id))
}

/// char_byte returns the byte a "char" value holds, which prints as an octal escape when it is not ASCII.
fn char_byte(text: &str) -> u8 {
    match text.strip_prefix('\\') {
        Some(octal) if octal.len() == 3 => u8::from_str_radix(octal, 8).unwrap_or(b'\\'),
        _ => text.as_bytes().first().copied().unwrap_or(0),
    }
}

/// hash_oid returns the OID an element of an oidvector holds.
fn hash_oid(value: &Value) -> u32 {
    match value {
        Value::Oid(oid) => *oid,
        Value::Reg(reg) => reg.oid,
        _ => 0,
    }
}

/// hash_value hashes a value of a type as the type's hash support function does, into 32 bits without a seed and
/// 64 bits with one, taking the type from the value when the type is 0.
pub fn hash_value(value: &Value, type_oid: u32, seed: Option<u64>) -> Result<u64> {
    let type_oid = if type_oid == 0 { super::value_type(value) } else { type_oid };
    Ok(match value {
        Value::Bool(b) => uint32(*b as u32, seed),
        Value::Int2(v) => uint32(*v as i32 as u32, seed),
        Value::Int4(v) | Value::Date(v) => uint32(*v as u32, seed),
        Value::Int8(v) | Value::Time(v) | Value::Timestamp(v) | Value::TimestampTz(v) => int8(*v, seed),
        Value::Oid(v) => uint32(*v, seed),
        Value::Reg(reg) => uint32(reg.oid, seed),
        Value::Float4(v) => float8(*v as f64, seed),
        Value::Float8(v) => float8(*v, seed),
        Value::Numeric(n) => numeric(n, seed),
        Value::Text(t) if type_oid == CHAR => uint32(char_byte(t) as i8 as i32 as u32, seed),
        Value::Text(t) if type_oid == BPCHAR => bytes(t.trim_end_matches(' ').as_bytes(), seed),
        Value::Text(t) => bytes(t.as_bytes(), seed),
        Value::Bytea(b) => bytes(b, seed),
        Value::Uuid(u) => bytes(u, seed),
        Value::TimeTz(time, zone) => int8(*time, seed) ^ uint32(*zone as u32, seed),
        Value::Interval(iv) => int8(iv.cmp_key() as i64, seed),
        Value::Jsonb(j) => jsonb(j, seed),
        Value::Enum(e) => uint32(enum_oid(e), seed),
        Value::Array(array) if type_oid == OIDVECTOR => {
            let oids: Vec<u8> = array.values.iter().flat_map(|v| hash_oid(v).to_le_bytes()).collect();
            bytes(&oids, seed)
        }
        Value::Array(array) => {
            check_hashable(array.element, seed)?;
            let mut hash = 1;
            for element in &array.values {
                let element_hash = if element.is_null() { 0 } else { hash_value(element, array.element, seed)? };
                hash = combine(hash, element_hash);
            }
            truncate(hash, seed)
        }
        Value::Range(r) => {
            let subtype = crate::rangetypes::range_type(r.type_oid).map_or(0, |t| t.subtype);
            check_hashable(subtype, seed)?;
            range(r, subtype, seed)?
        }
        Value::Multirange(m) => {
            let subtype = crate::rangetypes::multirange_type(m.type_oid).map_or(0, |t| t.subtype);
            check_hashable(subtype, seed)?;
            let mut hash = 1;
            for r in &m.ranges {
                hash = combine(hash, range(r, subtype, seed)?);
            }
            truncate(hash, seed)
        }
        Value::Composite(c) => {
            let types: Vec<u32> = match crate::usertypes::get(c.type_oid).map(|t| t.kind.clone()) {
                Some(crate::usertypes::Kind::Composite(attributes)) => attributes.iter().map(|a| a.1.oid).collect(),
                _ => c.fields.iter().map(super::value_type).collect(),
            };
            record(&c.fields, &types, seed)?
        }
        Value::Record(fields) => record(fields, &fields.iter().map(super::value_type).collect::<Vec<_>>(), seed)?,
        Value::Base(base) => match base.type_oid {
            MACADDR => bytes(&base.data, seed),
            INET | 650 => bytes(&[&base.data[..2], &base.data[4..]].concat(), seed),
            PG_LSN => int8(i64::from_be_bytes(base.data[..8].try_into().unwrap_or_default()), seed),
            other => return Err(no_hash(other, seed)),
        },
        _ => return Err(no_hash(type_oid, seed)),
    })
}

/// record hashes the fields of a row as Postgres' hash_record does, checking that every field's type has a hash
/// function.
fn record(fields: &[Value], types: &[u32], seed: Option<u64>) -> Result<u64> {
    let mut hash = 0;
    for (field, &type_oid) in fields.iter().zip(types) {
        check_hashable(type_oid, seed)?;
        let field_hash = if field.is_null() { 0 } else { hash_value(field, type_oid, seed)? };
        hash = combine(hash, field_hash);
    }
    Ok(truncate(hash, seed))
}
