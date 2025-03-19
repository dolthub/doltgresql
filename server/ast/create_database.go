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
	"regexp"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeCreateDatabase handles *tree.CreateDatabase nodes.
func nodeCreateDatabase(ctx *Context, node *tree.CreateDatabase) (*vitess.DBDDL, error) {
	var charsets []*vitess.CharsetAndCollate

	if len(node.Template) > 0 {
		// TODO: special casing "template0", as some tests make use of it and we need them to pass for now
		if node.Template != "template0" {
			return nil, errors.Errorf("TEMPLATE clause is not yet supported")
		}
	}
	if len(node.Encoding) > 0 {
		logrus.Warnf("unsupported clause ENCODING, ignoring")
	}
	if len(node.Strategy) > 0 {
		return nil, errors.Errorf("STRATEGY clause is not yet supported")
	}
	if len(node.Locale) > 0 {
		return nil, errors.Errorf("LOCALE clause is not yet supported")
	}
	if len(node.Collate) > 0 {
		collation, charset, err := parseLocaleString(node.Collate)
		if err != nil {
			return nil, err
		}

		if collation == "" {
			logrus.Warnf("unsupported LC_COLLATE, ignoring")
		} else {
			charsets = append(charsets,
				&vitess.CharsetAndCollate{
					Type:  "CHARACTER SET",
					Value: charset,
				},
				&vitess.CharsetAndCollate{
					Type:  "COLLATE",
					Value: collation,
				},
			)
		}
	}
	if len(node.CType) > 0 {
		logrus.Warnf("CTYPE clause is not yet supported, ignoring")
	}
	if len(node.IcuLocale) > 0 {
		return nil, errors.Errorf("ICU_LOCALE clause is not yet supported")
	}
	if len(node.IcuRules) > 0 {
		return nil, errors.Errorf("TEMPLATE clause is not yet supported")
	}
	if len(node.LocaleProvider) > 0 {
		return nil, errors.Errorf("LOCALE_PROVIDER clause is not yet supported")
	}
	if len(node.CollationVersion) > 0 {
		return nil, errors.Errorf("COLLATION_VERSION clause is not yet supported")
	}
	if len(node.Tablespace) > 0 {
		return nil, errors.Errorf("TABLESPACE clause is not yet supported")
	}
	// TODO: some clauses have default values in case of not being defined.
	// ALLOW_CONNECTIONS defaults to TRUE
	if node.AllowConnections != nil {
		return nil, errors.Errorf("ALLOW_CONNECTIONS clause is not yet supported")
	}
	// CONNECTION LIMIT defaults to -1
	if node.ConnectionLimit != nil {
		return nil, errors.Errorf("CONNECTION LIMIT clause is not yet supported")
	}
	// IS_TEMPLATE defaults to FALSE
	if node.IsTemplate != nil {
		return nil, errors.Errorf("IS_TEMPLATE clause is not yet supported")
	}
	if node.Oid != nil {
		return nil, errors.Errorf("OID clause is not yet supported")
	}

	return &vitess.DBDDL{
		Action:           vitess.CreateStr,
		SchemaOrDatabase: "database",
		DBName:           node.Name.String(),
		IfNotExists:      node.IfNotExists,
		CharsetCollate:   charsets,
	}, nil
}

var collationRegex = regexp.MustCompile(`^(?P<Language>[^_]+)_?(?P<Region>[^.]+)?\.?(?P<CodePage>\d+)?$`)

// parseLocaleString attempts to parse the locale string given to extract a mysql collation we can use
func parseLocaleString(collation string) (string, string, error) {
	// FindStringSubmatchIndex returns the indices of the matched elements
	match := collationRegex.FindStringSubmatch(collation)

	result := make(map[string]string)
	for i, name := range collationRegex.SubexpNames() {
		if i > 0 && i <= len(match) {
			result[name] = match[i]
		}
	}

	if result["Language"] == "" {
		return "", "", errors.Errorf("malformed collation: %s", collation)
	}

	switch strings.ToLower(result["Language"]) {
	case "english", "en":
		return "latin1_general_cs", "latin1", nil
	}

	return "", "", nil
}
