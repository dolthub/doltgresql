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

// Copyright 2012, Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in licenses/BSD-vitess.txt.

// Portions of this file are additionally subject to the following
// license and copyright.
//
// Copyright 2015 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

// This code was derived from https://github.com/youtube/vitess.

package tree

import (
	"github.com/dolthub/doltgresql/postgres/parser/lex"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
)

var _ Statement = &CreateDatabase{}

// CreateDatabase represents a CREATE DATABASE statement.
type CreateDatabase struct {
	IfNotExists      bool
	Name             Name
	Owner            string
	Template         string
	Encoding         string
	Strategy         string
	Locale           string
	Collate          string
	CType            string
	IcuLocale        string
	IcuRules         string
	LocaleProvider   string
	CollationVersion string
	Tablespace       string
	AllowConnections Expr // default is true
	ConnectionLimit  Expr // default is -1
	IsTemplate       Expr // default is false
	Oid              Expr
}

// NewCreateDatabase builds a CREATE DATABASE statement from options given in any order, rejecting repeated options.
func NewCreateDatabase(name Name, ifNotExists bool, options []KVOption) (*CreateDatabase, error) {
	node := &CreateDatabase{Name: name, IfNotExists: ifNotExists}
	seen := make(map[Name]struct{}, len(options))
	for _, option := range options {
		if _, ok := seen[option.Key]; ok {
			return nil, pgerror.New(pgcode.Syntax, "conflicting or redundant options")
		}
		seen[option.Key] = struct{}{}
		switch option.Key {
		case "owner":
			node.Owner = string(*option.Value.(*DString))
		case "template":
			node.Template = string(*option.Value.(*DString))
		case "encoding":
			node.Encoding = string(*option.Value.(*DString))
		case "strategy":
			node.Strategy = string(*option.Value.(*DString))
		case "locale":
			node.Locale = string(*option.Value.(*DString))
		case "lc_collate":
			node.Collate = string(*option.Value.(*DString))
		case "lc_ctype":
			node.CType = string(*option.Value.(*DString))
		case "icu_locale":
			node.IcuLocale = string(*option.Value.(*DString))
		case "icu_rules":
			node.IcuRules = string(*option.Value.(*DString))
		case "locale_provider":
			node.LocaleProvider = string(*option.Value.(*DString))
		case "collation_version":
			node.CollationVersion = string(*option.Value.(*DString))
		case "tablespace":
			node.Tablespace = string(*option.Value.(*DString))
		case "allow_connections":
			node.AllowConnections = option.Value
		case "connection_limit":
			node.ConnectionLimit = option.Value
		case "is_template":
			node.IsTemplate = option.Value
		case "oid":
			node.Oid = option.Value
		}
	}
	return node, nil
}

// Format implements the NodeFormatter interface.
func (node *CreateDatabase) Format(ctx *FmtCtx) {
	ctx.WriteString("CREATE DATABASE ")
	if node.IfNotExists {
		ctx.WriteString("IF NOT EXISTS ")
	}
	ctx.FormatNode(&node.Name)
	if node.Owner != "" {
		ctx.WriteString(" OWNER = ")
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, node.Owner, ctx.flags.EncodeFlags())
	}
	if node.Template != "" {
		ctx.WriteString(" TEMPLATE = ")
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, node.Template, ctx.flags.EncodeFlags())
	}
	if node.Encoding != "" {
		ctx.WriteString(" ENCODING = ")
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, node.Encoding, ctx.flags.EncodeFlags())
	}
	if node.Strategy != "" {
		ctx.WriteString(" STRATEGY = ")
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, node.Strategy, ctx.flags.EncodeFlags())
	}
	if node.Locale != "" {
		ctx.WriteString(" LOCALE = ")
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, node.Locale, ctx.flags.EncodeFlags())
	}
	if node.Collate != "" {
		ctx.WriteString(" LC_COLLATE = ")
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, node.Collate, ctx.flags.EncodeFlags())
	}
	if node.CType != "" {
		ctx.WriteString(" LC_CTYPE = ")
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, node.CType, ctx.flags.EncodeFlags())
	}
	if node.IcuLocale != "" {
		ctx.WriteString(" ICU_LOCALE = ")
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, node.IcuLocale, ctx.flags.EncodeFlags())
	}
	if node.IcuRules != "" {
		ctx.WriteString(" ICU_RULES = ")
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, node.IcuRules, ctx.flags.EncodeFlags())
	}
	if node.LocaleProvider != "" {
		ctx.WriteString(" LOCALE_PROVIDER = ")
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, node.LocaleProvider, ctx.flags.EncodeFlags())
	}
	if node.CollationVersion != "" {
		ctx.WriteString(" COLLATION_VERSION = ")
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, node.CollationVersion, ctx.flags.EncodeFlags())
	}
	if node.Tablespace != "" {
		ctx.WriteString(" TABLESPACE = ")
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, node.Tablespace, ctx.flags.EncodeFlags())
	}
	if node.AllowConnections != nil {
		ctx.WriteString(" ALLOW_CONNECTIONS = ")
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, node.AllowConnections.String(), ctx.flags.EncodeFlags())
	}
	if node.ConnectionLimit != nil {
		ctx.WriteString(" CONNECTION LIMIT = ")
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, node.ConnectionLimit.String(), ctx.flags.EncodeFlags())
	}
	if node.IsTemplate != nil {
		ctx.WriteString(" IS_TEMPLATE = ")
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, node.IsTemplate.String(), ctx.flags.EncodeFlags())
	}
	if node.Oid != nil {
		ctx.WriteString(" OID = ")
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, node.Oid.String(), ctx.flags.EncodeFlags())
	}
}
