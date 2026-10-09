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

//! The built-in base types that Go lacks, held as `Value::Base` values whose stored bytes are Postgres' binary
//! format: the geometric types, the network address types, money, pg_lsn, and the text search types.

mod geometric;
mod money;
mod network;
mod textsearch;
mod tid;

use crate::extensions::BaseType;

/// get returns the definition of a built-in base type by its OID.
pub fn get(type_oid: u32) -> Option<&'static BaseType> {
    Some(match type_oid {
        27 => &tid::TID,
        600 => &geometric::POINT,
        601 => &geometric::LSEG,
        602 => &geometric::PATH,
        603 => &geometric::BOX,
        604 => &geometric::POLYGON,
        628 => &geometric::LINE,
        650 => &network::CIDR,
        718 => &geometric::CIRCLE,
        790 => &money::MONEY,
        829 => &network::MACADDR,
        869 => &network::INET,
        3220 => &money::PG_LSN,
        3614 => &textsearch::TSVECTOR,
        3615 => &textsearch::TSQUERY,
        _ => return None,
    })
}
