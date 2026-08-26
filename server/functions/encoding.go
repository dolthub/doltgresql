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

package functions

import (
	"strings"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

// postgresEncoding describes a PostgreSQL encoding name. An entry with neither passThrough nor encoder set is a valid
// PostgreSQL encoding that DoltgreSQL recognizes but does not support as a conversion destination yet.
type postgresEncoding struct {
	name        string
	id          int32
	aliases     []string
	encoder     encoding.Encoding
	passThrough bool
}

// postgresEncodings is the shared source of truth for encoding-aware functions. Conversion support is intentionally
// curated: UTF8/SQL_ASCII, ISO-8859/Latin, Windows and KOI8 single-byte encodings, and the commonly used Japanese,
// Korean, and Chinese encodings available in x/text. The remaining PostgreSQL encodings stay registered so callers get
// an actionable unsupported-feature error instead of being told that a valid encoding name is invalid.
var postgresEncodings = []postgresEncoding{
	{name: "SQL_ASCII", id: 0, aliases: []string{"SQLASCII"}, passThrough: true},
	{name: "EUC_JP", id: 1, encoder: japanese.EUCJP},
	{name: "EUC_CN", id: 2},
	{name: "EUC_KR", id: 3, encoder: korean.EUCKR},
	{name: "EUC_TW", id: 4},
	{name: "EUC_JIS_2004", id: 5},
	{name: "UTF8", id: 6, aliases: []string{"UNICODE"}, passThrough: true},
	{name: "MULE_INTERNAL", id: 7},
	{name: "LATIN1", id: 8, aliases: []string{"ISO_8859_1"}, encoder: charmap.ISO8859_1},
	{name: "LATIN2", id: 9, aliases: []string{"ISO_8859_2"}, encoder: charmap.ISO8859_2},
	{name: "LATIN3", id: 10, aliases: []string{"ISO_8859_3"}, encoder: charmap.ISO8859_3},
	{name: "LATIN4", id: 11, aliases: []string{"ISO_8859_4"}, encoder: charmap.ISO8859_4},
	{name: "LATIN5", id: 12, aliases: []string{"ISO_8859_9"}, encoder: charmap.ISO8859_9},
	{name: "LATIN6", id: 13, aliases: []string{"ISO_8859_10"}, encoder: charmap.ISO8859_10},
	{name: "LATIN7", id: 14, aliases: []string{"ISO_8859_13"}, encoder: charmap.ISO8859_13},
	{name: "LATIN8", id: 15, aliases: []string{"ISO_8859_14"}, encoder: charmap.ISO8859_14},
	{name: "LATIN9", id: 16, aliases: []string{"ISO_8859_15"}, encoder: charmap.ISO8859_15},
	{name: "LATIN10", id: 17, aliases: []string{"ISO_8859_16"}, encoder: charmap.ISO8859_16},
	{name: "WIN1256", id: 18, aliases: []string{"WINDOWS_1256"}, encoder: charmap.Windows1256},
	{name: "WIN1258", id: 19, aliases: []string{"WINDOWS_1258"}, encoder: charmap.Windows1258},
	{name: "WIN866", id: 20, aliases: []string{"ALT", "WINDOWS_866"}, encoder: charmap.CodePage866},
	{name: "WIN874", id: 21, aliases: []string{"WINDOWS_874"}, encoder: charmap.Windows874},
	{name: "KOI8R", id: 22, encoder: charmap.KOI8R},
	{name: "WIN1251", id: 23, aliases: []string{"WIN", "WINDOWS_1251"}, encoder: charmap.Windows1251},
	{name: "WIN1252", id: 24, aliases: []string{"WINDOWS_1252"}, encoder: charmap.Windows1252},
	{name: "ISO_8859_5", id: 25, encoder: charmap.ISO8859_5},
	{name: "ISO_8859_6", id: 26, encoder: charmap.ISO8859_6},
	{name: "ISO_8859_7", id: 27, encoder: charmap.ISO8859_7},
	{name: "ISO_8859_8", id: 28, encoder: charmap.ISO8859_8},
	{name: "WIN1250", id: 29, aliases: []string{"WINDOWS_1250"}, encoder: charmap.Windows1250},
	{name: "WIN1253", id: 30, aliases: []string{"WINDOWS_1253"}, encoder: charmap.Windows1253},
	{name: "WIN1254", id: 31, aliases: []string{"WINDOWS_1254"}, encoder: charmap.Windows1254},
	{name: "WIN1255", id: 32, aliases: []string{"WINDOWS_1255"}, encoder: charmap.Windows1255},
	{name: "WIN1257", id: 33, aliases: []string{"WINDOWS_1257"}, encoder: charmap.Windows1257},
	{name: "KOI8U", id: 34, encoder: charmap.KOI8U},
	{name: "SJIS", id: 35, aliases: []string{"SHIFT_JIS"}, encoder: japanese.ShiftJIS},
	{name: "BIG5", id: 36, encoder: traditionalchinese.Big5},
	{name: "GBK", id: 37, encoder: simplifiedchinese.GBK},
	{name: "UHC", id: 38, encoder: korean.EUCKR},
	{name: "GB18030", id: 39, encoder: simplifiedchinese.GB18030},
	{name: "JOHAB", id: 40},
	{name: "SHIFT_JIS_2004", id: 41},
}

var postgresEncodingByName = func() map[string]*postgresEncoding {
	encodings := make(map[string]*postgresEncoding, len(postgresEncodings)*2)
	for i := range postgresEncodings {
		definition := &postgresEncodings[i]
		encodings[normalizeEncodingName(definition.name)] = definition
		for _, alias := range definition.aliases {
			encodings[normalizeEncodingName(alias)] = definition
		}
	}
	return encodings
}()

// lookupPostgresEncoding returns the PostgreSQL encoding matching name or one of its aliases.
func lookupPostgresEncoding(name string) *postgresEncoding {
	return postgresEncodingByName[normalizeEncodingName(name)]
}

// normalizeEncodingName matches PostgreSQL's case-insensitive treatment of encoding names and aliases.
func normalizeEncodingName(name string) string {
	var sb strings.Builder
	for _, r := range strings.ToLower(name) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			sb.WriteRune(r)
		}
	}
	return sb.String()
}
