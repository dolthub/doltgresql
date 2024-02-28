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

// Copyright 2015 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tree

import "strings"

var _ Statement = &AlterTable{}

// AlterTable represents an ALTER TABLE statement.
type AlterTable struct {
	IfExists bool
	Table    *UnresolvedObjectName
	Cmds     AlterTableCmds
}

// Format implements the NodeFormatter interface.
func (node *AlterTable) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER TABLE ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	ctx.FormatNode(node.Table)
	ctx.FormatNode(&node.Cmds)
}

// HoistAddColumnConstraints converts column constraints in ADD COLUMN commands,
// stored in node.Cmds, into top-level commands to add those constraints.
// Currently, this only applies to checks. For example, the ADD COLUMN in
//
//	ALTER TABLE t ADD COLUMN a INT CHECK (a < 1)
//
// is transformed into two commands, as in
//
//	ALTER TABLE t ADD COLUMN a INT, ADD CONSTRAINT check_a CHECK (a < 1)
//
// (with an auto-generated name).
//
// Note that some SQL databases require that a constraint attached to a column
// to refer only to the column it is attached to. We follow Postgres' behavior,
// however, in omitting this restriction by blindly hoisting all column
// constraints. For example, the following statement is accepted in
// CockroachDB and Postgres, but not necessarily other SQL databases:
//
//	ALTER TABLE t ADD COLUMN a INT CHECK (a < b)
func (node *AlterTable) HoistAddColumnConstraints() {
	var normalizedCmds AlterTableCmds

	for _, cmd := range node.Cmds {
		normalizedCmds = append(normalizedCmds, cmd)

		if t, ok := cmd.(*AlterTableAddColumn); ok {
			d := t.ColumnDef
			for _, checkExpr := range d.CheckExprs {
				normalizedCmds = append(normalizedCmds,
					&AlterTableAddConstraint{
						ConstraintDef: &CheckConstraintTableDef{
							Expr: checkExpr.Expr,
							Name: checkExpr.ConstraintName,
						},
						ValidationBehavior: ValidationDefault,
					},
				)
			}
			d.CheckExprs = nil
			if d.HasFKConstraint() {
				var targetCol NameList
				if d.References.Col != "" {
					targetCol = append(targetCol, d.References.Col)
				}
				fk := &ForeignKeyConstraintTableDef{
					Table:    *d.References.Table,
					FromCols: NameList{d.Name},
					ToCols:   targetCol,
					Name:     d.References.ConstraintName,
					Actions:  d.References.Actions,
					Match:    d.References.Match,
				}
				constraint := &AlterTableAddConstraint{
					ConstraintDef:      fk,
					ValidationBehavior: ValidationDefault,
				}
				normalizedCmds = append(normalizedCmds, constraint)
				d.References.Table = nil
			}
		}
	}
	node.Cmds = normalizedCmds
}

// AlterTableCmds represents a list of table alterations.
type AlterTableCmds []AlterTableCmd

// Format implements the NodeFormatter interface.
func (node *AlterTableCmds) Format(ctx *FmtCtx) {
	for i, n := range *node {
		if i > 0 {
			ctx.WriteString(",")
		}
		ctx.FormatNode(n)
	}
}

// AlterTableCmd represents a table modification operation.
type AlterTableCmd interface {
	NodeFormatter
	// Placeholder function to ensure that only desired types
	// (AlterTable*) conform to the AlterTableCmd interface.
	alterTableCmd()
}

func (*AlterTableAddColumn) alterTableCmd()            {}
func (*AlterTableAddConstraint) alterTableCmd()        {}
func (*AlterTableAlterColumnType) alterTableCmd()      {}
func (*AlterTableAlterConstraint) alterTableCmd()      {}
func (*AlterTableCluster) alterTableCmd()              {}
func (*AlterTableColSetStorage) alterTableCmd()        {}
func (*AlterTableComputed) alterTableCmd()             {}
func (*AlterTableConstraintUsingIndex) alterTableCmd() {}
func (*AlterTableDropColumn) alterTableCmd()           {}
func (*AlterTableDropConstraint) alterTableCmd()       {}
func (*AlterTableDropNotNull) alterTableCmd()          {}
func (*AlterTableDropExprIden) alterTableCmd()         {}
func (*AlterTableInherit) alterTableCmd()              {}
func (*AlterTableOfType) alterTableCmd()               {}
func (*AlterTableOwner) alterTableCmd()                {}
func (*AlterTableRenameColumn) alterTableCmd()         {}
func (*AlterTableRenameConstraint) alterTableCmd()     {}
func (*AlterTableReplicaIdentity) alterTableCmd()      {}
func (*AlterTableRowLevelSecurity) alterTableCmd()     {}
func (*AlterTableRule) alterTableCmd()                 {}
func (*AlterTableSetAccessMethod) alterTableCmd()      {}
func (*AlterTableSetAttribution) alterTableCmd()       {}
func (*AlterTableSetCompression) alterTableCmd()       {}
func (*AlterTableSetDefault) alterTableCmd()           {}
func (*AlterTableSetLog) alterTableCmd()               {}
func (*AlterTableSetNotNull) alterTableCmd()           {}
func (*AlterTableSetStatistics) alterTableCmd()        {}
func (*AlterTableSetStorage) alterTableCmd()           {}
func (*AlterTableSetTablespace) alterTableCmd()        {}
func (*AlterTableTrigger) alterTableCmd()              {}
func (*AlterTableValidateConstraint) alterTableCmd()   {}

var _ AlterTableCmd = &AlterTableAddColumn{}
var _ AlterTableCmd = &AlterTableAddConstraint{}
var _ AlterTableCmd = &AlterTableAlterColumnType{}
var _ AlterTableCmd = &AlterTableAlterConstraint{}
var _ AlterTableCmd = &AlterTableCluster{}
var _ AlterTableCmd = &AlterTableColSetStorage{}
var _ AlterTableCmd = &AlterTableComputed{}
var _ AlterTableCmd = &AlterTableConstraintUsingIndex{}
var _ AlterTableCmd = &AlterTableDropColumn{}
var _ AlterTableCmd = &AlterTableDropConstraint{}
var _ AlterTableCmd = &AlterTableDropNotNull{}
var _ AlterTableCmd = &AlterTableDropExprIden{}
var _ AlterTableCmd = &AlterTableInherit{}
var _ AlterTableCmd = &AlterTableOfType{}
var _ AlterTableCmd = &AlterTableOwner{}
var _ AlterTableCmd = &AlterTableRenameColumn{}
var _ AlterTableCmd = &AlterTableRenameConstraint{}
var _ AlterTableCmd = &AlterTableReplicaIdentity{}
var _ AlterTableCmd = &AlterTableRowLevelSecurity{}
var _ AlterTableCmd = &AlterTableRule{}
var _ AlterTableCmd = &AlterTableSetAccessMethod{}
var _ AlterTableCmd = &AlterTableSetAttribution{}
var _ AlterTableCmd = &AlterTableSetCompression{}
var _ AlterTableCmd = &AlterTableSetDefault{}
var _ AlterTableCmd = &AlterTableSetLog{}
var _ AlterTableCmd = &AlterTableSetNotNull{}
var _ AlterTableCmd = &AlterTableSetStatistics{}
var _ AlterTableCmd = &AlterTableSetStorage{}
var _ AlterTableCmd = &AlterTableSetTablespace{}
var _ AlterTableCmd = &AlterTableTrigger{}
var _ AlterTableCmd = &AlterTableValidateConstraint{}

// ColumnMutationCmd is the subset of AlterTableCmds that modify an
// existing column.
type ColumnMutationCmd interface {
	AlterTableCmd
	GetColumn() Name
}

// AlterTableAddColumn represents an ADD COLUMN command.
type AlterTableAddColumn struct {
	IfNotExists bool
	ColumnDef   *ColumnTableDef
}

// Format implements the NodeFormatter interface.
func (node *AlterTableAddColumn) Format(ctx *FmtCtx) {
	ctx.WriteString(" ADD COLUMN ")
	if node.IfNotExists {
		ctx.WriteString("IF NOT EXISTS ")
	}
	ctx.FormatNode(node.ColumnDef)
}

// ValidationBehavior specifies whether or not a constraint is validated.
type ValidationBehavior int

const (
	// ValidationDefault is the default validation behavior (immediate).
	ValidationDefault ValidationBehavior = iota
	// ValidationSkip skips validation of any existing data.
	ValidationSkip
)

// AlterTableAddConstraint represents an ADD CONSTRAINT command.
type AlterTableAddConstraint struct {
	ConstraintDef      ConstraintTableDef
	ValidationBehavior ValidationBehavior
}

// Format implements the NodeFormatter interface.
func (node *AlterTableAddConstraint) Format(ctx *FmtCtx) {
	ctx.WriteString(" ADD ")
	ctx.FormatNode(node.ConstraintDef)
	if node.ValidationBehavior == ValidationSkip {
		ctx.WriteString(" NOT VALID")
	}
}

// AlterTableAlterColumnType represents an ALTER TABLE ALTER COLUMN TYPE command.
type AlterTableAlterColumnType struct {
	Collation string
	Column    Name
	ToType    ResolvableTypeReference
	Using     Expr
}

// Format implements the NodeFormatter interface.
func (node *AlterTableAlterColumnType) Format(ctx *FmtCtx) {
	ctx.WriteString(" ALTER COLUMN ")
	ctx.FormatNode(&node.Column)
	ctx.WriteString(" SET DATA TYPE ")
	ctx.WriteString(node.ToType.SQLString())
	if len(node.Collation) > 0 {
		ctx.WriteString(" COLLATE ")
		ctx.WriteString(node.Collation)
	}
	if node.Using != nil {
		ctx.WriteString(" USING ")
		ctx.FormatNode(node.Using)
	}
}

// GetColumn implements the ColumnMutationCmd interface.
func (node *AlterTableAlterColumnType) GetColumn() Name {
	return node.Column
}

// AlterTableAlterConstraint represents an ALTER CONSTRAINT command.
type AlterTableAlterConstraint struct {
	Constraint Name
	Deferrable DeferrableMode
	Initially  InitiallyMode
}

// Format implements the NodeFormatter interface.
func (node *AlterTableAlterConstraint) Format(ctx *FmtCtx) {
	ctx.WriteString(" ALTER CONSTRAINT ")
	ctx.FormatNode(&node.Constraint)
	switch node.Deferrable {
	case Deferrable:
		ctx.WriteString(" DEFERRABLE")
		switch node.Initially {
		case InitiallyImmediate:
			ctx.WriteString(" INITIALLY IMMEDIATE")
		case InitiallyDeferred:
			ctx.WriteString(" INITIALLY DEFERRED")
		default:
		}
	case NotDeferrable:
		ctx.WriteString(" NOT DEFERRABLE")
	default:
	}
}

// AlterTableCluster represents an ALTER TABLE { CLUSTER ON ... | WITHOUT CLUSTER } command.
type AlterTableCluster struct {
	OnIndex Name
	Without bool
}

// Format implements the NodeFormatter interface.
func (node *AlterTableCluster) Format(ctx *FmtCtx) {
	if node.Without {
		ctx.WriteString(" SET WITHOUT CLUSTER")
	} else {
		ctx.WriteString(" CLUSTER ON ")
		ctx.FormatNode(&node.OnIndex)
	}
}

type StorageType int

const (
	StoragePlain StorageType = iota
	StorageExternal
	StorageExtended
	StorageMain
)

// AlterTableColSetStorage represents an ALTER COLUMN SET STORAGE command
// to remove the computed-ness from a column.
type AlterTableColSetStorage struct {
	Column Name
	Type   StorageType
}

// GetColumn implements the ColumnMutationCmd interface.
func (node *AlterTableColSetStorage) GetColumn() Name {
	return node.Column
}

// Format implements the NodeFormatter interface.
func (node *AlterTableColSetStorage) Format(ctx *FmtCtx) {
	ctx.WriteString(" ALTER COLUMN ")
	ctx.FormatNode(&node.Column)
	ctx.WriteString(" SET STORAGE ")
	switch node.Type {
	case StoragePlain:
		ctx.WriteString("PLAIN")
	case StorageExternal:
		ctx.WriteString("EXTERNAL")
	case StorageExtended:
		ctx.WriteString("EXTENDED")
	case StorageMain:
		ctx.WriteString("MAIN")
	}
}

type AlterColComputed struct {
	ByDefault bool
	Options   SequenceOptions
	IsRestart bool
	Restart   Expr
}

// AlterTableComputed represents an ALTER TABLE ALTER COLUMN SET GENERATED command.
type AlterTableComputed struct {
	Column  Name
	Defs    []AlterColComputed
	IsAdd   bool
	AddDefs ColumnQualification
}

// Format implements the NodeFormatter interface.
func (node *AlterTableComputed) Format(ctx *FmtCtx) {
	ctx.WriteString(" ALTER COLUMN ")
	ctx.FormatNode(&node.Column)
	if node.IsAdd {

	} else {
		for i, s := range node.Defs {
			if i != 0 {
				ctx.WriteByte(' ')
			}
			if s.ByDefault {
				ctx.WriteString("SET GENERATED BY DEFAULT")
			} else if s.Options == nil {
				ctx.WriteString("SET ")
				ctx.FormatNode(&s.Options)
			} else if s.IsRestart {
				ctx.WriteString("RESTART ")
				if s.Restart != nil {
					ctx.WriteString("WITH ")
					ctx.FormatNode(s.Restart)
				}
			} else {
				ctx.WriteString("SET GENERATED ALWAYS")
			}
		}
	}
}

// GetColumn implements the ColumnMutationCmd interface.
func (node *AlterTableComputed) GetColumn() Name {
	return node.Column
}

// AlterTableConstraintUsingIndex represents an ALTER TABLE ADD table_constraint USING INDEX command.
type AlterTableConstraintUsingIndex struct {
	Constraint Name
	IsUnique   bool
	Index      Name
	Deferrable DeferrableMode
	Initially  InitiallyMode
}

// Format implements the NodeFormatter interface.
func (node *AlterTableConstraintUsingIndex) Format(ctx *FmtCtx) {
	if node.Constraint != "" {
		ctx.WriteString(" CONSTRAINT")
		ctx.FormatNode(&node.Constraint)
	}
	if node.IsUnique {
		ctx.WriteString(" UNIQUE")
	} else {
		ctx.WriteString(" PRIMARY KEY")
	}
	ctx.WriteString(" USING INDEX")
	ctx.FormatNode(&node.Index)
	switch node.Deferrable {
	case Deferrable:
		ctx.WriteString(" DEFERRABLE")
		switch node.Initially {
		case InitiallyImmediate:
			ctx.WriteString(" INITIALLY IMMEDIATE")
		case InitiallyDeferred:
			ctx.WriteString(" INITIALLY DEFERRED")
		default:
		}
	case NotDeferrable:
		ctx.WriteString(" NOT DEFERRABLE")
	default:
	}
}

// AlterTableDropColumn represents a DROP COLUMN command.
type AlterTableDropColumn struct {
	IfExists     bool
	Column       Name
	DropBehavior DropBehavior
}

// Format implements the NodeFormatter interface.
func (node *AlterTableDropColumn) Format(ctx *FmtCtx) {
	ctx.WriteString(" DROP COLUMN ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	ctx.FormatNode(&node.Column)
	if node.DropBehavior != DropDefault {
		ctx.Printf(" %s", node.DropBehavior)
	}
}

// AlterTableDropConstraint represents a DROP CONSTRAINT command.
type AlterTableDropConstraint struct {
	IfExists     bool
	Constraint   Name
	DropBehavior DropBehavior
}

// Format implements the NodeFormatter interface.
func (node *AlterTableDropConstraint) Format(ctx *FmtCtx) {
	ctx.WriteString(" DROP CONSTRAINT ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	ctx.FormatNode(&node.Constraint)
	if node.DropBehavior != DropDefault {
		ctx.Printf(" %s", node.DropBehavior)
	}
}

// AlterTableDropNotNull represents an ALTER COLUMN DROP NOT NULL
// command.
type AlterTableDropNotNull struct {
	Column Name
}

// GetColumn implements the ColumnMutationCmd interface.
func (node *AlterTableDropNotNull) GetColumn() Name {
	return node.Column
}

// Format implements the NodeFormatter interface.
func (node *AlterTableDropNotNull) Format(ctx *FmtCtx) {
	ctx.WriteString(" ALTER COLUMN ")
	ctx.FormatNode(&node.Column)
	ctx.WriteString(" DROP NOT NULL")
}

// AlterTableDropExprIden represents an ALTER COLUMN DROP EXPRESSION | IDENTITY command
// to remove the computed-ness from a column.
type AlterTableDropExprIden struct {
	Column     Name
	IsIdentity bool
	IfExists   bool
}

// GetColumn implements the ColumnMutationCmd interface.
func (node *AlterTableDropExprIden) GetColumn() Name {
	return node.Column
}

// Format implements the NodeFormatter interface.
func (node *AlterTableDropExprIden) Format(ctx *FmtCtx) {
	ctx.WriteString(" ALTER COLUMN ")
	ctx.FormatNode(&node.Column)
	if node.IsIdentity {
		ctx.WriteString(" DROP IDENTITY")
	} else {
		ctx.WriteString(" DROP EXPRESSION")
	}
	if node.IfExists {
		ctx.WriteString(" IF EXISTS")
	}
}

// AlterTableInherit represents an ALTER TABLE { INHERIT | NO INHERIT } ... command.
type AlterTableInherit struct {
	Inherit bool
	Table   TableName
}

// Format implements the NodeFormatter interface.
func (node *AlterTableInherit) Format(ctx *FmtCtx) {
	if node.Inherit {
		ctx.WriteString(" INHERIT")
	} else {
		ctx.WriteString(" NO INHERIT")
	}
	ctx.FormatNode(&node.Table)
}

// AlterTableOfType represents an ALTER TABLE { OF type_name | NOT OF } command.
type AlterTableOfType struct {
	Type  ResolvableTypeReference
	NotOf bool
}

// Format implements the NodeFormatter interface.
func (node *AlterTableOfType) Format(ctx *FmtCtx) {
	if node.NotOf {
		ctx.WriteString(" NOT OF")
	} else {
		ctx.WriteString(" OF ")
		ctx.WriteString(node.Type.SQLString())
	}
}

// AlterTableOwner represents an ALTER TABLE OWNER TO command.
type AlterTableOwner struct {
	Owner string
}

// Format implements the NodeFormatter interface.
func (node *AlterTableOwner) Format(ctx *FmtCtx) {
	ctx.WriteString(" OWNER TO ")
	ctx.FormatNameP(&node.Owner)
}

// AlterTableRenameColumn represents an ALTER TABLE RENAME [COLUMN] command.
type AlterTableRenameColumn struct {
	Column  Name
	NewName Name
}

// Format implements the NodeFormatter interface.
func (node *AlterTableRenameColumn) Format(ctx *FmtCtx) {
	ctx.WriteString(" RENAME COLUMN ")
	ctx.FormatNode(&node.Column)
	ctx.WriteString(" TO ")
	ctx.FormatNode(&node.NewName)
}

// AlterTableRenameConstraint represents an ALTER TABLE RENAME CONSTRAINT command.
type AlterTableRenameConstraint struct {
	Constraint Name
	NewName    Name
}

// Format implements the NodeFormatter interface.
func (node *AlterTableRenameConstraint) Format(ctx *FmtCtx) {
	ctx.WriteString(" RENAME CONSTRAINT ")
	ctx.FormatNode(&node.Constraint)
	ctx.WriteString(" TO ")
	ctx.FormatNode(&node.NewName)
}

type ReplicaIdentity int

const (
	ReplicaIdentityDefault ReplicaIdentity = iota
	ReplicaIdentityUsingIndex
	ReplicaIdentityFull
	ReplicaIdentityNothing
)

// AlterTableReplicaIdentity represents an ALTER TABLE REPLICA IDENTITY command.
type AlterTableReplicaIdentity struct {
	Type  ReplicaIdentity
	Index Name
}

// Format implements the NodeFormatter interface.
func (node *AlterTableReplicaIdentity) Format(ctx *FmtCtx) {
	ctx.WriteString(" REPLICA IDENTITY")
	switch node.Type {
	case ReplicaIdentityDefault:
		ctx.WriteString(" DEFAULT")
	case ReplicaIdentityUsingIndex:
		ctx.WriteString(" USING INDEX ")
		ctx.FormatNode(&node.Index)
	case ReplicaIdentityFull:
		ctx.WriteString(" FULL")
	case ReplicaIdentityNothing:
		ctx.WriteString(" NOTHING")
	}
}

type RowLevelSecurity int

const (
	RowLevelSecurityDisable RowLevelSecurity = iota
	RowLevelSecurityEnable
	RowLevelSecurityForce
	RowLevelSecurityNoForce
)

// AlterTableRowLevelSecurity represents an ALTER TABLE ... ROW LEVEL SECURITY command.
type AlterTableRowLevelSecurity struct {
	Type RowLevelSecurity
}

// Format implements the NodeFormatter interface.
func (node *AlterTableRowLevelSecurity) Format(ctx *FmtCtx) {
	switch node.Type {
	case RowLevelSecurityDisable:
		ctx.WriteString(" DISABLE")
	case RowLevelSecurityEnable:
		ctx.WriteString(" ENABLE")
	case RowLevelSecurityForce:
		ctx.WriteString(" FORCE")
	case RowLevelSecurityNoForce:
		ctx.WriteString(" NO FORCE")
	}
	ctx.WriteString(" ROW LEVEL SECURITY")
}

// AlterTableRule represents an ALTER TABLE {DISABLE|ENABLE} RULE command.
type AlterTableRule struct {
	Disable   bool
	Rule      string
	IsReplica bool
	IsAlways  bool
}

// Format implements the NodeFormatter interface.
func (node *AlterTableRule) Format(ctx *FmtCtx) {
	if node.Disable {
		ctx.WriteString(" DISABLE")
	} else {
		ctx.WriteString(" ENABLE")
	}
	if node.IsReplica {
		ctx.WriteString(" REPLICA")
	} else if node.IsAlways {
		ctx.WriteString(" ALWAYS")
	}
	ctx.WriteString(" RULE")
	ctx.FormatNameP(&node.Rule)
}

// AlterTableSetAccessMethod represents an ALTER TABLE SET ACCESS METHOD ... command.
type AlterTableSetAccessMethod struct {
	Method string
}

// Format implements the NodeFormatter interface.
func (node *AlterTableSetAccessMethod) Format(ctx *FmtCtx) {
	ctx.WriteString(" SET ACCESS METHOD ")
	ctx.WriteString(node.Method)
}

// AlterTableSetAttribution represents an ALTER COLUMN SET | RESET ATTRIBUTION command
// to remove the computed-ness from a column.
type AlterTableSetAttribution struct {
	Column Name
	Reset  bool
	Params StorageParams
}

// GetColumn implements the ColumnMutationCmd interface.
func (node *AlterTableSetAttribution) GetColumn() Name {
	return node.Column
}

// Format implements the NodeFormatter interface.
func (node *AlterTableSetAttribution) Format(ctx *FmtCtx) {
	ctx.WriteString(" ALTER COLUMN ")
	ctx.FormatNode(&node.Column)
	if node.Reset {
		ctx.WriteString(" RESET ")
	} else {
		ctx.WriteString(" SET ")
	}
	ctx.FormatNode(&node.Params)
}

// AlterTableSetCompression represents an ALTER COLUMN SET COMPRESSION command
// to remove the computed-ness from a column.
type AlterTableSetCompression struct {
	Column      Name
	Compression string
}

// GetColumn implements the ColumnMutationCmd interface.
func (node *AlterTableSetCompression) GetColumn() Name {
	return node.Column
}

// Format implements the NodeFormatter interface.
func (node *AlterTableSetCompression) Format(ctx *FmtCtx) {
	ctx.WriteString(" ALTER COLUMN ")
	ctx.FormatNode(&node.Column)
	ctx.WriteString(" SET COMPRESSION")
	ctx.WriteString(node.Compression)
}

// AlterTableSetDefault represents an ALTER COLUMN SET DEFAULT
// or DROP DEFAULT command.
type AlterTableSetDefault struct {
	Column  Name
	Default Expr
}

// GetColumn implements the ColumnMutationCmd interface.
func (node *AlterTableSetDefault) GetColumn() Name {
	return node.Column
}

// Format implements the NodeFormatter interface.
func (node *AlterTableSetDefault) Format(ctx *FmtCtx) {
	ctx.WriteString(" ALTER COLUMN ")
	ctx.FormatNode(&node.Column)
	if node.Default == nil {
		ctx.WriteString(" DROP DEFAULT")
	} else {
		ctx.WriteString(" SET DEFAULT ")
		ctx.FormatNode(node.Default)
	}
}

// AlterTableSetLog represents an ALTER TABLE SET { LOGGED | UNLOGGED }command.
type AlterTableSetLog struct {
	Logged bool
}

// Format implements the NodeFormatter interface.
func (node *AlterTableSetLog) Format(ctx *FmtCtx) {
	ctx.WriteString(" SET ")
	if node.Logged {
		ctx.WriteString("LOGGED")
	} else {
		ctx.WriteString("UNLOGGED")
	}
}

// AlterTableSetNotNull represents an ALTER COLUMN SET NOT NULL
// command.
type AlterTableSetNotNull struct {
	Column Name
}

// GetColumn implements the ColumnMutationCmd interface.
func (node *AlterTableSetNotNull) GetColumn() Name {
	return node.Column
}

// Format implements the NodeFormatter interface.
func (node *AlterTableSetNotNull) Format(ctx *FmtCtx) {
	ctx.WriteString(" ALTER COLUMN ")
	ctx.FormatNode(&node.Column)
	ctx.WriteString(" SET NOT NULL")
}

// AlterTableSetStatistics represents an ALTER COLUMN SET STATISTICS command
// to remove the computed-ness from a column.
type AlterTableSetStatistics struct {
	Column Name
	Num    Expr
}

// GetColumn implements the ColumnMutationCmd interface.
func (node *AlterTableSetStatistics) GetColumn() Name {
	return node.Column
}

// Format implements the NodeFormatter interface.
func (node *AlterTableSetStatistics) Format(ctx *FmtCtx) {
	ctx.WriteString(" ALTER COLUMN ")
	ctx.FormatNode(&node.Column)
	ctx.WriteString(" SET STATISTICS ")
	ctx.FormatNode(node.Num)
}

// AlterTableSetStorage represents an ALTER TABLE { SET | RESET } ( ... ) command.
type AlterTableSetStorage struct {
	Params  StorageParams
	IsReset bool
}

// Format implements the NodeFormatter interface.
func (node *AlterTableSetStorage) Format(ctx *FmtCtx) {
	ctx.WriteString(" SET TABLESPACE")
	if node.IsReset {
		ctx.WriteString(" RESET ( ")
	} else {
		ctx.WriteString(" SET ( ")
	}
	ctx.FormatNode(&node.Params)
	ctx.WriteString(" )")
}

// AlterTableSetTablespace represents an ALTER TABLE SET TABLESPACE ... command.
type AlterTableSetTablespace struct {
	Tablespace string
}

// Format implements the NodeFormatter interface.
func (node *AlterTableSetTablespace) Format(ctx *FmtCtx) {
	ctx.WriteString(" SET TABLESPACE ")
	ctx.WriteString(node.Tablespace)
}

// AlterTableTrigger represents an ALTER TABLE {DISABLE|ENABLE} TRIGGER command.
type AlterTableTrigger struct {
	Disable   bool
	Trigger   string
	IsReplica bool
	IsAlways  bool
}

// Format implements the NodeFormatter interface.
func (node *AlterTableTrigger) Format(ctx *FmtCtx) {
	if node.Disable {
		ctx.WriteString(" DISABLE")
	} else {
		ctx.WriteString(" ENABLE")
	}
	if node.IsReplica {
		ctx.WriteString(" REPLICA")
	} else if node.IsAlways {
		ctx.WriteString(" ALWAYS")
	}
	ctx.WriteString(" TRIGGER")
	ctx.FormatNameP(&node.Trigger)
}

// AlterTableValidateConstraint represents a VALIDATE CONSTRAINT command.
type AlterTableValidateConstraint struct {
	Constraint Name
}

// Format implements the NodeFormatter interface.
func (node *AlterTableValidateConstraint) Format(ctx *FmtCtx) {
	ctx.WriteString(" VALIDATE CONSTRAINT ")
	ctx.FormatNode(&node.Constraint)
}

var _ Statement = &AlterTableSetSchema{}

// AlterTableSetSchema represents an ALTER TABLE SET SCHEMA statement.
type AlterTableSetSchema struct {
	Name     *UnresolvedObjectName
	Schema   string
	IfExists bool

	IsMaterialized bool
	IsSequence     bool
}

// Format implements the NodeFormatter interface.
func (node *AlterTableSetSchema) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER ")
	if node.IsMaterialized {
		ctx.WriteString("MATERIALIZED VIEW")
	} else if node.IsSequence {
		ctx.WriteString(" SEQUENCE ")
	} else {
		ctx.WriteString("TABLE")
	}
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	node.Name.Format(ctx)
	ctx.WriteString(" SET SCHEMA ")
	ctx.WriteString(node.Schema)
}

var _ Statement = &AlterTableAllInTablespace{}

// AlterTableAllInTablespace represents an ALTER { TABLE | MATERIALIZED VIEW } ALL IN TABLESPACE ... statement.
type AlterTableAllInTablespace struct {
	Name       Name
	OwnedBy    []string
	Tablespace string
	NoWait     bool

	IsMaterialized bool
}

// Format implements the NodeFormatter interface.
func (node *AlterTableAllInTablespace) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER ")
	if node.IsMaterialized {
		ctx.WriteString("MATERIALIZED VIEW")
	} else {
		ctx.WriteString("TABLE")
	}
	ctx.WriteString(" ALL IN TABLESPACE ")
	ctx.FormatNode(&node.Name)
	if node.OwnedBy != nil {
		ctx.WriteString(" OWNED BY ")
		ctx.WriteString(strings.Join(node.OwnedBy, ", "))
	}
	ctx.WriteString(" SET TABLESPACE ")
	ctx.WriteString(node.Tablespace)
	if node.NoWait {
		ctx.WriteString(" NOWAIT")
	}
}

type DetachPartition int

const (
	DetachPartitionNone DetachPartition = iota
	DetachPartitionConcurrently
	DetachPartitionFinalize
)

var _ Statement = &AlterTablePartition{}

// AlterTablePartition represents an ALTER TABLE { ATTACH | DETACH } PARTITION ...
// command.
type AlterTablePartition struct {
	Name       *UnresolvedObjectName
	IfExists   bool
	Partition  Name
	Spec       PartitionBoundSpec
	IsDetach   bool
	DetachType DetachPartition
}

// Format implements the NodeFormatter interface.
func (node *AlterTablePartition) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER TABLE ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	node.Name.Format(ctx)
	if node.IsDetach {
		ctx.WriteString(" DETACH PARTITION ")
		ctx.FormatNode(&node.Partition)
		switch node.DetachType {
		case DetachPartitionNone:
		case DetachPartitionConcurrently:
			ctx.WriteString(" CONCURRENTLY")
		case DetachPartitionFinalize:
			ctx.WriteString(" FINALIZE")
		}
	} else {
		ctx.WriteString(" ATTACH PARTITION ")
		ctx.FormatNode(&node.Partition)
		ctx.WriteByte(' ')
		ctx.FormatNode(&node.Spec)
	}
}
