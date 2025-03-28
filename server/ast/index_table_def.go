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
	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeIndexTableDef handles *tree.IndexTableDef nodes. The parser does not store type information in the index
// definition (PRIMARY KEY, UNIQUE, etc.) so it must be added to this definition by the caller.
func nodeIndexTableDef(ctx *Context, node *tree.IndexTableDef) (*vitess.IndexDefinition, error) {
	if node == nil {
		return nil, nil
	}
	if node.IndexParams.IncludeColumns != nil {
		return nil, errors.Errorf("include columns is not yet supported")
	}
	if len(node.IndexParams.StorageParams) > 0 {
		logrus.Warn("storage params are not yet supported, ignoring")
	}
	if node.IndexParams.Tablespace != "" {
		logrus.Warn("tablespace is not yet supported, ignoring")
	}

	columns, err := nodeIndexElemList(ctx, node.Columns)
	if err != nil {
		return nil, err
	}

	return &vitess.IndexDefinition{
		Info: &vitess.IndexInfo{
			Type: "",
			Name: vitess.NewColIdent(string(node.Name)),
		},
		Columns: columns,
	}, nil
}
