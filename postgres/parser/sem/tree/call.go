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

package tree

var _ Statement = &Call{}

// Call represents a CALL statement.
type Call struct {
	Procedure *FuncExpr
}

// StatementType implements the interface Statement.
func (c *Call) StatementType() StatementType {
	return Rows
}

// StatementTag implements the interface Statement.
func (c *Call) StatementTag() string {
	return "CALL"
}

// Format implements the interface Statement.
func (c *Call) Format(ctx *FmtCtx) {
	ctx.WriteString("CALL ")
	ctx.FormatNode(c.Procedure)
}

// String implements the interface Statement.
func (c *Call) String() string {
	return AsString(c)
}
