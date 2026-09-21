// Copyright 2026 Dolthub, Inc.
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

// Do represents an anonymous procedural code block.
type Do struct {
	Language string
	Code     string
}

var _ Statement = (*Do)(nil)

// Format implements the NodeFormatter interface.
func (d *Do) Format(ctx *FmtCtx) {
	ctx.WriteString("DO ")
	if d.Language != "" {
		ctx.WriteString("LANGUAGE ")
		ctx.WriteString(d.Language)
		ctx.WriteByte(' ')
	}
	lex.EncodeSQLString(&ctx.Buffer, d.Code)
}

// String returns the statement formatted as SQL.
func (d *Do) String() string { return AsString(d) }
