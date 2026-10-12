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

use std::net::{Ipv4Addr, Ipv6Addr};

use crate::decode::reader::Reader;

/// inet_text renders a binary inet or cidr. An inet hides its prefix length when it covers the whole address, and a
/// cidr always shows it.
pub(crate) fn inet_text(oid: u32, r: &mut Reader<'_>) -> Result<String, String> {
    let family = r.u8()?;
    let bits = r.u8()?;
    let _is_cidr = r.u8()?;
    let length = r.u8()? as usize;
    let address = r.take(length)?;
    let (text, max_bits) = match (family, length) {
        (2, 4) => (Ipv4Addr::new(address[0], address[1], address[2], address[3]).to_string(), 32),
        (3, 16) => {
            let octets: [u8; 16] = address.try_into().unwrap();
            (Ipv6Addr::from(octets).to_string(), 128)
        }
        _ => return Err(format!("invalid inet family {family} with {length} address bytes")),
    };
    if oid == 869 && bits == max_bits { Ok(text) } else { Ok(format!("{text}/{bits}")) }
}

/// macaddr_text renders a macaddr or macaddr8 as colon-separated hexadecimal bytes.
pub(crate) fn macaddr_text(bytes: &[u8]) -> String {
    bytes.iter().map(|b| format!("{b:02x}")).collect::<Vec<_>>().join(":")
}
