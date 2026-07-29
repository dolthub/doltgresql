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

package id

// init registers the OIDs for the built-in operator families and operator classes. Operator families and classes are
// identified by two segments: the access method name and the family/class name.
//
// Operator family OIDs match the fixed OIDs assigned by Postgres (pg_opfamily.dat). Operator class OIDs in Postgres
// are assigned dynamically during initdb and are not stable across versions, so Doltgres assigns its own fixed OIDs
// in the 15000+ range (which no other built-in assignments use).
func init() {
	// btree operator families
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "btree", "array_ops"), 397)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "btree", "bit_ops"), 423)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "btree", "bool_ops"), 424)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "btree", "bpchar_ops"), 426)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "btree", "bytea_ops"), 428)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "btree", "char_ops"), 429)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "btree", "datetime_ops"), 434)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "btree", "float_ops"), 1970)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "btree", "network_ops"), 1974)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "btree", "integer_ops"), 1976)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "btree", "interval_ops"), 1982)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "btree", "numeric_ops"), 1988)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "btree", "oid_ops"), 1989)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "btree", "text_ops"), 1994)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "btree", "time_ops"), 1996)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "btree", "timetz_ops"), 2000)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "btree", "varbit_ops"), 2002)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "btree", "uuid_ops"), 2968)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "btree", "record_ops"), 2994)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "btree", "jsonb_ops"), 4033)

	// hash operator families
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "hash", "bpchar_ops"), 427)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "hash", "char_ops"), 431)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "hash", "datetime_ops"), 435)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "hash", "bytea_ops"), 624)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "hash", "float_ops"), 1971)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "hash", "integer_ops"), 1977)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "hash", "oid_ops"), 1990)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "hash", "text_ops"), 1995)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "hash", "numeric_ops"), 1998)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "hash", "bool_ops"), 2222)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "hash", "uuid_ops"), 2969)
	globalCache.setBuiltIn(NewId(Section_OperatorFamily, "hash", "jsonb_ops"), 4034)

	// btree operator classes (Doltgres-assigned OIDs, see comment above)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "array_ops"), 15000)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "bool_ops"), 15001)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "bpchar_ops"), 15002)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "bytea_ops"), 15003)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "char_ops"), 15004)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "date_ops"), 15005)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "float4_ops"), 15006)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "float8_ops"), 15007)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "int2_ops"), 15008)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "int4_ops"), 15009)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "int8_ops"), 15010)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "interval_ops"), 15011)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "jsonb_ops"), 15012)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "name_ops"), 15013)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "numeric_ops"), 15014)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "oid_ops"), 15015)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "record_ops"), 15016)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "text_ops"), 15017)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "time_ops"), 15018)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "timestamp_ops"), 15019)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "timestamptz_ops"), 15020)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "timetz_ops"), 15021)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "uuid_ops"), 15022)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "btree", "varchar_ops"), 15023)

	// hash operator classes (Doltgres-assigned OIDs, see comment above)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "hash", "bool_ops"), 15100)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "hash", "bpchar_ops"), 15101)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "hash", "bytea_ops"), 15102)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "hash", "char_ops"), 15103)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "hash", "date_ops"), 15104)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "hash", "float4_ops"), 15105)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "hash", "float8_ops"), 15106)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "hash", "int2_ops"), 15107)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "hash", "int4_ops"), 15108)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "hash", "int8_ops"), 15109)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "hash", "jsonb_ops"), 15110)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "hash", "numeric_ops"), 15111)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "hash", "oid_ops"), 15112)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "hash", "text_ops"), 15113)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "hash", "timestamp_ops"), 15114)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "hash", "timestamptz_ops"), 15115)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "hash", "uuid_ops"), 15116)
	globalCache.setBuiltIn(NewId(Section_OperatorClass, "hash", "varchar_ops"), 15117)
}
