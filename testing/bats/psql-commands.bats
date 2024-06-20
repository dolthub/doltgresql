#!/usr/bin/env bats
load $BATS_TEST_DIRNAME/setup/common.bash

setup() {
    setup_common
    start_sql_server
    query_server <<SQL
    CREATE TABLE test1 (pk BIGINT PRIMARY KEY, v1 SMALLINT);
    CREATE TABLE test2 (pk BIGINT PRIMARY KEY, v1 INTEGER, v2 SMALLINT);
    INSERT INTO test1 VALUES (1, 2), (6, 7);
    INSERT INTO test2 VALUES (3, 4, 5), (8, 9, 0);
    CREATE VIEW testview AS SELECT * FROM test1;
SQL
}

teardown() {
    teardown_common
}

@test 'psql-commands: \l' {
    run query_server -c "\l"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "information_schema" ]] || false
    [[ "$output" =~ "postgres" ]] || false
    [[ "$output" =~ "doltgres" ]] || false
}

# TODO: These should not include pg_catalog tables
@test 'psql-commands: \dt' {
    run query_server --csv -c "\dt"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "public,pg_aggregate,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_am,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_amop,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_amproc,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_attrdef,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_attribute,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_auth_members,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_authid,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_cast,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_class,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_collation,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_constraint,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_conversion,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_database,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_db_role_setting,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_default_acl,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_depend,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_description,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_enum,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_event_trigger,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_extension,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_foreign_data_wrapper,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_foreign_server,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_foreign_table,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_index,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_inherits,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_init_privs,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_language,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_largeobject,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_largeobject_metadata,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_namespace,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_opclass,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_opfamily,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_parameter_acl,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_partitioned_table,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_proc,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_sequence,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_trigger,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_type,table,postgres" ]] || false
    [[ "$output" =~ "public,test1,table,postgres" ]] || false
    [[ "$output" =~ "public,test2,table,postgres" ]] || false
    [ "${#lines[@]}" -eq 42 ]
}

@test 'psql-commands: \d' {
    run query_server --csv -c "\d"
    echo "$output"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "public,pg_aggregate,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_am,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_amop,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_amproc,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_attrdef,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_attribute,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_auth_members,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_authid,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_cast,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_class,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_collation,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_constraint,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_conversion,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_database,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_db_role_setting,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_default_acl,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_depend,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_description,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_enum,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_event_trigger,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_extension,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_foreign_data_wrapper,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_foreign_server,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_foreign_table,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_index,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_inherits,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_init_privs,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_language,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_largeobject,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_largeobject_metadata,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_namespace,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_opclass,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_opfamily,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_parameter_acl,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_partitioned_table,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_proc,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_sequence,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_trigger,table,postgres" ]] || false
    [[ "$output" =~ "public,pg_type,table,postgres" ]] || false
    [[ "$output" =~ "public,test1,table,postgres" ]] || false
    [[ "$output" =~ "public,test2,table,postgres" ]] || false
    [ "${#lines[@]}" -eq 42 ]
}

@test 'psql-commands: \d table' {
    skip "this command has not yet been implemented"
}

@test 'psql-commands: \dn' {
    run query_server --csv -c "\dn"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "public,pg_database_owner" ]] || false
    [ "${#lines[@]}" -eq 2 ]
}

@test 'psql-commands: \df' {
    run query_server -c "\df"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "0 rows" ]] || false
}

@test 'psql-commands: \dv' {
    skip "need to reimplement CREATE VIEW support"
    run query_server --csv -c "\dv"
    [ "$status" -eq 0 ]
    [[ "$output" =~ "public,testview,view,postgres" ]] || false
    [ "${#lines[@]}" -eq 2 ]
}

@test 'psql-commands: \du' {
    skip "users have not yet been implemented"
}
