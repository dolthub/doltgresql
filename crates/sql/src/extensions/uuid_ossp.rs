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

//! The uuid-ossp extension, version 1.1, which generates universally unique identifiers.

use super::{Control, Extension, Implementation, Routine};
use crate::catalog::ColumnType;
use crate::error::Result;
use crate::query::Ctx;
use crate::types::Value;

/// The RFC 4122 namespace identifiers.
const NAMESPACE_DNS: [u8; 16] = *b"\x6b\xa7\xb8\x10\x9d\xad\x11\xd1\x80\xb4\x00\xc0\x4f\xd4\x30\xc8";
const NAMESPACE_URL: [u8; 16] = *b"\x6b\xa7\xb8\x11\x9d\xad\x11\xd1\x80\xb4\x00\xc0\x4f\xd4\x30\xc8";
const NAMESPACE_OID: [u8; 16] = *b"\x6b\xa7\xb8\x12\x9d\xad\x11\xd1\x80\xb4\x00\xc0\x4f\xd4\x30\xc8";
const NAMESPACE_X500: [u8; 16] = *b"\x6b\xa7\xb8\x14\x9d\xad\x11\xd1\x80\xb4\x00\xc0\x4f\xd4\x30\xc8";

/// extension returns the definition of the extension.
pub fn extension() -> Extension {
    let routine = |name: &str, params: Vec<(&'static str, String)>, implementation: Implementation| Routine {
        name: name.into(),
        symbol: name.into(),
        params,
        returns: "uuid".into(),
        strict: true,
        implementation,
    };
    let namespace_params = || vec![("namespace", "uuid".to_string()), ("name", "text".to_string())];
    Extension {
        name: "uuid-ossp",
        control: Control {
            default_version: "1.1",
            comment: "generate universally unique identifiers (UUIDs)",
            superuser: true,
            trusted: true,
            relocatable: true,
        },
        types: Vec::new(),
        routines: vec![
            routine("uuid_nil", Vec::new(), |_, _, _| Ok(Value::Uuid([0; 16]))),
            routine("uuid_ns_dns", Vec::new(), |_, _, _| Ok(Value::Uuid(NAMESPACE_DNS))),
            routine("uuid_ns_url", Vec::new(), |_, _, _| Ok(Value::Uuid(NAMESPACE_URL))),
            routine("uuid_ns_oid", Vec::new(), |_, _, _| Ok(Value::Uuid(NAMESPACE_OID))),
            routine("uuid_ns_x500", Vec::new(), |_, _, _| Ok(Value::Uuid(NAMESPACE_X500))),
            routine("uuid_generate_v1", Vec::new(), generate_v1),
            routine("uuid_generate_v1mc", Vec::new(), generate_v1),
            routine("uuid_generate_v3", namespace_params(), generate_v3),
            routine("uuid_generate_v4", Vec::new(), generate_v4),
            routine("uuid_generate_v5", namespace_params(), generate_v5),
        ],
        operators: Vec::new(),
        casts: Vec::new(),
        aggregates: Vec::new(),
        operator_classes: Vec::new(),
        access_methods: Vec::new(),
    }
}

/// versioned sets a uuid's version and its RFC 4122 variant.
fn versioned(mut uuid: [u8; 16], version: u8) -> Value {
    uuid[6] = (uuid[6] & 0x0f) | (version << 4);
    uuid[8] = (uuid[8] & 0x3f) | 0x80;
    Value::Uuid(uuid)
}

/// generate_v1 returns a version 1 uuid from the time, a random clock sequence, and a random multicast node, which
/// serves both uuid_generate_v1 and uuid_generate_v1mc since Doltgres never reveals a MAC address.
fn generate_v1(_: &mut Ctx<'_>, _: &[Value], _: ColumnType) -> Result<Value> {
    const GREGORIAN_OFFSET: u64 = 0x01b2_1dd2_1381_4000;
    let since_epoch = std::time::SystemTime::now().duration_since(std::time::UNIX_EPOCH).unwrap_or_default();
    let ticks = GREGORIAN_OFFSET + since_epoch.as_nanos() as u64 / 100;
    let mut uuid: [u8; 16] = rand::random();
    uuid[..4].copy_from_slice(&(ticks as u32).to_be_bytes());
    uuid[4..6].copy_from_slice(&((ticks >> 32) as u16).to_be_bytes());
    uuid[6..8].copy_from_slice(&((ticks >> 48) as u16).to_be_bytes());
    uuid[10] |= 0x03;
    Ok(versioned(uuid, 1))
}

/// name_input returns the bytes that versions 3 and 5 hash: a namespace uuid followed by a name.
fn name_input(args: &[Value]) -> Vec<u8> {
    let mut input = match &args[0] {
        Value::Uuid(namespace) => namespace.to_vec(),
        _ => Vec::new(),
    };
    if let Value::Text(name) = &args[1] {
        input.extend_from_slice(name.as_bytes());
    }
    input
}

/// generate_v3 returns a version 3 uuid from the MD5 hash of a namespace and a name.
fn generate_v3(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    use md5::{Digest, Md5};
    let hash = Md5::digest(name_input(args));
    Ok(versioned(hash.into(), 3))
}

/// generate_v4 returns a random version 4 uuid.
fn generate_v4(_: &mut Ctx<'_>, _: &[Value], _: ColumnType) -> Result<Value> {
    Ok(versioned(rand::random(), 4))
}

/// generate_v5 returns a version 5 uuid from the SHA-1 hash of a namespace and a name.
fn generate_v5(_: &mut Ctx<'_>, args: &[Value], _: ColumnType) -> Result<Value> {
    use sha1::{Digest, Sha1};
    let hash = Sha1::digest(name_input(args));
    let mut uuid = [0u8; 16];
    uuid.copy_from_slice(&hash[..16]);
    Ok(versioned(uuid, 5))
}
