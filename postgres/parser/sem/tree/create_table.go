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

import (
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/roleoption"
	"github.com/dolthub/doltgresql/postgres/parser/types"
)

// TableDef represents a column, index or constraint definition within a CREATE
// TABLE statement.
type TableDef interface {
	NodeFormatter
	// Placeholder function to ensure that only desired types (*TableDef) conform
	// to the TableDef interface.
	tableDef()
}

func (*ColumnTableDef) tableDef()               {}
func (*IndexTableDef) tableDef()                {}
func (*ForeignKeyConstraintTableDef) tableDef() {}
func (*CheckConstraintTableDef) tableDef()      {}
func (*ExcludeConstraintTableDef) tableDef()    {}
func (*LikeTableDef) tableDef()                 {}

// TableDefs represents a list of table definitions.
type TableDefs []TableDef

// Format implements the NodeFormatter interface.
func (node *TableDefs) Format(ctx *FmtCtx) {
	for i, n := range *node {
		if i > 0 {
			ctx.WriteString(", ")
		}
		ctx.FormatNode(n)
	}
}

// Nullability represents either NULL, NOT NULL or an unspecified value (silent
// NULL).
type Nullability int

// The values for NullType.
const (
	NotNull Nullability = iota
	Null
	SilentNull
)

// ColumnTableDef represents a column definition within a CREATE TABLE
// statement.
type ColumnTableDef struct {
	Name        Name
	Type        ResolvableTypeReference
	Compression string
	Collation   string
	IsSerial    bool
	Nullable    struct {
		Nullability    Nullability
		ConstraintName Name
	}
	// only UNIQUE, PRIMARY KEY, EXCLUDE, and REFERENCES (foreign key) constraints accept this clause
	PrimaryKey struct {
		IsPrimaryKey bool
	}
	Unique               bool
	UniqueConstraintName Name
	UniqueDeferrable     DeferrableMode
	UniqueInitially      InitiallyMode
	DefaultExpr          struct {
		Expr           Expr
		ConstraintName Name
	}
	CheckExprs []ColumnTableDefCheckExpr
	References struct {
		Table          *TableName
		Col            Name
		ConstraintName Name
		Actions        ReferenceActions
		Match          CompositeKeyMatchMethod
		Deferrable     DeferrableMode
		Initially      InitiallyMode
	}
	Computed struct {
		Computed  bool
		ByDefault bool
		Expr      Expr
		Options   SequenceOptions
	}
}

// ColumnTableDefCheckExpr represents a check constraint on a column definition
// within a CREATE TABLE statement.
type ColumnTableDefCheckExpr struct {
	Expr           Expr
	ConstraintName Name
	NoInherit      bool
}

func processCollationOnType(name Name, ref ResolvableTypeReference, c string) (*types.T, error) {
	// At the moment, only string types can be collated. User defined types
	//  like enums don't support collations, so check this at parse time.
	typ, ok := GetStaticallyKnownType(ref)
	if !ok {
		return nil, pgerror.Newf(pgcode.DatatypeMismatch,
			"COLLATE declaration for non-string-typed column %q", name)
	}
	switch typ.Family() {
	case types.StringFamily:
		return types.MakeCollatedString(typ, string(c)), nil
	case types.CollatedStringFamily:
		return nil, pgerror.Newf(pgcode.Syntax,
			"multiple COLLATE declarations for column %q", name)
	case types.ArrayFamily:
		elemTyp, err := processCollationOnType(name, typ.ArrayContents(), c)
		if err != nil {
			return nil, err
		}
		return types.MakeArray(elemTyp), nil
	default:
		return nil, pgerror.Newf(pgcode.DatatypeMismatch,
			"COLLATE declaration for non-string-typed column %q", name)
	}
}

// NewColumnTableDef constructs a column definition for a CreateTable statement.
func NewColumnTableDef(
	name Name,
	typRef ResolvableTypeReference,
	compression string,
	collation string,
	qualifications []NamedColumnQualification,
) (*ColumnTableDef, error) {
	var isSerial bool
	if typRef != nil {
		isSerial = IsReferenceSerialType(typRef)
	}
	d := &ColumnTableDef{
		Name:     name,
		Type:     typRef,
		IsSerial: isSerial,
	}
	d.Nullable.Nullability = SilentNull
	//if collation != "" {
	//	_, err := language.Parse(collation)
	//	if err != nil {
	//		return nil, pgerror.Wrapf(err, pgcode.Syntax, "invalid locale %s", collation)
	//	}
	//	collatedTyp, err := processCollationOnType(name, d.Type, collation)
	//	if err != nil {
	//		return nil, err
	//	}
	//	d.Type = collatedTyp
	//}
	for _, c := range qualifications {
		switch t := c.Qualification.(type) {
		case *ColumnDefault:
			if d.HasDefaultExpr() {
				return nil, pgerror.Newf(pgcode.Syntax,
					"multiple default values specified for column %q", name)
			}
			d.DefaultExpr.Expr = t.Expr
			d.DefaultExpr.ConstraintName = c.Name
		case NotNullConstraint:
			if d.Nullable.Nullability == Null {
				return nil, pgerror.Newf(pgcode.Syntax,
					"conflicting NULL/NOT NULL declarations for column %q", name)
			}
			d.Nullable.Nullability = NotNull
			d.Nullable.ConstraintName = c.Name
		case NullConstraint:
			if d.Nullable.Nullability == NotNull {
				return nil, pgerror.Newf(pgcode.Syntax,
					"conflicting NULL/NOT NULL declarations for column %q", name)
			}
			d.Nullable.Nullability = Null
			d.Nullable.ConstraintName = c.Name
		case UniqueConstraint:
			if t.IsPrimary {
				d.PrimaryKey.IsPrimaryKey = true
			} else {
				d.Unique = true
			}
			d.UniqueConstraintName = c.Name
			d.UniqueDeferrable = c.Deferrable
			d.UniqueInitially = c.Initially
		case *ColumnCheckConstraint:
			d.CheckExprs = append(d.CheckExprs, ColumnTableDefCheckExpr{
				Expr:           t.Expr,
				ConstraintName: c.Name,
			})
		case *ColumnFKConstraint:
			//if d.HasFKConstraint() {
			//	return nil, pgerror.Newf(pgcode.InvalidTableDefinition,
			//		"multiple foreign key constraints specified for column %q", name)
			//}
			d.References.Table = &t.Table
			d.References.Col = t.Col
			d.References.ConstraintName = c.Name
			d.References.Actions = t.Actions
			d.References.Match = t.Match
			d.References.Deferrable = c.Deferrable
			d.References.Initially = c.Initially
		case *ColumnComputedDef:
			d.Computed.Computed = true
			d.Computed.ByDefault = t.ByDefault
			d.Computed.Expr = t.Expr
			d.Computed.Options = t.Options
		default:
			return nil, errors.AssertionFailedf("unexpected column qualification: %T", c)
		}
	}
	return d, nil
}

// HasDefaultExpr returns if the ColumnTableDef has a default expression.
func (node *ColumnTableDef) HasDefaultExpr() bool {
	return node.DefaultExpr.Expr != nil
}

// HasFKConstraint returns if the ColumnTableDef has a foreign key constraint.
func (node *ColumnTableDef) HasFKConstraint() bool {
	return node.References.Table != nil
}

// IsComputed returns if the ColumnTableDef is a computed column.
func (node *ColumnTableDef) IsComputed() bool {
	return node.Computed.Computed
}

// Format implements the NodeFormatter interface.
func (node *ColumnTableDef) Format(ctx *FmtCtx) {
	ctx.FormatNode(&node.Name)

	// ColumnTableDef node type will not be specified if it represents a CREATE
	// TABLE ... AS query.
	if node.Type != nil {
		ctx.WriteByte(' ')
		ctx.WriteString(node.columnTypeString())
	}

	if node.Nullable.Nullability != SilentNull && node.Nullable.ConstraintName != "" {
		ctx.WriteString(" CONSTRAINT ")
		ctx.FormatNode(&node.Nullable.ConstraintName)
	}
	switch node.Nullable.Nullability {
	case Null:
		ctx.WriteString(" NULL")
	case NotNull:
		ctx.WriteString(" NOT NULL")
	default:
	}
	for _, checkExpr := range node.CheckExprs {
		if checkExpr.ConstraintName != "" {
			ctx.WriteString(" CONSTRAINT ")
			ctx.FormatNode(&checkExpr.ConstraintName)
		}
		ctx.WriteString(" CHECK (")
		ctx.FormatNode(checkExpr.Expr)
		ctx.WriteByte(')')
		if checkExpr.NoInherit {
			ctx.WriteString(" NO INHERIT")
		}
	}
	if node.HasDefaultExpr() {
		if node.DefaultExpr.ConstraintName != "" {
			ctx.WriteString(" CONSTRAINT ")
			ctx.FormatNode(&node.DefaultExpr.ConstraintName)
		}
		ctx.WriteString(" DEFAULT ")
		ctx.FormatNode(node.DefaultExpr.Expr)
	}
	if node.IsComputed() {
		ctx.WriteString(" GENERATED")
		if node.Computed.ByDefault {
			ctx.WriteString(" BY DEFAULT")
		} else {
			ctx.WriteString(" ALWAYS")
		}
		ctx.WriteString(" AS")
		if node.Computed.Expr != nil {
			ctx.WriteString(" ( ")
			ctx.FormatNode(node.Computed.Expr)
			ctx.WriteString(" ) STORED")
		} else {
			ctx.WriteString(" IDENTITY")
			if node.Computed.Options != nil {
				ctx.WriteString(" ( ")
				ctx.FormatNode(&node.Computed.Options)
				ctx.WriteString(" )")
			}
		}
	}
	if node.PrimaryKey.IsPrimaryKey || node.Unique {
		if node.UniqueConstraintName != "" {
			ctx.WriteString(" CONSTRAINT ")
			ctx.FormatNode(&node.UniqueConstraintName)
		}
		if node.PrimaryKey.IsPrimaryKey {
			ctx.WriteString(" PRIMARY KEY")
		} else if node.Unique {
			ctx.WriteString(" UNIQUE")
		}
		switch node.UniqueDeferrable {
		case Deferrable:
			ctx.WriteString(" DEFERRABLE")
			switch node.UniqueInitially {
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
	if node.HasFKConstraint() {
		if node.References.ConstraintName != "" {
			ctx.WriteString(" CONSTRAINT ")
			ctx.FormatNode(&node.References.ConstraintName)
		}
		ctx.WriteString(" REFERENCES ")
		ctx.FormatNode(node.References.Table)
		if node.References.Col != "" {
			ctx.WriteString(" (")
			ctx.FormatNode(&node.References.Col)
			ctx.WriteByte(')')
		}
		if node.References.Match != MatchSimple {
			ctx.WriteByte(' ')
			ctx.WriteString(node.References.Match.String())
		}
		ctx.FormatNode(&node.References.Actions)
		switch node.UniqueDeferrable {
		case Deferrable:
			ctx.WriteString(" DEFERRABLE")
			switch node.UniqueInitially {
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
}

func (node *ColumnTableDef) columnTypeString() string {
	if node.IsSerial {
		// Map INT types to SERIAL keyword.
		// TODO (rohany): This should be pushed until type resolution occurs.
		//  However, the argument is that we deal with serial at parse time only,
		//  so we handle those cases here.
		switch MustBeStaticallyKnownType(node.Type).Width() {
		case 16:
			return "SERIAL2"
		case 32:
			return "SERIAL4"
		}
		return "SERIAL8"
	}
	return node.Type.SQLString()
}

// String implements the fmt.Stringer interface.
func (node *ColumnTableDef) String() string { return AsString(node) }

// NamedColumnQualification wraps a NamedColumnQualification with a name.
type NamedColumnQualification struct {
	Name          Name
	Qualification ColumnQualification
	Collation     string
	Deferrable    DeferrableMode
	Initially     InitiallyMode
}

type InitiallyMode int

// InitiallyMode values.
const (
	UnspecifiedInitiallyMode InitiallyMode = iota
	InitiallyDeferred
	InitiallyImmediate
)

// ColumnQualification represents a constraint on a column.
type ColumnQualification interface {
	columnQualification()
}

func (NotNullConstraint) columnQualification()      {}
func (NullConstraint) columnQualification()         {}
func (*ColumnCheckConstraint) columnQualification() {}
func (*ColumnDefault) columnQualification()         {}
func (*ColumnComputedDef) columnQualification()     {}
func (UniqueConstraint) columnQualification()       {}
func (*ColumnFKConstraint) columnQualification()    {}

// NotNullConstraint represents NOT NULL on a column.
type NotNullConstraint struct{}

// NullConstraint represents NULL on a column.
type NullConstraint struct{}

// ColumnCheckConstraint represents either a check on a column.
type ColumnCheckConstraint struct {
	Expr      Expr
	NoInherit bool
}

// ColumnDefault represents a DEFAULT clause for a column.
type ColumnDefault struct {
	Expr Expr
}

// ColumnComputedDef represents the description of a computed column (GENERATED ... clause).
type ColumnComputedDef struct {
	Expr      Expr
	ByDefault bool
	Options   SequenceOptions
}

// UniqueConstraint represents UNIQUE and PRIMARY KEY on a column.
type UniqueConstraint struct {
	NullsDistinct bool
	IndexParams   IndexParams
	IsPrimary     bool
}

// ColumnFKConstraint represents a Foreign Key constaint on a column (REFERENCES ... clause).
type ColumnFKConstraint struct {
	Table   TableName
	Col     Name // empty-string means use PK
	Actions ReferenceActions
	Match   CompositeKeyMatchMethod
}

// IndexParams is sub-clause used in UNIQUE, PRIMARY KEY, and EXCLUDE constraints.
type IndexParams struct {
	IncludeColumns IndexElemList // names only
	StorageParams  StorageParams
	Tablespace     Name
}

func (node *IndexParams) Format(ctx *FmtCtx) {
	if node.IncludeColumns != nil {
		ctx.WriteString(" INCLUDE ( ")
		ctx.FormatNode(&node.IncludeColumns)
		ctx.WriteString(" )")
	}
	if node.StorageParams != nil {
		ctx.WriteString(" WITH ( ")
		ctx.FormatNode(&node.IncludeColumns)
		ctx.WriteString(" )")
	}
	if node.Tablespace != "" {
		ctx.WriteString(" USING INDEX TABLESPACE ")
		ctx.FormatNode(&node.Tablespace)
	}
}

// IndexTableDef represents an index definition within a CREATE TABLE
// statement.
type IndexTableDef struct {
	Name        Name
	Columns     IndexElemList
	IndexParams IndexParams
}

// Format implements the NodeFormatter interface.
func (node *IndexTableDef) Format(ctx *FmtCtx) {
	if node.Name != "" {
		ctx.WriteString("CONSTRAINT ")
		ctx.FormatNode(&node.Name)
		ctx.WriteByte(' ')
	}
	ctx.WriteByte('(')
	ctx.FormatNode(&node.Columns)
	ctx.WriteByte(')')
	ctx.FormatNode(&node.IndexParams)
}

// ConstraintTableDef represents a table constraint definition within a CREATE TABLE
// statement.
type ConstraintTableDef interface {
	TableDef
	// Placeholder function to ensure that only desired types
	// (*ConstraintTableDef) conform to the ConstraintTableDef interface.
	constraintTableDef()

	// SetName replaces the name of the definition in-place. Used in the parser.
	SetName(name Name)
}

func (*CheckConstraintTableDef) constraintTableDef()      {}
func (*UniqueConstraintTableDef) constraintTableDef()     {}
func (*ExcludeConstraintTableDef) constraintTableDef()    {}
func (*ForeignKeyConstraintTableDef) constraintTableDef() {}

// CheckConstraintTableDef represents a check constraint within a CREATE
// TABLE statement.
type CheckConstraintTableDef struct {
	Name      Name
	Expr      Expr
	NoInherit bool
}

// SetName implements the ConstraintTableDef interface.
func (node *CheckConstraintTableDef) SetName(name Name) {
	node.Name = name
}

// Format implements the NodeFormatter interface.
func (node *CheckConstraintTableDef) Format(ctx *FmtCtx) {
	if node.Name != "" {
		ctx.WriteString("CONSTRAINT ")
		ctx.FormatNode(&node.Name)
		ctx.WriteByte(' ')
	}
	ctx.WriteString("CHECK (")
	ctx.FormatNode(node.Expr)
	ctx.WriteByte(')')
	if node.NoInherit {
		ctx.WriteString(" NO INHERIT")
	}
}

// UniqueConstraintTableDef represents a unique constraint within a CREATE
// TABLE statement.
type UniqueConstraintTableDef struct {
	IndexTableDef
	NullsNotDistinct bool
	PrimaryKey       bool
}

// SetName implements the TableDef interface.
func (node *UniqueConstraintTableDef) SetName(name Name) {
	node.Name = name
}

// Format implements the NodeFormatter interface.
func (node *UniqueConstraintTableDef) Format(ctx *FmtCtx) {
	if node.Name != "" {
		ctx.WriteString("CONSTRAINT ")
		ctx.FormatNode(&node.Name)
		ctx.WriteByte(' ')
	}
	if node.PrimaryKey {
		ctx.WriteString("PRIMARY KEY ")
	} else {
		ctx.WriteString("UNIQUE ")
	}
	ctx.WriteByte('(')
	ctx.FormatNode(&node.Columns)
	ctx.WriteByte(')')
	ctx.FormatNode(&node.IndexParams)
}

type ExcludeElement struct {
	Elem IndexElem
	With Operator
}

// ExcludeConstraintTableDef represents a FOREIGN KEY constraint in the AST.
type ExcludeConstraintTableDef struct {
	IndexTableDef
	Using     string
	Predicate Expr
}

// Format implements the NodeFormatter interface.
func (node *ExcludeConstraintTableDef) Format(ctx *FmtCtx) {
	if node.Name != "" {
		ctx.WriteString("CONSTRAINT ")
		ctx.FormatNode(&node.Name)
		ctx.WriteByte(' ')
	}
	ctx.WriteString("EXCLUDE ")
	if node.Using != "" {
		ctx.WriteString("USING ")
		ctx.WriteString(node.Using)
	}
	ctx.WriteString(" ( ")
	ctx.FormatNode(&node.Columns)
	ctx.WriteString(" )")
	ctx.FormatNode(&node.IndexParams)
	if node.Predicate != nil {
		ctx.WriteByte(' ')
		ctx.FormatNode(node.Predicate)
	}
}

// SetName implements the ConstraintTableDef interface.
func (node *ExcludeConstraintTableDef) SetName(name Name) {
	node.Name = name
}

type RefAction struct {
	Action  ReferenceAction
	Columns NameList // used for SET NULL or SET DEFAULT
}

func (node *RefAction) Format(ctx *FmtCtx) {
	ctx.WriteString(node.Action.String())
	if node.Columns != nil {
		ctx.WriteByte(' ')
		ctx.FormatNode(&node.Columns)
	}
}

// ReferenceAction is the method used to maintain referential integrity through
// foreign keys.
type ReferenceAction int

// The values for ReferenceAction.
const (
	NoAction ReferenceAction = iota
	Restrict
	SetNull
	SetDefault
	Cascade
)

var referenceActionName = [...]string{
	NoAction:   "NO ACTION",
	Restrict:   "RESTRICT",
	SetNull:    "SET NULL",
	SetDefault: "SET DEFAULT",
	Cascade:    "CASCADE",
}

func (ra ReferenceAction) String() string {
	return referenceActionName[ra]
}

// ReferenceActions contains the actions specified to maintain referential
// integrity through foreign keys for different operations.
type ReferenceActions struct {
	Delete RefAction
	Update RefAction
}

// Format implements the NodeFormatter interface.
func (node *ReferenceActions) Format(ctx *FmtCtx) {
	if node.Delete.Action != NoAction {
		ctx.WriteString(" ON DELETE ")
		ctx.FormatNode(&node.Delete)
	}
	if node.Update.Action != NoAction {
		ctx.WriteString(" ON UPDATE ")
		ctx.FormatNode(&node.Update)
	}
}

// CompositeKeyMatchMethod is the algorithm use when matching composite keys.
// See https://github.com/cockroachdb/cockroach/issues/20305 or
// https://www.postgresql.org/docs/11/sql-createtable.html for details on the
// different composite foreign key matching methods.
type CompositeKeyMatchMethod int

// The values for CompositeKeyMatchMethod.
const (
	MatchSimple CompositeKeyMatchMethod = iota
	MatchFull
	MatchPartial // Note: PARTIAL not actually supported at this point.
)

var compositeKeyMatchMethodName = [...]string{
	MatchSimple:  "MATCH SIMPLE",
	MatchFull:    "MATCH FULL",
	MatchPartial: "MATCH PARTIAL",
}

func (c CompositeKeyMatchMethod) String() string {
	return compositeKeyMatchMethodName[c]
}

// ForeignKeyConstraintTableDef represents a FOREIGN KEY constraint in the AST.
type ForeignKeyConstraintTableDef struct {
	Name     Name
	FromCols NameList
	Table    TableName
	ToCols   NameList
	Actions  ReferenceActions
	Match    CompositeKeyMatchMethod
}

// Format implements the NodeFormatter interface.
func (node *ForeignKeyConstraintTableDef) Format(ctx *FmtCtx) {
	if node.Name != "" {
		ctx.WriteString("CONSTRAINT ")
		ctx.FormatNode(&node.Name)
		ctx.WriteByte(' ')
	}
	ctx.WriteString("FOREIGN KEY (")
	ctx.FormatNode(&node.FromCols)
	ctx.WriteString(") REFERENCES ")
	ctx.FormatNode(&node.Table)

	if len(node.ToCols) > 0 {
		ctx.WriteByte(' ')
		ctx.WriteByte('(')
		ctx.FormatNode(&node.ToCols)
		ctx.WriteByte(')')
	}

	if node.Match != MatchSimple {
		ctx.WriteByte(' ')
		ctx.WriteString(node.Match.String())
	}

	ctx.FormatNode(&node.Actions)
}

// SetName implements the ConstraintTableDef interface.
func (node *ForeignKeyConstraintTableDef) SetName(name Name) {
	node.Name = name
}

// PartitionByType is an enum of each type of partitioning (LIST/RANGE).
type PartitionByType string

const (
	// PartitionByList indicates a PARTITION BY LIST clause.
	PartitionByList PartitionByType = "LIST"
	// PartitionByRange indicates a PARTITION BY RANGE clause.
	PartitionByRange PartitionByType = "RANGE"
	// PartitionByHash indicates a PARTITION BY HASH clause.
	PartitionByHash PartitionByType = "HASH"
)

// PartitionBy represents an PARTITION BY definition within a CREATE/ALTER
// TABLE/INDEX statement.
type PartitionBy struct {
	// Exactly one of List or Range or Hash is required to be non-empty.
	Type  PartitionByType
	Elems IndexElemList
}

// Format implements the NodeFormatter interface.
func (node *PartitionBy) Format(ctx *FmtCtx) {
	ctx.WriteString(` PARTITION BY `)
	ctx.WriteString(string(node.Type))
	ctx.WriteString(" ( ")
	for i := range node.Elems {
		if i > 0 {
			ctx.WriteString(", ")
		}
		ctx.FormatNode(&node.Elems[i])
	}
	ctx.WriteString(" )")
}

// StorageParam is a key-value parameter for table storage.
type StorageParam struct {
	Key   Name
	Value Expr
}

// StorageParams is a list of StorageParams.
type StorageParams []StorageParam

// Format implements the NodeFormatter interface.
func (o *StorageParams) Format(ctx *FmtCtx) {
	for i := range *o {
		n := &(*o)[i]
		if i > 0 {
			ctx.WriteString(", ")
		}
		ctx.FormatNode(&n.Key)
		if n.Value != nil {
			ctx.WriteString(` = `)
			ctx.FormatNode(n.Value)
		}
	}
}

// CreateTableOnCommitSetting represents the CREATE TABLE ... ON COMMIT <action>
// parameters.
type CreateTableOnCommitSetting uint32

const (
	// CreateTableOnCommitUnset indicates that ON COMMIT was unset.
	CreateTableOnCommitUnset CreateTableOnCommitSetting = iota
	// CreateTableOnCommitPreserveRows indicates that ON COMMIT PRESERVE ROWS was set.
	CreateTableOnCommitPreserveRows
	// CreateTableOnCommitDeleteRows indicates that ON COMMIT DELETE ROWS was set.
	CreateTableOnCommitDeleteRows
	// CreateTableOnCommitDrop indicates that ON COMMIT DROP was set.
	CreateTableOnCommitDrop
)

// CreateTable represents a CREATE TABLE statement.
type CreateTable struct {
	IfNotExists   bool
	Table         TableName
	Inherits      TableNames
	PartitionBy   *PartitionBy
	Persistence   Persistence
	StorageParams StorageParams
	OnCommit      CreateTableOnCommitSetting
	Using         string
	Tablespace    Name
	// In CREATE...AS queries, Defs represents a list of ColumnTableDefs, one for
	// each column, and a ConstraintTableDef for each constraint on a subset of
	// these columns.
	Defs       TableDefs
	AsSource   *Select
	WithNoData bool
}

// As returns true if this table represents a CREATE TABLE ... AS statement,
// false otherwise.
func (node *CreateTable) As() bool {
	return node.AsSource != nil
}

// AsHasUserSpecifiedPrimaryKey returns true if a CREATE TABLE ... AS statement
// has a PRIMARY KEY constraint specified.
func (node *CreateTable) AsHasUserSpecifiedPrimaryKey() bool {
	if node.As() {
		for _, def := range node.Defs {
			if d, ok := def.(*ColumnTableDef); !ok {
				return false
			} else if d.PrimaryKey.IsPrimaryKey {
				return true
			}
		}
	}
	return false
}

// Format implements the NodeFormatter interface.
func (node *CreateTable) Format(ctx *FmtCtx) {
	ctx.WriteString("CREATE ")
	switch node.Persistence {
	case PersistenceTemporary:
		ctx.WriteString("TEMPORARY ")
	case PersistenceUnlogged:
		ctx.WriteString("UNLOGGED ")
	}
	ctx.WriteString("TABLE ")
	if node.IfNotExists {
		ctx.WriteString("IF NOT EXISTS ")
	}
	ctx.FormatNode(&node.Table)
	node.FormatBody(ctx)
}

// FormatBody formats the "body" of the create table definition - everything
// but the CREATE TABLE tableName part.
func (node *CreateTable) FormatBody(ctx *FmtCtx) {
	if node.As() {
		if len(node.Defs) > 0 {
			ctx.WriteString(" (")
			ctx.FormatNode(&node.Defs)
			ctx.WriteByte(')')
		}
		if node.Using != "" {
			ctx.WriteString(" USING ")
			ctx.WriteString(node.Using)
		}
		if node.StorageParams != nil {
			ctx.FormatNode(&node.StorageParams)
		}
		switch node.OnCommit {
		case CreateTableOnCommitPreserveRows:
			ctx.WriteString(" ON COMMIT PRESERVE ROWS")
		case CreateTableOnCommitDeleteRows:
			ctx.WriteString(" ON COMMIT DELETE ROWS")
		case CreateTableOnCommitDrop:
			ctx.WriteString(" ON COMMIT DROP")
		}
		if node.Tablespace != "" {
			ctx.WriteString(" TABLESPACE ")
			ctx.FormatNode(&node.Tablespace)
		}
		ctx.WriteString(" AS ")
		ctx.FormatNode(node.AsSource)
		if node.WithNoData {
			ctx.WriteString(" WITH NO DATA")
		}
	} else {
		ctx.WriteString(" (")
		ctx.FormatNode(&node.Defs)
		ctx.WriteByte(')')

		if node.PartitionBy != nil {
			ctx.FormatNode(node.PartitionBy)
		}
		// No storage parameters are implemented, so we never list the storage
		// parameters in the output format.
	}
}

// HoistConstraints finds column check and foreign key constraints defined
// inline with their columns and makes them table-level constraints, stored in
// n.Defs. For example, the foreign key constraint in
//
//	CREATE TABLE foo (a INT REFERENCES bar(a))
//
// gets pulled into a top-level constraint like:
//
//	CREATE TABLE foo (a INT, FOREIGN KEY (a) REFERENCES bar(a))
//
// Similarly, the CHECK constraint in
//
//	CREATE TABLE foo (a INT CHECK (a < 1), b INT)
//
// gets pulled into a top-level constraint like:
//
//	CREATE TABLE foo (a INT, b INT, CHECK (a < 1))
//
// Note that some SQL databases require that a constraint attached to a column
// to refer only to the column it is attached to. We follow Postgres' behavior,
// however, in omitting this restriction by blindly hoisting all column
// constraints. For example, the following table definition is accepted in
// CockroachDB and Postgres, but not necessarily other SQL databases:
//
//	CREATE TABLE foo (a INT CHECK (a < b), b INT)
//
// Unique constraints are not hoisted.
func (node *CreateTable) HoistConstraints() {
	for _, d := range node.Defs {
		if col, ok := d.(*ColumnTableDef); ok {
			for _, checkExpr := range col.CheckExprs {
				node.Defs = append(node.Defs,
					&CheckConstraintTableDef{
						Expr: checkExpr.Expr,
						Name: checkExpr.ConstraintName,
					},
				)
			}
			col.CheckExprs = nil
			if col.HasFKConstraint() {
				var targetCol NameList
				if col.References.Col != "" {
					targetCol = append(targetCol, col.References.Col)
				}
				node.Defs = append(node.Defs, &ForeignKeyConstraintTableDef{
					Table:    *col.References.Table,
					FromCols: NameList{col.Name},
					ToCols:   targetCol,
					Name:     col.References.ConstraintName,
					Actions:  col.References.Actions,
					Match:    col.References.Match,
				})
				col.References.Table = nil
			}
		}
	}
}

// LikeTableDef represents a LIKE table declaration on a CREATE TABLE statement.
type LikeTableDef struct {
	Name    TableName
	Options []LikeTableOption
}

// LikeTableOption represents an individual INCLUDING / EXCLUDING statement
// on a LIKE table declaration.
type LikeTableOption struct {
	Excluded bool
	Opt      LikeTableOpt
}

// Format implements the NodeFormatter interface.
func (def *LikeTableDef) Format(ctx *FmtCtx) {
	ctx.WriteString("LIKE ")
	ctx.FormatNode(&def.Name)
	for _, o := range def.Options {
		ctx.WriteString(" ")
		ctx.FormatNode(o)
	}
}

// Format implements the NodeFormatter interface.
func (l LikeTableOption) Format(ctx *FmtCtx) {
	if l.Excluded {
		ctx.WriteString("EXCLUDING ")
	} else {
		ctx.WriteString("INCLUDING ")
	}
	ctx.WriteString(l.Opt.String())
}

// LikeTableOpt represents one of the types of things that can be included or
// excluded in a LIKE table declaration. It's a bitmap, where each of the Opt
// values is a single enabled bit in the map.
type LikeTableOpt int

// The values for LikeTableOpt.
const (
	LikeTableOptConstraints LikeTableOpt = 1 << iota
	LikeTableOptDefaults
	LikeTableOptGenerated
	LikeTableOptIndexes

	// Make sure this field stays last!
	likeTableOptInvalid
)

// LikeTableOptAll is the full LikeTableOpt bitmap.
const LikeTableOptAll = ^likeTableOptInvalid

// Has returns true if the receiver has the other options bits set.
func (o LikeTableOpt) Has(other LikeTableOpt) bool {
	return int(o)&int(other) != 0
}

func (o LikeTableOpt) String() string {
	switch o {
	case LikeTableOptConstraints:
		return "CONSTRAINTS"
	case LikeTableOptDefaults:
		return "DEFAULTS"
	case LikeTableOptGenerated:
		return "GENERATED"
	case LikeTableOptIndexes:
		return "INDEXES"
	case LikeTableOptAll:
		return "ALL"
	default:
		panic("unknown like table opt" + strconv.Itoa(int(o)))
	}
}

// ToRoleOptions converts KVOptions to a roleoption.List using
// typeAsString to convert exprs to strings.
func (o KVOptions) ToRoleOptions(
	typeAsStringOrNull func(e Expr, op string) (func() (bool, string, error), error), op string,
) (roleoption.List, error) {
	roleOptions := make(roleoption.List, len(o))

	for i, ro := range o {
		option, err := roleoption.ToOption(ro.Key.String())
		if err != nil {
			return nil, err
		}

		if ro.Value != nil {
			if ro.Value == DNull {
				roleOptions[i] = roleoption.RoleOption{
					Option: option, HasValue: true, Value: func() (bool, string, error) {
						return true, "", nil
					},
				}
			} else {
				strFn, err := typeAsStringOrNull(ro.Value, op)
				if err != nil {
					return nil, err
				}

				if err != nil {
					return nil, err
				}
				roleOptions[i] = roleoption.RoleOption{
					Option: option, Value: strFn, HasValue: true,
				}
			}
		} else {
			roleOptions[i] = roleoption.RoleOption{
				Option: option, HasValue: false,
			}
		}
	}

	return roleOptions, nil
}

func (o *KVOptions) formatAsRoleOptions(ctx *FmtCtx) {
	for _, option := range *o {
		ctx.WriteString(" ")
		ctx.WriteString(
			// "_" replaces space (" ") in YACC for handling tree.Name formatting.
			strings.ReplaceAll(
				strings.ToUpper(option.Key.String()), "_", " "),
		)

		// Password is a special case.
		if strings.ToUpper(option.Key.String()) == "PASSWORD" {
			ctx.WriteString(" ")
			if ctx.flags.HasFlags(FmtShowPasswords) {
				ctx.FormatNode(option.Value)
			} else {
				ctx.WriteString("*****")
			}
		} else if option.Value == DNull {
			ctx.WriteString(" ")
			ctx.FormatNode(option.Value)
		} else if option.Value != nil {
			ctx.WriteString(" ")
			ctx.FormatNode(option.Value)
		}
	}
}
