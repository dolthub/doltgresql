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
