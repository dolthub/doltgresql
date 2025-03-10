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

// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in licenses/BSD-vitess.txt.

// Portions of this file are additionally subject to the following
// license and copyright.
//
// Copyright 2015 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// This code was derived from https://github.com/youtube/vitess.

package tree

import "fmt"

// Instructions for creating new types: If a type needs to satisfy an
// interface, declare that function along with that interface. This
// will help users identify the list of types to which they can assert
// those interfaces. If the member of a type has a string with a
// predefined list of values, declare those values as const following
// the type. For interfaces that define dummy functions to
// consolidate a set of types, define the function as typeName().
// This will help avoid name collisions.

// StatementType is the enumerated type for Statement return styles on
// the wire.
type StatementType int

//go:generate stringer -type=StatementType
const (
	// Ack indicates that the statement does not have a meaningful
	// return. Examples include SET, BEGIN, COMMIT.
	Ack StatementType = iota
	// DDL indicates that the statement mutates the database schema.
	//
	// Note: this is the type indicated back to the client; it is not a
	// sufficient test for schema mutation for planning purposes. There
	// are schema-modifying statements (e.g. CREATE TABLE AS) which
	// report RowsAffected to the client, not DDL.
	// Use CanModifySchema() below instead.
	DDL
	// RowsAffected indicates that the statement returns the count of
	// affected rows.
	RowsAffected
	// Rows indicates that the statement returns the affected rows after
	// the statement was applied.
	Rows
	// CopyIn indicates a COPY FROM statement.
	CopyIn
	// Unknown indicates that the statement does not have a known
	// return style at the time of parsing. This is not first in the
	// enumeration because it is more convenient to have Ack as a zero
	// value, and because the use of Unknown should be an explicit choice.
	// The primary example of this statement type is EXECUTE, where the
	// statement type depends on the statement type of the prepared statement
	// being executed.
	Unknown
)

// Statement represents a statement.
type Statement interface {
	fmt.Stringer
	NodeFormatter
	StatementType() StatementType
	// StatementTag is a short string identifying the type of statement
	// (usually a single verb). This is different than the Stringer output,
	// which is the actual statement (including args).
	// TODO(dt): Currently tags are always pg-compatible in the future it
	// might make sense to pass a tag format specifier.
	StatementTag() string
}

// canModifySchema is to be implemented by statements that can modify
// the database schema but may have StatementType() != DDL.
// See CanModifySchema() below.
type canModifySchema interface {
	modifiesSchema() bool
}

// CanModifySchema returns true if the statement can modify
// the database schema.
func CanModifySchema(stmt Statement) bool {
	if stmt.StatementType() == DDL {
		return true
	}
	scm, ok := stmt.(canModifySchema)
	return ok && scm.modifiesSchema()
}

// CanWriteData returns true if the statement can modify data.
func CanWriteData(stmt Statement) bool {
	switch stmt.(type) {
	// Normal write operations.
	case *Insert, *Delete, *Update, *Truncate:
		return true
	// Import operations.
	case *CopyFrom, *Import, *Restore:
		return true
	// CockroachDB extensions.
	case *Split, *Unsplit, *Relocate, *Scatter:
		return true
	}
	return false
}

// IsStmtParallelized determines if a given statement's execution should be
// parallelized. This means that its results should be mocked out, and that
// it should be run asynchronously and in parallel with other statements that
// are independent.
func IsStmtParallelized(stmt Statement) bool {
	parallelizedRetClause := func(ret ReturningClause) bool {
		_, ok := ret.(*ReturningNothing)
		return ok
	}
	switch s := stmt.(type) {
	case *Delete:
		return parallelizedRetClause(s.Returning)
	case *Insert:
		return parallelizedRetClause(s.Returning)
	case *Update:
		return parallelizedRetClause(s.Returning)
	}
	return false
}

// HiddenFromShowQueries is a pseudo-interface to be implemented
// by statements that should not show up in SHOW QUERIES (and are hence
// not cancellable using CANCEL QUERIES either). Usually implemented by
// statements that spawn jobs.
type HiddenFromShowQueries interface {
	hiddenFromShowQueries()
}

// ObserverStatement is a marker interface for statements which are allowed to
// run regardless of the current transaction state: statements other than
// rollback are generally rejected if the session is in a failed transaction
// state, but it's convenient to allow some statements (e.g. "show syntax; set
// tracing").
// Such statements are not expected to modify the database, the transaction or
// session state (other than special cases such as enabling/disabling tracing).
//
// These statements short-circuit the regular execution - they don't get planned
// (there are no corresponding planNodes). The connExecutor recognizes them and
// handles them.
type ObserverStatement interface {
	observerStatement()
}

// CCLOnlyStatement is a marker interface for statements that require
// a CCL binary for successful planning or execution.
// It is used to enhance error messages when attempting to use these
// statements in non-CCL binaries.
type CCLOnlyStatement interface {
	cclOnlyStatement()
}

var _ CCLOnlyStatement = &Backup{}
var _ CCLOnlyStatement = &ShowBackup{}
var _ CCLOnlyStatement = &Restore{}
var _ CCLOnlyStatement = &CreateChangefeed{}
var _ CCLOnlyStatement = &Import{}
var _ CCLOnlyStatement = &Export{}
var _ CCLOnlyStatement = &ScheduledBackup{}

// StatementType implements the Statement interface.
func (*AlterAggregate) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*AlterAggregate) StatementTag() string { return "ALTER AGGREGATE" }

func (*AlterAggregate) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*AlterCollation) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*AlterCollation) StatementTag() string { return "ALTER COLLATION" }

func (*AlterCollation) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*AlterConversion) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*AlterConversion) StatementTag() string { return "ALTER CONVERSION" }

func (*AlterConversion) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*AlterDatabase) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*AlterDatabase) StatementTag() string { return "ALTER DATABASE" }

func (*AlterDatabase) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*AlterDomain) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*AlterDomain) StatementTag() string { return "ALTER DOMAIN" }

func (*AlterDomain) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*AlterDefaultPrivileges) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*AlterDefaultPrivileges) StatementTag() string { return "ALTER DEFAULT PRIVILEGES" }

func (*AlterDefaultPrivileges) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*AlterFunction) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*AlterFunction) StatementTag() string { return "ALTER FUNCTION" }

func (*AlterFunction) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*AlterIndex) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*AlterIndex) StatementTag() string { return "ALTER INDEX" }

func (*AlterIndex) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*AlterIndexAllInTablespace) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*AlterIndexAllInTablespace) StatementTag() string { return "ALTER INDEX ALL IN TABLESPACE" }

func (*AlterIndexAllInTablespace) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*AlterLanguage) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*AlterLanguage) StatementTag() string { return "ALTER LANGUAGE" }

func (*AlterLanguage) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*AlterMaterializedView) StatementType() StatementType { return DDL }

// StatementTag implements the Statement interface.
func (*AlterMaterializedView) StatementTag() string { return "ALTER MATERIALIZED VIEW" }

func (*AlterMaterializedView) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*AlterProcedure) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*AlterProcedure) StatementTag() string { return "ALTER PROCEDURE" }

func (*AlterProcedure) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*AlterTable) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*AlterTable) StatementTag() string { return "ALTER TABLE" }

func (*AlterTable) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*AlterTableSetSchema) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*AlterTableSetSchema) StatementTag() string { return "ALTER TABLE SET SCHEMA" }

func (*AlterTableSetSchema) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*AlterTableAllInTablespace) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*AlterTableAllInTablespace) StatementTag() string { return "ALTER TABLE ALL IN TABLESPACE" }

func (*AlterTableAllInTablespace) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*AlterTablePartition) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*AlterTablePartition) StatementTag() string { return "ALTER TABLE ... PARTITION" }

func (*AlterTablePartition) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*AlterTrigger) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*AlterTrigger) StatementTag() string { return "ALTER TRIGGER" }

func (*AlterTrigger) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*AlterSchema) StatementType() StatementType { return DDL }

// StatementTag implements the Statement interface.
func (*AlterSchema) StatementTag() string { return "ALTER SCHEMA" }

func (*AlterSchema) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*AlterType) StatementType() StatementType { return DDL }

// StatementTag implements the Statement interface.
func (*AlterType) StatementTag() string { return "ALTER TYPE" }

func (*AlterType) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*AlterView) StatementType() StatementType { return DDL }

// StatementTag implements the Statement interface.
func (*AlterView) StatementTag() string { return "ALTER VIEW" }

func (*AlterView) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*AlterSequence) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*AlterSequence) StatementTag() string { return "ALTER SEQUENCE" }

// StatementType implements the Statement interface.
func (*AlterRole) StatementType() StatementType { return Ack }

// StatementTag returns a short string identifying the type of statement.
func (*AlterRole) StatementTag() string { return "ALTER ROLE" }

func (*AlterRole) cclOnlyStatement() {}

func (*AlterRole) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*Analyze) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*Analyze) StatementTag() string { return "ANALYZE" }

// StatementType implements the Statement interface.
func (*Backup) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*Backup) StatementTag() string { return "BACKUP" }

func (*Backup) cclOnlyStatement() {}

func (*Backup) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*ScheduledBackup) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ScheduledBackup) StatementTag() string { return "SCHEDULED BACKUP" }

func (*ScheduledBackup) cclOnlyStatement() {}

func (*ScheduledBackup) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*BeginTransaction) StatementType() StatementType { return Ack }

// StatementTag returns a short string identifying the type of statement.
func (*BeginTransaction) StatementTag() string { return "BEGIN" }

// StatementType implements the Statement interface.
func (*BeginEndBlock) StatementType() StatementType { return Unknown }

// StatementTag returns a short string identifying the type of statement.
func (*BeginEndBlock) StatementTag() string { return "BEGIN ... END" }

// StatementType implements the Statement interface.
func (*ControlJobs) StatementType() StatementType { return RowsAffected }

// StatementTag returns a short string identifying the type of statement.
func (n *ControlJobs) StatementTag() string {
	return fmt.Sprintf("%s JOBS", JobCommandToStatement[n.Command])
}

// StatementType implements the Statement interface.
func (*ControlSchedules) StatementType() StatementType { return RowsAffected }

// StatementTag returns a short string identifying the type of statement.
func (n *ControlSchedules) StatementTag() string {
	return fmt.Sprintf("%s SCHEDULES", n.Command)
}

// StatementType implements the Statement interface.
func (*ControlJobsForSchedules) StatementType() StatementType { return RowsAffected }

// StatementTag returns a short string identifying the type of statement.
func (n *ControlJobsForSchedules) StatementTag() string {
	return fmt.Sprintf("%s JOBS FOR SCHEDULES", JobCommandToStatement[n.Command])
}

// StatementType implements the Statement interface.
func (*CancelQueries) StatementType() StatementType { return RowsAffected }

// StatementTag returns a short string identifying the type of statement.
func (*CancelQueries) StatementTag() string { return "CANCEL QUERIES" }

// StatementType implements the Statement interface.
func (*CancelSessions) StatementType() StatementType { return RowsAffected }

// StatementTag returns a short string identifying the type of statement.
func (*CancelSessions) StatementTag() string { return "CANCEL SESSIONS" }

// StatementType implements the Statement interface.
func (*CannedOptPlan) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*CannedOptPlan) StatementTag() string { return "PREPARE AS OPT PLAN" }

// StatementType implements the Statement interface.
func (*Comment) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*Comment) StatementTag() string { return "COMMENT" }

// StatementType implements the Statement interface.
func (*CommitTransaction) StatementType() StatementType { return Ack }

// StatementTag returns a short string identifying the type of statement.
func (*CommitTransaction) StatementTag() string { return "COMMIT" }

// StatementType implements the Statement interface.
func (*CopyFrom) StatementType() StatementType { return CopyIn }

// StatementTag returns a short string identifying the type of statement.
func (*CopyFrom) StatementTag() string { return "COPY" }

// StatementType implements the Statement interface.
func (*CreateAggregate) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*CreateAggregate) StatementTag() string { return "CREATE AGGREGATE" }

// StatementType implements the Statement interface.
func (*CreateChangefeed) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (n *CreateChangefeed) StatementTag() string {
	if n.SinkURI == nil {
		return "EXPERIMENTAL CHANGEFEED"
	}
	return "CREATE CHANGEFEED"
}

func (*CreateChangefeed) cclOnlyStatement() {}

// StatementType implements the Statement interface.
func (*CreateDatabase) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*CreateDatabase) StatementTag() string { return "CREATE DATABASE" }

// StatementType implements the Statement interface.
func (*CreateDomain) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*CreateDomain) StatementTag() string { return "CREATE DOMAIN" }

// StatementType implements the Statement interface.
func (*CreateExtension) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*CreateExtension) StatementTag() string { return "CREATE EXTENSION" }

// StatementType implements the Statement interface.
func (*CreateFunction) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*CreateFunction) StatementTag() string { return "CREATE FUNCTION" }

// StatementType implements the Statement interface.
func (*CreateIndex) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*CreateIndex) StatementTag() string { return "CREATE INDEX" }

// StatementType implements the Statement interface.
func (*CreateLanguage) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*CreateLanguage) StatementTag() string { return "CREATE LANGUAGE" }

// StatementType implements the Statement interface.
func (*CreateMaterializedView) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*CreateMaterializedView) StatementTag() string { return "CREATE MATERIALIZED VIEW" }

// StatementType implements the Statement interface.
func (*CreateProcedure) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*CreateProcedure) StatementTag() string { return "CREATE VIEW" }

// StatementType implements the Statement interface.
func (n *CreateSchema) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (n *CreateSchema) StatementTag() string {
	return "CREATE SCHEMA"
}

// modifiesSchema implements the canModifySchema interface.
func (*CreateSchema) modifiesSchema() bool { return true }

// StatementType implements the Statement interface.
func (n *CreateTable) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (n *CreateTable) StatementTag() string {
	if n.As() {
		return "CREATE TABLE AS"
	}
	return "CREATE TABLE"
}

// modifiesSchema implements the canModifySchema interface.
func (*CreateTable) modifiesSchema() bool { return true }

// StatementType implements the Statement interface.
func (*CreateTrigger) StatementType() StatementType { return DDL }

// StatementTag implements the Statement interface.
func (*CreateTrigger) StatementTag() string { return "CREATE TRIGGER" }

func (*CreateTrigger) modifiesSchema() bool { return true }

// StatementType implements the Statement interface.
func (*CreateType) StatementType() StatementType { return DDL }

// StatementTag implements the Statement interface.
func (*CreateType) StatementTag() string { return "CREATE TYPE" }

func (*CreateType) modifiesSchema() bool { return true }

// StatementType implements the Statement interface.
func (*CreateRole) StatementType() StatementType { return Ack }

// StatementTag returns a short string identifying the type of statement.
func (*CreateRole) StatementTag() string { return "CREATE ROLE" }

func (*CreateRole) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*CreateView) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*CreateView) StatementTag() string { return "CREATE VIEW" }

// StatementType implements the Statement interface.
func (*CreateSequence) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*CreateSequence) StatementTag() string { return "CREATE SEQUENCE" }

// StatementType implements the Statement interface.
func (*CreateStats) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*CreateStats) StatementTag() string { return "CREATE STATISTICS" }

// StatementType implements the Statement interface.
func (*Deallocate) StatementType() StatementType { return Ack }

// StatementTag returns a short string identifying the type of statement.
func (n *Deallocate) StatementTag() string {
	// Postgres distinguishes the command tags for these two cases of Deallocate statements.
	if n.Name == "" {
		return "DEALLOCATE ALL"
	}
	return "DEALLOCATE"
}

// StatementType implements the Statement interface.
func (*Discard) StatementType() StatementType { return Ack }

// StatementTag returns a short string identifying the type of statement.
func (*Discard) StatementTag() string { return "DISCARD" }

// StatementType implements the Statement interface.
func (n *Delete) StatementType() StatementType { return n.Returning.statementType() }

// StatementTag returns a short string identifying the type of statement.
func (*Delete) StatementTag() string { return "DELETE" }

// StatementType implements the Statement interface.
func (*DropAggregate) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*DropAggregate) StatementTag() string { return "DROP AGGREGATE" }

// StatementType implements the Statement interface.
func (*DropDatabase) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*DropDatabase) StatementTag() string { return "DROP DATABASE" }

// StatementType implements the Statement interface.
func (*DropDomain) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*DropDomain) StatementTag() string { return "DROP DOMAIN" }

// StatementType implements the Statement interface.
func (*DropExtension) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*DropExtension) StatementTag() string { return "DROP EXTENSION" }

// StatementType implements the Statement interface.
func (*DropFunction) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*DropFunction) StatementTag() string { return "DROP FUNCTION" }

// StatementType implements the Statement interface.
func (*DropIndex) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*DropIndex) StatementTag() string { return "DROP INDEX" }

// StatementType implements the Statement interface.
func (*DropLanguage) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*DropLanguage) StatementTag() string { return "DROP LANGUAGE" }

// StatementType implements the Statement interface.
func (*DropProcedure) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*DropProcedure) StatementTag() string { return "DROP PROCEDURE" }

// StatementType implements the Statement interface.
func (*DropTable) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*DropTable) StatementTag() string { return "DROP TABLE" }

// StatementType implements the Statement interface.
func (*DropTrigger) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*DropTrigger) StatementTag() string { return "DROP TRIGGER" }

// StatementType implements the Statement interface.
func (*DropView) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*DropView) StatementTag() string { return "DROP VIEW" }

// StatementType implements the Statement interface.
func (*DropSequence) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*DropSequence) StatementTag() string { return "DROP SEQUENCE" }

// StatementType implements the Statement interface.
func (*DropRole) StatementType() StatementType { return Ack }

// StatementTag returns a short string identifying the type of statement.
func (*DropRole) StatementTag() string { return "DROP ROLE" }

func (*DropRole) cclOnlyStatement() {}

func (*DropRole) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*DropType) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*DropType) StatementTag() string { return "DROP TYPE" }

// StatementType implements the Statement interface.
func (*DropSchema) StatementType() StatementType { return DDL }

// StatementTag implements the Statement interface.
func (*DropSchema) StatementTag() string { return "DROP SCHEMA" }

// StatementType implements the Statement interface.
func (*Execute) StatementType() StatementType { return Unknown }

// StatementTag returns a short string identifying the type of statement.
func (*Execute) StatementTag() string { return "EXECUTE" }

// StatementType implements the Statement interface.
func (*Explain) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*Explain) StatementTag() string { return "EXPLAIN" }

// StatementType implements the Statement interface.
func (*ExplainAnalyzeDebug) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ExplainAnalyzeDebug) StatementTag() string { return "EXPLAIN ANALYZE (DEBUG)" }

// StatementType implements the Statement interface.
func (*Export) StatementType() StatementType { return Rows }

func (*Export) cclOnlyStatement() {}

// StatementTag returns a short string identifying the type of statement.
func (*Export) StatementTag() string { return "EXPORT" }

// StatementType implements the Statement interface.
func (*Grant) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*Grant) StatementTag() string { return "GRANT" }

// StatementType implements the Statement interface.
func (*GrantRole) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*GrantRole) StatementTag() string { return "GRANT" }

// StatementType implements the Statement interface.
func (n *Insert) StatementType() StatementType { return n.Returning.statementType() }

// StatementTag returns a short string identifying the type of statement.
func (*Insert) StatementTag() string { return "INSERT" }

// StatementType implements the Statement interface.
func (n *Import) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*Import) StatementTag() string { return "IMPORT" }

func (*Import) cclOnlyStatement() {}

// StatementType implements the Statement interface.
func (*ParenSelect) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ParenSelect) StatementTag() string { return "SELECT" }

// StatementType implements the Statement interface.
func (*Prepare) StatementType() StatementType { return Ack }

// StatementTag returns a short string identifying the type of statement.
func (*Prepare) StatementTag() string { return "PREPARE" }

// StatementType implements the Statement interface.
func (*RefreshMaterializedView) StatementType() StatementType { return DDL }

// StatementTag implements the Statement interface.
func (*RefreshMaterializedView) StatementTag() string { return "REFRESH MATERIALIZED VIEW" }

// StatementType implements the Statement interface.
func (*ReleaseSavepoint) StatementType() StatementType { return Ack }

// StatementTag returns a short string identifying the type of statement.
func (*ReleaseSavepoint) StatementTag() string { return "RELEASE" }

// StatementType implements the Statement interface.
func (*RenameColumn) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*RenameColumn) StatementTag() string { return "RENAME COLUMN" }

// StatementType implements the Statement interface.
func (*RenameDatabase) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*RenameDatabase) StatementTag() string { return "RENAME DATABASE" }

// StatementType implements the Statement interface.
func (*ReparentDatabase) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*ReparentDatabase) StatementTag() string { return "CONVERT TO SCHEMA" }

// StatementType implements the Statement interface.
func (*RenameIndex) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*RenameIndex) StatementTag() string { return "RENAME INDEX" }

// StatementType implements the Statement interface.
func (*RenameTable) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (n *RenameTable) StatementTag() string {
	if n.IsMaterialized {
		return "RENAME MATERIALIZED VIEW"
	} else if n.IsSequence {
		return "RENAME SEQUENCE"
	}
	return "RENAME TABLE"
}

// StatementType implements the Statement interface.
func (*Relocate) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (n *Relocate) StatementTag() string {
	if n.RelocateLease {
		return "EXPERIMENTAL_RELOCATE LEASE"
	}
	return "EXPERIMENTAL_RELOCATE"
}

// StatementType implements the Statement interface.
func (*ResetAll) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ResetAll) StatementTag() string { return "RESET ALL" }

// StatementType implements the Statement interface.
func (*Restore) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*Restore) StatementTag() string { return "RESTORE" }

func (*Restore) cclOnlyStatement() {}

func (*Restore) hiddenFromShowQueries() {}

// StatementType implements the Statement interface.
func (*Return) StatementType() StatementType { return Unknown }

// StatementTag returns a short string identifying the type of statement.
func (*Return) StatementTag() string { return "RETURN" }

// StatementType implements the Statement interface.
func (*Revoke) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*Revoke) StatementTag() string { return "REVOKE" }

// StatementType implements the Statement interface.
func (*RevokeRole) StatementType() StatementType { return DDL }

// StatementTag returns a short string identifying the type of statement.
func (*RevokeRole) StatementTag() string { return "REVOKE" }

// StatementType implements the Statement interface.
func (*RollbackToSavepoint) StatementType() StatementType { return Ack }

// StatementTag returns a short string identifying the type of statement.
func (*RollbackToSavepoint) StatementTag() string { return "ROLLBACK" }

// StatementType implements the Statement interface.
func (*RollbackTransaction) StatementType() StatementType { return Ack }

// StatementTag returns a short string identifying the type of statement.
func (*RollbackTransaction) StatementTag() string { return "ROLLBACK" }

// StatementType implements the Statement interface.
func (*Savepoint) StatementType() StatementType { return Ack }

// StatementTag returns a short string identifying the type of statement.
func (*Savepoint) StatementTag() string { return "SAVEPOINT" }

// StatementType implements the Statement interface.
func (*Scatter) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*Scatter) StatementTag() string { return "SCATTER" }

// StatementType implements the Statement interface.
func (*Scrub) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (n *Scrub) StatementTag() string { return "SCRUB" }

// StatementType implements the Statement interface.
func (*Select) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*Select) StatementTag() string { return "SELECT" }

// StatementType implements the Statement interface.
func (*SelectClause) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*SelectClause) StatementTag() string { return "SELECT" }

// StatementType implements the Statement interface.
func (*SetVar) StatementType() StatementType { return Ack }

// StatementTag returns a short string identifying the type of statement.
func (*SetVar) StatementTag() string { return "SET" }

// StatementType implements the Statement interface.
func (*SetConstraints) StatementType() StatementType { return Ack }

// StatementTag returns a short string identifying the type of statement.
func (*SetConstraints) StatementTag() string { return "SET CONSTRAINTS" }

// StatementType implements the Statement interface.
func (*SetRole) StatementType() StatementType { return Ack }

// StatementTag returns a short string identifying the type of statement.
func (*SetRole) StatementTag() string { return "SET ROLE" }

// StatementType implements the Statement interface.
func (*SetTransaction) StatementType() StatementType { return Ack }

// StatementTag returns a short string identifying the type of statement.
func (*SetTransaction) StatementTag() string { return "SET TRANSACTION" }

// StatementType implements the Statement interface.
func (*SetSessionAuthorization) StatementType() StatementType { return Ack }

// StatementTag returns a short string identifying the type of statement.
func (*SetSessionAuthorization) StatementTag() string { return "SET" }

// StatementType implements the Statement interface.
func (*SetSessionCharacteristics) StatementType() StatementType { return Ack }

// StatementTag returns a short string identifying the type of statement.
func (*SetSessionCharacteristics) StatementTag() string { return "SET" }

// StatementType implements the Statement interface.
func (*ShowVar) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowVar) StatementTag() string { return "SHOW" }

// StatementType implements the Statement interface.
func (*ShowColumns) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowColumns) StatementTag() string { return "SHOW COLUMNS" }

// StatementType implements the Statement interface.
func (*ShowCreate) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowCreate) StatementTag() string { return "SHOW CREATE" }

// StatementType implements the Statement interface.
func (*ShowBackup) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowBackup) StatementTag() string { return "SHOW BACKUP" }

func (*ShowBackup) cclOnlyStatement() {}

// StatementType implements the Statement interface.
func (*ShowDatabases) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowDatabases) StatementTag() string { return "SHOW DATABASES" }

// StatementType implements the Statement interface.
func (*ShowEnums) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowEnums) StatementTag() string { return "SHOW ENUMS" }

// StatementType implements the Statement interface.
func (*ShowTypes) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowTypes) StatementTag() string { return "SHOW TYPES" }

// StatementType implements the Statement interface.
func (*ShowTraceForSession) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowTraceForSession) StatementTag() string { return "SHOW TRACE FOR SESSION" }

// StatementType implements the Statement interface.
func (*ShowGrants) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowGrants) StatementTag() string { return "SHOW GRANTS" }

// StatementType implements the Statement interface.
func (*ShowDatabaseIndexes) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowDatabaseIndexes) StatementTag() string { return "SHOW INDEXES FROM DATABASE" }

// StatementType implements the Statement interface.
func (*ShowIndexes) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowIndexes) StatementTag() string { return "SHOW INDEXES FROM TABLE" }

// StatementType implements the Statement interface.
func (*ShowPartitions) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of the statement.
func (*ShowPartitions) StatementTag() string { return "SHOW PARTITIONS" }

// StatementType implements the Statement interface.
func (*ShowQueries) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowQueries) StatementTag() string { return "SHOW QUERIES" }

// StatementType implements the Statement interface.
func (*ShowJobs) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowJobs) StatementTag() string { return "SHOW JOBS" }

// StatementType implements the Statement interface.
func (*ShowRoleGrants) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowRoleGrants) StatementTag() string { return "SHOW GRANTS ON ROLE" }

// StatementType implements the Statement interface.
func (*ShowSessions) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowSessions) StatementTag() string { return "SHOW SESSIONS" }

// StatementType implements the Statement interface.
func (*ShowTableStats) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowTableStats) StatementTag() string { return "SHOW STATISTICS" }

// StatementType implements the Statement interface.
func (*ShowHistogram) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowHistogram) StatementTag() string { return "SHOW HISTOGRAM" }

// StatementType implements the Statement interface.
func (*ShowSchedules) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowSchedules) StatementTag() string { return "SHOW SCHEDULES" }

// StatementType implements the Statement interface.
func (*ShowSyntax) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowSyntax) StatementTag() string { return "SHOW SYNTAX" }

func (*ShowSyntax) observerStatement() {}

// StatementType implements the Statement interface.
func (*ShowTransactionStatus) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowTransactionStatus) StatementTag() string { return "SHOW TRANSACTION STATUS" }

func (*ShowTransactionStatus) observerStatement() {}

// StatementType implements the Statement interface.
func (*ShowSavepointStatus) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowSavepointStatus) StatementTag() string { return "SHOW SAVEPOINT STATUS" }

func (*ShowSavepointStatus) observerStatement() {}

// StatementType implements the Statement interface.
func (*ShowLastQueryStatistics) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowLastQueryStatistics) StatementTag() string { return "SHOW LAST QUERY STATISTICS" }

func (*ShowLastQueryStatistics) observerStatement() {}

// StatementType implements the Statement interface.
func (*ShowUsers) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowUsers) StatementTag() string { return "SHOW USERS" }

// StatementType implements the Statement interface.
func (*ShowRoles) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowRoles) StatementTag() string { return "SHOW ROLES" }

// StatementType implements the Statement interface.
func (*ShowFingerprints) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowFingerprints) StatementTag() string { return "SHOW EXPERIMENTAL_FINGERPRINTS" }

// StatementType implements the Statement interface.
func (*ShowConstraints) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowConstraints) StatementTag() string { return "SHOW CONSTRAINTS" }

// StatementType implements the Statement interface.
func (*ShowTables) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowTables) StatementTag() string { return "SHOW TABLES" }

// StatementType implements the Statement interface
func (*ShowTransactions) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowTransactions) StatementTag() string { return "SHOW TRANSACTIONS" }

// StatementType implements the Statement interface.
func (*ShowSchemas) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowSchemas) StatementTag() string { return "SHOW SCHEMAS" }

// StatementType implements the Statement interface.
func (*ShowSequences) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ShowSequences) StatementTag() string { return "SHOW SCHEMAS" }

// StatementType implements the Statement interface.
func (*Split) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*Split) StatementTag() string { return "SPLIT" }

// StatementType implements the Statement interface.
func (*Unsplit) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*Unsplit) StatementTag() string { return "UNSPLIT" }

// StatementType implements the Statement interface.
func (*Truncate) StatementType() StatementType { return Ack }

// StatementTag returns a short string identifying the type of statement.
func (*Truncate) StatementTag() string { return "TRUNCATE" }

// modifiesSchema implements the canModifySchema interface.
func (*Truncate) modifiesSchema() bool { return true }

// StatementType implements the Statement interface.
func (n *Update) StatementType() StatementType { return n.Returning.statementType() }

// StatementTag returns a short string identifying the type of statement.
func (*Update) StatementTag() string { return "UPDATE" }

// StatementType implements the Statement interface.
func (*UnionClause) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*UnionClause) StatementTag() string { return "UNION" }

// StatementType implements the Statement interface.
func (*ValuesClause) StatementType() StatementType { return Rows }

// StatementTag returns a short string identifying the type of statement.
func (*ValuesClause) StatementTag() string { return "VALUES" }

func (n *AlterAggregate) String() string            { return AsString(n) }
func (n *AlterCollation) String() string            { return AsString(n) }
func (n *AlterConversion) String() string           { return AsString(n) }
func (n *AlterDatabase) String() string             { return AsString(n) }
func (n *AlterDomain) String() string               { return AsString(n) }
func (n *AlterDefaultPrivileges) String() string    { return AsString(n) }
func (n *AlterFunction) String() string             { return AsString(n) }
func (n *AlterIndex) String() string                { return AsString(n) }
func (n *AlterIndexAllInTablespace) String() string { return AsString(n) }
func (n *AlterLanguage) String() string             { return AsString(n) }
func (n *AlterMaterializedView) String() string     { return AsString(n) }
func (n *AlterProcedure) String() string            { return AsString(n) }
func (n *AlterSchema) String() string               { return AsString(n) }
func (n *AlterTable) String() string                { return AsString(n) }
func (n *AlterTableAllInTablespace) String() string { return AsString(n) }
func (n *AlterTableSetSchema) String() string       { return AsString(n) }
func (n *AlterTablePartition) String() string       { return AsString(n) }
func (n *AlterTrigger) String() string              { return AsString(n) }
func (n *AlterType) String() string                 { return AsString(n) }
func (n *AlterView) String() string                 { return AsString(n) }
func (n *AlterRole) String() string                 { return AsString(n) }
func (n *AlterSequence) String() string             { return AsString(n) }
func (n *Analyze) String() string                   { return AsString(n) }
func (n *Backup) String() string                    { return AsString(n) }
func (n *BeginTransaction) String() string          { return AsString(n) }
func (n *BeginEndBlock) String() string             { return AsString(n) }
func (n *ControlJobs) String() string               { return AsString(n) }
func (n *ControlSchedules) String() string          { return AsString(n) }
func (n *ControlJobsForSchedules) String() string   { return AsString(n) }
func (n *CancelQueries) String() string             { return AsString(n) }
func (n *CancelSessions) String() string            { return AsString(n) }
func (n *CannedOptPlan) String() string             { return AsString(n) }
func (n *Comment) String() string                   { return AsString(n) }
func (n *CommitTransaction) String() string         { return AsString(n) }
func (n *CopyFrom) String() string                  { return AsString(n) }
func (n *CreateAggregate) String() string           { return AsString(n) }
func (n *CreateChangefeed) String() string          { return AsString(n) }
func (n *CreateDatabase) String() string            { return AsString(n) }
func (n *CreateDomain) String() string              { return AsString(n) }
func (n *CreateExtension) String() string           { return AsString(n) }
func (n *CreateFunction) String() string            { return AsString(n) }
func (n *CreateIndex) String() string               { return AsString(n) }
func (n *CreateLanguage) String() string            { return AsString(n) }
func (n *CreateMaterializedView) String() string    { return AsString(n) }
func (n *CreateProcedure) String() string           { return AsString(n) }
func (n *CreateRole) String() string                { return AsString(n) }
func (n *CreateTable) String() string               { return AsString(n) }
func (n *CreateTrigger) String() string             { return AsString(n) }
func (n *CreateType) String() string                { return AsString(n) }
func (n *CreateSchema) String() string              { return AsString(n) }
func (n *CreateSequence) String() string            { return AsString(n) }
func (n *CreateStats) String() string               { return AsString(n) }
func (n *CreateView) String() string                { return AsString(n) }
func (n *Deallocate) String() string                { return AsString(n) }
func (n *Delete) String() string                    { return AsString(n) }
func (n *DropAggregate) String() string             { return AsString(n) }
func (n *DropDatabase) String() string              { return AsString(n) }
func (n *DropDomain) String() string                { return AsString(n) }
func (n *DropExtension) String() string             { return AsString(n) }
func (n *DropFunction) String() string              { return AsString(n) }
func (n *DropIndex) String() string                 { return AsString(n) }
func (n *DropLanguage) String() string              { return AsString(n) }
func (n *DropProcedure) String() string             { return AsString(n) }
func (n *DropSchema) String() string                { return AsString(n) }
func (n *DropTable) String() string                 { return AsString(n) }
func (n *DropTrigger) String() string               { return AsString(n) }
func (n *DropType) String() string                  { return AsString(n) }
func (n *DropView) String() string                  { return AsString(n) }
func (n *DropSequence) String() string              { return AsString(n) }
func (n *DropRole) String() string                  { return AsString(n) }
func (n *Execute) String() string                   { return AsString(n) }
func (n *Explain) String() string                   { return AsString(n) }
func (n *ExplainAnalyzeDebug) String() string       { return AsString(n) }
func (n *Export) String() string                    { return AsString(n) }
func (n *Grant) String() string                     { return AsString(n) }
func (n *GrantRole) String() string                 { return AsString(n) }
func (n *Insert) String() string                    { return AsString(n) }
func (n *Import) String() string                    { return AsString(n) }
func (n *ParenSelect) String() string               { return AsString(n) }
func (n *Prepare) String() string                   { return AsString(n) }
func (n *ReleaseSavepoint) String() string          { return AsString(n) }
func (n *Relocate) String() string                  { return AsString(n) }
func (n *RefreshMaterializedView) String() string   { return AsString(n) }
func (n *RenameColumn) String() string              { return AsString(n) }
func (n *RenameDatabase) String() string            { return AsString(n) }
func (n *ReparentDatabase) String() string          { return AsString(n) }
func (n *RenameIndex) String() string               { return AsString(n) }
func (n *RenameTable) String() string               { return AsString(n) }
func (n *ResetAll) String() string                  { return AsString(n) }
func (n *Restore) String() string                   { return AsString(n) }
func (n *Return) String() string                    { return AsString(n) }
func (n *Revoke) String() string                    { return AsString(n) }
func (n *RevokeRole) String() string                { return AsString(n) }
func (n *RollbackToSavepoint) String() string       { return AsString(n) }
func (n *RollbackTransaction) String() string       { return AsString(n) }
func (n *Savepoint) String() string                 { return AsString(n) }
func (n *Scatter) String() string                   { return AsString(n) }
func (n *ScheduledBackup) String() string           { return AsString(n) }
func (n *Scrub) String() string                     { return AsString(n) }
func (n *Select) String() string                    { return AsString(n) }
func (n *SelectClause) String() string              { return AsString(n) }
func (n *SetConstraints) String() string            { return AsString(n) }
func (n *SetRole) String() string                   { return AsString(n) }
func (n *SetSessionAuthorization) String() string   { return AsString(n) }
func (n *SetSessionCharacteristics) String() string { return AsString(n) }
func (n *SetTransaction) String() string            { return AsString(n) }
func (n *SetVar) String() string                    { return AsString(n) }
func (n *ShowBackup) String() string                { return AsString(n) }
func (n *ShowColumns) String() string               { return AsString(n) }
func (n *ShowConstraints) String() string           { return AsString(n) }
func (n *ShowCreate) String() string                { return AsString(n) }
func (n *ShowDatabases) String() string             { return AsString(n) }
func (n *ShowDatabaseIndexes) String() string       { return AsString(n) }
func (n *ShowEnums) String() string                 { return AsString(n) }
func (n *ShowGrants) String() string                { return AsString(n) }
func (n *ShowHistogram) String() string             { return AsString(n) }
func (n *ShowSchedules) String() string             { return AsString(n) }
func (n *ShowIndexes) String() string               { return AsString(n) }
func (n *ShowPartitions) String() string            { return AsString(n) }
func (n *ShowJobs) String() string                  { return AsString(n) }
func (n *ShowQueries) String() string               { return AsString(n) }
func (n *ShowRoleGrants) String() string            { return AsString(n) }
func (n *ShowRoles) String() string                 { return AsString(n) }
func (n *ShowSavepointStatus) String() string       { return AsString(n) }
func (n *ShowSchemas) String() string               { return AsString(n) }
func (n *ShowSequences) String() string             { return AsString(n) }
func (n *ShowSessions) String() string              { return AsString(n) }
func (n *ShowSyntax) String() string                { return AsString(n) }
func (n *ShowTableStats) String() string            { return AsString(n) }
func (n *ShowTables) String() string                { return AsString(n) }
func (n *ShowTypes) String() string                 { return AsString(n) }
func (n *ShowTraceForSession) String() string       { return AsString(n) }
func (n *ShowTransactionStatus) String() string     { return AsString(n) }
func (n *ShowTransactions) String() string          { return AsString(n) }
func (n *ShowLastQueryStatistics) String() string   { return AsString(n) }
func (n *ShowUsers) String() string                 { return AsString(n) }
func (n *ShowVar) String() string                   { return AsString(n) }
func (n *ShowFingerprints) String() string          { return AsString(n) }
func (n *Split) String() string                     { return AsString(n) }
func (n *Unsplit) String() string                   { return AsString(n) }
func (n *Truncate) String() string                  { return AsString(n) }
func (n *UnionClause) String() string               { return AsString(n) }
func (n *Update) String() string                    { return AsString(n) }
func (n *ValuesClause) String() string              { return AsString(n) }
