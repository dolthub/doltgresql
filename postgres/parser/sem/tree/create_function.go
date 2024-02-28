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

import "strings"

var _ Statement = &CreateFunction{}

// CreateFunction represents a CREATE FUNCTION statement.
type CreateFunction struct {
	Name    *UnresolvedObjectName
	Replace bool
	Args    RoutineArgs
	RetType []SimpleColumnDef
	Options []RoutineOption
}

// Format implements the NodeFormatter interface.
func (node *CreateFunction) Format(ctx *FmtCtx) {
	ctx.WriteString("CREATE ")
	if node.Replace {
		ctx.WriteString("OR REPLACE ")
	}
	ctx.WriteString("FUNCTION ")
	ctx.FormatNode(node.Name)
	if len(node.Args) != 0 {
		ctx.WriteString(" (")
		ctx.FormatNode(node.Args)
		ctx.WriteString(" )")
	}
	if node.RetType != nil {
		if len(node.RetType) == 1 && node.RetType[0].Name == "" {
			ctx.WriteString("RETURNS ")
			ctx.WriteString(node.RetType[0].Type.SQLString())
		} else {
			ctx.WriteString("RETURNS TABLE ")
			for i, t := range node.RetType {
				if i != 0 {
					ctx.WriteString(", ")
				}
				ctx.FormatNode(&t.Name)
				ctx.WriteByte(' ')
				ctx.WriteString(t.Type.SQLString())
			}
		}
	}
	for i, option := range node.Options {
		if i != 0 {
			ctx.WriteByte(' ')
		}
		ctx.FormatNode(option)
	}
}

type SimpleColumnDef struct {
	Name Name
	Type ResolvableTypeReference
}

type RoutineArgs []*RoutineArg

// RoutineArg represents a routine argument. It can be used by FUNCTIONs and PROCEDUREs.
type RoutineArg struct {
	Mode    string
	Name    Name
	Type    ResolvableTypeReference
	Default Expr
}

// Format implements the NodeFormatter interface.
func (node RoutineArgs) Format(ctx *FmtCtx) {
	for i, t := range node {
		if i != 0 {
			ctx.WriteString(", ")
		}
		if t.Mode != "" {
			ctx.WriteString(t.Mode)
			ctx.WriteByte(' ')
		}
		if t.Name != "" {
			ctx.FormatNode(&t.Name)
			ctx.WriteByte(' ')
		}
		t.Type.SQLString()
	}
}

type FunctionOption int8

const (
	OptionLanguage FunctionOption = 1 + iota
	OptionTransform
	OptionWindow
	OptionVolatility
	OptionLeakProof
	OptionNullInput
	OptionSecurity
	OptionParallel
	OptionCost
	OptionRows
	OptionSupport
	OptionSet
	OptionAs1
	OptionAs2
	OptionSqlBody
	OptionReset // For ALTER { FUNCTION | PROCEDURE } use only
)

type RoutineOption struct {
	OptionType FunctionOption
	// these members cannot be defined more than once
	Language       string
	TransformTypes []ResolvableTypeReference
	Volatility     Volatility
	IsLeakProof    bool
	NullInput      NullInput
	External       bool
	Definer        bool // true if Definer, false if Invoker
	Parallel       Parallel
	Cost           Expr // positive number
	Rows           Expr // positive number
	Support        string
	SetVar         *SetVar
	Definition     string // It can be an internal function name, the path to an object file, an SQL command, or text in a procedural language.
	ObjFile        string
	LinkSymbol     string
	SqlBody        Statement

	// For ALTER { FUNCTION | PROCEDURE } use only
	ResetParam string
	ResetAll   bool
}

// Format implements the NodeFormatter interface.
func (node RoutineOption) Format(ctx *FmtCtx) {
	switch node.OptionType {
	case OptionLanguage:
		ctx.WriteString("LANGUAGE ")
		ctx.WriteString(node.Language)
	case OptionTransform:
		ctx.WriteString("TRANSFORM ")
		for i, t := range node.TransformTypes {
			if i != 0 {
				ctx.WriteString(", ")
			}
			ctx.WriteString("FOR TYPE ")
			ctx.WriteString(t.SQLString())
		}
	case OptionWindow:
		ctx.WriteString("WINDOW")
	case OptionVolatility:
		switch node.Volatility {
		case VolatilityImmutable:
			ctx.WriteString("IMMUTABLE")
		case VolatilityStable:
			ctx.WriteString("STABLE")
		case VolatilityVolatile:
			ctx.WriteString("VOLATILE")
		default:
		}
	case OptionLeakProof:
		if !node.IsLeakProof {
			ctx.WriteString("NOT ")
		}
		ctx.WriteString("LEAKPROOF")
	case OptionNullInput:
		ctx.WriteString(strings.ToUpper(string(node.NullInput)))
	case OptionSecurity:
		if node.External {
			ctx.WriteString("EXTERNAL ")
		}
		ctx.WriteString("SECURITY ")
		if node.Definer {
			ctx.WriteString("DEFINER")
		} else {
			ctx.WriteString("INVOKER")
		}
	case OptionParallel:
		ctx.WriteString("PARALLEL")
		ctx.WriteString(strings.ToUpper(string(node.Parallel)))
	case OptionCost:
		ctx.WriteString("COST ")
		ctx.FormatNode(node.Cost)
	case OptionRows:
		ctx.WriteString("ROWS ")
		ctx.FormatNode(node.Rows)
	case OptionSupport:
		ctx.WriteString("SUPPORT ")
		ctx.WriteString(node.Support)
	case OptionSet:
		ctx.WriteString("SET ")
		ctx.FormatNode(node.SetVar)
	case OptionAs1:
		ctx.WriteString("AS ")
		ctx.WriteString(node.Definition)
	case OptionAs2:
		ctx.WriteString("AS ")
		ctx.WriteString(node.ObjFile)
		ctx.WriteString(", ")
		ctx.WriteString(node.LinkSymbol)
	case OptionSqlBody:
		ctx.FormatNode(node.SqlBody)
	case OptionReset:
		ctx.WriteString("RESET ")
		if node.ResetAll {
			ctx.WriteString("ALL")
		} else {
			ctx.WriteString(node.ResetParam)
		}
	}
}

type NullInput string

const (
	CalledOnNullInput      NullInput = "called on null input" // default
	ReturnsNullOnNullInput NullInput = "returns null on null input"
	StrictNullInput        NullInput = "strict"
)

type Parallel string

const (
	ParallelUnsafe     Parallel = "unsafe"
	ParallelRestricted Parallel = "restricted"
	ParallelSafe       Parallel = "safe"
)

var _ Statement = &BeginEndBlock{}

// 'BEGIN ATOMIC ... END' and 'RETURN' statements are used in `sql_body` of FUNCTIONs and PROCEDUREs.

// BeginEndBlock represents a BEGIN ATOMIC ... END block with one or more statements nested within
type BeginEndBlock struct {
	Statements []Statement
}

func (node *BeginEndBlock) Format(ctx *FmtCtx) {
	ctx.WriteString("BEGIN ATOMIC")
	for i, s := range node.Statements {
		if i != 0 {
			ctx.WriteString("; ")
		}
		ctx.FormatNode(s)
	}
	ctx.WriteString(" END")
}

var _ Statement = &Return{}

type Return struct {
	Expr Expr
}

func (node *Return) Format(ctx *FmtCtx) {
	ctx.WriteString("RETURN ")
	ctx.FormatNode(node.Expr)
}
