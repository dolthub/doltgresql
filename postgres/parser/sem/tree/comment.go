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

package tree

import "github.com/dolthub/doltgresql/postgres/parser/lex"

var _ Statement = &Comment{}

// Comment represents a COMMENT statement.
type Comment struct {
	Object  CommentObject
	Comment *string
}

// Format implements the NodeFormatter interface.
func (node *Comment) Format(ctx *FmtCtx) {
	ctx.WriteString("COMMENT ON ")
	ctx.FormatNode(node.Object)
	ctx.WriteString(" IS ")
	if node.Comment != nil {
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, *node.Comment, ctx.flags.EncodeFlags())
	} else {
		ctx.WriteString("NULL")
	}
}

// CommentObject represents an object type definition for COMMENT ON command.
type CommentObject interface {
	NodeFormatter
	commentObject()
}

func (*CommentOnAccessMethod) commentObject()            {}
func (*CommentOnAggregate) commentObject()               {}
func (*CommentOnCast) commentObject()                    {}
func (*CommentOnCollation) commentObject()               {}
func (*CommentOnColumn) commentObject()                  {}
func (*CommentOnConstraintOnTable) commentObject()       {}
func (*CommentOnConstraintOnDomain) commentObject()      {}
func (*CommentOnConversion) commentObject()              {}
func (*CommentOnDatabase) commentObject()                {}
func (*CommentOnDomain) commentObject()                  {}
func (*CommentOnExtension) commentObject()               {}
func (*CommentOnEventTrigger) commentObject()            {}
func (*CommentOnForeignDataWrapper) commentObject()      {}
func (*CommentOnForeignTable) commentObject()            {}
func (*CommentOnFunction) commentObject()                {}
func (*CommentOnIndex) commentObject()                   {}
func (*CommentOnLargeObject) commentObject()             {}
func (*CommentOnMaterializedView) commentObject()        {}
func (*CommentOnOperator) commentObject()                {}
func (*CommentOnOperatorClass) commentObject()           {}
func (*CommentOnOperatorFamily) commentObject()          {}
func (*CommentOnPolicy) commentObject()                  {}
func (*CommentOnLanguage) commentObject()                {}
func (*CommentOnProcedure) commentObject()               {}
func (*CommentOnPublication) commentObject()             {}
func (*CommentOnRole) commentObject()                    {}
func (*CommentOnRoutine) commentObject()                 {}
func (*CommentOnRule) commentObject()                    {}
func (*CommentOnSchema) commentObject()                  {}
func (*CommentOnSequence) commentObject()                {}
func (*CommentOnServer) commentObject()                  {}
func (*CommentOnStatistics) commentObject()              {}
func (*CommentOnSubscription) commentObject()            {}
func (*CommentOnTable) commentObject()                   {}
func (*CommentOnTablespace) commentObject()              {}
func (*CommentOnTextSearchConfiguration) commentObject() {}
func (*CommentOnTextSearchDictionary) commentObject()    {}
func (*CommentOnTextSearchParser) commentObject()        {}
func (*CommentOnTextSearchTemplate) commentObject()      {}
func (*CommentOnTransformFor) commentObject()            {}
func (*CommentOnTrigger) commentObject()                 {}
func (*CommentOnType) commentObject()                    {}
func (*CommentOnView) commentObject()                    {}

var _ CommentObject = &CommentOnAccessMethod{}
var _ CommentObject = &CommentOnAggregate{}
var _ CommentObject = &CommentOnCast{}
var _ CommentObject = &CommentOnCollation{}
var _ CommentObject = &CommentOnColumn{}
var _ CommentObject = &CommentOnConstraintOnTable{}
var _ CommentObject = &CommentOnConstraintOnDomain{}
var _ CommentObject = &CommentOnConversion{}
var _ CommentObject = &CommentOnDatabase{}
var _ CommentObject = &CommentOnDomain{}
var _ CommentObject = &CommentOnExtension{}
var _ CommentObject = &CommentOnEventTrigger{}
var _ CommentObject = &CommentOnForeignDataWrapper{}
var _ CommentObject = &CommentOnForeignTable{}
var _ CommentObject = &CommentOnFunction{}
var _ CommentObject = &CommentOnIndex{}
var _ CommentObject = &CommentOnLargeObject{}
var _ CommentObject = &CommentOnMaterializedView{}
var _ CommentObject = &CommentOnOperator{}
var _ CommentObject = &CommentOnOperatorClass{}
var _ CommentObject = &CommentOnOperatorFamily{}
var _ CommentObject = &CommentOnPolicy{}
var _ CommentObject = &CommentOnLanguage{}
var _ CommentObject = &CommentOnProcedure{}
var _ CommentObject = &CommentOnPublication{}
var _ CommentObject = &CommentOnRole{}
var _ CommentObject = &CommentOnRoutine{}
var _ CommentObject = &CommentOnRule{}
var _ CommentObject = &CommentOnSchema{}
var _ CommentObject = &CommentOnSequence{}
var _ CommentObject = &CommentOnServer{}
var _ CommentObject = &CommentOnStatistics{}
var _ CommentObject = &CommentOnSubscription{}
var _ CommentObject = &CommentOnTable{}
var _ CommentObject = &CommentOnTablespace{}
var _ CommentObject = &CommentOnTextSearchConfiguration{}
var _ CommentObject = &CommentOnTextSearchDictionary{}
var _ CommentObject = &CommentOnTextSearchParser{}
var _ CommentObject = &CommentOnTextSearchTemplate{}
var _ CommentObject = &CommentOnTransformFor{}
var _ CommentObject = &CommentOnTrigger{}
var _ CommentObject = &CommentOnType{}
var _ CommentObject = &CommentOnView{}

// CommentOnAccessMethod represents COMMENT ON ACCESS METHOD command.
type CommentOnAccessMethod struct {
	Name *UnresolvedObjectName
}

// Format implements the NodeFormatter interface.
func (node *CommentOnAccessMethod) Format(ctx *FmtCtx) {
	ctx.WriteString("ACCESS METHOD ")
	ctx.FormatNode(node.Name)
}

// CommentOnAggregate represents COMMENT ON AGGREGATE command.
type CommentOnAggregate struct {
	Name   Name
	AggSig *AggregateSignature
}

// Format implements the NodeFormatter interface.
func (node *CommentOnAggregate) Format(ctx *FmtCtx) {
	ctx.WriteString("AGGREGATE ")
	ctx.FormatNode(&node.Name)
	ctx.WriteString(" ( ")
	ctx.FormatNode(node.AggSig)
	ctx.WriteString(" )")
}

// CommentOnCast represents COMMENT ON CAST command.
type CommentOnCast struct {
	Source ResolvableTypeReference
	Target ResolvableTypeReference
}

// Format implements the NodeFormatter interface.
func (node *CommentOnCast) Format(ctx *FmtCtx) {
	ctx.WriteString("CAST ( ")
	ctx.WriteString(node.Source.SQLString())
	ctx.WriteString(" AS ")
	ctx.WriteString(node.Target.SQLString())
	ctx.WriteString(" )")
}

// CommentOnCollation represents COMMENT ON COLLATION command.
type CommentOnCollation struct {
	Name string
}

// Format implements the NodeFormatter interface.
func (node *CommentOnCollation) Format(ctx *FmtCtx) {
	ctx.WriteString("COLLATION ")
	ctx.WriteString(node.Name)
}

// CommentOnColumn represents COMMENT ON COLUMN command.
type CommentOnColumn struct {
	*ColumnItem
}

// Format implements the NodeFormatter interface.
func (node *CommentOnColumn) Format(ctx *FmtCtx) {
	ctx.WriteString("COLUMN ")
	ctx.FormatNode(node.ColumnItem)
}

// CommentOnConstraintOnTable represents COMMENT ON CONSTRAINT ON table command.
type CommentOnConstraintOnTable struct {
	Constraint Name
	Table      *UnresolvedObjectName
}

// Format implements the NodeFormatter interface.
func (node *CommentOnConstraintOnTable) Format(ctx *FmtCtx) {
	ctx.WriteString("CONSTRAINT ")
	ctx.FormatNode(&node.Constraint)
	ctx.WriteString(" ON ")
	ctx.FormatNode(node.Table)
}

// CommentOnConstraintOnDomain represents COMMENT ON CONSTRAINT ON DOMAIN command.
type CommentOnConstraintOnDomain struct {
	Constraint Name
	Domain     *UnresolvedObjectName
}

// Format implements the NodeFormatter interface.
func (node *CommentOnConstraintOnDomain) Format(ctx *FmtCtx) {
	ctx.WriteString("CONSTRAINT ")
	ctx.FormatNode(&node.Constraint)
	ctx.WriteString(" ON DOMAIN ")
	ctx.FormatNode(node.Domain)
}

// CommentOnConversion represents COMMENT ON CONVERSION command.
type CommentOnConversion struct {
	Name Name
}

// Format implements the NodeFormatter interface.
func (node *CommentOnConversion) Format(ctx *FmtCtx) {
	ctx.WriteString("CONVERSION ")
	ctx.FormatNode(&node.Name)
}

// CommentOnDatabase represents COMMENT ON DATABASE command.
type CommentOnDatabase struct {
	Name Name
}

// Format implements the NodeFormatter interface.
func (node *CommentOnDatabase) Format(ctx *FmtCtx) {
	ctx.WriteString("DATABASE ")
	ctx.FormatNode(&node.Name)
}

// CommentOnDomain represents COMMENT ON DOMAIN command.
type CommentOnDomain struct {
	Name *UnresolvedObjectName
}

// Format implements the NodeFormatter interface.
func (node *CommentOnDomain) Format(ctx *FmtCtx) {
	ctx.WriteString(" DOMAIN ")
	ctx.FormatNode(node.Name)
}

// CommentOnExtension represents COMMENT ON EXTENSION command.
type CommentOnExtension struct {
	Name Name
}

// Format implements the NodeFormatter interface.
func (node *CommentOnExtension) Format(ctx *FmtCtx) {
	ctx.WriteString("EXTENSION ")
	ctx.FormatNode(&node.Name)
}

// CommentOnEventTrigger represents COMMENT ON EVENT TRIGGER command.
type CommentOnEventTrigger struct {
	Name Name
}

// Format implements the NodeFormatter interface.
func (node *CommentOnEventTrigger) Format(ctx *FmtCtx) {
	ctx.WriteString("EVENT TRIGGER ")
	ctx.FormatNode(&node.Name)
}

// CommentOnForeignDataWrapper represents COMMENT ON FOREIGN DATA WRAPPER command.
type CommentOnForeignDataWrapper struct {
	Name Name
}

// Format implements the NodeFormatter interface.
func (node *CommentOnForeignDataWrapper) Format(ctx *FmtCtx) {
	ctx.WriteString("FOREIGN DATA WRAPPER ")
	ctx.FormatNode(&node.Name)
}

// CommentOnForeignTable represents COMMENT ON FOREIGN TABLE command.
type CommentOnForeignTable struct {
	Name Name
}

// Format implements the NodeFormatter interface.
func (node *CommentOnForeignTable) Format(ctx *FmtCtx) {
	ctx.WriteString("FOREIGN TABLE ")
	ctx.FormatNode(&node.Name)
}

// CommentOnFunction represents COMMENT ON FUNCTION command.
type CommentOnFunction struct {
	Name *UnresolvedObjectName
	Args RoutineArgs
}

// Format implements the NodeFormatter interface.
func (node *CommentOnFunction) Format(ctx *FmtCtx) {
	ctx.WriteString("FUNCTION ")
	ctx.FormatNode(node.Name)
	if node.Args != nil {
		ctx.WriteString(" ( ")
		ctx.FormatNode(node.Args)
		ctx.WriteString(" )")
	}
}

// CommentOnIndex represents COMMENT ON INDEX command.
type CommentOnIndex struct {
	Index TableIndexName
}

// Format implements the NodeFormatter interface.
func (node *CommentOnIndex) Format(ctx *FmtCtx) {
	ctx.WriteString("INDEX ")
	ctx.FormatNode(&node.Index)
}

// CommentOnLargeObject represents COMMENT ON LARGE OBJECT command.
type CommentOnLargeObject struct {
	Oid Expr
}

// Format implements the NodeFormatter interface.
func (node *CommentOnLargeObject) Format(ctx *FmtCtx) {
	ctx.WriteString("LARGE OBJECT ")
	ctx.FormatNode(node.Oid)
}

// CommentOnMaterializedView represents COMMENT ON MATERIALIZED VIEW command.
type CommentOnMaterializedView struct {
	Name *UnresolvedObjectName
}

// Format implements the NodeFormatter interface.
func (node *CommentOnMaterializedView) Format(ctx *FmtCtx) {
	ctx.WriteString("MATERIALIZED VIEW ")
	ctx.FormatNode(node.Name)
}

// CommentOnOperator represents COMMENT ON OPERATOR command.
type CommentOnOperator struct {
	Op    Operator
	Left  ResolvableTypeReference
	Right ResolvableTypeReference
}

// Format implements the NodeFormatter interface.
func (node *CommentOnOperator) Format(ctx *FmtCtx) {
	ctx.WriteString("OPERATOR ")
	switch op := node.Op.(type) {
	case UnaryOperator:
		ctx.WriteString(op.String())
	case BinaryOperator:
		ctx.WriteString(op.String())
	case ComparisonOperator:
		ctx.WriteString(op.String())
	}
	ctx.WriteString(" ( ")
	ctx.WriteString(node.Left.SQLString())
	ctx.WriteString(" , ")
	ctx.WriteString(node.Right.SQLString())
	ctx.WriteString(" )")
}

// CommentOnOperatorClass represents COMMENT ON OPERATOR CLASS command.
type CommentOnOperatorClass struct {
	Name      Name
	IdxMethod string
}

// Format implements the NodeFormatter interface.
func (node *CommentOnOperatorClass) Format(ctx *FmtCtx) {
	ctx.WriteString("OPERATOR CLASS ")
	ctx.FormatNode(&node.Name)
	ctx.WriteString(" USING ")
	ctx.WriteString(node.IdxMethod)
}

// CommentOnOperatorFamily represents COMMENT ON FAMILY command.
type CommentOnOperatorFamily struct {
	Name      Name
	IdxMethod string
}

// Format implements the NodeFormatter interface.
func (node *CommentOnOperatorFamily) Format(ctx *FmtCtx) {
	ctx.WriteString("OPERATOR FAMILY ")
	ctx.FormatNode(&node.Name)
	ctx.WriteString(" USING ")
	ctx.WriteString(node.IdxMethod)
}

// CommentOnPolicy represents COMMENT ON POLICY command.
type CommentOnPolicy struct {
	Policy Name
	Table  *UnresolvedObjectName
}

// Format implements the NodeFormatter interface.
func (node *CommentOnPolicy) Format(ctx *FmtCtx) {
	ctx.WriteString("POLICY ")
	ctx.FormatNode(&node.Policy)
	ctx.WriteString(" ON ")
	ctx.FormatNode(node.Table)
}

// CommentOnLanguage represents COMMENT ON LANGUAGE command.
type CommentOnLanguage struct {
	Name       Name
	Procedural bool
}

// Format implements the NodeFormatter interface.
func (node *CommentOnLanguage) Format(ctx *FmtCtx) {
	if node.Procedural {
		ctx.WriteString("PROCEDURAL ")
	}
	ctx.WriteString("LANGUAGE ")
	ctx.FormatNode(&node.Name)
}

// CommentOnProcedure represents COMMENT ON PROCEDURE command.
type CommentOnProcedure struct {
	Name *UnresolvedObjectName
	Args RoutineArgs
}

// Format implements the NodeFormatter interface.
func (node *CommentOnProcedure) Format(ctx *FmtCtx) {
	ctx.WriteString("PROCEDURE ")
	ctx.FormatNode(node.Name)
	if node.Args != nil {
		ctx.WriteString(" ( ")
		ctx.FormatNode(node.Args)
		ctx.WriteString(" )")
	}
}

// CommentOnPublication represents COMMENT ON PUBLICATION command.
type CommentOnPublication struct {
	Name Name
}

// Format implements the NodeFormatter interface.
func (node *CommentOnPublication) Format(ctx *FmtCtx) {
	ctx.WriteString("PUBLICATION ")
	ctx.FormatNode(&node.Name)
}

// CommentOnRole represents COMMENT ON ROLE command.
type CommentOnRole struct {
	Name string
}

// Format implements the NodeFormatter interface.
func (node *CommentOnRole) Format(ctx *FmtCtx) {
	ctx.WriteString("ROLE ")
	ctx.WriteString(node.Name)
}

// CommentOnRoutine represents COMMENT ON ROUTINE command.
type CommentOnRoutine struct {
	Name *UnresolvedObjectName
	Args RoutineArgs
}

// Format implements the NodeFormatter interface.
func (node *CommentOnRoutine) Format(ctx *FmtCtx) {
	ctx.WriteString("ROUTINE ")
	ctx.FormatNode(node.Name)
	if node.Args != nil {
		ctx.WriteString(" ( ")
		ctx.FormatNode(node.Args)
		ctx.WriteString(" )")
	}
}

// CommentOnRule represents COMMENT ON RULE command.
type CommentOnRule struct {
	Rule  Name
	Table *UnresolvedObjectName
}

// Format implements the NodeFormatter interface.
func (node *CommentOnRule) Format(ctx *FmtCtx) {
	ctx.WriteString("RULE ")
	ctx.FormatNode(&node.Rule)
	ctx.WriteString(" ON ")
	ctx.FormatNode(node.Table)
}

// CommentOnSchema represents COMMENT ON SCHEMA command.
type CommentOnSchema struct {
	Name string
}

// Format implements the NodeFormatter interface.
func (node *CommentOnSchema) Format(ctx *FmtCtx) {
	ctx.WriteString("SCHEMA ")
	ctx.WriteString(node.Name)
}

// CommentOnSequence represents COMMENT ON SEQUENCE command.
type CommentOnSequence struct {
	Name *UnresolvedObjectName
}

// Format implements the NodeFormatter interface.
func (node *CommentOnSequence) Format(ctx *FmtCtx) {
	ctx.WriteString("SEQUENCE ")
	ctx.FormatNode(node.Name)
}

// CommentOnServer represents COMMENT ON SERVER command.
type CommentOnServer struct {
	Name Name
}

// Format implements the NodeFormatter interface.
func (node *CommentOnServer) Format(ctx *FmtCtx) {
	ctx.WriteString("SERVER ")
	ctx.FormatNode(&node.Name)
}

// CommentOnStatistics represents COMMENT ON STATISTICS command.
type CommentOnStatistics struct {
	Name Name
}

// Format implements the NodeFormatter interface.
func (node *CommentOnStatistics) Format(ctx *FmtCtx) {
	ctx.WriteString("STATISTICS ")
	ctx.FormatNode(&node.Name)
}

// CommentOnSubscription represents COMMENT ON SUBSCRIPTION command.
type CommentOnSubscription struct {
	Name Name
}

// Format implements the NodeFormatter interface.
func (node *CommentOnSubscription) Format(ctx *FmtCtx) {
	ctx.WriteString("SUBSCRIPTION ")
	ctx.FormatNode(&node.Name)
}

// CommentOnTable represents COMMENT ON TABLE command.
type CommentOnTable struct {
	Name *UnresolvedObjectName
}

// Format implements the NodeFormatter interface.
func (node *CommentOnTable) Format(ctx *FmtCtx) {
	ctx.WriteString("TABLE ")
	ctx.FormatNode(node.Name)
}

// CommentOnTablespace represents COMMENT ON TABLESPACE command.
type CommentOnTablespace struct {
	Name Name
}

// Format implements the NodeFormatter interface.
func (node *CommentOnTablespace) Format(ctx *FmtCtx) {
	ctx.WriteString("TABLESPACE ")
	ctx.FormatNode(&node.Name)
}

// CommentOnTextSearchConfiguration represents COMMENT ON TEXT SEARCH CONFIGURATION command.
type CommentOnTextSearchConfiguration struct {
	Name Name
}

// Format implements the NodeFormatter interface.
func (node *CommentOnTextSearchConfiguration) Format(ctx *FmtCtx) {
	ctx.WriteString("TEXT SEARCH CONFIGURATION ")
	ctx.FormatNode(&node.Name)
}

// CommentOnTextSearchDictionary represents COMMENT ON TEXT SEARCH DICTIONARY command.
type CommentOnTextSearchDictionary struct {
	Name Name
}

// Format implements the NodeFormatter interface.
func (node *CommentOnTextSearchDictionary) Format(ctx *FmtCtx) {
	ctx.WriteString("TEXT SEARCH DICTIONARY ")
	ctx.FormatNode(&node.Name)
}

// CommentOnTextSearchParser represents COMMENT ON TEXT SEARCH PARSER command.
type CommentOnTextSearchParser struct {
	Name Name
}

// Format implements the NodeFormatter interface.
func (node *CommentOnTextSearchParser) Format(ctx *FmtCtx) {
	ctx.WriteString("TEXT SEARCH PARSER ")
	ctx.FormatNode(&node.Name)
}

// CommentOnTextSearchTemplate represents COMMENT ON TEXT SEARCH TEMPLATE command.
type CommentOnTextSearchTemplate struct {
	Name Name
}

// Format implements the NodeFormatter interface.
func (node *CommentOnTextSearchTemplate) Format(ctx *FmtCtx) {
	ctx.WriteString("TEXT SEARCH TEMPLATE ")
	ctx.FormatNode(&node.Name)
}

// CommentOnTransformFor represents COMMENT ON TRANSFORM FOR command.
type CommentOnTransformFor struct {
	Type     ResolvableTypeReference
	Language Name
}

// Format implements the NodeFormatter interface.
func (node *CommentOnTransformFor) Format(ctx *FmtCtx) {
	ctx.WriteString("TRANSFORM FOR ")
	ctx.WriteString(node.Type.SQLString())
	ctx.WriteString(" LANGUAGE ")
	ctx.FormatNode(&node.Language)
}

// CommentOnTrigger represents COMMENT ON TRIGGER command.
type CommentOnTrigger struct {
	Trigger Name
	Table   *UnresolvedObjectName
}

// Format implements the NodeFormatter interface.
func (node *CommentOnTrigger) Format(ctx *FmtCtx) {
	ctx.WriteString("TRIGGER ")
	ctx.FormatNode(&node.Trigger)
	ctx.WriteString(" ON ")
	ctx.FormatNode(node.Table)
}

// CommentOnType represents COMMENT ON TYPE command.
type CommentOnType struct {
	Name *UnresolvedObjectName
}

// Format implements the NodeFormatter interface.
func (node *CommentOnType) Format(ctx *FmtCtx) {
	ctx.WriteString("TYPE ")
	ctx.FormatNode(node.Name)
}

// CommentOnView represents COMMENT ON VIEW command.
type CommentOnView struct {
	Name *UnresolvedObjectName
}

// Format implements the NodeFormatter interface.
func (node *CommentOnView) Format(ctx *FmtCtx) {
	ctx.WriteString("VIEW ")
	ctx.FormatNode(node.Name)
}
