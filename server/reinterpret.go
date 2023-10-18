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

package server

import (
	"fmt"

	"github.com/dolthub/doltgresql/postgres/parser/parser"
)

// reinterpretQuery takes the given Postgres query, and reinterprets it as a query that will work with the handler.
func (l *Listener) reinterpretQuery(query string) (string, error) {
	s, err := parser.Parse(query)
	if err != nil {
		return "", err
	}
	if len(s) > 1 {
		return "", fmt.Errorf("only a single statement at a time is currently supported")
	}
	return s[0].AST.String(), nil
}
