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

package _go

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestConvertTo(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "convert_to",
			SetUpScript: []string{
				`CREATE TABLE supported_encodings (name TEXT PRIMARY KEY);`,
				`INSERT INTO supported_encodings VALUES ` +
					`('SQL_ASCII'),('SQLASCII'),('EUC_JP'),('EUC_KR'),('UTF8'),('utf-8'),('UNICODE'),` +
					`('LATIN1'),('ISO_8859_1'),('LATIN2'),('ISO_8859_2'),('LATIN3'),('ISO_8859_3'),` +
					`('LATIN4'),('ISO_8859_4'),('LATIN5'),('ISO_8859_9'),('LATIN6'),('ISO_8859_10'),` +
					`('LATIN7'),('ISO_8859_13'),('LATIN8'),('ISO_8859_14'),('LATIN9'),('ISO_8859_15'),` +
					`('LATIN10'),('ISO_8859_16'),('WIN1256'),('WIN1258'),('WIN866'),('ALT'),('WIN874'),` +
					`('KOI8R'),('WIN1251'),('WIN'),('WIN1252'),('ISO_8859_5'),('ISO_8859_6'),('ISO_8859_7'),` +
					`('ISO_8859_8'),('WIN1250'),('WIN1253'),('WIN1254'),('WIN1255'),('WIN1257'),('KOI8U'),` +
					`('SJIS'),('SHIFT_JIS'),('BIG5'),('GBK'),('UHC'),('GB18030'),` +
					`('WINDOWS-866'),('WINDOWS-874'),('WINDOWS-1250'),('WINDOWS-1251'),('WINDOWS-1252'),` +
					`('WINDOWS-1253'),('WINDOWS-1254'),('WINDOWS-1255'),('WINDOWS-1256'),('WINDOWS-1257'),` +
					`('WINDOWS-1258');`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT count(*), bool_and(pg_char_to_encoding(name) >= 0), ` +
						`bool_and(encode(convert_to('x', name::name), 'hex') = '78') FROM supported_encodings;`,
					Expected: []sql.Row{{int64(63), "t", "t"}},
				},
				{
					Query:    `SELECT encode(convert_to('x', 'UTF8'), 'hex');`,
					Expected: []sql.Row{{"78"}},
				},
				{
					Query:    `SELECT encode(convert_to('café', 'utf-8'), 'hex');`,
					Expected: []sql.Row{{"636166c3a9"}},
				},
				{
					Query:    `SELECT encode(convert_to('日本', 'EUC_JP'), 'hex');`,
					Expected: []sql.Row{{"c6fccbdc"}},
				},
				{
					Query: `SELECT encode(convert_to('한', 'UHC'), 'hex'), encode(convert_to('Ж', 'ALT'), 'hex'), ` +
						`encode(convert_to('Ж', 'WIN'), 'hex'), encode(convert_to('Ж', 'ISO_8859_5'), 'hex');`,
					Expected: []sql.Row{{"c7d1", "86", "c6", "b6"}},
				},
				{
					Query: `SELECT encode(convert_to('Ж', 'KOI8R'), 'hex'), encode(convert_to('日本', 'SJIS'), 'hex'), ` +
						`encode(convert_to('한', 'EUC_KR'), 'hex'), encode(convert_to('한', 'UHC'), 'hex'), ` +
						`encode(convert_to('中文', 'BIG5'), 'hex'), encode(convert_to('中文', 'GBK'), 'hex'), ` +
						`encode(convert_to('😄', 'GB18030'), 'hex');`,
					Expected: []sql.Row{{"f6", "93fa967b", "c7d1", "c7d1", "a4a4a4e5", "d6d0cec4", "9439fd30"}},
				},
				{
					Query: `SELECT encode(convert_to('€', 'WIN1252'), 'hex'), ` +
						`encode(convert_to('€', 'WINDOWS-1252'), 'hex'), encode(convert_to('Ж', 'WIN866'), 'hex'), ` +
						`encode(convert_to('Ж', 'WINDOWS-866'), 'hex');`,
					Expected: []sql.Row{{"80", "80", "86", "86"}},
				},
				{
					Query: `SELECT encode(convert_to('café', 'LATIN1'), 'hex'), ` +
						`encode(convert_to('café', 'ISO_8859_1'), 'hex'), encode(convert_to('日本', 'SJIS'), 'hex'), ` +
						`encode(convert_to('日本', 'SHIFT_JIS'), 'hex');`,
					Expected: []sql.Row{{"636166e9", "636166e9", "93fa967b", "93fa967b"}},
				},
				{
					Query: `SELECT pg_char_to_encoding('ALT'), pg_char_to_encoding('WIN'), pg_char_to_encoding('UHC'), ` +
						`pg_char_to_encoding('ISO_8859_5'), pg_char_to_encoding('EUC_CN');`,
					Expected: []sql.Row{{int32(20), int32(23), int32(38), int32(25), int32(2)}},
				},
				{
					Query: `SELECT pg_char_to_encoding('WINDOWS-866'), pg_char_to_encoding('WINDOWS-874'), ` +
						`pg_char_to_encoding('WINDOWS-1250'), pg_char_to_encoding('WINDOWS-1251'), ` +
						`pg_char_to_encoding('WINDOWS-1252'), pg_char_to_encoding('WINDOWS-1253'), ` +
						`pg_char_to_encoding('WINDOWS-1254'), pg_char_to_encoding('WINDOWS-1255'), ` +
						`pg_char_to_encoding('WINDOWS-1256'), pg_char_to_encoding('WINDOWS-1257'), ` +
						`pg_char_to_encoding('WINDOWS-1258');`,
					Expected: []sql.Row{{int32(20), int32(21), int32(29), int32(23), int32(24), int32(30),
						int32(31), int32(32), int32(18), int32(33), int32(19)}},
				},
				{
					Query:    `SELECT encode(convert_to('', 'UTF8'), 'hex');`,
					Expected: []sql.Row{{""}},
				},
				{
					Query:    `SELECT convert_to(NULL, 'UTF8'), convert_to('x', NULL);`,
					Expected: []sql.Row{{nil, nil}},
				},
				{
					Query:           `SELECT convert_to('€', 'LATIN1');`,
					ExpectedErr:     `character with byte sequence 0xe2 0x82 0xac in encoding "UTF8" has no equivalent in encoding "LATIN1"`,
					ExpectedErrCode: "22P05",
				},
				{
					Query:           `SELECT convert_to('x', 'EUC_CN');`,
					ExpectedErr:     `destination encoding "EUC_CN" is recognized but not yet supported; request support at https://github.com/dolthub/doltgresql/issues`,
					ExpectedErrCode: "0A000",
				},
				{
					Query:           `SELECT convert_to('x', 'bogus');`,
					ExpectedErr:     `invalid destination encoding name "bogus"`,
					ExpectedErrCode: "22023",
				},
			},
		},
	})
}
