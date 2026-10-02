// Copyright 2026 Dolthub, Inc.
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
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgnodes "github.com/dolthub/doltgresql/server/node"
)

// nodeDeclareCursor handles *tree.DeclareCursor nodes.
func nodeDeclareCursor(ctx *Context, node *tree.DeclareCursor) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	if node.Options&tree.CursorOptionScroll != 0 && node.Options&tree.CursorOptionNoScroll != 0 {
		return nil, pgerror.New(pgcode.InvalidCursorDefinition, "cannot specify both SCROLL and NO SCROLL")
	}
	if node.Options&tree.CursorOptionAsensitive != 0 && node.Options&tree.CursorOptionInsensitive != 0 {
		return nil, pgerror.New(pgcode.InvalidCursorDefinition, "cannot specify both ASENSITIVE and INSENSITIVE")
	}
	if node.Options&tree.CursorOptionBinary != 0 {
		//TODO: support BINARY cursors, which return their rows in the binary format over the simple query protocol
		return NotYetSupportedError("BINARY cursors are not yet supported")
	}
	if len(node.Select.Locking) > 0 {
		strength := node.Select.Locking[0].Strength.String()
		if node.Options&tree.CursorOptionHold != 0 {
			return nil, errors.WithDetail(pgerror.Newf(pgcode.FeatureNotSupported,
				"DECLARE CURSOR WITH HOLD ... %s is not supported", strength), "Holdable cursors must be READ ONLY.")
		}
		if node.Options&tree.CursorOptionScroll != 0 {
			return nil, errors.WithDetail(pgerror.Newf(pgcode.FeatureNotSupported,
				"DECLARE SCROLL CURSOR ... %s is not supported", strength), "Scrollable cursors must be READ ONLY.")
		}
		if node.Options&tree.CursorOptionInsensitive != 0 {
			return nil, errors.WithDetail(pgerror.Newf(pgcode.FeatureNotSupported,
				"DECLARE INSENSITIVE CURSOR ... %s is not supported", strength), "Insensitive cursors must be READ ONLY.")
		}
	}
	selectStmt, err := nodeSelect(ctx, node.Select)
	if err != nil {
		return nil, err
	}
	return vitess.InjectedStatement{
		Statement: &pgnodes.DeclareCursor{
			Name:         string(node.Name),
			Select:       selectStmt,
			IsHoldable:   node.Options&tree.CursorOptionHold != 0,
			IsScrollable: node.Options&tree.CursorOptionScroll != 0,
		},
	}, nil
}
