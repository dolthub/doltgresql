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

package analyzer

import (
	"fmt"
	"io"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/index"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
)

// TODO: come up with a better name
func ResolvedTable(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector) (sql.Node, transform.TreeIdentity, error) {
	return transform.Node(node, func(n sql.Node) (sql.Node, transform.TreeIdentity, error) {
		if filter, ok := n.(*plan.Filter); ok {
			return transform.Node(filter, resolvedTableFilter)
		}
		return n, transform.SameTree, nil
	})
}

// TODO: come up with a better name
func resolvedTableFilter(n sql.Node) (sql.Node, transform.TreeIdentity, error) {
	switch n := n.(type) {
	case *plan.ResolvedTable:
		if newTable, ok, err := resolvedTableFilterInner(n.UnderlyingTable()); err != nil {
			return n, transform.SameTree, err
		} else if ok {
			nt, err := n.WithTable(newTable)
			return nt, transform.NewTree, err
		}
		return n, transform.SameTree, nil
	default:
		return n, transform.SameTree, nil
	}
}

// TODO: come up with a better name
func resolvedTableFilterInner(table sql.Table) (sql.Table, bool, error) {
	switch table := table.(type) {
	case *sqle.AlterableDoltTable:
		return &DoltgresTable{table.WritableDoltTable.DoltTable}, true, nil
	case *sqle.WritableDoltTable:
		return &DoltgresTable{table.DoltTable}, true, nil
	case *sqle.DoltTable:
		return &DoltgresTable{table}, true, nil
	case *DoltgresTable:
		return table, false, nil
	default:
		return nil, false, fmt.Errorf("unknown table type: %T", table)
	}
}

// TODO: doc
type DoltgresTable struct {
	*sqle.DoltTable
}

var _ sql.Table = (*DoltgresTable)(nil)
var _ sql.ProjectedTable = (*DoltgresTable)(nil)
var _ sql.IndexSearchableTable = (*DoltgresTable)(nil)

// IndexedAccess implements the sql.IndexSearchableTable interface.
func (dt *DoltgresTable) IndexedAccess(lookup sql.IndexLookup) sql.IndexedTable {
	return &IndexedDoltgresTable{
		DoltTable: dt.DoltTable,
		idx:       lookup.Index,
	}
}

// LookupForExpressions implements the sql.IndexSearchableTable interface.
func (dt *DoltgresTable) LookupForExpressions(ctx *sql.Context, exprs []sql.Expression) (sql.IndexLookup, error) {
	allIndexes, err := dt.DoltTable.GetIndexes(ctx)
	if err != nil {
		return sql.IndexLookup{}, err
	}
	if len(allIndexes) == 0 {
		return sql.IndexLookup{}, nil
	}
	var potentialIndex sql.Index
	var potentialType sql.Type
	for _, expr := range exprs {
		transform.InspectExpr(expr, func(expr sql.Expression) bool {
			getField, ok := expr.(*expression.GetField)
			if !ok {
				return false
			}
			qualifiedFieldName := fmt.Sprintf("%s.%s", getField.Table(), getField.Name())
			// TODO: handle partial and composite indexes
			if potentialIndex != nil {
				if qualifiedFieldName != potentialIndex.Expressions()[0] {
					potentialIndex = nil
				}
			} else {
				for _, idx := range allIndexes {
					// TODO: handle composite indexes, POC is easier for non-composite indexes
					colNames := idx.Expressions()
					if len(colNames) != 1 {
						continue
					}
					if qualifiedFieldName == colNames[0] {
						// TODO: handle multiple indexes on the same column
						potentialIndex = idx
						potentialType = getField.Type()
						allIndexes = nil
						break
					}
				}
			}
			return true
		})
	}
	if potentialIndex == nil {
		return sql.IndexLookup{}, nil
	}
	var startExprs []sql.Expression
	var stopExprs []sql.Expression
	for _, expr := range exprs {
		switch expr := expr.(type) {
		case *pgexprs.BinaryOperator:
			switch expr.Operator() {
			case framework.Operator_BinaryEqual:
				startExpr := framework.GetBinaryFunction(framework.Operator_BinaryGreaterOrEqual).Compile("internal_index_start", expr.Left(), expr.Right())
				stopExpr := framework.GetBinaryFunction(framework.Operator_BinaryGreaterThan).Compile("internal_index_stop", expr.Left(), expr.Right())
				if startExpr == nil || stopExpr == nil {
					// TODO: maybe error if we can't find the complementary operators?
					return sql.IndexLookup{}, nil
				}
				startExprs = []sql.Expression{startExpr}
				stopExprs = []sql.Expression{stopExpr}
			case framework.Operator_BinaryGreaterThan:
				startExprs = []sql.Expression{expr}
				stopExprs = []sql.Expression{pgexprs.NewRawLiteralBool(false)}
			case framework.Operator_BinaryGreaterOrEqual:
				// TODO
			case framework.Operator_BinaryLessThan:
				// TODO
			case framework.Operator_BinaryLessOrEqual:
				// TODO
			default:
				// We only support the above operator types, since those are the ones used for b-tree indexing
				return sql.IndexLookup{}, nil
			}
		default:
			// TODO: support other expression types
			return sql.IndexLookup{}, nil
		}
	}
	return sql.IndexLookup{
		Index: specialIndexContainer{
			Index:       potentialIndex,
			startExprs:  startExprs,
			stopExprs:   stopExprs,
			filterExprs: exprs,
		},
		Ranges: sql.RangeCollection{ // TODO: figure out if this is even necessary
			sql.Range{
				sql.RangeColumnExpr{
					LowerBound: sql.BelowNull{},
					UpperBound: sql.AboveAll{},
					Typ:        potentialType,
				},
			},
		},
	}, nil
}

// PreciseMatch implements the sql.IndexSearchableTable interface.
func (dt *DoltgresTable) PreciseMatch() bool {
	return false // TODO: determine if true or false
}

// SkipIndexCosting implements the sql.IndexSearchableTable interface.
func (dt *DoltgresTable) SkipIndexCosting() bool {
	return true
}

// WithProjections implements the sql.ProjectedTable interface.
func (dt *DoltgresTable) WithProjections(colNames []string) sql.Table {
	return &DoltgresTable{dt.DoltTable.WithProjections(colNames).(*sqle.DoltTable)}
}

// TODO: doc
type IndexedDoltgresTable struct {
	*sqle.DoltTable
	idx sql.Index
}

var _ sql.IndexedTable = (*IndexedDoltgresTable)(nil)

// LookupPartitions implements the sql.IndexedTable interface.
func (idt *IndexedDoltgresTable) LookupPartitions(ctx *sql.Context, lookup sql.IndexLookup) (sql.PartitionIter, error) {
	indexContainer := lookup.Index.(specialIndexContainer)
	lookup.Index = indexContainer.Index
	return &specialIndexPartitionIter{
		specialIndexPartition: specialIndexPartition{
			specialIndexContainer: indexContainer,
			lookup:                lookup,
		},
		used: false,
	}, nil
}

// Partitions implements the sql.Table interface.
func (idt *IndexedDoltgresTable) Partitions(ctx *sql.Context) (sql.PartitionIter, error) {
	return nil, fmt.Errorf("%T: Partitions is invalid on this table", idt)
}

// PartitionRows implements the sql.Table interface.
func (idt *IndexedDoltgresTable) PartitionRows(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	part := partition.(specialIndexPartition)
	return index.RawIndexIterator(ctx, idt.DoltTable, idt.DoltTable.ProjectedTags(), idt.PrimaryKeySchema(), part.lookup, part.startExprs, part.stopExprs)
}

// WithProjections implements the sql.ProjectedTable interface.
func (idt *IndexedDoltgresTable) WithProjections(colNames []string) sql.Table {
	return &DoltgresTable{idt.DoltTable.WithProjections(colNames).(*sqle.DoltTable)}
}

// TODO: doc
type specialIndexContainer struct {
	sql.Index
	startExprs  []sql.Expression
	stopExprs   []sql.Expression
	filterExprs []sql.Expression
}

// TODO: doc
type specialIndexPartitionIter struct {
	specialIndexPartition
	used bool
}

// TODO: doc
type specialIndexPartition struct {
	specialIndexContainer
	lookup sql.IndexLookup
}

var _ sql.PartitionIter = (*specialIndexPartitionIter)(nil)
var _ sql.Partition = specialIndexPartition{}

// Close implements the sql.PartitionIter interface.
func (iter *specialIndexPartitionIter) Close(context *sql.Context) error {
	return nil
}

// Next implements the sql.PartitionIter interface.
func (iter *specialIndexPartitionIter) Next(context *sql.Context) (sql.Partition, error) {
	if iter.used {
		return nil, io.EOF
	}
	iter.used = true
	return iter.specialIndexPartition, nil
}

// Key implements the sql.Partition interface.
func (s specialIndexPartition) Key() []byte {
	return nil
}
