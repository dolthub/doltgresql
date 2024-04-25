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

package pg_catalog

import (
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/types"
)

// Below are BASE TABLEs
const (
	// PgAggregateTableName is the name of the pg_aggregate table.
	PgAggregateTableName = "pg_aggregate"
	// PgAmTableName is the name of the pg_am table.
	PgAmTableName = "pg_am"
	// PgAmOpTableName is the name of the pg_amop table.
	PgAmOpTableName = "pg_amop"
	// PgAmProcTableName is the name of the pg_amproc table.
	PgAmProcTableName = "pg_amproc"
	// PgAttrDefTableName is the name of the pg_attrdef table.
	PgAttrDefTableName = "pg_attrdef"
	// PgAttributeTableName is the name of the pg_attribute table.
	PgAttributeTableName = "pg_attribute"
	// PgAuthMembersTableName is the name of the pg_auth_members table.
	PgAuthMembersTableName = "pg_auth_members"
	// PgAuthIdTableName is the name of the pg_authid table.
	PgAuthIdTableName = "pg_authid"
	// PgCastTableName is the name of the pg_cast table.
	PgCastTableName = "pg_cast"
	// PgClassTableName is the name of the pg_class table.
	PgClassTableName = "pg_class"
	// PgCollationTableName is the name of the pg_collation table.
	PgCollationTableName = "pg_collation"
	// PgConstraintTableName is the name of the pg_constraint table.
	PgConstraintTableName = "pg_constraint"
	// PgConversionTableName is the name of the pg_conversion table.
	PgConversionTableName = "pg_conversion"
	// PgDatabaseTableName is the name of the pg_database table.
	PgDatabaseTableName = "pg_database"
	// PgDbRoleSettingTableName is the name of the pg_db_role_setting table.
	PgDbRoleSettingTableName = "pg_db_role_setting"
	// PgDefaultAclTableName is the name of the pg_default_acl table.
	PgDefaultAclTableName = "pg_default_acl"
	// PgDependTableName is the name of the pg_depend table.
	PgDependTableName = "pg_depend"
	// PgDescriptionTableName is the name of the pg_description table.
	PgDescriptionTableName = "pg_description"
	// PgEnumTableName is the name of the pg_enum table.
	PgEnumTableName = "pg_enum"
	// PgEventTriggerTableName is the name of the pg_event_trigger table.
	PgEventTriggerTableName = "pg_event_trigger"
	// PgExtensionTableName is the name of the pg_extension table.
	PgExtensionTableName = "pg_extension"
	// PgForeignDataWrapperTableName is the name of the pg_foreign_data_wrapper table.
	PgForeignDataWrapperTableName = "pg_foreign_data_wrapper"
	// PgForeignServerTableName is the name of the pg_foreign_server table.
	PgForeignServerTableName = "pg_foreign_server"
	// PgForeignTableTableName is the name of the pg_foreign_table table.
	PgForeignTableTableName = "pg_foreign_table"
	// PgIndexTableName is the name of the pg_index table.
	PgIndexTableName = "pg_index"
	// PgInheritsTableName is the name of the pg_inherits table.
	PgInheritsTableName = "pg_inherits"
	// PgInitPrivsTableName is the name of the pg_init_privs table.
	PgInitPrivsTableName = "pg_init_privs"
	// PgLanguageTableName is the name of the pg_language table.
	PgLanguageTableName = "pg_language"
	// PgLargeObjectTableName is the name of the pg_largeobject table.
	PgLargeObjectTableName = "pg_largeobject"
	// PgLargeObjectMetadataTableName is the name of the pg_largeobject_metadata table.
	PgLargeObjectMetadataTableName = "pg_largeobject_metadata"
	// PgNamespaceTableName is the name of the pg_namespace table.
	PgNamespaceTableName = "pg_namespace"
	// PgOpClassTableName is the name of the pg_opclass table.
	PgOpClassTableName = "pg_opclass"
	// PgOperatorTableName is the name of the pg_operator table.
	PgOperatorTableName = "pg_operator"
	// PgOpFamilyTableName is the name of the pg_opfamily table.
	PgOpFamilyTableName = "pg_opfamily"
	// PgParameterAclTableName is the name of the pg_parameter_acl table.
	PgParameterAclTableName = "pg_parameter_acl"
	// PgPartitionedTableTableName is the name of the pg_partitioned_table table.
	PgPartitionedTableTableName = "pg_partitioned_table"
	// PgPolicyTableName is the name of the pg_policy table.
	PgPolicyTableName = "pg_policy"
	// PgProcTableName is the name of the pg_proc table.
	PgProcTableName = "pg_proc"
	// PgPublicationTableName is the name of the pg_publication table.
	PgPublicationTableName = "pg_publication"
	// PgPublicationNamespaceTableName is the name of the pg_publication_namespace table.
	PgPublicationNamespaceTableName = "pg_publication_namespace"
	// PgPublicationRelTableName is the name of the pg_publication_rel table.
	PgPublicationRelTableName = "pg_publication_rel"
	// PgRangeTableName is the name of the pg_range table.
	PgRangeTableName = "pg_range"
	// PgReplicationOriginTableName is the name of the pg_replication_origin table.
	PgReplicationOriginTableName = "pg_replication_origin"
	// PgRewriteTableName is the name of the pg_rewrite table.
	PgRewriteTableName = "pg_rewrite"
	// PgSecLabelTableName is the name of the pg_seclabel table.
	PgSecLabelTableName = "pg_seclabel"
	// PgSequenceTableName is the name of the pg_sequence table.
	PgSequenceTableName = "pg_sequence"
	// PgShDependTableName is the name of the pg_shdepend table.
	PgShDependTableName = "pg_shdepend"
	// PgShDescriptionTableName is the name of the pg_shdescription table.
	PgShDescriptionTableName = "pg_shdescription"
	// PgShSecLabelTableName is the name of the pg_shseclabel table.
	PgShSecLabelTableName = "pg_shseclabel"
	// PgStatisticsTableName is the name of the pg_statistic table.
	PgStatisticsTableName = "pg_statistic"
	// PgStatisticsExtTableName is the name of the pg_statistic_ext table.
	PgStatisticsExtTableName = "pg_statistic_ext"
	// PgStatisticsExtDataTableName is the name of the pg_statistic_ext_data table.
	PgStatisticsExtDataTableName = "pg_statistic_ext_data"
	// PgSubscriptionTableName is the name of the pg_subscription table.
	PgSubscriptionTableName = "pg_subscription"
	// PgSubscriptionRelTableName is the name of the pg_subscription_rel table.
	PgSubscriptionRelTableName = "pg_subscription_rel"
	// PgTablespaceTableName is the name of the pg_tablespace table.
	PgTablespaceTableName = "pg_tablespace"
	// PgTransformTableName is the name of the pg_transform table.
	PgTransformTableName = "pg_transform"
	// PgTriggerTableName is the name of the pg_trigger table.
	PgTriggerTableName = "pg_trigger"
	// PgTsConfigTableName is the name of the pg_ts_config table.
	PgTsConfigTableName = "pg_ts_config"
	// PgTsConfigMapTableName is the name of the pg_ts_config_map table.
	PgTsConfigMapTableName = "pg_ts_config_map"
	// PgTsDictTableName is the name of the pg_ts_dict table.
	PgTsDictTableName = "pg_ts_dict"
	// PgTsParserTableName is the name of the pg_ts_parser table.
	PgTsParserTableName = "pg_ts_parser"
	// PgTsTemplateTableName is the name of the pg_ts_template table.
	PgTsTemplateTableName = "pg_ts_template"
	// PgTypeTableName is the name of the pg_type table.
	PgTypeTableName = "pg_type"
	// PgUserMappingTableName is the name of the pg_user_mapping table.
	PgUserMappingTableName = "pg_user_mapping"
)

// Below are VIEWs
const (
	// PgAvailableExtensionVersionsTableName is the name of the pg_available_extension_versions table.
	PgAvailableExtensionVersionsTableName = "pg_available_extension_versions"
	// PgAvailableExtensionsTableName is the name of the pg_available_extensions table.
	PgAvailableExtensionsTableName = "pg_available_extensions"
	// PgBackendMemoryContextsTableName is the name of the pg_backend_memory_contexts table.
	PgBackendMemoryContextsTableName = "pg_backend_memory_contexts"
	// PgConfigTableName is the name of the pg_config table.
	PgConfigTableName = "pg_config"
	// PgCursorsTableName is the name of the pg_cursors table.
	PgCursorsTableName = "pg_cursors"
	// PgFileSettingsTableName is the name of the pg_file_settings table.
	PgFileSettingsTableName = "pg_file_settings"
	// PgGroupTableName is the name of the pg_group table.
	PgGroupTableName = "pg_group"
	// PgHbaFileRulesTableName is the name of the pg_hba_file_rules table.
	PgHbaFileRulesTableName = "pg_hba_file_rules"
	// PgIdentFileMappingsTableName is the name of the pg_ident_file_mappings table.
	PgIdentFileMappingsTableName = "pg_ident_file_mappings"
	// PgIndexesTableName is the name of the pg_indexes table.
	PgIndexesTableName = "pg_indexes"
	// PgLocksTableName is the name of the pg_locks table.
	PgLocksTableName = "pg_locks"
	// PgMatViewsTableName is the name of the pg_matviews table.
	PgMatViewsTableName = "pg_matviews"
	// PgPoliciesTableName is the name of the pg_policies table.
	PgPoliciesTableName = "pg_policies"
	// PgPreparedStatementsTableName is the name of the pg_prepared_statements table.
	PgPreparedStatementsTableName = "pg_prepared_statements"
	// PgPreparedXactsTableName is the name of the pg_prepared_xacts table.
	PgPreparedXactsTableName = "pg_prepared_xacts"
	// PgPublicationTablesTableName is the name of the pg_publication_tables table.
	PgPublicationTablesTableName = "pg_publication_tables"
	// PgReplicationOriginStatusTableName is the name of the pg_replication_origin_status table.
	PgReplicationOriginStatusTableName = "pg_replication_origin_status"
	// PgReplicationSlotsTableName is the name of the pg_replication_slots table.
	PgReplicationSlotsTableName = "pg_replication_slots"
	// PgRolesTableName is the name of the pg_roles table.
	PgRolesTableName = "pg_roles"
	// PgRulesTableName is the name of the pg_rules table.
	PgRulesTableName = "pg_rules"
	// PgSecLabelsTableName is the name of the pg_seclabels table.
	PgSecLabelsTableName = "pg_seclabels"
	// PgSequencesTableName is the name of the pg_sequences table.
	PgSequencesTableName = "pg_sequences"
	// PgSettingsTableName is the name of the pg_settings table.
	PgSettingsTableName = "pg_settings"
	// PgShadowTableName is the name of the pg_shadow table.
	PgShadowTableName = "pg_shadow"
	// PgShMemAllocationsTableName is the name of the pg_shmem_allocations table.
	PgShMemAllocationsTableName = "pg_shmem_allocations"
	// PgStatActivityTableName is the name of the pg_stat_activity table.
	PgStatActivityTableName = "pg_stat_activity"
	// PgStatAllIndexesTableName is the name of the pg_stat_all_indexes table.
	PgStatAllIndexesTableName = "pg_stat_all_indexes"
	// PgStatAllTablesTableName is the name of the pg_stat_all_tables table.
	PgStatAllTablesTableName = "pg_stat_all_tables"
	// PgStatArchiverTableName is the name of the pg_stat_archiver table.
	PgStatArchiverTableName = "pg_stat_archiver"
	// PgStatBgWriterTableName is the name of the pg_stat_bgwriter table.
	PgStatBgWriterTableName = "pg_stat_bgwriter"
	// PgStatDatabaseTableName is the name of the pg_stat_database table.
	PgStatDatabaseTableName = "pg_stat_database"
	// PgStatDatabaseConflictsTableName is the name of the pg_stat_database_conflicts table.
	PgStatDatabaseConflictsTableName = "pg_stat_database_conflicts"
	// PgStatGssapiTableName is the name of the pg_stat_gssapi table.
	PgStatGssapiTableName = "pg_stat_gssapi"
	// PgStatIoTableName is the name of the pg_stat_io table.
	PgStatIoTableName = "pg_stat_io"
	// PgStatProgressAnalyzeTableName is the name of the pg_stat_progress_analyze table.
	PgStatProgressAnalyzeTableName = "pg_stat_progress_analyze"
	// PgStatProgressBaseBackupTableName is the name of the pg_stat_progress_basebackup table.
	PgStatProgressBaseBackupTableName = "pg_stat_progress_basebackup"
	// PgStatProgressClusterTableName is the name of the pg_stat_progress_cluster table.
	PgStatProgressClusterTableName = "pg_stat_progress_cluster"
	// PgStatProgressCopyTableName is the name of the pg_stat_progress_copy table.
	PgStatProgressCopyTableName = "pg_stat_progress_copy"
	// PgStatProgressCreateIndexTableName is the name of the pg_stat_progress_create_index table.
	PgStatProgressCreateIndexTableName = "pg_stat_progress_create_index"
	// PgStatProgressVacuumTableName is the name of the pg_stat_progress_vacuum table.
	PgStatProgressVacuumTableName = "pg_stat_progress_vacuum"
	// PgStatRecoveryPrefetchTableName is the name of the pg_stat_recovery_prefetch table.
	PgStatRecoveryPrefetchTableName = "pg_stat_recovery_prefetch"
	// PgStatReplicationTableName is the name of the pg_stat_replication table.
	PgStatReplicationTableName = "pg_stat_replication"
	// PgStatReplicationSlotsTableName is the name of the pg_stat_replication_slots table.
	PgStatReplicationSlotsTableName = "pg_stat_replication_slots"
	// PgStatSlruTableName is the name of the pg_stat_slru table.
	PgStatSlruTableName = "pg_stat_slru"
	// PgStatSslTableName is the name of the pg_stat_ssl table.
	PgStatSslTableName = "pg_stat_ssl"
	// PgStatSubscriptionTableName is the name of the pg_stat_subscription table.
	PgStatSubscriptionTableName = "pg_stat_subscription"
	// PgStatSubscriptionStatsTableName is the name of the pg_stat_subscription_stats table.
	PgStatSubscriptionStatsTableName = "pg_stat_subscription_stats"
	// PgStatSysIndexesTableName is the name of the pg_stat_sys_indexes table.
	PgStatSysIndexesTableName = "pg_stat_sys_indexes"
	// PgStatSysTablesTableName is the name of the pg_stat_sys_tables table.
	PgStatSysTablesTableName = "pg_stat_sys_tables"
	// PgStatUserFunctionsTableName is the name of the pg_stat_user_functions table.
	PgStatUserFunctionsTableName = "pg_stat_user_functions"
	// PgStatUserIndexesTableName is the name of the pg_stat_user_indexes table.
	PgStatUserIndexesTableName = "pg_stat_user_indexes"
	// PgStatUserTablesTableName is the name of the pg_stat_user_tables table.
	PgStatUserTablesTableName = "pg_stat_user_tables"
	// PgStatWalTableName is the name of the pg_stat_wal table.
	PgStatWalTableName = "pg_stat_wal"
	// PgStatWalReceiverTableName is the name of the pg_stat_wal_receiver table.
	PgStatWalReceiverTableName = "pg_stat_wal_receiver"
	// PgStatXactAllTablesTableName is the name of the pg_stat_xact_all_tables table.
	PgStatXactAllTablesTableName = "pg_stat_xact_all_tables"
	// PgStatXactSysTablesTableName is the name of the pg_stat_xact_sys_tables table.
	PgStatXactSysTablesTableName = "pg_stat_xact_sys_tables"
	// PgStatXactUserFunctionsTableName is the name of the pg_stat_xact_user_functions table.
	PgStatXactUserFunctionsTableName = "pg_stat_xact_user_functions"
	// PgStatXactUserTablesTableName is the name of the pg_stat_xact_user_tables table.
	PgStatXactUserTablesTableName = "pg_stat_xact_user_tables"
	// PgStatIoAllIndexesTableName is the name of the pg_statio_all_indexes table.
	PgStatIoAllIndexesTableName = "pg_statio_all_indexes"
	// PgStatIoAllSequencesTableName is the name of the pg_statio_all_sequences table.
	PgStatIoAllSequencesTableName = "pg_statio_all_sequences"
	// PgStatIoAllTablesTableName is the name of the pg_statio_all_tables table.
	PgStatIoAllTablesTableName = "pg_statio_all_tables"
	// PgStatIoSysIndexesTableName is the name of the pg_statio_sys_indexes table.
	PgStatIoSysIndexesTableName = "pg_statio_sys_indexes"
	// PgStatIoSysSequencesTableName is the name of the pg_statio_sys_sequences table.
	PgStatIoSysSequencesTableName = "pg_statio_sys_sequences"
	// PgStatIoSysTablesTableName is the name of the pg_statio_sys_tables table.
	PgStatIoSysTablesTableName = "pg_statio_sys_tables"
	// PgStatIoUserIndexesTableName is the name of the pg_statio_user_indexes table.
	PgStatIoUserIndexesTableName = "pg_statio_user_indexes"
	// PgStatIoUserSequencesTableName is the name of the pg_statio_user_sequences table.
	PgStatIoUserSequencesTableName = "pg_statio_user_sequences"
	// PgStatIoUserTablesTableName is the name of the pg_statio_user_tables table.
	PgStatIoUserTablesTableName = "pg_statio_user_tables"
	// PgStatsTableName is the name of the pg_stats table.
	PgStatsTableName = "pg_stats"
	// PgStatsExtTableName is the name of the pg_stats_ext table.
	PgStatsExtTableName = "pg_stats_ext"
	// PgStatsExtExprsTableName is the name of the pg_stats_ext_exprs table.
	PgStatsExtExprsTableName = "pg_stats_ext_exprs"
	// PgTablesTableName is the name of the pg_tables table.
	PgTablesTableName = "pg_tables"
	// PgTimezoneAbbrevsTableName is the name of the pg_timezone_abbrevs table.
	PgTimezoneAbbrevsTableName = "pg_timezone_abbrevs"
	// PgTimezoneNamesTableName is the name of the pg_timezone_names table.
	PgTimezoneNamesTableName = "pg_timezone_names"
	// PgUserTableName is the name of the pg_user table.
	PgUserTableName = "pg_user"
	// PgUserMappingsTableName is the name of the pg_user_mappings table.
	PgUserMappingsTableName = "pg_user_mappings"
	// PgViewsTableName is the name of the pg_views table.
	PgViewsTableName = "pg_views"
)

// Oid is an Object Identifier Type.
var Oid = types.Int64

// Xid is a Transaction Identifier Type.
var Xid = types.Int64

// Name is an internal type for object names. The length, max_identifier_length, is defined from NAMEDATALEN - 1.
var Name = types.VarCharType{Length: 63}

// Char is a single-byte internal type.
var Char = types.VarCharType{Length: 1}

// ==================================
//
// 		pg_catalog table schemas
//
// ==================================

var pgAggregateSchema = sql.Schema{}

var pgAmSchema = sql.Schema{}

var pgAmOpSchema = sql.Schema{}

var pgAmProcSchema = sql.Schema{}

var pgAttrDefSchema = sql.Schema{}

var pgAttributeSchema = sql.Schema{}

var pgAuthMembersSchema = sql.Schema{}

var pgAuthIdSchema = sql.Schema{}

var pgCastSchema = sql.Schema{}

// TODO: Not all types are accurate
var pgClassSchema = sql.Schema{
	{Name: "oid", Type: Oid, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relname", Type: Name, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relnamespace", Type: Oid, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "reltype", Type: Oid, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "reloftype", Type: Oid, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relowner", Type: Oid, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relam", Type: Oid, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relfilenode", Type: Oid, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "reltablespace", Type: Oid, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relpages", Type: types.Int32, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "reltuples", Type: types.Float32, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relallvisible", Type: types.Int32, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "reltoastrelid", Type: Oid, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relhasindex", Type: types.Bool, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relisshared", Type: types.Bool, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relpersistence", Type: Char, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relkind", Type: Char, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relnatts", Type: types.Int16, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relchecks", Type: types.Int16, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relhasrules", Type: types.Bool, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relhastriggers", Type: types.Bool, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relhassubclass", Type: types.Bool, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relrowsecurity", Type: types.Bool, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relforcerowsecurity", Type: types.Bool, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relispopulated", Type: types.Bool, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relreplident", Type: Char, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relispartition", Type: types.Bool, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relrewrite", Type: Oid, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relfrozenxid", Type: Xid, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relminmxid", Type: Xid, Default: nil, Nullable: false, Source: PgClassTableName},
	{Name: "relacl", Type: types.Int16, Default: nil, Nullable: true, Source: PgClassTableName},
	{Name: "reloptions", Type: types.Int16, Default: nil, Nullable: true, Source: PgClassTableName},
	{Name: "relpartbound", Type: types.Int16, Default: nil, Nullable: true, Source: PgClassTableName},
}

var pgCollationSchema = sql.Schema{}

var pgConstraintSchema = sql.Schema{}

var pgConversionSchema = sql.Schema{}

// TODO: Implement the rest of pg_database
var pgDatabaseSchema = sql.Schema{
	// {Name: "oid", Type: Oid, Default: nil, Nullable: false, Source: PgDatabaseTableName},
	{Name: "datname", Type: Name, Default: nil, Nullable: false, Source: PgDatabaseTableName},
	// {Name: "datdba", Type: Oid, Default: nil, Nullable: false, Source: PgDatabaseTableName},
	// {Name: "encoding", Type: types.Int32, Default: nil, Nullable: false, Source: PgDatabaseTableName},
	// {Name: "datlocprovider", Type: Char, Default: nil, Nullable: false, Source: PgDatabaseTableName},
	// {Name: "datistemplate", Type: types.Bool, Default: nil, Nullable: false, Source: PgDatabaseTableName},
	// {Name: "datallowconn", Type: types.Bool, Default: nil, Nullable: false, Source: PgDatabaseTableName},
	// {Name: "datconnlimit", Type: types.Int32, Default: nil, Nullable: false, Source: PgDatabaseTableName},
	// {Name: "datfrozenxid", Type: Xid, Default: nil, Nullable: false, Source: PgDatabaseTableName},
	// {Name: "datminmxid", Type: Xid, Default: nil, Nullable: false, Source: PgDatabaseTableName},
	// {Name: "dattablespace", Type: Oid, Default: nil, Nullable: false, Source: PgDatabaseTableName},
	// {Name: "datcollate", Type: types.Text, Default: nil, Nullable: false, Source: PgDatabaseTableName},
	// {Name: "datctype", Type: types.Text, Default: nil, Nullable: false, Source: PgDatabaseTableName},
	// {Name: "daticulocale", Type: types.Text, Default: nil, Nullable: false, Source: PgDatabaseTableName},
	// {Name: "daticurules", Type: types.Text, Default: nil, Nullable: false, Source: PgDatabaseTableName},
	// {Name: "datcollversion", Type: types.Text, Default: nil, Nullable: false, Source: PgDatabaseTableName},
	// {Name: "datacl", Type: []aclitem, Default: nil, Nullable: false, Source: PgDatabaseTableName},
}

var pgDbRoleSettingSchema = sql.Schema{}

var pgDefaultAclSchema = sql.Schema{}

var pgDependSchema = sql.Schema{}

var pgDescriptionSchema = sql.Schema{}

var pgEnumSchema = sql.Schema{}

var pgEventTriggerSchema = sql.Schema{}

var pgExtensionSchema = sql.Schema{}

var pgForeignDataWrapperSchema = sql.Schema{}

var pgForeignServerSchema = sql.Schema{}

var pgForeignTableSchema = sql.Schema{}

var pgIndexSchema = sql.Schema{}

var pgInheritsSchema = sql.Schema{}

var pgInitPrivsSchema = sql.Schema{}

var pgLanguageSchema = sql.Schema{}

var pgLargeObjectSchema = sql.Schema{}

var pgLargeObjectMetadataSchema = sql.Schema{}

var pgNamespaceSchema = sql.Schema{}

var pgOpClassSchema = sql.Schema{}

var pgOperatorSchema = sql.Schema{}

var pgOpFamilySchema = sql.Schema{}

var pgParameterAclSchema = sql.Schema{}

var pgPartitionedTableSchema = sql.Schema{}

var pgPolicySchema = sql.Schema{}

var pgProcSchema = sql.Schema{}

var pgPublicationSchema = sql.Schema{}

var pgPublicationNamespaceSchema = sql.Schema{}

var pgPublicationRelSchema = sql.Schema{}

var pgRangeSchema = sql.Schema{}

var pgReplicationOriginSchema = sql.Schema{}

var pgRewriteSchema = sql.Schema{}

var pgSecLabelSchema = sql.Schema{}

var pgSequenceSchema = sql.Schema{}

var pgShDependSchema = sql.Schema{}

var pgShDescriptionSchema = sql.Schema{}

var pgShSecLabelSchema = sql.Schema{}

var pgStatisticsSchema = sql.Schema{}

var pgStatisticsExtSchema = sql.Schema{}

var pgStatisticsExtDataSchema = sql.Schema{}

var pgSubscriptionSchema = sql.Schema{}

var pgSubscriptionRelSchema = sql.Schema{}

var pgTablespaceSchema = sql.Schema{}

var pgTransformSchema = sql.Schema{}

var pgTriggerSchema = sql.Schema{}

var pgTsConfigSchema = sql.Schema{}

var pgTsConfigMapSchema = sql.Schema{}

var pgTsDictSchema = sql.Schema{}

var pgTsParserSchema = sql.Schema{}

var pgTsTemplateSchema = sql.Schema{}

// TODO: Not all types are accurate
var pgTypeSchema = sql.Schema{
	{Name: "oid", Type: Oid, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typname", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typnamespace", Type: Oid, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typowner", Type: Oid, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typlen", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typbyval", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typtype", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typcategory", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typispreferred", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typisdefined", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typdelim", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typrelid", Type: Oid, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typsubscript", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typelem", Type: Oid, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typarray", Type: Oid, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typinput", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typoutput", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typreceive", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typsend", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typmodin", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typmodout", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typanalyze", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typalign", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typstorage", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typnotnull", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typbasetype", Type: Oid, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typtypmod", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typndims", Type: types.Int32, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typcollation", Type: Oid, Default: nil, Nullable: false, Source: PgTypeTableName},
	{Name: "typdefaultbin", Type: types.Int32, Default: nil, Nullable: true, Source: PgTypeTableName},
	{Name: "typdefault", Type: types.Int32, Default: nil, Nullable: true, Source: PgTypeTableName},
	{Name: "typacl", Type: types.Int32, Default: nil, Nullable: true, Source: PgTypeTableName},
}

var pgUserMappingSchema = sql.Schema{}

var pgCatalogDb = &pgCatalogDatabase{
	name: PgCatalogName,
	tables: map[string]sql.Table{
		PgAggregateTableName: &pgCatalogTable{
			name:   PgAggregateTableName,
			schema: pgAggregateSchema,
			reader: emptyRowIter,
		},
		PgAmTableName: &pgCatalogTable{
			name:   PgAmTableName,
			schema: pgAmSchema,
			reader: emptyRowIter,
		},
		PgAmOpTableName: &pgCatalogTable{
			name:   PgAmOpTableName,
			schema: pgAmOpSchema,
			reader: emptyRowIter,
		},
		PgAmProcTableName: &pgCatalogTable{
			name:   PgAmProcTableName,
			schema: pgAmProcSchema,
			reader: emptyRowIter,
		},
		PgAttrDefTableName: &pgCatalogTable{
			name:   PgAttrDefTableName,
			schema: pgAttrDefSchema,
			reader: emptyRowIter,
		},
		PgAttributeTableName: &pgCatalogTable{
			name:   PgAttributeTableName,
			schema: pgAttributeSchema,
			reader: emptyRowIter,
		},
		PgAuthMembersTableName: &pgCatalogTable{
			name:   PgAuthMembersTableName,
			schema: pgAuthMembersSchema,
			reader: emptyRowIter,
		},
		PgAuthIdTableName: &pgCatalogTable{
			name:   PgAuthIdTableName,
			schema: pgAuthIdSchema,
			reader: emptyRowIter,
		},
		PgCastTableName: &pgCatalogTable{
			name:   PgCastTableName,
			schema: pgCastSchema,
			reader: emptyRowIter,
		},
		PgClassTableName: &pgCatalogTable{
			name:   PgClassTableName,
			schema: pgClassSchema,
			reader: emptyRowIter,
		},
		PgCollationTableName: &pgCatalogTable{
			name:   PgCollationTableName,
			schema: pgCollationSchema,
			reader: emptyRowIter,
		},
		PgConstraintTableName: &pgCatalogTable{
			name:   PgConstraintTableName,
			schema: pgConstraintSchema,
			reader: emptyRowIter,
		},
		PgConversionTableName: &pgCatalogTable{
			name:   PgConversionTableName,
			schema: pgConversionSchema,
			reader: emptyRowIter,
		},
		PgDatabaseTableName: &pgCatalogTable{
			name:   PgDatabaseTableName,
			schema: pgDatabaseSchema,
			reader: databaseRowIter,
		},
		PgDbRoleSettingTableName: &pgCatalogTable{
			name:   PgDbRoleSettingTableName,
			schema: pgDbRoleSettingSchema,
			reader: emptyRowIter,
		},
		PgDefaultAclTableName: &pgCatalogTable{
			name:   PgDefaultAclTableName,
			schema: pgDefaultAclSchema,
			reader: emptyRowIter,
		},
		PgDependTableName: &pgCatalogTable{
			name:   PgDependTableName,
			schema: pgDependSchema,
			reader: emptyRowIter,
		},
		PgDescriptionTableName: &pgCatalogTable{
			name:   PgDescriptionTableName,
			schema: pgDescriptionSchema,
			reader: emptyRowIter,
		},
		PgEnumTableName: &pgCatalogTable{
			name:   PgEnumTableName,
			schema: pgEnumSchema,
			reader: emptyRowIter,
		},
		PgEventTriggerTableName: &pgCatalogTable{
			name:   PgEventTriggerTableName,
			schema: pgEventTriggerSchema,
			reader: emptyRowIter,
		},
		PgExtensionTableName: &pgCatalogTable{
			name:   PgExtensionTableName,
			schema: pgExtensionSchema,
			reader: emptyRowIter,
		},
		PgForeignDataWrapperTableName: &pgCatalogTable{
			name:   PgForeignDataWrapperTableName,
			schema: pgForeignDataWrapperSchema,
			reader: emptyRowIter,
		},
		PgForeignServerTableName: &pgCatalogTable{
			name:   PgForeignServerTableName,
			schema: pgForeignServerSchema,
			reader: emptyRowIter,
		},
		PgForeignTableTableName: &pgCatalogTable{
			name:   PgForeignTableTableName,
			schema: pgForeignTableSchema,
			reader: emptyRowIter,
		},
		PgIndexTableName: &pgCatalogTable{
			name:   PgIndexTableName,
			schema: pgIndexSchema,
			reader: emptyRowIter,
		},
		PgInheritsTableName: &pgCatalogTable{
			name:   PgInheritsTableName,
			schema: pgInheritsSchema,
			reader: emptyRowIter,
		},
		PgInitPrivsTableName: &pgCatalogTable{
			name:   PgInitPrivsTableName,
			schema: pgInitPrivsSchema,
			reader: emptyRowIter,
		},
		PgLanguageTableName: &pgCatalogTable{
			name:   PgLanguageTableName,
			schema: pgLanguageSchema,
			reader: emptyRowIter,
		},
		PgLargeObjectTableName: &pgCatalogTable{
			name:   PgLargeObjectTableName,
			schema: pgLargeObjectSchema,
			reader: emptyRowIter,
		},
		PgLargeObjectMetadataTableName: &pgCatalogTable{
			name:   PgLargeObjectMetadataTableName,
			schema: pgLargeObjectMetadataSchema,
			reader: emptyRowIter,
		},
		PgNamespaceTableName: &pgCatalogTable{
			name:   PgNamespaceTableName,
			schema: pgNamespaceSchema,
			reader: emptyRowIter,
		},
		PgOpClassTableName: &pgCatalogTable{
			name:   PgOpClassTableName,
			schema: pgOpClassSchema,
			reader: emptyRowIter,
		},
		PgOperatorTableName: &pgCatalogTable{
			name:   PgOperatorTableName,
			schema: pgOperatorSchema,
			reader: emptyRowIter,
		},
		PgOpFamilyTableName: &pgCatalogTable{
			name:   PgOpFamilyTableName,
			schema: pgOpFamilySchema,
			reader: emptyRowIter,
		},
		PgParameterAclTableName: &pgCatalogTable{
			name:   PgParameterAclTableName,
			schema: pgParameterAclSchema,
			reader: emptyRowIter,
		},
		PgPartitionedTableTableName: &pgCatalogTable{
			name:   PgPartitionedTableTableName,
			schema: pgPartitionedTableSchema,
			reader: emptyRowIter,
		},
		PgPolicyTableName: &pgCatalogTable{
			name:   PgPolicyTableName,
			schema: pgPolicySchema,
			reader: emptyRowIter,
		},
		PgProcTableName: &pgCatalogTable{
			name:   PgProcTableName,
			schema: pgProcSchema,
			reader: emptyRowIter,
		},
		PgPublicationTableName: &pgCatalogTable{
			name:   PgPublicationTableName,
			schema: pgPublicationSchema,
			reader: emptyRowIter,
		},
		PgPublicationNamespaceTableName: &pgCatalogTable{
			name:   PgPublicationNamespaceTableName,
			schema: pgPublicationNamespaceSchema,
			reader: emptyRowIter,
		},
		PgPublicationRelTableName: &pgCatalogTable{
			name:   PgPublicationRelTableName,
			schema: pgPublicationRelSchema,
			reader: emptyRowIter,
		},
		PgRangeTableName: &pgCatalogTable{
			name:   PgRangeTableName,
			schema: pgRangeSchema,
			reader: emptyRowIter,
		},
		PgReplicationOriginTableName: &pgCatalogTable{
			name:   PgReplicationOriginTableName,
			schema: pgReplicationOriginSchema,
			reader: emptyRowIter,
		},
		PgRewriteTableName: &pgCatalogTable{
			name:   PgRewriteTableName,
			schema: pgRewriteSchema,
			reader: emptyRowIter,
		},
		PgSecLabelTableName: &pgCatalogTable{
			name:   PgSecLabelTableName,
			schema: pgSecLabelSchema,
			reader: emptyRowIter,
		},
		PgSequenceTableName: &pgCatalogTable{
			name:   PgSequenceTableName,
			schema: pgSequenceSchema,
			reader: emptyRowIter,
		},
		PgShDependTableName: &pgCatalogTable{
			name:   PgShDependTableName,
			schema: pgShDependSchema,
			reader: emptyRowIter,
		},
		PgShDescriptionTableName: &pgCatalogTable{
			name:   PgShDescriptionTableName,
			schema: pgShDescriptionSchema,
			reader: emptyRowIter,
		},
		PgShSecLabelTableName: &pgCatalogTable{
			name:   PgShSecLabelTableName,
			schema: pgShSecLabelSchema,
			reader: emptyRowIter,
		},
		PgStatisticsTableName: &pgCatalogTable{
			name:   PgStatisticsTableName,
			schema: pgStatisticsSchema,
			reader: emptyRowIter,
		},
		PgStatisticsExtTableName: &pgCatalogTable{
			name:   PgStatisticsExtTableName,
			schema: pgStatisticsExtSchema,
			reader: emptyRowIter,
		},
		PgStatisticsExtDataTableName: &pgCatalogTable{
			name:   PgStatisticsExtDataTableName,
			schema: pgStatisticsExtDataSchema,
			reader: emptyRowIter,
		},
		PgSubscriptionTableName: &pgCatalogTable{
			name:   PgSubscriptionTableName,
			schema: pgSubscriptionSchema,
			reader: emptyRowIter,
		},
		PgSubscriptionRelTableName: &pgCatalogTable{
			name:   PgSubscriptionRelTableName,
			schema: pgSubscriptionRelSchema,
			reader: emptyRowIter,
		},
		PgTablespaceTableName: &pgCatalogTable{
			name:   PgTablespaceTableName,
			schema: pgTablespaceSchema,
			reader: emptyRowIter,
		},
		PgTransformTableName: &pgCatalogTable{
			name:   PgTransformTableName,
			schema: pgTransformSchema,
			reader: emptyRowIter,
		},
		PgTriggerTableName: &pgCatalogTable{
			name:   PgTriggerTableName,
			schema: pgTriggerSchema,
			reader: emptyRowIter,
		},
		PgTsConfigTableName: &pgCatalogTable{
			name:   PgTsConfigTableName,
			schema: pgTsConfigSchema,
			reader: emptyRowIter,
		},
		PgTsConfigMapTableName: &pgCatalogTable{
			name:   PgTsConfigMapTableName,
			schema: pgTsConfigMapSchema,
			reader: emptyRowIter,
		},
		PgTsDictTableName: &pgCatalogTable{
			name:   PgTsDictTableName,
			schema: pgTsDictSchema,
			reader: emptyRowIter,
		},
		PgTsParserTableName: &pgCatalogTable{
			name:   PgTsParserTableName,
			schema: pgTsParserSchema,
			reader: emptyRowIter,
		},
		PgTsTemplateTableName: &pgCatalogTable{
			name:   PgTsTemplateTableName,
			schema: pgTsTemplateSchema,
			reader: emptyRowIter,
		},
		PgTypeTableName: &pgCatalogTable{
			name:   PgTypeTableName,
			schema: pgTypeSchema,
			reader: emptyRowIter,
		},
		PgUserMappingTableName: &pgCatalogTable{
			name:   PgUserMappingTableName,
			schema: pgUserMappingSchema,
			reader: emptyRowIter,
		},
	},
}
