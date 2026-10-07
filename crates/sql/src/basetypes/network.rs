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

//! The network address types, read and written as Postgres' network.c, inet_net_pton.c, inet_net_ntop.c, and mac.c
//! do, and stored as their binary formats: inet, cidr, and macaddr.

use std::cmp::Ordering;

use crate::error::{PgError, Result, code};
use crate::extensions::BaseType;

/// IPV4 and IPV6 are the address families of Postgres' binary format.
const IPV4: u8 = 2;
const IPV6: u8 = 3;

/// INET is the inet type.
pub const INET: BaseType = BaseType {
    name: "inet",
    input: |text, _| network_in(text, false),
    output: network_out,
    receive: |bytes, _| network_recv(bytes, false),
    send: <[u8]>::to_vec,
    typmod_in: |_| Ok(-1),
    typmod: |_, _| Ok(()),
    compare: network_cmp,
    vector: None,
};

/// CIDR is the cidr type.
pub const CIDR: BaseType = BaseType {
    name: "cidr",
    input: |text, _| network_in(text, true),
    output: network_out,
    receive: |bytes, _| network_recv(bytes, true),
    send: <[u8]>::to_vec,
    typmod_in: |_| Ok(-1),
    typmod: |_, _| Ok(()),
    compare: network_cmp,
    vector: None,
};

/// MACADDR is the macaddr type.
pub const MACADDR: BaseType = BaseType {
    name: "macaddr",
    input: |text, _| macaddr_in(text),
    output: |bytes| bytes.iter().map(|b| format!("{b:02x}")).collect::<Vec<_>>().join(":"),
    receive: |bytes, _| match bytes.len() {
        6 => Ok(bytes.to_vec()),
        n if n < 6 => Err(PgError::new(code::PROTOCOL_VIOLATION, "insufficient data left in message")),
        _ => Err(PgError::new(code::INVALID_BINARY_REPRESENTATION, "incorrect binary data format")),
    },
    send: <[u8]>::to_vec,
    typmod_in: |_| Ok(-1),
    typmod: |_, _| Ok(()),
    compare: <[u8]>::cmp,
    vector: None,
};

/// address_size returns the bytes of an address of a family.
fn address_size(family: u8) -> usize {
    if family == IPV4 { 4 } else { 16 }
}

/// network_in reads an inet or cidr value, as network_in does, storing its family, bits, whether it is a cidr, its
/// address size, and its address.
fn network_in(text: &str, cidr: bool) -> Result<Vec<u8>> {
    let type_name = if cidr { "cidr" } else { "inet" };
    let invalid = || {
        PgError::new(
            code::INVALID_TEXT_REPRESENTATION,
            format!("invalid input syntax for type {type_name}: \"{text}\""),
        )
    };
    let family = if text.contains(':') { IPV6 } else { IPV4 };
    let mut address = [0u8; 16];
    let bits = match (family, cidr) {
        (IPV4, true) => cidr_pton_ipv4(text.as_bytes(), &mut address),
        (IPV4, false) => net_pton_ipv4(text.as_bytes(), &mut address),
        _ => pton_ipv6(text.as_bytes(), &mut address),
    }
    .ok_or_else(invalid)?;
    let max = if family == IPV4 { 32 } else { 128 };
    if bits > max {
        return Err(invalid());
    }
    if cidr && !address_ok(&address, bits, family) {
        return Err(PgError {
            detail: Some("Value has bits set to right of mask.".into()),
            ..PgError::new(code::INVALID_TEXT_REPRESENTATION, format!("invalid cidr value: \"{text}\""))
        });
    }
    let size = address_size(family);
    let mut out = vec![family, bits as u8, u8::from(cidr), size as u8];
    out.extend_from_slice(&address[..size]);
    Ok(out)
}

/// cidr_pton_ipv4 reads an IPv4 network, whose octets may be abbreviated and whose bits default to its class, as
/// inet_cidr_pton_ipv4 does, returning its bits.
fn cidr_pton_ipv4(src: &[u8], dst: &mut [u8; 16]) -> Option<u32> {
    let mut i = 0;
    let mut written = 0;
    let next = |i: &mut usize| {
        let ch = src.get(*i).copied().unwrap_or(0);
        *i += 1;
        ch
    };
    let mut ch = next(&mut i);
    if ch == b'0' && matches!(src.get(i), Some(b'x' | b'X')) && src.get(i + 1).is_some_and(u8::is_ascii_hexdigit) {
        i += 1;
        let (mut dirty, mut tmp) = (0, 0u8);
        loop {
            ch = next(&mut i);
            if ch == 0 || !ch.is_ascii_hexdigit() {
                break;
            }
            let n = (ch as char).to_digit(16)? as u8;
            tmp = if dirty == 0 { n } else { (tmp << 4) | n };
            dirty += 1;
            if dirty == 2 {
                *dst.get_mut(written).filter(|_| written < 4)? = tmp;
                written += 1;
                dirty = 0;
            }
        }
        if dirty != 0 {
            *dst.get_mut(written).filter(|_| written < 4)? = tmp << 4;
            written += 1;
        }
    } else if ch.is_ascii_digit() {
        loop {
            let mut tmp = 0u32;
            loop {
                tmp = tmp * 10 + u32::from(ch - b'0');
                if tmp > 255 {
                    return None;
                }
                ch = next(&mut i);
                if ch == 0 || !ch.is_ascii_digit() {
                    break;
                }
            }
            *dst.get_mut(written).filter(|_| written < 4)? = tmp as u8;
            written += 1;
            if ch == 0 || ch == b'/' {
                break;
            }
            if ch != b'.' {
                return None;
            }
            ch = next(&mut i);
            if !ch.is_ascii_digit() {
                return None;
            }
        }
    } else {
        return None;
    }
    let mut bits: Option<u32> = None;
    if ch == b'/' && src.get(i).is_some_and(u8::is_ascii_digit) && written > 0 {
        ch = next(&mut i);
        let mut value = 0;
        loop {
            value = value * 10 + u32::from(ch - b'0');
            if value > 32 {
                return None;
            }
            ch = next(&mut i);
            if ch == 0 || !ch.is_ascii_digit() {
                break;
            }
        }
        bits = Some(value);
    }
    if ch != 0 || written == 0 {
        return None;
    }
    let bits = match bits {
        Some(bits) => bits,
        None => {
            let first = dst[0];
            let class = match first {
                240.. => 32,
                224.. => 8,
                192.. => 24,
                128.. => 16,
                _ => 8,
            };
            let class = class.max(written as u32 * 8);
            if class == 8 && first == 224 { 4 } else { class }
        }
    };
    Some(bits)
}

/// net_pton_ipv4 reads an IPv4 address with an optional `/bits`, which must name all four octets unless the bits
/// cover fewer, as inet_net_pton_ipv4 does, returning its bits.
fn net_pton_ipv4(src: &[u8], dst: &mut [u8; 16]) -> Option<u32> {
    let mut i = 0;
    let mut written = 0;
    let at = |i: usize| src.get(i).copied().unwrap_or(0);
    let mut ch = at(i);
    i += 1;
    while ch.is_ascii_digit() {
        let mut tmp = 0u32;
        loop {
            tmp = tmp * 10 + u32::from(ch - b'0');
            if tmp > 255 {
                return None;
            }
            ch = at(i);
            i += 1;
            if ch == 0 || !ch.is_ascii_digit() {
                break;
            }
        }
        if written == 4 {
            return None;
        }
        dst[written] = tmp as u8;
        written += 1;
        if ch == 0 || ch == b'/' {
            break;
        }
        if ch != b'.' {
            return None;
        }
        ch = at(i);
        i += 1;
    }
    let mut bits: Option<u32> = None;
    if ch == b'/' && at(i).is_ascii_digit() && written > 0 {
        ch = at(i);
        i += 1;
        let mut value = 0;
        loop {
            value = value * 10 + u32::from(ch - b'0');
            if value > 32 {
                return None;
            }
            ch = at(i);
            i += 1;
            if ch == 0 || !ch.is_ascii_digit() {
                break;
            }
        }
        bits = Some(value);
    }
    if ch != 0 || written == 0 {
        return None;
    }
    let bits = match bits {
        Some(bits) => bits,
        None if written == 4 => 32,
        None => return None,
    };
    if bits / 8 > written as u32 {
        return None;
    }
    Some(bits)
}

/// get_bits reads the decimal bits after an IPv6 address's slash, without leading zeros, as getbits does.
fn get_bits(src: &[u8]) -> Option<u32> {
    if src.is_empty() || (src.len() > 1 && src[0] == b'0') || !src.iter().all(u8::is_ascii_digit) {
        return None;
    }
    let value: u32 = std::str::from_utf8(src).ok()?.parse().ok()?;
    (value <= 128).then_some(value)
}

/// get_v4 reads the dotted IPv4 address that ends an IPv6 address, with optional bits, as getv4 does.
fn get_v4(src: &[u8], dst: &mut [u8]) -> Option<Option<u32>> {
    let (mut value, mut digits, mut written) = (0u32, 0, 0);
    for (i, &ch) in src.iter().enumerate() {
        if ch.is_ascii_digit() {
            if digits != 0 && value == 0 {
                return None;
            }
            digits += 1;
            value = value * 10 + u32::from(ch - b'0');
            if value > 255 {
                return None;
            }
        } else if ch == b'.' || ch == b'/' {
            if written > 3 {
                return None;
            }
            dst[written] = value as u8;
            written += 1;
            if ch == b'/' {
                return get_bits(&src[i + 1..]).map(Some);
            }
            value = 0;
            digits = 0;
        } else {
            return None;
        }
    }
    if digits == 0 || written > 3 {
        return None;
    }
    dst[written] = value as u8;
    Some(None)
}

/// pton_ipv6 reads an IPv6 address with an optional `/bits`, as inet_cidr_pton_ipv6 does, returning its bits.
fn pton_ipv6(src: &[u8], dst: &mut [u8; 16]) -> Option<u32> {
    let mut tmp = [0u8; 16];
    let mut tp = 0;
    let mut colon: Option<usize> = None;
    let mut i = 0;
    if src.first() == Some(&b':') {
        if src.get(1) != Some(&b':') {
            return None;
        }
        i = 1;
    }
    let mut token = i;
    let (mut saw_digit, mut value, mut digits) = (false, 0u32, 0);
    let mut bits: Option<u32> = None;
    while i < src.len() {
        let ch = src[i];
        i += 1;
        if let Some(n) = (ch as char).to_digit(16) {
            value = (value << 4) | n;
            digits += 1;
            if digits > 4 {
                return None;
            }
            saw_digit = true;
            continue;
        }
        if ch == b':' {
            token = i;
            if !saw_digit {
                if colon.is_some() {
                    return None;
                }
                colon = Some(tp);
                continue;
            } else if i == src.len() {
                return None;
            }
            if tp + 2 > 16 {
                return None;
            }
            tmp[tp] = (value >> 8) as u8;
            tmp[tp + 1] = value as u8;
            tp += 2;
            saw_digit = false;
            digits = 0;
            value = 0;
            continue;
        }
        if ch == b'.' && tp + 4 <= 16 {
            let v4 = get_v4(&src[token..], &mut tmp[tp..tp + 4]);
            if let Some(v4_bits) = v4 {
                tp += 4;
                saw_digit = false;
                bits = v4_bits;
                break;
            }
        }
        if ch == b'/'
            && let Some(value) = get_bits(&src[i..])
        {
            bits = Some(value);
            break;
        }
        return None;
    }
    if saw_digit {
        if tp + 2 > 16 {
            return None;
        }
        tmp[tp] = (value >> 8) as u8;
        tmp[tp + 1] = value as u8;
        tp += 2;
    }
    if let Some(colon) = colon {
        if tp == 16 {
            return None;
        }
        let n = tp - colon;
        for k in 1..=n {
            tmp[16 - k] = tmp[colon + n - k];
            tmp[colon + n - k] = 0;
        }
        tp = 16;
    }
    if tp != 16 {
        return None;
    }
    *dst = tmp;
    Some(bits.unwrap_or(128))
}

/// address_ok reports whether an address has no bits set past its mask, as addressOK does.
fn address_ok(address: &[u8], bits: u32, family: u8) -> bool {
    let max = address_size(family) * 8;
    if bits as usize == max {
        return true;
    }
    let mut byte = bits as usize / 8;
    let mut mask: u8 = if bits.is_multiple_of(8) { 0xff } else { 0xff >> (bits % 8) };
    while byte < max / 8 {
        if address[byte] & mask != 0 {
            return false;
        }
        mask = 0xff;
        byte += 1;
    }
    true
}

/// network_out writes an inet or cidr value, as network_out does, with its bits unless an inet has all of them.
fn network_out(bytes: &[u8]) -> String {
    let (family, bits, cidr) = (bytes[0], u32::from(bytes[1]), bytes[2] != 0);
    let address = &bytes[4..];
    let mut out = if family == IPV4 {
        let octets: Vec<String> = address.iter().map(u8::to_string).collect();
        let mut out = octets.join(".");
        if bits != 32 {
            out.push_str(&format!("/{bits}"));
        }
        out
    } else {
        ntop_ipv6(address, bits)
    };
    if cidr && !out.contains('/') {
        out.push_str(&format!("/{bits}"));
    }
    out
}

/// ntop_ipv6 writes an IPv6 address, shortening its longest run of zero words to `::` and writing an embedded IPv4
/// address in dots, as inet_net_ntop_ipv6 does.
fn ntop_ipv6(address: &[u8], bits: u32) -> String {
    let words: Vec<u32> =
        address.as_chunks::<2>().0.iter().map(|[high, low]| u32::from(*high) << 8 | u32::from(*low)).collect();
    let (mut best, mut current): ((i32, i32), (i32, i32)) = ((-1, 0), (-1, 0));
    for (i, &word) in words.iter().enumerate() {
        if word == 0 {
            current = if current.0 == -1 { (i as i32, 1) } else { (current.0, current.1 + 1) };
        } else if current.0 != -1 {
            if best.0 == -1 || current.1 > best.1 {
                best = current;
            }
            current.0 = -1;
        }
    }
    if current.0 != -1 && (best.0 == -1 || current.1 > best.1) {
        best = current;
    }
    if best.0 != -1 && best.1 < 2 {
        best.0 = -1;
    }
    let mut out = String::new();
    for (i, &word) in words.iter().enumerate() {
        let i = i as i32;
        if best.0 != -1 && i >= best.0 && i < best.0 + best.1 {
            if i == best.0 {
                out.push(':');
            }
            continue;
        }
        if i != 0 {
            out.push(':');
        }
        let embedded = i == 6
            && best.0 == 0
            && (best.1 == 6 || (best.1 == 7 && words[7] != 1) || (best.1 == 5 && words[5] == 0xffff));
        if embedded {
            let octets: Vec<String> = address[12..].iter().map(u8::to_string).collect();
            out.push_str(&octets.join("."));
            break;
        }
        out.push_str(&format!("{word:x}"));
    }
    if best.0 != -1 && best.0 + best.1 == 8 {
        out.push(':');
    }
    if bits != 128 {
        out.push_str(&format!("/{bits}"));
    }
    out
}

/// network_recv reads the binary format of an inet or cidr value, as network_recv does.
fn network_recv(bytes: &[u8], cidr: bool) -> Result<Vec<u8>> {
    let type_name = if cidr { "cidr" } else { "inet" };
    let error = |what: &str| {
        PgError::new(code::INVALID_BINARY_REPRESENTATION, format!("invalid {what} in external \"{type_name}\" value"))
    };
    let [family, bits, _, size, ..] = *bytes else {
        return Err(PgError::new(code::PROTOCOL_VIOLATION, "insufficient data left in message"));
    };
    if family != IPV4 && family != IPV6 {
        return Err(error("address family"));
    }
    if usize::from(bits) > address_size(family) * 8 {
        return Err(error("bits"));
    }
    if usize::from(size) != address_size(family) {
        return Err(error("length"));
    }
    let address = bytes
        .get(4..4 + usize::from(size))
        .ok_or_else(|| PgError::new(code::PROTOCOL_VIOLATION, "insufficient data left in message"))?;
    if cidr && !address_ok(address, u32::from(bits), family) {
        return Err(PgError {
            detail: Some("Value has bits set to right of mask.".into()),
            ..PgError::new(code::INVALID_BINARY_REPRESENTATION, "invalid external \"cidr\" value")
        });
    }
    let mut out = vec![family, bits, u8::from(cidr), size];
    out.extend_from_slice(address);
    Ok(out)
}

/// bits_cmp compares the first bits of two addresses, as bitncmp does.
fn bits_cmp(left: &[u8], right: &[u8], bits: u32) -> Ordering {
    let bytes = (bits / 8) as usize;
    let order = left[..bytes].cmp(&right[..bytes]);
    if order != Ordering::Equal || bits.is_multiple_of(8) {
        return order;
    }
    let mask = 0xffu8 << (8 - bits % 8);
    (left[bytes] & mask).cmp(&(right[bytes] & mask))
}

/// network_cmp orders inet and cidr values by family, network, mask length, and then the whole address, as
/// network_cmp does.
fn network_cmp(left: &[u8], right: &[u8]) -> Ordering {
    if left[0] != right[0] {
        return left[0].cmp(&right[0]);
    }
    let (left_bits, right_bits) = (u32::from(left[1]), u32::from(right[1]));
    bits_cmp(&left[4..], &right[4..], left_bits.min(right_bits))
        .then(left_bits.cmp(&right_bits))
        .then_with(|| bits_cmp(&left[4..], &right[4..], address_size(left[0]) as u32 * 8))
}

/// MAC_FORMATS are the formats macaddr_in tries in turn, written as scanf conversions: `x` reads hex digits of any
/// length, `2` reads at most two, and other bytes match themselves.
const MAC_FORMATS: [&str; 7] = ["x:x:x:x:x:x", "x-x-x-x-x-x", "222:222", "222-222", "22.22.22", "22-22-22", "222222"];

/// scan_hex reads a hex number as scanf's `%x` does, with at most the width of bytes after leading whitespace.
fn scan_hex(text: &[u8], i: &mut usize, width: Option<usize>) -> Option<u64> {
    while text.get(*i).is_some_and(u8::is_ascii_whitespace) {
        *i += 1;
    }
    let start = *i;
    let limit = width.map_or(text.len(), |w| (start + w).min(text.len()));
    let negative = matches!(text.get(*i), Some(b'-'));
    if matches!(text.get(*i), Some(b'+' | b'-')) && *i < limit {
        *i += 1;
    }
    if text.get(*i) == Some(&b'0')
        && matches!(text.get(*i + 1), Some(b'x' | b'X'))
        && *i + 2 < limit
        && text.get(*i + 2).is_some_and(u8::is_ascii_hexdigit)
    {
        *i += 2;
    }
    let digits = *i;
    let mut value: u64 = 0;
    while *i < limit && text.get(*i).is_some_and(u8::is_ascii_hexdigit) {
        value = value.saturating_mul(16).saturating_add(u64::from((text[*i] as char).to_digit(16)?));
        *i += 1;
    }
    if *i == digits {
        return None;
    }
    Some(if negative { value.wrapping_neg() & 0xffff_ffff } else { value })
}

/// macaddr_in reads a MAC address in any of the formats Postgres accepts.
fn macaddr_in(text: &str) -> Result<Vec<u8>> {
    let bytes = text.as_bytes();
    for format in MAC_FORMATS {
        let mut i = 0;
        let mut octets = Vec::with_capacity(6);
        let mut matched = true;
        for spec in format.bytes() {
            let parsed = match spec {
                b'x' => scan_hex(bytes, &mut i, None),
                b'2' => scan_hex(bytes, &mut i, Some(2)),
                literal if bytes.get(i) == Some(&literal) => {
                    i += 1;
                    continue;
                }
                _ => None,
            };
            match parsed {
                Some(value) => octets.push(value),
                None => {
                    matched = false;
                    break;
                }
            }
        }
        if !matched || octets.len() != 6 || bytes[i..].iter().any(|b| !b.is_ascii_whitespace()) {
            continue;
        }
        if octets.iter().any(|&o| o > 255) {
            return Err(PgError::new(
                code::NUMERIC_VALUE_OUT_OF_RANGE,
                format!("invalid octet value in \"macaddr\" value: \"{text}\""),
            ));
        }
        return Ok(octets.into_iter().map(|o| o as u8).collect());
    }
    Err(PgError::new(code::INVALID_TEXT_REPRESENTATION, format!("invalid input syntax for type macaddr: \"{text}\"")))
}
