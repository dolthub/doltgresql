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

import (
	"fmt"
	"strings"

	"github.com/dolthub/doltgresql/postgres/parser/privilege"
)

var _ Statement = &AlterDefaultPrivileges{}

// AlterDefaultPrivileges represents a ALTER DEFAULT PRIVILEGES statement.
type AlterDefaultPrivileges struct {
	ForRole      bool
	TargetRoles  []string
	Privileges   privilege.List
	Target       TargetList
	Grantees     []string
	GrantOption  bool
	DropBehavior DropBehavior
	Grant        bool
}

// Format implements the NodeFormatter interface.
func (node *AlterDefaultPrivileges) Format(ctx *FmtCtx) {
	ctx.WriteString("ALTER DEFAULT PRIVILEGES ")
	if len(node.TargetRoles) > 0 {
		ctx.WriteString("FOR ")
		if node.ForRole {
			ctx.WriteString("ROLE ")
		} else {
			ctx.WriteString("USER ")
		}
		ctx.WriteString(strings.Join(node.TargetRoles, ", "))
	}
	if len(node.Target.InSchema) > 0 {
		ctx.WriteString("IN SCHEMAS ")
		ctx.WriteString(strings.Join(node.Target.InSchema, ", "))
	}
	if node.Grant {
		ctx.WriteString(" GRANT ")
		node.Privileges.Format(&ctx.Buffer)
		ctx.WriteString(fmt.Sprintf(" ON %sS TO ", strings.ToUpper(string(node.Target.TargetType))))
		ctx.WriteString(strings.Join(node.Grantees, ", "))
		if node.GrantOption {
			ctx.WriteString(" WITH GRANT OPTION")
		}
	} else {
		ctx.WriteString(" REVOKE ")
		if node.GrantOption {
			ctx.WriteString(" GRANT OPTION FOR ")
		}
		node.Privileges.Format(&ctx.Buffer)
		ctx.WriteString(fmt.Sprintf(" ON %sS FROM ", strings.ToUpper(string(node.Target.TargetType))))
		ctx.WriteString(strings.Join(node.Grantees, ", "))
		if node.DropBehavior.String() != "" {
			ctx.WriteByte(' ')
			ctx.WriteString(node.DropBehavior.String())
		}
	}
}
