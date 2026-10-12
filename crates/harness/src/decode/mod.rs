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

//! Renders binary-format values as Postgres text. Binary values carry no session state, so the rendering is
//! canonical: timestamptz is shown in UTC, floats use the shortest text that round-trips, and every other type
//! matches Postgres' output function under the default settings (DateStyle ISO, IntervalStyle postgres).

mod datetime;
mod geometric;
mod network;
mod numeric;
mod reader;
mod structured;

use reader::Reader;

/// decode_binary renders a binary-format value of the given type as text.
pub fn decode_binary(oid: u32, bytes: &[u8]) -> Result<String, String> {
    let mut r = Reader::new(bytes);
    let text = decode_value(oid, &mut r)?;
    r.expect_end(oid)?;
    Ok(text)
}

/// decode_value renders the value of the given type at the reader's position, consuming exactly its bytes.
fn decode_value(oid: u32, r: &mut Reader<'_>) -> Result<String, String> {
    Ok(match oid {
        16 => match r.u8()? {
            0 => "f".to_string(),
            1 => "t".to_string(),
            other => return Err(format!("invalid bool byte {other}")),
        },
        17 => bytea_hex(r.rest()),
        18 => internal_char(r.rest()),
        20 => r.i64()?.to_string(),
        21 => r.i16()?.to_string(),
        23 => r.i32()?.to_string(),
        26 | 28 | 29 => r.u32()?.to_string(),
        5069 => r.u64()?.to_string(),
        27 => {
            let block = r.u32()?;
            let offset = r.u16()?;
            format!("({block},{offset})")
        }
        700 => numeric::float4_text(f32::from_bits(r.u32()?)),
        701 => numeric::float8_text(f64::from_bits(r.u64()?)),
        1700 => numeric::numeric_text(r)?,
        1082 => datetime::date_text(r.i32()?),
        1083 => datetime::time_text(r.i64()?),
        1114 => datetime::timestamp_text(r.i64()?, false),
        1184 => datetime::timestamp_text(r.i64()?, true),
        1186 => {
            let microseconds = r.i64()?;
            let days = r.i32()?;
            let months = r.i32()?;
            datetime::interval_text(months, days, microseconds)
        }
        2950 => uuid_text(r.take(16)?),
        1560 | 1562 => bit_text(r)?,
        600 | 601 | 602 | 603 | 604 | 628 | 718 => geometric::geometric_text(oid, r)?,
        650 | 869 => network::inet_text(oid, r)?,
        829 => network::macaddr_text(r.take(6)?),
        774 => network::macaddr_text(r.take(8)?),
        2249 => structured::record_text(r)?,
        3802 => match r.u8()? {
            1 => String::from_utf8(r.rest().to_vec()).map_err(|err| format!("invalid UTF-8 in jsonb: {err}"))?,
            version => return Err(format!("unsupported jsonb binary version {version}")),
        },
        3614 => structured::tsvector_text(r)?,
        3904 | 3906 | 3908 | 3910 | 3912 | 3926 => structured::range_text(oid, r)?,
        4451 | 4532 | 4533 | 4534 | 4535 | 4536 => structured::multirange_text(oid, r)?,
        _ => {
            if let Some(element_oid) = array_element_oid(oid) {
                structured::array_text(oid, element_oid, r)?
            } else if is_text_like(oid) {
                String::from_utf8(r.rest().to_vec()).map_err(|err| format!("invalid UTF-8 in type {oid}: {err}"))?
            } else {
                return Err(format!("binary decoding is not implemented for type {oid}"));
            }
        }
    })
}

/// is_text_like reports whether a type's binary format is its text, as for text, varchar, name, json, and xml.
fn is_text_like(oid: u32) -> bool {
    matches!(oid, 19 | 25 | 114 | 142 | 1042 | 1043 | 705)
}

/// array_element_oid returns the element type of a built-in array type.
pub fn array_element_oid(oid: u32) -> Option<u32> {
    Some(match oid {
        143 => 142,
        199 => 114,
        271 => 5069,
        629 => 628,
        651 => 650,
        719 => 718,
        775 => 774,
        1000 => 16,
        1001 => 17,
        1002 => 18,
        1003 => 19,
        1005 => 21,
        1007 => 23,
        1009 => 25,
        1010 => 27,
        1011 => 28,
        1012 => 29,
        1014 => 1042,
        1015 => 1043,
        1016 => 20,
        1017 => 600,
        1018 => 601,
        1019 => 602,
        1020 => 603,
        1021 => 700,
        1022 => 701,
        1027 => 604,
        1028 => 26,
        1040 => 829,
        1041 => 869,
        1115 => 1114,
        1182 => 1082,
        1183 => 1083,
        1185 => 1184,
        1187 => 1186,
        1231 => 1700,
        1561 => 1560,
        1563 => 1562,
        2287 => 2249,
        2951 => 2950,
        3643 => 3614,
        3807 => 3802,
        3905 => 3904,
        3907 => 3906,
        3909 => 3908,
        3911 => 3910,
        3913 => 3912,
        3927 => 3926,
        _ => return None,
    })
}

/// bytea_hex renders bytes in bytea's hex output format.
fn bytea_hex(bytes: &[u8]) -> String {
    let mut text = String::with_capacity(2 + bytes.len() * 2);
    text.push_str("\\x");
    for byte in bytes {
        text.push_str(&format!("{byte:02x}"));
    }
    text
}

/// internal_char renders the "char" type, which shows bytes above 127 as an octal escape.
fn internal_char(bytes: &[u8]) -> String {
    match bytes.first() {
        None | Some(0) => String::new(),
        Some(&byte) if byte > 127 => format!("\\{byte:03o}"),
        Some(&byte) => (byte as char).to_string(),
    }
}

/// uuid_text renders a UUID in its hyphenated form.
fn uuid_text(bytes: &[u8]) -> String {
    let hex: String = bytes.iter().map(|b| format!("{b:02x}")).collect();
    format!("{}-{}-{}-{}-{}", &hex[0..8], &hex[8..12], &hex[12..16], &hex[16..20], &hex[20..32])
}

/// bit_text renders a bit or bit varying value.
fn bit_text(r: &mut Reader<'_>) -> Result<String, String> {
    let length = r.i32()?;
    if length < 0 {
        return Err(format!("invalid bit length {length}"));
    }
    let bytes = r.take((length as usize).div_ceil(8))?;
    Ok((0..length as usize).map(|i| if bytes[i / 8] & (0x80 >> (i % 8)) != 0 { '1' } else { '0' }).collect())
}
