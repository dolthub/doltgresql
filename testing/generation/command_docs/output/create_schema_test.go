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

func TestCreateSchema(t *testing.T) {
	tests := []QueryParses{
		Parses("CREATE SCHEMA schema_name"),
		Parses("CREATE SCHEMA schema_name AUTHORIZATION user_name"),
		Unimplemented("CREATE SCHEMA schema_name AUTHORIZATION CURRENT_ROLE"),
		Parses("CREATE SCHEMA schema_name AUTHORIZATION CURRENT_USER"),
		Parses("CREATE SCHEMA schema_name AUTHORIZATION SESSION_USER"),
		Unimplemented("CREATE SCHEMA schema_name CREATE TABLE tablename ( )"),
		Unimplemented("CREATE SCHEMA schema_name AUTHORIZATION user_name CREATE TABLE tablename ( )"),
		Unimplemented("CREATE SCHEMA schema_name AUTHORIZATION CURRENT_ROLE CREATE TABLE tablename ( )"),
		Unimplemented("CREATE SCHEMA schema_name AUTHORIZATION CURRENT_USER CREATE TABLE tablename ( )"),
		Unimplemented("CREATE SCHEMA schema_name AUTHORIZATION SESSION_USER CREATE TABLE tablename ( )"),
		Unimplemented("CREATE SCHEMA schema_name CREATE TABLE tablename ( ) CREATE TABLE tablename ( )"),
		Unimplemented("CREATE SCHEMA schema_name AUTHORIZATION user_name CREATE TABLE tablename ( ) CREATE TABLE tablename ( )"),
		Unimplemented("CREATE SCHEMA schema_name AUTHORIZATION CURRENT_ROLE CREATE TABLE tablename ( ) CREATE TABLE tablename ( )"),
		Unimplemented("CREATE SCHEMA schema_name AUTHORIZATION CURRENT_USER CREATE TABLE tablename ( ) CREATE TABLE tablename ( )"),
		Unimplemented("CREATE SCHEMA schema_name AUTHORIZATION SESSION_USER CREATE TABLE tablename ( ) CREATE TABLE tablename ( )"),
		Parses("CREATE SCHEMA AUTHORIZATION user_name"),
		Unimplemented("CREATE SCHEMA AUTHORIZATION CURRENT_ROLE"),
		Parses("CREATE SCHEMA AUTHORIZATION CURRENT_USER"),
		Parses("CREATE SCHEMA AUTHORIZATION SESSION_USER"),
		Unimplemented("CREATE SCHEMA AUTHORIZATION user_name CREATE TABLE tablename ( )"),
		Unimplemented("CREATE SCHEMA AUTHORIZATION CURRENT_ROLE CREATE TABLE tablename ( )"),
		Unimplemented("CREATE SCHEMA AUTHORIZATION CURRENT_USER CREATE TABLE tablename ( )"),
		Unimplemented("CREATE SCHEMA AUTHORIZATION SESSION_USER CREATE TABLE tablename ( )"),
		Unimplemented("CREATE SCHEMA AUTHORIZATION user_name CREATE TABLE tablename ( ) CREATE TABLE tablename ( )"),
		Unimplemented("CREATE SCHEMA AUTHORIZATION CURRENT_ROLE CREATE TABLE tablename ( ) CREATE TABLE tablename ( )"),
		Unimplemented("CREATE SCHEMA AUTHORIZATION CURRENT_USER CREATE TABLE tablename ( ) CREATE TABLE tablename ( )"),
		Unimplemented("CREATE SCHEMA AUTHORIZATION SESSION_USER CREATE TABLE tablename ( ) CREATE TABLE tablename ( )"),
		Parses("CREATE SCHEMA IF NOT EXISTS schema_name"),
		Parses("CREATE SCHEMA IF NOT EXISTS schema_name AUTHORIZATION user_name"),
		Unimplemented("CREATE SCHEMA IF NOT EXISTS schema_name AUTHORIZATION CURRENT_ROLE"),
		Parses("CREATE SCHEMA IF NOT EXISTS schema_name AUTHORIZATION CURRENT_USER"),
		Parses("CREATE SCHEMA IF NOT EXISTS schema_name AUTHORIZATION SESSION_USER"),
		Parses("CREATE SCHEMA IF NOT EXISTS AUTHORIZATION user_name"),
		Unimplemented("CREATE SCHEMA IF NOT EXISTS AUTHORIZATION CURRENT_ROLE"),
		Parses("CREATE SCHEMA IF NOT EXISTS AUTHORIZATION CURRENT_USER"),
		Parses("CREATE SCHEMA IF NOT EXISTS AUTHORIZATION SESSION_USER"),
	}
	RunTests(t, tests)
}
