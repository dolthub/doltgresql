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

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeWindow handles *tree.Window nodes.
func nodeWindow(ctx *Context, node tree.Window) (vitess.Window, error) {
	if len(node) == 0 {
		return nil, nil
	}
	windows := make(vitess.Window, len(node))
	for i, def := range node {
		windowDef, err := nodeWindowDef(ctx, def)
		if err != nil {
			return nil, err
		}
		windows[i] = windowDef
	}
	return windows, nil
}

// nodeWindowDef handles *tree.WindowDef nodes.
func nodeWindowDef(ctx *Context, node *tree.WindowDef) (*vitess.WindowDef, error) {
	if node == nil {
		return nil, nil
	}
	partitionBy, err := nodeExprs(ctx, node.Partitions)
	if err != nil {
		return nil, err
	}
	orderBy, err := nodeOrderBy(ctx, node.OrderBy)
	if err != nil {
		return nil, err
	}
	var frame *vitess.Frame
	if node.Frame != nil {
		var unit vitess.FrameUnit
		switch node.Frame.Mode {
		case tree.RANGE:
			unit = vitess.RangeUnit
		case tree.ROWS:
			unit = vitess.RowsUnit
		case tree.GROUPS:
			return nil, errors.Errorf("GROUPS is not yet supported")
		default:
			return nil, errors.Errorf("unknown window frame mode")
		}
		var bounds [2]*vitess.FrameBound
		for i, bound := range []*tree.WindowFrameBound{node.Frame.Bounds.StartBound, node.Frame.Bounds.EndBound} {
			if bound == nil {
				continue
			}
			var boundType vitess.BoundType
			switch bound.BoundType {
			case tree.UnboundedPreceding:
				boundType = vitess.UnboundedPreceding
			case tree.OffsetPreceding:
				boundType = vitess.ExprPreceding
			case tree.CurrentRow:
				boundType = vitess.CurrentRow
			case tree.OffsetFollowing:
				boundType = vitess.ExprFollowing
			case tree.UnboundedFollowing:
				boundType = vitess.UnboundedFollowing
			default:
				return nil, errors.Errorf("unknown window frame bound type")
			}
			boundExpr, err := nodeExpr(ctx, bound.OffsetExpr)
			if err != nil {
				return nil, err
			}
			bounds[i] = &vitess.FrameBound{
				Expr: boundExpr,
				Type: boundType,
			}
		}
		frame = &vitess.Frame{
			Unit: unit,
			Extent: &vitess.FrameExtent{
				Start: bounds[0],
				End:   bounds[1],
			},
		}
	}
	return &vitess.WindowDef{
		Name:        vitess.NewColIdent(string(node.Name)),
		NameRef:     vitess.NewColIdent(string(node.RefName)),
		PartitionBy: partitionBy,
		OrderBy:     orderBy,
		Frame:       frame,
	}, nil
}
