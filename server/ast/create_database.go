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

package ast

import (
	"fmt"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeCreateDatabase handles *tree.CreateDatabase nodes.
func nodeCreateDatabase(node *tree.CreateDatabase) (*vitess.DBDDL, error) {
	if len(node.Owner) > 0 {
		return nil, fmt.Errorf("OWNER clause is not yet supported")
	}
	if len(node.Template) > 0 {
		return nil, fmt.Errorf("TEMPLATE clause is not yet supported")
	}
	if len(node.Encoding) > 0 {
		return nil, fmt.Errorf("ENCODING clause is not yet supported")
	}
	if len(node.Strategy) > 0 {
		return nil, fmt.Errorf("STRATEGY clause is not yet supported")
	}
	if len(node.Locale) > 0 {
		return nil, fmt.Errorf("LOCALE clause is not yet supported")
	}
	if len(node.Collate) > 0 {
		return nil, fmt.Errorf("LC_COLLATE clause is not yet supported")
	}
	if len(node.CType) > 0 {
		return nil, fmt.Errorf("LC_CTYPE clause is not yet supported")
	}
	if len(node.IcuLocale) > 0 {
		return nil, fmt.Errorf("ICU_LOCALE clause is not yet supported")
	}
	if len(node.IcuRules) > 0 {
		return nil, fmt.Errorf("TEMPLATE clause is not yet supported")
	}
	if len(node.LocaleProvider) > 0 {
		return nil, fmt.Errorf("LOCALE_PROVIDER clause is not yet supported")
	}
	if len(node.CollationVersion) > 0 {
		return nil, fmt.Errorf("COLLATION_VERSION clause is not yet supported")
	}
	if len(node.Tablespace) > 0 {
		return nil, fmt.Errorf("TABLESPACE clause is not yet supported")
	}
	// TODO: some clauses have default values in case of not being defined.
	// ALLOW_CONNECTIONS defaults to TRUE
	if node.AllowConnections != nil {
		return nil, fmt.Errorf("ALLOW_CONNECTIONS clause is not yet supported")
	}
	// CONNECTION LIMIT defaults to -1
	if node.ConnectionLimit != nil {
		return nil, fmt.Errorf("CONNECTION LIMIT clause is not yet supported")
	}
	// IS_TEMPLATE defaults to FALSE
	if node.IsTemplate != nil {
		return nil, fmt.Errorf("IS_TEMPLATE clause is not yet supported")
	}
	if node.Oid != nil {
		return nil, fmt.Errorf("OID clause is not yet supported")
	}

	return &vitess.DBDDL{
		Action:      vitess.CreateStr,
		SchemaOrDatabase: "database",
		DBName:      node.Name.String(),
		IfNotExists: node.IfNotExists,
	}, nil
}
