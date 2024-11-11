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

package auth

import "github.com/dolthub/doltgresql/utils"

// AuthContext contains the auth portion of the context when converting from the Postgres AST to the Vitess AST.
type AuthContext struct {
	authType utils.Stack[string]
}

// NewAuthContext returns a new *AuthContext.
func NewAuthContext() *AuthContext {
	return &AuthContext{}
}

// PushAuthType pushes the given AuthType into the context's stack.
func (ctx *AuthContext) PushAuthType(authType string) {
	ctx.authType.Push(authType)
}

// PeekAuthType returns the AuthType that is on the top of the stack. This does not remove it from the stack. Returns
// AuthType_IGNORE if the stack is empty.
func (ctx *AuthContext) PeekAuthType() string {
	if ctx.authType.Empty() {
		return AuthType_IGNORE
	}
	return ctx.authType.Peek()
}

// PopAuthType returns the AuthType that is on the top of the stack. This also removes it from the stack. Returns
// AuthType_IGNORE if the stack is empty.
func (ctx *AuthContext) PopAuthType() string {
	if ctx.authType.Empty() {
		return AuthType_IGNORE
	}
	return ctx.authType.Pop()
}
