// Copyright 2024 Dolthub, Inc.
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

package regression

import (
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
)

func TestMoney(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_money)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_money,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE TABLE money_data (m money);`,
			},
			{
				Statement: `INSERT INTO money_data VALUES ('123');`,
			},
			{
				Statement: `SELECT * FROM money_data;`,
				Results:   []sql.Row{{`$123.00`}},
			},
			{
				Statement: `SELECT m + '123' FROM money_data;`,
				Results:   []sql.Row{{`$246.00`}},
			},
			{
				Statement: `SELECT m + '123.45' FROM money_data;`,
				Results:   []sql.Row{{`$246.45`}},
			},
			{
				Statement: `SELECT m - '123.45' FROM money_data;`,
				Results:   []sql.Row{{`-$0.45`}},
			},
			{
				Statement: `SELECT m / '2'::money FROM money_data;`,
				Results:   []sql.Row{{61.5}},
			},
			{
				Statement: `SELECT m * 2 FROM money_data;`,
				Results:   []sql.Row{{`$246.00`}},
			},
			{
				Statement: `SELECT 2 * m FROM money_data;`,
				Results:   []sql.Row{{`$246.00`}},
			},
			{
				Statement: `SELECT m / 2 FROM money_data;`,
				Results:   []sql.Row{{`$61.50`}},
			},
			{
				Statement: `SELECT m * 2::int2 FROM money_data;`,
				Results:   []sql.Row{{`$246.00`}},
			},
			{
				Statement: `SELECT 2::int2 * m FROM money_data;`,
				Results:   []sql.Row{{`$246.00`}},
			},
			{
				Statement: `SELECT m / 2::int2 FROM money_data;`,
				Results:   []sql.Row{{`$61.50`}},
			},
			{
				Statement: `SELECT m * 2::int8 FROM money_data;`,
				Results:   []sql.Row{{`$246.00`}},
			},
			{
				Statement: `SELECT 2::int8 * m FROM money_data;`,
				Results:   []sql.Row{{`$246.00`}},
			},
			{
				Statement: `SELECT m / 2::int8 FROM money_data;`,
				Results:   []sql.Row{{`$61.50`}},
			},
			{
				Statement: `SELECT m * 2::float8 FROM money_data;`,
				Results:   []sql.Row{{`$246.00`}},
			},
			{
				Statement: `SELECT 2::float8 * m FROM money_data;`,
				Results:   []sql.Row{{`$246.00`}},
			},
			{
				Statement: `SELECT m / 2::float8 FROM money_data;`,
				Results:   []sql.Row{{`$61.50`}},
			},
			{
				Statement: `SELECT m * 2::float4 FROM money_data;`,
				Results:   []sql.Row{{`$246.00`}},
			},
			{
				Statement: `SELECT 2::float4 * m FROM money_data;`,
				Results:   []sql.Row{{`$246.00`}},
			},
			{
				Statement: `SELECT m / 2::float4 FROM money_data;`,
				Results:   []sql.Row{{`$61.50`}},
			},
			{
				Statement: `SELECT m = '$123.00' FROM money_data;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT m != '$124.00' FROM money_data;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT m <= '$123.00' FROM money_data;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT m >= '$123.00' FROM money_data;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT m < '$124.00' FROM money_data;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT m > '$122.00' FROM money_data;`,
				Results:   []sql.Row{{true}},
			},
			{
				Statement: `SELECT m = '$123.01' FROM money_data;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT m != '$123.00' FROM money_data;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT m <= '$122.99' FROM money_data;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT m >= '$123.01' FROM money_data;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT m > '$124.00' FROM money_data;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT m < '$122.00' FROM money_data;`,
				Results:   []sql.Row{{false}},
			},
			{
				Statement: `SELECT cashlarger(m, '$124.00') FROM money_data;`,
				Results:   []sql.Row{{`$124.00`}},
			},
			{
				Statement: `SELECT cashsmaller(m, '$124.00') FROM money_data;`,
				Results:   []sql.Row{{`$123.00`}},
			},
			{
				Statement: `SELECT cash_words(m) FROM money_data;`,
				Results:   []sql.Row{{`One hundred twenty three dollars and zero cents`}},
			},
			{
				Statement: `SELECT cash_words(m + '1.23') FROM money_data;`,
				Results:   []sql.Row{{`One hundred twenty four dollars and twenty three cents`}},
			},
			{
				Statement: `DELETE FROM money_data;`,
			},
			{
				Statement: `INSERT INTO money_data VALUES ('$123.45');`,
			},
			{
				Statement: `SELECT * FROM money_data;`,
				Results:   []sql.Row{{`$123.45`}},
			},
			{
				Statement: `DELETE FROM money_data;`,
			},
			{
				Statement: `INSERT INTO money_data VALUES ('$123.451');`,
			},
			{
				Statement: `SELECT * FROM money_data;`,
				Results:   []sql.Row{{`$123.45`}},
			},
			{
				Statement: `DELETE FROM money_data;`,
			},
			{
				Statement: `INSERT INTO money_data VALUES ('$123.454');`,
			},
			{
				Statement: `SELECT * FROM money_data;`,
				Results:   []sql.Row{{`$123.45`}},
			},
			{
				Statement: `DELETE FROM money_data;`,
			},
			{
				Statement: `INSERT INTO money_data VALUES ('$123.455');`,
			},
			{
				Statement: `SELECT * FROM money_data;`,
				Results:   []sql.Row{{`$123.46`}},
			},
			{
				Statement: `DELETE FROM money_data;`,
			},
			{
				Statement: `INSERT INTO money_data VALUES ('$123.456');`,
			},
			{
				Statement: `SELECT * FROM money_data;`,
				Results:   []sql.Row{{`$123.46`}},
			},
			{
				Statement: `DELETE FROM money_data;`,
			},
			{
				Statement: `INSERT INTO money_data VALUES ('$123.459');`,
			},
			{
				Statement: `SELECT * FROM money_data;`,
				Results:   []sql.Row{{`$123.46`}},
			},
			{
				Statement: `SELECT '1234567890'::money;`,
				Results:   []sql.Row{{`$1,234,567,890.00`}},
			},
			{
				Statement: `SELECT '12345678901234567'::money;`,
				Results:   []sql.Row{{`$12,345,678,901,234,567.00`}},
			},
			{
				Statement:   `SELECT '123456789012345678'::money;`,
				ErrorString: `value "123456789012345678" is out of range for type money`,
			},
			{
				Statement:   `SELECT '9223372036854775807'::money;`,
				ErrorString: `value "9223372036854775807" is out of range for type money`,
			},
			{
				Statement: `SELECT '-12345'::money;`,
				Results:   []sql.Row{{`-$12,345.00`}},
			},
			{
				Statement: `SELECT '-1234567890'::money;`,
				Results:   []sql.Row{{`-$1,234,567,890.00`}},
			},
			{
				Statement: `SELECT '-12345678901234567'::money;`,
				Results:   []sql.Row{{`-$12,345,678,901,234,567.00`}},
			},
			{
				Statement:   `SELECT '-123456789012345678'::money;`,
				ErrorString: `value "-123456789012345678" is out of range for type money`,
			},
			{
				Statement:   `SELECT '-9223372036854775808'::money;`,
				ErrorString: `value "-9223372036854775808" is out of range for type money`,
			},
			{
				Statement: `SELECT '(1)'::money;`,
				Results:   []sql.Row{{`-$1.00`}},
			},
			{
				Statement: `SELECT '($123,456.78)'::money;`,
				Results:   []sql.Row{{`-$123,456.78`}},
			},
			{
				Statement: `SELECT '-92233720368547758.08'::money;`,
				Results:   []sql.Row{{`-$92,233,720,368,547,758.08`}},
			},
			{
				Statement: `SELECT '92233720368547758.07'::money;`,
				Results:   []sql.Row{{`$92,233,720,368,547,758.07`}},
			},
			{
				Statement:   `SELECT '-92233720368547758.09'::money;`,
				ErrorString: `value "-92233720368547758.09" is out of range for type money`,
			},
			{
				Statement:   `SELECT '92233720368547758.08'::money;`,
				ErrorString: `value "92233720368547758.08" is out of range for type money`,
			},
			{
				Statement:   `SELECT '-92233720368547758.085'::money;`,
				ErrorString: `value "-92233720368547758.085" is out of range for type money`,
			},
			{
				Statement:   `SELECT '92233720368547758.075'::money;`,
				ErrorString: `value "92233720368547758.075" is out of range for type money`,
			},
			{
				Statement: `SELECT '878.08'::money / 11::float8;`,
				Results:   []sql.Row{{`$79.83`}},
			},
			{
				Statement: `SELECT '878.08'::money / 11::float4;`,
				Results:   []sql.Row{{`$79.83`}},
			},
			{
				Statement: `SELECT '878.08'::money / 11::bigint;`,
				Results:   []sql.Row{{`$79.82`}},
			},
			{
				Statement: `SELECT '878.08'::money / 11::int;`,
				Results:   []sql.Row{{`$79.82`}},
			},
			{
				Statement: `SELECT '878.08'::money / 11::smallint;`,
				Results:   []sql.Row{{`$79.82`}},
			},
			{
				Statement: `SELECT '90000000000000099.00'::money / 10::bigint;`,
				Results:   []sql.Row{{`$9,000,000,000,000,009.90`}},
			},
			{
				Statement: `SELECT '90000000000000099.00'::money / 10::int;`,
				Results:   []sql.Row{{`$9,000,000,000,000,009.90`}},
			},
			{
				Statement: `SELECT '90000000000000099.00'::money / 10::smallint;`,
				Results:   []sql.Row{{`$9,000,000,000,000,009.90`}},
			},
			{
				Statement: `SELECT 1234567890::money;`,
				Results:   []sql.Row{{`$1,234,567,890.00`}},
			},
			{
				Statement: `SELECT 12345678901234567::money;`,
				Results:   []sql.Row{{`$12,345,678,901,234,567.00`}},
			},
			{
				Statement: `SELECT (-12345)::money;`,
				Results:   []sql.Row{{`-$12,345.00`}},
			},
			{
				Statement: `SELECT (-1234567890)::money;`,
				Results:   []sql.Row{{`-$1,234,567,890.00`}},
			},
			{
				Statement: `SELECT (-12345678901234567)::money;`,
				Results:   []sql.Row{{`-$12,345,678,901,234,567.00`}},
			},
			{
				Statement: `SELECT 1234567890::int4::money;`,
				Results:   []sql.Row{{`$1,234,567,890.00`}},
			},
			{
				Statement: `SELECT 12345678901234567::int8::money;`,
				Results:   []sql.Row{{`$12,345,678,901,234,567.00`}},
			},
			{
				Statement: `SELECT 12345678901234567::numeric::money;`,
				Results:   []sql.Row{{`$12,345,678,901,234,567.00`}},
			},
			{
				Statement: `SELECT (-1234567890)::int4::money;`,
				Results:   []sql.Row{{`-$1,234,567,890.00`}},
			},
			{
				Statement: `SELECT (-12345678901234567)::int8::money;`,
				Results:   []sql.Row{{`-$12,345,678,901,234,567.00`}},
			},
			{
				Statement: `SELECT (-12345678901234567)::numeric::money;`,
				Results:   []sql.Row{{`-$12,345,678,901,234,567.00`}},
			},
			{
				Statement: `SELECT '12345678901234567'::money::numeric;`,
				Results:   []sql.Row{{12345678901234567.00}},
			},
			{
				Statement: `SELECT '-12345678901234567'::money::numeric;`,
				Results:   []sql.Row{{-12345678901234567.00}},
			},
			{
				Statement: `SELECT '92233720368547758.07'::money::numeric;`,
				Results:   []sql.Row{{92233720368547758.07}},
			},
			{
				Statement: `SELECT '-92233720368547758.08'::money::numeric;`,
				Results:   []sql.Row{{-92233720368547758.08}},
			},
		},
	})
}
