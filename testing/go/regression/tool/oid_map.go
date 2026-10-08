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

package main

import (
	"strconv"
	"strings"
)

// minUserOID is the first OID that Postgres assigns to user-created objects. OIDs below this value belong to the
// system catalogs, which are expected to be stable, so we never map them.
const minUserOID = 16384

// OIDMap tracks the mapping from OIDs that were recorded in the original Postgres session to the OIDs that the
// Doltgres server assigned to the same objects during the replay. Clients (psql in particular) read OIDs from
// catalog queries and embed them verbatim in follow-up queries, so a replay against a server with different OID
// assignments must translate those embedded OIDs for the follow-ups to have any chance of succeeding.
//
// Mappings are learned during row comparison: whenever an `oid`-typed result column contains a user-object OID that
// differs between the recording and the replay (and the rest of the comparison succeeds), the pair is recorded.
type OIDMap struct {
	oids map[uint32]uint32
}

// NewOIDMap returns a new *OIDMap.
func NewOIDMap() *OIDMap {
	return &OIDMap{oids: make(map[uint32]uint32)}
}

// Get returns the replay OID that the given recorded OID maps to.
func (om *OIDMap) Get(recorded uint32) (uint32, bool) {
	mapped, ok := om.oids[recorded]
	return mapped, ok
}

// PutAll records all of the given recorded-to-replay OID pairs. Later learnings overwrite earlier ones, since a
// dropped and recreated object may reuse an OID on one side only.
func (om *OIDMap) PutAll(replacements map[uint32]uint32) {
	for recorded, actual := range replacements {
		om.oids[recorded] = actual
	}
}

// RewriteQuery replaces every standalone numeric token that matches a recorded OID with its replay OID. Tokens that
// are part of an identifier (adjacent to letters, digits, or underscores) are left untouched. Replacements are not
// rescanned, so a replacement value can never be mistaken for another recorded OID.
func (om *OIDMap) RewriteQuery(query string) string {
	if len(om.oids) == 0 {
		return query
	}
	var sb *strings.Builder
	last := 0
	for i := 0; i < len(query); {
		if !isDigit(query[i]) {
			i++
			continue
		}
		start := i
		for i < len(query) && isDigit(query[i]) {
			i++
		}
		// A user OID has at least 5 digits; also reject digit runs that are part of an identifier
		if i-start < 5 || i-start > 10 ||
			(start > 0 && isWordChar(query[start-1])) ||
			(i < len(query) && isWordChar(query[i])) {
			continue
		}
		parsed, err := strconv.ParseUint(query[start:i], 10, 32)
		if err != nil {
			continue
		}
		mapped, ok := om.oids[uint32(parsed)]
		if !ok {
			continue
		}
		if sb == nil {
			sb = &strings.Builder{}
			sb.Grow(len(query) + 16)
		}
		sb.WriteString(query[last:start])
		sb.WriteString(strconv.FormatUint(uint64(mapped), 10))
		last = i
	}
	if sb == nil {
		return query
	}
	sb.WriteString(query[last:])
	return sb.String()
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isWordChar(c byte) bool {
	return isDigit(c) || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

// cellToOID interprets a decoded result cell as an OID if possible. Cells arrive as whatever type the pgtype scan
// produced, which differs between the recorded Postgres response and the Doltgres response, so this accepts every
// integer representation along with numeric strings.
func cellToOID(cell interface{}) (uint32, bool) {
	switch val := cell.(type) {
	case uint32:
		return val, true
	case int64:
		if val >= 0 && val <= 4294967295 {
			return uint32(val), true
		}
	case uint64:
		if val <= 4294967295 {
			return uint32(val), true
		}
	case int32:
		if val >= 0 {
			return uint32(val), true
		}
	case string:
		parsed, err := strconv.ParseUint(val, 10, 32)
		if err == nil {
			return uint32(parsed), true
		}
	}
	return 0, false
}
