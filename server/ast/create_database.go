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
	if len(node.Template) > 0 {
		return nil, fmt.Errorf("templates are not yet supported")
	}
	if len(node.Encoding) > 0 {
		return nil, fmt.Errorf("encodings are not yet supported")
	}
	if len(node.Collate) > 0 {
		return nil, fmt.Errorf("collations are not yet supported")
	}
	if len(node.CType) > 0 {
		return nil, fmt.Errorf("ctypes are not yet supported")
	}
	return &vitess.DBDDL{
		Action:      vitess.CreateStr,
		DBName:      node.Name.String(),
		IfNotExists: node.IfNotExists,
	}, nil
}
