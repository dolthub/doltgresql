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

// Copyright 2020 The Cockroach Authors.
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

var _ Statement = &AlterAggregate{}

// AlterAggregate represents a ALTER AGGREGATE statement.
type AlterAggregate struct {
	Name   Name
	AggSig *AggregateSignature
	Rename Name
	Owner  string
	Schema string
}

// Format implements the NodeFormatter interface.
func (node *AlterAggregate) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER AGGREGATE ")
	ctx.FormatNode(&node.Name)
	ctx.WriteString(" (")
	node.AggSig.Format(ctx)
	ctx.WriteString(" )")

	if node.Rename != "" {
		ctx.WriteString(" RENAME ")
		ctx.FormatNode(&node.Rename)
	} else if node.Owner != "" {
		ctx.WriteString(" OWNER TO ")
		ctx.FormatNameP(&node.Owner)
	} else if node.Schema != "" {
		ctx.WriteString(" SET SCHEMA ")
		ctx.FormatNameP(&node.Schema)
	}
}

// AggregateSignature represents an aggregate_signature clause.
type AggregateSignature struct {
	All     bool
	Arg     *AggregateArg
	OrderBy *AggregateArg
}

// Format implements the NodeFormatter interface.
func (node *AggregateSignature) Format(ctx *FmtCtx) {
	if node.All {
		ctx.WriteString(" * ")
	} else {
		if node.Arg != nil {
			ctx.WriteByte(' ')
			node.Arg.Format(ctx)
		}
		if node.OrderBy != nil {
			ctx.WriteString(" ORDER BY ")
			node.Arg.Format(ctx)
		}
	}
}

// AggregateArg represents an aggregate argument(s).
type AggregateArg struct {
	Mode  string
	Name  Name
	Types []ResolvableTypeReference
}

// Format implements the NodeFormatter interface.
func (node *AggregateArg) Format(ctx *FmtCtx) {
	ctx.WriteString(node.Mode)
	if node.Name != "" {
		ctx.WriteByte(' ')
		ctx.FormatNode(&node.Name)
	}
	types := make([]string, len(node.Types))
	for i, t := range node.Types {
		types[i] = t.SQLString()
	}

	if len(types) > 0 {
		ctx.WriteByte(' ')
		ctx.WriteString(strings.Join(types, ", "))
	}
}
