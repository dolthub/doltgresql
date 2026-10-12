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

/// The text format code.
pub const TEXT: i16 = 0;
/// The binary format code.
pub const BINARY: i16 = 1;

/// BINARY_OIDS lists every type OID that pgx v5.9.2's default type map transfers in the binary format. It was
/// produced by asking pgtype.Map.FormatCodeForOID about every OID below 100000, so every other OID uses text.
const BINARY_OIDS: [u32; 94] = [
    16, 17, 18, 20, 21, 23, 26, 27, 28, 29, 143, 199, 271, 600, 601, 602, 603, 604, 628, 629, 650, 651, 700, 701, 718,
    719, 774, 829, 869, 1000, 1001, 1002, 1003, 1005, 1007, 1009, 1010, 1011, 1012, 1014, 1015, 1016, 1017, 1018, 1019,
    1020, 1021, 1022, 1027, 1028, 1040, 1041, 1082, 1083, 1114, 1115, 1182, 1183, 1184, 1185, 1186, 1187, 1231, 1560,
    1561, 1562, 1563, 1700, 2249, 2287, 2950, 2951, 3614, 3643, 3807, 3904, 3905, 3906, 3907, 3908, 3909, 3910, 3911,
    3912, 3913, 3926, 3927, 4451, 4532, 4533, 4534, 4535, 4536, 5069,
];

/// format_code_for_oid returns the format that pgx requests for results of the type, and prefers for parameters of
/// the type.
pub fn format_code_for_oid(oid: u32) -> i16 {
    if BINARY_OIDS.binary_search(&oid).is_ok() { BINARY } else { TEXT }
}
