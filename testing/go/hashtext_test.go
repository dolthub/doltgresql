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

func TestHashText(t *testing.T) {
	RunScripts(t, []ScriptTest{
		{
			Name: "hashtext",
			SetUpScript: []string{
				`CREATE TABLE hashtext_values (id INT PRIMARY KEY, value TEXT);`,
				`INSERT INTO hashtext_values VALUES (1, repeat('x', 20000));`,
				`CREATE TABLE hashtext_lengths (length INT PRIMARY KEY);`,
				`INSERT INTO hashtext_lengths VALUES (0), (1), (2), (3), (4), (5), (6), (7), (8), (9), (10), ` +
					`(11), (12), (13), (23), (24), (25);`,
			},
			Assertions: []ScriptTestAssertion{
				{
					Query: `SELECT length, hashtext(repeat('x', length)) FROM hashtext_lengths ORDER BY length;`,
					Expected: []sql.Row{
						{int32(0), int32(-1477818771)},
						{int32(1), int32(1074944137)},
						{int32(2), int32(-1086392228)},
						{int32(3), int32(-1992236649)},
						{int32(4), int32(-1379736791)},
						{int32(5), int32(-370454118)},
						{int32(6), int32(1489915569)},
						{int32(7), int32(-66683019)},
						{int32(8), int32(-2126973000)},
						{int32(9), int32(1651296771)},
						{int32(10), int32(755764456)},
						{int32(11), int32(-1494243903)},
						{int32(12), int32(631527812)},
						{int32(13), int32(28686851)},
						{int32(23), int32(597544042)},
						{int32(24), int32(1380215333)},
						{int32(25), int32(733930510)},
					},
				},
				{
					Query:    `SELECT hashtext('');`,
					Expected: []sql.Row{{int32(-1477818771)}},
				},
				{
					Query:    `SELECT hashtext('abc');`,
					Expected: []sql.Row{{int32(-785388649)}},
				},
				{
					Query:    `SELECT hashtext('12345678901'), hashtext('123456789012'), hashtext('1234567890123');`,
					Expected: []sql.Row{{int32(1650060602), int32(-2102057603), int32(437480032)}},
				},
				{
					Query:    `SELECT octet_length('12345678901é'), hashtext('12345678901é');`,
					Expected: []sql.Row{{int32(13), int32(-295778157)}},
				},
				{
					Query:    `SELECT hashtext('café'), hashtext('日本');`,
					Expected: []sql.Row{{int32(103771354), int32(-1851216170)}},
				},
				{
					Query:    `SELECT hashtext(repeat('x', 1000));`,
					Expected: []sql.Row{{int32(-1157355676)}},
				},
				{
					Query:    `SELECT hashtext(value) FROM hashtext_values WHERE id = 1;`,
					Expected: []sql.Row{{int32(-1519670832)}},
				},
				{
					Query:    `SELECT hashtext(NULL);`,
					Expected: []sql.Row{{nil}},
				},
			},
		},
	})
}
