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

var _ Statement = &CreateLanguage{}

// CreateLanguage represents a CREATE LANGUAGE statement.
type CreateLanguage struct {
	Name       Name
	Replace    bool
	Trusted    bool
	Procedural bool
	Handler    *LanguageHandler
}

func (node *CreateLanguage) Format(ctx *FmtCtx) {
	ctx.WriteString("CREATE ")
	if node.Replace {
		ctx.WriteString("REPLACE ")
	}
	if node.Trusted {
		ctx.WriteString("TRUSTED ")
	}
	if node.Procedural {
		ctx.WriteString("PROCEDURAL ")
	}
	ctx.WriteString("LANGUAGE ")
	ctx.FormatNode(&node.Name)
	if node.Handler != nil {
		ctx.FormatNode(node.Handler)
	}
}

type LanguageHandler struct {
	Handler   *UnresolvedObjectName
	Inline    *UnresolvedObjectName
	Validator *UnresolvedObjectName
}

func (h *LanguageHandler) Format(ctx *FmtCtx) {
	ctx.WriteString(" HANDLER ")
	ctx.FormatNode(h.Handler)
	if h.Inline != nil {
		ctx.WriteString(" INLINE ")
		ctx.FormatNode(h.Inline)
	}
	if h.Validator != nil {
		ctx.WriteString(" VALIDATOR ")
		ctx.FormatNode(h.Validator)
	}
}
