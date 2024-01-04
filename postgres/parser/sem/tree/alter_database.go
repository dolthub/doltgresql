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

import (
	"fmt"
	"strings"
)

// AlterDatabaseOwner represents a ALTER DATABASE OWNER TO statement.
type AlterDatabaseOwner struct {
	Name  Name
	Owner string
}

// Format implements the NodeFormatter interface.
func (node *AlterDatabaseOwner) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER DATABASE ")
	ctx.FormatNode(&node.Name)
	ctx.WriteString(" OWNER TO ")
	ctx.FormatNameP(&node.Owner)
}

type AlterDatabaseOption string

// Names of options on ALTER DATABASE.
const (
	OptAllowConnections AlterDatabaseOption = "ALLOW_CONNECTIONS"
	OptConnectionLimit                      = "CONNECTION LIMIT"
	OptIsTemplate                           = "IS_TEMPLATE"
)

// DatabaseOption represents a ALTER DATABASE option.
type DatabaseOption struct {
	Opt AlterDatabaseOption
	Val Expr
}

// AlterDatabaseOptions represents a ALTER DATABASE OWNER TO statement.
type AlterDatabaseOptions struct {
	Name    Name
	Options []DatabaseOption
}

// Format implements the NodeFormatter interface.
func (node *AlterDatabaseOptions) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER DATABASE ")
	ctx.FormatNode(&node.Name)
	opts := make([]string, len(node.Options))
	for i, opt := range node.Options {
		opts[i] = fmt.Sprintf("%s %s", opt.Opt, AsString(opt.Val))
	}

	if len(opts) > 0 {
		ctx.WriteString(fmt.Sprintf(" WITH %s", strings.Join(opts, " ")))
	}
}
