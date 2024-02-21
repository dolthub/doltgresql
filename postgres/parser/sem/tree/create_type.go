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

import "github.com/dolthub/doltgresql/postgres/parser/lex"

// CreateTypeVariety represents a particular variety of user defined types.
type CreateTypeVariety int

//go:generate stringer -type=CreateTypeVariety
const (
	_ CreateTypeVariety = iota
	// Enum represents an ENUM user defined type.
	Enum
	// Composite represents a composite user defined type.
	Composite
	// Range represents a RANGE user defined type.
	Range
	// Base represents a base user defined type.
	Base
	// Shell represents a shell user defined type.
	Shell
	// Domain represents a DOMAIN user defined type.
	Domain
)

var _ Statement = &CreateType{}

// CreateType represents a CREATE TYPE statement.
type CreateType struct {
	TypeName *UnresolvedObjectName
	Variety  CreateTypeVariety
	// EnumLabels is set when this represents a CREATE TYPE ... AS ENUM statement.
	EnumLabels []string
}

var _ Statement = &CreateType{}

// Format implements the NodeFormatter interface.
func (node *CreateType) Format(ctx *FmtCtx) {
	ctx.WriteString("CREATE TYPE ")
	ctx.WriteString(node.TypeName.String())
	ctx.WriteString(" ")
	switch node.Variety {
	case Enum:
		ctx.WriteString("AS ENUM (")
		for i := range node.EnumLabels {
			if i > 0 {
				ctx.WriteString(", ")
			}
			lex.EncodeSQLString(&ctx.Buffer, node.EnumLabels[i])
		}
		ctx.WriteString(")")
	}
}
