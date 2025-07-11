// Portions Copyright (c) 1996-2015, PostgreSQL Global Development Group
// Portions Copyright (c) 1994, Regents of the University of California
// Portions Copyright 2023 Dolthub, Inc.

// Portions of this file are additionally subject to the following
// license and copyright.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied. See the License for the specific language governing
// permissions and limitations under the License.

%{
package parser

import (
    "fmt"
    "strings"

    "go/constant"

    "github.com/dolthub/doltgresql/postgres/parser/geo/geopb"
    "github.com/dolthub/doltgresql/postgres/parser/roachpb"
    "github.com/dolthub/doltgresql/postgres/parser/lex"
    "github.com/dolthub/doltgresql/postgres/parser/privilege"
    "github.com/dolthub/doltgresql/postgres/parser/roleoption"
    "github.com/dolthub/doltgresql/postgres/parser/sem/tree"
    "github.com/dolthub/doltgresql/postgres/parser/types"
    "github.com/lib/pq/oid"
)

// MaxUint is the maximum value of an uint.
const MaxUint = ^uint(0)
// MaxInt is the maximum value of an int.
const MaxInt = int(MaxUint >> 1)

func unimplemented(sqllex sqlLexer, feature string) int {
    sqllex.(*lexer).Unimplemented(feature)
    return 1
}

func purposelyUnimplemented(sqllex sqlLexer, feature string, reason string) int {
    sqllex.(*lexer).PurposelyUnimplemented(feature, reason)
    return 1
}

func setErr(sqllex sqlLexer, err error) int {
    sqllex.(*lexer).setErr(err)
    return 1
}

func unimplementedWithIssue(sqllex sqlLexer, issue int) int {
    sqllex.(*lexer).UnimplementedWithIssue(issue)
    return 1
}

func unimplementedWithIssueDetail(sqllex sqlLexer, issue int, detail string) int {
    sqllex.(*lexer).UnimplementedWithIssueDetail(issue, detail)
    return 1
}
%}

%{
// sqlSymUnion represents a union of types, providing accessor methods
// to retrieve the underlying type stored in the union's empty interface.
// The purpose of the sqlSymUnion struct is to reduce the memory footprint of
// the sqlSymType because only one value (of a variety of types) is ever needed
// to be stored in the union field at a time.
//
// By using an empty interface, we lose the type checking previously provided
// by yacc and the Go compiler when dealing with union values. Instead, runtime
// type assertions must be relied upon in the methods below, and as such, the
// parser should be thoroughly tested whenever new syntax is added.
//
// It is important to note that when assigning values to sqlSymUnion.val, all
// nil values should be typed so that they are stored as nil instances in the
// empty interface, instead of setting the empty interface to nil. This means
// that:
//     $$ = []String(nil)
// should be used, instead of:
//     $$ = nil
// to assign a nil string slice to the union.
type sqlSymUnion struct {
    val interface{}
}

// The following accessor methods come in three forms, depending on the
// type of the value being accessed and whether a nil value is admissible
// for the corresponding grammar rule.
// - Values and pointers are directly type asserted from the empty
//   interface, regardless of whether a nil value is admissible or
//   not. A panic occurs if the type assertion is incorrect; no panic occurs
//   if a nil is not expected but present. (TODO(knz): split this category of
//   accessor in two; with one checking for unexpected nils.)
//   Examples: bool(), tableIndexName().
//
// - Interfaces where a nil is admissible are handled differently
//   because a nil instance of an interface inserted into the empty interface
//   becomes a nil instance of the empty interface and therefore will fail a
//   direct type assertion. Instead, a guarded type assertion must be used,
//   which returns nil if the type assertion fails.
//   Examples: expr(), stmt().
//
// - Interfaces where a nil is not admissible are implemented as a direct
//   type assertion, which causes a panic to occur if an unexpected nil
//   is encountered.
//   Examples: tblDef().
//
func (u *sqlSymUnion) numVal() *tree.NumVal {
    return u.val.(*tree.NumVal)
}
func (u *sqlSymUnion) strVal() *tree.StrVal {
    if stmt, ok := u.val.(*tree.StrVal); ok {
        return stmt
    }
    return nil
}
func (u *sqlSymUnion) placeholder() *tree.Placeholder {
    return u.val.(*tree.Placeholder)
}
func (u *sqlSymUnion) bool() bool {
    return u.val.(bool)
}
func (u *sqlSymUnion) strPtr() *string {
    return u.val.(*string)
}
func (u *sqlSymUnion) strs() []string {
    return u.val.([]string)
}
func (u *sqlSymUnion) newTableIndexName() *tree.TableIndexName {
    tn := u.val.(tree.TableIndexName)
    return &tn
}
func (u *sqlSymUnion) tableIndexName() tree.TableIndexName {
    return u.val.(tree.TableIndexName)
}
func (u *sqlSymUnion) newTableIndexNames() tree.TableIndexNames {
    return u.val.(tree.TableIndexNames)
}
func (u *sqlSymUnion) nameList() tree.NameList {
    return u.val.(tree.NameList)
}
func (u *sqlSymUnion) unresolvedName() *tree.UnresolvedName {
    return u.val.(*tree.UnresolvedName)
}
func (u *sqlSymUnion) unresolvedObjectName() *tree.UnresolvedObjectName {
    return u.val.(*tree.UnresolvedObjectName)
}
func (u *sqlSymUnion) unresolvedObjectNames() []*tree.UnresolvedObjectName {
    return u.val.([]*tree.UnresolvedObjectName)
}
func (u *sqlSymUnion) functionReference() tree.FunctionReference {
    return u.val.(tree.FunctionReference)
}
func (u *sqlSymUnion) tablePatterns() tree.TablePatterns {
    return u.val.(tree.TablePatterns)
}
func (u *sqlSymUnion) tableNames() tree.TableNames {
    return u.val.(tree.TableNames)
}
func (u *sqlSymUnion) indexFlags() *tree.IndexFlags {
    return u.val.(*tree.IndexFlags)
}
func (u *sqlSymUnion) arraySubscript() *tree.ArraySubscript {
    return u.val.(*tree.ArraySubscript)
}
func (u *sqlSymUnion) arraySubscripts() tree.ArraySubscripts {
    if as, ok := u.val.(tree.ArraySubscripts); ok {
        return as
    }
    return nil
}
func (u *sqlSymUnion) stmt() tree.Statement {
    if stmt, ok := u.val.(tree.Statement); ok {
        return stmt
    }
    return nil
}
func (u *sqlSymUnion) stmts() []tree.Statement {
    if stmt, ok := u.val.([]tree.Statement); ok {
        return stmt
    }
    return nil
}
func (u *sqlSymUnion) cte() *tree.CTE {
    if cte, ok := u.val.(*tree.CTE); ok {
        return cte
    }
    return nil
}
func (u *sqlSymUnion) ctes() []*tree.CTE {
    return u.val.([]*tree.CTE)
}
func (u *sqlSymUnion) with() *tree.With {
    if with, ok := u.val.(*tree.With); ok {
        return with
    }
    return nil
}
func (u *sqlSymUnion) slct() *tree.Select {
    return u.val.(*tree.Select)
}
func (u *sqlSymUnion) selectStmt() tree.SelectStatement {
    return u.val.(tree.SelectStatement)
}
func (u *sqlSymUnion) colDef() *tree.ColumnTableDef {
    return u.val.(*tree.ColumnTableDef)
}
func (u *sqlSymUnion) constraintDef() tree.ConstraintTableDef {
    return u.val.(tree.ConstraintTableDef)
}
func (u *sqlSymUnion) tblDef() tree.TableDef {
    return u.val.(tree.TableDef)
}
func (u *sqlSymUnion) tblDefs() tree.TableDefs {
    return u.val.(tree.TableDefs)
}
func (u *sqlSymUnion) likeTableOption() tree.LikeTableOption {
    return u.val.(tree.LikeTableOption)
}
func (u *sqlSymUnion) likeTableOptionList() []tree.LikeTableOption {
    return u.val.([]tree.LikeTableOption)
}
func (u *sqlSymUnion) colQual() tree.NamedColumnQualification {
    return u.val.(tree.NamedColumnQualification)
}
func (u *sqlSymUnion) colQualElem() tree.ColumnQualification {
    return u.val.(tree.ColumnQualification)
}
func (u *sqlSymUnion) colQuals() []tree.NamedColumnQualification {
    return u.val.([]tree.NamedColumnQualification)
}
func (u *sqlSymUnion) storageParam() tree.StorageParam {
    return u.val.(tree.StorageParam)
}
func (u *sqlSymUnion) storageParams() []tree.StorageParam {
    if params, ok := u.val.([]tree.StorageParam); ok {
        return params
    }
    return nil
}
func (u *sqlSymUnion) persistence() tree.Persistence {
 return u.val.(tree.Persistence)
}
func (u *sqlSymUnion) colType() *types.T {
    if colType, ok := u.val.(*types.T); ok && colType != nil {
        return colType
    }
    return nil
}
func (u *sqlSymUnion) tableRefCols() []tree.ColumnID {
    if refCols, ok := u.val.([]tree.ColumnID); ok {
        return refCols
    }
    return nil
}
func (u *sqlSymUnion) colTypes() []*types.T {
    return u.val.([]*types.T)
}
func (u *sqlSymUnion) int32() int32 {
    return u.val.(int32)
}
func (u *sqlSymUnion) int64() int64 {
    return u.val.(int64)
}
func (u *sqlSymUnion) seqOpt() tree.SequenceOption {
    return u.val.(tree.SequenceOption)
}
func (u *sqlSymUnion) seqOpts() []tree.SequenceOption {
    return u.val.([]tree.SequenceOption)
}
func (u *sqlSymUnion) expr() tree.Expr {
    if expr, ok := u.val.(tree.Expr); ok {
        return expr
    }
    return nil
}
func (u *sqlSymUnion) exprs() tree.Exprs {
    return u.val.(tree.Exprs)
}
func (u *sqlSymUnion) selExpr() tree.SelectExpr {
    return u.val.(tree.SelectExpr)
}
func (u *sqlSymUnion) selExprs() tree.SelectExprs {
    return u.val.(tree.SelectExprs)
}
func (u *sqlSymUnion) retClause() tree.ReturningClause {
        return u.val.(tree.ReturningClause)
}
func (u *sqlSymUnion) aliasClause() tree.AliasClause {
    return u.val.(tree.AliasClause)
}
func (u *sqlSymUnion) asOfClause() tree.AsOfClause {
    return u.val.(tree.AsOfClause)
}
func (u *sqlSymUnion) tblExpr() tree.TableExpr {
    return u.val.(tree.TableExpr)
}
func (u *sqlSymUnion) tblExprs() tree.TableExprs {
    return u.val.(tree.TableExprs)
}
func (u *sqlSymUnion) from() tree.From {
    return u.val.(tree.From)
}
func (u *sqlSymUnion) int32s() []int32 {
    return u.val.([]int32)
}
func (u *sqlSymUnion) joinCond() tree.JoinCond {
    return u.val.(tree.JoinCond)
}
func (u *sqlSymUnion) when() *tree.When {
    return u.val.(*tree.When)
}
func (u *sqlSymUnion) whens() []*tree.When {
    return u.val.([]*tree.When)
}
func (u *sqlSymUnion) lockingClause() tree.LockingClause {
    return u.val.(tree.LockingClause)
}
func (u *sqlSymUnion) lockingItem() *tree.LockingItem {
    return u.val.(*tree.LockingItem)
}
func (u *sqlSymUnion) lockingStrength() tree.LockingStrength {
    return u.val.(tree.LockingStrength)
}
func (u *sqlSymUnion) lockingWaitPolicy() tree.LockingWaitPolicy {
    return u.val.(tree.LockingWaitPolicy)
}
func (u *sqlSymUnion) updateExpr() *tree.UpdateExpr {
    return u.val.(*tree.UpdateExpr)
}
func (u *sqlSymUnion) updateExprs() tree.UpdateExprs {
    return u.val.(tree.UpdateExprs)
}
func (u *sqlSymUnion) limit() *tree.Limit {
    return u.val.(*tree.Limit)
}
func (u *sqlSymUnion) targetList() tree.TargetList {
    return u.val.(tree.TargetList)
}
func (u *sqlSymUnion) targetListPtr() *tree.TargetList {
    return u.val.(*tree.TargetList)
}
func (u *sqlSymUnion) privilegeType() privilege.Kind {
    return u.val.(privilege.Kind)
}
func (u *sqlSymUnion) privilegeList() privilege.List {
    return u.val.(privilege.List)
}
func (u *sqlSymUnion) onConflict() *tree.OnConflict {
    return u.val.(*tree.OnConflict)
}
func (u *sqlSymUnion) orderBy() tree.OrderBy {
    return u.val.(tree.OrderBy)
}
func (u *sqlSymUnion) order() *tree.Order {
    return u.val.(*tree.Order)
}
func (u *sqlSymUnion) orders() []*tree.Order {
    return u.val.([]*tree.Order)
}
func (u *sqlSymUnion) groupBy() tree.GroupBy {
    return u.val.(tree.GroupBy)
}
func (u *sqlSymUnion) windowFrame() *tree.WindowFrame {
    return u.val.(*tree.WindowFrame)
}
func (u *sqlSymUnion) windowFrameBounds() tree.WindowFrameBounds {
    return u.val.(tree.WindowFrameBounds)
}
func (u *sqlSymUnion) windowFrameBound() *tree.WindowFrameBound {
    return u.val.(*tree.WindowFrameBound)
}
func (u *sqlSymUnion) windowFrameExclusion() tree.WindowFrameExclusion {
    return u.val.(tree.WindowFrameExclusion)
}
func (u *sqlSymUnion) distinctOn() tree.DistinctOn {
    return u.val.(tree.DistinctOn)
}
func (u *sqlSymUnion) dir() tree.Direction {
    return u.val.(tree.Direction)
}
func (u *sqlSymUnion) nullsOrder() tree.NullsOrder {
    return u.val.(tree.NullsOrder)
}
func (u *sqlSymUnion) alterTableCmd() tree.AlterTableCmd {
    return u.val.(tree.AlterTableCmd)
}
func (u *sqlSymUnion) alterTableCmds() tree.AlterTableCmds {
    return u.val.(tree.AlterTableCmds)
}
func (u *sqlSymUnion) alterIndexCmd() tree.AlterIndexCmd {
    return u.val.(tree.AlterIndexCmd)
}
func (u *sqlSymUnion) isoLevel() tree.IsolationLevel {
    return u.val.(tree.IsolationLevel)
}
func (u *sqlSymUnion) userPriority() tree.UserPriority {
    return u.val.(tree.UserPriority)
}
func (u *sqlSymUnion) readWriteMode() tree.ReadWriteMode {
    return u.val.(tree.ReadWriteMode)
}
func (u *sqlSymUnion) deferrableMode() tree.DeferrableMode {
    return u.val.(tree.DeferrableMode)
}
func (u *sqlSymUnion) idxElem() tree.IndexElem {
    return u.val.(tree.IndexElem)
}
func (u *sqlSymUnion) idxElems() tree.IndexElemList {
    return u.val.(tree.IndexElemList)
}
func (u *sqlSymUnion) dropBehavior() tree.DropBehavior {
    return u.val.(tree.DropBehavior)
}
func (u *sqlSymUnion) validationBehavior() tree.ValidationBehavior {
    return u.val.(tree.ValidationBehavior)
}
func (u *sqlSymUnion) partitionBy() *tree.PartitionBy {
    return u.val.(*tree.PartitionBy)
}
func (u *sqlSymUnion) createTableOnCommitSetting() tree.CreateTableOnCommitSetting {
    return u.val.(tree.CreateTableOnCommitSetting)
}
func (u *sqlSymUnion) tuples() []*tree.Tuple {
    return u.val.([]*tree.Tuple)
}
func (u *sqlSymUnion) tuple() *tree.Tuple {
    return u.val.(*tree.Tuple)
}
func (u *sqlSymUnion) windowDef() *tree.WindowDef {
    return u.val.(*tree.WindowDef)
}
func (u *sqlSymUnion) window() tree.Window {
    return u.val.(tree.Window)
}
func (u *sqlSymUnion) op() tree.Operator {
    return u.val.(tree.Operator)
}
func (u *sqlSymUnion) cmpOp() tree.ComparisonOperator {
    return u.val.(tree.ComparisonOperator)
}
func (u *sqlSymUnion) intervalTypeMetadata() types.IntervalTypeMetadata {
    return u.val.(types.IntervalTypeMetadata)
}
func (u *sqlSymUnion) kvOption() tree.KVOption {
    return u.val.(tree.KVOption)
}
func (u *sqlSymUnion) kvOptions() []tree.KVOption {
    if colType, ok := u.val.([]tree.KVOption); ok {
        return colType
    }
    return nil
}
func (u *sqlSymUnion) backupOptions() *tree.BackupOptions {
  return u.val.(*tree.BackupOptions)
}
func (u *sqlSymUnion) copyOptions() *tree.CopyOptions {
  return u.val.(*tree.CopyOptions)
}
func (u *sqlSymUnion) restoreOptions() *tree.RestoreOptions {
  return u.val.(*tree.RestoreOptions)
}
func (u *sqlSymUnion) transactionModes() tree.TransactionModes {
    return u.val.(tree.TransactionModes)
}
func (u *sqlSymUnion) compositeKeyMatchMethod() tree.CompositeKeyMatchMethod {
  return u.val.(tree.CompositeKeyMatchMethod)
}
func (u *sqlSymUnion) refAction() tree.RefAction {
    return u.val.(tree.RefAction)
}
func (u *sqlSymUnion) referenceActions() tree.ReferenceActions {
    return u.val.(tree.ReferenceActions)
}
func (u *sqlSymUnion) createStatsOptions() *tree.CreateStatsOptions {
    return u.val.(*tree.CreateStatsOptions)
}
func (u *sqlSymUnion) scrubOptions() tree.ScrubOptions {
    return u.val.(tree.ScrubOptions)
}
func (u *sqlSymUnion) scrubOption() tree.ScrubOption {
    return u.val.(tree.ScrubOption)
}
func (u *sqlSymUnion) resolvableFuncRefFromName() tree.ResolvableFunctionReference {
    return tree.ResolvableFunctionReference{FunctionReference: u.unresolvedName()}
}
func (u *sqlSymUnion) rowsFromExpr() *tree.RowsFromExpr {
    return u.val.(*tree.RowsFromExpr)
}
func (u *sqlSymUnion) stringOrPlaceholderOptList() tree.StringOrPlaceholderOptList {
    return u.val.(tree.StringOrPlaceholderOptList)
}
func (u *sqlSymUnion) listOfStringOrPlaceholderOptList() []tree.StringOrPlaceholderOptList {
    return u.val.([]tree.StringOrPlaceholderOptList)
}
func (u *sqlSymUnion) fullBackupClause() *tree.FullBackupClause {
    return u.val.(*tree.FullBackupClause)
}
func (u *sqlSymUnion) geoShapeType() geopb.ShapeType {
  return u.val.(geopb.ShapeType)
}
func newNameFromStr(s string) *tree.Name {
    return (*tree.Name)(&s)
}
func (u *sqlSymUnion) typeReference() tree.ResolvableTypeReference {
    return u.val.(tree.ResolvableTypeReference)
}
func (u *sqlSymUnion) typeReferences() []tree.ResolvableTypeReference {
    return u.val.([]tree.ResolvableTypeReference)
}
func (u *sqlSymUnion) alterTypeAddValuePlacement() *tree.AlterTypeAddValuePlacement {
    return u.val.(*tree.AlterTypeAddValuePlacement)
}
func (u *sqlSymUnion) scheduleState() tree.ScheduleState {
  return u.val.(tree.ScheduleState)
}
func (u *sqlSymUnion) executorType() tree.ScheduledJobExecutorType {
  return u.val.(tree.ScheduledJobExecutorType)
}
func (u *sqlSymUnion) refreshDataOption() tree.RefreshDataOption {
  return u.val.(tree.RefreshDataOption)
}
func (u *sqlSymUnion) aggregateSignature() *tree.AggregateSignature {
  return u.val.(*tree.AggregateSignature)
}
func (u *sqlSymUnion) routineArg() *tree.RoutineArg {
  return u.val.(*tree.RoutineArg)
}
func (u *sqlSymUnion) routineArgs() []*tree.RoutineArg {
  return u.val.([]*tree.RoutineArg)
}
func (u *sqlSymUnion) databaseOption() tree.DatabaseOption {
    return u.val.(tree.DatabaseOption)
}
func (u *sqlSymUnion) databaseOptionList() []tree.DatabaseOption {
    return u.val.([]tree.DatabaseOption)
}
func (u *sqlSymUnion) setVar() *tree.SetVar {
    return u.val.(*tree.SetVar)
}
func (u *sqlSymUnion) routine() tree.Routine {
  return u.val.(tree.Routine)
}
func (u *sqlSymUnion) routines() []tree.Routine {
  return u.val.([]tree.Routine)
}
func (u *sqlSymUnion) privForCols() tree.PrivForCols {
  return u.val.(tree.PrivForCols)
}
func (u *sqlSymUnion) privForColsList() []tree.PrivForCols {
  return u.val.([]tree.PrivForCols)
}
func (u *sqlSymUnion) alterDefaultPrivileges() *tree.AlterDefaultPrivileges {
  return u.val.(*tree.AlterDefaultPrivileges)
}
func (u *sqlSymUnion) simpleColumnDef() tree.SimpleColumnDef {
  return u.val.(tree.SimpleColumnDef)
}
func (u *sqlSymUnion) simpleColumnDefs() []tree.SimpleColumnDef {
  return u.val.([]tree.SimpleColumnDef)
}
func (u *sqlSymUnion) routineOption() tree.RoutineOption {
  return u.val.(tree.RoutineOption)
}
func (u *sqlSymUnion) routineOptions() []tree.RoutineOption {
  return u.val.([]tree.RoutineOption)
}
func (u *sqlSymUnion) routineWithArgs() []tree.RoutineWithArgs {
  return u.val.([]tree.RoutineWithArgs)
}
func (u *sqlSymUnion) opClass() *tree.IndexElemOpClass {
  return u.val.(*tree.IndexElemOpClass)
}
func (u *sqlSymUnion) opClassOptions() []tree.IndexElemOpClassOption {
  return u.val.([]tree.IndexElemOpClassOption)
}
func (u *sqlSymUnion) initiallyMode() tree.InitiallyMode {
    return u.val.(tree.InitiallyMode)
}
func (u *sqlSymUnion) constraintIdxParams() tree.IndexParams {
    return u.val.(tree.IndexParams)
}
func (u *sqlSymUnion) partitionByType() tree.PartitionByType {
    return u.val.(tree.PartitionByType)
}
func (u *sqlSymUnion) partitionBoundSpec() tree.PartitionBoundSpec {
    return u.val.(tree.PartitionBoundSpec)
}
func (u *sqlSymUnion) alterColComputed() tree.AlterColComputed {
    return u.val.(tree.AlterColComputed)
}
func (u *sqlSymUnion) alterColComputedList() []tree.AlterColComputed {
    return u.val.([]tree.AlterColComputed)
}
func (u *sqlSymUnion) storageType() tree.StorageType {
    return u.val.(tree.StorageType)
}
func (u *sqlSymUnion) detachPartition() tree.DetachPartition {
    return u.val.(tree.DetachPartition)
}
func (u *sqlSymUnion) viewOption() tree.ViewOption {
    return u.val.(tree.ViewOption)
}
func (u *sqlSymUnion) viewOptions() tree.ViewOptions {
    return u.val.(tree.ViewOptions)
}
func (u *sqlSymUnion) viewCheckOption() tree.ViewCheckOption {
    return u.val.(tree.ViewCheckOption)
}
func (u *sqlSymUnion) alterViewCmd() tree.AlterViewCmd {
    return u.val.(tree.AlterViewCmd)
}
func (u *sqlSymUnion) triggerDeferrableMode() tree.TriggerDeferrableMode {
    return u.val.(tree.TriggerDeferrableMode)
}
func (u *sqlSymUnion) triggerRelations() tree.TriggerRelations {
    return u.val.(tree.TriggerRelations)
}
func (u *sqlSymUnion) triggerEvent() tree.TriggerEvent {
    return u.val.(tree.TriggerEvent)
}
func (u *sqlSymUnion) triggerEvents() tree.TriggerEvents {
    return u.val.(tree.TriggerEvents)
}
func (u *sqlSymUnion) triggerTime() tree.TriggerTime {
    return u.val.(tree.TriggerTime)
}
func (u *sqlSymUnion) languageHandler() *tree.LanguageHandler {
    return u.val.(*tree.LanguageHandler)
}
func (u *sqlSymUnion) compositeTypeElems() []tree.CompositeTypeElem {
    return u.val.([]tree.CompositeTypeElem)
}
func (u *sqlSymUnion) rangeTypeOption() tree.RangeTypeOption {
    return u.val.(tree.RangeTypeOption)
}
func (u *sqlSymUnion) rangeTypeOptions() []tree.RangeTypeOption {
    return u.val.([]tree.RangeTypeOption)
}
func (u *sqlSymUnion) baseTypeOption() tree.BaseTypeOption {
    return u.val.(tree.BaseTypeOption)
}
func (u *sqlSymUnion) baseTypeOptions() []tree.BaseTypeOption {
    return u.val.([]tree.BaseTypeOption)
}
func (u *sqlSymUnion) alterAttributeAction() tree.AlterAttributeAction {
    return u.val.(tree.AlterAttributeAction)
}
func (u *sqlSymUnion) alterAttributeActions() []tree.AlterAttributeAction {
    return u.val.([]tree.AlterAttributeAction)
}
func (u *sqlSymUnion) domainConstraint() tree.DomainConstraint {
    return u.val.(tree.DomainConstraint)
}
func (u *sqlSymUnion) domainConstraints() []tree.DomainConstraint {
    return u.val.([]tree.DomainConstraint)
}
func (u *sqlSymUnion) alterDomainCmd() tree.AlterDomainCmd {
    return u.val.(tree.AlterDomainCmd)
}
func (u *sqlSymUnion) createAggOption() tree.CreateAggOption {
    return u.val.(tree.CreateAggOption)
}
func (u *sqlSymUnion) createAggOptions() []tree.CreateAggOption {
    return u.val.([]tree.CreateAggOption)
}
func (u *sqlSymUnion) aggregatesToDrop() []tree.AggregateToDrop {
    return u.val.([]tree.AggregateToDrop)
}
func (u *sqlSymUnion) vacuumOptions() tree.VacuumOptions {
    return u.val.(tree.VacuumOptions)
}
func (u *sqlSymUnion) vacuumOption() *tree.VacuumOption {
    return u.val.(*tree.VacuumOption)
}
func (u *sqlSymUnion) vacuumTableAndCols() *tree.VacuumTableAndCols {
    return u.val.(*tree.VacuumTableAndCols)
}
func (u *sqlSymUnion) vacuumTableAndColsList() tree.VacuumTableAndColsList {
    return u.val.(tree.VacuumTableAndColsList)
}
%}

// NB: the %token definitions must come before the %type definitions in this
// file to work around a bug in goyacc. See #16369 for more details.

// Non-keyword token types.
%token <str> IDENT SCONST BCONST BITCONST
%token <*tree.NumVal> ICONST FCONST
%token <*tree.Placeholder> PLACEHOLDER
%token <str> TYPECAST TYPEANNOTATE DOT_DOT
%token <str> LESS_EQUALS GREATER_EQUALS NOT_EQUALS
%token <str> NOT_REGMATCH REGIMATCH NOT_REGIMATCH
%token <str> TEXTSEARCHMATCH
%token <str> ERROR

// If you want to make any keyword changes, add the new keyword here as well as
// to the appropriate one of the reserved-or-not-so-reserved keyword lists,
// below; search this file for "Keyword category lists".

// Ordinary key words in alphabetical order.
%token <str> ABORT ACCESS ACTION ADD ADMIN AFTER AGGREGATE
%token <str> ALIGNMENT ALL ALLOW_CONNECTIONS ALTER ALWAYS ANALYSE ANALYZE AND AND_AND ANY ANNOTATE_TYPE ARRAY AS ASC
%token <str> ASYMMETRIC AT ATOMIC ATTACH ATTRIBUTE AUTHORIZATION AUTO AUTOMATIC

%token <str> BACKUP BACKUPS BASETYPE BEFORE BEGIN BETWEEN BIGINT BIGSERIAL BINARY BIT
%token <str> FORMAT CSV HEADER
%token <str> BUCKET_COUNT 
%token <str> BOOLEAN BOTH BOX2D BUFFER_USAGE_LIMIT BUNDLE BY BYPASSRLS

%token <str> CACHE CHAIN CALL CALLED CANCEL CANCELQUERY CANONICAL CASCADE CASCADED CASE CAST CATEGORY CBRT
%token <str> CHANGEFEED CHAR CHARACTER CHARACTERISTICS CHECK CHECK_OPTION CLASS CLOSE
%token <str> CLUSTER COALESCE COLLATABLE COLLATE COLLATION COLLATION_VERSION COLUMN COLUMNS COMBINEFUNC COMMENT COMMENTS
%token <str> COMMIT COMMITTED COMPACT COMPLETE COMPRESSION CONCAT CONCURRENTLY CONFIGURATION CONFIGURATIONS CONFIGURE
%token <str> CONFLICT CONNECT CONNECTION CONSTRAINT CONSTRAINTS CONTAINS CONTROLCHANGEFEED
%token <str> CONTROLJOB CONVERSION CONVERT COPY COST CREATE CREATEDB CREATELOGIN CREATEROLE
%token <str> CROSS CUBE CURRENT CURRENT_CATALOG CURRENT_DATE CURRENT_SCHEMA
%token <str> CURRENT_ROLE CURRENT_TIME CURRENT_TIMESTAMP
%token <str> CURRENT_USER CYCLE

%token <str> DATA DATABASE DATABASES DATE DAY DEALLOCATE DEC DECIMAL DECLARE
%token <str> DEFAULT DEFAULTS DEFERRABLE DEFERRED DEFINER DELETE DELIMITER DEPENDS DESC DESCRIBE DESERIALFUNC DESTINATION
%token <str> DETACH DETACHED DICTIONARY DISABLE DISABLE_PAGE_SKIPPING DISCARD DISTINCT DO DOMAIN DOUBLE DROP

%token <str> EACH ELEMENT ELSE ENABLE ENCODING ENCRYPTION_PASSPHRASE ENCRYPTED END ENUM ENUMS ESCAPE EVENT
%token <str> EXCEPT EXCLUDE EXCLUDING EXISTS EXECUTE EXECUTION EXPERIMENTAL
%token <str> EXPERIMENTAL_FINGERPRINTS EXPERIMENTAL_REPLICA
%token <str> EXPERIMENTAL_AUDIT EXPIRATION EXPLAIN EXPORT EXPRESSION
%token <str> EXTENDED EXTENSION EXTERNAL EXTRACT EXTRACT_DURATION

%token <str> FALSE FAMILY FETCH FETCHVAL FETCHTEXT FETCHVAL_PATH FETCHTEXT_PATH
%token <str> FILES FILTER FINALFUNC FINALFUNC_EXTRA FINALFUNC_MODIFY FINALIZE FIRST FLOAT FLOAT4 FLOAT8 FLOORDIV
%token <str> FOLLOWING FOR FORCE FORCE_INDEX FOREIGN FREEZE FROM FULL FUNCTION FUNCTIONS

%token <str> GENERATED GEOGRAPHY GEOMETRY GEOMETRYM GEOMETRYZ GEOMETRYZM
%token <str> GEOMETRYCOLLECTION GEOMETRYCOLLECTIONM GEOMETRYCOLLECTIONZ GEOMETRYCOLLECTIONZM
%token <str> GLOBAL GRANT GRANTED GRANTS GREATEST GROUP GROUPING GROUPS

%token <str> HANDLER HASH HAVING HIGH HISTOGRAM HOUR HYPOTHETICAL

%token <str> ICU_LOCALE ICU_RULES IDENTITY
%token <str> IF IFERROR IFNULL IGNORE_FOREIGN_KEYS ILIKE IMMEDIATE IMMUTABLE IMPORT
%token <str> IN INCLUDE INCLUDING INCREMENT INCREMENTAL INET INET_CONTAINED_BY_OR_EQUALS
%token <str> INET_CONTAINS_OR_EQUALS INDEX INDEX_CLEANUP INDEXES INHERIT INHERITS INITCOND INJECT INLINE INPUT INTERLEAVE INITIALLY
%token <str> INNER INOUT INSERT INSTEAD INT INTEGER INTERNALLENGTH
%token <str> INTERSECT INTERVAL INTO INTO_DB INVERTED INVOKER IS ISERROR ISNULL ISOLATION IS_TEMPLATE

%token <str> JOB JOBS JOIN JSON JSONB JSON_SOME_EXISTS JSON_ALL_EXISTS

%token <str> KEY KEYS KMS KV

%token <str> LANGUAGE LARGE LAST LATERAL LATEST LC_CTYPE LC_COLLATE
%token <str> LEADING LEAKPROOF LEASE LEAST LEFT LESS LEVEL LIKE LIMIT
%token <str> LINESTRING LINESTRINGM LINESTRINGZ LINESTRINGZM LIST
%token <str> LOCAL LOCALE LOCALE_PROVIDER LOCALTIME LOCALTIMESTAMP LOCKED LOGGED LOGIN LOOKUP LOW LSHIFT

%token <str> MAIN MATCH MATERIALIZED MAXVALUE MERGE METHOD MFINALFUNC MFINALFUNC_EXTRA MFINALFUNC_MODIFY
%token <str> MINITCOND MINUTE MINVALUE MINVFUNC MODIFYCLUSTERSETTING MODULUS MONTH MSFUNC MSPACE MSSPACE MSTYPE
%token <str> MULTILINESTRING MULTILINESTRINGM MULTILINESTRINGZ MULTILINESTRINGZM MULTIPOINT MULTIPOINTM
%token <str> MULTIPOINTZ MULTIPOINTZM MULTIPOLYGON MULTIPOLYGONM MULTIPOLYGONZ MULTIPOLYGONZM MULTIRANGE_TYPE_NAME

%token <str> NAN NAME NAMES NATURAL NEVER NEW NEXT NO NOCANCELQUERY NOCONTROLCHANGEFEED NOCONTROLJOB
%token <str> NOBYPASSRLS NOCREATEDB NOCREATELOGIN NOCREATEROLE NOINHERIT NOLOGIN NOMODIFYCLUSTERSETTING NOREPLICATION NOSUPERUSER NO_INDEX_JOIN
%token <str> NONE NORMAL NOT NOTHING NOTNULL NOVIEWACTIVITY NOWAIT NULL NULLIF NULLS NUMERIC YES

%token <str> OBJECT OF OFF OFFSET OID OIDS OIDVECTOR OLD ON ONLY ONLY_DATABASE_STATS OPT OPTION OPTIONS OR
%token <str> ORDER ORDINALITY OTHERS OUT OUTER OUTPUT OVER OVERLAPS OVERLAY OWNED OWNER OPERATOR

%token <str> PARALLEL PARAMETER PARENT PARSER PARTIAL PARTITION PARTITIONS PASSEDBYVALUE PASSWORD PAUSE PAUSED PHYSICAL
%token <str> PLACING PLAIN PLAN PLANS POINT POINTM POINTZ POINTZM POLICY POLYGON POLYGONM POLYGONZ POLYGONZM
%token <str> POSITION PRECEDING PRECISION PREFERRED PREPARE PRESERVE PRIMARY PRIORITY PRIVILEGES
%token <str> PROCEDURAL PROCEDURE PROCEDURES PROCESS_MAIN PROCESS_TOAST PUBLIC PUBLICATION

%token <str> QUERIES QUERY

%token <str> RANGE RANGES READ READ_ONLY READ_WRITE REAL RECEIVE RECURSIVE RECURRING REF REFERENCES REFERENCING REFRESH
%token <str> REGCLASS REGPROC REGPROCEDURE REGNAMESPACE REGTYPE REINDEX RELEASE REMAINDER
%token <str> REMOVE_PATH RENAME REPEATABLE REPLACE REPLICA REPLICATION RESET RESTART RESTORE RESTRICT RESTRICTED RESUME
%token <str> RETRY RETURN RETURNING RETURNS REVISION_HISTORY REVOKE RIGHT
%token <str> ROLE ROLES ROUTINE ROUTINES ROLLBACK ROLLUP ROW ROWS RSHIFT RULE RUNNING

%token <str> SAFE SAVEPOINT SCATTER SCHEDULE SCHEDULES SCHEMA SCHEMAS SCRUB SEARCH SECOND SECURITY
%token <str> SECURITY_BARRIER SECURITY_INVOKER SEED SELECT SEND
%token <str> SERIALFUNC SERIALIZABLE SERVER SESSION SESSIONS SESSION_USER SET SETOF SETTING SETTINGS SEQUENCE SEQUENCES SFUNC
%token <str> SHARE SHAREABLE SHOW SIMILAR SIMPLE SKIP SKIP_LOCKED SKIP_DATABASE_STATS SKIP_MISSING_FOREIGN_KEYS
%token <str> SKIP_MISSING_SEQUENCES SKIP_MISSING_SEQUENCE_OWNERS SKIP_MISSING_VIEWS SMALLINT SMALLSERIAL SNAPSHOT SOME
%token <str> SORTOP SPLIT SQL SQRT SSPACE STABLE START STATEMENT STATISTICS STATUS STDIN STRATEGY STRICT STRING
%token <str> STORAGE STORE STORED STYPE SUBSCRIPT SUBSCRIPTION SUBSTRING SUBTYPE SUBTYPE_DIFF SUBTYPE_OPCLASS
%token <str> SUPERUSER SUPPORT SYMMETRIC SYNTAX SYSID SYSTEM

%token <str> TABLE TABLES TABLESPACE TEMP TEMPLATE TEMPORARY TEXT THEN
%token <str> TIES TIME TIMETZ TIMESTAMP TIMESTAMPTZ TO THROTTLING TRAILING TRACE TRACING
%token <str> TRANSACTION TRANSACTIONS TRANSFORM TREAT TRIGGER TRIM TRUE
%token <str> TRUNCATE TRUSTED TYPE TYPES TYPMOD_IN TYPMOD_OUT

%token <str> UNBOUNDED UNCOMMITTED UNION UNIQUE UNKNOWN UNLOGGED UNSAFE UNSPLIT
%token <str> UPDATE UPSERT UNTIL USAGE USE USER USERS USING UUID

%token <str> VACUUM VALID VALIDATE VALIDATOR VALUE VALUES VERBOSE
%token <str> VARBIT VARCHAR VARIABLE VARIADIC VARYING VERSION VIEW VIEWACTIVITY VIRTUAL VOLATILE

%token <str> WHEN WHERE WINDOW WITH WITHIN WITHOUT WORK WRAPPER WRITE

%token <str> XML

%token <str> YAML YEAR

%token <str> ZONE

// The grammar thinks these are keywords, but they are not in any category
// and so can never be entered directly. The filter in scan.go creates these
// tokens when required (based on looking one token ahead).
//
// NOT_LA exists so that productions such as NOT LIKE can be given the same
// precedence as LIKE; otherwise they'd effectively have the same precedence as
// NOT, at least with respect to their left-hand subexpression. WITH_LA is
// needed to make the grammar LALR(1). GENERATED_ALWAYS is needed to support
// the Postgres syntax for computed columns along with our family related
// extensions (CREATE FAMILY/CREATE FAMILY family_name).
%token NOT_LA WITH_LA AS_LA GENERATED_ALWAYS

%union {
  id    int32
  pos   int32
  str   string
  union sqlSymUnion
}

%type <tree.Statement> stmt_block
%type <tree.Statement> stmt
%type <tree.Statement> non_transaction_stmt

%type <tree.Statement> alter_stmt
%type <tree.Statement> alter_ddl_stmt
%type <tree.Statement> alter_table_stmt
%type <tree.Statement> alter_trigger_stmt
%type <tree.Statement> alter_index_stmt
%type <tree.Statement> alter_materialized_view_stmt
%type <tree.Statement> alter_function_stmt
%type <tree.Statement> alter_language_stmt
%type <tree.Statement> alter_procedure_stmt
%type <tree.Statement> alter_view_stmt
%type <tree.Statement> alter_sequence_stmt
%type <tree.Statement> alter_database_stmt
%type <tree.Statement> alter_role_stmt
%type <tree.Statement> alter_type_stmt
%type <tree.Statement> alter_schema_stmt
%type <tree.Statement> alter_domain_stmt

// ALTER TABLE
%type <tree.Statement> alter_onetable_stmt
%type <tree.Statement> alter_rename_table_stmt
%type <tree.Statement> alter_table_set_schema_stmt
%type <tree.Statement> alter_table_all_in_tablespace_stmt
%type <tree.Statement> alter_table_parition_stmt

// ALTER DATABASE
%type <tree.Statement> alter_rename_database_stmt
%type <tree.Statement> alter_database_to_schema_stmt
%type <tree.Statement> opt_alter_database

// ALTER DEFAULT PRIVILEGES
%type <tree.Statement> alter_default_privileges_stmt adp_abbreviated_grant_or_revoke

// ALTER INDEX
%type <tree.Statement> alter_oneindex_stmt
%type <tree.Statement> alter_rename_index_stmt
%type <tree.Statement> alter_index_all_in_tablespace_stmt

// ALTER MATERIALIZED VIEW
%type <tree.Statement> alter_materialized_view_rename_stmt
%type <tree.Statement> alter_materialized_view_set_schema_stmt
%type <tree.Statement> alter_materialized_view_all_in_tablespace_stmt

// ALTER SEQUENCE
%type <tree.Statement> alter_rename_sequence_stmt
%type <tree.Statement> alter_sequence_options_stmt
%type <tree.Statement> alter_sequence_set_schema_stmt
%type <tree.Statement> alter_sequence_set_log_stmt
%type <tree.Statement> alter_sequence_owner_to_stmt

%type <tree.Statement> alter_aggregate_stmt
%type <tree.Statement> alter_collation_stmt
%type <tree.Statement> alter_conversion_stmt

%type <tree.Statement> backup_stmt
%type <tree.Statement> begin_stmt

%type <tree.Statement> call_stmt

%type <tree.Statement> cancel_stmt
%type <tree.Statement> cancel_jobs_stmt
%type <tree.Statement> cancel_queries_stmt
%type <tree.Statement> cancel_sessions_stmt

// SCRUB
%type <tree.Statement> scrub_stmt
%type <tree.Statement> scrub_database_stmt
%type <tree.Statement> scrub_table_stmt
%type <tree.ScrubOptions> opt_scrub_options_clause
%type <tree.ScrubOptions> scrub_option_list
%type <tree.ScrubOption> scrub_option

%type <tree.Statement> comment_stmt
%type <tree.Statement> commit_stmt
%type <tree.Statement> copy_from_stmt

%type <tree.Statement> create_stmt
%type <tree.Statement> create_changefeed_stmt
%type <tree.Statement> create_ddl_stmt
%type <tree.Statement> create_ddl_stmt_schema_element
%type <tree.Statement> create_database_stmt
%type <tree.Statement> create_index_stmt
%type <tree.Statement> create_role_stmt
%type <tree.Statement> create_schedule_for_backup_stmt
%type <tree.Statement> create_extension_stmt
%type <tree.Statement> create_function_stmt
%type <tree.Statement> create_language_stmt
%type <tree.Statement> create_procedure_stmt
%type <tree.Statement> create_schema_stmt
%type <tree.Statement> create_table_stmt
%type <tree.Statement> create_table_as_stmt
%type <tree.Statement> create_view_stmt
%type <tree.Statement> create_materialized_view_stmt
%type <tree.Statement> create_sequence_stmt
%type <tree.Statement> create_trigger_stmt
%type <tree.Statement> create_domain_stmt
%type <tree.Statement> create_aggregate_stmt

%type <tree.Statement> create_aggregate_args_only_stmt
%type <tree.Statement> create_aggregate_order_by_args_stmt
%type <tree.Statement> create_aggregate_old_syntax_stmt

%type <tree.Statement> create_stats_stmt
%type <*tree.CreateStatsOptions> opt_create_stats_options
%type <*tree.CreateStatsOptions> create_stats_option_list
%type <*tree.CreateStatsOptions> create_stats_option

%type <tree.Statement> create_type_stmt
%type <tree.Statement> delete_stmt
%type <tree.Statement> discard_stmt

%type <tree.Statement> drop_stmt
%type <tree.Statement> drop_ddl_stmt
%type <tree.Statement> drop_database_stmt
%type <tree.Statement> drop_index_stmt
%type <tree.Statement> drop_role_stmt
%type <tree.Statement> drop_schema_stmt
%type <tree.Statement> drop_table_stmt
%type <tree.Statement> drop_trigger_stmt
%type <tree.Statement> drop_type_stmt
%type <tree.Statement> drop_view_stmt
%type <tree.Statement> drop_domain_stmt
%type <tree.Statement> drop_sequence_stmt
%type <tree.Statement> drop_extension_stmt
%type <tree.Statement> drop_language_stmt
%type <tree.Statement> drop_function_stmt
%type <tree.Statement> drop_procedure_stmt
%type <tree.Statement> drop_aggregate_stmt

%type <tree.Statement> analyze_stmt
%type <tree.Statement> explain_stmt
%type <tree.Statement> describe_table_stmt
%type <tree.Statement> prepare_stmt
%type <tree.Statement> preparable_stmt
%type <tree.Statement> row_source_extension_stmt
%type <tree.Statement> export_stmt
%type <tree.Statement> execute_stmt
%type <tree.Statement> deallocate_stmt
%type <tree.Statement> grant_stmt
%type <tree.Statement> insert_stmt
%type <tree.Statement> import_stmt
%type <tree.Statement> pause_stmt pause_jobs_stmt pause_schedules_stmt
%type <*tree.Select>   for_schedules_clause
%type <tree.Statement> release_stmt
%type <tree.Statement> reset_stmt
%type <tree.Statement> resume_stmt resume_jobs_stmt resume_schedules_stmt
%type <tree.Statement> drop_schedule_stmt
%type <tree.Statement> restore_stmt
%type <tree.StringOrPlaceholderOptList> string_or_placeholder_opt_list
%type <[]tree.StringOrPlaceholderOptList> list_of_string_or_placeholder_opt_list
%type <tree.Statement> revoke_stmt
%type <tree.Statement> refresh_stmt
%type <*tree.Select> select_stmt
%type <tree.Statement> abort_stmt
%type <tree.Statement> rollback_stmt
%type <tree.Statement> savepoint_stmt

%type <tree.Statement> schema_element
%type <[]tree.Statement> schema_element_list opt_schema_element_list stmt_list
%type <tree.Statement> set_stmt
%type <tree.Statement> set_session_or_local_stmt
%type <tree.Statement> set_transaction_stmt
%type <tree.Statement> set_constraints_stmt
%type <tree.Statement> set_exprs_internal
%type <tree.Statement> generic_set_single_config
%type <tree.Statement> set_session_or_local_cmd
%type <tree.Statement> set_session_authorization
%type <tree.Statement> set_var
%type <tree.Statement> set_special_syntax
%type <tree.Statement> set_names
%type <tree.Statement> set_role
%type <tree.Statement> begin_end_block
%type <tree.Statement> sql_body

%type <tree.Statement> show_stmt
%type <tree.Statement> show_backup_stmt
%type <tree.Statement> show_columns_stmt
%type <tree.Statement> show_constraints_stmt
%type <tree.Statement> show_create_stmt
%type <tree.Statement> show_databases_stmt
%type <tree.Statement> show_enums_stmt
%type <tree.Statement> show_fingerprints_stmt
%type <tree.Statement> show_grants_stmt
%type <tree.Statement> show_histogram_stmt
%type <tree.Statement> show_indexes_stmt
%type <tree.Statement> show_partitions_stmt
%type <tree.Statement> show_jobs_stmt
%type <tree.Statement> show_queries_stmt
%type <tree.Statement> show_roles_stmt
%type <tree.Statement> show_schemas_stmt
%type <tree.Statement> show_sequences_stmt
%type <tree.Statement> show_session_stmt
%type <tree.Statement> show_sessions_stmt
%type <tree.Statement> show_savepoint_stmt
%type <tree.Statement> show_stats_stmt
%type <tree.Statement> show_syntax_stmt
%type <tree.Statement> show_last_query_stats_stmt
%type <tree.Statement> show_tables_stmt
%type <tree.Statement> show_trace_stmt
%type <tree.Statement> show_transaction_stmt
%type <tree.Statement> show_transactions_stmt
%type <tree.Statement> show_types_stmt
%type <tree.Statement> show_users_stmt
%type <tree.Statement> show_schedules_stmt

%type <str> session_var
%type <*string> comment_text

%type <tree.Statement> transaction_stmt
%type <tree.Statement> truncate_stmt
%type <tree.Statement> update_stmt
%type <tree.Statement> upsert_stmt
%type <tree.Statement> use_stmt

%type <tree.Statement> close_cursor_stmt
%type <tree.Statement> declare_cursor_stmt
%type <tree.Statement> reindex_stmt

%type <tree.Statement> vacuum_stmt
%type <*tree.VacuumOption> vacuum_option legacy_vacuum_option
%type <tree.VacuumOptions> opt_vacuum_option_list vacuum_option_list legacy_vacuum_option_list
%type <*tree.VacuumTableAndCols> vacuum_table_and_cols 
%type <tree.VacuumTableAndColsList> opt_vacuum_table_and_cols_list
%type <str> auto_on_off

%type <[]string> opt_incremental
%type <tree.KVOption> kv_option
%type <[]tree.KVOption> kv_option_list opt_with_options opt_with_schedule_options
%type <*tree.BackupOptions> opt_with_backup_options backup_options backup_options_list
%type <*tree.RestoreOptions> opt_with_restore_options restore_options restore_options_list
%type <*tree.CopyOptions> copy_options copy_options_list
%type <*tree.CopyOptions> opt_legacy_copy_options legacy_copy_options legacy_copy_options_list
%type <str> import_format
%type <tree.StorageParam> storage_parameter
%type <[]tree.StorageParam> storage_parameter_list opt_table_with opt_with_storage_parameter_list attribution_list

%type <*tree.Select> select_no_parens
%type <tree.SelectStatement> select_clause select_with_parens simple_select empty_select values_clause table_clause simple_select_clause
%type <tree.LockingClause> for_locking_clause opt_for_locking_clause for_locking_items
%type <*tree.LockingItem> for_locking_item
%type <tree.LockingStrength> for_locking_strength
%type <tree.LockingWaitPolicy> opt_nowait_or_skip
%type <tree.SelectStatement> set_operation

%type <tree.Expr> alter_column_default opt_arg_default
%type <tree.Direction> opt_asc_desc
%type <tree.NullsOrder> opt_nulls_order
%type <[]tree.RoutineWithArgs> function_name_with_args_list

%type <*tree.AggregateSignature> aggregate_signature
%type <*tree.RoutineArg> routine_arg routine_arg_with_default
%type <[]*tree.RoutineArg> routine_arg_list opt_routine_args opt_routine_args_with_paren
%type <[]*tree.RoutineArg> routine_arg_with_default_list opt_routine_arg_with_default_list
%type <tree.SimpleColumnDef> returns_table_col_def
%type <[]tree.SimpleColumnDef> opt_returns_table_col_def_list
%type <tree.RoutineOption> function_option create_function_option alter_function_option create_procedure_option alter_procedure_option
%type <[]tree.RoutineOption> create_function_option_list alter_function_option_list create_procedure_option_list alter_procedure_option_list

%type <tree.Routine> routine_with_args
%type <[]tree.Routine> routine_with_args_list

%type <tree.CreateAggOption> create_agg_args_only_option create_agg_order_by_args_option
%type <tree.CreateAggOption> create_agg_old_syntax_option create_agg_common_option create_agg_parallel_option
%type <[]tree.CreateAggOption> create_agg_args_only_option_list create_agg_order_by_args_option_list create_agg_old_syntax_option_list
%type <[]tree.AggregateToDrop> drop_aggregates

%type <tree.DatabaseOption> opt_database_options
%type <[]tree.DatabaseOption> opt_database_options_list opt_database_with_options

%type <tree.AlterTableCmd> alter_table_action enable_or_disable_trigger enable_or_disable_rule
%type <tree.AlterTableCmd> alter_opt_column_options alter_materialized_view_opt_column_options
%type <tree.AlterTableCmd> row_level_security replica_identity_option alter_materialized_view_action
%type <tree.AlterTableCmds> alter_table_actions alter_table_cmd alter_materialized_view_cmd alter_materialized_view_actions
%type <tree.AlterColComputed> alter_column_set_seq_elem
%type <[]tree.AlterColComputed> alter_column_set_seq_elem_list
%type <tree.AlterIndexCmd> alter_index_cmd
%type <tree.StorageType> col_storage_option
%type <bool> unique_or_primary logged_or_unlogged opt_nowait opt_no opt_view_recursive
%type <str> trigger_name trigger_option
%type <tree.AlterViewCmd> alter_view_cmd

%type <tree.DropBehavior> opt_drop_behavior
%type <tree.ValidationBehavior> opt_validate_behavior

%type <str> opt_owner opt_template opt_encoding opt_strategy opt_locale opt_lc_collate opt_lc_ctype opt_icu_locale
%type <str> opt_icu_rules opt_locale_provider opt_collation_version opt_tablespace opt_using_index_tablespace

%type <tree.IsolationLevel> transaction_iso_level
%type <tree.UserPriority> transaction_user_priority
%type <tree.ReadWriteMode> transaction_read_mode
%type <tree.DeferrableMode> deferrable_mode opt_deferrable_mode

%type <str> name opt_name opt_name_parens use_db_name
%type <str> privilege savepoint_name
%type <tree.KVOption> role_option password_clause valid_until_clause
%type <tree.Operator> subquery_op
%type <*tree.UnresolvedName> func_name func_name_no_crdb_extra
%type <str> opt_compression 

%type <[]tree.CompositeTypeElem> type_composite_list opt_type_composite_list
%type <tree.RangeTypeOption> type_range_option
%type <[]tree.RangeTypeOption> type_range_optional_list
%type <tree.BaseTypeOption> type_base_option type_property
%type <[]tree.BaseTypeOption> type_base_optional_list type_property_list
%type <tree.AlterAttributeAction> alter_attribute_action
%type <[]tree.AlterAttributeAction> alter_attribute_action_list

%type <tree.DomainConstraint> domain_constraint
%type <[]tree.DomainConstraint> domain_constraint_list opt_domain_constraint_list
%type <tree.AlterDomainCmd> alter_domain_cmd

%type <str> cursor_name database_name index_name opt_index_name column_name insert_column_item statistics_name window_name
%type <str> table_alias_name constraint_name target_name opt_from_ref_table
%type <*tree.UnresolvedObjectName> collation_name 
%type <str> db_object_name_component
%type <*tree.UnresolvedObjectName> table_name standalone_index_name sequence_name type_name routine_name aggregate_name
%type <*tree.UnresolvedObjectName> view_name db_object_name simple_db_object_name complex_db_object_name  opt_collate
%type <*tree.UnresolvedObjectName> db_object_name_no_keywords simple_db_object_name_no_keywords complex_db_object_name_no_keywords
%type <[]*tree.UnresolvedObjectName> type_name_list
%type <str> schema_name opt_schema_name opt_schema opt_version tablespace_name partition_name
%type <[]string> schema_name_list role_spec_list opt_role_list opt_owned_by_list
%type <*tree.UnresolvedName> table_pattern complex_table_pattern
%type <*tree.UnresolvedName> column_path prefixed_column_path column_path_with_star
%type <tree.TableExpr> insert_target create_stats_target analyze_target

%type <*tree.UnresolvedObjectName> opt_handler_inline opt_handler_validator
%type <*tree.LanguageHandler> opt_language_handler

%type <tree.ViewOptions> view_options opt_with_view_options
%type <tree.ViewOption> view_option
%type <tree.ViewCheckOption> opt_with_check_option

%type <tree.Expr> opt_when
%type <bool> opt_for_each old_or_new opt_constraint opt_not_valid
%type <tree.TriggerDeferrableMode> opt_trigger_deferrable_mode
%type <tree.TriggerRelations> trigger_relations opt_trigger_relations
%type <tree.TriggerEvent> trigger_event
%type <tree.TriggerEvents> trigger_events
%type <tree.TriggerTime> trigger_time

%type <*tree.TableIndexName> table_index_name
%type <tree.TableIndexNames> table_index_name_list

%type <tree.Operator> math_op operator

%type <tree.IsolationLevel> iso_level
%type <tree.UserPriority> user_priority

%type <tree.TableDefs> opt_table_elem_list table_elem_list opt_table_of_elem_list table_of_elem_list
%type <[]tree.LikeTableOption> like_table_option_list
%type <tree.LikeTableOption> like_table_option
%type <tree.CreateTableOnCommitSetting> opt_create_table_on_commit
%type <*tree.PartitionBy> opt_partition_by partition_by
%type <tree.PartitionByType> partition_by_type
%type <empty> opt_all_clause
%type <str> explain_verb
%type <bool> distinct_clause opt_external definer_or_invoker opt_not opt_col_with_options
%type <tree.DistinctOn> distinct_on_clause
%type <tree.NameList> opt_column_list insert_column_list opt_stats_columns opt_of_cols
%type <tree.OrderBy> sort_clause single_sort_clause opt_sort_clause
%type <[]*tree.Order> sortby_list
%type <tree.IndexParams> constraint_index_params
%type <tree.IndexElemList> index_params index_params_name_only opt_index_params_name_only opt_include_index_cols partition_index_params exclude_elems
%type <tree.NameList> name_list opt_name_list privilege_list
%type <[]int32> opt_array_bounds
%type <tree.From> from_clause
%type <tree.TableExprs> from_list opt_from_list
%type <tree.TablePatterns> table_pattern_list single_table_pattern_list
%type <tree.TableNames> table_name_list opt_locked_rels opt_inherits
%type <tree.Exprs> expr_list opt_expr_list tuple1_ambiguous_values tuple1_unambiguous_values
%type <*tree.Tuple> expr_tuple1_ambiguous expr_tuple_unambiguous
%type <tree.SelectExprs> target_list
%type <tree.UpdateExprs> set_clause_list
%type <*tree.UpdateExpr> set_clause multiple_set_clause
%type <tree.ArraySubscripts> array_subscripts
%type <tree.GroupBy> group_clause
%type <*tree.Limit> select_limit opt_select_limit
%type <tree.TableNames> relation_expr_list
%type <tree.ReturningClause> returning_clause
%type <empty> opt_using_clause
%type <tree.RefreshDataOption> opt_clear_data

%type <[]tree.SequenceOption> create_seq_option_list opt_create_seq_option_list opt_create_seq_option_list_with_parens
%type <[]tree.SequenceOption> alter_seq_option_list opt_alter_seq_option_list
%type <tree.SequenceOption> create_seq_option_elem alter_seq_option_elem seq_as_type seq_increment
%type <tree.SequenceOption> seq_minvalue seq_maxvalue seq_start seq_cache seq_cycle seq_owned_by seq_restart

%type <bool> all_or_distinct opt_cascade opt_if_exists opt_restrict opt_trusted opt_procedural
%type <bool> with_comment opt_with_force opt_create_as_with_data
%type <bool> boolean_value boolean_value_for_vacuum_opt
%type <empty> join_outer
%type <tree.JoinCond> join_qual
%type <str> join_type
%type <str> opt_join_hint

%type <tree.Exprs> extract_list
%type <tree.Exprs> overlay_list
%type <tree.Exprs> position_list
%type <tree.Exprs> substr_list
%type <tree.Exprs> trim_list
%type <tree.Exprs> execute_param_clause
%type <types.IntervalTypeMetadata> opt_interval_qualifier interval_qualifier interval_second
%type <tree.Expr> overlay_placing
%type <tree.Exprs> var_list

%type <bool> opt_unique opt_concurrently opt_cluster
%type <str> opt_using_method

%type <*tree.Limit> limit_clause offset_clause opt_limit_clause
%type <tree.Expr> select_fetch_first_value
%type <empty> row_or_rows
%type <empty> first_or_next

%type <tree.Statement> insert_rest
%type <tree.NameList> opt_col_def_list
%type <*tree.OnConflict> on_conflict

%type <tree.Statement> begin_transaction
%type <tree.TransactionModes> transaction_mode_list transaction_mode

%type <bool> opt_only opt_nulls_distinct
%type <*tree.IndexElemOpClass> opt_opclass
%type <[]tree.IndexElemOpClassOption> opclass_option_list
%type <*tree.ColumnTableDef> alter_column_def create_table_column_def
%type <tree.TableDef> table_elem table_of_elem
%type <tree.Expr> where_clause opt_where_clause where_clause_paren opt_where_clause_paren
%type <*tree.ArraySubscript> array_subscript
%type <tree.Expr> opt_slice_bound
%type <*tree.IndexFlags> opt_index_flags
%type <*tree.IndexFlags> index_flags_param
%type <*tree.IndexFlags> index_flags_param_list
%type <tree.Expr> a_expr b_expr c_expr d_expr typed_literal
%type <tree.Expr> substr_from substr_for
%type <tree.Expr> in_expr partition_bound_expr
%type <tree.Expr> having_clause
%type <tree.Expr> array_expr
%type <tree.Expr> interval_value
%type <[]tree.ResolvableTypeReference> type_list prep_type_clause for_type_list
%type <tree.Exprs> array_expr_list int_expr_list
%type <*tree.Tuple> row labeled_row
%type <tree.Expr> case_expr case_arg case_default
%type <*tree.When> when_clause
%type <[]*tree.When> when_clause_list
%type <tree.ComparisonOperator> sub_type
%type <tree.Expr> numeric_only opt_allow_connections opt_connection_limit opt_is_template opt_oid
%type <tree.AliasClause> alias_clause opt_alias_clause
%type <bool> opt_ordinality opt_compact
%type <*tree.Order> sortby
%type <tree.IndexElem> index_elem index_elem_name_only partition_index_elem
%type <tree.TableExpr> table_ref numeric_table_ref func_table table_ref_options
%type <tree.Exprs> rowsfrom_list
%type <tree.Expr> rowsfrom_item
%type <tree.TableExpr> joined_table
%type <*tree.UnresolvedObjectName> relation_expr
%type <tree.TableExpr> table_expr_opt_alias_idx table_name_opt_idx
%type <tree.SelectExpr> target_elem
%type <*tree.UpdateExpr> single_set_clause
%type <tree.AsOfClause> as_of_clause opt_as_of_clause
%type <tree.Expr> opt_changefeed_sink

%type <str> explain_option_name explain_option_value
%type <[]string> explain_option_list opt_enum_val_list enum_val_list

%type <tree.ResolvableTypeReference> typename simple_typename cast_target
%type <*types.T> const_typename
%type <*tree.AlterTypeAddValuePlacement> opt_add_val_placement
%type <bool> opt_timezone
%type <*types.T> numeric opt_numeric_modifiers
%type <*types.T> opt_float
%type <*types.T> character_with_length character_without_length
%type <*types.T> const_datetime interval_type
%type <*types.T> bit_with_length bit_without_length
%type <*types.T> character_base
%type <*types.T> geo_shape_type
%type <*types.T> const_geo
%type <str> extract_arg
%type <bool> opt_varying opt_no_inherit

%type <*tree.NumVal> signed_iconst only_signed_iconst
%type <*tree.NumVal> signed_fconst only_signed_fconst
%type <int32> signed_iconst32
%type <int32> iconst32
%type <int64> signed_iconst64
%type <int64> iconst64
%type <tree.Expr> var_value opt_var_value opt_restart
%type <str> unrestricted_name type_function_name type_function_name_no_crdb_extra simple_ident
%type <str> non_reserved_word
%type <str> non_reserved_word_or_sconst
%type <str> role_spec owner_to set_schema opt_role set_tablespace
%type <tree.Expr> zone_value
%type <tree.Expr> string_or_placeholder
%type <tree.Expr> string_or_placeholder_list

%type <str> unreserved_keyword type_func_name_keyword type_func_name_no_crdb_extra_keyword
%type <str> col_name_keyword reserved_keyword cockroachdb_extra_reserved_keyword extra_var_value

%type <tree.ResolvableTypeReference> complex_type_name
%type <str> general_type_name

%type <tree.ConstraintTableDef> table_constraint table_constraint_elem
%type <[]tree.NamedColumnQualification> col_constraint_list
%type <tree.NamedColumnQualification> col_qualification
%type <tree.ColumnQualification> col_qualification_elem col_qual_generated_identity
%type <tree.CompositeKeyMatchMethod> key_match
%type <tree.ReferenceActions> reference_actions
%type <tree.RefAction> reference_action reference_on_delete reference_on_update
%type <tree.InitiallyMode> opt_initially

%type <tree.Expr> func_application func_expr_common_subexpr special_function
%type <tree.Expr> func_expr func_expr_windowless
%type <empty> opt_with
%type <*tree.With> with_clause opt_with_clause
%type <[]*tree.CTE> cte_list
%type <*tree.CTE> common_table_expr
%type <bool> materialize_clause

%type <tree.DetachPartition> detach_partition_type
%type <tree.PartitionBoundSpec> partition_of partition_bound_spec
%type <tree.Expr> within_group_clause
%type <tree.Expr> filter_clause
%type <tree.Exprs> opt_partition_clause partition_bound_expr_list
%type <tree.Window> window_clause window_definition_list
%type <*tree.WindowDef> window_definition over_clause window_specification
%type <str> opt_existing_window_name
%type <*tree.WindowFrame> opt_frame_clause
%type <tree.WindowFrameBounds> frame_extent
%type <*tree.WindowFrameBound> frame_bound
%type <tree.WindowFrameExclusion> opt_frame_exclusion

%type <[]tree.ColumnID> opt_tableref_col_list tableref_col_list

%type <tree.TargetList> targets_table targets_roles changefeed_targets other_targets targets targets_for_alter_def_priv
%type <*tree.TargetList> opt_on_targets_roles opt_backup_targets
%type <tree.NameList> for_grantee_clause
%type <privilege.List> privileges
%type <tree.PrivForCols> privilege_for_cols
%type <[]tree.PrivForCols> privilege_for_cols_list privileges_for_cols
%type <[]tree.KVOption> opt_role_options role_options
%type <str> opt_grant_role_with admin_inherit_set option_true_false opt_granted_by

%type <tree.Expr> opt_alter_column_using

%type <tree.Persistence> opt_temp opt_persistence_temp_table opt_persistence_sequence
%type <bool> role_or_group_or_user role_or_user opt_with_grant_option opt_grant_option_for

%type <tree.Expr>  cron_expr opt_description sconst_or_placeholder
%type <*tree.FullBackupClause> opt_full_backup_clause
%type <tree.ScheduleState> schedule_state
%type <tree.ScheduledJobExecutorType> opt_schedule_executor_type

// Precedence: lowest to highest
%nonassoc  VALUES              // see values_clause
%nonassoc  SET                 // see table_expr_opt_alias_idx
%left      UNION EXCEPT
%left      INTERSECT
%left      OR
%left      AND
%right     NOT
%nonassoc  IS ISNULL NOTNULL   // IS sets precedence for IS NULL, etc
%nonassoc  '<' '>' '=' LESS_EQUALS GREATER_EQUALS NOT_EQUALS
%nonassoc  '~' BETWEEN DEFERRABLE IN LIKE ILIKE SIMILAR NOT_REGMATCH REGIMATCH NOT_REGIMATCH NOT_LA TEXTSEARCHMATCH
%nonassoc  ESCAPE              // ESCAPE must be just above LIKE/ILIKE/SIMILAR
%nonassoc  CONTAINS CONTAINED_BY '?' JSON_SOME_EXISTS JSON_ALL_EXISTS
%nonassoc  OVERLAPS
%left      POSTFIXOP           // dummy for postfix OP rules
// To support target_elem without AS, we must give IDENT an explicit priority
// between POSTFIXOP and OP. We can safely assign the same priority to various
// unreserved keywords as needed to resolve ambiguities (this can't have any
// bad effects since obviously the keywords will still behave the same as if
// they weren't keywords). We need to do this for PARTITION, RANGE, ROWS,
// GROUPS to support opt_existing_window_name; and for RANGE, ROWS, GROUPS so
// that they can follow a_expr without creating postfix-operator problems; and
// for NULL so that it can follow b_expr in col_constraint_list without creating
// postfix-operator problems.
//
// To support CUBE and ROLLUP in GROUP BY without reserving them, we give them
// an explicit priority lower than '(', so that a rule with CUBE '(' will shift
// rather than reducing a conflicting rule that takes CUBE as a function name.
// Using the same precedence as IDENT seems right for the reasons given above.
//
// The frame_bound productions UNBOUNDED PRECEDING and UNBOUNDED FOLLOWING are
// even messier: since UNBOUNDED is an unreserved keyword (per spec!), there is
// no principled way to distinguish these from the productions a_expr
// PRECEDING/FOLLOWING. We hack this up by giving UNBOUNDED slightly lower
// precedence than PRECEDING and FOLLOWING. At present this doesn't appear to
// cause UNBOUNDED to be treated differently from other unreserved keywords
// anywhere else in the grammar, but it's definitely risky. We can blame any
// funny behavior of UNBOUNDED on the SQL standard, though.
%nonassoc  UNBOUNDED         // ideally should have same precedence as IDENT
%nonassoc  IDENT NULL PARTITION RANGE ROWS GROUPS PRECEDING FOLLOWING CUBE ROLLUP
%left      CONCAT FETCHVAL FETCHTEXT FETCHVAL_PATH FETCHTEXT_PATH REMOVE_PATH  // multi-character ops
%left      '|'
%left      '#'
%left      '&'
%left      LSHIFT RSHIFT INET_CONTAINS_OR_EQUALS INET_CONTAINED_BY_OR_EQUALS AND_AND SQRT CBRT
%left      '@'
%left      '+' '-'
%left      '*' '/' FLOORDIV '%'
%left      '^'
%left      OPERATOR
// Unary Operators
%left      AT                // sets precedence for AT TIME ZONE
%left      COLLATE
%right     UMINUS
%left      '[' ']'
%left      '(' ')'
%left      TYPEANNOTATE
%left      TYPECAST
%left      '.'
// These might seem to be low-precedence, but actually they are not part
// of the arithmetic hierarchy at all in their use as JOIN operators.
// We make them high-precedence to support their use as function names.
// They wouldn't be given a precedence at all, were it not that we need
// left-associativity among the JOIN rules themselves.
%left      JOIN CROSS LEFT FULL RIGHT INNER NATURAL
%right     HELPTOKEN

%%

stmt_block:
  stmt
  {
    sqllex.(*lexer).SetStmt($1.stmt())
  }

stmt:
  HELPTOKEN { return helpWith(sqllex, "") }
| non_transaction_stmt
| transaction_stmt  // help texts in sub-rule
| /* EMPTY */
  {
    $$.val = tree.Statement(nil)
  }

non_transaction_stmt:
  preparable_stmt   // help texts in sub-rule
| analyze_stmt      // EXTEND WITH HELP: ANALYZE
| call_stmt
| copy_from_stmt
| comment_stmt
| execute_stmt      // EXTEND WITH HELP: EXECUTE
| deallocate_stmt   // EXTEND WITH HELP: DEALLOCATE
| discard_stmt      // EXTEND WITH HELP: DISCARD
| grant_stmt        // EXTEND WITH HELP: GRANT
| prepare_stmt      // EXTEND WITH HELP: PREPARE
| revoke_stmt       // EXTEND WITH HELP: REVOKE
| savepoint_stmt    // EXTEND WITH HELP: SAVEPOINT
| release_stmt      // EXTEND WITH HELP: RELEASE
| refresh_stmt      // EXTEND WITH HELP: REFRESH
| set_stmt // help texts in sub-rule
| close_cursor_stmt
| declare_cursor_stmt
| reindex_stmt
| vacuum_stmt

stmt_list:
  non_transaction_stmt
  {
    $$.val = []tree.Statement{$1.stmt()}
  }
| stmt_list ';' non_transaction_stmt
  {
    $$.val = append($1.stmts(), $3.stmt())
  }

// %Help: ALTER
// %Category: Group
// %Text: ALTER TABLE, ALTER INDEX, ALTER VIEW, ALTER SEQUENCE, ALTER DATABASE, ALTER USER, ALTER ROLE
alter_stmt:
  alter_ddl_stmt        // help texts in sub-rule
| alter_role_stmt       // EXTEND WITH HELP: ALTER ROLE
| alter_aggregate_stmt  // EXTEND WITH HELP: ALTER AGGREGATE
| alter_collation_stmt  // EXTEND WITH HELP: ALTER COLLATION
| alter_conversion_stmt // EXTEND WITH HELP: ALTER CONVERSION
| ALTER error           // SHOW HELP: ALTER

alter_ddl_stmt:
  alter_table_stmt              // EXTEND WITH HELP: ALTER TABLE
| alter_index_stmt              // EXTEND WITH HELP: ALTER INDEX
| alter_function_stmt           // EXTEND WITH HELP: ALTER FUNCTION
| alter_procedure_stmt          // EXTEND WITH HELP: ALTER PROCEDURE
| alter_view_stmt               // EXTEND WITH HELP: ALTER VIEW
| alter_materialized_view_stmt  // EXTEND WITH HELP: ALTER MATERIALIZED VIEW
| alter_sequence_stmt           // EXTEND WITH HELP: ALTER SEQUENCE
| alter_database_stmt           // EXTEND WITH HELP: ALTER DATABASE
| alter_default_privileges_stmt // EXTEND WITH HELP: ALTER DEFAULT PRIVILEGES
| alter_schema_stmt             // EXTEND WITH HELP: ALTER SCHEMA
| alter_type_stmt               // EXTEND WITH HELP: ALTER TYPE
| alter_trigger_stmt            // EXTEND WITH HELP: ALTER TRIGGER
| alter_language_stmt           // EXTEND WITH HELP: ALTER LANGUAGE
| alter_domain_stmt             // EXTEND WITH HELP: ALTER DOMAIN

// %Help: ALTER TABLE - change the definition of a table
// %Category: DDL
// %Text:
// ALTER TABLE [IF EXISTS] <tablename> <command> [, ...]
//
// Commands:
//   ALTER TABLE ... ADD [COLUMN] [IF NOT EXISTS] <colname> <type> [<qualifiers...>]
//   ALTER TABLE ... ADD <constraint>
//   ALTER TABLE ... DROP [COLUMN] [IF EXISTS] <colname> [RESTRICT | CASCADE]
//   ALTER TABLE ... DROP CONSTRAINT [IF EXISTS] <constraintname> [RESTRICT | CASCADE]
//   ALTER TABLE ... ALTER [COLUMN] <colname> {SET DEFAULT <expr> | DROP DEFAULT}
//   ALTER TABLE ... ALTER [COLUMN] <colname> DROP NOT NULL
//   ALTER TABLE ... ALTER [COLUMN] <colname> DROP STORED
//   ALTER TABLE ... ALTER [COLUMN] <colname> [SET DATA] TYPE <type> [COLLATE <collation>]
//   ALTER TABLE ... RENAME TO <newname>
//   ALTER TABLE ... RENAME [COLUMN] <colname> TO <newname>
//   ALTER TABLE ... VALIDATE CONSTRAINT <constraintname>
//   ALTER TABLE ... SET SCHEMA <newschemaname>
//
// Column qualifiers:
//   [CONSTRAINT <constraintname>] {NULL | NOT NULL | UNIQUE | PRIMARY KEY | CHECK (<expr>) | DEFAULT <expr>}
//   REFERENCES <tablename> [( <colnames...> )]
//   COLLATE <collationname>
//
// %SeeAlso: WEBDOCS/alter-table.html
alter_table_stmt:
  alter_onetable_stmt
| alter_rename_table_stmt
| alter_table_set_schema_stmt
| alter_table_all_in_tablespace_stmt
| alter_table_parition_stmt
// ALTER TABLE has its error help token here because the ALTER TABLE
// prefix is spread over multiple non-terminals.
| ALTER TABLE error     // SHOW HELP: ALTER TABLE

// %Help: ALTER VIEW - change the definition of a view
// %Category: DDL
// %Text:
// ALTER [MATERIALIZED] VIEW [IF EXISTS] <name> RENAME TO <newname>
// ALTER [MATERIALIZED] VIEW [IF EXISTS] <name> SET SCHEMA <newschemaname>
// %SeeAlso: WEBDOCS/alter-view.html
alter_view_stmt:
  ALTER VIEW relation_expr alter_view_cmd
  {
    $$.val = &tree.AlterView{Name: $3.unresolvedObjectName(), IfExists: false, Cmd: $4.alterViewCmd()}
  }
| ALTER VIEW IF EXISTS relation_expr alter_view_cmd
  {
    $$.val = &tree.AlterView{Name: $5.unresolvedObjectName(), IfExists: true, Cmd: $6.alterViewCmd()}
  }
// ALTER VIEW has its error help token here because the ALTER VIEW
// prefix is spread over multiple non-terminals.
| ALTER VIEW error // SHOW HELP: ALTER VIEW

alter_view_cmd:
  ALTER opt_column column_name alter_column_default
  {
    $$.val = &tree.AlterViewSetDefault{Column: tree.Name($3), Default: $4.expr()}
  }
| owner_to
  {
    $$.val = &tree.AlterViewOwnerTo{Owner: $1}
  }
| RENAME opt_column column_name TO column_name
  {
    $$.val = &tree.AlterViewRenameColumn{Column: tree.Name($3), NewName: tree.Name($5)}
  }
| RENAME TO view_name
  {
    $$.val = &tree.AlterViewRenameTo{Rename: $3.unresolvedObjectName()}
  }
| set_schema
  {
    $$.val = &tree.AlterViewSetSchema{Schema: $1}
  }
| SET '(' view_options ')'
  {
    $$.val = &tree.AlterViewSetOption{Params: $3.viewOptions()}
  }
| RESET '(' view_options ')'
  {
    $$.val = &tree.AlterViewSetOption{Reset: true, Params: $3.viewOptions()}
  }

// %Help: ALTER SEQUENCE - change the definition of a sequence
// %Category: DDL
// %Text:
// ALTER SEQUENCE [IF EXISTS] <name>
//   [INCREMENT <increment>]
//   [MINVALUE <minvalue> | NO MINVALUE]
//   [MAXVALUE <maxvalue> | NO MAXVALUE]
//   [START <start>]
//   [[NO] CYCLE]
// ALTER SEQUENCE [IF EXISTS] <name> RENAME TO <newname>
// ALTER SEQUENCE [IF EXISTS] <name> SET SCHEMA <newschemaname>
alter_sequence_stmt:
  alter_rename_sequence_stmt
| alter_sequence_options_stmt
| alter_sequence_set_schema_stmt
| alter_sequence_set_log_stmt
| alter_sequence_owner_to_stmt

alter_sequence_options_stmt:
  ALTER SEQUENCE sequence_name opt_alter_seq_option_list
  {
    $$.val = &tree.AlterSequence{Name: $3.unresolvedObjectName(), Options: $4.seqOpts(), IfExists: false}
  }
| ALTER SEQUENCE IF EXISTS sequence_name opt_alter_seq_option_list
  {
    $$.val = &tree.AlterSequence{Name: $5.unresolvedObjectName(), Options: $6.seqOpts(), IfExists: true}
  }

alter_sequence_set_log_stmt:
  ALTER SEQUENCE relation_expr SET logged_or_unlogged
  {
    $$.val = &tree.AlterSequence{Name: $3.unresolvedObjectName(), IfExists: false, SetLog: true, Logged: $5.bool()}
  }
| ALTER SEQUENCE IF EXISTS relation_expr SET logged_or_unlogged
  {
    $$.val = &tree.AlterSequence{Name: $5.unresolvedObjectName(), IfExists: true, SetLog: true, Logged: $7.bool()}
  }

alter_sequence_owner_to_stmt:
  ALTER SEQUENCE relation_expr owner_to
  {
    $$.val = &tree.AlterSequence{Name: $3.unresolvedObjectName(), IfExists: true, Owner: $4}
  }

// %Help: ALTER DATABASE - change the definition of a database
// %Category: DDL
// %Text:
// ALTER DATABASE <name> RENAME TO <newname>
// ALTER DATABASE <name> OWNER TO <newowner>
// %SeeAlso: WEBDOCS/alter-database.html
alter_database_stmt:
  alter_rename_database_stmt
| opt_alter_database
| alter_database_to_schema_stmt
// ALTER DATABASE has its error help token here because the ALTER DATABASE
// prefix is spread over multiple non-terminals.
| ALTER DATABASE error // SHOW HELP: ALTER DATABASE

opt_alter_database:
  ALTER DATABASE database_name owner_to
  {
    $$.val = &tree.AlterDatabase{Name: tree.Name($3), Owner: $4}
  }
| ALTER DATABASE database_name opt_database_with_options
  {
    $$.val = &tree.AlterDatabase{Name: tree.Name($3), Options: $4.databaseOptionList()}
  }
| ALTER DATABASE database_name set_tablespace
  {
    $$.val = &tree.AlterDatabase{Name: tree.Name($3), Tablespace: $4}
  }
| ALTER DATABASE database_name REFRESH COLLATION VERSION
  {
    $$.val = &tree.AlterDatabase{Name: tree.Name($3), RefreshCollationVersion: true}
  }
| ALTER DATABASE database_name SET generic_set_single_config
  {
    $$.val = &tree.AlterDatabase{Name: tree.Name($3), SetVar: $5.setVar()}
  }
| ALTER DATABASE database_name RESET name
  {
    $$.val = &tree.AlterDatabase{Name: tree.Name($3), Tablespace: $5}
  }
| ALTER DATABASE database_name RESET ALL
  {
    $$.val = &tree.AlterDatabase{Name: tree.Name($3), ResetAll: true}
  }

opt_database_with_options:
  /* EMPTY */
  {
    $$.val = []tree.DatabaseOption(nil)
  }
| opt_database_options_list
  {
    $$.val = $1.databaseOptionList()
  }
| WITH opt_database_options_list
  {
    $$.val = $2.databaseOptionList()
  }

opt_database_options_list:
  opt_database_options
  {
    $$.val = []tree.DatabaseOption{$1.databaseOption()}
  }
| opt_database_options_list opt_database_options
  {
    $$.val = append($1.databaseOptionList(), $2.databaseOption())
  }

opt_database_options:
  ALLOW_CONNECTIONS a_expr
  {
    $$.val = tree.DatabaseOption{Opt: tree.OptAllowConnections, Val: $2.expr()}
  }
| CONNECTION LIMIT a_expr
  {
    $$.val = tree.DatabaseOption{Opt: tree.OptConnectionLimit, Val: $3.expr()}
  }
| IS_TEMPLATE a_expr
  {
    $$.val = tree.DatabaseOption{Opt: tree.OptIsTemplate, Val: $2.expr()}
  }

alter_default_privileges_stmt:
  ALTER DEFAULT PRIVILEGES adp_abbreviated_grant_or_revoke
  {
    $$.val = $4.alterDefaultPrivileges()
  }
| ALTER DEFAULT PRIVILEGES FOR role_or_user opt_role_list adp_abbreviated_grant_or_revoke
  {
    adp := $7.alterDefaultPrivileges()
    adp.ForRole = $5.bool()
    adp.TargetRoles = $6.strs()
    $$.val = adp
  }
| ALTER DEFAULT PRIVILEGES IN SCHEMA schema_name_list adp_abbreviated_grant_or_revoke
  {
    adp := $7.alterDefaultPrivileges()
    adp.Target.InSchema = $6.strs()
    $$.val = adp
  }
| ALTER DEFAULT PRIVILEGES FOR role_or_user opt_role_list IN SCHEMA schema_name_list adp_abbreviated_grant_or_revoke
  {
    adp := $10.alterDefaultPrivileges()
    adp.ForRole = $5.bool()
    adp.TargetRoles = $6.strs()
    adp.Target.InSchema = $9.strs()
    $$.val = adp
  }

adp_abbreviated_grant_or_revoke:
  GRANT privileges ON targets_for_alter_def_priv TO opt_role_list opt_with_grant_option
  {
    $$.val = &tree.AlterDefaultPrivileges{Privileges: $2.privilegeList(), Target: $4.targetList(), Grantees: $6.strs(), GrantOption: $7.bool(), Grant: true}
  }
| REVOKE opt_grant_option_for privileges ON targets_for_alter_def_priv FROM opt_role_list opt_drop_behavior
  {
    $$.val = &tree.AlterDefaultPrivileges{GrantOption: $2.bool(), Privileges: $3.privilegeList(), Target: $5.targetList(), Grantees: $7.strs(), DropBehavior: $8.dropBehavior()}
  }

opt_grant_option_for:
  /* EMPTY */
  {
    $$.val = false
  }
| GRANT OPTION FOR
  {
    $$.val = true
  }

targets_for_alter_def_priv:
  TABLES
  {
    $$.val = tree.TargetList{TargetType: privilege.Table}
  }
| SEQUENCES
  {
    $$.val = tree.TargetList{TargetType: privilege.Sequence}
  }
| FUNCTIONS
  {
    $$.val = tree.TargetList{TargetType: privilege.Function}
  }
| ROUTINES
  {
    $$.val = tree.TargetList{TargetType: privilege.Routine}
  }
| TYPES
  {
    $$.val = tree.TargetList{TargetType: privilege.Type}
  }
| SCHEMAS
  {
    $$.val = tree.TargetList{TargetType: privilege.Schema}
  }

opt_role_list:
  opt_role
  {
    $$.val = []string{$1}
  }
| opt_role_list ',' opt_role
  {
    $$.val = append($1.strs(), $3)
  }

// option 'PUBLIC' is under 'unreserved_keywords', so it's included in 'role_spec' rule.
opt_role:
  role_spec
  {
    $$ = string($1)
  }
| GROUP role_spec
  {
    $$ = string($1) + " " + string($2)
  }

// %Help: ALTER INDEX - change the definition of an index
// %Category: DDL
// %Text:
// ALTER INDEX [IF EXISTS] <idxname> <command>
//
// Commands:
//   ALTER INDEX ... RENAME TO <newname>
//
// %SeeAlso: WEBDOCS/alter-index.html
alter_index_stmt:
  alter_oneindex_stmt
| alter_rename_index_stmt
| alter_index_all_in_tablespace_stmt
// ALTER INDEX has its error help token here because the ALTER INDEX
// prefix is spread over multiple non-terminals.
| ALTER INDEX error // SHOW HELP: ALTER INDEX

alter_language_stmt:
  ALTER opt_procedural LANGUAGE name RENAME TO name
  {
    $$.val = &tree.AlterLanguage{Name: tree.Name($4), Procedural: $2.bool(), NewName: tree.Name($7)}
  }
| ALTER opt_procedural LANGUAGE name owner_to
  {
    $$.val = &tree.AlterLanguage{Name: tree.Name($4), Procedural: $2.bool(), Owner: $5}
  }

alter_domain_stmt:
  ALTER DOMAIN type_name alter_domain_cmd
  {
    $$.val = &tree.AlterDomain{
      Name: $3.unresolvedObjectName(),
      Cmd: $4.alterDomainCmd(),
    }
  }

alter_domain_cmd:
  SET DEFAULT a_expr
  {
    $$.val = &tree.AlterDomainSetDrop{IsSet: true, Default: $3.expr()}
  }
| DROP DEFAULT
  {
    $$.val = &tree.AlterDomainSetDrop{IsSet: false}
  }
| SET NOT NULL
  {
    $$.val = &tree.AlterDomainSetDrop{IsSet: true, NotNull: true}
  }
| DROP NOT NULL
  {
    $$.val = &tree.AlterDomainSetDrop{IsSet: false, NotNull: true}
  }
| ADD domain_constraint opt_not_valid
  {
    $$.val = &tree.AlterDomainConstraint{
      Action: tree.AlterDomainAddConstraint,
      Constraint: $2.domainConstraint(),
      NotValid: $3.bool(),
    }
  }
| DROP CONSTRAINT constraint_name opt_drop_behavior
  {
    $$.val = &tree.AlterDomainConstraint{
      Action: tree.AlterDomainDropConstraint,
      IfExists: false,
      ConstraintName: tree.Name($3),
      DropBehavior: $4.dropBehavior(),
    }
  }
| DROP CONSTRAINT IF EXISTS constraint_name opt_drop_behavior
  {
    $$.val = &tree.AlterDomainConstraint{
      Action: tree.AlterDomainDropConstraint,
      IfExists: true,
      ConstraintName: tree.Name($5),
      DropBehavior: $6.dropBehavior(),
    }
  }
| RENAME CONSTRAINT constraint_name TO constraint_name
  {
    $$.val = &tree.AlterDomainConstraint{
      Action: tree.AlterDomainRenameConstraint,
      ConstraintName: tree.Name($3),
      NewName: tree.Name($5),
    }
  }
| VALIDATE CONSTRAINT constraint_name
  {
    $$.val = &tree.AlterDomainConstraint{
      Action: tree.AlterDomainValidateConstraint,
      ConstraintName: tree.Name($3),
    }
  }
| owner_to
  {
    $$.val = &tree.AlterDomainOwner{Owner: $1}
  }
| RENAME TO name
  {
    $$.val = &tree.AlterDomainRename{NewName: $3}
  }
| set_schema
  {
    $$.val = &tree.AlterDomainSetSchema{Schema: $1}
  }

opt_not_valid:
  /* EMPTY */
  {
    $$.val = false
  }
| NOT VALID
  {
    $$.val = true
  }

alter_function_stmt:
  ALTER FUNCTION routine_name opt_routine_args_with_paren alter_function_option_list opt_restrict
  {
    $$.val = &tree.AlterFunction{Name: $3.unresolvedObjectName(), Args: $4.routineArgs(), Options: $5.routineOptions(), Restrict: $6.bool()}
  }
| ALTER FUNCTION routine_name opt_routine_args_with_paren RENAME TO routine_name
  {
    $$.val = &tree.AlterFunction{Name: $3.unresolvedObjectName(), Args: $4.routineArgs(), Rename: $7.unresolvedObjectName()}
  }
| ALTER FUNCTION routine_name opt_routine_args_with_paren owner_to
  {
    $$.val = &tree.AlterFunction{Name: $3.unresolvedObjectName(), Args: $4.routineArgs(), Owner: $5}
  }
| ALTER FUNCTION routine_name opt_routine_args_with_paren set_schema
  {
    $$.val = &tree.AlterFunction{Name: $3.unresolvedObjectName(), Args: $4.routineArgs(), Schema: $5}
  }
| ALTER FUNCTION routine_name opt_routine_args_with_paren opt_no DEPENDS ON EXTENSION name
  {
    $$.val = &tree.AlterFunction{Name: $3.unresolvedObjectName(), Args: $4.routineArgs(), No: $5.bool(), Extension: $9}
  }

opt_restrict:
  /* EMPTY */
  {
    $$.val = false
  }
| RESTRICT
  {
    $$.val = true
  }

alter_procedure_stmt:
  ALTER PROCEDURE routine_name opt_routine_args_with_paren alter_procedure_option_list opt_restrict
  {
    $$.val = &tree.AlterProcedure{Name: $3.unresolvedObjectName(), Args: $4.routineArgs(), Options: $5.routineOptions(), Restrict: $6.bool()}
  }
| ALTER PROCEDURE routine_name opt_routine_args_with_paren RENAME TO routine_name
  {
    $$.val = &tree.AlterProcedure{Name: $3.unresolvedObjectName(), Args: $4.routineArgs(), Rename: $7.unresolvedObjectName()}
  }
| ALTER PROCEDURE routine_name opt_routine_args_with_paren owner_to
  {
    $$.val = &tree.AlterProcedure{Name: $3.unresolvedObjectName(), Args: $4.routineArgs(), Owner: $5}
  }
| ALTER PROCEDURE routine_name opt_routine_args_with_paren set_schema
  {
    $$.val = &tree.AlterProcedure{Name: $3.unresolvedObjectName(), Args: $4.routineArgs(), Schema: $5}
  }
| ALTER PROCEDURE routine_name opt_routine_args_with_paren opt_no DEPENDS ON EXTENSION name
  {
    $$.val = &tree.AlterProcedure{Name: $3.unresolvedObjectName(), Args: $4.routineArgs(), No: $5.bool(), Extension: $9}
  }

alter_onetable_stmt:
  ALTER TABLE relation_expr alter_table_cmd
  {
    $$.val = &tree.AlterTable{Table: $3.unresolvedObjectName(), IfExists: false, Cmds: $4.alterTableCmds()}
  }
| ALTER TABLE IF EXISTS relation_expr alter_table_cmd
  {
    $$.val = &tree.AlterTable{Table: $5.unresolvedObjectName(), IfExists: true, Cmds: $6.alterTableCmds()}
  }

alter_oneindex_stmt:
  ALTER INDEX table_index_name alter_index_cmd
  {
    $$.val = &tree.AlterIndex{Index: $3.tableIndexName(), IfExists: false, Cmd: $4.alterIndexCmd()}
  }
| ALTER INDEX IF EXISTS table_index_name alter_index_cmd
  {
    $$.val = &tree.AlterIndex{Index: $5.tableIndexName(), IfExists: true, Cmd: $6.alterIndexCmd()}
  }
| ALTER INDEX table_index_name ATTACH PARTITION index_name
  {
    $$.val = &tree.AlterIndex{Index: $3.tableIndexName(), Cmd: &tree.AlterIndexAttachPartition{Index: tree.UnrestrictedName($6)}}
  }
| ALTER INDEX table_index_name opt_no DEPENDS ON EXTENSION name
  {
    $$.val = &tree.AlterIndex{Index: $3.tableIndexName(), Cmd: &tree.AlterIndexExtension{No: $4.bool(), Extension: $8}}
  }

opt_no:
  /* EMPTY */
  {
    $$.val = false
  }
| NO
  {
    $$.val = true
  }

alter_index_all_in_tablespace_stmt:
  ALTER INDEX ALL IN TABLESPACE tablespace_name opt_owned_by_list set_tablespace opt_nowait
  {
    $$.val = &tree.AlterIndexAllInTablespace{
      Name: tree.Name($6), OwnedBy: $7.strs(), Tablespace: $8, NoWait: $9.bool(),
    }
  }

alter_table_cmd:
  // ALTER TABLE <name> RENAME [COLUMN] <name> TO <newname>
  RENAME opt_column column_name TO column_name
  {
    $$.val = tree.AlterTableCmds{&tree.AlterTableRenameColumn{Column: tree.Name($3), NewName: tree.Name($5) }}
  }
  // ALTER TABLE <name> RENAME CONSTRAINT <name> TO <newname>
| RENAME CONSTRAINT constraint_name TO constraint_name
  {
    $$.val = tree.AlterTableCmds{&tree.AlterTableRenameConstraint{Constraint: tree.Name($3), NewName: tree.Name($5) }}
  }
| alter_table_actions

alter_table_actions:
  alter_table_action
  {
    $$.val = tree.AlterTableCmds{$1.alterTableCmd()}
  }
| alter_table_actions ',' alter_table_action
  {
    $$.val = append($1.alterTableCmds(), $3.alterTableCmd())
  }

alter_table_action:
  // ALTER TABLE <name> ADD <coldef>
  ADD alter_column_def
  {
    $$.val = &tree.AlterTableAddColumn{IfNotExists: false, ColumnDef: $2.colDef()}
  }
  // ALTER TABLE <name> ADD IF NOT EXISTS <coldef>
| ADD IF NOT EXISTS alter_column_def
  {
    $$.val = &tree.AlterTableAddColumn{IfNotExists: true, ColumnDef: $5.colDef()}
  }
  // ALTER TABLE <name> ADD COLUMN <coldef>
| ADD COLUMN alter_column_def
  {
    $$.val = &tree.AlterTableAddColumn{IfNotExists: false, ColumnDef: $3.colDef()}
  }
  // ALTER TABLE <name> ADD COLUMN IF NOT EXISTS <coldef>
| ADD COLUMN IF NOT EXISTS alter_column_def
  {
    $$.val = &tree.AlterTableAddColumn{IfNotExists: true, ColumnDef: $6.colDef()}
  }
  // ALTER TABLE <name> DROP [COLUMN] <colname> [RESTRICT|CASCADE]
| DROP opt_column column_name opt_drop_behavior
  {
    $$.val = &tree.AlterTableDropColumn{
      IfExists: false,
      Column: tree.Name($3),
      DropBehavior: $4.dropBehavior(),
    }
  }
  // ALTER TABLE <name> DROP [COLUMN] IF EXISTS <colname> [RESTRICT|CASCADE]
| DROP opt_column IF EXISTS column_name opt_drop_behavior
  {
    $$.val = &tree.AlterTableDropColumn{
      IfExists: true,
      Column: tree.Name($5),
      DropBehavior: $6.dropBehavior(),
    }
  }
| ALTER opt_column alter_opt_column_options
  {
    $$.val = $3.alterTableCmd()
  }
  // ALTER TABLE <name> ADD CONSTRAINT ...
| ADD table_constraint opt_validate_behavior
  {
    $$.val = &tree.AlterTableAddConstraint{
      ConstraintDef: $2.constraintDef(),
      ValidationBehavior: $3.validationBehavior(),
    }
  }
  // ALTER TABLE <name> ADD CONSTRAINT ... USING INDEX
| ADD CONSTRAINT constraint_name unique_or_primary USING INDEX index_name opt_deferrable_mode opt_initially
  {
    $$.val = tree.AlterTableConstraintUsingIndex{Constraint: tree.Name($3), IsUnique: $4.bool(), Index: tree.Name($7), Deferrable: $8.deferrableMode(), Initially: $9.initiallyMode()}
  }
  // ALTER TABLE <name> ALTER CONSTRAINT ...
| ALTER CONSTRAINT constraint_name opt_deferrable_mode opt_initially
  {
    $$.val = &tree.AlterTableAlterConstraint{Constraint: tree.Name($3), Deferrable: $4.deferrableMode(), Initially: $5.initiallyMode()}
  }
  // ALTER TABLE <name> VALIDATE CONSTRAINT ...
| VALIDATE CONSTRAINT constraint_name
  {
    $$.val = &tree.AlterTableValidateConstraint{
      Constraint: tree.Name($3),
    }
  }
  // ALTER TABLE <name> DROP CONSTRAINT IF EXISTS <name> [RESTRICT|CASCADE]
| DROP CONSTRAINT IF EXISTS constraint_name opt_drop_behavior
  {
    $$.val = &tree.AlterTableDropConstraint{
      IfExists: true,
      Constraint: tree.Name($5),
      DropBehavior: $6.dropBehavior(),
    }
  }
  // ALTER TABLE <name> DROP CONSTRAINT <name> [RESTRICT|CASCADE]
| DROP CONSTRAINT constraint_name opt_drop_behavior
  {
    $$.val = &tree.AlterTableDropConstraint{
      IfExists: false,
      Constraint: tree.Name($3),
      DropBehavior: $4.dropBehavior(),
    }
  }
  // ALTER TABLE <name> ... TRIGGER ...
| enable_or_disable_trigger
  // ALTER TABLE <name> ... RULE ...
| enable_or_disable_rule
  // ALTER TABLE <name> ... LEVEL SECURITY
| row_level_security
  // ALTER TABLE CLUSTER ON
| CLUSTER ON index_name
  {
    $$.val = &tree.AlterTableCluster{OnIndex: tree.Name($3)}
  }
  // ALTER TABLE <name> ... SET WITHOUT CLUSTER
| SET WITHOUT CLUSTER
  {
    $$.val = &tree.AlterTableCluster{Without: true}
  }
| SET WITHOUT OIDS { /* used for backward-compatibility, so no-op */ }
| SET ACCESS METHOD non_reserved_word_or_sconst
  {
    $$.val = &tree.AlterTableSetAccessMethod{Method: $4}
  }
| set_tablespace
  {
    $$.val = &tree.AlterTableSetTablespace{Tablespace: $1}
  }
| SET logged_or_unlogged
  {
    $$.val = &tree.AlterTableSetLog{Logged: $2.bool()}
  }
| SET '(' storage_parameter_list ')'
  {
    $$.val = &tree.AlterTableSetStorage{Params: $3.storageParams()}
  }
| RESET '(' storage_parameter_list ')'
  {
    $$.val = &tree.AlterTableSetStorage{Params: $3.storageParams(), IsReset: true}
  }
| INHERIT table_name
  {
    $$.val = &tree.AlterTableInherit{Inherit: true, Table: $2.unresolvedObjectName().ToTableName()}
  }
| NO INHERIT table_name
  {
    $$.val = &tree.AlterTableInherit{Table: $3.unresolvedObjectName().ToTableName()}
  }
| OF type_name
  {
    $$.val = &tree.AlterTableOfType{Type: $2.typeReference()}
  }
| NOT OF
  {
    $$.val = &tree.AlterTableOfType{NotOf: true}
  }
  // ALTER TABLE <name> OWNER TO <newowner>
| owner_to
  {
    $$.val = &tree.AlterTableOwner{
      Owner: $1,
    }
  }
| replica_identity_option

alter_materialized_view_actions:
  alter_materialized_view_action
  {
    $$.val = tree.AlterTableCmds{$1.alterTableCmd()}
  }
| alter_materialized_view_actions ',' alter_materialized_view_action
  {
    $$.val = append($1.alterTableCmds(), $3.alterTableCmd())
  }

alter_materialized_view_action:
  ALTER opt_column alter_materialized_view_opt_column_options
  {
    $$.val = $3.alterTableCmd()
  }
| CLUSTER ON index_name
  {
    $$.val = &tree.AlterTableCluster{OnIndex: tree.Name($3)}
  }
| SET WITHOUT CLUSTER
  {
    $$.val = &tree.AlterTableCluster{Without: true}
  }
| SET ACCESS METHOD non_reserved_word_or_sconst
  {
    $$.val = &tree.AlterTableSetAccessMethod{Method: $4}
  }
| set_tablespace
  {
    $$.val = &tree.AlterTableSetTablespace{Tablespace: $1}
  }
| SET '(' storage_parameter_list ')'
  {
    $$.val = &tree.AlterTableSetStorage{Params: $3.storageParams()}
  }
| RESET '(' storage_parameter_list ')'
  {
    $$.val = &tree.AlterTableSetStorage{Params: $3.storageParams(), IsReset: true}
  }
| owner_to
  {
    $$.val = &tree.AlterTableOwner{
      Owner: $1,
    }
  }

enable_or_disable_trigger:
  DISABLE TRIGGER trigger_option
  {
    $$.val = &tree.AlterTableTrigger{Disable: true, Trigger: $3}
  }
| ENABLE TRIGGER trigger_option
  {
    $$.val = &tree.AlterTableTrigger{Trigger: $3}
  }
| ENABLE REPLICA TRIGGER trigger_name
  {
    $$.val = &tree.AlterTableTrigger{Trigger: $4, IsReplica: true}
  }
| ENABLE ALWAYS TRIGGER trigger_name
  {
    $$.val = &tree.AlterTableTrigger{Trigger: $4, IsAlways: true}
  }

enable_or_disable_rule:
  DISABLE RULE name
  {
    $$.val = &tree.AlterTableRule{Disable: true, Rule: $3}
  }
| ENABLE RULE name
  {
    $$.val = &tree.AlterTableRule{Rule: $3}
  }
| ENABLE REPLICA RULE name
  {
    $$.val = &tree.AlterTableRule{Rule: $3, IsReplica: true}
  }
| ENABLE ALWAYS RULE name
  {
    $$.val = &tree.AlterTableRule{Rule: $3, IsAlways: true}
  }

replica_identity_option:
  REPLICA IDENTITY DEFAULT
  {
    $$.val = &tree.AlterTableReplicaIdentity{Type: tree.ReplicaIdentityDefault}
  }
| REPLICA IDENTITY USING INDEX index_name
  {
    $$.val = &tree.AlterTableReplicaIdentity{Type: tree.ReplicaIdentityUsingIndex, Index: tree.Name($5)}
  }
| REPLICA IDENTITY FULL
  {
    $$.val = &tree.AlterTableReplicaIdentity{Type: tree.ReplicaIdentityFull}
  }
| REPLICA IDENTITY NOTHING
  {
    $$.val = &tree.AlterTableReplicaIdentity{Type: tree.ReplicaIdentityNothing}
  }

logged_or_unlogged:
  LOGGED
  {
    $$.val = true
  }
| UNLOGGED
  {
    $$.val = false
  }

row_level_security:
  DISABLE ROW LEVEL SECURITY
  {
    $$.val = &tree.AlterTableRowLevelSecurity{Type: tree.RowLevelSecurityDisable}
  }
| ENABLE
  {
    $$.val = &tree.AlterTableRowLevelSecurity{Type: tree.RowLevelSecurityEnable}
  }
| FORCE
  {
    $$.val = &tree.AlterTableRowLevelSecurity{Type: tree.RowLevelSecurityForce}
  }
| NO FORCE
  {
    $$.val = &tree.AlterTableRowLevelSecurity{Type: tree.RowLevelSecurityNoForce}
  }

trigger_option:
  trigger_name
| ALL
| USER

unique_or_primary:
  UNIQUE
  {
    $$.val = true
  }
| PRIMARY
  {
    $$.val = false
  }

alter_opt_column_options:
  // ALTER TABLE <name> ALTER [COLUMN] <colname>
  //     [SET DATA] TYPE <typename>
  //     [ COLLATE collation ]
  //     [ USING <expression> ]
  column_name opt_set_data TYPE typename opt_collate opt_alter_column_using
  {
    $$.val = &tree.AlterTableAlterColumnType{
      Column: tree.Name($1),
      ToType: $4.typeReference(),
      Collation: $5.unresolvedObjectName().UnquotedString(),
      Using: $6.expr(),
    }
  }
  // ALTER TABLE <name> ALTER [COLUMN] <colname> {SET DEFAULT <expr>|DROP DEFAULT}
| column_name alter_column_default
  {
    $$.val = &tree.AlterTableSetDefault{Column: tree.Name($1), Default: $2.expr()}
  }
  // ALTER TABLE <name> ALTER [COLUMN] <colname> SET NOT NULL
| column_name SET NOT NULL
  {
    $$.val = &tree.AlterTableSetNotNull{Column: tree.Name($1)}
  }
  // ALTER TABLE <name> ALTER [COLUMN] <colname> DROP NOT NULL
| column_name DROP NOT NULL
  {
    $$.val = &tree.AlterTableDropNotNull{Column: tree.Name($1)}
  }
  // ALTER TABLE <name> ALTER [COLUMN] <colname> DROP EXPRESSION
| column_name DROP EXPRESSION opt_if_exists
  {
    $$.val = &tree.AlterTableDropExprIden{Column: tree.Name($1), IfExists: $4.bool()}
  }
| column_name ADD col_qual_generated_identity
  {
    $$.val = &tree.AlterTableComputed{Column: tree.Name($1), IsAdd: true, AddDefs: $3.colQualElem()}
  }
| column_name alter_column_set_seq_elem_list
  {
    $$.val = &tree.AlterTableComputed{Column: tree.Name($1), Defs: $2.alterColComputedList()}
  }
| column_name DROP IDENTITY opt_if_exists
  {
    $$.val = &tree.AlterTableDropExprIden{Column: tree.Name($1), IsIdentity: true, IfExists: $4.bool()}
  }
| alter_materialized_view_opt_column_options

alter_materialized_view_opt_column_options:
  column_name SET STATISTICS signed_iconst
  {
    $$.val = &tree.AlterTableSetStatistics{Column: tree.Name($1), Num: $4.expr()}
  }
| column_name SET '(' attribution_list ')'
  {
    $$.val = &tree.AlterTableSetAttribution{Column: tree.Name($1), Params: $4.storageParams()}
  }
| column_name RESET '(' attribution_list ')'
  {
    $$.val = &tree.AlterTableSetAttribution{Column: tree.Name($1), Reset: true, Params: $4.storageParams()}
  }
| column_name SET STORAGE col_storage_option
  {
    $$.val = &tree.AlterTableColSetStorage{Column: tree.Name($1), Type: $4.storageType()}
  }
| column_name SET COMPRESSION unrestricted_name
  {
    $$.val = &tree.AlterTableSetCompression{Column: tree.Name($1), Compression: $4}
  }

/* re-using StorageParam for attribution key-value pair */
attribution_list: storage_parameter_list

col_storage_option:
  PLAIN
  {
    $$.val = tree.StoragePlain
  }
| EXTERNAL
  {
    $$.val = tree.StorageExternal
  }
| EXTENDED
  {
    $$.val = tree.StorageExtended
  }
| MAIN
  {
    $$.val = tree.StorageMain
  }

alter_column_set_seq_elem_list:
  alter_column_set_seq_elem
  {
    $$.val = []tree.AlterColComputed{$1.alterColComputed()}
  }
| alter_column_set_seq_elem_list alter_column_set_seq_elem
  {
    $$.val = append($1.alterColComputedList(), $2.alterColComputed())
  }

alter_column_set_seq_elem:
  SET GENERATED_ALWAYS ALWAYS
  {
    $$.val = tree.AlterColComputed{}
  }
| SET GENERATED BY DEFAULT
  {
    $$.val = tree.AlterColComputed{ByDefault: true}
  }
| SET create_seq_option_elem
  {
    $$.val = tree.AlterColComputed{Options: tree.SequenceOptions{$2.seqOpt()}}
  }
| RESTART opt_restart
  {
    $$.val = tree.AlterColComputed{IsRestart: true, Restart: $2.expr()}
  }

opt_restart:
  opt_with signed_iconst
  {
    $$.val = $2.expr()
  }

opt_if_exists:
  /* EMPTY */
  {
    $$.val = false
  }
| IF EXISTS
  {
    $$.val = true
  }

alter_index_cmd:
  set_tablespace
  {
    $$.val = &tree.AlterIndexSetTablespace{Tablespace: $1}
  }
| SET '(' storage_parameter_list ')'
  {
    $$.val = &tree.AlterIndexSetStorage{Params: $3.storageParams()}
  }
| RESET '(' storage_parameter_list ')'
  {
    $$.val = &tree.AlterIndexSetStorage{IsReset: true, Params: $3.storageParams()}
  }
| ALTER opt_column iconst32 SET STATISTICS iconst32
  {
    $$.val = &tree.AlterIndexSetStatistics{ColumnIdx: $3.expr(), Stats: $6.expr()}
  }

alter_column_default:
  SET DEFAULT a_expr
  {
    $$.val = $3.expr()
  }
| DROP DEFAULT
  {
    $$.val = nil
  }

opt_alter_column_using:
  USING a_expr
  {
     $$.val = $2.expr()
  }
| /* EMPTY */
  {
     $$.val = nil
  }

opt_drop_behavior:
  CASCADE
  {
    $$.val = tree.DropCascade
  }
| RESTRICT
  {
    $$.val = tree.DropRestrict
  }
| /* EMPTY */
  {
    $$.val = tree.DropDefault
  }

opt_validate_behavior:
  NOT VALID
  {
    $$.val = tree.ValidationSkip
  }
| /* EMPTY */
  {
    $$.val = tree.ValidationDefault
  }

// %Help: ALTER TYPE - change the definition of a type.
// %Category: DDL
// %Text: ALTER TYPE <typename> <command>
//
// Commands:
//   ALTER TYPE ... ADD VALUE [IF NOT EXISTS] <value> [ { BEFORE | AFTER } <value> ]
//   ALTER TYPE ... RENAME VALUE <oldname> TO <newname>
//   ALTER TYPE ... RENAME TO <newname>
//   ALTER TYPE ... SET SCHEMA <newschemaname>
//   ALTER TYPE ... OWNER TO {<newowner> | CURRENT_ROLE | CURRENT_USER | SESSION_USER }
//   ALTER TYPE ... RENAME ATTRIBUTE <oldname> TO <newname> [ CASCADE | RESTRICT ]
//   ALTER TYPE ... <attributeaction> [, ... ]
//
// Attribute action:
//   ADD ATTRIBUTE <name> <type> [ COLLATE <collation> ] [ CASCADE | RESTRICT ]
//   DROP ATTRIBUTE [IF EXISTS] <name> [ CASCADE | RESTRICT ]
//   ALTER ATTRIBUTE <name> [ SET DATA ] TYPE <type> [ COLLATE <collation> ] [ CASCADE | RESTRICT ]
//
// %SeeAlso: WEBDOCS/alter-type.html
alter_type_stmt:
  ALTER TYPE type_name owner_to
  {
    $$.val = &tree.AlterType{
      Type: $3.unresolvedObjectName(),
      Cmd: &tree.AlterTypeOwner{
        Owner: $4,
      },
    }
  }
| ALTER TYPE type_name RENAME TO name
  {
    $$.val = &tree.AlterType{
      Type: $3.unresolvedObjectName(),
      Cmd: &tree.AlterTypeRename{
        NewName: $6,
      },
    }
  }
| ALTER TYPE type_name set_schema
  {
    $$.val = &tree.AlterType{
      Type: $3.unresolvedObjectName(),
      Cmd: &tree.AlterTypeSetSchema{
        Schema: $4,
      },
    }
  }
| ALTER TYPE type_name RENAME ATTRIBUTE column_name TO column_name opt_drop_behavior
  {
    $$.val = &tree.AlterType{
      Type: $3.unresolvedObjectName(),
      Cmd: &tree.AlterTypeRenameAttribute{
        ColName: tree.Name($6),
        NewColName: tree.Name($8),
        DropBehavior: $9.dropBehavior(),
      },
    }
  }
| ALTER TYPE type_name alter_attribute_action_list
  {
    $$.val = &tree.AlterType{
      Type: $3.unresolvedObjectName(),
      Cmd: &tree.AlterTypeAlterAttribute{
        Actions: $4.alterAttributeActions(),
      },
    }
  }
| ALTER TYPE type_name ADD VALUE SCONST opt_add_val_placement
  {
    $$.val = &tree.AlterType{
      Type: $3.unresolvedObjectName(),
      Cmd: &tree.AlterTypeAddValue{
        NewVal: $6,
        IfNotExists: false,
        Placement: $7.alterTypeAddValuePlacement(),
      },
    }
  }
| ALTER TYPE type_name ADD VALUE IF NOT EXISTS SCONST opt_add_val_placement
  {
    $$.val = &tree.AlterType{
      Type: $3.unresolvedObjectName(),
      Cmd: &tree.AlterTypeAddValue{
        NewVal: $9,
        IfNotExists: true,
        Placement: $10.alterTypeAddValuePlacement(),
      },
    }
  }
| ALTER TYPE type_name RENAME VALUE SCONST TO SCONST
  {
    $$.val = &tree.AlterType{
      Type: $3.unresolvedObjectName(),
      Cmd: &tree.AlterTypeRenameValue{
        OldVal: $6,
        NewVal: $8,
      },
    }
  }
| ALTER TYPE type_name SET '(' type_property_list ')'
  {
    $$.val = &tree.AlterType{
      Type: $3.unresolvedObjectName(),
      Cmd: &tree.AlterTypeSetProperty{
        Properties: $6.baseTypeOptions(),
      },
    }
  }

alter_attribute_action_list:
  alter_attribute_action
  {
    $$.val = []tree.AlterAttributeAction{$1.alterAttributeAction()}
  }
| alter_attribute_action_list ',' alter_attribute_action
  {
    $$.val = append($1.alterAttributeActions(), $3.alterAttributeAction())
  }

alter_attribute_action:
  ADD ATTRIBUTE column_name type_name opt_collate opt_drop_behavior
  {
    $$.val = tree.AlterAttributeAction{
      Action: "add",
      ColName: tree.Name($3),
      TypeName: $4.typeReference(),
      Collate: $5.unresolvedObjectName().UnquotedString(),
      DropBehavior: $6.dropBehavior(),
    }
  }
| DROP ATTRIBUTE column_name opt_drop_behavior
  {
    $$.val = tree.AlterAttributeAction{
      Action: "drop",
      ColName: tree.Name($3),
      DropBehavior: $4.dropBehavior(),
    }
  }
| DROP ATTRIBUTE IF EXISTS column_name opt_drop_behavior
  {
    $$.val = tree.AlterAttributeAction{
      Action: "drop",
      ColName: tree.Name($3),
      IfExists: true,
      DropBehavior: $6.dropBehavior(),
    }
  }
| ALTER ATTRIBUTE column_name TYPE type_name opt_collate opt_drop_behavior
  {
    $$.val = tree.AlterAttributeAction{
      Action: "alter",
      ColName: tree.Name($3),
      TypeName: $5.typeReference(),
      Collate: $6.unresolvedObjectName().UnquotedString(),
      DropBehavior: $7.dropBehavior(),
    }
  }
| ALTER ATTRIBUTE column_name SET DATA TYPE type_name opt_collate opt_drop_behavior
  {
    $$.val = tree.AlterAttributeAction{
      Action: "alter",
      ColName: tree.Name($3),
      TypeName: $7.typeReference(),
      Collate: $8.unresolvedObjectName().UnquotedString(),
      DropBehavior: $9.dropBehavior(),
    }
  }

type_property_list:
  type_property
  {
    $$.val = []tree.BaseTypeOption{$1.baseTypeOption()}
  }
| type_property_list ',' type_property
  {
    $$.val = append($1.baseTypeOptions(), $3.baseTypeOption())
  }

set_schema:
  SET SCHEMA schema_name
  {
    $$ = $3
  }

set_tablespace:
  SET TABLESPACE tablespace_name
  {
    $$ = $3
  }

opt_add_val_placement:
  BEFORE SCONST
  {
    $$.val = &tree.AlterTypeAddValuePlacement{
       Before: true,
       ExistingVal: $2,
    }
  }
| AFTER SCONST
  {
    $$.val = &tree.AlterTypeAddValuePlacement{
       Before: false,
       ExistingVal: $2,
    }
  }
| /* EMPTY */
  {
    $$.val = (*tree.AlterTypeAddValuePlacement)(nil)
  }

opt_owned_by_list:
  /* EMPTY */
  {
    $$.val = []string(nil)
  }
| OWNED BY role_spec_list
  {
    $$.val = $3.strs()
  }

role_spec_list:
  role_spec
  {
    $$.val = []string{$1}
  }
| role_spec_list ',' role_spec
  {
    $$.val = append($1.strs(), $3)
  }

role_spec:
  non_reserved_word_or_sconst
| CURRENT_ROLE
| CURRENT_USER
| SESSION_USER

owner_to:
  OWNER TO role_spec
  {
    $$ = $3
  }

alter_trigger_stmt:
  ALTER TRIGGER trigger_name ON table_name RENAME TO trigger_name
  {
    $$.val = &tree.AlterTrigger{Name: tree.Name($3), OnTable: $5.unresolvedObjectName().ToTableName(), NewName: tree.Name($8)}
  }
| ALTER TRIGGER trigger_name ON table_name opt_no DEPENDS ON EXTENSION name
  {
    $$.val = &tree.AlterTrigger{Name: tree.Name($3), OnTable: $5.unresolvedObjectName().ToTableName(), No: $6.bool(), Extension: $10}
  }

alter_aggregate_stmt:
  ALTER AGGREGATE aggregate_name '(' aggregate_signature ')' RENAME TO unrestricted_name
  {
    $$.val = &tree.AlterAggregate{Name: $3.unresolvedObjectName(), AggSig: $5.aggregateSignature(), Rename: tree.Name($9)}
  }
| ALTER AGGREGATE aggregate_name '(' aggregate_signature ')' owner_to
  {
    $$.val = &tree.AlterAggregate{Name: $3.unresolvedObjectName(), AggSig: $5.aggregateSignature(), Owner: $7}
  }
| ALTER AGGREGATE aggregate_name '(' aggregate_signature ')' set_schema
  {
    $$.val = &tree.AlterAggregate{Name: $3.unresolvedObjectName(), AggSig: $5.aggregateSignature(), Schema: $7}
  }

aggregate_signature:
  '*'
  {
    $$.val = &tree.AggregateSignature{All: true}
  }
| routine_arg_list
  {
    $$.val = &tree.AggregateSignature{Args: $1.routineArgs()}
  }
| routine_arg_list ORDER BY routine_arg_list
  {
    $$.val = &tree.AggregateSignature{Args: $1.routineArgs(), OrderByArgs: $4.routineArgs()}
  }
| ORDER BY routine_arg_list
  {
    $$.val = &tree.AggregateSignature{OrderByArgs: $3.routineArgs()}
  }

opt_routine_args_with_paren:
  /* EMPTY */
  {
    $$.val = []*tree.RoutineArg(nil)
  }
| '(' opt_routine_args ')'
  {
    $$.val = $2.routineArgs()
  }

opt_routine_args:
  /* EMPTY */
  {
    $$.val = []*tree.RoutineArg(nil)
  }
| routine_arg_list
  {
    $$.val = $1.routineArgs()
  }

routine_arg_list:
  routine_arg
  {
    $$.val = []*tree.RoutineArg{$1.routineArg()}
  }
| routine_arg_list ',' routine_arg
  {
    $$.val = append($1.routineArgs(), $3.routineArg())
  }

routine_arg:
  typename
  {
    $$.val = &tree.RoutineArg{Mode: tree.RoutineArgModeIn, Type: $1.typeReference()}
  }
| type_function_name typename
  {
    $$.val = &tree.RoutineArg{Mode: tree.RoutineArgModeIn, Name: tree.Name($1), Type: $2.typeReference()}
  }
| IN typename
  {
    $$.val = &tree.RoutineArg{Mode: tree.RoutineArgModeIn, Type: $2.typeReference()}
  }
| IN type_function_name typename
  {
    $$.val = &tree.RoutineArg{Mode: tree.RoutineArgModeIn, Name: tree.Name($2), Type: $3.typeReference()}
  }
| VARIADIC typename
  {
    $$.val = &tree.RoutineArg{Mode: tree.RoutineArgModeVariadic, Type: $2.typeReference()}
  }
| VARIADIC type_function_name typename
  {
    $$.val = &tree.RoutineArg{Mode: tree.RoutineArgModeVariadic, Name: tree.Name($2), Type: $3.typeReference()}
  }
| OUT typename
  {
    $$.val = &tree.RoutineArg{Mode: tree.RoutineArgModeOut, Type: $2.typeReference()}
  }
| OUT type_function_name typename
  {
    $$.val = &tree.RoutineArg{Mode: tree.RoutineArgModeOut, Name: tree.Name($2), Type: $3.typeReference()}
  }
| INOUT typename
  {
    $$.val = &tree.RoutineArg{Mode: tree.RoutineArgModeInout, Type: $2.typeReference()}
  }
| INOUT type_function_name typename
  {
    $$.val = &tree.RoutineArg{Mode: tree.RoutineArgModeInout, Name: tree.Name($2), Type: $3.typeReference()}
  }

alter_collation_stmt:
  ALTER COLLATION unrestricted_name REFRESH VERSION
  {
    $$.val = &tree.AlterCollation{Name: tree.Name($3), RefreshVersion: true}
  }
| ALTER COLLATION unrestricted_name RENAME TO unrestricted_name
  {
    $$.val = &tree.AlterCollation{Name: tree.Name($3), Rename: tree.Name($6)}
  }
| ALTER COLLATION unrestricted_name owner_to
  {
    $$.val = &tree.AlterCollation{Name: tree.Name($3), Owner: $4}
  }
| ALTER COLLATION unrestricted_name set_schema
  {
    $$.val = &tree.AlterCollation{Name: tree.Name($3), Schema: $4}
  }

alter_conversion_stmt:
  ALTER CONVERSION unrestricted_name RENAME TO unrestricted_name
  {
    $$.val = &tree.AlterConversion{Name: tree.Name($3), Rename: tree.Name($6)}
  }
| ALTER CONVERSION unrestricted_name owner_to
  {
    $$.val = &tree.AlterConversion{Name: tree.Name($3), Owner: $4}
  }
| ALTER CONVERSION unrestricted_name set_schema
  {
    $$.val = &tree.AlterConversion{Name: tree.Name($3), Schema: $4}
  }

// %Help: REFRESH - recalculate a materialized view
// %Category: Misc
// %Text:
// REFRESH MATERIALIZED VIEW [CONCURRENTLY] view_name [WITH [NO] DATA]
refresh_stmt:
  REFRESH MATERIALIZED VIEW opt_concurrently view_name opt_clear_data
  {
    $$.val = &tree.RefreshMaterializedView{
      Name: $5.unresolvedObjectName(),
      Concurrently: $4.bool(),
      RefreshDataOption: $6.refreshDataOption(),
    }
  }
| REFRESH error // SHOW HELP: REFRESH

opt_clear_data:
  WITH DATA
  {
    $$.val = tree.RefreshDataWithData
  }
| WITH NO DATA
  {
    $$.val = tree.RefreshDataClear
  }
| /* EMPTY */
  {
    $$.val = tree.RefreshDataDefault
  }

// %Help: BACKUP - back up data to external storage
// %Category: CCL
// %Text:
// BACKUP <targets...> TO <location...>
//        [ AS OF SYSTEM TIME <expr> ]
//        [ INCREMENTAL FROM <location...> ]
//        [ WITH <option> [= <value>] [, ...] ]
//
// Targets:
//    Empty targets list: backup full cluster.
//    TABLE <pattern> [, ...]
//    DATABASE <databasename> [, ...]
//
// Location:
//    "[scheme]://[host]/[path to backup]?[parameters]"
//
// Options:
//    revision_history: enable revision history
//    encryption_passphrase="secret": encrypt backups
//    kms="[kms_provider]://[kms_host]/[master_key_identifier]?[parameters]" : encrypt backups using KMS
//    detached: execute backup job asynchronously, without waiting for its completion
//
// %SeeAlso: RESTORE, WEBDOCS/backup.html
backup_stmt:
  BACKUP opt_backup_targets INTO sconst_or_placeholder IN string_or_placeholder_opt_list opt_as_of_clause opt_with_backup_options
  {
    $$.val = &tree.Backup{
      Targets: $2.targetListPtr(),
      To: $6.stringOrPlaceholderOptList(),
      Nested: true,
      AppendToLatest: false,
      Subdir: $4.expr(),
      AsOf: $7.asOfClause(),
      Options: *$8.backupOptions(),
    }
  }
| BACKUP opt_backup_targets INTO string_or_placeholder_opt_list opt_as_of_clause opt_with_backup_options
  {
    $$.val = &tree.Backup{
      Targets: $2.targetListPtr(),
      To: $4.stringOrPlaceholderOptList(),
      Nested: true,
      AsOf: $5.asOfClause(),
      Options: *$6.backupOptions(),
    }
  }
| BACKUP opt_backup_targets INTO LATEST IN string_or_placeholder_opt_list opt_as_of_clause opt_with_backup_options
  {
    $$.val = &tree.Backup{
      Targets: $2.targetListPtr(),
      To: $6.stringOrPlaceholderOptList(),
      Nested: true,
      AppendToLatest: true,
      AsOf: $7.asOfClause(),
      Options: *$8.backupOptions(),
    }
  }
| BACKUP opt_backup_targets TO string_or_placeholder_opt_list opt_as_of_clause opt_incremental opt_with_backup_options
  {
    $$.val = &tree.Backup{
      Targets: $2.targetListPtr(),
      To: $4.stringOrPlaceholderOptList(),
      IncrementalFrom: $6.exprs(),
      AsOf: $5.asOfClause(),
      Options: *$7.backupOptions(),
    }
  }
| BACKUP error // SHOW HELP: BACKUP

opt_backup_targets:
  /* EMPTY -- full cluster */
  {
    $$.val = (*tree.TargetList)(nil)
  }
| targets_table
  {
    t := $1.targetList()
    $$.val = &t
  }

// Optional backup options.
opt_with_backup_options:
  WITH backup_options_list
  {
    $$.val = $2.backupOptions()
  }
| WITH OPTIONS '(' backup_options_list ')'
  {
    $$.val = $4.backupOptions()
  }
| /* EMPTY */
  {
    $$.val = &tree.BackupOptions{}
  }

backup_options_list:
  // Require at least one option
  backup_options
  {
    $$.val = $1.backupOptions()
  }
| backup_options_list ',' backup_options
  {
    if err := $1.backupOptions().CombineWith($3.backupOptions()); err != nil {
      return setErr(sqllex, err)
    }
  }

// List of valid backup options.
backup_options:
  ENCRYPTION_PASSPHRASE '=' string_or_placeholder
  {
    $$.val = &tree.BackupOptions{EncryptionPassphrase: $3.expr()}
  }
| REVISION_HISTORY
  {
    $$.val = &tree.BackupOptions{CaptureRevisionHistory: true}
  }
| DETACHED
  {
    $$.val = &tree.BackupOptions{Detached: true}
  }
| KMS '=' string_or_placeholder_opt_list
	{
    $$.val = &tree.BackupOptions{EncryptionKMSURI: $3.stringOrPlaceholderOptList()}
	}
// %Help: CREATE SCHEDULE FOR BACKUP - backup data periodically
// %Category: CCL
// %Text:
// CREATE SCHEDULE [<description>]
// FOR BACKUP [<targets>] TO <location...>
// [WITH <backup_option>[=<value>] [, ...]]
// RECURRING [crontab|NEVER] [FULL BACKUP <crontab|ALWAYS>]
// [WITH EXPERIMENTAL SCHEDULE OPTIONS <schedule_option>[= <value>] [, ...] ]
//
// All backups run in UTC timezone.
//
// Description:
//   Optional description (or name) for this schedule
//
// Targets:
//   empty targets: Backup entire cluster
//   DATABASE <pattern> [, ...]: comma separated list of databases to backup.
//   TABLE <pattern> [, ...]: comma separated list of tables to backup.
//
// Location:
//   "[scheme]://[host]/[path prefix to backup]?[parameters]"
//   Backup schedule will create subdirectories under this location to store
//   full and periodic backups.
//
// WITH <options>:
//   Options specific to BACKUP: See BACKUP options
//
// RECURRING <crontab>:
//   The RECURRING expression specifies when we backup.  By default these are incremental
//   backups that capture changes since the last backup, writing to the "current" backup.
//
//   Schedule specified as a string in crontab format.
//   All times in UTC.
//     "5 0 * * *": run schedule 5 minutes past midnight.
//     "@daily": run daily, at midnight
//   See https://en.wikipedia.org/wiki/Cron
//
// FULL BACKUP <crontab|ALWAYS>:
//   The optional FULL BACKUP '<cron expr>' clause specifies when we'll start a new full backup,
//   which becomes the "current" backup when complete.
//   If FULL BACKUP ALWAYS is specified, then the backups triggered by the RECURRING clause will
//   always be full backups. For free users, this is the only accepted value of FULL BACKUP.
//
//   If the FULL BACKUP clause is omitted, we will select a reasonable default:
//      * RECURRING <= 1 hour: we default to FULL BACKUP '@daily';
//      * RECURRING <= 1 day:  we default to FULL BACKUP '@weekly';
//      * Otherwise: we default to FULL BACKUP ALWAYS.
//
//  SCHEDULE OPTIONS:
//   The schedule can be modified by specifying the following options (which are considered
//   to be experimental at this time):
//   * first_run=TIMESTAMPTZ:
//     execute the schedule at the specified time. If not specified, the default is to execute
//     the scheduled based on it's next RECURRING time.
//   * on_execution_failure='[retry|reschedule|pause]':
//     If an error occurs during the execution, handle the error based as:
//     * retry: retry execution right away
//     * reschedule: retry execution by rescheduling it based on its RECURRING expression.
//       This is the default.
//     * pause: pause this schedule.  Requires manual intervention to unpause.
//   * on_previous_running='[start|skip|wait]':
//     If the previous backup started by this schedule still running, handle this as:
//     * start: start this execution anyway, even if the previous one still running.
//     * skip: skip this execution, reschedule it based on RECURRING (or change_capture_period)
//       expression.
//     * wait: wait for the previous execution to complete.  This is the default.
//   * ignore_existing_backups
//     If backups were already created in the destination in which a new schedule references,
//     this flag must be passed in to acknowledge that the new schedule may be backing up different
//     objects.
//
// %SeeAlso: BACKUP
create_schedule_for_backup_stmt:
  CREATE SCHEDULE /*$3=*/opt_description FOR BACKUP /*$6=*/opt_backup_targets INTO
  /*$8=*/string_or_placeholder_opt_list /*$9=*/opt_with_backup_options
  /*$10=*/cron_expr /*$11=*/opt_full_backup_clause /*$12=*/opt_with_schedule_options
  {
    $$.val = &tree.ScheduledBackup{
      ScheduleLabel:    $3.expr(),
      Recurrence:       $10.expr(),
      FullBackup:       $11.fullBackupClause(),
      To:               $8.stringOrPlaceholderOptList(),
      Targets:          $6.targetListPtr(),
      BackupOptions:    *($9.backupOptions()),
      ScheduleOptions:  $12.kvOptions(),
    }
  }
| CREATE SCHEDULE error  // SHOW HELP: CREATE SCHEDULE FOR BACKUP

opt_description:
  string_or_placeholder
| /* EMPTY */
  {
     $$.val = nil
  }


// sconst_or_placeholder matches a simple string, or a placeholder.
sconst_or_placeholder:
  SCONST
  {
    $$.val =  tree.NewStrVal($1)
  }
| PLACEHOLDER
  {
    p := $1.placeholder()
    sqllex.(*lexer).UpdateNumPlaceholders(p)
    $$.val = p
  }

cron_expr:
  RECURRING sconst_or_placeholder
  // Can't use string_or_placeholder here due to conflict on NEVER branch above
  // (is NEVER a keyword or a variable?).
  {
    $$.val = $2.expr()
  }

opt_full_backup_clause:
  FULL BACKUP sconst_or_placeholder
  // Can't use string_or_placeholder here due to conflict on ALWAYS branch below
  // (is ALWAYS a keyword or a variable?).
  {
    $$.val = &tree.FullBackupClause{Recurrence: $3.expr()}
  }
| FULL BACKUP ALWAYS
  {
    $$.val = &tree.FullBackupClause{AlwaysFull: true}
  }
| /* EMPTY */
  {
    $$.val = (*tree.FullBackupClause)(nil)
  }

opt_with_schedule_options:
  WITH SCHEDULE OPTIONS kv_option_list
  {
    $$.val = $4.kvOptions()
  }
| WITH SCHEDULE OPTIONS '(' kv_option_list ')'
  {
    $$.val = $5.kvOptions()
  }
| /* EMPTY */
  {
    $$.val = nil
  }


// %Help: RESTORE - restore data from external storage
// %Category: CCL
// %Text:
// RESTORE <targets...> FROM <location...>
//         [ AS OF SYSTEM TIME <expr> ]
//         [ WITH <option> [= <value>] [, ...] ]
//
// Targets:
//    TABLE <pattern> [, ...]
//    DATABASE <databasename> [, ...]
//
// Locations:
//    "[scheme]://[host]/[path to backup]?[parameters]"
//
// Options:
//    into_db: specify target database
//    skip_missing_foreign_keys: remove foreign key constraints before restoring
//    skip_missing_sequences: ignore sequence dependencies
//    skip_missing_views: skip restoring views because of dependencies that cannot be restored
//    skip_missing_sequence_owners: remove sequence-table ownership dependencies before restoring
//    encryption_passphrase=passphrase: decrypt BACKUP with specified passphrase
//    kms="[kms_provider]://[kms_host]/[master_key_identifier]?[parameters]" : decrypt backups using KMS
//    detached: execute restore job asynchronously, without waiting for its completion
// %SeeAlso: BACKUP, WEBDOCS/restore.html
restore_stmt:
  RESTORE FROM list_of_string_or_placeholder_opt_list opt_as_of_clause opt_with_restore_options
  {
    $$.val = &tree.Restore{
    DescriptorCoverage: tree.AllDescriptors,
    From: $3.listOfStringOrPlaceholderOptList(),
    AsOf: $4.asOfClause(),
    Options: *($5.restoreOptions()),
    }
  }
| RESTORE targets_table FROM list_of_string_or_placeholder_opt_list opt_as_of_clause opt_with_restore_options
  {
    $$.val = &tree.Restore{
    Targets: $2.targetList(),
    From: $4.listOfStringOrPlaceholderOptList(),
    AsOf: $5.asOfClause(),
    Options: *($6.restoreOptions()),
    }
  }
| RESTORE targets_table FROM string_or_placeholder IN list_of_string_or_placeholder_opt_list opt_as_of_clause opt_with_restore_options
  {
    $$.val = &tree.Restore{
      Targets: $2.targetList(),
      Subdir: $4.expr(),
      From: $6.listOfStringOrPlaceholderOptList(),
      AsOf: $7.asOfClause(),
      Options: *($8.restoreOptions()),
    }
  }
| RESTORE error // SHOW HELP: RESTORE

string_or_placeholder_opt_list:
  string_or_placeholder
  {
    $$.val = tree.StringOrPlaceholderOptList{$1.expr()}
  }
| '(' string_or_placeholder_list ')'
  {
    $$.val = tree.StringOrPlaceholderOptList($2.exprs())
  }

list_of_string_or_placeholder_opt_list:
  string_or_placeholder_opt_list
  {
    $$.val = []tree.StringOrPlaceholderOptList{$1.stringOrPlaceholderOptList()}
  }
| list_of_string_or_placeholder_opt_list ',' string_or_placeholder_opt_list
  {
    $$.val = append($1.listOfStringOrPlaceholderOptList(), $3.stringOrPlaceholderOptList())
  }

// Optional restore options.
opt_with_restore_options:
  WITH restore_options_list
  {
    $$.val = $2.restoreOptions()
  }
| WITH OPTIONS '(' restore_options_list ')'
  {
    $$.val = $4.restoreOptions()
  }
| /* EMPTY */
  {
    $$.val = &tree.RestoreOptions{}
  }

restore_options_list:
  // Require at least one option
  restore_options
  {
    $$.val = $1.restoreOptions()
  }
| restore_options_list ',' restore_options
  {
    if err := $1.restoreOptions().CombineWith($3.restoreOptions()); err != nil {
      return setErr(sqllex, err)
    }
  }

// List of valid restore options.
restore_options:
  ENCRYPTION_PASSPHRASE '=' string_or_placeholder
  {
    $$.val = &tree.RestoreOptions{EncryptionPassphrase: $3.expr()}
  }
| KMS '=' string_or_placeholder_opt_list
	{
    $$.val = &tree.RestoreOptions{DecryptionKMSURI: $3.stringOrPlaceholderOptList()}
	}
| INTO_DB '=' string_or_placeholder
  {
    $$.val = &tree.RestoreOptions{IntoDB: $3.expr()}
  }
| SKIP_MISSING_FOREIGN_KEYS
  {
    $$.val = &tree.RestoreOptions{SkipMissingFKs: true}
  }
| SKIP_MISSING_SEQUENCES
  {
    $$.val = &tree.RestoreOptions{SkipMissingSequences: true}
  }
| SKIP_MISSING_SEQUENCE_OWNERS
  {
    $$.val = &tree.RestoreOptions{SkipMissingSequenceOwners: true}
  }
| SKIP_MISSING_VIEWS
  {
    $$.val = &tree.RestoreOptions{SkipMissingViews: true}
  }
| DETACHED
  {
    $$.val = &tree.RestoreOptions{Detached: true}
  }

import_format:
  name
  {
    $$ = strings.ToUpper($1)
  }

// %Help: IMPORT - load data from file in a distributed manner
// %Category: CCL
// %Text:
// -- Import both schema and table data:
// IMPORT [ TABLE <tablename> FROM ]
//        <format> <datafile>
//        [ WITH <option> [= <value>] [, ...] ]
//
// -- Import using specific schema, use only table data from external file:
// IMPORT TABLE <tablename>
//        { ( <elements> ) | CREATE USING <schemafile> }
//        <format>
//        DATA ( <datafile> [, ...] )
//        [ WITH <option> [= <value>] [, ...] ]
//
// Formats:
//    CSV
//    DELIMITED
//    MYSQLDUMP
//    PGCOPY
//    PGDUMP
//
// Options:
//    distributed = '...'
//    sstsize = '...'
//    temp = '...'
//    delimiter = '...'      [CSV, PGCOPY-specific]
//    nullif = '...'         [CSV, PGCOPY-specific]
//    comment = '...'        [CSV-specific]
//
// %SeeAlso: CREATE TABLE
import_stmt:
 IMPORT import_format '(' string_or_placeholder ')' opt_with_options
  {
    /* SKIP DOC */
    $$.val = &tree.Import{Bundle: true, FileFormat: $2, Files: tree.Exprs{$4.expr()}, Options: $6.kvOptions()}
  }
| IMPORT import_format string_or_placeholder opt_with_options
  {
    $$.val = &tree.Import{Bundle: true, FileFormat: $2, Files: tree.Exprs{$3.expr()}, Options: $4.kvOptions()}
  }
| IMPORT TABLE table_name FROM import_format '(' string_or_placeholder ')' opt_with_options
  {
    /* SKIP DOC */
    name := $3.unresolvedObjectName().ToTableName()
    $$.val = &tree.Import{Bundle: true, Table: &name, FileFormat: $5, Files: tree.Exprs{$7.expr()}, Options: $9.kvOptions()}
  }
| IMPORT TABLE table_name FROM import_format string_or_placeholder opt_with_options
  {
    name := $3.unresolvedObjectName().ToTableName()
    $$.val = &tree.Import{Bundle: true, Table: &name, FileFormat: $5, Files: tree.Exprs{$6.expr()}, Options: $7.kvOptions()}
  }
| IMPORT TABLE table_name CREATE USING string_or_placeholder import_format DATA '(' string_or_placeholder_list ')' opt_with_options
  {
    name := $3.unresolvedObjectName().ToTableName()
    $$.val = &tree.Import{Table: &name, CreateFile: $6.expr(), FileFormat: $7, Files: $10.exprs(), Options: $12.kvOptions()}
  }
| IMPORT TABLE table_name '(' table_elem_list ')' import_format DATA '(' string_or_placeholder_list ')' opt_with_options
  {
    name := $3.unresolvedObjectName().ToTableName()
    $$.val = &tree.Import{Table: &name, CreateDefs: $5.tblDefs(), FileFormat: $7, Files: $10.exprs(), Options: $12.kvOptions()}
  }
| IMPORT INTO table_name '(' insert_column_list ')' import_format DATA '(' string_or_placeholder_list ')' opt_with_options
  {
    name := $3.unresolvedObjectName().ToTableName()
    $$.val = &tree.Import{Table: &name, Into: true, IntoCols: $5.nameList(), FileFormat: $7, Files: $10.exprs(), Options: $12.kvOptions()}
  }
| IMPORT INTO table_name import_format DATA '(' string_or_placeholder_list ')' opt_with_options
  {
    name := $3.unresolvedObjectName().ToTableName()
    $$.val = &tree.Import{Table: &name, Into: true, IntoCols: nil, FileFormat: $4, Files: $7.exprs(), Options: $9.kvOptions()}
  }
| IMPORT error // SHOW HELP: IMPORT

// %Help: EXPORT - export data to file in a distributed manner
// %Category: CCL
// %Text:
// EXPORT INTO <format> <datafile> [WITH <option> [= value] [,...]] FROM <query>
//
// Formats:
//    CSV
//
// Options:
//    delimiter = '...'   [CSV-specific]
//
// %SeeAlso: SELECT
export_stmt:
  EXPORT INTO import_format string_or_placeholder opt_with_options FROM select_stmt
  {
    $$.val = &tree.Export{Query: $7.slct(), FileFormat: $3, File: $4.expr(), Options: $5.kvOptions()}
  }
| EXPORT error // SHOW HELP: EXPORT

string_or_placeholder:
  non_reserved_word_or_sconst
  {
    $$.val = tree.NewStrVal($1)
  }
| PLACEHOLDER
  {
    p := $1.placeholder()
    sqllex.(*lexer).UpdateNumPlaceholders(p)
    $$.val = p
  }

string_or_placeholder_list:
  string_or_placeholder
  {
    $$.val = tree.Exprs{$1.expr()}
  }
| string_or_placeholder_list ',' string_or_placeholder
  {
    $$.val = append($1.exprs(), $3.expr())
  }

opt_incremental:
  INCREMENTAL FROM string_or_placeholder_list
  {
    $$.val = $3.exprs()
  }
| /* EMPTY */
  {
    $$.val = tree.Exprs(nil)
  }

kv_option:
  name '=' string_or_placeholder
  {
    $$.val = tree.KVOption{Key: tree.Name($1), Value: $3.expr()}
  }
|  name
  {
    $$.val = tree.KVOption{Key: tree.Name($1)}
  }
|  SCONST '=' string_or_placeholder
  {
    $$.val = tree.KVOption{Key: tree.Name($1), Value: $3.expr()}
  }
|  SCONST
  {
    $$.val = tree.KVOption{Key: tree.Name($1)}
  }

kv_option_list:
  kv_option
  {
    $$.val = []tree.KVOption{$1.kvOption()}
  }
|  kv_option_list ',' kv_option
  {
    $$.val = append($1.kvOptions(), $3.kvOption())
  }

opt_with_options:
  WITH kv_option_list
  {
    $$.val = $2.kvOptions()
  }
| WITH OPTIONS '(' kv_option_list ')'
  {
    $$.val = $4.kvOptions()
  }
| /* EMPTY */
  {
    $$.val = nil
  }

// %Help: CALL - invoke a procedure
// %Category: Misc
// %Text: CALL <name> ( [ <expr> [, ...] ] )
// %SeeAlso: CREATE PROCEDURE
call_stmt:
  CALL func_application
  {
    $$.val = &tree.Call{Procedure: $2.expr().(*tree.FuncExpr)}
  }

// The COPY grammar in postgres has 3 different versions, all of which are supported by postgres:
// 1) The "really old" syntax from v7.2 and prior
// 2) Pre 9.0 using hard-wired, space-separated options
// 3) The current and preferred options using comma-separated generic identifiers instead of keywords.
// We currently support only the #2 format.
// See the comment for CopyStmt in https://github.com/postgres/postgres/blob/master/src/backend/parser/gram.y.
copy_from_stmt:
 COPY table_name opt_column_list FROM SCONST opt_with '(' copy_options_list ')'
  {
    name := $2.unresolvedObjectName().ToTableName()
    $$.val = &tree.CopyFrom{
       Table: name,
       File: $5,
       Columns: $3.nameList(),
       Stdin: false,
       Options: *$8.copyOptions(),
    }
  }
| COPY table_name opt_column_list FROM SCONST opt_legacy_copy_options
  {
    name := $2.unresolvedObjectName().ToTableName()
    $$.val = &tree.CopyFrom{
       Table: name,
       File: $5,
       Columns: $3.nameList(),
       Stdin: false,
       Options: *$6.copyOptions(),
    }
  }
| COPY table_name opt_column_list FROM STDIN opt_with '(' copy_options_list ')'
  {
    name := $2.unresolvedObjectName().ToTableName()
    $$.val = &tree.CopyFrom{
       Table: name,
       Columns: $3.nameList(),
       Stdin: true,
       Options: *$8.copyOptions(),
    }
  }
| COPY table_name opt_column_list FROM STDIN opt_legacy_copy_options
  {
    name := $2.unresolvedObjectName().ToTableName()
    $$.val = &tree.CopyFrom{
       Table: name,
       Columns: $3.nameList(),
       Stdin: true,
       Options: *$6.copyOptions(),
    }
  }


// legacy_copy_options represent the previous format that PostgreSQL supported for
// specifying COPY FROM options. They do not use the WITH keyword and do not use parens.
opt_legacy_copy_options:
  legacy_copy_options_list
  {
    $$.val = $1.copyOptions()
  }
| /* EMPTY */
  {
    $$.val = &tree.CopyOptions{}
  }

legacy_copy_options_list:
  legacy_copy_options
  {
    $$.val = $1.copyOptions()
  }
| legacy_copy_options_list ',' legacy_copy_options
  {
    if err := $1.copyOptions().CombineWith($3.copyOptions()); err != nil {
      return setErr(sqllex, err)
    }
  }

legacy_copy_options:
 BINARY
  {
    $$.val = &tree.CopyOptions{CopyFormat: tree.CopyFormatBinary}
  }
| DELIMITER SCONST
  {
    $$.val = &tree.CopyOptions{Delimiter: $2}
  }
| CSV
  {
    $$.val = &tree.CopyOptions{CopyFormat: tree.CopyFormatCsv}
  }
| HEADER
  {
    $$.val = &tree.CopyOptions{Header: true}
  }

copy_options_list:
  copy_options
  {
    $$.val = $1.copyOptions()
  }
| copy_options_list ',' copy_options
  {
    if err := $1.copyOptions().CombineWith($3.copyOptions()); err != nil {
      return setErr(sqllex, err)
    }
  }

copy_options:
 FORMAT CSV
  {
    $$.val = &tree.CopyOptions{CopyFormat: tree.CopyFormatCsv}
  }
| FORMAT TEXT
  {
    $$.val = &tree.CopyOptions{CopyFormat: tree.CopyFormatText}
  }
| FORMAT BINARY
  {
    $$.val = &tree.CopyOptions{CopyFormat: tree.CopyFormatBinary}
  }
| HEADER
  {
    $$.val = &tree.CopyOptions{Header: true}
  }
| HEADER boolean_value
  {
    $$.val = &tree.CopyOptions{Header: $2.val.(bool)}
  }
| DELIMITER SCONST
  {
    $$.val = &tree.CopyOptions{Delimiter: $2}
  }

// %Help: CANCEL
// %Category: Group
// %Text: CANCEL JOBS, CANCEL QUERIES, CANCEL SESSIONS
cancel_stmt:
  cancel_jobs_stmt     // EXTEND WITH HELP: CANCEL JOBS
| cancel_queries_stmt  // EXTEND WITH HELP: CANCEL QUERIES
| cancel_sessions_stmt // EXTEND WITH HELP: CANCEL SESSIONS
| CANCEL error         // SHOW HELP: CANCEL

// %Help: CANCEL JOBS - cancel background jobs
// %Category: Misc
// %Text:
// CANCEL JOBS <selectclause>
// CANCEL JOB <jobid>
// %SeeAlso: SHOW JOBS, PAUSE JOBS, RESUME JOBS
cancel_jobs_stmt:
  CANCEL JOB a_expr
  {
    $$.val = &tree.ControlJobs{
      Jobs: &tree.Select{
        Select: &tree.ValuesClause{Rows: []tree.Exprs{tree.Exprs{$3.expr()}}},
      },
      Command: tree.CancelJob,
    }
  }
| CANCEL JOB error // SHOW HELP: CANCEL JOBS
| CANCEL JOBS select_stmt
  {
    $$.val = &tree.ControlJobs{Jobs: $3.slct(), Command: tree.CancelJob}
  }
| CANCEL JOBS for_schedules_clause
  {
    $$.val = &tree.ControlJobsForSchedules{Schedules: $3.slct(), Command: tree.CancelJob}
  }
| CANCEL JOBS error // SHOW HELP: CANCEL JOBS

// %Help: CANCEL QUERIES - cancel running queries
// %Category: Misc
// %Text:
// CANCEL QUERIES [IF EXISTS] <selectclause>
// CANCEL QUERY [IF EXISTS] <expr>
// %SeeAlso: SHOW QUERIES
cancel_queries_stmt:
  CANCEL QUERY a_expr
  {
    $$.val = &tree.CancelQueries{
      Queries: &tree.Select{
        Select: &tree.ValuesClause{Rows: []tree.Exprs{tree.Exprs{$3.expr()}}},
      },
      IfExists: false,
    }
  }
| CANCEL QUERY IF EXISTS a_expr
  {
    $$.val = &tree.CancelQueries{
      Queries: &tree.Select{
        Select: &tree.ValuesClause{Rows: []tree.Exprs{tree.Exprs{$5.expr()}}},
      },
      IfExists: true,
    }
  }
| CANCEL QUERY error // SHOW HELP: CANCEL QUERIES
| CANCEL QUERIES select_stmt
  {
    $$.val = &tree.CancelQueries{Queries: $3.slct(), IfExists: false}
  }
| CANCEL QUERIES IF EXISTS select_stmt
  {
    $$.val = &tree.CancelQueries{Queries: $5.slct(), IfExists: true}
  }
| CANCEL QUERIES error // SHOW HELP: CANCEL QUERIES

// %Help: CANCEL SESSIONS - cancel open sessions
// %Category: Misc
// %Text:
// CANCEL SESSIONS [IF EXISTS] <selectclause>
// CANCEL SESSION [IF EXISTS] <sessionid>
// %SeeAlso: SHOW SESSIONS
cancel_sessions_stmt:
  CANCEL SESSION a_expr
  {
   $$.val = &tree.CancelSessions{
      Sessions: &tree.Select{
        Select: &tree.ValuesClause{Rows: []tree.Exprs{tree.Exprs{$3.expr()}}},
      },
      IfExists: false,
    }
  }
| CANCEL SESSION IF EXISTS a_expr
  {
   $$.val = &tree.CancelSessions{
      Sessions: &tree.Select{
        Select: &tree.ValuesClause{Rows: []tree.Exprs{tree.Exprs{$5.expr()}}},
      },
      IfExists: true,
    }
  }
| CANCEL SESSION error // SHOW HELP: CANCEL SESSIONS
| CANCEL SESSIONS select_stmt
  {
    $$.val = &tree.CancelSessions{Sessions: $3.slct(), IfExists: false}
  }
| CANCEL SESSIONS IF EXISTS select_stmt
  {
    $$.val = &tree.CancelSessions{Sessions: $5.slct(), IfExists: true}
  }
| CANCEL SESSIONS error // SHOW HELP: CANCEL SESSIONS

comment_stmt:
  COMMENT ON ACCESS METHOD db_object_name IS comment_text
  {
    $$.val = &tree.Comment{
      Object: &tree.CommentOnAccessMethod{Name: $5.unresolvedObjectName()},
      Comment: $7.strPtr(),
    }
  }
| COMMENT ON AGGREGATE aggregate_name '(' aggregate_signature ')' IS comment_text
  {
    $$.val = &tree.Comment{
      Object: &tree.CommentOnAggregate{Name: $4.unresolvedObjectName(), AggSig: $6.aggregateSignature()},
      Comment: $9.strPtr(),
    }
  }
| COMMENT ON CAST '(' typename AS typename ')' IS comment_text
  {
    $$.val = &tree.Comment{
      Object: &tree.CommentOnCast{Source: $5.typeReference(), Target: $7.typeReference()},
      Comment: $10.strPtr(),
    }
  }
| COMMENT ON COLLATION collation_name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnCollation{Name: $4.unresolvedObjectName().UnquotedString()}, Comment: $6.strPtr()}
  }
| COMMENT ON COLUMN column_path IS comment_text
  {
    varName, err := $4.unresolvedName().NormalizeVarName()
    if err != nil {
      return setErr(sqllex, err)
    }
    columnItem, ok := varName.(*tree.ColumnItem)
    if !ok {
      sqllex.Error(fmt.Sprintf("invalid column name: %q", tree.ErrString($4.unresolvedName())))
            return 1
    }
    $$.val = &tree.Comment{
      Object: &tree.CommentOnColumn{ColumnItem: columnItem},
      Comment: $6.strPtr(),
    }
  }
| COMMENT ON CONSTRAINT constraint_name ON table_name IS comment_text
  {
    $$.val = &tree.Comment{
      Object: &tree.CommentOnConstraintOnTable{Constraint: tree.Name($4), Table: $6.unresolvedObjectName()},
      Comment: $8.strPtr(),
    }
  }
| COMMENT ON CONSTRAINT constraint_name ON DOMAIN type_name IS comment_text
  {
    $$.val = &tree.Comment{
      Object: &tree.CommentOnConstraintOnDomain{Constraint: tree.Name($4), Domain: $7.unresolvedObjectName()},
      Comment: $9.strPtr(),
    }
  }
| COMMENT ON CONVERSION name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnConversion{Name: tree.Name($4)}, Comment: $6.strPtr()}
  }
| COMMENT ON DATABASE database_name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnDatabase{Name: tree.Name($4)}, Comment: $6.strPtr()}
  }
| COMMENT ON DOMAIN type_name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnDomain{Name: $4.unresolvedObjectName()}, Comment: $6.strPtr()}
  }
| COMMENT ON EXTENSION name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnExtension{Name: tree.Name($4)}, Comment: $6.strPtr()}
  }
| COMMENT ON EVENT TRIGGER name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnEventTrigger{Name: tree.Name($5)}, Comment: $7.strPtr()}
  }
| COMMENT ON FOREIGN DATA WRAPPER name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnForeignDataWrapper{Name: tree.Name($6)}, Comment: $8.strPtr()}
  }
| COMMENT ON FOREIGN TABLE name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnForeignTable{Name: tree.Name($5)}, Comment: $7.strPtr()}
  }
| COMMENT ON FUNCTION routine_name opt_routine_args_with_paren IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnFunction{Name: $4.unresolvedObjectName(), Args: $5.routineArgs()}, Comment: $7.strPtr()}
  }
| COMMENT ON INDEX table_index_name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnIndex{Index: $4.tableIndexName()}, Comment: $6.strPtr()}
  }
| COMMENT ON LARGE OBJECT signed_iconst IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnLargeObject{Oid: $5.expr()}, Comment: $7.strPtr()}
  }
| COMMENT ON MATERIALIZED VIEW relation_expr IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnMaterializedView{Name: $5.unresolvedObjectName()}, Comment: $7.strPtr()}
  }
| COMMENT ON OPERATOR operator '(' typename ',' typename ')' IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnOperator{Op: $4.op(), Left: $6.typeReference(), Right: $8.typeReference()}, Comment: $11.strPtr()}
  }
| COMMENT ON OPERATOR CLASS name USING name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnOperatorClass{Name: tree.Name($5), IdxMethod: $7}, Comment: $9.strPtr()}
  }
| COMMENT ON OPERATOR FAMILY name USING name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnOperatorFamily{Name: tree.Name($5), IdxMethod: $7}, Comment: $9.strPtr()}
  }
| COMMENT ON POLICY name ON table_name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnPolicy{Policy: tree.Name($4), Table: $6.unresolvedObjectName()}, Comment: $8.strPtr()}
  }
| COMMENT ON LANGUAGE name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnLanguage{Name: tree.Name($4)}, Comment: $6.strPtr()}
  }
| COMMENT ON PROCEDURAL LANGUAGE name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnLanguage{Name: tree.Name($5), Procedural: true}, Comment: $7.strPtr()}
  }
| COMMENT ON PROCEDURE routine_name opt_routine_args_with_paren IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnProcedure{Name: $4.unresolvedObjectName(), Args: $5.routineArgs()}, Comment: $7.strPtr()}
  }
| COMMENT ON PUBLICATION name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnPublication{Name: tree.Name($4)}, Comment: $6.strPtr()}
  }
| COMMENT ON ROLE role_spec IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnRole{Name: $4}, Comment: $6.strPtr()}
  }
| COMMENT ON ROUTINE routine_name opt_routine_args_with_paren IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnRoutine{Name: $4.unresolvedObjectName(), Args: $5.routineArgs()}, Comment: $7.strPtr()}
  }
| COMMENT ON RULE name ON table_name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnRule{Rule: tree.Name($4), Table: $6.unresolvedObjectName()}, Comment: $8.strPtr()}
  }
| COMMENT ON SCHEMA schema_name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnSchema{Name: $4}, Comment: $6.strPtr()}
  }
| COMMENT ON SEQUENCE sequence_name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnSequence{Name: $4.unresolvedObjectName()}, Comment: $6.strPtr()}
  }
| COMMENT ON SERVER name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnServer{Name: tree.Name($4)}, Comment: $6.strPtr()}
  }
| COMMENT ON STATISTICS name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnStatistics{Name: tree.Name($4)}, Comment: $6.strPtr()}
  }
| COMMENT ON SUBSCRIPTION name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnSubscription{Name: tree.Name($4)}, Comment: $6.strPtr()}
  }
| COMMENT ON TABLE table_name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnTable{Name: $4.unresolvedObjectName()}, Comment: $6.strPtr()}
  }
| COMMENT ON TABLESPACE tablespace_name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnTablespace{Name: tree.Name($4)}, Comment: $6.strPtr()}
  }
| COMMENT ON TEXT SEARCH CONFIGURATION name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnTextSearchConfiguration{Name: tree.Name($6)}, Comment: $8.strPtr()}
  }
| COMMENT ON TEXT SEARCH DICTIONARY name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnTextSearchDictionary{Name: tree.Name($6)}, Comment: $8.strPtr()}
  }
| COMMENT ON TEXT SEARCH PARSER name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnTextSearchParser{Name: tree.Name($6)}, Comment: $8.strPtr()}
  }
| COMMENT ON TEXT SEARCH TEMPLATE name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnTextSearchTemplate{Name: tree.Name($6)}, Comment: $8.strPtr()}
  }
| COMMENT ON TRANSFORM FOR typename LANGUAGE name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnTransformFor{Type: $5.typeReference(), Language: tree.Name($7)}, Comment: $9.strPtr()}
  }
| COMMENT ON TRIGGER trigger_name ON table_name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnTrigger{Trigger: tree.Name($4), Table: $6.unresolvedObjectName()}, Comment: $8.strPtr()}
  }
| COMMENT ON TYPE type_name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnType{Name: $4.unresolvedObjectName()}, Comment: $6.strPtr()}
  }
| COMMENT ON VIEW view_name IS comment_text
  {
    $$.val = &tree.Comment{Object: &tree.CommentOnView{Name: $4.unresolvedObjectName()}, Comment: $6.strPtr()}
  }

comment_text:
  SCONST
  {
    t := $1
    $$.val = &t
  }
| NULL
  {
    var str *string
    $$.val = str
  }

// %Help: CREATE
// %Category: Group
// %Text:
// CREATE DATABASE, CREATE TABLE, CREATE INDEX, CREATE TABLE AS,
// CREATE USER, CREATE VIEW, CREATE SEQUENCE, CREATE STATISTICS,
// CREATE ROLE, CREATE TYPE
create_stmt:
  create_role_stmt     // EXTEND WITH HELP: CREATE ROLE
| create_ddl_stmt      // help texts in sub-rule
| create_stats_stmt    // EXTEND WITH HELP: CREATE STATISTICS
| create_schedule_for_backup_stmt   // EXTEND WITH HELP: CREATE SCHEDULE FOR BACKUP
| create_function_stmt // EXTEND WITH HELP: CREATE FUNCTION
| create_procedure_stmt // EXTEND WITH HELP: CREATE PROCEDURE
| create_extension_stmt // EXTEND WITH HELP: CREATE EXTENSION
| create_language_stmt  // EXTEND WITH HELP: CREATE LANGUAGE
| create_aggregate_stmt // EXTEND WITH HELP: CREATE AGGREGATE
| create_unsupported   {}
| CREATE error         // SHOW HELP: CREATE

create_unsupported:
  CREATE CAST error { return unimplemented(sqllex, "create cast") }
| CREATE CONVERSION error { return unimplemented(sqllex, "create conversion") }
| CREATE DEFAULT CONVERSION error { return unimplemented(sqllex, "create def conv") }
| CREATE FOREIGN TABLE error { return unimplemented(sqllex, "create foreign table") }
| CREATE OPERATOR error { return unimplemented(sqllex, "create operator") }
| CREATE PUBLICATION error { return unimplemented(sqllex, "create publication") }
| CREATE opt_or_replace RULE error { return unimplemented(sqllex, "create rule") }
| CREATE SERVER error { return unimplemented(sqllex, "create server") }
| CREATE SUBSCRIPTION error { return unimplemented(sqllex, "create subscription") }
| CREATE TEXT error { return unimplementedWithIssueDetail(sqllex, 7821, "create text") }

create_aggregate_stmt:
  create_aggregate_args_only_stmt
| create_aggregate_order_by_args_stmt
| create_aggregate_old_syntax_stmt

create_aggregate_args_only_stmt:
  CREATE AGGREGATE aggregate_name '(' opt_routine_args ')' '(' SFUNC '=' name ',' STYPE '=' type_name create_agg_args_only_option_list ')'
  { $$.val = &tree.CreateAggregate{Name: $3.unresolvedObjectName(), Args: $5.routineArgs(), SFunc: $10, SType: $14.typeReference(), AggOptions: $15.createAggOptions()} }
| CREATE OR REPLACE AGGREGATE aggregate_name '(' opt_routine_args ')' '(' SFUNC '=' name ',' STYPE '=' type_name create_agg_args_only_option_list ')'
  { $$.val = &tree.CreateAggregate{Name: $5.unresolvedObjectName(), Replace: true, Args: $7.routineArgs(), SFunc: $12, SType: $16.typeReference(), AggOptions: $17.createAggOptions()} }

create_agg_args_only_option_list:
  /* EMPTY */
  { $$.val = []tree.CreateAggOption(nil) }
| create_agg_args_only_option
  { $$.val = []tree.CreateAggOption{$1.createAggOption()} }
| create_agg_args_only_option_list ',' create_agg_args_only_option
  { $$.val = append($1.createAggOptions(), $3.createAggOption()) }

create_agg_args_only_option:
  create_agg_old_syntax_option
| create_agg_parallel_option

create_aggregate_order_by_args_stmt:
  CREATE AGGREGATE aggregate_name '(' opt_routine_args ORDER BY routine_arg_list ')' '(' SFUNC '=' name ',' STYPE '=' type_name create_agg_order_by_args_option_list ')'
  { $$.val = &tree.CreateAggregate{Name: $3.unresolvedObjectName(), Args: $5.routineArgs(), OrderByArgs: $8.routineArgs(), SFunc: $13, SType: $17.typeReference(), AggOptions: $18.createAggOptions()} }
| CREATE OR REPLACE AGGREGATE aggregate_name '(' opt_routine_args ORDER BY routine_arg_list ')' '(' SFUNC '=' name ',' STYPE '=' type_name create_agg_order_by_args_option_list ')'
  { $$.val = &tree.CreateAggregate{Name: $5.unresolvedObjectName(), Replace: true, Args: $7.routineArgs(), OrderByArgs: $10.routineArgs(), SFunc: $15, SType: $19.typeReference(), AggOptions: $20.createAggOptions()} }

create_agg_order_by_args_option_list:
  /* EMPTY */
  { $$.val = []tree.CreateAggOption(nil) }
| create_agg_order_by_args_option
  { $$.val = []tree.CreateAggOption{$1.createAggOption()} }
| create_agg_order_by_args_option_list ',' create_agg_order_by_args_option
  { $$.val = append($1.createAggOptions(), $3.createAggOption()) }

create_agg_order_by_args_option:
  create_agg_common_option
| create_agg_parallel_option
| HYPOTHETICAL
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeHypothetical} }

create_aggregate_old_syntax_stmt:
  CREATE AGGREGATE aggregate_name '(' BASETYPE '=' type_name ',' SFUNC '=' name ',' STYPE '=' type_name create_agg_old_syntax_option_list ')'
  { $$.val = &tree.CreateAggregate{Name: $3.unresolvedObjectName(), BaseType: $7.typeReference(), SFunc: $11, SType: $15.typeReference(), AggOptions: $16.createAggOptions()} }
| CREATE OR REPLACE AGGREGATE aggregate_name '(' BASETYPE '=' type_name ',' SFUNC '=' name ',' STYPE '=' type_name create_agg_old_syntax_option_list ')'
  { $$.val = &tree.CreateAggregate{Name: $5.unresolvedObjectName(), Replace: true, BaseType: $9.typeReference(), SFunc: $13, SType: $17.typeReference(), AggOptions: $18.createAggOptions()} }

create_agg_old_syntax_option_list:
  /* EMPTY */
  { $$.val = []tree.CreateAggOption(nil) }
| create_agg_old_syntax_option
  { $$.val = []tree.CreateAggOption{$1.createAggOption()} }
| create_agg_old_syntax_option_list ',' create_agg_old_syntax_option
  { $$.val = append($1.createAggOptions(), $3.createAggOption()) }

create_agg_old_syntax_option:
  create_agg_common_option
| COMBINEFUNC '=' name
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeCombineFunc, StrVal: $3} }
| SERIALFUNC '=' name
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeSerialFunc, StrVal: $3} }
| DESERIALFUNC '=' name
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeDeserialFunc, StrVal: $3} }
| MSFUNC '=' name
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeMSFunc, StrVal: $3} }
| MINVFUNC '=' name
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeMInvFunc, StrVal: $3} }
| MSTYPE '=' type_name
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeMSType, TypeVal: $3.typeReference()} }
| MSSPACE '=' iconst64
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeMSSpace, IntVal: $3.expr()} }
| MFINALFUNC '=' name
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeMFinalFunc, StrVal: $3} }
| MFINALFUNC_EXTRA '=' TRUE
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeMFinalFuncExtra, BoolVal: true} }
| MFINALFUNC_EXTRA '=' FALSE
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeMFinalFuncExtra, BoolVal: false} }
| MFINALFUNC_MODIFY '=' READ_ONLY
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeMFinalFuncModify, FinalFuncModify: tree.FinalFuncModifyReadOnly} }
| MFINALFUNC_MODIFY '=' SHAREABLE
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeMFinalFuncModify, FinalFuncModify: tree.FinalFuncModifyShareable} }
| MFINALFUNC_MODIFY '=' READ_WRITE
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeMFinalFuncModify, FinalFuncModify: tree.FinalFuncModifyReadWrite} }
| MINITCOND '=' a_expr
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeMInitCond, CondVal: $3.expr()} }
| SORTOP '=' math_op
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeSortOp, SortOp: $3.op()} }

create_agg_common_option:
  SSPACE '=' iconst64
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeSSpace, IntVal: $3.expr()} }
| FINALFUNC '=' name
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeFinalFunc, StrVal: $3} }
| FINALFUNC_EXTRA '=' TRUE
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeFinalFuncExtra, BoolVal: true} }
| FINALFUNC_EXTRA '=' FALSE
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeFinalFuncExtra, BoolVal: false} }
| FINALFUNC_MODIFY '=' READ_ONLY
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeFinalFuncModify, FinalFuncModify: tree.FinalFuncModifyReadOnly} }
| FINALFUNC_MODIFY '=' SHAREABLE
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeFinalFuncModify, FinalFuncModify: tree.FinalFuncModifyShareable} }
| FINALFUNC_MODIFY '=' READ_WRITE
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeFinalFuncModify, FinalFuncModify: tree.FinalFuncModifyReadWrite} }
| INITCOND '=' a_expr
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeInitCond, CondVal: $3.expr()} }

create_agg_parallel_option:
  PARALLEL '=' SAFE
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeParallel, Parallel: tree.ParallelUnsafe} }
| PARALLEL '=' RESTRICTED
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeParallel, Parallel: tree.ParallelRestricted} }
| PARALLEL '=' UNSAFE
  { $$.val = tree.CreateAggOption{Option: tree.AggOptTypeParallel, Parallel: tree.ParallelSafe} }

create_domain_stmt:
  CREATE DOMAIN type_name opt_as typename opt_collate opt_arg_default opt_domain_constraint_list
  {
    $$.val = &tree.CreateDomain{
      TypeName: $3.unresolvedObjectName(),
      DataType: $5.typeReference(),
      Collate: $6.unresolvedObjectName().UnquotedString(),
      Default: $7.expr(),
      Constraints: $8.domainConstraints(),
    }
  }

opt_domain_constraint_list:
  /* EMPTY */
  {
    $$.val = []tree.DomainConstraint(nil)
  }
| domain_constraint_list
  {
    $$.val = $1.domainConstraints()
  }

domain_constraint_list:
  domain_constraint
  {
    $$.val = []tree.DomainConstraint{$1.domainConstraint()}
  }
| domain_constraint_list domain_constraint
  {
    $$.val = append($1.domainConstraints(), $2.domainConstraint())
  }

domain_constraint:
  NOT NULL
  {
    $$.val = tree.DomainConstraint{NotNull: true}
  }
| NULL
  {
    $$.val = tree.DomainConstraint{}
  }
| CHECK '(' a_expr ')'
  {
    $$.val = tree.DomainConstraint{Check: $3.expr()}
  }
| CONSTRAINT constraint_name NOT NULL
  {
    $$.val = tree.DomainConstraint{Constraint: tree.Name($2), NotNull: true}
  }
| CONSTRAINT constraint_name NULL
  {
    $$.val = tree.DomainConstraint{Constraint: tree.Name($2)}
  }
| CONSTRAINT constraint_name CHECK '(' a_expr ')'
  {
    $$.val = tree.DomainConstraint{Constraint: tree.Name($2), Check: $5.expr()}
  }

create_language_stmt:
  CREATE opt_trusted opt_procedural LANGUAGE name opt_language_handler
  {
    $$.val = &tree.CreateLanguage{Name: tree.Name($5), Replace: false, Trusted: $2.bool(), Procedural: $3.bool(), Handler: $6.languageHandler()}
  }
| CREATE OR REPLACE opt_trusted opt_procedural LANGUAGE name opt_language_handler
  {
    $$.val = &tree.CreateLanguage{Name: tree.Name($7), Replace: true, Trusted: $4.bool(), Procedural: $5.bool(), Handler: $8.languageHandler()}
  }

opt_language_handler:
  /* EMPTY */
  {
    $$.val = (*tree.LanguageHandler)(nil)
  }
| HANDLER routine_name opt_handler_inline opt_handler_validator
  {
    $$.val = &tree.LanguageHandler{
      Handler: $2.unresolvedObjectName(),
      Inline: $3.unresolvedObjectName(),
      Validator: $4.unresolvedObjectName(),
    }
  }

opt_handler_inline:
  /* EMPTY */
  {
    $$.val = (*tree.UnresolvedObjectName)(nil)
  }
| INLINE routine_name
  {
    $$.val = $2.unresolvedObjectName()
  }

opt_handler_validator:
  /* EMPTY */
  {
    $$.val = (*tree.UnresolvedObjectName)(nil)
  }
| VALIDATOR routine_name
  {
    $$.val = $2.unresolvedObjectName()
  }

create_extension_stmt:
  CREATE EXTENSION name opt_with opt_schema opt_version opt_cascade
  {
    $$.val = &tree.CreateExtension{Name: tree.Name($3), Schema: $5, Version: $6, Cascade: $7.bool()}
  }
| CREATE EXTENSION IF NOT EXISTS name opt_with opt_schema opt_version opt_cascade
  {
    $$.val = &tree.CreateExtension{Name: tree.Name($6), IfNotExists: true, Schema: $8, Version: $9, Cascade: $10.bool()}
  }

create_procedure_stmt:
  CREATE PROCEDURE routine_name opt_routine_arg_with_default_list create_procedure_option_list
  {
    $$.val = &tree.CreateProcedure{Name: $3.unresolvedObjectName(), Args: $4.routineArgs(), Options: $5.routineOptions()}
  }
| CREATE OR REPLACE PROCEDURE routine_name opt_routine_arg_with_default_list create_procedure_option_list
  {
    $$.val = &tree.CreateProcedure{Name: $5.unresolvedObjectName(), Replace: true, Args: $6.routineArgs(), Options: $7.routineOptions()}
  }

create_function_stmt:
  CREATE FUNCTION routine_name opt_routine_arg_with_default_list create_function_option_list
  {
    $$.val = &tree.CreateFunction{Name: $3.unresolvedObjectName(), Args: $4.routineArgs(), Options: $5.routineOptions()}
  }
| CREATE FUNCTION routine_name opt_routine_arg_with_default_list RETURNS typename create_function_option_list
  {
    $$.val = &tree.CreateFunction{Name: $3.unresolvedObjectName(), Args: $4.routineArgs(), RetType: []tree.SimpleColumnDef{tree.SimpleColumnDef{Type: $6.typeReference()}}, Options: $7.routineOptions()}
  }
| CREATE FUNCTION routine_name opt_routine_arg_with_default_list RETURNS SETOF typename create_function_option_list
  {
    $$.val = &tree.CreateFunction{Name: $3.unresolvedObjectName(), Args: $4.routineArgs(), SetOf: true, RetType: []tree.SimpleColumnDef{tree.SimpleColumnDef{Type: $7.typeReference()}}, Options: $8.routineOptions()}
  }
| CREATE FUNCTION routine_name opt_routine_arg_with_default_list RETURNS TABLE '(' opt_returns_table_col_def_list ')' create_function_option_list
  {
    $$.val = &tree.CreateFunction{Name: $3.unresolvedObjectName(), Args: $4.routineArgs(), RetType: $8.simpleColumnDefs(), Options: $10.routineOptions()}
  }
| CREATE OR REPLACE FUNCTION routine_name opt_routine_arg_with_default_list create_function_option_list
  {
    $$.val = &tree.CreateFunction{Name: $5.unresolvedObjectName(), Replace: true, Args: $6.routineArgs(), Options: $7.routineOptions()}
  }
| CREATE OR REPLACE FUNCTION routine_name opt_routine_arg_with_default_list RETURNS typename create_function_option_list
  {
    $$.val = &tree.CreateFunction{Name: $5.unresolvedObjectName(), Replace: true, Args: $6.routineArgs(), RetType: []tree.SimpleColumnDef{tree.SimpleColumnDef{Type: $8.typeReference()}}, Options: $9.routineOptions()}
  }
| CREATE OR REPLACE FUNCTION routine_name opt_routine_arg_with_default_list RETURNS SETOF typename create_function_option_list
  {
    $$.val = &tree.CreateFunction{Name: $5.unresolvedObjectName(), Replace: true, Args: $6.routineArgs(), SetOf: true, RetType: []tree.SimpleColumnDef{tree.SimpleColumnDef{Type: $9.typeReference()}}, Options: $10.routineOptions()}
  }
| CREATE OR REPLACE FUNCTION routine_name opt_routine_arg_with_default_list RETURNS TABLE '(' opt_returns_table_col_def_list ')' create_function_option_list
  {
    $$.val = &tree.CreateFunction{Name: $5.unresolvedObjectName(), Replace: true, Args: $6.routineArgs(), RetType: $10.simpleColumnDefs(), Options: $12.routineOptions()}
  }

opt_returns_table_col_def_list:
  returns_table_col_def
  {
    $$.val = []tree.SimpleColumnDef{$1.simpleColumnDef()}
  }
| opt_returns_table_col_def_list ',' returns_table_col_def
  {
    $$.val = append($1.simpleColumnDefs(), $3.simpleColumnDef())
  }

returns_table_col_def:
  column_name typename
  {
    $$.val = tree.SimpleColumnDef{Name: tree.Name($1), Type: $2.typeReference()}
  }

opt_routine_arg_with_default_list:
  /* EMPTY */
  {
    $$.val = []*tree.RoutineArg{}
  }
| '(' ')'
  {
    $$.val = []*tree.RoutineArg{}
  }
| '(' routine_arg_with_default_list ')'
  {
    $$.val = $2.routineArgs()
  }

routine_arg_with_default_list:
  routine_arg_with_default
  {
    $$.val = []*tree.RoutineArg{$1.routineArg()}
  }
| routine_arg_with_default_list ',' routine_arg_with_default
  {
    $$.val = append($1.routineArgs(), $3.routineArg())
  }

routine_arg_with_default:
  routine_arg opt_arg_default
  {
    arg := $1.routineArg()
    arg.Default = $2.expr()
    $$.val = arg
  }

opt_arg_default:
  /* EMPTY */
  {
    $$.val = nil
  }
| opt_default a_expr
  {
    $$.val = $2.expr()
  }

opt_default:
  DEFAULT {}
| '=' {}

alter_function_option_list:
  alter_function_option
  {
    $$.val = []tree.RoutineOption{$1.routineOption()}
  }
| alter_function_option_list alter_function_option
  {
    $$.val = append($1.routineOptions(), $2.routineOption())
  }

alter_function_option:
  function_option
| alter_procedure_option

create_function_option_list:
  create_function_option
  {
    $$.val = []tree.RoutineOption{$1.routineOption()}
  }
| create_function_option_list create_function_option
  {
    $$.val = append($1.routineOptions(), $2.routineOption())
  }

create_function_option:
  create_procedure_option
| function_option
| WINDOW
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionWindow}
  }

function_option:
  IMMUTABLE
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionVolatility, Volatility: tree.VolatilityImmutable}
  }
| STABLE
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionVolatility, Volatility: tree.VolatilityStable}
  }
| VOLATILE
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionVolatility, Volatility: tree.VolatilityVolatile}
  }
| opt_not LEAKPROOF
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionLeakProof, IsLeakProof: $1.bool()}
  }
| CALLED ON NULL INPUT
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionNullInput, NullInput: tree.CalledOnNullInput}
  }
| RETURNS NULL ON NULL INPUT
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionNullInput, NullInput: tree.ReturnsNullOnNullInput}
  }
| STRICT
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionNullInput, NullInput: tree.StrictNullInput}
  }
| PARALLEL UNSAFE
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionParallel, Parallel: tree.ParallelUnsafe}
  }
| PARALLEL RESTRICTED
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionParallel, Parallel: tree.ParallelRestricted}
  }
| PARALLEL SAFE
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionParallel, Parallel: tree.ParallelSafe}
  }
| COST signed_iconst
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionCost, Cost: $2.expr()}
  }
| ROWS signed_iconst
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionRows, Rows: $2.expr()}
  }
| SUPPORT name
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionSupport, Support: $2}
  }

alter_procedure_option_list:
  alter_procedure_option
  {
    $$.val = []tree.RoutineOption{$1.routineOption()}
  }
| alter_procedure_option_list alter_procedure_option
  {
    $$.val = append($1.routineOptions(), $2.routineOption())
  }

alter_procedure_option:
 opt_external SECURITY definer_or_invoker
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionSecurity, External: $1.bool(), Definer: $3.bool()}
  }
| SET generic_set_single_config
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionSet, SetVar: $2.setVar()}
  }
| RESET name
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionReset, ResetParam: $2}
  }
| RESET ALL
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionReset, ResetAll: true}
  }

create_procedure_option_list:
  create_procedure_option
  {
    $$.val = []tree.RoutineOption{$1.routineOption()}
  }
| create_procedure_option_list create_procedure_option
  {
    $$.val = append($1.routineOptions(), $2.routineOption())
  }

create_procedure_option:
  LANGUAGE name
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionLanguage, Language: $2}
  }
| TRANSFORM for_type_list
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionTransform, TransformTypes: $2.typeReferences()}
  }
| opt_external SECURITY definer_or_invoker
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionSecurity, External: $1.bool(), Definer: $3.bool()}
  }
| SET generic_set_single_config
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionSet, SetVar: $2.setVar()}
  }
| AS SCONST
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionAs1, Definition: $2}
  }
| AS SCONST ',' SCONST
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionAs2, ObjFile: $2, LinkSymbol: $4}
  }
| sql_body
  {
    $$.val = tree.RoutineOption{OptionType: tree.OptionSqlBody, SqlBody: $1.stmt()}
  }

sql_body:
  RETURN a_expr
  {
    $$.val = &tree.Return{Expr: $2.expr()}
  }
| begin_end_block
  {
    $$.val = $1.stmt()
  }

definer_or_invoker:
  DEFINER
  {
    $$.val = true
  }
| INVOKER
  {
    $$.val = false
  }

opt_external:
  /* EMPTY */
  {
    $$.val = false
  }
| EXTERNAL
  {
    $$.val = true
  }

opt_not:
  /* EMPTY */
  {
    $$.val = true
  }
| NOT
  {
    $$.val = false
  }

for_type_list:
  FOR TYPE typename
  {
    $$.val = []tree.ResolvableTypeReference{$3.typeReference()}
  }
| for_type_list ',' FOR TYPE typename
  {
    $$.val = append($1.typeReferences(), $5.typeReference())
  }

begin_end_block:
  BEGIN ATOMIC END
  {
    $$.val = &tree.BeginEndBlock{}
  }
| BEGIN ATOMIC stmt_list ';' END
  {
    $$.val = &tree.BeginEndBlock{Statements: $3.stmts()}
  }

opt_schema:
  /* EMPTY */
  {
    $$ = ""
  }
| SCHEMA schema_name
  {
    $$ = $2
  }

opt_version:
  /* EMPTY */
  {
    $$ = ""
  }
| VERSION name
  {
    $$ = $2
  }

opt_cascade:
  /* EMPTY */
  {
    $$.val = false
  }
| CASCADE
  {
    $$.val = true
  }

opt_or_replace:
  OR REPLACE {}
| /* EMPTY */ {}

opt_trusted:
  /* EMPTY */
  {
    $$.val = false
  }
| TRUSTED
  {
    $$.val = true
  }

opt_procedural:
  /* EMPTY */
  {
    $$.val = false
  }
| PROCEDURAL
  {
    $$.val = true
  }

drop_unsupported:
  DROP CAST error { return unimplemented(sqllex, "drop cast") }
| DROP COLLATION error { return unimplemented(sqllex, "drop collation") }
| DROP CONVERSION error { return unimplemented(sqllex, "drop conversion") }
| DROP FOREIGN TABLE error { return unimplemented(sqllex, "drop foreign table") }
| DROP FOREIGN DATA error { return unimplemented(sqllex, "drop fdw") }
| DROP OPERATOR error { return unimplemented(sqllex, "drop operator") }
| DROP PUBLICATION error { return unimplemented(sqllex, "drop publication") }
| DROP RULE error { return unimplemented(sqllex, "drop rule") }
| DROP SERVER error { return unimplemented(sqllex, "drop server") }
| DROP SUBSCRIPTION error { return unimplemented(sqllex, "drop subscription") }
| DROP TEXT error { return unimplementedWithIssueDetail(sqllex, 7821, "drop text") }

drop_aggregate_stmt:
  DROP AGGREGATE drop_aggregates opt_drop_behavior
  {
    $$.val = &tree.DropAggregate{Aggregates: $3.aggregatesToDrop(), DropBehavior: $4.dropBehavior()}
  }
| DROP AGGREGATE IF EXISTS drop_aggregates opt_drop_behavior
  {
    $$.val = &tree.DropAggregate{Aggregates: $5.aggregatesToDrop(), IfExists: true, DropBehavior: $6.dropBehavior()}
  }

drop_aggregates:
  aggregate_name '(' aggregate_signature ')'
  {
    $$.val = []tree.AggregateToDrop{{Name: $1.unresolvedObjectName(), AggSig: $3.aggregateSignature()}}
  }
| drop_aggregates ',' aggregate_name '(' aggregate_signature ')'
  {
    $$.val = append($1.aggregatesToDrop(), tree.AggregateToDrop{Name: $3.unresolvedObjectName(), AggSig: $5.aggregateSignature()})
  }

drop_domain_stmt:
  DROP DOMAIN table_name_list opt_drop_behavior
  {
    $$.val = &tree.DropDomain{Names: $3.tableNames(), DropBehavior: $4.dropBehavior()}
  }
| DROP DOMAIN IF EXISTS table_name_list opt_drop_behavior
  {
    $$.val = &tree.DropDomain{Names: $5.tableNames(), IfExists: true, DropBehavior: $6.dropBehavior()}
  }

drop_language_stmt:
  DROP opt_procedural LANGUAGE name opt_drop_behavior
  {
    $$.val = &tree.DropLanguage{Name: tree.Name($4), Procedural: $2.bool(), IfExists: false, DropBehavior: $5.dropBehavior()}
  }
| DROP opt_procedural LANGUAGE IF EXISTS name opt_drop_behavior
  {
    $$.val = &tree.DropLanguage{Name: tree.Name($6), Procedural: $2.bool(), IfExists: true, DropBehavior: $7.dropBehavior()}
  }

drop_extension_stmt:
  DROP EXTENSION name_list opt_drop_behavior
  {
    $$.val = &tree.DropExtension{Names: $3.nameList(), DropBehavior: $4.dropBehavior()}
  }
| DROP EXTENSION IF EXISTS name_list opt_drop_behavior
  {
    $$.val = &tree.DropExtension{Names: $5.nameList(), IfExists: true, DropBehavior: $6.dropBehavior()}
  }

drop_procedure_stmt:
  DROP PROCEDURE function_name_with_args_list opt_drop_behavior
  {
    $$.val = &tree.DropProcedure{Procedures: $3.routineWithArgs(), DropBehavior: $4.dropBehavior()}
  }
| DROP PROCEDURE IF EXISTS function_name_with_args_list opt_drop_behavior
  {
    $$.val = &tree.DropProcedure{Procedures: $5.routineWithArgs(), IfExists: true, DropBehavior: $6.dropBehavior()}
  }

drop_function_stmt:
  DROP FUNCTION function_name_with_args_list opt_drop_behavior
  {
    $$.val = &tree.DropFunction{Functions: $3.routineWithArgs(), DropBehavior: $4.dropBehavior()}
  }
| DROP FUNCTION IF EXISTS function_name_with_args_list opt_drop_behavior
  {
    $$.val = &tree.DropFunction{Functions: $5.routineWithArgs(), IfExists: true, DropBehavior: $6.dropBehavior()}
  }

function_name_with_args_list:
  db_object_name opt_routine_arg_with_default_list
  {
    $$.val = []tree.RoutineWithArgs{{Name: $1.unresolvedObjectName(), Args: $2.routineArgs()}}
  }
| function_name_with_args_list ',' db_object_name opt_routine_arg_with_default_list
  {
    $$.val = append($1.routineWithArgs(), tree.RoutineWithArgs{Name: $3.unresolvedObjectName(), Args: $4.routineArgs()})
  }

create_ddl_stmt:
  create_changefeed_stmt
| create_database_stmt // EXTEND WITH HELP: CREATE DATABASE
| create_schema_stmt   // EXTEND WITH HELP: CREATE SCHEMA
| create_type_stmt     // EXTEND WITH HELP: CREATE TYPE
| create_domain_stmt    // EXTEND WITH HELP: CREATE DOMAIN
| create_ddl_stmt_schema_element // help texts in sub-rule

create_ddl_stmt_schema_element:
  create_index_stmt    // EXTEND WITH HELP: CREATE INDEX
| create_table_stmt    // EXTEND WITH HELP: CREATE TABLE
| create_table_as_stmt // EXTEND WITH HELP: CREATE TABLE ... AS
// Error case for both CREATE TABLE and CREATE TABLE ... AS in one
| CREATE opt_persistence_temp_table TABLE error   // SHOW HELP: CREATE TABLE
| create_view_stmt     // EXTEND WITH HELP: CREATE VIEW
| create_materialized_view_stmt // EXTEND WITH HELP: CREATE MATERIALIZED VIEW
| create_sequence_stmt // EXTEND WITH HELP: CREATE SEQUENCE
| create_trigger_stmt  // EXTEND WITH HELP: CREATE TRIGGER

// %Help: CREATE STATISTICS - create a new table statistic
// %Category: Misc
// %Text:
// CREATE STATISTICS <statisticname>
//   [ON <colname> [, ...]]
//   FROM <tablename> [AS OF SYSTEM TIME <expr>]
create_stats_stmt:
  CREATE STATISTICS statistics_name opt_stats_columns FROM create_stats_target opt_create_stats_options
  {
    $$.val = &tree.CreateStats{
      Name: tree.Name($3),
      ColumnNames: $4.nameList(),
      Table: $6.tblExpr(),
      Options: *$7.createStatsOptions(),
    }
  }
| CREATE STATISTICS error // SHOW HELP: CREATE STATISTICS

opt_stats_columns:
  ON name_list
  {
    $$.val = $2.nameList()
  }
| /* EMPTY */
  {
    $$.val = tree.NameList(nil)
  }

create_stats_target:
  table_name
  {
    $$.val = $1.unresolvedObjectName()
  }
| '[' iconst64 ']'
  {
    /* SKIP DOC */
    $$.val = &tree.TableRef{
      TableID: $2.int64(),
    }
  }

opt_create_stats_options:
  WITH OPTIONS create_stats_option_list
  {
    /* SKIP DOC */
    $$.val = $3.createStatsOptions()
  }
// Allow AS OF SYSTEM TIME without WITH OPTIONS, for consistency with other
// statements.
| as_of_clause
  {
    $$.val = &tree.CreateStatsOptions{
      AsOf: $1.asOfClause(),
    }
  }
| /* EMPTY */
  {
    $$.val = &tree.CreateStatsOptions{}
  }

create_stats_option_list:
  create_stats_option
  {
    $$.val = $1.createStatsOptions()
  }
| create_stats_option_list create_stats_option
  {
    a := $1.createStatsOptions()
    b := $2.createStatsOptions()
    if err := a.CombineWith(b); err != nil {
      return setErr(sqllex, err)
    }
    $$.val = a
  }

create_stats_option:
  THROTTLING FCONST
  {
    /* SKIP DOC */
    value, _ := constant.Float64Val($2.numVal().AsConstantValue())
    if value < 0.0 || value >= 1.0 {
      sqllex.Error("THROTTLING fraction must be between 0 and 1")
      return 1
    }
    $$.val = &tree.CreateStatsOptions{
      Throttling: value,
    }
  }
| as_of_clause
  {
    $$.val = &tree.CreateStatsOptions{
      AsOf: $1.asOfClause(),
    }
  }

create_changefeed_stmt:
  CREATE CHANGEFEED FOR changefeed_targets opt_changefeed_sink opt_with_options
  {
    $$.val = &tree.CreateChangefeed{
      Targets: $4.targetList(),
      SinkURI: $5.expr(),
      Options: $6.kvOptions(),
    }
  }
| EXPERIMENTAL CHANGEFEED FOR changefeed_targets opt_with_options
  {
    /* SKIP DOC */
    $$.val = &tree.CreateChangefeed{
      Targets: $4.targetList(),
      Options: $5.kvOptions(),
    }
  }

changefeed_targets:
  single_table_pattern_list
  {
    $$.val = tree.TargetList{Tables: $1.tablePatterns()}
  }
| TABLE single_table_pattern_list
  {
    $$.val = tree.TargetList{Tables: $2.tablePatterns()}
  }

single_table_pattern_list:
  table_name
  {
    $$.val = tree.TablePatterns{$1.unresolvedObjectName().ToUnresolvedName()}
  }
| single_table_pattern_list ',' table_name
  {
    $$.val = append($1.tablePatterns(), $3.unresolvedObjectName().ToUnresolvedName())
  }


opt_changefeed_sink:
  INTO string_or_placeholder
  {
    $$.val = $2.expr()
  }
| /* EMPTY */
  {
    /* SKIP DOC */
    $$.val = nil
  }

// %Help: DELETE - delete rows from a table
// %Category: DML
// %Text: DELETE FROM <tablename> [WHERE <expr>]
//               [ORDER BY <exprs...>]
//               [LIMIT <expr>]
//               [RETURNING <exprs...>]
// %SeeAlso: WEBDOCS/delete.html
delete_stmt:
  opt_with_clause DELETE FROM table_expr_opt_alias_idx opt_using_clause opt_where_clause opt_sort_clause opt_limit_clause returning_clause
  {
    $$.val = &tree.Delete{
      With: $1.with(),
      Table: $4.tblExpr(),
      Where: tree.NewWhere(tree.AstWhere, $6.expr()),
      OrderBy: $7.orderBy(),
      Limit: $8.limit(),
      Returning: $9.retClause(),
    }
  }
| opt_with_clause DELETE error // SHOW HELP: DELETE

opt_using_clause:
  USING from_list { return unimplementedWithIssueDetail(sqllex, 40963, "delete using") }
| /* EMPTY */ { }

// %Help: DISCARD - reset the session to its initial state
// %Category: Cfg
// %Text: DISCARD ALL
discard_stmt:
  DISCARD ALL
  {
    $$.val = &tree.Discard{Mode: tree.DiscardModeAll}
  }
| DISCARD PLANS { return unimplemented(sqllex, "discard plans") }
| DISCARD SEQUENCES { return unimplemented(sqllex, "discard sequences") }
| DISCARD TEMP { return unimplemented(sqllex, "discard temp") }
| DISCARD TEMPORARY { return unimplemented(sqllex, "discard temp") }
| DISCARD error // SHOW HELP: DISCARD

// %Help: DROP
// %Category: Group
// %Text:
// DROP DATABASE, DROP INDEX, DROP TABLE, DROP VIEW, DROP SEQUENCE,
// DROP USER, DROP ROLE, DROP TYPE
drop_stmt:
  drop_ddl_stmt      // help texts in sub-rule
| drop_role_stmt     // EXTEND WITH HELP: DROP ROLE
| drop_schedule_stmt // EXTEND WITH HELP: DROP SCHEDULES
| drop_function_stmt // EXTEND WITH HELP: DROP FUNCTION
| drop_procedure_stmt // EXTEND WITH HELP: DROP PROCEDURE
| drop_domain_stmt   // EXTEND WITH HELP: DROP DOMAIN
| drop_extension_stmt // EXTEND WITH HELP: DROP EXTENSION
| drop_language_stmt // EXTEND WITH HELP: DROP LANGUAGE
| drop_aggregate_stmt // EXTEND WITH HELP: DROP AGGREGATE
| drop_unsupported   {}
| DROP error         // SHOW HELP: DROP

drop_ddl_stmt:
  drop_database_stmt // EXTEND WITH HELP: DROP DATABASE
| drop_index_stmt    // EXTEND WITH HELP: DROP INDEX
| drop_table_stmt    // EXTEND WITH HELP: DROP TABLE
| drop_trigger_stmt  // EXTEND WITH HELP: DROP TRIGGER
| drop_view_stmt     // EXTEND WITH HELP: DROP VIEW
| drop_sequence_stmt // EXTEND WITH HELP: DROP SEQUENCE
| drop_schema_stmt   // EXTEND WITH HELP: DROP SCHEMA
| drop_type_stmt     // EXTEND WITH HELP: DROP TYPE

// %Help: DROP VIEW - remove a view
// %Category: DDL
// %Text: DROP [MATERIALIZED] VIEW [IF EXISTS] <tablename> [, ...] [CASCADE | RESTRICT]
// %SeeAlso: WEBDOCS/drop-index.html
drop_view_stmt:
  DROP VIEW table_name_list opt_drop_behavior
  {
    $$.val = &tree.DropView{Names: $3.tableNames(), IfExists: false, DropBehavior: $4.dropBehavior()}
  }
| DROP VIEW IF EXISTS table_name_list opt_drop_behavior
  {
    $$.val = &tree.DropView{Names: $5.tableNames(), IfExists: true, DropBehavior: $6.dropBehavior()}
  }
| DROP MATERIALIZED VIEW table_name_list opt_drop_behavior
  {
    $$.val = &tree.DropView{
      Names: $4.tableNames(),
      IfExists: false,
      DropBehavior: $5.dropBehavior(),
      IsMaterialized: true,
    }
  }
| DROP MATERIALIZED VIEW IF EXISTS table_name_list opt_drop_behavior
  {
    $$.val = &tree.DropView{
      Names: $6.tableNames(),
      IfExists: true,
      DropBehavior: $7.dropBehavior(),
      IsMaterialized: true,
    }
  }
| DROP VIEW error // SHOW HELP: DROP VIEW

// %Help: DROP SEQUENCE - remove a sequence
// %Category: DDL
// %Text: DROP SEQUENCE [IF EXISTS] <sequenceName> [, ...] [CASCADE | RESTRICT]
// %SeeAlso: DROP
drop_sequence_stmt:
  DROP SEQUENCE table_name_list opt_drop_behavior
  {
    $$.val = &tree.DropSequence{Names: $3.tableNames(), IfExists: false, DropBehavior: $4.dropBehavior()}
  }
| DROP SEQUENCE IF EXISTS table_name_list opt_drop_behavior
  {
    $$.val = &tree.DropSequence{Names: $5.tableNames(), IfExists: true, DropBehavior: $6.dropBehavior()}
  }
| DROP SEQUENCE error // SHOW HELP: DROP VIEW

// %Help: DROP TABLE - remove a table
// %Category: DDL
// %Text: DROP TABLE [IF EXISTS] <tablename> [, ...] [CASCADE | RESTRICT]
// %SeeAlso: WEBDOCS/drop-table.html
drop_table_stmt:
  DROP TABLE table_name_list opt_drop_behavior
  {
    $$.val = &tree.DropTable{Names: $3.tableNames(), IfExists: false, DropBehavior: $4.dropBehavior()}
  }
| DROP TABLE IF EXISTS table_name_list opt_drop_behavior
  {
    $$.val = &tree.DropTable{Names: $5.tableNames(), IfExists: true, DropBehavior: $6.dropBehavior()}
  }
| DROP TABLE error // SHOW HELP: DROP TABLE

drop_trigger_stmt:
  DROP TRIGGER trigger_name ON table_name opt_drop_behavior
  {
    $$.val = &tree.DropTrigger{Name: tree.Name($3), OnTable: $5.unresolvedObjectName().ToTableName(), DropBehavior: $6.dropBehavior()}
  }
| DROP TRIGGER IF EXISTS trigger_name ON table_name opt_drop_behavior
  {
    $$.val = &tree.DropTrigger{Name: tree.Name($5), OnTable: $7.unresolvedObjectName().ToTableName(), DropBehavior: $8.dropBehavior()}
  }

// %Help: DROP INDEX - remove an index
// %Category: DDL
// %Text: DROP INDEX [CONCURRENTLY] [IF EXISTS] <idxname> [, ...] [CASCADE | RESTRICT]
// %SeeAlso: WEBDOCS/drop-index.html
drop_index_stmt:
  DROP INDEX opt_concurrently table_index_name_list opt_drop_behavior
  {
    $$.val = &tree.DropIndex{
      IndexList: $4.newTableIndexNames(),
      IfExists: false,
      DropBehavior: $5.dropBehavior(),
      Concurrently: $3.bool(),
    }
  }
| DROP INDEX opt_concurrently IF EXISTS table_index_name_list opt_drop_behavior
  {
    $$.val = &tree.DropIndex{
      IndexList: $6.newTableIndexNames(),
      IfExists: true,
      DropBehavior: $7.dropBehavior(),
      Concurrently: $3.bool(),
    }
  }
| DROP INDEX error // SHOW HELP: DROP INDEX

// %Help: DROP DATABASE - remove a database
// %Category: DDL
// %Text: DROP DATABASE [IF EXISTS] <databasename> [ [ WITH ] ( option [, ...] ) ]
// %SeeAlso: WEBDOCS/drop-database.html
drop_database_stmt:
  DROP DATABASE database_name opt_with_force
  {
    $$.val = &tree.DropDatabase{
      Name: tree.Name($3),
      IfExists: false,
      Force: $4.bool(),
    }
  }
| DROP DATABASE IF EXISTS database_name opt_with_force
  {
    $$.val = &tree.DropDatabase{
      Name: tree.Name($5),
      IfExists: true,
      Force: $6.bool(),
    }
  }
| DROP DATABASE error // SHOW HELP: DROP DATABASE

opt_with_force:
  /* EMPTY */
  {
    $$.val = false
  }
| opt_with '(' force_list ')'
  {
    $$.val = true
  }

force_list:
  FORCE
| force_list ',' FORCE

// %Help: DROP TYPE - remove a type
// %Category: DDL
// %Text: DROP TYPE [IF EXISTS] <type_name> [, ...] [CASCADE | RESTRICT]
drop_type_stmt:
  DROP TYPE type_name_list opt_drop_behavior
  {
    $$.val = &tree.DropType{
      Names: $3.unresolvedObjectNames(),
      IfExists: false,
      DropBehavior: $4.dropBehavior(),
    }
  }
| DROP TYPE IF EXISTS type_name_list opt_drop_behavior
  {
    $$.val = &tree.DropType{
      Names: $5.unresolvedObjectNames(),
      IfExists: true,
      DropBehavior: $6.dropBehavior(),
    }
  }
| DROP TYPE error // SHOW HELP: DROP TYPE

type_name_list:
  type_name
  {
    $$.val = []*tree.UnresolvedObjectName{$1.unresolvedObjectName()}
  }
| type_name_list ',' type_name
  {
    $$.val = append($1.unresolvedObjectNames(), $3.unresolvedObjectName())
  }

// %Help: DROP SCHEMA - remove a schema
// %Category: DDL
// %Text: DROP SCHEMA [IF EXISTS] <schema_name> [, ...] [CASCADE | RESTRICT]
drop_schema_stmt:
  DROP SCHEMA schema_name_list opt_drop_behavior
  {
    $$.val = &tree.DropSchema{
      Names: $3.strs(),
      IfExists: false,
      DropBehavior: $4.dropBehavior(),
    }
  }
| DROP SCHEMA IF EXISTS schema_name_list opt_drop_behavior
  {
    $$.val = &tree.DropSchema{
      Names: $5.strs(),
      IfExists: true,
      DropBehavior: $6.dropBehavior(),
    }
  }
| DROP SCHEMA error // SHOW HELP: DROP SCHEMA

schema_name_list:
  schema_name
  {
    $$.val = []string{$1}
  }
| schema_name_list ',' schema_name
  {
    $$.val = append($1.strs(), $3)
  }

// %Help: DROP ROLE - remove a user
// %Category: Priv
// %Text: DROP ROLE [IF EXISTS] <user> [, ...]
// %SeeAlso: CREATE ROLE, SHOW ROLE
drop_role_stmt:
  DROP role_or_group_or_user string_or_placeholder_list
  {
    $$.val = &tree.DropRole{Names: $3.exprs(), IfExists: false, IsRole: $2.bool()}
  }
| DROP role_or_group_or_user IF EXISTS string_or_placeholder_list
  {
    $$.val = &tree.DropRole{Names: $5.exprs(), IfExists: true, IsRole: $2.bool()}
  }
| DROP role_or_group_or_user error // SHOW HELP: DROP ROLE

table_name_list:
  table_name
  {
    name := $1.unresolvedObjectName().ToTableName()
    $$.val = tree.TableNames{name}
  }
| table_name_list ',' table_name
  {
    name := $3.unresolvedObjectName().ToTableName()
    $$.val = append($1.tableNames(), name)
  }

// %Help: ANALYZE - collect table statistics
// %Category: Misc
// %Text:
// ANALYZE [<tablename>]
//
// %SeeAlso: CREATE STATISTICS
analyze_stmt:
  ANALYZE
  {
    $$.val = &tree.Analyze{}
  }
| ANALYZE analyze_target
  {
    $$.val = &tree.Analyze{
      Table: $2.tblExpr(),
    }
  }
| ANALYZE error // SHOW HELP: ANALYZE
| ANALYSE analyze_target
  {
    $$.val = &tree.Analyze{
      Table: $2.tblExpr(),
    }
  }
| ANALYSE error // SHOW HELP: ANALYZE

analyze_target:
  table_name
  {
    $$.val = $1.unresolvedObjectName()
  }

explain_verb:
  EXPLAIN
| DESCRIBE
| DESC

// %Help: EXPLAIN - show the logical plan of a query
// %Category: Misc
// %Text:
// EXPLAIN <statement>
// EXPLAIN ([PLAN ,] <planoptions...> ) <statement>
// EXPLAIN [ANALYZE] (DISTSQL) <statement>
// EXPLAIN ANALYZE [(DISTSQL)] <statement>
//
// Explainable statements:
//     SELECT, CREATE, DROP, ALTER, INSERT, UPSERT, UPDATE, DELETE,
//     SHOW, EXPLAIN
//
// Plan options:
//     TYPES, VERBOSE, OPT
//
// %SeeAlso: WEBDOCS/explain.html
explain_stmt:
  explain_verb preparable_stmt
  {
    var err error
    $$.val, err = tree.MakeExplain(nil /* options */, $2.stmt())
    if err != nil {
      return setErr(sqllex, err)
    }
  }
| explain_verb error // SHOW HELP: EXPLAIN
| explain_verb '(' explain_option_list ')' preparable_stmt
  {
    var err error
    $$.val, err = tree.MakeExplain($3.strs(), $5.stmt())
    if err != nil {
      return setErr(sqllex, err)
    }
  }
| explain_verb ANALYZE preparable_stmt
  {
    var err error
    $$.val, err = tree.MakeExplain([]string{"DISTSQL", "ANALYZE"}, $3.stmt())
    if err != nil {
      return setErr(sqllex, err)
    }
  }
| explain_verb ANALYSE preparable_stmt
  {
    var err error
    $$.val, err = tree.MakeExplain([]string{"DISTSQL", "ANALYZE"}, $3.stmt())
    if err != nil {
      return setErr(sqllex, err)
    }
  }
| explain_verb ANALYZE '(' explain_option_list ')' preparable_stmt
  {
    var err error
    $$.val, err = tree.MakeExplain(append($4.strs(), "ANALYZE"), $6.stmt())
    if err != nil {
      return setErr(sqllex, err)
    }
  }
| explain_verb ANALYSE '(' explain_option_list ')' preparable_stmt
  {
    var err error
    $$.val, err = tree.MakeExplain(append($4.strs(), "ANALYZE"), $6.stmt())
    if err != nil {
      return setErr(sqllex, err)
    }
  }
// This second error rule is necessary, because otherwise
// preparable_stmt also provides "selectclause := '(' error ..." and
// cause a help text for the select clause, which will be confusing in
// the context of EXPLAIN.
| explain_verb '(' error // SHOW HELP: EXPLAIN

describe_table_stmt:
  // DELETE, UPDATE, etc. are non-reserved words in the grammar, so we can't use a table_name here, as it causes 
  // conflicts with EXPLAIN UPDATE ... etc.
  explain_verb db_object_name_no_keywords opt_as_of_clause
  {
    asOf := $3.asOfClause()
    $$.val = &tree.Explain{ExplainOptions: tree.ExplainOptions{}, TableName: $2.unresolvedObjectName(), AsOf: &asOf}
  }

preparable_stmt:
  alter_stmt     // help texts in sub-rule
| backup_stmt    // EXTEND WITH HELP: BACKUP
| cancel_stmt    // help texts in sub-rule
| create_stmt    // help texts in sub-rule
| delete_stmt    // EXTEND WITH HELP: DELETE
| drop_stmt      // help texts in sub-rule
| explain_stmt   // EXTEND WITH HELP: EXPLAIN
| describe_table_stmt
| import_stmt    // EXTEND WITH HELP: IMPORT
| insert_stmt    // EXTEND WITH HELP: INSERT
| pause_stmt     // help texts in sub-rule
| reset_stmt     // help texts in sub-rule
| restore_stmt   // EXTEND WITH HELP: RESTORE
| resume_stmt    // help texts in sub-rule
| export_stmt    // EXTEND WITH HELP: EXPORT
| scrub_stmt     // help texts in sub-rule
| select_stmt    // help texts in sub-rule
  {
    $$.val = $1.slct()
  }
| show_stmt         // help texts in sub-rule
| truncate_stmt     // EXTEND WITH HELP: TRUNCATE
| update_stmt       // EXTEND WITH HELP: UPDATE
| upsert_stmt       // EXTEND WITH HELP: UPSERT

// These are statements that can be used as a data source using the special
// syntax with brackets. These are a subset of preparable_stmt.
row_source_extension_stmt:
  delete_stmt       // EXTEND WITH HELP: DELETE
| explain_stmt      // EXTEND WITH HELP: EXPLAIN
| insert_stmt       // EXTEND WITH HELP: INSERT
| select_stmt       // help texts in sub-rule
  {
    $$.val = $1.slct()
  }
| show_stmt         // help texts in sub-rule
| update_stmt       // EXTEND WITH HELP: UPDATE
| upsert_stmt       // EXTEND WITH HELP: UPSERT

explain_option_list:
  explain_option_name explain_option_value // boolean_opt is ignored
  {
    $$.val = []string{$1}
  }
| explain_option_list ',' explain_option_name explain_option_value
  {
    $$.val = append($1.strs(), $3)
  }

explain_option_value:
  /* EMPTY */
  {
    $$ = ""
  }
| TRUE
| FALSE
| OFF
| ON
| TEXT
| XML
| JSON
| YAML

// %Help: PREPARE - prepare a statement for later execution
// %Category: Misc
// %Text: PREPARE <name> [ ( <types...> ) ] AS <query>
// %SeeAlso: EXECUTE, DEALLOCATE, DISCARD
prepare_stmt:
  PREPARE table_alias_name prep_type_clause AS preparable_stmt
  {
    $$.val = &tree.Prepare{
      Name: tree.Name($2),
      Types: $3.typeReferences(),
      Statement: $5.stmt(),
    }
  }
| PREPARE table_alias_name prep_type_clause AS OPT PLAN SCONST
  {
    /* SKIP DOC */
    $$.val = &tree.Prepare{
      Name: tree.Name($2),
      Types: $3.typeReferences(),
      Statement: &tree.CannedOptPlan{Plan: $7},
    }
  }
| PREPARE error // SHOW HELP: PREPARE

prep_type_clause:
  '(' type_list ')'
  {
    $$.val = $2.typeReferences();
  }
| /* EMPTY */
  {
    $$.val = []tree.ResolvableTypeReference(nil)
  }

// %Help: EXECUTE - execute a statement prepared previously
// %Category: Misc
// %Text: EXECUTE <name> [ ( <exprs...> ) ]
// %SeeAlso: PREPARE, DEALLOCATE, DISCARD
execute_stmt:
  EXECUTE table_alias_name execute_param_clause
  {
    $$.val = &tree.Execute{
      Name: tree.Name($2),
      Params: $3.exprs(),
    }
  }
| EXECUTE table_alias_name execute_param_clause DISCARD ROWS
  {
    /* SKIP DOC */
    $$.val = &tree.Execute{
      Name: tree.Name($2),
      Params: $3.exprs(),
      DiscardRows: true,
    }
  }
| EXECUTE error // SHOW HELP: EXECUTE

execute_param_clause:
  '(' expr_list ')'
  {
    $$.val = $2.exprs()
  }
| /* EMPTY */
  {
    $$.val = tree.Exprs(nil)
  }

// %Help: DEALLOCATE - remove a prepared statement
// %Category: Misc
// %Text: DEALLOCATE [PREPARE] { <name> | ALL }
// %SeeAlso: PREPARE, EXECUTE, DISCARD
deallocate_stmt:
  DEALLOCATE name
  {
    $$.val = &tree.Deallocate{Name: tree.Name($2)}
  }
| DEALLOCATE PREPARE name
  {
    $$.val = &tree.Deallocate{Name: tree.Name($3)}
  }
| DEALLOCATE ALL
  {
    $$.val = &tree.Deallocate{}
  }
| DEALLOCATE PREPARE ALL
  {
    $$.val = &tree.Deallocate{}
  }
| DEALLOCATE error // SHOW HELP: DEALLOCATE

// %Help: GRANT - define access privileges and role memberships
// %Category: Priv
// %Text:
// Grant privileges:
//   GRANT {ALL | <privileges...> } ON <targets...> TO <grantees...>
// Grant role membership:
//   GRANT <roles...> TO <grantees...> [WITH ADMIN OPTION]
//
// Privileges:
//   CREATE, DROP, GRANT, SELECT, INSERT, DELETE, UPDATE, USAGE
//
// Targets:
//   DATABASE <databasename> [, ...]
//   [TABLE] [<databasename> .] { <tablename> | * } [, ...]
//   TYPE <typename> [, <typename>]...
//   SCHEMA <schemaname> [, <schemaname>]...
//
// %SeeAlso: REVOKE, WEBDOCS/grant.html
grant_stmt:
  GRANT privileges_for_cols ON targets_table TO role_spec_list opt_with_grant_option opt_granted_by
  {
    $$.val = &tree.Grant{PrivsWithCols: $2.privForColsList(), Targets: $4.targetList(), Grantees: $6.strs(), WithGrantOption: $7.bool(), GrantedBy: $8}
  }
| GRANT privileges ON targets TO role_spec_list opt_with_grant_option opt_granted_by
  {
    $$.val = &tree.Grant{Privileges: $2.privilegeList(), Targets: $4.targetList(), Grantees: $6.strs(), WithGrantOption: $7.bool(), GrantedBy: $8}
  }
| GRANT privilege_list TO role_spec_list opt_grant_role_with opt_granted_by
  {
    $$.val = &tree.GrantRole{Roles: $2.nameList(), Members: $4.strs(), WithOption: $5, GrantedBy: $6}
  }
| GRANT error // SHOW HELP: GRANT

targets:
  targets_table
  {
    $$.val = $1.targetList()
  }
| other_targets
  {
    $$.val = $1.targetList()
  }

// these can be extended to more detailed rules
other_targets:
  SEQUENCE name_list
  {
    $$.val = tree.TargetList{TargetType: privilege.Sequence, Sequences: $2.nameList()}
  }
| DATABASE name_list
  {
    $$.val = tree.TargetList{TargetType: privilege.Database, Databases: $2.nameList()}
  }
| DOMAIN name_list
  {
    $$.val = tree.TargetList{TargetType: privilege.Domain, Names: $2.nameList().ToStrings()}
  }
| FOREIGN DATA WRAPPER name_list
  {
    $$.val = tree.TargetList{TargetType: privilege.ForeignDataWrapper, Names: $4.nameList().ToStrings()}
  }
| FOREIGN SERVER name_list
  {
    $$.val = tree.TargetList{TargetType: privilege.ForeignServer, Names: $3.nameList().ToStrings()}
  }
| FUNCTION routine_with_args_list
  {
    $$.val = tree.TargetList{TargetType: privilege.Routine, Routines: $2.routines()}
  }
| PROCEDURE routine_with_args_list
  {
    $$.val = tree.TargetList{TargetType: privilege.Routine, Routines: $2.routines()}
  }
| ROUTINE routine_with_args_list
  {
    $$.val = tree.TargetList{TargetType: privilege.Routine, Routines: $2.routines()}
  }
| LANGUAGE name_list
  {
    $$.val = tree.TargetList{TargetType: privilege.Language, Names: $2.nameList().ToStrings()}
  }
| LARGE OBJECT int_expr_list
  {
    $$.val = tree.TargetList{TargetType: privilege.LargeObject, LargeObjects: $3.exprs()}
  }
| PARAMETER name_list
  {
    $$.val = tree.TargetList{TargetType: privilege.Parameter, Names: $2.nameList().ToStrings()}
  }
| SCHEMA schema_name_list
  {
    $$.val = tree.TargetList{TargetType: privilege.Schema, Names: $2.strs()}
  }
| TABLESPACE name_list
  {
    $$.val = tree.TargetList{TargetType: privilege.Tablespace, Names: $2.nameList().ToStrings()}
  }
| TYPE type_name_list
  {
    $$.val = tree.TargetList{TargetType: privilege.Type, Types: $2.unresolvedObjectNames()}
  }
| ALL SEQUENCES IN SCHEMA schema_name_list
  {
    $$.val = tree.TargetList{TargetType: privilege.Sequence, InSchema: $5.strs()}
  }
| ALL FUNCTIONS IN SCHEMA schema_name_list
  {
    $$.val = tree.TargetList{TargetType: privilege.Function, InSchema: $5.strs()}
  }
| ALL PROCEDURES IN SCHEMA schema_name_list
  {
    $$.val = tree.TargetList{TargetType: privilege.Procedure, InSchema: $5.strs()}
  }
| ALL ROUTINES IN SCHEMA schema_name_list
  {
    $$.val = tree.TargetList{TargetType: privilege.Routine, InSchema: $5.strs()}
  }
| ALL TABLES IN SCHEMA schema_name_list
  {
    $$.val = tree.TargetList{TargetType: privilege.Table, InSchema: $5.strs()}
  }

routine_with_args_list:
  routine_with_args
  {
    $$.val = []tree.Routine{$1.routine()}
  }
| routine_with_args_list ',' routine_with_args
  {
    $$.val = append($1.routines(), $3.routine())
  }

routine_with_args:
  name
  {
    $$.val = tree.Routine{Name: tree.Name($1), Args: nil}
  }
| name '(' opt_routine_args ')'
  {
    $$.val = tree.Routine{Name: tree.Name($1), Args: $3.routineArgs()}
  }

opt_with_grant_option:
  /* EMPTY */
  {
    $$.val = false
  }
| WITH GRANT OPTION
  {
    $$.val = true
  }

opt_grant_role_with:
  /* EMPTY */
  {
    $$ = ""
  }
| WITH admin_inherit_set option_true_false
  {
    $$ = string($2) + " " + string($3)
  }

admin_inherit_set:
  ADMIN
| INHERIT
| SET
  {
    $$ = $1
  }

boolean_value:
  TRUE
  {
    $$.val = true
  }
| FALSE
  {
    $$.val = false
  }
| 't'
  {
    $$.val = true
  }
| 'f'
  {
    $$.val = false
  }
| YES
  {
    $$.val = true
  }
| NO
  {
    $$.val = false
  }
| 'y'
  {
    $$.val = true
  }
| 'n'
  {
    $$.val = false
  }
| ICONST
  {
    $$.val = $1.int64() != 0
  }

option_true_false:
  OPTION
| TRUE
| FALSE
  {
    $$ = $1
  }

opt_granted_by:
  /* EMPTY */
  {
    $$ = ""
  }
| GRANTED BY role_spec
  {
    $$ = $3
  }

// %Help: REVOKE - remove access privileges and role memberships
// %Category: Priv
// %Text:
// Revoke privileges:
//   REVOKE {ALL | <privileges...> } ON <targets...> FROM <grantees...>
// Revoke role membership:
//   REVOKE [ADMIN OPTION FOR] <roles...> FROM <grantees...>
//
// Privileges:
//   CREATE, DROP, GRANT, SELECT, INSERT, DELETE, UPDATE, USAGE
//
// Targets:
//   DATABASE <databasename> [, <databasename>]...
//   [TABLE] [<databasename> .] { <tablename> | * } [, ...]
//   TYPE <typename> [, <typename>]...
//   SCHEMA <schemaname> [, <schemaname]...
//
// %SeeAlso: GRANT, WEBDOCS/revoke.html
revoke_stmt:
  REVOKE privileges_for_cols ON targets_table FROM role_spec_list opt_granted_by opt_drop_behavior
  {
    $$.val = &tree.Revoke{PrivsWithCols: $2.privForColsList(), Targets: $4.targetList(), Grantees: $6.strs(), GrantedBy: $7, DropBehavior: $8.dropBehavior()}
  }
| REVOKE GRANT OPTION FOR privileges_for_cols ON targets_table FROM role_spec_list opt_granted_by opt_drop_behavior
  {
    $$.val = &tree.Revoke{PrivsWithCols: $5.privForColsList(), Targets: $7.targetList(), Grantees: $9.strs(), GrantOptionFor: true, GrantedBy: $10, DropBehavior: $11.dropBehavior()}
  }
| REVOKE privileges ON targets FROM role_spec_list opt_granted_by opt_drop_behavior
  {
    $$.val = &tree.Revoke{Privileges: $2.privilegeList(), Targets: $4.targetList(), Grantees: $6.strs(), GrantedBy: $7, DropBehavior: $8.dropBehavior()}
  }
| REVOKE GRANT OPTION FOR privileges ON targets FROM role_spec_list opt_granted_by opt_drop_behavior
  {
    $$.val = &tree.Revoke{Privileges: $5.privilegeList(), Targets: $7.targetList(), Grantees: $9.strs(), GrantOptionFor: true, GrantedBy: $10, DropBehavior: $11.dropBehavior()}
  }
| REVOKE privilege_list FROM role_spec_list opt_granted_by opt_drop_behavior
  {
    $$.val = &tree.RevokeRole{Roles: $2.nameList(), Members: $4.strs(), GrantedBy: $5, DropBehavior: $6.dropBehavior()}
  }
| REVOKE admin_inherit_set OPTION FOR privilege_list FROM role_spec_list opt_granted_by opt_drop_behavior
  {
    $$.val = &tree.RevokeRole{Roles: $5.nameList(), Members: $7.strs(), Option: $2, GrantedBy: $8, DropBehavior: $9.dropBehavior()}
  }
| REVOKE error // SHOW HELP: REVOKE

privileges_for_cols:
  ALL '(' name_list ')'
  {
    $$.val = []tree.PrivForCols{tree.PrivForCols{Privilege: privilege.ALL, ColNames: $3.nameList()}}
  }
| ALL PRIVILEGES '(' name_list ')'
  {
    $$.val = []tree.PrivForCols{tree.PrivForCols{Privilege: privilege.ALL, ColNames: $4.nameList()}}
  }
| privilege_for_cols_list
  {
     $$.val = $1.privForColsList()
  }

privilege_for_cols_list:
  privilege_for_cols
  {
    $$.val = []tree.PrivForCols{$1.privForCols()}
  }
| privilege_for_cols_list ',' privilege_for_cols
  {
    $$.val = append($1.privForColsList(), $3.privForCols())
  }

privilege_for_cols:
  privilege '(' name_list ')'
  {
    privKind, err := privilege.KindFromString($1)
    if err != nil {
      return setErr(sqllex, err)
    }
    $$.val = tree.PrivForCols{Privilege: privKind, ColNames: $3.nameList()}
  }

// ALL is always by itself.
privileges:
  ALL
  {
    $$.val = privilege.List{privilege.ALL}
  }
| ALL PRIVILEGES
  {
    $$.val = privilege.List{privilege.ALL}
  }
| privilege_list
  {
     privList, err := privilege.ListFromStrings($1.nameList().ToStrings())
     if err != nil {
       return setErr(sqllex, err)
     }
     $$.val = privList
  }

privilege_list:
  privilege
  {
    $$.val = tree.NameList{tree.Name($1)}
  }
| privilege_list ',' privilege
  {
    $$.val = append($1.nameList(), tree.Name($3))
  }

// Privileges are parsed at execution time to avoid having to make them reserved.
// Any privileges above `col_name_keyword` should be listed here.
// The full list is in sql/privilege/privilege.go.
privilege:
  name
| SELECT
| REFERENCES
| CREATE
| CONNECT
  {
    $$ = string($1)
  }
| ALTER SYSTEM
  {
    $$ = string($1) + " " + string($2)
  }

// %Help: RESET - reset a session variable to its default value
// %Category: Cfg
// %Text: RESET [SESSION] <var>
reset_stmt:
  RESET name
  {
    name := $2
    if name == "role" {
      $$.val = &tree.SetRole{Reset: true}
    } else {
      $$.val = &tree.SetVar{Name: $2, Values:tree.Exprs{tree.DefaultVal{}}}
    }
  }
// TIME ZONE is special: it is two tokens, but is really the identifier "TIME ZONE".
| RESET TIME ZONE
  {
    $$.val = &tree.SetVar{Name: "timezone", Values:tree.Exprs{tree.DefaultVal{}}}
  }
| RESET ALL
  {
    $$.val = &tree.ResetAll{}
  }
| RESET SESSION AUTHORIZATION
  {
    $$.val = &tree.SetSessionAuthorization{}
  }
| RESET error // SHOW HELP: RESET

// USE is the MSSQL/MySQL equivalent of SET DATABASE. Alias it for convenience.
// %Help: USE - set the current database
// %Category: Cfg
// %Text: USE <dbname>
//
// "USE <dbname>" is an alias for "SET [SESSION] database = <dbname>".
// %SeeAlso: SET SESSION, WEBDOCS/set-vars.html
use_stmt:
  USE use_db_name
  {
    $$.val = &tree.SetVar{Name: "database", Values: tree.Exprs{tree.NewStrVal($2)}}
  }
| USE error // SHOW HELP: USE

// use_db_name is for the db name in USE <dbname>. We extend the simple syntax to allow for mydb/main without quoting.
use_db_name:
  SCONST
  {
    $$ = $1
  }
| db_object_name_component
  {
    $$ = $1
  }
| db_object_name_component '/' db_object_name_component
  {
    $$ = fmt.Sprintf("%s/%s", $1, $3)
  }

// SET statements including e.g. SET TRANSACTION
set_stmt:
  set_transaction_stmt // EXTEND WITH HELP: SET TRANSACTION
| set_constraints_stmt  // EXTEND WITH HELP: SET CONSTRAINTS
| set_exprs_internal   { /* SKIP DOC */ }
| set_session_or_local_stmt
| use_stmt

// %Help: SCRUB - run checks against databases or tables
// %Category: Experimental
// %Text:
// EXPERIMENTAL SCRUB TABLE <table> ...
// EXPERIMENTAL SCRUB DATABASE <database>
//
// The various checks that ca be run with SCRUB includes:
//   - Physical table data (encoding)
//   - Secondary index integrity
//   - Constraint integrity (NOT NULL, CHECK, FOREIGN KEY, UNIQUE)
// %SeeAlso: SCRUB TABLE, SCRUB DATABASE
scrub_stmt:
  scrub_table_stmt
| scrub_database_stmt
| EXPERIMENTAL SCRUB error // SHOW HELP: SCRUB

// %Help: SCRUB DATABASE - run scrub checks on a database
// %Category: Experimental
// %Text:
// EXPERIMENTAL SCRUB DATABASE <database>
//                             [AS OF SYSTEM TIME <expr>]
//
// All scrub checks will be run on the database. This includes:
//   - Physical table data (encoding)
//   - Secondary index integrity
//   - Constraint integrity (NOT NULL, CHECK, FOREIGN KEY, UNIQUE)
// %SeeAlso: SCRUB TABLE, SCRUB
scrub_database_stmt:
  EXPERIMENTAL SCRUB DATABASE database_name opt_as_of_clause
  {
    $$.val = &tree.Scrub{Typ: tree.ScrubDatabase, Database: tree.Name($4), AsOf: $5.asOfClause()}
  }
| EXPERIMENTAL SCRUB DATABASE error // SHOW HELP: SCRUB DATABASE

// %Help: SCRUB TABLE - run scrub checks on a table
// %Category: Experimental
// %Text:
// SCRUB TABLE <tablename>
//             [AS OF SYSTEM TIME <expr>]
//             [WITH OPTIONS <option> [, ...]]
//
// Options:
//   EXPERIMENTAL SCRUB TABLE ... WITH OPTIONS INDEX ALL
//   EXPERIMENTAL SCRUB TABLE ... WITH OPTIONS INDEX (<index>...)
//   EXPERIMENTAL SCRUB TABLE ... WITH OPTIONS CONSTRAINT ALL
//   EXPERIMENTAL SCRUB TABLE ... WITH OPTIONS CONSTRAINT (<constraint>...)
//   EXPERIMENTAL SCRUB TABLE ... WITH OPTIONS PHYSICAL
// %SeeAlso: SCRUB DATABASE, SRUB
scrub_table_stmt:
  EXPERIMENTAL SCRUB TABLE table_name opt_as_of_clause opt_scrub_options_clause
  {
    $$.val = &tree.Scrub{
      Typ: tree.ScrubTable,
      Table: $4.unresolvedObjectName(),
      AsOf: $5.asOfClause(),
      Options: $6.scrubOptions(),
    }
  }
| EXPERIMENTAL SCRUB TABLE error // SHOW HELP: SCRUB TABLE

opt_scrub_options_clause:
  WITH OPTIONS scrub_option_list
  {
    $$.val = $3.scrubOptions()
  }
| /* EMPTY */
  {
    $$.val = tree.ScrubOptions{}
  }

scrub_option_list:
  scrub_option
  {
    $$.val = tree.ScrubOptions{$1.scrubOption()}
  }
| scrub_option_list ',' scrub_option
  {
    $$.val = append($1.scrubOptions(), $3.scrubOption())
  }

scrub_option:
  INDEX ALL
  {
    $$.val = &tree.ScrubOptionIndex{}
  }
| INDEX '(' name_list ')'
  {
    $$.val = &tree.ScrubOptionIndex{IndexNames: $3.nameList()}
  }
| CONSTRAINT ALL
  {
    $$.val = &tree.ScrubOptionConstraint{}
  }
| CONSTRAINT '(' name_list ')'
  {
    $$.val = &tree.ScrubOptionConstraint{ConstraintNames: $3.nameList()}
  }
| PHYSICAL
  {
    $$.val = &tree.ScrubOptionPhysical{}
  }

to_or_eq:
  '='
| TO

set_exprs_internal:
  /* SET ROW serves to accelerate parser.parseExprs().
     It cannot be used by clients. */
  SET ROW '(' expr_list ')'
  {
    $$.val = &tree.SetVar{Values: $4.exprs()}
  }

// %Help: SET TRANSACTION - configure the transaction settings
// %Category: Txn
// %Text:
// SET [SESSION] TRANSACTION <txnparameters...>
//
// Transaction parameters:
//    ISOLATION LEVEL { SNAPSHOT | SERIALIZABLE }
//    PRIORITY { LOW | NORMAL | HIGH }
//    AS OF SYSTEM TIME <expr>
//    [NOT] DEFERRABLE
set_transaction_stmt:
  SET TRANSACTION transaction_mode_list
  {
    $$.val = &tree.SetTransaction{Modes: $3.transactionModes()}
  }
| SET TRANSACTION SNAPSHOT transaction_mode_list
  {
    $$.val = &tree.SetTransaction{Modes: $4.transactionModes()}
  }
| SET SESSION CHARACTERISTICS AS TRANSACTION transaction_mode_list
  {
    $$.val = &tree.SetSessionCharacteristics{Modes: $6.transactionModes()}
  }

// %Help: SET CONSTRAINTS - configure the constraints settings
// %Category: Cfg
// %Text:
// SET CONSTRAINTS { ALL | name [, ...] } { DEFERRED | IMMEDIATE }
set_constraints_stmt:
  SET CONSTRAINTS ALL DEFERRED
  {
    $$.val = &tree.SetConstraints{All: true, Deferred: true}
  }
| SET CONSTRAINTS ALL IMMEDIATE
  {
    $$.val = &tree.SetConstraints{All: true, Deferred: false}
  }
| SET CONSTRAINTS name_list DEFERRED
  {
    $$.val = &tree.SetConstraints{Names: $3.nameList(), Deferred: true}
  }
| SET CONSTRAINTS name_list IMMEDIATE
  {
    $$.val = &tree.SetConstraints{Names: $3.nameList(), Deferred: false}
  }

// %Help: SET SESSION - change a session variable
// %Category: Cfg
// %Text:
// SET [ SESSION | LOCAL ] <var> { TO | = } <values...>
// SET [ SESSION | LOCAL ] TIME ZONE <tz>
// SET [ SESSION | LOCAL ] ROLE role_name
// SET [ SESSION | LOCAL ] ROLE NONE
//
// %SeeAlso: SET TRANSACTION
// WEBDOCS/set-vars.html
set_session_or_local_stmt:
  SET set_session_or_local_cmd
  {
    $$.val = $2.stmt()
  }
| SET SESSION set_session_or_local_cmd
  {
    $$.val = $3.stmt()
  }
| SET LOCAL set_session_or_local_cmd
  {
    setStmt := $3.stmt()
    setStmt.(tree.SetStmt).SetLocalSetStmt()
    $$.val = setStmt
  }

set_session_or_local_cmd:
  set_var
| set_session_authorization
| set_role

generic_set_single_config:
  name '.' name to_or_eq var_list
  {
    $$.val = &tree.SetVar{Namespace: $1, Name: $3, Values: $5.exprs()}
  }
  // var_value includes DEFAULT expr
| name to_or_eq var_list
  {
    $$.val = &tree.SetVar{Name: $1, Values: $3.exprs()}
  }
| name FROM CURRENT
  {
    $$.val = &tree.SetVar{Name: $1, FromCurrent: true}
  }

var_list:
  var_value
  {
    $$.val = tree.Exprs{$1.expr()}
  }
| var_list ',' var_value
  {
    $$.val = append($1.exprs(), $3.expr())
  }

set_var:
  generic_set_single_config
// "SET TIME ZONE value is an alias for SET timezone TO value."
| TIME ZONE zone_value
  {
    $$.val = &tree.SetVar{Name: "timezone", Values: tree.Exprs{$3.expr()}}
  }
| set_special_syntax

set_special_syntax:
// "SET SCHEMA 'value' is an alias for SET search_path TO value. Only
// one schema can be specified using this syntax."
  SCHEMA SCONST
  {
    $$.val = &tree.SetVar{Name: "search_path", Values: tree.Exprs{tree.NewStrVal($2)}}
  }
// See comment for the non-terminal for SET NAMES below.
| set_names
// "SET SEED 'value' is an alias for SET seed TO value. Sets the internal seed
// for the random number generator (the function random)."
| SEED numeric_only
  {
    $$.val = &tree.SetVar{Name: "geqo_seed", Values: tree.Exprs{$2.expr()}}
  }

set_session_authorization:
  SESSION AUTHORIZATION DEFAULT
  {
    $$.val = &tree.SetSessionAuthorization{}
  }
| SESSION AUTHORIZATION non_reserved_word_or_sconst
  {
    $$.val = &tree.SetSessionAuthorization{Username: $3}
  }

set_role:
  ROLE non_reserved_word_or_sconst
  {
    name := $2
    if name == "none" {
      $$.val = &tree.SetRole{None: true}
    } else {
      $$.val = &tree.SetRole{Name: $2}
    }
  }

// SET NAMES is the SQL standard syntax for SET client_encoding.
// "SET NAMES value is an alias for SET client_encoding TO value."
// See https://www.postgresql.org/docs/10/static/sql-set.html
// Also see https://www.postgresql.org/docs/9.6/static/multibyte.html#AEN39236
set_names:
  NAMES SCONST
  {
    /* SKIP DOC */
    $$.val = &tree.SetVar{Name: "client_encoding", Values: tree.Exprs{tree.NewStrVal($2)}}
  }
| NAMES
  {
    /* SKIP DOC */
    $$.val = &tree.SetVar{Name: "client_encoding", Values: tree.Exprs{tree.DefaultVal{}}}
  }

var_value:
  a_expr
| extra_var_value
  {
    $$.val = tree.Expr(&tree.UnresolvedName{NumParts: 1, Parts: tree.NameParts{$1}})
  }

// The RHS of a SET statement can contain any valid expression, which
// themselves can contain identifiers like TRUE, FALSE. These are parsed
// as column names (via a_expr) and later during semantic analysis
// assigned their special value.
//
// In addition, for compatibility with CockroachDB we need to support
// the reserved keyword ON (to go along OFF, which is a valid column name).
//
// Finally, in PostgreSQL the CockroachDB-reserved words "index",
// "nothing", etc. are not special and are valid in SET. These need to
// be allowed here too.
extra_var_value:
  ON
| cockroachdb_extra_reserved_keyword

iso_level:
  READ UNCOMMITTED
  {
    $$.val = tree.SerializableIsolation
  }
| READ COMMITTED
  {
    $$.val = tree.SerializableIsolation
  }
| SNAPSHOT
  {
    $$.val = tree.SerializableIsolation
  }
| REPEATABLE READ
  {
    $$.val = tree.SerializableIsolation
  }
| SERIALIZABLE
  {
    $$.val = tree.SerializableIsolation
  }

user_priority:
  LOW
  {
    $$.val = tree.Low
  }
| NORMAL
  {
    $$.val = tree.Normal
  }
| HIGH
  {
    $$.val = tree.High
  }

// Timezone values can be:
// - a string such as 'pst8pdt'
// - an identifier such as "pst8pdt"
// - an integer or floating point number
// - a time interval per SQL99
zone_value:
  SCONST
  {
    $$.val = tree.NewStrVal($1)
  }
| IDENT
  {
    $$.val = tree.NewStrVal($1)
  }
| interval_value
  {
    $$.val = $1.expr()
  }
| numeric_only
| DEFAULT
  {
    $$.val = tree.DefaultVal{}
  }
| LOCAL
  {
    $$.val = tree.NewStrVal($1)
  }

// %Help: SHOW
// %Category: Group
// %Text:
// SHOW BACKUP, SHOW CLUSTER SETTING, SHOW COLUMNS, SHOW CONSTRAINTS,
// SHOW CREATE, SHOW DATABASES, SHOW ENUMS, SHOW HISTOGRAM, SHOW INDEXES, SHOW
// PARTITIONS, SHOW JOBS, SHOW QUERIES, SHOW RANGE, SHOW RANGES,
// SHOW ROLES, SHOW SCHEMAS, SHOW SEQUENCES, SHOW SESSION, SHOW SESSIONS,
// SHOW STATISTICS, SHOW SYNTAX, SHOW TABLES, SHOW TRACE, SHOW TRANSACTION,
// SHOW TRANSACTIONS, SHOW TYPES, SHOW USERS, SHOW LAST QUERY STATISTICS, SHOW SCHEDULES
show_stmt:
  show_backup_stmt          // EXTEND WITH HELP: SHOW BACKUP
| show_columns_stmt         // EXTEND WITH HELP: SHOW COLUMNS
| show_constraints_stmt     // EXTEND WITH HELP: SHOW CONSTRAINTS
| show_create_stmt          // EXTEND WITH HELP: SHOW CREATE
| show_databases_stmt       // EXTEND WITH HELP: SHOW DATABASES
| show_enums_stmt           // EXTEND WITH HELP: SHOW ENUMS
| show_types_stmt           // EXTEND WITH HELP: SHOW TYPES
| show_fingerprints_stmt
| show_grants_stmt          // EXTEND WITH HELP: SHOW GRANTS
| show_histogram_stmt       // EXTEND WITH HELP: SHOW HISTOGRAM
| show_indexes_stmt         // EXTEND WITH HELP: SHOW INDEXES
| show_partitions_stmt      // EXTEND WITH HELP: SHOW PARTITIONS
| show_jobs_stmt            // EXTEND WITH HELP: SHOW JOBS
| show_schedules_stmt       // EXTEND WITH HELP: SHOW SCHEDULES
| show_queries_stmt         // EXTEND WITH HELP: SHOW QUERIES
| show_roles_stmt           // EXTEND WITH HELP: SHOW ROLES
| show_savepoint_stmt       // EXTEND WITH HELP: SHOW SAVEPOINT
| show_schemas_stmt         // EXTEND WITH HELP: SHOW SCHEMAS
| show_sequences_stmt       // EXTEND WITH HELP: SHOW SEQUENCES
| show_session_stmt         // EXTEND WITH HELP: SHOW SESSION
| show_sessions_stmt        // EXTEND WITH HELP: SHOW SESSIONS
| show_stats_stmt           // EXTEND WITH HELP: SHOW STATISTICS
| show_syntax_stmt          // EXTEND WITH HELP: SHOW SYNTAX
| show_tables_stmt          // EXTEND WITH HELP: SHOW TABLES
| show_trace_stmt           // EXTEND WITH HELP: SHOW TRACE
| show_transaction_stmt     // EXTEND WITH HELP: SHOW TRANSACTION
| show_transactions_stmt    // EXTEND WITH HELP: SHOW TRANSACTIONS
| show_users_stmt           // EXTEND WITH HELP: SHOW USERS
| SHOW error                // SHOW HELP: SHOW
| show_last_query_stats_stmt // EXTEND WITH HELP: SHOW LAST QUERY STATISTICS

// Cursors are not yet supported by CockroachDB. CLOSE ALL is safe to no-op
// since there will be no open cursors.
close_cursor_stmt:
	CLOSE ALL { }
| CLOSE cursor_name { return unimplementedWithIssue(sqllex, 41412) }

declare_cursor_stmt:
	DECLARE { return unimplementedWithIssue(sqllex, 41412) }

reindex_stmt:
  REINDEX TABLE error
  {
    /* SKIP DOC */
    return purposelyUnimplemented(sqllex, "reindex table", "CockroachDB does not require reindexing.")
  }
| REINDEX INDEX error
  {
    /* SKIP DOC */
    return purposelyUnimplemented(sqllex, "reindex index", "CockroachDB does not require reindexing.")
  }
| REINDEX DATABASE error
  {
    /* SKIP DOC */
    return purposelyUnimplemented(sqllex, "reindex database", "CockroachDB does not require reindexing.")
  }
| REINDEX SYSTEM error
  {
    /* SKIP DOC */
    return purposelyUnimplemented(sqllex, "reindex system", "CockroachDB does not require reindexing.")
  }

vacuum_stmt:
  VACUUM opt_vacuum_option_list opt_vacuum_table_and_cols_list
  {
     $$.val = &tree.Vacuum{
       Options: $2.vacuumOptions(),
       TablesAndCols: $3.vacuumTableAndColsList(),
     }
  }
| VACUUM legacy_vacuum_option_list opt_vacuum_table_and_cols_list
  {
     $$.val = &tree.Vacuum{
       Options: $2.vacuumOptions(),
       TablesAndCols: $3.vacuumTableAndColsList(),
     }
  }
 

opt_vacuum_option_list:
  '(' vacuum_option_list ')' 
  {
    $$.val = $2.vacuumOptions()
  }
| /* EMPTY */
  {
    $$.val = tree.VacuumOptions(nil)
  }
  
vacuum_option_list:
  vacuum_option
  {
    $$.val = tree.VacuumOptions{$1.vacuumOption()}
  }
| vacuum_option_list ',' vacuum_option
  {
    $$.val = append($1.vacuumOptions(), $3.vacuumOption())
  }

vacuum_option:
  FULL boolean_value_for_vacuum_opt
  {
    $$.val = &tree.VacuumOption{
    	Option: "FULL",
    	Value:  $2,
    }
  }
| FREEZE boolean_value_for_vacuum_opt
  {
    $$.val = &tree.VacuumOption{
    	Option: "FREEZE",
    	Value:  $2,
    }
  }
| VERBOSE boolean_value_for_vacuum_opt
  {
    $$.val = &tree.VacuumOption{
    	Option: "VERBOSE",
    	Value:  $2,
    }
  }
| ANALYZE boolean_value_for_vacuum_opt
  {
    $$.val = &tree.VacuumOption{
    	Option: "ANALYZE",
    	Value:  $2,
    }
  }
| DISABLE_PAGE_SKIPPING boolean_value_for_vacuum_opt
  {
    $$.val = &tree.VacuumOption{
    	Option: "DISABLE_PAGE_SKIPPING",
    	Value:  $2,
    }
  }
| SKIP_LOCKED boolean_value_for_vacuum_opt
  {
    $$.val = &tree.VacuumOption{
    	Option: "SKIP_LOCKED",
    	Value:  $2,
    }
  }
| INDEX_CLEANUP auto_on_off
  {
    $$.val = &tree.VacuumOption{
    	Option: "INDEX_CLEANUP",
    	Value:  $2,
    }
  }
| PROCESS_MAIN boolean_value_for_vacuum_opt
  {
    $$.val = &tree.VacuumOption{
    	Option: "PROCESS_MAIN",
    	Value:  $2,
    }
  }
| PROCESS_TOAST boolean_value_for_vacuum_opt
  {
    $$.val = &tree.VacuumOption{
    	Option: "PROCESS_TOAST",
    	Value:  $2,
    }
  }
| TRUNCATE boolean_value_for_vacuum_opt
  {
    $$.val = &tree.VacuumOption{
    	Option: "TRUNCATE",
    	Value:  $2,
    }
  }
| PARALLEL ICONST
  {
    $$.val = &tree.VacuumOption{
    	Option: "PARALLEL",
    	Value:  $2,
    }
  }
| SKIP_DATABASE_STATS boolean_value_for_vacuum_opt
  {
    $$.val = &tree.VacuumOption{
    	Option: "SKIP_DATABASE_STATS",
    	Value:  $2,
    }
  }
| ONLY_DATABASE_STATS boolean_value_for_vacuum_opt
  {
    $$.val = &tree.VacuumOption{
    	Option: "ONLY_DATABASE_STATS",
    	Value:  $2,
    }
  }
| BUFFER_USAGE_LIMIT ICONST
  {
    $$.val = &tree.VacuumOption{
    	Option: "BUFFER_USAGE_LIMIT",
    	Value:  $2,
    }
  }

// Boolean constants for vacuum options. An empty value here is considered true
boolean_value_for_vacuum_opt:
  /* EMPTY */
  {
    $$.val = true
  }
| boolean_value
  {
    $$.val = $1
  }
  
auto_on_off:
  AUTO
  {
    $$ = "AUTO"
  }
| ON
  {
    $$ = "ON"
  }
| OFF
  {
    $$ = "OFF"
  }

legacy_vacuum_option_list:
  legacy_vacuum_option
  {
    $$.val = tree.VacuumOptions{$1.vacuumOption()}
  }
| legacy_vacuum_option_list legacy_vacuum_option
  {
    $$.val = append($1.vacuumOptions(), $2.vacuumOption())
  }
  
// this rule is actually more lenient than postgres: we can parse these four terms in any order
legacy_vacuum_option:
  FULL
  {
    $$.val = &tree.VacuumOption{
    	Option: "FULL",
    	Value:  true,
    }
  }
| FREEZE
  {
    $$.val = &tree.VacuumOption{
    	Option: "FREEZE",
    	Value:  true,
    }
  }
| VERBOSE
  {
    $$.val = &tree.VacuumOption{
    	Option: "VERBOSE",
    	Value:  true,
    }
  }
| ANALYZE
  {
    $$.val = &tree.VacuumOption{
    	Option: "ANALYZE",
    	Value:  true,
    }
  }  

opt_vacuum_table_and_cols_list:
  /* EMPTY */
  {
    $$.val = (tree.VacuumTableAndColsList)(nil)
  }
| vacuum_table_and_cols 
  {
    $$.val = tree.VacuumTableAndColsList{$1.vacuumTableAndCols()}
  }
| vacuum_table_and_cols ',' opt_vacuum_table_and_cols_list
  {
    list := $3.vacuumTableAndColsList()
    $$.val = append(list, $1.vacuumTableAndCols())
  }
  
vacuum_table_and_cols:
  table_name 
  {
    $$.val = &tree.VacuumTableAndCols{Name: $1.unresolvedObjectName()}
  }
| table_name '(' name_list ')'
  {
     $$.val = &tree.VacuumTableAndCols{
     	Name: $1.unresolvedObjectName(),
     	Cols: $3.nameList(),
     }
  }
    
// %Help: SHOW SESSION - display session variables
// %Category: Cfg
// %Text: SHOW [SESSION] { <var> | ALL }
// %SeeAlso: WEBDOCS/show-vars.html
show_session_stmt:
  SHOW session_var         { $$.val = &tree.ShowVar{Name: $2} }
| SHOW SESSION session_var { $$.val = &tree.ShowVar{Name: $3} }
| SHOW SESSION error // SHOW HELP: SHOW SESSION

session_var:
  IDENT
| IDENT '.' IDENT
  {
    $$ = $1 + "." + $3
  }
// Although ALL, SESSION_USER and DATABASE are identifiers for the
// purpose of SHOW, they lex as separate token types, so they need
// separate rules.
| ALL
| DATABASE
// SET NAMES is standard SQL for SET client_encoding.
// See https://www.postgresql.org/docs/9.6/static/multibyte.html#AEN39236
| NAMES { $$ = "client_encoding" }
| SESSION_USER
// TIME ZONE is special: it is two tokens, but is really the identifier "TIME ZONE".
| TIME ZONE { $$ = "timezone" }

// %Help: SHOW STATISTICS - display table statistics (experimental)
// %Category: Experimental
// %Text: SHOW STATISTICS [USING JSON] FOR TABLE <table_name>
//
// Returns the available statistics for a table.
// The statistics can include a histogram ID, which can
// be used with SHOW HISTOGRAM.
// If USING JSON is specified, the statistics and histograms
// are encoded in JSON format.
// %SeeAlso: SHOW HISTOGRAM
show_stats_stmt:
  SHOW STATISTICS FOR TABLE table_name
  {
    $$.val = &tree.ShowTableStats{Table: $5.unresolvedObjectName()}
  }
| SHOW STATISTICS USING JSON FOR TABLE table_name
  {
    /* SKIP DOC */
    $$.val = &tree.ShowTableStats{Table: $7.unresolvedObjectName(), UsingJSON: true}
  }
| SHOW STATISTICS error // SHOW HELP: SHOW STATISTICS

// %Help: SHOW HISTOGRAM - display histogram (experimental)
// %Category: Experimental
// %Text: SHOW HISTOGRAM <histogram_id>
//
// Returns the data in the histogram with the
// given ID (as returned by SHOW STATISTICS).
// %SeeAlso: SHOW STATISTICS
show_histogram_stmt:
  SHOW HISTOGRAM ICONST
  {
    /* SKIP DOC */
    id, err := $3.numVal().AsInt64()
    if err != nil {
      return setErr(sqllex, err)
    }
    $$.val = &tree.ShowHistogram{HistogramID: id}
  }
| SHOW HISTOGRAM error // SHOW HELP: SHOW HISTOGRAM

// %Help: SHOW BACKUP - list backup contents
// %Category: CCL
// %Text: SHOW BACKUP [SCHEMAS|FILES|RANGES] <location>
// %SeeAlso: WEBDOCS/show-backup.html
show_backup_stmt:
  SHOW BACKUPS IN string_or_placeholder
 {
    $$.val = &tree.ShowBackup{
      InCollection:    $4.expr(),
    }
  }
| SHOW BACKUP string_or_placeholder opt_with_options
  {
    $$.val = &tree.ShowBackup{
      Details: tree.BackupDefaultDetails,
      Path:    $3.expr(),
      Options: $4.kvOptions(),
    }
  }
| SHOW BACKUP string_or_placeholder IN string_or_placeholder opt_with_options
  {
    $$.val = &tree.ShowBackup{
      Details: tree.BackupDefaultDetails,
      Path:    $3.expr(),
      InCollection: $5.expr(),
      Options: $6.kvOptions(),
    }
  }
| SHOW BACKUP SCHEMAS string_or_placeholder opt_with_options
  {
    $$.val = &tree.ShowBackup{
      Details: tree.BackupDefaultDetails,
      ShouldIncludeSchemas: true,
      Path:    $4.expr(),
      Options: $5.kvOptions(),
    }
  }
| SHOW BACKUP RANGES string_or_placeholder opt_with_options
  {
    /* SKIP DOC */
    $$.val = &tree.ShowBackup{
      Details: tree.BackupRangeDetails,
      Path:    $4.expr(),
      Options: $5.kvOptions(),
    }
  }
| SHOW BACKUP FILES string_or_placeholder opt_with_options
  {
    /* SKIP DOC */
    $$.val = &tree.ShowBackup{
      Details: tree.BackupFileDetails,
      Path:    $4.expr(),
      Options: $5.kvOptions(),
    }
  }
| SHOW BACKUP error // SHOW HELP: SHOW BACKUP

// %Help: SHOW COLUMNS - list columns in relation
// %Category: DDL
// %Text: SHOW COLUMNS FROM <tablename>
// %SeeAlso: WEBDOCS/show-columns.html
show_columns_stmt:
  SHOW COLUMNS FROM table_name with_comment
  {
    $$.val = &tree.ShowColumns{Table: $4.unresolvedObjectName(), WithComment: $5.bool()}
  }
| SHOW COLUMNS error // SHOW HELP: SHOW COLUMNS

// %Help: SHOW PARTITIONS - list partition information
// %Category: DDL
// %Text: SHOW PARTITIONS FROM { TABLE <table> | INDEX <index> | DATABASE <database> }
// %SeeAlso: WEBDOCS/show-partitions.html
show_partitions_stmt:
  SHOW PARTITIONS FROM TABLE table_name
  {
    $$.val = &tree.ShowPartitions{IsTable: true, Table: $5.unresolvedObjectName()}
  }
| SHOW PARTITIONS FROM DATABASE database_name
  {
    $$.val = &tree.ShowPartitions{IsDB: true, Database: tree.Name($5)}
  }
| SHOW PARTITIONS FROM INDEX table_index_name
  {
    $$.val = &tree.ShowPartitions{IsIndex: true, Index: $5.tableIndexName()}
  }
| SHOW PARTITIONS FROM INDEX table_name '@' '*'
  {
    $$.val = &tree.ShowPartitions{IsTable: true, Table: $5.unresolvedObjectName()}
  }
| SHOW PARTITIONS error // SHOW HELP: SHOW PARTITIONS

// %Help: SHOW DATABASES - list databases
// %Category: DDL
// %Text: SHOW DATABASES
// %SeeAlso: WEBDOCS/show-databases.html
show_databases_stmt:
  SHOW DATABASES with_comment
  {
    $$.val = &tree.ShowDatabases{WithComment: $3.bool()}
  }
| SHOW DATABASES error // SHOW HELP: SHOW DATABASES

// %Help: SHOW ENUMS - list enums
// %Category: Misc
// %Text: SHOW ENUMS
show_enums_stmt:
  SHOW ENUMS
  {
    $$.val = &tree.ShowEnums{}
  }
| SHOW ENUMS error // SHOW HELP: SHOW ENUMS

// %Help: SHOW TYPES - list user defined types
// %Category: Misc
// %Text: SHOW TYPES
show_types_stmt:
  SHOW TYPES
  {
    $$.val = &tree.ShowTypes{}
  }
| SHOW TYPES error // SHOW HELP: SHOW TYPES

// %Help: SHOW GRANTS - list grants
// %Category: Priv
// %Text:
// Show privilege grants:
//   SHOW GRANTS [ON <targets...>] [FOR <users...>]
// Show role grants:
//   SHOW GRANTS ON ROLE [<roles...>] [FOR <grantees...>]
//
// %SeeAlso: WEBDOCS/show-grants.html
show_grants_stmt:
  SHOW GRANTS opt_on_targets_roles for_grantee_clause
  {
    lst := $3.targetListPtr()
    if lst != nil && lst.ForRoles {
      $$.val = &tree.ShowRoleGrants{Roles: lst.Roles, Grantees: $4.nameList()}
    } else {
      $$.val = &tree.ShowGrants{Targets: lst, Grantees: $4.nameList()}
    }
  }
| SHOW GRANTS error // SHOW HELP: SHOW GRANTS

// %Help: SHOW INDEXES - list indexes
// %Category: DDL
// %Text: SHOW INDEXES FROM { <tablename> | DATABASE <database_name> } [WITH COMMENT]
// %SeeAlso: WEBDOCS/show-index.html
show_indexes_stmt:
  SHOW INDEX FROM table_name with_comment
  {
    $$.val = &tree.ShowIndexes{Table: $4.unresolvedObjectName(), WithComment: $5.bool()}
  }
| SHOW INDEX error // SHOW HELP: SHOW INDEXES
| SHOW INDEX FROM DATABASE database_name with_comment
  {
    $$.val = &tree.ShowDatabaseIndexes{Database: tree.Name($5), WithComment: $6.bool()}
  }
| SHOW INDEXES FROM table_name with_comment
  {
    $$.val = &tree.ShowIndexes{Table: $4.unresolvedObjectName(), WithComment: $5.bool()}
  }
| SHOW INDEXES FROM DATABASE database_name with_comment
  {
    $$.val = &tree.ShowDatabaseIndexes{Database: tree.Name($5), WithComment: $6.bool()}
  }
| SHOW INDEXES error // SHOW HELP: SHOW INDEXES
| SHOW KEYS FROM table_name with_comment
  {
    $$.val = &tree.ShowIndexes{Table: $4.unresolvedObjectName(), WithComment: $5.bool()}
  }
| SHOW KEYS FROM DATABASE database_name with_comment
  {
    $$.val = &tree.ShowDatabaseIndexes{Database: tree.Name($5), WithComment: $6.bool()}
  }
| SHOW KEYS error // SHOW HELP: SHOW INDEXES

// %Help: SHOW CONSTRAINTS - list constraints
// %Category: DDL
// %Text: SHOW CONSTRAINTS FROM <tablename>
// %SeeAlso: WEBDOCS/show-constraints.html
show_constraints_stmt:
  SHOW CONSTRAINT FROM table_name
  {
    $$.val = &tree.ShowConstraints{Table: $4.unresolvedObjectName()}
  }
| SHOW CONSTRAINT error // SHOW HELP: SHOW CONSTRAINTS
| SHOW CONSTRAINTS FROM table_name
  {
    $$.val = &tree.ShowConstraints{Table: $4.unresolvedObjectName()}
  }
| SHOW CONSTRAINTS error // SHOW HELP: SHOW CONSTRAINTS

// %Help: SHOW QUERIES - list running queries
// %Category: Misc
// %Text: SHOW [ALL] [CLUSTER | LOCAL] QUERIES
// %SeeAlso: CANCEL QUERIES
show_queries_stmt:
  SHOW opt_cluster QUERIES
  {
    $$.val = &tree.ShowQueries{All: false, Cluster: $2.bool()}
  }
| SHOW opt_cluster QUERIES error // SHOW HELP: SHOW QUERIES
| SHOW ALL opt_cluster QUERIES
  {
    $$.val = &tree.ShowQueries{All: true, Cluster: $3.bool()}
  }
| SHOW ALL opt_cluster QUERIES error // SHOW HELP: SHOW QUERIES

opt_cluster:
  /* EMPTY */
  { $$.val = true }
| CLUSTER
  { $$.val = true }
| LOCAL
  { $$.val = false }

// %Help: SHOW JOBS - list background jobs
// %Category: Misc
// %Text:
// SHOW [AUTOMATIC] JOBS [select clause]
// SHOW JOBS FOR SCHEDULES [select clause]
// SHOW JOB <jobid>
// %SeeAlso: CANCEL JOBS, PAUSE JOBS, RESUME JOBS
show_jobs_stmt:
  SHOW AUTOMATIC JOBS
  {
    $$.val = &tree.ShowJobs{Automatic: true}
  }
| SHOW JOBS
  {
    $$.val = &tree.ShowJobs{Automatic: false}
  }
| SHOW AUTOMATIC JOBS error // SHOW HELP: SHOW JOBS
| SHOW JOBS error // SHOW HELP: SHOW JOBS
| SHOW JOBS select_stmt
  {
    $$.val = &tree.ShowJobs{Jobs: $3.slct()}
  }
| SHOW JOBS WHEN COMPLETE select_stmt
  {
    $$.val = &tree.ShowJobs{Jobs: $5.slct(), Block: true}
  }
| SHOW JOBS for_schedules_clause
  {
    $$.val = &tree.ShowJobs{Schedules: $3.slct()}
  }
| SHOW JOBS select_stmt error // SHOW HELP: SHOW JOBS
| SHOW JOB a_expr
  {
    $$.val = &tree.ShowJobs{
      Jobs: &tree.Select{
        Select: &tree.ValuesClause{Rows: []tree.Exprs{tree.Exprs{$3.expr()}}},
      },
    }
  }
| SHOW JOB WHEN COMPLETE a_expr
  {
    $$.val = &tree.ShowJobs{
      Jobs: &tree.Select{
        Select: &tree.ValuesClause{Rows: []tree.Exprs{tree.Exprs{$5.expr()}}},
      },
      Block: true,
    }
  }
| SHOW JOB error // SHOW HELP: SHOW JOBS

// %Help: SHOW SCHEDULES - list periodic schedules
// %Category: Misc
// %Text:
// SHOW [RUNNING | PAUSED] SCHEDULES [FOR BACKUP]
// SHOW SCHEDULE <schedule_id>
// %SeeAlso: PAUSE SCHEDULES, RESUME SCHEDULES, DROP SCHEDULES
show_schedules_stmt:
  SHOW SCHEDULES opt_schedule_executor_type
  {
    $$.val = &tree.ShowSchedules{
      WhichSchedules: tree.SpecifiedSchedules,
      ExecutorType: $3.executorType(),
    }
  }
| SHOW SCHEDULES opt_schedule_executor_type error // SHOW HELP: SHOW SCHEDULES
| SHOW schedule_state SCHEDULES opt_schedule_executor_type
  {
    $$.val = &tree.ShowSchedules{
      WhichSchedules: $2.scheduleState(),
      ExecutorType: $4.executorType(),
    }
  }
| SHOW schedule_state SCHEDULES opt_schedule_executor_type error // SHOW HELP: SHOW SCHEDULES
| SHOW SCHEDULE a_expr
  {
    $$.val = &tree.ShowSchedules{
      WhichSchedules: tree.SpecifiedSchedules,
      ScheduleID:  $3.expr(),
    }
  }
| SHOW SCHEDULE error  // SHOW HELP: SHOW SCHEDULES

schedule_state:
  RUNNING
  {
    $$.val = tree.ActiveSchedules
  }
| PAUSED
  {
    $$.val = tree.PausedSchedules
  }

opt_schedule_executor_type:
  /* Empty */
  {
    $$.val = tree.InvalidExecutor
  }
| FOR BACKUP
  {
    $$.val = tree.ScheduledBackupExecutor
  }

// %Help: SHOW TRACE - display an execution trace
// %Category: Misc
// %Text:
// SHOW [COMPACT] [KV] TRACE FOR SESSION
// %SeeAlso: EXPLAIN
show_trace_stmt:
  SHOW opt_compact TRACE FOR SESSION
  {
    $$.val = &tree.ShowTraceForSession{TraceType: tree.ShowTraceRaw, Compact: $2.bool()}
  }
| SHOW opt_compact TRACE error // SHOW HELP: SHOW TRACE
| SHOW opt_compact KV TRACE FOR SESSION
  {
    $$.val = &tree.ShowTraceForSession{TraceType: tree.ShowTraceKV, Compact: $2.bool()}
  }
| SHOW opt_compact KV error // SHOW HELP: SHOW TRACE
| SHOW opt_compact EXPERIMENTAL_REPLICA TRACE FOR SESSION
  {
    /* SKIP DOC */
    $$.val = &tree.ShowTraceForSession{TraceType: tree.ShowTraceReplica, Compact: $2.bool()}
  }
| SHOW opt_compact EXPERIMENTAL_REPLICA error // SHOW HELP: SHOW TRACE

opt_compact:
  COMPACT { $$.val = true }
| /* EMPTY */ { $$.val = false }

// %Help: SHOW SESSIONS - list open client sessions
// %Category: Misc
// %Text: SHOW [ALL] [CLUSTER | LOCAL] SESSIONS
// %SeeAlso: CANCEL SESSIONS
show_sessions_stmt:
  SHOW opt_cluster SESSIONS
  {
    $$.val = &tree.ShowSessions{Cluster: $2.bool()}
  }
| SHOW opt_cluster SESSIONS error // SHOW HELP: SHOW SESSIONS
| SHOW ALL opt_cluster SESSIONS
  {
    $$.val = &tree.ShowSessions{All: true, Cluster: $3.bool()}
  }
| SHOW ALL opt_cluster SESSIONS error // SHOW HELP: SHOW SESSIONS

// %Help: SHOW TABLES - list tables
// %Category: DDL
// %Text: SHOW TABLES [FROM <databasename> [ . <schemaname> ] ] [WITH COMMENT]
// %SeeAlso: WEBDOCS/show-tables.html
show_tables_stmt:
  SHOW TABLES FROM name '.' name with_comment
  {
    $$.val = &tree.ShowTables{ObjectNamePrefix:tree.ObjectNamePrefix{
        CatalogName: tree.Name($4),
        ExplicitCatalog: true,
        SchemaName: tree.Name($6),
        ExplicitSchema: true,
    },
    WithComment: $7.bool()}
  }
| SHOW TABLES FROM name with_comment
  {
    $$.val = &tree.ShowTables{ObjectNamePrefix:tree.ObjectNamePrefix{
        // Note: the schema name may be interpreted as database name,
        // see name_resolution.go.
        SchemaName: tree.Name($4),
        ExplicitSchema: true,
    },
    WithComment: $5.bool()}
  }
| SHOW TABLES with_comment
  {
    $$.val = &tree.ShowTables{WithComment: $3.bool()}
  }
| SHOW TABLES error // SHOW HELP: SHOW TABLES

// %Help: SHOW TRANSACTIONS - list open client transactions across the cluster
// %Category: Misc
// %Text: SHOW [ALL] [CLUSTER | LOCAL] TRANSACTIONS
show_transactions_stmt:
  SHOW opt_cluster TRANSACTIONS
  {
    $$.val = &tree.ShowTransactions{Cluster: $2.bool()}
  }
| SHOW opt_cluster TRANSACTIONS error // SHOW HELP: SHOW TRANSACTIONS
| SHOW ALL opt_cluster TRANSACTIONS
  {
    $$.val = &tree.ShowTransactions{All: true, Cluster: $3.bool()}
  }
| SHOW ALL opt_cluster TRANSACTIONS error // SHOW HELP: SHOW TRANSACTIONS

with_comment:
  WITH COMMENT { $$.val = true }
| /* EMPTY */  { $$.val = false }

// %Help: SHOW SCHEMAS - list schemas
// %Category: DDL
// %Text: SHOW SCHEMAS [FROM <databasename> ]
show_schemas_stmt:
  SHOW SCHEMAS FROM name
  {
    $$.val = &tree.ShowSchemas{Database: tree.Name($4)}
  }
| SHOW SCHEMAS
  {
    $$.val = &tree.ShowSchemas{}
  }
| SHOW SCHEMAS error // SHOW HELP: SHOW SCHEMAS

// %Help: SHOW SEQUENCES - list sequences
// %Category: DDL
// %Text: SHOW SEQUENCES [FROM <databasename> ]
show_sequences_stmt:
  SHOW SEQUENCES FROM name
  {
    $$.val = &tree.ShowSequences{Database: tree.Name($4)}
  }
| SHOW SEQUENCES
  {
    $$.val = &tree.ShowSequences{}
  }
| SHOW SEQUENCES error // SHOW HELP: SHOW SEQUENCES

// %Help: SHOW SYNTAX - analyze SQL syntax
// %Category: Misc
// %Text: SHOW SYNTAX <string>
show_syntax_stmt:
  SHOW SYNTAX SCONST
  {
    /* SKIP DOC */
    $$.val = &tree.ShowSyntax{Statement: $3}
  }
| SHOW SYNTAX error // SHOW HELP: SHOW SYNTAX

// %Help: SHOW LAST QUERY STATISTICS - display statistics for the last query issued
// %Category: Misc
// %Text: SHOW LAST QUERY STATISTICS
show_last_query_stats_stmt:
  SHOW LAST QUERY STATISTICS
  {
   /* SKIP DOC */
   $$.val = &tree.ShowLastQueryStatistics{}
  }

// %Help: SHOW SAVEPOINT - display current savepoint properties
// %Category: Cfg
// %Text: SHOW SAVEPOINT STATUS
show_savepoint_stmt:
  SHOW SAVEPOINT STATUS
  {
    $$.val = &tree.ShowSavepointStatus{}
  }
| SHOW SAVEPOINT error // SHOW HELP: SHOW SAVEPOINT

// %Help: SHOW TRANSACTION - display current transaction properties
// %Category: Cfg
// %Text: SHOW TRANSACTION {ISOLATION LEVEL | PRIORITY | STATUS}
// %SeeAlso: WEBDOCS/show-transaction.html
show_transaction_stmt:
  SHOW TRANSACTION ISOLATION LEVEL
  {
    /* SKIP DOC */
    $$.val = &tree.ShowVar{Name: "transaction_isolation"}
  }
| SHOW TRANSACTION PRIORITY
  {
    /* SKIP DOC */
    $$.val = &tree.ShowVar{Name: "transaction_priority"}
  }
| SHOW TRANSACTION STATUS
  {
    /* SKIP DOC */
    $$.val = &tree.ShowTransactionStatus{}
  }
| SHOW TRANSACTION error // SHOW HELP: SHOW TRANSACTION

// %Help: SHOW CREATE - display the CREATE statement for a table, sequence or view
// %Category: DDL
// %Text: SHOW CREATE [ TABLE | SEQUENCE | VIEW ] <tablename>
// %SeeAlso: WEBDOCS/show-create-table.html
show_create_stmt:
  SHOW CREATE table_name
  {
    $$.val = &tree.ShowCreate{Name: $3.unresolvedObjectName()}
  }
| SHOW CREATE create_kw table_name
  {
    /* SKIP DOC */
    $$.val = &tree.ShowCreate{Name: $4.unresolvedObjectName()}
  }
| SHOW CREATE error // SHOW HELP: SHOW CREATE

create_kw:
  TABLE
| VIEW
| SEQUENCE

// %Help: SHOW USERS - list defined users
// %Category: Priv
// %Text: SHOW USERS
// %SeeAlso: CREATE USER, DROP USER, WEBDOCS/show-users.html
show_users_stmt:
  SHOW USERS
  {
    $$.val = &tree.ShowUsers{}
  }
| SHOW USERS error // SHOW HELP: SHOW USERS

// %Help: SHOW ROLES - list defined roles
// %Category: Priv
// %Text: SHOW ROLES
// %SeeAlso: CREATE ROLE, ALTER ROLE, DROP ROLE
show_roles_stmt:
  SHOW ROLES
  {
    $$.val = &tree.ShowRoles{}
  }
| SHOW ROLES error // SHOW HELP: SHOW ROLES

show_fingerprints_stmt:
  SHOW EXPERIMENTAL_FINGERPRINTS FROM TABLE table_name
  {
    /* SKIP DOC */
    $$.val = &tree.ShowFingerprints{Table: $5.unresolvedObjectName()}
  }

opt_on_targets_roles:
  ON targets_roles
  {
    tmp := $2.targetList()
    $$.val = &tmp
  }
| /* EMPTY */
  {
    $$.val = (*tree.TargetList)(nil)
  }

// targets_table is a non-terminal for a list of privilege targets, a list of tables.
//
// This rule is complex and cannot be decomposed as a tree of
// non-terminals because it must resolve syntax ambiguities in the
// SHOW GRANTS ON ROLE statement. It was constructed as follows.
//
// 1. Start with the desired definition of targets:
//
//    targets ::=
//        table_pattern_list
//        TABLE table_pattern_list
//        DATABASE name_list
//
// 2. Now we must disambiguate the first rule "table_pattern_list"
//    between one that recognizes ROLE and one that recognizes
//    "<some table pattern list>". So first, inline the definition of
//    table_pattern_list.
//
//    targets ::=
//        table_pattern                          # <- here
//        table_pattern_list ',' table_pattern   # <- here
//        TABLE table_pattern_list
//        DATABASE name_list
//
// 3. We now must disambiguate the "ROLE" inside the prefix "table_pattern".
//    However having "table_pattern_list" as prefix is cumbersome, so swap it.
//
//    targets ::=
//        table_pattern
//        table_pattern ',' table_pattern_list   # <- here
//        TABLE table_pattern_list
//        DATABASE name_list
//
// 4. The rule that has table_pattern followed by a comma is now
//    non-problematic, because it will never match "ROLE" followed
//    by an optional name list (neither "ROLE;" nor "ROLE <ident>"
//    would match). We just need to focus on the first one "table_pattern".
//    This needs to tweak "table_pattern".
//
//    Here we could inline table_pattern but now we do not have to any
//    more, we just need to create a variant of it which is
//    unambiguous with a single ROLE keyword. That is, we need a
//    table_pattern which cannot contain a single name. We do
//    this as follows.
//
//    targets ::=
//        complex_table_pattern                  # <- here
//        table_pattern ',' table_pattern_list
//        TABLE table_pattern_list
//        DATABASE name_list
//    complex_table_pattern ::=
//        name '.' unrestricted_name
//        name '.' unrestricted_name '.' unrestricted_name
//        name '.' unrestricted_name '.' '*'
//        name '.' '*'
//        '*'
//
// 5. At this point the rule cannot start with a simple identifier any
//    more, keyword or not. But more importantly, any token sequence
//    that starts with ROLE cannot be matched by any of these remaining
//    rules. This means that the prefix is now free to use, without
//    ambiguity. We do this as follows, to gain a syntax rule for "ROLE
//    <namelist>". (We will handle a ROLE with no name list below.)
//
//    targets ::=
//        ROLE name_list                        # <- here
//        complex_table_pattern
//        table_pattern ',' table_pattern_list
//        TABLE table_pattern_list
//        DATABASE name_list
//
// 6. Now on to the finishing touches. First we would like to regain the
//    ability to use "<tablename>" when the table name is a simple
//    identifier. This is done as follows:
//
//    targets ::=
//        ROLE name_list
//        name                                  # <- here
//        complex_table_pattern
//        table_pattern ',' table_pattern_list
//        TABLE table_pattern_list
//        DATABASE name_list
//
// 7. Then, we want to recognize "ROLE" without any subsequent name
//    list. This requires some care: we can not add "ROLE" to the set of
//    rules above, because "name" would then overlap. To disambiguate,
//    we must first inline "name" as follows:
//
//    targets ::=
//        ROLE name_list
//        IDENT                    # <- here, always <table>
//        col_name_keyword         # <- here, always <table>
//        unreserved_keyword       # <- here, either ROLE or <table>
//        complex_table_pattern
//        table_pattern ',' table_pattern_list
//        TABLE table_pattern_list
//        DATABASE name_list
//
// 8. And now the rule is sufficiently simple that we can disambiguate
//    in the action, like this:
//
//    targets ::=
//        ...
//        unreserved_keyword {
//             if $1 == "role" { /* handle ROLE */ }
//             else { /* handle ON <tablename> */ }
//        }
//        ...
//
//   (but see the comment on the action of this sub-rule below for
//   more nuance.)
//
// Tada!
targets_table:
  IDENT
  {
    $$.val = tree.TargetList{TargetType: privilege.Table, Tables: tree.TablePatterns{&tree.UnresolvedName{NumParts:1, Parts: tree.NameParts{$1}}}}
  }
| col_name_keyword
  {
    $$.val = tree.TargetList{TargetType: privilege.Table, Tables: tree.TablePatterns{&tree.UnresolvedName{NumParts:1, Parts: tree.NameParts{$1}}}}
  }
| unreserved_keyword
  {
    // This sub-rule is meant to support both ROLE and other keywords
    // used as table name without the TABLE prefix. The keyword ROLE
    // here can have two meanings:
    //
    // - for all statements except SHOW GRANTS, it must be interpreted
    //   as a plain table name.
    // - for SHOW GRANTS specifically, it must be handled as an ON ROLE
    //   specifier without a name list (the rule with a name list is separate,
    //   see above).
    //
    // Yet we want to use a single "targets" non-terminal for all
    // statements that use targets, to share the code. This action
    // achieves this as follows:
    //
    // - for all statements (including SHOW GRANTS), it populates the
    //   Tables list in TargetList{} with the given name. This will
    //   include the given keyword as table pattern in all cases,
    //   including when the keyword was ROLE.
    //
    // - if ROLE was specified, it remembers this fact in the ForRoles
    //   field. This distinguishes `ON ROLE` (where "role" is
    //   specified as keyword), which triggers the special case in
    //   SHOW GRANTS, from `ON "role"` (where "role" is specified as
    //   identifier), which is always handled as a table name.
    //
    //   Both `ON ROLE` and `ON "role"` populate the Tables list in the same way,
    //   so that other statements than SHOW GRANTS don't observe any difference.
    //
    // Arguably this code is a bit too clever. Future work should aim
    // to remove the special casing of SHOW GRANTS altogether instead
    // of increasing (or attempting to modify) the grey magic occurring
    // here.
    $$.val = tree.TargetList{
      Tables: tree.TablePatterns{&tree.UnresolvedName{NumParts:1, Parts: tree.NameParts{$1}}},
      ForRoles: $1 == "role", // backdoor for "SHOW GRANTS ON ROLE" (no name list)
    }
  }
| complex_table_pattern
  {
    $$.val = tree.TargetList{TargetType: privilege.Table, Tables: tree.TablePatterns{$1.unresolvedName()}}
  }
| table_pattern ',' table_pattern_list
  {
    remainderPats := $3.tablePatterns()
    $$.val = tree.TargetList{TargetType: privilege.Table, Tables: append(tree.TablePatterns{$1.unresolvedName()}, remainderPats...)}
  }
| TABLE table_pattern_list
  {
    $$.val = tree.TargetList{TargetType: privilege.Table, Tables: $2.tablePatterns()}
  }

// target_roles is the variant of targets which recognizes ON ROLES
// with a name list. This cannot be included in targets_table directly
// because some statements must not recognize this syntax.
targets_roles:
  ROLE name_list
  {
     $$.val = tree.TargetList{ForRoles: true, Roles: $2.nameList()}
  }
| targets_table

for_grantee_clause:
  FOR name_list
  {
    $$.val = $2.nameList()
  }
| /* EMPTY */
  {
    $$.val = tree.NameList(nil)
  }


// %Help: PAUSE
// %Category: Misc
// %Text:
//
// Pause various background tasks and activities.
//
// PAUSE JOBS, PAUSE SCHEDULES
pause_stmt:
  pause_jobs_stmt       // EXTEND WITH HELP: PAUSE JOBS
| pause_schedules_stmt  // EXTEND WITH HELP: PAUSE SCHEDULES
| PAUSE error           // SHOW HELP: PAUSE

// %Help: RESUME
// %Category: Misc
// %Text:
//
// Resume various background tasks and activities.
//
// RESUME JOBS, RESUME SCHEDULES
resume_stmt:
  resume_jobs_stmt       // EXTEND WITH HELP: RESUME JOBS
| resume_schedules_stmt  // EXTEND WITH HELP: RESUME SCHEDULES
| RESUME error           // SHOW HELP: RESUME

// %Help: PAUSE JOBS - pause background jobs
// %Category: Misc
// %Text:
// PAUSE JOBS <selectclause>
// PAUSE JOB <jobid>
// %SeeAlso: SHOW JOBS, CANCEL JOBS, RESUME JOBS
pause_jobs_stmt:
  PAUSE JOB a_expr
  {
    $$.val = &tree.ControlJobs{
      Jobs: &tree.Select{
        Select: &tree.ValuesClause{Rows: []tree.Exprs{tree.Exprs{$3.expr()}}},
      },
      Command: tree.PauseJob,
    }
  }
| PAUSE JOB error // SHOW HELP: PAUSE JOBS
| PAUSE JOBS select_stmt
  {
    $$.val = &tree.ControlJobs{Jobs: $3.slct(), Command: tree.PauseJob}
  }
| PAUSE JOBS for_schedules_clause
  {
    $$.val = &tree.ControlJobsForSchedules{Schedules: $3.slct(), Command: tree.PauseJob}
  }
| PAUSE JOBS error // SHOW HELP: PAUSE JOBS


for_schedules_clause:
  FOR SCHEDULES select_stmt
  {
    $$.val = $3.slct()
  }
| FOR SCHEDULE a_expr
  {
   $$.val = &tree.Select{
     Select: &tree.ValuesClause{Rows: []tree.Exprs{tree.Exprs{$3.expr()}}},
   }
  }

// %Help: PAUSE SCHEDULES - pause scheduled jobs
// %Category: Misc
// %Text:
// PAUSE SCHEDULES <selectclause>
//   select clause: select statement returning schedule id to pause.
// PAUSE SCHEDULE <scheduleID>
// %SeeAlso: RESUME SCHEDULES, SHOW JOBS, CANCEL JOBS
pause_schedules_stmt:
  PAUSE SCHEDULE a_expr
  {
    $$.val = &tree.ControlSchedules{
      Schedules: &tree.Select{
        Select: &tree.ValuesClause{Rows: []tree.Exprs{tree.Exprs{$3.expr()}}},
      },
      Command: tree.PauseSchedule,
    }
  }
| PAUSE SCHEDULE error // SHOW HELP: PAUSE SCHEDULES
| PAUSE SCHEDULES select_stmt
  {
    $$.val = &tree.ControlSchedules{
      Schedules: $3.slct(),
      Command: tree.PauseSchedule,
    }
  }
| PAUSE SCHEDULES error // SHOW HELP: PAUSE SCHEDULES

// %Help: CREATE SCHEMA - create a new schema
// %Category: DDL
// %Text:
// CREATE SCHEMA [IF NOT EXISTS] { <schemaname> | [<schemaname>] AUTHORIZATION <rolename> }
create_schema_stmt:
  CREATE SCHEMA schema_name opt_schema_element_list
  {
    $$.val = &tree.CreateSchema{
      Schema: $3,
      SchemaElements: $4.stmts(),
    }
  }
| CREATE SCHEMA opt_schema_name AUTHORIZATION role_spec opt_schema_element_list
  {
    $$.val = &tree.CreateSchema{
      Schema: $3,
      AuthRole: $5,
      SchemaElements: $6.stmts(),
    }
  }
| CREATE SCHEMA IF NOT EXISTS schema_name
  {
    $$.val = &tree.CreateSchema{
      Schema: $6,
      IfNotExists: true,
    }
  }
| CREATE SCHEMA IF NOT EXISTS opt_schema_name AUTHORIZATION role_spec
  {
    $$.val = &tree.CreateSchema{
      Schema: $6,
      IfNotExists: true,
      AuthRole: $8,
    }
  }
| CREATE SCHEMA error // SHOW HELP: CREATE SCHEMA

opt_schema_element_list:
  /* EMPTY */
  {
  $$.val = nil
  }
| schema_element_list
  {
  $$.val = $1.stmts()
  }

schema_element_list:
  schema_element
  {
    $$.val = []tree.Statement{$1.stmt()}
  }
| schema_element_list schema_element
  {
    $$.val = append($1.stmts(), $2.stmt())
  }

schema_element:
  create_ddl_stmt_schema_element
| grant_stmt

// %Help: ALTER SCHEMA - alter an existing schema
// %Category: DDL
// %Text:
//
// Commands:
//   ALTER SCHEMA ... RENAME TO <newschemaname>
//   ALTER SCHEMA ... OWNER TO {<newowner> | CURRENT_ROLE | CURRENT_USER | SESSION_USER }
alter_schema_stmt:
  ALTER SCHEMA schema_name RENAME TO schema_name
  {
    $$.val = &tree.AlterSchema{
      Schema: $3,
      Cmd: &tree.AlterSchemaRename{
        NewName: $6,
      },
    }
  }
| ALTER SCHEMA schema_name owner_to
  {
    $$.val = &tree.AlterSchema{
      Schema: $3,
      Cmd: &tree.AlterSchemaOwner{
        Owner: $4,
      },
    }
  }
| ALTER SCHEMA error // SHOW HELP: ALTER SCHEMA

// %Help: CREATE TABLE - create a new table
// %Category: DDL
// %Text:
// CREATE [[GLOBAL | LOCAL] {TEMPORARY | TEMP}] TABLE [IF NOT EXISTS] <tablename> ( <elements...> ) [<interleave>] [<on_commit>]
// CREATE [[GLOBAL | LOCAL] {TEMPORARY | TEMP}] TABLE [IF NOT EXISTS] <tablename> [( <colnames...> )] AS <source> [<interleave>] [<on commit>]
//
// Table elements:
//    <name> <type> [<qualifiers...>]
//    [UNIQUE | INVERTED] INDEX [<name>] ( <colname> [ASC | DESC] [, ...] )
//                            [USING HASH WITH BUCKET_COUNT = <shard_buckets>] [{STORING | INCLUDE | COVERING} ( <colnames...> )] [<interleave>]
//    FAMILY [<name>] ( <colnames...> )
//    [CONSTRAINT <name>] <constraint>
//
// Table constraints:
//    PRIMARY KEY ( <colnames...> ) [USING HASH WITH BUCKET_COUNT = <shard_buckets>]
//    FOREIGN KEY ( <colnames...> ) REFERENCES <tablename> [( <colnames...> )] [ON DELETE {NO ACTION | RESTRICT}] [ON UPDATE {NO ACTION | RESTRICT}]
//    UNIQUE ( <colnames... ) [{STORING | INCLUDE | COVERING} ( <colnames...> )] [<interleave>]
//    CHECK ( <expr> )
//
// Column qualifiers:
//   [CONSTRAINT <constraintname>] {NULL | NOT NULL | UNIQUE | PRIMARY KEY | CHECK (<expr>) | DEFAULT <expr>}
//   FAMILY <familyname>, CREATE [IF NOT EXISTS] FAMILY [<familyname>]
//   REFERENCES <tablename> [( <colnames...> )] [ON DELETE {NO ACTION | RESTRICT}] [ON UPDATE {NO ACTION | RESTRICT}]
//   COLLATE <collationname>
//   AS ( <expr> ) STORED
//
// Interleave clause:
//    INTERLEAVE IN PARENT <tablename> ( <colnames...> ) [CASCADE | RESTRICT]
//
// On commit clause:
//    ON COMMIT {PRESERVE ROWS | DROP | DELETE ROWS}
//
// %SeeAlso: SHOW TABLES, CREATE VIEW, SHOW CREATE,
// WEBDOCS/create-table.html
// WEBDOCS/create-table-as.html
create_table_stmt:
  CREATE opt_persistence_temp_table TABLE table_name '(' opt_table_elem_list ')' opt_inherits opt_partition_by opt_using_method opt_table_with opt_create_table_on_commit opt_tablespace
  {
    $$.val = &tree.CreateTable{
      Table: $4.unresolvedObjectName().ToTableName(),
      IfNotExists: false,
      Defs: $6.tblDefs(),
      Inherits: $8.tableNames(),
      PartitionBy: $9.partitionBy(),
      Persistence: $2.persistence(),
      Using: $10,
      StorageParams: $11.storageParams(),
      OnCommit: $12.createTableOnCommitSetting(),
      Tablespace: tree.Name($13),
    }
  }
| CREATE opt_persistence_temp_table TABLE IF NOT EXISTS table_name '(' opt_table_elem_list ')' opt_inherits opt_partition_by opt_using_method opt_table_with opt_create_table_on_commit opt_tablespace
  {
    $$.val = &tree.CreateTable{
      Table: $7.unresolvedObjectName().ToTableName(),
      IfNotExists: true,
      Defs: $9.tblDefs(),
      Inherits: $11.tableNames(),
      PartitionBy: $12.partitionBy(),
      Persistence: $2.persistence(),
      Using: $13,
      StorageParams: $14.storageParams(),
      OnCommit: $15.createTableOnCommitSetting(),
      Tablespace: tree.Name($16),
    }
  }
| CREATE opt_persistence_temp_table TABLE table_name OF type_name opt_table_of_elem_list opt_partition_by opt_using_method opt_table_with opt_create_table_on_commit opt_tablespace
  {
    $$.val = &tree.CreateTable{
      Table: $4.unresolvedObjectName().ToTableName(),
      IfNotExists: false,
      OfType: $6.unresolvedObjectName(),
      Defs: $7.tblDefs(),
      PartitionBy: $8.partitionBy(),
      Persistence: $2.persistence(),
      Using: $9,
      StorageParams: $10.storageParams(),
      OnCommit: $11.createTableOnCommitSetting(),
      Tablespace: tree.Name($12),
    }
  }
| CREATE opt_persistence_temp_table TABLE IF NOT EXISTS table_name OF type_name opt_table_of_elem_list opt_partition_by opt_using_method opt_table_with opt_create_table_on_commit opt_tablespace
  {
    $$.val = &tree.CreateTable{
      Table: $7.unresolvedObjectName().ToTableName(),
      IfNotExists: true,
      OfType: $9.unresolvedObjectName(),
      Defs: $10.tblDefs(),
      PartitionBy: $11.partitionBy(),
      Persistence: $2.persistence(),
      Using: $12,
      StorageParams: $13.storageParams(),
      OnCommit: $14.createTableOnCommitSetting(),
      Tablespace: tree.Name($15),
    }
  }
| CREATE opt_persistence_temp_table TABLE table_name PARTITION OF table_name opt_table_of_elem_list partition_of opt_partition_by opt_using_method opt_table_with opt_create_table_on_commit opt_tablespace
  {
    $$.val = &tree.CreateTable{
      Table: $4.unresolvedObjectName().ToTableName(),
      IfNotExists: false,
      PartitionOf: $7.unresolvedObjectName().ToTableName(),
      PartitionBoundSpec: $9.partitionBoundSpec(),
      Defs: $8.tblDefs(),
      PartitionBy: $10.partitionBy(),
      Persistence: $2.persistence(),
      Using: $11,
      StorageParams: $12.storageParams(),
      OnCommit: $13.createTableOnCommitSetting(),
      Tablespace: tree.Name($14),
    }
  }
| CREATE opt_persistence_temp_table TABLE IF NOT EXISTS table_name PARTITION OF table_name opt_table_of_elem_list partition_of opt_partition_by opt_using_method opt_table_with opt_create_table_on_commit opt_tablespace
  {
    $$.val = &tree.CreateTable{
      Table: $7.unresolvedObjectName().ToTableName(),
      IfNotExists: true,
      PartitionOf: $10.unresolvedObjectName().ToTableName(),
      PartitionBoundSpec: $12.partitionBoundSpec(),
      Defs: $11.tblDefs(),
      PartitionBy: $13.partitionBy(),
      Persistence: $2.persistence(),
      Using: $14,
      StorageParams: $15.storageParams(),
      OnCommit: $16.createTableOnCommitSetting(),
      Tablespace: tree.Name($17),
    }
  }

partition_of:
  FOR VALUES partition_bound_spec
  {
    $$.val = $3.partitionBoundSpec()
  }
| DEFAULT
  {
    $$.val = tree.PartitionBoundSpec{IsDefault: true}
  }

partition_bound_spec:
  IN '(' partition_bound_expr_list ')'
  {
    $$.val = tree.PartitionBoundSpec{Type: tree.PartitionBoundIn, From: $3.exprs()}
  }
| FROM '(' partition_bound_expr_list ')' TO '(' partition_bound_expr_list ')'
  {
    $$.val = tree.PartitionBoundSpec{Type: tree.PartitionBoundFromTo, From: $3.exprs(), To: $7.exprs()}
  }
| WITH '(' MODULUS ICONST ',' REMAINDER ICONST ')'
  {
    $$.val = tree.PartitionBoundSpec{Type: tree.PartitionBoundWith, From: tree.Exprs{$4.expr()}, To: tree.Exprs{$7.expr()}}
  }

partition_bound_expr_list:
  partition_bound_expr
  {
    $$.val = tree.Exprs{$1.expr()}
  }
| partition_bound_expr_list ',' partition_bound_expr
  {
    $$.val = append($1.exprs(), $3.expr())
  }

partition_bound_expr:
  a_expr

opt_table_of_elem_list:
  '(' table_of_elem_list ')'
  {
    $$.val = $2.tblDefs()
  }
| /* EMPTY */
  {
    $$.val = tree.TableDefs(nil)
  }

table_of_elem_list:
  table_of_elem
  {
    $$.val = tree.TableDefs{$1.tblDef()}
  }
| table_of_elem_list ',' table_of_elem
  {
    $$.val = append($1.tblDefs(), $3.tblDef())
  }

table_of_elem:
  column_name opt_col_with_options col_constraint_list
  {
    tableDef, err := tree.NewColumnTableDef(tree.Name($1), nil, "", "", $3.colQuals())
    if err != nil {
      return setErr(sqllex, err)
    }
    $$.val = tableDef
  }
| table_constraint
  {
    $$.val = $1.constraintDef()
  }

opt_col_with_options:
  /* EMPTY */
  {
    $$.val = false
  }
| WITH OPTIONS
  {
    $$.val = true
  }

opt_inherits:
  /* EMPTY */
  {
    $$.val = tree.TableNames(nil)
  }
| INHERITS '(' table_name_list ')'
  {
    $$.val = $3.tableNames()
  }

opt_table_with:
  opt_with_storage_parameter_list
| WITHOUT OIDS
  {
    /* SKIP DOC */
    /* this is also the default in CockroachDB */
    $$.val = nil
  }
| WITH OIDS error
  {
    return unimplemented(sqllex, "create table with oids")
  }

opt_with_storage_parameter_list:
  /* EMPTY */
  {
    $$.val = nil
  }
| WITH '(' storage_parameter_list ')'
  {
    /* SKIP DOC */
    $$.val = $3.storageParams()
  }

opt_create_table_on_commit:
  /* EMPTY */
  {
    $$.val = tree.CreateTableOnCommitUnset
  }
| ON COMMIT PRESERVE ROWS
  {
    $$.val = tree.CreateTableOnCommitPreserveRows
  }
| ON COMMIT DELETE ROWS error
  {
    $$.val = tree.CreateTableOnCommitDeleteRows
  }
| ON COMMIT DROP error
  {
    $$.val = tree.CreateTableOnCommitDrop
  }

storage_parameter:
  name opt_var_value
  {
    $$.val = tree.StorageParam{Key: tree.Name($1), Value: $2.expr()}
  }
|  SCONST opt_var_value
  {
    $$.val = tree.StorageParam{Key: tree.Name($1), Value: $2.expr()}
  }

opt_var_value:
  /* EMPTY */
  {
    $$.val = nil
  }
| '=' var_value
  {
    $$.val = $2.expr()
  }

storage_parameter_list:
  storage_parameter
  {
    $$.val = []tree.StorageParam{$1.storageParam()}
  }
|  storage_parameter_list ',' storage_parameter
  {
    $$.val = append($1.storageParams(), $3.storageParam())
  }

create_table_as_stmt:
  CREATE opt_persistence_temp_table TABLE table_name opt_index_params_name_only opt_using_method opt_table_with opt_create_table_on_commit opt_tablespace AS select_stmt opt_create_as_with_data
  {
    name := $4.unresolvedObjectName().ToTableName()
    tblDefs := tree.ConvertIdxElemsToTblDefsForColumnNameOnly($5.idxElems())
    $$.val = &tree.CreateTable{
      Table: name,
      Defs: tblDefs,
      Using: $6,
      Tablespace: tree.Name($9),
      AsSource: $11.slct(),
      StorageParams: $7.storageParams(),
      OnCommit: $8.createTableOnCommitSetting(),
      Persistence: $2.persistence(),
      WithNoData: $12.bool(),
    }
  }
| CREATE opt_persistence_temp_table TABLE IF NOT EXISTS table_name opt_index_params_name_only opt_using_method opt_table_with opt_create_table_on_commit opt_tablespace AS select_stmt opt_create_as_with_data
  {
    name := $7.unresolvedObjectName().ToTableName()
    tblDefs := tree.ConvertIdxElemsToTblDefsForColumnNameOnly($8.idxElems())
    $$.val = &tree.CreateTable{
      Table: name,
      IfNotExists: true,
      Using: $9,
      Tablespace: tree.Name($12),
      Defs: tblDefs,
      AsSource: $14.slct(),
      StorageParams: $10.storageParams(),
      OnCommit: $11.createTableOnCommitSetting(),
      Persistence: $2.persistence(),
      WithNoData: $15.bool(),
    }
  }

opt_create_as_with_data:
  /* EMPTY */
  {
    $$.val = false
  }
| WITH DATA
  {
    $$.val = false
  }
| WITH NO DATA
  {
    $$.val = true
  }

/*
 * Redundancy here is needed to avoid shift/reduce conflicts,
 * since TEMP is not a reserved word.  See also OptTempTableName.
 *
 * NOTE: we accept both GLOBAL and LOCAL options.  They currently do nothing,
 * but future versions might consider GLOBAL to request SQL-spec-compliant
 * temp table behavior.  Since we have no modules the
 * LOCAL keyword is really meaningless; furthermore, some other products
 * implement LOCAL as meaning the same as our default temp table behavior,
 * so we'll probably continue to treat LOCAL as a noise word.
 *
 * NOTE: PG only accepts GLOBAL/LOCAL keywords for temp tables -- not sequences
 * and views. These keywords are no-ops in PG. This behavior is replicated by
 * making the distinction between opt_temp and opt_persistence_temp_table.
 */
 opt_temp:
  TEMPORARY         { $$.val = tree.PersistenceTemporary }
| TEMP              { $$.val = tree.PersistenceTemporary }
| /*EMPTY*/         { $$.val = tree.PersistencePermanent }

opt_persistence_sequence:
  opt_temp
| UNLOGGED          { $$.val = tree.PersistenceUnlogged }

opt_persistence_temp_table:
  opt_persistence_sequence
| LOCAL TEMPORARY   { $$.val = tree.PersistenceTemporary }
| LOCAL TEMP        { $$.val = tree.PersistenceTemporary }
| GLOBAL TEMPORARY  { $$.val = tree.PersistenceTemporary }
| GLOBAL TEMP       { $$.val = tree.PersistenceTemporary }

opt_table_elem_list:
  table_elem_list
| /* EMPTY */
  {
    $$.val = tree.TableDefs(nil)
  }

table_elem_list:
  table_elem
  {
    $$.val = tree.TableDefs{$1.tblDef()}
  }
| table_elem_list ',' table_elem
  {
    $$.val = append($1.tblDefs(), $3.tblDef())
  }

table_elem:
  create_table_column_def
  {
    $$.val = $1.colDef()
  }
| table_constraint
  {
    $$.val = $1.constraintDef()
  }
| LIKE table_name like_table_option_list
  {
    $$.val = &tree.LikeTableDef{
      Name: $2.unresolvedObjectName().ToTableName(),
      Options: $3.likeTableOptionList(),
    }
  }

like_table_option_list:
  like_table_option_list INCLUDING like_table_option
  {
    $$.val = append($1.likeTableOptionList(), $3.likeTableOption())
  }
| like_table_option_list EXCLUDING like_table_option
  {
    opt := $3.likeTableOption()
    opt.Excluded = true
    $$.val = append($1.likeTableOptionList(), opt)
  }
| /* EMPTY */
  {
    $$.val = []tree.LikeTableOption(nil)
  }

like_table_option:
  COMMENTS			{ return unimplementedWithIssueDetail(sqllex, 47071, "like table in/excluding comments") }
| CONSTRAINTS		{ $$.val = tree.LikeTableOption{Opt: tree.LikeTableOptConstraints} }
| DEFAULTS			{ $$.val = tree.LikeTableOption{Opt: tree.LikeTableOptDefaults} }
| IDENTITY	  	{ return unimplementedWithIssueDetail(sqllex, 47071, "like table in/excluding identity") }
| GENERATED			{ $$.val = tree.LikeTableOption{Opt: tree.LikeTableOptGenerated} }
| INDEXES			{ $$.val = tree.LikeTableOption{Opt: tree.LikeTableOptIndexes} }
| STATISTICS		{ return unimplementedWithIssueDetail(sqllex, 47071, "like table in/excluding statistics") }
| STORAGE			{ return unimplementedWithIssueDetail(sqllex, 47071, "like table in/excluding storage") }
| ALL				{ $$.val = tree.LikeTableOption{Opt: tree.LikeTableOptAll} }

opt_partition_by:
  partition_by
| /* EMPTY */
  {
    $$.val = (*tree.PartitionBy)(nil)
  }

partition_by:
  PARTITION BY partition_by_type '(' partition_index_params ')'
  {
    $$.val = &tree.PartitionBy{Type: $3.partitionByType(), Elems: $5.idxElems()}
  }

partition_by_type:
  LIST
  {
    $$.val = tree.PartitionByList
  }
| RANGE
  {
    $$.val = tree.PartitionByRange
  }
| HASH
  {
    $$.val = tree.PartitionByHash
  }

partition_index_params:
  partition_index_elem
  {
    $$.val = tree.IndexElemList{$1.idxElem()}
  }
| partition_index_params ',' partition_index_elem
  {
    $$.val = append($1.idxElems(), $3.idxElem())
  }

partition_index_elem:
  name opt_collate opt_opclass
  {
    $$.val = tree.IndexElem{Column: tree.Name($1), Collation: $2.unresolvedObjectName().UnquotedString(), OpClass: $3.opClass()}
  }
| '(' a_expr ')' opt_collate opt_opclass
  {
    $$.val = tree.IndexElem{Expr: $2.expr(), Collation: $4.unresolvedObjectName().UnquotedString(), OpClass: $5.opClass()}
  }

alter_column_def:
  column_name typename opt_collate col_constraint_list
  {
    tableDef, err := tree.NewColumnTableDef(tree.Name($1), $2.typeReference(), "", $3.unresolvedObjectName().UnquotedString(), $4.colQuals())
    if err != nil {
      return setErr(sqllex, err)
    }
    $$.val = tableDef
  }

// Treat SERIAL pseudo-types as separate case so that types.T does not have to
// support them as first-class types (e.g. they should not be supported as CAST
// target types).
create_table_column_def:
  column_name typename opt_compression opt_collate col_constraint_list
  {
    tableDef, err := tree.NewColumnTableDef(tree.Name($1), $2.typeReference(), $3, $4.unresolvedObjectName().UnquotedString(), $5.colQuals())
    if err != nil {
      return setErr(sqllex, err)
    }
    $$.val = tableDef
  }

opt_compression:
  /* EMPTY */
  {
    $$ = ""
  }
| COMPRESSION unrestricted_name
  {
    $$ = $2
  }

col_constraint_list:
  col_constraint_list col_qualification
  {
    $$.val = append($1.colQuals(), $2.colQual())
  }
| /* EMPTY */
  {
    $$.val = []tree.NamedColumnQualification(nil)
  }

col_qualification:
  CONSTRAINT constraint_name col_qualification_elem opt_deferrable_mode opt_initially
  {
    $$.val = tree.NamedColumnQualification{Name: tree.Name($2), Qualification: $3.colQualElem(), Deferrable: $4.deferrableMode(), Initially: $5.initiallyMode()}
  }
| col_qualification_elem opt_deferrable_mode opt_initially
  {
    $$.val = tree.NamedColumnQualification{Qualification: $1.colQualElem(), Deferrable: $2.deferrableMode(), Initially: $3.initiallyMode()}
  }

// DEFAULT NULL is already the default for Postgres. But define it here and
// carry it forward into the system to make it explicit.
// - thomas 1998-09-13
//
// WITH NULL and NULL are not SQL-standard syntax elements, so leave them
// out. Use DEFAULT NULL to explicitly indicate that a column may have that
// value. WITH NULL leads to shift/reduce conflicts with WITH TIME ZONE anyway.
// - thomas 1999-01-08
//
// DEFAULT expression must be b_expr not a_expr to prevent shift/reduce
// conflict on NOT (since NOT might start a subsequent NOT NULL constraint, or
// be part of a_expr NOT LIKE or similar constructs).
col_qualification_elem:
  NOT NULL
  {
    $$.val = tree.NotNullConstraint{}
  }
| NULL
  {
    $$.val = tree.NullConstraint{}
  }
| CHECK '(' a_expr ')' opt_no_inherit
  {
    $$.val = &tree.ColumnCheckConstraint{Expr: $3.expr(), NoInherit: $5.bool()}
  }
| DEFAULT b_expr
  {
    $$.val = &tree.ColumnDefault{Expr: $2.expr()}
  }
| GENERATED_ALWAYS ALWAYS AS '(' a_expr ')' STORED
  {
    $$.val = &tree.ColumnComputedDef{Expr: $5.expr()}
  }
| col_qual_generated_identity
| UNIQUE opt_nulls_distinct constraint_index_params
  {
    $$.val = tree.UniqueConstraint{
      NullsDistinct: $2.bool(),
      IndexParams: $3.constraintIdxParams(),
    }
  }
| PRIMARY KEY constraint_index_params
  {
    $$.val = tree.UniqueConstraint{
      IsPrimary: true,
      IndexParams: $3.constraintIdxParams(),
    }
  }
| REFERENCES table_name opt_name_parens key_match reference_actions
  {
    name := $2.unresolvedObjectName().ToTableName()
    $$.val = &tree.ColumnFKConstraint{
      Table: name,
      Col: tree.Name($3),
      Actions: $5.referenceActions(),
      Match: $4.compositeKeyMatchMethod(),
    }
  }

col_qual_generated_identity:
  GENERATED_ALWAYS ALWAYS AS IDENTITY opt_create_seq_option_list_with_parens
  {
    $$.val = &tree.ColumnComputedDef{Options: $5.seqOpts()}
  }
| GENERATED BY DEFAULT AS IDENTITY opt_create_seq_option_list_with_parens
  {
    $$.val = &tree.ColumnComputedDef{ByDefault: true, Options: $6.seqOpts()}
  }

opt_include_index_cols:
  /* EMPTY */
  {
    $$.val = tree.IndexElemList(nil)
  }
| INCLUDE '(' index_params_name_only ')'
  {
    $$.val = $3.idxElems()
  }

opt_using_index_tablespace:
  /* EMPTY */
  {
    $$ = ""
  }
| USING INDEX TABLESPACE tablespace_name
  {
    $$ = $4
  }

opt_create_seq_option_list_with_parens:
  '(' create_seq_option_list ')'
  {
    $$.val = $2.seqOpts()
  }
| /* EMPTY */
  {
    $$.val = []tree.SequenceOption(nil)
  }

opt_no_inherit:
  /* EMPTY */
  {
    $$.val = false
  }
| NO INHERIT
  {
    $$.val = true
  }

table_constraint:
  CONSTRAINT constraint_name table_constraint_elem opt_deferrable_mode opt_initially
  {
    $$.val = $3.constraintDef()
    $$.val.(tree.ConstraintTableDef).SetName(tree.Name($2))
  }
| table_constraint_elem opt_deferrable_mode opt_initially
  {
    $$.val = $1.constraintDef()
  }

// table_constraint_elem specifies constraint syntax which is not embedded into a
// column definition. col_qualification_elem specifies the embedded form.
// - thomas 1997-12-03
table_constraint_elem:
  CHECK '(' a_expr ')' opt_no_inherit
  {
    $$.val = &tree.CheckConstraintTableDef{
      Expr: $3.expr(),
      NoInherit: $5.bool(),
    }
  }
| UNIQUE opt_nulls_distinct '(' index_params ')' constraint_index_params
  {
    $$.val = &tree.UniqueConstraintTableDef{
      IndexTableDef: tree.IndexTableDef{
        Columns: $4.idxElems(),
        IndexParams: $6.constraintIdxParams(),
      },
    }
  }
| PRIMARY KEY '(' index_params ')' constraint_index_params
  {
    $$.val = &tree.UniqueConstraintTableDef{
      IndexTableDef: tree.IndexTableDef{
        Columns: $4.idxElems(),
	IndexParams: $6.constraintIdxParams(),
      },
      PrimaryKey: true,
    }
  }
| EXCLUDE opt_using_method '(' exclude_elems ')' constraint_index_params opt_where_clause_paren
  {
    $$.val = &tree.ExcludeConstraintTableDef{
      IndexTableDef: tree.IndexTableDef{
        Columns: $4.idxElems(),
	IndexParams: $6.constraintIdxParams(),
      },
      Predicate: $7.expr(),
    }
  }
| FOREIGN KEY '(' name_list ')' REFERENCES table_name opt_column_list key_match reference_actions
  {
    name := $7.unresolvedObjectName().ToTableName()
    $$.val = &tree.ForeignKeyConstraintTableDef{
      Table: name,
      FromCols: $4.nameList(),
      ToCols: $8.nameList(),
      Match: $9.compositeKeyMatchMethod(),
      Actions: $10.referenceActions(),
    }
  }

exclude_elems:
  index_elem WITH math_op
  {
    el := $1.idxElem()
    el.ExcludeOp = $3.op()
    $$.val = tree.IndexElemList{el}
  }
| exclude_elems ',' index_elem WITH math_op
  {
    el := $3.idxElem()
    el.ExcludeOp = $5.op()
    $$.val = append($1.idxElems(), el)
  }

opt_index_params_name_only:
  /* EMPTY */
  {
    $$.val = tree.IndexElemList(nil)
  }
| '(' index_params_name_only ')'
  {
    $$.val = $2.idxElems()
  }

index_params_name_only:
  index_elem_name_only
  {
    $$.val = tree.IndexElemList{$1.idxElem()}
  }
| index_params_name_only ',' index_elem_name_only
  {
    $$.val = append($1.idxElems(), $3.idxElem())
  }

index_elem_name_only:
  column_name
  {
    $$.val = tree.IndexElem{Column: tree.Name($1)}
  }

constraint_index_params:
  opt_include_index_cols opt_with_storage_parameter_list opt_using_index_tablespace
  {
    $$.val = tree.IndexParams{
      IncludeColumns: $1.idxElems(),
      StorageParams:  $2.storageParams(),
      Tablespace:     tree.Name($3),
    }
  }

opt_deferrable_mode:
  /* EMPTY */
  {
    $$.val = tree.UnspecifiedDeferrableMode
  }
| deferrable_mode

opt_initially:
  /* EMPTY */
  {
    $$.val = tree.UnspecifiedInitiallyMode
  }
| INITIALLY DEFERRED
  {
    $$.val = tree.InitiallyDeferred
  }
| INITIALLY IMMEDIATE
  {
    $$.val = tree.InitiallyImmediate
  }

opt_column_list:
  '(' name_list ')'
  {
    $$.val = $2.nameList()
  }
| /* EMPTY */
  {
    $$.val = tree.NameList(nil)
  }

opt_only:
  /* EMPTY */
  {
    $$.val = false
  }
| ONLY
  {
    $$.val = true
  }

opt_nulls_distinct:
  NULLS opt_not DISTINCT
  {
    $$.val = $2.bool()
  }
| /* EMPTY */
  {
    $$.val = true
  }

// https://www.postgresql.org/docs/10/sql-createtable.html
//
// "A value inserted into the referencing column(s) is matched against
// the values of the referenced table and referenced columns using the
// given match type. There are three match types: MATCH FULL, MATCH
// PARTIAL, and MATCH SIMPLE (which is the default). MATCH FULL will
// not allow one column of a multicolumn foreign key to be null unless
// all foreign key columns are null; if they are all null, the row is
// not required to have a match in the referenced table. MATCH SIMPLE
// allows any of the foreign key columns to be null; if any of them
// are null, the row is not required to have a match in the referenced
// table. MATCH PARTIAL is not yet implemented. (Of course, NOT NULL
// constraints can be applied to the referencing column(s) to prevent
// these cases from arising.)"
key_match:
  MATCH SIMPLE
  {
    $$.val = tree.MatchSimple
  }
| MATCH FULL
  {
    $$.val = tree.MatchFull
  }
| MATCH PARTIAL
  {
    $$.val = tree.MatchPartial
  }
| /* EMPTY */
  {
    $$.val = tree.MatchSimple
  }

// We combine the update and delete actions into one value temporarily for
// simplicity of parsing, and then break them down again in the calling
// production.
reference_actions:
  reference_on_update
  {
     $$.val = tree.ReferenceActions{Update: $1.refAction()}
  }
| reference_on_delete
  {
     $$.val = tree.ReferenceActions{Delete: $1.refAction()}
  }
| reference_on_update reference_on_delete
  {
    $$.val = tree.ReferenceActions{Update: $1.refAction(), Delete: $2.refAction()}
  }
| reference_on_delete reference_on_update
  {
    $$.val = tree.ReferenceActions{Delete: $1.refAction(), Update: $2.refAction()}
  }
| /* EMPTY */
  {
    $$.val = tree.ReferenceActions{}
  }

reference_on_update:
  ON UPDATE reference_action
  {
    $$.val = $3.refAction()
  }

reference_on_delete:
  ON DELETE reference_action
  {
    $$.val = $3.refAction()
  }

reference_action:
// NO ACTION is currently the default behavior. It is functionally the same as
// RESTRICT.
  NO ACTION
  {
    $$.val = tree.RefAction{Action: tree.NoAction}
  }
| RESTRICT
  {
    $$.val = tree.RefAction{Action: tree.Restrict}
  }
| CASCADE
  {
    $$.val = tree.RefAction{Action: tree.Cascade}
  }
| SET NULL opt_column_list
  {
    $$.val = tree.RefAction{Action: tree.SetNull, Columns: $3.nameList()}
  }
| SET DEFAULT opt_column_list
  {
    $$.val = tree.RefAction{Action: tree.SetDefault, Columns: $3.nameList()}
  }

// %Help: CREATE SEQUENCE - create a new sequence
// %Category: DDL
// %Text:
// CREATE [ { TEMPORARY | TEMP } | UNLOGGED ] SEQUENCE [ IF NOT EXISTS ] name
//    [ AS data_type ]
//    [ INCREMENT [ BY ] increment ]
//    [ MINVALUE minvalue | NO MINVALUE ] [ MAXVALUE maxvalue | NO MAXVALUE ]
//    [ START [ WITH ] start ] [ CACHE cache ] [ [ NO ] CYCLE ]
//    [ OWNED BY { table_name.column_name | NONE } ]
//
// %SeeAlso: CREATE TABLE
create_sequence_stmt:
  CREATE opt_persistence_sequence SEQUENCE sequence_name opt_create_seq_option_list
  {
    name := $4.unresolvedObjectName().ToTableName()
    $$.val = &tree.CreateSequence {
      Name: name,
      Persistence: $2.persistence(),
      Options: $5.seqOpts(),
    }
  }
| CREATE opt_persistence_sequence SEQUENCE IF NOT EXISTS sequence_name opt_create_seq_option_list
  {
    name := $7.unresolvedObjectName().ToTableName()
    $$.val = &tree.CreateSequence {
      Name: name,
      Persistence: $2.persistence(),
      IfNotExists: true,
      Options: $8.seqOpts(),
    }
  }

opt_create_seq_option_list:
  create_seq_option_list
| /* EMPTY */
  {
    $$.val = []tree.SequenceOption(nil)
  }

create_seq_option_list:
  create_seq_option_elem
  {
    $$.val = []tree.SequenceOption{$1.seqOpt()}
  }
| create_seq_option_list create_seq_option_elem
  {
    $$.val = append($1.seqOpts(), $2.seqOpt())
  }

create_seq_option_elem:
  seq_as_type
| seq_increment
| seq_minvalue
| seq_maxvalue
| seq_start
| seq_cache
| seq_cycle
| seq_owned_by

opt_alter_seq_option_list:
  alter_seq_option_list
| /* EMPTY */
  {
    $$.val = []tree.SequenceOption(nil)
  }

alter_seq_option_list:
  alter_seq_option_elem
  {
    $$.val = []tree.SequenceOption{$1.seqOpt()}
  }
| alter_seq_option_list alter_seq_option_elem
  {
    $$.val = append($1.seqOpts(), $2.seqOpt())
  }

alter_seq_option_elem:
  create_seq_option_elem
| seq_restart

seq_as_type:
  AS typename
  {
    $$.val = tree.SequenceOption{Name: tree.SeqOptAs, AsType: $2.typeReference()}
  }

seq_increment:
  INCREMENT signed_iconst64
  {
    x := $2.int64()
    $$.val = tree.SequenceOption{Name: tree.SeqOptIncrement, IntVal: &x}
  }
| INCREMENT BY signed_iconst64
  {
    x := $3.int64()
    $$.val = tree.SequenceOption{Name: tree.SeqOptIncrement, IntVal: &x}
  }

seq_minvalue:
  MINVALUE signed_iconst64
  {
    x := $2.int64()
    $$.val = tree.SequenceOption{Name: tree.SeqOptMinValue, IntVal: &x}
  }
| NO MINVALUE
  {
    $$.val = tree.SequenceOption{Name: tree.SeqOptMinValue}
  }

seq_maxvalue:
  MAXVALUE signed_iconst64
  {
    x := $2.int64()
    $$.val = tree.SequenceOption{Name: tree.SeqOptMaxValue, IntVal: &x}
  }
| NO MAXVALUE
  {
    $$.val = tree.SequenceOption{Name: tree.SeqOptMaxValue}
  }

seq_start:
  START opt_with signed_iconst64
  {
    x := $3.int64()
    $$.val = tree.SequenceOption{Name: tree.SeqOptStart, IntVal: &x}
  }

seq_cache:
  CACHE signed_iconst64
  {
    x := $2.int64()
    $$.val = tree.SequenceOption{Name: tree.SeqOptCache, IntVal: &x}
  }

seq_cycle:
  CYCLE
  {
    $$.val = tree.SequenceOption{Name: tree.SeqOptCycle}
  }
| NO CYCLE
  {
    $$.val = tree.SequenceOption{Name: tree.SeqOptNoCycle}
  }

seq_owned_by:
  OWNED BY column_path
  {
    varName, err := $3.unresolvedName().NormalizeVarName()
    if err != nil {
      return setErr(sqllex, err)
    }
    columnItem, ok := varName.(*tree.ColumnItem)
    if !ok {
      sqllex.Error(fmt.Sprintf("invalid column name: %q", tree.ErrString($3.unresolvedName())))
            return 1
    }
    $$.val = tree.SequenceOption{Name: tree.SeqOptOwnedBy, ColumnItemVal: columnItem}
  }
| OWNED BY NONE
  {
    $$.val = tree.SequenceOption{Name: tree.SeqOptOwnedBy, ColumnItemVal: nil}
  }

seq_restart:
  RESTART
  {
    $$.val = tree.SequenceOption{Name: tree.SeqOptRestart}
  }
| RESTART opt_with signed_iconst64
  {
    x := $3.int64()
    $$.val = tree.SequenceOption{Name: tree.SeqOptRestart, IntVal: &x}
  }

// %Help: TRUNCATE - empty one or more tables
// %Category: DML
// %Text: TRUNCATE [TABLE] <tablename> [, ...] [CASCADE | RESTRICT]
// %SeeAlso: WEBDOCS/truncate.html
truncate_stmt:
  TRUNCATE opt_table relation_expr_list opt_drop_behavior
  {
    $$.val = &tree.Truncate{Tables: $3.tableNames(), DropBehavior: $4.dropBehavior()}
  }
| TRUNCATE error // SHOW HELP: TRUNCATE

password_clause:
  PASSWORD non_reserved_word_or_sconst
  {
    $$.val = tree.KVOption{Key: tree.Name($1), Value: tree.NewDString($2)}
  }
| ENCRYPTED PASSWORD non_reserved_word_or_sconst
  {
    $$.val = tree.KVOption{Key: tree.Name($2), Value: tree.NewDString($3)}
  }
| PASSWORD NULL
  {
    $$.val = tree.KVOption{Key: tree.Name($1), Value: tree.DNull}
  }

create_trigger_stmt:
  CREATE opt_constraint TRIGGER trigger_name trigger_time trigger_events ON table_name opt_from_ref_table
  opt_trigger_deferrable_mode opt_trigger_relations opt_for_each opt_when EXECUTE function_or_procedure routine_name '(' opt_name_list ')'
  {
    $$.val = &tree.CreateTrigger{
      Replace: false,
      Constraint: $2.bool(),
      Name: tree.Name($4),
      Time: $5.triggerTime(),
      Events: $6.triggerEvents(),
      OnTable: $8.unresolvedObjectName().ToTableName(),
      RefTable: tree.Name($9),
      Deferrable: $10.triggerDeferrableMode(),
      Relations: $11.triggerRelations(),
      ForEachRow: $12.bool(),
      When: $13.expr(),
      FuncName: $16.unresolvedObjectName(),
      Args: $18.nameList(),
    }
  }
| CREATE OR REPLACE opt_constraint TRIGGER trigger_name trigger_time trigger_events ON table_name opt_from_ref_table
  opt_trigger_deferrable_mode opt_trigger_relations opt_for_each opt_when EXECUTE function_or_procedure routine_name '(' opt_name_list ')'
  {
    $$.val = &tree.CreateTrigger{
      Replace: true,
      Constraint: $4.bool(),
      Name: tree.Name($6),
      Time: $7.triggerTime(),
      Events: $8.triggerEvents(),
      OnTable: $10.unresolvedObjectName().ToTableName(),
      RefTable: tree.Name($11),
      Deferrable: $12.triggerDeferrableMode(),
      Relations: $13.triggerRelations(),
      ForEachRow: $14.bool(),
      When: $15.expr(),
      FuncName: $18.unresolvedObjectName(),
      Args: $20.nameList(),
    }
  }
  
function_or_procedure:
  FUNCTION
| PROCEDURE

opt_when:
  /* EMPTY */
  {
    $$.val = nil
  }
| WHEN '(' a_expr ')'
  {
    $$.val = $3.expr()
  }

opt_for_each:
  /* EMPTY */
  {
    $$.val = false
  }
| FOR opt_each ROW
  {
    $$.val = true
  }
| FOR opt_each STATEMENT
  {
    $$.val = false
  }

opt_each:
  /* EMPTY */
| EACH

opt_trigger_relations:
  /* EMPTY */
  {
    $$.val = tree.TriggerRelations(nil)
  }
| REFERENCING trigger_relations
  {
    $$.val = $2.triggerRelations()
  }

trigger_relations:
  old_or_new TABLE opt_as name
  {
    $$.val = tree.TriggerRelations{{IsOld: $1.bool(), Name: $4}}
  }
| trigger_relations old_or_new TABLE opt_as name
  {
    $$.val = append($1.triggerRelations(), tree.TriggerRelation{IsOld: $2.bool(), Name: $5})
  }

old_or_new:
  OLD
  {
    $$.val = true
  }
| NEW
  {
    $$.val = false
  }

opt_as:
  /* EMPTY */
| AS

opt_trigger_deferrable_mode:
  /* EMPTY */
  {
    $$.val = tree.TriggerNotDeferrable
  }
| DEFERRABLE
  {
    $$.val = tree.TriggerDeferrable
  }
| INITIALLY IMMEDIATE
  {
    $$.val = tree.TriggerDeferrable
  }
| INITIALLY DEFERRED
  {
    $$.val = tree.TriggerInitiallyDeferred
  }
| NOT_LA DEFERRABLE
  {
    $$.val = tree.TriggerNotDeferrable
  }
| DEFERRABLE INITIALLY IMMEDIATE
  {
    $$.val = tree.TriggerDeferrable
  }
| DEFERRABLE INITIALLY DEFERRED
  {
    $$.val = tree.TriggerInitiallyDeferred
  }

opt_from_ref_table:
  /* EMPTY */
  {
    $$ = ""
  }
| FROM name
  {
    $$ = $2
  }

trigger_events:
  trigger_event
  {
    $$.val = tree.TriggerEvents{$1.triggerEvent()}
  }
| trigger_events OR trigger_event
  {
    $$.val = append($1.triggerEvents(), $3.triggerEvent())
  }

trigger_event:
  INSERT
  {
    $$.val = tree.TriggerEvent{Type: tree.TriggerEventInsert}
  }
| UPDATE opt_of_cols
  {
    $$.val = tree.TriggerEvent{Type: tree.TriggerEventUpdate, Cols: $2.nameList()}
  }
| DELETE
  {
    $$.val = tree.TriggerEvent{Type: tree.TriggerEventDelete}
  }
| TRUNCATE
  {
    $$.val = tree.TriggerEvent{Type: tree.TriggerEventTruncate}
  }

opt_of_cols:
  /* EMPTY */
  {
    $$.val = tree.NameList(nil)
  }
| OF name_list
  {
    $$.val = $2.nameList()
  }

trigger_time:
  BEFORE
  {
    $$.val = tree.TriggerTimeBefore
  }
| AFTER
  {
    $$.val = tree.TriggerTimeAfter
  }
| INSTEAD OF
  {
    $$.val = tree.TriggerTimeInsteadOf
  }

opt_constraint:
  /* EMPTY */
  {
    $$.val = false
  }
| CONSTRAINT
  {
    $$.val = true
  }

// %Help: CREATE ROLE - define a new role
// %Category: Priv
// %Text: CREATE ROLE [IF NOT EXISTS] <name> [ [WITH] <OPTIONS...> ]
// %SeeAlso: ALTER ROLE, DROP ROLE, SHOW ROLES
create_role_stmt:
  CREATE role_or_group_or_user non_reserved_word_or_sconst opt_role_options
  {
    $$.val = &tree.CreateRole{Name: $3, KVOptions: $4.kvOptions(), IsRole: $2.bool()}
  }
| CREATE role_or_group_or_user IF NOT EXISTS non_reserved_word_or_sconst opt_role_options
  {
    $$.val = &tree.CreateRole{Name: $6, IfNotExists: true, KVOptions: $7.kvOptions(), IsRole: $2.bool()}
  }
| CREATE role_or_group_or_user error // SHOW HELP: CREATE ROLE

// %Help: ALTER ROLE - alter a role
// %Category: Priv
// %Text: ALTER ROLE <name> [WITH] <options...>
// %SeeAlso: CREATE ROLE, DROP ROLE, SHOW ROLES
alter_role_stmt:
  ALTER role_or_group_or_user non_reserved_word_or_sconst opt_role_options
{
  $$.val = &tree.AlterRole{Name: $3, KVOptions: $4.kvOptions(), IsRole: $2.bool()}
}
| ALTER role_or_group_or_user IF EXISTS non_reserved_word_or_sconst opt_role_options
{
  $$.val = &tree.AlterRole{Name: $5, IfExists: true, KVOptions: $6.kvOptions(), IsRole: $2.bool()}
}
| ALTER role_or_group_or_user error // SHOW HELP: ALTER ROLE

// "CREATE GROUP is now an alias for CREATE ROLE"
// https://www.postgresql.org/docs/10/static/sql-creategroup.html
role_or_group_or_user:
  role_or_user
  {
    $$ = $1
  }
| GROUP
  {
    /* SKIP DOC */
    $$.val = true
  }

role_or_user:
  ROLE
  {
    $$.val = true
  }
| USER
  {
    $$.val = false
  }

// %Help: CREATE VIEW - create a new view
// %Category: DDL
// %Text: CREATE [ OR REPLACE ] [ TEMP | TEMPORARY ] [ RECURSIVE ] VIEW name [ ( column_name [, ...] ) ]
//    	  [ WITH ( view_option_name [= view_option_value] [, ... ] ) ]
//    	  AS <source>
//    	  [ WITH [ CASCADED | LOCAL ] CHECK OPTION ]
// %SeeAlso: CREATE TABLE, SHOW CREATE, WEBDOCS/create-view.html
create_view_stmt:
  CREATE opt_temp opt_view_recursive VIEW view_name opt_column_list opt_with_view_options AS select_stmt opt_with_check_option
  {
    name := $5.unresolvedObjectName().ToTableName()
    $$.val = &tree.CreateView{
      Name: name,
      Replace: false,
      Persistence: $2.persistence(),
      IsRecursive: $3.bool(),
      ColumnNames: $6.nameList(),
      Options: $7.viewOptions(),
      AsSource: $9.slct(),
      CheckOption: $10.viewCheckOption(),
    }
  }
// We cannot use a rule like opt_or_replace here as that would cause a conflict
// with the opt_temp rule.
| CREATE OR REPLACE opt_temp opt_view_recursive VIEW view_name opt_column_list opt_with_view_options AS select_stmt opt_with_check_option
  {
    name := $7.unresolvedObjectName().ToTableName()
    $$.val = &tree.CreateView{
      Name: name,
      Replace: true,
      Persistence: $4.persistence(),
      IsRecursive: $5.bool(),
      ColumnNames: $8.nameList(),
      Options: $9.viewOptions(),
      AsSource: $11.slct(),
      CheckOption: $12.viewCheckOption(),
    }
  }

opt_with_check_option:
  /* EMPTY */
  {
    $$.val = tree.ViewCheckOptionUnspecified
  }
| WITH CHECK OPTION
  {
    $$.val = tree.ViewCheckOptionCascaded
  }
| WITH CASCADED CHECK OPTION
  {
    $$.val = tree.ViewCheckOptionCascaded
  }
| WITH LOCAL CHECK OPTION
  {
    $$.val = tree.ViewCheckOptionLocal
  }

opt_with_view_options:
  /* EMPTY */
  {
    $$.val = tree.ViewOptions(nil)
  }
| WITH '(' view_options ')'
  {
    $$.val = $3.viewOptions()
  }

view_options:
  view_option
  {
    $$.val = tree.ViewOptions{$1.viewOption()}
  }
| view_options ',' view_option
  {
    $$.val = append($1.viewOptions(), $3.viewOption())
  }

view_option:
  CHECK_OPTION { $$.val = tree.ViewOption{Name: $1, CheckOpt: "cascaded"} }
| CHECK_OPTION '=' SCONST { $$.val = tree.ViewOption{Name: $1, CheckOpt: $3} }
| SECURITY_BARRIER { $$.val = tree.ViewOption{Name: $1, Security: false} }
| SECURITY_BARRIER '=' TRUE { $$.val = tree.ViewOption{Name: $1, Security: true} }
| SECURITY_BARRIER '=' FALSE { $$.val = tree.ViewOption{Name: $1, Security: false} }
| SECURITY_INVOKER { $$.val = tree.ViewOption{Name: $1, Security: false} }
| SECURITY_INVOKER '=' TRUE { $$.val = tree.ViewOption{Name: $1, Security: true} }
| SECURITY_INVOKER '=' FALSE { $$.val = tree.ViewOption{Name: $1, Security: false} }

create_materialized_view_stmt:
  CREATE MATERIALIZED VIEW view_name opt_column_list opt_using_method opt_with_storage_parameter_list opt_tablespace AS select_stmt opt_create_as_with_data
  {
    name := $4.unresolvedObjectName().ToTableName()
    $$.val = &tree.CreateMaterializedView{
      Name: name,
      ColumnNames: $5.nameList(),
      AsSource: $10.slct(),
      Using: $6,
      Params: $7.storageParams(),
      Tablespace: tree.Name($8),
      WithNoData: $11.bool(),
    }
  }
| CREATE MATERIALIZED VIEW IF NOT EXISTS view_name opt_column_list opt_using_method opt_with_storage_parameter_list opt_tablespace AS select_stmt opt_create_as_with_data
  {
    name := $7.unresolvedObjectName().ToTableName()
    $$.val = &tree.CreateMaterializedView{
      Name: name,
      ColumnNames: $8.nameList(),
      AsSource: $13.slct(),
      IfNotExists: true,
      Using: $9,
      Params: $10.storageParams(),
      Tablespace: tree.Name($11),
      WithNoData: $14.bool(),
    }
  }

role_option:
  SUPERUSER
  {
    $$.val = tree.KVOption{Key: tree.Name($1), Value: nil}
  }
| NOSUPERUSER
  {
    $$.val = tree.KVOption{Key: tree.Name($1), Value: nil}
  }
| CREATEDB
  {
    $$.val = tree.KVOption{Key: tree.Name($1), Value: nil}
  }
| NOCREATEDB
  {
    $$.val = tree.KVOption{Key: tree.Name($1), Value: nil}
  }
| CREATEROLE
  {
    $$.val = tree.KVOption{Key: tree.Name($1), Value: nil}
  }
| NOCREATEROLE
  {
    $$.val = tree.KVOption{Key: tree.Name($1), Value: nil}
  }
| INHERIT
  {
    $$.val = tree.KVOption{Key: tree.Name($1), Value: nil}
  }
| NOINHERIT
  {
    $$.val = tree.KVOption{Key: tree.Name($1), Value: nil}
  }
| LOGIN
  {
    $$.val = tree.KVOption{Key: tree.Name($1), Value: nil}
  }
| NOLOGIN
  {
    $$.val = tree.KVOption{Key: tree.Name($1), Value: nil}
  }
| REPLICATION
  {
    $$.val = tree.KVOption{Key: tree.Name($1), Value: nil}
  }
| NOREPLICATION
  {
    $$.val = tree.KVOption{Key: tree.Name($1), Value: nil}
  }
| BYPASSRLS
  {
    $$.val = tree.KVOption{Key: tree.Name($1), Value: nil}
  }
| NOBYPASSRLS
  {
    $$.val = tree.KVOption{Key: tree.Name($1), Value: nil}
  }
| CONNECTION LIMIT signed_iconst32
  {
    $$.val = tree.KVOption{Key: tree.Name(fmt.Sprintf("%s_%s", $1, $2)), Value: tree.NewDInt(tree.DInt($3.val.(int32)))}
  }
| SYSID ICONST
  {
    $$.val = tree.KVOption{Key: tree.Name($1), Value: nil}
  }
| password_clause
| valid_until_clause


role_options:
  role_option
  {
    $$.val = []tree.KVOption{$1.kvOption()}
  }
|  role_options role_option
  {
    $$.val = append($1.kvOptions(), $2.kvOption())
  }

opt_role_options:
  opt_with role_options
  {
    $$.val = $2.kvOptions()
  }
| /* EMPTY */
  {
    $$.val = nil
  }

valid_until_clause:
  VALID UNTIL non_reserved_word_or_sconst
  {
    $$.val = tree.KVOption{Key: tree.Name(fmt.Sprintf("%s_%s", $1, $2)), Value: tree.NewDString($2)}
  }
| VALID UNTIL NULL
  {
    $$.val = tree.KVOption{Key: tree.Name(fmt.Sprintf("%s_%s", $1, $2)), Value: tree.DNull}
  }

opt_view_recursive:
  /* EMPTY */
  {
    $$.val = false
  }
| RECURSIVE
  {
    $$.val = true
  }

// %Help: CREATE TYPE -- create a type
// %Category: DDL
// %Text: CREATE TYPE <type_name> AS ENUM (...)
create_type_stmt:
  // Record/Composite types.
  CREATE TYPE type_name AS '(' opt_type_composite_list ')'
  {
    $$.val = &tree.CreateType{
      TypeName: $3.unresolvedObjectName(),
      Variety: tree.Composite,
      Composite: tree.CompositeType{Types: $6.compositeTypeElems()},
    }
  }
  // Enum types.
| CREATE TYPE type_name AS ENUM '(' opt_enum_val_list ')'
  {
    $$.val = &tree.CreateType{
      TypeName: $3.unresolvedObjectName(),
      Variety: tree.Enum,
      Enum: tree.EnumType{Labels: $7.strs()},
    }
  }
  // Range types.
| CREATE TYPE type_name AS RANGE '(' SUBTYPE '=' typename type_range_optional_list ')'
  {
    $$.val = &tree.CreateType{
      TypeName: $3.unresolvedObjectName(),
      Variety: tree.Range,
      Range: tree.RangeType{
        Subtype: $9.typeReference(),
        Options: $10.rangeTypeOptions(),
      },
    }
  }
  // Base (primitive) types.
| CREATE TYPE type_name '(' INPUT '=' name ',' OUTPUT '=' name type_base_optional_list ')'
  {
    $$.val = &tree.CreateType{
      TypeName: $3.unresolvedObjectName(),
      Variety: tree.Base,
      Base: tree.BaseType{
        Input: $7,
        Output: $11,
	Options: $12.baseTypeOptions(),
      },
    }
  }
  // Shell types, gateway to define base types using the previous syntax.
| CREATE TYPE type_name
  {
    $$.val = &tree.CreateType{
      TypeName: $3.unresolvedObjectName(),
      Variety: tree.Shell,
    }
  }

opt_type_composite_list:
  /* EMPTY */
  {
    $$.val = []tree.CompositeTypeElem{}
  }
| type_composite_list
  {
    $$.val = $1.compositeTypeElems()
  }

type_composite_list:
  name typename opt_collate
  {
    $$.val = []tree.CompositeTypeElem{{AttrName: $1, Type: $2.typeReference(), Collate: $3.unresolvedObjectName().UnquotedString()}}
  }
| type_composite_list ',' name typename opt_collate
  {
    $$.val = append($1.compositeTypeElems(), tree.CompositeTypeElem{AttrName: $3, Type: $4.typeReference(), Collate: $5.unresolvedObjectName().UnquotedString()})
  }

type_range_optional_list:
  /* EMPTY */
  {
    $$.val = []tree.RangeTypeOption(nil)
  }
| type_range_option
  {
    $$.val = []tree.RangeTypeOption{$1.rangeTypeOption()}
  }
| type_range_optional_list ',' type_range_option
  {
    $$.val = append($1.rangeTypeOptions(), $3.rangeTypeOption())
  }

type_range_option:
  SUBTYPE_OPCLASS '=' name
  { $$.val = tree.RangeTypeOption{Option: tree.RangeTypeSubtypeOpClass, StrVal: $3} }
| COLLATION '=' collation_name
  { $$.val = tree.RangeTypeOption{Option: tree.RangeTypeCollation, StrVal: $3.unresolvedObjectName().UnquotedString()} }
| CANONICAL '=' name
  { $$.val = tree.RangeTypeOption{Option: tree.RangeTypeCanonical, StrVal: $3} }
| SUBTYPE_DIFF '=' name
  { $$.val = tree.RangeTypeOption{Option: tree.RangeTypeSubtypeDiff, StrVal: $3} }
| MULTIRANGE_TYPE_NAME '=' type_name
  { $$.val = tree.RangeTypeOption{Option: tree.RangeTypeMultiRangeTypeName, MRTypeName: $3.unresolvedObjectName()} }

type_base_optional_list:
  /* EMPTY */
  {
    $$.val = []tree.BaseTypeOption(nil)
  }
| type_base_option
  {
    $$.val = []tree.BaseTypeOption{$1.baseTypeOption()}
  }
| type_base_optional_list ',' type_base_option
  {
    $$.val = append($1.baseTypeOptions(), $3.baseTypeOption())
  }

type_base_option:
  type_property
| INTERNALLENGTH '=' signed_iconst64
  {
    $$.val = tree.BaseTypeOption{Option: tree.BaseTypeInternalLength, InternalLength: $3.int64()}
  }
| INTERNALLENGTH '=' VARIABLE
  {
    $$.val = tree.BaseTypeOption{
      Option: tree.BaseTypeInternalLength,
      InternalLength: -1,
    }
  }
| PASSEDBYVALUE
  { $$.val = tree.BaseTypeOption{Option: tree.BaseTypePassedByValue} }
| ALIGNMENT '=' name
  { $$.val = tree.BaseTypeOption{Option: tree.BaseTypeAlignment, StrVal: $3} }
| LIKE '=' typename
  { $$.val = tree.BaseTypeOption{Option: tree.BaseTypeLikeType, TypeVal: $3.typeReference()} }
| CATEGORY '=' name
  { $$.val = tree.BaseTypeOption{Option: tree.BaseTypeCategory, StrVal: $3} }
| PREFERRED '=' TRUE
  { $$.val = tree.BaseTypeOption{Option: tree.BaseTypePreferred, BoolVal: true} }
| PREFERRED '=' FALSE
  { $$.val = tree.BaseTypeOption{Option: tree.BaseTypePreferred, BoolVal: false} }
| DEFAULT '=' a_expr
  { $$.val = tree.BaseTypeOption{Option: tree.BaseTypeDefault, Default: $3.expr()} }
| ELEMENT '=' typename
  { $$.val = tree.BaseTypeOption{Option: tree.BaseTypeElement, TypeVal: $3.typeReference()} }
| DELIMITER '=' non_reserved_word_or_sconst
  { $$.val = tree.BaseTypeOption{Option: tree.BaseTypeDelimiter, StrVal: $3} }
| COLLATABLE '=' TRUE
  { $$.val = tree.BaseTypeOption{Option: tree.BaseTypeCollatable, BoolVal: true} }
| COLLATABLE '=' FALSE
  { $$.val = tree.BaseTypeOption{Option: tree.BaseTypeCollatable, BoolVal: false} }

type_property:
  RECEIVE '=' name
  { $$.val = tree.BaseTypeOption{Option: tree.BaseTypeReceive, StrVal: $3} }
| SEND '=' name
  { $$.val = tree.BaseTypeOption{Option: tree.BaseTypeSend, StrVal: $3} }
| TYPMOD_IN '=' name
  { $$.val = tree.BaseTypeOption{Option: tree.BaseTypeTypModIn, StrVal: $3} }
| TYPMOD_OUT '=' name
  { $$.val = tree.BaseTypeOption{Option: tree.BaseTypeTypeModOut, StrVal: $3} }
| ANALYZE '=' name
  { $$.val = tree.BaseTypeOption{Option: tree.BaseTypeAnalyze, StrVal: $3} }
| SUBSCRIPT '=' name
  { $$.val = tree.BaseTypeOption{Option: tree.BaseTypeSubscript, StrVal: $3} }
| STORAGE '=' name
  { $$.val = tree.BaseTypeOption{Option: tree.BaseTypeStorage, StrVal: $3} }

opt_enum_val_list:
  enum_val_list
  {
    $$.val = $1.strs()
  }
| /* EMPTY */
  {
    $$.val = []string(nil)
  }

enum_val_list:
  SCONST
  {
    $$.val = []string{$1}
  }
| enum_val_list ',' SCONST
  {
    $$.val = append($1.strs(), $3)
  }

// %Help: CREATE INDEX - create a new index
// %Category: DDL
// %Text:
// CREATE [UNIQUE] INDEX [CONCURRENTLY] [IF NOT EXISTS] [<idxname>]
//        ON [ONLY] <tablename> [USING <method>]
//        ( { column_name | ( expression ) } [ COLLATE collation ] [ opclass [ ( opclass_parameter = value [, ... ] ) ] ] [ ASC | DESC ] [ NULLS { FIRST | LAST } ] [, ...] )
//        [INCLUDE ( column_name, [,...] )]
//        [NULLS [ NOT ] DISTINCT]
//        [WITH <storage_parameter_list>]
//        [TABLESPACE tablespace_name]
//        [WHERE <where_conds...>]
create_index_stmt:
  CREATE opt_unique INDEX opt_concurrently opt_index_name ON opt_only table_name opt_using_method '(' index_params ')' opt_include_index_cols opt_nulls_distinct opt_with_storage_parameter_list opt_tablespace opt_where_clause
  {
    table := $8.unresolvedObjectName().ToTableName()
    $$.val = &tree.CreateIndex{
      Name:          tree.Name($5),
      Table:         table,
      Unique:        $2.bool(),
      Concurrently:  $4.bool(),
      Only:          $7.bool(),
      Using:         $9,
      Columns:       $11.idxElems(),
      IndexParams:   tree.IndexParams{IncludeColumns: $13.idxElems(), StorageParams: $15.storageParams(), Tablespace: tree.Name($16)},
      NullsDistinct: $14.bool(),
      Predicate:     $17.expr(),
    }
  }
| CREATE opt_unique INDEX opt_concurrently IF NOT EXISTS index_name ON opt_only table_name opt_using_method '(' index_params ')' opt_include_index_cols opt_nulls_distinct opt_with_storage_parameter_list opt_tablespace opt_where_clause
  {
    table := $11.unresolvedObjectName().ToTableName()
    $$.val = &tree.CreateIndex{
      Name:          tree.Name($8),
      Table:         table,
      Unique:        $2.bool(),
      Concurrently:  $4.bool(),
      IfNotExists:   true,
      Only:          $10.bool(),
      Using:         $12,
      Columns:       $14.idxElems(),
      IndexParams:   tree.IndexParams{IncludeColumns: $16.idxElems(), StorageParams: $18.storageParams(), Tablespace: tree.Name($19)},
      NullsDistinct: $17.bool(),
      Predicate:     $20.expr(),
    }
  }
| CREATE opt_unique INDEX error // SHOW HELP: CREATE INDEX

opt_using_method:
  USING name
  {
    $$ = $2
  }
| /* EMPTY */
  {
    $$ = ""
  }

opt_concurrently:
  CONCURRENTLY
  {
    $$.val = true
  }
| /* EMPTY */
  {
    $$.val = false
  }

opt_unique:
  UNIQUE
  {
    $$.val = true
  }
| /* EMPTY */
  {
    $$.val = false
  }

index_params:
  index_elem
  {
    $$.val = tree.IndexElemList{$1.idxElem()}
  }
| index_params ',' index_elem
  {
    $$.val = append($1.idxElems(), $3.idxElem())
  }

// Index attributes can be either simple column references, or arbitrary
// expressions in parens. For backwards-compatibility reasons, we allow an
// expression that is just a function call to be written without parens.
index_elem:
  name opt_collate opt_opclass opt_asc_desc opt_nulls_order
  {
    $$.val = tree.IndexElem{Column: tree.Name($1), Collation: $2.unresolvedObjectName().UnquotedString(), OpClass: $3.opClass(), Direction: $4.dir(), NullsOrder: $5.nullsOrder()}
  }
| '(' a_expr ')' opt_collate opt_opclass opt_asc_desc opt_nulls_order
  {
    $$.val = tree.IndexElem{Expr: $2.expr(), Collation: $4.unresolvedObjectName().UnquotedString(), OpClass: $5.opClass(), Direction: $6.dir(), NullsOrder: $7.nullsOrder()}
  }

opt_opclass:
  /* EMPTY */
  {
    $$.val = (*tree.IndexElemOpClass)(nil)
  }
| IDENT
  {
    $$.val = &tree.IndexElemOpClass{Name: $1}
  }
| IDENT '(' opclass_option_list ')'
  {
    $$.val = &tree.IndexElemOpClass{Name: $1, Options: $3.opClassOptions()}
  }

opclass_option_list:
  name '=' a_expr
  {
    $$.val = []tree.IndexElemOpClassOption{{Param: $1, Val: $3.expr()}}
  }
| opclass_option_list ',' name '=' a_expr
  {
    $$.val = append($1.opClassOptions(), tree.IndexElemOpClassOption{Param: $3, Val: $5.expr()})
  }

opt_collate:
  COLLATE collation_name { $$ = $2 }
| /* EMPTY */ 
  {
    // TODO: this instantiates a zero-part object name, which then has the empty string for its String() output. It
    // would probably be better to return a nil object in this case, but that requires deeper changes.  
    $$.val = tree.NewUnresolvedName().GetUnresolvedObjectName() 
  }

opt_asc_desc:
  ASC
  {
    $$.val = tree.Ascending
  }
| DESC
  {
    $$.val = tree.Descending
  }
| /* EMPTY */
  {
    $$.val = tree.DefaultDirection
  }

alter_database_to_schema_stmt:
  ALTER DATABASE database_name CONVERT TO SCHEMA WITH PARENT database_name
  {
    $$.val = &tree.ReparentDatabase{Name: tree.Name($3), Parent: tree.Name($9)}
  }

alter_rename_database_stmt:
  ALTER DATABASE database_name RENAME TO database_name
  {
    $$.val = &tree.RenameDatabase{Name: tree.Name($3), NewName: tree.Name($6)}
  }

alter_rename_table_stmt:
  ALTER TABLE relation_expr RENAME TO table_name
  {
    name := $3.unresolvedObjectName()
    newName := $6.unresolvedObjectName()
    $$.val = &tree.RenameTable{Name: name, NewName: newName, IfExists: false}
  }
| ALTER TABLE IF EXISTS relation_expr RENAME TO table_name
  {
    name := $5.unresolvedObjectName()
    newName := $8.unresolvedObjectName()
    $$.val = &tree.RenameTable{Name: name, NewName: newName, IfExists: true}
  }

alter_table_set_schema_stmt:
  ALTER TABLE relation_expr set_schema
  {
    $$.val = &tree.AlterTableSetSchema{
      Name: $3.unresolvedObjectName(), Schema: $4, IfExists: false,
    }
  }
| ALTER TABLE IF EXISTS relation_expr set_schema
  {
    $$.val = &tree.AlterTableSetSchema{
      Name: $5.unresolvedObjectName(), Schema: $6, IfExists: true,
    }
  }

alter_table_all_in_tablespace_stmt:
  ALTER TABLE ALL IN TABLESPACE tablespace_name opt_owned_by_list set_tablespace opt_nowait
  {
    $$.val = &tree.AlterTableAllInTablespace{
      Name: tree.Name($6), OwnedBy: $7.strs(), Tablespace: $8, NoWait: $9.bool(),
    }
  }

alter_table_parition_stmt:
  ALTER TABLE table_name ATTACH PARTITION partition_name partition_of
  {
    $$.val = &tree.AlterTablePartition{
      Name: $3.unresolvedObjectName(), IfExists: false, Partition: tree.Name($6), Spec: $7.partitionBoundSpec(),
    }
  }
| ALTER TABLE IF EXISTS table_name ATTACH PARTITION partition_name partition_of
  {
    $$.val = &tree.AlterTablePartition{
      Name: $5.unresolvedObjectName(), IfExists: true, Partition: tree.Name($8), Spec: $9.partitionBoundSpec(),
    }
  }
| ALTER TABLE table_name DETACH PARTITION partition_name detach_partition_type
  {
    $$.val = &tree.AlterTablePartition{
      Name: $3.unresolvedObjectName(), IfExists: false, Partition: tree.Name($6), IsDetach: true, DetachType: $7.detachPartition(),
    }
  }
| ALTER TABLE IF EXISTS table_name DETACH PARTITION partition_name detach_partition_type
  {
    $$.val = &tree.AlterTablePartition{
      Name: $5.unresolvedObjectName(), IfExists: true, Partition: tree.Name($8), IsDetach: true, DetachType: $9.detachPartition(),
    }
  }

opt_nowait:
  /* EMPTY */
  {
    $$.val = false
  }
| NOWAIT
  {
    $$.val = true
  }

detach_partition_type:
  /* EMPTY */
  {
    $$.val = tree.DetachPartitionNone
  }
| CONCURRENTLY
  {
    $$.val = tree.DetachPartitionConcurrently
  }
| FINALIZE
  {
    $$.val = tree.DetachPartitionFinalize
  }

alter_sequence_set_schema_stmt:
  ALTER SEQUENCE relation_expr set_schema
  {
    $$.val = &tree.AlterTableSetSchema{
      Name: $3.unresolvedObjectName(), Schema: $4, IfExists: false, IsSequence: true,
    }
  }
| ALTER SEQUENCE IF EXISTS relation_expr set_schema
  {
    $$.val = &tree.AlterTableSetSchema{
      Name: $5.unresolvedObjectName(), Schema: $6, IfExists: true, IsSequence: true,
    }
  }

alter_materialized_view_stmt:
  ALTER MATERIALIZED VIEW relation_expr alter_materialized_view_cmd
  {
    $$.val = &tree.AlterMaterializedView{Name: $4.unresolvedObjectName(), IfExists: false, Cmds: $5.alterTableCmds()}
  }
| ALTER MATERIALIZED VIEW IF EXISTS relation_expr alter_materialized_view_cmd
  {
    $$.val = &tree.AlterMaterializedView{Name: $6.unresolvedObjectName(), IfExists: true, Cmds: $7.alterTableCmds()}
  }
| ALTER MATERIALIZED VIEW table_name opt_no DEPENDS ON EXTENSION name
  {
    $$.val = &tree.AlterMaterializedView{Name: $4.unresolvedObjectName(), No: $5.bool(), Extension: $9}
  }
| alter_materialized_view_rename_stmt
| alter_materialized_view_set_schema_stmt
| alter_materialized_view_all_in_tablespace_stmt

alter_materialized_view_cmd:
  alter_materialized_view_actions
  {
    $$.val = $1.alterTableCmds()
  }
| RENAME opt_column column_name TO column_name
  {
    $$.val = tree.AlterTableCmds{&tree.AlterTableRenameColumn{Column: tree.Name($3), NewName: tree.Name($5)}}
  }

alter_materialized_view_rename_stmt:
  ALTER MATERIALIZED VIEW relation_expr RENAME TO view_name
  {
    name := $4.unresolvedObjectName()
    newName := $7.unresolvedObjectName()
    $$.val = &tree.RenameTable{Name: name, NewName: newName, IfExists: false, IsMaterialized: true}
  }
| ALTER MATERIALIZED VIEW IF EXISTS relation_expr RENAME TO view_name
  {
    name := $6.unresolvedObjectName()
    newName := $9.unresolvedObjectName()
    $$.val = &tree.RenameTable{Name: name, NewName: newName, IfExists: false, IsMaterialized: true}
  }

alter_materialized_view_set_schema_stmt:
  ALTER MATERIALIZED VIEW relation_expr set_schema
  {
    $$.val = &tree.AlterTableSetSchema{
      Name: $4.unresolvedObjectName(), Schema: $5, IfExists: false, IsMaterialized: true,
    }
  }
| ALTER MATERIALIZED VIEW IF EXISTS relation_expr set_schema
  {
    $$.val = &tree.AlterTableSetSchema{
      Name: $6.unresolvedObjectName(), Schema: $7, IfExists: true, IsMaterialized: true,
    }
  }

alter_materialized_view_all_in_tablespace_stmt:
  ALTER MATERIALIZED VIEW ALL IN TABLESPACE tablespace_name opt_owned_by_list set_tablespace opt_nowait
  {
    $$.val = &tree.AlterTableAllInTablespace{
      Name: tree.Name($7), OwnedBy: $8.strs(), Tablespace: $9, NoWait: $10.bool(), IsMaterialized: true,
    }
  }

alter_rename_sequence_stmt:
  ALTER SEQUENCE relation_expr RENAME TO sequence_name
  {
    name := $3.unresolvedObjectName()
    newName := $6.unresolvedObjectName()
    $$.val = &tree.RenameTable{Name: name, NewName: newName, IfExists: false, IsSequence: true}
  }
| ALTER SEQUENCE IF EXISTS relation_expr RENAME TO sequence_name
  {
    name := $5.unresolvedObjectName()
    newName := $8.unresolvedObjectName()
    $$.val = &tree.RenameTable{Name: name, NewName: newName, IfExists: true, IsSequence: true}
  }

alter_rename_index_stmt:
  ALTER INDEX table_index_name RENAME TO index_name
  {
    $$.val = &tree.RenameIndex{Index: $3.newTableIndexName(), NewName: tree.UnrestrictedName($6), IfExists: false}
  }
| ALTER INDEX IF EXISTS table_index_name RENAME TO index_name
  {
    $$.val = &tree.RenameIndex{Index: $5.newTableIndexName(), NewName: tree.UnrestrictedName($8), IfExists: true}
  }

opt_column:
  COLUMN {}
| /* EMPTY */ {}

opt_set_data:
  SET DATA {}
| /* EMPTY */ {}

// %Help: RELEASE - complete a sub-transaction
// %Category: Txn
// %Text: RELEASE [SAVEPOINT] <savepoint name>
// %SeeAlso: SAVEPOINT, WEBDOCS/savepoint.html
release_stmt:
  RELEASE savepoint_name
  {
    $$.val = &tree.ReleaseSavepoint{Savepoint: tree.Name($2)}
  }
| RELEASE error // SHOW HELP: RELEASE

// %Help: RESUME JOBS - resume background jobs
// %Category: Misc
// %Text:
// RESUME JOBS <selectclause>
// RESUME JOB <jobid>
// %SeeAlso: SHOW JOBS, CANCEL JOBS, PAUSE JOBS
resume_jobs_stmt:
  RESUME JOB a_expr
  {
    $$.val = &tree.ControlJobs{
      Jobs: &tree.Select{
        Select: &tree.ValuesClause{Rows: []tree.Exprs{tree.Exprs{$3.expr()}}},
      },
      Command: tree.ResumeJob,
    }
  }
| RESUME JOB error // SHOW HELP: RESUME JOBS
| RESUME JOBS select_stmt
  {
    $$.val = &tree.ControlJobs{Jobs: $3.slct(), Command: tree.ResumeJob}
  }
| RESUME JOBS for_schedules_clause
  {
    $$.val = &tree.ControlJobsForSchedules{Schedules: $3.slct(), Command: tree.ResumeJob}
  }
| RESUME JOBS error // SHOW HELP: RESUME JOBS

// %Help: RESUME SCHEDULES - resume executing scheduled jobs
// %Category: Misc
// %Text:
// RESUME SCHEDULES <selectclause>
//  selectclause: select statement returning schedule IDs to resume.
//
// RESUME SCHEDULES <jobid>
//
// %SeeAlso: PAUSE SCHEDULES, SHOW JOBS, RESUME JOBS
resume_schedules_stmt:
  RESUME SCHEDULE a_expr
  {
    $$.val = &tree.ControlSchedules{
      Schedules: &tree.Select{
        Select: &tree.ValuesClause{Rows: []tree.Exprs{tree.Exprs{$3.expr()}}},
      },
      Command: tree.ResumeSchedule,
    }
  }
| RESUME SCHEDULE error // SHOW HELP: RESUME SCHEDULES
| RESUME SCHEDULES select_stmt
  {
    $$.val = &tree.ControlSchedules{
      Schedules: $3.slct(),
      Command: tree.ResumeSchedule,
    }
  }
| RESUME SCHEDULES error // SHOW HELP: RESUME SCHEDULES

// %Help: DROP SCHEDULES - destroy specified schedules
// %Category: Misc
// %Text:
// DROP SCHEDULES <selectclause>
//  selectclause: select statement returning schedule IDs to resume.
//
// DROP SCHEDULES <jobid>
//
// %SeeAlso: PAUSE SCHEDULES, SHOW JOBS, CANCEL JOBS
drop_schedule_stmt:
  DROP SCHEDULE a_expr
  {
    $$.val = &tree.ControlSchedules{
      Schedules: &tree.Select{
        Select: &tree.ValuesClause{Rows: []tree.Exprs{tree.Exprs{$3.expr()}}},
      },
      Command: tree.DropSchedule,
    }
  }
| DROP SCHEDULE error // SHOW HELP: DROP SCHEDULES
| DROP SCHEDULES select_stmt
  {
    $$.val = &tree.ControlSchedules{
      Schedules: $3.slct(),
      Command: tree.DropSchedule,
    }
  }
| DROP SCHEDULES error // SHOW HELP: DROP SCHEDULES

// %Help: SAVEPOINT - start a sub-transaction
// %Category: Txn
// %Text: SAVEPOINT <savepoint name>
// %SeeAlso: RELEASE, WEBDOCS/savepoint.html
savepoint_stmt:
  SAVEPOINT name
  {
    $$.val = &tree.Savepoint{Name: tree.Name($2)}
  }
| SAVEPOINT error // SHOW HELP: SAVEPOINT

// BEGIN / START / COMMIT / END / ROLLBACK / ...
transaction_stmt:
  begin_stmt    // EXTEND WITH HELP: BEGIN
| commit_stmt   // EXTEND WITH HELP: COMMIT
| rollback_stmt // EXTEND WITH HELP: ROLLBACK
| abort_stmt    /* SKIP DOC */

// %Help: BEGIN - start a transaction
// %Category: Txn
// %Text:
// BEGIN [TRANSACTION] [ <txnparameter> [[,] ...] ]
// START TRANSACTION [ <txnparameter> [[,] ...] ]
//
// Transaction parameters:
//    ISOLATION LEVEL { SNAPSHOT | SERIALIZABLE }
//    PRIORITY { LOW | NORMAL | HIGH }
//
// %SeeAlso: COMMIT, ROLLBACK, WEBDOCS/begin-transaction.html
begin_stmt:
  BEGIN opt_work_transaction begin_transaction
  {
    $$.val = $3.stmt()
  }
| BEGIN error // SHOW HELP: BEGIN
| START TRANSACTION begin_transaction
  {
    $$.val = $3.stmt()
  }
| START error // SHOW HELP: BEGIN

// %Help: COMMIT - commit the current transaction
// %Category: Txn
// %Text:
// COMMIT [TRANSACTION]
// END [TRANSACTION]
// %SeeAlso: BEGIN, ROLLBACK, WEBDOCS/commit-transaction.html
commit_stmt:
  COMMIT opt_transaction_chain
  {
    $$.val = &tree.CommitTransaction{}
  }
| COMMIT error // SHOW HELP: COMMIT
| END opt_transaction_chain
  {
    $$.val = &tree.CommitTransaction{}
  }
| END error // SHOW HELP: COMMIT

abort_stmt:
  ABORT opt_transaction_chain
  {
    $$.val = &tree.RollbackTransaction{}
  }

opt_transaction_chain:
  opt_work_transaction opt_chain {}

opt_chain:
  /* EMPTY */ {}
| AND CHAIN {}
| AND NO CHAIN {}

// %Help: ROLLBACK - abort the current (sub-)transaction
// %Category: Txn
// %Text:
// ROLLBACK [TRANSACTION]
// ROLLBACK [TRANSACTION] TO [SAVEPOINT] <savepoint name>
// %SeeAlso: BEGIN, COMMIT, SAVEPOINT, WEBDOCS/rollback-transaction.html
rollback_stmt:
  ROLLBACK opt_transaction_chain
  {
     $$.val = &tree.RollbackTransaction{}
  }
| ROLLBACK opt_work_transaction TO savepoint_name
  {
     $$.val = &tree.RollbackToSavepoint{Savepoint: tree.Name($4)}
  }
| ROLLBACK error // SHOW HELP: ROLLBACK

opt_work_transaction:
  TRANSACTION {}
| WORK        {}
| /* EMPTY */ {}

savepoint_name:
  SAVEPOINT name
  {
    $$ = $2
  }
| name
  {
    $$ = $1
  }

begin_transaction:
  transaction_mode_list
  {
    $$.val = &tree.BeginTransaction{Modes: $1.transactionModes()}
  }
| /* EMPTY */
  {
    $$.val = &tree.BeginTransaction{}
  }

transaction_mode_list:
  transaction_mode
  {
    $$.val = $1.transactionModes()
  }
| transaction_mode_list ',' transaction_mode
  {
    a := $1.transactionModes()
    b := $3.transactionModes()
    err := a.Merge(b)
    if err != nil { return setErr(sqllex, err) }
    $$.val = a
  }

transaction_mode:
  transaction_iso_level
  {
    /* SKIP DOC */
    $$.val = tree.TransactionModes{Isolation: $1.isoLevel()}
  }
| transaction_user_priority
  {
    $$.val = tree.TransactionModes{UserPriority: $1.userPriority()}
  }
| transaction_read_mode
  {
    $$.val = tree.TransactionModes{ReadWriteMode: $1.readWriteMode()}
  }
| as_of_clause
  {
    $$.val = tree.TransactionModes{AsOf: $1.asOfClause()}
  }
| deferrable_mode
  {
    $$.val = tree.TransactionModes{Deferrable: $1.deferrableMode()}
  }

transaction_user_priority:
  PRIORITY user_priority
  {
    $$.val = $2.userPriority()
  }

transaction_iso_level:
  ISOLATION LEVEL iso_level
  {
    $$.val = $3.isoLevel()
  }

transaction_read_mode:
  READ ONLY
  {
    $$.val = tree.ReadOnly
  }
| READ WRITE
  {
    $$.val = tree.ReadWrite
  }

deferrable_mode:
  DEFERRABLE
  {
    $$.val = tree.Deferrable
  }
| NOT_LA DEFERRABLE
  {
    $$.val = tree.NotDeferrable
  }

// %Help: CREATE DATABASE - create a new database
// %Category: DDL
// %Text: CREATE DATABASE [IF NOT EXISTS] <name>
// %SeeAlso: WEBDOCS/create-database.html
create_database_stmt:
  CREATE DATABASE database_name opt_with opt_owner opt_template opt_encoding opt_strategy opt_locale opt_lc_collate opt_lc_ctype opt_icu_locale opt_icu_rules opt_locale_provider opt_collation_version opt_tablespace opt_allow_connections opt_connection_limit opt_is_template opt_oid
  {
    $$.val = &tree.CreateDatabase{
      Name: tree.Name($3),
      Owner: $5,
      Template: $6,
      Encoding: $7,
      Strategy: $8,
      Locale: $9,
      Collate: $10,
      CType: $11,
      IcuLocale: $12,
      IcuRules: $13,
      LocaleProvider: $14,
      CollationVersion: $15,
      Tablespace: $16,
      AllowConnections: $17.expr(),
      ConnectionLimit: $18.expr(),
      IsTemplate: $19.expr(),
      Oid: $20.expr(),
    }
  }
| CREATE DATABASE IF NOT EXISTS database_name opt_with opt_owner opt_template opt_encoding opt_strategy opt_locale opt_lc_collate opt_lc_ctype opt_icu_locale opt_icu_rules opt_locale_provider opt_collation_version opt_tablespace opt_allow_connections opt_connection_limit opt_is_template opt_oid
  {
    $$.val = &tree.CreateDatabase{
      IfNotExists: true,
      Name: tree.Name($6),
      Owner: $8,
      Template: $9,
      Encoding: $10,
      Strategy: $11,
      Locale: $12,
      Collate: $13,
      CType: $14,
      IcuLocale: $15,
      IcuRules: $16,
      LocaleProvider: $17,
      CollationVersion: $18,
      Tablespace: $19,
      AllowConnections: $20.expr(),
      ConnectionLimit: $21.expr(),
      IsTemplate: $22.expr(),
      Oid: $23.expr(),
    }
   }
| CREATE DATABASE error // SHOW HELP: CREATE DATABASE

// Optional parameters can be written in any order, not only the order illustrated above.
opt_owner:
  OWNER opt_equal role_spec
  {
    $$ = $3
  }
| /* EMPTY */
  {
    $$ = ""
  }

opt_template:
  TEMPLATE opt_equal non_reserved_word_or_sconst
  {
    $$ = $3
  }
| /* EMPTY */
  {
    $$ = ""
  }

opt_encoding:
  ENCODING opt_equal non_reserved_word_or_sconst
  {
    $$ = $3
  }
| /* EMPTY */
  {
    $$ = ""
  }

opt_strategy:
  STRATEGY opt_equal non_reserved_word_or_sconst
  {
    $$ = $3
  }
| /* EMPTY */
  {
    $$ = ""
  }

opt_locale:
  LOCALE opt_equal non_reserved_word_or_sconst
  {
    $$ = $3
  }
| /* EMPTY */
  {
    $$ = ""
  }

opt_lc_collate:
  LC_COLLATE opt_equal non_reserved_word_or_sconst
  {
    $$ = $3
  }
| /* EMPTY */
  {
    $$ = ""
  }

opt_lc_ctype:
  LC_CTYPE opt_equal non_reserved_word_or_sconst
  {
    $$ = $3
  }
| /* EMPTY */
  {
    $$ = ""
  }

opt_icu_locale:
  ICU_LOCALE opt_equal non_reserved_word_or_sconst
  {
    $$ = $3
  }
| /* EMPTY */
  {
    $$ = ""
  }

opt_icu_rules:
  ICU_RULES opt_equal non_reserved_word_or_sconst
  {
    $$ = $3
  }
| /* EMPTY */
  {
    $$ = ""
  }

opt_locale_provider:
  LOCALE_PROVIDER opt_equal non_reserved_word_or_sconst
  {
    $$ = $3
  }
| /* EMPTY */
  {
    $$ = ""
  }

opt_collation_version:
  COLLATION_VERSION opt_equal non_reserved_word_or_sconst
  {
    $$ = $3
  }
| /* EMPTY */
  {
    $$ = ""
  }

opt_tablespace:
  TABLESPACE opt_equal tablespace_name
  {
    $$ = $3
  }
| /* EMPTY */
  {
    $$ = ""
  }

opt_allow_connections:
  ALLOW_CONNECTIONS opt_equal a_expr
  {
    $$.val = $3.expr()
  }
| /* EMPTY */
  {
    $$.val = nil
  }

opt_connection_limit:
  CONNECTION LIMIT opt_equal signed_iconst
  {
    $$.val = $4.expr()
  }
| /* EMPTY */
  {
    $$.val = nil
  }

opt_is_template:
  IS_TEMPLATE opt_equal a_expr
  {
    $$.val = $3.expr()
  }
| /* EMPTY */
  {
    $$.val = nil
  }

opt_oid:
  OID opt_equal signed_iconst
  {
    $$.val = $3.expr()
  }
| /* EMPTY */
  {
    $$.val = nil
  }

opt_equal:
  '=' {}
| /* EMPTY */ {}

// %Help: INSERT - create new rows in a table
// %Category: DML
// %Text:
// INSERT INTO <tablename> [[AS] <name>] [( <colnames...> )]
//        <selectclause>
//        [ON CONFLICT {
//          [( <colnames...> )] [WHERE <arbiter_predicate>] DO NOTHING |
//          ( <colnames...> ) [WHERE <index_predicate>] DO UPDATE SET ... [WHERE <expr>]
//        }
//        [RETURNING <exprs...>]
// %SeeAlso: UPSERT, UPDATE, DELETE, WEBDOCS/insert.html
insert_stmt:
  opt_with_clause INSERT INTO insert_target insert_rest returning_clause
  {
    $$.val = $5.stmt()
    $$.val.(*tree.Insert).With = $1.with()
    $$.val.(*tree.Insert).Table = $4.tblExpr()
    $$.val.(*tree.Insert).Returning = $6.retClause()
  }
| opt_with_clause INSERT INTO insert_target insert_rest on_conflict returning_clause
  {
    $$.val = $5.stmt()
    $$.val.(*tree.Insert).With = $1.with()
    $$.val.(*tree.Insert).Table = $4.tblExpr()
    $$.val.(*tree.Insert).OnConflict = $6.onConflict()
    $$.val.(*tree.Insert).Returning = $7.retClause()
  }
| opt_with_clause INSERT error // SHOW HELP: INSERT

// %Help: UPSERT - create or replace rows in a table
// %Category: DML
// %Text:
// UPSERT INTO <tablename> [AS <name>] [( <colnames...> )]
//        <selectclause>
//        [RETURNING <exprs...>]
// %SeeAlso: INSERT, UPDATE, DELETE, WEBDOCS/upsert.html
upsert_stmt:
  opt_with_clause UPSERT INTO insert_target insert_rest returning_clause
  {
    $$.val = $5.stmt()
    $$.val.(*tree.Insert).With = $1.with()
    $$.val.(*tree.Insert).Table = $4.tblExpr()
    $$.val.(*tree.Insert).OnConflict = &tree.OnConflict{}
    $$.val.(*tree.Insert).Returning = $6.retClause()
  }
| opt_with_clause UPSERT error // SHOW HELP: UPSERT

insert_target:
  table_name
  {
    name := $1.unresolvedObjectName().ToTableName()
    $$.val = &name
  }
// Can't easily make AS optional here, because VALUES in insert_rest would have
// a shift/reduce conflict with VALUES as an optional alias. We could easily
// allow unreserved_keywords as optional aliases, but that'd be an odd
// divergence from other places. So just require AS for now.
| table_name AS table_alias_name
  {
    name := $1.unresolvedObjectName().ToTableName()
    $$.val = &tree.AliasedTableExpr{Expr: &name, As: tree.AliasClause{Alias: tree.Name($3)}}
  }
| numeric_table_ref
  {
    $$.val = $1.tblExpr()
  }

insert_rest:
  select_stmt
  {
    $$.val = &tree.Insert{Rows: $1.slct()}
  }
| '(' insert_column_list ')' select_stmt
  {
    $$.val = &tree.Insert{Columns: $2.nameList(), Rows: $4.slct()}
  }
| DEFAULT VALUES
  {
    $$.val = &tree.Insert{Rows: &tree.Select{}}
  }

insert_column_list:
  insert_column_item
  {
    $$.val = tree.NameList{tree.Name($1)}
  }
| insert_column_list ',' insert_column_item
  {
    $$.val = append($1.nameList(), tree.Name($3))
  }

// insert_column_item represents the target of an INSERT/UPSERT or one
// of the LHS operands in an UPDATE SET statement.
//
//    INSERT INTO foo (x, y) VALUES ...
//                     ^^^^ here
//
//    UPDATE foo SET x = 1+2, (y, z) = (4, 5)
//                   ^^ here   ^^^^ here
//
// Currently CockroachDB only supports simple column names in this
// position. The rule below can be extended to support a sequence of
// field subscript or array indexing operators to designate a part of
// a field, when partial updates are to be supported. This likely will
// be needed together with support for composite types (#27792).
insert_column_item:
  column_name
| column_name '.' error { return unimplementedWithIssue(sqllex, 27792) }

on_conflict:
  ON CONFLICT DO NOTHING
  {
    $$.val = &tree.OnConflict{
      Columns: tree.NameList(nil),
      DoNothing: true,
    }
  }
| ON CONFLICT '(' name_list ')' opt_where_clause DO NOTHING
  {
    $$.val = &tree.OnConflict{
      Columns: $4.nameList(),
      ArbiterPredicate: $6.expr(),
      DoNothing: true,
    }
  }
| ON CONFLICT '(' name_list ')' opt_where_clause DO UPDATE SET set_clause_list opt_where_clause
  {
    $$.val = &tree.OnConflict{
      Columns: $4.nameList(),
      ArbiterPredicate: $6.expr(),
      Exprs: $10.updateExprs(),
      Where: tree.NewWhere(tree.AstWhere, $11.expr()),
    }
  }
| ON CONFLICT ON CONSTRAINT constraint_name { return unimplementedWithIssue(sqllex, 28161) }

returning_clause:
  RETURNING target_list
  {
    ret := tree.ReturningExprs($2.selExprs())
    $$.val = &ret
  }
| RETURNING NOTHING
  {
    $$.val = tree.ReturningNothingClause
  }
| /* EMPTY */
  {
    $$.val = tree.AbsentReturningClause
  }

// %Help: UPDATE - update rows of a table
// %Category: DML
// %Text:
// UPDATE <tablename> [[AS] <name>]
//        SET ...
//        [WHERE <expr>]
//        [ORDER BY <exprs...>]
//        [LIMIT <expr>]
//        [RETURNING <exprs...>]
// %SeeAlso: INSERT, UPSERT, DELETE, WEBDOCS/update.html
update_stmt:
  opt_with_clause UPDATE table_expr_opt_alias_idx
    SET set_clause_list opt_from_list opt_where_clause opt_sort_clause opt_limit_clause returning_clause
  {
    $$.val = &tree.Update{
      With: $1.with(),
      Table: $3.tblExpr(),
      Exprs: $5.updateExprs(),
      From: $6.tblExprs(),
      Where: tree.NewWhere(tree.AstWhere, $7.expr()),
      OrderBy: $8.orderBy(),
      Limit: $9.limit(),
      Returning: $10.retClause(),
    }
  }
| opt_with_clause UPDATE error // SHOW HELP: UPDATE

opt_from_list:
  FROM from_list {
    $$.val = $2.tblExprs()
  }
| /* EMPTY */ {
    $$.val = tree.TableExprs{}
}

set_clause_list:
  set_clause
  {
    $$.val = tree.UpdateExprs{$1.updateExpr()}
  }
| set_clause_list ',' set_clause
  {
    $$.val = append($1.updateExprs(), $3.updateExpr())
  }

// TODO(knz): The LHS in these can be extended to support
// a path to a field member when compound types are supported.
// Keep it simple for now.
set_clause:
  single_set_clause
| multiple_set_clause

single_set_clause:
  column_name '=' a_expr
  {
    $$.val = &tree.UpdateExpr{Names: tree.NameList{tree.Name($1)}, Expr: $3.expr()}
  }
| column_name '.' error { return unimplementedWithIssue(sqllex, 27792) }

multiple_set_clause:
  '(' insert_column_list ')' '=' in_expr
  {
    $$.val = &tree.UpdateExpr{Tuple: true, Names: $2.nameList(), Expr: $5.expr()}
  }

// A complete SELECT statement looks like this.
//
// The rule returns either a single select_stmt node or a tree of them,
// representing a set-operation tree.
//
// There is an ambiguity when a sub-SELECT is within an a_expr and there are
// excess parentheses: do the parentheses belong to the sub-SELECT or to the
// surrounding a_expr?  We don't really care, but bison wants to know. To
// resolve the ambiguity, we are careful to define the grammar so that the
// decision is staved off as long as possible: as long as we can keep absorbing
// parentheses into the sub-SELECT, we will do so, and only when it's no longer
// possible to do that will we decide that parens belong to the expression. For
// example, in "SELECT (((SELECT 2)) + 3)" the extra parentheses are treated as
// part of the sub-select. The necessity of doing it that way is shown by
// "SELECT (((SELECT 2)) UNION SELECT 2)". Had we parsed "((SELECT 2))" as an
// a_expr, it'd be too late to go back to the SELECT viewpoint when we see the
// UNION.
//
// This approach is implemented by defining a nonterminal select_with_parens,
// which represents a SELECT with at least one outer layer of parentheses, and
// being careful to use select_with_parens, never '(' select_stmt ')', in the
// expression grammar. We will then have shift-reduce conflicts which we can
// resolve in favor of always treating '(' <select> ')' as a
// select_with_parens. To resolve the conflicts, the productions that conflict
// with the select_with_parens productions are manually given precedences lower
// than the precedence of ')', thereby ensuring that we shift ')' (and then
// reduce to select_with_parens) rather than trying to reduce the inner
// <select> nonterminal to something else. We use UMINUS precedence for this,
// which is a fairly arbitrary choice.
//
// To be able to define select_with_parens itself without ambiguity, we need a
// nonterminal select_no_parens that represents a SELECT structure with no
// outermost parentheses. This is a little bit tedious, but it works.
//
// In non-expression contexts, we use select_stmt which can represent a SELECT
// with or without outer parentheses.
select_stmt:
  select_no_parens %prec UMINUS
| select_with_parens %prec UMINUS
  {
    $$.val = &tree.Select{Select: $1.selectStmt()}
  }

select_with_parens:
  '(' select_no_parens ')'
  {
    $$.val = &tree.ParenSelect{Select: $2.slct()}
  }
| '(' select_with_parens ')'
  {
    $$.val = &tree.ParenSelect{Select: &tree.Select{Select: $2.selectStmt()}}
  }

// This rule parses the equivalent of the standard's <query expression>. The
// duplicative productions are annoying, but hard to get rid of without
// creating shift/reduce conflicts.
//
//      The locking clause (FOR UPDATE etc) may be before or after
//      LIMIT/OFFSET. In <=7.2.X, LIMIT/OFFSET had to be after FOR UPDATE We
//      now support both orderings, but prefer LIMIT/OFFSET before the locking
//      clause.
//      - 2002-08-28 bjm
select_no_parens:
  simple_select
  {
    $$.val = &tree.Select{Select: $1.selectStmt()}
  }
| select_clause sort_clause
  {
    $$.val = &tree.Select{Select: $1.selectStmt(), OrderBy: $2.orderBy()}
  }
| select_clause opt_sort_clause for_locking_clause opt_select_limit
  {
    $$.val = &tree.Select{Select: $1.selectStmt(), OrderBy: $2.orderBy(), Limit: $4.limit(), Locking: $3.lockingClause()}
  }
| select_clause opt_sort_clause select_limit opt_for_locking_clause
  {
    $$.val = &tree.Select{Select: $1.selectStmt(), OrderBy: $2.orderBy(), Limit: $3.limit(), Locking: $4.lockingClause()}
  }
| with_clause select_clause
  {
    $$.val = &tree.Select{With: $1.with(), Select: $2.selectStmt()}
  }
| with_clause select_clause sort_clause
  {
    $$.val = &tree.Select{With: $1.with(), Select: $2.selectStmt(), OrderBy: $3.orderBy()}
  }
| with_clause select_clause opt_sort_clause for_locking_clause opt_select_limit
  {
    $$.val = &tree.Select{With: $1.with(), Select: $2.selectStmt(), OrderBy: $3.orderBy(), Limit: $5.limit(), Locking: $4.lockingClause()}
  }
| with_clause select_clause opt_sort_clause select_limit opt_for_locking_clause
  {
    $$.val = &tree.Select{With: $1.with(), Select: $2.selectStmt(), OrderBy: $3.orderBy(), Limit: $4.limit(), Locking: $5.lockingClause()}
  }

for_locking_clause:
  for_locking_items { $$.val = $1.lockingClause() }
| FOR READ ONLY     { $$.val = (tree.LockingClause)(nil) }

opt_for_locking_clause:
  for_locking_clause { $$.val = $1.lockingClause() }
| /* EMPTY */        { $$.val = (tree.LockingClause)(nil) }

for_locking_items:
  for_locking_item
  {
    $$.val = tree.LockingClause{$1.lockingItem()}
  }
| for_locking_items for_locking_item
  {
    $$.val = append($1.lockingClause(), $2.lockingItem())
  }

for_locking_item:
  for_locking_strength opt_locked_rels opt_nowait_or_skip
  {
    $$.val = &tree.LockingItem{
      Strength:   $1.lockingStrength(),
      Targets:    $2.tableNames(),
      WaitPolicy: $3.lockingWaitPolicy(),
    }
  }

for_locking_strength:
  FOR UPDATE        { $$.val = tree.ForUpdate }
| FOR NO KEY UPDATE { $$.val = tree.ForNoKeyUpdate }
| FOR SHARE         { $$.val = tree.ForShare }
| FOR KEY SHARE     { $$.val = tree.ForKeyShare }

opt_locked_rels:
  /* EMPTY */        { $$.val = tree.TableNames{} }
| OF table_name_list { $$.val = $2.tableNames() }

opt_nowait_or_skip:
  /* EMPTY */ { $$.val = tree.LockWaitBlock }
| SKIP LOCKED { $$.val = tree.LockWaitSkip }
| NOWAIT      { $$.val = tree.LockWaitError }

select_clause:
// We only provide help if an open parenthesis is provided, because
// otherwise the rule is ambiguous with the top-level statement list.
  '(' error // SHOW HELP: <SELECTCLAUSE>
| simple_select
| select_with_parens

// This rule parses SELECT statements that can appear within set operations,
// including UNION, INTERSECT and EXCEPT. '(' and ')' can be used to specify
// the ordering of the set operations. Without '(' and ')' we want the
// operations to be ordered per the precedence specs at the head of this file.
//
// As with select_no_parens, simple_select cannot have outer parentheses, but
// can have parenthesized subclauses.
//
// Note that sort clauses cannot be included at this level --- SQL requires
//       SELECT foo UNION SELECT bar ORDER BY baz
// to be parsed as
//       (SELECT foo UNION SELECT bar) ORDER BY baz
// not
//       SELECT foo UNION (SELECT bar ORDER BY baz)
//
// Likewise for WITH, FOR UPDATE and LIMIT. Therefore, those clauses are
// described as part of the select_no_parens production, not simple_select.
// This does not limit functionality, because you can reintroduce these clauses
// inside parentheses.
//
// NOTE: only the leftmost component select_stmt should have INTO. However,
// this is not checked by the grammar; parse analysis must check it.
//
// %Help: <SELECTCLAUSE> - access tabular data
// %Category: DML
// %Text:
// Select clause:
//   TABLE <tablename>
//   VALUES ( <exprs...> ) [ , ... ]
//   SELECT ... [ { INTERSECT | UNION | EXCEPT } [ ALL | DISTINCT ] <selectclause> ]
simple_select:
  simple_select_clause // EXTEND WITH HELP: SELECT
| empty_select
| values_clause        // EXTEND WITH HELP: VALUES
| table_clause         // EXTEND WITH HELP: TABLE
| set_operation

// Postgres allows select expressions to be omitted, when causes the select statement to
// return empty rows for any matches. Changing the existing select rules to make from_list
// optional cause shift/reduce conflicts, so this rule was added to work around that.
empty_select:
  SELECT FROM from_list opt_where_clause
    group_clause having_clause window_clause
  {
    $$.val = &tree.SelectClause{
      Exprs:   make(tree.SelectExprs, 0, 0),
      From:    tree.From{Tables: $3.tblExprs()},
      Where:   tree.NewWhere(tree.AstWhere, $4.expr()),
      GroupBy: $5.groupBy(),
      Having:  tree.NewWhere(tree.AstHaving, $6.expr()),
      Window:  $7.window(),
    }
  }

// %Help: SELECT - retrieve rows from a data source and compute a result
// %Category: DML
// %Text:
// SELECT [DISTINCT [ ON ( <expr> [ , ... ] ) ] ]
//        { <expr> [[AS] <name>] | [ [<dbname>.] <tablename>. ] * } [, ...]
//        [ FROM <source> ]
//        [ WHERE <expr> ]
//        [ GROUP BY <expr> [ , ... ] ]
//        [ HAVING <expr> ]
//        [ WINDOW <name> AS ( <definition> ) ]
//        [ { UNION | INTERSECT | EXCEPT } [ ALL | DISTINCT ] <selectclause> ]
//        [ ORDER BY <expr> [ ASC | DESC ] [, ...] ]
//        [ LIMIT { <expr> | ALL } ]
//        [ OFFSET <expr> [ ROW | ROWS ] ]
// %SeeAlso: WEBDOCS/select-clause.html
simple_select_clause:
  SELECT opt_all_clause target_list
    from_clause opt_where_clause
    group_clause having_clause window_clause
  {
    $$.val = &tree.SelectClause{
      Exprs:   $3.selExprs(),
      From:    $4.from(),
      Where:   tree.NewWhere(tree.AstWhere, $5.expr()),
      GroupBy: $6.groupBy(),
      Having:  tree.NewWhere(tree.AstHaving, $7.expr()),
      Window:  $8.window(),
    }
  }
| SELECT distinct_clause target_list
    from_clause opt_where_clause
    group_clause having_clause window_clause
  {
    $$.val = &tree.SelectClause{
      Distinct: $2.bool(),
      Exprs:    $3.selExprs(),
      From:     $4.from(),
      Where:    tree.NewWhere(tree.AstWhere, $5.expr()),
      GroupBy:  $6.groupBy(),
      Having:   tree.NewWhere(tree.AstHaving, $7.expr()),
      Window:   $8.window(),
    }
  }
| SELECT distinct_on_clause target_list
    from_clause opt_where_clause
    group_clause having_clause window_clause
  {
    $$.val = &tree.SelectClause{
      Distinct:   true,
      DistinctOn: $2.distinctOn(),
      Exprs:      $3.selExprs(),
      From:       $4.from(),
      Where:      tree.NewWhere(tree.AstWhere, $5.expr()),
      GroupBy:    $6.groupBy(),
      Having:     tree.NewWhere(tree.AstHaving, $7.expr()),
      Window:     $8.window(),
    }
  }
| SELECT error // SHOW HELP: SELECT

set_operation:
  select_clause UNION all_or_distinct select_clause
  {
    $$.val = &tree.UnionClause{
      Type:  tree.UnionOp,
      Left:  &tree.Select{Select: $1.selectStmt()},
      Right: &tree.Select{Select: $4.selectStmt()},
      All:   $3.bool(),
    }
  }
| select_clause INTERSECT all_or_distinct select_clause
  {
    $$.val = &tree.UnionClause{
      Type:  tree.IntersectOp,
      Left:  &tree.Select{Select: $1.selectStmt()},
      Right: &tree.Select{Select: $4.selectStmt()},
      All:   $3.bool(),
    }
  }
| select_clause EXCEPT all_or_distinct select_clause
  {
    $$.val = &tree.UnionClause{
      Type:  tree.ExceptOp,
      Left:  &tree.Select{Select: $1.selectStmt()},
      Right: &tree.Select{Select: $4.selectStmt()},
      All:   $3.bool(),
    }
  }

// %Help: TABLE - select an entire table
// %Category: DML
// %Text: TABLE <tablename>
// %SeeAlso: SELECT, VALUES, WEBDOCS/table-expressions.html
table_clause:
  TABLE table_ref
  {
    $$.val = &tree.SelectClause{
      Exprs:       tree.SelectExprs{tree.StarSelectExpr()},
      From:        tree.From{Tables: tree.TableExprs{$2.tblExpr()}},
      TableSelect: true,
    }
  }
| TABLE error // SHOW HELP: TABLE

// SQL standard WITH clause looks like:
//
// WITH [ RECURSIVE ] <query name> [ (<column> [, ...]) ]
//        AS [ [ NOT ] MATERIALIZED ] (query) [ SEARCH or CYCLE clause ]
//
// We don't currently support the SEARCH or CYCLE clause.
//
// Recognizing WITH_LA here allows a CTE to be named TIME or ORDINALITY.
with_clause:
  WITH cte_list
  {
    $$.val = &tree.With{CTEList: $2.ctes()}
  }
| WITH_LA cte_list
  {
    /* SKIP DOC */
    $$.val = &tree.With{CTEList: $2.ctes()}
  }
| WITH RECURSIVE cte_list
  {
    $$.val = &tree.With{Recursive: true, CTEList: $3.ctes()}
  }

cte_list:
  common_table_expr
  {
    $$.val = []*tree.CTE{$1.cte()}
  }
| cte_list ',' common_table_expr
  {
    $$.val = append($1.ctes(), $3.cte())
  }

materialize_clause:
  MATERIALIZED
  {
    $$.val = true
  }
| NOT MATERIALIZED
  {
    $$.val = false
  }

common_table_expr:
  table_alias_name opt_column_list AS '(' preparable_stmt ')'
    {
      $$.val = &tree.CTE{
        Name: tree.AliasClause{Alias: tree.Name($1), Cols: $2.nameList() },
        Mtr: tree.MaterializeClause{
          Set: false,
        },
        Stmt: $5.stmt(),
      }
    }
| table_alias_name opt_column_list AS materialize_clause '(' preparable_stmt ')'
    {
      $$.val = &tree.CTE{
        Name: tree.AliasClause{Alias: tree.Name($1), Cols: $2.nameList() },
        Mtr: tree.MaterializeClause{
          Materialize: $4.bool(),
          Set: true,
        },
        Stmt: $6.stmt(),
      }
    }

opt_with:
  WITH {}
| /* EMPTY */ {}

opt_with_clause:
  with_clause
  {
    $$.val = $1.with()
  }
| /* EMPTY */
  {
    $$.val = nil
  }

opt_table:
  TABLE {}
| /* EMPTY */ {}

all_or_distinct:
  ALL
  {
    $$.val = true
  }
| DISTINCT
  {
    $$.val = false
  }
| /* EMPTY */
  {
    $$.val = false
  }

distinct_clause:
  DISTINCT
  {
    $$.val = true
  }

distinct_on_clause:
  DISTINCT ON '(' expr_list ')'
  {
    $$.val = tree.DistinctOn($4.exprs())
  }

opt_all_clause:
  ALL {}
| /* EMPTY */ {}

opt_sort_clause:
  sort_clause
  {
    $$.val = $1.orderBy()
  }
| /* EMPTY */
  {
    $$.val = tree.OrderBy(nil)
  }

sort_clause:
  ORDER BY sortby_list
  {
    $$.val = tree.OrderBy($3.orders())
  }

single_sort_clause:
  ORDER BY sortby
  {
    $$.val = tree.OrderBy([]*tree.Order{$3.order()})
  }
| ORDER BY sortby ',' sortby_list
  {
    sqllex.Error("multiple ORDER BY clauses are not supported in this function")
    return 1
  }

sortby_list:
  sortby
  {
    $$.val = []*tree.Order{$1.order()}
  }
| sortby_list ',' sortby
  {
    $$.val = append($1.orders(), $3.order())
  }

sortby:
  a_expr opt_asc_desc opt_nulls_order
  {
    /* FORCE DOC */
    dir := $2.dir()
    nullsOrder := $3.nullsOrder()
    // We currently only support the opposite of Postgres defaults.
    if nullsOrder != tree.DefaultNullsOrder {
      if dir == tree.Descending && nullsOrder == tree.NullsFirst {
        return unimplementedWithIssue(sqllex, 6224)
      }
      if dir != tree.Descending && nullsOrder == tree.NullsLast {
        return unimplementedWithIssue(sqllex, 6224)
      }
    }
    $$.val = &tree.Order{
      OrderType:  tree.OrderByColumn,
      Expr:       $1.expr(),
      Direction:  dir,
      NullsOrder: nullsOrder,
    }
  }
| PRIMARY KEY table_name opt_asc_desc
  {
    name := $3.unresolvedObjectName().ToTableName()
    $$.val = &tree.Order{OrderType: tree.OrderByIndex, Direction: $4.dir(), Table: name}
  }
| INDEX table_name '@' index_name opt_asc_desc
  {
    name := $2.unresolvedObjectName().ToTableName()
    $$.val = &tree.Order{
      OrderType: tree.OrderByIndex,
      Direction: $5.dir(),
      Table:     name,
      Index:     tree.UnrestrictedName($4),
    }
  }

opt_nulls_order:
  NULLS FIRST
  {
    $$.val = tree.NullsFirst
  }
| NULLS LAST
  {
    $$.val = tree.NullsLast
  }
| /* EMPTY */
  {
    $$.val = tree.DefaultNullsOrder
  }

// TODO(pmattis): Support ordering using arbitrary math ops?
// | a_expr USING math_op {}

select_limit:
  limit_clause offset_clause
  {
    if $1.limit() == nil {
      $$.val = $2.limit()
    } else {
      $$.val = $1.limit()
      $$.val.(*tree.Limit).Offset = $2.limit().Offset
    }
  }
| offset_clause limit_clause
  {
    $$.val = $1.limit()
    if $2.limit() != nil {
      $$.val.(*tree.Limit).Count = $2.limit().Count
      $$.val.(*tree.Limit).LimitAll = $2.limit().LimitAll
    }
  }
| limit_clause
| offset_clause

opt_select_limit:
  select_limit { $$.val = $1.limit() }
| /* EMPTY */  { $$.val = (*tree.Limit)(nil) }

opt_limit_clause:
  limit_clause
| /* EMPTY */ { $$.val = (*tree.Limit)(nil) }

limit_clause:
  LIMIT ALL
  {
    $$.val = &tree.Limit{LimitAll: true}
  }
| LIMIT a_expr
  {
    if $2.expr() == nil {
      $$.val = (*tree.Limit)(nil)
    } else {
      $$.val = &tree.Limit{Count: $2.expr()}
    }
  }
// SQL:2008 syntax
// To avoid shift/reduce conflicts, handle the optional value with
// a separate production rather than an opt_ expression. The fact
// that ONLY is fully reserved means that this way, we defer any
// decision about what rule reduces ROW or ROWS to the point where
// we can see the ONLY token in the lookahead slot.
| FETCH first_or_next select_fetch_first_value row_or_rows ONLY
  {
    $$.val = &tree.Limit{Count: $3.expr()}
  }
| FETCH first_or_next row_or_rows ONLY
	{
    $$.val = &tree.Limit{
      Count: tree.NewNumVal(constant.MakeInt64(1), "" /* origString */, false /* negative */),
    }
  }

offset_clause:
  OFFSET a_expr
  {
    $$.val = &tree.Limit{Offset: $2.expr()}
  }
  // SQL:2008 syntax
  // The trailing ROW/ROWS in this case prevent the full expression
  // syntax. c_expr is the best we can do.
| OFFSET select_fetch_first_value row_or_rows
  {
    $$.val = &tree.Limit{Offset: $2.expr()}
  }

// Allowing full expressions without parentheses causes various parsing
// problems with the trailing ROW/ROWS key words. SQL spec only calls for
// <simple value specification>, which is either a literal or a parameter (but
// an <SQL parameter reference> could be an identifier, bringing up conflicts
// with ROW/ROWS). We solve this by leveraging the presence of ONLY (see above)
// to determine whether the expression is missing rather than trying to make it
// optional in this rule.
//
// c_expr covers almost all the spec-required cases (and more), but it doesn't
// cover signed numeric literals, which are allowed by the spec. So we include
// those here explicitly.
select_fetch_first_value:
  c_expr
| only_signed_iconst
| only_signed_fconst

// noise words
row_or_rows:
  ROW {}
| ROWS {}

first_or_next:
  FIRST {}
| NEXT {}

// This syntax for group_clause tries to follow the spec quite closely.
// However, the spec allows only column references, not expressions,
// which introduces an ambiguity between implicit row constructors
// (a,b) and lists of column references.
//
// We handle this by using the a_expr production for what the spec calls
// <ordinary grouping set>, which in the spec represents either one column
// reference or a parenthesized list of column references. Then, we check the
// top node of the a_expr to see if it's an RowExpr, and if so, just grab and
// use the list, discarding the node. (this is done in parse analysis, not here)
//
// Each item in the group_clause list is either an expression tree or a
// GroupingSet node of some type.
group_clause:
  GROUP BY expr_list
  {
    $$.val = tree.GroupBy($3.exprs())
  }
| /* EMPTY */
  {
    $$.val = tree.GroupBy(nil)
  }

having_clause:
  HAVING a_expr
  {
    $$.val = $2.expr()
  }
| /* EMPTY */
  {
    $$.val = tree.Expr(nil)
  }

// Given "VALUES (a, b)" in a table expression context, we have to
// decide without looking any further ahead whether VALUES is the
// values clause or a set-generating function. Since VALUES is allowed
// as a function name both interpretations are feasible. We resolve
// the shift/reduce conflict by giving the first values_clause
// production a higher precedence than the VALUES token has, causing
// the parser to prefer to reduce, in effect assuming that the VALUES
// is not a function name.
//
// %Help: VALUES - select a given set of values
// %Category: DML
// %Text: VALUES ( <exprs...> ) [, ...]
// %SeeAlso: SELECT, TABLE, WEBDOCS/table-expressions.html
values_clause:
  VALUES '(' expr_list ')' %prec UMINUS
  {
    $$.val = &tree.ValuesClause{Rows: []tree.Exprs{$3.exprs()}}
  }
| VALUES error // SHOW HELP: VALUES
| values_clause ',' '(' expr_list ')'
  {
    valNode := $1.selectStmt().(*tree.ValuesClause)
    valNode.Rows = append(valNode.Rows, $4.exprs())
    $$.val = valNode
  }

// clauses common to all optimizable statements:
//  from_clause   - allow list of both JOIN expressions and table names
//  where_clause  - qualifications for joins or restrictions

from_clause:
  FROM from_list
  {
    $$.val = tree.From{Tables: $2.tblExprs()}
  }
| FROM error // SHOW HELP: <SOURCE>
| /* EMPTY */
  {
    $$.val = tree.From{}
  }

from_list:
  table_ref
  {
    $$.val = tree.TableExprs{$1.tblExpr()}
  }
| from_list ',' table_ref
  {
    $$.val = append($1.tblExprs(), $3.tblExpr())
  }

index_flags_param:
  FORCE_INDEX '=' index_name
  {
     $$.val = &tree.IndexFlags{Index: tree.UnrestrictedName($3)}
  }
| FORCE_INDEX '=' '[' iconst64 ']'
  {
    /* SKIP DOC */
    $$.val = &tree.IndexFlags{IndexID: tree.IndexID($4.int64())}
  }
| ASC
  {
    /* SKIP DOC */
    $$.val = &tree.IndexFlags{Direction: tree.Ascending}
  }
| DESC
  {
    /* SKIP DOC */
    $$.val = &tree.IndexFlags{Direction: tree.Descending}
  }
|
  NO_INDEX_JOIN
  {
    $$.val = &tree.IndexFlags{NoIndexJoin: true}
  }
|
  IGNORE_FOREIGN_KEYS
  {
    /* SKIP DOC */
    $$.val = &tree.IndexFlags{IgnoreForeignKeys: true}
  }

index_flags_param_list:
  index_flags_param
  {
    $$.val = $1.indexFlags()
  }
|
  index_flags_param_list ',' index_flags_param
  {
    a := $1.indexFlags()
    b := $3.indexFlags()
    if err := a.CombineWith(b); err != nil {
      return setErr(sqllex, err)
    }
    $$.val = a
  }

opt_index_flags:
  '@' index_name
  {
    $$.val = &tree.IndexFlags{Index: tree.UnrestrictedName($2)}
  }
| '@' '[' iconst64 ']'
  {
    $$.val = &tree.IndexFlags{IndexID: tree.IndexID($3.int64())}
  }
| '@' '{' index_flags_param_list '}'
  {
    flags := $3.indexFlags()
    if err := flags.Check(); err != nil {
      return setErr(sqllex, err)
    }
    $$.val = flags
  }
| /* EMPTY */
  {
    $$.val = (*tree.IndexFlags)(nil)
  }

// %Help: <SOURCE> - define a data source for SELECT
// %Category: DML
// %Text:
// Data sources:
//   <tablename> [ @ { <idxname> | <indexflags> } ]
//   <tablefunc> ( <exprs...> )
//   ( { <selectclause> | <source> } )
//   <source> [AS] <alias> [( <colnames...> )]
//   <source> [ <jointype> ] JOIN <source> ON <expr>
//   <source> [ <jointype> ] JOIN <source> USING ( <colnames...> )
//   <source> NATURAL [ <jointype> ] JOIN <source>
//   <source> CROSS JOIN <source>
//   <source> WITH ORDINALITY
//   '[' EXPLAIN ... ']'
//   '[' SHOW ... ']'
//
// Index flags:
//   '{' FORCE_INDEX = <idxname> [, ...] '}'
//   '{' NO_INDEX_JOIN [, ...] '}'
//   '{' IGNORE_FOREIGN_KEYS [, ...] '}'
//
// Join types:
//   { INNER | { LEFT | RIGHT | FULL } [OUTER] } [ { HASH | MERGE | LOOKUP } ]
//
// %SeeAlso: WEBDOCS/table-expressions.html
table_ref:
numeric_table_ref table_ref_options
  {
    /* SKIP DOC */
    $$ = $2
    $$.val.(*tree.AliasedTableExpr).Expr = $1.tblExpr()
  }
| relation_expr table_ref_options
  {
    /* SKIP DOC */
    $$ = $2
    name := $1.unresolvedObjectName().ToTableName()
    $$.val.(*tree.AliasedTableExpr).Expr = &name
  }
| select_with_parens opt_ordinality opt_alias_clause
  {
    $$.val = &tree.AliasedTableExpr{
      Expr:       &tree.Subquery{Select: $1.selectStmt()},
      Ordinality: $2.bool(),
      As:         $3.aliasClause(),
    }
  }
| LATERAL select_with_parens opt_ordinality opt_alias_clause
  {
    $$.val = &tree.AliasedTableExpr{
      Expr:       &tree.Subquery{Select: $2.selectStmt()},
      Ordinality: $3.bool(),
      Lateral:    true,
      As:         $4.aliasClause(),
    }
  }
| joined_table
  {
    $$.val = $1.tblExpr()
  }
| '(' joined_table ')' opt_ordinality alias_clause
  {
    $$.val = &tree.AliasedTableExpr{Expr: &tree.ParenTableExpr{Expr: $2.tblExpr()}, Ordinality: $4.bool(), As: $5.aliasClause()}
  }
| func_table opt_ordinality opt_alias_clause
  {
    f := $1.tblExpr()
    $$.val = &tree.AliasedTableExpr{
      Expr: f,
      Ordinality: $2.bool(),
      // Technically, LATERAL is always implied on an SRF, but including it
      // here makes re-printing the AST slightly tricky.
      As: $3.aliasClause(),
    }
  }
| LATERAL func_table opt_ordinality opt_alias_clause
  {
    f := $2.tblExpr()
    $$.val = &tree.AliasedTableExpr{
      Expr: f,
      Ordinality: $3.bool(),
      Lateral: true,
      As: $4.aliasClause(),
    }
  }
// The following syntax is a CockroachDB extension:
//     SELECT ... FROM [ EXPLAIN .... ] WHERE ...
//     SELECT ... FROM [ SHOW .... ] WHERE ...
//     SELECT ... FROM [ INSERT ... RETURNING ... ] WHERE ...
// A statement within square brackets can be used as a table expression (data source).
// We use square brackets for two reasons:
// - the grammar would be terribly ambiguous if we used simple
//   parentheses or no parentheses at all.
// - it carries visual semantic information, by marking the table
//   expression as radically different from the other things.
//   If a user does not know this and encounters this syntax, they
//   will know from the unusual choice that something rather different
//   is going on and may be pushed by the unusual syntax to
//   investigate further in the docs.
| '[' row_source_extension_stmt ']' opt_ordinality opt_alias_clause
  {
    $$.val = &tree.AliasedTableExpr{Expr: &tree.StatementSource{ Statement: $2.stmt() }, Ordinality: $4.bool(), As: $5.aliasClause() }
  }

 // table_ref_options is the set of all possible combinations of AS OF and alias, since the optional versions of those
 // rules create shift/reduce conflicts if they're combined in same rule
table_ref_options:
  opt_index_flags opt_ordinality
  {
    /* SKIP DOC */
    $$.val = &tree.AliasedTableExpr{
        IndexFlags: $1.indexFlags(),
        Ordinality: $2.bool(),
    }
  }
| opt_index_flags opt_ordinality alias_clause
  {
    /* SKIP DOC */
    $$.val = &tree.AliasedTableExpr{
        IndexFlags: $1.indexFlags(),
        Ordinality: $2.bool(),
        As:         $3.aliasClause(),
    }
  }
| opt_index_flags opt_ordinality as_of_clause
  {
    /* SKIP DOC */
    asOf := $3.asOfClause()    
    $$.val = &tree.AliasedTableExpr{
        IndexFlags: $1.indexFlags(),
        Ordinality: $2.bool(),
        AsOf:       &asOf,
    }
  }
| opt_index_flags opt_ordinality as_of_clause AS table_alias_name opt_column_list
  {
    /* SKIP DOC */
    alias := tree.AliasClause{Alias: tree.Name($5), Cols: $6.nameList()}
    asOf := $3.asOfClause()
    $$.val = &tree.AliasedTableExpr{
        IndexFlags: $1.indexFlags(),
        Ordinality: $2.bool(),
        AsOf:       &asOf,
        As:         alias,
    }
  }
| opt_index_flags opt_ordinality as_of_clause table_alias_name opt_column_list
  {
    /* SKIP DOC */
    alias := tree.AliasClause{Alias: tree.Name($4), Cols: $5.nameList()}
    asOf := $3.asOfClause()
    $$.val = &tree.AliasedTableExpr{
        IndexFlags: $1.indexFlags(),
        Ordinality: $2.bool(),
        AsOf:       &asOf,
        As:         alias,
    }
  }

numeric_table_ref:
  '[' iconst64 opt_tableref_col_list alias_clause ']'
  {
    /* SKIP DOC */
    $$.val = &tree.TableRef{
      TableID: $2.int64(),
      Columns: $3.tableRefCols(),
      As:      $4.aliasClause(),
    }
  }

func_table:
  func_expr_windowless
  {
    $$.val = &tree.RowsFromExpr{Items: tree.Exprs{$1.expr()}}
  }
| ROWS FROM '(' rowsfrom_list ')'
  {
    $$.val = &tree.RowsFromExpr{Items: $4.exprs()}
  }

rowsfrom_list:
  rowsfrom_item
  { $$.val = tree.Exprs{$1.expr()} }
| rowsfrom_list ',' rowsfrom_item
  { $$.val = append($1.exprs(), $3.expr()) }

rowsfrom_item:
  func_expr_windowless opt_col_def_list
  {
    $$.val = $1.expr()
  }

opt_col_def_list:
  /* EMPTY */
  { }
| AS '(' error
  { return unimplemented(sqllex, "ROWS FROM with col_def_list") }

opt_tableref_col_list:
  /* EMPTY */               { $$.val = nil }
| '(' ')'                   { $$.val = []tree.ColumnID{} }
| '(' tableref_col_list ')' { $$.val = $2.tableRefCols() }

tableref_col_list:
  iconst64
  {
    $$.val = []tree.ColumnID{tree.ColumnID($1.int64())}
  }
| tableref_col_list ',' iconst64
  {
    $$.val = append($1.tableRefCols(), tree.ColumnID($3.int64()))
  }

opt_ordinality:
  WITH_LA ORDINALITY
  {
    $$.val = true
  }
| /* EMPTY */
  {
    $$.val = false
  }

// It may seem silly to separate joined_table from table_ref, but there is
// method in SQL's madness: if you don't do it this way you get reduce- reduce
// conflicts, because it's not clear to the parser generator whether to expect
// alias_clause after ')' or not. For the same reason we must treat 'JOIN' and
// 'join_type JOIN' separately, rather than allowing join_type to expand to
// empty; if we try it, the parser generator can't figure out when to reduce an
// empty join_type right after table_ref.
//
// Note that a CROSS JOIN is the same as an unqualified INNER JOIN, and an
// INNER JOIN/ON has the same shape but a qualification expression to limit
// membership. A NATURAL JOIN implicitly matches column names between tables
// and the shape is determined by which columns are in common. We'll collect
// columns during the later transformations.

joined_table:
  '(' joined_table ')'
  {
    $$.val = &tree.ParenTableExpr{Expr: $2.tblExpr()}
  }
| table_ref CROSS opt_join_hint JOIN table_ref
  {
    $$.val = &tree.JoinTableExpr{JoinType: tree.AstCross, Left: $1.tblExpr(), Right: $5.tblExpr(), Hint: $3}
  }
| table_ref join_type opt_join_hint JOIN table_ref join_qual
  {
    $$.val = &tree.JoinTableExpr{JoinType: $2, Left: $1.tblExpr(), Right: $5.tblExpr(), Cond: $6.joinCond(), Hint: $3}
  }
| table_ref JOIN table_ref join_qual
  {
    $$.val = &tree.JoinTableExpr{Left: $1.tblExpr(), Right: $3.tblExpr(), Cond: $4.joinCond()}
  }
| table_ref NATURAL join_type opt_join_hint JOIN table_ref
  {
    $$.val = &tree.JoinTableExpr{JoinType: $3, Left: $1.tblExpr(), Right: $6.tblExpr(), Cond: tree.NaturalJoinCond{}, Hint: $4}
  }
| table_ref NATURAL JOIN table_ref
  {
    $$.val = &tree.JoinTableExpr{Left: $1.tblExpr(), Right: $4.tblExpr(), Cond: tree.NaturalJoinCond{}}
  }

alias_clause:
  AS table_alias_name opt_column_list
  {
    $$.val = tree.AliasClause{Alias: tree.Name($2), Cols: $3.nameList()}
  }
| table_alias_name opt_column_list
  {
    $$.val = tree.AliasClause{Alias: tree.Name($1), Cols: $2.nameList()}
  }

opt_alias_clause:
  alias_clause
| /* EMPTY */
  {
    $$.val = tree.AliasClause{}
  }

// as_of_clause is limited to constants and a few function expressions. The entire expressoin tree is too permissive, 
// and causes many conflicts elsewhere in the rest of the grammar.
// These clauses are chosen carefully from the d_expr list.
as_of_clause:
 AS_LA OF SYSTEM TIME SCONST
  {
    $$.val = tree.AsOfClause{Expr: tree.NewStrVal($5)}
  }
| AS_LA OF SYSTEM TIME typed_literal
  {
    $$.val = tree.AsOfClause{Expr: $5.expr()}
  }
| AS_LA OF SYSTEM TIME func_expr_common_subexpr
  {
    $$.val = tree.AsOfClause{Expr: $5.expr()}
  }
| AS_LA OF SYSTEM TIME func_application
  {
    $$.val = tree.AsOfClause{Expr: $5.expr()}
  }
| AS_LA OF SCONST
  {
    $$.val = tree.AsOfClause{Expr: tree.NewStrVal($3)}
  }
| AS_LA OF typed_literal
  {
    $$.val = tree.AsOfClause{Expr: $3.expr()}
  }
| AS_LA OF func_expr_common_subexpr
  {
    $$.val = tree.AsOfClause{Expr: $3.expr()}
  }
| AS_LA OF func_application
  {
    $$.val = tree.AsOfClause{Expr: $3.expr()}
  }
  
opt_as_of_clause:
  as_of_clause
| /* EMPTY */
  {
    $$.val = tree.AsOfClause{}
  }

join_type:
  FULL join_outer
  {
    $$ = tree.AstFull
  }
| LEFT join_outer
  {
    $$ = tree.AstLeft
  }
| RIGHT join_outer
  {
    $$ = tree.AstRight
  }
| INNER
  {
    $$ = tree.AstInner
  }

// OUTER is just noise...
join_outer:
  OUTER {}
| /* EMPTY */ {}

// Join hint specifies that the join in the query should use a
// specific method.

// The semantics are as follows:
//  - HASH forces a hash join; in other words, it disables merge and lookup
//    join. A hash join is always possible; even if there are no equality
//    columns - we consider cartesian join a degenerate case of the hash join
//    (one bucket).
//  - MERGE forces a merge join, even if it requires resorting both sides of
//    the join.
//  - LOOKUP forces a lookup join into the right side; the right side must be
//    a table with a suitable index. `LOOKUP` can only be used with INNER and
//    LEFT joins (though this is not enforced by the syntax).
//  - If it is not possible to use the algorithm in the hint, an error is
//    returned.
//  - When a join hint is specified, the two tables will not be reordered
//    by the optimizer.
opt_join_hint:
  HASH
  {
    $$ = tree.AstHash
  }
| MERGE
  {
    $$ = tree.AstMerge
  }
| LOOKUP
  {
    $$ = tree.AstLookup
  }
| /* EMPTY */
  {
    $$ = ""
  }

// JOIN qualification clauses
// Possibilities are:
//      USING ( column list ) allows only unqualified column names,
//          which must match between tables.
//      ON expr allows more general qualifications.
//
// We return USING as a List node, while an ON-expr will not be a List.
join_qual:
  USING '(' name_list ')'
  {
    $$.val = &tree.UsingJoinCond{Cols: $3.nameList()}
  }
| ON a_expr
  {
    $$.val = &tree.OnJoinCond{Expr: $2.expr()}
  }

relation_expr:
  table_name              { $$.val = $1.unresolvedObjectName() }
| table_name '*'          { $$.val = $1.unresolvedObjectName() }
| ONLY table_name         { $$.val = $2.unresolvedObjectName() }
| ONLY table_name '*'     { $$.val = $2.unresolvedObjectName() }
| ONLY '(' table_name ')' { $$.val = $3.unresolvedObjectName() }

relation_expr_list:
  relation_expr
  {
    name := $1.unresolvedObjectName().ToTableName()
    $$.val = tree.TableNames{name}
  }
| relation_expr_list ',' relation_expr
  {
    name := $3.unresolvedObjectName().ToTableName()
    $$.val = append($1.tableNames(), name)
  }

// Given "UPDATE foo set set ...", we have to decide without looking any
// further ahead whether the first "set" is an alias or the UPDATE's SET
// keyword. Since "set" is allowed as a column name both interpretations are
// feasible. We resolve the shift/reduce conflict by giving the first
// table_expr_opt_alias_idx production a higher precedence than the SET token
// has, causing the parser to prefer to reduce, in effect assuming that the SET
// is not an alias.
table_expr_opt_alias_idx:
  table_name_opt_idx %prec UMINUS
  {
     $$.val = $1.tblExpr()
  }
| table_name_opt_idx table_alias_name
  {
     alias := $1.tblExpr().(*tree.AliasedTableExpr)
     alias.As = tree.AliasClause{Alias: tree.Name($2)}
     $$.val = alias
  }
| table_name_opt_idx AS table_alias_name
  {
     alias := $1.tblExpr().(*tree.AliasedTableExpr)
     alias.As = tree.AliasClause{Alias: tree.Name($3)}
     $$.val = alias
  }
| numeric_table_ref opt_index_flags
  {
    /* SKIP DOC */
    $$.val = &tree.AliasedTableExpr{
      Expr: $1.tblExpr(),
      IndexFlags: $2.indexFlags(),
    }
  }

table_name_opt_idx:
  table_name opt_index_flags
  {
    name := $1.unresolvedObjectName().ToTableName()
    $$.val = &tree.AliasedTableExpr{
      Expr: &name,
      IndexFlags: $2.indexFlags(),
    }
  }

where_clause_paren:
  WHERE '(' a_expr ')'
  {
    $$.val = $3.expr()
  }

opt_where_clause_paren:
  where_clause_paren
| /* EMPTY */
  {
    $$.val = tree.Expr(nil)
  }

where_clause:
  WHERE a_expr
  {
    $$.val = $2.expr()
  }

opt_where_clause:
  where_clause
| /* EMPTY */
  {
    $$.val = tree.Expr(nil)
  }

// Type syntax
//   SQL introduces a large amount of type-specific syntax.
//   Define individual clauses to handle these cases, and use
//   the generic case to handle regular type-extensible Postgres syntax.
//   - thomas 1997-10-10

typename:
  simple_typename opt_array_bounds
  {
    if bounds := $2.int32s(); bounds != nil {
      var err error
      $$.val, err = arrayOf($1.typeReference(), bounds)
      if err != nil {
        return setErr(sqllex, err)
      }
    } else {
      $$.val = $1.typeReference()
    }
  }
  // SQL standard syntax, currently only one-dimensional
  // Undocumented but support for potential Postgres compat
| simple_typename ARRAY '[' ICONST ']' {
    /* SKIP DOC */
    var err error
    $$.val, err = arrayOf($1.typeReference(), nil)
    if err != nil {
      return setErr(sqllex, err)
    }
  }
| simple_typename ARRAY '[' ICONST ']' '[' error { return unimplementedWithIssue(sqllex, 32552) }
| simple_typename ARRAY {
    var err error
    $$.val, err = arrayOf($1.typeReference(), nil)
    if err != nil {
      return setErr(sqllex, err)
    }
  }

cast_target:
  typename
  {
    $$.val = $1.typeReference()
  }

opt_array_bounds:
  // TODO(justin): reintroduce multiple array bounds
  // opt_array_bounds '[' ']' { $$.val = append($1.int32s(), -1) }
  '[' ']' { $$.val = []int32{-1} }
| '[' ']' '[' error { return unimplementedWithIssue(sqllex, 32552) }
| '[' ICONST ']'
  {
    /* SKIP DOC */
    bound, err := $2.numVal().AsInt32()
    if err != nil {
      return setErr(sqllex, err)
    }
    $$.val = []int32{bound}
  }
| '[' ICONST ']' '[' error { return unimplementedWithIssue(sqllex, 32552) }
| /* EMPTY */ { $$.val = []int32(nil) }

// general_type_name is a variant of type_or_function_name but does not
// include some extra keywords (like FAMILY) which cause ambiguity with
// parsing of typenames in certain contexts.
general_type_name:
  type_function_name_no_crdb_extra

// complex_type_name mirrors the rule for complex_db_object_name, but uses
// general_type_name rather than db_object_name_component to avoid conflicts.
complex_type_name:
  general_type_name '.' unrestricted_name
  {
    aIdx := sqllex.(*lexer).NewAnnotation()
    res, err := tree.NewUnresolvedObjectName(2, [3]string{$3, $1}, aIdx)
    if err != nil { return setErr(sqllex, err) }
    $$.val = res
  }
| general_type_name '.' unrestricted_name '.' unrestricted_name
  {
    aIdx := sqllex.(*lexer).NewAnnotation()
    res, err := tree.NewUnresolvedObjectName(3, [3]string{$5, $3, $1}, aIdx)
    if err != nil { return setErr(sqllex, err) }
    $$.val = res
  }

simple_typename:
  general_type_name
  {
    /* FORCE DOC */
    // See https://www.postgresql.org/docs/9.1/static/datatype-character.html
    // Postgres supports a special character type named "char" (with the quotes)
    // that is a single-character column type. It's used by system tables.
    // Eventually this clause will be used to parse user-defined types as well,
    // since their names can be quoted.
    if $1 == "char" {
      $$.val = types.MakeQChar(0)
    } else if $1 == "serial" {
        switch sqllex.(*lexer).nakedIntType.Width() {
        case 32:
          $$.val = &types.Serial4Type
        default:
          $$.val = &types.Serial8Type
        }
    } else {
      // Check the the type is one of our "non-keyword" type names.
      // Otherwise, package it up as a type reference for later.
      var ok bool
      var err error
      $$.val, ok, _ = types.TypeForNonKeywordTypeName($1)
      if !ok {
        aIdx := sqllex.(*lexer).NewAnnotation()
        $$.val, err = tree.NewUnresolvedObjectName(1, [3]string{$1}, aIdx)
        if err != nil { return setErr(sqllex, err) }
      }
    }
  }
| '@' iconst32
  {
    id := $2.int32()
    $$.val = &tree.OIDTypeReference{OID: oid.Oid(id)}
  }
| complex_type_name
  {
    $$.val = $1.typeReference()
  }
| const_typename
| bit_with_length
| character_with_length
| interval_type
| POINT error { return unimplementedWithIssueDetail(sqllex, 21286, "point") } // needed or else it generates a syntax error.
| POLYGON error { return unimplementedWithIssueDetail(sqllex, 21286, "polygon") } // needed or else it generates a syntax error.

geo_shape_type:
  POINT { $$.val = geopb.ShapeType_Point }
| POINTM error { return unimplementedWithIssueDetail(sqllex, 53091, "XYM_type") } // { $$.val = geopb.ShapeType_PointM }
| POINTZ error { return unimplementedWithIssueDetail(sqllex, 53091, "XYZ_type") } // { $$.val = geopb.ShapeType_PointZ }
| POINTZM error { return unimplementedWithIssueDetail(sqllex, 53091, "XYZM_type") } // { $$.val = geopb.ShapeType_PointZM }
| LINESTRING { $$.val = geopb.ShapeType_LineString }
| LINESTRINGM error { return unimplementedWithIssueDetail(sqllex, 53091, "XYM_type") } // { $$.val = geopb.ShapeType_LineStringM }
| LINESTRINGZ error { return unimplementedWithIssueDetail(sqllex, 53091, "XYZ_type") } // { $$.val = geopb.ShapeType_LineStringZ }
| LINESTRINGZM error { return unimplementedWithIssueDetail(sqllex, 53091, "XYZM_type") } // { $$.val = geopb.ShapeType_LineStringZM }
| POLYGON { $$.val = geopb.ShapeType_Polygon }
| POLYGONM error { return unimplementedWithIssueDetail(sqllex, 53091, "XYM_type") } // { $$.val = geopb.ShapeType_PolygonM }
| POLYGONZ error { return unimplementedWithIssueDetail(sqllex, 53091, "XYZ_type") } // { $$.val = geopb.ShapeType_PolygonZ }
| POLYGONZM error { return unimplementedWithIssueDetail(sqllex, 53091, "XYZM_type") } // { $$.val = geopb.ShapeType_PolygonZM }
| MULTIPOINT { $$.val = geopb.ShapeType_MultiPoint }
| MULTIPOINTM error { return unimplementedWithIssueDetail(sqllex, 53091, "XYM_type") } // { $$.val = geopb.ShapeType_MultiPointM }
| MULTIPOINTZ error { return unimplementedWithIssueDetail(sqllex, 53091, "XYZ_type") } // { $$.val = geopb.ShapeType_MultiPointZ }
| MULTIPOINTZM error { return unimplementedWithIssueDetail(sqllex, 53091, "XYZM_type") } // { $$.val = geopb.ShapeType_MultiPointZM }
| MULTILINESTRING { $$.val = geopb.ShapeType_MultiLineString }
| MULTILINESTRINGM error { return unimplementedWithIssueDetail(sqllex, 53091, "XYM_type") } // { $$.val = geopb.ShapeType_MultiLineStringM }
| MULTILINESTRINGZ error { return unimplementedWithIssueDetail(sqllex, 53091, "XYZ_type") } // { $$.val = geopb.ShapeType_MultiLineStringZ }
| MULTILINESTRINGZM error { return unimplementedWithIssueDetail(sqllex, 53091, "XYZM_type") } // { $$.val = geopb.ShapeType_MultiLineStringZM }
| MULTIPOLYGON { $$.val = geopb.ShapeType_MultiPolygon }
| MULTIPOLYGONM error { return unimplementedWithIssueDetail(sqllex, 53091, "XYM_type") } // { $$.val = geopb.ShapeType_MultiPolygonM }
| MULTIPOLYGONZ error { return unimplementedWithIssueDetail(sqllex, 53091, "XYZ_type") } // { $$.val = geopb.ShapeType_MultiPolygonZ }
| MULTIPOLYGONZM error { return unimplementedWithIssueDetail(sqllex, 53091, "XYZM_type") } // { $$.val = geopb.ShapeType_MultiPolygonZM }
| GEOMETRYCOLLECTION { $$.val = geopb.ShapeType_GeometryCollection }
| GEOMETRYCOLLECTIONM error { return unimplementedWithIssueDetail(sqllex, 53091, "XYM_type") } // { $$.val = geopb.ShapeType_GeometryCollectionM }
| GEOMETRYCOLLECTIONZ error { return unimplementedWithIssueDetail(sqllex, 53091, "XYZ_type") } // { $$.val = geopb.ShapeType_GeometryCollectionZ }
| GEOMETRYCOLLECTIONZM error { return unimplementedWithIssueDetail(sqllex, 53091, "XYZM_type") } // { $$.val = geopb.ShapeType_GeometryCollectionZM }
| GEOMETRY { $$.val = geopb.ShapeType_Geometry }
| GEOMETRYM error { return unimplementedWithIssueDetail(sqllex, 53091, "XYM_type") } // { $$.val = geopb.ShapeType_GeometryM }
| GEOMETRYZ error { return unimplementedWithIssueDetail(sqllex, 53091, "XYZ_type") } // { $$.val = geopb.ShapeType_GeometryZ }
| GEOMETRYZM error { return unimplementedWithIssueDetail(sqllex, 53091, "XYZM_type") } // { $$.val = geopb.ShapeType_GeometryZM }

const_geo:
  GEOGRAPHY { $$.val = types.Geography }
| GEOMETRY  { $$.val = types.Geometry }
| BOX2D     { $$.val = types.Box2D }
| GEOMETRY '(' geo_shape_type ')'
  {
    $$.val = types.MakeGeometry($3.geoShapeType(), 0)
  }
| GEOGRAPHY '(' geo_shape_type ')'
  {
    $$.val = types.MakeGeography($3.geoShapeType(), 0)
  }
| GEOMETRY '(' geo_shape_type ',' signed_iconst ')'
  {
    val, err := $5.numVal().AsInt32()
    if err != nil {
      return setErr(sqllex, err)
    }
    $$.val = types.MakeGeometry($3.geoShapeType(), geopb.SRID(val))
  }
| GEOGRAPHY '(' geo_shape_type ',' signed_iconst ')'
  {
    val, err := $5.numVal().AsInt32()
    if err != nil {
      return setErr(sqllex, err)
    }
    $$.val = types.MakeGeography($3.geoShapeType(), geopb.SRID(val))
  }

// We have a separate const_typename to allow defaulting fixed-length types
// such as CHAR() and BIT() to an unspecified length. SQL9x requires that these
// default to a length of one, but this makes no sense for constructs like CHAR
// 'hi' and BIT '0101', where there is an obvious better choice to make. Note
// that interval_type is not included here since it must be pushed up higher
// in the rules to accommodate the postfix options (e.g. INTERVAL '1'
// YEAR). Likewise, we have to handle the generic-type-name case in
// a_expr_const to avoid premature reduce/reduce conflicts against function
// names.
const_typename:
  numeric
| bit_without_length
| character_without_length
| const_datetime
| const_geo

opt_numeric_modifiers:
  '(' iconst32 ')'
  {
    dec, err := newDecimal($2.int32(), 0)
    if err != nil {
      return setErr(sqllex, err)
    }
    $$.val = dec
  }
| '(' iconst32 ',' iconst32 ')'
  {
    dec, err := newDecimal($2.int32(), $4.int32())
    if err != nil {
      return setErr(sqllex, err)
    }
    $$.val = dec
  }
| /* EMPTY */
  {
    $$.val = nil
  }

// SQL numeric data types
numeric:
  INT
  {
    $$.val = types.Int4
  }
| INTEGER
  {
    $$.val = types.Int4
  }
| SMALLINT
  {
    $$.val = types.Int2
  }
| BIGINT
  {
    $$.val = types.Int
  }
| REAL
  {
    $$.val = types.Float4
  }
| FLOAT opt_float
  {
    $$.val = $2.colType()
  }
| DOUBLE PRECISION
  {
    $$.val = types.Float
  }
| DECIMAL opt_numeric_modifiers
  {
    typ := $2.colType()
    if typ == nil {
      typ = types.Decimal
    }
    $$.val = typ
  }
| DEC opt_numeric_modifiers
  {
    typ := $2.colType()
    if typ == nil {
      typ = types.Decimal
    }
    $$.val = typ
  }
| NUMERIC opt_numeric_modifiers
  {
    typ := $2.colType()
    if typ == nil {
      typ = types.Decimal
    }
    $$.val = typ
  }
| BOOLEAN
  {
    $$.val = types.Bool
  }

opt_float:
  '(' ICONST ')'
  {
    nv := $2.numVal()
    prec, err := nv.AsInt64()
    if err != nil {
      return setErr(sqllex, err)
    }
    typ, err := newFloat(prec)
    if err != nil {
      return setErr(sqllex, err)
    }
    $$.val = typ
  }
| /* EMPTY */
  {
    $$.val = types.Float
  }

bit_with_length:
  BIT opt_varying '(' iconst32 ')'
  {
    bit, err := newBitType($4.int32(), $2.bool())
    if err != nil { return setErr(sqllex, err) }
    $$.val = bit
  }
| VARBIT '(' iconst32 ')'
  {
    bit, err := newBitType($3.int32(), true)
    if err != nil { return setErr(sqllex, err) }
    $$.val = bit
  }

bit_without_length:
  BIT
  {
    $$.val = types.MakeBit(1)
  }
| BIT VARYING
  {
    $$.val = types.VarBit
  }
| VARBIT
  {
    $$.val = types.VarBit
  }

character_with_length:
  character_base '(' iconst32 ')'
  {
    colTyp := *$1.colType()
    n := $3.int32()
    if n == 0 {
      sqllex.Error(fmt.Sprintf("length for type %s must be at least 1", colTyp.SQLString()))
      return 1
    }
    $$.val = types.MakeScalar(types.StringFamily, colTyp.Oid(), colTyp.Precision(), n, colTyp.Locale())
  }

character_without_length:
  character_base
  {
    $$.val = $1.colType()
  }

character_base:
  char_aliases
  {
    $$.val = types.MakeChar(1)
  }
| char_aliases VARYING
  {
    $$.val = types.VarChar
  }
| VARCHAR
  {
    $$.val = types.VarChar
  }
| STRING
  {
    $$.val = types.String
  }

char_aliases:
  CHAR
| CHARACTER

opt_varying:
  VARYING     { $$.val = true }
| /* EMPTY */ { $$.val = false }

// SQL date/time types
const_datetime:
  DATE
  {
    $$.val = types.Date
  }
| TIME opt_timezone
  {
    if $2.bool() {
      $$.val = types.TimeTZ
    } else {
      $$.val = types.Time
    }
  }
| TIME '(' iconst32 ')' opt_timezone
  {
    prec := $3.int32()
    if prec < 0 || prec > 6 {
      sqllex.Error(fmt.Sprintf("precision %d out of range", prec))
      return 1
    }
    if $5.bool() {
      $$.val = types.MakeTimeTZ(prec)
    } else {
      $$.val = types.MakeTime(prec)
    }
  }
| TIMETZ                             { $$.val = types.TimeTZ }
| TIMETZ '(' iconst32 ')'
  {
    prec := $3.int32()
    if prec < 0 || prec > 6 {
      sqllex.Error(fmt.Sprintf("precision %d out of range", prec))
      return 1
    }
    $$.val = types.MakeTimeTZ(prec)
  }
| TIMESTAMP opt_timezone
  {
    if $2.bool() {
      $$.val = types.TimestampTZ
    } else {
      $$.val = types.Timestamp
    }
  }
| TIMESTAMP '(' iconst32 ')' opt_timezone
  {
    prec := $3.int32()
    if prec < 0 || prec > 6 {
      sqllex.Error(fmt.Sprintf("precision %d out of range", prec))
      return 1
    }
    if $5.bool() {
      $$.val = types.MakeTimestampTZ(prec)
    } else {
      $$.val = types.MakeTimestamp(prec)
    }
  }
| TIMESTAMPTZ
  {
    $$.val = types.TimestampTZ
  }
| TIMESTAMPTZ '(' iconst32 ')'
  {
    prec := $3.int32()
    if prec < 0 || prec > 6 {
      sqllex.Error(fmt.Sprintf("precision %d out of range", prec))
      return 1
    }
    $$.val = types.MakeTimestampTZ(prec)
  }

opt_timezone:
  WITH_LA TIME ZONE { $$.val = true; }
| WITHOUT TIME ZONE { $$.val = false; }
| /*EMPTY*/         { $$.val = false; }

interval_type:
  INTERVAL
  {
    $$.val = types.Interval
  }
| INTERVAL interval_qualifier
  {
    $$.val = types.MakeInterval($2.intervalTypeMetadata())
  }
| INTERVAL '(' iconst32 ')'
  {
    prec := $3.int32()
    if prec < 0 || prec > 6 {
      sqllex.Error(fmt.Sprintf("precision %d out of range", prec))
      return 1
    }
    $$.val = types.MakeInterval(types.IntervalTypeMetadata{Precision: prec, PrecisionIsSet: true})
  }

interval_qualifier:
  YEAR
  {
    $$.val = types.IntervalTypeMetadata{
      DurationField: types.IntervalDurationField{
        DurationType: types.IntervalDurationType_YEAR,
      },
    }
  }
| MONTH
  {
    $$.val = types.IntervalTypeMetadata{
      DurationField: types.IntervalDurationField{
        DurationType: types.IntervalDurationType_MONTH,
      },
    }
  }
| DAY
  {
    $$.val = types.IntervalTypeMetadata{
      DurationField: types.IntervalDurationField{
        DurationType: types.IntervalDurationType_DAY,
      },
    }
  }
| HOUR
  {
    $$.val = types.IntervalTypeMetadata{
      DurationField: types.IntervalDurationField{
        DurationType: types.IntervalDurationType_HOUR,
      },
    }
  }
| MINUTE
  {
    $$.val = types.IntervalTypeMetadata{
      DurationField: types.IntervalDurationField{
        DurationType: types.IntervalDurationType_MINUTE,
      },
    }
  }
| interval_second
  {
    $$.val = $1.intervalTypeMetadata()
  }
// Like Postgres, we ignore the left duration field. See explanation:
// https://www.postgresql.org/message-id/20110510040219.GD5617%40tornado.gateway.2wire.net
| YEAR TO MONTH
  {
    $$.val = types.IntervalTypeMetadata{
      DurationField: types.IntervalDurationField{
        FromDurationType: types.IntervalDurationType_YEAR,
        DurationType: types.IntervalDurationType_MONTH,
      },
    }
  }
| DAY TO HOUR
  {
    $$.val = types.IntervalTypeMetadata{
      DurationField: types.IntervalDurationField{
        FromDurationType: types.IntervalDurationType_DAY,
        DurationType: types.IntervalDurationType_HOUR,
      },
    }
  }
| DAY TO MINUTE
  {
    $$.val = types.IntervalTypeMetadata{
      DurationField: types.IntervalDurationField{
        FromDurationType: types.IntervalDurationType_DAY,
        DurationType: types.IntervalDurationType_MINUTE,
      },
    }
  }
| DAY TO interval_second
  {
    ret := $3.intervalTypeMetadata()
    ret.DurationField.FromDurationType = types.IntervalDurationType_DAY
    $$.val = ret
  }
| HOUR TO MINUTE
  {
    $$.val = types.IntervalTypeMetadata{
      DurationField: types.IntervalDurationField{
        FromDurationType: types.IntervalDurationType_HOUR,
        DurationType: types.IntervalDurationType_MINUTE,
      },
    }
  }
| HOUR TO interval_second
  {
    ret := $3.intervalTypeMetadata()
    ret.DurationField.FromDurationType = types.IntervalDurationType_HOUR
    $$.val = ret
  }
| MINUTE TO interval_second
  {
    $$.val = $3.intervalTypeMetadata()
    ret := $3.intervalTypeMetadata()
    ret.DurationField.FromDurationType = types.IntervalDurationType_MINUTE
    $$.val = ret
  }

opt_interval_qualifier:
  interval_qualifier
| /* EMPTY */
  {
    $$.val = nil
  }

interval_second:
  SECOND
  {
    $$.val = types.IntervalTypeMetadata{
      DurationField: types.IntervalDurationField{
        DurationType: types.IntervalDurationType_SECOND,
      },
    }
  }
| SECOND '(' iconst32 ')'
  {
    prec := $3.int32()
    if prec < 0 || prec > 6 {
      sqllex.Error(fmt.Sprintf("precision %d out of range", prec))
      return 1
    }
    $$.val = types.IntervalTypeMetadata{
      DurationField: types.IntervalDurationField{
        DurationType: types.IntervalDurationType_SECOND,
      },
      PrecisionIsSet: true,
      Precision: prec,
    }
  }

// General expressions. This is the heart of the expression syntax.
//
// We have two expression types: a_expr is the unrestricted kind, and b_expr is
// a subset that must be used in some places to avoid shift/reduce conflicts.
// For example, we can't do BETWEEN as "BETWEEN a_expr AND a_expr" because that
// use of AND conflicts with AND as a boolean operator. So, b_expr is used in
// BETWEEN and we remove boolean keywords from b_expr.
//
// Note that '(' a_expr ')' is a b_expr, so an unrestricted expression can
// always be used by surrounding it with parens.
//
// c_expr is all the productions that are common to a_expr and b_expr; it's
// factored out just to eliminate redundant coding.
//
// Be careful of productions involving more than one terminal token. By
// default, bison will assign such productions the precedence of their last
// terminal, but in nearly all cases you want it to be the precedence of the
// first terminal instead; otherwise you will not get the behavior you expect!
// So we use %prec annotations freely to set precedences.
a_expr:
  c_expr
| a_expr TYPECAST cast_target
  {
    $$.val = &tree.CastExpr{Expr: $1.expr(), Type: $3.typeReference(), SyntaxMode: tree.CastShort}
  }
| a_expr TYPEANNOTATE typename
  {
    $$.val = &tree.AnnotateTypeExpr{Expr: $1.expr(), Type: $3.typeReference(), SyntaxMode: tree.AnnotateShort}
  }
| a_expr COLLATE collation_name
  {
    $$.val = &tree.CollateExpr{Expr: $1.expr(), Locale: $3.unresolvedObjectName().UnquotedString()}
  }
| a_expr AT TIME ZONE a_expr %prec AT
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction("timezone"), Exprs: tree.Exprs{$5.expr(), $1.expr()}}
  }
  // These operators must be called out explicitly in order to make use of
  // bison's automatic operator-precedence handling. All other operator names
  // are handled by the generic productions using "OP", below; and all those
  // operators will have the same precedence.
  //
  // If you add more explicitly-known operators, be sure to add them also to
  // b_expr and to the math_op list below.
| '+' a_expr %prec UMINUS
  {
    // Unary plus is a no-op. Desugar immediately.
    $$.val = $2.expr()
  }
| '-' a_expr %prec UMINUS
  {
    $$.val = unaryNegation($2.expr())
  }
| '~' a_expr %prec UMINUS
  {
    $$.val = &tree.UnaryExpr{Operator: tree.UnaryComplement, Expr: $2.expr()}
  }
| SQRT a_expr
  {
    $$.val = &tree.UnaryExpr{Operator: tree.UnarySqrt, Expr: $2.expr()}
  }
| CBRT a_expr
  {
    $$.val = &tree.UnaryExpr{Operator: tree.UnaryCbrt, Expr: $2.expr()}
  }
| '@' a_expr
  {
    $$.val = &tree.UnaryExpr{Operator: tree.UnaryAbsolute, Expr: $2.expr()}
  }
| a_expr '+' a_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.Plus, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr '-' a_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.Minus, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr '*' a_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.Mult, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr '/' a_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.Div, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr FLOORDIV a_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.FloorDiv, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr '%' a_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.Mod, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr '^' a_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.Pow, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr '#' a_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.Bitxor, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr '&' a_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.Bitand, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr '|' a_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.Bitor, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr '<' a_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.LT, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr '>' a_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.GT, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr '?' a_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.JSONExists, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr JSON_SOME_EXISTS a_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.JSONSomeExists, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr JSON_ALL_EXISTS a_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.JSONAllExists, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr CONTAINS a_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.Contains, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr CONTAINED_BY a_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.ContainedBy, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr '=' a_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.EQ, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr CONCAT a_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.Concat, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr LSHIFT a_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.LShift, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr RSHIFT a_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.RShift, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr FETCHVAL a_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.JSONFetchVal, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr FETCHTEXT a_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.JSONFetchText, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr FETCHVAL_PATH a_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.JSONFetchValPath, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr FETCHTEXT_PATH a_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.JSONFetchTextPath, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr REMOVE_PATH a_expr
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction("json_remove_path"), Exprs: tree.Exprs{$1.expr(), $3.expr()}}
  }
| a_expr INET_CONTAINED_BY_OR_EQUALS a_expr
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction("inet_contained_by_or_equals"), Exprs: tree.Exprs{$1.expr(), $3.expr()}}
  }
| a_expr AND_AND a_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.Overlaps, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr INET_CONTAINS_OR_EQUALS a_expr
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction("inet_contains_or_equals"), Exprs: tree.Exprs{$1.expr(), $3.expr()}}
  }
| a_expr LESS_EQUALS a_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.LE, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr GREATER_EQUALS a_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.GE, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr NOT_EQUALS a_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.NE, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr AND a_expr
  {
    $$.val = &tree.AndExpr{Left: $1.expr(), Right: $3.expr()}
  }
| a_expr OR a_expr
  {
    $$.val = &tree.OrExpr{Left: $1.expr(), Right: $3.expr()}
  }
| NOT a_expr
  {
    $$.val = &tree.NotExpr{Expr: $2.expr()}
  }
| NOT_LA a_expr %prec NOT
  {
    $$.val = &tree.NotExpr{Expr: $2.expr()}
  }
| a_expr LIKE a_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.Like, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr LIKE a_expr ESCAPE a_expr %prec ESCAPE
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction("like_escape"), Exprs: tree.Exprs{$1.expr(), $3.expr(), $5.expr()}}
  }
| a_expr NOT_LA LIKE a_expr %prec NOT_LA
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.NotLike, Left: $1.expr(), Right: $4.expr()}
  }
| a_expr NOT_LA LIKE a_expr ESCAPE a_expr %prec ESCAPE
 {
   $$.val = &tree.FuncExpr{Func: tree.WrapFunction("not_like_escape"), Exprs: tree.Exprs{$1.expr(), $4.expr(), $6.expr()}}
 }
| a_expr ILIKE a_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.ILike, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr ILIKE a_expr ESCAPE a_expr %prec ESCAPE
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction("ilike_escape"), Exprs: tree.Exprs{$1.expr(), $3.expr(), $5.expr()}}
  }
| a_expr NOT_LA ILIKE a_expr %prec NOT_LA
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.NotILike, Left: $1.expr(), Right: $4.expr()}
  }
| a_expr NOT_LA ILIKE a_expr ESCAPE a_expr %prec ESCAPE
 {
   $$.val = &tree.FuncExpr{Func: tree.WrapFunction("not_ilike_escape"), Exprs: tree.Exprs{$1.expr(), $4.expr(), $6.expr()}}
 }
| a_expr SIMILAR TO a_expr %prec SIMILAR
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.SimilarTo, Left: $1.expr(), Right: $4.expr()}
  }
| a_expr SIMILAR TO a_expr ESCAPE a_expr %prec ESCAPE
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction("similar_to_escape"), Exprs: tree.Exprs{$1.expr(), $4.expr(), $6.expr()}}
  }
| a_expr NOT_LA SIMILAR TO a_expr %prec NOT_LA
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.NotSimilarTo, Left: $1.expr(), Right: $5.expr()}
  }
| a_expr NOT_LA SIMILAR TO a_expr ESCAPE a_expr %prec ESCAPE
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction("not_similar_to_escape"), Exprs: tree.Exprs{$1.expr(), $5.expr(), $7.expr()}}
  }
| a_expr '~' a_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.RegMatch, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr NOT_REGMATCH a_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.NotRegMatch, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr REGIMATCH a_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.RegIMatch, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr NOT_REGIMATCH a_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.NotRegIMatch, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr TEXTSEARCHMATCH a_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.TextSearchMatch, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr IS NAN %prec IS
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.EQ, Left: $1.expr(), Right: tree.NewStrVal("NaN")}
  }
| a_expr IS NOT NAN %prec IS
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.NE, Left: $1.expr(), Right: tree.NewStrVal("NaN")}
  }
| a_expr IS NULL %prec IS
  {
    $$.val = &tree.IsNullExpr{Expr: $1.expr()}
  }
| a_expr ISNULL %prec IS
  {
    $$.val = &tree.IsNullExpr{Expr: $1.expr()}
  }
| a_expr IS NOT NULL %prec IS
  {
    $$.val = &tree.IsNotNullExpr{Expr: $1.expr()}
  }
| a_expr NOTNULL %prec IS
  {
    $$.val = &tree.IsNotNullExpr{Expr: $1.expr()}
  }
| row OVERLAPS row { return unimplemented(sqllex, "overlaps") }
| a_expr IS TRUE %prec IS
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.IsNotDistinctFrom, Left: $1.expr(), Right: tree.MakeDBool(true)}
  }
| a_expr IS NOT TRUE %prec IS
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.IsDistinctFrom, Left: $1.expr(), Right: tree.MakeDBool(true)}
  }
| a_expr IS FALSE %prec IS
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.IsNotDistinctFrom, Left: $1.expr(), Right: tree.MakeDBool(false)}
  }
| a_expr IS NOT FALSE %prec IS
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.IsDistinctFrom, Left: $1.expr(), Right: tree.MakeDBool(false)}
  }
| a_expr IS UNKNOWN %prec IS
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.IsNotDistinctFrom, Left: $1.expr(), Right: tree.DNull}
  }
| a_expr IS NOT UNKNOWN %prec IS
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.IsDistinctFrom, Left: $1.expr(), Right: tree.DNull}
  }
| a_expr IS DISTINCT FROM a_expr %prec IS
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.IsDistinctFrom, Left: $1.expr(), Right: $5.expr()}
  }
| a_expr IS NOT DISTINCT FROM a_expr %prec IS
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.IsNotDistinctFrom, Left: $1.expr(), Right: $6.expr()}
  }
| a_expr IS OF '(' type_list ')' %prec IS
  {
    $$.val = &tree.IsOfTypeExpr{Expr: $1.expr(), Types: $5.typeReferences()}
  }
| a_expr IS NOT OF '(' type_list ')' %prec IS
  {
    $$.val = &tree.IsOfTypeExpr{Not: true, Expr: $1.expr(), Types: $6.typeReferences()}
  }
| a_expr BETWEEN opt_asymmetric b_expr AND a_expr %prec BETWEEN
  {
    $$.val = &tree.RangeCond{Left: $1.expr(), From: $4.expr(), To: $6.expr()}
  }
| a_expr NOT_LA BETWEEN opt_asymmetric b_expr AND a_expr %prec NOT_LA
  {
    $$.val = &tree.RangeCond{Not: true, Left: $1.expr(), From: $5.expr(), To: $7.expr()}
  }
| a_expr BETWEEN SYMMETRIC b_expr AND a_expr %prec BETWEEN
  {
    $$.val = &tree.RangeCond{Symmetric: true, Left: $1.expr(), From: $4.expr(), To: $6.expr()}
  }
| a_expr NOT_LA BETWEEN SYMMETRIC b_expr AND a_expr %prec NOT_LA
  {
    $$.val = &tree.RangeCond{Not: true, Symmetric: true, Left: $1.expr(), From: $5.expr(), To: $7.expr()}
  }
| a_expr IN in_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.In, Left: $1.expr(), Right: $3.expr()}
  }
| a_expr NOT_LA IN in_expr %prec NOT_LA
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.NotIn, Left: $1.expr(), Right: $4.expr()}
  }
| a_expr subquery_op sub_type a_expr %prec CONCAT
  {
    op := $3.cmpOp()
    subOp := $2.op()
    subOpCmp, ok := subOp.(tree.ComparisonOperator)
    if !ok {
      sqllex.Error(fmt.Sprintf("%s %s <array> is invalid because %q is not a boolean operator",
        subOp, op, subOp))
      return 1
    }
    $$.val = &tree.ComparisonExpr{
      Operator: op,
      SubOperator: subOpCmp,
      Left: $1.expr(),
      Right: $4.expr(),
    }
  }
| DEFAULT
  {
    $$.val = tree.DefaultVal{}
  }
// The UNIQUE predicate is a standard SQL feature but not yet implemented
// in PostgreSQL (as of 10.5).
| UNIQUE '(' error { return unimplemented(sqllex, "UNIQUE predicate") }
// Below here we special-case the OPERATOR(...) syntax only for operators in the pg_catalog schema. 
// This is to support particular psql commands that require it.
| a_expr OPERATOR '(' schema_name '.' '+' ')' a_expr
  {
    $$.val = &tree.BinaryExpr{Schema: tree.Name($4), Operator: tree.Plus, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' '-' ')' a_expr
  {
    $$.val = &tree.BinaryExpr{Schema: tree.Name($4), Operator: tree.Minus, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' '*' ')' a_expr
  {
    $$.val = &tree.BinaryExpr{Schema: tree.Name($4), Operator: tree.Mult, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' '/' ')' a_expr
  {
    $$.val = &tree.BinaryExpr{Schema: tree.Name($4), Operator: tree.Div, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' FLOORDIV ')' a_expr
  {
    $$.val = &tree.BinaryExpr{Schema: tree.Name($4), Operator: tree.FloorDiv, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' '%' ')' a_expr
  {
    $$.val = &tree.BinaryExpr{Schema: tree.Name($4), Operator: tree.Mod, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' '^' ')' a_expr
  {
    $$.val = &tree.BinaryExpr{Schema: tree.Name($4), Operator: tree.Pow, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' '#' ')' a_expr
  {
    $$.val = &tree.BinaryExpr{Schema: tree.Name($4), Operator: tree.Bitxor, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' '&' ')' a_expr
  {
    $$.val = &tree.BinaryExpr{Schema: tree.Name($4), Operator: tree.Bitand, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' '|' ')' a_expr
  {
    $$.val = &tree.BinaryExpr{Schema: tree.Name($4), Operator: tree.Bitor, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' '<' ')' a_expr
  {
    $$.val = &tree.ComparisonExpr{Schema: tree.Name($4), Operator: tree.LT, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' '>' ')' a_expr
  {
    $$.val = &tree.ComparisonExpr{Schema: tree.Name($4), Operator: tree.GT, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' '?' ')' a_expr
  {
    $$.val = &tree.ComparisonExpr{Schema: tree.Name($4), Operator: tree.JSONExists, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' JSON_SOME_EXISTS ')' a_expr
  {
    $$.val = &tree.ComparisonExpr{Schema: tree.Name($4), Operator: tree.JSONSomeExists, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' JSON_ALL_EXISTS ')' a_expr
  {
    $$.val = &tree.ComparisonExpr{Schema: tree.Name($4), Operator: tree.JSONAllExists, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' '=' ')' a_expr
  {
    $$.val = &tree.ComparisonExpr{Schema: tree.Name($4), Operator: tree.EQ, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' CONCAT ')' a_expr
  {
    $$.val = &tree.BinaryExpr{Schema: tree.Name($4), Operator: tree.Concat, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' LSHIFT ')' a_expr
  {
    $$.val = &tree.BinaryExpr{Schema: tree.Name($4), Operator: tree.LShift, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' RSHIFT ')' a_expr
  {
    $$.val = &tree.BinaryExpr{Schema: tree.Name($4), Operator: tree.RShift, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' FETCHVAL ')' a_expr
  {
    $$.val = &tree.BinaryExpr{Schema: tree.Name($4), Operator: tree.JSONFetchVal, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' FETCHTEXT ')' a_expr
  {
    $$.val = &tree.BinaryExpr{Schema: tree.Name($4), Operator: tree.JSONFetchText, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' FETCHVAL_PATH ')' a_expr
  {
    $$.val = &tree.BinaryExpr{Schema: tree.Name($4), Operator: tree.JSONFetchValPath, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' FETCHTEXT_PATH ')' a_expr
  {
    $$.val = &tree.BinaryExpr{Schema: tree.Name($4), Operator: tree.JSONFetchTextPath, Left: $1.expr(), Right: $8.expr()}
  }
| a_expr OPERATOR '(' schema_name '.' REMOVE_PATH ')' a_expr
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunctionSchema("json_remove_path", $4), Exprs: tree.Exprs{$1.expr(), $8.expr()}}
  }
 | a_expr OPERATOR '(' schema_name '.' INET_CONTAINED_BY_OR_EQUALS ')' a_expr
   {
     $$.val = &tree.FuncExpr{Func: tree.WrapFunctionSchema("inet_contained_by_or_equals", $4), Exprs: tree.Exprs{$1.expr(), $8.expr()}}
   }
 | a_expr OPERATOR '(' schema_name '.' AND_AND ')' a_expr
   {
     $$.val = &tree.ComparisonExpr{Schema: tree.Name($4), Operator: tree.Overlaps, Left: $1.expr(), Right: $8.expr()}
   }
 | a_expr OPERATOR '(' schema_name '.' INET_CONTAINS_OR_EQUALS ')' a_expr
   {
     $$.val = &tree.FuncExpr{Func: tree.WrapFunctionSchema("inet_contains_or_equals", $4), Exprs: tree.Exprs{$1.expr(), $8.expr()}}
   }
 | a_expr OPERATOR '(' schema_name '.' LESS_EQUALS ')' a_expr
   {
     $$.val = &tree.ComparisonExpr{Schema: tree.Name($4), Operator: tree.LE, Left: $1.expr(), Right: $8.expr()}
   }
 | a_expr OPERATOR '(' schema_name '.' GREATER_EQUALS ')' a_expr
   {
     $$.val = &tree.ComparisonExpr{Schema: tree.Name($4), Operator: tree.GE, Left: $1.expr(), Right: $8.expr()}
   }
 | a_expr OPERATOR '(' schema_name '.' NOT_EQUALS ')' a_expr
   {
     $$.val = &tree.ComparisonExpr{Schema: tree.Name($4), Operator: tree.NE, Left: $1.expr(), Right: $8.expr()}
   }
 | a_expr OPERATOR '(' schema_name '.' '~' ')' a_expr
   {
     $$.val = &tree.ComparisonExpr{Schema: tree.Name($4), Operator: tree.RegMatch, Left: $1.expr(), Right: $8.expr()}
   }
 | a_expr OPERATOR '(' schema_name '.' NOT_REGMATCH ')' a_expr
   {
     $$.val = &tree.ComparisonExpr{Schema: tree.Name($4), Operator: tree.NotRegMatch, Left: $1.expr(), Right: $8.expr()}
   }
 | a_expr OPERATOR '(' schema_name '.' REGIMATCH ')' a_expr
   {
     $$.val = &tree.ComparisonExpr{Schema: tree.Name($4), Operator: tree.RegIMatch, Left: $1.expr(), Right: $8.expr()}
   }
 | a_expr OPERATOR '(' schema_name '.' NOT_REGIMATCH ')' a_expr
   {
     $$.val = &tree.ComparisonExpr{Schema: tree.Name($4), Operator: tree.NotRegIMatch, Left: $1.expr(), Right: $8.expr()}
   }
 | a_expr OPERATOR '(' schema_name '.' TEXTSEARCHMATCH ')' a_expr
   {
     $$.val = &tree.ComparisonExpr{Schema: tree.Name($4), Operator: tree.TextSearchMatch, Left: $1.expr(), Right: $8.expr()}
   }

// Restricted expressions
//
// b_expr is a subset of the complete expression syntax defined by a_expr.
//
// Presently, AND, NOT, IS, and IN are the a_expr keywords that would cause
// trouble in the places where b_expr is used. For simplicity, we just
// eliminate all the boolean-keyword-operator productions from b_expr.
b_expr:
  c_expr
| b_expr TYPECAST cast_target
  {
    $$.val = &tree.CastExpr{Expr: $1.expr(), Type: $3.typeReference(), SyntaxMode: tree.CastShort}
  }
| b_expr TYPEANNOTATE typename
  {
    $$.val = &tree.AnnotateTypeExpr{Expr: $1.expr(), Type: $3.typeReference(), SyntaxMode: tree.AnnotateShort}
  }
| '+' b_expr %prec UMINUS
  {
    $$.val = $2.expr()
  }
| '-' b_expr %prec UMINUS
  {
    $$.val = unaryNegation($2.expr())
  }
| '~' b_expr %prec UMINUS
  {
    $$.val = &tree.UnaryExpr{Operator: tree.UnaryComplement, Expr: $2.expr()}
  }
| b_expr '+' b_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.Plus, Left: $1.expr(), Right: $3.expr()}
  }
| b_expr '-' b_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.Minus, Left: $1.expr(), Right: $3.expr()}
  }
| b_expr '*' b_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.Mult, Left: $1.expr(), Right: $3.expr()}
  }
| b_expr '/' b_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.Div, Left: $1.expr(), Right: $3.expr()}
  }
| b_expr FLOORDIV b_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.FloorDiv, Left: $1.expr(), Right: $3.expr()}
  }
| b_expr '%' b_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.Mod, Left: $1.expr(), Right: $3.expr()}
  }
| b_expr '^' b_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.Pow, Left: $1.expr(), Right: $3.expr()}
  }
| b_expr '#' b_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.Bitxor, Left: $1.expr(), Right: $3.expr()}
  }
| b_expr '&' b_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.Bitand, Left: $1.expr(), Right: $3.expr()}
  }
| b_expr '|' b_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.Bitor, Left: $1.expr(), Right: $3.expr()}
  }
| b_expr '<' b_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.LT, Left: $1.expr(), Right: $3.expr()}
  }
| b_expr '>' b_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.GT, Left: $1.expr(), Right: $3.expr()}
  }
| b_expr '=' b_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.EQ, Left: $1.expr(), Right: $3.expr()}
  }
| b_expr CONCAT b_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.Concat, Left: $1.expr(), Right: $3.expr()}
  }
| b_expr LSHIFT b_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.LShift, Left: $1.expr(), Right: $3.expr()}
  }
| b_expr RSHIFT b_expr
  {
    $$.val = &tree.BinaryExpr{Operator: tree.RShift, Left: $1.expr(), Right: $3.expr()}
  }
| b_expr LESS_EQUALS b_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.LE, Left: $1.expr(), Right: $3.expr()}
  }
| b_expr GREATER_EQUALS b_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.GE, Left: $1.expr(), Right: $3.expr()}
  }
| b_expr NOT_EQUALS b_expr
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.NE, Left: $1.expr(), Right: $3.expr()}
  }
| b_expr IS DISTINCT FROM b_expr %prec IS
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.IsDistinctFrom, Left: $1.expr(), Right: $5.expr()}
  }
| b_expr IS NOT DISTINCT FROM b_expr %prec IS
  {
    $$.val = &tree.ComparisonExpr{Operator: tree.IsNotDistinctFrom, Left: $1.expr(), Right: $6.expr()}
  }
| b_expr IS OF '(' type_list ')' %prec IS
  {
    $$.val = &tree.IsOfTypeExpr{Expr: $1.expr(), Types: $5.typeReferences()}
  }
| b_expr IS NOT OF '(' type_list ')' %prec IS
  {
    $$.val = &tree.IsOfTypeExpr{Not: true, Expr: $1.expr(), Types: $6.typeReferences()}
  }

// Productions that can be used in both a_expr and b_expr.
//
// Note: productions that refer recursively to a_expr or b_expr mostly cannot
// appear here. However, it's OK to refer to a_exprs that occur inside
// parentheses, such as function arguments; that cannot introduce ambiguity to
// the b_expr syntax.
//
c_expr:
  d_expr
| d_expr array_subscripts
  {
    $$.val = &tree.IndirectionExpr{
      Expr: $1.expr(),
      Indirection: $2.arraySubscripts(),
    }
  }
| case_expr
| EXISTS select_with_parens
  {
    $$.val = &tree.Subquery{Select: $2.selectStmt(), Exists: true}
  }

// Productions that can be followed by a postfix operator.
//
// Currently we support array indexing (see c_expr above).
//
// TODO(knz/jordan): this is the rule that can be extended to support
// composite types (#27792) with e.g.:
//
//     | '(' a_expr ')' field_access_ops
//
//     [...]
//
//     // field_access_ops supports the notations:
//     // - .a
//     // - .a[123]
//     // - .a.b[123][5456].c.d
//     // NOT [123] directly, this is handled in c_expr above.
//
//     field_access_ops:
//       field_access_op
//     | field_access_op other_subscripts
//
//     field_access_op:
//       '.' name
//     other_subscripts:
//       other_subscript
//     | other_subscripts other_subscript
//     other_subscript:
//        field_access_op
//     |  array_subscripts

d_expr:
  ICONST
  {
    $$.val = $1.numVal()
  }
| FCONST
  {
    $$.val = $1.numVal()
  }
| SCONST
  {
    $$.val = tree.NewStrVal($1)
  }
| BCONST
  {
    $$.val = tree.NewBytesStrVal($1)
  }
| BITCONST
  {
    d, err := tree.ParseDBitArray($1)
    if err != nil { return setErr(sqllex, err) }
    $$.val = d
  }
| func_name '(' expr_list opt_sort_clause ')' SCONST { return unimplemented(sqllex, $1.unresolvedName().String() + "(...) SCONST") }
| typed_literal
  {
    $$.val = $1.expr()
  }
| interval_value
  {
    $$.val = $1.expr()
  }
| TRUE
  {
    $$.val = tree.MakeDBool(true)
  }
| FALSE
  {
    $$.val = tree.MakeDBool(false)
  }
| NULL
  {
    $$.val = tree.DNull
  }
| column_path_with_star
  {
    $$.val = tree.Expr($1.unresolvedName())
  }
| PLACEHOLDER
  {
    p := $1.placeholder()
    sqllex.(*lexer).UpdateNumPlaceholders(p)
    $$.val = p
  }
// TODO(knz/jordan): extend this for compound types. See explanation above.
| '(' a_expr ')' '.' '*'
  {
    $$.val = &tree.TupleStar{Expr: $2.expr()}
  }
| '(' a_expr ')' '.' unrestricted_name
  {
    $$.val = &tree.ColumnAccessExpr{Expr: $2.expr(), ColName: $5 }
  }
| '(' a_expr ')' '.' '@' ICONST
  {
    idx, err := $6.numVal().AsInt32()
    if err != nil || idx <= 0 { return setErr(sqllex, err) }
    $$.val = &tree.ColumnAccessExpr{Expr: $2.expr(), ByIndex: true, ColIndex: int(idx-1)}
  }
| '(' a_expr ')'
  {
    $$.val = &tree.ParenExpr{Expr: $2.expr()}
  }
| func_expr
| select_with_parens %prec UMINUS
  {
    $$.val = &tree.Subquery{Select: $1.selectStmt()}
  }
| labeled_row
  {
    $$.val = $1.tuple()
  }
| ARRAY select_with_parens %prec UMINUS
  {
    $$.val = &tree.ArrayFlatten{Subquery: &tree.Subquery{Select: $2.selectStmt()}}
  }
| ARRAY row
  {
    $$.val = &tree.Array{Exprs: $2.tuple().Exprs}
  }
| ARRAY array_expr
  {
    $$.val = $2.expr()
  }
| GROUPING '(' expr_list ')' { return unimplemented(sqllex, "d_expr grouping") }

func_application:
  func_name '(' ')'
  {
    $$.val = &tree.FuncExpr{Func: $1.resolvableFuncRefFromName()}
  }
| func_name '(' expr_list opt_sort_clause ')'
  {
    $$.val = &tree.FuncExpr{Func: $1.resolvableFuncRefFromName(), Exprs: $3.exprs(), OrderBy: $4.orderBy(), AggType: tree.GeneralAgg}
  }
| func_name '(' VARIADIC a_expr opt_sort_clause ')' { return unimplemented(sqllex, "variadic") }
| func_name '(' expr_list ',' VARIADIC a_expr opt_sort_clause ')' { return unimplemented(sqllex, "variadic") }
| func_name '(' ALL expr_list opt_sort_clause ')'
  {
    $$.val = &tree.FuncExpr{Func: $1.resolvableFuncRefFromName(), Type: tree.AllFuncType, Exprs: $4.exprs(), OrderBy: $5.orderBy(), AggType: tree.GeneralAgg}
  }
// TODO(ridwanmsharif): Once DISTINCT is supported by window aggregates,
// allow ordering to be specified below.
| func_name '(' DISTINCT expr_list ')'
  {
    $$.val = &tree.FuncExpr{Func: $1.resolvableFuncRefFromName(), Type: tree.DistinctFuncType, Exprs: $4.exprs()}
  }
| func_name '(' '*' ')'
  {
    $$.val = &tree.FuncExpr{Func: $1.resolvableFuncRefFromName(), Exprs: tree.Exprs{tree.StarExpr()}}
  }
| func_name '(' error { return helpWithFunction(sqllex, $1.resolvableFuncRefFromName()) }

// typed_literal represents expressions like INT '4', or generally <TYPE> SCONST.
// This rule handles both the case of qualified and non-qualified typenames.
typed_literal:
  // The key here is that none of the keywords in the func_name_no_crdb_extra
  // production can overlap with the type rules in const_typename, otherwise
  // we will have conflicts between this rule and the one below.
  func_name_no_crdb_extra SCONST
  {
    name := $1.unresolvedName()
    if name.NumParts == 1 {
      typName := name.Parts[0]
      /* FORCE DOC */
      // See https://www.postgresql.org/docs/9.1/static/datatype-character.html
      // Postgres supports a special character type named "char" (with the quotes)
      // that is a single-character column type. It's used by system tables.
      // Eventually this clause will be used to parse user-defined types as well,
      // since their names can be quoted.
      if typName == "char" {
        $$.val = &tree.CastExpr{Expr: tree.NewStrVal($2), Type: types.MakeQChar(0), SyntaxMode: tree.CastPrepend}
      } else if typName == "serial" {
        switch sqllex.(*lexer).nakedIntType.Width() {
        case 32:
          $$.val = &tree.CastExpr{Expr: tree.NewStrVal($2), Type: &types.Serial4Type, SyntaxMode: tree.CastPrepend}
        default:
          $$.val = &tree.CastExpr{Expr: tree.NewStrVal($2), Type: &types.Serial8Type, SyntaxMode: tree.CastPrepend}
        }
      } else {
        // Check the the type is one of our "non-keyword" type names.
        // Otherwise, package it up as a type reference for later.
        // However, if the type name is one of our known unsupported
        // types, return an unimplemented error message.
        var typ tree.ResolvableTypeReference
        var ok bool
        var err error
        var unimp int
        typ, ok, unimp = types.TypeForNonKeywordTypeName(typName)
        if !ok {
          switch unimp {
            case 0:
              // In this case, we don't think this type is one of our
              // known unsupported types, so make a type reference for it.
              aIdx := sqllex.(*lexer).NewAnnotation()
              typ, err = name.ToUnresolvedObjectName(aIdx)
              if err != nil { return setErr(sqllex, err) }
            case -1:
              return unimplemented(sqllex, "type name " + typName)
            default:
              return unimplementedWithIssueDetail(sqllex, unimp, typName)
          }
        }
      $$.val = &tree.CastExpr{Expr: tree.NewStrVal($2), Type: typ, SyntaxMode: tree.CastPrepend}
      }
    } else {
      aIdx := sqllex.(*lexer).NewAnnotation()
      res, err := name.ToUnresolvedObjectName(aIdx)
      if err != nil { return setErr(sqllex, err) }
      $$.val = &tree.CastExpr{Expr: tree.NewStrVal($2), Type: res, SyntaxMode: tree.CastPrepend}
    }
  }
| const_typename SCONST
  {
    $$.val = &tree.CastExpr{Expr: tree.NewStrVal($2), Type: $1.colType(), SyntaxMode: tree.CastPrepend}
  }
| bit_with_length SCONST
  {
    $$.val = &tree.CastExpr{Expr: tree.NewStrVal($2), Type: $1.colType(), SyntaxMode: tree.CastPrepend}
  }
| character_with_length SCONST
  {
    $$.val = &tree.CastExpr{Expr: tree.NewStrVal($2), Type: $1.colType(), SyntaxMode: tree.CastPrepend}
  }

// func_expr and its cousin func_expr_windowless are split out from c_expr just
// so that we have classifications for "everything that is a function call or
// looks like one". This isn't very important, but it saves us having to
// document which variants are legal in places like "FROM function()" or the
// backwards-compatible functional-index syntax for CREATE INDEX. (Note that
// many of the special SQL functions wouldn't actually make any sense as
// functional index entries, but we ignore that consideration here.)
func_expr:
  func_application within_group_clause filter_clause over_clause
  {
    f := $1.expr().(*tree.FuncExpr)
    w := $2.expr().(*tree.FuncExpr)
    if w.AggType != 0 {
      f.AggType = w.AggType
      f.OrderBy = w.OrderBy
    }
    f.Filter = $3.expr()
    f.WindowDef = $4.windowDef()
    $$.val = f
  }
| func_expr_common_subexpr
  {
    $$.val = $1.expr()
  }

// As func_expr but does not accept WINDOW functions directly (but they can
// still be contained in arguments for functions etc). Use this when window
// expressions are not allowed, where needed to disambiguate the grammar
// (e.g. in CREATE INDEX).
func_expr_windowless:
  func_application { $$.val = $1.expr() }
| func_expr_common_subexpr { $$.val = $1.expr() }

// Special expressions that are considered to be functions.
func_expr_common_subexpr:
  COLLATION FOR '(' a_expr ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction("pg_collation_for"), Exprs: tree.Exprs{$4.expr()}}
  }
| CURRENT_DATE
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1)}
  }
| CURRENT_SCHEMA
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1)}
  }
// Special identifier current_catalog is equivalent to current_database().
// https://www.postgresql.org/docs/10/static/functions-info.html
| CURRENT_CATALOG
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction("current_database")}
  }
| CURRENT_TIMESTAMP
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1)}
  }
| CURRENT_TIME
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1)}
  }
| LOCALTIMESTAMP
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1)}
  }
| LOCALTIME
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1)}
  }
| CURRENT_USER
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1)}
  }
// Special identifier current_role is equivalent to current_user.
// https://www.postgresql.org/docs/10/static/functions-info.html
| CURRENT_ROLE
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction("current_user")}
  }
| SESSION_USER
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction("current_user")}
  }
| USER
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction("current_user")}
  }
| CAST '(' a_expr AS cast_target ')'
  {
    $$.val = &tree.CastExpr{Expr: $3.expr(), Type: $5.typeReference(), SyntaxMode: tree.CastExplicit}
  }
| ANNOTATE_TYPE '(' a_expr ',' typename ')'
  {
    $$.val = &tree.AnnotateTypeExpr{Expr: $3.expr(), Type: $5.typeReference(), SyntaxMode: tree.AnnotateExplicit}
  }
| IF '(' a_expr ',' a_expr ',' a_expr ')'
  {
    $$.val = &tree.IfExpr{Cond: $3.expr(), True: $5.expr(), Else: $7.expr()}
  }
| IFERROR '(' a_expr ',' a_expr ',' a_expr ')'
  {
    $$.val = &tree.IfErrExpr{Cond: $3.expr(), Else: $5.expr(), ErrCode: $7.expr()}
  }
| IFERROR '(' a_expr ',' a_expr ')'
  {
    $$.val = &tree.IfErrExpr{Cond: $3.expr(), Else: $5.expr()}
  }
| ISERROR '(' a_expr ')'
  {
    $$.val = &tree.IfErrExpr{Cond: $3.expr()}
  }
| ISERROR '(' a_expr ',' a_expr ')'
  {
    $$.val = &tree.IfErrExpr{Cond: $3.expr(), ErrCode: $5.expr()}
  }
| NULLIF '(' a_expr ',' a_expr ')'
  {
    $$.val = &tree.NullIfExpr{Expr1: $3.expr(), Expr2: $5.expr()}
  }
| IFNULL '(' a_expr ',' a_expr ')'
  {
    $$.val = &tree.CoalesceExpr{Name: "IFNULL", Exprs: tree.Exprs{$3.expr(), $5.expr()}}
  }
| COALESCE '(' expr_list ')'
  {
    $$.val = &tree.CoalesceExpr{Name: "COALESCE", Exprs: $3.exprs()}
  }
| special_function

special_function:
  CURRENT_DATE '(' ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1)}
  }
| CURRENT_DATE '(' error { return helpWithFunctionByName(sqllex, $1) }
| CURRENT_SCHEMA '(' ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1)}
  }
| CURRENT_SCHEMA '(' error { return helpWithFunctionByName(sqllex, $1) }
| CURRENT_TIMESTAMP '(' ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1)}
  }
| CURRENT_TIMESTAMP '(' a_expr ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1), Exprs: tree.Exprs{$3.expr()}}
  }
| CURRENT_TIMESTAMP '(' error { return helpWithFunctionByName(sqllex, $1) }
| CURRENT_TIME '(' ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1)}
  }
| CURRENT_TIME '(' a_expr ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1), Exprs: tree.Exprs{$3.expr()}}
  }
| CURRENT_TIME '(' error { return helpWithFunctionByName(sqllex, $1) }
| LOCALTIMESTAMP '(' ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1)}
  }
| LOCALTIMESTAMP '(' a_expr ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1), Exprs: tree.Exprs{$3.expr()}}
  }
| LOCALTIMESTAMP '(' error { return helpWithFunctionByName(sqllex, $1) }
| LOCALTIME '(' ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1)}
  }
| LOCALTIME '(' a_expr ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1), Exprs: tree.Exprs{$3.expr()}}
  }
| LOCALTIME '(' error { return helpWithFunctionByName(sqllex, $1) }
| CURRENT_USER '(' ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1)}
  }
| CURRENT_USER '(' error { return helpWithFunctionByName(sqllex, $1) }
| EXTRACT '(' extract_list ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1), Exprs: $3.exprs()}
  }
| EXTRACT '(' error { return helpWithFunctionByName(sqllex, $1) }
| EXTRACT_DURATION '(' extract_list ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1), Exprs: $3.exprs()}
  }
| EXTRACT_DURATION '(' error { return helpWithFunctionByName(sqllex, $1) }
| OVERLAY '(' overlay_list ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1), Exprs: $3.exprs()}
  }
| OVERLAY '(' error { return helpWithFunctionByName(sqllex, $1) }
| POSITION '(' position_list ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction("strpos"), Exprs: $3.exprs()}
  }
| SUBSTRING '(' substr_list ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1), Exprs: $3.exprs()}
  }
| SUBSTRING '(' error { return helpWithFunctionByName(sqllex, $1) }
| TREAT '(' a_expr AS typename ')' { return unimplemented(sqllex, "treat") }
| TRIM '(' BOTH trim_list ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction("btrim"), Exprs: $4.exprs()}
  }
| TRIM '(' LEADING trim_list ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction("ltrim"), Exprs: $4.exprs()}
  }
| TRIM '(' TRAILING trim_list ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction("rtrim"), Exprs: $4.exprs()}
  }
| TRIM '(' trim_list ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction("btrim"), Exprs: $3.exprs()}
  }
| GREATEST '(' expr_list ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1), Exprs: $3.exprs()}
  }
| GREATEST '(' error { return helpWithFunctionByName(sqllex, $1) }
| LEAST '(' expr_list ')'
  {
    $$.val = &tree.FuncExpr{Func: tree.WrapFunction($1), Exprs: $3.exprs()}
  }
| LEAST '(' error { return helpWithFunctionByName(sqllex, $1) }


// Aggregate decoration clauses
within_group_clause:
  WITHIN GROUP '(' single_sort_clause ')'
  {
    $$.val = &tree.FuncExpr{OrderBy: $4.orderBy(), AggType: tree.OrderedSetAgg}
  }
| /* EMPTY */
  {
    $$.val = &tree.FuncExpr{}
  }

filter_clause:
  FILTER '(' WHERE a_expr ')'
  {
    $$.val = $4.expr()
  }
| /* EMPTY */
  {
    $$.val = tree.Expr(nil)
  }

// Window Definitions
window_clause:
  WINDOW window_definition_list
  {
    $$.val = $2.window()
  }
| /* EMPTY */
  {
    $$.val = tree.Window(nil)
  }

window_definition_list:
  window_definition
  {
    $$.val = tree.Window{$1.windowDef()}
  }
| window_definition_list ',' window_definition
  {
    $$.val = append($1.window(), $3.windowDef())
  }

window_definition:
  window_name AS window_specification
  {
    n := $3.windowDef()
    n.Name = tree.Name($1)
    $$.val = n
  }

over_clause:
  OVER window_specification
  {
    $$.val = $2.windowDef()
  }
| OVER window_name
  {
    $$.val = &tree.WindowDef{Name: tree.Name($2)}
  }
| /* EMPTY */
  {
    $$.val = (*tree.WindowDef)(nil)
  }

window_specification:
  '(' opt_existing_window_name opt_partition_clause
    opt_sort_clause opt_frame_clause ')'
  {
    $$.val = &tree.WindowDef{
      RefName: tree.Name($2),
      Partitions: $3.exprs(),
      OrderBy: $4.orderBy(),
      Frame: $5.windowFrame(),
    }
  }

// If we see PARTITION, RANGE, ROWS, or GROUPS as the first token after the '('
// of a window_specification, we want the assumption to be that there is no
// existing_window_name; but those keywords are unreserved and so could be
// names. We fix this by making them have the same precedence as IDENT and
// giving the empty production here a slightly higher precedence, so that the
// shift/reduce conflict is resolved in favor of reducing the rule. These
// keywords are thus precluded from being an existing_window_name but are not
// reserved for any other purpose.
opt_existing_window_name:
  name
| /* EMPTY */ %prec CONCAT
  {
    $$ = ""
  }

opt_partition_clause:
  PARTITION BY expr_list
  {
    $$.val = $3.exprs()
  }
| /* EMPTY */
  {
    $$.val = tree.Exprs(nil)
  }

opt_frame_clause:
  RANGE frame_extent opt_frame_exclusion
  {
    $$.val = &tree.WindowFrame{
      Mode: tree.RANGE,
      Bounds: $2.windowFrameBounds(),
      Exclusion: $3.windowFrameExclusion(),
    }
  }
| ROWS frame_extent opt_frame_exclusion
  {
    $$.val = &tree.WindowFrame{
      Mode: tree.ROWS,
      Bounds: $2.windowFrameBounds(),
      Exclusion: $3.windowFrameExclusion(),
    }
  }
| GROUPS frame_extent opt_frame_exclusion
  {
    $$.val = &tree.WindowFrame{
      Mode: tree.GROUPS,
      Bounds: $2.windowFrameBounds(),
      Exclusion: $3.windowFrameExclusion(),
    }
  }
| /* EMPTY */
  {
    $$.val = (*tree.WindowFrame)(nil)
  }

frame_extent:
  frame_bound
  {
    startBound := $1.windowFrameBound()
    switch {
    case startBound.BoundType == tree.UnboundedFollowing:
      sqllex.Error("frame start cannot be UNBOUNDED FOLLOWING")
      return 1
    case startBound.BoundType == tree.OffsetFollowing:
      sqllex.Error("frame starting from following row cannot end with current row")
      return 1
    }
    $$.val = tree.WindowFrameBounds{StartBound: startBound}
  }
| BETWEEN frame_bound AND frame_bound
  {
    startBound := $2.windowFrameBound()
    endBound := $4.windowFrameBound()
    switch {
    case startBound.BoundType == tree.UnboundedFollowing:
      sqllex.Error("frame start cannot be UNBOUNDED FOLLOWING")
      return 1
    case endBound.BoundType == tree.UnboundedPreceding:
      sqllex.Error("frame end cannot be UNBOUNDED PRECEDING")
      return 1
    case startBound.BoundType == tree.CurrentRow && endBound.BoundType == tree.OffsetPreceding:
      sqllex.Error("frame starting from current row cannot have preceding rows")
      return 1
    case startBound.BoundType == tree.OffsetFollowing && endBound.BoundType == tree.OffsetPreceding:
      sqllex.Error("frame starting from following row cannot have preceding rows")
      return 1
    case startBound.BoundType == tree.OffsetFollowing && endBound.BoundType == tree.CurrentRow:
      sqllex.Error("frame starting from following row cannot have preceding rows")
      return 1
    }
    $$.val = tree.WindowFrameBounds{StartBound: startBound, EndBound: endBound}
  }

// This is used for both frame start and frame end, with output set up on the
// assumption it's frame start; the frame_extent productions must reject
// invalid cases.
frame_bound:
  UNBOUNDED PRECEDING
  {
    $$.val = &tree.WindowFrameBound{BoundType: tree.UnboundedPreceding}
  }
| UNBOUNDED FOLLOWING
  {
    $$.val = &tree.WindowFrameBound{BoundType: tree.UnboundedFollowing}
  }
| CURRENT ROW
  {
    $$.val = &tree.WindowFrameBound{BoundType: tree.CurrentRow}
  }
| a_expr PRECEDING
  {
    $$.val = &tree.WindowFrameBound{
      OffsetExpr: $1.expr(),
      BoundType: tree.OffsetPreceding,
    }
  }
| a_expr FOLLOWING
  {
    $$.val = &tree.WindowFrameBound{
      OffsetExpr: $1.expr(),
      BoundType: tree.OffsetFollowing,
    }
  }

opt_frame_exclusion:
  EXCLUDE CURRENT ROW
  {
    $$.val = tree.ExcludeCurrentRow
  }
| EXCLUDE GROUP
  {
    $$.val = tree.ExcludeGroup
  }
| EXCLUDE TIES
  {
    $$.val = tree.ExcludeTies
  }
| EXCLUDE NO OTHERS
  {
    // EXCLUDE NO OTHERS is equivalent to omitting the frame exclusion clause.
    $$.val = tree.NoExclusion
  }
| /* EMPTY */
  {
    $$.val = tree.NoExclusion
  }

// Supporting nonterminals for expressions.

// Explicit row production.
//
// SQL99 allows an optional ROW keyword, so we can now do single-element rows
// without conflicting with the parenthesized a_expr production. Without the
// ROW keyword, there must be more than one a_expr inside the parens.
row:
  ROW '(' opt_expr_list ')'
  {
    $$.val = &tree.Tuple{Exprs: $3.exprs(), Row: true}
  }
| expr_tuple_unambiguous
  {
    $$.val = $1.tuple()
  }

labeled_row:
  row
| '(' row AS name_list ')'
  {
    t := $2.tuple()
    labels := $4.nameList()
    t.Labels = make([]string, len(labels))
    for i, l := range labels {
      t.Labels[i] = string(l)
    }
    $$.val = t
  }

sub_type:
  ANY
  {
    $$.val = tree.Any
  }
| SOME
  {
    $$.val = tree.Some
  }
| ALL
  {
    $$.val = tree.All
  }

/* TODO: not all operators are included */
operator:
  subquery_op
| '~' { $$.val = tree.RegMatch }
| SQRT { $$.val = tree.UnarySqrt }
| CBRT { $$.val = tree.UnaryCbrt }
| '?' { $$.val = tree.JSONExists }
| JSON_SOME_EXISTS { $$.val = tree.JSONSomeExists }
| JSON_ALL_EXISTS { $$.val = tree.JSONAllExists }
| CONTAINS { $$.val = tree.Contains }
| CONTAINED_BY { $$.val = tree.ContainedBy }
| CONCAT { $$.val = tree.Concat }
| LSHIFT { $$.val = tree.LShift }
| RSHIFT { $$.val = tree.RShift }
| FETCHVAL { $$.val = tree.JSONFetchVal }
| FETCHTEXT { $$.val = tree.JSONFetchText }
| FETCHVAL_PATH { $$.val = tree.JSONFetchValPath }
| FETCHTEXT_PATH { $$.val = tree.JSONFetchTextPath }
| AND_AND { $$.val = tree.Overlaps }
| TEXTSEARCHMATCH { $$.val = tree.TextSearchMatch }

math_op:
  '+' { $$.val = tree.Plus  }
| '-' { $$.val = tree.Minus }
| '*' { $$.val = tree.Mult  }
| '/' { $$.val = tree.Div   }
| FLOORDIV { $$.val = tree.FloorDiv }
| '%' { $$.val = tree.Mod    }
| '&' { $$.val = tree.Bitand }
| '|' { $$.val = tree.Bitor  }
| '^' { $$.val = tree.Pow }
| '#' { $$.val = tree.Bitxor }
| '<' { $$.val = tree.LT }
| '>' { $$.val = tree.GT }
| '=' { $$.val = tree.EQ }
| LESS_EQUALS    { $$.val = tree.LE }
| GREATER_EQUALS { $$.val = tree.GE }
| NOT_EQUALS     { $$.val = tree.NE }
| '@' { $$.val = tree.UnaryAbsolute }

subquery_op:
  math_op
| LIKE         { $$.val = tree.Like     }
| NOT_LA LIKE  { $$.val = tree.NotLike  }
| ILIKE        { $$.val = tree.ILike    }
| NOT_LA ILIKE { $$.val = tree.NotILike }
  // cannot put SIMILAR TO here, because SIMILAR TO is a hack.
  // the regular expression is preprocessed by a function (similar_escape),
  // and the ~ operator for posix regular expressions is used.
  //        x SIMILAR TO y     ->    x ~ similar_escape(y)
  // this transformation is made on the fly by the parser upwards.
  // however the SubLink structure which handles any/some/all stuff
  // is not ready for such a thing.

// expr_tuple1_ambiguous is a tuple expression with at least one expression.
// The allowable syntax is:
// ( )         -- empty tuple.
// ( E )       -- just one value, this is potentially ambiguous with
//             -- grouping parentheses. The ambiguity is resolved
//             -- by only allowing expr_tuple1_ambiguous on the RHS
//             -- of a IN expression.
// ( E, E, E ) -- comma-separated values, no trailing comma allowed.
// ( E, )      -- just one value with a comma, makes the syntax unambiguous
//             -- with grouping parentheses. This is not usually produced
//             -- by SQL clients, but can be produced by pretty-printing
//             -- internally in CockroachDB.
expr_tuple1_ambiguous:
  '(' ')'
  {
    $$.val = &tree.Tuple{}
  }
| '(' tuple1_ambiguous_values ')'
  {
    $$.val = &tree.Tuple{Exprs: $2.exprs()}
  }

tuple1_ambiguous_values:
  a_expr
  {
    $$.val = tree.Exprs{$1.expr()}
  }
| a_expr ','
  {
    $$.val = tree.Exprs{$1.expr()}
  }
| a_expr ',' expr_list
  {
     $$.val = append(tree.Exprs{$1.expr()}, $3.exprs()...)
  }

// expr_tuple_unambiguous is a tuple expression with zero or more
// expressions. The allowable syntax is:
// ( )         -- zero values
// ( E, )      -- just one value. This is unambiguous with the (E) grouping syntax.
// ( E, E, E ) -- comma-separated values, more than 1.
expr_tuple_unambiguous:
  '(' ')'
  {
    $$.val = &tree.Tuple{}
  }
| '(' tuple1_unambiguous_values ')'
  {
    $$.val = &tree.Tuple{Exprs: $2.exprs()}
  }

tuple1_unambiguous_values:
  a_expr ','
  {
    $$.val = tree.Exprs{$1.expr()}
  }
| a_expr ',' expr_list
  {
     $$.val = append(tree.Exprs{$1.expr()}, $3.exprs()...)
  }

opt_expr_list:
  expr_list
| /* EMPTY */
  {
    $$.val = tree.Exprs(nil)
  }

expr_list:
  a_expr
  {
    $$.val = tree.Exprs{$1.expr()}
  }
| expr_list ',' a_expr
  {
    $$.val = append($1.exprs(), $3.expr())
  }

type_list:
  typename
  {
    $$.val = []tree.ResolvableTypeReference{$1.typeReference()}
  }
| type_list ',' typename
  {
    $$.val = append($1.typeReferences(), $3.typeReference())
  }

array_expr:
  '[' opt_expr_list ']'
  {
    $$.val = &tree.Array{Exprs: $2.exprs()}
  }
| '[' array_expr_list ']'
  {
    $$.val = &tree.Array{Exprs: $2.exprs()}
  }

array_expr_list:
  array_expr
  {
    $$.val = tree.Exprs{$1.expr()}
  }
| array_expr_list ',' array_expr
  {
    $$.val = append($1.exprs(), $3.expr())
  }

extract_list:
  extract_arg FROM a_expr
  {
    $$.val = tree.Exprs{tree.NewStrVal($1), $3.expr()}
  }
| expr_list
  {
    $$.val = $1.exprs()
  }

// TODO(vivek): Narrow down to just IDENT once the other
// terms are not keywords.
extract_arg:
  IDENT
| YEAR
| MONTH
| DAY
| HOUR
| MINUTE
| SECOND
| SCONST

// OVERLAY() arguments
// SQL99 defines the OVERLAY() function:
//   - overlay(text placing text from int for int)
//   - overlay(text placing text from int)
// and similarly for binary strings
overlay_list:
  a_expr overlay_placing substr_from substr_for
  {
    $$.val = tree.Exprs{$1.expr(), $2.expr(), $3.expr(), $4.expr()}
  }
| a_expr overlay_placing substr_from
  {
    $$.val = tree.Exprs{$1.expr(), $2.expr(), $3.expr()}
  }
| expr_list
  {
    $$.val = $1.exprs()
  }

overlay_placing:
  PLACING a_expr
  {
    $$.val = $2.expr()
  }

// position_list uses b_expr not a_expr to avoid conflict with general IN
position_list:
  b_expr IN b_expr
  {
    $$.val = tree.Exprs{$3.expr(), $1.expr()}
  }
| /* EMPTY */
  {
    $$.val = tree.Exprs(nil)
  }

// SUBSTRING() arguments
// SQL9x defines a specific syntax for arguments to SUBSTRING():
//   - substring(text from int for int)
//   - substring(text from int) get entire string from starting point "int"
//   - substring(text for int) get first "int" characters of string
//   - substring(text from pattern) get entire string matching pattern
//   - substring(text from pattern for escape) same with specified escape char
// We also want to support generic substring functions which accept
// the usual generic list of arguments. So we will accept both styles
// here, and convert the SQL9x style to the generic list for further
// processing. - thomas 2000-11-28
substr_list:
  a_expr substr_from substr_for
  {
    $$.val = tree.Exprs{$1.expr(), $2.expr(), $3.expr()}
  }
| a_expr substr_for substr_from
  {
    $$.val = tree.Exprs{$1.expr(), $3.expr(), $2.expr()}
  }
| a_expr substr_from
  {
    $$.val = tree.Exprs{$1.expr(), $2.expr()}
  }
| a_expr substr_for
  {
    $$.val = tree.Exprs{$1.expr(), tree.NewDInt(1), $2.expr()}
  }
| opt_expr_list
  {
    $$.val = $1.exprs()
  }

substr_from:
  FROM a_expr
  {
    $$.val = $2.expr()
  }

substr_for:
  FOR a_expr
  {
    $$.val = $2.expr()
  }

trim_list:
  a_expr FROM expr_list
  {
    $$.val = append($3.exprs(), $1.expr())
  }
| FROM expr_list
  {
    $$.val = $2.exprs()
  }
| expr_list
  {
    $$.val = $1.exprs()
  }

in_expr:
  select_with_parens
  {
    $$.val = &tree.Subquery{Select: $1.selectStmt()}
  }
| expr_tuple1_ambiguous

// Define SQL-style CASE clause.
// - Full specification
//      CASE WHEN a = b THEN c ... ELSE d END
// - Implicit argument
//      CASE a WHEN b THEN c ... ELSE d END
case_expr:
  CASE case_arg when_clause_list case_default END
  {
    $$.val = &tree.CaseExpr{Expr: $2.expr(), Whens: $3.whens(), Else: $4.expr()}
  }

when_clause_list:
  // There must be at least one
  when_clause
  {
    $$.val = []*tree.When{$1.when()}
  }
| when_clause_list when_clause
  {
    $$.val = append($1.whens(), $2.when())
  }

when_clause:
  WHEN a_expr THEN a_expr
  {
    $$.val = &tree.When{Cond: $2.expr(), Val: $4.expr()}
  }

case_default:
  ELSE a_expr
  {
    $$.val = $2.expr()
  }
| /* EMPTY */
  {
    $$.val = tree.Expr(nil)
  }

case_arg:
  a_expr
| /* EMPTY */
  {
    $$.val = tree.Expr(nil)
  }

array_subscript:
  '[' a_expr ']'
  {
    $$.val = &tree.ArraySubscript{Begin: $2.expr()}
  }
| '[' opt_slice_bound ':' opt_slice_bound ']'
  {
    $$.val = &tree.ArraySubscript{Begin: $2.expr(), End: $4.expr(), Slice: true}
  }

opt_slice_bound:
  a_expr
| /*EMPTY*/
  {
    $$.val = tree.Expr(nil)
  }

array_subscripts:
  array_subscript
  {
    $$.val = tree.ArraySubscripts{$1.arraySubscript()}
  }
| array_subscripts array_subscript
  {
    $$.val = append($1.arraySubscripts(), $2.arraySubscript())
  }

opt_asymmetric:
  ASYMMETRIC {}
| /* EMPTY */ {}

target_list:
  target_elem
  {
    $$.val = tree.SelectExprs{$1.selExpr()}
  }
| target_list ',' target_elem
  {
    $$.val = append($1.selExprs(), $3.selExpr())
  }

target_elem:
  a_expr AS target_name
  {
    $$.val = tree.SelectExpr{Expr: $1.expr(), As: tree.UnrestrictedName($3)}
  }
  // We support omitting AS only for column labels that aren't any known
  // keyword. There is an ambiguity against postfix operators: is "a ! b" an
  // infix expression, or a postfix expression and a column label?  We prefer
  // to resolve this as an infix expression, which we accomplish by assigning
  // IDENT a precedence higher than POSTFIXOP.
| a_expr IDENT
  {
    $$.val = tree.SelectExpr{Expr: $1.expr(), As: tree.UnrestrictedName($2)}
  }
| a_expr
  {
    $$.val = tree.SelectExpr{Expr: $1.expr()}
  }
| '*'
  {
    $$.val = tree.StarSelectExpr()
  }

// Names and constants.

table_index_name_list:
  table_index_name
  {
    $$.val = tree.TableIndexNames{$1.newTableIndexName()}
  }
| table_index_name_list ',' table_index_name
  {
    $$.val = append($1.newTableIndexNames(), $3.newTableIndexName())
  }

table_pattern_list:
  table_pattern
  {
    $$.val = tree.TablePatterns{$1.unresolvedName()}
  }
| table_pattern_list ',' table_pattern
  {
    $$.val = append($1.tablePatterns(), $3.unresolvedName())
  }

// An index can be specified in a few different ways:
//
//   - with explicit table name:
//       <table>@<index>
//       <schema>.<table>@<index>
//       <catalog/db>.<table>@<index>
//       <catalog/db>.<schema>.<table>@<index>
//
//   - without explicit table name:
//       <index>
//       <schema>.<index>
//       <catalog/db>.<index>
//       <catalog/db>.<schema>.<index>
table_index_name:
  table_name '@' index_name
  {
    name := $1.unresolvedObjectName().ToTableName()
    $$.val = tree.TableIndexName{
       Table: name,
       Index: tree.UnrestrictedName($3),
    }
  }
| standalone_index_name
  {
    // Treat it as a table name, then pluck out the ObjectName.
    name := $1.unresolvedObjectName().ToTableName()
    indexName := tree.UnrestrictedName(name.ObjectName)
    name.ObjectName = ""
    $$.val = tree.TableIndexName{
        Table: name,
        Index: indexName,
    }
  }

// table_pattern selects zero or more tables using a wildcard.
// Accepted patterns:
// - Patterns accepted by db_object_name
//   <table>
//   <schema>.<table>
//   <catalog/db>.<schema>.<table>
// - Wildcards:
//   <db/catalog>.<schema>.*
//   <schema>.*
//   *
table_pattern:
  simple_db_object_name
  {
     $$.val = $1.unresolvedObjectName().ToUnresolvedName()
  }
| complex_table_pattern

// complex_table_pattern is the part of table_pattern which recognizes
// every pattern not composed of a single identifier.
complex_table_pattern:
  complex_db_object_name
  {
     $$.val = $1.unresolvedObjectName().ToUnresolvedName()
  }
| db_object_name_component '.' unrestricted_name '.' '*'
  {
     $$.val = &tree.UnresolvedName{Star: true, NumParts: 3, Parts: tree.NameParts{"", $3, $1}}
  }
| db_object_name_component '.' '*'
  {
     $$.val = &tree.UnresolvedName{Star: true, NumParts: 2, Parts: tree.NameParts{"", $1}}
  }
| '*'
  {
     $$.val = &tree.UnresolvedName{Star: true, NumParts: 1}
  }

opt_name_list:
  /* Empty */
  {
    $$.val = tree.NameList(nil)
  }
| name_list
  {
    $$.val = $1.nameList()
  }

name_list:
  name
  {
    $$.val = tree.NameList{tree.Name($1)}
  }
| name_list ',' name
  {
    $$.val = append($1.nameList(), tree.Name($3))
  }
  
// Constants
numeric_only:
  signed_iconst
| signed_fconst

int_expr_list:
  signed_iconst
  {
    $$.val = tree.Exprs{$1.expr()}
  }
| int_expr_list ',' signed_iconst
  {
    $$.val = append($1.exprs(), $3.expr())
  }

signed_iconst:
  ICONST
| only_signed_iconst

only_signed_iconst:
  '+' ICONST
  {
    $$.val = $2.numVal()
  }
| '-' ICONST
  {
    n := $2.numVal()
    n.SetNegative()
    $$.val = n
  }

signed_fconst:
  FCONST
| only_signed_fconst

only_signed_fconst:
  '+' FCONST
  {
    $$.val = $2.numVal()
  }
| '-' FCONST
  {
    n := $2.numVal()
    n.SetNegative()
    $$.val = n
  }

// signed_iconst32 is a variant of signed_iconst which only accepts (signed) integer literals that fit in an int32.
signed_iconst32:
  signed_iconst
  {
    val, err := $1.numVal().AsInt32()
    if err != nil { return setErr(sqllex, err) }
    $$.val = val
  }

// iconst32 accepts only unsigned integer literals that fit in an int32.
iconst32:
  ICONST
  {
    val, err := $1.numVal().AsInt32()
    if err != nil { return setErr(sqllex, err) }
    $$.val = val
  }

// signed_iconst64 is a variant of signed_iconst which only accepts (signed) integer literals that fit in an int64.
// If you use signed_iconst, you have to call AsInt64(), which returns an error if the value is too big.
// This rule just doesn't match in that case.
signed_iconst64:
  signed_iconst
  {
    val, err := $1.numVal().AsInt64()
    if err != nil { return setErr(sqllex, err) }
    $$.val = val
  }

// iconst64 accepts only unsigned integer literals that fit in an int64.
iconst64:
  ICONST
  {
    val, err := $1.numVal().AsInt64()
    if err != nil { return setErr(sqllex, err) }
    $$.val = val
  }

interval_value:
  INTERVAL SCONST opt_interval_qualifier
  {
    var err error
    var d tree.Datum
    if $3.val == nil {
      d, err = tree.ParseDInterval($2)
    } else {
      d, err = tree.ParseDIntervalWithTypeMetadata($2, $3.intervalTypeMetadata())
    }
    if err != nil { return setErr(sqllex, err) }
    $$.val = d
  }
| INTERVAL '(' iconst32 ')' SCONST
  {
    prec := $3.int32()
    if prec < 0 || prec > 6 {
      sqllex.Error(fmt.Sprintf("precision %d out of range", prec))
      return 1
    }
    d, err := tree.ParseDIntervalWithTypeMetadata($5, types.IntervalTypeMetadata{
      Precision: prec,
      PrecisionIsSet: true,
    })
    if err != nil { return setErr(sqllex, err) }
    $$.val = d
  }

// Name classification hierarchy.
//
// IDENT is the lexeme returned by the lexer for identifiers that match no
// known keyword. In most cases, we can accept certain keywords as names, not
// only IDENTs. We prefer to accept as many such keywords as possible to
// minimize the impact of "reserved words" on programmers. So, we divide names
// into several possible classes. The classification is chosen in part to make
// keywords acceptable as names wherever possible.

// Names specific to syntactic positions.
//
// The non-terminals "name", "unrestricted_name", "non_reserved_word",
// "unreserved_keyword", "non_reserved_word_or_sconst" etc. defined
// below are low-level, structural constructs.
//
// They are separate only because having them all as one rule would
// make the rest of the grammar ambiguous. However, because they are
// separate the question is then raised throughout the rest of the
// grammar: which of the name non-terminals should one use when
// defining a grammar rule?  Is an index a "name" or
// "unrestricted_name"? A partition? What about an index option?
//
// To make the decision easier, this section of the grammar creates
// meaningful, purpose-specific aliases to the non-terminals. These
// both make it easier to decide "which one should I use in this
// context" and also improves the readability of
// automatically-generated syntax diagrams.

// Note: newlines between non-terminals matter to the doc generator.

collation_name:        db_object_name

index_name:            unrestricted_name

opt_index_name:        opt_name

target_name:           unrestricted_name

constraint_name:       name

database_name:         name

column_name:           name

table_alias_name:      name

statistics_name:       name

window_name:           name

view_name:             table_name

trigger_name:          name

type_name:             db_object_name

sequence_name:         db_object_name

schema_name:           name

opt_schema_name:       opt_name

table_name:            db_object_name

standalone_index_name: db_object_name

explain_option_name:
  non_reserved_word
| ANALYZE
| VERBOSE

cursor_name:           name

tablespace_name:       name

partition_name:        name

routine_name:         db_object_name

aggregate_name:       db_object_name

// Names for column references.
// Accepted patterns:
// <colname>
// <table>.<colname>
// <schema>.<table>.<colname>
// <catalog/db>.<schema>.<table>.<colname>
//
// Note: the rule for accessing compound types, if those are ever
// supported, is not to be handled here. The syntax `a.b.c.d....y.z`
// in `select a.b.c.d from t` *always* designates a column `z` in a
// table `y`, regardless of the meaning of what's before.
column_path:
  name
  {
      $$.val = &tree.UnresolvedName{NumParts:1, Parts: tree.NameParts{$1}}
  }
| prefixed_column_path

prefixed_column_path:
  db_object_name_component '.' unrestricted_name
  {
      $$.val = &tree.UnresolvedName{NumParts:2, Parts: tree.NameParts{$3,$1}}
  }
| db_object_name_component '.' unrestricted_name '.' unrestricted_name
  {
      $$.val = &tree.UnresolvedName{NumParts:3, Parts: tree.NameParts{$5,$3,$1}}
  }
| db_object_name_component '.' unrestricted_name '.' unrestricted_name '.' unrestricted_name
  {
      $$.val = &tree.UnresolvedName{NumParts:4, Parts: tree.NameParts{$7,$5,$3,$1}}
  }

// Names for column references and wildcards.
// Accepted patterns:
// - those from column_path
// - <table>.*
// - <schema>.<table>.*
// - <catalog/db>.<schema>.<table>.*
// The single unqualified star is handled separately by target_elem.
column_path_with_star:
  column_path
| db_object_name_component '.' unrestricted_name '.' unrestricted_name '.' '*'
  {
    $$.val = &tree.UnresolvedName{Star:true, NumParts:4, Parts: tree.NameParts{"",$5,$3,$1}}
  }
| db_object_name_component '.' unrestricted_name '.' '*'
  {
    $$.val = &tree.UnresolvedName{Star:true, NumParts:3, Parts: tree.NameParts{"",$3,$1}}
  }
| db_object_name_component '.' '*'
  {
    $$.val = &tree.UnresolvedName{Star:true, NumParts:2, Parts: tree.NameParts{"",$1}}
  }

// Names for functions.
// The production for a qualified func_name has to exactly match the production
// for a column_path, because we cannot tell which we are parsing until
// we see what comes after it ('(' or SCONST for a func_name, anything else for
// a name).
// However we cannot use column_path directly, because for a single function name
// we allow more possible tokens than a simple column name.
func_name:
  type_function_name
  {
    $$.val = &tree.UnresolvedName{NumParts:1, Parts: tree.NameParts{$1}}
  }
| prefixed_column_path

// func_name_no_crdb_extra is the same rule as func_name, but does not
// contain some CRDB specific keywords like FAMILY.
func_name_no_crdb_extra:
  type_function_name_no_crdb_extra
  {
    $$.val = &tree.UnresolvedName{NumParts:1, Parts: tree.NameParts{$1}}
  }
| prefixed_column_path

// Names for database objects (tables, sequences, views, stored functions).
// Accepted patterns:
// <table>
// <schema>.<table>
// <catalog/db>.<schema>.<table>
db_object_name:
  simple_db_object_name
| complex_db_object_name

// Version of db_object_name that does not contain any CRDB specific keywords.
db_object_name_no_keywords:
  simple_db_object_name_no_keywords
| complex_db_object_name_no_keywords

// simple_db_object_name is the part of db_object_name that recognizes
// simple identifiers.
simple_db_object_name:
  db_object_name_component
  {
    aIdx := sqllex.(*lexer).NewAnnotation()
    res, err := tree.NewUnresolvedObjectName(1, [3]string{$1}, aIdx)
    if err != nil { return setErr(sqllex, err) }
    $$.val = res
  }
  
// simple_db_object_name_no_keywords is the part of db_object_name_no_keywords that recognizes
// simple identifiers.
simple_db_object_name_no_keywords:
  simple_ident
  {
    aIdx := sqllex.(*lexer).NewAnnotation()
    res, err := tree.NewUnresolvedObjectName(1, [3]string{$1}, aIdx)
    if err != nil { return setErr(sqllex, err) }
    $$.val = res
  }

// complex_db_object_name is the part of db_object_name that recognizes
// composite names (not simple identifiers).
// It is split away from db_object_name in order to enable the definition
// of table_pattern.
complex_db_object_name:
  db_object_name_component '.' unrestricted_name
  {
    aIdx := sqllex.(*lexer).NewAnnotation()
    res, err := tree.NewUnresolvedObjectName(2, [3]string{$3, $1}, aIdx)
    if err != nil { return setErr(sqllex, err) }
    $$.val = res
  }
| db_object_name_component '.' unrestricted_name '.' unrestricted_name
  {
    aIdx := sqllex.(*lexer).NewAnnotation()
    res, err := tree.NewUnresolvedObjectName(3, [3]string{$5, $3, $1}, aIdx)
    if err != nil { return setErr(sqllex, err) }
    $$.val = res
  }
 
complex_db_object_name_no_keywords:
  simple_ident '.' simple_ident
  {
    aIdx := sqllex.(*lexer).NewAnnotation()
    res, err := tree.NewUnresolvedObjectName(2, [3]string{$3, $1}, aIdx)
    if err != nil { return setErr(sqllex, err) }
    $$.val = res
  }
| simple_ident '.' simple_ident '.' simple_ident
  {
    aIdx := sqllex.(*lexer).NewAnnotation()
    res, err := tree.NewUnresolvedObjectName(3, [3]string{$5, $3, $1}, aIdx)
    if err != nil { return setErr(sqllex, err) }
    $$.val = res
  }
  
// simple_ident is a more restricted version of restricted_name that disallows most keywords
simple_ident:
  IDENT
| PUBLIC // PUBLIC is a keyword, but its use as the default schema makes it nice to include here

// DB object name component -- this cannot not include any reserved
// keyword because of ambiguity after FROM, but we've been too lax
// with reserved keywords and made INDEX and FAMILY reserved, so we're
// trying to gain them back here.
db_object_name_component:
  name
| cockroachdb_extra_reserved_keyword

// General name --- names that can be column, table, etc names.
name:
  IDENT
| unreserved_keyword
| col_name_keyword

opt_name:
  name
| /* EMPTY */
  {
    $$ = ""
  }

opt_name_parens:
  '(' name ')'
  {
    $$ = $2
  }
| /* EMPTY */
  {
    $$ = ""
  }

// Structural, low-level names

// Non-reserved word and also string literal constants.
non_reserved_word_or_sconst:
  non_reserved_word
| SCONST

// Type/function identifier --- names that can be type or function names.
type_function_name:
  IDENT
| unreserved_keyword
| type_func_name_keyword

// Type/function identifier without CRDB extra reserved keywords.
type_function_name_no_crdb_extra:
  IDENT
| unreserved_keyword
| type_func_name_no_crdb_extra_keyword

// Any not-fully-reserved word --- these names can be, eg, variable names.
non_reserved_word:
  IDENT
| unreserved_keyword
| col_name_keyword
| type_func_name_keyword

// Unrestricted name --- allowable names when there is no ambiguity with even
// reserved keywords, like in "AS" clauses. This presently includes *all*
// Postgres keywords.
unrestricted_name:
  IDENT
| unreserved_keyword
| col_name_keyword
| type_func_name_keyword
| reserved_keyword

// Keyword category lists. Generally, every keyword present in the Postgres
// grammar should appear in exactly one of these lists.
//
// Put a new keyword into the first list that it can go into without causing
// shift or reduce conflicts. The earlier lists define "less reserved"
// categories of keywords.
//
// "Unreserved" keywords --- available for use as any kind of name.
unreserved_keyword:
  ABORT
| ACCESS
| ACTION
| ADD
| ADMIN
| AFTER
| AGGREGATE
| ALIGNMENT
| ALLOW_CONNECTIONS
| ALTER
| ALWAYS
| AT
| ATOMIC
| ATTACH
| ATTRIBUTE
| AUTO
| AUTOMATIC
| BACKUP
| BACKUPS
| BASETYPE
| BEFORE
| BEGIN
| BINARY
| BUCKET_COUNT
| BUFFER_USAGE_LIMIT
| BUNDLE
| BY
| BYPASSRLS
| CACHE
| CALL
| CALLED
| CANCEL
| CANCELQUERY
| CANONICAL
| CASCADE
| CASCADED
| CATEGORY
| CHAIN
| CHANGEFEED
| CHECK_OPTION
| CLASS
| CLOSE
| CLUSTER
| COLLATABLE
| COLLATION_VERSION
| COLUMNS
| COMBINEFUNC
| COMMENT
| COMMENTS
| COMMIT
| COMMITTED
| COMPACT
| COMPLETE
| COMPRESSION
| CONFIGURATION
| CONFIGURATIONS
| CONFIGURE
| CONFLICT
| CONNECTION
| CONSTRAINTS
| CONTROLCHANGEFEED
| CONTROLJOB
| CONVERSION
| CONVERT
| COPY
| COST
| CREATEDB
| CREATELOGIN
| CREATEROLE
| CSV
| CUBE
| CURRENT
| CYCLE
| DATA
| DATABASE
| DATABASES
| DAY
| DEALLOCATE
| DECLARE
| DEFAULTS
| DEFERRED
| DEFINER
| DELETE
| DELIMITER
| DEPENDS
| DESERIALFUNC
| DESTINATION
| DETACH
| DETACHED
| DICTIONARY
| DISABLE
| DISABLE_PAGE_SKIPPING
| DISCARD
| DOMAIN
| DOUBLE
| DROP
| EACH
| ENABLE
| ENCODING
| ENCRYPTED
| ENCRYPTION_PASSPHRASE
| ENUM
| ENUMS
| ESCAPE
| EVENT
| EXCLUDE
| EXCLUDING
| EXECUTE
| EXECUTION
| EXPERIMENTAL
| EXPERIMENTAL_AUDIT
| EXPERIMENTAL_FINGERPRINTS
| EXPERIMENTAL_REPLICA
| EXPIRATION
| EXPLAIN
| EXPORT
| EXPRESSION
| EXTENDED
| EXTENSION
| EXTERNAL
| FAMILY
| FILES
| FILTER
| FINALFUNC
| FINALFUNC_EXTRA
| FINALFUNC_MODIFY
| FINALIZE
| FIRST
| FOLLOWING
| FORCE
| FORCE_INDEX
| FORMAT
| FUNCTION
| FUNCTIONS
| GENERATED
| GEOMETRYCOLLECTION
| GEOMETRYCOLLECTIONM
| GEOMETRYCOLLECTIONZ
| GEOMETRYCOLLECTIONZM
| GEOMETRYM
| GEOMETRYZ
| GEOMETRYZM
| GLOBAL
| GRANTED
| GRANTS
| GROUPS
| HANDLER
| HASH
| HEADER
| HIGH
| HISTOGRAM
| HOUR
| HYPOTHETICAL
| ICU_LOCALE
| ICU_RULES
| IDENTITY
| IGNORE_FOREIGN_KEYS
| IMMEDIATE
| IMMUTABLE
| IMPORT
| INCLUDE
| INCLUDING
| INCREMENT
| INCREMENTAL
| INDEXES
| INDEX_CLEANUP
| INHERIT
| INHERITS
| INITCOND
| INJECT
| INLINE
| INPUT
| INSERT
| INSTEAD
| INTERLEAVE
| INTERNALLENGTH
| INTO_DB
| INVERTED
| INVOKER
| ISOLATION
| IS_TEMPLATE
| JOB
| JOBS
| JSON
| KEY
| KEYS
| KMS
| KV
| LANGUAGE
| LARGE
| LAST
| LATEST
| LC_COLLATE
| LC_CTYPE
| LEAKPROOF
| LEASE
| LESS
| LEVEL
| LINESTRING
| LIST
| LOCAL
| LOCALE
| LOCALE_PROVIDER
| LOCKED
| LOGGED
| LOGIN
| LOOKUP
| LOW
| MAIN
| MATCH
| MATERIALIZED
| MAXVALUE
| MERGE
| METHOD
| MFINALFUNC
| MFINALFUNC_EXTRA
| MFINALFUNC_MODIFY
| MINITCOND
| MINUTE
| MINVALUE
| MINVFUNC
| MODIFYCLUSTERSETTING
| MODULUS
| MONTH
| MSFUNC
| MSPACE
| MSSPACE
| MSTYPE
| MULTILINESTRING
| MULTILINESTRINGM
| MULTILINESTRINGZ
| MULTILINESTRINGZM
| MULTIPOINT
| MULTIPOINTM
| MULTIPOINTZ
| MULTIPOINTZM
| MULTIPOLYGON
| MULTIPOLYGONM
| MULTIPOLYGONZ
| MULTIPOLYGONZM
| MULTIRANGE_TYPE_NAME
| NAMES
| NAN
| NEVER
| NEW
| NEXT
| NO
| NOBYPASSRLS
| NOCANCELQUERY
| NOCONTROLCHANGEFEED
| NOCONTROLJOB
| NOCREATEDB
| NOCREATELOGIN
| NOCREATEROLE
| NOINHERIT
| NOLOGIN
| NOMODIFYCLUSTERSETTING
| NOREPLICATION
| NORMAL
| NOSUPERUSER
| NOVIEWACTIVITY
| NOWAIT
| NO_INDEX_JOIN
| NULLS
| OBJECT
| OF
| OFF
| OID
| OIDS
| OLD
| ONLY_DATABASE_STATS
| OPERATOR
| OPT
| OPTION
| OPTIONS
| ORDINALITY
| OTHERS
| OUTPUT
| OVER
| OWNED
| OWNER
| PARALLEL
| PARAMETER
| PARENT
| PARSER
| PARTIAL
| PARTITION
| PARTITIONS
| PASSEDBYVALUE
| PASSWORD
| PAUSE
| PAUSED
| PHYSICAL
| PLAIN
| PLAN
| PLANS
| POINTM
| POINTZ
| POINTZM
| POLICY
| POLYGONM
| POLYGONZ
| POLYGONZM
| PRECEDING
| PREFERRED
| PREPARE
| PRESERVE
| PRIORITY
| PRIVILEGES
| PROCEDURAL
| PROCEDURE
| PROCEDURES
| PROCESS_MAIN
| PROCESS_TOAST
| PUBLIC
| PUBLICATION
| QUERIES
| QUERY
| RANGE
| RANGES
| READ
| READ_ONLY
| READ_WRITE
| RECEIVE
| RECURRING
| RECURSIVE
| REF
| REFERENCING
| REFRESH
| REINDEX
| RELEASE
| REMAINDER
| RENAME
| REPEATABLE
| REPLACE
| REPLICA
| REPLICATION
| RESET
| RESTART
| RESTORE
| RESTRICT
| RESTRICTED
| RESUME
| RETRY
| RETURN
| RETURNS
| REVISION_HISTORY
| REVOKE
| ROLE
| ROLES
| ROLLBACK
| ROLLUP
| ROUTINE
| ROUTINES
| ROWS
| RULE
| RUNNING
| SAFE
| SAVEPOINT
| SCATTER
| SCHEDULE
| SCHEDULES
| SCHEMA
| SCHEMAS
| SCRUB
| SEARCH
| SECOND
| SECURITY
| SECURITY_BARRIER
| SECURITY_INVOKER
| SEED
| SEND
| SEQUENCE
| SEQUENCES
| SERIALFUNC
| SERIALIZABLE
| SERVER
| SESSION
| SESSIONS
| SET
| SETTING
| SETTINGS
| SFUNC
| SHARE
| SHAREABLE
| SHOW
| SIMPLE
| SKIP
| SKIP_DATABASE_STATS
| SKIP_LOCKED
| SKIP_MISSING_FOREIGN_KEYS
| SKIP_MISSING_SEQUENCES
| SKIP_MISSING_SEQUENCE_OWNERS
| SKIP_MISSING_VIEWS
| SNAPSHOT
| SORTOP
| SPLIT
| SQL
| SSPACE
| STABLE
| START
| STATEMENT
| STATISTICS
| STATUS
| STDIN
| STORAGE
| STORE
| STORED
| STRATEGY
| STRICT
| STYPE
| SUBSCRIPT
| SUBSCRIPTION
| SUBTYPE
| SUBTYPE_DIFF
| SUBTYPE_OPCLASS
| SUPERUSER
| SUPPORT
| SYNTAX
| SYSID
| SYSTEM
| TABLES
| TABLESPACE
| TEMP
| TEMPLATE
| TEMPORARY
| TEXT
| THROTTLING
| TIES
| TRACE
| TRANSACTION
| TRANSACTIONS
| TRANSFORM
| TRIGGER
| TRUNCATE
| TRUSTED
| TYPE
| TYPES
| TYPMOD_IN
| TYPMOD_OUT
| UNBOUNDED
| UNCOMMITTED
| UNKNOWN
| UNLOGGED
| UNSAFE
| UNSPLIT
| UNTIL
| UPDATE
| UPSERT
| USAGE
| USE
| USERS
| VACUUM
| VALID
| VALIDATE
| VALIDATOR
| VALUE
| VARIABLE
| VARYING
| VERSION
| VIEW
| VIEWACTIVITY
| WITHIN
| WITHOUT
| WRITE
| XML
| YAML
| YEAR
| YES
| ZONE

// Column identifier --- keywords that can be column, table, etc names.
//
// Many of these keywords will in fact be recognized as type or function names
// too; but they have special productions for the purpose, and so can't be
// treated as "generic" type or function names.
//
// The type names appearing here are not usable as function names because they
// can be followed by '(' in typename productions, which looks too much like a
// function call for an LR(1) parser.
col_name_keyword:
  ANNOTATE_TYPE
| BETWEEN
| BIGINT
| BIT
| BOOLEAN
| BOX2D
| CHAR
| CHARACTER
| CHARACTERISTICS
| COALESCE
| DEC
| DECIMAL
| EXISTS
| EXTRACT
| EXTRACT_DURATION
| FLOAT
| GEOGRAPHY
| GEOMETRY
| GREATEST
| GROUPING
| IF
| IFERROR
| IFNULL
| INOUT
| INT
| INTEGER
| INTERVAL
| ISERROR
| LEAST
| NULLIF
| NUMERIC
| OUT
| OVERLAY
| POINT
| POLYGON
| POSITION
| PRECISION
| REAL
| ROW
| SETOF
| SMALLINT
| STRING
| SUBSTRING
| TIME
| TIMETZ
| TIMESTAMP
| TIMESTAMPTZ
| TREAT
| TRIM
| VALUES
| VARBIT
| VARCHAR
| VIRTUAL
| VOLATILE
| WORK

// type_func_name_keyword contains both the standard set of
// type_func_name_keyword's along with the set of CRDB extensions.
type_func_name_keyword:
  type_func_name_no_crdb_extra_keyword

// Type/function identifier --- keywords that can be type or function names.
//
// Most of these are keywords that are used as operators in expressions; in
// general such keywords can't be column names because they would be ambiguous
// with variables, but they are unambiguous as function identifiers.
//
// Do not include POSITION, SUBSTRING, etc here since they have explicit
// productions in a_expr to support the goofy SQL9x argument syntax.
// - thomas 2000-11-28
//
// *** DO NOT ADD COCKROACHDB-SPECIFIC KEYWORDS HERE ***
type_func_name_no_crdb_extra_keyword:
  AUTHORIZATION
| COLLATION
| CROSS
| FULL
| INNER
| ILIKE
| IS
| ISNULL
| JOIN
| LEFT
| LIKE
| NATURAL
| NONE
| NOTNULL
| OUTER
| OVERLAPS
| RIGHT
| SIMILAR

// Reserved keyword --- these keywords are usable only as a unrestricted_name.
//
// Keywords appear here if they could not be distinguished from variable, type,
// or function names in some contexts.
//
// *** NEVER ADD KEYWORDS HERE ***
//
// See cockroachdb_extra_reserved_keyword below.
//
reserved_keyword:
  ALL
| ANALYSE
| ANALYZE
| AND
| ANY
| ARRAY
| AS
| ASC
| ASYMMETRIC
| BOTH
| CASE
| CAST
| CHECK
| COLLATE
| COLUMN
| CONCURRENTLY
| CONNECT
| CONSTRAINT
| CREATE
| CURRENT_CATALOG
| CURRENT_DATE
| CURRENT_ROLE
| CURRENT_SCHEMA
| CURRENT_TIME
| CURRENT_TIMESTAMP
| CURRENT_USER
| DEFAULT
| DEFERRABLE
| DESC
| DESCRIBE
| DISTINCT
| DO
| ELEMENT
| ELSE
| END
| EXCEPT
| FALSE
| FETCH
| FOR
| FOREIGN
| FREEZE
| FROM
| GRANT
| GROUP
| HAVING
| IN
| INITIALLY
| INTERSECT
| INTO
| LATERAL
| LEADING
| LIMIT
| LOCALTIME
| LOCALTIMESTAMP
| NOT
| NULL
| OFFSET
| ON
| ONLY
| OR
| ORDER
| PLACING
| PRIMARY
| REFERENCES
| RETURNING
| SELECT
| SESSION_USER
| SOME
| SYMMETRIC
| TABLE
| THEN
| TO
| TRAILING
| TRUE
| UNION
| UNIQUE
| USER
| USING
| VARIADIC
| VERBOSE
| WHEN
| WHERE
| WINDOW
| WITH
| WRAPPER
| cockroachdb_extra_reserved_keyword

// Reserved keywords in CockroachDB, in addition to those reserved in
// PostgreSQL.
//
// *** REFRAIN FROM ADDING KEYWORDS HERE ***
//
// Adding keywords here creates non-resolvable incompatibilities with
// postgres clients.
cockroachdb_extra_reserved_keyword:
  INDEX
| NOTHING

%%
