// Copyright 2023 Dolthub, Inc.
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

package output

import "testing"

func TestCreateStatistics(t *testing.T) {
	tests := []QueryParses{
		Unimplemented("CREATE STATISTICS statistics_name ON ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ON ( expression ) FROM table_name"),
		Parses("CREATE STATISTICS statistics_name ON column_name , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ON column_name , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind ) ON column_name , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind ) ON column_name , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind , statistics_kind ) ON column_name , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind , statistics_kind ) ON column_name , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ON ( expression ) , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ON ( expression ) , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind ) ON ( expression ) , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind ) ON ( expression ) , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind , statistics_kind ) ON ( expression ) , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind , statistics_kind ) ON ( expression ) , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ON column_name , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ON column_name , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind ) ON column_name , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind ) ON column_name , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind , statistics_kind ) ON column_name , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind , statistics_kind ) ON column_name , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ON ( expression ) , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ON ( expression ) , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind ) ON ( expression ) , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind ) ON ( expression ) , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind , statistics_kind ) ON ( expression ) , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind , statistics_kind ) ON ( expression ) , ( expression ) FROM table_name"),
		Parses("CREATE STATISTICS statistics_name ON column_name , column_name , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ON column_name , column_name , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind ) ON column_name , column_name , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind ) ON column_name , column_name , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind , statistics_kind ) ON column_name , column_name , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind , statistics_kind ) ON column_name , column_name , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ON ( expression ) , column_name , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ON ( expression ) , column_name , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind ) ON ( expression ) , column_name , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind ) ON ( expression ) , column_name , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind , statistics_kind ) ON ( expression ) , column_name , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind , statistics_kind ) ON ( expression ) , column_name , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ON column_name , ( expression ) , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ON column_name , ( expression ) , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind ) ON column_name , ( expression ) , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind ) ON column_name , ( expression ) , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind , statistics_kind ) ON column_name , ( expression ) , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind , statistics_kind ) ON column_name , ( expression ) , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ON ( expression ) , ( expression ) , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ON ( expression ) , ( expression ) , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind ) ON ( expression ) , ( expression ) , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind ) ON ( expression ) , ( expression ) , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind , statistics_kind ) ON ( expression ) , ( expression ) , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind , statistics_kind ) ON ( expression ) , ( expression ) , column_name FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ON column_name , column_name , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ON column_name , column_name , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind ) ON column_name , column_name , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind ) ON column_name , column_name , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind , statistics_kind ) ON column_name , column_name , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind , statistics_kind ) ON column_name , column_name , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ON ( expression ) , column_name , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ON ( expression ) , column_name , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind ) ON ( expression ) , column_name , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind ) ON ( expression ) , column_name , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind , statistics_kind ) ON ( expression ) , column_name , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind , statistics_kind ) ON ( expression ) , column_name , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ON column_name , ( expression ) , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ON column_name , ( expression ) , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind ) ON column_name , ( expression ) , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind ) ON column_name , ( expression ) , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind , statistics_kind ) ON column_name , ( expression ) , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind , statistics_kind ) ON column_name , ( expression ) , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ON ( expression ) , ( expression ) , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ON ( expression ) , ( expression ) , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind ) ON ( expression ) , ( expression ) , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind ) ON ( expression ) , ( expression ) , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS statistics_name ( statistics_kind , statistics_kind ) ON ( expression ) , ( expression ) , ( expression ) FROM table_name"),
		Unimplemented("CREATE STATISTICS IF NOT EXISTS statistics_name ( statistics_kind , statistics_kind ) ON ( expression ) , ( expression ) , ( expression ) FROM table_name"),
	}
	RunTests(t, tests)
}
