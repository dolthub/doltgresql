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

// AlterSequence represents an ALTER SEQUENCE statement, except in the case of
// ALTER SEQUENCE <seqName> RENAME TO <newSeqName>, which is represented by a
// RenameTable node.
type AlterSequence struct {
	IfExists bool
	Name     *UnresolvedObjectName
	Options  SequenceOptions
	SetLog   bool
	Logged   bool
	Owner    string
}

// Format implements the NodeFormatter interface.
func (node *AlterSequence) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER SEQUENCE ")
	if node.IfExists {
		ctx.WriteString("IF EXISTS ")
	}
	ctx.FormatNode(node.Name)
	if node.Owner != "" {
		ctx.WriteString(" OWNER TO ")
		ctx.WriteString(node.Owner)
	} else if node.SetLog {
		if node.Logged {
			ctx.WriteString(" SET LOGGED")
		} else {
			ctx.WriteString(" SET UNLOGGED")
		}
	} else {
		ctx.FormatNode(&node.Options)
	}

}
