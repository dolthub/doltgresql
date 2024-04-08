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

// DropBehavior represents options for dropping schema elements.
type DropBehavior int

// DropBehavior values.
const (
	DropDefault DropBehavior = iota
	DropRestrict
	DropCascade
)

var dropBehaviorName = [...]string{
	DropDefault:  "",
	DropRestrict: "RESTRICT",
	DropCascade:  "CASCADE",
}

func (d DropBehavior) String() string {
	return dropBehaviorName[d]
}

var _ Statement = &DropAggregate{}

// DropAggregate represents a DROP AGGREGATE statement.
type DropAggregate struct {
	Aggregates   []AggregateToDrop
	IfExists     bool
	DropBehavior DropBehavior
}

type AggregateToDrop struct {
	Name   *UnresolvedObjectName
	AggSig *AggregateSignature
}

// Format implements the NodeFormatter interface.
func (node *DropAggregate) Format(ctx *FmtCtx) {
	ctx.WriteString("DROP AGGREGATE ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	for i, agg := range node.Aggregates {
		if i != 0 {
			ctx.WriteString(" , ")
		}
		ctx.FormatNode(agg.Name)
		ctx.WriteString(" ( ")
		ctx.FormatNode(agg.AggSig)
		ctx.WriteString(" ) ")
	}
	switch node.DropBehavior {
	case DropDefault:
	default:
		ctx.WriteByte(' ')
		ctx.WriteString(dropBehaviorName[node.DropBehavior])
	}
}

var _ Statement = &DropDatabase{}

// DropDatabase represents a DROP DATABASE statement.
type DropDatabase struct {
	Name     Name
	IfExists bool
	Force    bool
}

// Format implements the NodeFormatter interface.
func (node *DropDatabase) Format(ctx *FmtCtx) {
	ctx.WriteString("DROP DATABASE ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	ctx.FormatNode(&node.Name)
	if node.Force {
		ctx.WriteString(" WITH ( FORCE )")
	}
}

var _ Statement = &DropDomain{}

// DropDomain represents a DROP DOMAIN statement.
type DropDomain struct {
	Names        NameList
	IfExists     bool
	DropBehavior DropBehavior
}

// Format implements the NodeFormatter interface.
func (node *DropDomain) Format(ctx *FmtCtx) {
	ctx.WriteString("DROP DATABASE ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	ctx.FormatNode(&node.Names)
	switch node.DropBehavior {
	case DropDefault:
	default:
		ctx.WriteByte(' ')
		ctx.WriteString(dropBehaviorName[node.DropBehavior])
	}
}

var _ Statement = &DropExtension{}

// DropExtension represents a DROP EXTENSION statement.
type DropExtension struct {
	Names        NameList
	IfExists     bool
	DropBehavior DropBehavior
}

// Format implements the NodeFormatter interface.
func (node *DropExtension) Format(ctx *FmtCtx) {
	ctx.WriteString("DROP EXTENSION ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	ctx.FormatNode(&node.Names)
	switch node.DropBehavior {
	case DropDefault:
	default:
		ctx.WriteByte(' ')
		ctx.WriteString(dropBehaviorName[node.DropBehavior])
	}
}

var _ Statement = &DropFunction{}

// DropFunction represents a DROP FUNCTION statement.
type DropFunction struct {
	Functions    []RoutineWithArgs
	IfExists     bool
	DropBehavior DropBehavior
}

// Format implements the NodeFormatter interface.
func (node *DropFunction) Format(ctx *FmtCtx) {
	ctx.WriteString("DROP FUNCTION ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	for i, f := range node.Functions {
		if i != 0 {
			ctx.WriteString(", ")
		}
		ctx.FormatNode(&f)
	}
	switch node.DropBehavior {
	case DropDefault:
	default:
		ctx.WriteByte(' ')
		ctx.WriteString(dropBehaviorName[node.DropBehavior])
	}
}

// RoutineWithArgs represents the routine name and its arguments, if any, for DROP { FUNCTION | PROCEDURE } statement.
type RoutineWithArgs struct {
	Name *UnresolvedObjectName
	Args RoutineArgs
}

func (node *RoutineWithArgs) Format(ctx *FmtCtx) {
	ctx.FormatNode(node.Name)
	ctx.WriteString(" (")
	if len(node.Args) != 0 {
		node.Args.Format(ctx)
	}
	ctx.WriteString(" )")
}

var _ Statement = &DropIndex{}

// DropIndex represents a DROP INDEX statement.
type DropIndex struct {
	IndexList    TableIndexNames
	IfExists     bool
	DropBehavior DropBehavior
	Concurrently bool
}

// Format implements the NodeFormatter interface.
func (node *DropIndex) Format(ctx *FmtCtx) {
	ctx.WriteString("DROP INDEX ")
	if node.Concurrently {
		ctx.WriteString("CONCURRENTLY ")
	}
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	ctx.FormatNode(&node.IndexList)
	if node.DropBehavior != DropDefault {
		ctx.WriteByte(' ')
		ctx.WriteString(node.DropBehavior.String())
	}
}

var _ Statement = &DropLanguage{}

// DropLanguage represents a DROP LANGUAGE statement.
type DropLanguage struct {
	Name         Name
	Procedural   bool
	IfExists     bool
	DropBehavior DropBehavior
}

// Format implements the NodeFormatter interface.
func (node *DropLanguage) Format(ctx *FmtCtx) {
	ctx.WriteString("DROP ")
	if node.Procedural {
		ctx.WriteString("PROCEDURAL ")
	}
	ctx.WriteString("LANGUAGE ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	ctx.FormatNode(&node.Name)
	if node.DropBehavior != DropDefault {
		ctx.WriteByte(' ')
		ctx.WriteString(node.DropBehavior.String())
	}
}

var _ Statement = &DropProcedure{}

// DropProcedure represents a DROP PROCEDURE statement.
type DropProcedure struct {
	Procedures   []RoutineWithArgs
	IfExists     bool
	DropBehavior DropBehavior
}

// Format implements the NodeFormatter interface.
func (node *DropProcedure) Format(ctx *FmtCtx) {
	ctx.WriteString("DROP PROCEDURE ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	for i, f := range node.Procedures {
		if i != 0 {
			ctx.WriteString(", ")
		}
		ctx.FormatNode(&f)
	}
	switch node.DropBehavior {
	case DropDefault:
	default:
		ctx.WriteByte(' ')
		ctx.WriteString(dropBehaviorName[node.DropBehavior])
	}
}

var _ Statement = &DropTable{}

// DropTable represents a DROP TABLE statement.
type DropTable struct {
	Names        TableNames
	IfExists     bool
	DropBehavior DropBehavior
}

// Format implements the NodeFormatter interface.
func (node *DropTable) Format(ctx *FmtCtx) {
	ctx.WriteString("DROP TABLE ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	ctx.FormatNode(&node.Names)
	if node.DropBehavior != DropDefault {
		ctx.WriteByte(' ')
		ctx.WriteString(node.DropBehavior.String())
	}
}

var _ Statement = &DropTrigger{}

// DropTrigger represents a DROP TRIGGER statement.
type DropTrigger struct {
	Name         Name
	IfExists     bool
	OnTable      TableName
	DropBehavior DropBehavior
}

// Format implements the NodeFormatter interface.
func (node *DropTrigger) Format(ctx *FmtCtx) {
	ctx.WriteString("DROP TRIGGER ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	ctx.FormatNode(&node.Name)
	ctx.WriteString(" ON ")
	ctx.FormatNode(&node.OnTable)
	switch node.DropBehavior {
	case DropDefault:
	default:
		ctx.WriteByte(' ')
		ctx.WriteString(dropBehaviorName[node.DropBehavior])
	}
}

var _ Statement = &DropView{}

// DropView represents a DROP VIEW statement.
type DropView struct {
	Names          TableNames
	IfExists       bool
	DropBehavior   DropBehavior
	IsMaterialized bool
}

// Format implements the NodeFormatter interface.
func (node *DropView) Format(ctx *FmtCtx) {
	ctx.WriteString("DROP ")
	if node.IsMaterialized {
		ctx.WriteString("MATERIALIZED ")
	}
	ctx.WriteString("VIEW ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	ctx.FormatNode(&node.Names)
	if node.DropBehavior != DropDefault {
		ctx.WriteByte(' ')
		ctx.WriteString(node.DropBehavior.String())
	}
}

var _ Statement = &DropSequence{}

// DropSequence represents a DROP SEQUENCE statement.
type DropSequence struct {
	Names        TableNames
	IfExists     bool
	DropBehavior DropBehavior
}

// Format implements the NodeFormatter interface.
func (node *DropSequence) Format(ctx *FmtCtx) {
	ctx.WriteString("DROP SEQUENCE ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	ctx.FormatNode(&node.Names)
	if node.DropBehavior != DropDefault {
		ctx.WriteByte(' ')
		ctx.WriteString(node.DropBehavior.String())
	}
}

var _ Statement = &DropRole{}

// DropRole represents a DROP ROLE statement
type DropRole struct {
	Names    Exprs
	IsRole   bool
	IfExists bool
}

// Format implements the NodeFormatter interface.
func (node *DropRole) Format(ctx *FmtCtx) {
	ctx.WriteString("DROP")
	if node.IsRole {
		ctx.WriteString(" ROLE ")
	} else {
		ctx.WriteString(" USER ")
	}
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	ctx.FormatNode(&node.Names)
}

// DropType represents a DROP TYPE command.
type DropType struct {
	Names        []*UnresolvedObjectName
	IfExists     bool
	DropBehavior DropBehavior
}

var _ Statement = &DropType{}

// Format implements the NodeFormatter interface.
func (node *DropType) Format(ctx *FmtCtx) {
	ctx.WriteString("DROP TYPE ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	for i := range node.Names {
		if i > 0 {
			ctx.WriteString(", ")
		}
		ctx.FormatNode(node.Names[i])
	}
	if node.DropBehavior != DropDefault {
		ctx.WriteByte(' ')
		ctx.WriteString(node.DropBehavior.String())
	}
}

// DropSchema represents a DROP SCHEMA command.
type DropSchema struct {
	Names        []string
	IfExists     bool
	DropBehavior DropBehavior
}

var _ Statement = &DropSchema{}

// Format implements the NodeFormatter interface.
func (node *DropSchema) Format(ctx *FmtCtx) {
	ctx.WriteString("DROP SCHEMA ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	for i := range node.Names {
		if i > 0 {
			ctx.WriteString(", ")
		}
		ctx.FormatNameP(&node.Names[i])
	}
	if node.DropBehavior != DropDefault {
		ctx.WriteString(" ")
		ctx.WriteString(node.DropBehavior.String())
	}
}
