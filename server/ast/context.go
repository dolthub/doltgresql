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

package ast

import (
	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/server/auth"
)

// Context contains any relevant context for the AST conversion. For example, the auth system uses the context to
// determine which larger statement an expression exists in, which may influence how the expression should handle
// authorization.
type Context struct {
	authContext   *auth.AuthContext
	originalQuery string
}

// NewContext returns a new *Context.
func NewContext(postgresStmt parser.Statement) *Context {
	return &Context{
		authContext:   auth.NewAuthContext(),
		originalQuery: postgresStmt.SQL,
	}
}

// Auth returns the portion that handles authentication.
func (ctx *Context) Auth() *auth.AuthContext {
	return ctx.authContext
}
