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

type AlterDatabaseOption string

// Names of options on ALTER DATABASE.
const (
	OptAllowConnections AlterDatabaseOption = "ALLOW_CONNECTIONS"
	OptConnectionLimit  AlterDatabaseOption = "CONNECTION LIMIT"
	OptIsTemplate       AlterDatabaseOption = "IS_TEMPLATE"
)

// DatabaseOption represents a ALTER DATABASE option.
type DatabaseOption struct {
	Opt AlterDatabaseOption
	Val Expr
}

var _ Statement = &AlterDatabase{}

// AlterDatabase represents a ALTER DATABASE statement.
type AlterDatabase struct {
	Name    Name
	Options []DatabaseOption
	// Rename is handled by RenameDatabase
	Owner                   string
	Tablespace              string
	RefreshCollationVersion bool
	SetVar                  *SetVar
	ResetVar                string
	ResetAll                bool
}

// Format implements the NodeFormatter interface.
func (node *AlterDatabase) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER DATABASE ")
	ctx.FormatNode(&node.Name)
	if len(node.Options) > 0 {
		opts := make([]string, len(node.Options))
		for i, opt := range node.Options {
			opts[i] = fmt.Sprintf("%s %s", opt.Opt, AsString(opt.Val))
		}
		ctx.WriteString(fmt.Sprintf(" WITH %s", strings.Join(opts, " ")))
	} else if node.Owner != "" {
		ctx.WriteString(" OWNER TO ")
		ctx.FormatNameP(&node.Owner)
	} else if node.Tablespace != "" {
		ctx.WriteString(" SET TABLESPACE ")
		ctx.FormatNameP(&node.Tablespace)
	} else if node.RefreshCollationVersion {
		ctx.WriteString(" REFRESH COLLATION VERSION")
	} else if node.SetVar != nil {
		ctx.WriteByte(' ')
		node.SetVar.Format(ctx)
	} else if node.ResetVar != "" {
		ctx.WriteString(" RESET ")
		ctx.FormatNameP(&node.ResetVar)
	} else if node.ResetAll {
		ctx.WriteString(" RESET ALL")
	}
}
