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

package pgcatalog

import "os"

// PgCatalogName is a constant to the pg_catalog name.
const PgCatalogName = "pg_catalog"

// includeSystemTables is a flag to determine whether to include system tables in the pg_catalog tables.
var includeSystemTables = true

// Init initializes everything necessary for the pg_catalog tables.
func Init() {
	InitIncludeSystemTables()
	InitPgAggregate()
	InitPgAm()
	InitPgAmop()
	InitPgAmproc()
	InitPgAttrdef()
	InitPgAttribute()
	InitPgAuthMembers()
	InitPgAuthid()
	InitPgAvailableExtensionVersions()
	InitPgAvailableExtensions()
	InitPgBackendMemoryContexts()
	InitPgCast()
	InitPgClass()
	InitPgCollation()
	InitPgConfig()
	InitPgConstraint()
	InitPgConversion()
	InitPgCursors()
	InitPgDatabase()
	InitPgDbRoleSetting()
	InitPgDefaultAcl()
	InitPgDepend()
	InitPgDescription()
	InitPgEnum()
	InitPgEventTrigger()
	InitPgExtension()
	InitPgFileSettings()
	InitPgForeignDataWrapper()
	InitPgForeignServer()
	InitPgForeignTable()
	InitPgGroup()
	InitPgHbaFileRules()
	InitPgIdentFileMappings()
	InitPgIndex()
	InitPgIndexes()
	InitPgInherits()
	InitPgInitPrivs()
	InitPgLanguage()
	InitPgLargeobject()
	InitPgLargeobjectMetadata()
	InitPgLocks()
	InitPgMatviews()
	InitPgNamespace()
	InitPgOpclass()
	InitPgOperator()
	InitPgOpfamily()
	InitPgParameterAcl()
	InitPgPartitionedTable()
	InitPgPolicies()
	InitPgPolicy()
	InitPgPreparedStatements()
	InitPgPreparedXacts()
	InitPgProc()
	InitPgPublication()
	InitPgPublicationNamespace()
	InitPgPublicationRel()
	InitPgPublicationTables()
	InitPgRange()
	InitPgReplicationOrigin()
	InitPgReplicationOriginStatus()
	InitPgReplicationSlots()
	InitPgRewrite()
	InitPgRoles()
	InitPgRules()
	InitPgSeclabel()
	InitPgSeclabels()
	InitPgSequence()
	InitPgSequences()
	InitPgSettings()
	InitPgShadow()
	InitPgShdepend()
	InitPgShdescription()
	InitPgShmemAllocations()
	InitPgShseclabel()
	InitPgStatActivity()
	InitPgStatAllIndexes()
	InitPgStatAllTables()
	InitPgStatArchiver()
	InitPgStatBgwriter()
	InitPgStatDatabase()
	InitPgStatDatabaseConflicts()
	InitPgStatGssapi()
	InitPgStatProgressAnalyze()
	InitPgStatProgressBasebackup()
	InitPgStatProgressCluster()
	InitPgStatProgressCopy()
	InitPgStatProgressCreateIndex()
	InitPgStatProgressVacuum()
	InitPgStatRecoveryPrefetch()
	InitPgStatReplication()
	InitPgStatReplicationSlots()
	InitPgStatSlru()
	InitPgStatSsl()
	InitPgStatSubscription()
	InitPgStatSubscriptionStats()
	InitPgStatSysIndexes()
	InitPgStatSysTables()
	InitPgStatUserFunctions()
	InitPgStatUserIndexes()
	InitPgStatUserTables()
	InitPgStatWal()
	InitPgStatWalReceiver()
	InitPgStatXactAllTables()
	InitPgStatXactSysTables()
	InitPgStatXactUserFunctions()
	InitPgStatXactUserTables()
	InitPgStatioAllIndexes()
	InitPgStatioAllSequences()
	InitPgStatioAllTables()
	InitPgStatioSysIndexes()
	InitPgStatioSysSequences()
	InitPgStatioSysTables()
	InitPgStatioUserIndexes()
	InitPgStatioUserSequences()
	InitPgStatioUserTables()
	InitPgStatistic()
	InitPgStatisticExt()
	InitPgStatisticExtData()
	InitPgStats()
	InitPgStatsExt()
	InitPgStatsExtExprs()
	InitPgSubscription()
	InitPgSubscriptionRel()
	InitPgTables()
	InitPgTablespace()
	InitPgTimezoneAbbrevs()
	InitPgTimezoneNames()
	InitPgTransform()
	InitPgTrigger()
	InitPgTsConfig()
	InitPgTsConfigMap()
	InitPgTsDict()
	InitPgTsParser()
	InitPgTsTemplate()
	InitPgType()
	InitPgUser()
	InitPgUserMapping()
	InitPgUserMappings()
	InitPgViews()
}

func InitIncludeSystemTables() {
	if _, ok := os.LookupEnv("REGRESSION_TESTING"); ok {
		// In CI regression tests, we exclude system tables to make them faster.
		// None of them rely on the presence of system tables.
		includeSystemTables = false
	}
}
