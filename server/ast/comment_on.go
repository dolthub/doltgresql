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

// nodeCommentOnColumn handles *tree.CommentOnColumn nodes.
func nodeCommentOnColumn(node *tree.CommentOnColumn) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	return nil, fmt.Errorf("COMMENT ON COLUMN is not yet supported")
}

// nodeCommentOnDatabase handles *tree.CommentOnDatabase nodes.
func nodeCommentOnDatabase(node *tree.CommentOnDatabase) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	return nil, fmt.Errorf("COMMENT ON DATABASE is not yet supported")
}

// nodeCommentOnIndex handles *tree.CommentOnIndex nodes.
func nodeCommentOnIndex(node *tree.CommentOnIndex) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	return nil, fmt.Errorf("COMMENT ON INDEX is not yet supported")
}

// nodeCommentOnTable handles *tree.CommentOnTable nodes.
func nodeCommentOnTable(node *tree.CommentOnTable) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	return nil, fmt.Errorf("COMMENT ON TABLE is not yet supported")
}
