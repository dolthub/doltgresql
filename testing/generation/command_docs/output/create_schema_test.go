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
		Converts("CREATE SCHEMA schema_name"),
		Converts("CREATE SCHEMA schema_name AUTHORIZATION user_name"),
		Converts("CREATE SCHEMA schema_name AUTHORIZATION CURRENT_ROLE"),
		Converts("CREATE SCHEMA schema_name AUTHORIZATION CURRENT_USER"),
		Converts("CREATE SCHEMA schema_name AUTHORIZATION SESSION_USER"),
		Converts("CREATE SCHEMA schema_name CREATE TABLE tablename ( )"),
		Converts("CREATE SCHEMA schema_name AUTHORIZATION user_name CREATE TABLE tablename ( )"),
		Converts("CREATE SCHEMA schema_name AUTHORIZATION CURRENT_ROLE CREATE TABLE tablename ( )"),
		Converts("CREATE SCHEMA schema_name AUTHORIZATION CURRENT_USER CREATE TABLE tablename ( )"),
		Converts("CREATE SCHEMA schema_name AUTHORIZATION SESSION_USER CREATE TABLE tablename ( )"),
		Converts("CREATE SCHEMA schema_name CREATE TABLE tablename ( ) CREATE TABLE tablename ( )"),
		Converts("CREATE SCHEMA schema_name AUTHORIZATION user_name CREATE TABLE tablename ( ) CREATE TABLE tablename ( )"),
		Converts("CREATE SCHEMA schema_name AUTHORIZATION CURRENT_ROLE CREATE TABLE tablename ( ) CREATE TABLE tablename ( )"),
		Converts("CREATE SCHEMA schema_name AUTHORIZATION CURRENT_USER CREATE TABLE tablename ( ) CREATE TABLE tablename ( )"),
		Converts("CREATE SCHEMA schema_name AUTHORIZATION SESSION_USER CREATE TABLE tablename ( ) CREATE TABLE tablename ( )"),
		Converts("CREATE SCHEMA AUTHORIZATION user_name"),
		Converts("CREATE SCHEMA AUTHORIZATION CURRENT_ROLE"),
		Converts("CREATE SCHEMA AUTHORIZATION CURRENT_USER"),
		Converts("CREATE SCHEMA AUTHORIZATION SESSION_USER"),
		Converts("CREATE SCHEMA AUTHORIZATION user_name CREATE TABLE tablename ( )"),
		Converts("CREATE SCHEMA AUTHORIZATION CURRENT_ROLE CREATE TABLE tablename ( )"),
		Converts("CREATE SCHEMA AUTHORIZATION CURRENT_USER CREATE TABLE tablename ( )"),
		Converts("CREATE SCHEMA AUTHORIZATION SESSION_USER CREATE TABLE tablename ( )"),
		Converts("CREATE SCHEMA AUTHORIZATION user_name CREATE TABLE tablename ( ) CREATE TABLE tablename ( )"),
		Converts("CREATE SCHEMA AUTHORIZATION CURRENT_ROLE CREATE TABLE tablename ( ) CREATE TABLE tablename ( )"),
		Converts("CREATE SCHEMA AUTHORIZATION CURRENT_USER CREATE TABLE tablename ( ) CREATE TABLE tablename ( )"),
		Converts("CREATE SCHEMA AUTHORIZATION SESSION_USER CREATE TABLE tablename ( ) CREATE TABLE tablename ( )"),
		Converts("CREATE SCHEMA IF NOT EXISTS schema_name"),
		Converts("CREATE SCHEMA IF NOT EXISTS schema_name AUTHORIZATION user_name"),
		Converts("CREATE SCHEMA IF NOT EXISTS schema_name AUTHORIZATION CURRENT_ROLE"),
		Converts("CREATE SCHEMA IF NOT EXISTS schema_name AUTHORIZATION CURRENT_USER"),
		Converts("CREATE SCHEMA IF NOT EXISTS schema_name AUTHORIZATION SESSION_USER"),
		Converts("CREATE SCHEMA IF NOT EXISTS AUTHORIZATION user_name"),
		Converts("CREATE SCHEMA IF NOT EXISTS AUTHORIZATION CURRENT_ROLE"),
		Converts("CREATE SCHEMA IF NOT EXISTS AUTHORIZATION CURRENT_USER"),
		Converts("CREATE SCHEMA IF NOT EXISTS AUTHORIZATION SESSION_USER"),
	}
	RunTests(t, tests)
}
