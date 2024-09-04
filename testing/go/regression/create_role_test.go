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
)

func TestCreateRole(t *testing.T) {
	t.Skip()
	_ = RunTests(t, RegressionFileName_create_role)
}

func init() {
	RegisterRegressionFile(RegressionFile{
		RegressionFileName: RegressionFileName_create_role,
		DependsOn:          []RegressionFileName{RegressionFileName_test_setup},
		Statements: []RegressionFileStatement{
			{
				Statement: `CREATE ROLE regress_role_super SUPERUSER;`,
			},
			{
				Statement: `CREATE ROLE regress_role_admin CREATEDB CREATEROLE REPLICATION BYPASSRLS;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_role_admin;`,
			},
			{
				Statement:   `CREATE ROLE regress_nosuch_superuser SUPERUSER;`,
				ErrorString: `must be superuser to create superusers`,
			},
			{
				Statement:   `CREATE ROLE regress_nosuch_replication_bypassrls REPLICATION BYPASSRLS;`,
				ErrorString: `must be superuser to create replication users`,
			},
			{
				Statement:   `CREATE ROLE regress_nosuch_replication REPLICATION;`,
				ErrorString: `must be superuser to create replication users`,
			},
			{
				Statement:   `CREATE ROLE regress_nosuch_bypassrls BYPASSRLS;`,
				ErrorString: `must be superuser to create bypassrls users`,
			},
			{
				Statement: `CREATE ROLE regress_createdb CREATEDB;`,
			},
			{
				Statement: `CREATE ROLE regress_createrole CREATEROLE;`,
			},
			{
				Statement: `CREATE ROLE regress_login LOGIN;`,
			},
			{
				Statement: `CREATE ROLE regress_inherit INHERIT;`,
			},
			{
				Statement: `CREATE ROLE regress_connection_limit CONNECTION LIMIT 5;`,
			},
			{
				Statement: `CREATE ROLE regress_encrypted_password ENCRYPTED PASSWORD 'foo';`,
			},
			{
				Statement: `CREATE ROLE regress_password_null PASSWORD NULL;`,
			},
			{
				Statement: `CREATE ROLE regress_noiseword SYSID 12345;`,
			},
			{
				Statement:   `CREATE ROLE regress_nosuch_super IN ROLE regress_role_super;`,
				ErrorString: `must be superuser to alter superusers`,
			},
			{
				Statement:   `CREATE ROLE regress_nosuch_dbowner IN ROLE pg_database_owner;`,
				ErrorString: `role "pg_database_owner" cannot have explicit members`,
			},
			{
				Statement: `CREATE ROLE regress_inroles ROLE
	regress_role_super, regress_createdb, regress_createrole, regress_login,
	regress_inherit, regress_connection_limit, regress_encrypted_password, regress_password_null;`,
			},
			{
				Statement:   `CREATE ROLE regress_nosuch_recursive ROLE regress_nosuch_recursive;`,
				ErrorString: `role "regress_nosuch_recursive" is a member of role "regress_nosuch_recursive"`,
			},
			{
				Statement: `CREATE ROLE regress_adminroles ADMIN
	regress_role_super, regress_createdb, regress_createrole, regress_login,
	regress_inherit, regress_connection_limit, regress_encrypted_password, regress_password_null;`,
			},
			{
				Statement:   `CREATE ROLE regress_nosuch_admin_recursive ADMIN regress_nosuch_admin_recursive;`,
				ErrorString: `role "regress_nosuch_admin_recursive" is a member of role "regress_nosuch_admin_recursive"`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_createrole;`,
			},
			{
				Statement:   `CREATE DATABASE regress_nosuch_db;`,
				ErrorString: `permission denied to create database`,
			},
			{
				Statement: `CREATE ROLE regress_plainrole;`,
			},
			{
				Statement: `CREATE ROLE regress_rolecreator CREATEROLE;`,
			},
			{
				Statement: `CREATE ROLE regress_tenant CREATEDB CREATEROLE LOGIN INHERIT CONNECTION LIMIT 5;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_tenant;`,
			},
			{
				Statement: `CREATE TABLE tenant_table (i integer);`,
			},
			{
				Statement: `CREATE INDEX tenant_idx ON tenant_table(i);`,
			},
			{
				Statement: `CREATE VIEW tenant_view AS SELECT * FROM pg_catalog.pg_class;`,
			},
			{
				Statement: `REVOKE ALL PRIVILEGES ON tenant_table FROM PUBLIC;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_createrole;`,
			},
			{
				Statement:   `DROP INDEX tenant_idx;`,
				ErrorString: `must be owner of index tenant_idx`,
			},
			{
				Statement:   `ALTER TABLE tenant_table ADD COLUMN t text;`,
				ErrorString: `must be owner of table tenant_table`,
			},
			{
				Statement:   `DROP TABLE tenant_table;`,
				ErrorString: `must be owner of table tenant_table`,
			},
			{
				Statement:   `ALTER VIEW tenant_view OWNER TO regress_role_admin;`,
				ErrorString: `must be owner of view tenant_view`,
			},
			{
				Statement:   `DROP VIEW tenant_view;`,
				ErrorString: `must be owner of view tenant_view`,
			},
			{
				Statement:   `REASSIGN OWNED BY regress_tenant TO regress_createrole;`,
				ErrorString: `permission denied to reassign objects`,
			},
			{
				Statement: `CREATE ROLE regress_read_all_data IN ROLE pg_read_all_data;`,
			},
			{
				Statement: `CREATE ROLE regress_write_all_data IN ROLE pg_write_all_data;`,
			},
			{
				Statement: `CREATE ROLE regress_monitor IN ROLE pg_monitor;`,
			},
			{
				Statement: `CREATE ROLE regress_read_all_settings IN ROLE pg_read_all_settings;`,
			},
			{
				Statement: `CREATE ROLE regress_read_all_stats IN ROLE pg_read_all_stats;`,
			},
			{
				Statement: `CREATE ROLE regress_stat_scan_tables IN ROLE pg_stat_scan_tables;`,
			},
			{
				Statement: `CREATE ROLE regress_read_server_files IN ROLE pg_read_server_files;`,
			},
			{
				Statement: `CREATE ROLE regress_write_server_files IN ROLE pg_write_server_files;`,
			},
			{
				Statement: `CREATE ROLE regress_execute_server_program IN ROLE pg_execute_server_program;`,
			},
			{
				Statement: `CREATE ROLE regress_signal_backend IN ROLE pg_signal_backend;`,
			},
			{
				Statement: `SET SESSION AUTHORIZATION regress_role_admin;`,
			},
			{
				Statement:   `DROP ROLE regress_nosuch_superuser;`,
				ErrorString: `role "regress_nosuch_superuser" does not exist`,
			},
			{
				Statement:   `DROP ROLE regress_nosuch_replication_bypassrls;`,
				ErrorString: `role "regress_nosuch_replication_bypassrls" does not exist`,
			},
			{
				Statement:   `DROP ROLE regress_nosuch_replication;`,
				ErrorString: `role "regress_nosuch_replication" does not exist`,
			},
			{
				Statement:   `DROP ROLE regress_nosuch_bypassrls;`,
				ErrorString: `role "regress_nosuch_bypassrls" does not exist`,
			},
			{
				Statement:   `DROP ROLE regress_nosuch_super;`,
				ErrorString: `role "regress_nosuch_super" does not exist`,
			},
			{
				Statement:   `DROP ROLE regress_nosuch_dbowner;`,
				ErrorString: `role "regress_nosuch_dbowner" does not exist`,
			},
			{
				Statement:   `DROP ROLE regress_nosuch_recursive;`,
				ErrorString: `role "regress_nosuch_recursive" does not exist`,
			},
			{
				Statement:   `DROP ROLE regress_nosuch_admin_recursive;`,
				ErrorString: `role "regress_nosuch_admin_recursive" does not exist`,
			},
			{
				Statement: `DROP ROLE regress_plainrole;`,
			},
			{
				Statement: `DROP ROLE regress_createdb;`,
			},
			{
				Statement: `DROP ROLE regress_createrole;`,
			},
			{
				Statement: `DROP ROLE regress_login;`,
			},
			{
				Statement: `DROP ROLE regress_inherit;`,
			},
			{
				Statement: `DROP ROLE regress_connection_limit;`,
			},
			{
				Statement: `DROP ROLE regress_encrypted_password;`,
			},
			{
				Statement: `DROP ROLE regress_password_null;`,
			},
			{
				Statement: `DROP ROLE regress_noiseword;`,
			},
			{
				Statement: `DROP ROLE regress_inroles;`,
			},
			{
				Statement: `DROP ROLE regress_adminroles;`,
			},
			{
				Statement: `DROP ROLE regress_rolecreator;`,
			},
			{
				Statement: `DROP ROLE regress_read_all_data;`,
			},
			{
				Statement: `DROP ROLE regress_write_all_data;`,
			},
			{
				Statement: `DROP ROLE regress_monitor;`,
			},
			{
				Statement: `DROP ROLE regress_read_all_settings;`,
			},
			{
				Statement: `DROP ROLE regress_read_all_stats;`,
			},
			{
				Statement: `DROP ROLE regress_stat_scan_tables;`,
			},
			{
				Statement: `DROP ROLE regress_read_server_files;`,
			},
			{
				Statement: `DROP ROLE regress_write_server_files;`,
			},
			{
				Statement: `DROP ROLE regress_execute_server_program;`,
			},
			{
				Statement: `DROP ROLE regress_signal_backend;`,
			},
			{
				Statement:   `DROP ROLE regress_tenant;`,
				ErrorString: `role "regress_tenant" cannot be dropped because some objects depend on it`,
			},
			{
				Statement: `owner of view tenant_view
DROP ROLE regress_role_super;`,
				ErrorString: `must be superuser to drop superusers`,
			},
			{
				Statement:   `DROP ROLE regress_role_admin;`,
				ErrorString: `current user cannot be dropped`,
			},
			{
				Statement: `RESET SESSION AUTHORIZATION;`,
			},
			{
				Statement: `DROP INDEX tenant_idx;`,
			},
			{
				Statement: `DROP TABLE tenant_table;`,
			},
			{
				Statement: `DROP VIEW tenant_view;`,
			},
			{
				Statement: `DROP ROLE regress_tenant;`,
			},
			{
				Statement: `DROP ROLE regress_role_admin;`,
			},
			{
				Statement: `DROP ROLE regress_role_super;`,
			},
		},
	})
}
